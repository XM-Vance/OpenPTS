-- 碳交易行情表（碳交易业务模块）
-- 收纳中国碳市场 CEA（全国碳排放配额）、CCER（国家核证自愿减排量），
-- 以及原「市场行情」中的 EUA（欧盟碳排放配额）碳价。
--
-- 全国统一行情，与 md_* 系列一样为「共享参考数据」，不做省份(org_id)隔离：
-- 各省/总部看到的是同一套碳价，因此在「全部省」视图下也可正常读写。
CREATE TABLE IF NOT EXISTS carbon_quotes (
    id          SERIAL PRIMARY KEY,
    product     TEXT NOT NULL,                 -- CEA / CCER / EUA
    trade_date  DATE NOT NULL,
    open_price  DOUBLE PRECISION,              -- 开盘价
    high_price  DOUBLE PRECISION,              -- 最高价
    low_price   DOUBLE PRECISION,              -- 最低价
    close_price DOUBLE PRECISION,              -- 收盘价
    volume      DOUBLE PRECISION,              -- 成交量（吨 CO₂）
    turnover    DOUBLE PRECISION,              -- 成交额（元 / 欧元）
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (product, trade_date)
);

CREATE INDEX IF NOT EXISTS idx_carbon_quotes_product_date
    ON carbon_quotes (product, trade_date DESC);

-- 把原「市场行情」中既有的 EUA 碳价迁移进碳交易模块（产品标记为 EUA）。
-- md_carbon_eua 表保留（不删），但会从市场行情白名单中下线，碳价的唯一入口改为本表。
INSERT INTO carbon_quotes (product, trade_date, open_price, high_price, low_price, close_price, volume, created_at)
SELECT 'EUA', trade_date, open_price, high_price, low_price, close_price, volume, created_at
FROM md_carbon_eua
ON CONFLICT (product, trade_date) DO NOTHING;
