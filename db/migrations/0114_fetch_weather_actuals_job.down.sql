-- 115 down: 删除气象实况聚合定时任务。
DELETE FROM scheduled_jobs WHERE name = 'fetch_weather_actuals';
