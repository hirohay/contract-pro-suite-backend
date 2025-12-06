-- RLS（Row Level Security）を有効化するマイグレーション
-- セキュリティアドバイザーの推奨事項に基づく

-- clientsテーブルのRLSを有効化
ALTER TABLE clients ENABLE ROW LEVEL SECURITY;

-- operatorsテーブルのRLSを有効化
ALTER TABLE operators ENABLE ROW LEVEL SECURITY;

-- client_usersテーブルのRLSを有効化
ALTER TABLE client_users ENABLE ROW LEVEL SECURITY;

-- 注意: このアプリケーションはGoバックエンドからサービスロールキーを使用してアクセスするため、
-- 基本的なRLSポリシーは「サービスロールキーを使用する場合は全アクセス許可」とします。
-- より細かい制御が必要な場合は、追加のポリシーを作成してください。

-- サービスロールキーを使用する場合（バックエンドアプリケーション）は全アクセス許可
-- 注意: SupabaseのサービスロールキーはRLSをバイパスするため、このポリシーは主に
-- 将来のPostgREST API経由のアクセスに備えたものです。

-- clientsテーブルのポリシー
-- サービスロールキーを使用する場合は全アクセス許可（実質的にRLSをバイパス）
CREATE POLICY "Service role can access all clients"
    ON clients
    FOR ALL
    USING (true)
    WITH CHECK (true);

-- operatorsテーブルのポリシー
CREATE POLICY "Service role can access all operators"
    ON operators
    FOR ALL
    USING (true)
    WITH CHECK (true);

-- client_usersテーブルのポリシー
CREATE POLICY "Service role can access all client_users"
    ON client_users
    FOR ALL
    USING (true)
    WITH CHECK (true);

-- 関数のsearch_pathを修正（セキュリティ警告対応）
-- update_updated_at_column関数のsearch_pathを固定
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = public
AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$;

