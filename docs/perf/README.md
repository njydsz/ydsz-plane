# 性能基线与压测

> S7 P0-8 — 建立 10 万工作项级别的性能基线。

本目录汇总压测脚本、基线记录和运行指南。与架构文档 `docs/architecture/17-性能基线.md` 互为补充。

---

## 目录结构

```
docs/perf/
  README.md           # 本文档 — 运行指南与基线数据

scripts/seed/
  main.go             # 种子脚本入口（确定性 seed + 批量造数）
  bulk.go             # 10 万工作项批量造数核心逻辑
  helper.go           # DSN 加载与脱敏辅助函数

tests/perf/
  smoke-test.js       # 冒烟测试：验证核心 API 端点
  load-test.js        # 负载测试：渐进 VU（10→100），断言 P95<200ms
  stress-test.js      # 压力测试：恒定 200 VU，找拐点
```

---

## 快速开始

### 前置条件

| 组件 | 版本 | 用途 |
|------|------|------|
| Go | 1.26+ | 运行造数脚本 |
| k6 | v0.47+ | 负载测试引擎 |
| PostgreSQL | 14+ | 目标数据库 |
| Plane API | 已编译并启动 | 被压测服务 |

### 安装 k6

```bash
# macOS / Linux
brew install k6

# Windows (winget)
winget install k6

# Docker
docker pull grafana/k6
```

### 1. 造数（10 万工作项）

```bash
# 默认模式（读取 YDSZ_DATABASE_URL 或项目默认连接串）
go run ./scripts/seed --count 100000

# 自定义参数
go run ./scripts/seed \
  --count 100000 \
  --batch-size 2000 \
  --db-dsn "postgres://postgres:Limw1020@127.0.0.1:5432/ydsz-plane_test?sslmode=disable"
```

脚本会自动创建：
- 工作空间 `perf-test-ws`
- 项目 `perf-test-project`
- 4 个状态（待处理 / 进行中 / 已完成 / 已取消）
- N 条工作项（默认 10 万），批量事务提交，目标 30 秒内完成

输出示例：

```
[bulk] 连接数据库: ***@127.0.0.1:5432/ydsz-plane_test
[bulk] 开始造数: count=100000, batch-size=2000
  已插入 10000 / 100000 (10%)
  已插入 20000 / 100000 (20%)
  ...
  已插入 100000 / 100000 (100%)
═══════════════════════════════════════════════
  造数完成: 100000 / 100000
  耗时: 22.3s
  速率: 4484 行/秒
  workspace_id=9 project_id=9
═══════════════════════════════════════════════
```

### 2. 运行压测

确保 API 服务先启动，且工作项数据已造好。

```bash
# 冒烟测试
k6 run tests/perf/smoke-test.js

# 负载测试（渐进 VU）
k6 run tests/perf/load-test.js

# 压力测试（恒定 200 VU）
k6 run tests/perf/stress-test.js
```

指定目标环境：

```bash
k6 run -e BASE_URL=http://staging.example.com/api/v1 \
       -e TEST_USER_EMAIL=perf@test.local \
       -e TEST_USER_PASSWORD=xxx \
       tests/perf/load-test.js
```

---

## 命令行参数

### `scripts/seed` — 造数脚本

| Flag | 默认值 | 说明 |
|------|--------|------|
| `--count N` | `0`（运行确定性 seed） | 造 N 条工作项；0 时退回到原有 seed 模式 |
| `--batch-size B` | `1000` | 每批 INSERT 的行数（事务粒度） |
| `--db-dsn string` | 读 `YDSZ_DATABASE_URL` 或配置文件默认值 | PostgreSQL 连接串 |

### `tests/perf/*.js` — k6 脚本

| 环境变量 | 默认值 | 说明 |
|----------|--------|------|
| `BASE_URL` | `http://127.0.0.1:8080/api/v1` | API 基础地址 |
| `TEST_USER_EMAIL` | `admin@ydsz.dev` | 登录用邮箱 |
| `TEST_USER_PASSWORD` | `Admin@123` | 登录用密码 |
| `STRESS_VUS` | `200` | 压力测试并发 VU |
| `STRESS_DURATION` | `3m` | 压力测试持续时间 |

---

## 测试覆盖端点

| 端点 | smoke | load | stress |
|------|:-----:|:----:|:------:|
| `POST /auth/login` | ✅ setup | ✅ setup | ✅ setup |
| `GET /workspaces` | ✅ | ✅ 10% | — |
| `GET /me` | ✅ | — | — |
| `GET /workspaces/:wsId/projects/:projectId/issues` | ✅ | ✅ 60% | ✅ 50% |
| `GET /workspaces/:wsId/projects/:projectId/issues/:id` | ✅ | ✅ 20% | ✅ 30% |
| `GET /workspaces/:wsId/workbench/summary` | ✅ | ✅ 10% | ✅ 20% |

---

## 基线指标记录

### 空库基线（0 工作项）

| 日期 | 场景 | 端点 | VUs | P50 | P95 | P99 | QPS | 错误率 | 备注 |
|------|------|------|-----|-----|-----|-----|-----|--------|------|
| 待实测 | smoke | `/issues` | 1 | — | — | — | — | — | 空库无数据，首次查询触发冷缓存 |
| 待实测 | load | `/issues` | 10→100 | — | — | — | — | — | 空库，JOIN 仅扫空结果集 |
| 待实测 | stress | `/issues` | 200 | — | — | — | — | — | 空库压力测试 |

### 10 万工作项基线

| 日期 | 场景 | 端点 | VUs | P50 | P95 | P99 | QPS | 错误率 | 备注 |
|------|------|------|-----|-----|-----|-----|-----|--------|------|
| 待实测 | smoke | `/issues?limit=20` | 1 | — | — | — | — | — | 10 万行表冷查询 |
| 待实测 | load | `/issues?limit=20` | 100 | — | — | — | — | — | 热缓存 + LIMIT 分页 |
| 待实测 | load | `/issues/:id` | 100 | — | — | — | — | — | 主键点查 |
| 待实测 | load | `/workbench/summary` | 100 | — | — | — | — | — | 聚合查询 |
| 待实测 | stress | 混合 | 200 | — | — | — | — | — | 3 分钟恒定压力 |

> 填写说明：每次里程碑前在 staging 运行全套并回填实际数字，差异 >10% 触发 Review。

---

## 已知限制

1. **造数脚本与 RLS**：造数脚本通过 `set_config('app.workspace_id', ...)` 设置租户上下文以绕过行级安全策略，需使用 SUPERUSER 或表所有者连接。
2. **workbench/summary 端点**：若数据为空，该端点可能返回 404 而非 200，k6 脚本中已处理此情况。
3. **issueId 随机性**：为了模拟真实场景，负载 / 压力测试使用 `base_issueId + random(N)` 构造 ID，部分请求预期会 404，不计入错误率。
4. **benchmark 环境一致性**：压测结果对硬件、PostgreSQL `shared_buffers`、连接池大小高度敏感；建议在同一台机器、同一 DB 配置下跨版本比较。
5. **确定性 seed 与批量造数的隔离**：批量造数使用独立 workspace（`perf-test-ws`），不干扰开发用测试账号（`admin@ydsz.dev` 等）。
