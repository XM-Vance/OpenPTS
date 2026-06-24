DROP TABLE IF EXISTS intent_customer_meter_reads_daily;
DROP TABLE IF EXISTS intent_customer_load_curve_daily;

DROP TRIGGER IF EXISTS customer_monthly_energy_set_updated_at ON customer_monthly_energy;
DROP TABLE IF EXISTS customer_monthly_energy;

DROP TRIGGER IF EXISTS user_load_set_updated_at ON user_load_data;
DROP TABLE IF EXISTS user_load_data;

DROP TRIGGER IF EXISTS unified_load_set_updated_at ON unified_load_curve;
DROP TABLE IF EXISTS unified_load_curve;
