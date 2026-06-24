-- 0046: 回滚性能索引
-- 按创建顺序删除所有 0046 新增的索引（IF EXISTS 确保幂等）。

-- ─── 外键列索引 ─────────────────────────────────────
DROP INDEX IF EXISTS idx_customers_created_by;
DROP INDEX IF EXISTS idx_intent_customer_converted;
DROP INDEX IF EXISTS idx_retail_pkg_created_by;
DROP INDEX IF EXISTS idx_retail_contract_created_by;
DROP INDEX IF EXISTS idx_intent_retail_sim_intent;
DROP INDEX IF EXISTS idx_intent_retail_sim_pkg;
DROP INDEX IF EXISTS idx_intent_wholesale_intent;
DROP INDEX IF EXISTS idx_retail_settle_daily_customer;
DROP INDEX IF EXISTS idx_retail_price_customer;
DROP INDEX IF EXISTS idx_cust_alert_ack_by;
DROP INDEX IF EXISTS idx_green_power_status;
DROP INDEX IF EXISTS idx_rolling_trades_status;
DROP INDEX IF EXISTS idx_bidding_records_status;
DROP INDEX IF EXISTS idx_task_cmd_requested_by;
DROP INDEX IF EXISTS idx_daily_release_released_by;
DROP INDEX IF EXISTS idx_system_alert_resolved_by;
DROP INDEX IF EXISTS idx_actual_op_operator;
DROP INDEX IF EXISTS idx_file_attachments_uploaded_by;
DROP INDEX IF EXISTS idx_agents_created_by;
DROP INDEX IF EXISTS idx_bonds_created_by;
DROP INDEX IF EXISTS idx_da_sim_created_by;

-- ─── created_at 索引 ────────────────────────────────
DROP INDEX IF EXISTS idx_auth_email_chl_created;
DROP INDEX IF EXISTS idx_auth_sec_chl_created;
DROP INDEX IF EXISTS idx_intent_customer_created;
DROP INDEX IF EXISTS idx_task_exec_created;
DROP INDEX IF EXISTS idx_actual_op_created;
DROP INDEX IF EXISTS idx_rolling_match_created;
DROP INDEX IF EXISTS idx_rpa_jobs_created;
DROP INDEX IF EXISTS idx_rpa_runs_created;
DROP INDEX IF EXISTS idx_contract_price_daily_created;
DROP INDEX IF EXISTS idx_customer_profit_created;
DROP INDEX IF EXISTS idx_monthly_trade_review_created;
DROP INDEX IF EXISTS idx_rolling_match_quotes_created;
DROP INDEX IF EXISTS idx_monthly_manual_created;
DROP INDEX IF EXISTS idx_approval_requests_created;
DROP INDEX IF EXISTS idx_file_attachments_created;
DROP INDEX IF EXISTS idx_da_sim_scenarios_created;
DROP INDEX IF EXISTS idx_green_power_created;
DROP INDEX IF EXISTS idx_rolling_trades_created;
DROP INDEX IF EXISTS idx_bidding_records_created;
DROP INDEX IF EXISTS idx_vpp_dispatches_created;
DROP INDEX IF EXISTS idx_storage_declaration_created;

-- ─── 状态过滤索引 ───────────────────────────────────
DROP INDEX IF EXISTS idx_task_commands_status_created;
DROP INDEX IF EXISTS idx_task_exec_records_status_time;
DROP INDEX IF EXISTS idx_approval_status_created;
DROP INDEX IF EXISTS idx_daily_release_status;
DROP INDEX IF EXISTS idx_da_sim_scenarios_status;

-- ─── 多租户查询索引 ─────────────────────────────────
DROP INDEX IF EXISTS idx_customers_org_created;
DROP INDEX IF EXISTS idx_users_org_active;

-- ─── 业务热点覆盖索引 ──────────────────────────────
DROP INDEX IF EXISTS idx_audit_logs_path_prefix;
DROP INDEX IF EXISTS idx_weather_forecast_loc_target;
DROP INDEX IF EXISTS idx_weather_actuals_loc_date;
DROP INDEX IF EXISTS idx_retail_contract_active_period;
DROP INDEX IF EXISTS idx_bonds_expire_status;
