-- 0103 down: 移除市场行情采集任务
DELETE FROM scheduled_jobs WHERE name = 'fetch_market_data';
