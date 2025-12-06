package middleware

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"contract-pro-suite/internal/shared/config"
	"contract-pro-suite/services/auth/repository"
	"contract-pro-suite/services/auth/usecase"
	db "contract-pro-suite/sqlc"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

const (
	clientIDContextKey contextKey = "client_id"
	clientContextKey   contextKey = "client"
)

// ExtractSlugFromHost サブドメインからslugを抽出
func ExtractSlugFromHost(host string, baseDomain string) (string, error) {
	// ポート番号を除去
	if idx := strings.Index(host, ":"); idx != -1 {
		host = host[:idx]
	}

	// ベースドメインで終わるか確認
	if !strings.HasSuffix(host, "."+baseDomain) && host != baseDomain {
		return "", errors.New("invalid domain")
	}

	// サブドメイン部分を抽出
	if host == baseDomain {
		return "", errors.New("no subdomain found")
	}

	// "subdomain.baseDomain" から "subdomain" を抽出
	subdomain := strings.TrimSuffix(host, "."+baseDomain)
	if subdomain == "" {
		return "", errors.New("empty subdomain")
	}

	return subdomain, nil
}

// ValidateSubdomain サブドメインが許可されたドメインか検証
func ValidateSubdomain(host string, cfg *config.Config) error {
	if !cfg.EnableSubdomainValidation {
		return nil // 検証が無効な場合は常に許可
	}

	allowedDomains := cfg.AllowedDomains()
	for _, allowedDomain := range allowedDomains {
		if strings.HasSuffix(host, allowedDomain) || host == allowedDomain {
			return nil
		}
	}

	return errors.New("domain not allowed")
}

// GetClientIDBySlug slugからclient_id（UUID）を取得
func GetClientIDBySlug(ctx context.Context, slug string, clientRepo repository.ClientRepository) (uuid.UUID, error) {
	client, err := clientRepo.GetBySlug(ctx, slug)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to get client by slug: %w", err)
	}

	if !client.ClientID.Valid {
		return uuid.Nil, errors.New("client not found")
	}

	return client.ClientID.Bytes, nil
}

// GetClientBySlug slugからclient情報を取得（キャッシュ対応のため、将来的に拡張可能）
func GetClientBySlug(ctx context.Context, slug string, clientRepo repository.ClientRepository) (db.Client, error) {
	return clientRepo.GetBySlug(ctx, slug)
}

// ExtractClientID リクエストからclient_id（UUID）を抽出（優先順位順）
func ExtractClientID(r *http.Request, cfg *config.Config, clientRepo repository.ClientRepository) (uuid.UUID, error) {
	ctx := r.Context()

	// 優先1: サブドメイン（Hostヘッダー）から取得
	host := r.Host
	if host != "" {
		// サブドメインの検証
		if err := ValidateSubdomain(host, cfg); err == nil {
			slug, err := ExtractSlugFromHost(host, cfg.BaseDomain)
			if err == nil && slug != "" {
				clientID, err := GetClientIDBySlug(ctx, slug, clientRepo)
				if err == nil {
					return clientID, nil
				}
			}
		}
	}

	// 優先2: URLパラメータから取得（開発環境のみ）
	if cfg.AppEnv == "development" {
		if clientIDStr := chi.URLParam(r, "client_id"); clientIDStr != "" {
			clientID, err := uuid.Parse(clientIDStr)
			if err == nil {
				return clientID, nil
			}
		}
	}

	// 優先3: ヘッダーから取得（開発環境のみ）
	if cfg.AppEnv == "development" {
		if clientIDStr := r.Header.Get("X-Client-ID"); clientIDStr != "" {
			clientID, err := uuid.Parse(clientIDStr)
			if err == nil {
				return clientID, nil
			}
		}
	}

	return uuid.Nil, errors.New("client_id not found")
}

// GetClientIDFromContext コンテキストからclient_idを取得
func GetClientIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	clientID, ok := ctx.Value(clientIDContextKey).(uuid.UUID)
	return clientID, ok
}

// GetClientFromContext コンテキストからclient情報を取得
func GetClientFromContext(ctx context.Context) (db.Client, bool) {
	client, ok := ctx.Value(clientContextKey).(db.Client)
	return client, ok
}

// TenantMiddleware テナント（クライアント）検証ミドルウェア
func TenantMiddleware(cfg *config.Config, clientRepo repository.ClientRepository, authUsecase usecase.AuthUsecase) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// リクエストからclient_idを抽出
			clientID, err := ExtractClientID(r, cfg, clientRepo)
			if err != nil {
				http.Error(w, "Client ID required", http.StatusBadRequest)
				return
			}

			// client_idの存在確認と有効性チェック
			client, err := clientRepo.GetByID(r.Context(), clientID)
			if err != nil {
				http.Error(w, "Invalid client", http.StatusForbidden)
				return
			}

			// クライアントのステータスチェック（ACTIVEのみ許可）
			if client.Status != "ACTIVE" {
				http.Error(w, "Client is not active", http.StatusForbidden)
				return
			}

			// ユーザーのクライアントアクセス権限検証
			userCtx, ok := GetEnhancedUserContext(r.Context())
			if ok {
				// authUsecaseのValidateClientAccessを呼び出す
				if err := authUsecase.ValidateClientAccess(r.Context(), userCtx, clientID); err != nil {
					http.Error(w, "Client access denied", http.StatusForbidden)
					return
				}
			}

			// コンテキストにclient_idとclient情報を追加
			ctx := context.WithValue(r.Context(), clientIDContextKey, clientID)
			ctx = context.WithValue(ctx, clientContextKey, client)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
