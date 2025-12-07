-- 権限管理テーブル作成
-- E-R図: mvp-er-diagram.md の「1. アカウント & 権限管理」セクションに完全準拠

-- operator_assignments（オペレーター割当）テーブル
CREATE TABLE operator_assignments (
    client_id uuid NOT NULL REFERENCES clients(client_id) ON DELETE RESTRICT,
    operator_id uuid NOT NULL REFERENCES operators(operator_id) ON DELETE RESTRICT,
    role text NOT NULL CHECK (role IN ('ADMIN', 'OPERATOR', 'VIEWER')),
    status text NOT NULL DEFAULT 'ACTIVE' CHECK (status IN ('ACTIVE', 'INACTIVE', 'SUSPENDED')),
    assigned_at timestamptz NOT NULL DEFAULT now(),
    unassigned_at timestamptz,
    deleted_at timestamptz,
    deleted_by uuid,
    PRIMARY KEY (client_id, operator_id)
);

-- operator_assignmentsのインデックス
CREATE INDEX idx_operator_assignments_client_id ON operator_assignments(client_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_operator_assignments_operator_id ON operator_assignments(operator_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_operator_assignments_status ON operator_assignments(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_operator_assignments_role ON operator_assignments(role) WHERE deleted_at IS NULL;

-- client_roles（クライアントロール）テーブル
CREATE TABLE client_roles (
    role_id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    client_id uuid NOT NULL REFERENCES clients(client_id) ON DELETE RESTRICT,
    code text NOT NULL,
    name text NOT NULL,
    description text,
    is_system boolean NOT NULL DEFAULT false,
    deleted_at timestamptz,
    deleted_by uuid,
    created_at timestamptz NOT NULL DEFAULT now(),
    -- クライアント内でcodeの一意性を保証
    UNIQUE(client_id, code)
);

-- client_rolesのインデックス
CREATE INDEX idx_client_roles_client_id ON client_roles(client_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_client_roles_code ON client_roles(code) WHERE deleted_at IS NULL;
CREATE INDEX idx_client_roles_is_system ON client_roles(is_system) WHERE deleted_at IS NULL;

-- client_role_permissions（クライアントロール権限）テーブル
CREATE TABLE client_role_permissions (
    role_id uuid NOT NULL REFERENCES client_roles(role_id) ON DELETE CASCADE,
    feature text NOT NULL,
    action text NOT NULL,
    granted boolean NOT NULL DEFAULT true,
    conditions jsonb NOT NULL DEFAULT '{}',
    deleted_at timestamptz,
    deleted_by uuid,
    created_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (role_id, feature, action)
);

-- client_role_permissionsのインデックス
CREATE INDEX idx_client_role_permissions_role_id ON client_role_permissions(role_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_client_role_permissions_feature ON client_role_permissions(feature) WHERE deleted_at IS NULL;
CREATE INDEX idx_client_role_permissions_action ON client_role_permissions(action) WHERE deleted_at IS NULL;
CREATE INDEX idx_client_role_permissions_feature_action ON client_role_permissions(feature, action) WHERE deleted_at IS NULL;

-- client_user_roles（クライアントユーザーロール）テーブル
CREATE TABLE client_user_roles (
    client_id uuid NOT NULL REFERENCES clients(client_id) ON DELETE RESTRICT,
    client_user_id uuid NOT NULL REFERENCES client_users(client_user_id) ON DELETE RESTRICT,
    role_id uuid NOT NULL REFERENCES client_roles(role_id) ON DELETE RESTRICT,
    assigned_at timestamptz NOT NULL DEFAULT now(),
    revoked_at timestamptz,
    deleted_at timestamptz,
    deleted_by uuid,
    PRIMARY KEY (client_id, client_user_id, role_id)
);

-- client_user_rolesのインデックス
CREATE INDEX idx_client_user_roles_client_id ON client_user_roles(client_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_client_user_roles_client_user_id ON client_user_roles(client_user_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_client_user_roles_role_id ON client_user_roles(role_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_client_user_roles_revoked_at ON client_user_roles(revoked_at) WHERE deleted_at IS NULL AND revoked_at IS NULL;

-- updated_at自動更新トリガーをclient_rolesに設定
CREATE TRIGGER update_client_roles_updated_at BEFORE UPDATE ON client_roles
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

