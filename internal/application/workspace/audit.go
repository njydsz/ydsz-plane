// Package workspace — 审计日志领域服务。
//
// 设计参考: GitLab Event System、GitHub Audit Log。
// 所有管理操作（workspace、member、invitation、project CRUD）通过该服务写入 audit_logs。
// 审计记录 once written 即不可更新/删除（数据库只增约束）。
package workspace

import (
	"context"
	"encoding/json"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// AuditEntry 是一次审计记录的参数。
type AuditEntry struct {
	WorkspaceID int64
	ActorID     int64
	Action      string // 枚举：workspace.create / invitation.send / member.role_change / ...
	Target      string // 目标描述，如被操作人 email、project identifier
	Detail      map[string]any
	IP          string
}

// AuditService 管理审计日志写入。
type AuditService struct {
	db *pgxpool.Pool
}

// NewAuditService 创建审计服务。
func NewAuditService(db *pgxpool.Pool) *AuditService {
	return &AuditService{db: db}
}

// Record 同步记录审计事件（非阻塞）。
// 在大厂级实现中，这里应走 Kafka/SNS；当前用 PG 直接写入。
func (s *AuditService) Record(ctx context.Context, e AuditEntry) {
	detail, _ := json.Marshal(e.Detail)
	_, _ = s.db.Exec(ctx, `
		INSERT INTO audit_logs (workspace_id, actor_id, action, target, detail, ip)
		VALUES ($1, $2, $3, $4, $5, $6::inet)`,
		e.WorkspaceID, e.ActorID, e.Action, e.Target, string(detail), e.IP)
}

// RecordFromGin 从 gin context 提取 IP 和 actor 后记录。
func (s *AuditService) RecordFromGin(c *gin.Context, wsID int64, action, target string, detail map[string]any) {
	s.Record(c.Request.Context(), AuditEntry{
		WorkspaceID: wsID,
		ActorID:     c.GetInt64("user_id"),
		Action:      action,
		Target:      target,
		Detail:      detail,
		IP:          c.ClientIP(),
	})
}

// AuditLogVM 是审计日志的视图模型。
type AuditLogVM struct {
	ID        int64           `json:"id"`
	ActorID   int64           `json:"actor_id"`
	ActorName string          `json:"actor_name,omitempty"`
	Action    string          `json:"action"`
	Target    string          `json:"target,omitempty"`
	DetailRaw json.RawMessage `json:"detail,omitempty"`
	IP        string          `json:"ip,omitempty"`
	CreatedAt time.Time       `json:"created_at"`
}

// List 返回工作空间的审计日志列表（按时间倒序）。
func (s *AuditService) List(ctx context.Context, wsID int64, limit int) ([]AuditLogVM, error) {
	rows, err := s.db.Query(ctx, `
		SELECT a.id, a.actor_id, coalesce(u.display_name,'') AS actor_name,
		       a.action, coalesce(a.target,''), coalesce(a.detail,'{}'::jsonb),
		       coalesce(a.ip::text,''), a.created_at
		FROM audit_logs a
		LEFT JOIN users u ON u.id = a.actor_id
		WHERE a.workspace_id = $1
		ORDER BY a.created_at DESC
		LIMIT $2`, wsID, limit)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()

	var out []AuditLogVM
	for rows.Next() {
		var r AuditLogVM
		if err := rows.Scan(&r.ID, &r.ActorID, &r.ActorName, &r.Action, &r.Target, &r.DetailRaw, &r.IP, &r.CreatedAt); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
		out = append(out, r)
	}
	return out, nil
}
