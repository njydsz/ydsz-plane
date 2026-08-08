-- 迁移 0026：Issue 版本快照审计（对标 Plane 的 IssueVersion / Activity History）
-- 每次 issues 表发生 UPDATE 时，由应用层插入旧值快照，支持回溯与对比。
-- 遵循项目多租户约束：含 workspace_id NOT NULL，开启 RLS。

CREATE TABLE IF NOT EXISTS issue_versions (
    id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    workspace_id    BIGINT NOT NULL REFERENCES workspaces(id),
    project_id      BIGINT NOT NULL REFERENCES projects(id),
    issue_id        BIGINT NOT NULL,            -- 指向 issues.id（兼容分表后逻辑 ID）
    version         INT NOT NULL,               -- 递增版本号（与 issues.version 对应）
    snapshot        JSONB NOT NULL,             -- 完整字段快照（不含 description 大字段可配置裁剪）
    changed_fields  TEXT[] DEFAULT '{}',        -- 本次变更的字段名列表
    change_type     TEXT NOT NULL DEFAULT 'update'
                    CHECK (change_type IN ('create','update','delete','transition')),
    created_by      BIGINT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (issue_id, version) WHERE true       -- 同一 issue 版本号唯一
);

-- RLS（对齐项目现有租户隔离策略）
ALTER TABLE issue_versions ENABLE ROW LEVEL SECURITY;
ALTER TABLE issue_versions FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON issue_versions
    USING (workspace_id = current_setting('app.workspace_id', true)::bigint);

-- 索引
CREATE INDEX idx_issue_versions_issue   ON issue_versions(workspace_id, issue_id, version DESC);
CREATE INDEX idx_issue_versions_project ON issue_versions(workspace_id, project_id, created_at DESC);
CREATE INDEX idx_issue_versions_actor   ON issue_versions(workspace_id, created_by) WHERE created_by IS NOT NULL;

COMMENT ON TABLE issue_versions IS '工作项版本快照：记录每次变更前的字段状态，支撑审计回溯与变更对比';
COMMENT ON COLUMN issue_versions.snapshot IS '变更前完整字段快照（JSONB，对应 BaseWorkitem 结构）';
COMMENT ON COLUMN issue_versions.changed_fields IS '本次变更涉及的字段名，便于 diff 视图渲染';
COMMENT ON COLUMN issue_versions.change_type IS '变更类型：create(创建) / update(字段更新) / delete(软删除) / transition(状态流转)';
COMMENT ON COLUMN issue_versions.version IS '递增版本号；与 issues.version 一一对应';
