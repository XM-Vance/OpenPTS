DROP TABLE IF EXISTS actual_operation;
DROP TABLE IF EXISTS rolling_match_snapshots;
DROP TABLE IF EXISTS intent_customer_monthly_wholesale;
DROP TABLE IF EXISTS intent_customer_monthly_retail_simulation;
DROP TABLE IF EXISTS real_time_generation;

DROP TRIGGER IF EXISTS system_alerts_set_updated_at ON system_alerts;
DROP TABLE IF EXISTS system_alerts;
