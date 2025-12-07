package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	db "contract-pro-suite/sqlc"
)

// ClientUserRoleRepository クライアントユーザーロールリポジトリ
type ClientUserRoleRepository interface {
	GetByUserAndRole(ctx context.Context, clientID, clientUserID, roleID uuid.UUID) (db.ClientUserRole, error)
	GetByUserID(ctx context.Context, clientID, clientUserID uuid.UUID) ([]db.ClientUserRole, error)
	GetByRoleID(ctx context.Context, clientID, roleID uuid.UUID) ([]db.ClientUserRole, error)
	Create(ctx context.Context, params db.CreateClientUserRoleParams) (db.ClientUserRole, error)
	Revoke(ctx context.Context, clientID, clientUserID, roleID uuid.UUID) error
	Delete(ctx context.Context, clientID, clientUserID, roleID uuid.UUID, deletedBy uuid.UUID) error
}

type clientUserRoleRepository struct {
	queries *db.Queries
}

// NewClientUserRoleRepository クライアントユーザーロールリポジトリを作成
func NewClientUserRoleRepository(queries *db.Queries) ClientUserRoleRepository {
	return &clientUserRoleRepository{
		queries: queries,
	}
}

func (r *clientUserRoleRepository) GetByUserAndRole(ctx context.Context, clientID, clientUserID, roleID uuid.UUID) (db.ClientUserRole, error) {
	return r.queries.GetClientUserRole(ctx, db.GetClientUserRoleParams{
		ClientID:     pgtype.UUID{Bytes: clientID, Valid: true},
		ClientUserID: pgtype.UUID{Bytes: clientUserID, Valid: true},
		RoleID:       pgtype.UUID{Bytes: roleID, Valid: true},
	})
}

func (r *clientUserRoleRepository) GetByUserID(ctx context.Context, clientID, clientUserID uuid.UUID) ([]db.ClientUserRole, error) {
	return r.queries.GetClientUserRolesByUserID(ctx, db.GetClientUserRolesByUserIDParams{
		ClientID:     pgtype.UUID{Bytes: clientID, Valid: true},
		ClientUserID: pgtype.UUID{Bytes: clientUserID, Valid: true},
	})
}

func (r *clientUserRoleRepository) GetByRoleID(ctx context.Context, clientID, roleID uuid.UUID) ([]db.ClientUserRole, error) {
	return r.queries.GetClientUserRolesByRoleID(ctx, db.GetClientUserRolesByRoleIDParams{
		ClientID: pgtype.UUID{Bytes: clientID, Valid: true},
		RoleID:   pgtype.UUID{Bytes: roleID, Valid: true},
	})
}

func (r *clientUserRoleRepository) Create(ctx context.Context, params db.CreateClientUserRoleParams) (db.ClientUserRole, error) {
	return r.queries.CreateClientUserRole(ctx, params)
}

func (r *clientUserRoleRepository) Revoke(ctx context.Context, clientID, clientUserID, roleID uuid.UUID) error {
	return r.queries.RevokeClientUserRole(ctx, db.RevokeClientUserRoleParams{
		ClientID:     pgtype.UUID{Bytes: clientID, Valid: true},
		ClientUserID: pgtype.UUID{Bytes: clientUserID, Valid: true},
		RoleID:       pgtype.UUID{Bytes: roleID, Valid: true},
	})
}

func (r *clientUserRoleRepository) Delete(ctx context.Context, clientID, clientUserID, roleID uuid.UUID, deletedBy uuid.UUID) error {
	return r.queries.DeleteClientUserRole(ctx, db.DeleteClientUserRoleParams{
		ClientID:     pgtype.UUID{Bytes: clientID, Valid: true},
		ClientUserID: pgtype.UUID{Bytes: clientUserID, Valid: true},
		RoleID:       pgtype.UUID{Bytes: roleID, Valid: true},
		DeletedBy:    pgtype.UUID{Bytes: deletedBy, Valid: true},
	})
}

