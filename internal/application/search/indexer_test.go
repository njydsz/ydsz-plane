// Package search 搜索索引器纯逻辑测试。
//
// 覆盖 indexer.go 中不依赖数据库的纯函数：
//   - eventDocType：领域事件类型 → doc_type 提取（issue.created → issue）
//   - validDocType：doc_type 合法性校验
//   - sourceTable / sourceColumnsFor：源表 → search_documents 的投影 SQL 片段
//
// 背景（P0 缺陷修复回归）：
//   0008 迁移仅为 issues 建立索引触发器，sprints/versions 从不被索引。
//   indexer.go 的三类对象同步（SyncIssue/SyncSprint/SyncVersion）依赖
//   sourceColumnsFor 正确投影各源表字段，此处必须锁定。
package search

import (
	"strings"
	"testing"
)

func TestEventDocType(t *testing.T) {
	tests := []struct {
		eventType string
		want      string
	}{
		{"issue.created", "issue"},
		{"issue.updated", "issue"},
		{"issue.deleted", "issue"},
		{"sprint.started", "sprint"},
		{"version.released", "version"},
		{"no-dot", ""},
		{"", ""},
		{".created", ""}, // 空前缀
	}
	for _, tc := range tests {
		if got := eventDocType(tc.eventType); got != tc.want {
			t.Errorf("eventDocType(%q) = %q, want %q", tc.eventType, got, tc.want)
		}
	}
}

func TestValidDocType(t *testing.T) {
	valid := []string{"issue", "sprint", "version"}
	for _, dt := range valid {
		if _, ok := validDocType(dt); !ok {
			t.Errorf("validDocType(%q) = false, want true", dt)
		}
	}
	invalid := []string{"user", "project", "", "Issue", "issuex"}
	for _, dt := range invalid {
		if _, ok := validDocType(dt); ok {
			t.Errorf("validDocType(%q) = true, want false", dt)
		}
	}
}

func TestSourceTable(t *testing.T) {
	tests := []struct {
		typ  DocType
		want string
	}{
		{DocTypeIssue, "issues"},
		{DocTypeSprint, "sprints"},
		{DocTypeVersion, "versions"},
		{DocType(""), ""},
		{DocType("nope"), ""},
	}
	for _, tc := range tests {
		if got := sourceTable(tc.typ); got != tc.want {
			t.Errorf("sourceTable(%q) = %q, want %q", tc.typ, got, tc.want)
		}
	}
}

// TestSourceColumnsFor 校验各类型投影 SQL 的关键片段。
// 断言不依赖完整字符串（避免脆弱），只锁定语义关键点：
//   - doc_type 字面量（'issue'/'sprint'/'version'）
//   - identifier 来源（issue→sequence_id::text，version→semver，sprint→NULL）
//   - content 来源（sprint→goal，version→description）
func TestSourceColumnsFor(t *testing.T) {
	issue := sourceColumnsFor(DocTypeIssue)
	for _, want := range []string{"'issue'", "src.name", "src.sequence_id::text", "src.description_stripped", "src.type_code"} {
		if !strings.Contains(issue, want) {
			t.Errorf("issue columns missing %q in: %s", want, issue)
		}
	}
	if strings.Contains(issue, "src.goal") {
		t.Error("issue columns must not reference goal")
	}

	sprint := sourceColumnsFor(DocTypeSprint)
	for _, want := range []string{"'sprint'", "src.goal", "NULL", "src.status"} {
		if !strings.Contains(sprint, want) {
			t.Errorf("sprint columns missing %q in: %s", want, sprint)
		}
	}

	version := sourceColumnsFor(DocTypeVersion)
	for _, want := range []string{"'version'", "src.semver", "src.description"} {
		if !strings.Contains(version, want) {
			t.Errorf("version columns missing %q in: %s", want, version)
		}
	}

	if got := sourceColumnsFor(DocType("")); got != "" {
		t.Errorf("sourceColumnsFor(unknown) = %q, want empty", got)
	}
}
