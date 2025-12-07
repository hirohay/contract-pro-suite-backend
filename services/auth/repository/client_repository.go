package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	db "contract-pro-suite/sqlc"
)

// ClientRepository クライアントリポジトリ
type ClientRepository interface {
	GetByID(ctx context.Context, clientID uuid.UUID) (db.Client, error)
	GetBySlug(ctx context.Context, slug string) (db.Client, error)
	GetByCompanyCode(ctx context.Context, companyCode string) (db.Client, error)
	List(ctx context.Context, limit, offset int32) ([]db.Client, error)
	Create(ctx context.Context, params db.CreateClientParams) (db.Client, error)
	Update(ctx context.Context, params db.UpdateClientParams) (db.Client, error)
	Delete(ctx context.Context, clientID uuid.UUID, deletedBy uuid.UUID) error
}

type clientRepository struct {
	queries *db.Queries
}

// NewClientRepository クライアントリポジトリを作成
func NewClientRepository(queries *db.Queries) ClientRepository {
	return &clientRepository{
		queries: queries,
	}
}

func (r *clientRepository) GetByID(ctx context.Context, clientID uuid.UUID) (db.Client, error) {
	return r.queries.GetClient(ctx, pgtype.UUID{Bytes: clientID, Valid: true})
}

func (r *clientRepository) GetBySlug(ctx context.Context, slug string) (db.Client, error) {
	return r.queries.GetClientBySlug(ctx, slug)
}

func (r *clientRepository) GetByCompanyCode(ctx context.Context, companyCode string) (db.Client, error) {
	return r.queries.GetClientByCompanyCode(ctx, pgtype.Text{String: companyCode, Valid: companyCode != ""})
}

func (r *clientRepository) List(ctx context.Context, limit, offset int32) ([]db.Client, error) {
	return r.queries.ListClients(ctx, db.ListClientsParams{
		Limit:  limit,
		Offset: offset,
	})
}

func (r *clientRepository) Create(ctx context.Context, params db.CreateClientParams) (db.Client, error) {
	return r.queries.CreateClient(ctx, params)
}

func (r *clientRepository) Update(ctx context.Context, params db.UpdateClientParams) (db.Client, error) {
	return r.queries.UpdateClient(ctx, params)
}

func (r *clientRepository) Delete(ctx context.Context, clientID uuid.UUID, deletedBy uuid.UUID) error {
	return r.queries.DeleteClient(ctx, db.DeleteClientParams{
		ClientID:  pgtype.UUID{Bytes: clientID, Valid: true},
		DeletedBy: pgtype.UUID{Bytes: deletedBy, Valid: true},
	})
}
