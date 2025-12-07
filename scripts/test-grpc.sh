#!/bin/bash

# gRPCサーバーの起動とテスト実行スクリプト
# ユーザー作成から認証テストまで一連のフローを実行

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

# デバッグ出力関数
debug() {
    if [ "${DEBUG:-0}" = "1" ]; then
        echo "[DEBUG] $@" >&2
    fi
}

# 環境変数検証関数
validate_environment() {
    local required_vars=("SUPABASE_URL" "SUPABASE_SERVICE_ROLE_KEY" "SUPABASE_DB_URL")
    local missing_vars=()
    
    for var in "${required_vars[@]}"; do
        if [ -z "${!var}" ]; then
            missing_vars+=("$var")
        fi
    done
    
    if [ ${#missing_vars[@]} -gt 0 ]; then
        echo "Error: The following required environment variables are not set:" >&2
        for var in "${missing_vars[@]}"; do
            echo "  - $var" >&2
        done
        exit 1
    fi
    
    # GRPC_PORTが設定されていない場合はデフォルト値を使用
    if [ -z "$GRPC_PORT" ]; then
        export GRPC_PORT="8081"
        debug "GRPC_PORT not set, using default: 8081"
    fi
}

# grpcurlのインストール確認
check_grpcurl() {
    if ! command -v grpcurl &> /dev/null; then
        echo "Error: grpcurl is not installed" >&2
        echo "Please install grpcurl:" >&2
        echo "  macOS: brew install grpcurl" >&2
        echo "  Linux: go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest" >&2
        exit 1
    fi
}

# 環境変数の検証
validate_environment

# grpcurlの確認
check_grpcurl

# テスト用の変数
TEST_USER_EMAIL="test-operator-$(date +%s)@example.com"
TEST_USER_PASSWORD="TestPassword123!"
TEST_USER_FIRST_NAME="Test"
TEST_USER_LAST_NAME="Operator"

# gRPCサーバーのURL
GRPC_URL="localhost:${GRPC_PORT:-8081}"

# テストカウンター
PASSED=0
FAILED=0

# テスト関数
test_passed() {
    echo "✅ PASSED: $1"
    ((PASSED++))
    debug "Test passed: $1"
}

test_failed() {
    echo "❌ FAILED: $1"
    ((FAILED++))
    debug "Test failed: $1"
}

# エラー時の詳細情報表示関数
show_error_details() {
    local test_name="$1"
    local error_message="$2"
    
    echo ""
    echo "=========================================="
    echo "Error Details: $test_name"
    echo "=========================================="
    echo "Error: $error_message"
    echo ""
    
    if [ -f /tmp/api-server.log ]; then
        echo "Server logs (last 50 lines):"
        echo "----------------------------------------"
        tail -50 /tmp/api-server.log
        echo ""
    fi
    
    if [ "${DEBUG:-0}" = "1" ]; then
        echo "Environment variables:"
        echo "  SUPABASE_URL: ${SUPABASE_URL:0:50}..."
        echo "  GRPC_URL: $GRPC_URL"
        echo "  TEST_USER_EMAIL: $TEST_USER_EMAIL"
        echo ""
    fi
}

# リトライ関数（指数バックオフ）
retry_with_backoff() {
    local max_attempts="$1"
    local delay="$2"
    local command="$3"
    local attempt=1
    
    while [ $attempt -le $max_attempts ]; do
        debug "Attempt $attempt/$max_attempts: $command"
        if eval "$command"; then
            return 0
        fi
        
        if [ $attempt -lt $max_attempts ]; then
            debug "Waiting ${delay}s before retry..."
            sleep $delay
            delay=$((delay * 2))  # 指数バックオフ
        fi
        attempt=$((attempt + 1))
    done
    
    return 1
}

# クリーンアップ関数
cleanup() {
    echo ""
    echo "=========================================="
    echo "Cleaning up..."
    echo "=========================================="
    
    # サーバーを停止
    if [ ! -z "$SERVER_PID" ]; then
        debug "Stopping server (PID: $SERVER_PID)"
        kill $SERVER_PID 2>/dev/null || true
        wait $SERVER_PID 2>/dev/null || true
    fi
    
    # テストユーザーを削除
    if [ ! -z "$TEST_USER_ID" ]; then
        debug "Deleting test user: $TEST_USER_ID"
        cleanup_user "$TEST_USER_ID" || true
    fi
    
    echo "Cleanup completed"
}

# シグナルハンドラー
trap cleanup EXIT INT TERM

# サーバーの起動
echo "=========================================="
echo "Starting gRPC server..."
echo "=========================================="

# 既存のサーバーログをクリア
> /tmp/api-server.log

# サーバーをバックグラウンドで起動
debug "Starting server: go run ./cmd/api"
go run ./cmd/api > /tmp/api-server.log 2>&1 &
SERVER_PID=$!

debug "Server PID: $SERVER_PID"

# サーバーの起動を待つ（gRPCサーバーのリスト確認で確認）
echo "Waiting for server to start..."
if retry_with_backoff 30 1 "grpcurl -plaintext $GRPC_URL list > /dev/null 2>&1"; then
    echo "Server is ready!"
    debug "Server gRPC check passed"
else
    echo "Error: Server failed to start within 30 seconds"
    show_error_details "Server startup" "Server failed to respond to gRPC list command"
    exit 1
fi

# テスト1: gRPCサービスのリスト確認
echo ""
echo "=========================================="
echo "Test 1: gRPC Service List"
echo "=========================================="

debug "Listing gRPC services: grpcurl -plaintext $GRPC_URL list"
services=$(grpcurl -plaintext $GRPC_URL list 2>&1)

if echo "$services" | grep -q "auth.AuthService"; then
    test_passed "gRPC service list"
    echo "Available services:"
    echo "$services"
else
    test_failed "gRPC service list"
    show_error_details "gRPC service list" "AuthService not found in service list"
    exit 1
fi

# テスト2: ユーザー作成（Supabase Auth）
echo ""
echo "=========================================="
echo "Test 2: Create User (Supabase Auth)"
echo "=========================================="

debug "Creating user: $TEST_USER_EMAIL"
TEST_USER_ID=$(create_user "$TEST_USER_EMAIL" "$TEST_USER_PASSWORD" "$TEST_USER_FIRST_NAME" "$TEST_USER_LAST_NAME")
create_user_exit_code=$?

if [ $create_user_exit_code -eq 0 ] && [ ! -z "$TEST_USER_ID" ]; then
    test_passed "User creation"
    echo "Created user ID: $TEST_USER_ID"
    debug "User created successfully: $TEST_USER_ID"
else
    test_failed "User creation"
    show_error_details "User creation" "Failed to create user (exit code: $create_user_exit_code, user_id: ${TEST_USER_ID:-empty})"
    exit 1
fi

# テスト3: データベースにオペレーター情報を登録
echo ""
echo "=========================================="
echo "Test 3: Register Operator in Database"
echo "=========================================="

debug "Registering operator: $TEST_USER_EMAIL"
register_operator "$TEST_USER_ID" "$TEST_USER_EMAIL" "$TEST_USER_FIRST_NAME" "$TEST_USER_LAST_NAME"
register_operator_exit_code=$?

if [ $register_operator_exit_code -eq 0 ]; then
    test_passed "Operator registration"
    debug "Operator registration completed"
else
    test_failed "Operator registration"
    show_error_details "Operator registration" "Failed to register operator (exit code: $register_operator_exit_code)"
    echo "Note: You may need to manually register the operator in the database"
fi

# テスト4: ログインしてJWTトークンを取得
echo ""
echo "=========================================="
echo "Test 4: Login and Get JWT Token"
echo "=========================================="

debug "Logging in user: $TEST_USER_EMAIL"
JWT_TOKEN=$(login_user "$TEST_USER_EMAIL" "$TEST_USER_PASSWORD")
login_user_exit_code=$?

if [ $login_user_exit_code -eq 0 ] && [ ! -z "$JWT_TOKEN" ]; then
    test_passed "User login and JWT token retrieval"
    echo "JWT Token: ${JWT_TOKEN:0:50}..."
    debug "JWT token retrieved successfully (length: ${#JWT_TOKEN})"
else
    test_failed "User login and JWT token retrieval"
    show_error_details "User login" "Failed to login user (exit code: $login_user_exit_code, token: ${JWT_TOKEN:-empty})"
    exit 1
fi

# テスト5: gRPC GetMeエンドポイントのテスト
echo ""
echo "=========================================="
echo "Test 5: gRPC GetMe Endpoint"
echo "=========================================="

# クライアントIDを取得（デフォルトまたは環境変数から）
CLIENT_ID="${DEFAULT_CLIENT_ID:-00000000-0000-0000-0000-000000000000}"

debug "Calling gRPC GetMe: grpcurl -plaintext -H 'authorization: Bearer $JWT_TOKEN' -H 'x-client-id: $CLIENT_ID' $GRPC_URL auth.AuthService/GetMe"

# gRPCリクエストを実行
response=$(grpcurl -plaintext \
    -H "authorization: Bearer $JWT_TOKEN" \
    -H "x-client-id: $CLIENT_ID" \
    -d '{}' \
    $GRPC_URL \
    auth.AuthService/GetMe 2>&1)

grpc_exit_code=$?

if [ $grpc_exit_code -eq 0 ]; then
    test_passed "gRPC GetMe endpoint"
    echo "Response:"
    echo "$response" | jq . 2>/dev/null || echo "$response"
    
    # レスポンスの検証
    if echo "$response" | grep -q "user_id"; then
        test_passed "GetMe response validation"
    else
        test_failed "GetMe response validation"
        show_error_details "GetMe response" "Response does not contain user_id"
    fi
else
    test_failed "gRPC GetMe endpoint"
    show_error_details "gRPC GetMe" "Failed to call GetMe (exit code: $grpc_exit_code, response: $response)"
fi

    # テスト6: SignupClientエンドポイントのテスト
    echo ""
    echo "=========================================="
    echo "Test 6: gRPC SignupClient Endpoint"
    echo "=========================================="

    # テスト用のクライアント情報を生成
    TEST_CLIENT_NAME="Test Client $(date +%s)"
    TEST_CLIENT_SLUG="test-client-$(date +%s)"
    TEST_CLIENT_COMPANY_CODE="TEST$(date +%s)"
    TEST_ADMIN_EMAIL="admin-$(date +%s)@test.com"
    TEST_ADMIN_PASSWORD="TestPassword123!"
    TEST_ADMIN_FIRST_NAME="Admin"
    TEST_ADMIN_LAST_NAME="User"

    debug "Calling gRPC SignupClient: grpcurl -plaintext -d '{...}' $GRPC_URL auth.AuthService/SignupClient"

    # gRPCリクエストを実行
    signup_request=$(cat <<EOF
{
  "name": "$TEST_CLIENT_NAME",
  "company_code": "$TEST_CLIENT_COMPANY_CODE",
  "slug": "$TEST_CLIENT_SLUG",
  "admin_email": "$TEST_ADMIN_EMAIL",
  "admin_password": "$TEST_ADMIN_PASSWORD",
  "admin_first_name": "$TEST_ADMIN_FIRST_NAME",
  "admin_last_name": "$TEST_ADMIN_LAST_NAME"
}
EOF
)

    response=$(grpcurl -plaintext \
        -d "$signup_request" \
        $GRPC_URL \
        auth.AuthService/SignupClient 2>&1)

    grpc_exit_code=$?

    if [ $grpc_exit_code -eq 0 ]; then
        test_passed "gRPC SignupClient endpoint"
        echo "Response:"
        echo "$response" | jq . 2>/dev/null || echo "$response"
        
        # レスポンスの検証
        if echo "$response" | grep -q "client_id"; then
            test_passed "SignupClient response validation"
            
            # client_idとadmin_user_idを抽出
            CLIENT_ID=$(echo "$response" | jq -r '.client_id // empty' 2>/dev/null || echo "$response" | grep -o '"client_id":"[^"]*' | head -1 | cut -d'"' -f4)
            ADMIN_USER_ID=$(echo "$response" | jq -r '.admin_user_id // empty' 2>/dev/null || echo "$response" | grep -o '"admin_user_id":"[^"]*' | head -1 | cut -d'"' -f4)
            
            if [ ! -z "$CLIENT_ID" ] && [ ! -z "$ADMIN_USER_ID" ]; then
                debug "Client ID: $CLIENT_ID"
                debug "Admin User ID: $ADMIN_USER_ID"
            fi
        else
            test_failed "SignupClient response validation"
            show_error_details "SignupClient response" "Response does not contain client_id"
        fi
    else
        test_failed "gRPC SignupClient endpoint"
        show_error_details "gRPC SignupClient" "Failed to call SignupClient (exit code: $grpc_exit_code, response: $response)"
    fi

    # テスト結果のサマリー
    echo ""
    echo "=========================================="
    echo "Test Summary"
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

