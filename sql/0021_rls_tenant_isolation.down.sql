-- =============================================================
-- 迁移 0021 (回滚): 禁用 RLS
-- =============================================================

DO $$
DECLARE
    tbl TEXT;
    tables TEXT[] := ARRAY[
        'workspaces',
        'workspace_members',
        'projects',
        'project_members',
        'modules',
        'labels',
        'states',
        'state_transitions',
        'issues',
        'issue_assignees',
        'issue_labels',
        'issue_modules',
        'issue_activities',
        'issue_comments',
        'issue_relations',
        'issue_dependencies',
        'sprints',
        'sprint_issues',
        'sprint_snapshots',
        'versions',
        'automation_rules',
        'rule_executions',
        'automation_templates',
        'notifications',
        'notification_preferences',
        'api_tokens',
        'attachments',
        'invitations',
        'intake_channels',
        'intake_issues',
        'workbench_configs',
        'search_documents',
        'dashboard_configs',
        'dashboard_widgets',
        'dashboard_snapshots',
        'metric_snapshots',
        'audit_logs',
        'domain_events',
        'webhook_subscriptions',
        'webhook_logs',
        'preferences',
        'estimate_points',
        'project_templates'
    ];
BEGIN
    FOREACH tbl IN ARRAY tables
    LOOP
        IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = tbl) THEN
            EXECUTE format('ALTER TABLE %I DISABLE ROW LEVEL SECURITY', tbl);
        END IF;
    END LOOP;
END
$$;
