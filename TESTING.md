# APIサーバー起動と動作確認手順

## 前提条件

1. `.env`ファイルが作成され、必要な環境変数が設定されていること
2. Supabaseでマイグレーションが実行済みであること
3. Go 1.25以上がインストールされていること

## APIサーバーの起動

### 方法1: 直接実行

```bash
cd backend
go run ./cmd/api
```

### 方法2: ビルドして実行

```bash
cd backend
go build ./cmd/api
./api
```

サーバーが正常に起動すると、以下のメッセージが表示されます：

```
Server started on port 8080
```

## 動作確認

### 1. ヘルスチェックエンドポイント

データベース接続を含むヘルスチェック：

```bash
curl http://localhost:8080/health
```

**期待されるレスポンス**:
- ステータスコード: `200 OK`
- レスポンスボディ: `OK`

**エラー時のレスポンス**:
- ステータスコード: `503 Service Unavailable`
- レスポンスボディ: `Database connection failed`

### 2. 認証エンドポイント（JWTトークン必要）

現在のユーザー情報を取得：

```bash
curl -H "Authorization: Bearer YOUR_JWT_TOKEN" http://localhost:8080/api/v1/auth/me
```

**期待されるレスポンス** (正常時):
```json
{
  "user_id": "uuid",
  "user_type": "OPERATOR" | "CLIENT_USER",
  "email": "user@example.com",
  "client_id": "uuid" // オプション
}
```

**エラー時のレスポンス**:
- ステータスコード: `401 Unauthorized` (JWTトークンがない、または無効)
- ステータスコード: `403 Forbidden` (ユーザー情報の取得に失敗)

### 3. JWTトークンの取得方法

Supabase Authを使用してJWTトークンを取得：

1. Supabaseダッシュボードでユーザーを作成
2. Supabase Auth APIを使用してログイン
3. レスポンスから`access_token`を取得

**例（curl）**:
```bash
curl -X POST 'https://[YOUR-PROJECT-REF].supabase.co/auth/v1/token?grant_type=password' \
  -H "apikey: YOUR_ANON_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password"
  }'
```

レスポンスから`access_token`を取得し、APIリクエストの`Authorization`ヘッダーに設定します。

## トラブルシューティング

### データベース接続エラー

**エラー**: `Failed to connect to database`

**解決方法**:
1. `.env`ファイルの`SUPABASE_DB_URL`が正しいか確認
2. Supabaseのデータベースが起動しているか確認
3. ファイアウォール設定を確認

### JWT検証エラー

**エラー**: `Invalid token`

**解決方法**:
1. `.env`ファイルの`SUPABASE_JWT_SECRET`が正しいか確認
2. JWTトークンが有効期限内か確認
3. JWTトークンがSupabaseから発行されたものか確認

### ユーザー情報取得エラー

**エラー**: `user not found`

**解決方法**:
1. Supabase Authでユーザーが作成されているか確認
2. `operators`または`client_users`テーブルにユーザー情報が登録されているか確認
3. JWTトークンの`sub`（ユーザーID）がデータベースの`operator_id`または`client_user_id`と一致するか確認

## 開発時の注意事項

- サーバーを停止するには`Ctrl+C`を押します
- グレースフルシャットダウンが実装されているため、安全に停止できます
- ログは標準出力に表示されます

