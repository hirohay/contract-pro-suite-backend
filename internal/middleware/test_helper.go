package middleware

import (
	"context"
	"contract-pro-suite/services/auth/domain"
)

// SetEnhancedUserContextForTest テスト用に拡張されたユーザーコンテキストを設定
// この関数はテスト専用で、本番コードでは使用しない
func SetEnhancedUserContextForTest(ctx context.Context, userCtx *domain.UserContext) context.Context {
	return context.WithValue(ctx, enhancedUserContextKey, userCtx)
}

