# Ydsz Plane v0.2 — 运维手册

> 面向 SRE / DevOps 的部署、监控、备份、升级与故障排查手册。
> 参考：Google SRE Book、美团运维手册规范、等保三级运维基线。

---

## 1. 部署架构

```
┌──────────────────────────────────────────────────────┐
│                    nginx / CDN                        │
│              TLS 1.2+ / WAF / Rate Limit              │
└─────────────┬────────────────────────────┬───────────┘
              │                            │
       ┌──────▼──────┐              ┌──────▼──────┐
       │  API Server  │              │  Worker     │
       │  (Go/Gin)    │              │  (Go)       │
       │  :8080       │              │             │
       └──┬───┬───┬──┘              └──────┬──────┘
          │   │   │                        │
    ┌─────┘   │   └─────┐           ┌──────┘
    │         │         │           │
┌───▼──┐  ┌──▼───┐  ┌──▼────┐  ┌───▼────┐
│ PG 18│  │Redis8│  │Rabbit │  │ MinIO  │
│      │  │      │  │ MQ 4  │  │ (可选) │
└──────┘  └──────┘  └───────┘  └────────┘
```

### 1.1 最小服务器配置
| 组件 | CPU | 内存 | 磁盘 |
|------|-----|------|------|
| API Server | 2 | 2GB | - |
| Worker | 1 | 1GB | - |
| PostgreSQL | 2 | 4GB | 100GB SSD |
| Redis | 1 | 1GB | - |
| RabbitMQ | 1 | 2GB | 20GB |

### 1.2 Docker Compose 部署（最快）
```bash
git clone https://github.com/njydsz/ydsz-plane.git
cd ydsz-plane
cp .env.example .env   # 修改密钥
docker compose up -d
# 浏览器打开 http://localhost:8080
```

---

## 2. 数据库管理

### 2.1 迁移
```bash
# 应用全部迁移
go run ./cmd/migrate up

# 回滚一步
go run ./cmd/migrate down 1

# 查看版本
go run ./cmd/migrate version
```

迁移文件位于 `sql/` 目录，命名格式 `{version}_{name}.{up|down}.sql`，由 golang-migrate 执行。**只能向前回滚**，不支持跳版本。

### 2.2 备份策略
```
# 每日全量（凌晨 03:00）
pg_dump -Fc $YDSZ_DATABASE_URL > backup_$(date +%Y%m%d).pgdump

# WAL 持续归档（ postgresql.conf ）
archive_mode = on
archive_command = 'cp %p /wal_archive/%f'

# 保留周期
全量备份：30 天
WAL 归档：7 天
```

### 2.3 恢复演练（每季度）
```bash
# 1. 还原全量备份
pg_restore -d ydsz_plane_restore backup_20260807.pgdump

# 2. 应用后续 WAL（如需要指定时间点恢复）
# 执行 recovery_target_time = '2026-08-07 10:00:00'

# 3. 验证：启动测试容器 + 登录 + 抽查数据
```

### 2.4 索引维护
新增 0019_perf_indexes migration 后（CREATE INDEX CONCURRENTLY 在线不加锁），建议：
```sql
-- 查看索引大小与使用率
SELECT indexrelname, pg_size_pretty(pg_relation_size(indexrelid)),
       idx_scan, idx_tup_read, idx_tup_fetch
FROM pg_stat_user_indexes WHERE schemaname = 'public'
ORDER BY pg_relation_size(indexrelid) DESC;
```

### 2.5 VACUUM / ANALYZE
已配置 `autovacuum` 默认值。高写入场景（日均百万行）：
```sql
ALTER SYSTEM SET autovacuum_vacuum_scale_factor = 0.05;
ALTER SYSTEM SET autovacuum_analyze_scale_factor = 0.025;
SELECT pg_reload_conf();
```

---

## 3. 监控与告警

### 3.1 Prometheus 指标
- Endpoint：`GET /metrics`
- 命名空间：`ydsz_*`
- 核心指标：
  - `ydsz_http_requests_total{method,route,status}` — RED Rate
  - `ydsz_http_request_duration_seconds` — RED Duration
  - `ydsz_db_query_duration_seconds` — 数据库查询耗时
  - `ydsz_worker_jobs_total` — Worker 处理量

### 3.2 Grafana Dashboard
提供了 `deploy/grafana/dashboard.json`，包含：
- 每秒请求量、P95 延迟、错误率
- DB 连接池占用、慢查询
- Worker 队列堆积

### 3.3 健康检查
```
GET /health      → 200 OK（应用存活）
GET /health/db   → 200 OK / 503（PG 连通）
GET /health/redis → 200 OK / 503（Redis 连通）
```

### 3.4 告警规则示例
```yaml
# API P95 延迟 > 500ms 持续 5 分钟
- alert: YdszHighLatency
  expr: histogram_quantile(0.95, rate(ydsz_http_request_duration_seconds_bucket[5m])) > 0.5
  for: 5m
  severity: warning

# 错误率 > 5% 持续 2 分钟
- alert: YdszHighErrorRate
  expr: rate(ydsz_http_requests_total{status=~"5.."}[2m]) / rate(ydsz_http_requests_total[2m]) > 0.05
  for: 2m
  severity: critical
```

---

## 4. 日志管理

### 4.1 日志格式
结构化 JSON 日志（stdout），可由 Fluent Bit / Filebeat 采集。

### 4.2 关键字段
```json
{
  "ts": "2026-08-07T10:00:00Z",
  "level": "info",
  "caller": "httpapi/router.go:42",
  "msg": "request completed",
  "status": 200,
  "latency_ms": 12.3,
  "method": "GET",
  "path": "/api/v1/workspaces/1/projects",
  "workspace_id": 1,
  "user_id": "abc-123"
}
```

### 4.3 敏感信息脱敏
密码、Token、email 等字段不会出现在日志中。脱敏规则在 `internal/infrastructure/logger/` 中维护。

---

## 5. 故障排查

### 5.1 常见问题速查

| 症状 | 诊断 | 处置 |
|------|------|------|
| 502 Bad Gateway | Check API 存活：`curl /health` | 重启 API |
| DB 连接池耗尽 | `ydsz_db_pool_active / max` | 调大 `Database.MaxConns` |
| 限流 429 | 查看 `X-Remaining` header | 提升限流上限或联系管理员 |
| Outbox 堆积 | RabbitMQ queue length 监控 | 检查 Worker 存活 + Outbox relay |
| 登录失败率高 | Redis 计数异常 | 检查 Redis TTL |

### 5.2 数据库慢查询排查
```sql
-- 查看慢 queries (> 1s)
SELECT calls, mean_time, query
FROM pg_stat_statements
WHERE mean_time > 1000
ORDER BY mean_time DESC
LIMIT 20;
```

### 5.3 应急联系人表
| 角色 | 联系方式 | 职责 |
|------|---------|------|
| SRE on-call | `:pager:rod://...` | 基础设施故障 |
| DBA | `:mail:dba@...` | 数据库级故障 |
| 业务负责人 | `:mail:pm@...` | 业务逻辑 Bug 评估 |

---

## 6. 安全加固速查

- ✅ 登录失败锁定（5 次 / 15 分钟）
- ✅ Redis 令牌桶限流（API 级）
- ✅ CSP + HSTS + X-Frame-Options 安全头
- ✅ RLS 租户隔离（数据库层）
- ✅ 审计日志全量覆盖
- ⚠️ 审计日志按月分区（待 v0.3 实施）
- ⚠️ SSRF egress 代理（待 v0.3 实施）

**定期任务**：
- 每周：`pg_dump` 全量备份验证
- 每月：`VACUUM ANALYZE` + 索引 rebuild
- 每季度：恢复演练
- 每年：安全复测（等保测评）

---

## 7. 升级流程

详见 [upgrade-guide.md](./upgrade-guide.md)。

---

## 8. 容量规划

| 用户规模 | 配置规格 | 说明 |
|---------|---------|------|
| ≤ 100 | 最小部署 | 单 PG 实例即可 |
| 100-1000 | 2 API + PG 主从 | Redis 独立部署 |
| 1000-5000 | K8s + PG 读写分离 | ES 全文检索独立集群 |
| 5000+ | 多区域 + ShardingSphere | 见架构文档 |

PostgreSQL 单项目 1M 工作项量级已验证（P95 < 200ms，施加 0019_perf_indexes 后）。
