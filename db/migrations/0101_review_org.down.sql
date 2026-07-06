-- 回滚 0101：复盘两表 org_id（仅在一次性测试库执行；唯一键回退到旧单列，
-- 多省存量会冲突，故生产/共享库不得 down）。

-- day_ahead_trade_review
DROP INDEX IF EXISTS day_ahead_trade_review_org_uniq;
DROP INDEX IF EXISTS idx_day_ahead_trade_review_org;
ALTER TABLE day_ahead_trade_review DROP COLUMN IF EXISTS org_id;
ALTER TABLE day_ahead_trade_review ADD CONSTRAINT day_ahead_trade_review_trading_date_key UNIQUE (trading_date);

-- monthly_trade_review
DROP INDEX IF EXISTS monthly_trade_review_org_uniq;
DROP INDEX IF EXISTS idx_monthly_trade_review_org;
ALTER TABLE monthly_trade_review DROP COLUMN IF EXISTS org_id;
ALTER TABLE monthly_trade_review ADD CONSTRAINT monthly_trade_review_operating_month_key UNIQUE (operating_month);
