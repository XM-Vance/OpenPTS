-- 回滚 0105：结算/审批/撮合域 org_id（仅在一次性测试库执行；唯一键回退到旧单列，
-- 多省存量会冲突，故生产/共享库不得 down。遵守 AGENTS.md §3）。

-- contracts_aggregated_daily
DROP INDEX IF EXISTS contracts_aggregated_daily_org_uniq;
DROP INDEX IF EXISTS idx_contracts_aggregated_daily_org;
ALTER TABLE contracts_aggregated_daily DROP COLUMN IF EXISTS org_id;
ALTER TABLE contracts_aggregated_daily ADD CONSTRAINT contracts_aggregated_daily_date_key UNIQUE (date);

-- spot_settlement_period
DROP INDEX IF EXISTS spot_settlement_period_org_uniq;
DROP INDEX IF EXISTS idx_spot_settlement_period_org;
ALTER TABLE spot_settlement_period DROP COLUMN IF EXISTS org_id;
ALTER TABLE spot_settlement_period ADD CONSTRAINT spot_settlement_period_settlement_date_period_key UNIQUE (settlement_date, period);

-- spot_settlement_daily
DROP INDEX IF EXISTS spot_settlement_daily_org_uniq;
DROP INDEX IF EXISTS idx_spot_settlement_daily_org;
ALTER TABLE spot_settlement_daily DROP COLUMN IF EXISTS org_id;
ALTER TABLE spot_settlement_daily ADD CONSTRAINT spot_settlement_daily_settlement_date_key UNIQUE (settlement_date);

-- rolling_match_snapshots
DROP INDEX IF EXISTS rolling_match_snapshots_org_uniq;
DROP INDEX IF EXISTS idx_rolling_match_snapshots_org;
ALTER TABLE rolling_match_snapshots DROP COLUMN IF EXISTS org_id;
ALTER TABLE rolling_match_snapshots ADD CONSTRAINT rolling_match_snapshots_trade_date_delivery_date_key UNIQUE (trade_date, delivery_date);

-- approval_requests
DROP INDEX IF EXISTS idx_approval_requests_org;
ALTER TABLE approval_requests DROP COLUMN IF EXISTS org_id;

-- monthly_manual_data
DROP INDEX IF EXISTS idx_monthly_manual_data_org;
ALTER TABLE monthly_manual_data DROP COLUMN IF EXISTS org_id;

-- rolling_match_quotes
DROP INDEX IF EXISTS idx_rolling_match_quotes_org;
ALTER TABLE rolling_match_quotes DROP COLUMN IF EXISTS org_id;
