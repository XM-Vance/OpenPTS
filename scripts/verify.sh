#!/usr/bin/env bash
# PTIS 端到端验证脚本：登录 → 调用 15+ 关键端点 → 输出 OK/FAIL。
# 用法：bash scripts/verify.sh
# 环境变量：BASE_URL（默认 http://localhost:8080）、ADMIN_USER / ADMIN_PASS

set -e
BASE_URL="${BASE_URL:-http://localhost:8080}"
ADMIN_USER="${ADMIN_USER:-admin}"
# 种子程序随机生成 admin 密码，运行前请通过 ADMIN_PASS 传入（见 seed 输出）
ADMIN_PASS="${ADMIN_PASS:?请通过 ADMIN_PASS 传入 admin 密码（种子程序启动时随机生成并打印）}"

pass=0
fail=0

check() {
  local name="$1"
  local cmd="$2"
  local expect="${3:-200}"
  status=$(eval "$cmd" 2>/dev/null || echo 0)
  if [ "$status" = "$expect" ]; then
    printf "  \033[32m✓\033[0m %s (%s)\n" "$name" "$status"
    pass=$((pass+1))
  else
    printf "  \033[31m✗\033[0m %s (期望 %s 实际 %s)\n" "$name" "$expect" "$status"
    fail=$((fail+1))
  fi
}

echo "→ 健康检查"
check "/health"  "curl -s -o /dev/null -w '%{http_code}' $BASE_URL/health"
check "/metrics" "curl -s -o /dev/null -w '%{http_code}' $BASE_URL/metrics"

echo "→ 登录"
TOKEN=$(curl -s -X POST "$BASE_URL/api/v1/auth/login" \
  -H 'Content-Type: application/json' \
  -d "{\"username\":\"$ADMIN_USER\",\"password\":\"$ADMIN_PASS\"}" \
  | sed -n 's/.*"token":"\([^"]*\)".*/\1/p')

if [ -z "$TOKEN" ]; then
  echo "  ✗ 登录失败"; exit 1
fi
echo "  ✓ token=${TOKEN:0:40}..."
AH="Authorization: Bearer $TOKEN"

echo "→ 业务核心端点"
check "/auth/me"                  "curl -s -o /dev/null -w '%{http_code}' -H '$AH' $BASE_URL/api/v1/auth/me"
check "/dashboard/summary"        "curl -s -o /dev/null -w '%{http_code}' -H '$AH' $BASE_URL/api/v1/dashboard/summary"
check "/customers"                "curl -s -o /dev/null -w '%{http_code}' -H '$AH' $BASE_URL/api/v1/customers"
check "/retail/contracts"         "curl -s -o /dev/null -w '%{http_code}' -H '$AH' $BASE_URL/api/v1/retail/contracts"
check "/settlement/daily"         "curl -s -o /dev/null -w '%{http_code}' -H '$AH' $BASE_URL/api/v1/settlement/daily"
check "/settlement/monthly"       "curl -s -o /dev/null -w '%{http_code}' -H '$AH' $BASE_URL/api/v1/settlement/monthly"
check "/price/trend/daily"        "curl -s -o /dev/null -w '%{http_code}' -H '$AH' $BASE_URL/api/v1/price/trend/daily"
check "/analytics/alerts"         "curl -s -o /dev/null -w '%{http_code}' -H '$AH' $BASE_URL/api/v1/analytics/alerts"
check "/storage/stations"         "curl -s -o /dev/null -w '%{http_code}' -H '$AH' $BASE_URL/api/v1/storage/stations"
check "/freq/clearing"            "curl -s -o /dev/null -w '%{http_code}' -H '$AH' $BASE_URL/api/v1/freq/clearing"
check "/scheduler/jobs"           "curl -s -o /dev/null -w '%{http_code}' -H '$AH' $BASE_URL/api/v1/scheduler/jobs"
check "/audit/logs"               "curl -s -o /dev/null -w '%{http_code}' -H '$AH' $BASE_URL/api/v1/audit/logs"
check "/approvals"                "curl -s -o /dev/null -w '%{http_code}' -H '$AH' $BASE_URL/api/v1/approvals"
check "/approvals/templates"      "curl -s -o /dev/null -w '%{http_code}' -H '$AH' '$BASE_URL/api/v1/approvals/templates?resource=retail_contracts'"
check "/system/security/overview" "curl -s -o /dev/null -w '%{http_code}' -H '$AH' '$BASE_URL/api/v1/system/security/overview?hours=24'"
check "/docs (Swagger UI)"        "curl -s -o /dev/null -w '%{http_code}' $BASE_URL/docs"
check "/docs/openapi.yaml"        "curl -s -o /dev/null -w '%{http_code}' $BASE_URL/docs/openapi.yaml"

echo
echo "→ 限流验证（连续 5 次登录，第 4+ 次应返回 429）"
for i in 1 2 3 4 5; do
  code=$(curl -s -o /dev/null -w '%{http_code}' \
    -X POST -H 'Content-Type: application/json' \
    -d '{"username":"_bogus","password":"x"}' \
    "$BASE_URL/api/v1/auth/login")
  echo "  第 $i 次：$code"
done

echo
printf "\033[32m通过: %d\033[0m  \033[31m失败: %d\033[0m\n" "$pass" "$fail"
exit $((fail > 0 ? 1 : 0))
