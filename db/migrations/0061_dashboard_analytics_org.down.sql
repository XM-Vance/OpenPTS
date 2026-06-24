-- 任务 #11+#12 回滚: 大屏聚合 + 分析域 org_id

-- trade_strategies
DROP INDEX IF EXISTS trade_strategies_org_name_uniq;
DROP INDEX IF EXISTS idx_trade_strategies_org;
ALTER TABLE trade_strategies DROP COLUMN IF EXISTS org_id;
ALTER TABLE trade_strategies ADD CONSTRAINT trade_strategies_strategy_name_key UNIQUE (strategy_name);

-- load_characteristics
DROP INDEX IF EXISTS load_characteristics_org_cust_month_uniq;
DROP INDEX IF EXISTS idx_load_characteristics_org;
ALTER TABLE load_characteristics DROP COLUMN IF EXISTS org_id;
ALTER TABLE load_characteristics ADD CONSTRAINT load_characteristics_customer_id_analysis_month_key UNIQUE (customer_id, analysis_month);
