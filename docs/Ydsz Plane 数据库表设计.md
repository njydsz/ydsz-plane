# Ydsz Plane 数据库表设计

> 基于产品需求文档（PRD V1.0）+ 架构设计文档 + 用户确认的设计约束  
> 数据库：PostgreSQL 16+（兼容达梦/人大金仓）  
> 最后更新：2026-08-10（V1.2 PRD 对齐修订版）

## 修订记录

| 版本 | 日期 | 变更说明     |
|------|------|----------|
| V1.0 | 2026-08-08 | 初始版本133 张表 |

## 全局设计约定

### 字段规范

| 约定 | 规格 |
|------|------|
| 主键 | `id BIGINT` — 雪花算法（美团 Leaf/Snowflake），递增有序，应用层生成 |
| 短代码 | `code VARCHAR(50)` — 用户可选按规则生成的业务标识符（如 PROJ-001 / PROJ1-TASK-1001） |
| 名称 | `name VARCHAR(255) NOT NULL` — 业务名称 |
| 状态 | `status VARCHAR(50) NOT NULL` — 状态（CHECK约束） |
| 租户隔离 | `tenant_id BIGINT NOT NULL` — 所有业务表必备，数据隔离唯一维度 |
| 软删除 | `deleted BOOLEAN NOT NULL DEFAULT false` - 回收站支持 |
| 创建人 | `created_by BIGINT NOT NULL` — 用户表逻辑外键 |
| 更新人 | `updated_by BIGINT NOT NULL` — 用户表逻辑外键 |
| 创建时戳 | `created_at TIMESTAMPTZ NOT NULL DEFAULT now()` |
| 更新时戳 | `updated_at TIMESTAMPTZ NOT NULL DEFAULT now()` |
| 无物理外键 | 所有关联通过代码层维护，逻辑外键字段加索引 |

### 统一基础字段模板

以下 10 个字段是所有业务表的固定骨架，每张表无一例外均携带：

```sql
id            BIGINT PRIMARY KEY,        -- 雪花ID，应用层生成
code          VARCHAR(50),               -- 用户可选短代码标识（如 REQ-1001），可与 tenant_id 联合唯一
name          VARCHAR(255) NOT NULL,     -- 名称
status        VARCHAR(50) NOT NULL,      -- 状态（CHECK约束具体值）
deleted       BOOLEAN NOT NULL DEFAULT false,
tenant_id     BIGINT NOT NULL,           -- 租户（组织）ID，数据隔离维度
created_by    BIGINT NOT NULL,           -- 创建人（逻辑外键 → users.id）
created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
updated_by    BIGINT NOT NULL,           -- 最后更新人（逻辑外键 → users.id）
updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
```

> 关联弱实体表（如 entity_labels、sprint_assignments 等 M2M 表）与日志/快照/审计类表可酌情不携带 `code` / `name`（置 NULL），其余 8 个字段必须完整。

### 逻辑外键索引惯例

逻辑外键字段统一命名为 `xxx_id`，并在之上建立普通索引：

```sql
CREATE INDEX "{table}_{xxx}_id_idx" ON "{table}" (tenant_id, {xxx}_id) WHERE NOT deleted;
CREATE INDEX "{table}_code_idx" ON "{table}" (tenant_id, code) WHERE code IS NOT NULL AND NOT deleted;
```

### 雪花 ID 注意事项

- 类型 `BIGINT`，非自增，由美团 Leaf 或 Snowflake 变体生成
- 不透明，不可枚举，不过期
- 可排序（时间递增），适合 B+ 树索引与范围查询
- 表注释标注 `Snowflake ID`

### 层级与排序约定

- WBS 三级层级：`parent_id` 自引用（Epic→Feature→Story / 主任务→子任务→子子任务 / 主缺陷→子缺陷→子子缺陷），应用层限制最大深度 3 级并做循环校验
- 同级排序：`sort_order DOUBLE PRECISION DEFAULT 65535`（小数插值重排，避免批量 UPDATE）
- 知识文档树、项目文档目录同样使用 `parent_id` + `sort_order`

---

## ER 总览

```
tenants ──< users ──< user_roles >── roles ──< role_menus >── menus
  │           ├──< tenant_members（空间四级角色 owner/admin/member/guest）
  │           └──< user_preferences
  │
  ├──< project_groups ──< projects ──< project_members / project_configs / project_sequences
  │        │                    （project_templates 为创建期预设，不运行时关联）
  │        ├──< states / state_transitions（按 project_id + entity_type 隔离）
  │        ├──< estimate_points
  │        ├──< requirement / task / defect
  │        │        ├──< requirement_assignees / task_assignees / defect_assignees
  │        │        ├──< requirement_labels / task_labels / defect_labels
  │        │        ├──< requirement_watchers / task_watchers / defect_watchers
  │        │        ├──< requirement_comments / task_comments / defect_comments
  │        │        ├──< requirement_activities / task_activities / defect_activities
  │        │        ├──< requirement_attachments / task_attachments / defect_attachments
  │        │        ├──< requirement_timelogs / task_timelogs / defect_timelogs
  │        │        ├──< requirement_modules / task_modules / defect_modules
  │        │        ├──< requirement_relations / task_relations / defect_relations
  │        │        ├──< task_dependencies / cross_entity_relations
  │        │        └──< code_links（commit/branch/PR 关联）
  │        ├──< modules / labels
  │        ├──< versions ──< version_sprint_relations >── sprints
  │        │        ├──< version_delivery_snapshots
  │        │        ├──< version_reports（交付报告/Release Notes）
  │        │        └──< sprint_requirements / sprint_tasks / sprint_defects
  │        │        └──< sprint_snapshots
  │        ├──< intake_channels ──< intake_issues
  │        └──< content_templates（需求/任务/缺陷/文档/知识库模板）
  │
  ├──< reviews ──< review_assignments（多态：需求评审/文档评审/知识库评审/迭代复盘）
  │        └── review_templates
  ├──< knowledge_spaces ──< knowledge_space_members
  │        └──< knowledge_pages ──< knowledge_page_versions
  │        │        └──< knowledge_page_requirements / knowledge_page_tasks / knowledge_page_defects
  │        │        └──< knowledge_page_labels / knowledge_page_comments / knowledge_page_attachments
  ├──< documents ──< document_versions / document_links / document_comments / document_attachments
  ├──< share_links（多态分享：document/knowledge_page/page/dashboard/saved_view）
  ├──< notifications ──< notification_deliveries / notification_subscriptions / notification_digests / notification_preferences
  ├──< dashboards ──< dashboard_widgets ──< dashboard_widget_data
  │        └──< dashboard_templates / dashboard_snapshots
  ├──< workbench_configs / workbench_todos / workbench_templates
  ├──< recent_items / view_preferences / saved_views / favorite_items
  ├──< search_history / search_bookmarks / search_documents（ES降级全文检索）
  ├──< calendar_events（站会/评审会/个人日程）
  ├──< automation_rules ──< rule_executions / automation_templates
  ├──< webhooks ──< webhook_logs
  ├──< sso_providers ──< sso_sessions / sso_links（OIDC/SAML 账号绑定）
  ├──< biz_entity_relations（跨类型关联：需求↔任务↔缺陷）
  ├──< data_jobs（导入/导出/归档）
  └──< audit_logs / domain_events
```

---

## 1、租户与权限域

### 1. `tenants` — 租户（组织机构）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| code | VARCHAR(50) | UNIQUE | 租户编码 |
| name | VARCHAR(255) | NOT NULL | 组织名称 |
| slug | VARCHAR(100) | NOT NULL, UNIQUE | 域名/URL标识 |
| logo_url | TEXT | | 组织Logo |
| owner_id | BIGINT | NOT NULL, 逻辑外键 → users.id | 空间 Owner（PRD 1.4.1：创建人自动成为 Owner） |
| timezone | VARCHAR(50) | DEFAULT 'Asia/Shanghai' | 默认时区（PRD 1.4.1） |
| language | VARCHAR(20) | DEFAULT 'zh-CN' | 默认语言（多语言） |
| brand_config | JSONB | | 品牌定制（主题色/登录页，PRD 品牌定制） |
| status | VARCHAR(50) | NOT NULL DEFAULT 'active' CHECK (active/disabled/archived/expired) | archived=归档只读可恢复 |
| max_projects | INT | DEFAULT 10 | 最大项目数 |
| max_users | INT | DEFAULT 50 | 最大用户数 |
| expired_at | TIMESTAMPTZ | | 服务到期时间 |
| config | JSONB | | 租户级配置 |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

> 说明：tenants 为基础租户表，不含基础字段模板中的 deleted/tenant_id/created_by/updated_by（它是租户级根表）。

### 2. `users` — 用户

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| ~~基础字段~~ | — | — | 见模板 |
| id | BIGINT | PK | 雪花ID |
| code | VARCHAR(50) | 租户内唯一 | 工号/用户编码 |
| name | VARCHAR(255) | NOT NULL | 显示名 |
| email | VARCHAR(255) | NOT NULL | 邮箱 |
| phone | VARCHAR(20) | | 手机号 |
| password_hash | TEXT | | 密码哈希 |
| avatar_url | TEXT | | 头像URL |
| status | VARCHAR(50) | NOT NULL DEFAULT 'active' CHECK (active/inactive/locked) | 状态 |
| is_super_admin | BOOLEAN | DEFAULT false | 系统级超管 |
| last_login_at | TIMESTAMPTZ | | 最后登录时间 |
| deleted | BOOLEAN | DEFAULT false | 软删除标记（true=已删除） |
| tenant_id | BIGINT | NOT NULL | 归属租户 |
| created_by | BIGINT | NOT NULL | 创建人（逻辑外键 → users.id） |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_by | BIGINT | NOT NULL | 最后更新人（逻辑外键 → users.id） |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

> UNIQUE(tenant_id, email) / UNIQUE(tenant_id, phone) WHERE NOT deleted — 租户内唯一

### 3. `roles` — 角色

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| ~~基础字段~~ | — | — | 见模板 |
| id | BIGINT | PK | 雪花ID |
| code | VARCHAR(50) | 租户内唯一 | 角色编码（PROJECT_ADMIN / DEVELOPER / TESTER） |
| name | VARCHAR(255) | NOT NULL | 角色名称 |
| description | TEXT | | 描述 |
| status | VARCHAR(50) | NOT NULL DEFAULT 'active' | 状态 |
| is_system | BOOLEAN | DEFAULT false | 系统内置角色不可删除 |
| role_scope | VARCHAR(50) | DEFAULT 'tenant' CHECK (tenant/project) | 作用范围 |
| deleted | BOOLEAN | DEFAULT false | 软删除标记（true=已删除） |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |
| created_by | BIGINT | NOT NULL | 创建人（逻辑外键 → users.id） |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_by | BIGINT | NOT NULL | 最后更新人（逻辑外键 → users.id） |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

### 4. `menus` — 菜单 / 权限资源

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| ~~基础字段~~ | — | — | 见模板 |
| id | BIGINT | PK | 雪花ID |
| code | VARCHAR(100) | NOT NULL, UNIQUE | 权限编码（project:issue:create） |
| name | VARCHAR(255) | NOT NULL | 菜单/按钮名称 |
| menu_type | VARCHAR(20) | NOT NULL CHECK (menu/button/api) | 资源类型 |
| parent_id | BIGINT | 逻辑外键 → menus.id | 父菜单（0=根） |
| path | VARCHAR(255) | | 路由路径 |
| icon | VARCHAR(100) | | 图标 |
| sort_order | INT | DEFAULT 0 | 排序 |
| status | VARCHAR(50) | DEFAULT 'active' | 状态 |
| deleted | BOOLEAN | DEFAULT false | 菜单一般为系统级，不使用 tenant_id |
| created_by | BIGINT | NOT NULL | 创建人（逻辑外键 → users.id） |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_by | BIGINT | NOT NULL | 最后更新人（逻辑外键 → users.id） |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

> 说明：menus 为系统级资源表，不含 tenant_id。通过 role_menus 与角色关联。

### 5. `user_roles` — 用户角色关联

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| user_id | BIGINT | NOT NULL, 逻辑外键 → users.id | 用户 ID（逻辑外键 → users.id） |
| role_id | BIGINT | NOT NULL, 逻辑外键 → roles.id | 角色 ID（逻辑外键 → roles.id） |
| project_id | BIGINT | 逻辑外键 → projects.id（nullable） | 项目级角色可为空 |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |
| created_by | BIGINT | NOT NULL | 创建人（逻辑外键 → users.id） |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |

> UNIQUE(tenant_id, user_id, role_id, project_id) — 同一用户不重复授予

### 6. `role_menus` — 角色菜单关联

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| role_id | BIGINT | NOT NULL, 逻辑外键 → roles.id | 角色 ID（逻辑外键 → roles.id） |
| menu_id | BIGINT | NOT NULL, 逻辑外键 → menus.id | 菜单 ID（逻辑外键 → menus.id） |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |

> UNIQUE(role_id, menu_id)

---

## 2、项目管理域

### 7. `projects` — 项目

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| ~~基础字段~~ | — | — | 见模板 |
| id | BIGINT | PK | 雪花ID |
| code | VARCHAR(50) | NOT NULL | 项目标识（YD / PLANE） |
| name | VARCHAR(255) | NOT NULL | 项目名称 |
| description | TEXT | | 描述 |
| project_type | VARCHAR(50) | DEFAULT 'scrum' CHECK (scrum/kanban/waterfall) | 项目类型（Scrum/Kanban/瀑布） |
| status | VARCHAR(50) | DEFAULT 'active' CHECK (active/archived) | 状态 |
| logo_props | JSONB | | Logo/Emoji/封面属性 |
| lead_id | BIGINT | 逻辑外键 → users.id | 项目负责人 |
| network | VARCHAR(20) | NOT NULL DEFAULT 'private' CHECK (private/public/internal) | 网络类型（PRD 2.5：私有/公开/内部，替代布尔 is_public） |
| group_id | BIGINT | 逻辑外键 → project_groups.id | 所属项目集/分组（PRD 2.2，nullable） |
| template_id | BIGINT | 逻辑外键 → project_templates.id | 创建来源模板（追溯用） |
| default_assignee_id | BIGINT | 逻辑外键 → users.id | 默认负责人 |
| progress | SMALLINT | DEFAULT 0 CHECK (0-100) | 进度百分比 0-100 |
| sort_order | DOUBLE PRECISION | DEFAULT 65535 | 同级排序值（小数插值重排） |
| deleted | BOOLEAN | DEFAULT false | 软删除标记（true=已删除） |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |
| created_by | BIGINT | NOT NULL | 创建人（逻辑外键 → users.id） |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_by | BIGINT | NOT NULL | 最后更新人（逻辑外键 → users.id） |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |
| version | INT | DEFAULT 1 | 乐观锁版本号（并发控制） |

> UNIQUE(tenant_id, code) — 租户内code唯一

### 8. `project_members` — 项目成员

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| project_id | BIGINT | NOT NULL, 逻辑外键 | 项目 ID（逻辑外键 → projects.id） |
| user_id | BIGINT | NOT NULL, 逻辑外键 | 用户 ID（逻辑外键 → users.id） |
| role | VARCHAR(50) | NOT NULL CHECK (admin/developer/tester/viewer) | 项目内角色 |
| join_type | VARCHAR(20) | DEFAULT 'direct' CHECK (direct/invitation) | 加入方式（直接/邀请） |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |
| joined_at | TIMESTAMPTZ | DEFAULT now() | 加入时间 |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |

> UNIQUE(tenant_id, project_id, user_id)

### 9. `project_configs` — 项目功能模块开关

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| project_id | BIGINT | NOT NULL | 项目 ID（逻辑外键 → projects.id） |
| config_key | VARCHAR(100) | NOT NULL | 模块键（intake/sprint/version/estimate/pages/knowledge） |
| enabled | BOOLEAN | DEFAULT true | 功能模块是否启用 |
| config_json | JSONB | | 模块级配置 |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

> UNIQUE(tenant_id, project_id, config_key)

### 10. `project_sequences` — 实体发号器

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| project_id | BIGINT | NOT NULL | 项目 ID（逻辑外键 → projects.id） |
| entity_type | VARCHAR(20) | NOT NULL CHECK (requirement/task/defect) | 实体类型（需求/任务/缺陷） |
| next_value | BIGINT | DEFAULT 1 | 下一个序号 |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

> UNIQUE(tenant_id, project_id, entity_type) — 同一项目每种类型独立发号

---

## 3、状态机域

### 11. `states` — 状态

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| ~~基础字段~~ | — | — | 见模板 |
| id | BIGINT | PK | 雪花ID |
| code | VARCHAR(50) | | 状态编码 |
| slug | VARCHAR(100) | | 机器名（API/过滤用，PRD 6.1） |
| name | VARCHAR(100) | NOT NULL | 状态名 |
| description | TEXT | | 描述 |
| color | VARCHAR(7) | | 色值 HEX |
| entity_type | VARCHAR(20) | NOT NULL CHECK (requirement/task/defect) | 适用类型（需求/任务/缺陷独立） |
| project_id | BIGINT | 逻辑外键 → projects.id | 所属项目；NULL = 租户级默认状态（PRD 6.1：每项目独立状态机） |
| state_group | VARCHAR(20) | NOT NULL CHECK (backlog/unstarted/started/completed/cancelled/triage) | 状态分组（待办/未开始/进行中/完成/取消/分类） |
| is_triage | BOOLEAN | DEFAULT false | 是否分诊态 |
| is_default | BOOLEAN | DEFAULT false | 是否新建默认 |
| sort_order | INT | DEFAULT 0 | 同级排序值（小数插值重排） |
| status | VARCHAR(50) | DEFAULT 'active' | 状态 |
| deleted | BOOLEAN | DEFAULT false | 软删除标记（true=已删除） |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |
| created_by | BIGINT | NOT NULL | 创建人（逻辑外键 → users.id） |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_by | BIGINT | NOT NULL | 最后更新人（逻辑外键 → users.id） |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

### 12. `state_transitions` — 状态流转规则

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| ~~基础字段~~ | — | — | 见模板 |
| id | BIGINT | PK | 雪花ID |
| entity_type | VARCHAR(20) | NOT NULL CHECK (requirement/task/defect) | 实体类型（requirement/task/defect） |
| project_id | BIGINT | 逻辑外键 → projects.id | 项目级流转规则；NULL = 租户级默认 |
| from_state_id | BIGINT | 逻辑外键 | 起始状态 |
| to_state_id | BIGINT | 逻辑外键 | 目标状态 |
| required_fields | JSONB | DEFAULT '[]' | 必填字段（如缺陷解决时必填 root_cause_category） |
| allowed_roles | JSONB | DEFAULT '[]' | 允许操作的角色 |
| status | VARCHAR(50) | DEFAULT 'active' | 状态 |
| deleted | BOOLEAN | DEFAULT false | 软删除标记（true=已删除） |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |
| created_by | BIGINT | NOT NULL | 创建人（逻辑外键 → users.id） |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_by | BIGINT | NOT NULL | 最后更新人（逻辑外键 → users.id） |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

> UNIQUE(tenant_id, project_id, entity_type, from_state_id, to_state_id)

---

## 4、估算域

### 13. `estimate_points` — 估算体系

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| ~~基础字段~~ | — | — | 见模板 |
| id | BIGINT | PK | 雪花ID |
| code | VARCHAR(50) | | 业务短代码标识（用户按规则生成） |
| name | VARCHAR(255) | NOT NULL | 名称 |
| description | TEXT | | 描述 |
| points_config | JSONB | NOT NULL | 估算选项 [{"label":"1点","value":1}] |
| is_default | BOOLEAN | DEFAULT false | 是否默认估算体系 |
| status | VARCHAR(50) | DEFAULT 'active' | 状态 |
| deleted | BOOLEAN | DEFAULT false | 软删除标记（true=已删除） |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |
| created_by | BIGINT | NOT NULL | 创建人（逻辑外键 → users.id） |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_by | BIGINT | NOT NULL | 最后更新人（逻辑外键 → users.id） |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

---

## 5、版本日域

### 14. `versions` — 版本日

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| ~~基础字段~~ | — | — | 见模板 |
| id | BIGINT | PK | 雪花ID |
| code | VARCHAR(50) | | 版本编码 |
| name | VARCHAR(255) | NOT NULL | 版本名称 |
| description | TEXT | | 描述 |
| project_id | BIGINT | NOT NULL, 逻辑外键 → projects.id | 归属项目（版本日是项目内对象） |
| version_number | VARCHAR(50) | NOT NULL | 语义化版本号（semver: 1.2.3，应用层校验） |
| template_type | VARCHAR(20) | DEFAULT 'standard' CHECK (standard/hotfix/major) | 发布模板类型（PRD 3.2：常规/热修复/大版本） |
| status | VARCHAR(50) | DEFAULT 'planning' CHECK (planning/active/released/archived) | 状态 |
| start_date | DATE | | 开始日期 |
| release_date | DATE | | 发布日期 |
| released_at | TIMESTAMPTZ | | 实际发布时间 |
| owner_id | BIGINT | 逻辑外键 → users.id | 版本负责人 |
| checklist | JSONB | | 发布检查清单 |
| release_notes | TEXT | | 发布说明 |
| progress_snapshot | JSONB | | 进度快照（缓存） |
| deleted | BOOLEAN | DEFAULT false | 软删除标记（true=已删除） |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |
| created_by | BIGINT | NOT NULL | 创建人（逻辑外键 → users.id） |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_by | BIGINT | NOT NULL | 最后更新人（逻辑外键 → users.id） |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |
| version | INT | DEFAULT 1 | 乐观锁版本号（并发控制） |

### 15. `version_delivery_snapshots` — 版本日交付快照

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| version_id | BIGINT | NOT NULL | 版本日 ID（逻辑外键 → versions.id） |
| snapshot_date | DATE | NOT NULL | 快照日期 |
| total_count | INT | | 总数量 |
| completed_count | INT | | 已完成数量 |
| total_points | NUMERIC(10,2) | | 总故事点 |
| completed_points | NUMERIC(10,2) | | 已完成故事点 |
| defect_count | INT | | 缺陷数量 |
| details | JSONB | | 详细数据 |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |

> UNIQUE(tenant_id, version_id, snapshot_date)

---

## 6、迭代域

### 16. `sprints` — 迭代

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| ~~基础字段~~ | — | — | 见模板 |
| id | BIGINT | PK | 雪花ID |
| code | VARCHAR(50) | | 业务短代码标识（用户按规则生成） |
| name | VARCHAR(255) | NOT NULL | 名称 |
| goal | TEXT | | 迭代目标 |
| description | TEXT | | 描述 |
| project_id | BIGINT | NOT NULL, 逻辑外键 → projects.id | 归属项目 |
| sprint_number | INT | NOT NULL | 迭代序号 |
| start_date | DATE | NOT NULL | 开始日期 |
| end_date | DATE | NOT NULL | 结束日期 |
| started_at | TIMESTAMPTZ | | 实际启动时间 |
| completed_at | TIMESTAMPTZ | | 实际完成时间 |
| close_strategy | VARCHAR(20) | CHECK (backlog/next_sprint) | 结束时未完成需求/任务/缺陷的处理方式 |
| retrospective_document_id | BIGINT | 逻辑外键 → documents.id | 复盘报告文档（PRD 4.4.3 自动复盘） |
| owner_id | BIGINT | 逻辑外键 → users.id | 迭代负责人（PRD 5.2 owned_by） |
| status | VARCHAR(50) | DEFAULT 'planned' CHECK (planned/active/completed/cancelled) | 状态 |
| capacity | NUMERIC(8,2) | | 容量（人天） |
| velocity | NUMERIC(8,2) | | 速率 |
| display_filters | JSONB | | 显示过滤配置 |
| viewport | JSONB | | 视图状态 |
| progress_snapshot | JSONB | | 进度快照缓存 |
| is_mid_sprint_change | BOOLEAN | DEFAULT false | 中途变更标记 |
| deleted | BOOLEAN | DEFAULT false | 软删除标记（true=已删除） |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |
| created_by | BIGINT | NOT NULL | 创建人（逻辑外键 → users.id） |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_by | BIGINT | NOT NULL | 最后更新人（逻辑外键 → users.id） |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |
| version | INT | DEFAULT 1 | 乐观锁版本号（并发控制） |

### 17. `version_sprint_relations` — 版本日-迭代关联

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| version_id | BIGINT | NOT NULL | 版本日 ID（逻辑外键 → versions.id） |
| sprint_id | BIGINT | NOT NULL | 迭代 ID（逻辑外键 → sprints.id） |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |

> UNIQUE(tenant_id, version_id, sprint_id)

### 18. `sprint_snapshots` — 迭代每日快照（燃尽图数据）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| sprint_id | BIGINT | NOT NULL | 迭代 ID（逻辑外键 → sprints.id） |
| snapshot_date | DATE | NOT NULL | 快照日期 |
| remaining_points | NUMERIC(8,2) | | 剩余故事点 |
| completed_points | NUMERIC(8,2) | | 已完成故事点 |
| remaining_count | INT | | 剩余数量 |
| completed_count | INT | | 已完成数量 |
| details | JSONB | | 快照明细（各状态分布等 JSON） |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |

> UNIQUE(tenant_id, sprint_id, snapshot_date)

### 19. `sprint_requirements` — 迭代-需求关联
### 20. `sprint_tasks` — 迭代-任务关联
### 21. `sprint_defects` — 迭代-缺陷关联

三表结构一致（仅逻辑外键目标表不同），以 `sprint_requirements` 为模板：

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| sprint_id | BIGINT | NOT NULL, 逻辑外键 → sprints.id | 迭代 ID（逻辑外键 → sprints.id） |
| requirement_id | BIGINT | NOT NULL, 逻辑外键 → requirement.id | 需求 ID（逻辑外键 → requirement.id） |
| is_mid_sprint | BOOLEAN | DEFAULT false | 中途加入 |
| added_by | BIGINT | 逻辑外键 → users.id | 加入人 ID |
| added_at | TIMESTAMPTZ | DEFAULT now() | 加入时间 |
| removed_at | TIMESTAMPTZ | | 中途移除（复盘统计） |
| removed_by | BIGINT | 逻辑外键 → users.id | 移除人 ID |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |

> UNIQUE(tenant_id, sprint_id, requirement_id)
> sprint_tasks: requirement_id → task_id
> sprint_defects: requirement_id → defect_id

---

## 7、需求表

### 22. `requirement` — 需求

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| ~~基础字段~~ | — | — | 见模板 |
| id | BIGINT | PK | 雪花ID |
| public_id | UUID | NOT NULL DEFAULT gen_random_uuid() | 外部引用UUID（不外泄数据库自增ID） |
| code | VARCHAR(50) | NOT NULL | 需求编码（YD-1001） |
| name | VARCHAR(500) | NOT NULL | 需求标题 |
| description_json | JSONB | | 富文本文档模型 |
| description_html | TEXT | | 富文本 HTML 描述 |
| description_stripped | TEXT | | 纯文本（检索） |
| status | | | 继承基础字段 |
| priority | VARCHAR(20) | DEFAULT 'medium' CHECK (urgent/high/medium/low/none) | 优先级（紧急/高/中/低/无） |
| severity | SMALLINT | CHECK (1-5) | 业务影响严重程度：致命/严重/一般/提示/建议（PRD 4.1 建议补充，选填） |
| requirement_type | VARCHAR(50) | DEFAULT 'functional' CHECK (functional/non_functional/bug_fix/optimization) | 需求类型（功能/非功能/缺陷修复/优化） |
| source | VARCHAR(50) | CHECK (customer/internal/competitor/other) | 来源 |
| acceptance_criteria | JSONB | | 验收标准 |
| story_points | SMALLINT | CHECK (0-12) | 故事点 |
| progress | SMALLINT | DEFAULT 0 CHECK (0-100) | 进度百分比 0-100 |
| project_id | BIGINT | NOT NULL | 项目 ID（逻辑外键 → projects.id） |
| module_id | BIGINT | | 模块 |
| state_id | BIGINT | NOT NULL | 状态 ID（逻辑外键 → states.id） |
| sprint_id | BIGINT | | 迭代 ID（逻辑外键 → sprints.id） |
| version_id | BIGINT | | 首次发布版本 |
| estimate_id | BIGINT | | 估算方案 |
| parent_id | BIGINT | 逻辑外键 → requirement.id | 父需求（WBS） |
| depth | SMALLINT | NOT NULL DEFAULT 1 CHECK (1-3) | WBS层级深度（1=顶层 2=子需求 3=孙需求） |
| assignee_id | BIGINT | | 主要负责人 |
| start_date | DATE | | 开始日期 |
| target_date | DATE | | 目标/截止日期 |
| completed_at | TIMESTAMPTZ | | 完成时间 |
| is_draft | BOOLEAN | DEFAULT false | 是否草稿 |
| sort_order | DOUBLE PRECISION | DEFAULT 65535 | 同级排序值（小数插值重排） |
| external_source | VARCHAR(50) | | 外部系统来源（jira/tapd/csv_import，增量导入用） |
| external_id | VARCHAR(255) | | 外部系统ID（增量导入去重键，PRD 9.5） |
| custom_fields | JSONB | | 自定义字段扩展（已废弃，迁移至 work_item_custom_fields 表，保留向后兼容） |
| archived_at | TIMESTAMPTZ | | 归档时间（NULL=未归档；归档后从默认列表隐藏，支持回收站恢复） |
| version | INT | NOT NULL DEFAULT 1 | 乐观锁版本号（UPDATE 条件带 version，冲突返回 409） |
| deleted | BOOLEAN | DEFAULT false | 软删除标记（true=已删除） |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |
| created_by | BIGINT | NOT NULL | 创建人（逻辑外键 → users.id） |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_by | BIGINT | NOT NULL | 最后更新人（逻辑外键 → users.id） |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

> UNIQUE(tenant_id, code)
> UNIQUE(tenant_id, external_source, external_id) WHERE external_id IS NOT NULL — 导入幂等
> UNIQUE(tenant_id, public_id) WHERE NOT deleted — 外部引用唯一

### 23. `task` — 任务

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| ~~基础字段~~ | — | — | 见模板 |
| id | BIGINT | PK | 雪花ID |
| public_id | UUID | NOT NULL DEFAULT gen_random_uuid() | 外部引用UUID |
| code | VARCHAR(50) | NOT NULL | 任务编码 |
| name | VARCHAR(500) | NOT NULL | 名称 |
| description_json | JSONB | | JSON 结构描述 |
| description_html | TEXT | | 富文本 HTML 描述 |
| description_stripped | TEXT | | 纯文本描述（搜索用） |
| status | | | 状态 |
| priority | VARCHAR(20) | DEFAULT 'medium' | 优先级（高/中/低等） |
| task_type | VARCHAR(50) | DEFAULT 'development' CHECK (development/testing/documentation/design/other) | 任务类型（开发/测试/文档/设计/其他） |
| category | VARCHAR(50) | CHECK (frontend/backend/qa/doc/design/other) | 任务分类（前端/后端/QA/文档/设计） |
| estimated_effort | NUMERIC(8,2) | | 预估工时（小时，PRD 6.3.2 工时管理） |
| actual_effort | NUMERIC(8,2) | | 实际工时（小时，汇总自 entity_timelogs） |
| remaining_effort | NUMERIC(8,2) | | 剩余工时（默认 = 预估 - 实际，可手动修正） |
| story_points | SMALLINT | | 故事点估算 |
| progress | SMALLINT | DEFAULT 0 CHECK (0-100) | 进度百分比 0-100 |
| delay_reason | VARCHAR(50) | CHECK (requirement_change/resource/blocked/other) | 延期原因（需求变更/资源/阻塞/其他） |
| project_id | BIGINT | NOT NULL | 项目 ID（逻辑外键 → projects.id） |
| module_id | BIGINT | | 模块 ID（逻辑外键 → modules.id） |
| state_id | BIGINT | NOT NULL | 状态 ID（逻辑外键 → states.id） |
| sprint_id | BIGINT | | 迭代 ID（逻辑外键 → sprints.id） |
| version_id | BIGINT | | 版本日 ID（逻辑外键 → versions.id） |
| estimate_id | BIGINT | | 估算体系 ID（逻辑外键 → estimate_points.id） |
| parent_id | BIGINT | 逻辑外键 → task.id | 父任务 |
| depth | SMALLINT | NOT NULL DEFAULT 1 CHECK (1-3) | WBS层级深度 |
| requirement_id | BIGINT | 逻辑外键 → requirement.id | 归属需求 |
| assignee_id | BIGINT | | 负责人 ID（逻辑外键 → users.id） |
| start_date | DATE | | 开始日期 |
| target_date | DATE | | 目标/截止日期 |
| completed_at | TIMESTAMPTZ | | 完成时间 |
| sort_order | DOUBLE PRECISION | DEFAULT 65535 | 同级排序值（小数插值重排） |
| external_source | VARCHAR(50) | | 外部系统来源 |
| external_id | VARCHAR(255) | | 外部系统ID（导入去重） |
| custom_fields | JSONB | | 自定义字段扩展（已废弃，迁移至 work_item_custom_fields 表） |
| archived_at | TIMESTAMPTZ | | 归档时间 |
| version | INT | NOT NULL DEFAULT 1 | 乐观锁版本号 |
| deleted | BOOLEAN | DEFAULT false | 软删除标记（true=已删除） |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |
| created_by | BIGINT | NOT NULL | 创建人（逻辑外键 → users.id） |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_by | BIGINT | NOT NULL | 最后更新人（逻辑外键 → users.id） |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

> UNIQUE(tenant_id, code)
> UNIQUE(tenant_id, external_source, external_id) WHERE external_id IS NOT NULL
> UNIQUE(tenant_id, public_id) WHERE NOT deleted

### 24. `defect` — 缺陷

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| ~~基础字段~~ | — | — | 见模板 |
| id | BIGINT | PK | 雪花ID |
| public_id | UUID | NOT NULL DEFAULT gen_random_uuid() | 外部引用UUID |
| code | VARCHAR(50) | NOT NULL | 缺陷编码 |
| name | VARCHAR(500) | NOT NULL | 名称 |
| description_json | JSONB | | JSON 结构描述 |
| description_html | TEXT | | 富文本 HTML 描述 |
| description_stripped | TEXT | | 纯文本描述（搜索用） |
| status | | | 状态 |
| priority | VARCHAR(20) | DEFAULT 'medium' | 优先级（高/中/低等） |
| severity | SMALLINT | NOT NULL CHECK (1-5) | 致命/严重/一般/提示/建议 |
| defect_type | VARCHAR(50) | CHECK (functional/performance/ui/security/compatibility/other) | 缺陷类型（功能/性能/UI/安全/兼容性等） |
| found_phase | VARCHAR(50) | NOT NULL CHECK (unit/integration/uat/production/customer) | 发现阶段（单元/集成/UAT/线上/客户） |
| root_cause_category | VARCHAR(50) | CHECK (requirement/technical/environment/data/other) | 解决时必填（经 state_transitions.required_fields 强制） |
| resolution | VARCHAR(30) | CHECK (fixed/cannot_reproduce/wont_fix/duplicate/by_design/external) | 解决方案（PRD 4.3 Resolution） |
| environment | JSONB | | 环境信息 |
| reproduce_steps | JSONB | NOT NULL | 复现步骤 {steps, expected, actual} |
| fix_steps | JSONB | | 修复步骤（解决时填写） |
| regression_risk | VARCHAR(20) | CHECK (low/medium/high) | 回归风险评级 |
| is_draft | BOOLEAN | DEFAULT false | 是否草稿 |
| project_id | BIGINT | NOT NULL | 项目 ID（逻辑外键 → projects.id） |
| module_id | BIGINT | | 模块 ID（逻辑外键 → modules.id） |
| state_id | BIGINT | NOT NULL | 状态 ID（逻辑外键 → states.id） |
| sprint_id | BIGINT | | 迭代 ID（逻辑外键 → sprints.id） |
| parent_id | BIGINT | 逻辑外键 → defect.id | 父缺陷（WBS） |
| depth | SMALLINT | NOT NULL DEFAULT 1 CHECK (1-3) | WBS层级深度 |
| found_version_id | BIGINT | | 发现版本 |
| fix_version_id | BIGINT | | 修复版本 |
| requirement_id | BIGINT | | 关联需求 |
| task_id | BIGINT | | 关联任务 |
| verifier_id | BIGINT | | 验证人 |
| assignee_id | BIGINT | | 处理人 ID（逻辑外键 → users.id） |
| start_date | DATE | | 开始日期 |
| target_date | DATE | | 目标/截止日期 |
| resolved_at | TIMESTAMPTZ | | 修复时间（缺陷龄/MTTR 分析） |
| resolved_by | BIGINT | 逻辑外键 → users.id | 修复人 |
| verified_at | TIMESTAMPTZ | | 验证通过时间 |
| reopen_count | INT | DEFAULT 0 | 重新打开次数（质量分析：返工率） |
| completed_at | TIMESTAMPTZ | | 完成时间 |
| sort_order | DOUBLE PRECISION | DEFAULT 65535 | 同级排序值（小数插值重排） |
| external_source | VARCHAR(50) | | 外部系统来源 |
| external_id | VARCHAR(255) | | 外部系统ID（导入去重） |
| custom_fields | JSONB | | 自定义字段扩展（已废弃，迁移至 work_item_custom_fields 表） |
| archived_at | TIMESTAMPTZ | | 归档时间 |
| version | INT | NOT NULL DEFAULT 1 | 乐观锁版本号 |
| deleted | BOOLEAN | DEFAULT false | 软删除标记（true=已删除） |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |
| created_by | BIGINT | NOT NULL | 创建人（逻辑外键 → users.id） |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_by | BIGINT | NOT NULL | 最后更新人（逻辑外键 → users.id） |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

> UNIQUE(tenant_id, code)
> UNIQUE(tenant_id, external_source, external_id) WHERE external_id IS NOT NULL
> UNIQUE(tenant_id, public_id) WHERE NOT deleted

---

## 8、需求关联表

> 以下每张关联表的逻辑外键目标明确（需求/任务/缺陷各有独立的关联表，无多态 entity_type），均携带 `tenant_id` + 索引。

### 25. `requirement_assignees` — 需求指派人

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| requirement_id | BIGINT | NOT NULL, 逻辑外键 → requirement.id | 需求 ID（逻辑外键 → requirement.id） |
| user_id | BIGINT | NOT NULL, 逻辑外键 → users.id | 用户 ID（逻辑外键 → users.id） |
| role_type | VARCHAR(20) | DEFAULT 'assignee' CHECK (assignee/reviewer/tester) | 角色类型（负责人/评审人/测试人） |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |

> UNIQUE(tenant_id, requirement_id, user_id, role_type)

### 26. `requirement_labels` — 需求标签

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| requirement_id | BIGINT | NOT NULL, 逻辑外键 → requirement.id | 需求 ID（逻辑外键 → requirement.id） |
| label_id | BIGINT | NOT NULL, 逻辑外键 → labels.id | 标签 ID（逻辑外键 → labels.id） |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |

> UNIQUE(tenant_id, requirement_id, label_id)

### 27. `requirement_watchers` — 需求关注人

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| requirement_id | BIGINT | NOT NULL, 逻辑外键 → requirement.id | 需求 ID（逻辑外键 → requirement.id） |
| user_id | BIGINT | NOT NULL, 逻辑外键 → users.id | 用户 ID（逻辑外键 → users.id） |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |

> UNIQUE(tenant_id, requirement_id, user_id)

### 28. `requirement_comments` — 需求评论

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| ~~基础字段~~ | — | — | 见模板 |
| id | BIGINT | PK | 雪花ID |
| code | VARCHAR(50) | | 业务短代码标识（用户按规则生成） |
| name | VARCHAR(255) | | 标题（可空） |
| requirement_id | BIGINT | NOT NULL, 逻辑外键 → requirement.id | 需求 ID（逻辑外键 → requirement.id） |
| content_html | TEXT | | 富文本 HTML 内容 |
| content_stripped | TEXT | | 纯文本内容（搜索用） |
| parent_id | BIGINT | 逻辑外键 → requirement_comments.id | 嵌套回复 |
| status | VARCHAR(50) | DEFAULT 'active' | 状态 |
| deleted | BOOLEAN | DEFAULT false | 软删除标记（true=已删除） |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |
| created_by | BIGINT | NOT NULL | 创建人（逻辑外键 → users.id） |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_by | BIGINT | NOT NULL | 最后更新人（逻辑外键 → users.id） |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

### 29. `requirement_activities` — 需求活动日志（分区表）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID（时间有序） |
| requirement_id | BIGINT | NOT NULL, 逻辑外键 → requirement.id | 需求 ID（逻辑外键 → requirement.id） |
| project_id | BIGINT | NOT NULL | 项目 ID（逻辑外键 → projects.id） |
| verb | VARCHAR(50) | NOT NULL | 行为动词（created/updated/...） |
| field | VARCHAR(100) | | 变更字段名 |
| old_value | TEXT | | 变更前值 |
| new_value | TEXT | | 变更后值 |
| actor_id | BIGINT | NOT NULL | 操作人 ID |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |

> 分区策略：按月 RANGE 分区，12个月后归档冷备。

### 30. `requirement_attachments` — 需求附件

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| ~~基础字段~~ | — | — | 见模板 |
| id | BIGINT | PK | 雪花ID |
| code | VARCHAR(50) | | 业务短代码标识（用户按规则生成） |
| name | VARCHAR(255) | NOT NULL | 文件名 |
| requirement_id | BIGINT | NOT NULL, 逻辑外键 → requirement.id | 需求 ID（逻辑外键 → requirement.id） |
| file_size | BIGINT | | 文件大小（字节） |
| mime_type | VARCHAR(100) | | 文件 MIME 类型 |
| storage_path | TEXT | NOT NULL | 存储路径 |
| storage_type | VARCHAR(20) | DEFAULT 's3' | 存储类型（s3/oss/local） |
| status | VARCHAR(50) | DEFAULT 'active' | 状态 |
| deleted | BOOLEAN | DEFAULT false | 软删除标记（true=已删除） |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |
| created_by | BIGINT | NOT NULL | 创建人（逻辑外键 → users.id） |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_by | BIGINT | NOT NULL | 最后更新人（逻辑外键 → users.id） |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

### 31. `requirement_timelogs` — 需求工时

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| requirement_id | BIGINT | NOT NULL, 逻辑外键 → requirement.id | 需求 ID（逻辑外键 → requirement.id） |
| user_id | BIGINT | NOT NULL | 用户 ID（逻辑外键 → users.id） |
| log_date | DATE | NOT NULL | 记录日期 |
| hours | NUMERIC(6,2) | NOT NULL | 工时数（小时） |
| description | VARCHAR(500) | | 描述 |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |

> UNIQUE(tenant_id, requirement_id, user_id, log_date)

### 32. `requirement_modules` — 需求-模块关联

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| requirement_id | BIGINT | NOT NULL, 逻辑外键 → requirement.id | 需求 ID（逻辑外键 → requirement.id） |
| module_id | BIGINT | NOT NULL, 逻辑外键 → modules.id | 模块 ID（逻辑外键 → modules.id） |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |

> UNIQUE(tenant_id, requirement_id, module_id)

### 33. `requirement_relations` — 需求关联关系

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| requirement_id | BIGINT | NOT NULL, 逻辑外键 → requirement.id | 需求 ID（逻辑外键 → requirement.id） |
| related_requirement_id | BIGINT | NOT NULL, 逻辑外键 → requirement.id | 关联需求 ID |
| relation_type | VARCHAR(50) | NOT NULL CHECK (duplicate/relates_to/blocked_by/implemented_by) | 关联类型（重复/相关/阻塞/实现） |
| description | VARCHAR(500) | | 描述 |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |
| created_by | BIGINT | NOT NULL | 创建人（逻辑外键 → users.id） |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |

> UNIQUE(tenant_id, requirement_id, related_requirement_id, relation_type)

---

## 9、任务关联表

> 任务表独有：`task_dependencies`（FS/SS/FF/SF 四种依赖类型，PRD 6.3.3）。

### 34. `task_assignees` — 任务指派人

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| task_id | BIGINT | NOT NULL, 逻辑外键 → task.id | 任务 ID（逻辑外键 → task.id） |
| user_id | BIGINT | NOT NULL, 逻辑外键 → users.id | 用户 ID（逻辑外键 → users.id） |
| role_type | VARCHAR(20) | DEFAULT 'assignee' CHECK (assignee/reviewer/tester) | 角色类型（负责人/评审人/测试人） |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |

> UNIQUE(tenant_id, task_id, user_id, role_type)

### 35. `task_labels` — 任务标签

| id | BIGINT | PK | 雪花ID |
| task_id | BIGINT | NOT NULL, 逻辑外键 → task.id | 任务 ID（逻辑外键 → task.id） |
| label_id | BIGINT | NOT NULL, 逻辑外键 → labels.id | 标签 ID（逻辑外键 → labels.id） |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |

> UNIQUE(tenant_id, task_id, label_id)

### 36. `task_watchers` — 任务关注人

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| task_id | BIGINT | NOT NULL, 逻辑外键 → task.id | 任务 ID（逻辑外键 → task.id） |
| user_id | BIGINT | NOT NULL, 逻辑外键 → users.id | 用户 ID（逻辑外键 → users.id） |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |

> UNIQUE(tenant_id, task_id, user_id)

### 37. `task_comments` — 任务评论

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| ~~基础字段~~ | — | — | 见模板 |
| id | BIGINT | PK | 雪花ID |
| code | VARCHAR(50) | | 业务短代码标识（用户按规则生成） |
| name | VARCHAR(255) | | 名称 |
| task_id | BIGINT | NOT NULL, 逻辑外键 → task.id | 任务 ID（逻辑外键 → task.id） |
| content_html | TEXT | | 富文本 HTML 内容 |
| content_stripped | TEXT | | 纯文本内容（搜索用） |
| parent_id | BIGINT | 逻辑外键 → task_comments.id | 父级 ID（自引用） |
| status | VARCHAR(50) | DEFAULT 'active' | 状态 |
| deleted | BOOLEAN | DEFAULT false | 软删除标记（true=已删除） |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |
| created_by | BIGINT | NOT NULL | 创建人（逻辑外键 → users.id） |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_by | BIGINT | NOT NULL | 最后更新人（逻辑外键 → users.id） |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

### 38. `task_activities` — 任务活动日志（分区表）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| task_id | BIGINT | NOT NULL, 逻辑外键 → task.id | 任务 ID（逻辑外键 → task.id） |
| project_id | BIGINT | NOT NULL | 项目 ID（逻辑外键 → projects.id） |
| verb | VARCHAR(50) | NOT NULL | 行为动词（created/updated/...） |
| field | VARCHAR(100) | | 变更字段名 |
| old_value | TEXT | | 变更前值 |
| new_value | TEXT | | 变更后值 |
| actor_id | BIGINT | NOT NULL | 操作人 ID |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |

> 分区策略：按月 RANGE 分区。

### 39. `task_attachments` — 任务附件

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| ~~基础字段~~ | — | — | 见模板 |
| id | BIGINT | PK | 雪花ID |
| code | VARCHAR(50) | | 业务短代码标识（用户按规则生成） |
| name | VARCHAR(255) | NOT NULL | 文件名 |
| task_id | BIGINT | NOT NULL, 逻辑外键 → task.id | 任务 ID（逻辑外键 → task.id） |
| file_size | BIGINT | | 文件大小（字节） |
| mime_type | VARCHAR(100) | | 文件 MIME 类型 |
| storage_path | TEXT | NOT NULL | 存储路径 |
| storage_type | VARCHAR(20) | DEFAULT 's3' | 存储类型（s3/oss/local） |
| status | VARCHAR(50) | DEFAULT 'active' | 状态 |
| deleted | BOOLEAN | DEFAULT false | 软删除标记（true=已删除） |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |
| created_by | BIGINT | NOT NULL | 创建人（逻辑外键 → users.id） |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_by | BIGINT | NOT NULL | 最后更新人（逻辑外键 → users.id） |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

### 40. `task_timelogs` — 任务工时

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| task_id | BIGINT | NOT NULL, 逻辑外键 → task.id | 任务 ID（逻辑外键 → task.id） |
| user_id | BIGINT | NOT NULL | 用户 ID（逻辑外键 → users.id） |
| log_date | DATE | NOT NULL | 记录日期 |
| hours | NUMERIC(6,2) | NOT NULL | 工时数（小时） |
| started_at | TIMESTAMPTZ | | 计时开始 |
| ended_at | TIMESTAMPTZ | | 计时结束 |
| description | VARCHAR(500) | | 描述 |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |

> UNIQUE(tenant_id, task_id, user_id, log_date)

### 41. `task_modules` — 任务-模块关联

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| task_id | BIGINT | NOT NULL, 逻辑外键 → task.id | 任务 ID（逻辑外键 → task.id） |
| module_id | BIGINT | NOT NULL, 逻辑外键 → modules.id | 模块 ID（逻辑外键 → modules.id） |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |

> UNIQUE(tenant_id, task_id, module_id)

### 42. `task_dependencies` — 任务依赖

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| task_id | BIGINT | NOT NULL, 逻辑外键 → task.id | 依赖方 |
| depends_on_task_id | BIGINT | NOT NULL, 逻辑外键 → task.id | 被依赖方 |
| dependency_type | VARCHAR(10) | NOT NULL CHECK (FS/SS/FF/SF) | 依赖类型（FS/SS/FF/SF） |
| lag_days | INT | DEFAULT 0 | 滞后天数 |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |

> UNIQUE(tenant_id, task_id, depends_on_task_id)

### 43. `task_relations` — 任务关联关系

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| task_id | BIGINT | NOT NULL, 逻辑外键 → task.id | 任务 ID（逻辑外键 → task.id） |
| related_task_id | BIGINT | NOT NULL, 逻辑外键 → task.id | 关联任务 ID |
| relation_type | VARCHAR(50) | NOT NULL CHECK (duplicate/relates_to/blocked_by) | 关联类型（重复/相关/阻塞） |
| description | VARCHAR(500) | | 描述 |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |
| created_by | BIGINT | NOT NULL | 创建人（逻辑外键 → users.id） |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |

> UNIQUE(tenant_id, task_id, related_task_id, relation_type)

---

## 10、缺陷关联表

### 44. `defect_assignees` — 缺陷指派人

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| defect_id | BIGINT | NOT NULL, 逻辑外键 → defect.id | 缺陷 ID（逻辑外键 → defect.id） |
| user_id | BIGINT | NOT NULL, 逻辑外键 → users.id | 用户 ID（逻辑外键 → users.id） |
| role_type | VARCHAR(20) | DEFAULT 'assignee' CHECK (assignee/reviewer/verifier) | 角色类型（负责人/评审人/验证人） |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |

> UNIQUE(tenant_id, defect_id, user_id, role_type)

### 45. `defect_labels` — 缺陷标签

| id | BIGINT | PK | 雪花ID |
| defect_id | BIGINT | NOT NULL, 逻辑外键 → defect.id | 缺陷 ID（逻辑外键 → defect.id） |
| label_id | BIGINT | NOT NULL, 逻辑外键 → labels.id | 标签 ID（逻辑外键 → labels.id） |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |

> UNIQUE(tenant_id, defect_id, label_id)

### 46. `defect_watchers` — 缺陷关注人

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| defect_id | BIGINT | NOT NULL, 逻辑外键 → defect.id | 缺陷 ID（逻辑外键 → defect.id） |
| user_id | BIGINT | NOT NULL, 逻辑外键 → users.id | 用户 ID（逻辑外键 → users.id） |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |

> UNIQUE(tenant_id, defect_id, user_id)

### 47. `defect_comments` — 缺陷评论

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| ~~基础字段~~ | — | — | 见模板 |
| id | BIGINT | PK | 雪花ID |
| code | VARCHAR(50) | | 业务短代码标识（用户按规则生成） |
| name | VARCHAR(255) | | 名称 |
| defect_id | BIGINT | NOT NULL, 逻辑外键 → defect.id | 缺陷 ID（逻辑外键 → defect.id） |
| content_html | TEXT | | 富文本 HTML 内容 |
| content_stripped | TEXT | | 纯文本内容（搜索用） |
| parent_id | BIGINT | 逻辑外键 → defect_comments.id | 父级 ID（自引用） |
| status | VARCHAR(50) | DEFAULT 'active' | 状态 |
| deleted | BOOLEAN | DEFAULT false | 软删除标记（true=已删除） |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |
| created_by | BIGINT | NOT NULL | 创建人（逻辑外键 → users.id） |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_by | BIGINT | NOT NULL | 最后更新人（逻辑外键 → users.id） |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

### 48. `defect_activities` — 缺陷活动日志（分区表）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| defect_id | BIGINT | NOT NULL, 逻辑外键 → defect.id | 缺陷 ID（逻辑外键 → defect.id） |
| project_id | BIGINT | NOT NULL | 项目 ID（逻辑外键 → projects.id） |
| verb | VARCHAR(50) | NOT NULL | 行为动词（created/updated/...） |
| field | VARCHAR(100) | | 变更字段名 |
| old_value | TEXT | | 变更前值 |
| new_value | TEXT | | 变更后值 |
| actor_id | BIGINT | NOT NULL | 操作人 ID |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |

> 分区策略：按月 RANGE 分区。

### 49. `defect_attachments` — 缺陷附件

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| ~~基础字段~~ | — | — | 见模板 |
| id | BIGINT | PK | 雪花ID |
| code | VARCHAR(50) | | 业务短代码标识（用户按规则生成） |
| name | VARCHAR(255) | NOT NULL | 文件名 |
| defect_id | BIGINT | NOT NULL, 逻辑外键 → defect.id | 缺陷 ID（逻辑外键 → defect.id） |
| file_size | BIGINT | | 文件大小（字节） |
| mime_type | VARCHAR(100) | | 文件 MIME 类型 |
| storage_path | TEXT | NOT NULL | 存储路径 |
| storage_type | VARCHAR(20) | DEFAULT 's3' | 存储类型（s3/oss/local） |
| status | VARCHAR(50) | DEFAULT 'active' | 状态 |
| deleted | BOOLEAN | DEFAULT false | 软删除标记（true=已删除） |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |
| created_by | BIGINT | NOT NULL | 创建人（逻辑外键 → users.id） |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_by | BIGINT | NOT NULL | 最后更新人（逻辑外键 → users.id） |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

### 50. `defect_timelogs` — 缺陷工时

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| defect_id | BIGINT | NOT NULL, 逻辑外键 → defect.id | 缺陷 ID（逻辑外键 → defect.id） |
| user_id | BIGINT | NOT NULL | 用户 ID（逻辑外键 → users.id） |
| log_date | DATE | NOT NULL | 记录日期 |
| hours | NUMERIC(6,2) | NOT NULL | 工时数（小时） |
| description | VARCHAR(500) | | 描述 |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |

> UNIQUE(tenant_id, defect_id, user_id, log_date)

### 51. `defect_modules` — 缺陷-模块关联

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| defect_id | BIGINT | NOT NULL, 逻辑外键 → defect.id | 缺陷 ID（逻辑外键 → defect.id） |
| module_id | BIGINT | NOT NULL, 逻辑外键 → modules.id | 模块 ID（逻辑外键 → modules.id） |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |

> UNIQUE(tenant_id, defect_id, module_id)

### 52. `defect_relations` — 缺陷关联关系

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| defect_id | BIGINT | NOT NULL, 逻辑外键 → defect.id | 缺陷 ID（逻辑外键 → defect.id） |
| related_defect_id | BIGINT | NOT NULL, 逻辑外键 → defect.id | 关联缺陷 ID |
| relation_type | VARCHAR(50) | NOT NULL CHECK (duplicate/relates_to/blocked_by) | 关联类型（重复/相关/阻塞） |
| description | VARCHAR(500) | | 描述 |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |
| created_by | BIGINT | NOT NULL | 创建人（逻辑外键 → users.id） |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |

> UNIQUE(tenant_id, defect_id, related_defect_id, relation_type)

---

## 11、跨类型关联

> 需求↔任务、需求↔缺陷、任务↔缺陷之间的跨类型关联（如"需求测试产生的缺陷"、"任务关联需求"），通过本表承载。同类型关联各自走 `requirement_relations` / `task_relations` / `defect_relations`。

### 53. `cross_entity_relations` — 跨类型关联

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| source_type | VARCHAR(20) | NOT NULL CHECK (requirement/task/defect) | 源实体类型 |
| source_id | BIGINT | NOT NULL | 源实体 ID |
| target_type | VARCHAR(20) | NOT NULL CHECK (requirement/task/defect) | 目标实体类型 |
| target_id | BIGINT | NOT NULL | 目标实体 ID |
| relation_type | VARCHAR(50) | NOT NULL CHECK (duplicate/relates_to/blocked_by/implemented_by) | 关联类型（重复/相关/阻塞/实现） |
| description | VARCHAR(500) | | 描述 |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |
| created_by | BIGINT | NOT NULL | 创建人（逻辑外键 → users.id） |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |

> CHECK(source_type <> target_type) — 仅允许跨类型

---

## 12、标签与模块

### 54. `labels` — 标签

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| ~~基础字段~~ | — | — | 见模板 |
| id | BIGINT | PK | 雪花ID |
| code | VARCHAR(50) | | 业务短代码标识（用户按规则生成） |
| name | VARCHAR(100) | NOT NULL | 名称 |
| color | VARCHAR(7) | | 色值 |
| description | TEXT | | 描述 |
| label_group | VARCHAR(50) | | 分组 |
| project_id | BIGINT | 逻辑外键 → projects.id | 所属项目；NULL = 租户级通用标签 |
| status | VARCHAR(50) | DEFAULT 'active' | 状态 |
| deleted | BOOLEAN | DEFAULT false | 软删除标记（true=已删除） |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |
| created_by | BIGINT | NOT NULL | 创建人（逻辑外键 → users.id） |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_by | BIGINT | NOT NULL | 最后更新人（逻辑外键 → users.id） |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

### 55. `modules` — 模块

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| ~~基础字段~~ | — | — | 见模板 |
| id | BIGINT | PK | 雪花ID |
| code | VARCHAR(50) | | 业务短代码标识（用户按规则生成） |
| name | VARCHAR(255) | NOT NULL | 名称 |
| description | TEXT | | 描述 |
| project_id | BIGINT | 逻辑外键 → projects.id | 所属项目；NULL = 租户级模块 |
| owner_id | BIGINT | | 模块负责人 |
| target_version_id | BIGINT | | 目标交付版本日 |
| status | VARCHAR(50) | DEFAULT 'active' | 状态 |
| deleted | BOOLEAN | DEFAULT false | 软删除标记（true=已删除） |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |
| created_by | BIGINT | NOT NULL | 创建人（逻辑外键 → users.id） |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_by | BIGINT | NOT NULL | 最后更新人（逻辑外键 → users.id） |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |



## 13、收件箱域

### 56. `intake_channels` — 收件箱渠道

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| ~~基础字段~~ | — | — | 见模板 |
| id | BIGINT | PK | 雪花ID |
| code | VARCHAR(50) | | 业务短代码标识（用户按规则生成） |
| name | VARCHAR(255) | NOT NULL | 名称 |
| description | TEXT | | 描述 |
| project_id | BIGINT | NOT NULL, 逻辑外键 → projects.id | 关联项目（PRD 9.3.1：创建 Channel 需关联项目） |
| is_public | BOOLEAN | DEFAULT true | 公开（无需登录）/私有（仅成员） |
| default_entity_type | VARCHAR(20) | DEFAULT 'requirement' | 默认转正实体类型 |
| portal_slug | VARCHAR(100) | UNIQUE | 门户URL |
| auto_assign_rules | JSONB | | 自动分配规则（JSON） |
| status | VARCHAR(50) | DEFAULT 'active' | 状态 |
| deleted | BOOLEAN | DEFAULT false | 软删除标记（true=已删除） |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |
| created_by | BIGINT | NOT NULL | 创建人（逻辑外键 → users.id） |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_by | BIGINT | NOT NULL | 最后更新人（逻辑外键 → users.id） |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

### 57. `intake_issues` — 收件箱提交

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| ~~基础字段~~ | — | — | 见模板 |
| id | BIGINT | PK | 雪花ID |
| code | VARCHAR(50) | | 业务短代码标识（用户按规则生成） |
| name | VARCHAR(255) | NOT NULL | 反馈标题 |
| description_json | JSONB | | JSON 结构描述 |
| description_html | TEXT | | 富文本 HTML 描述 |
| channel_id | BIGINT | NOT NULL, 逻辑外键 → intake_channels.id | 来源渠道 |
| project_id | BIGINT | NOT NULL, 逻辑外键 → projects.id | 归属项目（冗余自 channel，便于查询） |
| tracking_code | VARCHAR(32) | NOT NULL, UNIQUE | 外部跟踪 ID（提交者凭此+邮箱查询进度，PRD 9.3.2） |
| priority | VARCHAR(20) | CHECK (urgent/high/medium/low/none) | 提交者可选填 |
| status | VARCHAR(50) | DEFAULT 'open' CHECK (open/accepted/rejected/converted/archived) | 状态 |
| reject_reason | VARCHAR(500) | | 拒绝原因（通知提交者） |
| duplicate_of_id | BIGINT | 逻辑外键 → intake_issues.id | 重复标记（AI 重复检测，P2） |
| converted_entity_type | VARCHAR(20) | | 转正后类型 |
| converted_entity_id | BIGINT | | 转正后ID |
| submitter_name | VARCHAR(255) | | 外部提交人 |
| submitter_email | VARCHAR(255) | | 提交者邮箱（跟踪查询凭证） |
| submitter_phone | VARCHAR(20) | | 提交者手机号 |
| submitted_at | TIMESTAMPTZ | DEFAULT now() | 提交时间 |
| reviewed_by | BIGINT | | 审核人 ID（逻辑外键 → users.id） |
| reviewed_at | TIMESTAMPTZ | | 审核时间 |
| deleted | BOOLEAN | DEFAULT false | 软删除标记（true=已删除） |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |
| created_by | BIGINT | NOT NULL | 创建人（逻辑外键 → users.id） |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_by | BIGINT | NOT NULL | 最后更新人（逻辑外键 → users.id） |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

---

## 14、知识库域

### 58. `knowledge_spaces` — 知识库空间

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| ~~基础字段~~ | — | — | 见模板 |
| id | BIGINT | PK | 雪花ID |
| code | VARCHAR(50) | | 业务短代码标识（用户按规则生成） |
| name | VARCHAR(255) | NOT NULL | 名称 |
| slug | VARCHAR(100) | NOT NULL | 空间唯一标识（URL） |
| description | TEXT | | 描述 |
| cover_image_url | TEXT | | 封面图 URL |
| is_private | BOOLEAN | DEFAULT false | 是否私有（私有仅成员可见） |
| default_permission | VARCHAR(20) | DEFAULT 'view' | 默认权限（view/edit/comment） |
| owner_id | BIGINT | | 空间负责人 |
| project_id | BIGINT | | 关联项目（nullable） |
| status | VARCHAR(50) | DEFAULT 'active' | 状态 |
| deleted | BOOLEAN | DEFAULT false | 软删除标记（true=已删除） |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |
| created_by | BIGINT | NOT NULL | 创建人（逻辑外键 → users.id） |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_by | BIGINT | NOT NULL | 最后更新人（逻辑外键 → users.id） |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

### 59. `knowledge_pages` — 知识文档

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| ~~基础字段~~ | — | — | 见模板 |
| id | BIGINT | PK | 雪花ID |
| code | VARCHAR(50) | | 业务短代码标识（用户按规则生成） |
| name | VARCHAR(255) | NOT NULL | 名称 |
| content_md | TEXT | | Markdown 原文内容 |
| content_html | TEXT | | 富文本 HTML 内容 |
| content_stripped | TEXT | | 纯文本内容（搜索用） |
| space_id | BIGINT | NOT NULL | 知识库空间 ID（逻辑外键 → knowledge_spaces.id） |
| parent_id | BIGINT | 逻辑外键 → knowledge_pages.id | 目录层级（文档树，PRD 10.4 无限层级） |
| sort_order | DOUBLE PRECISION | DEFAULT 65535 | 同级排序（拖拽排序） |
| status | VARCHAR(50) | DEFAULT 'draft' CHECK (draft/published/archived) | 状态 |
| version_num | INT | DEFAULT 1 | 当前版本号 |
| is_pinned | BOOLEAN | DEFAULT false | 是否置顶 |
| is_featured | BOOLEAN | DEFAULT false | 是否推荐 |
| view_count | INT | DEFAULT 0 | 浏览量统计 |
| reviewer_id | BIGINT | | 评审人 ID（逻辑外键 → users.id） |
| published_at | TIMESTAMPTZ | | 发布时间 |
| deleted | BOOLEAN | DEFAULT false | 软删除标记（true=已删除） |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |
| created_by | BIGINT | NOT NULL | 创建人（逻辑外键 → users.id） |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_by | BIGINT | NOT NULL | 最后更新人（逻辑外键 → users.id） |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

> 标签通过 `knowledge_page_labels`（结构与 `requirement_labels` 一致）承载；订阅复用 `notification_subscriptions`。
> 评审通过 `reviews`（entity_type='knowledge_page'）承载，reviewer_id 仅作当前待审人冗余。

### 60. `knowledge_page_versions` — 知识文档版本
| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| page_id | BIGINT | NOT NULL | 知识文档 ID（逻辑外键 → knowledge_pages.id） |
| version_num | INT | NOT NULL | 版本号 |
| content_md | TEXT | | Markdown 内容快照 |
| content_html | TEXT | | 富文本 HTML 内容 |
| change_summary | VARCHAR(500) | | 版本变更摘要 |
| created_by | BIGINT | NOT NULL | 创建人（逻辑外键 → users.id） |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |

> UNIQUE(tenant_id, page_id, version_num)

### 61. `knowledge_page_requirements` — 知识文档-需求关联
### 62. `knowledge_page_tasks` — 知识文档-任务关联
### 63. `knowledge_page_defects` — 知识文档-缺陷关联

三表结构一致，以 `knowledge_page_requirements` 为模板：

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| page_id | BIGINT | NOT NULL, 逻辑外键 → knowledge_pages.id | 知识文档 ID（逻辑外键 → knowledge_pages.id） |
| requirement_id | BIGINT | NOT NULL, 逻辑外键 → requirement.id | 需求 ID（逻辑外键 → requirement.id） |
| relation_type | VARCHAR(20) | DEFAULT 'references' | 关联类型（引用等） |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |

> UNIQUE(tenant_id, page_id, requirement_id)
> knowledge_page_tasks: requirement_id → task_id; knowledge_page_defects: requirement_id → defect_id

### 64. `knowledge_page_labels` — 知识文档标签
### 65. `knowledge_page_comments` — 知识文档评论
### 66. `knowledge_page_attachments` — 知识文档附件

三表结构与对应的 `requirement_*` 表一致（`requirement_id` → `page_id`，逻辑外键 → `knowledge_pages.id`），此处省略重复字段定义。

---

## 15、文档管理域

### 67. `documents` — 项目文档

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| ~~基础字段~~ | — | — | 见模板 |
| id | BIGINT | PK | 雪花ID |
| code | VARCHAR(50) | | 业务短代码标识（用户按规则生成） |
| name | VARCHAR(255) | NOT NULL | 名称 |
| description_stripped | TEXT | | 纯文本描述（搜索用） |
| content_json | JSONB | | JSON 结构内容 |
| document_category | VARCHAR(50) | DEFAULT 'other' CHECK (prd/design/api/test/checklist/requirements/other) | 文档分类（PRD/设计/接口/测试报告等） |
| document_format | VARCHAR(20) | DEFAULT 'rich_text' CHECK (rich_text/markdown/wiki) | 文档格式（富文本/Markdown/Wiki） |
| project_id | BIGINT | NOT NULL, 逻辑外键 → projects.id | 归属项目 |
| parent_id | BIGINT | 逻辑外键 → documents.id | 目录树父节点（PRD 10.3.1 文档目录） |
| sort_order | DOUBLE PRECISION | DEFAULT 65535 | 目录内拖拽排序 |
| status | VARCHAR(50) | DEFAULT 'draft' CHECK (draft/reviewing/approved/archived) | 状态 |
| entity_type | VARCHAR(20) | | 关联类型 |
| entity_id | BIGINT | | 关联ID |
| sprint_id | BIGINT | | 迭代 ID（逻辑外键 → sprints.id） |
| version_id | BIGINT | | 版本日 ID（逻辑外键 → versions.id） |
| review_required | BOOLEAN | DEFAULT false | 是否需要评审 |
| approved_by | BIGINT | | 审批通过人 ID |
| approved_at | TIMESTAMPTZ | | 审批通过时间 |
| deleted | BOOLEAN | DEFAULT false | 软删除标记（true=已删除） |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |
| created_by | BIGINT | NOT NULL | 创建人（逻辑外键 → users.id） |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_by | BIGINT | NOT NULL | 最后更新人（逻辑外键 → users.id） |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

### 68. `document_versions` — 文档版本

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| document_id | BIGINT | NOT NULL | 项目文档 ID（逻辑外键 → documents.id） |
| version_num | INT | NOT NULL | 版本号 |
| content_json | JSONB | | JSON 结构内容 |
| content_html | TEXT | | 富文本 HTML 内容 |
| change_summary | VARCHAR(500) | | 版本变更摘要 |
| created_by | BIGINT | NOT NULL | 创建人（逻辑外键 → users.id） |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |

> UNIQUE(tenant_id, document_id, version_num)

### 69. `document_links` — 多态关联

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| document_id | BIGINT | NOT NULL | 项目文档 ID（逻辑外键 → documents.id） |
| linkable_type | VARCHAR(20) | NOT NULL | requirement/task/defect/sprint/version |
| linkable_id | BIGINT | NOT NULL | 关联对象 ID |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |

> UNIQUE(tenant_id, document_id, linkable_type, linkable_id)

### 70. `document_comments` — 文档评论
### 71. `document_attachments` — 文档附件

结构与 `requirement_comments` / `requirement_attachments` 一致（`requirement_id` → `document_id`，逻辑外键 → `documents.id`），省略重复字段定义。

---

## 16、通知域

### 72. `notifications` — 通知（分区表）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| ~~基础字段~~ | — | — | 见模板 |
| id | BIGINT | PK | 雪花ID |
| code | VARCHAR(50) | | 业务短代码标识（用户按规则生成） |
| name | VARCHAR(500) | NOT NULL | 通知摘要（代替 title） |
| recipient_id | BIGINT | NOT NULL | 接收人 ID（逻辑外键 → users.id） |
| actor_id | BIGINT | | 操作人 ID |
| requirement_id | BIGINT | | 关联需求（逻辑外键 → requirements.id；三选一） |
| task_id | BIGINT | | 关联任务（逻辑外键 → tasks.id；三选一） |
| defect_id | BIGINT | | 关联缺陷（逻辑外键 → defects.id；三选一） |
| event_type | VARCHAR(50) | NOT NULL | 事件类型 |
| message | TEXT | NOT NULL | 通知正文 |
| message_template | VARCHAR(255) | | 模板标识 |
| extra_data | JSONB | | 扩展数据（跳转链接/上下文 JSON） |
| is_read | BOOLEAN | DEFAULT false | 是否已读 |
| read_at | TIMESTAMPTZ | | 已读时间 |
| is_archived | BOOLEAN | DEFAULT false | 已归档（PRD 9.3：已读通知自动归档） |
| priority | VARCHAR(20) | DEFAULT 'normal' | 通知优先级（normal/high） |
| status | VARCHAR(50) | DEFAULT 'active' | 状态 |
| deleted | BOOLEAN | DEFAULT false | 软删除标记（true=已删除） |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |
| created_by | BIGINT | NOT NULL | 创建人（逻辑外键 → users.id） |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_by | BIGINT | NOT NULL | 最后更新人（逻辑外键 → users.id） |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

-- 业务约束：三选一（至少一个非空）
-- CHECK (num_nonnulls(requirement_id, task_id, defect_id) >= 1)
-- 或通过触发器保证唯一引用

> 分区策略：按月分区，已读90天后自动清理。

### 73. `notification_subscriptions` — 通知订阅

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| user_id | BIGINT | NOT NULL | 用户 ID（逻辑外键 → users.id） |
| project_id | BIGINT | | 项目 ID（逻辑外键 → projects.id） |
| requirement_id | BIGINT | | 关联需求（逻辑外键 → requirements.id；三选一） |
| task_id | BIGINT | | 关联任务（逻辑外键 → tasks.id；三选一） |
| defect_id | BIGINT | | 关联缺陷（逻辑外键 → defects.id；三选一） |
| event_types | JSONB | NOT NULL | 订阅事件列表 |
| channels | JSONB | NOT NULL | [in_app/email/im/sms] |
| is_enabled | BOOLEAN | DEFAULT true | 订阅是否启用 |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

-- 业务约束：subscribe to project or specific entity (三选一)
-- CHECK (project_id IS NOT NULL OR num_nonnulls(requirement_id, task_id, defect_id) >= 1)

> UNIQUE(tenant_id, user_id, project_id, requirement_id, task_id, defect_id)

### 74. `notification_deliveries` — 通知投递记录

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| notification_id | BIGINT | NOT NULL | 通知 ID（逻辑外键 → notifications.id） |
| channel | VARCHAR(20) | NOT NULL | in_app/email/im/sms |
| status | VARCHAR(20) | DEFAULT 'pending' | pending/sent/failed |
| sent_at | TIMESTAMPTZ | | 发送时间 |
| error_code | VARCHAR(50) | | 发送错误码 |
| error_message | TEXT | | 错误信息 |
| retry_count | INT | DEFAULT 0 | 重试次数 |
| external_id | VARCHAR(200) | | 外部系统ID |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |

### 75. `notification_digests` — 通知汇总配置

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| user_id | BIGINT | NOT NULL | 用户 ID（逻辑外键 → users.id） |
| frequency | VARCHAR(20) | DEFAULT 'realtime' CHECK (realtime/daily/weekly/monthly) | 通知汇总频率（实时/每日/每周/每月） |
| day_of_week | SMALLINT | 1-7 | weekly 时用 |
| time_of_day | time | | daily 发送时间 |
| timezone | VARCHAR(50) | DEFAULT 'Asia/Shanghai' | 时区（决定汇总时间点） |
| quiet_hours_start | TIME | | 免打扰开始（如 22:00，PRD 9.3 免打扰时段） |
| quiet_hours_end | TIME | | 免打扰结束（如 08:00；跨天由应用层处理） |
| muted_event_types | JSONB | | 静音事件类型列表 |
| last_sent_at | TIMESTAMPTZ | | 上次发送时间 |
| is_enabled | BOOLEAN | DEFAULT true | 是否启用汇总 |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

> UNIQUE(tenant_id, user_id)

### 76.1. `notification_preferences` — 用户通知偏好

> 对应 PRD 9.3 通知系统：按「项目×事件类型×渠道」三维开关；支持免打扰（DND）时段与摘要投递模式。与 notification_subscriptions 互补——subscriptions 管实体级订阅，preferences 管用户全局投递策略。

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| user_id | BIGINT | NOT NULL | 用户 ID（逻辑外键 → users.id） |
| project_id | BIGINT | NOT NULL | 项目维度隔离（0 = 全局默认） |
| event_types | JSONB | NOT NULL DEFAULT '[]' | 订阅事件类型列表 |
| channels | JSONB | NOT NULL DEFAULT '["in_app"]' | 启用渠道 [in_app/email/im/sms] |
| digest | VARCHAR(16) | NOT NULL DEFAULT 'realtime' CHECK (realtime/daily/weekly) | 投递模式：realtime(实时) / daily(每日08:30聚合) / weekly(每周一聚合) |
| dnd_enabled | BOOLEAN | NOT NULL DEFAULT false | 免打扰开关 |
| dnd_start | TIME | DEFAULT '22:00:00' | 免打扰开始时间（HH:MM 用户时区） |
| dnd_end | TIME | DEFAULT '08:00:00' | 免打扰结束时间；高优事件(mention/automation.failed)可豁免 |
| is_enabled | BOOLEAN | NOT NULL DEFAULT true | 全局通知开关 |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

> UNIQUE(tenant_id, user_id, project_id)

---

## 17、效率增强域

### 77. `dashboards` — 仪表盘

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| ~~基础字段~~ | — | — | 见模板 |
| id | BIGINT | PK | 雪花ID |
| code | VARCHAR(50) | | 业务短代码标识（用户按规则生成） |
| name | VARCHAR(255) | NOT NULL | 名称 |
| description | TEXT | | 描述 |
| layout_config | JSONB | | 布局 |
| scope | VARCHAR(20) | DEFAULT 'project' CHECK (project/tenant/personal) | 仪表盘作用域（项目/租户/个人） |
| project_id | BIGINT | | 项目 ID（逻辑外键 → projects.id） |
| owner_id | BIGINT | NOT NULL | 归属人 |
| is_default | BOOLEAN | DEFAULT false | 是否默认仪表盘 |
| is_shared | BOOLEAN | DEFAULT false | 是否共享 |
| status | VARCHAR(50) | DEFAULT 'active' | 状态 |
| deleted | BOOLEAN | DEFAULT false | 软删除标记（true=已删除） |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |
| created_by | BIGINT | NOT NULL | 创建人（逻辑外键 → users.id） |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_by | BIGINT | NOT NULL | 最后更新人（逻辑外键 → users.id） |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

### 78. `dashboard_widgets` — 仪表盘小部件

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| ~~基础字段~~ | — | — | 见模板 |
| id | BIGINT | PK | 雪花ID |
| code | VARCHAR(50) | | 业务短代码标识（用户按规则生成） |
| name | VARCHAR(255) | NOT NULL | 名称 |
| dashboard_id | BIGINT | NOT NULL | 仪表盘 ID（逻辑外键 → dashboards.id） |
| widget_type | VARCHAR(50) | NOT NULL CHECK (chart/counter/list/table/calendar/gantt) | 小部件类型（图表/计数器/列表/表格等） |
| data_source_type | VARCHAR(50) | | 数据源类型 |
| config | JSONB | NOT NULL | 配置 |
| query_config | JSONB | | 查询配置 |
| position_x | INT | DEFAULT 0 | X 坐标（栅格） |
| position_y | INT | DEFAULT 0 | Y 坐标（栅格） |
| width | INT | DEFAULT 6 | 宽度（栅格列数） |
| height | INT | DEFAULT 4 | 高度（栅格行数） |
| refresh_interval | INT | DEFAULT 300 | 秒 |
| status | VARCHAR(50) | DEFAULT 'active' | 状态 |
| deleted | BOOLEAN | DEFAULT false | 软删除标记（true=已删除） |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |
| created_by | BIGINT | NOT NULL | 创建人（逻辑外键 → users.id） |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_by | BIGINT | NOT NULL | 最后更新人（逻辑外键 → users.id） |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

### 79. `dashboard_templates` — 仪表盘模板

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| ~~基础字段~~ | — | — | 见模板 |
| id | BIGINT | PK | 雪花ID |
| code | VARCHAR(50) | UNIQUE | 业务短代码标识（用户按规则生成） |
| name | VARCHAR(255) | NOT NULL | 名称 |
| description | TEXT | | 描述 |
| thumbnail_url | TEXT | | 缩略图 URL |
| layout_config | JSONB | | 布局配置（JSON） |
| widgets_template | JSONB | | 小部件模板配置（JSON） |
| scope | VARCHAR(20) | DEFAULT 'system' | system/tenant |
| applicable_project_types | JSONB | | 适用项目类型 |
| status | VARCHAR(50) | DEFAULT 'active' | 状态 |
| deleted | BOOLEAN | DEFAULT false | 软删除标记（true=已删除） |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |
| created_by | BIGINT | NOT NULL | 创建人（逻辑外键 → users.id） |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_by | BIGINT | NOT NULL | 最后更新人（逻辑外键 → users.id） |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

### 80. `dashboard_snapshots` — 仪表盘快照

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| dashboard_id | BIGINT | NOT NULL | 仪表盘 ID（逻辑外键 → dashboards.id） |
| snapshot_name | VARCHAR(255) | | 快照名称 |
| snapshot_data | JSONB | NOT NULL | 快照数据（JSON） |
| created_by | BIGINT | NOT NULL | 创建人（逻辑外键 → users.id） |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |

### 81. `workbench_configs` — 个人工作台

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| user_id | BIGINT | NOT NULL | 用户 ID（逻辑外键 → users.id） |
| layout_config | JSONB | NOT NULL | 工作台布局配置（JSON） |
| widgets_config | JSONB | NOT NULL | 小部件配置（JSON） |
| quick_filters | JSONB | | 快捷筛选配置（JSON） |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

> UNIQUE(tenant_id, user_id)

### 82. `workbench_templates` — 工作台模板

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| ~~基础字段~~ | — | — | 见模板 |
| id | BIGINT | PK | 雪花ID |
| code | VARCHAR(50) | UNIQUE | 业务短代码标识（用户按规则生成） |
| name | VARCHAR(255) | NOT NULL | 名称 |
| description | TEXT | | 描述 |
| layout_config | JSONB | | 布局配置（JSON） |
| widgets_config | JSONB | | 小部件配置（JSON） |
| applicable_roles | JSONB | | 适用角色列表 |
| status | VARCHAR(50) | DEFAULT 'active' | 状态 |
| deleted | BOOLEAN | DEFAULT false | 软删除标记（true=已删除） |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |
| created_by | BIGINT | NOT NULL | 创建人（逻辑外键 → users.id） |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_by | BIGINT | NOT NULL | 最后更新人（逻辑外键 → users.id） |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

### 83. `recent_items` — 最近访问

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| user_id | BIGINT | NOT NULL | 用户 ID（逻辑外键 → users.id） |
| entity_type | VARCHAR(20) | NOT NULL | 实体类型（requirement/task/defect） |
| entity_id | BIGINT | NOT NULL | 实体 ID |
| project_id | BIGINT | | 项目 ID（逻辑外键 → projects.id） |
| view_count | INT | DEFAULT 1 | 访问次数 |
| last_accessed_at | TIMESTAMPTZ | DEFAULT now() | 最后访问时间 |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |

> UNIQUE(tenant_id, user_id, entity_type, entity_id)

### 84. `view_preferences` — 视图偏好

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| user_id | BIGINT | NOT NULL | 用户 ID（逻辑外键 → users.id） |
| project_id | BIGINT | NOT NULL | 项目 ID（逻辑外键 → projects.id） |
| view_type | VARCHAR(30) | NOT NULL CHECK (kanban/list/calendar/gantt/timeline/table) | 视图类型（看板/列表/日历/甘特等） |
| entity_scope | VARCHAR(20) | DEFAULT 'all' CHECK (all/requirement/task/defect) | 视图范围 |
| filters | JSONB | | 过滤条件 |
| sort_config | JSONB | | 排序 |
| columns_config | JSONB | | 列配置 |
| group_by | VARCHAR(50) | | 分组字段 |
| viewport_state | JSONB | | 状态 |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

> UNIQUE(tenant_id, user_id, project_id, view_type, entity_scope)

### 85. `search_history` — 搜索历史

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| user_id | BIGINT | NOT NULL | 用户 ID（逻辑外键 → users.id） |
| query_text | VARCHAR(500) | NOT NULL | 搜索关键词 |
| filters_snapshot | JSONB | | 过滤条件快照（JSON） |
| entity_types | JSONB | | 搜索范围 |
| result_count | INT | | 结果数量 |
| response_time_ms | INT | | 搜索耗时（毫秒） |
| last_used_at | TIMESTAMPTZ | DEFAULT now() | 最后使用时间 |
| use_count | INT | DEFAULT 1 | 使用次数 |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |

### 86. `search_bookmarks` — 搜索书签

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| ~~基础字段~~ | — | — | 见模板 |
| id | BIGINT | PK | 雪花ID |
| code | VARCHAR(50) | | 业务短代码标识（用户按规则生成） |
| name | VARCHAR(255) | NOT NULL | 名称 |
| query_text | VARCHAR(500) | NOT NULL | 收藏的搜索词 |
| filters_snapshot | JSONB | | 过滤条件快照（JSON） |
| entity_types | JSONB | | 限定实体类型（JSON） |
| status | VARCHAR(50) | DEFAULT 'active' | 状态 |
| deleted | BOOLEAN | DEFAULT false | 软删除标记（true=已删除） |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |
| created_by | BIGINT | NOT NULL | 创建人（逻辑外键 → users.id） |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_by | BIGINT | NOT NULL | 最后更新人（逻辑外键 → users.id） |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

### 87.1. `search_documents` — 搜索文档索引（ES 降级兜底）

> 对应架构10 全文检索：ES 正常时本表不参与查询；ES 不可用时降级为 PG `to_tsvector` 全文检索源。由 Outbox 消费者异步写入，保持与工作项/文档/知识库内容同步。

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| project_id | BIGINT | NOT NULL | 项目 ID（逻辑外键 → projects.id） |
| doc_type | VARCHAR(30) | NOT NULL | requirement/task/defect/document/knowledge_page |
| doc_id | BIGINT | NOT NULL | 被索引实体ID |
| title | TEXT | NOT NULL | 索引标题 |
| identifier | VARCHAR(50) | | 工作项编码（如 YD-1001） |
| content | TEXT | | 纯文本内容（tsvector 计算源） |
| search_tsv | TSVECTOR | | PG 全文检索向量 |
| metadata | JSONB | NOT NULL DEFAULT '{}' | 附加属性（attribution/sprint/labels/status 等，用于过滤） |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

> UNIQUE(tenant_id, doc_type, doc_id)
> INDEX GIN (search_tsv) — 全文检索降级索引
> INDEX (tenant_id, project_id, doc_type) — 按项目+类型过滤

---

## 18、自动化域

### 88. `automation_rules` — 自动化规则

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| ~~基础字段~~ | — | — | 见模板 |
| id | BIGINT | PK | 雪花ID |
| code | VARCHAR(50) | | 业务短代码标识（用户按规则生成） |
| name | VARCHAR(255) | NOT NULL | 名称 |
| description | TEXT | | 描述 |
| trigger_event | VARCHAR(100) | NOT NULL | 触发事件 |
| trigger_config | JSONB | | 触发配置 |
| conditions | JSONB | NOT NULL | 执行条件 |
| actions | JSONB | NOT NULL | 执行动作 |
| entity_scope | VARCHAR(20) | DEFAULT 'all' CHECK (all/requirement/task/defect) | 适用类型 |
| priority | INT | DEFAULT 0 | 优先级（越小越先执行） |
| execution_count | INT | DEFAULT 0 | 累计执行次数 |
| success_count | INT | DEFAULT 0 | 成功数量 |
| fail_count | INT | DEFAULT 0 | 失败数量 |
| last_executed_at | TIMESTAMPTZ | | 最后执行时间 |
| last_error | TEXT | | 最近一次错误信息 |
| dry_run | BOOLEAN | DEFAULT false | 试运行模式 |
| status | VARCHAR(50) | DEFAULT 'active' CHECK (active/paused/disabled) | 状态 |
| deleted | BOOLEAN | DEFAULT false | 软删除标记（true=已删除） |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |
| created_by | BIGINT | NOT NULL | 创建人（逻辑外键 → users.id） |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_by | BIGINT | NOT NULL | 最后更新人（逻辑外键 → users.id） |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

### 89. `rule_executions` — 规则执行记录（分区表）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| rule_id | BIGINT | NOT NULL | 自动化规则 ID（逻辑外键 → automation_rules.id） |
| trigger_event_type | VARCHAR(100) | NOT NULL | 触发事件类型 |
| trigger_entity_type | VARCHAR(20) | | 触发实体类型 |
| trigger_entity_id | BIGINT | | 触发实体 ID |
| input_data | JSONB | | 输入数据（JSON） |
| results | JSONB | | 执行结果（JSON） |
| status | VARCHAR(20) | DEFAULT 'pending' CHECK (pending/running/success/failed/cancelled) | 状态 |
| error_message | TEXT | | 错误信息 |
| duration_ms | INT | | 执行耗时（毫秒） |
| executed_at | TIMESTAMPTZ | DEFAULT now() | 执行时间 |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |

> 分区策略：按月分区，30天TTL。

### 90. `automation_templates` — 自动化模板

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| ~~基础字段~~ | — | — | 见模板 |
| id | BIGINT | PK | 雪花ID |
| code | VARCHAR(50) | UNIQUE | 业务短代码标识（用户按规则生成） |
| name | VARCHAR(255) | NOT NULL | 名称 |
| description | TEXT | | 描述 |
| category | VARCHAR(50) | | 模板分类（issue/sprint/version/member） |
| icon | VARCHAR(100) | | 模板图标 |
| trigger_event | VARCHAR(100) | NOT NULL | 触发事件 |
| trigger_config_template | JSONB | | 触发器配置模板（JSON） |
| conditions_template | JSONB | | 条件配置模板（JSON） |
| actions_template | JSONB | | 动作配置模板（JSON） |
| applicable_entity_types | JSONB | | 适用实体类型列表 |
| applicable_project_types | JSONB | | 适用项目类型列表 |
| popularity | INT | DEFAULT 0 | 使用次数 |
| rating | NUMERIC(3,2) | | 评分 |
| is_builtin | BOOLEAN | DEFAULT false | 是否系统内置 |
| status | VARCHAR(50) | DEFAULT 'active' | 状态 |
| deleted | BOOLEAN | DEFAULT false | 软删除标记（true=已删除） |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |
| created_by | BIGINT | NOT NULL | 创建人（逻辑外键 → users.id） |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_by | BIGINT | NOT NULL | 最后更新人（逻辑外键 → users.id） |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

---

## 19、集成域

### 91. `webhooks` — Webhook 配置

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| ~~基础字段~~ | — | — | 见模板 |
| id | BIGINT | PK | 雪花ID |
| code | VARCHAR(50) | | 业务短代码标识（用户按规则生成） |
| name | VARCHAR(255) | NOT NULL | 名称 |
| description | TEXT | | 描述 |
| url | TEXT | NOT NULL | 回调 URL |
| secret | VARCHAR(128) | NOT NULL | 签名密钥（HMAC-SHA256） |
| project_id | BIGINT | 逻辑外键 → projects.id | 项目级 Webhook；NULL = 租户级 |
| events | JSONB | NOT NULL | 订阅事件（30+ 事件类型，PRD 9.6） |
| entity_scope | VARCHAR(20) | DEFAULT 'all' | 实体范围 |
| headers | JSONB | | 自定义请求头 |
| timeout_ms | INT | DEFAULT 10000 | 超时时间（毫秒） |
| retry_config | JSONB | | 重试配置 |
| is_active | BOOLEAN | DEFAULT true | 是否启用 |
| last_triggered_at | TIMESTAMPTZ | | 最后触发时间 |
| last_error | TEXT | | 最近一次错误信息 |
| status | VARCHAR(50) | DEFAULT 'active' | 状态 |
| deleted | BOOLEAN | DEFAULT false | 软删除标记（true=已删除） |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |
| created_by | BIGINT | NOT NULL | 创建人（逻辑外键 → users.id） |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_by | BIGINT | NOT NULL | 最后更新人（逻辑外键 → users.id） |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

### 92. `webhook_logs` — Webhook 投递日志（分区表）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| webhook_id | BIGINT | NOT NULL | Webhook ID（逻辑外键 → webhooks.id） |
| event_type | VARCHAR(100) | NOT NULL | 事件类型 |
| entity_type | VARCHAR(20) | | 实体类型（requirement/task/defect） |
| entity_id | BIGINT | | 实体 ID |
| payload | JSONB | NOT NULL | 事件载荷（JSON） |
| request_headers | JSONB | | 请求头（JSON） |
| response_status | INT | | 响应状态码 |
| response_body | TEXT | | 响应体 |
| response_time_ms | INT | | 响应耗时（毫秒） |
| retry_count | INT | DEFAULT 0 | 重试次数 |
| delivery_status | VARCHAR(20) | DEFAULT 'pending' CHECK (pending/sent/failed) | 投递状态（待发送/已发送/失败） |
| error_message | TEXT | | 错误信息 |
| delivered_at | TIMESTAMPTZ | | 投递时间 |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |

> 分区策略：按月分区，30天TTL。

### 93. `api_tokens` — API Token

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| ~~基础字段~~ | — | — | 见模板 |
| id | BIGINT | PK | 雪花ID |
| code | VARCHAR(50) | | 业务短代码标识（用户按规则生成） |
| name | VARCHAR(255) | NOT NULL | 名称 |
| token_hash | TEXT | NOT NULL, UNIQUE | Token哈希 |
| token_prefix | VARCHAR(10) | | Token前缀 |
| scopes | JSONB | NOT NULL | 权限范围 |
| ip_whitelist | JSONB | | IP白列表 |
| rate_limit | INT | | 速率限制 |
| entity_type | VARCHAR(20) | DEFAULT 'user' CHECK (user/service) | 实体类型（requirement/task/defect） |
| expires_at | TIMESTAMPTZ | | 过期时间（PRD 1.6 api_token.expires_at） |
| last_used_at | TIMESTAMPTZ | | 最后使用时间 |
| last_used_ip | INET | | 最后使用 IP |
| status | VARCHAR(50) | DEFAULT 'active' | 状态 |
| deleted | BOOLEAN | DEFAULT false | 软删除标记（true=已删除） |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |
| created_by | BIGINT | NOT NULL | 创建人（逻辑外键 → users.id） |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_by | BIGINT | NOT NULL | 最后更新人（逻辑外键 → users.id） |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

### 94. `sso_providers` — SSO 配置

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| ~~基础字段~~ | — | — | 见模板 |
| id | BIGINT | PK | 雪花ID |
| code | VARCHAR(50) | | 业务短代码标识（用户按规则生成） |
| name | VARCHAR(100) | NOT NULL | 名称 |
| provider_type | VARCHAR(20) | NOT NULL CHECK (oidc/saml/ldap) | 提供商类型（OIDC/SAML/LDAP） |
| description | TEXT | | 描述 |
| config | JSONB | NOT NULL | 配置（含敏感字段加密） |
| default_role_id | BIGINT | | 默认角色 |
| user_mapping_rules | JSONB | | 用户属性映射 |
| auto_create_user | BOOLEAN | DEFAULT true | 是否自动创建用户 |
| status | VARCHAR(50) | DEFAULT 'active' | 状态 |
| deleted | BOOLEAN | DEFAULT false | 软删除标记（true=已删除） |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |
| created_by | BIGINT | NOT NULL | 创建人（逻辑外键 → users.id） |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_by | BIGINT | NOT NULL | 最后更新人（逻辑外键 → users.id） |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

### 95. `sso_sessions` — SSO 会话

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| user_id | BIGINT | NOT NULL | 用户 ID（逻辑外键 → users.id） |
| provider_id | BIGINT | NOT NULL | SSO 提供商 ID（逻辑外键 → sso_providers.id） |
| external_user_id | VARCHAR(255) | | 外部系统用户ID |
| session_data | JSONB | | SSO 会话数据（JSON） |
| ip_address | INET | | IP 地址 |
| user_agent | TEXT | | 浏览器/客户端 User-Agent |
| login_at | TIMESTAMPTZ | DEFAULT now() | 登录时间 |
| logout_at | TIMESTAMPTZ | | 登出时间 |
| expires_at | TIMESTAMPTZ | | 过期时间 |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |

> UNIQUE(tenant_id, provider_id, external_user_id)

### 96.1. `sso_links` — SSO 账号关联

> 对应架构04 SSO 设计：OIDC/SAML 登录成功后，将外部 IdP 账号与系统用户绑定（upsert）。同一 IdP+subject 唯一对应一个系统用户，支持一人绑定多个 IdP。

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| user_id | BIGINT | NOT NULL, 逻辑外键 → users.id | 关联的系统用户 |
| provider_id | BIGINT | NOT NULL, 逻辑外键 → sso_providers.id | SSO 提供者 |
| sso_subject | VARCHAR(255) | NOT NULL | 外部 IdP 的唯一标识（OIDC sub / SAML NameID） |
| sso_email | VARCHAR(255) | | 外部 IdP 返回的邮箱 |
| sso_display_name | VARCHAR(255) | | 外部 IdP 返回的显示名 |
| last_login_at | TIMESTAMPTZ | | 最后通过 SSO 登录时间 |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |

> UNIQUE(tenant_id, provider_id, sso_subject) — 同一 IdP 同一 subject 仅绑定一个用户
> INDEX (tenant_id, user_id) — 查询用户绑定的所有 IdP

---

## 20、事件总线域

### 97. `domain_events` — 领域事件（Outbox）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| aggregate_type | VARCHAR(50) | NOT NULL | 聚合类型 |
| aggregate_id | BIGINT | NOT NULL | 聚合 ID |
| event_type | VARCHAR(100) | NOT NULL | 事件类型 |
| event_version | INT | DEFAULT 1 | 事件版本 |
| payload | JSONB | NOT NULL | 事件载荷（JSON） |
| metadata | JSONB | | 元数据（JSON） |
| trace_id | VARCHAR(64) | | 链路追踪 |
| occurred_at | TIMESTAMPTZ | DEFAULT now() | 事件发生时间 |
| published_at | TIMESTAMPTZ | | 发布时间 |
| published_to | JSONB | [] | 已发布的消费者 |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |

### 98. `processed_events` — 消费者幂等记录

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| consumer | VARCHAR(100) | NOT NULL | 消费者标识 |
| event_id | BIGINT | NOT NULL | 事件 ID（逻辑外键 → domain_events.id） |
| processed_at | TIMESTAMPTZ | DEFAULT now() | 处理时间 |
| result_status | VARCHAR(20) | DEFAULT 'success' | 处理结果状态 |
| error_message | TEXT | | 错误信息 |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |

> PRIMARY KEY (tenant_id, consumer, event_id)

### 99. `dlq_events` — 死信队列

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| event_id | BIGINT | NOT NULL | 原始事件 ID（逻辑外键 → domain_events.id） |
| consumer | VARCHAR(100) | NOT NULL | 消费者标识 |
| error_type | VARCHAR(50) | | 错误类型 |
| error_message | TEXT | | 错误信息 |
| retry_count | INT | DEFAULT 0 | 重试次数 |
| original_payload | JSONB | | 原始事件载荷（JSON） |
| status | VARCHAR(20) | DEFAULT 'pending' CHECK (pending/processing/resolved/discarded) | 状态 |
| resolved_at | TIMESTAMPTZ | | 解决时间 |
| resolved_by | BIGINT | | 解决人 ID |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

### 100. `idempotency_keys` — API幂等键

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| key | VARCHAR(64) | PK | 幂等键Hash |
| user_id | BIGINT | NOT NULL | 用户 ID（逻辑外键 → users.id） |
| method | VARCHAR(10) | NOT NULL | 请求方法 |
| path | VARCHAR(500) | NOT NULL | 请求路径 |
| request_hash | VARCHAR(64) | | 请求体Hash |
| response_body | JSONB | | 缓存响应 |
| response_status | INT | | 缓存状态码 |
| expires_at | TIMESTAMPTZ | NOT NULL | 过期时间 |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |

---

## 21、安全审计域

### 101. `password_reset_tokens` — 密码重置Token

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| user_id | BIGINT | NOT NULL | 用户 ID（逻辑外键 → users.id） |
| token_hash | TEXT | NOT NULL, UNIQUE | Token 哈希值（不存明文） |
| expires_at | TIMESTAMPTZ | NOT NULL | 过期时间 |
| used_at | TIMESTAMPTZ | | 使用时间 |
| ip_address | INET | | IP 地址 |
| user_agent | TEXT | | 浏览器/客户端 User-Agent |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |

### 102. `audit_logs` — 审计日志（分区表）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| user_id | BIGINT | | 操作用户 |
| username | VARCHAR(255) | | 冗余（用户删除后保留） |
| action | VARCHAR(100) | NOT NULL | 操作 |
| resource_type | VARCHAR(50) | | 资源类型 |
| resource_id | BIGINT | | 资源ID |
| resource_name | VARCHAR(255) | | 资源名称冗余 |
| entity_type | VARCHAR(20) | | 实体类型 |
| entity_id | BIGINT | | 实体 ID |
| method | VARCHAR(10) | | HTTP方法 |
| path | VARCHAR(500) | | 请求路径 |
| ip_address | INET | | IP 地址 |
| user_agent | TEXT | | 浏览器/客户端 User-Agent |
| request_body | JSONB | | 请求体 |
| response_status | INT | | 响应状态码 |
| changes | JSONB | | 变更详情 |
| risk_level | VARCHAR(20) | DEFAULT 'low' | 风险等级（low/medium/high） |
| session_id | VARCHAR(64) | | 会话 ID |
| trace_id | VARCHAR(64) | | 链路追踪 ID |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |
| operated_at | TIMESTAMPTZ | DEFAULT now() | 操作时间 |

> 分区策略：按月分区，12个月后归档。

### 103. `login_attempts` — 登录尝试记录

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| user_id | BIGINT | | 用户 ID（逻辑外键 → users.id） |
| username_attempted | VARCHAR(255) | | 尝试登录的用户名 |
| success | BOOLEAN | NOT NULL | 是否成功 |
| failure_reason | VARCHAR(100) | | 失败原因 |
| ip_address | INET | | IP 地址 |
| user_agent | TEXT | | 浏览器/客户端 User-Agent |
| geo_location | VARCHAR(100) | | 地理位置 |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |

> 按月分区，30天TTL。

### 104. `invitations` — 邀请

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| ~~基础字段~~ | — | — | 见模板 |
| id | BIGINT | PK | 雪花ID |
| code | VARCHAR(50) | | 业务短代码标识（用户按规则生成） |
| name | VARCHAR(255) | NOT NULL | 备注名 |
| invitee_email | VARCHAR(255) | NOT NULL | 被邀请人邮箱 |
| invitee_name | VARCHAR(255) | | 被邀请人姓名 |
| project_id | BIGINT | | 项目 ID（逻辑外键 → projects.id） |
| role | VARCHAR(50) | NOT NULL | 邀请角色 |
| invite_type | VARCHAR(20) | DEFAULT 'project' CHECK (tenant/project) | 邀请类型（租户级/项目级） |
| token_hash | VARCHAR(128) | NOT NULL, UNIQUE | Token 哈希值（不存明文） |
| expire_at | TIMESTAMPTZ | NOT NULL | 邀请链接过期时间（默认 7 天） |
| accepted_at | TIMESTAMPTZ | | 接受邀请时间 |
| accepted_user_id | BIGINT | | 接受邀请的用户 ID |
| rejected_at | TIMESTAMPTZ | | 拒绝邀请时间 |
| status | VARCHAR(20) | DEFAULT 'pending' CHECK (pending/accepted/rejected/expired) | 状态 |
| welcome_message | TEXT | | 欢迎语（邀请邮件内容） |
| deleted | BOOLEAN | DEFAULT false | 软删除标记（true=已删除） |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |
| created_by | BIGINT | NOT NULL | 创建人（逻辑外键 → users.id） |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_by | BIGINT | NOT NULL | 最后更新人（逻辑外键 → users.id） |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

---

## 22、效能度量域

### 105. `metric_snapshots` — 效能指标快照

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| metric_type | VARCHAR(50) | NOT NULL | 指标类型 |
| metric_name | VARCHAR(100) | NOT NULL | 指标名称 |
| project_id | BIGINT | | 项目 ID（逻辑外键 → projects.id） |
| sprint_id | BIGINT | | 迭代 ID（逻辑外键 → sprints.id） |
| version_id | BIGINT | | 版本日 ID（逻辑外键 → versions.id） |
| entity_scope | VARCHAR(20) | DEFAULT 'all' | 指标作用域（all/requirement/task/defect） |
| aggregation_dim | VARCHAR(50) | | 维度（type/module/state） |
| dim_value | VARCHAR(100) | | 维度值 |
| value | NUMERIC(16,4) | | 指标值 |
| count_val | BIGINT | | 计数值 |
| ratio_val | NUMERIC(8,4) | | 比率值 |
| details | JSONB | | 详细数据 |
| snapshot_date | DATE | NOT NULL | 快照日期 |
| snapshot_hour | SMALLINT | | 小时粒度 |
| period_type | VARCHAR(20) | DEFAULT 'daily' CHECK (hourly/daily/weekly/monthly) | 统计周期类型（小时/日/周/月） |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |

> UNIQUE(tenant_id, metric_type, project_id, sprint_id, aggregation_dim, dim_value, snapshot_date, period_type)

### 106. `metric_adjustments` — 指标调整

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| ~~基础字段~~ | — | — | 见模板 |
| id | BIGINT | PK | 雪花ID |
| code | VARCHAR(50) | | 业务短代码标识（用户按规则生成） |
| name | VARCHAR(255) | NOT NULL | 名称 |
| metric_type | VARCHAR(50) | NOT NULL | 指标类型（效率/质量/交付） |
| project_id | BIGINT | | 项目 ID（逻辑外键 → projects.id） |
| adjustment_date | DATE | NOT NULL | 调整生效日期 |
| delta | NUMERIC(16,4) | NOT NULL | 修正量（可为负） |
| reason | VARCHAR(500) | | 原因 |
| status | VARCHAR(50) | DEFAULT 'approved' | 状态 |
| deleted | BOOLEAN | DEFAULT false | 软删除标记（true=已删除） |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |
| created_by | BIGINT | NOT NULL | 创建人（逻辑外键 → users.id） |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_by | BIGINT | NOT NULL | 最后更新人（逻辑外键 → users.id） |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

### 107. `metric_definitions` — 指标定义

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| ~~基础字段~~ | — | — | 见模板 |
| id | BIGINT | PK | 雪花ID |
| code | VARCHAR(50) | UNIQUE | 业务短代码标识（用户按规则生成） |
| name | VARCHAR(100) | NOT NULL | 名称 |
| description | TEXT | | 描述 |
| metric_category | VARCHAR(50) | NOT NULL | velocity/burndown/lead_time/... |
| calculation_method | VARCHAR(50) | | 计算方式 |
| calculation_formula | TEXT | | 计算公式 |
| unit | VARCHAR(20) | | 单位 |
| default_config | JSONB | | 默认配置（JSON） |
| dimensions | JSONB | | 可用维度 |
| visualization_config | JSONB | | 可视化配置 |
| is_builtin | BOOLEAN | DEFAULT false | 是否系统内置 |
| status | VARCHAR(50) | DEFAULT 'active' | 状态 |
| deleted | BOOLEAN | DEFAULT false | 软删除标记（true=已删除） |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |
| created_by | BIGINT | NOT NULL | 创建人（逻辑外键 → users.id） |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_by | BIGINT | NOT NULL | 最后更新人（逻辑外键 → users.id） |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

---

## 23、治理域

### 108. `risk_rules` — 风险规则

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| ~~基础字段~~ | — | — | 见模板 |
| id | BIGINT | PK | 雪花ID |
| code | VARCHAR(50) | | 业务短代码标识（用户按规则生成） |
| name | VARCHAR(255) | NOT NULL | 名称 |
| description | TEXT | | 描述 |
| risk_type | VARCHAR(50) | NOT NULL CHECK (delay/quality/resource/scope/other) | 风险类型（延期/质量/资源/范围/其他） |
| trigger_conditions | JSONB | NOT NULL | 触发条件 |
| threshold_config | JSONB | | 阈值 |
| severity | VARCHAR(20) | DEFAULT 'medium' CHECK (low/medium/high/critical) | 严重程度（低/中/高/严重） |
| auto_actions | JSONB | | 自动动作 |
| applicable_entity_types | JSONB | | 适用实体（requirement/task/defect） |
| evaluation_frequency | VARCHAR(20) | DEFAULT 'daily' | 评估频率（每日等） |
| status | VARCHAR(50) | DEFAULT 'active' | 状态 |
| deleted | BOOLEAN | DEFAULT false | 软删除标记（true=已删除） |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |
| created_by | BIGINT | NOT NULL | 创建人（逻辑外键 → users.id） |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_by | BIGINT | NOT NULL | 最后更新人（逻辑外键 → users.id） |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

### 109. `risk_alerts` — 风险告警

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| ~~基础字段~~ | — | — | 见模板 |
| id | BIGINT | PK | 雪花ID |
| code | VARCHAR(50) | | 业务短代码标识（用户按规则生成） |
| name | VARCHAR(500) | NOT NULL | 名称 |
| rule_id | BIGINT | NOT NULL | 自动化规则 ID（逻辑外键 → automation_rules.id） |
| entity_type | VARCHAR(20) | | 实体类型（requirement/task/defect） |
| entity_id | BIGINT | | 实体 ID |
| project_id | BIGINT | | 项目 ID（逻辑外键 → projects.id） |
| sprint_id | BIGINT | | 迭代 ID（逻辑外键 → sprints.id） |
| alert_level | VARCHAR(20) | NOT NULL CHECK (info/warning/critical) | 告警级别（信息/警告/严重） |
| alert_content | TEXT | | 告警内容描述 |
| metric_snapshot | JSONB | | 触发时的指标快照（JSON） |
| suggested_actions | JSONB | | 建议处理动作（JSON 列表） |
| is_acknowledged | BOOLEAN | DEFAULT false | 是否已确认 |
| acknowledged_by | BIGINT | | 确认人 ID |
| acknowledged_at | TIMESTAMPTZ | | 确认时间 |
| resolution_note | TEXT | | 处理说明 |
| status | VARCHAR(50) | DEFAULT 'open' | 状态 |
| deleted | BOOLEAN | DEFAULT false | 软删除标记（true=已删除） |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |
| created_by | BIGINT | NOT NULL | 创建人（逻辑外键 → users.id） |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_by | BIGINT | NOT NULL | 最后更新人（逻辑外键 → users.id） |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

### 110. `deployment_events` — 部署事件

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| ~~基础字段~~ | — | — | 见模板 |
| id | BIGINT | PK | 雪花ID |
| code | VARCHAR(50) | | 业务短代码标识（用户按规则生成） |
| name | VARCHAR(255) | NOT NULL | 名称 |
| description | TEXT | | 描述 |
| project_id | BIGINT | | 项目 ID（逻辑外键 → projects.id） |
| version_id | BIGINT | | 版本日 ID（逻辑外键 → versions.id） |
| environment | VARCHAR(50) | NOT NULL | dev/staging/production |
| deployment_type | VARCHAR(30) | DEFAULT 'standard' | standard/rollback/hotfix |
| deployed_version | VARCHAR(100) | | 部署版本 |
| previous_version | VARCHAR(100) | | 上一版本 |
| changes_summary | TEXT | | 变更摘要 |
| artifacts | JSONB | | 构建产物 |
| status | VARCHAR(50) | DEFAULT 'pending' CHECK (pending/in_progress/success/failed/rolled_back) | 状态 |
| started_at | TIMESTAMPTZ | | 开始时间 |
| completed_at | TIMESTAMPTZ | | 完成时间 |
| duration_ms | INT | | 执行耗时（毫秒） |
| trigger_source | VARCHAR(30) | | manual/webhook/automation |
| metadata | JSONB | | 部署元数据（JSON） |
| deleted | BOOLEAN | DEFAULT false | 软删除标记（true=已删除） |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |
| created_by | BIGINT | NOT NULL | 创建人（逻辑外键 → users.id） |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_by | BIGINT | NOT NULL | 最后更新人（逻辑外键 → users.id） |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

---

## 24、内容管理域

### 111. `pages` — 静态页面（CMS）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| ~~基础字段~~ | — | — | 见模板 |
| id | BIGINT | PK | 雪花ID |
| code | VARCHAR(50) | UNIQUE | 页面标识 |
| name | VARCHAR(255) | NOT NULL | 名称 |
| title | VARCHAR(255) | NOT NULL | SEO标题 |
| slug | VARCHAR(255) | NOT NULL, UNIQUE | URL标识 |
| content_html | TEXT | | 富文本 HTML 内容 |
| content_json | JSONB | | JSON 结构内容 |
| description | VARCHAR(500) | | SEO描述 |
| keywords | VARCHAR(500) | | 页面关键词（SEO） |
| template | VARCHAR(50) | DEFAULT 'default' | 模板 |
| layout_type | VARCHAR(30) | DEFAULT 'full' | 页面布局类型 |
| cover_image_url | TEXT | | 封面图 URL |
| view_count | BIGINT | DEFAULT 0 | 浏览量 |
| is_published | BOOLEAN | DEFAULT false | 是否已发布 |
| published_at | TIMESTAMPTZ | | 发布时间 |
| expire_at | TIMESTAMPTZ | | 页面过期时间（营销页限时） |
| status | VARCHAR(50) | DEFAULT 'draft' CHECK (draft/published/archived) | 状态 |
| deleted | BOOLEAN | DEFAULT false | 软删除标记（true=已删除） |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |
| created_by | BIGINT | NOT NULL | 创建人（逻辑外键 → users.id） |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_by | BIGINT | NOT NULL | 最后更新人（逻辑外键 → users.id） |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

### 112. `page_templates` — 页面模板

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| ~~基础字段~~ | — | — | 见模板 |
| id | BIGINT | PK | 雪花ID |
| code | VARCHAR(50) | UNIQUE | 业务短代码标识（用户按规则生成） |
| name | VARCHAR(255) | NOT NULL | 名称 |
| description | TEXT | | 描述 |
| html_template | TEXT | | HTML 模板内容 |
| css_content | TEXT | | CSS 样式内容 |
| preview_image_url | TEXT | | 预览图 URL |
| content_slots | JSONB | | 内容插槽 |
| status | VARCHAR(50) | DEFAULT 'active' | 状态 |
| deleted | BOOLEAN | DEFAULT false | 软删除标记（true=已删除） |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |
| created_by | BIGINT | NOT NULL | 创建人（逻辑外键 → users.id） |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_by | BIGINT | NOT NULL | 最后更新人（逻辑外键 → users.id） |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

### 113. `page_shares` — 页面分享

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| page_id | BIGINT | NOT NULL | 知识文档 ID（逻辑外键 → knowledge_pages.id） |
| share_token | VARCHAR(128) | NOT NULL, UNIQUE | 分享令牌（URL 访问凭证） |
| share_type | VARCHAR(20) | DEFAULT 'public' CHECK (public/password/restricted) | 分享类型（公开/密码/指定成员） |
| password_hash | VARCHAR(128) | | 访问密码哈希 |
| expire_at | TIMESTAMPTZ | | 链接过期时间 |
| access_count | INT | DEFAULT 0 | 访问次数 |
| max_access_count | INT | | 最大访问次数限制 |
| is_active | BOOLEAN | DEFAULT true | 是否启用 |
| last_accessed_at | TIMESTAMPTZ | | 最后访问时间 |
| created_by | BIGINT | NOT NULL | 创建人（逻辑外键 → users.id） |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |

---

## 25、组织与项目扩展域（V1.1 新增）

> 对应 PRD 8.1 工作空间管理（成员四级角色、多语言/时区）与 8.2 项目管理（项目模板、项目集）。

### 114. `tenant_members` — 租户成员（空间成员身份）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |
| user_id | BIGINT | NOT NULL, 逻辑外键 → users.id | 用户 ID（逻辑外键 → users.id） |
| role | VARCHAR(20) | NOT NULL DEFAULT 'member' CHECK (owner/admin/member/guest) | 空间四级角色（PRD 1.4.2/1.4.3） |
| join_type | VARCHAR(20) | DEFAULT 'invitation' CHECK (direct/invitation/import) | 加入方式（import=CSV 批量导入） |
| invited_by | BIGINT | 逻辑外键 → users.id | 邀请人 |
| is_active | BOOLEAN | DEFAULT true | 移除成员置 false（已关联需求/任务/缺陷保留但无权限） |
| joined_at | TIMESTAMPTZ | DEFAULT now() | 加入时间 |
| last_active_at | TIMESTAMPTZ | | 最后活跃时间（成员列表展示） |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

> UNIQUE(tenant_id, user_id)
> 说明：与 RBAC 的关系——`tenant_members.role` 管空间成员身份与宏观权限档（Owner/Admin/Member/Guest），`user_roles`+`role_menus` 管细粒度权限点；四级角色映射到系统内置 roles，二者并存互补。

### 115. `user_preferences` — 用户偏好

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| user_id | BIGINT | NOT NULL, 逻辑外键 → users.id | 用户 ID（逻辑外键 → users.id） |
| language | VARCHAR(20) | DEFAULT 'zh-CN' | 界面语言 |
| timezone | VARCHAR(50) | DEFAULT 'Asia/Shanghai' | 个人时区 |
| theme | VARCHAR(20) | DEFAULT 'light' CHECK (light/dark/system) | 主题 |
| default_project_id | BIGINT | 逻辑外键 → projects.id | 默认项目（工作台上下文自动带入） |
| home_page | VARCHAR(50) | DEFAULT 'workbench' | 登录后首页 |
| extra | JSONB | | 其他偏好（快捷键/列表密度等） |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

> UNIQUE(tenant_id, user_id)

### 116. `project_templates` — 项目模板

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| ~~基础字段~~ | — | — | 见模板 |
| id | BIGINT | PK | 雪花ID |
| code | VARCHAR(50) | 租户内唯一 | 业务短代码标识（用户按规则生成） |
| name | VARCHAR(255) | NOT NULL | 模板名（敏捷/瀑布/通用，PRD 2.4.1） |
| description | TEXT | | 描述 |
| project_type | VARCHAR(50) | DEFAULT 'scrum' CHECK (scrum/kanban/waterfall) | 项目类型（Scrum/Kanban/瀑布） |
| thumbnail_url | TEXT | | 模板缩略图 URL |
| workflow_preset | JSONB | | 预设状态机（states/state_transitions 种子数据） |
| modules_preset | JSONB | | 预设功能模块开关（intake/sprint/version/estimate） |
| config_preset | JSONB | | 其他预设（估算体系/标签/角色） |
| is_builtin | BOOLEAN | DEFAULT false | 内置模板 |
| usage_count | INT | DEFAULT 0 | 使用次数 |
| status | VARCHAR(50) | DEFAULT 'active' | 状态 |
| deleted | BOOLEAN | DEFAULT false | 软删除标记（true=已删除） |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |
| created_by | BIGINT | NOT NULL | 创建人（逻辑外键 → users.id） |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_by | BIGINT | NOT NULL | 最后更新人（逻辑外键 → users.id） |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

### 117. `project_groups` — 项目集 / 项目分组

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| ~~基础字段~~ | — | — | 见模板 |
| id | BIGINT | PK | 雪花ID |
| code | VARCHAR(50) | | 业务短代码标识（用户按规则生成） |
| name | VARCHAR(255) | NOT NULL | 项目集名称（PRD 2.2 项目分组/分类） |
| description | TEXT | | 描述 |
| owner_id | BIGINT | 逻辑外键 → users.id | PMO 负责人 |
| sort_order | DOUBLE PRECISION | DEFAULT 65535 | 同级排序值（小数插值重排） |
| status | VARCHAR(50) | DEFAULT 'active' | 状态 |
| deleted | BOOLEAN | DEFAULT false | 软删除标记（true=已删除） |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |
| created_by | BIGINT | NOT NULL | 创建人（逻辑外键 → users.id） |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_by | BIGINT | NOT NULL | 最后更新人（逻辑外键 → users.id） |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

---

## 26、评审工作流域（V1.1 新增）

> 对应 PRD 5.3.3 需求评审工作流、10.3.2 文档评审、10.4 知识库评审、4.4.3 迭代复盘评审人。统一多态评审模型，避免每种实体各建一套审批表。

### 118. `review_templates` — 评审模板

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| ~~基础字段~~ | — | — | 见模板 |
| id | BIGINT | PK | 雪花ID |
| code | VARCHAR(50) | | 业务短代码标识（用户按规则生成） |
| name | VARCHAR(255) | NOT NULL | 模板名 |
| entity_scope | VARCHAR(30) | NOT NULL CHECK (requirement/document/knowledge_page/sprint) | 适用对象 |
| dimensions | JSONB | NOT NULL | 评审维度 [{name, max_score, weight, required}]（PRD：评审模板可自定义评审维度） |
| pass_rule | JSONB | | 通过规则 {min_score, min_approvals, require_all} |
| is_builtin | BOOLEAN | DEFAULT false | 是否系统内置 |
| status | VARCHAR(50) | DEFAULT 'active' | 状态 |
| deleted | BOOLEAN | DEFAULT false | 软删除标记（true=已删除） |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |
| created_by | BIGINT | NOT NULL | 创建人（逻辑外键 → users.id） |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_by | BIGINT | NOT NULL | 最后更新人（逻辑外键 → users.id） |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

### 119. `reviews` — 评审单（多态）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| ~~基础字段~~ | — | — | 见模板 |
| id | BIGINT | PK | 雪花ID |
| code | VARCHAR(50) | | 评审单号 |
| name | VARCHAR(500) | NOT NULL | 评审标题 |
| entity_type | VARCHAR(30) | NOT NULL CHECK (requirement/document/knowledge_page/sprint) | 被评审对象类型 |
| entity_id | BIGINT | NOT NULL | 被评审对象ID |
| project_id | BIGINT | 逻辑外键 → projects.id | 冗余便于按项目过滤 |
| template_id | BIGINT | 逻辑外键 → review_templates.id | 使用模板 |
| status | VARCHAR(20) | DEFAULT 'pending' CHECK (pending/approved/rejected/cancelled) | 状态 |
| score_summary | JSONB | | 评分汇总 {avg, per_dimension} |
| round | INT | DEFAULT 1 | 评审轮次（驳回后重提递增） |
| submitted_at | TIMESTAMPTZ | DEFAULT now() | 提交时间 |
| completed_at | TIMESTAMPTZ | | 完成时间 |
| deleted | BOOLEAN | DEFAULT false | 软删除标记（true=已删除） |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |
| created_by | BIGINT | NOT NULL | 创建人（逻辑外键 → users.id） |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_by | BIGINT | NOT NULL | 最后更新人（逻辑外键 → users.id） |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

> INDEX (tenant_id, entity_type, entity_id, status)

### 120. `review_assignments` — 评审人记录

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| review_id | BIGINT | NOT NULL, 逻辑外键 → reviews.id | 评审单 ID（逻辑外键 → reviews.id） |
| reviewer_id | BIGINT | NOT NULL, 逻辑外键 → users.id | 评审人（支持多人评审） |
| status | VARCHAR(20) | DEFAULT 'pending' CHECK (pending/approved/rejected/abstained) | 状态 |
| score | NUMERIC(5,2) | | 总分 |
| dimension_scores | JSONB | | 分维度打分 [{dimension, score, comment}] |
| comment | TEXT | | 评审意见 |
| reviewed_at | TIMESTAMPTZ | | 审核时间 |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |

> UNIQUE(tenant_id, review_id, reviewer_id, round 由 reviews 侧控制)

---

## 27、版本日报告域（V1.1 新增）

> 对应 PRD 3.4.2：发布时自动生成交付报告与 Release Notes（缺陷数、通过率、准出率、7日/30日留存）。

### 121. `version_reports` — 版本日交付报告

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| version_id | BIGINT | NOT NULL, 逻辑外键 → versions.id | 版本日 ID（逻辑外键 → versions.id） |
| report_type | VARCHAR(30) | NOT NULL CHECK (delivery_report/release_notes) | 报告类型（交付报告/发布说明） |
| content_md | TEXT | | 渲染内容（Release Notes 可按模板定制） |
| metrics | JSONB | | 结构化指标 {defect_count, pass_rate, exit_rate, retention_7d, retention_30d, ...} |
| generated_by | BIGINT | 逻辑外键 → users.id；NULL=系统生成 | 生成人（NULL=系统自动） |
| generated_at | TIMESTAMPTZ | DEFAULT now() | 生成时间 |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |

> UNIQUE(tenant_id, version_id, report_type, generated_at::date) — 同日同类型不重复生成

---

## 28、模板、权限与分享域（V1.1 新增）

> 对应 PRD 5.1 需求模板、7.1 缺陷模板库、10.3.2 文档模板、10.4 知识库模板（技术方案/ADR/PRD）与文档权限（空间级四级+文档级覆盖）；分享对应 10.3.1 文档分享、10.5 公开链接（密码+有效期）。

### 122. `knowledge_space_members` — 知识库空间成员

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| space_id | BIGINT | NOT NULL, 逻辑外键 → knowledge_spaces.id | 知识库空间 ID（逻辑外键 → knowledge_spaces.id） |
| user_id | BIGINT | NOT NULL, 逻辑外键 → users.id | 用户 ID（逻辑外键 → users.id） |
| role | VARCHAR(20) | NOT NULL DEFAULT 'viewer' CHECK (owner/admin/editor/viewer) | 空间级四级权限（PRD 10.4 文档权限 P0） |
| page_overrides | JSONB | | 文档级权限覆盖 [{page_id, role}]（继承或覆盖） |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |
| created_by | BIGINT | NOT NULL | 创建人（逻辑外键 → users.id） |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |

> UNIQUE(tenant_id, space_id, user_id)

### 123. `content_templates` — 内容模板（工作项/文档/知识库统一）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| ~~基础字段~~ | — | — | 见模板 |
| id | BIGINT | PK | 雪花ID |
| code | VARCHAR(50) | | 业务短代码标识（用户按规则生成） |
| name | VARCHAR(255) | NOT NULL | 模板名 |
| description | TEXT | | 描述 |
| template_scope | VARCHAR(20) | NOT NULL CHECK (work_item/document/knowledge_page) | 模板域 |
| entity_type | VARCHAR(20) | CHECK (requirement/task/defect) | work_item 域时指定类型 |
| content | JSONB | NOT NULL | 模板内容（预填字段值/富文本骨架） |
| project_id | BIGINT | 逻辑外键 → projects.id | 项目级模板；NULL = 租户级 |
| is_builtin | BOOLEAN | DEFAULT false | 是否系统内置 |
| usage_count | INT | DEFAULT 0 | 使用次数 |
| status | VARCHAR(50) | DEFAULT 'active' | 状态 |
| deleted | BOOLEAN | DEFAULT false | 软删除标记（true=已删除） |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |
| created_by | BIGINT | NOT NULL | 创建人（逻辑外键 → users.id） |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_by | BIGINT | NOT NULL | 最后更新人（逻辑外键 → users.id） |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

### 124. `share_links` — 资源分享链接（多态）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| resource_type | VARCHAR(30) | NOT NULL CHECK (document/knowledge_page/page/dashboard/saved_view/version_report) | 被分享资源类型 |
| resource_id | BIGINT | NOT NULL | 资源 ID |
| share_token | VARCHAR(128) | NOT NULL, UNIQUE | 访问令牌 |
| share_type | VARCHAR(20) | DEFAULT 'public' CHECK (public/password/restricted) | restricted=指定成员可见 |
| allowed_user_ids | JSONB | | restricted 时的用户白名单 |
| password_hash | VARCHAR(128) | | 访问密码 |
| expire_at | TIMESTAMPTZ | | 有效期（PRD：支持设置链接有效期） |
| access_count | INT | DEFAULT 0 | 访问次数 |
| max_access_count | INT | | 最大访问次数限制 |
| is_active | BOOLEAN | DEFAULT true | 是否启用 |
| last_accessed_at | TIMESTAMPTZ | | 最后访问时间 |
| created_by | BIGINT | NOT NULL | 创建人（逻辑外键 → users.id） |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |

> 说明：CMS 的 `page_shares` 保留（含页面特有统计语义）；文档/知识库/仪表盘/视图等分享统一走本表。

---

## 29、个人效率扩展域（V1.1 新增）

> 对应 PRD 9.2 个人工作台（TodoItem 置顶/排序）、9.5 保存/分享视图（个人/团队/系统）、11.1 星标、9.2 今日日程。

### 125. `workbench_todos` — 我的待办

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| user_id | BIGINT | NOT NULL, 逻辑外键 → users.id | 用户 ID（逻辑外键 → users.id） |
| entity_type | VARCHAR(20) | NOT NULL CHECK (requirement/task/defect) | 实体类型（requirement/task/defect） |
| entity_id | BIGINT | NOT NULL | 实体 ID |
| project_id | BIGINT | | 冗余便于跨项目分组 |
| is_pinned | BOOLEAN | DEFAULT false | 置顶 |
| sort_order | DOUBLE PRECISION | DEFAULT 65535 | 手动拖拽排序（PRD 9.2 拖拽调整优先级顺序） |
| note | VARCHAR(500) | | 个人备注 |
| added_at | TIMESTAMPTZ | DEFAULT now() | 加入时间 |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |

> UNIQUE(tenant_id, user_id, entity_type, entity_id)

### 126. `saved_views` — 保存视图（个人/团队/系统）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| ~~基础字段~~ | — | — | 见模板 |
| id | BIGINT | PK | 雪花ID |
| code | VARCHAR(50) | | 业务短代码标识（用户按规则生成） |
| name | VARCHAR(255) | NOT NULL | 视图名（如「PMO 视图」「QA 视图」） |
| project_id | BIGINT | NOT NULL, 逻辑外键 → projects.id | 项目 ID（逻辑外键 → projects.id） |
| view_type | VARCHAR(30) | NOT NULL CHECK (kanban/list/calendar/gantt/timeline/table/spreadsheet) | 视图类型（看板/列表/日历/甘特等） |
| entity_scope | VARCHAR(20) | DEFAULT 'all' CHECK (all/requirement/task/defect) | 视图实体范围（全部/需求/任务/缺陷） |
| scope | VARCHAR(20) | DEFAULT 'personal' CHECK (personal/team/system) | PRD 9.5：个人/团队/系统默认视图 |
| owner_id | BIGINT | NOT NULL, 逻辑外键 → users.id | 负责人 ID（逻辑外键 → users.id） |
| filters | JSONB | | 过滤条件（JSON） |
| sort_config | JSONB | | 排序配置（JSON） |
| columns_config | JSONB | | 列配置（JSON，列表/表格视图） |
| group_by | VARCHAR(50) | | 分组维度 |
| is_default | BOOLEAN | DEFAULT false | 项目默认打开视图 |
| share_token | VARCHAR(64) | UNIQUE | 链接分享（PRD 9.5 分享视图） |
| status | VARCHAR(50) | DEFAULT 'active' | 状态 |
| deleted | BOOLEAN | DEFAULT false | 软删除标记（true=已删除） |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |
| created_by | BIGINT | NOT NULL | 创建人（逻辑外键 → users.id） |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_by | BIGINT | NOT NULL | 最后更新人（逻辑外键 → users.id） |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

> 说明：与 `view_preferences`（一人一项目一视图的偏好记忆）互补——`saved_views` 是可命名、可共享、多份的视图定义。

### 127. `favorite_items` — 星标收藏（多态）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| user_id | BIGINT | NOT NULL | 用户 ID（逻辑外键 → users.id） |
| entity_type | VARCHAR(30) | NOT NULL CHECK (requirement/task/defect/project/document/knowledge_page/sprint/version) | 实体类型（requirement/task/defect） |
| entity_id | BIGINT | NOT NULL | 实体 ID |
| sort_order | DOUBLE PRECISION | DEFAULT 65535 | 同级排序值（小数插值重排） |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |

> UNIQUE(tenant_id, user_id, entity_type, entity_id)

### 128. `calendar_events` — 日程 / 会议

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| ~~基础字段~~ | — | — | 见模板 |
| id | BIGINT | PK | 雪花ID |
| code | VARCHAR(50) | | 业务短代码标识（用户按规则生成） |
| name | VARCHAR(255) | NOT NULL | 标题 |
| event_type | VARCHAR(20) | NOT NULL CHECK (meeting/standup/review/personal/release) | 会议/站会/评审/个人日程/发布 |
| description | TEXT | | 描述 |
| project_id | BIGINT | 逻辑外键 → projects.id | 项目 ID（逻辑外键 → projects.id） |
| sprint_id | BIGINT | 逻辑外键 → sprints.id | 站会关联迭代 |
| entity_type | VARCHAR(20) | | 关联对象（如评审会关联 reviews） |
| entity_id | BIGINT | | 实体 ID |
| start_at | TIMESTAMPTZ | NOT NULL | 开始时间 |
| end_at | TIMESTAMPTZ | NOT NULL | 结束时间 |
| rrule | VARCHAR(255) | | 周期性规则（RFC 5545，如每日站会） |
| organizer_id | BIGINT | NOT NULL, 逻辑外键 → users.id | 组织者 ID（逻辑外键 → users.id） |
| attendees | JSONB | | 参会人 [{user_id, rsvp}] |
| meeting_url | TEXT | | 会议链接 |
| status | VARCHAR(50) | DEFAULT 'active' | 状态 |
| deleted | BOOLEAN | DEFAULT false | 软删除标记（true=已删除） |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |
| created_by | BIGINT | NOT NULL | 创建人（逻辑外键 → users.id） |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_by | BIGINT | NOT NULL | 最后更新人（逻辑外键 → users.id） |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

---

## 30、工程联动与数据任务域（V1.1 新增）

> 对应 PRD 6.1 关联代码提交（智能提交）、11.3 关联修复 PR、9.1 WidgetData 缓存、9.5 导入导出（字段映射/增量导入）、1.2 数据归档/导出。

### 129. `code_links` — 代码关联（commit / branch / PR）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| entity_type | VARCHAR(20) | NOT NULL CHECK (requirement/task/defect) | 实体类型（requirement/task/defect） |
| entity_id | BIGINT | NOT NULL | 实体 ID |
| link_type | VARCHAR(20) | NOT NULL CHECK (commit/branch/pull_request/tag) | 关联类型（提交/分支/PR/标签） |
| provider | VARCHAR(20) | NOT NULL CHECK (github/gitlab/gitee/cnb/other) | 代码平台 |
| repo_name | VARCHAR(255) | | 仓库（org/repo） |
| external_id | VARCHAR(255) | NOT NULL | commit_sha / branch 名 / MR IID |
| title | VARCHAR(500) | | 提交信息/PR 标题 |
| url | TEXT | | 跳转链接 |
| author_name | VARCHAR(255) | | 提交人 |
| status | VARCHAR(30) | | PR 状态（open/merged/closed） |
| committed_at | TIMESTAMPTZ | | 提交时间 |
| raw | JSONB | | 原始 payload |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |

> UNIQUE(tenant_id, entity_type, entity_id, link_type, provider, external_id) — 智能提交幂等
> INDEX (tenant_id, provider, external_id) — 按 commit/PR 反查关联

### 130. `dashboard_widget_data` — 小部件数据缓存

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| widget_id | BIGINT | NOT NULL, 逻辑外键 → dashboard_widgets.id | 小部件 ID（逻辑外键 → dashboard_widgets.id） |
| query_hash | VARCHAR(64) | NOT NULL | 查询条件哈希（含全局时间筛选上下文） |
| data | JSONB | NOT NULL | 渲染数据 |
| refreshed_at | TIMESTAMPTZ | DEFAULT now() | 刷新时间 |
| expires_at | TIMESTAMPTZ | | 按 widget.refresh_interval 计算 |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |

> UNIQUE(tenant_id, widget_id, query_hash)
> 说明：按 PRD 9.1 卡片刷新频率（实时/1h/4h/每日）分级缓存，失效后异步重算。

### 131. `data_jobs` — 数据导入 / 导出 / 归档任务

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| job_type | VARCHAR(20) | NOT NULL CHECK (import/export/archive) | 任务类型（导入/导出/归档） |
| resource_type | VARCHAR(30) | NOT NULL | work_item/member/document/defect/... |
| project_id | BIGINT | 逻辑外键 → projects.id | 项目 ID（逻辑外键 → projects.id） |
| file_name | VARCHAR(255) | | 原始文件名 |
| file_path | TEXT | | 文件存储路径（结果文件或上传文件） |
| file_format | VARCHAR(10) | CHECK (csv/xlsx/json/md/html/pdf/docx) | 文件格式（CSV/Excel/JSON 等） |
| field_mapping | JSONB | | 字段映射配置（PRD 9.5 外部字段→系统字段） |
| dedup_strategy | JSONB | | 增量导入策略 {match_by: external_id/code, on_conflict: update/skip} |
| status | VARCHAR(20) | DEFAULT 'pending' CHECK (pending/running/success/partial/failed/cancelled) | 状态 |
| total_count | INT | | 总数量 |
| success_count | INT | | 成功数量 |
| fail_count | INT | | 失败数量 |
| skip_count | INT | | 跳过数量 |
| result_detail | JSONB | | 逐行错误明细 |
| error_message | TEXT | | 错误信息 |
| started_at | TIMESTAMPTZ | | 开始时间 |
| finished_at | TIMESTAMPTZ | | 完成时间 |
| expires_at | TIMESTAMPTZ | | 结果文件保留期 |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |
| created_by | BIGINT | NOT NULL | 创建人（逻辑外键 → users.id） |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |

---

## 31、跨类型关联与自定义字段域（V1.2 新增）

> 对应架构04 数据模型：跨类型关联（需求↔任务↔缺陷多态关联）、自定义字段扩展表（替代 JSONB 内联方案，支持独立索引与类型校验）。

### 132. `biz_entity_relations` — 跨类型关联

> 跨需求/任务/缺陷的多态关联表。与各类型内部的 *_relations 表（同类型内关联）互补，本表专管跨类型关联（如需求被任务实现、缺陷关联到需求等）。对应 PRD 4.4 需求-任务-缺陷交互模型。

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| project_id | BIGINT | NOT NULL | 项目 ID（逻辑外键 → projects.id） |
| source_type | VARCHAR(20) | NOT NULL CHECK (requirement/task/defect) | 源实体类型 |
| source_id | BIGINT | NOT NULL | 源实体ID |
| target_type | VARCHAR(20) | NOT NULL CHECK (requirement/task/defect) | 目标实体类型 |
| target_id | BIGINT | NOT NULL | 目标实体ID |
| relation_type | VARCHAR(30) | NOT NULL CHECK (implemented_by/relates_to/duplicate/blocked_by/parent_child/found_in/fixed_in/verified_in) | 关联类型 |
| created_by | BIGINT | NOT NULL | 创建人（逻辑外键 → users.id） |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |

> UNIQUE(tenant_id, source_type, source_id, target_type, target_id, relation_type)
> INDEX (tenant_id, source_type, source_id) — 按源实体查关联
> INDEX (tenant_id, target_type, target_id) — 按目标实体反查

### 133. `work_item_custom_fields` — 工作项自定义字段

> 对应 PRD 5.1 需求模板/7.1 缺陷模板的自定义字段能力。取代三张主表内联的 `custom_fields JSONB` 方案，支持独立索引、类型校验（field_schema）与按字段名过滤。每种工作项类型维护独立的扩展表（requirement_ext / task_ext / defect_ext），此处统一描述结构。

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| project_id | BIGINT | NOT NULL | 项目 ID（逻辑外键 → projects.id） |
| entity_type | VARCHAR(20) | NOT NULL CHECK (requirement/task/defect) | 工作项类型 |
| entity_id | BIGINT | NOT NULL | 工作项ID |
| field_name | TEXT | NOT NULL | 自定义字段名 |
| field_value | JSONB | NOT NULL | 字段值（JSONB 支持多种类型） |
| field_schema | JSONB | NOT NULL | 字段类型定义 {type, label, options, validation, ...} |
| created_by | BIGINT | NOT NULL | 创建人（逻辑外键 → users.id） |
| tenant_id | BIGINT | NOT NULL | 租户（组织）ID，数据隔离维度 |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

> UNIQUE(tenant_id, entity_type, entity_id, field_name)
> INDEX (tenant_id, entity_type, field_name) — 按字段名查询
> 说明：实际物理实现为三张独立表 `requirement_ext` / `task_ext` / `defect_ext`，结构一致，FK 分别指向对应主表（ON DELETE CASCADE）。

---

## 分区策略汇总

| 表 | 分区方式 | 保留策略 |
|----|----------|----------|
| entity_activities | 按月 `created_at` RANGE | 12个月后归档冷备 |
| notifications | 按月 `created_at` RANGE | 已读90天后自动清理 |
| notification_deliveries | 按月 `created_at` RANGE | 30天后自动清理 |
| webhook_logs | 按月 `created_at` RANGE | 30天（自动drop分区） |
| rule_executions | 按月 `created_at` RANGE | 30天 |
| login_attempts | 按月 `created_at` RANGE | 30天 |
| audit_logs | 按月 `operated_at` RANGE | 12个月后归档 |
| domain_events | 不分区 | 发布后保留7天 |

---

## 逻辑外键索引模板

```sql
-- 单字段索引
CREATE INDEX "{table}_{field}_idx" ON "{table}" (tenant_id, {field}) WHERE NOT deleted;

-- 复合索引（list视图常用）
CREATE INDEX "{table}_{a}_{b}_idx"
    ON "{table}" (tenant_id, {a}, {b})
    WHERE NOT deleted;

-- code唯一索引
CREATE UNIQUE INDEX "{table}_code_idx"
    ON "{table}" (tenant_id, code)
    WHERE code IS NOT NULL AND NOT deleted;

-- 全文检索降级（无ES时使用）
CREATE INDEX "{table}_fts_idx"
    ON "{table}" USING gin(to_tsvector('simple', coalesce(name,'') || ' ' || coalesce(description_stripped,'')));
```

---

## 索引统计清单

| 索引类型 | 涉及表 | 数量估算 |
|----------|--------|----------|
| 主键 | 全部133张 | 133 |
| 唯一约束 | code字段 + public_id + 联合唯一 | ~90 |
| 逻辑外键索引 | tenant_id + xxx_id | ~165 |
| 复合查询索引 | list视图排序 | ~55 |
| GIN索引 | 全文检索（降级）+ search_tsv | ~6 |

---

## 信创/方言适配点

| PostgreSQL 特性 | 达梦 DM8 | 人大金仓 KingbaseES |
|-----------------|----------|---------------------|
| `gen_random_uuid()` | 应用层生成UUID | 应用层生成UUID |
| `BIGINT` (雪花ID) | `BIGINT` | `BIGINT` |
| `BOOLEAN` | `BIT` 或 `INTEGER` | `BOOLEAN` |
| `JSONB` | `TEXT` + JSON校验 | `JSONB` |
| `TIMESTAMPTZ` | `TIMESTAMP WITH TIME ZONE` | `TIMESTAMPTZ` |
| `INET` | `VARCHAR(45)` | `INET` |
| `CHECK` 约束 | 支持 | 支持 |
| RLS | VPD替代 | 原生RLS |
| 声明式分区 | `RANGE` | `RANGE` |
| `to_tsvector` | 全文索引插件 | 全文索引 |
| `TSVECTOR` | `TEXT` + 全文索引 | `TSVECTOR` |
| `TEXT[]` (数组) | `TEXT` + 分隔符或 JSON | `TEXT[]` |
| `GIN` 索引 | `BITMAP` / 全文索引 | `GIN` |

---
