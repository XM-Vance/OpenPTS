-- 0014: 杂项业务表
-- 覆盖 v1：system_alerts, real_time_generation,
--         intent_customer_monthly_retail_simulation, intent_customer_monthly_wholesale,
--         rolling_match_snapshots, actual_operation

-- ─── 系统告警 ──────────────────────────────────────
CREATE TABLE IF NOT EXISTS system_alerts (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    alert_type          VARCHAR(64) NOT NULL,
    severity            VARCHAR(16) NOT NULL,                    -- info / warn / critical
    message             TEXT NOT NULL,
    affected_resources  JSONB NOT NULL DEFAULT '[]'::jsonb,
    resolved            BOOLEAN NOT NULL DEFAULT FALSE,
    resolved_by         UUID REFERENCES users(id),
    resolved_at         TIMESTAMPTZ,
    note                TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_system_alert_type ON system_alerts(alert_type);
CREATE INDEX IF NOT EXISTS idx_system_alert_severity ON system_alerts(severity);
CREATE INDEX IF NOT EXISTS idx_system_alert_unresolved ON system_alerts(resolved) WHERE resolved = FALSE;
CREATE INDEX IF NOT EXISTS idx_system_alert_created ON system_alerts(created_at);

DROP TRIGGER IF EXISTS system_alerts_set_updated_at ON system_alerts;
CREATE TRIGGER system_alerts_set_updated_at
    BEFORE UPDATE ON system_alerts
    FOR EACH ROW EXECUTE FUNCTION trg_set_updated_at();

-- ─── 实时发电 ──────────────────────────────────────
CREATE TABLE IF NOT EXISTS real_time_generation (
    id              BIGSERIAL PRIMARY KEY,
    date            DATE NOT NULL,
    period          INT NOT NULL,
    generation_mwh  NUMERIC(18,4),
    source_type     VARCHAR(64),                              -- 火电 / 水电 / 风电 / 光伏
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (date, period, source_type)
);

CREATE INDEX IF NOT EXISTS idx_rt_generation_date ON real_time_generation(date);

-- ─── 意向客户月度零售模拟 ──────────────────────────
CREATE TABLE IF NOT EXISTS intent_customer_monthly_retail_simulation (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    intent_id       UUID NOT NULL REFERENCES intent_customers(id) ON DELETE CASCADE,
    month           VARCHAR(7) NOT NULL,
    total_energy    NUMERIC(18,4),
    estimated_cost  NUMERIC(18,4),
    package_id      UUID REFERENCES retail_packages(id) ON DELETE SET NULL,
    details         JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (intent_id, month, package_id)
);

CREATE INDEX IF NOT EXISTS idx_intent_monthly_retail_month ON intent_customer_monthly_retail_simulation(month);

-- ─── 意向客户月度批发 ──────────────────────────────
CREATE TABLE IF NOT EXISTS intent_customer_monthly_wholesale (
    id                UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    intent_id         UUID NOT NULL REFERENCES intent_customers(id) ON DELETE CASCADE,
    month             VARCHAR(7) NOT NULL,
    wholesale_volume  NUMERIC(18,4),
    wholesale_cost    NUMERIC(18,4),
    details           JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (intent_id, month)
);

-- ─── 月内行情快照（交易策略）──────────────────────
CREATE TABLE IF NOT EXISTS rolling_match_snapshots (
    id             UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    trade_date     DATE NOT NULL,
    delivery_date  DATE NOT NULL,
    record_count   INT NOT NULL DEFAULT 0,
    snapshot_data  JSONB NOT NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (trade_date, delivery_date)
);

CREATE INDEX IF NOT EXISTS idx_rolling_match_trade ON rolling_match_snapshots(trade_date);

-- ─── 实际操作记录（交易成交）──────────────────────
CREATE TABLE IF NOT EXISTS actual_operation (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    trade_date      DATE NOT NULL,
    delivery_date   DATE NOT NULL,
    operation_type  VARCHAR(64) NOT NULL,
    volume_mwh      NUMERIC(18,4),
    price           NUMERIC(14,4),
    profit_loss     NUMERIC(18,4),
    operator        VARCHAR(64),
    extra           JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_actual_op_trade ON actual_operation(trade_date);
CREATE INDEX IF NOT EXISTS idx_actual_op_delivery ON actual_operation(delivery_date);
