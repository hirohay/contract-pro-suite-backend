package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"contract-pro-suite/services/auth/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestEnhancedAuthMiddleware(t *testing.T) {
	mockUsecase := new(MockAuthUsecase)
	testUserID := uuid.New()
	userCtx := &domain.UserContext{
		UserID:   testUserID,
		UserType: domain.UserTypeOperator,
		Email:    "test@example.com",
	}

	tests := []struct {
		name           string
		setupMock      func()
		hasJWTContext  bool
		expectedStatus int
	}{
		{
			name: "successful user context retrieval",
			setupMock: func() {
				mockUsecase.On("GetUserContext", mock.Anything, "test-user-id").Return(userCtx, nil)
			},
			hasJWTContext:  true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "no JWT context",
			setupMock:      func() {},
			hasJWTContext:  false,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "user not found",
			setupMock: func() {
				mockUsecase.On("GetUserContext", mock.Anything, "test-user-id").Return(nil, assert.AnError)
			},
			hasJWTContext:  true,
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUsecase.ExpectedCalls = nil
			mockUsecase.Calls = nil
			tt.setupMock()

			handler := EnhancedAuthMiddleware(mockUsecase)
			req := httptest.NewRequest("GET", "/", nil)

			ctx := req.Context()
			if tt.hasJWTContext {
				ctx = context.WithValue(ctx, userContextKey, &UserContext{
					UserID: "test-user-id",
					Email:  "test@example.com",
					Role:   "authenticated",
				})
			}
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()
			handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})).ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}

func TestGetEnhancedUserContext(t *testing.T) {
	testUserID := uuid.New()
	userCtx := &domain.UserContext{
		UserID:   testUserID,
		UserType: domain.UserTypeOperator,
		Email:    "test@example.com",
	}

	tests := []struct {
		name   string
		ctx    context.Context
		wantOk bool
		wantID uuid.UUID
	}{
		{
			name: "enhanced user context exists",
			ctx: context.WithValue(
				context.Background(),
				enhancedUserContextKey,
				userCtx,
			),
			wantOk: true,
			wantID: testUserID,
		},
		{
			name:   "no enhanced user context",
			ctx:    context.Background(),
			wantOk: false,
			wantID: uuid.Nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := GetEnhancedUserContext(tt.ctx)
			assert.Equal(t, tt.wantOk, ok)
			if tt.wantOk {
				assert.Equal(t, tt.wantID, got.UserID)
			}
		})
	}
}
