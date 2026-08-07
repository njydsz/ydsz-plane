// Package issue — 内置状态模板集（项目创建时复制到 states + state_transitions 表）。
package issue

// --- 内置状态定义 ---

// DevFlowStates 研发流状态（requirement/task 默认）。
// 参考: GitHub Projects / Linear workflow
var DevFlowStates = []State{
	{Name: "Backlog", Group: GroupBacklog, Color: "#8DA2C2", Sequence: 1000},
	{Name: "Todo", Group: GroupBacklog, Color: "#A2B8D8", Sequence: 2000, IsDefault: true},
	{Name: "In Progress", Group: GroupStarted, Color: "#F59E0B", Sequence: 3000},
	{Name: "In Review", Group: GroupStarted, Color: "#8B5CF6", Sequence: 4000},
	{Name: "Done", Group: GroupCompleted, Color: "#10B981", Sequence: 5000},
	{Name: "Cancelled", Group: GroupCancelled, Color: "#9CA3AF", Sequence: 6000},
}

// DefectFlowStates 缺陷流状态。
var DefectFlowStates = []State{
	{Name: "New", Group: GroupBacklog, Color: "#EF4444", Sequence: 1000, IsDefault: true},
	{Name: "Confirmed", Group: GroupStarted, Color: "#F59E0B", Sequence: 2000},
	{Name: "In Progress", Group: GroupStarted, Color: "#3B82F6", Sequence: 3000},
	{Name: "Fixed", Group: GroupStarted, Color: "#8B5CF6", Sequence: 4000},
	{Name: "Verifying", Group: GroupStarted, Color: "#A855F7", Sequence: 5000},
	{Name: "Closed", Group: GroupCompleted, Color: "#10B981", Sequence: 6000},
	{Name: "Rejected", Group: GroupCancelled, Color: "#9CA3AF", Sequence: 7000},
	{Name: "Reopened", Group: GroupStarted, Color: "#F97316", Sequence: 8000},
}

// RequirementFlowStates 需求评审流。
var RequirementFlowStates = []State{
	{Name: "Draft", Group: GroupBacklog, Color: "#9CA3AF", Sequence: 1000, IsDefault: true},
	{Name: "Reviewing", Group: GroupStarted, Color: "#3B82F6", Sequence: 2000},
	{Name: "Accepted", Group: GroupCompleted, Color: "#10B981", Sequence: 3000},
	{Name: "Rejected", Group: GroupCancelled, Color: "#EF4444", Sequence: 4000},
	{Name: "Verified", Group: GroupCompleted, Color: "#059669", Sequence: 5000},
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
	"Fixed -> Verifying":     {"root_cause_category"},
	"Verifying -> Closed":    {"fix_version_id"},
	"New -> Rejected":        {"root_cause_category"},
	"Confirmed -> Rejected":  {"root_cause_category"},
}

// TemplateSet 项目状态模板套装（包含一套状态+流转）。
type TemplateSet struct {
	Name       string
	TypeCode   IssueTypeCode // 空=通用（requirement+task 共用 dev_flow）
	States     []State
	Transitions []TransitionKey
}

// DefaultTemplates 项目创建时默认使用的模板套装。
// 一个项目同时拥有：requirement/task 用 dev_flow，缺陷用 defect_flow。
var DefaultTemplates = []TemplateSet{
	{
		Name:       "dev_flow",
		TypeCode:   "", // 通用
		States:     DevFlowStates,
		Transitions: BuiltInTransitions["dev_flow"],
	},
	{
		Name:       "defect_flow",
		TypeCode:   TypeDefect,
		States:     DefectFlowStates,
		Transitions: BuiltInTransitions["defect_flow"],
	},
}
