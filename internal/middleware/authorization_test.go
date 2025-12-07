package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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

func TestExtractPermissionFromRoute(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		path           string
		wantFeature    string
		wantAction     string
	}{
		{
			name:        "GET request",
			method:      "GET",
			path:        "/api/v1/contracts",
			wantFeature: "contracts",
			wantAction:  "READ",
		},
		{
			name:        "POST request",
			method:      "POST",
			path:        "/api/v1/contracts",
			wantFeature: "contracts",
			wantAction:  "WRITE",
		},
		{
			name:        "PUT request",
			method:      "PUT",
			path:        "/api/v1/contracts/123",
			wantFeature: "contracts",
			wantAction:  "WRITE",
		},
		{
			name:        "DELETE request",
			method:      "DELETE",
			path:        "/api/v1/contracts/123",
			wantFeature: "contracts",
			wantAction:  "DELETE",
		},
		{
			name:        "PATCH request",
			method:      "PATCH",
			path:        "/api/v1/contracts/123",
			wantFeature: "contracts",
			wantAction:  "WRITE",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			feature, action := ExtractPermissionFromRoute(tt.method, tt.path)
			assert.Equal(t, tt.wantFeature, feature)
			assert.Equal(t, tt.wantAction, action)
		})
	}
}

func TestRequirePermission(t *testing.T) {
	mockUsecase := new(MockAuthUsecase)
	userCtx := &domain.UserContext{
		UserID:   uuid.New(),
		UserType: domain.UserTypeOperator,
		Email:    "test@example.com",
	}

	tests := []struct {
		name           string
		setupMock      func()
		expectedStatus int
	}{
		{
			name: "permission granted",
			setupMock: func() {
				mockUsecase.On("CheckPermission", mock.Anything, mock.Anything, "contracts", "READ").Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "permission denied",
			setupMock: func() {
				mockUsecase.On("CheckPermission", mock.Anything, mock.Anything, "contracts", "READ").Return(assert.AnError)
			},
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUsecase.ExpectedCalls = nil
			mockUsecase.Calls = nil
			tt.setupMock()

			handler := RequirePermission(mockUsecase, "contracts", "READ")
			req := httptest.NewRequest("GET", "/", nil)
			ctx := context.WithValue(req.Context(), enhancedUserContextKey, userCtx)
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()
			handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})).ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}

func TestRequireUserType(t *testing.T) {
	tests := []struct {
		name           string
		userType       domain.UserType
		allowedTypes   []domain.UserType
		expectedStatus int
	}{
		{
			name:           "operator allowed",
			userType:       domain.UserTypeOperator,
			allowedTypes:   []domain.UserType{domain.UserTypeOperator},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "client user not allowed",
			userType:       domain.UserTypeClientUser,
			allowedTypes:   []domain.UserType{domain.UserTypeOperator},
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userCtx := &domain.UserContext{
				UserID:   uuid.New(),
				UserType: tt.userType,
				Email:    "test@example.com",
			}

			handler := RequireUserType(tt.allowedTypes...)
			req := httptest.NewRequest("GET", "/", nil)
			ctx := context.WithValue(req.Context(), enhancedUserContextKey, userCtx)
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()
			handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})).ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}

