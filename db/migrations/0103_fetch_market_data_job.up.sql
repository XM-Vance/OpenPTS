-- 0103: 注册市场行情采集定时任务（每日 18:00 收盘后刷新 md_* 表）
-- 配套 scheduler.FetchMarketData handler（exec AKShare 脚本）。
INSERT INTO scheduled_jobs (name, description, cron_expr, handler, enabled) VALUES
    ('fetch_market_data', '采集市场行情数据（AKShare → md_* 表，宏观/燃料/期货/汇率/利率）', '0 0 18 * * *', 'fetch_market_data', true)
ON CONFLICT (name) DO NOTHING;
