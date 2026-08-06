# Ydsz Plane — 常用开发命令（见 docs/architecture/03）

.PHONY: dev up down migrate seed lint test build openapi reindex

# 本地开发：基础设施容器 + 后端热重载(air) + 前端 vite
dev: up
	@echo "→ start api (air) & web (vite) in two terminals:"
	@echo "  make dev-api"
	@echo "  make dev-web"

up:
	docker compose -f deployments/docker-compose.yml up -d postgres redis nats

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

lint:
	golangci-lint run ./...
	cd web && pnpm lint

test:
	go test ./... -race -count=1
	cd web && pnpm test

build:
	go build ./...
	cd web && pnpm build

openapi:
	swag init -g cmd/api/main.go -o api/openapi

reindex:
	go run ./scripts/reindex

fmt:
	gofmt -w .
	cd web && pnpm format
