-- 0010_dashboard: project dashboard + widget data + risk alerts + templates (Sprint 9 — M5 仪表盘)
-- 参考: Plane / Jira dashboard / Linear project overview

-- dashboard_widgets: widget 配置（per-project）
CREATE TABLE dashboard_widgets (
    id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    project_id      BIGINT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    widget_type     TEXT NOT NULL CHECK (widget_type IN (
        'progress_overview','burndown','velocity','priority_split',
        'state_distribution','overdue_list','blocked_list',
        'risk_alert','recent_activity','team_workload'
    )),
    title           TEXT NOT NULL,
    grid_x          INT NOT NULL DEFAULT 0,
    grid_y          INT NOT NULL DEFAULT 0,
    grid_w          INT NOT NULL DEFAULT 4,
    grid_h          INT NOT NULL DEFAULT 3,
    config          JSONB NOT NULL DEFAULT '{}'::jsonb,
    is_visible      BOOLEAN NOT NULL DEFAULT TRUE,
    sort_order      INT NOT NULL DEFAULT 0,
    user_id         BIGINT REFERENCES users(id) ON DELETE CASCADE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_dashboard_widgets_project ON dashboard_widgets(project_id, sort_order);
CREATE INDEX idx_dashboard_widgets_user ON dashboard_widgets(user_id) WHERE user_id IS NOT NULL;

ALTER TABLE dashboard_widgets ENABLE ROW LEVEL SECURITY;
ALTER TABLE dashboard_widgets FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON dashboard_widgets
    USING (project_id IN (SELECT id FROM projects WHERE workspace_id = current_setting('app.workspace_id', true)::bigint));

CREATE TRIGGER trg_dashboard_widgets_updated_at BEFORE UPDATE ON dashboard_widgets
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- dashboard_snapshots: widget 数据快照加速首屏
CREATE TABLE dashboard_snapshots (
    id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    project_id      BIGINT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    widget_type     TEXT NOT NULL,
    data            JSONB NOT NULL DEFAULT '{}'::jsonb,
    refreshed_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (project_id, widget_type)
);
CREATE INDEX idx_dashboard_snapshots_project ON dashboard_snapshots(project_id);
CREATE INDEX idx_dashboard_snapshots_refreshed ON dashboard_snapshots(refreshed_at);

ALTER TABLE dashboard_snapshots ENABLE ROW LEVEL SECURITY;
ALTER TABLE dashboard_snapshots FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON dashboard_snapshots
    USING (project_id IN (SELECT id FROM projects WHERE workspace_id = current_setting('app.workspace_id', true)::bigint));

-- risk_rules: 风险预警规则
CREATE TABLE risk_rules (
    id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    workspace_id    BIGINT NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    project_id      BIGINT REFERENCES projects(id) ON DELETE CASCADE,
    rule_name       TEXT NOT NULL,
    rule_type       TEXT NOT NULL CHECK (rule_type IN (
        'overdue_issue','overdue_sprint','blocked_count','sla_breach',
        'stalled_progress','high_priority_open'
    )),
    condition_json  JSONB NOT NULL DEFAULT '{}'::jsonb,
    notify_channels TEXT[] NOT NULL DEFAULT '{}',
    is_active       BOOLEAN NOT NULL DEFAULT TRUE,
    last_triggered  TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_risk_rules_workspace ON risk_rules(workspace_id);
CREATE INDEX idx_risk_rules_project ON risk_rules(project_id) WHERE project_id IS NOT NULL;
CREATE INDEX idx_risk_rules_active ON risk_rules(is_active) WHERE is_active = TRUE;

ALTER TABLE risk_rules ENABLE ROW LEVEL SECURITY;
ALTER TABLE risk_rules FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON risk_rules
    USING (workspace_id = current_setting('app.workspace_id', true)::bigint);

CREATE TRIGGER trg_risk_rules_updated_at BEFORE UPDATE ON risk_rules
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- risk_alerts: 风险告警记录
CREATE TABLE risk_alerts (
    id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    workspace_id    BIGINT NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    project_id      BIGINT REFERENCES projects(id) ON DELETE CASCADE,
    rule_id         BIGINT NOT NULL REFERENCES risk_rules(id) ON DELETE CASCADE,
    severity        TEXT NOT NULL DEFAULT 'medium' CHECK (severity IN ('info','low','medium','high','critical')),
    title           TEXT NOT NULL,
    description     TEXT,
    metadata        JSONB NOT NULL DEFAULT '{}'::jsonb,
    is_resolved     BOOLEAN NOT NULL DEFAULT FALSE,
    resolved_at     TIMESTAMPTZ,
    resolved_by     BIGINT REFERENCES users(id),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_risk_alerts_project ON risk_alerts(project_id, created_at DESC) WHERE NOT is_resolved;
CREATE INDEX idx_risk_alerts_workspace ON risk_alerts(workspace_id, created_at DESC) WHERE NOT is_resolved;
CREATE INDEX idx_risk_alerts_unresolved ON risk_alerts(workspace_id, severity) WHERE NOT is_resolved;

ALTER TABLE risk_alerts ENABLE ROW LEVEL SECURITY;
ALTER TABLE risk_alerts FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON risk_alerts
    USING (workspace_id = current_setting('app.workspace_id', true)::bigint);

-- dashboard_templates: 仪表盘预设模板
CREATE TABLE dashboard_templates (
    id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name            TEXT NOT NULL,
    slug            TEXT NOT NULL UNIQUE,
    description     TEXT,
    layout          JSONB NOT NULL DEFAULT '{}'::jsonb,
    icon            TEXT,
    category        TEXT NOT NULL DEFAULT 'agile',
    is_default      BOOLEAN NOT NULL DEFAULT FALSE,
    sort_order      INT NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_dashboard_templates_category ON dashboard_templates(category, sort_order);

INSERT INTO dashboard_templates (name, slug, description, layout, icon, category, is_default, sort_order) VALUES
('项目概览', 'project-overview', '项目级核心指标：进度、趋势、风险点', '{"widgets":[{"type":"progress_overview","x":0,"y":0,"w":12,"h":2},{"type":"burndown","x":0,"y":2,"w":6,"h":4},{"type":"risk_alert","x":6,"y":2,"w":3,"h":4},{"type":"overdue_list","x":9,"y":2,"w":3,"h":4},{"type":"recent_activity","x":0,"y":6,"w":6,"h":4},{"type":"team_workload","x":6,"y":6,"w":6,"h":4}]}', 'chart', 'agile', TRUE, 1),
('项目管理', 'pmo-dashboard', 'PMO 视角：多维度统计 + 风险清单', '{"widgets":[{"type":"progress_overview","x":0,"y":0,"w":12,"h":2},{"type":"state_distribution","x":0,"y":2,"w":4,"h":4},{"type":"priority_split","x":4,"y":2,"w":4,"h":4},{"type":"velocity","x":8,"y":2,"w":4,"h":4},{"type":"risk_alert","x":0,"y":6,"w":6,"h":3},{"type":"overdue_list","x":6,"y":6,"w":6,"h":3}]}', 'monitor', 'pmo', FALSE, 2),
('质量看板', 'quality-dashboard', '质量指标：缺陷趋势 + 阻塞分析', '{"widgets":[{"type":"priority_split","x":0,"y":0,"w":6,"h":3},{"type":"blocked_list","x":6,"y":0,"w":6,"h":3},{"type":"burndown","x":0,"y":3,"w":12,"h":4},{"type":"recent_activity","x":0,"y":7,"w":12,"h":4}]}', 'bug', 'quality', FALSE, 3);
