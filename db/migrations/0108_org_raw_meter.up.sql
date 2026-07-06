-- 多租户隔离收口（批次四）：电表原始数据分区表补 org_id。
-- raw_meter_data / raw_mp_data 是 PARTITION BY RANGE(date) 分区表，给父表 ADD COLUMN
-- 会自动传播到所有子分区，回填 UPDATE 也会作用于全部分区。当前主键 (date, id) 不含
-- org 维度（电表 id 本身全局唯一，无需 org 去重），故仅加列+回填+索引，不动主键。
-- 样板参照 0101/0055。

-- raw_meter_data
ALTER TABLE raw_meter_data ADD COLUMN IF NOT EXISTS org_id uuid REFERENCES organizations(id);
UPDATE raw_meter_data SET org_id = (SELECT id FROM organizations WHERE code='default') WHERE org_id IS NULL;
CREATE INDEX IF NOT EXISTS idx_raw_meter_data_org ON raw_meter_data(org_id);

-- raw_mp_data
ALTER TABLE raw_mp_data ADD COLUMN IF NOT EXISTS org_id uuid REFERENCES organizations(id);
UPDATE raw_mp_data SET org_id = (SELECT id FROM organizations WHERE code='default') WHERE org_id IS NULL;
CREATE INDEX IF NOT EXISTS idx_raw_mp_data_org ON raw_mp_data(org_id);
