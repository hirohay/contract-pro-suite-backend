package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	authmiddleware "contract-pro-suite/internal/middleware"
	"contract-pro-suite/internal/shared/config"
	"contract-pro-suite/internal/shared/db"
	"contract-pro-suite/services/auth/handler"
	"contract-pro-suite/services/auth/repository"
	"contract-pro-suite/services/auth/usecase"
	dbgen "contract-pro-suite/sqlc"
)

func main() {
	// 設定を読み込む
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if err := cfg.Validate(); err != nil {
		log.Fatalf("Invalid config: %v", err)
	}

	// データベース接続
	database, err := db.New(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	// sqlcクエリを作成
	queries := dbgen.New(database.Pool)

	// リポジトリを作成
	clientRepo := repository.NewClientRepository(queries)
	operatorRepo := repository.NewOperatorRepository(queries)
	clientUserRepo := repository.NewClientUserRepository(queries)

	// ユースケースを作成
	authUsecase := usecase.NewAuthUsecase(operatorRepo, clientUserRepo, clientRepo)

	// ハンドラを作成
	authHandler := handler.NewAuthHandler(authUsecase)

	// ルーターを設定
	r := chi.NewRouter()

	// ミドルウェア
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// CORS設定（簡易版）
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", cfg.CORSOrigin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
			w.Header().Set("Access-Control-Allow-Credentials", "true")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	})

	// APIルート
	r.Route("/api/v1", func(r chi.Router) {
		// 認証ミドルウェア
		r.Use(authmiddleware.AuthMiddleware(cfg))
		r.Use(authmiddleware.EnhancedAuthMiddleware(authUsecase))

		// 認証ハンドラ
		authHandler.RegisterRoutes(r)
	})

	// ヘルスチェック
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		if err := database.HealthCheck(ctx); err != nil {
			http.Error(w, "Database connection failed", http.StatusServiceUnavailable)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// サーバーを起動
	srv := &http.Server{
		Addr:    ":" + cfg.AppPort,
		Handler: r,
	}

	// グレースフルシャットダウン
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	log.Printf("Server started on port %s", cfg.AppPort)

	// シグナル待機
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

