# S14 · Phase 3 — 微服务拆分

> 版本：v1.0（参考美团/字节/Google 工程标准制定）
> 依据：`README.md` 路线图 Phase 3 + 互联网大厂演进最佳实践
> 日期：2026-08-08

---

## 0. 总览与优先级矩阵

S14 聚焦**微服务拆分**。按大厂"价值交付优先、风险前置"原则排序：

| 优先级 | 模块 | 交付物 | 风险 | 业务价值 | 预估工期 |
|--------|------|--------|------|----------|----------|
| **P0** | 微服务拆分 | 模块依赖图 + 边界分析报告 | 低 | 中（指导后续拆分方向） | 3 天 |
| **P2-A** | 微服务拆分 | API 契约 Proto 化 + 事件总线解耦 | 中 | 中（必要前置条件） | 1 周 |
| **P2-B** | 微服务拆分 | 通知服务独立部署首个微服务 | 中 | 高（独立扩缩容） | 1 周 |
| **P3** | 微服务拆分 | 搜索服务独立部署 | 中 | 高 | 1 周 |

**核心决策**：微服务拆分采用渐进式（先分析、先解耦、逐步剥离），绝不"big bang"重写。

---

## 1. 微服务拆分

### 1.1 拆分原则（对标美团/字节内部架构规范）

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

### 1.2 服务拓扑（目标架构）

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

### 1.3 分阶段实施路线图

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

### 1.4 服务治理（对标美团 OCTO/字节 ServiceMesh）

拆分后需补齐的治理能力：

| 能力 | 实现方案 | 优先级 |
|------|----------|--------|
| 服务注册与发现 | Consul / Nacos（共用现有 RabbitMQ 基础设施时先简化） | P2 |
| 服务间通信 | gRPC + Protobuf（Quest/调用走 gRPC；前端→API 仍走 REST） | P2 |
| 链路追踪 | OTel + Jaeger（沿用现有，扩展至跨服务 trace） | P2 |
| 熔断降级 | Client 端 circuit breaker（-go-zero 模式或自实现） | P2 |
| 配置中心 | Viper + PostgreSQL（渐进式，替代 Nacos 重依赖） | P3 |
| API Gateway | Nginx + Lua / APISIX（替代硬编码路由） | P3 |

### 1.5 拆分反模式检查表（每阶段上线前自查）

- [ ] 禁止拆分后跨服务本地事务（替换为 Saga / Outbox + 幂等消费）
- [ ] 禁止跨服务共享数据库表（每个服务独占 schema）
- [ ] 禁止同步 RPC 链 > 3 跳（A→B→C→D 需要合并或事件化）
- [ ] 禁止"分布式单体"（服务部署仍必须同步 → 说明边界没切对）

---

## 2. 实施排期

### Week 1（P0：分析）

| 天数 | 微服务拆分 |
|------|-----------|
| D1-D3 | 模块依赖图绘制（import graph）+ 边界分析报告 |
| D4 | API 调用频次统计脚本 |
| D5 | 拆分就绪度评审 |

### Week 2–3（P2-A：微服务拆分准备）

- Proto 契约定义（notification / search / webhook）
- 本地 gPRC Service 实现（进程内调用验证）
- 全量测试双跑（monolith vs gRPC）等价验证
- 性能基线确认

### Week 4（P2-B：通知服务独立部署）

- Schema 拆分（expand-contract 灰度迁移）
- Notification Service 独立进程 + 消费者切割
- Docker Compose 集成部署
- 灰度验证 + 监控告警

### Week 5–6（P3：搜索服务独立）

- 搜索服务拆分上线
- S14 整体验证 + 文档更新

---

## 3. 退出检查表

**S14 整体验证**：
- [ ] Notification Service 独立部署，功能等价于单体
- [ ] Search Service 独立部署，ES 读写无退化
- [ ] 全量 API 兼容性回归通过
- [ ] 性能基线回归（1000 并发 / P95 ≤ 200ms）
- [ ] 部署文档更新（docker-compose.s14.yml + Helm 扩展）
- [ ] 运维 Runbook 覆盖新组件告警
- [ ] CHANGELOG + Release Notes 更新
