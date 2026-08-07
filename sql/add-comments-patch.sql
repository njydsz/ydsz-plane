-- ============================================================================
-- 补丁名称: ydsz-plane 全库注释补丁 (add-comments-patch.sql)
-- 生成日期: 2026-08-08
-- 目标数据库: PostgreSQL 18, schema: public
-- 用途: 为 ydsz-plane 项目全部 58 张表及其字段、索引、触发器添加中文注释，
--       实现 100% 注释覆盖率。
--
-- 使用方法:
--   psql -U ydsz_app -d ydsz-plane -f add-comments-patch.sql
--
-- 注意事项:
--   1. 已存在注释的表/issue_comments, notification_preferences, notifications, versions
--      仅补充缺失字段的注释，不覆盖已有内容。
--   2. 所有 COMMENT 语句均为幂等操作，可重复执行。
--   3. 部分索引注释中包含 WHERE 条件含义说明。
-- ============================================================================

-- ============================================================================
-- 核心工作项域
-- ============================================================================

-- ============================================================
-- 表: issues — 核心工作项表
-- ============================================================
COMMENT ON TABLE public.issues IS '[核心工作项域]工作项主表（承载需求/任务/缺陷三类工作项，项目内 sequence_id 唯一标识如 YD-123，支持3级父子层级，乐观锁并发控制）';

COMMENT ON COLUMN public.issues.id IS '自增主键（内部使用，不对外暴露）';
COMMENT ON COLUMN public.issues.public_id IS '对外暴露的 UUID 主键，API 与前端使用此字段';
COMMENT ON COLUMN public.issues.workspace_id IS '关联 workspaces.id（租户隔离列，RLS 策略依据）';
COMMENT ON COLUMN public.issues.project_id IS '关联 projects.id（所属项目）';
COMMENT ON COLUMN public.issues.sequence_id IS '项目内自增序列号，配合 identifier 展示为 YD-123 格式';
COMMENT ON COLUMN public.issues.type_code IS '工作项类型: requirement(需求) / task(任务) / defect(缺陷)';
COMMENT ON COLUMN public.issues.parent_id IS '关联 issues.id（父工作项，支持层级结构）';
COMMENT ON COLUMN public.issues.depth IS '层级深度: 1=顶层 / 2=子项 / 3=孙项，最大 3';
COMMENT ON COLUMN public.issues.name IS '工作项标题';
COMMENT ON COLUMN public.issues.description_json IS '工作项描述的 TipTap JSON 格式（富文本编辑器原始数据，文档节点树结构）';
COMMENT ON COLUMN public.issues.description_html IS '工作项描述的 HTML 渲染结果';
COMMENT ON COLUMN public.issues.description_stripped IS '纯文本摘要（去除富文本标记，供全文检索使用）';
COMMENT ON COLUMN public.issues.state_id IS '关联 states.id（当前状态）';
COMMENT ON COLUMN public.issues.priority IS '优先级: urgent(紧急) / high(高) / medium(中) / low(低) / none(无)';
COMMENT ON COLUMN public.issues.severity IS '严重程度（缺陷专有）: 1=致命 / 2=严重 / 3=一般 / 4=轻微 / 5=建议';
COMMENT ON COLUMN public.issues.found_phase IS '发现阶段（缺陷专有）: unit(单元测试) / integration(集成测试) / uat(验收测试) / production(生产环境) / customer(客户反馈)';
COMMENT ON COLUMN public.issues.root_cause_category IS '根因分类（缺陷专有）: requirement(需求问题) / technical(技术问题) / environment(环境问题) / data(数据问题)';
COMMENT ON COLUMN public.issues.verifier_id IS '关联 users.id（缺陷验证人）';
COMMENT ON COLUMN public.issues.environment IS '环境信息 JSON（缺陷专有），结构: {os, browser, version, ...}';
COMMENT ON COLUMN public.issues.reproduce_steps IS '复现步骤 JSON（缺陷专有），结构: {steps: [], expected: "", actual: ""}';
COMMENT ON COLUMN public.issues.category IS '工作项分类（任务专有）: frontend / backend / qa / doc / design / other';
COMMENT ON COLUMN public.issues.actual_effort IS '实际花费工时（小时，NUMERIC(8,2)）';
COMMENT ON COLUMN public.issues.remaining_effort IS '剩余预估工时（小时，NUMERIC(8,2)）';
COMMENT ON COLUMN public.issues.delay_reason IS '延期原因（任务专有）: requirement_change(需求变更) / resource(资源不足) / blocked(被阻塞) / other(其他)';
COMMENT ON COLUMN public.issues.source IS '需求来源（需求专有）: customer(客户) / internal(内部) / competitor(竞品)';
COMMENT ON COLUMN public.issues.point IS '故事点（0-12，敏捷估算用）';
COMMENT ON COLUMN public.issues.sprint_id IS '关联 sprints.id（所属迭代）';
COMMENT ON COLUMN public.issues.progress IS '完成进度百分比（0-100）';
COMMENT ON COLUMN public.issues.start_date IS '实际开始日期';
COMMENT ON COLUMN public.issues.target_date IS '目标完成日期';
COMMENT ON COLUMN public.issues.completed_at IS '实际完成时间（含时区）';
COMMENT ON COLUMN public.issues.is_draft IS '是否为草稿: true=草稿 / false=已发布';
COMMENT ON COLUMN public.issues.sort_order IS '看板列内排序权重（默认 65535，越小越靠前）';
COMMENT ON COLUMN public.issues.created_by IS '关联 users.id（创建者）';
COMMENT ON COLUMN public.issues.created_at IS '创建时间（含时区）';
COMMENT ON COLUMN public.issues.updated_at IS '最后更新时间（含时区，由触发器自动维护）';
COMMENT ON COLUMN public.issues.deleted_at IS '软删除时间（含时区），NULL 表示未删除';
COMMENT ON COLUMN public.issues.version IS '乐观锁版本号（每次更新自增，冲突返回 409）';
COMMENT ON COLUMN public.issues.found_version_id IS '关联 versions.id（发现缺陷时的版本）';
COMMENT ON COLUMN public.issues.fix_version_id IS '关联 versions.id（计划修复的版本）';
COMMENT ON COLUMN public.issues.release_version_id IS '关联 versions.id（首次发布的版本）';
COMMENT ON COLUMN public.issues.search_tsv IS '全文检索向量（自动生成，simple 配置，供 ES 降级使用）';

-- ============================================================
-- 表: labels — 标签表
-- ============================================================
COMMENT ON TABLE public.labels IS '[核心工作项域]项目级标签（用于工作项分类、筛选与可视化标识）';

COMMENT ON COLUMN public.labels.id IS '自增主键';
COMMENT ON COLUMN public.labels.workspace_id IS '关联 workspaces.id（租户隔离列）';
COMMENT ON COLUMN public.labels.project_id IS '关联 projects.id（所属项目）';
COMMENT ON COLUMN public.labels.name IS '标签名称（项目内唯一展示名）';
COMMENT ON COLUMN public.labels.color IS '标签颜色（十六进制，如 #8DA2C2）';
COMMENT ON COLUMN public.labels.description IS '标签说明文本';
COMMENT ON COLUMN public.labels.created_by IS '关联 users.id（创建者）';
COMMENT ON COLUMN public.labels.created_at IS '创建时间（含时区）';
COMMENT ON COLUMN public.labels.updated_at IS '最后更新时间（含时区）';
COMMENT ON COLUMN public.labels.deleted_at IS '软删除时间（含时区），NULL 表示未删除';

-- ============================================================
-- 表: modules — 模块/组件表
-- ============================================================
COMMENT ON TABLE public.modules IS '[核心工作项域]项目模块/组件分解结构（将项目划分为功能模块，每模块可关联工作项）';

COMMENT ON COLUMN public.modules.id IS '自增主键';
COMMENT ON COLUMN public.modules.workspace_id IS '关联 workspaces.id（租户隔离列）';
COMMENT ON COLUMN public.modules.project_id IS '关联 projects.id（所属项目）';
COMMENT ON COLUMN public.modules.name IS '模块名称';
COMMENT ON COLUMN public.modules.description IS '模块描述';
COMMENT ON COLUMN public.modules.lead_id IS '关联 users.id（模块负责人）';
COMMENT ON COLUMN public.modules.status IS '模块状态: active(进行中) / completed(已完成) / cancelled(已取消)';
COMMENT ON COLUMN public.modules.start_date IS '模块开始日期';
COMMENT ON COLUMN public.modules.target_date IS '模块目标日期';
COMMENT ON COLUMN public.modules.sort_order IS '排序权重（默认 65535，越小越靠前）';
COMMENT ON COLUMN public.modules.created_by IS '关联 users.id（创建者）';
COMMENT ON COLUMN public.modules.created_at IS '创建时间（含时区）';
COMMENT ON COLUMN public.modules.updated_at IS '最后更新时间（含时区）';
COMMENT ON COLUMN public.modules.deleted_at IS '软删除时间（含时区），NULL 表示未删除';

-- ============================================================
-- 表: issue_labels — 工作项-标签关联表
-- ============================================================
COMMENT ON TABLE public.issue_labels IS '[核心工作项域]工作项与标签的多对多关联关系';

COMMENT ON COLUMN public.issue_labels.issue_id IS '关联 issues.id（工作项）';
COMMENT ON COLUMN public.issue_labels.label_id IS '关联 labels.id（标签）';

-- ============================================================
-- 表: issue_modules — 工作项-模块关联表
-- ============================================================
COMMENT ON TABLE public.issue_modules IS '[核心工作项域]工作项与模块的多对多关联关系';

COMMENT ON COLUMN public.issue_modules.issue_id IS '关联 issues.id（工作项）';
COMMENT ON COLUMN public.issue_modules.module_id IS '关联 modules.id（模块）';

-- ============================================================
-- 表: issue_assignees — 工作项指派人员表
-- ============================================================
COMMENT ON TABLE public.issue_assignees IS '[核心工作项域]工作项负责人多对多关联（一个工作项可指派多人）';

COMMENT ON COLUMN public.issue_assignees.issue_id IS '关联 issues.id（工作项）';
COMMENT ON COLUMN public.issue_assignees.user_id IS '关联 users.id（被指派人员）';
COMMENT ON COLUMN public.issue_assignees.assigned_at IS '指派时间（含时区）';
COMMENT ON COLUMN public.issue_assignees.assigned_by IS '关联 users.id（执行指派操作的管理员）';

-- ============================================================
-- 表: issue_watchers — 工作项关注者表
-- ============================================================
COMMENT ON TABLE public.issue_watchers IS '[核心工作项域]工作项关注者多对多关联（关注者接收该工作项变更通知）';

COMMENT ON COLUMN public.issue_watchers.issue_id IS '关联 issues.id（被关注的工作项）';
COMMENT ON COLUMN public.issue_watchers.user_id IS '关联 users.id（关注者）';
COMMENT ON COLUMN public.issue_watchers.created_at IS '关注时间（含时zone）';

-- ============================================================
-- 表: issue_comments — 工作项评论表（已有注释，补充缺失字段）
-- ============================================================
COMMENT ON COLUMN public.issue_comments.id IS '自增主键';
COMMENT ON COLUMN public.issue_comments.workspace_id IS '关联 workspaces.id（租户隔离列）';
COMMENT ON COLUMN public.issue_comments.project_id IS '关联 projects.id（所属项目）';
COMMENT ON COLUMN public.issue_comments.issue_id IS '关联 issues.id（所属工作项）';
COMMENT ON COLUMN public.issue_comments.content_json IS 'TipTap 编辑器的 JSON 输出';
COMMENT ON COLUMN public.issue_comments.content_html IS '评论的 HTML 渲染结果';
COMMENT ON COLUMN public.issue_comments.content_stripped IS '评论纯文本摘要（去除富文本标记，供搜索使用）';
COMMENT ON COLUMN public.issue_comments.created_by IS '关联 users.id（评论作者）';
COMMENT ON COLUMN public.issue_comments.mentions IS '@提及的用户 ID 数组';
COMMENT ON COLUMN public.issue_comments.parent_id IS '父评论 ID（嵌套回复）';
COMMENT ON COLUMN public.issue_comments.is_edited IS '是否已被编辑: true=已编辑 / false=原文';
COMMENT ON COLUMN public.issue_comments.edited_at IS '最后编辑时间（含时区）';
COMMENT ON COLUMN public.issue_comments.created_at IS '创建时间（含时区）';
COMMENT ON COLUMN public.issue_comments.updated_at IS '最后更新时间（含时区）';

-- ============================================================
-- 表: issue_activities — 工作项活动表（按月分区）
-- ============================================================
COMMENT ON TABLE public.issue_activities IS '[核心工作项域]工作项变更历史流水（按月分区归档，记录所有字段变更事件）';

COMMENT ON COLUMN public.issue_activities.id IS '自增主键';
COMMENT ON COLUMN public.issue_activities.workspace_id IS '关联 workspaces.id（租户隔离列）';
COMMENT ON COLUMN public.issue_activities.project_id IS '关联 projects.id（所属项目）';
COMMENT ON COLUMN public.issue_activities.issue_id IS '关联 issues.id（所属工作项）';
COMMENT ON COLUMN public.issue_activities.verb IS '操作类型: created(创建) / updated(更新) / transitioned(状态流转) / attached(附件) / linked(关联) / unlinked(取消关联) / commented(评论)';
COMMENT ON COLUMN public.issue_activities.field IS '变更的字段名（verb=updated 时有效）';
COMMENT ON COLUMN public.issue_activities.old_value IS '字段变更前的值（纯文本）';
COMMENT ON COLUMN public.issue_activities.new_value IS '字段变更后的值（纯文本）';
COMMENT ON COLUMN public.issue_activities.old_ref IS '变更前的复杂引用 JSON（如状态对象 {id, name, group}）';
COMMENT ON COLUMN public.issue_activities.new_ref IS '变更后的复杂引用 JSON';
COMMENT ON COLUMN public.issue_activities.actor_id IS '关联 users.id（操作执行人）';
COMMENT ON COLUMN public.issue_activities.actor_email IS '操作人邮箱（冗余存储，方便审计展示）';
COMMENT ON COLUMN public.issue_activities.actor_name IS '操作人显示名（冗余存储）';
COMMENT ON COLUMN public.issue_activities.created_at IS '活动发生时间（含时区）';

-- ============================================================
-- 表: issue_dependencies — 工作项依赖关系表
-- ============================================================
COMMENT ON TABLE public.issue_dependencies IS '[核心工作项域]工作项间依赖关系（FS/SS/FF/SF 四种工程依赖类型，支持 lag_days 延迟）';

COMMENT ON COLUMN public.issue_dependencies.id IS '自增主键';
COMMENT ON COLUMN public.issue_dependencies.workspace_id IS '关联 workspaces.id（租户隔离列）';
COMMENT ON COLUMN public.issue_dependencies.project_id IS '关联 projects.id（所属项目）';
COMMENT ON COLUMN public.issue_dependencies.predecessor_id IS '关联 issues.id（前置工作项，即依赖的源头）';
COMMENT ON COLUMN public.issue_dependencies.successor_id IS '关联 issues.id（后续工作项，即被依赖方）';
COMMENT ON COLUMN public.issue_dependencies.dependency_type IS '依赖类型: FS(完成-开始) / SS(开始-开始) / FF(完成-完成) / SF(开始-完成)';
COMMENT ON COLUMN public.issue_dependencies.lag_days IS '延迟天数（正数表示延后，负数表示提前）';
COMMENT ON COLUMN public.issue_dependencies.created_by IS '关联 users.id（创建者）';
COMMENT ON COLUMN public.issue_dependencies.created_at IS '创建时间（含时区）';

-- ============================================================
-- 表: issue_relations — 工作项关联关系表
-- ============================================================
COMMENT ON TABLE public.issue_relations IS '[核心工作项域]工作项间的语义关联（重复、相关、阻塞、顺序、实现等多种关系类型）';

COMMENT ON COLUMN public.issue_relations.id IS '自增主键';
COMMENT ON COLUMN public.issue_relations.workspace_id IS '关联 workspaces.id（租户隔离列）';
COMMENT ON COLUMN public.issue_relations.project_id IS '关联 projects.id（所属项目）';
COMMENT ON COLUMN public.issue_relations.source_issue_id IS '关联 issues.id（源工作项）';
COMMENT ON COLUMN public.issue_relations.target_issue_id IS '关联 issues.id（目标工作项）';
COMMENT ON COLUMN public.issue_relations.relation_type IS '关联类型: duplicate(重复) / relates_to(相关) / blocked_by(被阻塞) / start_before(先于开始) / finish_before(先于完成) / implemented_by(由...实现)';
COMMENT ON COLUMN public.issue_relations.created_by IS '关联 users.id（创建者）';
COMMENT ON COLUMN public.issue_relations.created_at IS '创建时间（含时区）';

-- ============================================================
-- 表: time_logs — 工时记录表
-- ============================================================
COMMENT ON TABLE public.time_logs IS '[核心工作项域]工作项工时记录（用户每日填报的实际花费时间，单位: 分钟）';

COMMENT ON COLUMN public.time_logs.id IS '自增主键';
COMMENT ON COLUMN public.time_logs.workspace_id IS '关联 workspaces.id（租户隔离列）';
COMMENT ON COLUMN public.time_logs.project_id IS '关联 projects.id（所属项目）';
COMMENT ON COLUMN public.time_logs.issue_id IS '关联 issues.id（工时所属工作项）';
COMMENT ON COLUMN public.time_logs.user_id IS '关联 users.id（工时填报人）';
COMMENT ON COLUMN public.time_logs.spent_date IS '工时消耗日期（默认当天）';
COMMENT ON COLUMN public.time_logs.duration_minutes IS '花费时长（分钟，范围 1-1440）';
COMMENT ON COLUMN public.time_logs.description IS '工时描述（工作内容说明）';
COMMENT ON COLUMN public.time_logs.created_at IS '创建时间（含时区）';
COMMENT ON COLUMN public.time_logs.updated_at IS '最后更新时间（含时区）';
COMMENT ON COLUMN public.time_logs.deleted_at IS '软删除时间（含时区），NULL 表示未删除';

-- ============================================================
-- 表: attachments — 附件表
-- ============================================================
COMMENT ON TABLE public.attachments IS '[核心工作项域]多态附件存储（支持关联到 issue/comment/workspace/project 四种实体）';

COMMENT ON COLUMN public.attachments.id IS '自增主键';
COMMENT ON COLUMN public.attachments.workspace_id IS '关联 workspaces.id（租户隔离列）';
COMMENT ON COLUMN public.attachments.project_id IS '关联 projects.id（所属项目）';
COMMENT ON COLUMN public.attachments.entity_type IS '关联实体类型: issue(工作项) / comment(评论) / workspace(工作空间) / project(项目)';
COMMENT ON COLUMN public.attachments.entity_id IS '关联实体的 ID（配合 entity_type 组成多态关联）';
COMMENT ON COLUMN public.attachments.file_name IS '原始文件名';
COMMENT ON COLUMN public.attachments.file_size IS '文件大小（字节）';
COMMENT ON COLUMN public.attachments.content_type IS 'MIME 类型（默认 application/octet-stream）';
COMMENT ON COLUMN public.attachments.storage_key IS '对象存储中的文件路径/key';
COMMENT ON COLUMN public.attachments.storage_url IS '文件访问 URL';
COMMENT ON COLUMN public.attachments.thumb_key IS '缩略图存储 key（图片/视频文件专用）';
COMMENT ON COLUMN public.attachments.uploaded_by IS '关联 users.id（上传者）';
COMMENT ON COLUMN public.attachments.deleted_at IS '软删除时间（含时区），NULL 表示未删除';
COMMENT ON COLUMN public.attachments.created_at IS '上传时间（含时区）';
COMMENT ON COLUMN public.attachments.updated_at IS '最后更新时间（含时区）';

-- ============================================================================
-- 项目管理域
-- ============================================================================

-- ============================================================
-- 表: projects — 项目表
-- ============================================================
COMMENT ON TABLE public.projects IS '[项目管理域]工作空间下的项目（最小业务容器，所有工作项归属项目）';

COMMENT ON COLUMN public.projects.id IS '自增主键';
COMMENT ON COLUMN public.projects.workspace_id IS '关联 workspaces.id（所属工作空间）';
COMMENT ON COLUMN public.projects.public_id IS '对外暴露的 UUID 主键';
COMMENT ON COLUMN public.projects.name IS '项目名称';
COMMENT ON COLUMN public.projects.slug IS '项目 URL 标识（小写字母+数字+短横线，空间内唯一）';

COMMENT ON COLUMN public.projects.identifier IS '项目标识符（大写字母，如 YD，用于工作项编号前缀）';
COMMENT ON COLUMN public.projects.description IS '项目描述';
COMMENT ON COLUMN public.projects.network IS '可见性: public(空间内所有成员可见) / private(仅项目成员可见)';
COMMENT ON COLUMN public.projects.icon IS '项目图标（Lucide 图标名或 emoji）';
COMMENT ON COLUMN public.projects.color IS '项目颜色（十六进制）';
COMMENT ON COLUMN public.projects.status IS '项目状态: active(进行中) / archived(已归档)';
COMMENT ON COLUMN public.projects.sort_order IS '排序权重（默认 65535）';
COMMENT ON COLUMN public.projects.created_by IS '关联 users.id（创建者）';
COMMENT ON COLUMN public.projects.created_at IS '创建时间（含时区）';
COMMENT ON COLUMN public.projects.updated_at IS '最后更新时间（含时区）';
COMMENT ON COLUMN public.projects.deleted_at IS '软删除时间（含时区），NULL 表示未删除';
COMMENT ON COLUMN public.projects.template IS '项目模板: agile(敏捷) / waterfall(瀑布) / generic(通用)';

-- ============================================================
-- 表: sprints — 迭代/冲刺表
-- ============================================================
COMMENT ON TABLE public.sprints IS '[项目管理域]敏捷迭代/冲刺（包含容量规划、目标设定、状态流转，乐观锁并发控制）';

COMMENT ON COLUMN public.sprints.id IS '自增主键';
COMMENT ON COLUMN public.sprints.workspace_id IS '关联 workspaces.id（租户隔离列）';
COMMENT ON COLUMN public.sprints.project_id IS '关联 projects.id（所属项目）';
COMMENT ON COLUMN public.sprints.name IS '迭代名称';
COMMENT ON COLUMN public.sprints.description IS '迭代描述';
COMMENT ON COLUMN public.sprints.goal IS '迭代目标（Sprint Goal）';
COMMENT ON COLUMN public.sprints.status IS '迭代状态: planned(未开始) / active(进行中) / completed(已完成)';
COMMENT ON COLUMN public.sprints.start_date IS '迭代开始日期';
COMMENT ON COLUMN public.sprints.end_date IS '迭代结束日期';
COMMENT ON COLUMN public.sprints.capacity IS '迭代容量（团队可用工时，NUMERIC(10,2) 小时）';
COMMENT ON COLUMN public.sprints.owner_id IS '关联 users.id（迭代负责人/Scrum Master）';
COMMENT ON COLUMN public.sprints.viewport IS '迭代视图配置 JSON（组件布局、筛选条件等）';
COMMENT ON COLUMN public.sprints.review_snapshot IS '评审快照 JSON（评审会议数据）';
COMMENT ON COLUMN public.sprints.started_at IS '实际启动时间（含时区）';
COMMENT ON COLUMN public.sprints.completed_at IS '完成时间（含时区）';
COMMENT ON COLUMN public.sprints.created_by IS '关联 users.id（创建者）';
COMMENT ON COLUMN public.sprints.created_at IS '创建时间（含时区）';
COMMENT ON COLUMN public.sprints.updated_at IS '最后更新时间（含时区）';
COMMENT ON COLUMN public.sprints.deleted_at IS '软删除时间（含时区），NULL 表示未删除';
COMMENT ON COLUMN public.sprints.version_id IS '关联 versions.id（首次发布版本）';

-- ============================================================
-- 表: sprint_issues — 迭代-工作项关联表
-- ============================================================
COMMENT ON TABLE public.sprint_issues IS '[项目管理域]迭代与工作项的多对多关联（含中途加入标记与排序）';

COMMENT ON COLUMN public.sprint_issues.sprint_id IS '关联 sprints.id（所属迭代）';
COMMENT ON COLUMN public.sprint_issues.issue_id IS '关联 issues.id（工作项）';
COMMENT ON COLUMN public.sprint_issues.added_midway IS '是否中途加入迭代: true=迭代启动后加入 / false=规划时加入';
COMMENT ON COLUMN public.sprint_issues.sort_order IS '迭代内排序权重（默认 65535）';
COMMENT ON COLUMN public.sprint_issues.added_at IS '加入迭代时间（含时区）';
COMMENT ON COLUMN public.sprint_issues.added_by IS '关联 users.id（执行加入操作的人）';

-- ============================================================
-- 表: sprint_snapshots — 迭代快照表
-- ============================================================
COMMENT ON TABLE public.sprint_snapshots IS '[项目管理域]迭代燃尽图/速率等每日快照数据（JSONB 压缩存储历史指标）';

COMMENT ON COLUMN public.sprint_snapshots.id IS '自增主键';
COMMENT ON COLUMN public.sprint_snapshots.workspace_id IS '关联 workspaces.id（租户隔离列）';
COMMENT ON COLUMN public.sprint_snapshots.project_id IS '关联 projects.id（所属项目）';
COMMENT ON COLUMN public.sprint_snapshots.sprint_id IS '关联 sprints.id（所属迭代）';
COMMENT ON COLUMN public.sprint_snapshots.snapshot_date IS '快照日期';
COMMENT ON COLUMN public.sprint_snapshots.data IS '快照数据 JSON（burndown 点数、完成率、速率等指标）';
COMMENT ON COLUMN public.sprint_snapshots.created_at IS '创建时间（含时区）';
COMMENT ON COLUMN public.sprint_snapshots.deleted_at IS '软删除时间（含时区），NULL 表示未删除';

-- ============================================================
-- 表: versions — 版本发布表（已有 version 字段注释，补充其他字段）
-- ============================================================
COMMENT ON TABLE public.versions IS '[项目管理域]版本发布管理（支持 SemVer 规范，含检查清单与发布说明）';

COMMENT ON COLUMN public.versions.id IS '自增主键';
COMMENT ON COLUMN public.versions.workspace_id IS '关联 workspaces.id（租户隔离列）';
COMMENT ON COLUMN public.versions.project_id IS '关联 projects.id（所属项目）';
COMMENT ON COLUMN public.versions.name IS '版本名称';
COMMENT ON COLUMN public.versions.semver IS '语义化版本号（如 1.2.3-beta.1，符合 SemVer 规范）';
COMMENT ON COLUMN public.versions.description IS '版本描述';
COMMENT ON COLUMN public.versions.status IS '版本状态: planning(规划中) / active(开发中) / released(已发布) / archived(已归档)';
COMMENT ON COLUMN public.versions.checklist IS '发布检查清单 JSON 数组（最多 50 项）';
COMMENT ON COLUMN public.versions.release_notes IS '发布说明';
COMMENT ON COLUMN public.versions.delivered_at IS '实际交付时间（含时区）';
COMMENT ON COLUMN public.versions.target_date IS '目标发布日期';
COMMENT ON COLUMN public.versions.archived_at IS '归档时间（含时区）';
COMMENT ON COLUMN public.versions.created_by IS '关联 users.id（创建者）';
COMMENT ON COLUMN public.versions.created_at IS '创建时间（含时区）';
COMMENT ON COLUMN public.versions.updated_at IS '最后更新时间（含时区）';
COMMENT ON COLUMN public.versions.deleted_at IS '软删除时间（含时区），NULL 表示未删除';
COMMENT ON COLUMN public.versions.start_date IS '开发开始日期';
COMMENT ON COLUMN public.versions.end_date IS '开发结束日期';

-- ============================================================
-- 表: version_delivery_snapshots — 版本交付快照表
-- ============================================================
COMMENT ON TABLE public.version_delivery_snapshots IS '[项目管理域]版本交付过程记录（保存进度与质量维度的时间线数据）';

COMMENT ON COLUMN public.version_delivery_snapshots.id IS '自增主键';
COMMENT ON COLUMN public.version_delivery_snapshots.version_id IS '关联 versions.id（所属版本）';
COMMENT ON COLUMN public.version_delivery_snapshots.workspace_id IS '关联 workspaces.id（租户隔离列）';
COMMENT ON COLUMN public.version_delivery_snapshots.progress IS '进度快照 JSON（完成率、阻塞数、剩余工作项等）';
COMMENT ON COLUMN public.version_delivery_snapshots.quality IS '质量快照 JSON（缺陷数、解决率、回归率等）';
COMMENT ON COLUMN public.version_delivery_snapshots.release_notes IS '当时版本的发布说明快照';
COMMENT ON COLUMN public.version_delivery_snapshots.snapshot_at IS '快照记录时间（含时区）';

-- ============================================================================
-- 工作流域
-- ============================================================================

-- ============================================================
-- 表: states — 状态表
-- ============================================================
COMMENT ON TABLE public.states IS '[工作流域]项目工作项状态定义（每个项目可自定义状态集，状态归属到 4 大分组之一）';

COMMENT ON COLUMN public.states.id IS '自增主键';
COMMENT ON COLUMN public.states.workspace_id IS '关联 workspaces.id（租户隔离列）';
COMMENT ON COLUMN public.states.project_id IS '关联 projects.id（所属项目）';
COMMENT ON COLUMN public.states.name IS '状态名称（如 进行中、待测试）';
COMMENT ON COLUMN public.states."group" IS '状态分组: backlog(待办) / started(进行中) / completed(已完成) / cancelled(已取消)';
COMMENT ON COLUMN public.states.color IS '状态颜色标识（十六进制，如 #8DA2C2）';
COMMENT ON COLUMN public.states.sequence IS '状态排序权重（默认 65535，越小越靠前）';
COMMENT ON COLUMN public.states.is_default IS '是否为新建工作项的默认状态: true=默认 / false=非默认';
COMMENT ON COLUMN public.states.created_at IS '创建时间（含时区）';
COMMENT ON COLUMN public.states.updated_at IS '最后更新时间（含时区）';
COMMENT ON COLUMN public.states.deleted_at IS '软删除时间（含时区），NULL 表示未删除';
COMMENT ON COLUMN public.states.template_set IS '模板归属集合: dev_flow(研发流) / defect_flow(缺陷流) / requirement_flow(需求评审流) / custom(自定义)';

-- ============================================================
-- 表: state_transitions — 状态流转规则表
-- ============================================================
COMMENT ON TABLE public.state_transitions IS '[工作流域]状态流转规则（定义 from→to 的合法路径、必填字段与权限约束）';

COMMENT ON COLUMN public.state_transitions.id IS '自增主键';
COMMENT ON COLUMN public.state_transitions.workspace_id IS '关联 workspaces.id（租户隔离列）';
COMMENT ON COLUMN public.state_transitions.project_id IS '关联 projects.id（所属项目）';
COMMENT ON COLUMN public.state_transitions.type_code IS '适用的工作项类型: requirement/task/defect/all（all 表示所有类型）';
COMMENT ON COLUMN public.state_transitions.from_state_id IS '关联 states.id（起始状态）';
COMMENT ON COLUMN public.state_transitions.to_state_id IS '关联 states.id（目标状态）';
COMMENT ON COLUMN public.state_transitions.required_fields IS '流转时需要填写的字段列表 JSON 数组（如 ["root_cause_category"]）';
COMMENT ON COLUMN public.state_transitions.created_at IS '创建时间（含时区）';
COMMENT ON COLUMN public.state_transitions.updated_at IS '最后更新时间（含时区）';

-- ============================================================================
-- 租户与权限域
-- ============================================================================

-- ============================================================
-- 表: users — 用户表
-- ============================================================
COMMENT ON TABLE public.users IS '[租户与权限域]系统用户（平台级账号，可跨工作空间加入多个空间）';

COMMENT ON COLUMN public.users.id IS '自增主键';
COMMENT ON COLUMN public.users.public_id IS '对外暴露的 UUID 主键';
COMMENT ON COLUMN public.users.email IS '用户邮箱（系统内唯一，登录凭证之一）';
COMMENT ON COLUMN public.users.password_hash IS '密码 bcrypt 哈希（OIDC/OAuth 登录用户可为空）';
COMMENT ON COLUMN public.users.display_name IS '用户显示名';
COMMENT ON COLUMN public.users.avatar_url IS '头像图片 URL';
COMMENT ON COLUMN public.users.is_active IS '账号是否激活: true=激活 / false=已禁用';
COMMENT ON COLUMN public.users.timezone IS '用户时区（默认 Asia/Shanghai）';
COMMENT ON COLUMN public.users.created_at IS '注册时间（含时区）';
COMMENT ON COLUMN public.users.updated_at IS '最后更新时间（含时区）';
COMMENT ON COLUMN public.users.deleted_at IS '软删除时间（含时区），NULL 表示未删除';

-- ============================================================
-- 表: workspaces — 工作空间/租户表
-- ============================================================
COMMENT ON TABLE public.workspaces IS '[租户与权限域]工作空间/租户（顶级隔离容器，所有业务表 workspace_id 指向此表）';

COMMENT ON COLUMN public.workspaces.id IS '自增主键';
COMMENT ON COLUMN public.workspaces.name IS '工作空间名称';
COMMENT ON COLUMN public.workspaces.slug IS '工作空间 URL 标识（全局唯一，非归档状态不可重复）';
COMMENT ON COLUMN public.workspaces.logo_url IS '工作空间 Logo URL';
COMMENT ON COLUMN public.workspaces.timezone IS '默认时区（默认 Asia/Shanghai）';
COMMENT ON COLUMN public.workspaces.language IS '默认语言（如 zh-CN, en-US）';
COMMENT ON COLUMN public.workspaces.owner_id IS '关联 users.id（空间所有者）';
COMMENT ON COLUMN public.workspaces.status IS '空间状态: active(活跃) / archived(已归档)';
COMMENT ON COLUMN public.workspaces.created_at IS '创建时间（含时区）';
COMMENT ON COLUMN public.workspaces.updated_at IS '最后更新时间（含时区）';

-- ============================================================
-- 表: invitations — 邀请表
-- ============================================================
COMMENT ON TABLE public.invitations IS '[租户与权限域]工作空间成员邀请（通过邮件发送邀请链接，7 天有效）';

COMMENT ON COLUMN public.invitations.id IS '自增主键';
COMMENT ON COLUMN public.invitations.workspace_id IS '关联 workspaces.id（被邀请的工作空间）';
COMMENT ON COLUMN public.invitations.inviter_id IS '关联 users.id（发送邀请的人）';
COMMENT ON COLUMN public.invitations.email IS '被邀请人邮箱';
COMMENT ON COLUMN public.invitations.role IS '邀请角色: admin(管理员) / member(成员) / guest(访客)';
COMMENT ON COLUMN public.invitations.token_hash IS '邀请令牌哈希（用于验证邀请链接）';
COMMENT ON COLUMN public.invitations.message IS '邀请附言';
COMMENT ON COLUMN public.invitations.status IS '邀请状态: pending(待接受) / accepted(已接受) / revoked(已撤销) / expired(已过期)';
COMMENT ON COLUMN public.invitations.expires_at IS '过期时间（含时区，默认创建后 7 天）';
COMMENT ON COLUMN public.invitations.accepted_at IS '接受时间（含时区）';
COMMENT ON COLUMN public.invitations.created_at IS '创建时间（含时区）';
COMMENT ON COLUMN public.invitations.updated_at IS '最后更新时间（含时区）';

-- ============================================================
-- 表: workspace_members — 工作空间成员表
-- ============================================================
COMMENT ON TABLE public.workspace_members IS '[租户与权限域]工作空间成员关系（用户在不同空间有不同角色）';

COMMENT ON COLUMN public.workspace_members.workspace_id IS '关联 workspaces.id（工作空间）';
COMMENT ON COLUMN public.workspace_members.user_id IS '关联 users.id（成员用户）';
COMMENT ON COLUMN public.workspace_members.role IS '成员角色: owner(所有者) / admin(管理员) / member(普通成员) / guest(访客)';
COMMENT ON COLUMN public.workspace_members.joined_at IS '加入时间（含时区）';


COMMENT ON COLUMN public.projects.identifier IS '项目标识符（大写字母，如 YD，用于工作项编号前缀）';
COMMENT ON COLUMN public.projects.description IS '项目描述';
COMMENT ON COLUMN public.projects.network IS '可见性: public(空间内所有成员可见) / private(仅项目成员可见)';
COMMENT ON COLUMN public.projects.icon IS '项目图标（Lucide 图标名或 emoji）';
COMMENT ON COLUMN public.projects.color IS '项目颜色（十六进制）';
COMMENT ON COLUMN public.projects.status IS '项目状态: active(进行中) / archived(已归档)';
COMMENT ON COLUMN public.projects.sort_order IS '排序权重（默认 65535）';
COMMENT ON COLUMN public.projects.created_by IS '关联 users.id（创建者）';
COMMENT ON COLUMN public.projects.created_at IS '创建时间（含时区）';
COMMENT ON COLUMN public.projects.updated_at IS '最后更新时间（含时区）';
COMMENT ON COLUMN public.projects.deleted_at IS '软删除时间（含时区），NULL 表示未删除';
COMMENT ON COLUMN public.projects.template IS '项目模板: agile(敏捷) / waterfall(瀑布) / generic(通用)';

-- ============================================================
-- 表: sprints — 迭代/冲刺表
-- ============================================================
COMMENT ON TABLE public.sprints IS '[项目管理域]敏捷迭代/冲刺（包含容量规划、目标设定、状态流转，乐观锁并发控制）';

COMMENT ON COLUMN public.sprints.id IS '自增主键';
COMMENT ON COLUMN public.sprints.workspace_id IS '关联 workspaces.id（租户隔离列）';
COMMENT ON COLUMN public.sprints.project_id IS '关联 projects.id（所属项目）';
COMMENT ON COLUMN public.sprints.name IS '迭代名称';
COMMENT ON COLUMN public.sprints.description IS '迭代描述';
COMMENT ON COLUMN public.sprints.goal IS '迭代目标（Sprint Goal）';
COMMENT ON COLUMN public.sprints.status IS '迭代状态: planned(未开始) / active(进行中) / completed(已完成)';
COMMENT ON COLUMN public.sprints.start_date IS '迭代开始日期';
COMMENT ON COLUMN public.sprints.end_date IS '迭代结束日期';
COMMENT ON COLUMN public.sprints.capacity IS '迭代容量（团队可用工时，NUMERIC(10,2) 小时）';
COMMENT ON COLUMN public.sprints.owner_id IS '关联 users.id（迭代负责人/Scrum Master）';
COMMENT ON COLUMN public.sprints.viewport IS '迭代视图配置 JSON（组件布局、筛选条件等）';
COMMENT ON COLUMN public.sprints.review_snapshot IS '评审快照 JSON（评审会议数据）';
COMMENT ON COLUMN public.sprints.started_at IS '实际启动时间（含时区）';
COMMENT ON COLUMN public.sprints.completed_at IS '完成时间（含时区）';
COMMENT ON COLUMN public.sprints.created_by IS '关联 users.id（创建者）';
COMMENT ON COLUMN public.sprints.created_at IS '创建时间（含时区）';
COMMENT ON COLUMN public.sprints.updated_at IS '最后更新时间（含时区）';
COMMENT ON COLUMN public.sprints.deleted_at IS '软删除时间（含时区），NULL 表示未删除';
COMMENT ON COLUMN public.sprints.version_id IS '关联 versions.id（首次发布版本）';

-- ============================================================
-- 表: sprint_issues — 迭代-工作项关联表
-- ============================================================
COMMENT ON TABLE public.sprint_issues IS '[项目管理域]迭代与工作项的多对多关联（含中途加入标记与排序）';

COMMENT ON COLUMN public.sprint_issues.sprint_id IS '关联 sprints.id（所属迭代）';
COMMENT ON COLUMN public.sprint_issues.issue_id IS '关联 issues.id（工作项）';
COMMENT ON COLUMN public.sprint_issues.added_midway IS '是否中途加入迭代: true=迭代启动后加入 / false=规划时加入';
COMMENT ON COLUMN public.sprint_issues.sort_order IS '迭代内排序权重（默认 65535）';
COMMENT ON COLUMN public.sprint_issues.added_at IS '加入迭代时间（含时区）';
COMMENT ON COLUMN public.sprint_issues.added_by IS '关联 users.id（执行加入操作的人）';

-- ============================================================
-- 表: sprint_snapshots — 迭代快照表
-- ============================================================
COMMENT ON TABLE public.sprint_snapshots IS '[项目管理域]迭代燃尽图/速率等每日快照数据（JSONB 压缩存储历史指标）';

COMMENT ON COLUMN public.sprint_snapshots.id IS '自增主键';
COMMENT ON COLUMN public.sprint_snapshots.workspace_id IS '关联 workspaces.id（租户隔离列）';
COMMENT ON COLUMN public.sprint_snapshots.project_id IS '关联 projects.id（所属项目）';
COMMENT ON COLUMN public.sprint_snapshots.sprint_id IS '关联 sprints.id（所属迭代）';
COMMENT ON COLUMN public.sprint_snapshots.snapshot_date IS '快照日期';
COMMENT ON COLUMN public.sprint_snapshots.data IS '快照数据 JSON（burndown 点数、完成率、速率等指标）';
COMMENT ON COLUMN public.sprint_snapshots.created_at IS '创建时间（含时区）';
COMMENT ON COLUMN public.sprint_snapshots.deleted_at IS '软删除时间（含时区），NULL 表示未删除';

-- ============================================================
-- 表: versions — 版本发布表（已有 version 字段注释，补充其他字段）
-- ============================================================
COMMENT ON TABLE public.versions IS '[项目管理域]版本发布管理（支持 SemVer 规范，含检查清单与发布说明）';

COMMENT ON COLUMN public.versions.id IS '自增主键';
COMMENT ON COLUMN public.versions.workspace_id IS '关联 workspaces.id（租户隔离列）';
COMMENT ON COLUMN public.versions.project_id IS '关联 projects.id（所属项目）';
COMMENT ON COLUMN public.versions.name IS '版本名称';
COMMENT ON COLUMN public.versions.semver IS '语义化版本号（如 1.2.3-beta.1，符合 SemVer 规范）';
COMMENT ON COLUMN public.versions.description IS '版本描述';
COMMENT ON COLUMN public.versions.status IS '版本状态: planning(规划中) / active(开发中) / released(已发布) / archived(已归档)';
COMMENT ON COLUMN public.versions.checklist IS '发布检查清单 JSON 数组（最多 50 项）';
COMMENT ON COLUMN public.versions.release_notes IS '发布说明';
COMMENT ON COLUMN public.versions.delivered_at IS '实际交付时间（含时区）';
COMMENT ON COLUMN public.versions.target_date IS '目标发布日期';
COMMENT ON COLUMN public.versions.archived_at IS '归档时间（含时区）';
COMMENT ON COLUMN public.versions.created_by IS '关联 users.id（创建者）';
COMMENT ON COLUMN public.versions.created_at IS '创建时间（含时区）';
COMMENT ON COLUMN public.versions.updated_at IS '最后更新时间（含时区）';
COMMENT ON COLUMN public.versions.deleted_at IS '软删除时间（含时区），NULL 表示未删除';
COMMENT ON COLUMN public.versions.start_date IS '开发开始日期';
COMMENT ON COLUMN public.versions.end_date IS '开发结束日期';

-- ============================================================
-- 表: version_delivery_snapshots — 版本交付快照表
-- ============================================================
COMMENT ON TABLE public.version_delivery_snapshots IS '[项目管理域]版本交付过程记录（保存进度与质量维度的时间线数据）';

COMMENT ON COLUMN public.version_delivery_snapshots.id IS '自增主键';
COMMENT ON COLUMN public.version_delivery_snapshots.version_id IS '关联 versions.id（所属版本）';
COMMENT ON COLUMN public.version_delivery_snapshots.workspace_id IS '关联 workspaces.id（租户隔离列）';
COMMENT ON COLUMN public.version_delivery_snapshots.progress IS '进度快照 JSON（完成率、阻塞数、剩余工作项等）';
COMMENT ON COLUMN public.version_delivery_snapshots.quality IS '质量快照 JSON（缺陷数、解决率、回归率等）';
COMMENT ON COLUMN public.version_delivery_snapshots.release_notes IS '当时版本的发布说明快照';
COMMENT ON COLUMN public.version_delivery_snapshots.snapshot_at IS '快照记录时间（含时区）';

-- ============================================================================
-- 工作流域
-- ============================================================================

-- ============================================================
-- 表: states — 状态表
-- ============================================================
COMMENT ON TABLE public.states IS '[工作流域]项目工作项状态定义（每个项目可自定义状态集，状态归属到 4 大分组之一）';

COMMENT ON COLUMN public.states.id IS '自增主键';
COMMENT ON COLUMN public.states.workspace_id IS '关联 workspaces.id（租户隔离列）';
COMMENT ON COLUMN public.states.project_id IS '关联 projects.id（所属项目）';
COMMENT ON COLUMN public.states.name IS '状态名称（如 进行中、待测试）';
COMMENT ON COLUMN public.states."group" IS '状态分组: backlog(待办) / started(进行中) / completed(已完成) / cancelled(已取消)';
COMMENT ON COLUMN public.states.color IS '状态颜色标识（十六进制，如 #8DA2C2）';
COMMENT ON COLUMN public.states.sequence IS '状态排序权重（默认 65535，越小越靠前）';
COMMENT ON COLUMN public.states.is_default IS '是否为新建工作项的默认状态: true=默认 / false=非默认';
COMMENT ON COLUMN public.states.created_at IS '创建时间（含时区）';
COMMENT ON COLUMN public.states.updated_at IS '最后更新时间（含时区）';
COMMENT ON COLUMN public.states.deleted_at IS '软删除时间（含时区），NULL 表示未删除';
COMMENT ON COLUMN public.states.template_set IS '模板归属集合: dev_flow(研发流) / defect_flow(缺陷流) / requirement_flow(需求评审流) / custom(自定义)';

-- ============================================================
-- 表: state_transitions — 状态流转规则表
-- ============================================================
COMMENT ON TABLE public.state_transitions IS '[工作流域]状态流转规则（定义 from→to 的合法路径、必填字段与权限约束）';

COMMENT ON COLUMN public.state_transitions.id IS '自增主键';
COMMENT ON COLUMN public.state_transitions.workspace_id IS '关联 workspaces.id（租户隔离列）';
COMMENT ON COLUMN public.state_transitions.project_id IS '关联 projects.id（所属项目）';
COMMENT ON COLUMN public.state_transitions.type_code IS '适用的工作项类型: requirement/task/defect/all（all 表示所有类型）';
COMMENT ON COLUMN public.state_transitions.from_state_id IS '关联 states.id（起始状态）';
COMMENT ON COLUMN public.state_transitions.to_state_id IS '关联 states.id（目标状态）';
COMMENT ON COLUMN public.state_transitions.required_fields IS '流转时需要填写的字段列表 JSON 数组（如 ["root_cause_category"]）';
COMMENT ON COLUMN public.state_transitions.created_at IS '创建时间（含时区）';
COMMENT ON COLUMN public.state_transitions.updated_at IS '最后更新时间（含时区）';

-- ============================================================================
-- 租户与权限域
-- ============================================================================

-- ============================================================
-- 表: users — 用户表
-- ============================================================
COMMENT ON TABLE public.users IS '[租户与权限域]系统用户（平台级账号，可跨工作空间加入多个空间）';

COMMENT ON COLUMN public.users.id IS '自增主键';
COMMENT ON COLUMN public.users.public_id IS '对外暴露的 UUID 主键';
COMMENT ON COLUMN public.users.email IS '用户邮箱（系统内唯一，登录凭证之一）';
COMMENT ON COLUMN public.users.password_hash IS '密码 bcrypt 哈希（OIDC/OAuth 登录用户可为空）';
COMMENT ON COLUMN public.users.display_name IS '用户显示名';
COMMENT ON COLUMN public.users.avatar_url IS '头像图片 URL';
COMMENT ON COLUMN public.users.is_active IS '账号是否激活: true=激活 / false=已禁用';
COMMENT ON COLUMN public.users.timezone IS '用户时区（默认 Asia/Shanghai）';
COMMENT ON COLUMN public.users.created_at IS '注册时间（含时区）';
COMMENT ON COLUMN public.users.updated_at IS '最后更新时间（含时区）';
COMMENT ON COLUMN public.users.deleted_at IS '软删除时间（含时区），NULL 表示未删除';

-- ============================================================
-- 表: workspaces — 工作空间/租户表
-- ============================================================
COMMENT ON TABLE public.workspaces IS '[租户与权限域]工作空间/租户（顶级隔离容器，所有业务表 workspace_id 指向此表）';

COMMENT ON COLUMN public.workspaces.id IS '自增主键';
COMMENT ON COLUMN public.workspaces.name IS '工作空间名称';
COMMENT ON COLUMN public.workspaces.slug IS '工作空间 URL 标识（全局唯一，非归档状态不可重复）';
COMMENT ON COLUMN public.workspaces.logo_url IS '工作空间 Logo URL';
COMMENT ON COLUMN public.workspaces.timezone IS '默认时区（默认 Asia/Shanghai）';
COMMENT ON COLUMN public.workspaces.language IS '默认语言（如 zh-CN, en-US）';
COMMENT ON COLUMN public.workspaces.owner_id IS '关联 users.id（空间所有者）';
COMMENT ON COLUMN public.workspaces.status IS '空间状态: active(活跃) / archived(已归档)';
COMMENT ON COLUMN public.workspaces.created_at IS '创建时间（含时区）';
COMMENT ON COLUMN public.workspaces.updated_at IS '最后更新时间（含时区）';

-- ============================================================
-- 表: invitations — 邀请表
-- ============================================================
COMMENT ON TABLE public.invitations IS '[租户与权限域]工作空间成员邀请（通过邮件发送邀请链接，7 天有效）';

COMMENT ON COLUMN public.invitations.id IS '自增主键';
COMMENT ON COLUMN public.invitations.workspace_id IS '关联 workspaces.id（被邀请的工作空间）';
COMMENT ON COLUMN public.invitations.inviter_id IS '关联 users.id（发送邀请的人）';
COMMENT ON COLUMN public.invitations.email IS '被邀请人邮箱';
COMMENT ON COLUMN public.invitations.role IS '邀请角色: admin(管理员) / member(成员) / guest(访客)';
COMMENT ON COLUMN public.invitations.token_hash IS '邀请令牌哈希（用于验证邀请链接）';
COMMENT ON COLUMN public.invitations.message IS '邀请附言';
COMMENT ON COLUMN public.invitations.status IS '邀请状态: pending(待接受) / accepted(已接受) / revoked(已撤销) / expired(已过期)';
COMMENT ON COLUMN public.invitations.expires_at IS '过期时间（含时区，默认创建后 7 天）';
COMMENT ON COLUMN public.invitations.accepted_at IS '接受时间（含时区）';
COMMENT ON COLUMN public.invitations.created_at IS '创建时间（含时区）';
COMMENT ON COLUMN public.invitations.updated_at IS '最后更新时间（含时区）';

-- ============================================================
-- 表: workspace_members — 工作空间成员表
-- ============================================================
COMMENT ON TABLE public.workspace_members IS '[租户与权限域]工作空间成员关系（用户在不同空间有不同角色）';

COMMENT ON COLUMN public.workspace_members.workspace_id IS '关联 workspaces.id（工作空间）';
COMMENT ON COLUMN public.workspace_members.user_id IS '关联 users.id（成员用户）';
COMMENT ON COLUMN public.workspace_members.role IS '成员角色: owner(所有者) / admin(管理员) / member(普通成员) / guest(访客)';
COMMENT ON COLUMN public.workspace_members.joined_at IS '加入时间（含时区）';


-- ============================================================================
-- 通知域
-- ============================================================================

-- ============================================================
-- 表: notifications — 通知消息表（已有注释，补充缺失字段）
-- ============================================================
COMMENT ON COLUMN public.notifications.id IS '自增主键';
COMMENT ON COLUMN public.notifications.workspace_id IS '关联 workspaces.id（租户隔离列）';
COMMENT ON COLUMN public.notifications.recipient_id IS '关联 users.id（通知接收人）';
COMMENT ON COLUMN public.notifications.event_type IS '事件类型: issue.created/issue.assigned/issue.status_changed/comment.created/sprint.started/sprint.completed/version.released/member.added';
COMMENT ON COLUMN public.notifications.entity_type IS '关联对象类型: issue/sprint/version/project/workspace/comment';
COMMENT ON COLUMN public.notifications.entity_id IS '关联对象的 ID';
COMMENT ON COLUMN public.notifications.title IS '通知标题';
COMMENT ON COLUMN public.notifications.body IS '通知正文（富文本）';
COMMENT ON COLUMN public.notifications.action_url IS '点击通知跳转的 URL';
COMMENT ON COLUMN public.notifications.actor_id IS '关联 users.id（触发该通知的操作人）';
COMMENT ON COLUMN public.notifications.actor_name IS '触发通知的操作人显示名（冗余存储）';
COMMENT ON COLUMN public.notifications.is_read IS '是否已读: true=已读 / false=未读';
COMMENT ON COLUMN public.notifications.is_archived IS '是否归档: true=已归档 / false=正常';
COMMENT ON COLUMN public.notifications.read_at IS '阅读时间（含时区）';
COMMENT ON COLUMN public.notifications.channel IS '通知渠道: in_app(站内)/email/sms/wecom/dingtalk/feishu';
COMMENT ON COLUMN public.notifications.payload IS '通知附加数据 JSON（包含跳转上下文等）';
COMMENT ON COLUMN public.notifications.created_at IS '创建时间（含时区）';

-- ============================================================
-- 表: notification_preferences — 通知偏好表（已有表注释，补充字段）
-- ============================================================
COMMENT ON COLUMN public.notification_preferences.id IS '自增主键';
COMMENT ON COLUMN public.notification_preferences.user_id IS '关联 users.id（用户）';
COMMENT ON COLUMN public.notification_preferences.workspace_id IS '关联 workspaces.id（工作空间）';
COMMENT ON COLUMN public.notification_preferences.event_types IS '订阅的事件类型列表 JSON 数组';
COMMENT ON COLUMN public.notification_preferences.channels IS '通知渠道配置 JSON 数组（如 ["in_app", "email"]）';
COMMENT ON COLUMN public.notification_preferences.digest IS '聚合发送频率: realtime(实时) / daily(每日) / weekly(每周)';
COMMENT ON COLUMN public.notification_preferences.dnd_enabled IS '是否启用免打扰: true=启用 / false=关闭';
COMMENT ON COLUMN public.notification_preferences.dnd_start IS '免打扰开始时间（默认 22:00）';
COMMENT ON COLUMN public.notification_preferences.dnd_end IS '免打扰结束时间（默认 08:00）';
COMMENT ON COLUMN public.notification_preferences.is_enabled IS '是否启用通知: true=启用 / false=关闭';
COMMENT ON COLUMN public.notification_preferences.created_at IS '创建时间（含时区）';
COMMENT ON COLUMN public.notification_preferences.updated_at IS '最后更新时间（含时区）';

-- ============================================================
-- 表: notification_deliveries — 通知投递记录表
-- ============================================================
COMMENT ON TABLE public.notification_deliveries IS '[通知域]通知多渠道投递记录（追踪每条通知的发送状态与重试）';

COMMENT ON COLUMN public.notification_deliveries.id IS '自增主键';
COMMENT ON COLUMN public.notification_deliveries.notification_id IS '关联 notifications.id（所属通知）';
COMMENT ON COLUMN public.notification_deliveries.channel IS '投递渠道: in_app / email / sms / wecom / dingtalk / feishu';
COMMENT ON COLUMN public.notification_deliveries.status IS '投递状态: pending(待发送) / sent(已发送) / failed(失败) / skipped(跳过)';
COMMENT ON COLUMN public.notification_deliveries.recipient IS '接收方标识（邮箱/手机号/用户ID等）';
COMMENT ON COLUMN public.notification_deliveries.sent_at IS '实际发送时间（含时区）';
COMMENT ON COLUMN public.notification_deliveries.error_msg IS '发送失败时的错误信息';
COMMENT ON COLUMN public.notification_deliveries.retry_count IS '已重试次数';
COMMENT ON COLUMN public.notification_deliveries.created_at IS '创建时间（含时区）';
COMMENT ON COLUMN public.notification_deliveries.next_retry_at IS '下次重试时间（含时区）';

-- ============================================================
-- 表: notification_digests — 通知摘要/聚合
-- ============================================================
COMMENT ON TABLE public.notification_digests IS '[通知域]通知聚合摘要（将多条通知打包为每日/每周摘要定时发送）';

COMMENT ON COLUMN public.notification_digests.id IS '自增主键';
COMMENT ON COLUMN public.notification_digests.user_id IS '关联 users.id（摘要接收人）';
COMMENT ON COLUMN public.notification_digests.workspace_id IS '关联 workspaces.id（工作空间）';
COMMENT ON COLUMN public.notification_digests.digest_type IS '摘要类型: daily(每日) / weekly(每周)';
COMMENT ON COLUMN public.notification_digests.notification_ids IS '聚合的通知 ID 数组';
COMMENT ON COLUMN public.notification_digests.status IS '摘要状态: pending(待发送) / sent(已发送) / failed(失败)';
COMMENT ON COLUMN public.notification_digests.scheduled_for IS '计划发送时间（含时区）';
COMMENT ON COLUMN public.notification_digests.sent_at IS '实际发送时间（含时区）';
COMMENT ON COLUMN public.notification_digests.created_at IS '创建时间（含时区）';

-- ============================================================================
-- 入口工单域
-- ============================================================================

-- ============================================================
-- 表: intake_channels — 入口渠道表
-- ============================================================
COMMENT ON TABLE public.intake_channels IS '[入口工单域]工单接收渠道（公开提交入口，可配置默认类型、指派规则和自定义字段）';

COMMENT ON COLUMN public.intake_channels.id IS '自增主键';
COMMENT ON COLUMN public.intake_channels.workspace_id IS '关联 workspaces.id（租户隔离列）';
COMMENT ON COLUMN public.intake_channels.project_id IS '关联 projects.id（工单转入的项目，可为空）';
COMMENT ON COLUMN public.intake_channels.slug IS '渠道 URL 标识（空间内唯一，is_active 时不可重复）';
COMMENT ON COLUMN public.intake_channels.name IS '渠道名称';
COMMENT ON COLUMN public.intake_channels.description IS '渠道说明';
COMMENT ON COLUMN public.intake_channels.is_public IS '是否公开访问: true=无需登录即可提交 / false=需要认证';
COMMENT ON COLUMN public.intake_channels.default_issue_type IS '工单默认类型: requirement / task / defect';
COMMENT ON COLUMN public.intake_channels.default_priority IS '默认优先级（0=无, 1=紧急, 2=高, 3=中, 4=低）';
COMMENT ON COLUMN public.intake_channels.auto_assign_rules IS '自动指派规则 JSON 数组';
COMMENT ON COLUMN public.intake_channels.rate_limit_per_min IS '每分钟提交限流数（默认 20）';
COMMENT ON COLUMN public.intake_channels.require_captcha IS '是否需要验证码: true=需要 / false=不需要';
COMMENT ON COLUMN public.intake_channels.custom_fields IS '自定义字段配置 JSON 数组';
COMMENT ON COLUMN public.intake_channels.branding IS '品牌配置 JSON（标题、Logo、说明文字）';
COMMENT ON COLUMN public.intake_channels.notify_on_submit IS '提交时是否通知管理员: true=通知 / false=不通知';
COMMENT ON COLUMN public.intake_channels.notify_users IS '提交时通知的管理员用户 ID 数组';
COMMENT ON COLUMN public.intake_channels.is_active IS '是否启用: true=启用 / false=禁用';
COMMENT ON COLUMN public.intake_channels.created_by IS '关联 users.id（创建者）';
COMMENT ON COLUMN public.intake_channels.created_at IS '创建时间（含时区）';
COMMENT ON COLUMN public.intake_channels.updated_at IS '最后更新时间（含时区）';

-- ============================================================
-- 表: intake_issues — 入口工单表
-- ============================================================
COMMENT ON TABLE public.intake_issues IS '[入口工单域]通过入口渠道提交的工单（可转换为正式 issues 工作项）';

COMMENT ON COLUMN public.intake_issues.id IS '自增主键';
COMMENT ON COLUMN public.intake_issues.channel_id IS '关联 intake_channels.id（提交渠道）';
COMMENT ON COLUMN public.intake_issues.workspace_id IS '关联 workspaces.id（租户隔离列）';
COMMENT ON COLUMN public.intake_issues.project_id IS '关联 projects.id（转入项目，可为空）';
COMMENT ON COLUMN public.intake_issues.tracking_id IS '工单追踪编号（空间内唯一，对外展示）';
COMMENT ON COLUMN public.intake_issues.submitter_name IS '提交者姓名';
COMMENT ON COLUMN public.intake_issues.submitter_email IS '提交者邮箱';
COMMENT ON COLUMN public.intake_issues.submitter_user_id IS '关联 users.id（如果提交者是系统用户）';
COMMENT ON COLUMN public.intake_issues.title IS '工单标题';
COMMENT ON COLUMN public.intake_issues.description IS '工单描述（纯文本）';
COMMENT ON COLUMN public.intake_issues.issue_type IS '工单类型: requirement / task / defect';
COMMENT ON COLUMN public.intake_issues.priority IS '优先级（0=无, 1=紧急, 2=高, 3=中, 4=低）';
COMMENT ON COLUMN public.intake_issues.custom_fields IS '自定义字段值 JSON';
COMMENT ON COLUMN public.intake_issues.attachment_ids IS '关联附件 ID 数组';
COMMENT ON COLUMN public.intake_issues.status IS '工单状态: open(待处理) / accepted(已接受) / rejected(已拒绝) / archived(已归档)';
COMMENT ON COLUMN public.intake_issues.status_reason IS '状态变更原因';
COMMENT ON COLUMN public.intake_issues.converted_issue_id IS '关联 issues.id（转换后的正式工作项 ID）';
COMMENT ON COLUMN public.intake_issues.assigned_to IS '关联 users.id（工单指定处理人）';
COMMENT ON COLUMN public.intake_issues.reviewed_by IS '关联 users.id（审核人）';
COMMENT ON COLUMN public.intake_issues.reviewed_at IS '审核时间（含时区）';
COMMENT ON COLUMN public.intake_issues.notify_on_status IS '状态变更时是否通知提交者: true=通知 / false=不通知';
COMMENT ON COLUMN public.intake_issues.created_at IS '创建时间（含时区）';
COMMENT ON COLUMN public.intake_issues.updated_at IS '最后更新时间（含时区）';


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
COMMENT ON COLUMN public.dashboard_snapshots.widget_type IS '组件类型（与 dashboard_widgets.widget_type 枚举值相同）';
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


-- ============================================================================
-- 风险与度量域
-- ============================================================================

-- ============================================================
-- 表: risk_rules — 风险规则表
-- ============================================================
COMMENT ON TABLE public.risk_rules IS '[风险与度量域]风险检测规则（预定义阈值条件，触发时自动生成 risk_alerts 告警）';

COMMENT ON COLUMN public.risk_rules.id IS '自增主键';
COMMENT ON COLUMN public.risk_rules.workspace_id IS '关联 workspaces.id（租户隔离列）';
COMMENT ON COLUMN public.risk_rules.project_id IS '关联 projects.id（项目级规则，为空表示空间级）';
COMMENT ON COLUMN public.risk_rules.rule_name IS '规则名称';
COMMENT ON COLUMN public.risk_rules.rule_type IS '规则类型: overdue_issue(逾期工作项) / overdue_sprint(逾期迭代) / blocked_count(阻塞数超标) / sla_breach(SLA违约) / stalled_progress(进度停滞) / high_priority_open(高优未关闭)';
COMMENT ON COLUMN public.risk_rules.condition_json IS '规则条件配置 JSON（阈值、比较运算符、检测频率等）';
COMMENT ON COLUMN public.risk_rules.notify_channels IS '告警通知渠道数组（in_app / email / webhook）';
COMMENT ON COLUMN public.risk_rules.is_active IS '规则是否启用: true=启用监控 / false=暂停';
COMMENT ON COLUMN public.risk_rules.last_triggered IS '最近一次触发时间（含时区）';
COMMENT ON COLUMN public.risk_rules.created_at IS '创建时间（含时区）';
COMMENT ON COLUMN public.risk_rules.updated_at IS '最后更新时间（含时区）';

-- ============================================================
-- 表: risk_alerts — 风险告警表
-- ============================================================
COMMENT ON TABLE public.risk_alerts IS '[风险与度量域]风险告警记录（由 risk_rules 自动生成，需要人工确认和解决）';

COMMENT ON COLUMN public.risk_alerts.id IS '自增主键';
COMMENT ON COLUMN public.risk_alerts.workspace_id IS '关联 workspaces.id（租户隔离列）';
COMMENT ON COLUMN public.risk_alerts.project_id IS '关联 projects.id（告警所属项目）';
COMMENT ON COLUMN public.risk_alerts.rule_id IS '关联 risk_rules.id（触发规则）';
COMMENT ON COLUMN public.risk_alerts.severity IS '告警严重度: info(信息) / low(低) / medium(中) / high(高) / critical(严重)';
COMMENT ON COLUMN public.risk_alerts.title IS '告警标题';
COMMENT ON COLUMN public.risk_alerts.description IS '告警描述（详细说明了触发原因）';
COMMENT ON COLUMN public.risk_alerts.metadata IS '告警元数据 JSON（触发时的上下文快照）';
COMMENT ON COLUMN public.risk_alerts.is_resolved IS '是否已解决: true=已解决 / false=未解决';
COMMENT ON COLUMN public.risk_alerts.resolved_at IS '解决时间（含时区）';
COMMENT ON COLUMN public.risk_alerts.resolved_by IS '关联 users.id（解决人）';
COMMENT ON COLUMN public.risk_alerts.created_at IS '创建时间（含时区）';

-- ============================================================
-- 表: metric_snapshots — 指标快照表
-- ============================================================
COMMENT ON TABLE public.metric_snapshots IS '[风险与度量域]项目/空间级效能指标每日快照（支撑趋势分析、燃尽图、速率图）';

COMMENT ON COLUMN public.metric_snapshots.id IS '自增主键';
COMMENT ON COLUMN public.metric_snapshots.workspace_id IS '关联 workspaces.id（租户隔离列）';
COMMENT ON COLUMN public.metric_snapshots.project_id IS '关联 projects.id（项目级指标，为空表示空间级）';
COMMENT ON COLUMN public.metric_snapshots.granularity IS '粒度: daily(日级) / sprint(迭代级) / version(版本级)';
COMMENT ON COLUMN public.metric_snapshots.ref_id IS '关联对象 ID（sprint 粒度时为 sprints.id，version 粒度时为 versions.id）';
COMMENT ON COLUMN public.metric_snapshots.metric IS '指标名称（如 velocity、burndown_points、bug_count）';
COMMENT ON COLUMN public.metric_snapshots.value IS '指标值（NUMERIC(12,4)）';
COMMENT ON COLUMN public.metric_snapshots.dimensions IS '维度标签 JSON（按工作项类型、优先级等切分的子指标）';
COMMENT ON COLUMN public.metric_snapshots.snapshot_date IS '快照日期';
COMMENT ON COLUMN public.metric_snapshots.created_at IS '创建时间（含时区）';

-- ============================================================
-- 表: metric_adjustments — 指标调整记录表
-- ============================================================
COMMENT ON TABLE public.metric_adjustments IS '[风险与度量域]指标人工修正记录（对异常指标进行手动调整的审计日志）';

COMMENT ON COLUMN public.metric_adjustments.id IS '自增主键';
COMMENT ON COLUMN public.metric_adjustments.workspace_id IS '关联 workspaces.id（租户隔离列）';
COMMENT ON COLUMN public.metric_adjustments.project_id IS '关联 projects.id（所属项目）';
COMMENT ON COLUMN public.metric_adjustments.snapshot_id IS '关联 metric_snapshots.id（被修正的快照）';
COMMENT ON COLUMN public.metric_adjustments.metric IS '被修正的指标名';
COMMENT ON COLUMN public.metric_adjustments.snapshot_date IS '快照日期';
COMMENT ON COLUMN public.metric_adjustments.original_value IS '原始指标值';
COMMENT ON COLUMN public.metric_adjustments.adjusted_value IS '修正后的指标值';
COMMENT ON COLUMN public.metric_adjustments.reason IS '修正原因（必填）';
COMMENT ON COLUMN public.metric_adjustments.adjusted_by IS '关联 users.id（修正人）';
COMMENT ON COLUMN public.metric_adjustments.created_at IS '创建时间（含时区）';

-- ============================================================================
-- 集成与扩展域
-- ============================================================================

-- ============================================================
-- 表: webhooks — Webhook 配置表
-- ============================================================
COMMENT ON TABLE public.webhooks IS '[集成与扩展域]Webhook 配置（项目级 HTTP 回调，事件触发时推送 JSON 报文）';

COMMENT ON COLUMN public.webhooks.id IS '自增主键';
COMMENT ON COLUMN public.webhooks.workspace_id IS '关联 workspaces.id（租户隔离列）';
COMMENT ON COLUMN public.webhooks.project_id IS '关联 projects.id（项目级 webhook，可为空表示空间级）';
COMMENT ON COLUMN public.webhooks.name IS 'Webhook 名称';
COMMENT ON COLUMN public.webhooks.target_url IS '回调目标 URL（必须 http:// 或 https:// 开头）';
COMMENT ON COLUMN public.webhooks.secret IS '签名密钥（用于 HMAC 签名验证）';
COMMENT ON COLUMN public.webhooks.events IS '订阅事件类型数组（如 issue.created, issue.updated）';
COMMENT ON COLUMN public.webhooks.is_active IS '是否启用: true=启用 / false=禁用';
COMMENT ON COLUMN public.webhooks.last_error IS '最近一次发送失败的错误信息';
COMMENT ON COLUMN public.webhooks.last_triggered IS '最近一次触发时间（含时区）';
COMMENT ON COLUMN public.webhooks.last_status IS '最近一次执行状态: success / failed';
COMMENT ON COLUMN public.webhooks.unhealthy_at IS '判定为不健康的时间（含时区连续失败超过阈值）';
COMMENT ON COLUMN public.webhooks.created_by IS '关联 users.id（创建者）';
COMMENT ON COLUMN public.webhooks.created_at IS '创建时间（含时区）';
COMMENT ON COLUMN public.webhooks.updated_at IS '最后更新时间（含时区）';

-- ============================================================
-- 表: webhook_logs — Webhook 日志表
-- ============================================================
COMMENT ON TABLE public.webhook_logs IS '[集成与扩展域]Webhook 投递日志（每次回调尝试的完整请求/响应记录，支持重放排错）';

COMMENT ON COLUMN public.webhook_logs.id IS '自增主键';
COMMENT ON COLUMN public.webhook_logs.webhook_id IS '关联 webhooks.id（所属 webhook）';
COMMENT ON COLUMN public.webhook_logs.workspace_id IS '关联 workspaces.id（租户隔离列）';
COMMENT ON COLUMN public.webhook_logs.delivery_id IS '投递 ID（用于去重和追踪）';
COMMENT ON COLUMN public.webhook_logs.event_type IS '触发事件类型';
COMMENT ON COLUMN public.webhook_logs.event_id IS '关联 domain_events.id（触发事件 ID）';
COMMENT ON COLUMN public.webhook_logs.request_url IS '请求目标 URL';
COMMENT ON COLUMN public.webhook_logs.request_method IS 'HTTP 方法（默认 POST）';
COMMENT ON COLUMN public.webhook_logs.request_headers IS '请求头 JSON';
COMMENT ON COLUMN public.webhook_logs.request_body IS '请求体（JSON 序列化的事件数据）';
COMMENT ON COLUMN public.webhook_logs.response_status IS 'HTTP 响应状态码';
COMMENT ON COLUMN public.webhook_logs.response_body IS 'HTTP 响应体';
COMMENT ON COLUMN public.webhook_logs.response_headers IS '响应头 JSON';
COMMENT ON COLUMN public.webhook_logs.status IS '投递状态: success / failed / pending / retrying';
COMMENT ON COLUMN public.webhook_logs.attempt IS '尝试次数';
COMMENT ON COLUMN public.webhook_logs.duration_ms IS '请求耗时（毫秒）';
COMMENT ON COLUMN public.webhook_logs.error IS '错误信息';
COMMENT ON COLUMN public.webhook_logs.occurred_at IS '发生时间（含时区）';

-- ============================================================
-- 表: api_tokens — API 令牌表
-- ============================================================
COMMENT ON TABLE public.api_tokens IS '[集成与扩展域]用户 API 令牌（用于编程访问 API，支持 scope 权限控制与撤销）';

COMMENT ON COLUMN public.api_tokens.id IS '自增主键';
COMMENT ON COLUMN public.api_tokens.user_id IS '关联 users.id（令牌所属用户）';
COMMENT ON COLUMN public.api_tokens.name IS '令牌名称（仅展示用途）';
COMMENT ON COLUMN public.api_tokens.token_hash IS '令牌哈希（存储 hash，原始值仅创建时返回一次）';
COMMENT ON COLUMN public.api_tokens.scopes IS '权限范围 JSON 数组（如 ["read:workspace", "write:issues"]）';
COMMENT ON COLUMN public.api_tokens.last_used_at IS '最近一次使用时间（含时区）';
COMMENT ON COLUMN public.api_tokens.expires_at IS '过期时间（含时区）';
COMMENT ON COLUMN public.api_tokens.revoked_at IS '撤销时间（含时区），NULL 表示未撤销';
COMMENT ON COLUMN public.api_tokens.created_at IS '创建时间（含时区）';
COMMENT ON COLUMN public.api_tokens.updated_at IS '最后更新时间（含时区）';

-- ============================================================
-- 表: deployment_events — 部署事件表
-- ============================================================
COMMENT ON TABLE public.deployment_events IS '[集成与扩展域]部署流水线事件（接收 CI/CD 回调，追踪代码部署状态与关联）';

COMMENT ON COLUMN public.deployment_events.id IS '自增主键';
COMMENT ON COLUMN public.deployment_events.workspace_id IS '关联 workspaces.id（租户隔离列）';
COMMENT ON COLUMN public.deployment_events.project_id IS '关联 projects.id（部署项目）';
COMMENT ON COLUMN public.deployment_events.deployment_id IS '部署 ID（用于幂等去重）';
COMMENT ON COLUMN public.deployment_events.env IS '部署环境: development / staging / production / testing';
COMMENT ON COLUMN public.deployment_events.status IS '部署状态: success(成功) / failed(失败) / rolled_back(已回滚)';
COMMENT ON COLUMN public.deployment_events.commit_sha IS '部署的 Git commit SHA';
COMMENT ON COLUMN public.deployment_events.started_at IS '部署开始时间（含时区）';
COMMENT ON COLUMN public.deployment_events.deployed_at IS '部署完成时间（含时区）';
COMMENT ON COLUMN public.deployment_events.source IS '事件来源: webhook(回调) / api(接口) / cli(命令行)';
COMMENT ON COLUMN public.deployment_events.metadata IS '部署元数据 JSON（流水线名称、执行人等）';
COMMENT ON COLUMN public.deployment_events.created_at IS '创建时间（含时区）';

-- ============================================================================
-- 系统基础设施域
-- ============================================================================

-- ============================================================
-- 表: domain_events — 领域事件表（Outbox 模式）
-- ============================================================
COMMENT ON TABLE public.domain_events IS '[系统基础设施域]领域事件表（Transactional Outbox 模式，保证业务操作与事件发布的原子性）';

COMMENT ON COLUMN public.domain_events.id IS '自增主键';
COMMENT ON COLUMN public.domain_events.workspace_id IS '关联 workspaces.id（租户隔离列）';
COMMENT ON COLUMN public.domain_events.aggregate_type IS '聚合根类型（如 issue、sprint、version）';
COMMENT ON COLUMN public.domain_events.aggregate_id IS '聚合根 ID';
COMMENT ON COLUMN public.domain_events.event_type IS '事件类型（如 issue.status_changed、issue.created）';
COMMENT ON COLUMN public.domain_events.payload IS '事件数据 JSON（事件详情）';
COMMENT ON COLUMN public.domain_events.occurred_at IS '事件发生时间（含时区）';
COMMENT ON COLUMN public.domain_events.published_at IS '事件发布时间（含时区），NULL 表示未发布';

-- ============================================================
-- 表: idempotency_keys — API 幂等键表
-- ============================================================
COMMENT ON TABLE public.idempotency_keys IS '[系统基础设施域]API 幂等键（防止重复提交，存储首次响应用于重放）';

COMMENT ON COLUMN public.idempotency_keys.key IS '幂等键（UUID，客户端生成）';
COMMENT ON COLUMN public.idempotency_keys.user_id IS '关联 users.id（请求用户）';
COMMENT ON COLUMN public.idempotency_keys.response IS '首次响应 JSON（用于重复请求时直接返回）';
COMMENT ON COLUMN public.idempotency_keys.created_at IS '创建时间（含时区）';

-- ============================================================
-- 表: password_reset_tokens — 密码重置令牌表
-- ============================================================
COMMENT ON TABLE public.password_reset_tokens IS '[系统基础设施域]密码重置令牌（一次性使用，过期后失效）';

COMMENT ON COLUMN public.password_reset_tokens.id IS '自增主键';
COMMENT ON COLUMN public.password_reset_tokens.user_id IS '关联 users.id（申请重置的用户）';
COMMENT ON COLUMN public.password_reset_tokens.token_hash IS '令牌哈希（通过邮件发送原始值）';
COMMENT ON COLUMN public.password_reset_tokens.expires_at IS '过期时间（含时区）';
COMMENT ON COLUMN public.password_reset_tokens.used_at IS '使用时间（含时区），NULL 表示未使用';
COMMENT ON COLUMN public.password_reset_tokens.created_at IS '创建时间（含时区）';

-- ============================================================
-- 表: audit_logs — 审计日志表
-- ============================================================
COMMENT ON TABLE public.audit_logs IS '[系统基础设施域]操作审计日志（记录所有实体变更和关键操作，供合规审查）';

COMMENT ON COLUMN public.audit_logs.id IS '自增主键';
COMMENT ON COLUMN public.audit_logs.workspace_id IS '关联 workspaces.id（工作空间上下文）';
COMMENT ON COLUMN public.audit_logs.actor_id IS '关联 users.id（操作执行人，NULL 表示系统操作）';
COMMENT ON COLUMN public.audit_logs.action IS '操作类型（如 issue.status_changed、version.released）';
COMMENT ON COLUMN public.audit_logs.target IS '操作目标标识（如工作项编号、版本号）';
COMMENT ON COLUMN public.audit_logs.detail IS '操作详情 JSON（变更前后值等）';
COMMENT ON COLUMN public.audit_logs.ip IS '操作来源 IP 地址（inet 类型）';
COMMENT ON COLUMN public.audit_logs.created_at IS '操作时间（含时区）';

-- ============================================================
-- 表: schema_migrations — 数据库迁移版本表
-- ============================================================
COMMENT ON TABLE public.schema_migrations IS '[系统基础设施域]数据库迁移版本（记录已应用的迁移编号和脏标记）';

COMMENT ON COLUMN public.schema_migrations.version IS '迁移版本号（对应迁移文件名中的数字前缀）';
COMMENT ON COLUMN public.schema_migrations.dirty IS '是否处于脏状态: true=上次迁移中途失败 / false=正常';

-- ============================================================
-- 表: project_sequences — 项目序列发号器表
-- ============================================================
COMMENT ON TABLE public.project_sequences IS '[系统基础设施域]项目工作项序列发号器（每个项目一行，原子递增生成 sequence_id）';

COMMENT ON COLUMN public.project_sequences.project_id IS '关联 projects.id（所属项目，主键）';
COMMENT ON COLUMN public.project_sequences.next_value IS '下一个序列号（从 1 开始，允许跳号）';



-- ============================================================================
-- 触发器注释（统一区块）
-- ============================================================================

-- ============================================================
-- updated_at 自动维护触发器（适用于所有含 updated_at 的表）
-- ============================================================
COMMENT ON TRIGGER trg_api_tokens_updated_at ON public.api_tokens IS '自动维护 updated_at = now()，BEFORE UPDATE 时触发';
COMMENT ON TRIGGER trg_automation_rules_updated_at ON public.automation_rules IS '自动维护 updated_at = now()，BEFORE UPDATE 时触发';
COMMENT ON TRIGGER trg_dashboard_widgets_updated_at ON public.dashboard_widgets IS '自动维护 updated_at = now()，BEFORE UPDATE 时触发';
COMMENT ON TRIGGER trg_invitations_updated_at ON public.invitations IS '自动维护 updated_at = now()，BEFORE UPDATE 时触发';
COMMENT ON TRIGGER trg_issues_updated_at ON public.issues IS '自动维护 updated_at = now()，BEFORE UPDATE 时触发';
COMMENT ON TRIGGER trg_labels_updated_at ON public.labels IS '自动维护 updated_at = now()，BEFORE UPDATE 时触发';
COMMENT ON TRIGGER trg_metric_adjustments_updated_at ON public.metric_adjustments IS '自动维护 updated_at = now()，BEFORE UPDATE 时触发';
COMMENT ON TRIGGER trg_modules_updated_at ON public.modules IS '自动维护 updated_at = now()，BEFORE UPDATE 时触发';
COMMENT ON TRIGGER trg_projects_updated_at ON public.projects IS '自动维护 updated_at = now()，BEFORE UPDATE 时触发';
COMMENT ON TRIGGER trg_risk_rules_updated_at ON public.risk_rules IS '自动维护 updated_at = now()，BEFORE UPDATE 时触发';
COMMENT ON TRIGGER trg_search_bookmarks_updated_at ON public.search_bookmarks IS '自动维护 updated_at = now()，BEFORE UPDATE 时触发';
COMMENT ON TRIGGER trg_search_documents_updated_at ON public.search_documents IS '自动维护 updated_at = now()，BEFORE UPDATE 时触发';
COMMENT ON TRIGGER trg_sprints_updated_at ON public.sprints IS '自动维护 updated_at = now()，BEFORE UPDATE 时触发';
COMMENT ON TRIGGER trg_states_updated_at ON public.states IS '自动维护 updated_at = now()，BEFORE UPDATE 时触发';
COMMENT ON TRIGGER trg_time_logs_updated_at ON public.time_logs IS '自动维护 updated_at = now()，BEFORE UPDATE 时触发';
COMMENT ON TRIGGER trg_users_updated_at ON public.users IS '自动维护 updated_at = now()，BEFORE UPDATE 时触发';
COMMENT ON TRIGGER trg_versions_updated_at ON public.versions IS '自动维护 updated_at = now()，BEFORE UPDATE 时触发';
COMMENT ON TRIGGER trg_workbench_configs_updated_at ON public.workbench_configs IS '自动维护 updated_at = now()，BEFORE UPDATE 时触发';
COMMENT ON TRIGGER trg_workspaces_updated_at ON public.workspaces IS '自动维护 updated_at = now()，BEFORE UPDATE 时触发';

-- ============================================================
-- 乐观锁版本号自增触发器
-- ============================================================
COMMENT ON TRIGGER trg_versions_bump_version ON public.versions IS '乐观锁版本号自增: NEW.version = OLD.version + 1，BEFORE UPDATE 时触发';

-- ============================================================
-- 搜索索引同步触发器
-- ============================================================
COMMENT ON TRIGGER trg_issue_search_sync ON public.issues IS '工作项插入/更新时同步到 search_documents 全文检索表（仅 deleted_at IS NULL 时）';
COMMENT ON TRIGGER trg_issue_search_cleanup ON public.issues IS '工作项软删除（deleted_at 变为非空）时清理 search_documents 对应记录';
COMMENT ON TRIGGER trg_sprint_search_sync ON public.sprints IS '迭代插入/更新时同步到 search_documents 全文检索表（仅 deleted_at IS NULL 时）';
COMMENT ON TRIGGER trg_sprint_search_cleanup ON public.sprints IS '迭代软删除（deleted_at 变为非空）时清理 search_documents 对应记录';
COMMENT ON TRIGGER trg_version_search_sync ON public.versions IS '版本插入/更新时同步到 search_documents 全文检索表（仅 deleted_at IS NULL 时）';
COMMENT ON TRIGGER trg_version_search_cleanup ON public.versions IS '版本软删除（deleted_at 变为非空）时清理 search_documents 对应记录';

-- ============================================================
-- 最近访问时间更新触发器
-- ============================================================
COMMENT ON TRIGGER trg_recent_items_touch ON public.recent_items IS '每次 UPDATE 时自动将 accessed_at 重置为 now()';

-- ============================================================================
-- 部分索引注释（说明 WHERE 条件含义）
-- ============================================================================

-- ================ issues 表索引 ================
COMMENT ON INDEX public.idx_issues_project_state IS '按项目+状态+排序查询工作项列表（WHERE deleted_at IS NULL，看板视图主索引）';
COMMENT ON INDEX public.idx_issues_list_covering IS '工作项列表排序查询（WHERE deleted_at IS NULL，按 updated_at 倒序）';
COMMENT ON INDEX public.idx_issues_project_sequence IS '工作项编号唯一性保证（WHERE deleted_at IS NULL，sequence_id 在项目内唯一）';
COMMENT ON INDEX public.idx_issues_public_id IS '按 public_id 查询工作项（WHERE deleted_at IS NULL，UUID 部分唯一索引）';
COMMENT ON INDEX public.idx_issues_parent IS '查找子工作项（WHERE deleted_at IS NULL AND parent_id IS NOT NULL，层级关系查询）';
COMMENT ON INDEX public.idx_issues_target_date IS '按目标日期查询未完成工作项（WHERE deleted_at IS NULL AND completed_at IS NULL）';
COMMENT ON INDEX public.idx_issues_target_date_covering IS '目标日期覆盖索引（WHERE deleted_at IS NULL AND target_date IS NOT NULL）';
COMMENT ON INDEX public.idx_issues_search_tsv IS '工作项全文检索 GIN 索引（基于 search_tsv 字段，ES 降级使用）';
COMMENT ON INDEX public.idx_issues_priority_covering IS '高优工作项覆盖索引（WHERE deleted_at IS NULL AND priority IN (urgent, high)，优先级筛选视图）';
COMMENT ON INDEX public.idx_issues_fix_version IS '按修复版本查询工作项（WHERE deleted_at IS NULL AND fix_version_id IS NOT NULL）';
COMMENT ON INDEX public.idx_issues_found_version IS '按发现版本查询缺陷（WHERE deleted_at IS NULL AND found_version_id IS NOT NULL）';
COMMENT ON INDEX public.idx_issues_release_version IS '按发布版本查询工作项（WHERE deleted_at IS NULL AND release_version_id IS NOT NULL）';
COMMENT ON INDEX public.idx_issues_type IS '按类型查询工作项（WHERE deleted_at IS NULL，类型筛选）';
COMMENT ON INDEX public.idx_issues_type_covering IS '按类型查询并排序（WHERE deleted_at IS NULL，类型视图）';
COMMENT ON INDEX public.idx_issues_workspace_project IS '按空间+项目查询工作项（WHERE deleted_at IS NULL，列表视图）';
COMMENT ON INDEX public.idx_issues_created IS '按创建时间倒序查询（项目内排序）';
COMMENT ON INDEX public.idx_issues_state_covering IS '看板列内按状态+排序字段查询（WHERE deleted_at IS NULL）';

-- ================ sprints 表索引 ================
COMMENT ON INDEX public.idx_one_active_sprint_per_project IS '保证每项目最多一个激活迭代（WHERE status = active AND deleted_at IS NULL）';
COMMENT ON INDEX public.idx_sprints_active_unique IS '激活迭代唯一性+按项目查询（WHERE status = active AND deleted_at IS NULL）';
COMMENT ON INDEX public.idx_sprints_project_status IS '按项目+状态查询迭代列表（WHERE deleted_at IS NULL）';
COMMENT ON INDEX public.idx_sprints_version IS '按发布版本查找关联迭代（WHERE deleted_at IS NULL）';

-- ================ versions 表索引 ================
COMMENT ON INDEX public.idx_versions_project_status IS '按项目+状态查询版本列表（WHERE deleted_at IS NULL）';
COMMENT ON INDEX public.idx_versions_unique_semver IS '同项目下 SemVer 唯一性保证（WHERE deleted_at IS NULL）';
COMMENT ON INDEX public.idx_versions_workspace IS '按空间查询版本列表（WHERE deleted_at IS NULL）';

-- ================ issues 关联表索引 ================
COMMENT ON INDEX public.idx_issue_deps_pred IS '查找指定工作项的所有后继依赖（按 predecessor_id）';
COMMENT ON INDEX public.idx_issue_deps_succ IS '查找指定工作项的所有前驱依赖（按 successor_id）';
COMMENT ON INDEX public.idx_issue_relations_source IS '查找源工作项的所有关联关系（按 source_issue_id）';
COMMENT ON INDEX public.idx_issue_relations_target IS '查找目标工作项的所有关联关系（按 target_issue_id）';
COMMENT ON INDEX public.idx_issue_assignees_user IS '查找用户负责的所有工作项（按 user_id）';
COMMENT ON INDEX public.idx_issue_watchers_user IS '查找用户关注的所有工作项（按 user_id）';
COMMENT ON INDEX public.idx_sprint_issues_issue IS '按工作项反查所在迭代（按 issue_id）';

-- ================ notification 模块索引 ================
COMMENT ON INDEX public.idx_notifications_recipient_unread IS '查询用户未读通知（WHERE is_archived = false，按 created_at 倒序）';
COMMENT ON INDEX public.idx_notifications_entity IS '按关联实体类型+ID 查找通知';
COMMENT ON INDEX public.idx_notifications_archived IS '查询归档通知（WHERE is_archived = true）';
COMMENT ON INDEX public.idx_deliveries_next_retry IS '查询待重试投递（WHERE status = pending，按 next_retry_at）';
COMMENT ON INDEX public.idx_deliveries_notification IS '按通知 ID 查找投递记录';
COMMENT ON INDEX public.idx_deliveries_status IS '按投递状态查询（WHERE status = pending）';
COMMENT ON INDEX public.idx_digests_pending IS '查询待发送的聚合摘要（WHERE status = pending，按 scheduled_for）';

-- ================ search 模块索引 ================
COMMENT ON INDEX public.idx_search_documents_tsv IS '搜索文档 GIN 全文检索索引';
COMMENT ON INDEX public.idx_search_documents_unique IS '搜索文档唯一性保证（workspace_id + doc_type + doc_id）';
COMMENT ON INDEX public.idx_search_documents_workspace IS '按空间+类型查找搜索文档';
COMMENT ON INDEX public.idx_search_documents_project IS '按项目+类型查找搜索文档';
COMMENT ON INDEX public.idx_search_history_user IS '按用户查找搜索历史（按 searched_at 倒序）';

-- ================ 其他模块索引 ================
COMMENT ON INDEX public.idx_attachments_entity IS '按实体类型+ID 查找附件（WHERE deleted_at IS NULL）';
COMMENT ON INDEX public.idx_attachments_uploader IS '按上传者查找附件（WHERE deleted_at IS NULL）';
COMMENT ON INDEX public.idx_audit_logs_ws_time IS '按空间+时间查询审计日志';
COMMENT ON INDEX public.idx_time_logs_issue IS '按工作项查询工时记录（WHERE deleted_at IS NULL）';
COMMENT ON INDEX public.idx_time_logs_user_date IS '按用户+日期查询工时记录（WHERE deleted_at IS NULL）';
COMMENT ON INDEX public.idx_risk_alerts_project IS '按项目查询未解决风险告警（WHERE NOT is_resolved）';
COMMENT ON INDEX public.idx_risk_alerts_unresolved IS '按空间+严重度查询未解决告警（WHERE NOT is_resolved）';
COMMENT ON INDEX public.idx_recent_items_user IS '按用户查询最近访问记录（按 accessed_at 倒序）';
COMMENT ON INDEX public.idx_intake_issues_channel IS '按渠道查询入口工单（按 created_at 倒序）';
COMMENT ON INDEX public.idx_intake_issues_status IS '按空间+状态查询入口工单';
COMMENT ON INDEX public.idx_events_unpublished IS '查询未发布领域事件（WHERE published_at IS NULL，Outbox 轮询）';
COMMENT ON INDEX public.idx_password_reset_tokens_expires IS '按过期时间查询令牌（用于清理任务）';
COMMENT ON INDEX public.idx_password_reset_tokens_user_active IS '查找用户有效重置令牌（WHERE used_at IS NULL，唯一约束）';
COMMENT ON INDEX public.idx_deployment_events_project IS '按项目查询部署事件（按 deployed_at 倒序）';
COMMENT ON INDEX public.idx_deployment_events_ws IS '按空间查询部署事件';

-- ============================================================================
-- 全部对象注释注入完毕。补丁执行完毕。
-- ============================================================================

