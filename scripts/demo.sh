#!/usr/bin/env bash
# OpenPTS 一键演示脚本（本地，无需 Docker）。
#
# 做这些事：
#   1. 检查 PostgreSQL / migrate CLI / Node 是否可用
#   2. 建专用演示库 openpts_demo（不复用你已有的库）
#   3. 跑全部数据库迁移
#   4. 创建 admin 账号（演示用固定密码 admin123）
#   5. 启动后端（DEMO_MODE=true，预测端点返回合成数据）
#   6. 注入演示数据（客户/负荷/价格/结算等）
#   7. 启动前端
#   8. 打印访问地址与登录账号
#
# 用法：  bash scripts/demo.sh
# 停止：  bash scripts/demo.sh stop
# 重置：  bash scripts/demo.sh reset   （删演示库重来）
set -euo pipefail

PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DB_NAME="openpts_demo"
BACKEND_PORT="${BACKEND_PORT:-8090}"
FRONTEND_PORT="${FRONTEND_PORT:-3000}"
ADMIN_USER="admin"
ADMIN_PASS="admin123"   # 演示用固定密码，便于记忆；生产请用 make seed（随机密码）
LOG_DIR="/tmp/openpts-demo"
PID_DIR="/tmp/openpts-demo"

# PG 连接（默认本地，可用环境变量覆盖）
PG_USER="${PG_USER:-$(whoami)}"
PG_HOST="${PG_HOST:-localhost}"
PG_PORT="${PG_PORT:-5432}"
DATABASE_URL="postgres://${PG_USER}@${PG_HOST}:${PG_PORT}/${DB_NAME}?sslmode=disable"

mkdir -p "$LOG_DIR" "$PID_DIR"

# ── 颜色 ──
G() { printf "\033[32m%s\033[0m\n" "$1"; }
Y() { printf "\033[33m%s\033[0m\n" "$1"; }
R() { printf "\033[31m%s\033[0m\n" "$1"; }
B() { printf "\033[36m%s\033[0m\n" "$1"; }

check_cmd() {
  command -v "$1" >/dev/null 2>&1 || { R "❌ 缺少依赖：$1（请先安装）"; exit 1; }
}

# ── 停止 ──
do_stop() {
  B "停止 OpenPTS 演示服务..."
  for name in backend frontend; do
    pidf="$PID_DIR/$name.pid"
    if [ -f "$pidf" ]; then
      pid="$(cat "$pidf")"
      kill "$pid" 2>/dev/null && G "  ✅ 已停止 $name (PID $pid)" || Y "  $name 未在运行"
      rm -f "$pidf"
    fi
  done
  # 清理可能残留的 go run 子进程 / next dev
  pkill -f "go run ./cmd/server.*$BACKEND_PORT" 2>/dev/null || true
  pkill -f "next dev.*$FRONTEND_PORT" 2>/dev/null || true
  G "已停止。演示库 $DB_NAME 保留（reset 可删除）。"
}

# ── 重置 ──
do_reset() {
  do_stop || true
  B "删除演示库 $DB_NAME..."
  psql -h "$PG_HOST" -p "$PG_PORT" -U "$PG_USER" -d postgres -c \
    "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname='$DB_NAME' AND pid<>pg_backend_pid();" >/dev/null 2>&1 || true
  psql -h "$PG_HOST" -p "$PG_PORT" -U "$PG_USER" -d postgres -c "DROP DATABASE IF EXISTS $DB_NAME;" >/dev/null 2>&1
  G "✅ 已删除 $DB_NAME"
}

# ── 启动 ──
do_start() {
  B "════════════════════════════════════════════"
  B "  OpenPTS 一键演示"
  B "════════════════════════════════════════════"

  # 1. 检查依赖
  B "[1/8] 检查依赖..."
  check_cmd psql
  check_cmd migrate
  check_cmd node
  check_cmd go
  G "  ✅ psql / migrate / node / go 均可用"

  # 2. 建库
  B "[2/8] 创建演示数据库 $DB_NAME..."
  psql -h "$PG_HOST" -p "$PG_PORT" -U "$PG_USER" -d postgres -tc \
    "SELECT 1 FROM pg_database WHERE datname='$DB_NAME'" | grep -q 1 || \
    psql -h "$PG_HOST" -p "$PG_PORT" -U "$PG_USER" -d postgres -c "CREATE DATABASE $DB_NAME;" >/dev/null
  G "  ✅ 数据库就绪"

  # 3. 迁移
  B "[3/8] 执行数据库迁移..."
  migrate -path "$PROJECT_DIR/db/migrations" \
    -database "$DATABASE_URL" up >/dev/null 2>&1 || { R "❌ 迁移失败"; exit 1; }
  G "  ✅ 迁移完成"

  # 4. seed（固定密码）
  B "[4/8] 创建管理员账号..."
  (cd "$PROJECT_DIR/backend" && \
    DATABASE_URL="$DATABASE_URL" JWT_SECRET="demo-only-not-for-production-use-32chars" \
    go run ./cmd/seed -username "$ADMIN_USER" -password "$ADMIN_PASS" >/dev/null 2>&1) || true
  G "  ✅ 账号就绪（admin / admin123）"

  # 5. 启动后端（demo 模式）
  B "[5/8] 启动后端（DEMO_MODE，端口 ${BACKEND_PORT}）..."
  JWT=$(openssl rand -base64 48 2>/dev/null || echo "demo-only-insecure-secret-min-32-characters-long-xx")
  (cd "$PROJECT_DIR/backend" && \
    PORT="$BACKEND_PORT" ENVIRONMENT=development LOG_LEVEL=warn \
    DATABASE_URL="$DATABASE_URL" JWT_SECRET="$JWT" JWT_TTL_HOURS=8 \
    DEMO_MODE=true \
    go run ./cmd/server >"$LOG_DIR/backend.log" 2>&1 &) 
  # 记录 go run 父进程（取最新的）
  sleep 1
  pgrep -f "go run ./cmd/server" | tail -1 > "$PID_DIR/backend.pid" 2>/dev/null || true
  # 等后端就绪
  for i in $(seq 1 20); do
    curl -sf "http://localhost:$BACKEND_PORT/health" >/dev/null 2>&1 && break
    sleep 1
  done
  curl -sf "http://localhost:$BACKEND_PORT/health" >/dev/null 2>&1 && G "  ✅ 后端已启动" || { R "❌ 后端启动失败，见 $LOG_DIR/backend.log"; tail -5 "$LOG_DIR/backend.log"; exit 1; }

  # 6. 注入演示数据
  B "[6/8] 注入演示数据..."
  TOKEN=$(curl -s -X POST "http://localhost:$BACKEND_PORT/api/v1/auth/login" \
    -H 'Content-Type: application/json' \
    -d "{\"username\":\"$ADMIN_USER\",\"password\":\"$ADMIN_PASS\"}" \
    | sed -n 's/.*"token":"\([^"]*\)".*/\1/p')
  ORGID=$(curl -s "http://localhost:$BACKEND_PORT/api/v1/auth/me" \
    -H "Authorization: Bearer $TOKEN" \
    | sed -n 's/.*"id":"\([0-9a-f-]*\)".*code.*default.*/\1/p' | head -1)
  H_AUTH=(-H "Authorization: Bearer $TOKEN" -H "X-Org-Id: $ORGID" -H 'Content-Type: application/json')
  # 先给负荷/价格建演示数据（预测端点需要历史）
  for ep in "/api/v1/load/demo-data" "/api/v1/price/demo-data" "/api/v1/settlement/demo-data" \
            "/api/v1/analytics/demo-data" "/api/v1/weather/demo-data" "/api/v1/price/tou-rules/demo-data"; do
    curl -sf -X POST "http://localhost:$BACKEND_PORT$ep" "${H_AUTH[@]}" \
      -d '{"days":30}' >/dev/null 2>&1 || true
  done
  G "  ✅ 演示数据已注入"

  # 7. 启动前端
  B "[7/8] 启动前端（端口 ${FRONTEND_PORT}）..."
  (cd "$PROJECT_DIR/frontend" && \
    NEXT_PUBLIC_API_BASE="http://localhost:$BACKEND_PORT" \
    npm run dev >"$LOG_DIR/frontend.log" 2>&1 &) 
  sleep 1
  pgrep -f "next dev" | tail -1 > "$PID_DIR/frontend.pid" 2>/dev/null || true
  for i in $(seq 1 25); do
    curl -sf "http://localhost:$FRONTEND_PORT" >/dev/null 2>&1 && break
    sleep 1
  done
  curl -sf "http://localhost:$FRONTEND_PORT" >/dev/null 2>&1 && G "  ✅ 前端已启动" || { R "❌ 前端启动失败，见 $LOG_DIR/frontend.log"; tail -5 "$LOG_DIR/frontend.log"; }

  # 8. 打印信息
  B "[8/8] 完成！"
  echo ""
  G "════════════════════════════════════════════════════"
  G "  🎉 OpenPTS 演示已就绪"
  G "════════════════════════════════════════════════════"
  echo ""
  B "  访问地址：http://localhost:$FRONTEND_PORT"
  B "  登录账号：$ADMIN_USER"
  B "  密码：    $ADMIN_PASS"
  echo ""
  Y "  ⚠️ 演示模式（DEMO_MODE）：预测端点返回合成数据，非真实算法。"
  Y "  ⚠️ 演示用固定密码 admin123，请勿用于生产。"
  echo ""
  B "  停止：  bash scripts/demo.sh stop"
  B "  重置：  bash scripts/demo.sh reset"
  B "  日志：  $LOG_DIR/{backend,frontend}.log"
  echo ""
}

case "${1:-start}" in
  start) do_start ;;
  stop)  do_stop ;;
  reset) do_reset ;;
  *) echo "用法: bash scripts/demo.sh [start|stop|reset]"; exit 1 ;;
esac
