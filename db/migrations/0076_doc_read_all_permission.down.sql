-- Reverse 0076: remove the doc read_all permission and its role assignments.
DELETE FROM role_permissions WHERE permission_code = 'document_management:read_all';
DELETE FROM auth_permissions WHERE code = 'document_management:read_all';
