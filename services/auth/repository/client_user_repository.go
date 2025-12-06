package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	db "contract-pro-suite/sqlc"
)

// ClientUserRepository クライアントユーザーリポジトリ
type ClientUserRepository interface {
	GetByID(ctx context.Context, clientUserID uuid.UUID) (db.ClientUser, error)
	GetByEmail(ctx context.Context, clientID uuid.UUID, email string) (db.ClientUser, error)
	List(ctx context.Context, clientID uuid.UUID, limit, offset int32) ([]db.ClientUser, error)
	Create(ctx context.Context, params db.CreateClientUserParams) (db.ClientUser, error)
	Update(ctx context.Context, params db.UpdateClientUserParams) (db.ClientUser, error)
	Delete(ctx context.Context, clientUserID uuid.UUID, deletedBy uuid.UUID) error
}

type clientUserRepository struct {
	queries *db.Queries
}

// NewClientUserRepository クライアントユーザーリポジトリを作成
func NewClientUserRepository(queries *db.Queries) ClientUserRepository {
	return &clientUserRepository{
		queries: queries,
	}
}

func (r *clientUserRepository) GetByID(ctx context.Context, clientUserID uuid.UUID) (db.ClientUser, error) {
	return r.queries.GetClientUser(ctx, pgtype.UUID{Bytes: clientUserID, Valid: true})
}

func (r *clientUserRepository) GetByEmail(ctx context.Context, clientID uuid.UUID, email string) (db.ClientUser, error) {
	return r.queries.GetClientUserByEmail(ctx, db.GetClientUserByEmailParams{
		ClientID: pgtype.UUID{Bytes: clientID, Valid: true},
		Email:     email,
	})
}

func (r *clientUserRepository) List(ctx context.Context, clientID uuid.UUID, limit, offset int32) ([]db.ClientUser, error) {
	return r.queries.ListClientUsers(ctx, db.ListClientUsersParams{
		ClientID: pgtype.UUID{Bytes: clientID, Valid: true},
		Limit:    limit,
		Offset:   offset,
	})
}

func (r *clientUserRepository) Create(ctx context.Context, params db.CreateClientUserParams) (db.ClientUser, error) {
	return r.queries.CreateClientUser(ctx, params)
}

func (r *clientUserRepository) Update(ctx context.Context, params db.UpdateClientUserParams) (db.ClientUser, error) {
	return r.queries.UpdateClientUser(ctx, params)
}

func (r *clientUserRepository) Delete(ctx context.Context, clientUserID uuid.UUID, deletedBy uuid.UUID) error {
	return r.queries.DeleteClientUser(ctx, db.DeleteClientUserParams{
		ClientUserID: pgtype.UUID{Bytes: clientUserID, Valid: true},
		DeletedBy:    pgtype.UUID{Bytes: deletedBy, Valid: true},
	})
}
