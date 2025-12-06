package config

import (
	"fmt"
	"os"
	"time"

	"github.com/kelseyhightower/envconfig"
)

// Config アプリケーション設定
type Config struct {
	// アプリケーション設定
	AppEnv  string `envconfig:"APP_ENV" default:"development"`
	AppPort string `envconfig:"APP_PORT" default:"8080"`

	// Supabase設定
	SupabaseDBURL          string `envconfig:"SUPABASE_DB_URL" required:"true"`
	SupabaseServiceRoleKey string `envconfig:"SUPABASE_SERVICE_ROLE_KEY" required:"true"`
	SupabaseJWTSecret     string `envconfig:"SUPABASE_JWT_SECRET" required:"true"`
	SupabaseURL            string `envconfig:"SUPABASE_URL" required:"true"`

	// テナント設定
	DefaultClientID string `envconfig:"DEFAULT_CLIENT_ID" default:"00000000-0000-0000-0000-000000000000"`

	// CORS設定
	CORSOrigin string `envconfig:"CORS_ORIGIN" default:"http://localhost:3001"`

	// データベース接続設定
	DBMaxConns        int           `envconfig:"DB_MAX_CONNS" default:"25"`
	DBMinConns        int           `envconfig:"DB_MIN_CONNS" default:"5"`
	DBMaxConnLifetime time.Duration `envconfig:"DB_MAX_CONN_LIFETIME" default:"5m"`
	DBMaxConnIdleTime time.Duration `envconfig:"DB_MAX_CONN_IDLE_TIME" default:"1m"`
}

// Load 環境変数から設定を読み込む
func Load() (*Config, error) {
	var cfg Config

	// .envファイルがあれば読み込む（開発環境用）
	// 本番環境では環境変数を直接設定することを想定
	if _, err := os.Stat(".env"); err == nil {
		// .envファイルの読み込みは外部ライブラリ（godotenv等）を使用する場合
		// 今回は環境変数から直接読み込む方式を採用
	}

	if err := envconfig.Process("", &cfg); err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	return &cfg, nil
}

// Validate 設定の妥当性を検証
func (c *Config) Validate() error {
	if c.SupabaseDBURL == "" {
		return fmt.Errorf("SUPABASE_DB_URL is required")
	}
	if c.SupabaseServiceRoleKey == "" {
		return fmt.Errorf("SUPABASE_SERVICE_ROLE_KEY is required")
	}
	if c.SupabaseJWTSecret == "" {
		return fmt.Errorf("SUPABASE_JWT_SECRET is required")
	}
	if c.SupabaseURL == "" {
		return fmt.Errorf("SUPABASE_URL is required")
	}
	return nil
}

