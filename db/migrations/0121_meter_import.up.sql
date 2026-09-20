-- 138: 计量数据底座（WP0.3）——raw_meter_data 补幂等键与客户关联。
-- 此前 raw_meter_data 建表（0005 分区表 + 0108 org_id）后零写入路径：
--   1) 无 (meter_id, date) 唯一约束，重复导入会堆重复行 → 加唯一索引支撑 ON CONFLICT 幂等 upsert；
--   2) 无 customer 关联，无法聚合到 user_load_data → 加 customer_id（导入时按户号/客户名解析）。
-- 分区表父表 ADD COLUMN / CREATE INDEX 自动传播到全部子分区（同 0108 先例）。
-- 表当前为空表，加唯一索引无历史去重问题。

ALTER TABLE raw_meter_data ADD COLUMN IF NOT EXISTS customer_id uuid REFERENCES customers(id);
CREATE UNIQUE INDEX IF NOT EXISTS uq_raw_meter_meter_date ON raw_meter_data(meter_id, date);
CREATE INDEX IF NOT EXISTS idx_raw_meter_customer_date ON raw_meter_data(customer_id, date);
