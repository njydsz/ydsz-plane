// Package collaboration — 实时协作应用服务。
package collaboration

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// Service 应用服务。
type Service struct {
	db *pgxpool.Pool
	// broadcaster 可选的广播回调（由基础设施层注入 ws.Hub）。
	broadcaster func(op *Operation)
}

// NewService 创建服务。
func NewService(db *pgxpool.Pool) *Service {
	return &Service{db: db}
}

// SetBroadcaster 注入广播回调。
func (s *Service) SetBroadcaster(fn func(op *Operation)) {
	s.broadcaster = fn
}

// --- 错误 ---

var (
	// ErrSessionNotFound 会话不存在。
	ErrSessionNotFound = errors.New("collaboration session not found")
	// ErrVersionConflict 版本冲突（乐观锁失败）。
	ErrVersionConflict = errors.New("document version conflict")
	// ErrDocumentNotFound 文档不存在。
	ErrDocumentNotFound = errors.New("document not found")
)

// 操作空闲阈值：超过该时长未心跳视为 idle。
const idleThreshold = 90 * time.Second

// --- 会话 ---

// JoinSession 加入（或续期）文档编辑会话。
func (s *Service) JoinSession(ctx context.Context, in JoinSessionInput) (*CollaborationSession, error) {
	if !in.DocumentType.IsValid() {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "document_type", Reason: "invalid value"})
	}
	if in.DocumentID <= 0 || in.UserID <= 0 {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "document_id", Reason: "required"})
	}

	// 幂等：同一用户 + 同一文档复用已有 active 会话。
	var existing CollaborationSession
	err := s.db.QueryRow(ctx, `
		SELECT id, workspace_id, document_type, document_id, user_id,
			COALESCE(user_name,''), COALESCE(display_name,''), status, base_version, last_seen_at, created_at
		FROM collaboration_sessions
		WHERE workspace_id = $1 AND document_type = $2 AND document_id = $3 AND user_id = $4
		  AND status IN ('active','idle')
		ORDER BY last_seen_at DESC LIMIT 1`,
		in.WorkspaceID, in.DocumentType, in.DocumentID, in.UserID,
	).Scan(&existing.ID, &existing.WorkspaceID, &existing.DocumentType, &existing.DocumentID,
		&existing.UserID, &existing.UserName, &existing.DisplayName, &existing.Status,
		&existing.BaseVersion, &existing.LastSeenAt, &existing.CreatedAt)
	if err == nil {
		// 续期
		if _, uerr := s.db.Exec(ctx, `
			UPDATE collaboration_sessions
			SET status = 'active', last_seen_at = now(), user_name = COALESCE(NULLIF($1,''), user_name)
			WHERE id = $2`, nullIfEmpty(in.UserName), existing.ID); uerr != nil {
			return nil, errs.ErrInternal.Wrap(uerr)
		}
		existing.Status = SessionActive
		existing.LastSeenAt = time.Now()
		if in.UserName != "" {
			existing.UserName = in.UserName
		}
		return &existing, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, errs.ErrInternal.Wrap(err)
	}

	var sess CollaborationSession
	err = s.db.QueryRow(ctx, `
		INSERT INTO collaboration_sessions
			(id, workspace_id, document_type, document_id, user_id, user_name, status, base_version)
		VALUES ($1,$2,$3,$4,$5,$6,'active',0)
		RETURNING id, workspace_id, document_type, document_id, user_id,
			COALESCE(user_name,''), COALESCE(display_name,''), status, base_version, last_seen_at, created_at`,
		genID(), in.WorkspaceID, in.DocumentType, in.DocumentID, in.UserID,
		nullIfEmpty(in.UserName),
	).Scan(&sess.ID, &sess.WorkspaceID, &sess.DocumentType, &sess.DocumentID,
		&sess.UserID, &sess.UserName, &sess.DisplayName, &sess.Status,
		&sess.BaseVersion, &sess.LastSeenAt, &sess.CreatedAt)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}

	s.broadcast(in.WorkspaceID, &Operation{
		DocumentType: in.DocumentType,
		DocumentID:   in.DocumentID,
		UserID:       in.UserID,
		UserName:     in.UserName,
		OpType:       "join",
	})
	return &sess, nil
}

// LeaveSession 离开会话。
func (s *Service) LeaveSession(ctx context.Context, workspaceID int64, documentType string, documentID, userID int64) error {
	_, err := s.db.Exec(ctx, `
		UPDATE collaboration_sessions SET status = 'closed', last_seen_at = now()
		WHERE workspace_id = $1 AND document_type = $2 AND document_id = $3 AND user_id = $4
		  AND status IN ('active','idle')`,
		workspaceID, documentType, documentID, userID)
	if err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	s.broadcast(workspaceID, &Operation{
		DocumentType: DocumentType(documentType),
		DocumentID:   documentID,
		UserID:       userID,
		OpType:       "leave",
	})
	return nil
}

// Heartbeat 心跳续期。
func (s *Service) Heartbeat(ctx context.Context, in HeartbeatInput) error {
	tag, err := s.db.Exec(ctx, `
		UPDATE collaboration_sessions
		SET status = 'active', last_seen_at = now(), base_version = $1
		WHERE workspace_id = $2 AND document_type = $3 AND document_id = $4 AND user_id = $5
		  AND status IN ('active','idle')`,
		in.BaseVersion, in.WorkspaceID, in.DocumentType, in.DocumentID, in.UserID)
	if err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	if tag.RowsAffected() == 0 {
		return ErrSessionNotFound
	}
	return nil
}

// ListSessions 列出文档的活跃编辑者。
func (s *Service) ListSessions(ctx context.Context, opts ListSessionsOptions) ([]*CollaborationSession, error) {
	// 顺带清理空闲会话。
	_, _ = s.db.Exec(ctx, `
		UPDATE collaboration_sessions SET status = 'idle'
		WHERE document_type = $1 AND document_id = $2 AND status = 'active'
		  AND last_seen_at < now() - $3::interval`,
		opts.DocumentType, opts.DocumentID, "90 seconds")

	rows, err := s.db.Query(ctx, `
		SELECT id, workspace_id, document_type, document_id, user_id,
			COALESCE(user_name,''), COALESCE(display_name,''), status, base_version, last_seen_at, created_at
		FROM collaboration_sessions
		WHERE workspace_id = $1 AND document_type = $2 AND document_id = $3
		  AND status IN ('active','idle')
		ORDER BY last_seen_at DESC
		LIMIT $4`,
		opts.WorkspaceID, opts.DocumentType, opts.DocumentID, opts.Limit)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()

	var out []*CollaborationSession
	for rows.Next() {
		var sess CollaborationSession
		if err := rows.Scan(&sess.ID, &sess.WorkspaceID, &sess.DocumentType, &sess.DocumentID,
			&sess.UserID, &sess.UserName, &sess.DisplayName, &sess.Status,
			&sess.BaseVersion, &sess.LastSeenAt, &sess.CreatedAt); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
		out = append(out, &sess)
	}
	return out, rows.Err()
}

// --- 操作与版本 ---

// SubmitOperation 提交协作操作（乐观锁校验）。
//
// 若 base_version 与当前文档版本不一致，返回冲突（Accepted=false），
// 由客户端拉取最新内容后重试。
func (s *Service) SubmitOperation(ctx context.Context, in SubmitOperationInput) (*SubmitResult, error) {
	if !in.DocumentType.IsValid() {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "document_type", Reason: "invalid value"})
	}
	if in.OpType == "" {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "op_type", Reason: "required"})
	}

	// 读取当前版本（不存在则初始化为 0）。
	var current int64
	err := s.db.QueryRow(ctx, `
		SELECT version FROM collaboration_documents
		WHERE document_type = $1 AND document_id = $2`,
		in.DocumentType, in.DocumentID).Scan(&current)
	if errors.Is(err, pgx.ErrNoRows) {
		current = 0
	} else if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}

	if in.OpType == "edit" && in.BaseVersion != current {
		return &SubmitResult{
			Accepted:   false,
			NewVersion: current,
			Reason:     "文档已被他人更新，请刷新后重试",
		}, nil
	}

	// 自增版本。
	var newVersion int64
	if err := s.db.QueryRow(ctx, `
		INSERT INTO collaboration_documents (id, document_type, document_id, version, updated_by)
		VALUES ($1,$2,$3,1,$4)
		ON CONFLICT (document_type, document_id) DO UPDATE SET
			version = collaboration_documents.version + 1,
			updated_by = EXCLUDED.updated_by,
			updated_at = now()
		RETURNING version`,
		genID(), in.DocumentType, in.DocumentID, in.UserID).Scan(&newVersion); err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}

	op := &Operation{
		ID:           genID(),
		WorkspaceID:  in.WorkspaceID,
		DocumentType: in.DocumentType,
		DocumentID:   in.DocumentID,
		UserID:       in.UserID,
		UserName:     in.UserName,
		OpType:       in.OpType,
		BaseVersion:  in.BaseVersion,
		TargetVersion: newVersion,
		Payload:      in.Payload,
		CreatedAt:    time.Now(),
	}

	// 记录操作日志。
	if _, err := s.db.Exec(ctx, `
		INSERT INTO collaboration_operations
			(id, workspace_id, document_type, document_id, user_id, user_name, op_type, base_version, target_version, payload)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		op.ID, op.WorkspaceID, op.DocumentType, op.DocumentID, op.UserID,
		nullIfEmpty(op.UserName), op.OpType, op.BaseVersion, op.TargetVersion,
		nullIfEmpty(op.Payload)); err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}

	s.broadcast(in.WorkspaceID, op)
	return &SubmitResult{Accepted: true, NewVersion: newVersion, Operation: op}, nil
}

// GetDocumentVersion 获取文档当前版本。
func (s *Service) GetDocumentVersion(ctx context.Context, documentType DocumentType, documentID int64) (int64, error) {
	var v int64
	err := s.db.QueryRow(ctx, `
		SELECT version FROM collaboration_documents
		WHERE document_type = $1 AND document_id = $2`, documentType, documentID).Scan(&v)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, nil
	}
	if err != nil {
		return 0, errs.ErrInternal.Wrap(err)
	}
	return v, nil
}

// --- 广播 ---

func (s *Service) broadcast(workspaceID int64, op *Operation) {
	if s.broadcaster == nil || workspaceID <= 0 {
		return
	}
	s.broadcaster(op)
}

// --- helpers ---

func nullIfEmpty(s string) any {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return s
}

var _ = idleThreshold
