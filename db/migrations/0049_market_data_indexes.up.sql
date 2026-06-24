-- 0049: market-data 各表日期列索引补齐
-- 0048 仅给 2 张天气表建了索引；其余表查询为 WHERE 日期列 >= $1 ORDER BY 日期列 DESC，
-- 缺索引会全表扫+排序。这里给各表日期列补 DESC 索引（覆盖过滤+排序+MIN/MAX）。

CREATE INDEX IF NOT EXISTS idx_md_macro_gdp_date ON md_macro_gdp(stat_date DESC);
CREATE INDEX IF NOT EXISTS idx_md_macro_cpi_date ON md_macro_cpi(stat_date DESC);
CREATE INDEX IF NOT EXISTS idx_md_macro_ppi_date ON md_macro_ppi(stat_date DESC);
CREATE INDEX IF NOT EXISTS idx_md_macro_pmi_date ON md_macro_pmi(stat_date DESC);
CREATE INDEX IF NOT EXISTS idx_md_macro_electricity_date ON md_macro_electricity(stat_date DESC);
CREATE INDEX IF NOT EXISTS idx_md_macro_m2_date ON md_macro_m2(stat_date DESC);
CREATE INDEX IF NOT EXISTS idx_md_macro_industrial_output_date ON md_macro_industrial_output(stat_date DESC);
CREATE INDEX IF NOT EXISTS idx_md_fuel_wti_date ON md_fuel_wti(trade_date DESC);
CREATE INDEX IF NOT EXISTS idx_md_fuel_natgas_hh_date ON md_fuel_natgas_hh(trade_date DESC);
CREATE INDEX IF NOT EXISTS idx_md_fuel_ine_crude_date ON md_fuel_ine_crude(trade_date DESC);
CREATE INDEX IF NOT EXISTS idx_md_fuel_cn_oil_price_date ON md_fuel_cn_oil_price(adjust_date DESC);
CREATE INDEX IF NOT EXISTS idx_md_futures_rb_date ON md_futures_rb(trade_date DESC);
CREATE INDEX IF NOT EXISTS idx_md_futures_i_date ON md_futures_i(trade_date DESC);
CREATE INDEX IF NOT EXISTS idx_md_futures_al_date ON md_futures_al(trade_date DESC);
CREATE INDEX IF NOT EXISTS idx_md_futures_au_date ON md_futures_au(trade_date DESC);
CREATE INDEX IF NOT EXISTS idx_md_futures_cu_date ON md_futures_cu(trade_date DESC);
CREATE INDEX IF NOT EXISTS idx_md_futures_zn_date ON md_futures_zn(trade_date DESC);
CREATE INDEX IF NOT EXISTS idx_md_futures_hc_date ON md_futures_hc(trade_date DESC);
CREATE INDEX IF NOT EXISTS idx_md_futures_fg_date ON md_futures_fg(trade_date DESC);
CREATE INDEX IF NOT EXISTS idx_md_futures_sa_date ON md_futures_sa(trade_date DESC);
CREATE INDEX IF NOT EXISTS idx_md_futures_zc_date ON md_futures_zc(trade_date DESC);
CREATE INDEX IF NOT EXISTS idx_md_rate_shibor_date ON md_rate_shibor(trade_date DESC);
CREATE INDEX IF NOT EXISTS idx_md_rate_lpr_date ON md_rate_lpr(stat_date DESC);
CREATE INDEX IF NOT EXISTS idx_md_fx_usdcny_date ON md_fx_usdcny(trade_date DESC);
CREATE INDEX IF NOT EXISTS idx_md_bond_zh_us_yield_date ON md_bond_zh_us_yield(trade_date DESC);
CREATE INDEX IF NOT EXISTS idx_md_index_dxy_date ON md_index_dxy(trade_date DESC);
CREATE INDEX IF NOT EXISTS idx_md_index_bdi_date ON md_index_bdi(trade_date DESC);
CREATE INDEX IF NOT EXISTS idx_md_carbon_eua_date ON md_carbon_eua(trade_date DESC);
