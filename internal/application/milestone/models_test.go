// Package milestone — 领域模型单元测试。
//
// 覆盖：MilestoneStatus 校验、genID 唯一性与时间戳参与。
package milestone

import (
	"testing"
)

func TestMilestoneStatus_IsValid(t *testing.T) {
	t.Parallel()

	valid := []MilestoneStatus{
		StatusNotStarted, StatusInProgress, StatusCompleted, StatusMissed,
	}
	for _, s := range valid {
		if !s.IsValid() {
			t.Errorf("MilestoneStatus(%q) should be valid", s)
		}
	}

	invalid := []MilestoneStatus{"", "unknown", "NOT_STARTED", "active"}
	for _, s := range invalid {
		if s.IsValid() {
			t.Errorf("MilestoneStatus(%q) should be invalid", s)
		}
	}
}

func TestGenID_Unique(t *testing.T) {
	t.Parallel()

	// 按毫秒分组验证：同毫秒内低 22 位随机数有小概率碰撞，
	// 跨毫秒由时间戳高位保证唯一。
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
			continue
		}
		seenFull[id] = struct{}{}
		seenMs[id>>22] = struct{}{}
	}
	if collisions > 10 {
		t.Errorf("genID() collisions = %d/1000, too high (want <= 10)", collisions)
	}
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
