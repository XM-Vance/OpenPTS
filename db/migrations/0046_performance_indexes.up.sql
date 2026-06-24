-- 0046: 性能索引补齐
-- 为外键列（缺少索引）、频繁过滤列（created_at / status）、
-- 以及多租户查询模式（org_id）添加索引。
-- 所有索引使用 IF NOT EXISTS 确保幂等。

-- ─── 外键列索引补齐 ─────────────────────────────────
-- 以下 FK 列在原迁移中未建索引，按 JOIN / WHERE 频率排序。

-- customers.created_by（按创建人查客户）
CREATE INDEX IF NOT EXISTS idx_customers_created_by ON customers(created_by) WHERE created_by IS NOT NULL;

-- intent_customers.converted_to（查转正关联）
CREATE INDEX IF NOT EXISTS idx_intent_customer_converted ON intent_customers(converted_to) WHERE converted_to IS NOT NULL;

-- retail_packages.created_by
CREATE INDEX IF NOT EXISTS idx_retail_pkg_created_by ON retail_packages(created_by) WHERE created_by IS NOT NULL;

-- retail_contracts.created_by
CREATE INDEX IF NOT EXISTS idx_retail_contract_created_by ON retail_contracts(created_by) WHERE created_by IS NOT NULL;

-- contract_price_daily.contract_id（已有复合索引含 contract_id，但单列查询也常见）
-- 无需单独建，idx_contract_price_daily_date 已含

-- intent_customer_monthly_retail_simulation.intent_id
CREATE INDEX IF NOT EXISTS idx_intent_retail_sim_intent ON intent_customer_monthly_retail_simulation(intent_id);

-- intent_customer_monthly_retail_simulation.package_id
CREATE INDEX IF NOT EXISTS idx_intent_retail_sim_pkg ON intent_customer_monthly_retail_simulation(package_id) WHERE package_id IS NOT NULL;

-- intent_customer_monthly_wholesale.intent_id
CREATE INDEX IF NOT EXISTS idx_intent_wholesale_intent ON intent_customer_monthly_wholesale(intent_id);

-- retail_settlement_daily.customer_id（已有复合 UNIQUE，但 FK JOIN 需索引）
CREATE INDEX IF NOT EXISTS idx_retail_settle_daily_customer ON retail_settlement_daily(customer_id);

-- retail_settlement_prices.customer_id
CREATE INDEX IF NOT EXISTS idx_retail_price_customer ON retail_settlement_prices(customer_id);

-- customer_characteristics.customer_id（UNIQUE 已含，无需额外索引）

-- customer_anomaly_alerts.acknowledged_by
CREATE INDEX IF NOT EXISTS idx_cust_alert_ack_by ON customer_anomaly_alerts(acknowledged_by) WHERE acknowledged_by IS NOT NULL;

-- customer_profit.customer_id（UNIQUE 已含，无需额外索引）

-- contract_progress.contract_id（UNIQUE 已含，无需额外索引）

-- load_characteristics.customer_id（UNIQUE 已含，无需额外索引）

-- customer_analysis.customer_id（UNIQUE 已含，无需额外索引）

-- green_power_trades.status（低基数码，偏索引）
CREATE INDEX IF NOT EXISTS idx_green_power_status ON green_power_trades(status) WHERE status != 'completed';

-- rolling_trades.status
CREATE INDEX IF NOT EXISTS idx_rolling_trades_status ON rolling_trades(status) WHERE status != 'completed';

-- bidding_records.status
CREATE INDEX IF NOT EXISTS idx_bidding_records_status ON bidding_records(status) WHERE status != 'completed';

-- task_commands.requested_by
CREATE INDEX IF NOT EXISTS idx_task_cmd_requested_by ON task_commands(requested_by) WHERE requested_by IS NOT NULL;

-- daily_release.released_by
CREATE INDEX IF NOT EXISTS idx_daily_release_released_by ON daily_release(released_by) WHERE released_by IS NOT NULL;

-- system_alerts.resolved_by
CREATE INDEX IF NOT EXISTS idx_system_alert_resolved_by ON system_alerts(resolved_by) WHERE resolved_by IS NOT NULL;

-- actual_operation.operator
CREATE INDEX IF NOT EXISTS idx_actual_op_operator ON actual_operation(operator) WHERE operator IS NOT NULL;

-- file_attachments.uploaded_by
CREATE INDEX IF NOT EXISTS idx_file_attachments_uploaded_by ON file_attachments(uploaded_by) WHERE uploaded_by IS NOT NULL;

-- agents.created_by
CREATE INDEX IF NOT EXISTS idx_agents_created_by ON agents(created_by) WHERE created_by IS NOT NULL;

-- bonds.created_by
CREATE INDEX IF NOT EXISTS idx_bonds_created_by ON bonds(created_by) WHERE created_by IS NOT NULL;

-- da_simulation_scenarios.created_by
CREATE INDEX IF NOT EXISTS idx_da_sim_created_by ON da_simulation_scenarios(created_by) WHERE created_by IS NOT NULL;

-- ─── created_at 索引补齐（时间范围查询热点）──────────
-- 这些表有 created_at 但缺少按时间倒序查询的索引

CREATE INDEX IF NOT EXISTS idx_auth_email_chl_created ON auth_email_challenges(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_auth_sec_chl_created ON auth_security_challenges(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_intent_customer_created ON intent_customers(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_task_exec_created ON task_execution_records(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_actual_op_created ON actual_operation(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_rolling_match_created ON rolling_match_snapshots(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_rpa_jobs_created ON rpa_jobs(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_rpa_runs_created ON rpa_runs(started_at DESC);
CREATE INDEX IF NOT EXISTS idx_contract_price_daily_created ON contract_price_daily(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_customer_profit_created ON customer_profit(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_monthly_trade_review_created ON monthly_trade_review(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_rolling_match_quotes_created ON rolling_match_quotes(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_monthly_manual_created ON monthly_manual_data(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_approval_requests_created ON approval_requests(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_file_attachments_created ON file_attachments(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_da_sim_scenarios_created ON da_simulation_scenarios(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_green_power_created ON green_power_trades(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_rolling_trades_created ON rolling_trades(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_bidding_records_created ON bidding_records(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_vpp_dispatches_created ON vpp_dispatches(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_storage_declaration_created ON storage_declaration(created_at DESC);

-- ─── 状态过滤索引补齐 ──────────────────────────────
-- 常见「WHERE status = X」过滤

CREATE INDEX IF NOT EXISTS idx_task_commands_status_created ON task_commands(status, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_task_exec_records_status_time ON task_execution_records(status, execution_time DESC);
CREATE INDEX IF NOT EXISTS idx_approval_status_created ON approval_requests(status, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_daily_release_status ON daily_release(status) WHERE status != 'completed';
CREATE INDEX IF NOT EXISTS idx_da_sim_scenarios_status ON da_simulation_scenarios(status) WHERE status != 'settled';

-- ─── 多租户查询优化 ────────────────────────────────
-- 按 org_id 过滤 + 时间排序（未来租户隔离查询热点）

CREATE INDEX IF NOT EXISTS idx_customers_org_created ON customers(org_id, created_at DESC) WHERE org_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_users_org_active ON users(org_id, is_active) WHERE org_id IS NOT NULL;

-- ─── 业务热点查询覆盖索引 ──────────────────────────
-- 审计日志：按路径前缀查（API 调用统计）
CREATE INDEX IF NOT EXISTS idx_audit_logs_path_prefix ON audit_logs(path, created_at DESC);

-- 天气预报：按地点 + 预报日期范围
CREATE INDEX IF NOT EXISTS idx_weather_forecast_loc_target ON weather_forecasts(location_name, target_date);

-- 天气实测：按地点 + 日期范围
CREATE INDEX IF NOT EXISTS idx_weather_actuals_loc_date ON weather_actuals(location_name, date);

-- 零售合同：有效合同查询（状态 + 日期范围）
CREATE INDEX IF NOT EXISTS idx_retail_contract_active_period
    ON retail_contracts(status, purchase_start_month, purchase_end_month)
    WHERE status = 'active';

-- 保函：按到期日排序（到期提醒查询）
CREATE INDEX IF NOT EXISTS idx_bonds_expire_status ON bonds(status, expire_date) WHERE status = 'active';
