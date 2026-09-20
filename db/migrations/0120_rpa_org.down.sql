-- 132 down: 回退 RPA org_id 隔离（仅结构，丢省归属）。
-- ⚠️ 生产/共享库不要执行（会丢 org_id 列及省归属）；仅用于一次性测试库结构回退。
-- 注：PostgreSQL 的 ADD CONSTRAINT 不支持 IF NOT EXISTS（只有 ADD COLUMN 支持），
-- 用 DO 块 + exception 捕获实现幂等。
ALTER TABLE rpa_runs DROP COLUMN IF EXISTS org_id;
DROP INDEX IF EXISTS idx_rpa_runs_org;

DROP INDEX IF EXISTS rpa_jobs_org_name_uniq;
DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'rpa_jobs_name_key') THEN
    ALTER TABLE rpa_jobs ADD CONSTRAINT rpa_jobs_name_key UNIQUE (name);
  END IF;
END $$;
DROP INDEX IF EXISTS idx_rpa_jobs_org;
ALTER TABLE rpa_jobs DROP COLUMN IF EXISTS org_id;
