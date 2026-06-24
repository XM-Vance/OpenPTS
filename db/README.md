# 数据库迁移

使用 [golang-migrate](https://github.com/golang-migrate/migrate) 管理。

## 文件命名规范

```
NNNN_name.up.sql       前进脚本
NNNN_name.down.sql     回滚脚本
```

- `NNNN`：4 位整数序号，从 0001 开始
- `name`：snake_case，描述本次变更

## 本地执行迁移

```bash
# 安装 golang-migrate CLI
brew install golang-migrate                                # macOS
# 或参考 https://github.com/golang-migrate/migrate/releases

# 前进到最新
migrate -path ./migrations \
        -database "postgres://ptis:***@localhost:5432/ptis?sslmode=disable" \
        up

# 回滚一步
migrate -path ./migrations \
        -database "postgres://ptis:***@localhost:5432/ptis?sslmode=disable" \
        down 1
```

## docker-compose 自动初始化

`docker-compose.yml` 通过一个独立的 `migrate` 服务（基于 golang-migrate）在 Postgres 就绪后
按文件名顺序执行 `db/migrations/` 下的迁移脚本（`*.up.sql` / `*.down.sql` 成对）。
该服务运行完毕即退出，后续可通过 `make migrate-up` / `make migrate-down` 手动增减版本。


## 阶段 1 规划

阶段 1 将基于 v1 MongoDB 集合反推完整 schema。预计涉及表（按业务域分组）：

| 业务域 | 主要表 |
|---|---|
| 用户权限 | users, roles, permissions, user_roles, role_permissions |
| 客户档案 | customers, customer_tags, customer_contacts |
| 负荷 | load_records, load_forecasts, load_diagnoses |
| 价格 | price_records, price_forecasts, grid_agency_prices |
| 零售 | retail_contracts, retail_packages, retail_settlements |
| 储能 | storage_stations, storage_declarations |
| 调频 | freq_regulation_records, freq_comp_fees |
| 调度 | scheduled_tasks, task_runs, rpa_monitors |
| 日志 | audit_logs, system_logs |
