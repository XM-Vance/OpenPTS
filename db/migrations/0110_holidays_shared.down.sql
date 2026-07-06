-- 回滚 0110：恢复 holidays 的 org_id 隔离。
-- ⚠️ 仅一次性测试库用：回滚后所有现存假日归到 FJ（其他省的记录在 up 的去重步骤已删，
-- 无法恢复）。生产/共享库不建议 down。
ALTER TABLE holidays ADD COLUMN IF NOT EXISTS org_id uuid REFERENCES organizations(id);
UPDATE holidays SET org_id = (SELECT id FROM organizations WHERE code='default') WHERE org_id IS NULL;
CREATE INDEX IF NOT EXISTS idx_holidays_org ON holidays(org_id);
DROP INDEX IF EXISTS holidays_holiday_date_key;
CREATE UNIQUE INDEX IF NOT EXISTS holidays_org_uniq ON holidays(org_id, holiday_date);
