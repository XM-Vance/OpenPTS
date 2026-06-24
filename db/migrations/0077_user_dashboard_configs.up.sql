-- 0077: 用户仪表盘配置表 — 轻量级个性化定制
-- 每个用户保存自己的 widget 布局（显示/隐藏 + 排列顺序）

CREATE TABLE IF NOT EXISTS user_dashboard_configs (
    user_id   UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    config    JSONB NOT NULL DEFAULT '{}'::jsonb,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
