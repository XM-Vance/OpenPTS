-- 回滚 0107：预测/电网/碳/告警/负荷等域 org_id（仅在一次性测试库执行；多省存量会冲突，生产/共享库不得 down）。

-- system_alerts
DROP INDEX IF EXISTS idx_system_alerts_org;
ALTER TABLE system_alerts DROP COLUMN IF EXISTS org_id;

-- real_time_generation
DROP INDEX IF EXISTS real_time_generation_org_uniq;
DROP INDEX IF EXISTS idx_real_time_generation_org;
ALTER TABLE real_time_generation DROP COLUMN IF EXISTS org_id;
ALTER TABLE real_time_generation ADD CONSTRAINT real_time_generation_date_period_source_type_key UNIQUE (date, period, source_type);

-- daily_release
DROP INDEX IF EXISTS daily_release_org_uniq;
DROP INDEX IF EXISTS idx_daily_release_org;
ALTER TABLE daily_release DROP COLUMN IF EXISTS org_id;
ALTER TABLE daily_release ADD CONSTRAINT daily_release_release_date_key UNIQUE (release_date);

-- unified_load_curve
DROP INDEX IF EXISTS unified_load_curve_org_uniq;
DROP INDEX IF EXISTS idx_unified_load_curve_org;
ALTER TABLE unified_load_curve DROP COLUMN IF EXISTS org_id;
ALTER TABLE unified_load_curve ADD CONSTRAINT unified_load_curve_customer_id_date_key UNIQUE (customer_id, date);

-- mechanism_energy_monthly
DROP INDEX IF EXISTS mechanism_energy_monthly_org_uniq;
DROP INDEX IF EXISTS idx_mechanism_energy_monthly_org;
ALTER TABLE mechanism_energy_monthly DROP COLUMN IF EXISTS org_id;
ALTER TABLE mechanism_energy_monthly ADD CONSTRAINT mechanism_energy_monthly_month_key UNIQUE (month);

-- actual_operation
DROP INDEX IF EXISTS idx_actual_operation_org;
ALTER TABLE actual_operation DROP COLUMN IF EXISTS org_id;

-- medium_term_load_forecast
DROP INDEX IF EXISTS idx_medium_term_load_forecast_org;
ALTER TABLE medium_term_load_forecast DROP COLUMN IF EXISTS org_id;

-- storage_schedules
DROP INDEX IF EXISTS idx_storage_schedules_org;
ALTER TABLE storage_schedules DROP COLUMN IF EXISTS org_id;

-- tariff_schemes
DROP INDEX IF EXISTS idx_tariff_schemes_org;
ALTER TABLE tariff_schemes DROP COLUMN IF EXISTS org_id;

-- carbon_emissions
DROP INDEX IF EXISTS idx_carbon_emissions_org;
ALTER TABLE carbon_emissions DROP COLUMN IF EXISTS org_id;

-- market_clearings（独立唯一索引，重建为原名）
DROP INDEX IF EXISTS market_clearings_org_uniq;
DROP INDEX IF EXISTS idx_market_clearings_org;
ALTER TABLE market_clearings DROP COLUMN IF EXISTS org_id;
CREATE UNIQUE INDEX idx_clearing_unique ON market_clearings(market_type, trade_date, period);

-- grid_calculations
DROP INDEX IF EXISTS idx_grid_calculations_org;
ALTER TABLE grid_calculations DROP COLUMN IF EXISTS org_id;

-- forecasts
DROP INDEX IF EXISTS idx_forecasts_org;
ALTER TABLE forecasts DROP COLUMN IF EXISTS org_id;
