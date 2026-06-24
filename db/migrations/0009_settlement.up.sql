-- 0009: 结算
-- 覆盖 v1：settlement_daily, spot_settlement_daily, spot_settlement_period,
--         contracts_aggregated_daily, retail_settlement_daily, retail_settlement_prices

-- ─── 批发日结算（核心结算表）──────────────────────
CREATE TABLE IF NOT EXISTS settlement_daily (
    id                      UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    operating_date          DATE NOT NULL,
    version                 VARCHAR(32) NOT NULL DEFAULT 'PRELIMINARY',
                            -- PRELIMINARY / PLATFORM_DAILY / PLATFORM_MONTHLY
    period_details          JSONB NOT NULL,                              -- [{period, contract{}, day_ahead{}, real_time{}}]
    contract_fee            NUMERIC(18,4),
    day_ahead_fee           NUMERIC(18,4),
    real_time_fee           NUMERIC(18,4),
    total_energy_fee        NUMERIC(18,4),
    energy_avg_price        NUMERIC(14,4),
    deviation_recovery_fee  NUMERIC(18,4),
    extra                   JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (operating_date, version)
);

CREATE INDEX IF NOT EXISTS idx_settlement_daily_date ON settlement_daily(operating_date);
CREATE INDEX IF NOT EXISTS idx_settlement_daily_version ON settlement_daily(version);

DROP TRIGGER IF EXISTS settlement_daily_set_updated_at ON settlement_daily;
CREATE TRIGGER settlement_daily_set_updated_at
    BEFORE UPDATE ON settlement_daily
    FOR EACH ROW EXECUTE FUNCTION trg_set_updated_at();

-- ─── 现货日结算 ────────────────────────────────────
CREATE TABLE IF NOT EXISTS spot_settlement_daily (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    settlement_date DATE NOT NULL UNIQUE,
    period_details  JSONB NOT NULL,
    total_volume    NUMERIC(18,4),
    total_fee       NUMERIC(18,4),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_spot_settle_daily_date ON spot_settlement_daily(settlement_date);

DROP TRIGGER IF EXISTS spot_settle_daily_set_updated_at ON spot_settlement_daily;
CREATE TRIGGER spot_settle_daily_set_updated_at
    BEFORE UPDATE ON spot_settlement_daily
    FOR EACH ROW EXECUTE FUNCTION trg_set_updated_at();

-- ─── 现货分时（48 点单条记录便于聚合）─────────────
CREATE TABLE IF NOT EXISTS spot_settlement_period (
    id               BIGSERIAL PRIMARY KEY,
    settlement_date  DATE NOT NULL,
    period           INT NOT NULL,
    volume           NUMERIC(18,4),
    price            NUMERIC(14,4),
    fee              NUMERIC(18,4),
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (settlement_date, period)
);

CREATE INDEX IF NOT EXISTS idx_spot_settle_period_date ON spot_settlement_period(settlement_date);

-- ─── 合同日汇总（衍生表，可考虑物化视图）──────────
CREATE TABLE IF NOT EXISTS contracts_aggregated_daily (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    date                DATE NOT NULL UNIQUE,
    contract_volume     NUMERIC(18,4),
    contract_avg_price  NUMERIC(14,4),
    contract_fee        NUMERIC(18,4),
    mechanism_volume    NUMERIC(18,4),
    extra               JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_contracts_agg_daily_date ON contracts_aggregated_daily(date);

DROP TRIGGER IF EXISTS contracts_agg_daily_set_updated_at ON contracts_aggregated_daily;
CREATE TRIGGER contracts_agg_daily_set_updated_at
    BEFORE UPDATE ON contracts_aggregated_daily
    FOR EACH ROW EXECUTE FUNCTION trg_set_updated_at();

-- ─── 零售日结算 ────────────────────────────────────
CREATE TABLE IF NOT EXISTS retail_settlement_daily (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    customer_id     UUID NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    date            DATE NOT NULL,
    contract_id     UUID REFERENCES retail_contracts(id) ON DELETE SET NULL,
    package_name    VARCHAR(255),                                -- 冗余便于查询
    period_details  JSONB NOT NULL,                              -- [{period_type, load_mwh, unit_price, fee}]
    total_load_mwh  NUMERIC(18,4),
    total_fee       NUMERIC(18,4),
    avg_price       NUMERIC(14,4),
    extra           JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (customer_id, date)
);

CREATE INDEX IF NOT EXISTS idx_retail_settle_daily_date ON retail_settlement_daily(date);
CREATE INDEX IF NOT EXISTS idx_retail_settle_daily_contract ON retail_settlement_daily(contract_id);

DROP TRIGGER IF EXISTS retail_settle_daily_set_updated_at ON retail_settlement_daily;
CREATE TRIGGER retail_settle_daily_set_updated_at
    BEFORE UPDATE ON retail_settlement_daily
    FOR EACH ROW EXECUTE FUNCTION trg_set_updated_at();

-- ─── 零售结算价（衍生）─────────────────────────────
CREATE TABLE IF NOT EXISTS retail_settlement_prices (
    id                    UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    date                  DATE NOT NULL,
    customer_id           UUID NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    final_prices          JSONB NOT NULL,                        -- {period_type: price}
    price_ratio_adjusted  NUMERIC(8,4),
    is_capped             BOOLEAN NOT NULL DEFAULT FALSE,
    nominal_avg_price     NUMERIC(14,4),
    cap_price             NUMERIC(14,4),
    created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (date, customer_id)
);

CREATE INDEX IF NOT EXISTS idx_retail_price_date ON retail_settlement_prices(date);
