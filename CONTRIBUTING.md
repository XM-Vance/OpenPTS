# 贡献指南

感谢你参与本项目！以下是协作约定。

## 分支与提交

- 主分支受保护，**禁止直推**，所有改动走 Pull Request。
- 从主分支切出特性分支：
  ```bash
  git checkout main && git pull
  git checkout -b feat/简述      # 新功能：feat/；修复：fix/；文档：docs/
  ```
- 提交前本地自检：
  - 后端：`cd backend && go build ./... && go vet ./... && go test ./...`
  - 前端：`cd frontend && npm run type-check && npm run lint`

## 代码规范

- **后端（Go）**：遵循 `gofmt` / `go vet`；新增 repository 保持纯 CRUD，业务逻辑放 handler；多租户读写必须经 `OrgFilter` 过滤。
- **前端（TypeScript）**：`tsc --noEmit` 必须通过；API 客户端按业务域一文件（`lib/api/<域>.ts`）；复用 `components/` 下既有组件。
- **数据库迁移**：用 `golang-migrate` 成对的 `*.up.sql` / `*.down.sql`；**迁移必须可回滚**（down 能撤清 up 的改动）；迁移文件序号递增，不修改已合入的历史迁移。

## 算法接入的边界

本项目是**通用运营框架**，不含算法/规则内核。贡献时请注意：

- ✅ 欢迎：业务 CRUD、鉴权/权限、UI 优化、文档解析、性能、测试、通用基础设施。
- ⚠️ 算法相关（负荷/价格预测、竞价策略）：保持**接口契约**稳定即可，具体算法实现由各部署方私有接入，不要把具体算法逻辑提交进来。
- ⚠️ 规则数据：`trade_rules` 等表保持**通用结构**，不要提交具体省份/市场的规则参数数据。

## 测试

- 后端单测：`make test-be`
- 前端单测：`make test-fe`
- 契约校验（前后端 API 一致性）：`bash scripts/check_contract.sh`

## 多租户铁律

所有业务表都带 `org_id`，**写操作必须校验当前活跃组织**、**读操作必须按组织过滤**。
新增任何带数据的表/接口时，务必接入多租户隔离，避免数据越权。详见 `docs/ARCHITECTURE.md`。

## 提交 PR

1. 推送分支并开 PR，按模板描述改动与测试情况。
2. 确保 CI 全绿。
3. 等待 review；通过后合并。
