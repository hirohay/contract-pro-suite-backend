package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"contract-pro-suite/internal/middleware"
	"contract-pro-suite/services/auth/domain"
	"contract-pro-suite/services/auth/usecase"
)

// AuthHandler 認証ハンドラ
type AuthHandler struct {
	authUsecase usecase.AuthUsecase
}

// NewAuthHandler 認証ハンドラを作成
func NewAuthHandler(authUsecase usecase.AuthUsecase) *AuthHandler {
	return &AuthHandler{
		authUsecase: authUsecase,
	}
}

// RegisterRoutes ルートを登録
func (h *AuthHandler) RegisterRoutes(r chi.Router) {
	r.Route("/auth", func(r chi.Router) {
		r.Get("/me", h.GetMe)
	})
}

// GetMe 現在のユーザー情報を取得
func (h *AuthHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	// 拡張されたユーザーコンテキストを取得
	userCtx, ok := middleware.GetEnhancedUserContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// レスポンス
	response := UserResponse{
		UserID:   userCtx.UserID.String(),
		UserType: userCtx.UserType,
		Email:    userCtx.Email,
	}

	if userCtx.ClientID.String() != "00000000-0000-0000-0000-000000000000" {
		response.ClientID = userCtx.ClientID.String()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// UserResponse ユーザー情報レスポンス
type UserResponse struct {
	UserID   string          `json:"user_id"`
	UserType domain.UserType `json:"user_type"`
	Email    string          `json:"email"`
	ClientID string          `json:"client_id,omitempty"`
}

