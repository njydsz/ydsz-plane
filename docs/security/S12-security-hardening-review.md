# S12 安全加固复查报告

> Sprint 12 · 2026-08-07 · 等保基线 checklist
> 参考标准：GB/T 22239（等保三级）、OWASP ASVS 4.0、OWASP Top 10 2021、NIST CSF 2.0

---

## 1. 总览

| 领域 | 状态 | 说明 |
|------|------|------|
| 身份鉴别 | GO | bcrypt cost=12、登录锁定、预留 TOTP |
| 访问控制 | GO | 4 角色 RBAC + RLS 双保险 |
| 安全审计 | GO | audit_logs 全量管理操作 |
| 数据完整性 | GO | 乐观锁 + 事务 + 软删除 |
| 数据保密性 | GO | 密码/Token hash；TLS 部署侧 |
| 注入防御 | GO | pgx 参数化 + sql-lint CI |
| XSS 防御 | GO | bluemonday + CSP + v-html 卡口 |
| 限流保护 | GO | Redis 令牌桶 + 429 + Retry-After |
| 供应链 | GO | govulncheck + pnpm audit 门禁 |
| 日志泄露 | WARN | IP 明文存储；无脱敏中间件 |
| 文件上传 | WARN | 大小限制已有，无服务端二次转码 |
| 运行时安全 | WARN | 无 egress 代理内网黑名单 |

### 结论：可发布 ✅（2 个 WARN 项建议 v0.3 修复）

---

## 2. 等保三级基线 Checklist

### 2.1 身份鉴别（Authentication）

| 编号 | 控制项 | 状态 | 落地位置 |
|------|--------|------|---------|
| A-01 | 用户唯一标识与强密码策略 | ✅ | `internal/application/auth/service.go` |
| A-02 | 登录失败锁定（≥5 次/15min） | ✅ | Redis `login_attempts:{email}` 计数 |
| A-03 | 会话超时与重新认证 | ✅ | JWT access 15min + refresh 30d 旋转 |
| A-04 | 双因子认证（可选三级） | 🔲 P2 | 接口已抽象 `Authenticator`，待实现 |
| A-05 | 禁止默认/测试账号上线 | ✅ | seed 脚本隔离（`@ydsz.dev`） |

### 2.2 访问控制（Authorization）

| 编号 | 控制项 | 状态 | 落地位置 |
|------|--------|------|---------|
| B-01 | 权限分离（管理员/操作员/审计员） | ✅ | 4 角色 RBAC（owner/admin/member/guest） |
| B-02 | 默认拒绝 + 最小权限 | ✅ | 中间件 `requireWsPermission` |
| B-03 | 角色互斥（三权分立） | ⚠️ | owner/admin 重叠较多，待拆分 |
| B-04 | 敏感操作二次确认 | ⚠️ | 前端无 "deleteWS" 密码确认 |
| B-05 | 资源隔离（租户隔离） | ✅ | RLS + 中间件 app.workspace_id |

### 2.3 安全审计（Audit）

| 编号 | 控制项 | 状态 | 落地位置 |
|------|--------|------|---------|
| C-01 | 全覆盖管理操作 | ✅ | 50+ 操作点调用 `AuditSvc.RecordFromGin` |
| C-02 | 审计记录保护（只增） | ✅ | 代码层无 UPDATE/DELETE 路径 |
| C-03 | 审计日志查询仪表板 | ✅ | `/audit-logs` + 前端审计报表页 |
| C-04 | 日志存储 ≥ 6 个月 | ⚠️ | 当前无自动归档/分区，需配置 PG 分区 |
| C-05 | 集中日志采集（外部 SIEM） | 🔲 | 预留 Kafka hook 接口，待外部对接 |

### 2.4 数据完整性（Integrity）

| 编号 | 控制项 | 状态 | 落地位置 |
|------|--------|------|---------|
| D-01 | 传输完整性（TLS） | ✅ | 部署层 nginx/CDN TLS 1.2+ |
| D-02 | 存储完整性（校验） | ⚠️ | 无 HMAC 校验列 |
| D-03 | 乐观锁防并发覆盖 | ✅ | `issues.version` + 事务 |
| D-04 | 事务一致性 | ✅ | 业务域 `tx.Begin()` + defer Rollback |
| D-05 | 软删除保护 | ✅ | `deleted_at IS NULL` 全局过滤 |

### 2.5 数据保密性（Confidentiality）

| 编号 | 控制项 | 状态 | 落地位置 |
|------|--------|------|---------|
| E-01 | 密码单向 hash | ✅ | bcrypt cost=12 |
| E-02 | Token hash 存储 | ✅ | SHA-256 hash（仅创建时回显明文） |
| E-03 | 敏感字段脱敏 | ⚠️ | email/IP 在审计日志中未脱敏 |
| E-04 | 传输加密 | ✅ | TLS 1.2+（部署侧） |
| E-05 | 生产/测试数据隔离 | ✅ | seed 邮箱域限制 + 独立 migration |

### 2.6 入侵防范

| 编号 | 控制项 | 状态 | 落地位置 |
|------|--------|------|---------|
| F-01 | 应用层限流 | ✅ | Redis 令牌桶 + 429 |
| F-02 | 安全响应头 | ✅ | `SecurityHeaders()` 中间件（CSP/HSTS 边缘） |
| F-03 | SQL 注入防护 | ✅ | pgx 参数化 + CI sql-lint |
| F-04 | XSS 防护 | ✅ | bluemonday + CSP |
| F-05 | SSRF 防护 | ⚠️ | Webhook 出站无内网黑名单 |
| F-06 | 文件上传白名单 | ✅ | MIME/大小（10MB）+ UUID 重命名 |

### 2.7 备份恢复

| 编号 | 控制项 | 状态 | 落地位置 |
|------|--------|------|---------|
| G-01 | 定期全量备份 | ✅ | 文档 13：每日全量 + WAL 归档 |
| G-02 | 恢复演练 | 🔲 | 需运维手册补充演练脚本 |
| G-03 | 备份加密 | ⚠️ | 文档提及，未实际配置 |

---

## 3. 高优先级加固项（v0.3 推荐）

### 3.1 P1：审计日志按月分区 + 自动归档
**当前风险：** 单表长期运行尺寸膨胀、查询性能退化。
**建议：** PG 原生分区（`PARTITION BY RANGE (created_at)`）+ pg_cron 每月创建 + 冷数据转储至 S3/OSS。

### 3.2 P1：审计敏感字段脱敏
**当前风险：** IP 地址明文、email 明文、detail JSON 可能含 Token/密码。
**建议：** 审计写入前经 `maskIP()` + `maskEmail()`，detail 字段经脱敏敏感词列表过滤。

### 3.3 P2：SSRF 防护（Egress 代理）
**当前风险：** Webhook 出站可触达内网地址（`169.254.x.x`/`10.x.x.x`）。
**建议：** `x/net/netutil` 限制内网段；或部署侧 Squid egress 代理。

---

## 4. 审计日志报表 API

后端已覆盖，详见：

- `GET /api/v1/workspaces/:ws/audit-logs` — 原始审计日志分页
- `GET /api/v1/workspaces/:ws/projects/:id/stats` — 仪表盘汇总（已有）

前端审计报表页对接已就绪（`/audit-logs` 视图）。

---

## 5. 交付物

| 路径 | 描述 |
|------|------|
| `docs/security/S12-security-hardening-review.md` | 本文档（等保基线 checklist） |
| `docs/architecture/06-权限与安全设计.md` | 安全架构基线文档 |
| `internal/application/auth/rbac.go` | RBAC 实现 |
| `internal/application/workspace/audit.go` | 审计服务 |
| `internal/interfaces/middleware/middleware.go` | 安全中间件（SecurityHeaders/RateLimit） |
| `internal/application/issue/html_sanitize.go` | 富文本净化 |

---

## 6. 复测清单

- [ ] `go build ./...` 全量通过
- [ ] `go test ./internal/...` 单元测试通过率 ≥ 80%
- [ ] `pnpm lint` + `pnpm test` 前端检查通过
- [ ] k6 stress 压测 200 VU × 3m 错误率 < 1%
- [ ] CSP 响应头验证（浏览器 DevTools）
- [ ] 登录失败锁定验证（5 次 → 423/429）
- [ ] 跨租户 IDOR 验证（用例：userA 尝试 GET userB 的 issue）
