-- 审计日志：所有写操作（POST/PUT/DELETE）通过中间件自动落库。
-- 复用 0015 已建的 system 模块与 system:read/write 权限点。

CREATE TABLE IF NOT EXISTS audit_logs (
    id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id       uuid,
    username      text,
    method        text NOT NULL,            -- POST / PUT / DELETE
    path          text NOT NULL,            -- /api/v1/customers/:id
    resource      text,                     -- customers / retail_contracts ...
    resource_id   text,                     -- 路径参数 :id（如有）
    status_code   integer NOT NULL,
    ip            inet,
    user_agent    text,
    duration_ms   integer NOT NULL DEFAULT 0,
    error_message text,                     -- 失败时记录错误摘要
    created_at    timestamp NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_audit_logs_created
    ON audit_logs(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_logs_user_created
    ON audit_logs(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_logs_resource
    ON audit_logs(resource, created_at DESC);
