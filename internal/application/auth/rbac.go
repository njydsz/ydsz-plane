// Package auth — RBAC 领域模型（角色枚举 + 权限矩阵）。
//
// 设计参考：GitHub / GitLab 工作空间成员模型。
// "谁能做什么"的唯一事实来源是 Roles[role] 映射。
package auth

import (
	"context"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/njydsz/ydsz-plane/pkg/errs"
)

/* ------------------------------------------------------------------ */
/* 权限常量                                                              */
/* ------------------------------------------------------------------ */

const (
	// 工作空间管理
	PermWorkspaceRead   = "workspace:read"
	PermWorkspaceUpdate = "workspace:update"
	PermWorkspaceDelete = "workspace:delete"
	// 成员管理
	PermMemberInvite    = "member:invite"
	PermMemberRemove    = "member:remove"
	PermMemberChangeRole = "member:change_role"
	// 项目管理
	PermProjectCreate = "project:create"
	PermProjectDelete = "project:delete"
	// 审计
	PermAuditRead = "audit:read"
	// 工作项管理
	PermIssueCreate = "issue:create"
	PermIssueDelete = "issue:delete"
	// 版本管理
	PermVersionCreate = "version:create"
	PermVersionRelease = "version:release"
	PermVersionDelete = "version:delete"
	PermVersionUpdate = "version:update"
	// 自动化管理（S11）
	PermProjectAutomation = "project:automation"
)

/* ------------------------------------------------------------------ */
/* 角色枚举与排序                                                         */
/* ------------------------------------------------------------------ */

// WorkspaceRole 是成员级别。
type WorkspaceRole string

const (
	RoleOwner  WorkspaceRole = "owner"
	RoleAdmin  WorkspaceRole = "admin"
	RoleMember WorkspaceRole = "member"
	RoleGuest  WorkspaceRole = "guest"
)

// IsValid 报告角色是否为合法枚举值。
func (r WorkspaceRole) IsValid() bool {
	switch r {
	case RoleOwner, RoleAdmin, RoleMember, RoleGuest:
		return true
	default:
		return false
	}
}

// IsAtLeast 报告角色 r 是否满足最低所需角色级别 min。
func (r WorkspaceRole) IsAtLeast(min WorkspaceRole) bool {
	levels := map[WorkspaceRole]int{
		RoleGuest: 0, RoleMember: 1, RoleAdmin: 2, RoleOwner: 3,
	}
	return levels[r] >= levels[min]
}

/* ------------------------------------------------------------------ */
/* 权限矩阵（唯一事实来源）                                                 */
/* ------------------------------------------------------------------ */

// Roles 将每个角色映射到其被授予的权限集合。
var Roles = map[WorkspaceRole][]string{
	RoleOwner: {
		PermWorkspaceRead, PermWorkspaceUpdate, PermWorkspaceDelete,
		PermMemberInvite, PermMemberRemove, PermMemberChangeRole,
		PermProjectCreate, PermProjectDelete,
		PermAuditRead,
		PermIssueCreate, PermIssueDelete,
		PermVersionCreate, PermVersionRelease, PermVersionDelete, PermVersionUpdate,
		PermProjectAutomation,
	},
	RoleAdmin: {
		PermWorkspaceRead, PermWorkspaceUpdate,
		PermMemberInvite, PermMemberRemove,
		PermProjectCreate, PermProjectDelete,
		PermAuditRead,
		PermIssueCreate, PermIssueDelete,
		PermVersionCreate, PermVersionRelease, PermVersionDelete, PermVersionUpdate,
		PermProjectAutomation,
	},
	RoleMember: {
		PermWorkspaceRead,
		PermProjectCreate,
		PermIssueCreate,
		PermVersionCreate, PermVersionUpdate,
	},
	RoleGuest: {
		PermWorkspaceRead,
	},
}

// RolePermissionSet 返回权限集合，便于 O(1) 查找（每次调用构建）。
func RolePermissionSet(role WorkspaceRole) map[string]struct{} {
	set := make(map[string]struct{}, len(Roles[role]))
	for _, p := range Roles[role] {
		set[p] = struct{}{}
	}
	return set
}

/* ------------------------------------------------------------------ */
/* WorkspaceMembership                                                  */
/* ------------------------------------------------------------------ */

// WorkspaceMembership 记录用户与工作空间的关系。
type WorkspaceMembership struct {
	WorkspaceID int64
	UserID      int64
	Role        WorkspaceRole
	JoinedAt    string
}

// HasPermission 检查某角色是否携带指定权限。
func (m WorkspaceMembership) HasPermission(perm string) bool {
	_, ok := RolePermissionSet(m.Role)[perm]
	return ok
}

// WorkspaceMembershipStore 从数据库解析 工作空间 → 角色 查询。
type WorkspaceMembershipStore struct {
	db *pgxpool.Pool
}

// NewWorkspaceMembershipStore 构造该 store。
func NewWorkspaceMembershipStore(db *pgxpool.Pool) *WorkspaceMembershipStore {
	return &WorkspaceMembershipStore{db: db}
}

// ResolveRole 返回用户在某工作空间的成员关系；无关系时返回 ErrForbidden /
// ErrNotFound 的 AppError（抽象 404 与 403 的区别，
// 避免向非成员泄露工作空间是否存在）。
func (s *WorkspaceMembershipStore) ResolveRole(ctx context.Context, wsID, userID int64) (WorkspaceMembership, error) {
	var (
		role     string
		joinedAt string
	)
	err := s.db.QueryRow(ctx, `
		SELECT role, joined_at::text
		FROM workspace_members
		WHERE workspace_id = $1 AND user_id = $2`, wsID, userID).Scan(&role, &joinedAt)
	if err != nil {
		// 隐藏工作空间存在性：非成员统一看到 403（ErrForbidden）
		return WorkspaceMembership{}, errs.ErrForbidden
	}
	return WorkspaceMembership{
		WorkspaceID: wsID,
		UserID:      userID,
		Role:        WorkspaceRole(strings.ToLower(role)),
		JoinedAt:    joinedAt,
	}, nil
}
