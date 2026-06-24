-- 回滚文档解析管线。
-- docling_documents 原表自始未动，历史数据仍在原表，直接删除新表即可。
DROP TABLE IF EXISTS document_applies;
DROP TABLE IF EXISTS document_extractions;
DROP TABLE IF EXISTS documents;
