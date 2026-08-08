# S14 · Phase 3 — 微服务拆分与原生桌面客户端

> 版本：v1.0（参考美团/字节/Google 工程标准制定）
> 依据：`README.md` 路线图 Phase 3 + 互联网大厂演进最佳实践
> 日期：2026-08-08

---

## 0. 总览与优先级矩阵

S14 涵盖两大能力域，**独立交付、互不包含**。按大厂"价值交付优先、风险前置"原则排序：

| 优先级 | 模块 | 交付物 | 风险 | 业务价值 | 预估工期 |
|--------|------|--------|------|----------|----------|
| **P0-A** | 桌面客户端 | Wails v2 + Vue 3 技术验证报告 | 低 | 高（覆盖无浏览器场景） | 2 天 |
| **P0-B** | 微服务拆分 | 模块依赖图 + 边界分析报告 | 低 | 中（指导后续拆分方向） | 3 天 |
| **P1** | 桌面客户端 | MVP 应用（认证 + 看板 + 托盘 + 通知） | 中 | 高 | 2 周 |
| **P2-A** | 微服务拆分 | API 契约 Proto 化 + 事件总线解耦 | 中 | 中（必要前置条件） | 1 周 |
| **P2-B** | 微服务拆分 | 通知服务独立部署首个微服务 | 中 | 高（独立扩缩容） | 1 周 |
| **P3-A** | 桌面客户端 | 离线缓存 + 自动更新 + 多窗口 | 中 | 中 | 1 周 |
| **P3-B** | 微服务拆分 | 搜索服务独立部署 | 中 | 高 | 1 周 |

**核心决策**：桌面客户端先做（独立交付、不扰动现有架构），微服务拆分采用渐进式（先分析、先解耦、逐步剥离），绝不"big bang"重写。

---

## 1. 原生桌面客户端

### 1.1 技术选型

参考 2026 年桌面 GUI 生态（对标 Linear/Notion/Slack 客户端架构）：

| 维度 | Wails v2 | Tauri v2 | Fyne | Electron |
|------|----------|----------|------|----------|
| 后端语言 | **Go ✅（复用现有技术栈）** | Rust | Go | Node.js |
| 前端技术 | **Vue 3 ✅（复用 web 代码）** | React/Vue/Svelte | Go 自绘 | React/Vue |
| 空包体积 | ~12MB | ~5MB | ~15MB | ~150MB |
| 内存占用 | ~80MB | ~50MB | ~60MB | ~300MB |
| 编译速度 | 快（增量 ~5s） | 慢 | 快 | N/A |
| 移动端 | 仅桌面 | 支持 | 仅桌面 | 仅桌面 |
| 信创兼容 | WebView2（Win11 自带） | WebView2 | 自绘（最佳） | Chromium 自带 |

**结论**：Wails v2

- **决定性因素**：团队技术栈 = Go + Vue 3，现有 `web/src/` 的路由、Store、API 封装、组件库可**大量复用**，前端同学零学习成本切入桌面端
- **兜底方案**：Wails 依赖 WebView2（Windows 自带 99.1% 渗透率），信创/旧版 Windows 系统提供 WebView2 Runtime 离线安装引导
- **渐进式能力**：Wails 天然支持将桌面应用改造为 PWA，未来可平滑过渡到 Tauri（如需移动端）

### 1.2 架构设计

```
Ydsz Plane Desktop (Wails v2)
├── cmd/desktop/main.go          # Wails 应用入口（Go 端）
├── desktop/                     # 桌面端专属 Go 逻辑
│   ├── auth/
│   │   ├── store.go             # Token 持久化（OS Keychain: Windows DPAPI/macOS Keychain/libsecret）
│   │   ├── sso.go               # 系统浏览器 OAuth/OIDC 回调（localhost 临时 HTTP server）
│   │   └── session.go           # 会话状态机（自动刷新 + 过期登出）
│   ├── tray/
│   │   ├── tray.go              # 系统托盘（图标 + 右键菜单 + 未读角标）
│   │   └── notification.go      # 原生系统通知（Windows Toast/macOS NotificationCenter）
│   ├── window/
│   │   ├── manager.go           # 多窗口管理（主窗口 / 快速创建浮层）
│   │   └── shortcuts.go         # 全局快捷键（Ctrl+Shift+Y 唤起、Cmd+N 新建）
│   ├── cache/
│   │   ├── sqlite.go            # 本地 SQLite 缓存（工作项/项目元数据）
│   │   └── sync.go              # 后台同步策略（增量拉取 + 冲突解决）
│   ├── updater/
│   │   └── updater.go           # 自动更新（Squirrel.Windows / appimage / dmg）
│   └── bridge/
│       └── events.go            # Go → JS 事件推送（复制现有 ws-hub 模式）
├── frontend/                    # 桌面端前端（独立于 web，共享 packages/）
│   ├── wailsjs/                 # Wails 自动生成（Go ↔ JS 绑定）
│   ├── src/
│   │   ├── views/               # 桌面端视图（精简版：侧边栏 + 看板 + 详情）
│   │   ├── layouts/             # DesktopLayout（无浏览器 chrome）
│   │   └── lib/
│   │       ├── bridge.ts        # 封装 wails Go 方法（替代部分 axios 调用）
│   │       └── keychain.ts      # 调用 Go Keychain 封装
│   └── shared/                  # ← 符号链接/引用 web/packages/ 共享组件
└── wails.json                   # Wails 配置（应用元数据/构建选项）
```

### 1.3 与 Web 端的复用策略

| 层次 | 复用方式 | 说明 |
|------|----------|------|
| `web/packages/ui/` | **直接依赖** | 基础组件（Button/Dropdown/Modal）无平台差异 |
| `web/src/api/` | **直接依赖** | axios 封装、OpenAPI 生成 client 完全复用 |
| `web/src/stores/` | **选择性复用** | Pinia auth/ui store 复用；realtime store 适配 Wails 事件桥 |
| `web/src/router/` | **独立实现** | 桌面端路由更扁平（无浏览器 URL 栏需求） |
| `web/src/views/` | **选择性复用** | IssueDetailView 适配桌面布局；看板/列表复用 |
| `web/src/design/` | **直接依赖** | 设计令牌（CSS 变量）100% 复用 |
| `web/src/modules/` | **独立实现** | 桌面端只保留核心域（workspace/project/issue/dashboard） |

### 1.4 桌面端 MVP 功能清单

**认证系统（P1 优先）**：
- JWT Token 持久化（Go 端加密存储到 OS Keychain）
- 会话自动续期（在 Go 端完成 refresh 轮换，前端无感知）
- OAuth/OIDC 系统浏览器回调（`wails:localhost` 临时 HTTP server 捕获回调参数）
- 多工作空间切换

**核心视图**：
- 工作空间侧边栏（项目列表 + 快速导航）
- 项目看板（拖拽 + WIP 限制 + 组视图）
- 工作项详情（状态流转 + 评论 + 活动日志）
- 工作台（跨空间聚合待办）

**系统集成**：
- 系统托盘（最小化到托盘 + 未读通知角标）
- 原生系统通知（Windows Toast / macOS UNUserNotification）
- 全局快捷键（唤起应用 / 快速创建工作项）
- 应用徽章（Dock/任务栏未读计数）

### 1.5 信创兼容策略

Windows 信创环境（统信 UOS/麒麟 Kylin）：
- 统信 UOS 已内置 **GNOME WebView**，Wails 通过 WebKitGTK 渲染
- 麒麟 V10（x86）：内置 **CEF（Chromium Embedded Framework）**，提供兼容层
- 兜底方案：检测 WebView2/WebKitGTK 不可用时，引导用户下载 WebView2 Runtime（打包在应用内 ~130MB）

### 1.6 构建与分发

```yaml
# .github/workflows/desktop.yml
targets:
  windows: { os: windows-latest, arch: [amd64, arm64] }
  macOS:   { os: macos-latest, arch: [amd64, arm64] }   # Universal Binary
  linux:   { os: ubuntu-22.04, arch: [amd64] }           # AppImage + deb
output:
  windows: YdszPlane-Setup-{version}.exe  # NSIS 安装包
  macOS:   YdszPlane-{version}.dmg        # Apple Notarization
  linux:   YdszPlane-{version}.AppImage
更新机制:
  windows: Squirrel.Windows（增量更新）
  macOS:   Squirrel.Mac 或 Wails 内置
  linux:   AppImageUpdate（Zsync）
签名: Windows EV Code Signing + macOS Developer ID
```

---

## 2. 微服务拆分

### 2.1 拆分原则（对标美团/字节内部架构规范）

互联网大厂微服务拆分的**三大不拆原则**：
1. **不拆"调用稠密"边界** — 若模块间同步调用 > 50 次/秒，暂缓拆分（网络开销 > 收益）
2. **不拆"事务耦合"边界** — 若两个模块共享同一数据库事务（非事件通信），暂缓拆分
3. **拆分前必先"契约先行"** — API 必须 Proto 化 + 版本兼容策略就绪

**Ydsz Plane 拆分就绪度评估**：

| 模块 | 独立事务 | 独立存储 | 事件边界 | 就绪度 | 拆分顺序 |
|------|----------|----------|----------|--------|----------|
| **notification** | ✅ | ✅（PG 独立 schema 就绪） | ✅（纯消费端，无同步出站调用） | 🟢 高 | **Phase-1 首选** |
| **search** | ✅ | ✅（ES 已独立） | ✅（纯消费端，PG→ES 单向） | 🟢 高 | **Phase-2** |
| **webhook** | ✅ | ✅（webhook_logs 独立表） | ✅（出站 HTTP 独立） | 🟢 高 | Phase-3 |
| **automation** | ⚠️（聚合查询 issue 状态） | ⚠️（读写 rule_executions） | ⚠️（需查询 project/issue） | 🟡 中 | Phase-4 |
| **metrics** | ⚠️（聚合查询 sprint/version/issue） | ✅（metric_snapshots） | ✅（纯消费端） | 🟡 中 | Phase-5 |
| **issue** | ❌（核心聚合） | ❌（主库） | ❌（所有域都依赖红色事件） | 🔴 低 | ❌ 长期保持单体 |

### 2.2 服务拓扑（目标架构）

```
                      ┌───────────────────────────────┐
                      │    Nginx / API Gateway          │
                      │  (路由 / 限流 / 认证聚合)         │
                      └──────────────┬────────────────┘
                                     │
          ┌─────────────────────────▼──────────────────────────┐
          │          核心服务（ydsz-plane-core）                 │
          │  单体保留: auth · workspace · project · issue        │
          │           sprint · version · intake · ai · kb        │
          └───────────────┬───────────────────────┬────────────┘
                          │ 异步事件（RabbitMQ）    │ 同步 gRPC（按需）
          ┌───────────────▼─────────┐   ┌─────────▼──────────┐
          │  Notification Service   │   │   Search Service    │
          │  (独立部署/独立扩缩容)      │   │  (ES 读写/分词/对账)  │
          │  消费: *.created/updated │   │  消费: issue.*/page.* │
          │  投递: 站内/邮件/IM/Webhook│   │  输出: ES bulk index  │
          └─────────────────────────┘   └─────────────────────┘
                    未来扩展 ↓
          ┌─────────────────────────┐   ┌──────────────────────┐
          │  Webhook Service        │   │   Metrics Service     │
          │  (出站 HTTP/签名/重试)     │   │  (离线聚合/DORA/CFD)   │
          └─────────────────────────┘   └──────────────────────┘
```

### 2.3 分阶段实施路线图

#### Phase-0：契约先行 + 事件总线解耦（P2-A）

**目标**：让模块具备"可拆"条件——先解耦，不拆进程。

```go
// 步骤 1: 定义 gRPC 服务契约（跨服务调用规范）
// api/proto/notification/v1/notification.proto
service NotificationService {
  rpc ListNotifications(ListNotificationsReq) returns (ListNotificationsResp);
  rpc MarkRead(MarkReadReq) returns (MarkReadResp);
  rpc GetUnreadCount(GetUnreadCountReq) returns (GetUnreadCountResp);
  rpc ListPreferences(ListPreferencesReq) returns (ListPreferencesResp);
  rpc UpdatePreference(UpdatePreferenceReq) returns (UpdatePreferenceResp);
}

// 步骤 2: 抽象事件内部门接口（为 RPC 替换本地调用做准备）
// pkg/eventbus/publisher.go
type EventPublisher interface {
    Publish(ctx context.Context, event DomainEvent) error
    // 未来扩展：PublishRemote → RabbitMQ Publish
    // 当前实现：LocalPublisher → 本地 direct call
}

// 步骤 3: Notification 模块接口化 -- 定义 gRPC Service 接口
// internal/application/notification/service_interface.go
type NotificationModule interface {
    Dispatch(ctx context.Context, event DomainEvent) error
    GetUnreadCount(ctx context.Context, userID int64) (int, error)
    SendTestNotification(...) error
}
```

**Phase-0 验收标准**：
- [ ] 至少 3 个模块完成 Proto 契约定义
- [ ] Notification 模块的 HTTP Handler 切换为 gRPC Service 实现（内部进程调用，功能等价）
- [ ] 全量测试通过（单元 + 集成 + E2E）
- [ ] 性能基线回归无退化（k6 ±5%）

#### Phase-1：通知服务独立部署（P2-B）

Notification 模块是最理想的"第一个微服务"候选：
- **纯消费端**：仅消费 RabbitMQ 事件，出站仅需 HTTP（邮件/IM/Webhook）
- **独立存储就绪**：`notifications` / `notification_subscriptions` 表已具备独立迁移条件
- **独立扩缩容价值高**：高频通知场景下（千人团队），通知投递可独立水平扩展

```yaml
# docker-compose.s14.yml（增量部署）
version: "3.8"
services:
  # ----- 原有服务（同 S12） -----
  api: { image: ydsz-plane-api:latest, ... }
  worker: { image: ydsz-plane-worker:latest, ... }
  
  # ----- 新增：独立通知服务 -----
  notification-svc:
    build: ./cmd/notification-service
    environment:
      - RABBITMQ_URL=amqp://rabbitmq:5672
      - DATABASE_URL=postgres://notification_db:5432/notifications
      - SMTP_HOST=mailpit
      - WEBHOOK_ENABLED=true
    depends_on: [rabbitmq, notification_db]
    deploy:
      replicas: 2
      resources: { limits: { cpus: '1', memory: 512M } }
  
  # 通知服务独立数据库（schema 拆分）
  notification_db:
    image: postgres:18
    volumes:
      - ./sql/notification_schema_v1.sql:/docker-entrypoint-initdb.d/
      - notif_pg_data:/var/lib/postgresql/data
```

**Phase-1 实施关键点**：

1. **数据库 Schema 拆分**（expand-contract 零停机迁移）：
   - 新建 `notification_db` PG 实例
   - `notification` 表 triggers + logical replication 同步新增数据
   - 灰度切换 writer → 停止同步 → 切换 reader → 下线旧表

2. **进程边界切割**：
   - 抽离 `cmd/notification-service/main.go`（从 worker 中独立）
   - Stop consuming notification tasks from main worker
   - 新增 Notification Service gRPC server

3. **配置热切换**：通过 feature flag 控制通知路由（单体 → 独立服务灰度）

```go
// cmd/api/feature_flags.go
type NotificationRouting string

const (
    RoutingMonolith  NotificationRouting = "monolith"      // 默认
    RoutingHybrid    NotificationRouting = "hybrid"       // 灰度：部分通知走独立服务
    RoutingService   NotificationRouting = "service"      // 全量走独立服务
)

var currentRouting = RoutingMonolith  // 由配置中心/Consul 控制
```

### 2.4 服务治理（对标美团 OCTO/字节 ServiceMesh）

拆分后需补齐的治理能力：

| 能力 | 实现方案 | 优先级 |
|------|----------|--------|
| 服务注册与发现 | Consul / Nacos（共用现有 RabbitMQ 基础设施时先简化） | P2 |
| 服务间通信 | gRPC + Protobuf（Quest/调用走 gRPC；前端→API 仍走 REST） | P2 |
| 链路追踪 | OTel + Jaeger（沿用现有，扩展至跨服务 trace） | P2 |
| 熔断降级 | Client 端 circuit breaker（-go-zero 模式或自实现） | P2 |
| 配置中心 | Viper + PostgreSQL（渐进式，替代 Nacos 重依赖） | P3 |
| API Gateway | Nginx + Lua / APISIX（替代硬编码路由） | P3 |

### 2.5 拆分反模式检查表（每阶段上线前自查）

- [ ] 禁止拆分后跨服务本地事务（替换为 Saga / Outbox + 幂等消费）
- [ ] 禁止跨服务共享数据库表（每个服务独占 schema）
- [ ] 禁止同步 RPC 链 > 3 跳（A→B→C→D 需要合并或事件化）
- [ ] 禁止"分布式单体"（服务部署仍必须同步 → 说明边界没切对）

---

## 3. 实施排期

### Week 1（P0：技术验证 + 分析）

| 天数 | 桌面客户端 | 微服务拆分 |
|------|-----------|-----------|
| D1-D2 | Wails v2 环境搭建 + 最小可运行 Demo | — |
| D3 | — | 模块依赖图绘制（import graph）|
| D4 | Token 持久化 PoC（Windows DPAPI Keychain） | API 调用频次统计脚本 |
| D5 | 构建产物 Demo 验证 + 技术选型评审通过 | 边界分析报告 + 拆分就绪度评审 |

### Week 2–3（P1：桌面客户端 MVP）

- 桌面项目骨架 + Wails 配置
- 认证流程（Token 存储 + 自动刷新 + SSO 回调）
- 工作空间选择 → 项目看板 → 工作项详情
- 系统托盘 + 全局快捷键 + 应用图标
- E2E 冒烟测试（桌面端 Playwright/Spectron）

### Week 4–5（P2-A：微服务拆分准备）

- Proto 契约定义（notification / search / webhook）
- 本地 gPRC Service 实现（进程内调用验证）
- 全量测试双跑（monolith vs gRPC）等价验证
- 性能基线确认

### Week 6（P2-B：通知服务独立部署）

- Schema 拆分（expand-contract 灰度迁移）
- Notification Service 独立进程 + 消费者切割
- Docker Compose 集成部署
- 灰度验证 + 监控告警

### Week 7–8（P3：桌面增强 + 搜索服务独立）

- 桌面端离线缓存（SQLite + 后台 sync）
- 桌面端自动更新（Squirrel/dmg/AppImage）
- 搜索服务拆分上线
- S14 整体验证 + 文档更新

---

## 4. 退出检查表

**S14 整体验证**：
- [ ] 桌面客户端 macOS/Windows/Linux 三平台构建通过
- [ ] 桌面端 E2E 冒烟测试通过（登录→看板→详情→通知）
- [ ] 桌面端性能基线：启动 < 1s / 内存 < 150MB
- [ ] Notification Service 独立部署，功能等价于单体
- [ ] Search Service 独立部署，ES 读写无退化
- [ ] 全量 API 兼容性回归通过
- [ ] 性能基线回归（1000 并发 / P95 ≤ 200ms）
- [ ] 部署文档更新（docker-compose.s14.yml + Helm 扩展）
- [ ] 运维 Runbook 覆盖新组件告警
- [ ] CHANGELOG + Release Notes 更新
