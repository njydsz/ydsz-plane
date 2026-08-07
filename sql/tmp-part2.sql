
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
