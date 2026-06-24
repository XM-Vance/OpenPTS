-- 任务 1: 客户档案域 org_id
-- customers 已有 org_id（0033），把为空的回填到默认组织。
UPDATE customers SET org_id = (SELECT id FROM organizations WHERE code='default')
  WHERE org_id = (SELECT id FROM organizations WHERE code='default')
     OR (org_id IS NULL);

-- 唯一约束：UNIQUE(user_name, source) → UNIQUE(org_id, user_name, source)
ALTER TABLE customers DROP CONSTRAINT IF EXISTS customers_user_name_source_unique;
CREATE UNIQUE INDEX IF NOT EXISTS customers_org_user_source_uniq ON customers(org_id, user_name, source);

-- intent_customers
ALTER TABLE intent_customers ADD COLUMN IF NOT EXISTS org_id uuid REFERENCES organizations(id);
UPDATE intent_customers SET org_id = (SELECT id FROM organizations WHERE code='default') WHERE org_id IS NULL;
CREATE INDEX IF NOT EXISTS idx_intent_customers_org ON intent_customers(org_id);

-- intent_customer_load_curve_daily
ALTER TABLE intent_customer_load_curve_daily ADD COLUMN IF NOT EXISTS org_id uuid REFERENCES organizations(id);
UPDATE intent_customer_load_curve_daily SET org_id = (SELECT id FROM organizations WHERE code='default') WHERE org_id IS NULL;
CREATE INDEX IF NOT EXISTS idx_intent_load_curve_org ON intent_customer_load_curve_daily(org_id);

-- intent_customer_meter_reads_daily
ALTER TABLE intent_customer_meter_reads_daily ADD COLUMN IF NOT EXISTS org_id uuid REFERENCES organizations(id);
UPDATE intent_customer_meter_reads_daily SET org_id = (SELECT id FROM organizations WHERE code='default') WHERE org_id IS NULL;
CREATE INDEX IF NOT EXISTS idx_intent_meter_reads_org ON intent_customer_meter_reads_daily(org_id);

-- intent_customer_monthly_retail_simulation
ALTER TABLE intent_customer_monthly_retail_simulation ADD COLUMN IF NOT EXISTS org_id uuid REFERENCES organizations(id);
UPDATE intent_customer_monthly_retail_simulation SET org_id = (SELECT id FROM organizations WHERE code='default') WHERE org_id IS NULL;
CREATE INDEX IF NOT EXISTS idx_intent_retail_sim_org ON intent_customer_monthly_retail_simulation(org_id);
ALTER TABLE intent_customer_monthly_retail_simulation DROP CONSTRAINT IF EXISTS intent_customer_monthly_retail_s_intent_id_month_package_id_key;
CREATE UNIQUE INDEX IF NOT EXISTS intent_retail_sim_org_uniq ON intent_customer_monthly_retail_simulation(org_id, intent_id, month, package_id);

-- intent_customer_monthly_wholesale
ALTER TABLE intent_customer_monthly_wholesale ADD COLUMN IF NOT EXISTS org_id uuid REFERENCES organizations(id);
UPDATE intent_customer_monthly_wholesale SET org_id = (SELECT id FROM organizations WHERE code='default') WHERE org_id IS NULL;
CREATE INDEX IF NOT EXISTS idx_intent_wholesale_org ON intent_customer_monthly_wholesale(org_id);
ALTER TABLE intent_customer_monthly_wholesale DROP CONSTRAINT IF EXISTS intent_customer_monthly_wholesale_intent_id_month_key;
CREATE UNIQUE INDEX IF NOT EXISTS intent_wholesale_org_uniq ON intent_customer_monthly_wholesale(org_id, intent_id, month);

-- customer_analysis
ALTER TABLE customer_analysis ADD COLUMN IF NOT EXISTS org_id uuid REFERENCES organizations(id);
UPDATE customer_analysis SET org_id = (SELECT id FROM organizations WHERE code='default') WHERE org_id IS NULL;
CREATE INDEX IF NOT EXISTS idx_customer_analysis_org ON customer_analysis(org_id);
ALTER TABLE customer_analysis DROP CONSTRAINT IF EXISTS customer_analysis_customer_id_analysis_month_key;
CREATE UNIQUE INDEX IF NOT EXISTS customer_analysis_org_uniq ON customer_analysis(org_id, customer_id, analysis_month);

-- customer_anomaly_alerts
ALTER TABLE customer_anomaly_alerts ADD COLUMN IF NOT EXISTS org_id uuid REFERENCES organizations(id);
UPDATE customer_anomaly_alerts SET org_id = (SELECT id FROM organizations WHERE code='default') WHERE org_id IS NULL;
CREATE INDEX IF NOT EXISTS idx_customer_alerts_org ON customer_anomaly_alerts(org_id);

-- customer_characteristics
ALTER TABLE customer_characteristics ADD COLUMN IF NOT EXISTS org_id uuid REFERENCES organizations(id);
UPDATE customer_characteristics SET org_id = (SELECT id FROM organizations WHERE code='default') WHERE org_id IS NULL;
CREATE INDEX IF NOT EXISTS idx_customer_chars_org ON customer_characteristics(org_id);
ALTER TABLE customer_characteristics DROP CONSTRAINT IF EXISTS customer_characteristics_customer_id_data_date_key;
CREATE UNIQUE INDEX IF NOT EXISTS customer_chars_org_uniq ON customer_characteristics(org_id, customer_id, data_date);

-- customer_demo_aliases
ALTER TABLE customer_demo_aliases ADD COLUMN IF NOT EXISTS org_id uuid REFERENCES organizations(id);
UPDATE customer_demo_aliases SET org_id = (SELECT id FROM organizations WHERE code='default') WHERE org_id IS NULL;
CREATE INDEX IF NOT EXISTS idx_customer_demo_aliases_org ON customer_demo_aliases(org_id);

-- customer_monthly_energy
ALTER TABLE customer_monthly_energy ADD COLUMN IF NOT EXISTS org_id uuid REFERENCES organizations(id);
UPDATE customer_monthly_energy SET org_id = (SELECT id FROM organizations WHERE code='default') WHERE org_id IS NULL;
CREATE INDEX IF NOT EXISTS idx_customer_energy_org ON customer_monthly_energy(org_id);
ALTER TABLE customer_monthly_energy DROP CONSTRAINT IF EXISTS customer_monthly_energy_customer_id_month_key;
CREATE UNIQUE INDEX IF NOT EXISTS customer_energy_org_uniq ON customer_monthly_energy(org_id, customer_id, month);

-- customer_profit
ALTER TABLE customer_profit ADD COLUMN IF NOT EXISTS org_id uuid REFERENCES organizations(id);
UPDATE customer_profit SET org_id = (SELECT id FROM organizations WHERE code='default') WHERE org_id IS NULL;
CREATE INDEX IF NOT EXISTS idx_customer_profit_org ON customer_profit(org_id);
ALTER TABLE customer_profit DROP CONSTRAINT IF EXISTS customer_profit_customer_id_operating_month_key;
CREATE UNIQUE INDEX IF NOT EXISTS customer_profit_org_uniq ON customer_profit(org_id, customer_id, operating_month);
