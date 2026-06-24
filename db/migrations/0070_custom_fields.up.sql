-- 70: 自定义字段定义表
CREATE TABLE custom_field_definitions (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id       UUID REFERENCES organizations(id),
    entity_type  VARCHAR(32) NOT NULL,  -- customer / contract / document / intent_customer
    field_key    VARCHAR(64) NOT NULL,  -- 英文 key，如 credit_rating
    field_label  VARCHAR(128) NOT NULL, -- 中文显示名，如"信用等级"
    field_type   VARCHAR(16) NOT NULL DEFAULT 'text', -- text/number/date/select/multiselect/textarea
    options      JSONB,                 -- select/multiselect 的选项列表
    default_value TEXT,
    is_required  BOOLEAN NOT NULL DEFAULT false,
    is_searchable BOOLEAN NOT NULL DEFAULT false,
    sort_order   INTEGER NOT NULL DEFAULT 0,
    created_by   UUID REFERENCES users(id),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(org_id, entity_type, field_key)
);

CREATE INDEX idx_cfd_org_entity ON custom_field_definitions(org_id, entity_type);
