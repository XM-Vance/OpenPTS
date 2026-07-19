-- 0124 down: 移除气象数据采集定时任务注册。
-- scheduled_jobs 删除后，job_runs 记录由 ON DELETE CASCADE 自动清理。
DELETE FROM scheduled_jobs WHERE name = 'fetch_weather_data';
