-- 0047: 回滚剩余无索引外键的补齐
DROP INDEX IF EXISTS idx_auth_email_chl_user;
DROP INDEX IF EXISTS idx_user_roles_granted_by;
