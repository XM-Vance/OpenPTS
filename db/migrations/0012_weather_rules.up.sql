-- 0012: 气象数据 + 分时电价规则 + 机制电量
-- 覆盖 v1：weather_locations, weather_actuals, weather_forecasts, tou_rules,
--         mechanism_energy_monthly

-- ─── 气象区域 ──────────────────────────────────────
CREATE TABLE IF NOT EXISTS weather_locations (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name        VARCHAR(128) NOT NULL UNIQUE,            -- 被 customers.location 引用
    province    VARCHAR(64),
    city        VARCHAR(64),
    latitude    NUMERIC(10,6),
    longitude   NUMERIC(10,6),
    gis         JSONB,                                    -- {type, coordinates}
    extra       JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

DROP TRIGGER IF EXISTS weather_locations_set_updated_at ON weather_locations;
CREATE TRIGGER weather_locations_set_updated_at
    BEFORE UPDATE ON weather_locations
    FOR EACH ROW EXECUTE FUNCTION trg_set_updated_at();

-- ─── 实测气象 ──────────────────────────────────────
CREATE TABLE IF NOT EXISTS weather_actuals (
    id              BIGSERIAL PRIMARY KEY,
    location_name   VARCHAR(128) NOT NULL REFERENCES weather_locations(name) ON UPDATE CASCADE,
    date            DATE NOT NULL,
    avg_temp        NUMERIC(6,2),
    max_temp        NUMERIC(6,2),
    min_temp        NUMERIC(6,2),
    humidity        NUMERIC(5,2),
    wind_speed      NUMERIC(6,2),
    precipitation   NUMERIC(6,2),
    extra           JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (location_name, date)
);

CREATE INDEX IF NOT EXISTS idx_weather_actual_date ON weather_actuals(date);

-- ─── 气象预报 ──────────────────────────────────────
CREATE TABLE IF NOT EXISTS weather_forecasts (
    id                 BIGSERIAL PRIMARY KEY,
    location_name      VARCHAR(128) NOT NULL REFERENCES weather_locations(name) ON UPDATE CASCADE,
    forecast_date      DATE NOT NULL,                    -- 做预报那天
    target_date        DATE NOT NULL,                    -- 被预报那天
    temp_forecast      NUMERIC(6,2),
    humidity_forecast  NUMERIC(5,2),
    wind_forecast      NUMERIC(6,2),
    extra              JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (location_name, forecast_date, target_date)
);

CREATE INDEX IF NOT EXISTS idx_weather_forecast_target ON weather_forecasts(target_date);

-- ─── 分时电价规则 ──────────────────────────────────
CREATE TABLE IF NOT EXISTS tou_rules (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    rule_name       VARCHAR(128) NOT NULL,
    effective_from  DATE NOT NULL,
    effective_to    DATE,
    periods         JSONB NOT NULL,                       -- [{period, tou_type, start_time, end_time}]
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_tou_rules_effective ON tou_rules(effective_from);

DROP TRIGGER IF EXISTS tou_rules_set_updated_at ON tou_rules;
CREATE TRIGGER tou_rules_set_updated_at
    BEFORE UPDATE ON tou_rules
    FOR EACH ROW EXECUTE FUNCTION trg_set_updated_at();

-- ─── 机制电量月度 ──────────────────────────────────
CREATE TABLE IF NOT EXISTS mechanism_energy_monthly (
    id                       UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    month                    VARCHAR(7) NOT NULL UNIQUE,   -- YYYY-MM
    total_mechanism_energy   NUMERIC(18,4),
    by_customer              JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at               TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at               TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

DROP TRIGGER IF EXISTS mechanism_energy_set_updated_at ON mechanism_energy_monthly;
CREATE TRIGGER mechanism_energy_set_updated_at
    BEFORE UPDATE ON mechanism_energy_monthly
    FOR EACH ROW EXECUTE FUNCTION trg_set_updated_at();
