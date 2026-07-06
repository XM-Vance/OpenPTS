-- 多租户隔离收口（批次一）：结算/审批/撮合域核心表补 org_id。
-- 这批表此前建表未带 org_id（0026/0029/0009），读写全表扫描导致跨租户数据互窜，
-- 是审查报告 P0-B1 / P0-D1 点名的真实跨租户泄露面。本迁移逐表：加 org_id 列 →
-- 回填 FJ → 建索引 → 唯一键升级为含 org 维度（避免多省 ON CONFLICT 互相覆盖）。
-- 样板参照 0101_review_org / 0058_da_settle_org。down 仅一次性库验证用。

-- rolling_match_quotes（无唯一约束，仅加列+回填+索引）
ALTER TABLE rolling_match_quotes ADD COLUMN IF NOT EXISTS org_id uuid REFERENCES organizations(id);
UPDATE rolling_match_quotes SET org_id = (SELECT id FROM organizations WHERE code='default') WHERE org_id IS NULL;
CREATE INDEX IF NOT EXISTS idx_rolling_match_quotes_org ON rolling_match_quotes(org_id);

-- monthly_manual_data（无唯一约束，仅加列+回填+索引）
ALTER TABLE monthly_manual_data ADD COLUMN IF NOT EXISTS org_id uuid REFERENCES organizations(id);
UPDATE monthly_manual_data SET org_id = (SELECT id FROM organizations WHERE code='default') WHERE org_id IS NULL;
CREATE INDEX IF NOT EXISTS idx_monthly_manual_data_org ON monthly_manual_data(org_id);

-- approval_requests（无唯一约束，仅加列+回填+索引；审批 payload 含敏感金额必须隔离）
ALTER TABLE approval_requests ADD COLUMN IF NOT EXISTS org_id uuid REFERENCES organizations(id);
UPDATE approval_requests SET org_id = (SELECT id FROM organizations WHERE code='default') WHERE org_id IS NULL;
CREATE INDEX IF NOT EXISTS idx_approval_requests_org ON approval_requests(org_id);

-- rolling_match_snapshots（列级 UNIQUE(trade_date, delivery_date) → 含 org）
ALTER TABLE rolling_match_snapshots ADD COLUMN IF NOT EXISTS org_id uuid REFERENCES organizations(id);
UPDATE rolling_match_snapshots SET org_id = (SELECT id FROM organizations WHERE code='default') WHERE org_id IS NULL;
CREATE INDEX IF NOT EXISTS idx_rolling_match_snapshots_org ON rolling_match_snapshots(org_id);
ALTER TABLE rolling_match_snapshots DROP CONSTRAINT IF EXISTS rolling_match_snapshots_trade_date_delivery_date_key;
CREATE UNIQUE INDEX IF NOT EXISTS rolling_match_snapshots_org_uniq ON rolling_match_snapshots(org_id, trade_date, delivery_date);

-- spot_settlement_daily（列级 settlement_date UNIQUE → 含 org）
ALTER TABLE spot_settlement_daily ADD COLUMN IF NOT EXISTS org_id uuid REFERENCES organizations(id);
UPDATE spot_settlement_daily SET org_id = (SELECT id FROM organizations WHERE code='default') WHERE org_id IS NULL;
CREATE INDEX IF NOT EXISTS idx_spot_settlement_daily_org ON spot_settlement_daily(org_id);
ALTER TABLE spot_settlement_daily DROP CONSTRAINT IF EXISTS spot_settlement_daily_settlement_date_key;
CREATE UNIQUE INDEX IF NOT EXISTS spot_settlement_daily_org_uniq ON spot_settlement_daily(org_id, settlement_date);

-- spot_settlement_period（子句 UNIQUE(settlement_date, period) → 含 org）
ALTER TABLE spot_settlement_period ADD COLUMN IF NOT EXISTS org_id uuid REFERENCES organizations(id);
UPDATE spot_settlement_period SET org_id = (SELECT id FROM organizations WHERE code='default') WHERE org_id IS NULL;
CREATE INDEX IF NOT EXISTS idx_spot_settlement_period_org ON spot_settlement_period(org_id);
ALTER TABLE spot_settlement_period DROP CONSTRAINT IF EXISTS spot_settlement_period_settlement_date_period_key;
CREATE UNIQUE INDEX IF NOT EXISTS spot_settlement_period_org_uniq ON spot_settlement_period(org_id, settlement_date, period);

-- contracts_aggregated_daily（列级 date UNIQUE → 含 org）
ALTER TABLE contracts_aggregated_daily ADD COLUMN IF NOT EXISTS org_id uuid REFERENCES organizations(id);
UPDATE contracts_aggregated_daily SET org_id = (SELECT id FROM organizations WHERE code='default') WHERE org_id IS NULL;
CREATE INDEX IF NOT EXISTS idx_contracts_aggregated_daily_org ON contracts_aggregated_daily(org_id);
ALTER TABLE contracts_aggregated_daily DROP CONSTRAINT IF EXISTS contracts_aggregated_daily_date_key;
CREATE UNIQUE INDEX IF NOT EXISTS contracts_aggregated_daily_org_uniq ON contracts_aggregated_daily(org_id, date);
