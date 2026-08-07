-- 0007_version_core: versions + issue 缺陷版本字段 (Sprint 6 — M4 质量与版本)
-- 参考 docs/architecture/08-迭代与版本设计.md
-- 注意: 版本与迭代的 1:N 关联在 0009_version_fix 迁移中实现 (sprints.version_id FK)

-- -----------------------------------------------------------------
-- versions: 版本聚合根（跨迭代的发布聚合）
-- status: planning | active | released | archived
-- 属性: name, semver, description, start_date, end_date, target_date(版本日) 等
-- -----------------------------------------------------------------
CREATE TABLE versions (
    id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    workspace_id    BIGINT NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    project_id      BIGINT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name            TEXT NOT NULL,                        -- 版本展示名（如"用户画像一期"）
    semver          TEXT NOT NULL,                        -- 语义版本号（如 1.2.3）
    description     TEXT,
    status          TEXT NOT NULL DEFAULT 'planning'
                    CHECK (status IN ('planning','active','released','archived')),
    checklist       JSONB NOT NULL DEFAULT '[]'::jsonb,  -- 发布检查清单 [{id,label,required,checked}]
    release_notes   TEXT,                                 -- 已生成的 Release Notes（可编辑）
    delivered_at    TIMESTAMPTZ,                          -- 实际发布时间
    target_date     DATE,                                 -- 计划发布日
    archived_at     TIMESTAMPTZ,
    created_by      BIGINT NOT NULL REFERENCES users(id),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,
    -- 项目内 semver 未删除时唯一
    CONSTRAINT versions_semver_valid CHECK (semver ~ '^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(-[0-9A-Za-z\-.]+)?(\+[0-9A-Za-z\-.]+)?$')
);
CREATE INDEX idx_versions_project_status ON versions(project_id, status) WHERE deleted_at IS NULL;
CREATE INDEX idx_versions_workspace ON versions(workspace_id) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX idx_versions_unique_semver ON versions(project_id, semver) WHERE deleted_at IS NULL;

ALTER TABLE versions ENABLE ROW LEVEL SECURITY;
ALTER TABLE versions FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON versions
    USING (workspace_id = current_setting('app.workspace_id', true)::bigint)
    WITH CHECK (workspace_id = current_setting('app.workspace_id', true)::bigint);

-- -----------------------------------------------------------------
-- version_sprints: 已弃用（由 0009_version_fix 迁移为 sprints.version_id FK）
-- 保留此注释仅为历史参考，实际表不再创建
-- -----------------------------------------------------------------
CREATE TABLE version_sprints (
    version_id      BIGINT NOT NULL REFERENCES versions(id) ON DELETE CASCADE,
    sprint_id       BIGINT NOT NULL REFERENCES sprints(id) ON DELETE CASCADE,
    added_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    added_by        BIGINT REFERENCES users(id),
    PRIMARY KEY (version_id, sprint_id)
);
CREATE INDEX idx_version_sprints_sprint ON version_sprints(sprint_id);

ALTER TABLE version_sprints ENABLE ROW LEVEL SECURITY;
ALTER TABLE version_sprints FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON version_sprints
    USING (version_id IN (SELECT id FROM versions WHERE workspace_id = current_setting('app.workspace_id', true)::bigint));

-- -----------------------------------------------------------------
-- versions.updated_at 触发器
-- -----------------------------------------------------------------
CREATE TRIGGER trg_versions_updated_at BEFORE UPDATE ON versions
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- -----------------------------------------------------------------
-- issues: 缺陷版本关联字段
-- found_version_id:   发现版本（必填 — 与 found_phase 类似）
-- fix_version_id:     修复版本（缺陷解决时关联）
-- release_version_id: 首次发布版本（需求/任务完成时的首个发版版本）
-- -----------------------------------------------------------------
ALTER TABLE issues ADD COLUMN found_version_id   BIGINT REFERENCES versions(id);
ALTER TABLE issues ADD COLUMN fix_version_id     BIGINT REFERENCES versions(id);
ALTER TABLE issues ADD COLUMN release_version_id BIGINT REFERENCES versions(id);

CREATE INDEX idx_issues_found_version   ON issues(project_id, found_version_id) WHERE deleted_at IS NULL AND found_version_id IS NOT NULL;
CREATE INDEX idx_issues_fix_version     ON issues(project_id, fix_version_id) WHERE deleted_at IS NULL AND fix_version_id IS NOT NULL;
CREATE INDEX idx_issues_release_version ON issues(project_id, release_version_id) WHERE deleted_at IS NULL AND release_version_id IS NOT NULL;
