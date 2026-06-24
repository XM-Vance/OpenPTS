-- 文档智能解析(docling)接入 RBAC：新增 document_management 模块及其权限点,
-- 按既有分配规则授予系统角色(模式同 0015_auth_seed)。

-- ─── 模块 ───
INSERT INTO auth_modules (code, name, menu_group, route_paths, sort_order) VALUES
    ('document_management', '文档解析', '系统', ARRAY['/documents'], 145)
ON CONFLICT (code) DO NOTHING;

-- ─── 派生 read/write/delete 权限点 ───
INSERT INTO auth_permissions (code, name, module_code, action, permission_type)
SELECT
    m.code || ':' || a.action,
    m.name || ' - ' || a.label,
    m.code,
    a.action,
    CASE WHEN a.action = 'delete' THEN 'critical' ELSE 'normal' END
FROM auth_modules m
CROSS JOIN (VALUES
    ('read',   '查看'),
    ('write',  '编辑'),
    ('delete', '删除')
) AS a(action, label)
WHERE m.code = 'document_management'
ON CONFLICT (code) DO NOTHING;

-- ─── 角色授权(与 0015 一致：super_admin 全部 / admin read+write / analyst、viewer 只读) ───
INSERT INTO role_permissions (role_code, permission_code)
SELECT 'super_admin', code FROM auth_permissions WHERE module_code = 'document_management'
ON CONFLICT DO NOTHING;

INSERT INTO role_permissions (role_code, permission_code)
SELECT 'admin', code FROM auth_permissions
WHERE module_code = 'document_management' AND action IN ('read', 'write')
ON CONFLICT DO NOTHING;

INSERT INTO role_permissions (role_code, permission_code)
SELECT 'analyst', code FROM auth_permissions
WHERE module_code = 'document_management' AND action = 'read'
ON CONFLICT DO NOTHING;

INSERT INTO role_permissions (role_code, permission_code)
SELECT 'viewer', code FROM auth_permissions
WHERE module_code = 'document_management' AND action = 'read'
ON CONFLICT DO NOTHING;
