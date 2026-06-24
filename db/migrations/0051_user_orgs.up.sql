-- 0051: 用户↔组织多对多（支撑「一人管多租户」）+ 总部标记 + 存量回填默认组织。

-- 用户可访问的组织（多对多）。users.org_id 仍作为「主/默认活跃组织」。
CREATE TABLE IF NOT EXISTS user_orgs (
    user_id    uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    org_id     uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    created_at timestamp NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, org_id)
);
CREATE INDEX IF NOT EXISTS idx_user_orgs_org ON user_orgs(org_id);

-- 总部标记：is_hq 用户可访问/切换全部组织。
ALTER TABLE users ADD COLUMN IF NOT EXISTS is_hq boolean NOT NULL DEFAULT false;

-- 存量回填：现有用户主组织 = 默认组织（org_id 为空的）。
UPDATE users
   SET org_id = (SELECT id FROM organizations WHERE code = 'default')
 WHERE org_id IS NULL;

-- 现有用户加入 user_orgs（默认组织）。
INSERT INTO user_orgs (user_id, org_id)
SELECT u.id, (SELECT id FROM organizations WHERE code = 'default')
  FROM users u
ON CONFLICT DO NOTHING;

-- 让现有 super_admin 成为总部（开箱即有一个能看全部组织的人）。
UPDATE users SET is_hq = true
 WHERE id IN (SELECT user_id FROM user_roles WHERE role_code = 'super_admin');
