# Ydsz Plane 数据库表设计

> 基于产品需求文档（PRD V1.0）+ 架构设计文档 + 用户确认的设计约束  
> 数据库：PostgreSQL 16+（兼容达梦/人大金仓）  
> 最后更新：2026-08-10（V2.0 审查修订版）

## 修订记录

| 版本 | 日期 | 变更说明 |
|------|------|----------|
| V1.0 | 2026-08-08 | 初始版本 133 张表 |
| V2.0 | 2026-08-10 | 审查修订：ENUM 状态、租户分离、字段排序标准化 |

## 全局设计约定

### 字段规范

| 约定 | 规格 |
|------|------|
| 主键 | `id BIGINT PRIMARY KEY` — 雪花算法（美团 Leaf/Snowflake），递增有序，应用层生成 |
| 短代码 | `code VARCHAR(50)` — 用户可选按规则生成的业务标识符 |
| 名称 | `name VARCHAR(255) NOT NULL` — 业务名称 |
| 状态 | `status <ENUM_TYPE> NOT NULL` — 状态（使用 PostgreSQL ENUM 类型） |
| 租户隔离 | `tenant_id BIGINT NOT NULL` — 业务表必备，数据隔离唯一维度 |
| 软删除 | `deleted BOOLEAN NOT NULL DEFAULT false` |
| 创建人/更新人 | `created_by/updated_by BIGINT NOT NULL` — 逻辑外键 |
| 时间戳 | `created_at/updated_at TIMESTAMPTZ NOT NULL DEFAULT now()` |
| 无物理外键 | 所有关联通过代码层维护，逻辑外键字段加索引 |

### 字段排序规则

所有业务表统一按以下顺序排列字段：

```
1. id              BIGINT PRIMARY KEY
2. code            VARCHAR(50)          -- 如有
3. name            VARCHAR(255) NOT NULL -- 如有
~~ 业务专用字段（按业务逻辑排序）~~
n-8. status        <ENUM>               -- 状态
n-7. deleted       BOOLEAN DEFAULT false
n-6. tenant_id     BIGINT NOT NULL      -- 租户ID（系统级表可省略）
n-5. created_by    BIGINT NOT NULL
n-4. created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
n-3. updated_by    BIGINT NOT NULL
n-2. updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
```

### ENUM 状态类型定义

```sql
-- 租户状态
CREATE TYPE tenant_status AS ENUM ('active', 'disabled', 'archived', 'expired');

-- 用户状态
CREATE TYPE user_status AS ENUM ('active', 'inactive', 'locked');

-- 通用实体状态
CREATE TYPE entity_status AS ENUM ('active', 'inactive', 'archived');

-- 项目状态
CREATE TYPE project_status AS ENUM ('active', 'archived');

-- 需求/任务/缺陷状态（通过 states 表动态定义，此处为默认值）
CREATE TYPE work_item_status AS ENUM ('draft', 'active', 'completed', 'cancelled');

-- 迭代状态
CREATE TYPE sprint_status AS ENUM ('planned', 'active', 'completed');

-- 版本状态
CREATE TYPE version_status AS ENUM ('planning', 'active', 'released', 'archived');

-- 入口工单状态
CREATE TYPE intake_issue_status AS ENUM ('open', 'accepted', 'rejected', 'archived');

-- 通知状态
CREATE TYPE notification_status AS ENUM ('pending', 'sent', 'failed', 'read');

-- 工作空间角色
CREATE TYPE workspace_role_enum AS ENUM ('owner', 'admin', 'member', 'guest');

-- 项目角色
CREATE TYPE project_role_enum AS ENUM ('admin', 'member');
```

### 三层次数据隔离模型

```
租户（tenants）
  └── 工作空间（workspaces）
        └── 项目（projects）
              └── 工作项（task / requirement / defect）
```

- **租户（tenants）**：组织级实体，代表公司/部门/团队
- **工作空间（workspaces）**：租户下的协作空间，包含多个项目
- **项目（projects）**：工作空间下的具体项目管理单元

### 系统级表（无需 tenant_id）

以下表为系统全局配置，不按租户隔离：

- `menus` — 菜单/权限资源（系统级）
- `roles` — 角色（系统预置 + 租户自定义）
- `automation_templates` — 自动化模板（系统预设）
- `dashboard_templates` — 仪表盘模板（系统预设）
- `sso_providers` — SSO 提供方配置

### 逻辑外键索引惯例

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

- WBS 三级层级：`parent_id` 自引用，应用层限制最大深度 3 级并做循环校验
- 同级排序：`sort_order DOUBLE PRECISION DEFAULT 65535`（小数插值重排，避免批量 UPDATE）
- 知识文档树、项目文档目录同样使用 `parent_id` + `sort_order`

---

## ER 总览

```
tenants ──< users ──< user_roles >── roles ──< role_menus >── menus
  │           ├──< tenant_members
  │           └──< user_preferences
  │
  ├──< workspaces ──< workspace_members
  │        └──< projects ──< project_members / project_configs / project_sequences
  │              ├──< states / state_transitions
  │              ├──< estimate_points
  │              ├──< requirement / task / defect
  │              │        ├──< xxx_assignees / xxx_labels / xxx_watchers
  │              │        ├──< xxx_comments / xxx_activities / xxx_attachments
  │              │        ├──< xxx_timelogs / xxx_modules / xxx_relations
  │              │        └──< task_dependencies / biz_entity_relations
  │              ├──< modules / labels
  │              ├──< versions ──< version_sprint_relations >── sprints
  │              │        ├──< version_delivery_snapshots
  │              │        └──< sprint_requirements / sprint_tasks / sprint_defects
  │              ├──< intake_channels ──< intake_issues
  │              └──< content_templates
  │
  ├──< reviews ──< review_assignments
  ├──< knowledge_spaces ──< knowledge_pages
  ├──< documents ──< document_versions
  ├──< share_links
  ├──< notifications ──< notification_deliveries / notification_subscriptions
  ├──< dashboards ──< dashboard_widgets
  ├──< workbench_configs
  ├──< recent_items / view_preferences / saved_views
  ├──< search_history / search_bookmarks / search_documents
  ├──< calendar_events
  ├──< automation_rules ──< rule_executions
  ├──< webhooks ──< webhook_logs
  ├──< biz_entity_relations
  ├──< data_jobs
  └──< audit_logs / domain_events
```

---

## 1、租户与权限域

### 1. `tenants` — 租户（组织机构）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| code | VARCHAR(50) | UNIQUE | 租户编码（如 MEITUAN） |
| name | VARCHAR(255) | NOT NULL | 组织名称（如"美团基础研发平台"） |
| slug | VARCHAR(100) | NOT NULL, UNIQUE | URL 标识 |
| logo_url | TEXT | | 组织 Logo |
| owner_id | BIGINT | NOT NULL, 逻辑外键 → users.id | 租户 Owner |
| timezone | VARCHAR(50) | DEFAULT 'Asia/Shanghai' | 默认时区 |
| language | VARCHAR(20) | DEFAULT 'zh-CN' | 默认语言 |
| brand_config | JSONB | | 品牌定制（主题色/登录页） |
| status | tenant_status | NOT NULL DEFAULT 'active' | active/disabled/archived/expired |
| max_projects | INT | DEFAULT 10 | 最大项目数 |
| max_users | INT | DEFAULT 50 | 最大用户数 |
| max_workspaces | INT | DEFAULT 5 | 最大工作空间数 |
| expired_at | TIMESTAMPTZ | | 服务到期时间 |
| config | JSONB | | 租户级配置 |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

> 说明：tenants 为组织级根表，不含 deleted/tenant_id/created_by/updated_by。一个租户可包含多个工作空间。

### 2. `workspaces` — 工作空间

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| code | VARCHAR(50) | 租户内唯一 | 工作空间编码 |
| name | VARCHAR(255) | NOT NULL | 工作空间名称 |
| slug | VARCHAR(100) | NOT NULL, 租户内唯一 | URL 友好标识 |
| tenant_id | BIGINT | NOT NULL, 逻辑外键 → tenants.id | 所属租户 |
| logo_url | TEXT | | Logo |
| owner_id | BIGINT | NOT NULL | 空间 Owner |
| timezone | VARCHAR(50) | DEFAULT 'Asia/Shanghai' | 时区 |
| language | VARCHAR(20) | DEFAULT 'zh-CN' | 语言 |
| status | entity_status | NOT NULL DEFAULT 'active' | 状态 |
| max_projects | INT | DEFAULT 20 | 空间级项目上限 |
| config | JSONB | | 空间级配置 |
| deleted | BOOLEAN | DEFAULT false | 软删除 |
| created_by | BIGINT | NOT NULL | 创建人 |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_by | BIGINT | NOT NULL | 更新人 |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

> 说明：workspaces 是租户下的协作空间，一个工作空间包含多个项目。它与 tenants 是 N:1 关系。

### 3. `workspace_members` — 工作空间成员

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| tenant_id | BIGINT | NOT NULL | 租户ID |
| workspace_id | BIGINT | NOT NULL | 工作空间ID |
| user_id | BIGINT | NOT NULL | 用户ID |
| role | workspace_role_enum | NOT NULL DEFAULT 'member' | owner/admin/member/guest |
| status | entity_status | NOT NULL DEFAULT 'active' | 状态 |
| joined_at | TIMESTAMPTZ | DEFAULT now() | 加入时间 |
| deleted | BOOLEAN | DEFAULT false | 软删除 |
| created_by | BIGINT | NOT NULL | 创建人 |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_by | BIGINT | NOT NULL | 更新人 |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

> UNIQUE(tenant_id, workspace_id, user_id) WHERE NOT deleted

### 4. `users` — 用户

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| code | VARCHAR(50) | | 工号/用户编码 |
| name | VARCHAR(255) | NOT NULL | 显示名 |
| email | VARCHAR(255) | NOT NULL | 邮箱 |
| phone | VARCHAR(20) | | 手机号 |
| password_hash | TEXT | | 密码哈希 |
| avatar_url | TEXT | | 头像URL |
| status | user_status | NOT NULL DEFAULT 'active' | 状态 |
| is_super_admin | BOOLEAN | DEFAULT false | 系统级超管 |
| last_login_at | TIMESTAMPTZ | | 最后登录时间 |
| tenant_id | BIGINT | NOT NULL | 归属租户 |
| deleted | BOOLEAN | DEFAULT false | 软删除 |
| created_by | BIGINT | NOT NULL | 创建人 |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_by | BIGINT | NOT NULL | 更新人 |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

> UNIQUE(tenant_id, email) WHERE NOT deleted

### 5. `roles` — 角色

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| code | VARCHAR(50) | 租户内唯一 | 角色编码 |
| name | VARCHAR(255) | NOT NULL | 角色名称 |
| description | TEXT | | 描述 |
| status | entity_status | NOT NULL DEFAULT 'active' | 状态 |
| is_system | BOOLEAN | DEFAULT false | 系统内置角色 |
| role_scope | VARCHAR(50) | CHECK (tenant/project) | 作用范围 |
| deleted | BOOLEAN | DEFAULT false | 软删除 |
| created_by | BIGINT | NOT NULL | 创建人 |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_by | BIGINT | NOT NULL | 更新人 |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

> 说明：roles 为系统级表，不含 tenant_id。通过 role_menus 与角色关联。

### 6. `menus` — 菜单 / 权限资源

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| code | VARCHAR(100) | NOT NULL, UNIQUE | 权限编码 |
| name | VARCHAR(255) | NOT NULL | 菜单/按钮名称 |
| menu_type | VARCHAR(20) | NOT NULL CHECK (menu/button/api) | 资源类型 |
| parent_id | BIGINT | | 父菜单 |
| path | VARCHAR(255) | | 路由路径 |
| icon | VARCHAR(100) | | 图标 |
| sort_order | INT | DEFAULT 0 | 排序 |
| status | entity_status | DEFAULT 'active' | 状态 |
| deleted | BOOLEAN | DEFAULT false | 软删除 |
| created_by | BIGINT | NOT NULL | 创建人 |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_by | BIGINT | NOT NULL | 更新人 |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

> 说明：menus 为系统级资源表，不含 tenant_id。

### 7. `user_roles` — 用户角色关联

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| tenant_id | BIGINT | NOT NULL | 租户ID |
| user_id | BIGINT | NOT NULL | 用户ID |
| role_id | BIGINT | NOT NULL | 角色ID |
| status | entity_status | NOT NULL DEFAULT 'active' | 状态 |
| deleted | BOOLEAN | DEFAULT false | 软删除 |
| created_by | BIGINT | NOT NULL | 创建人 |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_by | BIGINT | NOT NULL | 更新人 |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |

---

## 2、项目域

### 8. `projects` — 项目

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| code | VARCHAR(50) | 工作空间内唯一 | 项目编码 |
| name | VARCHAR(255) | NOT NULL | 项目名称 |
| tenant_id | BIGINT | NOT NULL | 租户ID |
| workspace_id | BIGINT | NOT NULL | 工作空间ID |
| identifier | VARCHAR(20) | | 项目标识符（大写 2-10 字符） |
| slug | VARCHAR(100) | | URL 友好标识 |
| description | TEXT | | 项目描述 |
| icon | VARCHAR(50) | | 项目图标 |
| cover_image_url | TEXT | | 封面图片 |
| network | VARCHAR(20) | DEFAULT 'private' | 可见性 |
| template | VARCHAR(20) | DEFAULT 'generic' | 项目模板 |
| status | project_status | NOT NULL DEFAULT 'active' | 状态 |
| modules | JSONB | | 功能模块开关 |
| start_date | DATE | | 开始日期 |
| target_date | DATE | | 目标日期 |
| owner_id | BIGINT | NOT NULL | 项目负责人 |
| version | INT | DEFAULT 1 | 乐观锁版本号 |
| sort_order | DOUBLE PRECISION | DEFAULT 65535 | 排序权重 |
| deleted | BOOLEAN | DEFAULT false | 软删除 |
| created_by | BIGINT | NOT NULL | 创建人 |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_by | BIGINT | NOT NULL | 更新人 |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

### 9. `project_members` — 项目成员

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| tenant_id | BIGINT | NOT NULL | 租户ID |
| workspace_id | BIGINT | NOT NULL | 工作空间ID |
| project_id | BIGINT | NOT NULL | 项目ID |
| user_id | BIGINT | NOT NULL | 用户ID |
| role | project_role_enum | NOT NULL DEFAULT 'member' | 角色 |
| joined_at | TIMESTAMPTZ | DEFAULT now() | 加入时间 |
| status | entity_status | NOT NULL DEFAULT 'active' | 状态 |
| deleted | BOOLEAN | DEFAULT false | 软删除 |
| created_by | BIGINT | NOT NULL | 创建人 |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_by | BIGINT | NOT NULL | 更新人 |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

### 10. `project_sequences` — 项目序列发号器

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| tenant_id | BIGINT | NOT NULL | 租户ID |
| workspace_id | BIGINT | NOT NULL | 工作空间ID |
| project_id | BIGINT | NOT NULL, UNIQUE | 项目ID |
| sequence_id | BIGINT | NOT NULL DEFAULT 0 | 当前序号 |
| status | entity_status | NOT NULL DEFAULT 'active' | 状态 |
| deleted | BOOLEAN | DEFAULT false | 软删除 |
| created_by | BIGINT | NOT NULL | 创建人 |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_by | BIGINT | NOT NULL | 更新人 |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

---

## 3、工作项域

### 11. `task` — 任务

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| code | VARCHAR(50) | | 任务编码 |
| name | VARCHAR(255) | NOT NULL | 任务名称 |
| tenant_id | BIGINT | NOT NULL | 租户ID |
| workspace_id | BIGINT | NOT NULL | 工作空间ID |
| project_id | BIGINT | NOT NULL | 项目ID |
| sequence_id | BIGINT | NOT NULL | 序列号 |
| public_id | UUID | DEFAULT gen_random_uuid() | 公开ID |
| parent_id | BIGINT | | 父任务ID |
| depth | SMALLINT | DEFAULT 1 CHECK (1-3) | 层级深度 |
| description_json | JSONB | | 描述（JSON） |
| description_html | TEXT | | 描述（HTML） |
| description_stripped | TEXT | | 描述（纯文本） |
| state_id | BIGINT | NOT NULL | 状态ID |
| priority | VARCHAR(20) | DEFAULT 'none' | 优先级 |
| category | VARCHAR(20) | | 分类 |
| actual_effort | NUMERIC(8,2) | | 实际工时 |
| remaining_effort | NUMERIC(8,2) | | 剩余工时 |
| delay_reason | VARCHAR(50) | | 延期原因 |
| point | SMALLINT | CHECK (0-12) | 故事点 |
| estimate_point_id | BIGINT | | 估算ID |
| sprint_id | BIGINT | | 迭代ID |
| version_id | BIGINT | | 版本ID |
| progress | SMALLINT | DEFAULT 0 CHECK (0-100) | 进度 |
| start_date | DATE | | 开始日期 |
| target_date | DATE | | 目标日期 |
| completed_at | TIMESTAMPTZ | | 完成时间 |
| is_draft | BOOLEAN | DEFAULT false | 草稿标记 |
| archived_at | TIMESTAMPTZ | | 归档时间 |
| sort_order | DOUBLE PRECISION | DEFAULT 65535 | 排序权重 |
| version | INT | DEFAULT 1 | 乐观锁 |
| status | work_item_status | NOT NULL DEFAULT 'active' | 状态 |
| deleted | BOOLEAN | DEFAULT false | 软删除 |
| created_by | BIGINT | NOT NULL | 创建人 |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_by | BIGINT | NOT NULL | 更新人 |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

### 12. `requirement` — 需求

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| code | VARCHAR(50) | | 需求编码 |
| name | VARCHAR(255) | NOT NULL | 需求名称 |
| tenant_id | BIGINT | NOT NULL | 租户ID |
| workspace_id | BIGINT | NOT NULL | 工作空间ID |
| project_id | BIGINT | NOT NULL | 项目ID |
| sequence_id | BIGINT | NOT NULL | 序列号 |
| public_id | UUID | DEFAULT gen_random_uuid() | 公开ID |
| parent_id | BIGINT | | 父需求ID |
| depth | SMALLINT | DEFAULT 1 CHECK (1-3) | 层级深度 |
| description_json | JSONB | | 描述（JSON） |
| description_html | TEXT | | 描述（HTML） |
| description_stripped | TEXT | | 描述（纯文本） |
| state_id | BIGINT | NOT NULL | 状态ID |
| priority | VARCHAR(20) | DEFAULT 'none' | 优先级 |
| source | VARCHAR(50) | | 来源 |
| acceptance_criteria | JSONB | | 验收标准 |
| business_value | TEXT | | 业务价值 |
| review_status | VARCHAR(20) | | 评审状态 |
| point | SMALLINT | CHECK (0-12) | 故事点 |
| estimate_point_id | BIGINT | | 估算ID |
| sprint_id | BIGINT | | 迭代ID |
| version_id | BIGINT | | 版本ID |
| progress | SMALLINT | DEFAULT 0 CHECK (0-100) | 进度 |
| start_date | DATE | | 开始日期 |
| target_date | DATE | | 目标日期 |
| completed_at | TIMESTAMPTZ | | 完成时间 |
| is_draft | BOOLEAN | DEFAULT false | 草稿标记 |
| archived_at | TIMESTAMPTZ | | 归档时间 |
| sort_order | DOUBLE PRECISION | DEFAULT 65535 | 排序权重 |
| version | INT | DEFAULT 1 | 乐观锁 |
| status | work_item_status | NOT NULL DEFAULT 'active' | 状态 |
| deleted | BOOLEAN | DEFAULT false | 软删除 |
| created_by | BIGINT | NOT NULL | 创建人 |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_by | BIGINT | NOT NULL | 更新人 |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

### 13. `defect` — 缺陷

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| code | VARCHAR(50) | | 缺陷编码 |
| name | VARCHAR(255) | NOT NULL | 缺陷名称 |
| tenant_id | BIGINT | NOT NULL | 租户ID |
| workspace_id | BIGINT | NOT NULL | 工作空间ID |
| project_id | BIGINT | NOT NULL | 项目ID |
| sequence_id | BIGINT | NOT NULL | 序列号 |
| public_id | UUID | DEFAULT gen_random_uuid() | 公开ID |
| parent_id | BIGINT | | 父缺陷ID |
| depth | SMALLINT | DEFAULT 1 CHECK (1-3) | 层级深度 |
| description_json | JSONB | | 描述（JSON） |
| description_html | TEXT | | 描述（HTML） |
| description_stripped | TEXT | | 描述（纯文本） |
| state_id | BIGINT | NOT NULL | 状态ID |
| priority | VARCHAR(20) | DEFAULT 'none' | 优先级 |
| severity | SMALLINT | NOT NULL CHECK (1-5) | 严重度 |
| found_phase | VARCHAR(20) | | 发现阶段 |
| found_version_id | BIGINT | | 发现版本 |
| fix_version_id | BIGINT | | 修复版本 |
| root_cause_category | VARCHAR(50) | | 根因分类 |
| verifier_id | BIGINT | | 验证人 |
| environment | JSONB | | 环境信息 |
| reproduce_steps | JSONB | | 复现步骤 |
| fix_steps | JSONB | | 修复步骤 |
| regression_risk | VARCHAR(20) | | 回归风险 |
| point | SMALLINT | CHECK (0-12) | 故事点 |
| estimate_point_id | BIGINT | | 估算ID |
| sprint_id | BIGINT | | 迭代ID |
| version_id | BIGINT | | 版本ID |
| progress | SMALLINT | DEFAULT 0 CHECK (0-100) | 进度 |
| start_date | DATE | | 开始日期 |
| target_date | DATE | | 目标日期 |
| completed_at | TIMESTAMPTZ | | 完成时间 |
| is_draft | BOOLEAN | DEFAULT false | 草稿标记 |
| archived_at | TIMESTAMPTZ | | 归档时间 |
| sort_order | DOUBLE PRECISION | DEFAULT 65535 | 排序权重 |
| version | INT | DEFAULT 1 | 乐观锁 |
| status | work_item_status | NOT NULL DEFAULT 'active' | 状态 |
| deleted | BOOLEAN | DEFAULT false | 软删除 |
| created_by | BIGINT | NOT NULL | 创建人 |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_by | BIGINT | NOT NULL | 更新人 |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

---

## 4、工作项关联表

### 14. `task_assignees` — 任务执行人

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| tenant_id | BIGINT | NOT NULL | 租户ID |
| workspace_id | BIGINT | NOT NULL | 工作空间ID |
| project_id | BIGINT | NOT NULL | 项目ID |
| task_id | BIGINT | NOT NULL | 任务ID |
| user_id | BIGINT | NOT NULL | 用户ID |
| status | entity_status | NOT NULL DEFAULT 'active' | 状态 |
| deleted | BOOLEAN | DEFAULT false | 软删除 |
| created_by | BIGINT | NOT NULL | 创建人 |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_by | BIGINT | NOT NULL | 更新人 |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

### 15. `task_labels` — 任务标签

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| tenant_id | BIGINT | NOT NULL | 租户ID |
| workspace_id | BIGINT | NOT NULL | 工作空间ID |
| project_id | BIGINT | NOT NULL | 项目ID |
| task_id | BIGINT | NOT NULL | 任务ID |
| label_id | BIGINT | NOT NULL | 标签ID |
| status | entity_status | NOT NULL DEFAULT 'active' | 状态 |
| deleted | BOOLEAN | DEFAULT false | 软删除 |
| created_by | BIGINT | NOT NULL | 创建人 |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_by | BIGINT | NOT NULL | 更新人 |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

### 16. `task_modules` — 任务模块关联

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| tenant_id | BIGINT | NOT NULL | 租户ID |
| workspace_id | BIGINT | NOT NULL | 工作空间ID |
| project_id | BIGINT | NOT NULL | 项目ID |
| task_id | BIGINT | NOT NULL | 任务ID |
| module_id | BIGINT | NOT NULL | 模块ID |
| status | entity_status | NOT NULL DEFAULT 'active' | 状态 |
| deleted | BOOLEAN | DEFAULT false | 软删除 |
| created_by | BIGINT | NOT NULL | 创建人 |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_by | BIGINT | NOT NULL | 更新人 |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

### 17. `task_watchers` — 任务关注人

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| tenant_id | BIGINT | NOT NULL | 租户ID |
| workspace_id | BIGINT | NOT NULL | 工作空间ID |
| project_id | BIGINT | NOT NULL | 项目ID |
| task_id | BIGINT | NOT NULL | 任务ID |
| user_id | BIGINT | NOT NULL | 用户ID |
| status | entity_status | NOT NULL DEFAULT 'active' | 状态 |
| deleted | BOOLEAN | DEFAULT false | 软删除 |
| created_by | BIGINT | NOT NULL | 创建人 |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_by | BIGINT | NOT NULL | 更新人 |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

### 18. `task_relations` — 任务关联关系

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| tenant_id | BIGINT | NOT NULL | 租户ID |
| workspace_id | BIGINT | NOT NULL | 工作空间ID |
| project_id | BIGINT | NOT NULL | 项目ID |
| source_task_id | BIGINT | NOT NULL | 源任务 |
| target_task_id | BIGINT | NOT NULL | 目标任务 |
| relation_type | VARCHAR(50) | NOT NULL | 关联类型 |
| status | entity_status | NOT NULL DEFAULT 'active' | 状态 |
| deleted | BOOLEAN | DEFAULT false | 软删除 |
| created_by | BIGINT | NOT NULL | 创建人 |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_by | BIGINT | NOT NULL | 更新人 |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

### 19. `task_comments` — 任务评论

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| tenant_id | BIGINT | NOT NULL | 租户ID |
| workspace_id | BIGINT | NOT NULL | 工作空间ID |
| project_id | BIGINT | NOT NULL | 项目ID |
| task_id | BIGINT | NOT NULL | 任务ID |
| content_json | JSONB | NOT NULL | 内容（JSON） |
| content_html | TEXT | NOT NULL | 内容（HTML） |
| parent_id | BIGINT | | 父评论 |
| deleted | BOOLEAN | DEFAULT false | 软删除 |
| created_by | BIGINT | NOT NULL | 创建人 |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_by | BIGINT | | 更新人 |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

### 20. `task_activities` — 任务活动日志

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| tenant_id | BIGINT | NOT NULL | 租户ID |
| workspace_id | BIGINT | NOT NULL | 工作空间ID |
| project_id | BIGINT | NOT NULL | 项目ID |
| task_id | BIGINT | NOT NULL | 任务ID |
| verb | VARCHAR(50) | NOT NULL | 动作类型 |
| field_name | VARCHAR(100) | | 字段名 |
| old_value | TEXT | | 旧值 |
| new_value | TEXT | | 新值 |
| actor_id | BIGINT | NOT NULL | 操作人 |
| status | entity_status | NOT NULL DEFAULT 'active' | 状态 |
| deleted | BOOLEAN | DEFAULT false | 软删除 |
| created_by | BIGINT | NOT NULL | 创建人 |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间（分区键） |
| updated_by | BIGINT | NOT NULL | 更新人 |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

> 说明：按月分区表

---

> 需求（requirement_*）和缺陷（defect_*）的关联表结构与任务（task_*）一致，包含：requirement_assignees / requirement_labels / requirement_modules / requirement_watchers / requirement_relations / requirement_comments / requirement_activities / requirement_attachments / requirement_timelogs，以及 defect_assignees / defect_labels / defect_modules / defect_watchers / defect_relations / defect_comments / defect_activities / defect_attachments / defect_timelogs。每张表均包含字段：id、tenant_id、workspace_id、project_id、对应实体ID、业务字段、status、deleted、created_by、created_at、updated_by、updated_at。

---

## 5、任务/需求/缺陷扩展表

### 21. `task_ext` — 任务扩展字段

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| tenant_id | BIGINT | NOT NULL | 租户ID |
| workspace_id | BIGINT | NOT NULL | 工作空间ID |
| project_id | BIGINT | NOT NULL | 项目ID |
| task_id | BIGINT | NOT NULL | 任务ID |
| field_name | VARCHAR(100) | NOT NULL | 字段名 |
| field_value | JSONB | NOT NULL | 字段值 |
| field_schema | JSONB | | 字段 Schema |
| status | entity_status | NOT NULL DEFAULT 'active' | 状态 |
| deleted | BOOLEAN | DEFAULT false | 软删除 |
| created_by | BIGINT | NOT NULL | 创建人 |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_by | BIGINT | NOT NULL | 更新人 |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

> 说明：requirement_ext 和 defect_ext 结构相同。

---

## 6、迭代域

### 22. `sprints` — 迭代

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| code | VARCHAR(50) | | 迭代编码 |
| name | VARCHAR(255) | NOT NULL | 迭代名称 |
| tenant_id | BIGINT | NOT NULL | 租户ID |
| workspace_id | BIGINT | NOT NULL | 工作空间ID |
| project_id | BIGINT | NOT NULL | 项目ID |
| description | TEXT | | 描述 |
| goal | TEXT | | 迭代目标 |
| start_date | DATE | | 开始日期 |
| end_date | DATE | | 结束日期 |
| capacity | NUMERIC(10,2) | | 团队容量 |
| owner_id | BIGINT | | 负责人 |
| viewport | JSONB | DEFAULT '{}' | 视口配置 |
| review_snapshot | JSONB | | 复盘快照 |
| started_at | TIMESTAMPTZ | | 实际开始时间 |
| completed_at | TIMESTAMPTZ | | 实际完成时间 |
| version_id | BIGINT | | 关联版本 |
| status | sprint_status | NOT NULL DEFAULT 'planned' | planned/active/completed |
| version | INT | DEFAULT 1 | 乐观锁 |
| deleted | BOOLEAN | DEFAULT false | 软删除 |
| created_by | BIGINT | NOT NULL | 创建人 |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_by | BIGINT | NOT NULL | 更新人 |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

### 23. `sprint_snapshots` — 迭代快照

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| tenant_id | BIGINT | NOT NULL | 租户ID |
| workspace_id | BIGINT | NOT NULL | 工作空间ID |
| project_id | BIGINT | NOT NULL | 项目ID |
| sprint_id | BIGINT | NOT NULL | 迭代ID |
| snapshot_date | DATE | NOT NULL | 快照日期 |
| data | JSONB | DEFAULT '{}' | 快照数据 |
| status | entity_status | NOT NULL DEFAULT 'active' | 状态 |
| deleted | BOOLEAN | DEFAULT false | 软删除 |
| created_by | BIGINT | NOT NULL | 创建人 |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_by | BIGINT | NOT NULL | 更新人 |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

---

## 7、版本域

### 24. `versions` — 版本

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| code | VARCHAR(50) | | 版本编码 |
| name | VARCHAR(255) | NOT NULL | 版本名称 |
| tenant_id | BIGINT | NOT NULL | 租户ID |
| workspace_id | BIGINT | NOT NULL | 工作空间ID |
| project_id | BIGINT | NOT NULL | 项目ID |
| semver | VARCHAR(50) | NOT NULL | 语义化版本号 |
| description | TEXT | | 描述 |
| checklist | JSONB | DEFAULT '[]' | 发布检查清单 |
| release_notes | TEXT | | Release Notes |
| start_date | DATE | | 计划开始 |
| end_date | DATE | | 计划结束 |
| target_date | DATE | | 发布日期 |
| delivered_at | TIMESTAMPTZ | | 实际发布 |
| archived_at | TIMESTAMPTZ | | 归档时间 |
| status | version_status | NOT NULL DEFAULT 'planning' | planning/active/released/archived |
| version | INT | DEFAULT 1 | 乐观锁 |
| deleted | BOOLEAN | DEFAULT false | 软删除 |
| created_by | BIGINT | NOT NULL | 创建人 |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_by | BIGINT | NOT NULL | 更新人 |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

### 25. `version_delivery_snapshots` — 版本交付快照

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| tenant_id | BIGINT | NOT NULL | 租户ID |
| workspace_id | BIGINT | NOT NULL | 工作空间ID |
| project_id | BIGINT | NOT NULL | 项目ID |
| version_id | BIGINT | NOT NULL | 版本ID |
| progress | JSONB | DEFAULT '{}' | 进度快照 |
| quality | JSONB | DEFAULT '{}' | 质量快照 |
| release_notes | TEXT | | 发布说明 |
| status | entity_status | NOT NULL DEFAULT 'active' | 状态 |
| deleted | BOOLEAN | DEFAULT false | 软删除 |
| created_by | BIGINT | NOT NULL | 创建人 |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_by | BIGINT | NOT NULL | 更新人 |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

---

## 8、状态域

### 26. `states` — 状态

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| name | VARCHAR(255) | NOT NULL | 状态名称 |
| tenant_id | BIGINT | NOT NULL | 租户ID |
| workspace_id | BIGINT | NOT NULL | 工作空间ID |
| project_id | BIGINT | NOT NULL | 项目ID |
| group | VARCHAR(50) | NOT NULL | 状态组 |
| color | VARCHAR(20) | DEFAULT '#8DA2C2' | 颜色 |
| sequence | DOUBLE PRECISION | DEFAULT 65535 | 排序 |
| is_default | BOOLEAN | DEFAULT false | 是否默认 |
| applicable_types | TEXT[] | DEFAULT '{all}' | 适用类型 |
| template_set | VARCHAR(50) | DEFAULT 'custom' | 模板集 |
| status | entity_status | NOT NULL DEFAULT 'active' | 状态 |
| deleted | BOOLEAN | DEFAULT false | 软删除 |
| created_by | BIGINT | NOT NULL | 创建人 |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_by | BIGINT | NOT NULL | 更新人 |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

### 27. `state_transitions` — 状态流转规则

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| tenant_id | BIGINT | NOT NULL | 租户ID |
| workspace_id | BIGINT | NOT NULL | 工作空间ID |
| project_id | BIGINT | NOT NULL | 项目ID |
| type_code | VARCHAR(20) | DEFAULT 'all' | 工作项类型 |
| from_state_id | BIGINT | NOT NULL | 起始状态 |
| to_state_id | BIGINT | NOT NULL | 目标状态 |
| required_fields | JSONB | DEFAULT '[]' | 必填字段 |
| status | entity_status | NOT NULL DEFAULT 'active' | 状态 |
| deleted | BOOLEAN | DEFAULT false | 软删除 |
| created_by | BIGINT | NOT NULL | 创建人 |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_by | BIGINT | NOT NULL | 更新人 |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

---

## 9、模块/标签/估算域

### 28. `modules` — 模块

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| code | VARCHAR(50) | | 模块编码 |
| name | VARCHAR(255) | NOT NULL | 模块名称 |
| tenant_id | BIGINT | NOT NULL | 租户ID |
| workspace_id | BIGINT | NOT NULL | 工作空间ID |
| project_id | BIGINT | NOT NULL | 项目ID |
| public_id | UUID | DEFAULT gen_random_uuid() | 公开ID |
| description | TEXT | | 描述 |
| lead_id | BIGINT | | 负责人 |
| sort_order | FLOAT8 | DEFAULT 65535 | 排序 |
| status | entity_status | NOT NULL DEFAULT 'active' | 状态 |
| deleted | BOOLEAN | DEFAULT false | 软删除 |
| created_by | BIGINT | NOT NULL | 创建人 |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_by | BIGINT | NOT NULL | 更新人 |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

### 29. `labels` — 标签

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| code | VARCHAR(50) | | 标签编码 |
| name | VARCHAR(255) | NOT NULL | 标签名称 |
| tenant_id | BIGINT | NOT NULL | 租户ID |
| workspace_id | BIGINT | NOT NULL | 工作空间ID |
| project_id | BIGINT | NOT NULL | 项目ID |
| color | VARCHAR(20) | | 颜色 |
| description | TEXT | | 描述 |
| status | entity_status | NOT NULL DEFAULT 'active' | 状态 |
| deleted | BOOLEAN | DEFAULT false | 软删除 |
| created_by | BIGINT | NOT NULL | 创建人 |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_by | BIGINT | NOT NULL | 更新人 |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

### 30. `estimate_points` — 估算体系

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| code | VARCHAR(50) | NOT NULL | 估算编码 |
| name | VARCHAR(255) | NOT NULL | 估算名称 |
| tenant_id | BIGINT | NOT NULL | 租户ID |
| workspace_id | BIGINT | NOT NULL | 工作空间ID |
| description | TEXT | | 描述 |
| points_config | JSONB | NOT NULL | 估算配置 |
| is_default | BOOLEAN | DEFAULT false | 是否默认 |
| status | entity_status | NOT NULL DEFAULT 'active' | 状态 |
| deleted | BOOLEAN | DEFAULT false | 软删除 |
| created_by | BIGINT | NOT NULL | 创建人 |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_by | BIGINT | NOT NULL | 更新人 |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

---

## 10、自动化域

### 31. `automation_rules` — 自动化规则

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| code | VARCHAR(50) | | 规则编码 |
| name | VARCHAR(255) | NOT NULL | 规则名称 |
| tenant_id | BIGINT | NOT NULL | 租户ID |
| workspace_id | BIGINT | NOT NULL | 工作空间ID |
| project_id | BIGINT | | 项目ID |
| description | TEXT | | 描述 |
| trigger_type | VARCHAR(50) | NOT NULL | 触发类型 |
| conditions | JSONB | DEFAULT '{}' | 条件配置 |
| actions | JSONB | DEFAULT '{}' | 动作配置 |
| sort_order | INT | DEFAULT 0 | 排序 |
| status | entity_status | NOT NULL DEFAULT 'active' | 状态 |
| deleted | BOOLEAN | DEFAULT false | 软删除 |
| created_by | BIGINT | NOT NULL | 创建人 |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_by | BIGINT | NOT NULL | 更新人 |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

### 32. `rule_executions` — 规则执行日志

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| tenant_id | BIGINT | NOT NULL | 租户ID |
| workspace_id | BIGINT | NOT NULL | 工作空间ID |
| project_id | BIGINT | | 项目ID |
| rule_id | BIGINT | NOT NULL | 规则ID |
| trigger_event_id | BIGINT | | 触发事件 |
| duration_ms | INT | | 执行耗时 |
| error_message | TEXT | | 错误信息 |
| context_json | JSONB | | 上下文 |
| trigger_depth | SMALLINT | DEFAULT 0 | 触发深度 |
| via_automation | BOOLEAN | DEFAULT false | 自动触发标记 |
| status | work_item_status | NOT NULL | 执行状态 |
| deleted | BOOLEAN | DEFAULT false | 软删除 |
| created_by | BIGINT | NOT NULL | 创建人 |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间（分区键） |
| updated_by | BIGINT | NOT NULL | 更新人 |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

> 说明：按月分区表

### 33. `automation_templates` — 自动化模板

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| code | VARCHAR(50) | NOT NULL, UNIQUE | 模板编码 |
| name | VARCHAR(255) | NOT NULL | 模板名称 |
| description | TEXT | | 描述 |
| category | VARCHAR(50) | | 分类 |
| icon | VARCHAR(50) | | 图标 |
| template_config | JSONB | NOT NULL | 模板配置 |
| is_default | BOOLEAN | DEFAULT false | 是否默认模板 |
| sort_order | INT | DEFAULT 0 | 排序 |
| status | entity_status | NOT NULL DEFAULT 'active' | 状态 |
| deleted | BOOLEAN | DEFAULT false | 软删除 |
| created_by | BIGINT | NOT NULL | 创建人 |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_by | BIGINT | NOT NULL | 更新人 |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

> 说明：automation_templates 为系统预设，不含 tenant_id。

---

## 11、仪表盘域

### 34. `dashboards` — 仪表盘

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| code | VARCHAR(50) | | 仪表盘编码 |
| name | VARCHAR(255) | NOT NULL | 仪表盘名称 |
| tenant_id | BIGINT | NOT NULL | 租户ID |
| workspace_id | BIGINT | NOT NULL | 工作空间ID |
| project_id | BIGINT | | 项目ID |
| description | TEXT | | 描述 |
| layout | JSONB | DEFAULT '{}' | 布局配置 |
| status | entity_status | NOT NULL DEFAULT 'active' | 状态 |
| deleted | BOOLEAN | DEFAULT false | 软删除 |
| created_by | BIGINT | NOT NULL | 创建人 |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_by | BIGINT | NOT NULL | 更新人 |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

### 35. `dashboard_widgets` — 仪表盘组件

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| tenant_id | BIGINT | NOT NULL | 租户ID |
| workspace_id | BIGINT | NOT NULL | 工作空间ID |
| project_id | BIGINT | | 项目ID |
| dashboard_id | BIGINT | NOT NULL | 仪表盘ID |
| widget_type | VARCHAR(50) | NOT NULL | 组件类型 |
| name | VARCHAR(255) | NOT NULL | 组件名称 |
| config | JSONB | DEFAULT '{}' | 组件配置 |
| user_id | BIGINT | | 用户ID（个性化） |
| sort_order | INT | DEFAULT 0 | 排序 |
| status | entity_status | NOT NULL DEFAULT 'active' | 状态 |
| deleted | BOOLEAN | DEFAULT false | 软删除 |
| created_by | BIGINT | NOT NULL | 创建人 |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_by | BIGINT | NOT NULL | 更新人 |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

### 36. `dashboard_snapshots` — 仪表盘快照

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| tenant_id | BIGINT | NOT NULL | 租户ID |
| workspace_id | BIGINT | NOT NULL | 工作空间ID |
| project_id | BIGINT | NOT NULL | 项目ID |
| widget_type | VARCHAR(50) | NOT NULL | 组件类型 |
| refreshed_at | TIMESTAMPTZ | DEFAULT now() | 刷新时间 |
| data | JSONB | DEFAULT '{}' | 快照数据 |
| status | entity_status | NOT NULL DEFAULT 'active' | 状态 |
| deleted | BOOLEAN | DEFAULT false | 软删除 |
| created_by | BIGINT | NOT NULL | 创建人 |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_by | BIGINT | NOT NULL | 更新人 |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

### 37. `dashboard_templates` — 仪表盘模板

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| code | VARCHAR(50) | NOT NULL, UNIQUE | 模板编码 |
| name | VARCHAR(255) | NOT NULL | 模板名称 |
| description | TEXT | | 描述 |
| category | VARCHAR(50) | | 分类 |
| layout | JSONB | DEFAULT '{}' | 布局配置 |
| icon | VARCHAR(50) | | 图标 |
| is_default | BOOLEAN | DEFAULT false | 是否默认 |
| sort_order | INT | DEFAULT 0 | 排序 |
| status | entity_status | NOT NULL DEFAULT 'active' | 状态 |
| deleted | BOOLEAN | DEFAULT false | 软删除 |
| created_by | BIGINT | NOT NULL | 创建人 |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_by | BIGINT | NOT NULL | 更新人 |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

> 说明：dashboard_templates 为系统预设，不含 tenant_id。

---

## 12、通知域

### 38. `notifications` — 通知

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| tenant_id | BIGINT | NOT NULL | 租户ID |
| workspace_id | BIGINT | NOT NULL | 工作空间ID |
| recipient_id | BIGINT | NOT NULL | 接收人 |
| actor_id | BIGINT | | 触发人 |
| notification_type | VARCHAR(50) | NOT NULL | 通知类型 |
| title | VARCHAR(255) | NOT NULL | 标题 |
| content | TEXT | | 内容 |
| entity_type | VARCHAR(50) | | 关联实体类型 |
| entity_id | BIGINT | | 关联实体ID |
| is_read | BOOLEAN | DEFAULT false | 是否已读 |
| is_archived | BOOLEAN | DEFAULT false | 是否归档 |
| read_at | TIMESTAMPTZ | | 阅读时间 |
| status | notification_status | NOT NULL DEFAULT 'pending' | 状态 |
| deleted | BOOLEAN | DEFAULT false | 软删除 |
| created_by | BIGINT | NOT NULL | 创建人 |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_by | BIGINT | NOT NULL | 更新人 |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

### 39. `notification_deliveries` — 通知投递

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| tenant_id | BIGINT | NOT NULL | 租户ID |
| workspace_id | BIGINT | NOT NULL | 工作空间ID |
| notification_id | BIGINT | NOT NULL | 通知ID |
| channel | VARCHAR(20) | NOT NULL | 投递渠道 |
| status | notification_status | NOT NULL DEFAULT 'pending' | 状态 |
| retry_count | INT | DEFAULT 0 | 重试次数 |
| next_retry_at | TIMESTAMPTZ | | 下次重试 |
| delivered_at | TIMESTAMPTZ | | 投递时间 |
| error_message | TEXT | | 错误信息 |
| deleted | BOOLEAN | DEFAULT false | 软删除 |
| created_by | BIGINT | NOT NULL | 创建人 |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_by | BIGINT | NOT NULL | 更新人 |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

### 40. `notification_preferences` — 通知偏好

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| tenant_id | BIGINT | NOT NULL | 租户ID |
| workspace_id | BIGINT | NOT NULL | 工作空间ID |
| user_id | BIGINT | NOT NULL | 用户ID |
| channel_settings | JSONB | DEFAULT '{}' | 渠道设置 |
| mute_all | BOOLEAN | DEFAULT false | 全部免打扰 |
| digest_enabled | BOOLEAN | DEFAULT true | 摘要启用 |
| digest_schedule | VARCHAR(20) | DEFAULT 'daily' | 摘要频率 |
| status | entity_status | NOT NULL DEFAULT 'active' | 状态 |
| deleted | BOOLEAN | DEFAULT false | 软删除 |
| created_by | BIGINT | NOT NULL | 创建人 |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_by | BIGINT | NOT NULL | 更新人 |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

### 41. `notification_digests` — 通知摘要

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| tenant_id | BIGINT | NOT NULL | 租户ID |
| workspace_id | BIGINT | NOT NULL | 工作空间ID |
| user_id | BIGINT | NOT NULL | 用户ID |
| digest_type | VARCHAR(20) | NOT NULL | 摘要类型 |
| content | JSONB | DEFAULT '{}' | 摘要内容 |
| scheduled_for | TIMESTAMPTZ | | 计划投递时间 |
| sent_at | TIMESTAMPTZ | | 实际投递时间 |
| status | notification_status | NOT NULL DEFAULT 'pending' | 状态 |
| deleted | BOOLEAN | DEFAULT false | 软删除 |
| created_by | BIGINT | NOT NULL | 创建人 |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_by | BIGINT | NOT NULL | 更新人 |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

---

## 13、搜索域

### 42. `search_documents` — 搜索文档索引

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| tenant_id | BIGINT | NOT NULL | 租户ID |
| workspace_id | BIGINT | NOT NULL | 工作空间ID |
| project_id | BIGINT | | 项目ID |
| doc_type | VARCHAR(50) | NOT NULL | 文档类型 |
| doc_id | BIGINT | NOT NULL | 文档ID |
| title | VARCHAR(255) | | 标题 |
| identifier | VARCHAR(100) | | 标识符 |
| content | TEXT | | 内容 |
| search_tsv | tsvector | | 全文索引向量 |
| metadata | JSONB | DEFAULT '{}' | 元数据 |
| status | entity_status | NOT NULL DEFAULT 'active' | 状态 |
| deleted | BOOLEAN | DEFAULT false | 软删除 |
| created_by | BIGINT | NOT NULL | 创建人 |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_by | BIGINT | NOT NULL | 更新人 |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

### 43. `search_history` — 搜索历史

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| tenant_id | BIGINT | NOT NULL | 租户ID |
| workspace_id | BIGINT | NOT NULL | 工作空间ID |
| user_id | BIGINT | NOT NULL | 用户ID |
| query | TEXT | NOT NULL | 查询语句 |
| filters | JSONB | DEFAULT '{}' | 过滤条件 |
| result_count | INT | DEFAULT 0 | 结果数 |
| status | entity_status | NOT NULL DEFAULT 'active' | 状态 |
| deleted | BOOLEAN | DEFAULT false | 软删除 |
| created_by | BIGINT | NOT NULL | 创建人 |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_by | BIGINT | NOT NULL | 更新人 |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

### 44. `search_bookmarks` — 搜索收藏

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| tenant_id | BIGINT | NOT NULL | 租户ID |
| workspace_id | BIGINT | NOT NULL | 工作空间ID |
| project_id | BIGINT | | 项目ID |
| user_id | BIGINT | NOT NULL | 用户ID |
| name | VARCHAR(255) | NOT NULL | 收藏名称 |
| query | TEXT | | 查询语句 |
| filters | JSONB | DEFAULT '{}' | 过滤条件 |
| is_shared | BOOLEAN | DEFAULT false | 是否共享 |
| sort_order | FLOAT8 | DEFAULT 65535 | 排序 |
| status | entity_status | NOT NULL DEFAULT 'active' | 状态 |
| deleted | BOOLEAN | DEFAULT false | 软删除 |
| created_by | BIGINT | NOT NULL | 创建人 |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_by | BIGINT | NOT NULL | 更新人 |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

---

## 14、风险与度量域

### 45. `risk_rules` — 风险规则

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| code | VARCHAR(50) | | 规则编码 |
| name | VARCHAR(255) | NOT NULL | 规则名称 |
| tenant_id | BIGINT | NOT NULL | 租户ID |
| workspace_id | BIGINT | NOT NULL | 工作空间ID |
| project_id | BIGINT | | 项目ID |
| rule_type | VARCHAR(50) | NOT NULL | 规则类型 |
| condition_json | JSONB | DEFAULT '{}' | 条件配置 |
| notify_channels | TEXT[] | DEFAULT '{}' | 通知渠道 |
| is_active | BOOLEAN | DEFAULT true | 是否启用 |
| last_triggered | TIMESTAMPTZ | | 最后触发 |
| status | entity_status | NOT NULL DEFAULT 'active' | 状态 |
| deleted | BOOLEAN | DEFAULT false | 软删除 |
| created_by | BIGINT | NOT NULL | 创建人 |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_by | BIGINT | NOT NULL | 更新人 |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

### 46. `risk_alerts` — 风险告警

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| tenant_id | BIGINT | NOT NULL | 租户ID |
| workspace_id | BIGINT | NOT NULL | 工作空间ID |
| project_id | BIGINT | | 项目ID |
| rule_id | BIGINT | NOT NULL | 规则ID |
| severity | VARCHAR(20) | DEFAULT 'medium' | 严重度 |
| title | VARCHAR(255) | NOT NULL | 标题 |
| description | TEXT | | 描述 |
| metadata | JSONB | DEFAULT '{}' | 元数据 |
| is_resolved | BOOLEAN | DEFAULT false | 是否已解决 |
| resolved_at | TIMESTAMPTZ | | 解决时间 |
| resolved_by | BIGINT | | 解决人 |
| status | entity_status | NOT NULL DEFAULT 'active' | 状态 |
| deleted | BOOLEAN | DEFAULT false | 软删除 |
| created_by | BIGINT | NOT NULL | 创建人 |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_by | BIGINT | NOT NULL | 更新人 |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

### 47. `metric_snapshots` — 指标快照

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| tenant_id | BIGINT | NOT NULL | 租户ID |
| workspace_id | BIGINT | NOT NULL | 工作空间ID |
| project_id | BIGINT | | 项目ID |
| granularity | VARCHAR(20) | NOT NULL | 粒度 |
| ref_id | BIGINT | | 引用ID |
| metric | VARCHAR(50) | NOT NULL | 指标名 |
| snapshot_date | DATE | NOT NULL | 快照日期 |
| value | NUMERIC | | 指标值 |
| metadata | JSONB | DEFAULT '{}' | 元数据 |
| status | entity_status | NOT NULL DEFAULT 'active' | 状态 |
| deleted | BOOLEAN | DEFAULT false | 软删除 |
| created_by | BIGINT | NOT NULL | 创建人 |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_by | BIGINT | NOT NULL | 更新人 |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

### 48. `metric_adjustments` — 指标调整

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| tenant_id | BIGINT | NOT NULL | 租户ID |
| workspace_id | BIGINT | NOT NULL | 工作空间ID |
| project_id | BIGINT | | 项目ID |
| snapshot_id | BIGINT | | 快照ID |
| metric | VARCHAR(50) | NOT NULL | 指标名 |
| original_value | NUMERIC | | 原始值 |
| adjusted_value | NUMERIC | | 调整值 |
| reason | TEXT | | 调整原因 |
| adjusted_by | BIGINT | | 调整人 |
| status | entity_status | NOT NULL DEFAULT 'active' | 状态 |
| deleted | BOOLEAN | DEFAULT false | 软删除 |
| created_by | BIGINT | NOT NULL | 创建人 |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_by | BIGINT | NOT NULL | 更新人 |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

---

## 15、入口工单域

### 49. `intake_channels` — 入口渠道

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| code | VARCHAR(50) | | 渠道编码 |
| name | VARCHAR(255) | NOT NULL | 渠道名称 |
| tenant_id | BIGINT | NOT NULL | 租户ID |
| workspace_id | BIGINT | NOT NULL | 工作空间ID |
| project_id | BIGINT | | 项目ID |
| slug | VARCHAR(100) | NOT NULL | URL标识 |
| description | TEXT | | 描述 |
| is_active | BOOLEAN | DEFAULT true | 是否启用 |
| config | JSONB | DEFAULT '{}' | 渠道配置 |
| status | entity_status | NOT NULL DEFAULT 'active' | 状态 |
| deleted | BOOLEAN | DEFAULT false | 软删除 |
| created_by | BIGINT | NOT NULL | 创建人 |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_by | BIGINT | NOT NULL | 更新人 |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

### 50. `intake_issues` — 入口工单

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| code | VARCHAR(50) | | 工单编码 |
| name | VARCHAR(255) | NOT NULL | 工单标题 |
| tenant_id | BIGINT | NOT NULL | 租户ID |
| workspace_id | BIGINT | NOT NULL | 工作空间ID |
| project_id | BIGINT | | 项目ID |
| channel_id | BIGINT | NOT NULL | 渠道ID |
| tracking_id | VARCHAR(50) | | 跟踪ID |
| submitter_name | VARCHAR(255) | | 提交人姓名 |
| submitter_email | VARCHAR(255) | NOT NULL | 提交人邮箱 |
| description | TEXT | | 描述 |
| priority | VARCHAR(20) | DEFAULT 'medium' | 优先级 |
| status | intake_issue_status | NOT NULL DEFAULT 'open' | 状态 |
| linked_entity_type | VARCHAR(50) | | 关联实体类型 |
| linked_entity_id | BIGINT | | 关联实体ID |
| resolved_at | TIMESTAMPTZ | | 解决时间 |
| resolved_by | BIGINT | | 解决人 |
| deleted | BOOLEAN | DEFAULT false | 软删除 |
| created_by | BIGINT | NOT NULL | 创建人 |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_by | BIGINT | NOT NULL | 更新人 |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

---

## 16、Webhooks 与审计

### 51. `webhooks` — Webhook 配置

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| code | VARCHAR(50) | | Webhook编码 |
| name | VARCHAR(100) | NOT NULL | 名称 |
| tenant_id | BIGINT | NOT NULL | 租户ID |
| workspace_id | BIGINT | NOT NULL | 工作空间ID |
| project_id | BIGINT | | 项目ID |
| target_url | TEXT | NOT NULL | 目标URL |
| secret | VARCHAR(255) | NOT NULL | HMAC密钥 |
| events | TEXT[] | DEFAULT '{}' | 事件白名单 |
| is_active | BOOLEAN | DEFAULT true | 启用 |
| last_error | TEXT | | 最后错误 |
| last_triggered | TIMESTAMPTZ | | 最后触发 |
| last_status | VARCHAR(20) | | 最后状态 |
| unhealthy_at | TIMESTAMPTZ | | 异常时间 |
| status | entity_status | NOT NULL DEFAULT 'active' | 状态 |
| deleted | BOOLEAN | DEFAULT false | 软删除 |
| created_by | BIGINT | NOT NULL | 创建人 |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_by | BIGINT | NOT NULL | 更新人 |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

### 52. `webhook_logs` — Webhook 投递日志

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| tenant_id | BIGINT | NOT NULL | 租户ID |
| workspace_id | BIGINT | NOT NULL | 工作空间ID |
| webhook_id | BIGINT | NOT NULL | WebhookID |
| delivery_id | VARCHAR(64) | NOT NULL | 投递ID |
| event_type | VARCHAR(80) | NOT NULL | 事件类型 |
| event_id | BIGINT | | 事件ID |
| request_url | TEXT | NOT NULL | 请求URL |
| request_method | VARCHAR(10) | DEFAULT 'POST' | 请求方法 |
| request_headers | JSONB | | 请求头 |
| request_body | TEXT | | 请求体 |
| response_status | INT | | 响应状态码 |
| response_body | TEXT | | 响应体 |
| response_headers | JSONB | | 响应头 |
| attempt | SMALLINT | DEFAULT 1 | 尝试次数 |
| duration_ms | INT | | 耗时 |
| error | TEXT | | 错误 |
| status | VARCHAR(20) | NOT NULL | 投递状态 |
| deleted | BOOLEAN | DEFAULT false | 软删除 |
| created_by | BIGINT | NOT NULL | 创建人 |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间（分区键） |
| updated_by | BIGINT | NOT NULL | 更新人 |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

> 说明：按月分区表

---

## 17、工作台与视图偏好

### 53. `workbench_configs` — 工作台配置

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| tenant_id | BIGINT | NOT NULL | 租户ID |
| workspace_id | BIGINT | NOT NULL | 工作空间ID |
| project_id | BIGINT | | 项目ID |
| user_id | BIGINT | NOT NULL | 用户ID |
| layout | JSONB | DEFAULT '{}' | 布局配置 |
| widget_states | JSONB | DEFAULT '{}' | 组件状态 |
| focus_enabled | BOOLEAN | DEFAULT false | 专注模式 |
| status | entity_status | NOT NULL DEFAULT 'active' | 状态 |
| deleted | BOOLEAN | DEFAULT false | 软删除 |
| created_by | BIGINT | NOT NULL | 创建人 |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_by | BIGINT | NOT NULL | 更新人 |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

### 54. `view_preferences` — 视图偏好

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| tenant_id | BIGINT | NOT NULL | 租户ID |
| workspace_id | BIGINT | NOT NULL | 工作空间ID |
| project_id | BIGINT | | 项目ID |
| user_id | BIGINT | NOT NULL | 用户ID |
| view_type | VARCHAR(20) | NOT NULL | 视图类型 |
| layout | VARCHAR(20) | DEFAULT 'list' | 布局 |
| columns | JSONB | DEFAULT '[]' | 列配置 |
| filters | JSONB | DEFAULT '{}' | 过滤条件 |
| sort | JSONB | DEFAULT '{}' | 排序配置 |
| extra | JSONB | DEFAULT '{}' | 扩展配置 |
| status | entity_status | NOT NULL DEFAULT 'active' | 状态 |
| deleted | BOOLEAN | DEFAULT false | 软删除 |
| created_by | BIGINT | NOT NULL | 创建人 |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_by | BIGINT | NOT NULL | 更新人 |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

### 55. `recent_items` — 最近访问

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| tenant_id | BIGINT | NOT NULL | 租户ID |
| workspace_id | BIGINT | NOT NULL | 工作空间ID |
| project_id | BIGINT | | 项目ID |
| user_id | BIGINT | NOT NULL | 用户ID |
| item_type | VARCHAR(20) | NOT NULL | 访问类型 |
| item_id | BIGINT | NOT NULL | 关联ID |
| title | VARCHAR(255) | | 标题 |
| identifier | VARCHAR(100) | | 标识符 |
| accessed_at | TIMESTAMPTZ | DEFAULT now() | 访问时间 |
| status | entity_status | NOT NULL DEFAULT 'active' | 状态 |
| deleted | BOOLEAN | DEFAULT false | 软删除 |
| created_by | BIGINT | NOT NULL | 创建人 |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_by | BIGINT | NOT NULL | 更新人 |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

---

## 18、知识库与文档

### 56. `knowledge_spaces` — 知识空间

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| code | VARCHAR(50) | | 编码 |
| name | VARCHAR(255) | NOT NULL | 名称 |
| tenant_id | BIGINT | NOT NULL | 租户ID |
| workspace_id | BIGINT | NOT NULL | 工作空间ID |
| description | TEXT | | 描述 |
| icon | VARCHAR(50) | | 图标 |
| status | entity_status | NOT NULL DEFAULT 'active' | 状态 |
| deleted | BOOLEAN | DEFAULT false | 软删除 |
| created_by | BIGINT | NOT NULL | 创建人 |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_by | BIGINT | NOT NULL | 更新人 |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

### 57. `knowledge_pages` — 知识页面

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| code | VARCHAR(50) | | 编码 |
| name | VARCHAR(255) | NOT NULL | 名称 |
| tenant_id | BIGINT | NOT NULL | 租户ID |
| workspace_id | BIGINT | NOT NULL | 工作空间ID |
| knowledge_space_id | BIGINT | NOT NULL | 知识空间ID |
| public_id | UUID | DEFAULT gen_random_uuid() | 公开ID |
| parent_id | BIGINT | | 父页面 |
| depth | SMALLINT | DEFAULT 1 CHECK (1-3) | 层级 |
| content_json | JSONB | | 内容（JSON） |
| content_html | TEXT | | 内容（HTML） |
| sort_order | DOUBLE PRECISION | DEFAULT 65535 | 排序 |
| version | INT | DEFAULT 1 | 乐观锁 |
| status | entity_status | NOT NULL DEFAULT 'active' | 状态 |
| deleted | BOOLEAN | DEFAULT false | 软删除 |
| created_by | BIGINT | NOT NULL | 创建人 |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_by | BIGINT | NOT NULL | 更新人 |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

---

## 19、入口与 SSO

### 58. `sso_providers` — SSO 提供方

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| name | VARCHAR(255) | NOT NULL | 名称 |
| protocol | VARCHAR(20) | DEFAULT 'oidc' | 协议类型 |
| issuer_url | TEXT | | Issuer URL |
| client_id | VARCHAR(255) | NOT NULL | 客户端ID |
| client_secret | TEXT | NOT NULL | 客户端密钥 |
| redirect_uri | TEXT | NOT NULL | 回调URL |
| auth_url | TEXT | | 认证URL |
| token_url | TEXT | | Token URL |
| userinfo_url | TEXT | | 用户信息URL |
| jwks_url | TEXT | | JWKS URL |
| sso_url | TEXT | | SSO URL |
| idp_issuer | VARCHAR(255) | | IDP Issuer |
| idp_certificate | TEXT | | IDP 证书 |
| skip_signature | BOOLEAN | DEFAULT false | 跳过签名 |
| scopes | TEXT | DEFAULT 'openid email profile' | 范围 |
| auto_create_user | BOOLEAN | DEFAULT true | 自动创建用户 |
| default_role | VARCHAR(20) | DEFAULT 'member' | 默认角色 |
| attribute_mapping | JSONB | DEFAULT '{}' | 属性映射 |
| enabled | BOOLEAN | DEFAULT true | 启用 |
| status | entity_status | NOT NULL DEFAULT 'active' | 状态 |
| deleted | BOOLEAN | DEFAULT false | 软删除 |
| created_by | BIGINT | NOT NULL | 创建人 |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_by | BIGINT | NOT NULL | 更新人 |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

> 说明：sso_providers 为系统级配置，不含 tenant_id。

---

## 20、其他系统表

### 59. `api_tokens` — API 令牌

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| tenant_id | BIGINT | NOT NULL | 租户ID |
| user_id | BIGINT | NOT NULL | 用户ID |
| name | VARCHAR(255) | NOT NULL | 令牌名称 |
| token_hash | VARCHAR(255) | NOT NULL, UNIQUE | 令牌哈希 |
| scopes | TEXT[] | DEFAULT '{}' | 权限范围 |
| expires_at | TIMESTAMPTZ | | 过期时间 |
| last_used_at | TIMESTAMPTZ | | 最后使用 |
| revoked_at | TIMESTAMPTZ | | 撤销时间 |
| status | entity_status | NOT NULL DEFAULT 'active' | 状态 |
| deleted | BOOLEAN | DEFAULT false | 软删除 |
| created_by | BIGINT | NOT NULL | 创建人 |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_by | BIGINT | NOT NULL | 更新人 |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

### 60. `audit_logs` — 审计日志

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| tenant_id | BIGINT | NOT NULL | 租户ID |
| workspace_id | BIGINT | | 工作空间ID |
| actor_id | BIGINT | | 操作人 |
| action | VARCHAR(100) | NOT NULL | 操作类型 |
| target_type | VARCHAR(50) | | 目标类型 |
| target_id | BIGINT | | 目标ID |
| details | JSONB | DEFAULT '{}' | 详情 |
| ip_address | INET | | IP地址 |
| user_agent | TEXT | | 用户代理 |
| status | entity_status | NOT NULL DEFAULT 'active' | 状态 |
| deleted | BOOLEAN | DEFAULT false | 软删除 |
| created_by | BIGINT | NOT NULL | 创建人 |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间（分区键） |
| updated_by | BIGINT | NOT NULL | 更新人 |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

> 说明：按月分区表

### 61. `domain_events` — 领域事件

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| tenant_id | BIGINT | NOT NULL | 租户ID |
| workspace_id | BIGINT | | 工作空间ID |
| event_type | VARCHAR(100) | NOT NULL | 事件类型 |
| aggregate_type | VARCHAR(50) | | 聚合类型 |
| aggregate_id | BIGINT | | 聚合ID |
| payload | JSONB | NOT NULL | 事件数据 |
| metadata | JSONB | DEFAULT '{}' | 元数据 |
| published_at | TIMESTAMPTZ | | 发布时间 |
| status | entity_status | NOT NULL DEFAULT 'active' | 状态 |
| deleted | BOOLEAN | DEFAULT false | 软删除 |
| created_by | BIGINT | NOT NULL | 创建人 |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_by | BIGINT | NOT NULL | 更新人 |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

### 62. `invitations` — 邀请

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | BIGINT | PK | 雪花ID |
| tenant_id | BIGINT | NOT NULL | 租户ID |
| workspace_id | BIGINT | NOT NULL | 工作空间ID |
| email | VARCHAR(255) | NOT NULL | 邮箱 |
| inviter_id | BIGINT | NOT NULL | 邀请人 |
| role | workspace_role_enum | NOT NULL | 角色 |
| token_hash | VARCHAR(255) | NOT NULL | 令牌哈希 |
| message | TEXT | | 邀请消息 |
| expires_at | TIMESTAMPTZ | NOT NULL | 过期时间 |
| accepted_at | TIMESTAMPTZ | | 接受时间 |
| revoked_at | TIMESTAMPTZ | | 撤销时间 |
| status | VARCHAR(20) | NOT NULL | 状态 |
| deleted | BOOLEAN | DEFAULT false | 软删除 |
| created_by | BIGINT | NOT NULL | 创建人 |
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间 |
| updated_by | BIGINT | NOT NULL | 更新人 |
| updated_at | TIMESTAMPTZ | DEFAULT now() | 更新时间 |

> UNIQUE(token_hash)

---

## 21、数据隔离说明

### 需要租户数据隔离的表

所有业务表均需包含 `tenant_id` 字段，实现行级数据隔离：

- 租户域：tenants, users, user_roles, roles, menus
- 工作空间域：workspaces, workspace_members, invitations
- 项目域：projects, project_members, project_sequences, modules, labels, estimate_points
- 工作项域：task, requirement, defect, task_*, requirement_*, defect_*
- 迭代域：sprints, sprint_snapshots
- 版本域：versions, version_delivery_snapshots
- 状态域：states, state_transitions
- 自动化域：automation_rules, rule_executions
- 仪表盘域：dashboards, dashboard_widgets, dashboard_snapshots
- 通知域：notifications, notification_deliveries, notification_preferences, notification_digests
- 搜索域：search_documents, search_history, search_bookmarks
- 风险度量：risk_rules, risk_alerts, metric_snapshots, metric_adjustments
- 入口工单：intake_channels, intake_issues
- 工作台：workbench_configs, view_preferences, recent_items
- 文档：knowledge_spaces, knowledge_pages, documents, pages
- 集成：webhooks, webhook_logs, sso_providers, deployment_events
- 其他：api_tokens, audit_logs, domain_events

### 不需要租户数据隔离的表

系统级配置表，所有租户共享：

- `menus` — 菜单/权限资源（系统预定义）
- `roles` — 角色定义（系统预置 + 租户自定义混合，通过 is_system 区分）
- `automation_templates` — 自动化模板（系统预设）
- `dashboard_templates` — 仪表盘模板（系统预设）
- `sso_providers` — SSO 提供方（系统级配置）
- `schema_migrations` — 迁移版本记录（系统级）
