#!/bin/bash
# 生成 HTTP API 服务端代码（统一管理）
# 用法: ./scripts/generate-api-service.sh <service-name>
# 示例: ./scripts/generate-api-service.sh gateway-service

set -e

SERVICE_NAME=$1
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
SERVICE_DIR="${ROOT_DIR}/${SERVICE_NAME}"
API_DIR="${ROOT_DIR}/proto/api/${SERVICE_NAME}"

if [ -z "$SERVICE_NAME" ]; then
    echo "错误: 请指定服务名称"
    echo "用法: $0 <service-name>"
    echo "示例: $0 gateway-service"
    exit 1
fi

if [ ! -d "$API_DIR" ]; then
    echo "错误: API 目录不存在: $API_DIR"
    exit 1
fi

if [ ! -d "$SERVICE_DIR" ]; then
    echo "错误: 服务目录不存在: $SERVICE_DIR"
    echo "提示: 请先创建服务目录，或使用 goctl api new 创建"
    exit 1
fi

echo "正在生成 HTTP API 服务端代码..."
echo "服务名称: $SERVICE_NAME"
echo "服务目录: $SERVICE_DIR"
echo "API 目录: $API_DIR"

API_FILES=$(find "$API_DIR" -name "*.api" | sort)
if [ -z "$API_FILES" ]; then
    echo "错误: 在 $API_DIR 中未找到 .api 文件"
    exit 1
fi

API_FILE=$(echo "$API_FILES" | head -n 1)
FILE_COUNT=$(echo "$API_FILES" | wc -l | tr -d ' ')
if [ "$FILE_COUNT" -gt 1 ]; then
    echo "⚠️  发现多个 .api 文件，默认使用: $API_FILE"
fi

goctl api go \
  --api="$API_FILE" \
  --dir="$SERVICE_DIR"

echo "✅ HTTP API 服务端代码生成完成！"
