-- 回滚 0055: p0/p1 表 org_id

-- pre_settlement_daily
DROP INDEX IF EXISTS pre_settlement_daily_org_uniq;
DROP INDEX IF EXISTS idx_pre_settlement_daily_org;
ALTER TABLE pre_settlement_daily DROP COLUMN IF EXISTS org_id;
ALTER TABLE pre_settlement_daily ADD CONSTRAINT pre_settlement_daily_operating_date_key UNIQUE (operating_date);

-- total_load_daily
DROP INDEX IF EXISTS total_load_daily_org_uniq;
DROP INDEX IF EXISTS idx_total_load_daily_org;
ALTER TABLE total_load_daily DROP COLUMN IF EXISTS org_id;
ALTER TABLE total_load_daily ADD CONSTRAINT total_load_daily_data_date_key UNIQUE (data_date);

-- forecast_accuracy
DROP INDEX IF EXISTS forecast_accuracy_org_uniq;
DROP INDEX IF EXISTS idx_forecast_accuracy_org;
ALTER TABLE forecast_accuracy DROP COLUMN IF EXISTS org_id;
ALTER TABLE forecast_accuracy ADD CONSTRAINT forecast_accuracy_forecast_target_forecast_date_key UNIQUE (forecast_target, forecast_date);

-- mechanism_energy_plan
DROP INDEX IF EXISTS mechanism_energy_plan_org_uniq;
DROP INDEX IF EXISTS idx_mechanism_energy_plan_org;
ALTER TABLE mechanism_energy_plan DROP COLUMN IF EXISTS org_id;
ALTER TABLE mechanism_energy_plan ADD CONSTRAINT mechanism_energy_plan_operating_month_voltage_level_key UNIQUE (operating_month, voltage_level);

-- medium_load_forecast
DROP INDEX IF EXISTS medium_load_forecast_org_uniq;
DROP INDEX IF EXISTS idx_medium_load_forecast_org;
ALTER TABLE medium_load_forecast DROP COLUMN IF EXISTS org_id;
ALTER TABLE medium_load_forecast ADD CONSTRAINT medium_load_forecast_forecast_month_key UNIQUE (forecast_month);

-- market_analysis_daily
DROP INDEX IF EXISTS market_analysis_daily_org_uniq;
DROP INDEX IF EXISTS idx_market_analysis_daily_org;
ALTER TABLE market_analysis_daily DROP COLUMN IF EXISTS org_id;
ALTER TABLE market_analysis_daily ADD CONSTRAINT market_analysis_daily_trade_date_key UNIQUE (trade_date);

-- holidays
DROP INDEX IF EXISTS holidays_org_uniq;
DROP INDEX IF EXISTS idx_holidays_org;
ALTER TABLE holidays DROP COLUMN IF EXISTS org_id;
ALTER TABLE holidays ADD CONSTRAINT holidays_holiday_date_key UNIQUE (holiday_date);

-- typical_curves
DROP INDEX IF EXISTS typical_curves_org_uniq;
DROP INDEX IF EXISTS idx_typical_curves_org;
ALTER TABLE typical_curves DROP COLUMN IF EXISTS org_id;
ALTER TABLE typical_curves ADD CONSTRAINT typical_curves_name_key UNIQUE (name);
