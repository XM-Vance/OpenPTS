-- 81: 政策文件库 — 文档解析「确认入库」可把政策文件归纳为结构化条目，并关联来源解析文档
CREATE TABLE policy_documents (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id         UUID REFERENCES organizations(id),
    document_id    UUID REFERENCES documents(id) ON DELETE SET NULL,  -- 来源解析文档
    title          TEXT NOT NULL,                          -- 政策标题
    doc_no         TEXT,                                   -- 文号
    category       VARCHAR(32),                            -- 分类（市场规则/补贴/准入/其他）
    effective_date DATE,                                   -- 生效日期
    summary        TEXT,                                   -- 要点摘要
    source         VARCHAR(16) NOT NULL DEFAULT 'manual',  -- manual / document
    created_by     UUID REFERENCES users(id),
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_policy_documents_org ON policy_documents(org_id);
CREATE INDEX idx_policy_documents_effective ON policy_documents(effective_date DESC);
