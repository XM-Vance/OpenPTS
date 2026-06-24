-- 回滚：UUID → SERIAL（仅用于回退）
ALTER TABLE forecasts DROP CONSTRAINT forecasts_pkey;
ALTER TABLE forecasts DROP COLUMN id;
ALTER TABLE forecasts ADD COLUMN id SERIAL PRIMARY KEY;

ALTER TABLE grid_calculations DROP CONSTRAINT grid_calculations_pkey;
ALTER TABLE grid_calculations DROP COLUMN id;
ALTER TABLE grid_calculations ADD COLUMN id SERIAL PRIMARY KEY;

ALTER TABLE market_clearings DROP CONSTRAINT market_clearings_pkey;
ALTER TABLE market_clearings DROP COLUMN id;
ALTER TABLE market_clearings ADD COLUMN id SERIAL PRIMARY KEY;

ALTER TABLE carbon_emissions DROP CONSTRAINT carbon_emissions_pkey;
ALTER TABLE carbon_emissions DROP COLUMN id;
ALTER TABLE carbon_emissions ADD COLUMN id SERIAL PRIMARY KEY;

ALTER TABLE tariff_schemes DROP CONSTRAINT tariff_schemes_pkey;
ALTER TABLE tariff_schemes DROP COLUMN id;
ALTER TABLE tariff_schemes ADD COLUMN id SERIAL PRIMARY KEY;

ALTER TABLE storage_schedules DROP CONSTRAINT storage_schedules_pkey;
ALTER TABLE storage_schedules DROP COLUMN id;
ALTER TABLE storage_schedules ADD COLUMN id SERIAL PRIMARY KEY;
