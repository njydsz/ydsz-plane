-- 工作项分表拆分：将issues表拆分为task/defect/requirement三张独立表
-- 遵循项目多租户约束：所有新表含workspace_id NOT NULL，首列索引，开启RLS
-- 遵循项目主键约束：id BIGINT自增主键，对外暴露public_id UUID
-- 遵循项目时间字段约束：created_at/updated_at为TIMESTAMPTZ，updated_at由触发器维护
-- 遵循项目乐观锁约束：version INT NOT NULL DEFAULT 1，支持CAS冲突检测
-- 遵循项目软删除约束：deleted_at TIMESTAMPTZ NULL，唯一索引带WHERE deleted_at IS NULL

-- 1. 创建task表（任务专用）
CREATE TABLE IF NOT EXISTS task (
    id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    public_id       UUID NOT NULL DEFAULT gen_random_uuid(),
    workspace_id    BIGINT NOT NULL REFERENCES workspaces(id),
    project_id      BIGINT NOT NULL REFERENCES projects(id),
    sequence_id     BIGINT NOT NULL,
    parent_id       BIGINT REFERENCES task(id),
    depth           SMALLINT NOT NULL DEFAULT 1 CHECK (depth BETWEEN 1 AND 3),
    name            TEXT NOT NULL,
    description_json    JSONB,
    description_html    TEXT,
    description_stripped TEXT,
    state_id        BIGINT NOT NULL REFERENCES states(id),
    priority        TEXT NOT NULL DEFAULT 'none'
                    CHECK (priority IN ('urgent','high','medium','low','none')),
    -- task专属字段
    category        TEXT CHECK (category IN ('frontend','backend','qa','doc','design','devops','other')),
    actual_effort   NUMERIC(8,2),
    remaining_effort NUMERIC(8,2),
    delay_reason    TEXT CHECK (delay_reason IN ('requirement_change','resource','blocked','other')),
    -- 公共字段
    point           SMALLINT CHECK (point BETWEEN 0 AND 12),
    estimate_point_id BIGINT REFERENCES estimate_points(id),
    sprint_id       BIGINT REFERENCES sprints(id),
    version_id      BIGINT REFERENCES versions(id),
    progress        SMALLINT NOT NULL DEFAULT 0 CHECK (progress BETWEEN 0 AND 100),
    start_date      DATE,
    target_date     DATE,
    completed_at    TIMESTAMPTZ,
    is_draft        BOOLEAN NOT NULL DEFAULT FALSE,
    sort_order      DOUBLE PRECISION NOT NULL DEFAULT 65535,
    created_by      BIGINT NOT NULL REFERENCES users(id),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,
    version         INT NOT NULL DEFAULT 1,
    UNIQUE (project_id, sequence_id),
    UNIQUE (public_id) WHERE deleted_at IS NULL
);

-- task表RLS配置（对齐项目现有租户隔离策略）
ALTER TABLE task ENABLE ROW LEVEL SECURITY;
ALTER TABLE task FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON task
    USING (workspace_id = current_setting('app.workspace_id', true)::bigint)
    WITH CHECK (workspace_id = current_setting('app.workspace_id', true)::bigint);

-- task表索引（对齐项目现有索引规范）
CREATE INDEX idx_task_project_state   ON task(workspace_id, project_id, state_id)  WHERE deleted_at IS NULL;
CREATE INDEX idx_task_project_sprint  ON task(workspace_id, project_id, sprint_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_task_parent          ON task(parent_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_task_target_date     ON task(workspace_id, project_id, target_date) WHERE deleted_at IS NULL AND completed_at IS NULL;
CREATE INDEX idx_task_updated         ON task(workspace_id, updated_at DESC);
CREATE INDEX idx_task_fts             ON task USING gin(to_tsvector('simple', coalesce(name,'') || ' ' || coalesce(description_stripped,'')));
CREATE INDEX idx_task_sort            ON task(project_id, state_id, sort_order);
CREATE INDEX idx_task_assignee       ON task USING gin(assignee_ids) WHERE deleted_at IS NULL;


-- 2. 创建requirement表（需求专用）
CREATE TABLE IF NOT EXISTS requirement (
    id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    public_id       UUID NOT NULL DEFAULT gen_random_uuid(),
    workspace_id    BIGINT NOT NULL REFERENCES workspaces(id),
    project_id      BIGINT NOT NULL REFERENCES projects(id),
    sequence_id     BIGINT NOT NULL,
    parent_id       BIGINT REFERENCES requirement(id),
    depth           SMALLINT NOT NULL DEFAULT 1 CHECK (depth BETWEEN 1 AND 3),
    name            TEXT NOT NULL,
    description_json    JSONB,
    description_html    TEXT,
    description_stripped TEXT,
    state_id        BIGINT NOT NULL REFERENCES states(id),
    priority        TEXT NOT NULL DEFAULT 'none'
                    CHECK (priority IN ('urgent','high','medium','low','none')),
    -- requirement专属字段
    source          TEXT CHECK (source IN ('customer','internal','competitor','other')),
    acceptance_criteria JSONB,
    business_value  TEXT,
    review_status   TEXT CHECK (review_status IN ('draft','reviewing','accepted','rejected','verified')),
    -- 公共字段
    point           SMALLINT CHECK (point BETWEEN 0 AND 12),
    estimate_point_id BIGINT REFERENCES estimate_points(id),
    sprint_id       BIGINT REFERENCES sprints(id),
    version_id      BIGINT REFERENCES versions(id),
    progress        SMALLINT NOT NULL DEFAULT 0 CHECK (progress BETWEEN 0 AND 100),
    start_date      DATE,
    target_date     DATE,
    completed_at    TIMESTAMPTZ,
    is_draft        BOOLEAN NOT NULL DEFAULT FALSE,
    sort_order      DOUBLE PRECISION NOT NULL DEFAULT 65535,
    created_by      BIGINT NOT NULL REFERENCES users(id),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,
    version         INT NOT NULL DEFAULT 1,
    UNIQUE (project_id, sequence_id),
    UNIQUE (public_id) WHERE deleted_at IS NULL
);

-- requirement表RLS配置
ALTER TABLE requirement ENABLE ROW LEVEL SECURITY;
ALTER TABLE requirement FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON requirement
    USING (workspace_id = current_setting('app.workspace_id', true)::bigint)
    WITH CHECK (workspace_id = current_setting('app.workspace_id', true)::bigint);

-- requirement表索引
CREATE INDEX idx_requirement_project_state   ON requirement(workspace_id, project_id, state_id)  WHERE deleted_at IS NULL;
CREATE INDEX idx_requirement_project_sprint  ON requirement(workspace_id, project_id, sprint_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_requirement_parent          ON requirement(parent_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_requirement_target_date     ON requirement(workspace_id, project_id, target_date) WHERE deleted_at IS NULL AND completed_at IS NULL;
CREATE INDEX idx_requirement_updated         ON requirement(workspace_id, updated_at DESC);
CREATE INDEX idx_requirement_fts             ON requirement USING gin(to_tsvector('simple', coalesce(name,'') || ' ' || coalesce(description_stripped,'')));
CREATE INDEX idx_requirement_sort            ON requirement(project_id, state_id, sort_order);


-- 3. 创建defect表（缺陷专用）
CREATE TABLE IF NOT EXISTS defect (
    id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    public_id       UUID NOT NULL DEFAULT gen_random_uuid(),
    workspace_id    BIGINT NOT NULL REFERENCES workspaces(id),
    project_id      BIGINT NOT NULL REFERENCES projects(id),
    sequence_id     BIGINT NOT NULL,
    parent_id       BIGINT REFERENCES defect(id),
    depth           SMALLINT NOT NULL DEFAULT 1 CHECK (depth BETWEEN 1 AND 3),
    name            TEXT NOT NULL,
    description_json    JSONB,
    description_html    TEXT,
    description_stripped TEXT,
    state_id        BIGINT NOT NULL REFERENCES states(id),
    priority        TEXT NOT NULL DEFAULT 'none'
                    CHECK (priority IN ('urgent','high','medium','low','none')),
    -- defect专属字段
    severity        SMALLINT NOT NULL CHECK (severity BETWEEN 1 AND 5),
    found_phase     TEXT NOT NULL CHECK (found_phase IN ('unit','integration','uat','production','customer')),
    found_version_id BIGINT REFERENCES versions(id),
    fix_version_id   BIGINT REFERENCES versions(id),
    root_cause_category TEXT CHECK (root_cause_category IN ('requirement','technical','environment','data')),
    verifier_id     BIGINT REFERENCES users(id),
    environment     JSONB,
    reproduce_steps JSONB NOT NULL,
    fix_steps       JSONB,
    regression_risk TEXT CHECK (regression_risk IN ('low','medium','high')),
    -- 公共字段
    point           SMALLINT CHECK (point BETWEEN 0 AND 12),
    estimate_point_id BIGINT REFERENCES estimate_points(id),
    sprint_id       BIGINT REFERENCES sprints(id),
    version_id      BIGINT REFERENCES versions(id),
    progress        SMALLINT NOT NULL DEFAULT 0 CHECK (progress BETWEEN 0 AND 100),
    start_date      DATE,
    target_date     DATE,
    completed_at    TIMESTAMPTZ,
    is_draft        BOOLEAN NOT NULL DEFAULT FALSE,
    sort_order      DOUBLE PRECISION NOT NULL DEFAULT 65535,
    created_by      BIGINT NOT NULL REFERENCES users(id),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,
    version         INT NOT NULL DEFAULT 1,
    UNIQUE (project_id, sequence_id),
    UNIQUE (public_id) WHERE deleted_at IS NULL
);

-- defect表RLS配置
ALTER TABLE defect ENABLE ROW LEVEL SECURITY;
ALTER TABLE defect FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON defect
    USING (workspace_id = current_setting('app.workspace_id', true)::bigint)
    WITH CHECK (workspace_id = current_setting('app.workspace_id', true)::bigint);

-- defect表索引
CREATE INDEX idx_defect_project_state   ON defect(workspace_id, project_id, state_id)  WHERE deleted_at IS NULL;
CREATE INDEX idx_defect_project_sprint  ON defect(workspace_id, project_id, sprint_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_defect_parent          ON defect(parent_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_defect_target_date     ON defect(workspace_id, project_id, target_date) WHERE deleted_at IS NULL AND completed_at IS NULL;
CREATE INDEX idx_defect_updated         ON defect(workspace_id, updated_at DESC);
CREATE INDEX idx_defect_fts             ON defect USING gin(to_tsvector('simple', coalesce(name,'') || ' ' || coalesce(description_stripped,'')));
CREATE INDEX idx_defect_sort            ON defect(project_id, state_id, sort_order);
CREATE INDEX idx_defect_severity        ON defect(workspace_id, project_id, severity) WHERE deleted_at IS NULL AND completed_at IS NULL;
CREATE INDEX idx_defect_root_cause      ON defect(workspace_id, project_id, root_cause_category) WHERE deleted_at IS NULL;


-- 4. 创建扩展属性表（支持三类工作项自定义字段）
CREATE TABLE IF NOT EXISTS task_ext (
    id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    workspace_id    BIGINT NOT NULL REFERENCES workspaces(id),
    project_id      BIGINT NOT NULL REFERENCES projects(id),
    task_id         BIGINT NOT NULL REFERENCES task(id) ON DELETE CASCADE,
    field_name      TEXT NOT NULL,
    field_value     JSONB NOT NULL,
    field_schema    JSONB NOT NULL,
    created_by      BIGINT NOT NULL REFERENCES users(id),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (task_id, field_name)
);
ALTER TABLE task_ext ENABLE ROW LEVEL SECURITY;
ALTER TABLE task_ext FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON task_ext
    USING (workspace_id = current_setting('app.workspace_id', true)::bigint)
    WITH CHECK (workspace_id = current_setting('app.workspace_id', true)::bigint);

CREATE TABLE IF NOT EXISTS requirement_ext (
    id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    workspace_id    BIGINT NOT NULL REFERENCES workspaces(id),
    project_id      BIGINT NOT NULL REFERENCES projects(id),
    requirement_id  BIGINT NOT NULL REFERENCES requirement(id) ON DELETE CASCADE,
    field_name      TEXT NOT NULL,
    field_value     JSONB NOT NULL,
    field_schema    JSONB NOT NULL,
    created_by      BIGINT NOT NULL REFERENCES users(id),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (requirement_id, field_name)
);
ALTER TABLE requirement_ext ENABLE ROW LEVEL SECURITY;
ALTER TABLE requirement_ext FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON requirement_ext
    USING (workspace_id = current_setting('app.workspace_id', true)::bigint)
    WITH CHECK (workspace_id = current_setting('app.workspace_id', true)::bigint);

CREATE TABLE IF NOT EXISTS defect_ext (
    id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    workspace_id    BIGINT NOT NULL REFERENCES workspaces(id),
    project_id      BIGINT NOT NULL REFERENCES projects(id),
    defect_id       BIGINT NOT NULL REFERENCES defect(id) ON DELETE CASCADE,
    field_name      TEXT NOT NULL,
    field_value     JSONB NOT NULL,
    field_schema    JSONB NOT NULL,
    created_by      BIGINT NOT NULL REFERENCES users(id),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (defect_id, field_name)
);
ALTER TABLE defect_ext ENABLE ROW LEVEL SECURITY;
ALTER TABLE defect_ext FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON defect_ext
    USING (workspace_id = current_setting('app.workspace_id', true)::bigint)
    WITH CHECK (workspace_id = current_setting('app.workspace_id', true)::bigint);


-- 5. 创建通用工作项关联关系表（替代原单表内的关联逻辑，支持三类工作项互相关联）
CREATE TABLE IF NOT EXISTS biz_entity_relation (
    id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    workspace_id    BIGINT NOT NULL REFERENCES workspaces(id),
    project_id      BIGINT NOT NULL REFERENCES projects(id),
    source_type     TEXT NOT NULL CHECK (source_type IN ('task','requirement','defect')),
    source_id       BIGINT NOT NULL,
    target_type     TEXT NOT NULL CHECK (target_type IN ('task','requirement','defect')),
    target_id       BIGINT NOT NULL,
    relation_type   TEXT NOT NULL CHECK (relation_type IN ('implemented_by','relates_to','duplicate','blocked_by','parent_child','found_in','fixed_in','verified_in')),
    created_by      BIGINT NOT NULL REFERENCES users(id),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (source_type, source_id, target_type, target_id, relation_type)
);
ALTER TABLE biz_entity_relation ENABLE ROW LEVEL SECURITY;
ALTER TABLE biz_entity_relation FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON biz_entity_relation
    USING (workspace_id = current_setting('app.workspace_id', true)::bigint)
    WITH CHECK (workspace_id = current_setting('app.workspace_id', true)::bigint);
CREATE INDEX idx_biz_entity_relation_source ON biz_entity_relation(workspace_id, source_type, source_id) ;
CREATE INDEX idx_biz_entity_relation_target ON biz_entity_relation(workspace_id, target_type, target_id) ;


-- 6. 适配states表，支持按工作项类型区分
ALTER TABLE states ADD COLUMN IF NOT EXISTS applicable_types TEXT[] NOT NULL DEFAULT '{"all"}' ;
CREATE INDEX idx_states_applicable_types ON states USING gin(applicable_types) ;


-- 7. 适配state_transitions表，明确适用类型
ALTER TABLE state_transitions ALTER COLUMN type_code DROP DEFAULT ;
UPDATE state_transitions SET type_code = 'all' WHERE type_code = '' ;
CREATE INDEX idx_state_transitions_type ON state_transitions(project_id, type_code) ;
