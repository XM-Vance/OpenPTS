-- 98 回滚：删除第二批补齐的外键索引。

DROP INDEX IF EXISTS idx_tag_def_created_by;
DROP INDEX IF EXISTS idx_cfd_created_by;
DROP INDEX IF EXISTS idx_approval_requests_reviewed_by;
DROP INDEX IF EXISTS idx_doc_applies_applied_by;
DROP INDEX IF EXISTS idx_documents_uploaded_by;
