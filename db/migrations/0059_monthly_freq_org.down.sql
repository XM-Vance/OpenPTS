-- 任务 9 回滚: 月度结算 / 频繁调频域 org_id

-- freq_comp_fee
DROP INDEX IF EXISTS freq_comp_fee_org_uniq;
DROP INDEX IF EXISTS idx_freq_comp_org;
ALTER TABLE freq_comp_fee DROP COLUMN IF EXISTS org_id;
ALTER TABLE freq_comp_fee ADD CONSTRAINT freq_comp_fee_settlement_date_frequency_type_key UNIQUE (settlement_date, frequency_type);

-- frequency_regulation_demand
DROP INDEX IF EXISTS frequency_regulation_demand_org_uniq;
DROP INDEX IF EXISTS idx_freq_demand_org;
ALTER TABLE frequency_regulation_demand DROP COLUMN IF EXISTS org_id;
ALTER TABLE frequency_regulation_demand ADD CONSTRAINT frequency_regulation_demand_settlement_date_key UNIQUE (settlement_date);

-- frequency_regulation_clearing
DROP INDEX IF EXISTS frequency_regulation_clearing_org_uniq;
DROP INDEX IF EXISTS idx_freq_clearing_org;
ALTER TABLE frequency_regulation_clearing DROP COLUMN IF EXISTS org_id;
ALTER TABLE frequency_regulation_clearing ADD CONSTRAINT frequency_regulation_clearing_settlement_date_regulation_ty_key UNIQUE (settlement_date, regulation_type);

-- batch_monthly_settlement
DROP INDEX IF EXISTS batch_monthly_settlement_org_uniq;
DROP INDEX IF EXISTS idx_batch_monthly_settlement_org;
ALTER TABLE batch_monthly_settlement DROP COLUMN IF EXISTS org_id;
ALTER TABLE batch_monthly_settlement ADD CONSTRAINT batch_monthly_settlement_operating_month_key UNIQUE (operating_month);
