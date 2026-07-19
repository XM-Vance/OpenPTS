-- 114 down: 回滚 weather_locations 种子。
-- 注意：只删"由本 seed 从 md_weather_hydrology_daily 派生且 province/city 为 NULL"的站点
-- （即 ETL 自动建的），保留用户手动创建的站点（有 province/city 或被 customers.location 引用的）。
DELETE FROM weather_locations
WHERE province IS NULL AND city IS NULL
  AND name IN (SELECT location_name FROM md_weather_hydrology_daily);
