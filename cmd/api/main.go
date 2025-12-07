package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/fx"

	"contract-pro-suite/internal/interceptor"
	"contract-pro-suite/internal/shared/config"
	sharedfx "contract-pro-suite/internal/shared/fx"
	pbauth "contract-pro-suite/proto/proto/auth"
	authfx "contract-pro-suite/services/auth/fx"
	"contract-pro-suite/services/auth/repository"
	"contract-pro-suite/services/auth/server"
	"contract-pro-suite/services/auth/usecase"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	app := fx.New(
		// 共有モジュール（config、db、sqlc）
		sharedfx.NewSharedModule(),
		// 認証サービスモジュール
		authfx.NewAuthModule(),
		// gRPCサーバーの起動
		fx.Invoke(startGRPCServer),
		// グレースフルシャットダウン
		fx.Invoke(registerShutdown),
	)

	app.Run()
}

// startGRPCServer gRPCサーバーを起動
func startGRPCServer(
	lc fx.Lifecycle,
	cfg *config.Config,
	authServer *server.AuthServer,
	authUsecase usecase.AuthUsecase,
	clientRepo repository.ClientRepository,
) error {
	// gRPCサーバーの作成
	grpcServer := grpc.NewServer(
		// インターセプターの適用順序が重要
		grpc.ChainUnaryInterceptor(
			// 1. 監査ログインターセプター（最初に適用）
			interceptor.AuditInterceptor(),
			// 2. JWT検証インターセプター
			interceptor.AuthInterceptor(cfg),
			// 3. ユーザー情報取得インターセプター
			interceptor.EnhancedAuthInterceptor(authUsecase),
			// 4. テナント検証インターセプター
			interceptor.TenantInterceptor(cfg, clientRepo, authUsecase),
		),
	)

	// 認証サービスを登録
	pbauth.RegisterAuthServiceServer(grpcServer, authServer)

	// gRPCリフレクションを有効化（開発環境用、テスト用）
	reflection.Register(grpcServer)

	// リスナーの作成
	lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	// ライフサイクル管理
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			log.Printf("gRPC server starting on port %s", cfg.GRPCPort)
			go func() {
				if err := grpcServer.Serve(lis); err != nil {
					log.Fatalf("gRPC server failed to serve: %v", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Println("Shutting down gRPC server...")
			grpcServer.GracefulStop()
			return nil
		},
	})

	return nil
}

// registerShutdown グレースフルシャットダウンを登録
func registerShutdown(lc fx.Lifecycle) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			// シグナル待機を別goroutineで実行
			go func() {
				quit := make(chan os.Signal, 1)
				signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
				<-quit
				log.Println("Received shutdown signal")
				os.Exit(0)
			}()
			return nil
		},
	})
}
