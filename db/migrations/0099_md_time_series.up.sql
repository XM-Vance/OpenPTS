-- 99: 统一时序数据表 md_time_series（EAV 模型）—— md_* 多表扩展性问题的目标架构。
--
-- 背景（审查报告 3.7）：当前有 30 张 md_* 表（md_futures_cu/md_macro_gdp/...），
-- 每加一个品种就加一张表，不可维护。本迁移引入统一 EAV 表作为目标存储：
--   md_time_series(instrument_code, date, value, ...)
-- 一张表承载所有时序品种，新增品种只需 INSERT，无需改 schema。
--
-- 本迁移为「降级方案」第一步：仅建表，不迁移存量数据、不改动旧表与现有读取代码。
-- 旧的 30 张 md_* 表与 market_data_repository.go 保持原样运行（向后兼容）。
-- 后续分阶段把读写切到本表，详见 docs/md_timeseries_migration.md。
--
-- 字段说明：
--   instrument_code  品种代码（如 cu/gdp/wti/dxy），等价于旧表名去掉 md_<category>_ 前缀
--   category         分类（macro/fuel/futures/rate/fx/index/carbon/weather）
--   obs_date         观测日期（日频）或观测时刻（含时间）
--   frequency        频率（daily/monthly/quarterly/...）
--   value            观测值（统一 numeric，旧表各列 open/high/.../cpi_yoy 等统一进这里）
--   attrs            额外属性（如开高低收量、同比/环比等），用 JSONB 兜底多值品种
--   source           数据来源（wind/choice/manual/...）

CREATE TABLE IF NOT EXISTS md_time_series (
    id              BIGSERIAL PRIMARY KEY,
    instrument_code VARCHAR(64)  NOT NULL,
    category        VARCHAR(32)  NOT NULL,
    obs_date        DATE         NOT NULL,
    frequency       VARCHAR(16)  NOT NULL DEFAULT 'daily',
    value           NUMERIC(24, 8),
    attrs           JSONB,
    source          VARCHAR(32),
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT now()
);

-- 按品种查时序的主索引（含日期序），覆盖 99% 查询。
CREATE INDEX IF NOT EXISTS idx_md_ts_code_date
    ON md_time_series(instrument_code, obs_date DESC);
-- 按分类筛选所有品种。
CREATE INDEX IF NOT EXISTS idx_md_ts_category
    ON md_time_series(category, obs_date DESC);
-- 去重：同一品种同一日期同一频率一条记录。
CREATE UNIQUE INDEX IF NOT EXISTS uq_md_ts_code_date_freq
    ON md_time_series(instrument_code, obs_date, frequency);
