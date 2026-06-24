-- 日前交易复盘：每个交易日一条聚合（每日 96 时段聚合后的总览）。

CREATE TABLE IF NOT EXISTS day_ahead_trade_review (
    id                 uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    trading_date       date NOT NULL UNIQUE,
    declared_energy_mwh double precision NOT NULL DEFAULT 0,
    cleared_energy_mwh  double precision NOT NULL DEFAULT 0,
    avg_declared_price  double precision NOT NULL DEFAULT 0,
    avg_cleared_price   double precision NOT NULL DEFAULT 0,
    revenue             double precision NOT NULL DEFAULT 0,
    note                text,
    created_at          timestamp NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_da_trade_review_date
    ON day_ahead_trade_review(trading_date DESC);
