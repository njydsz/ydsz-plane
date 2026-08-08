// Package dlq — DLQ 域单元测试（纯函数部分）。
package dlq

import "testing"

func TestParseAggregateType(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"issue.created", "issue"},
		{"issue.status_changed", "issue"},
		{"sprint.completed", "sprint"},
		{"version.released", "version"},
		{"deployment.recorded", "deployment"},
		{"no_dot", ""},
		{"", ""},
	}
	for _, tc := range cases {
		if got := parseAggregateType(tc.in); got != tc.want {
			t.Fatalf("parseAggregateType(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestTruncateReason(t *testing.T) {
	if got := truncateReason(nil); got != "" {
		t.Fatalf("nil headers should yield empty reason, got %q", got)
	}
	long := make([]byte, 600)
	for i := range long {
		long[i] = 'a'
	}
	got := truncate(string(long), 500)
	if len(got) != 503 { // 500 + "..."
		t.Fatalf("truncate length = %d, want 503", len(got))
	}
}
