#!/bin/bash

# Protocol Buffersコード生成スクリプト

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

cd "$BACKEND_DIR"

# Go binディレクトリをPATHに追加
export PATH="$PATH:$HOME/go/bin"

# protocがインストールされているか確認
if ! command -v protoc &> /dev/null; then
    echo "Error: protoc is not installed" >&2
    echo "Please install protoc:" >&2
    echo "  macOS: brew install protobuf" >&2
    echo "  Linux: apt-get install protobuf-compiler" >&2
    exit 1
fi

# protoファイルのディレクトリ
PROTO_DIR="proto"
OUT_DIR="proto"

# Protocol Buffersコードを生成
echo "Generating Protocol Buffers code..."

protoc \
  --go_out="$OUT_DIR" \
  --go_opt=paths=source_relative \
  --go-grpc_out="$OUT_DIR" \
  --go-grpc_opt=paths=source_relative \
  "$PROTO_DIR/auth/auth.proto"

echo "Protocol Buffers code generated successfully!"
echo "Generated files:"
echo "  - $OUT_DIR/auth/auth.pb.go"
echo "  - $OUT_DIR/auth/auth_grpc.pb.go"

