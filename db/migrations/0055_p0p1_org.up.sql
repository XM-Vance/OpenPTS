-- 任务 #5: p0/p1 表 org_id
-- pre_settlement_daily, total_load_daily, forecast_accuracy, mechanism_energy_plan,
-- medium_load_forecast, market_analysis_daily, holidays, typical_curves

-- pre_settlement_daily（有 UNIQUE(operating_date) → UNIQUE(org_id, operating_date)）
ALTER TABLE pre_settlement_daily ADD COLUMN IF NOT EXISTS org_id uuid REFERENCES organizations(id);
UPDATE pre_settlement_daily SET org_id = (SELECT id FROM organizations WHERE code='default') WHERE org_id IS NULL;
CREATE INDEX IF NOT EXISTS idx_pre_settlement_daily_org ON pre_settlement_daily(org_id);
ALTER TABLE pre_settlement_daily DROP CONSTRAINT IF EXISTS pre_settlement_daily_operating_date_key;
CREATE UNIQUE INDEX IF NOT EXISTS pre_settlement_daily_org_uniq ON pre_settlement_daily(org_id, operating_date);

-- total_load_daily（有 UNIQUE(data_date) → UNIQUE(org_id, data_date)）
ALTER TABLE total_load_daily ADD COLUMN IF NOT EXISTS org_id uuid REFERENCES organizations(id);
UPDATE total_load_daily SET org_id = (SELECT id FROM organizations WHERE code='default') WHERE org_id IS NULL;
CREATE INDEX IF NOT EXISTS idx_total_load_daily_org ON total_load_daily(org_id);
ALTER TABLE total_load_daily DROP CONSTRAINT IF EXISTS total_load_daily_data_date_key;
CREATE UNIQUE INDEX IF NOT EXISTS total_load_daily_org_uniq ON total_load_daily(org_id, data_date);

-- forecast_accuracy（有 UNIQUE(forecast_target, forecast_date) → UNIQUE(org_id, forecast_target, forecast_date)）
ALTER TABLE forecast_accuracy ADD COLUMN IF NOT EXISTS org_id uuid REFERENCES organizations(id);
UPDATE forecast_accuracy SET org_id = (SELECT id FROM organizations WHERE code='default') WHERE org_id IS NULL;
CREATE INDEX IF NOT EXISTS idx_forecast_accuracy_org ON forecast_accuracy(org_id);
ALTER TABLE forecast_accuracy DROP CONSTRAINT IF EXISTS forecast_accuracy_forecast_target_forecast_date_key;
CREATE UNIQUE INDEX IF NOT EXISTS forecast_accuracy_org_uniq ON forecast_accuracy(org_id, forecast_target, forecast_date);

-- mechanism_energy_plan（有 UNIQUE(operating_month, voltage_level) → UNIQUE(org_id, operating_month, voltage_level)）
ALTER TABLE mechanism_energy_plan ADD COLUMN IF NOT EXISTS org_id uuid REFERENCES organizations(id);
UPDATE mechanism_energy_plan SET org_id = (SELECT id FROM organizations WHERE code='default') WHERE org_id IS NULL;
CREATE INDEX IF NOT EXISTS idx_mechanism_energy_plan_org ON mechanism_energy_plan(org_id);
ALTER TABLE mechanism_energy_plan DROP CONSTRAINT IF EXISTS mechanism_energy_plan_operating_month_voltage_level_key;
CREATE UNIQUE INDEX IF NOT EXISTS mechanism_energy_plan_org_uniq ON mechanism_energy_plan(org_id, operating_month, voltage_level);

-- medium_load_forecast（有 UNIQUE(forecast_month) → UNIQUE(org_id, forecast_month)）
ALTER TABLE medium_load_forecast ADD COLUMN IF NOT EXISTS org_id uuid REFERENCES organizations(id);
UPDATE medium_load_forecast SET org_id = (SELECT id FROM organizations WHERE code='default') WHERE org_id IS NULL;
CREATE INDEX IF NOT EXISTS idx_medium_load_forecast_org ON medium_load_forecast(org_id);
ALTER TABLE medium_load_forecast DROP CONSTRAINT IF EXISTS medium_load_forecast_forecast_month_key;
CREATE UNIQUE INDEX IF NOT EXISTS medium_load_forecast_org_uniq ON medium_load_forecast(org_id, forecast_month);

-- market_analysis_daily（有 UNIQUE(trade_date) → UNIQUE(org_id, trade_date)）
ALTER TABLE market_analysis_daily ADD COLUMN IF NOT EXISTS org_id uuid REFERENCES organizations(id);
UPDATE market_analysis_daily SET org_id = (SELECT id FROM organizations WHERE code='default') WHERE org_id IS NULL;
CREATE INDEX IF NOT EXISTS idx_market_analysis_daily_org ON market_analysis_daily(org_id);
ALTER TABLE market_analysis_daily DROP CONSTRAINT IF EXISTS market_analysis_daily_trade_date_key;
CREATE UNIQUE INDEX IF NOT EXISTS market_analysis_daily_org_uniq ON market_analysis_daily(org_id, trade_date);

-- holidays（有 UNIQUE(holiday_date) → UNIQUE(org_id, holiday_date)）
ALTER TABLE holidays ADD COLUMN IF NOT EXISTS org_id uuid REFERENCES organizations(id);
UPDATE holidays SET org_id = (SELECT id FROM organizations WHERE code='default') WHERE org_id IS NULL;
CREATE INDEX IF NOT EXISTS idx_holidays_org ON holidays(org_id);
ALTER TABLE holidays DROP CONSTRAINT IF EXISTS holidays_holiday_date_key;
CREATE UNIQUE INDEX IF NOT EXISTS holidays_org_uniq ON holidays(org_id, holiday_date);

-- typical_curves（有 UNIQUE(name) → UNIQUE(org_id, name)）
ALTER TABLE typical_curves ADD COLUMN IF NOT EXISTS org_id uuid REFERENCES organizations(id);
UPDATE typical_curves SET org_id = (SELECT id FROM organizations WHERE code='default') WHERE org_id IS NULL;
CREATE INDEX IF NOT EXISTS idx_typical_curves_org ON typical_curves(org_id);
ALTER TABLE typical_curves DROP CONSTRAINT IF EXISTS typical_curves_name_key;
CREATE UNIQUE INDEX IF NOT EXISTS typical_curves_org_uniq ON typical_curves(org_id, name);
