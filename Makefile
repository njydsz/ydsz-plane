# Ydsz Plane — 常用开发命令（见 docs/architecture/03）

.PHONY: dev up down migrate seed lint test build openapi reindex

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

dev-api:
	air -c .air.toml || go run ./cmd/api

dev-worker:
	go run ./cmd/worker

dev-web:
	cd web && pnpm dev

GOPKGS := ./cmd/... ./internal/... ./pkg/... ./scripts/...

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

# 性能压测：需 k6 已安装且后端已启动并 seed
perf:
	k6 run scripts/k6/smoke.js

perf-json:
	k6 run --out json=scripts/k6/result.json scripts/k6/smoke.js

fmt:
	gofmt -w .
	cd web && pnpm format
