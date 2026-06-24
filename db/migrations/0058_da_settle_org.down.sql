-- 任务 8 回滚: 现货交易域 org_id

-- settlement_daily
DROP INDEX IF EXISTS settlement_daily_org_uniq;
DROP INDEX IF EXISTS idx_settlement_daily_org;
ALTER TABLE settlement_daily DROP COLUMN IF EXISTS org_id;
ALTER TABLE settlement_daily ADD CONSTRAINT settlement_daily_operating_date_version_key UNIQUE (operating_date, version);

-- da_simulation_scenarios
DROP INDEX IF EXISTS idx_da_sim_scenarios_org;
ALTER TABLE da_simulation_scenarios DROP COLUMN IF EXISTS org_id;
