-- 0123 down: 回滚 md_weather_wind_hourly 新增列（恢复到 0048 的 8 列结构）。
-- 仅一次性测试库使用；生产/共享库不要执行（AGENTS.md §3 down 守卫已拦）。
-- 注意：down 会丢弃这些列的历史数据，但 0123 之前本就写不进这些数据，无损。

ALTER TABLE md_weather_wind_hourly
  DROP COLUMN IF EXISTS city_name,
  DROP COLUMN IF EXISTS province_name,
  DROP COLUMN IF EXISTS apparent_temperature,
  DROP COLUMN IF EXISTS wind_speed_10m,
  DROP COLUMN IF EXISTS wind_direction_10m,
  DROP COLUMN IF EXISTS pressure_msl,
  DROP COLUMN IF EXISTS precipitation;
