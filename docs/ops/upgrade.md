# 升级指南

> 配套 [运维手册](./README.md) · [发布管理](../architecture/15-发布管理.md)

---

## 1. 版本规范

遵循 [Semantic Versioning](https://semver.org/)：
- `v<major>.<minor>.<patch>`
- incompatible DB 变更升 major（v0.x → v1.0 时发生）
- 递增功能升 minor
- 补丁修复升 patch

---

## 2. 升级前准备

### 2.1 阅读 CHANGELOG
每个版本 tag 附 CHANGELOG，关注：
- Breaking changes（DB / API / 配置）
- 新必备环境变量
- 迁移注意事项

### 2.2 环境备份
升级前必做零时快照（详见 [备份与恢复 SOP](./backup-recovery.md)）：
```bash
# PostgreSQL
pg_dump -h <host> -U <user> -Fc ydsz > backup-$(date +%F).dump

# Redis（可选，缓存可重建）
redis-cli BGSAVE
cp /var/lib/redis/dump.rdb dump-$(date +%F).rdb

# 配置文件
cp .env .env.bak.$(date +%F)
cp docker-compose.yml docker-compose.yml.bak.$(date +%F)
```

### 2.3 Staging 验证
先在 staging 环境执行升级，24h 无异常再上 prod。

---

## 3. Compose 升级步骤

```bash
# 1. 停服前通知用户（可选维护页）
echo "Ydsz Plane 将于 10:00-10:15 进行升级维护" > /tmp/maintenance.html

# 2. 拉取新代码
git fetch --tags
git checkout v<new-version>

# 3. 更新配置（对照 CHANGELOG 的新增环境变量）
cp .env .env.bak
vim .env

# 4. 拉取新镜像
docker compose -f docs/deployments/docker-compose.yml pull

# 5. 滚动停启
docker compose up -d --remove-orphans

# 6. 自动迁移（镜像启动入口已含 migrate）
docker compose exec api migrate up

# 7. 冒烟验证
curl -f http://localhost:8080/healthz
curl -f http://localhost:8080/readyz
make test-e2e --base-url=http://localhost:5173

# 8. 确认 Worker 消费恢复
docker compose logs worker --tail=20 | grep -E "consumer|relay" 
```

---

## 4. K8s 升级步骤

```bash
# 1. 更新镜像 tag
helm upgrade ydsz-plane ./charts/ydsz-plane \
  --set api.image.tag=v<new-version> \
  --set worker.image.tag=v<new-version>

# 2. 等待滚动更新
kubectl rollout status deployment/api
kubectl rollout status deployment/worker

# 3. 迁移（Job 模式）
kubectl create job migrate-$(date +%s) --from=cronjob/db-migrate

# 4. 冒烟验证
kubectl exec deploy/api -- curl -f localhost:8080/readyz
```

---

## 5. 回滚

### 5.1 镜像回滚
```bash
# Compose
docker compose up -d --force-recreate
# （或 tag 回退：修改 image tag 后重新 up）

# K8s
helm rollback ydsz-plane <revision>
kubectl rollout undo deployment/api
```

### 5.2 数据库回滚
- 正向迁移已是 expand-contract，无需回退脚本即可让上一版本继续工作
- 若出现异常需全量恢复：
```bash
pg_restore -h <host> -U <user> -d ydsz backup-<date>.dump
```

---

## 6. 迁移注意事项

### expand-contract 原则
Ydsz Plane 迁移遵循以下约定：
1. **Expand**：新版本加列/索引/表 — 旧版本不读新列，正常运行
2. **Migrate**：双写期间补齐存量数据
3. **Contract**：下一版本移除旧列/旧表

**升级无需回滚脚本**（兼容旧版本最后一个 release 的数据形态）。

### 关键迁移风险点
| 场景 | 风险 | 缓解 |
|------|------|------|
| 大表 DDL | 锁表 / 复制延迟 | 使用 `CREATE INDEX CONCURRENTLY`；长 DDL 用 `pgroll` |
| 新增 NOT NULL 列 | 插入失败 | 先 ADD NULLABLE → backfill → SET NOT NULL |
| 新增枚举值 | 校验错误 | `ALTER TYPE ... ADD VALUE`（PG 11+ 无需事务）|
| 删除列 | 旧版本读失败 | expand-contract 下一版本删 |

---

## 7. 升级后验收清单

- [ ] `/healthz` / `/readyz` 200
- [ ] `/metrics` 指标上报正常
- [ ] 核心 API（登录 / 看板 / 详情 / 搜索）冒烟返回 200
- [ ] Worker 日志无 ERROR；Outbox Relay 无积压
- [ ] WebSocket 推送正常（`/ws/:id`）
- [ ] E2E 关键旅程通过
- [ ] 数据库迁移版本号与 tag 对齐（`schema_migrations.table_schema_version`）
