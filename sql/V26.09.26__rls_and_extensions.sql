-- ===========================================================================
-- V26.09.26 · RLS 行级安全策略 + 必要扩展启用
-- 依据：docs/architecture/04-数据模型设计.md §全局约定（RLS 行级安全）
--       docs/architecture/06-权限与安全设计.md
-- ===========================================================================
-- 设计原则：
--   1. 共享 Schema + tenant_id 行级隔离（对标 Plane / Jira Cloud 多租户模式）
--   2. 应用层通过 SET app.current_tenant 设置当前租户，SET LOCAL 会话级生效
--   3. FORCE ROW LEVEL SECURITY 确保表 owner 也受策略约束（防 DBA 绕过）
--   4. pg_stat_statements 用于慢查询巡检（架构文档 §17 性能基线）
-- ===========================================================================

-- ═══════════════════════════════════════════════════════════════════════
-- 第一部分：扩展启用
-- ═══════════════════════════════════════════════════════════════════════

-- pg_trgm：模糊搜索与相似度搜索（补充 tsvector 未覆盖的 LIKE %xxx% 场景）
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- pg_stat_statements：SQL 执行统计（慢查询巡检、P95 基线对比）
CREATE EXTENSION IF NOT EXISTS pg_stat_statements;

-- ═══════════════════════════════════════════════════════════════════════
-- 第二部分：租户上下文辅助函数（SECURITY DEFINER = 以定义者权限执行）
-- ═══════════════════════════════════════════════════════════════════════

-- set_tenant：在会话/事务级设置当前租户 ID
-- 用法：每次 HTTP 请求进入时由 middleware 调用 SELECT set_tenant(?)
CREATE OR REPLACE FUNCTION set_tenant(p_tenant BIGINT) RETURNS VOID AS $$
BEGIN
    PERFORM set_config('app.current_tenant', p_tenant::TEXT, false);
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- get_tenant：读取当前会话租户 ID（默认 1 = 系统默认租户）
CREATE OR REPLACE FUNCTION get_tenant() RETURNS BIGINT AS $$
BEGIN
    RETURN COALESCE(NULLIF(current_setting('app.current_tenant', true), '')::BIGINT, 1);
END;
$$ LANGUAGE plpgsql STABLE;

-- ═══════════════════════════════════════════════════════════════════════
-- 第三部分：批量启用 RLS + 创建隔离策略（幂等）
-- 说明：所有含 tenant_id 列的业务表均启用行级安全。
--       系统级表（如 flyway_schema_history 等无 tenant_id 的表）无需启用。
-- ═══════════════════════════════════════════════════════════════════════

DO $$
DECLARE
    t TEXT;
    tables TEXT[] := ARRAY[
        -- 系统/租户级
        'tenants', 'workspaces', 'workspace_members', 'users', 'roles', 'user_roles',
        'menus', 'tenant_members',
        -- 项目/工作项级
        'projects', 'project_members', 'project_sequences',
        'task', 'requirement', 'defect',
        'task_assignees', 'task_labels', 'task_modules', 'task_watchers',
        'task_relations', 'task_comments', 'task_activities', 'task_ext',
        'labels', 'modules', 'estimate_points',
        'sprints', 'sprint_snapshots', 'versions', 'version_delivery_snapshots',
        'states', 'state_transitions',
        -- 知识库
        'knowledge_spaces', 'knowledge_pages', 'knowledge_page_versions',
        'knowledge_page_relations', 'document_links', 'page_templates', 'page_shares',
        -- 自动化/仪表盘
        'automation_rules', 'rule_executions', 'automation_templates',
        'dashboards', 'dashboard_widgets', 'dashboard_snapshots', 'dashboard_templates',
        -- 通知
        'notifications', 'notification_deliveries', 'notification_preferences', 'notification_digests',
        -- 搜索
        'search_documents', 'search_history', 'search_bookmarks',
        -- 风险/指标
        'risk_rules', 'risk_alerts', 'metric_snapshots', 'metric_adjustments',
        -- 集成
        'webhooks', 'webhook_logs', 'workbench_configs', 'workbench_templates',
        'view_preferences', 'recent_items',
        -- 收件箱
        'intake_channels', 'intake_issues',
        -- 权限/审计
        'sso_providers', 'sso_links', 'api_tokens', 'audit_logs',
        'domain_events', 'invitations', 'user_preferences',
        'role_permissions', 'defect_extra', 'issue_dependencies'
    ];
BEGIN
    FOREACH t IN ARRAY tables LOOP
        -- 跳过不存在的表（幂等安全）
        IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = t AND table_schema = 'public') THEN
            -- 检查表是否含 tenant_id 列
            IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = t AND column_name = 'tenant_id' AND table_schema = 'public') THEN
                EXECUTE format('ALTER TABLE %I ENABLE ROW LEVEL SECURITY', t);
                EXECUTE format('ALTER TABLE %I FORCE ROW LEVEL SECURITY', t);

                -- 删除旧策略（幂等），重建
                EXECUTE format('DROP POLICY IF EXISTS %I ON %I', t || '_tenant_isolation', t);

                -- 创建隔离策略：仅允许访问当前租户的数据
                EXECUTE format(
                    'CREATE POLICY %I ON %I USING (tenant_id = get_tenant() OR tenant_id = 1)',
                    t || '_tenant_isolation', t
                );
            END IF;
        END IF;
    END LOOP;
END;
$$;
