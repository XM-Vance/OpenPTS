-- 0007: 价格数据（现货、日前、国网代购、节点）
-- 覆盖 v1：real_time_spot_price, day_ahead_spot_price, day_ahead_econ_price,
--         price_sgcc, node_spot_price_daily

-- ─── 实时现货价 ────────────────────────────────────
CREATE TABLE IF NOT EXISTS real_time_spot_price (
    id               BIGSERIAL PRIMARY KEY,
    date             DATE NOT NULL,
    period           INT NOT NULL,                            -- 1 ~ 96 (15min) 或 1 ~ 48 (30min)
    price_rt         NUMERIC(14,4),                           -- 实时价
    node_price_rt    NUMERIC(14,4),                           -- 节点实时价
    price_da         NUMERIC(14,4),                           -- 日前价（冗余便于联合查询）
    price_da_econ    NUMERIC(14,4),                           -- 日前经济价（冗余）
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (date, period)
);

CREATE INDEX IF NOT EXISTS idx_rt_price_date ON real_time_spot_price(date);

-- ─── 日前现货价 ────────────────────────────────────
CREATE TABLE IF NOT EXISTS day_ahead_spot_price (
    id                 BIGSERIAL PRIMARY KEY,
    date               DATE NOT NULL,
    period             INT NOT NULL,
    price_da           NUMERIC(14,4),
    price_da_econ      NUMERIC(14,4),
    price_da_forecast  NUMERIC(14,4),
    created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (date, period)
);

CREATE INDEX IF NOT EXISTS idx_da_price_date ON day_ahead_spot_price(date);

-- ─── 日前经济价（独立表）───────────────────────────
CREATE TABLE IF NOT EXISTS day_ahead_econ_price (
    id          BIGSERIAL PRIMARY KEY,
    date        DATE NOT NULL,
    period      INT NOT NULL,
    price_econ  NUMERIC(14,4) NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (date, period)
);

CREATE INDEX IF NOT EXISTS idx_da_econ_date ON day_ahead_econ_price(date);

-- ─── 国网代购价（月级）────────────────────────────
CREATE TABLE IF NOT EXISTS price_sgcc (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    month           VARCHAR(7) NOT NULL UNIQUE,               -- YYYY-MM
    price_sgcc      NUMERIC(14,4) NOT NULL,
    effective_from  DATE NOT NULL,
    effective_to    DATE,
    source          VARCHAR(64),                              -- 文件来源/导入批次
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_price_sgcc_effective ON price_sgcc(effective_from);

DROP TRIGGER IF EXISTS price_sgcc_set_updated_at ON price_sgcc;
CREATE TRIGGER price_sgcc_set_updated_at
    BEFORE UPDATE ON price_sgcc
    FOR EACH ROW EXECUTE FUNCTION trg_set_updated_at();

-- ─── 节点现货价 ────────────────────────────────────
CREATE TABLE IF NOT EXISTS node_spot_price_daily (
    id          BIGSERIAL PRIMARY KEY,
    date        DATE NOT NULL,
    period      INT NOT NULL,
    node_id     VARCHAR(64) NOT NULL,
    price       NUMERIC(14,4) NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (date, period, node_id)
);

CREATE INDEX IF NOT EXISTS idx_node_price_date ON node_spot_price_daily(date);
CREATE INDEX IF NOT EXISTS idx_node_price_node ON node_spot_price_daily(node_id);
