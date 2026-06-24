#!/usr/bin/env bash
# =============================================================================
# OpenPTS V1 — PostgreSQL 备份脚本
#
# 功能：
#   - pg_dump 全量导出，gzip 压缩
#   - 文件名带时间戳：ptis_backup_YYYYMMDD_HHMMSS.sql.gz
#   - 保留最近 7 天备份，自动清理旧的
#   - 备份目录可通过环境变量 BACKUP_DIR 配置
#
# 用法：
#   ./scripts/pg_backup.sh                          # 默认备份到 ./backups
#   BACKUP_DIR=/data/backups ./scripts/pg_backup.sh  # 指定目录
#
# Docker Compose 内使用（在容器外执行，调用容器内的 pg_dump）：
#   docker compose exec -T postgres pg_dump -U ptis ptis | gzip > backups/ptis_backup_$(date +%Y%m%d_%H%M%S).sql.gz
#
# ─── cron job 参考用法 ───────────────────────────────────────────────
# 每天凌晨 2:00 自动备份（crontab -e 添加）：
#   0 2 * * * cd /path/to/ptis-web-v3 && ./scripts/pg_backup.sh >> /var/log/ptis_backup.log 2>&1
#
# 或者用 docker compose 的一行命令版本：
#   0 2 * * * cd /path/to/ptis-web-v3 && docker compose exec -T postgres pg_dump -U ptis ptis | gzip > /path/to/backups/ptis_backup_$(date +\%Y\%m\%d_\%H\%M\%S).sql.gz
# =============================================================================

set -euo pipefail

# ─── 可配置参数 ───────────────────────────────────────────────────────
BACKUP_DIR="${BACKUP_DIR:-./backups}"
KEEP_DAYS="${KEEP_DAYS:-7}"

# 数据库连接参数（与 docker-compose.yml 中保持一致）
PG_HOST="${PG_HOST:-localhost}"
PG_PORT="${PG_PORT:-5432}"
PG_USER="${PG_USER:-ptis}"
PG_DB="${PG_DB:-ptis}"
PG_PASSWORD="${PG_PASSWORD:-ptis_dev}"

export PGPASSWORD="${PG_PASSWORD}"

# ─── 时间戳 ───────────────────────────────────────────────────────────
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
FILENAME="ptis_backup_${TIMESTAMP}.sql.gz"
BACKUP_PATH="${BACKUP_DIR}/${FILENAME}"

# ─── 创建备份目录 ────────────────────────────────────────────────────
mkdir -p "${BACKUP_DIR}"

echo "[${TIMESTAMP}] 开始备份 PostgreSQL 数据库 '${PG_DB}' ..."

# ─── 检测运行模式 ────────────────────────────────────────────────────
if command -v docker &>/dev/null && docker compose version &>/dev/null 2>&1; then
    # 检查 postgres 容器是否运行中
    if docker compose ps postgres --format '{{.State}}' 2>/dev/null | grep -q 'running'; then
        echo "检测到 Docker Compose 环境，使用容器内 pg_dump ..."
        docker compose exec -T postgres \
            pg_dump -U "${PG_USER}" -d "${PG_DB}" \
            | gzip > "${BACKUP_PATH}"
    else
        # 容器未运行，尝试本地 pg_dump
        echo "Postgres 容器未运行，尝试本地 pg_dump ..."
        pg_dump -h "${PG_HOST}" -p "${PG_PORT}" -U "${PG_USER}" -d "${PG_DB}" \
            | gzip > "${BACKUP_PATH}"
    fi
else
    # 纯本地环境
    pg_dump -h "${PG_HOST}" -p "${PG_PORT}" -U "${PG_USER}" -d "${PG_DB}" \
        | gzip > "${BACKUP_PATH}"
fi

# ─── 验证备份文件 ────────────────────────────────────────────────────
if [ -f "${BACKUP_PATH}" ]; then
    FILE_SIZE=$(du -h "${BACKUP_PATH}" | cut -f1)
    echo "备份完成: ${BACKUP_PATH} (${FILE_SIZE})"
else
    echo "备份失败：文件未生成"
    exit 1
fi

# ─── 清理旧备份（保留最近 N 天） ─────────────────────────────────────
DELETED=$(find "${BACKUP_DIR}" -name "ptis_backup_*.sql.gz" -type f -mtime +"${KEEP_DAYS}" -print -delete | wc -l)
if [ "${DELETED}" -gt 0 ]; then
    echo "已清理 ${DELETED} 个超过 ${KEEP_DAYS} 天的旧备份"
fi

echo "[${TIMESTAMP}] 备份流程结束"
