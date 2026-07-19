-- 129: 用户 API Key —— 每个用户可生成自己的 API Key，用于 MCP/外部脚本以本人身份连 ptis。
--
-- 设计要点：
-- - 明文 Key 只在创建时返回一次；库存 bcrypt 哈希（与密码同算法，但独立存储）。
-- - key_prefix（前 12 字符明文）用于展示和按前缀反查，避免全表 bcrypt 比对。
-- - 支持多 Key、吊销（revoked_at）、过期（expires_at）、最后使用时间（last_used_at）。
-- - 吊销后通过 exchange 端点换 JWT 会 401，MCP/脚本连接自然失效。
--
-- 与 users 的关系：user_id 外键；用户删则级联删 Key（和该用户的会话等一致）。
-- 共享/部署级表（不加 org_id）：Key 归属用户，权限由用户的 RBAC 权限码决定。

CREATE TABLE IF NOT EXISTS user_api_keys (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name         TEXT NOT NULL DEFAULT 'default',   -- 用户起的名，如「我的Cursor」
    key_prefix   VARCHAR(20) NOT NULL,               -- 明文前缀，用于展示+反查（如 ptis_abc12345）
    key_hash     TEXT NOT NULL,                       -- 完整 Key 的 bcrypt 哈希
    last_used_at TIMESTAMPTZ,
    expires_at   TIMESTAMPTZ,                         -- 可空 = 永不过期
    revoked_at   TIMESTAMPTZ,                         -- 可空 = 未吊销
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_user_api_keys_user   ON user_api_keys(user_id);
CREATE INDEX IF NOT EXISTS idx_user_api_keys_prefix ON user_api_keys(key_prefix);
