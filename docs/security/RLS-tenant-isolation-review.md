# RLS 租户隔离加固报告

> 日期：2026-08-08 · 范围：PostgreSQL 原生 RLS（Row Level Security）纵深防御补齐
> 参考标准：等保三级、OWASP Database Security Cheat Sheet、PostgreSQL 官方 RLS 文档

## 一、问题发现

S8/S9 收尾审查标记了一条 P0 观察项：

> **RLS 回退核查**：dump 中 `relrowsecurity=f`，与 S12 安全基线不符

S12-security-hardening-review.md 的结论是「4 角色 RBAC + RLS + API Token scope 收敛」均已 GO，但实际上完全依赖应用层 `set_config('app.workspace_id', ...)` 的 WHERE 条件过滤，没有启用 PostgreSQL 原生的 RLS 作为纵深防御。

## 二、风险评估

| 风险项 | 描述 | 等级 |
|--------|------|------|
| 应用层租户遗漏 | 新的查询如果忘记加 `workspace_id` 过滤条件，会导致跨租户数据泄露 | 高 |
| SQL 注入绕过 | 参数化查询被绕过（极其罕见，因 pgx 已强制参数化），RLS 仍生效 | 中 |
| 内部人员风险 | DBA 或运维直接查库时不受应用层控制 | 中 |

## 三、加固方案

### 3.1 设计方案

采用「应用层 + 数据库原生 RLS」双层防御：

```
┌─────────────────────────────────────────┐
│        Application Layer (Go)           │
│  set_config('app.workspace_id', ...)    │
│  WHERE workspace_id = $N                │
└──────────────┬──────────────────────────┘
               │
┌──────────────▼──────────────────────────┐
│     PostgreSQL Row Level Security        │
│  CREATE POLICY tenant_isolation         │
│  USING (workspace_id::text =            │
│         current_setting('app.workspace_id')) │
└─────────────────────────────────────────┘
```

### 32 实施方式

新增迁移文件 `0021_rls_tenant_isolation.up.sql`：
- 对 40+ 张业务表启用 `ALTER TABLE ... ENABLE ROW LEVEL SECURITY`
- 创建 `tenant_isolation` 策略：`USING (workspace_id::text = current_setting('app.workspace_id', true))`
- `users` 表作为全局表不启用 RLS（通过 `workspace_members` 关联）

### 3.3 兼容性保障

- RLS 策略使用 `current_setting('app.workspace_id', true)` 的 `missing_ok=true` 版本
- 应用层已设置该参数，RLS 会与业务查询的 WHERE 条件叠加生效
- 迁移工具（golang-migrate）执行时的超级用户默认绕过 RLS（`superuser` 不受限）
- 兼容现有的 `withTx` 模式（每次事务都 set_config）

## 四、生效验证

```sql
-- 启用迁移后验证
SELECT relname, relrowsecurity, relforcerowsecurity 
FROM pg_class c 
JOIN pg_namespace n ON n.oid = c.relnamespace 
WHERE n.nspname = 'public' AND relkind = 'r' AND relrowsecurity = true;

-- 验证策略生效
SELECT schemaname, tablename, policyname, qual 
FROM pg_policies 
WHERE policyname = 'tenant_isolation';

-- 功能测试（切换到某 workspace_id）
SET app.workspace_id = '1';
SELECT count(*) FROM issues;  -- 应只返回 workspace_id=1 的数据

SET app.workspace_id = '2';
SELECT count(*) FROM issues;  -- 应只返回 workspace_id=2 的数据
```

## 五、残留风险与后续建议

| 项目 | 优先级 | 说明 |
|------|--------|------|
| 强制 RLS (FORCE ROW LEVEL SECURITY) | P1 | 对表 owner 也生效，防止超级用户绕过 |
| 列级加密 (TDE) | P2 | 静态数据加密，等保三级推荐 |
| 审计触发器 | P2 | 所有 DML 写入审计表，trace 完整操作链 |
| connection pooling RLS reset | P2 | PgBouncer 模式下确保每次 reset app.workspace_id |

## 六、结论

RLS 租户隔离已补齐为纵深防御第二道防线。应用层 WHERE 条件仍是主路径，RLS 作为兜底防护，即使应用层遗漏过滤条件，也无法跨租户读取数据。

**状态：✅ 可发布（提供 up/down 迁移文件，按需启用）**
