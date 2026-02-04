#!/bin/bash
# Neptune 项目代码生成工具版本检查与自动安装脚本
# 确保团队成员使用统一的工具版本，避免生成代码合并冲突
# 版本不一致时自动安装正确版本，无需开发者手动操作
set -e

# ============================================
# 版本要求配置（统一使用项目最高版本）
# ============================================
REQUIRED_PROTOC_GEN_GO="v1.36.6"
REQUIRED_PROTOC_GEN_GO_GRPC="v1.5.1"
REQUIRED_GOCTL="v1.9.2"
REQUIRED_PROTOC="29.3"  # protoc 版本不带 v 前缀，格式如 libprotoc 29.3

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo "============================================"
echo "Neptune 代码生成工具版本检查"
echo "============================================"
echo ""

# 安装 protoc-gen-go
install_protoc_gen_go() {
    echo -e "${BLUE}  → 自动安装 protoc-gen-go@${REQUIRED_PROTOC_GEN_GO}...${NC}"
    go install google.golang.org/protobuf/cmd/protoc-gen-go@${REQUIRED_PROTOC_GEN_GO}
    echo -e "${GREEN}  → 安装完成${NC}"
}

# 安装 protoc-gen-go-grpc
install_protoc_gen_go_grpc() {
    echo -e "${BLUE}  → 自动安装 protoc-gen-go-grpc@${REQUIRED_PROTOC_GEN_GO_GRPC}...${NC}"
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@${REQUIRED_PROTOC_GEN_GO_GRPC}
    echo -e "${GREEN}  → 安装完成${NC}"
}

# 安装 goctl
install_goctl() {
    echo -e "${BLUE}  → 自动安装 goctl@${REQUIRED_GOCTL}...${NC}"
    go install github.com/zeromicro/go-zero/tools/goctl@${REQUIRED_GOCTL}
    echo -e "${GREEN}  → 安装完成${NC}"
}

# 检查并自动安装 protoc-gen-go
check_protoc_gen_go() {
    echo -n "检查 protoc-gen-go... "
    
    if ! command -v protoc-gen-go &> /dev/null; then
        echo -e "${YELLOW}未安装${NC}"
        install_protoc_gen_go
        return
    fi
    
    # 获取版本号 (格式: protoc-gen-go v1.36.6)
    CURRENT=$(protoc-gen-go --version 2>&1 | grep -oE 'v[0-9]+\.[0-9]+\.[0-9]+' || echo "unknown")
    
    if [ "$CURRENT" == "$REQUIRED_PROTOC_GEN_GO" ]; then
        echo -e "${GREEN}✓ ${CURRENT}${NC}"
    else
        echo -e "${YELLOW}${CURRENT} → ${REQUIRED_PROTOC_GEN_GO}${NC}"
        install_protoc_gen_go
    fi
}

# 检查并自动安装 protoc-gen-go-grpc
check_protoc_gen_go_grpc() {
    echo -n "检查 protoc-gen-go-grpc... "
    
    if ! command -v protoc-gen-go-grpc &> /dev/null; then
        echo -e "${YELLOW}未安装${NC}"
        install_protoc_gen_go_grpc
        return
    fi
    
    # 获取版本号 (格式: protoc-gen-go-grpc 1.5.1)
    CURRENT=$(protoc-gen-go-grpc --version 2>&1 | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' || echo "unknown")
    
    if [ "v${CURRENT}" == "$REQUIRED_PROTOC_GEN_GO_GRPC" ]; then
        echo -e "${GREEN}✓ v${CURRENT}${NC}"
    else
        echo -e "${YELLOW}v${CURRENT} → ${REQUIRED_PROTOC_GEN_GO_GRPC}${NC}"
        install_protoc_gen_go_grpc
    fi
}

# 检查并自动安装 goctl
check_goctl() {
    echo -n "检查 goctl... "
    
    if ! command -v goctl &> /dev/null; then
        echo -e "${YELLOW}未安装${NC}"
        install_goctl
        return
    fi
    
    # 获取版本号 (格式: goctl version 1.9.2 或 goctl version 1.9.2 darwin/arm64)
    CURRENT=$(goctl --version 2>&1 | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' | head -1 || echo "unknown")
    
    if [ "v${CURRENT}" == "$REQUIRED_GOCTL" ]; then
        echo -e "${GREEN}✓ v${CURRENT}${NC}"
    else
        echo -e "${YELLOW}v${CURRENT} → ${REQUIRED_GOCTL}${NC}"
        install_goctl
    fi
}

# 检查 protoc 版本（仅提示，不自动安装）
check_protoc() {
    echo -n "检查 protoc... "
    
    if ! command -v protoc &> /dev/null; then
        echo -e "${RED}未安装${NC}"
        echo -e "  ${YELLOW}protoc 需手动安装，请参考: https://grpc.io/docs/protoc-installation/${NC}"
        exit 1
    fi
    
    # 获取版本号 (格式: libprotoc 29.3 或 libprotoc 5.29.3)
    CURRENT=$(protoc --version 2>&1 | grep -oE '[0-9]+\.[0-9]+(\.[0-9]+)?' || echo "unknown")
    
    if [ "$CURRENT" == "$REQUIRED_PROTOC" ]; then
        echo -e "${GREEN}✓ v${CURRENT}${NC}"
    else
        echo -e "${YELLOW}⚠ v${CURRENT} (推荐 v${REQUIRED_PROTOC})${NC}"
        # protoc 版本差异通常不会导致代码冲突，只是警告，不阻塞
    fi
}

# 执行所有检查（自动安装）
check_protoc_gen_go
check_protoc_gen_go_grpc
check_goctl
check_protoc

echo ""
echo "============================================"
echo -e "${GREEN}工具版本检查完成 ✓${NC}"
exit 0
