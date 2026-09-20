-- 139: 表计负荷聚合作业（WP0.3）。每日 23:40 把 raw_meter_data 近 2 天
-- 按 (customer, date) 汇聚（Σ 表计曲线×倍率）写入 user_load_data + unified_load_curve。
INSERT INTO scheduled_jobs (name, description, cron_expr, handler) VALUES
    ('aggregate_meter_load', '表计负荷聚合（raw_meter_data → user_load_data + unified_load_curve，多表计按倍率求和）', '0 40 23 * * *', 'aggregate_meter_load')
ON CONFLICT (name) DO NOTHING;
