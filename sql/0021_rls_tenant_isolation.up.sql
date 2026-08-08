-- =============================================================
-- 迁移 0021: 启用 PostgreSQL 原生 RLS（Row Level Security）作为纵深防御
-- 
-- 设计说明：
--   项目主路径使用应用层租户隔离（set_config('app.workspace_id', ...) + WHERE 条件），
--   RLS 作为第二道防线防止遗漏租户过滤的查询泄露数据。
--   
--   RLS 策略: 每个表的 SELECT/INSERT/UPDATE/DELETE 均通过 current_setting('app.workspace_id')
--   与表中 workspace_id 列做等值匹配。
--
-- 参考标准: 等保三级、OWASP Database Security Cheat Sheet
-- =============================================================

-- 启用 RLS 的表列表
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
        -- 检查表是否存在（避免迁移失败）
        IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = tbl) THEN
            EXECUTE format('ALTER TABLE %I ENABLE ROW LEVEL SECURITY', tbl);
            
            -- 如果不存在策略，创建默认策略
            IF NOT EXISTS (
                SELECT 1 FROM pg_policies 
                WHERE schemaname = 'public' AND tablename = tbl AND policyname = 'tenant_isolation'
            ) THEN
                EXECUTE format(
                    'CREATE POLICY tenant_isolation ON %I USING (workspace_id::text = current_setting(''app.workspace_id'', true))',
                    tbl
                );
            END IF;
        END IF;
    END LOOP;
END
$$;

-- 为没有 workspace_id 的表创建特殊策略（如 users 表通过 workspace_members 关联）
-- users 表不启用 RLS（全局表，通过 workspace_members 控制访问）

-- 迁移元数据记录
INSERT INTO ydsz_dump_state (version, applied_at) VALUES (21, now())
ON CONFLICT (version) DO UPDATE SET applied_at = now();
