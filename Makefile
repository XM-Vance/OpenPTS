# PTIS 交付脚本：一键起栈 / 迁移 / 种子 / 测试 / 截图归档。
# 用法：make help
#
# 依赖：docker compose / migrate (golang-migrate) / curl / jq（可选）

.PHONY: help hooks up up-local down restart logs ps migrate-up migrate-down seed test test-be test-fe lint clean smoke verify k6-smoke monitoring monitoring-down

# up-local 交叉编译的目标架构：默认取本机（Apple Silicon→arm64，x86→amd64）；
# 给 x86 服务器出包时显式覆盖：make up-local LOCAL_GOARCH=amd64
LOCAL_GOARCH ?= $(shell uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/')

DEFAULT_GOAL := help

## ─── 基础信息 ───
help: ## 显示此帮助
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

hooks: ## 启用 git 钩子（推送前本地自检，见 .githooks/pre-push）
	git config core.hooksPath .githooks
	@echo "已启用 .githooks（推送前自检）。临时跳过：git push --no-verify"

## ─── Docker 编排 ───
up: ## 构建并启动全部服务（含 nginx，容器内多阶段构建）
	docker compose up -d --build

up-local: ## 宿主机预构建后启动（离线/内网环境，不拉取 golang/node 基础镜像）
	cd backend && \
	  CGO_ENABLED=0 GOOS=linux GOARCH=$(LOCAL_GOARCH) go build -trimpath -ldflags="-s -w" -o ptis-server-linux ./cmd/server && \
	  CGO_ENABLED=0 GOOS=linux GOARCH=$(LOCAL_GOARCH) go build -trimpath -ldflags="-s -w" -o ptis-seed-linux   ./cmd/seed
	cd frontend && npm ci --no-audit --no-fund && rm -f .env.local && npm run build
	docker compose -f docker-compose.yml -f docker-compose.override.yml -f docker-compose.local-build.yml up -d --build

down: ## 停止全部服务（保留卷）
	docker compose down

restart: ## 重启 backend + frontend
	docker compose restart backend frontend

logs: ## 跟踪 backend 日志
	docker compose logs -f backend

ps: ## 服务状态
	docker compose ps

## ─── 数据库 ───
migrate-up: ## 应用所有迁移
	docker compose --profile migrate run --rm migrate up

migrate-down: ## 回滚 1 步
	docker compose --profile migrate run --rm migrate down 1

seed: ## 写入种子数据（随机生成 admin 密码并打印 + 系统角色）
	docker compose exec backend /app/seed

## ─── 监控栈 ───
monitoring: ## 启动监控栈（Prom + Grafana + Alertmanager + exporters）
	docker compose -f docker-compose.monitoring.yml up -d

monitoring-down: ## 停止监控栈
	docker compose -f docker-compose.monitoring.yml down

## ─── 测试 ───
test: test-be test-fe ## 跑全部测试

test-be: ## Go 单测（auth + middleware）
	cd backend && go test -race ./...

test-fe: ## 前端 vitest 单测
	cd frontend && npm test

k6-smoke: ## k6 冒烟测试（5 个热点接口）
	k6 run perf/smoke.js

## ─── 联调验证 ───
smoke: ## 健康检查 + 指标端点（无需 docker，要求后端已在本机 :8080 运行）
	@curl -s http://localhost:8080/health | head -c 200 && echo
	@curl -s http://localhost:8080/metrics | head -c 200 && echo

verify: ## 完整功能验证清单（依次调多个关键端点）
	@bash scripts/verify.sh

## ─── 工具 ───
lint: ## go vet + tsc
	cd backend && go vet ./...
	cd frontend && npx tsc --noEmit

clean: ## 清理构建产物
	rm -rf backend/dist frontend/.next frontend/out
