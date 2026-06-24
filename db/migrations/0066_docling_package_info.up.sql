-- 66: 增加 package_info 字段（套餐结构化信息）
ALTER TABLE docling_documents ADD COLUMN IF NOT EXISTS package_info JSONB;
COMMENT ON COLUMN docling_documents.package_info IS '套餐结构化信息：套餐名/类型/计价方式/浮动比例/绿电比例等';
