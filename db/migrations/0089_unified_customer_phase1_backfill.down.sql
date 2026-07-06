-- 回滚 1b：删除回填进来的意向客户（customer_diagnosis 随 ON DELETE CASCADE 一并删）。
-- 仅删带 _phase1_intent 标记的回填行，不影响原 intent_customers 与正式客户。
DELETE FROM customers WHERE extra->>'_phase1_intent' = 'true' AND lifecycle_stage = 'intent';
