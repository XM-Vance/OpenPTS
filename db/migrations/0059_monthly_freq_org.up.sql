-- 任务 9: 月度结算 / 频繁调频域 org_id

-- batch_monthly_settlement
ALTER TABLE batch_monthly_settlement ADD COLUMN IF NOT EXISTS org_id uuid REFERENCES organizations(id);
UPDATE batch_monthly_settlement SET org_id = (SELECT id FROM organizations WHERE code='default') WHERE org_id IS NULL;
CREATE INDEX IF NOT EXISTS idx_batch_monthly_settlement_org ON batch_monthly_settlement(org_id);
ALTER TABLE batch_monthly_settlement DROP CONSTRAINT IF EXISTS batch_monthly_settlement_operating_month_key;
CREATE UNIQUE INDEX IF NOT EXISTS batch_monthly_settlement_org_uniq ON batch_monthly_settlement(org_id, operating_month);

-- frequency_regulation_clearing
ALTER TABLE frequency_regulation_clearing ADD COLUMN IF NOT EXISTS org_id uuid REFERENCES organizations(id);
UPDATE frequency_regulation_clearing SET org_id = (SELECT id FROM organizations WHERE code='default') WHERE org_id IS NULL;
CREATE INDEX IF NOT EXISTS idx_freq_clearing_org ON frequency_regulation_clearing(org_id);
ALTER TABLE frequency_regulation_clearing DROP CONSTRAINT IF EXISTS frequency_regulation_clearing_settlement_date_regulation_ty_key;
CREATE UNIQUE INDEX IF NOT EXISTS frequency_regulation_clearing_org_uniq ON frequency_regulation_clearing(org_id, settlement_date, regulation_type);

-- frequency_regulation_demand
ALTER TABLE frequency_regulation_demand ADD COLUMN IF NOT EXISTS org_id uuid REFERENCES organizations(id);
UPDATE frequency_regulation_demand SET org_id = (SELECT id FROM organizations WHERE code='default') WHERE org_id IS NULL;
CREATE INDEX IF NOT EXISTS idx_freq_demand_org ON frequency_regulation_demand(org_id);
ALTER TABLE frequency_regulation_demand DROP CONSTRAINT IF EXISTS frequency_regulation_demand_settlement_date_key;
CREATE UNIQUE INDEX IF NOT EXISTS frequency_regulation_demand_org_uniq ON frequency_regulation_demand(org_id, settlement_date);

-- freq_comp_fee
ALTER TABLE freq_comp_fee ADD COLUMN IF NOT EXISTS org_id uuid REFERENCES organizations(id);
UPDATE freq_comp_fee SET org_id = (SELECT id FROM organizations WHERE code='default') WHERE org_id IS NULL;
CREATE INDEX IF NOT EXISTS idx_freq_comp_org ON freq_comp_fee(org_id);
ALTER TABLE freq_comp_fee DROP CONSTRAINT IF EXISTS freq_comp_fee_settlement_date_frequency_type_key;
CREATE UNIQUE INDEX IF NOT EXISTS freq_comp_fee_org_uniq ON freq_comp_fee(org_id, settlement_date, frequency_type);
