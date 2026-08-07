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
  Go · Vue 3 · PostgreSQL · Redis · RabbitMQ · Docker
</p>

---

## 项目状态

> **当前阶段：M0 ✅ + M0.5 ✅ + M1 ✅ + M2/M3/M4 ✅ + M5 ✅ — MVP v0.1 核心闭环（附件/评论/通知/WebSocket/导出/性能基线）全部就绪，待首轮 staging 压测回填基线后正式打版**
> 最后更新：2026-08-07 · 架构基线版本 v1.3

Ydsz Plane 是一款开源、自托管的现代项目管理工具，专为中小型敏捷开发团队设计。M0 工程基座、M0.5 增强基座（RBAC/可观测性/安全纵深/Swagger）、M1 租户与项目骨架（Workspace CRUD/成员邀请/审计）、M2 工作项核心（Issue 全生命周期/状态机/关联/依赖/WBS）、M3 迭代管理（Sprint 生命周期/燃尽图/速率统计）、M4 版本日（版本聚合/发布/交付报告）、M5 MVP 闭环（附件/评论/通知/WebSocket/导出）的后端 API 与前端视图已全部就绪。数据库迁移脚本从 0001 递进至 0017（users → workspaces → issue_core → state_templates → sprint_core → version_core → notifications → issue_comments → attachments → 等），覆盖全部核心域 schema。

## 已完成能力

截至 2026-08-08，仓库中已实现的能力：

| 模块 | 说明 | 状态 |
|------|------|------|
| 工程基座 | Monorepo（Go module + pnpm workspace）、CI（lint/test(race)/build/e2e-smoke）、Docker Compose 全栈、Makefile | ✅ |
| 鉴权链路 | 注册 / 登录（bcrypt + JWT access/refresh）、Cookie 会话、401 单飞刷新重放、忘记 / 重置密码端点 | ✅ |
| RBAC | Owner/Admin/Member/Guest 四角色 × workspace_members 域、10 项权限中间件、权限校验 | ✅ |
| API 文档 | swaggo 注解 + Swagger UI（`/swagger/index.html`）、Bearer 鉴权说明、30+ 端点 | ✅ |
| 邮件服务 | SMTP 抽象（Noop / SMTP 自动切换）、双版本 MIME 模板（密码重置 / 邀请） | ✅ |
| 可观测性 | zap 结构化日志、Prometheus RED 指标（`/metrics`）、Go runtime 默认收集 | ✅ |
| 安全纵深 | CSP / HSTS / COOP / CORP / Permissions-Policy / X-Frame-Options / X-Content-Type-Options 等 8 项安全头 | ✅ |
| 中间件链 | RequestID → Recovery → CORS → SecurityHeaders → AccessLog → Metrics → Auth | ✅ |
| Workspace API | CRUD（创建/读取/更新/归档/恢复）、Slug 唯一校验、项目集合管理 | ✅ |
| 成员与邀请 | 成员列表/角色切换/移除、邮箱邀请（token+7 天有效+可撤销）、邀请审核模式 | ✅ |
| 审计日志 | 空间级管理操作全量记录（invitation/workspace/member 操作）+ AuditService + 查询端点 | ✅ |
| 前端骨架 | Vue 3.5 + Vite 6 + Pinia、WorkspaceLayout、设计令牌（CSS 变量）、路由守卫 | ✅ |
| 前端视图 | 28 个页面（Login/Register/ForgotPassword/ResetPassword/Workspace/List/Settings/Project/List/Board/List/Settings/IssueDetail/Sprint/List/Planning/Detail/Standup/Version/List/Detail/Release/Report） | ✅ |
| 数据持久层 | pgx 连接池、租户上下文（SET LOCAL app.workspace_id）、RLS 策略模板、迁移工具 | ✅ |
| 事件骨架 | 事务型 Outbox 表 + Relay（DB → RabbitMQ EventExchange）、Asynq Worker | ✅ |
| 数据库迁移 | 0001~0017（users / workspaces / issue_core / state_templates / sprint_core / version_core / notifications / issue_comments / attachments / 等） | ✅ |
| 种子数据 | 5 用户 + 3 工作空间 + 多角色成员（owner/admin/member/guest）+ 幂等执行 | ✅ |
| Issue API | CRUD + 状态流转 + 活动日志 + 工时记录 + 关联（6 种关系）+ 依赖（FS/SS/FF/SF + lag_days）+ WBS 三级 + 类型差异化字段（defect/task/requirement）+ CSV/xlsx 导出 | ✅ |
| 协同增强 | 评论（富文本 + @提及 + 嵌套回复）、附件（MinIO 预签名上传/下载/预览）、通知（站内信 + 偏好 + 铃铛 + 列表页）、WebSocket（Redis Pub/Sub 扇出 + 断线补偿） | ✅ |
| 缺陷分析 | 聚合查询（严重程度/发现阶段/模块/根因/缺陷龄/周趋势）+ 明细导出（CSV / xlsx，支持版本过滤） | ✅ |
| CI/CD | GitHub Actions：后端 lint/test(race)/覆盖率门禁/govulncheck、前端 lint/typecheck/test/build/audit、CodeQL 安全分析、AI Code Review、按需 k6 性能压测 | ✅ |
| 测试治理 | Go 单元/集成测试（errs/middleware/workspace/attachment/search/notification 等）、前端 Vitest 组件测试 + Playwright E2E 冒烟 | ✅ |
| 版本发布 | CHANGELOG、发布管理文档、GitHub Release workflow（打 tag 自动构建产物） | ✅ |
| 依赖安全 | govulncheck + pnpm audit + CodeQL 三层防线，依赖安全治理文档 | ✅ |
| State API | 项目管理状态集 + 状态流转规则（state_transitions 按项目×类型维度配置） | ✅ |
| Sprint API | CRUD + 生命周期（start/complete）+ 燃尽图数据 + Backlog 查询 + 容量规划 + 速率建议 + 复盘快照 | ✅ |
| Version API | CRUD + 状态机（activate/release/archive）+ 进度聚合 + 交付报告 + Release Notes 生成 + 缺陷面板过滤 + 迭代聚合 | ✅ |

## 技术栈

| 层次 | 技术选型 | 说明 |
|------|----------|------|
| 后端 | Go 1.26.5 + Gin 1.12 | 模块化单体（DDD 轻量分层） |
| 前端 | Vue 3.5 + TypeScript + Vite 6 | 组合式 API、Pinia 状态管理 |
| 数据库 | PostgreSQL 18 | ACID + JSONB + RLS 租户隔离 + 信创方言预留 |
| 缓存 | Redis 8 | 限流、分布式锁、会话辅助、WebSocket 扇出 |
| 事件总线 | RabbitMQ 4 | Outbox 投递、可靠事件投递（替代 Redis Streams/NATS） |
| 任务队列 | RabbitMQ 4 | 异步任务（通知、索引、Webhook、自动化）延迟/优先级/死信/Retry-with-backoff |
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
              │  Infrastructure（PG / Redis / RabbitMQ / ES）   │
              └───────┬───────────────────────┬────────────────┘
                      │ 写事件 (Outbox)        │ 读
        ┌─────────────▼──────────┐   ┌────────▼───────┐
        │  ydsz-plane-worker     │   │ PostgreSQL 18  │
        │  (Asynq + Outbox Relay)│   │ Redis 8        │
        │  · 通知投递             │   │   · Asynq      │
        │  · ES 索引同步          │   │   · Cache      │
        │  · Webhook 分发         │   │   · RateLimit  │
        │  · 自动化规则执行        │   │ RabbitMQ 4     │
        │  · 迭代快照 / 效能计算   │   │   · Event Bus  │
        │                        │   │   · DLX / DLQ  │
        └────────────────────────┘   └────────────────┘
```

**双中间件分工**

| 中间件 | 用途 | 选型理由 |
|--------|------|----------|
| Redis 8 | 缓存、分布式锁、限流、WebSocket 扇出 | 低延迟、单二进制部署、复用现有运维基础 |
| RabbitMQ 4 | 任务队列 + 事件总线（Outbox → Exchange → Queue）、DLX/DLQ、延迟队列、Topic 路由 | 可靠投递（consumer acks + publisher confirms）、死信队列、灵活路由模式 |

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
│   │   ├── events/            # Outbox Relay（DB → Redis Streams）
│   │   └── telemetry/         # zap 结构化日志
│   ├── interfaces/
│   │   ├── http/              # Gin 路由 + Handler
│   │   └── middleware/        # 中间件链
│   └── config/                # Viper 配置加载 + 校验
├── pkg/
│   └── errs/                  # 统一错误类型 + 错误码注册
├── sql/                       # 递增编号迁移脚本（0001_init.up/down.sql）
├── scripts/seed/              # 开发环境种子数据
├── docs/
│   ├── architecture/          # 架构设计文档（14 份，含完整设计）
│   ├── deployments/           # Docker 部署配置（compose + Dockerfile + nginx）
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

# 启动核心服务（PostgreSQL + Redis + RabbitMQ + API + Worker + Web）
docker compose -f docs/deployments/docker-compose.yml up -d

# 启动完整栈（额外包含 Elasticsearch + MinIO）
docker compose -f docs/deployments/docker-compose.yml --profile full up -d

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
# 1. 启动基础设施容器 (PostgreSQL + Redis + RabbitMQ + Mailpit)
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
| RabbitMQ Mgmt | http://127.0.0.1:15672 | RabbitMQ Management UI (guest/guest) |
| Mailpit UI | http://127.0.0.1:8025 | 邮件测试 Web UI |

> 完整服务连接信息（Redis/MySQL/MinIO 等）请参考 [本地开发环境配置](docs/本地开发环境配置.md)

## Makefile 命令速查

| 命令 | 说明 |
|------|------|
| `make dev` | 启动基础设施容器 + 提示 dev-api / dev-web |
| `make up` | 启动基础设施（pg + redis + rabbitmq + mailpit） |
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
# 系统
GET  /healthz                              健康检查
GET  /readyz                               就绪检查（含 PG / Redis 探针）
GET  /metrics                              Prometheus RED 指标
GET  /swagger/index.html                   Swagger UI（dev 模式）

# 鉴权
POST /api/v1/auth/login                    邮箱 + 密码登录（返回 access/refresh + cookie）
POST /api/v1/auth/refresh                  通过 refresh cookie 换发令牌对
POST /api/v1/auth/register                 新用户注册（需 features.registration_open=true）
POST /api/v1/auth/forgot-password          触发密码重置邮件（202，防枚举）
POST /api/v1/auth/reset-password           用一次性 token 重置密码
GET  /api/v1/me                            Bearer / Cookie 获取当前用户简介

# 工作空间
GET  /api/v1/workspaces                    列出当前用户的工作空间
POST /api/v1/workspaces                    创建工作空间
GET  /api/v1/workspaces/slug/:slug         通过 slug 查 ID

# 工作空间作用域
GET  /api/v1/workspaces/:id                工作空间详情
PATCH /api/v1/workspaces/:id               更新工作空间
DELETE /api/v1/workspaces/:id              归档工作空间
GET  /api/v1/workspaces/:id/members        成员列表
PATCH /api/v1/workspaces/:id/members/:uid  修改成员角色
DELETE /api/v1/workspaces/:id/members/:uid 移除成员
POST /api/v1/workspaces/:id/invitations    发送邀请
GET  /api/v1/workspaces/:id/invitations    邀请列表
DELETE /api/v1/workspaces/:id/invitations/:iid  撤销邀请
GET  /api/v1/workspaces/:id/audit-logs     审计日志

# 项目
GET  /api/v1/workspaces/:id/projects       项目列表
POST /api/v1/workspaces/:id/projects       创建项目
GET  /api/v1/workspaces/:id/projects/:pid  项目详情
PATCH /api/v1/workspaces/:id/projects/:pid 更新项目
DELETE /api/v1/workspaces/:id/projects/:pid 归档项目

# 工作项
GET  /api/v1/workspaces/:id/projects/:pid/states        状态列表
GET  /api/v1/workspaces/:id/projects/:pid/issues        工作项列表（过滤/排序/搜索/分页）
POST /api/v1/workspaces/:id/projects/:pid/issues        创建工作项
GET  /api/v1/workspaces/:id/projects/:pid/issues/:iid   工作项详情
PATCH /api/v1/workspaces/:id/projects/:pid/issues/:iid   更新工作项
DELETE /api/v1/workspaces/:id/projects/:pid/issues/:iid   删除工作项
POST /api/v1/workspaces/:id/projects/:pid/issues/:iid/transition  状态流转
GET  /api/v1/workspaces/:id/projects/:pid/issues/:iid/activities  活动日志
GET  /api/v1/workspaces/:id/projects/:pid/issues/:iid/time-logs   工时记录
POST /api/v1/workspaces/:id/projects/:pid/issues/:iid/time-logs   记录工时
GET  /api/v1/workspaces/:id/projects/:pid/issues/:iid/relations   关联列表
POST /api/v1/workspaces/:id/projects/:pid/issues/:iid/relations   创建关联
DELETE /api/v1/workspaces/:id/projects/:pid/issues/:iid/relations/:rid  删除关联
GET  /api/v1/workspaces/:id/projects/:pid/issues/:iid/dependencies    依赖列表
POST /api/v1/workspaces/:id/projects/:pid/issues/:iid/dependencies    创建依赖
DELETE /api/v1/workspaces/:id/projects/:pid/issues/:iid/dependencies/:did  删除依赖

# 迭代
GET  /api/v1/workspaces/:id/projects/:pid/sprints        迭代列表
POST /api/v1/workspaces/:id/projects/:pid/sprints        创建迭代
GET  /api/v1/workspaces/:id/projects/:pid/sprints/backlog   Backlog
GET  /api/v1/workspaces/:id/projects/:pid/sprints/suggest-capacity  速率建议
GET  /api/v1/workspaces/:id/projects/:pid/sprints/:sid   迭代详情
PATCH /api/v1/workspaces/:id/projects/:pid/sprints/:sid   更新迭代
DELETE /api/v1/workspaces/:id/projects/:pid/sprints/:sid   删除迭代
POST /api/v1/workspaces/:id/projects/:pid/sprints/:sid:start     启动迭代
POST /api/v1/workspaces/:id/projects/:pid/sprints/:sid:complete  结束迭代
GET  /api/v1/workspaces/:id/projects/:pid/sprints/:sid/progress   迭代进度
GET  /api/v1/workspaces/:id/projects/:pid/sprints/:sid/issues    迭代工作项
POST /api/v1/workspaces/:id/projects/:pid/sprints/:sid/issues    加入迭代
DELETE /api/v1/workspaces/:id/projects/:pid/sprints/:sid/issues/:iid  移出迭代
GET  /api/v1/workspaces/:id/projects/:pid/sprints/:sid/burndown  燃尽图
GET  /api/v1/workspaces/:id/projects/:pid/sprints/:sid/review    复盘数据

# 版本日
GET  /api/v1/workspaces/:id/projects/:pid/versions          版本列表
POST /api/v1/workspaces/:id/projects/:pid/versions          创建版本
GET  /api/v1/workspaces/:id/projects/:pid/versions/defects  缺陷跨版本过滤
GET  /api/v1/workspaces/:id/projects/:pid/versions/:vid     版本详情
PATCH /api/v1/workspaces/:id/projects/:pid/versions/:vid     更新版本
DELETE /api/v1/workspaces/:id/projects/:pid/versions/:vid     删除版本
POST /api/v1/workspaces/:id/projects/:pid/versions/:vid/activate  激活版本
POST /api/v1/workspaces/:id/projects/:pid/versions/:vid/release   发布版本
POST /api/v1/workspaces/:id/projects/:pid/versions/:vid/archive   归档版本
GET  /api/v1/workspaces/:id/projects/:pid/versions/:vid/progress  版本进度
GET  /api/v1/workspaces/:id/projects/:pid/versions/:vid/quality   质量指标
GET  /api/v1/workspaces/:id/projects/:pid/versions/:vid/delivery-report  交付报告
GET  /api/v1/workspaces/:id/projects/:pid/versions/:vid/release-notes     Release Notes
POST /api/v1/workspaces/:id/projects/:pid/versions/:vid/release-notes/regenerate  重生成 Notes
GET  /api/v1/workspaces/:id/projects/:pid/versions/:vid/defects   缺陷面板
GET  /api/v1/workspaces/:id/projects/:pid/versions/:vid/sprints   聚合迭代列表
POST /api/v1/workspaces/:id/projects/:pid/versions/:vid/sprints   添加迭代
DELETE /api/v1/workspaces/:id/projects/:pid/versions/:vid/sprints/:sid  移除迭代

# 邀请
POST /api/v1/invitations/accept            接受邀请
GET  /api/v1/invitations/:token            邀请预览（公开）
```

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
| 15 | [发布管理](docs/architecture/15-发布管理.md) | 版本规范、分支保护、发布门禁 |
| 16 | [依赖安全治理](docs/architecture/16-依赖安全治理.md) | govulncheck、npm audit、漏洞响应 |
| 17 | [性能基线与压测](docs/architecture/17-性能基线.md) | 性能目标、k6 脚本、基线回归 |

## 路线图

### M0 工程基座 ✅ 已完成（S1，W2）

- [x] Monorepo 初始化（Go module + pnpm workspace）
- [x] CI/CD（GitHub Actions: lint / test(race) / build / e2e-smoke）
- [x] Docker Compose 全栈 + Makefile
- [x] Gin 骨架 + 中间件链 + 健康检查
- [x] 数据库迁移系统 + RLS 模板 + 租户上下文
- [x] Outbox + Asynq + Redis Streams 事件骨架
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

### M1 租户与项目骨架（S2 ✅ 已完成）

#### 后端 API 已交付 ✅

- [x] users / workspaces / workspace_members / password_reset_tokens / audit_logs / invitations / audit_logs
- [x] 鉴权链路（注册/登录/刷新/RBAC）
- [x] 邮件抽象 + 模板
- [x] **Workspace CRUD API**（创建/读取/更新/归档/恢复）
- [x] **成员邀请**（invitations 表 + 邮件链接 + 7 天有效 + 审核模式）
- [x] **审计日志** 埋点（空间级管理操作）+ AuditService + 查询端点
- [x] **项目 CRUD API**（创建/读取/更新/归档 + Identifier 生成 + RLS 策略）

#### 前端已交付 ✅

- [x] 空间列表 / 创建向导 / 设置页（信息/成员/邀请）
- [ ] API Token 管理页（创建/吊销/scopes） ⏳ P2
- [x] WorkspaceLayout + 项目导航 + 路由守卫
- [x] 邀请接受页（公开预览 + 登录后签收）

### M2 工作项核心（S3–S4 ✅ 后端 API 已交付，前端视图就绪）

- [x] Issue 主表 + 状态机配置 + 三级 WBS + sequence_id 发号器
- [x] 需求 / 任务 / 缺陷全流程（类型差异化字段 + 状态流转 + optimistic lock）
- [x] 看板 / 列表视图前端组件（KanbanBoardView + IssueListView + IssueFilter）
- [x] 活动日志时间线（ActivityService + issue_activities 表）
- [x] 工时记录（TimeLogService + time_logs 表）
- [x] 关联 6 种关系 + 依赖 FS/SS/FF/SF + lag_days（RelationService）
- [x] 缺陷 P0 字段（severity/found_phase/root_cause/environment/reproduce_steps）
- [x] 状态流转规则（state_transitions 按 project×type 维度配置）
- [x] 工作项详情页 + 关联面板 + 缺陷面板前端

### M3 迭代管理（S5 ✅ 后端 API 已交付，前端视图就绪）

- [x] Sprint 生命周期（start/complete + 唯一 active 约束）
- [x] 燃尽图数据接口（sprint_snapshots 表 + burndown 端点）
- [x] Backlog 查询 + 容量规划 + 速率建议（VelocityStats）
- [x] 迭代规划视图（Backlog↔Sprint 拖拽）+ 迭代详情 + 站会模式
- [x] 复盘数据（ReviewSnapshot + 中途加项标记 added_midway）
- [x] 前端视图：SprintList / SprintDetail / SprintPlanning / SprintStandup / BurndownChart

### M4 版本日（S6 ✅ 后端 API 已交付，前端视图就绪）

- [x] Version CRUD + 状态机（planning → active → released → archived）
- [x] Semver 校验 + Release Notes 模板渲染 + 交付报告数据接口
- [x] 进度聚合（跨迭代完成度汇总 + progress_snapshot）
- [x] 缺陷发现/修复版本联动过滤 + 缺陷面板
- [x] 发布检查清单（checklist JSON）+ 准出校验
- [x] 前端视图：VersionList / VersionDetail / VersionRelease / DeliveryReport / DefectPanel

### M5 MVP v0.1 发布（S7 ✅ 已收尾）

- [x] MinIO 附件上传/下载 + 图片粘贴
- [x] 评论 issue_comments（富文本 + @提及 + 嵌套回复）
- [x] 通知系统（notifications + 站内信 + 默认规则 + 偏好设置）
- [x] WebSocket 实时推送（Redis Pub/Sub 多节点扇出）
- [x] MVP 加固：错误页/空态/加载态全链路统一组件
- [x] 缺陷导出（CSV / xlsx）+ 缺陷分析聚合与明细导出
- [x] 性能基线：造数脚本（100 万级）+ k6 三件套 + CI 压测流水线

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
