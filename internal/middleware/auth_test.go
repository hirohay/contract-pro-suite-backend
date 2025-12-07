package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"contract-pro-suite/internal/shared/config"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestAuthMiddleware(t *testing.T) {
	cfg := &config.Config{
		SupabaseJWTSecret: "test-secret-key-for-jwt-signing",
	}

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
		setupToken     func() string
	}{
		{
			name:           "valid token",
			authHeader:     "Bearer ",
			expectedStatus: http.StatusOK,
			setupToken: func() string {
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
					"sub":   "123e4567-e89b-12d3-a456-426614174000",
					"email": "test@example.com",
					"role":  "authenticated",
					"exp":   time.Now().Add(time.Hour).Unix(),
				})
				tokenString, _ := token.SignedString([]byte(cfg.SupabaseJWTSecret))
				return tokenString
			},
		},
		{
			name:           "missing authorization header",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
			setupToken:     func() string { return "" },
		},
		{
			name:           "invalid authorization format",
			authHeader:     "InvalidFormat token",
			expectedStatus: http.StatusUnauthorized,
			setupToken:     func() string { return "" },
		},
		{
			name:           "invalid token",
			authHeader:     "Bearer invalid-token",
			expectedStatus: http.StatusUnauthorized,
			setupToken:     func() string { return "invalid-token" },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenString := tt.setupToken()
			authHeader := tt.authHeader
			if tokenString != "" && authHeader == "Bearer " {
				authHeader = "Bearer " + tokenString
			}

			handler := AuthMiddleware(cfg)
			req := httptest.NewRequest("GET", "/", nil)
			if authHeader != "" {
				req.Header.Set("Authorization", authHeader)
			}

			rr := httptest.NewRecorder()
			handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})).ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}

func TestGetUserContext(t *testing.T) {
	tests := []struct {
		name   string
		ctx    context.Context
		wantOk bool
		wantID string
	}{
		{
			name: "user context exists",
			ctx: context.WithValue(
				context.Background(),
				userContextKey,
				&UserContext{
					UserID: "123e4567-e89b-12d3-a456-426614174000",
					Email:  "test@example.com",
					Role:   "authenticated",
				},
			),
			wantOk: true,
			wantID: "123e4567-e89b-12d3-a456-426614174000",
		},
		{
			name:   "no user context",
			ctx:    context.Background(),
			wantOk: false,
			wantID: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userCtx, ok := GetUserContext(tt.ctx)
			assert.Equal(t, tt.wantOk, ok)
			if tt.wantOk {
				assert.Equal(t, tt.wantID, userCtx.UserID)
			}
		})
	}
}

func TestGetStringClaim(t *testing.T) {
	claims := jwt.MapClaims{
		"sub":   "123e4567-e89b-12d3-a456-426614174000",
		"email": "test@example.com",
		"role":  "authenticated",
		"num":   123,
	}

	tests := []struct {
		name     string
		key      string
		expected string
	}{
		{
			name:     "existing string claim",
			key:      "sub",
			expected: "123e4567-e89b-12d3-a456-426614174000",
		},
		{
			name:     "non-string claim",
			key:      "num",
			expected: "",
		},
		{
			name:     "non-existent claim",
			key:      "nonexistent",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getStringClaim(claims, tt.key)
			assert.Equal(t, tt.expected, result)
		})
	}
}
