-- 注入合同自动到期归档调度任务。
-- 与 0018 的 3 个内置任务一样，启动时由 scheduler.Start 自动加载执行。

INSERT INTO scheduled_jobs (name, description, cron_expr, handler) VALUES
    ('expire_contracts', '把购电期已截止的活跃合同自动标记 expired', '0 0 2 * * *', 'expire_contracts')
ON CONFLICT (name) DO NOTHING;
