// Package aiagent — Agent 应用服务。
package aiagent

import (
	"context"
	"encoding/json"
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
}

// NewService 创建服务。
func NewService(db *pgxpool.Pool) *Service { return &Service{db: db} }

// --- 错误 ---

var (
	// ErrAgentNotFound Agent 不存在。
	ErrAgentNotFound = errors.New("agent not found")
	// ErrConversationNotFound 会话不存在。
	ErrConversationNotFound = errors.New("conversation not found")
	// ErrActionNotFound 动作不存在。
	ErrActionNotFound = errors.New("action not found")
	// ErrActionNotPending 动作不在待处理状态。
	ErrActionNotPending = errors.New("action is not pending")
	// ErrCreditExhausted 额度耗尽。
	ErrCreditExhausted = errors.New("daily credit limit exceeded")
)

// --- Agent ---

// CreateAgent 创建 Agent。
func (s *Service) CreateAgent(ctx context.Context, in AgentInput) (*Agent, error) {
	if strings.TrimSpace(in.Name) == "" {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "name", Reason: "required"})
	}
	if in.WorkspaceID <= 0 {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "workspace_id", Reason: "required"})
	}
	if in.Model == "" {
		in.Model = "gpt-4o-mini"
	}
	if in.DailyCreditLimit <= 0 {
		in.DailyCreditLimit = 100
	}

	var a Agent
	err := s.db.QueryRow(ctx, `
		INSERT INTO ai_agents (id, workspace_id, name, description, instructions, triggers, skills,
			scope_project_ids, scope_work_item_types, model, temperature, daily_credit_limit, created_by)
		VALUES ($1,$2,$3,$4,$5,$6::text[],$7::text[],$8::bigint[],$9::text[],$10,$11,$12,$13)
		RETURNING id, workspace_id, name, COALESCE(description,''), COALESCE(instructions,''),
			triggers, skills, scope_project_ids, scope_work_item_types, model, temperature,
			daily_credit_limit, used_credits_today, status, created_by, created_at, updated_at`,
		genID(), in.WorkspaceID, in.Name, nullIfEmpty(in.Description),
		nullIfEmpty(in.Instructions), strSlice(in.Triggers), strSlice(in.Skills),
		int64Slice(in.ScopeProjectIDs), strSlice(in.ScopeWorkItemTypes),
		in.Model, in.Temperature, in.DailyCreditLimit, in.CreatedBy,
	).Scan(&a.ID, &a.WorkspaceID, &a.Name, &a.Description, &a.Instructions,
		&a.Triggers, &a.Skills, &a.ScopeProjectIDs, &a.ScopeWorkItemTypes,
		&a.Model, &a.Temperature, &a.DailyCreditLimit, &a.UsedCreditsToday,
		&a.Status, &a.CreatedBy, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	return &a, nil
}

// GetAgent 查询 Agent。
func (s *Service) GetAgent(ctx context.Context, id int64) (*Agent, error) {
	var a Agent
	err := s.db.QueryRow(ctx, `
		SELECT id, workspace_id, name, COALESCE(description,''), COALESCE(instructions,''),
			triggers, skills, scope_project_ids, scope_work_item_types, model, temperature,
			daily_credit_limit, used_credits_today, status, created_by, created_at, updated_at
		FROM ai_agents WHERE id = $1`, id,
	).Scan(&a.ID, &a.WorkspaceID, &a.Name, &a.Description, &a.Instructions,
		&a.Triggers, &a.Skills, &a.ScopeProjectIDs, &a.ScopeWorkItemTypes,
		&a.Model, &a.Temperature, &a.DailyCreditLimit, &a.UsedCreditsToday,
		&a.Status, &a.CreatedBy, &a.CreatedAt, &a.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrAgentNotFound
	}
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	return &a, nil
}

// ListAgents 列出工作空间的 Agent。
func (s *Service) ListAgents(ctx context.Context, workspaceID int64) ([]*Agent, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, workspace_id, name, COALESCE(description,''), COALESCE(instructions,''),
			triggers, skills, scope_project_ids, scope_work_item_types, model, temperature,
			daily_credit_limit, used_credits_today, status, created_by, created_at, updated_at
		FROM ai_agents WHERE workspace_id = $1 ORDER BY created_at DESC`, workspaceID)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()

	var out []*Agent
	for rows.Next() {
		var a Agent
		if err := rows.Scan(&a.ID, &a.WorkspaceID, &a.Name, &a.Description, &a.Instructions,
			&a.Triggers, &a.Skills, &a.ScopeProjectIDs, &a.ScopeWorkItemTypes,
			&a.Model, &a.Temperature, &a.DailyCreditLimit, &a.UsedCreditsToday,
			&a.Status, &a.CreatedBy, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
		out = append(out, &a)
	}
	return out, rows.Err()
}

// UpdateAgent 更新 Agent。
func (s *Service) UpdateAgent(ctx context.Context, id int64, in AgentInput) (*Agent, error) {
	var a Agent
	err := s.db.QueryRow(ctx, `
		UPDATE ai_agents SET
			name = COALESCE(NULLIF($1,''), name),
			description = COALESCE(NULLIF($2,''), description),
			instructions = COALESCE(NULLIF($3,''), instructions),
			triggers = COALESCE($4::text[], triggers),
			skills = COALESCE($5::text[], skills),
			scope_project_ids = COALESCE($6::bigint[], scope_project_ids),
			scope_work_item_types = COALESCE($7::text[], scope_work_item_types),
			model = COALESCE(NULLIF($8,''), model),
			temperature = COALESCE($9, temperature),
			daily_credit_limit = COALESCE($10, daily_credit_limit),
			updated_at = now()
		WHERE id = $11
		RETURNING id, workspace_id, name, COALESCE(description,''), COALESCE(instructions,''),
			triggers, skills, scope_project_ids, scope_work_item_types, model, temperature,
			daily_credit_limit, used_credits_today, status, created_by, created_at, updated_at`,
		nullIfEmpty(in.Name), nullIfEmpty(in.Description), nullIfEmpty(in.Instructions),
		strSlice(in.Triggers), strSlice(in.Skills), int64Slice(in.ScopeProjectIDs),
		strSlice(in.ScopeWorkItemTypes), nullIfEmpty(in.Model),
		nullIfZeroFloat(in.Temperature), nullIfZeroInt(in.DailyCreditLimit), id,
	).Scan(&a.ID, &a.WorkspaceID, &a.Name, &a.Description, &a.Instructions,
		&a.Triggers, &a.Skills, &a.ScopeProjectIDs, &a.ScopeWorkItemTypes,
		&a.Model, &a.Temperature, &a.DailyCreditLimit, &a.UsedCreditsToday,
		&a.Status, &a.CreatedBy, &a.CreatedAt, &a.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrAgentNotFound
	}
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	return &a, nil
}

// DeleteAgent 删除 Agent。
func (s *Service) DeleteAgent(ctx context.Context, id int64) error {
	tag, err := s.db.Exec(ctx, `DELETE FROM ai_agents WHERE id = $1`, id)
	if err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	if tag.RowsAffected() == 0 {
		return ErrAgentNotFound
	}
	return nil
}

// --- 记忆 ---

// SetMemory 写入记忆（按 scope+scope_id+key 幂等 upsert）。
func (s *Service) SetMemory(ctx context.Context, workspaceID int64, scope MemoryScope, scopeID int64, key, content, source string, userID int64) (*AIWorkspaceMemory, error) {
	if scope == "" {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "scope", Reason: "required"})
	}
	if strings.TrimSpace(content) == "" {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "content", Reason: "required"})
	}

	var m AIWorkspaceMemory
	err := s.db.QueryRow(ctx, `
		INSERT INTO ai_memory (id, workspace_id, scope, scope_id, key, content, source, created_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		ON CONFLICT (workspace_id, scope, scope_id, key) DO UPDATE SET
			content = EXCLUDED.content,
			source = EXCLUDED.source,
			updated_at = now()
		RETURNING id, workspace_id, scope, scope_id, key, content, source, created_by, created_at, updated_at`,
		genID(), workspaceID, string(scope), scopeID, nullIfEmpty(key), content,
		nullIfEmpty(source), userID,
	).Scan(&m.ID, &m.WorkspaceID, &m.Scope, &m.ScopeID, &m.Key, &m.Content,
		&m.Source, &m.CreatedBy, &m.CreatedAt, &m.UpdatedAt)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	return &m, nil
}

// ListMemory 列出记忆（可按 scope/scope_id 过滤）。
func (s *Service) ListMemory(ctx context.Context, workspaceID int64, scope MemoryScope, scopeID int64) ([]*AIWorkspaceMemory, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, workspace_id, scope, scope_id, key, content, source, created_by, created_at, updated_at
		FROM ai_memory
		WHERE workspace_id = $1
		  AND ($2 = '' OR scope = $2)
		  AND ($3 = 0 OR scope_id = $3)
		ORDER BY updated_at DESC
		LIMIT 200`, workspaceID, string(scope), scopeID)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()

	var out []*AIWorkspaceMemory
	for rows.Next() {
		var m AIWorkspaceMemory
		if err := rows.Scan(&m.ID, &m.WorkspaceID, &m.Scope, &m.ScopeID, &m.Key,
			&m.Content, &m.Source, &m.CreatedBy, &m.CreatedAt, &m.UpdatedAt); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
		out = append(out, &m)
	}
	return out, rows.Err()
}

// DeleteMemory 删除记忆。
func (s *Service) DeleteMemory(ctx context.Context, id, workspaceID int64) error {
	_, err := s.db.Exec(ctx, `DELETE FROM ai_memory WHERE id = $1 AND workspace_id = $2`, id, workspaceID)
	if err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	return nil
}

// BuildContext 汇总当前作用域的 AI 上下文（用于注入 prompt）。
func (s *Service) BuildContext(ctx context.Context, workspaceID int64, scopeIDs map[MemoryScope]int64) (string, error) {
	var sb strings.Builder
	for scope, id := range scopeIDs {
		items, err := s.ListMemory(ctx, workspaceID, scope, id)
		if err != nil {
			return "", err
		}
		if len(items) == 0 {
			continue
		}
		sb.WriteString("## ")
		sb.WriteString(string(scope))
		sb.WriteString(" context\n")
		for _, m := range items {
			if m.Key != "" {
				sb.WriteString("- ")
				sb.WriteString(m.Key)
				sb.WriteString(": ")
			}
			sb.WriteString(m.Content)
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
	}
	return sb.String(), nil
}

// --- 会话 ---

// CreateConversation 创建会话。
func (s *Service) CreateConversation(ctx context.Context, in ChatInput) (*Conversation, error) {
	if !IsValidMode(string(in.Mode)) {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "mode", Reason: "invalid value"})
	}
	if strings.TrimSpace(in.Message) == "" {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "message", Reason: "required"})
	}
	title := in.Message
	if len([]rune(title)) > 60 {
		title = string([]rune(title)[:60]) + "..."
	}

	var conv Conversation
	err := s.db.QueryRow(ctx, `
		INSERT INTO ai_conversations (id, workspace_id, agent_id, user_id, mode, title)
		VALUES ($1,$2,$3,$4,$5,$6)
		RETURNING id, workspace_id, agent_id, user_id, mode, title, created_at, updated_at`,
		genID(), in.WorkspaceID, in.AgentID, in.UserID, string(in.Mode), title,
	).Scan(&conv.ID, &conv.WorkspaceID, &conv.AgentID, &conv.UserID, &conv.Mode,
		&conv.Title, &conv.CreatedAt, &conv.UpdatedAt)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	return &conv, nil
}

// ListConversations 列出会话。
func (s *Service) ListConversations(ctx context.Context, workspaceID, userID int64) ([]*Conversation, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, workspace_id, agent_id, user_id, mode, title, created_at, updated_at
		FROM ai_conversations
		WHERE workspace_id = $1 AND user_id = $2
		ORDER BY updated_at DESC LIMIT 100`, workspaceID, userID)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()

	var out []*Conversation
	for rows.Next() {
		var c Conversation
		if err := rows.Scan(&c.ID, &c.WorkspaceID, &c.AgentID, &c.UserID, &c.Mode,
			&c.Title, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
		out = append(out, &c)
	}
	return out, rows.Err()
}

// GetConversation 查询会话（含消息）。
func (s *Service) GetConversation(ctx context.Context, id int64) (*Conversation, []*Message, error) {
	var c Conversation
	err := s.db.QueryRow(ctx, `
		SELECT id, workspace_id, agent_id, user_id, mode, title, created_at, updated_at
		FROM ai_conversations WHERE id = $1`, id,
	).Scan(&c.ID, &c.WorkspaceID, &c.AgentID, &c.UserID, &c.Mode, &c.Title, &c.CreatedAt, &c.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil, ErrConversationNotFound
	}
	if err != nil {
		return nil, nil, errs.ErrInternal.Wrap(err)
	}

	msgs, err := s.listMessages(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	return &c, msgs, nil
}

func (s *Service) listMessages(ctx context.Context, conversationID int64) ([]*Message, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, conversation_id, role, content, created_at
		FROM ai_messages WHERE conversation_id = $1 ORDER BY created_at ASC`, conversationID)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()

	var out []*Message
	for rows.Next() {
		var m Message
		if err := rows.Scan(&m.ID, &m.ConversationID, &m.Role, &m.Content, &m.CreatedAt); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
		out = append(out, &m)
	}
	return out, rows.Err()
}

// AppendMessage 追加消息。
func (s *Service) AppendMessage(ctx context.Context, conversationID int64, role, content string, actions []AIAction) (*Message, error) {
	var m Message
	err := s.db.QueryRow(ctx, `
		INSERT INTO ai_messages (id, conversation_id, role, content)
		VALUES ($1,$2,$3,$4)
		RETURNING id, conversation_id, role, content, created_at`,
		genID(), conversationID, role, content,
	).Scan(&m.ID, &m.ConversationID, &m.Role, &m.Content, &m.CreatedAt)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}

	for i := range actions {
		actions[i].MessageID = m.ID
		if _, err := s.insertAction(ctx, &actions[i]); err != nil {
			return nil, err
		}
	}
	m.Actions = actions
	_, _ = s.db.Exec(ctx, `UPDATE ai_conversations SET updated_at = now() WHERE id = $1`, conversationID)
	return &m, nil
}

func (s *Service) insertAction(ctx context.Context, a *AIAction) (*AIAction, error) {
	if err := s.db.QueryRow(ctx, `
		INSERT INTO ai_actions (id, message_id, type, payload, status)
		VALUES ($1,$2,$3,$4,'pending')
		RETURNING id, message_id, type, payload, status, COALESCE(result,''), COALESCE(error_message,''),
			executed_by, executed_at, created_at`,
		genID(), a.MessageID, string(a.Type), a.Payload,
	).Scan(&a.ID, &a.MessageID, &a.Type, &a.Payload, &a.Status, &a.Result,
		&a.ErrorMessage, &a.ExecutedBy, &a.ExecutedAt, &a.CreatedAt); err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	return a, nil
}

// ListActions 列出会话的待处理动作。
func (s *Service) ListPendingActions(ctx context.Context, conversationID int64) ([]*AIAction, error) {
	rows, err := s.db.Query(ctx, `
		SELECT a.id, a.message_id, a.type, a.payload, a.status, COALESCE(a.result,''),
			COALESCE(a.error_message,''), a.executed_by, a.executed_at, a.created_at
		FROM ai_actions a
		JOIN ai_messages m ON m.id = a.message_id
		WHERE m.conversation_id = $1 AND a.status = 'pending'
		ORDER BY a.created_at ASC`, conversationID)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()

	var out []*AIAction
	for rows.Next() {
		var a AIAction
		if err := rows.Scan(&a.ID, &a.MessageID, &a.Type, &a.Payload, &a.Status,
			&a.Result, &a.ErrorMessage, &a.ExecutedBy, &a.ExecutedAt, &a.CreatedAt); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
		out = append(out, &a)
	}
	return out, rows.Err()
}

// ResolveAction 处理动作（批准/拒绝/标记执行）。
func (s *Service) ResolveAction(ctx context.Context, actionID int64, status string, executedBy int64, result, errMsg string) (*AIAction, error) {
	var a AIAction
	err := s.db.QueryRow(ctx, `
		UPDATE ai_actions
		SET status = $1,
			result = COALESCE(NULLIF($2,''), result),
			error_message = COALESCE(NULLIF($3,''), error_message),
			executed_by = COALESCE($4, executed_by),
			executed_at = CASE WHEN $1 IN ('executed','failed') THEN now() ELSE executed_at END
		WHERE id = $5 AND status = 'pending'
		RETURNING id, message_id, type, payload, status, COALESCE(result,''), COALESCE(error_message,''),
			executed_by, executed_at, created_at`,
		status, result, errMsg, executedBy, actionID,
	).Scan(&a.ID, &a.MessageID, &a.Type, &a.Payload, &a.Status, &a.Result,
		&a.ErrorMessage, &a.ExecutedBy, &a.ExecutedAt, &a.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrActionNotPending
	}
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	return &a, nil
}

// --- 动作执行 ---

// ActionPayload 动作载荷（通用 JSON）。
type ActionPayload struct {
	ProjectID    int64   `json:"project_id,omitempty"`
	WorkItemType string  `json:"work_item_type,omitempty"`
	Title        string  `json:"title,omitempty"`
	Description  string  `json:"description,omitempty"`
	StateID      int64   `json:"state_id,omitempty"`
	AssigneeID   int64   `json:"assignee_id,omitempty"`
	Priority     string  `json:"priority,omitempty"`
	LabelIDs     []int64 `json:"label_ids,omitempty"`
	ModuleID     int64   `json:"module_id,omitempty"`
	Name         string  `json:"name,omitempty"`
	Content      string  `json:"content,omitempty"`
	Hours        float64 `json:"hours,omitempty"`
	IDs          []int64 `json:"ids,omitempty"`
}

// ExecuteAction 执行已批准的动作。
//
// 说明：本实现覆盖工作项的创建/更新/状态/指派/评论/工时等核心动作。
// 未支持的类型返回错误（不静默失败）。
func (s *Service) ExecuteAction(ctx context.Context, action *AIAction, executedBy int64) (*AIAction, error) {
	var payload ActionPayload
	if action.Payload != "" {
		if err := json.Unmarshal([]byte(action.Payload), &payload); err != nil {
			return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "payload", Reason: "invalid JSON"})
		}
	}

	var result string
	var execErr error

	switch action.Type {
	case ActionCreateWorkItem:
		execErr = s.execCreateWorkItem(ctx, &payload)
	case ActionUpdateWorkItem:
		execErr = s.execUpdateWorkItem(ctx, &payload)
	case ActionUpdateState:
		execErr = s.execUpdateState(ctx, &payload)
	case ActionAssign:
		execErr = s.execAssign(ctx, &payload)
	case ActionAddComment:
		execErr = s.execAddComment(ctx, &payload)
	case ActionLogTime:
		execErr = s.execLogTime(ctx, &payload)
	case ActionCreateModule:
		execErr = s.execCreateModule(ctx, &payload)
	case ActionCreateCycle:
		execErr = s.execCreateCycle(ctx, &payload)
	case ActionDeleteWorkItem:
		execErr = s.execDeleteWorkItem(ctx, &payload)
	default:
		execErr = errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "type", Reason: "unsupported action type"})
	}

	if execErr != nil {
		return s.ResolveAction(ctx, action.ID, "failed", executedBy, "", execErr.Error())
	}
	return s.ResolveAction(ctx, action.ID, "executed", executedBy, result, "")
}

// --- helpers ---

func nullIfEmpty(s string) any {
	if strings.TrimSpace(s) == "" {
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

func int64Slice(v []int64) any {
	if len(v) == 0 {
		return nil
	}
	return v
}

func nullIfZeroFloat(f float64) any {
	if f == 0 {
		return nil
	}
	return f
}

func nullIfZeroInt(n int) any {
	if n == 0 {
		return nil
	}
	return n
}

var _ = time.Now

// getActionByID 按 ID 查询动作（供执行端点使用）。
func (s *Service) getActionByID(ctx context.Context, id int64) (*AIAction, error) {
	var a AIAction
	err := s.db.QueryRow(ctx, `
		SELECT id, message_id, type, payload, status, COALESCE(result,''),
			COALESCE(error_message,''), executed_by, executed_at, created_at
		FROM ai_actions WHERE id = $1`, id,
	).Scan(&a.ID, &a.MessageID, &a.Type, &a.Payload, &a.Status, &a.Result,
		&a.ErrorMessage, &a.ExecutedBy, &a.ExecutedAt, &a.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrActionNotFound
	}
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	return &a, nil
}
