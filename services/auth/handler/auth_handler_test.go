package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"contract-pro-suite/internal/middleware"
	"contract-pro-suite/services/auth/domain"
	"contract-pro-suite/services/auth/usecase"
)

// MockAuthUsecase モック認証ユースケース
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

func TestNewAuthHandler(t *testing.T) {
	mockUsecase := new(MockAuthUsecase)
	handler := NewAuthHandler(mockUsecase)

	assert.NotNil(t, handler)
	assert.Equal(t, mockUsecase, handler.authUsecase)
}

func TestGetMe(t *testing.T) {
	mockUsecase := new(MockAuthUsecase)
	handler := NewAuthHandler(mockUsecase)

	testUserID := uuid.New()
	testClientID := uuid.New()
	userCtx := &domain.UserContext{
		UserID:   testUserID,
		UserType: domain.UserTypeOperator,
		Email:    "test@example.com",
		ClientID: testClientID,
	}

	tests := []struct {
		name           string
		hasUserContext bool
		userCtx        *domain.UserContext
		expectedStatus int
		expectedBody   func() map[string]interface{}
	}{
		{
			name:           "successful response with client_id",
			hasUserContext: true,
			userCtx:        userCtx,
			expectedStatus: http.StatusOK,
			expectedBody: func() map[string]interface{} {
				return map[string]interface{}{
					"user_id":   testUserID.String(),
					"user_type": "OPERATOR",
					"email":     "test@example.com",
					"client_id": testClientID.String(),
				}
			},
		},
		{
			name:           "successful response without client_id",
			hasUserContext: true,
			userCtx: &domain.UserContext{
				UserID:   testUserID,
				UserType: domain.UserTypeOperator,
				Email:    "test@example.com",
				ClientID: uuid.Nil,
			},
			expectedStatus: http.StatusOK,
			expectedBody: func() map[string]interface{} {
				return map[string]interface{}{
					"user_id":   testUserID.String(),
					"user_type": "OPERATOR",
					"email":     "test@example.com",
				}
			},
		},
		{
			name:           "unauthorized - no user context",
			hasUserContext: false,
			userCtx:        nil,
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/auth/me", nil)

			ctx := req.Context()
			if tt.hasUserContext {
				// middlewareパッケージのテストヘルパー関数を使用してコンテキストを設定
				ctx = middleware.SetEnhancedUserContextForTest(ctx, tt.userCtx)
			}
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()
			handler.GetMe(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.expectedBody != nil {
				var response map[string]interface{}
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				assert.NoError(t, err)

				expected := tt.expectedBody()
				for key, value := range expected {
					assert.Equal(t, value, response[key], "Mismatch for key: %s", key)
				}
			}
		})
	}
}

