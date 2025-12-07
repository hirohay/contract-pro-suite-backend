-- company_codeを任意項目（NULL許可）に変更
-- JIPDEC標準企業コードは任意項目とする

-- UNIQUE制約を削除（NULL値を許可するため）
ALTER TABLE clients DROP CONSTRAINT IF EXISTS clients_company_code_key;

-- NOT NULL制約を削除
ALTER TABLE clients ALTER COLUMN company_code DROP NOT NULL;

-- NULL値を許可するUNIQUE制約を追加（NULL値はUNIQUE制約の対象外）
-- ただし、PostgreSQLではNULL値はUNIQUE制約の対象外なので、通常のUNIQUE制約でOK
-- 複数のNULL値が許可されるが、非NULL値は一意である必要がある
CREATE UNIQUE INDEX idx_clients_company_code_unique ON clients(company_code) WHERE company_code IS NOT NULL AND deleted_at IS NULL;

-- 既存のインデックスを削除（新しいUNIQUEインデックスで代替）
DROP INDEX IF EXISTS idx_clients_company_code;

