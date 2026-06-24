-- 72: 审批增加实体关联
ALTER TABLE approval_requests ADD COLUMN IF NOT EXISTS entity_type VARCHAR(32);
ALTER TABLE approval_requests ADD COLUMN IF NOT EXISTS entity_id UUID;

CREATE INDEX IF NOT EXISTS idx_approval_entity ON approval_requests(entity_type, entity_id);
