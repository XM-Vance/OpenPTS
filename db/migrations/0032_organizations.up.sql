-- 多租户骨架：组织表 + users.org_id 列。
-- 默认插入 default 组织；现有用户全部归属 default，向后兼容。
-- 业务表（customers / retail_contracts 等）暂不强制 org_id；未来按需逐域引入。

CREATE TABLE IF NOT EXISTS organizations (
    id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    code         text NOT NULL UNIQUE,
    name         text NOT NULL,
    is_active    boolean NOT NULL DEFAULT true,
    created_at   timestamp NOT NULL DEFAULT now()
);

INSERT INTO organizations (code, name) VALUES ('default', '默认组织')
ON CONFLICT (code) DO NOTHING;

-- users 加 org_id；现有用户回填 default
ALTER TABLE users ADD COLUMN IF NOT EXISTS org_id uuid REFERENCES organizations(id);

UPDATE users
SET org_id = (SELECT id FROM organizations WHERE code = 'default')
WHERE org_id IS NULL;

CREATE INDEX IF NOT EXISTS idx_users_org ON users(org_id);
