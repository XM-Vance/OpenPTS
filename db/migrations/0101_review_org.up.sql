-- 多租户隔离收口：月度交易复盘 / 日前交易复盘两张表此前漏加 org_id（0055 批量
-- 加 org_id 时未覆盖），唯一键仅 (operating_month) / (trading_date)，导致多省共享一行、
-- 跨省 ON CONFLICT 互相覆盖。本迁移补 org_id + 回填 FJ + 唯一键升级为含 org 维度。

-- monthly_trade_review（UNIQUE(operating_month) → UNIQUE(org_id, operating_month)）
ALTER TABLE monthly_trade_review ADD COLUMN IF NOT EXISTS org_id uuid REFERENCES organizations(id);
UPDATE monthly_trade_review SET org_id = (SELECT id FROM organizations WHERE code='FJ') WHERE org_id IS NULL;
CREATE INDEX IF NOT EXISTS idx_monthly_trade_review_org ON monthly_trade_review(org_id);
ALTER TABLE monthly_trade_review DROP CONSTRAINT IF EXISTS monthly_trade_review_operating_month_key;
CREATE UNIQUE INDEX IF NOT EXISTS monthly_trade_review_org_uniq ON monthly_trade_review(org_id, operating_month);

-- day_ahead_trade_review（UNIQUE(trading_date) → UNIQUE(org_id, trading_date)）
ALTER TABLE day_ahead_trade_review ADD COLUMN IF NOT EXISTS org_id uuid REFERENCES organizations(id);
UPDATE day_ahead_trade_review SET org_id = (SELECT id FROM organizations WHERE code='FJ') WHERE org_id IS NULL;
CREATE INDEX IF NOT EXISTS idx_day_ahead_trade_review_org ON day_ahead_trade_review(org_id);
ALTER TABLE day_ahead_trade_review DROP CONSTRAINT IF EXISTS day_ahead_trade_review_trading_date_key;
CREATE UNIQUE INDEX IF NOT EXISTS day_ahead_trade_review_org_uniq ON day_ahead_trade_review(org_id, trading_date);
