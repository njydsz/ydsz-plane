// Package automation — scheduled 触发器调度器单元测试。
package automation

import (
	"testing"
	"time"
)

func TestCronMatches(t *testing.T) {
	// 本地时区基准时刻：2026-08-08 09:00（周六）——覆盖模板 "0 9 * * *"
	base := time.Date(2026, 8, 8, 9, 0, 0, 0, time.Local)

	cases := []struct {
		name  string
		cron  string
		at    time.Time
		match bool
	}{
		{"daily 9am exact", "0 9 * * *", base, true},
		{"daily 9am minute off", "0 9 * * *", base.Add(time.Minute), false},
		{"daily 9am hour off", "0 9 * * *", base.Add(time.Hour), false},
		{"daily 2am", "0 2 * * *", time.Date(2026, 8, 8, 2, 0, 0, 0, time.Local), true},
		{"star minute", "* * * * *", time.Date(2026, 8, 8, 13, 42, 0, 0, time.Local), true},
		{"every 15 min", "*/15 * * * *", time.Date(2026, 8, 8, 10, 45, 0, 0, time.Local), true},
		{"every 15 min off", "*/15 * * * *", time.Date(2026, 8, 8, 10, 44, 0, 0, time.Local), false},
		{"range hour", "0 9-18 * * *", base, true},
		{"range hour off", "0 9-18 * * *", base.Add(10 * time.Hour), false},
		{"or list minute", "0,30 9 * * *", base, true},
		{"or list minute off", "0,30 9 * * *", base.Add(5 * time.Minute), false},
		{"weekday match (Sat)", "0 9 * * 6", base, true},
		{"weekday no match (Sun)", "0 9 * * 0", base, false},
		{"step range", "0 9 */2 * *", base, true}, // 8 日为偶数日
		{"bad field count", "0 9 * *", base, false},
		{"non numeric", "abc * * * *", base, false},
		{"step bad", "*/0 * * * *", base, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := CronMatches(tc.cron, tc.at); got != tc.match {
				t.Fatalf("CronMatches(%q, %v) = %v, want %v", tc.cron, tc.at, got, tc.match)
			}
		})
	}
}

func TestNumberFromAny(t *testing.T) {
	cases := []struct {
		in   any
		want float64
		ok   bool
	}{
		{float64(24), 24, true},
		{24, 24, true},
		{int64(24), 24, true},
		{"24", 24, true},
		{"abc", 0, false},
		{nil, 0, false},
	}
	for _, tc := range cases {
		got, ok := numberFromAny(tc.in)
		if ok != tc.ok || (ok && got != tc.want) {
			t.Fatalf("numberFromAny(%v) = (%v,%v), want (%v,%v)", tc.in, got, ok, tc.want, tc.ok)
		}
	}
}
