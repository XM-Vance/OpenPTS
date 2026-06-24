-- P1 v1 缺口补齐：系统总负荷、中期预测、预测准确率、机制电量、市场分析。

-- V1 系统总负荷（日维度 96 点）
CREATE TABLE IF NOT EXISTS total_load_daily (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    data_date   date NOT NULL UNIQUE,
    region      text NOT NULL DEFAULT 'system',
    curve_96    double precision[] NOT NULL CHECK (cardinality(curve_96) = 96),
    peak_load   double precision NOT NULL DEFAULT 0,    -- MW
    valley_load double precision NOT NULL DEFAULT 0,
    avg_load    double precision NOT NULL DEFAULT 0,
    total_mwh   double precision NOT NULL DEFAULT 0,
    created_at  timestamp NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_total_load_date
    ON total_load_daily(data_date DESC);

-- V2 中期负荷预测（月维度）。
-- 注：v1 老库已有 medium_term_load_forecast（日维度 jsonb），v2 用月维度新表。
CREATE TABLE IF NOT EXISTS medium_load_forecast (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    forecast_month  text NOT NULL UNIQUE,         -- YYYY-MM
    predicted_mwh   double precision NOT NULL,
    actual_mwh      double precision,             -- 月底实测填入
    peak_mw         double precision NOT NULL DEFAULT 0,
    growth_rate     double precision,             -- 同比 %
    weather_factor  double precision,             -- 0.85-1.15
    economic_factor double precision,             -- 0.9-1.1
    note            text,
    created_at      timestamp NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_mlf_month
    ON medium_load_forecast(forecast_month DESC);

-- V3 预测准确率（按预测目标维度，与 v1 accuracy_service 对齐）
CREATE TABLE IF NOT EXISTS forecast_accuracy (
    id                 uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    forecast_target    text NOT NULL,             -- load / price / freq
    forecast_date      date NOT NULL,
    predicted_value    double precision NOT NULL,
    actual_value       double precision,
    mape               double precision,          -- %
    rmse               double precision,
    model_version      text,
    created_at         timestamp NOT NULL DEFAULT now(),
    UNIQUE (forecast_target, forecast_date)
);

CREATE INDEX IF NOT EXISTS idx_fa_target_date
    ON forecast_accuracy(forecast_target, forecast_date DESC);

-- V4 机制电量（中长期电力合约）。注：v1 老库已有 mechanism_energy_monthly，用 v2 新名 mechanism_energy_plan
CREATE TABLE IF NOT EXISTS mechanism_energy_plan (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    operating_month text NOT NULL,
    voltage_level   text NOT NULL,
    planned_mwh     double precision NOT NULL,
    actual_mwh      double precision NOT NULL DEFAULT 0,
    deviation_mwh   double precision NOT NULL DEFAULT 0,
    contract_price  double precision NOT NULL,
    settle_amount   double precision NOT NULL DEFAULT 0,
    created_at      timestamp NOT NULL DEFAULT now(),
    UNIQUE (operating_month, voltage_level)
);

CREATE INDEX IF NOT EXISTS idx_mech_energy_month
    ON mechanism_energy_plan(operating_month DESC);

-- V5 市场分析（现货市场每日宏观指标）
CREATE TABLE IF NOT EXISTS market_analysis_daily (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    trade_date      date NOT NULL UNIQUE,
    high_price      double precision NOT NULL,
    low_price       double precision NOT NULL,
    avg_price       double precision NOT NULL,
    volatility      double precision NOT NULL,    -- 波动率（标准差/均值）
    total_volume_mwh double precision NOT NULL,
    peak_valley_gap double precision NOT NULL,    -- 峰谷差
    market_sentiment text,                        -- bullish / neutral / bearish
    created_at      timestamp NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_market_analysis_date
    ON market_analysis_daily(trade_date DESC);
