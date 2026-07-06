-- 回填属数据修正、不可逆（无法区分"本就 NULL"与"回填得来"）；
-- 把正确的 org_id 改回 NULL 会重新引入跨省泄漏，故 down 为 no-op。
SELECT 1;
