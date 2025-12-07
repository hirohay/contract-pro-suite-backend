package repository

import (
	"context"
	"testing"

	db "contract-pro-suite/sqlc"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestClientUserRepository_GetByID_ClientIsolation クライアント分離のテスト
func TestClientUserRepository_GetByID_ClientIsolation(t *testing.T) {
	// 注意: このテストは実際のデータベース接続が必要です
	// 統合テスト環境でのみ実行してください
	t.Skip("統合テスト環境でのみ実行")

	ctx := context.Background()
	queries := db.New(nil) // 実際のデータベース接続が必要
	repo := NewClientUserRepository(queries)

	// テストデータの準備
	clientID1 := uuid.New()
	clientID2 := uuid.New()
	clientUserID1 := uuid.New()
	clientUserID2 := uuid.New()

	// クライアント1のユーザーを作成
	_, err := queries.CreateClientUser(ctx, db.CreateClientUserParams{
		ClientUserID: pgtype.UUID{Bytes: clientUserID1, Valid: true},
		ClientID:     pgtype.UUID{Bytes: clientID1, Valid: true},
		Email:        "user1@client1.com",
		FirstName:    "User1",
		LastName:     "Client1",
		Status:       "ACTIVE",
		Settings:     []byte("{}"),
	})
	require.NoError(t, err)

	// クライアント2のユーザーを作成
	_, err = queries.CreateClientUser(ctx, db.CreateClientUserParams{
		ClientUserID: pgtype.UUID{Bytes: clientUserID2, Valid: true},
		ClientID:     pgtype.UUID{Bytes: clientID2, Valid: true},
		Email:        "user2@client2.com",
		FirstName:    "User2",
		LastName:     "Client2",
		Status:       "ACTIVE",
		Settings:     []byte("{}"),
	})
	require.NoError(t, err)

	t.Run("正しいクライアントIDで取得できる", func(t *testing.T) {
		user, err := repo.GetByID(ctx, clientID1, clientUserID1)
		assert.NoError(t, err)
		assert.Equal(t, clientUserID1, user.ClientUserID.Bytes)
		assert.Equal(t, clientID1, user.ClientID.Bytes)
	})

	t.Run("異なるクライアントIDで取得できない（セキュリティチェック）", func(t *testing.T) {
		// クライアント1のユーザーIDで、クライアント2のIDを指定して取得を試みる
		_, err := repo.GetByID(ctx, clientID2, clientUserID1)
		// エラーが発生するか、結果が空であることを確認
		assert.Error(t, err, "異なるクライアントのユーザーを取得できてはいけない")
	})

	t.Run("存在しないクライアントIDで取得できない", func(t *testing.T) {
		nonExistentClientID := uuid.New()
		_, err := repo.GetByID(ctx, nonExistentClientID, clientUserID1)
		assert.Error(t, err, "存在しないクライアントIDで取得できてはいけない")
	})
}

// TestClientUserRepository_Update_ClientIsolation 更新時のクライアント分離テスト
func TestClientUserRepository_Update_ClientIsolation(t *testing.T) {
	t.Skip("統合テスト環境でのみ実行")

	ctx := context.Background()
	queries := db.New(nil) // 実際のデータベース接続が必要
	repo := NewClientUserRepository(queries)

	// テストデータの準備
	clientID1 := uuid.New()
	clientID2 := uuid.New()
	clientUserID1 := uuid.New()

	// クライアント1のユーザーを作成
	_, err := queries.CreateClientUser(ctx, db.CreateClientUserParams{
		ClientUserID: pgtype.UUID{Bytes: clientUserID1, Valid: true},
		ClientID:     pgtype.UUID{Bytes: clientID1, Valid: true},
		Email:        "user1@client1.com",
		FirstName:    "User1",
		LastName:     "Client1",
		Status:       "ACTIVE",
		Settings:     []byte("{}"),
	})
	require.NoError(t, err)

	t.Run("正しいクライアントIDで更新できる", func(t *testing.T) {
		params := db.UpdateClientUserParams{
			ClientUserID: pgtype.UUID{Bytes: clientUserID1, Valid: true},
			ClientID:     pgtype.UUID{Bytes: clientID1, Valid: true},
			FirstName:    "Updated",
			LastName:     "Name",
			Email:        "updated@client1.com",
			Status:       "ACTIVE",
			Settings:     []byte("{}"),
		}
		_, err := repo.Update(ctx, clientID1, params)
		assert.NoError(t, err)
	})

	t.Run("異なるクライアントIDで更新できない（セキュリティチェック）", func(t *testing.T) {
		params := db.UpdateClientUserParams{
			ClientUserID: pgtype.UUID{Bytes: clientUserID1, Valid: true},
			ClientID:     pgtype.UUID{Bytes: clientID1, Valid: true},
			FirstName:    "Hacked",
			LastName:     "Name",
			Email:        "hacked@client1.com",
			Status:       "ACTIVE",
			Settings:     []byte("{}"),
		}
		// クライアント2のIDを指定して更新を試みる
		_, err := repo.Update(ctx, clientID2, params)
		// エラーが発生するか、更新が失敗することを確認
		assert.Error(t, err, "異なるクライアントのユーザーを更新できてはいけない")
	})
}

// TestClientUserRepository_Delete_ClientIsolation 削除時のクライアント分離テスト
func TestClientUserRepository_Delete_ClientIsolation(t *testing.T) {
	t.Skip("統合テスト環境でのみ実行")

	ctx := context.Background()
	queries := db.New(nil) // 実際のデータベース接続が必要
	repo := NewClientUserRepository(queries)

	// テストデータの準備
	clientID1 := uuid.New()
	clientID2 := uuid.New()
	clientUserID1 := uuid.New()
	deletedBy := uuid.New()

	// クライアント1のユーザーを作成
	_, err := queries.CreateClientUser(ctx, db.CreateClientUserParams{
		ClientUserID: pgtype.UUID{Bytes: clientUserID1, Valid: true},
		ClientID:     pgtype.UUID{Bytes: clientID1, Valid: true},
		Email:        "user1@client1.com",
		FirstName:    "User1",
		LastName:     "Client1",
		Status:       "ACTIVE",
		Settings:     []byte("{}"),
	})
	require.NoError(t, err)

	t.Run("正しいクライアントIDで削除できる", func(t *testing.T) {
		err := repo.Delete(ctx, clientID1, clientUserID1, deletedBy)
		assert.NoError(t, err)
	})

	t.Run("異なるクライアントIDで削除できない（セキュリティチェック）", func(t *testing.T) {
		// クライアント1のユーザーIDで、クライアント2のIDを指定して削除を試みる
		// 注意: 前のテストで削除されているため、新しいユーザーを作成する必要がある
		clientUserID2 := uuid.New()
		_, err := queries.CreateClientUser(ctx, db.CreateClientUserParams{
			ClientUserID: pgtype.UUID{Bytes: clientUserID2, Valid: true},
			ClientID:     pgtype.UUID{Bytes: clientID1, Valid: true},
			Email:        "user2@client1.com",
			FirstName:    "User2",
			LastName:     "Client1",
			Status:       "ACTIVE",
			Settings:     []byte("{}"),
		})
		require.NoError(t, err)

		// クライアント2のIDを指定して削除を試みる
		err = repo.Delete(ctx, clientID2, clientUserID2, deletedBy)
		// エラーが発生するか、削除が失敗することを確認
		assert.Error(t, err, "異なるクライアントのユーザーを削除できてはいけない")
	})
}
