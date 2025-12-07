package usecase

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"contract-pro-suite/internal/shared/config"
	"contract-pro-suite/internal/shared/db"
	"contract-pro-suite/services/auth/domain"
	"contract-pro-suite/services/auth/repository"

	dbgen "contract-pro-suite/sqlc"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// SignupClientParams クライアント登録パラメータ
type SignupClientParams struct {
	// クライアント情報
	Name                   string
	CompanyCode            string
	Slug                   string
	ESignMode              string
	RetentionDefaultMonths int32
	Settings               string

	// 管理者ユーザー情報
	AdminEmail      string
	AdminPassword   string
	AdminFirstName  string
	AdminLastName   string
	AdminDepartment *string
	AdminPosition   *string
}

// SignupClientResult クライアント登録結果
type SignupClientResult struct {
	ClientID    uuid.UUID
	ClientName  string
	AdminUserID uuid.UUID
	AdminEmail  string
}

// AuthUsecase 認証ユースケース
type AuthUsecase interface {
	// GetUserContext JWTから取得したユーザーIDでデータベースからユーザー情報と権限を取得
	GetUserContext(ctx context.Context, jwtUserID string) (*domain.UserContext, error)

	// ValidateClientAccess ユーザーのクライアントアクセス権限を検証
	ValidateClientAccess(ctx context.Context, userCtx *domain.UserContext, clientID uuid.UUID) error

	// CheckPermission 権限チェック（将来の実装用）
	CheckPermission(ctx context.Context, userCtx *domain.UserContext, feature, action string) error

	// SignupClient サービス利用開始時のアカウント登録（クライアント + 管理者ユーザー作成）
	SignupClient(ctx context.Context, params SignupClientParams) (*SignupClientResult, error)
}

type authUsecase struct {
	operatorRepo              repository.OperatorRepository
	clientUserRepo            repository.ClientUserRepository
	clientRepo                repository.ClientRepository
	operatorAssignmentRepo    repository.OperatorAssignmentRepository
	clientRoleRepo            repository.ClientRoleRepository
	clientRolePermissionRepo  repository.ClientRolePermissionRepository
	clientUserRoleRepo        repository.ClientUserRoleRepository
	cfg                       *config.Config
	database                  *db.DB
}

// NewAuthUsecase 認証ユースケースを作成
func NewAuthUsecase(
	operatorRepo repository.OperatorRepository,
	clientUserRepo repository.ClientUserRepository,
	clientRepo repository.ClientRepository,
	operatorAssignmentRepo repository.OperatorAssignmentRepository,
	clientRoleRepo repository.ClientRoleRepository,
	clientRolePermissionRepo repository.ClientRolePermissionRepository,
	clientUserRoleRepo repository.ClientUserRoleRepository,
	cfg *config.Config,
	database *db.DB,
) AuthUsecase {
	return &authUsecase{
		operatorRepo:             operatorRepo,
		clientUserRepo:           clientUserRepo,
		clientRepo:               clientRepo,
		operatorAssignmentRepo:   operatorAssignmentRepo,
		clientRoleRepo:           clientRoleRepo,
		clientRolePermissionRepo: clientRolePermissionRepo,
		clientUserRoleRepo:       clientUserRoleRepo,
		cfg:                      cfg,
		database:                 database,
	}
}

// GetUserContext JWTから取得したユーザーIDでデータベースからユーザー情報を取得
func (u *authUsecase) GetUserContext(ctx context.Context, jwtUserID string) (*domain.UserContext, error) {
	userUUID, err := uuid.Parse(jwtUserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	// まずoperatorsテーブルで検索
	operator, err := u.operatorRepo.GetByID(ctx, userUUID)
	if err == nil && operator.OperatorID.Valid {
		// オペレーターの場合、operator_assignmentsテーブルからクライアントIDを取得
		// 最初のアクティブな割り当てを取得（将来は複数クライアント対応が必要）
		assignments, err := u.operatorAssignmentRepo.GetByOperatorID(ctx, userUUID)
		if err == nil && len(assignments) > 0 {
			// 最初のアクティブな割り当てのクライアントIDを使用
			clientID := uuidFromPGType(assignments[0].ClientID)
			return &domain.UserContext{
				UserID:   uuidFromPGType(operator.OperatorID),
				UserType: domain.UserTypeOperator,
				Email:    operator.Email,
				ClientID: clientID,
			}, nil
		}
		// 割り当てがない場合でもオペレーターとして返す（ClientIDは空）
		return &domain.UserContext{
			UserID:   uuidFromPGType(operator.OperatorID),
			UserType: domain.UserTypeOperator,
			Email:    operator.Email,
			ClientID: uuid.Nil,
		}, nil
	}

	// operatorsテーブルにない場合、client_usersテーブルで検索
	// 注意: client_usersはclient_idが必要なので、全件検索は非効率
	// 実際の実装では、JWTにclient_idを含めるか、別の方法で特定する必要がある
	// ここでは簡易実装として、emailで検索（複数クライアントに同じemailが存在する可能性があるため注意）

	// 将来的には、JWTにclient_idを含めるか、サブドメインからclient_idを取得する実装が必要
	return nil, errors.New("user not found")
}

// ValidateClientAccess ユーザーのクライアントアクセス権限を検証
func (u *authUsecase) ValidateClientAccess(ctx context.Context, userCtx *domain.UserContext, clientID uuid.UUID) error {
	switch userCtx.UserType {
	case domain.UserTypeOperator:
		// オペレーターの場合、operator_assignmentsテーブルで割り当てを確認
		assignments, err := u.operatorAssignmentRepo.GetByOperatorID(ctx, userCtx.UserID)
		if err != nil {
			return fmt.Errorf("failed to get operator assignments: %w", err)
		}
		// アクティブな割り当てがあるか確認
		for _, assignment := range assignments {
			if uuidFromPGType(assignment.ClientID) == clientID && assignment.Status == "ACTIVE" {
				return nil
			}
		}
		return errors.New("client access denied: operator not assigned to client")
	case domain.UserTypeClientUser:
		// クライアントユーザーの場合、client_idが一致するか確認
		if userCtx.ClientID != clientID {
			return errors.New("client access denied")
		}
		return nil
	default:
		return errors.New("unknown user type")
	}
}

// CheckPermission 権限チェック
func (u *authUsecase) CheckPermission(ctx context.Context, userCtx *domain.UserContext, feature, action string) error {
	switch userCtx.UserType {
	case domain.UserTypeOperator:
		// オペレーターの場合、operator_assignmentsテーブルでロールを確認
		if userCtx.ClientID == uuid.Nil {
			return errors.New("operator not assigned to any client")
		}
		assignment, err := u.operatorAssignmentRepo.GetByClientAndOperator(ctx, userCtx.ClientID, userCtx.UserID)
		if err != nil {
			return fmt.Errorf("failed to get operator assignment: %w", err)
		}
		if assignment.Status != "ACTIVE" {
			return errors.New("operator assignment is not active")
		}
		// ロールに基づいて権限チェック
		switch assignment.Role {
		case "ADMIN":
			// ADMIN: 全操作可能
			return nil
		case "OPERATOR":
			// OPERATOR: 一部操作可能（将来拡張可能、現時点では全操作可能）
			return nil
		case "VIEWER":
			// VIEWER: 閲覧のみ
			if action != "READ" {
				return errors.New("viewer can only read")
			}
			return nil
		default:
			return fmt.Errorf("unknown operator role: %s", assignment.Role)
		}
	case domain.UserTypeClientUser:
		// クライアントユーザーの場合、client_user_roles → client_role_permissionsで権限確認
		// 1. client_user_rolesテーブルからユーザーのロールを取得
		userRoles, err := u.clientUserRoleRepo.GetByUserID(ctx, userCtx.ClientID, userCtx.UserID)
		if err != nil {
			return fmt.Errorf("failed to get user roles: %w", err)
		}
		if len(userRoles) == 0 {
			return errors.New("user has no roles assigned")
		}
		// 2. 各ロールのclient_role_permissionsテーブルから権限を確認
		for _, userRole := range userRoles {
			roleID := uuidFromPGType(userRole.RoleID)
			permissions, err := u.clientRolePermissionRepo.GetByRoleID(ctx, roleID)
			if err != nil {
				continue // エラーが発生しても次のロールを確認
			}
			// 3. featureとactionの組み合わせが許可されているか確認
			for _, perm := range permissions {
				if perm.Feature == feature && perm.Action == action && perm.Granted {
					return nil // 権限が見つかった
				}
			}
		}
		return errors.New("permission denied")
	default:
		return errors.New("unknown user type")
	}
}

// SignupClient サービス利用開始時のアカウント登録（クライアント + 管理者ユーザー作成）
func (u *authUsecase) SignupClient(ctx context.Context, params SignupClientParams) (*SignupClientResult, error) {
	// 1. クライアント情報のバリデーション（slug, company_codeの重複チェック）
	if _, err := u.clientRepo.GetBySlug(ctx, params.Slug); err == nil {
		return nil, fmt.Errorf("slug already exists: %s", params.Slug)
	}
	// company_codeが指定されている場合のみ重複チェック
	if params.CompanyCode != "" {
		if _, err := u.clientRepo.GetByCompanyCode(ctx, params.CompanyCode); err == nil {
			return nil, fmt.Errorf("company_code already exists: %s", params.CompanyCode)
		}
	}

	// デフォルト値の設定
	eSignMode := params.ESignMode
	if eSignMode == "" {
		eSignMode = "WITNESS_OTP"
	}
	retentionMonths := params.RetentionDefaultMonths
	if retentionMonths == 0 {
		retentionMonths = 84
	}
	settingsJSON := params.Settings
	if settingsJSON == "" {
		settingsJSON = "{}"
	}

	// データベーストランザクション開始
	tx, err := u.database.Pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	queries := dbgen.New(tx)

	// 2. clientsテーブルにクライアントを登録
	// company_codeが空文字列の場合はNULLに変換
	companyCode := pgtype.Text{String: params.CompanyCode, Valid: params.CompanyCode != ""}
	clientParams := dbgen.CreateClientParams{
		Slug:                   params.Slug,
		CompanyCode:            companyCode,
		Name:                   params.Name,
		ESignMode:              eSignMode,
		RetentionDefaultMonths: retentionMonths,
		Status:                 "ACTIVE",
		Settings:               []byte(settingsJSON),
	}

	client, err := queries.CreateClient(ctx, clientParams)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	clientID := uuidFromPGType(client.ClientID)

	// 3. デフォルトロールと権限を作成
	if err := u.createDefaultRoles(ctx, tx, clientID); err != nil {
		return nil, fmt.Errorf("failed to create default roles: %w", err)
	}

	// 4. Supabase Auth Admin APIで管理者ユーザーを作成
	adminUserID, err := u.createSupabaseUser(ctx, params.AdminEmail, params.AdminPassword, params.AdminFirstName, params.AdminLastName)
	if err != nil {
		return nil, fmt.Errorf("failed to create supabase user: %w", err)
	}

	// 5. client_usersテーブルに管理者ユーザーを登録
	department := pgtype.Text{String: "", Valid: false}
	if params.AdminDepartment != nil && *params.AdminDepartment != "" {
		department = pgtype.Text{String: *params.AdminDepartment, Valid: true}
	}
	position := pgtype.Text{String: "", Valid: false}
	if params.AdminPosition != nil && *params.AdminPosition != "" {
		position = pgtype.Text{String: *params.AdminPosition, Valid: true}
	}

	clientUserParams := dbgen.CreateClientUserParams{
		ClientUserID: pgtype.UUID{Bytes: adminUserID, Valid: true},
		ClientID:     pgtype.UUID{Bytes: clientID, Valid: true},
		Email:        params.AdminEmail,
		FirstName:    params.AdminFirstName,
		LastName:     params.AdminLastName,
		Department:   department,
		Position:     position,
		Settings:     []byte("{}"),
		Status:       "ACTIVE",
	}

	_, err = queries.CreateClientUser(ctx, clientUserParams)
	if err != nil {
		return nil, fmt.Errorf("failed to create client user: %w", err)
	}

	// 6. 管理者ユーザーにsystem_adminロールを割り当て
	systemAdminRole, err := queries.GetClientRoleByCode(ctx, dbgen.GetClientRoleByCodeParams{
		ClientID: pgtype.UUID{Bytes: clientID, Valid: true},
		Code:     "system_admin",
	})
	if err == nil {
		now := time.Now()
		_, err = queries.CreateClientUserRole(ctx, dbgen.CreateClientUserRoleParams{
			ClientID:     pgtype.UUID{Bytes: clientID, Valid: true},
			ClientUserID: pgtype.UUID{Bytes: adminUserID, Valid: true},
			RoleID:       systemAdminRole.RoleID,
			AssignedAt:   pgtype.Timestamptz{Time: now, Valid: true},
		})
		if err != nil {
			return nil, fmt.Errorf("failed to assign system_admin role: %w", err)
		}
	}

	// 7. トランザクションコミット
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// 6. レスポンス返却
	return &SignupClientResult{
		ClientID:    clientID,
		ClientName:  client.Name,
		AdminUserID: adminUserID,
		AdminEmail:  params.AdminEmail,
	}, nil
}

// createSupabaseUser Supabase Auth Admin APIでユーザーを作成
func (u *authUsecase) createSupabaseUser(ctx context.Context, email, password, firstName, lastName string) (uuid.UUID, error) {
	url := fmt.Sprintf("%s/auth/v1/admin/users", u.cfg.SupabaseURL)

	reqBody := map[string]interface{}{
		"email":         email,
		"password":      password,
		"email_confirm": true,
		"user_metadata": map[string]string{
			"first_name": firstName,
			"last_name":  lastName,
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("apikey", u.cfg.SupabaseServiceRoleKey)
	req.Header.Set("Authorization", "Bearer "+u.cfg.SupabaseServiceRoleKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return uuid.Nil, fmt.Errorf("supabase auth error: status %d, body: %s", resp.StatusCode, string(body))
	}

	var result struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return uuid.Nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	userID, err := uuid.Parse(result.ID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to parse user ID: %w", err)
	}

	return userID, nil
}

// uuidFromPGType pgtype.UUIDからuuid.UUIDに変換
func uuidFromPGType(pgUUID pgtype.UUID) uuid.UUID {
	if !pgUUID.Valid {
		return uuid.Nil
	}
	return pgUUID.Bytes
}
