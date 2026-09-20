-- 156: 负荷数据校准（WP6.4）——计量口径 vs 结算口径的系数台账。
-- 恢复被 YAGNI 删除的校准功能（docs/api-contract-backlog.md:38）：基于
-- raw_meter_data 与 user_load_data 差异计算校准系数并回写（可追溯、可回滚）。

CREATE TABLE load_calibrations (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id         UUID NOT NULL REFERENCES organizations(id),
    customer_id    UUID NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    period_month   VARCHAR(7) NOT NULL,             -- YYYY-MM
    meter_kwh      DOUBLE PRECISION NOT NULL DEFAULT 0, -- 表计侧合计 kWh
    system_kwh     DOUBLE PRECISION NOT NULL DEFAULT 0, -- 系统侧合计 kWh
    coefficient    DOUBLE PRECISION NOT NULL DEFAULT 1, -- 校准系数 = meter/system
    status         VARCHAR(16) NOT NULL DEFAULT 'preview', -- preview/applied/void
    applied_at     TIMESTAMPTZ,
    applied_by     UUID REFERENCES users(id),
    note           TEXT,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (org_id, customer_id, period_month)
);

CREATE INDEX idx_load_calibrations_org_month ON load_calibrations(org_id, period_month DESC);
