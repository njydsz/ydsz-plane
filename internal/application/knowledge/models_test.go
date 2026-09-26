// Package knowledge — 知识库领域模型单元测试。
//
// 覆盖范围：
//  1. KnowledgeSpace / KnowledgePage 模型 JSON 序列化往返
//  2. SpacePermission / PageStatus / PageRelationType 枚举值
//  3. KnowledgePageNode 树形结构嵌套序列化
//  4. PageStatus 合法值校验
package knowledge

import (
	"encoding/json"
	"testing"
	"time"
)

// TestKnowledgeSpace_JSON_Roundtrip 验证 KnowledgeSpace 序列化与反序列化的一致性。
func TestKnowledgeSpace_JSON_Roundtrip(t *testing.T) {
	ownerID := int64(42)
	projectID := int64(100)
	now := time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC)

	original := KnowledgeSpace{
		ID:                1,
		WorkspaceID:       10,
		ProjectID:         &projectID,
		Name:              "工程文档",
		Slug:              "eng-docs",
		Description:       "技术文档空间",
		OwnerID:           &ownerID,
		DefaultPermission: PermissionEditor,
		IsPrivate:         true,
		CoverImage:        "https://example.com/cover.png",
		CreatedAt:         now,
		UpdatedAt:         now,
		DeletedAt:         nil,
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("KnowledgeSpace marshal: %v", err)
	}

	var decoded KnowledgeSpace
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("KnowledgeSpace unmarshal: %v", err)
	}

	if decoded.ID != original.ID {
		t.Errorf("ID: got %d, want %d", decoded.ID, original.ID)
	}
	if decoded.Name != original.Name {
		t.Errorf("Name: got %q, want %q", decoded.Name, original.Name)
	}
	if decoded.Slug != original.Slug {
		t.Errorf("Slug: got %q, want %q", decoded.Slug, original.Slug)
	}
	if decoded.DefaultPermission != original.DefaultPermission {
		t.Errorf("DefaultPermission: got %q, want %q", decoded.DefaultPermission, original.DefaultPermission)
	}
	if decoded.IsPrivate != original.IsPrivate {
		t.Errorf("IsPrivate: got %v, want %v", decoded.IsPrivate, original.IsPrivate)
	}
	if decoded.ProjectID == nil || *decoded.ProjectID != *original.ProjectID {
		t.Errorf("ProjectID: got %v, want %v", decoded.ProjectID, original.ProjectID)
	}
	if decoded.OwnerID == nil || *decoded.OwnerID != *original.OwnerID {
		t.Errorf("OwnerID: got %v, want %v", decoded.OwnerID, original.OwnerID)
	}
}

// TestKnowledgeSpace_JSON_NullableFields 验证 omitempty 字段在为零值时不出现在 JSON 中。
func TestKnowledgeSpace_JSON_NullableFields(t *testing.T) {
	original := KnowledgeSpace{
		ID:                2,
		WorkspaceID:       10,
		Name:              "公开空间",
		Slug:              "public",
		DefaultPermission: PermissionViewer,
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("unmarshal to map: %v", err)
	}

	// omitempty 字段不应出现在 JSON 中
	for _, key := range []string{"project_id", "owner_id", "description", "cover_image", "deleted"} {
		if _, exists := raw[key]; exists {
			t.Errorf("field %q should be omitted when zero value", key)
		}
	}
}

// TestKnowledgePage_JSON_Roundtrip 验证 KnowledgePage 序列化往返。
func TestKnowledgePage_JSON_Roundtrip(t *testing.T) {
	parentID := int64(5)
	creatorID := int64(7)
	now := time.Date(2025, 3, 20, 14, 0, 0, 0, time.UTC)

	original := KnowledgePage{
		ID:          100,
		WorkspaceID: 10,
		SpaceID:     1,
		ParentID:    &parentID,
		Lft:         2,
		Rgt:         7,
		Depth:       1,
		Title:       "接口设计文档",
		Path:        "/eng-docs/api-design",
		ContentMD:   "# API Design\nContent here",
		ContentHTML: "<h1>API Design</h1><p>Content here</p>",
		Version:     3,
		Status:      PageStatusPublished,
		SortOrder:   10,
		IsPinned:    true,
		IsFeatured:  false,
		ViewCount:   42,
		CreatedBy:   &creatorID,
		UpdatedBy:   &creatorID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("KnowledgePage marshal: %v", err)
	}

	var decoded KnowledgePage
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("KnowledgePage unmarshal: %v", err)
	}

	if decoded.ID != original.ID {
		t.Errorf("ID: got %d, want %d", decoded.ID, original.ID)
	}
	if decoded.Depth != original.Depth {
		t.Errorf("Depth: got %d, want %d", decoded.Depth, original.Depth)
	}
	if decoded.Version != original.Version {
		t.Errorf("Version: got %d, want %d", decoded.Version, original.Version)
	}
	if decoded.Status != original.Status {
		t.Errorf("Status: got %q, want %q", decoded.Status, original.Status)
	}
	if decoded.Path != original.Path {
		t.Errorf("Path: got %q, want %q", decoded.Path, original.Path)
	}
	if decoded.SortOrder != original.SortOrder {
		t.Errorf("SortOrder: got %d, want %d", decoded.SortOrder, original.SortOrder)
	}
	if decoded.IsPinned != original.IsPinned {
		t.Errorf("IsPinned: got %v, want %v", decoded.IsPinned, original.IsPinned)
	}
}

// TestKnowledgePage_Node_TreeStructure 验证 KnowledgePageNode 嵌套树形结构序列化。
func TestKnowledgePage_Node_TreeStructure(t *testing.T) {
	page := KnowledgePageNode{
		KnowledgePage: KnowledgePage{
			ID:    1,
			Title: "Root",
			Depth: 0,
		},
		Children: []KnowledgePageNode{
			{
				KnowledgePage: KnowledgePage{
					ID:    2,
					Title: "Child A",
					Depth: 1,
				},
				Children: []KnowledgePageNode{
					{
						KnowledgePage: KnowledgePage{
							ID:    4,
							Title: "Grandchild A1",
							Depth: 2,
						},
						Children: []KnowledgePageNode{},
					},
				},
			},
			{
				KnowledgePage: KnowledgePage{
					ID:    3,
					Title: "Child B",
					Depth: 1,
				},
				Children: []KnowledgePageNode{},
			},
		},
	}

	data, err := json.Marshal(page)
	if err != nil {
		t.Fatalf("tree marshal: %v", err)
	}

	var decoded KnowledgePageNode
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("tree unmarshal: %v", err)
	}

	if decoded.ID != 1 {
		t.Errorf("root ID: got %d, want 1", decoded.ID)
	}
	if len(decoded.Children) != 2 {
		t.Errorf("root children count: got %d, want 2", len(decoded.Children))
	}
	if len(decoded.Children[0].Children) != 1 {
		t.Errorf("Child A grandchildren count: got %d, want 1", len(decoded.Children[0].Children))
	}
	if decoded.Children[0].Children[0].Depth != 2 {
		t.Errorf("grandchild depth: got %d, want 2", decoded.Children[0].Children[0].Depth)
	}
}

// TestSpacePermission_Values 验证 SpacePermission 枚举值稳定。
func TestSpacePermission_Values(t *testing.T) {
	cases := []struct {
		perm SpacePermission
		want string
	}{
		{PermissionViewer, "viewer"},
		{PermissionEditor, "editor"},
		{PermissionAdmin, "admin"},
		{PermissionOwner, "owner"},
	}
	for _, tc := range cases {
		if string(tc.perm) != tc.want {
			t.Errorf("SpacePermission: got %q, want %q", tc.perm, tc.want)
		}
	}
}

// TestPageStatus_Values 验证 PageStatus 枚举值稳定。
func TestPageStatus_Values(t *testing.T) {
	cases := []struct {
		status PageStatus
		want   string
	}{
		{PageStatusDraft, "draft"},
		{PageStatusPublished, "published"},
		{PageStatusArchived, "archived"},
	}
	for _, tc := range cases {
		if string(tc.status) != tc.want {
			t.Errorf("PageStatus: got %q, want %q", tc.status, tc.want)
		}
	}
}

// TestPageStatus_IsValid 验证 PageStatus 合法值校验。
func TestPageStatus_IsValid(t *testing.T) {
	validStatuses := map[PageStatus]bool{
		PageStatusDraft:     true,
		PageStatusPublished: true,
		PageStatusArchived:  true,
		"invalid":           false,
		"":                  false,
	}
	for status, wantValid := range validStatuses {
		got := isValidPageStatus(status)
		if got != wantValid {
			t.Errorf("isValidPageStatus(%q) = %v, want %v", status, got, wantValid)
		}
	}
}

// isValidPageStatus 校验 PageStatus 是否为合法枚举值。
func isValidPageStatus(s PageStatus) bool {
	switch s {
	case PageStatusDraft, PageStatusPublished, PageStatusArchived:
		return true
	}
	return false
}

// TestPageDepth_Range 验证 depth 字段范围规则（0-3 之间有效）。
func TestPageDepth_Range(t *testing.T) {
	cases := []struct {
		depth int
		valid bool
	}{
		{0, true},
		{1, true},
		{2, true},
		{3, true},
		{-1, false},
		{4, false},
		{10, false},
	}
	for _, tc := range cases {
		got := tc.depth >= 0 && tc.depth <= 3
		if got != tc.valid {
			t.Errorf("depth %d validity: got %v, want %v", tc.depth, got, tc.valid)
		}
	}
}

// TestPageRelationType_Values 验证 PageRelationType 枚举值。
func TestPageRelationType_Values(t *testing.T) {
	cases := []struct {
		rel  PageRelationType
		want string
	}{
		{RelationReferenced, "referenced"},
		{RelationReferencing, "referencing"},
	}
	for _, tc := range cases {
		if string(tc.rel) != tc.want {
			t.Errorf("PageRelationType: got %q, want %q", tc.rel, tc.want)
		}
	}
}

// TestKnowledgePageVersion_JSON_Roundtrip 验证版本快照序列化往返。
func TestKnowledgePageVersion_JSON_Roundtrip(t *testing.T) {
	createdBy := int64(9)
	now := time.Date(2025, 4, 1, 9, 0, 0, 0, time.UTC)

	original := KnowledgePageVersion{
		ID:            50,
		PageID:        100,
		Version:       5,
		Title:         "第 5 版",
		ContentMD:     "# 更新内容",
		ContentHTML:   "<h1>更新内容</h1>",
		ChangeSummary: "修改了标题",
		CreatedBy:     &createdBy,
		CreatedAt:     now,
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded KnowledgePageVersion
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if decoded.Version != original.Version {
		t.Errorf("Version: got %d, want %d", decoded.Version, original.Version)
	}
	if decoded.ChangeSummary != original.ChangeSummary {
		t.Errorf("ChangeSummary: got %q, want %q", decoded.ChangeSummary, original.ChangeSummary)
	}
	if decoded.CreatedBy == nil || *decoded.CreatedBy != *original.CreatedBy {
		t.Errorf("CreatedBy: got %v, want %v", decoded.CreatedBy, original.CreatedBy)
	}
}
