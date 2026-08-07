// Package issue — 项目模板预设配置（Sprint 12 — v0.2）。
//
// 定义三种预设项目模板：敏捷(Agile)、瀑布(Waterfall)、通用(Generic)。
// 每种模板决定项目创建时注入的状态集与流转规则。
package issue

// ProjectTemplateCode 项目模板代码。
type ProjectTemplateCode string

const (
	TemplateAgile    ProjectTemplateCode = "agile"
	TemplateWaterfall ProjectTemplateCode = "waterfall"
	TemplateGeneric   ProjectTemplateCode = "generic"
)

// ProjectTemplate 项目模板定义（状态集 + 缺陷流绑定）。
type ProjectTemplate struct {
	Code        ProjectTemplateCode `json:"code"`
	Name        string              `json:"name"`
	Description string              `json:"description"`
	// ApplyDevFlow 是否注入研发流状态（requirement/task 通用）。
	ApplyDevFlow bool `json:"apply_dev_flow"`
	// ApplyDefectFlow 是否注入缺陷流状态。
	ApplyDefectFlow bool `json:"apply_defect_flow"`
	// ApplyRequirementFlow 是否注入需求评审流状态（需配合 ApplyDevFlow=true）。
	ApplyRequirementFlow bool `json:"apply_requirement_flow"`
}

// BuiltInProjectTemplates 预置模板清单（前端模板选择器数据源）。
var BuiltInProjectTemplates = []ProjectTemplate{
	{
		Code:                 TemplateAgile,
		Name:                 "敏捷研发",
		Description:          "Scrum 风格：产品待办 → 迭代冲刺，含完整缺陷管理流",
		ApplyDevFlow:         true,
		ApplyDefectFlow:      true,
		ApplyRequirementFlow: false,
	},
	{
		Code:                 TemplateWaterfall,
		Name:                 "瀑布交付",
		Description:          "V 模型：需求评审 → 开发 → 验证/验收，强调需求追踪",
		ApplyDevFlow:         true,
		ApplyDefectFlow:      false,
		ApplyRequirementFlow: true,
	},
	{
		Code:                 TemplateGeneric,
		Name:                 "通用看板",
		Description:          "轻量看板：待办 → 进行中 → 已完成，零配置起步",
		ApplyDevFlow:         true,
		ApplyDefectFlow:      false,
		ApplyRequirementFlow: false,
	},
}

// ValidateTemplateCode 校验模板代码是否合法。
func ValidateTemplateCode(code string) bool {
	switch ProjectTemplateCode(code) {
	case TemplateAgile, TemplateWaterfall, TemplateGeneric:
		return true
	default:
		return false
	}
}

// ProjectTemplateByCode 按代码查找预置模板。未找到时回退到 Generic。
func ProjectTemplateByCode(code string) ProjectTemplate {
	for _, t := range BuiltInProjectTemplates {
		if t.Code == ProjectTemplateCode(code) {
			return t
		}
	}
	return BuiltInProjectTemplates[2] // generic
}
