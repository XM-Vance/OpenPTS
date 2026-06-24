-- 文件附件：关联到任意资源（合同 / 客户 / 月度结算 等）。
-- object_key 存对象存储中的 key（如 minio bucket 内的路径），content_type / size 缓存元信息。

CREATE TABLE IF NOT EXISTS file_attachments (
    id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    resource     text NOT NULL,                  -- customers / retail_contracts / monthly_settlement ...
    resource_id  text NOT NULL,                  -- 关联资源主键（uuid 或字符串）
    filename     text NOT NULL,                  -- 原始文件名
    object_key   text NOT NULL UNIQUE,           -- 对象存储 key
    content_type text NOT NULL DEFAULT 'application/octet-stream',
    size         bigint NOT NULL DEFAULT 0,
    uploaded_by  text,                           -- 上传人 username
    note         text,
    created_at   timestamp NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_file_attachments_resource
    ON file_attachments(resource, resource_id, created_at DESC);
