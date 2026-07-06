-- P4 收尾：剩余结算货币列 double precision → numeric(18,4)。
-- 仅货币列；电量(daily_energy/cumulative_energy/energy_kwh)保持 float。float → numeric 放大精度，存量无损。

ALTER TABLE bonds
    ALTER COLUMN amount TYPE numeric(18,4) USING amount::numeric(18,4);

ALTER TABLE contract_price_daily
    ALTER COLUMN unit_price        TYPE numeric(18,4) USING unit_price::numeric(18,4),
    ALTER COLUMN daily_amount      TYPE numeric(18,4) USING daily_amount::numeric(18,4),
    ALTER COLUMN cumulative_amount TYPE numeric(18,4) USING cumulative_amount::numeric(18,4);

ALTER TABLE solar_revenue_settlement
    ALTER COLUMN revenue    TYPE numeric(18,4) USING revenue::numeric(18,4),
    ALTER COLUMN avg_price  TYPE numeric(18,4) USING avg_price::numeric(18,4),
    ALTER COLUMN subsidy    TYPE numeric(18,4) USING subsidy::numeric(18,4),
    ALTER COLUMN net_income TYPE numeric(18,4) USING net_income::numeric(18,4);
