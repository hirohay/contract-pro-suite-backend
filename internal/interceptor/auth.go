package interceptor

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"contract-pro-suite/internal/shared/config"
	"contract-pro-suite/services/auth/domain"
	"contract-pro-suite/services/auth/repository"
	"contract-pro-suite/services/auth/usecase"
)

type contextKey string

const (
	userContextKey         contextKey = "user"
	enhancedUserContextKey contextKey = "enhanced_user"
	clientIDContextKey     contextKey = "client_id"
)

// isPublicMethod 認証不要の公開メソッドかどうかを判定
func isPublicMethod(methodName string) bool {
	publicMethods := []string{
		"/auth.AuthService/SignupClient",
	}
	for _, publicMethod := range publicMethods {
		if methodName == publicMethod {
			return true
		}
	}
	return false
}

// AuthInterceptor JWT検証インターセプター
func AuthInterceptor(cfg *config.Config) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// 認証不要の公開メソッドの場合はスキップ
		if isPublicMethod(info.FullMethod) {
			return handler(ctx, req)
		}

		// メタデータからAuthorizationヘッダーを取得
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Errorf(codes.Unauthenticated, "metadata not found")
		}

		authHeaders := md.Get("authorization")
		if len(authHeaders) == 0 {
			return nil, status.Errorf(codes.Unauthenticated, "authorization header required")
		}

		authHeader := authHeaders[0]
		if authHeader == "" {
			return nil, status.Errorf(codes.Unauthenticated, "authorization header required")
		}

		// "Bearer <token>"形式からトークンを抽出
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return nil, status.Errorf(codes.Unauthenticated, "invalid authorization header format")
		}

		tokenString := parts[1]

		// JWTトークンを検証
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// 署名アルゴリズムを確認
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			// Supabase JWT Secretを使用
			return []byte(cfg.SupabaseJWTSecret), nil
		})

		if err != nil || !token.Valid {
			return nil, status.Errorf(codes.Unauthenticated, "invalid token")
		}

		// クレームからユーザー情報を取得
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return nil, status.Errorf(codes.Unauthenticated, "invalid token claims")
		}

		// ユーザー情報をコンテキストに追加
		userCtx := &UserContext{
			UserID: getStringClaim(claims, "sub"),
			Email:  getStringClaim(claims, "email"),
			Role:   getStringClaim(claims, "role"),
		}

		ctx = context.WithValue(ctx, userContextKey, userCtx)
		return handler(ctx, req)
	}
}

// UserContext JWTから取得したユーザー情報
type UserContext struct {
	UserID string // Supabase AuthのユーザーID（sub）
	Email  string
	Role   string // Supabase Authのロール
}

// GetUserContext コンテキストからユーザー情報を取得
func GetUserContext(ctx context.Context) (*UserContext, bool) {
	userCtx, ok := ctx.Value(userContextKey).(*UserContext)
	return userCtx, ok
}

// EnhancedAuthInterceptor ユーザー情報取得インターセプター
func EnhancedAuthInterceptor(authUsecase usecase.AuthUsecase) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// 認証不要の公開メソッドの場合はスキップ
		if isPublicMethod(info.FullMethod) {
			return handler(ctx, req)
		}

		// 基本のJWT検証インターセプターで設定されたUserContextを取得
		jwtUserCtx, ok := GetUserContext(ctx)
		if !ok {
			return nil, status.Errorf(codes.Unauthenticated, "unauthorized")
		}

		// データベースからユーザー情報と権限を取得
		userCtx, err := authUsecase.GetUserContext(ctx, jwtUserCtx.UserID)
		if err != nil {
			return nil, status.Errorf(codes.PermissionDenied, "forbidden")
		}

		// 拡張されたユーザーコンテキストを設定
		ctx = context.WithValue(ctx, enhancedUserContextKey, userCtx)
		return handler(ctx, req)
	}
}

// GetEnhancedUserContext 拡張されたユーザーコンテキストを取得
func GetEnhancedUserContext(ctx context.Context) (*domain.UserContext, bool) {
	userCtx, ok := ctx.Value(enhancedUserContextKey).(*domain.UserContext)
	return userCtx, ok
}

// SetEnhancedUserContextForTest テスト用: 拡張されたユーザーコンテキストを設定
func SetEnhancedUserContextForTest(ctx context.Context, userCtx *domain.UserContext) context.Context {
	return context.WithValue(ctx, enhancedUserContextKey, userCtx)
}

// TenantInterceptor テナント検証インターセプター
func TenantInterceptor(
	cfg *config.Config,
	clientRepo repository.ClientRepository,
	authUsecase usecase.AuthUsecase,
) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// 認証不要の公開メソッドの場合はスキップ（クライアント登録時はclient_idが存在しないため）
		if isPublicMethod(info.FullMethod) {
			return handler(ctx, req)
		}

		// メタデータからclient_idを取得（優先順位: サブドメイン > x-client-idヘッダー > デフォルト値）
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Errorf(codes.InvalidArgument, "metadata not found")
		}

		// 優先1: サブドメイン（authorityメタデータ）から取得
		clientID, err := extractClientIDFromSubdomain(ctx, md, cfg, clientRepo)
		if err == nil {
			// サブドメインから正常に取得できた場合
		} else {
			// 優先2: x-client-idヘッダーから取得を試みる
			clientIDHeaders := md.Get("x-client-id")
			var clientIDStr string
			if len(clientIDHeaders) > 0 {
				clientIDStr = clientIDHeaders[0]
			}

			if clientIDStr == "" {
				// 優先3: デフォルトのクライアントIDを使用
				clientIDStr = cfg.DefaultClientID
			}

			// UUIDとしてパース
			clientID, err = uuid.Parse(clientIDStr)
			if err != nil {
				return nil, status.Errorf(codes.InvalidArgument, "invalid client_id format")
			}
		}

		// client_idの存在確認と有効性チェック
		client, err := clientRepo.GetByID(ctx, clientID)
		if err != nil {
			return nil, status.Errorf(codes.PermissionDenied, "invalid client")
		}

		// クライアントのステータスチェック（ACTIVEのみ許可）
		if client.Status != "ACTIVE" {
			return nil, status.Errorf(codes.PermissionDenied, "client is not active")
		}

		// コンテキストにclient_idを追加
		ctx = context.WithValue(ctx, clientIDContextKey, clientID)

		// ユーザーのクライアントアクセス権限検証
		userCtx, ok := GetEnhancedUserContext(ctx)
		if ok {
			// authUsecaseのValidateClientAccessを呼び出す
			if err := authUsecase.ValidateClientAccess(ctx, userCtx, clientID); err != nil {
				return nil, status.Errorf(codes.PermissionDenied, "client access denied")
			}
		}

		return handler(ctx, req)
	}
}

// GetClientIDFromContext コンテキストからclient_idを取得
func GetClientIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	clientID, ok := ctx.Value(clientIDContextKey).(uuid.UUID)
	return clientID, ok
}

// ExtractClientIDFromSubdomain gRPCメタデータからサブドメインを抽出し、clientIDを取得（再利用可能な関数）
func ExtractClientIDFromSubdomain(
	ctx context.Context,
	md metadata.MD,
	cfg *config.Config,
	clientRepo repository.ClientRepository,
) (uuid.UUID, error) {
	// gRPCのauthorityメタデータからHost情報を取得
	authorities := md.Get(":authority")
	if len(authorities) == 0 {
		return uuid.Nil, errors.New("authority not found")
	}

	host := authorities[0]
	if host == "" {
		return uuid.Nil, errors.New("empty authority")
	}

	// サブドメインの検証
	if err := ValidateSubdomain(host, cfg); err != nil {
		return uuid.Nil, fmt.Errorf("subdomain validation failed: %w", err)
	}

	// サブドメインからslugを抽出
	slug, err := ExtractSlugFromHost(host, cfg.BaseDomain)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to extract slug: %w", err)
	}

	// slugからclientIDを取得（再利用可能な関数を使用）
	return GetClientIDBySlug(ctx, slug, clientRepo)
}

// GetClientIDBySlug slugからclient_id（UUID）を取得（再利用可能な関数）
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

// ExtractSlugFromHost サブドメインからslugを抽出（再利用可能な関数）
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

// ValidateSubdomain サブドメインが許可されたドメインか検証（再利用可能な関数）
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

// extractClientIDFromSubdomain gRPCメタデータからサブドメインを抽出し、clientIDを取得（内部用、ExtractClientIDFromSubdomainを使用）
func extractClientIDFromSubdomain(
	ctx context.Context,
	md metadata.MD,
	cfg *config.Config,
	clientRepo repository.ClientRepository,
) (uuid.UUID, error) {
	return ExtractClientIDFromSubdomain(ctx, md, cfg, clientRepo)
}

// getStringClaim クレームから文字列値を取得
func getStringClaim(claims jwt.MapClaims, key string) string {
	if val, ok := claims[key].(string); ok {
		return val
	}
	return ""
}
