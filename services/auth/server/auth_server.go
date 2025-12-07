package server

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"contract-pro-suite/internal/interceptor"
	pbauth "contract-pro-suite/proto/proto/auth"
	"contract-pro-suite/services/auth/usecase"
)

// AuthServer 認証gRPCサーバー
type AuthServer struct {
	pbauth.UnimplementedAuthServiceServer
	authUsecase usecase.AuthUsecase
}

// NewAuthServer 認証gRPCサーバーを作成
func NewAuthServer(authUsecase usecase.AuthUsecase) *AuthServer {
	return &AuthServer{
		authUsecase: authUsecase,
	}
}

// GetMe 現在のユーザー情報を取得
func (s *AuthServer) GetMe(ctx context.Context, req *pbauth.GetMeRequest) (*pbauth.GetMeResponse, error) {
	// 拡張されたユーザーコンテキストを取得
	// interceptor.GetEnhancedUserContextは内部でcontextKey型を使用しているため、
	// 同じキー型で設定されたコンテキストから値を取得できる
	userCtx, ok := interceptor.GetEnhancedUserContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "unauthorized")
	}

	// レスポンスを作成
	resp := &pbauth.GetMeResponse{
		UserId:   userCtx.UserID.String(),
		Email:    userCtx.Email,
		UserType: string(userCtx.UserType),
	}

	// client_idが設定されている場合は追加
	if userCtx.ClientID.String() != "00000000-0000-0000-0000-000000000000" {
		clientIDStr := userCtx.ClientID.String()
		resp.ClientId = &clientIDStr
	}

	return resp, nil
}

// SignupClient サービス利用開始時のアカウント登録（クライアント + 管理者ユーザー作成）
func (s *AuthServer) SignupClient(ctx context.Context, req *pbauth.SignupClientRequest) (*pbauth.SignupClientResponse, error) {
	// リクエストのバリデーション
	if req.Name == "" {
		return nil, status.Errorf(codes.InvalidArgument, "name is required")
	}
	if req.Slug == "" {
		return nil, status.Errorf(codes.InvalidArgument, "slug is required")
	}
	if req.AdminEmail == "" {
		return nil, status.Errorf(codes.InvalidArgument, "admin_email is required")
	}
	if req.AdminPassword == "" {
		return nil, status.Errorf(codes.InvalidArgument, "admin_password is required")
	}
	if req.AdminFirstName == "" {
		return nil, status.Errorf(codes.InvalidArgument, "admin_first_name is required")
	}
	if req.AdminLastName == "" {
		return nil, status.Errorf(codes.InvalidArgument, "admin_last_name is required")
	}

	// デフォルト値の設定
	eSignMode := req.GetESignMode()
	if eSignMode == "" {
		eSignMode = "WITNESS_OTP"
	}
	retentionMonths := req.GetRetentionDefaultMonths()
	if retentionMonths == 0 {
		retentionMonths = 84
	}
	settings := req.GetSettings()
	if settings == "" {
		settings = "{}"
	}

	// オプションフィールドの処理
	var adminDepartment *string
	if dept := req.GetAdminDepartment(); dept != "" {
		adminDepartment = &dept
	}
	var adminPosition *string
	if pos := req.GetAdminPosition(); pos != "" {
		adminPosition = &pos
	}

	// オプションフィールドの処理
	companyCode := req.GetCompanyCode() // オプション（JIPDEC標準企業コード）

	// ユースケースを呼び出し
	params := usecase.SignupClientParams{
		Name:                   req.Name,
		CompanyCode:            companyCode,
		Slug:                   req.Slug,
		ESignMode:              eSignMode,
		RetentionDefaultMonths: retentionMonths,
		Settings:               settings,
		AdminEmail:             req.AdminEmail,
		AdminPassword:          req.AdminPassword,
		AdminFirstName:         req.AdminFirstName,
		AdminLastName:          req.AdminLastName,
		AdminDepartment:        adminDepartment,
		AdminPosition:          adminPosition,
	}

	result, err := s.authUsecase.SignupClient(ctx, params)
	if err != nil {
		// エラーの種類に応じて適切なgRPCステータスコードを返す
		errMsg := err.Error()
		companyCode := req.GetCompanyCode()
		if errMsg == "slug already exists: "+req.Slug || (companyCode != "" && errMsg == "company_code already exists: "+companyCode) {
			return nil, status.Errorf(codes.AlreadyExists, "%s", errMsg)
		}
		return nil, status.Errorf(codes.Internal, "failed to signup client: %v", err)
	}

	// レスポンスを作成
	return &pbauth.SignupClientResponse{
		ClientId:    result.ClientID.String(),
		ClientName:  result.ClientName,
		AdminUserId: result.AdminUserID.String(),
		AdminEmail:  result.AdminEmail,
	}, nil
}
