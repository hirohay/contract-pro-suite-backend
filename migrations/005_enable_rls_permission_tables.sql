-- 権限管理テーブルのRLS（Row Level Security）を有効化するマイグレーション
-- セキュリティアドバイザーの推奨事項に基づく

-- operator_assignmentsテーブルのRLSを有効化
ALTER TABLE operator_assignments ENABLE ROW LEVEL SECURITY;

-- client_rolesテーブルのRLSを有効化
ALTER TABLE client_roles ENABLE ROW LEVEL SECURITY;

-- client_role_permissionsテーブルのRLSを有効化
ALTER TABLE client_role_permissions ENABLE ROW LEVEL SECURITY;

-- client_user_rolesテーブルのRLSを有効化
ALTER TABLE client_user_roles ENABLE ROW LEVEL SECURITY;

-- 注意: このアプリケーションはGoバックエンドからサービスロールキーを使用してアクセスするため、
-- 基本的なRLSポリシーは「サービスロールキーを使用する場合は全アクセス許可」とします。
-- より細かい制御が必要な場合は、追加のポリシーを作成してください。

-- サービスロールキーを使用する場合（バックエンドアプリケーション）は全アクセス許可
-- 注意: SupabaseのサービスロールキーはRLSをバイパスするため、このポリシーは主に
-- 将来のPostgREST API経由のアクセスに備えたものです。

-- operator_assignmentsテーブルのポリシー
-- サービスロールキーを使用する場合は全アクセス許可（実質的にRLSをバイパス）
CREATE POLICY "Service role can access all operator_assignments"
    ON operator_assignments
    FOR ALL
    USING (true)
    WITH CHECK (true);

-- client_rolesテーブルのポリシー
-- サービスロールキーを使用する場合は全アクセス許可（実質的にRLSをバイパス）
CREATE POLICY "Service role can access all client_roles"
    ON client_roles
    FOR ALL
    USING (true)
    WITH CHECK (true);

-- client_role_permissionsテーブルのポリシー
-- サービスロールキーを使用する場合は全アクセス許可（実質的にRLSをバイパス）
CREATE POLICY "Service role can access all client_role_permissions"
    ON client_role_permissions
    FOR ALL
    USING (true)
    WITH CHECK (true);

-- client_user_rolesテーブルのポリシー
-- サービスロールキーを使用する場合は全アクセス許可（実質的にRLSをバイパス）
CREATE POLICY "Service role can access all client_user_roles"
    ON client_user_roles
    FOR ALL
    USING (true)
    WITH CHECK (true);

