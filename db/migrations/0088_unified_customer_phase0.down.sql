-- 回滚 Phase 0（仅一次性库验证可逆性时使用）。
-- 注：不动 intent_customers.org_id（那是 0052 的，不属本迁移）。
DROP TABLE IF EXISTS customer_diagnosis;

DROP INDEX IF EXISTS idx_customers_agent_id;
DROP INDEX IF EXISTS idx_customers_lifecycle_stage;
ALTER TABLE customers DROP CONSTRAINT IF EXISTS customers_lifecycle_stage_chk;
ALTER TABLE customers DROP COLUMN IF EXISTS agent_id;
ALTER TABLE customers DROP COLUMN IF EXISTS lifecycle_stage;
