// Package workflow — 工作流与审批域。
//
// 对标 Plane / Jira Workflow 的状态机能力：
//   - Workflow           —— 项目级工作流定义（状态流转规则的容器）；
//   - WorkflowTransition —— 状态流转规则（含审批配置与转换条件）；
//   - WorkflowApproval   —— 审批记录（pending / approved / rejected）。
//
// 数据表：
//   - workflows（id 由应用层 genID 生成）
//   - workflow_transitions（id 由应用层 genID 生成）
//   - workflow_approvals（id 由应用层 genID 生成）
//
// 设计要点：
//   - 一个项目可拥有多个工作流，其中至多一个为默认（is_default）；
//   - 流转规则校验：仅允许已定义的 from → to 流转；
//   - 审批门禁：requires_approval = true 时，执行流转前必须有已批准的审批记录；
//   - 多审批人：approver_ids 非空时仅列表内用户可审批，否则按 approver_role 角色匹配。
package workflow

import (
	"crypto/rand"
	"encoding/binary"
	"time"
)

// ApprovalStatus 审批状态枚举。
type ApprovalStatus string

const (
	// ApprovalPending 待审批。
	ApprovalPending ApprovalStatus = "pending"
	// ApprovalApproved 已批准。
	ApprovalApproved ApprovalStatus = "approved"
	// ApprovalRejected 已驳回。
	ApprovalRejected ApprovalStatus = "rejected"
)

// IsValid 校验审批状态是否合法。
func (s ApprovalStatus) IsValid() bool {
	switch s {
	case ApprovalPending, ApprovalApproved, ApprovalRejected:
		return true
	default:
		return false
	}
}

// IsTerminal 报告审批状态是否为终态（不再变化）。
func (s ApprovalStatus) IsTerminal() bool {
	switch s {
	case ApprovalApproved, ApprovalRejected:
		return true
	default:
		return false
	}
}

// 工作项类型常量（审批记录关联的工作项类型）。
const (
	WorkItemTypeRequirement = "requirement"
	WorkItemTypeTask        = "task"
	WorkItemTypeDefect      = "defect"
)

// IsValidWorkItemType 校验工作项类型是否合法。
func IsValidWorkItemType(t string) bool {
	switch t {
	case WorkItemTypeRequirement, WorkItemTypeTask, WorkItemTypeDefect:
		return true
	default:
		return false
	}
}

// Workflow 项目级工作流定义。
type Workflow struct {
	ID          int64     `json:"id"`
	ProjectID   int64     `json:"project_id"`
	Name        string    `json:"name"` // e.g. "默认工作流"
	Description string    `json:"description,omitempty"`
	IsDefault   bool      `json:"is_default"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// WorkflowTransition 状态流转规则。
type WorkflowTransition struct {
	ID          int64  `json:"id"`
	WorkflowID  int64  `json:"workflow_id"`
	FromStateID int64  `json:"from_state_id"`
	ToStateID   int64  `json:"to_state_id"`
	Name        string `json:"name"` // transition label
	// 审批配置
	RequiresApproval bool    `json:"requires_approval"`
	ApproverRole     string  `json:"approver_role,omitempty"` // admin/member/owner or empty
	ApproverIDs      []int64 `json:"approver_ids,omitempty"`  // specific approvers
	// 转换条件
	Conditions string    `json:"conditions,omitempty"` // JSON conditions (optional)
	CreatedAt  time.Time `json:"created_at"`
}

// WorkflowApproval 审批记录。
type WorkflowApproval struct {
	ID           int64          `json:"id"`
	TransitionID int64          `json:"transition_id"`
	WorkItemID   int64          `json:"work_item_id"`
	WorkItemType string         `json:"work_item_type"`
	RequesterID  int64          `json:"requester_id"`
	ApproverID   int64          `json:"approver_id"`
	Status       ApprovalStatus `json:"status"` // pending/approved/rejected
	Comment      string         `json:"comment,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
	ResolvedAt   *time.Time     `json:"resolved_at,omitempty"`
}

// WorkflowDetail 工作流详情（含全部流转规则）。
type WorkflowDetail struct {
	*Workflow
	Transitions []WorkflowTransition `json:"transitions"`
}

// TransitionHistory 工作项的流转历史记录（审批 + 流转快照）。
type TransitionHistory struct {
	ApprovalID   int64          `json:"approval_id"`
	TransitionID int64          `json:"transition_id"`
	FromStateID  int64          `json:"from_state_id"`
	ToStateID    int64          `json:"to_state_id"`
	Name         string         `json:"name,omitempty"`
	Status       ApprovalStatus `json:"status"`
	RequesterID  int64          `json:"requester_id"`
	ApproverID   int64          `json:"approver_id"`
	Comment      string         `json:"comment,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
	ResolvedAt   *time.Time     `json:"resolved_at,omitempty"`
}

// --- Input Types ---

// CreateWorkflowInput 创建流转入参。
type CreateWorkflowInput struct {
	ProjectID   int64
	Name        string
	Description string
	IsDefault   bool
	CreatedBy   int64
}

// UpdateWorkflowInput 更新流转入参（字段为指针，nil 表示不更新）。
type UpdateWorkflowInput struct {
	Name        *string
	Description *string
	IsDefault   *bool
}

// ListWorkflowsOptions 工作流列表查询选项。
type ListWorkflowsOptions struct {
	ProjectID int64
	Limit     int
	Offset    int
}

// CreateTransitionInput 创建流转规则入参。
type CreateTransitionInput struct {
	WorkflowID       int64
	FromStateID      int64
	ToStateID        int64
	Name             string
	RequiresApproval bool
	ApproverRole     string
	ApproverIDs      []int64
	Conditions       string
}

// ValidateTransitionInput 校验流转入参。
type ValidateTransitionInput struct {
	WorkflowID  int64
	FromStateID int64
	ToStateID   int64
	WorkItemID  int64
}

// ValidateResult 流转校验结果。
type ValidateResult struct {
	Allowed          bool   `json:"allowed"`
	RequiresApproval bool   `json:"requires_approval"`
	Reason           string `json:"reason,omitempty"`
}

// SubmitApprovalInput 提交审批入参。
type SubmitApprovalInput struct {
	TransitionID int64
	WorkItemID   int64
	WorkItemType string
	RequesterID  int64
	ApproverID   int64
	Comment      string
}

// ResolveApprovalInput 审批处理入参。
type ResolveApprovalInput struct {
	ApprovalID int64
	ApproverID int64
	Comment    string
}

// ExecuteTransitionInput 执行流转入参。
type ExecuteTransitionInput struct {
	WorkflowID   int64
	TransitionID int64
	WorkItemID   int64
	WorkItemType string
	OperatorID   int64
}

// ListPendingApprovalsOptions 待审批列表查询选项。
type ListPendingApprovalsOptions struct {
	WorkspaceID int64
	UserID      int64
	Limit       int
	Offset      int
}

// --- ID 生成 ---

// genID 生成基于时间戳的伪随机 ID（与 calendar / intake / datajob 包策略一致）。
//
// 结构：高 42 位为毫秒时间戳，低 22 位为密码学随机数。
// 同一毫秒内可生成 4,194,304 个不重复 ID，时间戳保证跨毫秒单调递增。
func genID() int64 {
	var b [8]byte
	_, _ = rand.Read(b[:])
	randBits := int64(binary.LittleEndian.Uint64(b[:]) & 0x3FFFFF)
	return (time.Now().UnixMilli() << 22) | randBits
}
