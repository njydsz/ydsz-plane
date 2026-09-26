// Package pages — 项目文档页面领域模型单元测试。
//
// 覆盖范围：
//  1. Page 模型 JSON 序列化往返
//  2. DocumentVersion / DocumentLink / PageTemplate 模型字段
//  3. PageShare 序列化（含 PasswordHash json:"-" 排除）
//  4. PublicSharePageView 序列化
//  5. defaultSortOrder 常量
//  6. CreateTemplateInput / UpdateTemplateInput 零值行为
package pages

import (
	"encoding/json"
	"testing"
	"time"
)

// TestPage_JSON_Roundtrip 验证 Page 模型 JSON 序列化往返。
func TestPage_JSON_Roundtrip(t *testing.T) {
	parentID := int64(10)
	now := time.Date(2025, 5, 1, 12, 0, 0, 0, time.UTC)

	original := Page{
		ID:                  1,
		PublicID:            "abc-123",
		WorkspaceID:         10,
		ProjectID:           100,
		Name:                "API 文档",
		DescriptionJSON:     []byte(`{"type":"doc"}`),
		DescriptionHTML:     "<h1>API</h1>",
		DescriptionStripped: "API",
		ParentID:            &parentID,
		SortOrder:           65535,
		Category:            "api",
		CreatedBy:           7,
		CreatedAt:           now,
		UpdatedAt:           now,
		Version:             1,
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Page marshal: %v", err)
	}

	var decoded Page
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Page unmarshal: %v", err)
	}

	if decoded.ID != original.ID {
		t.Errorf("ID: got %d, want %d", decoded.ID, original.ID)
	}
	if decoded.PublicID != original.PublicID {
		t.Errorf("PublicID: got %q, want %q", decoded.PublicID, original.PublicID)
	}
	if decoded.Name != original.Name {
		t.Errorf("Name: got %q, want %q", decoded.Name, original.Name)
	}
	if decoded.Version != original.Version {
		t.Errorf("Version: got %d, want %d", decoded.Version, original.Version)
	}
	if decoded.SortOrder != original.SortOrder {
		t.Errorf("SortOrder: got %f, want %f", decoded.SortOrder, original.SortOrder)
	}
	if decoded.ParentID == nil || *decoded.ParentID != *original.ParentID {
		t.Errorf("ParentID: got %v, want %v", decoded.ParentID, original.ParentID)
	}
}

// TestPage_JSON_NullableFields 验证 Page omitempty 字段。
func TestPage_JSON_NullableFields(t *testing.T) {
	original := Page{
		ID:          2,
		WorkspaceID: 10,
		ProjectID:   100,
		Name:        "Simple Page",
		CreatedBy:   1,
	}

	// ParentID/DeletedAt 为零值（nil）时应被省略
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if _, exists := raw["parent_id"]; exists {
		t.Errorf("parent_id should be omitted when nil")
	}
	if _, exists := raw["deleted"]; exists {
		t.Errorf("deleted should be omitted when nil")
	}
	if _, exists := raw["description_json"]; exists {
		t.Errorf("description_json should be omitted when empty")
	}
}

// TestPageShare_PasswordHash_Excluded 验证 PageShare.PasswordHash 不序列化到 JSON。
func TestPageShare_PasswordHash_Excluded(t *testing.T) {
	expiresAt := time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC)
	share := PageShare{
		ID:           1,
		PageID:       100,
		WorkspaceID:  10,
		ProjectID:    50,
		Token:        "abc123def456",
		IsActive:     true,
		PasswordHash: "$2a$10$hashed-secret-value",
		ExpiresAt:    &expiresAt,
		CreatedBy:    7,
	}

	data, err := json.Marshal(share)
	if err != nil {
		t.Fatalf("PageShare marshal: %v", err)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if _, exists := raw["password_hash"]; exists {
		t.Errorf("password_hash should be excluded from JSON output (json:\"-\")")
	}

	// 但 token 应该存在
	if _, exists := raw["token"]; !exists {
		t.Errorf("token should exist in JSON output")
	}

	// is_active 应该存在
	if _, exists := raw["is_active"]; !exists {
		t.Errorf("is_active should exist in JSON output")
	}
}

// TestPageShare_JSON_Roundtrip 验证 PageShare 完整序列化往返。
func TestPageShare_JSON_Roundtrip(t *testing.T) {
	share := PageShare{
		ID:          5,
		PageID:      200,
		WorkspaceID: 10,
		ProjectID:   50,
		Token:       "share-token-xyz",
		IsActive:    true,
		CreatedBy:   1,
	}

	data, err := json.Marshal(share)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded PageShare
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if decoded.Token != share.Token {
		t.Errorf("Token: got %q, want %q", decoded.Token, share.Token)
	}
	if decoded.IsActive != share.IsActive {
		t.Errorf("IsActive: got %v, want %v", decoded.IsActive, share.IsActive)
	}
	if decoded.PageID != share.PageID {
		t.Errorf("PageID: got %d, want %d", decoded.PageID, share.PageID)
	}
}

// TestDocumentVersion_JSON_Roundtrip 验证 DocumentVersion 序列化往返。
func TestDocumentVersion_JSON_Roundtrip(t *testing.T) {
	now := time.Date(2025, 6, 15, 9, 0, 0, 0, time.UTC)
	original := DocumentVersion{
		ID:            30,
		PageID:        100,
		VersionNumber: 5,
		ContentMD:     "# v5",
		ContentHTML:   "<h1>v5</h1>",
		CreatedBy:     7,
		CreatedAt:     now,
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded DocumentVersion
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if decoded.VersionNumber != original.VersionNumber {
		t.Errorf("VersionNumber: got %d, want %d", decoded.VersionNumber, original.VersionNumber)
	}
	if decoded.ContentMD != original.ContentMD {
		t.Errorf("ContentMD: got %q, want %q", decoded.ContentMD, original.ContentMD)
	}
}

// TestDocumentLink_JSON_Roundtrip 验证 DocumentLink 序列化往返。
func TestDocumentLink_JSON_Roundtrip(t *testing.T) {
	now := time.Date(2025, 7, 1, 10, 0, 0, 0, time.UTC)
	original := DocumentLink{
		ID:           1,
		PageID:       100,
		LinkableType: "issue",
		LinkableID:   500,
		CreatedBy:    7,
		CreatedAt:    now,
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded DocumentLink
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if decoded.LinkableType != original.LinkableType {
		t.Errorf("LinkableType: got %q, want %q", decoded.LinkableType, original.LinkableType)
	}
	if decoded.LinkableID != original.LinkableID {
		t.Errorf("LinkableID: got %d, want %d", decoded.LinkableID, original.LinkableID)
	}
}

// TestPageTemplate_JSON_Roundtrip 验证 PageTemplate 序列化往返。
func TestPageTemplate_JSON_Roundtrip(t *testing.T) {
	now := time.Date(2025, 8, 1, 0, 0, 0, 0, time.UTC)
	original := PageTemplate{
		ID:          1,
		WorkspaceID: 10,
		ProjectID:   0, // workspace-level template
		Name:        "通用模板",
		ContentHTML: "<p>模板内容</p>",
		CreatedBy:   1,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded PageTemplate
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if decoded.Name != original.Name {
		t.Errorf("Name: got %q, want %q", decoded.Name, original.Name)
	}
	if decoded.ProjectID != 0 {
		t.Errorf("ProjectID: got %d, want 0", decoded.ProjectID)
	}
}

// TestPublicSharePageView_JSON_Roundtrip 验证公开分享视图序列化。
func TestPublicSharePageView_JSON_Roundtrip(t *testing.T) {
	now := time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC)
	original := PublicSharePageView{
		PageID:          100,
		WorkspaceID:     10,
		ProjectID:       50,
		Name:            "公开页面",
		DescriptionHTML: "<p>公开内容</p>",
		DescriptionJSON: []byte(`{"type":"doc"}`),
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded PublicSharePageView
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if decoded.PageID != original.PageID {
		t.Errorf("PageID: got %d, want %d", decoded.PageID, original.PageID)
	}
	if decoded.Name != original.Name {
		t.Errorf("Name: got %q, want %q", decoded.Name, original.Name)
	}
}

// TestDefaultSortOrder_Constant 验证默认排序常量。
func TestDefaultSortOrder_Constant(t *testing.T) {
	if defaultSortOrder != 65535.0 {
		t.Errorf("defaultSortOrder = %f, want 65535.0", defaultSortOrder)
	}
}

// TestCreatePageInput_ZeroValues 验证 CreatePageInput 零值行为。
func TestCreatePageInput_ZeroValues(t *testing.T) {
	input := CreatePageInput{}
	if input.ParentID != nil {
		t.Errorf("zero ParentID = %v, want nil", input.ParentID)
	}
	if input.SortOrder != nil {
		t.Errorf("zero SortOrder = %v, want nil", input.SortOrder)
	}
}
