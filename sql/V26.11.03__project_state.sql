-- V26.11.03__project_state.sql — 项目状态与更新
--
-- 对标 Plane Project States + Updates：
--   - project_states           ：项目状态/优先级/起止日期（每项目一条）
--   - project_updates          ：项目更新（状态快照 + 文本）
--   - project_update_reactions ：更新表情回应
--   - project_update_comments  ：更新评论（支持回复）

-- ---------------------------------------------------------------------------
-- project_states
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS project_states (
    id          BIGINT       PRIMARY KEY,
    project_id  BIGINT       NOT NULL UNIQUE,
    status      VARCHAR(20)  NOT NULL DEFAULT 'on_track',
    priority    VARCHAR(20)  NOT NULL DEFAULT 'medium',
    start_date  DATE,
    target_date DATE,
    owner_id    BIGINT,
    description TEXT,
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    CONSTRAINT chk_project_state_status CHECK (status IN ('on_track','at_risk','off_track','completed','paused')),
    CONSTRAINT chk_project_state_priority CHECK (priority IN ('urgent','high','medium','low','none'))
);

CREATE INDEX IF NOT EXISTS idx_project_states_status ON project_states (status);

-- ---------------------------------------------------------------------------
-- project_updates
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS project_updates (
    id                BIGINT      PRIMARY KEY,
    workspace_id      BIGINT      NOT NULL,
    project_id        BIGINT      NOT NULL,
    author_id         BIGINT      NOT NULL,
    status            VARCHAR(20) NOT NULL,
    content           TEXT        NOT NULL,
    progress_percent  NUMERIC(5,2) NOT NULL DEFAULT 0,
    initiative_id     BIGINT,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT chk_project_update_status CHECK (status IN ('on_track','at_risk','off_track','completed','paused'))
);

CREATE INDEX IF NOT EXISTS idx_project_updates_project ON project_updates (project_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_project_updates_initiative ON project_updates (initiative_id) WHERE initiative_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_project_updates_workspace ON project_updates (workspace_id, created_at DESC);

-- ---------------------------------------------------------------------------
-- project_update_reactions
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS project_update_reactions (
    id         BIGINT      PRIMARY KEY,
    update_id  BIGINT      NOT NULL,
    user_id    BIGINT      NOT NULL,
    emoji      VARCHAR(32) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (update_id, user_id, emoji)
);

CREATE INDEX IF NOT EXISTS idx_project_update_reactions_update ON project_update_reactions (update_id);

-- ---------------------------------------------------------------------------
-- project_update_comments
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS project_update_comments (
    id         BIGINT      PRIMARY KEY,
    update_id  BIGINT      NOT NULL,
    author_id  BIGINT      NOT NULL,
    parent_id  BIGINT,
    content    TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT chk_project_update_comment_not_self_parent CHECK (parent_id IS NULL OR parent_id <> id)
);

CREATE INDEX IF NOT EXISTS idx_project_update_comments_update ON project_update_comments (update_id, created_at);
CREATE INDEX IF NOT EXISTS idx_project_update_comments_parent ON project_update_comments (parent_id) WHERE parent_id IS NOT NULL;

COMMENT ON TABLE project_states IS '项目状态（on_track/at_risk/off_track/completed/paused）';
COMMENT ON TABLE project_updates IS '项目更新（状态快照 + 文本）';
COMMENT ON TABLE project_update_reactions IS '项目更新表情回应';
COMMENT ON TABLE project_update_comments IS '项目更新评论';
