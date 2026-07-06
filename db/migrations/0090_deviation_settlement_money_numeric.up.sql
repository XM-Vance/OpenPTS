-- P4 stage-one：偏差结算金额列 double precision → numeric(18,4)。
-- 消除浮点漂移，使「偏差费 + 考核费 == 总结算」与 SUM 聚合精确成立。
-- 仅金额列；电量(_mwh)/偏差率(rate) 保持 double precision（物理量/比率，非账目）。
-- float → numeric 为放大精度，存量数据无损。
ALTER TABLE deviation_settlement
    ALTER COLUMN deviation_cost   TYPE numeric(18,4) USING deviation_cost::numeric(18,4),
    ALTER COLUMN penalty_cost     TYPE numeric(18,4) USING penalty_cost::numeric(18,4),
    ALTER COLUMN total_settlement TYPE numeric(18,4) USING total_settlement::numeric(18,4);
