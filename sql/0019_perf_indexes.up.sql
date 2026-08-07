-- 0019_perf_indexes: 100 万工作项量级的覆盖索引与复合索引优化。
--
-- 基准假设：单项目 1M 工作项，覆盖互联网中型团队 3-5 年存量数据。
-- 设计原则（参考美团/字节内部 DBA 规范）：
--   1. 高频查询走覆盖索引（Index-Only Scan），避免回表
--   2. 复合索引列顺序按"等值过滤 → 排序 → 范围"原则
--   3. 避免冗余索引，每次写入额外维护成本可控（< 10%）
--   4. 局部索引（WHERE deleted_at IS NULL）减少索引体积
-- 注：迁移中移除 CONCURRENTLY（不能在事务内执行），生产环境可手动补加

CREATE INDEX IF NOT EXISTS idx_issues_list_covering
    ON issues (project_id, updated_at DESC)
    INCLUDE (public_id, sequence_id, type_code, name, state_id,
             priority, point, target_date, progress, created_by)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_issues_state_covering
    ON issues (project_id, state_id, sort_order)
    INCLUDE (id, sequence_id, type_code, name, priority, point, target_date, updated_at)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_issues_priority_covering
    ON issues (project_id, priority, updated_at DESC)
    INCLUDE (id, sequence_id, type_code, name, state_id, point, target_date)
    WHERE deleted_at IS NULL AND priority IN ('urgent', 'high');

CREATE INDEX IF NOT EXISTS idx_issues_target_date_covering
    ON issues (project_id, target_date)
    INCLUDE (id, sequence_id, type_code, name, state_id, priority, progress)
    WHERE deleted_at IS NULL AND target_date IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_issues_type_covering
    ON issues (project_id, type_code, created_at DESC)
    INCLUDE (id, sequence_id, name, state_id, priority, point, target_date)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_issue_assignees_covering
    ON issue_assignees (user_id)
    INCLUDE (issue_id, assigned_at);

CREATE INDEX IF NOT EXISTS idx_activities_issue_covering
    ON issue_activities (issue_id, created_at DESC)
    INCLUDE (verb, field, old_value, new_value, actor_id, actor_email, actor_name);
