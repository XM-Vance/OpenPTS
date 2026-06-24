# AGENTS.md — 多 Agent 协作与提交规则

本仓库由多个 AI agent 协作开发。**动手前先读完本文件。** 违反这些规则会导致 CI 失败、
数据库损坏、或与其他 agent 的工作冲突。

---

## 0. 防冲突（最重要——这些坑真实发生过）

- **一个 agent 一个分支，一个独立工作目录。** 不要两个 agent 同时改同一分支、同一工作目录。
- **不要在共享工作目录里跑后台/并发 git 命令**（`git pull` / `checkout` / `merge`）。并发 git
  会把工作区搞成半合并的乱状态。git 操作一律前台、串行。
- **开工前先 `git fetch` 并查看现有开放 PR 与分支**（`gh pr list`、`git branch -r`），
  确认没有别的 agent 已经在做同一件事（曾出现两个 agent 各修一遍同一个 flaky 测试）。
- 改动**全局单点资源**（迁移序号、契约白名单、共享配置）前，务必基于**最新 main**，避免撞号/重复。

## 1. 分支与提交流程

> ⚠️ **main 已开启分支保护（强制，对管理员也生效）**：禁止直接 `git push origin main`；
> 必须走 PR，且 6 道 CI（backend / frontend / algo / docling / contract / release-guard）
> 全绿才能合并；禁止 force-push / 删除 main。直推会被 GitHub 拒绝——**别浪费时间试直推**。

1. 从**最新 main** 切分支：`git checkout main && git pull && git checkout -b type/简述`
   （type ∈ feat / fix / refactor / perf / chore / docs）。
2. 改动 → **本地验证全过**（见 §2）→ `git push`（推分支，不是 main）→ 开 PR → **CI 全绿后**再合并。
3. commit message 用中文：`type(scope): 描述`，若是 AI 协作可在结尾加：
   `Co-Authored-By: <模型名> <邮箱>`
4. PR 描述必须写清：**做了什么 / 为什么 / 怎么验证的**（仓库有 PR 模板，按模板填）。
5. 不要 force-push 共享分支；不要把别人的提交 rebase 掉。
6. **紧急逃生口**（仅 CI 基础设施抖动等卡死时，由仓库 owner 操作）：
   Settings → Branches 临时关闭保护，或 `gh api -X PUT .../branches/main/protection`
   把 `enforce_admins` 置 false，合并后立即恢复。**正常情况下不要动它。**

## 2. 提交前必须本地通过（CI 有六道闸，先在本地过）

- **后端**：`cd backend && go build ./... && go vet ./... && go test ./...`
- **前端**：`cd frontend && npx tsc --noEmit && npm run lint && npm run build`
  （lint 是 `--max-warnings 0`，一个 warning 都会挂）
- **Python(docling/algo)**：`python3 -m py_compile` 相关文件能过。
- 「本地能跑」不等于「CI 过」。六道闸：`backend / frontend / algo / docling / contract / release-guard`。
- **克隆后先 `make hooks`** 启用推送前自检（`.githooks/pre-push`，按改动路径自动跑上述检查，
  未过拦下推送；确需跳过用 `git push --no-verify`）。

## 3. 数据库迁移（高危区，最容易出事）

- 新迁移用**下一个序号**（查 `db/migrations/` 当前最大号 +1），**up + down 成对**、必须可逆。
- **可逆性只在一次性库验证**：`createdb ptis_chk` → `migrate up` → `migrate down 1` → `migrate up`。
  **绝不在 `ptis_dev` 或生产库跑 `migrate down`**（曾误删过 ptis_dev 全部数据）。
- **业务表**加 `org_id`（多租户隔离）；**共享参考表不加**（`md_*`、`carbon_quotes`、
  气象观测表、`auth_*`/`users`/`roles`/`*_permissions`）。
- 唯一约束要并入 `org_id`（`UNIQUE (org_id, ...)`）。

## 4. 多租户铁律（省份 = 组织）

- 业务**写入**必须作用于活跃省；总部「全部省」(`X-Org-Id: *`) 下写业务数据要返回 **400**，
  统一用 `handler.respondOrgRequired(c, err)` + repo 层 `db.ErrOrgRequired`。别绕过。
- 读写按活跃省过滤复用现成模式：`db.OrgFilter(ctx)` / `db.WithOrg(ctx, org)`。
- 不要给共享参考表加省份过滤（全国统一数据，各省看同一份）。

## 5. 契约门禁（前端 ↔ 后端）

- 前端 `lib/api/*.ts` 里新增的每个 `/api/v1/...` 调用，后端**必须有对应路由**，
  否则 contract 闸报 404 失败（401/403/405 算路由存在，只有 404 算失败）。
- 删除前端 api 模块里的死函数时，**同步删除** `scripts/check_contract.sh` 里 `KNOWN_PENDING`
  对应条目。**不要把死代码塞进白名单**——白名单只登记「确实要做、暂未实现」的端点。

## 6. 安全（release-guard 会卡）

- **绝不提交** `.env`、API 密钥（如 `ZHIPU_API_KEY`）、`*.env.backup`、编译二进制、`node_modules`。
- 密钥只走环境变量；仓库里只放 `.env.example` 占位。
- `.claude/settings.local.json` 等个人本地配置已在 `.gitignore`，不要提交。

## 7. 不要打断别人正在用的环境

- 本地验证起实例用**临时端口**（如 `:8099`），不要重启/杀掉别人正在用的服务
  （默认本地：后端 `:8081`、前端 `:3000`）。
- 不要在 `ptis_dev` 上做破坏性操作（删表、`migrate down`、清数据）；要可逆性测试就建一次性库。

## 8. 代码风格随大流

- 前端图表 recharts 走懒加载（`page.tsx` + `_view.tsx`/`_charts.tsx` + `next/dynamic`）；
  派生数据用 `useMemo`（否则 lint 挂）。菜单项绑权限码，按 RBAC 控制可见性。
- 后端按领域拆 repo/handler 文件；错误信息、注释用中文，与周边代码一致。

---

发布流程见仓库 README 与 Makefile；本文件聚焦「多 agent 怎么安全地往仓库提交」。
