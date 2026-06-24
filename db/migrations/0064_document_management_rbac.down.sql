-- 回滚:先删授权,再删权限点,最后删模块。
DELETE FROM role_permissions
WHERE permission_code IN (
    SELECT code FROM auth_permissions WHERE module_code = 'document_management'
);
DELETE FROM auth_permissions WHERE module_code = 'document_management';
DELETE FROM auth_modules WHERE code = 'document_management';
