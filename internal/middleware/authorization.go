package middleware

import (
	"net/http"
	"strings"

	"contract-pro-suite/services/auth/domain"
	"contract-pro-suite/services/auth/usecase"
)

// RequirePermission 指定された権限を要求するミドルウェア
func RequirePermission(authUsecase usecase.AuthUsecase, feature, action string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 拡張されたユーザーコンテキストを取得
			userCtx, ok := GetEnhancedUserContext(r.Context())
			if !ok {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// 権限チェック
			if err := authUsecase.CheckPermission(r.Context(), userCtx, feature, action); err != nil {
				http.Error(w, "Forbidden: insufficient permissions", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r.WithContext(r.Context()))
		})
	}
}

// RequireUserType 指定されたユーザータイプを要求するミドルウェア
func RequireUserType(userTypes ...domain.UserType) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 拡張されたユーザーコンテキストを取得
			userCtx, ok := GetEnhancedUserContext(r.Context())
			if !ok {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// ユーザータイプのチェック
			allowed := false
			for _, userType := range userTypes {
				if userCtx.UserType == userType {
					allowed = true
					break
				}
			}

			if !allowed {
				http.Error(w, "Forbidden: user type not allowed", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r.WithContext(r.Context()))
		})
	}
}

// RequireOperatorRole オペレーターのロールを要求するミドルウェア（将来の実装用）
// 現時点ではoperator_assignmentsテーブルが未実装のため、簡易実装
func RequireOperatorRole(authUsecase usecase.AuthUsecase, roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 拡張されたユーザーコンテキストを取得
			userCtx, ok := GetEnhancedUserContext(r.Context())
			if !ok {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// オペレーターのみ許可
			if userCtx.UserType != domain.UserTypeOperator {
				http.Error(w, "Forbidden: operator required", http.StatusForbidden)
				return
			}

			// 将来の実装: operator_assignmentsテーブルからロールを取得してチェック
			// 現時点では一旦許可（実装後に詳細なチェックを追加）
			_ = roles // 未使用変数の警告を回避

			next.ServeHTTP(w, r.WithContext(r.Context()))
		})
	}
}

// PermissionConfig 権限設定（ルートごとに設定）
type PermissionConfig struct {
	Feature string
	Action  string
}

// RequirePermissions 複数の権限を要求するミドルウェア（いずれか一つでも許可されていればOK）
func RequirePermissions(authUsecase usecase.AuthUsecase, permissions ...PermissionConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 拡張されたユーザーコンテキストを取得
			userCtx, ok := GetEnhancedUserContext(r.Context())
			if !ok {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// いずれかの権限があれば許可
			hasPermission := false
			for _, perm := range permissions {
				if err := authUsecase.CheckPermission(r.Context(), userCtx, perm.Feature, perm.Action); err == nil {
					hasPermission = true
					break
				}
			}

			if !hasPermission {
				http.Error(w, "Forbidden: insufficient permissions", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r.WithContext(r.Context()))
		})
	}
}

// RequireAllPermissions 複数の権限を要求するミドルウェア（すべての権限が必要）
func RequireAllPermissions(authUsecase usecase.AuthUsecase, permissions ...PermissionConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 拡張されたユーザーコンテキストを取得
			userCtx, ok := GetEnhancedUserContext(r.Context())
			if !ok {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// すべての権限が必要
			for _, perm := range permissions {
				if err := authUsecase.CheckPermission(r.Context(), userCtx, perm.Feature, perm.Action); err != nil {
					http.Error(w, "Forbidden: insufficient permissions", http.StatusForbidden)
					return
				}
			}

			next.ServeHTTP(w, r.WithContext(r.Context()))
		})
	}
}

// ExtractPermissionFromRoute ルートから権限情報を抽出（将来の実装用）
// 例: "/api/v1/contracts" → feature: "contracts", action: "READ" (GETの場合)
func ExtractPermissionFromRoute(method, path string) (feature, action string) {
	// パスから機能名を抽出
	parts := strings.Split(strings.TrimPrefix(path, "/api/v1/"), "/")
	if len(parts) > 0 {
		feature = parts[0]
	}

	// HTTPメソッドからアクションを決定
	switch method {
	case "GET":
		action = "READ"
	case "POST":
		action = "WRITE"
	case "PUT", "PATCH":
		action = "WRITE"
	case "DELETE":
		action = "DELETE"
	default:
		action = "READ"
	}

	return feature, action
}
