# 架构设计

> **关于算法服务**：本文档描述的是完整平台架构。本开源骨架已移除独立的 Python
> 算法服务（`algo-service/`）及其对应的负荷/价格预测、竞价策略等算法实现，
> 仅保留 Go 网关、Next.js 前端、数据库与文档解析服务。
> 下文出现的"算法服务"相关描述说明的是**算法接入点**：相关预测端点
> （如 `POST /api/v1/load/forecast`、`POST /api/v1/price/forecast`）在骨架中
> 返回 `501 Not Implemented`，二次开发者可在这些端点接入自有算法。详见各 handler
> 中的接入指引注释。

## 一、整体拓扑

```
┌──────────┐     ┌───────────────┐     ┌──────────────┐
│ 浏览器   │ ──▶ │ Next.js 前端  │ ──▶ │ Go 网关 (Gin) │
└──────────┘     │ (App Router)  │     │              │
                 └───────────────┘     │   鉴权 / RBAC │
                                        │   CRUD       │
                                        │   编排       │
                                        └──────┬───────┘
                                               │
                              ┌────────────────┼────────────────┐
                              ▼                                 ▼
                       ┌──────────────┐                ┌──────────────────┐
                       │ PostgreSQL16 │                │ Python 算法服务  │
                       │              │                │ (FastAPI)        │
                       │ 业务数据     │                │ - 负荷预测       │
                       │ 用户权限     │                │ - 价格预测       │
                       │ 任务调度     │                │ - DTW / 统计     │
                       └──────────────┘                │ - 客户特征分析   │
                                                       └──────────────────┘
```

## 二、技术选型理由

### 为什么 Go 而非 Node/Rust？
- **vs Node**：单二进制部署、并发模型简洁、内存可控
- **vs Rust**：Web 后端场景 Rust 的性能优势不显著，但开发成本明显更高
- **生态**：Gin / sqlc / golang-migrate 在国内外都成熟

### 为什么保留 Python 做算法服务？
v1 的负荷预测、价格预测、客户特征分析重度依赖 pandas / scikit-learn / statsmodels / scipy / fastdtw。
强制重写为 Go 会导致：
1. 大量算法需要重新实现（统计模型、DTW 算法、时间序列）
2. 数据科学迭代变慢（Python 调试 vs Go 重新编译）
3. 与领域知识脱节（电力交易领域 Python 模板与论文居多）

折中：Go 处理 HTTP、鉴权、CRUD、编排；Python 仅提供算法服务，通过 HTTP/JSON 调用。

### 为什么 Next.js + shadcn 而非保留 MUI？
- shadcn 不是组件库，是"组件方案"，组件代码在自己仓库，可控性最高
- Tailwind 比 MUI 的 sx 系统更直观，且无运行时开销
- App Router 的 RSC 模式适合数据密集型业务页面（直接在服务端拉数据）

### 为什么 PostgreSQL 而非保留 MongoDB？
- 电力结算、合同、套餐这类业务本质是关系型
- v1 在 MongoDB 上做事务、join 一直比较别扭
- Postgres 的 JSONB 字段足够覆盖半结构化场景
- 与 sqlc 配合可获得编译期 SQL 类型安全

## 三、模块边界（职责清单）

### 前端（frontend/）
- UI 渲染、路由、表单校验、图表
- 状态管理：@tanstack/react-query（服务端状态） + React 自带（局部状态）
- 鉴权：JWT 存 httpOnly Cookie；权限 Hook 控制按钮/菜单可见性
- 不直接调用算法服务，统一经 Go 网关

### Go 网关（backend/）
- HTTP API（RESTful，少量场景考虑 GraphQL 或聚合接口）
- JWT 鉴权 + RBAC 权限校验中间件
- 直接读写 Postgres（sqlc 生成类型安全代码）
- 编排算法服务调用（带超时、重试、熔断）
- 任务调度（定时任务、RPA 触发）

### Python 算法服务（algo-service/）
- 仅提供算法计算接口，**不直接读数据库**（数据由 Go 网关传入）
- 输入：业务参数 + 数据数组（JSON）
- 输出：计算结果（JSON）
- 无状态、可水平扩展

### PostgreSQL
- 单一可信数据源
- 所有持久化通过 Go 网关写入
- 算法服务的中间结果若需持久化，由 Go 网关回写

## 四、跨服务通信

| 通道 | 协议 | 备注 |
|---|---|---|
| 前端 → Go 网关 | HTTPS / REST + JSON | JWT bearer token |
| Go 网关 → Python 算法 | HTTP / JSON | 内网，超时 30s，自动重试 1 次 |
| Go 网关 → Postgres | TCP / pgx | 连接池，最大 20 |

## 五、关键依赖

### Backend (Go)
- `gin-gonic/gin` —— HTTP 框架
- `jackc/pgx/v5` —— Postgres 驱动
- `sqlc-dev/sqlc` —— SQL → Go 代码生成
- `golang-migrate/migrate` —— 数据库迁移
- `golang-jwt/jwt/v5` —— JWT
- `rs/zerolog` —— 结构化日志
- `go-playground/validator` —— 参数校验
- `robfig/cron/v3` —— 定时任务

### Algo Service (Python)
- `fastapi` + `uvicorn[standard]`
- `pandas` / `numpy` / `scipy`
- `scikit-learn` / `statsmodels`
- `fastdtw`
- `httpx`（如需调用外部）

### Frontend (Next.js)
- `next` / `react` (19)
- `@tanstack/react-query` —— 服务端状态
- `react-hook-form` + `zod` —— 表单 + 校验
- `recharts` —— 图表（v1 同款，迁移成本低）
- `shadcn/ui` 组件 + `tailwindcss`
- `lucide-react` —— 图标
- `date-fns` —— 时间

## 六、阶段计划

### 阶段 0：基础设施骨架 ✅
本阶段交付物。

### 阶段 1：横切关注点
- 数据库 schema 设计（基于 v1 MongoDB 集合反推 27 个领域模型）
- 鉴权体系：JWT + RBAC（沿用 v1 `docs/spec/权限与鉴权规则.md` 的粒度）
- Go 中间件：日志、错误、限流、CORS、超时
- Go ↔ Python 内部协议
- 前端：布局、主题、路由、API 客户端、权限 Hook、基础组件
- shadcn 关键组件按需 add

### 阶段 2：27 个业务模块迁移
按价值排序，每个模块独立验收。
推荐顺序：
1. 用户权限管理（先有它才能测其他）
2. 客户档案管理
3. 短期负荷预测 / 预测基础数据
4. 价格预测 / 现货价格分析
5. 零售合同 / 套餐 / 结算
6. 储能 / 调频
7. 其他

### 阶段 3：调度、RPA 监控、日志告警
### 阶段 4：性能压测、监控、部署

## 七、与 v1 的对照

| 项 | v1 | v2 |
|---|---|---|
| 前端框架 | React 19 + MUI 7 | Next.js 15 + shadcn/Tailwind |
| 前端路由 | react-router 7 | Next App Router |
| 后端框架 | FastAPI | Gin（网关） + FastAPI（算法） |
| 数据库 | MongoDB | PostgreSQL 16 |
| 表单 | react-hook-form + zod | 同 v1（保留） |
| 图表 | recharts | 同 v1（保留） |
| 鉴权 | JWT + 自研权限 | JWT + RBAC（重新设计） |
| 调度 | APScheduler | robfig/cron + Postgres |
| 时间字段 | naive datetime | Postgres `timestamptz` |

## 八、实际部署拓扑（v2 已就绪）

```
                              ┌─────────────────┐
                              │  浏览器 / SPA   │
                              └────────┬────────┘
                                       │ HTTPS 443
                              ┌────────▼────────┐
                              │  Nginx / Caddy  │     ← Caddy 自动 TLS 备选
                              └────────┬────────┘
            ┌────────────────────┬─────┴────────────────────┐
            │ /api/*  /health    │                          │ /*
            │ /metrics  /docs    │                          │
   ┌────────▼────────┐  ┌────────▼────────┐        ┌────────▼─────────┐
   │  backend (Go)   │  │ algo-service    │        │  frontend (Next) │
   │  - Gin + pgx    │  │  - FastAPI      │        │  - SSR + standalone
   │  - JWT + RBAC   │  │  - sklearn/pd   │        └──────────────────┘
   │  - cron 调度    │  │  - 负荷/价格预测
   │  - 审计中间件   │  └────────┬────────┘
   └────────┬────────┘           │
            │ pgx 连接池          │ HTTP (algo.Client)
   ┌────────▼─────────────────────┘
   │  PostgreSQL 16
   │  - 50+ 业务表（含 26 迁移）
   │  - audit_logs / scheduled_jobs / job_runs
   └──────────────────
   监控旁路：
   backend /metrics ──► Prometheus ──► Alertmanager ──► alert-webhook-proxy ──► 飞书/邮件
                            ↑
                    node-exporter / cAdvisor / postgres-exporter
                            ↓
                          Grafana (预制 dashboard)
```

## 九、ADR（架构决策记录）

### ADR-001：选 Go + Gin 而非 NodeJS / Java 作网关
- **背景**：v1 用 FastAPI 直接连 MongoDB，承担网关与算法两职。
- **决策**：拆分。**网关用 Go**，算法保留 Python。
- **理由**：
  - Go 静态二进制 + distroless 镜像 < 30 MB，启动 < 100 ms；
  - 强类型 + race detector + 内建 cron 工具链；
  - 不需要 JIT 预热，适合容器化弹性扩缩；
  - 算法生态（pandas/sklearn）仍由 Python 承担，互不打架。
- **代价**：团队需要熟悉 Go；与 Python 双语言协作通过 HTTP（algo.Client）。

### ADR-002：Postgres 替代 MongoDB
- **背景**：v1 用 MongoDB，存在跨集合查询无法 JOIN 的痛点。
- **决策**：全量切换 PostgreSQL 16。
- **理由**：
  - JSON 字段保留（`jsonb`），保留半结构化弹性；
  - 跨表 JOIN / CTE / FILTER 一句 SQL 完成（典型：dashboard summary 9 个 KPI 单查询）；
  - 96 时段负荷曲线用 `double precision[]`，pgx 类型映射稳定；
  - golang-migrate 双向迁移文件（up/down）支持回滚。
- **代价**：数据不迁移（v2 从零部署），v1 仍冻结可查。

### ADR-003：前端用 Next.js 15 App Router 而非 SPA
- **背景**：v1 React + react-router 全部 CSR，首屏白屏感强。
- **决策**：迁移到 Next 15 App Router + Server Components 默认。
- **理由**：
  - 文件即路由（如 `app/(main)/dashboard/page.tsx`），结构清晰；
  - 路由组 `(main)` 套 AppShell，与登录页解耦；
  - standalone 构建产出 200 MB 容器，含最小 node_modules；
  - SSR 可逐步开启（目前 dashboard 等仍是 `'use client'`，但路由保留升级潜力）。
- **代价**：useSearchParams / next/navigation 需要 use client，初期心智成本。

### ADR-004：RBAC 用「权限码字符串」而非数字位运算
- **决策**：权限统一形如 `module:action`（如 `customer_management:write`）。
- **理由**：
  - 可读性强，handler 上下文一眼看明白；
  - 数据库表 `auth_permissions(code)` 主键就是字符串；
  - Postgres 索引足够快，5 分钟内存缓存（`PermissionService`）兜底。
- **代价**：理论上不如位运算紧凑，但实测无瓶颈。

### ADR-005：业务监控走 Prometheus（而非自研日志聚合）
- **决策**：backend 暴露 `/metrics`，全栈监控用 Prom + Grafana。
- **理由**：
  - 直方图 + label 维度即开即用，无需自研；
  - 11 条告警规则 → Alertmanager → 飞书机器人闭环；
  - 与 cAdvisor / node-exporter / postgres-exporter 同一栈，开销极低。
- **代价**：需要熟悉 PromQL；告警卡片需要自适配（已通过 alert-webhook-proxy 解决）。

### ADR-006：调度任务进程内（robfig/cron）而非 Celery / Sidekiq
- **决策**：用 `github.com/robfig/cron/v3` 在 backend 进程内调度。
- **理由**：
  - 当前规模仅 3-10 个任务，无需独立 worker 集群；
  - 状态持久化进 `scheduled_jobs` + `job_runs` 表，重启不丢；
  - 同进程方便调用 repo 与 pool，避免引入消息队列。
- **代价**：水平扩多实例时需要加 leader election 或单实例运行；当前接受。

### ADR-007：API 文档手写 OpenAPI 而非 swag 生成
- **决策**：在 `backend/internal/handler/openapi.yaml` 手写规范，go:embed 进二进制。
- **理由**：
  - swag 注解会污染 handler；
  - 手写一份 yaml 比扫注解可控，且 IDE 中可校验；
  - Swagger UI 通过 CDN 加载，无需第三方 Go 库。
- **代价**：增删接口时需手工同步 yaml（CI 可加 lint 校验，下一步）。

## 十、迁移历史

| 迁移 | 主题 |
|---|---|
| 0001-0014 | 初始 schema + 鉴权 + 客户/零售/负荷/价格/结算/调频/储能/分析 14 张核心表 |
| 0015 | 14 模块 / 42 权限 / 4 角色种子 |
| 0016 | user_load_data.curve_96 类型改为 `double precision[]`（pgx 兼容） |
| 0017 | storage_stations + storage_daily_operation |
| 0018 | scheduled_jobs + job_runs（B 批） |
| 0019 | audit_logs（C 批） |
| 0020-0024 | 月度结算 / 日前复盘 / 气象 / RPA / 合同电价（D 批） |
| 0025 | grid_agency_price + storage_declaration（E 批） |
| 0026 | customer_profit + monthly_trade_review + rolling_match_quotes + monthly_manual_data（F 批） |

至 0026 共 **26 个迁移文件**，所有 down 文件均按反向顺序 `DROP IF EXISTS`，可一键回滚。

## 十一、测试金字塔

| 层 | 工具 | 位置 | 数量 |
|---|---|---|---|
| 单测（Go） | `go test` | `backend/internal/auth/*_test.go`、`backend/internal/middleware/audit_test.go` | 11 用例 |
| 单测（前端） | vitest + Testing Library | `frontend/tests/*.test.ts` | 9 用例 |
| 端到端 | Playwright | `e2e/specs/*.spec.ts` | 8 用例（鉴权 3 + 仪表盘 4 + 客户 CRUD 1） |
| 性能 | k6 | `perf/smoke.js`、`perf/load.js` | 冒烟 5 接口 + 阶梯压测 |
| CI 集成 | GitHub Actions | `.github/workflows/ci.yml` | 6 job 并行/串行（backend / frontend / algo / docker / migrations / e2e） |

