-- 0047: 补齐剩余无索引外键
-- 0046 已系统性补齐绝大多数外键/热点索引；本迁移仅补 0046 之后审计仍发现
-- 「有外键约束但首列无索引」的两列，避免父表删除/JOIN 时全表扫描。
-- 均使用 IF NOT EXISTS 保证幂等。

-- auth_email_challenges.user_id：按用户查邮箱验证挑战（鉴权流程 JOIN）
CREATE INDEX IF NOT EXISTS idx_auth_email_chl_user ON auth_email_challenges(user_id);

-- user_roles.granted_by：按授予人审计 + 用户删除时的 FK 反向检查
CREATE INDEX IF NOT EXISTS idx_user_roles_granted_by ON user_roles(granted_by) WHERE granted_by IS NOT NULL;
