-- 市场行情数据表（MySQL market_data 同步）
-- 包含：宏观经济、燃料能源、期货、利率汇率、碳价、大宗指数、风速、水文

-- ─── 宏观经济（月度/季度） ───

CREATE TABLE IF NOT EXISTS md_macro_gdp (
    id SERIAL PRIMARY KEY,
    stat_date DATE NOT NULL UNIQUE,
    gdp_yoy DOUBLE PRECISION,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS md_macro_cpi (
    id SERIAL PRIMARY KEY,
    stat_date DATE NOT NULL UNIQUE,
    cpi_yoy DOUBLE PRECISION,
    cpi_mom DOUBLE PRECISION,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS md_macro_ppi (
    id SERIAL PRIMARY KEY,
    stat_date DATE NOT NULL UNIQUE,
    ppi_yoy DOUBLE PRECISION,
    ppi_mom DOUBLE PRECISION,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS md_macro_pmi (
    id SERIAL PRIMARY KEY,
    stat_date DATE NOT NULL UNIQUE,
    pmi_value DOUBLE PRECISION,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS md_macro_electricity (
    id SERIAL PRIMARY KEY,
    stat_date DATE NOT NULL UNIQUE,
    total_usage DOUBLE PRECISION,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS md_macro_m2 (
    id SERIAL PRIMARY KEY,
    stat_date DATE NOT NULL UNIQUE,
    m2_yoy DOUBLE PRECISION,
    m2_balance DOUBLE PRECISION,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS md_macro_industrial_output (
    id SERIAL PRIMARY KEY,
    stat_date DATE NOT NULL UNIQUE,
    yoy_growth DOUBLE PRECISION,
    cum_growth DOUBLE PRECISION,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- ─── 燃料能源（日频） ───

CREATE TABLE IF NOT EXISTS md_fuel_wti (
    id SERIAL PRIMARY KEY,
    trade_date DATE NOT NULL UNIQUE,
    open_price DOUBLE PRECISION,
    high_price DOUBLE PRECISION,
    low_price DOUBLE PRECISION,
    close_price DOUBLE PRECISION,
    volume DOUBLE PRECISION,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS md_fuel_natgas_hh (
    id SERIAL PRIMARY KEY,
    trade_date DATE NOT NULL UNIQUE,
    open_price DOUBLE PRECISION,
    high_price DOUBLE PRECISION,
    low_price DOUBLE PRECISION,
    close_price DOUBLE PRECISION,
    volume DOUBLE PRECISION,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS md_fuel_ine_crude (
    id SERIAL PRIMARY KEY,
    trade_date DATE NOT NULL UNIQUE,
    open_price DOUBLE PRECISION,
    high_price DOUBLE PRECISION,
    low_price DOUBLE PRECISION,
    close_price DOUBLE PRECISION,
    volume DOUBLE PRECISION,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS md_fuel_cn_oil_price (
    id SERIAL PRIMARY KEY,
    adjust_date DATE NOT NULL UNIQUE,
    gasoline_price DOUBLE PRECISION,
    diesel_price DOUBLE PRECISION,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- ─── 期货（日频） ───

CREATE TABLE IF NOT EXISTS md_futures_rb (
    id SERIAL PRIMARY KEY,
    trade_date DATE NOT NULL UNIQUE,
    open_price DOUBLE PRECISION,
    high_price DOUBLE PRECISION,
    low_price DOUBLE PRECISION,
    close_price DOUBLE PRECISION,
    volume DOUBLE PRECISION,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS md_futures_i (
    id SERIAL PRIMARY KEY,
    trade_date DATE NOT NULL UNIQUE,
    open_price DOUBLE PRECISION,
    high_price DOUBLE PRECISION,
    low_price DOUBLE PRECISION,
    close_price DOUBLE PRECISION,
    volume DOUBLE PRECISION,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS md_futures_al (
    id SERIAL PRIMARY KEY,
    trade_date DATE NOT NULL UNIQUE,
    open_price DOUBLE PRECISION,
    high_price DOUBLE PRECISION,
    low_price DOUBLE PRECISION,
    close_price DOUBLE PRECISION,
    volume DOUBLE PRECISION,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS md_futures_au (
    id SERIAL PRIMARY KEY,
    trade_date DATE NOT NULL UNIQUE,
    open_price DOUBLE PRECISION,
    high_price DOUBLE PRECISION,
    low_price DOUBLE PRECISION,
    close_price DOUBLE PRECISION,
    volume DOUBLE PRECISION,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS md_futures_cu (
    id SERIAL PRIMARY KEY,
    trade_date DATE NOT NULL UNIQUE,
    open_price DOUBLE PRECISION,
    high_price DOUBLE PRECISION,
    low_price DOUBLE PRECISION,
    close_price DOUBLE PRECISION,
    volume DOUBLE PRECISION,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS md_futures_zn (
    id SERIAL PRIMARY KEY,
    trade_date DATE NOT NULL UNIQUE,
    open_price DOUBLE PRECISION,
    high_price DOUBLE PRECISION,
    low_price DOUBLE PRECISION,
    close_price DOUBLE PRECISION,
    volume DOUBLE PRECISION,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS md_futures_hc (
    id SERIAL PRIMARY KEY,
    trade_date DATE NOT NULL UNIQUE,
    open_price DOUBLE PRECISION,
    high_price DOUBLE PRECISION,
    low_price DOUBLE PRECISION,
    close_price DOUBLE PRECISION,
    volume DOUBLE PRECISION,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS md_futures_fg (
    id SERIAL PRIMARY KEY,
    trade_date DATE NOT NULL UNIQUE,
    open_price DOUBLE PRECISION,
    high_price DOUBLE PRECISION,
    low_price DOUBLE PRECISION,
    close_price DOUBLE PRECISION,
    volume DOUBLE PRECISION,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS md_futures_sa (
    id SERIAL PRIMARY KEY,
    trade_date DATE NOT NULL UNIQUE,
    open_price DOUBLE PRECISION,
    high_price DOUBLE PRECISION,
    low_price DOUBLE PRECISION,
    close_price DOUBLE PRECISION,
    volume DOUBLE PRECISION,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS md_futures_zc (
    id SERIAL PRIMARY KEY,
    trade_date DATE NOT NULL UNIQUE,
    open_price DOUBLE PRECISION,
    high_price DOUBLE PRECISION,
    low_price DOUBLE PRECISION,
    close_price DOUBLE PRECISION,
    volume DOUBLE PRECISION,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- ─── 利率（日频） ───

CREATE TABLE IF NOT EXISTS md_rate_shibor (
    id SERIAL PRIMARY KEY,
    trade_date DATE NOT NULL UNIQUE,
    overnight DOUBLE PRECISION,
    week_1 DOUBLE PRECISION,
    month_1 DOUBLE PRECISION,
    month_3 DOUBLE PRECISION,
    month_6 DOUBLE PRECISION,
    year_1 DOUBLE PRECISION,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS md_rate_lpr (
    id SERIAL PRIMARY KEY,
    stat_date DATE NOT NULL UNIQUE,
    lpr_1y DOUBLE PRECISION,
    lpr_5y DOUBLE PRECISION,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- ─── 汇率 + 国债（日频） ───

CREATE TABLE IF NOT EXISTS md_fx_usdcny (
    id SERIAL PRIMARY KEY,
    trade_date DATE NOT NULL UNIQUE,
    open_price DOUBLE PRECISION,
    high_price DOUBLE PRECISION,
    low_price DOUBLE PRECISION,
    close_price DOUBLE PRECISION,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS md_bond_zh_us_yield (
    id SERIAL PRIMARY KEY,
    trade_date DATE NOT NULL UNIQUE,
    cn_2y DOUBLE PRECISION,
    cn_5y DOUBLE PRECISION,
    cn_10y DOUBLE PRECISION,
    cn_30y DOUBLE PRECISION,
    us_2y DOUBLE PRECISION,
    us_5y DOUBLE PRECISION,
    us_10y DOUBLE PRECISION,
    us_30y DOUBLE PRECISION,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- ─── 指数（日频） ───

CREATE TABLE IF NOT EXISTS md_index_dxy (
    id SERIAL PRIMARY KEY,
    trade_date DATE NOT NULL UNIQUE,
    open_price DOUBLE PRECISION,
    high_price DOUBLE PRECISION,
    low_price DOUBLE PRECISION,
    close_price DOUBLE PRECISION,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS md_index_bdi (
    id SERIAL PRIMARY KEY,
    trade_date DATE NOT NULL UNIQUE,
    bdi_value DOUBLE PRECISION,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- ─── 碳价（日频） ───

CREATE TABLE IF NOT EXISTS md_carbon_eua (
    id SERIAL PRIMARY KEY,
    trade_date DATE NOT NULL UNIQUE,
    open_price DOUBLE PRECISION,
    high_price DOUBLE PRECISION,
    low_price DOUBLE PRECISION,
    close_price DOUBLE PRECISION,
    volume DOUBLE PRECISION,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- ─── 风速（小时频） ───

CREATE TABLE IF NOT EXISTS md_weather_wind_hourly (
    id SERIAL PRIMARY KEY,
    location_code VARCHAR(20) NOT NULL,
    location_name VARCHAR(50) NOT NULL,
    lat NUMERIC(6,2) NOT NULL,
    lon NUMERIC(6,2) NOT NULL,
    obs_time TIMESTAMPTZ NOT NULL,
    wind_speed_100m DOUBLE PRECISION,
    wind_dir_100m SMALLINT,
    temperature_2m DOUBLE PRECISION,
    humidity_2m DOUBLE PRECISION,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (location_code, obs_time)
);

CREATE INDEX idx_md_wind_time ON md_weather_wind_hourly(obs_time);
CREATE INDEX idx_md_wind_loc ON md_weather_wind_hourly(location_code);

-- ─── 水文气象（日频） ───

CREATE TABLE IF NOT EXISTS md_weather_hydrology_daily (
    id SERIAL PRIMARY KEY,
    location_code VARCHAR(20) NOT NULL,
    location_name VARCHAR(50) NOT NULL,
    lat NUMERIC(6,2) NOT NULL,
    lon NUMERIC(6,2) NOT NULL,
    obs_date DATE NOT NULL,
    temp_mean DOUBLE PRECISION,
    humidity_mean DOUBLE PRECISION,
    precipitation_sum DOUBLE PRECISION,
    rain_sum DOUBLE PRECISION,
    et0_evapotranspiration DOUBLE PRECISION,
    wind_speed_10m_mean DOUBLE PRECISION,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (location_code, obs_date)
);

CREATE INDEX idx_md_hydro_date ON md_weather_hydrology_daily(obs_date);
CREATE INDEX idx_md_hydro_loc ON md_weather_hydrology_daily(location_code);
