// Package timeutil 提供 API 层时间参数解析工具函数。
//
// 统一时间格式为 RFC3339（2006-01-02T15:04:05Z07:00），
// 也兼容 ISO 8601 / 仅日期格式（2006-01-02），并将空值处理为零值 time.Time。
//
// 设计目标：
//   - 所有 API 时间查询参数通过本包解析，避免各 handler 自行 parse 的错误码不一致。
//   - 不支持 Unix 时间戳格式（出于安全与可读性考虑，API 层一律使用 RFC3339）。
package timeutil

import (
	"fmt"
	"net/http"
	"time"
)

// Layouts 时间解析支持的格式列表，按优先顺序尝试。
var Layouts = []string{
	time.RFC3339,          // 2006-01-02T15:04:05Z07:00
	"2006-01-02T15:04:05", // 无时区的 ISO 8601
	"2006-01-02",          // 纯日期
}

// ParseTime 按支持的格式解析单个时间字符串。
func ParseTime(s string) (time.Time, error) {
	for _, layout := range Layouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unsupported time format: %q (expected RFC3339 or YYYY-MM-DD)", s)
}

// ParseAPITimeRange 从 HTTP 请求的 query 参数中解析起止时间。
//
// 参数：
//   - queryKey: 查询参数名（不含 _after/_before 后缀）。
//   - defaultStart/defaultEnd: 未提供对应参数时的默认值（可为 zero time 表示不限制）。
//
// 例如 ParseAPITimeRange(r, "created",默认30天前, now) 会解析 created_after / created_before。
// 如果解析失败返回 error，handler 应写出 400 VALIDATION_ERROR。
func ParseAPITimeRange(r *http.Request, queryKey string, defaultStart, defaultEnd time.Time) (start, end time.Time, err error) {
	start, end = defaultStart, defaultEnd

	afterKey := queryKey + "_after"
	beforeKey := queryKey + "_before"

	if v := r.URL.Query().Get(afterKey); v != "" {
		t, parseErr := ParseTime(v)
		if parseErr != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("%s: %w", afterKey, parseErr)
		}
		start = t
	}
	if v := r.URL.Query().Get(beforeKey); v != "" {
		t, parseErr := ParseTime(v)
		if parseErr != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("%s: %w", beforeKey, parseErr)
		}
		end = t
	}

	return start, end, nil
}
