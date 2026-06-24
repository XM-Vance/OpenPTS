-- 第二批 v1 复刻：电网代理价、储能申报。
-- 注：E4 TOU 规则复用 0014 已建的 tou_rules（rule_name / periods jsonb）表。

-- E5 电网代理购电价
CREATE TABLE IF NOT EXISTS grid_agency_price (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    operating_month text NOT NULL,                   -- YYYY-MM
    voltage_level   text NOT NULL,                   -- 380V / 10kV / 35kV / 110kV
    avg_price       double precision NOT NULL,       -- 平均代理价（元/MWh）
    peak_price      double precision NOT NULL,
    flat_price      double precision NOT NULL,
    valley_price    double precision NOT NULL,
    created_at      timestamp NOT NULL DEFAULT now(),
    UNIQUE (operating_month, voltage_level)
);

CREATE INDEX IF NOT EXISTS idx_grid_agency_month
    ON grid_agency_price(operating_month DESC);

-- E6 储能申报策略（每日 96 时段充放电）
CREATE TABLE IF NOT EXISTS storage_declaration (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    station_id      uuid NOT NULL REFERENCES storage_stations(id) ON DELETE CASCADE,
    declared_date   date NOT NULL,
    charge_mw       double precision[] NOT NULL CHECK (cardinality(charge_mw) = 96),
    discharge_mw    double precision[] NOT NULL CHECK (cardinality(discharge_mw) = 96),
    expected_revenue double precision NOT NULL DEFAULT 0,
    strategy_note   text,
    created_at      timestamp NOT NULL DEFAULT now(),
    UNIQUE (station_id, declared_date)
);

CREATE INDEX IF NOT EXISTS idx_storage_declaration_date
    ON storage_declaration(declared_date DESC);
