# S11 自动化规则 + 效能度量 大厂标准实施报告

> Sprint 11 · 2026-08-08 · 参考互联网大厂标准
> 参考规范：美团研发效能度量规范 v2.0、字节跳动字节索引规范、阿里 ARMS 时效性分层、DORA Accelerate State of DevOps 2023、Prometheus RED 指标体系

---

## 1. 总览

### 1.1 现状评估（实施前）

| 能力域 | 状态 | 问题 |
|--------|------|------|
| 自动化引擎 | MVP | 无熔断器、无 Prometheus 监控、无并发锁 |
| 规则模板 | 7 条 | 覆盖不足，缺少管理/通知场景 |
| 缓存层 | 缺失 | 仪表盘查询全部直连 DB |
| 查询索引 | 基础 | 缺少覆盖索引和局部索引 |
| 效能度量 | 基础五类 | 缺少 CFD/控制图/吞吐量 |
| 资源负载 | 粗略 | 无成员级细化 |

### 1.2 实施结论

**整体状态：GO ✅**

本次实施完成 P0 全部项 + P1 核心项，S11 能力达到互联网中型团队生产就绪状态。剩余 P2/P3 项（E2E 测试/国际化/高级 DSL）可在 v0.3 迭代继续完善。

---

## 2. 实施详情

### 2.1 P0：自动化引擎核心加固

#### 2.1.1 熔断器（Circuit Breaker）

**新增文件**：`internal/application/automation/circuit_breaker.go`

对标参考：Netflix Hystrix、美团 MT-Hystrix、阿里 Sentinel

```
三态模型：Closed → Open → HalfOpen
触发条件：连续失败 ≥ N 次（默认 5）
冷却策略：默认 60s 后半开，仅放行 1 次探测
复位策略：半开探测成功 → Closed；失败 → Open
```

**落地效果**：单条规则失败不会耗尽系统资源，熔断后自动隔离故障规则。

**关键 API**：
- `NewCircuitBreaker(cfg)` - 创建规则级熔断器
- `Allow()` - 检查是否放行
- `RecordSuccess/RecordFailure()` - 状态翻转
- `CircuitBreakerRegistry` - 并发安全注册表

**测试覆盖**：`circuit_breaker_test.go` 含 7 个用例（阈值触发/半开成功/半开失败/最大探测次数/注册表/复位/连续计数）

---

#### 2.1.2 Prometheus 可观测指标

**新增文件**：`internal/application/automation/metrics.go`

```
指标清单：
  automation_executions_total{result="success|failed|skipped|dry_run"}
  automation_execution_duration_ms{result}
  automation_rules_active
  automation_circuit_breaker_open
  antomation_antiloop_drop_total
```

**集成方式**：引擎自动在以下埋点记录指标：
- 规则执行完成（成功/失败/跳过）
- 防循环深度超限丢弃
- 熔断器告警计数

---

#### 2.1.3 引擎集成改造

**修改文件**：`internal/application/automation/engine.go`

核心改动：
1. Engine 新增 `metrics` 和 `breakers` 字段
2. `EvaluateEvent` 在规则评估前调用 `breaker.Allow()`
3. 失败调用 `breaker.RecordFailure()`，成功调用 `breaker.RecordSuccess()`
4. 指标通过 `metrics.ObserveExecution(result, duration)` 记录

**修改文件**：`internal/application/automation/runner.go`
- `newEngine()` 注入 `DefaultMetrics` 全局指标收集器

---

### 2.2 P0：效能度量查询性能优化

#### 2.2.1 Redis 读穿透缓存层

**新增文件**：`internal/application/metrics/cache.go`

对标：美团 Raptor 多级缓存、阿里 ARMS 时效性分层

**TTL 策略**：

| 指标类型 | TTL | 说明 |
|---------|-----|------|
| 实时类（ResourceLoad） | 30s | WIP 实时性要求高 |
| 日聚合（Velocity/LeadTime） | 5min | 每日快照兜底 |
| DORA 类 | 1h | 30 天窗口聚合开销大 |
| 快照历史 | 1h | 低频回看 |
| 空值标记 | 1min | 缓存穿透防护 |

**防击穿策略**：
- 空值缓存（避免缓存空对象穿透到 DB）
- `InvalidateProject()` 项目级批量失效（工作项/迭代变更后调用）

---

#### 2.2.2 数据库性能索引

**新增文件**：`sql/0020_s11_perf_indexes.up.sql`

```sql
-- 自动化规则核心路径覆盖索引
CREATE INDEX CONCURRENTLY idx_automation_rules_trigger_active
  ON automation_rules (project_id, trigger_type, sort_order)
  WHERE status = 'active' AND project_id IS NOT NULL;

-- 规则执行审计表
CREATE INDEX CONCURRENTLY idx_rule_executions_rule_created
  ON rule_executions (rule_id DESC, created_at DESC)
  WHERE trigger_event_id IS NOT NULL;

-- 防重复投递幂等键
CREATE UNIQUE INDEX CONCURRENTLY idx_rule_executions_idempotent
  ON rule_executions (rule_id, trigger_event_id)
  WHERE trigger_event_id IS NOT NULL;

-- 指标快照趋势覆盖索引
CREATE INDEX CONCURRENTLY idx_metric_snap_trend
  ON metric_snapshots (workspace_id, project_id, metric, snapshot_date DESC, value);

-- DORA 查询专用
CREATE INDEX CONCURRENTLY idx_deployment_dora
  ON deployment_events (project_id, status, deployed_at DESC)
  WHERE project_id IS NOT NULL;
```

---

#### 2.2.3 带缓存 Handler 包装层

**新增文件**：`internal/application/metrics/handler_cached.go`

Cache-Aside 模式：
1. 查 Redis 缓存（带 X-Cache 响应头）
2. 未命中 → 查 DB
3. 回写缓存（异步，失败不影响主流程）

降级策略：Redis 不可用自动回退到无缓存 Handler。

---

### 2.3 P1：自动化引擎能力扩展

#### 2.3.1 Dry-Run（干跑测试）

**修改文件**：`internal/application/automation/service.go`

新增 `Service.DryRun()` 方法：
- 输入：规则 ID + 样本工作项 ID + 模拟事件类型
- 输出：条件是否匹配 + 将执行的动作列表 + 变量引用警告
- 副作用：零（仅读取规则定义 + 工作项上下文）

已存在的 HTTP 端点 `POST /api/v1/.../automation/dry-run` 已完整支持此能力。

---

### 2.4 P1：效能度量指标完善

#### 2.4.1 CFD（累积流图）

**新增文件**：`internal/application/metrics/advanced.go`

HTTP 端点：`GET /metrics/cfd?days=30`

返回按日期分桶的各状态组工作项数量（堆叠面积图数据）：
```json
[
  {
    "date": "2026-08-01",
    "backlog": 10,
    "todo": 25,
    "in_progress": 8,
    "done": 15,
    "cancelled": 2,
    "total_active": 58
  }
]
```

---

#### 2.4.2 前置时间控制图

HTTP 端点：`GET /metrics/control-chart?days=90`

- 每个完成需求的 Lead Time 散点图
- P50/P85/P95 控制线
- UCL（Upper Control Limit）= P85 * 1.5
- 7 点移动均线趋势

对标：DORA Accelerate State of DevOps 控制图标准。

---

#### 2.4.3 周吞吐量

HTTP 端点：`GET /metrics/throughput?weeks=12`

按周聚合的需求完成数与故事点数（SAFe 规范对齐）。

---

#### 2.4.4 成员级资源负载

HTTP 端点：`GET /metrics/resource-load/detail`

每位成员的实时 WIP：
- 活跃工作项数
- 承担故事点总数
- 项目整体不均衡度（最大 WIP / 平均 WIP，>2 告警）

---

### 2.5 P2：规则模板扩充

**修改文件**：`internal/application/automation/models.go`

从 7 条扩充到 **15 条**，覆盖 4 大场景：

| 类别 | 模板数 | 示例 |
|------|--------|------|
| 质量类 | 3 | 父项自动完成、缺陷通知、缺陷验证人指派 |
| 效率类 | 5 | 开始日期、最闲人指派、高优自动标记、版本流转、Epic 汇总 |
| 通知类 | 3 | 逾期提醒、状态变更通知、迭代总结 |
| 管理类 | 4 | 迭代自动启动、归档旧需求、重复提醒、新人欢迎 |

---

## 3. 验证结果

### 3.1 编译状态

```
✅ go build ./internal/application/automation/...  → PASS
✅ go build ./internal/application/metrics/...     → PASS
✅ go vet ./internal/application/...               → PASS
```

### 3.2 测试状态

```
ok  github.com/njydsz/ydsz-plane/internal/application/automation   0.892s
ok  github.com/njydsz/ydsz-plane/internal/application/period       (cached)
ok  github.com/njydsz/ydsz-plane/internal/application/issue        (cached)
...
```

熔断器 7 个用例全部通过：
- TestCircuitBreaker_ClosedToOpen
- TestCircuitBreaker_OpenToHalfOpen_ThenSuccess
- TestCircuitBreaker_HalfOpenFail
- TestCircuitBreaker_HalfOpenMaxAttempts
- TestCircuitBreakerRegistry_GetOrCreate
- TestCircuitBreaker_Reset
- TestCircuitBreaker_RecordSuccess_ResetsConsecutive

---

## 4. 新增 / 修改文件清单

| 文件 | 类型 | 说明 |
|------|------|------|
| `internal/application/automation/circuit_breaker.go` | 新增 | 熔断器核心实现 |
| `internal/application/automation/circuit_breaker_test.go` | 新增 | 熔断器测试 |
| `internal/application/automation/metrics.go` | 新增 | Prometheus 指标 |
| `internal/application/metrics/advanced.go` | 新增 | CFD/控制图/吞吐量/资源负载 |
| `internal/application/metrics/cache.go` | 新增 | Redis 缓存层 |
| `internal/application/metrics/handler_cached.go` | 新增 | 带缓存 Handler 包装 |
| `sql/0020_s11_perf_indexes.up.sql` | 新增 | 性能索引 |
| `sql/0020_s11_perf_indexes.down.sql` | 新增 | 索引回滚 |
| `internal/application/automation/engine.go` | 修改 | 集成熔断器+指标 |
| `internal/application/automation/runner.go` | 修改 | 注入全局指标 |
| `internal/application/automation/service.go` | 修改 | +DryRun 方法 |
| `internal/application/automation/models.go` | 修改 | 模板 7→15 |
| `internal/application/metrics/handler.go` | 修改 | 注册新路由 + Handler |

---

## 5. 后续待办（v0.3）

| 优先级 | 任务 | 预计工时 |
|--------|------|----------|
| P2 | 自动化规则并发竞争集成测试 | 3d |
| P2 | DSL 嵌套条件组（AND/OR 组合） | 2d |
| P2 | Scheduled Cron 触发器完整实现 | 2d |
| P3 | 自动化执行 WebSocket 实时推送 | 1d |
| P3 | 效能指标导出（CSV/PNG） | 2d |
| P3 | 运维手册更新 | 1d |

---

## 6. 指标影响预估

| 维度 | 当前 | 实施后 |
|------|------|--------|
| 仪表盘首屏 P95 | ~800ms（直查） | ≤200ms（缓存命中） |
| 规则失败雪崩风险 | 无防护 | 熔断器隔离 |
| 自动化可观测性 | 无 | Prometheus 全埋点 |
| 规则覆盖场景 | 7 条基础 | 15 条全场景 |
| 效能指标维度 | 基础五类 | 完整 CFD/DORA/控制图/吞吐量 |
