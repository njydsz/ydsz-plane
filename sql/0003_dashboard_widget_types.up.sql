-- 0003: 仪表盘布局增强 — 扩展 widget 类型约束
--
-- 变更内容:
--   1. 新增 widget 类型: version_burndown / module_distribution / dora / project_compare
--   2. dashboard_widgets 已具备 grid_x/grid_y/grid_w/grid_h 字段（0-11 列网格），无需加列

BEGIN;

-- 重建 CHECK 约束，纳入全部受支持的 widget 类型。
ALTER TABLE dashboard_widgets
    DROP CONSTRAINT IF EXISTS dashboard_widgets_widget_type_check;

ALTER TABLE dashboard_widgets
    ADD CONSTRAINT dashboard_widgets_widget_type_check
    CHECK (widget_type = ANY (ARRAY[
        'progress_overview', 'burndown', 'velocity', 'priority_split',
        'state_distribution', 'overdue_list', 'blocked_list', 'risk_alert',
        'recent_activity', 'team_workload', 'version_burndown',
        'module_distribution', 'dora', 'project_compare'
    ]::text[]));

COMMIT;
