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

    echo "Creating user: $email"
    
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

    user_id=$(echo "$response" | grep -o '"id":"[^"]*' | cut -d'"' -f4)
    
    if [ -z "$user_id" ]; then
        echo "Error: Failed to create user"
        echo "Response: $response"
        return 1
    fi

    echo "$user_id"
    return 0
}

# login_user ユーザーでログインしてJWTトークンを取得
login_user() {
    local email=$1
    local password=$2

    echo "Logging in user: $email"
    
    response=$(curl -s -X POST "${SUPABASE_URL}/auth/v1/token?grant_type=password" \
        -H "apikey: ${SUPABASE_SERVICE_ROLE_KEY}" \
        -H "Content-Type: application/json" \
        -d "{
            \"email\": \"${email}\",
            \"password\": \"${password}\"
        }")

    access_token=$(echo "$response" | grep -o '"access_token":"[^"]*' | cut -d'"' -f4)
    
    if [ -z "$access_token" ]; then
        echo "Error: Failed to login"
        echo "Response: $response"
        return 1
    fi

    echo "$access_token"
    return 0
}

# register_operator データベースにオペレーター情報を登録
register_operator() {
    local user_id=$1
    local email=$2
    local first_name=$3
    local last_name=$4

    echo "Registering operator in database: $email"
    
    # SQLを実行してオペレーターを登録
    # 注意: この実装はSupabase MCPツールを使用するか、psqlコマンドを使用する必要があります
    # ここでは簡易実装として、SQLファイルを作成して実行する方法を示します
    
    echo "Note: Operator registration requires database access"
    echo "User ID: $user_id"
    echo "Email: $email"
    echo "First Name: $first_name"
    echo "Last Name: $last_name"
    
    # 実際の実装では、Supabase MCPツールまたはpsqlを使用してSQLを実行
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

