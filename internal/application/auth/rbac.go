// Package auth — RBAC 领域模型（角色枚举 + 权限码常量）。
//
// 设计对标行业主流竞品：
//   GitLab  = Instance Admin（系统） > Group Owner（空间） > Maintainer > Developer > Reporter > Guest
//   Jira    = Jira Administrator（系统） > Project Admin（空间） > 各项角色
//   Plane   = Admin > Workspace Owner > Member > Guest
//   ONES    = 系统管理员 > 空间管理员 > 普通成员
//
// 关键分层（自顶向下）：
//   admin（系统管理员）= 平台级 / L5 / level=100
//   owner（空间管理员）= 工作空间级 / L4 / level=80
//   pm / po / techlead / qalead = L3 / level=50
//   dev = L2 / level=30
//   guest = L1 / level=10

package auth

import (
	"context"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/njydsz/ydsz-plane/pkg/errs"
)

/* ------------------------------------------------------------------ */
/* 权限常量（~55 个）                                                   */
/* ------------------------------------------------------------------ */

const (
	// ===== 系统级（仅 admin 持有，owner 不可见）=====
	PermSystemConfig        = "system:config"         // 系统通用配置：SMTP / SSO / 注册开关 / 安全策略 / 许可证
	PermSystemUserRead      = "system:user:read"      // 查看平台所有用户
	PermSystemUserManage    = "system:user:manage"    // 创建 / 禁用 / 重置密码任意平台用户
	PermSystemWorkspaceList = "system:workspace:list" // 列出所有工作空间
	PermSystemWorkspaceMgmt = "system:workspace:manage" // 归档 / 删除 / 转移工作空间所有权
	PermSystemAuditRead     = "system:audit:read"     // 全平台审计日志

	// ===== 工作空间级 =====
	PermWorkspaceRead      = "workspace:read"
	PermWorkspaceUpdate    = "workspace:update"
	PermWorkspaceDelete    = "workspace:delete"
	PermWorkspaceTransfer  = "workspace:transfer"   // 转移工作空间所有权

	// 项目
	PermProjectRead   = "project:read"
	PermProjectCreate = "project:create"
	PermProjectUpdate = "project:update"
	PermProjectDelete = "project:delete"

	// 工作项
	PermIssueRead           = "issue:read"
	PermIssueCreate         = "issue:create"
	PermIssueEditOwn        = "issue:edit_own"
	PermIssueEditAll        = "issue:edit_all"
	PermIssueDelete         = "issue:delete"
	PermIssueTransition     = "issue:transition"
	PermIssueReassign       = "issue:reassign"
	PermIssueChangePriority = "issue:change_priority"
	PermIssueManageSprint   = "issue:manage_sprint"

	// 成员
	PermMemberInvite     = "member:invite"
	PermMemberRemove     = "member:remove"
	PermMemberChangeRole = "member:change_role"

	// 迭代
	PermSprintRead      = "sprint:read"
	PermSprintCreate    = "sprint:create"
	PermSprintUpdate    = "sprint:update"
	PermSprintDelete    = "sprint:delete"
	PermSprintLifecycle = "sprint:lifecycle"
	PermSprintPlan      = "sprint:plan"

	// 版本
	PermVersionRead    = "version:read"
	PermVersionCreate  = "version:create"
	PermVersionUpdate  = "version:update"
	PermVersionRelease = "version:release"
	PermVersionDelete  = "version:delete"

	// 质量
	PermDefectCreate = "defect:create"
	PermQAReport     = "qa:report"

	// 效能
	PermAnalyticsRead   = "analytics:read"
	PermAnalyticsExport = "analytics:export"

	// 自动化 / 集成
	PermAutomationManage = "automation:manage"
	PermDeployReport     = "deploy:report"

	// 审计 / 收件箱 / 知识库 / Webhook
	PermAuditRead     = "audit:read"
	PermWebhookManage = "webhook:manage"
	PermIntakeManage  = "intake:manage"
	PermPagesManage   = "pages:manage"

	// 评论 / 关联
	PermCommentModerate = "comment:moderate"
	PermRelationManage  = "relation:manage"

	// 字段级
	PermFieldEditSeverity = "field:edit_severity"
	PermFieldEditEffort   = "field:edit_effort"
	PermFieldEditDeadline = "field:edit_deadline"

	// 菜单级
	PermMenuSettings = "menu:settings"
	PermMenuAudit    = "menu:audit"
)

/* ------------------------------------------------------------------ */
/* 角色枚举（8 个）                                                     */
/* ------------------------------------------------------------------ */

type WorkspaceRole string

const (
	// RoleAdmin 系统管理员 — 平台级最高权限，可进入并管理所有工作空间。
	RoleAdmin WorkspaceRole = "admin"
	// RoleOwner 空间管理员 — 某个工作空间内的最高权限，可管理空间下所有资源。
	RoleOwner WorkspaceRole = "owner"
	// RolePM 项目经理
	RolePM WorkspaceRole = "pm"
	// RolePO 产品经理
	RolePO WorkspaceRole = "po"
	// RoleTechLead 技术经理
	RoleTechLead WorkspaceRole = "techlead"
	// RoleQALead 测试经理
	RoleQALead WorkspaceRole = "qalead"
	// RoleDev 开发（前端/后端统一）
	RoleDev WorkspaceRole = "dev"
	// RoleGuest 访客（只读）
	RoleGuest WorkspaceRole = "guest"
)

func (r WorkspaceRole) IsValid() bool {
	switch r {
	case RoleAdmin, RoleOwner, RolePM, RolePO, RoleTechLead, RoleQALead, RoleDev, RoleGuest:
		return true
	default:
		return false
	}
}

func (r WorkspaceRole) String() string { return string(r) }

// Level 返回角色层级数值（用于前端比大小）。
//
//	admin          100  系统管理员（平台级）
//	owner           80  空间管理员（空间级最高）
//	pm/po/techlead  50  经理级
//	dev             30  执行者
//	guest           10  只读协作者
func (r WorkspaceRole) Level() int {
	switch r {
	case RoleAdmin:
		return 100
	case RoleOwner:
		return 80
	case RolePM, RolePO, RoleTechLead, RoleQALead:
		return 50
	case RoleDev:
		return 30
	case RoleGuest:
		return 10
	default:
		return 0
	}
}

// IsAtLeast 报告角色 r 是否满足最低所需角色级别 min。
func (r WorkspaceRole) IsAtLeast(min WorkspaceRole) bool {
	return r.Level() >= min.Level()
}

/* ------------------------------------------------------------------ */
/* DB-backed WorkspaceMembershipStore                                   */
/* ------------------------------------------------------------------ */

// WorkspaceMembership 记录用户与工作空间的关系。
type WorkspaceMembership struct {
	WorkspaceID int64
	UserID      int64
	Role        WorkspaceRole
	IsActive    bool
	JoinedAt    string
}

// WorkspaceMembershipStore 从数据库解析 工作空间 → 角色 查询。
type WorkspaceMembershipStore struct {
	db *pgxpool.Pool
}

func NewWorkspaceMembershipStore(db *pgxpool.Pool) *WorkspaceMembershipStore {
	return &WorkspaceMembershipStore{db: db}
}

// ResolveRole 返回用户在某工作空间的成员关系；无关系 / is_active=false 时返回 ErrForbidden。
func (s *WorkspaceMembershipStore) ResolveRole(ctx context.Context, wsID, userID int64) (WorkspaceMembership, error) {
	var (
		role     string
		joinedAt string
		isActive bool
	)
	err := s.db.QueryRow(ctx, `
		SELECT role, joined_at::text, is_active
		FROM workspace_members
		WHERE workspace_id = $1 AND user_id = $2`, wsID, userID).Scan(&role, &joinedAt, &isActive)
	if err != nil {
		return WorkspaceMembership{}, errs.ErrForbidden
	}
	if !isActive {
		return WorkspaceMembership{}, errs.ErrForbidden
	}
	return WorkspaceMembership{
		WorkspaceID: wsID,
		UserID:      userID,
		Role:        WorkspaceRole(strings.ToLower(role)),
		IsActive:    isActive,
		JoinedAt:    joinedAt,
	}, nil
}
