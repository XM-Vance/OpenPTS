-- 0005: 原始负荷数据（高频写入，按月分区）
-- 覆盖 v1：raw_meter_data, raw_mp_data

-- ─── 原始电表数据 ──────────────────────────────────
-- 主键含分区键 date；按月分区，2026 年 12 个月预建分区
CREATE TABLE IF NOT EXISTS raw_meter_data (
    id           BIGSERIAL,
    meter_id     VARCHAR(64) NOT NULL,
    multiplier   NUMERIC(10,4) NOT NULL DEFAULT 1,
    date         DATE NOT NULL,
    period_data  JSONB NOT NULL,                          -- [{period, value}, ...]
    data_source  VARCHAR(32) NOT NULL DEFAULT 'platform', -- platform / manual / import
    imported_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (date, id)
) PARTITION BY RANGE (date);

CREATE INDEX IF NOT EXISTS idx_raw_meter_data_meter_date ON raw_meter_data(meter_id, date);
CREATE INDEX IF NOT EXISTS idx_raw_meter_data_imported ON raw_meter_data(imported_at);

DO $$
DECLARE
    yr INT := EXTRACT(YEAR FROM CURRENT_DATE)::INT;
    mo INT;
    start_date DATE;
    end_date DATE;
    partition_name TEXT;
BEGIN
    FOR mo IN 1..12 LOOP
        start_date := make_date(yr, mo, 1);
        end_date := start_date + INTERVAL '1 month';
        partition_name := format('raw_meter_data_y%sm%s', yr, lpad(mo::text, 2, '0'));
        EXECUTE format(
            'CREATE TABLE IF NOT EXISTS %I PARTITION OF raw_meter_data FOR VALUES FROM (%L) TO (%L)',
            partition_name, start_date, end_date
        );
    END LOOP;
END $$;

-- ─── 原始计量点数据 ────────────────────────────────
CREATE TABLE IF NOT EXISTS raw_mp_data (
    id           BIGSERIAL,
    mp_id        VARCHAR(64) NOT NULL,
    customer_id  UUID,                                     -- 软关联（来源系统可能不知道 customer_id）
    meta         JSONB NOT NULL DEFAULT '{}'::jsonb,       -- {customer_name, account_id, ...}
    date         DATE NOT NULL,
    period_data  JSONB NOT NULL,
    data_source  VARCHAR(32) NOT NULL DEFAULT 'platform',
    imported_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (date, id)
) PARTITION BY RANGE (date);

CREATE INDEX IF NOT EXISTS idx_raw_mp_data_mp_date ON raw_mp_data(mp_id, date);
CREATE INDEX IF NOT EXISTS idx_raw_mp_data_customer ON raw_mp_data(customer_id);

DO $$
DECLARE
    yr INT := EXTRACT(YEAR FROM CURRENT_DATE)::INT;
    mo INT;
    start_date DATE;
    end_date DATE;
    partition_name TEXT;
BEGIN
    FOR mo IN 1..12 LOOP
        start_date := make_date(yr, mo, 1);
        end_date := start_date + INTERVAL '1 month';
        partition_name := format('raw_mp_data_y%sm%s', yr, lpad(mo::text, 2, '0'));
        EXECUTE format(
            'CREATE TABLE IF NOT EXISTS %I PARTITION OF raw_mp_data FOR VALUES FROM (%L) TO (%L)',
            partition_name, start_date, end_date
        );
    END LOOP;
END $$;
