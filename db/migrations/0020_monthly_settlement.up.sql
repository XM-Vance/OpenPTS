-- 月度结算汇总：每月一行，电量电费 + 政策项。
-- 演示数据：batch_monthly_settlement 表 + 12 个月示例。

CREATE TABLE IF NOT EXISTS batch_monthly_settlement (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    operating_month text NOT NULL UNIQUE,         -- YYYY-MM
    settled_energy_mwh double precision NOT NULL DEFAULT 0,
    energy_fee        double precision NOT NULL DEFAULT 0,    -- 电能量电费
    capacity_fee      double precision NOT NULL DEFAULT 0,    -- 容量电费
    ancillary_fee     double precision NOT NULL DEFAULT 0,    -- 辅助服务分摊
    policy_subsidy    double precision NOT NULL DEFAULT 0,    -- 政策补贴
    total_fee         double precision NOT NULL DEFAULT 0,    -- 上述合计（不含补贴）
    version           text NOT NULL DEFAULT 'PRELIMINARY',
    created_at        timestamp NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_batch_monthly_settlement_month
    ON batch_monthly_settlement(operating_month DESC);
