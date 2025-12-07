package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	db "contract-pro-suite/sqlc"
)

// ClientRolePermissionRepository クライアントロール権限リポジトリ
type ClientRolePermissionRepository interface {
	GetByRoleFeatureAction(ctx context.Context, roleID uuid.UUID, feature, action string) (db.ClientRolePermission, error)
	GetByRoleID(ctx context.Context, roleID uuid.UUID) ([]db.ClientRolePermission, error)
	GetByFeatureAndAction(ctx context.Context, roleID uuid.UUID, feature, action string) ([]db.ClientRolePermission, error)
	Create(ctx context.Context, params db.CreateClientRolePermissionParams) (db.ClientRolePermission, error)
	Update(ctx context.Context, params db.UpdateClientRolePermissionParams) (db.ClientRolePermission, error)
	Delete(ctx context.Context, roleID uuid.UUID, feature, action string, deletedBy uuid.UUID) error
	DeleteByRoleID(ctx context.Context, roleID uuid.UUID, deletedBy uuid.UUID) error
}

type clientRolePermissionRepository struct {
	queries *db.Queries
}

// NewClientRolePermissionRepository クライアントロール権限リポジトリを作成
func NewClientRolePermissionRepository(queries *db.Queries) ClientRolePermissionRepository {
	return &clientRolePermissionRepository{
		queries: queries,
	}
}

func (r *clientRolePermissionRepository) GetByRoleFeatureAction(ctx context.Context, roleID uuid.UUID, feature, action string) (db.ClientRolePermission, error) {
	return r.queries.GetClientRolePermission(ctx, db.GetClientRolePermissionParams{
		RoleID:  pgtype.UUID{Bytes: roleID, Valid: true},
		Feature: feature,
		Action:  action,
	})
}

func (r *clientRolePermissionRepository) GetByRoleID(ctx context.Context, roleID uuid.UUID) ([]db.ClientRolePermission, error) {
	return r.queries.GetClientRolePermissionsByRoleID(ctx, pgtype.UUID{Bytes: roleID, Valid: true})
}

func (r *clientRolePermissionRepository) GetByFeatureAndAction(ctx context.Context, roleID uuid.UUID, feature, action string) ([]db.ClientRolePermission, error) {
	return r.queries.GetClientRolePermissionsByFeatureAndAction(ctx, db.GetClientRolePermissionsByFeatureAndActionParams{
		RoleID:  pgtype.UUID{Bytes: roleID, Valid: true},
		Feature: feature,
		Action:  action,
	})
}

func (r *clientRolePermissionRepository) Create(ctx context.Context, params db.CreateClientRolePermissionParams) (db.ClientRolePermission, error) {
	return r.queries.CreateClientRolePermission(ctx, params)
}

func (r *clientRolePermissionRepository) Update(ctx context.Context, params db.UpdateClientRolePermissionParams) (db.ClientRolePermission, error) {
	return r.queries.UpdateClientRolePermission(ctx, params)
}

func (r *clientRolePermissionRepository) Delete(ctx context.Context, roleID uuid.UUID, feature, action string, deletedBy uuid.UUID) error {
	return r.queries.DeleteClientRolePermission(ctx, db.DeleteClientRolePermissionParams{
		RoleID:    pgtype.UUID{Bytes: roleID, Valid: true},
		Feature:   feature,
		Action:    action,
		DeletedBy: pgtype.UUID{Bytes: deletedBy, Valid: true},
	})
}

func (r *clientRolePermissionRepository) DeleteByRoleID(ctx context.Context, roleID uuid.UUID, deletedBy uuid.UUID) error {
	return r.queries.DeleteClientRolePermissionsByRoleID(ctx, db.DeleteClientRolePermissionsByRoleIDParams{
		RoleID:    pgtype.UUID{Bytes: roleID, Valid: true},
		DeletedBy: pgtype.UUID{Bytes: deletedBy, Valid: true},
	})
}

