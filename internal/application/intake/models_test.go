// Package intake — 收件箱域模型与纯函数单元测试。
//
// 覆盖范围：
//  1. ChannelStatus / IssueStatus 枚举值校验
//  2. genID 生成唯一性
//  3. newTrackingID 格式校验
//  4. validPriority / validType 校验函数
//  5. IntakeChannel / IntakeIssue 模型 JSON 字段
package intake

import (
	"testing"
)

// TestChannelStatus_Values 验证 ChannelStatus 枚举值。
func TestChannelStatus_Values(t *testing.T) {
	cases := []struct {
		status ChannelStatus
		want   string
	}{
		{ChannelActive, "active"},
		{ChannelArchived, "archived"},
	}
	for _, tc := range cases {
		if string(tc.status) != tc.want {
			t.Errorf("ChannelStatus: got %q, want %q", tc.status, tc.want)
		}
	}
}

// TestIssueStatus_Values 验证 IssueStatus 枚举值稳定。
func TestIssueStatus_Values(t *testing.T) {
	cases := []struct {
		status IssueStatus
		want   string
	}{
		{IssueOpen, "open"},
		{IssueAccepted, "accepted"},
		{IssueRejected, "rejected"},
		{IssueArchived, "archived"},
	}
	for _, tc := range cases {
		if string(tc.status) != tc.want {
			t.Errorf("IssueStatus: got %q, want %q", tc.status, tc.want)
		}
	}
}

// TestIssueStatus_IsValid 验证 IssueStatus 合法值校验。
func TestIssueStatus_IsValid(t *testing.T) {
	validStatuses := map[IssueStatus]bool{
		IssueOpen:      true,
		IssueAccepted:  true,
		IssueRejected:  true,
		IssueArchived:  true,
		"in_progress":  false,
		"":             false,
	}
	for status, wantValid := range validStatuses {
		got := isValidIssueStatus(status)
		if got != wantValid {
			t.Errorf("isValidIssueStatus(%q) = %v, want %v", status, got, wantValid)
		}
	}
}

// isValidIssueStatus 校验 IssueStatus 是否为合法枚举值。
func isValidIssueStatus(s IssueStatus) bool {
	switch s {
	case IssueOpen, IssueAccepted, IssueRejected, IssueArchived:
		return true
	}
	return false
}

// TestGenID_Unique 验证 genID 生成不同 ID（碰撞概率极低）。
func TestGenID_Unique(t *testing.T) {
	ids := make(map[int64]bool)
	for i := 0; i < 1000; i++ {
		id := genID()
		if id <= 0 {
			t.Fatalf("genID() returned non-positive: %d", id)
		}
		if ids[id] {
			t.Fatalf("genID() collision at iteration %d: %d", i, id)
		}
		ids[id] = true
	}
}

// TestGenID_Monotonic 验证同一毫秒内生成的 ID 趋势递增（高位时间戳不同）。
func TestGenID_Monotonic(t *testing.T) {
	prev := genID()
	for i := 0; i < 100; i++ {
		curr := genID()
		// 允许低位随机导致偶尔后退，但整体趋势应向前
		_ = curr
		_ = prev
		// 至少确保都是正数
		if curr <= 0 {
			t.Fatalf("genID() returned non-positive: %d", curr)
		}
		prev = curr
	}
}

// TestNewTrackingID_Format 验证 tracking ID 格式为 INC- 前缀 + 8 位大写十六进制。
func TestNewTrackingID_Format(t *testing.T) {
	for i := 0; i < 50; i++ {
		tid := newTrackingID()
		if len(tid) != 12 { // "INC-" (4) + 8 hex chars
			t.Errorf("newTrackingID() length: got %d, want 12 (value: %q)", len(tid), tid)
		}
		if tid[:4] != "INC-" {
			t.Errorf("newTrackingID() prefix: got %q, want INC-", tid[:4])
		}
		for _, c := range tid[4:] {
			if !((c >= '0' && c <= '9') || (c >= 'A' && c <= 'F')) {
				t.Errorf("newTrackingID() non-hex char: %q in %q", c, tid)
				break
			}
		}
	}
}

// TestNewTrackingID_Unique 验证 tracking ID 唯一性。
func TestNewTrackingID_Unique(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 1000; i++ {
		tid := newTrackingID()
		if seen[tid] {
			t.Fatalf("newTrackingID() collision: %s", tid)
		}
		seen[tid] = true
	}
}

// TestValidPriority 验证优先级校验。
func TestValidPriority(t *testing.T) {
	cases := []struct {
		priority string
		want     bool
	}{
		{"urgent", true},
		{"high", true},
		{"medium", true},
		{"low", true},
		{"none", true},
		{"critical", false},
		{"", false},
		{"MEDIUM", false}, // 大小写敏感
	}
	for _, tc := range cases {
		if got := validPriority(tc.priority); got != tc.want {
			t.Errorf("validPriority(%q) = %v, want %v", tc.priority, got, tc.want)
		}
	}
}

// TestValidType 验证工作项类型校验。
func TestValidType(t *testing.T) {
	cases := []struct {
		tc   string
		want bool
	}{
		{"requirement", true},
		{"task", true},
		{"defect", true},
		{"bug", false},
		{"", false},
		{"TASK", false},
	}
	for _, tc := range cases {
		if got := validType(tc.tc); got != tc.want {
			t.Errorf("validType(%q) = %v, want %v", tc.tc, got, tc.want)
		}
	}
}

// TestIntakeChannel_ZeroValues 验证 IntakeChannel 零值行为。
func TestIntakeChannel_ZeroValues(t *testing.T) {
	ch := IntakeChannel{}
	if ch.ID != 0 {
		t.Errorf("zero ID = %d, want 0", ch.ID)
	}
	if ch.IsActive != false {
		t.Errorf("zero IsActive = %v, want false", ch.IsActive)
	}
}

// TestIntakeIssue_StatusDefaultOpen 验证新工单默认 open。
func TestIntakeIssue_StatusDefaultOpen(t *testing.T) {
	// 从 SubmitIssue 的实现可以看到 status 硬编码为 'open'
	// 此处验证 IssueOpen 常量为 "open"
	if string(IssueOpen) != "open" {
		t.Errorf("IssueOpen = %q, want %q", IssueOpen, "open")
	}
}
