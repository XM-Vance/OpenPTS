-- 0041: 光伏系列模块
-- 光伏电站表
CREATE TABLE IF NOT EXISTS solar_stations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    station_name VARCHAR(200) NOT NULL,
    location VARCHAR(200) NOT NULL DEFAULT '',
    capacity_kw DOUBLE PRECISION NOT NULL DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    installed_date DATE,
    latitude DOUBLE PRECISION,
    longitude DOUBLE PRECISION,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_solar_stations_status ON solar_stations (status);

-- 光伏发电预测表
CREATE TABLE IF NOT EXISTS solar_generation_forecast (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    station_id UUID NOT NULL REFERENCES solar_stations(id) ON DELETE CASCADE,
    forecast_date DATE NOT NULL,
    period INT NOT NULL CHECK (period >= 1 AND period <= 96),
    forecast_power_kw DOUBLE PRECISION NOT NULL DEFAULT 0,
    actual_power_kw DOUBLE PRECISION,
    deviation_rate DOUBLE PRECISION,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_solar_forecast_station_date ON solar_generation_forecast (station_id, forecast_date);
CREATE UNIQUE INDEX IF NOT EXISTS idx_solar_forecast_uniq ON solar_generation_forecast (station_id, forecast_date, period);

-- 光伏收益结算表
CREATE TABLE IF NOT EXISTS solar_revenue_settlement (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    station_id UUID NOT NULL REFERENCES solar_stations(id) ON DELETE CASCADE,
    settlement_month VARCHAR(7) NOT NULL,
    energy_kwh DOUBLE PRECISION NOT NULL DEFAULT 0,
    revenue DOUBLE PRECISION NOT NULL DEFAULT 0,
    avg_price DOUBLE PRECISION NOT NULL DEFAULT 0,
    subsidy DOUBLE PRECISION NOT NULL DEFAULT 0,
    net_income DOUBLE PRECISION NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_solar_revenue_station_month ON solar_revenue_settlement (station_id, settlement_month);
CREATE UNIQUE INDEX IF NOT EXISTS idx_solar_revenue_uniq ON solar_revenue_settlement (station_id, settlement_month);
