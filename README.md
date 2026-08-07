<p align="center">
  <h1 align="center">Ydsz Plane</h1>
  <p align="center">
    面向中国软件团队的开源项目管理平台
  </p>
</p>

<p align="center">
  <a href="#快速开始"><b>快速开始</b></a> ·
  <a href="docs/architecture/README.md"><b>架构文档</b></a> ·
  <a href="#路线图"><b>路线图</b></a> ·
  <a href="#贡献指南"><b>参与贡献</b></a>
</p>

<p align="center">
  Go · Vue 3 · PostgreSQL · Redis · NATS · Docker
</p>

---

## 项目状态

> **当前阶段：M0 ✅ + M0.5（增强基座）✅ — M1 架构就绪，待 Workspace CRUD API & 前端落地**
> 最后更新：2026-08-08 · 架构基线版本 v1.1

Ydsz Plane 是一款开源、自托管的现代项目管理工具，专为中小型敏捷开发团队设计。M0 工程基座及 M0.5 增强基座（完整可观测性、RBAC、安全纵深、API 文档）已全部完工；数据库 schema（users / workspaces / workspace_members / password_reset_tokens / audit_logs / domain_events）已就位，下一步落地 Workspace CRUD API 及前端即可进入 M2 工作项核心。

## 已完成能力

截至 2026-08-08，仓库中已实现的能力：

| 模块 | 说明 | 状态 |
|------|------|------|
| 工程基座 | Monorepo（Go module + pnpm workspace）、CI（lint/test/build/e2e-smoke）、Docker Compose（pg/redis/nats）、Makefile | ✅ |
| 鉴权链路 | 注册 / 登录（bcrypt + JWT access/refresh）、Cookie 会话、401 单飞刷新重放、忘记 / 重置密码端点 | ✅ |
| RBAC | Owner/Admin/Member/Guest 四角色 × workspace_members 域、10 项权限中间件、权限 | ✅ |
| API 文档 | swaggo 注解 + Swagger UI（`/swagger/index.html`）、7 条端点、Bearer 鉴权说明 | ✅ |
| 邮件服务 | SMTP 抽象（Noop / SMTP 自动切换）、双版本 MIME 模板（密码重置 / 邀请） | ✅ |
| 可观测性 | zap 结构化日志、Prometheus RED 指标（`/metrics`）、Go runtime 默认收集 | ✅ |
| 安全纵深 | CSP / HSTS / COOP / CORP / Permissions-Policy / X-Frame-Options / X-Content-Type-Options 等 8 项安全头 | ✅ |
| 中间件链 | RequestID → Recovery → CORS → SecurityHeaders → RateLimit → AccessLog → Metrics → Auth | ✅ |
| 前端骨架 | Vue 3.5 + Vite 6 + Pinia、登录页（亮/暗主题）、Axios 拦截器、auth store | ✅ |
| 数据持久层 | pgx 连接池、租户上下文（SET LOCAL app.workspace_id）、RLS 策略模板、迁移工具 | ✅ |
| 事件骨架 | 事务型 Outbox 表 + Relay（DB → NATS）、Asynq Worker（default/notifications/automation 队列） | ✅ |
| 数据库迁移 | 0001_init（users / workspaces / workspace_members / domain_events / audit_logs）+ 0002_password_reset | ✅ |
| 种子数据 | 5 用户 + 3 工作空间 + 多角色成员（owner/admin/member/guest）+ 幂等执行 | ✅ |

## 技术栈

| 层次 | 技术选型 | 说明 |
|------|----------|------|
| 后端 | Go 1.25 + Gin 1.12 | 模块化单体（DDD 轻量分层） |
| 前端 | Vue 3.5 + TypeScript + Vite 6 | 组合式 API、Pinia 状态管理 |
| 数据库 | PostgreSQL 16 | ACID + JSONB + RLS 租户隔离 + 信创方言预留 |
| 缓存 | Redis 7 | 限流、分布式锁、会话辅助 |
| 事件 | NATS 2.10（JetStream） | Outbox 投递、实时推送扇出 |
| 任务队列 | Asynq | 异步任务（通知、索引、Webhook、自动化） |
| 全文检索 | Elasticsearch 8（可选 profile） | 全局搜索、分词（IK） |
| 对象存储 | MinIO（可选 profile） | 附件、Logo |
| 部署 | Docker Compose（一键）/ K8s（Phase 3） | 信创兼容：openEuler/麒麟 + ARM64 |

## 系统架构

```
                        ┌────────────────────────────┐
                        │     Web SPA (Vue3+Vite)     │
                        │  Design Tokens / Plane 风格  │
                        └──────────────┬─────────────┘
                                       │ HTTPS / WSS
                        ┌──────────────▼─────────────┐
                        │   Nginx（反代/静态/限流）    │
                        └──────────────┬─────────────┘
                                       │
              ┌────────────────────────▼────────────────────────┐
              │            ydsz-plane-api (Go + Gin)            │
              │  Interfaces（HTTP + Middleware）                │
              │  Application Services（用例编排/事务边界）        │
              │  Domain（限界上下文：iam / workspace / project   │
              │      / issue / sprint / version / ...）         │
              │  Infrastructure（PG / Redis / ES / NATS）       │
              └───────┬───────────────────────┬────────────────┘
                      │ 写事件 (Outbox)        │ 读
        ┌─────────────▼──────────┐   ┌────────▼───────┐
        │  ydsz-plane-worker     │   │ PostgreSQL 16  │
        │  (Asynq + Outbox Relay)│   │ Redis 7        │
        │  · 通知投递             │   │ NATS / ES / MinIO
        │  · ES 索引同步          │   └────────────────┘
        │  · Webhook 分发         │
        │  · 自动化规则执行        │
        │  · 迭代快照 / 效能计算   │
        └────────────────────────┘
```

## 项目结构

```
ydsz-plane/
├── cmd/
│   ├── api/main.go            # API Server 入口
│   ├── worker/main.go         # 异步 Worker（Outbox Relay + Asynq Consumer）
│   └── migrate/main.go        # 数据库迁移执行（golang-migrate）
├── internal/
│   ├── application/
│   │   └── auth/              # 登录 / 令牌签发 / 令牌解析（已实现）
│   ├── infrastructure/
│   │   ├── persistence/       # pgx 连接池 + 租户上下文
│   │   ├── cache/             # Redis 客户端
│   │   ├── events/            # Outbox Relay（DB → NATS）
│   │   └── telemetry/         # zap 结构化日志
│   ├── interfaces/
│   │   ├── http/              # Gin 路由 + Handler
│   │   └── middleware/        # 中间件链
│   └── config/                # Viper 配置加载 + 校验
├── pkg/
│   └── errs/                  # 统一错误类型 + 错误码注册
├── migrations/                # 递增编号迁移脚本（0001_init.up/down.sql）
├── scripts/seed/              # 开发环境种子数据
├── deployments/
│   ├── docker-compose.yml     # 核心栈 + full profile（ES + MinIO）
│   ├── Dockerfile.api         # 多阶段构建（api + worker + migrate）
│   ├── Dockerfile.web         # 前端 Nginx 静态服务
│   └── nginx/web.conf
├── docs/
│   ├── architecture/          # 架构设计文档（14 份，含完整设计）
│   └── Ydsz Plane PRD-终极完整版.docx
├── web/                       # 前端（pnpm workspace）
│   ├── src/
│   │   ├── api/               # Axios 客户端（含 401 自动刷新）
│   │   ├── views/             # 页面（Login / Home / ProjectList / NotFound）
│   │   ├── stores/            # Pinia（auth）
│   │   ├── router/            # 路由 + 权限守卫
│   │   ├── layouts/           # WorkspaceLayout
│   │   └── design/            # 设计令牌（CSS 变量）
│   └── packages/
│       ├── design-tokens/     # 主题令牌包
│       └── ui/                # 基础组件门面（建设中）
├── .github/workflows/ci.yml   # lint / test(race) / build / e2e-smoke
├── Makefile                   # dev / up / migrate / seed / lint / test
└── .env.example               # 环境变量模板（YDSZ_ 前缀）
```

## 快速开始

> 📖 **[本地开发环境配置](docs/本地开发环境配置.md)** - 包含服务地址、账号密码、快速启动指南

### Docker Compose（推荐）

```bash
# 克隆项目
git clone https://github.com/njydsz/ydsz-plane.git
cd ydsz-plane

# 启动核心服务（PostgreSQL + Redis + NATS + API + Worker + Web）
docker compose -f deployments/docker-compose.yml up -d

# 启动完整栈（额外包含 Elasticsearch + MinIO）
docker compose -f deployments/docker-compose.yml --profile full up -d

# 访问 http://localhost （前端） / API 运行在 8080
```

### 本地开发

#### 前置要求

- Go 1.25+
- Node.js 20+
- pnpm 10+
- Docker & Docker Compose

#### 启动基础设施 + 后端

```bash
# 1. 启动基础设施容器 (PostgreSQL + Redis + NATS + Mailpit)
make up

# 2. 配置环境变量（已默认配置为本地开发环境，可直接复制）
cp .env.example .env

# 3. 运行数据库迁移
make migrate      # go run ./cmd/migrate up

# 4. 导入种子数据（admin@ydsz.dev / Admin@123）
make seed         # go run ./scripts/seed

# 5. 启动 API 服务（热重载需要 air）
make dev-api      # air -c .air.toml
# 或者
go run ./cmd/api
```

#### 启动前端

```bash
cd web

# 安装依赖
pnpm install

# 启动开发服务器（http://127.0.0.1:5173）
pnpm dev
```

#### 启动 Worker（可选，本地调试异步任务时使用）

```bash
make dev-worker   # go run ./cmd/worker
```

#### 常用服务端口

| 服务 | 地址 | 说明 |
|------|------|------|
| API Server | http://127.0.0.1:8080 | 后端 API |
| Swagger UI | http://127.0.0.1:8080/swagger/index.html | API 文档 |
| Web (Vite) | http://127.0.0.1:5173 | 前端开发服务器 |
| Mailpit UI | http://127.0.0.1:8025 | 邮件测试 Web UI |

> 完整服务连接信息（Redis/MySQL/MinIO 等）请参考 [本地开发环境配置](docs/本地开发环境配置.md)

## Makefile 命令速查

| 命令 | 说明 |
|------|------|
| `make dev` | 启动基础设施容器 + 提示 dev-api / dev-web |
| `make up` | 启动 pg + redis + nats |
| `make down` | 停止所有容器 |
| `make migrate` | 执行数据库迁移到最新 |
| `make migrate-down` | 回滚 1 步 |
| `make seed` | 导入种子数据（幂等） |
| `make dev-api` | 启动 API（air 热重载） |
| `make dev-worker` | 启动 Worker |
| `make dev-web` | 启动前端 dev server |
| `make lint` | Go + 前端全量 lint |
| `make test` | Go（race）+ 前端全量测试 |
| `make build` | Go + 前端全量构建 |
| `make openapi` | 生成 OpenAPI 规范（swaggo） |

## API 设计概览

- 风格：REST + 统一 envelope；版本前缀 `/api/v1`
- 认证：Cookie（Web SPA）/ Bearer Token（API 调用）
- 错误格式：`{"error":{"code":"AUTH.INVALID_CREDENTIALS","message":"...","request_id":"..."}}`
- 中间件：RequestID → Recovery → CORS → RateLimit → Auth

已实现的路由（dev server http://localhost:8080）：

```
GET  /healthz                              健康检查
GET  /readyz                               就绪检查（含 PG / Redis 探针）
GET  /metrics                              Prometheus RED 指标
GET  /swagger/index.html                   Swagger UI（dev 模式）
POST /api/v1/auth/login                    邮箱 + 密码登录（返回 access/refresh + cookie）
POST /api/v1/auth/refresh                  通过 refresh cookie 换发令牌对
POST /api/v1/auth/register                 新用户注册（需 features.registration_open=true）
POST /api/v1/auth/forgot-password          触发密码重置邮件（202，防枚举）
POST /api/v1/auth/reset-password           用一次性 token 重置密码
GET  /api/v1/me                            Bearer / Cookie 获取当前用户简介
WS   /api/v1/workspaces/:id/members        列工作空间成员（workspace:read 权限校验示例）
```

本地凭证：`make docker up && make migrate && make seed`

## 架构文档

详细设计见 [`docs/architecture/`](docs/architecture/README.md)，共 14 份文档覆盖从工程基座到测试质量的全链路设计：

| # | 文档 | 内容 |
|---|------|------|
| 01 | [系统架构设计](docs/architecture/01-系统架构设计.md) | 总体架构、ADR 决策、领域上下文 |
| 02 | [架构开发详细计划](docs/architecture/02-架构开发详细计划.md) | 12 Sprint WBS、里程碑、风险登记 |
| 03 | [工程基座与开发规范](docs/architecture/03-工程基座与开发规范.md) | Monorepo、CI/CD、分支策略、编码规范 |
| 04 | [数据模型设计](docs/architecture/04-数据模型设计.md) | DDL、索引、RLS、信创方言 |
| 05 | [API 设计规范](docs/architecture/05-API设计规范.md) | REST 约定、错误码、幂等、限流、WS |
| 06 | [权限与安全设计](docs/architecture/06-权限与安全设计.md) | RBAC、认证、威胁模型、等保三级 |
| 07 | [工作项与状态机设计](docs/architecture/07-工作项与状态机设计.md) | Issue 聚合、WBS、状态机 |
| 08 | [迭代与版本日设计](docs/architecture/08-迭代与版本日设计.md) | Sprint 生命周期、燃尽图、发布 |
| 09 | [通知与实时推送设计](docs/architecture/09-通知与实时推送设计.md) | 通知管道、订阅、WS 扇出 |
| 10 | [全局搜索设计](docs/architecture/10-全局搜索设计.md) | ES mapping、类 JQL 语法 |
| 11 | [仪表盘与效能度量设计](docs/architecture/11-仪表盘与效能度量设计.md) | Widget 框架、DORA 指标 |
| 12 | [开放集成设计](docs/architecture/12-开放集成设计.md) | Webhook、OpenAPI 治理、自动化 DSL |
| 13 | [部署运维与可靠性设计](docs/architecture/13-部署运维与可靠性设计.md) | SLO、备份、容量、发布管理 |
| 14 | [测试策略与质量保障](docs/architecture/14-测试策略与质量保障.md) | 测试金字塔、专项清单 |

## 路线图

### M0 工程基座 ✅ 已完成（S1，W2）

- [x] Monorepo 初始化（Go module + pnpm workspace）
- [x] CI/CD（GitHub Actions: lint / test(race) / build / e2e-smoke）
- [x] Docker Compose 全栈 + Makefile
- [x] Gin 骨架 + 中间件链 + 健康检查
- [x] 数据库迁移系统 + RLS 模板 + 租户上下文
- [x] Outbox + Asynq + NATS 事件骨架
- [x] 前端骨架 + 设计令牌 + 登录页 + 鉴权链路
- [x] 种子数据（5 用户 + 3 工作空间，幂等）

### M0.5 增强基座 ✅ 已完成（额外加固，S1.5）

- [x] Prometheus RED 指标 + `/metrics` + Go runtime 收集器
- [x] RBAC 四角色 × 10 项权限中间件 + 单测
- [x] 安全纵深（CSP/HSTS/COOP/CORP/Permissions-Policy/X-Frame-Options 等 8 项头）
- [x] swaggo 完整 Swagger 注解 + Swagger UI + Bearer 鉴权
- [x] 邮件 SMTP 抽象（Noop/SMTP 自动切换）+ 双 MIME 模板（重置/邀请）
- [x] JWT Secret dev 随机生成 + production fail-fast
- [x] 密码重置 token 表（0002 迁移）+ 端点接入
- [x] Axios 拦截器（401 单飞刷新 + 限流 429 回调 + Request ID）
- [x] 统一 authApi Service 层 + auth Store

### M1 租户与项目骨架（S2，架构就绪 → CRUD & 前端待落地）

#### 后端 schema 已就绪 ✅

- [x] users / workspaces / workspace_members / password_reset_tokens / audit_logs
- [x] 鉴权链路（注册/登录/刷新/RBAC）
- [x] 邮件抽象 + 模板
- [ ] **Workspace CRUD API**（创建/读取/更新/归档/恢复）
- [ ] **成员邀请**（invitations 表 + 邮件链接 + 7 天有效 + 审核模式）
- [ ] **API Token** 管理（创建/吊销/scopes）
- [ ] **审计日志** 埋点（空间级管理操作）

#### 前端待建设 ⏳

- [ ] 空间列表 / 创建向导 / 设置页（信息/成员/邀请）
- [ ] 成员管理页（列表/筛选/角色切换/移除二次确认）
- [ ] 注册 / 忘记密码 / 重置密码页
- [ ] WorkspaceLayout 完整侧边栏 + 项目导航

### M2 工作项核心（S3–S4）

- [ ] Issue 主表 + 状态机配置 + 三级 WBS
- [ ] 需求 / 任务 / 缺陷全流程
- [ ] 看板 / 列表视图
- [ ] 活动日志时间线

### M3–M4：迭代与质量（S5–S7）

- [ ] Sprint 生命周期、燃尽图
- [ ] 版本日聚合、Release Notes
- [ ] 附件管理、站内通知
- [ ] MVP v0.1 发布

### Phase 2+（S8–S12）

- [ ] 全局搜索（ES + 类 JQL 语法）
- [ ] 项目仪表盘、个人工作台
- [ ] 通知多渠道（IM 企微/钉钉/飞书）
- [ ] Webhook / OpenAPI 集成
- [ ] 自动化引擎（Trigger-Condition-Action）
- [ ] 研发效能度量（DORA、CFD）

### Phase 3–4（远期）

- [ ] 知识库、收件箱
- [ ] 甘特图 / 日历 / 电子表格视图
- [ ] SSO / SAML 集成
- [ ] 国际化、PWA
- [ ] 数据迁移工具（Jira / 云效 / ONES 导入）
- [ ] AI 功能（智能分配、重复检测）

## 竞品对标

| 能力域 | 云效 | TAPD | ONES | PingCode | Jira | Ydsz Plane |
|--------|------|------|------|----------|------|------------|
| 需求管理 | ● | ● | ● | ● | ● | △ |
| 任务管理 | ● | ● | ● | ● | ● | △ |
| 缺陷管理 | ● | ● | ● | ● | ● | △ |
| 迭代管理 | ● | ● | ● | ● | ● | △ |
| 项目仪表盘 | ● | ● | ● | ● | ● | △ |
| 效能度量 | ● | ○ | ● | ● | ○ | △ |
| **开源自托管** | ✕ | ✕ | △ | △ | ✕ | **●** |
| **信创兼容** | ✕ | ✕ | △ | △ | ✕ | **●** |

> ● 已支持 / △ 设计中 / ○ 不支持 / ✕ 不适用

## 信创兼容

- **操作系统**：麒麟 V10 / 统信 UOS / openEuler
- **CPU 架构**：x86_64 / ARM64（鲲鹏 / 飞腾）
- **数据库**：PostgreSQL / 达梦 / 人大金仓（方言层抽象）
- **中间件**：Nginx / 东方通 / 宝兰德
- **浏览器**：360 安全浏览器 / 奇安信
- **密码算法**：预留 SM2/SM3/SM4 国密算法接口（build tag 隔离）
- **等保合规**：满足等保三级基线要求

## 性能指标（目标）

| 指标 | 目标值 |
|------|--------|
| 页面加载 P95 | ≤ 2s |
| API 响应 P95 | ≤ 200ms |
| 并发用户数 | ≥ 1000 |
| 单项目工作项 | ≥ 100 万 |
| 可用性 SLA | ≥ 99.9% |

## 贡献指南

### 分支策略（Trunk-Based）

```
main（受保护，永远可发布）
 ├── feature/xxx      → PR → squash merge
 ├── fix/xxx          → PR → squash merge
 └── release/v0.x     → 仅 cherry-pick 修复 → tag
```

### 提交规范

遵循 [Conventional Commits](https://www.conventionalcommits.org/)：

- `feat:` 新功能
- `fix:` 修复 Bug
- `docs:` 文档更新
- `refactor:` 重构
- `test:` 测试
- `chore:` 构建/工具变动

### 代码规范

- **Go**：`golangci-lint`（配置见 `.golangci.yml`）
- **TypeScript / Vue**：`eslint` + `prettier`
- **提交前**：`make lint && make test`

## 许可协议

Ydsz Plane 采用 [MIT License](LICENSE) 开源协议。

## 致谢

本项目的设计参考了以下优秀的开源项目与商业产品：

- [Plane](https://github.com/makeplane/plane) - 项目设计语言与信息架构参考
- [GitLab](https://gitlab.com/gitlab-org/gitlab) - DevOps 平台
- [云效](https://www.aliyun.com/product/yunxiao)
- [TAPD](https://www.tapd.cn)
- [ONES](https://ones.cn)

---

<p align="center">
  ⭐ 如果觉得这个项目对你有帮助，欢迎 <b>Star</b> 支持我们
</p>
