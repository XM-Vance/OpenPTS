-- 0008: 预测结果
-- 覆盖 v1：price_forecast_results, medium_term_load_forecast

-- ─── 价格预测结果（含准确率回溯）───────────────────
CREATE TABLE IF NOT EXISTS price_forecast_results (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    forecast_date       DATE NOT NULL,                       -- 做预测的那一天
    target_date         DATE NOT NULL,                       -- 被预测的目标日
    price_da_forecast   NUMERIC(14,4)[],                     -- 48 点日前价预测
    price_rt_forecast   NUMERIC(14,4)[],                     -- 48 点实时价预测
    forecast_method     VARCHAR(64),                         -- 算法标识
    accuracy_metrics    JSONB NOT NULL DEFAULT '{}'::jsonb,  -- {mape, rmse, mae, ...}
    operator            VARCHAR(64),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (forecast_date, target_date, forecast_method)
);

CREATE INDEX IF NOT EXISTS idx_price_forecast_target ON price_forecast_results(target_date);
CREATE INDEX IF NOT EXISTS idx_price_forecast_made ON price_forecast_results(forecast_date);

-- ─── 中期负荷预测（30 ~ 90 天）─────────────────────
CREATE TABLE IF NOT EXISTS medium_term_load_forecast (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    forecast_date       DATE NOT NULL,
    daily_forecasts     JSONB NOT NULL,                      -- [{target_date, total_load, curve[], actual_load, wmape}]
    monthly_forecasts   JSONB NOT NULL DEFAULT '[]'::jsonb,
    operator            VARCHAR(64),
    method              VARCHAR(64),
    extra               JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_mt_forecast_date ON medium_term_load_forecast(forecast_date);

DROP TRIGGER IF EXISTS mt_forecast_set_updated_at ON medium_term_load_forecast;
CREATE TRIGGER mt_forecast_set_updated_at
    BEFORE UPDATE ON medium_term_load_forecast
    FOR EACH ROW EXECUTE FUNCTION trg_set_updated_at();
