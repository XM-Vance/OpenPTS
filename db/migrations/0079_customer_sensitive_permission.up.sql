-- 迁移 #79: 客户档案字段级权限 — 仅授权用户可查看客户名称/简称/客户经理/地址等敏感信息
-- 无此权限的用户看到的字段会自动脱敏（如 "某***公司"、"张***"、"东***路"）

INSERT INTO auth_permissions (code, name, module_code, action, permission_type)
VALUES (
    'customer_management:view_sensitive',
    '查看客户敏感信息',
    'customer_management',
    'view_sensitive',
    'data'
)
ON CONFLICT (code) DO NOTHING;

-- super_admin / admin 默认拥有此权限
INSERT INTO role_permissions (role_code, permission_code)
VALUES
    ('super_admin', 'customer_management:view_sensitive'),
    ('admin', 'customer_management:view_sensitive')
ON CONFLICT (role_code, permission_code) DO NOTHING;
