// Package workspace — 工作空间成员管理。
//
// 职责：成员列表查询、角色变更、移除成员。
package workspace

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// Member 工作空间成员信息（API 响应）。
type Member struct {
	ID          int64     `json:"id"`
	Email       string    `json:"email"`
	DisplayName string    `json:"display_name"`
	AvatarURL   string    `json:"avatar_url,omitempty"`
	Role        string    `json:"role"`
	JoinedAt    time.Time `json:"joined_at"`
}

// MemberService 成员管理服务。
type MemberService struct {
	db *pgxpool.Pool
}

// NewMemberService 创建成员服务。
func NewMemberService(db *pgxpool.Pool) *MemberService {
	return &MemberService{db: db}
}

// List 列出工作空间内的全部成员。
func (s *MemberService) List(ctx context.Context, wsID int64) ([]Member, error) {
	rows, err := s.db.Query(ctx, `
		SELECT u.id, u.email, u.display_name, coalesce(u.avatar_url,''), wm.role, wm.joined_at
		FROM workspace_members wm
		JOIN users u ON u.id = wm.user_id
		WHERE wm.workspace_id = $1 AND u.is_active
		ORDER BY
			CASE wm.role WHEN 'owner' THEN 0 WHEN 'admin' THEN 1 WHEN 'member' THEN 2 ELSE 3 END,
			wm.joined_at ASC`, wsID)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()

	var out []Member
	for rows.Next() {
		var m Member
		if err := rows.Scan(&m.ID, &m.Email, &m.DisplayName, &m.AvatarURL, &m.Role, &m.JoinedAt); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
		out = append(out, m)
	}
	return out, nil
}

// ChangeRole 修改成员角色（仅 owner/admin 可操作；不可改 owner 自身）。
func (s *MemberService) ChangeRole(ctx context.Context, wsID, targetUserID int64, newRole string) error {
	if newRole != "admin" && newRole != "member" && newRole != "guest" {
		return errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "role", Reason: "无效的角色"})
	}
	tag, err := s.db.Exec(ctx, `
		UPDATE workspace_members SET role = $1
		WHERE workspace_id = $2 AND user_id = $3 AND role <> 'owner'`,
		newRole, wsID, targetUserID)
	if err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	if tag.RowsAffected() == 0 {
		var curRole string
		err := s.db.QueryRow(ctx, `SELECT role FROM workspace_members WHERE workspace_id = $1 AND user_id = $2`, wsID, targetUserID).Scan(&curRole)
		if errors.Is(err, pgx.ErrNoRows) {
			return errs.ErrNotFound
		}
		if curRole == "owner" {
			return errs.New("MEMBER.CANNOT_CHANGE_OWNER", "不可修改 Owner 的角色", 403)
		}
		return errs.ErrForbidden
	}
	return nil
}

// RemoveMember 移除成员（仅 owner/admin 可操作；不可移除 owner）。
func (s *MemberService) RemoveMember(ctx context.Context, wsID, targetUserID int64) error {
	tag, err := s.db.Exec(ctx, `
		DELETE FROM workspace_members
		WHERE workspace_id = $1 AND user_id = $2 AND role <> 'owner'`,
		wsID, targetUserID)
	if err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	if tag.RowsAffected() == 0 {
		var curRole string
		err := s.db.QueryRow(ctx, `SELECT role FROM workspace_members WHERE workspace_id = $1 AND user_id = $2`, wsID, targetUserID).Scan(&curRole)
		if errors.Is(err, pgx.ErrNoRows) {
			return errs.ErrNotFound
		}
		if curRole == "owner" {
			return errs.New("MEMBER.CANNOT_REMOVE_OWNER", "不可移除工作空间 Owner", 403)
		}
		return errs.ErrForbidden
	}
	return nil
}

// Count 返回工作空间成员数。
func (s *MemberService) Count(ctx context.Context, wsID int64) (int64, error) {
	var n int64
	err := s.db.QueryRow(ctx, `SELECT count(*) FROM workspace_members WHERE workspace_id = $1`, wsID).Scan(&n)
	return n, err
}

// IsMember 判断用户是否为该工作空间成员。
func (s *MemberService) IsMember(ctx context.Context, wsID, userID int64) (bool, error) {
	var role string
	err := s.db.QueryRow(ctx, `SELECT role FROM workspace_members WHERE workspace_id = $1 AND user_id = $2`, wsID, userID).Scan(&role)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
