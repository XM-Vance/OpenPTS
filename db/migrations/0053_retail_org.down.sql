-- 回滚 0053: 零售域 org_id

-- retail_contracts（无唯一约束）
DROP INDEX IF EXISTS idx_retail_contracts_org;
ALTER TABLE retail_contracts DROP COLUMN IF EXISTS org_id;

-- retail_packages
DROP INDEX IF EXISTS retail_packages_org_name_uniq;
DROP INDEX IF EXISTS idx_retail_packages_org;
ALTER TABLE retail_packages DROP COLUMN IF EXISTS org_id;
ALTER TABLE retail_packages ADD CONSTRAINT retail_packages_package_name_key UNIQUE (package_name);

-- retail_settlement_daily
DROP INDEX IF EXISTS retail_settlement_daily_org_uniq;
DROP INDEX IF EXISTS idx_retail_settlement_daily_org;
ALTER TABLE retail_settlement_daily DROP COLUMN IF EXISTS org_id;
ALTER TABLE retail_settlement_daily ADD CONSTRAINT retail_settlement_daily_customer_id_date_key UNIQUE (customer_id, date);

-- retail_settlement_prices
DROP INDEX IF EXISTS retail_settlement_prices_org_uniq;
DROP INDEX IF EXISTS idx_retail_settlement_prices_org;
ALTER TABLE retail_settlement_prices DROP COLUMN IF EXISTS org_id;
ALTER TABLE retail_settlement_prices ADD CONSTRAINT retail_settlement_prices_date_customer_id_key UNIQUE (date, customer_id);

-- retail_monthly_settlement
DROP INDEX IF EXISTS retail_monthly_settlement_org_uniq;
DROP INDEX IF EXISTS idx_retail_monthly_settlement_org;
ALTER TABLE retail_monthly_settlement DROP COLUMN IF EXISTS org_id;
ALTER TABLE retail_monthly_settlement ADD CONSTRAINT retail_monthly_settlement_contract_id_operating_month_key UNIQUE (contract_id, operating_month);
