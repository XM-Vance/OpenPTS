-- 0004: 零售合同 + 套餐 + 定价模型
-- 覆盖 v1 集合：pricing_models / retail_packages / retail_contracts

-- ─── 定价模型 ──────────────────────────────────────
CREATE TABLE IF NOT EXISTS pricing_models (
    code           VARCHAR(64) PRIMARY KEY,
    display_name   VARCHAR(128) NOT NULL,
    package_type   VARCHAR(64) NOT NULL,                  -- 月度 / 月内 / 分时 / 其他
    pricing_mode   VARCHAR(64) NOT NULL,                  -- 固定 / 浮动 / 联动
    floating_type  VARCHAR(64),                           -- 上浮率 / 比例 / 公式
    formula_html   TEXT,                                  -- 渲染用 HTML 公式
    config         JSONB NOT NULL DEFAULT '{}'::jsonb,    -- 模型参数
    enabled        BOOLEAN NOT NULL DEFAULT TRUE,
    sort_order     INT NOT NULL DEFAULT 0,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_pricing_model_type ON pricing_models(package_type);
CREATE INDEX IF NOT EXISTS idx_pricing_model_enabled ON pricing_models(enabled);

DROP TRIGGER IF EXISTS pricing_models_set_updated_at ON pricing_models;
CREATE TRIGGER pricing_models_set_updated_at
    BEFORE UPDATE ON pricing_models
    FOR EACH ROW EXECUTE FUNCTION trg_set_updated_at();

-- ─── 零售套餐 ──────────────────────────────────────
CREATE TABLE IF NOT EXISTS retail_packages (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    package_name    VARCHAR(255) NOT NULL UNIQUE,
    package_type    VARCHAR(64) NOT NULL,
    model_code      VARCHAR(64) REFERENCES pricing_models(code) ON DELETE RESTRICT,
    pricing_config  JSONB NOT NULL DEFAULT '{}'::jsonb,    -- 价格参数（按 model_code 的 schema 解析）
    is_green_power  BOOLEAN NOT NULL DEFAULT FALSE,
    validation      JSONB NOT NULL DEFAULT '{}'::jsonb,    -- 约束规则
    status          VARCHAR(32) NOT NULL DEFAULT 'active', -- active / archived / draft
    description     TEXT,
    created_by      UUID REFERENCES users(id),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_retail_pkg_type ON retail_packages(package_type);
CREATE INDEX IF NOT EXISTS idx_retail_pkg_status ON retail_packages(status);
CREATE INDEX IF NOT EXISTS idx_retail_pkg_model ON retail_packages(model_code);

DROP TRIGGER IF EXISTS retail_packages_set_updated_at ON retail_packages;
CREATE TRIGGER retail_packages_set_updated_at
    BEFORE UPDATE ON retail_packages
    FOR EACH ROW EXECUTE FUNCTION trg_set_updated_at();

-- ─── 零售合同 ──────────────────────────────────────
CREATE TABLE IF NOT EXISTS retail_contracts (
    id                          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    customer_id                 UUID NOT NULL REFERENCES customers(id) ON DELETE RESTRICT,
    package_id                  UUID NOT NULL REFERENCES retail_packages(id) ON DELETE RESTRICT,
    package_name_snapshot       VARCHAR(255) NOT NULL,                  -- 签约时套餐名快照
    pricing_config_snapshot     JSONB NOT NULL DEFAULT '{}'::jsonb,     -- 签约时价格参数快照
    purchasing_energy_mwh       NUMERIC(18,4) NOT NULL,
    green_power_ratio           NUMERIC(5,2),                            -- 0.00 ~ 100.00
    purchase_start_month        VARCHAR(7) NOT NULL,                     -- YYYY-MM
    purchase_end_month          VARCHAR(7) NOT NULL,                     -- YYYY-MM
    signed_at                   TIMESTAMPTZ,
    effective_from              DATE,
    effective_to                DATE,
    status                      VARCHAR(32) NOT NULL DEFAULT 'active',   -- active / expired / terminated
    extra                       JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_by                  UUID REFERENCES users(id),
    created_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_retail_contract_period CHECK (purchase_start_month <= purchase_end_month)
);

CREATE INDEX IF NOT EXISTS idx_retail_contract_customer ON retail_contracts(customer_id);
CREATE INDEX IF NOT EXISTS idx_retail_contract_package ON retail_contracts(package_id);
CREATE INDEX IF NOT EXISTS idx_retail_contract_status ON retail_contracts(status);
CREATE INDEX IF NOT EXISTS idx_retail_contract_period ON retail_contracts(purchase_start_month, purchase_end_month);

DROP TRIGGER IF EXISTS retail_contracts_set_updated_at ON retail_contracts;
CREATE TRIGGER retail_contracts_set_updated_at
    BEFORE UPDATE ON retail_contracts
    FOR EACH ROW EXECUTE FUNCTION trg_set_updated_at();
