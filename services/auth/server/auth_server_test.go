package server

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"contract-pro-suite/internal/interceptor"
	pbauth "contract-pro-suite/proto/proto/auth"
	"contract-pro-suite/services/auth/domain"
	"contract-pro-suite/services/auth/usecase"
)

// stringPtr 文字列ポインタを生成するヘルパー関数
func stringPtr(s string) *string {
	return &s
}

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

func TestAuthServer_GetMe(t *testing.T) {
	tests := []struct {
		name           string
		userCtx        *domain.UserContext
		userCtxExists  bool
		expectedStatus codes.Code
		expectedError  bool
	}{
		{
			name: "成功: オペレーター",
			userCtx: &domain.UserContext{
				UserID:   uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
				UserType: domain.UserTypeOperator,
				Email:    "operator@example.com",
				ClientID: uuid.MustParse("00000000-0000-0000-0000-000000000000"),
			},
			userCtxExists:  true,
			expectedStatus: codes.OK,
			expectedError:  false,
		},
		{
			name: "成功: クライアントユーザー（client_idあり）",
			userCtx: &domain.UserContext{
				UserID:   uuid.MustParse("123e4567-e89b-12d3-a456-426614174001"),
				UserType: domain.UserTypeClientUser,
				Email:    "client@example.com",
				ClientID: uuid.MustParse("123e4567-e89b-12d3-a456-426614174002"),
			},
			userCtxExists:  true,
			expectedStatus: codes.OK,
			expectedError:  false,
		},
		{
			name:           "失敗: ユーザーコンテキストなし",
			userCtx:        nil,
			userCtxExists:  false,
			expectedStatus: codes.Unauthenticated,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// モックの準備
			mockUsecase := new(MockAuthUsecase)
			authServer := NewAuthServer(mockUsecase)

			// コンテキストの準備
			ctx := context.Background()
			if tt.userCtxExists {
				ctx = interceptor.SetEnhancedUserContextForTest(ctx, tt.userCtx)
			}

			// GetMeを呼び出し
			req := &pbauth.GetMeRequest{}
			resp, err := authServer.GetMe(ctx, req)

			// エラーチェック
			if tt.expectedError {
				assert.Error(t, err)
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.expectedStatus, st.Code())
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, tt.userCtx.UserID.String(), resp.UserId)
				assert.Equal(t, tt.userCtx.Email, resp.Email)
				assert.Equal(t, string(tt.userCtx.UserType), resp.UserType)

				// client_idのチェック
				if tt.userCtx.ClientID.String() != "00000000-0000-0000-0000-000000000000" {
					assert.NotNil(t, resp.ClientId)
					assert.Equal(t, tt.userCtx.ClientID.String(), *resp.ClientId)
				} else {
					// client_idがnilの場合は、resp.ClientIdもnilまたは空
					if resp.ClientId != nil {
						assert.Equal(t, "", *resp.ClientId)
					}
				}
			}
		})
	}
}

func TestAuthServer_SignupClient(t *testing.T) {
	tests := []struct {
		name           string
		req            *pbauth.SignupClientRequest
		mockResult     *usecase.SignupClientResult
		mockError      error
		expectedStatus codes.Code
		expectedError  bool
	}{
		{
			name: "成功: クライアント登録と管理者ユーザー作成",
			req: &pbauth.SignupClientRequest{
				Name:           "Test Client",
				CompanyCode:    stringPtr("TEST001"),
				Slug:           "test-client",
				AdminEmail:     "admin@test.com",
				AdminPassword:  "Password123!",
				AdminFirstName: "Admin",
				AdminLastName:  "User",
			},
			mockResult: &usecase.SignupClientResult{
				ClientID:    uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
				ClientName:  "Test Client",
				AdminUserID: uuid.MustParse("123e4567-e89b-12d3-a456-426614174001"),
				AdminEmail:  "admin@test.com",
			},
			mockError:      nil,
			expectedStatus: codes.OK,
			expectedError:  false,
		},
		{
			name: "失敗: 必須フィールド不足（name）",
			req: &pbauth.SignupClientRequest{
				CompanyCode:    stringPtr("TEST001"),
				Slug:           "test-client",
				AdminEmail:     "admin@test.com",
				AdminPassword:  "Password123!",
				AdminFirstName: "Admin",
				AdminLastName:  "User",
			},
			mockResult:     nil,
			mockError:      nil,
			expectedStatus: codes.InvalidArgument,
			expectedError:  true,
		},
		{
			name: "失敗: slug重複",
			req: &pbauth.SignupClientRequest{
				Name:           "Test Client",
				CompanyCode:    stringPtr("TEST001"),
				Slug:           "existing-slug",
				AdminEmail:     "admin@test.com",
				AdminPassword:  "Password123!",
				AdminFirstName: "Admin",
				AdminLastName:  "User",
			},
			mockResult:     nil,
			mockError:      fmt.Errorf("slug already exists: existing-slug"),
			expectedStatus: codes.AlreadyExists,
			expectedError:  true,
		},
		{
			name: "失敗: company_code重複",
			req: &pbauth.SignupClientRequest{
				Name:           "Test Client",
				CompanyCode:    stringPtr("EXISTING"),
				Slug:           "test-client",
				AdminEmail:     "admin@test.com",
				AdminPassword:  "Password123!",
				AdminFirstName: "Admin",
				AdminLastName:  "User",
			},
			mockResult:     nil,
			mockError:      fmt.Errorf("company_code already exists: EXISTING"),
			expectedStatus: codes.AlreadyExists,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// モックの準備
			mockUsecase := new(MockAuthUsecase)
			authServer := NewAuthServer(mockUsecase)

			// 成功ケースの場合のみモックを設定
			if !tt.expectedError && tt.mockResult != nil {
				mockUsecase.On("SignupClient", mock.Anything, mock.Anything).Return(tt.mockResult, tt.mockError)
			} else if tt.expectedError && tt.mockError != nil {
				mockUsecase.On("SignupClient", mock.Anything, mock.Anything).Return(nil, tt.mockError)
			}

			// SignupClientを呼び出し
			ctx := context.Background()
			resp, err := authServer.SignupClient(ctx, tt.req)

			// エラーチェック
			if tt.expectedError {
				assert.Error(t, err)
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.expectedStatus, st.Code())
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, tt.mockResult.ClientID.String(), resp.ClientId)
				assert.Equal(t, tt.mockResult.ClientName, resp.ClientName)
				assert.Equal(t, tt.mockResult.AdminUserID.String(), resp.AdminUserId)
				assert.Equal(t, tt.mockResult.AdminEmail, resp.AdminEmail)
			}

			mockUsecase.AssertExpectations(t)
		})
	}
}
