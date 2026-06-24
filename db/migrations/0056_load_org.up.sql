-- 任务 6: 负荷数据域 org_id
ALTER TABLE user_load_data ADD COLUMN IF NOT EXISTS org_id uuid REFERENCES organizations(id);
UPDATE user_load_data SET org_id = (SELECT id FROM organizations WHERE code='default') WHERE org_id IS NULL;
CREATE INDEX IF NOT EXISTS idx_user_load_data_org ON user_load_data(org_id);
ALTER TABLE user_load_data DROP CONSTRAINT IF EXISTS user_load_data_customer_id_date_key;
CREATE UNIQUE INDEX IF NOT EXISTS user_load_data_org_uniq ON user_load_data(org_id, customer_id, date);
