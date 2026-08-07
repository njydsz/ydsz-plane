// Package workbench 工作台纯逻辑测试：URL 映射。
package workbench

import "testing"

// TestBuildItemURL 验证工作台各类快捷项的前端跳转 URL 映射。
// 对齐 search.buildDocURL 的约定，保证工作台/搜索结果跳转一致。
func TestBuildItemURL(t *testing.T) {
	tests := []struct {
		name      string
		itemType  string
		itemID    int64
		projectID int64
		want      string
	}{
		{"issue", "issue", 10, 5, "/projects/5/issues/10"},
		{"sprint", "sprint", 11, 5, "/projects/5/sprints/11"},
		{"version", "version", 12, 5, "/projects/5/versions/12"},
		{"project board", "project", 7, 7, "/projects/7/board"},
		{"unknown type", "user", 1, 2, ""},
		{"empty type", "", 1, 2, ""},
		{"zero ids", "issue", 0, 0, "/projects/0/issues/0"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := buildItemURL(tc.itemType, tc.itemID, tc.projectID); got != tc.want {
				t.Errorf("buildItemURL(%q, %d, %d) = %q, want %q",
					tc.itemType, tc.itemID, tc.projectID, got, tc.want)
			}
		})
	}
}
