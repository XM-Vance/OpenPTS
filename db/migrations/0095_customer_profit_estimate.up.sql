-- Phase 3：客户利润区分「签约前测算」与「签约后实际结算」，
-- 避免销售预估污染财务实际（is_estimate=false 实际；true 测算/预估）。
ALTER TABLE customer_profit ADD COLUMN IF NOT EXISTS is_estimate boolean NOT NULL DEFAULT false;

-- 唯一键纳入 is_estimate：同省·同客户·同月下，测算与实际各一行可并存。
-- （原索引 customer_profit_org_uniq 由 0052 多租户改造建立，键为 org_id,customer_id,operating_month。）
DROP INDEX IF EXISTS customer_profit_org_uniq;
CREATE UNIQUE INDEX customer_profit_org_uniq
    ON customer_profit (org_id, customer_id, operating_month, is_estimate);

CREATE INDEX IF NOT EXISTS idx_customer_profit_is_estimate ON customer_profit (is_estimate);
