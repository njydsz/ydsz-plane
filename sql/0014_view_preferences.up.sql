-- 0014_view_preferences.up.sql — 视图偏好持久化（看板/列表布局、列配置、过滤条件）
BEGIN;

CREATE TABLE IF NOT EXISTS view_preferences (
    id            BIGSERIAL PRIMARY KEY,
    workspace_id  BIGINT NOT NULL,
    project_id    BIGINT NOT NULL,
    user_id       BIGINT NOT NULL,
    view_type     VARCHAR(20) NOT NULL,           -- 'kanban' | 'list' | 'calendar' | 'gantt'
    layout        VARCHAR(20) NOT NULL DEFAULT 'list',
    columns       JSONB NOT NULL DEFAULT '[]',    -- 列表列配置（key/label/width/visible/order）
    filters       JSONB NOT NULL DEFAULT '{}',    -- 过滤条件（ListIssuesParams 快照）
    sort          JSONB NOT NULL DEFAULT '{}',    -- 排序配置（{field, desc}）
    extra         JSONB NOT NULL DEFAULT '{}',    -- 扩展配置（如看板分组字段、密度）
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (workspace_id, project_id, user_id, view_type)
);

CREATE INDEX IF NOT EXISTS idx_view_prefs_user
    ON view_preferences (workspace_id, user_id);

COMMIT;
