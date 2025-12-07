package interceptor

import (
	"context"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"contract-pro-suite/internal/shared/config"
	"contract-pro-suite/services/auth/domain"
	"contract-pro-suite/services/auth/repository"
	"contract-pro-suite/services/auth/usecase"
	"github.com/jackc/pgx/v5/pgtype"
	dbgen "contract-pro-suite/sqlc"
)

// MockAuthUsecase モックAuthUsecase
type MockAuthUsecase struct {
	mock.Mock
}

func (m *MockAuthUsecase) GetUserContext(ctx context.Context, jwtUserID string) (*domain.UserContext, error) {
	args := m.Called(ctx, jwtUserID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.UserContext), args.Error(1)
}

func (m *MockAuthUsecase) ValidateClientAccess(ctx context.Context, userCtx *domain.UserContext, clientID uuid.UUID) error {
	args := m.Called(ctx, userCtx, clientID)
	return args.Error(0)
}

func (m *MockAuthUsecase) CheckPermission(ctx context.Context, userCtx *domain.UserContext, feature, action string) error {
	args := m.Called(ctx, userCtx, feature, action)
	return args.Error(0)
}

func (m *MockAuthUsecase) SignupClient(ctx context.Context, params usecase.SignupClientParams) (*usecase.SignupClientResult, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*usecase.SignupClientResult), args.Error(1)
}

// MockClientRepository モックClientRepository
type MockClientRepository struct {
	mock.Mock
}

func (m *MockClientRepository) GetByID(ctx context.Context, clientID uuid.UUID) (dbgen.Client, error) {
	args := m.Called(ctx, clientID)
	if args.Get(0) == nil {
		return dbgen.Client{}, args.Error(1)
	}
	return args.Get(0).(dbgen.Client), args.Error(1)
}

func (m *MockClientRepository) GetBySlug(ctx context.Context, slug string) (dbgen.Client, error) {
	args := m.Called(ctx, slug)
	if args.Get(0) == nil {
		return dbgen.Client{}, args.Error(1)
	}
	return args.Get(0).(dbgen.Client), args.Error(1)
}

func (m *MockClientRepository) GetByCompanyCode(ctx context.Context, companyCode string) (dbgen.Client, error) {
	args := m.Called(ctx, companyCode)
	if args.Get(0) == nil {
		return dbgen.Client{}, args.Error(1)
	}
	return args.Get(0).(dbgen.Client), args.Error(1)
}

func (m *MockClientRepository) List(ctx context.Context, limit, offset int32) ([]dbgen.Client, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]dbgen.Client), args.Error(1)
}

func (m *MockClientRepository) Create(ctx context.Context, params dbgen.CreateClientParams) (dbgen.Client, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return dbgen.Client{}, args.Error(1)
	}
	return args.Get(0).(dbgen.Client), args.Error(1)
}

func (m *MockClientRepository) Update(ctx context.Context, params dbgen.UpdateClientParams) (dbgen.Client, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return dbgen.Client{}, args.Error(1)
	}
	return args.Get(0).(dbgen.Client), args.Error(1)
}

func (m *MockClientRepository) Delete(ctx context.Context, clientID uuid.UUID, deletedBy uuid.UUID) error {
	args := m.Called(ctx, clientID, deletedBy)
	return args.Error(0)
}

// createTestJWT テスト用のJWTトークンを作成
func createTestJWT(secret string, claims jwt.MapClaims) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(secret))
	return tokenString
}

func TestAuthInterceptor(t *testing.T) {
	tests := []struct {
		name           string
		authHeader     string
		jwtSecret      string
		claims         jwt.MapClaims
		expectedStatus codes.Code
		expectedError  bool
	}{
		{
			name:       "成功: 有効なJWTトークン",
			authHeader: "Bearer ",
			jwtSecret:  "test-secret",
			claims: jwt.MapClaims{
				"sub":   "123e4567-e89b-12d3-a456-426614174000",
				"email": "test@example.com",
				"role":  "authenticated",
			},
			expectedStatus: codes.OK,
			expectedError:  false,
		},
		{
			name:           "失敗: Authorizationヘッダーなし",
			authHeader:     "",
			jwtSecret:      "test-secret",
			claims:         nil,
			expectedStatus: codes.Unauthenticated,
			expectedError:  true,
		},
		{
			name:           "失敗: 無効な形式",
			authHeader:     "Invalid",
			jwtSecret:      "test-secret",
			claims:         nil,
			expectedStatus: codes.Unauthenticated,
			expectedError:  true,
		},
		{
			name:           "失敗: 無効なJWTトークン",
			authHeader:     "Bearer invalid-token",
			jwtSecret:      "test-secret",
			claims:         nil,
			expectedStatus: codes.Unauthenticated,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				SupabaseJWTSecret: tt.jwtSecret,
			}

			// JWTトークンを作成
			var tokenString string
			if tt.claims != nil {
				tokenString = createTestJWT(tt.jwtSecret, tt.claims)
				tt.authHeader = "Bearer " + tokenString
			}

			// メタデータを作成
			md := metadata.New(map[string]string{})
			if tt.authHeader != "" {
				md.Set("authorization", tt.authHeader)
			}
			ctx := metadata.NewIncomingContext(context.Background(), md)

			// インターセプターを適用
			interceptor := AuthInterceptor(cfg)
			handler := func(ctx context.Context, req interface{}) (interface{}, error) {
				// コンテキストからユーザー情報を取得
				userCtx, ok := GetUserContext(ctx)
				if !ok {
					return nil, status.Errorf(codes.Internal, "user context not found")
				}
				return userCtx, nil
			}

			resp, err := interceptor(ctx, nil, &grpc.UnaryServerInfo{
				FullMethod: "/test.Test/Test",
			}, handler)

			if tt.expectedError {
				assert.Error(t, err)
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.expectedStatus, st.Code())
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				userCtx, ok := resp.(*UserContext)
				assert.True(t, ok)
				if tt.claims != nil {
					assert.Equal(t, tt.claims["sub"], userCtx.UserID)
					assert.Equal(t, tt.claims["email"], userCtx.Email)
				}
			}
		})
	}
}

func TestEnhancedAuthInterceptor(t *testing.T) {
	tests := []struct {
		name           string
		userCtxExists  bool
		userCtx        *UserContext
		expectedUserCtx *domain.UserContext
		expectedError  error
		expectedStatus codes.Code
	}{
		{
			name:          "成功: ユーザーコンテキスト取得",
			userCtxExists: true,
			userCtx: &UserContext{
				UserID: "123e4567-e89b-12d3-a456-426614174000",
				Email:  "test@example.com",
				Role:   "authenticated",
			},
			expectedUserCtx: &domain.UserContext{
				UserID:   uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
				UserType: domain.UserTypeOperator,
				Email:    "test@example.com",
			},
			expectedError:  nil,
			expectedStatus: codes.OK,
		},
		{
			name:           "失敗: ユーザーコンテキストなし",
			userCtxExists:  false,
			userCtx:        nil,
			expectedUserCtx: nil,
			expectedError:   nil,
			expectedStatus: codes.Unauthenticated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUsecase := new(MockAuthUsecase)
			ctx := context.Background()

			if tt.userCtxExists {
				ctx = context.WithValue(ctx, userContextKey, tt.userCtx)
				if tt.expectedUserCtx != nil {
					mockUsecase.On("GetUserContext", ctx, tt.userCtx.UserID).Return(tt.expectedUserCtx, tt.expectedError)
				}
			}

			interceptor := EnhancedAuthInterceptor(mockUsecase)
			handler := func(ctx context.Context, req interface{}) (interface{}, error) {
				userCtx, ok := GetEnhancedUserContext(ctx)
				if !ok {
					return nil, status.Errorf(codes.Internal, "enhanced user context not found")
				}
				return userCtx, nil
			}

			resp, err := interceptor(ctx, nil, &grpc.UnaryServerInfo{
				FullMethod: "/test.Test/Test",
			}, handler)

			if tt.expectedStatus != codes.OK {
				assert.Error(t, err)
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.expectedStatus, st.Code())
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
			}

			mockUsecase.AssertExpectations(t)
		})
	}
}

func TestTenantInterceptor(t *testing.T) {
	tests := []struct {
		name           string
		clientIDHeader string
		clientID       uuid.UUID
		clientStatus   string
		expectedStatus codes.Code
		expectedError  bool
	}{
		{
			name:           "成功: 有効なクライアントID",
			clientIDHeader: "123e4567-e89b-12d3-a456-426614174000",
			clientID:       uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
			clientStatus:   "ACTIVE",
			expectedStatus: codes.OK,
			expectedError:  false,
		},
		{
			name:           "失敗: 無効なクライアントID形式",
			clientIDHeader: "invalid-uuid",
			clientID:       uuid.Nil,
			clientStatus:   "",
			expectedStatus: codes.InvalidArgument,
			expectedError:  true,
		},
		{
			name:           "失敗: クライアントが非アクティブ",
			clientIDHeader: "123e4567-e89b-12d3-a456-426614174000",
			clientID:       uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
			clientStatus:   "INACTIVE",
			expectedStatus: codes.PermissionDenied,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				DefaultClientID: "00000000-0000-0000-0000-000000000000",
			}

			mockClientRepo := new(MockClientRepository)
			mockUsecase := new(MockAuthUsecase)

			// メタデータを作成
			md := metadata.New(map[string]string{})
			if tt.clientIDHeader != "" {
				md.Set("x-client-id", tt.clientIDHeader)
			}
			ctx := metadata.NewIncomingContext(context.Background(), md)

			// クライアント情報をモック
			if tt.clientID != uuid.Nil {
				client := dbgen.Client{
					ClientID: pgtype.UUID{Bytes: tt.clientID, Valid: true},
					Status:   tt.clientStatus,
				}
				mockClientRepo.On("GetByID", ctx, tt.clientID).Return(client, nil)
			}

			// 型アサーションでインポートを使用していることを示す
			var _ repository.ClientRepository = mockClientRepo
			var _ usecase.AuthUsecase = mockUsecase

			interceptor := TenantInterceptor(cfg, mockClientRepo, mockUsecase)
			handler := func(ctx context.Context, req interface{}) (interface{}, error) {
				return "success", nil
			}

			resp, err := interceptor(ctx, nil, &grpc.UnaryServerInfo{
				FullMethod: "/test.Test/Test",
			}, handler)

			if tt.expectedError {
				assert.Error(t, err)
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.expectedStatus, st.Code())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, "success", resp)
			}

			mockClientRepo.AssertExpectations(t)
		})
	}
}

