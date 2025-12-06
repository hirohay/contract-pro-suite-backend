package middleware

import (
	"context"
	"net/http"

	"contract-pro-suite/services/auth/domain"
	"contract-pro-suite/services/auth/usecase"
)

// EnhancedAuthMiddleware JWT検証とユーザー情報取得を統合したミドルウェア
func EnhancedAuthMiddleware(authUsecase usecase.AuthUsecase) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 基本のJWT検証ミドルウェアで設定されたUserContextを取得
			jwtUserCtx, ok := GetUserContext(r.Context())
			if !ok {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// データベースからユーザー情報と権限を取得
			userCtx, err := authUsecase.GetUserContext(r.Context(), jwtUserCtx.UserID)
			if err != nil {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			// 拡張されたユーザーコンテキストを設定
			ctx := context.WithValue(r.Context(), enhancedUserContextKey, userCtx)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

type enhancedContextKey string

const enhancedUserContextKey enhancedContextKey = "enhanced_user"

// GetEnhancedUserContext 拡張されたユーザーコンテキストを取得
func GetEnhancedUserContext(ctx context.Context) (*domain.UserContext, bool) {
	userCtx, ok := ctx.Value(enhancedUserContextKey).(*domain.UserContext)
	return userCtx, ok
}

