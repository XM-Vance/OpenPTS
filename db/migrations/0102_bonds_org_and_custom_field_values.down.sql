-- 0102 down: 回滚 bonds org_id 与 custom_field_values 表
-- ⚠️ 仅在一次性验证库执行；生产/ptis_dev 禁止（会丢数据）。

DROP TABLE IF EXISTS custom_field_values;

DROP INDEX IF EXISTS bonds_org_name_uniq;
DROP INDEX IF EXISTS idx_bonds_org;
-- 注意：回填到 FJ 的 org_id 无法精准还原为 NULL（无法区分原有与回填），保守不删列。
-- 如需彻底回滚，可手动：ALTER TABLE bonds DROP COLUMN org_id;
