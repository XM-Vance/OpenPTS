-- 调度器增强（P0-D2/D3/D4）：给 scheduled_jobs 增加重试与交易日历配置列。
-- - max_retries：任务失败时的最大重试次数（默认 0 = 不重试，保持旧行为）
-- - trade_day_only：是否仅在交易日（跳过周末 + 法定假日）执行
-- 这两列配合 scheduler.go 的互斥锁 + 重试循环 + 自定义 tradeDaySchedule 实现原子化。
ALTER TABLE scheduled_jobs ADD COLUMN IF NOT EXISTS max_retries integer NOT NULL DEFAULT 0;
ALTER TABLE scheduled_jobs ADD COLUMN IF NOT EXISTS trade_day_only boolean NOT NULL DEFAULT false;

-- fetch_market_data（每日 18:00 采集市场行情）开启交易日历过滤 + 1 次重试：
-- 周末/法定假日无行情，跳过；采集脚本偶发失败时重试一次。
UPDATE scheduled_jobs
   SET trade_day_only = true, max_retries = 1
 WHERE name = 'fetch_market_data';
