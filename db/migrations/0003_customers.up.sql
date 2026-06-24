-- 0003: 客户档案 + 演示别名 + 意向客户
-- 覆盖 v1 集合：customer_archives / customer_demo_aliases / intent_customer_profiles

-- ─── 客户档案 ──────────────────────────────────────
CREATE TABLE IF NOT EXISTS customers (
    id            UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_name     VARCHAR(255) NOT NULL,                  -- 客户全称
    short_name    VARCHAR(255),                           -- 简称
    location      VARCHAR(255),                           -- 关联 weather_locations.name
    source        VARCHAR(64),                            -- 来源：开发/导入/外部
    manager       VARCHAR(64),                            -- 客户经理
    tags          TEXT[] NOT NULL DEFAULT '{}',
    accounts      JSONB NOT NULL DEFAULT '[]'::jsonb,     -- [{account_id, meters[...]}]
    is_demo       BOOLEAN NOT NULL DEFAULT FALSE,         -- 演示客户标记
    extra         JSONB NOT NULL DEFAULT '{}'::jsonb,     -- 可扩展字段
    created_by    UUID REFERENCES users(id),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_customers_user_name ON customers(user_name);
CREATE INDEX IF NOT EXISTS idx_customers_short_name ON customers(short_name);
CREATE INDEX IF NOT EXISTS idx_customers_location ON customers(location);
CREATE INDEX IF NOT EXISTS idx_customers_manager ON customers(manager);
CREATE INDEX IF NOT EXISTS idx_customers_tags ON customers USING GIN(tags);
CREATE INDEX IF NOT EXISTS idx_customers_accounts ON customers USING GIN(accounts);

DROP TRIGGER IF EXISTS customers_set_updated_at ON customers;
CREATE TRIGGER customers_set_updated_at
    BEFORE UPDATE ON customers
    FOR EACH ROW EXECUTE FUNCTION trg_set_updated_at();

-- ─── 演示客户别名（脱敏映射）───────────────────────
CREATE TABLE IF NOT EXISTS customer_demo_aliases (
    customer_id    UUID PRIMARY KEY REFERENCES customers(id) ON DELETE CASCADE,
    display_name   VARCHAR(255) NOT NULL,
    original_name  VARCHAR(255) NOT NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_demo_alias_display ON customer_demo_aliases(display_name);

-- ─── 意向客户（尚未签约，仅诊断分析）───────────────
CREATE TABLE IF NOT EXISTS intent_customers (
    id                UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    customer_name     VARCHAR(255) NOT NULL,
    meters            JSONB NOT NULL DEFAULT '[]'::jsonb,    -- [{meter_id, name, multiplier}]
    coverage_start    DATE,
    coverage_end      DATE,
    coverage_days     INT,
    completeness      NUMERIC(5,2),                           -- 0.00 ~ 100.00
    avg_daily_load    NUMERIC(14,4),
    status            VARCHAR(32) NOT NULL DEFAULT 'pending', -- pending / converted / archived
    converted_to      UUID REFERENCES customers(id),          -- 转正后指向正式客户
    extra             JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_intent_customer_name ON intent_customers(customer_name);
CREATE INDEX IF NOT EXISTS idx_intent_customer_status ON intent_customers(status);
CREATE INDEX IF NOT EXISTS idx_intent_customer_coverage ON intent_customers(coverage_start, coverage_end);

DROP TRIGGER IF EXISTS intent_customers_set_updated_at ON intent_customers;
CREATE TRIGGER intent_customers_set_updated_at
    BEFORE UPDATE ON intent_customers
    FOR EACH ROW EXECUTE FUNCTION trg_set_updated_at();
