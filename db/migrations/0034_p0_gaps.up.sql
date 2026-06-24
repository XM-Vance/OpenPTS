-- P0 v1 缺口补齐：零售月度结算、预结算明细、预测基础数据。

-- U1 零售月度结算（合同维度）
CREATE TABLE IF NOT EXISTS retail_monthly_settlement (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    contract_id     uuid NOT NULL REFERENCES retail_contracts(id) ON DELETE CASCADE,
    operating_month text NOT NULL,                       -- YYYY-MM
    contract_energy_mwh    double precision NOT NULL DEFAULT 0,  -- 合同应购
    actual_energy_mwh      double precision NOT NULL DEFAULT 0,  -- 实际用电
    weighted_avg_price     double precision NOT NULL DEFAULT 0,  -- 加权均价
    receivable_amount      double precision NOT NULL DEFAULT 0,  -- 应收
    actual_amount          double precision NOT NULL DEFAULT 0,  -- 实收
    deviation_energy_mwh   double precision NOT NULL DEFAULT 0,  -- 偏差电量（actual - contract）
    penalty_amount         double precision NOT NULL DEFAULT 0,  -- 违约金
    note            text,
    created_at      timestamp NOT NULL DEFAULT now(),
    UNIQUE (contract_id, operating_month)
);

CREATE INDEX IF NOT EXISTS idx_retail_monthly_settle_month
    ON retail_monthly_settlement(operating_month DESC, contract_id);

-- U2 预结算明细（日维度）
CREATE TABLE IF NOT EXISTS pre_settlement_daily (
    id                uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    operating_date    date NOT NULL UNIQUE,
    declared_curve_96 double precision[] NOT NULL CHECK (cardinality(declared_curve_96) = 96),
    cleared_curve_96  double precision[] NOT NULL CHECK (cardinality(cleared_curve_96) = 96),
    spot_price_96     double precision[] NOT NULL CHECK (cardinality(spot_price_96) = 96),
    total_declared    double precision NOT NULL DEFAULT 0, -- MWh
    total_cleared     double precision NOT NULL DEFAULT 0,
    total_deviation   double precision NOT NULL DEFAULT 0,
    deviation_ratio   double precision NOT NULL DEFAULT 0, -- (cleared-declared)/declared
    energy_revenue    double precision NOT NULL DEFAULT 0,
    deviation_penalty double precision NOT NULL DEFAULT 0,
    final_amount      double precision NOT NULL DEFAULT 0,
    created_at        timestamp NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_pre_settle_date
    ON pre_settlement_daily(operating_date DESC);

-- U3 预测基础数据：节假日 + 典型日曲线
CREATE TABLE IF NOT EXISTS holidays (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    holiday_date date NOT NULL UNIQUE,
    name        text NOT NULL,
    kind        text NOT NULL DEFAULT 'public', -- public / makeup_workday
    note        text,
    created_at  timestamp NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_holidays_date ON holidays(holiday_date);

CREATE TABLE IF NOT EXISTS typical_curves (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name        text NOT NULL UNIQUE,
    season      text NOT NULL,                  -- spring / summer / autumn / winter
    day_type    text NOT NULL,                  -- workday / weekend / holiday
    curve_96    double precision[] NOT NULL CHECK (cardinality(curve_96) = 96),
    note        text,
    enabled     boolean NOT NULL DEFAULT true,
    created_at  timestamp NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_typical_curves_season_daytype
    ON typical_curves(season, day_type);
