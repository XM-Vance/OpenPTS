-- 任务 8: 现货交易域（日前模拟 + 结算）org_id

-- da_simulation_scenarios
ALTER TABLE da_simulation_scenarios ADD COLUMN IF NOT EXISTS org_id uuid REFERENCES organizations(id);
UPDATE da_simulation_scenarios SET org_id = (SELECT id FROM organizations WHERE code='default') WHERE org_id IS NULL;
CREATE INDEX IF NOT EXISTS idx_da_sim_scenarios_org ON da_simulation_scenarios(org_id);

-- settlement_daily
ALTER TABLE settlement_daily ADD COLUMN IF NOT EXISTS org_id uuid REFERENCES organizations(id);
UPDATE settlement_daily SET org_id = (SELECT id FROM organizations WHERE code='default') WHERE org_id IS NULL;
CREATE INDEX IF NOT EXISTS idx_settlement_daily_org ON settlement_daily(org_id);
ALTER TABLE settlement_daily DROP CONSTRAINT IF EXISTS settlement_daily_operating_date_version_key;
CREATE UNIQUE INDEX IF NOT EXISTS settlement_daily_org_uniq ON settlement_daily(org_id, operating_date, version);
