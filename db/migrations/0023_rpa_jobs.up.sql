-- RPA 任务监控：外部 RPA 流程的定义 + 运行记录。
-- 与 scheduled_jobs 区分：scheduled_jobs 是 v2 进程内 cron；rpa_jobs 是接外部 RPA 平台运行结果。

CREATE TABLE IF NOT EXISTS rpa_jobs (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name        text NOT NULL UNIQUE,
    description text,
    schedule    text,                       -- 计划描述（自然语言或 cron）
    enabled     boolean NOT NULL DEFAULT true,
    last_run_at timestamp,
    last_status text,                       -- success / failed / running
    created_at  timestamp NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS rpa_runs (
    id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    rpa_job_id    uuid NOT NULL REFERENCES rpa_jobs(id) ON DELETE CASCADE,
    started_at    timestamp NOT NULL DEFAULT now(),
    finished_at   timestamp,
    status        text NOT NULL,
    duration_sec  integer,
    output_files  integer NOT NULL DEFAULT 0,
    output_bytes  bigint NOT NULL DEFAULT 0,
    error         text
);

CREATE INDEX IF NOT EXISTS idx_rpa_runs_job_started
    ON rpa_runs(rpa_job_id, started_at DESC);
