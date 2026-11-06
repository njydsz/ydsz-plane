// Package projectstate — 领域模型单元测试。
//
// 覆盖：优先级合法性校验。
package projectstate

import "testing"

func TestIsValidPriority(t *testing.T) {
	t.Parallel()

	valid := []string{"urgent", "high", "medium", "low", "none"}
	for _, p := range valid {
		if !IsValidPriority(p) {
			t.Errorf("IsValidPriority(%q) = false, want true", p)
		}
	}

	invalid := []string{"", "URGENT", "critical", "unknown", " high "}
	for _, p := range invalid {
		if IsValidPriority(p) {
			t.Errorf("IsValidPriority(%q) = true, want false", p)
		}
	}
}
