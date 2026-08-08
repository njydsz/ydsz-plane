# 部署手册

> 配套 [运维手册](./README.md) · [部署架构设计](../architecture/13-部署运维与可靠性设计.md)

---

## 1. 部署形态

### 1.1 Docker Compose（推荐）

最快上手方式，适合 10–1000 人团队。

```bash
git clone https://github.com/njydsz/ydsz-plane.git
cd ydsz-plane

# 1. 配置环境变量
cp .env.example .env
# 必须修改：YDSZ_AUTH_JWT_SECRET（≥32 字节）、YDSZ_BASE_URL

# 2. 一键启动核心栈（API + Worker + Web + PG + Redis + RabbitMQ + Mailpit）
docker compose -f docs/deployments/docker-compose.yml up -d

# 3. 完整栈（额外 ES + MinIO）
docker compose -f docs/deployments/docker-compose.yml --profile full up -d

# 4. 首次启动需执行迁移
docker compose exec api migrate up
docker compose exec api seed   # 幂等
```

### 1.2 Kubernetes（Phase 3 候选）

Helm Chart 结构（`docs/deployments/helm/`，Phase 3 正式交付）：
- `api` Deployment（HPA，基于 CPU/内存/QPS 自动伸缩）
- `worker` Deployment（长驻消费者组）
- `web` Deployment（静态资源 + Nginx）
- PostgreSQL / Redis / RabbitMQ 建议使用云托管（CrunchyData / ElastiCache / CloudAMQP）

### 1.3 可信/信创环境

- 操作系统：麒麟 V10 / 统信 UOS / openEuler 22.03
- CPU：x86_64 或 ARM64（鲲鹏/飞腾）；Docker 镜像双架构构建
- 数据库：达梦 / 人大金仓；通过方言层（`pkg/dialect`）抽象，`build tag` 切换
- TLS 国密由上游网关（Tengine / 东方通）终结

---

## 2. 最小/推荐规格

| 形态 | 规格 | 支撑 |
|------|------|------|
| 最小（试用） | 2C4G ×1（all-in-one） | 50 用户 / 5 万工作项 |
| 推荐（生产） | api 2C4G ×2 + worker 2C4G + PG 4C16G + Redis 2C4G | 1000 并发 / 100 万工作项 |

**增长触发条件**：PG CPU >60% 持续 1 周 → 升配；ES heap >70% → 加节点。

---

## 3. 高可用架构

### PostgreSQL
- 主从流复制（`synchronous_commit = remote_apply`，关键业务路径）
- Patroni + etcd 做故障转移
- 每日全量（pg_basebackup）+ WAL 连续归档
- 读写分离（Hot Standby 承接报表/快照查询）

### Redis
- 缓存层 AOF；不持状态，DB 兜底
- 部署 Sentinel（≥3 节点）做主从切换

### RabbitMQ
- 镜像队列（`ha-mode: all` + `ha-sync-mode: automatic`）
- 单集群 ≥3 节点 Outbox Relay / Task Worker 多副本消费

### API / Worker
- API：无状态水平扩展，前置 Nginx/Envoy 反代
- Worker：长消费者多副本竞争消费（work queue 调度语义）

---

## 4. 监控与可观测

### 4.1 Prometheus + Grafana
- API `/metrics` 暴露 RED 指标（request rate / error / duration）
- 预置 4 张面板：总览 / API / 依赖（PG+Redis+RabbitMQ） / 业务

### 4.2 日志
- zap JSON，字段规范：`ts/level/msg/request_id/tenant/user/trace_id`
- 收集：Fluent Bit / Vector → Loki 或 ELK
- 保留：热 30 天 + 冷 12 个月

### 4.3 追踪
- OTel SDK，生产采样率 5%（错误 100% 采样）
- `trace_id` 注入日志与错误响应 `Error-Trace-Id` 头

### 4.4 健康检查
| 探活 | 用途 |
|------|------|
| `/healthz` | 进程存活（K8s livenessProbe）|
| `/readyz` | 依赖探活 PG+Redis（K8s readinessProbe）|

---

## 5. 发布与回滚

### 5.1 发布流程
1. PR 合入 main → CI lint/test/e2e/build 绿
2. 新建 tag `v<semver>` → CI 发布流水线触发
3. Staging 环境部署 → 冒烟 + 人工验收 ≥24h
4. Prod 部署（蓝绿或滚动更新 K8s）

### 5.2 回滚策略
- 镜像：回退到上一 tag
- DB：迁移采用 expand-contract 模式（向后兼容 1 个版本），无需回滚脚本
- 关键配置变更：先在 staging 验证

### 5.3 发布窗口
- 推荐：工作日 10:00–17:00（避开周五下午）
- 例外：hotfix，需 Admin 审批

---

## 6. 安全检查清单

| 检查项 | 说明 |
|--------|------|
| JWT Secret 长度 ≥32 字节 | `config.go` validate() 强制 |
| 响应头 CSP/HSTS 等 8 项 | SecurityHeaders 中间件默认启用 |
| RLS 策略生效 | FORCE ROW LEVEL SECURITY + app.workspace_id |
| API /metrics /swagger 不暴露公网 | 防火墙或 Ingress 限内网 |
| RabbitMQ 管理 UI 内网访问 | 不暴露 15672 |
| 数据库不直连公网 | VPN/堡垒机/对等连接 |
| `YDSZ_AUTH_JWT_SECRET` 定期轮换 | 轮换后所有用户重新登录 |
| 依赖漏洞扫描 | `make vuln`（govulncheck + pnpm audit）|

---

## 7. 验收跑分

```bash
make build         # Go + 前端双构建
make lint          # golangci-lint + eslint
make test          # Go race + 前端全量测试
make test-e2e      # Playwright E2E
make perf-smoke    # k6 冒烟
```

新建部署完成后进入 [备份与恢复 SOP](./backup-recovery.md) 完成零号备份。
