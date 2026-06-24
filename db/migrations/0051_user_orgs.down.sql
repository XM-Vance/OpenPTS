-- 0051 回滚：删多对多表与总部标记。
-- （users.org_id 的省份回填由 0050.down 统一重置回 default。）
DROP INDEX IF EXISTS idx_user_orgs_org;
DROP TABLE IF EXISTS user_orgs;
ALTER TABLE users DROP COLUMN IF EXISTS is_hq;
