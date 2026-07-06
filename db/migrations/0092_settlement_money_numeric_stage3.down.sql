-- 回滚：numeric(18,4) → double precision（仅一次性库验证可逆性时使用；会丢失精度）。
ALTER TABLE green_power_trades
    ALTER COLUMN price  TYPE double precision USING price::double precision,
    ALTER COLUMN amount TYPE double precision USING amount::double precision;

ALTER TABLE mechanism_energy_plan
    ALTER COLUMN contract_price TYPE double precision USING contract_price::double precision,
    ALTER COLUMN settle_amount  TYPE double precision USING settle_amount::double precision;

ALTER TABLE vpp_dispatches
    ALTER COLUMN revenue TYPE double precision USING revenue::double precision;

ALTER TABLE day_ahead_trade_review
    ALTER COLUMN avg_declared_price TYPE double precision USING avg_declared_price::double precision,
    ALTER COLUMN avg_cleared_price  TYPE double precision USING avg_cleared_price::double precision,
    ALTER COLUMN revenue            TYPE double precision USING revenue::double precision;

ALTER TABLE storage_declaration
    ALTER COLUMN expected_revenue TYPE double precision USING expected_revenue::double precision;
