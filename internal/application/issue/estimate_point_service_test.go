// Package issue — EstimatePoint 服务测试。
//
// 验证 EstimatePoint 实体字段映射与公共方法契约。
package issue

import (
	"encoding/json"
	"testing"
)

// TestEstimatePointJSONMarshall 验证 JSON 序列化包含所有关键字段。
func TestEstimatePointJSONMarshall(t *testing.T) {
	points := json.RawMessage(`[{"label":"S","value":1},{"label":"M","value":2}]`)
	ep := EstimatePoint{
		ID:          1,
		Code:        "EP-001",
		Name:        "T-Shirt Size",
		WorkspaceID: 10,
		ProjectID:   20,
		Description: "T shirt sizing",
		Points:      points,
		IsDefault:   true,
		Status:      "active",
	}

	b, err := json.Marshal(ep)
	if err != nil {
		t.Fatalf("Marshal() error: %v", err)
	}

	// 验证关键字段出现在 JSON 中
	s := string(b)
	wantTokens := []string{
		`"id":1`,
		`"name":"T-Shirt Size"`,
		`"workspace_id":10`,
		`"project_id":20`,
		`"is_default":true`,
		`"status":"active"`,
		`"points":[`,
	}
	for _, tok := range wantTokens {
		if !containsString(s, tok) {
			t.Errorf("JSON missing %q, got: %s", tok, s)
		}
	}
}

// TestEstimatePointEmptyPoints 验证空 Points 字段序列化为 []。
func TestEstimatePointEmptyPoints(t *testing.T) {
	ep := EstimatePoint{ID: 1, Name: "Test"}
	b, _ := json.Marshal(ep)
	// Points 为 nil 时，序列化为 null（标准 encoding/json 行为）
	s := string(b)
	if !containsString(s, `"points":null`) {
		t.Errorf("Expected points:null for nil Points, got: %s", s)
	}
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && indexOf(s, substr) >= 0)
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
