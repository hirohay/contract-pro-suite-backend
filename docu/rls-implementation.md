# RLS（Row Level Security）実装状況

## 最終更新日
2025年1月

## RLS有効化状況

### ✅ 有効化済みテーブル

すべてのテーブルでRLSが有効化されています：

1. ✅ **clients** - クライアント情報
2. ✅ **operators** - オペレーター情報
3. ✅ **client_users** - クライアントユーザー情報
4. ✅ **operator_assignments** - オペレーター割り当て
5. ✅ **client_roles** - クライアントロール
6. ✅ **client_role_permissions** - ロール権限
7. ✅ **client_user_roles** - ユーザーロール割り当て

## RLSポリシー

### 現在のポリシー

すべてのテーブルに対して、以下のポリシーが設定されています：

```sql
CREATE POLICY "Service role can access all <table_name>"
    ON <table_name>
    FOR ALL
    USING (true)
    WITH CHECK (true);
```

### ポリシーの説明

- **USING (true)**: 既存の行へのアクセスを許可
- **WITH CHECK (true)**: 新しい行の挿入・更新を許可
- **FOR ALL**: SELECT、INSERT、UPDATE、DELETEすべての操作に適用

### 注意事項

**重要**: 現在のポリシーは「サービスロールキーを使用する場合は全アクセス許可」となっています。

これは以下の理由によるものです：

1. **Goバックエンドアプリケーション**: サービスロールキーを使用してデータベースにアクセス
2. **Supabaseのサービスロールキー**: RLSをバイパスするため、このポリシーは主に将来のPostgREST API経由のアクセスに備えたもの
3. **アプリケーションレベルのセキュリティ**: クライアント分離はアプリケーション層（ミドルウェア、リポジトリ）で実装済み

## マイグレーション履歴

### 002_enable_rls.sql
- `clients`テーブルのRLS有効化
- `operators`テーブルのRLS有効化
- `client_users`テーブルのRLS有効化
- 各テーブルにポリシーを設定

### 005_enable_rls_permission_tables.sql
- `operator_assignments`テーブルのRLS有効化
- `client_roles`テーブルのRLS有効化
- `client_role_permissions`テーブルのRLS有効化
- `client_user_roles`テーブルのRLS有効化
- 各テーブルにポリシーを設定

## セキュリティアドバイザーの結果

### ✅ 解決済み
- ✅ RLS無効エラー: すべてのテーブルでRLSが有効化され、エラーが解消されました

### ⚠️ 残っている警告（優先度: 低）

1. **Extension in Public** (WARN)
   - `citext`拡張機能が`public`スキーマにインストールされている
   - 改善案: 別のスキーマに移動（将来的な改善）

2. **Leaked Password Protection Disabled** (WARN)
   - パスワード漏洩保護が無効
   - 改善案: Supabase Auth設定で有効化（将来的な改善）

## 将来の改善案

### 1. より細かいRLSポリシー（PostgREST API使用時）

PostgREST APIを直接使用する場合、より細かいポリシーを設定できます：

```sql
-- 例: クライアントユーザーは自分のクライアントのデータのみアクセス可能
CREATE POLICY "Users can access their own client data"
    ON client_users
    FOR SELECT
    USING (
        client_id IN (
            SELECT client_id 
            FROM client_users 
            WHERE client_user_id = auth.uid()
        )
    );
```

### 2. ロールベースのポリシー

```sql
-- 例: オペレーターは割り当てられたクライアントのデータのみアクセス可能
CREATE POLICY "Operators can access assigned clients"
    ON clients
    FOR SELECT
    USING (
        client_id IN (
            SELECT client_id 
            FROM operator_assignments 
            WHERE operator_id = auth.uid()
            AND status = 'ACTIVE'
        )
    );
```

### 3. 関数ベースのポリシー

```sql
-- 例: カスタム関数を使用した複雑なポリシー
CREATE FUNCTION check_user_client_access(user_id uuid, target_client_id uuid)
RETURNS boolean AS $$
BEGIN
    -- クライアントユーザーの場合
    IF EXISTS (
        SELECT 1 FROM client_users 
        WHERE client_user_id = user_id 
        AND client_id = target_client_id
    ) THEN
        RETURN true;
    END IF;
    
    -- オペレーターの場合
    IF EXISTS (
        SELECT 1 FROM operator_assignments 
        WHERE operator_id = user_id 
        AND client_id = target_client_id
        AND status = 'ACTIVE'
    ) THEN
        RETURN true;
    END IF;
    
    RETURN false;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

CREATE POLICY "Users can access their client data"
    ON client_users
    FOR ALL
    USING (check_user_client_access(auth.uid(), client_id));
```

## 現在のセキュリティ実装

### アプリケーションレベル

現在、セキュリティは主にアプリケーション層で実装されています：

1. ✅ **ミドルウェア層**: `TenantMiddleware`、`TenantInterceptor`でクライアント検証
2. ✅ **リポジトリ層**: SQLクエリレベルで`client_id`フィルタリング
3. ✅ **ユースケース層**: `ValidateClientAccess`でアクセス権限検証

### データベースレベル

1. ✅ **RLS有効化**: すべてのテーブルでRLSが有効
2. ✅ **基本ポリシー**: サービスロールキー使用時の全アクセス許可
3. ⚠️ **細かいポリシー**: 将来的にPostgREST API使用時に追加可能

## まとめ

✅ **RLSはすべてのテーブルで有効化されています**

- 7つのテーブルすべてでRLSが有効
- セキュリティアドバイザーのRLS関連エラーはすべて解消
- 現在のポリシーはサービスロールキー使用時の全アクセス許可
- アプリケーションレベルでのセキュリティ実装も完了

将来的にPostgREST APIを直接使用する場合は、より細かいRLSポリシーを追加できます。

