# OpenPTS · 开放式电力交易系统

> **OpenPTS** = **Open** **P**ower **T**rading **S**ystem —— 一个可二次开发的电力交易平台开源骨架。

> ⚠️ **WIP — 尚未公开发布**。本项目为本地开发分支，尚未发布到公开仓库。

一个可二次开发的**开放式电力交易系统脚手架**。基于成熟的生产架构（Go 网关 + Next.js + PostgreSQL），
覆盖电力零售交易运营的核心业务域（客户、合同、结算、价格、负荷、文档解析等），
**算法与预测能力留作接入点**，由二次开发者按自有策略实现。

## 为什么是"骨架"？

完整的电力交易平台 = **通用运营框架** + **算法/规则内核**。
本项目开放的是**通用运营框架**部分：

| 组成 | 本项目是否提供 |
|---|---|
| 鉴权 / RBAC / 多租户 / 审计 | ✅ 完整 |
| 客户、零售合同、结算、价格、负荷等业务 CRUD | ✅ 完整 |
| 文档解析（电费单/合同 OCR → 结构化） | ✅ 完整（docling 服务） |
| 负荷预测 / 价格预测 | ⚠️ 仅留接口（返回 501），**算法由你接入** |
| 竞价策略 / 市场出清 / 潮流计算 | ❌ 未包含（可自行扩展） |
| 具体省份交易规则数据 | ❌ 不含（规则表结构通用，数据由你填充） |

这样设计的目的：让同业开发者能在一个**能跑通的框架**上协作，而把各自的核心算法/规则留在自己手里。

## 技术栈

| 层 | 选型 |
|---|---|
| 后端网关 | Go + Gin + sqlc + JWT(v5) |
| 前端 | Next.js 15 (App Router) + shadcn/ui + Tailwind + TypeScript |
| 数据库 | PostgreSQL 16 |
| 对象存储 | MinIO（S3 兼容） |
| 文档解析 | docling-service（PyMuPDF + GLM 视觉 OCR → 结构化） |
| 容器编排 | Docker Compose |
| 反向代理 | Nginx / 可选 Caddy (HTTPS) |

## 业务模块

已实现的核心域（均含完整 CRUD + 权限控制）：

- **客户管理** — 客户档案、意向客户、客户分析、客户电量、代理商、保函
- **零售管理** — 零售合同、合同进度、绿电交易
- **价格管理** — 现货价格、TOU 时段、代理购电价
- **负荷管理** — 系统总负荷、气象数据
- **结算管理** — 日/月结算、预结算、偏差管理、手工数据
- **运营管理** — 交易规则（通用规则表）、光伏监控/结算、市场行情
- **文档中心** — 文档解析（OCR→结构化→入库）、政策文件
- **系统支撑** — 用户/角色/组织/权限/菜单/审计/调度任务/RPA/大屏

## 目录结构

```
├── backend/           Go 网关（HTTP API、鉴权、CRUD、多租户）
├── docling-service/   文档解析微服务（GLM 视觉 OCR → 结构化）
├── frontend/          Next.js 前端
├── db/migrations/     数据库迁移脚本（golang-migrate）
├── scripts/           工具脚本（迁移、备份、契约校验等）
├── docs/              架构、API 契约、审计分区等文档
├── monitoring/        Prometheus + Grafana 配置
├── perf/              k6 性能压测脚本
├── e2e/               Playwright 端到端测试
├── nginx/             Nginx 配置
├── docker-compose.yml 生产编排
└── Caddyfile          可选 HTTPS（Caddy）
```

## 快速开始

### 前置要求
- Docker + Docker Compose
- （本地开发）Go 1.25+、Node.js 20+

### 一键启动（Docker）

```bash
# 1. 准备环境变量
cp .env.example .env
#    修改 .env 中的 JWT_SECRET / POSTGRES_PASSWORD / MINIO_SECRET_KEY
#    （至少 32 字符的随机串；可用 `openssl rand -base64 48`）
#    如需文档解析：填 ZHIPU_API_KEY（智谱开放平台）；仅做业务 CRUD 可不填。

# 2. 构建并启动
docker compose up -d --build

# 3. 执行数据库迁移
docker compose --profile migrate run --rm migrate up

# 4. 写入种子数据（随机生成 admin 密码并打印到日志）
docker compose exec backend /app/seed

# 5. 访问 http://localhost
#    用上一步打印的 admin 账号 + 随机密码登录
```

### 本地开发（不走 Docker）

```bash
# 后端
cd backend
go build ./...      # 编译
go test ./...       # 测试
go run ./cmd/server # 启动（需配 DATABASE_URL / JWT_SECRET 等环境变量）

# 前端
cd frontend
npm install
npm run dev         # 开发模式 http://localhost:3000
```

## 算法接入指引（二次开发）

骨架在以下位置预留了算法接入点，接入自有算法即可启用：

| 能力 | 接入位置 | 说明 |
|---|---|---|
| 负荷预测 | `backend/internal/handler/load.go` 的 `Forecast` 方法 | 取历史 96 点曲线（`repo.GetRecentCurves`）→ 调你的预测服务 → 返回 `algoForecastResponse` |
| 价格预测 | `backend/internal/handler/price.go` 的 `Forecast` 方法 | 取历史 48 点价格（`repo.GetRecentDayAheadCurves`）→ 调你的预测服务 |
| 交易规则 | `trade_rules` 表（通用 key/value 结构） | 填入你所在省份/市场的规则参数，前端 `/trade-rules` 页面可 CRUD |

每个接入点的文件顶部都有详细的注释指引（搜索 `// 接入指引`）。
请求/响应 DTO（`algoForecastRequest` / `algoForecastResponse`）已在 `load.go` 定义好，可直接复用。

## 协作与提交

详见 [CONTRIBUTING.md](CONTRIBUTING.md)。克隆后建议先 `make hooks` 启用推送前自检。

## License

待定（发布前确定）。
