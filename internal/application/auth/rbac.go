// Package auth — RBAC 领域模型（角色枚举 + 权限码常量）。
//
// 设计参考：GitHub / GitLab / Plane / Linear 工作空间成员模型。
// "谁能做什么"的事实来源之一是所有 PermXxx 常量；真正的矩阵由 DB role_permissions 表承载。
//
// 与旧版本的兼容性：
//   - 角色枚举 WorkspaceRole 保持不变（owner/admin/member/guest 已保留；新增 pm/po/techlead/qalead/dev）
//   - 权限点常量从 ~18 个扩展到 ~50 个，用于 DB role_permissions 表的 permission_code 列
//   - 旧 HasPermission 方法仍然可用（走 DB 查询），便于旧调用方平滑过渡
package auth

import (
	"context"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/njydsz/ydsz-plane/pkg/errs"
)

/* ------------------------------------------------------------------ */
/* 权限常量（50 个）                                                    */
/* ------------------------------------------------------------------ */

const (
	// 工作空间
	PermWorkspaceRead   = "workspace:read"
	PermWorkspaceUpdate = "workspace:update"
	PermWorkspaceDelete = "workspace:delete"

	// 项目
	PermProjectRead   = "project:read"
	PermProjectCreate = "project:create"
	PermProjectUpdate = "project:update"
	PermProjectDelete = "project:delete"

	// 工作项
	PermIssueRead          = "issue:read"
	PermIssueCreate        = "issue:create"
	PermIssueEditOwn       = "issue:edit_own"
	PermIssueEditAll       = "issue:edit_all"
	PermIssueDelete        = "issue:delete"
	PermIssueTransition    = "issue:transition"
	PermIssueReassign      = "issue:reassign"
	PermIssueChangePriority = "issue:change_priority"
	PermIssueManageSprint  = "issue:manage_sprint"

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

	// 审计 / 收件箱 / 知识库
	PermAuditRead       = "audit:read"
	PermWebhookManage   = "webhook:manage"
	PermIntakeManage    = "intake:manage"
	PermPagesManage     = "pages:manage"

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
	RoleOwner    WorkspaceRole = "owner"
	RoleAdmin    WorkspaceRole = "admin"
	RolePM       WorkspaceRole = "pm"
	RolePO       WorkspaceRole = "po"
	RoleTechLead WorkspaceRole = "techlead"
	RoleQALead   WorkspaceRole = "qalead"
	RoleDev      WorkspaceRole = "dev"
	RoleGuest    WorkspaceRole = "guest"
)

func (r WorkspaceRole) IsValid() bool {
	switch r {
	case RoleOwner, RoleAdmin, RolePM, RolePO, RoleTechLead, RoleQALead, RoleDev, RoleGuest:
		return true
	default:
		return false
	}
}

func (r WorkspaceRole) String() string { return string(r) }

// Level 返回角色层级数值（用于前端比大小）。
func (r WorkspaceRole) Level() int {
	switch r {
	case RoleOwner:
		return 100
	case RoleAdmin:
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
