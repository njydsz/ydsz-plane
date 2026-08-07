// Package workspace — 项目模板域（Sprint 12 — v0.2）。
//
// 提供预置模板列表（agile/waterfall/generic），前端动态渲染项目创建向导。
package workspace

import (
	issueTpl "github.com/njydsz/ydsz-plane/internal/application/issue"
)

// TemplateService 项目模板域服务。
type TemplateService struct{}

// NewTemplateService 创建模板服务。
func NewTemplateService() *TemplateService {
	return &TemplateService{}
}

// ListTemplates 返回全部预置项目模板。
func (s *TemplateService) ListTemplates() []issueTpl.ProjectTemplate {
	return issueTpl.BuiltInProjectTemplates
}

// ValidateTemplate 校验模板代码是否合法（用于创建/更新时防御）。
func (s *TemplateService) ValidateTemplate(code string) bool {
	return issueTpl.ValidateTemplateCode(code)
}

// DefaultTemplateCode 返回默认模板代码。
func (s *TemplateService) DefaultTemplateCode() string {
	return string(issueTpl.TemplateGeneric)
}
