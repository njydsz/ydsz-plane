
-- ============================================================================
-- 自动化域
-- ============================================================================

-- ============================================================
-- 表: automation_rules — 自动化规则表
-- ============================================================
COMMENT ON TABLE public.automation_rules IS '[自动化域]工作项自动化规则（DSL 定义 trigger+condition+action，支持多类型触发器）';

COMMENT ON COLUMN public.automation_rules.id IS '自增主键';
COMMENT ON COLUMN public.automation_rules.workspace_id IS '关联 workspaces.id（租户隔离列）';
COMMENT ON COLUMN public.automation_rules.project_id IS '关联 projects.id（项目级规则为空表示空间级）';
COMMENT ON COLUMN public.automation_rules.name IS '规则名称';
COMMENT ON COLUMN public.automation_rules.description IS '规则描述';
COMMENT ON COLUMN public.automation_rules.dsl IS '规则 DSL JSON（trigger / conditions / actions 三段式结构）';
COMMENT ON COLUMN public.automation_rules.trigger_type IS '触发器类型: issue.created / issue.updated / issue.status_changed / version.released / scheduled 等';
COMMENT ON COLUMN public.automation_rules.action_count IS '规则包含的动作数量';
COMMENT ON COLUMN public.automation_rules.status IS '规则状态: draft(草稿) / active(启用) / disabled(禁用) / error(执行出错)';
COMMENT ON COLUMN public.automation_rules.created_by IS '关联 users.id（创建者）';
COMMENT ON COLUMN public.automation_rules.last_run_at IS '最近一次执行时间（含时区）';
COMMENT ON COLUMN public.automation_rules.last_error IS '最近一次错误信息';
COMMENT ON COLUMN public.automation_rules.consecutive_failures IS '连续失败次数（达到阈值自动置为 error）';
COMMENT ON COLUMN public.automation_rules.execution_count IS '累计执行次数';
COMMENT ON COLUMN public.automation_rules.sort_order IS '排序权重（越小越靠前）';
COMMENT ON COLUMN public.automation_rules.created_at IS '创建时间（含时区）';
COMMENT ON COLUMN public.automation_rules.updated_at IS '最后更新时间（含时区）';

-- ============================================================
-- 表: automation_templates — 自动化模板表
-- ============================================================
COMMENT ON TABLE public.automation_templates IS '[自动化域]预置自动化规则模板（用户可一键启用，快速配置常见自动化场景）';

COMMENT ON COLUMN public.automation_templates.id IS '自增主键';
COMMENT ON COLUMN public.automation_templates.name IS '模板名称';
COMMENT ON COLUMN public.automation_templates.slug IS '模板标识符（系统内唯一）';
COMMENT ON COLUMN public.automation_templates.description IS '模板说明';
COMMENT ON COLUMN public.automation_templates.category IS '模板分类: efficiency(效率) / quality(质量) / notification(通知)';
COMMENT ON COLUMN public.automation_templates.dsl_template IS '模板 DSL JSON（预置的 trigger+condition+action 结构）';
COMMENT ON COLUMN public.automation_templates.icon IS '模板图标名（Lucide）';
COMMENT ON COLUMN public.automation_templates.sort_order IS '排序权重';
COMMENT ON COLUMN public.automation_templates.is_recommended IS '是否推荐: true=在首页推荐展示 / false=普通';
COMMENT ON COLUMN public.automation_templates.created_at IS '创建时间（含时区）';

-- ============================================================
-- 表: rule_executions — 规则执行日志表
-- ============================================================
COMMENT ON TABLE public.rule_executions IS '[自动化域]自动化规则执行日志（记录每次触发的匹配、执行和耗时信息）';

COMMENT ON COLUMN public.rule_executions.id IS '自增主键';
COMMENT ON COLUMN public.rule_executions.workspace_id IS '关联 workspaces.id（租户隔离列）';
COMMENT ON COLUMN public.rule_executions.project_id IS '关联 projects.id（项目级规则执行）';
COMMENT ON COLUMN public.rule_executions.rule_id IS '关联 automation_rules.id（执行规则）';
COMMENT ON COLUMN public.rule_executions.trigger_event_id IS '关联 domain_events.id（触发事件 ID）';
COMMENT ON COLUMN public.rule_executions.status IS '执行状态: matched(已匹配但未执行) / skipped(跳过条件不满足) / success(成功) / failed(失败) / dry_run(试运行)';
COMMENT ON COLUMN public.rule_executions.duration_ms IS '执行耗时（毫秒）';
COMMENT ON COLUMN public.rule_executions.error_message IS '执行失败时的错误信息';
COMMENT ON COLUMN public.rule_executions.context_json IS '执行上下文 JSON（当时的 issue 快照等）';
COMMENT ON COLUMN public.rule_executions.trigger_depth IS '触发深度（防止递归触发，0=直接触发）';
COMMENT ON COLUMN public.rule_executions.via_automation IS '是否由其他自动化规则间接触发';
COMMENT ON COLUMN public.rule_executions.created_at IS '创建时间（含时区）';

-- ============================================================================
-- 仪表盘域
-- ============================================================================

-- ============================================================
-- 表: dashboard_widgets — 仪表盘组件表
-- ============================================================
COMMENT ON TABLE public.dashboard_widgets IS '[仪表盘域]仪表盘组件实例（项目级可配置的图表/列表/数据卡组件，支持布局定位）';

COMMENT ON COLUMN public.dashboard_widgets.id IS '自增主键';
COMMENT ON COLUMN public.dashboard_widgets.project_id IS '关联 projects.id（所属项目）';
COMMENT ON COLUMN public.dashboard_widgets.widget_type IS '组件类型: progress_overview(进度概览) / burndown(燃尽图) / velocity(速率图) / priority_split(优先级分布) / state_distribution(状态分布) / overdue_list(逾期列表) / blocked_list(阻塞列表) / risk_alert(风险告警) / recent_activity(近期活动) / team_workload(团队负载)';
COMMENT ON COLUMN public.dashboard_widgets.title IS '组件标题';
COMMENT ON COLUMN public.dashboard_widgets.grid_x IS '网格 X 坐标（列位置）';
COMMENT ON COLUMN public.dashboard_widgets.grid_y IS '网格 Y 坐标（行位置）';
COMMENT ON COLUMN public.dashboard_widgets.grid_w IS '网格宽度（占几列，默认 4）';
COMMENT ON COLUMN public.dashboard_widgets.grid_h IS '网格高度（占几行，默认 3）';
COMMENT ON COLUMN public.dashboard_widgets.config IS '组件配置 JSON（数据源、过滤条件、展示参数）';
COMMENT ON COLUMN public.dashboard_widgets.is_visible IS '是否可见: true=显示 / false=隐藏';
COMMENT ON COLUMN public.dashboard_widgets.sort_order IS '排序权重';
COMMENT ON COLUMN public.dashboard_widgets.user_id IS '关联 users.id（NULL 表示项目级组件，非空表示个人组件）';
COMMENT ON COLUMN public.dashboard_widgets.created_at IS '创建时间（含时区）';
COMMENT ON COLUMN public.dashboard_widgets.updated_at IS '最后更新时间（含时区）';

-- ============================================================
-- 表: dashboard_templates — 仪表盘模板表
-- ============================================================
COMMENT ON TABLE public.dashboard_templates IS '[仪表盘域]预置仪表盘布局模板（一键初始化项目仪表盘，如项目概览、质量看板等）';

COMMENT ON COLUMN public.dashboard_templates.id IS '自增主键';
COMMENT ON COLUMN public.dashboard_templates.name IS '模板名称';
COMMENT ON COLUMN public.dashboard_templates.slug IS '模板标识符（系统内唯一）';
COMMENT ON COLUMN public.dashboard_templates.description IS '模板说明';
COMMENT ON COLUMN public.dashboard_templates.layout IS '布局配置 JSON（widgets 数组的默认位置与尺寸）';
COMMENT ON COLUMN public.dashboard_templates.icon IS '模板图标名';
COMMENT ON COLUMN public.dashboard_templates.category IS '模板分类: agile(敏捷) / pmo(项目管理) / quality(质量)';
COMMENT ON COLUMN public.dashboard_templates.is_default IS '是否为默认模板: true=创建项目时自动应用';
COMMENT ON COLUMN public.dashboard_templates.sort_order IS '排序权重';
COMMENT ON COLUMN public.dashboard_templates.created_at IS '创建时间（含时区）';

-- ============================================================
-- 表: dashboard_snapshots — 仪表盘快照表
-- ============================================================
COMMENT ON TABLE public.dashboard_snapshots IS '[仪表盘域]仪表盘组件数据快照（定时刷新缓存，避免实时查询开销）';

COMMENT ON COLUMN public.dashboard_snapshots.id IS '自增主键';
COMMENT ON COLUMN public.dashboard_snapshots.project_id IS '关联 projects.id（所属项目）';
COMMENT ON COLUMN public.dashboard_widgets.widget_type IS '组件类型（同 dashboard_widgets.widget_type）';
COMMENT ON COLUMN public.dashboard_snapshots.data IS '快照数据 JSON（组件渲染所需的全部数据）';
COMMENT ON COLUMN public.dashboard_snapshots.refreshed_at IS '最后刷新时间（含时区）';

-- ============================================================================
-- 功能区
-- ============================================================================

-- ============================================================
-- 表: workbench_configs — 工作台配置表
-- ============================================================
COMMENT ON TABLE public.workbench_configs IS '[功能区]用户工作台个性化布局配置（每人/每项目可有独立工作台布局）';

COMMENT ON COLUMN public.workbench_configs.id IS '自增主键';
COMMENT ON COLUMN public.workbench_configs.workspace_id IS '关联 workspaces.id（租户隔离列）';
COMMENT ON COLUMN public.workbench_configs.project_id IS '关联 projects.id（NULL 表示空间级工作台）';
COMMENT ON COLUMN public.workbench_configs.user_id IS '关联 users.id（配置所属用户）';
COMMENT ON COLUMN public.workbench_configs.layout IS '工作台布局 JSON（组件排列、尺寸等）';
COMMENT ON COLUMN public.workbench_configs.widget_states IS '组件状态 JSON（折叠/展开/过滤条件等）';
COMMENT ON COLUMN public.workbench_configs.focus_enabled IS '是否启用专注模式: true=仅显示待办和专注计时器';
COMMENT ON COLUMN public.workbench_configs.updated_at IS '最后更新时间（含时区）';

-- ============================================================
-- 表: workbench_templates — 工作台模板表
-- ============================================================
COMMENT ON TABLE public.workbench_templates IS '[功能区]预置工作台布局模板（如敏捷开发、项目监控、个人专注等模式一键切换）';

COMMENT ON COLUMN public.workbench_templates.id IS '自增主键';
COMMENT ON COLUMN public.workbench_templates.name IS '模板名称';
COMMENT ON COLUMN public.workbench_templates.slug IS '模板标识符（系统内唯一）';
COMMENT ON COLUMN public.workbench_templates.description IS '模板说明';
COMMENT ON COLUMN public.workbench_templates.layout IS '布局 JSON（预置组件排列方案）';
COMMENT ON COLUMN public.workbench_templates.icon IS '模板图标名';
COMMENT ON COLUMN public.workbench_templates.is_default IS '是否为默认模板: true=新用户默认使用';
COMMENT ON COLUMN public.workbench_templates.sort_order IS '排序权重';
COMMENT ON COLUMN public.workbench_templates.created_at IS '创建时间（含时区）';

-- ============================================================
-- 表: view_preferences — 视图偏好表
-- ============================================================
COMMENT ON TABLE public.view_preferences IS '[功能区]视图展示偏好（如列表/看板的列定义、筛选条件、排序方式）';

COMMENT ON COLUMN public.view_preferences.id IS '自增主键';
COMMENT ON COLUMN public.view_preferences.workspace_id IS '关联 workspaces.id（租户隔离列）';
COMMENT ON COLUMN public.view_preferences.project_id IS '关联 projects.id（所属项目）';
COMMENT ON COLUMN public.view_preferences.user_id IS '关联 users.id（偏好所属用户）';
COMMENT ON COLUMN public.view_preferences.view_type IS '视图类型: list(列表) / kanban(看板) / calendar(日历) / spreadsheet(表格) / gantt(甘特图)';
COMMENT ON COLUMN public.view_preferences.layout IS '布局模式: list / kanban / calendar / spreadsheet / gantt 等';
COMMENT ON COLUMN public.view_preferences.columns IS '已配置列的 JSON 数组（列宽、可见性、顺序）';
COMMENT ON COLUMN public.view_preferences.filters IS '视图过滤条件 JSON';
COMMENT ON COLUMN public.view_preferences.sort IS '排序规则 JSON（字段、方向）';
COMMENT ON COLUMN public.view_preferences.extra IS '额外视图参数 JSON（视图特有配置）';
COMMENT ON COLUMN public.view_preferences.created_at IS '创建时间（含时区）';
COMMENT ON COLUMN public.view_preferences.updated_at IS '最后更新时间（含时区）';

-- ============================================================
-- 表: recent_items — 最近访问记录表
-- ============================================================
COMMENT ON TABLE public.recent_items IS '[功能区]用户最近访问记录（工作台"最近访问"功能的数据来源）';

COMMENT ON COLUMN public.recent_items.id IS '自增主键';
COMMENT ON COLUMN public.recent_items.workspace_id IS '关联 workspaces.id（租户隔离列）';
COMMENT ON COLUMN public.recent_items.user_id IS '关联 users.id（记录所属用户）';
COMMENT ON COLUMN public.recent_items.item_type IS '访问对象类型: project / issue / sprint / version';
COMMENT ON COLUMN public.recent_items.item_id IS '访问对象的 ID';
COMMENT ON COLUMN public.recent_items.project_id IS '关联 projects.id（便于按项目筛选）';
COMMENT ON COLUMN public.recent_items.title IS '对象标题（冗余存储，避免关联查询）';
COMMENT ON COLUMN public.recent_items.identifier IS '对象标识符（如 YD-123，冗余存储）';
COMMENT ON COLUMN public.recent_items.accessed_at IS '访问时间（含时区，触发器自动更新）';

-- ============================================================
-- 表: search_bookmarks — 搜索书签表
-- ============================================================
COMMENT ON TABLE public.search_bookmarks IS '[功能区]用户保存的搜索书签（常用搜索条件的快捷入口，可共享给团队）';

COMMENT ON COLUMN public.search_bookmarks.id IS '自增主键';
COMMENT ON COLUMN public.search_bookmarks.workspace_id IS '关联 workspaces.id（租户隔离列）';
COMMENT ON COLUMN public.search_bookmarks.project_id IS '关联 projects.id（项目级书签，可为空）';
COMMENT ON COLUMN public.search_bookmarks.user_id IS '关联 users.id（书签所属用户）';
COMMENT ON COLUMN public.search_bookmarks.name IS '书签名称';
COMMENT ON COLUMN public.search_bookmarks.query IS '搜索关键词';
COMMENT ON COLUMN public.search_bookmarks.filters IS '搜索过滤条件 JSON';
COMMENT ON COLUMN public.search_bookmarks.is_shared IS '是否共享: true=项目成员可见 / false=仅个人可见';
COMMENT ON COLUMN public.search_bookmarks.sort_order IS '排序权重';
COMMENT ON COLUMN public.search_bookmarks.created_at IS '创建时间（含时区）';
COMMENT ON COLUMN public.search_bookmarks.updated_at IS '最后更新时间（含时区）';
COMMENT ON COLUMN public.search_bookmarks.deleted_at IS '软删除时间（含时区），NULL 表示未删除';

-- ============================================================
-- 表: search_documents — 搜索文档索引表
-- ============================================================
COMMENT ON TABLE public.search_documents IS '[功能区]全文检索文档索引（触发器自动从 issues/sprints/versions 同步，供 PG 全文检索降级使用）';

COMMENT ON COLUMN public.search_documents.id IS '自增主键';
COMMENT ON COLUMN public.search_documents.workspace_id IS '关联 workspaces.id（租户隔离列）';
COMMENT ON COLUMN public.search_documents.project_id IS '关联 projects.id（所属项目）';
COMMENT ON COLUMN public.search_documents.doc_type IS '文档类型: issue / sprint / version';
COMMENT ON COLUMN public.search_documents.doc_id IS '原始表中的 ID（与 doc_type 组成唯一索引）';
COMMENT ON COLUMN public.search_documents.title IS '文档标题（工作项名/迭代名/版本名）';
COMMENT ON COLUMN public.search_documents.identifier IS '文档标识符（如 YD-123 或版本号）';
COMMENT ON COLUMN public.search_documents.content IS '文档正文内容（纯文本，用于全文检索）';
COMMENT ON COLUMN public.search_documents.search_tsv IS '全文检索 tsvector（simple 配置，由触发器维护）';
COMMENT ON COLUMN public.search_documents.metadata IS '元数据 JSON（type_code、state_id、priority 等可筛选字段）';
COMMENT ON COLUMN public.search_documents.updated_at IS '最后更新时间（含时区）';

-- ============================================================
-- 表: search_history — 搜索历史表
-- ============================================================
COMMENT ON TABLE public.search_history IS '[功能区]用户搜索历史记录（用于搜索建议和最近搜索展示）';

COMMENT ON COLUMN public.search_history.id IS '自增主键';
COMMENT ON COLUMN public.search_history.workspace_id IS '关联 workspaces.id（租户隔离列）';
COMMENT ON COLUMN public.search_history.user_id IS '关联 users.id（搜索用户）';
COMMENT ON COLUMN public.search_history.query IS '搜索关键词原文';
COMMENT ON COLUMN public.search_history.filters IS '搜索当时使用的过滤条件 JSON';
COMMENT ON COLUMN public.search_history.result_count IS '搜索结果数量';
COMMENT ON COLUMN public.search_history.searched_at IS '搜索时间（含时区）';
