package fx

import (
	"go.uber.org/fx"

	"contract-pro-suite/services/auth/repository"
	"contract-pro-suite/services/auth/server"
	"contract-pro-suite/services/auth/usecase"
	dbgen "contract-pro-suite/sqlc"
)

// NewAuthModule 認証サービスのfxモジュールを作成
func NewAuthModule() fx.Option {
	return fx.Options(
		// リポジトリの提供
		fx.Provide(func(queries *dbgen.Queries) repository.ClientRepository {
			return repository.NewClientRepository(queries)
		}),
		fx.Provide(func(queries *dbgen.Queries) repository.OperatorRepository {
			return repository.NewOperatorRepository(queries)
		}),
		fx.Provide(func(queries *dbgen.Queries) repository.ClientUserRepository {
			return repository.NewClientUserRepository(queries)
		}),
		// ユースケースの提供
		fx.Provide(func(
			operatorRepo repository.OperatorRepository,
			clientUserRepo repository.ClientUserRepository,
			clientRepo repository.ClientRepository,
		) usecase.AuthUsecase {
			return usecase.NewAuthUsecase(operatorRepo, clientUserRepo, clientRepo)
		}),
		// gRPCサーバーの提供
		fx.Provide(server.NewAuthServer),
	)
}

