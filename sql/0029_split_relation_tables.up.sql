-- 拆分原来的通用关联表为三个工作项独立的关联表，对齐三个独立主表的设计
-- 所有关联表包含workspace_id字段，开启RLS租户隔离，结构和原通用表对齐

-- ==================== 任务工作项关联表 ====================
-- 1. 任务指派人关联表
CREATE TABLE IF NOT EXISTS task_assignees (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    workspace_id BIGINT NOT NULL REFERENCES workspaces(id),
    task_id BIGINT NOT NULL REFERENCES task(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (task_id, user_id)
);
ALTER TABLE task_assignees ENABLE ROW LEVEL SECURITY;
ALTER TABLE task_assignees FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON task_assignees
    USING (workspace_id = current_setting('app.workspace_id', true)::bigint)
    WITH CHECK (workspace_id = current_setting('app.workspace_id', true)::bigint);
CREATE INDEX idx_task_assignee_task ON task_assignees(workspace_id, task_id) ;
CREATE INDEX idx_task_assignee_user ON task_assignees(workspace_id, user_id) ;

-- 2. 任务标签关联表
CREATE TABLE IF NOT EXISTS task_labels (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    workspace_id BIGINT NOT NULL REFERENCES workspaces(id),
    task_id BIGINT NOT NULL REFERENCES task(id) ON DELETE CASCADE,
    label_id BIGINT NOT NULL REFERENCES labels(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (task_id, label_id)
);
ALTER TABLE task_labels ENABLE ROW LEVEL SECURITY;
ALTER TABLE task_labels FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON task_labels
    USING (workspace_id = current_setting('app.workspace_id', true)::bigint)
    WITH CHECK (workspace_id = current_setting('app.workspace_id', true)::bigint);
CREATE INDEX idx_task_label_task ON task_labels(workspace_id, task_id) ;
CREATE INDEX idx_task_label_label ON task_labels(workspace_id, label_id) ;

-- 3. 任务模块关联表
CREATE TABLE IF NOT EXISTS task_modules (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    workspace_id BIGINT NOT NULL REFERENCES workspaces(id),
    task_id BIGINT NOT NULL REFERENCES task(id) ON DELETE CASCADE,
    module_id BIGINT NOT NULL REFERENCES modules(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (task_id, module_id)
);
ALTER TABLE task_modules ENABLE ROW LEVEL SECURITY;
ALTER TABLE task_modules FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON task_modules
    USING (workspace_id = current_setting('app.workspace_id', true)::bigint)
    WITH CHECK (workspace_id = current_setting('app.workspace_id', true)::bigint);
CREATE INDEX idx_task_module_task ON task_modules(workspace_id, task_id) ;
CREATE INDEX idx_task_module_module ON task_modules(workspace_id, module_id) ;

-- 4. 任务关注人关联表
CREATE TABLE IF NOT EXISTS task_watchers (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    workspace_id BIGINT NOT NULL REFERENCES workspaces(id),
    task_id BIGINT NOT NULL REFERENCES task(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (task_id, user_id)
);
ALTER TABLE task_watchers ENABLE ROW LEVEL SECURITY;
ALTER TABLE task_watchers FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON task_watchers
    USING (workspace_id = current_setting('app.workspace_id', true)::bigint)
    WITH CHECK (workspace_id = current_setting('app.workspace_id', true)::bigint);
CREATE INDEX idx_task_watcher_task ON task_watchers(workspace_id, task_id) ;
CREATE INDEX idx_task_watcher_user ON task_watchers(workspace_id, user_id) ;

-- 5. 任务关联关系表
CREATE TABLE IF NOT EXISTS task_relations (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    workspace_id BIGINT NOT NULL REFERENCES workspaces(id),
    source_task_id BIGINT NOT NULL REFERENCES task(id) ON DELETE CASCADE,
    target_task_id BIGINT NOT NULL REFERENCES task(id) ON DELETE CASCADE,
    relation_type TEXT NOT NULL CHECK (relation_type IN ('duplicate','relates_to','blocked_by','start_before','finish_before')),
    created_by BIGINT NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (source_task_id, target_task_id, relation_type)
);
ALTER TABLE task_relations ENABLE ROW LEVEL SECURITY;
ALTER TABLE task_relations FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON task_relations
    USING (workspace_id = current_setting('app.workspace_id', true)::bigint)
    WITH CHECK (workspace_id = current_setting('app.workspace_id', true)::bigint);
CREATE INDEX idx_task_relation_source ON task_relations(workspace_id, source_task_id) ;
CREATE INDEX idx_task_relation_target ON task_relations(workspace_id, target_task_id) ;

-- 6. 任务评论表
CREATE TABLE IF NOT EXISTS task_comments (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    workspace_id BIGINT NOT NULL REFERENCES workspaces(id),
    task_id BIGINT NOT NULL REFERENCES task(id) ON DELETE CASCADE,
    content_json JSONB NOT NULL,
    content_html TEXT NOT NULL,
    parent_id BIGINT REFERENCES task_comments(id) ON DELETE CASCADE,
    created_by BIGINT NOT NULL REFERENCES users(id),
    updated_by BIGINT REFERENCES users(id),
    deleted_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
ALTER TABLE task_comments ENABLE ROW LEVEL SECURITY;
ALTER TABLE task_comments FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON task_comments
    USING (workspace_id = current_setting('app.workspace_id', true)::bigint)
    WITH CHECK (workspace_id = current_setting('app.workspace_id', true)::bigint);
CREATE INDEX idx_task_comment_task ON task_comments(workspace_id, task_id) WHERE deleted_at IS NULL;

-- 7. 任务活动记录表（分区表，按月分区，原issue_activities拆分）
CREATE TABLE IF NOT EXISTS task_activities (
    id BIGINT GENERATED ALWAYS AS IDENTITY,
    workspace_id BIGINT NOT NULL REFERENCES workspaces(id),
    task_id BIGINT NOT NULL REFERENCES task(id) ON DELETE CASCADE,
    verb TEXT NOT NULL CHECK (verb IN ('created','updated','transitioned','attached','linked','commented')),
    field_name TEXT,
    old_value TEXT,
    new_value TEXT,
    actor_id BIGINT NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (id, created_at)
) PARTITION BY RANGE (created_at);
ALTER TABLE task_activities ENABLE ROW LEVEL SECURITY;
ALTER TABLE task_activities FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON task_activities
    USING (workspace_id = current_setting('app.workspace_id', true)::bigint)
    WITH CHECK (workspace_id = current_setting('app.workspace_id', true)::bigint);
CREATE INDEX idx_task_activity_task ON task_activities(workspace_id, task_id, created_at DESC);
-- 创建默认分区，后续可以按月份新增分区
CREATE TABLE task_activities_default PARTITION OF task_activities DEFAULT;


-- ==================== 需求工作项关联表 ====================
-- 1. 需求标签关联表
CREATE TABLE IF NOT EXISTS requirement_labels (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    workspace_id BIGINT NOT NULL REFERENCES workspaces(id),
    requirement_id BIGINT NOT NULL REFERENCES requirement(id) ON DELETE CASCADE,
    label_id BIGINT NOT NULL REFERENCES labels(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (requirement_id, label_id)
);
ALTER TABLE requirement_labels ENABLE ROW LEVEL SECURITY;
ALTER TABLE requirement_labels FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON requirement_labels
    USING (workspace_id = current_setting('app.workspace_id', true)::bigint)
    WITH CHECK (workspace_id = current_setting('app.workspace_id', true)::bigint);
CREATE INDEX idx_requirement_label_req ON requirement_labels(workspace_id, requirement_id) ;
CREATE INDEX idx_requirement_label_label ON requirement_labels(workspace_id, label_id) ;

-- 2. 需求关注人关联表
CREATE TABLE IF NOT EXISTS requirement_watchers (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    workspace_id BIGINT NOT NULL REFERENCES workspaces(id),
    requirement_id BIGINT NOT NULL REFERENCES requirement(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (requirement_id, user_id)
);
ALTER TABLE requirement_watchers ENABLE ROW LEVEL SECURITY;
ALTER TABLE requirement_watchers FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON requirement_watchers
    USING (workspace_id = current_setting('app.workspace_id', true)::bigint)
    WITH CHECK (workspace_id = current_setting('app.workspace_id', true)::bigint);
CREATE INDEX idx_requirement_watcher_req ON requirement_watchers(workspace_id, requirement_id) ;
CREATE INDEX idx_requirement_watcher_user ON requirement_watchers(workspace_id, user_id) ;

-- 3. 需求关联关系表
CREATE TABLE IF NOT EXISTS requirement_relations (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    workspace_id BIGINT NOT NULL REFERENCES workspaces(id),
    source_requirement_id BIGINT NOT NULL REFERENCES requirement(id) ON DELETE CASCADE,
    target_requirement_id BIGINT NOT NULL REFERENCES requirement(id) ON DELETE CASCADE,
    relation_type TEXT NOT NULL CHECK (relation_type IN ('duplicate','relates_to','implemented_by')),
    created_by BIGINT NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (source_requirement_id, target_requirement_id, relation_type)
);
ALTER TABLE requirement_relations ENABLE ROW LEVEL SECURITY;
ALTER TABLE requirement_relations FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON requirement_relations
    USING (workspace_id = current_setting('app.workspace_id', true)::bigint)
    WITH CHECK (workspace_id = current_setting('app.workspace_id', true)::bigint);
CREATE INDEX idx_requirement_relation_source ON requirement_relations(workspace_id, source_requirement_id) ;
CREATE INDEX idx_requirement_relation_target ON requirement_relations(workspace_id, target_requirement_id) ;

-- 4. 需求评论表
CREATE TABLE IF NOT EXISTS requirement_comments (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    workspace_id BIGINT NOT NULL REFERENCES workspaces(id),
    requirement_id BIGINT NOT NULL REFERENCES requirement(id) ON DELETE CASCADE,
    content_json JSONB NOT NULL,
    content_html TEXT NOT NULL,
    parent_id BIGINT REFERENCES requirement_comments(id) ON DELETE CASCADE,
    created_by BIGINT NOT NULL REFERENCES users(id),
    updated_by BIGINT REFERENCES users(id),
    deleted_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
ALTER TABLE requirement_comments ENABLE ROW LEVEL SECURITY;
ALTER TABLE requirement_comments FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON requirement_comments
    USING (workspace_id = current_setting('app.workspace_id', true)::bigint)
    WITH CHECK (workspace_id = current_setting('app.workspace_id', true)::bigint);
CREATE INDEX idx_requirement_comment_req ON requirement_comments(workspace_id, requirement_id) WHERE deleted_at IS NULL;

-- 5. 需求活动记录表（分区表）
CREATE TABLE IF NOT EXISTS requirement_activities (
    id BIGINT GENERATED ALWAYS AS IDENTITY,
    workspace_id BIGINT NOT NULL REFERENCES workspaces(id),
    requirement_id BIGINT NOT NULL REFERENCES requirement(id) ON DELETE CASCADE,
    verb TEXT NOT NULL CHECK (verb IN ('created','updated','transitioned','attached','linked','commented')),
    field_name TEXT,
    old_value TEXT,
    new_value TEXT,
    actor_id BIGINT NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (id, created_at)
) PARTITION BY RANGE (created_at);
ALTER TABLE requirement_activities ENABLE ROW LEVEL SECURITY;
ALTER TABLE requirement_activities FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON requirement_activities
    USING (workspace_id = current_setting('app.workspace_id', true)::bigint)
    WITH CHECK (workspace_id = current_setting('app.workspace_id', true)::bigint);
CREATE INDEX idx_requirement_activity_req ON requirement_activities(workspace_id, requirement_id, created_at DESC);
CREATE TABLE requirement_activities_default PARTITION OF requirement_activities DEFAULT;


-- ==================== 缺陷工作项关联表 ====================
-- 1. 缺陷关注人关联表
CREATE TABLE IF NOT EXISTS defect_watchers (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    workspace_id BIGINT NOT NULL REFERENCES workspaces(id),
    defect_id BIGINT NOT NULL REFERENCES defect(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (defect_id, user_id)
);
ALTER TABLE defect_watchers ENABLE ROW LEVEL SECURITY;
ALTER TABLE defect_watchers FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON defect_watchers
    USING (workspace_id = current_setting('app.workspace_id', true)::bigint)
    WITH CHECK (workspace_id = current_setting('app.workspace_id', true)::bigint);
CREATE INDEX idx_defect_watcher_defect ON defect_watchers(workspace_id, defect_id) ;
CREATE INDEX idx_defect_watcher_user ON defect_watchers(workspace_id, user_id) ;

-- 2. 缺陷关联关系表
CREATE TABLE IF NOT EXISTS defect_relations (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    workspace_id BIGINT NOT NULL REFERENCES workspaces(id),
    source_defect_id BIGINT NOT NULL REFERENCES defect(id) ON DELETE CASCADE,
    target_defect_id BIGINT NOT NULL REFERENCES defect(id) ON DELETE CASCADE,
    relation_type TEXT NOT NULL CHECK (relation_type IN ('duplicate','relates_to','blocked_by','found_in','fixed_in','verified_in')),
    created_by BIGINT NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (source_defect_id, target_defect_id, relation_type)
);
ALTER TABLE defect_relations ENABLE ROW LEVEL SECURITY;
ALTER TABLE defect_relations FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON defect_relations
    USING (workspace_id = current_setting('app.workspace_id', true)::bigint)
    WITH CHECK (workspace_id = current_setting('app.workspace_id', true)::bigint);
CREATE INDEX idx_defect_relation_source ON defect_relations(workspace_id, source_defect_id) ;
CREATE INDEX idx_defect_relation_target ON defect_relations(workspace_id, target_defect_id) ;

-- 3. 缺陷评论表
CREATE TABLE IF NOT EXISTS defect_comments (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    workspace_id BIGINT NOT NULL REFERENCES workspaces(id),
    defect_id BIGINT NOT NULL REFERENCES defect(id) ON DELETE CASCADE,
    content_json JSONB NOT NULL,
    content_html TEXT NOT NULL,
    parent_id BIGINT REFERENCES defect_comments(id) ON DELETE CASCADE,
    created_by BIGINT NOT NULL REFERENCES users(id),
    updated_by BIGINT REFERENCES users(id),
    deleted_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
ALTER TABLE defect_comments ENABLE ROW LEVEL SECURITY;
ALTER TABLE defect_comments FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON defect_comments
    USING (workspace_id = current_setting('app.workspace_id', true)::bigint)
    WITH CHECK (workspace_id = current_setting('app.workspace_id', true)::bigint);
CREATE INDEX idx_defect_comment_defect ON defect_comments(workspace_id, defect_id) WHERE deleted_at IS NULL;

-- 4. 缺陷活动记录表（分区表）
CREATE TABLE IF NOT EXISTS defect_activities (
    id BIGINT GENERATED ALWAYS AS IDENTITY,
    workspace_id BIGINT NOT NULL REFERENCES workspaces(id),
    defect_id BIGINT NOT NULL REFERENCES defect(id) ON DELETE CASCADE,
    verb TEXT NOT NULL CHECK (verb IN ('created','updated','transitioned','attached','linked','commented','verified')),
    field_name TEXT,
    old_value TEXT,
    new_value TEXT,
    actor_id BIGINT NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (id, created_at)
) PARTITION BY RANGE (created_at);
ALTER TABLE defect_activities ENABLE ROW LEVEL SECURITY;
ALTER TABLE defect_activities FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON defect_activities
    USING (workspace_id = current_setting('app.workspace_id', true)::bigint)
    WITH CHECK (workspace_id = current_setting('app.workspace_id', true)::bigint);
CREATE INDEX idx_defect_activity_defect ON defect_activities(workspace_id, defect_id, created_at DESC);
CREATE TABLE defect_activities_default PARTITION OF defect_activities DEFAULT;

-- ==================== 跨类型关联表 ====================
-- 跨类型工作项关联，替代原来的issue_relations通用关联表
CREATE TABLE IF NOT EXISTS biz_entity_relations (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    workspace_id BIGINT NOT NULL REFERENCES workspaces(id),
    project_id BIGINT NOT NULL REFERENCES projects(id),
    source_type TEXT NOT NULL CHECK (source_type IN ('task','requirement','defect')),
    source_id BIGINT NOT NULL,
    target_type TEXT NOT NULL CHECK (target_type IN ('task','requirement','defect')),
    target_id BIGINT NOT NULL,
    relation_type TEXT NOT NULL CHECK (relation_type IN ('implemented_by','relates_to','duplicate','blocked_by','parent_child','found_in','fixed_in','verified_in')),
    created_by BIGINT NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (source_type, source_id, target_type, target_id, relation_type)
);
ALTER TABLE biz_entity_relations ENABLE ROW LEVEL SECURITY;
ALTER TABLE biz_entity_relations FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON biz_entity_relations
    USING (workspace_id = current_setting('app.workspace_id', true)::bigint)
    WITH CHECK (workspace_id = current_setting('app.workspace_id', true)::bigint);
CREATE INDEX idx_biz_entity_rel_source ON biz_entity_relations(workspace_id, source_type, source_id) ;
CREATE INDEX idx_biz_entity_rel_target ON biz_entity_relations(workspace_id, target_type, target_id) ;

-- ==================== 删除旧通用关联表 ====================
-- 所有旧issues通用关联表全部删除，已经无用处
DROP TABLE IF EXISTS issue_assignees;
DROP TABLE IF EXISTS issue_labels;
DROP TABLE IF EXISTS issue_modules;
DROP TABLE IF EXISTS issue_watchers;
DROP TABLE IF EXISTS issue_relations;
DROP TABLE IF EXISTS issue_dependencies;
DROP TABLE IF EXISTS issue_reactions;
DROP TABLE IF EXISTS issue_votes;
DROP TABLE IF EXISTS issue_subscriptions;
DROP TABLE IF EXISTS issue_comments;
DROP TABLE IF EXISTS issue_activities;
DROP TABLE IF EXISTS issue_sequences;
DROP TABLE IF EXISTS project_sequences;
