-- 0002: 权限系统（RBAC + 安全增强）
-- 覆盖 v1 集合：auth_modules / auth_permissions / auth_roles / user_roles（隐式）
--              / role_permissions（隐式）/ auth_sessions
--              / auth_security_challenges / auth_email_challenges / auth_trusted_devices

-- ─── users 表补充字段 ──────────────────────────────
ALTER TABLE users ADD COLUMN IF NOT EXISTS email VARCHAR(128);
ALTER TABLE users ADD COLUMN IF NOT EXISTS phone VARCHAR(32);
ALTER TABLE users ADD COLUMN IF NOT EXISTS last_login_at TIMESTAMPTZ;
ALTER TABLE users ADD COLUMN IF NOT EXISTS last_login_ip INET;

CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);

-- ─── 权限模块（菜单分组）───────────────────────────
CREATE TABLE IF NOT EXISTS auth_modules (
    code         VARCHAR(64) PRIMARY KEY,
    name         VARCHAR(128) NOT NULL,
    menu_group   VARCHAR(64),
    route_paths  TEXT[] NOT NULL DEFAULT '{}',
    sort_order   INT NOT NULL DEFAULT 0,
    is_active    BOOLEAN NOT NULL DEFAULT TRUE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

DROP TRIGGER IF EXISTS auth_modules_set_updated_at ON auth_modules;
CREATE TRIGGER auth_modules_set_updated_at
    BEFORE UPDATE ON auth_modules
    FOR EACH ROW EXECUTE FUNCTION trg_set_updated_at();

-- ─── 权限点 ────────────────────────────────────────
CREATE TABLE IF NOT EXISTS auth_permissions (
    code             VARCHAR(128) PRIMARY KEY,
    name             VARCHAR(128) NOT NULL,
    module_code      VARCHAR(64) NOT NULL REFERENCES auth_modules(code) ON DELETE CASCADE,
    action           VARCHAR(32) NOT NULL,                       -- read / write / delete / export / approve
    permission_type  VARCHAR(32) NOT NULL DEFAULT 'normal',      -- normal / critical
    is_active        BOOLEAN NOT NULL DEFAULT TRUE,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_auth_perm_module ON auth_permissions(module_code);
CREATE INDEX IF NOT EXISTS idx_auth_perm_type ON auth_permissions(permission_type);

DROP TRIGGER IF EXISTS auth_permissions_set_updated_at ON auth_permissions;
CREATE TRIGGER auth_permissions_set_updated_at
    BEFORE UPDATE ON auth_permissions
    FOR EACH ROW EXECUTE FUNCTION trg_set_updated_at();

-- ─── 角色 ──────────────────────────────────────────
CREATE TABLE IF NOT EXISTS auth_roles (
    code        VARCHAR(64) PRIMARY KEY,
    name        VARCHAR(128) NOT NULL,
    description TEXT,
    is_system   BOOLEAN NOT NULL DEFAULT FALSE,
    is_active   BOOLEAN NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

DROP TRIGGER IF EXISTS auth_roles_set_updated_at ON auth_roles;
CREATE TRIGGER auth_roles_set_updated_at
    BEFORE UPDATE ON auth_roles
    FOR EACH ROW EXECUTE FUNCTION trg_set_updated_at();

-- ─── 用户 ↔ 角色（多对多）──────────────────────────
CREATE TABLE IF NOT EXISTS user_roles (
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_code   VARCHAR(64) NOT NULL REFERENCES auth_roles(code) ON DELETE CASCADE,
    granted_by  UUID REFERENCES users(id),
    granted_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, role_code)
);

CREATE INDEX IF NOT EXISTS idx_user_roles_user ON user_roles(user_id);
CREATE INDEX IF NOT EXISTS idx_user_roles_role ON user_roles(role_code);

-- ─── 角色 ↔ 权限（多对多）──────────────────────────
CREATE TABLE IF NOT EXISTS role_permissions (
    role_code        VARCHAR(64) NOT NULL REFERENCES auth_roles(code) ON DELETE CASCADE,
    permission_code  VARCHAR(128) NOT NULL REFERENCES auth_permissions(code) ON DELETE CASCADE,
    granted_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (role_code, permission_code)
);

CREATE INDEX IF NOT EXISTS idx_role_perms_perm ON role_permissions(permission_code);

-- ─── 会话（用于 token 撤销）────────────────────────
CREATE TABLE IF NOT EXISTS auth_sessions (
    id             UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id        UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    session_token  VARCHAR(255) NOT NULL UNIQUE,
    ip             INET,
    user_agent     TEXT,
    device_id      VARCHAR(128),
    expires_at     TIMESTAMPTZ NOT NULL,
    revoked_at     TIMESTAMPTZ,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_sessions_user ON auth_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_token ON auth_sessions(session_token);
CREATE INDEX IF NOT EXISTS idx_sessions_expires ON auth_sessions(expires_at);

-- ─── 安全挑战（异常登录二次验证）───────────────────
CREATE TABLE IF NOT EXISTS auth_security_challenges (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    challenge_type  VARCHAR(32) NOT NULL,                        -- captcha / sms / email / totp
    status          VARCHAR(32) NOT NULL DEFAULT 'pending',      -- pending / passed / failed / expired
    payload         JSONB NOT NULL DEFAULT '{}',
    expires_at      TIMESTAMPTZ NOT NULL,
    completed_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_sec_chl_user ON auth_security_challenges(user_id);
CREATE INDEX IF NOT EXISTS idx_sec_chl_status ON auth_security_challenges(status);

-- ─── 邮件验证码 ────────────────────────────────────
CREATE TABLE IF NOT EXISTS auth_email_challenges (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id     UUID REFERENCES users(id) ON DELETE CASCADE,
    email       VARCHAR(128) NOT NULL,
    code        VARCHAR(16) NOT NULL,
    purpose     VARCHAR(32) NOT NULL,                            -- reset_password / verify_email / bind_email
    used_at     TIMESTAMPTZ,
    expires_at  TIMESTAMPTZ NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_email_chl_email ON auth_email_challenges(email);
CREATE INDEX IF NOT EXISTS idx_email_chl_expires ON auth_email_challenges(expires_at);

-- ─── 可信设备 ──────────────────────────────────────
CREATE TABLE IF NOT EXISTS auth_trusted_devices (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    device_id       VARCHAR(128) NOT NULL,
    device_name     VARCHAR(128),
    last_login_at   TIMESTAMPTZ,
    last_login_ip   INET,
    expires_at      TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, device_id)
);

CREATE INDEX IF NOT EXISTS idx_trusted_dev_user ON auth_trusted_devices(user_id);
