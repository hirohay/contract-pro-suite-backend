#!/bin/bash

# テストデータ準備スクリプト
# テスト実行前に必要なデータ（クライアント、オペレーター）を準備します

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

cd "$BACKEND_DIR"

# 環境変数の読み込み
if [ ! -f .env ]; then
    echo "Error: .env file not found"
    exit 1
fi

export $(cat .env | grep -v '^#' | xargs)

echo "=========================================="
echo "Preparing Test Data"
echo "=========================================="

# 1. テスト用クライアントの作成
echo ""
echo "1. Creating test client..."
CLIENT_SQL="INSERT INTO clients (client_id, name, company_code, slug, status)
VALUES ('00000000-0000-0000-0000-000000000000'::uuid, 'Test Client', 'TEST', 'test', 'ACTIVE')
ON CONFLICT (client_id) DO UPDATE SET 
    name = EXCLUDED.name, 
    status = EXCLUDED.status, 
    updated_at = now();"

echo "SQL: $CLIENT_SQL"
echo "Note: Execute this SQL manually using Supabase MCP or SQL Editor"

# 2. テスト用オペレーターの確認（既存のオペレーターを確認）
echo ""
echo "2. Checking existing operators..."
echo "Note: Operators will be registered during test execution"
echo ""

echo "Test data preparation completed!"
echo "Note: Execute the SQL above using Supabase MCP or SQL Editor before running tests"

