package server

import (
	"context"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"contract-pro-suite/internal/interceptor"
	pbauth "contract-pro-suite/proto/auth"
	"contract-pro-suite/services/auth/usecase"
	dbgen "contract-pro-suite/sqlc"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
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

// ListClientUsers クライアントユーザー一覧取得
func (s *AuthServer) ListClientUsers(ctx context.Context, req *pbauth.ListClientUsersRequest) (*pbauth.ListClientUsersResponse, error) {
	// ユーザーコンテキストを取得
	userCtx, ok := interceptor.GetEnhancedUserContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "unauthorized")
	}

	// パラメータの取得
	limit := req.GetLimit()
	offset := req.GetOffset()

	// ユースケースを呼び出し
	users, total, err := s.authUsecase.ListClientUsers(ctx, userCtx, limit, offset)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list client users: %v", err)
	}

	// レスポンスを作成
	pbUsers := make([]*pbauth.ClientUser, len(users))
	for i, user := range users {
		pbUsers[i] = convertClientUserToPB(user)
	}

	return &pbauth.ListClientUsersResponse{
		Users: pbUsers,
		Total: total,
	}, nil
}

// GetClientUser クライアントユーザー詳細取得
func (s *AuthServer) GetClientUser(ctx context.Context, req *pbauth.GetClientUserRequest) (*pbauth.GetClientUserResponse, error) {
	// ユーザーコンテキストを取得
	userCtx, ok := interceptor.GetEnhancedUserContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "unauthorized")
	}

	// リクエストのバリデーション
	clientUserID, err := uuid.Parse(req.GetClientUserId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid client_user_id: %v", err)
	}

	// ユースケースを呼び出し
	user, err := s.authUsecase.GetClientUser(ctx, userCtx, clientUserID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get client user: %v", err)
	}

	// レスポンスを作成
	return &pbauth.GetClientUserResponse{
		User: convertClientUserToPB(user),
	}, nil
}

// CreateClientUser クライアントユーザー作成
func (s *AuthServer) CreateClientUser(ctx context.Context, req *pbauth.CreateClientUserRequest) (*pbauth.CreateClientUserResponse, error) {
	// ユーザーコンテキストを取得
	userCtx, ok := interceptor.GetEnhancedUserContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "unauthorized")
	}

	// リクエストのバリデーション
	if req.GetEmail() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "email is required")
	}
	if req.GetPassword() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "password is required")
	}
	if req.GetFirstName() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "first_name is required")
	}
	if req.GetLastName() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "last_name is required")
	}

	// パラメータの構築
	params := usecase.CreateClientUserParams{
		Email:     req.GetEmail(),
		Password:  req.GetPassword(),
		FirstName: req.GetFirstName(),
		LastName:  req.GetLastName(),
	}

	if req.Department != nil {
		params.Department = req.Department
	}
	if req.Position != nil {
		params.Position = req.Position
	}
	if req.Settings != nil {
		params.Settings = *req.Settings
	}

	// ユースケースを呼び出し
	user, err := s.authUsecase.CreateClientUser(ctx, userCtx, params)
	if err != nil {
		errMsg := err.Error()
		if errMsg == "email already exists: "+req.GetEmail() {
			return nil, status.Errorf(codes.AlreadyExists, "%s", errMsg)
		}
		return nil, status.Errorf(codes.Internal, "failed to create client user: %v", err)
	}

	// レスポンスを作成
	return &pbauth.CreateClientUserResponse{
		User: convertClientUserToPB(user),
	}, nil
}

// UpdateClientUser クライアントユーザー更新
func (s *AuthServer) UpdateClientUser(ctx context.Context, req *pbauth.UpdateClientUserRequest) (*pbauth.UpdateClientUserResponse, error) {
	// ユーザーコンテキストを取得
	userCtx, ok := interceptor.GetEnhancedUserContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "unauthorized")
	}

	// リクエストのバリデーション
	clientUserID, err := uuid.Parse(req.GetClientUserId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid client_user_id: %v", err)
	}

	// パラメータの構築
	params := usecase.UpdateClientUserParams{}

	if req.Email != nil {
		params.Email = req.Email
	}
	if req.FirstName != nil {
		params.FirstName = req.FirstName
	}
	if req.LastName != nil {
		params.LastName = req.LastName
	}
	if req.Department != nil {
		params.Department = req.Department
	}
	if req.Position != nil {
		params.Position = req.Position
	}
	if req.Settings != nil {
		params.Settings = req.Settings
	}
	if req.Status != nil {
		params.Status = req.Status
	}

	// ユースケースを呼び出し
	user, err := s.authUsecase.UpdateClientUser(ctx, userCtx, clientUserID, params)
	if err != nil {
		errMsg := err.Error()
		if req.Email != nil && errMsg == "email already exists: "+*req.Email {
			return nil, status.Errorf(codes.AlreadyExists, "%s", errMsg)
		}
		return nil, status.Errorf(codes.Internal, "failed to update client user: %v", err)
	}

	// レスポンスを作成
	return &pbauth.UpdateClientUserResponse{
		User: convertClientUserToPB(user),
	}, nil
}

// DeleteClientUser クライアントユーザー削除
func (s *AuthServer) DeleteClientUser(ctx context.Context, req *pbauth.DeleteClientUserRequest) (*pbauth.DeleteClientUserResponse, error) {
	// ユーザーコンテキストを取得
	userCtx, ok := interceptor.GetEnhancedUserContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "unauthorized")
	}

	// リクエストのバリデーション
	clientUserID, err := uuid.Parse(req.GetClientUserId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid client_user_id: %v", err)
	}

	// ユースケースを呼び出し
	if err := s.authUsecase.DeleteClientUser(ctx, userCtx, clientUserID); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete client user: %v", err)
	}

	// レスポンスを作成
	return &pbauth.DeleteClientUserResponse{}, nil
}

// convertClientUserToPB dbgen.ClientUserをpbauth.ClientUserに変換
func convertClientUserToPB(user dbgen.ClientUser) *pbauth.ClientUser {
	pbUser := &pbauth.ClientUser{
		ClientUserId: uuidFromPGType(user.ClientUserID).String(),
		ClientId:     uuidFromPGType(user.ClientID).String(),
		Email:        user.Email,
		FirstName:    user.FirstName,
		LastName:     user.LastName,
		Status:       user.Status,
		Settings:     string(user.Settings),
	}

	// CreatedAtの変換
	if user.CreatedAt.Valid {
		pbUser.CreatedAt = user.CreatedAt.Time.Format(time.RFC3339)
	} else {
		pbUser.CreatedAt = time.Time{}.Format(time.RFC3339)
	}

	// UpdatedAtの変換
	if user.UpdatedAt.Valid {
		pbUser.UpdatedAt = user.UpdatedAt.Time.Format(time.RFC3339)
	} else {
		pbUser.UpdatedAt = time.Time{}.Format(time.RFC3339)
	}

	if user.Department.Valid {
		pbUser.Department = &user.Department.String
	}
	if user.Position.Valid {
		pbUser.Position = &user.Position.String
	}

	return pbUser
}

// uuidFromPGType pgtype.UUIDからuuid.UUIDに変換
func uuidFromPGType(pgUUID pgtype.UUID) uuid.UUID {
	if !pgUUID.Valid {
		return uuid.Nil
	}
	return pgUUID.Bytes
}
