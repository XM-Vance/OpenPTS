-- 73: 交易规则配置
CREATE TABLE trade_rules (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id         UUID REFERENCES organizations(id),
    rule_key       VARCHAR(64) NOT NULL,    -- deviation_penalty_rate
    rule_value     TEXT NOT NULL,            -- 5%
    rule_category  VARCHAR(32),              -- settlement/deviation/green/registration
    description    TEXT,
    effective_date DATE,
    expiry_date    DATE,
    source_doc_id  UUID,                     -- 关联解析的政策文档
    is_active      BOOLEAN NOT NULL DEFAULT true,
    created_by     UUID REFERENCES users(id),
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(org_id, rule_key, effective_date)
);

CREATE INDEX idx_trade_rules_org_cat ON trade_rules(org_id, rule_category);
CREATE INDEX idx_trade_rules_active ON trade_rules(is_active, effective_date);
