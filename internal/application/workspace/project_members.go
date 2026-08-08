// Package workspace — 项目级成员管理。
//
// 项目成员是工作空间成员的子集，拥有项目级角色（admin / member）。
// workspace_members 控制"能否进入空间"，project_members 控制"能否进入特定项目"。
// 项目可见性为 public 时，空间成员自动获得项目读权限（不再需要显式加入）；
// project_members 用于 private/internal 项目的成员白名单，以及项目级 admin 的维护。
package workspace

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// ProjectMember 项目成员信息（API 响应）。
type ProjectMember struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Email     string    `json:"email"`
	DisplayName string  `json:"display_name"`
	AvatarURL string    `json:"avatar_url,omitempty"`
	Role      string    `json:"role"`
	JoinedAt  time.Time `json:"joined_at"`
	CreatedBy int64     `json:"created_by,omitempty"`
}

// ProjectMemberService 项目成员管理服务。
type ProjectMemberService struct {
	db *pgxpool.Pool
}

// NewProjectMemberService 创建项目成员服务。
func NewProjectMemberService(db *pgxpool.Pool) *ProjectMemberService {
	return &ProjectMemberService{db: db}
}

// List 列出项目内的全部成员。
func (s *ProjectMemberService) List(ctx context.Context, wsID, projectID int64) ([]ProjectMember, error) {
	rows, err := s.db.Query(ctx, `
		SELECT pm.id, pm.user_id, u.email, u.display_name, coalesce(u.avatar_url,''),
		       pm.role, pm.joined_at, coalesce(pm.created_by, 0)
		FROM project_members pm
		JOIN users u ON u.id = pm.user_id
		WHERE pm.workspace_id = $1 AND pm.project_id = $2 AND u.is_active
		ORDER BY
			CASE pm.role WHEN 'admin' THEN 0 ELSE 1 END,
			pm.joined_at ASC`, wsID, projectID)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()

	out := make([]ProjectMember, 0)
	for rows.Next() {
		var m ProjectMember
		if err := rows.Scan(&m.ID, &m.UserID, &m.Email, &m.DisplayName, &m.AvatarURL,
			&m.Role, &m.JoinedAt, &m.CreatedBy); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
		out = append(out, m)
	}
	return out, nil
}

// AddMember 将工作空间成员加入项目。
// 前置条件：调用者须为项目 admin 或 workspace admin/owner；targetUserID 须为同空间成员。
func (s *ProjectMemberService) AddMember(ctx context.Context, wsID, projectID, targetUserID, adderID int64, role string) error {
	if role != "admin" && role != "member" {
		return errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "role", Reason: "无效的项目角色"})
	}
	// 校验 adder 是否有权限加入（是 workspace owner/admin 或 project admin）
	if err := s.checkProjectManagePermission(ctx, wsID, projectID, adderID); err != nil {
		return err
	}
	// 校验 target 是否为同 workspace 成员（public/internal 项目：workspace 成员可直接加入；
	// private 项目：只有 workspace 成员才可被加为项目成员）
	var wsRole string
	err := s.db.QueryRow(ctx, `SELECT role FROM workspace_members WHERE workspace_id = $1 AND user_id = $2`,
		wsID, targetUserID).Scan(&wsRole)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errs.New("PROJECT_MEMBER.NOT_WS_MEMBER", "目标用户不是该工作空间的成员", 403)
		}
		return errs.ErrInternal.Wrap(err)
	}
	// 校验是否已是项目成员
	var existingID int64
	err = s.db.QueryRow(ctx, `SELECT id FROM project_members WHERE workspace_id = $1 AND project_id = $2 AND user_id = $3`,
		wsID, projectID, targetUserID).Scan(&existingID)
	if err == nil {
		return errs.New("PROJECT_MEMBER.ALREADY_EXISTS", "该用户已是项目成员", 409)
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return errs.ErrInternal.Wrap(err)
	}

	tag, err := s.db.Exec(ctx, `
		INSERT INTO project_members (workspace_id, project_id, user_id, role, created_by)
		VALUES ($1, $2, $3, $4, $5)`,
		wsID, projectID, targetUserID, role, adderID)
	if err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	if tag.RowsAffected() == 0 {
		return errs.ErrInternal
	}
	return nil
}

// ChangeRole 修改项目成员角色。
func (s *ProjectMemberService) ChangeRole(ctx context.Context, wsID, projectID, targetUserID int64, newRole string) error {
	if newRole != "admin" && newRole != "member" {
		return errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "role", Reason: "无效的项目角色"})
	}
	tag, err := s.db.Exec(ctx, `
		UPDATE project_members SET role = $1, updated_at = now()
		WHERE workspace_id = $2 AND project_id = $3 AND user_id = $4`,
		newRole, wsID, projectID, targetUserID)
	if err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	if tag.RowsAffected() == 0 {
		return errs.ErrNotFound
	}
	return nil
}

// RemoveMember 从项目中移除成员。
func (s *ProjectMemberService) RemoveMember(ctx context.Context, wsID, projectID, targetUserID int64) error {
	tag, err := s.db.Exec(ctx, `
		DELETE FROM project_members
		WHERE workspace_id = $1 AND project_id = $2 AND user_id = $3`,
		wsID, projectID, targetUserID)
	if err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	if tag.RowsAffected() == 0 {
		return errs.ErrNotFound
	}
	return nil
}

// GetProjectRole 返回用户在该项目中的角色（若未显式加入则为空字符串）。
// 用于前端判断"此人是否有项目级管理权限（admin）"。
func (s *ProjectMemberService) GetProjectRole(ctx context.Context, wsID, projectID, userID int64) (string, error) {
	var role string
	err := s.db.QueryRow(ctx, `SELECT role FROM project_members WHERE workspace_id = $1 AND project_id = $2 AND user_id = $3`,
		wsID, projectID, userID).Scan(&role)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", nil
		}
		return "", errs.ErrInternal.Wrap(err)
	}
	return role, nil
}

// AddCreatorAsAdmin 将项目创建者加入为项目 admin（创建项目时调用）。
func (s *ProjectMemberService) AddCreatorAsAdmin(ctx context.Context, wsID, projectID, creatorID int64) error {
	_, err := s.db.Exec(ctx, `
		INSERT INTO project_members (workspace_id, project_id, user_id, role, created_by)
		VALUES ($1, $2, $3, 'admin', $3)
		ON CONFLICT (workspace_id, project_id, user_id) DO UPDATE SET role = 'admin'`,
		wsID, projectID, creatorID)
	if err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	return nil
}

// --- Internal helpers ---

// checkProjectManagePermission 检查用户是否具备该项目的"成员管理"权限。
// 权限规则：workspace owner/admin 或 project admin 可管理项目成员。
func (s *ProjectMemberService) checkProjectManagePermission(ctx context.Context, wsID, projectID, userID int64) error {
	var wsRole string
	err := s.db.QueryRow(ctx, `SELECT role FROM workspace_members WHERE workspace_id = $1 AND user_id = $2`,
		wsID, userID).Scan(&wsRole)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errs.ErrForbidden
		}
		return errs.ErrInternal.Wrap(err)
	}
	if wsRole == "owner" || wsRole == "admin" {
		return nil
	}
	// workspace 角色较低时，检查是否 project admin
	var projectRole string
	err = s.db.QueryRow(ctx, `SELECT role FROM project_members WHERE workspace_id = $1 AND project_id = $2 AND user_id = $3`,
		wsID, projectID, userID).Scan(&projectRole)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errs.ErrForbidden
		}
		return errs.ErrInternal.Wrap(err)
	}
	if projectRole != "admin" {
		return errs.ErrForbidden
	}
	return nil
}
