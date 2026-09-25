# Ydsz Plane — 常用开发命令（见 docs/architecture/03）
# GNU Make 4+ compatible

# --- Phony targets (no file outputs) ---
.PHONY: help dev up down migrate migrate-down seed seed-scale \
        dev-api dev-worker dev-web smoke \
        lint test test-only coverage coverage-html test-html \
        build openapi gen build-types-dev \
        reindex fmt dev-secrets vet \
        build-notification-svc build-search-svc build-microservices \
        run-notification-svc run-search-svc \
        perf-smoke perf-load perf-stress perf-json

# --- Help (self-documenting) ---
## help: 列出所有可用 make 目标及说明
help:
	@echo "Ydsz Plane — 开发命令"
	@echo ""
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z0-9_-]+:.*?## / {printf "  \033[36m%-22s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)
	@echo ""
	@echo "带参数目标示例："
	@echo "  make seed-scale COUNT=100000 PROJECT=1 WORKERS=8"
	@echo "  make perf-smoke BASE_URL=http://127.0.0.1:8080/api/v1"

# --- Infrastructure ---
## dev: 启动基础设施容器并提示开发终端命令
dev: up
	@echo "→ start api (air) & web (vite) in two terminals:"
	@echo "  make dev-api"
	@echo "  make dev-web"

## up: 启动基础设施容器 (postgres + redis + mailpit)
up:
	docker compose -f deployments/docker-compose.yml up -d postgres redis mailpit

## up-full: 启动完整基础设施容器（含 ES + MinIO）
up-full:
	docker compose -f deployments/docker-compose.yml --profile full up -d

## down: 停止所有容器
down:
	docker compose -f deployments/docker-compose.yml down

# --- Database ---
## migrate: 执行数据库迁移（upgrade）
migrate:
	go run ./cmd/migrate up

## migrate-down: 回滚一次迁移
migrate-down:
	go run ./cmd/migrate down 1

## seed: 迁移后初始化种子数据
seed: migrate
	go run ./scripts/seed

## seed-scale: 大规模造数（性能基线压测，用法: make seed-scale COUNT=100000 PROJECT=1 WORKERS=8）
seed-scale: migrate
	go run ./scripts/seed-scale -count=$(or $(COUNT),1000000) -project=$(or $(PROJECT),1) -workers=$(or $(WORKERS),8)

# --- Dev servers ---
## dev-api: 启动 API 热重载（air fallback go run）
dev-api:
	air -c .air.toml || go run ./cmd/api

## dev-worker: 启动异步 Worker
dev-worker:
	go run ./cmd/worker

## dev-web: 启动前端开发服务器（vite）
dev-web:
	cd web && pnpm dev

# --- Local smoke test ---
## smoke: 本地冒烟（构建 + 启动 + 校验 healthz/readyz）
smoke:
	@echo "→ build + smoke: 启动 API 并校验 /healthz 与 /readyz"
	go build -o /tmp/ydsz-api ./cmd/api
	/tmp/ydsz-api > /tmp/ydsz-api.log 2>&1 & echo $$! > /tmp/ydsz-api.pid
	@for i in $$(seq 1 30); do \
		if curl -sf http://localhost:8080/healthz > /dev/null 2>&1; then \
			echo "  ✅ API started after $${i}s"; break; \
		fi; sleep 1; \
	done
	@echo "→ /healthz:" && curl -s http://localhost:8080/healthz
	@echo ""
	@echo "→ /readyz:"  && curl -s http://localhost:8080/readyz
	@echo ""
	@if [ -f /tmp/ydsz-api.pid ]; then kill $$(cat /tmp/ydsz-api.pid) 2>/dev/null || true; fi

# --- Lint & vet ---
## lint: golangci-lint + 前端 lint
lint: vet
	golangci-lint run $(GOPKGS)
	cd web && pnpm lint

## vet: 静态分析（go vet 全包）
vet:
	go vet $(GOPKGS)

# --- Test ---
GOPKGS := $(shell go list ./... | grep -v '/web/node_modules/')

## test: go vet + 全量测试（race），输出 go-test-report.txt
test: vet
	go test $(GOPKGS) -race -count=1 2>&1 | tee go-test-report.txt
	cd web && pnpm test

## test-only: 纯 go test（不带 vet，用于快速迭代）
test-only:
	go test $(GOPKGS) -race -count=1 2>&1 | tee go-test-report.txt

## coverage: 本地覆盖率（coverage.out + 函数级报告）
coverage:
	go test $(GOPKGS) -count=1 -coverprofile=coverage.out
	go tool cover -func=coverage.out

## coverage-html: 覆盖率 HTML 报告（coverage.html）
coverage-html:
	go test $(GOPKGS) -count=1 -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html

## test-html: 运行测试并打开 coverage HTML（跨平台）
test-html: coverage-html
	@echo "Opening coverage report..."
ifeq ($(OS),Windows_NT)
	@start coverage.html
else
	@xdg-open coverage.html 2>/dev/null || open coverage.html 2>/dev/null || echo "coverage.html ready"
endif

# --- OpenAPI & type generation ---
## openapi: 生成 Swagger 文档（swag init）
openapi:
	swag init -g cmd/api/main.go --output docs/swagger --parseDependency --parseInternal
	@echo "→ Swagger UI: http://localhost:8080/swagger/index.html"

## gen: 同步生成 swagger + 前端类型（保持前后端一致）
gen: openapi build-types-dev

build-types-dev:
	cd web && npx openapi-typescript ../docs/swagger/swagger.yaml -o src/types/api.generated.ts --path-params-as-hash-map false
	@echo "Generated web/src/types/api.generated.ts"

# --- DevOnly secret tool ---
## dev-secrets: 生成 YDSZ_SSO_SECRET_KEY 并写入 .env.local（gitignored，不覆盖已有文件）
dev-secrets:
	@echo "→ DevOnly: 生成随机 SSO secret key（32 字节 hex）"
	go run ./cmd/dev_secrets
	@echo "⚠  本工具仅用于本地开发！生产环境必须通过 Vault / K8s Secret 注入。"

# --- Reindex ---
## reindex: 重建搜索索引
reindex:
	go run ./scripts/reindex

# --- Format ---
## fmt: 格式化 Go 代码 + 前端代码
fmt:
	gofmt -w .
	cd web && pnpm format

# --- Build ---
## build: 构建后端（全包）+ 前端
build:
	go build $(GOPKGS)
	cd web && pnpm build

# --- S14: 微服务独立构建 ---
## build-notification-svc: 构建通知服务独立二进制
build-notification-svc:
	CGO_ENABLED=1 go build -ldflags="-s -w" -o bin/notification-service ./cmd/notification-service

## build-search-svc: 构建搜索服务独立二进制
build-search-svc:
	CGO_ENABLED=1 go build -ldflags="-s -w" -o bin/search-service ./cmd/search-service

## build-microservices: 同时构建两个微服务
build-microservices: build-notification-svc build-search-svc

## run-notification-svc: 运行通知服务（需先 build-notification-svc）
run-notification-svc: build-notification-svc
	./bin/notification-service

## run-search-svc: 运行搜索服务（需先 build-search-svc）
run-search-svc: build-search-svc
	./bin/search-service

# --- Performance benchmarks ---
## perf-smoke: k6 冒烟测试（核心端点可用性）
perf-smoke:
	k6 run -e BASE_URL=$(or $(BASE_URL),http://127.0.0.1:8080/api/v1) \
		-e TEST_USER_EMAIL=$(or $(TEST_USER_EMAIL),admin@njydsz.com) \
		-e TEST_USER_PASSWORD=$(or $(TEST_USER_PASSWORD),Admin@1020) \
		tests/perf/smoke-test.js

## perf-load: k6 负载测试（10→100 VU，断言 P95<200ms）
perf-load:
	k6 run -e BASE_URL=$(or $(BASE_URL),http://127.0.0.1:8080/api/v1) \
		-e TEST_USER_EMAIL=$(or $(TEST_USER_EMAIL),admin@njydsz.com) \
		-e TEST_USER_PASSWORD=$(or $(TEST_USER_PASSWORD),Admin@1020) \
		tests/perf/load-test.js

## perf-stress: k6 压力测试（200 VU 恒定 3 分钟）
perf-stress:
	k6 run -e BASE_URL=$(or $(BASE_URL),http://127.0.0.1:8080/api/v1) \
		-e TEST_USER_EMAIL=$(or $(TEST_USER_EMAIL),admin@njydsz.com) \
		-e TEST_USER_PASSWORD=$(or $(TEST_USER_PASSWORD),Admin@1020) \
		tests/perf/stress-test.js

## perf-json: k6 负载测试并输出 JSON 报告
perf-json:
	k6 run --out json=docs/perf/result.json \
		-e BASE_URL=$(or $(BASE_URL),http://127.0.0.1:8080/api/v1) \
		-e TEST_USER_EMAIL=$(or $(TEST_USER_EMAIL),admin@njydsz.com) \
		-e TEST_USER_PASSWORD=$(or $(TEST_USER_PASSWORD),Admin@1020) \
		tests/perf/load-test.js
