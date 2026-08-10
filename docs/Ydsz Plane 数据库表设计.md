# Ydsz Plane 数据库表设计

> 基于产品需求文档（PRD V1.0）+ 架构设计文档 + 用户确认的设计约束  
> 数据库：PostgreSQL 16+（兼容达梦/人大金仓）  
> 最后更新：2026-08-10（V1.1 架构师评审修订版）

## 修订记录

| 版本 | 日期 | 变更说明 |
|------|------|----------|
---

---

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

### 多态 entity_type 取值约定

所有多态关联表（`entity_type` + `entity_id`）的 `entity_type` 统一取值域：

| 类别 | 取值 |
|------|------|
| 核心实体 | `requirement` / `task` / `defect` |
| 扩展实体 | `knowledge_page` / `document` / `intake_issue` / `sprint` / `version` / `project` |

- 评论、附件、标签、关注人等多态表默认承载三类核心实体（需求/任务/缺陷）；按 PRD 需要扩展承载知识文档评论、项目文档附件、收件箱附件等扩展实体，各表在 CHECK 约束中声明实际支持的子集
- 跨类型关联（需求↔任务、缺陷→需求/任务）一律走 `entity_relations`，不新增专用外键，保证关联模型统一

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
  │        ├──< requirement / task / defect  ──< entity_assignees / entity_labels / entity_watchers
  │        │        │                                   └──< entity_comments / entity_activities / entity_attachments
  │        │        │                                   └──< entity_relations / entity_dependencies / entity_timelogs
  │        │        └──< code_links（commit/branch/PR 关联）
  │        ├──< modules / labels / entity_modules
  │        ├──< versions ──< version_sprint_relations >── sprints ──< sprint_assignments
  │        │        ├──< version_delivery_snapshots      └──< sprint_snapshots
  │        │        └──< version_reports（交付报告/Release Notes）
  │        ├──< intake_channels ──< intake_issues
  │        └──< content_templates（需求/任务/缺陷/文档/知识库模板）
  │
  ├──< reviews ──< review_assignments（多态：需求评审/文档评审/知识库评审/迭代复盘）
  │        └── review_templates
  ├──< knowledge_spaces ──< knowledge_space_members
  │        └──< knowledge_pages ──< knowledge_page_versions
  │                                 └──< knowledge_page_links
  ├──< documents ──< document_versions / document_links
  ├──< share_links（多态分享：document/knowledge_page/page/dashboard/saved_view）
  ├──< notifications ──< notification_deliveries / notification_subscriptions / notification_digests
  ├──< dashboards ──< dashboard_widgets ──< dashboard_widget_data
  │        └──< dashboard_templates / dashboard_snapshots
  ├──< workbench_configs / workbench_todos / workbench_templates
  ├──< recent_items / view_preferences / saved_views / favorite_items
  ├──< search_history / search_bookmarks
  ├──< calendar_events（站会/评审会/个人日程）
  ├──< automation_rules ──< rule_executions / automation_templates
  ├──< webhooks ──< webhook_logs
  ├──< data_jobs（导入/导出/归档）
  └──< audit_logs / domain_events
```

---

## 一、租户与权限域

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
| created_at | TIMESTAMPTZ | DEFAULT now() | |
| updated_at | TIMESTAMPTZ | DEFAULT now() | |

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
| status | VARCHAR(50) | NOT NULL DEFAULT 'active' CHECK (active/inactive/locked) | |
| is_super_admin | BOOLEAN | DEFAULT false | 系统级超管 |
| last_login_at | TIMESTAMPTZ | | 最后登录时间 |
| deleted | BOOLEAN | DEFAULT false | |
| tenant_id | BIGINT | NOT NULL | 归属租户 |
| created_by | BIGINT | NOT NULL | |
| created_at | TIMESTAMPTZ | DEFAULT now() | |
| updated_by | BIGINT | NOT NULL | |
| updated_at | TIMESTAMPTZ | DEFAULT now() | |

> UNIQUE(tenant_id, email) / UNIQUE(tenant_id, phone) WHERE NOT deleted — 租户内唯一

### 3. `roles` — 角色

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| ~~基础字段~~ | — | — | 见模板 |
| id | BIGINT | PK | 雪花ID |
| code | VARCHAR(50) | 租户内唯一 | 角色编码（PROJECT_ADMIN / DEVELOPER / TESTER） |
| name | VARCHAR(255) | NOT NULL | 角色名称 |
| description | TEXT | | 描述 |
| status | VARCHAR(50) | NOT NULL DEFAULT 'active' | |
| is_system | BOOLEAN | DEFAULT false | 系统内置角色不可删除 |
| role_scope | VARCHAR(50) | DEFAULT 'tenant' CHECK (tenant/project) | 作用范围 |
| deleted | BOOLEAN | DEFAULT false | |
| tenant_id | BIGINT | NOT NULL | |
| created_by | BIGINT | NOT NULL | |
| created_at | TIMESTAMPTZ | DEFAULT now() | |
| updated_by | BIGINT | NOT NULL | |
| updated_at | TIMESTAMPTZ | DEFAULT now() | |

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
| status | VARCHAR(50) | DEFAULT 'active' | |
| deleted | BOOLEAN | DEFAULT false | 菜单一般为系统级，不使用 tenant_id |
| created_by | BIGINT | NOT NULL | |
| created_at | TIMESTAMPTZ | DEFAULT now() | |
| updated_by | BIGINT | NOT NULL | |
| updated_at | TIMESTAMPTZ | DEFAULT now() | |

> 说明：menus 为系统级资源表，不含 tenant_id。通过 role_menus 与角色关联。

### 5. `user_roles` — 用户角色关联

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| user_id | BIGINT | NOT NULL, 逻辑外键 → users.id | |
| role_id | BIGINT | NOT NULL, 逻辑外键 → roles.id | |
| project_id | BIGINT | 逻辑外键 → projects.id（nullable） | 项目级角色可为空 |
| tenant_id | BIGINT | NOT NULL | |
| created_by | BIGINT | NOT NULL | |
| created_at | TIMESTAMPTZ | DEFAULT now() | |

> UNIQUE(tenant_id, user_id, role_id, project_id) — 同一用户不重复授予

### 6. `role_menus` — 角色菜单关联

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| role_id | BIGINT | NOT NULL, 逻辑外键 → roles.id | |
| menu_id | BIGINT | NOT NULL, 逻辑外键 → menus.id | |
| tenant_id | BIGINT | NOT NULL | |
| created_at | TIMESTAMPTZ | DEFAULT now() | |

> UNIQUE(role_id, menu_id)

---

## 二、项目管理域

### 7. `projects` — 项目

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| ~~基础字段~~ | — | — | 见模板 |
| id | BIGINT | PK | 雪花ID |
| code | VARCHAR(50) | NOT NULL | 项目标识（YD / PLANE） |
| name | VARCHAR(255) | NOT NULL | 项目名称 |
| description | TEXT | | 描述 |
| project_type | VARCHAR(50) | DEFAULT 'scrum' CHECK (scrum/kanban/waterfall) | |
| status | VARCHAR(50) | DEFAULT 'active' CHECK (active/archived) | |
| logo_props | JSONB | | Logo/Emoji/封面属性 |
| lead_id | BIGINT | 逻辑外键 → users.id | 项目负责人 |
| network | VARCHAR(20) | NOT NULL DEFAULT 'private' CHECK (private/public/internal) | 网络类型（PRD 2.5：私有/公开/内部，替代布尔 is_public） |
| group_id | BIGINT | 逻辑外键 → project_groups.id | 所属项目集/分组（PRD 2.2，nullable） |
| template_id | BIGINT | 逻辑外键 → project_templates.id | 创建来源模板（追溯用） |
| default_assignee_id | BIGINT | 逻辑外键 → users.id | 默认负责人 |
| progress | SMALLINT | DEFAULT 0 CHECK (0-100) | |
| sort_order | DOUBLE PRECISION | DEFAULT 65535 | |
| deleted | BOOLEAN | DEFAULT false | |
| tenant_id | BIGINT | NOT NULL | |
| created_by | BIGINT | NOT NULL | |
| created_at | TIMESTAMPTZ | DEFAULT now() | |
| updated_by | BIGINT | NOT NULL | |
| updated_at | TIMESTAMPTZ | DEFAULT now() | |
| version | INT | DEFAULT 1 | |

> UNIQUE(tenant_id, code) — 租户内code唯一

### 8. `project_members` — 项目成员

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| project_id | BIGINT | NOT NULL, 逻辑外键 | |
| user_id | BIGINT | NOT NULL, 逻辑外键 | |
| role | VARCHAR(50) | NOT NULL CHECK (admin/developer/tester/viewer) | 项目内角色 |
| join_type | VARCHAR(20) | DEFAULT 'direct' CHECK (direct/invitation) | |
| tenant_id | BIGINT | NOT NULL | |
| joined_at | TIMESTAMPTZ | DEFAULT now() | |
| created_at | TIMESTAMPTZ | DEFAULT now() | |

> UNIQUE(tenant_id, project_id, user_id)

### 9. `project_configs` — 项目功能模块开关

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| project_id | BIGINT | NOT NULL | |
| config_key | VARCHAR(100) | NOT NULL | 模块键（intake/sprint/version/estimate/pages/knowledge） |
| enabled | BOOLEAN | DEFAULT true | |
| config_json | JSONB | | 模块级配置 |
| tenant_id | BIGINT | NOT NULL | |
| created_at | TIMESTAMPTZ | DEFAULT now() | |
| updated_at | TIMESTAMPTZ | DEFAULT now() | |

> UNIQUE(tenant_id, project_id, config_key)

### 10. `project_sequences` — 实体发号器

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| project_id | BIGINT | NOT NULL | |
| entity_type | VARCHAR(20) | NOT NULL CHECK (requirement/task/defect) | 实体类型（需求/任务/缺陷） |
| next_value | BIGINT | DEFAULT 1 | 下一个序号 |
| tenant_id | BIGINT | NOT NULL | |
| updated_at | TIMESTAMPTZ | DEFAULT now() | |

> UNIQUE(tenant_id, project_id, entity_type) — 同一项目每种类型独立发号

---

## 三、状态机域

### 11. `states` — 状态

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| ~~基础字段~~ | — | — | 见模板 |
| id | BIGINT | PK | 雪花ID |
| code | VARCHAR(50) | | 状态编码 |
| slug | VARCHAR(100) | | 机器名（API/过滤用，PRD 6.1） |
| name | VARCHAR(100) | NOT NULL | 状态名 |
| description | TEXT | | |
| color | VARCHAR(7) | | 色值 HEX |
| entity_type | VARCHAR(20) | NOT NULL CHECK (requirement/task/defect) | 适用类型（需求/任务/缺陷独立） |
| project_id | BIGINT | 逻辑外键 → projects.id | 所属项目；NULL = 租户级默认状态（PRD 6.1：每项目独立状态机） |
| state_group | VARCHAR(20) | NOT NULL CHECK (backlog/unstarted/started/completed/cancelled/triage) | |
| is_triage | BOOLEAN | DEFAULT false | 是否分诊态 |
| is_default | BOOLEAN | DEFAULT false | 是否新建默认 |
| sort_order | INT | DEFAULT 0 | |
| status | VARCHAR(50) | DEFAULT 'active' | |
| deleted | BOOLEAN | DEFAULT false | |
| tenant_id | BIGINT | NOT NULL | |
| created_by | BIGINT | NOT NULL | |
| created_at | TIMESTAMPTZ | DEFAULT now() | |
| updated_by | BIGINT | NOT NULL | |
| updated_at | TIMESTAMPTZ | DEFAULT now() | |

### 12. `state_transitions` — 状态流转规则

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| ~~基础字段~~ | — | — | 见模板 |
| id | BIGINT | PK | 雪花ID |
| entity_type | VARCHAR(20) | NOT NULL CHECK (requirement/task/defect) | |
| project_id | BIGINT | 逻辑外键 → projects.id | 项目级流转规则；NULL = 租户级默认 |
| from_state_id | BIGINT | 逻辑外键 | 起始状态 |
| to_state_id | BIGINT | 逻辑外键 | 目标状态 |
| required_fields | JSONB | DEFAULT '[]' | 必填字段（如缺陷解决时必填 root_cause_category） |
| allowed_roles | JSONB | DEFAULT '[]' | 允许操作的角色 |
| status | VARCHAR(50) | DEFAULT 'active' | |
| deleted | BOOLEAN | DEFAULT false | |
| tenant_id | BIGINT | NOT NULL | |
| created_by | BIGINT | NOT NULL | |
| created_at | TIMESTAMPTZ | DEFAULT now() | |
| updated_by | BIGINT | NOT NULL | |
| updated_at | TIMESTAMPTZ | DEFAULT now() | |

> UNIQUE(tenant_id, project_id, entity_type, from_state_id, to_state_id)

---

## 四、估算域

### 13. `estimate_points` — 估算体系

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| ~~基础字段~~ | — | — | 见模板 |
| id | BIGINT | PK | 雪花ID |
| code | VARCHAR(50) | | |
| name | VARCHAR(255) | NOT NULL | |
| description | TEXT | | |
| points_config | JSONB | NOT NULL | 估算选项 [{"label":"1点","value":1}] |
| is_default | BOOLEAN | DEFAULT false | |
| status | VARCHAR(50) | DEFAULT 'active' | |
| deleted | BOOLEAN | DEFAULT false | |
| tenant_id | BIGINT | NOT NULL | |
| created_by | BIGINT | NOT NULL | |
| created_at | TIMESTAMPTZ | DEFAULT now() | |
| updated_by | BIGINT | NOT NULL | |
| updated_at | TIMESTAMPTZ | DEFAULT now() | |

---

## 五、版本日域

### 14. `versions` — 版本日

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| ~~基础字段~~ | — | — | 见模板 |
| id | BIGINT | PK | 雪花ID |
| code | VARCHAR(50) | | 版本编码 |
| name | VARCHAR(255) | NOT NULL | 版本名称 |
| description | TEXT | | |
| project_id | BIGINT | NOT NULL, 逻辑外键 → projects.id | 归属项目（版本日是项目内对象） |
| version_number | VARCHAR(50) | NOT NULL | 语义化版本号（semver: 1.2.3，应用层校验） |
| template_type | VARCHAR(20) | DEFAULT 'standard' CHECK (standard/hotfix/major) | 发布模板类型（PRD 3.2：常规/热修复/大版本） |
| status | VARCHAR(50) | DEFAULT 'planning' CHECK (planning/active/released/archived) | |
| start_date | DATE | | 开始日期 |
| release_date | DATE | | 发布日期 |
| released_at | TIMESTAMPTZ | | 实际发布时间 |
| owner_id | BIGINT | 逻辑外键 → users.id | 版本负责人 |
| checklist | JSONB | | 发布检查清单 |
| release_notes | TEXT | | 发布说明 |
| progress_snapshot | JSONB | | 进度快照（缓存） |
| deleted | BOOLEAN | DEFAULT false | |
| tenant_id | BIGINT | NOT NULL | |
| created_by | BIGINT | NOT NULL | |
| created_at | TIMESTAMPTZ | DEFAULT now() | |
| updated_by | BIGINT | NOT NULL | |
| updated_at | TIMESTAMPTZ | DEFAULT now() | |
| version | INT | DEFAULT 1 | |

### 15. `version_delivery_snapshots` — 版本日交付快照

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| version_id | BIGINT | NOT NULL | |
| snapshot_date | DATE | NOT NULL | |
| total_count | INT | | |
| completed_count | INT | | |
| total_points | NUMERIC(10,2) | | |
| completed_points | NUMERIC(10,2) | | |
| defect_count | INT | | |
| details | JSONB | | 详细数据 |
| tenant_id | BIGINT | NOT NULL | |
| created_at | TIMESTAMPTZ | DEFAULT now() | |

> UNIQUE(tenant_id, version_id, snapshot_date)

---

## 六、迭代域

### 16. `sprints` — 迭代

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| ~~基础字段~~ | — | — | 见模板 |
| id | BIGINT | PK | 雪花ID |
| code | VARCHAR(50) | | |
| name | VARCHAR(255) | NOT NULL | |
| goal | TEXT | | 迭代目标 |
| description | TEXT | | |
| project_id | BIGINT | NOT NULL, 逻辑外键 → projects.id | 归属项目 |
| sprint_number | INT | NOT NULL | 迭代序号 |
| start_date | DATE | NOT NULL | |
| end_date | DATE | NOT NULL | |
| started_at | TIMESTAMPTZ | | 实际启动时间 |
| completed_at | TIMESTAMPTZ | | 实际完成时间 |
| close_strategy | VARCHAR(20) | CHECK (backlog/next_sprint) | 结束时未完成需求/任务/缺陷的处理方式 |
| retrospective_document_id | BIGINT | 逻辑外键 → documents.id | 复盘报告文档（PRD 4.4.3 自动复盘） |
| owner_id | BIGINT | 逻辑外键 → users.id | 迭代负责人（PRD 5.2 owned_by） |
| status | VARCHAR(50) | DEFAULT 'planned' CHECK (planned/active/completed/cancelled) | |
| capacity | NUMERIC(8,2) | | 容量（人天） |
| velocity | NUMERIC(8,2) | | 速率 |
| display_filters | JSONB | | 显示过滤配置 |
| viewport | JSONB | | 视图状态 |
| progress_snapshot | JSONB | | 进度快照缓存 |
| is_mid_sprint_change | BOOLEAN | DEFAULT false | 中途变更标记 |
| deleted | BOOLEAN | DEFAULT false | |
| tenant_id | BIGINT | NOT NULL | |
| created_by | BIGINT | NOT NULL | |
| created_at | TIMESTAMPTZ | DEFAULT now() | |
| updated_by | BIGINT | NOT NULL | |
| updated_at | TIMESTAMPTZ | DEFAULT now() | |
| version | INT | DEFAULT 1 | |

### 17. `version_sprint_relations` — 版本日-迭代关联

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| version_id | BIGINT | NOT NULL | |
| sprint_id | BIGINT | NOT NULL | |
| tenant_id | BIGINT | NOT NULL | |
| created_at | TIMESTAMPTZ | DEFAULT now() | |

> UNIQUE(tenant_id, version_id, sprint_id)

### 18. `sprint_snapshots` — 迭代每日快照（燃尽图数据）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| sprint_id | BIGINT | NOT NULL | |
| snapshot_date | DATE | NOT NULL | |
| remaining_points | NUMERIC(8,2) | | |
| completed_points | NUMERIC(8,2) | | |
| remaining_count | INT | | |
| completed_count | INT | | |
| details | JSONB | | |
| tenant_id | BIGINT | NOT NULL | |
| created_at | TIMESTAMPTZ | DEFAULT now() | |

> UNIQUE(tenant_id, sprint_id, snapshot_date)

### 19. `sprint_assignments` — 迭代-实体关联（多态）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| sprint_id | BIGINT | NOT NULL | |
| entity_type | VARCHAR(20) | NOT NULL CHECK (requirement/task/defect) | |
| entity_id | BIGINT | NOT NULL | 需求/任务/缺陷 ID |
| is_mid_sprint | BOOLEAN | DEFAULT false | 中途加入 |
| added_by | BIGINT | 逻辑外键 → users.id | |
| added_at | TIMESTAMPTZ | DEFAULT now() | |
| removed_at | TIMESTAMPTZ | | 中途移除时间（复盘统计加入/移除，PRD 4.4.3） |
| removed_by | BIGINT | 逻辑外键 → users.id | |
| tenant_id | BIGINT | NOT NULL | |
| created_at | TIMESTAMPTZ | DEFAULT now() | |

> UNIQUE(tenant_id, sprint_id, entity_type, entity_id)

---

## 七、需求表

### 20. `requirement` — 需求

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| ~~基础字段~~ | — | — | 见模板 |
| id | BIGINT | PK | 雪花ID |
| code | VARCHAR(50) | NOT NULL | 需求编码（YD-1001） |
| name | VARCHAR(500) | NOT NULL | 需求标题 |
| description_json | JSONB | | 富文本文档模型 |
| description_html | TEXT | | |
| description_stripped | TEXT | | 纯文本（检索） |
| status | | | 继承基础字段 |
| priority | VARCHAR(20) | DEFAULT 'medium' CHECK (urgent/high/medium/low/none) | |
| severity | SMALLINT | CHECK (1-5) | 业务影响严重程度：致命/严重/一般/提示/建议（PRD 4.1 建议补充，选填） |
| requirement_type | VARCHAR(50) | DEFAULT 'functional' CHECK (functional/non_functional/bug_fix/optimization) | |
| source | VARCHAR(50) | CHECK (customer/internal/competitor/other) | 来源 |
| acceptance_criteria | JSONB | | 验收标准 |
| story_points | SMALLINT | CHECK (0-12) | 故事点 |
| progress | SMALLINT | DEFAULT 0 CHECK (0-100) | |
| project_id | BIGINT | NOT NULL | |
| module_id | BIGINT | | 模块 |
| state_id | BIGINT | NOT NULL | |
| sprint_id | BIGINT | | |
| version_id | BIGINT | | 首次发布版本 |
| estimate_id | BIGINT | | 估算方案 |
| parent_id | BIGINT | 逻辑外键 → requirement.id | 父需求（WBS） |
| assignee_id | BIGINT | | 主要负责人 |
| start_date | DATE | | |
| target_date | DATE | | |
| completed_at | TIMESTAMPTZ | | |
| is_draft | BOOLEAN | DEFAULT false | |
| sort_order | DOUBLE PRECISION | DEFAULT 65535 | |
| external_source | VARCHAR(50) | | 外部系统来源（jira/tapd/csv_import，增量导入用） |
| external_id | VARCHAR(255) | | 外部系统ID（增量导入去重键，PRD 9.5） |
| custom_fields | JSONB | | 自定义字段扩展 |
| deleted | BOOLEAN | DEFAULT false | |
| tenant_id | BIGINT | NOT NULL | |
| created_by | BIGINT | NOT NULL | |
| created_at | TIMESTAMPTZ | DEFAULT now() | |
| updated_by | BIGINT | NOT NULL | |
| updated_at | TIMESTAMPTZ | DEFAULT now() | |

> UNIQUE(tenant_id, code)
> UNIQUE(tenant_id, external_source, external_id) WHERE external_id IS NOT NULL — 导入幂等

### 21. `task` — 任务

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| ~~基础字段~~ | — | — | 见模板 |
| id | BIGINT | PK | 雪花ID |
| code | VARCHAR(50) | NOT NULL | 任务编码 |
| name | VARCHAR(500) | NOT NULL | |
| description_json | JSONB | | |
| description_html | TEXT | | |
| description_stripped | TEXT | | |
| status | | | |
| priority | VARCHAR(20) | DEFAULT 'medium' | |
| task_type | VARCHAR(50) | DEFAULT 'development' CHECK (development/testing/documentation/design/other) | |
| category | VARCHAR(50) | CHECK (frontend/backend/qa/doc/design/other) | |
| estimated_effort | NUMERIC(8,2) | | 预估工时（小时，PRD 6.3.2 工时管理） |
| actual_effort | NUMERIC(8,2) | | 实际工时（小时，汇总自 entity_timelogs） |
| remaining_effort | NUMERIC(8,2) | | 剩余工时（默认 = 预估 - 实际，可手动修正） |
| story_points | SMALLINT | | |
| progress | SMALLINT | DEFAULT 0 CHECK (0-100) | |
| delay_reason | VARCHAR(50) | CHECK (requirement_change/resource/blocked/other) | |
| project_id | BIGINT | NOT NULL | |
| module_id | BIGINT | | |
| state_id | BIGINT | NOT NULL | |
| sprint_id | BIGINT | | |
| version_id | BIGINT | | |
| estimate_id | BIGINT | | |
| parent_id | BIGINT | 逻辑外键 → task.id | 父任务 |
| requirement_id | BIGINT | 逻辑外键 → requirement.id | 归属需求 |
| assignee_id | BIGINT | | |
| start_date | DATE | | |
| target_date | DATE | | |
| completed_at | TIMESTAMPTZ | | |
| sort_order | DOUBLE PRECISION | DEFAULT 65535 | |
| external_source | VARCHAR(50) | | 外部系统来源 |
| external_id | VARCHAR(255) | | 外部系统ID（导入去重） |
| custom_fields | JSONB | | |
| deleted | BOOLEAN | DEFAULT false | |
| tenant_id | BIGINT | NOT NULL | |
| created_by | BIGINT | NOT NULL | |
| created_at | TIMESTAMPTZ | DEFAULT now() | |
| updated_by | BIGINT | NOT NULL | |
| updated_at | TIMESTAMPTZ | DEFAULT now() | |

> UNIQUE(tenant_id, code)
> UNIQUE(tenant_id, external_source, external_id) WHERE external_id IS NOT NULL

### 22. `defect` — 缺陷

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| ~~基础字段~~ | — | — | 见模板 |
| id | BIGINT | PK | 雪花ID |
| code | VARCHAR(50) | NOT NULL | 缺陷编码 |
| name | VARCHAR(500) | NOT NULL | |
| description_json | JSONB | | |
| description_html | TEXT | | |
| description_stripped | TEXT | | |
| status | | | |
| priority | VARCHAR(20) | DEFAULT 'medium' | |
| severity | SMALLINT | NOT NULL CHECK (1-5) | 致命/严重/一般/提示/建议 |
| defect_type | VARCHAR(50) | CHECK (functional/performance/ui/security/compatibility/other) | |
| found_phase | VARCHAR(50) | NOT NULL CHECK (unit/integration/uat/production/customer) | |
| root_cause_category | VARCHAR(50) | CHECK (requirement/technical/environment/data/other) | 解决时必填（经 state_transitions.required_fields 强制） |
| resolution | VARCHAR(30) | CHECK (fixed/cannot_reproduce/wont_fix/duplicate/by_design/external) | 解决方案（PRD 4.3 Resolution） |
| environment | JSONB | | 环境信息 |
| reproduce_steps | JSONB | | 复现步骤 {steps, expected, actual} |
| project_id | BIGINT | NOT NULL | |
| module_id | BIGINT | | |
| state_id | BIGINT | NOT NULL | |
| sprint_id | BIGINT | | |
| found_version_id | BIGINT | | 发现版本 |
| fix_version_id | BIGINT | | 修复版本 |
| requirement_id | BIGINT | | 关联需求 |
| task_id | BIGINT | | 关联任务 |
| verifier_id | BIGINT | | 验证人 |
| assignee_id | BIGINT | | |
| target_date | DATE | | |
| resolved_at | TIMESTAMPTZ | | 修复时间（缺陷龄/MTTR 分析） |
| resolved_by | BIGINT | 逻辑外键 → users.id | 修复人 |
| verified_at | TIMESTAMPTZ | | 验证通过时间 |
| reopen_count | INT | DEFAULT 0 | 重新打开次数（质量分析：返工率） |
| completed_at | TIMESTAMPTZ | | |
| sort_order | DOUBLE PRECISION | DEFAULT 65535 | |
| external_source | VARCHAR(50) | | 外部系统来源 |
| external_id | VARCHAR(255) | | 外部系统ID（导入去重） |
| custom_fields | JSONB | | |
| deleted | BOOLEAN | DEFAULT false | |
| tenant_id | BIGINT | NOT NULL | |
| created_by | BIGINT | NOT NULL | |
| created_at | TIMESTAMPTZ | DEFAULT now() | |
| updated_by | BIGINT | NOT NULL | |
| updated_at | TIMESTAMPTZ | DEFAULT now() | |

> UNIQUE(tenant_id, code)
> UNIQUE(tenant_id, external_source, external_id) WHERE external_id IS NOT NULL

---

## 八、多态关联表

> 以下关联表通过 `entity_type` + `entity_id` 多态模式统一承载需求、任务、缺陷三类实体的关联关系。`entity_type` 默认取值 `requirement` | `task` | `defect`；按「多态 entity_type 取值约定」，评论、附件、标签、关注人四张表额外支持 `knowledge_page` / `document` / `intake_issue` 扩展实体，各表 CHECK 约束声明实际子集。

### M2M 关联（通用结构）

#### 23. `entity_assignees` — 指派人关联（多对多）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| entity_type | VARCHAR(20) | NOT NULL CHECK (requirement/task/defect) | |
| entity_id | BIGINT | NOT NULL | |
| user_id | BIGINT | NOT NULL, 逻辑外键 → users.id | |
| role_type | VARCHAR(20) | DEFAULT 'assignee' CHECK (assignee/reviewer/tester) | |
| tenant_id | BIGINT | NOT NULL | |
| created_at | TIMESTAMPTZ | DEFAULT now() | |

> UNIQUE(tenant_id, entity_type, entity_id, user_id, role_type)

#### 24. `entity_labels` — 标签关联（多对多）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| entity_type | VARCHAR(20) | NOT NULL | |
| entity_id | BIGINT | NOT NULL | |
| label_id | BIGINT | NOT NULL, 逻辑外键 → labels.id | |
| tenant_id | BIGINT | NOT NULL | |
| created_at | TIMESTAMPTZ | DEFAULT now() | |

> UNIQUE(tenant_id, entity_type, entity_id, label_id)

#### 25. `entity_watchers` — 关注关联

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| entity_type | VARCHAR(20) | NOT NULL | |
| entity_id | BIGINT | NOT NULL | |
| user_id | BIGINT | NOT NULL | |
| created_at | TIMESTAMPTZ | DEFAULT now() | |
| tenant_id | BIGINT | NOT NULL | |

> UNIQUE(tenant_id, entity_type, entity_id, user_id)

### 1:N 关联

#### 26. `entity_comments` — 评论

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| ~~基础字段~~ | — | — | 见模板 |
| id | BIGINT | PK | 雪花ID |
| code VARCHAR(50) | — | — | 空 |
| name VARCHAR(255) | — | — | 标题（可空） |
| entity_type | VARCHAR(20) | NOT NULL | |
| entity_id | BIGINT | NOT NULL | |
| content_html | TEXT | | |
| content_stripped | TEXT | | |
| parent_id | BIGINT | 逻辑外键 → entity_comments.id（嵌套回复） | |
| status | VARCHAR(50) | DEFAULT 'active' | |
| deleted | BOOLEAN | DEFAULT false | |
| tenant_id | BIGINT | NOT NULL | |
| created_by | BIGINT | NOT NULL | |
| created_at | TIMESTAMPTZ | DEFAULT now() | |
| updated_by | BIGINT | NOT NULL | |
| updated_at | TIMESTAMPTZ | DEFAULT now() | |

#### 27. `entity_activities` — 活动日志（分区表）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID（时间有序） |
| entity_type | VARCHAR(20) | NOT NULL | |
| entity_id | BIGINT | NOT NULL | |
| project_id | BIGINT | NOT NULL | |
| verb | VARCHAR(50) | NOT NULL | created/updated/status_changed/commented/... |
| field | VARCHAR(100) | | |
| old_value | TEXT | | |
| new_value | TEXT | | |
| actor_id | BIGINT | NOT NULL | |
| tenant_id | BIGINT | NOT NULL | |
| created_at | TIMESTAMPTZ | DEFAULT now() | |

> 分区策略：按月 RANGE 分区，12个月后归档冷备。

#### 28. `entity_attachments` — 附件

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| ~~基础字段~~ | — | — | 见模板 |
| id | BIGINT | PK | 雪花ID |
| code | VARCHAR(50) | | |
| name | VARCHAR(255) | NOT NULL | 文件名 |
| entity_type | VARCHAR(20) | NOT NULL | |
| entity_id | BIGINT | NOT NULL | |
| file_size | BIGINT | | |
| mime_type | VARCHAR(100) | | |
| storage_path | TEXT | NOT NULL | |
| storage_type | VARCHAR(20) | DEFAULT 's3' | |
| status | VARCHAR(50) | DEFAULT 'active' | |
| deleted | BOOLEAN | DEFAULT false | |
| tenant_id | BIGINT | NOT NULL | |
| created_by | BIGINT | NOT NULL | |
| created_at | TIMESTAMPTZ | DEFAULT now() | |
| updated_by | BIGINT | NOT NULL | |
| updated_at | TIMESTAMPTZ | DEFAULT now() | |

#### 29. `entity_timelogs` — 工时记录

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| entity_type | VARCHAR(20) | NOT NULL | |
| entity_id | BIGINT | NOT NULL | |
| user_id | BIGINT | NOT NULL | |
| log_date | DATE | NOT NULL | |
| hours | NUMERIC(6,2) | NOT NULL | |
| started_at | TIMESTAMPTZ | | 计时开始（计时器模式，PRD 6.3.2 自动计时） |
| ended_at | TIMESTAMPTZ | | 计时结束（NULL = 计时中，同一用户同时仅一条进行中） |
| description | VARCHAR(500) | | |
| tenant_id | BIGINT | NOT NULL | |
| created_at | TIMESTAMPTZ | DEFAULT now() | |
| updated_at | TIMESTAMPTZ | DEFAULT now() | |

> UNIQUE(tenant_id, entity_type, entity_id, user_id, log_date)

### 关联关系（带方向的独立记录）

#### 30. `entity_relations` — 关联关系

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| source_type | VARCHAR(20) | NOT NULL | 源类型 |
| source_id | BIGINT | NOT NULL | 源ID |
| target_type | VARCHAR(20) | NOT NULL | 目标类型 |
| target_id | BIGINT | NOT NULL | 目标ID |
| relation_type | VARCHAR(50) | NOT NULL CHECK (duplicate/relates_to/blocked_by/start_before/finish_before/implemented_by) | |
| description | VARCHAR(500) | | |
| tenant_id | BIGINT | NOT NULL | |
| created_by | BIGINT | NOT NULL | |
| created_at | TIMESTAMPTZ | DEFAULT now() | |

> UNIQUE(tenant_id, source_type, source_id, target_type, target_id, relation_type)

#### 31. `entity_dependencies` — 任务依赖

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| dependent_type | VARCHAR(20) | NOT NULL | 依赖方类型 |
| dependent_id | BIGINT | NOT NULL | 依赖方ID |
| depends_on_type | VARCHAR(20) | NOT NULL | |
| depends_on_id | BIGINT | NOT NULL | 被依赖方 |
| dependency_type | VARCHAR(10) | CHECK (FS/SS/FF/SF) | |
| lag_days | INT | DEFAULT 0 | |
| tenant_id | BIGINT | NOT NULL | |
| created_at | TIMESTAMPTZ | DEFAULT now() | |

> UNIQUE(tenant_id, dependent_type, dependent_id, depends_on_type, depends_on_id)

---

## 九、标签与模块

### 32. `labels` — 标签

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| ~~基础字段~~ | — | — | 见模板 |
| id | BIGINT | PK | 雪花ID |
| code | VARCHAR(50) | | |
| name | VARCHAR(100) | NOT NULL | |
| color | VARCHAR(7) | | 色值 |
| description | TEXT | | |
| label_group | VARCHAR(50) | | 分组 |
| project_id | BIGINT | 逻辑外键 → projects.id | 所属项目；NULL = 租户级通用标签 |
| status | VARCHAR(50) | DEFAULT 'active' | |
| deleted | BOOLEAN | DEFAULT false | |
| tenant_id | BIGINT | NOT NULL | |
| created_by | BIGINT | NOT NULL | |
| created_at | TIMESTAMPTZ | DEFAULT now() | |
| updated_by | BIGINT | NOT NULL | |
| updated_at | TIMESTAMPTZ | DEFAULT now() | |

### 33. `modules` — 模块

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| ~~基础字段~~ | — | — | 见模板 |
| id | BIGINT | PK | 雪花ID |
| code | VARCHAR(50) | | |
| name | VARCHAR(255) | NOT NULL | |
| description | TEXT | | |
| project_id | BIGINT | 逻辑外键 → projects.id | 所属项目；NULL = 租户级模块（PRD 3.4：工作空间/项目维度维护） |
| owner_id | BIGINT | | 模块负责人 |
| target_version_id | BIGINT | | 目标交付版本日（PRD 3.4） |
| status | VARCHAR(50) | DEFAULT 'active' | |
| deleted | BOOLEAN | DEFAULT false | |
| tenant_id | BIGINT | NOT NULL | |
| created_by | BIGINT | NOT NULL | |
| created_at | TIMESTAMPTZ | DEFAULT now() | |
| updated_by | BIGINT | NOT NULL | |
| updated_at | TIMESTAMPTZ | DEFAULT now() | |

### 34. `entity_modules` — 模块关联

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| entity_type | VARCHAR(20) | NOT NULL | |
| entity_id | BIGINT | NOT NULL | |
| module_id | BIGINT | NOT NULL | |
| tenant_id | BIGINT | NOT NULL | |
| created_at | TIMESTAMPTZ | DEFAULT now() | |

> UNIQUE(tenant_id, entity_type, entity_id, module_id)

---

## 十、收件箱域

### 35. `intake_channels` — 收件箱渠道

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| ~~基础字段~~ | — | — | 见模板 |
| id | BIGINT | PK | 雪花ID |
| code | VARCHAR(50) | | |
| name | VARCHAR(255) | NOT NULL | |
| description | TEXT | | |
| project_id | BIGINT | NOT NULL, 逻辑外键 → projects.id | 关联项目（PRD 9.3.1：创建 Channel 需关联项目） |
| is_public | BOOLEAN | DEFAULT true | 公开（无需登录）/私有（仅成员） |
| default_entity_type | VARCHAR(20) | DEFAULT 'requirement' | |
| portal_slug | VARCHAR(100) | UNIQUE | 门户URL |
| auto_assign_rules | JSONB | | |
| status | VARCHAR(50) | DEFAULT 'active' | |
| deleted | BOOLEAN | DEFAULT false | |
| tenant_id | BIGINT | NOT NULL | |
| created_by | BIGINT | NOT NULL | |
| created_at | TIMESTAMPTZ | DEFAULT now() | |
| updated_by | BIGINT | NOT NULL | |
| updated_at | TIMESTAMPTZ | DEFAULT now() | |

### 36. `intake_issues` — 收件箱提交

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| ~~基础字段~~ | — | — | 见模板 |
| id | BIGINT | PK | 雪花ID |
| code | VARCHAR(50) | | |
| name | VARCHAR(255) | NOT NULL | 反馈标题 |
| description_json | JSONB | | |
| description_html | TEXT | | |
| channel_id | BIGINT | NOT NULL, 逻辑外键 → intake_channels.id | 来源渠道 |
| project_id | BIGINT | NOT NULL, 逻辑外键 → projects.id | 归属项目（冗余自 channel，便于查询） |
| tracking_code | VARCHAR(32) | NOT NULL, UNIQUE | 外部跟踪 ID（提交者凭此+邮箱查询进度，PRD 9.3.2） |
| priority | VARCHAR(20) | CHECK (urgent/high/medium/low/none) | 提交者可选填 |
| status | VARCHAR(50) | DEFAULT 'open' CHECK (open/accepted/rejected/converted/archived) | |
| reject_reason | VARCHAR(500) | | 拒绝原因（通知提交者） |
| duplicate_of_id | BIGINT | 逻辑外键 → intake_issues.id | 重复标记（AI 重复检测，P2） |
| converted_entity_type | VARCHAR(20) | | 转正后类型 |
| converted_entity_id | BIGINT | | 转正后ID |
| submitter_name | VARCHAR(255) | | 外部提交人 |
| submitter_email | VARCHAR(255) | | |
| submitter_phone | VARCHAR(20) | | |
| submitted_at | TIMESTAMPTZ | DEFAULT now() | |
| reviewed_by | BIGINT | | |
| reviewed_at | TIMESTAMPTZ | | |
| deleted | BOOLEAN | DEFAULT false | |
| tenant_id | BIGINT | NOT NULL | |
| created_by | BIGINT | NOT NULL | |
| created_at | TIMESTAMPTZ | DEFAULT now() | |
| updated_by | BIGINT | NOT NULL | |
| updated_at | TIMESTAMPTZ | DEFAULT now() | |

---

## 十一、知识库域

### 37. `knowledge_spaces` — 知识库空间

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| ~~基础字段~~ | — | — | 见模板 |
| id | BIGINT | PK | 雪花ID |
| code | VARCHAR(50) | | |
| name | VARCHAR(255) | NOT NULL | |
| slug | VARCHAR(100) | NOT NULL | |
| description | TEXT | | |
| cover_image_url | TEXT | | |
| is_private | BOOLEAN | DEFAULT false | |
| default_permission | VARCHAR(20) | DEFAULT 'view' | |
| owner_id | BIGINT | | 空间负责人 |
| project_id | BIGINT | | 关联项目（nullable） |
| status | VARCHAR(50) | DEFAULT 'active' | |
| deleted | BOOLEAN | DEFAULT false | |
| tenant_id | BIGINT | NOT NULL | |
| created_by | BIGINT | NOT NULL | |
| created_at | TIMESTAMPTZ | DEFAULT now() | |
| updated_by | BIGINT | NOT NULL | |
| updated_at | TIMESTAMPTZ | DEFAULT now() | |

### 38. `knowledge_pages` — 知识文档

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| ~~基础字段~~ | — | — | 见模板 |
| id | BIGINT | PK | 雪花ID |
| code | VARCHAR(50) | | |
| name | VARCHAR(255) | NOT NULL | |
| content_md | TEXT | | |
| content_html | TEXT | | |
| content_stripped | TEXT | | |
| space_id | BIGINT | NOT NULL | |
| parent_id | BIGINT | 逻辑外键 → knowledge_pages.id | 目录层级（文档树，PRD 10.4 无限层级） |
| sort_order | DOUBLE PRECISION | DEFAULT 65535 | 同级排序（拖拽排序） |
| status | VARCHAR(50) | DEFAULT 'draft' CHECK (draft/published/archived) | |
| version_num | INT | DEFAULT 1 | |
| is_pinned | BOOLEAN | DEFAULT false | |
| is_featured | BOOLEAN | DEFAULT false | |
| view_count | INT | DEFAULT 0 | |
| reviewer_id | BIGINT | | |
| published_at | TIMESTAMPTZ | | |
| deleted | BOOLEAN | DEFAULT false | |
| tenant_id | BIGINT | NOT NULL | |
| created_by | BIGINT | NOT NULL | |
| created_at | TIMESTAMPTZ | DEFAULT now() | |
| updated_by | BIGINT | NOT NULL | |
| updated_at | TIMESTAMPTZ | DEFAULT now() | |

> 标签复用 `entity_labels`（entity_type='knowledge_page'）；评论复用 `entity_comments`（支持段落锚点写入 content_html 锚点 ID）；订阅复用 `notification_subscriptions`（entity_type='knowledge_page'）。
> 评审通过 `reviews`（entity_type='knowledge_page'）承载，reviewer_id 仅作当前待审人冗余。

### 39. `knowledge_page_versions` — 知识文档版本
| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| page_id | BIGINT | NOT NULL | |
| version_num | INT | NOT NULL | |
| content_md | TEXT | | |
| content_html | TEXT | | |
| change_summary | VARCHAR(500) | | |
| created_by | BIGINT | NOT NULL | |
| created_at | TIMESTAMPTZ | DEFAULT now() | |
| tenant_id | BIGINT | NOT NULL | |

> UNIQUE(tenant_id, page_id, version_num)

### 40. `knowledge_page_links` — 知识文档-实体关联

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| page_id | BIGINT | NOT NULL | |
| entity_type | VARCHAR(20) | NOT NULL CHECK (requirement/task/defect) | |
| entity_id | BIGINT | NOT NULL | |
| relation_type | VARCHAR(20) | DEFAULT 'references' | |
| tenant_id | BIGINT | NOT NULL | |
| created_at | TIMESTAMPTZ | DEFAULT now() | |

> UNIQUE(tenant_id, page_id, entity_type, entity_id)

---

## 十二、文档管理域

### 41. `documents` — 项目文档

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| ~~基础字段~~ | — | — | 见模板 |
| id | BIGINT | PK | 雪花ID |
| code | VARCHAR(50) | | |
| name | VARCHAR(255) | NOT NULL | |
| description_stripped | TEXT | | |
| content_json | JSONB | | |
| document_category | VARCHAR(50) | DEFAULT 'other' CHECK (prd/design/api/test/checklist/requirements/other) | |
| document_format | VARCHAR(20) | DEFAULT 'rich_text' CHECK (rich_text/markdown/wiki) | |
| project_id | BIGINT | NOT NULL, 逻辑外键 → projects.id | 归属项目 |
| parent_id | BIGINT | 逻辑外键 → documents.id | 目录树父节点（PRD 10.3.1 文档目录） |
| sort_order | DOUBLE PRECISION | DEFAULT 65535 | 目录内拖拽排序 |
| status | VARCHAR(50) | DEFAULT 'draft' CHECK (draft/reviewing/approved/archived) | |
| entity_type | VARCHAR(20) | | 关联类型 |
| entity_id | BIGINT | | 关联ID |
| sprint_id | BIGINT | | |
| version_id | BIGINT | | |
| review_required | BOOLEAN | DEFAULT false | |
| approved_by | BIGINT | | |
| approved_at | TIMESTAMPTZ | | |
| deleted | BOOLEAN | DEFAULT false | |
| tenant_id | BIGINT | NOT NULL | |
| created_by | BIGINT | NOT NULL | |
| created_at | TIMESTAMPTZ | DEFAULT now() | |
| updated_by | BIGINT | NOT NULL | |
| updated_at | TIMESTAMPTZ | DEFAULT now() | |

### 42. `document_versions` — 文档版本

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| document_id | BIGINT | NOT NULL | |
| version_num | INT | NOT NULL | |
| content_json | JSONB | | |
| content_html | TEXT | | |
| change_summary | VARCHAR(500) | | |
| created_by | BIGINT | NOT NULL | |
| created_at | TIMESTAMPTZ | DEFAULT now() | |
| tenant_id | BIGINT | NOT NULL | |

> UNIQUE(tenant_id, document_id, version_num)

### 43. `document_links` — 多态关联

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| document_id | BIGINT | NOT NULL | |
| linkable_type | VARCHAR(20) | NOT NULL | requirement/task/defect/sprint/version |
| linkable_id | BIGINT | NOT NULL | |
| tenant_id | BIGINT | NOT NULL | |
| created_at | TIMESTAMPTZ | DEFAULT now() | |

> UNIQUE(tenant_id, document_id, linkable_type, linkable_id)

---

## 十三、通知域

### 44. `notifications` — 通知（分区表）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| ~~基础字段~~ | — | — | 见模板 |
| id | BIGINT | PK | 雪花ID |
| code | VARCHAR(50) | | |
| name | VARCHAR(500) | NOT NULL | 通知摘要（代替 title） |
| recipient_id | BIGINT | NOT NULL | |
| actor_id | BIGINT | | |
| entity_type | VARCHAR(20) | | 关联类型 |
| entity_id | BIGINT | | 关联ID |
| event_type | VARCHAR(50) | NOT NULL | |
| message | TEXT | NOT NULL | |
| message_template | VARCHAR(255) | | 模板标识 |
| extra_data | JSONB | | |
| is_read | BOOLEAN | DEFAULT false | |
| read_at | TIMESTAMPTZ | | |
| is_archived | BOOLEAN | DEFAULT false | 已归档（PRD 9.3：已读通知自动归档） |
| priority | VARCHAR(20) | DEFAULT 'normal' | |
| status | VARCHAR(50) | DEFAULT 'active' | |
| deleted | BOOLEAN | DEFAULT false | |
| tenant_id | BIGINT | NOT NULL | |
| created_by | BIGINT | NOT NULL | |
| created_at | TIMESTAMPTZ | DEFAULT now() | |
| updated_by | BIGINT | NOT NULL | |
| updated_at | TIMESTAMPTZ | DEFAULT now() | |

> 分区策略：按月分区，已读90天后自动清理。

### 45. `notification_subscriptions` — 通知订阅

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| user_id | BIGINT | NOT NULL | |
| project_id | BIGINT | | |
| entity_type | VARCHAR(20) | | |
| entity_id | BIGINT | | |
| event_types | JSONB | NOT NULL | 订阅事件列表 |
| channels | JSONB | NOT NULL | [in_app/email/im/sms] |
| is_enabled | BOOLEAN | DEFAULT true | |
| tenant_id | BIGINT | NOT NULL | |
| created_at | TIMESTAMPTZ | DEFAULT now() | |
| updated_at | TIMESTAMPTZ | DEFAULT now() | |

> UNIQUE(tenant_id, user_id, project_id, entity_type, entity_id)

### 46. `notification_deliveries` — 通知投递记录

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| notification_id | BIGINT | NOT NULL | |
| channel | VARCHAR(20) | NOT NULL | in_app/email/im/sms |
| status | VARCHAR(20) | DEFAULT 'pending' | pending/sent/failed |
| sent_at | TIMESTAMPTZ | | |
| error_code | VARCHAR(50) | | |
| error_message | TEXT | | |
| retry_count | INT | DEFAULT 0 | |
| external_id | VARCHAR(200) | | 外部系统ID |
| tenant_id | BIGINT | NOT NULL | |
| created_at | TIMESTAMPTZ | DEFAULT now() | |

### 47. `notification_digests` — 通知汇总配置

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| user_id | BIGINT | NOT NULL | |
| frequency | VARCHAR(20) | DEFAULT 'realtime' CHECK (realtime/daily/weekly/monthly) | |
| day_of_week | SMALLINT | 1-7 | weekly 时用 |
| time_of_day | time | | daily 发送时间 |
| timezone | VARCHAR(50) | DEFAULT 'Asia/Shanghai' | |
| quiet_hours_start | TIME | | 免打扰开始（如 22:00，PRD 9.3 免打扰时段） |
| quiet_hours_end | TIME | | 免打扰结束（如 08:00；跨天由应用层处理） |
| muted_event_types | JSONB | | 静音事件类型列表 |
| last_sent_at | TIMESTAMPTZ | | |
| is_enabled | BOOLEAN | DEFAULT true | |
| tenant_id | BIGINT | NOT NULL | |
| created_at | TIMESTAMPTZ | DEFAULT now() | |
| updated_at | TIMESTAMPTZ | DEFAULT now() | |

> UNIQUE(tenant_id, user_id)

---

## 十四、效率增强域

### 48. `dashboards` — 仪表盘

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| ~~基础字段~~ | — | — | 见模板 |
| id | BIGINT | PK | 雪花ID |
| code | VARCHAR(50) | | |
| name | VARCHAR(255) | NOT NULL | |
| description | TEXT | | |
| layout_config | JSONB | | 布局 |
| scope | VARCHAR(20) | DEFAULT 'project' CHECK (project/tenant/personal) | |
| project_id | BIGINT | | |
| owner_id | BIGINT | NOT NULL | 归属人 |
| is_default | BOOLEAN | DEFAULT false | |
| is_shared | BOOLEAN | DEFAULT false | |
| status | VARCHAR(50) | DEFAULT 'active' | |
| deleted | BOOLEAN | DEFAULT false | |
| tenant_id | BIGINT | NOT NULL | |
| created_by | BIGINT | NOT NULL | |
| created_at | TIMESTAMPTZ | DEFAULT now() | |
| updated_by | BIGINT | NOT NULL | |
| updated_at | TIMESTAMPTZ | DEFAULT now() | |

### 49. `dashboard_widgets` — 仪表盘小部件

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| ~~基础字段~~ | — | — | 见模板 |
| id | BIGINT | PK | 雪花ID |
| code | VARCHAR(50) | | |
| name | VARCHAR(255) | NOT NULL | |
| dashboard_id | BIGINT | NOT NULL | |
| widget_type | VARCHAR(50) | NOT NULL CHECK (chart/counter/list/table/calendar/gantt) | |
| data_source_type | VARCHAR(50) | | 数据源类型 |
| config | JSONB | NOT NULL | 配置 |
| query_config | JSONB | | 查询配置 |
| position_x | INT | DEFAULT 0 | |
| position_y | INT | DEFAULT 0 | |
| width | INT | DEFAULT 6 | |
| height | INT | DEFAULT 4 | |
| refresh_interval | INT | DEFAULT 300 | 秒 |
| status | VARCHAR(50) | DEFAULT 'active' | |
| deleted | BOOLEAN | DEFAULT false | |
| tenant_id | BIGINT | NOT NULL | |
| created_by | BIGINT | NOT NULL | |
| created_at | TIMESTAMPTZ | DEFAULT now() | |
| updated_by | BIGINT | NOT NULL | |
| updated_at | TIMESTAMPTZ | DEFAULT now() | |

### 50. `dashboard_templates` — 仪表盘模板

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| ~~基础字段~~ | — | — | 见模板 |
| id | BIGINT | PK | 雪花ID |
| code | VARCHAR(50) | UNIQUE | |
| name | VARCHAR(255) | NOT NULL | |
| description | TEXT | | |
| thumbnail_url | TEXT | | |
| layout_config | JSONB | | |
| widgets_template | JSONB | | |
| scope | VARCHAR(20) | DEFAULT 'system' | system/tenant |
| applicable_project_types | JSONB | | 适用项目类型 |
| status | VARCHAR(50) | DEFAULT 'active' | |
| deleted | BOOLEAN | DEFAULT false | |
| tenant_id | BIGINT | NOT NULL | |
| created_by | BIGINT | NOT NULL | |
| created_at | TIMESTAMPTZ | DEFAULT now() | |
| updated_by | BIGINT | NOT NULL | |
| updated_at | TIMESTAMPTZ | DEFAULT now() | |

### 51. `dashboard_snapshots` — 仪表盘快照

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| dashboard_id | BIGINT | NOT NULL | |
| snapshot_name | VARCHAR(255) | | |
| snapshot_data | JSONB | NOT NULL | |
| created_by | BIGINT | NOT NULL | |
| created_at | TIMESTAMPTZ | DEFAULT now() | |
| tenant_id | BIGINT | NOT NULL | |

### 52. `workbench_configs` — 个人工作台

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| user_id | BIGINT | NOT NULL | |
| layout_config | JSONB | NOT NULL | |
| widgets_config | JSONB | NOT NULL | |
| quick_filters | JSONB | | |
| tenant_id | BIGINT | NOT NULL | |
| created_at | TIMESTAMPTZ | DEFAULT now() | |
| updated_at | TIMESTAMPTZ | DEFAULT now() | |

> UNIQUE(tenant_id, user_id)

### 53. `workbench_templates` — 工作台模板

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| ~~基础字段~~ | — | — | 见模板 |
| id | BIGINT | PK | 雪花ID |
| code | VARCHAR(50) | UNIQUE | |
| name | VARCHAR(255) | NOT NULL | |
| description | TEXT | | |
| layout_config | JSONB | | |
| widgets_config | JSONB | | |
| applicable_roles | JSONB | | |
| status | VARCHAR(50) | DEFAULT 'active' | |
| deleted | BOOLEAN | DEFAULT false | |
| tenant_id | BIGINT | NOT NULL | |
| created_by | BIGINT | NOT NULL | |
| created_at | TIMESTAMPTZ | DEFAULT now() | |
| updated_by | BIGINT | NOT NULL | |
| updated_at | TIMESTAMPTZ | DEFAULT now() | |

### 54. `recent_items` — 最近访问

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| user_id | BIGINT | NOT NULL | |
| entity_type | VARCHAR(20) | NOT NULL | |
| entity_id | BIGINT | NOT NULL | |
| project_id | BIGINT | | |
| view_count | INT | DEFAULT 1 | |
| last_accessed_at | TIMESTAMPTZ | DEFAULT now() | |
| tenant_id | BIGINT | NOT NULL | |
| created_at | TIMESTAMPTZ | DEFAULT now() | |

> UNIQUE(tenant_id, user_id, entity_type, entity_id)

### 55. `view_preferences` — 视图偏好

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| user_id | BIGINT | NOT NULL | |
| project_id | BIGINT | NOT NULL | |
| view_type | VARCHAR(30) | NOT NULL CHECK (kanban/list/calendar/gantt/timeline/table) | |
| entity_scope | VARCHAR(20) | DEFAULT 'all' CHECK (all/requirement/task/defect) | 视图范围 |
| filters | JSONB | | 过滤条件 |
| sort_config | JSONB | | 排序 |
| columns_config | JSONB | | 列配置 |
| group_by | VARCHAR(50) | | 分组字段 |
| viewport_state | JSONB | | 状态 |
| tenant_id | BIGINT | NOT NULL | |
| created_at | TIMESTAMPTZ | DEFAULT now() | |
| updated_at | TIMESTAMPTZ | DEFAULT now() | |

> UNIQUE(tenant_id, user_id, project_id, view_type, entity_scope)

### 56. `search_history` — 搜索历史

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| user_id | BIGINT | NOT NULL | |
| query_text | VARCHAR(500) | NOT NULL | |
| filters_snapshot | JSONB | | |
| entity_types | JSONB | | 搜索范围 |
| result_count | INT | | |
| response_time_ms | INT | | |
| last_used_at | TIMESTAMPTZ | DEFAULT now() | |
| use_count | INT | DEFAULT 1 | |
| tenant_id | BIGINT | NOT NULL | |
| created_at | TIMESTAMPTZ | DEFAULT now() | |

### 57. `search_bookmarks` — 搜索书签

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| ~~基础字段~~ | — | — | 见模板 |
| id | BIGINT | PK | 雪花ID |
| code | VARCHAR(50) | | |
| name | VARCHAR(255) | NOT NULL | |
| query_text | VARCHAR(500) | NOT NULL | |
| filters_snapshot | JSONB | | |
| entity_types | JSONB | | |
| status | VARCHAR(50) | DEFAULT 'active' | |
| deleted | BOOLEAN | DEFAULT false | |
| tenant_id | BIGINT | NOT NULL | |
| created_by | BIGINT | NOT NULL | |
| created_at | TIMESTAMPTZ | DEFAULT now() | |
| updated_by | BIGINT | NOT NULL | |
| updated_at | TIMESTAMPTZ | DEFAULT now() | |

---

## 十五、自动化域

### 58. `automation_rules` — 自动化规则

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| ~~基础字段~~ | — | — | 见模板 |
| id | BIGINT | PK | 雪花ID |
| code | VARCHAR(50) | | |
| name | VARCHAR(255) | NOT NULL | |
| description | TEXT | | |
| trigger_event | VARCHAR(100) | NOT NULL | 触发事件 |
| trigger_config | JSONB | | 触发配置 |
| conditions | JSONB | NOT NULL | 执行条件 |
| actions | JSONB | NOT NULL | 执行动作 |
| entity_scope | VARCHAR(20) | DEFAULT 'all' CHECK (all/requirement/task/defect) | 适用类型 |
| priority | INT | DEFAULT 0 | 优先级（越小越先执行） |
| execution_count | INT | DEFAULT 0 | |
| success_count | INT | DEFAULT 0 | |
| fail_count | INT | DEFAULT 0 | |
| last_executed_at | TIMESTAMPTZ | | |
| last_error | TEXT | | |
| dry_run | BOOLEAN | DEFAULT false | 试运行模式 |
| status | VARCHAR(50) | DEFAULT 'active' CHECK (active/paused/disabled) | |
| deleted | BOOLEAN | DEFAULT false | |
| tenant_id | BIGINT | NOT NULL | |
| created_by | BIGINT | NOT NULL | |
| created_at | TIMESTAMPTZ | DEFAULT now() | |
| updated_by | BIGINT | NOT NULL | |
| updated_at | TIMESTAMPTZ | DEFAULT now() | |

### 59. `rule_executions` — 规则执行记录（分区表）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| rule_id | BIGINT | NOT NULL | |
| trigger_event_type | VARCHAR(100) | NOT NULL | |
| trigger_entity_type | VARCHAR(20) | | |
| trigger_entity_id | BIGINT | | |
| input_data | JSONB | | |
| results | JSONB | | |
| status | VARCHAR(20) | DEFAULT 'pending' CHECK (pending/running/success/failed/cancelled) | |
| error_message | TEXT | | |
| duration_ms | INT | | |
| executed_at | TIMESTAMPTZ | DEFAULT now() | |
| tenant_id | BIGINT | NOT NULL | |
| created_at | TIMESTAMPTZ | DEFAULT now() | |

> 分区策略：按月分区，30天TTL。

### 60. `automation_templates` — 自动化模板

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| ~~基础字段~~ | — | — | 见模板 |
| id | BIGINT | PK | 雪花ID |
| code | VARCHAR(50) | UNIQUE | |
| name | VARCHAR(255) | NOT NULL | |
| description | TEXT | | |
| category | VARCHAR(50) | | |
| icon | VARCHAR(100) | | |
| trigger_event | VARCHAR(100) | NOT NULL | |
| trigger_config_template | JSONB | | |
| conditions_template | JSONB | | |
| actions_template | JSONB | | |
| applicable_entity_types | JSONB | | |
| applicable_project_types | JSONB | | |
| popularity | INT | DEFAULT 0 | 使用次数 |
| rating | NUMERIC(3,2) | | 评分 |
| is_builtin | BOOLEAN | DEFAULT false | |
| status | VARCHAR(50) | DEFAULT 'active' | |
| deleted | BOOLEAN | DEFAULT false | |
| tenant_id | BIGINT | NOT NULL | |
| created_by | BIGINT | NOT NULL | |
| created_at | TIMESTAMPTZ | DEFAULT now() | |
| updated_by | BIGINT | NOT NULL | |
| updated_at | TIMESTAMPTZ | DEFAULT now() | |

---

## 十六、集成域

### 61. `webhooks` — Webhook 配置

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| ~~基础字段~~ | — | — | 见模板 |
| id | BIGINT | PK | 雪花ID |
| code | VARCHAR(50) | | |
| name | VARCHAR(255) | NOT NULL | |
| description | TEXT | | |
| url | TEXT | NOT NULL | |
| secret | VARCHAR(128) | NOT NULL | 签名密钥（HMAC-SHA256） |
| project_id | BIGINT | 逻辑外键 → projects.id | 项目级 Webhook；NULL = 租户级 |
| events | JSONB | NOT NULL | 订阅事件（30+ 事件类型，PRD 9.6） |
| entity_scope | VARCHAR(20) | DEFAULT 'all' | 实体范围 |
| headers | JSONB | | 自定义请求头 |
| timeout_ms | INT | DEFAULT 10000 | |
| retry_config | JSONB | | 重试配置 |
| is_active | BOOLEAN | DEFAULT true | |
| last_triggered_at | TIMESTAMPTZ | | |
| last_error | TEXT | | |
| status | VARCHAR(50) | DEFAULT 'active' | |
| deleted | BOOLEAN | DEFAULT false | |
| tenant_id | BIGINT | NOT NULL | |
| created_by | BIGINT | NOT NULL | |
| created_at | TIMESTAMPTZ | DEFAULT now() | |
| updated_by | BIGINT | NOT NULL | |
| updated_at | TIMESTAMPTZ | DEFAULT now() | |

### 62. `webhook_logs` — Webhook 投递日志（分区表）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| webhook_id | BIGINT | NOT NULL | |
| event_type | VARCHAR(100) | NOT NULL | |
| entity_type | VARCHAR(20) | | |
| entity_id | BIGINT | | |
| payload | JSONB | NOT NULL | |
| request_headers | JSONB | | |
| response_status | INT | | |
| response_body | TEXT | | |
| response_time_ms | INT | | |
| retry_count | INT | DEFAULT 0 | |
| delivery_status | VARCHAR(20) | DEFAULT 'pending' CHECK (pending/sent/failed) | |
| error_message | TEXT | | |
| delivered_at | TIMESTAMPTZ | | |
| tenant_id | BIGINT | NOT NULL | |
| created_at | TIMESTAMPTZ | DEFAULT now() | |

> 分区策略：按月分区，30天TTL。

### 63. `api_tokens` — API Token

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| ~~基础字段~~ | — | — | 见模板 |
| id | BIGINT | PK | 雪花ID |
| code | VARCHAR(50) | | |
| name | VARCHAR(255) | NOT NULL | |
| token_hash | TEXT | NOT NULL, UNIQUE | Token哈希 |
| token_prefix | VARCHAR(10) | | Token前缀 |
| scopes | JSONB | NOT NULL | 权限范围 |
| ip_whitelist | JSONB | | IP白列表 |
| rate_limit | INT | | 速率限制 |
| entity_type | VARCHAR(20) | DEFAULT 'user' CHECK (user/service) | |
| expires_at | TIMESTAMPTZ | | 过期时间（PRD 1.6 api_token.expires_at） |
| last_used_at | TIMESTAMPTZ | | |
| last_used_ip | INET | | |
| status | VARCHAR(50) | DEFAULT 'active' | |
| deleted | BOOLEAN | DEFAULT false | |
| tenant_id | BIGINT | NOT NULL | |
| created_by | BIGINT | NOT NULL | |
| created_at | TIMESTAMPTZ | DEFAULT now() | |
| updated_by | BIGINT | NOT NULL | |
| updated_at | TIMESTAMPTZ | DEFAULT now() | |

### 64. `sso_providers` — SSO 配置

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| ~~基础字段~~ | — | — | 见模板 |
| id | BIGINT | PK | 雪花ID |
| code | VARCHAR(50) | | |
| name | VARCHAR(100) | NOT NULL | |
| provider_type | VARCHAR(20) | NOT NULL CHECK (oidc/saml/ldap) | |
| description | TEXT | | |
| config | JSONB | NOT NULL | 配置（含敏感字段加密） |
| default_role_id | BIGINT | | 默认角色 |
| user_mapping_rules | JSONB | | 用户属性映射 |
| auto_create_user | BOOLEAN | DEFAULT true | |
| status | VARCHAR(50) | DEFAULT 'active' | |
| deleted | BOOLEAN | DEFAULT false | |
| tenant_id | BIGINT | NOT NULL | |
| created_by | BIGINT | NOT NULL | |
| created_at | TIMESTAMPTZ | DEFAULT now() | |
| updated_by | BIGINT | NOT NULL | |
| updated_at | TIMESTAMPTZ | DEFAULT now() | |

### 65. `sso_sessions` — SSO 会话

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| user_id | BIGINT | NOT NULL | |
| provider_id | BIGINT | NOT NULL | |
| external_user_id | VARCHAR(255) | | 外部系统用户ID |
| session_data | JSONB | | |
| ip_address | INET | | |
| user_agent | TEXT | | |
| login_at | TIMESTAMPTZ | DEFAULT now() | |
| logout_at | TIMESTAMPTZ | | |
| expires_at | TIMESTAMPTZ | | |
| tenant_id | BIGINT | NOT NULL | |
| created_at | TIMESTAMPTZ | DEFAULT now() | |

> UNIQUE(tenant_id, provider_id, external_user_id)

---

## 十七、事件总线域

### 66. `domain_events` — 领域事件（Outbox）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| aggregate_type | VARCHAR(50) | NOT NULL | |
| aggregate_id | BIGINT | NOT NULL | |
| event_type | VARCHAR(100) | NOT NULL | |
| event_version | INT | DEFAULT 1 | |
| payload | JSONB | NOT NULL | |
| metadata | JSONB | | |
| trace_id | VARCHAR(64) | | 链路追踪 |
| occurred_at | TIMESTAMPTZ | DEFAULT now() | |
| published_at | TIMESTAMPTZ | | |
| published_to | JSONB | [] | 已发布的消费者 |
| tenant_id | BIGINT | NOT NULL | |
| created_at | TIMESTAMPTZ | DEFAULT now() | |

### 67. `processed_events` — 消费者幂等记录

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| consumer | VARCHAR(100) | NOT NULL | 消费者标识 |
| event_id | BIGINT | NOT NULL | |
| processed_at | TIMESTAMPTZ | DEFAULT now() | |
| result_status | VARCHAR(20) | DEFAULT 'success' | |
| error_message | TEXT | | |
| tenant_id | BIGINT | NOT NULL | |

> PRIMARY KEY (tenant_id, consumer, event_id)

### 68. `dlq_events` — 死信队列

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| event_id | BIGINT | NOT NULL | |
| consumer | VARCHAR(100) | NOT NULL | |
| error_type | VARCHAR(50) | | |
| error_message | TEXT | | |
| retry_count | INT | DEFAULT 0 | |
| original_payload | JSONB | | |
| status | VARCHAR(20) | DEFAULT 'pending' CHECK (pending/processing/resolved/discarded) | |
| resolved_at | TIMESTAMPTZ | | |
| resolved_by | BIGINT | | |
| tenant_id | BIGINT | NOT NULL | |
| created_at | TIMESTAMPTZ | DEFAULT now() | |
| updated_at | TIMESTAMPTZ | DEFAULT now() | |

### 69. `idempotency_keys` — API幂等键

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| key | VARCHAR(64) | PK | 幂等键Hash |
| user_id | BIGINT | NOT NULL | |
| method | VARCHAR(10) | NOT NULL | 请求方法 |
| path | VARCHAR(500) | NOT NULL | 请求路径 |
| request_hash | VARCHAR(64) | | 请求体Hash |
| response_body | JSONB | | 缓存响应 |
| response_status | INT | | 缓存状态码 |
| expires_at | TIMESTAMPTZ | NOT NULL | |
| created_at | TIMESTAMPTZ | DEFAULT now() | |
| tenant_id | BIGINT | NOT NULL | |

---

## 十八、安全审计域

### 70. `password_reset_tokens` — 密码重置Token

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| user_id | BIGINT | NOT NULL | |
| token_hash | TEXT | NOT NULL, UNIQUE | |
| expires_at | TIMESTAMPTZ | NOT NULL | |
| used_at | TIMESTAMPTZ | | |
| ip_address | INET | | |
| user_agent | TEXT | | |
| tenant_id | BIGINT | NOT NULL | |
| created_at | TIMESTAMPTZ | DEFAULT now() | |

### 71. `audit_logs` — 审计日志（分区表）

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
| entity_id | BIGINT | | |
| method | VARCHAR(10) | | HTTP方法 |
| path | VARCHAR(500) | | |
| ip_address | INET | | |
| user_agent | TEXT | | |
| request_body | JSONB | | |
| response_status | INT | | |
| changes | JSONB | | 变更详情 |
| risk_level | VARCHAR(20) | DEFAULT 'low' | |
| session_id | VARCHAR(64) | | |
| trace_id | VARCHAR(64) | | |
| tenant_id | BIGINT | NOT NULL | |
| operated_at | TIMESTAMPTZ | DEFAULT now() | |

> 分区策略：按月分区，12个月后归档。

### 72. `login_attempts` — 登录尝试记录

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| user_id | BIGINT | | |
| username_attempted | VARCHAR(255) | | |
| success | BOOLEAN | NOT NULL | |
| failure_reason | VARCHAR(100) | | |
| ip_address | INET | | |
| user_agent | TEXT | | |
| geo_location | VARCHAR(100) | | |
| tenant_id | BIGINT | NOT NULL | |
| created_at | TIMESTAMPTZ | DEFAULT now() | |

> 按月分区，30天TTL。

### 73. `invitations` — 邀请

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| ~~基础字段~~ | — | — | 见模板 |
| id | BIGINT | PK | 雪花ID |
| code | VARCHAR(50) | | |
| name | VARCHAR(255) | NOT NULL | 备注名 |
| invitee_email | VARCHAR(255) | NOT NULL | |
| invitee_name | VARCHAR(255) | | |
| project_id | BIGINT | | |
| role | VARCHAR(50) | NOT NULL | 邀请角色 |
| invite_type | VARCHAR(20) | DEFAULT 'project' CHECK (tenant/project) | |
| token_hash | VARCHAR(128) | NOT NULL, UNIQUE | |
| expire_at | TIMESTAMPTZ | NOT NULL | |
| accepted_at | TIMESTAMPTZ | | |
| accepted_user_id | BIGINT | | |
| rejected_at | TIMESTAMPTZ | | |
| status | VARCHAR(20) | DEFAULT 'pending' CHECK (pending/accepted/rejected/expired) | |
| welcome_message | TEXT | | |
| deleted | BOOLEAN | DEFAULT false | |
| tenant_id | BIGINT | NOT NULL | |
| created_by | BIGINT | NOT NULL | |
| created_at | TIMESTAMPTZ | DEFAULT now() | |
| updated_by | BIGINT | NOT NULL | |
| updated_at | TIMESTAMPTZ | DEFAULT now() | |

---

## 十九、效能度量域

### 74. `metric_snapshots` — 效能指标快照

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| metric_type | VARCHAR(50) | NOT NULL | 指标类型 |
| metric_name | VARCHAR(100) | NOT NULL | 指标名称 |
| project_id | BIGINT | | |
| sprint_id | BIGINT | | |
| version_id | BIGINT | | |
| entity_scope | VARCHAR(20) | DEFAULT 'all' | |
| aggregation_dim | VARCHAR(50) | | 维度（type/module/state） |
| dim_value | VARCHAR(100) | | 维度值 |
| value | NUMERIC(16,4) | | 指标值 |
| count_val | BIGINT | | 计数值 |
| ratio_val | NUMERIC(8,4) | | 比率值 |
| details | JSONB | | 详细数据 |
| snapshot_date | DATE | NOT NULL | |
| snapshot_hour | SMALLINT | | 小时粒度 |
| period_type | VARCHAR(20) | DEFAULT 'daily' CHECK (hourly/daily/weekly/monthly) | |
| tenant_id | BIGINT | NOT NULL | |
| created_at | TIMESTAMPTZ | DEFAULT now() | |

> UNIQUE(tenant_id, metric_type, project_id, sprint_id, aggregation_dim, dim_value, snapshot_date, period_type)

### 75. `metric_adjustments` — 指标调整

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| ~~基础字段~~ | — | — | 见模板 |
| id | BIGINT | PK | 雪花ID |
| code | VARCHAR(50) | | |
| name | VARCHAR(255) | NOT NULL | |
| metric_type | VARCHAR(50) | NOT NULL | |
| project_id | BIGINT | | |
| adjustment_date | DATE | NOT NULL | |
| delta | NUMERIC(16,4) | NOT NULL | |
| reason | VARCHAR(500) | | |
| status | VARCHAR(50) | DEFAULT 'approved' | |
| deleted | BOOLEAN | DEFAULT false | |
| tenant_id | BIGINT | NOT NULL | |
| created_by | BIGINT | NOT NULL | |
| created_at | TIMESTAMPTZ | DEFAULT now() | |
| updated_by | BIGINT | NOT NULL | |
| updated_at | TIMESTAMPTZ | DEFAULT now() | |

### 76. `metric_definitions` — 指标定义

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| ~~基础字段~~ | — | — | 见模板 |
| id | BIGINT | PK | 雪花ID |
| code | VARCHAR(50) | UNIQUE | |
| name | VARCHAR(100) | NOT NULL | |
| description | TEXT | | |
| metric_category | VARCHAR(50) | NOT NULL | velocity/burndown/lead_time/... |
| calculation_method | VARCHAR(50) | | 计算方式 |
| calculation_formula | TEXT | | |
| unit | VARCHAR(20) | | 单位 |
| default_config | JSONB | | |
| dimensions | JSONB | | 可用维度 |
| visualization_config | JSONB | | 可视化配置 |
| is_builtin | BOOLEAN | DEFAULT false | |
| status | VARCHAR(50) | DEFAULT 'active' | |
| deleted | BOOLEAN | DEFAULT false | |
| tenant_id | BIGINT | NOT NULL | |
| created_by | BIGINT | NOT NULL | |
| created_at | TIMESTAMPTZ | DEFAULT now() | |
| updated_by | BIGINT | NOT NULL | |
| updated_at | TIMESTAMPTZ | DEFAULT now() | |

---

## 二十、治理域

### 77. `risk_rules` — 风险规则

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| ~~基础字段~~ | — | — | 见模板 |
| id | BIGINT | PK | 雪花ID |
| code | VARCHAR(50) | | |
| name | VARCHAR(255) | NOT NULL | |
| description | TEXT | | |
| risk_type | VARCHAR(50) | NOT NULL CHECK (delay/quality/resource/scope/other) | |
| trigger_conditions | JSONB | NOT NULL | 触发条件 |
| threshold_config | JSONB | | 阈值 |
| severity | VARCHAR(20) | DEFAULT 'medium' CHECK (low/medium/high/critical) | |
| auto_actions | JSONB | | 自动动作 |
| applicable_entity_types | JSONB | | 适用实体（requirement/task/defect） |
| evaluation_frequency | VARCHAR(20) | DEFAULT 'daily' | |
| status | VARCHAR(50) | DEFAULT 'active' | |
| deleted | BOOLEAN | DEFAULT false | |
| tenant_id | BIGINT | NOT NULL | |
| created_by | BIGINT | NOT NULL | |
| created_at | TIMESTAMPTZ | DEFAULT now() | |
| updated_by | BIGINT | NOT NULL | |
| updated_at | TIMESTAMPTZ | DEFAULT now() | |

### 78. `risk_alerts` — 风险告警

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| ~~基础字段~~ | — | — | 见模板 |
| id | BIGINT | PK | 雪花ID |
| code | VARCHAR(50) | | |
| name | VARCHAR(500) | NOT NULL | |
| rule_id | BIGINT | NOT NULL | |
| entity_type | VARCHAR(20) | | |
| entity_id | BIGINT | | |
| project_id | BIGINT | | |
| sprint_id | BIGINT | | |
| alert_level | VARCHAR(20) | NOT NULL CHECK (info/warning/critical) | |
| alert_content | TEXT | | |
| metric_snapshot | JSONB | | |
| suggested_actions | JSONB | | |
| is_acknowledged | BOOLEAN | DEFAULT false | |
| acknowledged_by | BIGINT | | |
| acknowledged_at | TIMESTAMPTZ | | |
| resolution_note | TEXT | | |
| status | VARCHAR(50) | DEFAULT 'open' | |
| deleted | BOOLEAN | DEFAULT false | |
| tenant_id | BIGINT | NOT NULL | |
| created_by | BIGINT | NOT NULL | |
| created_at | TIMESTAMPTZ | DEFAULT now() | |
| updated_by | BIGINT | NOT NULL | |
| updated_at | TIMESTAMPTZ | DEFAULT now() | |

### 79. `deployment_events` — 部署事件

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| ~~基础字段~~ | — | — | 见模板 |
| id | BIGINT | PK | 雪花ID |
| code | VARCHAR(50) | | |
| name | VARCHAR(255) | NOT NULL | |
| description | TEXT | | |
| project_id | BIGINT | | |
| version_id | BIGINT | | |
| environment | VARCHAR(50) | NOT NULL | dev/staging/production |
| deployment_type | VARCHAR(30) | DEFAULT 'standard' | standard/rollback/hotfix |
| deployed_version | VARCHAR(100) | | 部署版本 |
| previous_version | VARCHAR(100) | | 上一版本 |
| changes_summary | TEXT | | |
| artifacts | JSONB | | 构建产物 |
| status | VARCHAR(50) | DEFAULT 'pending' CHECK (pending/in_progress/success/failed/rolled_back) | |
| started_at | TIMESTAMPTZ | | |
| completed_at | TIMESTAMPTZ | | |
| duration_ms | INT | | |
| trigger_source | VARCHAR(30) | | manual/webhook/automation |
| metadata | JSONB | | |
| deleted | BOOLEAN | DEFAULT false | |
| tenant_id | BIGINT | NOT NULL | |
| created_by | BIGINT | NOT NULL | |
| created_at | TIMESTAMPTZ | DEFAULT now() | |
| updated_by | BIGINT | NOT NULL | |
| updated_at | TIMESTAMPTZ | DEFAULT now() | |

---

## 二十一、内容管理域

### 80. `pages` — 静态页面（CMS）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| ~~基础字段~~ | — | — | 见模板 |
| id | BIGINT | PK | 雪花ID |
| code | VARCHAR(50) | UNIQUE | 页面标识 |
| name | VARCHAR(255) | NOT NULL | |
| title | VARCHAR(255) | NOT NULL | SEO标题 |
| slug | VARCHAR(255) | NOT NULL, UNIQUE | URL标识 |
| content_html | TEXT | | |
| content_json | JSONB | | |
| description | VARCHAR(500) | | SEO描述 |
| keywords | VARCHAR(500) | | |
| template | VARCHAR(50) | DEFAULT 'default' | 模板 |
| layout_type | VARCHAR(30) | DEFAULT 'full' | |
| cover_image_url | TEXT | | |
| view_count | BIGINT | DEFAULT 0 | |
| is_published | BOOLEAN | DEFAULT false | |
| published_at | TIMESTAMPTZ | | |
| expire_at | TIMESTAMPTZ | | |
| status | VARCHAR(50) | DEFAULT 'draft' CHECK (draft/published/archived) | |
| deleted | BOOLEAN | DEFAULT false | |
| tenant_id | BIGINT | NOT NULL | |
| created_by | BIGINT | NOT NULL | |
| created_at | TIMESTAMPTZ | DEFAULT now() | |
| updated_by | BIGINT | NOT NULL | |
| updated_at | TIMESTAMPTZ | DEFAULT now() | |

### 81. `page_templates` — 页面模板

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| ~~基础字段~~ | — | — | 见模板 |
| id | BIGINT | PK | 雪花ID |
| code | VARCHAR(50) | UNIQUE | |
| name | VARCHAR(255) | NOT NULL | |
| description | TEXT | | |
| html_template | TEXT | | |
| css_content | TEXT | | |
| preview_image_url | TEXT | | |
| content_slots | JSONB | | 内容插槽 |
| status | VARCHAR(50) | DEFAULT 'active' | |
| deleted | BOOLEAN | DEFAULT false | |
| tenant_id | BIGINT | NOT NULL | |
| created_by | BIGINT | NOT NULL | |
| created_at | TIMESTAMPTZ | DEFAULT now() | |
| updated_by | BIGINT | NOT NULL | |
| updated_at | TIMESTAMPTZ | DEFAULT now() | |

### 82. `page_shares` — 页面分享

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| page_id | BIGINT | NOT NULL | |
| share_token | VARCHAR(128) | NOT NULL, UNIQUE | |
| share_type | VARCHAR(20) | DEFAULT 'public' CHECK (public/password/restricted) | |
| password_hash | VARCHAR(128) | | |
| expire_at | TIMESTAMPTZ | | |
| access_count | INT | DEFAULT 0 | |
| max_access_count | INT | | |
| is_active | BOOLEAN | DEFAULT true | |
| last_accessed_at | TIMESTAMPTZ | | |
| created_by | BIGINT | NOT NULL | |
| created_at | TIMESTAMPTZ | DEFAULT now() | |
| tenant_id | BIGINT | NOT NULL | |

---

## 二十二、组织与项目扩展域（V1.1 新增）

> 对应 PRD 8.1 工作空间管理（成员四级角色、多语言/时区）与 8.2 项目管理（项目模板、项目集）。

### 83. `tenant_members` — 租户成员（空间成员身份）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| tenant_id | BIGINT | NOT NULL | |
| user_id | BIGINT | NOT NULL, 逻辑外键 → users.id | |
| role | VARCHAR(20) | NOT NULL DEFAULT 'member' CHECK (owner/admin/member/guest) | 空间四级角色（PRD 1.4.2/1.4.3） |
| join_type | VARCHAR(20) | DEFAULT 'invitation' CHECK (direct/invitation/import) | 加入方式（import=CSV 批量导入） |
| invited_by | BIGINT | 逻辑外键 → users.id | 邀请人 |
| is_active | BOOLEAN | DEFAULT true | 移除成员置 false（已关联需求/任务/缺陷保留但无权限） |
| joined_at | TIMESTAMPTZ | DEFAULT now() | |
| last_active_at | TIMESTAMPTZ | | 最后活跃时间（成员列表展示） |
| created_at | TIMESTAMPTZ | DEFAULT now() | |
| updated_at | TIMESTAMPTZ | DEFAULT now() | |

> UNIQUE(tenant_id, user_id)
> 说明：与 RBAC 的关系——`tenant_members.role` 管空间成员身份与宏观权限档（Owner/Admin/Member/Guest），`user_roles`+`role_menus` 管细粒度权限点；四级角色映射到系统内置 roles，二者并存互补。

### 84. `user_preferences` — 用户偏好

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| user_id | BIGINT | NOT NULL, 逻辑外键 → users.id | |
| language | VARCHAR(20) | DEFAULT 'zh-CN' | 界面语言 |
| timezone | VARCHAR(50) | DEFAULT 'Asia/Shanghai' | 个人时区 |
| theme | VARCHAR(20) | DEFAULT 'light' CHECK (light/dark/system) | 主题 |
| default_project_id | BIGINT | 逻辑外键 → projects.id | 默认项目（工作台上下文自动带入） |
| home_page | VARCHAR(50) | DEFAULT 'workbench' | 登录后首页 |
| extra | JSONB | | 其他偏好（快捷键/列表密度等） |
| tenant_id | BIGINT | NOT NULL | |
| created_at | TIMESTAMPTZ | DEFAULT now() | |
| updated_at | TIMESTAMPTZ | DEFAULT now() | |

> UNIQUE(tenant_id, user_id)

### 85. `project_templates` — 项目模板

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| ~~基础字段~~ | — | — | 见模板 |
| id | BIGINT | PK | 雪花ID |
| code | VARCHAR(50) | 租户内唯一 | |
| name | VARCHAR(255) | NOT NULL | 模板名（敏捷/瀑布/通用，PRD 2.4.1） |
| description | TEXT | | |
| project_type | VARCHAR(50) | DEFAULT 'scrum' CHECK (scrum/kanban/waterfall) | |
| thumbnail_url | TEXT | | |
| workflow_preset | JSONB | | 预设状态机（states/state_transitions 种子数据） |
| modules_preset | JSONB | | 预设功能模块开关（intake/sprint/version/estimate） |
| config_preset | JSONB | | 其他预设（估算体系/标签/角色） |
| is_builtin | BOOLEAN | DEFAULT false | 内置模板 |
| usage_count | INT | DEFAULT 0 | |
| status | VARCHAR(50) | DEFAULT 'active' | |
| deleted | BOOLEAN | DEFAULT false | |
| tenant_id | BIGINT | NOT NULL | |
| created_by | BIGINT | NOT NULL | |
| created_at | TIMESTAMPTZ | DEFAULT now() | |
| updated_by | BIGINT | NOT NULL | |
| updated_at | TIMESTAMPTZ | DEFAULT now() | |

### 86. `project_groups` — 项目集 / 项目分组

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| ~~基础字段~~ | — | — | 见模板 |
| id | BIGINT | PK | 雪花ID |
| code | VARCHAR(50) | | |
| name | VARCHAR(255) | NOT NULL | 项目集名称（PRD 2.2 项目分组/分类） |
| description | TEXT | | |
| owner_id | BIGINT | 逻辑外键 → users.id | PMO 负责人 |
| sort_order | DOUBLE PRECISION | DEFAULT 65535 | |
| status | VARCHAR(50) | DEFAULT 'active' | |
| deleted | BOOLEAN | DEFAULT false | |
| tenant_id | BIGINT | NOT NULL | |
| created_by | BIGINT | NOT NULL | |
| created_at | TIMESTAMPTZ | DEFAULT now() | |
| updated_by | BIGINT | NOT NULL | |
| updated_at | TIMESTAMPTZ | DEFAULT now() | |

---

## 二十三、评审工作流域（V1.1 新增）

> 对应 PRD 5.3.3 需求评审工作流、10.3.2 文档评审、10.4 知识库评审、4.4.3 迭代复盘评审人。统一多态评审模型，避免每种实体各建一套审批表。

### 87. `review_templates` — 评审模板

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| ~~基础字段~~ | — | — | 见模板 |
| id | BIGINT | PK | 雪花ID |
| code | VARCHAR(50) | | |
| name | VARCHAR(255) | NOT NULL | 模板名 |
| entity_scope | VARCHAR(30) | NOT NULL CHECK (requirement/document/knowledge_page/sprint) | 适用对象 |
| dimensions | JSONB | NOT NULL | 评审维度 [{name, max_score, weight, required}]（PRD：评审模板可自定义评审维度） |
| pass_rule | JSONB | | 通过规则 {min_score, min_approvals, require_all} |
| is_builtin | BOOLEAN | DEFAULT false | |
| status | VARCHAR(50) | DEFAULT 'active' | |
| deleted | BOOLEAN | DEFAULT false | |
| tenant_id | BIGINT | NOT NULL | |
| created_by | BIGINT | NOT NULL | |
| created_at | TIMESTAMPTZ | DEFAULT now() | |
| updated_by | BIGINT | NOT NULL | |
| updated_at | TIMESTAMPTZ | DEFAULT now() | |

### 88. `reviews` — 评审单（多态）

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
| status | VARCHAR(20) | DEFAULT 'pending' CHECK (pending/approved/rejected/cancelled) | |
| score_summary | JSONB | | 评分汇总 {avg, per_dimension} |
| round | INT | DEFAULT 1 | 评审轮次（驳回后重提递增） |
| submitted_at | TIMESTAMPTZ | DEFAULT now() | |
| completed_at | TIMESTAMPTZ | | |
| deleted | BOOLEAN | DEFAULT false | |
| tenant_id | BIGINT | NOT NULL | |
| created_by | BIGINT | NOT NULL | |
| created_at | TIMESTAMPTZ | DEFAULT now() | |
| updated_by | BIGINT | NOT NULL | |
| updated_at | TIMESTAMPTZ | DEFAULT now() | |

> INDEX (tenant_id, entity_type, entity_id, status)

### 89. `review_assignments` — 评审人记录

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| review_id | BIGINT | NOT NULL, 逻辑外键 → reviews.id | |
| reviewer_id | BIGINT | NOT NULL, 逻辑外键 → users.id | 评审人（支持多人评审） |
| status | VARCHAR(20) | DEFAULT 'pending' CHECK (pending/approved/rejected/abstained) | |
| score | NUMERIC(5,2) | | 总分 |
| dimension_scores | JSONB | | 分维度打分 [{dimension, score, comment}] |
| comment | TEXT | | 评审意见 |
| reviewed_at | TIMESTAMPTZ | | |
| tenant_id | BIGINT | NOT NULL | |
| created_at | TIMESTAMPTZ | DEFAULT now() | |

> UNIQUE(tenant_id, review_id, reviewer_id, round 由 reviews 侧控制)

---

## 二十四、版本日报告域（V1.1 新增）

> 对应 PRD 3.4.2：发布时自动生成交付报告与 Release Notes（缺陷数、通过率、准出率、7日/30日留存）。

### 90. `version_reports` — 版本日交付报告

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| version_id | BIGINT | NOT NULL, 逻辑外键 → versions.id | |
| report_type | VARCHAR(30) | NOT NULL CHECK (delivery_report/release_notes) | |
| content_md | TEXT | | 渲染内容（Release Notes 可按模板定制） |
| metrics | JSONB | | 结构化指标 {defect_count, pass_rate, exit_rate, retention_7d, retention_30d, ...} |
| generated_by | BIGINT | 逻辑外键 → users.id；NULL=系统生成 | |
| generated_at | TIMESTAMPTZ | DEFAULT now() | |
| tenant_id | BIGINT | NOT NULL | |
| created_at | TIMESTAMPTZ | DEFAULT now() | |

> UNIQUE(tenant_id, version_id, report_type, generated_at::date) — 同日同类型不重复生成

---

## 二十五、模板、权限与分享域（V1.1 新增）

> 对应 PRD 5.1 需求模板、7.1 缺陷模板库、10.3.2 文档模板、10.4 知识库模板（技术方案/ADR/PRD）与文档权限（空间级四级+文档级覆盖）；分享对应 10.3.1 文档分享、10.5 公开链接（密码+有效期）。

### 91. `knowledge_space_members` — 知识库空间成员

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| space_id | BIGINT | NOT NULL, 逻辑外键 → knowledge_spaces.id | |
| user_id | BIGINT | NOT NULL, 逻辑外键 → users.id | |
| role | VARCHAR(20) | NOT NULL DEFAULT 'viewer' CHECK (owner/admin/editor/viewer) | 空间级四级权限（PRD 10.4 文档权限 P0） |
| page_overrides | JSONB | | 文档级权限覆盖 [{page_id, role}]（继承或覆盖） |
| tenant_id | BIGINT | NOT NULL | |
| created_by | BIGINT | NOT NULL | |
| created_at | TIMESTAMPTZ | DEFAULT now() | |

> UNIQUE(tenant_id, space_id, user_id)

### 92. `content_templates` — 内容模板（工作项/文档/知识库统一）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| ~~基础字段~~ | — | — | 见模板 |
| id | BIGINT | PK | 雪花ID |
| code | VARCHAR(50) | | |
| name | VARCHAR(255) | NOT NULL | 模板名 |
| description | TEXT | | |
| template_scope | VARCHAR(20) | NOT NULL CHECK (work_item/document/knowledge_page) | 模板域 |
| entity_type | VARCHAR(20) | CHECK (requirement/task/defect) | work_item 域时指定类型 |
| content | JSONB | NOT NULL | 模板内容（预填字段值/富文本骨架） |
| project_id | BIGINT | 逻辑外键 → projects.id | 项目级模板；NULL = 租户级 |
| is_builtin | BOOLEAN | DEFAULT false | |
| usage_count | INT | DEFAULT 0 | |
| status | VARCHAR(50) | DEFAULT 'active' | |
| deleted | BOOLEAN | DEFAULT false | |
| tenant_id | BIGINT | NOT NULL | |
| created_by | BIGINT | NOT NULL | |
| created_at | TIMESTAMPTZ | DEFAULT now() | |
| updated_by | BIGINT | NOT NULL | |
| updated_at | TIMESTAMPTZ | DEFAULT now() | |

### 93. `share_links` — 资源分享链接（多态）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| resource_type | VARCHAR(30) | NOT NULL CHECK (document/knowledge_page/page/dashboard/saved_view/version_report) | 被分享资源类型 |
| resource_id | BIGINT | NOT NULL | |
| share_token | VARCHAR(128) | NOT NULL, UNIQUE | 访问令牌 |
| share_type | VARCHAR(20) | DEFAULT 'public' CHECK (public/password/restricted) | restricted=指定成员可见 |
| allowed_user_ids | JSONB | | restricted 时的用户白名单 |
| password_hash | VARCHAR(128) | | 访问密码 |
| expire_at | TIMESTAMPTZ | | 有效期（PRD：支持设置链接有效期） |
| access_count | INT | DEFAULT 0 | |
| max_access_count | INT | | |
| is_active | BOOLEAN | DEFAULT true | |
| last_accessed_at | TIMESTAMPTZ | | |
| created_by | BIGINT | NOT NULL | |
| created_at | TIMESTAMPTZ | DEFAULT now() | |
| tenant_id | BIGINT | NOT NULL | |

> 说明：CMS 的 `page_shares` 保留（含页面特有统计语义）；文档/知识库/仪表盘/视图等分享统一走本表。

---

## 二十六、个人效率扩展域（V1.1 新增）

> 对应 PRD 9.2 个人工作台（TodoItem 置顶/排序）、9.5 保存/分享视图（个人/团队/系统）、11.1 星标、9.2 今日日程。

### 94. `workbench_todos` — 我的待办

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| user_id | BIGINT | NOT NULL, 逻辑外键 → users.id | |
| entity_type | VARCHAR(20) | NOT NULL CHECK (requirement/task/defect) | |
| entity_id | BIGINT | NOT NULL | |
| project_id | BIGINT | | 冗余便于跨项目分组 |
| is_pinned | BOOLEAN | DEFAULT false | 置顶 |
| sort_order | DOUBLE PRECISION | DEFAULT 65535 | 手动拖拽排序（PRD 9.2 拖拽调整优先级顺序） |
| note | VARCHAR(500) | | 个人备注 |
| added_at | TIMESTAMPTZ | DEFAULT now() | |
| tenant_id | BIGINT | NOT NULL | |

> UNIQUE(tenant_id, user_id, entity_type, entity_id)

### 95. `saved_views` — 保存视图（个人/团队/系统）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| ~~基础字段~~ | — | — | 见模板 |
| id | BIGINT | PK | 雪花ID |
| code | VARCHAR(50) | | |
| name | VARCHAR(255) | NOT NULL | 视图名（如「PMO 视图」「QA 视图」） |
| project_id | BIGINT | NOT NULL, 逻辑外键 → projects.id | |
| view_type | VARCHAR(30) | NOT NULL CHECK (kanban/list/calendar/gantt/timeline/table/spreadsheet) | |
| entity_scope | VARCHAR(20) | DEFAULT 'all' CHECK (all/requirement/task/defect) | |
| scope | VARCHAR(20) | DEFAULT 'personal' CHECK (personal/team/system) | PRD 9.5：个人/团队/系统默认视图 |
| owner_id | BIGINT | NOT NULL, 逻辑外键 → users.id | |
| filters | JSONB | | |
| sort_config | JSONB | | |
| columns_config | JSONB | | |
| group_by | VARCHAR(50) | | |
| is_default | BOOLEAN | DEFAULT false | 项目默认打开视图 |
| share_token | VARCHAR(64) | UNIQUE | 链接分享（PRD 9.5 分享视图） |
| status | VARCHAR(50) | DEFAULT 'active' | |
| deleted | BOOLEAN | DEFAULT false | |
| tenant_id | BIGINT | NOT NULL | |
| created_by | BIGINT | NOT NULL | |
| created_at | TIMESTAMPTZ | DEFAULT now() | |
| updated_by | BIGINT | NOT NULL | |
| updated_at | TIMESTAMPTZ | DEFAULT now() | |

> 说明：与 `view_preferences`（一人一项目一视图的偏好记忆）互补——`saved_views` 是可命名、可共享、多份的视图定义。

### 96. `favorite_items` — 星标收藏（多态）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| user_id | BIGINT | NOT NULL | |
| entity_type | VARCHAR(30) | NOT NULL CHECK (requirement/task/defect/project/document/knowledge_page/sprint/version) | |
| entity_id | BIGINT | NOT NULL | |
| sort_order | DOUBLE PRECISION | DEFAULT 65535 | |
| tenant_id | BIGINT | NOT NULL | |
| created_at | TIMESTAMPTZ | DEFAULT now() | |

> UNIQUE(tenant_id, user_id, entity_type, entity_id)

### 97. `calendar_events` — 日程 / 会议

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| ~~基础字段~~ | — | — | 见模板 |
| id | BIGINT | PK | 雪花ID |
| code | VARCHAR(50) | | |
| name | VARCHAR(255) | NOT NULL | 标题 |
| event_type | VARCHAR(20) | NOT NULL CHECK (meeting/standup/review/personal/release) | 会议/站会/评审/个人日程/发布 |
| description | TEXT | | |
| project_id | BIGINT | 逻辑外键 → projects.id | |
| sprint_id | BIGINT | 逻辑外键 → sprints.id | 站会关联迭代 |
| entity_type | VARCHAR(20) | | 关联对象（如评审会关联 reviews） |
| entity_id | BIGINT | | |
| start_at | TIMESTAMPTZ | NOT NULL | |
| end_at | TIMESTAMPTZ | NOT NULL | |
| rrule | VARCHAR(255) | | 周期性规则（RFC 5545，如每日站会） |
| organizer_id | BIGINT | NOT NULL, 逻辑外键 → users.id | |
| attendees | JSONB | | 参会人 [{user_id, rsvp}] |
| meeting_url | TEXT | | 会议链接 |
| status | VARCHAR(50) | DEFAULT 'active' | |
| deleted | BOOLEAN | DEFAULT false | |
| tenant_id | BIGINT | NOT NULL | |
| created_by | BIGINT | NOT NULL | |
| created_at | TIMESTAMPTZ | DEFAULT now() | |
| updated_by | BIGINT | NOT NULL | |
| updated_at | TIMESTAMPTZ | DEFAULT now() | |

---

## 二十七、工程联动与数据任务域（V1.1 新增）

> 对应 PRD 6.1 关联代码提交（智能提交）、11.3 关联修复 PR、9.1 WidgetData 缓存、9.5 导入导出（字段映射/增量导入）、1.2 数据归档/导出。

### 98. `code_links` — 代码关联（commit / branch / PR）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| entity_type | VARCHAR(20) | NOT NULL CHECK (requirement/task/defect) | |
| entity_id | BIGINT | NOT NULL | |
| link_type | VARCHAR(20) | NOT NULL CHECK (commit/branch/pull_request/tag) | |
| provider | VARCHAR(20) | NOT NULL CHECK (github/gitlab/gitee/cnb/other) | 代码平台 |
| repo_name | VARCHAR(255) | | 仓库（org/repo） |
| external_id | VARCHAR(255) | NOT NULL | commit_sha / branch 名 / MR IID |
| title | VARCHAR(500) | | 提交信息/PR 标题 |
| url | TEXT | | 跳转链接 |
| author_name | VARCHAR(255) | | 提交人 |
| status | VARCHAR(30) | | PR 状态（open/merged/closed） |
| committed_at | TIMESTAMPTZ | | |
| raw | JSONB | | 原始 payload |
| tenant_id | BIGINT | NOT NULL | |
| created_at | TIMESTAMPTZ | DEFAULT now() | |

> UNIQUE(tenant_id, entity_type, entity_id, link_type, provider, external_id) — 智能提交幂等
> INDEX (tenant_id, provider, external_id) — 按 commit/PR 反查关联

### 99. `dashboard_widget_data` — 小部件数据缓存

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| widget_id | BIGINT | NOT NULL, 逻辑外键 → dashboard_widgets.id | |
| query_hash | VARCHAR(64) | NOT NULL | 查询条件哈希（含全局时间筛选上下文） |
| data | JSONB | NOT NULL | 渲染数据 |
| refreshed_at | TIMESTAMPTZ | DEFAULT now() | |
| expires_at | TIMESTAMPTZ | | 按 widget.refresh_interval 计算 |
| tenant_id | BIGINT | NOT NULL | |

> UNIQUE(tenant_id, widget_id, query_hash)
> 说明：按 PRD 9.1 卡片刷新频率（实时/1h/4h/每日）分级缓存，失效后异步重算。

### 100. `data_jobs` — 数据导入 / 导出 / 归档任务

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| job_type | VARCHAR(20) | NOT NULL CHECK (import/export/archive) | |
| resource_type | VARCHAR(30) | NOT NULL | work_item/member/document/defect/... |
| project_id | BIGINT | 逻辑外键 → projects.id | |
| file_name | VARCHAR(255) | | 原始文件名 |
| file_path | TEXT | | 文件存储路径（结果文件或上传文件） |
| file_format | VARCHAR(10) | CHECK (csv/xlsx/json/md/html/pdf/docx) | |
| field_mapping | JSONB | | 字段映射配置（PRD 9.5 外部字段→系统字段） |
| dedup_strategy | JSONB | | 增量导入策略 {match_by: external_id/code, on_conflict: update/skip} |
| status | VARCHAR(20) | DEFAULT 'pending' CHECK (pending/running/success/partial/failed/cancelled) | |
| total_count | INT | | |
| success_count | INT | | |
| fail_count | INT | | |
| skip_count | INT | | |
| result_detail | JSONB | | 逐行错误明细 |
| error_message | TEXT | | |
| started_at | TIMESTAMPTZ | | |
| finished_at | TIMESTAMPTZ | | |
| expires_at | TIMESTAMPTZ | | 结果文件保留期 |
| tenant_id | BIGINT | NOT NULL | |
| created_by | BIGINT | NOT NULL | |
| created_at | TIMESTAMPTZ | DEFAULT now() | |

---

## 分区策略汇总

| 表 | 分区方式 | 保留策略 |
|----|----------|----------|
| entity_activities | 按月 `created_at` RANGE | 12个月后归档冷备 |
| notifications | 按月 `created_at` RANGE | 已读90天后自动清理 |
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
| 主键 | 全部100张 | 100 |
| 唯一约束 | code字段 + 联合唯一 | ~75 |
| 逻辑外键索引 | tenant_id + xxx_id | ~150 |
| 复合查询索引 | list视图排序 | ~50 |
| GIN索引 | 全文检索（降级） | ~5 |

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
| `GIN` 索引 | `BITMAP` / 全文索引 | `GIN` |

---

> 文档覆盖 **100 张表**：原有 82 张（租户权限6 + 项目5 + 状态机2 + 估算1 + 版本日1 + 迭代5 + 需求/任务/缺陷3 + 多态关联12 + 标签模块3 + 收件箱2 + 知识库4 + 文档管理3 + 通知4 + 效率增强10 + 自动化3 + 集成6 + 事件总线4 + 安全审计4 + 效能度量3 + 治理3 + CMS3）+ V1.1 新增 18 张（组织与项目扩展4 + 评审工作流3 + 版本日报告1 + 模板权限与分享3 + 个人效率扩展4 + 工程联动与数据任务3）。
