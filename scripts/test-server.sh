#!/bin/bash

# APIサーバーの起動とテスト実行スクリプト
# ユーザー作成から認証テストまで一連のフローを実行

set -e

# スクリプトのディレクトリを取得
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

cd "$BACKEND_DIR"

# 環境変数の読み込み
if [ ! -f .env ]; then
    echo "Error: .env file not found"
    echo "Please create .env file with required environment variables"
    exit 1
fi

export $(cat .env | grep -v '^#' | xargs)

# テストユーティリティを読み込み
source "$SCRIPT_DIR/test-utils.sh"

# テスト用のユーザー情報
TEST_USER_EMAIL="${TEST_USER_EMAIL:-test-operator-$(date +%s)@example.com}"
TEST_USER_PASSWORD="${TEST_USER_PASSWORD:-TestPassword123!}"
TEST_USER_FIRST_NAME="${TEST_USER_FIRST_NAME:-Test}"
TEST_USER_LAST_NAME="${TEST_USER_LAST_NAME:-Operator}"

# サーバーのポート
API_PORT="${APP_PORT:-8080}"
API_URL="http://localhost:${API_PORT}"

# クリーンアップ関数
cleanup() {
    echo ""
    echo "Cleaning up..."
    
    # サーバープロセスを停止
    if [ ! -z "$SERVER_PID" ]; then
        echo "Stopping server (PID: $SERVER_PID)..."
        kill $SERVER_PID 2>/dev/null || true
        wait $SERVER_PID 2>/dev/null || true
    fi
    
    # テストユーザーを削除
    if [ ! -z "$TEST_USER_ID" ]; then
        cleanup_user "$TEST_USER_ID"
    fi
    
    echo "Cleanup completed"
}

# シグナルハンドラー
trap cleanup EXIT INT TERM

# テスト結果のカウンター
PASSED=0
FAILED=0

# テスト関数
test_passed() {
    echo "✅ PASSED: $1"
    ((PASSED++))
}

test_failed() {
    echo "❌ FAILED: $1"
    ((FAILED++))
}

# サーバーの起動
echo "=========================================="
echo "Starting API server..."
echo "=========================================="

# サーバーをバックグラウンドで起動
go run ./cmd/api > /tmp/api-server.log 2>&1 &
SERVER_PID=$!

# サーバーの起動を待つ
echo "Waiting for server to start..."
for i in {1..30}; do
    if curl -s "${API_URL}/health" > /dev/null 2>&1; then
        echo "Server is ready!"
        break
    fi
    if [ $i -eq 30 ]; then
        echo "Error: Server failed to start within 30 seconds"
        echo "Server logs:"
        cat /tmp/api-server.log
        exit 1
    fi
    sleep 1
done

# テスト1: ヘルスチェック
echo ""
echo "=========================================="
echo "Test 1: Health Check"
echo "=========================================="

response=$(curl -s -w "\n%{http_code}" "${API_URL}/health")
body=$(echo "$response" | head -n -1)
status_code=$(echo "$response" | tail -n 1)

if [ "$status_code" = "200" ] && [ "$body" = "OK" ]; then
    test_passed "Health check endpoint"
else
    test_failed "Health check endpoint (status: $status_code, body: $body)"
fi

# テスト2: ユーザー作成（Supabase Auth）
echo ""
echo "=========================================="
echo "Test 2: Create User (Supabase Auth)"
echo "=========================================="

TEST_USER_ID=$(create_user "$TEST_USER_EMAIL" "$TEST_USER_PASSWORD" "$TEST_USER_FIRST_NAME" "$TEST_USER_LAST_NAME")

if [ $? -eq 0 ] && [ ! -z "$TEST_USER_ID" ]; then
    test_passed "User creation"
    echo "Created user ID: $TEST_USER_ID"
else
    test_failed "User creation"
    exit 1
fi

# テスト3: データベースにオペレーター情報を登録
echo ""
echo "=========================================="
echo "Test 3: Register Operator in Database"
echo "=========================================="

# Supabase MCPツールを使用してオペレーターを登録
echo "Registering operator using Supabase MCP..."

# 注意: この実装はSupabase MCPツールを使用します
# 実際の実装では、SQLを実行してオペレーターを登録する必要があります

register_operator "$TEST_USER_ID" "$TEST_USER_EMAIL" "$TEST_USER_FIRST_NAME" "$TEST_USER_LAST_NAME"

if [ $? -eq 0 ]; then
    test_passed "Operator registration"
else
    test_failed "Operator registration"
fi

# テスト4: ログインしてJWTトークンを取得
echo ""
echo "=========================================="
echo "Test 4: Login and Get JWT Token"
echo "=========================================="

JWT_TOKEN=$(login_user "$TEST_USER_EMAIL" "$TEST_USER_PASSWORD")

if [ $? -eq 0 ] && [ ! -z "$JWT_TOKEN" ]; then
    test_passed "User login and JWT token retrieval"
    echo "JWT Token: ${JWT_TOKEN:0:50}..."
else
    test_failed "User login and JWT token retrieval"
    exit 1
fi

# テスト5: 認証エンドポイントのテスト
echo ""
echo "=========================================="
echo "Test 5: Authentication Endpoint (/api/v1/auth/me)"
echo "=========================================="

response=$(curl -s -w "\n%{http_code}" \
    -H "Authorization: Bearer $JWT_TOKEN" \
    "${API_URL}/api/v1/auth/me")
body=$(echo "$response" | head -n -1)
status_code=$(echo "$response" | tail -n 1)

if [ "$status_code" = "200" ]; then
    echo "Response body: $body"
    
    # レスポンスにuser_idが含まれているか確認
    if echo "$body" | grep -q "\"user_id\""; then
        test_passed "Authentication endpoint"
    else
        test_failed "Authentication endpoint (user_id not found in response)"
    fi
else
    test_failed "Authentication endpoint (status: $status_code, body: $body)"
fi

# テスト結果の表示
echo ""
echo "=========================================="
echo "Test Results"
echo "=========================================="
echo "Passed: $PASSED"
echo "Failed: $FAILED"
echo ""

if [ $FAILED -eq 0 ]; then
    echo "✅ All tests passed!"
    exit 0
else
    echo "❌ Some tests failed"
    exit 1
fi

