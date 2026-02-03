#!/bin/bash
# 生成 RPC 客户端代码（仅 pb 文件）
# 用法: ./scripts/generate-rpc-client.sh <service-name> <target-dir>
# 示例: ./scripts/generate-rpc-client.sh user gateway-service

set -e

SERVICE_NAME=$1
TARGET_DIR=${2:-"gateway-service"}
PROTO_DIR="../proto/rpc/${SERVICE_NAME}"

if [ -z "$SERVICE_NAME" ]; then
    echo "错误: 请指定服务名称"
    echo "用法: $0 <service-name> [target-dir]"
    echo "示例: $0 user gateway-service"
    exit 1
fi

if [ ! -d "$PROTO_DIR" ]; then
    echo "错误: Proto 目录不存在: $PROTO_DIR"
    exit 1
fi

if [ ! -d "$TARGET_DIR" ]; then
    echo "错误: 目标目录不存在: $TARGET_DIR"
    exit 1
fi

echo "正在生成 RPC 客户端代码..."
echo "服务名称: $SERVICE_NAME"
echo "目标目录: $TARGET_DIR"
echo "Proto 目录: $PROTO_DIR"

cd "$TARGET_DIR"

# 创建 pb 目录（如果不存在）
mkdir -p pb

# 查找所有 proto 文件
PROTO_FILES=$(find "$PROTO_DIR" -name "*.proto" | tr '\n' ' ')

if [ -z "$PROTO_FILES" ]; then
    echo "错误: 在 $PROTO_DIR 中未找到 proto 文件"
    exit 1
fi

# 生成 pb 文件（仅客户端）
goctl rpc protoc \
  --proto_path="$PROTO_DIR" \
  --proto_path="../proto/rpc/common-service" \
  $PROTO_FILES \
  --go_out=./pb \
  --go-grpc_out=./pb

echo "✅ RPC 客户端代码生成完成！"
echo "生成的文件位置: pb/${SERVICE_NAME}/"
