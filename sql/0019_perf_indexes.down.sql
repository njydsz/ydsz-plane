-- 0019_perf_indexes 回滚：删除性能优化索引。

DROP INDEX IF EXISTS idx_issues_list_covering;
DROP INDEX IF EXISTS idx_issues_state_covering;
DROP INDEX IF EXISTS idx_issues_priority_covering;
DROP INDEX IF EXISTS idx_issues_target_date_covering;
DROP INDEX IF EXISTS idx_issues_type_covering;
DROP INDEX IF EXISTS idx_issue_assignees_covering;
DROP INDEX IF EXISTS idx_activities_issue_covering;
