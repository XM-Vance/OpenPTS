# 前端 API 契约 backlog（P1-7）

本文件登记前端 `lib/api/*.ts` 中 `TODO(P1): 后端无对应路由` 标记的端点——
这些函数当前**回退到已有的相近路由**（功能可用，但语义不精确），或调用尚不存在的路由（运行时 404）。

> **不是死代码**：契约门禁（`scripts/check_contract.sh`）扫描的是已注册路由，
> 这些函数回退到相近路由，故门禁不报错。本清单用于追踪「应当实现的专用路由」。

## 已补齐（P1-7 首批）

- [x] `GET /api/v1/retail/contracts/:id` — 单个合同详情（2026-06 补齐 handler + 路由）
  - `lib/api/retail-contracts.ts` `getContract` / `getRetailContractById`

## 待补齐 backlog（按域）

### 负荷特性 `load-characteristics.ts`（12 处）
GET `/load/characteristics/overview`、`/overview/distribution`、`/overview/scatter-data`、`/overview/tag-changes`、`/customers`、`/customer/:id`、`/customer/:id/history`、`/customer/:id/alerts`、`/customer/:id/daily-trend`、`/customer/:id/monthly-energy`、POST `/analyze/batch/all`、`/alerts/:id/acknowledge`

### 负荷数据 `load-data.ts`（11 处）
GET `/load-data/customers`、`/export/mp-missing`、`/signed-customers`、`/customers/:id`、`/customers/:id/calendar`、`/customers/:id/curves`、POST `/load-data/reaggregate`、`/calibration/preview`、`/calibration/calculate`、`/calibration/apply`、`/calibration/details`

### 零售月结算 `retail-settlement.ts`（9 处）
GET `/retail/monthly-settlement/monthly-customers`、`/monthly-chart-data`、`/monthly-progress/:jobId`、`/daily`、`/monthly-customer-detail`、POST `/calculate`、`/monthly-calc`

### 负荷预测 `load-forecast.ts`（5 处）
GET `/load/forecast/versions`、`/data`、`/customers`、`/performance-overview`、`/accuracy`（后者后端有 `/forecast/accuracy`，路径不一致）

### 零售合同 `retail-contracts.ts`（3 处，PDF 相关）
GET `/retail/contracts/:id/pdf`、POST `/retail/contracts/:id/upload-pdf`、GET `/retail/contracts/:id/has-pdf`

### 价格趋势 `contract-price-trend.ts`（3 处）
GET `/price/trend/price-trend`、`/curve-analysis`、`/quantity-structure`

### 现货市场 `spot-market.ts`（2 处）
GET `/price/spot-market/statistics`、`/price-curve`

### 合同日均价 `contract-price.ts`（1 处）
GET `/retail/price-daily/daily-summary`

### 竞价策略 `bidding-strategy.ts`（1 处）
GET `/trade/bidding/statistics`

---

## 处理原则（对照 AGENTS.md §5）
- 补齐某个路由后，**同步删除**对应 `TODO(P1)` 注释与本清单条目。
- 不要把回退函数塞进 `KNOWN_PENDING` 白名单——白名单只登记「确实要做、暂未实现且会 404」的端点；
  这些函数已回退到相近路由（不 404），属于「语义不精确」而非「断裂」。
- 实现 backlog 路由时，业务表加 `org_id`，读取按活跃省过滤（`db.OrgFilter`）。
