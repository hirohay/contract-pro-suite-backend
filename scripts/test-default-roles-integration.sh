#!/bin/bash

# デフォルトロール作成機能の統合テストスクリプト
# SignupClientでクライアント作成後、デフォルトロールと権限が正しく作成されることを確認します

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
echo "デフォルトロール作成機能の統合テスト"
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
    
    echo "Cleanup completed"
}

# シグナルハンドラー
trap cleanup EXIT INT TERM

# サーバーのポート
GRPC_PORT="${GRPC_PORT:-8081}"

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
> /tmp/api-server-default-roles.log

# サーバーをバックグラウンドで起動
echo "サーバーを起動中..."
go run ./cmd/api > /tmp/api-server-default-roles.log 2>&1 &
SERVER_PID=$!

echo "Server PID: $SERVER_PID"

# サーバーの起動を待つ
echo "サーバーの起動を待機中..."
if retry_with_backoff 30 1 "grpcurl -plaintext localhost:${GRPC_PORT} list > /dev/null 2>&1"; then
    echo "サーバーが起動しました！"
    test_passed "サーバーの起動"
else
    echo "Error: サーバーが30秒以内に起動しませんでした"
    test_failed "サーバーの起動"
    echo ""
    echo "サーバーログ（最後の50行）:"
    tail -50 /tmp/api-server-default-roles.log
    exit 1
fi

# テスト1: SignupClientでクライアント作成
echo ""
echo "=========================================="
echo "テスト1: SignupClientでクライアント作成"
echo "=========================================="
echo ""

if ! command -v grpcurl >/dev/null 2>&1; then
    echo "Error: grpcurlがインストールされていません"
    echo "インストール方法:"
    echo "  macOS: brew install grpcurl"
    echo "  Linux: go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest"
    test_failed "grpcurlのインストール確認"
    exit 1
fi

TEST_CLIENT_SLUG="test-client-$(date +%s)"
TEST_CLIENT_NAME="Test Client $(date +%s)"
TEST_ADMIN_EMAIL="admin-$(date +%s)@test.com"
TEST_ADMIN_PASSWORD="TestPassword123!"
TEST_ADMIN_FIRST_NAME="Admin"
TEST_ADMIN_LAST_NAME="User"

GRPC_REQUEST=$(cat <<EOF
{
  "name": "$TEST_CLIENT_NAME",
  "slug": "$TEST_CLIENT_SLUG",
  "admin_email": "$TEST_ADMIN_EMAIL",
  "admin_password": "$TEST_ADMIN_PASSWORD",
  "admin_first_name": "$TEST_ADMIN_FIRST_NAME",
  "admin_last_name": "$TEST_ADMIN_LAST_NAME"
}
EOF
)

GRPC_RESPONSE=$(grpcurl -plaintext -d "$GRPC_REQUEST" localhost:${GRPC_PORT} auth.AuthService/SignupClient 2>&1)
grpc_exit_code=$?

if [ $grpc_exit_code -eq 0 ] && (echo "$GRPC_RESPONSE" | grep -q "clientId" || echo "$GRPC_RESPONSE" | grep -q "client_id"); then
    test_passed "SignupClientでクライアント作成"
    echo "Response: $GRPC_RESPONSE"
    
    # クライアントIDを抽出（キャメルケースとスネークケースの両方に対応）
    CLIENT_ID=$(echo "$GRPC_RESPONSE" | grep -o '"clientId":"[^"]*' | cut -d'"' -f4 || echo "")
    
    if [ -z "$CLIENT_ID" ]; then
        CLIENT_ID=$(echo "$GRPC_RESPONSE" | grep -o '"client_id":"[^"]*' | cut -d'"' -f4 || echo "")
    fi
    
    if [ -z "$CLIENT_ID" ]; then
        # jqを使用して抽出を試みる
        if command -v jq >/dev/null 2>&1; then
            CLIENT_ID=$(echo "$GRPC_RESPONSE" | jq -r '.clientId // .client_id // empty' 2>/dev/null)
        fi
    fi
    
    if [ ! -z "$CLIENT_ID" ]; then
        echo "作成されたクライアントID: $CLIENT_ID"
    else
        test_failed "クライアントIDの抽出"
        echo "Response: $GRPC_RESPONSE"
        exit 1
    fi
else
    test_failed "SignupClientでクライアント作成"
    echo "Error: $GRPC_RESPONSE"
    exit 1
fi

# テスト2: デフォルトロールの存在確認
echo ""
echo "=========================================="
echo "テスト2: デフォルトロールの存在確認"
echo "=========================================="
echo ""

if [ -z "$CLIENT_ID" ]; then
    test_failed "クライアントIDが取得できませんでした"
    exit 1
fi

# Supabase MCPまたはpsqlを使用してデータベースにクエリ
if command -v psql >/dev/null 2>&1 && [ ! -z "$SUPABASE_DB_URL" ]; then
    echo "データベースからデフォルトロールを確認中..."
    
    # ロール数の確認
    ROLES_COUNT=$(psql "$SUPABASE_DB_URL" -t -c "SELECT COUNT(*) FROM client_roles WHERE client_id = '$CLIENT_ID'::uuid AND deleted_at IS NULL;" 2>/dev/null | tr -d ' ')
    
    if [ "$ROLES_COUNT" = "4" ]; then
        test_passed "デフォルトロールの数（4つ）"
    else
        test_failed "デフォルトロールの数（期待: 4, 実際: $ROLES_COUNT）"
    fi
    
    # 各ロールの存在確認
    SYSTEM_ADMIN=$(psql "$SUPABASE_DB_URL" -t -c "SELECT COUNT(*) FROM client_roles WHERE client_id = '$CLIENT_ID'::uuid AND code = 'system_admin' AND deleted_at IS NULL;" 2>/dev/null | tr -d ' ')
    BUSINESS_ADMIN=$(psql "$SUPABASE_DB_URL" -t -c "SELECT COUNT(*) FROM client_roles WHERE client_id = '$CLIENT_ID'::uuid AND code = 'business_admin' AND deleted_at IS NULL;" 2>/dev/null | tr -d ' ')
    MEMBER=$(psql "$SUPABASE_DB_URL" -t -c "SELECT COUNT(*) FROM client_roles WHERE client_id = '$CLIENT_ID'::uuid AND code = 'member' AND deleted_at IS NULL;" 2>/dev/null | tr -d ' ')
    READONLY=$(psql "$SUPABASE_DB_URL" -t -c "SELECT COUNT(*) FROM client_roles WHERE client_id = '$CLIENT_ID'::uuid AND code = 'readonly' AND deleted_at IS NULL;" 2>/dev/null | tr -d ' ')
    
    if [ "$SYSTEM_ADMIN" = "1" ] && [ "$BUSINESS_ADMIN" = "1" ] && [ "$MEMBER" = "1" ] && [ "$READONLY" = "1" ]; then
        test_passed "デフォルトロールの存在確認（system_admin, business_admin, member, readonly）"
    else
        test_failed "デフォルトロールの存在確認（system_admin: $SYSTEM_ADMIN, business_admin: $BUSINESS_ADMIN, member: $MEMBER, readonly: $READONLY）"
    fi
    
    # is_systemフラグの確認
    IS_SYSTEM_COUNT=$(psql "$SUPABASE_DB_URL" -t -c "SELECT COUNT(*) FROM client_roles WHERE client_id = '$CLIENT_ID'::uuid AND is_system = true AND deleted_at IS NULL;" 2>/dev/null | tr -d ' ')
    
    if [ "$IS_SYSTEM_COUNT" = "4" ]; then
        test_passed "デフォルトロールのis_systemフラグ（すべてtrue）"
    else
        test_failed "デフォルトロールのis_systemフラグ（期待: 4, 実際: $IS_SYSTEM_COUNT）"
    fi
else
    echo "Note: psqlがインストールされていないか、SUPABASE_DB_URLが設定されていません"
    echo "手動で以下のSQLを実行してデフォルトロールを確認してください:"
    echo "SELECT * FROM client_roles WHERE client_id = '$CLIENT_ID'::uuid AND deleted_at IS NULL;"
fi

# テスト3: 各ロールの権限確認
echo ""
echo "=========================================="
echo "テスト3: 各ロールの権限確認"
echo "=========================================="
echo ""

if command -v psql >/dev/null 2>&1 && [ ! -z "$SUPABASE_DB_URL" ]; then
    # system_adminロールの権限確認（全機能の全アクション = 36個）
    SYSTEM_ADMIN_ROLE_ID=$(psql "$SUPABASE_DB_URL" -t -c "SELECT role_id FROM client_roles WHERE client_id = '$CLIENT_ID'::uuid AND code = 'system_admin' AND deleted_at IS NULL LIMIT 1;" 2>/dev/null | tr -d ' ')
    
    if [ ! -z "$SYSTEM_ADMIN_ROLE_ID" ]; then
        SYSTEM_ADMIN_PERM_COUNT=$(psql "$SUPABASE_DB_URL" -t -c "SELECT COUNT(*) FROM client_role_permissions WHERE role_id = '$SYSTEM_ADMIN_ROLE_ID'::uuid AND deleted_at IS NULL AND granted = true;" 2>/dev/null | tr -d ' ')
        
        if [ "$SYSTEM_ADMIN_PERM_COUNT" = "36" ]; then
            test_passed "system_adminロールの権限数（36個）"
        else
            test_failed "system_adminロールの権限数（期待: 36, 実際: $SYSTEM_ADMIN_PERM_COUNT）"
        fi
    fi
    
    # business_adminロールの権限確認（system_settingsはREADのみ + その他8機能×4アクション = 33個）
    BUSINESS_ADMIN_ROLE_ID=$(psql "$SUPABASE_DB_URL" -t -c "SELECT role_id FROM client_roles WHERE client_id = '$CLIENT_ID'::uuid AND code = 'business_admin' AND deleted_at IS NULL LIMIT 1;" 2>/dev/null | tr -d ' ')
    
    if [ ! -z "$BUSINESS_ADMIN_ROLE_ID" ]; then
        BUSINESS_ADMIN_PERM_COUNT=$(psql "$SUPABASE_DB_URL" -t -c "SELECT COUNT(*) FROM client_role_permissions WHERE role_id = '$BUSINESS_ADMIN_ROLE_ID'::uuid AND deleted_at IS NULL AND granted = true;" 2>/dev/null | tr -d ' ')
        
        if [ "$BUSINESS_ADMIN_PERM_COUNT" = "33" ]; then
            test_passed "business_adminロールの権限数（33個）"
        else
            test_failed "business_adminロールの権限数（期待: 33, 実際: $BUSINESS_ADMIN_PERM_COUNT）"
        fi
    fi
    
    # memberロールの権限確認（system_settingsとapprovalsはREADのみ + その他7機能×3アクション = 23個）
    MEMBER_ROLE_ID=$(psql "$SUPABASE_DB_URL" -t -c "SELECT role_id FROM client_roles WHERE client_id = '$CLIENT_ID'::uuid AND code = 'member' AND deleted_at IS NULL LIMIT 1;" 2>/dev/null | tr -d ' ')
    
    if [ ! -z "$MEMBER_ROLE_ID" ]; then
        MEMBER_PERM_COUNT=$(psql "$SUPABASE_DB_URL" -t -c "SELECT COUNT(*) FROM client_role_permissions WHERE role_id = '$MEMBER_ROLE_ID'::uuid AND deleted_at IS NULL AND granted = true;" 2>/dev/null | tr -d ' ')
        
        if [ "$MEMBER_PERM_COUNT" = "23" ]; then
            test_passed "memberロールの権限数（23個）"
        else
            test_failed "memberロールの権限数（期待: 23, 実際: $MEMBER_PERM_COUNT）"
        fi
    fi
    
    # readonlyロールの権限確認（全機能READのみ = 9個）
    READONLY_ROLE_ID=$(psql "$SUPABASE_DB_URL" -t -c "SELECT role_id FROM client_roles WHERE client_id = '$CLIENT_ID'::uuid AND code = 'readonly' AND deleted_at IS NULL LIMIT 1;" 2>/dev/null | tr -d ' ')
    
    if [ ! -z "$READONLY_ROLE_ID" ]; then
        READONLY_PERM_COUNT=$(psql "$SUPABASE_DB_URL" -t -c "SELECT COUNT(*) FROM client_role_permissions WHERE role_id = '$READONLY_ROLE_ID'::uuid AND deleted_at IS NULL AND granted = true;" 2>/dev/null | tr -d ' ')
        
        if [ "$READONLY_PERM_COUNT" = "9" ]; then
            test_passed "readonlyロールの権限数（9個）"
        else
            test_failed "readonlyロールの権限数（期待: 9, 実際: $READONLY_PERM_COUNT）"
        fi
    fi
else
    echo "Note: psqlがインストールされていないか、SUPABASE_DB_URLが設定されていません"
    echo "手動で以下のSQLを実行して権限を確認してください:"
    echo "SELECT cr.code, COUNT(crp.*) FROM client_roles cr"
    echo "JOIN client_role_permissions crp ON cr.role_id = crp.role_id"
    echo "WHERE cr.client_id = '$CLIENT_ID'::uuid AND cr.deleted_at IS NULL AND crp.deleted_at IS NULL AND crp.granted = true"
    echo "GROUP BY cr.code;"
fi

# テスト4: 管理者ユーザーのロール割り当て確認
echo ""
echo "=========================================="
echo "テスト4: 管理者ユーザーのロール割り当て確認"
echo "=========================================="
echo ""

if command -v psql >/dev/null 2>&1 && [ ! -z "$SUPABASE_DB_URL" ]; then
    # 管理者ユーザーIDを取得（キャメルケースとスネークケースの両方に対応）
    ADMIN_USER_ID=$(echo "$GRPC_RESPONSE" | grep -o '"adminUserId":"[^"]*' | cut -d'"' -f4 || echo "")
    
    if [ -z "$ADMIN_USER_ID" ]; then
        ADMIN_USER_ID=$(echo "$GRPC_RESPONSE" | grep -o '"admin_user_id":"[^"]*' | cut -d'"' -f4 || echo "")
    fi
    
    if [ -z "$ADMIN_USER_ID" ] && command -v jq >/dev/null 2>&1; then
        ADMIN_USER_ID=$(echo "$GRPC_RESPONSE" | jq -r '.adminUserId // .admin_user_id // empty' 2>/dev/null)
    fi
    
    if [ ! -z "$ADMIN_USER_ID" ]; then
        # 管理者ユーザーにsystem_adminロールが割り当てられているか確認
        ADMIN_ROLE_ASSIGNMENT=$(psql "$SUPABASE_DB_URL" -t -c "SELECT COUNT(*) FROM client_user_roles cur JOIN client_roles cr ON cur.role_id = cr.role_id WHERE cur.client_user_id = '$ADMIN_USER_ID'::uuid AND cr.code = 'system_admin' AND cur.client_id = '$CLIENT_ID'::uuid;" 2>/dev/null | tr -d ' ')
        
        if [ "$ADMIN_ROLE_ASSIGNMENT" = "1" ]; then
            test_passed "管理者ユーザーへのsystem_adminロール割り当て"
        else
            test_failed "管理者ユーザーへのsystem_adminロール割り当て（期待: 1, 実際: $ADMIN_ROLE_ASSIGNMENT）"
        fi
    else
        test_failed "管理者ユーザーIDの取得"
    fi
else
    echo "Note: psqlがインストールされていないか、SUPABASE_DB_URLが設定されていません"
    echo "手動で以下のSQLを実行してロール割り当てを確認してください:"
    echo "SELECT * FROM client_user_roles WHERE client_id = '$CLIENT_ID'::uuid;"
fi

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

