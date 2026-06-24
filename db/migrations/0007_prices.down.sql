DROP TABLE IF EXISTS node_spot_price_daily;

DROP TRIGGER IF EXISTS price_sgcc_set_updated_at ON price_sgcc;
DROP TABLE IF EXISTS price_sgcc;

DROP TABLE IF EXISTS day_ahead_econ_price;
DROP TABLE IF EXISTS day_ahead_spot_price;
DROP TABLE IF EXISTS real_time_spot_price;
