-- ===========================================================================
-- V26.10.01 · RLS 写保护修复 — 为所有已启用 RLS 的表补 WITH CHECK
--
-- 安全漏洞：原 V26.09.26 迁移仅添加了 USING（读保护），未添加 WITH CHECK
-- （写保护），导致攻击者可以跨租户写入数据。
--
-- 修复：为所有含 tenant_id 的业务表追加 WITH CHECK 条件，确保
-- INSERT/UPDATE 的 tenant_id 必须匹配当前会话租户。
-- ===========================================================================

DO $$
DECLARE
    t TEXT;
    tables TEXT[] := ARRAY[
        'tenants', 'workspaces', 'workspace_members', 'users', 'roles', 'user_roles',
        'menus', 'tenant_members',
        'projects', 'project_members', 'project_sequences',
        'task', 'requirement', 'defect',
        'task_assignees', 'task_labels', 'task_modules', 'task_watchers',
        'task_relations', 'task_comments', 'task_activities', 'task_ext',
        'labels', 'modules', 'estimate_points',
        'sprints', 'sprint_snapshots', 'versions', 'version_delivery_snapshots',
        'states', 'state_transitions',
        'knowledge_spaces', 'knowledge_pages', 'knowledge_page_versions',
        'knowledge_page_relations', 'document_links', 'page_templates', 'page_shares',
        'automation_rules', 'rule_executions', 'automation_templates',
        'dashboards', 'dashboard_widgets', 'dashboard_snapshots', 'dashboard_templates',
        'notifications', 'notification_deliveries', 'notification_preferences', 'notification_digests',
        'search_documents', 'search_history', 'search_bookmarks',
        'risk_rules', 'risk_alerts', 'metric_snapshots', 'metric_adjustments',
        'webhooks', 'webhook_logs', 'workbench_configs', 'workbench_templates',
        'view_preferences', 'recent_items',
        'intake_channels', 'intake_issues',
        'sso_links', 'api_tokens', 'audit_logs',
        'domain_events', 'invitations', 'user_preferences',
        'role_permissions', 'defect_extra', 'issue_dependencies'
    ];
    pol_name TEXT;
BEGIN
    FOREACH t IN ARRAY tables LOOP
        -- 仅处理存在且已启用 RLS + 含 tenant_id 的表
        IF EXISTS (
            SELECT 1 FROM information_schema.tables
            WHERE table_name = t AND table_schema = 'public'
        ) AND EXISTS (
            SELECT 1 FROM information_schema.columns
            WHERE table_name = t AND column_name = 'tenant_id' AND table_schema = 'public'
        ) THEN
            pol_name := t || '_tenant_isolation';

            -- 检查策略是否存在
            IF EXISTS (
                SELECT 1 FROM pg_policies
                WHERE tablename = t AND policyname = pol_name
            ) THEN
                -- 删除旧策略（仅有 USING）
                EXECUTE format('DROP POLICY %I ON %I', pol_name, t);

                -- 重建：同时带 USING + WITH CHECK
                EXECUTE format(
                    'CREATE POLICY %I ON %I '
                    'USING (tenant_id = get_tenant() OR tenant_id = 1) '
                    'WITH CHECK (tenant_id = get_tenant() OR tenant_id = 1)',
                    pol_name, t
                );
            END IF;
        END IF;
    END LOOP;
END;
$$;
