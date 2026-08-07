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

## 技术栈

| 层次 | 技术选型 | 说明 |
|------|----------|------|
| 后端 | Go 1.26.5 + Gin 1.12 | 模块化单体（DDD 轻量分层） |
| 前端 | Vue 3.5 + TypeScript + Vite 6 | 组合式 API、Pinia 状态管理 |
| 数据库 | PostgreSQL 18 | ACID + JSONB + RLS 租户隔离 + 全文检索（tsvector） + 信创方言预留 |
| 缓存 | Redis 8 | 限流、分布式锁、会话辅助、WebSocket 扇出 |
| 事件总线 | RabbitMQ 4 | Outbox 投递、可靠事件投递（替代 Redis Streams/NATS） |
| 任务队列 | RabbitMQ 4 | 异步任务（通知、索引、Webhook、自动化）延迟/优先级/死信/Retry-with-backoff |
| 全文检索 | PostgreSQL FTS（当前） / Elasticsearch 8（可选 profile） | 全局搜索、分词（IK） |
| 对象存储 | MinIO（可选 profile） | 附件、Logo |
| 部署 | Docker Compose（一键）/ K8s（Phase 3） | 信创兼容：openEuler/麒麟 + ARM64 |

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
│   │   ├── apitoken/          # API Token 管理与认证
│   │   ├── attachment/        # 附件上传/下载/预览（MinIO 预签名）
│   │   ├── auth/              # 登录 / 令牌签发 / RBAC / 密码重置
│   │   ├── automation/        # 自动化规则引擎（DSL / 执行器 / 模板）
│   │   ├── dashboard/         # 项目仪表盘（Widget / 风险告警 / 模板）
│   │   ├── intake/            # 收件箱（提交通道 / 工单审核 / 公开表单）
│   │   ├── issue/             # 工作项核心（CRUD / 状态机 / 关联 / 依赖 / 缺陷分析）
│   │   ├── metrics/           # 效能度量（DORA / 速度 / 前置时间 / 质量）
│   │   ├── notification/      # 通知系统（站内信 / 偏好 / 多渠道分发）
│   │   ├── preference/        # 视图偏好持久化
│   │   ├── search/            # 全文搜索（PG FTS / 多对象检索 / 书签）
│   │   ├── sprint/            # 迭代管理（生命周期 / 燃尽图 / 容量 / 复盘）
│   │   ├── version/           # 版本日（CRUD / 状态机 / 交付报告 / Release Notes）
│   │   ├── webhook/           # Webhook 管理（订阅 / 投递 / 签名 / 重试）
│   │   ├── workbench/         # 个人工作台（任务分桶 / 概览 / 模板）
│   │   └── workspace/         # 工作空间（CRUD / 成员 / 邀请 / 审计 / 项目）
│   ├── infrastructure/
│   │   ├── cache/             # Redis 客户端
│   │   ├── events/            # Outbox Relay（DB → RabbitMQ）
│   │   ├── mail/              # SMTP 邮件服务 + 模板
│   │   ├── mq/                # RabbitMQ 连接 + 任务队列
│   │   ├── persistence/       # pgx 连接池 + 租户上下文
│   │   ├── storage/           # MinIO/S3 对象存储客户端
│   │   ├── telemetry/         # zap 结构化日志 + Prometheus 指标
│   │   └── ws/                # WebSocket Hub（多节点扇出）
│   ├── interfaces/
│   │   ├── http/              # Gin 路由 + Handler + DTO + 导出基础设施
│   │   └── middleware/        # 中间件链（RequestID/Recovery/CORS/Security/Auth/RBAC/Metrics）
│   └── config/                # Viper 配置加载 + 校验
├── pkg/
│   └── errs/                  # 统一错误类型 + 错误码注册
├── sql/                       # 递增编号迁移脚本（0001~0017）
├── scripts/
│   ├── seed/                  # 开发环境种子数据
│   ├── seed-scale/            # 大规模造数（100 万工作项）
│   └── k6/                    # k6 压测脚本（smoke / load / stress）
├── tests/                     # 后端集成测试 + 性能压测（k6）
├── docs/
│   ├── architecture/          # 架构设计文档（17 份，含完整设计）
│   ├── deployments/           # Docker 部署配置（compose + Dockerfile + nginx）
│   ├── perf/                  # 性能基线报告归档
│   └── Ydsz Plane PRD-终极完整版.docx
├── web/                       # 前端（pnpm workspace）
│   ├── src/
│   │   ├── api/               # Axios 客户端（含 401 自动刷新）+ Service 层
│   │   ├── views/             # 31 个页面组件
│   │   ├── stores/            # Pinia 状态管理（auth / workspace / sprint / ...）
│   │   ├── router/            # 路由 + 权限守卫
│   │   ├── layouts/           # WorkspaceLayout
│   │   ├── components/        # 通用组件 + 仪表盘 Widget 组件
│   │   ├── design/            # 设计令牌（CSS 变量）
│   │   └── lib/               # 工具库（toast / ws-client / ...）
│   └── packages/
│       ├── design-tokens/     # 主题令牌包
│       └── ui/                # 基础组件门面（建设中）
├── .github/workflows/         # ci.yml / codeql.yml / perf.yml / release.yml
├── Makefile                   # dev / up / migrate / seed / seed-scale / lint / test / build / openapi
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
| `make seed-scale` | 大规模造数（100 万工作项，8 并发） |
| `make dev-api` | 启动 API（air 热重载） |
| `make dev-worker` | 启动 Worker |
| `make dev-web` | 启动前端 dev server |
| `make lint` | Go + 前端全量 lint |
| `make test` | Go（race）+ 前端全量测试 |
| `make build` | Go + 前端全量构建 |
| `make openapi` | 生成 OpenAPI 规范（swaggo） |

## API 设计概览

- 风格：REST + 统一 envelope；版本前缀 `/api/v1`
- 认证：Cookie（Web SPA）/ Bearer Token（API 调用）/ API Token（第三方集成）
- 错误格式：`{"error":{"code":"AUTH.INVALID_CREDENTIALS","message":"...","request_id":"..."}}`
- 中间件：RequestID → Recovery → CORS → SecurityHeaders → AccessLog → Metrics → Auth

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
GET  /api/v1/me                            Bearer / Cookie / API Token 获取当前用户简介

# API Token（用户级）
GET  /api/v1/me/api-tokens                 列出我的 Token
POST /api/v1/me/api-tokens                 创建 Token
DELETE /api/v1/me/api-tokens/:token_id     吊销 Token

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
PATCH /api/v1/workspaces/:id/projects/:pid  更新项目
DELETE /api/v1/workspaces/:id/projects/:pid  归档项目

# 工作项
GET  /api/v1/workspaces/:id/projects/:pid/states        状态列表
GET  /api/v1/workspaces/:id/projects/:pid/issues        工作项列表（过滤/排序/搜索/分页）
POST /api/v1/workspaces/:id/projects/:pid/issues        创建工作项
POST /api/v1/workspaces/:id/projects/:pid/issues/batch  批量创建工作项
GET  /api/v1/workspaces/:id/projects/:pid/issues/export 导出（CSV / xlsx）
GET  /api/v1/workspaces/:id/projects/:pid/issues/:iid   工作项详情
PATCH /api/v1/workspaces/:id/projects/:pid/issues/:iid   更新工作项
DELETE /api/v1/workspaces/:id/projects/:pid/issues/:iid   删除工作项
POST /api/v1/workspaces/:id/projects/:pid/issues/:iid/transition  状态流转
GET  /api/v1/workspaces/:id/projects/:pid/issues/:iid/activities  活动日志
GET  /api/v1/workspaces/:id/projects/:pid/issues/:iid/time-logs   工时记录
POST /api/v1/workspaces/:id/projects/:pid/issues/:iid/time-logs   记录工时
PATCH /api/v1/workspaces/:id/projects/:pid/issues/:iid/time-logs/:log_id  更新工时
DELETE /api/v1/workspaces/:id/projects/:pid/issues/:iid/time-logs/:log_id  删除工时
GET  /api/v1/workspaces/:id/projects/:pid/issues/:iid/relations   关联列表
POST /api/v1/workspaces/:id/projects/:pid/issues/:iid/relations   创建关联
DELETE /api/v1/workspaces/:id/projects/:pid/issues/:iid/relations/:rid  删除关联
GET  /api/v1/workspaces/:id/projects/:pid/issues/:iid/dependencies    依赖列表
POST /api/v1/workspaces/:id/projects/:pid/issues/:iid/dependencies    创建依赖
DELETE /api/v1/workspaces/:id/projects/:pid/issues/:iid/dependencies/:did  删除依赖
PATCH /api/v1/workspaces/:id/projects/:pid/issues/:iid/reorder        看板排序
GET  /api/v1/workspaces/:id/projects/:pid/issues/:iid/comments       评论列表
POST /api/v1/workspaces/:id/projects/:pid/issues/:iid/comments       创建评论
PATCH /api/v1/workspaces/:id/projects/:pid/issues/:iid/comments/:cid 更新评论
DELETE /api/v1/workspaces/:id/projects/:pid/issues/:iid/comments/:cid 删除评论

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

# 搜索（工作空间级 + 项目级）
GET  /api/v1/workspaces/:id/search                 跨项目全文搜索
GET  /api/v1/workspaces/:id/search/history         搜索历史
DELETE /api/v1/workspaces/:id/search/history       清空搜索历史
DELETE /api/v1/workspaces/:id/search/history/:id   删除单条历史
GET  /api/v1/workspaces/:id/search/bookmarks       搜索书签列表
POST /api/v1/workspaces/:id/search/bookmarks       创建搜索书签
PATCH /api/v1/workspaces/:id/search/bookmarks/:id  更新搜索书签
DELETE /api/v1/workspaces/:id/search/bookmarks/:id  删除搜索书签
GET  /api/v1/workspaces/:id/projects/:pid/search   项目级搜索
GET  /api/v1/workspaces/:id/projects/:pid/search/history
GET  /api/v1/workspaces/:id/projects/:pid/search/bookmarks

# 工作台（工作空间级 + 项目级）
GET  /api/v1/workspaces/:id/workbench/summary      个人工作台首屏聚合
GET  /api/v1/workspaces/:id/workbench/config       工作台配置
PUT  /api/v1/workspaces/:id/workbench/config       保存配置
GET  /api/v1/workspaces/:id/workbench/recent       最近访问
POST /api/v1/workspaces/:id/workbench/recent       记录访问
GET  /api/v1/workspaces/:id/workbench/templates    工作台模板
POST /api/v1/workspaces/:id/workbench/templates/apply  应用模板
GET  /api/v1/workspaces/:id/projects/:pid/workbench/summary
GET  /api/v1/workspaces/:id/projects/:pid/workbench/config

# 仪表盘（项目级 + 工作空间级）
GET  /api/v1/workspaces/:id/projects/:pid/dashboard          仪表盘数据
GET  /api/v1/workspaces/:id/projects/:pid/dashboard/widgets   Widget 列表
POST /api/v1/workspaces/:id/projects/:pid/dashboard/widgets   创建 Widget
PATCH /api/v1/workspaces/:id/projects/:pid/dashboard/widgets/:id  更新 Widget
DELETE /api/v1/workspaces/:id/projects/:pid/dashboard/widgets/:id  删除 Widget
GET  /api/v1/workspaces/:id/projects/:pid/dashboard/alerts   风险告警
POST /api/v1/workspaces/:id/projects/:pid/dashboard/alerts/:id/resolve  解决告警
GET  /api/v1/workspaces/:id/projects/:pid/dashboard/templates  仪表盘模板
GET  /api/v1/workspaces/:id/dashboard/alerts                 工作空间级告警

# 效能度量
GET  /api/v1/workspaces/:id/projects/:pid/metrics/velocity       迭代速率
GET  /api/v1/workspaces/:id/projects/:pid/metrics/velocity/trend 速率趋势
GET  /api/v1/workspaces/:id/projects/:pid/metrics/lead-time      前置时间分布
GET  /api/v1/workspaces/:id/projects/:pid/metrics/quality        质量指标
GET  /api/v1/workspaces/:id/projects/:pid/metrics/dora           DORA 四指标
GET  /api/v1/workspaces/:id/projects/:pid/metrics/resource-load  资源负载
POST /api/v1/workspaces/:id/projects/:pid/metrics/deployments    上报部署事件
GET  /api/v1/workspaces/:id/projects/:pid/metrics/snapshots      每日快照

# Webhook 管理
GET  /api/v1/workspaces/:id/webhooks                 Webhook 列表
POST /api/v1/workspaces/:id/webhooks                 创建 Webhook
GET  /api/v1/workspaces/:id/webhooks/:wid           详情
PATCH /api/v1/workspaces/:id/webhooks/:wid          更新
DELETE /api/v1/workspaces/:id/webhooks/:wid         删除
GET  /api/v1/workspaces/:id/webhooks/:wid/logs      投递日志
POST /api/v1/workspaces/:id/webhooks/:wid/test      测试投递（Ping）
POST /api/v1/workspaces/:id/webhooks/:wid/logs/:lid/retry  手动重试

# Intake 收件箱（管理端）
GET  /api/v1/workspaces/:id/intake/channels         通道列表
POST /api/v1/workspaces/:id/intake/channels         创建通道
GET  /api/v1/workspaces/:id/intake/channels/:cid   通道详情
PATCH /api/v1/workspaces/:id/intake/channels/:cid  更新通道
DELETE /api/v1/workspaces/:id/intake/channels/:cid  删除通道
GET  /api/v1/workspaces/:id/intake/issues           工单列表
GET  /api/v1/workspaces/:id/intake/issues/:iid      工单详情
POST /api/v1/workspaces/:id/intake/issues/:iid/review  审核工单
POST /api/v1/workspaces/:id/intake/issues/:iid/convert  转正为正式工作项

# Intake 公开 API（免登录）
GET  /api/v1/public/intake/channels/:workspace/:slug      获取通道配置（渲染表单）
POST /api/v1/public/intake/channels/:workspace/:slug/submit  公开提交工单
GET  /api/v1/public/intake/track                           提交者跟踪查询

# 自动化规则（项目级）
GET  /api/v1/workspaces/:id/projects/:pid/automation             规则列表
POST /api/v1/workspaces/:id/projects/:pid/automation             创建规则
GET  /api/v1/workspaces/:id/projects/:pid/automation/templates   内置模板列表
POST /api/v1/workspaces/:id/projects/:pid/automation/dry-run     试运行规则
POST /api/v1/workspaces/:id/projects/:pid/automation/from-template  从模板创建
GET  /api/v1/workspaces/:id/projects/:pid/automation/executions  执行审计日志
GET  /api/v1/workspaces/:id/projects/:pid/automation/:rid        规则详情
PATCH /api/v1/workspaces/:id/projects/:pid/automation/:rid        更新规则
DELETE /api/v1/workspaces/:id/projects/:pid/automation/:rid        删除规则
POST /api/v1/workspaces/:id/projects/:pid/automation/:rid/toggle   启用/禁用

# 视图偏好
GET  /api/v1/workspaces/:id/projects/:pid/preferences      获取偏好
PUT  /api/v1/workspaces/:id/projects/:pid/preferences      保存偏好

# 缺陷分析
GET  /api/v1/workspaces/:id/projects/:pid/analytics/defects       聚合查询
GET  /api/v1/workspaces/:id/projects/:pid/analytics/defects/export  明细导出

# 通知
GET  /api/v1/workspaces/:id/notifications                通知列表
POST /api/v1/workspaces/:id/notifications/read-all       全部已读
GET  /api/v1/workspaces/:id/notifications/unread-count   未读计数
GET  /api/v1/workspaces/:id/notifications/preferences    通知偏好
PUT  /api/v1/workspaces/:id/notifications/preferences    保存偏好

# 附件
POST /api/v1/workspaces/:id/projects/:pid/attachments         获取预签名上传 URL
GET  /api/v1/workspaces/:id/projects/:pid/attachments/:aid    获取预签名下载 URL
GET  /api/v1/workspaces/:id/projects/:pid/attachments/:aid/preview  获取预览 URL

# WebSocket
GET  /ws/:workspace_id           实时推送（Redis Pub/Sub 扇出 + 断线补偿）

# 邀请
POST /api/v1/invitations/accept            接受邀请
GET  /api/v1/invitations/:token            邀请预览（公开）
```

## 架构文档

详细设计见 [`docs/architecture/`](docs/architecture/README.md)，共 17 份文档覆盖从工程基座到性能基线的全链路设计：

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
- [x] swaggo 完整 Swagger 注解 + Swagger UI + Bearer 鉴权 + API Token 鉴权
- [x] 邮件 SMTP 抽象（Noop/SMTP 自动切换）+ 双 MIME 模板（重置/邀请）
- [x] JWT Secret dev 随机生成 + production fail-fast
- [x] 密码重置 token 表（0002 迁移）+ 端点接入
- [x] Axios 拦截器（401 单飞刷新 + 限流 429 回调 + Request ID）
- [x] 统一 authApi Service 层 + auth Store

### M1 租户与项目骨架（S2 ✅ 已完成）

#### 后端 API 已交付 ✅

- [x] users / workspaces / workspace_members / password_reset_tokens / audit_logs / invitations / audit_logs
- [x] 鉴权链路（注册/登录/刷新/RBAC）+ API Token 管理
- [x] 邮件抽象 + 模板
- [x] **Workspace CRUD API**（创建/读取/更新/归档/恢复）
- [x] **成员邀请**（invitations 表 + 邮件链接 + 7 天有效 + 审核模式）
- [x] **审计日志** 埋点（空间级管理操作）+ AuditService + 查询端点
- [x] **项目 CRUD API**（创建/读取/更新/归档 + Identifier 生成 + RLS 策略）

#### 前端已交付 ✅

- [x] 空间列表 / 创建向导 / 设置页（信息/成员/邀请）
- [x] WorkspaceLayout + 项目导航 + 路由守卫
- [x] 邀请接受页（公开预览 + 登录后签收）

### M2 工作项核心（S3–S4 ✅ 后端 API 已交付，前端视图就绪）

- [x] Issue 主表 + 状态机配置 + 三级 WBS + sequence_id 发号器
- [x] 需求 / 任务 / 缺陷全流程（类型差异化字段 + 状态流转 + optimistic lock）
- [x] 看板 / 列表视图前端组件（KanbanBoardView + IssueListView + IssueFilter）
- [x] 活动日志时间线（ActivityService + issue_activities 表）
- [x] 工时记录 CRUD（TimeLogService + time_logs 表）
- [x] 关联 6 种关系 + 依赖 FS/SS/FF/SF + lag_days（RelationService）
- [x] 缺陷 P0 字段（severity/found_phase/root_cause/environment/reproduce_steps）
- [x] 状态流转规则（state_transitions 按 project×type 维度配置）
- [x] 工作项详情页 + 关联面板 + 缺陷面板前端
- [x] 评论 CRUD（updateComment / deleteComment）
- [x] 批量操作（batchIssues）+ 看板排序（reorderIssue）

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

### M5 MVP v0.1 发布（S7 ✅ 已收尾 — 2026-08-07）

- [x] MinIO 附件上传/下载 + 图片粘贴
- [x] 评论 issue_comments（富文本 + @提及 + 嵌套回复）
- [x] 通知系统（notifications + 站内信 + 默认规则矩阵 + 偏好设置 + 铃铛面板 + 列表页）
- [x] WebSocket 实时推送（Redis Pub/Sub 多节点扇出 + 断线补偿）
- [x] MVP 加固：错误页/空态/加载态全链路统一组件
- [x] 缺陷导出（CSV / xlsx）+ 缺陷分析聚合与明细导出
- [x] 性能基线：造数脚本（100 万级）+ k6 三件套 + CI 压测流水线
- [x] 通用导出基础设施（CSV UTF-8 BOM / 最小合法 OOXML xlsx）
- [x] 通知分发 Worker 支撑（next_retry_at + 部分索引）

### M6 开放与智能（S8–S11 ✅ 后端 API 已交付，前端视图就绪）

#### S8 搜索（✅ 已完成）

- [x] PostgreSQL FTS（search_tsv tsvector + GIN 索引 + 触发器自动同步）
- [x] 多对象检索（issues / sprints / versions 通过 search_documents 表）
- [x] 搜索历史（最近 50 条 / user + 快速回显）
- [x] 搜索书签（保存的过滤器 + 共享）
- [x] 关键词高亮（mark 标签）
- [x] 前端视图：SearchView（左过滤器 + 中间结果 + 分组 + 历史/书签）

#### S9 工作台 / 仪表盘 / 视图偏好（✅ 已完成）

- [x] 个人工作台首屏（WorkbenchSummary：任务分桶 / 迭代概览 / 最近访问 / 快捷操作）
- [x] 工作台配置持久化 + 模板应用
- [x] 项目仪表盘 Widget 框架（10 种 Widget 类型）
- [x] 仪表盘数据快照（dashboard_snapshots 加速首屏）
- [x] 风险预警规则与告警（6 种 rule_type + 自动触发 + 解决流程）
- [x] 仪表盘预设模板（项目概览 / 项目管理 / 质量看板）
- [x] 视图偏好持久化（view_preferences：看板/列表布局、列配置、过滤条件、排序）
- [x] 前端视图：WorkbenchView + DashboardView + 通知偏好设置

#### S10 Webhook / Intake（✅ 已完成）

- [x] Webhook 订阅管理（CRUD + HMAC-SHA256 签名密钥）
- [x] Webhook 投递日志（30 天保留 + 响应详情 + duration）
- [x] Webhook 测试（Ping）+ 手动重试 + unhealthy 自动标记
- [x] Intake 收件通道 CRUD（公开表单 + 限流 + 验证码 + 自定义字段）
- [x] Intake 工单审核（tracking 回执 + 自动分配 + 审核转正）
- [x] 公开提交 API（免登录 + 提交者跟踪查询）

#### S11 自动化规则 / 效能度量（✅ 已完成）

- [x] 自动化规则引擎（Trigger-Condition-Action DSL, JSONB 存储）
- [x] 规则状态机（draft/active/disabled/error + 连续失败自动降级）
- [x] 7 条内置规则模板（子项完成自动流转 / 逾期提醒 / 版本发布流转 / Epic 点数汇总 / 自动开始日期 / 最闲人指派 / 新缺陷通知）
- [x] 规则执行审计（rule_executions + 防循环 + dry-run）
- [x] 效能度量（DORA 四指标 / 迭代速率趋势 / 前置时间分布 / 质量指标 / 资源负载 WIP）
- [x] 部署事件上报（deployment_events + 幂等）
- [x] 每日快照（metric_snapshots + 管理员数据校准）

### Phase 3+（S12+）

- [ ] 通知多渠道（IM 企微/钉钉/飞书/邮件摘要）
- [ ] 全局搜索 ES 升级（IK 分词 + 类 JQL 语法解析器）
- [ ] 甘特图 / 日历 / 电子表格视图
- [ ] SSO / SAML 集成
- [ ] 国际化、PWA
- [ ] 数据迁移工具（Jira / 云效 / ONES 导入）
- [ ] AI 功能（智能分配、重复检测）
- [ ] 信创数据库实测（达梦 / 人大金仓 + 国密算法）

## 竞品对标

| 能力域 | 云效 | TAPD | ONES | PingCode | Jira | Ydsz Plane |
|--------|------|------|------|----------|------|------------|
| 需求管理 | ● | ● | ● | ● | ● | ● |
| 任务管理 | ● | ● | ● | ● | ● | ● |
| 缺陷管理 | ● | ● | ● | ● | ● | ● |
| 迭代管理 | ● | ● | ● | ● | ● | ● |
| 版本管理 | ● | ○ | ● | ● | ● | ● |
| 项目仪表盘 | ● | ● | ● | ● | ● | ● |
| 全文搜索 | ● | ● | ● | ● | ● | ● |
| 工作台 | ● | ● | ● | ● | ● | ● |
| 自动化 | ● | ○ | ● | ● | ● | ● |
| Webhook | ● | ○ | ● | ● | ● | ● |
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

- [Plane](https://github.com/makeplane/plane) 
- [GitLab](https://gitlab.com/gitlab-org/gitlab) 
- [云效](https://www.aliyun.com/product/yunxiao)
- [TAPD](https://www.tapd.cn)
- [ONES](https://ones.cn)

---
