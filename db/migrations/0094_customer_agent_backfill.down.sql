-- 回滚：清空本次按 source=agent_name 回填写入的 agent_id。
-- 0088 起 agent_id 默认 NULL、本迁移是首个写入者，故 up→down→up 可逆。
-- （Phase 2 之后若客户表单已开始手工维护 agent_id，则只回退仍与 source 同名的行，
--   手工绑定的不受影响——这是保守且安全的下界。）
UPDATE customers c
SET agent_id = NULL
FROM agents a
WHERE c.agent_id = a.id
  AND c.source IS NOT NULL
  AND c.source = a.agent_name
  AND c.org_id = a.org_id;
