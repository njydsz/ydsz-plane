// Package search 搜索纯逻辑测试：toTSQuery 与 buildDocURL。
package search

import (
	"testing"
)

// TestToTSQuery 验证用户输入转 tsquery 的安全性：过滤控制字符、前缀匹配。
func TestToTSQuery(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"foo", "foo:*"},
		{"foo bar", "foo:* & bar:*"},
		{"", ""},
		{"   ", ""},
		{"foo  bar", "foo:* & bar:*"},
		{"FOO-BAR", "FOO-BAR:*"},
		// 控制字符被过滤（非 32-126 ASCII 字符被移除）
		{"foo\x00bar", "foobar:*"},
	}
	for _, tc := range tests {
		if got := toTSQuery(tc.in); got != tc.want {
			t.Errorf("toTSQuery(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

// TestBuildDocURL 验证不同类型搜索结果的前端跳转 URL。
func TestBuildDocURL(t *testing.T) {
	tests := []struct {
		docType string
		docID   int64
		projID  int64
		want    string
	}{
		{"issue", 10, 5, "/projects/5/issues/10"},
		{"sprint", 11, 5, "/projects/5/sprints/11"},
		{"version", 12, 5, "/projects/5/versions/12"},
		{"unknown", 1, 2, ""},
		{"", 1, 2, ""},
	}
	for _, tc := range tests {
		if got := buildDocURL(tc.docType, tc.docID, tc.projID); got != tc.want {
			t.Errorf("buildDocURL(%q, %d, %d) = %q, want %q",
				tc.docType, tc.docID, tc.projID, got, tc.want)
		}
	}
}
