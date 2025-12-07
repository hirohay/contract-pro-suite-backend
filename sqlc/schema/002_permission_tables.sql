-- 権限管理テーブルのスキーマ定義
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
    UNIQUE(client_id, code)
);

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

