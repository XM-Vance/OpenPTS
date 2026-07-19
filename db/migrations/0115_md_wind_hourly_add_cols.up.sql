-- 0123: 扩展 md_weather_wind_hourly 表结构，对齐 Open-Meteo 采集脚本字段。
--
-- 背景：fetch_weather_data.py 的 upsert_hourly 写入 15 列，而 0048 建表仅 8 列，
-- 导致 hourly 数据因列名不存在而静默失败（脚本捕获异常后 continue）。
-- 本迁移给表补齐脚本采集但表里缺失的列，使逐小时观测能正常落库。
-- 已有列（无需改）：location_code, location_name, lat, lon, obs_time,
--                  wind_speed_100m, wind_dir_100m, temperature_2m, humidity_2m。

ALTER TABLE md_weather_wind_hourly
  ADD COLUMN IF NOT EXISTS city_name            VARCHAR(50),
  ADD COLUMN IF NOT EXISTS province_name        VARCHAR(50),
  ADD COLUMN IF NOT EXISTS apparent_temperature DOUBLE PRECISION,
  ADD COLUMN IF NOT EXISTS wind_speed_10m       DOUBLE PRECISION,
  ADD COLUMN IF NOT EXISTS wind_direction_10m   SMALLINT,
  ADD COLUMN IF NOT EXISTS pressure_msl         DOUBLE PRECISION,
  ADD COLUMN IF NOT EXISTS precipitation        DOUBLE PRECISION;
