// Package issue — 内置状态模板集（项目创建时复制到 states + state_transitions 表）。
package issue

// --- 内置状态定义 ---

// RequirementDevFlowStates 需求研发流状态。
var RequirementDevFlowStates = []State{
	{Name: "Backlog", Group: GroupBacklog, Color: "#8DA2C2", Sequence: 1000, ApplicableTypes: []string{"requirement"}},
	{Name: "Todo", Group: GroupBacklog, Color: "#A2B8D8", Sequence: 2000, IsDefault: true, ApplicableTypes: []string{"requirement"}},
	{Name: "In Progress", Group: GroupStarted, Color: "#F59E0B", Sequence: 3000, ApplicableTypes: []string{"requirement"}},
	{Name: "In Review", Group: GroupStarted, Color: "#8B5CF6", Sequence: 4000, ApplicableTypes: []string{"requirement"}},
	{Name: "Done", Group: GroupCompleted, Color: "#10B981", Sequence: 5000, ApplicableTypes: []string{"requirement"}},
	{Name: "Cancelled", Group: GroupCancelled, Color: "#9CA3AF", Sequence: 6000, ApplicableTypes: []string{"requirement"}},
}

// TaskDevFlowStates 任务研发流状态。
var TaskDevFlowStates = []State{
	{Name: "Backlog", Group: GroupBacklog, Color: "#8DA2C2", Sequence: 1000, ApplicableTypes: []string{"task"}},
	{Name: "Todo", Group: GroupBacklog, Color: "#A2B8D8", Sequence: 2000, IsDefault: true, ApplicableTypes: []string{"task"}},
	{Name: "In Progress", Group: GroupStarted, Color: "#F59E0B", Sequence: 3000, ApplicableTypes: []string{"task"}},
	{Name: "In Review", Group: GroupStarted, Color: "#8B5CF6", Sequence: 4000, ApplicableTypes: []string{"task"}},
	{Name: "Done", Group: GroupCompleted, Color: "#10B981", Sequence: 5000, ApplicableTypes: []string{"task"}},
	{Name: "Cancelled", Group: GroupCancelled, Color: "#9CA3AF", Sequence: 6000, ApplicableTypes: []string{"task"}},
}

// DefectFlowStates 缺陷流状态。
var DefectFlowStates = []State{
	{Name: "New", Group: GroupBacklog, Color: "#EF4444", Sequence: 1000, IsDefault: true, ApplicableTypes: []string{"defect"}},
	{Name: "Confirmed", Group: GroupStarted, Color: "#F59E0B", Sequence: 2000, ApplicableTypes: []string{"defect"}},
	{Name: "In Progress", Group: GroupStarted, Color: "#3B82F6", Sequence: 3000, ApplicableTypes: []string{"defect"}},
	{Name: "Fixed", Group: GroupStarted, Color: "#8B5CF6", Sequence: 4000, ApplicableTypes: []string{"defect"}},
	{Name: "Verifying", Group: GroupStarted, Color: "#A855F7", Sequence: 5000, ApplicableTypes: []string{"defect"}},
	{Name: "Closed", Group: GroupCompleted, Color: "#10B981", Sequence: 6000, ApplicableTypes: []string{"defect"}},
	{Name: "Rejected", Group: GroupCancelled, Color: "#9CA3AF", Sequence: 7000, ApplicableTypes: []string{"defect"}},
	{Name: "Reopened", Group: GroupStarted, Color: "#F97316", Sequence: 8000, ApplicableTypes: []string{"defect"}},
}

// EpicFlowStates 史诗流状态（顶层容器，对标 Plane Epic 工作流）。
var EpicFlowStates = []State{
	{Name: "Planning", Group: GroupBacklog, Color: "#6366F1", Sequence: 1000, IsDefault: true, ApplicableTypes: []string{"epic"}},
	{Name: "In Progress", Group: GroupStarted, Color: "#F59E0B", Sequence: 2000, ApplicableTypes: []string{"epic"}},
	{Name: "Done", Group: GroupCompleted, Color: "#10B981", Sequence: 3000, ApplicableTypes: []string{"epic"}},
	{Name: "Cancelled", Group: GroupCancelled, Color: "#9CA3AF", Sequence: 4000, ApplicableTypes: []string{"epic"}},
}

// RequirementReviewFlowStates 需求评审流（可选启用）。
var RequirementReviewFlowStates = []State{
	{Name: "Draft", Group: GroupBacklog, Color: "#9CA3AF", Sequence: 1000, IsDefault: true, ApplicableTypes: []string{"requirement"}},
	{Name: "Reviewing", Group: GroupStarted, Color: "#3B82F6", Sequence: 2000, ApplicableTypes: []string{"requirement"}},
	{Name: "Accepted", Group: GroupCompleted, Color: "#10B981", Sequence: 3000, ApplicableTypes: []string{"requirement"}},
	{Name: "Rejected", Group: GroupCancelled, Color: "#EF4444", Sequence: 4000, ApplicableTypes: []string{"requirement"}},
	{Name: "Verified", Group: GroupCompleted, Color: "#059669", Sequence: 5000, ApplicableTypes: []string{"requirement"}},
}

// --- 流转规则 ---

// TransitionKey 用于查找唯一流转规则。
type TransitionKey struct {
	From string
	To   string
}

// BuiltInTransitions 内置流转规则（按状态名索引）。
// 使用状态名而非 ID，方便项目创建后 ID 映射。
var BuiltInTransitions = map[string][]TransitionKey{
	"dev_flow": {
		{From: "Backlog", To: "Todo"},
		{From: "Backlog", To: "In Progress"},
		{From: "Todo", To: "Backlog"},
		{From: "Todo", To: "In Progress"},
		{From: "Todo", To: "Done"},
		{From: "In Progress", To: "Todo"},
		{From: "In Progress", To: "In Review"},
		{From: "In Progress", To: "Done"},
		{From: "In Review", To: "In Progress"},
		{From: "In Review", To: "Done"},
		{From: "Done", To: "In Progress"},
		{From: "Done", To: "Cancelled"},
		{From: "*", To: "Cancelled"}, // 任意状态都可取消
	},
	"defect_flow": {
		{From: "New", To: "Confirmed"},
		{From: "New", To: "Rejected"},
		{From: "Confirmed", To: "In Progress"},
		{From: "Confirmed", To: "Rejected"},
		{From: "In Progress", To: "Fixed"},
		{From: "In Progress", To: "Rejected"},
		{From: "Fixed", To: "Verifying"},
		{From: "Verifying", To: "Closed"},
		{From: "Verifying", To: "Rejected"}, // 验证不通过 = 驳回（到 Rejected 或回到 In Progress）
		{From: "Closed", To: "Reopened"},
		{From: "Rejected", To: "Reopened"},
		{From: "*", To: "Rejected"},
	},
	"epic_flow": {
		{From: "Planning", To: "In Progress"},
		{From: "Planning", To: "Done"},
		{From: "Planning", To: "Cancelled"},
		{From: "In Progress", To: "Planning"},
		{From: "In Progress", To: "Done"},
		{From: "In Progress", To: "Cancelled"},
		{From: "Done", To: "In Progress"},
		{From: "*", To: "Cancelled"},
	},
	"requirement_flow": {
		{From: "Draft", To: "Reviewing"},
		{From: "Reviewing", To: "Accepted"},
		{From: "Reviewing", To: "Rejected"},
		{From: "Accepted", To: "Verified"},
		{From: "Rejected", To: "Draft"}, // 被拒后可重新起草
	},
}

// RequiredFieldsForTransition 状态流转必填字段（某些流转要求字段必填）。
// key: "from_state_name -> to_state_name"
var RequiredFieldsForTransition = map[string][]string{
	"Fixed -> Verifying":    {"root_cause_category"},
	"Verifying -> Closed":   {"fix_version_id"},
	"New -> Rejected":       {"root_cause_category"},
	"Confirmed -> Rejected": {"root_cause_category"},
}

// TemplateSet 项目状态模板套装（包含一套状态+流转）。
type TemplateSet struct {
	Name        string
	TypeCode    IssueTypeCode // 空=通用（requirement+task 共用 dev_flow）
	States      []State
	Transitions []TransitionKey
}

// DefaultTemplates 项目创建时默认使用的模板套装。
// 多态工作项模型（S12 重构）下，需求与任务各自维护独立状态集
// （RequirementDevFlowStates / TaskDevFlowStates），缺陷用 defect_flow。
// 生产路径（project_init.go）按类型分别复制；本表供测试与文档校验。
var DefaultTemplates = []TemplateSet{
	{
		Name:        "epic_flow",
		TypeCode:    TypeEpic,
		States:      EpicFlowStates,
		Transitions: BuiltInTransitions["epic_flow"],
	},
	{
		Name:        "requirement_dev_flow",
		TypeCode:    TypeRequirement,
		States:      RequirementDevFlowStates,
		Transitions: BuiltInTransitions["dev_flow"],
	},
	{
		Name:        "task_dev_flow",
		TypeCode:    TypeTask,
		States:      TaskDevFlowStates,
		Transitions: BuiltInTransitions["dev_flow"],
	},
	{
		Name:        "defect_flow",
		TypeCode:    TypeDefect,
		States:      DefectFlowStates,
		Transitions: BuiltInTransitions["defect_flow"],
	},
}
