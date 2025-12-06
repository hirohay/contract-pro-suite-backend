# データベースマイグレーション手順

## Supabaseでのマイグレーション実行

### 方法1: SupabaseダッシュボードのSQL Editorを使用

1. Supabaseダッシュボードにログイン
2. プロジェクトを選択
3. 左メニューから「SQL Editor」を選択
4. 「New query」をクリック
5. `migrations/001_initial_auth_tables.sql`の内容をコピー＆ペースト
6. 「Run」ボタンをクリックして実行

### 方法2: Supabase CLIを使用

```bash
# Supabase CLIがインストールされている場合
supabase db push
```

## 環境変数の設定

`backend/.env`ファイルを作成し、以下の環境変数を設定してください：

```bash
# アプリケーション設定
APP_ENV=development
APP_PORT=8080

# Supabase設定
# Supabaseダッシュボードの「Settings」→「Database」から取得
SUPABASE_DB_URL=postgresql://postgres:[YOUR-PASSWORD]@db.[YOUR-PROJECT-REF].supabase.co:5432/postgres
# Supabaseダッシュボードの「Settings」→「API」から取得
SUPABASE_SERVICE_ROLE_KEY=your-supabase-service-role-key
# Supabaseダッシュボードの「Settings」→「API」→「JWT Settings」から取得
SUPABASE_JWT_SECRET=your-supabase-jwt-secret
# SupabaseプロジェクトURL
SUPABASE_URL=https://[YOUR-PROJECT-REF].supabase.co

# テナント設定
DEFAULT_CLIENT_ID=00000000-0000-0000-0000-000000000000

# CORS設定
CORS_ORIGIN=http://localhost:3001

# データベース接続設定
DB_MAX_CONNS=25
DB_MIN_CONNS=5
DB_MAX_CONN_LIFETIME=5m
DB_MAX_CONN_IDLE_TIME=1m
```

### 環境変数の取得方法

1. **SUPABASE_DB_URL**: 
   - Supabaseダッシュボード → Settings → Database
   - 「Connection string」の「URI」をコピー
   - `[YOUR-PASSWORD]`を実際のパスワードに置き換え

2. **SUPABASE_SERVICE_ROLE_KEY**:
   - Supabaseダッシュボード → Settings → API
   - 「Project API keys」の「service_role」キーをコピー

3. **SUPABASE_JWT_SECRET**:
   - Supabaseダッシュボード → Settings → API
   - 「JWT Settings」の「JWT Secret」をコピー

4. **SUPABASE_URL**:
   - Supabaseダッシュボード → Settings → API
   - 「Project URL」をコピー

## 注意事項

- `.env`ファイルは`.gitignore`に含まれているため、Gitにコミットされません
- 本番環境では環境変数を直接設定するか、適切なシークレット管理サービスを使用してください

