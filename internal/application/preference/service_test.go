// Package preference — 视图偏好域单元测试。
//
// 覆盖范围：
//  1. ViewType 枚举值校验
//  2. SavedViewScope 枚举值校验
//  3. ViewPreference 模型 JSON 序列化往返
//  4. SavedView 模型 JSON 序列化往返
//  5. defaultJSON 辅助函数
//  6. joinStrings 辅助函数
package preference

import (
	"encoding/json"
	"testing"
)

// TestViewType_Values 验证 ViewType 枚举值（list / kanban / gantt / calendar）。
func TestViewType_Values(t *testing.T) {
	cases := []struct {
		viewType ViewType
		want     string
	}{
		{ViewKanban, "kanban"},
		{ViewList, "list"},
		{ViewCalendar, "calendar"},
		{ViewGantt, "gantt"},
	}
	for _, tc := range cases {
		if string(tc.viewType) != tc.want {
			t.Errorf("ViewType: got %q, want %q", tc.viewType, tc.want)
		}
	}
}

// TestViewType_IsValid 验证 ViewType 合法值校验。
func TestViewType_IsValid(t *testing.T) {
	validTypes := map[ViewType]bool{
		ViewKanban:   true,
		ViewList:     true,
		ViewCalendar: true,
		ViewGantt:    true,
		"":           false,
		"board":      false,
		"KANBAN":     false, // 大小写敏感
	}
	for vt, wantValid := range validTypes {
		got := isValidViewType(vt)
		if got != wantValid {
			t.Errorf("isValidViewType(%q) = %v, want %v", vt, got, wantValid)
		}
	}
}

// isValidViewType 校验 ViewType 是否为合法枚举值。
func isValidViewType(vt ViewType) bool {
	switch vt {
	case ViewKanban, ViewList, ViewCalendar, ViewGantt:
		return true
	}
	return false
}

// TestSavedViewScope_Values 验证 SavedViewScope 枚举值。
func TestSavedViewScope_Values(t *testing.T) {
	cases := []struct {
		scope SavedViewScope
		want  string
	}{
		{ScopePersonal, "personal"},
		{ScopeTeam, "team"},
		{ScopeDefault, "default"},
	}
	for _, tc := range cases {
		if string(tc.scope) != tc.want {
			t.Errorf("SavedViewScope: got %q, want %q", tc.scope, tc.want)
		}
	}
}

// TestSavedViewScope_IsValid 验证 SavedViewScope 合法值校验。
func TestSavedViewScope_IsValid(t *testing.T) {
	cases := []struct {
		scope SavedViewScope
		want  bool
	}{
		{ScopePersonal, true},
		{ScopeTeam, true},
		{ScopeDefault, true},
		{"public", false},
		{"", false},
	}
	for _, tc := range cases {
		got := isValidSavedViewScope(tc.scope)
		if got != tc.want {
			t.Errorf("isValidSavedViewScope(%q) = %v, want %v", tc.scope, got, tc.want)
		}
	}
}

// isValidSavedViewScope 校验 SavedViewScope 合法性。
func isValidSavedViewScope(s SavedViewScope) bool {
	switch s {
	case ScopePersonal, ScopeTeam, ScopeDefault:
		return true
	}
	return false
}

// TestViewPreference_JSON_Roundtrip 验证 ViewPreference 序列化往返。
func TestViewPreference_JSON_Roundtrip(t *testing.T) {
	original := ViewPreference{
		ID:          1,
		WorkspaceID: 10,
		ProjectID:   100,
		UserID:      7,
		ViewType:    ViewKanban,
		Layout:      "list",
		Columns:     json.RawMessage(`["name","status","priority"]`),
		Filters:     json.RawMessage(`{"priority":"high"}`),
		Sort:        json.RawMessage(`{"field":"created_at","dir":"desc"}`),
		Extra:       json.RawMessage(`{"group_by":"state"}`),
		CreatedAt:   "2025-01-15T10:30:00Z",
		UpdatedAt:   "2025-01-15T11:00:00Z",
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("ViewPreference marshal: %v", err)
	}

	var decoded ViewPreference
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("ViewPreference unmarshal: %v", err)
	}

	if decoded.ID != original.ID {
		t.Errorf("ID: got %d, want %d", decoded.ID, original.ID)
	}
	if decoded.ViewType != original.ViewType {
		t.Errorf("ViewType: got %q, want %q", decoded.ViewType, original.ViewType)
	}
	if decoded.Layout != original.Layout {
		t.Errorf("Layout: got %q, want %q", decoded.Layout, original.Layout)
	}
	if string(decoded.Columns) != string(original.Columns) {
		t.Errorf("Columns: got %s, want %s", decoded.Columns, original.Columns)
	}
}

// TestSavedView_JSON_Roundtrip 验证 SavedView 序列化往返。
func TestSavedView_JSON_Roundtrip(t *testing.T) {
	original := SavedView{
		ID:          5,
		WorkspaceID: 10,
		ProjectID:   100,
		Name:        "我的看板视图",
		Type:        ViewKanban,
		Scope:       ScopePersonal,
		Config:      json.RawMessage(`{"columns":["name","status"]}`),
		OwnerID:     7,
		IsShared:    false,
		CreatedAt:   "2025-03-01T08:00:00Z",
		UpdatedAt:   "2025-03-01T08:00:00Z",
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("SavedView marshal: %v", err)
	}

	var decoded SavedView
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("SavedView unmarshal: %v", err)
	}

	if decoded.Name != original.Name {
		t.Errorf("Name: got %q, want %q", decoded.Name, original.Name)
	}
	if decoded.Type != original.Type {
		t.Errorf("Type: got %q, want %q", decoded.Type, original.Type)
	}
	if decoded.Scope != original.Scope {
		t.Errorf("Scope: got %q, want %q", decoded.Scope, original.Scope)
	}
	if decoded.OwnerID != original.OwnerID {
		t.Errorf("OwnerID: got %d, want %d", decoded.OwnerID, original.OwnerID)
	}
	if decoded.IsShared != original.IsShared {
		t.Errorf("IsShared: got %v, want %v", decoded.IsShared, original.IsShared)
	}
}

// TestDefaultJSON_Empty 验证空 RawMessage 使用 fallback。
func TestDefaultJSON_Empty(t *testing.T) {
	got := defaultJSON(nil, "[]")
	if string(got) != "[]" {
		t.Errorf("defaultJSON(nil, \"[]\") = %s, want []", got)
	}

	got = defaultJSON(json.RawMessage{}, "{}")
	if string(got) != "{}" {
		t.Errorf("defaultJSON(empty, \"{}\") = %s, want {}", got)
	}
}

// TestDefaultJSON_NonEmpty 验证非空 RawMessage 原样返回。
func TestDefaultJSON_NonEmpty(t *testing.T) {
	original := json.RawMessage(`{"key":"value"}`)
	got := defaultJSON(original, "{}")
	if string(got) != `{"key":"value"}` {
		t.Errorf("defaultJSON(non-empty) = %s, want {\"key\":\"value\"}", got)
	}
}

// TestJoinStrings 验证 joinStrings 拼接函数。
func TestJoinStrings(t *testing.T) {
	cases := []struct {
		parts []string
		sep   string
		want  string
	}{
		{nil, ", ", ""},
		{[]string{"a"}, ", ", "a"},
		{[]string{"a", "b", "c"}, ", ", "a, b, c"},
		{[]string{"2 = $2", "3 = $3"}, " AND ", "2 = $2 AND 3 = $3"},
		{[]string{}, "-", ""},
	}
	for _, tc := range cases {
		got := joinStrings(tc.parts, tc.sep)
		if got != tc.want {
			t.Errorf("joinStrings(%v, %q) = %q, want %q", tc.parts, tc.sep, got, tc.want)
		}
	}
}

// TestViewPreference_JSON_EmptyColumns 验证空 JSON 字段序列化为空对象。
func TestViewPreference_JSON_EmptyColumns(t *testing.T) {
	vp := ViewPreference{
		ViewType: ViewList,
		Columns:  nil,
	}

	data, err := json.Marshal(vp)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	// nil RawMessage 序列化为 null
	if val, exists := raw["columns"]; exists {
		if string(val) != "null" {
			t.Errorf("expected columns = null, got %s", val)
		}
	}
}
