package fx

import (
	"go.uber.org/fx"

	"contract-pro-suite/internal/shared/config"
	"contract-pro-suite/internal/shared/db"
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
		fx.Provide(func(queries *dbgen.Queries) repository.OperatorAssignmentRepository {
			return repository.NewOperatorAssignmentRepository(queries)
		}),
		fx.Provide(func(queries *dbgen.Queries) repository.ClientRoleRepository {
			return repository.NewClientRoleRepository(queries)
		}),
		fx.Provide(func(queries *dbgen.Queries) repository.ClientRolePermissionRepository {
			return repository.NewClientRolePermissionRepository(queries)
		}),
		fx.Provide(func(queries *dbgen.Queries) repository.ClientUserRoleRepository {
			return repository.NewClientUserRoleRepository(queries)
		}),
		// ユースケースの提供
		fx.Provide(func(
			operatorRepo repository.OperatorRepository,
			clientUserRepo repository.ClientUserRepository,
			clientRepo repository.ClientRepository,
			operatorAssignmentRepo repository.OperatorAssignmentRepository,
			clientRoleRepo repository.ClientRoleRepository,
			clientRolePermissionRepo repository.ClientRolePermissionRepository,
			clientUserRoleRepo repository.ClientUserRoleRepository,
			cfg *config.Config,
			database *db.DB,
		) usecase.AuthUsecase {
			return usecase.NewAuthUsecase(
				operatorRepo,
				clientUserRepo,
				clientRepo,
				operatorAssignmentRepo,
				clientRoleRepo,
				clientRolePermissionRepo,
				clientUserRoleRepo,
				cfg,
				database,
			)
		}),
		// gRPCサーバーの提供
		fx.Provide(server.NewAuthServer),
	)
}

