-- 0125: 给 md_weather_hydrology_daily 加 temp_max / temp_min 列。
-- Open-Meteo daily 接口本就返回 temperature_2m_max/min，此前脚本未采集、表也无此列，
-- 导致下游 weather_actuals 的 max_temp/min_temp 恒 NULL（前端 ActualsSummary 显示 0）。
-- 配套：脚本 upsert_hydrology 补采这两字段；FetchWeatherActuals ETL 透传。
ALTER TABLE md_weather_hydrology_daily
  ADD COLUMN IF NOT EXISTS temp_max DOUBLE PRECISION,
  ADD COLUMN IF NOT EXISTS temp_min DOUBLE PRECISION;
