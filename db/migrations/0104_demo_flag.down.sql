-- 0104 down: 移除 is_demo 列（仅一次性验证库执行）
ALTER TABLE customer_profit DROP COLUMN IF EXISTS is_demo;
ALTER TABLE forecast_accuracy DROP COLUMN IF EXISTS is_demo;
ALTER TABLE monthly_trade_review DROP COLUMN IF EXISTS is_demo;
ALTER TABLE day_ahead_trade_review DROP COLUMN IF EXISTS is_demo;
ALTER TABLE settlement_daily DROP COLUMN IF EXISTS is_demo;
ALTER TABLE deviation_settlement DROP COLUMN IF EXISTS is_demo;
ALTER TABLE rolling_trades DROP COLUMN IF EXISTS is_demo;
ALTER TABLE green_power_trades DROP COLUMN IF EXISTS is_demo;
ALTER TABLE bidding_records DROP COLUMN IF EXISTS is_demo;
ALTER TABLE carbon_quotes DROP COLUMN IF EXISTS is_demo;
