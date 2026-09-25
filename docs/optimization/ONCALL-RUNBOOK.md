# On-call Runbook（P2-8）

> 适用角色：运维工程师 / On-call  
> 目标：快速定位并处置常见故障，降低 MTTR  
> 最后更新：2025-07-10

---

## 1. Outbox 堆积排查

### 症状

- 事件未被消费，`outbox_events` 表持续增长
- 下游系统数据不一致（如搜索结果滞后、缓存未更新）

### 检查步骤

```sql
-- 查看堆积数量（超过 5 分钟未发布的事件）
SELECT COUNT(*)
FROM outbox_events
WHERE published_at IS NULL
  AND created_at < NOW() - INTERVAL '5 min';
```

进一步排查：

```sql
-- 按事件类型分布查看
SELECT event_type, COUNT(*)
FROM outbox_events
WHERE published_at IS NULL
GROUP BY event_type;
```

```bash
# 检查 worker 进程状态
kubectl get pods -l app=plane-worker

# 查看 worker 最近日志
kubectl logs -l app=plane-worker --tail=100 --since=10m
```

### 常见原因

| 原因 | 特征 |
|------|------|
| RabbitMQ 连接中断 | worker 日志出现 `connection closed` / `amqp: connection reset` |
| Relay 协程 panic | 日志中出现 stack trace 或 `goroutine panic` |
| Worker 进程 down | 所有 worker Pod 处于 CrashLoopBackOff 或已退出 |
| 事件发布逻辑异常 | 单条特定事件反复发布失败，阻塞后续事件 |

### 修复步骤

1. 滚动重启 worker：`kubectl rollout restart deployment/plane-worker`
2. 检查 RabbitMQ 连接：确认 Broker endpoint 可达，无网络抖动
3. 查看应用日志：关注 error/warn 级别日志中 `outbox` 或 `relay` 相关内容
4. 若单条事件阻塞：手动将该事件标记为已发布（`UPDATE outbox_events SET published_at = NOW() WHERE id = '<stuck_id>'`），再分析失败原因

### Prometheus 告警

```promql
plane_outbox_lag_seconds > 30
```

告警含义：最近一条成功发布的 outbox 事件与当前时间差超过 30 秒。

---

## 2. ES 与 DB 不一致修复

### 症状

- 搜索结果中缺失或多出工作项
- DB 中存在的记录在 ES 搜索中查不到（或反之）

### 检查步骤

```bash
# 对比 DB 与 ES 的 work_items 数量
curl -s http://es-cluster:9200/plane_work_items/_count | jq '.count'

# 对比 DB count
# 通过应用 API 或直连数据库查询 SELECT COUNT(*) FROM issues
```

```sql
-- 检查 outbox_events 中 search 相关事件堆积
SELECT event_type, COUNT(*)
FROM outbox_events
WHERE published_at IS NULL
  AND event_type LIKE '%search%'
GROUP BY event_type;
```

```bash
# 检查 ES 集群健康状态
curl -s http://es-cluster:9200/_cluster/health | jq '{status, number_of_nodes, unassigned_shards}'

# 检查 ES 磁盘水位
curl -s http://es-cluster:9200/_cat/allocation?v
```

### 修复步骤

1. 触发全量 Reindex API：

```bash
curl -X POST "http://plane-api:8000/internal/v1/reindex" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json"
```

2. 检查 ES 集群状态：若 status 为 `red`，排查丢失分片或下线节点
3. 检查磁盘：磁盘使用率超过 85%（high watermark）时 ES 会变为 read-only，需扩容或清理旧索引

### 预防措施

- 定期（每日）对比 `outbox_events` 中 `search.index` 事件的消费延迟，纳入监控大盘
- 设置告警：`plane_outbox_event_age{type="search.index"} > 60s`

---

## 3. RabbitMQ 连接异常处置

### 症状

- 消息堆积：Ready 消息数持续上涨，消费者进度停滞
- `publisher-confirm` 超时：消息发布方日志中出现 `confirm timeout`
- Outbox Relay 出现断断续续的失败

### 检查步骤

```bash
# 检查 RabbitMQ 节点状态
rabbitmq-diagnostics status | head -30

# 检查连接数和通道数
rabbitmq-diagnostics connections | wc -l
rabbitmq-diagnostics channels | wc -l
```

```bash
# 查看 connection recovery 关键日志
kubectl logs -l app=plane-worker --since=30m | grep -iE "reconnect|connection.*closed|amqp.*error"
```

```bash
# 检查当前 MQ 队列状态
rabbitmqctl list_queues name messages_ready messages_unacked consumers
```

### 修复步骤

1. 检查网络连通性：确认 worker Pod 到 RabbitMQ endpoint 的端口无防火墙阻断
2. 确认 DNS 解析正常（若使用域名访问 RabbitMQ）
3. 重启 MQ connection watchdog：滚动重启 worker 进程以重置连接管理器
4. 若集群不健康：按 RabbitMQ 运维手册排查节点分区/脑裂，必要时强制清理故障节点

### 关键代码路径

```
internal/infrastructure/mq/rabbitmq.go  -> reconnectionWatchdog()
```

该 Watchdog 负责检测连接丢失并自动执行退避重连。若日志显示 watchdog 本身 panic，需优先修复代码逻辑。

---

## 4. Worker DLQ 告警处理

### 症状

- 消费者处理失败超过重试上限，消息进入 Dead Letter Queue
- DLQ 队列 `plane.events.dlq` 出现堆积

### 检查步骤

```bash
# 查看 DLQ 队列中的消息数
rabbitmqctl list_queues name messages_ready | grep dlq

# 读取 DLQ 消息内容（取样）
rabbitmqctl get queue=plane.events.dlq count=10 ackmode=ack_requeue_true
```

```bash
# 查看 worker 消费失败的日志
kubectl logs -l app=plane-worker --since=15m | grep -iE "dlq|dead.letter|max.retries|consume.*failed"
```

```bash
# 检查 DLQ 服务进程状态
kubectl get pods -l app=dlq-processor
```

### 修复步骤

1. **定位失败原因**：根据错误日志判断是瞬时可恢复（如下游超时）还是需人工介入（如数据损坏）
2. **重新投递**：若确认原始消息可正常处理，重新将消息投递至主队列：

```bash
rabbitmqadmin publish exchange=<dlq_exchange> routing_key=<main_queue> payload_file=<msg.json>
```

3. **跳过并 ACK**：若消息内容本身有问题（如非法 payload），记录问题并在 DLQ 中消费掉（ACK）
4. 修复根因后，确认 DLQ 堆积量持续下降

### 关键代码路径

```
internal/application/dlq/         -- DLQ 消息处理逻辑
```

---

## 附录：告警快速索引

| 告警名称 | PromQL 表达式 | 优先级 |
|----------|--------------|--------|
| Outbox 堆积 | `plane_outbox_lag_seconds > 30` | P2 |
| Outbox 事件年龄 | `plane_outbox_event_age{type="search.index"} > 60` | P2 |
| DLQ 堆积 | `rabbitmq_queue_messages{queue=~"*.dlq"} > 0` | P1 |
| MQ 连接异常 | `plane_mq_connection_status != 1` | P1 |
| ES 集群状态 | `elasticsearch_cluster_health_status{color="red"}` | P1 |

> 联系方式：遇到无法处理的告警，执行 escalate 流程 — 联系后端值班负责人（#on-call-backend）。
