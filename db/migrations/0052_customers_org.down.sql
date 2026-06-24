-- 回滚 0052: 客户档案域 org_id

-- 恢复 customers 原始唯一约束
DROP INDEX IF EXISTS customers_org_user_source_uniq;
ALTER TABLE customers ADD CONSTRAINT customers_user_name_source_unique UNIQUE (user_name, source);

-- intent_customers
DROP INDEX IF EXISTS idx_intent_customers_org;
ALTER TABLE intent_customers DROP COLUMN IF EXISTS org_id;

-- intent_customer_load_curve_daily
DROP INDEX IF EXISTS idx_intent_load_curve_org;
ALTER TABLE intent_customer_load_curve_daily DROP COLUMN IF EXISTS org_id;

-- intent_customer_meter_reads_daily
DROP INDEX IF EXISTS idx_intent_meter_reads_org;
ALTER TABLE intent_customer_meter_reads_daily DROP COLUMN IF EXISTS org_id;

-- intent_customer_monthly_retail_simulation
DROP INDEX IF EXISTS intent_retail_sim_org_uniq;
DROP INDEX IF EXISTS idx_intent_retail_sim_org;
ALTER TABLE intent_customer_monthly_retail_simulation DROP COLUMN IF EXISTS org_id;
ALTER TABLE intent_customer_monthly_retail_simulation ADD CONSTRAINT intent_customer_monthly_retail_s_intent_id_month_package_id_key UNIQUE (intent_id, month, package_id);

-- intent_customer_monthly_wholesale
DROP INDEX IF EXISTS intent_wholesale_org_uniq;
DROP INDEX IF EXISTS idx_intent_wholesale_org;
ALTER TABLE intent_customer_monthly_wholesale DROP COLUMN IF EXISTS org_id;
ALTER TABLE intent_customer_monthly_wholesale ADD CONSTRAINT intent_customer_monthly_wholesale_intent_id_month_key UNIQUE (intent_id, month);

-- customer_analysis
DROP INDEX IF EXISTS customer_analysis_org_uniq;
DROP INDEX IF EXISTS idx_customer_analysis_org;
ALTER TABLE customer_analysis DROP COLUMN IF EXISTS org_id;
ALTER TABLE customer_analysis ADD CONSTRAINT customer_analysis_customer_id_analysis_month_key UNIQUE (customer_id, analysis_month);

-- customer_anomaly_alerts
DROP INDEX IF EXISTS idx_customer_alerts_org;
ALTER TABLE customer_anomaly_alerts DROP COLUMN IF EXISTS org_id;

-- customer_characteristics
DROP INDEX IF EXISTS customer_chars_org_uniq;
DROP INDEX IF EXISTS idx_customer_chars_org;
ALTER TABLE customer_characteristics DROP COLUMN IF EXISTS org_id;
ALTER TABLE customer_characteristics ADD CONSTRAINT customer_characteristics_customer_id_data_date_key UNIQUE (customer_id, data_date);

-- customer_demo_aliases
DROP INDEX IF EXISTS idx_customer_demo_aliases_org;
ALTER TABLE customer_demo_aliases DROP COLUMN IF EXISTS org_id;

-- customer_monthly_energy
DROP INDEX IF EXISTS customer_energy_org_uniq;
DROP INDEX IF EXISTS idx_customer_energy_org;
ALTER TABLE customer_monthly_energy DROP COLUMN IF EXISTS org_id;
ALTER TABLE customer_monthly_energy ADD CONSTRAINT customer_monthly_energy_customer_id_month_key UNIQUE (customer_id, month);

-- customer_profit
DROP INDEX IF EXISTS customer_profit_org_uniq;
DROP INDEX IF EXISTS idx_customer_profit_org;
ALTER TABLE customer_profit DROP COLUMN IF EXISTS org_id;
ALTER TABLE customer_profit ADD CONSTRAINT customer_profit_customer_id_operating_month_key UNIQUE (customer_id, operating_month);
