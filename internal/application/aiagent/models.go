// Package aiagent — Plane AI 风格的 Agent 能力域。
//
// 对标 Plane AI 的三种模式：
//   - Ask：只读问答（基于工作空间上下文回答）；
//   - Build：生成行动计划并以「动作卡片」形式待用户确认；
//   - Autopilot：直接执行动作。
//
// 另含：
//   - AI Memory：用户/项目/工作空间级别的持久化上下文，供模型引用；
//   - Agent 定义：可配置的助手（身份、指令、触发器、技能、记忆、额度）。
//
// 数据表：
//   - ai_agents
//   - ai_memory
//   - ai_conversations
//   - ai_messages
//   - ai_actions
package aiagent

import (
	"crypto/rand"
	"encoding/binary"
	"time"
)

// Mode AI 交互模式。
type Mode string

const (
	// ModeAsk 只读问答。
	ModeAsk Mode = "ask"
	// ModeBuild 生成待确认的行动计划。
	ModeBuild Mode = "build"
	// ModeAutopilot 直接执行。
	ModeAutopilot Mode = "autopilot"
)

// IsValidMode 校验模式。
func IsValidMode(m string) bool {
	switch Mode(m) {
	case ModeAsk, ModeBuild, ModeAutopilot:
		return true
	}
	return false
}

// Agent AI 助手定义。
type Agent struct {
	ID                 int64     `json:"id"`
	WorkspaceID        int64     `json:"workspace_id"`
	Name               string    `json:"name"`
	Description        string    `json:"description"`
	Instructions       string    `json:"instructions"`
	Triggers           []string  `json:"triggers"`
	Skills             []string  `json:"skills"`
	ScopeProjectIDs    []int64   `json:"scope_project_ids"`
	ScopeWorkItemTypes []string  `json:"scope_work_item_types"`
	Model              string    `json:"model"`
	Temperature        float64   `json:"temperature"`
	DailyCreditLimit   int       `json:"daily_credit_limit"`
	UsedCreditsToday   int       `json:"used_credits_today"`
	Status             string    `json:"status"`
	CreatedBy          int64     `json:"created_by"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// MemoryScope 记忆作用域。
type MemoryScope string

const (
	// ScopeUser 用户级记忆。
	ScopeUser MemoryScope = "user"
	// ScopeProject 项目级记忆。
	ScopeProject MemoryScope = "project"
	// ScopeWorkspace 工作空间级记忆。
	ScopeWorkspace MemoryScope = "workspace"
)

// AIWorkspaceMemory 持久化的 AI 上下文记忆。
type AIWorkspaceMemory struct {
	ID          int64       `json:"id"`
	WorkspaceID int64       `json:"workspace_id"`
	Scope       MemoryScope `json:"scope"`
	ScopeID     int64       `json:"scope_id"`
	Key         string      `json:"key"`
	Content     string      `json:"content"`
	Source      string      `json:"source"`
	CreatedBy   int64       `json:"created_by"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

// Conversation AI 会话。
type Conversation struct {
	ID          int64     `json:"id"`
	WorkspaceID int64     `json:"workspace_id"`
	AgentID     int64     `json:"agent_id"`
	UserID      int64     `json:"user_id"`
	Mode        Mode      `json:"mode"`
	Title       string    `json:"title"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Message 会话消息。
type Message struct {
	ID             int64      `json:"id"`
	ConversationID int64      `json:"conversation_id"`
	Role           string     `json:"role"`
	Content        string     `json:"content"`
	Actions        []AIAction `json:"actions,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
}

// ActionType 动作类型。
type ActionType string

const (
	ActionCreateWorkItem ActionType = "create_work_item"
	ActionUpdateWorkItem ActionType = "update_work_item"
	ActionCreateCycle    ActionType = "create_cycle"
	ActionCreateModule   ActionType = "create_module"
	ActionCreatePage     ActionType = "create_page"
	ActionUpdateState    ActionType = "update_state"
	ActionAssign         ActionType = "assign"
	ActionAddComment     ActionType = "add_comment"
	ActionLogTime        ActionType = "log_time"
	ActionBulkUpdate     ActionType = "bulk_update"
	ActionDeleteWorkItem ActionType = "delete_work_item"
)

// AIAction 待执行/已执行的动作。
type AIAction struct {
	ID           int64      `json:"id"`
	MessageID    int64      `json:"message_id"`
	Type         ActionType `json:"type"`
	Payload      string     `json:"payload"`
	Status       string     `json:"status"`
	Result       string     `json:"result,omitempty"`
	ErrorMessage string     `json:"error_message,omitempty"`
	ExecutedBy   int64      `json:"executed_by,omitempty"`
	ExecutedAt   *time.Time `json:"executed_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
}

// --- Input Types ---

// ChatInput 对话入参。
type ChatInput struct {
	WorkspaceID    int64
	UserID         int64
	AgentID        int64
	Mode           Mode
	Message        string
	ProjectID      int64
	ConversationID int64
}

// AgentInput 创建/更新 Agent。
type AgentInput struct {
	WorkspaceID        int64
	Name               string
	Description        string
	Instructions       string
	Triggers           []string
	Skills             []string
	ScopeProjectIDs    []int64
	ScopeWorkItemTypes []string
	Model              string
	Temperature        float64
	DailyCreditLimit   int
	CreatedBy          int64
}

// --- ID 生成 ---

func genID() int64 {
	var b [8]byte
	_, _ = rand.Read(b[:])
	randBits := int64(binary.LittleEndian.Uint64(b[:]) & 0x3FFFFF)
	return (time.Now().UnixMilli() << 22) | randBits
}
