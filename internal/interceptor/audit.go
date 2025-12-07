package interceptor

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

// AuditInterceptor 認証・認可のログを記録するインターセプター
func AuditInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		start := time.Now()

		// リクエストを処理
		resp, err := handler(ctx, req)

		// ログを記録
		auditLog := AuditLog{
			Timestamp:  start,
			Method:     "gRPC",
			Path:       info.FullMethod,
			StatusCode: getStatusCode(err),
		}

		// ユーザー情報を取得（利用可能な場合）
		if userCtx, ok := GetEnhancedUserContext(ctx); ok {
			auditLog.UserID = userCtx.UserID.String()
			auditLog.UserType = string(userCtx.UserType)
			if userCtx.ClientID != uuid.Nil {
				auditLog.ClientID = userCtx.ClientID.String()
			}
		}

		// client_idを取得（利用可能な場合）
		if clientID, ok := GetClientIDFromContext(ctx); ok && auditLog.ClientID == "" {
			auditLog.ClientID = clientID.String()
		}

		// エラーログ（4xx, 5xxステータスコード相当）
		if err != nil {
			if st, ok := status.FromError(err); ok {
				auditLog.Error = st.Message()
			} else {
				auditLog.Error = err.Error()
			}
		}

		// 構造化ログを出力（JSON形式）
		logJSON(auditLog)

		return resp, err
	}
}

// getStatusCode gRPCエラーからHTTPステータスコード相当の値を取得
func getStatusCode(err error) int {
	if err == nil {
		return 200 // OK
	}
	if st, ok := status.FromError(err); ok {
		switch st.Code() {
		case codes.OK:
			return 200
		case codes.InvalidArgument:
			return 400
		case codes.Unauthenticated:
			return 401
		case codes.PermissionDenied:
			return 403
		case codes.NotFound:
			return 404
		case codes.Internal:
			return 500
		default:
			return 500
		}
	}
	return 500
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

