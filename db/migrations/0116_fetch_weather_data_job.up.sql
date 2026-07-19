-- 0124: 注册气象数据采集定时任务（每日 17:00，Open-Meteo → md_weather_* 表）。
-- 配套 scheduler.FetchWeatherData handler：exec scripts/data-collection/fetch_weather_data.py。
-- 时序：17:00 采气象原始数据 → 18:00 采市场行情 → 19:00 聚合 md_weather → weather_actuals。
-- 气象数据非交易日特有，不设 trade_day_only（天天采）。
INSERT INTO scheduled_jobs (name, description, cron_expr, handler, enabled) VALUES
    ('fetch_weather_data', '采集气象数据（Open-Meteo → md_weather_wind_hourly/md_weather_hydrology_daily，风电场逐时+水库水文逐日）', '0 0 17 * * *', 'fetch_weather_data', true)
ON CONFLICT (name) DO NOTHING;
