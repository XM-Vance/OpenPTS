-- 调度任务（系统级 cron 定义）+ 执行记录。
-- 任务由 Go 进程内调度器（robfig/cron）按 cron 表达式触发，handler 字段决定具体执行体。
-- 复用 0015 已建的 task_scheduler 模块与 task_scheduler:read/write 权限点。

CREATE TABLE IF NOT EXISTS scheduled_jobs (
    id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name          text NOT NULL UNIQUE,
    description   text,
    cron_expr     text NOT NULL,
    handler       text NOT NULL,            -- 调度器内注册的处理器名（如 "cleanup_tokens"）
    enabled       boolean NOT NULL DEFAULT true,
    last_run_at   timestamp,
    last_status   text,                     -- success / failed
    last_error    text,
    created_at    timestamp NOT NULL DEFAULT now(),
    updated_at    timestamp NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS job_runs (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    job_id      uuid NOT NULL REFERENCES scheduled_jobs(id) ON DELETE CASCADE,
    started_at  timestamp NOT NULL DEFAULT now(),
    finished_at timestamp,
    status      text NOT NULL,              -- running / success / failed
    error       text,
    duration_ms integer,
    trigger     text NOT NULL DEFAULT 'cron' -- cron / manual
);

CREATE INDEX IF NOT EXISTS idx_job_runs_job_started
    ON job_runs(job_id, started_at DESC);

-- 预置 3 个内置任务（cron 表达式为 6 段制：秒 分 时 日 月 周）。
INSERT INTO scheduled_jobs (name, description, cron_expr, handler) VALUES
    ('cleanup_expired_tokens', '清理过期登录会话',           '0 */15 * * * *', 'cleanup_tokens'),
    ('aggregate_daily_active', '汇总日活用户',               '0 5 0 * * *',    'aggregate_daily_active'),
    ('refresh_dashboard_kpi',  '刷新仪表盘 KPI 缓存（占位）', '0 0 * * * *',    'refresh_dashboard_kpi')
ON CONFLICT (name) DO NOTHING;
