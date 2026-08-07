// Package auth — RBAC domain model (role enum + permission matrix).
//
// Design reference: GitHub / GitLab workspace membership model.
// Single source of truth for "who can do what" lives in Roles[role].
package auth

import (
	"context"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/njydsz/ydsz-plane/pkg/errs"
)

/* ------------------------------------------------------------------ */
/* Permission constants                                                 */
/* ------------------------------------------------------------------ */

const (
	// Workspace management
	PermWorkspaceRead   = "workspace:read"
	PermWorkspaceUpdate = "workspace:update"
	PermWorkspaceDelete = "workspace:delete"
	// Membership management
	PermMemberInvite    = "member:invite"
	PermMemberRemove    = "member:remove"
	PermMemberChangeRole = "member:change_role"
	// Project management
	PermProjectCreate = "project:create"
	PermProjectDelete = "project:delete"
	// Issue management
	PermIssueCreate = "issue:create"
	PermIssueDelete = "issue:delete"
)

/* ------------------------------------------------------------------ */
/* Role enum + ordering                                                 */
/* ------------------------------------------------------------------ */

// WorkspaceRole is a membership level.
type WorkspaceRole string

const (
	RoleOwner  WorkspaceRole = "owner"
	RoleAdmin  WorkspaceRole = "admin"
	RoleMember WorkspaceRole = "member"
	RoleGuest  WorkspaceRole = "guest"
)

func (r WorkspaceRole) IsValid() bool {
	switch r {
	case RoleOwner, RoleAdmin, RoleMember, RoleGuest:
		return true
	default:
		return false
	}
}

// IsAtLeast returns whether role r satisfies a minimum required role level.
func (r WorkspaceRole) IsAtLeast(min WorkspaceRole) bool {
	levels := map[WorkspaceRole]int{
		RoleGuest: 0, RoleMember: 1, RoleAdmin: 2, RoleOwner: 3,
	}
	return levels[r] >= levels[min]
}

/* ------------------------------------------------------------------ */
/* Permission matrix (single source of truth)                           */
/* ------------------------------------------------------------------ */

// Roles maps each role to the set of granted permissions.
var Roles = map[WorkspaceRole][]string{
	RoleOwner: {
		PermWorkspaceRead, PermWorkspaceUpdate, PermWorkspaceDelete,
		PermMemberInvite, PermMemberRemove, PermMemberChangeRole,
		PermProjectCreate, PermProjectDelete,
		PermIssueCreate, PermIssueDelete,
	},
	RoleAdmin: {
		PermWorkspaceRead, PermWorkspaceUpdate,
		PermMemberInvite, PermMemberRemove,
		PermProjectCreate, PermProjectDelete,
		PermIssueCreate, PermIssueDelete,
	},
	RoleMember: {
		PermWorkspaceRead,
		PermProjectCreate,
		PermIssueCreate,
	},
	RoleGuest: {
		PermWorkspaceRead,
	},
}

// RolePermissionSet returns the set for O(1) lookup (cached per call).
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

// WorkspaceMembership captures the user's relation to a workspace.
type WorkspaceMembership struct {
	WorkspaceID int64
	UserID      int64
	Role        WorkspaceRole
	JoinedAt    string
}

// HasPermission checks if a role carries the given permission.
func (m WorkspaceMembership) HasPermission(perm string) bool {
	_, ok := RolePermissionSet(m.Role)[perm]
	return ok
}

// WorkspaceMembershipStore resolves workspace → role lookups from DB.
type WorkspaceMembershipStore struct {
	db *pgxpool.Pool
}

// NewWorkspaceMembershipStore constructs the store.
func NewWorkspaceMembershipStore(db *pgxpool.Pool) *WorkspaceMembershipStore {
	return &WorkspaceMembershipStore{db: db}
}

// ResolveRole returns the user's membership in a workspace, or an ErrForbidden
// / ErrNotFound AppError (abstracted from the 404 vs 403 distinction to avoid
// leaking workspace existence to non-members).
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
		// Hide workspace existence: non-members see 403 (ErrForbidden)
		return WorkspaceMembership{}, errs.ErrForbidden
	}
	return WorkspaceMembership{
		WorkspaceID: wsID,
		UserID:      userID,
		Role:        WorkspaceRole(strings.ToLower(role)),
		JoinedAt:    joinedAt,
	}, nil
}
