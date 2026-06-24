-- 回滚 #79
DELETE FROM role_permissions
WHERE permission_code = 'customer_management:view_sensitive';

DELETE FROM auth_permissions WHERE code = 'customer_management:view_sensitive';
