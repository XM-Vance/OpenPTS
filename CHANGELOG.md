# 变更日志

本项目的关键变更记录。版本号遵循 [语义化版本](https://semver.org/lang/zh-CN/)。

## [0.2.1] — 2026-09-20

从生产分支增量同步：全站 UI 改造 + 4 项通用业务 + 安全加固。剥离算法真实化系列（WP2-WP5）、省份结算引擎（闽/赣/皖/苏）、钉钉、MCP。

### ✨ 新增

- **🎨 全站 UI 改造**：数据蓝品牌色体系、暗色模式全站适配（徽章/斑马纹对比度）、图表调色板（palette.ts）、骨架屏加载（Skeleton/ChartLoading）、空态组件化（EmptyState）、数值右对齐、ChartContainer 图表白屏修复。
- **计量数据底座**：`raw_meter_data` 表计数据导入通道（严格限流）+ 多表计聚合任务（`aggregate_meter_load` 定时任务）。
- **负荷校准台账**：表计 vs 系统口径系数（预览保存/列表/应用/作废 4 端点）。
- **合同文档生成恢复**：自包含 HTML 打印即 PDF（替代 gofpdf 归档依赖）。
- **意向客户转正式**：前端操作接线（前后端闭环）。
- **全局搜索**：顶栏搜索真正可用（客户/合同/文档三域）。

### 🔒 安全加固

- JWT_SECRET 占位值识别扩展 + 生产环境长度校验（≥32 字符，RFC 8729）。
- RPA 多租户隔离（rpa_jobs/rpa_runs 补 org_id + 唯一键，迁移 0120）。
- scheduler exec 加固 + 任务运行指标（`ptis_scheduler_job_runs_total`）。

### 🧹 清理

- 删除 10 个无引用的死 api 模块（auth-management/customer-analysis/retail-contracts 等）。
- OpenAPI 契约同步扩展并剥离 19 个引擎端点（batches/jx/ah/calculate 等仅存在于完整版）。

### 迁移（0120-0123）

| 号 | 内容 |
|---|---|
| 0120 | rpa_org（RPA 多租户隔离） |
| 0121-0122 | meter_import + meter_aggregate_job（计量底座） |
| 0123 | load_calibration（负荷校准台账） |

### ⚠️ 不含（开源边界）

- 算法真实化系列（LEAR 电价预测/负荷预测 ML/撮合引擎/策略回测/储能 LP/调频清算/VPP 聚合/碳持仓/绿电闭环）
- 省份结算引擎（江苏/赣/皖/闽）与结算批次状态机
- 钉钉（含审批告警双通道）、MCP 算法工具

### 📌 已知事项

- `github.com/xuri/excelize/v2@v2.11.0` 存在 [GO-2026-6452](https://pkg.go.dev/vuln/GO-2026-6452)（特制 xlsx 可触发 panic，DoS 级非 RCE，仅影响 Excel 上传解析场景）。上游暂无修复版，发布后请升级。CI 的 security job 为仅报告模式不阻断。

---

## [0.2.0] — 2026-07-19

从生产分支同步通用改进 + 新增 3 个零耦合模块，剥离省份结算引擎与算法服务。

### ✨ 新增

- **用户 API Key 体系**：用户可生成/吊销个人 API Key（`openpts_` 前缀 + bcrypt 哈希存储），凭 Key 换 JWT 后即可让外部脚本/工具以本人身份访问 API。前 4 个端点：`POST/GET/DELETE /api/v1/auth/api-keys` + `POST /api/v1/auth/api-key/exchange`。
- **OpenTelemetry 分布式追踪**：backend 启动时初始化 OTel TracerProvider（OTLP HTTP 导出），Gin 入口经 otelgin 自动埋点；连不上 Tempo 静默降级，业务零影响。监控栈（`docker-compose.monitoring.yml`）新增 Tempo 服务，Grafana 可查询瀑布图。
- **气象数据采集闭环**：补全 5 个迁移（weather_locations 种子、fetch_weather_actuals/fetch_weather_data 定时任务、md_weather_wind_hourly 字段对齐、md_hydrology 加 temp_max/min）+ scheduler 注册 `FetchWeatherData`/`FetchWeatherActuals` handler，打通 Open-Meteo → md_weather_* → weather_actuals 的完整 ETL。
- **CI 安全扫描**：新增 security job（govulncheck + npm audit，仅报告不阻断）。

### 🔒 安全加固

- **500 错误不再泄露内部信息**：market_data/integration 的 3 处 500 响应改用通用文案。
- **提权防护**：SetRoles / SetPermissions / SetUserOrgs 三个高危接口要求 super_admin 身份，且禁止自我赋权（改自己的角色/组织须另一名超管操作）。
- **DELETE 权限码统一**：trade-rules 等删除操作改用 `:delete` 权限码（独立于 `:write`）。
- **合同甲乙方识别方向修正**：document importers 优先用甲方（用电客户），乙方仅兜底。
- **金额校验工具**：新增 `util.go` 的 `requireNonNegFloat`/`requireNonNegDecimal` 校验。

### ♻️ 变更

- **删除保函管理（bonds）模块**：迁移 0118 DROP bonds 表 + 清理菜单种子；前端 bonds 页/api client/菜单项一并移除。
- **依赖升级**：backend minio-go 7.2.1 / excelize 2.11.0 / golang.org/x/crypto；frontend @types/node 26 / @vitejs/plugin-react 5 / lucide-react 1.x / eslint patch。
- **清理死代码**：删除 sqlc 脚手架（backend/sqlc.yaml）。

### 🩹 修复

- dashboard 500（现货 items:null + DELETE 权限码）
- 气象两页重做（去假数据，改风电出力视角）
- market-data 分页（LIMIT 5000 → page/page_size + total + has_more）
- document apply 0 行返回 422 + 拒绝老式 .xls

### ⚠️ 不含（开源边界，未同步）

- 钉钉集成（不上传）
- 售电知识库 + MCP server（依赖算法服务，暂不引入）
- 文档转换引擎 / 省份结算引擎 / 负荷预测持久化 / 撮合引擎（算法或省份特定）
- 删除 docling-service（保留 docling 保证文档解析闭环可用）

---

## [0.1.0] — 2026-06-24

首个开源发布：从生产系统剥离算法与省份规则后的通用运营骨架。

### ✨ 核心内容

- **完整运营框架**：Go 网关（300+ API、JWT 鉴权、RBAC、多租户、审计）+ Next.js 15 前端（30+ 业务页面）+ PostgreSQL（83 对迁移、~190 张表）+ docling 文档解析。
- **业务域**：客户、零售合同、结算、价格、负荷、绿电、调频、储能、文档解析等 ~30 个域。
- **算法接入点**：负荷/价格预测端点返回 501 + 详细接入指引，由二次开发者接入自有算法。
- **一键 Demo**：`bash scripts/demo.sh` 两分钟体验完整系统（含合成数据 + 合成预测）。
- **开源配套**：CI（Go + Next.js）、Issue/PR 模板、MIT 协议、效果截图、《为什么做 OpenPTS》文章、Roadmap、5 个 Good First Issue。

---

[0.2.0]: https://github.com/XM-Vance/OpenPTS/releases/tag/v0.2.0
[0.1.0]: https://github.com/XM-Vance/OpenPTS/releases/tag/v0.1.0
[0.2.1]: https://github.com/XM-Vance/OpenPTS/releases/tag/v0.2.1
