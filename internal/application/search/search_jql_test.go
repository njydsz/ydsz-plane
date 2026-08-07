// Package search JQL 子句 → 过滤器映射测试。
package search

import (
	"reflect"
	"testing"

	"github.com/njydsz/ydsz-plane/pkg/searchql"
)

// TestMergeJQLClauseToFilters 验证类 JQL 子句正确合并到 Filters map。
// 覆盖 searchql 支持的字段映射，确保 JQL 查询（如 project:TP type:task）能驱动 PG FTS 过滤。
func TestMergeJQLClauseToFilters(t *testing.T) {
	tests := []struct {
		name   string
		clause searchql.Clause
		key    string
		want   any
	}{
		{"project", searchql.Clause{Field: "project", Value: "CORE"}, "project_identifier", "CORE"},
		{"type", searchql.Clause{Field: "type", Value: "task"}, "type_code", "task"},
		{"status", searchql.Clause{Field: "status", Value: "done"}, "state_name", "done"},
		{"priority", searchql.Clause{Field: "priority", Value: "urgent"}, "priority", "urgent"},
		{"assignee", searchql.Clause{Field: "assignee", Value: "dev"}, "assignee_id", "dev"},
		{"reporter", searchql.Clause{Field: "reporter", Value: "pm"}, "created_by", "pm"},
		{"sprint", searchql.Clause{Field: "sprint", Value: "S1"}, "sprint_id", "S1"},
		{"version", searchql.Clause{Field: "version", Value: "1.0.0"}, "version_id", "1.0.0"},
		{"identifier", searchql.Clause{Field: "identifier", Value: "TP-1"}, "identifier", "TP-1"},
		{"negated value still mapped", searchql.Clause{Field: "type", Value: "bug", Negated: true}, "type_code", "bug"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			filters := map[string]any{}
			mergeJQLClauseToFilters(filters, tc.clause)
			if got, ok := filters[tc.key]; !ok {
				t.Errorf("key %q not set, filters=%v", tc.key, filters)
			} else if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("filters[%q] = %v, want %v", tc.key, got, tc.want)
			}
		})
	}
}

// TestMergeJQLClauseToFiltersUnknownField 未知字段不应污染过滤器。
func TestMergeJQLClauseToFiltersUnknownField(t *testing.T) {
	filters := map[string]any{"existing": "kept"}
	mergeJQLClauseToFilters(filters, searchql.Clause{Field: "unknown_field", Value: "x"})
	if len(filters) != 1 {
		t.Errorf("unknown field polluted filters: %v", filters)
	}
	if filters["existing"] != "kept" {
		t.Errorf("existing filter lost: %v", filters)
	}
}
