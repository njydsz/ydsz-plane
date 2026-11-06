// Package customer — 客户管理（CRM 层）应用域。
//
// 对标 Plane Business / HubSpot 轻量 CRM：
//   - Customer：客户记录（含行业、规模、收入、客户阶段、套餐、区域等自定义属性）；
//   - CustomerRequest：客户请求（关联到 intake 工单或正式工作项），
//     用于跟踪客户从「提出诉求」到「解决交付」的全生命周期。
//
// 数据表：customers / customer_requests（id 由应用层生成，见 genID）。
package customer

import (
	"crypto/rand"
	"encoding/binary"
	"time"
)

// CustomerStage 客户阶段。
type CustomerStage string

const (
	// StageLead 潜在客户（尚未签约）。
	StageLead CustomerStage = "lead"
	// StageProspect 意向客户（商务跟进中）。
	StageProspect CustomerStage = "prospect"
	// StageCustomer 正式客户（已签约）。
	StageCustomer CustomerStage = "customer"
	// StageChurned 流失客户。
	StageChurned CustomerStage = "churned"
)

// IsValid 校验客户阶段是否合法。
func (s CustomerStage) IsValid() bool {
	switch s {
	case StageLead, StageProspect, StageCustomer, StageChurned:
		return true
	}
	return false
}

// CustomerTier 客户套餐档位。
type CustomerTier string

const (
	// TierFree 免费版。
	TierFree CustomerTier = "free"
	// TierPro 专业版。
	TierPro CustomerTier = "pro"
	// TierEnterprise 企业版。
	TierEnterprise CustomerTier = "enterprise"
)

// IsValid 校验客户套餐档位是否合法。
func (t CustomerTier) IsValid() bool {
	switch t {
	case TierFree, TierPro, TierEnterprise:
		return true
	}
	return false
}

// RequestStatus 客户请求状态。
type RequestStatus string

const (
	// RequestNew 新建（待受理）。
	RequestNew RequestStatus = "new"
	// RequestAcknowledged 已受理（确认收到，待排期）。
	RequestAcknowledged RequestStatus = "acknowledged"
	// RequestInProgress 处理中。
	RequestInProgress RequestStatus = "in_progress"
	// RequestResolved 已解决。
	RequestResolved RequestStatus = "resolved"
	// RequestClosed 已关闭。
	RequestClosed RequestStatus = "closed"
)

// IsValid 校验请求状态是否合法。
func (s RequestStatus) IsValid() bool {
	switch s {
	case RequestNew, RequestAcknowledged, RequestInProgress, RequestResolved, RequestClosed:
		return true
	}
	return false
}

// RequestPriority 客户请求优先级。
type RequestPriority string

const (
	// PriorityUrgent 紧急。
	PriorityUrgent RequestPriority = "urgent"
	// PriorityHigh 高。
	PriorityHigh RequestPriority = "high"
	// PriorityMedium 中。
	PriorityMedium RequestPriority = "medium"
	// PriorityLow 低。
	PriorityLow RequestPriority = "low"
)

// IsValid 校验优先级是否合法。
func (p RequestPriority) IsValid() bool {
	switch p {
	case PriorityUrgent, PriorityHigh, PriorityMedium, PriorityLow:
		return true
	}
	return false
}

// RequestSource 客户请求来源。
type RequestSource string

const (
	// SourceEmail 邮件。
	SourceEmail RequestSource = "email"
	// SourceForm 表单。
	SourceForm RequestSource = "form"
	// SourceInApp 应用内。
	SourceInApp RequestSource = "in_app"
	// SourcePhone 电话。
	SourcePhone RequestSource = "phone"
)

// IsValid 校验请求来源是否合法。
func (s RequestSource) IsValid() bool {
	switch s {
	case SourceEmail, SourceForm, SourceInApp, SourcePhone:
		return true
	}
	return false
}

// Customer 客户记录。
type Customer struct {
	ID               int64         `json:"id"`
	WorkspaceID      int64         `json:"workspace_id"`
	Name             string        `json:"name"`
	Email            string        `json:"email"`
	Website          string        `json:"website"`
	Industry         string        `json:"industry"`
	Stage            CustomerStage `json:"stage"` // lead/prospect/customer/churned
	Tier             CustomerTier  `json:"tier"`  // free/pro/enterprise
	Region           string        `json:"region"`
	Employees        int           `json:"employees"`
	Revenue          string        `json:"revenue"`
	AccountManagerID int64         `json:"account_manager_id"`
	Notes            string        `json:"notes"`
	CreatedBy        int64         `json:"created_by"`
	CreatedAt        time.Time     `json:"created_at"`
	UpdatedAt        time.Time     `json:"updated_at"`
}

// CustomerRequest 客户请求（关联到 intake 工单或正式工作项）。
type CustomerRequest struct {
	ID           int64           `json:"id"`
	WorkspaceID  int64           `json:"workspace_id"`
	CustomerID   int64           `json:"customer_id"`
	Title        string          `json:"title"`
	Description  string          `json:"description"`
	Status       RequestStatus   `json:"status"`       // new/acknowledged/in_progress/resolved/closed
	Priority     RequestPriority `json:"priority"`     // urgent/high/medium/low
	Source       RequestSource   `json:"source"`       // email/form/in_app/phone
	WorkItemID   *int64          `json:"work_item_id"` // 关联的工作项
	WorkItemType string          `json:"work_item_type"`
	AssignedTo   int64           `json:"assigned_to"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}

// CustomerStats 单个客户的请求统计。
type CustomerStats struct {
	CustomerID int64          `json:"customer_id"`
	Total      int            `json:"total"`
	ByStatus   map[string]int `json:"by_status"`
}

// --- Input Types ---

// CreateCustomerInput 创建客户入参。
type CreateCustomerInput struct {
	WorkspaceID      int64
	Name             string
	Email            string
	Website          string
	Industry         string
	Stage            string
	Tier             string
	Region           string
	Employees        int
	Revenue          string
	AccountManagerID int64
	Notes            string
	CreatedBy        int64
}

// UpdateCustomerInput 更新客户入参（仅非 nil 字段被更新）。
type UpdateCustomerInput struct {
	Name             *string
	Email            *string
	Website          *string
	Industry         *string
	Stage            *string
	Tier             *string
	Region           *string
	Employees        *int
	Revenue          *string
	AccountManagerID *int64
	Notes            *string
}

// ListCustomersFilter 客户列表筛选。
type ListCustomersFilter struct {
	WorkspaceID int64
	Stage       string // lead/prospect/customer/churned；空=全部
	Tier        string // free/pro/enterprise；空=全部
	Industry    string
	Search      string
	Limit       int
	Offset      int
}

// CreateCustomerRequestInput 创建客户请求入参。
type CreateCustomerRequestInput struct {
	WorkspaceID  int64
	CustomerID   int64
	Title        string
	Description  string
	Status       string
	Priority     string
	Source       string
	WorkItemID   *int64
	WorkItemType string
	AssignedTo   int64
}

// UpdateCustomerRequestInput 更新客户请求入参（仅非 nil 字段被更新）。
type UpdateCustomerRequestInput struct {
	Title        *string
	Description  *string
	Status       *string
	Priority     *string
	Source       *string
	CustomerID   *int64
	WorkItemID   *int64
	WorkItemType *string
	AssignedTo   *int64
}

// ListCustomerRequestsFilter 客户请求列表筛选。
type ListCustomerRequestsFilter struct {
	WorkspaceID int64
	CustomerID  *int64
	Status      string
	Priority    string
	Limit       int
	Offset      int
}

// LinkRequestInput 关联工作项入参。
type LinkRequestInput struct {
	WorkspaceID  int64
	RequestID    int64
	WorkItemID   int64
	WorkItemType string // requirement/task/defect
}

// --- ID 生成 ---

// genID 生成应用层 BIGINT 主键（与其它应用层包策略一致）。
func genID() int64 {
	var b [8]byte
	_, _ = rand.Read(b[:])
	randBits := int64(binary.LittleEndian.Uint64(b[:]) & 0x3FFFFF)
	return (time.Now().UnixMilli() << 22) | randBits
}
