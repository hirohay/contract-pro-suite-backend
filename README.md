# ContractProSuite Backend

ContractProSuiteのGoバックエンドアプリケーションです。

## 概要

モジュラーモノリス構成のGoアプリケーションで、以下の技術スタックを使用しています：

- **Language**: Go 1.25+
- **API Protocol**: gRPC (Protocol Buffers + google.golang.org/grpc, Port 8081)
- **Dependency Injection**: uber-go/fx (ランタイムDI、モジュールベース)
- **Database**: PostgreSQL (Supabase) + pgx/v5
- **Query Builder**: sqlc
- **Authentication**: Supabase Auth (gotrue-go)
- **Testing**: testify

## プロジェクト構造

```
backend/
├── cmd/
│   └── api/              # エントリポイント (main.go)
├── internal/
│   ├── shared/          # ロガー/設定/DB接続/トレース/メトリクス
│   │   ├── config/      # 設定管理
│   │   ├── db/          # データベース接続
│   │   └── fx/          # fxモジュール定義
│   ├── events/          # Pub/Sub抽象
│   └── interceptor/     # gRPCインターセプター（認証・認可・ロギング）
├── services/
│   └── auth/            # 認証サービス
│       ├── server/      # gRPCサーバー
│       ├── client/      # 外部APIクライアント
│       ├── usecase/     # ビジネスロジック
│       ├── repository/  # データアクセス
│       ├── domain/      # ドメインモデル
│       └── fx/          # fxモジュール定義
├── proto/               # Protocol Buffers定義
│   └── auth/           # 認証サービスのprotoファイル
├── migrations/          # データベースマイグレーション
├── sqlc/               # sqlc設定・生成物
└── pkg/                # 共通ライブラリ
```

## セットアップ

### 前提条件

- Go 1.25以上
- PostgreSQL (Supabase)
- sqlc (インストール済み)
- protoc (Protocol Buffersコンパイラ)

### 環境変数

`.env` ファイルを作成し、以下の環境変数を設定してください：

```bash
# Supabase
SUPABASE_DB_URL=postgresql://<user>:<password>@<host>:5432/<db>
SUPABASE_SERVICE_ROLE_KEY=your-service-role-key

# アプリ設定
GRPC_PORT=8081         # gRPC サーバーポート
APP_ENV=development
DEFAULT_CLIENT_ID=00000000-0000-0000-0000-000000000000
```

### 依存パッケージのインストール

```bash
go mod download
```

### ビルド

```bash
go build ./cmd/api
```

### 実行

```bash
./api
```

サーバーはgRPCサーバーとして起動します：
- **gRPC Server**: `localhost:8081`

### Protocol Buffersコード生成

Protocol BuffersからGoコードを生成する必要があります：

```bash
# protocのインストール（macOSの場合）
brew install protobuf

# Protocol Buffersコードの生成
protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       proto/auth/auth.proto
```

または、Makefileを使用する場合：

```bash
make generate-proto
```

### 依存性注入（DI）

このプロジェクトでは`uber-go/fx`を使用して依存性注入を行います：

```bash
# fxのインストール
go get go.uber.org/fx
```

各サービスはfxモジュールとして定義され、`main.go`で統合されます。

## 開発

### テスト実行

```bash
go test ./...
```

### コードフォーマット

```bash
go fmt ./...
```

### リンター

```bash
golangci-lint run
```

## マイグレーション

データベースマイグレーションの実行方法については、[MIGRATION.md](MIGRATION.md)を参照してください。

## テスト

APIサーバーの起動と動作確認については、[TESTING.md](TESTING.md)を参照してください。

```bash
golangci-lint run
```

## ライセンス

Copyright (c) ContractProSuite開発チーム

