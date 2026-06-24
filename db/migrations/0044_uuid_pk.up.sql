-- 统一新表主键风格为 UUID（与全库其它表一致）
-- 仅对 0043 新增的 6 张表操作，且当前表无数据

-- 1. forecasts: id SERIAL → UUID
ALTER TABLE forecasts DROP CONSTRAINT forecasts_pkey;
ALTER TABLE forecasts DROP COLUMN id;
ALTER TABLE forecasts ADD COLUMN id UUID PRIMARY KEY DEFAULT gen_random_uuid();
-- 重建索引（id 列变了，需重建）
DROP INDEX IF EXISTS idx_forecasts_type_date;
DROP INDEX IF EXISTS idx_forecasts_type_region;
CREATE INDEX idx_forecasts_type_date ON forecasts(type, target_date);
CREATE INDEX idx_forecasts_type_region ON forecasts(type, region, target_date);

-- 2. grid_calculations: id SERIAL → UUID
ALTER TABLE grid_calculations DROP CONSTRAINT grid_calculations_pkey;
ALTER TABLE grid_calculations DROP COLUMN id;
ALTER TABLE grid_calculations ADD COLUMN id UUID PRIMARY KEY DEFAULT gen_random_uuid();
DROP INDEX IF EXISTS idx_grid_calc_type;
CREATE INDEX idx_grid_calc_type ON grid_calculations(calc_type, snapshot_time);

-- 3. market_clearings: id SERIAL → UUID
ALTER TABLE market_clearings DROP CONSTRAINT market_clearings_pkey;
ALTER TABLE market_clearings DROP COLUMN id;
ALTER TABLE market_clearings ADD COLUMN id UUID PRIMARY KEY DEFAULT gen_random_uuid();
-- idx_clearing_unique 不含 id，无需改

-- 4. carbon_emissions: id SERIAL → UUID
ALTER TABLE carbon_emissions DROP CONSTRAINT carbon_emissions_pkey;
ALTER TABLE carbon_emissions DROP COLUMN id;
ALTER TABLE carbon_emissions ADD COLUMN id UUID PRIMARY KEY DEFAULT gen_random_uuid();
DROP INDEX IF EXISTS idx_carbon_entity;
CREATE INDEX idx_carbon_entity ON carbon_emissions(entity_type, entity_id, period_start);

-- 5. tariff_schemes: id SERIAL → UUID
ALTER TABLE tariff_schemes DROP CONSTRAINT tariff_schemes_pkey;
ALTER TABLE tariff_schemes DROP COLUMN id;
ALTER TABLE tariff_schemes ADD COLUMN id UUID PRIMARY KEY DEFAULT gen_random_uuid();
DROP INDEX IF EXISTS idx_tariff_name;
CREATE INDEX idx_tariff_name ON tariff_schemes(name, valid_from);

-- 6. storage_schedules: id SERIAL → UUID
ALTER TABLE storage_schedules DROP CONSTRAINT storage_schedules_pkey;
ALTER TABLE storage_schedules DROP COLUMN id;
ALTER TABLE storage_schedules ADD COLUMN id UUID PRIMARY KEY DEFAULT gen_random_uuid();
DROP INDEX IF EXISTS idx_storage_schedule;
CREATE INDEX idx_storage_schedule ON storage_schedules(storage_id, schedule_date);
