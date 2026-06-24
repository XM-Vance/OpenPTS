-- 合同电价日维度：每个合同每天一条单价记录。

CREATE TABLE IF NOT EXISTS contract_price_daily (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    contract_id     uuid NOT NULL REFERENCES retail_contracts(id) ON DELETE CASCADE,
    price_date      date NOT NULL,
    unit_price      double precision NOT NULL,           -- 元/MWh
    daily_energy    double precision NOT NULL DEFAULT 0, -- 当日执行电量 MWh
    daily_amount    double precision NOT NULL DEFAULT 0, -- 当日金额 = unit_price * daily_energy
    cumulative_energy double precision NOT NULL DEFAULT 0,
    cumulative_amount double precision NOT NULL DEFAULT 0,
    created_at      timestamp NOT NULL DEFAULT now(),
    UNIQUE (contract_id, price_date)
);

CREATE INDEX IF NOT EXISTS idx_contract_price_daily_date
    ON contract_price_daily(contract_id, price_date DESC);
