-- 文档解析管线重设计：原件/解析件双存档 + 结构化提取 + 人工确认入库。
--
-- documents            文档主表（状态机 uploaded→parsing→parsed/failed）
-- document_extractions 结构化提取字段（合同甲乙方/电量电价、Excel 行数据……带类型与置信度）
-- document_applies     「确认入库」审计（哪个文档、写进哪张业务表、多少行、谁、何时）
--
-- 旧表 docling_documents 保留不动（独立微服务 standalone 模式仍可写），
-- 其历史数据平移进 documents（无原件，status 直接 parsed）。

CREATE TABLE IF NOT EXISTS documents (
    id                  uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id              uuid,                              -- 归属省份；上传时取活跃省
    filename            VARCHAR(512) NOT NULL,
    content_type        TEXT NOT NULL DEFAULT 'application/octet-stream',
    size                BIGINT NOT NULL DEFAULT 0,
    sha256              TEXT,                              -- 原件指纹，用于去重
    source_kind         VARCHAR(16) NOT NULL DEFAULT 'pdf',-- pdf/image/word/excel/csv
    original_object_key TEXT,                              -- MinIO 原件 key（历史平移行为 NULL）
    parsed_object_key   TEXT,                              -- MinIO 解析件(.md) key
    doc_type            VARCHAR(32),                       -- 合同/政策/规则/账单/资质/客户清单/负荷数据/结算单/其他
    status              VARCHAR(16) NOT NULL DEFAULT 'uploaded', -- uploaded/parsing/parsed/failed
    page_count          INTEGER NOT NULL DEFAULT 0,
    text_content        TEXT,                              -- 全文 markdown（同时落解析件文件）
    tables              JSONB,
    entities            JSONB,                             -- 正则实体（兜底展示）
    summary             TEXT,
    error               TEXT,                              -- status=failed 时的原因
    uploaded_by         TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_documents_org    ON documents(org_id);
CREATE INDEX IF NOT EXISTS idx_documents_status ON documents(status);
CREATE INDEX IF NOT EXISTS idx_documents_type   ON documents(doc_type);
CREATE INDEX IF NOT EXISTS idx_documents_sha    ON documents(sha256);

CREATE TABLE IF NOT EXISTS document_extractions (
    id          BIGSERIAL PRIMARY KEY,
    document_id uuid NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    group_no    INTEGER NOT NULL DEFAULT 0,  -- 行组：0=文档级字段；Excel/明细按行 1..n
    field_key   TEXT NOT NULL,               -- party_a / energy_mwh / customer_name ...
    field_label TEXT NOT NULL,               -- 甲方 / 合同电量 / 客户名称 ...
    value_text  TEXT,
    value_num   DOUBLE PRECISION,            -- 数值字段冗余（金额/电量/电价）
    value_date  DATE,                        -- 日期字段冗余
    unit        TEXT,
    confidence  REAL,                        -- 0~1；excel 直读=1，GLM 提取按模型给
    source      VARCHAR(16) NOT NULL DEFAULT 'glm',  -- glm/regex/excel
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_doc_extractions_doc ON document_extractions(document_id);

CREATE TABLE IF NOT EXISTS document_applies (
    id           BIGSERIAL PRIMARY KEY,
    document_id  uuid NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    target       TEXT NOT NULL,              -- customers / user_load_data / batch_monthly_settlement
    applied_rows INTEGER NOT NULL DEFAULT 0,
    detail       JSONB,                      -- 行级结果（成功/跳过原因）
    applied_by   TEXT,
    applied_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_doc_applies_doc ON document_applies(document_id);

-- 历史数据平移：docling_documents → documents（无原件，标记已解析）
INSERT INTO documents (org_id, filename, doc_type, status, source_kind,
                       text_content, tables, entities, summary, size, created_at, updated_at)
SELECT org_id, filename, doc_type, 'parsed', 'pdf',
       text_content, tables, entities, summary, COALESCE(file_size, 0), created_at, updated_at
FROM docling_documents;
