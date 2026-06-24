-- customers 加 org_id（指向 organizations），现有数据回填到 default 组织。
ALTER TABLE customers ADD COLUMN IF NOT EXISTS org_id uuid REFERENCES organizations(id);

UPDATE customers
SET org_id = (SELECT id FROM organizations WHERE code = 'default')
WHERE org_id IS NULL;

CREATE INDEX IF NOT EXISTS idx_customers_org ON customers(org_id);
