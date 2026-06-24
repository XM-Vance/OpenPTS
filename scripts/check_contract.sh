#!/usr/bin/env bash
# ──────────────────────────────────────────────────────────────
# PTIS 契约闸脚本
# 从前端 lib/api 及 app 目录提取所有 /api/v1/... 路径，
# 逐条 curl 后端验证；404 即报告失败。
# ──────────────────────────────────────────────────────────────
set -euo pipefail

BASE_URL="${1:-http://localhost:8080}"
FRONTEND_DIR="${2:-frontend}"
MAX_WAIT="${CONTRACT_MAX_WAIT:-60}"   # 等后端启动的最大秒数

# ── 1. 等待后端就绪 ──────────────────────────────────────────
echo "⏳ 等待后端就绪 (${BASE_URL}) ..."
elapsed=0
until curl -sf -o /dev/null "${BASE_URL}/health" 2>/dev/null; do
  sleep 2
  elapsed=$((elapsed + 2))
  if [ "$elapsed" -ge "$MAX_WAIT" ]; then
    echo "::error::后端在 ${MAX_WAIT}s 内未就绪，放弃"
    exit 1
  fi
done
echo "✅ 后端已就绪"

# ── 2. 提取「可达」api 模块的 /api/v1/... 路径 ────────────────────────
# 只扫被 app/components/hooks 真正 import 的 api 模块（剔除无人引用的死/重复模块），
# 避免对死代码误报。新增页面调用、或活模块出现的缺口仍会被发现。
PATHS=$(mktemp)
trap 'rm -f "$PATHS"' EXIT
: > "$PATHS"

for f in "${FRONTEND_DIR}"/lib/api/*.ts; do
  base=$(basename "$f" .ts)
  [ "$base" = "client" ] && continue
  # 模块是否被任一页面/组件/Hook import（单双引号皆可）
  if grep -rqsE "api/${base}[\"']" \
      "${FRONTEND_DIR}/app" "${FRONTEND_DIR}/components" "${FRONTEND_DIR}/lib/hooks" 2>/dev/null; then
    grep -hoE '/api/v1/[a-zA-Z0-9_/.-]+' "$f" 2>/dev/null | sed 's:/*$::' >> "$PATHS" || true
  fi
done
# 页面内联书写的 /api/v1 路径也纳入
grep -rhoE '/api/v1/[a-zA-Z0-9_/.-]+' "${FRONTEND_DIR}/app" 2>/dev/null | sed 's:/*$::' >> "$PATHS" || true
sort -u "$PATHS" -o "$PATHS"

TOTAL=$(wc -l < "$PATHS")
if [ "$TOTAL" -eq 0 ]; then
  echo "⚠️  未找到任何 /api/v1/... 路径，跳过契约检查"
  exit 0
fi
echo "📋 共发现 ${TOTAL} 条 API 路径，开始校验..."
echo "────────────────────────────────────────"

# 已知待办白名单：后端端点尚未实现、已登记跟踪的路径。
# 这些 404 只告警、不致 CI 失败；新出现的 404 仍会硬失败。
# 清空一项即代表该端点已补齐（应同步删除此处）。
#
# 2026-06：曾登记 8 条历史遗留路径；对应前端死函数已全部删除
# （不再被扫描），条目随之清空。当前应保持为空。
KNOWN_PENDING="
"

# ── 3. 逐条 curl 校验 ───────────────────────────────────────
PASS=0
FAIL=0
WARN=0
FAILED_PATHS=""

while IFS= read -r path; do
  HTTP_CODE=$(curl -s -o /dev/null -w '%{http_code}' "${BASE_URL}${path}" 2>/dev/null || true)

  if [ "$HTTP_CODE" = "404" ]; then
    if echo "$KNOWN_PENDING" | grep -qxF "$path"; then
      echo "⚠️  404  ${path}  (已登记待办，不阻断)"
      WARN=$((WARN + 1))
    else
      echo "❌  404  ${path}"
      FAIL=$((FAIL + 1))
      FAILED_PATHS="${FAILED_PATHS}\\n  ${path}"
    fi
  else
    # 200/401/403/500 等都算"路由存在"，只有 404 算契约失败
    echo "✅  ${HTTP_CODE}  ${path}"
    PASS=$((PASS + 1))
  fi
done < "$PATHS"

# ── 4. 汇总 ──────────────────────────────────────────────────
echo "────────────────────────────────────────"
echo "结果: ✅ ${PASS} 通过  ⚠️ ${WARN} 待办  ❌ ${FAIL} 失败  (共 ${TOTAL})"

if [ "$FAIL" -ne 0 ]; then
  echo ""
  echo "❌ 以下路径返回 404（后端未实现）:"
  echo -e "$FAILED_PATHS"
  exit 1
fi

echo "🎉 契约闸通过！"
exit 0
