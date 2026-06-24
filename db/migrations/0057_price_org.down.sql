-- 任务 7 回滚: 价格域 org_id
DROP INDEX IF EXISTS grid_agency_price_org_uniq;
DROP INDEX IF EXISTS idx_grid_agency_price_org;
ALTER TABLE grid_agency_price DROP COLUMN IF EXISTS org_id;
ALTER TABLE grid_agency_price ADD CONSTRAINT grid_agency_price_operating_month_voltage_level_key UNIQUE (operating_month, voltage_level);

DROP INDEX IF EXISTS idx_tou_rules_org;
ALTER TABLE tou_rules DROP COLUMN IF EXISTS org_id;

DROP INDEX IF EXISTS day_ahead_spot_price_org_uniq;
DROP INDEX IF EXISTS idx_day_ahead_spot_price_org;
ALTER TABLE day_ahead_spot_price DROP COLUMN IF EXISTS org_id;
ALTER TABLE day_ahead_spot_price ADD CONSTRAINT day_ahead_spot_price_date_period_key UNIQUE (date, period);

DROP INDEX IF EXISTS contract_price_daily_org_uniq;
DROP INDEX IF EXISTS idx_contract_price_daily_org;
ALTER TABLE contract_price_daily DROP COLUMN IF EXISTS org_id;
ALTER TABLE contract_price_daily ADD CONSTRAINT contract_price_daily_contract_id_price_date_key UNIQUE (contract_id, price_date);
