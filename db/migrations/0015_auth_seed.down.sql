-- 回滚仅清除本迁移插入的种子数据；不删除表结构（结构在 0002 管理）。
DELETE FROM role_permissions WHERE role_code IN ('super_admin', 'admin', 'analyst', 'viewer');
DELETE FROM auth_roles       WHERE code IN ('super_admin', 'admin', 'analyst', 'viewer');
DELETE FROM auth_permissions;       -- 0015 是 auth_permissions 唯一种子来源
DELETE FROM auth_modules;           -- 0015 是 auth_modules 唯一种子来源
