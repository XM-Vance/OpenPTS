-- 0102: bonds 多租户隔离 + custom_field_values 实体值存储表
-- 背景：bonds 表此前无 org_id，跨省共享，违反多租户隔离铁律（与其他主数据不一致）。
--       custom_field 只有 definitions 表，缺实体值存储，导致自定义字段无法真正使用。

-- ─── bonds 加 org_id（仿 0101 样板）───
ALTER TABLE bonds ADD COLUMN IF NOT EXISTS org_id uuid REFERENCES organizations(id);
UPDATE bonds SET org_id = (SELECT id FROM organizations WHERE code='FJ') WHERE org_id IS NULL;
CREATE INDEX IF NOT EXISTS idx_bonds_org ON bonds(org_id);
-- 保函按省隔离：同省内 name 唯一，跨省可同名
CREATE UNIQUE INDEX IF NOT EXISTS bonds_org_name_uniq ON bonds(org_id, name);

-- ─── custom_field_values 实体值存储表 ───
-- 让 custom_field_definitions 定义的字段能真正写入到具体客户/合同等实体上。
CREATE TABLE IF NOT EXISTS custom_field_values (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id        UUID REFERENCES organizations(id),
    definition_id UUID NOT NULL REFERENCES custom_field_definitions(id) ON DELETE CASCADE,
    entity_id     UUID NOT NULL,          -- 客户/合同/文档等实体 id（按 definition.entity_type 解释）
    value         JSONB,                  -- 字段值（文本/数字/选项统一存 JSONB）
    created_by    UUID REFERENCES users(id),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(org_id, definition_id, entity_id)
);
CREATE INDEX IF NOT EXISTS idx_cfv_org_entity ON custom_field_values(org_id, entity_id);
CREATE INDEX IF NOT EXISTS idx_cfv_definition ON custom_field_values(definition_id);
