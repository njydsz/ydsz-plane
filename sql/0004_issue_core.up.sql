-- 0004_issue_core: issues + states + modules + labels + M2M + activities + time_logs (Sprint 3 — M2 数据层)
-- 参考 docs/architecture/04-数据模型设计.md + 07-工作项与状态机设计.md

-- -----------------------------------------------------------------
-- states: 工作项状态（项目维度，从模板复制）
-- group: backlog / started / completed / cancelled
-- -----------------------------------------------------------------
CREATE TABLE states (
    id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    workspace_id    BIGINT NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    project_id      BIGINT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name            TEXT NOT NULL,
    "group"           TEXT NOT NULL CHECK ("group" IN ('backlog','started','completed','cancelled')),
    color           TEXT NOT NULL DEFAULT '#8DA2C2',
    sequence        DOUBLE PRECISION NOT NULL DEFAULT 65535,
    is_default      BOOLEAN NOT NULL DEFAULT FALSE,     -- 行创建工作项时的默认状态
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ
);
CREATE INDEX idx_states_project ON states(project_id, sequence) WHERE deleted_at IS NULL;

ALTER TABLE states ENABLE ROW LEVEL SECURITY;
ALTER TABLE states FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON states
    USING (workspace_id = current_setting('app.workspace_id', true)::bigint)
    WITH CHECK (workspace_id = current_setting('app.workspace_id', true)::bigint);

-- -----------------------------------------------------------------
-- modules: 项目模块/组件（如前端/后端/部署）
-- -----------------------------------------------------------------
CREATE TABLE modules (
    id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    workspace_id    BIGINT NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    project_id      BIGINT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name            TEXT NOT NULL,
    description     TEXT,
    lead_id         BIGINT REFERENCES users(id),
    status          TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active','completed','cancelled')),
    start_date      DATE,
    target_date     DATE,
    sort_order      DOUBLE PRECISION NOT NULL DEFAULT 65535,
    created_by      BIGINT NOT NULL REFERENCES users(id),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ
);
CREATE INDEX idx_modules_project ON modules(project_id) WHERE deleted_at IS NULL;

ALTER TABLE modules ENABLE ROW LEVEL SECURITY;
ALTER TABLE modules FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON modules
    USING (workspace_id = current_setting('app.workspace_id', true)::bigint)
    WITH CHECK (workspace_id = current_setting('app.workspace_id', true)::bigint);

-- -----------------------------------------------------------------
-- labels: 项目标签
-- -----------------------------------------------------------------
CREATE TABLE labels (
    id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    workspace_id    BIGINT NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    project_id      BIGINT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name            TEXT NOT NULL,
    color           TEXT NOT NULL DEFAULT '#8DA2C2',
    description     TEXT,
    created_by      BIGINT NOT NULL REFERENCES users(id),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ
);
CREATE INDEX idx_labels_project ON labels(project_id) WHERE deleted_at IS NULL;

ALTER TABLE labels ENABLE ROW LEVEL SECURITY;
ALTER TABLE labels FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON labels
    USING (workspace_id = current_setting('app.workspace_id', true)::bigint)
    WITH CHECK (workspace_id = current_setting('app.workspace_id', true)::bigint);

-- -----------------------------------------------------------------
-- issues: 系统最核心表（统一需求/任务/缺陷）
-- -----------------------------------------------------------------
CREATE TABLE issues (
    id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    public_id       UUID NOT NULL DEFAULT gen_random_uuid(),
    workspace_id    BIGINT NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    project_id      BIGINT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    sequence_id     BIGINT NOT NULL,                     -- 项目内自增，配合 identifier 展示 YD-123
    type_code       TEXT NOT NULL CHECK (type_code IN ('requirement','task','defect')),
    parent_id       BIGINT REFERENCES issues(id),
    depth           SMALLINT NOT NULL DEFAULT 1 CHECK (depth BETWEEN 1 AND 3),
    name            TEXT NOT NULL,
    description_json    JSONB,
    description_html    TEXT,
    description_stripped TEXT,                           -- tsvector 检索源
    state_id        BIGINT NOT NULL REFERENCES states(id),
    priority        TEXT NOT NULL DEFAULT 'none'
                    CHECK (priority IN ('urgent','high','medium','low','none')),
    -- defect 专有
    severity        SMALLINT CHECK (severity BETWEEN 1 AND 5),
    found_phase     TEXT CHECK (found_phase IN ('unit','integration','uat','production','customer')),
    root_cause_category TEXT CHECK (root_cause_category IN ('requirement','technical','environment','data')),
    verifier_id     BIGINT REFERENCES users(id),
    environment     JSONB,
    reproduce_steps JSONB,                               -- {steps, expected, actual}
    -- task 专有
    category        TEXT,                                -- frontend/backend/qa/doc
    actual_effort   NUMERIC(8,2),
    remaining_effort NUMERIC(8,2),
    delay_reason    TEXT CHECK (delay_reason IN ('requirement_change','resource','blocked','other')),
    -- requirement 专有
    source          TEXT,                                -- customer/internal/competitor
    -- 通用
    point           SMALLINT CHECK (point BETWEEN 0 AND 12),
    sprint_id       BIGINT,                              -- 预留，Sprint 5 起可用
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
    CONSTRAINT defect_required CHECK (
        type_code <> 'defect' OR (severity IS NOT NULL AND found_phase IS NOT NULL)
    )
);

CREATE UNIQUE INDEX idx_issues_project_sequence ON issues(project_id, sequence_id) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX idx_issues_public_id ON issues(public_id) WHERE deleted_at IS NULL;

-- 索引
CREATE INDEX idx_issues_workspace_project   ON issues(workspace_id, project_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_issues_project_state       ON issues(project_id, state_id, sort_order) WHERE deleted_at IS NULL;
CREATE INDEX idx_issues_parent              ON issues(parent_id) WHERE deleted_at IS NULL AND parent_id IS NOT NULL;
CREATE INDEX idx_issues_target_date         ON issues(project_id, target_date) WHERE deleted_at IS NULL AND completed_at IS NULL;
CREATE INDEX idx_issues_type                ON issues(project_id, type_code) WHERE deleted_at IS NULL;
CREATE INDEX idx_issues_updated             ON issues(project_id, updated_at DESC);
CREATE INDEX idx_issues_created             ON issues(project_id, created_at DESC);

ALTER TABLE issues ENABLE ROW LEVEL SECURITY;
ALTER TABLE issues FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON issues
    USING (workspace_id = current_setting('app.workspace_id', true)::bigint)
    WITH CHECK (workspace_id = current_setting('app.workspace_id', true)::bigint);

-- -----------------------------------------------------------------
-- project_sequences: 工作项编号发号器
-- -----------------------------------------------------------------
CREATE TABLE project_sequences (
    project_id BIGINT PRIMARY KEY REFERENCES projects(id) ON DELETE CASCADE,
    next_value BIGINT NOT NULL DEFAULT 1
);

-- -----------------------------------------------------------------
-- state_transitions: 状态流转规则（按项目×类型维度）
-- -----------------------------------------------------------------
CREATE TABLE state_transitions (
    id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    workspace_id    BIGINT NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    project_id      BIGINT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    type_code       TEXT NOT NULL DEFAULT 'all',                 -- all | requirement | task | defect
    from_state_id   BIGINT NOT NULL REFERENCES states(id),
    to_state_id     BIGINT NOT NULL REFERENCES states(id),
    required_fields JSONB NOT NULL DEFAULT '[]',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (project_id, type_code, from_state_id, to_state_id)
);
CREATE INDEX idx_state_transitions_lookup ON state_transitions(project_id, type_code, from_state_id);

ALTER TABLE state_transitions ENABLE ROW LEVEL SECURITY;
ALTER TABLE state_transitions FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON state_transitions
    USING (workspace_id = current_setting('app.workspace_id', true)::bigint)
    WITH CHECK (workspace_id = current_setting('app.workspace_id', true)::bigint);

-- -----------------------------------------------------------------
-- M2M: issue_assignees
-- -----------------------------------------------------------------
CREATE TABLE issue_assignees (
    issue_id    BIGINT NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
    user_id     BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    assigned_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    assigned_by BIGINT REFERENCES users(id),
    PRIMARY KEY (issue_id, user_id)
);
CREATE INDEX idx_issue_assignees_user ON issue_assignees(user_id);

-- -----------------------------------------------------------------
-- M2M: issue_labels
-- -----------------------------------------------------------------
CREATE TABLE issue_labels (
    issue_id    BIGINT NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
    label_id    BIGINT NOT NULL REFERENCES labels(id) ON DELETE CASCADE,
    PRIMARY KEY (issue_id, label_id)
);

-- -----------------------------------------------------------------
-- M2M: issue_modules
-- -----------------------------------------------------------------
CREATE TABLE issue_modules (
    issue_id    BIGINT NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
    module_id   BIGINT NOT NULL REFERENCES modules(id) ON DELETE CASCADE,
    PRIMARY KEY (issue_id, module_id)
);

-- -----------------------------------------------------------------
-- M2M: issue_watchers
-- -----------------------------------------------------------------
CREATE TABLE issue_watchers (
    issue_id    BIGINT NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
    user_id     BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (issue_id, user_id)
);
CREATE INDEX idx_issue_watchers_user ON issue_watchers(user_id);

-- -----------------------------------------------------------------
-- issue_activities: 字段级变更日志
-- -----------------------------------------------------------------
CREATE TABLE issue_activities (
    id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    workspace_id    BIGINT NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    project_id      BIGINT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    issue_id        BIGINT NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
    verb            TEXT NOT NULL CHECK (verb IN ('created','updated','transitioned','attached','linked','unlinked','commented')),
    field           TEXT,                                -- 变更字段名
    old_value       TEXT,
    new_value       TEXT,
    old_ref         JSONB,                               -- 复杂引用
    new_ref         JSONB,
    actor_id        BIGINT REFERENCES users(id),
    actor_email     TEXT,
    actor_name      TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_activities_issue ON issue_activities(issue_id, created_at DESC);
CREATE INDEX idx_activities_project ON issue_activities(project_id, created_at DESC);

ALTER TABLE issue_activities ENABLE ROW LEVEL SECURITY;
ALTER TABLE issue_activities FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON issue_activities
    USING (workspace_id = current_setting('app.workspace_id', true)::bigint)
    WITH CHECK (workspace_id = current_setting('app.workspace_id', true)::bigint);

-- -----------------------------------------------------------------
-- time_logs: 工时记录
-- -----------------------------------------------------------------
CREATE TABLE time_logs (
    id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    workspace_id    BIGINT NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    project_id      BIGINT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    issue_id        BIGINT NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
    user_id         BIGINT NOT NULL REFERENCES users(id),
    spent_date      DATE NOT NULL DEFAULT current_date,
    duration_minutes INTEGER NOT NULL CHECK (duration_minutes > 0 AND duration_minutes <= 1440),
    description     TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ
);
CREATE INDEX idx_time_logs_issue ON time_logs(issue_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_time_logs_user_date ON time_logs(user_id, spent_date) WHERE deleted_at IS NULL;

ALTER TABLE time_logs ENABLE ROW LEVEL SECURITY;
ALTER TABLE time_logs FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON time_logs
    USING (workspace_id = current_setting('app.workspace_id', true)::bigint)
    WITH CHECK (workspace_id = current_setting('app.workspace_id', true)::bigint);

-- -----------------------------------------------------------------
-- issue_relations: 6 种关系
-- type: duplicate | relates_to | blocked_by | start_before | finish_before | implemented_by
-- -----------------------------------------------------------------
CREATE TABLE issue_relations (
    id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    workspace_id    BIGINT NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    project_id      BIGINT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    source_issue_id BIGINT NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
    target_issue_id BIGINT NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
    relation_type   TEXT NOT NULL CHECK (relation_type IN ('duplicate','relates_to','blocked_by','start_before','finish_before','implemented_by')),
    created_by      BIGINT NOT NULL REFERENCES users(id),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (source_issue_id, target_issue_id, relation_type),
    CONSTRAINT no_self_relation CHECK (source_issue_id <> target_issue_id)
);
CREATE INDEX idx_issue_relations_source ON issue_relations(source_issue_id);
CREATE INDEX idx_issue_relations_target ON issue_relations(target_issue_id);

ALTER TABLE issue_relations ENABLE ROW LEVEL SECURITY;
ALTER TABLE issue_relations FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON issue_relations
    USING (workspace_id = current_setting('app.workspace_id', true)::bigint)
    WITH CHECK (workspace_id = current_setting('app.workspace_id', true)::bigint);

-- -----------------------------------------------------------------
-- issue_dependencies: FS/SS/FF/SF + lag_days
-- kind: finish_to_start(FS) | start_to_start(SS) | finish_to_finish(FF) | start_to_finish(SF)
-- -----------------------------------------------------------------
CREATE TABLE issue_dependencies (
    id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    workspace_id    BIGINT NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    project_id      BIGINT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    predecessor_id  BIGINT NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
    successor_id    BIGINT NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
    dependency_type TEXT NOT NULL CHECK (dependency_type IN ('FS','SS','FF','SF')),
    lag_days        INTEGER NOT NULL DEFAULT 0,
    created_by      BIGINT NOT NULL REFERENCES users(id),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (predecessor_id, successor_id, dependency_type),
    CONSTRAINT no_self_dependency CHECK (predecessor_id <> successor_id)
);
CREATE INDEX idx_issue_deps_pred ON issue_dependencies(predecessor_id);
CREATE INDEX idx_issue_deps_succ ON issue_dependencies(successor_id);

ALTER TABLE issue_dependencies ENABLE ROW LEVEL SECURITY;
ALTER TABLE issue_dependencies FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON issue_dependencies
    USING (workspace_id = current_setting('app.workspace_id', true)::bigint)
    WITH CHECK (workspace_id = current_setting('app.workspace_id', true)::bigint);

-- -----------------------------------------------------------------
-- updated_at 触发器
-- -----------------------------------------------------------------
CREATE TRIGGER trg_states_updated_at BEFORE UPDATE ON states
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE TRIGGER trg_modules_updated_at BEFORE UPDATE ON modules
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE TRIGGER trg_labels_updated_at BEFORE UPDATE ON labels
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE TRIGGER trg_issues_updated_at BEFORE UPDATE ON issues
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE TRIGGER trg_time_logs_updated_at BEFORE UPDATE ON time_logs
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
