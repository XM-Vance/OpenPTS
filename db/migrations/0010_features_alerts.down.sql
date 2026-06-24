DROP TABLE IF EXISTS analysis_history_log;

DROP TRIGGER IF EXISTS cust_alert_set_updated_at ON customer_anomaly_alerts;
DROP TABLE IF EXISTS customer_anomaly_alerts;

DROP TRIGGER IF EXISTS cust_char_set_updated_at ON customer_characteristics;
DROP TABLE IF EXISTS customer_characteristics;
