-- P4 stage-two：其余结算/账单域金额列 double precision → numeric(18,4)。
-- 仅金额列；电量(_mwh)、毛利率(gross_margin %)、活跃客户数等保持 float/int（非账目）。
-- float → numeric 为放大精度，存量数据无损。

-- F1 客户利润（gross_profit = revenue - cost 须精确）
ALTER TABLE customer_profit
    ALTER COLUMN revenue      TYPE numeric(18,4) USING revenue::numeric(18,4),
    ALTER COLUMN cost         TYPE numeric(18,4) USING cost::numeric(18,4),
    ALTER COLUMN gross_profit TYPE numeric(18,4) USING gross_profit::numeric(18,4);

-- F2 月度交易复盘（gross_profit = retail_revenue - wholesale_cost 须精确）
ALTER TABLE monthly_trade_review
    ALTER COLUMN retail_revenue TYPE numeric(18,4) USING retail_revenue::numeric(18,4),
    ALTER COLUMN wholesale_cost TYPE numeric(18,4) USING wholesale_cost::numeric(18,4),
    ALTER COLUMN gross_profit   TYPE numeric(18,4) USING gross_profit::numeric(18,4);

-- U1 零售月度结算（应收/实收/违约金/加权均价）
ALTER TABLE retail_monthly_settlement
    ALTER COLUMN weighted_avg_price TYPE numeric(18,4) USING weighted_avg_price::numeric(18,4),
    ALTER COLUMN receivable_amount  TYPE numeric(18,4) USING receivable_amount::numeric(18,4),
    ALTER COLUMN actual_amount      TYPE numeric(18,4) USING actual_amount::numeric(18,4),
    ALTER COLUMN penalty_amount     TYPE numeric(18,4) USING penalty_amount::numeric(18,4);

-- 批量月度结算（total_fee = energy_fee + capacity_fee + ancillary_fee 须精确）
ALTER TABLE batch_monthly_settlement
    ALTER COLUMN energy_fee     TYPE numeric(18,4) USING energy_fee::numeric(18,4),
    ALTER COLUMN capacity_fee   TYPE numeric(18,4) USING capacity_fee::numeric(18,4),
    ALTER COLUMN ancillary_fee  TYPE numeric(18,4) USING ancillary_fee::numeric(18,4),
    ALTER COLUMN policy_subsidy TYPE numeric(18,4) USING policy_subsidy::numeric(18,4),
    ALTER COLUMN total_fee      TYPE numeric(18,4) USING total_fee::numeric(18,4);
