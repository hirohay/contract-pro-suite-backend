# gRPCサーバー起動と動作確認手順

## 前提条件

1. `.env`ファイルが作成され、必要な環境変数が設定されていること
2. Supabaseでマイグレーションが実行済みであること
3. Go 1.25以上がインストールされていること
4. `grpcurl`がインストールされていること（テスト用）

## gRPCサーバーの起動

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
gRPC server starting on port 8081
```

## 動作確認

### 1. gRPCサービスのリスト確認

利用可能なgRPCサービスを確認：

```bash
grpcurl -plaintext localhost:8081 list
```

**期待される出力**:
```
auth.AuthService
```

### 2. 認証エンドポイント（JWTトークン必要）

現在のユーザー情報を取得：

```bash
grpcurl -plaintext \
  -H "authorization: Bearer YOUR_JWT_TOKEN" \
  -H "x-client-id: YOUR_CLIENT_ID" \
  -d '{}' \
  localhost:8081 \
  auth.AuthService/GetMe
```

**期待されるレスポンス** (正常時):
```json
{
  "userId": "uuid",
  "email": "user@example.com",
  "userType": "OPERATOR" | "CLIENT_USER",
  "clientId": "uuid" // オプション
}
```

**エラー時のレスポンス**:
- gRPCステータスコード: `UNAUTHENTICATED` (JWTトークンがない、または無効)
- gRPCステータスコード: `PERMISSION_DENIED` (ユーザー情報の取得に失敗、またはクライアントアクセス権限なし)

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

