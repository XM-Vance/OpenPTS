-- 0118 down: 恢复「保函管理」表结构与菜单种子（仅结构回退，数据不恢复）。
-- 表结构 = 0040 原始列 + 0102 的 org_id 列 + 0093 的 amount numeric(18,4) 精度。
-- ⚠️ 历史保函数据在 up 时已 DROP，无法恢复；此 down 仅用于一次性测试库的结构回退。

-- 1) 重建表结构
CREATE TABLE IF NOT EXISTS bonds (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(200) NOT NULL,
    bond_type VARCHAR(50) NOT NULL DEFAULT '',
    amount numeric(18,4) NOT NULL DEFAULT 0,
    issuer VARCHAR(200) NOT NULL DEFAULT '',
    beneficiary VARCHAR(200) NOT NULL DEFAULT '',
    issue_date DATE,
    expire_date DATE,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    description TEXT NOT NULL DEFAULT '',
    created_by UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    org_id uuid REFERENCES organizations(id)
);

-- 0040 索引
CREATE INDEX IF NOT EXISTS idx_bonds_status ON bonds (status);
CREATE INDEX IF NOT EXISTS idx_bonds_expire_date ON bonds (expire_date);
-- 0046 性能索引
CREATE INDEX IF NOT EXISTS idx_bonds_created_by ON bonds(created_by);
CREATE INDEX IF NOT EXISTS idx_bonds_expire_status ON bonds(status, expire_date) WHERE status = 'active';
-- 0102 多租户索引
CREATE INDEX IF NOT EXISTS idx_bonds_org ON bonds(org_id);
CREATE UNIQUE INDEX IF NOT EXISTS bonds_org_name_uniq ON bonds(org_id, name);

-- 2) 恢复菜单种子（与 0080 中原始定义一致）
INSERT INTO menu_pages (code, label, href, icon, sort_order, group_name, is_required) VALUES
('page:bonds', '保函管理', '/bonds', 'FileKey', 17, '客户管理', FALSE)
ON CONFLICT (code) DO NOTHING;
