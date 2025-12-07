package fx

import (
	"go.uber.org/fx"

	"contract-pro-suite/internal/shared/config"
	"contract-pro-suite/internal/shared/db"
	dbgen "contract-pro-suite/sqlc"
)

// NewSharedModule 共有モジュール（config、db、sqlc）を作成
func NewSharedModule() fx.Option {
	return fx.Options(
		// 設定の読み込み
		fx.Provide(config.Load),
		// データベース接続
		fx.Provide(db.New),
		// sqlcクエリの作成
		fx.Provide(func(database *db.DB) *dbgen.Queries {
			return dbgen.New(database.Pool)
		}),
	)
}

