#!/bin/bash

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 项目信息
BINARY_NAME="rst"
VERSION=${1:-"0.1.0"}
BUILD_DIR="./bin"
DIST_DIR="./dist"

# 创建构建目录
mkdir -p $BUILD_DIR
mkdir -p $DIST_DIR

echo -e "${GREEN}Building $BINARY_NAME version $VERSION...${NC}"

# 获取当前时间戳
BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S')
# 获取 Git 提交哈希
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# 构建标志
LDFLAGS="-X 'github.com/budyaya/resty-stress-tester/pkg/version.Version=$VERSION'"
LDFLAGS="$LDFLAGS -X 'github.com/budyaya/resty-stress-tester/pkg/version.BuildTime=$BUILD_TIME'"
LDFLAGS="$LDFLAGS -X 'github.com/budyaya/resty-stress-tester/pkg/version.GitCommit=$GIT_COMMIT'"
LDFLAGS="$LDFLAGS -X 'github.com/budyaya/resty-stress-tester/pkg/version.GoVersion=$(go version)'"

echo -e "${YELLOW}Build info:${NC}"
echo "  Version:    $VERSION"
echo "  Build Time: $BUILD_TIME"
echo "  Git Commit: $GIT_COMMIT"
echo "  Go Version: $(go version)"

# 构建主程序
echo -e "\n${GREEN}Building main binary...${NC}"
go build -ldflags "$LDFLAGS" -o $BUILD_DIR/$BINARY_NAME ./cmd/rst

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Build successful!${NC}"
    echo -e "Binary: $BUILD_DIR/$BINARY_NAME"
    
    # 显示二进制信息
    echo -e "\n${YELLOW}Binary information:${NC}"
    $BUILD_DIR/$BINARY_NAME -version
else
    echo -e "${RED}✗ Build failed!${NC}"
    exit 1
fi

# 构建 Docker 镜像（可选）
if [ "$2" == "docker" ]; then
    echo -e "\n${GREEN}Building Docker image...${NC}"
    docker build -t $BINARY_NAME:$VERSION .
fi

echo -e "\n${GREEN}Build completed!${NC}"