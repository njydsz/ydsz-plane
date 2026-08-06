# Ydsz Plane

<p align="center">
  <b>面向中国软件团队的开源项目管理平台</b>
  <br />
  对标云效 · TAPD · ONES · PingCode，本土化研发全流程
</p>

<p align="center">
  <a href="#快速开始"><b>快速开始</b></a> ·
  <a href="#在线体验"><b>在线体验</b></a> ·
  <a href="https://github.com/ydszopen/ydsz-plane/wiki"><b>文档</b></a> ·
  <a href="#路线图"><b>路线图</b></a> ·
  <a href="#贡献指南"><b>参与贡献</b></a>
</p>

<p align="center">
  技术栈：Go · Vue 3 · PostgreSQL · Redis · Docker
</p>

---

## 简介

**Ydsz Plane** 是一款开源、自托管的现代项目管理工具，专为中小型敏捷开发团队设计。融合了云效、TAPD、ONES 在需求管理、任务管理、迭代协同方面的本土化实践，补齐项目仪表盘、个人工作台、研发效能度量等能力，致力于成为符合国内研发现状的「开源 PM SaaS 底座」。

## 在线体验

> 公共 Demo 部署中... 敬请期待。

## 核心功能

| 模块 | 说明 | 状态 |
|------|------|------|
| 工作空间 | 组织级租户隔离，成员邀请、角色权限、SSO 集成 | ✅ 已完成 |
| 项目管理 | 项目 CRUD、模板复用、模块配置、网络类型 | ✅ 已完成 |
| 需求管理 | 需求池、Epic→Feature→Story 三级 WBS、评审工作流 | ✅ 已完成 |
| 任务管理 | WBS 子任务、工时管理、任务依赖（FS/SS/FF/SF） | ✅ 已完成 |
| 缺陷管理 | 缺陷跟踪、状态机、根因分析、严重程度 | ✅ 已完成 |
| 迭代管理 | Sprint 创建/启动/结束、容量规划、燃尽图 | ✅ 已完成 |
| 版本日管理 | 多迭代聚合、Release Notes、交付报告 | 🚧 规划中 |
| 项目仪表盘 | 可配置卡片、多项目聚合、预警规则 | 🚧 规划中 |
| 个人工作台 | 待办聚合、快捷操作、Focus Mode | 🚧 规划中 |
| 全局搜索 | 全文检索、类 JQL 语法、过滤器联动 | 🚧 规划中 |
| 通知中心 | 站内/邮件/IM 多渠道、订阅配置 | 🚧 规划中 |
| Webhook & API | 开放集成、签名验证、自动重试 | 🚧 规划中 |
| 知识库 | Markdown 文档、版本管理、评审流程 | 📋 规划中 |
| 研发效能 | DORA 指标、CFD 分析、累计流图 | 📋 规划中 |
| 收件箱 | 外部反馈收集、Intake Issue 转正 | 📋 规划中 |

## 架构设计

### 技术栈

| 层次 | 技术选型 | 说明 |
|------|----------|------|
| 后端 | Go 1.21+ + Gin | 高并发性能、编译部署简单 |
| 前端 | Vue 3.5 + TypeScript + Vite 6 | 组合式 API、类型安全 |
| UI 组件库 | Element Plus | 企业级设计、Vue 3 原生支持 |
| 状态管理 | Pinia | Vue 3 官方推荐 |
| 数据库 | PostgreSQL 16 | ACID 事务、JSON 支持、全文检索 |
| 缓存 | Redis 7 | 高性能 K/V、发布订阅 |
| 全文检索 | Elasticsearch 8 | 复杂查询、聚合分析 |
| 对象存储 | MinIO / S3 | 附件存储、私有部署 |
| 消息队列 | NATS / Asynq | 异步任务、事件驱动 |
| 部署 | Docker Compose / K8s | 一键部署、信创兼容 |

### 系统架构

```
┌────────────────────────────────────────────────────────────────┐
│                      展示层 (Vue 3 + Vite)                      │
│    SPA + Element Plus + Pinia + Vue Router + ECharts           │
└────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌────────────────────────────────────────────────────────────────┐
│                    API 网关 (Nginx / Traefik)                    │
│           负载均衡 / SSL 终结 / 静态资源 / Rate Limit           │
└────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌────────────────────────────────────────────────────────────────┐
│                   业务服务层 (Go + Gin)                          │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌───────────┐ │
│  │ Workspace   │ │ Project     │ │ Issue       │ │ Sprint    │ │
│  │ Service     │ │ Service     │ │ Service     │ │ Service   │ │
│  └─────────────┘ └─────────────┘ └─────────────┘ └───────────┘ │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌───────────┐ │
│  │ Search      │ │ Notif       │ │ Intake      │ │ Automate  │ │
│  │ Service     │ │ Service     │ │ Service     │ │ Service   │ │
│  └─────────────┘ └─────────────┘ └─────────────┘ └───────────┘ │
└────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌────────────────────────────────────────────────────────────────┐
│                    数据持久层                                    │
│    PostgreSQL    │    Redis    │    MinIO    │    ES           │
└────────────────────────────────────────────────────────────────┘
```

### 后端项目结构

后端采用 **DDD 轻量级分层架构**（Domain-Driven Design），模块化设计便于团队并行开发：

```
ydsz-plane/
├── cmd/                    # 入口程序
│   ├── api/                # API Server
│   └── worker/             # 异步任务 Worker
├── internal/               # 内部业务代码
│   ├── domain/             # 领域模型与业务规则
│   │   ├── workspace/
│   │   ├── project/
│   │   ├── issue/
│   │   ├── sprint/
│   │   └── version/
│   ├── application/        # 应用服务（用例编排）
│   ├── infrastructure/     # 基础设施实现
│   │   ├── persistence/    # 持久化（PostgreSQL）
│   │   ├── cache/          # Redis 缓存
│   │   ├── search/         # Elasticsearch
│   │   ├── queue/          # 消息队列
│   │   └── storage/        # 对象存储
│   └── interfaces/         # 接口层
│       ├── http/           # HTTP Handler
│       └── middleware/     # 中间件
├── pkg/                    # 可公开的工具包
├── api/                    # API 定义（OpenAPI / Proto）
├── configs/                # 配置文件模板
├── migrations/             # 数据库迁移脚本
└── scripts/                # 部署/运维脚本
```

### 前端项目结构

前端采用 **Vue 3 + TypeScript + Vite**，组合式 API + Pinia 状态管理：

```
ydsz-plane-web/
├── src/
│   ├── api/                # API 请求封装（Axios）
│   ├── assets/             # 静态资源
│   ├── components/         # 通用组件
│   ├── composables/        # 组合式函数
│   ├── layouts/            # 布局组件
│   ├── router/             # 路由配置
│   ├── stores/             # Pinia 状态管理
│   ├── types/              # TypeScript 类型
│   ├── utils/              # 工具函数
│   ├── views/              # 页面组件
│   ├── App.vue
│   └── main.ts
├── public/                 # 公共静态资源
└── tests/                  # 单元测试
```

## 快速开始

### Docker Compose（推荐）

```bash
# 克隆项目
git clone https://github.com/ydszopen/ydsz-plane.git
cd ydsz-plane

# 一键启动
docker compose up -d

# 访问 http://localhost:8080
```

### 本地开发

#### 前置要求

- Go 1.21+
- Node.js 20+
- PostgreSQL 16
- Redis 7

#### 启动后端

```bash
cd backend

# 安装依赖
go mod download

# 配置环境变量
cp .env.example .env
# 编辑 .env 配置数据库连接

# 运行数据库迁移
go run cmd/migrate/main.go

# 启动 API Server
go run cmd/api/main.go
```

#### 启动前端

```bash
cd web

# 安装依赖
pnpm install

# 启动开发服务器
pnpm dev

# 访问 http://localhost:5173
```

## 竞品对标

| 能力域 | 云效 | TAPD | ONES | PingCode | Jira | Ydsz Plane |
|--------|------|------|------|----------|------|------------|
| 需求管理 | ● | ● | ● | ● | ● | ● |
| 任务管理 | ● | ● | ● | ● | ● | ● |
| 缺陷管理 | ● | ● | ● | ● | ● | ● |
| 模块管理 | ● | ● | ● | ● | ● | △ |
| 迭代管理 | ● | ● | ● | ● | ● | ● |
| 版本日管理 | ● | ○ | ● | ● | ● | △ |
| 项目仪表盘 | ● | ● | ● | ● | ● | △ |
| 个人工作台 | ● | ● | ● | ● | ● | △ |
| 收件箱 | ● | ● | ● | ● | ● | ● |
| 效能度量 | ● | ○ | ● | ● | ○ | △ |
| 自动化 | ● | ● | ● | ● | ● | △ |
| 知识库 | ● | ○ | ● | ● | ● | △ |
| **开源自托管** | ✕ | ✕ | △ | △ | ✕ | **●** |
| **信创兼容** | ✕ | ✕ | △ | △ | ✕ | **●** |

> ● 已支持 / △ 部分支持或规划中 / ○ 不支持 / ✕ 不适用

## 路线图

### Phase 1 — MVP（当前）

- [x] 工作空间管理（CRUD、成员邀请、角色权限）
- [x] 项目管理（CRUD、Identifier、模板、归档）
- [x] 工作项管理（需求/任务/缺陷、WBS 层级、状态流转）
- [x] 迭代管理（创建/启动/结束、燃尽图）
- [x] 用户认证与权限体系
- [x] Docker Compose 一键部署

### Phase 2 — 功能完善

- [ ] 版本日管理（多迭代聚合、Release Notes）
- [ ] 项目仪表盘（可配置卡片、多项目聚合）
- [ ] 个人工作台（待办聚合、快捷操作）
- [ ] 全局搜索（全文检索、类 JQL 语法）
- [ ] 通知中心（多渠道推送、订阅配置）
- [ ] Webhook & OpenAPI（开放集成能力）
- [ ] 附件管理（对象存储集成）

### Phase 3 — 进阶增强

- [ ] 知识库（Markdown 文档、版本管理、评审）
- [ ] 收件箱（Intake Issue 转正流程）
- [ ] 自动化引擎（Trigger-Condition-Action）
- [ ] 研发效能度量（DORA 指标、CFD 分析）
- [ ] 甘特图 / 日历视图 / 电子表格视图
- [ ] SSO/SAML 集成

### Phase 4 — 生态扩展

- [ ] 国际化（中/英/日多语言）
- [ ] 移动端适配 / PWA
- [ ] 数据迁移工具（Jira/云效/ONES 导入）
- [ ] 插件市场 / 应用商店
- [ ] AI 功能（智能分配、重复检测、需求分析）

## 性能指标

| 指标 | 目标值 |
|------|--------|
| 页面加载 P95 | ≤ 2s |
| API 响应 P95 | ≤ 200ms |
| 并发用户数 | ≥ 1000 |
| 单项目工作项 | ≥ 100 万 |
| 可用性 SLA | ≥ 99.9% |

## 信创兼容

Ydsz Plane 从设计之初即考虑信创合规需求：

- **操作系统**：麒麟 V10 / 统信 UOS / openEuler
- **CPU 架构**：x86_64 / ARM64（鲲鹏 / 飞腾）
- **数据库**：PostgreSQL / 达梦 / 人大金仓
- **中间件**：Nginx / 东方通 / 宝兰德
- **浏览器**：360 安全浏览器 / 奇安信
- **密码算法**：支持 SM2/SM3/SM4 国密算法
- **等保合规**：满足等保三级基线要求

## 非功能需求

| 维度 | 措施 |
|------|------|
| **安全** | JWT + RBAC、HTTPS/TLS、bcrypt、审计日志、CORS/CSP/Rate Limit |
| **性能** | Redis 多级缓存、CDN、ES 全文检索、连接池、异步任务 |
| **可用** | 无状态服务、多副本、跨可用区部署、健康检查 |
| **可运维** | Docker Compose / Helm Chart、Prometheus + Grafana 监控、ELK 日志 |
| **可扩展** | 微服务就绪、读写分离、消息队列削峰 |

## 贡献指南

### 开发环境搭建

1. Fork 本仓库
2. Clone 你的 Fork：`git clone https://github.com/<your-name>/ydsz-plane.git`
3. 创建特性分支：`git checkout -b feature/your-feature`
4. 提交更改：`git commit -m "feat: your feature"`
5. 推送分支：`git push origin feature/your-feature`
6. 提交 Pull Request

### 提交规范

遵循 [Conventional Commits](https://www.conventionalcommits.org/)：

- `feat:` 新功能
- `fix:` 修复 Bug
- `docs:` 文档更新
- `refactor:` 重构
- `test:` 测试
- `chore:` 构建/工具变动

### 代码规范

- **Go**：`golangci-lint` + [Uber Go Style Guide](https://github.com/uber-go/guide)
- **TypeScript**：`eslint` + `prettier`
- **Vue**：组合式 API + `<script setup>`

## 社区与反馈

- **GitHub Issues**：Bug 反馈、功能建议
- **GitHub Discussions**：使用问答、设计讨论
- **微信群**（待创建）
- **邮件**：ydszopen@meituan.com

## 许可协议

Ydsz Plane 采用 [MIT License](LICENSE) 开源协议。

```
MIT License

Copyright (c) 2026 Ydsz OpenSource

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
```

## 致谢

本项目的设计参考了以下优秀的开源项目与商业产品：

- [Plane](https://github.com/makeplane/plane) - 开源项目管理工具
- [GitLab](https://gitlab.com/gitlab-org/gitlab) - DevOps 平台
- [Taiga](https://github.com/taigaio) - 敏捷项目管理
- [Focalboard](https://github.com/mattermost/focalboard) - 看板工具
- [云效](https://www.aliyun.com/product/yunxiao) - 阿里云研发平台
- [TAPD](https://www.tapd.cn) - 腾讯敏捷协作平台
- [ONES](https://ones.cn) - 企业级研发管理平台

---

<p align="center">
  ⭐ 如果觉得这个项目对你有帮助，欢迎 <b>Star</b> 支持我们
</p>
