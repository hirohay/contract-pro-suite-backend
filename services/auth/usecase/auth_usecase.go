package usecase

import (
	"context"
	"errors"
	"fmt"

	"contract-pro-suite/services/auth/domain"
	"contract-pro-suite/services/auth/repository"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// AuthUsecase 認証ユースケース
type AuthUsecase interface {
	// GetUserContext JWTから取得したユーザーIDでデータベースからユーザー情報と権限を取得
	GetUserContext(ctx context.Context, jwtUserID string) (*domain.UserContext, error)

	// ValidateClientAccess ユーザーのクライアントアクセス権限を検証
	ValidateClientAccess(ctx context.Context, userCtx *domain.UserContext, clientID uuid.UUID) error

	// CheckPermission 権限チェック（将来の実装用）
	CheckPermission(ctx context.Context, userCtx *domain.UserContext, feature, action string) error
}

type authUsecase struct {
	operatorRepo   repository.OperatorRepository
	clientUserRepo repository.ClientUserRepository
	clientRepo     repository.ClientRepository
}

// NewAuthUsecase 認証ユースケースを作成
func NewAuthUsecase(
	operatorRepo repository.OperatorRepository,
	clientUserRepo repository.ClientUserRepository,
	clientRepo repository.ClientRepository,
) AuthUsecase {
	return &authUsecase{
		operatorRepo:   operatorRepo,
		clientUserRepo: clientUserRepo,
		clientRepo:     clientRepo,
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
		// オペレーターの場合、operator_assignmentsテーブルからクライアントIDを取得する必要がある
		// 現時点では最初の割り当てを取得（将来は複数クライアント対応が必要）
		return &domain.UserContext{
			UserID:   uuidFromPGType(operator.OperatorID),
			UserType: domain.UserTypeOperator,
			Email:    operator.Email,
			// ClientIDはoperator_assignmentsテーブルから取得する必要がある（現時点では空）
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
		// 現時点では簡易実装（将来はoperator_assignmentsテーブルの実装が必要）
		// ここでは一旦許可（実装後に詳細なチェックを追加）
		return nil
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
		// 現時点では簡易実装（将来はoperator_assignmentsテーブルの実装が必要）
		// ADMIN: 全操作可能
		// OPERATOR: 一部操作可能
		// VIEWER: 閲覧のみ
		// 現時点では一旦許可（実装後に詳細なチェックを追加）
		return nil
	case domain.UserTypeClientUser:
		// クライアントユーザーの場合、client_role_permissionsテーブルで権限を確認
		// 現時点では簡易実装（将来はclient_user_rolesとclient_role_permissionsテーブルの実装が必要）
		// 1. client_user_rolesテーブルからユーザーのロールを取得
		// 2. 各ロールのclient_role_permissionsテーブルから権限を取得
		// 3. featureとactionの組み合わせが許可されているか確認
		// 現時点では一旦許可（実装後に詳細なチェックを追加）
		return nil
	default:
		return errors.New("unknown user type")
	}
}

// uuidFromPGType pgtype.UUIDからuuid.UUIDに変換
func uuidFromPGType(pgUUID pgtype.UUID) uuid.UUID {
	if !pgUUID.Valid {
		return uuid.Nil
	}
	return pgUUID.Bytes
}
