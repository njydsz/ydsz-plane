# 常见告警处置 Runbook

> 配套 [运维手册](./README.md) · [告警体系](../architecture/13-部署运维与可靠性设计.md)

---

## 告警分级

| 级别 | 触达方式 | 响应时限 | 升级条件 |
|------|----------|----------|----------|
| P0 | 电话+IM+短信 | 5 min | 15 min 内未响应 → 升级 oncall-b |
| P1 | IM + 邮件 | 30 min | 1h 未解决 → oncall 介入 |
| P2 | IM | 工作时间内 | 当天解决 |
| P3 | 周报汇总 | 计划内排期 | 下个迭代修复 |

---

## Runbook 索引

1. [API 5xx 突增](#1-api-5xx-突增)
2. [PostgreSQL 连接耗尽](#2-postgresql-连接耗尽)
3. [PostgreSQL 慢查询](#3-postgresql-慢查询)
4. [RabbitMQ 队列堆积](#4-rabbitmq-队列堆积)
5. [Redis 内存溢出](#5-redis-内存溢出)
6. [磁盘空间告警](#6-磁盘空间告警)
7. [Worker 消费停滞](#7-worker-消费停滞)
8. [索引延迟（search lag）](#8-索引延迟search-lag)
9. [ES 黄/红集群（可选 profile）](#9-es-黄红集群可选-profile)
10. [证书即将过期](#10-证书即将过期)

---

## 1. API 5xx 突增

**症状**：`rate(http_requests_total{status=~"5.."}[5m]) > 0.01`

**诊断**：
```bash
# 查看错误分布
docker compose logs api --tail=200 | grep -E "error|ERROR|panic"
curl -s http://localhost:8080/metrics | grep "http_requests_total{status=\"5"

# 检查最近部署
docker compose events --since 30m
```

**常见原因与处置**：
- 最近部署导致：`docker compose up -d --force-recreate` 回滚到上一 tag
- PG 连接池耗尽：检查 PgBouncer / `cfg.Database.MaxConns`
- 资源不足（OOM/free memory）：扩容节点
- 死锁/慢查询：见 [PostgreSQL 慢查询](#3-postgresql-慢查询)

---

## 2. PostgreSQL 连接耗尽

**症状**：`pq: too many clients` / API 503

**诊断**：
```sql
SELECT count(*), state FROM pg_stat_activity GROUP BY state;
SELECT * FROM pg_stat_activity WHERE state = 'idle in transaction';
```

**处置**：
```bash
# 紧急：释放空闲事务
docker compose exec api psql -c "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE state='idle in transaction' AND now()-state_change > interval '5min';"

# 长期：调高 PgBouncer pool_size 或增加 pgx MaxConns
```

---

## 3. PostgreSQL 慢查询

**症状**：API P95 > 200ms；DB CPU 飙升

**诊断**：
```sql
SELECT query, mean_exec_time, calls, total_exec_time
FROM pg_stat_statements
ORDER BY mean_exec_time DESC LIMIT 20;
```

**处置**：
- `CREATE INDEX CONCURRENTLY` 优化缺失索引
- `EXPLAIN (ANALYZE, BUFFERS)` 定位执行计划异常
- 热点查询加 Redis 缓存
- 大表分区（如 `issues` 按 `workspace_id` 分区）

---

## 4. RabbitMQ 队列积压

**症状**：管理 UI 队列深度 >10k；outbox lag >10s

**诊断**：
```bash
# 查看队列深度
docker compose exec rabbitmq rabbitmqctl list_queues name messages consumers

# 查看 consumer 是否存活
docker compose logs worker | grep -E "search consumer|notif consumer|automation consumer"
```

**处置**：
- worker 挂了 → `docker compose restart worker`
- consumer 卡死 → 扩容 worker 副本
- 死信队列堆积 → `rabbitmqctl purge_queue plane.dlx`（需确认）
- 大量事务回滚导致 outbox 重试风暴 → 排查业务异常根因

---

## 5. Redis 内存溢出

**症状**：`OOM command not allowed` / 缓存命中率骤降

**诊断**：
```bash
redis-cli INFO memory
redis-cli --bigkeys
```

**处置**：
- 紧急：`redis-cli CONFIG SET maxmemory-policy allkeys-lru`（若未设）
- 加大内存：增加 Redis 实例规格
- 清理：`redis-cli --scan --pattern "ydsz:cache:*" | xargs redis-cli DEL`（谨慎）

---

## 6. 磁盘空间告警

**症状**：磁盘使用 >80%

**诊断**：
```bash
df -h
du -sh /var/lib/postgresql/data
du -sh /archive   # WAL 归档
```

**处置**：
- WAL 归档堆积：清理过期归档（`find /archive -type f -mtime +7 -delete`）
- 业务表膨胀：`VACUUM FULL`（低峰期）或 pg_repack
- 扩展磁盘

---

## 7. Worker 消费停滞

**症状**：outbox 表持续增长；队列有消息但 consumer=0

**诊断**：
```bash
docker compose logs worker --tail=100
curl -s http://localhost:8080/healthz   # worker 无 /hexlthz，看消息投递是否正常
SELECT count(*) FROM outbox WHERE published=false;
```

**处置**：
```bash
docker compose restart worker
# 若反复出现 → 检查 mq 连接：docker compose logs api | grep -i "rabbitmq\|amqp"
```

---

## 8. 索引延迟（search lag）

**症状**：搜索不到刚创建的 issue

**诊断**：
```bash
docker compose logs worker | grep "search consumer"
# 检查 search_documents 与源表数据差异
docker compose exec api psql -c "SELECT count(*) FROM issues;\nSELECT count(*) FROM search_documents WHERE doc_type='issue';"
```

**处置**：
- 通过运维接口触发回填：`POST /api/v1/admin/search/backfill`（如有）
- 或直接：`go run ./cmd/worker -backfill-only`
- 修复消费者异常、观察 lag

---

## 9. ES 黄/红集群（可选 profile）

**症状**：`health.status != green`

**诊断**：
```bash
curl -s http://localhost:9200/_cat/health
curl -s http://localhost:9200/_cat/shards?h=index,shard,state
```

**处置**：
- 黄（未分配副本）：扩容 data 节点 或 调低副本数
- 红（主分片丢失）：恢复快照；ES 未启用则依赖 PG 重建

> 注：[当前默认部署使用 PG FTS](./deployment.md)，ES 仅作为 `full profile` 可选启用；默认场景此 Runbook 不触发。

---

## 10. 证书即将过期

**症状**：`cert_expiry_days < 14`

**处置**：
```bash
# acme/letsencrypt
certbot renew --force-renew

# 自签证书：提前 30 天重签，Nginx 热加载
nginx -s reload
```

---

## 黑屏处置原则

1. **先看告警**：Prometheus Alertmanager / 短信 / 邮件
2. **看日志**：`docker compose logs --tail=100`
3. **看指标**：Grafana 相关面板
4. **快速止血**：扩容 / 重启 / 切换流量
5. **复盘**：事后出 incident report，更新本 Runbook

---

## 常用排查命令速查

```bash
# 整体状态
docker compose ps
docker system df

# 实时日志（key word）
dockercompose logs -f --tail=0 | grep -iE "error|panic|fatal"

# PG 活动会话
docker compose exec api psql -c "SELECT pid,usename,query_start, state, LEFT(query,80) FROM pg_stat_activity WHERE state<>'idle' ORDER BY query_start;"

# RabbitMQ 队列
docker compose exec rabbitmq rabbitmqctl list_queues name messages consumers memory

# Redis
docker compose exec redis redis-cli ping
```
