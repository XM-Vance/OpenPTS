-- 回滚：删除测算行后恢复原 org 唯一键（否则测算/实际同键撞约束），再删列。
DROP INDEX IF EXISTS idx_customer_profit_is_estimate;
DROP INDEX IF EXISTS customer_profit_org_uniq;
DELETE FROM customer_profit WHERE is_estimate = true;
CREATE UNIQUE INDEX customer_profit_org_uniq
    ON customer_profit (org_id, customer_id, operating_month);
ALTER TABLE customer_profit DROP COLUMN IF EXISTS is_estimate;
