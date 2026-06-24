-- 71: 统一标签定义表
CREATE TABLE tag_definitions (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id      UUID REFERENCES organizations(id),
    name        VARCHAR(64) NOT NULL,
    color       VARCHAR(7) DEFAULT '#3B82F6',
    entity_type VARCHAR(32) NOT NULL DEFAULT 'customer', -- customer/contract/document/intent_customer
    is_active   BOOLEAN NOT NULL DEFAULT true,
    sort_order  INTEGER NOT NULL DEFAULT 0,
    created_by  UUID REFERENCES users(id),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(org_id, name, entity_type)
);
