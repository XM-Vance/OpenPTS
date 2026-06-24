.PHONY: run build test tidy lint clean seed

# 本地运行网关
run:
	go run ./cmd/server

# 编译产物
build:
	mkdir -p bin
	go build -o bin/server ./cmd/server
	go build -o bin/seed ./cmd/seed

# 单元测试
test:
	go test -v -race ./...

# 整理依赖
tidy:
	go mod tidy

# 代码检查（需要 golangci-lint）
lint:
	golangci-lint run ./...

# 创建默认管理员 admin（随机密码，启动时打印到日志）
seed:
	go run ./cmd/seed

# 清理产物
clean:
	rm -rf bin
