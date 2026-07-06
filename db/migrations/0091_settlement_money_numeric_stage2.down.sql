-- 回滚：numeric(18,4) → double precision（仅一次性库验证可逆性时使用；会丢失精度）。
ALTER TABLE customer_profit
    ALTER COLUMN revenue      TYPE double precision USING revenue::double precision,
    ALTER COLUMN cost         TYPE double precision USING cost::double precision,
    ALTER COLUMN gross_profit TYPE double precision USING gross_profit::double precision;

ALTER TABLE monthly_trade_review
    ALTER COLUMN retail_revenue TYPE double precision USING retail_revenue::double precision,
    ALTER COLUMN wholesale_cost TYPE double precision USING wholesale_cost::double precision,
    ALTER COLUMN gross_profit   TYPE double precision USING gross_profit::double precision;

ALTER TABLE retail_monthly_settlement
    ALTER COLUMN weighted_avg_price TYPE double precision USING weighted_avg_price::double precision,
    ALTER COLUMN receivable_amount  TYPE double precision USING receivable_amount::double precision,
    ALTER COLUMN actual_amount      TYPE double precision USING actual_amount::double precision,
    ALTER COLUMN penalty_amount     TYPE double precision USING penalty_amount::double precision;

ALTER TABLE batch_monthly_settlement
    ALTER COLUMN energy_fee     TYPE double precision USING energy_fee::double precision,
    ALTER COLUMN capacity_fee   TYPE double precision USING capacity_fee::double precision,
    ALTER COLUMN ancillary_fee  TYPE double precision USING ancillary_fee::double precision,
    ALTER COLUMN policy_subsidy TYPE double precision USING policy_subsidy::double precision,
    ALTER COLUMN total_fee      TYPE double precision USING total_fee::double precision;
