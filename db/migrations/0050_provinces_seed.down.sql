-- 0050 回滚：把指向各省组织的 users.org_id 重置回 default，再删除各省组织。
-- （user_orgs 由 0051.down 先行删除，故此处无 FK 冲突。）
UPDATE users
   SET org_id = (SELECT id FROM organizations WHERE code = 'default')
 WHERE org_id IN (SELECT id FROM organizations WHERE code IN ('AH','JX','FJ','JS','SH','SN'));

DELETE FROM organizations WHERE code IN ('AH','JX','FJ','JS','SH','SN');
