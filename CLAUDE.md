# CLAUDE.md

本项目的协作与提交规则见 **[AGENTS.md](AGENTS.md)** —— 动手改代码或提交前请先读完。

要点速记（详情见 AGENTS.md）：
- **main 已分支保护（强制）**：禁止直接 push main，必须走 PR + 6 道 CI 全绿才能合；直推会被拒。
- 一个 agent 一分支一工作目录；git 操作前台串行，先 `git fetch` + 看现有 PR 防重复。
- 提交前本地必过：后端 `go build/vet/test`，前端 `tsc + lint(--max-warnings 0) + build`。
- 迁移可逆、只在一次性库测 down，**绝不在 ptis_dev/生产跑 down**；业务表加 org_id，共享表不加。
- 多租户：业务写入按活跃省，「全部省」写入返回 400（`respondOrgRequired`）。
- 契约门禁：前端调的 `/api/v1` 路径后端必须存在；删死代码同步清白名单。
- 绝不提交 .env / 密钥 / 二进制。
