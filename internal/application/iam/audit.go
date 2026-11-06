// Package iam — 权限审计日志。
//
// 记录权限变更的完整上下文（谁/何时/对哪个角色/改了什么/为什么），
// 支持合规审计与权限回溯。
package iam

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// AuditAction 审计动作类型。
type AuditAction string

const (
	// AuditGrant 授予权限。
	AuditGrant AuditAction = "grant"
	// AuditRevoke 回收权限。
	AuditRevoke AuditAction = "revoke"
	// AuditReplace 全量替换。
	AuditReplace AuditAction = "replace"
)

// AuditEntry 审计条目。
type AuditEntry struct {
	ID          int64           `json:"id"`
	WorkspaceID int64           `json:"workspace_id"`
	ActorID     int64           `json:"actor_id"`
	ActorName   string          `json:"actor_name,omitempty"`
	RoleSlug    string          `json:"role_slug"`
	Action      AuditAction     `json:"action"`
	Permissions []string        `json:"permissions"`
	Before      json.RawMessage `json:"before,omitempty"`
	After       json.RawMessage `json:"after,omitempty"`
	Reason      string          `json:"reason,omitempty"`
	IP          string          `json:"ip,omitempty"`
	UserAgent   string          `json:"user_agent,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
}

// AuditQuery 查询条件。
type AuditQuery struct {
	WorkspaceID int64
	RoleSlug    string
	ActorID     int64
	Action      string
	Limit       int
	Offset      int
}

// AuditService 审计服务。
type AuditService struct {
	db *pgxpool.Pool
}

// NewAuditService 构造服务。
func NewAuditService(db *pgxpool.Pool) *AuditService {
	return &AuditService{db: db}
}

// RecordInput 记录入参。
type RecordInput struct {
	WorkspaceID int64
	ActorID     int64
	ActorName   string
	RoleSlug    string
	Action      AuditAction
	Permissions []string
	Before      []byte
	After       []byte
	Reason      string
	IP          string
	UserAgent   string
}

// Record 记录权限变更。
func (s *AuditService) Record(ctx context.Context, in RecordInput) (*AuditEntry, error) {
	if in.WorkspaceID <= 0 || in.ActorID <= 0 {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "workspace_id/actor_id", Reason: "required"})
	}
	if in.RoleSlug == "" {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "role_slug", Reason: "required"})
	}

	beforeJSON := json.RawMessage("null")
	if len(in.Before) > 0 {
		beforeJSON = json.RawMessage(in.Before)
	}
	afterJSON := json.RawMessage("null")
	if len(in.After) > 0 {
		afterJSON = json.RawMessage(in.After)
	}

	var e AuditEntry
	err := s.db.QueryRow(ctx, `
		INSERT INTO permission_audit_logs
			(id, workspace_id, actor_id, actor_name, role_slug, action,
			 permissions, before, after, reason, ip, user_agent)
		VALUES ($1,$2,$3,$4,$5,$6,$7::text[],$8::jsonb,$9::jsonb,$10,$11,$12)
		RETURNING id, workspace_id, actor_id, COALESCE(actor_name,''),
			role_slug, action, permissions, COALESCE(before,'null'::jsonb),
			COALESCE(after,'null'::jsonb), COALESCE(reason,''), COALESCE(ip,''),
			COALESCE(user_agent,''), created_at`,
		genID(), in.WorkspaceID, in.ActorID, nullIfEmpty(in.ActorName),
		in.RoleSlug, string(in.Action), strSlice(in.Permissions),
		beforeJSON, afterJSON, nullIfEmpty(in.Reason),
		nullIfEmpty(in.IP), nullIfEmpty(in.UserAgent),
	).Scan(&e.ID, &e.WorkspaceID, &e.ActorID, &e.ActorName, &e.RoleSlug,
		&e.Action, &e.Permissions, &e.Before, &e.After, &e.Reason,
		&e.IP, &e.UserAgent, &e.CreatedAt)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	return &e, nil
}

// List 查询审计日志。
func (s *AuditService) List(ctx context.Context, q AuditQuery) ([]*AuditEntry, error) {
	conds := []string{"workspace_id = $1"}
	args := []any{q.WorkspaceID}
	if q.RoleSlug != "" {
		args = append(args, q.RoleSlug)
		conds = append(conds, "role_slug = $"+itoa(len(args)))
	}
	if q.ActorID > 0 {
		args = append(args, q.ActorID)
		conds = append(conds, "actor_id = $"+itoa(len(args)))
	}
	if q.Action != "" {
		args = append(args, q.Action)
		conds = append(conds, "action = $"+itoa(len(args)))
	}
	if q.Limit <= 0 || q.Limit > 500 {
		q.Limit = 100
	}
	args = append(args, q.Limit)
	if q.Offset < 0 {
		q.Offset = 0
	}
	args = append(args, q.Offset)

	rows, err := s.db.Query(ctx, `
		SELECT id, workspace_id, actor_id, COALESCE(actor_name,''),
			role_slug, action, permissions, COALESCE(before,'null'::jsonb),
			COALESCE(after,'null'::jsonb), COALESCE(reason,''), COALESCE(ip,''),
			COALESCE(user_agent,''), created_at
		FROM permission_audit_logs
		WHERE `+joinComma(conds)+`
		ORDER BY created_at DESC
		LIMIT $`+itoa(len(args)-1)+` OFFSET $`+itoa(len(args)), args...)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()

	var out []*AuditEntry
	for rows.Next() {
		var e AuditEntry
		if err := rows.Scan(&e.ID, &e.WorkspaceID, &e.ActorID, &e.ActorName,
			&e.RoleSlug, &e.Action, &e.Permissions, &e.Before, &e.After,
			&e.Reason, &e.IP, &e.UserAgent, &e.CreatedAt); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
		out = append(out, &e)
	}
	return out, rows.Err()
}

// GetByID 查询单条。
func (s *AuditService) GetByID(ctx context.Context, id int64) (*AuditEntry, error) {
	var e AuditEntry
	err := s.db.QueryRow(ctx, `
		SELECT id, workspace_id, actor_id, COALESCE(actor_name,''),
			role_slug, action, permissions, COALESCE(before,'null'::jsonb),
			COALESCE(after,'null'::jsonb), COALESCE(reason,''), COALESCE(ip,''),
			COALESCE(user_agent,''), created_at
		FROM permission_audit_logs WHERE id = $1`, id,
	).Scan(&e.ID, &e.WorkspaceID, &e.ActorID, &e.ActorName, &e.RoleSlug,
		&e.Action, &e.Permissions, &e.Before, &e.After, &e.Reason,
		&e.IP, &e.UserAgent, &e.CreatedAt)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	return &e, nil
}

// --- helpers ---

func genID() int64 {
	var b [8]byte
	_, _ = rand.Read(b[:])
	randBits := int64(binary.LittleEndian.Uint64(b[:]) & 0x3FFFFF)
	return (time.Now().UnixMilli() << 22) | randBits
}

func nullIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func strSlice(v []string) any {
	if len(v) == 0 {
		return nil
	}
	return v
}

func itoa(n int) string {
	if n < 10 {
		return string(rune('0' + n))
	}
	buf := []byte{}
	for n > 0 {
		buf = append([]byte{byte('0' + n%10)}, buf...)
		n /= 10
	}
	return string(buf)
}

func joinComma(parts []string) string {
	out := ""
	for i, p := range parts {
		if i > 0 {
			out += ", "
		}
		out += p
	}
	return out
}
