-- 回滚：numeric(18,4) → double precision（仅一次性库验证可逆性时使用；会丢失精度）。
ALTER TABLE bonds
    ALTER COLUMN amount TYPE double precision USING amount::double precision;

ALTER TABLE contract_price_daily
    ALTER COLUMN unit_price        TYPE double precision USING unit_price::double precision,
    ALTER COLUMN daily_amount      TYPE double precision USING daily_amount::double precision,
    ALTER COLUMN cumulative_amount TYPE double precision USING cumulative_amount::double precision;

ALTER TABLE solar_revenue_settlement
    ALTER COLUMN revenue    TYPE double precision USING revenue::double precision,
    ALTER COLUMN avg_price  TYPE double precision USING avg_price::double precision,
    ALTER COLUMN subsidy    TYPE double precision USING subsidy::double precision,
    ALTER COLUMN net_income TYPE double precision USING net_income::double precision;
