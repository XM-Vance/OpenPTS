-- 0013: 任务调度 + 执行记录 + 日志（分区）+ 日发布
-- 覆盖 v1：task_commands, task_execution_records, task_execution_logs, daily_release

-- ─── 任务命令（手工触发/取消等）────────────────────
CREATE TABLE IF NOT EXISTS task_commands (
    id             UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    command_id     VARCHAR(128) NOT NULL UNIQUE,
    pipeline_name  VARCHAR(128) NOT NULL,
    task_key       VARCHAR(128) NOT NULL,
    command_type   VARCHAR(32) NOT NULL,                   -- run / pause / cancel / retry
    status         VARCHAR(32) NOT NULL DEFAULT 'pending', -- pending / running / completed / failed
    payload        JSONB NOT NULL DEFAULT '{}'::jsonb,
    requested_by   UUID REFERENCES users(id),
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    started_at     TIMESTAMPTZ,
    completed_at   TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_task_cmd_pipeline ON task_commands(pipeline_name);
CREATE INDEX IF NOT EXISTS idx_task_cmd_status ON task_commands(status);
CREATE INDEX IF NOT EXISTS idx_task_cmd_created ON task_commands(created_at);

-- ─── 任务执行记录 ──────────────────────────────────
CREATE TABLE IF NOT EXISTS task_execution_records (
    id                UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    pipeline_name     VARCHAR(128) NOT NULL,
    task_key          VARCHAR(128) NOT NULL,
    execution_time    TIMESTAMPTZ NOT NULL,
    status            VARCHAR(32) NOT NULL,                 -- running / success / failed / cancelled
    records_inserted  INT NOT NULL DEFAULT 0,
    records_updated   INT NOT NULL DEFAULT 0,
    records_skipped   INT NOT NULL DEFAULT 0,
    duration_seconds  NUMERIC(8,3),
    error_message     TEXT,
    extra             JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_task_exec_pipeline_time ON task_execution_records(pipeline_name, execution_time);
CREATE INDEX IF NOT EXISTS idx_task_exec_status ON task_execution_records(status);

-- ─── 任务执行日志（高频写入，按月分区）─────────────
CREATE TABLE IF NOT EXISTS task_execution_logs (
    id             BIGSERIAL,
    pipeline_name  VARCHAR(128) NOT NULL,
    task_key       VARCHAR(128) NOT NULL,
    log_level      VARCHAR(16) NOT NULL,                    -- debug / info / warn / error
    message        TEXT NOT NULL,
    extra          JSONB NOT NULL DEFAULT '{}'::jsonb,
    timestamp      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (timestamp, id)
) PARTITION BY RANGE (timestamp);

CREATE INDEX IF NOT EXISTS idx_task_log_pipeline_time ON task_execution_logs(pipeline_name, timestamp);
CREATE INDEX IF NOT EXISTS idx_task_log_level ON task_execution_logs(log_level);

DO $$
DECLARE
    yr INT := EXTRACT(YEAR FROM CURRENT_DATE)::INT;
    mo INT;
    start_ts TIMESTAMPTZ;
    end_ts TIMESTAMPTZ;
    partition_name TEXT;
BEGIN
    FOR mo IN 1..12 LOOP
        start_ts := make_timestamptz(yr, mo, 1, 0, 0, 0);
        end_ts := start_ts + INTERVAL '1 month';
        partition_name := format('task_execution_logs_y%sm%s', yr, lpad(mo::text, 2, '0'));
        EXECUTE format(
            'CREATE TABLE IF NOT EXISTS %I PARTITION OF task_execution_logs FOR VALUES FROM (%L) TO (%L)',
            partition_name, start_ts, end_ts
        );
    END LOOP;
END $$;

-- ─── 日发布记录 ────────────────────────────────────
CREATE TABLE IF NOT EXISTS daily_release (
    id                    UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    release_date          DATE NOT NULL UNIQUE,
    released_collections  TEXT[] NOT NULL DEFAULT '{}',
    status                VARCHAR(32) NOT NULL DEFAULT 'pending',
    released_by           UUID REFERENCES users(id),
    released_at           TIMESTAMPTZ,
    note                  TEXT,
    created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at            TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

DROP TRIGGER IF EXISTS daily_release_set_updated_at ON daily_release;
CREATE TRIGGER daily_release_set_updated_at
    BEFORE UPDATE ON daily_release
    FOR EACH ROW EXECUTE FUNCTION trg_set_updated_at();
