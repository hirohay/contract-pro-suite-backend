# ContractProSuite Backend

ContractProSuiteのGoバックエンドアプリケーションです。

## 概要

モジュラーモノリス構成のGoアプリケーションで、以下の技術スタックを使用しています：

- **Language**: Go 1.25+
- **HTTP Framework**: chi/v5
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
│   ├── events/          # Pub/Sub抽象
│   └── middleware/     # 認証・認可・ロギング
├── services/
│   └── auth/            # 認証サービス
│       ├── handler/    # HTTPハンドラ
│       ├── usecase/    # ビジネスロジック
│       ├── repository/ # データアクセス
│       └── domain/     # ドメインモデル
├── migrations/          # データベースマイグレーション
├── sqlc/               # sqlc設定・生成物
└── pkg/                # 共通ライブラリ
```

## セットアップ

### 前提条件

- Go 1.25以上
- PostgreSQL (Supabase)
- sqlc (インストール済み)

### 環境変数

`.env` ファイルを作成し、以下の環境変数を設定してください：

```bash
# Supabase
SUPABASE_DB_URL=postgresql://<user>:<password>@<host>:5432/<db>
SUPABASE_SERVICE_ROLE_KEY=your-service-role-key

# アプリ設定
APP_PORT=8080
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

