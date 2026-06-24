# ptis-backend (Go 网关)

电力交易信息系统 v2 的 HTTP 网关。

## 职责
- 接收前端 HTTP 请求
- JWT 鉴权 + RBAC 权限校验
- 直接读写 PostgreSQL（sqlc 生成代码）
- 编排 Python 算法服务调用
- 定时任务调度

## 目录
```
backend/
├── cmd/server/             启动入口
├── internal/
│   ├── server/             路由组装
│   ├── handler/            HTTP handler
│   ├── middleware/         中间件（阶段 1 填充）
│   ├── config/             配置加载（阶段 1 填充）
│   └── db/                 sqlc 生成代码 + 仓储层（阶段 1 填充）
├── go.mod
├── Dockerfile
└── Makefile
```

## 本地启动
```bash
# 1. 安装依赖
go mod tidy

# 2. 启动 Postgres
cd .. && docker compose up -d postgres

# 3. 运行
cd backend
make run
```

健康检查：
```bash
curl http://localhost:8080/health
curl http://localhost:8080/api/v1/ping
```

## 阶段计划
- 阶段 0：最小可运行骨架（**当前**）
- 阶段 1：补充中间件、sqlc、鉴权、配置
- 阶段 2：按业务模块新增 handler / repository
