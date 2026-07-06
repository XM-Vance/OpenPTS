-- 100: 日活用户统计表（让 scheduler 的 AggregateDailyActive 把 DAU 落库，供趋势查询）。
-- 数据源：auth_sessions（每日去重 user_id），由调度任务聚合写入。
CREATE TABLE IF NOT EXISTS dau_daily (
    date         DATE PRIMARY KEY,
    user_count   INTEGER NOT NULL DEFAULT 0,
    computed_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
