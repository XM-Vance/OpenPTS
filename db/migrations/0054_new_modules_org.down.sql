-- 回滚 0054: 新模块 org_id

-- contract_progress
DROP INDEX IF EXISTS contract_progress_org_uniq;
DROP INDEX IF EXISTS idx_contract_progress_org;
ALTER TABLE contract_progress DROP COLUMN IF EXISTS org_id;
ALTER TABLE contract_progress ADD CONSTRAINT contract_progress_contract_id_operating_month_key UNIQUE (contract_id, operating_month);

-- deviation_settlement
DROP INDEX IF EXISTS deviation_settlement_org_uniq;
DROP INDEX IF EXISTS idx_deviation_settlement_org;
ALTER TABLE deviation_settlement DROP COLUMN IF EXISTS org_id;
ALTER TABLE deviation_settlement ADD CONSTRAINT deviation_settlement_operating_date_category_key UNIQUE (operating_date, category);

-- green_power_trades
DROP INDEX IF EXISTS idx_green_power_trades_org;
ALTER TABLE green_power_trades DROP COLUMN IF EXISTS org_id;

-- rolling_trades
DROP INDEX IF EXISTS idx_rolling_trades_org;
ALTER TABLE rolling_trades DROP COLUMN IF EXISTS org_id;

-- spot_market_daily
DROP INDEX IF EXISTS spot_market_daily_org_uniq;
DROP INDEX IF EXISTS idx_spot_market_daily_org;
ALTER TABLE spot_market_daily DROP COLUMN IF EXISTS org_id;
ALTER TABLE spot_market_daily ADD CONSTRAINT spot_market_daily_trade_date_key UNIQUE (trade_date);

-- vpp_dispatches
DROP INDEX IF EXISTS idx_vpp_dispatches_org;
ALTER TABLE vpp_dispatches DROP COLUMN IF EXISTS org_id;

-- vpp_resources
DROP INDEX IF EXISTS vpp_resources_org_uniq;
DROP INDEX IF EXISTS idx_vpp_resources_org;
ALTER TABLE vpp_resources DROP COLUMN IF EXISTS org_id;
ALTER TABLE vpp_resources ADD CONSTRAINT vpp_resources_resource_name_key UNIQUE (resource_name);

-- bidding_records
DROP INDEX IF EXISTS bidding_records_org_uniq;
DROP INDEX IF EXISTS idx_bidding_records_org;
ALTER TABLE bidding_records DROP COLUMN IF EXISTS org_id;
ALTER TABLE bidding_records ADD CONSTRAINT bidding_records_trade_date_bidding_session_key UNIQUE (trade_date, bidding_session);
