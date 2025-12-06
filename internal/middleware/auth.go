package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"contract-pro-suite/internal/shared/config"
)

type contextKey string

const userContextKey contextKey = "user"

// UserContext JWTから取得したユーザー情報
type UserContext struct {
	UserID string // Supabase AuthのユーザーID（sub）
	Email  string
	Role   string // Supabase Authのロール
}

// AuthMiddleware Supabase JWTトークンを検証するミドルウェア
func AuthMiddleware(cfg *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Authorizationヘッダーからトークンを取得
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Authorization header required", http.StatusUnauthorized)
				return
			}

			// "Bearer <token>"形式からトークンを抽出
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
				return
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
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			// クレームからユーザー情報を取得
			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				http.Error(w, "Invalid token claims", http.StatusUnauthorized)
				return
			}

			// ユーザー情報をコンテキストに追加
			userCtx := &UserContext{
				UserID: getStringClaim(claims, "sub"),
				Email:  getStringClaim(claims, "email"),
				Role:   getStringClaim(claims, "role"),
			}

			ctx := context.WithValue(r.Context(), userContextKey, userCtx)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserContext コンテキストからユーザー情報を取得
func GetUserContext(ctx context.Context) (*UserContext, bool) {
	userCtx, ok := ctx.Value(userContextKey).(*UserContext)
	return userCtx, ok
}

// getStringClaim クレームから文字列値を取得
func getStringClaim(claims jwt.MapClaims, key string) string {
	if val, ok := claims[key].(string); ok {
		return val
	}
	return ""
}

