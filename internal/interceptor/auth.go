package interceptor

import (
	"context"
	"errors"
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

// AuthInterceptor JWT検証インターセプター
func AuthInterceptor(cfg *config.Config) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
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
		// メタデータからclient_idを取得
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Errorf(codes.InvalidArgument, "metadata not found")
		}

		// x-client-idヘッダーから取得を試みる
		clientIDHeaders := md.Get("x-client-id")
		var clientIDStr string
		if len(clientIDHeaders) > 0 {
			clientIDStr = clientIDHeaders[0]
		}

		// TODO: サブドメインからの取得も実装する必要がある
		// 現時点ではx-client-idヘッダーからのみ取得

		if clientIDStr == "" {
			// デフォルトのクライアントIDを使用
			clientIDStr = cfg.DefaultClientID
		}

		// UUIDとしてパース
		clientID, err := uuid.Parse(clientIDStr)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid client_id format")
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

// getStringClaim クレームから文字列値を取得
func getStringClaim(claims jwt.MapClaims, key string) string {
	if val, ok := claims[key].(string); ok {
		return val
	}
	return ""
}

