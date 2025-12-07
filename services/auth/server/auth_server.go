package server

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"contract-pro-suite/internal/interceptor"
	pbauth "contract-pro-suite/proto/proto/auth"
	"contract-pro-suite/services/auth/usecase"
)

// AuthServer 認証gRPCサーバー
type AuthServer struct {
	pbauth.UnimplementedAuthServiceServer
	authUsecase usecase.AuthUsecase
}

// NewAuthServer 認証gRPCサーバーを作成
func NewAuthServer(authUsecase usecase.AuthUsecase) *AuthServer {
	return &AuthServer{
		authUsecase: authUsecase,
	}
}

// GetMe 現在のユーザー情報を取得
func (s *AuthServer) GetMe(ctx context.Context, req *pbauth.GetMeRequest) (*pbauth.GetMeResponse, error) {
	// 拡張されたユーザーコンテキストを取得
	// interceptor.GetEnhancedUserContextは内部でcontextKey型を使用しているため、
	// 同じキー型で設定されたコンテキストから値を取得できる
	userCtx, ok := interceptor.GetEnhancedUserContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "unauthorized")
	}

	// レスポンスを作成
	resp := &pbauth.GetMeResponse{
		UserId:   userCtx.UserID.String(),
		Email:    userCtx.Email,
		UserType: string(userCtx.UserType),
	}

	// client_idが設定されている場合は追加
	if userCtx.ClientID.String() != "00000000-0000-0000-0000-000000000000" {
		clientIDStr := userCtx.ClientID.String()
		resp.ClientId = &clientIDStr
	}

	return resp, nil
}

