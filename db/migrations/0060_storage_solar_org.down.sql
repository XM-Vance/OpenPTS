-- 任务 10+12 回滚: 储能 + 光伏 + 储能申报 org_id

-- solar_revenue_settlement
DROP INDEX IF EXISTS idx_solar_revenue_org;
ALTER TABLE solar_revenue_settlement DROP COLUMN IF EXISTS org_id;

-- solar_generation_forecast
DROP INDEX IF EXISTS idx_solar_forecast_org;
ALTER TABLE solar_generation_forecast DROP COLUMN IF EXISTS org_id;

-- solar_stations
DROP INDEX IF EXISTS idx_solar_stations_org;
ALTER TABLE solar_stations DROP COLUMN IF EXISTS org_id;

-- storage_declaration
DROP INDEX IF EXISTS storage_declaration_org_uniq;
DROP INDEX IF EXISTS idx_storage_declaration_org;
ALTER TABLE storage_declaration DROP COLUMN IF EXISTS org_id;
ALTER TABLE storage_declaration ADD CONSTRAINT storage_declaration_station_id_declared_date_key UNIQUE (station_id, declared_date);

-- storage_daily_operation
DROP INDEX IF EXISTS storage_daily_op_org_uniq;
DROP INDEX IF EXISTS idx_storage_daily_op_org;
ALTER TABLE storage_daily_operation DROP COLUMN IF EXISTS org_id;
ALTER TABLE storage_daily_operation ADD CONSTRAINT storage_daily_operation_station_id_operation_date_key UNIQUE (station_id, operation_date);

-- storage_stations
DROP INDEX IF EXISTS storage_stations_org_name_uniq;
DROP INDEX IF EXISTS idx_storage_stations_org;
ALTER TABLE storage_stations DROP COLUMN IF EXISTS org_id;
ALTER TABLE storage_stations ADD CONSTRAINT storage_stations_name_key UNIQUE (name);
