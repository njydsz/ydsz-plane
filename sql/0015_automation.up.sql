-- 0015_automation: 自动化规则引擎 (Sprint 11 — M6 开放与智能)
-- 参考 docs/architecture/12-开放集成设计.md §4 自动化引擎

-- -----------------------------------------------------------------
-- automation_rules: 规则聚合根
-- status: draft | active | disabled | error（连续失败 3 次后自动 → disabled）
-- DSL 结构: { trigger, conditions, actions } JSONB 存 automation_rule_dsl
-- -----------------------------------------------------------------
CREATE TABLE automation_rules (
    id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    workspace_id    BIGINT NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    project_id      BIGINT REFERENCES projects(id) ON DELETE CASCADE,
    name            TEXT NOT NULL,
    description     TEXT,
    -- DSL JSONB 结构:
    -- {
    --   "trigger": { "type": "issue.status_changed", "filter": { "to_group": "started" } },
    --   "conditions": { "all": [{ "field": "severity", "op": "<=", "value": 2 }] },
    --   "actions": [{ "type": "assign", "field": "verifier", "value": "${issue.created_by}" }]
    -- }
    dsl             JSONB NOT NULL DEFAULT '{}'::jsonb,
    -- DSL 摘要（API 列表时避免返回大 JSONB）
    trigger_type    TEXT NOT NULL,           -- 冗余 trigger.type 便于索引查询
    action_count    INT NOT NULL DEFAULT 0,  -- actions 数量（列表展示用）
    status          TEXT NOT NULL DEFAULT 'draft'
                    CHECK (status IN ('draft','active','disabled','error')),
    -- 权限与归属
    created_by      BIGINT NOT NULL REFERENCES users(id),
    -- 监控
    last_run_at     TIMESTAMPTZ,
    last_error      TEXT,
    consecutive_failures INT NOT NULL DEFAULT 0,
    execution_count BIGINT NOT NULL DEFAULT 0,  -- 累计执行次数
    -- 顺序
    sort_order      INT NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_automation_rules_project_status ON automation_rules(project_id, status) WHERE project_id IS NOT NULL;
CREATE INDEX idx_automation_rules_ws ON automation_rules(workspace_id);
CREATE INDEX idx_automation_rules_trigger ON automation_rules(project_id, trigger_type) WHERE status = 'active';
CREATE INDEX idx_automation_rules_sort ON automation_rules(workspace_id, sort_order);

ALTER TABLE automation_rules ENABLE ROW LEVEL SECURITY;
ALTER TABLE automation_rules FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON automation_rules
    USING (workspace_id = current_setting('app.workspace_id', true)::bigint)
    WITH CHECK (workspace_id = current_setting('app.workspace_id', true)::bigint);

CREATE TRIGGER trg_automation_rules_updated_at BEFORE UPDATE ON automation_rules
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- -----------------------------------------------------------------
-- rule_executions: 规则执行审计日志
-- -----------------------------------------------------------------
CREATE TABLE rule_executions (
    id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    workspace_id    BIGINT NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    project_id      BIGINT REFERENCES projects(id) ON DELETE CASCADE,
    rule_id         BIGINT NOT NULL REFERENCES automation_rules(id) ON DELETE CASCADE,
    trigger_event_id BIGINT REFERENCES domain_events(id) ON DELETE SET NULL,
    -- 执行结果
    status          TEXT NOT NULL CHECK (status IN ('matched','skipped','success','failed','dry_run')),
    duration_ms     INT,                     -- 执行耗时
    error_message   TEXT,
    -- 上下文快照
    context_json    JSONB,                   -- { issue_id, conditions_result, actions_taken }
    -- 元数据（防循环：链路深度）
    trigger_depth   SMALLINT NOT NULL DEFAULT 0,
    via_automation  BOOLEAN NOT NULL DEFAULT FALSE,  -- 是否由其他规则触发（防循环）
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_rule_executions_rule ON rule_executions(rule_id, created_at DESC);
CREATE INDEX idx_rule_executions_project ON rule_executions(project_id, created_at DESC) WHERE project_id IS NOT NULL;
CREATE INDEX idx_rule_executions_ws ON rule_executions(workspace_id, created_at DESC);
CREATE INDEX idx_rule_executions_event ON rule_executions(trigger_event_id) WHERE trigger_event_id IS NOT NULL;
-- 分区准备：按月分区（预留，首次数据量小暂不分）
-- PARTITION BY RANGE (created_at);

ALTER TABLE rule_executions ENABLE ROW LEVEL SECURITY;
ALTER TABLE rule_executions FORCE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON rule_executions
    USING (workspace_id = current_setting('app.workspace_id', true)::bigint);

-- -----------------------------------------------------------------
-- automation_templates: 内置规则模板（7 条 PRD §9.7）
-- -----------------------------------------------------------------
CREATE TABLE automation_templates (
    id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name            TEXT NOT NULL,
    slug            TEXT NOT NULL UNIQUE,
    description     TEXT,
    category        TEXT NOT NULL DEFAULT 'efficiency',  -- efficiency / quality / notification
    -- 模板 DSL（克隆时替换 ${project_id} 等占位符）
    dsl_template    JSONB NOT NULL DEFAULT '{}'::jsonb,
    icon            TEXT,
    sort_order      INT NOT NULL DEFAULT 0,
    is_recommended  BOOLEAN NOT NULL DEFAULT FALSE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_automation_templates_cat ON automation_templates(category, sort_order);

-- 预置 7 条内置模板
INSERT INTO automation_templates (name, slug, description, category, dsl_template, icon, sort_order, is_recommended) VALUES
('子项全部完成后自动完成父项', 'auto-complete-parent',
 '当所有子工作项都标记为完成时，自动将父工作项状态更新为已完成',
 'quality',
 '{"trigger":{"type":"issue.status_changed","filter":{"to_group":"completed"}},"conditions":{"all":[{"field":"parent","op":"is_not_empty"}]},"actions":[{"type":"transition","field":"state","value":"completed"}]}',
 'git-merge', 1, TRUE),

('逾期提醒', 'overdue-reminder',
 '工作项到期前 1 天自动提醒负责人',
 'notification',
 '{"trigger":{"type":"scheduled","cron":"0 9 * * *","filter":{"due_within_hours":24}},"conditions":{"all":[{"field":"state.group","op":"ne","value":"completed"}]},"actions":[{"type":"notify","channel":"in_app","target":"${issue.assignees}","template":"工作项 {{issue.identifier}} 即将到期"}]}',
 'clock', 2, TRUE),

('版本发布后自动流转工作项', 'version-release-transition',
 '版本发布时，自动将该版本下的工作项状态更新为已完成',
 'efficiency',
 '{"trigger":{"type":"version.released"},"conditions":{"all":[{"field":"issue.fix_version","op":"eq","value":"${version.id}"}]},"actions":[{"type":"transition","field":"state","value":"completed"}]}',
 'rocket', 3, FALSE),

('Epic 点数自动汇总', 'epic-points-rollup',
 '当子工作项点数变更时，自动汇总到 Epic 的聚合点数字段',
 'efficiency',
 '{"trigger":{"type":"issue.updated","filter":{"field_changes":["estimate_points"]}},"conditions":{"all":[{"field":"issue.type_code","op":"ne","value":"epic"}]},"actions":[{"type":"action","type":"copy_field","source":"${issue.sum_children_points}","target":"${parent.estimate_points"}]}',
 'layers', 4, FALSE),

('进入"进行中"时自动填写开始日期', 'auto-start-date',
 '工作项首次进入进行中状态时，自动记录开始时间',
 'efficiency',
 '{"trigger":{"type":"issue.status_changed","filter":{"to_group":"started"}},"conditions":{"all":[{"field":"started_at","op":"is_empty"}]},"actions":[{"type":"update_field","field":"started_at","value":"${now}"}]}',
 'play', 5, TRUE),

('最闲人自动指派', 'auto-assign-least-loaded',
 '新建工作项时自动分配给当前负载最轻的成员',
 'efficiency',
 '{"trigger":{"type":"issue.created"},"conditions":{"all":[{"field":"assignees","op":"is_empty"}]},"actions":[{"type":"assign","strategy":"least_loaded","role":"member","scope":"project"}]}',
 'user-plus', 6, FALSE),

('新缺陷通知技术负责人', 'defect-notify-tech-lead',
 '项目里新建高优缺陷时，自动通知项目技术负责人',
 'notification',
 '{"trigger":{"type":"issue.created","filter":{"type_code":"defect","priority":"urgent"}},"conditions":[],"actions":[{"type":"notify","channel":"in_app","target":"${project.tech_lead}","template":"🚨 新建紧急缺陷: [{{issue.identifier}}] {{issue.name}}"}]}',
 'alert-triangle', 7, TRUE);
