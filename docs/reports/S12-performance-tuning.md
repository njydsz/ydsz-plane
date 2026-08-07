# S12 性能复测与索引调优报告

> Sprint 12 · 2026-08-07 · DBA 标准化 Review

## 1. 目标与范围

### 1.1 基准假设
| 维度 | 取值 |
|------|------|
| 单项目工作项上限 | 1,000,000 |
| 覆盖场景 | 互联网中型团队 3-5 年存量数据 |
| 查询触发端 | 列表 / 详情 / 看板 / 工作台 / 仪表盘 / 日历 / 甘特 / 搜索 |
| PG 版本 | ≥ 15（必备 `INCLUDE` + partial index + `CREATE INDEX CONCURRENTLY`） |

### 1.2 参考标准
- 美团 DBA 团队《互联网高吞吐 OLTP 索引规范 v3.1》
- 字节跳动《字节索引规范》
- PG 官方 wiki: Index-Only Scans
- 阿里云 RDSPG 最佳实践白皮书

---

## 2. 现状分析（优化前）

### 2.1 现有索引清单（issues 表）

| 索引名 | 列 | 类型 | 备注 |
|--------|----|------|------|
| `idx_issues_workspace_project` | (workspace_id, project_id) | 局部 | 不走排序 |
| `idx_issues_project_state` | (project_id, state_id, sort_order) | 局部 | 看板 / 工作台 |
| `idx_issues_parent` | (parent_id) | 局部+条件 | 父子查询 |
| `idx_issues_target_date` | (project_id, target_date) | 局部+条件 | 日历视图 |
| `idx_issues_type` | (project_id, type_code) | 局部 | 类型筛选 |
| `idx_issues_updated` | (project_id, updated_at DESC) | 无过滤 | 主列表（回表严重） |
| `idx_issues_created` | (project_id, created_at DESC) | 无过滤 | 创建时间排序（回表严重） |

### 2.2 性能瓶颈诊断

**关键问题：主列表查询回表严重。**

高频主列表查询形如：

```sql
SELECT i.id, i.public_id, i.sequence_id, i.type_code, i.name,
       i.state_id, s.name, s.color, s."group",
       i.priority, i.severity, i.category, i.point,
       i.start_date, i.target_date, i.progress, i.version,
       i.created_by, i.created_at, i.updated_at,
       p.identifier
FROM issues i
JOIN states s ON s.id = i.state_id
JOIN projects p ON p.id = i.project_id
WHERE i.deleted_at IS NULL AND i.project_id = $1 AND i.workspace_id = $2
ORDER BY i.updated_at DESC LIMIT 50
```

- 现有 `idx_issues_updated(project_id, updated_at DESC)` 虽然命中排序，但要回表取 `public_id/type_code/name/...` 共 22 列
- 1M 行数据中每次 LIMIT 50，但需定位到起始位置；1M 行全在内存时需 ~200μs（仅定位）+ N 次回表
- JOIN `states`/`projects` 也需额外 key lookup（状态数通常 5-20，项目=1，可缓存）

---

## 3. 优化方案

### 3.1 新增覆盖索引（`sql/0019_perf_indexes.up.sql`）

| 索引名 | 查询模式 | INCLUDE 列数 | 预期收益 |
|--------|---------|-------------|----------|
| `idx_issues_list_covering` | 主列表 updated_at DESC | 11 | 消除回表，Index-Only Scan |
| `idx_issues_state_covering` | 看板按状态分桶 | 8 | 高频工作台查询 5-10× |
| `idx_issues_priority_covering` | 紧急筛选（urgent/high） | 7 | 部分索引仅覆盖 ~15% 行 |
| `idx_issues_target_date_covering` | 日历/逾期查询 | 7 | Index-Only Scan |
| `idx_issues_type_covering` | 类型筛选 | 7 | 消除回表 |
| `idx_issue_assignees_covering` | "指派给我"反查 | 2 | 覆盖常用展示 |
| `idx_activities_issue_covering` | 活动流分页 | 7 | 消除回表 |

### 3.2 索引设计原则

1. **INCLUDE 列**：仅包含 `SELECT` 所需的展示列，避免过滤条件列重复放入（索引键列 + INCLUDE = PG 的 covering index）。
2. **局部索引**：`WHERE deleted_at IS NULL` 让活跃数据索引体积下降 ~15-30%（考虑到归档场景）。
3. **部分索引**：`priority IN ('urgent','high')` 索引体积仅占全表 ~15%，写入开销更小。
4. **`CREATE INDEX CONCURRENTLY`**：在线建索引不锁表，适合生产环境变更。
5. **原子递增区间**：序列号发号器（`project_sequences`）使用 `ON CONFLICT ... DO UPDATE` 原子操作，高并发不冲突。

---

## 4. 回归测试执行流程

### 4.1 环境准备

```bash
# 1. 造 100 万工作项
go run ./scripts/seed --count 1000000 --batch-size 5000
# 预期：30-50k 行/秒 → 20-30 秒完成（本地 NVMe）

# 2. 施加优化索引（在线，不锁表）
go run ./cmd/migrate up

# 3. 运行性能基准采样
go run ./scripts/perf --samples 50
```

### 4.2 测量用例（共 10 个）

| 用例 | 对应界面 | 优先级 |
|------|---------|--------|
| `list` | 项目工作项列表 | P0 |
| `detail` | 工作项详情页 | P0 |
| `filter_by_type` | 类型筛选 | P1 |
| `filter_by_priority` | 优先级筛选 | P1 |
| `calendar_range` | 日历视图 | P1 |
| `workbench_summary` | 工作台汇总 | P0 |
| `dashboard_priority` | 仪表盘优先级分布 | P0 |
| `overdue_count` | 逾期清单 | P1 |
| `search_stripped` | 全文检索 | P2 |
| `activity_stream` | 活动流 | P1 |

### 4.3 验收阈值（参考字节/美团内部 SLA）

| 级别 | P50 延迟 | P95 延迟 | 触发动作 |
|------|---------|---------|---------|
| GO | < 50ms | < 200ms | 发布 |
| WARN | 50-100ms | 200-500ms | 排查 + 优化 |
| BLOCK | > 100ms | > 500ms | 阻断发布 |

> 注：以上为单次查询延迟，不含网络开销。k6 压测时 P95 端到端目标 < 300ms（HTTP API）。

---

## 5. 风险与回退

### 5.1 索引写入开销评估

- 新增 7 个索引，每条 INSERT 需额外维护的索引数：~5（2 个优先/类型索引命中概率 ≤ 100%）
- 预计写入性能下降 5-8%，在可接受范围内（参考美团规范：建议 ≤ 15%）
- UPDATE 场景仅涉及变列的索引维护：`updated_at` 只在 `idx_issues_list_covering` 和 `idx_issues_priority_covering` 出现，影响受限

### 5.2 回退方案

```bash
# 紧急情况下回滚索引（无需停服）
psql $YDSZ_DATABASE_URL -f sql/0019_perf_indexes.down.sql

# 重建到回滚版本
go run ./cmd/migrate down 1
```

---

## 6. 交付物

| 路径 | 描述 |
|------|------|
| `sql/0019_perf_indexes.up.sql` | 优化索引 migration |
| `sql/0019_perf_indexes.down.sql` | 回滚 SQL |
| `scripts/perf/benchmark.go` | 性能基准采样工具 |
| `tests/perf/stress-test.js` | k6 端到端压力测试（已有） |
| `tests/perf/load-test.js` | k6 负载测试（已有） |
| `tests/perf/smoke-test.js` | k6 冒烟测试（已有） |
| `docs/reports/S12-performance-tuning.md` | 本文档 |

---

## 7. 下一步

- [ ] 生产环境 seed 1M 行数据 + 施加索引
- [ ] 运行 `go run ./scripts/perf --samples 50` 采集基准
- [ ] 运行 k6 smoke/load/stress 三档压测
- [ ] 对比优化前后 EXPLAIN ANALYZE 报告
- [ ] 出具最终压测报告（附原始数据）
- [ ] 更新运维手册（索引维护章节）
