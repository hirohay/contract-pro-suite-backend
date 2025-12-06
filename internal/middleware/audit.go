package middleware

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// AuditLog 監査ログの構造
type AuditLog struct {
	Timestamp  time.Time `json:"timestamp"`
	UserID     string    `json:"user_id,omitempty"`
	ClientID   string    `json:"client_id,omitempty"`
	Method     string    `json:"method"`
	Path       string    `json:"path"`
	StatusCode int       `json:"status_code"`
	UserType   string    `json:"user_type,omitempty"`
	Error      string    `json:"error,omitempty"`
}

// AuditMiddleware 認証・認可のログを記録するミドルウェア
func AuditMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// レスポンスをラップしてステータスコードを記録
			rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			// リクエストを処理
			next.ServeHTTP(rw, r)

			// ログを記録
			auditLog := AuditLog{
				Timestamp:  start,
				Method:     r.Method,
				Path:       r.URL.Path,
				StatusCode: rw.statusCode,
			}

			// ユーザー情報を取得（利用可能な場合）
			if userCtx, ok := GetEnhancedUserContext(r.Context()); ok {
				auditLog.UserID = userCtx.UserID.String()
				auditLog.UserType = string(userCtx.UserType)
				if userCtx.ClientID != uuid.Nil {
					auditLog.ClientID = userCtx.ClientID.String()
				}
			}

			// client_idを取得（利用可能な場合）
			if clientID, ok := GetClientIDFromContext(r.Context()); ok && auditLog.ClientID == "" {
				auditLog.ClientID = clientID.String()
			}

			// エラーログ（4xx, 5xxステータスコード）
			if rw.statusCode >= 400 {
				auditLog.Error = http.StatusText(rw.statusCode)
			}

			// 構造化ログを出力（JSON形式）
			logJSON(auditLog)
		})
	}
}

// responseWriter ステータスコードを記録するためのResponseWriterラッパー
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// logJSON 構造化ログをJSON形式で出力
func logJSON(auditLog AuditLog) {
	logBytes, err := json.Marshal(auditLog)
	if err != nil {
		log.Printf("Failed to marshal audit log: %v", err)
		return
	}
	log.Println(string(logBytes))
}

