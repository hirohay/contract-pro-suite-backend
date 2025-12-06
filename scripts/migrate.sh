#!/bin/bash

# Supabaseマイグレーション実行スクリプト
# 使用方法: ./scripts/migrate.sh [migration-file]

set -e

MIGRATION_FILE=${1:-"migrations/001_initial_auth_tables.sql"}

if [ ! -f "$MIGRATION_FILE" ]; then
    echo "Error: Migration file not found: $MIGRATION_FILE"
    exit 1
fi

echo "Migration file: $MIGRATION_FILE"
echo ""
echo "このスクリプトはSupabase CLIを使用してマイグレーションを実行します。"
echo "Supabase CLIがインストールされていない場合は、SupabaseダッシュボードのSQL Editorを使用してください。"
echo ""
echo "Supabase CLIを使用する場合:"
echo "  supabase db push"
echo ""
echo "または、SupabaseダッシュボードのSQL Editorで以下のファイルの内容を実行してください:"
echo "  $MIGRATION_FILE"
echo ""

# Supabase CLIがインストールされているか確認
if command -v supabase &> /dev/null; then
    echo "Supabase CLIが見つかりました。"
    echo "マイグレーションを実行しますか？ (y/n)"
    read -r response
    if [[ "$response" =~ ^[Yy]$ ]]; then
        supabase db push
    else
        echo "マイグレーションをスキップしました。"
    fi
else
    echo "Supabase CLIが見つかりませんでした。"
    echo "SupabaseダッシュボードのSQL Editorを使用してください。"
fi

