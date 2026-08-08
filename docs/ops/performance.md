# 性能压测与调优

> 配套 [运维手册](./README.md) · [性能基线](../architecture/17-性能基线.md) · [压测报告](../perf/README.md)

---

## 1. 性能目标（SLO）

| 指标 | 目标 |
|------|------|
| API 可用性 | ≥99.9%/30d |
| API 响应 P95 | ≤200ms |
| 页面加载 P95 | ≤2s |
| 事件投递延迟 P95（DB → worker） | ≤5s |
| 搜索索引延迟 P95 | ≤10s |
| 并发用户数 | ≥1000 |
| 单项目工作项 | ≥100 万 |

压测入口：`make perf-smoke` / `make perf-load` / `make perf-stress`（基于 k6）。

---

## 2. 压测场景

### 2.1 冒烟（50 VU，5 分钟）
验证 CI 环境下核心接口基线不退化。

### 2.2 负载（500 VU，30 分钟）
工作日平均负荷模拟。

### 2.3 压力（1000 VU，10 分钟）
峰值流量验收。

| 场景 | 脚本 |
|------|------|
| 登录 | `scripts/k6/smoke.js` |
| 看板渲染 | `scripts/k6/load.js` |
| 工作项详情 | `scripts/k6/stress.js` |
| 全局搜索 | `scripts/k6/search.js` |
| 创建工作项 | `scripts/k6/create.js` |

---

## 3. 压测执行规范

```bash
# 前置：准备 100 万工作项数据
make seed-scale

# 冒烟
make perf-smoke

# 负载
make perf-load
```

**禁止事项**：
- 禁止在 Prod 环境直接压测（隔离 staging / perf 环境）
- 禁止无监控窗口下长时间高 VU 压测（每次 ≤30 分钟）

---

## 4. 压测指标采集

### 4.1 后端 RED
- request_rate / error_rate / duration（P50/P95/P99）

### 4.2 依赖资源
- PG：`pg_stat_statements`、`pg_stat_activity`、连接池使用率
- Redis：命中/miss、命令吞吐
- RabbitMQ：queue depth、ack rate、consumer utilisation

### 4.3 主机层
- CPU、内存、网络 IO、磁盘 IOPS（node_exporter）

### 压制报告归档

每次全量压测后，报告归档 `docs/perf/YYYY-MM-DD-<scenario>.md`：
- 场景与 VU
- 关键指标表（含与基线对比）
- 瓶颈定位
- 改进建议

---

## 5. 常见瓶颈与调优

### 5.1 PG 连接池耗尽
- 现象：API 503 / `too many clients`
- 调优：`MaxConns`、PgBouncer 事务级 pooling、慢查询优化

### 5.2 看板 N+1 查询
- 现象：单接口多次 DB 查询
- 调优：批量预加载（`select ... where id = any($1)`），DTO 聚合到 Service 层

### 5.3 全文搜索慢
- 现象：search P95 >2s
- 调优：GIN 索引重建（`REINDEX`），`search_tsv` 列统计信息更新（`ANALYZE search_documents`）

### 5.4 附件下载带宽
- 现象：下载大对象时延迟飙升
- 调优：CDN + 预签名 URL 直接到 MinIO，不经过 API 反代

### 5.5 Worker 消费瓶颈
- 现象：outbox lag >10s
- 调优：增加 worker 副本；拆分队列粒度；批量消费（prefetch）

### 5.6 RabbitMQ 内存/磁盘告警
- 现象：内存高位 + paging
- 调优：`x-max-length` 限深；`delivery_mode=1`（非持久）针对实时任务；提高 consumer prefetch

---

## 6. 容量冗余建议

| 资源 | 基线水位 | 扩容触发 |
|------|----------|----------|
| PG CPU | <50% | >60% 持续 1 周 |
| PG 磁盘 | <60% | >80% |
| Redis 内存 | <50% | >70% |
| RabbitMQ 队列深 | <1000 | >10000 |
| API Pod CPU | <50% | >70% |
| Worker CPU | <50% | >70% |

---

## 7. 性能门禁（CI）

`make perf-smoke` 在 CI 中运行，任一场景 P95 >200ms 则 PR 合入失败（perf 门禁）。

压测回归周期：
- 每日：自动化冒烟（CI）
- 每周：负载（perf 环境）
- 每 Release 前：全量三件套
