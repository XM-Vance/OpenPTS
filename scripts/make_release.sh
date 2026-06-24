#!/usr/bin/env bash
# OpenPTS 安全打包脚本 — 排除所有敏感文件
# 用法: bash scripts/make_release.sh [输出文件名]
set -euo pipefail

OUT="${1:-ptis-source.tar.gz}"

if [ "$OUT" != "ptis-source.tar.gz" ]; then
    # 同步离线包配置：从根配置生成 offline-package 配置副本
    # 这样维护者只需维护根配置，离线包配置自动同步
    echo ""
    echo "同步离线包配置..."
    mkdir -p offline-package/config/nginx
    cp docker-compose.yml offline-package/config/docker-compose.yml
    cp nginx/nginx.conf offline-package/config/nginx/nginx.conf
    cp .env.example offline-package/config/.env.example
    echo "  offline-package/config/ 已从根配置同步"
fi

tar --exclude='./.env' \
    --exclude='*/.env' \
    --exclude='*.env.backup*' \
    --exclude='*/ptis-server' \
    --exclude='*/node_modules' \
    --exclude='__pycache__' \
    --exclude='*.py[cod]' \
    --exclude='*.py.backup' \
    --exclude='*.tsbuildinfo' \
    --exclude='./frontend/.next' \
    --exclude='.git' \
    --exclude='.DS_Store' \
    --exclude='backups' \
    --exclude='./offline-package/images' \
    --exclude='./offline-package/ptis_db.sql' \
    -czf "$OUT" .

echo ""
echo "校验敏感文件..."
if tar -tzf "$OUT" | grep -E '\.env$|\.backup|__pycache__'; then
    echo "⚠️  仍有敏感文件，请检查！"
    exit 1
else
    SIZE=$(ls -lh "$OUT" | awk '{print $5}')
    echo "✅ 干净！包大小: $SIZE → $OUT"
fi
