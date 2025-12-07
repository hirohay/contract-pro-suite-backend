package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	db "contract-pro-suite/sqlc"
)

// OperatorAssignmentRepository オペレーター割当リポジトリ
type OperatorAssignmentRepository interface {
	GetByClientAndOperator(ctx context.Context, clientID, operatorID uuid.UUID) (db.OperatorAssignment, error)
	GetByOperatorID(ctx context.Context, operatorID uuid.UUID) ([]db.OperatorAssignment, error)
	GetByClientID(ctx context.Context, clientID uuid.UUID) ([]db.OperatorAssignment, error)
	Create(ctx context.Context, params db.CreateOperatorAssignmentParams) (db.OperatorAssignment, error)
	Update(ctx context.Context, params db.UpdateOperatorAssignmentParams) (db.OperatorAssignment, error)
	Delete(ctx context.Context, clientID, operatorID uuid.UUID, deletedBy uuid.UUID) error
}

type operatorAssignmentRepository struct {
	queries *db.Queries
}

// NewOperatorAssignmentRepository オペレーター割当リポジトリを作成
func NewOperatorAssignmentRepository(queries *db.Queries) OperatorAssignmentRepository {
	return &operatorAssignmentRepository{
		queries: queries,
	}
}

func (r *operatorAssignmentRepository) GetByClientAndOperator(ctx context.Context, clientID, operatorID uuid.UUID) (db.OperatorAssignment, error) {
	return r.queries.GetOperatorAssignment(ctx, db.GetOperatorAssignmentParams{
		ClientID:   pgtype.UUID{Bytes: clientID, Valid: true},
		OperatorID: pgtype.UUID{Bytes: operatorID, Valid: true},
	})
}

func (r *operatorAssignmentRepository) GetByOperatorID(ctx context.Context, operatorID uuid.UUID) ([]db.OperatorAssignment, error) {
	return r.queries.GetOperatorAssignmentsByOperatorID(ctx, pgtype.UUID{Bytes: operatorID, Valid: true})
}

func (r *operatorAssignmentRepository) GetByClientID(ctx context.Context, clientID uuid.UUID) ([]db.OperatorAssignment, error) {
	return r.queries.GetOperatorAssignmentsByClientID(ctx, pgtype.UUID{Bytes: clientID, Valid: true})
}

func (r *operatorAssignmentRepository) Create(ctx context.Context, params db.CreateOperatorAssignmentParams) (db.OperatorAssignment, error) {
	return r.queries.CreateOperatorAssignment(ctx, params)
}

func (r *operatorAssignmentRepository) Update(ctx context.Context, params db.UpdateOperatorAssignmentParams) (db.OperatorAssignment, error) {
	return r.queries.UpdateOperatorAssignment(ctx, params)
}

func (r *operatorAssignmentRepository) Delete(ctx context.Context, clientID, operatorID uuid.UUID, deletedBy uuid.UUID) error {
	return r.queries.DeleteOperatorAssignment(ctx, db.DeleteOperatorAssignmentParams{
		ClientID:   pgtype.UUID{Bytes: clientID, Valid: true},
		OperatorID: pgtype.UUID{Bytes: operatorID, Valid: true},
		DeletedBy:  pgtype.UUID{Bytes: deletedBy, Valid: true},
	})
}

