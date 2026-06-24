-- 任务 #11+#12: 大屏聚合 + 分析域 org_id

-- load_characteristics: UNIQUE(customer_id, analysis_month) → UNIQUE(org_id, customer_id, analysis_month)
ALTER TABLE load_characteristics ADD COLUMN IF NOT EXISTS org_id uuid REFERENCES organizations(id);
UPDATE load_characteristics SET org_id = (SELECT id FROM organizations WHERE code='default') WHERE org_id IS NULL;
CREATE INDEX IF NOT EXISTS idx_load_characteristics_org ON load_characteristics(org_id);
ALTER TABLE load_characteristics DROP CONSTRAINT IF EXISTS load_characteristics_customer_id_analysis_month_key;
CREATE UNIQUE INDEX IF NOT EXISTS load_characteristics_org_cust_month_uniq ON load_characteristics(org_id, customer_id, analysis_month);

-- trade_strategies: UNIQUE(strategy_name) → UNIQUE(org_id, strategy_name)
ALTER TABLE trade_strategies ADD COLUMN IF NOT EXISTS org_id uuid REFERENCES organizations(id);
UPDATE trade_strategies SET org_id = (SELECT id FROM organizations WHERE code='default') WHERE org_id IS NULL;
CREATE INDEX IF NOT EXISTS idx_trade_strategies_org ON trade_strategies(org_id);
ALTER TABLE trade_strategies DROP CONSTRAINT IF EXISTS trade_strategies_strategy_name_key;
CREATE UNIQUE INDEX IF NOT EXISTS trade_strategies_org_name_uniq ON trade_strategies(org_id, strategy_name);
