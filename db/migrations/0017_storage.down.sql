DROP TRIGGER IF EXISTS storage_op_set_updated_at ON storage_daily_operation;
DROP TABLE IF EXISTS storage_daily_operation;

DROP TRIGGER IF EXISTS storage_stations_set_updated_at ON storage_stations;
DROP TABLE IF EXISTS storage_stations;
