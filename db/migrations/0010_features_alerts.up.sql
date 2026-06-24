-- 0010: 客户特征 + 异动告警 + 分析历史
-- 覆盖 v1：customer_characteristics, customer_anomaly_alerts, analysis_history_log

-- ─── 客户特征画像（月度快照）───────────────────────
CREATE TABLE IF NOT EXISTS customer_characteristics (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    customer_id         UUID NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    data_date           DATE NOT NULL,
    long_term           JSONB NOT NULL DEFAULT '{}'::jsonb,    -- 长期特征指标
    short_term          JSONB NOT NULL DEFAULT '{}'::jsonb,    -- 短期特征指标
    tags                TEXT[] NOT NULL DEFAULT '{}',
    regularity_score    NUMERIC(8,4),
    quality_rating      VARCHAR(32),
    baseline_curve      NUMERIC(14,4)[],                       -- 典型日基线曲线
    extra               JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (customer_id, data_date)
);

CREATE INDEX IF NOT EXISTS idx_cust_char_date ON customer_characteristics(data_date);
CREATE INDEX IF NOT EXISTS idx_cust_char_tags ON customer_characteristics USING GIN(tags);

DROP TRIGGER IF EXISTS cust_char_set_updated_at ON customer_characteristics;
CREATE TRIGGER cust_char_set_updated_at
    BEFORE UPDATE ON customer_characteristics
    FOR EACH ROW EXECUTE FUNCTION trg_set_updated_at();

-- ─── 客户异动告警 ──────────────────────────────────
CREATE TABLE IF NOT EXISTS customer_anomaly_alerts (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    customer_id         UUID NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    alert_date          DATE NOT NULL,
    alert_type          VARCHAR(64) NOT NULL,                  -- load_drop / shape_change / quality_drop / ...
    severity            VARCHAR(16) NOT NULL,                  -- info / warn / critical
    confidence          NUMERIC(5,2),                           -- 0.00 ~ 100.00
    reason              TEXT,
    metrics             JSONB NOT NULL DEFAULT '{}'::jsonb,
    rule_id             VARCHAR(64),
    acknowledged        BOOLEAN NOT NULL DEFAULT FALSE,
    acknowledged_by     UUID REFERENCES users(id),
    acknowledged_at     TIMESTAMPTZ,
    note                TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_cust_alert_customer ON customer_anomaly_alerts(customer_id);
CREATE INDEX IF NOT EXISTS idx_cust_alert_date ON customer_anomaly_alerts(alert_date);
CREATE INDEX IF NOT EXISTS idx_cust_alert_type ON customer_anomaly_alerts(alert_type);
CREATE INDEX IF NOT EXISTS idx_cust_alert_unack ON customer_anomaly_alerts(acknowledged) WHERE acknowledged = FALSE;

DROP TRIGGER IF EXISTS cust_alert_set_updated_at ON customer_anomaly_alerts;
CREATE TRIGGER cust_alert_set_updated_at
    BEFORE UPDATE ON customer_anomaly_alerts
    FOR EACH ROW EXECUTE FUNCTION trg_set_updated_at();

-- ─── 分析历史 ──────────────────────────────────────
CREATE TABLE IF NOT EXISTS analysis_history_log (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    customer_id     UUID REFERENCES customers(id) ON DELETE SET NULL,
    date            DATE NOT NULL,
    tags            TEXT[] NOT NULL DEFAULT '{}',
    rule_ids        TEXT[] NOT NULL DEFAULT '{}',
    metrics         JSONB NOT NULL DEFAULT '{}'::jsonb,
    execution_time  NUMERIC(8,3),                              -- 秒
    operator        VARCHAR(64),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_analysis_log_customer ON analysis_history_log(customer_id);
CREATE INDEX IF NOT EXISTS idx_analysis_log_date ON analysis_history_log(date);
