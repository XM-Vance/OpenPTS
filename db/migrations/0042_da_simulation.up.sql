-- 0042: 日前模拟模块

-- 模拟场景表
CREATE TABLE IF NOT EXISTS da_simulation_scenarios (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(200) NOT NULL,
    description TEXT DEFAULT '',
    sim_date DATE NOT NULL,
    total_volume_mwh DOUBLE PRECISION NOT NULL DEFAULT 0,
    avg_price DOUBLE PRECISION NOT NULL DEFAULT 0,
    total_cost DOUBLE PRECISION NOT NULL DEFAULT 0,
    profit DOUBLE PRECISION NOT NULL DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'submitted', 'settled')),
    created_by UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_da_sim_scenarios_date ON da_simulation_scenarios(sim_date);
CREATE INDEX IF NOT EXISTS idx_da_sim_scenarios_status ON da_simulation_scenarios(status);

-- 模拟时段结果表
CREATE TABLE IF NOT EXISTS da_simulation_period_results (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    scenario_id UUID NOT NULL REFERENCES da_simulation_scenarios(id) ON DELETE CASCADE,
    period INT NOT NULL CHECK (period >= 1 AND period <= 96),
    declared_volume_mwh DOUBLE PRECISION NOT NULL DEFAULT 0,
    simulated_price DOUBLE PRECISION NOT NULL DEFAULT 0,
    simulated_cost DOUBLE PRECISION NOT NULL DEFAULT 0,
    spot_actual_price DOUBLE PRECISION,
    settlement_amount DOUBLE PRECISION NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_da_sim_period_scenario ON da_simulation_period_results(scenario_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_da_sim_period_uniq ON da_simulation_period_results(scenario_id, period);
