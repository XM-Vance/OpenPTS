# audit_logs 按月分区方案

## 背景

`audit_logs` 是写多读少的高增长表：每个 POST/PUT/DELETE 请求都会落一条，按 100 req/s 估算每天 ~430 万行、每月 ~1.3 亿行。
现状用 B-Tree 索引 + DELETE 清理，3-6 个月后 vacuum 不及时会引起索引膨胀与查询变慢。

阶段 R3 先上**索引优化**（0030 迁移），后续切换到**声明式分区**。

## 何时切换

满足以下任一条件时立即切：
- 单表 ≥ 5,000 万行
- 安全大屏 P95 > 1s
- `vacuum_analyze` 时间 > 30 分钟
- 准备实施数据保留策略（自动删除 6 个月前数据）

## 迁移步骤（非破坏式，零停机）

```sql
-- 1) 重命名旧表
ALTER TABLE audit_logs RENAME TO audit_logs_legacy;

-- 2) 建分区主表（结构一致，按 created_at 分区）
CREATE TABLE audit_logs (
    id            uuid NOT NULL DEFAULT gen_random_uuid(),
    user_id       uuid,
    username      text,
    method        text NOT NULL,
    path          text NOT NULL,
    resource      text,
    resource_id   text,
    status_code   integer NOT NULL,
    ip            inet,
    user_agent    text,
    duration_ms   integer NOT NULL DEFAULT 0,
    error_message text,
    created_at    timestamp NOT NULL DEFAULT now(),
    PRIMARY KEY (id, created_at)
) PARTITION BY RANGE (created_at);

-- 3) 建索引（自动传播到分区）
CREATE INDEX ON audit_logs(created_at DESC);
CREATE INDEX ON audit_logs(user_id, created_at DESC);
CREATE INDEX ON audit_logs(resource, created_at DESC) WHERE resource IS NOT NULL;

-- 4) 建初始分区（覆盖过去 3 个月 + 未来 1 个月）
CREATE TABLE audit_logs_y2026m01 PARTITION OF audit_logs
    FOR VALUES FROM ('2026-01-01') TO ('2026-02-01');
CREATE TABLE audit_logs_y2026m02 PARTITION OF audit_logs
    FOR VALUES FROM ('2026-02-01') TO ('2026-03-01');
-- ... 按月类推

-- 5) 默认分区兜底（防意外越界）
CREATE TABLE audit_logs_default PARTITION OF audit_logs DEFAULT;

-- 6) 把历史数据搬过去（小批量分批，避免锁表）
INSERT INTO audit_logs
SELECT * FROM audit_logs_legacy
WHERE created_at >= '2026-01-01' AND created_at < '2026-02-01';

-- 校验后删除 legacy
DROP TABLE audit_logs_legacy;
```

## 自动化滚动

需要一个定时任务，每月初创建下一个月的分区。可加到 `scheduled_jobs`：

```sql
-- 调度任务名：rotate_audit_partitions，cron：0 0 0 1 * *（每月 1 号 00:00）
-- handler 内逻辑：
--   1) 计算下个月的起止日期
--   2) IF NOT EXISTS 创建分区 audit_logs_yYYYYmMM
--   3) 删除 6 个月前的旧分区（DROP TABLE，比 DELETE 快几个数量级）
```

handler Go 代码骨架：

```go
func RotateAuditPartitions(ctx context.Context, pool *db.Pool) error {
    now := time.Now()
    next := time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, time.Local)
    afterNext := next.AddDate(0, 1, 0)
    table := fmt.Sprintf("audit_logs_y%dm%02d", next.Year(), next.Month())

    _, err := pool.Exec(ctx, fmt.Sprintf(
        `CREATE TABLE IF NOT EXISTS %s PARTITION OF audit_logs
         FOR VALUES FROM ('%s') TO ('%s')`,
        table, next.Format("2006-01-02"), afterNext.Format("2006-01-02"),
    ))
    if err != nil {
        return err
    }

    // 删除 6 个月前的旧分区
    drop := now.AddDate(0, -6, 0)
    dropTable := fmt.Sprintf("audit_logs_y%dm%02d", drop.Year(), drop.Month())
    _, _ = pool.Exec(ctx, "DROP TABLE IF EXISTS "+dropTable)
    return nil
}
```

## 性能预期

| 维度 | 单表 | 月分区 |
|---|---|---|
| 5000 万行后 SELECT P95 | 300-500 ms | 30-50 ms（命中单分区） |
| 删除 6 个月前数据 | DELETE 数小时 + VACUUM 数小时 | `DROP TABLE` < 1 秒 |
| 索引大小 | ~20 GB | 单分区 ~1-2 GB |

## 风险与回退

- **风险**：客户端如果用 ID 查询且不带 created_at，需要扫所有分区
  - **缓解**：handler 已统一加 ORDER BY created_at DESC LIMIT；查单 ID 场景极少
- **回退**：分区表语义和普通表一致，必要时 `ALTER TABLE audit_logs DETACH PARTITION` 单独导出，再合并
