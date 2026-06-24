-- 0045: 将剩余 BIGSERIAL 主键统一为 UUID（与全库其它表一致）
-- 覆盖 0007、0009、0012、0014 中仍使用 BIGSERIAL 的表
-- 注意：这些表中 id 列无外键引用，可安全执行 DROP/ADD

-- ─── 1. real_time_spot_price — id BIGSERIAL → UUID ────────
ALTER TABLE real_time_spot_price DROP CONSTRAINT real_time_spot_price_pkey;
ALTER TABLE real_time_spot_price DROP COLUMN id;
ALTER TABLE real_time_spot_price ADD COLUMN id UUID PRIMARY KEY DEFAULT gen_random_uuid();
-- idx_rt_price_date 不含 id，无需改

-- ─── 2. day_ahead_spot_price — id BIGSERIAL → UUID ───────
ALTER TABLE day_ahead_spot_price DROP CONSTRAINT day_ahead_spot_price_pkey;
ALTER TABLE day_ahead_spot_price DROP COLUMN id;
ALTER TABLE day_ahead_spot_price ADD COLUMN id UUID PRIMARY KEY DEFAULT gen_random_uuid();
-- idx_da_price_date 不含 id，无需改

-- ─── 3. day_ahead_econ_price — id BIGSERIAL → UUID ───────
ALTER TABLE day_ahead_econ_price DROP CONSTRAINT day_ahead_econ_price_pkey;
ALTER TABLE day_ahead_econ_price DROP COLUMN id;
ALTER TABLE day_ahead_econ_price ADD COLUMN id UUID PRIMARY KEY DEFAULT gen_random_uuid();
-- idx_da_econ_date 不含 id，无需改

-- ─── 4. node_spot_price_daily — id BIGSERIAL → UUID ──────
ALTER TABLE node_spot_price_daily DROP CONSTRAINT node_spot_price_daily_pkey;
ALTER TABLE node_spot_price_daily DROP COLUMN id;
ALTER TABLE node_spot_price_daily ADD COLUMN id UUID PRIMARY KEY DEFAULT gen_random_uuid();
-- idx_node_price_date / idx_node_price_node 不含 id，无需改

-- ─── 5. spot_settlement_period — id BIGSERIAL → UUID ─────
ALTER TABLE spot_settlement_period DROP CONSTRAINT spot_settlement_period_pkey;
ALTER TABLE spot_settlement_period DROP COLUMN id;
ALTER TABLE spot_settlement_period ADD COLUMN id UUID PRIMARY KEY DEFAULT gen_random_uuid();
-- idx_spot_settle_period_date 不含 id，无需改

-- ─── 6. weather_actuals — id BIGSERIAL → UUID ────────────
ALTER TABLE weather_actuals DROP CONSTRAINT weather_actuals_pkey;
ALTER TABLE weather_actuals DROP COLUMN id;
ALTER TABLE weather_actuals ADD COLUMN id UUID PRIMARY KEY DEFAULT gen_random_uuid();
-- idx_weather_actual_date 不含 id，无需改

-- ─── 7. weather_forecasts — id BIGSERIAL → UUID ──────────
ALTER TABLE weather_forecasts DROP CONSTRAINT weather_forecasts_pkey;
ALTER TABLE weather_forecasts DROP COLUMN id;
ALTER TABLE weather_forecasts ADD COLUMN id UUID PRIMARY KEY DEFAULT gen_random_uuid();
-- idx_weather_forecast_target 不含 id，无需改

-- ─── 8. real_time_generation — id BIGSERIAL → UUID ──────
ALTER TABLE real_time_generation DROP CONSTRAINT real_time_generation_pkey;
ALTER TABLE real_time_generation DROP COLUMN id;
ALTER TABLE real_time_generation ADD COLUMN id UUID PRIMARY KEY DEFAULT gen_random_uuid();
-- idx_rt_generation_date 不含 id，无需改
