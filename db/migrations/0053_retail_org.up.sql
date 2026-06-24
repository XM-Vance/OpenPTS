-- 任务 1: 零售域 org_id
-- retail_contracts, retail_packages, retail_settlement_daily, retail_settlement_prices, retail_monthly_settlement

-- retail_contracts（无唯一约束）
ALTER TABLE retail_contracts ADD COLUMN IF NOT EXISTS org_id uuid REFERENCES organizations(id);
UPDATE retail_contracts SET org_id = (SELECT id FROM organizations WHERE code='default') WHERE org_id IS NULL;
CREATE INDEX IF NOT EXISTS idx_retail_contracts_org ON retail_contracts(org_id);

-- retail_packages（有 UNIQUE(package_name) → UNIQUE(org_id, package_name)）
ALTER TABLE retail_packages ADD COLUMN IF NOT EXISTS org_id uuid REFERENCES organizations(id);
UPDATE retail_packages SET org_id = (SELECT id FROM organizations WHERE code='default') WHERE org_id IS NULL;
CREATE INDEX IF NOT EXISTS idx_retail_packages_org ON retail_packages(org_id);
ALTER TABLE retail_packages DROP CONSTRAINT IF EXISTS retail_packages_package_name_key;
CREATE UNIQUE INDEX IF NOT EXISTS retail_packages_org_name_uniq ON retail_packages(org_id, package_name);

-- retail_settlement_daily（有 UNIQUE(customer_id, date) → UNIQUE(org_id, customer_id, date)）
ALTER TABLE retail_settlement_daily ADD COLUMN IF NOT EXISTS org_id uuid REFERENCES organizations(id);
UPDATE retail_settlement_daily SET org_id = (SELECT id FROM organizations WHERE code='default') WHERE org_id IS NULL;
CREATE INDEX IF NOT EXISTS idx_retail_settlement_daily_org ON retail_settlement_daily(org_id);
ALTER TABLE retail_settlement_daily DROP CONSTRAINT IF EXISTS retail_settlement_daily_customer_id_date_key;
CREATE UNIQUE INDEX IF NOT EXISTS retail_settlement_daily_org_uniq ON retail_settlement_daily(org_id, customer_id, date);

-- retail_settlement_prices（有 UNIQUE(date, customer_id) → UNIQUE(org_id, date, customer_id)）
ALTER TABLE retail_settlement_prices ADD COLUMN IF NOT EXISTS org_id uuid REFERENCES organizations(id);
UPDATE retail_settlement_prices SET org_id = (SELECT id FROM organizations WHERE code='default') WHERE org_id IS NULL;
CREATE INDEX IF NOT EXISTS idx_retail_settlement_prices_org ON retail_settlement_prices(org_id);
ALTER TABLE retail_settlement_prices DROP CONSTRAINT IF EXISTS retail_settlement_prices_date_customer_id_key;
CREATE UNIQUE INDEX IF NOT EXISTS retail_settlement_prices_org_uniq ON retail_settlement_prices(org_id, date, customer_id);

-- retail_monthly_settlement（有 UNIQUE(contract_id, operating_month) → UNIQUE(org_id, contract_id, operating_month)）
ALTER TABLE retail_monthly_settlement ADD COLUMN IF NOT EXISTS org_id uuid REFERENCES organizations(id);
UPDATE retail_monthly_settlement SET org_id = (SELECT id FROM organizations WHERE code='default') WHERE org_id IS NULL;
CREATE INDEX IF NOT EXISTS idx_retail_monthly_settlement_org ON retail_monthly_settlement(org_id);
ALTER TABLE retail_monthly_settlement DROP CONSTRAINT IF EXISTS retail_monthly_settlement_contract_id_operating_month_key;
CREATE UNIQUE INDEX IF NOT EXISTS retail_monthly_settlement_org_uniq ON retail_monthly_settlement(org_id, contract_id, operating_month);
