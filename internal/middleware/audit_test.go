package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"contract-pro-suite/services/auth/domain"
)

func TestAuditMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		path           string
		statusCode     int
		hasUserContext bool
		hasClientID    bool
	}{
		{
			name:           "successful request with user context",
			method:         "GET",
			path:           "/api/v1/test",
			statusCode:     http.StatusOK,
			hasUserContext: true,
			hasClientID:    true,
		},
		{
			name:           "error request",
			method:         "POST",
			path:           "/api/v1/test",
			statusCode:     http.StatusBadRequest,
			hasUserContext: false,
			hasClientID:    false,
		},
		{
			name:           "request without user context",
			method:         "GET",
			path:           "/health",
			statusCode:     http.StatusOK,
			hasUserContext: false,
			hasClientID:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := AuditMiddleware()
			req := httptest.NewRequest(tt.method, tt.path, nil)

			ctx := req.Context()
			if tt.hasUserContext {
				userCtx := &domain.UserContext{
					UserID:   uuid.New(),
					UserType: domain.UserTypeOperator,
					Email:    "test@example.com",
				}
				if tt.hasClientID {
					userCtx.ClientID = uuid.New()
				}
				ctx = context.WithValue(ctx, enhancedUserContextKey, userCtx)
			}
			if tt.hasClientID {
				clientID := uuid.New()
				ctx = context.WithValue(ctx, clientIDContextKey, clientID)
			}
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()
			handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
			})).ServeHTTP(rr, req)

			assert.Equal(t, tt.statusCode, rr.Code)
		})
	}
}

func TestResponseWriter(t *testing.T) {
	rr := httptest.NewRecorder()
	rw := &responseWriter{
		ResponseWriter: rr,
		statusCode:     http.StatusOK,
	}

	rw.WriteHeader(http.StatusCreated)
	assert.Equal(t, http.StatusCreated, rw.statusCode)
	assert.Equal(t, http.StatusCreated, rr.Code)
}

func TestAuditLog(t *testing.T) {
	log := AuditLog{
		Timestamp:  time.Now(),
		UserID:     "123e4567-e89b-12d3-a456-426614174000",
		ClientID:   "223e4567-e89b-12d3-a456-426614174000",
		Method:     "GET",
		Path:       "/api/v1/test",
		StatusCode: http.StatusOK,
		UserType:   "OPERATOR",
	}

	// 構造体が正しく定義されているか確認
	assert.Equal(t, "GET", log.Method)
	assert.Equal(t, "/api/v1/test", log.Path)
	assert.Equal(t, http.StatusOK, log.StatusCode)
}

