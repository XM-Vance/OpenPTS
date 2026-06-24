#!/bin/bash
# PTIS 前端一键重建脚本（宿主机构建 + Dockerfile.local 打镜像的快速迭代路径）
# 用法: bash rebuild-frontend.sh
# 说明: 路径基于脚本位置自动推导，无需修改；完整离线启动见根目录 make up-local。

set -e

FRONTEND_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
COMPOSE_DIR="$(dirname "$FRONTEND_DIR")"

echo ">>> 1/4 安装依赖..."
cd "$FRONTEND_DIR"
npm install --quiet 2>/dev/null

echo ">>> 2/4 构建 Next.js..."
# 确保 .env.local 不存在（避免浏览器直连 Docker 内部地址导致 Network Error）
rm -f .env.local
npm run build

echo ">>> 3/4 打包 Docker 镜像..."
docker build -f Dockerfile.local -t ptis-frontend:latest .

echo ">>> 4/4 重建并启动容器..."
cd "$COMPOSE_DIR"
docker compose create frontend
docker start ptis-frontend

echo ">>> 等待启动..."
sleep 8

STATUS=$(docker inspect --format='{{.State.Health.Status}}' ptis-frontend 2>/dev/null || echo "unknown")
echo ""
echo "========================================="
echo "  前端容器状态: $STATUS"
echo "  访问地址: http://localhost/ （经 nginx 反代；前端容器不再直接对外暴露 3000）"
echo "========================================="
