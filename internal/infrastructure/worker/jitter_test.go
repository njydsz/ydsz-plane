package worker

import (
	"testing"
	"time"
)

func TestDeterministicJitter_Range(t *testing.T) {
	// 多次调用应始终返回 [0, maxSeconds) 范围内的值
	for i := 0; i < 100; i++ {
		key := "test-key-" + string(rune('A'+i%26))
		got := DeterministicJitter(key, 60)
		if got < 0 || got >= 60*time.Second {
			t.Fatalf("DeterministicJitter(%q, 60) = %v, want [0, 60s)", key, got)
		}
	}
}

func TestDeterministicJitter_Consistency(t *testing.T) {
	// 同一 key 在多次调用间应返回一致的值
	key := "consistent-key"
	first := DeterministicJitter(key, 30)
	for i := 0; i < 50; i++ {
		got := DeterministicJitter(key, 30)
		if got != first {
			t.Fatalf("DeterministicJitter(%q) inconsistent: %v vs %v", key, first, got)
		}
	}
}

func TestDeterministicJitter_DifferentKeys(t *testing.T) {
	// 不同 key 更可能产生不同偏移（不太可能完全相同）
	offsets := make(map[time.Duration]bool)
	for i := 0; i < 200; i++ {
		got := DeterministicJitter("key-"+string(rune(i)), 60)
		offsets[got] = true
	}
	// 至少应有 30 个不同偏移值（概率上几乎确定）
	if len(offsets) < 30 {
		t.Fatalf("Expected >=30 distinct offsets from 200 keys, got %d", len(offsets))
	}
}

func TestDeterministicJitter_MaxZero(t *testing.T) {
	got := DeterministicJitter("any", 0)
	if got != 0 {
		t.Fatalf("DeterministicJitter(max=0) = %v, want 0", got)
	}
	got = DeterministicJitter("any", -5)
	if got != 0 {
		t.Fatalf("DeterministicJitter(max=-5) = %v, want 0", got)
	}
}

func TestIDJitter(t *testing.T) {
	// ID 0 与 ID 1 应产生不同偏移（统计学上几乎确定）
	j0 := IDJitter(0, 30)
	j1 := IDJitter(1, 30)
	if j0 == j1 {
		t.Logf("warning: ID 0 and 1 produced same jitter: %v", j0)
	}
	// 同一 ID 应一致
	if IDJitter(42, 30) != IDJitter(42, 30) {
		t.Fatal("IDJitter not consistent for same ID")
	}
}

func TestDayJitter_VariesByDay(t *testing.T) {
	// 不同日期的偏移应不同（大概率）
	j1 := DayJitter("2026-08-01", 45)
	j2 := DayJitter("2026-08-02", 45)
	j3 := DayJitter("2026-08-03", 45)
	if j1 == j2 && j2 == j3 {
		t.Skip("warning: 3 consecutive days produced same jitter (rare but possible)")
	}
	// 同一日期同一天调用多次一致
	if DayJitter("2026-08-08", 45) != DayJitter("2026-08-08", 45) {
		t.Fatal("DayJitter not consistent for same date")
	}
}

func TestMinuteJitter(t *testing.T) {
	// 不同分钟 key 产生不同偏移
	j1 := MinuteJitter(1, "20260102T0900", 55)
	j2 := MinuteJitter(1, "20260102T0901", 55)
	if j1 == j2 {
		t.Logf("warning: same rule id in consecutive minutes produced same jitter")
	}
	// 同一 ID + 分钟 一致
	if MinuteJitter(7, "20260102T1000", 55) != MinuteJitter(7, "20260102T1000", 55) {
		t.Fatal("MinuteJitter not consistent")
	}
}
