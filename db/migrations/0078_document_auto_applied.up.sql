-- 文档自动入库标记：解析后置信度达标时自动写入业务表，标记此字段。
ALTER TABLE documents ADD COLUMN IF NOT EXISTS auto_applied boolean NOT NULL DEFAULT false;

COMMENT ON COLUMN documents.auto_applied IS 'true=系统自动入库（高置信度），false=需人工确认入库';
