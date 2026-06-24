-- 任务 7: 价格域 org_id
-- contract_price_daily
ALTER TABLE contract_price_daily ADD COLUMN IF NOT EXISTS org_id uuid REFERENCES organizations(id);
UPDATE contract_price_daily SET org_id = (SELECT id FROM organizations WHERE code='default') WHERE org_id IS NULL;
CREATE INDEX IF NOT EXISTS idx_contract_price_daily_org ON contract_price_daily(org_id);
ALTER TABLE contract_price_daily DROP CONSTRAINT IF EXISTS contract_price_daily_contract_id_price_date_key;
CREATE UNIQUE INDEX IF NOT EXISTS contract_price_daily_org_uniq ON contract_price_daily(org_id, contract_id, price_date);

-- day_ahead_spot_price (现货趋势聚合)
ALTER TABLE day_ahead_spot_price ADD COLUMN IF NOT EXISTS org_id uuid REFERENCES organizations(id);
UPDATE day_ahead_spot_price SET org_id = (SELECT id FROM organizations WHERE code='default') WHERE org_id IS NULL;
CREATE INDEX IF NOT EXISTS idx_day_ahead_spot_price_org ON day_ahead_spot_price(org_id);
ALTER TABLE day_ahead_spot_price DROP CONSTRAINT IF EXISTS day_ahead_spot_price_date_period_key;
CREATE UNIQUE INDEX IF NOT EXISTS day_ahead_spot_price_org_uniq ON day_ahead_spot_price(org_id, date, period);

-- tou_rules
ALTER TABLE tou_rules ADD COLUMN IF NOT EXISTS org_id uuid REFERENCES organizations(id);
UPDATE tou_rules SET org_id = (SELECT id FROM organizations WHERE code='default') WHERE org_id IS NULL;
CREATE INDEX IF NOT EXISTS idx_tou_rules_org ON tou_rules(org_id);

-- grid_agency_price
ALTER TABLE grid_agency_price ADD COLUMN IF NOT EXISTS org_id uuid REFERENCES organizations(id);
UPDATE grid_agency_price SET org_id = (SELECT id FROM organizations WHERE code='default') WHERE org_id IS NULL;
CREATE INDEX IF NOT EXISTS idx_grid_agency_price_org ON grid_agency_price(org_id);
ALTER TABLE grid_agency_price DROP CONSTRAINT IF EXISTS grid_agency_price_operating_month_voltage_level_key;
CREATE UNIQUE INDEX IF NOT EXISTS grid_agency_price_org_uniq ON grid_agency_price(org_id, operating_month, voltage_level);
