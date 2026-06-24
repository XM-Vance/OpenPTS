-- 0083: Hermes 等自动化账号专用的最小权限角色 agent（只读起步）。
-- 写权限（如 document_management:write）不在此默认授予；需要归档/写入时请在
-- 「角色管理」中按需为 agent 开启，遵循最小权限 + 写操作可审计/可人工确认的原则。

INSERT INTO auth_roles (code, name, description, is_system)
VALUES ('agent', 'Agent（自动化）',
        '供 Hermes 等自动化账号使用的最小权限角色：默认只读；需要归档/写入时在角色管理中按需开启写权限。',
        false)
ON CONFLICT (code) DO NOTHING;

-- 只授予存在的只读权限码（用 SELECT 过滤，避免授予不存在的码导致外键失败）
INSERT INTO role_permissions (role_code, permission_code)
SELECT 'agent', code FROM auth_permissions
WHERE code IN (
  'document_management:read',
  'customer_management:read',
  'retail_management:read',
  'settlement_management:read',
  'price_management:read',
  'analytics:read'
)
ON CONFLICT (role_code, permission_code) DO NOTHING;
