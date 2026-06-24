-- 气象数据：日维度，关联到地点。

CREATE TABLE IF NOT EXISTS weather_data (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    obs_date    date NOT NULL,
    location    text NOT NULL,
    temp_high   double precision,
    temp_low    double precision,
    humidity    double precision,           -- 0-100
    precip_mm   double precision,
    wind_kmh    double precision,
    load_factor double precision,           -- 0-1，气温对负荷的影响系数（经验值）
    description text,                       -- 天气描述：晴 / 多云 / 阵雨
    created_at  timestamp NOT NULL DEFAULT now(),
    UNIQUE (obs_date, location)
);

CREATE INDEX IF NOT EXISTS idx_weather_date
    ON weather_data(obs_date DESC);
