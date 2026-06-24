-- 系统配置：全局参数中心，业务模块按 key 读取。

CREATE TABLE IF NOT EXISTS system_settings (
    key          text PRIMARY KEY,
    value        text NOT NULL,
    value_type   text NOT NULL DEFAULT 'string',  -- string / number / bool / json
    category     text NOT NULL DEFAULT 'general',
    description  text,
    is_editable  boolean NOT NULL DEFAULT true,
    is_sensitive boolean NOT NULL DEFAULT false,  -- 敏感值返回时掩码
    updated_by   text,
    created_at   timestamp NOT NULL DEFAULT now(),
    updated_at   timestamp NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_system_settings_category
    ON system_settings(category);

-- 预置常用配置
INSERT INTO system_settings (key, value, value_type, category, description) VALUES
    ('app.name',                    '电力交易信息系统',     'string', 'general',   '应用名称'),
    ('app.version',                 'v2.0.0',                  'string', 'general',   '版本号'),
    ('jwt.ttl_hours',               '24',                      'number', 'security',  'JWT 有效期（小时）'),
    ('cache.dashboard_ttl_seconds', '30',                      'number', 'cache',     '仪表盘 KPI 缓存 TTL（秒）'),
    ('cache.series_ttl_seconds',    '300',                     'number', 'cache',     '时间序列缓存 TTL（秒）'),
    ('alert.pending_threshold',     '50',                      'number', 'alert',     '待处理告警堆积阈值'),
    ('alert.critical_threshold',    '10',                      'number', 'alert',     '严重告警堆积阈值'),
    ('audit.retention_days',        '180',                     'number', 'audit',     '审计日志保留天数'),
    ('feature.enable_pdf_generation','true',                    'bool',   'feature',   '启用合同 PDF 生成'),
    ('feature.enable_attachments',  'true',                     'bool',   'feature',   '启用附件上传'),
    ('feature.enable_approvals',    'true',                     'bool',   'feature',   '启用审批流'),
    ('feature.enable_websocket',    'true',                     'bool',   'feature',   '启用 WebSocket'),
    ('mask.default_for_demo',       'true',                     'bool',   'privacy',   '演示客户默认脱敏')
ON CONFLICT (key) DO NOTHING;
