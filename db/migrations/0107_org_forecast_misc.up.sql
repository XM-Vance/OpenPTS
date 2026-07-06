-- 多租户隔离收口（批次三）：预测/电网/碳/告警/负荷等域补 org_id。
-- 审查报告 P0-D1 系统性问题：0043 integration_tables、0042/0008/0014/0012/0006
-- 多张业务表建表未带 org_id。本迁移批量补齐。样板参照 0101/0055。

-- forecasts（无唯一约束，仅加列+回填+索引）
ALTER TABLE forecasts ADD COLUMN IF NOT EXISTS org_id uuid REFERENCES organizations(id);
UPDATE forecasts SET org_id = (SELECT id FROM organizations WHERE code='FJ') WHERE org_id IS NULL;
CREATE INDEX IF NOT EXISTS idx_forecasts_org ON forecasts(org_id);

-- grid_calculations（无唯一约束，仅加列+回填+索引）
ALTER TABLE grid_calculations ADD COLUMN IF NOT EXISTS org_id uuid REFERENCES organizations(id);
UPDATE grid_calculations SET org_id = (SELECT id FROM organizations WHERE code='FJ') WHERE org_id IS NULL;
CREATE INDEX IF NOT EXISTS idx_grid_calculations_org ON grid_calculations(org_id);

-- market_clearings（独立唯一索引 idx_clearing_unique → 含 org；用 DROP INDEX 非 DROP CONSTRAINT）
DROP INDEX IF EXISTS idx_clearing_unique;
ALTER TABLE market_clearings ADD COLUMN IF NOT EXISTS org_id uuid REFERENCES organizations(id);
UPDATE market_clearings SET org_id = (SELECT id FROM organizations WHERE code='FJ') WHERE org_id IS NULL;
CREATE INDEX IF NOT EXISTS idx_market_clearings_org ON market_clearings(org_id);
CREATE UNIQUE INDEX IF NOT EXISTS market_clearings_org_uniq ON market_clearings(org_id, market_type, trade_date, period);

-- carbon_emissions（无唯一约束，仅加列+回填+索引）
ALTER TABLE carbon_emissions ADD COLUMN IF NOT EXISTS org_id uuid REFERENCES organizations(id);
UPDATE carbon_emissions SET org_id = (SELECT id FROM organizations WHERE code='FJ') WHERE org_id IS NULL;
CREATE INDEX IF NOT EXISTS idx_carbon_emissions_org ON carbon_emissions(org_id);

-- tariff_schemes（无唯一约束，仅加列+回填+索引）
ALTER TABLE tariff_schemes ADD COLUMN IF NOT EXISTS org_id uuid REFERENCES organizations(id);
UPDATE tariff_schemes SET org_id = (SELECT id FROM organizations WHERE code='FJ') WHERE org_id IS NULL;
CREATE INDEX IF NOT EXISTS idx_tariff_schemes_org ON tariff_schemes(org_id);

-- storage_schedules（无唯一约束，仅加列+回填+索引）
ALTER TABLE storage_schedules ADD COLUMN IF NOT EXISTS org_id uuid REFERENCES organizations(id);
UPDATE storage_schedules SET org_id = (SELECT id FROM organizations WHERE code='FJ') WHERE org_id IS NULL;
CREATE INDEX IF NOT EXISTS idx_storage_schedules_org ON storage_schedules(org_id);

-- medium_term_load_forecast（无唯一约束，仅加列+回填+索引）
ALTER TABLE medium_term_load_forecast ADD COLUMN IF NOT EXISTS org_id uuid REFERENCES organizations(id);
UPDATE medium_term_load_forecast SET org_id = (SELECT id FROM organizations WHERE code='FJ') WHERE org_id IS NULL;
CREATE INDEX IF NOT EXISTS idx_medium_term_load_forecast_org ON medium_term_load_forecast(org_id);

-- actual_operation（无唯一约束，仅加列+回填+索引）
ALTER TABLE actual_operation ADD COLUMN IF NOT EXISTS org_id uuid REFERENCES organizations(id);
UPDATE actual_operation SET org_id = (SELECT id FROM organizations WHERE code='FJ') WHERE org_id IS NULL;
CREATE INDEX IF NOT EXISTS idx_actual_operation_org ON actual_operation(org_id);

-- mechanism_energy_monthly（列级 month UNIQUE → 含 org）
ALTER TABLE mechanism_energy_monthly ADD COLUMN IF NOT EXISTS org_id uuid REFERENCES organizations(id);
UPDATE mechanism_energy_monthly SET org_id = (SELECT id FROM organizations WHERE code='FJ') WHERE org_id IS NULL;
CREATE INDEX IF NOT EXISTS idx_mechanism_energy_monthly_org ON mechanism_energy_monthly(org_id);
ALTER TABLE mechanism_energy_monthly DROP CONSTRAINT IF EXISTS mechanism_energy_monthly_month_key;
CREATE UNIQUE INDEX IF NOT EXISTS mechanism_energy_monthly_org_uniq ON mechanism_energy_monthly(org_id, month);

-- unified_load_curve（子句 UNIQUE(customer_id, date) → 含 org）
ALTER TABLE unified_load_curve ADD COLUMN IF NOT EXISTS org_id uuid REFERENCES organizations(id);
UPDATE unified_load_curve SET org_id = (SELECT id FROM organizations WHERE code='FJ') WHERE org_id IS NULL;
CREATE INDEX IF NOT EXISTS idx_unified_load_curve_org ON unified_load_curve(org_id);
ALTER TABLE unified_load_curve DROP CONSTRAINT IF EXISTS unified_load_curve_customer_id_date_key;
CREATE UNIQUE INDEX IF NOT EXISTS unified_load_curve_org_uniq ON unified_load_curve(org_id, customer_id, date);

-- daily_release（列级 release_date UNIQUE → 含 org）
ALTER TABLE daily_release ADD COLUMN IF NOT EXISTS org_id uuid REFERENCES organizations(id);
UPDATE daily_release SET org_id = (SELECT id FROM organizations WHERE code='FJ') WHERE org_id IS NULL;
CREATE INDEX IF NOT EXISTS idx_daily_release_org ON daily_release(org_id);
ALTER TABLE daily_release DROP CONSTRAINT IF EXISTS daily_release_release_date_key;
CREATE UNIQUE INDEX IF NOT EXISTS daily_release_org_uniq ON daily_release(org_id, release_date);

-- real_time_generation（子句 UNIQUE(date, period, source_type) → 含 org）
ALTER TABLE real_time_generation ADD COLUMN IF NOT EXISTS org_id uuid REFERENCES organizations(id);
UPDATE real_time_generation SET org_id = (SELECT id FROM organizations WHERE code='FJ') WHERE org_id IS NULL;
CREATE INDEX IF NOT EXISTS idx_real_time_generation_org ON real_time_generation(org_id);
ALTER TABLE real_time_generation DROP CONSTRAINT IF EXISTS real_time_generation_date_period_source_type_key;
CREATE UNIQUE INDEX IF NOT EXISTS real_time_generation_org_uniq ON real_time_generation(org_id, date, period, source_type);

-- system_alerts（无唯一约束，仅加列+回填+索引）
ALTER TABLE system_alerts ADD COLUMN IF NOT EXISTS org_id uuid REFERENCES organizations(id);
UPDATE system_alerts SET org_id = (SELECT id FROM organizations WHERE code='FJ') WHERE org_id IS NULL;
CREATE INDEX IF NOT EXISTS idx_system_alerts_org ON system_alerts(org_id);
