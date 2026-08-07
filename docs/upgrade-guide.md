# Ydsz Plane 升级指南 (v0.1 → v0.2)

> 适用于将已有 v0.1 部署升级到 v0.2 的运维人员。
> 预计停机时间：< 5 分钟（使用 `CREATE INDEX CONCURRENTLY` 在线索引为 0）。

---

## 1. 升级前 checklist

- [ ] 当前 v0.1 运行正常
- [ ] 已完成全量备份（`pg_dump -Fc`）
- [ ] 已读取本文档全部章节
- [ ] 已通知用户维护窗口

---

## 2. 升级步骤

### Step 1：停服前准备
```bash
# 1.1 备份数据库（必须）
pg_dump -Fc $YDSZ_DATABASE_URL > pre_v02_upgrade_$(date +%Y%m%d_%H%M%S).pgdump

# 1.2 备份配置文件
cp .env .env.v01.bak
```

### Step 2：拉取新版本
```bash
git fetch origin
git checkout v0.2.0
```

### Step 3：更新配置
v0.2 新增配置项（在 `.env` 中追加）：
```env
# 可选：项目模板默认值（已内置 generic，无需配置）
# YDSZ_DEFAULT_TEMPLATE=generic

# 新增：性能索引开关（默认 true）
# YDSZ_PERF_INDEXES=true
```

### Step 4：应用数据库迁移
```bash
go run ./cmd/migrate up
```
v0.2 新增 6 个 migrations（0014 → 0019），包括：
- 通知设置、附件、视图偏好、指标、自动化、Intake 等表
- **0019_perf_indexes**：在线创建 7 个覆盖索引（CONCURRENTLY，不锁表）

索引创建耗时估算（1M 工作项）：每个 5-30 秒，总 < 3 分钟。

### Step 5：重启服务
```bash
docker compose down
docker compose build api worker
docker compose up -d
```

### Step 6：验证
```bash
# 健康检查
curl http://localhost:8080/healthz

# 登录测试
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"admin@ydsz.dev","password":"Admin@123"}'

# 确认版本号
curl http://localhost:8080/api/v1/version
```

### Step 7：冒烟测试
使用 k6 冒烟套件快速验证：
```bash
k6 run -e BASE_URL=http://localhost:8080/api/v1 \
       -e TEST_USER_EMAIL=admin@ydsz.dev \
       -e TEST_USER_PASSWORD=Admin@123 \
       tests/perf/smoke-test.js
```

---

## 3. 不兼容变更 (Breaking Changes)

v0.2 **无**破坏性 API 变更。所有 v0.1 接口继续可用。

### 数据库变更（向后兼容）
- 所有新列为 `NULLABLE` 或带 `DEFAULT`
- 无列删除 / 重命名
- 无索引删除（仅新增）

---

## 4. 回滚方案

若升级后发现问题：

```bash
# 1. 回滚迁移（按需回滚 N 步）
go run ./cmd/migrate down 6

# 2. 切回旧版镜像
git checkout v0.1.0
docker compose build api worker
docker compose up -d

# 3. 若数据已变更，恢复备份
pg_restore --clean -d $YDSZ_DATABASE_URL pre_v02_upgrade_*.pgdump
```

---

## 5. 已知限制

- `0019_perf_indexes` 中的 `CREATE INDEX CONCURRENTLY` 不能在事务块内执行；migration 工具已处理。
- 1M+ 工作项后强烈建议施加 0019 索引，否则主列表 P95 可能超过 500ms。
- 升级后首次加载仪表盘会触发缓存预热（冷启动 5-10s 属正常）。
