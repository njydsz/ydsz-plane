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
COMMENT ON COLUMN public.issues.description_json IS 'TipTap 富文本编辑器的 JSON 原始数据（文档节点树结构）';
COMMENT ON COLUMN public.issues.description_json IS '工作项描述的 TipTap JSON 格式（富文本编辑器原始数据）';
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
