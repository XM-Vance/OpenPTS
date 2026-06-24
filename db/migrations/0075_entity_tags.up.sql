-- entity_tags: 实体-标签关联表
CREATE TABLE IF NOT EXISTS entity_tags (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    entity_type VARCHAR(50)  NOT NULL,   -- customer / contract / document / agent …
    entity_id   UUID         NOT NULL,
    tag_id      UUID         NOT NULL REFERENCES tag_definitions(id) ON DELETE CASCADE,
    created_by  UUID,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    UNIQUE (entity_type, entity_id, tag_id)
);

CREATE INDEX idx_entity_tags_entity ON entity_tags (entity_type, entity_id);
CREATE INDEX idx_entity_tags_tag    ON entity_tags (tag_id);
