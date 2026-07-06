-- 回滚 0106：价格域 org_id（仅在一次性测试库执行；多省存量会冲突，生产/共享库不得 down）。

-- carbon_quotes
DROP INDEX IF EXISTS carbon_quotes_org_uniq;
DROP INDEX IF EXISTS idx_carbon_quotes_org;
ALTER TABLE carbon_quotes DROP COLUMN IF EXISTS org_id;
ALTER TABLE carbon_quotes ADD CONSTRAINT carbon_quotes_product_trade_date_key UNIQUE (product, trade_date);

-- price_forecast_results
DROP INDEX IF EXISTS price_forecast_results_org_uniq;
DROP INDEX IF EXISTS idx_price_forecast_results_org;
ALTER TABLE price_forecast_results DROP COLUMN IF EXISTS org_id;
ALTER TABLE price_forecast_results ADD CONSTRAINT price_forecast_results_forecast_date_target_date_forecast_method_key UNIQUE (forecast_date, target_date, forecast_method);

-- price_sgcc
DROP INDEX IF EXISTS price_sgcc_org_uniq;
DROP INDEX IF EXISTS idx_price_sgcc_org;
ALTER TABLE price_sgcc DROP COLUMN IF EXISTS org_id;
ALTER TABLE price_sgcc ADD CONSTRAINT price_sgcc_month_key UNIQUE (month);

-- node_spot_price_daily
DROP INDEX IF EXISTS node_spot_price_daily_org_uniq;
DROP INDEX IF EXISTS idx_node_spot_price_daily_org;
ALTER TABLE node_spot_price_daily DROP COLUMN IF EXISTS org_id;
ALTER TABLE node_spot_price_daily ADD CONSTRAINT node_spot_price_daily_date_period_node_id_key UNIQUE (date, period, node_id);

-- day_ahead_econ_price
DROP INDEX IF EXISTS day_ahead_econ_price_org_uniq;
DROP INDEX IF EXISTS idx_day_ahead_econ_price_org;
ALTER TABLE day_ahead_econ_price DROP COLUMN IF EXISTS org_id;
ALTER TABLE day_ahead_econ_price ADD CONSTRAINT day_ahead_econ_price_date_period_key UNIQUE (date, period);

-- real_time_spot_price
DROP INDEX IF EXISTS real_time_spot_price_org_uniq;
DROP INDEX IF EXISTS idx_real_time_spot_price_org;
ALTER TABLE real_time_spot_price DROP COLUMN IF EXISTS org_id;
ALTER TABLE real_time_spot_price ADD CONSTRAINT real_time_spot_price_date_period_key UNIQUE (date, period);
