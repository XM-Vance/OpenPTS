-- 0017: 储能站点 + 日运营记录
-- 储能模块缺表，0011 文件名虽含 storage 但只建了调频 3 张表；这里补齐。

CREATE TABLE IF NOT EXISTS storage_stations (
    id             UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name           VARCHAR(128) NOT NULL UNIQUE,
    capacity_mwh   NUMERIC(14,4) NOT NULL,
    max_power_mw   NUMERIC(14,4) NOT NULL,
    location       VARCHAR(255),
    status         VARCHAR(32) NOT NULL DEFAULT 'active', -- active / maintenance / offline
    extra          JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

DROP TRIGGER IF EXISTS storage_stations_set_updated_at ON storage_stations;
CREATE TRIGGER storage_stations_set_updated_at
    BEFORE UPDATE ON storage_stations
    FOR EACH ROW EXECUTE FUNCTION trg_set_updated_at();

CREATE TABLE IF NOT EXISTS storage_daily_operation (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    station_id      UUID NOT NULL REFERENCES storage_stations(id) ON DELETE CASCADE,
    operation_date  DATE NOT NULL,
    charge_mwh      NUMERIC(14,4) NOT NULL DEFAULT 0, -- 当日充电量
    discharge_mwh   NUMERIC(14,4) NOT NULL DEFAULT 0, -- 当日放电量
    revenue         NUMERIC(18,4),                     -- 当日收益（¥）
    avg_soc         NUMERIC(5,2),                      -- 平均荷电状态 0..100
    cycles          NUMERIC(6,2),                      -- 等效循环次数
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (station_id, operation_date)
);

CREATE INDEX IF NOT EXISTS idx_storage_op_date ON storage_daily_operation(operation_date);

DROP TRIGGER IF EXISTS storage_op_set_updated_at ON storage_daily_operation;
CREATE TRIGGER storage_op_set_updated_at
    BEFORE UPDATE ON storage_daily_operation
    FOR EACH ROW EXECUTE FUNCTION trg_set_updated_at();
