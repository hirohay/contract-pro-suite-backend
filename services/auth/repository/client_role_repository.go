package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	db "contract-pro-suite/sqlc"
)

// ClientRoleRepository クライアントロールリポジトリ
type ClientRoleRepository interface {
	GetByID(ctx context.Context, roleID uuid.UUID) (db.ClientRole, error)
	GetByCode(ctx context.Context, clientID uuid.UUID, code string) (db.ClientRole, error)
	List(ctx context.Context, clientID uuid.UUID) ([]db.ClientRole, error)
	Create(ctx context.Context, params db.CreateClientRoleParams) (db.ClientRole, error)
	Update(ctx context.Context, params db.UpdateClientRoleParams) (db.ClientRole, error)
	Delete(ctx context.Context, roleID uuid.UUID, deletedBy uuid.UUID) error
}

type clientRoleRepository struct {
	queries *db.Queries
}

// NewClientRoleRepository クライアントロールリポジトリを作成
func NewClientRoleRepository(queries *db.Queries) ClientRoleRepository {
	return &clientRoleRepository{
		queries: queries,
	}
}

func (r *clientRoleRepository) GetByID(ctx context.Context, roleID uuid.UUID) (db.ClientRole, error) {
	return r.queries.GetClientRole(ctx, pgtype.UUID{Bytes: roleID, Valid: true})
}

func (r *clientRoleRepository) GetByCode(ctx context.Context, clientID uuid.UUID, code string) (db.ClientRole, error) {
	return r.queries.GetClientRoleByCode(ctx, db.GetClientRoleByCodeParams{
		ClientID: pgtype.UUID{Bytes: clientID, Valid: true},
		Code:     code,
	})
}

func (r *clientRoleRepository) List(ctx context.Context, clientID uuid.UUID) ([]db.ClientRole, error) {
	return r.queries.ListClientRoles(ctx, pgtype.UUID{Bytes: clientID, Valid: true})
}

func (r *clientRoleRepository) Create(ctx context.Context, params db.CreateClientRoleParams) (db.ClientRole, error) {
	return r.queries.CreateClientRole(ctx, params)
}

func (r *clientRoleRepository) Update(ctx context.Context, params db.UpdateClientRoleParams) (db.ClientRole, error) {
	return r.queries.UpdateClientRole(ctx, params)
}

func (r *clientRoleRepository) Delete(ctx context.Context, roleID uuid.UUID, deletedBy uuid.UUID) error {
	return r.queries.DeleteClientRole(ctx, db.DeleteClientRoleParams{
		RoleID:    pgtype.UUID{Bytes: roleID, Valid: true},
		DeletedBy: pgtype.UUID{Bytes: deletedBy, Valid: true},
	})
}

