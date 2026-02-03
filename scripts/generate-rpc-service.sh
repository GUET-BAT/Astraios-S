#!/bin/bash
# 生成 RPC 服务端代码和客户端代码（统一管理）
# 用法: ./scripts/generate-rpc-service.sh <service-name> [service-dir]
# 示例: ./scripts/generate-rpc-service.sh user user-service

set -e

SERVICE_NAME=$1
SERVICE_DIR=${2:-"${SERVICE_NAME}-service"}
PROTO_DIR="../proto/rpc/${SERVICE_NAME}"

if [ -z "$SERVICE_NAME" ]; then
    echo "错误: 请指定服务名称"
    echo "用法: $0 <service-name> [service-dir]"
    echo "示例: $0 user user-service"
    exit 1
fi

if [ ! -d "$PROTO_DIR" ]; then
    echo "错误: Proto 目录不存在: $PROTO_DIR"
    exit 1
fi

if [ ! -d "$SERVICE_DIR" ]; then
    echo "错误: 服务目录不存在: $SERVICE_DIR"
    echo "提示: 请先创建服务目录，或使用 goctl rpc new 创建"
    exit 1
fi

echo "正在生成 RPC 服务端代码和客户端代码..."
echo "服务名称: $SERVICE_NAME"
echo "服务目录: $SERVICE_DIR"
echo "Proto 目录: $PROTO_DIR"

cd "$SERVICE_DIR"

# 查找所有 proto 文件
PROTO_FILES=$(find "$PROTO_DIR" -name "*.proto" | tr '\n' ' ')

if [ -z "$PROTO_FILES" ]; then
    echo "错误: 在 $PROTO_DIR 中未找到 proto 文件"
    exit 1
fi

# 创建 pb 目录（如果不存在）
mkdir -p pb

# 生成服务端代码和 pb 文件
goctl rpc protoc \
  --proto_path="$PROTO_DIR" \
  --proto_path="../proto/rpc/common-service" \
  $PROTO_FILES \
  --go_out=./pb \
  --go-grpc_out=./pb \
  --zrpc_out=.

echo "✅ RPC 服务端代码生成完成！"
echo "生成的文件位置:"
echo "  - pb/${SERVICE_NAME}/"
echo "  - internal/logic/"
echo "  - internal/server/"
echo ""
echo "📝 下一步：在 client/ 目录下创建客户端封装代码"
echo "   参考示例: gateway-service/client/${SERVICE_NAME}.go.example"
