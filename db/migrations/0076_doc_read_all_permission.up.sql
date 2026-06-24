-- 0076: 新增文档查看全部权限 — 用于按角色控制文档可见范围
-- super_admin / admin / analyst 角色自动获得该权限（可看同省所有文档）
-- viewer 角色没有该权限（仅看自己上传的）

INSERT INTO auth_permissions (code, name, module_code, action, permission_type)
VALUES ('document_management:read_all', '文档查看全部', 'document_management', 'read_all', 'normal')
ON CONFLICT (code) DO NOTHING;

INSERT INTO role_permissions (role_code, permission_code)
SELECT r.code, 'document_management:read_all'
FROM auth_roles r
WHERE r.code IN ('super_admin', 'admin', 'analyst')
ON CONFLICT DO NOTHING;
