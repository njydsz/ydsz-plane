package automation

import (
	"testing"
	"time"
)

func TestCircuitBreaker_ClosedToOpen(t *testing.T) {
	cb := NewCircuitBreaker(CircuitBreakerConfig{
		FailureThreshold: 3,
		CoolingPeriod:    100 * time.Millisecond,
	})

	// 前 3 次放行
	if !cb.Allow() {
		t.Fatal("closed state should allow")
	}
	cb.RecordFailure()

	if !cb.Allow() {
		t.Fatal("1st failure below threshold, should still allow")
	}
	cb.RecordFailure()

	if !cb.Allow() {
		t.Fatal("2nd failure at threshold, should still allow (3rd call triggers)")
	}
	cb.RecordFailure()

	// 第 3 次失败触发熔断
	if cb.State() != CircuitOpen {
		t.Fatalf("expected Open after 3 failures, got %v", cb.State())
	}

	// Open 态拒绝
	if cb.Allow() {
		t.Fatal("open state should deny")
	}
}

func TestCircuitBreaker_OpenToHalfOpen_ThenSuccess(t *testing.T) {
	cb := NewCircuitBreaker(CircuitBreakerConfig{
		FailureThreshold:    1,
		CoolingPeriod:       50 * time.Millisecond,
		HalfOpenMaxRequests: 1,
	})

	// 触发熔断
	cb.RecordFailure()
	if cb.State() != CircuitOpen {
		t.Fatalf("expected Open, got %v", cb.State())
	}

	// 冷却期内仍拒绝
	if cb.Allow() {
		t.Fatal("open state during cooling should deny")
	}

	// 等待冷却期
	time.Sleep(60 * time.Millisecond)

	// 半开放行一次
	if !cb.Allow() {
		t.Fatal("half-open should allow probe")
	}

	// 探测成功 → 复位
	cb.RecordSuccess()
	if cb.State() != CircuitClosed {
		t.Fatalf("expected Closed after half-open success, got %v", cb.State())
	}
}

func TestCircuitBreaker_HalfOpenFail(t *testing.T) {
	cb := NewCircuitBreaker(CircuitBreakerConfig{
		FailureThreshold:    1,
		CoolingPeriod:       50 * time.Millisecond,
		HalfOpenMaxRequests: 1,
	})

	cb.RecordFailure() // Open
	time.Sleep(60 * time.Millisecond)

	if !cb.Allow() {
		t.Fatal("half-open should allow first probe")
	}
	cb.RecordFailure() // 探测失败 → Open

	if cb.State() != CircuitOpen {
		t.Fatalf("expected Open after half-open failure, got %v", cb.State())
	}
}

func TestCircuitBreaker_HalfOpenMaxAttempts(t *testing.T) {
	cb := NewCircuitBreaker(CircuitBreakerConfig{
		FailureThreshold:    1,
		CoolingPeriod:       50 * time.Millisecond,
		HalfOpenMaxRequests: 2, // 允许 2 次探测
	})

	cb.RecordFailure()
	time.Sleep(60 * time.Millisecond)

	// 应放行 2 次
	if !cb.Allow() {
		t.Fatal("first probe should allow")
	}
	if !cb.Allow() {
		t.Fatal("second probe should allow")
	}
	// 第 3 次拒绝
	if cb.Allow() {
		t.Fatal("third probe should deny (max exceeded)")
	}
}

func TestCircuitBreakerRegistry_GetOrCreate(t *testing.T) {
	reg := NewCircuitBreakerRegistry(DefaultCircuitBreakerConfig)

	cb1 := reg.GetOrCreate(100)
	cb2 := reg.GetOrCreate(100)
	if cb1 != cb2 {
		t.Fatal("registry should return same breaker for same rule")
	}

	cb3 := reg.GetOrCreate(200)
	if cb1 == cb3 {
		t.Fatal("different rule IDs should have different breakers")
	}

	reg.Remove(100)
	cb4 := reg.GetOrCreate(100)
	if cb1 == cb4 {
		t.Fatal("after Remove, should create new breaker")
	}
}

func TestCircuitBreaker_Reset(t *testing.T) {
	cb := NewCircuitBreaker(CircuitBreakerConfig{
		FailureThreshold: 1,
	})

	cb.RecordFailure()
	if cb.State() != CircuitOpen {
		t.Fatal("should be Open")
	}

	cb.Reset()
	if cb.State() != CircuitClosed {
		t.Fatal("Reset should move to Closed")
	}
	if !cb.Allow() {
		t.Fatal("after Reset should allow")
	}
}

func TestCircuitBreaker_RecordSuccess_ResetsConsecutive(t *testing.T) {
	cb := NewCircuitBreaker(CircuitBreakerConfig{
		FailureThreshold: 3,
	})

	cb.RecordFailure()
	cb.RecordFailure()
	cb.RecordSuccess() // 复位

	cb.RecordFailure()
	if cb.State() != CircuitClosed {
		t.Fatalf("after reset + 1 fail, should still be Closed, got %v", cb.State())
	}
}
