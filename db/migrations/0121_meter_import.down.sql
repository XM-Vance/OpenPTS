-- 138 down: 回滚计量底座列与索引（raw_meter_data 回到零写入死表状态；不影响其他表数据）。
DROP INDEX IF EXISTS idx_raw_meter_customer_date;
DROP INDEX IF EXISTS uq_raw_meter_meter_date;
ALTER TABLE raw_meter_data DROP COLUMN IF EXISTS customer_id;
