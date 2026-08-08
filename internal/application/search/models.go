// Package search — 全文搜索应用服务（PostgreSQL FTS + 多对象分组）。
//
// 对标:
//   - Plane: /api/workspaces/{slug}/search/
//   - Jira: /rest/api/latest/search with JQL
//   - Linear: search API with filters
//
// 设计要点:
//   - 搜索结果按对象类型分组（issues / sprints / versions）
//   - 支持高亮（PostgreSQL ts_headline）
//   - 支持过滤（状态/优先级/指派人/日期/类型）
//   - 搜索历史自动记录
//   - 收藏过滤器持久化
package search

import (
	"encoding/json"
	"time"
)

// --- Search Types ---

// DocType 搜索文档类型。
type DocType string

const (
	DocTypeIssue   DocType = "issue"
	DocTypeSprint  DocType = "sprint"
	DocTypeVersion DocType = "version"
)

// SearchQuery 搜索请求。
type SearchQuery struct {
	WorkspaceID int64    `json:"workspace_id"`
	ProjectID   int64    `json:"project_id"` // 0 = 全局搜索
	UserID      int64    `json:"user_id"`
	Query       string   `json:"query"` // 原始搜索词
	DocTypes    []string `json:"doc_types"` // 空 = 全部
	Filters     map[string]any `json:"filters"`
	Limit       int      `json:"limit"`
	Offset      int      `json:"offset"`
}

// SearchHit 单条搜索结果（JSON 对齐前端 SearchResultItem）。
type SearchHit struct {
	DocType     string  `json:"type"`                 // issue / sprint / version
	DocID       int64   `json:"id"`
	Title       string  `json:"name"`
	Identifier  string  `json:"identifier,omitempty"` // YD-123
	Description string  `json:"description"`          // 原始内容（供 debug）
	Highlight   string  `json:"highlight"`            // 取第一条高亮片段
	ProjectID   int64   `json:"project_id"`
	ProjectName string  `json:"project_name"`
	Rank        float64 `json:"rank"` // 相关性得分
	URL         string  `json:"url"`  // 前端跳转路径
}

// SearchResults 按类型分组的搜索结果（前端期望格式）。
type SearchResults struct {
	Issues   []SearchHit `json:"issues"`
	Sprints  []SearchHit `json:"sprints"`
	Versions []SearchHit `json:"versions"`
}

// MarshalJSON 确保空结果序列化为空数组 [] 而非 null，
// 以满足前端 TypeScript 代码和 E2E 测试的 Array.isArray 断言。
func (r SearchResults) MarshalJSON() ([]byte, error) {
	type alias SearchResults
	aux := struct{ alias }{alias: alias(r)}
	if aux.alias.Issues == nil {
		aux.alias.Issues = []SearchHit{}
	}
	if aux.alias.Sprints == nil {
		aux.alias.Sprints = []SearchHit{}
	}
	if aux.alias.Versions == nil {
		aux.alias.Versions = []SearchHit{}
	}
	return json.Marshal(aux.alias)
}

// SearchResponse 搜索响应。
type SearchResponse struct {
	Query       string        `json:"query"`
	Total       int           `json:"total"`        // 总命中数
	Results     SearchResults `json:"results"`      // 按类型分组（前端期望格式）
	Groups      []SearchGroup `json:"groups"`       // 保留向后兼容
	TimeMs      int64         `json:"time_ms"`      // 查询耗时
	Suggestions []string      `json:"suggestions"`  // 查询建议
	Backend     string        `json:"backend,omitempty"` // 搜索后端标识 (pg|es)，用于监控
	IsDegraded  bool          `json:"is_degraded,omitempty"` // JQL 解析是否降级
}

// SearchGroup 向后兼容的分组结果。
type SearchGroup struct {
	DocType string      `json:"doc_type"`
	Total   int64       `json:"total"`
	Hits    []SearchHit `json:"hits"`
}

// --- Search History ---

// SearchHistoryEntry 搜索历史记录。
type SearchHistoryEntry struct {
	ID          int64           `json:"id"`
	WorkspaceID int64           `json:"workspace_id"`
	UserID      int64           `json:"user_id"`
	Query       string          `json:"query"`
	Filters     map[string]any  `json:"filters"`
	ResultCount int             `json:"result_count"`
	SearchedAt  time.Time       `json:"searched_at"`
}

// RecordHistoryInput 记录搜索历史。
type RecordHistoryInput struct {
	WorkspaceID int64
	UserID      int64
	Query       string
	Filters     map[string]any
	ResultCount int
}

// --- Search Bookmark ---

// SearchBookmark 保存的搜索过滤器。
type SearchBookmark struct {
	ID          int64           `json:"id"`
	WorkspaceID int64           `json:"workspace_id"`
	ProjectID   *int64          `json:"project_id,omitempty"`
	UserID      int64           `json:"user_id"`
	Name        string          `json:"name"`
	Query       string          `json:"query"`
	Filters     map[string]any  `json:"filters"`
	IsShared    bool            `json:"is_shared"`
	SortOrder   float64         `json:"sort_order"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

// CreateBookmarkInput 创建收藏。
type CreateBookmarkInput struct {
	WorkspaceID int64
	ProjectID   *int64
	UserID      int64
	Name        string
	Query       string
	Filters     map[string]any
	IsShared    bool
}

// UpdateBookmarkInput 更新收藏。
type UpdateBookmarkInput struct {
	Name     *string
	Query    *string
	Filters  map[string]any
	IsShared *bool
}
