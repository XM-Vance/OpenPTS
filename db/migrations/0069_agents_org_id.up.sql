-- 69: agents 表增加 org_id 多租户隔离
ALTER TABLE agents ADD COLUMN IF NOT EXISTS org_id UUID REFERENCES organizations(id);

-- 为现有 agents 数据填充默认 org_id（取第一个组织）
UPDATE agents SET org_id = (SELECT id FROM organizations ORDER BY created_at LIMIT 1) WHERE org_id IS NULL;

CREATE INDEX IF NOT EXISTS idx_agents_org_id ON agents(org_id);
