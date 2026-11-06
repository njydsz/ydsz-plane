// Package collaboration — 实时协作域。
//
// 对标 Plane 的实时协作体验：
//   - 文档编辑会话（谁在编辑哪篇文档）；
//   - 字段级乐观锁（避免并发覆盖）；
//   - 变更实时广播（通过现有 ws.Hub）；
//   - 在线编辑者指示器。
//
// 设计取舍：
//   完整 CRDT（Y.js / Automerge）需要引入额外运行时与持久化层，
//   本实现采用「会话 + 版本向量 + 操作广播」的轻量方案，覆盖
//   Plane 协作体验的核心交互：多人同时可见、变更即时同步、
//   冲突可检测。后续可平滑替换为 CRDT 传输层。
//
// 数据表：
//   - collaboration_sessions
//   - collaboration_operations（操作日志，用于回放与冲突检测）
package collaboration

import (
	"crypto/rand"
	"encoding/binary"
	"time"
)

// DocumentType 文档类型。
type DocumentType string

const (
	// DocTypeProjectPage 项目文档。
	DocTypeProjectPage DocumentType = "project_page"
	// DocTypeKnowledgePage 知识库文档。
	DocTypeKnowledgePage DocumentType = "knowledge_page"
	// DocTypeIssueDescription 工作项描述。
	DocTypeIssueDescription DocumentType = "issue_description"
	// DocTypeSprintGoal 迭代目标。
	DocTypeSprintGoal DocumentType = "sprint_goal"
)

// IsValid 校验文档类型。
func (t DocumentType) IsValid() bool {
	switch t {
	case DocTypeProjectPage, DocTypeKnowledgePage, DocTypeIssueDescription, DocTypeSprintGoal:
		return true
	}
	return false
}

// SessionStatus 会话状态。
type SessionStatus string

const (
	// SessionActive 活跃编辑中。
	SessionActive SessionStatus = "active"
	// SessionIdle 空闲（超时未操作）。
	SessionIdle SessionStatus = "idle"
	// SessionClosed 已关闭。
	SessionClosed SessionStatus = "closed"
)

// CollaborationSession 文档编辑会话。
type CollaborationSession struct {
	ID          int64         `json:"id"`
	WorkspaceID int64         `json:"workspace_id"`
	DocumentType DocumentType `json:"document_type"`
	DocumentID  int64         `json:"document_id"`
	UserID      int64         `json:"user_id"`
	UserName    string        `json:"user_name,omitempty"`
	DisplayName string        `json:"display_name,omitempty"`
	Status      SessionStatus `json:"status"`
	// 客户端当前持有的文档版本（用于乐观锁）
	BaseVersion int64     `json:"base_version"`
	LastSeenAt  time.Time `json:"last_seen_at"`
	CreatedAt   time.Time `json:"created_at"`
}

// Operation 协作操作。
type Operation struct {
	ID           int64        `json:"id"`
	WorkspaceID  int64        `json:"workspace_id"`
	DocumentType DocumentType `json:"document_type"`
	DocumentID   int64        `json:"document_id"`
	UserID       int64        `json:"user_id"`
	UserName     string       `json:"user_name,omitempty"`
	// 操作类型：join / leave / edit / lock / unlock / cursor
	OpType string `json:"op_type"`
	// 基于版本（乐观锁）
	BaseVersion int64 `json:"base_version"`
	// 目标版本
	TargetVersion int64     `json:"target_version"`
	// 变更内容摘要（用于广播）
	Payload   string    `json:"payload,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// DocumentVersion 文档版本（乐观锁）。
type DocumentVersion struct {
	ID           int64        `json:"id"`
	DocumentType DocumentType `json:"document_type"`
	DocumentID   int64        `json:"document_id"`
	Version      int64        `json:"version"`
	UpdatedBy    int64        `json:"updated_by"`
	UpdatedAt    time.Time    `json:"updated_at"`
}

// --- Input Types ---

// JoinSessionInput 加入会话入参。
type JoinSessionInput struct {
	WorkspaceID  int64
	DocumentType DocumentType
	DocumentID   int64
	UserID       int64
	UserName     string
}

// HeartbeatInput 心跳入参。
type HeartbeatInput struct {
	WorkspaceID  int64
	DocumentType DocumentType
	DocumentID   int64
	UserID       int64
	BaseVersion  int64
}

// SubmitOperationInput 提交操作入参。
type SubmitOperationInput struct {
	WorkspaceID  int64
	DocumentType DocumentType
	DocumentID   int64
	UserID       int64
	UserName     string
	OpType       string
	BaseVersion  int64
	Payload      string
}

// ListSessionsOptions 会话查询选项。
type ListSessionsOptions struct {
	WorkspaceID  int64
	DocumentType DocumentType
	DocumentID   int64
	Limit        int
}

// --- 结果类型 ---

// SubmitResult 操作提交结果。
type SubmitResult struct {
	Accepted   bool       `json:"accepted"`
	NewVersion int64      `json:"new_version"`
	Reason     string     `json:"reason,omitempty"`
	Operation  *Operation `json:"operation,omitempty"`
}

// --- ID 生成 ---

// genID 生成基于时间戳的伪随机 ID。
func genID() int64 {
	var b [8]byte
	_, _ = rand.Read(b[:])
	randBits := int64(binary.LittleEndian.Uint64(b[:]) & 0x3FFFFF)
	return (time.Now().UnixMilli() << 22) | randBits
}
