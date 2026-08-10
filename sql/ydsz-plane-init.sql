-- ===========================================================================
-- Ydsz Plane 数据库初始化脚本（V2.0 最终版）
-- 数据库: PostgreSQL 16+
-- 设计要点:
--   1. 使用 PostgreSQL ENUM 类型定义所有状态字段
--   2. 三元数据隔离: tenants → workspaces → projects → work_items
--   3. 所有业务表统一携带 tenant_id (系统级表除外)
--   4. 主键使用 BIGINT + 雪花ID (应用层生成)
--   5. 字段排序标准化: id/code/name 开头, status/deleted/tenant_id/created_by/created_at/updated_by/updated_at 结尾
-- ===========================================================================

-- ===========================================================================
-- 第一部分: ENUM 类型定义
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
    CREATE TYPE sprint_status AS ENUM ('planned', 'active', 'completed');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

DO $$ BEGIN
    CREATE TYPE version_status AS ENUM ('planning', 'active', 'released', 'archived');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

DO $$ BEGIN
    CREATE TYPE intake_issue_status AS ENUM ('open', 'accepted', 'rejected', 'archived');
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
    CREATE TYPE work_item_status AS ENUM ('draft', 'active', 'completed', 'cancelled');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

DO $$ BEGIN
    CREATE TYPE attachment_status AS ENUM ('uploading', 'available', 'archived', 'deleted');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

DO $$ BEGIN
    CREATE TYPE webhook_delivery_status AS ENUM ('success', 'failed', 'retrying', 'pending');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

-- ===========================================================================
-- 第二部分: 触发器函数
-- ===========================================================================

CREATE OR REPLACE FUNCTION set_updated_at() RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION bump_version() RETURNS TRIGGER AS $$
BEGIN
    NEW.version = OLD.version + 1;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- ===========================================================================
-- 第三部分: 租户域表（根表）
-- ===========================================================================

-- 租户表（组织级根表，不含 tenant_id）
CREATE TABLE IF NOT EXISTS tenants (
    id              BIGINT PRIMARY KEY,
    code            VARCHAR(50) UNIQUE,
    name            VARCHAR(255) NOT NULL,
    slug            VARCHAR(100) NOT NULL UNIQUE,
    logo_url        TEXT,
    owner_id        BIGINT NOT NULL,
    timezone        VARCHAR(50) NOT NULL DEFAULT 'Asia/Shanghai',
    language        VARCHAR(20) NOT NULL DEFAULT 'zh-CN',
    brand_config    JSONB,
    status          tenant_status NOT NULL DEFAULT 'active',
    max_projects    INT DEFAULT 10,
    max_users       INT DEFAULT 50,
    max_workspaces  INT DEFAULT 5,
    expired_at      TIMESTAMPTZ,
    config          JSONB,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

COMMENT ON TABLE tenants IS '租户表（组织级实体：公司/部门/团队，顶层数据隔离根）';

-- ===========================================================================
-- 第四部分: 工作空间域
-- ===========================================================================

CREATE TABLE IF NOT EXISTS workspaces (
    id              BIGINT PRIMARY KEY,
    code            VARCHAR(50),
    name            VARCHAR(255) NOT NULL,
    slug            VARCHAR(100) NOT NULL,
    tenant_id       BIGINT NOT NULL,
    logo_url        TEXT,
    owner_id        BIGINT NOT NULL,
    timezone        VARCHAR(50) NOT NULL DEFAULT 'Asia/Shanghai',
    language        VARCHAR(20) NOT NULL DEFAULT 'zh-CN',
    status          entity_status NOT NULL DEFAULT 'active',
    max_projects    INT DEFAULT 20,
    config          JSONB,
    deleted         BOOLEAN NOT NULL DEFAULT false,
    created_by      BIGINT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_by      BIGINT NOT NULL,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, slug)
);

COMMENT ON TABLE workspaces IS '工作空间表（租户下的协作空间，一个空间包含多个项目）';

CREATE TABLE IF NOT EXISTS workspace_members (
    id              BIGINT PRIMARY KEY,
    tenant_id       BIGINT NOT NULL,
    workspace_id    BIGINT NOT NULL,
    user_id         BIGINT NOT NULL,
    role            workspace_role_enum NOT NULL DEFAULT 'member',
    status          entity_status NOT NULL DEFAULT 'active',
    joined_at       TIMESTAMPTZ DEFAULT now(),
    deleted         BOOLEAN NOT NULL DEFAULT false,
    created_by      BIGINT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_by      BIGINT NOT NULL,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, workspace_id, user_id)
);

COMMENT ON TABLE workspace_members IS '工作空间成员表（user↔workspace 多对多）';

-- ===========================================================================
-- 第五部分: 用户域
-- ===========================================================================

CREATE TABLE IF NOT EXISTS users (
    id              BIGINT PRIMARY KEY,
    code            VARCHAR(50),
    name            VARCHAR(255) NOT NULL,
    email           VARCHAR(255) NOT NULL,
    phone           VARCHAR(20),
    password_hash   TEXT,
    avatar_url      TEXT,
    status          user_status NOT NULL DEFAULT 'active',
    is_super_admin  BOOLEAN DEFAULT false,
    last_login_at   TIMESTAMPTZ,
    tenant_id       BIGINT NOT NULL,
    deleted         BOOLEAN NOT NULL DEFAULT false,
    created_by      BIGINT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_by      BIGINT NOT NULL,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, email),
    UNIQUE (tenant_id, phone)
);

COMMENT ON TABLE users IS '用户表（跨工作空间，认证信息）';

CREATE TABLE IF NOT EXISTS roles (
    id              BIGINT PRIMARY KEY,
    code            VARCHAR(50),
    name            VARCHAR(255) NOT NULL,
    description     TEXT,
    status          entity_status NOT NULL DEFAULT 'active',
    is_system       BOOLEAN DEFAULT false,
    role_scope      VARCHAR(50) DEFAULT 'tenant' CHECK (role_scope IN ('tenant', 'project')),
    deleted         BOOLEAN NOT NULL DEFAULT false,
    created_by      BIGINT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_by      BIGINT NOT NULL,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

COMMENT ON TABLE roles IS '角色表（系统级表，预置+自定义）';

CREATE TABLE IF NOT EXISTS menus (
    id              BIGINT PRIMARY KEY,
    code            VARCHAR(100) NOT NULL UNIQUE,
    name            VARCHAR(255) NOT NULL,
    menu_type       VARCHAR(20) NOT NULL CHECK (menu_type IN ('menu', 'button', 'api')),
    parent_id       BIGINT,
    path            VARCHAR(255),
    icon            VARCHAR(100),
    sort_order      INT DEFAULT 0,
    status          entity_status DEFAULT 'active',
    deleted         BOOLEAN DEFAULT false,
    created_by      BIGINT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_by      BIGINT NOT NULL,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

COMMENT ON TABLE menus IS '菜单/权限资源表（系统级表，不含tenant_id）';

CREATE TABLE IF NOT EXISTS user_roles (
    id              BIGINT PRIMARY KEY,
    tenant_id       BIGINT NOT NULL,
    user_id         BIGINT NOT NULL,
    role_id         BIGINT NOT NULL,
    status          entity_status NOT NULL DEFAULT 'active',
    deleted         BOOLEAN NOT NULL DEFAULT false,
    created_by      BIGINT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_by      BIGINT NOT NULL,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS invitations (
    id              BIGINT PRIMARY KEY,
    tenant_id       BIGINT NOT NULL,
    workspace_id    BIGINT NOT NULL,
    email           VARCHAR(255) NOT NULL,
    inviter_id      BIGINT NOT NULL,
    role            workspace_role_enum NOT NULL,
    token_hash      VARCHAR(255) NOT NULL,
    message         TEXT,
    expires_at      TIMESTAMPTZ NOT NULL,
    accepted_at     TIMESTAMPTZ,
    revoked_at      TIMESTAMPTZ,
    status          VARCHAR(20) NOT NULL DEFAULT 'pending',
    deleted         BOOLEAN NOT NULL DEFAULT false,
    created_by      BIGINT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_by      BIGINT NOT NULL,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (token_hash)
);

-- ===========================================================================
-- 第六部分: 项目域
-- ===========================================================================

CREATE TABLE IF NOT EXISTS projects (
    id              BIGINT PRIMARY KEY,
    code            VARCHAR(50),
    name            VARCHAR(255) NOT NULL,
    tenant_id       BIGINT NOT NULL,
    workspace_id    BIGINT NOT NULL,
    identifier      VARCHAR(20),
    slug            VARCHAR(100),
    description     TEXT,
    icon            VARCHAR(50),
    cover_image_url TEXT,
    network         VARCHAR(20) DEFAULT 'private' CHECK (network IN ('public', 'private', 'internal')),
    template        VARCHAR(20) DEFAULT 'generic' CHECK (template IN ('agile', 'waterfall', 'generic')),
    status          project_status NOT NULL DEFAULT 'active',
    modules         JSONB,
    start_date      DATE,
    target_date     DATE,
    owner_id        BIGINT NOT NULL,
    version         INT DEFAULT 1,
    sort_order      DOUBLE PRECISION DEFAULT 65535,
    deleted         BOOLEAN NOT NULL DEFAULT false,
    created_by      BIGINT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_by      BIGINT NOT NULL,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (workspace_id, identifier),
    UNIQUE (workspace_id, slug)
);

COMMENT ON TABLE projects IS '项目表';

CREATE TABLE IF NOT EXISTS project_members (
    id              BIGINT PRIMARY KEY,
    tenant_id       BIGINT NOT NULL,
    workspace_id    BIGINT NOT NULL,
    project_id      BIGINT NOT NULL,
    user_id         BIGINT NOT NULL,
    role            project_role_enum NOT NULL DEFAULT 'member',
    status          entity_status NOT NULL DEFAULT 'active',
    joined_at       TIMESTAMPTZ DEFAULT now(),
    deleted         BOOLEAN NOT NULL DEFAULT false,
    created_by      BIGINT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_by      BIGINT NOT NULL,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (workspace_id, project_id, user_id)
);

CREATE TABLE IF NOT EXISTS project_sequences (
    id              BIGINT PRIMARY KEY,
    tenant_id       BIGINT NOT NULL,
    workspace_id    BIGINT NOT NULL,
    project_id      BIGINT NOT NULL UNIQUE,
    sequence_id     BIGINT NOT NULL DEFAULT 0,
    status          entity_status NOT NULL DEFAULT 'active',
    deleted         BOOLEAN NOT NULL DEFAULT false,
    created_by      BIGINT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_by      BIGINT NOT NULL,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- ===========================================================================
-- 第七部分: 工作项域（核心）
-- ===========================================================================

CREATE TABLE IF NOT EXISTS task (
    id                  BIGINT PRIMARY KEY,
    code                VARCHAR(50),
    name                VARCHAR(255) NOT NULL,
    tenant_id           BIGINT NOT NULL,
    workspace_id        BIGINT NOT NULL,
    project_id          BIGINT NOT NULL,
    sequence_id         BIGINT NOT NULL,
    public_id           UUID NOT NULL DEFAULT gen_random_uuid(),
    parent_id           BIGINT,
    depth               SMALLINT NOT NULL DEFAULT 1 CHECK (depth BETWEEN 1 AND 3),
    description_json    JSONB,
    description_html    TEXT,
    description_stripped TEXT,
    state_id            BIGINT NOT NULL,
    priority            VARCHAR(20) NOT NULL DEFAULT 'none' CHECK (priority IN ('urgent', 'high', 'medium', 'low', 'none')),
    category            VARCHAR(20) CHECK (category IN ('frontend', 'backend', 'qa', 'doc', 'design', 'devops', 'other')),
    actual_effort       NUMERIC(8, 2),
    remaining_effort    NUMERIC(8, 2),
    delay_reason        VARCHAR(50) CHECK (delay_reason IN ('requirement_change', 'resource', 'blocked', 'other')),
    point               SMALLINT CHECK (point BETWEEN 0 AND 12),
    estimate_point_id   BIGINT,
    sprint_id           BIGINT,
    version_id          BIGINT,
    progress            SMALLINT NOT NULL DEFAULT 0 CHECK (progress BETWEEN 0 AND 100),
    start_date          DATE,
    target_date         DATE,
    completed_at        TIMESTAMPTZ,
    is_draft            BOOLEAN NOT NULL DEFAULT false,
    archived_at         TIMESTAMPTZ,
    sort_order          DOUBLE PRECISION NOT NULL DEFAULT 65535,
    version             INT NOT NULL DEFAULT 1,
    status              work_item_status NOT NULL DEFAULT 'active',
    deleted             BOOLEAN NOT NULL DEFAULT false,
    created_by          BIGINT NOT NULL,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_by          BIGINT NOT NULL,
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (project_id, sequence_id)
);

COMMENT ON TABLE task IS '任务工作项表';

CREATE TABLE IF NOT EXISTS requirement (
    id                  BIGINT PRIMARY KEY,
    code                VARCHAR(50),
    name                VARCHAR(255) NOT NULL,
    tenant_id           BIGINT NOT NULL,
    workspace_id        BIGINT NOT NULL,
    project_id          BIGINT NOT NULL,
    sequence_id         BIGINT NOT NULL,
    public_id           UUID NOT NULL DEFAULT gen_random_uuid(),
    parent_id           BIGINT,
    depth               SMALLINT NOT NULL DEFAULT 1 CHECK (depth BETWEEN 1 AND 3),
    description_json    JSONB,
    description_html    TEXT,
    description_stripped TEXT,
    state_id            BIGINT NOT NULL,
    priority            VARCHAR(20) NOT NULL DEFAULT 'none' CHECK (priority IN ('urgent', 'high', 'medium', 'low', 'none')),
    source              VARCHAR(50) CHECK (source IN ('customer', 'internal', 'competitor', 'other')),
    acceptance_criteria JSONB,
    business_value      TEXT,
    review_status       VARCHAR(20) CHECK (review_status IN ('draft', 'reviewing', 'accepted', 'rejected', 'verified')),
    point               SMALLINT CHECK (point BETWEEN 0 AND 12),
    estimate_point_id   BIGINT,
    sprint_id           BIGINT,
    version_id          BIGINT,
    progress            SMALLINT NOT NULL DEFAULT 0 CHECK (progress BETWEEN 0 AND 100),
    start_date          DATE,
    target_date         DATE,
    completed_at        TIMESTAMPTZ,
    is_draft            BOOLEAN NOT NULL DEFAULT false,
    archived_at         TIMESTAMPTZ,
    sort_order          DOUBLE PRECISION NOT NULL DEFAULT 65535,
    version             INT NOT NULL DEFAULT 1,
    status              work_item_status NOT NULL DEFAULT 'active',
    deleted             BOOLEAN NOT NULL DEFAULT false,
    created_by          BIGINT NOT NULL,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_by          BIGINT NOT NULL,
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (project_id, sequence_id)
);

COMMENT ON TABLE requirement IS '需求工作项表';

CREATE TABLE IF NOT EXISTS defect (
    id                  BIGINT PRIMARY KEY,
    code                VARCHAR(50),
    name                VARCHAR(255) NOT NULL,
    tenant_id           BIGINT NOT NULL,
    workspace_id        BIGINT NOT NULL,
    project_id          BIGINT NOT NULL,
    sequence_id         BIGINT NOT NULL,
    public_id           UUID NOT NULL DEFAULT gen_random_uuid(),
    parent_id           BIGINT,
    depth               SMALLINT NOT NULL DEFAULT 1 CHECK (depth BETWEEN 1 AND 3),
    description_json    JSONB,
    description_html    TEXT,
    description_stripped TEXT,
    state_id            BIGINT NOT NULL,
    priority            VARCHAR(20) NOT NULL DEFAULT 'none' CHECK (priority IN ('urgent', 'high', 'medium', 'low', 'none')),
    severity            SMALLINT NOT NULL CHECK (severity BETWEEN 1 AND 5),
    found_phase         VARCHAR(20) CHECK (found_phase IN ('unit', 'integration', 'uat', 'production', 'customer')),
    found_version_id    BIGINT,
    fix_version_id      BIGINT,
    root_cause_category VARCHAR(50) CHECK (root_cause_category IN ('requirement', 'technical', 'environment', 'data')),
    verifier_id         BIGINT,
    environment         JSONB,
    reproduce_steps     JSONB,
    fix_steps           JSONB,
    regression_risk     VARCHAR(20) CHECK (regression_risk IN ('low', 'medium', 'high')),
    point               SMALLINT CHECK (point BETWEEN 0 AND 12),
    estimate_point_id   BIGINT,
    sprint_id           BIGINT,
    version_id          BIGINT,
    progress            SMALLINT NOT NULL DEFAULT 0 CHECK (progress BETWEEN 0 AND 100),
    start_date          DATE,
    target_date         DATE,
    completed_at        TIMESTAMPTZ,
    is_draft            BOOLEAN NOT NULL DEFAULT false,
    archived_at         TIMESTAMPTZ,
    sort_order          DOUBLE PRECISION NOT NULL DEFAULT 65535,
    version             INT NOT NULL DEFAULT 1,
    status              work_item_status NOT NULL DEFAULT 'active',
    deleted             BOOLEAN NOT NULL DEFAULT false,
    created_by          BIGINT NOT NULL,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_by          BIGINT NOT NULL,
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (project_id, sequence_id)
);

COMMENT ON TABLE defect IS '缺陷工作项表';

-- 工作项扩展表
CREATE TABLE IF NOT EXISTS task_ext (
    id              BIGINT PRIMARY KEY,
    tenant_id       BIGINT NOT NULL,
    workspace_id    BIGINT NOT NULL,
    project_id      BIGINT NOT NULL,
    task_id         BIGINT NOT NULL,
    field_name      VARCHAR(100) NOT NULL,
    field_value     JSONB NOT NULL,
    field_schema    JSONB,
    status          entity_status NOT NULL DEFAULT 'active',
    deleted         BOOLEAN NOT NULL DEFAULT false,
    created_by      BIGINT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_by      BIGINT NOT NULL,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (task_id, field_name)
);

CREATE TABLE IF NOT EXISTS requirement_ext (
    id              BIGINT PRIMARY KEY,
    tenant_id       BIGINT NOT NULL,
    workspace_id    BIGINT NOT NULL,
    project_id      BIGINT NOT NULL,
    requirement_id  BIGINT NOT NULL,
    field_name      VARCHAR(100) NOT NULL,
    field_value     JSONB NOT NULL,
    field_schema    JSONB,
    status          entity_status NOT NULL DEFAULT 'active',
    deleted         BOOLEAN NOT NULL DEFAULT false,
    created_by      BIGINT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_by      BIGINT NOT NULL,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (requirement_id, field_name)
);

CREATE TABLE IF NOT EXISTS defect_ext (
    id              BIGINT PRIMARY KEY,
    tenant_id       BIGINT NOT NULL,
    workspace_id    BIGINT NOT NULL,
    project_id      BIGINT NOT NULL,
    defect_id       BIGINT NOT NULL,
    field_name      VARCHAR(100) NOT NULL,
    field_value     JSONB NOT NULL,
    field_schema    JSONB,
    status          entity_status NOT NULL DEFAULT 'active',
    deleted         BOOLEAN NOT NULL DEFAULT false,
    created_by      BIGINT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_by      BIGINT NOT NULL,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (defect_id, field_name)
);

-- ===========================================================================
-- 第八部分: 工作项关联表（分表设计）
-- ===========================================================================

-- Task 关联表
CREATE TABLE IF NOT EXISTS task_assignees (
    id              BIGINT PRIMARY KEY,
    tenant_id       BIGINT NOT NULL,
    workspace_id    BIGINT NOT NULL,
    project_id      BIGINT NOT NULL,
    task_id         BIGINT NOT NULL,
    user_id         BIGINT NOT NULL,
    status          entity_status NOT NULL DEFAULT 'active',
    deleted         BOOLEAN NOT NULL DEFAULT false,
    created_by      BIGINT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_by      BIGINT NOT NULL,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (task_id, user_id)
);

CREATE TABLE IF NOT EXISTS task_labels (
    id              BIGINT PRIMARY KEY,
    tenant_id       BIGINT NOT NULL,
    workspace_id    BIGINT NOT NULL,
    project_id      BIGINT NOT NULL,
    task_id         BIGINT NOT NULL,
    label_id        BIGINT NOT NULL,
    status          entity_status NOT NULL DEFAULT 'active',
    deleted         BOOLEAN NOT NULL DEFAULT false,
    created_by      BIGINT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_by      BIGINT NOT NULL,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (task_id, label_id)
);

CREATE TABLE IF NOT EXISTS task_modules (
    id              BIGINT PRIMARY KEY,
    tenant_id       BIGINT NOT NULL,
    workspace_id    BIGINT NOT NULL,
    project_id      BIGINT NOT NULL,
    task_id         BIGINT NOT NULL,
    module_id       BIGINT NOT NULL,
    status          entity_status NOT NULL DEFAULT 'active',
    deleted         BOOLEAN NOT NULL DEFAULT false,
    created_by      BIGINT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_by      BIGINT NOT NULL,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (task_id, module_id)
);

CREATE TABLE IF NOT EXISTS task_watchers (
    id              BIGINT PRIMARY KEY,
    tenant_id       BIGINT NOT NULL,
    workspace_id    BIGINT NOT NULL,
    project_id      BIGINT NOT NULL,
    task_id         BIGINT NOT NULL,
    user_id         BIGINT NOT NULL,
    status          entity_status NOT NULL DEFAULT 'active',
    deleted         BOOLEAN NOT NULL DEFAULT false,
    created_by      BIGINT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_by      BIGINT NOT NULL,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (task_id, user_id)
);

CREATE TABLE IF NOT EXISTS task_relations (
    id              BIGINT PRIMARY KEY,
    tenant_id       BIGINT NOT NULL,
    workspace_id    BIGINT NOT NULL,
    project_id      BIGINT NOT NULL,
    source_task_id  BIGINT NOT NULL,
    target_task_id  BIGINT NOT NULL,
    relation_type   VARCHAR(50) NOT NULL CHECK (relation_type IN ('duplicate', 'relates_to', 'blocked_by', 'start_before', 'finish_before')),
    status          entity_status NOT NULL DEFAULT 'active',
    deleted         BOOLEAN NOT NULL DEFAULT false,
    created_by      BIGINT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_by      BIGINT NOT NULL,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (source_task_id, target_task_id, relation_type)
);

CREATE TABLE IF NOT EXISTS task_comments (
    id              BIGINT PRIMARY KEY,
    tenant_id       BIGINT NOT NULL,
    workspace_id    BIGINT NOT NULL,
    project_id      BIGINT NOT NULL,
    task_id         BIGINT NOT NULL,
    content_json    JSONB NOT NULL,
    content_html    TEXT NOT NULL,
    parent_id       BIGINT,
    status          entity_status NOT NULL DEFAULT 'active',
    deleted         BOOLEAN NOT NULL DEFAULT false,
    created_by      BIGINT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_by      BIGINT,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- 需求/缺陷关联表（结构同 task_*）
CREATE TABLE IF NOT EXISTS requirement_assignees (
    id              BIGINT PRIMARY KEY,
    tenant_id       BIGINT NOT NULL,
    workspace_id    BIGINT NOT NULL,
    project_id      BIGINT NOT NULL,
    requirement_id  BIGINT NOT NULL,
    user_id         BIGINT NOT NULL,
    status          entity_status NOT NULL DEFAULT 'active',
    deleted         BOOLEAN NOT NULL DEFAULT false,
    created_by      BIGINT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_by      BIGINT NOT NULL,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (requirement_id, user_id)
);

CREATE TABLE IF NOT EXISTS requirement_labels (
    id              BIGINT PRIMARY KEY,
    tenant_id       BIGINT NOT NULL,
    workspace_id    BIGINT NOT NULL,
    project_id      BIGINT NOT NULL,
    requirement_id  BIGINT NOT NULL,
    label_id        BIGINT NOT NULL,
    status          entity_status NOT NULL DEFAULT 'active',
    deleted         BOOLEAN NOT NULL DEFAULT false,
    created_by      BIGINT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_by      BIGINT NOT NULL,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (requirement_id, label_id)
);

CREATE TABLE IF NOT EXISTS requirement_modules (
    id              BIGINT PRIMARY KEY,
    tenant_id       BIGINT NOT NULL,
    workspace_id    BIGINT NOT NULL,
    project_id      BIGINT NOT NULL,
    requirement_id  BIGINT NOT NULL,
    module_id       BIGINT NOT NULL,
    status          entity_status NOT NULL DEFAULT 'active',
    deleted         BOOLEAN NOT NULL DEFAULT false,
    created_by      BIGINT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_by      BIGINT NOT NULL,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (requirement_id, module_id)
);

CREATE TABLE IF NOT EXISTS requirement_watchers (
    id              BIGINT PRIMARY KEY,
    tenant_id       BIGINT NOT NULL,
    workspace_id    BIGINT NOT NULL,
    project_id      BIGINT NOT NULL,
    requirement_id  BIGINT NOT NULL,
    user_id         BIGINT NOT NULL,
    status          entity_status NOT NULL DEFAULT 'active',
    deleted         BOOLEAN NOT NULL DEFAULT false,
    created_by      BIGINT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_by      BIGINT NOT NULL,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (requirement_id, user_id)
);

CREATE TABLE IF NOT EXISTS requirement_relations (
    id                  BIGINT PRIMARY KEY,
    tenant_id           BIGINT NOT NULL,
    workspace_id        BIGINT NOT NULL,
    project_id          BIGINT NOT NULL,
    source_requirement_id BIGINT NOT NULL,
    target_requirement_id BIGINT NOT NULL,
    relation_type       VARCHAR(50) NOT NULL CHECK (relation_type IN ('duplicate', 'relates_to', 'blocked_by', 'start_before', 'finish_before')),
    status              entity_status NOT NULL DEFAULT 'active',
    deleted             BOOLEAN NOT NULL DEFAULT false,
    created_by          BIGINT NOT NULL,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_by          BIGINT NOT NULL,
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (source_requirement_id, target_requirement_id, relation_type)
);

CREATE TABLE IF NOT EXISTS requirement_comments (
    id              BIGINT PRIMARY KEY,
    tenant_id       BIGINT NOT NULL,
    workspace_id    BIGINT NOT NULL,
    project_id      BIGINT NOT NULL,
    requirement_id  BIGINT NOT NULL,
    content_json    JSONB NOT NULL,
    content_html    TEXT NOT NULL,
    parent_id       BIGINT,
    status          entity_status NOT NULL DEFAULT 'active',
    deleted         BOOLEAN NOT NULL DEFAULT false,
    created_by      BIGINT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_by      BIGINT,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS defect_assignees (
    id              BIGINT PRIMARY KEY,
    tenant_id       BIGINT NOT NULL,
    workspace_id    BIGINT NOT NULL,
    project_id      BIGINT NOT NULL,
    defect_id       BIGINT NOT NULL,
    user_id         BIGINT NOT NULL,
    status          entity_status NOT NULL DEFAULT 'active',
    deleted         BOOLEAN NOT NULL DEFAULT false,
    created_by      BIGINT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_by      BIGINT NOT NULL,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (defect_id, user_id)
);

CREATE TABLE IF NOT EXISTS defect_labels (
    id              BIGINT PRIMARY KEY,
    tenant_id       BIGINT NOT NULL,
    workspace_id    BIGINT NOT NULL,
    project_id      BIGINT NOT NULL,
    defect_id       BIGINT NOT NULL,
    label_id        BIGINT NOT NULL,
    status          entity_status NOT NULL DEFAULT 'active',
    deleted         BOOLEAN NOT NULL DEFAULT false,
    created_by      BIGINT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_by      BIGINT NOT NULL,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (defect_id, label_id)
);

CREATE TABLE IF NOT EXISTS defect_modules (
    id              BIGINT PRIMARY KEY,
    tenant_id       BIGINT NOT NULL,
    workspace_id    BIGINT NOT NULL,
    project_id      BIGINT NOT NULL,
    defect_id       BIGINT NOT NULL,
    module_id       BIGINT NOT NULL,
    status          entity_status NOT NULL DEFAULT 'active',
    deleted         BOOLEAN NOT NULL DEFAULT false,
    created_by      BIGINT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_by      BIGINT NOT NULL,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (defect_id, module_id)
);

CREATE TABLE IF NOT EXISTS defect_watchers (
    id              BIGINT PRIMARY KEY,
    tenant_id       BIGINT NOT NULL,
    workspace_id    BIGINT NOT NULL,
    project_id      BIGINT NOT NULL,
    defect_id       BIGINT NOT NULL,
    user_id         BIGINT NOT NULL,
    status          entity_status NOT NULL DEFAULT 'active',
    deleted         BOOLEAN NOT NULL DEFAULT false,
    created_by      BIGINT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_by      BIGINT NOT NULL,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (defect_id, user_id)
);

CREATE TABLE IF NOT EXISTS defect_relations (
    id                  BIGINT PRIMARY KEY,
    tenant_id           BIGINT NOT NULL,
    workspace_id        BIGINT NOT NULL,
    project_id          BIGINT NOT NULL,
    source_defect_id    BIGINT NOT NULL,
    target_defect_id    BIGINT NOT NULL,
    relation_type       VARCHAR(50) NOT NULL CHECK (relation_type IN ('duplicate', 'relates_to', 'blocked_by', 'start_before', 'finish_before')),
    status              entity_status NOT NULL DEFAULT 'active',
    deleted             BOOLEAN NOT NULL DEFAULT false,
    created_by          BIGINT NOT NULL,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_by          BIGINT NOT NULL,
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (source_defect_id, target_defect_id, relation_type)
);

CREATE TABLE IF NOT EXISTS defect_comments (
    id              BIGINT PRIMARY KEY,
    tenant_id       BIGINT NOT NULL,
    workspace_id    BIGINT NOT NULL,
    project_id      BIGINT NOT NULL,
    defect_id       BIGINT NOT NULL,
    content_json    JSONB NOT NULL,
    content_html    TEXT NOT NULL,
    parent_id       BIGINT,
    status          entity_status NOT NULL DEFAULT 'active',
    deleted         BOOLEAN NOT NULL DEFAULT false,
    created_by      BIGINT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_by      BIGINT,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- 跨类型关联表
CREATE TABLE IF NOT EXISTS biz_entity_relations (
    id              BIGINT PRIMARY KEY,
    tenant_id       BIGINT NOT NULL,
    workspace_id    BIGINT NOT NULL,
    project_id      BIGINT NOT NULL,
    source_type     VARCHAR(20) NOT NULL CHECK (source_type IN ('task', 'requirement', 'defect')),
    source_id       BIGINT NOT NULL,
    target_type     VARCHAR(20) NOT NULL CHECK (target_type IN ('task', 'requirement', 'defect')),
    target_id       BIGINT NOT NULL,
    relation_type   VARCHAR(50) NOT NULL CHECK (relation_type IN ('implemented_by', 'relates_to', 'duplicate', 'blocked_by', 'parent_child', 'found_in', 'fixed_in', 'verified_in')),
    status          entity_status NOT NULL DEFAULT 'active',
    deleted         BOOLEAN NOT NULL DEFAULT false,
    created_by      BIGINT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_by      BIGINT NOT NULL,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (source_type, source_id, target_type, target_id, relation_type)
);

-- ===========================================================================
-- 第九部分: 工时记录（分表）
-- ===========================================================================

CREATE TABLE IF NOT EXISTS task_timelogs (
    id              BIGINT PRIMARY KEY,
    tenant_id       BIGINT NOT NULL,
    workspace_id    BIGINT NOT NULL,
    project_id      BIGINT NOT NULL,
    task_id         BIGINT NOT NULL,
    user_id         BIGINT NOT NULL,
    spent_date      DATE NOT NULL DEFAULT CURRENT_DATE,
    duration_minutes INT NOT NULL CHECK (duration_minutes > 0 AND duration_minutes <= 1440),
    description     TEXT,
    status          entity_status NOT NULL DEFAULT 'active',
    deleted         BOOLEAN NOT NULL DEFAULT false,
    created_by      BIGINT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_by      BIGINT NOT NULL,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS requirement_timelogs (
    id              BIGINT PRIMARY KEY,
    tenant_id       BIGINT NOT NULL,
    workspace_id    BIGINT NOT NULL,
    project_id      BIGINT NOT NULL,
    requirement_id  BIGINT NOT NULL,
    user_id         BIGINT NOT NULL,
    spent_date      DATE NOT NULL DEFAULT CURRENT_DATE,
    duration_minutes INT NOT NULL CHECK (duration_minutes > 0 AND duration_minutes <= 1440),
    description     TEXT,
    status          entity_status NOT NULL DEFAULT 'active',
    deleted         BOOLEAN NOT NULL DEFAULT false,
    created_by      BIGINT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_by      BIGINT NOT NULL,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS defect_timelogs (
    id              BIGINT PRIMARY KEY,
    tenant_id       BIGINT NOT NULL,
    workspace_id    BIGINT NOT NULL,
    project_id      BIGINT NOT NULL,
    defect_id       BIGINT NOT NULL,
    user_id         BIGINT NOT NULL,
    spent_date      DATE NOT NULL DEFAULT CURRENT_DATE,
    duration_minutes INT NOT NULL CHECK (duration_minutes > 0 AND duration_minutes <= 1440),
    description     TEXT,
    status          entity_status NOT NULL DEFAULT 'active',
    deleted         BOOLEAN NOT NULL DEFAULT false,
    created_by      BIGINT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_by      BIGINT NOT NULL,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- ===========================================================================
-- 第十部分: 附件表（分表）
-- ===========================================================================

CREATE TABLE IF NOT EXISTS task_attachments (
    id              BIGINT PRIMARY KEY,
    tenant_id       BIGINT NOT NULL,
    workspace_id    BIGINT NOT NULL,
    project_id      BIGINT NOT NULL,
    task_id         BIGINT NOT NULL,
    file_name       VARCHAR(255) NOT NULL,
    file_size       BIGINT NOT NULL,
    file_type       VARCHAR(100) NOT NULL,
    storage_path    TEXT NOT NULL,
    thumbnail_path  TEXT,
    status          attachment_status NOT NULL DEFAULT 'available',
    deleted         BOOLEAN NOT NULL DEFAULT false,
    created_by      BIGINT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_by      BIGINT NOT NULL,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS requirement_attachments (
    id              BIGINT PRIMARY KEY,
    tenant_id       BIGINT NOT NULL,
    workspace_id    BIGINT NOT NULL,
    project_id      BIGINT NOT NULL,
    requirement_id  BIGINT NOT NULL,
    file_name       VARCHAR(255) NOT NULL,
    file_size       BIGINT NOT NULL,
    file_type       VARCHAR(100) NOT NULL,
    storage_path    TEXT NOT NULL,
    thumbnail_path  TEXT,
    status          attachment_status NOT NULL DEFAULT 'available',
    deleted         BOOLEAN NOT NULL DEFAULT false,
    created_by      BIGINT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_by      BIGINT NOT NULL,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS defect_attachments (
    id              BIGINT PRIMARY KEY,
    tenant_id       BIGINT NOT NULL,
    workspace_id    BIGINT NOT NULL,
    project_id      BIGINT NOT NULL,
    defect_id       BIGINT NOT NULL,
    file_name       VARCHAR(255) NOT NULL,
    file_size       BIGINT NOT NULL,
    file_type       VARCHAR(100) NOT NULL,
    storage_path    TEXT NOT NULL,
    thumbnail_path  TEXT,
    status          attachment_status NOT NULL DEFAULT 'available',
    deleted         BOOLEAN NOT NULL DEFAULT false,
    created_by      BIGINT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_by      BIGINT NOT NULL,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- ===========================================================================
-- 第十一部分: 迭代域
-- ===========================================================================

CREATE TABLE IF NOT EXISTS sprints (
    id              BIGINT PRIMARY KEY,
    code            VARCHAR(50),
    name            VARCHAR(255) NOT NULL,
    tenant_id       BIGINT NOT NULL,
    workspace_id    BIGINT NOT NULL,
    project_id      BIGINT NOT NULL,
    description     TEXT,
    goal            TEXT,
    start_date      DATE,
    end_date        DATE,
    capacity        NUMERIC(10, 2),
    owner_id        BIGINT,
    viewport        JSONB DEFAULT '{}',
    review_snapshot JSONB,
    started_at      TIMESTAMPTZ,
    completed_at    TIMESTAMPTZ,
    version_id      BIGINT,
    status          sprint_status NOT NULL DEFAULT 'planned',
    version         INT DEFAULT 1,
    deleted         BOOLEAN NOT NULL DEFAULT false,
    created_by      BIGINT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_by      BIGINT NOT NULL,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (start_date IS NULL OR end_date IS NULL OR start_date <= end_date)
);

CREATE TABLE IF NOT EXISTS sprint_snapshots (
    id              BIGINT PRIMARY KEY,
    tenant_id       BIGINT NOT NULL,
    workspace_id    BIGINT NOT NULL,
    project_id      BIGINT NOT NULL,
    sprint_id       BIGINT NOT NULL,
    snapshot_date   DATE NOT NULL,
    data            JSONB DEFAULT '{}',
    status          entity_status NOT NULL DEFAULT 'active',
    deleted         BOOLEAN NOT NULL DEFAULT false,
    created_by      BIGINT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_by      BIGINT NOT NULL,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (sprint_id, snapshot_date)
);

-- ===========================================================================
-- 第十二部分: 版本域
-- ===========================================================================

CREATE TABLE IF NOT EXISTS versions (
    id              BIGINT PRIMARY KEY,
    code            VARCHAR(50),
    name            VARCHAR(255) NOT NULL,
    tenant_id       BIGINT NOT NULL,
    workspace_id    BIGINT NOT NULL,
    project_id      BIGINT NOT NULL,
    semver          VARCHAR(50) NOT NULL,
    description     TEXT,
    checklist       JSONB DEFAULT '[]',
    release_notes   TEXT,
    start_date      DATE,
    end_date        DATE,
    target_date     DATE,
    delivered_at    TIMESTAMPTZ,
    archived_at     TIMESTAMPTZ,
    status          version_status NOT NULL DEFAULT 'planning',
    version         INT DEFAULT 1,
    deleted         BOOLEAN NOT NULL DEFAULT false,
    created_by      BIGINT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_by      BIGINT NOT NULL,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (start_date IS NULL OR end_date IS NULL OR start_date <= end_date),
    CHECK (semver ~ '^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(-[0-9A-Za-z\-.]+)?(\+[0-9A-Za-z\-.]+)?$'),
    UNIQUE (project_id, semver)
);

CREATE TABLE IF NOT EXISTS version_delivery_snapshots (
    id              BIGINT PRIMARY KEY,
    tenant_id       BIGINT NOT NULL,
    workspace_id    BIGINT NOT NULL,
    project_id      BIGINT NOT NULL,
    version_id      BIGINT NOT NULL,
    progress        JSONB DEFAULT '{}',
    quality         JSONB DEFAULT '{}',
    release_notes   TEXT,
    snapshot_at     TIMESTAMPTZ DEFAULT now(),
    status          entity_status NOT NULL DEFAULT 'active',
    deleted         BOOLEAN NOT NULL DEFAULT false,
    created_by      BIGINT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_by      BIGINT NOT NULL,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- ===========================================================================
-- 第十三部分: 状态域
-- ===========================================================================

CREATE TABLE IF NOT EXISTS states (
    id              BIGINT PRIMARY KEY,
    name            VARCHAR(255) NOT NULL,
    tenant_id       BIGINT NOT NULL,
    workspace_id    BIGINT NOT NULL,
    project_id      BIGINT NOT NULL,
    group           VARCHAR(50) NOT NULL CHECK (group IN ('backlog', 'unstarted', 'triage', 'started', 'completed', 'cancelled')),
    color           VARCHAR(20) DEFAULT '#8DA2C2',
    sequence        DOUBLE PRECISION DEFAULT 65535,
    is_default      BOOLEAN DEFAULT false,
    applicable_types TEXT[] DEFAULT '{all}',
    template_set    VARCHAR(50) DEFAULT 'custom' CHECK (template_set IN ('dev_flow', 'defect_flow', 'requirement_flow', 'custom')),
    status          entity_status NOT NULL DEFAULT 'active',
    deleted         BOOLEAN NOT NULL DEFAULT false,
    created_by      BIGINT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_by      BIGINT NOT NULL,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS state_transitions (
    id              BIGINT PRIMARY KEY,
    tenant_id       BIGINT NOT NULL,
    workspace_id    BIGINT NOT NULL,
    project_id      BIGINT NOT NULL,
    type_code       VARCHAR(20) DEFAULT 'all',
    from_state_id   BIGINT NOT NULL,
    to_state_id     BIGINT NOT NULL,
    required_fields JSONB DEFAULT '[]',
    status          entity_status NOT NULL DEFAULT 'active',
    deleted         BOOLEAN NOT NULL DEFAULT false,
    created_by      BIGINT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_by      BIGINT NOT NULL,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (project_id, type_code, from_state_id, to_state_id)
);

-- ===========================================================================
-- 第十四部分: 模块/标签/估算域
-- ===========================================================================

CREATE TABLE IF NOT EXISTS modules (
    id              BIGINT PRIMARY KEY,
    code            VARCHAR(50),
    name            VARCHAR(255) NOT NULL,
    tenant_id       BIGINT NOT NULL,
    workspace_id    BIGINT NOT NULL,
    project_id      BIGINT NOT NULL,
    public_id       UUID DEFAULT gen_random_uuid(),
    description     TEXT,
    lead_id         BIGINT,
    sort_order      FLOAT8 DEFAULT 65535,
    status          entity_status NOT NULL DEFAULT 'active',
    deleted         BOOLEAN NOT NULL DEFAULT false,
    created_by      BIGINT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_by      BIGINT NOT NULL,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, workspace_id, project_id, name)
);

CREATE TABLE IF NOT EXISTS labels (
    id              BIGINT PRIMARY KEY,
    code            VARCHAR(50),
    name            VARCHAR(255) NOT NULL,
    tenant_id       BIGINT NOT NULL,
    workspace_id    BIGINT NOT NULL,
    project_id      BIGINT NOT NULL,
    color           VARCHAR(20),
    description     TEXT,
    status          entity_status NOT NULL DEFAULT 'active',
    deleted         BOOLEAN NOT NULL DEFAULT false,
    created_by      BIGINT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_by      BIGINT NOT NULL,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, workspace_id, project_id, name)
);

CREATE TABLE IF NOT EXISTS estimate_points (
    id              BIGINT PRIMARY KEY,
    code            VARCHAR(50) NOT NULL,
    name            VARCHAR(255) NOT NULL,
    tenant_id       BIGINT NOT NULL,
    workspace_id    BIGINT NOT NULL,
    description     TEXT,
    points_config   JSONB NOT NULL,
    is_default      BOOLEAN NOT NULL DEFAULT false,
    status          entity_status NOT NULL DEFAULT 'active',
    deleted         BOOLEAN NOT NULL DEFAULT false,
    created_by      BIGINT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_by      BIGINT NOT NULL,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (workspace_id, code)
);

-- ===========================================================================
-- 第十五部分: 自动化域
-- ===========================================================================

CREATE TABLE IF NOT EXISTS automation_rules (
    id              BIGINT PRIMARY KEY,
    code            VARCHAR(50),
    name            VARCHAR(255) NOT NULL,
    tenant_id       BIGINT NOT NULL,
    workspace_id    BIGINT NOT NULL,
    project_id      BIGINT,
    description     TEXT,
    trigger_type    VARCHAR(50) NOT NULL,
    conditions      JSONB DEFAULT '{}',
    actions         JSONB DEFAULT '{}',
    sort_order      INT DEFAULT 0,
    status          entity_status NOT NULL DEFAULT 'active',
    deleted         BOOLEAN NOT NULL DEFAULT false,
    created_by      BIGINT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_by      BIGINT NOT NULL,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS rule_executions (
    id              BIGINT PRIMARY KEY,
    tenant_id       BIGINT NOT NULL,
    workspace_id    BIGINT NOT NULL,
    project_id      BIGINT,
    rule_id         BIGINT NOT NULL,
    trigger_event_id BIGINT,
    duration_ms     INT,
    error_message   TEXT,
    context_json    JSONB,
    trigger_depth   SMALLINT DEFAULT 0,
    via_automation  BOOLEAN DEFAULT false,
    status          work_item_status NOT NULL,
    deleted         BOOLEAN NOT NULL DEFAULT false,
    created_by      BIGINT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_by      BIGINT NOT NULL,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS automation_templates (
    id              BIGINT PRIMARY KEY,
    code            VARCHAR(50) NOT NULL UNIQUE,
    name            VARCHAR(255) NOT NULL,
    description     TEXT,
    category        VARCHAR(50),
    icon            VARCHAR(50),
    template_config JSONB NOT NULL,
    is_default      BOOLEAN DEFAULT false,
    sort_order      INT DEFAULT 0,
    status          entity_status NOT NULL DEFAULT 'active',
    deleted         BOOLEAN NOT NULL DEFAULT false,
    created_by      BIGINT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_by      BIGINT NOT NULL,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

COMMENT ON TABLE automation_templates IS '自动化模板表（系统级表，不含tenant_id）';

-- ===========================================================================
-- 第十六部分: 仪表盘域
-- ===========================================================================

CREATE TABLE IF NOT EXISTS dashboards (
    id              BIGINT PRIMARY KEY,
    code            VARCHAR(50),
    name            VARCHAR(255) NOT NULL,
    tenant_id       BIGINT NOT NULL,
    workspace_id    BIGINT NOT NULL,
    project_id      BIGINT,
    description     TEXT,
    layout          JSONB DEFAULT '{}',
    status          entity_status NOT NULL DEFAULT 'active',
    deleted         BOOLEAN NOT NULL DEFAULT false,
    created_by      BIGINT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_by      BIGINT NOT NULL,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS dashboard_widgets (
    id              BIGINT PRIMARY KEY,
    tenant_id       BIGINT NOT NULL,
    workspace_id    BIGINT NOT NULL,
    project_id      BIGINT,
    dashboard_id    BIGINT NOT NULL,
    widget_type     VARCHAR(50) NOT NULL,
    name            VARCHAR(255) NOT NULL,
    config          JSONB DEFAULT '{}',
    user_id         BIGINT,
    sort_order      INT DEFAULT 0,
    status          entity_status NOT NULL DEFAULT 'active',
    deleted         BOOLEAN NOT NULL DEFAULT false,
    created_by      BIGINT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_by      BIGINT NOT NULL,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (widget_type IN ('progress_overview', 'burndown', 'velocity', 'priority_split', 'state_distribution', 'overdue_list', 'blocked_list', 'risk_alert', 'recent_activity', 'team_workload', 'version_burndown', 'module_distribution', 'dora', 'project_compare'))
);

CREATE TABLE IF NOT EXISTS dashboard_snapshots (
    id              BIGINT PRIMARY KEY,
    tenant_id       BIGINT NOT NULL,
    workspace_id    BIGINT NOT NULL,
    project_id      BIGINT NOT NULL,
    dashboard_id    BIGINT NOT NULL,
    widget_type     VARCHAR(50) NOT NULL,
    refreshed_at    TIMESTAMPTZ DEFAULT now(),
    data            JSONB DEFAULT '{}',
    status          entity_status NOT NULL DEFAULT 'active',
    deleted         BOOLEAN NOT NULL DEFAULT false,
    created_by      BIGINT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_by      BIGINT NOT NULL,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (project_id, widget_type)
);

CREATE TABLE IF NOT EXISTS dashboard_templates (
    id              BIGINT PRIMARY KEY,
    code            VARCHAR(50) NOT NULL UNIQUE,
    name            VARCHAR(255) NOT NULL,
    description     TEXT,
    category        VARCHAR(50),
    layout          JSONB DEFAULT '{}',
    icon            VARCHAR(50),
    is_default      BOOLEAN DEFAULT false,
    sort_order      INT DEFAULT 0,
    status          entity_status NOT NULL DEFAULT 'active',
    deleted         BOOLEAN NOT NULL DEFAULT false,
    created_by      BIGINT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_by      BIGINT NOT NULL,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

COMMENT ON TABLE dashboard_templates IS '仪表盘模板表（系统级表，不含tenant_id）';

-- ===========================================================================
-- 第十七部分: 通知域
-- ===========================================================================

CREATE TABLE IF NOT EXISTS notifications (
    id              BIGINT PRIMARY KEY,
    tenant_id       BIGINT NOT NULL,
    workspace_id    BIGINT NOT NULL,
    recipient_id    BIGINT NOT NULL,
    actor_id        BIGINT,
    notification_type VARCHAR(50) NOT NULL,
    title           VARCHAR(255) NOT NULL,
    content         TEXT,
    entity_type     VARCHAR(50),
    entity_id       BIGINT,
    is_read         BOOLEAN DEFAULT false,
    is_archived     BOOLEAN DEFAULT false,
    read_at         TIMESTAMPTZ,
    status          notification_status NOT NULL DEFAULT 'pending',
    deleted         BOOLEAN NOT NULL DEFAULT false,
    created_by      BIGINT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_by      BIGINT NOT NULL,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS notification_deliveries (
    id              BIGINT PRIMARY KEY,
    tenant_id       BIGINT NOT NULL,
    workspace_id    BIGINT NOT NULL,
    notification_id BIGINT NOT NULL,
    channel         VARCHAR(20) NOT NULL,
    status          notification_status NOT NULL DEFAULT 'pending',
    retry_count     INT DEFAULT 0,
    next_retry_at   TIMESTAMPTZ,
    delivered_at    TIMESTAMPTZ,
    error_message   TEXT,
    deleted         BOOLEAN NOT NULL DEFAULT false,
    created_by      BIGINT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_by      BIGINT NOT NULL,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS notification_preferences (
    id              BIGINT PRIMARY KEY,
    tenant_id       BIGINT NOT NULL,
    workspace_id    BIGINT NOT NULL,
    user_id         BIGINT NOT NULL,
    channel_settings JSONB DEFAULT '{}',
    mute_all        BOOLEAN DEFAULT false,
    digest_enabled  BOOLEAN DEFAULT true,
    digest_schedule VARCHAR(20) DEFAULT 'daily',
    status          entity_status NOT NULL DEFAULT 'active',
    deleted         BOOLEAN NOT NULL DEFAULT false,
    created_by      BIGINT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_by      BIGINT NOT NULL,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, workspace_id)
);

CREATE TABLE IF NOT EXISTS notification_digests (
    id              BIGINT PRIMARY KEY,
    tenant_id       BIGINT NOT NULL,
    workspace_id    BIGINT NOT NULL,
    user_id         BIGINT NOT NULL,
    digest_type     VARCHAR(20) NOT NULL,
    content         JSONB DEFAULT '{}',
    scheduled_for   TIMESTAMPTZ,
    sent_at         TIMESTAMPTZ,
    status          notification_status NOT NULL DEFAULT 'pending',
    deleted         BOOLEAN NOT NULL DEFAULT false,
    created_by      BIGINT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_by      BIGINT NOT NULL,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- ===========================================================================
-- 第十八部分: 搜索域
-- ===========================================================================

CREATE TABLE IF NOT EXISTS search_documents (
    id              BIGINT PRIMARY KEY,
    tenant_id       BIGINT NOT NULL,
    workspace_id    BIGINT NOT NULL,
    project_id      BIGINT,
    doc_type        VARCHAR(50) NOT NULL CHECK (doc_type IN ('task', 'requirement', 'defect', 'sprint', 'version', 'page')),
    doc_id          BIGINT NOT NULL,
    title           VARCHAR(255),
    identifier      VARCHAR(100),
    content         TEXT,
    search_tsv      tsvector,
    metadata        JSONB DEFAULT '{}',
    status          entity_status NOT NULL DEFAULT 'active',
    deleted         BOOLEAN NOT NULL DEFAULT false,
    created_by      BIGINT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_by      BIGINT NOT NULL,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, workspace_id, doc_type, doc_id)
);

CREATE TABLE IF NOT EXISTS search_history (
    id              BIGINT PRIMARY KEY,
    tenant_id       BIGINT NOT NULL,
    workspace_id    BIGINT NOT NULL,
    user_id         BIGINT NOT NULL,
    query           TEXT NOT NULL,
    filters         JSONB DEFAULT '{}',
    result_count    INT DEFAULT 0,
    status          entity_status NOT NULL DEFAULT 'active',
    deleted         BOOLEAN NOT NULL DEFAULT false,
    created_by      BIGINT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_by      BIGINT NOT NULL,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS search_bookmarks (
    id              BIGINT PRIMARY KEY,
    tenant_id       BIGINT NOT NULL,
    workspace_id    BIGINT NOT NULL,
    project_id      BIGINT,
    user_id         BIGINT NOT NULL,
    name            VARCHAR(255) NOT NULL,
    query           TEXT,
    filters         JSONB DEFAULT '{}',
    is_shared       BOOLEAN DEFAULT false,
    sort_order      FLOAT8 DEFAULT 65535,
    status          entity_status NOT NULL DEFAULT 'active',
    deleted         BOOLEAN NOT NULL DEFAULT false,
    created_by      BIGINT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_by      BIGINT NOT NULL,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- ===========================================================================
-- 第十九部分: 风险与度量域
-- ===========================================================================

CREATE TABLE IF NOT EXISTS risk_rules (
    id              BIGINT PRIMARY KEY,
    code            VARCHAR(50),
    name            VARCHAR(255) NOT NULL,
    tenant_id       BIGINT NOT NULL,
    workspace_id    BIGINT NOT NULL,
    project_id      BIGINT,
    rule_type       VARCHAR(50) NOT NULL,
    condition_json  JSONB DEFAULT '{}',
    notify_channels TEXT[] DEFAULT '{}',
    is_active       BOOLEAN DEFAULT true,
    last_triggered  TIMESTAMPTZ,
    status          entity_status NOT NULL DEFAULT 'active',
    deleted         BOOLEAN NOT NULL DEFAULT false,
    created_by      BIGINT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_by      BIGINT NOT NULL,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (rule_type IN ('overdue_issue', 'overdue_sprint', 'blocked_count', 'sla_breach', 'stalled_progress', 'high_priority_open'))
);

CREATE TABLE IF NOT EXISTS risk_alerts (
    id              BIGINT PRIMARY KEY,
    tenant_id       BIGINT NOT NULL,
    workspace_id    BIGINT NOT NULL,
    project_id      BIGINT,
    rule_id         BIGINT NOT NULL,
    severity        VARCHAR(20) DEFAULT 'medium' CHECK (severity IN ('info', 'low', 'medium', 'high', 'critical')),
    title           VARCHAR(255) NOT NULL,
    description     TEXT,
    metadata        JSONB DEFAULT '{}',
    is_resolved     BOOLEAN DEFAULT false,
    resolved_at     TIMESTAMPTZ,
    resolved_by     BIGINT,
    status          entity_status NOT NULL DEFAULT 'active',
    deleted         BOOLEAN NOT NULL DEFAULT false,
    created_by      BIGINT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_by      BIGINT NOT NULL,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS metric_snapshots (
    id              BIGINT PRIMARY KEY,
    tenant_id       BIGINT NOT NULL,
    workspace_id    BIGINT NOT NULL,
    project_id      BIGINT,
    granularity     VARCHAR(20) NOT NULL CHECK (granularity IN ('daily', 'sprint', 'version')),
    ref_id          BIGINT,
    metric          VARCHAR(50) NOT NULL,
    snapshot_date   DATE NOT NULL,
    value           NUMERIC,
    metadata        JSONB DEFAULT '{}',
    status          entity_status NOT NULL DEFAULT 'active',
    deleted         BOOLEAN NOT NULL DEFAULT false,
    created_by      BIGINT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_by      BIGINT NOT NULL,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS metric_adjustments (
    id              BIGINT PRIMARY KEY,
    tenant_id       BIGINT NOT NULL,
    workspace_id    BIGINT NOT NULL,
    project_id      BIGINT,
    snapshot_id     BIGINT,
    metric          VARCHAR(50) NOT NULL,
    original_value  NUMERIC,
    adjusted_value  NUMERIC,
    reason          TEXT,
    adjusted_by     BIGINT,
    status          entity_status NOT NULL DEFAULT 'active',
    deleted         BOOLEAN NOT NULL DEFAULT false,
    created_by      BIGINT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_by      BIGINT NOT NULL,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- ===========================================================================
-- 第二十部分: 入口工单域
-- ===========================================================================

CREATE TABLE IF NOT EXISTS intake_channels (
    id              BIGINT PRIMARY KEY,
    code            VARCHAR(50),
    name            VARCHAR(255) NOT NULL,
    tenant_id       BIGINT NOT NULL,
    workspace_id    BIGINT NOT NULL,
    project_id      BIGINT,
    slug            VARCHAR(100) NOT NULL,
    description     TEXT,
    is_active       BOOLEAN DEFAULT true,
    config          JSONB DEFAULT '{}',
    status          entity_status NOT NULL DEFAULT 'active',
    deleted         BOOLEAN NOT NULL DEFAULT false,
    created_by      BIGINT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_by      BIGINT NOT NULL,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, workspace_id, slug)
);

CREATE TABLE IF NOT EXISTS intake_issues (
    id              BIGINT PRIMARY KEY,
    code            VARCHAR(50),
    name            VARCHAR(255) NOT NULL,
    tenant_id       BIGINT NOT NULL,
    workspace_id    BIGINT NOT NULL,
    project_id      BIGINT,
    channel_id      BIGINT NOT NULL,
    tracking_id     VARCHAR(50),
    submitter_name  VARCHAR(255),
    submitter_email VARCHAR(255) NOT NULL,
    description     TEXT,
    priority        VARCHAR(20) DEFAULT 'medium',
    status          intake_issue_status NOT NULL DEFAULT 'open',
    linked_entity_type VARCHAR(50),
    linked_entity_id BIGINT,
    resolved_at     TIMESTAMPTZ,
    resolved_by     BIGINT,
    deleted         BOOLEAN NOT NULL DEFAULT false,
    created_by      BIGINT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_by      BIGINT NOT NULL,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, workspace_id, tracking_id)
);

-- ===========================================================================
-- 第二十一部分: Webhooks
-- ===========================================================================

CREATE TABLE IF NOT EXISTS webhooks (
    id              BIGINT PRIMARY KEY,
    code            VARCHAR(50),
    name            VARCHAR(100) NOT NULL,
    tenant_id       BIGINT NOT NULL,
    workspace_id    BIGINT NOT NULL,
    project_id      BIGINT,
    target_url      TEXT NOT NULL CHECK (target_url ~ '^https?://'),
    secret          VARCHAR(255) NOT NULL,
    events          TEXT[] DEFAULT '{}',
    is_active       BOOLEAN DEFAULT true,
    last_error      TEXT,
    last_triggered  TIMESTAMPTZ,
    last_status     VARCHAR(20),
    unhealthy_at    TIMESTAMPTZ,
    status          entity_status NOT NULL DEFAULT 'active',
    deleted         BOOLEAN NOT NULL DEFAULT false,
    created_by      BIGINT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_by      BIGINT NOT NULL,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS webhook_logs (
    id              BIGINT PRIMARY KEY,
    tenant_id       BIGINT NOT NULL,
    workspace_id    BIGINT NOT NULL,
    webhook_id      BIGINT NOT NULL,
    delivery_id     VARCHAR(64) NOT NULL,
    event_type      VARCHAR(80) NOT NULL,
    event_id        BIGINT,
    request_url     TEXT NOT NULL,
    request_method  VARCHAR(10) DEFAULT 'POST',
    request_headers JSONB,
    request_body    TEXT,
    response_status INT,
    response_body   TEXT,
    response_headers JSONB,
    attempt         SMALLINT DEFAULT 1,
    duration_ms     INT,
    error           TEXT,
    status          webhook_delivery_status NOT NULL,
    deleted         BOOLEAN NOT NULL DEFAULT false,
    occurred_at     TIMESTAMPTZ DEFAULT now(),
    created_by      BIGINT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_by      BIGINT NOT NULL,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- ===========================================================================
-- 第二十二部分: 工作台与视图
-- ===========================================================================

CREATE TABLE IF NOT EXISTS workbench_configs (
    id              BIGINT PRIMARY KEY,
    tenant_id       BIGINT NOT NULL,
    workspace_id    BIGINT NOT NULL,
    project_id      BIGINT,
    user_id         BIGINT NOT NULL,
    layout          JSONB DEFAULT '{}',
    widget_states   JSONB DEFAULT '{}',
    focus_enabled   BOOLEAN DEFAULT false,
    status          entity_status NOT NULL DEFAULT 'active',
    deleted         BOOLEAN NOT NULL DEFAULT false,
    created_by      BIGINT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_by      BIGINT NOT NULL,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS view_preferences (
    id              BIGINT PRIMARY KEY,
    tenant_id       BIGINT NOT NULL,
    workspace_id    BIGINT NOT NULL,
    project_id      BIGINT,
    user_id         BIGINT NOT NULL,
    view_type       VARCHAR(20) NOT NULL,
    layout          VARCHAR(20) DEFAULT 'list',
    columns         JSONB DEFAULT '[]',
    filters         JSONB DEFAULT '{}',
    sort            JSONB DEFAULT '{}',
    extra           JSONB DEFAULT '{}',
    status          entity_status NOT NULL DEFAULT 'active',
    deleted         BOOLEAN NOT NULL DEFAULT false,
    created_by      BIGINT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_by      BIGINT NOT NULL,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (workspace_id, project_id, user_id, view_type)
);

CREATE TABLE IF NOT EXISTS recent_items (
    id              BIGINT PRIMARY KEY,
    tenant_id       BIGINT NOT NULL,
    workspace_id    BIGINT NOT NULL,
    project_id      BIGINT,
    user_id         BIGINT NOT NULL,
    item_type       VARCHAR(20) NOT NULL CHECK (item_type IN ('project', 'task', 'requirement', 'defect', 'sprint', 'version', 'page')),
    item_id         BIGINT NOT NULL,
    title           VARCHAR(255),
    identifier      VARCHAR(100),
    accessed_at     TIMESTAMPTZ DEFAULT now(),
    status          entity_status NOT NULL DEFAULT 'active',
    deleted         BOOLEAN NOT NULL DEFAULT false,
    created_by      BIGINT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_by      BIGINT NOT NULL,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, item_type, item_id)
);

-- ===========================================================================
-- 第二十三部分: 知识库与文档
-- ===========================================================================

CREATE TABLE IF NOT EXISTS knowledge_spaces (
    id              BIGINT PRIMARY KEY,
    code            VARCHAR(50),
    name            VARCHAR(255) NOT NULL,
    tenant_id       BIGINT NOT NULL,
    workspace_id    BIGINT NOT NULL,
    description     TEXT,
    icon            VARCHAR(50),
    status          entity_status NOT NULL DEFAULT 'active',
    deleted         BOOLEAN NOT NULL DEFAULT false,
    created_by      BIGINT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_by      BIGINT NOT NULL,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS knowledge_pages (
    id              BIGINT PRIMARY KEY,
    code            VARCHAR(50),
    name            VARCHAR(255) NOT NULL,
    tenant_id       BIGINT NOT NULL,
    workspace_id    BIGINT NOT NULL,
    knowledge_space_id BIGINT NOT NULL,
    public_id       UUID DEFAULT gen_random_uuid(),
    parent_id       BIGINT,
    depth           SMALLINT DEFAULT 1 CHECK (depth BETWEEN 1 AND 3),
    content_json    JSONB,
    content_html    TEXT,
    sort_order      DOUBLE PRECISION DEFAULT 65535,
    version         INT DEFAULT 1,
    status          entity_status NOT NULL DEFAULT 'active',
    deleted         BOOLEAN NOT NULL DEFAULT false,
    created_by      BIGINT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_by      BIGINT NOT NULL,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS pages (
    id              BIGINT PRIMARY KEY,
    public_id       UUID NOT NULL DEFAULT gen_random_uuid(),
    tenant_id       BIGINT NOT NULL,
    workspace_id    BIGINT NOT NULL,
    project_id      BIGINT NOT NULL,
    name            VARCHAR(255) NOT NULL,
    description_json JSONB,
    description_html TEXT,
    description_stripped TEXT,
    parent_id       BIGINT,
    sort_order      DOUBLE PRECISION NOT NULL DEFAULT 65535,
    version         INT NOT NULL DEFAULT 1,
    status          entity_status NOT NULL DEFAULT 'active',
    deleted         BOOLEAN NOT NULL DEFAULT false,
    created_by      BIGINT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_by      BIGINT NOT NULL,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- ===========================================================================
-- 第二十四部分: SSO / 部署事件 / 其他
-- ===========================================================================

CREATE TABLE IF NOT EXISTS sso_providers (
    id              BIGINT PRIMARY KEY,
    name            VARCHAR(255) NOT NULL,
    protocol        VARCHAR(20) DEFAULT 'oidc' CHECK (protocol IN ('oidc', 'saml')),
    issuer_url      TEXT,
    client_id       VARCHAR(255) NOT NULL,
    client_secret   TEXT NOT NULL,
    redirect_uri    TEXT NOT NULL,
    auth_url        TEXT,
    token_url       TEXT,
    userinfo_url    TEXT,
    jwks_url        TEXT,
    sso_url         TEXT,
    idp_issuer      VARCHAR(255),
    idp_certificate TEXT,
    skip_signature  BOOLEAN DEFAULT false,
    scopes          TEXT DEFAULT 'openid email profile',
    auto_create_user BOOLEAN DEFAULT true,
    default_role    VARCHAR(20) DEFAULT 'member',
    attribute_mapping JSONB DEFAULT '{}',
    enabled         BOOLEAN DEFAULT true,
    status          entity_status NOT NULL DEFAULT 'active',
    deleted         BOOLEAN NOT NULL DEFAULT false,
    created_by      BIGINT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_by      BIGINT NOT NULL,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

COMMENT ON TABLE sso_providers IS 'SSO提供方配置表（系统级表，不含tenant_id）';

CREATE TABLE IF NOT EXISTS deployment_events (
    id              BIGINT PRIMARY KEY,
    tenant_id       BIGINT NOT NULL,
    workspace_id    BIGINT NOT NULL,
    project_id      BIGINT,
    deployment_id   VARCHAR(64),
    env             VARCHAR(20) NOT NULL CHECK (env IN ('development', 'staging', 'production', 'testing')),
    status          VARCHAR(20) NOT NULL CHECK (status IN ('success', 'failed', 'rolled_back')),
    version         VARCHAR(100),
    deployed_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted         BOOLEAN NOT NULL DEFAULT false,
    created_by      BIGINT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_by      BIGINT NOT NULL,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (deployment_id, env, project_id)
);

CREATE TABLE IF NOT EXISTS api_tokens (
    id              BIGINT PRIMARY KEY,
    tenant_id       BIGINT NOT NULL,
    user_id         BIGINT NOT NULL,
    name            VARCHAR(255) NOT NULL,
    token_hash      VARCHAR(255) NOT NULL,
    scopes          TEXT[] DEFAULT '{}',
    expires_at      TIMESTAMPTZ,
    last_used_at    TIMESTAMPTZ,
    revoked_at      TIMESTAMPTZ,
    status          entity_status NOT NULL DEFAULT 'active',
    deleted         BOOLEAN NOT NULL DEFAULT false,
    created_by      BIGINT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_by      BIGINT NOT NULL,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (token_hash)
);

CREATE TABLE IF NOT EXISTS audit_logs (
    id              BIGINT PRIMARY KEY,
    tenant_id       BIGINT NOT NULL,
    workspace_id    BIGINT,
    actor_id        BIGINT,
    action          VARCHAR(100) NOT NULL,
    target_type     VARCHAR(50),
    target_id       BIGINT,
    details         JSONB DEFAULT '{}',
    ip_address      INET,
    user_agent      TEXT,
    status          entity_status NOT NULL DEFAULT 'active',
    deleted         BOOLEAN NOT NULL DEFAULT false,
    created_by      BIGINT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_by      BIGINT NOT NULL,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS domain_events (
    id              BIGINT PRIMARY KEY,
    tenant_id       BIGINT NOT NULL,
    workspace_id    BIGINT,
    event_type      VARCHAR(100) NOT NULL,
    aggregate_type  VARCHAR(50),
    aggregate_id    BIGINT,
    payload         JSONB NOT NULL,
    metadata        JSONB DEFAULT '{}',
    published_at    TIMESTAMPTZ,
    status          entity_status NOT NULL DEFAULT 'active',
    deleted         BOOLEAN NOT NULL DEFAULT false,
    created_by      BIGINT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_by      BIGINT NOT NULL,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS processed_events (
    event_id        BIGINT NOT NULL,
    consumer_id     VARCHAR(100) NOT NULL,
    processed_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    retry_count     INT NOT NULL DEFAULT 1,
    PRIMARY KEY (event_id, consumer_id)
);

CREATE TABLE IF NOT EXISTS dlq_events (
    id              BIGINT PRIMARY KEY,
    event_id        BIGINT,
    tenant_id       BIGINT,
    workspace_id    BIGINT,
    queue           VARCHAR(100) NOT NULL,
    exchange        VARCHAR(100) NOT NULL,
    routing_key     VARCHAR(200) DEFAULT '',
    payload         JSONB,
    error_reason    TEXT,
    resolved_at     TIMESTAMPTZ,
    resolved_by     VARCHAR(100),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS password_reset_tokens (
    id              BIGINT PRIMARY KEY,
    user_id         BIGINT NOT NULL,
    token_hash      VARCHAR(255) NOT NULL,
    expires_at      TIMESTAMPTZ NOT NULL,
    used_at         TIMESTAMPTZ,
    status          entity_status NOT NULL DEFAULT 'active',
    deleted         BOOLEAN NOT NULL DEFAULT false,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (token_hash)
);

CREATE TABLE IF NOT EXISTS idempotency_keys (
    key             VARCHAR(255) PRIMARY KEY,
    user_id         BIGINT,
    response_status INT,
    response_body   JSONB,
    expires_at      TIMESTAMPTZ NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS schema_migrations (
    version         BIGINT PRIMARY KEY,
    dirty           BOOLEAN NOT NULL DEFAULT false
);

-- ===========================================================================
-- 第二十五部分: 核心索引（关键查询路径）
-- ===========================================================================

-- 租户域索引
CREATE INDEX idx_workspaces_tenant ON workspaces (tenant_id) WHERE NOT deleted;
CREATE INDEX idx_workspace_members_tenant_ws ON workspace_members (tenant_id, workspace_id) WHERE NOT deleted;
CREATE INDEX idx_workspace_members_user ON workspace_members (user_id) WHERE NOT deleted;
CREATE INDEX idx_users_tenant_email ON users (tenant_id, email) WHERE NOT deleted;
CREATE INDEX idx_users_tenant_phone ON users (tenant_id, phone) WHERE phone IS NOT NULL AND NOT deleted;

-- 项目域索引
CREATE INDEX idx_projects_tenant_ws ON projects (tenant_id, workspace_id) WHERE NOT deleted;
CREATE INDEX idx_project_members_project ON project_members (tenant_id, workspace_id, project_id) WHERE NOT deleted;
CREATE INDEX idx_project_members_user ON project_members (user_id) WHERE NOT deleted;

-- 工作项索引
CREATE INDEX idx_task_project ON task (tenant_id, workspace_id, project_id) WHERE NOT deleted;
CREATE INDEX idx_task_parent ON task (parent_id) WHERE NOT deleted AND parent_id IS NOT NULL;
CREATE INDEX idx_task_sprint ON task (tenant_id, project_id, sprint_id) WHERE NOT deleted;
CREATE INDEX idx_task_target_date ON task (tenant_id, project_id, target_date) WHERE NOT deleted AND completed_at IS NULL;
CREATE INDEX idx_task_sort ON task (project_id, state_id, sort_order);

CREATE INDEX idx_requirement_project ON requirement (tenant_id, workspace_id, project_id) WHERE NOT deleted;
CREATE INDEX idx_requirement_parent ON requirement (parent_id) WHERE NOT deleted AND parent_id IS NOT NULL;
CREATE INDEX idx_requirement_sprint ON requirement (tenant_id, project_id, sprint_id) WHERE NOT deleted;

CREATE INDEX idx_defect_project ON defect (tenant_id, workspace_id, project_id) WHERE NOT deleted;
CREATE INDEX idx_defect_parent ON defect (parent_id) WHERE NOT deleted AND parent_id IS NOT NULL;
CREATE INDEX idx_defect_sprint ON defect (tenant_id, project_id, sprint_id) WHERE NOT deleted;
CREATE INDEX idx_defect_severity ON defect (tenant_id, project_id, severity) WHERE NOT deleted AND completed_at IS NULL;

-- 迭代版本索引
CREATE INDEX idx_sprints_project_status ON sprints (tenant_id, project_id, status) WHERE NOT deleted;
CREATE UNIQUE INDEX idx_sprints_active ON sprints (tenant_id, project_id) WHERE status = 'active' AND NOT deleted;
CREATE INDEX idx_versions_project ON versions (tenant_id, project_id) WHERE NOT deleted;

-- 状态域索引
CREATE INDEX idx_states_project ON states (tenant_id, project_id) WHERE NOT deleted;
CREATE INDEX idx_state_transitions_lookup ON state_transitions (tenant_id, project_id, type_code, from_state_id) WHERE NOT deleted;

-- 搜索索引
CREATE INDEX idx_search_documents_tsv ON search_documents USING gin(search_tsv);
CREATE INDEX idx_search_documents_type ON search_documents (tenant_id, workspace_id, doc_type);

-- 通知索引
CREATE INDEX idx_notifications_recipient ON notifications (tenant_id, recipient_id, is_read, created_at DESC) WHERE is_archived = false;

-- 其他索引
CREATE INDEX idx_biz_entity_relations_source ON biz_entity_relations (tenant_id, source_type, source_id);
CREATE INDEX idx_biz_entity_relations_target ON biz_entity_relations (tenant_id, target_type, target_id);

-- ===========================================================================
-- 第二十六部分: updated_at 触发器
-- ===========================================================================

DO $$
DECLARE
    tbl TEXT;
    tables TEXT[] := ARRAY[
        'tenants', 'workspaces', 'workspace_members', 'users', 'roles', 'menus',
        'user_roles', 'invitations', 'projects', 'project_members',
        'task', 'requirement', 'defect', 'task_ext', 'requirement_ext', 'defect_ext',
        'sprints', 'versions', 'states', 'modules', 'labels', 'estimate_points',
        'automation_rules', 'automation_templates', 'dashboards', 'dashboard_widgets',
        'dashboard_snapshots', 'dashboard_templates', 'notifications',
        'notification_deliveries', 'notification_preferences', 'notification_digests',
        'search_documents', 'search_history', 'search_bookmarks',
        'risk_rules', 'risk_alerts', 'metric_snapshots', 'metric_adjustments',
        'intake_channels', 'intake_issues', 'webhooks', 'webhook_logs',
        'workbench_configs', 'view_preferences', 'recent_items',
        'knowledge_spaces', 'knowledge_pages', 'pages', 'sso_providers',
        'deployment_events', 'api_tokens', 'audit_logs', 'domain_events',
        'workbench_configs', 'task_assignees', 'task_labels', 'task_modules',
        'task_watchers', 'task_relations', 'task_comments',
        'requirement_assignees', 'requirement_labels', 'requirement_modules',
        'requirement_watchers', 'requirement_relations', 'requirement_comments',
        'defect_assignees', 'defect_labels', 'defect_modules',
        'defect_watchers', 'defect_relations', 'defect_comments',
        'biz_entity_relations', 'task_timelogs', 'requirement_timelogs', 'defect_timelogs',
        'task_attachments', 'requirement_attachments', 'defect_attachments'
    ];
BEGIN
    FOREACH tbl IN ARRAY tables
    LOOP
        IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = tbl) THEN
            IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'trg_' || tbl || '_updated_at' AND tgrelid = ('public.' || tbl)::regclass) THEN
                EXECUTE format('CREATE TRIGGER trg_%s_updated_at BEFORE UPDATE ON %I FOR EACH ROW EXECUTE FUNCTION set_updated_at()', tbl, tbl);
            END IF;
        END IF;
    END LOOP;
END $$;

-- ===========================================================================
-- 第二十七部分: 种子数据（最小化）
-- ===========================================================================

-- 默认工作空间模板（供新用户初始化）
-- 注意: 完整的应用种子数据在 seeds/ 目录下独立维护

INSERT INTO schema_migrations (version, dirty) VALUES (1, false) ON CONFLICT DO NOTHING;

-- ===========================================================================
-- V2.0 初始化完毕
-- 所有表统一遵循:
--   1. ENUM 状态类型 (tenant_status/user_status/entity_status/project_status/sprint_status/version_status/...)
--   2. 三层次隔离: tenant_id → workspace_id → project_id
--   3. 雪花ID (BIGINT PRIMARY KEY)
--   4. 字段排序: id/code/name ... status/deleted/tenant_id/created_by/created_at/updated_by/updated_at
-- ===========================================================================
