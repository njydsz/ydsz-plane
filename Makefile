# Ydsz Plane — 常用开发命令（见 docs/architecture/03）

.PHONY: dev up down migrate seed seed-scale lint test build openapi reindex

# 本地开发：基础设施容器 + 后端热重载(air) + 前端 vite
dev: up
	@echo "→ start api (air) & web (vite) in two terminals:"
	@echo "  make dev-api"
	@echo "  make dev-web"

up:
	docker compose -f deployments/docker-compose.yml up -d postgres redis mailpit

up-full:
	docker compose -f deployments/docker-compose.yml --profile full up -d

down:
	docker compose -f deployments/docker-compose.yml down

migrate:
	go run ./cmd/migrate up

migrate-down:
	go run ./cmd/migrate down 1

seed: migrate
	go run ./scripts/seed

# 大规模造数（100 万需求/任务/缺陷，用于性能基线压测）
# 用法：make seed-scale COUNT=100000 PROJECT=1
seed-scale: migrate
	go run ./scripts/seed-scale -count=$(or $(COUNT),1000000) -project=$(or $(PROJECT),1) -workers=$(or $(WORKERS),8)

dev-api:
	air -c .air.toml || go run ./cmd/api

dev-worker:
	go run ./cmd/worker

dev-web:
	cd web && pnpm dev

GOPKGS := $(shell go list ./... | grep -v '/web/node_modules/')

lint:
	golangci-lint run $(GOPKGS)
	cd web && pnpm lint

test:
	go test $(GOPKGS) -race -count=1
	cd web && pnpm test

# 本地覆盖率：生成 coverage.out 并按函数列出覆盖情况
coverage:
	go test $(GOPKGS) -count=1 -coverprofile=coverage.out
	go tool cover -func=coverage.out

# 覆盖率 HTML 报告：coverage.html
coverage-html:
	go test $(GOPKGS) -count=1 -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html

build:
	go build $(GOPKGS)
	cd web && pnpm build

openapi:
	swag init -g cmd/api/main.go --output docs/swagger --parseDependency --parseInternal
	@echo "→ Swagger UI: http://localhost:8080/swagger/index.html"

reindex:
	go run ./scripts/reindex

# 性能压测：需 k6 已安装且后端已启动并 seed（脚本位于 tests/perf/）
# 用法：
#   make perf-smoke                      # 冒烟：核心端点可用性
#   make perf-load                       # 负载：10→100 VU，断言 P95<200ms
#   make perf-stress                     # 压力：200 VU 恒定 3 分钟
# 可用环境变量：BASE_URL / TEST_USER_EMAIL / TEST_USER_PASSWORD
perf-smoke:
	k6 run -e BASE_URL=$(or $(BASE_URL),http://127.0.0.1:8080/api/v1) \
		-e TEST_USER_EMAIL=$(or $(TEST_USER_EMAIL),admin@njydsz.com) \
		-e TEST_USER_PASSWORD=$(or $(TEST_USER_PASSWORD),Admin@1020) \
		tests/perf/smoke-test.js

perf-load:
	k6 run -e BASE_URL=$(or $(BASE_URL),http://127.0.0.1:8080/api/v1) \
		-e TEST_USER_EMAIL=$(or $(TEST_USER_EMAIL),admin@njydsz.com) \
		-e TEST_USER_PASSWORD=$(or $(TEST_USER_PASSWORD),Admin@1020) \
		tests/perf/load-test.js

perf-stress:
	k6 run -e BASE_URL=$(or $(BASE_URL),http://127.0.0.1:8080/api/v1) \
		-e TEST_USER_EMAIL=$(or $(TEST_USER_EMAIL),admin@njydsz.com) \
		-e TEST_USER_PASSWORD=$(or $(TEST_USER_PASSWORD),Admin@1020) \
		tests/perf/stress-test.js

perf-json:
	k6 run --out json=docs/perf/result.json \
		-e BASE_URL=$(or $(BASE_URL),http://127.0.0.1:8080/api/v1) \
		-e TEST_USER_EMAIL=$(or $(TEST_USER_EMAIL),admin@njydsz.com) \
		-e TEST_USER_PASSWORD=$(or $(TEST_USER_PASSWORD),Admin@1020) \
		tests/perf/load-test.js

fmt:
	gofmt -w .
	cd web && pnpm format

# --- S14: 微服务独立构建 ---

# 构建通知服务独立二进制
build-notification-svc:
	CGO_ENABLED=1 go build -ldflags="-s -w" -o bin/notification-service ./cmd/notification-service

# 构建搜索服务独立二进制
build-search-svc:
	CGO_ENABLED=1 go build -ldflags="-s -w" -o bin/search-service ./cmd/search-service

# 同时构建两个微服务
build-microservices: build-notification-svc build-search-svc

# 运行 S14 通知服务独立部署（依赖 notification_db 已初始化）
run-notification-svc: build-notification-svc
	./bin/notification-service

run-search-svc: build-search-svc
	./bin/search-service
