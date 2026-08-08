# 运维手册

> 面向 SRE / 运维工程师 / 平台管理员
> 配套：[部署架构设计](../architecture/13-部署运维与可靠性设计.md) · [数据模型](../architecture/04-数据模型设计.md) · [API 规范](../architecture/05-API设计规范.md)

目录：

- [部署手册](./deployment.md)
- [升级指南](./upgrade.md)
- [备份与恢复 SOP](./backup-recovery.md)
- [常见告警处置 Runbook](./runbook.md)
- [On-Call 轮值约定](./oncall.md)
- [性能压调优](./performance.md)

---

## 系统组件

| 组件 | 作用 | 依赖 |
|------|------|------|
| API Server（Go + Gin） | 业务 API，WebSocket | PostgreSQL, Redis |
| Worker（Go） | Outbox Relay + 任务消费者 | PostgreSQL, RabbitMQ |
| Web（Vue 3） | SPA 静态资源 | Nginx / CDN |
| PostgreSQL 18 | 主存储 + FTS | — |
| Redis 8 | 缓存 / 限流 / 分布式锁 / WS 扇出 | — |
| RabbitMQ 4 | 事件总线（topic） + 任务队列 | — |
| MinIO（可选）| 对象存储（附件/Logo）| — |

## 关键端口

| 端口 | 服务 | 说明 |
|------|------|------|
| 80 / 443 | Nginx → 前端 + API 反代 | 用户入口 |
| 8080 | API Server | Gin 监听 |
| 5432 | PostgreSQL | 数据库 |
| 6379 | Redis | 缓存 |
| 5672 / 15672 | RabbitMQ | 消息总线 / 管理 UI |
| 9000 / 9001 | MinIO | 对象存储 / Console |
| 8025 | Mailpit | 开发环境邮件测试 |

## 环境变量速查（YDSZ_ 前缀）

| 变量 | 默认 | 说明 |
|------|------|------|
| `YDSZ_DATABASE_URL` | 必填 | PostgreSQL 连接串 |
| `YDSZ_REDIS_ADDR` | `127.0.0.1:6379` | Redis 地址 |
| `YDSZ_RABBITMQ_URL` | `amqp://guest:guest@127.0.0.1:5672/` | AMQP 连接串 |
| `YDSZ_AUTH_JWT_SECRET` | dev 随机 | **生产必填，≥32 字节** |
| `YDSZ_EMAIL_*` | — | SMTP 凭据 |
| `YDSZ_MINIO_*` | — | 对象存储 |
| `YDSZ_LOG_LEVEL` | `info` | zap 日志级别 |
| `YDSZ_BASE_URL` | — | 前端基础 URL（邮件链接拼接）|

> 完整列表见项目根目录 `.env.example`。

## 健康检查

```
GET /healthz        # 进程存活
GET /readyz         # 依赖探活（PG + Redis）
GET /metrics        # Prometheus RED 指标
```

## 默认入口账号

种子账号（仅含种子场景可用）：
- `admin@ydsz.dev` / `Admin@123`

---

> 目录锚点：本手册与 [S12 设计计划 §12.7](../architecture/02-架构开发详细计划.md) 对齐，「部署/运维/升级/Runbook/Oncall」五份子文档构成 v0.2 发布交付物。
