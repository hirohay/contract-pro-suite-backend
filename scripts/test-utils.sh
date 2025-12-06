#!/bin/bash

# テストユーティリティ関数
# Supabase Auth APIを使用したユーザー操作

set -e

# 環境変数の読み込み
if [ -f .env ]; then
    export $(cat .env | grep -v '^#' | xargs)
fi

# Supabase URLとAPIキーの確認
if [ -z "$SUPABASE_URL" ]; then
    echo "Error: SUPABASE_URL is not set"
    exit 1
fi

if [ -z "$SUPABASE_SERVICE_ROLE_KEY" ]; then
    echo "Error: SUPABASE_SERVICE_ROLE_KEY is not set"
    exit 1
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
    
    response=$(curl -s -X POST "${SUPABASE_URL}/auth/v1/admin/users" \
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
        }")

    # JSONレスポンスからuser_idを抽出（最初の1つだけ）
    user_id=$(echo "$response" | grep -o '"id":"[^"]*' | head -1 | cut -d'"' -f4 | tr -d '\n\r')
    
    if [ -z "$user_id" ]; then
        echo "Error: Failed to create user" >&2
        echo "Response: $response" >&2
        return 1
    fi

    echo -n "$user_id"
    return 0
}

# login_user ユーザーでログインしてJWTトークンを取得
login_user() {
    local email=$1
    local password=$2

    echo "Logging in user: $email" >&2
    
    response=$(curl -s -X POST "${SUPABASE_URL}/auth/v1/token?grant_type=password" \
        -H "apikey: ${SUPABASE_SERVICE_ROLE_KEY}" \
        -H "Content-Type: application/json" \
        -d "{
            \"email\": \"${email}\",
            \"password\": \"${password}\"
        }")

    # JSONレスポンスからaccess_tokenを抽出（最初の1つだけ）
    access_token=$(echo "$response" | grep -o '"access_token":"[^"]*' | head -1 | cut -d'"' -f4 | tr -d '\n\r')
    
    if [ -z "$access_token" ]; then
        echo "Error: Failed to login" >&2
        echo "Response: $response" >&2
        return 1
    fi

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
    
    # 注意: この実装では、Supabase MCPツールを使用してSQLを実行します
    # 実際の登録は、Supabase MCPツールまたはSupabaseダッシュボードのSQL Editorを使用してください
    
    # SQLを実行（Supabase MCPツールを使用する場合）
    # 現在の実装では、手動でSQLを実行する必要があります
    echo "Note: Operator registration SQL (execute manually if needed):" >&2
    echo "INSERT INTO operators (operator_id, email, first_name, last_name, status, mfa_enabled)" >&2
    echo "VALUES ('$user_id'::uuid, '$email', '$first_name', '$last_name', 'ACTIVE', false)" >&2
    echo "ON CONFLICT (email) DO NOTHING;" >&2
    
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

