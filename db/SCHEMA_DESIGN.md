# 数据库 Schema 设计规范

## 一、命名规范

| 对象 | 规则 | 示例 |
|---|---|---|
| 表名 | snake_case 复数 | `customers`, `retail_contracts` |
| 主键 | `id`（UUID） | - |
| 外键 | `<引用表单数>_id` | `customer_id`, `package_id` |
| 时间戳 | `created_at` / `updated_at` | 类型必须 `TIMESTAMPTZ` |
| 状态字段 | `status` 或 `is_xxx` | `status VARCHAR(32)` / `is_active BOOLEAN` |
| 业务编码 | `code` 或 `<前缀>_code` | `role_code`, `module_code` |

## 二、字段类型对照（v1 MongoDB → v2 PostgreSQL）

| MongoDB | PostgreSQL | 说明 |
|---|---|---|
| ObjectId | `UUID` | 用 `uuid_generate_v4()` 生成 |
| String | `VARCHAR(N)` 或 `TEXT` | 已知长度用 VARCHAR，自由文本用 TEXT |
| Number | `NUMERIC(p,s)` / `INT` | 金额、电量统一 `NUMERIC(18,4)`；占比 `NUMERIC(5,2)` |
| Boolean | `BOOLEAN` | - |
| Date | `DATE` | 仅日期（如结算日） |
| ISODate | `TIMESTAMPTZ` | 含时区，**禁止 `TIMESTAMP`** |
| Array (同构) | `TEXT[]` / `NUMERIC[]` | 标签、列表 |
| Array (异构) | `JSONB` | period_details[48]、tags[] 含对象 |
| Object | `JSONB` | 嵌套结构，例如 `accounts`, `pricing_config` |

## 三、时间字段规则（强制）

- 全部使用 `TIMESTAMPTZ`（带时区）；不使用裸 `TIMESTAMP`
- 默认值 `NOW()`（数据库时区，部署在 Asia/Shanghai）
- `updated_at` 通过触发器 `trg_set_updated_at()` 自动维护（见 0001_init）

## 四、JSONB 的使用场景

**推荐用 JSONB**：
- 字段结构会演化的嵌套对象（如 `accounts[]` 内嵌 meters[]）
- 计算结果（如 48 点曲线 `curve_data`）
- 第三方配置（pricing_config、validation 规则）
- 临时扩展字段（`extra` 槽位）

**禁止用 JSONB**：
- 需要被频繁 join / where 的关键业务字段
- 数值聚合的金额、电量（必须 NUMERIC 单独列）
- 状态枚举（必须 VARCHAR 单独列）

## 五、索引策略

- 所有外键字段必须建索引
- 时间序列表必须建 `(customer_id, date)` 或 `(date)` 复合索引
- 数组字段查询频繁的用 GIN 索引（如 `tags` / `accounts`）
- 唯一约束用 UNIQUE 显式声明

## 六、外键删除策略

| 场景 | 策略 |
|---|---|
| 父子强一致（用户 ↔ 用户角色） | `ON DELETE CASCADE` |
| 业务实体（客户 ↔ 合同） | `ON DELETE RESTRICT`（删客户前必须解绑合同） |
| 快照字段（合同记录套餐名） | 不做外键 + 冗余字段 |

## 七、迁移文件命名

```
NNNN_<业务域>.up.sql        前进
NNNN_<业务域>.down.sql      回滚
```

- `NNNN`：4 位序号，从 0001 起
- 每个 up.sql 必须用 `CREATE TABLE IF NOT EXISTS` / `CREATE INDEX IF NOT EXISTS`
- 每个 down.sql 必须用 `DROP ... IF EXISTS`，避免回滚时报错

## 八、迁移路线图（基于 v1 数据模型反推）

| 序号 | 业务域 | 关键表 | 状态 |
|---|---|---|---|
| 0001 | 初始化 | users（最小版）+ 函数 + 扩展 | ✓ 已有 |
| 0002 | 用户权限 | auth_modules, auth_permissions, auth_roles, user_roles, role_permissions, auth_sessions, auth_security_challenges, auth_email_challenges, auth_trusted_devices | **本批次** |
| 0003 | 客户档案 | customers, customer_demo_aliases, intent_customers | **本批次** |
| 0004 | 零售合同 | pricing_models, retail_packages, retail_contracts | **本批次** |
| 0005 | 负荷（原始） | raw_meter_data, raw_mp_data（**需分区**） | 待批次 B |
| 0006 | 负荷（处理） | unified_load_curve, user_load_data, customer_monthly_energy, intent_customer_load_curve_daily, intent_customer_meter_reads_daily | 待批次 B |
| 0007 | 价格 | real_time_spot_price, day_ahead_spot_price, day_ahead_econ_price, price_sgcc, node_spot_price_daily | 待批次 B |
| 0008 | 预测 | price_forecast_results, medium_term_load_forecast | 待批次 B |
| 0009 | 结算 | settlement_daily, spot_settlement_daily, spot_settlement_period, contracts_aggregated_daily, retail_settlement_daily, retail_settlement_prices | 待批次 B |
| 0010 | 特征与告警 | customer_characteristics, customer_anomaly_alerts, analysis_history_log | 待批次 B |
| 0011 | 调频与储能 | frequency_regulation_clearing, frequency_regulation_demand, freq_comp_fee | 待批次 B |
| 0012 | 气象与规则 | weather_locations, weather_actuals, weather_forecasts, tou_rules, mechanism_energy_monthly | 待批次 B |
| 0013 | 任务调度 | task_commands, task_execution_records, task_execution_logs, daily_release | 待批次 B |
| 0014 | 杂项 | system_alerts, real_time_generation, intent_customer_monthly_retail_simulation, intent_customer_monthly_wholesale, rolling_match_snapshots, actual_operation | 待批次 B |

## 九、特殊设计：高频时序表（分区）

`raw_meter_data`、`raw_mp_data`、`real_time_spot_price`、`task_execution_logs` 等表预期日增量很大，将使用 PostgreSQL **声明式分区**：

```sql
CREATE TABLE raw_meter_data (
    ...
    date DATE NOT NULL,
    ...
) PARTITION BY RANGE (date);

-- 按月分区
CREATE TABLE raw_meter_data_y2026m01 PARTITION OF raw_meter_data
    FOR VALUES FROM ('2026-01-01') TO ('2026-02-01');
```

阶段 4 部署时引入 `pg_partman` 自动管理。

## 十、特殊设计：结算表的 period_details

v1 的 `settlement_daily.period_details[48]` 每个元素含 `{contract, day_ahead, real_time}` 等嵌套对象。

**v2 策略**：保留 JSONB 存储 `period_details`，关键聚合字段单独列出便于 SQL 查询：
- 顶层列：`total_energy_fee`, `contract_fee`, `day_ahead_fee`, `real_time_fee`, `energy_avg_price`
- 明细：`period_details JSONB`（48 个元素）

这样既保留了"按时段查询明细"的能力（JSONB 索引），又支持"按日/月聚合金额"的高效 SQL。
