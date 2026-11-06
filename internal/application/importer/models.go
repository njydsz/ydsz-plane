// Package importer — 外部平台数据导入。
//
// 对标 Plane 的 Importers：将外部项目管理工具的数据迁移到 Ydsz Plane。
//
// 支持的来源：
//   - CSV（通用，自定义列映射）
//   - Jira（Cloud / Data Center，通过 REST API）
//   - Linear（通过 GraphQL API）
//   - Asana（通过 REST API）
//
// 设计：
//   - Importer 接口 + 各平台实现（解析 → 统一中间模型 → 校验）；
//   - 统一中间模型 WorkItemRecord 屏蔽平台差异；
//   - 字段映射可自定义（CSV 列映射 / 平台字段映射）。
package importer

import (
	"context"
	"errors"
	"time"
)

// Source 导入来源。
type Source string

const (
	// SourceCSV 通用 CSV。
	SourceCSV Source = "csv"
	// SourceJira Atlassian Jira。
	SourceJira Source = "jira"
	// SourceLinear Linear。
	SourceLinear Source = "linear"
	// SourceAsana Asana。
	SourceAsana Source = "asana"
)

// IsValidSource 校验来源。
func IsValidSource(s string) bool {
	switch Source(s) {
	case SourceCSV, SourceJira, SourceLinear, SourceAsana:
		return true
	}
	return false
}

// WorkItemRecord 统一工作项中间模型。
type WorkItemRecord struct {
	ExternalID       string         `json:"external_id"`
	Title            string         `json:"title"`
	Description      string         `json:"description"`
	TypeCode         string         `json:"type_code"`
	Priority         string         `json:"priority"`
	StateName        string         `json:"state_name"`
	AssigneeEmail    string         `json:"assignee_email"`
	CreatorEmail     string         `json:"creator_email"`
	Labels           []string       `json:"labels"`
	CreatedAt        *time.Time     `json:"created_at,omitempty"`
	UpdatedAt        *time.Time     `json:"updated_at,omitempty"`
	DueDate          *time.Time     `json:"due_date,omitempty"`
	ParentExternalID string         `json:"parent_external_id,omitempty"`
	Estimate         *float64       `json:"estimate,omitempty"`
	Raw              map[string]any `json:"-"`
}

// ImportOptions 导入选项。
type ImportOptions struct {
	WorkspaceID    int64
	ProjectID      int64
	CreatedBy      int64
	ColumnMapping  map[string]string
	DefaultType    string
	ImportChildren bool
	BatchSize      int
}

// ImportError 单条错误。
type ImportError struct {
	ExternalID string `json:"external_id,omitempty"`
	Title      string `json:"title,omitempty"`
	Reason     string `json:"reason"`
}

// ImportResult 导入结果。
type ImportResult struct {
	Source       Source        `json:"source"`
	TotalRecords int           `json:"total_records"`
	Imported     int           `json:"imported"`
	Skipped      int           `json:"skipped"`
	Failed       int           `json:"failed"`
	Errors       []ImportError `json:"errors,omitempty"`
	StartedAt    time.Time     `json:"started_at"`
	CompletedAt  *time.Time    `json:"completed_at,omitempty"`
}

// Importer 导入器接口。
type Importer interface {
	Source() Source
	Fetch(ctx context.Context, creds Credentials, opts ImportOptions) ([]WorkItemRecord, error)
}

// Credentials 外部平台凭证。
type Credentials struct {
	APIToken   string `json:"api_token"`
	Email      string `json:"email"`
	BaseURL    string `json:"base_url"`
	ProjectKey string `json:"project_key"`
	TeamID     string `json:"team_id"`
}

// ErrIncompleteCredentials 凭证不完整。
var ErrIncompleteCredentials = errors.New("credentials incomplete")

// Validate 校验凭证完整性。
func (c Credentials) Validate(src Source) error {
	switch src {
	case SourceJira:
		if c.BaseURL == "" || c.Email == "" || c.APIToken == "" {
			return ErrIncompleteCredentials
		}
	case SourceLinear:
		if c.APIToken == "" {
			return ErrIncompleteCredentials
		}
	case SourceAsana:
		if c.APIToken == "" {
			return ErrIncompleteCredentials
		}
	}
	return nil
}
