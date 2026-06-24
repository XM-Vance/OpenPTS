-- v1 复刻第三批：F1 利润 / F2 月度复盘 / F3 撮合报价 / F4 手工数据。

-- F1 客户利润分析
CREATE TABLE IF NOT EXISTS customer_profit (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id     uuid NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    operating_month text NOT NULL,
    revenue         double precision NOT NULL DEFAULT 0,
    cost            double precision NOT NULL DEFAULT 0,
    gross_profit    double precision NOT NULL DEFAULT 0,
    gross_margin    double precision NOT NULL DEFAULT 0,    -- %
    energy_mwh      double precision NOT NULL DEFAULT 0,
    created_at      timestamp NOT NULL DEFAULT now(),
    UNIQUE (customer_id, operating_month)
);

CREATE INDEX IF NOT EXISTS idx_customer_profit_month
    ON customer_profit(operating_month DESC);

-- F2 月度交易复盘
CREATE TABLE IF NOT EXISTS monthly_trade_review (
    id                uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    operating_month   text NOT NULL UNIQUE,
    retail_revenue    double precision NOT NULL DEFAULT 0,
    wholesale_cost    double precision NOT NULL DEFAULT 0,
    gross_profit      double precision NOT NULL DEFAULT 0,
    gross_margin      double precision NOT NULL DEFAULT 0,
    active_customers  integer NOT NULL DEFAULT 0,
    total_energy_mwh  double precision NOT NULL DEFAULT 0,
    note              text,
    created_at        timestamp NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_monthly_trade_review_month
    ON monthly_trade_review(operating_month DESC);

-- F3 滚动撮合报价
CREATE TABLE IF NOT EXISTS rolling_match_quotes (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    match_date      date NOT NULL,
    match_session   integer NOT NULL,            -- 撮合时段 1-N
    side            text NOT NULL,               -- buy / sell
    declared_mw     double precision NOT NULL,
    cleared_mw      double precision NOT NULL DEFAULT 0,
    declared_price  double precision NOT NULL,   -- 元/MWh
    cleared_price   double precision NOT NULL DEFAULT 0,
    status          text NOT NULL DEFAULT 'cleared', -- cleared / partial / failed
    created_at      timestamp NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_match_quotes_date_session
    ON rolling_match_quotes(match_date DESC, match_session);

-- F4 月度手工数据
CREATE TABLE IF NOT EXISTS monthly_manual_data (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    operating_month text NOT NULL,
    category        text NOT NULL,                -- 收入 / 成本 / 偏差 / 其他
    item_name       text NOT NULL,
    value           double precision NOT NULL,
    unit            text NOT NULL DEFAULT '元',
    source          text,                         -- 数据来源（人工 / 系统 / 外部）
    note            text,
    created_by      text,
    created_at      timestamp NOT NULL DEFAULT now(),
    updated_at      timestamp NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_monthly_manual_data_month
    ON monthly_manual_data(operating_month DESC);
CREATE INDEX IF NOT EXISTS idx_monthly_manual_data_category
    ON monthly_manual_data(category);
