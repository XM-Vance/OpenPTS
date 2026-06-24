-- 0006: 处理后的负荷数据（汇聚、规范化、月度聚合）
-- 覆盖 v1：unified_load_curve, user_load_data, customer_monthly_energy,
--         intent_customer_load_curve_daily, intent_customer_meter_reads_daily

-- ─── 统一负荷曲线（汇聚多个计量点）─────────────────
CREATE TABLE IF NOT EXISTS unified_load_curve (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    customer_id         UUID NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    date                DATE NOT NULL,
    curve_data          JSONB NOT NULL,                       -- [{period, value}] 96 点或 48 点
    aggregation_method  VARCHAR(32) NOT NULL DEFAULT 'sum',   -- sum / weighted_avg
    total_daily_load    NUMERIC(18,4),
    data_quality        VARCHAR(32) NOT NULL DEFAULT 'normal', -- normal / missing / abnormal
    extra               JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (customer_id, date)
);

CREATE INDEX IF NOT EXISTS idx_unified_load_date ON unified_load_curve(date);
CREATE INDEX IF NOT EXISTS idx_unified_load_quality ON unified_load_curve(data_quality);

DROP TRIGGER IF EXISTS unified_load_set_updated_at ON unified_load_curve;
CREATE TRIGGER unified_load_set_updated_at
    BEFORE UPDATE ON unified_load_curve
    FOR EACH ROW EXECUTE FUNCTION trg_set_updated_at();

-- ─── 用户负荷数据（96 点规范化）────────────────────
CREATE TABLE IF NOT EXISTS user_load_data (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    customer_id     UUID NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    date            DATE NOT NULL,
    curve_96        NUMERIC(14,4)[] NOT NULL,                -- 96 点数组（15min/点）
    total_load      NUMERIC(18,4),
    quality_flag    VARCHAR(32) NOT NULL DEFAULT 'ok',       -- ok / missing / outlier
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (customer_id, date)
);

CREATE INDEX IF NOT EXISTS idx_user_load_date ON user_load_data(date);
CREATE INDEX IF NOT EXISTS idx_user_load_quality ON user_load_data(quality_flag);

DROP TRIGGER IF EXISTS user_load_set_updated_at ON user_load_data;
CREATE TRIGGER user_load_set_updated_at
    BEFORE UPDATE ON user_load_data
    FOR EACH ROW EXECUTE FUNCTION trg_set_updated_at();

-- ─── 客户月度电量汇总 ──────────────────────────────
CREATE TABLE IF NOT EXISTS customer_monthly_energy (
    id                UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    customer_id       UUID NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    month             VARCHAR(7) NOT NULL,                   -- YYYY-MM
    monthly_energy    NUMERIC(18,4) NOT NULL,
    avg_daily_energy  NUMERIC(18,4),
    variation_cv      NUMERIC(8,4),                           -- 变异系数
    extra             JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (customer_id, month)
);

CREATE INDEX IF NOT EXISTS idx_customer_monthly_energy_month ON customer_monthly_energy(month);

DROP TRIGGER IF EXISTS customer_monthly_energy_set_updated_at ON customer_monthly_energy;
CREATE TRIGGER customer_monthly_energy_set_updated_at
    BEFORE UPDATE ON customer_monthly_energy
    FOR EACH ROW EXECUTE FUNCTION trg_set_updated_at();

-- ─── 意向客户日负荷曲线 ────────────────────────────
CREATE TABLE IF NOT EXISTS intent_customer_load_curve_daily (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    intent_id   UUID NOT NULL REFERENCES intent_customers(id) ON DELETE CASCADE,
    date        DATE NOT NULL,
    curve_48    NUMERIC(14,4)[] NOT NULL,                    -- 48 点（30min/点）
    total_load  NUMERIC(18,4),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (intent_id, date)
);

CREATE INDEX IF NOT EXISTS idx_intent_load_curve_date ON intent_customer_load_curve_daily(date);

-- ─── 意向客户原表数据 ──────────────────────────────
CREATE TABLE IF NOT EXISTS intent_customer_meter_reads_daily (
    id            UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    intent_id     UUID NOT NULL REFERENCES intent_customers(id) ON DELETE CASCADE,
    date          DATE NOT NULL,
    meter_reads   JSONB NOT NULL,                            -- {meter_id: [period_data]}
    total_load    NUMERIC(18,4),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (intent_id, date)
);

CREATE INDEX IF NOT EXISTS idx_intent_meter_reads_date ON intent_customer_meter_reads_daily(date);
