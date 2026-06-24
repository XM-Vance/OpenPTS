-- 0037: 回滚新增模块表

DROP TABLE IF EXISTS trade_strategies CASCADE;
DROP TABLE IF EXISTS customer_analysis CASCADE;
DROP TABLE IF EXISTS load_characteristics CASCADE;
DROP TABLE IF EXISTS bidding_records CASCADE;
DROP TABLE IF EXISTS vpp_dispatches CASCADE;
DROP TABLE IF EXISTS vpp_resources CASCADE;
DROP TABLE IF EXISTS spot_market_daily CASCADE;
DROP TABLE IF EXISTS rolling_trades CASCADE;
DROP TABLE IF EXISTS green_power_trades CASCADE;
DROP TABLE IF EXISTS deviation_settlement CASCADE;
DROP TABLE IF EXISTS contract_progress CASCADE;
