-- 114: weather_locations 种子（从 md_weather_hydrology_daily 已采到的站点派生）。
-- weather_actuals/forecasts 有 FK 到 weather_locations.name，此表此前为空，导致两张死表无法写入。
-- 动态派生：脚本新采到新区时自动有站点（首次 ETL 运行也会兜底补建，见 scheduler.FetchWeatherActuals）。
-- 若 md_weather_hydrology_daily 此时为空（脚本未跑过），本迁移插 0 行——ETL handler 兜底补建。

INSERT INTO weather_locations (name, province, city, latitude, longitude)
SELECT DISTINCT
    h.location_name,
    NULL,                          -- province：md 表无此列，由 ETL handler 按需补
    NULL,                          -- city：同上
    AVG(h.lat),                    -- 同名站点取均值经纬度
    AVG(h.lon)
FROM md_weather_hydrology_daily h
WHERE h.location_name IS NOT NULL
  AND NOT EXISTS (SELECT 1 FROM weather_locations w WHERE w.name = h.location_name)
GROUP BY h.location_name
ON CONFLICT (name) DO NOTHING;
