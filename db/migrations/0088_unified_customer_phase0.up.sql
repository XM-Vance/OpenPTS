-- 统一客户+生命周期 Phase 0：加法式 schema（零行为变更，列暂不被代码读写）。
-- 目标：为"意向→服务"统一主体打地基——customers 加生命周期阶段 + 代理商外键；
-- 建客户诊断卫星表（承接意向期 meters/coverage/评分，转正后不丢）。
-- 注：intent_customers.org_id 已由 0052 添加并回填 FJ，本条不重复处理。
-- 序号说明：main 现到 0083；0084-0087 属 P4 金额精度分支（未合）。本条取 0088 避免合并冲突。

-- 1) customers：生命周期阶段 + 代理商外键
ALTER TABLE customers ADD COLUMN IF NOT EXISTS lifecycle_stage VARCHAR(16) NOT NULL DEFAULT 'service';
ALTER TABLE customers ADD COLUMN IF NOT EXISTS agent_id uuid REFERENCES agents(id);
-- 存量客户均为已签约 → 默认 'service'。阶段取值约束：线索/意向/服务/流失。
ALTER TABLE customers ADD CONSTRAINT customers_lifecycle_stage_chk
    CHECK (lifecycle_stage IN ('lead','intent','service','churned'));
CREATE INDEX IF NOT EXISTS idx_customers_lifecycle_stage ON customers(lifecycle_stage);
CREATE INDEX IF NOT EXISTS idx_customers_agent_id ON customers(agent_id);

-- 2) 客户诊断卫星表（按客户，多次诊断快照；转正后随客户保留）
CREATE TABLE IF NOT EXISTS customer_diagnosis (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id     uuid NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    org_id          uuid REFERENCES organizations(id),
    meters          jsonb NOT NULL DEFAULT '[]'::jsonb,    -- 电表 [{meter_id, name, multiplier}]
    coverage_start  date,
    coverage_end    date,
    coverage_days   int,
    completeness    numeric(5,2),                          -- 数据完整度 0-100
    avg_daily_load  numeric(14,4),
    data_score      numeric(5,2),                          -- 诊断评分 0-100
    load_score      numeric(5,2),
    coverage_score  numeric(5,2),
    overall_score   numeric(5,2),
    matched_package text,                                  -- 匹配到的零售套餐
    recommendation  text,
    diagnosed_at    timestamptz NOT NULL DEFAULT now(),
    created_at      timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_customer_diagnosis_customer ON customer_diagnosis(customer_id);
CREATE INDEX IF NOT EXISTS idx_customer_diagnosis_org ON customer_diagnosis(org_id);
