-- ===========================================================================
-- Ydsz Plane 数据库初始化脚本
-- 数据库: PostgreSQL 16+ (兼容达梦/人大金仓)
-- 依据  : docs/Ydsz Plane 数据库表设计.md (V2.1, 2026-08-10)
-- 表数量: 121 张
-- 生成  : 2026-08-17 (scripts/fix_db_consistency.py 一致性修复, 原自动生成后手动补齐)
-- 要点  :
--   1. 主键 BIGINT 雪花ID, 应用层生成, 非自增
--   2. 三层次隔离: tenant_id → workspace_id → project_id (系统级表除外)
--   3. 无物理外键, 逻辑外键字段全部建索引
--   4. 软删除 deleted BOOLEAN DEFAULT false, 查询索引均带 WHERE NOT deleted
--   5. 状态字段使用 PostgreSQL ENUM 类型
--   6. 字段排序标准化: id/code/name 开头, status/deleted/tenant_id/created_by/created_at/updated_by/updated_at 结尾
-- ===========================================================================
-- ===========================================================================

-- 第一部分: ENUM 类型定义 (幂等)

-- ===========================================================================

DO $$ BEGIN
    CREATE TYPE tenant_status AS ENUM ('active', 'disabled', 'archived', 'expired');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

DO $$ BEGIN
    CREATE TYPE user_status AS ENUM ('active', 'inactive', 'locked');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

DO $$ BEGIN
    CREATE TYPE entity_status AS ENUM ('active', 'inactive', 'archived');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

DO $$ BEGIN
    CREATE TYPE project_status AS ENUM ('active', 'archived');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

DO $$ BEGIN
    CREATE TYPE work_item_status AS ENUM ('draft', 'active', 'completed', 'cancelled');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

DO $$ BEGIN
    CREATE TYPE sprint_status AS ENUM ('planned', 'active', 'completed');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

DO $$ BEGIN
    CREATE TYPE version_status AS ENUM ('planning', 'active', 'released', 'archived');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

DO $$ BEGIN
    CREATE TYPE notification_status AS ENUM ('pending', 'sent', 'failed', 'read');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

DO $$ BEGIN
    CREATE TYPE workspace_role_enum AS ENUM ('owner', 'admin', 'member', 'guest');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

DO $$ BEGIN
    CREATE TYPE project_role_enum AS ENUM ('admin', 'member');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

DO $$ BEGIN
    CREATE TYPE attachment_status AS ENUM ('uploading', 'available', 'archived', 'deleted');
DO $$ BEGIN
    CREATE TYPE intake_issue_status AS ENUM ('open', 'accepted', 'rejected', 'archived');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

DO $$ BEGIN
    CREATE TYPE attachment_status AS ENUM ('uploading', 'available', 'archived', 'deleted');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

-- ===========================================================================
-- 第二部分: 通用触发器函数
-- ===========================================================================

-- 更新时间戳 (所有含 updated_at 的表)
CREATE OR REPLACE FUNCTION set_updated_at() RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- 乐观锁版本自增 (所有含 version 字段的表)
CREATE OR REPLACE FUNCTION bump_version() RETURNS TRIGGER AS $$
BEGIN
    NEW.version = OLD.version + 1;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- ===========================================================================
-- 序列（用于各表 Snowflake 风格 ID 生成）
-- ===========================================================================
CREATE SEQUENCE IF NOT EXISTS states_id_seq START 1 INCREMENT 1;
CREATE SEQUENCE IF NOT EXISTS state_transitions_id_seq START 1 INCREMENT 1;

-- ===========================================================================
-- 建表 (110 张)
-- ===========================================================================

--   1. tenants — 租户（组织机构）
CREATE TABLE IF NOT EXISTS tenants (
    id                       BIGINT PRIMARY KEY,
    code                     VARCHAR(50) UNIQUE,
    name                     VARCHAR(255) NOT NULL,
    slug                     VARCHAR(100) NOT NULL UNIQUE,
    logo_url                 TEXT,
    owner_id                 BIGINT NOT NULL DEFAULT 0,
    timezone                 VARCHAR(50) DEFAULT 'Asia/Shanghai',
    language                 VARCHAR(20) DEFAULT 'zh-CN',
    brand_config             JSONB,
    status                   tenant_status NOT NULL DEFAULT 'active',
    max_projects             INTEGER DEFAULT 10,
    max_users                INTEGER DEFAULT 50,
    max_workspaces           INTEGER DEFAULT 5,
    expired_at               TIMESTAMPTZ,
    config                   JSONB,
    created_at               TIMESTAMPTZ DEFAULT now(),
    updated_at               TIMESTAMPTZ DEFAULT now()
);

--   2. workspaces — 工作空间
CREATE TABLE IF NOT EXISTS workspaces (
    id                       BIGINT PRIMARY KEY,
    code                     VARCHAR(50),
    name                     VARCHAR(255) NOT NULL,
    slug                     VARCHAR(100) NOT NULL,
    tenant_id                BIGINT NOT NULL DEFAULT 1,
    logo_url                 TEXT,
    owner_id                 BIGINT NOT NULL DEFAULT 0,
    timezone                 VARCHAR(50) DEFAULT 'Asia/Shanghai',
    language                 VARCHAR(20) DEFAULT 'zh-CN',
    status                   entity_status NOT NULL DEFAULT 'active',
    max_projects             INTEGER DEFAULT 20,
    config                   JSONB,
    deleted                  BOOLEAN DEFAULT false,
    created_by               BIGINT NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ DEFAULT now(),
    updated_by               BIGINT NOT NULL DEFAULT 0,
    updated_at               TIMESTAMPTZ DEFAULT now()
);

--   3. workspace_members — 工作空间成员
CREATE TABLE IF NOT EXISTS workspace_members (
    id                       BIGINT PRIMARY KEY,
    tenant_id                BIGINT NOT NULL DEFAULT 1,
    workspace_id             BIGINT NOT NULL DEFAULT 0,
    user_id                  BIGINT NOT NULL,
    role                     workspace_role_enum NOT NULL DEFAULT 'member',
    status                   entity_status NOT NULL DEFAULT 'active',
    joined_at                TIMESTAMPTZ DEFAULT now(),
    deleted                  BOOLEAN DEFAULT false,
    created_by               BIGINT NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ DEFAULT now(),
    updated_by               BIGINT NOT NULL DEFAULT 0,
    updated_at               TIMESTAMPTZ DEFAULT now()
);

--   4. users — 用户
CREATE TABLE IF NOT EXISTS users (
    id                       BIGINT PRIMARY KEY,
    code                     VARCHAR(50),
    name                     VARCHAR(255) NOT NULL,
    email                    VARCHAR(255) NOT NULL,
    phone                    VARCHAR(20),
    password_hash            TEXT,
    avatar_url               TEXT,
    status                   user_status NOT NULL DEFAULT 'active',
    is_super_admin           BOOLEAN DEFAULT false,
    last_login_at            TIMESTAMPTZ,
    tenant_id                BIGINT NOT NULL DEFAULT 1,
    deleted                  BOOLEAN DEFAULT false,
    created_by               BIGINT NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ DEFAULT now(),
    updated_by               BIGINT NOT NULL DEFAULT 0,
    updated_at               TIMESTAMPTZ DEFAULT now()
);

--   5. roles — 角色
CREATE TABLE IF NOT EXISTS roles (
    id                       BIGINT PRIMARY KEY,
    code                     VARCHAR(50),
    name                     VARCHAR(255) NOT NULL,
    description              TEXT,
    status                   entity_status NOT NULL DEFAULT 'active',
    is_system                BOOLEAN DEFAULT false,
    role_scope               VARCHAR(50) CHECK (role_scope IN ('tenant', 'project')),
    deleted                  BOOLEAN DEFAULT false,
    created_by               BIGINT NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ DEFAULT now(),
    updated_by               BIGINT NOT NULL DEFAULT 0,
    updated_at               TIMESTAMPTZ DEFAULT now()
);

--   6. menus — 菜单 / 权限资源
CREATE TABLE IF NOT EXISTS menus (
    id                       BIGINT PRIMARY KEY,
    code                     VARCHAR(100) NOT NULL UNIQUE,
    name                     VARCHAR(255) NOT NULL,
    menu_type                VARCHAR(20) NOT NULL CHECK (menu_type IN ('menu', 'button', 'api')),
    parent_id                BIGINT,
    path                     VARCHAR(255),
    icon                     VARCHAR(100),
    sort_order               INTEGER DEFAULT 0,
    status                   entity_status DEFAULT 'active',
    deleted                  BOOLEAN DEFAULT false,
    created_by               BIGINT NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ DEFAULT now(),
    updated_by               BIGINT NOT NULL DEFAULT 0,
    updated_at               TIMESTAMPTZ DEFAULT now()
);

--   7. user_roles — 用户角色关联
CREATE TABLE IF NOT EXISTS user_roles (
    id                       BIGINT PRIMARY KEY,
    tenant_id                BIGINT NOT NULL DEFAULT 1,
    user_id                  BIGINT NOT NULL,
    role_id                  BIGINT NOT NULL,
    status                   entity_status NOT NULL DEFAULT 'active',
    deleted                  BOOLEAN DEFAULT false,
    created_by               BIGINT NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ DEFAULT now(),
    updated_by               BIGINT NOT NULL DEFAULT 0,
    updated_at               TIMESTAMPTZ DEFAULT now()
);

--   8. projects — 项目
CREATE TABLE IF NOT EXISTS projects (
    id                       BIGINT PRIMARY KEY,
    code                     VARCHAR(50),
    name                     VARCHAR(255) NOT NULL,
    tenant_id                BIGINT NOT NULL DEFAULT 1,
    workspace_id             BIGINT NOT NULL DEFAULT 0,
    identifier               VARCHAR(20),
    slug                     VARCHAR(100),
    description              TEXT,
    icon                     VARCHAR(50),
    cover_image_url          TEXT,
    network                  VARCHAR(20) DEFAULT 'private',
    template                 VARCHAR(20) DEFAULT 'generic',
    status                   project_status NOT NULL DEFAULT 'active',
    modules                  JSONB,
    start_date               DATE,
    target_date              DATE,
    owner_id                 BIGINT NOT NULL DEFAULT 0,
    version                  INTEGER DEFAULT 1,
    sort_order               DOUBLE PRECISION DEFAULT 65535,
    deleted                  BOOLEAN DEFAULT false,
    created_by               BIGINT NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ DEFAULT now(),
    updated_by               BIGINT NOT NULL DEFAULT 0,
    updated_at               TIMESTAMPTZ DEFAULT now()
);

--   9. project_members — 项目成员
CREATE TABLE IF NOT EXISTS project_members (
    id                       BIGINT PRIMARY KEY,
    tenant_id                BIGINT NOT NULL DEFAULT 1,
    workspace_id             BIGINT NOT NULL DEFAULT 0,
    project_id               BIGINT NOT NULL DEFAULT 0,
    user_id                  BIGINT NOT NULL,
    role                     project_role_enum NOT NULL DEFAULT 'member',
    joined_at                TIMESTAMPTZ DEFAULT now(),
    status                   entity_status NOT NULL DEFAULT 'active',
    deleted                  BOOLEAN DEFAULT false,
    created_by               BIGINT NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ DEFAULT now(),
    updated_by               BIGINT NOT NULL DEFAULT 0,
    updated_at               TIMESTAMPTZ DEFAULT now()
);

--  10. project_sequences — 项目序列发号器
CREATE TABLE IF NOT EXISTS project_sequences (
    id                       BIGINT PRIMARY KEY,
    tenant_id                BIGINT NOT NULL DEFAULT 1,
    workspace_id             BIGINT NOT NULL DEFAULT 0,
    project_id               BIGINT NOT NULL UNIQUE,
    sequence_id              BIGINT NOT NULL DEFAULT 0,
    status                   entity_status NOT NULL DEFAULT 'active',
    deleted                  BOOLEAN DEFAULT false,
    created_by               BIGINT NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ DEFAULT now(),
    updated_by               BIGINT NOT NULL DEFAULT 0,
    updated_at               TIMESTAMPTZ DEFAULT now()
);

--  11. task — 任务
CREATE TABLE IF NOT EXISTS task (
    id                       BIGINT PRIMARY KEY,
    code                     VARCHAR(50),
    name                     VARCHAR(255) NOT NULL,
    tenant_id                BIGINT NOT NULL DEFAULT 1,
    workspace_id             BIGINT NOT NULL DEFAULT 0,
    project_id               BIGINT NOT NULL DEFAULT 0,
    sequence_id              BIGINT NOT NULL,
    public_id                UUID DEFAULT gen_random_uuid(),
    parent_id                BIGINT,
    depth                    SMALLINT DEFAULT 1 CHECK (depth BETWEEN 1 AND 3),
    description_json         JSONB,
    description_html         TEXT,
    description_stripped     TEXT,
    state_id                 BIGINT NOT NULL,
    priority                 VARCHAR(20) DEFAULT 'none',
    category                 VARCHAR(20),
    actual_effort            NUMERIC(8,2),
    remaining_effort         NUMERIC(8,2),
    delay_reason             VARCHAR(50),
    point                    SMALLINT CHECK (point BETWEEN 0 AND 12),
    estimate_point_id        BIGINT,
    sprint_id                BIGINT,
    version_id               BIGINT,
    progress                 SMALLINT DEFAULT 0 CHECK (progress BETWEEN 0 AND 100),
    start_date               DATE,
    target_date              DATE,
    completed_at             TIMESTAMPTZ,
    is_draft                 BOOLEAN DEFAULT false,
    archived_at              TIMESTAMPTZ,
    sort_order               DOUBLE PRECISION DEFAULT 65535,
    version                  INTEGER DEFAULT 1,
    status                   work_item_status NOT NULL DEFAULT 'active',
    deleted                  BOOLEAN DEFAULT false,
    created_by               BIGINT NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ DEFAULT now(),
    updated_by               BIGINT NOT NULL DEFAULT 0,
    updated_at               TIMESTAMPTZ DEFAULT now()
);

--  12. requirement — 需求
CREATE TABLE IF NOT EXISTS requirement (
    id                       BIGINT PRIMARY KEY,
    code                     VARCHAR(50),
    name                     VARCHAR(255) NOT NULL,
    tenant_id                BIGINT NOT NULL DEFAULT 1,
    workspace_id             BIGINT NOT NULL DEFAULT 0,
    project_id               BIGINT NOT NULL DEFAULT 0,
    sequence_id              BIGINT NOT NULL,
    public_id                UUID DEFAULT gen_random_uuid(),
    parent_id                BIGINT,
    depth                    SMALLINT DEFAULT 1 CHECK (depth BETWEEN 1 AND 3),
    description_json         JSONB,
    description_html         TEXT,
    description_stripped     TEXT,
    state_id                 BIGINT NOT NULL,
    priority                 VARCHAR(20) DEFAULT 'none',
    source                   VARCHAR(50),
    acceptance_criteria      JSONB,
    business_value           TEXT,
    review_status            VARCHAR(20),
    point                    SMALLINT CHECK (point BETWEEN 0 AND 12),
    estimate_point_id        BIGINT,
    sprint_id                BIGINT,
    version_id               BIGINT,
    progress                 SMALLINT DEFAULT 0 CHECK (progress BETWEEN 0 AND 100),
    start_date               DATE,
    target_date              DATE,
    completed_at             TIMESTAMPTZ,
    is_draft                 BOOLEAN DEFAULT false,
    archived_at              TIMESTAMPTZ,
    sort_order               DOUBLE PRECISION DEFAULT 65535,
    version                  INTEGER DEFAULT 1,
    status                   work_item_status NOT NULL DEFAULT 'active',
    deleted                  BOOLEAN DEFAULT false,
    created_by               BIGINT NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ DEFAULT now(),
    updated_by               BIGINT NOT NULL DEFAULT 0,
    updated_at               TIMESTAMPTZ DEFAULT now()
);

--  13. defect — 缺陷
CREATE TABLE IF NOT EXISTS defect (
    id                       BIGINT PRIMARY KEY,
    code                     VARCHAR(50),
    name                     VARCHAR(255) NOT NULL,
    tenant_id                BIGINT NOT NULL DEFAULT 1,
    workspace_id             BIGINT NOT NULL DEFAULT 0,
    project_id               BIGINT NOT NULL DEFAULT 0,
    sequence_id              BIGINT NOT NULL,
    public_id                UUID DEFAULT gen_random_uuid(),
    parent_id                BIGINT,
    depth                    SMALLINT DEFAULT 1 CHECK (depth BETWEEN 1 AND 3),
    description_json         JSONB,
    description_html         TEXT,
    description_stripped     TEXT,
    state_id                 BIGINT NOT NULL,
    priority                 VARCHAR(20) DEFAULT 'none',
    severity                 SMALLINT NOT NULL CHECK (severity BETWEEN 1 AND 5),
    found_phase              VARCHAR(20),
    found_version_id         BIGINT,
    fix_version_id           BIGINT,
    root_cause_category      VARCHAR(50),
    verifier_id              BIGINT,
    environment              JSONB,
    reproduce_steps          JSONB,
    fix_steps                JSONB,
    regression_risk          VARCHAR(20),
    point                    SMALLINT CHECK (point BETWEEN 0 AND 12),
    estimate_point_id        BIGINT,
    sprint_id                BIGINT,
    version_id               BIGINT,
    progress                 SMALLINT DEFAULT 0 CHECK (progress BETWEEN 0 AND 100),
    start_date               DATE,
    target_date              DATE,
    completed_at             TIMESTAMPTZ,
    is_draft                 BOOLEAN DEFAULT false,
    archived_at              TIMESTAMPTZ,
    sort_order               DOUBLE PRECISION DEFAULT 65535,
    version                  INTEGER DEFAULT 1,
    status                   work_item_status NOT NULL DEFAULT 'active',
    deleted                  BOOLEAN DEFAULT false,
    created_by               BIGINT NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ DEFAULT now(),
    updated_by               BIGINT NOT NULL DEFAULT 0,
    updated_at               TIMESTAMPTZ DEFAULT now()
);

--  14. task_assignees — 任务执行人
CREATE TABLE IF NOT EXISTS task_assignees (
    id                       BIGINT PRIMARY KEY,
    tenant_id                BIGINT NOT NULL DEFAULT 1,
    workspace_id             BIGINT NOT NULL DEFAULT 0,
    project_id               BIGINT NOT NULL DEFAULT 0,
    task_id                  BIGINT NOT NULL,
    user_id                  BIGINT NOT NULL,
    status                   entity_status NOT NULL DEFAULT 'active',
    deleted                  BOOLEAN DEFAULT false,
    created_by               BIGINT NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ DEFAULT now(),
    updated_by               BIGINT NOT NULL DEFAULT 0,
    updated_at               TIMESTAMPTZ DEFAULT now()
);

--  15. task_labels — 任务标签
CREATE TABLE IF NOT EXISTS task_labels (
    id                       BIGINT PRIMARY KEY,
    tenant_id                BIGINT NOT NULL DEFAULT 1,
    workspace_id             BIGINT NOT NULL DEFAULT 0,
    project_id               BIGINT NOT NULL DEFAULT 0,
    task_id                  BIGINT NOT NULL,
    label_id                 BIGINT NOT NULL,
    status                   entity_status NOT NULL DEFAULT 'active',
    deleted                  BOOLEAN DEFAULT false,
    created_by               BIGINT NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ DEFAULT now(),
    updated_by               BIGINT NOT NULL DEFAULT 0,
    updated_at               TIMESTAMPTZ DEFAULT now()
);

--  16. task_modules — 任务模块关联
CREATE TABLE IF NOT EXISTS task_modules (
    id                       BIGINT PRIMARY KEY,
    tenant_id                BIGINT NOT NULL DEFAULT 1,
    workspace_id             BIGINT NOT NULL DEFAULT 0,
    project_id               BIGINT NOT NULL DEFAULT 0,
    task_id                  BIGINT NOT NULL,
    module_id                BIGINT NOT NULL,
    status                   entity_status NOT NULL DEFAULT 'active',
    deleted                  BOOLEAN DEFAULT false,
    created_by               BIGINT NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ DEFAULT now(),
    updated_by               BIGINT NOT NULL DEFAULT 0,
    updated_at               TIMESTAMPTZ DEFAULT now()
);

--  17. task_watchers — 任务关注人
CREATE TABLE IF NOT EXISTS task_watchers (
    id                       BIGINT PRIMARY KEY,
    tenant_id                BIGINT NOT NULL DEFAULT 1,
    workspace_id             BIGINT NOT NULL DEFAULT 0,
    project_id               BIGINT NOT NULL DEFAULT 0,
    task_id                  BIGINT NOT NULL,
    user_id                  BIGINT NOT NULL,
    status                   entity_status NOT NULL DEFAULT 'active',
    deleted                  BOOLEAN DEFAULT false,
    created_by               BIGINT NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ DEFAULT now(),
    updated_by               BIGINT NOT NULL DEFAULT 0,
    updated_at               TIMESTAMPTZ DEFAULT now()
);

--  18. task_relations — 任务关联关系
CREATE TABLE IF NOT EXISTS task_relations (
    id                       BIGINT PRIMARY KEY,
    tenant_id                BIGINT NOT NULL DEFAULT 1,
    workspace_id             BIGINT NOT NULL DEFAULT 0,
    project_id               BIGINT NOT NULL DEFAULT 0,
    source_task_id           BIGINT NOT NULL,
    target_task_id           BIGINT NOT NULL,
    relation_type            VARCHAR(50) NOT NULL,
    status                   entity_status NOT NULL DEFAULT 'active',
    deleted                  BOOLEAN DEFAULT false,
    created_by               BIGINT NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ DEFAULT now(),
    updated_by               BIGINT NOT NULL DEFAULT 0,
    updated_at               TIMESTAMPTZ DEFAULT now()
);

--  19. task_comments — 任务评论
CREATE TABLE IF NOT EXISTS task_comments (
    id                       BIGINT PRIMARY KEY,
    tenant_id                BIGINT NOT NULL DEFAULT 1,
    workspace_id             BIGINT NOT NULL DEFAULT 0,
    project_id               BIGINT NOT NULL DEFAULT 0,
    task_id                  BIGINT NOT NULL,
    content_json             JSONB NOT NULL,
    content_html             TEXT NOT NULL,
    content_stripped         TEXT,
    parent_id                BIGINT,
    mentions                 JSONB DEFAULT '[]',
    is_edited                BOOLEAN DEFAULT false,
    edited_at                TIMESTAMPTZ,
    status                   entity_status NOT NULL DEFAULT 'active',
    deleted                  BOOLEAN DEFAULT false,
    created_by               BIGINT NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ DEFAULT now(),
    updated_by               BIGINT,
    updated_at               TIMESTAMPTZ DEFAULT now()
);

--  20. task_activities — 任务活动日志
CREATE TABLE IF NOT EXISTS task_activities (
    id                       BIGINT PRIMARY KEY,
    tenant_id                BIGINT NOT NULL DEFAULT 1,
    workspace_id             BIGINT NOT NULL DEFAULT 0,
    project_id               BIGINT NOT NULL DEFAULT 0,
    task_id                  BIGINT NOT NULL,
    verb                     VARCHAR(50) NOT NULL,
    field_name               VARCHAR(100),
    old_value                TEXT,
    new_value                TEXT,
    actor_id                 BIGINT NOT NULL,
    status                   entity_status NOT NULL DEFAULT 'active',
    deleted                  BOOLEAN DEFAULT false,
    created_by               BIGINT NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ DEFAULT now(),
    updated_by               BIGINT NOT NULL DEFAULT 0,
    updated_at               TIMESTAMPTZ DEFAULT now()
);

--  21. task_ext — 任务扩展字段
CREATE TABLE IF NOT EXISTS task_ext (
    id                       BIGINT PRIMARY KEY,
    tenant_id                BIGINT NOT NULL DEFAULT 1,
    workspace_id             BIGINT NOT NULL DEFAULT 0,
    project_id               BIGINT NOT NULL DEFAULT 0,
    task_id                  BIGINT NOT NULL,
    field_name               VARCHAR(100) NOT NULL,
    field_value              JSONB NOT NULL,
    field_schema             JSONB,
    status                   entity_status NOT NULL DEFAULT 'active',
    deleted                  BOOLEAN DEFAULT false,
    created_by               BIGINT NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ DEFAULT now(),
    updated_by               BIGINT NOT NULL DEFAULT 0,
    updated_at               TIMESTAMPTZ DEFAULT now()
);

--  22. sprints — 迭代
CREATE TABLE IF NOT EXISTS sprints (
    id                       BIGINT PRIMARY KEY,
    code                     VARCHAR(50),
    name                     VARCHAR(255) NOT NULL,
    tenant_id                BIGINT NOT NULL DEFAULT 1,
    workspace_id             BIGINT NOT NULL DEFAULT 0,
    project_id               BIGINT NOT NULL DEFAULT 0,
    description              TEXT,
    goal                     TEXT,
    start_date               DATE,
    end_date                 DATE,
    capacity                 NUMERIC(10,2),
    owner_id                 BIGINT,
    viewport                 JSONB DEFAULT '{}',
    review_snapshot          JSONB,
    started_at               TIMESTAMPTZ,
    completed_at             TIMESTAMPTZ,
    version_id               BIGINT,
    status                   sprint_status NOT NULL DEFAULT 'planned',
    version                  INTEGER DEFAULT 1,
    deleted                  BOOLEAN DEFAULT false,
    created_by               BIGINT NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ DEFAULT now(),
    updated_by               BIGINT NOT NULL DEFAULT 0,
    updated_at               TIMESTAMPTZ DEFAULT now()
);

--  23. sprint_snapshots — 迭代快照
CREATE TABLE IF NOT EXISTS sprint_snapshots (
    id                       BIGINT PRIMARY KEY,
    tenant_id                BIGINT NOT NULL DEFAULT 1,
    workspace_id             BIGINT NOT NULL DEFAULT 0,
    project_id               BIGINT NOT NULL DEFAULT 0,
    sprint_id                BIGINT NOT NULL,
    snapshot_date            DATE NOT NULL,
    data                     JSONB DEFAULT '{}',
    status                   entity_status NOT NULL DEFAULT 'active',
    deleted                  BOOLEAN DEFAULT false,
    created_by               BIGINT NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ DEFAULT now(),
    updated_by               BIGINT NOT NULL DEFAULT 0,
    updated_at               TIMESTAMPTZ DEFAULT now()
);

--  24. versions — 版本
CREATE TABLE IF NOT EXISTS versions (
    id                       BIGINT PRIMARY KEY,
    code                     VARCHAR(50),
    name                     VARCHAR(255) NOT NULL,
    tenant_id                BIGINT NOT NULL DEFAULT 1,
    workspace_id             BIGINT NOT NULL DEFAULT 0,
    project_id               BIGINT NOT NULL DEFAULT 0,
    semver                   VARCHAR(50) NOT NULL,
    description              TEXT,
    checklist                JSONB DEFAULT '[]',
    release_notes            TEXT,
    start_date               DATE,
    end_date                 DATE,
    target_date              DATE,
    delivered_at             TIMESTAMPTZ,
    archived_at              TIMESTAMPTZ,
    status                   version_status NOT NULL DEFAULT 'planning',
    version                  INTEGER DEFAULT 1,
    deleted                  BOOLEAN DEFAULT false,
    created_by               BIGINT NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ DEFAULT now(),
    updated_by               BIGINT NOT NULL DEFAULT 0,
    updated_at               TIMESTAMPTZ DEFAULT now()
);

--  25. version_delivery_snapshots — 版本交付快照
CREATE TABLE IF NOT EXISTS version_delivery_snapshots (
    id                       BIGINT PRIMARY KEY,
    tenant_id                BIGINT NOT NULL DEFAULT 1,
    workspace_id             BIGINT NOT NULL DEFAULT 0,
    project_id               BIGINT NOT NULL DEFAULT 0,
    version_id               BIGINT NOT NULL,
    progress                 JSONB DEFAULT '{}',
    quality                  JSONB DEFAULT '{}',
    release_notes            TEXT,
    snapshot_at              TIMESTAMPTZ DEFAULT now(),
    status                   entity_status NOT NULL DEFAULT 'active',
    deleted                  BOOLEAN DEFAULT false,
    created_by               BIGINT NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ DEFAULT now(),
    updated_by               BIGINT NOT NULL DEFAULT 0,
    updated_at               TIMESTAMPTZ DEFAULT now()
);

--  26. states — 状态
CREATE TABLE IF NOT EXISTS states (
    id                       BIGINT PRIMARY KEY,
    name                     VARCHAR(255) NOT NULL,
    tenant_id                BIGINT NOT NULL DEFAULT 1,
    workspace_id             BIGINT NOT NULL DEFAULT 0,
    project_id               BIGINT NOT NULL DEFAULT 0,
    "group"                    VARCHAR(50) NOT NULL,
    color                    VARCHAR(20) DEFAULT '#8DA2C2',
    sequence                 DOUBLE PRECISION DEFAULT 65535,
    is_default               BOOLEAN DEFAULT false,
    applicable_types         TEXT[] DEFAULT '{all}',
    template_set             VARCHAR(50) DEFAULT 'custom',
    status                   entity_status NOT NULL DEFAULT 'active',
    deleted                  BOOLEAN DEFAULT false,
    created_by               BIGINT NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ DEFAULT now(),
    updated_by               BIGINT NOT NULL DEFAULT 0,
    updated_at               TIMESTAMPTZ DEFAULT now()
);

--  27. state_transitions — 状态流转规则
CREATE TABLE IF NOT EXISTS state_transitions (
    id                       BIGINT PRIMARY KEY,
    tenant_id                BIGINT NOT NULL DEFAULT 1,
    workspace_id             BIGINT NOT NULL DEFAULT 0,
    project_id               BIGINT NOT NULL DEFAULT 0,
    type_code                VARCHAR(20) DEFAULT 'all',
    from_state_id            BIGINT NOT NULL,
    to_state_id              BIGINT NOT NULL,
    required_fields          JSONB DEFAULT '[]',
    status                   entity_status NOT NULL DEFAULT 'active',
    deleted                  BOOLEAN DEFAULT false,
    created_by               BIGINT NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ DEFAULT now(),
    updated_by               BIGINT NOT NULL DEFAULT 0,
    updated_at               TIMESTAMPTZ DEFAULT now()
);

--  28. modules — 模块
CREATE TABLE IF NOT EXISTS modules (
    id                       BIGINT PRIMARY KEY,
    code                     VARCHAR(50),
    name                     VARCHAR(255) NOT NULL,
    tenant_id                BIGINT NOT NULL DEFAULT 1,
    workspace_id             BIGINT NOT NULL DEFAULT 0,
    project_id               BIGINT NOT NULL DEFAULT 0,
    public_id                UUID DEFAULT gen_random_uuid(),
    description              TEXT,
    lead_id                  BIGINT,
    sort_order               DOUBLE PRECISION DEFAULT 65535,
    status                   entity_status NOT NULL DEFAULT 'active',
    deleted                  BOOLEAN DEFAULT false,
    created_by               BIGINT NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ DEFAULT now(),
    updated_by               BIGINT NOT NULL DEFAULT 0,
    updated_at               TIMESTAMPTZ DEFAULT now()
);

--  28b. issue_dependencies — 任务依赖关系（FS/SS/FF/SF）
CREATE TABLE IF NOT EXISTS issue_dependencies (
    id                       BIGINT PRIMARY KEY,
    tenant_id                BIGINT NOT NULL DEFAULT 1,
    workspace_id             BIGINT NOT NULL DEFAULT 0,
    project_id               BIGINT NOT NULL DEFAULT 0,
    issue_id                 BIGINT NOT NULL,
    depends_on_id            BIGINT NOT NULL,
    dependency_type          VARCHAR(10) NOT NULL DEFAULT 'fs',
    lag_days                 INTEGER DEFAULT 0,
    status                   entity_status NOT NULL DEFAULT 'active',
    deleted                  BOOLEAN DEFAULT false,
    created_by               BIGINT NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ DEFAULT now(),
    updated_by               BIGINT NOT NULL DEFAULT 0,
    updated_at               TIMESTAMPTZ DEFAULT now(),
    UNIQUE (issue_id, depends_on_id)
);

CREATE INDEX IF NOT EXISTS idx_issue_deps_issue_id ON issue_dependencies (issue_id, project_id) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_issue_deps_depends_on ON issue_dependencies (depends_on_id, project_id) WHERE NOT deleted;

--  29. labels — 标签
CREATE TABLE IF NOT EXISTS labels (
    id                       BIGINT PRIMARY KEY,
    code                     VARCHAR(50),
    name                     VARCHAR(255) NOT NULL,
    tenant_id                BIGINT NOT NULL DEFAULT 1,
    workspace_id             BIGINT NOT NULL DEFAULT 0,
    project_id               BIGINT NOT NULL DEFAULT 0,
    color                    VARCHAR(20),
    description              TEXT,
    status                   entity_status NOT NULL DEFAULT 'active',
    deleted                  BOOLEAN DEFAULT false,
    created_by               BIGINT NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ DEFAULT now(),
    updated_by               BIGINT NOT NULL DEFAULT 0,
    updated_at               TIMESTAMPTZ DEFAULT now()
);

--  30. estimate_points — 估算体系
CREATE TABLE IF NOT EXISTS estimate_points (
    id                       BIGINT PRIMARY KEY,
    code                     VARCHAR(50) NOT NULL,
    name                     VARCHAR(255) NOT NULL,
    tenant_id                BIGINT NOT NULL DEFAULT 1,
    workspace_id             BIGINT NOT NULL DEFAULT 0,
    description              TEXT,
    points_config            JSONB NOT NULL,
    is_default               BOOLEAN DEFAULT false,
    status                   entity_status NOT NULL DEFAULT 'active',
    deleted                  BOOLEAN DEFAULT false,
    created_by               BIGINT NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ DEFAULT now(),
    updated_by               BIGINT NOT NULL DEFAULT 0,
    updated_at               TIMESTAMPTZ DEFAULT now()
);

--  31. automation_rules — 自动化规则
CREATE TABLE IF NOT EXISTS automation_rules (
    id                       BIGINT PRIMARY KEY,
    code                     VARCHAR(50),
    name                     VARCHAR(255) NOT NULL,
    tenant_id                BIGINT NOT NULL DEFAULT 1,
    workspace_id             BIGINT NOT NULL DEFAULT 0,
    project_id               BIGINT,
    description              TEXT,
    trigger_type             VARCHAR(50) NOT NULL,
    conditions               JSONB DEFAULT '{}',
    actions                  JSONB DEFAULT '{}',
    sort_order               INTEGER DEFAULT 0,
    status                   entity_status NOT NULL DEFAULT 'active',
    deleted                  BOOLEAN DEFAULT false,
    created_by               BIGINT NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ DEFAULT now(),
    updated_by               BIGINT NOT NULL DEFAULT 0,
    updated_at               TIMESTAMPTZ DEFAULT now()
);

--  32. rule_executions — 规则执行日志
CREATE TABLE IF NOT EXISTS rule_executions (
    id                       BIGINT PRIMARY KEY,
    tenant_id                BIGINT NOT NULL DEFAULT 1,
    workspace_id             BIGINT NOT NULL DEFAULT 0,
    project_id               BIGINT,
    rule_id                  BIGINT NOT NULL,
    trigger_event_id         BIGINT,
    duration_ms              INTEGER,
    error_message            TEXT,
    context_json             JSONB,
    trigger_depth            SMALLINT DEFAULT 0,
    via_automation           BOOLEAN DEFAULT false,
    status                   work_item_status NOT NULL,
    deleted                  BOOLEAN DEFAULT false,
    created_by               BIGINT NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ DEFAULT now(),
    updated_by               BIGINT NOT NULL DEFAULT 0,
    updated_at               TIMESTAMPTZ DEFAULT now()
);

--  33. automation_templates — 自动化模板
CREATE TABLE IF NOT EXISTS automation_templates (
    id                       BIGINT PRIMARY KEY,
    code                     VARCHAR(50) NOT NULL UNIQUE,
    name                     VARCHAR(255) NOT NULL,
    description              TEXT,
    category                 VARCHAR(50),
    icon                     VARCHAR(50),
    template_config          JSONB NOT NULL,
    is_default               BOOLEAN DEFAULT false,
    sort_order               INTEGER DEFAULT 0,
    status                   entity_status NOT NULL DEFAULT 'active',
    deleted                  BOOLEAN DEFAULT false,
    created_by               BIGINT NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ DEFAULT now(),
    updated_by               BIGINT NOT NULL DEFAULT 0,
    updated_at               TIMESTAMPTZ DEFAULT now()
);

--  34. dashboards — 仪表盘
CREATE TABLE IF NOT EXISTS dashboards (
    id                       BIGINT PRIMARY KEY,
    code                     VARCHAR(50),
    name                     VARCHAR(255) NOT NULL,
    tenant_id                BIGINT NOT NULL DEFAULT 1,
    workspace_id             BIGINT NOT NULL DEFAULT 0,
    project_id               BIGINT,
    description              TEXT,
    layout                   JSONB DEFAULT '{}',
    status                   entity_status NOT NULL DEFAULT 'active',
    deleted                  BOOLEAN DEFAULT false,
    created_by               BIGINT NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ DEFAULT now(),
    updated_by               BIGINT NOT NULL DEFAULT 0,
    updated_at               TIMESTAMPTZ DEFAULT now()
);

--  35. dashboard_widgets — 仪表盘组件
CREATE TABLE IF NOT EXISTS dashboard_widgets (
    id                       BIGINT PRIMARY KEY,
    tenant_id                BIGINT NOT NULL DEFAULT 1,
    workspace_id             BIGINT NOT NULL DEFAULT 0,
    project_id               BIGINT,
    dashboard_id             BIGINT NOT NULL,
    widget_type              VARCHAR(50) NOT NULL,
    name                     VARCHAR(255) NOT NULL,
    config                   JSONB DEFAULT '{}',
    user_id                  BIGINT,
    sort_order               INTEGER DEFAULT 0,
    status                   entity_status NOT NULL DEFAULT 'active',
    deleted                  BOOLEAN DEFAULT false,
    created_by               BIGINT NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ DEFAULT now(),
    updated_by               BIGINT NOT NULL DEFAULT 0,
    updated_at               TIMESTAMPTZ DEFAULT now()
);

--  36. dashboard_snapshots — 仪表盘快照
CREATE TABLE IF NOT EXISTS dashboard_snapshots (
    id                       BIGINT PRIMARY KEY,
    tenant_id                BIGINT NOT NULL DEFAULT 1,
    workspace_id             BIGINT NOT NULL DEFAULT 0,
    project_id               BIGINT NOT NULL DEFAULT 0,
    dashboard_id             BIGINT NOT NULL,
    widget_type              VARCHAR(50) NOT NULL,
    refreshed_at             TIMESTAMPTZ DEFAULT now(),
    data                     JSONB DEFAULT '{}',
    status                   entity_status NOT NULL DEFAULT 'active',
    deleted                  BOOLEAN DEFAULT false,
    created_by               BIGINT NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ DEFAULT now(),
    updated_by               BIGINT NOT NULL DEFAULT 0,
    updated_at               TIMESTAMPTZ DEFAULT now()
);

--  37. dashboard_templates — 仪表盘模板
CREATE TABLE IF NOT EXISTS dashboard_templates (
    id                       BIGINT PRIMARY KEY,
    code                     VARCHAR(50) NOT NULL UNIQUE,
    name                     VARCHAR(255) NOT NULL,
    description              TEXT,
    category                 VARCHAR(50),
    layout                   JSONB DEFAULT '{}',
    icon                     VARCHAR(50),
    is_default               BOOLEAN DEFAULT false,
    sort_order               INTEGER DEFAULT 0,
    status                   entity_status NOT NULL DEFAULT 'active',
    deleted                  BOOLEAN DEFAULT false,
    created_by               BIGINT NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ DEFAULT now(),
    updated_by               BIGINT NOT NULL DEFAULT 0,
    updated_at               TIMESTAMPTZ DEFAULT now()
);

--  38. notifications — 通知
CREATE TABLE IF NOT EXISTS notifications (
    id                       BIGINT PRIMARY KEY,
    tenant_id                BIGINT NOT NULL DEFAULT 1,
    workspace_id             BIGINT NOT NULL DEFAULT 0,
    recipient_id             BIGINT NOT NULL,
    actor_id                 BIGINT,
    notification_type        VARCHAR(50) NOT NULL,
    title                    VARCHAR(255) NOT NULL,
    content                  TEXT,
    entity_type              VARCHAR(50),
    entity_id                BIGINT,
    is_read                  BOOLEAN DEFAULT false,
    is_archived              BOOLEAN DEFAULT false,
    read_at                  TIMESTAMPTZ,
    status                   notification_status NOT NULL DEFAULT 'pending',
    deleted                  BOOLEAN DEFAULT false,
    created_by               BIGINT NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ DEFAULT now(),
    updated_by               BIGINT NOT NULL DEFAULT 0,
    updated_at               TIMESTAMPTZ DEFAULT now()
);

--  39. notification_deliveries — 通知投递
CREATE TABLE IF NOT EXISTS notification_deliveries (
    id                       BIGINT PRIMARY KEY,
    tenant_id                BIGINT NOT NULL DEFAULT 1,
    workspace_id             BIGINT NOT NULL DEFAULT 0,
    notification_id          BIGINT NOT NULL,
    channel                  VARCHAR(20) NOT NULL,
    status                   notification_status NOT NULL DEFAULT 'pending',
    retry_count              INTEGER DEFAULT 0,
    next_retry_at            TIMESTAMPTZ,
    delivered_at             TIMESTAMPTZ,
    error_message            TEXT,
    deleted                  BOOLEAN DEFAULT false,
    created_by               BIGINT NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ DEFAULT now(),
    updated_by               BIGINT NOT NULL DEFAULT 0,
    updated_at               TIMESTAMPTZ DEFAULT now()
);

--  40. notification_preferences — 通知偏好
CREATE TABLE IF NOT EXISTS notification_preferences (
    id                       BIGINT PRIMARY KEY,
    tenant_id                BIGINT NOT NULL DEFAULT 1,
    workspace_id             BIGINT NOT NULL DEFAULT 0,
    user_id                  BIGINT NOT NULL,
    channel_settings         JSONB DEFAULT '{}',
    mute_all                 BOOLEAN DEFAULT false,
    digest_enabled           BOOLEAN DEFAULT true,
    digest_schedule          VARCHAR(20) DEFAULT 'daily',
    status                   entity_status NOT NULL DEFAULT 'active',
    deleted                  BOOLEAN DEFAULT false,
    created_by               BIGINT NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ DEFAULT now(),
    updated_by               BIGINT NOT NULL DEFAULT 0,
    updated_at               TIMESTAMPTZ DEFAULT now()
);

--  41. notification_digests — 通知摘要
CREATE TABLE IF NOT EXISTS notification_digests (
    id                       BIGINT PRIMARY KEY,
    tenant_id                BIGINT NOT NULL DEFAULT 1,
    workspace_id             BIGINT NOT NULL DEFAULT 0,
    user_id                  BIGINT NOT NULL,
    digest_type              VARCHAR(20) NOT NULL,
    content                  JSONB DEFAULT '{}',
    scheduled_for            TIMESTAMPTZ,
    sent_at                  TIMESTAMPTZ,
    status                   notification_status NOT NULL DEFAULT 'pending',
    deleted                  BOOLEAN DEFAULT false,
    created_by               BIGINT NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ DEFAULT now(),
    updated_by               BIGINT NOT NULL DEFAULT 0,
    updated_at               TIMESTAMPTZ DEFAULT now()
);

--  42. search_documents — 搜索文档索引
CREATE TABLE IF NOT EXISTS search_documents (
    id                       BIGINT PRIMARY KEY,
    tenant_id                BIGINT NOT NULL DEFAULT 1,
    workspace_id             BIGINT NOT NULL DEFAULT 0,
    project_id               BIGINT,
    doc_type                 VARCHAR(50) NOT NULL,
    doc_id                   BIGINT NOT NULL,
    title                    VARCHAR(255),
    identifier               VARCHAR(100),
    content                  TEXT,
    search_tsv               tsvector,
    metadata                 JSONB DEFAULT '{}',
    status                   entity_status NOT NULL DEFAULT 'active',
    deleted                  BOOLEAN DEFAULT false,
    created_by               BIGINT NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ DEFAULT now(),
    updated_by               BIGINT NOT NULL DEFAULT 0,
    updated_at               TIMESTAMPTZ DEFAULT now()
);

--  43. search_history — 搜索历史
CREATE TABLE IF NOT EXISTS search_history (
    id                       BIGINT PRIMARY KEY,
    tenant_id                BIGINT NOT NULL DEFAULT 1,
    workspace_id             BIGINT NOT NULL DEFAULT 0,
    user_id                  BIGINT NOT NULL,
    query                    TEXT NOT NULL,
    filters                  JSONB DEFAULT '{}',
    result_count             INTEGER DEFAULT 0,
    status                   entity_status NOT NULL DEFAULT 'active',
    deleted                  BOOLEAN DEFAULT false,
    created_by               BIGINT NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ DEFAULT now(),
    updated_by               BIGINT NOT NULL DEFAULT 0,
    updated_at               TIMESTAMPTZ DEFAULT now()
);

--  44. search_bookmarks — 搜索收藏
CREATE TABLE IF NOT EXISTS search_bookmarks (
    id                       BIGINT PRIMARY KEY,
    tenant_id                BIGINT NOT NULL DEFAULT 1,
    workspace_id             BIGINT NOT NULL DEFAULT 0,
    project_id               BIGINT,
    user_id                  BIGINT NOT NULL,
    name                     VARCHAR(255) NOT NULL,
    query                    TEXT,
    filters                  JSONB DEFAULT '{}',
    is_shared                BOOLEAN DEFAULT false,
    sort_order               DOUBLE PRECISION DEFAULT 65535,
    status                   entity_status NOT NULL DEFAULT 'active',
    deleted                  BOOLEAN DEFAULT false,
    created_by               BIGINT NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ DEFAULT now(),
    updated_by               BIGINT NOT NULL DEFAULT 0,
    updated_at               TIMESTAMPTZ DEFAULT now()
);

--  45. risk_rules — 风险规则
CREATE TABLE IF NOT EXISTS risk_rules (
    id                       BIGINT PRIMARY KEY,
    code                     VARCHAR(50),
    name                     VARCHAR(255) NOT NULL,
    tenant_id                BIGINT NOT NULL DEFAULT 1,
    workspace_id             BIGINT NOT NULL DEFAULT 0,
    project_id               BIGINT,
    rule_type                VARCHAR(50) NOT NULL,
    condition_json           JSONB DEFAULT '{}',
    notify_channels          TEXT[] DEFAULT '{}',
    is_active                BOOLEAN DEFAULT true,
    last_triggered           TIMESTAMPTZ,
    status                   entity_status NOT NULL DEFAULT 'active',
    deleted                  BOOLEAN DEFAULT false,
    created_by               BIGINT NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ DEFAULT now(),
    updated_by               BIGINT NOT NULL DEFAULT 0,
    updated_at               TIMESTAMPTZ DEFAULT now()
);

--  46. risk_alerts — 风险告警
CREATE TABLE IF NOT EXISTS risk_alerts (
    id                       BIGINT PRIMARY KEY,
    tenant_id                BIGINT NOT NULL DEFAULT 1,
    workspace_id             BIGINT NOT NULL DEFAULT 0,
    project_id               BIGINT,
    rule_id                  BIGINT NOT NULL,
    severity                 VARCHAR(20) DEFAULT 'medium',
    title                    VARCHAR(255) NOT NULL,
    description              TEXT,
    metadata                 JSONB DEFAULT '{}',
    is_resolved              BOOLEAN DEFAULT false,
    resolved_at              TIMESTAMPTZ,
    resolved_by              BIGINT,
    status                   entity_status NOT NULL DEFAULT 'active',
    deleted                  BOOLEAN DEFAULT false,
    created_by               BIGINT NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ DEFAULT now(),
    updated_by               BIGINT NOT NULL DEFAULT 0,
    updated_at               TIMESTAMPTZ DEFAULT now()
);

--  47. metric_snapshots — 指标快照
CREATE TABLE IF NOT EXISTS metric_snapshots (
    id                       BIGINT PRIMARY KEY,
    tenant_id                BIGINT NOT NULL DEFAULT 1,
    workspace_id             BIGINT NOT NULL DEFAULT 0,
    project_id               BIGINT,
    granularity              VARCHAR(20) NOT NULL,
    ref_id                   BIGINT,
    metric                   VARCHAR(50) NOT NULL,
    snapshot_date            DATE NOT NULL,
    value                    NUMERIC,
    metadata                 JSONB DEFAULT '{}',
    status                   entity_status NOT NULL DEFAULT 'active',
    deleted                  BOOLEAN DEFAULT false,
    created_by               BIGINT NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ DEFAULT now(),
    updated_by               BIGINT NOT NULL DEFAULT 0,
    updated_at               TIMESTAMPTZ DEFAULT now()
);

--  48. metric_adjustments — 指标调整
CREATE TABLE IF NOT EXISTS metric_adjustments (
    id                       BIGINT PRIMARY KEY,
    tenant_id                BIGINT NOT NULL DEFAULT 1,
    workspace_id             BIGINT NOT NULL DEFAULT 0,
    project_id               BIGINT,
    snapshot_id              BIGINT,
    metric                   VARCHAR(50) NOT NULL,
    original_value           NUMERIC,
    adjusted_value           NUMERIC,
    reason                   TEXT,
    adjusted_by              BIGINT,
    status                   entity_status NOT NULL DEFAULT 'active',
    deleted                  BOOLEAN DEFAULT false,
    created_by               BIGINT NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ DEFAULT now(),
    updated_by               BIGINT NOT NULL DEFAULT 0,
    updated_at               TIMESTAMPTZ DEFAULT now()
);

--  51. webhooks — Webhook 配置
CREATE TABLE IF NOT EXISTS webhooks (
    id                       BIGINT PRIMARY KEY,
    code                     VARCHAR(50),
    name                     VARCHAR(100) NOT NULL,
    tenant_id                BIGINT NOT NULL DEFAULT 1,
    workspace_id             BIGINT NOT NULL DEFAULT 0,
    project_id               BIGINT,
    target_url               TEXT NOT NULL,
    secret                   VARCHAR(255) NOT NULL,
    events                   TEXT[] DEFAULT '{}',
    is_active                BOOLEAN DEFAULT true,
    last_error               TEXT,
    last_triggered           TIMESTAMPTZ,
    last_status              VARCHAR(20),
    unhealthy_at             TIMESTAMPTZ,
    status                   entity_status NOT NULL DEFAULT 'active',
    deleted                  BOOLEAN DEFAULT false,
    created_by               BIGINT NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ DEFAULT now(),
    updated_by               BIGINT NOT NULL DEFAULT 0,
    updated_at               TIMESTAMPTZ DEFAULT now()
);

--  52. webhook_logs — Webhook 投递日志
CREATE TABLE IF NOT EXISTS webhook_logs (
    id                       BIGINT PRIMARY KEY,
    tenant_id                BIGINT NOT NULL DEFAULT 1,
    workspace_id             BIGINT NOT NULL DEFAULT 0,
    webhook_id               BIGINT NOT NULL,
    delivery_id              VARCHAR(64) NOT NULL,
    event_type               VARCHAR(80) NOT NULL,
    event_id                 BIGINT,
    request_url              TEXT NOT NULL,
    request_method           VARCHAR(10) DEFAULT 'POST',
    request_headers          JSONB,
    request_body             TEXT,
    response_status          INTEGER,
    response_body            TEXT,
    response_headers         JSONB,
    attempt                  SMALLINT DEFAULT 1,
    duration_ms              INTEGER,
    error                    TEXT,
    status                   VARCHAR(20) NOT NULL,
    deleted                  BOOLEAN DEFAULT false,
    created_by               BIGINT NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ DEFAULT now(),
    updated_by               BIGINT NOT NULL DEFAULT 0,
    updated_at               TIMESTAMPTZ DEFAULT now()
);

--  53. workbench_configs — 工作台配置
CREATE TABLE IF NOT EXISTS workbench_configs (
    id                       BIGINT PRIMARY KEY,
    tenant_id                BIGINT NOT NULL DEFAULT 1,
    workspace_id             BIGINT NOT NULL DEFAULT 0,
    project_id               BIGINT,
    user_id                  BIGINT NOT NULL,
    layout                   JSONB DEFAULT '{}',
    widget_states            JSONB DEFAULT '{}',
    focus_enabled            BOOLEAN DEFAULT false,
    status                   entity_status NOT NULL DEFAULT 'active',
    deleted                  BOOLEAN DEFAULT false,
    created_by               BIGINT NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ DEFAULT now(),
    updated_by               BIGINT NOT NULL DEFAULT 0,
    updated_at               TIMESTAMPTZ DEFAULT now()
);

--  54. view_preferences — 视图偏好
CREATE TABLE IF NOT EXISTS view_preferences (
    id                       BIGINT PRIMARY KEY,
    tenant_id                BIGINT NOT NULL DEFAULT 1,
    workspace_id             BIGINT NOT NULL DEFAULT 0,
    project_id               BIGINT,
    user_id                  BIGINT NOT NULL,
    view_type                VARCHAR(20) NOT NULL,
    layout                   VARCHAR(20) DEFAULT 'list',
    columns                  JSONB DEFAULT '[]',
    filters                  JSONB DEFAULT '{}',
    sort                     JSONB DEFAULT '{}',
    extra                    JSONB DEFAULT '{}',
    status                   entity_status NOT NULL DEFAULT 'active',
    deleted                  BOOLEAN DEFAULT false,
    created_by               BIGINT NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ DEFAULT now(),
    updated_by               BIGINT NOT NULL DEFAULT 0,
    updated_at               TIMESTAMPTZ DEFAULT now()
);

--  55. recent_items — 最近访问
CREATE TABLE IF NOT EXISTS recent_items (
    id                       BIGINT PRIMARY KEY,
    tenant_id                BIGINT NOT NULL DEFAULT 1,
    workspace_id             BIGINT NOT NULL DEFAULT 0,
    project_id               BIGINT,
    user_id                  BIGINT NOT NULL,
    item_type                VARCHAR(20) NOT NULL,
    item_id                  BIGINT NOT NULL,
    title                    VARCHAR(255),
    identifier               VARCHAR(100),
    accessed_at              TIMESTAMPTZ DEFAULT now(),
    status                   entity_status NOT NULL DEFAULT 'active',
    deleted                  BOOLEAN DEFAULT false,
    created_by               BIGINT NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ DEFAULT now(),
    updated_by               BIGINT NOT NULL DEFAULT 0,
    updated_at               TIMESTAMPTZ DEFAULT now()
);

--  56. knowledge_spaces — 知识空间
CREATE TABLE IF NOT EXISTS knowledge_spaces (
    id                       BIGINT PRIMARY KEY,
    code                     VARCHAR(50),
    name                     VARCHAR(255) NOT NULL,
    tenant_id                BIGINT NOT NULL DEFAULT 1,
    workspace_id             BIGINT NOT NULL DEFAULT 0,
    description              TEXT,
    icon                     VARCHAR(50),
    status                   entity_status NOT NULL DEFAULT 'active',
    deleted                  BOOLEAN DEFAULT false,
    created_by               BIGINT NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ DEFAULT now(),
    updated_by               BIGINT NOT NULL DEFAULT 0,
    updated_at               TIMESTAMPTZ DEFAULT now()
);

--  57. knowledge_pages — 知识页面
CREATE TABLE IF NOT EXISTS knowledge_pages (
    id                       BIGINT PRIMARY KEY,
    code                     VARCHAR(50),
    name                     VARCHAR(255) NOT NULL,
    tenant_id                BIGINT NOT NULL DEFAULT 1,
    workspace_id             BIGINT NOT NULL DEFAULT 0,
    knowledge_space_id       BIGINT NOT NULL,
    public_id                UUID DEFAULT gen_random_uuid(),
    parent_id                BIGINT,
    depth                    SMALLINT DEFAULT 1 CHECK (depth BETWEEN 1 AND 3),
    content_json             JSONB,
    content_html             TEXT,
    sort_order               DOUBLE PRECISION DEFAULT 65535,
    version                  INTEGER DEFAULT 1,
    status                   entity_status NOT NULL DEFAULT 'active',
    deleted                  BOOLEAN DEFAULT false,
    created_by               BIGINT NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ DEFAULT now(),
    updated_by               BIGINT NOT NULL DEFAULT 0,
    updated_at               TIMESTAMPTZ DEFAULT now()
);

--  58. sso_providers — SSO 提供方
CREATE TABLE IF NOT EXISTS sso_providers (
    id                       BIGINT PRIMARY KEY,
    name                     VARCHAR(255) NOT NULL,
    protocol                 VARCHAR(20) DEFAULT 'oidc',
    issuer_url               TEXT,
    client_id                VARCHAR(255) NOT NULL,
    client_secret            TEXT NOT NULL,
    redirect_uri             TEXT NOT NULL,
    auth_url                 TEXT,
    token_url                TEXT,
    userinfo_url             TEXT,
    jwks_url                 TEXT,
    sso_url                  TEXT,
    idp_issuer               VARCHAR(255),
    idp_certificate          TEXT,
    skip_signature           BOOLEAN DEFAULT false,
    scopes                   TEXT DEFAULT 'openid email profile',
    auto_create_user         BOOLEAN DEFAULT true,
    default_role             VARCHAR(20) DEFAULT 'member',
    attribute_mapping        JSONB DEFAULT '{}',
    enabled                  BOOLEAN DEFAULT true,
    status                   entity_status NOT NULL DEFAULT 'active',
    deleted                  BOOLEAN DEFAULT false,
    created_by               BIGINT NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ DEFAULT now(),
    updated_by               BIGINT NOT NULL DEFAULT 0,
    updated_at               TIMESTAMPTZ DEFAULT now()
);

--  59. api_tokens — API 令牌
CREATE TABLE IF NOT EXISTS api_tokens (
    id                       BIGINT PRIMARY KEY,
    tenant_id                BIGINT NOT NULL DEFAULT 1,
    user_id                  BIGINT NOT NULL,
    name                     VARCHAR(255) NOT NULL,
    token_hash               VARCHAR(255) NOT NULL UNIQUE,
    scopes                   TEXT[] DEFAULT '{}',
    expires_at               TIMESTAMPTZ,
    last_used_at             TIMESTAMPTZ,
    revoked_at               TIMESTAMPTZ,
    status                   entity_status NOT NULL DEFAULT 'active',
    deleted                  BOOLEAN DEFAULT false,
    created_by               BIGINT NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ DEFAULT now(),
    updated_by               BIGINT NOT NULL DEFAULT 0,
    updated_at               TIMESTAMPTZ DEFAULT now()
);

--  60. audit_logs — 审计日志
CREATE TABLE IF NOT EXISTS audit_logs (
    id                       BIGINT PRIMARY KEY,
    tenant_id                BIGINT NOT NULL DEFAULT 1,
    workspace_id             BIGINT,
    actor_id                 BIGINT,
    action                   VARCHAR(100) NOT NULL,
    target_type              VARCHAR(50),
    target_id                BIGINT,
    details                  JSONB DEFAULT '{}',
    ip_address               INET,
    user_agent               TEXT,
    status                   entity_status NOT NULL DEFAULT 'active',
    deleted                  BOOLEAN DEFAULT false,
    created_by               BIGINT NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ DEFAULT now(),
    updated_by               BIGINT NOT NULL DEFAULT 0,
    updated_at               TIMESTAMPTZ DEFAULT now()
);

--  61. domain_events — 领域事件
CREATE TABLE IF NOT EXISTS domain_events (
    id                       BIGINT PRIMARY KEY,
    tenant_id                BIGINT NOT NULL DEFAULT 1,
    workspace_id             BIGINT,
    event_type               VARCHAR(100) NOT NULL,
    aggregate_type           VARCHAR(50),
    aggregate_id             BIGINT,
    payload                  JSONB NOT NULL,
    metadata                 JSONB DEFAULT '{}',
    published_at             TIMESTAMPTZ,
    status                   entity_status NOT NULL DEFAULT 'active',
    deleted                  BOOLEAN DEFAULT false,
    created_by               BIGINT NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ DEFAULT now(),
    updated_by               BIGINT NOT NULL DEFAULT 0,
    updated_at               TIMESTAMPTZ DEFAULT now()
);

--  62. invitations — 邀请
CREATE TABLE IF NOT EXISTS invitations (
    id                       BIGINT PRIMARY KEY,
    tenant_id                BIGINT NOT NULL DEFAULT 1,
    workspace_id             BIGINT NOT NULL DEFAULT 0,
    email                    VARCHAR(255) NOT NULL,
    inviter_id               BIGINT NOT NULL,
    role                     workspace_role_enum NOT NULL,
    token_hash               VARCHAR(255) NOT NULL,
    message                  TEXT,
    expires_at               TIMESTAMPTZ NOT NULL,
    accepted_at              TIMESTAMPTZ,
    revoked_at               TIMESTAMPTZ,
    status                   VARCHAR(20) NOT NULL,
    deleted                  BOOLEAN DEFAULT false,
    created_by               BIGINT NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ DEFAULT now(),
    updated_by               BIGINT NOT NULL DEFAULT 0,
    updated_at               TIMESTAMPTZ DEFAULT now(),
    UNIQUE (token_hash)
);

--  63. tenant_members — 租户成员
CREATE TABLE IF NOT EXISTS tenant_members (
    id                       BIGINT PRIMARY KEY,
    tenant_id                BIGINT NOT NULL DEFAULT 1,
    user_id                  BIGINT NOT NULL,
    role                     workspace_role_enum NOT NULL DEFAULT 'member',
    status                   entity_status NOT NULL DEFAULT 'active',
    is_owner                 BOOLEAN DEFAULT false,
    joined_at                TIMESTAMPTZ DEFAULT now(),
    deleted                  BOOLEAN DEFAULT false,
    created_by               BIGINT NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ DEFAULT now(),
    updated_by               BIGINT NOT NULL DEFAULT 0,
    updated_at               TIMESTAMPTZ DEFAULT now(),
    UNIQUE (tenant_id, user_id)
);

--  64. user_preferences — 用户偏好
CREATE TABLE IF NOT EXISTS user_preferences (
    id                       BIGINT PRIMARY KEY,
    tenant_id                BIGINT NOT NULL DEFAULT 1,
    user_id                  BIGINT NOT NULL,
    key                      VARCHAR(100) NOT NULL,
    value                    JSONB,
    status                   entity_status NOT NULL DEFAULT 'active',
    deleted                  BOOLEAN DEFAULT false,
    created_by               BIGINT NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ DEFAULT now(),
    updated_by               BIGINT NOT NULL DEFAULT 0,
    updated_at               TIMESTAMPTZ DEFAULT now(),
    UNIQUE (tenant_id, user_id, key)
);

--  65. role_menus — 角色-权限关联
CREATE TABLE IF NOT EXISTS role_menus (
    id                       BIGINT PRIMARY KEY,
    role_id                  BIGINT NOT NULL,
    menu_id                  BIGINT NOT NULL,
    status                   entity_status NOT NULL DEFAULT 'active',
    deleted                  BOOLEAN DEFAULT false,
    created_by               BIGINT NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ DEFAULT now(),
    updated_by               BIGINT NOT NULL DEFAULT 0,
    updated_at               TIMESTAMPTZ DEFAULT now(),
    UNIQUE (role_id, menu_id)
);

--  66. project_configs — 项目配置
CREATE TABLE IF NOT EXISTS project_configs (
    id                       BIGINT PRIMARY KEY,
    tenant_id                BIGINT NOT NULL DEFAULT 1,
    workspace_id             BIGINT NOT NULL DEFAULT 0,
    project_id               BIGINT NOT NULL UNIQUE,
    config                   JSONB DEFAULT '{}',
    status                   entity_status NOT NULL DEFAULT 'active',
    deleted                  BOOLEAN DEFAULT false,
    created_by               BIGINT NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ DEFAULT now(),
    updated_by               BIGINT NOT NULL DEFAULT 0,
    updated_at               TIMESTAMPTZ DEFAULT now()
);

--  67. version_sprint_relations — 版本-迭代关联
CREATE TABLE IF NOT EXISTS version_sprint_relations (
    id                       BIGINT PRIMARY KEY,
    tenant_id                BIGINT NOT NULL DEFAULT 1,
    workspace_id             BIGINT NOT NULL DEFAULT 0,
    project_id               BIGINT NOT NULL DEFAULT 0,
    version_id               BIGINT NOT NULL,
    sprint_id                BIGINT NOT NULL,
    status                   entity_status NOT NULL DEFAULT 'active',
    deleted                  BOOLEAN DEFAULT false,
    created_by               BIGINT NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ DEFAULT now(),
    updated_by               BIGINT NOT NULL DEFAULT 0,
    updated_at               TIMESTAMPTZ DEFAULT now(),
    UNIQUE (project_id, version_id, sprint_id)
);

--  68. sprint_requirements — 迭代需求关联
CREATE TABLE IF NOT EXISTS sprint_requirements (
    id                       BIGINT PRIMARY KEY,
    tenant_id                BIGINT NOT NULL DEFAULT 1,
    workspace_id             BIGINT NOT NULL DEFAULT 0,
    project_id               BIGINT NOT NULL DEFAULT 0,
    sprint_id                BIGINT NOT NULL,
    requirement_id           BIGINT NOT NULL,
    added_midway             BOOLEAN DEFAULT false,
    sort_order               DOUBLE PRECISION DEFAULT 65535,
    added_by                 BIGINT,
    added_at                 TIMESTAMPTZ DEFAULT now(),
    status                   entity_status NOT NULL DEFAULT 'active',
    deleted                  BOOLEAN DEFAULT false,
    created_by               BIGINT NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ DEFAULT now(),
    updated_by               BIGINT NOT NULL DEFAULT 0,
    updated_at               TIMESTAMPTZ DEFAULT now(),
    UNIQUE (project_id, sprint_id, requirement_id)
);

--  69. sprint_tasks — 迭代任务关联
CREATE TABLE IF NOT EXISTS sprint_tasks (
    id                       BIGINT PRIMARY KEY,
    tenant_id                BIGINT NOT NULL DEFAULT 1,
    workspace_id             BIGINT NOT NULL DEFAULT 0,
    project_id               BIGINT NOT NULL DEFAULT 0,
    sprint_id                BIGINT NOT NULL,
    task_id                  BIGINT NOT NULL,
    added_midway             BOOLEAN DEFAULT false,
    sort_order               DOUBLE PRECISION DEFAULT 65535,
    added_by                 BIGINT,
    added_at                 TIMESTAMPTZ DEFAULT now(),
    status                   entity_status NOT NULL DEFAULT 'active',
    deleted                  BOOLEAN DEFAULT false,
    created_by               BIGINT NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ DEFAULT now(),
    updated_by               BIGINT NOT NULL DEFAULT 0,
    updated_at               TIMESTAMPTZ DEFAULT now(),
    UNIQUE (project_id, sprint_id, task_id)
);

--  70. sprint_defects — 迭代缺陷关联
CREATE TABLE IF NOT EXISTS sprint_defects (
    id                       BIGINT PRIMARY KEY,
    tenant_id                BIGINT NOT NULL DEFAULT 1,
    workspace_id             BIGINT NOT NULL DEFAULT 0,
    project_id               BIGINT NOT NULL DEFAULT 0,
    sprint_id                BIGINT NOT NULL,
    defect_id                BIGINT NOT NULL,
    added_midway             BOOLEAN DEFAULT false,
    sort_order               DOUBLE PRECISION DEFAULT 65535,
    added_by                 BIGINT,
    added_at                 TIMESTAMPTZ DEFAULT now(),
    status                   entity_status NOT NULL DEFAULT 'active',
    deleted                  BOOLEAN DEFAULT false,
    created_by               BIGINT NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ DEFAULT now(),
    updated_by               BIGINT NOT NULL DEFAULT 0,
    updated_at               TIMESTAMPTZ DEFAULT now(),
    UNIQUE (project_id, sprint_id, defect_id)
);

--  71. content_templates — 内容模板
CREATE TABLE IF NOT EXISTS content_templates (
    id                       BIGINT PRIMARY KEY,
    tenant_id                BIGINT NOT NULL DEFAULT 1,
    workspace_id             BIGINT NOT NULL DEFAULT 0,
    project_id               BIGINT,
    name                     VARCHAR(255) NOT NULL,
    template_type            VARCHAR(50) NOT NULL,
    content_json             JSONB NOT NULL,
    content_html             TEXT,
    is_default               BOOLEAN DEFAULT false,
    status                   entity_status NOT NULL DEFAULT 'active',
    deleted                  BOOLEAN DEFAULT false,
    created_by               BIGINT NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ DEFAULT now(),
    updated_by               BIGINT NOT NULL DEFAULT 0,
    updated_at               TIMESTAMPTZ DEFAULT now()
);

--  72. reviews — 评审
CREATE TABLE IF NOT EXISTS reviews (
    id                       BIGINT PRIMARY KEY,
    tenant_id                BIGINT NOT NULL DEFAULT 1,
    workspace_id             BIGINT NOT NULL DEFAULT 0,
    project_id               BIGINT,
    name                     VARCHAR(255) NOT NULL,
    review_type              VARCHAR(50) NOT NULL,
    entity_type              VARCHAR(50) DEFAULT 'requirement',
    entity_id                BIGINT,
    status                   entity_status NOT NULL DEFAULT 'active',
    description              TEXT,
    due_date                 DATE,
    created_date             DATE DEFAULT CURRENT_DATE,
    completed_date           DATE,
    deleted                  BOOLEAN DEFAULT false,
    created_by               BIGINT NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ DEFAULT now(),
    updated_by               BIGINT NOT NULL DEFAULT 0,
    updated_at               TIMESTAMPTZ DEFAULT now()
);

--  73. review_assignments — 评审分配
CREATE TABLE IF NOT EXISTS review_assignments (
    id                       BIGINT PRIMARY KEY,
    tenant_id                BIGINT NOT NULL DEFAULT 1,
    workspace_id             BIGINT NOT NULL DEFAULT 0,
    project_id               BIGINT,
    review_id                BIGINT NOT NULL,
    assignee_id              BIGINT NOT NULL,
    status                   entity_status NOT NULL DEFAULT 'active',
    deleted                  BOOLEAN DEFAULT false,
    created_by               BIGINT NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ DEFAULT now(),
    updated_by               BIGINT NOT NULL DEFAULT 0,
    updated_at               TIMESTAMPTZ DEFAULT now(),
    UNIQUE (review_id, assignee_id)
);

--  74. documents — 文档
CREATE TABLE IF NOT EXISTS documents (
    id                       BIGINT PRIMARY KEY,
    code                     VARCHAR(50),
    name                     VARCHAR(255) NOT NULL,
    tenant_id                BIGINT NOT NULL DEFAULT 1,
    workspace_id             BIGINT NOT NULL DEFAULT 0,
    project_id               BIGINT,
    public_id                UUID DEFAULT gen_random_uuid(),
    description              TEXT,
    cover_image_url          TEXT,
    is_published             BOOLEAN DEFAULT false,
    is_archived              BOOLEAN DEFAULT false,
    sort_order               DOUBLE PRECISION DEFAULT 65535,
    status                   entity_status NOT NULL DEFAULT 'active',
    deleted                  BOOLEAN DEFAULT false,
    created_by               BIGINT NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ DEFAULT now(),
    updated_by               BIGINT NOT NULL DEFAULT 0,
    updated_at               TIMESTAMPTZ DEFAULT now()
);

--  75. document_versions — 文档版本
CREATE TABLE IF NOT EXISTS document_versions (
    id                       BIGINT PRIMARY KEY,
    tenant_id                BIGINT NOT NULL DEFAULT 1,
    workspace_id             BIGINT NOT NULL DEFAULT 0,
    project_id               BIGINT,
    document_id              BIGINT NOT NULL,
    version_number           INTEGER NOT NULL,
    change_summary           VARCHAR(255),
    content_json             JSONB,
    content_html             TEXT,
    status                   entity_status NOT NULL DEFAULT 'active',
    deleted                  BOOLEAN DEFAULT false,
    created_by               BIGINT NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ DEFAULT now(),
    updated_by               BIGINT NOT NULL DEFAULT 0,
    updated_at               TIMESTAMPTZ DEFAULT now(),
    UNIQUE (document_id, version_number)
);

--  76. share_links — 分享链接
CREATE TABLE IF NOT EXISTS share_links (
    id                       BIGINT PRIMARY KEY,
    tenant_id                BIGINT NOT NULL DEFAULT 1,
    workspace_id             BIGINT NOT NULL DEFAULT 0,
    project_id               BIGINT,
    entity_type              VARCHAR(50) NOT NULL,
    entity_id                BIGINT NOT NULL,
    share_token              VARCHAR(255) NOT NULL,
    scope                    VARCHAR(20) NOT NULL DEFAULT 'view',
    password_hash            VARCHAR(255),
    expires_at               TIMESTAMPTZ,
    is_active                BOOLEAN DEFAULT true,
    access_count             INTEGER DEFAULT 0,
    last_accessed_at         TIMESTAMPTZ,
    status                   entity_status NOT NULL DEFAULT 'active',
    deleted                  BOOLEAN DEFAULT false,
    created_by               BIGINT NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ DEFAULT now(),
    updated_by               BIGINT NOT NULL DEFAULT 0,
    updated_at               TIMESTAMPTZ DEFAULT now(),
    UNIQUE (share_token)
);

--  77. notification_subscriptions — 通知订阅
CREATE TABLE IF NOT EXISTS notification_subscriptions (
    id                       BIGINT PRIMARY KEY,
    tenant_id                BIGINT NOT NULL DEFAULT 1,
    workspace_id             BIGINT NOT NULL DEFAULT 0,
    project_id               BIGINT,
    user_id                  BIGINT NOT NULL,
    entity_type              VARCHAR(50),
    entity_id                BIGINT,
    event_types              TEXT[] DEFAULT '{}',
    status                   entity_status NOT NULL DEFAULT 'active',
    deleted                  BOOLEAN DEFAULT false,
    created_by               BIGINT NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ DEFAULT now(),
    updated_by               BIGINT NOT NULL DEFAULT 0,
    updated_at               TIMESTAMPTZ DEFAULT now(),
    UNIQUE (tenant_id, workspace_id, user_id, entity_type, entity_id)
);

--  78. saved_views — 保存视图
CREATE TABLE IF NOT EXISTS saved_views (
    id                       BIGINT PRIMARY KEY,
    tenant_id                BIGINT NOT NULL DEFAULT 1,
    workspace_id             BIGINT NOT NULL DEFAULT 0,
    project_id               BIGINT,
    user_id                  BIGINT NOT NULL,
    name                     VARCHAR(255) NOT NULL,
    view_type                VARCHAR(20) NOT NULL,
    filters                  JSONB DEFAULT '{}',
    columns                  JSONB DEFAULT '[]',
    sort                     JSONB DEFAULT '{}',
    is_shared                BOOLEAN DEFAULT false,
    sort_order               DOUBLE PRECISION DEFAULT 65535,
    status                   entity_status NOT NULL DEFAULT 'active',
    deleted                  BOOLEAN DEFAULT false,
    created_by               BIGINT NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ DEFAULT now(),
    updated_by               BIGINT NOT NULL DEFAULT 0,
    updated_at               TIMESTAMPTZ DEFAULT now()
);

--  79. calendar_events — 日历事件
CREATE TABLE IF NOT EXISTS calendar_events (
    id                       BIGINT PRIMARY KEY,
    tenant_id                BIGINT NOT NULL DEFAULT 1,
    workspace_id             BIGINT NOT NULL DEFAULT 0,
    project_id               BIGINT,
    title                    VARCHAR(255) NOT NULL,
    description              TEXT,
    start_time               TIMESTAMPTZ NOT NULL,
    end_time                 TIMESTAMPTZ NOT NULL,
    is_all_day               BOOLEAN DEFAULT false,
    location                 VARCHAR(255),
    event_type               VARCHAR(20) NOT NULL DEFAULT 'meeting',
    source_type              VARCHAR(20),
    source_id                BIGINT,
    idempotency_key          VARCHAR(100),
    organizer_id             BIGINT,
    status                   entity_status NOT NULL DEFAULT 'active',
    version                  INTEGER DEFAULT 1,
    deleted                  BOOLEAN DEFAULT false,
    created_by               BIGINT NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ DEFAULT now(),
    updated_by               BIGINT NOT NULL DEFAULT 0,
    updated_at               TIMESTAMPTZ DEFAULT now()
);

--  80. data_jobs — 数据任务
CREATE TABLE IF NOT EXISTS data_jobs (
    id                       BIGINT PRIMARY KEY,
    tenant_id                BIGINT NOT NULL DEFAULT 1,
    workspace_id             BIGINT NOT NULL DEFAULT 0,
    project_id               BIGINT,
    job_type                 VARCHAR(50) NOT NULL,
    name                     VARCHAR(255) NOT NULL,
    parameters               JSONB DEFAULT '{}',
    progress                 INTEGER DEFAULT 0 CHECK (progress BETWEEN 0 AND 100),
    status                   work_item_status NOT NULL DEFAULT 'active',
    error_message            TEXT,
    scheduled_at             TIMESTAMPTZ,
    executed_at              TIMESTAMPTZ,
    completed_at             TIMESTAMPTZ,
    duration_ms              BIGINT,
    triggered_by             BIGINT,
    deleted                  BOOLEAN DEFAULT false,
    created_by               BIGINT NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ DEFAULT now(),
    updated_by               BIGINT NOT NULL DEFAULT 0,
    updated_at               TIMESTAMPTZ DEFAULT now()
);

--  81. task_attachments — 任务附件
CREATE TABLE IF NOT EXISTS task_attachments (
    id                       BIGINT PRIMARY KEY,
    tenant_id                BIGINT NOT NULL DEFAULT 1,
    workspace_id             BIGINT NOT NULL DEFAULT 0,
    project_id               BIGINT NOT NULL DEFAULT 0,
    task_id                  BIGINT NOT NULL,
    file_name                VARCHAR(255) NOT NULL,
    file_size                BIGINT NOT NULL,
    file_type                VARCHAR(100) NOT NULL,
    storage_path             TEXT NOT NULL,
    thumbnail_path           TEXT,
    status                   attachment_status NOT NULL DEFAULT 'available',
    deleted                  BOOLEAN DEFAULT false,
    created_by               BIGINT NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ DEFAULT now(),
    updated_by               BIGINT NOT NULL DEFAULT 0,
    updated_at               TIMESTAMPTZ DEFAULT now()
);

--  82. task_timelogs — 任务工时
CREATE TABLE IF NOT EXISTS task_timelogs (
    id                       BIGINT PRIMARY KEY,
    tenant_id                BIGINT NOT NULL DEFAULT 1,
    workspace_id             BIGINT NOT NULL DEFAULT 0,
    project_id               BIGINT NOT NULL DEFAULT 0,
    task_id                  BIGINT NOT NULL,
    user_id                  BIGINT NOT NULL,
    spent_date               DATE NOT NULL DEFAULT CURRENT_DATE,
    duration_minutes         INTEGER NOT NULL CHECK (duration_minutes > 0 AND duration_minutes <= 1440),
    description              TEXT,
    status                   entity_status NOT NULL DEFAULT 'active',
    deleted                  BOOLEAN DEFAULT false,
    created_by               BIGINT NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ DEFAULT now(),
    updated_by               BIGINT NOT NULL DEFAULT 0,
    updated_at               TIMESTAMPTZ DEFAULT now()
);

--  83. requirement_assignees — 需求执行人
CREATE TABLE IF NOT EXISTS requirement_assignees (
    id                       BIGINT PRIMARY KEY,
    tenant_id                BIGINT NOT NULL DEFAULT 1,
    workspace_id             BIGINT NOT NULL DEFAULT 0,
    project_id               BIGINT NOT NULL DEFAULT 0,
    requirement_id           BIGINT NOT NULL,
    user_id                  BIGINT NOT NULL,
    status                   entity_status NOT NULL DEFAULT 'active',
    deleted                  BOOLEAN DEFAULT false,
    created_by               BIGINT NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ DEFAULT now(),
    updated_by               BIGINT NOT NULL DEFAULT 0,
    updated_at               TIMESTAMPTZ DEFAULT now(),
    UNIQUE (requirement_id, user_id)
);

--  84. requirement_labels — 需求标签
CREATE TABLE IF NOT EXISTS requirement_labels (
    id                       BIGINT PRIMARY KEY,
    tenant_id                BIGINT NOT NULL DEFAULT 1,
    workspace_id             BIGINT NOT NULL DEFAULT 0,
    project_id               BIGINT NOT NULL DEFAULT 0,
    requirement_id           BIGINT NOT NULL,
    label_id                 BIGINT NOT NULL,
    status                   entity_status NOT NULL DEFAULT 'active',
    deleted                  BOOLEAN DEFAULT false,
    created_by               BIGINT NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ DEFAULT now(),
    updated_by               BIGINT NOT NULL DEFAULT 0,
    updated_at               TIMESTAMPTZ DEFAULT now(),
    UNIQUE (requirement_id, label_id)
);

--  85. requirement_modules — 需求模块关联
CREATE TABLE IF NOT EXISTS requirement_modules (
    id                       BIGINT PRIMARY KEY,
    tenant_id                BIGINT NOT NULL DEFAULT 1,
    workspace_id             BIGINT NOT NULL DEFAULT 0,
    project_id               BIGINT NOT NULL DEFAULT 0,
    requirement_id           BIGINT NOT NULL,
    module_id                BIGINT NOT NULL,
    status                   entity_status NOT NULL DEFAULT 'active',
    deleted                  BOOLEAN DEFAULT false,
    created_by               BIGINT NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ DEFAULT now(),
    updated_by               BIGINT NOT NULL DEFAULT 0,
    updated_at               TIMESTAMPTZ DEFAULT now(),
    UNIQUE (requirement_id, module_id)
);

--  86. requirement_watchers — 需求关注人
CREATE TABLE IF NOT EXISTS requirement_watchers (
    id                       BIGINT PRIMARY KEY,
    tenant_id                BIGINT NOT NULL DEFAULT 1,
    workspace_id             BIGINT NOT NULL DEFAULT 0,
    project_id               BIGINT NOT NULL DEFAULT 0,
    requirement_id           BIGINT NOT NULL,
    user_id                  BIGINT NOT NULL,
    status                   entity_status NOT NULL DEFAULT 'active',
    deleted                  BOOLEAN DEFAULT false,
    created_by               BIGINT NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ DEFAULT now(),
    updated_by               BIGINT NOT NULL DEFAULT 0,
    updated_at               TIMESTAMPTZ DEFAULT now(),
    UNIQUE (requirement_id, user_id)
);

--  87. requirement_relations — 需求关联关系
CREATE TABLE IF NOT EXISTS requirement_relations (
    id                       BIGINT PRIMARY KEY,
    tenant_id                BIGINT NOT NULL DEFAULT 1,
    workspace_id             BIGINT NOT NULL DEFAULT 0,
    project_id               BIGINT NOT NULL DEFAULT 0,
    source_requirement_id    BIGINT NOT NULL,
    target_requirement_id    BIGINT NOT NULL,
    relation_type            VARCHAR(50) NOT NULL,
    status                   entity_status NOT NULL DEFAULT 'active',
    deleted                  BOOLEAN DEFAULT false,
    created_by               BIGINT NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ DEFAULT now(),
    updated_by               BIGINT NOT NULL DEFAULT 0,
    updated_at               TIMESTAMPTZ DEFAULT now(),
    UNIQUE (source_requirement_id, target_requirement_id, relation_type)
);

--  88. requirement_comments — 需求评论
CREATE TABLE IF NOT EXISTS requirement_comments (
    id                       BIGINT PRIMARY KEY,
    tenant_id                BIGINT NOT NULL DEFAULT 1,
    workspace_id             BIGINT NOT NULL DEFAULT 0,
    project_id               BIGINT NOT NULL DEFAULT 0,
    requirement_id           BIGINT NOT NULL,
    content_json             JSONB NOT NULL,
    content_html             TEXT NOT NULL,
    content_stripped         TEXT,
    parent_id                BIGINT,
    mentions                 JSONB DEFAULT '[]',
    is_edited                BOOLEAN DEFAULT false,
    edited_at                TIMESTAMPTZ,
    status                   entity_status NOT NULL DEFAULT 'active',
    deleted                  BOOLEAN DEFAULT false,
    created_by               BIGINT NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ DEFAULT now(),
    updated_by               BIGINT,
    updated_at               TIMESTAMPTZ DEFAULT now()
);

--  89. requirement_activities — 需求活动日志
CREATE TABLE IF NOT EXISTS requirement_activities (
    id                       BIGINT PRIMARY KEY,
    tenant_id                BIGINT NOT NULL DEFAULT 1,
    workspace_id             BIGINT NOT NULL DEFAULT 0,
    project_id               BIGINT NOT NULL DEFAULT 0,
    requirement_id           BIGINT NOT NULL,
    verb                     VARCHAR(50) NOT NULL,
    field_name               VARCHAR(100),
    old_value                TEXT,
    new_value                TEXT,
    actor_id                 BIGINT NOT NULL,
    status                   entity_status NOT NULL DEFAULT 'active',
    deleted                  BOOLEAN DEFAULT false,
    created_by               BIGINT NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ DEFAULT now(),
    updated_by               BIGINT NOT NULL DEFAULT 0,
    updated_at               TIMESTAMPTZ DEFAULT now()
);

--  90. requirement_timelogs — 需求工时
CREATE TABLE IF NOT EXISTS requirement_timelogs (
    id                       BIGINT PRIMARY KEY,
    tenant_id                BIGINT NOT NULL DEFAULT 1,
    workspace_id             BIGINT NOT NULL DEFAULT 0,
    project_id               BIGINT NOT NULL DEFAULT 0,
    requirement_id           BIGINT NOT NULL,
    user_id                  BIGINT NOT NULL,
    spent_date               DATE NOT NULL DEFAULT CURRENT_DATE,
    duration_minutes         INTEGER NOT NULL CHECK (duration_minutes > 0 AND duration_minutes <= 1440),
    description              TEXT,
    status                   entity_status NOT NULL DEFAULT 'active',
    deleted                  BOOLEAN DEFAULT false,
    created_by               BIGINT NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ DEFAULT now(),
    updated_by               BIGINT NOT NULL DEFAULT 0,
    updated_at               TIMESTAMPTZ DEFAULT now()
);

--  91. requirement_attachments — 需求附件
CREATE TABLE IF NOT EXISTS requirement_attachments (
    id                       BIGINT PRIMARY KEY,
    tenant_id                BIGINT NOT NULL DEFAULT 1,
    workspace_id             BIGINT NOT NULL DEFAULT 0,
    project_id               BIGINT NOT NULL DEFAULT 0,
    requirement_id           BIGINT NOT NULL,
    file_name                VARCHAR(255) NOT NULL,
    file_size                BIGINT NOT NULL,
    file_type                VARCHAR(100) NOT NULL,
    storage_path             TEXT NOT NULL,
    thumbnail_path           TEXT,
    status                   attachment_status NOT NULL DEFAULT 'available',
    deleted                  BOOLEAN DEFAULT false,
    created_by               BIGINT NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ DEFAULT now(),
    updated_by               BIGINT NOT NULL DEFAULT 0,
    updated_at               TIMESTAMPTZ DEFAULT now()
);

--  92. requirement_ext — 需求扩展字段
CREATE TABLE IF NOT EXISTS requirement_ext (
    id                       BIGINT PRIMARY KEY,
    tenant_id                BIGINT NOT NULL DEFAULT 1,
    workspace_id             BIGINT NOT NULL DEFAULT 0,
    project_id               BIGINT NOT NULL DEFAULT 0,
    requirement_id           BIGINT NOT NULL,
    field_name               VARCHAR(100) NOT NULL,
    field_value              JSONB NOT NULL,
    field_schema             JSONB,
    status                   entity_status NOT NULL DEFAULT 'active',
    deleted                  BOOLEAN DEFAULT false,
    created_by               BIGINT NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ DEFAULT now(),
    updated_by               BIGINT NOT NULL DEFAULT 0,
    updated_at               TIMESTAMPTZ DEFAULT now(),
    UNIQUE (requirement_id, field_name)
);

--  93. defect_assignees — 缺陷处理人
CREATE TABLE IF NOT EXISTS defect_assignees (
    id                       BIGINT PRIMARY KEY,
    tenant_id                BIGINT NOT NULL DEFAULT 1,
    workspace_id             BIGINT NOT NULL DEFAULT 0,
    project_id               BIGINT NOT NULL DEFAULT 0,
    defect_id                BIGINT NOT NULL,
    user_id                  BIGINT NOT NULL,
    status                   entity_status NOT NULL DEFAULT 'active',
    deleted                  BOOLEAN DEFAULT false,
    created_by               BIGINT NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ DEFAULT now(),
    updated_by               BIGINT NOT NULL DEFAULT 0,
    updated_at               TIMESTAMPTZ DEFAULT now(),
    UNIQUE (defect_id, user_id)
);

--  94. defect_labels — 缺陷标签
CREATE TABLE IF NOT EXISTS defect_labels (
    id                       BIGINT PRIMARY KEY,
    tenant_id                BIGINT NOT NULL DEFAULT 1,
    workspace_id             BIGINT NOT NULL DEFAULT 0,
    project_id               BIGINT NOT NULL DEFAULT 0,
    defect_id                BIGINT NOT NULL,
    label_id                 BIGINT NOT NULL,
    status                   entity_status NOT NULL DEFAULT 'active',
    deleted                  BOOLEAN DEFAULT false,
    created_by               BIGINT NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ DEFAULT now(),
    updated_by               BIGINT NOT NULL DEFAULT 0,
    updated_at               TIMESTAMPTZ DEFAULT now(),
    UNIQUE (defect_id, label_id)
);

--  95. defect_modules — 缺陷模块关联
CREATE TABLE IF NOT EXISTS defect_modules (
    id                       BIGINT PRIMARY KEY,
    tenant_id                BIGINT NOT NULL DEFAULT 1,
    workspace_id             BIGINT NOT NULL DEFAULT 0,
    project_id               BIGINT NOT NULL DEFAULT 0,
    defect_id                BIGINT NOT NULL,
    module_id                BIGINT NOT NULL,
    status                   entity_status NOT NULL DEFAULT 'active',
    deleted                  BOOLEAN DEFAULT false,
    created_by               BIGINT NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ DEFAULT now(),
    updated_by               BIGINT NOT NULL DEFAULT 0,
    updated_at               TIMESTAMPTZ DEFAULT now(),
    UNIQUE (defect_id, module_id)
);

--  96. defect_watchers — 缺陷关注人
CREATE TABLE IF NOT EXISTS defect_watchers (
    id                       BIGINT PRIMARY KEY,
    tenant_id                BIGINT NOT NULL DEFAULT 1,
    workspace_id             BIGINT NOT NULL DEFAULT 0,
    project_id               BIGINT NOT NULL DEFAULT 0,
    defect_id                BIGINT NOT NULL,
    user_id                  BIGINT NOT NULL,
    status                   entity_status NOT NULL DEFAULT 'active',
    deleted                  BOOLEAN DEFAULT false,
    created_by               BIGINT NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ DEFAULT now(),
    updated_by               BIGINT NOT NULL DEFAULT 0,
    updated_at               TIMESTAMPTZ DEFAULT now(),
    UNIQUE (defect_id, user_id)
);

--  97. defect_relations — 缺陷关联关系
CREATE TABLE IF NOT EXISTS defect_relations (
    id                       BIGINT PRIMARY KEY,
    tenant_id                BIGINT NOT NULL DEFAULT 1,
    workspace_id             BIGINT NOT NULL DEFAULT 0,
    project_id               BIGINT NOT NULL DEFAULT 0,
    source_defect_id         BIGINT NOT NULL,
    target_defect_id         BIGINT NOT NULL,
    relation_type            VARCHAR(50) NOT NULL,
    status                   entity_status NOT NULL DEFAULT 'active',
    deleted                  BOOLEAN DEFAULT false,
    created_by               BIGINT NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ DEFAULT now(),
    updated_by               BIGINT NOT NULL DEFAULT 0,
    updated_at               TIMESTAMPTZ DEFAULT now(),
    UNIQUE (source_defect_id, target_defect_id, relation_type)
);

--  98. defect_comments — 缺陷评论
CREATE TABLE IF NOT EXISTS defect_comments (
    id                       BIGINT PRIMARY KEY,
    tenant_id                BIGINT NOT NULL DEFAULT 1,
    workspace_id             BIGINT NOT NULL DEFAULT 0,
    project_id               BIGINT NOT NULL DEFAULT 0,
    defect_id                BIGINT NOT NULL,
    content_json             JSONB NOT NULL,
    content_html             TEXT NOT NULL,
    content_stripped         TEXT,
    parent_id                BIGINT,
    mentions                 JSONB DEFAULT '[]',
    is_edited                BOOLEAN DEFAULT false,
    edited_at                TIMESTAMPTZ,
    status                   entity_status NOT NULL DEFAULT 'active',
    deleted                  BOOLEAN DEFAULT false,
    created_by               BIGINT NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ DEFAULT now(),
    updated_by               BIGINT,
    updated_at               TIMESTAMPTZ DEFAULT now()
);

--  99. defect_activities — 缺陷活动日志
CREATE TABLE IF NOT EXISTS defect_activities (
    id                       BIGINT PRIMARY KEY,
    tenant_id                BIGINT NOT NULL DEFAULT 1,
    workspace_id             BIGINT NOT NULL DEFAULT 0,
    project_id               BIGINT NOT NULL DEFAULT 0,
    defect_id                BIGINT NOT NULL,
    verb                     VARCHAR(50) NOT NULL,
    field_name               VARCHAR(100),
    old_value                TEXT,
    new_value                TEXT,
    actor_id                 BIGINT NOT NULL,
    status                   entity_status NOT NULL DEFAULT 'active',
    deleted                  BOOLEAN DEFAULT false,
    created_by               BIGINT NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ DEFAULT now(),
    updated_by               BIGINT NOT NULL DEFAULT 0,
    updated_at               TIMESTAMPTZ DEFAULT now()
);

-- 100. defect_timelogs — 缺陷工时
CREATE TABLE IF NOT EXISTS defect_timelogs (
    id                       BIGINT PRIMARY KEY,
    tenant_id                BIGINT NOT NULL DEFAULT 1,
    workspace_id             BIGINT NOT NULL DEFAULT 0,
    project_id               BIGINT NOT NULL DEFAULT 0,
    defect_id                BIGINT NOT NULL,
    user_id                  BIGINT NOT NULL,
    spent_date               DATE NOT NULL DEFAULT CURRENT_DATE,
    duration_minutes         INTEGER NOT NULL CHECK (duration_minutes > 0 AND duration_minutes <= 1440),
    description              TEXT,
    status                   entity_status NOT NULL DEFAULT 'active',
    deleted                  BOOLEAN DEFAULT false,
    created_by               BIGINT NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ DEFAULT now(),
    updated_by               BIGINT NOT NULL DEFAULT 0,
    updated_at               TIMESTAMPTZ DEFAULT now()
);

-- 101. defect_attachments — 缺陷附件
CREATE TABLE IF NOT EXISTS defect_attachments (
    id                       BIGINT PRIMARY KEY,
    tenant_id                BIGINT NOT NULL DEFAULT 1,
    workspace_id             BIGINT NOT NULL DEFAULT 0,
    project_id               BIGINT NOT NULL DEFAULT 0,
    defect_id                BIGINT NOT NULL,
    file_name                VARCHAR(255) NOT NULL,
    file_size                BIGINT NOT NULL,
    file_type                VARCHAR(100) NOT NULL,
    storage_path             TEXT NOT NULL,
    thumbnail_path           TEXT,
    status                   attachment_status NOT NULL DEFAULT 'available',
    deleted                  BOOLEAN DEFAULT false,
    created_by               BIGINT NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ DEFAULT now(),
    updated_by               BIGINT NOT NULL DEFAULT 0,
    updated_at               TIMESTAMPTZ DEFAULT now()
);

-- 102. defect_ext — 缺陷扩展字段
CREATE TABLE IF NOT EXISTS defect_ext (
    id                       BIGINT PRIMARY KEY,
    tenant_id                BIGINT NOT NULL DEFAULT 1,
    workspace_id             BIGINT NOT NULL DEFAULT 0,
    project_id               BIGINT NOT NULL DEFAULT 0,
    defect_id                BIGINT NOT NULL,
    field_name               VARCHAR(100) NOT NULL,
    field_value              JSONB NOT NULL,
    field_schema             JSONB,
    status                   entity_status NOT NULL DEFAULT 'active',
    deleted                  BOOLEAN DEFAULT false,
    created_by               BIGINT NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ DEFAULT now(),
    updated_by               BIGINT NOT NULL DEFAULT 0,
    updated_at               TIMESTAMPTZ DEFAULT now(),
    UNIQUE (defect_id, field_name)
);

-- 103. biz_entity_relations — 跨类型实体关联
CREATE TABLE IF NOT EXISTS biz_entity_relations (
    id                       BIGINT PRIMARY KEY,
    tenant_id                BIGINT NOT NULL DEFAULT 1,
    workspace_id             BIGINT NOT NULL DEFAULT 0,
    project_id               BIGINT NOT NULL DEFAULT 0,
    source_type              VARCHAR(20) NOT NULL,
    source_id                BIGINT NOT NULL,
    target_type              VARCHAR(20) NOT NULL,
    target_id                BIGINT NOT NULL,
    relation_type            VARCHAR(50) NOT NULL,
    status                   entity_status NOT NULL DEFAULT 'active',
    deleted                  BOOLEAN DEFAULT false,
    created_by               BIGINT NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ DEFAULT now(),
    updated_by               BIGINT NOT NULL DEFAULT 0,
    updated_at               TIMESTAMPTZ DEFAULT now(),
    UNIQUE (source_type, source_id, target_type, target_id, relation_type)
);

-- 104. pages — 项目文档页面
CREATE TABLE IF NOT EXISTS pages (
    id                       BIGINT PRIMARY KEY,
    public_id                UUID NOT NULL DEFAULT gen_random_uuid(),
    tenant_id                BIGINT NOT NULL DEFAULT 1,
    workspace_id             BIGINT NOT NULL DEFAULT 0,
    project_id               BIGINT NOT NULL DEFAULT 0,
    name                     VARCHAR(255) NOT NULL,
    description_json         JSONB,
    description_html         TEXT,
    description_stripped     TEXT,
    parent_id                BIGINT,
    sort_order               DOUBLE PRECISION NOT NULL DEFAULT 65535,
    version                  INTEGER NOT NULL DEFAULT 1,
    status                   entity_status NOT NULL DEFAULT 'active',
    deleted                  BOOLEAN DEFAULT false,
    created_by               BIGINT NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ DEFAULT now(),
    updated_by               BIGINT NOT NULL DEFAULT 0,
    updated_at               TIMESTAMPTZ DEFAULT now()
);

-- 105. deployment_events — 部署事件
CREATE TABLE IF NOT EXISTS deployment_events (
    id                       BIGINT PRIMARY KEY,
    tenant_id                BIGINT NOT NULL DEFAULT 1,
    workspace_id             BIGINT NOT NULL DEFAULT 0,
    project_id               BIGINT,
    deployment_id            VARCHAR(64),
    env                      VARCHAR(20) NOT NULL,
    status                   VARCHAR(20) NOT NULL,
    version                  VARCHAR(100),
    deployed_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted                  BOOLEAN DEFAULT false,
    created_by               BIGINT NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ DEFAULT now(),
    updated_by               BIGINT NOT NULL DEFAULT 0,
    updated_at               TIMESTAMPTZ DEFAULT now(),
    UNIQUE (deployment_id, env, project_id)
);

-- 106. processed_events — 事件消费记录
CREATE TABLE IF NOT EXISTS processed_events (
    event_id                 BIGINT NOT NULL,
    consumer_id              VARCHAR(100) NOT NULL,
    processed_at             TIMESTAMPTZ NOT NULL DEFAULT now(),
    retry_count              INTEGER NOT NULL DEFAULT 1,
    PRIMARY KEY (event_id, consumer_id)
);

-- 107. dlq_events — 死信队列事件
CREATE TABLE IF NOT EXISTS dlq_events (
    id                       BIGINT PRIMARY KEY,
    event_id                 BIGINT,
    tenant_id                BIGINT,
    workspace_id             BIGINT,
    queue                    VARCHAR(100) NOT NULL,
    exchange                 VARCHAR(100) NOT NULL,
    routing_key              VARCHAR(200) DEFAULT '',
    payload                  JSONB,
    error_reason             TEXT,
    resolved_at              TIMESTAMPTZ,
    resolved_by              VARCHAR(100),
    created_at               TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- 108. password_reset_tokens — 密码重置令牌
CREATE TABLE IF NOT EXISTS password_reset_tokens (
    id                       BIGINT PRIMARY KEY,
    user_id                  BIGINT NOT NULL,
    token_hash               VARCHAR(255) NOT NULL,
    expires_at               TIMESTAMPTZ NOT NULL,
    used_at                  TIMESTAMPTZ,
    status                   entity_status NOT NULL DEFAULT 'active',
    deleted                  BOOLEAN DEFAULT false,
    created_at               TIMESTAMPTZ DEFAULT now(),
    UNIQUE (token_hash)
);

-- 109. idempotency_keys — API 幂等键
CREATE TABLE IF NOT EXISTS idempotency_keys (
    key                      VARCHAR(255) PRIMARY KEY,
    user_id                  BIGINT,
    response_status          INTEGER,
    response_body            JSONB,
    expires_at               TIMESTAMPTZ NOT NULL,
    created_at               TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- 110. schema_migrations — 迁移版本
CREATE TABLE IF NOT EXISTS schema_migrations (
    version                  BIGINT PRIMARY KEY,
    dirty                    BOOLEAN NOT NULL DEFAULT false
);

-- ===========================================================================
-- 表/字段注释
-- ===========================================================================

COMMENT ON TABLE tenants IS '租户（组织机构） (主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN tenants.id IS '雪花ID';
COMMENT ON COLUMN tenants.code IS '租户编码（如 MEITUAN）';
COMMENT ON COLUMN tenants.name IS '组织名称（如"美团基础研发平台"）';
COMMENT ON COLUMN tenants.slug IS 'URL 标识';
COMMENT ON COLUMN tenants.logo_url IS '组织 Logo';
COMMENT ON COLUMN tenants.owner_id IS '租户 Owner';
COMMENT ON COLUMN tenants.timezone IS '默认时区';
COMMENT ON COLUMN tenants.language IS '默认语言';
COMMENT ON COLUMN tenants.brand_config IS '品牌定制（主题色/登录页）';
COMMENT ON COLUMN tenants.status IS 'active/disabled/archived/expired';
COMMENT ON COLUMN tenants.max_projects IS '最大项目数';
COMMENT ON COLUMN tenants.max_users IS '最大用户数';
COMMENT ON COLUMN tenants.max_workspaces IS '最大工作空间数';
COMMENT ON COLUMN tenants.expired_at IS '服务到期时间';
COMMENT ON COLUMN tenants.config IS '租户级配置';
COMMENT ON COLUMN tenants.created_at IS '创建时间';
COMMENT ON COLUMN tenants.updated_at IS '更新时间';

COMMENT ON TABLE workspaces IS '工作空间 (主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN workspaces.id IS '雪花ID';
COMMENT ON COLUMN workspaces.code IS '工作空间编码';
COMMENT ON COLUMN workspaces.name IS '工作空间名称';
COMMENT ON COLUMN workspaces.slug IS 'URL 友好标识';
COMMENT ON COLUMN workspaces.tenant_id IS '所属租户';
COMMENT ON COLUMN workspaces.logo_url IS 'Logo';
COMMENT ON COLUMN workspaces.owner_id IS '空间 Owner';
COMMENT ON COLUMN workspaces.timezone IS '时区';
COMMENT ON COLUMN workspaces.language IS '语言';
COMMENT ON COLUMN workspaces.status IS '状态';
COMMENT ON COLUMN workspaces.max_projects IS '空间级项目上限';
COMMENT ON COLUMN workspaces.config IS '空间级配置';
COMMENT ON COLUMN workspaces.deleted IS '软删除';
COMMENT ON COLUMN workspaces.created_by IS '创建人';
COMMENT ON COLUMN workspaces.created_at IS '创建时间';
COMMENT ON COLUMN workspaces.updated_by IS '更新人';
COMMENT ON COLUMN workspaces.updated_at IS '更新时间';

COMMENT ON TABLE workspace_members IS '工作空间成员 (主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN workspace_members.id IS '雪花ID';
COMMENT ON COLUMN workspace_members.tenant_id IS '租户ID';
COMMENT ON COLUMN workspace_members.workspace_id IS '工作空间ID';
COMMENT ON COLUMN workspace_members.user_id IS '用户ID';
COMMENT ON COLUMN workspace_members.role IS 'owner/admin/member/guest';
COMMENT ON COLUMN workspace_members.status IS '状态';
COMMENT ON COLUMN workspace_members.joined_at IS '加入时间';
COMMENT ON COLUMN workspace_members.deleted IS '软删除';
COMMENT ON COLUMN workspace_members.created_by IS '创建人';
COMMENT ON COLUMN workspace_members.created_at IS '创建时间';
COMMENT ON COLUMN workspace_members.updated_by IS '更新人';
COMMENT ON COLUMN workspace_members.updated_at IS '更新时间';

COMMENT ON TABLE users IS '用户 (主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN users.id IS '雪花ID';
COMMENT ON COLUMN users.code IS '工号/用户编码';
COMMENT ON COLUMN users.name IS '显示名';
COMMENT ON COLUMN users.email IS '邮箱';
COMMENT ON COLUMN users.phone IS '手机号';
COMMENT ON COLUMN users.password_hash IS '密码哈希';
COMMENT ON COLUMN users.avatar_url IS '头像URL';
COMMENT ON COLUMN users.status IS '状态';
COMMENT ON COLUMN users.is_super_admin IS '系统级超管';
COMMENT ON COLUMN users.last_login_at IS '最后登录时间';
COMMENT ON COLUMN users.tenant_id IS '归属租户';
COMMENT ON COLUMN users.deleted IS '软删除';
COMMENT ON COLUMN users.created_by IS '创建人';
COMMENT ON COLUMN users.created_at IS '创建时间';
COMMENT ON COLUMN users.updated_by IS '更新人';
COMMENT ON COLUMN users.updated_at IS '更新时间';

COMMENT ON TABLE roles IS '角色 (主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN roles.id IS '雪花ID';
COMMENT ON COLUMN roles.code IS '角色编码';
COMMENT ON COLUMN roles.name IS '角色名称';
COMMENT ON COLUMN roles.description IS '描述';
COMMENT ON COLUMN roles.status IS '状态';
COMMENT ON COLUMN roles.is_system IS '系统内置角色';
COMMENT ON COLUMN roles.role_scope IS '作用范围';
COMMENT ON COLUMN roles.deleted IS '软删除';
COMMENT ON COLUMN roles.created_by IS '创建人';
COMMENT ON COLUMN roles.created_at IS '创建时间';
COMMENT ON COLUMN roles.updated_by IS '更新人';
COMMENT ON COLUMN roles.updated_at IS '更新时间';

COMMENT ON TABLE menus IS '菜单 / 权限资源 (主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN menus.id IS '雪花ID';
COMMENT ON COLUMN menus.code IS '权限编码';
COMMENT ON COLUMN menus.name IS '菜单/按钮名称';
COMMENT ON COLUMN menus.menu_type IS '资源类型';
COMMENT ON COLUMN menus.parent_id IS '父菜单';
COMMENT ON COLUMN menus.path IS '路由路径';
COMMENT ON COLUMN menus.icon IS '图标';
COMMENT ON COLUMN menus.sort_order IS '排序';
COMMENT ON COLUMN menus.status IS '状态';
COMMENT ON COLUMN menus.deleted IS '软删除';
COMMENT ON COLUMN menus.created_by IS '创建人';
COMMENT ON COLUMN menus.created_at IS '创建时间';
COMMENT ON COLUMN menus.updated_by IS '更新人';
COMMENT ON COLUMN menus.updated_at IS '更新时间';

COMMENT ON TABLE user_roles IS '用户角色关联 (主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN user_roles.id IS '雪花ID';
COMMENT ON COLUMN user_roles.tenant_id IS '租户ID';
COMMENT ON COLUMN user_roles.user_id IS '用户ID';
COMMENT ON COLUMN user_roles.role_id IS '角色ID';
COMMENT ON COLUMN user_roles.status IS '状态';
COMMENT ON COLUMN user_roles.deleted IS '软删除';
COMMENT ON COLUMN user_roles.created_by IS '创建人';
COMMENT ON COLUMN user_roles.created_at IS '创建时间';
COMMENT ON COLUMN user_roles.updated_by IS '更新人';
COMMENT ON COLUMN user_roles.updated_at IS '创建时间';

COMMENT ON TABLE projects IS '项目 (主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN projects.id IS '雪花ID';
COMMENT ON COLUMN projects.code IS '项目编码';
COMMENT ON COLUMN projects.name IS '项目名称';
COMMENT ON COLUMN projects.tenant_id IS '租户ID';
COMMENT ON COLUMN projects.workspace_id IS '工作空间ID';
COMMENT ON COLUMN projects.identifier IS '项目标识符（大写 2-10 字符）';
COMMENT ON COLUMN projects.slug IS 'URL 友好标识';
COMMENT ON COLUMN projects.description IS '项目描述';
COMMENT ON COLUMN projects.icon IS '项目图标';
COMMENT ON COLUMN projects.cover_image_url IS '封面图片';
COMMENT ON COLUMN projects.network IS '可见性';
COMMENT ON COLUMN projects.template IS '项目模板';
COMMENT ON COLUMN projects.status IS '状态';
COMMENT ON COLUMN projects.modules IS '功能模块开关';
COMMENT ON COLUMN projects.start_date IS '开始日期';
COMMENT ON COLUMN projects.target_date IS '目标日期';
COMMENT ON COLUMN projects.owner_id IS '项目负责人';
COMMENT ON COLUMN projects.version IS '乐观锁版本号';
COMMENT ON COLUMN projects.sort_order IS '排序权重';
COMMENT ON COLUMN projects.deleted IS '软删除';
COMMENT ON COLUMN projects.created_by IS '创建人';
COMMENT ON COLUMN projects.created_at IS '创建时间';
COMMENT ON COLUMN projects.updated_by IS '更新人';
COMMENT ON COLUMN projects.updated_at IS '更新时间';

COMMENT ON TABLE project_members IS '项目成员 (主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN project_members.id IS '雪花ID';
COMMENT ON COLUMN project_members.tenant_id IS '租户ID';
COMMENT ON COLUMN project_members.workspace_id IS '工作空间ID';
COMMENT ON COLUMN project_members.project_id IS '项目ID';
COMMENT ON COLUMN project_members.user_id IS '用户ID';
COMMENT ON COLUMN project_members.role IS '角色';
COMMENT ON COLUMN project_members.joined_at IS '加入时间';
COMMENT ON COLUMN project_members.status IS '状态';
COMMENT ON COLUMN project_members.deleted IS '软删除';
COMMENT ON COLUMN project_members.created_by IS '创建人';
COMMENT ON COLUMN project_members.created_at IS '创建时间';
COMMENT ON COLUMN project_members.updated_by IS '更新人';
COMMENT ON COLUMN project_members.updated_at IS '更新时间';

COMMENT ON TABLE project_sequences IS '项目序列发号器 (主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN project_sequences.id IS '雪花ID';
COMMENT ON COLUMN project_sequences.tenant_id IS '租户ID';
COMMENT ON COLUMN project_sequences.workspace_id IS '工作空间ID';
COMMENT ON COLUMN project_sequences.project_id IS '项目ID';
COMMENT ON COLUMN project_sequences.sequence_id IS '当前序号';
COMMENT ON COLUMN project_sequences.status IS '状态';
COMMENT ON COLUMN project_sequences.deleted IS '软删除';
COMMENT ON COLUMN project_sequences.created_by IS '创建人';
COMMENT ON COLUMN project_sequences.created_at IS '创建时间';
COMMENT ON COLUMN project_sequences.updated_by IS '更新人';
COMMENT ON COLUMN project_sequences.updated_at IS '更新时间';

COMMENT ON TABLE task IS '任务 (主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN task.id IS '雪花ID';
COMMENT ON COLUMN task.code IS '任务编码';
COMMENT ON COLUMN task.name IS '任务名称';
COMMENT ON COLUMN task.tenant_id IS '租户ID';
COMMENT ON COLUMN task.workspace_id IS '工作空间ID';
COMMENT ON COLUMN task.project_id IS '项目ID';
COMMENT ON COLUMN task.sequence_id IS '序列号';
COMMENT ON COLUMN task.public_id IS '公开ID';
COMMENT ON COLUMN task.parent_id IS '父任务ID';
COMMENT ON COLUMN task.depth IS '层级深度';
COMMENT ON COLUMN task.description_json IS '描述（JSON）';
COMMENT ON COLUMN task.description_html IS '描述（HTML）';
COMMENT ON COLUMN task.description_stripped IS '描述（纯文本）';
COMMENT ON COLUMN task.state_id IS '状态ID';
COMMENT ON COLUMN task.priority IS '优先级';
COMMENT ON COLUMN task.category IS '分类';
COMMENT ON COLUMN task.actual_effort IS '实际工时';
COMMENT ON COLUMN task.remaining_effort IS '剩余工时';
COMMENT ON COLUMN task.delay_reason IS '延期原因';
COMMENT ON COLUMN task.point IS '故事点';
COMMENT ON COLUMN task.estimate_point_id IS '估算ID';
COMMENT ON COLUMN task.sprint_id IS '迭代ID';
COMMENT ON COLUMN task.version_id IS '版本ID';
COMMENT ON COLUMN task.progress IS '进度';
COMMENT ON COLUMN task.start_date IS '开始日期';
COMMENT ON COLUMN task.target_date IS '目标日期';
COMMENT ON COLUMN task.completed_at IS '完成时间';
COMMENT ON COLUMN task.is_draft IS '草稿标记';
COMMENT ON COLUMN task.archived_at IS '归档时间';
COMMENT ON COLUMN task.sort_order IS '排序权重';
COMMENT ON COLUMN task.version IS '乐观锁';
COMMENT ON COLUMN task.status IS '状态';
COMMENT ON COLUMN task.deleted IS '软删除';
COMMENT ON COLUMN task.created_by IS '创建人';
COMMENT ON COLUMN task.created_at IS '创建时间';
COMMENT ON COLUMN task.updated_by IS '更新人';
COMMENT ON COLUMN task.updated_at IS '更新时间';

COMMENT ON TABLE requirement IS '需求 (主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN requirement.id IS '雪花ID';
COMMENT ON COLUMN requirement.code IS '需求编码';
COMMENT ON COLUMN requirement.name IS '需求名称';
COMMENT ON COLUMN requirement.tenant_id IS '租户ID';
COMMENT ON COLUMN requirement.workspace_id IS '工作空间ID';
COMMENT ON COLUMN requirement.project_id IS '项目ID';
COMMENT ON COLUMN requirement.sequence_id IS '序列号';
COMMENT ON COLUMN requirement.public_id IS '公开ID';
COMMENT ON COLUMN requirement.parent_id IS '父需求ID';
COMMENT ON COLUMN requirement.depth IS '层级深度';
COMMENT ON COLUMN requirement.description_json IS '描述（JSON）';
COMMENT ON COLUMN requirement.description_html IS '描述（HTML）';
COMMENT ON COLUMN requirement.description_stripped IS '描述（纯文本）';
COMMENT ON COLUMN requirement.state_id IS '状态ID';
COMMENT ON COLUMN requirement.priority IS '优先级';
COMMENT ON COLUMN requirement.source IS '来源';
COMMENT ON COLUMN requirement.acceptance_criteria IS '验收标准';
COMMENT ON COLUMN requirement.business_value IS '业务价值';
COMMENT ON COLUMN requirement.review_status IS '评审状态';
COMMENT ON COLUMN requirement.point IS '故事点';
COMMENT ON COLUMN requirement.estimate_point_id IS '估算ID';
COMMENT ON COLUMN requirement.sprint_id IS '迭代ID';
COMMENT ON COLUMN requirement.version_id IS '版本ID';
COMMENT ON COLUMN requirement.progress IS '进度';
COMMENT ON COLUMN requirement.start_date IS '开始日期';
COMMENT ON COLUMN requirement.target_date IS '目标日期';
COMMENT ON COLUMN requirement.completed_at IS '完成时间';
COMMENT ON COLUMN requirement.is_draft IS '草稿标记';
COMMENT ON COLUMN requirement.archived_at IS '归档时间';
COMMENT ON COLUMN requirement.sort_order IS '排序权重';
COMMENT ON COLUMN requirement.version IS '乐观锁';
COMMENT ON COLUMN requirement.status IS '状态';
COMMENT ON COLUMN requirement.deleted IS '软删除';
COMMENT ON COLUMN requirement.created_by IS '创建人';
COMMENT ON COLUMN requirement.created_at IS '创建时间';
COMMENT ON COLUMN requirement.updated_by IS '更新人';
COMMENT ON COLUMN requirement.updated_at IS '更新时间';

COMMENT ON TABLE defect IS '缺陷 (主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN defect.id IS '雪花ID';
COMMENT ON COLUMN defect.code IS '缺陷编码';
COMMENT ON COLUMN defect.name IS '缺陷名称';
COMMENT ON COLUMN defect.tenant_id IS '租户ID';
COMMENT ON COLUMN defect.workspace_id IS '工作空间ID';
COMMENT ON COLUMN defect.project_id IS '项目ID';
COMMENT ON COLUMN defect.sequence_id IS '序列号';
COMMENT ON COLUMN defect.public_id IS '公开ID';
COMMENT ON COLUMN defect.parent_id IS '父缺陷ID';
COMMENT ON COLUMN defect.depth IS '层级深度';
COMMENT ON COLUMN defect.description_json IS '描述（JSON）';
COMMENT ON COLUMN defect.description_html IS '描述（HTML）';
COMMENT ON COLUMN defect.description_stripped IS '描述（纯文本）';
COMMENT ON COLUMN defect.state_id IS '状态ID';
COMMENT ON COLUMN defect.priority IS '优先级';
COMMENT ON COLUMN defect.severity IS '严重度';
COMMENT ON COLUMN defect.found_phase IS '发现阶段';
COMMENT ON COLUMN defect.found_version_id IS '发现版本';
COMMENT ON COLUMN defect.fix_version_id IS '修复版本';
COMMENT ON COLUMN defect.root_cause_category IS '根因分类';
COMMENT ON COLUMN defect.verifier_id IS '验证人';
COMMENT ON COLUMN defect.environment IS '环境信息';
COMMENT ON COLUMN defect.reproduce_steps IS '复现步骤';
COMMENT ON COLUMN defect.fix_steps IS '修复步骤';
COMMENT ON COLUMN defect.regression_risk IS '回归风险';
COMMENT ON COLUMN defect.point IS '故事点';
COMMENT ON COLUMN defect.estimate_point_id IS '估算ID';
COMMENT ON COLUMN defect.sprint_id IS '迭代ID';
COMMENT ON COLUMN defect.version_id IS '版本ID';
COMMENT ON COLUMN defect.progress IS '进度';
COMMENT ON COLUMN defect.start_date IS '开始日期';
COMMENT ON COLUMN defect.target_date IS '目标日期';
COMMENT ON COLUMN defect.completed_at IS '完成时间';
COMMENT ON COLUMN defect.is_draft IS '草稿标记';
COMMENT ON COLUMN defect.archived_at IS '归档时间';
COMMENT ON COLUMN defect.sort_order IS '排序权重';
COMMENT ON COLUMN defect.version IS '乐观锁';
COMMENT ON COLUMN defect.status IS '状态';
COMMENT ON COLUMN defect.deleted IS '软删除';
COMMENT ON COLUMN defect.created_by IS '创建人';
COMMENT ON COLUMN defect.created_at IS '创建时间';
COMMENT ON COLUMN defect.updated_by IS '更新人';
COMMENT ON COLUMN defect.updated_at IS '更新时间';

COMMENT ON TABLE task_assignees IS '任务执行人 (主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN task_assignees.id IS '雪花ID';
COMMENT ON COLUMN task_assignees.tenant_id IS '租户ID';
COMMENT ON COLUMN task_assignees.workspace_id IS '工作空间ID';
COMMENT ON COLUMN task_assignees.project_id IS '项目ID';
COMMENT ON COLUMN task_assignees.task_id IS '任务ID';
COMMENT ON COLUMN task_assignees.user_id IS '用户ID';
COMMENT ON COLUMN task_assignees.status IS '状态';
COMMENT ON COLUMN task_assignees.deleted IS '软删除';
COMMENT ON COLUMN task_assignees.created_by IS '创建人';
COMMENT ON COLUMN task_assignees.created_at IS '创建时间';
COMMENT ON COLUMN task_assignees.updated_by IS '更新人';
COMMENT ON COLUMN task_assignees.updated_at IS '更新时间';

COMMENT ON TABLE task_labels IS '任务标签 (主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN task_labels.id IS '雪花ID';
COMMENT ON COLUMN task_labels.tenant_id IS '租户ID';
COMMENT ON COLUMN task_labels.workspace_id IS '工作空间ID';
COMMENT ON COLUMN task_labels.project_id IS '项目ID';
COMMENT ON COLUMN task_labels.task_id IS '任务ID';
COMMENT ON COLUMN task_labels.label_id IS '标签ID';
COMMENT ON COLUMN task_labels.status IS '状态';
COMMENT ON COLUMN task_labels.deleted IS '软删除';
COMMENT ON COLUMN task_labels.created_by IS '创建人';
COMMENT ON COLUMN task_labels.created_at IS '创建时间';
COMMENT ON COLUMN task_labels.updated_by IS '更新人';
COMMENT ON COLUMN task_labels.updated_at IS '更新时间';

COMMENT ON TABLE task_modules IS '任务模块关联 (主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN task_modules.id IS '雪花ID';
COMMENT ON COLUMN task_modules.tenant_id IS '租户ID';
COMMENT ON COLUMN task_modules.workspace_id IS '工作空间ID';
COMMENT ON COLUMN task_modules.project_id IS '项目ID';
COMMENT ON COLUMN task_modules.task_id IS '任务ID';
COMMENT ON COLUMN task_modules.module_id IS '模块ID';
COMMENT ON COLUMN task_modules.status IS '状态';
COMMENT ON COLUMN task_modules.deleted IS '软删除';
COMMENT ON COLUMN task_modules.created_by IS '创建人';
COMMENT ON COLUMN task_modules.created_at IS '创建时间';
COMMENT ON COLUMN task_modules.updated_by IS '更新人';
COMMENT ON COLUMN task_modules.updated_at IS '更新时间';

COMMENT ON TABLE task_watchers IS '任务关注人 (主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN task_watchers.id IS '雪花ID';
COMMENT ON COLUMN task_watchers.tenant_id IS '租户ID';
COMMENT ON COLUMN task_watchers.workspace_id IS '工作空间ID';
COMMENT ON COLUMN task_watchers.project_id IS '项目ID';
COMMENT ON COLUMN task_watchers.task_id IS '任务ID';
COMMENT ON COLUMN task_watchers.user_id IS '用户ID';
COMMENT ON COLUMN task_watchers.status IS '状态';
COMMENT ON COLUMN task_watchers.deleted IS '软删除';
COMMENT ON COLUMN task_watchers.created_by IS '创建人';
COMMENT ON COLUMN task_watchers.created_at IS '创建时间';
COMMENT ON COLUMN task_watchers.updated_by IS '更新人';
COMMENT ON COLUMN task_watchers.updated_at IS '更新时间';

COMMENT ON TABLE task_relations IS '任务关联关系 (主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN task_relations.id IS '雪花ID';
COMMENT ON COLUMN task_relations.tenant_id IS '租户ID';
COMMENT ON COLUMN task_relations.workspace_id IS '工作空间ID';
COMMENT ON COLUMN task_relations.project_id IS '项目ID';
COMMENT ON COLUMN task_relations.source_task_id IS '源任务';
COMMENT ON COLUMN task_relations.target_task_id IS '目标任务';
COMMENT ON COLUMN task_relations.relation_type IS '关联类型';
COMMENT ON COLUMN task_relations.status IS '状态';
COMMENT ON COLUMN task_relations.deleted IS '软删除';
COMMENT ON COLUMN task_relations.created_by IS '创建人';
COMMENT ON COLUMN task_relations.created_at IS '创建时间';
COMMENT ON COLUMN task_relations.updated_by IS '更新人';
COMMENT ON COLUMN task_relations.updated_at IS '更新时间';

COMMENT ON TABLE task_comments IS '任务评论 (主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN task_comments.id IS '雪花ID';
COMMENT ON COLUMN task_comments.tenant_id IS '租户ID';
COMMENT ON COLUMN task_comments.workspace_id IS '工作空间ID';
COMMENT ON COLUMN task_comments.project_id IS '项目ID';
COMMENT ON COLUMN task_comments.task_id IS '任务ID';
COMMENT ON COLUMN task_comments.content_json IS '内容（JSON）';
COMMENT ON COLUMN task_comments.content_html IS '内容（HTML）';
COMMENT ON COLUMN task_comments.parent_id IS '父评论';
COMMENT ON COLUMN task_comments.status IS '状态';
COMMENT ON COLUMN task_comments.deleted IS '软删除';
COMMENT ON COLUMN task_comments.created_by IS '创建人';
COMMENT ON COLUMN task_comments.created_at IS '创建时间';
COMMENT ON COLUMN task_comments.updated_by IS '更新人';
COMMENT ON COLUMN task_comments.updated_at IS '更新时间';
COMMENT ON COLUMN task_comments.content_stripped IS '去标签后的纯文本内容';
COMMENT ON COLUMN task_comments.mentions IS '@提及的用户ID列表';
COMMENT ON COLUMN task_comments.is_edited IS '是否被编辑过';
COMMENT ON COLUMN task_comments.edited_at IS '最后编辑时间';

COMMENT ON TABLE task_activities IS '任务活动日志 (主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN task_activities.id IS '雪花ID';
COMMENT ON COLUMN task_activities.tenant_id IS '租户ID';
COMMENT ON COLUMN task_activities.workspace_id IS '工作空间ID';
COMMENT ON COLUMN task_activities.project_id IS '项目ID';
COMMENT ON COLUMN task_activities.task_id IS '任务ID';
COMMENT ON COLUMN task_activities.verb IS '动作类型';
COMMENT ON COLUMN task_activities.field_name IS '字段名';
COMMENT ON COLUMN task_activities.old_value IS '旧值';
COMMENT ON COLUMN task_activities.new_value IS '新值';
COMMENT ON COLUMN task_activities.actor_id IS '操作人';
COMMENT ON COLUMN task_activities.status IS '状态';
COMMENT ON COLUMN task_activities.deleted IS '软删除';
COMMENT ON COLUMN task_activities.created_by IS '创建人';
COMMENT ON COLUMN task_activities.created_at IS '创建时间（分区键）';
COMMENT ON COLUMN task_activities.updated_by IS '更新人';
COMMENT ON COLUMN task_activities.updated_at IS '更新时间';

COMMENT ON TABLE task_ext IS '任务扩展字段 (主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN task_ext.id IS '雪花ID';
COMMENT ON COLUMN task_ext.tenant_id IS '租户ID';
COMMENT ON COLUMN task_ext.workspace_id IS '工作空间ID';
COMMENT ON COLUMN task_ext.project_id IS '项目ID';
COMMENT ON COLUMN task_ext.task_id IS '任务ID';
COMMENT ON COLUMN task_ext.field_name IS '字段名';
COMMENT ON COLUMN task_ext.field_value IS '字段值';
COMMENT ON COLUMN task_ext.field_schema IS '字段 Schema';
COMMENT ON COLUMN task_ext.status IS '状态';
COMMENT ON COLUMN task_ext.deleted IS '软删除';
COMMENT ON COLUMN task_ext.created_by IS '创建人';
COMMENT ON COLUMN task_ext.created_at IS '创建时间';
COMMENT ON COLUMN task_ext.updated_by IS '更新人';
COMMENT ON COLUMN task_ext.updated_at IS '更新时间';

COMMENT ON TABLE sprints IS '迭代 (主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN sprints.id IS '雪花ID';
COMMENT ON COLUMN sprints.code IS '迭代编码';
COMMENT ON COLUMN sprints.name IS '迭代名称';
COMMENT ON COLUMN sprints.tenant_id IS '租户ID';
COMMENT ON COLUMN sprints.workspace_id IS '工作空间ID';
COMMENT ON COLUMN sprints.project_id IS '项目ID';
COMMENT ON COLUMN sprints.description IS '描述';
COMMENT ON COLUMN sprints.goal IS '迭代目标';
COMMENT ON COLUMN sprints.start_date IS '开始日期';
COMMENT ON COLUMN sprints.end_date IS '结束日期';
COMMENT ON COLUMN sprints.capacity IS '团队容量';
COMMENT ON COLUMN sprints.owner_id IS '负责人';
COMMENT ON COLUMN sprints.viewport IS '视口配置';
COMMENT ON COLUMN sprints.review_snapshot IS '复盘快照';
COMMENT ON COLUMN sprints.started_at IS '实际开始时间';
COMMENT ON COLUMN sprints.completed_at IS '实际完成时间';
COMMENT ON COLUMN sprints.version_id IS '关联版本';
COMMENT ON COLUMN sprints.status IS 'planned/active/completed';
COMMENT ON COLUMN sprints.version IS '乐观锁';
COMMENT ON COLUMN sprints.deleted IS '软删除';
COMMENT ON COLUMN sprints.created_by IS '创建人';
COMMENT ON COLUMN sprints.created_at IS '创建时间';
COMMENT ON COLUMN sprints.updated_by IS '更新人';
COMMENT ON COLUMN sprints.updated_at IS '更新时间';

COMMENT ON TABLE sprint_snapshots IS '迭代快照 (主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN sprint_snapshots.id IS '雪花ID';
COMMENT ON COLUMN sprint_snapshots.tenant_id IS '租户ID';
COMMENT ON COLUMN sprint_snapshots.workspace_id IS '工作空间ID';
COMMENT ON COLUMN sprint_snapshots.project_id IS '项目ID';
COMMENT ON COLUMN sprint_snapshots.sprint_id IS '迭代ID';
COMMENT ON COLUMN sprint_snapshots.snapshot_date IS '快照日期';
COMMENT ON COLUMN sprint_snapshots.data IS '快照数据';
COMMENT ON COLUMN sprint_snapshots.status IS '状态';
COMMENT ON COLUMN sprint_snapshots.deleted IS '软删除';
COMMENT ON COLUMN sprint_snapshots.created_by IS '创建人';
COMMENT ON COLUMN sprint_snapshots.created_at IS '创建时间';
COMMENT ON COLUMN sprint_snapshots.updated_by IS '更新人';
COMMENT ON COLUMN sprint_snapshots.updated_at IS '更新时间';

COMMENT ON TABLE versions IS '版本 (主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN versions.id IS '雪花ID';
COMMENT ON COLUMN versions.code IS '版本编码';
COMMENT ON COLUMN versions.name IS '版本名称';
COMMENT ON COLUMN versions.tenant_id IS '租户ID';
COMMENT ON COLUMN versions.workspace_id IS '工作空间ID';
COMMENT ON COLUMN versions.project_id IS '项目ID';
COMMENT ON COLUMN versions.semver IS '语义化版本号';
COMMENT ON COLUMN versions.description IS '描述';
COMMENT ON COLUMN versions.checklist IS '发布检查清单';
COMMENT ON COLUMN versions.release_notes IS 'Release Notes';
COMMENT ON COLUMN versions.start_date IS '计划开始';
COMMENT ON COLUMN versions.end_date IS '计划结束';
COMMENT ON COLUMN versions.target_date IS '发布日期';
COMMENT ON COLUMN versions.delivered_at IS '实际发布';
COMMENT ON COLUMN versions.archived_at IS '归档时间';
COMMENT ON COLUMN versions.status IS 'planning/active/released/archived';
COMMENT ON COLUMN versions.version IS '乐观锁';
COMMENT ON COLUMN versions.deleted IS '软删除';
COMMENT ON COLUMN versions.created_by IS '创建人';
COMMENT ON COLUMN versions.created_at IS '创建时间';
COMMENT ON COLUMN versions.updated_by IS '更新人';
COMMENT ON COLUMN versions.updated_at IS '更新时间';

COMMENT ON TABLE version_delivery_snapshots IS '版本交付快照 (主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN version_delivery_snapshots.id IS '雪花ID';
COMMENT ON COLUMN version_delivery_snapshots.tenant_id IS '租户ID';
COMMENT ON COLUMN version_delivery_snapshots.workspace_id IS '工作空间ID';
COMMENT ON COLUMN version_delivery_snapshots.project_id IS '项目ID';
COMMENT ON COLUMN version_delivery_snapshots.version_id IS '版本ID';
COMMENT ON COLUMN version_delivery_snapshots.progress IS '进度快照';
COMMENT ON COLUMN version_delivery_snapshots.quality IS '质量快照';
COMMENT ON COLUMN version_delivery_snapshots.release_notes IS '发布说明';
COMMENT ON COLUMN version_delivery_snapshots.snapshot_at IS '快照时间';
COMMENT ON COLUMN version_delivery_snapshots.status IS '状态';
COMMENT ON COLUMN version_delivery_snapshots.deleted IS '软删除';
COMMENT ON COLUMN version_delivery_snapshots.created_by IS '创建人';
COMMENT ON COLUMN version_delivery_snapshots.created_at IS '创建时间';
COMMENT ON COLUMN version_delivery_snapshots.updated_by IS '更新人';
COMMENT ON COLUMN version_delivery_snapshots.updated_at IS '更新时间';

COMMENT ON TABLE states IS '状态 (主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN states.id IS '雪花ID';
COMMENT ON COLUMN states.name IS '状态名称';
COMMENT ON COLUMN states.tenant_id IS '租户ID';
COMMENT ON COLUMN states.workspace_id IS '工作空间ID';
COMMENT ON COLUMN states.project_id IS '项目ID';
COMMENT ON COLUMN states.group IS '状态组';
COMMENT ON COLUMN states.color IS '颜色';
COMMENT ON COLUMN states.sequence IS '排序';
COMMENT ON COLUMN states.is_default IS '是否默认';
COMMENT ON COLUMN states.applicable_types IS '适用类型';
COMMENT ON COLUMN states.template_set IS '模板集';
COMMENT ON COLUMN states.status IS '状态';
COMMENT ON COLUMN states.deleted IS '软删除';
COMMENT ON COLUMN states.created_by IS '创建人';
COMMENT ON COLUMN states.created_at IS '创建时间';
COMMENT ON COLUMN states.updated_by IS '更新人';
COMMENT ON COLUMN states.updated_at IS '更新时间';
COMMENT ON COLUMN states."group IS '状态分组（如 todo/in_progress/done）';

COMMENT ON TABLE state_transitions IS '状态流转规则 (主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN state_transitions.id IS '雪花ID';
COMMENT ON COLUMN state_transitions.tenant_id IS '租户ID';
COMMENT ON COLUMN state_transitions.workspace_id IS '工作空间ID';
COMMENT ON COLUMN state_transitions.project_id IS '项目ID';
COMMENT ON COLUMN state_transitions.type_code IS '工作项类型';
COMMENT ON COLUMN state_transitions.from_state_id IS '起始状态';
COMMENT ON COLUMN state_transitions.to_state_id IS '目标状态';
COMMENT ON COLUMN state_transitions.required_fields IS '必填字段';
COMMENT ON COLUMN state_transitions.status IS '状态';
COMMENT ON COLUMN state_transitions.deleted IS '软删除';
COMMENT ON COLUMN state_transitions.created_by IS '创建人';
COMMENT ON COLUMN state_transitions.created_at IS '创建时间';
COMMENT ON COLUMN state_transitions.updated_by IS '更新人';
COMMENT ON COLUMN state_transitions.updated_at IS '更新时间';

COMMENT ON TABLE modules IS '模块 (主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN modules.id IS '雪花ID';
COMMENT ON COLUMN modules.code IS '模块编码';
COMMENT ON COLUMN modules.name IS '模块名称';
COMMENT ON COLUMN modules.tenant_id IS '租户ID';
COMMENT ON COLUMN modules.workspace_id IS '工作空间ID';
COMMENT ON COLUMN modules.project_id IS '项目ID';
COMMENT ON COLUMN modules.public_id IS '公开ID';
COMMENT ON COLUMN modules.description IS '描述';
COMMENT ON COLUMN modules.lead_id IS '负责人';
COMMENT ON COLUMN modules.sort_order IS '排序';
COMMENT ON COLUMN modules.status IS '状态';
COMMENT ON COLUMN modules.deleted IS '软删除';
COMMENT ON COLUMN modules.created_by IS '创建人';
COMMENT ON COLUMN modules.created_at IS '创建时间';
COMMENT ON COLUMN modules.updated_by IS '更新人';
COMMENT ON COLUMN modules.updated_at IS '更新时间';

COMMENT ON TABLE labels IS '标签 (主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN labels.id IS '雪花ID';
COMMENT ON COLUMN labels.code IS '标签编码';
COMMENT ON COLUMN labels.name IS '标签名称';
COMMENT ON COLUMN labels.tenant_id IS '租户ID';
COMMENT ON COLUMN labels.workspace_id IS '工作空间ID';
COMMENT ON COLUMN labels.project_id IS '项目ID';
COMMENT ON COLUMN labels.color IS '颜色';
COMMENT ON COLUMN labels.description IS '描述';
COMMENT ON COLUMN labels.status IS '状态';
COMMENT ON COLUMN labels.deleted IS '软删除';
COMMENT ON COLUMN labels.created_by IS '创建人';
COMMENT ON COLUMN labels.created_at IS '创建时间';
COMMENT ON COLUMN labels.updated_by IS '更新人';
COMMENT ON COLUMN labels.updated_at IS '更新时间';

COMMENT ON TABLE estimate_points IS '估算体系 (主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN estimate_points.id IS '雪花ID';
COMMENT ON COLUMN estimate_points.code IS '估算编码';
COMMENT ON COLUMN estimate_points.name IS '估算名称';
COMMENT ON COLUMN estimate_points.tenant_id IS '租户ID';
COMMENT ON COLUMN estimate_points.workspace_id IS '工作空间ID';
COMMENT ON COLUMN estimate_points.description IS '描述';
COMMENT ON COLUMN estimate_points.points_config IS '估算配置';
COMMENT ON COLUMN estimate_points.is_default IS '是否默认';
COMMENT ON COLUMN estimate_points.status IS '状态';
COMMENT ON COLUMN estimate_points.deleted IS '软删除';
COMMENT ON COLUMN estimate_points.created_by IS '创建人';
COMMENT ON COLUMN estimate_points.created_at IS '创建时间';
COMMENT ON COLUMN estimate_points.updated_by IS '更新人';
COMMENT ON COLUMN estimate_points.updated_at IS '更新时间';

COMMENT ON TABLE automation_rules IS '自动化规则 (主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN automation_rules.id IS '雪花ID';
COMMENT ON COLUMN automation_rules.code IS '规则编码';
COMMENT ON COLUMN automation_rules.name IS '规则名称';
COMMENT ON COLUMN automation_rules.tenant_id IS '租户ID';
COMMENT ON COLUMN automation_rules.workspace_id IS '工作空间ID';
COMMENT ON COLUMN automation_rules.project_id IS '项目ID';
COMMENT ON COLUMN automation_rules.description IS '描述';
COMMENT ON COLUMN automation_rules.trigger_type IS '触发类型';
COMMENT ON COLUMN automation_rules.conditions IS '条件配置';
COMMENT ON COLUMN automation_rules.actions IS '动作配置';
COMMENT ON COLUMN automation_rules.sort_order IS '排序';
COMMENT ON COLUMN automation_rules.status IS '状态';
COMMENT ON COLUMN automation_rules.deleted IS '软删除';
COMMENT ON COLUMN automation_rules.created_by IS '创建人';
COMMENT ON COLUMN automation_rules.created_at IS '创建时间';
COMMENT ON COLUMN automation_rules.updated_by IS '更新人';
COMMENT ON COLUMN automation_rules.updated_at IS '更新时间';

COMMENT ON TABLE rule_executions IS '规则执行日志 (主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN rule_executions.id IS '雪花ID';
COMMENT ON COLUMN rule_executions.tenant_id IS '租户ID';
COMMENT ON COLUMN rule_executions.workspace_id IS '工作空间ID';
COMMENT ON COLUMN rule_executions.project_id IS '项目ID';
COMMENT ON COLUMN rule_executions.rule_id IS '规则ID';
COMMENT ON COLUMN rule_executions.trigger_event_id IS '触发事件';
COMMENT ON COLUMN rule_executions.duration_ms IS '执行耗时';
COMMENT ON COLUMN rule_executions.error_message IS '错误信息';
COMMENT ON COLUMN rule_executions.context_json IS '上下文';
COMMENT ON COLUMN rule_executions.trigger_depth IS '触发深度';
COMMENT ON COLUMN rule_executions.via_automation IS '自动触发标记';
COMMENT ON COLUMN rule_executions.status IS '执行状态';
COMMENT ON COLUMN rule_executions.deleted IS '软删除';
COMMENT ON COLUMN rule_executions.created_by IS '创建人';
COMMENT ON COLUMN rule_executions.created_at IS '创建时间（分区键）';
COMMENT ON COLUMN rule_executions.updated_by IS '更新人';
COMMENT ON COLUMN rule_executions.updated_at IS '更新时间';

COMMENT ON TABLE automation_templates IS '自动化模板 (主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN automation_templates.id IS '雪花ID';
COMMENT ON COLUMN automation_templates.code IS '模板编码';
COMMENT ON COLUMN automation_templates.name IS '模板名称';
COMMENT ON COLUMN automation_templates.description IS '描述';
COMMENT ON COLUMN automation_templates.category IS '分类';
COMMENT ON COLUMN automation_templates.icon IS '图标';
COMMENT ON COLUMN automation_templates.template_config IS '模板配置';
COMMENT ON COLUMN automation_templates.is_default IS '是否默认模板';
COMMENT ON COLUMN automation_templates.sort_order IS '排序';
COMMENT ON COLUMN automation_templates.status IS '状态';
COMMENT ON COLUMN automation_templates.deleted IS '软删除';
COMMENT ON COLUMN automation_templates.created_by IS '创建人';
COMMENT ON COLUMN automation_templates.created_at IS '创建时间';
COMMENT ON COLUMN automation_templates.updated_by IS '更新人';
COMMENT ON COLUMN automation_templates.updated_at IS '更新时间';

COMMENT ON TABLE dashboards IS '仪表盘 (主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN dashboards.id IS '雪花ID';
COMMENT ON COLUMN dashboards.code IS '仪表盘编码';
COMMENT ON COLUMN dashboards.name IS '仪表盘名称';
COMMENT ON COLUMN dashboards.tenant_id IS '租户ID';
COMMENT ON COLUMN dashboards.workspace_id IS '工作空间ID';
COMMENT ON COLUMN dashboards.project_id IS '项目ID';
COMMENT ON COLUMN dashboards.description IS '描述';
COMMENT ON COLUMN dashboards.layout IS '布局配置';
COMMENT ON COLUMN dashboards.status IS '状态';
COMMENT ON COLUMN dashboards.deleted IS '软删除';
COMMENT ON COLUMN dashboards.created_by IS '创建人';
COMMENT ON COLUMN dashboards.created_at IS '创建时间';
COMMENT ON COLUMN dashboards.updated_by IS '更新人';
COMMENT ON COLUMN dashboards.updated_at IS '更新时间';

COMMENT ON TABLE dashboard_widgets IS '仪表盘组件 (主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN dashboard_widgets.id IS '雪花ID';
COMMENT ON COLUMN dashboard_widgets.tenant_id IS '租户ID';
COMMENT ON COLUMN dashboard_widgets.workspace_id IS '工作空间ID';
COMMENT ON COLUMN dashboard_widgets.project_id IS '项目ID';
COMMENT ON COLUMN dashboard_widgets.dashboard_id IS '仪表盘ID';
COMMENT ON COLUMN dashboard_widgets.widget_type IS '组件类型';
COMMENT ON COLUMN dashboard_widgets.name IS '组件名称';
COMMENT ON COLUMN dashboard_widgets.config IS '组件配置';
COMMENT ON COLUMN dashboard_widgets.user_id IS '用户ID（个性化）';
COMMENT ON COLUMN dashboard_widgets.sort_order IS '排序';
COMMENT ON COLUMN dashboard_widgets.status IS '状态';
COMMENT ON COLUMN dashboard_widgets.deleted IS '软删除';
COMMENT ON COLUMN dashboard_widgets.created_by IS '创建人';
COMMENT ON COLUMN dashboard_widgets.created_at IS '创建时间';
COMMENT ON COLUMN dashboard_widgets.updated_by IS '更新人';
COMMENT ON COLUMN dashboard_widgets.updated_at IS '更新时间';

COMMENT ON TABLE dashboard_snapshots IS '仪表盘快照 (主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN dashboard_snapshots.id IS '雪花ID';
COMMENT ON COLUMN dashboard_snapshots.tenant_id IS '租户ID';
COMMENT ON COLUMN dashboard_snapshots.workspace_id IS '工作空间ID';
COMMENT ON COLUMN dashboard_snapshots.project_id IS '项目ID';
COMMENT ON COLUMN dashboard_snapshots.dashboard_id IS '仪表盘ID';
COMMENT ON COLUMN dashboard_snapshots.widget_type IS '组件类型';
COMMENT ON COLUMN dashboard_snapshots.refreshed_at IS '刷新时间';
COMMENT ON COLUMN dashboard_snapshots.data IS '快照数据';
COMMENT ON COLUMN dashboard_snapshots.status IS '状态';
COMMENT ON COLUMN dashboard_snapshots.deleted IS '软删除';
COMMENT ON COLUMN dashboard_snapshots.created_by IS '创建人';
COMMENT ON COLUMN dashboard_snapshots.created_at IS '创建时间';
COMMENT ON COLUMN dashboard_snapshots.updated_by IS '更新人';
COMMENT ON COLUMN dashboard_snapshots.updated_at IS '更新时间';

COMMENT ON TABLE dashboard_templates IS '仪表盘模板 (主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN dashboard_templates.id IS '雪花ID';
COMMENT ON COLUMN dashboard_templates.code IS '模板编码';
COMMENT ON COLUMN dashboard_templates.name IS '模板名称';
COMMENT ON COLUMN dashboard_templates.description IS '描述';
COMMENT ON COLUMN dashboard_templates.category IS '分类';
COMMENT ON COLUMN dashboard_templates.layout IS '布局配置';
COMMENT ON COLUMN dashboard_templates.icon IS '图标';
COMMENT ON COLUMN dashboard_templates.is_default IS '是否默认';
COMMENT ON COLUMN dashboard_templates.sort_order IS '排序';
COMMENT ON COLUMN dashboard_templates.status IS '状态';
COMMENT ON COLUMN dashboard_templates.deleted IS '软删除';
COMMENT ON COLUMN dashboard_templates.created_by IS '创建人';
COMMENT ON COLUMN dashboard_templates.created_at IS '创建时间';
COMMENT ON COLUMN dashboard_templates.updated_by IS '更新人';
COMMENT ON COLUMN dashboard_templates.updated_at IS '更新时间';

COMMENT ON TABLE notifications IS '通知 (主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN notifications.id IS '雪花ID';
COMMENT ON COLUMN notifications.tenant_id IS '租户ID';
COMMENT ON COLUMN notifications.workspace_id IS '工作空间ID';
COMMENT ON COLUMN notifications.recipient_id IS '接收人';
COMMENT ON COLUMN notifications.actor_id IS '触发人';
COMMENT ON COLUMN notifications.notification_type IS '通知类型';
COMMENT ON COLUMN notifications.title IS '标题';
COMMENT ON COLUMN notifications.content IS '内容';
COMMENT ON COLUMN notifications.entity_type IS '关联实体类型';
COMMENT ON COLUMN notifications.entity_id IS '关联实体ID';
COMMENT ON COLUMN notifications.is_read IS '是否已读';
COMMENT ON COLUMN notifications.is_archived IS '是否归档';
COMMENT ON COLUMN notifications.read_at IS '阅读时间';
COMMENT ON COLUMN notifications.status IS '状态';
COMMENT ON COLUMN notifications.deleted IS '软删除';
COMMENT ON COLUMN notifications.created_by IS '创建人';
COMMENT ON COLUMN notifications.created_at IS '创建时间';
COMMENT ON COLUMN notifications.updated_by IS '更新人';
COMMENT ON COLUMN notifications.updated_at IS '更新时间';

COMMENT ON TABLE notification_deliveries IS '通知投递 (主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN notification_deliveries.id IS '雪花ID';
COMMENT ON COLUMN notification_deliveries.tenant_id IS '租户ID';
COMMENT ON COLUMN notification_deliveries.workspace_id IS '工作空间ID';
COMMENT ON COLUMN notification_deliveries.notification_id IS '通知ID';
COMMENT ON COLUMN notification_deliveries.channel IS '投递渠道';
COMMENT ON COLUMN notification_deliveries.status IS '状态';
COMMENT ON COLUMN notification_deliveries.retry_count IS '重试次数';
COMMENT ON COLUMN notification_deliveries.next_retry_at IS '下次重试';
COMMENT ON COLUMN notification_deliveries.delivered_at IS '投递时间';
COMMENT ON COLUMN notification_deliveries.error_message IS '错误信息';
COMMENT ON COLUMN notification_deliveries.deleted IS '软删除';
COMMENT ON COLUMN notification_deliveries.created_by IS '创建人';
COMMENT ON COLUMN notification_deliveries.created_at IS '创建时间';
COMMENT ON COLUMN notification_deliveries.updated_by IS '更新人';
COMMENT ON COLUMN notification_deliveries.updated_at IS '更新时间';

COMMENT ON TABLE notification_preferences IS '通知偏好 (主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN notification_preferences.id IS '雪花ID';
COMMENT ON COLUMN notification_preferences.tenant_id IS '租户ID';
COMMENT ON COLUMN notification_preferences.workspace_id IS '工作空间ID';
COMMENT ON COLUMN notification_preferences.user_id IS '用户ID';
COMMENT ON COLUMN notification_preferences.channel_settings IS '渠道设置';
COMMENT ON COLUMN notification_preferences.mute_all IS '全部免打扰';
COMMENT ON COLUMN notification_preferences.digest_enabled IS '摘要启用';
COMMENT ON COLUMN notification_preferences.digest_schedule IS '摘要频率';
COMMENT ON COLUMN notification_preferences.status IS '状态';
COMMENT ON COLUMN notification_preferences.deleted IS '软删除';
COMMENT ON COLUMN notification_preferences.created_by IS '创建人';
COMMENT ON COLUMN notification_preferences.created_at IS '创建时间';
COMMENT ON COLUMN notification_preferences.updated_by IS '更新人';
COMMENT ON COLUMN notification_preferences.updated_at IS '更新时间';

COMMENT ON TABLE notification_digests IS '通知摘要 (主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN notification_digests.id IS '雪花ID';
COMMENT ON COLUMN notification_digests.tenant_id IS '租户ID';
COMMENT ON COLUMN notification_digests.workspace_id IS '工作空间ID';
COMMENT ON COLUMN notification_digests.user_id IS '用户ID';
COMMENT ON COLUMN notification_digests.digest_type IS '摘要类型';
COMMENT ON COLUMN notification_digests.content IS '摘要内容';
COMMENT ON COLUMN notification_digests.scheduled_for IS '计划投递时间';
COMMENT ON COLUMN notification_digests.sent_at IS '实际投递时间';
COMMENT ON COLUMN notification_digests.status IS '状态';
COMMENT ON COLUMN notification_digests.deleted IS '软删除';
COMMENT ON COLUMN notification_digests.created_by IS '创建人';
COMMENT ON COLUMN notification_digests.created_at IS '创建时间';
COMMENT ON COLUMN notification_digests.updated_by IS '更新人';
COMMENT ON COLUMN notification_digests.updated_at IS '更新时间';

COMMENT ON TABLE search_documents IS '搜索文档索引 (主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN search_documents.id IS '雪花ID';
COMMENT ON COLUMN search_documents.tenant_id IS '租户ID';
COMMENT ON COLUMN search_documents.workspace_id IS '工作空间ID';
COMMENT ON COLUMN search_documents.project_id IS '项目ID';
COMMENT ON COLUMN search_documents.doc_type IS '文档类型';
COMMENT ON COLUMN search_documents.doc_id IS '文档ID';
COMMENT ON COLUMN search_documents.title IS '标题';
COMMENT ON COLUMN search_documents.identifier IS '标识符';
COMMENT ON COLUMN search_documents.content IS '内容';
COMMENT ON COLUMN search_documents.search_tsv IS '全文索引向量';
COMMENT ON COLUMN search_documents.metadata IS '元数据';
COMMENT ON COLUMN search_documents.status IS '状态';
COMMENT ON COLUMN search_documents.deleted IS '软删除';
COMMENT ON COLUMN search_documents.created_by IS '创建人';
COMMENT ON COLUMN search_documents.created_at IS '创建时间';
COMMENT ON COLUMN search_documents.updated_by IS '更新人';
COMMENT ON COLUMN search_documents.updated_at IS '更新时间';

COMMENT ON TABLE search_history IS '搜索历史 (主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN search_history.id IS '雪花ID';
COMMENT ON COLUMN search_history.tenant_id IS '租户ID';
COMMENT ON COLUMN search_history.workspace_id IS '工作空间ID';
COMMENT ON COLUMN search_history.user_id IS '用户ID';
COMMENT ON COLUMN search_history.query IS '查询语句';
COMMENT ON COLUMN search_history.filters IS '过滤条件';
COMMENT ON COLUMN search_history.result_count IS '结果数';
COMMENT ON COLUMN search_history.status IS '状态';
COMMENT ON COLUMN search_history.deleted IS '软删除';
COMMENT ON COLUMN search_history.created_by IS '创建人';
COMMENT ON COLUMN search_history.created_at IS '创建时间';
COMMENT ON COLUMN search_history.updated_by IS '更新人';
COMMENT ON COLUMN search_history.updated_at IS '更新时间';

COMMENT ON TABLE search_bookmarks IS '搜索收藏 (主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN search_bookmarks.id IS '雪花ID';
COMMENT ON COLUMN search_bookmarks.tenant_id IS '租户ID';
COMMENT ON COLUMN search_bookmarks.workspace_id IS '工作空间ID';
COMMENT ON COLUMN search_bookmarks.project_id IS '项目ID';
COMMENT ON COLUMN search_bookmarks.user_id IS '用户ID';
COMMENT ON COLUMN search_bookmarks.name IS '收藏名称';
COMMENT ON COLUMN search_bookmarks.query IS '查询语句';
COMMENT ON COLUMN search_bookmarks.filters IS '过滤条件';
COMMENT ON COLUMN search_bookmarks.is_shared IS '是否共享';
COMMENT ON COLUMN search_bookmarks.sort_order IS '排序';
COMMENT ON COLUMN search_bookmarks.status IS '状态';
COMMENT ON COLUMN search_bookmarks.deleted IS '软删除';
COMMENT ON COLUMN search_bookmarks.created_by IS '创建人';
COMMENT ON COLUMN search_bookmarks.created_at IS '创建时间';
COMMENT ON COLUMN search_bookmarks.updated_by IS '更新人';
COMMENT ON COLUMN search_bookmarks.updated_at IS '更新时间';

COMMENT ON TABLE risk_rules IS '风险规则 (主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN risk_rules.id IS '雪花ID';
COMMENT ON COLUMN risk_rules.code IS '规则编码';
COMMENT ON COLUMN risk_rules.name IS '规则名称';
COMMENT ON COLUMN risk_rules.tenant_id IS '租户ID';
COMMENT ON COLUMN risk_rules.workspace_id IS '工作空间ID';
COMMENT ON COLUMN risk_rules.project_id IS '项目ID';
COMMENT ON COLUMN risk_rules.rule_type IS '规则类型';
COMMENT ON COLUMN risk_rules.condition_json IS '条件配置';
COMMENT ON COLUMN risk_rules.notify_channels IS '通知渠道';
COMMENT ON COLUMN risk_rules.is_active IS '是否启用';
COMMENT ON COLUMN risk_rules.last_triggered IS '最后触发';
COMMENT ON COLUMN risk_rules.status IS '状态';
COMMENT ON COLUMN risk_rules.deleted IS '软删除';
COMMENT ON COLUMN risk_rules.created_by IS '创建人';
COMMENT ON COLUMN risk_rules.created_at IS '创建时间';
COMMENT ON COLUMN risk_rules.updated_by IS '更新人';
COMMENT ON COLUMN risk_rules.updated_at IS '更新时间';

COMMENT ON TABLE risk_alerts IS '风险告警 (主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN risk_alerts.id IS '雪花ID';
COMMENT ON COLUMN risk_alerts.tenant_id IS '租户ID';
COMMENT ON COLUMN risk_alerts.workspace_id IS '工作空间ID';
COMMENT ON COLUMN risk_alerts.project_id IS '项目ID';
COMMENT ON COLUMN risk_alerts.rule_id IS '规则ID';
COMMENT ON COLUMN risk_alerts.severity IS '严重度';
COMMENT ON COLUMN risk_alerts.title IS '标题';
COMMENT ON COLUMN risk_alerts.description IS '描述';
COMMENT ON COLUMN risk_alerts.metadata IS '元数据';
COMMENT ON COLUMN risk_alerts.is_resolved IS '是否已解决';
COMMENT ON COLUMN risk_alerts.resolved_at IS '解决时间';
COMMENT ON COLUMN risk_alerts.resolved_by IS '解决人';
COMMENT ON COLUMN risk_alerts.status IS '状态';
COMMENT ON COLUMN risk_alerts.deleted IS '软删除';
COMMENT ON COLUMN risk_alerts.created_by IS '创建人';
COMMENT ON COLUMN risk_alerts.created_at IS '创建时间';
COMMENT ON COLUMN risk_alerts.updated_by IS '更新人';
COMMENT ON COLUMN risk_alerts.updated_at IS '更新时间';

COMMENT ON TABLE metric_snapshots IS '指标快照 (主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN metric_snapshots.id IS '雪花ID';
COMMENT ON COLUMN metric_snapshots.tenant_id IS '租户ID';
COMMENT ON COLUMN metric_snapshots.workspace_id IS '工作空间ID';
COMMENT ON COLUMN metric_snapshots.project_id IS '项目ID';
COMMENT ON COLUMN metric_snapshots.granularity IS '粒度';
COMMENT ON COLUMN metric_snapshots.ref_id IS '引用ID';
COMMENT ON COLUMN metric_snapshots.metric IS '指标名';
COMMENT ON COLUMN metric_snapshots.snapshot_date IS '快照日期';
COMMENT ON COLUMN metric_snapshots.value IS '指标值';
COMMENT ON COLUMN metric_snapshots.metadata IS '元数据';
COMMENT ON COLUMN metric_snapshots.status IS '状态';
COMMENT ON COLUMN metric_snapshots.deleted IS '软删除';
COMMENT ON COLUMN metric_snapshots.created_by IS '创建人';
COMMENT ON COLUMN metric_snapshots.created_at IS '创建时间';
COMMENT ON COLUMN metric_snapshots.updated_by IS '更新人';
COMMENT ON COLUMN metric_snapshots.updated_at IS '更新时间';

COMMENT ON TABLE metric_adjustments IS '指标调整 (主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN metric_adjustments.id IS '雪花ID';
COMMENT ON COLUMN metric_adjustments.tenant_id IS '租户ID';
COMMENT ON COLUMN metric_adjustments.workspace_id IS '工作空间ID';
COMMENT ON COLUMN metric_adjustments.project_id IS '项目ID';
COMMENT ON COLUMN metric_adjustments.snapshot_id IS '快照ID';
COMMENT ON COLUMN metric_adjustments.metric IS '指标名';
COMMENT ON COLUMN metric_adjustments.original_value IS '原始值';
COMMENT ON COLUMN metric_adjustments.adjusted_value IS '调整值';
COMMENT ON COLUMN metric_adjustments.reason IS '调整原因';
COMMENT ON COLUMN metric_adjustments.adjusted_by IS '调整人';
COMMENT ON COLUMN metric_adjustments.status IS '状态';
COMMENT ON COLUMN metric_adjustments.deleted IS '软删除';
COMMENT ON COLUMN metric_adjustments.created_by IS '创建人';
COMMENT ON COLUMN metric_adjustments.created_at IS '创建时间';
COMMENT ON COLUMN metric_adjustments.updated_by IS '更新人';
COMMENT ON COLUMN metric_adjustments.updated_at IS '更新时间';

COMMENT ON TABLE webhooks IS 'Webhook 配置 (主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN webhooks.id IS '雪花ID';
COMMENT ON COLUMN webhooks.code IS 'Webhook编码';
COMMENT ON COLUMN webhooks.name IS '名称';
COMMENT ON COLUMN webhooks.tenant_id IS '租户ID';
COMMENT ON COLUMN webhooks.workspace_id IS '工作空间ID';
COMMENT ON COLUMN webhooks.project_id IS '项目ID';
COMMENT ON COLUMN webhooks.target_url IS '目标URL';
COMMENT ON COLUMN webhooks.secret IS 'HMAC密钥';
COMMENT ON COLUMN webhooks.events IS '事件白名单';
COMMENT ON COLUMN webhooks.is_active IS '启用';
COMMENT ON COLUMN webhooks.last_error IS '最后错误';
COMMENT ON COLUMN webhooks.last_triggered IS '最后触发';
COMMENT ON COLUMN webhooks.last_status IS '最后状态';
COMMENT ON COLUMN webhooks.unhealthy_at IS '异常时间';
COMMENT ON COLUMN webhooks.status IS '状态';
COMMENT ON COLUMN webhooks.deleted IS '软删除';
COMMENT ON COLUMN webhooks.created_by IS '创建人';
COMMENT ON COLUMN webhooks.created_at IS '创建时间';
COMMENT ON COLUMN webhooks.updated_by IS '更新人';
COMMENT ON COLUMN webhooks.updated_at IS '更新时间';

COMMENT ON TABLE webhook_logs IS 'Webhook 投递日志 (主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN webhook_logs.id IS '雪花ID';
COMMENT ON COLUMN webhook_logs.tenant_id IS '租户ID';
COMMENT ON COLUMN webhook_logs.workspace_id IS '工作空间ID';
COMMENT ON COLUMN webhook_logs.webhook_id IS 'WebhookID';
COMMENT ON COLUMN webhook_logs.delivery_id IS '投递ID';
COMMENT ON COLUMN webhook_logs.event_type IS '事件类型';
COMMENT ON COLUMN webhook_logs.event_id IS '事件ID';
COMMENT ON COLUMN webhook_logs.request_url IS '请求URL';
COMMENT ON COLUMN webhook_logs.request_method IS '请求方法';
COMMENT ON COLUMN webhook_logs.request_headers IS '请求头';
COMMENT ON COLUMN webhook_logs.request_body IS '请求体';
COMMENT ON COLUMN webhook_logs.response_status IS '响应状态码';
COMMENT ON COLUMN webhook_logs.response_body IS '响应体';
COMMENT ON COLUMN webhook_logs.response_headers IS '响应头';
COMMENT ON COLUMN webhook_logs.attempt IS '尝试次数';
COMMENT ON COLUMN webhook_logs.duration_ms IS '耗时';
COMMENT ON COLUMN webhook_logs.error IS '错误';
COMMENT ON COLUMN webhook_logs.status IS '投递状态';
COMMENT ON COLUMN webhook_logs.deleted IS '软删除';
COMMENT ON COLUMN webhook_logs.created_by IS '创建人';
COMMENT ON COLUMN webhook_logs.created_at IS '创建时间（分区键）';
COMMENT ON COLUMN webhook_logs.updated_by IS '更新人';
COMMENT ON COLUMN webhook_logs.updated_at IS '更新时间';

COMMENT ON TABLE workbench_configs IS '工作台配置 (主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN workbench_configs.id IS '雪花ID';
COMMENT ON COLUMN workbench_configs.tenant_id IS '租户ID';
COMMENT ON COLUMN workbench_configs.workspace_id IS '工作空间ID';
COMMENT ON COLUMN workbench_configs.project_id IS '项目ID';
COMMENT ON COLUMN workbench_configs.user_id IS '用户ID';
COMMENT ON COLUMN workbench_configs.layout IS '布局配置';
COMMENT ON COLUMN workbench_configs.widget_states IS '组件状态';
COMMENT ON COLUMN workbench_configs.focus_enabled IS '专注模式';
COMMENT ON COLUMN workbench_configs.status IS '状态';
COMMENT ON COLUMN workbench_configs.deleted IS '软删除';
COMMENT ON COLUMN workbench_configs.created_by IS '创建人';
COMMENT ON COLUMN workbench_configs.created_at IS '创建时间';
COMMENT ON COLUMN workbench_configs.updated_by IS '更新人';
COMMENT ON COLUMN workbench_configs.updated_at IS '更新时间';

COMMENT ON TABLE view_preferences IS '视图偏好 (主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN view_preferences.id IS '雪花ID';
COMMENT ON COLUMN view_preferences.tenant_id IS '租户ID';
COMMENT ON COLUMN view_preferences.workspace_id IS '工作空间ID';
COMMENT ON COLUMN view_preferences.project_id IS '项目ID';
COMMENT ON COLUMN view_preferences.user_id IS '用户ID';
COMMENT ON COLUMN view_preferences.view_type IS '视图类型';
COMMENT ON COLUMN view_preferences.layout IS '布局';
COMMENT ON COLUMN view_preferences.columns IS '列配置';
COMMENT ON COLUMN view_preferences.filters IS '过滤条件';
COMMENT ON COLUMN view_preferences.sort IS '排序配置';
COMMENT ON COLUMN view_preferences.extra IS '扩展配置';
COMMENT ON COLUMN view_preferences.status IS '状态';
COMMENT ON COLUMN view_preferences.deleted IS '软删除';
COMMENT ON COLUMN view_preferences.created_by IS '创建人';
COMMENT ON COLUMN view_preferences.created_at IS '创建时间';
COMMENT ON COLUMN view_preferences.updated_by IS '更新人';
COMMENT ON COLUMN view_preferences.updated_at IS '更新时间';

COMMENT ON TABLE recent_items IS '最近访问 (主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN recent_items.id IS '雪花ID';
COMMENT ON COLUMN recent_items.tenant_id IS '租户ID';
COMMENT ON COLUMN recent_items.workspace_id IS '工作空间ID';
COMMENT ON COLUMN recent_items.project_id IS '项目ID';
COMMENT ON COLUMN recent_items.user_id IS '用户ID';
COMMENT ON COLUMN recent_items.item_type IS '访问类型';
COMMENT ON COLUMN recent_items.item_id IS '关联ID';
COMMENT ON COLUMN recent_items.title IS '标题';
COMMENT ON COLUMN recent_items.identifier IS '标识符';
COMMENT ON COLUMN recent_items.accessed_at IS '访问时间';
COMMENT ON COLUMN recent_items.status IS '状态';
COMMENT ON COLUMN recent_items.deleted IS '软删除';
COMMENT ON COLUMN recent_items.created_by IS '创建人';
COMMENT ON COLUMN recent_items.created_at IS '创建时间';
COMMENT ON COLUMN recent_items.updated_by IS '更新人';
COMMENT ON COLUMN recent_items.updated_at IS '更新时间';

COMMENT ON TABLE knowledge_spaces IS '知识空间 (主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN knowledge_spaces.id IS '雪花ID';
COMMENT ON COLUMN knowledge_spaces.code IS '编码';
COMMENT ON COLUMN knowledge_spaces.name IS '名称';
COMMENT ON COLUMN knowledge_spaces.tenant_id IS '租户ID';
COMMENT ON COLUMN knowledge_spaces.workspace_id IS '工作空间ID';
COMMENT ON COLUMN knowledge_spaces.description IS '描述';
COMMENT ON COLUMN knowledge_spaces.icon IS '图标';
COMMENT ON COLUMN knowledge_spaces.status IS '状态';
COMMENT ON COLUMN knowledge_spaces.deleted IS '软删除';
COMMENT ON COLUMN knowledge_spaces.created_by IS '创建人';
COMMENT ON COLUMN knowledge_spaces.created_at IS '创建时间';
COMMENT ON COLUMN knowledge_spaces.updated_by IS '更新人';
COMMENT ON COLUMN knowledge_spaces.updated_at IS '更新时间';

COMMENT ON TABLE knowledge_pages IS '知识页面 (主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN knowledge_pages.id IS '雪花ID';
COMMENT ON COLUMN knowledge_pages.code IS '编码';
COMMENT ON COLUMN knowledge_pages.name IS '名称';
COMMENT ON COLUMN knowledge_pages.tenant_id IS '租户ID';
COMMENT ON COLUMN knowledge_pages.workspace_id IS '工作空间ID';
COMMENT ON COLUMN knowledge_pages.knowledge_space_id IS '知识空间ID';
COMMENT ON COLUMN knowledge_pages.public_id IS '公开ID';
COMMENT ON COLUMN knowledge_pages.parent_id IS '父页面';
COMMENT ON COLUMN knowledge_pages.depth IS '层级';
COMMENT ON COLUMN knowledge_pages.content_json IS '内容（JSON）';
COMMENT ON COLUMN knowledge_pages.content_html IS '内容（HTML）';
COMMENT ON COLUMN knowledge_pages.sort_order IS '排序';
COMMENT ON COLUMN knowledge_pages.version IS '乐观锁';
COMMENT ON COLUMN knowledge_pages.status IS '状态';
COMMENT ON COLUMN knowledge_pages.deleted IS '软删除';
COMMENT ON COLUMN knowledge_pages.created_by IS '创建人';
COMMENT ON COLUMN knowledge_pages.created_at IS '创建时间';
COMMENT ON COLUMN knowledge_pages.updated_by IS '更新人';
COMMENT ON COLUMN knowledge_pages.updated_at IS '更新时间';

COMMENT ON TABLE sso_providers IS 'SSO 提供方 (主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN sso_providers.id IS '雪花ID';
COMMENT ON COLUMN sso_providers.name IS '名称';
COMMENT ON COLUMN sso_providers.protocol IS '协议类型';
COMMENT ON COLUMN sso_providers.issuer_url IS 'Issuer URL';
COMMENT ON COLUMN sso_providers.client_id IS '客户端ID';
COMMENT ON COLUMN sso_providers.client_secret IS '客户端密钥';
COMMENT ON COLUMN sso_providers.redirect_uri IS '回调URL';
COMMENT ON COLUMN sso_providers.auth_url IS '认证URL';
COMMENT ON COLUMN sso_providers.token_url IS 'Token URL';
COMMENT ON COLUMN sso_providers.userinfo_url IS '用户信息URL';
COMMENT ON COLUMN sso_providers.jwks_url IS 'JWKS URL';
COMMENT ON COLUMN sso_providers.sso_url IS 'SSO URL';
COMMENT ON COLUMN sso_providers.idp_issuer IS 'IDP Issuer';
COMMENT ON COLUMN sso_providers.idp_certificate IS 'IDP 证书';
COMMENT ON COLUMN sso_providers.skip_signature IS '跳过签名';
COMMENT ON COLUMN sso_providers.scopes IS '范围';
COMMENT ON COLUMN sso_providers.auto_create_user IS '自动创建用户';
COMMENT ON COLUMN sso_providers.default_role IS '默认角色';
COMMENT ON COLUMN sso_providers.attribute_mapping IS '属性映射';
COMMENT ON COLUMN sso_providers.enabled IS '启用';
COMMENT ON COLUMN sso_providers.status IS '状态';
COMMENT ON COLUMN sso_providers.deleted IS '软删除';
COMMENT ON COLUMN sso_providers.created_by IS '创建人';
COMMENT ON COLUMN sso_providers.created_at IS '创建时间';
COMMENT ON COLUMN sso_providers.updated_by IS '更新人';
COMMENT ON COLUMN sso_providers.updated_at IS '更新时间';

COMMENT ON TABLE api_tokens IS 'API 令牌 (主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN api_tokens.id IS '雪花ID';
COMMENT ON COLUMN api_tokens.tenant_id IS '租户ID';
COMMENT ON COLUMN api_tokens.user_id IS '用户ID';
COMMENT ON COLUMN api_tokens.name IS '令牌名称';
COMMENT ON COLUMN api_tokens.token_hash IS '令牌哈希';
COMMENT ON COLUMN api_tokens.scopes IS '权限范围';
COMMENT ON COLUMN api_tokens.expires_at IS '过期时间';
COMMENT ON COLUMN api_tokens.last_used_at IS '最后使用';
COMMENT ON COLUMN api_tokens.revoked_at IS '撤销时间';
COMMENT ON COLUMN api_tokens.status IS '状态';
COMMENT ON COLUMN api_tokens.deleted IS '软删除';
COMMENT ON COLUMN api_tokens.created_by IS '创建人';
COMMENT ON COLUMN api_tokens.created_at IS '创建时间';
COMMENT ON COLUMN api_tokens.updated_by IS '更新人';
COMMENT ON COLUMN api_tokens.updated_at IS '更新时间';

COMMENT ON TABLE audit_logs IS '审计日志 (主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN audit_logs.id IS '雪花ID';
COMMENT ON COLUMN audit_logs.tenant_id IS '租户ID';
COMMENT ON COLUMN audit_logs.workspace_id IS '工作空间ID';
COMMENT ON COLUMN audit_logs.actor_id IS '操作人';
COMMENT ON COLUMN audit_logs.action IS '操作类型';
COMMENT ON COLUMN audit_logs.target_type IS '目标类型';
COMMENT ON COLUMN audit_logs.target_id IS '目标ID';
COMMENT ON COLUMN audit_logs.details IS '详情';
COMMENT ON COLUMN audit_logs.ip_address IS 'IP地址';
COMMENT ON COLUMN audit_logs.user_agent IS '用户代理';
COMMENT ON COLUMN audit_logs.status IS '状态';
COMMENT ON COLUMN audit_logs.deleted IS '软删除';
COMMENT ON COLUMN audit_logs.created_by IS '创建人';
COMMENT ON COLUMN audit_logs.created_at IS '创建时间（分区键）';
COMMENT ON COLUMN audit_logs.updated_by IS '更新人';
COMMENT ON COLUMN audit_logs.updated_at IS '更新时间';

COMMENT ON TABLE domain_events IS '领域事件 (主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN domain_events.id IS '雪花ID';
COMMENT ON COLUMN domain_events.tenant_id IS '租户ID';
COMMENT ON COLUMN domain_events.workspace_id IS '工作空间ID';
COMMENT ON COLUMN domain_events.event_type IS '事件类型';
COMMENT ON COLUMN domain_events.aggregate_type IS '聚合类型';
COMMENT ON COLUMN domain_events.aggregate_id IS '聚合ID';
COMMENT ON COLUMN domain_events.payload IS '事件数据';
COMMENT ON COLUMN domain_events.metadata IS '元数据';
COMMENT ON COLUMN domain_events.published_at IS '发布时间';
COMMENT ON COLUMN domain_events.status IS '状态';
COMMENT ON COLUMN domain_events.deleted IS '软删除';
COMMENT ON COLUMN domain_events.created_by IS '创建人';
COMMENT ON COLUMN domain_events.created_at IS '创建时间';
COMMENT ON COLUMN domain_events.updated_by IS '更新人';
COMMENT ON COLUMN domain_events.updated_at IS '更新时间';

COMMENT ON TABLE invitations IS '邀请 (主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN invitations.id IS '雪花ID';
COMMENT ON COLUMN invitations.tenant_id IS '租户ID';
COMMENT ON COLUMN invitations.workspace_id IS '工作空间ID';
COMMENT ON COLUMN invitations.email IS '邮箱';
COMMENT ON COLUMN invitations.inviter_id IS '邀请人';
COMMENT ON COLUMN invitations.role IS '角色';
COMMENT ON COLUMN invitations.token_hash IS '令牌哈希';
COMMENT ON COLUMN invitations.message IS '邀请消息';
COMMENT ON COLUMN invitations.expires_at IS '过期时间';
COMMENT ON COLUMN invitations.accepted_at IS '接受时间';
COMMENT ON COLUMN invitations.revoked_at IS '撤销时间';
COMMENT ON COLUMN invitations.status IS '状态';
COMMENT ON COLUMN invitations.deleted IS '软删除';
COMMENT ON COLUMN invitations.created_by IS '创建人';
COMMENT ON COLUMN invitations.created_at IS '创建时间';
COMMENT ON COLUMN invitations.updated_by IS '更新人';
COMMENT ON COLUMN invitations.updated_at IS '更新时间';

COMMENT ON TABLE tenant_members IS '租户成员 (主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN tenant_members.id IS '雪花ID';
COMMENT ON COLUMN tenant_members.tenant_id IS '租户ID';
COMMENT ON COLUMN tenant_members.user_id IS '用户ID';
COMMENT ON COLUMN tenant_members.role IS '角色';
COMMENT ON COLUMN tenant_members.status IS '状态';
COMMENT ON COLUMN tenant_members.is_owner IS '是否所有者';
COMMENT ON COLUMN tenant_members.joined_at IS '加入时间';
COMMENT ON COLUMN tenant_members.deleted IS '软删除';
COMMENT ON COLUMN tenant_members.created_by IS '创建人';
COMMENT ON COLUMN tenant_members.created_at IS '创建时间';
COMMENT ON COLUMN tenant_members.updated_by IS '更新人';
COMMENT ON COLUMN tenant_members.updated_at IS '更新时间';

COMMENT ON TABLE user_preferences IS '用户偏好 (主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN user_preferences.id IS '雪花ID';
COMMENT ON COLUMN user_preferences.tenant_id IS '租户ID';
COMMENT ON COLUMN user_preferences.user_id IS '用户ID';
COMMENT ON COLUMN user_preferences.key IS '偏好键';
COMMENT ON COLUMN user_preferences.value IS '偏好值';
COMMENT ON COLUMN user_preferences.status IS '状态';
COMMENT ON COLUMN user_preferences.deleted IS '软删除';
COMMENT ON COLUMN user_preferences.created_by IS '创建人';
COMMENT ON COLUMN user_preferences.created_at IS '创建时间';
COMMENT ON COLUMN user_preferences.updated_by IS '更新人';
COMMENT ON COLUMN user_preferences.updated_at IS '更新时间';

COMMENT ON TABLE role_menus IS '角色-权限关联 (主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN role_menus.id IS '雪花ID';
COMMENT ON COLUMN role_menus.role_id IS '角色ID';
COMMENT ON COLUMN role_menus.menu_id IS '菜单ID';
COMMENT ON COLUMN role_menus.status IS '状态';
COMMENT ON COLUMN role_menus.deleted IS '软删除';
COMMENT ON COLUMN role_menus.created_by IS '创建人';
COMMENT ON COLUMN role_menus.created_at IS '创建时间';
COMMENT ON COLUMN role_menus.updated_by IS '更新人';
COMMENT ON COLUMN role_menus.updated_at IS '更新时间';

COMMENT ON TABLE project_configs IS '项目配置 (主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN project_configs.id IS '雪花ID';
COMMENT ON COLUMN project_configs.tenant_id IS '租户ID';
COMMENT ON COLUMN project_configs.workspace_id IS '工作空间ID';
COMMENT ON COLUMN project_configs.project_id IS '项目ID';
COMMENT ON COLUMN project_configs.config IS '项目配置';
COMMENT ON COLUMN project_configs.status IS '状态';
COMMENT ON COLUMN project_configs.deleted IS '软删除';
COMMENT ON COLUMN project_configs.created_by IS '创建人';
COMMENT ON COLUMN project_configs.created_at IS '创建时间';
COMMENT ON COLUMN project_configs.updated_by IS '更新人';
COMMENT ON COLUMN project_configs.updated_at IS '更新时间';

COMMENT ON TABLE version_sprint_relations IS '版本-迭代关联 (主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN version_sprint_relations.id IS '雪花ID';
COMMENT ON COLUMN version_sprint_relations.tenant_id IS '租户ID';
COMMENT ON COLUMN version_sprint_relations.workspace_id IS '工作空间ID';
COMMENT ON COLUMN version_sprint_relations.project_id IS '项目ID';
COMMENT ON COLUMN version_sprint_relations.version_id IS '版本ID';
COMMENT ON COLUMN version_sprint_relations.sprint_id IS '迭代ID';
COMMENT ON COLUMN version_sprint_relations.status IS '状态';
COMMENT ON COLUMN version_sprint_relations.deleted IS '软删除';
COMMENT ON COLUMN version_sprint_relations.created_by IS '创建人';
COMMENT ON COLUMN version_sprint_relations.created_at IS '创建时间';
COMMENT ON COLUMN version_sprint_relations.updated_by IS '更新人';
COMMENT ON COLUMN version_sprint_relations.updated_at IS '更新时间';

COMMENT ON TABLE sprint_requirements IS '迭代需求关联 (主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN sprint_requirements.id IS '雪花ID';
COMMENT ON COLUMN sprint_requirements.tenant_id IS '租户ID';
COMMENT ON COLUMN sprint_requirements.workspace_id IS '工作空间ID';
COMMENT ON COLUMN sprint_requirements.project_id IS '项目ID';
COMMENT ON COLUMN sprint_requirements.sprint_id IS '迭代ID';
COMMENT ON COLUMN sprint_requirements.requirement_id IS '需求ID';
COMMENT ON COLUMN sprint_requirements.status IS '状态';
COMMENT ON COLUMN sprint_requirements.deleted IS '软删除';
COMMENT ON COLUMN sprint_requirements.created_by IS '创建人';
COMMENT ON COLUMN sprint_requirements.created_at IS '创建时间';
COMMENT ON COLUMN sprint_requirements.updated_by IS '更新人';
COMMENT ON COLUMN sprint_requirements.updated_at IS '更新时间';
COMMENT ON COLUMN sprint_requirements.added_midway IS '是否迭代中途加入';
COMMENT ON COLUMN sprint_requirements.sort_order IS '迭代内排序权重';
COMMENT ON COLUMN sprint_requirements.added_by IS '添加人';
COMMENT ON COLUMN sprint_requirements.added_at IS '添加时间';

COMMENT ON TABLE sprint_tasks IS '迭代任务关联 (主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN sprint_tasks.id IS '雪花ID';
COMMENT ON COLUMN sprint_tasks.tenant_id IS '租户ID';
COMMENT ON COLUMN sprint_tasks.workspace_id IS '工作空间ID';
COMMENT ON COLUMN sprint_tasks.project_id IS '项目ID';
COMMENT ON COLUMN sprint_tasks.sprint_id IS '迭代ID';
COMMENT ON COLUMN sprint_tasks.task_id IS '任务ID';
COMMENT ON COLUMN sprint_tasks.status IS '状态';
COMMENT ON COLUMN sprint_tasks.deleted IS '软删除';
COMMENT ON COLUMN sprint_tasks.created_by IS '创建人';
COMMENT ON COLUMN sprint_tasks.created_at IS '创建时间';
COMMENT ON COLUMN sprint_tasks.updated_by IS '更新人';
COMMENT ON COLUMN sprint_tasks.updated_at IS '更新时间';
COMMENT ON COLUMN sprint_tasks.added_midway IS '是否迭代中途加入';
COMMENT ON COLUMN sprint_tasks.sort_order IS '迭代内排序权重';
COMMENT ON COLUMN sprint_tasks.added_by IS '添加人';
COMMENT ON COLUMN sprint_tasks.added_at IS '添加时间';

COMMENT ON TABLE sprint_defects IS '迭代缺陷关联 (主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN sprint_defects.id IS '雪花ID';
COMMENT ON COLUMN sprint_defects.tenant_id IS '租户ID';
COMMENT ON COLUMN sprint_defects.workspace_id IS '工作空间ID';
COMMENT ON COLUMN sprint_defects.project_id IS '项目ID';
COMMENT ON COLUMN sprint_defects.sprint_id IS '迭代ID';
COMMENT ON COLUMN sprint_defects.defect_id IS '缺陷ID';
COMMENT ON COLUMN sprint_defects.status IS '状态';
COMMENT ON COLUMN sprint_defects.deleted IS '软删除';
COMMENT ON COLUMN sprint_defects.created_by IS '创建人';
COMMENT ON COLUMN sprint_defects.created_at IS '创建时间';
COMMENT ON COLUMN sprint_defects.updated_by IS '更新人';
COMMENT ON COLUMN sprint_defects.updated_at IS '更新时间';
COMMENT ON COLUMN sprint_defects.added_midway IS '是否迭代中途加入';
COMMENT ON COLUMN sprint_defects.sort_order IS '迭代内排序权重';
COMMENT ON COLUMN sprint_defects.added_by IS '添加人';
COMMENT ON COLUMN sprint_defects.added_at IS '添加时间';

COMMENT ON TABLE content_templates IS '内容模板 (主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN content_templates.id IS '雪花ID';
COMMENT ON COLUMN content_templates.tenant_id IS '租户ID';
COMMENT ON COLUMN content_templates.workspace_id IS '工作空间ID';
COMMENT ON COLUMN content_templates.project_id IS '项目ID';
COMMENT ON COLUMN content_templates.name IS '模板名称';
COMMENT ON COLUMN content_templates.template_type IS '模板类型';
COMMENT ON COLUMN content_templates.content_json IS '内容 JSON';
COMMENT ON COLUMN content_templates.content_html IS '内容 HTML';
COMMENT ON COLUMN content_templates.is_default IS '是否默认';
COMMENT ON COLUMN content_templates.status IS '状态';
COMMENT ON COLUMN content_templates.deleted IS '软删除';
COMMENT ON COLUMN content_templates.created_by IS '创建人';
COMMENT ON COLUMN content_templates.created_at IS '创建时间';
COMMENT ON COLUMN content_templates.updated_by IS '更新人';
COMMENT ON COLUMN content_templates.updated_at IS '更新时间';

COMMENT ON TABLE reviews IS '评审 (主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN reviews.id IS '雪花ID';
COMMENT ON COLUMN reviews.tenant_id IS '租户ID';
COMMENT ON COLUMN reviews.workspace_id IS '工作空间ID';
COMMENT ON COLUMN reviews.project_id IS '项目ID';
COMMENT ON COLUMN reviews.name IS '评审名称';
COMMENT ON COLUMN reviews.review_type IS '评审类型';
COMMENT ON COLUMN reviews.status IS '状态';
COMMENT ON COLUMN reviews.description IS '描述';
COMMENT ON COLUMN reviews.due_date IS '截止日期';
COMMENT ON COLUMN reviews.created_date IS '创建日期';
COMMENT ON COLUMN reviews.completed_date IS '完成日期';
COMMENT ON COLUMN reviews.deleted IS '软删除';
COMMENT ON COLUMN reviews.created_by IS '创建人';
COMMENT ON COLUMN reviews.created_at IS '创建时间';
COMMENT ON COLUMN reviews.updated_by IS '更新人';
COMMENT ON COLUMN reviews.updated_at IS '更新时间';
COMMENT ON COLUMN reviews.entity_type IS '关联实体类型（如 requirement/task）';
COMMENT ON COLUMN reviews.entity_id IS '关联实体ID';

COMMENT ON TABLE review_assignments IS '评审分配 (主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN review_assignments.id IS '雪花ID';
COMMENT ON COLUMN review_assignments.tenant_id IS '租户ID';
COMMENT ON COLUMN review_assignments.workspace_id IS '工作空间ID';
COMMENT ON COLUMN review_assignments.project_id IS '项目ID';
COMMENT ON COLUMN review_assignments.review_id IS '评审ID';
COMMENT ON COLUMN review_assignments.assignee_id IS '被指派人';
COMMENT ON COLUMN review_assignments.status IS '状态';
COMMENT ON COLUMN review_assignments.deleted IS '软删除';
COMMENT ON COLUMN review_assignments.created_by IS '创建人';
COMMENT ON COLUMN review_assignments.created_at IS '创建时间';
COMMENT ON COLUMN review_assignments.updated_by IS '更新人';
COMMENT ON COLUMN review_assignments.updated_at IS '更新时间';

COMMENT ON TABLE documents IS '文档 (主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN documents.id IS '雪花ID';
COMMENT ON COLUMN documents.code IS '文档编码';
COMMENT ON COLUMN documents.name IS '文档名称';
COMMENT ON COLUMN documents.tenant_id IS '租户ID';
COMMENT ON COLUMN documents.workspace_id IS '工作空间ID';
COMMENT ON COLUMN documents.project_id IS '项目ID';
COMMENT ON COLUMN documents.public_id IS '公开ID';
COMMENT ON COLUMN documents.description IS '描述';
COMMENT ON COLUMN documents.cover_image_url IS '封面图片';
COMMENT ON COLUMN documents.is_published IS '是否发布';
COMMENT ON COLUMN documents.is_archived IS '是否归档';
COMMENT ON COLUMN documents.sort_order IS '排序';
COMMENT ON COLUMN documents.status IS '状态';
COMMENT ON COLUMN documents.deleted IS '软删除';
COMMENT ON COLUMN documents.created_by IS '创建人';
COMMENT ON COLUMN documents.created_at IS '创建时间';
COMMENT ON COLUMN documents.updated_by IS '更新人';
COMMENT ON COLUMN documents.updated_at IS '更新时间';

COMMENT ON TABLE document_versions IS '文档版本 (主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN document_versions.id IS '雪花ID';
COMMENT ON COLUMN document_versions.tenant_id IS '租户ID';
COMMENT ON COLUMN document_versions.workspace_id IS '工作空间ID';
COMMENT ON COLUMN document_versions.project_id IS '项目ID';
COMMENT ON COLUMN document_versions.document_id IS '文档ID';
COMMENT ON COLUMN document_versions.version_number IS '版本号';
COMMENT ON COLUMN document_versions.change_summary IS '变更摘要';
COMMENT ON COLUMN document_versions.content_json IS '内容 JSON';
COMMENT ON COLUMN document_versions.content_html IS '内容 HTML';
COMMENT ON COLUMN document_versions.status IS '状态';
COMMENT ON COLUMN document_versions.deleted IS '软删除';
COMMENT ON COLUMN document_versions.created_by IS '创建人';
COMMENT ON COLUMN document_versions.created_at IS '创建时间';
COMMENT ON COLUMN document_versions.updated_by IS '更新人';
COMMENT ON COLUMN document_versions.updated_at IS '更新时间';

COMMENT ON TABLE share_links IS '分享链接 (主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN share_links.id IS '雪花ID';
COMMENT ON COLUMN share_links.tenant_id IS '租户ID';
COMMENT ON COLUMN share_links.workspace_id IS '工作空间ID';
COMMENT ON COLUMN share_links.project_id IS '项目ID';
COMMENT ON COLUMN share_links.entity_type IS '实体类型';
COMMENT ON COLUMN share_links.entity_id IS '实体ID';
COMMENT ON COLUMN share_links.share_token IS '分享令牌';
COMMENT ON COLUMN share_links.scope IS '权限范围';
COMMENT ON COLUMN share_links.password_hash IS '访问密码';
COMMENT ON COLUMN share_links.expires_at IS '过期时间';
COMMENT ON COLUMN share_links.is_active IS '启用';
COMMENT ON COLUMN share_links.access_count IS '访问次数';
COMMENT ON COLUMN share_links.last_accessed_at IS '最后访问';
COMMENT ON COLUMN share_links.status IS '状态';
COMMENT ON COLUMN share_links.deleted IS '软删除';
COMMENT ON COLUMN share_links.created_by IS '创建人';
COMMENT ON COLUMN share_links.created_at IS '创建时间';
COMMENT ON COLUMN share_links.updated_by IS '更新人';
COMMENT ON COLUMN share_links.updated_at IS '更新时间';

COMMENT ON TABLE notification_subscriptions IS '通知订阅 (主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN notification_subscriptions.id IS '雪花ID';
COMMENT ON COLUMN notification_subscriptions.tenant_id IS '租户ID';
COMMENT ON COLUMN notification_subscriptions.workspace_id IS '工作空间ID';
COMMENT ON COLUMN notification_subscriptions.project_id IS '项目ID';
COMMENT ON COLUMN notification_subscriptions.user_id IS '用户ID';
COMMENT ON COLUMN notification_subscriptions.entity_type IS '实体类型';
COMMENT ON COLUMN notification_subscriptions.entity_id IS '实体ID';
COMMENT ON COLUMN notification_subscriptions.event_types IS '订阅事件';
COMMENT ON COLUMN notification_subscriptions.status IS '状态';
COMMENT ON COLUMN notification_subscriptions.deleted IS '软删除';
COMMENT ON COLUMN notification_subscriptions.created_by IS '创建人';
COMMENT ON COLUMN notification_subscriptions.created_at IS '创建时间';
COMMENT ON COLUMN notification_subscriptions.updated_by IS '更新人';
COMMENT ON COLUMN notification_subscriptions.updated_at IS '更新时间';

COMMENT ON TABLE saved_views IS '保存视图 (主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN saved_views.id IS '雪花ID';
COMMENT ON COLUMN saved_views.tenant_id IS '租户ID';
COMMENT ON COLUMN saved_views.workspace_id IS '工作空间ID';
COMMENT ON COLUMN saved_views.project_id IS '项目ID';
COMMENT ON COLUMN saved_views.user_id IS '用户ID';
COMMENT ON COLUMN saved_views.name IS '视图名称';
COMMENT ON COLUMN saved_views.view_type IS '视图类型';
COMMENT ON COLUMN saved_views.filters IS '过滤条件';
COMMENT ON COLUMN saved_views.columns IS '列配置';
COMMENT ON COLUMN saved_views.sort IS '排序配置';
COMMENT ON COLUMN saved_views.is_shared IS '是否共享';
COMMENT ON COLUMN saved_views.sort_order IS '排序';
COMMENT ON COLUMN saved_views.status IS '状态';
COMMENT ON COLUMN saved_views.deleted IS '软删除';
COMMENT ON COLUMN saved_views.created_by IS '创建人';
COMMENT ON COLUMN saved_views.created_at IS '创建时间';
COMMENT ON COLUMN saved_views.updated_by IS '更新人';
COMMENT ON COLUMN saved_views.updated_at IS '更新时间';

COMMENT ON TABLE calendar_events IS '日历事件 (主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN calendar_events.id IS '雪花ID';
COMMENT ON COLUMN calendar_events.tenant_id IS '租户ID';
COMMENT ON COLUMN calendar_events.workspace_id IS '工作空间ID';
COMMENT ON COLUMN calendar_events.project_id IS '项目ID';
COMMENT ON COLUMN calendar_events.title IS '事件标题';
COMMENT ON COLUMN calendar_events.description IS '描述';
COMMENT ON COLUMN calendar_events.start_time IS '开始时间';
COMMENT ON COLUMN calendar_events.end_time IS '结束时间';
COMMENT ON COLUMN calendar_events.is_all_day IS '全天事件';
COMMENT ON COLUMN calendar_events.location IS '地点';
COMMENT ON COLUMN calendar_events.event_type IS '事件类型';
COMMENT ON COLUMN calendar_events.source_type IS '来源类型';
COMMENT ON COLUMN calendar_events.source_id IS '来源ID';
COMMENT ON COLUMN calendar_events.idempotency_key IS '幂等键';
COMMENT ON COLUMN calendar_events.organizer_id IS '组织者';
COMMENT ON COLUMN calendar_events.status IS '状态';
COMMENT ON COLUMN calendar_events.version IS '乐观锁';
COMMENT ON COLUMN calendar_events.deleted IS '软删除';
COMMENT ON COLUMN calendar_events.created_by IS '创建人';
COMMENT ON COLUMN calendar_events.created_at IS '创建时间';
COMMENT ON COLUMN calendar_events.updated_by IS '更新人';
COMMENT ON COLUMN calendar_events.updated_at IS '更新时间';

COMMENT ON TABLE data_jobs IS '数据任务 (主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN data_jobs.id IS '雪花ID';
COMMENT ON COLUMN data_jobs.tenant_id IS '租户ID';
COMMENT ON COLUMN data_jobs.workspace_id IS '工作空间ID';
COMMENT ON COLUMN data_jobs.project_id IS '项目ID';
COMMENT ON COLUMN data_jobs.job_type IS '任务类型';
COMMENT ON COLUMN data_jobs.name IS '任务名称';
COMMENT ON COLUMN data_jobs.parameters IS '参数';
COMMENT ON COLUMN data_jobs.progress IS '进度';
COMMENT ON COLUMN data_jobs.status IS '状态';
COMMENT ON COLUMN data_jobs.error_message IS '错误信息';
COMMENT ON COLUMN data_jobs.scheduled_at IS '计划执行';
COMMENT ON COLUMN data_jobs.executed_at IS '开始执行';
COMMENT ON COLUMN data_jobs.completed_at IS '完成时间';
COMMENT ON COLUMN data_jobs.duration_ms IS '耗时';
COMMENT ON COLUMN data_jobs.triggered_by IS '触发人';
COMMENT ON COLUMN data_jobs.deleted IS '软删除';
COMMENT ON COLUMN data_jobs.created_by IS '创建人';
COMMENT ON COLUMN data_jobs.created_at IS '创建时间';
COMMENT ON COLUMN data_jobs.updated_by IS '更新人';
COMMENT ON COLUMN data_jobs.updated_at IS '更新时间';

COMMENT ON TABLE task_attachments IS '任务附件 (主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN task_attachments.id IS '雪花ID';
COMMENT ON COLUMN task_attachments.tenant_id IS '租户ID';
COMMENT ON COLUMN task_attachments.workspace_id IS '工作空间ID';
COMMENT ON COLUMN task_attachments.project_id IS '项目ID';
COMMENT ON COLUMN task_attachments.task_id IS '任务ID';
COMMENT ON COLUMN task_attachments.file_name IS '文件名';
COMMENT ON COLUMN task_attachments.file_size IS '文件大小';
COMMENT ON COLUMN task_attachments.file_type IS '文件类型';
COMMENT ON COLUMN task_attachments.storage_path IS '存储路径';
COMMENT ON COLUMN task_attachments.thumbnail_path IS '缩略图路径';
COMMENT ON COLUMN task_attachments.status IS '状态';
COMMENT ON COLUMN task_attachments.deleted IS '软删除';
COMMENT ON COLUMN task_attachments.created_by IS '创建人';
COMMENT ON COLUMN task_attachments.created_at IS '创建时间';
COMMENT ON COLUMN task_attachments.updated_by IS '更新人';
COMMENT ON COLUMN task_attachments.updated_at IS '更新时间';

COMMENT ON TABLE task_timelogs IS '任务工时 (主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN task_timelogs.id IS '雪花ID';
COMMENT ON COLUMN task_timelogs.tenant_id IS '租户ID';
COMMENT ON COLUMN task_timelogs.workspace_id IS '工作空间ID';
COMMENT ON COLUMN task_timelogs.project_id IS '项目ID';
COMMENT ON COLUMN task_timelogs.task_id IS '任务ID';
COMMENT ON COLUMN task_timelogs.user_id IS '记录人';
COMMENT ON COLUMN task_timelogs.spent_date IS '日期';
COMMENT ON COLUMN task_timelogs.duration_minutes IS '时长(分)';
COMMENT ON COLUMN task_timelogs.description IS '描述';
COMMENT ON COLUMN task_timelogs.status IS '状态';
COMMENT ON COLUMN task_timelogs.deleted IS '软删除';
COMMENT ON COLUMN task_timelogs.created_by IS '创建人';
COMMENT ON COLUMN task_timelogs.created_at IS '创建时间';
COMMENT ON COLUMN task_timelogs.updated_by IS '更新人';
COMMENT ON COLUMN task_timelogs.updated_at IS '更新时间';

COMMENT ON TABLE requirement_assignees IS '需求执行人 (主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN requirement_assignees.id IS '雪花ID';
COMMENT ON COLUMN requirement_assignees.tenant_id IS '租户ID';
COMMENT ON COLUMN requirement_assignees.workspace_id IS '工作空间ID';
COMMENT ON COLUMN requirement_assignees.project_id IS '项目ID';
COMMENT ON COLUMN requirement_assignees.requirement_id IS '需求ID';
COMMENT ON COLUMN requirement_assignees.user_id IS '用户ID';
COMMENT ON COLUMN requirement_assignees.status IS '状态';
COMMENT ON COLUMN requirement_assignees.deleted IS '软删除';
COMMENT ON COLUMN requirement_assignees.created_by IS '创建人';
COMMENT ON COLUMN requirement_assignees.created_at IS '创建时间';
COMMENT ON COLUMN requirement_assignees.updated_by IS '更新人';
COMMENT ON COLUMN requirement_assignees.updated_at IS '更新时间';

COMMENT ON TABLE requirement_labels IS '需求标签 (主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN requirement_labels.id IS '雪花ID';
COMMENT ON COLUMN requirement_labels.tenant_id IS '租户ID';
COMMENT ON COLUMN requirement_labels.workspace_id IS '工作空间ID';
COMMENT ON COLUMN requirement_labels.project_id IS '项目ID';
COMMENT ON COLUMN requirement_labels.requirement_id IS '需求ID';
COMMENT ON COLUMN requirement_labels.label_id IS '标签ID';
COMMENT ON COLUMN requirement_labels.status IS '状态';
COMMENT ON COLUMN requirement_labels.deleted IS '软删除';
COMMENT ON COLUMN requirement_labels.created_by IS '创建人';
COMMENT ON COLUMN requirement_labels.created_at IS '创建时间';
COMMENT ON COLUMN requirement_labels.updated_by IS '更新人';
COMMENT ON COLUMN requirement_labels.updated_at IS '更新时间';

COMMENT ON TABLE requirement_modules IS '需求模块关联 (主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN requirement_modules.id IS '雪花ID';
COMMENT ON COLUMN requirement_modules.tenant_id IS '租户ID';
COMMENT ON COLUMN requirement_modules.workspace_id IS '工作空间ID';
COMMENT ON COLUMN requirement_modules.project_id IS '项目ID';
COMMENT ON COLUMN requirement_modules.requirement_id IS '需求ID';
COMMENT ON COLUMN requirement_modules.module_id IS '模块ID';
COMMENT ON COLUMN requirement_modules.status IS '状态';
COMMENT ON COLUMN requirement_modules.deleted IS '软删除';
COMMENT ON COLUMN requirement_modules.created_by IS '创建人';
COMMENT ON COLUMN requirement_modules.created_at IS '创建时间';
COMMENT ON COLUMN requirement_modules.updated_by IS '更新人';
COMMENT ON COLUMN requirement_modules.updated_at IS '更新时间';

COMMENT ON TABLE requirement_watchers IS '需求关注人 (主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN requirement_watchers.id IS '雪花ID';
COMMENT ON COLUMN requirement_watchers.tenant_id IS '租户ID';
COMMENT ON COLUMN requirement_watchers.workspace_id IS '工作空间ID';
COMMENT ON COLUMN requirement_watchers.project_id IS '项目ID';
COMMENT ON COLUMN requirement_watchers.requirement_id IS '需求ID';
COMMENT ON COLUMN requirement_watchers.user_id IS '用户ID';
COMMENT ON COLUMN requirement_watchers.status IS '状态';
COMMENT ON COLUMN requirement_watchers.deleted IS '软删除';
COMMENT ON COLUMN requirement_watchers.created_by IS '创建人';
COMMENT ON COLUMN requirement_watchers.created_at IS '创建时间';
COMMENT ON COLUMN requirement_watchers.updated_by IS '更新人';
COMMENT ON COLUMN requirement_watchers.updated_at IS '更新时间';

COMMENT ON TABLE requirement_relations IS '需求关联关系 (主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN requirement_relations.id IS '雪花ID';
COMMENT ON COLUMN requirement_relations.tenant_id IS '租户ID';
COMMENT ON COLUMN requirement_relations.workspace_id IS '工作空间ID';
COMMENT ON COLUMN requirement_relations.project_id IS '项目ID';
COMMENT ON COLUMN requirement_relations.source_requirement_id IS '源需求ID';
COMMENT ON COLUMN requirement_relations.target_requirement_id IS '目标需求ID';
COMMENT ON COLUMN requirement_relations.relation_type IS '关系类型';
COMMENT ON COLUMN requirement_relations.status IS '状态';
COMMENT ON COLUMN requirement_relations.deleted IS '软删除';
COMMENT ON COLUMN requirement_relations.created_by IS '创建人';
COMMENT ON COLUMN requirement_relations.created_at IS '创建时间';
COMMENT ON COLUMN requirement_relations.updated_by IS '更新人';
COMMENT ON COLUMN requirement_relations.updated_at IS '更新时间';

COMMENT ON TABLE requirement_comments IS '需求评论 (主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN requirement_comments.id IS '雪花ID';
COMMENT ON COLUMN requirement_comments.tenant_id IS '租户ID';
COMMENT ON COLUMN requirement_comments.workspace_id IS '工作空间ID';
COMMENT ON COLUMN requirement_comments.project_id IS '项目ID';
COMMENT ON COLUMN requirement_comments.requirement_id IS '需求ID';
COMMENT ON COLUMN requirement_comments.content_json IS '内容 JSON';
COMMENT ON COLUMN requirement_comments.content_html IS '内容 HTML';
COMMENT ON COLUMN requirement_comments.parent_id IS '父评论ID';
COMMENT ON COLUMN requirement_comments.status IS '状态';
COMMENT ON COLUMN requirement_comments.deleted IS '软删除';
COMMENT ON COLUMN requirement_comments.created_by IS '创建人';
COMMENT ON COLUMN requirement_comments.created_at IS '创建时间';
COMMENT ON COLUMN requirement_comments.updated_by IS '更新人';
COMMENT ON COLUMN requirement_comments.updated_at IS '更新时间';
COMMENT ON COLUMN requirement_comments.content_stripped IS '去标签后的纯文本内容';
COMMENT ON COLUMN requirement_comments.mentions IS '@提及的用户ID列表';
COMMENT ON COLUMN requirement_comments.is_edited IS '是否被编辑过';
COMMENT ON COLUMN requirement_comments.edited_at IS '最后编辑时间';

COMMENT ON TABLE requirement_activities IS '需求活动日志 (主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN requirement_activities.id IS '雪花ID';
COMMENT ON COLUMN requirement_activities.tenant_id IS '租户ID';
COMMENT ON COLUMN requirement_activities.workspace_id IS '工作空间ID';
COMMENT ON COLUMN requirement_activities.project_id IS '项目ID';
COMMENT ON COLUMN requirement_activities.requirement_id IS '需求ID';
COMMENT ON COLUMN requirement_activities.verb IS '动作';
COMMENT ON COLUMN requirement_activities.field_name IS '字段名';
COMMENT ON COLUMN requirement_activities.old_value IS '旧值';
COMMENT ON COLUMN requirement_activities.new_value IS '新值';
COMMENT ON COLUMN requirement_activities.actor_id IS '操作人';
COMMENT ON COLUMN requirement_activities.status IS '状态';
COMMENT ON COLUMN requirement_activities.deleted IS '软删除';
COMMENT ON COLUMN requirement_activities.created_by IS '创建人';
COMMENT ON COLUMN requirement_activities.created_at IS '创建时间';
COMMENT ON COLUMN requirement_activities.updated_by IS '更新人';
COMMENT ON COLUMN requirement_activities.updated_at IS '更新时间';

COMMENT ON TABLE requirement_timelogs IS '需求工时 (主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN requirement_timelogs.id IS '雪花ID';
COMMENT ON COLUMN requirement_timelogs.tenant_id IS '租户ID';
COMMENT ON COLUMN requirement_timelogs.workspace_id IS '工作空间ID';
COMMENT ON COLUMN requirement_timelogs.project_id IS '项目ID';
COMMENT ON COLUMN requirement_timelogs.requirement_id IS '需求ID';
COMMENT ON COLUMN requirement_timelogs.user_id IS '记录人';
COMMENT ON COLUMN requirement_timelogs.spent_date IS '日期';
COMMENT ON COLUMN requirement_timelogs.duration_minutes IS '时长(分)';
COMMENT ON COLUMN requirement_timelogs.description IS '描述';
COMMENT ON COLUMN requirement_timelogs.status IS '状态';
COMMENT ON COLUMN requirement_timelogs.deleted IS '软删除';
COMMENT ON COLUMN requirement_timelogs.created_by IS '创建人';
COMMENT ON COLUMN requirement_timelogs.created_at IS '创建时间';
COMMENT ON COLUMN requirement_timelogs.updated_by IS '更新人';
COMMENT ON COLUMN requirement_timelogs.updated_at IS '更新时间';

COMMENT ON TABLE requirement_attachments IS '需求附件 (主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN requirement_attachments.id IS '雪花ID';
COMMENT ON COLUMN requirement_attachments.tenant_id IS '租户ID';
COMMENT ON COLUMN requirement_attachments.workspace_id IS '工作空间ID';
COMMENT ON COLUMN requirement_attachments.project_id IS '项目ID';
COMMENT ON COLUMN requirement_attachments.requirement_id IS '需求ID';
COMMENT ON COLUMN requirement_attachments.file_name IS '文件名';
COMMENT ON COLUMN requirement_attachments.file_size IS '文件大小';
COMMENT ON COLUMN requirement_attachments.file_type IS '文件类型';
COMMENT ON COLUMN requirement_attachments.storage_path IS '存储路径';
COMMENT ON COLUMN requirement_attachments.thumbnail_path IS '缩略图路径';
COMMENT ON COLUMN requirement_attachments.status IS '状态';
COMMENT ON COLUMN requirement_attachments.deleted IS '软删除';
COMMENT ON COLUMN requirement_attachments.created_by IS '创建人';
COMMENT ON COLUMN requirement_attachments.created_at IS '创建时间';
COMMENT ON COLUMN requirement_attachments.updated_by IS '更新人';
COMMENT ON COLUMN requirement_attachments.updated_at IS '更新时间';

COMMENT ON TABLE requirement_ext IS '需求扩展字段 (主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN requirement_ext.id IS '雪花ID';
COMMENT ON COLUMN requirement_ext.tenant_id IS '租户ID';
COMMENT ON COLUMN requirement_ext.workspace_id IS '工作空间ID';
COMMENT ON COLUMN requirement_ext.project_id IS '项目ID';
COMMENT ON COLUMN requirement_ext.requirement_id IS '需求ID';
COMMENT ON COLUMN requirement_ext.field_name IS '字段名';
COMMENT ON COLUMN requirement_ext.field_value IS '字段值';
COMMENT ON COLUMN requirement_ext.field_schema IS '字段元数据';
COMMENT ON COLUMN requirement_ext.status IS '状态';
COMMENT ON COLUMN requirement_ext.deleted IS '软删除';
COMMENT ON COLUMN requirement_ext.created_by IS '创建人';
COMMENT ON COLUMN requirement_ext.created_at IS '创建时间';
COMMENT ON COLUMN requirement_ext.updated_by IS '更新人';
COMMENT ON COLUMN requirement_ext.updated_at IS '更新时间';

COMMENT ON TABLE defect_assignees IS '缺陷处理人 (主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN defect_assignees.id IS '雪花ID';
COMMENT ON COLUMN defect_assignees.tenant_id IS '租户ID';
COMMENT ON COLUMN defect_assignees.workspace_id IS '工作空间ID';
COMMENT ON COLUMN defect_assignees.project_id IS '项目ID';
COMMENT ON COLUMN defect_assignees.defect_id IS '缺陷ID';
COMMENT ON COLUMN defect_assignees.user_id IS '用户ID';
COMMENT ON COLUMN defect_assignees.status IS '状态';
COMMENT ON COLUMN defect_assignees.deleted IS '软删除';
COMMENT ON COLUMN defect_assignees.created_by IS '创建人';
COMMENT ON COLUMN defect_assignees.created_at IS '创建时间';
COMMENT ON COLUMN defect_assignees.updated_by IS '更新人';
COMMENT ON COLUMN defect_assignees.updated_at IS '更新时间';

COMMENT ON TABLE defect_labels IS '缺陷标签 (主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN defect_labels.id IS '雪花ID';
COMMENT ON COLUMN defect_labels.tenant_id IS '租户ID';
COMMENT ON COLUMN defect_labels.workspace_id IS '工作空间ID';
COMMENT ON COLUMN defect_labels.project_id IS '项目ID';
COMMENT ON COLUMN defect_labels.defect_id IS '缺陷ID';
COMMENT ON COLUMN defect_labels.label_id IS '标签ID';
COMMENT ON COLUMN defect_labels.status IS '状态';
COMMENT ON COLUMN defect_labels.deleted IS '软删除';
COMMENT ON COLUMN defect_labels.created_by IS '创建人';
COMMENT ON COLUMN defect_labels.created_at IS '创建时间';
COMMENT ON COLUMN defect_labels.updated_by IS '更新人';
COMMENT ON COLUMN defect_labels.updated_at IS '更新时间';

COMMENT ON TABLE defect_modules IS '缺陷模块关联 (主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN defect_modules.id IS '雪花ID';
COMMENT ON COLUMN defect_modules.tenant_id IS '租户ID';
COMMENT ON COLUMN defect_modules.workspace_id IS '工作空间ID';
COMMENT ON COLUMN defect_modules.project_id IS '项目ID';
COMMENT ON COLUMN defect_modules.defect_id IS '缺陷ID';
COMMENT ON COLUMN defect_modules.module_id IS '模块ID';
COMMENT ON COLUMN defect_modules.status IS '状态';
COMMENT ON COLUMN defect_modules.deleted IS '软删除';
COMMENT ON COLUMN defect_modules.created_by IS '创建人';
COMMENT ON COLUMN defect_modules.created_at IS '创建时间';
COMMENT ON COLUMN defect_modules.updated_by IS '更新人';
COMMENT ON COLUMN defect_modules.updated_at IS '更新时间';

COMMENT ON TABLE defect_watchers IS '缺陷关注人 (主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN defect_watchers.id IS '雪花ID';
COMMENT ON COLUMN defect_watchers.tenant_id IS '租户ID';
COMMENT ON COLUMN defect_watchers.workspace_id IS '工作空间ID';
COMMENT ON COLUMN defect_watchers.project_id IS '项目ID';
COMMENT ON COLUMN defect_watchers.defect_id IS '缺陷ID';
COMMENT ON COLUMN defect_watchers.user_id IS '用户ID';
COMMENT ON COLUMN defect_watchers.status IS '状态';
COMMENT ON COLUMN defect_watchers.deleted IS '软删除';
COMMENT ON COLUMN defect_watchers.created_by IS '创建人';
COMMENT ON COLUMN defect_watchers.created_at IS '创建时间';
COMMENT ON COLUMN defect_watchers.updated_by IS '更新人';
COMMENT ON COLUMN defect_watchers.updated_at IS '更新时间';

COMMENT ON TABLE defect_relations IS '缺陷关联关系 (主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN defect_relations.id IS '雪花ID';
COMMENT ON COLUMN defect_relations.tenant_id IS '租户ID';
COMMENT ON COLUMN defect_relations.workspace_id IS '工作空间ID';
COMMENT ON COLUMN defect_relations.project_id IS '项目ID';
COMMENT ON COLUMN defect_relations.source_defect_id IS '源缺陷ID';
COMMENT ON COLUMN defect_relations.target_defect_id IS '目标缺陷ID';
COMMENT ON COLUMN defect_relations.relation_type IS '关系类型';
COMMENT ON COLUMN defect_relations.status IS '状态';
COMMENT ON COLUMN defect_relations.deleted IS '软删除';
COMMENT ON COLUMN defect_relations.created_by IS '创建人';
COMMENT ON COLUMN defect_relations.created_at IS '创建时间';
COMMENT ON COLUMN defect_relations.updated_by IS '更新人';
COMMENT ON COLUMN defect_relations.updated_at IS '更新时间';

COMMENT ON TABLE defect_comments IS '缺陷评论 (主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN defect_comments.id IS '雪花ID';
COMMENT ON COLUMN defect_comments.tenant_id IS '租户ID';
COMMENT ON COLUMN defect_comments.workspace_id IS '工作空间ID';
COMMENT ON COLUMN defect_comments.project_id IS '项目ID';
COMMENT ON COLUMN defect_comments.defect_id IS '缺陷ID';
COMMENT ON COLUMN defect_comments.content_json IS '内容 JSON';
COMMENT ON COLUMN defect_comments.content_html IS '内容 HTML';
COMMENT ON COLUMN defect_comments.parent_id IS '父评论ID';
COMMENT ON COLUMN defect_comments.status IS '状态';
COMMENT ON COLUMN defect_comments.deleted IS '软删除';
COMMENT ON COLUMN defect_comments.created_by IS '创建人';
COMMENT ON COLUMN defect_comments.created_at IS '创建时间';
COMMENT ON COLUMN defect_comments.updated_by IS '更新人';
COMMENT ON COLUMN defect_comments.updated_at IS '更新时间';
COMMENT ON COLUMN defect_comments.content_stripped IS '去标签后的纯文本内容';
COMMENT ON COLUMN defect_comments.mentions IS '@提及的用户ID列表';
COMMENT ON COLUMN defect_comments.is_edited IS '是否被编辑过';
COMMENT ON COLUMN defect_comments.edited_at IS '最后编辑时间';

COMMENT ON TABLE defect_activities IS '缺陷活动日志 (主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN defect_activities.id IS '雪花ID';
COMMENT ON COLUMN defect_activities.tenant_id IS '租户ID';
COMMENT ON COLUMN defect_activities.workspace_id IS '工作空间ID';
COMMENT ON COLUMN defect_activities.project_id IS '项目ID';
COMMENT ON COLUMN defect_activities.defect_id IS '缺陷ID';
COMMENT ON COLUMN defect_activities.verb IS '动作';
COMMENT ON COLUMN defect_activities.field_name IS '字段名';
COMMENT ON COLUMN defect_activities.old_value IS '旧值';
COMMENT ON COLUMN defect_activities.new_value IS '新值';
COMMENT ON COLUMN defect_activities.actor_id IS '操作人';
COMMENT ON COLUMN defect_activities.status IS '状态';
COMMENT ON COLUMN defect_activities.deleted IS '软删除';
COMMENT ON COLUMN defect_activities.created_by IS '创建人';
COMMENT ON COLUMN defect_activities.created_at IS '创建时间';
COMMENT ON COLUMN defect_activities.updated_by IS '更新人';
COMMENT ON COLUMN defect_activities.updated_at IS '更新时间';

COMMENT ON TABLE defect_timelogs IS '缺陷工时 (主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN defect_timelogs.id IS '雪花ID';
COMMENT ON COLUMN defect_timelogs.tenant_id IS '租户ID';
COMMENT ON COLUMN defect_timelogs.workspace_id IS '工作空间ID';
COMMENT ON COLUMN defect_timelogs.project_id IS '项目ID';
COMMENT ON COLUMN defect_timelogs.defect_id IS '缺陷ID';
COMMENT ON COLUMN defect_timelogs.user_id IS '记录人';
COMMENT ON COLUMN defect_timelogs.spent_date IS '日期';
COMMENT ON COLUMN defect_timelogs.duration_minutes IS '时长(分)';
COMMENT ON COLUMN defect_timelogs.description IS '描述';
COMMENT ON COLUMN defect_timelogs.status IS '状态';
COMMENT ON COLUMN defect_timelogs.deleted IS '软删除';
COMMENT ON COLUMN defect_timelogs.created_by IS '创建人';
COMMENT ON COLUMN defect_timelogs.created_at IS '创建时间';
COMMENT ON COLUMN defect_timelogs.updated_by IS '更新人';
COMMENT ON COLUMN defect_timelogs.updated_at IS '更新时间';

COMMENT ON TABLE defect_attachments IS '缺陷附件 (主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN defect_attachments.id IS '雪花ID';
COMMENT ON COLUMN defect_attachments.tenant_id IS '租户ID';
COMMENT ON COLUMN defect_attachments.workspace_id IS '工作空间ID';
COMMENT ON COLUMN defect_attachments.project_id IS '项目ID';
COMMENT ON COLUMN defect_attachments.defect_id IS '缺陷ID';
COMMENT ON COLUMN defect_attachments.file_name IS '文件名';
COMMENT ON COLUMN defect_attachments.file_size IS '文件大小';
COMMENT ON COLUMN defect_attachments.file_type IS '文件类型';
COMMENT ON COLUMN defect_attachments.storage_path IS '存储路径';
COMMENT ON COLUMN defect_attachments.thumbnail_path IS '缩略图路径';
COMMENT ON COLUMN defect_attachments.status IS '状态';
COMMENT ON COLUMN defect_attachments.deleted IS '软删除';
COMMENT ON COLUMN defect_attachments.created_by IS '创建人';
COMMENT ON COLUMN defect_attachments.created_at IS '创建时间';
COMMENT ON COLUMN defect_attachments.updated_by IS '更新人';
COMMENT ON COLUMN defect_attachments.updated_at IS '更新时间';

COMMENT ON TABLE defect_ext IS '缺陷扩展字段 (主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN defect_ext.id IS '雪花ID';
COMMENT ON COLUMN defect_ext.tenant_id IS '租户ID';
COMMENT ON COLUMN defect_ext.workspace_id IS '工作空间ID';
COMMENT ON COLUMN defect_ext.project_id IS '项目ID';
COMMENT ON COLUMN defect_ext.defect_id IS '缺陷ID';
COMMENT ON COLUMN defect_ext.field_name IS '字段名';
COMMENT ON COLUMN defect_ext.field_value IS '字段值';
COMMENT ON COLUMN defect_ext.field_schema IS '字段元数据';
COMMENT ON COLUMN defect_ext.status IS '状态';
COMMENT ON COLUMN defect_ext.deleted IS '软删除';
COMMENT ON COLUMN defect_ext.created_by IS '创建人';
COMMENT ON COLUMN defect_ext.created_at IS '创建时间';
COMMENT ON COLUMN defect_ext.updated_by IS '更新人';
COMMENT ON COLUMN defect_ext.updated_at IS '更新时间';

COMMENT ON TABLE biz_entity_relations IS '跨类型实体关联 (主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN biz_entity_relations.id IS '雪花ID';
COMMENT ON COLUMN biz_entity_relations.tenant_id IS '租户ID';
COMMENT ON COLUMN biz_entity_relations.workspace_id IS '工作空间ID';
COMMENT ON COLUMN biz_entity_relations.project_id IS '项目ID';
COMMENT ON COLUMN biz_entity_relations.source_type IS '源类型(task/requirement/defect)';
COMMENT ON COLUMN biz_entity_relations.source_id IS '源ID';
COMMENT ON COLUMN biz_entity_relations.target_type IS '目标类型';
COMMENT ON COLUMN biz_entity_relations.target_id IS '目标ID';
COMMENT ON COLUMN biz_entity_relations.relation_type IS '关系类型';
COMMENT ON COLUMN biz_entity_relations.status IS '状态';
COMMENT ON COLUMN biz_entity_relations.deleted IS '软删除';
COMMENT ON COLUMN biz_entity_relations.created_by IS '创建人';
COMMENT ON COLUMN biz_entity_relations.created_at IS '创建时间';
COMMENT ON COLUMN biz_entity_relations.updated_by IS '更新人';
COMMENT ON COLUMN biz_entity_relations.updated_at IS '更新时间';
COMMENT ON TABLE issue_dependencies IS '工作项依赖关系表（FS/SS/FF/SF + 滞后天数）(主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN issue_dependencies.id IS '雪花ID';
COMMENT ON COLUMN issue_dependencies.tenant_id IS '租户ID';
COMMENT ON COLUMN issue_dependencies.workspace_id IS '工作空间ID';
COMMENT ON COLUMN issue_dependencies.project_id IS '项目ID';
COMMENT ON COLUMN issue_dependencies.issue_id IS '工作项ID（依赖方）';
COMMENT ON COLUMN issue_dependencies.depends_on_id IS '被依赖工作项ID';
COMMENT ON COLUMN issue_dependencies.dependency_type IS '依赖类型（fs=完成-开始, ss=开始-开始, ff=完成-完成, sf=开始-完成）';
COMMENT ON COLUMN issue_dependencies.lag_days IS '滞后天数';
COMMENT ON COLUMN issue_dependencies.status IS '状态';
COMMENT ON COLUMN issue_dependencies.deleted IS '软删除';
COMMENT ON COLUMN issue_dependencies.created_by IS '创建人';
COMMENT ON COLUMN issue_dependencies.created_at IS '创建时间';
COMMENT ON COLUMN issue_dependencies.updated_by IS '更新人';
COMMENT ON COLUMN issue_dependencies.updated_at IS '更新时间';

COMMENT ON TABLE pages IS '项目文档页面 (主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN pages.id IS '雪花ID';
COMMENT ON COLUMN pages.public_id IS '公开ID';
COMMENT ON COLUMN pages.tenant_id IS '租户ID';
COMMENT ON COLUMN pages.workspace_id IS '工作空间ID';
COMMENT ON COLUMN pages.project_id IS '项目ID';
COMMENT ON COLUMN pages.name IS '页面名';
COMMENT ON COLUMN pages.description_json IS '描述 JSON';
COMMENT ON COLUMN pages.description_html IS '描述 HTML';
COMMENT ON COLUMN pages.description_stripped IS '纯文本';
COMMENT ON COLUMN pages.parent_id IS '父页面ID';
COMMENT ON COLUMN pages.sort_order IS '排序';
COMMENT ON COLUMN pages.version IS '乐观锁';
COMMENT ON COLUMN pages.status IS '状态';
COMMENT ON COLUMN pages.deleted IS '软删除';
COMMENT ON COLUMN pages.created_by IS '创建人';
COMMENT ON COLUMN pages.created_at IS '创建时间';
COMMENT ON COLUMN pages.updated_by IS '更新人';
COMMENT ON COLUMN pages.updated_at IS '更新时间';

COMMENT ON TABLE deployment_events IS '部署事件 (主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN deployment_events.id IS '雪花ID';
COMMENT ON COLUMN deployment_events.tenant_id IS '租户ID';
COMMENT ON COLUMN deployment_events.workspace_id IS '工作空间ID';
COMMENT ON COLUMN deployment_events.project_id IS '项目ID';
COMMENT ON COLUMN deployment_events.deployment_id IS '部署ID';
COMMENT ON COLUMN deployment_events.env IS '环境';
COMMENT ON COLUMN deployment_events.status IS '状态';
COMMENT ON COLUMN deployment_events.version IS '版本';
COMMENT ON COLUMN deployment_events.deployed_at IS '部署时间';
COMMENT ON COLUMN deployment_events.deleted IS '软删除';
COMMENT ON COLUMN deployment_events.created_by IS '创建人';
COMMENT ON COLUMN deployment_events.created_at IS '创建时间';
COMMENT ON COLUMN deployment_events.updated_by IS '更新人';
COMMENT ON COLUMN deployment_events.updated_at IS '更新时间';

COMMENT ON TABLE processed_events IS '事件消费记录 (主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN processed_events.event_id IS '事件ID';
COMMENT ON COLUMN processed_events.consumer_id IS '消费者ID';
COMMENT ON COLUMN processed_events.processed_at IS '处理时间';
COMMENT ON COLUMN processed_events.retry_count IS '重试次数';

COMMENT ON TABLE dlq_events IS '死信队列事件 (主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN dlq_events.id IS '雪花ID';
COMMENT ON COLUMN dlq_events.event_id IS '事件ID';
COMMENT ON COLUMN dlq_events.tenant_id IS '租户ID';
COMMENT ON COLUMN dlq_events.workspace_id IS '工作空间ID';
COMMENT ON COLUMN dlq_events.queue IS '队列';
COMMENT ON COLUMN dlq_events.exchange IS '交换机';
COMMENT ON COLUMN dlq_events.routing_key IS '路由键';
COMMENT ON COLUMN dlq_events.payload IS '载荷';
COMMENT ON COLUMN dlq_events.error_reason IS '错误原因';
COMMENT ON COLUMN dlq_events.resolved_at IS '解决时间';
COMMENT ON COLUMN dlq_events.resolved_by IS '解决人';
COMMENT ON COLUMN dlq_events.created_at IS '创建时间';

COMMENT ON TABLE password_reset_tokens IS '密码重置令牌 (主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN password_reset_tokens.id IS '雪花ID';
COMMENT ON COLUMN password_reset_tokens.user_id IS '用户ID';
COMMENT ON COLUMN password_reset_tokens.token_hash IS '令牌哈希';
COMMENT ON COLUMN password_reset_tokens.expires_at IS '过期时间';
COMMENT ON COLUMN password_reset_tokens.used_at IS '使用时间';
COMMENT ON COLUMN password_reset_tokens.status IS '状态';
COMMENT ON COLUMN password_reset_tokens.deleted IS '软删除';
COMMENT ON COLUMN password_reset_tokens.created_at IS '创建时间';

COMMENT ON TABLE idempotency_keys IS 'API 幂等键 (主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN idempotency_keys.key IS '幂等键';
COMMENT ON COLUMN idempotency_keys.user_id IS '用户ID';
COMMENT ON COLUMN idempotency_keys.response_status IS '响应状态码';
COMMENT ON COLUMN idempotency_keys.response_body IS '响应体';
COMMENT ON COLUMN idempotency_keys.expires_at IS '过期时间';
COMMENT ON COLUMN idempotency_keys.created_at IS '创建时间';

COMMENT ON TABLE schema_migrations IS '迁移版本 (主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN schema_migrations.version IS '版本号';
COMMENT ON COLUMN schema_migrations.dirty IS '脏标记';
COMMENT ON TABLE intake_channels IS '入口渠道 (主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN intake_channels.id IS '雪花ID';
COMMENT ON COLUMN intake_channels.code IS '渠道编码';
COMMENT ON COLUMN intake_channels.name IS '渠道名称';
COMMENT ON COLUMN intake_channels.slug IS 'URL标识';
COMMENT ON COLUMN intake_channels.tenant_id IS '租户ID';
COMMENT ON COLUMN intake_channels.workspace_id IS '工作空间ID';
COMMENT ON COLUMN intake_channels.project_id IS '项目ID';
COMMENT ON COLUMN intake_channels.description IS '描述';
COMMENT ON COLUMN intake_channels.is_active IS '是否启用';
COMMENT ON COLUMN intake_channels.config IS '渠道配置（JSONB）';
COMMENT ON COLUMN intake_channels.status IS '状态';
COMMENT ON COLUMN intake_channels.deleted IS '软删除';
COMMENT ON COLUMN intake_channels.created_by IS '创建人';
COMMENT ON COLUMN intake_channels.created_at IS '创建时间';
COMMENT ON COLUMN intake_channels.updated_by IS '更新人';
COMMENT ON COLUMN intake_channels.updated_at IS '更新时间';

COMMENT ON TABLE intake_issues IS '入口工单 (主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN intake_issues.id IS '雪花ID';
COMMENT ON COLUMN intake_issues.code IS '工单编码';
COMMENT ON COLUMN intake_issues.name IS '工单标题';
COMMENT ON COLUMN intake_issues.tenant_id IS '租户ID';
COMMENT ON COLUMN intake_issues.workspace_id IS '工作空间ID';
COMMENT ON COLUMN intake_issues.project_id IS '项目ID';
COMMENT ON COLUMN intake_issues.channel_id IS '渠道ID';
COMMENT ON COLUMN intake_issues.tracking_id IS '跟踪ID';
COMMENT ON COLUMN intake_issues.submitter_name IS '提交人姓名';
COMMENT ON COLUMN intake_issues.submitter_email IS '提交人邮箱';
COMMENT ON COLUMN intake_issues.description IS '描述';
COMMENT ON COLUMN intake_issues.priority IS '优先级（low/medium/high/urgent）';
COMMENT ON COLUMN intake_issues.status IS '状态（open/accepted/rejected/archived）';
COMMENT ON COLUMN intake_issues.linked_entity_type IS '关联实体类型（转正后关联）';
COMMENT ON COLUMN intake_issues.linked_entity_id IS '关联实体ID';
COMMENT ON COLUMN intake_issues.resolved_at IS '解决时间';
COMMENT ON COLUMN intake_issues.resolved_by IS '解决人';
COMMENT ON COLUMN intake_issues.deleted IS '软删除';
COMMENT ON COLUMN intake_issues.created_by IS '创建人';
COMMENT ON COLUMN intake_issues.created_at IS '创建时间';
COMMENT ON COLUMN intake_issues.updated_by IS '更新人';
COMMENT ON COLUMN intake_issues.updated_at IS '更新时间';

-- 以下为补齐的 11 张表（SQL 有、文档无）的注释


COMMENT ON TABLE defect_extra IS '缺陷扩展信息表（found_phase、reopened_at 等缺陷分析字段）(主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN defect_extra.id IS '雪花ID';
COMMENT ON COLUMN defect_extra.tenant_id IS '租户ID';
COMMENT ON COLUMN defect_extra.defect_id IS '缺陷ID';
COMMENT ON COLUMN defect_extra.found_phase IS '发现阶段（如 dev/test/prod）';
COMMENT ON COLUMN defect_extra.reopened_at IS '重开时间';
COMMENT ON COLUMN defect_extra.status IS '状态';
COMMENT ON COLUMN defect_extra.deleted IS '软删除';
COMMENT ON COLUMN defect_extra.created_by IS '创建人';
COMMENT ON COLUMN defect_extra.created_at IS '创建时间';
COMMENT ON COLUMN defect_extra.updated_at IS '更新时间';

COMMENT ON TABLE document_links IS '文档与业务实体关联关系表 (主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN document_links.id IS '雪花ID';
COMMENT ON COLUMN document_links.tenant_id IS '租户ID';
COMMENT ON COLUMN document_links.page_id IS '文档页面ID';
COMMENT ON COLUMN document_links.linkable_type IS '关联实体类型';
COMMENT ON COLUMN document_links.linkable_id IS '关联实体ID';
COMMENT ON COLUMN document_links.status IS '状态';
COMMENT ON COLUMN document_links.deleted IS '软删除';
COMMENT ON COLUMN document_links.created_by IS '创建人';
COMMENT ON COLUMN document_links.created_at IS '创建时间';
COMMENT ON COLUMN document_links.updated_at IS '更新时间';

COMMENT ON TABLE knowledge_page_relations IS '知识页与工作项关联关系表 (主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN knowledge_page_relations.id IS '雪花ID';
COMMENT ON COLUMN knowledge_page_relations.tenant_id IS '租户ID';
COMMENT ON COLUMN knowledge_page_relations.page_id IS '知识页ID';
COMMENT ON COLUMN knowledge_page_relations.workitem_id IS '工作项ID（需求/任务/缺陷）';
COMMENT ON COLUMN knowledge_page_relations.relation_type IS '关联类型（如 related/reference）';
COMMENT ON COLUMN knowledge_page_relations.status IS '状态';
COMMENT ON COLUMN knowledge_page_relations.deleted IS '软删除';
COMMENT ON COLUMN knowledge_page_relations.created_by IS '创建人';
COMMENT ON COLUMN knowledge_page_relations.created_at IS '创建时间';
COMMENT ON COLUMN knowledge_page_relations.updated_at IS '更新时间';

COMMENT ON TABLE knowledge_page_versions IS '知识页版本历史表 (主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN knowledge_page_versions.id IS '雪花ID';
COMMENT ON COLUMN knowledge_page_versions.tenant_id IS '租户ID';
COMMENT ON COLUMN knowledge_page_versions.page_id IS '知识页ID';
COMMENT ON COLUMN knowledge_page_versions.version IS '版本号';
COMMENT ON COLUMN knowledge_page_versions.title IS '版本标题';
COMMENT ON COLUMN knowledge_page_versions.content_md IS 'Markdown 内容';
COMMENT ON COLUMN knowledge_page_versions.content_html IS 'HTML 渲染内容';
COMMENT ON COLUMN knowledge_page_versions.change_summary IS '变更摘要';
COMMENT ON COLUMN knowledge_page_versions.status IS '状态';
COMMENT ON COLUMN knowledge_page_versions.deleted IS '软删除';
COMMENT ON COLUMN knowledge_page_versions.created_by IS '创建人';
COMMENT ON COLUMN knowledge_page_versions.created_at IS '创建时间';
COMMENT ON COLUMN knowledge_page_versions.updated_at IS '更新时间';

COMMENT ON TABLE page_shares IS '页面分享链接表（支持 token/密码/过期时间）(主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN page_shares.id IS '雪花ID';
COMMENT ON COLUMN page_shares.tenant_id IS '租户ID';
COMMENT ON COLUMN page_shares.page_id IS '页面ID';
COMMENT ON COLUMN page_shares.workspace_id IS '工作空间ID';
COMMENT ON COLUMN page_shares.project_id IS '项目ID';
COMMENT ON COLUMN page_shares.token IS '分享令牌';
COMMENT ON COLUMN page_shares.is_active IS '是否有效';
COMMENT ON COLUMN page_shares.password_hash IS '访问密码哈希';
COMMENT ON COLUMN page_shares.expires_at IS '过期时间';
COMMENT ON COLUMN page_shares.status IS '状态';
COMMENT ON COLUMN page_shares.deleted IS '软删除';
COMMENT ON COLUMN page_shares.created_by IS '创建人';
COMMENT ON COLUMN page_shares.created_at IS '创建时间';
COMMENT ON COLUMN page_shares.updated_at IS '更新时间';

COMMENT ON TABLE page_templates IS '页面模板表 (主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN page_templates.id IS '雪花ID';
COMMENT ON COLUMN page_templates.tenant_id IS '租户ID';
COMMENT ON COLUMN page_templates.workspace_id IS '工作空间ID';
COMMENT ON COLUMN page_templates.project_id IS '项目ID';
COMMENT ON COLUMN page_templates.name IS '模板名称';
COMMENT ON COLUMN page_templates.description IS '模板描述';
COMMENT ON COLUMN page_templates.content_html IS '模板 HTML 内容';
COMMENT ON COLUMN page_templates.category IS '模板分类';
COMMENT ON COLUMN page_templates.status IS '状态';
COMMENT ON COLUMN page_templates.deleted IS '软删除';
COMMENT ON COLUMN page_templates.created_by IS '创建人';
COMMENT ON COLUMN page_templates.created_at IS '创建时间';
COMMENT ON COLUMN page_templates.updated_by IS '更新人';
COMMENT ON COLUMN page_templates.updated_at IS '更新时间';

COMMENT ON TABLE role_permissions IS '角色-权限映射表（role_slug + permission_code）(主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN role_permissions.id IS '雪花ID';
COMMENT ON COLUMN role_permissions.tenant_id IS '租户ID';
COMMENT ON COLUMN role_permissions.role_slug IS '角色标识';
COMMENT ON COLUMN role_permissions.permission_code IS '权限编码';
COMMENT ON COLUMN role_permissions.status IS '状态';
COMMENT ON COLUMN role_permissions.deleted IS '软删除';
COMMENT ON COLUMN role_permissions.created_by IS '创建人';
COMMENT ON COLUMN role_permissions.created_at IS '创建时间';
COMMENT ON COLUMN role_permissions.updated_at IS '更新时间';

COMMENT ON TABLE sso_links IS 'SSO 用户与提供方关联表 (主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN sso_links.id IS '雪花ID';
COMMENT ON COLUMN sso_links.tenant_id IS '租户ID';
COMMENT ON COLUMN sso_links.user_id IS '用户ID';
COMMENT ON COLUMN sso_links.provider_id IS 'SSO 提供方ID';
COMMENT ON COLUMN sso_links.sso_subject IS 'SSO 主体标识';
COMMENT ON COLUMN sso_links.sso_email IS 'SSO 邮箱';
COMMENT ON COLUMN sso_links.sso_display_name IS 'SSO 显示名';
COMMENT ON COLUMN sso_links.last_login_at IS '最后登录时间';
COMMENT ON COLUMN sso_links.deleted IS '软删除';
COMMENT ON COLUMN sso_links.created_at IS '创建时间';
COMMENT ON COLUMN sso_links.updated_at IS '更新时间';

COMMENT ON TABLE sso_sessions IS 'SSO 认证会话表（OIDC/OAuth2 PKCE 流程状态）(主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN sso_sessions.id IS '雪花ID';
COMMENT ON COLUMN sso_sessions.tenant_id IS '租户ID';
COMMENT ON COLUMN sso_sessions.state IS 'OIDC state 参数';
COMMENT ON COLUMN sso_sessions.nonce IS 'OIDC nonce 参数';
COMMENT ON COLUMN sso_sessions.provider_id IS 'SSO 提供方ID';
COMMENT ON COLUMN sso_sessions.redirect_to IS '认证完成后重定向地址';
COMMENT ON COLUMN sso_sessions.ip_address IS '客户端 IP';
COMMENT ON COLUMN sso_sessions.user_agent IS '客户端 User-Agent';
COMMENT ON COLUMN sso_sessions.code_verifier IS 'PKCE code_verifier';
COMMENT ON COLUMN sso_sessions.status IS '状态（pending/success/failed）';
COMMENT ON COLUMN sso_sessions.user_id IS '关联用户ID';
COMMENT ON COLUMN sso_sessions.error_message IS '错误信息';
COMMENT ON COLUMN sso_sessions.expires_at IS '过期时间';
COMMENT ON COLUMN sso_sessions.completed_at IS '完成时间';
COMMENT ON COLUMN sso_sessions.deleted IS '软删除';
COMMENT ON COLUMN sso_sessions.created_at IS '创建时间';
COMMENT ON COLUMN sso_sessions.updated_at IS '更新时间';

COMMENT ON TABLE workbench_templates IS '工作台模板表 (主键为雪花ID/BIGINT, 应用层生成)';
COMMENT ON COLUMN workbench_templates.id IS '雪花ID';
COMMENT ON COLUMN workbench_templates.tenant_id IS '租户ID';
COMMENT ON COLUMN workbench_templates.name IS '模板名称';
COMMENT ON COLUMN workbench_templates.slug IS 'URL 标识';
COMMENT ON COLUMN workbench_templates.description IS '模板描述';
COMMENT ON COLUMN workbench_templates.layout IS '布局配置（JSONB）';
COMMENT ON COLUMN workbench_templates.icon IS '图标';
COMMENT ON COLUMN workbench_templates.is_default IS '是否默认';
COMMENT ON COLUMN workbench_templates.sort_order IS '排序权重';
COMMENT ON COLUMN workbench_templates.status IS '状态';
COMMENT ON COLUMN workbench_templates.deleted IS '软删除';
COMMENT ON COLUMN workbench_templates.created_by IS '创建人';
COMMENT ON COLUMN workbench_templates.created_at IS '创建时间';
COMMENT ON COLUMN workbench_templates.updated_by IS '更新人';
COMMENT ON COLUMN workbench_templates.updated_at IS '更新时间';


-- ===========================================================================

-- 第三部分: 触发器 (幂等安装)

-- ===========================================================================

DO $$
DECLARE
    tbl TEXT;
    tables TEXT[] := ARRAY[
    'tenants',
    'workspaces',
    'workspace_members',
    'users',
    'roles',
    'menus',
    'user_roles',
    'projects',
    'project_members',
    'project_sequences',
    'task',
    'requirement',
    'defect',
    'task_assignees',
    'task_labels',
    'task_modules',
    'task_watchers',
    'task_relations',
    'task_comments',
    'task_activities',
    'task_ext',
    'sprints',
    'sprint_snapshots',
    'versions',
    'version_delivery_snapshots',
    'states',
    'state_transitions',
    'modules',
    'labels',
    'estimate_points',
    'automation_rules',
    'rule_executions',
    'automation_templates',
    'dashboards',
    'dashboard_widgets',
    'dashboard_snapshots',
    'dashboard_templates',
    'notifications',
    'notification_deliveries',
    'notification_preferences',
    'notification_digests',
    'search_documents',
    'search_history',
    'search_bookmarks',
    'risk_rules',
    'risk_alerts',
    'metric_snapshots',
    'metric_adjustments',
    'webhooks',
    'webhook_logs',
    'workbench_configs',
    'view_preferences',
    'recent_items',
    'knowledge_spaces',
    'knowledge_pages',
    'sso_providers',
    'api_tokens',
    'audit_logs',
    'domain_events',
    'invitations',
    'tenant_members',
    'user_preferences',
    'role_menus',
    'project_configs',
    'version_sprint_relations',
    'sprint_requirements',
    'sprint_tasks',
    'sprint_defects',
    'content_templates',
    'reviews',
    'review_assignments',
    'documents',
    'document_versions',
    'share_links',
    'notification_subscriptions',
    'saved_views',
    'calendar_events',
    'data_jobs',
    'task_attachments',
    'task_timelogs',
    'requirement_assignees',
    'requirement_labels',
    'requirement_modules',
    'requirement_watchers',
    'requirement_relations',
    'requirement_comments',
    'requirement_activities',
    'requirement_timelogs',
    'requirement_attachments',
    'requirement_ext',
    'defect_assignees',
    'defect_labels',
    'defect_modules',
    'defect_watchers',
    'defect_relations',
    'defect_comments',
    'defect_activities',
    'defect_timelogs',
    'defect_attachments',
    'defect_ext',
    'biz_entity_relations',
    'pages',
    'intake_channels',
    'intake_issues',
    'deployment_events'
    ];
BEGIN
    FOREACH tbl IN ARRAY tables
    LOOP
        IF EXISTS (SELECT 1 FROM information_schema.tables
                   WHERE table_schema = 'public' AND table_name = tbl) THEN
            IF NOT EXISTS (SELECT 1 FROM pg_trigger
                           WHERE tgname = 'trg_' || tbl || '_updated_at'
                             AND tgrelid = ('public.' || tbl)::regclass) THEN
                EXECUTE format('CREATE TRIGGER trg_%s_updated_at BEFORE UPDATE ON %I
                                FOR EACH ROW EXECUTE FUNCTION set_updated_at()', tbl, tbl);
            END IF;
        END IF;
    END LOOP;
END $$;

DO $$
DECLARE
    tbl TEXT;
    tables TEXT[] := ARRAY[
    'projects',
    'task',
    'requirement',
    'defect',
    'sprints',
    'versions',
    'knowledge_pages',
    'calendar_events',
    'pages',
    'intake_channels',
    'intake_issues',
    'deployment_events'
    ];
BEGIN
    FOREACH tbl IN ARRAY tables
    LOOP
        IF EXISTS (SELECT 1 FROM information_schema.tables
                   WHERE table_schema = 'public' AND table_name = tbl) THEN
            IF NOT EXISTS (SELECT 1 FROM pg_trigger
                           WHERE tgname = 'trg_' || tbl || '_bump_version'
                             AND tgrelid = ('public.' || tbl)::regclass) THEN
                EXECUTE format('CREATE TRIGGER trg_%s_bump_version BEFORE UPDATE ON %I
                                FOR EACH ROW EXECUTE FUNCTION bump_version()', tbl, tbl);
            END IF;
        END IF;
    END LOOP;
END $$;

-- ===========================================================================
-- 逻辑外键索引与部分唯一索引
-- ===========================================================================

CREATE INDEX IF NOT EXISTS idx_tenants_owner_id ON tenants (owner_id) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_tenants_code ON tenants (code) WHERE code IS NOT NULL AND NOT deleted;

CREATE INDEX IF NOT EXISTS idx_workspaces_owner_id ON workspaces (tenant_id, owner_id) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_workspaces_code ON workspaces (tenant_id, code) WHERE code IS NOT NULL AND NOT deleted;

CREATE INDEX IF NOT EXISTS idx_workspace_members_user_id ON workspace_members (tenant_id, user_id) WHERE NOT deleted;
CREATE UNIQUE INDEX IF NOT EXISTS uq_workspace_members_tenant_id_workspace_id_user_id ON workspace_members (tenant_id, workspace_id, user_id) WHERE NOT deleted;

CREATE INDEX IF NOT EXISTS idx_users_code ON users (tenant_id, code) WHERE code IS NOT NULL AND NOT deleted;
CREATE UNIQUE INDEX IF NOT EXISTS uq_users_tenant_id_email ON users (tenant_id, email) WHERE NOT deleted;

CREATE INDEX IF NOT EXISTS idx_roles_code ON roles (code) WHERE code IS NOT NULL AND NOT deleted;

CREATE INDEX IF NOT EXISTS idx_menus_parent_id ON menus (parent_id) WHERE NOT deleted AND parent_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_menus_code ON menus (code) WHERE code IS NOT NULL AND NOT deleted;

CREATE INDEX IF NOT EXISTS idx_user_roles_user_id ON user_roles (tenant_id, user_id) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_user_roles_role_id ON user_roles (tenant_id, role_id) WHERE NOT deleted;

CREATE INDEX IF NOT EXISTS idx_projects_owner_id ON projects (tenant_id, owner_id) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_projects_code ON projects (tenant_id, code) WHERE code IS NOT NULL AND NOT deleted;

CREATE INDEX IF NOT EXISTS idx_project_members_user_id ON project_members (tenant_id, user_id) WHERE NOT deleted;

CREATE INDEX IF NOT EXISTS idx_project_sequences_sequence_id ON project_sequences (tenant_id, sequence_id) WHERE NOT deleted;

CREATE INDEX IF NOT EXISTS idx_task_sequence_id ON task (tenant_id, sequence_id) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_task_public_id ON task (tenant_id, public_id) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_task_parent_id ON task (parent_id) WHERE NOT deleted AND parent_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_task_state_id ON task (tenant_id, state_id) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_task_estimate_point_id ON task (tenant_id, estimate_point_id) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_task_sprint_id ON task (tenant_id, sprint_id) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_task_version_id ON task (tenant_id, version_id) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_task_code ON task (tenant_id, code) WHERE code IS NOT NULL AND NOT deleted;

CREATE INDEX IF NOT EXISTS idx_requirement_sequence_id ON requirement (tenant_id, sequence_id) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_requirement_public_id ON requirement (tenant_id, public_id) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_requirement_parent_id ON requirement (parent_id) WHERE NOT deleted AND parent_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_requirement_state_id ON requirement (tenant_id, state_id) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_requirement_estimate_point_id ON requirement (tenant_id, estimate_point_id) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_requirement_sprint_id ON requirement (tenant_id, sprint_id) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_requirement_version_id ON requirement (tenant_id, version_id) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_requirement_code ON requirement (tenant_id, code) WHERE code IS NOT NULL AND NOT deleted;

CREATE INDEX IF NOT EXISTS idx_defect_sequence_id ON defect (tenant_id, sequence_id) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_defect_public_id ON defect (tenant_id, public_id) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_defect_parent_id ON defect (parent_id) WHERE NOT deleted AND parent_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_defect_state_id ON defect (tenant_id, state_id) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_defect_found_version_id ON defect (tenant_id, found_version_id) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_defect_fix_version_id ON defect (tenant_id, fix_version_id) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_defect_verifier_id ON defect (tenant_id, verifier_id) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_defect_estimate_point_id ON defect (tenant_id, estimate_point_id) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_defect_sprint_id ON defect (tenant_id, sprint_id) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_defect_version_id ON defect (tenant_id, version_id) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_defect_code ON defect (tenant_id, code) WHERE code IS NOT NULL AND NOT deleted;

CREATE INDEX IF NOT EXISTS idx_task_assignees_task_id ON task_assignees (tenant_id, task_id) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_task_assignees_user_id ON task_assignees (tenant_id, user_id) WHERE NOT deleted;

CREATE INDEX IF NOT EXISTS idx_task_labels_task_id ON task_labels (tenant_id, task_id) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_task_labels_label_id ON task_labels (tenant_id, label_id) WHERE NOT deleted;

CREATE INDEX IF NOT EXISTS idx_task_modules_task_id ON task_modules (tenant_id, task_id) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_task_modules_module_id ON task_modules (tenant_id, module_id) WHERE NOT deleted;

CREATE INDEX IF NOT EXISTS idx_task_watchers_task_id ON task_watchers (tenant_id, task_id) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_task_watchers_user_id ON task_watchers (tenant_id, user_id) WHERE NOT deleted;

CREATE INDEX IF NOT EXISTS idx_task_relations_source_task_id ON task_relations (tenant_id, source_task_id) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_task_relations_target_task_id ON task_relations (tenant_id, target_task_id) WHERE NOT deleted;

CREATE INDEX IF NOT EXISTS idx_task_comments_task_id ON task_comments (tenant_id, task_id) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_task_comments_parent_id ON task_comments (parent_id) WHERE NOT deleted AND parent_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_task_activities_task_id ON task_activities (tenant_id, task_id) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_task_activities_actor_id ON task_activities (tenant_id, actor_id) WHERE NOT deleted;

CREATE INDEX IF NOT EXISTS idx_task_ext_task_id ON task_ext (tenant_id, task_id) WHERE NOT deleted;

CREATE INDEX IF NOT EXISTS idx_sprints_owner_id ON sprints (tenant_id, owner_id) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_sprints_version_id ON sprints (tenant_id, version_id) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_sprints_code ON sprints (tenant_id, code) WHERE code IS NOT NULL AND NOT deleted;

CREATE INDEX IF NOT EXISTS idx_sprint_snapshots_sprint_id ON sprint_snapshots (tenant_id, sprint_id) WHERE NOT deleted;

CREATE INDEX IF NOT EXISTS idx_versions_code ON versions (tenant_id, code) WHERE code IS NOT NULL AND NOT deleted;

CREATE INDEX IF NOT EXISTS idx_version_delivery_snapshots_version_id ON version_delivery_snapshots (tenant_id, version_id) WHERE NOT deleted;

CREATE INDEX IF NOT EXISTS idx_state_transitions_from_state_id ON state_transitions (tenant_id, from_state_id) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_state_transitions_to_state_id ON state_transitions (tenant_id, to_state_id) WHERE NOT deleted;

CREATE INDEX IF NOT EXISTS idx_modules_public_id ON modules (tenant_id, public_id) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_modules_lead_id ON modules (tenant_id, lead_id) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_modules_code ON modules (tenant_id, code) WHERE code IS NOT NULL AND NOT deleted;

CREATE INDEX IF NOT EXISTS idx_labels_code ON labels (tenant_id, code) WHERE code IS NOT NULL AND NOT deleted;

CREATE INDEX IF NOT EXISTS idx_estimate_points_code ON estimate_points (tenant_id, code) WHERE code IS NOT NULL AND NOT deleted;

CREATE INDEX IF NOT EXISTS idx_automation_rules_code ON automation_rules (tenant_id, code) WHERE code IS NOT NULL AND NOT deleted;

CREATE INDEX IF NOT EXISTS idx_rule_executions_rule_id ON rule_executions (tenant_id, rule_id) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_rule_executions_trigger_event_id ON rule_executions (tenant_id, trigger_event_id) WHERE NOT deleted;

CREATE INDEX IF NOT EXISTS idx_automation_templates_code ON automation_templates (code) WHERE code IS NOT NULL AND NOT deleted;

CREATE INDEX IF NOT EXISTS idx_dashboards_code ON dashboards (tenant_id, code) WHERE code IS NOT NULL AND NOT deleted;

CREATE INDEX IF NOT EXISTS idx_dashboard_widgets_dashboard_id ON dashboard_widgets (tenant_id, dashboard_id) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_dashboard_widgets_user_id ON dashboard_widgets (tenant_id, user_id) WHERE NOT deleted;

CREATE INDEX IF NOT EXISTS idx_dashboard_snapshots_dashboard_id ON dashboard_snapshots (tenant_id, dashboard_id) WHERE NOT deleted;

CREATE INDEX IF NOT EXISTS idx_dashboard_templates_code ON dashboard_templates (code) WHERE code IS NOT NULL AND NOT deleted;

CREATE INDEX IF NOT EXISTS idx_notifications_recipient_id ON notifications (tenant_id, recipient_id) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_notifications_actor_id ON notifications (tenant_id, actor_id) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_notifications_entity_id ON notifications (tenant_id, entity_id) WHERE NOT deleted;

CREATE INDEX IF NOT EXISTS idx_notification_deliveries_notification_id ON notification_deliveries (tenant_id, notification_id) WHERE NOT deleted;

CREATE INDEX IF NOT EXISTS idx_notification_preferences_user_id ON notification_preferences (tenant_id, user_id) WHERE NOT deleted;

CREATE INDEX IF NOT EXISTS idx_notification_digests_user_id ON notification_digests (tenant_id, user_id) WHERE NOT deleted;

CREATE INDEX IF NOT EXISTS idx_search_documents_doc_id ON search_documents (tenant_id, doc_id) WHERE NOT deleted;

CREATE INDEX IF NOT EXISTS idx_search_history_user_id ON search_history (tenant_id, user_id) WHERE NOT deleted;

CREATE INDEX IF NOT EXISTS idx_search_bookmarks_user_id ON search_bookmarks (tenant_id, user_id) WHERE NOT deleted;

CREATE INDEX IF NOT EXISTS idx_risk_rules_code ON risk_rules (tenant_id, code) WHERE code IS NOT NULL AND NOT deleted;

CREATE INDEX IF NOT EXISTS idx_risk_alerts_rule_id ON risk_alerts (tenant_id, rule_id) WHERE NOT deleted;

CREATE INDEX IF NOT EXISTS idx_metric_snapshots_ref_id ON metric_snapshots (tenant_id, ref_id) WHERE NOT deleted;

CREATE INDEX IF NOT EXISTS idx_metric_adjustments_snapshot_id ON metric_adjustments (tenant_id, snapshot_id) WHERE NOT deleted;

CREATE INDEX IF NOT EXISTS idx_webhooks_code ON webhooks (tenant_id, code) WHERE code IS NOT NULL AND NOT deleted;

CREATE INDEX IF NOT EXISTS idx_webhook_logs_webhook_id ON webhook_logs (tenant_id, webhook_id) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_webhook_logs_delivery_id ON webhook_logs (tenant_id, delivery_id) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_webhook_logs_event_id ON webhook_logs (tenant_id, event_id) WHERE NOT deleted;

CREATE INDEX IF NOT EXISTS idx_workbench_configs_user_id ON workbench_configs (tenant_id, user_id) WHERE NOT deleted;

CREATE INDEX IF NOT EXISTS idx_view_preferences_user_id ON view_preferences (tenant_id, user_id) WHERE NOT deleted;

CREATE INDEX IF NOT EXISTS idx_recent_items_user_id ON recent_items (tenant_id, user_id) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_recent_items_item_id ON recent_items (tenant_id, item_id) WHERE NOT deleted;

CREATE INDEX IF NOT EXISTS idx_knowledge_spaces_code ON knowledge_spaces (tenant_id, code) WHERE code IS NOT NULL AND NOT deleted;

CREATE INDEX IF NOT EXISTS idx_knowledge_pages_knowledge_space_id ON knowledge_pages (tenant_id, knowledge_space_id) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_knowledge_pages_public_id ON knowledge_pages (tenant_id, public_id) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_knowledge_pages_parent_id ON knowledge_pages (parent_id) WHERE NOT deleted AND parent_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_knowledge_pages_code ON knowledge_pages (tenant_id, code) WHERE code IS NOT NULL AND NOT deleted;

CREATE INDEX IF NOT EXISTS idx_sso_providers_client_id ON sso_providers (client_id) WHERE NOT deleted;

CREATE INDEX IF NOT EXISTS idx_api_tokens_user_id ON api_tokens (tenant_id, user_id) WHERE NOT deleted;

CREATE INDEX IF NOT EXISTS idx_audit_logs_actor_id ON audit_logs (tenant_id, actor_id) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_audit_logs_target_id ON audit_logs (tenant_id, target_id) WHERE NOT deleted;

CREATE INDEX IF NOT EXISTS idx_domain_events_aggregate_id ON domain_events (tenant_id, aggregate_id) WHERE NOT deleted;

CREATE INDEX IF NOT EXISTS idx_invitations_inviter_id ON invitations (tenant_id, inviter_id) WHERE NOT deleted;

CREATE INDEX IF NOT EXISTS idx_tenant_members_user_id ON tenant_members (tenant_id, user_id) WHERE NOT deleted;

CREATE INDEX IF NOT EXISTS idx_user_preferences_user_id ON user_preferences (tenant_id, user_id) WHERE NOT deleted;

CREATE INDEX IF NOT EXISTS idx_role_menus_role_id ON role_menus (role_id) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_role_menus_menu_id ON role_menus (menu_id) WHERE NOT deleted;

CREATE INDEX IF NOT EXISTS idx_version_sprint_relations_version_id ON version_sprint_relations (tenant_id, version_id) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_version_sprint_relations_sprint_id ON version_sprint_relations (tenant_id, sprint_id) WHERE NOT deleted;

CREATE INDEX IF NOT EXISTS idx_sprint_requirements_sprint_id ON sprint_requirements (tenant_id, sprint_id) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_sprint_requirements_requirement_id ON sprint_requirements (tenant_id, requirement_id) WHERE NOT deleted;

CREATE INDEX IF NOT EXISTS idx_sprint_tasks_sprint_id ON sprint_tasks (tenant_id, sprint_id) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_sprint_tasks_task_id ON sprint_tasks (tenant_id, task_id) WHERE NOT deleted;

CREATE INDEX IF NOT EXISTS idx_sprint_defects_sprint_id ON sprint_defects (tenant_id, sprint_id) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_sprint_defects_defect_id ON sprint_defects (tenant_id, defect_id) WHERE NOT deleted;

CREATE INDEX IF NOT EXISTS idx_review_assignments_review_id ON review_assignments (tenant_id, review_id) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_review_assignments_assignee_id ON review_assignments (tenant_id, assignee_id) WHERE NOT deleted;

CREATE INDEX IF NOT EXISTS idx_documents_public_id ON documents (tenant_id, public_id) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_documents_code ON documents (tenant_id, code) WHERE code IS NOT NULL AND NOT deleted;

CREATE INDEX IF NOT EXISTS idx_document_versions_document_id ON document_versions (tenant_id, document_id) WHERE NOT deleted;

CREATE INDEX IF NOT EXISTS idx_share_links_entity_id ON share_links (tenant_id, entity_id) WHERE NOT deleted;

CREATE INDEX IF NOT EXISTS idx_notification_subscriptions_user_id ON notification_subscriptions (tenant_id, user_id) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_notification_subscriptions_entity_id ON notification_subscriptions (tenant_id, entity_id) WHERE NOT deleted;

CREATE INDEX IF NOT EXISTS idx_saved_views_user_id ON saved_views (tenant_id, user_id) WHERE NOT deleted;

CREATE INDEX IF NOT EXISTS idx_calendar_events_source_id ON calendar_events (tenant_id, source_id) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_calendar_events_organizer_id ON calendar_events (tenant_id, organizer_id) WHERE NOT deleted;

CREATE INDEX IF NOT EXISTS idx_task_attachments_task_id ON task_attachments (tenant_id, task_id) WHERE NOT deleted;

CREATE INDEX IF NOT EXISTS idx_task_timelogs_task_id ON task_timelogs (tenant_id, task_id) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_task_timelogs_user_id ON task_timelogs (tenant_id, user_id) WHERE NOT deleted;

CREATE INDEX IF NOT EXISTS idx_requirement_assignees_requirement_id ON requirement_assignees (tenant_id, requirement_id) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_requirement_assignees_user_id ON requirement_assignees (tenant_id, user_id) WHERE NOT deleted;

CREATE INDEX IF NOT EXISTS idx_requirement_labels_requirement_id ON requirement_labels (tenant_id, requirement_id) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_requirement_labels_label_id ON requirement_labels (tenant_id, label_id) WHERE NOT deleted;

CREATE INDEX IF NOT EXISTS idx_requirement_modules_requirement_id ON requirement_modules (tenant_id, requirement_id) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_requirement_modules_module_id ON requirement_modules (tenant_id, module_id) WHERE NOT deleted;

CREATE INDEX IF NOT EXISTS idx_requirement_watchers_requirement_id ON requirement_watchers (tenant_id, requirement_id) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_requirement_watchers_user_id ON requirement_watchers (tenant_id, user_id) WHERE NOT deleted;

CREATE INDEX IF NOT EXISTS idx_requirement_relations_source_requirement_id ON requirement_relations (tenant_id, source_requirement_id) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_requirement_relations_target_requirement_id ON requirement_relations (tenant_id, target_requirement_id) WHERE NOT deleted;

CREATE INDEX IF NOT EXISTS idx_requirement_comments_requirement_id ON requirement_comments (tenant_id, requirement_id) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_requirement_comments_parent_id ON requirement_comments (parent_id) WHERE NOT deleted AND parent_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_requirement_activities_requirement_id ON requirement_activities (tenant_id, requirement_id) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_requirement_activities_actor_id ON requirement_activities (tenant_id, actor_id) WHERE NOT deleted;

CREATE INDEX IF NOT EXISTS idx_requirement_timelogs_requirement_id ON requirement_timelogs (tenant_id, requirement_id) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_requirement_timelogs_user_id ON requirement_timelogs (tenant_id, user_id) WHERE NOT deleted;

CREATE INDEX IF NOT EXISTS idx_requirement_attachments_requirement_id ON requirement_attachments (tenant_id, requirement_id) WHERE NOT deleted;

CREATE INDEX IF NOT EXISTS idx_requirement_ext_requirement_id ON requirement_ext (tenant_id, requirement_id) WHERE NOT deleted;

CREATE INDEX IF NOT EXISTS idx_defect_assignees_defect_id ON defect_assignees (tenant_id, defect_id) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_defect_assignees_user_id ON defect_assignees (tenant_id, user_id) WHERE NOT deleted;

CREATE INDEX IF NOT EXISTS idx_defect_labels_defect_id ON defect_labels (tenant_id, defect_id) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_defect_labels_label_id ON defect_labels (tenant_id, label_id) WHERE NOT deleted;

CREATE INDEX IF NOT EXISTS idx_defect_modules_defect_id ON defect_modules (tenant_id, defect_id) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_defect_modules_module_id ON defect_modules (tenant_id, module_id) WHERE NOT deleted;

CREATE INDEX IF NOT EXISTS idx_defect_watchers_defect_id ON defect_watchers (tenant_id, defect_id) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_defect_watchers_user_id ON defect_watchers (tenant_id, user_id) WHERE NOT deleted;

CREATE INDEX IF NOT EXISTS idx_defect_relations_source_defect_id ON defect_relations (tenant_id, source_defect_id) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_defect_relations_target_defect_id ON defect_relations (tenant_id, target_defect_id) WHERE NOT deleted;

CREATE INDEX IF NOT EXISTS idx_defect_comments_defect_id ON defect_comments (tenant_id, defect_id) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_defect_comments_parent_id ON defect_comments (parent_id) WHERE NOT deleted AND parent_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_defect_activities_defect_id ON defect_activities (tenant_id, defect_id) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_defect_activities_actor_id ON defect_activities (tenant_id, actor_id) WHERE NOT deleted;

CREATE INDEX IF NOT EXISTS idx_defect_timelogs_defect_id ON defect_timelogs (tenant_id, defect_id) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_defect_timelogs_user_id ON defect_timelogs (tenant_id, user_id) WHERE NOT deleted;

CREATE INDEX IF NOT EXISTS idx_defect_attachments_defect_id ON defect_attachments (tenant_id, defect_id) WHERE NOT deleted;

CREATE INDEX IF NOT EXISTS idx_defect_ext_defect_id ON defect_ext (tenant_id, defect_id) WHERE NOT deleted;

CREATE INDEX IF NOT EXISTS idx_biz_entity_relations_source_id ON biz_entity_relations (tenant_id, source_id) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_biz_entity_relations_target_id ON biz_entity_relations (tenant_id, target_id) WHERE NOT deleted;

CREATE INDEX IF NOT EXISTS idx_pages_public_id ON pages (tenant_id, public_id) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_pages_parent_id ON pages (parent_id) WHERE NOT deleted AND parent_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_deployment_events_deployment_id ON deployment_events (tenant_id, deployment_id) WHERE NOT deleted;

CREATE INDEX IF NOT EXISTS idx_processed_events_event_id ON processed_events (event_id) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_processed_events_consumer_id ON processed_events (consumer_id) WHERE NOT deleted;

CREATE INDEX IF NOT EXISTS idx_dlq_events_event_id ON dlq_events (tenant_id, event_id) WHERE NOT deleted;

CREATE INDEX IF NOT EXISTS idx_password_reset_tokens_user_id ON password_reset_tokens (user_id) WHERE NOT deleted;

CREATE INDEX IF NOT EXISTS idx_idempotency_keys_user_id ON idempotency_keys (user_id) WHERE NOT deleted;

-- ===========================================================================
-- 常用查询索引
-- ===========================================================================

-- 常用查询索引(核心业务路径, 可按需裁剪)
CREATE INDEX IF NOT EXISTS idx_task_scope ON task (tenant_id, workspace_id, project_id) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_task_sprint ON task (tenant_id, project_id, sprint_id) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_task_board ON task (project_id, state_id, sort_order) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_requirement_scope ON requirement (tenant_id, workspace_id, project_id) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_requirement_sprint ON requirement (tenant_id, project_id, sprint_id) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_requirement_board ON requirement (project_id, state_id, sort_order) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_defect_scope ON defect (tenant_id, workspace_id, project_id) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_defect_sprint ON defect (tenant_id, project_id, sprint_id) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_defect_board ON defect (project_id, state_id, sort_order) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_sprints_status ON sprints (tenant_id, project_id, status) WHERE NOT deleted;
CREATE UNIQUE INDEX IF NOT EXISTS uq_sprints_active ON sprints (tenant_id, project_id) WHERE status = 'active' AND NOT deleted;
CREATE INDEX IF NOT EXISTS idx_versions_project ON versions (tenant_id, project_id) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_states_project ON states (tenant_id, project_id) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_state_transitions_lookup ON state_transitions (tenant_id, project_id, type_code, from_state_id) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_search_documents_tsv ON search_documents USING gin (search_tsv);
CREATE INDEX IF NOT EXISTS idx_search_documents_type ON search_documents (tenant_id, workspace_id, doc_type) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_notifications_recipient ON notifications (tenant_id, recipient_id, is_read, created_at DESC) WHERE is_archived = false;
CREATE INDEX IF NOT EXISTS idx_biz_relations_source ON biz_entity_relations (tenant_id, source_type, source_id) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_biz_relations_target ON biz_entity_relations (tenant_id, target_type, target_id) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_task_activities_lookup ON task_activities (tenant_id, project_id, task_id, created_at DESC) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_requirement_activities_lookup ON requirement_activities (tenant_id, project_id, requirement_id, created_at DESC) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_defect_activities_lookup ON defect_activities (tenant_id, project_id, defect_id, created_at DESC) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_workbench_configs_user ON workbench_configs (tenant_id, user_id) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_documents_project ON documents (tenant_id, workspace_id, project_id) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_calendar_events_time ON calendar_events (tenant_id, workspace_id, start_time) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_data_jobs_status ON data_jobs (tenant_id, workspace_id, job_type, status) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_saved_views_user ON saved_views (tenant_id, workspace_id, user_id, view_type) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_notif_subs_user ON notification_subscriptions (tenant_id, user_id) WHERE NOT deleted;

-- 补齐: 11 张游离表的索引
CREATE INDEX IF NOT EXISTS idx_defect_extra_defect_id ON defect_extra (tenant_id, defect_id) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_document_links_page_id ON document_links (tenant_id, page_id) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_document_links_linkable ON document_links (tenant_id, linkable_type, linkable_id) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_knowledge_page_relations_page ON knowledge_page_relations (tenant_id, page_id) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_knowledge_page_relations_workitem ON knowledge_page_relations (tenant_id, workitem_id) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_knowledge_page_versions_page ON knowledge_page_versions (tenant_id, page_id) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_page_shares_page ON page_shares (tenant_id, page_id) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_page_shares_token ON page_shares (token) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_page_templates_project ON page_templates (tenant_id, project_id) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_role_permissions_role ON role_permissions (tenant_id, role_slug) WHERE NOT deleted;
CREATE UNIQUE INDEX IF NOT EXISTS uq_role_permissions ON role_permissions (tenant_id, role_slug, permission_code) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_sso_links_user ON sso_links (tenant_id, user_id) WHERE NOT deleted;
CREATE UNIQUE INDEX IF NOT EXISTS uq_sso_links_provider_subject ON sso_links (tenant_id, provider_id, sso_subject) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_sso_sessions_state ON sso_sessions (state) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_workbench_templates_tenant ON workbench_templates (tenant_id) WHERE NOT deleted;

-- 补齐: intake 表的索引
CREATE INDEX IF NOT EXISTS idx_intake_channels_tenant ON intake_channels (tenant_id, workspace_id) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_intake_channels_slug ON intake_channels (tenant_id, slug) WHERE NOT deleted;
CREATE UNIQUE INDEX IF NOT EXISTS uq_intake_channels_tenant_slug ON intake_channels (tenant_id, slug) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_intake_issues_tenant ON intake_issues (tenant_id, workspace_id) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_intake_issues_channel ON intake_issues (tenant_id, channel_id) WHERE NOT deleted;
CREATE INDEX IF NOT EXISTS idx_intake_issues_tracking ON intake_issues (tenant_id, tracking_id) WHERE tracking_id IS NOT NULL AND NOT deleted;
CREATE INDEX IF NOT EXISTS idx_intake_issues_submitter ON intake_issues (tenant_id, submitter_email) WHERE NOT deleted;


-- ===========================================================================
-- 种子数据 (最小化)
-- 说明: 应用级种子(系统角色/菜单/状态/模板等)在应用启动迁移中维护, 此处仅登记迁移版本。
-- ===========================================================================

INSERT INTO schema_migrations (version, dirty) VALUES (1, false) ON CONFLICT DO NOTHING;
-- ===========================================================================
-- 初始化完毕
-- ===========================================================================


-- ===========================================================================
-- 补充表（代码需要、设计文档遗漏的支撑表）
-- ===========================================================================
CREATE TABLE IF NOT EXISTS sso_sessions (
    id                       BIGINT PRIMARY KEY,
    tenant_id                BIGINT NOT NULL DEFAULT 1,
    state                    VARCHAR(255) NOT NULL,
    nonce                    VARCHAR(255),
    provider_id              BIGINT,
    redirect_to              TEXT,
    ip_address               VARCHAR(64),
    user_agent               TEXT,
    code_verifier            TEXT,
    status                   VARCHAR(20) DEFAULT 'pending',
    user_id                  BIGINT,
    error_message            TEXT,
    expires_at               TIMESTAMPTZ,
    completed_at             TIMESTAMPTZ,
    deleted                  BOOLEAN DEFAULT false,
    created_at               TIMESTAMPTZ DEFAULT now(),
    updated_at               TIMESTAMPTZ DEFAULT now()
);
CREATE TABLE IF NOT EXISTS sso_links (
    id                       BIGINT PRIMARY KEY,
    tenant_id                BIGINT NOT NULL DEFAULT 1,
    user_id                  BIGINT NOT NULL,
    provider_id              BIGINT NOT NULL,
    sso_subject              VARCHAR(255) NOT NULL,
    sso_email                VARCHAR(255),
    sso_display_name         VARCHAR(255),
    last_login_at            TIMESTAMPTZ,
    deleted                  BOOLEAN DEFAULT false,
    created_at               TIMESTAMPTZ DEFAULT now(),
    updated_at               TIMESTAMPTZ DEFAULT now()
);
CREATE TABLE IF NOT EXISTS workbench_templates (
    id                       BIGINT PRIMARY KEY,
    tenant_id                BIGINT NOT NULL DEFAULT 1,
    name                     VARCHAR(255) NOT NULL,
    slug                     VARCHAR(100) NOT NULL,
    description              TEXT,
    layout                   JSONB,
    icon                     VARCHAR(100),
    is_default               BOOLEAN DEFAULT false,
    sort_order               DOUBLE PRECISION DEFAULT 65535,
    status                   entity_status NOT NULL DEFAULT 'active',
    deleted                  BOOLEAN DEFAULT false,
    created_by               BIGINT NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ DEFAULT now(),
    updated_by               BIGINT NOT NULL DEFAULT 0,
    updated_at               TIMESTAMPTZ DEFAULT now()
);
CREATE TABLE IF NOT EXISTS knowledge_page_versions (
    id                       BIGINT PRIMARY KEY,
    tenant_id                BIGINT NOT NULL DEFAULT 1,
    page_id                  BIGINT NOT NULL,
    version                  INTEGER NOT NULL DEFAULT 1,
    title                    VARCHAR(255),
    content_md               TEXT,
    content_html             TEXT,
    change_summary           TEXT,
    status                   entity_status NOT NULL DEFAULT 'active',
    deleted                  BOOLEAN DEFAULT false,
    created_by               BIGINT NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ DEFAULT now(),
    updated_at               TIMESTAMPTZ DEFAULT now()
);
CREATE TABLE IF NOT EXISTS knowledge_page_relations (
    id                       BIGINT PRIMARY KEY,
    tenant_id                BIGINT NOT NULL DEFAULT 1,
    page_id                  BIGINT NOT NULL,
    workitem_id              BIGINT NOT NULL,
    relation_type            VARCHAR(50),
    status                   entity_status NOT NULL DEFAULT 'active',
    deleted                  BOOLEAN DEFAULT false,
    created_by               BIGINT NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ DEFAULT now(),
    updated_at               TIMESTAMPTZ DEFAULT now(),
    UNIQUE (page_id, workitem_id, relation_type)
);
CREATE TABLE IF NOT EXISTS document_links (
    id                       BIGINT PRIMARY KEY,
    tenant_id                BIGINT NOT NULL DEFAULT 1,
    page_id                  BIGINT NOT NULL,
    linkable_type            VARCHAR(50) NOT NULL,
    linkable_id              BIGINT NOT NULL,
    status                   entity_status NOT NULL DEFAULT 'active',
    deleted                  BOOLEAN DEFAULT false,
    created_by               BIGINT NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ DEFAULT now(),
    updated_at               TIMESTAMPTZ DEFAULT now()
);
CREATE TABLE IF NOT EXISTS page_templates (
    id                       BIGINT PRIMARY KEY,
    tenant_id                BIGINT NOT NULL DEFAULT 1,
    workspace_id             BIGINT NOT NULL DEFAULT 0,
    project_id               BIGINT NOT NULL DEFAULT 0,
    name                     VARCHAR(255) NOT NULL,
    description              TEXT,
    content_html             TEXT,
    category                 VARCHAR(100),
    status                   entity_status NOT NULL DEFAULT 'active',
    deleted                  BOOLEAN DEFAULT false,
    created_by               BIGINT NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ DEFAULT now(),
    updated_by               BIGINT NOT NULL DEFAULT 0,
    updated_at               TIMESTAMPTZ DEFAULT now()
);
CREATE TABLE IF NOT EXISTS page_shares (
    id                       BIGINT PRIMARY KEY,
    tenant_id                BIGINT NOT NULL DEFAULT 1,
    page_id                  BIGINT NOT NULL,
    workspace_id             BIGINT NOT NULL DEFAULT 0,
    project_id               BIGINT NOT NULL DEFAULT 0,
    token                    VARCHAR(128),
    is_active                BOOLEAN DEFAULT true,
    password_hash            VARCHAR(255),
    expires_at               TIMESTAMPTZ,
    status                   entity_status NOT NULL DEFAULT 'active',
    deleted                  BOOLEAN DEFAULT false,
    created_by               BIGINT NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ DEFAULT now(),
    updated_at               TIMESTAMPTZ DEFAULT now()
);
CREATE TABLE IF NOT EXISTS role_permissions (
    id                       BIGINT PRIMARY KEY,
    tenant_id                BIGINT NOT NULL DEFAULT 1,
    role_slug                VARCHAR(100) NOT NULL,
    permission_code          VARCHAR(100) NOT NULL,
    status                   entity_status NOT NULL DEFAULT 'active',
    deleted                  BOOLEAN DEFAULT false,
    created_by               BIGINT NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ DEFAULT now(),
    updated_at               TIMESTAMPTZ DEFAULT now()
);
CREATE TABLE IF NOT EXISTS defect_extra (
    id                       BIGINT PRIMARY KEY,
    tenant_id                BIGINT NOT NULL DEFAULT 1,
    defect_id                BIGINT NOT NULL,
    found_phase              VARCHAR(20),
    reopened_at              TIMESTAMPTZ,
    status                   entity_status NOT NULL DEFAULT 'active',
    deleted                  BOOLEAN DEFAULT false,
    created_by               BIGINT NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ DEFAULT now(),
    updated_at               TIMESTAMPTZ DEFAULT now()
);

--   120. intake_channels — 入口渠道
CREATE TABLE IF NOT EXISTS intake_channels (
    id                       BIGINT PRIMARY KEY,
    code                     VARCHAR(50),
    name                     VARCHAR(255) NOT NULL,
    slug                     VARCHAR(100) NOT NULL,
    tenant_id                BIGINT NOT NULL DEFAULT 1,
    workspace_id             BIGINT NOT NULL DEFAULT 0,
    project_id               BIGINT,
    description              TEXT,
    is_active                BOOLEAN DEFAULT true,
    config                   JSONB DEFAULT '{}',
    status                   entity_status NOT NULL DEFAULT 'active',
    deleted                  BOOLEAN DEFAULT false,
    created_by               BIGINT NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ DEFAULT now(),
    updated_by               BIGINT NOT NULL DEFAULT 0,
    updated_at               TIMESTAMPTZ DEFAULT now()
);

--   121. intake_issues — 入口工单
CREATE TABLE IF NOT EXISTS intake_issues (
    id                       BIGINT PRIMARY KEY,
    code                     VARCHAR(50),
    name                     VARCHAR(255) NOT NULL,
    tenant_id                BIGINT NOT NULL DEFAULT 1,
    workspace_id             BIGINT NOT NULL DEFAULT 0,
    project_id               BIGINT,
    channel_id               BIGINT NOT NULL DEFAULT 0,
    tracking_id              VARCHAR(50),
    submitter_name           VARCHAR(255),
    submitter_email          VARCHAR(255) NOT NULL,
    description              TEXT,
    priority                 VARCHAR(20) DEFAULT 'medium',
    status                   intake_issue_status NOT NULL DEFAULT 'open',
    linked_entity_type       VARCHAR(50),
    linked_entity_id         BIGINT,
    resolved_at              TIMESTAMPTZ,
    resolved_by              BIGINT,
    deleted                  BOOLEAN DEFAULT false,
    created_by               BIGINT NOT NULL DEFAULT 0,
    created_at               TIMESTAMPTZ DEFAULT now(),
    updated_by               BIGINT NOT NULL DEFAULT 0,
    updated_at               TIMESTAMPTZ DEFAULT now()
);

