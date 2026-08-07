-- 0006_sprint_core: sprints + sprint_issues + sprint_snapshots (Sprint 5 — M3 迭代与排期)
-- 参考 docs/architecture/08-迭代与版本日设计.md

-- -----------------------------------------------------------------
-- sprints: 迭代聚合根
-- status: planned | active | completed
-- -----------------------------------------------------------------
CREATE TABLE sprints (
    id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    workspace_id    BIGINT NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    project_id      BIGINT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name            TEXT NOT NULL,
    description     TEXT,
    goal            TEXT,                                -- 迭代目标
    status          TEXT NOT NULL DEFAULT 'planned'
                    CHECK (status IN ('planned','active','completed')),
    start_date      DATE,
    end_date        DATE,
    capacity        NUMERIC(10,2),                       -- 容量（故事点或人天，与 point 单位一致）
    owner_id        BIGINT REFERENCES users(id),
    viewport        JSONB NOT NULL DEFAULT '{}'::jsonb,  -- 视图偏好（折叠/展开）
    -- 复盘数据（结束瞬间快照）
    review_snapshot JSONB,                               -- {committed_points, completed_points, joined_points, removed_points, issues_breakdown}
    started_at      TIMESTAMPTZ,
    completed_at    TIMESTAMPTZ,
    created_by      BIGINT NOT NULL REFERENCES users(id),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ,
    CONSTRAINT sprint_date_range CHECK (
        start_date IS NULL OR end_date IS NULL OR start_date <= end_date
    )
);
CREATE INDEX idx_sprints_project_status ON sprints(project_id, status) WHERE deleted_at IS NULL;
CREATE INDEX idx_sprints_active_unique ON sprints(project_id, status)
    WHERE status = 'active' AND deleted_at IS NULL;  -- 辅助唯一性（配合 DB 约束）

ALTER TABLE sprints ENABLE ROW LEVEL SECURITY;
ALTER TABLE sprints FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON sprints
    USING (workspace_id = current_setting('app.workspace_id', true)::bigint)
    WITH CHECK (workspace_id = current_setting('app.workspace_id', true)::bigint);

-- -----------------------------------------------------------------
-- sprint_issues: 迭代与工作项 M2M
-- added_midway: 中途加入标记（影响速率统计口径）
-- -----------------------------------------------------------------
CREATE TABLE sprint_issues (
    sprint_id       BIGINT NOT NULL REFERENCES sprints(id) ON DELETE CASCADE,
    issue_id        BIGINT NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
    added_midway    BOOLEAN NOT NULL DEFAULT FALSE,      -- 迭代启动后加入（影响复盘数据）
    sort_order      DOUBLE PRECISION NOT NULL DEFAULT 65535,
    added_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    added_by        BIGINT REFERENCES users(id),
    PRIMARY KEY (sprint_id, issue_id)
);
CREATE INDEX idx_sprint_issues_issue ON sprint_issues(issue_id);

-- -----------------------------------------------------------------
-- sprint_snapshots: 每日快照（00:05 Cron）
-- 用于燃尽图 / 燃起图 / CFD
-- -----------------------------------------------------------------
CREATE TABLE sprint_snapshots (
    id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    workspace_id    BIGINT NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    project_id      BIGINT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    sprint_id       BIGINT NOT NULL REFERENCES sprints(id) ON DELETE CASCADE,
    snapshot_date   DATE NOT NULL,                       -- 快照UTC日期
    data            JSONB NOT NULL DEFAULT '{}'::jsonb,  -- 见 data shape 注释
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ
);
CREATE UNIQUE INDEX idx_sprint_snapshots_unique ON sprint_snapshots(sprint_id, snapshot_date);
-- data shape:
-- {
--   "total_points": 80,
--   "done_points": 32,
--   "total_issues": 25,
--   "done_issues": 10,
--   "by_state_group": {"backlog": 0, "unstarted": 30, "started": 18, "completed": 32},
--   "added_points": 5, "removed_points": 0
-- }
CREATE INDEX idx_sprintsnapshots_sprint_date ON sprint_snapshots(sprint_id, snapshot_date);
CREATE INDEX idx_sprintsnapshots_project ON sprint_snapshots(project_id) WHERE deleted_at IS NULL;

ALTER TABLE sprint_snapshots ENABLE ROW LEVEL SECURITY;
ALTER TABLE sprint_snapshots FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON sprint_snapshots
    USING (workspace_id = current_setting('app.workspace_id', true)::bigint)
    WITH CHECK (workspace_id = current_setting('app.workspace_id', true)::bigint);

-- -----------------------------------------------------------------
-- updated_at 触发器
-- -----------------------------------------------------------------
CREATE TRIGGER trg_sprints_updated_at BEFORE UPDATE ON sprints
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- -----------------------------------------------------------------
-- 辅助：迭代 active 唯一性约束
-- 项目维度仅允许一个 active 迭代
-- -----------------------------------------------------------------
CREATE UNIQUE INDEX idx_one_active_sprint_per_project
    ON sprints(project_id)
    WHERE status = 'active' AND deleted_at IS NULL;
