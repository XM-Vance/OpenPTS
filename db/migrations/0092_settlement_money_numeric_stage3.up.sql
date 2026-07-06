-- P4 stage-three（收尾）：其余结算货币列 double precision → numeric(18,4)。
-- 仅货币列；电量(_mw/_mwh)、曲线数组等保持 float。float → numeric 放大精度，存量无损。

ALTER TABLE green_power_trades
    ALTER COLUMN price  TYPE numeric(18,4) USING price::numeric(18,4),
    ALTER COLUMN amount TYPE numeric(18,4) USING amount::numeric(18,4);

ALTER TABLE mechanism_energy_plan
    ALTER COLUMN contract_price TYPE numeric(18,4) USING contract_price::numeric(18,4),
    ALTER COLUMN settle_amount  TYPE numeric(18,4) USING settle_amount::numeric(18,4);

ALTER TABLE vpp_dispatches
    ALTER COLUMN revenue TYPE numeric(18,4) USING revenue::numeric(18,4);

ALTER TABLE day_ahead_trade_review
    ALTER COLUMN avg_declared_price TYPE numeric(18,4) USING avg_declared_price::numeric(18,4),
    ALTER COLUMN avg_cleared_price  TYPE numeric(18,4) USING avg_cleared_price::numeric(18,4),
    ALTER COLUMN revenue            TYPE numeric(18,4) USING revenue::numeric(18,4);

ALTER TABLE storage_declaration
    ALTER COLUMN expected_revenue TYPE numeric(18,4) USING expected_revenue::numeric(18,4);
