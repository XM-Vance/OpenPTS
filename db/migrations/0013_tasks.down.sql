DROP TRIGGER IF EXISTS daily_release_set_updated_at ON daily_release;
DROP TABLE IF EXISTS daily_release;

-- 分区表 DROP 会一并清理所有子分区
DROP TABLE IF EXISTS task_execution_logs;

DROP TABLE IF EXISTS task_execution_records;
DROP TABLE IF EXISTS task_commands;
