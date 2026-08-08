# 备份与恢复 SOP

> 配套 [运维手册](./README.md) · [备份与灾备设计](../architecture/13-部署运维与可靠性设计.md)

---

## 1. 备份矩阵

| 对象 | 工具/方式 | 频率 | 保留 | RPO | RTO |
|------|-----------|------|------|-----|-----|
| PostgreSQL 全量 | `pg_dump -Fc` | 每日 02:00 UTC | 30 天 | ≤24h | ≤1h |
| PostgreSQL WAL | 持续归档（archive_command） | 实时 | 7 天 | ≤5min | ≤1h |
| Redis dump.rdb | `BGSAVE` | 每日 | 7 天 | 24h | ≤5min |
| MinIO 桶 | `mirror` | 每日 | 30 天 | 24h | ≤4h |
| 配置文件 | `.env` + `docker-compose.yml` | 每次变更 | 永久 | 0 | ≤5min |

---

## 2. PostgreSQL 备份

### 2.1 全量逻辑备份
```bash
pg_dump -h <host> -p 5432 -U ydsz -d ydsz \
  -Fc --no-owner --no-privileges \
  -f backup-$(date +%F-%H%M).dump

# 校验完整性
pg_restore --list backup-*.dump > /dev/null && echo "OK"
```

### 2.2 物理备份（大库推荐，支持 PITR）
```bash
pg_basebackup -h <host> -U replicator -D /backup/pg-$(date +%F) \
  -Ft -X fetch -P -v
```

### 2.3 WAL 归档
postgresql.conf:
```ini
wal_level = replica
archive_mode = on
archive_command = 'test ! -f /archive/%f && cp %p /archive/%f'
archive_timeout = 300
```

---

## 3. 恢复演练（季度必做）

### 3.1 PG 全量恢复
```bash
# 1. 暂停 API/Worker
docker compose stop api worker

# 2. 准备空实例
dropdb -h <host> -U ydsz ydsz
createdb -h <host> -U ydsz ydsz

# 3. 恢复
pg_restore -h <host> -U ydsz -d ydsz \
  -j 4 --no-owner --role=ydsz \
  backup-<date>.dump

# 4. 确认行数（与备库或历史快照对比）
psql -h <host> -U ydsz -d ydsz -c "SELECT count(*) FROM issues;"

# 5. 重新启动服务
docker compose up -d api worker
```

### 3.2 PITR（时间点恢复）
```bash
# 1. 停库
pg_ctl stop -D /var/lib/postgresql/data

# 2. 拿下一个 WAL 归档与 target time
cat > recovery.signal <<EOF
restore_command = 'cp /archive/%f %p'
recovery_target_time = '2026-08-08 10:30:00+08'
recovery_target_action = 'pause'
EOF

# 3. 启动并观察日志
pg_ctl start -D /var/lib/postgresql/data
```

---

## 4. 跨区域同步

```bash
# MinIO 异地镜像
mc mirror --watch src-minio/attachments backup-minio/attachments

# PG 流复制到异地只读从库
#（在异地实例配置 primary_conninfo 指向主库）
```

---

## 5. 备份完整性验证

每次备份后自动执行：
```bash
pg_restore --list backup-*.dump && \
sha256sum backup-*.dump > backup-*.sha256 && \
echo "backup OK $(date)" | slackcat -c alerts
```

季度恢复演练后填写《备份恢复演练记录表》：
- 演练日期
- 实际 RTO
- 数据完整性校验结果
- 问题与改进措施

---

## 6. 告警门

| 事件 | 严重度 | 处置 |
|------|--------|------|
| 备份脚本执行失败 | P0 | 立即重试 + oncall 介入 |
| RPO > 24h | P0 | on-call 10min 内响应 |
| 归档磁盘 >80% | P1 | 扩展磁盘或缩短保留期 |
| 校验失败 | P0 | 停服评估 + 重建备份 |
