-- 115: 注册气象实况聚合定时任务（每日 19:00，错开 fetch_market_data 18:00 让 md_weather 先采好）。
-- 配套 scheduler.FetchWeatherActuals handler：md_weather_hydrology_daily → weather_actuals（UPSERT）。
INSERT INTO scheduled_jobs (name, description, cron_expr, handler, enabled) VALUES
    ('fetch_weather_actuals', '聚合气象实况（md_weather_hydrology_daily → weather_actuals，接通死表）', '0 0 19 * * *', 'fetch_weather_actuals', true)
ON CONFLICT (name) DO NOTHING;
