-- Phase 4：旧意向子系统下线（改名归档，不删数据，最安全可逆）。
-- 自 Phase 1b 起意向客户已统一在 customers(stage='intent')，Phase 3b 又用 customer_diagnosis
-- 提供其负荷/分析；下列表均已无 live 代码引用，仅改名归档保留。
-- 表内 FK（4 张卫星表 intent_id→intent_customers、documents.intent_customer_id）按 OID 引用，
-- 随表改名自动跟随，无需改动。
ALTER TABLE IF EXISTS intent_customers RENAME TO intent_customers_deprecated;
ALTER TABLE IF EXISTS intent_customer_load_curve_daily RENAME TO intent_customer_load_curve_daily_deprecated;
ALTER TABLE IF EXISTS intent_customer_meter_reads_daily RENAME TO intent_customer_meter_reads_daily_deprecated;
ALTER TABLE IF EXISTS intent_customer_monthly_retail_simulation RENAME TO intent_customer_monthly_retail_simulation_deprecated;
ALTER TABLE IF EXISTS intent_customer_monthly_wholesale RENAME TO intent_customer_monthly_wholesale_deprecated;
