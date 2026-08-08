# S14 微服务拆分就绪度分析报告

> 分析日期：2026-08-08
> 分析范围：`internal/application/` 下 20 个业务模块
> 分析依据：Go import 静态分析 + 代码结构审查
> 等价参考：美团 OCTO 拆分评估框架、字节 Module Coupling Index

---

## 1. 模块依赖图（基于 Go import 分析）

### 1.1 内部模块依赖矩阵

| 模块 | 导入的内部模块 | 外部依赖 | 耦合度 | 独立存储 | 独立事务 | 就绪度 |
|------|---------------|----------|--------|----------|----------|--------|
| **notification** | `auth` 仅此一个 | RabbitMQ | 🟢 极低 | ✅ | ✅ | **🟢 就绪** |
| **search** | `auth` 仅此一个 | ES + RabbitMQ | 🟢 极低 | ✅ | ✅ | **🟢 就绪** |
| **webhook** | `auth` 仅此一个 | 出站 HTTP | 🟢 极低 | ✅ | ✅ | **🟢 就绪** |
| **metrics** | 无内部依赖 | Redis 缓存 | 🟢 极低 | ✅ | ✅ | **🟢 就绪** |
| **intake** | 无内部依赖 | — | 🟢 极低 | ✅ | ✅ | 🟢 就绪 |
| **version** | 无内部依赖 | PG + Redis | 🟡 低 | ✅ | ⚠️ 共享 Sprint 表 | 🟡 |
| **sprint** | 无内部依赖 | PG + Redis | 🟡 低 | ✅ | ⚠️ 共享 Issue 表 | 🟡 |
| **dashboard** | 无内部依赖 | PG | 🟡 低 | ✅ | ⚠️ 跨表聚合 | 🟡 |
| **preference** | 无内部依赖 | PG | 🟡 低 | ✅ | ✅ | 🟡 |
| **workbench** | 无内部依赖 | PG | 🟡 低 | ✅ | ⚠️ 跨项目聚合 | 🟡 |
| **pages** | 无内部依赖 | PG + ES | 🟡 低 | ✅ | ✅ | 🟡 |
| **attachment** | 无内部依赖 | MinIO | 🟡 低 | ✅ | ✅ | 🟡 |
| **automation** | `issue + notification + webhook` | PG + RabbitMQ | 🟠 中 | ⚠️ 跨模块读写 | ⚠️ | 🟡 |
| **ai** | 无内部依赖 | OpenAI/Claude API | 🟢 极低 | ✅ | ✅ | 🟡（归为单体更合适） |
| **apitoken** | 无内部依赖 | PG | 🟢 极低 | ✅ | ✅ | 🔴 与 auth 高度耦合 |
| **auth** | 无内部依赖 | PG + Redis | 🟢 极低 | ✅ | ✅ | 🔴 基础设施核心 |
| **dlq** | 无内部依赖 | PG | 🟢 极低 | ✅ | ✅ | 🔴 基础设施核心 |
| **issue** | `notification + auth` | PG + WS | 🔴 高（被依赖方） | ❌ 主库 | ❌ 核心聚合 | 🔴 核心域 |
| **workspace` | 无内部依赖+多子模块 | PG | 🟠 中 | ⚠️ | ⚠️ | 🔴 核心域 |

### 1.2 关键依赖路径

```
                      auth ← 所有模块的"基础设施依赖"
                       ↑
              ┌────────┴────────┐
              │                 │
          notification      search           webhook         metrics
              ↑                 ↑                ↑
              │                 │                │
          ┌───┴───┐             │          ┌─────┴──────┐
          │       │             │          │            │
      automation  │        automation    automation   自动化动作目标
          │       │             │
          └───────┴──────┬──────┘
                         │
                        issue ← 核心聚合根
                         ↑
              ┌─────────┼─────────┐
              │         │         │
          workspace  sprint   version
```

**核心发现**：
- 除 `automation` 和 `issue` 外，几乎所有模块对内部包只有 `auth` 一种横向依赖
- 现有架构已天然适合"按 DDD 限界上下文独立服务"拆分
- `issue` 是所有依赖的汇聚点 → 应永远保留在核心服务中

---

## 2. API 契约边界分析

### 2.1 现有 REST API 端点分布（约 120+ 路由）

| 模块 | 路由组 | GET | POST/PATCH/DELETE | 同步出站调用 |
|------|--------|-----|-------------------|-------------|
| auth | `/auth/*` | — | 5 | — |
| workspace | `/workspaces/:ws` | 6 | 8 | — |
| issue | `/projects/:pid/issues/*` | 6 | 8 | WebSocket 推送 |
| sprint | `/projects/:pid/sprints/*` | 4 | 6 | WebSocket 推送 |
| version | `/projects/:pid/versions/*` | 4 | 5 | — |
| search | `/projects/:pid/search` + `/ws/:ws/search` | 4 + 4 | 1 | — |
| notification | `/ws/:ws/notifications` | 2 | 3 | 出站 SMTP/IM/Webhook |
| webhook | `/ws/:ws/webhooks` | 2 | 4 | 出站 HTTP |
| metrics | `/projects/:pid/metrics` | 8 | 1 | — |
| automation | `/projects/:pid/automation` | 2 | 5 | **→ issue 状态变更** |
| intake | `/ws/:ws/intake` + `/public/intake` | 3 | 4 | **→ issue 创建** |

### 2.2 必须 Proto 化的接口（Phase-0 范围）

排序依据：接口跨进程调用的潜在频率 × 接口稳定性权重

| 优先级 | Proto 服务 | 接口数 | 理由 |
|--------|------------|--------|------|
| P0 | `NotificationService` | 6 RPCs | 首个独立微服务目标 |
| P0 | `SearchService` | 4 RPCs | 第二个独立目标 |
| P1 | `WebhookService` | 4 RPCs | 出站服务独立 |
| P1 | `MetricsService` | 6 RPCs | 统计聚合独立 |
| P2 | `AutomationService` | 8 RPCs | 依赖 issue 接口，需先稳定 event |
| P3 | `IssueService`（仅读） | 6 RPCs | 读操作可独立为 view svc |

---

## 3. 数据库 Schema 拆分就绪度

### 3.1 核心表归属分析

```sql
-- 独立 Schema 候选（完全属于单一模块）
-- 可直接迁移到 notification_db
notification_schema = {
    notifications,
    notification_subscriptions,
}

-- 独立 Schema 候选
search_schema = {
    search_history,
    search_bookmarks,
    -- ES 索引不属 PG，此处仅 bookmark 数据
}

-- 独立 Schema 候选
webhook_schema = {
    webhooks,
    webhook_logs,    -- 分区表，数据量大，独立库收益高
}

-- 独立 Schema 候选
metrics_schema = {
    metric_snapshots,
    -- CFD/控制图数据可通过重建获取
}

-- 跨模块共享表（保留在 core_db）
shared_once = {
    users,              -- auth
    workspaces,         -- workspace
    workspace_members,  -- workspace / auth
    projects,           -- workspace / project
    project_members,    -- workspace
    issues,             -- issue（核心聚合）
    sprints,            -- sprint / version
    versions,           -- version / sprint
    issues 相关 M2M 表,  -- issue / label / module / state
}
```

### 3.2 拆分 Migration 策略（Expand-Contract）

```
Step 1 [Expand]    新建 notification_db + 表结构（不影响现有 DB）
Step 2 [双写]      API 代码同时写入 core_db.notifications + notification_db.notifications
Step 3 [回填]      存量数据全量同步至 notification_db
Step 4 [切读]      灰度切读流量 1% → 10% → 50% → 100%
Step 5 [停写]      停止 core_db 写入，确认无 diff
Step 6 [Contract]  core_db.notifications 表标记 DEPRECATED → 下一版本删除

预估耗时：双写 1 天 + 回填 4h（1000 万条通知 = ~30min）+ 灰度 3 天
```

---

## 4. 推荐拆分顺序

**原则**：就绪度 从高到低 × 拆分收益 从高到低 × 风险 从低到高

### 第一批（S14 P2 执行）：独立通知服务

**理由**：
1. 独立依赖仅限 `auth`（垂直依赖）
2. 纯消费端：只消费 RabbitMQ 事件，出站无同步调用
3. 数据自治：`notifications` 表已物理独立
4. 扩展价值：通知投递是典型"流量尖峰"场景（千人团队同步 @ 所有人时）
5. 失败隔离：通知服务宕机不影响核心工作项 CRUD

### 第二批（S14 P3 执行）：独立搜索服务

**理由**：
1. ES 已独立，数据自治天然满足
2. 纯消费端：DB→ES 单向索引
3. 扩展价值：搜索/分析负载与 OLTP 查询完全不同（CPU/IO 模式差异大）
4. 故障隔离：搜索异常不影响核心业务

### 第三批 +（Phase 4+）可能的拆分

- **Webhook Service** — 出站 HTTP 独立（需解决重试幂等问题）
- **Metrics Service** — 离线聚合独立（计算密集型与 OLTP 分离）
- **File/Attachment Service**（如果 MinIO 流量增大）
- 暂不拆分：AI automation（太依赖 issue，RPC 链过深收益低）

---

## 5. 关键风险与缓解

| 风险 | 概率 | 缓解措施 |
|------|------|----------|
| Schema 拆分后跨库 JOIN 查询退化 | 高 | 冗余字段 + 事件同步宽表 |
| 分布式事务（通知状态 + 已读回写） | 中 | Outbox + 消费端幂等（event_id 去重） |
| gRPC 网络延迟取代内存调用 | 中 | 仅通知 1 次 RPC（≤1ms），通知事件本身异步 |
| 拆分后本地单元测试 mock 复杂化 | 中 | 保留 interface 抽象，单测走 mock，集成测试用 testcontainers |
| 团队运维成本上升 | 低（当前单人） | Docker Compose 一键起；后续逐步补 K8s / Helm |
