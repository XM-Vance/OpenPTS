-- 139 down: 移除表计负荷聚合作业定义（已聚合的数据保留）。
DELETE FROM scheduled_jobs WHERE name = 'aggregate_meter_load';
