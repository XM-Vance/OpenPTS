-- 0045 回滚：UUID → BIGSERIAL（仅用于回退）

ALTER TABLE real_time_spot_price DROP CONSTRAINT real_time_spot_price_pkey;
ALTER TABLE real_time_spot_price DROP COLUMN id;
ALTER TABLE real_time_spot_price ADD COLUMN id BIGSERIAL PRIMARY KEY;

ALTER TABLE day_ahead_spot_price DROP CONSTRAINT day_ahead_spot_price_pkey;
ALTER TABLE day_ahead_spot_price DROP COLUMN id;
ALTER TABLE day_ahead_spot_price ADD COLUMN id BIGSERIAL PRIMARY KEY;

ALTER TABLE day_ahead_econ_price DROP CONSTRAINT day_ahead_econ_price_pkey;
ALTER TABLE day_ahead_econ_price DROP COLUMN id;
ALTER TABLE day_ahead_econ_price ADD COLUMN id BIGSERIAL PRIMARY KEY;

ALTER TABLE node_spot_price_daily DROP CONSTRAINT node_spot_price_daily_pkey;
ALTER TABLE node_spot_price_daily DROP COLUMN id;
ALTER TABLE node_spot_price_daily ADD COLUMN id BIGSERIAL PRIMARY KEY;

ALTER TABLE spot_settlement_period DROP CONSTRAINT spot_settlement_period_pkey;
ALTER TABLE spot_settlement_period DROP COLUMN id;
ALTER TABLE spot_settlement_period ADD COLUMN id BIGSERIAL PRIMARY KEY;

ALTER TABLE weather_actuals DROP CONSTRAINT weather_actuals_pkey;
ALTER TABLE weather_actuals DROP COLUMN id;
ALTER TABLE weather_actuals ADD COLUMN id BIGSERIAL PRIMARY KEY;

ALTER TABLE weather_forecasts DROP CONSTRAINT weather_forecasts_pkey;
ALTER TABLE weather_forecasts DROP COLUMN id;
ALTER TABLE weather_forecasts ADD COLUMN id BIGSERIAL PRIMARY KEY;

ALTER TABLE real_time_generation DROP CONSTRAINT real_time_generation_pkey;
ALTER TABLE real_time_generation DROP COLUMN id;
ALTER TABLE real_time_generation ADD COLUMN id BIGSERIAL PRIMARY KEY;
