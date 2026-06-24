-- 审计日志读路径优化：高频查询是「按用户/方法/资源 + 时间倒序 LIMIT N」。
-- 加复合 covering 索引；status_code 失败筛选加偏索引。

CREATE INDEX IF NOT EXISTS idx_audit_logs_user_method_time
    ON audit_logs(user_id, method, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_audit_logs_resource_time
    ON audit_logs(resource, created_at DESC)
    WHERE resource IS NOT NULL;

-- 仅对失败请求（4xx/5xx）建偏索引，安全大屏热点查询用
CREATE INDEX IF NOT EXISTS idx_audit_logs_errors
    ON audit_logs(created_at DESC, status_code)
    WHERE status_code >= 400;

-- 仅对 DELETE 请求（敏感操作）建偏索引
CREATE INDEX IF NOT EXISTS idx_audit_logs_delete
    ON audit_logs(created_at DESC)
    WHERE method = 'DELETE';
