# P2-11 · RLS 跨数据库兼容验证文档

> 版本：v1.0
> 日期：2026-06-26
> 范围：ydsz-plane PostgreSQL 18 → 达梦 DM8 / 人大金仓 KingbaseES V8 RLS 迁移可行性评估
> 依赖：`internal/infrastructure/persistence/db.go`、`dialect.go`、`sql/ydsz-plane-init.sql`

---

## 1. PostgreSQL 18 RLS 实现确认（当前基线）

### 1.1 实现方式：SET LOCAL + app.workspace_id + POLICY

当前 PostgreSQL 的租户隔离**并非依赖原生 RLS POLICY**，而是通过应用层代码手动注入 session 变量，利用 WHERE 子句实现行级过滤。具体机制：

| 组件 | 文件 | 作用 |
|------|------|------|
| 连接池包装 | `internal/infrastructure/persistence/db.go` — `Pool.WithTenantTx()` | 事务开启后执行 `SELECT set_config('app.workspace_id', $1, true)` |
| 应用层调用 | 17 处 `tx.Exec(ctx, "SELECT set_config('app.workspace_id', $1, true)", ...)` | 各领域服务在事务入口注入当前 workspace ID |
| 过滤逻辑 | 业务 SQL WHERE 子句 | 开发者在每个查询中手动添加 `WHERE workspace_id = $N` 或 `WHERE workspace_id = current_setting('app.workspace_id')::bigint` |

### 1.2 设计特征

- **三层次隔离**：tenant_id → workspace_id → project_id
- `SET LOCAL` 仅在事务内生效，连接池场景安全
- 41 张业务表均含 `workspace_id BIGINT NOT NULL DEFAULT 0`
- 查询层强制拼接 `workspace_id` 过滤条件
- `search_documents` 等表注释提到 `FORCE ROW LEVEL SECURITY`，但当前迁移脚本中**未包含 `CREATE POLICY`**（grep 确认 POLICY 关键字在 ydsz-plane-init.sql 中 0 次出现）

### 1.3 现状结论

> ✅ **PostgreSQL 的租户隔离实际上走的是应用层强制 WHERE 路径，不是原生 RLS POLICY。** 这意味着跨数据库迁移时：
> - 不需要在目标数据库上实现 RLS POLICY 语法
> - 只需确保目标数据库支持 `SET SESSION`/`SET LOCAL`-风格的 session 变量
> - 或者直接在所有 SQL 中硬编码 `workspace_id = ?` 过滤条件

---

## 2. 达梦 DM8 RLS 现状分析

### 2.1 达梦是否支持原生 RLS

| 特性 | DM8 支持情况 |
|------|-------------|
| PostgreSQL 式 RLS POLICY | ❌ 不支持（DM8 不兼容 PG RLS 语法） |
| Oracle 式 Virtual Private Database (VPD) | ⚠️ 部分支持（通过 `DBMS_RLS` 包或触发器模拟） |
| Row-Level Security 关键字 | ❌ 不支持 |
| Session 上下文变量（CLIENT_IDENTIFIER 等价物） | ✅ 支持（通过 `DBMS_SESSION.SET_CONTEXT` 或 `SF_SET_SYSTEM_PARAM`） |

### 2.2 达梦 VPD 等价实现方式

达梦 DM8 可通过以下两种方式实现 PG RLS 的等价功能：

#### 方案 A：触发器 + 行级访问控制（推荐）

```sql
-- 1. 创建上下文命名空间
CREATE CONTEXT app_ctx app_info USING set_app_ctx;

-- 2. 创建上下文设置包
CREATE OR REPLACE PACKAGE set_app_ctx AS
    PROCEDURE set_workspace_id(p_id BIGINT);
END;
/

CREATE OR REPLACE PACKAGE BODY set_app_ctx AS
    PROCEDURE set_workspace_id(p_id BIGINT) IS
    BEGIN
        DBMS_SESSION.SET_CONTEXT('app_ctx', 'workspace_id', TO_CHAR(p_id));
    END;
END;
/

-- 3. 创建视图层过滤
CREATE OR REPLACE VIEW v_issues AS
SELECT * FROM issues
WHERE workspace_id = NVL(
    TO_NUMBER(SYS_CONTEXT('app_ctx', 'workspace_id')),
    0  -- 无上下文时返回空集
);
-- 注意：需要 INSTEAD OF 触发器支持写操作路由回基表
```

#### 方案 B：应用层强制 WHERE + BEGIN 脚本（最简单）

```sql
-- 达梦支持 SET 命令，但不保证所有 JDBC 驱动传递 session 变量
-- 最稳妥做法：应用层 Go 代码中，方言为 dameng 时直接拼接参数化 WHERE

-- 示例：BEGIN 脚本设置上下文（DM8 兼容）
BEGIN
    -- DM8 无 set_config，但有自定义系统过程
    SF_SET_SYSTEM_PARAM('app_workspace_id', '12345', 1);
END;
```

### 2.3 代码回退方案

如果达梦无法支持 session 变量机制，推荐在方言层自动注入 `WHERE workspace_id`：

```go
// internal/infrastructure/persistence/dialect.go
// DamengDialect 增加 QueryFilter 方法
func (d *DamengDialect) TenantFilter(workspaceID int64, paramOffset int) string {
    // 达梦参数占位符为 :1, :2, ...
    return fmt.Sprintf("workspace_id = :%d", paramOffset)
}
```

或在 Repository 层根据方言类型动态选择：

```go
// 当 DialectType == DialectDameng 时
// 所有查询自动追加参数化条件（不依赖 session 变量）
func (r *issueRepo) buildTenantFilter(dialect DialectType, wsID int64, baseQuery string) (string, []any) {
    if dialect == DialectDameng {
        return baseQuery + " AND workspace_id = $1", []any{wsID}
    }
    // postgres: 保持 set_config + current_setting 方案
    return baseQuery, nil
}
```

### 2.4 DM8 关键差异清单

| 差异点 | PostgreSQL | 达梦 DM8 | 兼容性风险 |
|--------|-----------|---------|-----------|
| Session 变量 | `set_config('app.workspace_id')` | `SF_SET_SYSTEM_PARAM` / 自定义包 | 中 — 需适配 |
| 参数化 WHERE | `$1::bigint` | `:N` | 低 — dialect 层已抽象 |
| `current_setting()` | 内置函数 | ❌ 不支持 | 高 — 需移除所有引用 |
| CTE / WITH 子句 | ✅ | ✅（有限） | 中 |
| `unnest()` 批量操作 | ✅ | ❌ 不支持 | 高 — batch 操作需重写 |
| `COPY FROM` 批量写入 | ✅ | ❌（用 `INSERT ALL` 或 JDBC batch） | 中 |
| JSONB 操作符 | `->>` / `@>` | `JSON_VALUE()` / `JSON_QUERY()` | 低 — dialect 层已抽象 |
| ILIKE | ✅ | `UPPER(col) LIKE UPPER(val)` | 低 — dialect 层已抽象 |
| `now()` 函数 | ✅ | `SYSDATE` 或 `CURRENT_TIMESTAMP` | 低 — dialect 层已抽象 |

---

## 3. 人大金仓 KingbaseES RLS 现状分析

### 3.1 金仓是否支持原生 RLS

| 特性 | KingbaseES V8 支持情况 |
|------|----------------------|
| PostgreSQL 式 RLS POLICY | ✅ **原生支持**（基于 PG 9.5+ RLS 语法完全兼容） |
| `ROW LEVEL SECURITY` | ✅ 支持 |
| `CREATE POLICY` | ✅ 支持（USING / WITH CHECK 子句均兼容） |
| `ALTER TABLE ... ENABLE ROW LEVEL SECURITY` | ✅ 支持 |
| `ALTER TABLE ... FORCE ROW LEVEL SECURITY` | ✅ 支持 |
| `set_config()` | ✅ 支持（PG 兼容） |
| `current_setting()` | ✅ 支持（PG 兼容） |

### 3.2 兼容级别评估

KingbaseES V8 基于 PostgreSQL 内核，RLS 实现与 PostgreSQL 高度兼容：

```
兼容度评估：★★★★★（95%+）
```

| 语法特性 | 兼容性 | 备注 |
|---------|--------|------|
| `CREATE POLICY name ON table USING (expr)` | ✅ 完全兼容 | 与 PG 15+ 语法一致 |
| `FOR SELECT / FOR INSERT / FOR UPDATE / FOR ALL` | ✅ 完全兼容 | — |
| `WITH CHECK` 表达式 | ✅ 完全兼容 | — |
| `current_setting('app.workspace_id')::bigint` | ✅ 完全兼容 | — |
| `set_config('app.workspace_id', val, true)` | ✅ 完全兼容 | — |
| `SECURITY INVOKER` / `SECURITY DEFINER` | ✅ 完全兼容 | — |

### 3.3 已知差异

| 差异点 | 影响等级 | 说明 |
|--------|---------|------|
| 扩展插件生态 | 中 | 部分 PG 扩展（如 pg_stat_statements）需单独评估 |
| 性能特征 | 低 | 相同 RLS 策略执行计划可能不同，需压测 |
| `gen_random_uuid()` | 低 | 金仓自带 `pgcrypto` 或使用内置 `uuid-ossp` |
| 全文搜索 `to_tsvector` | 低 | 金仓支持，但中文分词可能用不同内置解析器 |
| 索引类型 | 低 | BRIN/GiST/SP-GiST 兼容，但可能不支持 PG 18 新增特性 |

### 3.4 金仓迁移建议

> **KingbaseES 是三个数据库中迁移成本最低的**。由于与 PostgreSQL 内核兼容，当前 SET LOCAL + app.workspace_id 方案可直接运行，无需代码修改。甚至可以启用真正的 RLS POLICY 作为纵深防御层。

---

## 4. 推荐测试矩阵

### 4.1 核心隔离测试用例

| 测试编号 | 描述 | 预期结果 | 严重级别 |
|---------|------|---------|---------|
| RLS-TC-001 | Workspace A 用户读取 issues 表，不应看到 Workspace B 的数据 | 返回空集或 404 | P0 阻断 |
| RLS-TC-002 | Workspace A 用户更新 Workspace B 的 Issue，应拒绝 | 返回 0 rows affected | P0 阻断 |
| RLS-TC-003 | Workspace A 用户尝试 DELETE Workspace B 的 Issue | 返回 0 rows affected | P0 阻断 |
| RLS-TC-004 | 同一 workspace 内 member 可读写各自权限范围内的数据 | 正常返回 | P1 |
| RLS-TC-005 | Workspace owner 可访问 workspace 内所有项目数据 | 正常返回 | P1 |
| RLS-TC-006 | 无 workspace 上下文（app.workspace_id = NULL）时所有查询返回空 | 空集且无 ERROR | P1 |
| RLS-TC-007 | 批量 `unnest()` 插入 issues 时隔离是否生效 | 仅当前 workspace 数据可见 | P1 |
| RLC-TC-008 | 跨 workspace 的 relation/dependency 链接是否被 RLS 阻止 | 应返回 404 | P0 |

### 4.2 角色与权限矩阵

| 角色 | 读本租户 | 读跨租户 | 写本租户 | 写跨租户 | 删除本租户 |
|------|---------|---------|---------|---------|----------|
| workspace.owner | ✅ | ❌ | ✅ | ❌ | ✅ |
| workspace.admin | ✅ | ❌ | ✅ | ❌ | ✅ |
| workspace.member | ✅ | ❌ | ✅ (限自有) | ❌ | ❌ |
| workspace.guest | ✅ (只读) | ❌ | ❌ | ❌ | ❌ |

### 4.3 跨数据库测试矩阵

| 测试维度 | PostgreSQL 18 | 达梦 DM8 | KingbaseES V8 |
|---------|--------------|---------|--------------|
| 基本 CRUD 隔离 | ✅ 已验证 | ⬜ 待验证 | ⬜ 待验证 |
| 事务内 SET LOCAL 生效 | ✅ | ⬜ 待验证 | ⬜ 待验证 |
| 批量 unnest/copyFrom 隔离 | ✅ | ⬜ 待验证 | ⬜ 待验证 |
| 全文检索 + workspace 过滤 | ✅ | ⬜ 待验证 | ⬜ 待验证 |
| JSONB 操作 + workspace 过滤 | ✅ | ⬜ 待验证 | ⬜ 待验证 |
| FK 逻辑约束隔离 | ✅ | ⬜ 待验证 | ⬜ 待验证 |
| 并发写入隔离（race detector） | ✅ | ⬜ 待验证 | ⬜ 待验证 |
| 软删除 + workspace 过滤 | ✅ | ⬜ 待验证 | ⬜ 待验证 |

### 4.4 跨 DB Migration 验证清单

```markdown
#### DDL 兼容性
- [ ] PostgreSQL `BIGINT GENERATED ALWAYS AS IDENTITY` → DM8 `BIGINT IDENTITY` / Kingbase 兼容
- [ ] `GEN_RANDOM_UUID()` → DM8 `SYS_GUID()` / Kingbase `gen_random_uuid()`
- [ ] `TIMESTAMPTZ` → DM8 `TIMESTAMP WITH TIME ZONE` / Kingbase 兼容
- [ ] `JSONB` → DM8 `CLOB` + JSON函数 / Kingbase `JSONB`
- [ ] `CHECK` 约束内枚举 → DM8 `CHECK IN ()` / Kingbase 兼容
- [ ] `COMMENT ON TABLE/COLUMN` → DM8 `COMMENT ON` / Kingbase 兼容
- [ ] 部分索引 `WHERE deleted = false` → DM8 不支持部分索引 / Kingbase 兼容

#### 函数兼容性
- [ ] `current_setting()` → DM8 不支持 → **需适配层**
- [ ] `set_config()` → DM8 不支持 → **需适配层**
- [ ] `unnest()` → DM8 不支持 → **需自定义函数或应用层循环**
- [ ] `to_tsvector/to_tsquery` → DM8 `CONTAINS` / Kingbase 兼容
- [ ] `now()` → DM8 `SYSDATE` / Kingbase 兼容
- [ ] `array_agg` → DM8 `WM_CONCAT` / Kingbase 兼容

#### 触发器兼容性
- [ ] `AFTER UPDATE FOR EACH ROW` → DM8 `FOR EACH ROW` 兼容 / Kingbase 兼容
- [ ] `NEW`/`OLD` 引用 → DM8 `:NEW`/`:OLD` / Kingbase `NEW`/`OLD`
- [ ] 触发器内 `current_setting` → DM8 不支持

#### 索引兼容性
- [ ] B-tree 索引 → 全兼容
- [ ] GIN 索引 (JSONB) → DM8 不支持 / Kingbase 兼容
- [ ] GiST 索引 → DM8 不支持 / Kingbase 兼容
- [ ] 表达式索引 (`LOWER(email)`) → DM8 有限支持 / Kingbase 兼容
- [ ] 部分索引 → DM8 不支持 / Kingbase 兼容
```

---

## 5. 结论与风险等级

### 5.1 各数据库 RLS 方案推荐

| 数据库 | 推荐方案 | 迁移工作量 | 风险等级 |
|--------|---------|-----------|---------|
| PostgreSQL 18 | 维持现状（SET LOCAL + 应用层 WHERE） | — | 🟢 低 |
| KingbaseES V8 | 维持现状 + 可选启用 RLS POLICY 纵深防御 | 1-2 人天 | 🟢 低 |
| 达梦 DM8 | 应用层强制 WHERE + dialect.TenantFilter 适配 | 5-10 人天 | 🟡 中 |

### 5.2 达梦迁移风险详情

| 风险项 | 等级 | 缓解措施 |
|--------|------|---------|
| `current_setting('app.workspace_id')` 不可用 | 🔴 高 | 在 DamengDialect 中增加 `TenantFilter()` 方法，Repository 层根据 DialectType 自动切换 |
| `unnest()` 批量操作不可用 | 🟡 中 | 改写为应用层循环或利用 JDBC batch；或通过 DBMS_SQL 动态执行 |
| GIN/GiST 索引不可用 | 🟡 中 | 评估是否改用 B-tree + 函数索引；或接受全表扫描 + 应用层过滤 |
| `COPY FROM` 批量导入不可用 | 🟢 低 | 改写为 INSERT ALL 或分批 INSERT |
| MERGE INTO 替代 ON CONFLICT | 🟡 中 | 已在 dialect.go 中提供 MERGE INTO 语法模板 |
| 触发器语法差异 | 🟢 低 | DM8 触发器语法接近 Oracle，迁移时调整 `NEW/OLD` 引用即可 |

### 5.3 迁移优先级建议

```
Phase 1 (P0 - 阻断):
  □ 实现 DamengDialect.TenantFilter() 替代 current_setting 依赖
  □ 为 17 处 set_config 调用增加方言分支
  □ pg → dm：替换 3 处 unnest/copyFrom 批量操作

Phase 2 (P1 - 重要):
  □ 在 CI 中增加达梦/金仓的 SQL 语法 lint
  □ 编写跨库集成测试（至少覆盖 RLS-TC-001 ~ RLS-TC-006）
  □ 评估 KingbaseES 启用 RLS POLICY 作为安全纵深

Phase 3 (P2 - 改善):
  □ DM 触发器语法迁移验证
  □ 性能基线对比测试（PG vs DM8 vs KingbaseES）
  □ 自动化 DDL diff 工具验证（migration 脚本跨库兼容性）
```

### 5.4 最终结论

ydsz-plane 当前的租户隔离实现方案（应用层 SET LOCAL + WHERE）具有较好的跨数据库可移植性：

> **KingbaseES 可直接兼容，达梦需 dialect 层适配（约 5-10 人天工作量），不存在架构级阻断风险。**

建议在 dialect.go 中抽象出稳定的 `TenantFilter` / `BulkInsert` / `JSONBExtract` 接口，使 Repository 层代码无需关心底层数据库差异，实现真正的"配置驱动"信创适配。
