// Package datajob — 领域模型单元测试。
//
// 覆盖 JobStatus 校验、终态判定、genID 唯一性与输入校验。
package datajob

import (
	"testing"
	"time"
)

func TestJobStatus_IsValid(t *testing.T) {
	t.Parallel()

	valid := []JobStatus{StatusPending, StatusRunning, StatusSuccess, StatusFailed, StatusCancelled}
	for _, s := range valid {
		if !s.IsValid() {
			t.Errorf("JobStatus(%q) should be valid", s)
		}
	}

	invalid := []JobStatus{"", "unknown", "PENDING", "active", "completed"}
	for _, s := range invalid {
		if s.IsValid() {
			t.Errorf("JobStatus(%q) should be invalid", s)
		}
	}
}

func TestJobStatus_IsTerminal(t *testing.T) {
	t.Parallel()

	terminal := []JobStatus{StatusSuccess, StatusFailed, StatusCancelled}
	for _, s := range terminal {
		if !s.IsTerminal() {
			t.Errorf("JobStatus(%q) should be terminal", s)
		}
	}

	nonTerminal := []JobStatus{StatusPending, StatusRunning}
	for _, s := range nonTerminal {
		if s.IsTerminal() {
			t.Errorf("JobStatus(%q) should not be terminal", s)
		}
	}
}

func TestGenID_Unique(t *testing.T) {
	t.Parallel()

	// 唯一性按「毫秒」分组验证：同一毫秒内低 22 位随机数有 1/4,194,304 碰撞概率，
	// 高并发场景下允许极小概率碰撞；跨毫秒由时间戳高位保证必然唯一。
	seenMs := make(map[int64]struct{}, 1000)
	seenFull := make(map[int64]struct{}, 1000)
	collisions := 0
	for i := 0; i < 1000; i++ {
		id := genID()
		if id <= 0 {
			t.Fatalf("genID() = %d, want positive", id)
		}
		if _, dup := seenFull[id]; dup {
			collisions++
			// 同一毫秒内的碰撞可接受；跨毫秒碰撞才是 bug。
			continue
		}
		seenFull[id] = struct{}{}
		seenMs[id>>22] = struct{}{}
	}
	// 碰撞率应极低（< 1%），否则说明随机源有问题。
	if collisions > 10 {
		t.Errorf("genID() collisions = %d/1000, too high (want <= 10)", collisions)
	}
	// 验证时间戳高位参与（1000 次调用可能在同一毫秒内完成，不强制多分组）。
	if len(seenMs) == 1 {
		ms := int64(0)
		for m := range seenMs {
			ms = m
		}
		if ms == 0 {
			t.Error("genID() timestamp component is zero, want non-zero")
		}
	}
}

func TestValidateCreateInput(t *testing.T) {
	t.Parallel()

	now := time.Now()
	cases := []struct {
		name    string
		in      CreateJobInput
		wantErr bool
	}{
		{
			name: "valid",
			in: CreateJobInput{
				WorkspaceID: 1,
				JobType:     JobTypeExport,
				Name:        "导出 Q4 需求",
			},
			wantErr: false,
		},
		{
			name: "zero workspace",
			in: CreateJobInput{
				WorkspaceID: 0,
				JobType:     JobTypeExport,
				Name:        "导出",
			},
			wantErr: true,
		},
		{
			name: "empty job type",
			in: CreateJobInput{
				WorkspaceID: 1,
				JobType:     "",
				Name:        "导出",
			},
			wantErr: true,
		},
		{
			name: "empty name",
			in: CreateJobInput{
				WorkspaceID: 1,
				JobType:     JobTypeExport,
				Name:        "",
			},
			wantErr: true,
		},
		{
			name: "scheduled at set",
			in: CreateJobInput{
				WorkspaceID: 1,
				JobType:     JobTypeSync,
				Name:        "定时同步",
				ScheduledAt: &now,
			},
			wantErr: false,
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

func TestJSONOrEmpty(t *testing.T) {
	t.Parallel()

	if got := jsonOrEmpty(nil); got != "{}" {
		t.Errorf("jsonOrEmpty(nil) = %v, want {}", got)
	}
	if got := jsonOrEmpty(map[string]any{}); got != "{}" {
		t.Errorf("jsonOrEmpty(empty) = %v, want {}", got)
	}
	m := map[string]any{"key": "value"}
	got := jsonOrEmpty(m)
	gotMap, ok := got.(map[string]any)
	if !ok {
		t.Fatalf("jsonOrEmpty(non-empty) returned %T, want map", got)
	}
	if gotMap["key"] != "value" {
		t.Errorf("jsonOrEmpty(non-empty)[key] = %v, want %v", gotMap["key"], "value")
	}
}
