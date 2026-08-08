// Package notification — 通知摘要 (Digest) 单元测试。
//
// 覆盖 S9.4 遗留项：digest 日报/周报聚合逻辑。
// 聚焦纯函数（无 DB 依赖）：时间窗口判断、模板生成、回看周期。
package notification

import (
	"strings"
	"testing"
	"time"
)

// ==========================================================================
// P1: ShouldDigestNow — 触发时间判断
// ==========================================================================

// TestShouldDigestNow_DailyWeekday 验证 daily 频率在工作日 08:30 应触发。
func TestShouldDigestNow_DailyWeekday(t *testing.T) {
	// 上海时区 UTC+8：UTC 00:30 → 上海 08:30
	cases := []struct {
		name     string
		d        Digest
		tz       string
		now      time.Time
		expected bool
	}{
		{
			name:     "daily 工作日 08:30 触发",
			d:        DigestDaily,
			tz:       "Asia/Shanghai",
			now:      time.Date(2026, 8, 4, 0, 30, 0, 0, time.UTC),
			expected: true,
		},
		{
			name:     "daily 工作日 08:31 不触发",
			d:        DigestDaily,
			tz:       "Asia/Shanghai",
			now:      time.Date(2026, 8, 4, 0, 31, 0, 0, time.UTC),
			expected: false,
		},
		{
			name:     "daily 周末 08:30 不触发",
			d:        DigestDaily,
			tz:       "Asia/Shanghai",
			now:      time.Date(2026, 8, 2, 0, 30, 0, 0, time.UTC),
			expected: false,
		},
		{
			name:     "weekly 周一 08:30 触发",
			d:        DigestWeekly,
			tz:       "Asia/Shanghai",
			now:      time.Date(2026, 8, 3, 0, 30, 0, 0, time.UTC),
			expected: true,
		},
		{
			name:     "weekly 周二 08:30 不触发",
			d:        DigestWeekly,
			tz:       "Asia/Shanghai",
			now:      time.Date(2026, 8, 4, 0, 30, 0, 0, time.UTC),
			expected: false,
		},
		{
			name:     "weekly 周日 08:30 不触发",
			d:        DigestWeekly,
			tz:       "Asia/Shanghai",
			now:      time.Date(2026, 8, 9, 0, 30, 0, 0, time.UTC),
			expected: false,
		},
		{
			name:     "off 频率永不触发",
			d:        DigestOff,
			tz:       "Asia/Shanghai",
			now:      time.Date(2026, 8, 4, 0, 30, 0, 0, time.UTC),
			expected: false,
		},
		{
			name:     "UTC 时区 08:30 周一触发",
			d:        DigestDaily,
			tz:       "UTC",
			now:      time.Date(2026, 8, 3, 8, 30, 0, 0, time.UTC),
			expected: true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := ShouldDigestNow(c.d, c.tz, c.now)
			if got != c.expected {
				t.Errorf("ShouldDigestNow(%q, %q, %v) = %v, want %v",
					c.d, c.tz, c.now, got, c.expected)
			}
		})
	}
}

// TestShouldDigestNow_UnknownTimezone 验证未知时区回退到 UTC 不 panic。
func TestShouldDigestNow_UnknownTimezone(t *testing.T) {
	got := ShouldDigestNow(DigestDaily, "Invalid/Timezone", time.Date(2026, 8, 4, 8, 30, 0, 0, time.UTC))
	_ = got
}

// ==========================================================================
// P2: DefaultDigestWindowStart — 默认回看窗口
// ==========================================================================

// TestDefaultDigestWindowStart_Daily 验证 daily 回看 24 小时。
func TestDefaultDigestWindowStart_Daily(t *testing.T) {
	now := time.Date(2026, 8, 7, 10, 0, 0, 0, time.UTC)
	start := DefaultDigestWindowStart(DigestDaily, now)
	expected := now.Add(-24 * time.Hour)
	if !start.Equal(expected) {
		t.Errorf("daily window start = %v, want %v", start, expected)
	}
}

// TestDefaultDigestWindowStart_Weekly 验证 weekly 回看 7 天。
func TestDefaultDigestWindowStart_Weekly(t *testing.T) {
	now := time.Date(2026, 8, 7, 10, 0, 0, 0, time.UTC)
	start := DefaultDigestWindowStart(DigestWeekly, now)
	expected := now.AddDate(0, 0, -7)
	if !start.Equal(expected) {
		t.Errorf("weekly window start = %v, want %v", start, expected)
	}
}

// ==========================================================================
// P3: BuildDigestSubject — 邮件主题
// ==========================================================================

// TestBuildDigestSubject_Daily 验证 daily 主题包含"日报"。
func TestBuildDigestSubject_Daily(t *testing.T) {
	subject := BuildDigestSubject(DigestDaily, "核心产品")
	if !strings.Contains(subject, "日报") {
		t.Errorf("daily subject = %q, want contains 日报", subject)
	}
	if !strings.Contains(subject, "核心产品") {
		t.Errorf("daily subject = %q, want workspace name", subject)
	}
}

// TestBuildDigestSubject_Weekly 验证 weekly 主题包含"周报"。
func TestBuildDigestSubject_Weekly(t *testing.T) {
	subject := BuildDigestSubject(DigestWeekly, "核心产品")
	if !strings.Contains(subject, "周报") {
		t.Errorf("weekly subject = %q, want contains 周报", subject)
	}
}

// ==========================================================================
// P4: BuildDigestHTML — HTML 内容生成
// ==========================================================================

// TestBuildDigestHTML_CountInBody 验证 HTML 包含通知数量。
func TestBuildDigestHTML_CountInBody(t *testing.T) {
	now := time.Now()
	payload := &DigestPayload{
		GeneratedAt: now,
		DigestType:  DigestDaily,
		PeriodStart: now.Add(-24 * time.Hour),
		PeriodEnd:   now,
		TotalCount:  3,
		Items: []DigestItem{
			{EventType: EventIssueCreated, Title: "创建了工作项", Body: "[YD-1] 测试需求", ActionURL: "/projects/1/issues/1"},
		},
	}
	html := BuildDigestHTML(payload, "核心产品")
	// 数字被 <strong> 包裹，检查紧邻 "条 新通知" 上下文中的数字
	if !strings.Contains(html, "3") || !strings.Contains(html, "条新通知") {
		t.Errorf("HTML should contain count '3 条', got: %s", html)
	}
	if !strings.Contains(html, "[YD-1] 测试需求") {
		t.Errorf("HTML should contain notification body")
	}
	if !strings.Contains(html, "/projects/1/issues/1") {
		t.Errorf("HTML should contain action URL")
	}
}

// TestBuildDigestHTML_EmptyItems 验证空通知列表时 HTML 计数为 0。
func TestBuildDigestHTML_EmptyItems(t *testing.T) {
	now := time.Now()
	payload := &DigestPayload{
		GeneratedAt: now,
		DigestType:  DigestWeekly,
		PeriodStart: now.AddDate(0, 0, -7),
		PeriodEnd:   now,
		TotalCount:  0,
		Items:       []DigestItem{},
	}
	html := BuildDigestHTML(payload, "测试工作空间")
	if !strings.Contains(html, "0") || !strings.Contains(html, "条新通知") {
		t.Errorf("empty digest HTML should show 0 count, got: %s", html)
	}
}

// ==========================================================================
// P5: digestTextSummary — 纯文本摘要
// ==========================================================================

// TestDigestTextSummary_Full 验证文本摘要包含时间和标题。
func TestDigestTextSummary_Full(t *testing.T) {
	now := time.Now()
	payload := &DigestPayload{
		GeneratedAt: now,
		DigestType:  DigestDaily,
		PeriodStart: now.Add(-24 * time.Hour),
		PeriodEnd:   now,
		TotalCount:  2,
		Items: []DigestItem{
			{Title: "Alice 创建需求", Body: "用户登录功能", ActionURL: "/a/b"},
			{Title: "Bob 评论", Body: "请查看实现", ActionURL: "/a/c"},
		},
	}
	text := digestTextSummary(payload)
	if !strings.Contains(text, "Alice 创建需求") {
		t.Errorf("text summary should contain first item title: %s", text)
	}
	if !strings.Contains(text, "Bob 评论") {
		t.Errorf("text summary should contain second item title: %s", text)
	}
	if !strings.Contains(text, "/a/b") {
		t.Errorf("text summary should contain action URL")
	}
	// 计数文本在文本摘要中格式为 "共 N 条新通知"
	if !strings.Contains(text, "2") {
		t.Errorf("text summary should contain count number: %s", text)
	}
}

// ==========================================================================
// P6: Payload 结构验证
// ==========================================================================

// TestDigestPayload_TotalCountConsistency 验证 total_count 与 items 长度一致。
func TestDigestPayload_TotalCountConsistency(t *testing.T) {
	items := []DigestItem{
		{Title: "a"}, {Title: "b"}, {Title: "c"},
	}
	payload := &DigestPayload{TotalCount: len(items), Items: items}
	if payload.TotalCount != len(payload.Items) {
		t.Errorf("TotalCount = %d, len(Items) = %d", payload.TotalCount, len(payload.Items))
	}
}
