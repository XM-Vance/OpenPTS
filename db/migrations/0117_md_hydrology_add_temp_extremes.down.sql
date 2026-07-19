-- 0125 down: 回滚 md_weather_hydrology_daily 新增的温度极值列。
ALTER TABLE md_weather_hydrology_daily
  DROP COLUMN IF EXISTS temp_max,
  DROP COLUMN IF EXISTS temp_min;
