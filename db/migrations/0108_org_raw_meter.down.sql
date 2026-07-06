-- 回滚 0108：电表原始数据分区表 org_id（仅在一次性测试库执行；生产/共享库不得 down）。

-- raw_mp_data
DROP INDEX IF EXISTS idx_raw_mp_data_org;
ALTER TABLE raw_mp_data DROP COLUMN IF EXISTS org_id;

-- raw_meter_data
DROP INDEX IF EXISTS idx_raw_meter_data_org;
ALTER TABLE raw_meter_data DROP COLUMN IF EXISTS org_id;
