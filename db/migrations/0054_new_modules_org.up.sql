-- 任务 #4/#4b: 新模块 org_id
-- contract_progress, deviation_settlement, green_power_trades, rolling_trades,
-- spot_market_daily, vpp_dispatches, vpp_resources, bidding_records

-- contract_progress（有 UNIQUE(contract_id, operating_month) → UNIQUE(org_id, contract_id, operating_month)）
ALTER TABLE contract_progress ADD COLUMN IF NOT EXISTS org_id uuid REFERENCES organizations(id);
UPDATE contract_progress SET org_id = (SELECT id FROM organizations WHERE code='default') WHERE org_id IS NULL;
CREATE INDEX IF NOT EXISTS idx_contract_progress_org ON contract_progress(org_id);
ALTER TABLE contract_progress DROP CONSTRAINT IF EXISTS contract_progress_contract_id_operating_month_key;
CREATE UNIQUE INDEX IF NOT EXISTS contract_progress_org_uniq ON contract_progress(org_id, contract_id, operating_month);

-- deviation_settlement（有 UNIQUE(operating_date, category) → UNIQUE(org_id, operating_date, category)）
ALTER TABLE deviation_settlement ADD COLUMN IF NOT EXISTS org_id uuid REFERENCES organizations(id);
UPDATE deviation_settlement SET org_id = (SELECT id FROM organizations WHERE code='default') WHERE org_id IS NULL;
CREATE INDEX IF NOT EXISTS idx_deviation_settlement_org ON deviation_settlement(org_id);
ALTER TABLE deviation_settlement DROP CONSTRAINT IF EXISTS deviation_settlement_operating_date_category_key;
CREATE UNIQUE INDEX IF NOT EXISTS deviation_settlement_org_uniq ON deviation_settlement(org_id, operating_date, category);

-- green_power_trades（无唯一约束）
ALTER TABLE green_power_trades ADD COLUMN IF NOT EXISTS org_id uuid REFERENCES organizations(id);
UPDATE green_power_trades SET org_id = (SELECT id FROM organizations WHERE code='default') WHERE org_id IS NULL;
CREATE INDEX IF NOT EXISTS idx_green_power_trades_org ON green_power_trades(org_id);

-- rolling_trades（无唯一约束）
ALTER TABLE rolling_trades ADD COLUMN IF NOT EXISTS org_id uuid REFERENCES organizations(id);
UPDATE rolling_trades SET org_id = (SELECT id FROM organizations WHERE code='default') WHERE org_id IS NULL;
CREATE INDEX IF NOT EXISTS idx_rolling_trades_org ON rolling_trades(org_id);

-- spot_market_daily（有 UNIQUE(trade_date) → UNIQUE(org_id, trade_date)）
ALTER TABLE spot_market_daily ADD COLUMN IF NOT EXISTS org_id uuid REFERENCES organizations(id);
UPDATE spot_market_daily SET org_id = (SELECT id FROM organizations WHERE code='default') WHERE org_id IS NULL;
CREATE INDEX IF NOT EXISTS idx_spot_market_daily_org ON spot_market_daily(org_id);
ALTER TABLE spot_market_daily DROP CONSTRAINT IF EXISTS spot_market_daily_trade_date_key;
CREATE UNIQUE INDEX IF NOT EXISTS spot_market_daily_org_uniq ON spot_market_daily(org_id, trade_date);

-- vpp_dispatches（无唯一约束）
ALTER TABLE vpp_dispatches ADD COLUMN IF NOT EXISTS org_id uuid REFERENCES organizations(id);
UPDATE vpp_dispatches SET org_id = (SELECT id FROM organizations WHERE code='default') WHERE org_id IS NULL;
CREATE INDEX IF NOT EXISTS idx_vpp_dispatches_org ON vpp_dispatches(org_id);

-- vpp_resources（有 UNIQUE(resource_name) → UNIQUE(org_id, resource_name)）
ALTER TABLE vpp_resources ADD COLUMN IF NOT EXISTS org_id uuid REFERENCES organizations(id);
UPDATE vpp_resources SET org_id = (SELECT id FROM organizations WHERE code='default') WHERE org_id IS NULL;
CREATE INDEX IF NOT EXISTS idx_vpp_resources_org ON vpp_resources(org_id);
ALTER TABLE vpp_resources DROP CONSTRAINT IF EXISTS vpp_resources_resource_name_key;
CREATE UNIQUE INDEX IF NOT EXISTS vpp_resources_org_uniq ON vpp_resources(org_id, resource_name);

-- bidding_records（有 UNIQUE(trade_date, bidding_session) → UNIQUE(org_id, trade_date, bidding_session)）
ALTER TABLE bidding_records ADD COLUMN IF NOT EXISTS org_id uuid REFERENCES organizations(id);
UPDATE bidding_records SET org_id = (SELECT id FROM organizations WHERE code='default') WHERE org_id IS NULL;
CREATE INDEX IF NOT EXISTS idx_bidding_records_org ON bidding_records(org_id);
ALTER TABLE bidding_records DROP CONSTRAINT IF EXISTS bidding_records_trade_date_bidding_session_key;
CREATE UNIQUE INDEX IF NOT EXISTS bidding_records_org_uniq ON bidding_records(org_id, trade_date, bidding_session);
