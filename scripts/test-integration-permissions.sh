#!/bin/bash

# 権限管理機能の統合テストスクリプト
# サーバーを起動して、実際のAPIエンドポイントを使用してテストを実行します

# set -e を削除（エラー時に即座に終了しないようにする）
# set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

cd "$BACKEND_DIR"

# 環境変数の読み込み
if [ ! -f .env ]; then
    echo "Error: .env file not found"
    exit 1
fi

export $(cat .env | grep -v '^#' | xargs)

# テストユーティリティを読み込み
source "$SCRIPT_DIR/test-utils.sh"

echo "=========================================="
echo "権限管理機能の統合テスト"
echo "=========================================="
echo ""

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

# サーバーのポート
API_PORT="${APP_PORT:-8080}"
API_URL="http://localhost:${API_PORT}"

# リトライ関数（指数バックオフ）
retry_with_backoff() {
    local max_attempts="$1"
    local delay="$2"
    local command="$3"
    local attempt=1
    
    while [ $attempt -le $max_attempts ]; do
        if eval "$command"; then
            return 0
        fi
        
        if [ $attempt -lt $max_attempts ]; then
            sleep $delay
            delay=$((delay * 2))
        fi
        attempt=$((attempt + 1))
    done
    
    return 1
}

# サーバーの起動
echo "=========================================="
echo "サーバーの起動"
echo "=========================================="
echo ""

# 既存のサーバーログをクリア
> /tmp/api-server-permissions.log

# サーバーをバックグラウンドで起動
echo "サーバーを起動中..."
go run ./cmd/api > /tmp/api-server-permissions.log 2>&1 &
SERVER_PID=$!

echo "Server PID: $SERVER_PID"

# サーバーの起動を待つ
echo "サーバーの起動を待機中..."
if retry_with_backoff 30 1 "curl -s -f '${API_URL}/health' > /dev/null 2>&1"; then
    echo "サーバーが起動しました！"
    test_passed "サーバーの起動"
else
    echo "Error: サーバーが30秒以内に起動しませんでした"
    test_failed "サーバーの起動"
    echo ""
    echo "サーバーログ（最後の50行）:"
    tail -50 /tmp/api-server-permissions.log
    exit 1
fi

# テスト用のユーザー情報
TEST_USER_EMAIL="test-operator-$(date +%s)@example.com"
TEST_USER_PASSWORD="TestPassword123!"
TEST_USER_FIRST_NAME="Test"
TEST_USER_LAST_NAME="Operator"

# テスト1: ユーザー作成（Supabase Auth）
echo ""
echo "=========================================="
echo "テスト1: ユーザー作成（Supabase Auth）"
echo "=========================================="
echo ""

TEST_USER_ID=$(create_user "$TEST_USER_EMAIL" "$TEST_USER_PASSWORD" "$TEST_USER_FIRST_NAME" "$TEST_USER_LAST_NAME")
create_user_exit_code=$?

if [ $create_user_exit_code -eq 0 ] && [ ! -z "$TEST_USER_ID" ]; then
    test_passed "ユーザー作成"
    echo "作成されたユーザーID: $TEST_USER_ID"
else
    test_failed "ユーザー作成"
    exit 1
fi

# テスト2: データベースにオペレーター情報を登録
echo ""
echo "=========================================="
echo "テスト2: オペレーター登録"
echo "=========================================="
echo ""

# Supabase MCPツールを使用してオペレーターを登録
echo "Supabase MCPツールを使用してオペレーターを登録中..."
OPERATOR_SQL="INSERT INTO operators (operator_id, email, first_name, last_name, status, mfa_enabled)
VALUES ('$TEST_USER_ID'::uuid, '$TEST_USER_EMAIL', '$TEST_USER_FIRST_NAME', '$TEST_USER_LAST_NAME', 'ACTIVE', false)
ON CONFLICT (email) DO UPDATE SET
    operator_id = EXCLUDED.operator_id,
    first_name = EXCLUDED.first_name,
    last_name = EXCLUDED.last_name,
    status = EXCLUDED.status,
    mfa_enabled = EXCLUDED.mfa_enabled,
    updated_at = now();"

# psqlが利用可能な場合は使用
if command -v psql >/dev/null 2>&1 && [ ! -z "$SUPABASE_DB_URL" ]; then
    if echo "$OPERATOR_SQL" | psql "$SUPABASE_DB_URL" -q -t 2>&1; then
        test_passed "オペレーター登録"
    else
        test_failed "オペレーター登録"
        echo "Note: 手動でSQLを実行してください:"
        echo "$OPERATOR_SQL"
    fi
else
    echo "Note: psqlがインストールされていないか、SUPABASE_DB_URLが設定されていません"
    echo "手動でSQLを実行してください:"
    echo "$OPERATOR_SQL"
    echo ""
    echo "または、SupabaseダッシュボードのSQL Editorで実行してください"
    test_failed "オペレーター登録（手動実行が必要）"
fi

# テスト3: ログインしてJWTトークンを取得
echo ""
echo "=========================================="
echo "テスト3: ログインとJWTトークン取得"
echo "=========================================="
echo ""

JWT_TOKEN=$(login_user "$TEST_USER_EMAIL" "$TEST_USER_PASSWORD")
login_user_exit_code=$?

if [ $login_user_exit_code -eq 0 ] && [ ! -z "$JWT_TOKEN" ]; then
    test_passed "ログインとJWTトークン取得"
    echo "JWT Token: ${JWT_TOKEN:0:50}..."
else
    test_failed "ログインとJWTトークン取得"
    exit 1
fi

# テスト4: 認証エンドポイントのテスト
echo ""
echo "=========================================="
echo "テスト4: 認証エンドポイント（/api/v1/auth/me）"
echo "=========================================="
echo ""

CLIENT_ID="${DEFAULT_CLIENT_ID:-00000000-0000-0000-0000-000000000000}"
response=$(curl -s -w "\n%{http_code}" \
    -H "Authorization: Bearer $JWT_TOKEN" \
    -H "X-Client-ID: $CLIENT_ID" \
    "${API_URL}/api/v1/auth/me" 2>&1)
body=$(echo "$response" | sed '$d')
status_code=$(echo "$response" | tail -n 1)

if [ "$status_code" = "200" ]; then
    echo "Response body: $body"
    
    if echo "$body" | grep -q "\"user_id\""; then
        test_passed "認証エンドポイント"
    else
        test_failed "認証エンドポイント（user_idが見つかりません）"
    fi
else
    test_failed "認証エンドポイント（status: $status_code, body: $body）"
fi

# テスト5: クライアント作成とデフォルトロールの確認
echo ""
echo "=========================================="
echo "テスト5: クライアント作成とデフォルトロール"
echo "=========================================="
echo ""

# gRPCを使用してクライアントを作成
if command -v grpcurl >/dev/null 2>&1; then
    echo "gRPCを使用してクライアントを作成中..."
    
    TEST_CLIENT_SLUG="test-client-$(date +%s)"
    TEST_CLIENT_NAME="Test Client $(date +%s)"
    TEST_ADMIN_EMAIL="admin-$(date +%s)@test.com"
    TEST_ADMIN_PASSWORD="TestPassword123!"
    
    # gRPCリクエストを作成
    GRPC_REQUEST=$(cat <<EOF
{
  "name": "$TEST_CLIENT_NAME",
  "slug": "$TEST_CLIENT_SLUG",
  "admin_email": "$TEST_ADMIN_EMAIL",
  "admin_password": "$TEST_ADMIN_PASSWORD",
  "admin_first_name": "Admin",
  "admin_last_name": "User"
}
EOF
)
    
    # gRPCリクエストを実行（ポート8081はgRPCポート）
    GRPC_PORT="${GRPC_PORT:-8081}"
    GRPC_RESPONSE=$(grpcurl -plaintext -d "$GRPC_REQUEST" localhost:${GRPC_PORT} auth.AuthService/SignupClient 2>&1)
    
    if echo "$GRPC_RESPONSE" | grep -q "client_id"; then
        test_passed "クライアント作成（gRPC）"
        echo "Response: $GRPC_RESPONSE"
        
        # クライアントIDを抽出
        CLIENT_ID=$(echo "$GRPC_RESPONSE" | grep -o '"client_id":"[^"]*' | cut -d'"' -f4 || echo "")
        
        if [ ! -z "$CLIENT_ID" ]; then
            echo "作成されたクライアントID: $CLIENT_ID"
            
            # デフォルトロールが作成されているか確認
            echo ""
            echo "デフォルトロールの確認中..."
            
            if command -v psql >/dev/null 2>&1 && [ ! -z "$SUPABASE_DB_URL" ]; then
                ROLES_COUNT=$(psql "$SUPABASE_DB_URL" -t -c "SELECT COUNT(*) FROM client_roles WHERE client_id = '$CLIENT_ID'::uuid AND deleted_at IS NULL;" 2>/dev/null | tr -d ' ')
                
                if [ "$ROLES_COUNT" = "4" ]; then
                    test_passed "デフォルトロールの作成確認（4つのロール）"
                    
                    # 各ロールの存在確認
                    SYSTEM_ADMIN=$(psql "$SUPABASE_DB_URL" -t -c "SELECT COUNT(*) FROM client_roles WHERE client_id = '$CLIENT_ID'::uuid AND code = 'system_admin' AND deleted_at IS NULL;" 2>/dev/null | tr -d ' ')
                    BUSINESS_ADMIN=$(psql "$SUPABASE_DB_URL" -t -c "SELECT COUNT(*) FROM client_roles WHERE client_id = '$CLIENT_ID'::uuid AND code = 'business_admin' AND deleted_at IS NULL;" 2>/dev/null | tr -d ' ')
                    MEMBER=$(psql "$SUPABASE_DB_URL" -t -c "SELECT COUNT(*) FROM client_roles WHERE client_id = '$CLIENT_ID'::uuid AND code = 'member' AND deleted_at IS NULL;" 2>/dev/null | tr -d ' ')
                    READONLY=$(psql "$SUPABASE_DB_URL" -t -c "SELECT COUNT(*) FROM client_roles WHERE client_id = '$CLIENT_ID'::uuid AND code = 'readonly' AND deleted_at IS NULL;" 2>/dev/null | tr -d ' ')
                    
                    if [ "$SYSTEM_ADMIN" = "1" ] && [ "$BUSINESS_ADMIN" = "1" ] && [ "$MEMBER" = "1" ] && [ "$READONLY" = "1" ]; then
                        test_passed "デフォルトロールの詳細確認（system_admin, business_admin, member, readonly）"
                    else
                        test_failed "デフォルトロールの詳細確認"
                    fi
                else
                    test_failed "デフォルトロールの作成確認（期待: 4, 実際: $ROLES_COUNT）"
                fi
            else
                echo "Note: psqlがインストールされていないか、SUPABASE_DB_URLが設定されていません"
                echo "手動で以下のSQLを実行してデフォルトロールを確認してください:"
                echo "SELECT * FROM client_roles WHERE client_id = '$CLIENT_ID'::uuid AND deleted_at IS NULL;"
            fi
        fi
    else
        test_failed "クライアント作成（gRPC）"
        echo "Error: $GRPC_RESPONSE"
    fi
else
    echo "Note: grpcurlがインストールされていません"
    echo "クライアント作成のテストをスキップします"
    echo "期待される動作:"
    echo "  1. クライアント作成時に、systemAdmin, businessAdmin, member, readonlyロールが自動作成される"
    echo "  2. 各ロールに適切な権限が設定される"
fi

# テスト6: 権限チェックのテスト
echo ""
echo "=========================================="
echo "テスト6: 権限チェック機能"
echo "=========================================="
echo ""

echo "Note: 権限チェック用のエンドポイントが実装されたら、ここでテストを追加してください"
echo "期待される動作:"
echo "  1. オペレーターのADMINロールで権限チェックが成功する"
echo "  2. オペレーターのVIEWERロールで読み取り権限のみが許可される"
echo "  3. クライアントユーザーのロールに基づいて権限が正しく判定される"

# テスト結果の表示
echo ""
echo "=========================================="
echo "テスト結果サマリー"
echo "=========================================="
echo "Passed: $PASSED"
echo "Failed: $FAILED"
echo ""

if [ $FAILED -eq 0 ]; then
    echo "✅ すべてのテストが成功しました！"
    exit 0
else
    echo "❌ 一部のテストが失敗しました"
    exit 1
fi

