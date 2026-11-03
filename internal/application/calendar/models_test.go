// Package calendar — 领域模型单元测试。
//
// 覆盖 EventType 校验、EventStatus 终态判定、genID 唯一性与单调性。
package calendar

import (
	"testing"
	"time"
)

func TestEventType_IsValid(t *testing.T) {
	t.Parallel()

	valid := []EventType{EventMeeting, EventMilestone, EventRelease, EventDeadline, EventOther}
	for _, e := range valid {
		if !e.IsValid() {
			t.Errorf("EventType(%q) should be valid", e)
		}
	}

	invalid := []EventType{"", "unknown", "MEETING", "event"}
	for _, e := range invalid {
		if e.IsValid() {
			t.Errorf("EventType(%q) should be invalid", e)
		}
	}
}

func TestEventStatus_IsTerminal(t *testing.T) {
	t.Parallel()

	if !StatusArchived.IsTerminal() {
		t.Error("StatusArchived should be terminal")
	}
	if StatusActive.IsTerminal() {
		t.Error("StatusActive should not be terminal")
	}
}

func TestGenID_Unique(t *testing.T) {
	t.Parallel()

	seen := make(map[int64]struct{}, 1000)
	for i := 0; i < 1000; i++ {
		id := genID()
		if id <= 0 {
			t.Fatalf("genID() = %d, want positive", id)
		}
		if _, dup := seen[id]; dup {
			t.Fatalf("genID() collision at iteration %d: %d", i, id)
		}
		seen[id] = struct{}{}
	}
}

func TestGenID_MonotonicAcrossMillis(t *testing.T) {
	t.Parallel()

	// 同一毫秒内低 22 位随机，不保证单调；跨毫秒时高位时间戳递增，
	// 因此时间戳部分必然非递减。
	lastMs := int64(-1)
	for i := 0; i < 50; i++ {
		id := genID()
		ms := id >> 22
		if ms < lastMs {
			t.Fatalf("genID() timestamp went backwards: %d < %d", ms, lastMs)
		}
		lastMs = ms
	}
}

func TestCreateEventInput_Validation(t *testing.T) {
	t.Parallel()

	now := time.Now()
	cases := []struct {
		name    string
		in      CreateEventInput
		wantErr bool
	}{
		{
			name: "valid",
			in: CreateEventInput{
				WorkspaceID: 1,
				Title:       "评审会",
				StartTime:   now,
				EndTime:     now.Add(time.Hour),
				EventType:   EventMeeting,
			},
			wantErr: false,
		},
		{
			name: "empty title",
			in: CreateEventInput{
				WorkspaceID: 1,
				Title:       "",
				StartTime:   now,
				EndTime:     now.Add(time.Hour),
				EventType:   EventMeeting,
			},
			wantErr: true,
		},
		{
			name: "zero workspace",
			in: CreateEventInput{
				WorkspaceID: 0,
				Title:       "评审会",
				StartTime:   now,
				EndTime:     now.Add(time.Hour),
				EventType:   EventMeeting,
			},
			wantErr: true,
		},
		{
			name: "end before start",
			in: CreateEventInput{
				WorkspaceID: 1,
				Title:       "评审会",
				StartTime:   now,
				EndTime:     now.Add(-time.Hour),
				EventType:   EventMeeting,
			},
			wantErr: true,
		},
		{
			name: "invalid event type",
			in: CreateEventInput{
				WorkspaceID: 1,
				Title:       "评审会",
				StartTime:   now,
				EndTime:     now.Add(time.Hour),
				EventType:   EventType("unknown"),
			},
			wantErr: true,
		},
		{
			name: "missing start time",
			in: CreateEventInput{
				WorkspaceID: 1,
				Title:       "评审会",
				EndTime:     now.Add(time.Hour),
				EventType:   EventMeeting,
			},
			wantErr: true,
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			err := validateCreateInput(c.in)
			if (err != nil) != c.wantErr {
				t.Errorf("validateCreateInput() error = %v, wantErr %v", err, c.wantErr)
			}
		})
	}
}
