package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"contract-pro-suite/internal/shared/config"
)

// DB データベース接続プール
type DB struct {
	Pool *pgxpool.Pool
}

// New データベース接続プールを作成
func New(cfg *config.Config) (*DB, error) {
	poolConfig, err := pgxpool.ParseConfig(cfg.SupabaseDBURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database URL: %w", err)
	}

	// IPv6接続を許可（デフォルトのDialFuncを使用）
	// DialFuncを設定しないことで、pgxが自動的にIPv6/IPv4を選択

	// 接続プール設定
	poolConfig.MaxConns = int32(cfg.DBMaxConns)
	poolConfig.MinConns = int32(cfg.DBMinConns)
	poolConfig.MaxConnLifetime = cfg.DBMaxConnLifetime
	poolConfig.MaxConnIdleTime = cfg.DBMaxConnIdleTime
	
	// Supabaseの接続プーラー（pgbouncer）を使用している場合、
	// transactionモードでは準備済みステートメントの使用に制限がある
	// 接続URLがpooler.supabase.com:6543の場合は、直接接続（db.supabase.com:5432）に変更することを推奨
	// または、pgbouncerのsessionモードを使用する

	// 接続プールを作成
	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// 接続テスト
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DB{Pool: pool}, nil
}

// Close データベース接続を閉じる
func (db *DB) Close() {
	if db.Pool != nil {
		db.Pool.Close()
	}
}

// HealthCheck データベースのヘルスチェック
func (db *DB) HealthCheck(ctx context.Context) error {
	return db.Pool.Ping(ctx)
}

