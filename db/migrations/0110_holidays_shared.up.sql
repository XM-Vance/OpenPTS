-- 节假日表改为跨省共享参考表（P2-5）。
-- holidays 是政府法定假日 + 调休补班，全国统一，不应按省隔离。0055 误给它加了
-- org_id 并升级唯一键为 (org_id, holiday_date)，导致每省需各自维护一份相同数据。
-- 本迁移：去重保留 FJ 的数据 → 删 org_id 列 → 恢复 UNIQUE(holiday_date)。
-- 注意：多省已各自维护的情况下，回滚前应确认无差异；此处按 FJ 为准去重。

-- 1) 去重：同一天保留 org=FJ 的记录（FJ 是回填默认省，数据最全），删其他省的重复行
DELETE FROM holidays a USING holidays b
 WHERE a.holiday_date = b.holiday_date AND a.org_id <> b.org_id
   AND a.org_id <> (SELECT id FROM organizations WHERE code='default');

-- 2) 删 org_id 列（含外键、索引、唯一索引）
DROP INDEX IF EXISTS holidays_org_uniq;
DROP INDEX IF EXISTS idx_holidays_org;
ALTER TABLE holidays DROP COLUMN IF EXISTS org_id;

-- 3) 恢复单列唯一约束（全国共享，一天一条）
CREATE UNIQUE INDEX IF NOT EXISTS holidays_holiday_date_key ON holidays(holiday_date);
