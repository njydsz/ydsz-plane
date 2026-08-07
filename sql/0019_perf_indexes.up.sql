-- 0019_perf_indexes: 100 万工作项量级的覆盖索引与复合索引优化。
--
-- 基准假设：单项目 1M 工作项，覆盖互联网中型团队 3-5 年存量数据。
-- 设计原则（参考美团/字节内部 DBA 规范）：
--   1. 高频查询走覆盖索引（Index-Only Scan），避免回表
--   2. 复合索引列顺序按"等值过滤 → 排序 → 范围"原则
--   3. 避免冗余索引，每次写入额外维护成本可控（< 10%）
--   4. 局部索引（WHERE deleted_at IS NULL）减少索引体积

-- -----------------------------------------------------------------
-- 1. 主列表查询优化：project_list 默认排序(updated_at DESC)
--
-- 原索引 idx_issues_updated(project_id, updated_at DESC) 缺失 state/priority 等
-- 常用过滤列，导致大量回表。新增 covering index 包含常用展示列。
-- -----------------------------------------------------------------
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_issues_list_covering
    ON issues (project_id, updated_at DESC)
    INCLUDE (public_id, sequence_id, type_code, name, state_id,
             priority, point, target_date, progress, created_by)
    WHERE deleted_at IS NULL;

-- -----------------------------------------------------------------
-- 2. 按状态分桶的覆盖索引：工作台 / 看板常用场景
--    查询模式：WHERE project_id=? AND state_id=? ORDER BY sort_order
-- -----------------------------------------------------------------
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_issues_state_covering
    ON issues (project_id, state_id, sort_order)
    INCLUDE (id, sequence_id, type_code, name, priority, point, target_date, updated_at)
    WHERE deleted_at IS NULL;

-- -----------------------------------------------------------------
-- 3. 按优先级的覆盖索引：urgent/high 紧急工作项快速筛选
--    查询模式：WHERE project_id=? AND priority IN ('urgent','high') ORDER BY updated_at DESC
-- -----------------------------------------------------------------
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_issues_priority_covering
    ON issues (project_id, priority, updated_at DESC)
    INCLUDE (id, sequence_id, type_code, name, state_id, point, target_date)
    WHERE deleted_at IS NULL AND priority IN ('urgent', 'high');

-- -----------------------------------------------------------------
-- 4. 按目标日期的覆盖索引（优化日历视图/逾期清单）
--    查询模式：WHERE project_id=? AND target_date BETWEEN ? AND ?
-- -----------------------------------------------------------------
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_issues_target_date_covering
    ON issues (project_id, target_date)
    INCLUDE (id, sequence_id, type_code, name, state_id, priority, progress)
    WHERE deleted_at IS NULL AND target_date IS NOT NULL;

-- -----------------------------------------------------------------
-- 5. 按类型的覆盖索引（优化类型筛选：需求/任务/缺陷）
-- -----------------------------------------------------------------
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_issues_type_covering
    ON issues (project_id, type_code, created_at DESC)
    INCLUDE (id, sequence_id, name, state_id, priority, point, target_date)
    WHERE deleted_at IS NULL;

-- -----------------------------------------------------------------
-- 6. 指派给我的工作项：通过 issue_assignees 的反查能命中
--    原索引已有 idx_issue_assignees_user(user_id)，但无 covering。
-- -----------------------------------------------------------------
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_issue_assignees_covering
    ON issue_assignees (user_id)
    INCLUDE (issue_id, assigned_at);

-- -----------------------------------------------------------------
-- 7. 活动流查询优化：issue_activities 按 issue + 时间
--    工作台 / 详情页活动流常用，覆盖列包含展示所需全部字段。
-- -----------------------------------------------------------------
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_activities_issue_covering
    ON issue_activities (issue_id, created_at DESC)
    INCLUDE (verb, field, old_value, new_value, actor_id, actor_email, actor_name)
    WHERE true;  -- 活动记录不软删除
