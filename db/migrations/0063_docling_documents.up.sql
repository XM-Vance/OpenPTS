-- 文档智能解析结果表（docling-service 写入）
-- 上传文档经 GLM 视觉 OCR → 结构化抽取后落库。由独立的 docling 微服务写入,
-- 此处纳入统一迁移流作为 schema 唯一来源(服务侧 _ensure_table 仅作 standalone 兜底,
-- 同样 IF NOT EXISTS、同构)。
--
-- org_id 为 uuid,与全局多租户口径一致(早期服务用 VARCHAR 'FJ',已对齐为 uuid)。
-- 不设 FK:docling 为松耦合微服务、可独立于主库迁移启动,避免跨服务外键的启动期脆弱;
-- NULL 表示未归属省份(standalone/历史数据),消费方按 org 过滤时据此排除。
CREATE TABLE IF NOT EXISTS docling_documents (
    id           SERIAL PRIMARY KEY,
    org_id       uuid,                      -- 归属省份(组织),消费方据此做租户隔离
    filename     VARCHAR(512) NOT NULL,
    doc_type     VARCHAR(32),               -- 合同/政策/规则/账单/资质/其他
    text_content TEXT,                      -- 全文 markdown
    tables       JSONB,                     -- 抽取的表格
    entities     JSONB,                     -- 关键实体(金额/电价/电量/企业/日期)
    summary      TEXT,                      -- 摘要
    file_size    INTEGER,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_docling_doc_type ON docling_documents(doc_type);
CREATE INDEX IF NOT EXISTS idx_docling_org_id ON docling_documents(org_id);
