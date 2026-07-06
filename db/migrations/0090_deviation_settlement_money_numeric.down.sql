-- 回滚：numeric(18,4) → double precision（仅一次性库验证可逆性时使用；会丢失精度）。
ALTER TABLE deviation_settlement
    ALTER COLUMN deviation_cost   TYPE double precision USING deviation_cost::double precision,
    ALTER COLUMN penalty_cost     TYPE double precision USING penalty_cost::double precision,
    ALTER COLUMN total_settlement TYPE double precision USING total_settlement::double precision;
