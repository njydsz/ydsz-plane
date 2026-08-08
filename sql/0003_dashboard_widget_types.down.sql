-- 0003 回滚: 恢复原始 widget 类型约束（不含新增类型）

BEGIN;

ALTER TABLE dashboard_widgets
    DROP CONSTRAINT IF EXISTS dashboard_widgets_widget_type_check;

ALTER TABLE dashboard_widgets
    ADD CONSTRAINT dashboard_widgets_widget_type_check
    CHECK (widget_type = ANY (ARRAY[
        'progress_overview', 'burndown', 'velocity', 'priority_split',
        'state_distribution', 'overdue_list', 'blocked_list', 'risk_alert',
        'recent_activity', 'team_workload'
    ]::text[]));

COMMIT;
