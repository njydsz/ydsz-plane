// Package aiagent — Agent Run 生命周期。
//
// 对标 Plane Agents 的 Run 概念：每次 Agent 执行（用户发起或触发器触发）
// 都是一个 Run，具有明确的状态机与可观测的执行轨迹。
//
// 状态机：
//
//	created → in_progress → completed
//	                     ↘ awaiting（等待用户输入/确认）
//	                     ↘ stopping → stopped
//	                     ↘ failed
//	                     ↘ stale（超时未收敛）
//
// 数据表：ai_agent_runs、ai_run_steps
package aiagent

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// RunStatus Run 状态。
type RunStatus string

const (
	// RunCreated 已创建，尚未开始。
	RunCreated RunStatus = "created"
	// RunInProgress 执行中。
	RunInProgress RunStatus = "in_progress"
	// RunAwaiting 等待用户输入/确认。
	RunAwaiting RunStatus = "awaiting"
	// RunCompleted 已完成。
	RunCompleted RunStatus = "completed"
	// RunStopping 正在停止。
	RunStopping RunStatus = "stopping"
	// RunStopped 已停止。
	RunStopped RunStatus = "stopped"
	// RunFailed 失败。
	RunFailed RunStatus = "failed"
	// RunStale 超时未收敛。
	RunStale RunStatus = "stale"
)

// IsValidRunStatus 校验状态值。
func IsValidRunStatus(s string) bool {
	switch RunStatus(s) {
	case RunCreated, RunInProgress, RunAwaiting, RunCompleted,
		RunStopping, RunStopped, RunFailed, RunStale:
		return true
	}
	return false
}

// AgentRun 一次 Agent 执行。
type AgentRun struct {
	ID             int64     `json:"id"`
	WorkspaceID    int64     `json:"workspace_id"`
	AgentID        int64     `json:"agent_id"`
	UserID         int64     `json:"user_id"`
	ConversationID int64     `json:"conversation_id"`
	Mode           Mode      `json:"mode"`
	Status         RunStatus `json:"status"`
	// 触发方式：user / mention / assignment / event / schedule / api
	Trigger string `json:"trigger"`
	// 触发上下文（如触发的事件负载）
	TriggerPayload string `json:"trigger_payload,omitempty"`
	// 用户输入
	Input string `json:"input,omitempty"`
	// 最终输出摘要
	Output string `json:"output,omitempty"`
	// 错误信息
	ErrorMessage string `json:"error_message,omitempty"`
	// token 用量
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
	// 工具调用次数
	ToolCallCount int `json:"tool_call_count"`
	// 是否需要用户输入（awaiting 原因）
	NeedsInputReason string     `json:"needs_input_reason,omitempty"`
	StartedAt        *time.Time `json:"started_at,omitempty"`
	CompletedAt      *time.Time `json:"completed_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

// RunStep Run 执行步骤（可观测轨迹）。
type RunStep struct {
	ID    int64 `json:"id"`
	RunID int64 `json:"run_id"`
	// 步骤类型：thought / action / response / elicitation
	StepType string `json:"step_type"`
	Summary  string `json:"summary"`
	// 详细内容（JSON）
	Detail    string    `json:"detail,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// --- 服务方法 ---

// RunService Run 生命周期服务。
type RunService struct {
	db *pgxpool.Pool
}

// NewRunService 构造服务。
func NewRunService(db *pgxpool.Pool) *RunService { return &RunService{db: db} }

// CreateRun 创建 Run。
func (s *RunService) CreateRun(ctx context.Context, in CreateRunInput) (*AgentRun, error) {
	if in.AgentID <= 0 {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "agent_id", Reason: "required"})
	}
	if in.Input == "" {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "input", Reason: "required"})
	}
	if !IsValidMode(string(in.Mode)) {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "mode", Reason: "invalid value"})
	}

	var r AgentRun
	err := s.db.QueryRow(ctx, `
		INSERT INTO ai_agent_runs
			(id, workspace_id, agent_id, user_id, conversation_id, mode, status, trigger,
			 trigger_payload, input, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,'created',$7,$8,$9,now(),now())
		RETURNING id, workspace_id, agent_id, user_id, conversation_id, mode, status, trigger,
			COALESCE(trigger_payload,''), COALESCE(input,''), COALESCE(output,''),
			COALESCE(error_message,''), prompt_tokens, completion_tokens, total_tokens,
			tool_call_count, COALESCE(needs_input_reason,''), started_at, completed_at,
			created_at, updated_at`,
		genID(), in.WorkspaceID, in.AgentID, in.UserID, in.ConversationID,
		string(in.Mode), nullIfEmpty(in.Trigger), nullIfEmpty(in.TriggerPayload), in.Input,
	).Scan(&r.ID, &r.WorkspaceID, &r.AgentID, &r.UserID, &r.ConversationID, &r.Mode,
		&r.Status, &r.Trigger, &r.TriggerPayload, &r.Input, &r.Output, &r.ErrorMessage,
		&r.PromptTokens, &r.CompletionTokens, &r.TotalTokens, &r.ToolCallCount,
		&r.NeedsInputReason, &r.StartedAt, &r.CompletedAt, &r.CreatedAt, &r.UpdatedAt)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	return &r, nil
}

// GetRun 查询 Run。
func (s *RunService) GetRun(ctx context.Context, id int64) (*AgentRun, error) {
	var r AgentRun
	err := s.db.QueryRow(ctx, `
		SELECT id, workspace_id, agent_id, user_id, conversation_id, mode, status, trigger,
			COALESCE(trigger_payload,''), COALESCE(input,''), COALESCE(output,''),
			COALESCE(error_message,''), prompt_tokens, completion_tokens, total_tokens,
			tool_call_count, COALESCE(needs_input_reason,''), started_at, completed_at,
			created_at, updated_at
		FROM ai_agent_runs WHERE id = $1`, id,
	).Scan(&r.ID, &r.WorkspaceID, &r.AgentID, &r.UserID, &r.ConversationID, &r.Mode,
		&r.Status, &r.Trigger, &r.TriggerPayload, &r.Input, &r.Output, &r.ErrorMessage,
		&r.PromptTokens, &r.CompletionTokens, &r.TotalTokens, &r.ToolCallCount,
		&r.NeedsInputReason, &r.StartedAt, &r.CompletedAt, &r.CreatedAt, &r.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrRunNotFound
	}
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	return &r, nil
}

// ErrRunNotFound Run 不存在。
var ErrRunNotFound = errors.New("agent run not found")

// Transition 流转 Run 状态（带合法性校验）。
func (s *RunService) Transition(ctx context.Context, runID int64, to RunStatus, reason, output string) (*AgentRun, error) {
	if !IsValidRunStatus(string(to)) {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "status", Reason: "invalid value"})
	}
	if !validTransition(to) {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "status", Reason: "invalid transition target"})
	}

	var r AgentRun
	err := s.db.QueryRow(ctx, `
		UPDATE ai_agent_runs
		SET status = $1,
			needs_input_reason = COALESCE(NULLIF($2,''), needs_input_reason),
			output = COALESCE(NULLIF($3,''), output),
			started_at = CASE WHEN $1 = 'in_progress' AND started_at IS NULL THEN now() ELSE started_at END,
			completed_at = CASE WHEN $1 IN ('completed','stopped','failed','stale') THEN now() ELSE completed_at END,
			updated_at = now()
		WHERE id = $4 AND status NOT IN ('completed','stopped','failed','stale')
		RETURNING id, workspace_id, agent_id, user_id, conversation_id, mode, status, trigger,
			COALESCE(trigger_payload,''), COALESCE(input,''), COALESCE(output,''),
			COALESCE(error_message,''), prompt_tokens, completion_tokens, total_tokens,
			tool_call_count, COALESCE(needs_input_reason,''), started_at, completed_at,
			created_at, updated_at`,
		string(to), reason, output, runID,
	).Scan(&r.ID, &r.WorkspaceID, &r.AgentID, &r.UserID, &r.ConversationID, &r.Mode,
		&r.Status, &r.Trigger, &r.TriggerPayload, &r.Input, &r.Output, &r.ErrorMessage,
		&r.PromptTokens, &r.CompletionTokens, &r.TotalTokens, &r.ToolCallCount,
		&r.NeedsInputReason, &r.StartedAt, &r.CompletedAt, &r.CreatedAt, &r.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrRunNotFound
	}
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	return &r, nil
}

// validTransition 校验目标状态是否可达（终态不可再流转）。
func validTransition(to RunStatus) bool {
	switch to {
	case RunCreated, RunInProgress, RunAwaiting, RunCompleted,
		RunStopping, RunStopped, RunFailed, RunStale:
		return true
	}
	return false
}

// RecordUsage 记录 token 用量。
func (s *RunService) RecordUsage(ctx context.Context, runID int64, prompt, completion, toolCalls int) error {
	_, err := s.db.Exec(ctx, `
		UPDATE ai_agent_runs
		SET prompt_tokens = prompt_tokens + $1,
			completion_tokens = completion_tokens + $2,
			total_tokens = total_tokens + $1 + $2,
			tool_call_count = tool_call_count + $3,
			updated_at = now()
		WHERE id = $4`, prompt, completion, toolCalls, runID)
	if err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	return nil
}

// AddStep 追加执行步骤。
func (s *RunService) AddStep(ctx context.Context, runID int64, stepType, summary, detail string) (*RunStep, error) {
	var st RunStep
	err := s.db.QueryRow(ctx, `
		INSERT INTO ai_run_steps (id, run_id, step_type, summary, detail, created_at)
		VALUES ($1,$2,$3,$4,$5,now())
		RETURNING id, run_id, step_type, summary, COALESCE(detail,''), created_at`,
		genID(), runID, stepType, summary, nullIfEmpty(detail),
	).Scan(&st.ID, &st.RunID, &st.StepType, &st.Summary, &st.Detail, &st.CreatedAt)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	return &st, nil
}

// ListRuns 列出 Agent 的 Run。
func (s *RunService) ListRuns(ctx context.Context, agentID int64, limit int) ([]*AgentRun, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	rows, err := s.db.Query(ctx, `
		SELECT id, workspace_id, agent_id, user_id, conversation_id, mode, status, trigger,
			COALESCE(trigger_payload,''), COALESCE(input,''), COALESCE(output,''),
			COALESCE(error_message,''), prompt_tokens, completion_tokens, total_tokens,
			tool_call_count, COALESCE(needs_input_reason,''), started_at, completed_at,
			created_at, updated_at
		FROM ai_agent_runs
		WHERE agent_id = $1
		ORDER BY created_at DESC LIMIT $2`, agentID, limit)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()

	var out []*AgentRun
	for rows.Next() {
		var r AgentRun
		if err := rows.Scan(&r.ID, &r.WorkspaceID, &r.AgentID, &r.UserID, &r.ConversationID, &r.Mode,
			&r.Status, &r.Trigger, &r.TriggerPayload, &r.Input, &r.Output, &r.ErrorMessage,
			&r.PromptTokens, &r.CompletionTokens, &r.TotalTokens, &r.ToolCallCount,
			&r.NeedsInputReason, &r.StartedAt, &r.CompletedAt, &r.CreatedAt, &r.UpdatedAt); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
		out = append(out, &r)
	}
	return out, rows.Err()
}

// ListSteps 列出 Run 的执行步骤。
func (s *RunService) ListSteps(ctx context.Context, runID int64) ([]*RunStep, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, run_id, step_type, summary, COALESCE(detail,''), created_at
		FROM ai_run_steps WHERE run_id = $1 ORDER BY created_at ASC`, runID)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()

	var out []*RunStep
	for rows.Next() {
		var st RunStep
		if err := rows.Scan(&st.ID, &st.RunID, &st.StepType, &st.Summary, &st.Detail, &st.CreatedAt); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
		out = append(out, &st)
	}
	return out, rows.Err()
}

// StopRun 请求停止 Run（幂等）。
func (s *RunService) StopRun(ctx context.Context, runID int64) (*AgentRun, error) {
	return s.Transition(ctx, runID, RunStopping, "用户请求停止", "")
}

// --- Input ---

// CreateRunInput 创建 Run 入参。
type CreateRunInput struct {
	WorkspaceID    int64
	AgentID        int64
	UserID         int64
	ConversationID int64
	Mode           Mode
	Input          string
	Trigger        string
	TriggerPayload string
}

// --- migration ---

// RunMigrationSQL 返回建表 SQL（供 migrate 命令执行）。
const RunMigrationSQL = `
CREATE TABLE IF NOT EXISTS ai_agent_runs (
    id                BIGINT       PRIMARY KEY,
    workspace_id      BIGINT       NOT NULL,
    agent_id          BIGINT       NOT NULL,
    user_id           BIGINT       NOT NULL,
    conversation_id   BIGINT       NOT NULL DEFAULT 0,
    mode              VARCHAR(20)  NOT NULL,
    status            VARCHAR(20)  NOT NULL DEFAULT 'created',
    trigger           VARCHAR(20)  NOT NULL DEFAULT 'user',
    trigger_payload   TEXT,
    input             TEXT,
    output            TEXT,
    error_message     TEXT,
    prompt_tokens     INTEGER      NOT NULL DEFAULT 0,
    completion_tokens INTEGER      NOT NULL DEFAULT 0,
    total_tokens      INTEGER      NOT NULL DEFAULT 0,
    tool_call_count   INTEGER      NOT NULL DEFAULT 0,
    needs_input_reason TEXT,
    started_at        TIMESTAMPTZ,
    completed_at      TIMESTAMPTZ,
    created_at        TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ  NOT NULL DEFAULT now(),
    CONSTRAINT chk_ai_run_mode CHECK (mode IN ('ask','build','autopilot')),
    CONSTRAINT chk_ai_run_status CHECK (status IN
        ('created','in_progress','awaiting','completed','stopping','stopped','failed','stale'))
);

CREATE INDEX IF NOT EXISTS idx_ai_agent_runs_agent ON ai_agent_runs (agent_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_ai_agent_runs_status ON ai_agent_runs (status) WHERE status IN ('in_progress','awaiting','stopping');
CREATE INDEX IF NOT EXISTS idx_ai_agent_runs_workspace ON ai_agent_runs (workspace_id, created_at DESC);

CREATE TABLE IF NOT EXISTS ai_run_steps (
    id         BIGINT      PRIMARY KEY,
    run_id     BIGINT      NOT NULL,
    step_type  VARCHAR(20) NOT NULL,
    summary    VARCHAR(500) NOT NULL DEFAULT '',
    detail     TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT chk_ai_run_step_type CHECK (step_type IN ('thought','action','response','elicitation'))
);

CREATE INDEX IF NOT EXISTS idx_ai_run_steps_run ON ai_run_steps (run_id, created_at);

COMMENT ON TABLE ai_agent_runs IS 'Agent 执行记录（Run 生命周期）';
COMMENT ON TABLE ai_run_steps IS 'Agent 执行步骤（可观测轨迹）';
`

var _ = json.Marshal
var _ = time.Now
