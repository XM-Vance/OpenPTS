-- 0011: 调频与储能
-- 覆盖 v1：frequency_regulation_clearing, frequency_regulation_demand, freq_comp_fee

-- ─── 调频清算 ──────────────────────────────────────
CREATE TABLE IF NOT EXISTS frequency_regulation_clearing (
    id               UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    settlement_date  DATE NOT NULL,
    regulation_type  VARCHAR(32) NOT NULL,                    -- AGC / AVC / 其他
    cleared_volume   NUMERIC(18,4),
    cleared_price    NUMERIC(14,4),
    revenue          NUMERIC(18,4),
    extra            JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (settlement_date, regulation_type)
);

CREATE INDEX IF NOT EXISTS idx_freq_clearing_date ON frequency_regulation_clearing(settlement_date);

-- ─── 调频需求 ──────────────────────────────────────
CREATE TABLE IF NOT EXISTS frequency_regulation_demand (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    settlement_date DATE NOT NULL UNIQUE,
    demand_volume   NUMERIC(18,4),
    demand_price    NUMERIC(14,4),
    source          VARCHAR(64),
    extra           JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_freq_demand_date ON frequency_regulation_demand(settlement_date);

-- ─── 调频补偿费 ────────────────────────────────────
CREATE TABLE IF NOT EXISTS freq_comp_fee (
    id                UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    settlement_date   DATE NOT NULL,
    frequency_type    VARCHAR(32) NOT NULL,
    compensation_fee  NUMERIC(18,4),
    extra             JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (settlement_date, frequency_type)
);

CREATE INDEX IF NOT EXISTS idx_freq_comp_date ON freq_comp_fee(settlement_date);
