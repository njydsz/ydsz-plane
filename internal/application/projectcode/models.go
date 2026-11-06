// Package projectcode — Projects-as-Code（YAML 定义项目）域。
//
// 对标 Plane Compose：项目以 YAML 声明式定义（状态 / 标签 / 模块 / 成员 / 设置），
// 可纳入 Git 版本管理，并通过 API 或 CLI 拉取（export/pull）与下发（apply）。
//
// 能力分层：
//   - ParseSpec：YAML → ProjectSpec，并做结构 / 语义校验；
//   - ConfigStore：原始 YAML + 解析 JSON + sha256 + 版本号的持久化；
//   - Diff：既有项目 vs 规范的结构化差异；
//   - Apply：以规范为准创建 / 更新项目及其子资源（状态 / 标签 / 模块 / 成员）；
//   - Export：既有项目 → 规范（反向生成 YAML）。
//
// 数据表：
//   - project_code_configs（本域自有）
//   - projects / states / labels / modules / project_members（复用既有表）
package projectcode

import (
	"crypto/rand"
	"encoding/binary"
	"time"
)

// --- 规范结构（YAML / JSON 双向映射） ---

// ProjectSpec 项目 YAML 定义（解析后的规范结构）。
//
// YAML 示例：
//
//	version: "1.0"
//	name: "Apollo"
//	key: "APOLLO"
//	color: "#4B89FF"
//	icon: "🚀"
//	description: "Apollo program"
//	states:
//	  - name: "Todo"
//	    group: "backlog"
//	    color: "#8DA2C2"
//	labels:
//	  - name: "Backend"
//	    color: "#39758D"
//	modules:
//	  - name: "API"
//	    lead: "alice@example.com"
//	members:
//	  - email: "alice@example.com"
//	    role: "admin"
//	settings:
//	  default_assignee: "alice@example.com"
//	  estimate_points: [1,  2, 3, 5, 8]
type ProjectSpec struct {
	Version     string       `yaml:"version" json:"version"` // spec version, e.g. "1.0"
	Name        string       `yaml:"name" json:"name"`
	Key         string       `yaml:"key" json:"key"` // short identifier
	Color       string       `yaml:"color" json:"color"`
	Icon        string       `yaml:"icon" json:"icon"`
	Description string       `yaml:"description" json:"description"`
	States      []StateSpec  `yaml:"states" json:"states"`
	Labels      []LabelSpec  `yaml:"labels" json:"labels"`
	Modules     []ModuleSpec `yaml:"modules" json:"modules"`
	Members     []MemberSpec `yaml:"members" json:"members"`
	Settings    SpecSettings `yaml:"settings" json:"settings"`
}

// StateSpec 状态定义。
type StateSpec struct {
	Name  string `yaml:"name" json:"name"`
	Group string `yaml:"group" json:"group"` // backlog / started / completed / cancelled
	Color string `yaml:"color" json:"color"`
}

// LabelSpec 标签定义。
type LabelSpec struct {
	Name  string `yaml:"name" json:"name"`
	Color string `yaml:"color" json:"color"`
}

// ModuleSpec 模块定义。
type ModuleSpec struct {
	Name string `yaml:"name" json:"name"`
	Lead string `yaml:"lead" json:"lead"` // lead user email（可选）
}

// MemberSpec 成员定义。
type MemberSpec struct {
	Email string `yaml:"email" json:"email"`
	Role  string `yaml:"role" json:"role"` // admin / member
}

// SpecSettings 项目设置。
type SpecSettings struct {
	DefaultAssignee string `yaml:"default_assignee" json:"default_assignee"`
	EstimatePoints  []int  `yaml:"estimate_points" json:"estimate_points"`
}

// --- 持久化记录 ---

// Config 项目配置存储记录（原始 YAML + 解析 JSON + 版本）。
type Config struct {
	ID            int64        `json:"id"`
	WorkspaceID   int64        `json:"workspace_id"`
	ProjectID     int64        `json:"project_id,omitempty"`
	Name          string       `json:"name"`
	SpecKey       string       `json:"spec_key,omitempty"`
	Version       string       `json:"version"`
	YAML          string       `json:"yaml"`
	Spec          *ProjectSpec `json:"spec,omitempty"`
	SHA256        string       `json:"sha256"`
	ConfigVersion int          `json:"config_version"`
	CreatedBy     int64        `json:"created_by"`
	UpdatedBy     int64        `json:"updated_by"`
	CreatedAt     time.Time    `json:"created_at"`
	UpdatedAt     time.Time    `json:"updated_at"`
}

// --- 差异结构 ---

// ChangeKind 差异类型。
type ChangeKind string

const (
	// ChangeAdd 新增。
	ChangeAdd ChangeKind = "add"
	// ChangeUpdate 更新。
	ChangeUpdate ChangeKind = "update"
	// ChangeRemove 移除。
	ChangeRemove ChangeKind = "remove"
	// ChangeNoop 无变化。
	ChangeNoop ChangeKind = "noop"
)

// DiffItem 单条结构化差异。
type DiffItem struct {
	Kind     ChangeKind `json:"kind"`            // add / update / remove / noop
	Resource string     `json:"resource"`        // project / state / label / module / member / setting
	Name     string     `json:"name,omitempty"`  // 资源名（状态 / 标签 / 模块名等）
	Field    string     `json:"field,omitempty"` // 变更字段
	Old      any        `json:"old,omitempty"`   // 变更前值
	New      any        `json:"new,omitempty"`   // 变更后值
}

// DiffResult 差异汇总。
type DiffResult struct {
	ProjectID  int64      `json:"project_id,omitempty"`
	SpecName   string     `json:"spec_name"`
	HasChanges bool       `json:"has_changes"`
	Items      []DiffItem `json:"items"`
	// 便于 CLI 展示的分组计数
	AddCount    int `json:"add_count"`
	UpdateCount int `json:"update_count"`
	RemoveCount int `json:"remove_count"`
}

// --- 输入类型 ---

// SaveConfigInput 保存配置入参。
type SaveConfigInput struct {
	WorkspaceID int64
	ProjectID   int64 // 可选，关联既有项目
	Name        string
	YAML        string
	CreatedBy   int64
	UpdatedBy   int64
}

// ListConfigsOptions 列表查询选项。
type ListConfigsOptions struct {
	WorkspaceID int64
	ProjectID   int64 // 可选，按项目过滤
	Limit       int
	Offset      int
}

// ApplySpecInput 应用规范入参。
type ApplySpecInput struct {
	WorkspaceID int64
	ProjectID   int64 // 可选：>0 表示更新既有项目；0 表示新建
	Spec        *ProjectSpec
	CreatedBy   int64
	UpdatedBy   int64
}

// ApplyResult 应用结果。
type ApplyResult struct {
	ProjectID   int64  `json:"project_id"`
	Created     bool   `json:"created"`      // true=新建；false=更新
	ProjectSlug string `json:"project_slug"` // 项目 slug
	Identifier  string `json:"identifier"`
}

// --- ID 生成 ---

// genID 生成基于时间戳的伪随机 ID（与其它应用层包策略一致）。
func genID() int64 {
	var b [8]byte
	_, _ = rand.Read(b[:])
	randBits := int64(binary.LittleEndian.Uint64(b[:]) & 0x3FFFFF)
	return (time.Now().UnixMilli() << 22) | randBits
}
