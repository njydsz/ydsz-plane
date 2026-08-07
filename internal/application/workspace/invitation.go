// Package workspace — 邀请管理。
//
// 设计参考: GitHub / GitLab invitations workspace model。
// Token 仅以 SHA-256 hash 落库；原始 token 只在邮件中出现一次。
// 邀请状态：pending → accepted / revoked / expired（过期由查询过滤处理）。
package workspace

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/njydsz/ydsz-plane/internal/infrastructure/mail"
	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// Invitation 表示一条邀请记录（API 响应 DTO）。
	ID          int64      `json:"id"`
	WorkspaceID int64      `json:"workspace_id"`
	InviterID   int64      `json:"inviter_id"`
	InviterName string     `json:"inviter_name,omitempty"`
	Email       string     `json:"email"`
	Role        string     `json:"role"`
	Status      string     `json:"status"`
	Message     string     `json:"message,omitempty"`
	ExpiresAt   time.Time  `json:"expires_at"`
	AcceptedAt  *time.Time `json:"accepted_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

// InviteInput 发送邀请的入参。
type InviteInput struct {
	WorkspaceID int64
	InviterID   int64
	Email       string
	Role        string
	Message     string
}

// InvitationService 管理邀请的生命周期。
type InvitationService struct {
	db      *pgxpool.Pool
	mailer  mail.EmailService
	baseURL string
}

// NewInvitationService 创建邀请服务。
func NewInvitationService(db *pgxpool.Pool, mailer mail.EmailService, baseURL string) *InvitationService {
	return &InvitationService{db: db, mailer: mailer, baseURL: baseURL}
}

// Invite 创建邀请并发送邀请邮件。
func (s *InvitationService) Invite(ctx context.Context, in InviteInput) (*Invitation, error) {
	in.Email = strings.TrimSpace(strings.ToLower(in.Email))
	if in.Email == "" {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "email", Reason: "邮箱不能为空"})
	}
	if in.Role != "admin" && in.Role != "member" && in.Role != "guest" {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "role", Reason: "无效的角色"})
	}

	// 生成 token（仅此时有原始值，hash 落库）
	rawToken := generateSecureToken()
	hash := sha256Hex(rawToken)

	var inv *Invitation
	var inviterName, wsName string
	err := pgx.BeginTxFunc(ctx, s.db, pgx.TxOptions{}, func(tx pgx.Tx) error {
		// 撤销该邮箱下已有的 pending 邀请
		_, _ = tx.Exec(ctx, `
			UPDATE invitations SET status = 'revoked'
			WHERE workspace_id = $1 AND email = $2 AND status = 'pending' AND expires_at > now()`,
			in.WorkspaceID, in.Email)

		// 检查用户是否已是成员
		var existingUserID int64
		err := tx.QueryRow(ctx, `
			SELECT user_id FROM workspace_members
			WHERE workspace_id = $1 AND user_id = (SELECT id FROM users WHERE email = $2 AND is_active)`,
			in.WorkspaceID, in.Email).Scan(&existingUserID)
		if err == nil {
			return errs.New("INVITATION.ALREADY_MEMBER", "该用户已是工作空间成员", 409)
		}

		var i Invitation
		err = tx.QueryRow(ctx, `
			INSERT INTO invitations (workspace_id, inviter_id, email, role, token_hash, message, expires_at)
			VALUES ($1, $2, $3, $4, $5, $6, now() + interval '7 days')
			RETURNING id, workspace_id, inviter_id, email, role, status, message, expires_at, created_at`,
			in.WorkspaceID, in.InviterID, in.Email, in.Role, hash, in.Message).
			Scan(&i.ID, &i.WorkspaceID, &i.InviterID, &i.Email, &i.Role, &i.Status, &i.Message, &i.ExpiresAt, &i.CreatedAt)
		if err != nil {
			return errs.ErrInternal.Wrap(err)
		}

		// 查询 inviter 名称（用于邮件）
		_ = tx.QueryRow(ctx, `SELECT display_name FROM users WHERE id = $1`, in.InviterID).Scan(&inviterName)
		_ = tx.QueryRow(ctx, `SELECT name FROM workspaces WHERE id = $1`, in.WorkspaceID).Scan(&wsName)

		i.InviterName = inviterName
		inv = &i
		return nil
	})
	if err != nil {
		return nil, err
	}

	// 异步发送邀请邮件（事务外）
	go s.sendInviteEmail(inviterName, wsName, inv.Email, rawToken)

	return inv, nil
}

// Accept 用户接受邀请 → 写入 workspace_members + 标记 invitation accepted。
func (s *InvitationService) Accept(ctx context.Context, token string, userID int64) (*Invitation, error) {
	hash := sha256Hex(token)

	var inv *Invitation
	err := pgx.BeginTxFunc(ctx, s.db, pgx.TxOptions{}, func(tx pgx.Tx) error {
		var i Invitation
		err := tx.QueryRow(ctx, `
			SELECT id, workspace_id, inviter_id, email, role, status, expires_at, created_at
			FROM invitations WHERE token_hash = $1`, hash).
			Scan(&i.ID, &i.WorkspaceID, &i.InviterID, &i.Email, &i.Role, &i.Status, &i.ExpiresAt, &i.CreatedAt)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return errs.New("INVITATION.NOT_FOUND", "邀请不存在", 404)
			}
			return errs.ErrInternal.Wrap(err)
		}

		if i.Status != "pending" {
			return errs.New("INVITATION.NOT_PENDING", "邀请已被处理", 409)
		}
		if i.ExpiresAt.Before(time.Now()) {
			_, _ = tx.Exec(ctx, `UPDATE invitations SET status = 'expired' WHERE id = $1`, i.ID)
			return errs.New("INVITATION.EXPIRED", "邀请已过期", 410)
		}

		// 校验 user_id 对应的 email 与邀请 email 一致
		var invitedEmail string
		err = tx.QueryRow(ctx, `SELECT email FROM users WHERE id = $1 AND is_active`, userID).Scan(&invitedEmail)
		if err != nil {
			return errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "user", Reason: "无效的用户"})
		}
		if invitedEmail != i.Email {
			return errs.New("INVITATION.EMAIL_MISMATCH", "邀请邮箱与您当前登录账号不匹配", 403)
		}

		// 写入 workspace_members
		_, err = tx.Exec(ctx, `
			INSERT INTO workspace_members (workspace_id, user_id, role, joined_at)
			VALUES ($1, $2, $3, now())
			ON CONFLICT (workspace_id, user_id) DO UPDATE SET role = EXCLUDED.role`,
			i.WorkspaceID, userID, i.Role)
		if err != nil {
			return errs.ErrInternal.Wrap(err)
		}

		// 标记 accepted
		_, err = tx.Exec(ctx, `UPDATE invitations SET status = 'accepted', accepted_at = now() WHERE id = $1`, i.ID)
		if err != nil {
			return errs.ErrInternal.Wrap(err)
		}
		i.Status = "accepted"
		now := time.Now()
		i.AcceptedAt = &now
		inv = &i
		return nil
	})
	if err != nil {
		return nil, err
	}
	return inv, nil
}

// Revoke 撤销邀请。
func (s *InvitationService) Revoke(ctx context.Context, invID, wsID int64) error {
	tag, err := s.db.Exec(ctx, `
		UPDATE invitations SET status = 'revoked', updated_at = now()
		WHERE id = $1 AND workspace_id = $2 AND status = 'pending'`, invID, wsID)
	if err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	if tag.RowsAffected() == 0 {
		return errs.New("INVITATION.NOT_REVOKABLE", "邀请不存在或已处理", 409)
	}
	return nil
}

// ListByWorkspace 列出工作空间的邀请列表。
func (s *InvitationService) ListByWorkspace(ctx context.Context, wsID int64, statusFilter string) ([]Invitation, error) {
	query := `
		SELECT i.id, i.workspace_id, i.inviter_id, u.display_name, i.email, i.role, i.status,
		       i.message, i.expires_at, i.accepted_at, i.created_at
		FROM invitations i
		LEFT JOIN users u ON u.id = i.inviter_id
		WHERE i.workspace_id = $1`
	args := []any{wsID}
	if statusFilter != "" {
		query += " AND i.status = $" + strconv.Itoa(len(args)+1)
		args = append(args, statusFilter)
	}
	query += " ORDER BY i.created_at DESC"

	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()

	var out []Invitation
	for rows.Next() {
		var i Invitation
		if err := rows.Scan(&i.ID, &i.WorkspaceID, &i.InviterID, &i.InviterName,
			&i.Email, &i.Role, &i.Status, &i.Message, &i.ExpiresAt, &i.AcceptedAt, &i.CreatedAt); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
		out = append(out, i)
	}
	return out, nil
}

// InvitationPreview 邀请的公开预览信息。
type InvitationPreview struct {
	WorkspaceID   int64     `json:"workspace_id"`
	WorkspaceName string    `json:"workspace_name"`
	InviterName   string    `json:"inviter_name"`
	Email         string    `json:"email"`
	Role          string    `json:"role"`
	ExpiresAt     time.Time `json:"expires_at"`
	Status        string    `json:"status"`
}

// Preview 返回邀请的公开预览（接受前确认页用）。
func (s *InvitationService) Preview(ctx context.Context, token string) (*InvitationPreview, error) {
	hash := sha256Hex(token)
	var p InvitationPreview
	err := s.db.QueryRow(ctx, `
		SELECT i.workspace_id, w.name, u.display_name, i.email, i.role, i.expires_at, i.status
		FROM invitations i
		JOIN workspaces w ON w.id = i.workspace_id
		LEFT JOIN users u ON u.id = i.inviter_id
		WHERE i.token_hash = $1`, hash).
		Scan(&p.WorkspaceID, &p.WorkspaceName, &p.InviterName, &p.Email, &p.Role, &p.ExpiresAt, &p.Status)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.New("INVITATION.NOT_FOUND", "邀请不存在", 404)
		}
		return nil, errs.ErrInternal.Wrap(err)
	}
	return &p, nil
}

// --- Internal ---

func (s *InvitationService) sendInviteEmail(inviterName, wsName, toEmail, rawToken string) {
	if s.mailer == nil {
		return
	}
	acceptURL := fmt.Sprintf("%s/invite/%s", s.baseURL, rawToken)
	msg := mail.RenderInvitation(mail.InviteData{
		InviteeName:   strings.Split(toEmail, "@")[0],
		InviterName:   inviterName,
		WorkspaceName: wsName,
		AcceptURL:     acceptURL,
	})
	msg.To = toEmail
	_ = s.mailer.Send(msg)
}

// --- helpers ---

func generateSecureToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func sha256Hex(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}

