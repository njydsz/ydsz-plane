// Package shared — 需求/任务/缺陷三域的共享类型与常量。
//
// S18 P0-9 issue 模块拆分：将原 issue 包中 cross-cutting 的类型下沉到此包，
// 使 requirement / task / defect 三个子域通过 import 引用，消除循环依赖风险。
//
// 本包仅放：枚举、接口定义、DTO struct、错误类型。禁止放业务逻辑。
package shared

// IssueTypeCode 复用主 issue 包的类型别名，避免导入循环。
// 新代码应直接使用本常量。
type IssueTypeCode = string

const (
	// TypeTask 任务类型标识
	TypeTask IssueTypeCode = "task"
	// TypeRequirement 需求类型标识
	TypeRequirement IssueTypeCode = "requirement"
	// TypeDefect 缺陷类型标识
	TypeDefect IssueTypeCode = "defect"
	// TypeEpic 史诗类型标识
	TypeEpic IssueTypeCode = "epic"
)

// IsKnownType 判断字符串是否为合法的工作项类型码。
// 替代原 issue 包中的 ad-hoc 判断逻辑。
func IsKnownType(code string) bool {
	switch code {
	case TypeTask, TypeRequirement, TypeDefect, TypeEpic:
		return true
	}
	return false
}
