// Package telemetry — Prometheus 指标纯逻辑单元测试。
package telemetry

import "testing"

// TestClassStatus 校验 HTTP 状态码分类逻辑。
func TestClassStatus(t *testing.T) {
	cases := []struct {
		code int
		want string
	}{
		{100, "1xx"},
		{199, "1xx"},
		{200, "2xx"},
		{204, "2xx"},
		{299, "2xx"},
		{301, "3xx"},
		{304, "3xx"},
		{399, "3xx"},
		{400, "4xx"},
		{404, "4xx"},
		{429, "4xx"},
		{499, "4xx"},
		{500, "5xx"},
		{502, "5xx"},
		{599, "5xx"},
	}
	for _, tc := range cases {
		if got := classStatus(tc.code); got != tc.want {
			t.Errorf("classStatus(%d) = %s, want %s", tc.code, got, tc.want)
		}
	}
}
