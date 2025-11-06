#!/bin/bash

set -e

# 颜色定义
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

BINARY_NAME="rst"
INSTALL_DIR="/usr/local/bin"
BUILD_DIR="./bin"

# 检查是否以 root 权限运行
if [ "$EUID" -ne 0 ]; then
    echo -e "${RED}Please run as root or use sudo${NC}"
    exit 1
fi

# 检查二进制文件是否存在
if [ ! -f "$BUILD_DIR/$BINARY_NAME" ]; then
    echo -e "${RED}Binary not found. Please build first: make build${NC}"
    exit 1
fi

echo -e "${GREEN}Installing $BINARY_NAME to $INSTALL_DIR...${NC}"

# 备份旧版本（如果存在）
if [ -f "$INSTALL_DIR/$BINARY_NAME" ]; then
    BACKUP_FILE="$INSTALL_DIR/$BINARY_NAME.backup.$(date +%Y%m%d%H%M%S)"
    echo "Backing up existing version to $BACKUP_FILE"
    cp "$INSTALL_DIR/$BINARY_NAME" "$BACKUP_FILE"
fi

# 安装新版本
cp "$BUILD_DIR/$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME"
chmod +x "$INSTALL_DIR/$BINARY_NAME"

# 验证安装
if [ -f "$INSTALL_DIR/$BINARY_NAME" ]; then
    echo -e "${GREEN}✓ Installation successful!${NC}"
    echo -e "You can now run: $BINARY_NAME --help"
else
    echo -e "${RED}✗ Installation failed!${NC}"
    exit 1
fi