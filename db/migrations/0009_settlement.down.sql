DROP TABLE IF EXISTS retail_settlement_prices;

DROP TRIGGER IF EXISTS retail_settle_daily_set_updated_at ON retail_settlement_daily;
DROP TABLE IF EXISTS retail_settlement_daily;

DROP TRIGGER IF EXISTS contracts_agg_daily_set_updated_at ON contracts_aggregated_daily;
DROP TABLE IF EXISTS contracts_aggregated_daily;

DROP TABLE IF EXISTS spot_settlement_period;

DROP TRIGGER IF EXISTS spot_settle_daily_set_updated_at ON spot_settlement_daily;
DROP TABLE IF EXISTS spot_settlement_daily;

DROP TRIGGER IF EXISTS settlement_daily_set_updated_at ON settlement_daily;
DROP TABLE IF EXISTS settlement_daily;
