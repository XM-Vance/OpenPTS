-- 回滚 0109：移除 scheduled_jobs 的 max_retries / trade_day_only 列。
-- 注意：回滚后 fetch_market_data 会恢复为每日含节假日执行（旧行为）。
ALTER TABLE scheduled_jobs DROP COLUMN IF EXISTS max_retries;
ALTER TABLE scheduled_jobs DROP COLUMN IF EXISTS trade_day_only;
