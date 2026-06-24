-- 预测结果表（负荷预测 + 电价预测 + PV出力预测 + 储能调度）
-- 统一表结构，通过 type 区分预测类型
CREATE TABLE IF NOT EXISTS forecasts (
    id SERIAL PRIMARY KEY,
    type VARCHAR(30) NOT NULL,              -- load / price_day_ahead / price_realtime / pv_output / storage_schedule
    region VARCHAR(100) DEFAULT '',         -- 区域/省份
    target_date DATE NOT NULL,              -- 预测目标日期
    target_hour SMALLINT DEFAULT 0,         -- 预测目标小时(0-23)，日内预测用
    horizon_hours SMALLINT DEFAULT 24,      -- 预测时长(小时)
    value DOUBLE PRECISION NOT NULL,        -- 预测值
    lower_bound DOUBLE PRECISION,           -- 置信下界
    upper_bound DOUBLE PRECISION,           -- 置信上界
    actual_value DOUBLE PRECISION,          -- 实际值(回填)
    model_id VARCHAR(100) DEFAULT '',       -- 模型标识
    model_params JSONB DEFAULT '{}',        -- 模型参数
    features JSONB DEFAULT '{}',            -- 特征重要性等
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_forecasts_type_date ON forecasts(type, target_date);
CREATE INDEX idx_forecasts_type_region ON forecasts(type, region, target_date);

-- 电网计算结果表（潮流 + OPF + 短路 + 状态估计 + N-1校验）
CREATE TABLE IF NOT EXISTS grid_calculations (
    id SERIAL PRIMARY KEY,
    calc_type VARCHAR(30) NOT NULL,         -- powerflow / opf / short_circuit / state_estimation / n1_check
    scenario_name VARCHAR(200) DEFAULT '',  -- 场景名称
    network_id VARCHAR(100) DEFAULT '',     -- 网络模型ID
    snapshot_time TIMESTAMPTZ,              -- 计算时刻
    input_data JSONB DEFAULT '{}',          -- 输入快照
    result_data JSONB NOT NULL DEFAULT '{}',-- 计算结果
    constraints_violated JSONB DEFAULT '[]',-- 违约束列表
    is_feasible BOOLEAN DEFAULT TRUE,       -- 是否可行
    solve_time_ms INTEGER DEFAULT 0,        -- 求解耗时
    solver VARCHAR(50) DEFAULT '',          -- 求解器名称
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_grid_calc_type ON grid_calculations(calc_type, snapshot_time);

-- 市场出清结果表
CREATE TABLE IF NOT EXISTS market_clearings (
    id SERIAL PRIMARY KEY,
    market_type VARCHAR(30) NOT NULL,       -- day_ahead / realtime / ancillary
    trade_date DATE NOT NULL,               -- 交易日期
    period SMALLINT NOT NULL,               -- 时段编号(1-96 或 1-24)
    clearing_price DOUBLE PRECISION,        -- 出清价格(元/MWh)
    clearing_volume DOUBLE PRECISION,       -- 出清电量(MWh)
    marginal_price DOUBLE PRECISION,        -- 节点边际电价(LMP)
    congestion_cost DOUBLE PRECISION DEFAULT 0, -- 阻塞费用
    total_buy_volume DOUBLE PRECISION DEFAULT 0,
    total_sell_volume DOUBLE PRECISION DEFAULT 0,
    network_id VARCHAR(100) DEFAULT '',
    solver VARCHAR(50) DEFAULT '',
    solve_time_ms INTEGER DEFAULT 0,
    result_detail JSONB DEFAULT '{}',       -- 详细出清结果
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_clearing_unique ON market_clearings(market_type, trade_date, period);

-- 碳排放记录表
CREATE TABLE IF NOT EXISTS carbon_emissions (
    id SERIAL PRIMARY KEY,
    entity_type VARCHAR(30) NOT NULL,       -- customer / generator / region
    entity_id VARCHAR(100) NOT NULL,        -- 实体ID
    entity_name VARCHAR(200) DEFAULT '',    -- 实体名称
    period_start TIMESTAMPTZ NOT NULL,      -- 统计周期开始
    period_end TIMESTAMPTZ NOT NULL,        -- 统计周期结束
    energy_consumption DOUBLE PRECISION DEFAULT 0,   -- 电量(kWh)
    emission_factor DOUBLE PRECISION DEFAULT 0,       -- 排放因子(kgCO2e/kWh)
    carbon_emission DOUBLE PRECISION DEFAULT 0,       -- 碳排放量(kgCO2e)
    green_energy_ratio DOUBLE PRECISION DEFAULT 0,    -- 绿电比例
    is_green_certificate BOOLEAN DEFAULT FALSE,       -- 是否有绿证
    source VARCHAR(50) DEFAULT 'calculated',          -- 数据来源
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_carbon_entity ON carbon_emissions(entity_type, entity_id);
CREATE INDEX idx_carbon_period ON carbon_emissions(period_start, period_end);

-- 电价方案表(分时电价/合同电价)
CREATE TABLE IF NOT EXISTS tariff_schemes (
    id SERIAL PRIMARY KEY,
    name VARCHAR(200) NOT NULL,             -- 方案名称
    tariff_type VARCHAR(30) NOT NULL,       -- tou / spot / contract / hybrid
    region VARCHAR(100) DEFAULT '',
    valid_from DATE NOT NULL,               -- 生效日期
    valid_through DATE NOT NULL,            -- 失效日期
    periods JSONB NOT NULL DEFAULT '[]',    -- 时段定义 [{type:"peak",start:"08:00",end:"11:00",price:1.2}, ...]
    holidays JSONB DEFAULT '[]',            -- 节假日加价规则
    base_price DOUBLE PRECISION DEFAULT 0,  -- 基准电价
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- 储能调度计划表
CREATE TABLE IF NOT EXISTS storage_schedules (
    id SERIAL PRIMARY KEY,
    storage_id VARCHAR(100) NOT NULL,       -- 储能设备ID
    schedule_date DATE NOT NULL,            -- 调度日期
    period SMALLINT NOT NULL,               -- 时段(1-96)
    action VARCHAR(20) NOT NULL,            -- charge / discharge / idle
    power_mw DOUBLE PRECISION DEFAULT 0,    -- 功率(MW)，正=放电，负=充电
    soc_target DOUBLE PRECISION,            -- 目标SOC(%)
    energy_mwh DOUBLE PRECISION DEFAULT 0,  -- 电量(MWh)
    price_charge DOUBLE PRECISION DEFAULT 0,-- 充电电价
    price_discharge DOUBLE PRECISION DEFAULT 0, -- 放电电价
    profit DOUBLE PRECISION DEFAULT 0,      -- 预估收益
    optimization_id VARCHAR(100) DEFAULT '',-- 优化任务ID
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_storage_schedule ON storage_schedules(storage_id, schedule_date);
