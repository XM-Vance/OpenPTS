-- Phase 2：回填 customers.agent_id（历史靠 source 文本弱关联代理商名，改为 FK）。
-- 仅回填 agent_id 为空、且 source 恰好等于同省某代理商名的行；
-- 同省代理商名唯一性由业务保证，若偶有重名取任一不影响后续以 FK 为准的维护。
UPDATE customers c
SET agent_id = a.id
FROM agents a
WHERE c.agent_id IS NULL
  AND c.source IS NOT NULL
  AND c.source = a.agent_name
  AND c.org_id = a.org_id;
