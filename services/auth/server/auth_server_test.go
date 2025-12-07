package server

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"contract-pro-suite/internal/interceptor"
	"contract-pro-suite/services/auth/domain"
	pbauth "contract-pro-suite/proto/proto/auth"
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


