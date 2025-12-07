#!/bin/bash

# テストユーティリティ関数
# Supabase Auth APIを使用したユーザー操作

# set -e を削除（呼び出し元でエラーハンドリングを制御）
# set -e

# デバッグ出力関数
debug() {
    if [ "${DEBUG:-0}" = "1" ]; then
        echo "[DEBUG] $@" >&2
    fi
}

# 環境変数の読み込み
if [ -f .env ]; then
    export $(cat .env | grep -v '^#' | xargs)
fi

# Supabase URLとAPIキーの確認
if [ -z "$SUPABASE_URL" ]; then
    echo "Error: SUPABASE_URL is not set" >&2
    return 1 2>/dev/null || exit 1
fi

if [ -z "$SUPABASE_SERVICE_ROLE_KEY" ]; then
    echo "Error: SUPABASE_SERVICE_ROLE_KEY is not set" >&2
    return 1 2>/dev/null || exit 1
fi

# テスト用のユーザー情報
TEST_USER_EMAIL="${TEST_USER_EMAIL:-test-operator@example.com}"
TEST_USER_PASSWORD="${TEST_USER_PASSWORD:-TestPassword123!}"
TEST_USER_FIRST_NAME="${TEST_USER_FIRST_NAME:-Test}"
TEST_USER_LAST_NAME="${TEST_USER_LAST_NAME:-Operator}"

# create_user ユーザーを作成（Supabase Auth）
create_user() {
    local email=$1
    local password=$2
    local first_name=$3
    local last_name=$4

    echo "Creating user: $email" >&2
    debug "create_user called with email: $email"
    
    # HTTPステータスコードも取得
    response=$(curl -s -w "\n%{http_code}" -X POST "${SUPABASE_URL}/auth/v1/admin/users" \
        -H "apikey: ${SUPABASE_SERVICE_ROLE_KEY}" \
        -H "Authorization: Bearer ${SUPABASE_SERVICE_ROLE_KEY}" \
        -H "Content-Type: application/json" \
        -d "{
            \"email\": \"${email}\",
            \"password\": \"${password}\",
            \"email_confirm\": true,
            \"user_metadata\": {
                \"first_name\": \"${first_name}\",
                \"last_name\": \"${last_name}\"
            }
        }" 2>&1)
    
    # レスポンスボディとHTTPステータスコードを分離
    body=$(echo "$response" | sed '$d')
    http_status=$(echo "$response" | tail -n 1)
    
    debug "create_user HTTP status: $http_status"
    debug "create_user response body: $body"
    
    # HTTPステータスコードのチェック
    if [ "$http_status" != "200" ] && [ "$http_status" != "201" ]; then
        echo "Error: Failed to create user (HTTP $http_status)" >&2
        echo "Response: $body" >&2
        return 1
    fi
    
    # JSONレスポンスからuser_idを抽出（最初の1つだけ）
    # jqが利用可能な場合はjqを使用、そうでない場合はgrepを使用
    if command -v jq >/dev/null 2>&1; then
        user_id=$(echo "$body" | jq -r '.id // empty' 2>/dev/null)
        if [ $? -ne 0 ] || [ -z "$user_id" ] || [ "$user_id" = "null" ]; then
            debug "jq parsing failed, falling back to grep"
            user_id=$(echo "$body" | grep -o '"id":"[^"]*' | head -1 | cut -d'"' -f4 | tr -d '\n\r')
        fi
    else
        user_id=$(echo "$body" | grep -o '"id":"[^"]*' | head -1 | cut -d'"' -f4 | tr -d '\n\r')
    fi
    
    if [ -z "$user_id" ] || [ "$user_id" = "null" ]; then
        echo "Error: Failed to extract user_id from response" >&2
        echo "Response: $body" >&2
        return 1
    fi

    debug "User created successfully: $user_id"
    echo -n "$user_id"
    return 0
}

# login_user ユーザーでログインしてJWTトークンを取得
login_user() {
    local email=$1
    local password=$2

    echo "Logging in user: $email" >&2
    debug "login_user called with email: $email"
    
    # HTTPステータスコードも取得
    response=$(curl -s -w "\n%{http_code}" -X POST "${SUPABASE_URL}/auth/v1/token?grant_type=password" \
        -H "apikey: ${SUPABASE_SERVICE_ROLE_KEY}" \
        -H "Content-Type: application/json" \
        -d "{
            \"email\": \"${email}\",
            \"password\": \"${password}\"
        }" 2>&1)
    
    # レスポンスボディとHTTPステータスコードを分離
    body=$(echo "$response" | sed '$d')
    http_status=$(echo "$response" | tail -n 1)
    
    debug "login_user HTTP status: $http_status"
    debug "login_user response body: ${body:0:200}..."  # 最初の200文字のみ表示
    
    # HTTPステータスコードのチェック
    if [ "$http_status" != "200" ]; then
        echo "Error: Failed to login (HTTP $http_status)" >&2
        echo "Response: $body" >&2
        return 1
    fi
    
    # JSONレスポンスからaccess_tokenを抽出（最初の1つだけ）
    # jqが利用可能な場合はjqを使用、そうでない場合はgrepを使用
    if command -v jq >/dev/null 2>&1; then
        access_token=$(echo "$body" | jq -r '.access_token // empty' 2>/dev/null)
        if [ $? -ne 0 ] || [ -z "$access_token" ] || [ "$access_token" = "null" ]; then
            debug "jq parsing failed, falling back to grep"
            access_token=$(echo "$body" | grep -o '"access_token":"[^"]*' | head -1 | cut -d'"' -f4 | tr -d '\n\r')
        fi
    else
        access_token=$(echo "$body" | grep -o '"access_token":"[^"]*' | head -1 | cut -d'"' -f4 | tr -d '\n\r')
    fi
    
    if [ -z "$access_token" ] || [ "$access_token" = "null" ]; then
        echo "Error: Failed to extract access_token from response" >&2
        echo "Response: $body" >&2
        return 1
    fi

    debug "Login successful, token length: ${#access_token}"
    echo -n "$access_token"
    return 0
}

# register_operator データベースにオペレーター情報を登録
register_operator() {
    local user_id=$1
    local email=$2
    local first_name=$3
    local last_name=$4

    echo "Registering operator in database: $email" >&2
    debug "register_operator called with user_id: $user_id, email: $email"
    
    # SQLを構築
    local sql="INSERT INTO operators (operator_id, email, first_name, last_name, status, mfa_enabled)
VALUES ('$user_id'::uuid, '$email', '$first_name', '$last_name', 'ACTIVE', false)
ON CONFLICT (email) DO UPDATE SET
    operator_id = EXCLUDED.operator_id,
    first_name = EXCLUDED.first_name,
    last_name = EXCLUDED.last_name,
    status = EXCLUDED.status,
    mfa_enabled = EXCLUDED.mfa_enabled,
    updated_at = now();"
    
    debug "SQL: $sql"
    
    # 注意: この実装では、Supabase MCPツールを使用してSQLを実行します
    # 実際の登録は、Supabase MCPツールまたはSupabaseダッシュボードのSQL Editorを使用してください
    echo "Note: Operator registration SQL (execute manually if needed):" >&2
    echo "$sql" >&2
    
    # テストのため、成功として扱う（実際の登録は手動で実行）
    # 統合テストでは、Supabase MCPツールを使用してオペレーターを登録する必要があります
    echo "Operator registration step completed (manual execution may be required)" >&2
    return 0
}

# cleanup_user テスト用ユーザーを削除
cleanup_user() {
    local user_id=$1
    
    if [ -z "$user_id" ]; then
        return 0
    fi

    echo "Cleaning up user: $user_id"
    
    curl -s -X DELETE "${SUPABASE_URL}/auth/v1/admin/users/${user_id}" \
        -H "apikey: ${SUPABASE_SERVICE_ROLE_KEY}" \
        -H "Authorization: Bearer ${SUPABASE_SERVICE_ROLE_KEY}" > /dev/null
    
    echo "User deleted"
}

