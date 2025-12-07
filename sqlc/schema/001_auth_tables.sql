-- 認証関連テーブルのスキーマ定義
-- E-R図: mvp-er-diagram.md に完全準拠

-- UUID拡張機能を有効化
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- citext拡張機能を有効化（大文字小文字を区別しないテキスト型）
CREATE EXTENSION IF NOT EXISTS "citext";

-- clients（クライアント）テーブル
CREATE TABLE clients (
    client_id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    slug text NOT NULL UNIQUE,
    company_code text UNIQUE,  -- オプション（JIPDEC標準企業コード）
    name text NOT NULL,
    e_sign_mode text NOT NULL DEFAULT 'WITNESS_OTP' CHECK (e_sign_mode IN ('WITNESS_OTP', 'OTP_ONLY', 'CERTIFICATE', 'BIOMETRIC', 'SIMPLE_CLICK')),
    retention_default_months integer NOT NULL DEFAULT 84 CHECK (retention_default_months >= 12 AND retention_default_months <= 240),
    status text NOT NULL DEFAULT 'ACTIVE' CHECK (status IN ('ACTIVE', 'SUSPENDED', 'TERMINATED')),
    settings jsonb NOT NULL DEFAULT '{}',
    deleted_at timestamptz,
    deleted_by uuid,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

-- slugのインデックス（既にUNIQUE制約でインデックスが作成されるが、明示的に作成）
CREATE INDEX idx_clients_slug ON clients(slug) WHERE deleted_at IS NULL;
CREATE INDEX idx_clients_company_code ON clients(company_code) WHERE deleted_at IS NULL;
CREATE INDEX idx_clients_status ON clients(status) WHERE deleted_at IS NULL;

-- operators（オペレーター）テーブル
CREATE TABLE operators (
    operator_id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    email citext NOT NULL UNIQUE,
    first_name text NOT NULL,
    last_name text NOT NULL,
    status text NOT NULL DEFAULT 'ACTIVE' CHECK (status IN ('ACTIVE', 'INACTIVE', 'SUSPENDED')),
    mfa_enabled boolean NOT NULL DEFAULT false,
    password_hash text,
    salt text,
    last_login_at timestamptz,
    password_changed_at timestamptz,
    deleted_at timestamptz,
    deleted_by uuid,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

-- operatorsのインデックス
CREATE INDEX idx_operators_email ON operators(email) WHERE deleted_at IS NULL;
CREATE INDEX idx_operators_status ON operators(status) WHERE deleted_at IS NULL;

-- client_users（クライアントユーザー）テーブル
CREATE TABLE client_users (
    client_user_id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    client_id uuid NOT NULL REFERENCES clients(client_id) ON DELETE RESTRICT,
    email citext NOT NULL,
    first_name text NOT NULL,
    last_name text NOT NULL,
    department text,
    position text,
    settings jsonb NOT NULL DEFAULT '{}',
    status text NOT NULL DEFAULT 'ACTIVE' CHECK (status IN ('ACTIVE', 'INACTIVE', 'SUSPENDED')),
    deleted_at timestamptz,
    deleted_by uuid,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    -- クライアント内でemailの一意性を保証
    UNIQUE(client_id, email)
);

-- client_usersのインデックス
CREATE INDEX idx_client_users_client_id ON client_users(client_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_client_users_email ON client_users(email) WHERE deleted_at IS NULL;
CREATE INDEX idx_client_users_status ON client_users(status) WHERE deleted_at IS NULL;

-- updated_atを自動更新するトリガー関数
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- 各テーブルにupdated_at自動更新トリガーを設定
CREATE TRIGGER update_clients_updated_at BEFORE UPDATE ON clients
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_operators_updated_at BEFORE UPDATE ON operators
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_client_users_updated_at BEFORE UPDATE ON client_users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

