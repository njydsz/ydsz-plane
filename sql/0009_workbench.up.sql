-- 0009_workbench: personal workbench + layout persistence + recent items (Sprint 8 — M5 工作台)
-- 参考: Plane / Linear / Jira dashboard & workbench

-- -----------------------------------------------------------------
-- workbench_configs: 工作台布局持久化
-- 每个用户的每项目一个布局配置（widgets + 折叠状态）
-- -----------------------------------------------------------------
CREATE TABLE workbench_configs (
    id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    workspace_id    BIGINT NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    project_id      BIGINT REFERENCES projects(id) ON DELETE CASCADE,  -- NULL = workspace-level
    user_id         BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    layout          JSONB NOT NULL DEFAULT '{}'::jsonb,  -- 拖拽布局配置
    widget_states   JSONB NOT NULL DEFAULT '{}'::jsonb,  -- 各 widget 展开/折叠状态
    focus_enabled   BOOLEAN NOT NULL DEFAULT FALSE,    -- 当前是否开启 Focus Mode
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, COALESCE(project_id, 0))  -- 每用户每项目唯一
);
CREATE INDEX idx_workbench_configs_user ON workbench_configs(user_id, project_id);
CREATE INDEX idx_workbench_configs_project ON workbench_configs(project_id) WHERE project_id IS NOT NULL;

ALTER TABLE workbench_configs ENABLE ROW LEVEL SECURITY;
ALTER TABLE workbench_configs FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON workbench_configs
    USING (workspace_id = current_setting('app.workspace_id', true)::bigint)
    WITH CHECK (workspace_id = current_setting('app.workspace_id', true)::bigint);

CREATE TRIGGER trg_workbench_configs_updated_at BEFORE UPDATE ON workbench_configs
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- -----------------------------------------------------------------
-- recent_items: 最近访问项目/工作项/迭代 (workspace 内跨项目)
-- 用于工作台"最近访问"快速恢复上下文
-- -----------------------------------------------------------------
CREATE TABLE recent_items (
    id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    workspace_id    BIGINT NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    user_id         BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    item_type       TEXT NOT NULL CHECK (item_type IN ('project','issue','sprint','version')),
    item_id         BIGINT NOT NULL,
    project_id      BIGINT REFERENCES projects(id) ON DELETE CASCADE,
    title           TEXT,                                 -- 冗余快照避免 JOIN
    identifier      TEXT,                                 -- 如 YD-123
    accessed_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, item_type, item_id)
);
CREATE INDEX idx_recent_items_user ON recent_items(user_id, accessed_at DESC);
CREATE INDEX idx_recent_items_ws ON recent_items(workspace_id, user_id);

-- -----------------------------------------------------------------
-- updated_at 触发器 (replace accessed_at on conflict upsert)
-- -----------------------------------------------------------------
CREATE OR REPLACE FUNCTION fn_touch_recent_item()
RETURNS TRIGGER AS $$
BEGIN
    NEW.accessed_at := now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_recent_items_touch
    BEFORE UPDATE ON recent_items
    FOR EACH ROW EXECUTE FUNCTION fn_touch_recent_item();

-- -----------------------------------------------------------------
-- workbench_templates: 工作台模板（敏捷/通用/PMO 预设）
-- 预定义 widget 布局，供用户初始化工作台
-- -----------------------------------------------------------------
CREATE TABLE workbench_templates (
    id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name            TEXT NOT NULL,
    slug            TEXT NOT NULL UNIQUE,
    description     TEXT,
    layout          JSONB NOT NULL DEFAULT '{}'::jsonb,
    icon            TEXT,
    is_default      BOOLEAN NOT NULL DEFAULT FALSE,
    sort_order      INT NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_workbench_templates_default ON workbench_templates(is_default);

-- 插入模板数据
INSERT INTO workbench_templates (name, slug, description, layout, is_default, sort_order) VALUES
('敏捷开发', 'agile', 'Scrum/Kanban 团队的默认工作台，含迭代概览、待办事项、燃尽任务、阻塞工作项', '{"widgets":[{"type":"my_issues","w":6,"h":4},{"type":"sprint_overview","w":6,"h":3},{"type":"overdue","w":4,"h":2},{"type":"recent","w":4,"h":3},{"type":"quick_actions","w":4,"h":2}]}', TRUE, 1),
('项目监控', 'pmo', 'PMO/管理者视角，关注项目进度、逾期预警、团队速率', '{"widgets":[{"type":"project_overview","w":12,"h":2},{"type":"overdue","w":6,"h":3},{"type":"risk_alert","w":6,"h":3},{"type":"recent","w":6,"h":3},{"type":"quick_actions","w":6,"h":2}]}', FALSE, 2),
('个人聚焦', 'focus', '专注模式，只有待办和专注计时器', '{"widgets":[{"type":"my_issues","w":12,"h":6},{"type":"focus_timer","w":6,"h":2},{"type":"overdue","w":6,"h":2}]}', FALSE, 3);
