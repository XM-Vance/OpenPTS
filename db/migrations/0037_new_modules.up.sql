-- 0037: 新增模块表 — 签约进度、偏差结算、绿电交易、滚动撮合交易、现货市场、虚拟电厂、竞价管理、负荷特性、客户分析、交易策略、大屏聚合

-- 签约进度跟踪
CREATE TABLE IF NOT EXISTS contract_progress (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    contract_id UUID NOT NULL REFERENCES retail_contracts(id),
    operating_month VARCHAR(7) NOT NULL,
    planned_energy_mwh DOUBLE PRECISION NOT NULL DEFAULT 0,
    actual_energy_mwh DOUBLE PRECISION NOT NULL DEFAULT 0,
    completion_rate DOUBLE PRECISION NOT NULL DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'on_track',
    milestones JSONB DEFAULT '{}',
    note TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (contract_id, operating_month)
);

-- 偏差结算
CREATE TABLE IF NOT EXISTS deviation_settlement (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    operating_date DATE NOT NULL,
    declared_energy_mwh DOUBLE PRECISION NOT NULL DEFAULT 0,
    actual_energy_mwh DOUBLE PRECISION NOT NULL DEFAULT 0,
    deviation_energy_mwh DOUBLE PRECISION NOT NULL DEFAULT 0,
    deviation_rate DOUBLE PRECISION NOT NULL DEFAULT 0,
    deviation_cost DOUBLE PRECISION NOT NULL DEFAULT 0,
    penalty_cost DOUBLE PRECISION NOT NULL DEFAULT 0,
    total_settlement DOUBLE PRECISION NOT NULL DEFAULT 0,
    category VARCHAR(20) NOT NULL DEFAULT 'day_ahead',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (operating_date, category)
);

-- 绿电交易
CREATE TABLE IF NOT EXISTS green_power_trades (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    trade_date DATE NOT NULL,
    product_name VARCHAR(50) NOT NULL,
    energy_mwh DOUBLE PRECISION NOT NULL DEFAULT 0,
    price DOUBLE PRECISION NOT NULL DEFAULT 0,
    amount DOUBLE PRECISION NOT NULL DEFAULT 0,
    green_cert_count INTEGER NOT NULL DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    counterparty VARCHAR(100) NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- 滚动撮合交易
CREATE TABLE IF NOT EXISTS rolling_trades (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    trade_date DATE NOT NULL,
    trade_session INTEGER NOT NULL,
    side VARCHAR(10) NOT NULL DEFAULT 'buy',
    energy_mw DOUBLE PRECISION NOT NULL DEFAULT 0,
    declared_price DOUBLE PRECISION NOT NULL DEFAULT 0,
    cleared_price DOUBLE PRECISION NOT NULL DEFAULT 0,
    cleared_energy_mw DOUBLE PRECISION NOT NULL DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- 现货市场日数据
CREATE TABLE IF NOT EXISTS spot_market_daily (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    trade_date DATE NOT NULL UNIQUE,
    day_ahead_avg DOUBLE PRECISION NOT NULL DEFAULT 0,
    day_ahead_high DOUBLE PRECISION NOT NULL DEFAULT 0,
    day_ahead_low DOUBLE PRECISION NOT NULL DEFAULT 0,
    real_time_avg DOUBLE PRECISION NOT NULL DEFAULT 0,
    real_time_high DOUBLE PRECISION NOT NULL DEFAULT 0,
    real_time_low DOUBLE PRECISION NOT NULL DEFAULT 0,
    total_volume_mwh DOUBLE PRECISION NOT NULL DEFAULT 0,
    spread DOUBLE PRECISION NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- 虚拟电厂资源
CREATE TABLE IF NOT EXISTS vpp_resources (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    resource_name VARCHAR(100) NOT NULL UNIQUE,
    resource_type VARCHAR(30) NOT NULL DEFAULT 'storage',
    capacity_mw DOUBLE PRECISION NOT NULL DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    location VARCHAR(100) NOT NULL DEFAULT '',
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- 虚拟电厂调度记录
CREATE TABLE IF NOT EXISTS vpp_dispatches (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    dispatch_date DATE NOT NULL,
    resource_id UUID NOT NULL REFERENCES vpp_resources(id),
    dispatch_type VARCHAR(30) NOT NULL DEFAULT 'peak_shaving',
    dispatched_mw DOUBLE PRECISION NOT NULL DEFAULT 0,
    duration_min INTEGER NOT NULL DEFAULT 0,
    response_time_sec INTEGER NOT NULL DEFAULT 0,
    revenue DOUBLE PRECISION NOT NULL DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'completed',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- 竞价管理
CREATE TABLE IF NOT EXISTS bidding_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    trade_date DATE NOT NULL,
    bidding_session VARCHAR(20) NOT NULL,
    declared_mw DOUBLE PRECISION NOT NULL DEFAULT 0,
    declared_price DOUBLE PRECISION NOT NULL DEFAULT 0,
    cleared_mw DOUBLE PRECISION NOT NULL DEFAULT 0,
    cleared_price DOUBLE PRECISION NOT NULL DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    strategy VARCHAR(30) NOT NULL DEFAULT 'balanced',
    bid_curve JSONB,
    note TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (trade_date, bidding_session)
);

-- 负荷特性分析
CREATE TABLE IF NOT EXISTS load_characteristics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id UUID NOT NULL REFERENCES customers(id),
    analysis_month VARCHAR(7) NOT NULL,
    avg_daily_mwh DOUBLE PRECISION NOT NULL DEFAULT 0,
    peak_mw DOUBLE PRECISION NOT NULL DEFAULT 0,
    valley_mw DOUBLE PRECISION NOT NULL DEFAULT 0,
    peak_valley_ratio DOUBLE PRECISION NOT NULL DEFAULT 0,
    load_factor DOUBLE PRECISION NOT NULL DEFAULT 0,
    peak_hours DOUBLE PRECISION NOT NULL DEFAULT 0,
    load_type VARCHAR(20) NOT NULL DEFAULT 'industrial',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (customer_id, analysis_month)
);

-- 客户分析
CREATE TABLE IF NOT EXISTS customer_analysis (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id UUID NOT NULL REFERENCES customers(id),
    analysis_month VARCHAR(7) NOT NULL,
    energy_consumed_mwh DOUBLE PRECISION NOT NULL DEFAULT 0,
    bill_amount DOUBLE PRECISION NOT NULL DEFAULT 0,
    avg_unit_price DOUBLE PRECISION NOT NULL DEFAULT 0,
    peak_ratio DOUBLE PRECISION NOT NULL DEFAULT 0,
    score DOUBLE PRECISION NOT NULL DEFAULT 0,
    risk_level VARCHAR(10) NOT NULL DEFAULT 'low',
    tags TEXT[] DEFAULT '{}',
    extra JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (customer_id, analysis_month)
);

-- 交易策略
CREATE TABLE IF NOT EXISTS trade_strategies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    strategy_name VARCHAR(100) NOT NULL UNIQUE,
    strategy_type VARCHAR(30) NOT NULL DEFAULT 'arbitrage',
    target_market VARCHAR(30) NOT NULL DEFAULT 'day_ahead',
    parameters JSONB DEFAULT '{}',
    status VARCHAR(20) NOT NULL DEFAULT 'draft',
    backtest_return DOUBLE PRECISION,
    backtest_sharpe DOUBLE PRECISION,
    win_rate DOUBLE PRECISION,
    total_trades INTEGER NOT NULL DEFAULT 0,
    note TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- 索引
CREATE INDEX IF NOT EXISTS idx_contract_progress_month ON contract_progress(operating_month);
CREATE INDEX IF NOT EXISTS idx_deviation_settlement_date ON deviation_settlement(operating_date);
CREATE INDEX IF NOT EXISTS idx_green_power_trades_date ON green_power_trades(trade_date);
CREATE INDEX IF NOT EXISTS idx_rolling_trades_date ON rolling_trades(trade_date);
CREATE INDEX IF NOT EXISTS idx_vpp_dispatches_date ON vpp_dispatches(dispatch_date);
CREATE INDEX IF NOT EXISTS idx_vpp_dispatches_resource ON vpp_dispatches(resource_id);
CREATE INDEX IF NOT EXISTS idx_bidding_records_date ON bidding_records(trade_date);
CREATE INDEX IF NOT EXISTS idx_load_characteristics_month ON load_characteristics(analysis_month);
CREATE INDEX IF NOT EXISTS idx_customer_analysis_month ON customer_analysis(analysis_month);
CREATE INDEX IF NOT EXISTS idx_trade_strategies_type ON trade_strategies(strategy_type);
