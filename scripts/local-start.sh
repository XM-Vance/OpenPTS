#!/bin/bash
# =============================================================
# PTIS V1 本地原生启动脚本（不用 Docker）
# 前置：brew PostgreSQL@16 已启动
# 用法：
#   bash scripts/local-start.sh        启动
#   bash scripts/local-start.sh stop   停止
# =============================================================

PROJECT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
PID_DIR="$PROJECT_DIR/.pids"
LOG_DIR="$PROJECT_DIR/.logs"

mkdir -p "$PID_DIR" "$LOG_DIR"

is_running() {
    local pid_file="$1"
    [ -f "$pid_file" ] && kill -0 "$(cat "$pid_file")" 2>/dev/null
}

stop_service() {
    local name="$1" pid_file="$PID_DIR/$name.pid"
    if is_running "$pid_file"; then
        kill "$(cat "$pid_file")" 2>/dev/null
        rm -f "$pid_file"
        echo "  ✅ $name 已停止"
    else
        echo "  ⏭️  $name 未运行"
    fi
}

case "${1:-start}" in
    start)
        echo "========================================="
        echo "  PTIS V1 原生启动"
        echo "========================================="

        # 清理残留进程
        pkill -f "next-server" 2>/dev/null
        sleep 1

        # 检查 PostgreSQL
        if ! psql -U ptis -d ptis -c 'SELECT 1' > /dev/null 2>&1; then
            if psql -U vance -d ptis -c 'SELECT 1' > /dev/null 2>&1; then
                echo "  ✅ PostgreSQL 已连接 (vance 用户)"
            else
                echo "  ❌ PostgreSQL 未启动或无权限，请先："
                echo "     brew services start postgresql@16"
                exit 1
            fi
        else
            echo "  ✅ PostgreSQL 已连接"
        fi

        # ─── 后端 ───
        BPID="$PID_DIR/backend.pid"
        BLOG="$LOG_DIR/backend.log"
        [ -f "$BPID" ] && kill "$(cat "$BPID")" 2>/dev/null && rm -f "$BPID"
        echo "  启动 backend ..."
        (cd "$PROJECT_DIR" && exec ./backend/ptis-server) > "$BLOG" 2>&1 &
        echo $! > "$BPID"
        sleep 2
        if kill -0 $(cat "$BPID") 2>/dev/null; then
            echo "  ✅ backend (PID $(cat "$BPID"))"
        else
            echo "  ❌ backend 启动失败："
            tail -3 "$BLOG"
        fi

        # ─── 前端 ───
        FPID="$PID_DIR/frontend.pid"
        FLOG="$LOG_DIR/frontend.log"
        [ -f "$FPID" ] && kill "$(cat "$FPID")" 2>/dev/null && rm -f "$FPID"
        echo "  启动 frontend ..."
        (cd "$PROJECT_DIR/frontend" && NEXT_PUBLIC_API_BASE=http://localhost:8080 exec npx next dev) > "$FLOG" 2>&1 &
        echo $! > "$FPID"
        sleep 3
        if kill -0 $(cat "$FPID") 2>/dev/null; then
            echo "  ✅ frontend (PID $(cat "$FPID"))"
        else
            echo "  ❌ frontend 启动失败："
            tail -3 "$FLOG"
        fi

        echo ""
        echo "========================================="
        echo "  🎉 服务已启动"
        echo "========================================="
        echo "  前端：http://localhost:3000"
        echo "  后端：http://localhost:8080/health"
        echo "  算法：http://localhost:8200/health"
        echo ""
        echo "  日志目录：$LOG_DIR/"
        echo "  停止：bash scripts/local-start.sh stop"
        echo "========================================="
        ;;

    stop)
        echo "停止 PTIS V1 ..."
        stop_service frontend
        stop_service algo
        stop_service backend
        echo "✅ 全部停止"
        ;;

    status)
        for name in backend algo frontend; do
            if is_running "$PID_DIR/$name.pid"; then
                echo "  ✅ $name (PID $(cat "$PID_DIR/$name.pid"))"
            else
                echo "  ❌ $name 未运行"
            fi
        done
        ;;

    *)
        echo "用法: bash scripts/local-start.sh {start|stop|status}"
        ;;
esac
