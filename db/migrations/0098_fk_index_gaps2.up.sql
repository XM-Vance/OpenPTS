-- 98: 补齐外键缺索引（第二批）。
-- 审查报告 2.6 点名 14 个 FK 缺索引；第一批（0046/0047）已补 9 个，
-- 本迁移补齐剩余 5 个（均为 *_by 列引用 users，晚于索引补齐批次新增的表未回头补 FK 索引）。
-- 这些列在 JOIN/WHERE 时会全表扫描，随数据量增长性能下降。
--
-- 覆盖：
--   documents.uploaded_by              （text，引用 users.username/外部账号）
--   document_applies.applied_by        （text）
--   approval_requests.reviewed_by      （text）
--   custom_field_definitions.created_by（uuid，FK → users.id）
--   tag_definitions.created_by         （uuid，FK → users.id）
--
-- 说明：documents/document_applies/approval_requests 的 *_by 为 text（存用户名/外部 ID），
--   非 uuid FK 约束（见各自建表迁移），但同样需要索引加速「按操作人筛选」查询。

CREATE INDEX IF NOT EXISTS idx_documents_uploaded_by
    ON documents(uploaded_by);
CREATE INDEX IF NOT EXISTS idx_doc_applies_applied_by
    ON document_applies(applied_by);
CREATE INDEX IF NOT EXISTS idx_approval_requests_reviewed_by
    ON approval_requests(reviewed_by);
CREATE INDEX IF NOT EXISTS idx_cfd_created_by
    ON custom_field_definitions(created_by);
CREATE INDEX IF NOT EXISTS idx_tag_def_created_by
    ON tag_definitions(created_by);
