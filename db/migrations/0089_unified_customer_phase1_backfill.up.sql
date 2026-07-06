-- 统一客户 Phase 1b：把意向客户并入统一 customers(stage='intent') + customer_diagnosis。
-- 前置：0088（lifecycle_stage/agent_id/customer_diagnosis）+ 1a（customers 读取点已加 stage 过滤）。
--
-- 规则：
-- - 未转正意向(converted_to IS NULL 且 status<>'converted') → 新增 customers 行，
--   id 沿用 intent.id（过渡期 convert/前端引用一致），标记 extra._phase1_intent 便于回滚。
-- - 诊断数据(meters/coverage/completeness/avg_daily_load) → customer_diagnosis 快照。
-- - 已转正意向不处理（其 converted_to 已指向正式 service 客户）。

INSERT INTO customers (id, user_name, source, lifecycle_stage, org_id, extra, is_demo, created_at, updated_at)
SELECT ic.id, ic.customer_name, '意向客户', 'intent', ic.org_id,
       jsonb_build_object('_phase1_intent', true), false, ic.created_at, now()
FROM intent_customers ic
WHERE ic.converted_to IS NULL AND ic.status <> 'converted'
ON CONFLICT (id) DO NOTHING;

INSERT INTO customer_diagnosis (customer_id, org_id, meters, coverage_start, coverage_end,
       coverage_days, completeness, avg_daily_load, diagnosed_at, created_at)
SELECT ic.id, ic.org_id, COALESCE(ic.meters, '[]'::jsonb), ic.coverage_start, ic.coverage_end,
       ic.coverage_days, ic.completeness, ic.avg_daily_load, ic.created_at, now()
FROM intent_customers ic
WHERE ic.converted_to IS NULL AND ic.status <> 'converted'
  AND EXISTS (SELECT 1 FROM customers c WHERE c.id = ic.id AND c.extra->>'_phase1_intent' = 'true');
