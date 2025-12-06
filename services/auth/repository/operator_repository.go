package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	db "contract-pro-suite/sqlc"
)

// OperatorRepository オペレーターリポジトリ
type OperatorRepository interface {
	GetByID(ctx context.Context, operatorID uuid.UUID) (db.Operator, error)
	GetByEmail(ctx context.Context, email string) (db.Operator, error)
	List(ctx context.Context, limit, offset int32) ([]db.Operator, error)
	Create(ctx context.Context, params db.CreateOperatorParams) (db.Operator, error)
	Update(ctx context.Context, params db.UpdateOperatorParams) (db.Operator, error)
	Delete(ctx context.Context, operatorID uuid.UUID, deletedBy uuid.UUID) error
}

type operatorRepository struct {
	queries *db.Queries
}

// NewOperatorRepository オペレーターリポジトリを作成
func NewOperatorRepository(queries *db.Queries) OperatorRepository {
	return &operatorRepository{
		queries: queries,
	}
}

func (r *operatorRepository) GetByID(ctx context.Context, operatorID uuid.UUID) (db.Operator, error) {
	return r.queries.GetOperator(ctx, pgtype.UUID{Bytes: operatorID, Valid: true})
}

func (r *operatorRepository) GetByEmail(ctx context.Context, email string) (db.Operator, error) {
	return r.queries.GetOperatorByEmail(ctx, email)
}

func (r *operatorRepository) List(ctx context.Context, limit, offset int32) ([]db.Operator, error) {
	return r.queries.ListOperators(ctx, db.ListOperatorsParams{
		Limit:  limit,
		Offset: offset,
	})
}

func (r *operatorRepository) Create(ctx context.Context, params db.CreateOperatorParams) (db.Operator, error) {
	return r.queries.CreateOperator(ctx, params)
}

func (r *operatorRepository) Update(ctx context.Context, params db.UpdateOperatorParams) (db.Operator, error) {
	return r.queries.UpdateOperator(ctx, params)
}

func (r *operatorRepository) Delete(ctx context.Context, operatorID uuid.UUID, deletedBy uuid.UUID) error {
	return r.queries.DeleteOperator(ctx, db.DeleteOperatorParams{
		OperatorID: pgtype.UUID{Bytes: operatorID, Valid: true},
		DeletedBy:  pgtype.UUID{Bytes: deletedBy, Valid: true},
	})
}
