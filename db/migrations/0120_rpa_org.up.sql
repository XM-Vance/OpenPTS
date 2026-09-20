-- 132: RPA 任务/运行记录加 org_id 多租户隔离。
--
-- 背景：0023 建表时未加 org_id，导致跨省用户能看到彼此的 RPA 任务和运行记录，
-- demo 数据也跨省污染。违反 AGENTS.md §4 多租户铁律。
-- 改造参照 0056_load_org / 0106_org_prices 的模式：
--   1) 加 org_id 列（REFERENCES organizations）；
--   2) 存量数据回填到默认省 FJ（与 0056/0106 一致）；
--   3) rpa_jobs 唯一键由 (name) 改为 (org_id, name)；
--   4) 加索引。
-- down 可逆（仅结构回退，删 org_id 列会丢省归属但 RPA 数据本身保留）。

-- ─── rpa_jobs ───
ALTER TABLE rpa_jobs ADD COLUMN IF NOT EXISTS org_id uuid REFERENCES organizations(id);

-- 存量数据回填到默认省 FJ（与 0056_load_org / 0106_org_prices 一致）
UPDATE rpa_jobs SET org_id = (SELECT id FROM organizations WHERE code='FJ') WHERE org_id IS NULL;

CREATE INDEX IF NOT EXISTS idx_rpa_jobs_org ON rpa_jobs(org_id);

-- 唯一键 (name) → (org_id, name)：允许不同省同名任务
ALTER TABLE rpa_jobs DROP CONSTRAINT IF EXISTS rpa_jobs_name_key;
CREATE UNIQUE INDEX IF NOT EXISTS rpa_jobs_org_name_uniq ON rpa_jobs(org_id, name);

-- ─── rpa_runs（无独立唯一键，只加 org_id 列 + 索引，JOIN 时按 rpa_jobs.org_id 过滤）───
ALTER TABLE rpa_runs ADD COLUMN IF NOT EXISTS org_id uuid REFERENCES organizations(id);
UPDATE rpa_runs SET org_id = (SELECT id FROM organizations WHERE code='FJ') WHERE org_id IS NULL;
CREATE INDEX IF NOT EXISTS idx_rpa_runs_org ON rpa_runs(org_id);
