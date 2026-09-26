-- ===========================================================================
-- V26.10.02 · 高频查询性能索引补充
--
-- 修复列表分页全表 filesort、看板查询回表过滤租户、通知收件箱索引 WHERE
-- 条件不匹配、JSONB 高频字段缺 GIN 索引。
--
-- 依据：docs/optimization/S19-数据库表设计分析报告 第 2 节
-- ===========================================================================

-- 安全幂等：所有索引使用 IF NOT EXISTS，可重复执行

-- ═══════════════════════════════════════════════════════════════════════
-- 1. 核心业务表 — 列表分页 + 时间倒序（修复 filesort）
-- ═══════════════════════════════════════════════════════════════════════

CREATE INDEX IF NOT EXISTS idx_task_tenant_created_at
    ON task (tenant_id, created_at DESC) WHERE NOT deleted;

CREATE INDEX IF NOT EXISTS idx_requirement_tenant_created_at
    ON requirement (tenant_id, created_at DESC) WHERE NOT deleted;

CREATE INDEX IF NOT EXISTS idx_defect_tenant_created_at
    ON defect (tenant_id, created_at DESC) WHERE NOT deleted;

-- 项目内时间倒序分页
CREATE INDEX IF NOT EXISTS idx_task_project_created_at
    ON task (tenant_id, project_id, created_at DESC) WHERE NOT deleted;

CREATE INDEX IF NOT EXISTS idx_requirement_project_created_at
    ON requirement (tenant_id, project_id, created_at DESC) WHERE NOT deleted;

CREATE INDEX IF NOT EXISTS idx_defect_project_created_at
    ON defect (tenant_id, project_id, created_at DESC) WHERE NOT deleted;

-- ═══════════════════════════════════════════════════════════════════════
-- 2. 看板视图 — 含 tenant_id 前缀
-- ═══════════════════════════════════════════════════════════════════════

CREATE INDEX IF NOT EXISTS idx_task_board_tenant
    ON task (tenant_id, project_id, state_id, sort_order) WHERE NOT deleted;

CREATE INDEX IF NOT EXISTS idx_requirement_board_tenant
    ON requirement (tenant_id, project_id, state_id, sort_order) WHERE NOT deleted;

CREATE INDEX IF NOT EXISTS idx_defect_board_tenant
    ON defect (tenant_id, project_id, state_id, sort_order) WHERE NOT deleted;

-- ═══════════════════════════════════════════════════════════════════════
-- 3. 迭代查询
-- ═══════════════════════════════════════════════════════════════════════

CREATE INDEX IF NOT EXISTS idx_sprints_project_dates
    ON sprints (tenant_id, project_id, start_date, end_date) WHERE NOT deleted;

-- ═══════════════════════════════════════════════════════════════════════
-- 4. 审计日志 — 按时间范围查询
-- ═══════════════════════════════════════════════════════════════════════

CREATE INDEX IF NOT EXISTS idx_audit_logs_tenant_time
    ON audit_logs (tenant_id, created_at DESC) WHERE NOT deleted;

CREATE INDEX IF NOT EXISTS idx_audit_logs_action
    ON audit_logs (tenant_id, action, created_at DESC) WHERE NOT deleted;

-- ═══════════════════════════════════════════════════════════════════════
-- 5. 工时报表 — 按花费日期日期范围查询
-- ═══════════════════════════════════════════════════════════════════════

CREATE INDEX IF NOT EXISTS idx_task_timelogs_date
    ON task_timelogs (tenant_id, project_id, spent_date) WHERE NOT deleted;

CREATE INDEX IF NOT EXISTS idx_requirement_timelogs_date
    ON requirement_timelogs (tenant_id, project_id, spent_date) WHERE NOT deleted;

CREATE INDEX IF NOT EXISTS idx_defect_timelogs_date
    ON defect_timelogs (tenant_id, project_id, spent_date) WHERE NOT deleted;

-- ═══════════════════════════════════════════════════════════════════════
-- 6. 通知收件箱 — 重建索引（原索引 WHERE 条件与键列不匹配）
-- ═══════════════════════════════════════════════════════════════════════

-- 收件箱主查询：未删除 + 未归档，按时间倒序
DROP INDEX IF EXISTS idx_notifications_recipient;
CREATE INDEX IF NOT EXISTS idx_notifications_inbox
    ON notifications (tenant_id, recipient_id, created_at DESC)
    WHERE NOT deleted AND NOT is_archived;

-- 未读计数查询
CREATE INDEX IF NOT EXISTS idx_notifications_unread
    ON notifications (tenant_id, recipient_id)
    WHERE NOT deleted AND NOT is_archived AND NOT is_read;

-- ═══════════════════════════════════════════════════════════════════════
-- 7. JSONB GIN 索引
-- ═══════════════════════════════════════════════════════════════════════

-- projects.modules：功能模块开关（高频内部查询）
CREATE INDEX IF NOT EXISTS idx_projects_modules
    ON projects USING gin (modules jsonb_path_ops)
    WHERE modules IS NOT NULL AND NOT deleted;

-- automation_rules.conditions：条件匹配
CREATE INDEX IF NOT EXISTS idx_automation_rules_conditions
    ON automation_rules USING gin (conditions jsonb_path_ops);

-- tenants.brand_config：品牌配置内部查询
CREATE INDEX IF NOT EXISTS idx_tenants_brand_config
    ON tenants USING gin (brand_config)
    WHERE brand_config IS NOT NULL;

-- ═══════════════════════════════════════════════════════════════════════
-- 8. 清理冗余索引
-- ═══════════════════════════════════════════════════════════════════════

DROP INDEX IF EXISTS idx_notif_subs_user;
