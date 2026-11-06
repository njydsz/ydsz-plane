-- V26.11.03__teamspace.sql — 团队空间（Teamspaces）
--
-- 对标 Plane Teamspaces：团队 = 一组人 + 一组关联项目。
-- 聚合展示团队工作项、迭代 rollup 与进度。

-- ---------------------------------------------------------------------------
-- teamspaces
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS teamspaces (
    id          BIGINT       PRIMARY KEY,
    workspace_id BIGINT      NOT NULL,
    name        VARCHAR(200) NOT NULL,
    description TEXT,
    color       VARCHAR(20),
    lead_id     BIGINT,
    created_by  BIGINT,
    archived_at TIMESTAMPTZ,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_teamspaces_workspace_id ON teamspaces (workspace_id);
CREATE INDEX IF NOT EXISTS idx_teamspaces_workspace_active ON teamspaces (workspace_id, archived_at);

-- ---------------------------------------------------------------------------
-- teamspace_members
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS teamspace_members (
    teamspace_id BIGINT      NOT NULL,
    user_id      BIGINT      NOT NULL,
    role         VARCHAR(20) NOT NULL DEFAULT 'member',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (teamspace_id, user_id),
    CONSTRAINT chk_teamspace_member_role CHECK (role IN ('lead', 'member'))
);

CREATE INDEX IF NOT EXISTS idx_teamspace_members_user_id ON teamspace_members (user_id);

-- ---------------------------------------------------------------------------
-- teamspace_projects
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS teamspace_projects (
    teamspace_id BIGINT      NOT NULL,
    project_id   BIGINT      NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (teamspace_id, project_id)
);

CREATE INDEX IF NOT EXISTS idx_teamspace_projects_project_id ON teamspace_projects (project_id);

COMMENT ON TABLE teamspaces IS '团队空间（一组人 + 一组关联项目）';
COMMENT ON TABLE teamspace_members IS '团队成员';
COMMENT ON TABLE teamspace_projects IS '团队关联项目';
