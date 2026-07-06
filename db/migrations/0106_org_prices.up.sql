-- 多租户隔离收口（批次二）：价格域补 org_id。
-- 审查报告 P0-D1：0057 给 day_ahead_spot_price 补了 org_id，但同批 real_time_spot_price
-- 等价格表遗漏，导致各省价格数据互相覆盖/串读。本迁移补齐。样板参照 0101/0057。

-- real_time_spot_price（子句 UNIQUE(date, period) → 含 org）
ALTER TABLE real_time_spot_price ADD COLUMN IF NOT EXISTS org_id uuid REFERENCES organizations(id);
UPDATE real_time_spot_price SET org_id = (SELECT id FROM organizations WHERE code='FJ') WHERE org_id IS NULL;
CREATE INDEX IF NOT EXISTS idx_real_time_spot_price_org ON real_time_spot_price(org_id);
ALTER TABLE real_time_spot_price DROP CONSTRAINT IF EXISTS real_time_spot_price_date_period_key;
CREATE UNIQUE INDEX IF NOT EXISTS real_time_spot_price_org_uniq ON real_time_spot_price(org_id, date, period);

-- day_ahead_econ_price（子句 UNIQUE(date, period) → 含 org）
ALTER TABLE day_ahead_econ_price ADD COLUMN IF NOT EXISTS org_id uuid REFERENCES organizations(id);
UPDATE day_ahead_econ_price SET org_id = (SELECT id FROM organizations WHERE code='FJ') WHERE org_id IS NULL;
CREATE INDEX IF NOT EXISTS idx_day_ahead_econ_price_org ON day_ahead_econ_price(org_id);
ALTER TABLE day_ahead_econ_price DROP CONSTRAINT IF EXISTS day_ahead_econ_price_date_period_key;
CREATE UNIQUE INDEX IF NOT EXISTS day_ahead_econ_price_org_uniq ON day_ahead_econ_price(org_id, date, period);

-- node_spot_price_daily（子句 UNIQUE(date, period, node_id) → 含 org）
ALTER TABLE node_spot_price_daily ADD COLUMN IF NOT EXISTS org_id uuid REFERENCES organizations(id);
UPDATE node_spot_price_daily SET org_id = (SELECT id FROM organizations WHERE code='FJ') WHERE org_id IS NULL;
CREATE INDEX IF NOT EXISTS idx_node_spot_price_daily_org ON node_spot_price_daily(org_id);
ALTER TABLE node_spot_price_daily DROP CONSTRAINT IF EXISTS node_spot_price_daily_date_period_node_id_key;
CREATE UNIQUE INDEX IF NOT EXISTS node_spot_price_daily_org_uniq ON node_spot_price_daily(org_id, date, period, node_id);

-- price_sgcc（列级 month UNIQUE → 含 org）
ALTER TABLE price_sgcc ADD COLUMN IF NOT EXISTS org_id uuid REFERENCES organizations(id);
UPDATE price_sgcc SET org_id = (SELECT id FROM organizations WHERE code='FJ') WHERE org_id IS NULL;
CREATE INDEX IF NOT EXISTS idx_price_sgcc_org ON price_sgcc(org_id);
ALTER TABLE price_sgcc DROP CONSTRAINT IF EXISTS price_sgcc_month_key;
CREATE UNIQUE INDEX IF NOT EXISTS price_sgcc_org_uniq ON price_sgcc(org_id, month);

-- price_forecast_results（子句 UNIQUE(forecast_date, target_date, forecast_method) → 含 org）
ALTER TABLE price_forecast_results ADD COLUMN IF NOT EXISTS org_id uuid REFERENCES organizations(id);
UPDATE price_forecast_results SET org_id = (SELECT id FROM organizations WHERE code='FJ') WHERE org_id IS NULL;
CREATE INDEX IF NOT EXISTS idx_price_forecast_results_org ON price_forecast_results(org_id);
ALTER TABLE price_forecast_results DROP CONSTRAINT IF EXISTS price_forecast_results_forecast_date_target_date_forecast_method_key;
CREATE UNIQUE INDEX IF NOT EXISTS price_forecast_results_org_uniq ON price_forecast_results(org_id, forecast_date, target_date, forecast_method);

-- carbon_quotes（子句 UNIQUE(product, trade_date) → 含 org；碳价为全国共享参考数据，
-- 但当前业务表设计按省存储行情快照，故仍并入 org 维度保持与碳价走势页一致）
ALTER TABLE carbon_quotes ADD COLUMN IF NOT EXISTS org_id uuid REFERENCES organizations(id);
UPDATE carbon_quotes SET org_id = (SELECT id FROM organizations WHERE code='FJ') WHERE org_id IS NULL;
CREATE INDEX IF NOT EXISTS idx_carbon_quotes_org ON carbon_quotes(org_id);
ALTER TABLE carbon_quotes DROP CONSTRAINT IF EXISTS carbon_quotes_product_trade_date_key;
CREATE UNIQUE INDEX IF NOT EXISTS carbon_quotes_org_uniq ON carbon_quotes(org_id, product, trade_date);
