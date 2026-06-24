DROP TABLE IF EXISTS auth_trusted_devices;
DROP TABLE IF EXISTS auth_email_challenges;
DROP TABLE IF EXISTS auth_security_challenges;
DROP TABLE IF EXISTS auth_sessions;
DROP TABLE IF EXISTS role_permissions;
DROP TABLE IF EXISTS user_roles;

DROP TRIGGER IF EXISTS auth_roles_set_updated_at ON auth_roles;
DROP TABLE IF EXISTS auth_roles;

DROP TRIGGER IF EXISTS auth_permissions_set_updated_at ON auth_permissions;
DROP TABLE IF EXISTS auth_permissions;

DROP TRIGGER IF EXISTS auth_modules_set_updated_at ON auth_modules;
DROP TABLE IF EXISTS auth_modules;

ALTER TABLE users DROP COLUMN IF EXISTS last_login_ip;
ALTER TABLE users DROP COLUMN IF EXISTS last_login_at;
ALTER TABLE users DROP COLUMN IF EXISTS phone;
ALTER TABLE users DROP COLUMN IF EXISTS email;
