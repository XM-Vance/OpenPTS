-- 任务 10+12: 储能 + 光伏 + 储能申报 org_id

-- storage_stations: UNIQUE(name) → UNIQUE(org_id, name)
ALTER TABLE storage_stations ADD COLUMN IF NOT EXISTS org_id uuid REFERENCES organizations(id);
UPDATE storage_stations SET org_id = (SELECT id FROM organizations WHERE code='default') WHERE org_id IS NULL;
CREATE INDEX IF NOT EXISTS idx_storage_stations_org ON storage_stations(org_id);
ALTER TABLE storage_stations DROP CONSTRAINT IF EXISTS storage_stations_name_key;
CREATE UNIQUE INDEX IF NOT EXISTS storage_stations_org_name_uniq ON storage_stations(org_id, name);

-- storage_daily_operation: UNIQUE(station_id,operation_date) → UNIQUE(org_id,station_id,operation_date)
ALTER TABLE storage_daily_operation ADD COLUMN IF NOT EXISTS org_id uuid REFERENCES organizations(id);
UPDATE storage_daily_operation SET org_id = (SELECT id FROM organizations WHERE code='default') WHERE org_id IS NULL;
CREATE INDEX IF NOT EXISTS idx_storage_daily_op_org ON storage_daily_operation(org_id);
ALTER TABLE storage_daily_operation DROP CONSTRAINT IF EXISTS storage_daily_operation_station_id_operation_date_key;
CREATE UNIQUE INDEX IF NOT EXISTS storage_daily_op_org_uniq ON storage_daily_operation(org_id, station_id, operation_date);

-- storage_declaration: UNIQUE(station_id,declared_date) → UNIQUE(org_id,station_id,declared_date)
ALTER TABLE storage_declaration ADD COLUMN IF NOT EXISTS org_id uuid REFERENCES organizations(id);
UPDATE storage_declaration SET org_id = (SELECT id FROM organizations WHERE code='default') WHERE org_id IS NULL;
CREATE INDEX IF NOT EXISTS idx_storage_declaration_org ON storage_declaration(org_id);
ALTER TABLE storage_declaration DROP CONSTRAINT IF EXISTS storage_declaration_station_id_declared_date_key;
CREATE UNIQUE INDEX IF NOT EXISTS storage_declaration_org_uniq ON storage_declaration(org_id, station_id, declared_date);

-- solar_stations: 无唯一约束，只加 org_id + 索引
ALTER TABLE solar_stations ADD COLUMN IF NOT EXISTS org_id uuid REFERENCES organizations(id);
UPDATE solar_stations SET org_id = (SELECT id FROM organizations WHERE code='default') WHERE org_id IS NULL;
CREATE INDEX IF NOT EXISTS idx_solar_stations_org ON solar_stations(org_id);

-- solar_generation_forecast: 无唯一约束(索引idx_solar_forecast_uniq是UNIQUE INDEX不是约束)
ALTER TABLE solar_generation_forecast ADD COLUMN IF NOT EXISTS org_id uuid REFERENCES organizations(id);
UPDATE solar_generation_forecast SET org_id = (SELECT id FROM organizations WHERE code='default') WHERE org_id IS NULL;
CREATE INDEX IF NOT EXISTS idx_solar_forecast_org ON solar_generation_forecast(org_id);

-- solar_revenue_settlement: 无唯一约束(索引idx_solar_revenue_uniq是UNIQUE INDEX不是约束)
ALTER TABLE solar_revenue_settlement ADD COLUMN IF NOT EXISTS org_id uuid REFERENCES organizations(id);
UPDATE solar_revenue_settlement SET org_id = (SELECT id FROM organizations WHERE code='default') WHERE org_id IS NULL;
CREATE INDEX IF NOT EXISTS idx_solar_revenue_org ON solar_revenue_settlement(org_id);
