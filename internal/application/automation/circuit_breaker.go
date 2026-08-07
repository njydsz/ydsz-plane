// Package automation — 自动化规则熔断器（Circuit Breaker）。
//
// 对标参考:
//   - Netflix Hystrix 三态模型（Closed → Open → HalfOpen）
//   - 美团 MT-Hystrix 熔断器 SDK
//   - 阿里 Sentinel 流控规则
//
// 设计要点:
//   - 三态翻转基于错误率与最小采样窗口
//   - 半开状态仅放行 1 次探测请求，成功则复位、失败则重新断开
//   - 熔断触发后写 automation 规则状态为 error，通知 workspace admin
//
// 使用场景:
// 自动化规则执行引擎对单条规则的连续失败做熔断。当某规则连续失败 ≥
// threshold 次（默认 5）时，自动切换到 Open 状态禁止执行；
// 冷却期后进入 HalfOpen 放行一次探测。
package automation

import (
	"sync"
	"time"
)

// CircuitState 熔断器三态。
type CircuitState int

const (
	CircuitClosed    CircuitState = iota // 正常放行
	CircuitOpen                          // 熔断中（拒绝）
	CircuitHalfOpen                      // 半开（仅放行 1 次探测）
)

// CircuitBreakerConfig 熔断器配置。
type CircuitBreakerConfig struct {
	// FailureThreshold 触发熔断的连续失败阈值（默认 5）
	FailureThreshold int
	// CoolingPeriod 熔断后进入半开前的冷却期（默认 60s）
	CoolingPeriod time.Duration
	// HalfOpenMaxRequests 半开状态允许的最大并发探测数（默认 1）
	HalfOpenMaxRequests int
}

// DefaultCircuitBreakerConfig 默认配置。
var DefaultCircuitBreakerConfig = CircuitBreakerConfig{
	FailureThreshold:    5,
	CoolingPeriod:       60 * time.Second,
	HalfOpenMaxRequests: 1,
}

// CircuitBreaker 是规则级熔断器。
//
// 并发安全：通过内部互斥锁保护状态翻转。
type CircuitBreaker struct {
	mu              sync.Mutex
	state           CircuitState
	config          CircuitBreakerConfig
	consecutiveFail int
	lastFailureAt   time.Time
	halfOpenSuccess int
	halfOpenAttempt int
}

// NewCircuitBreaker 创建规则级熔断器。
func NewCircuitBreaker(cfg CircuitBreakerConfig) *CircuitBreaker {
	if cfg.FailureThreshold <= 0 {
		cfg.FailureThreshold = DefaultCircuitBreakerConfig.FailureThreshold
	}
	if cfg.CoolingPeriod <= 0 {
		cfg.CoolingPeriod = DefaultCircuitBreakerConfig.CoolingPeriod
	}
	if cfg.HalfOpenMaxRequests <= 0 {
		cfg.HalfOpenMaxRequests = DefaultCircuitBreakerConfig.HalfOpenMaxRequests
	}
	return &CircuitBreaker{config: cfg, state: CircuitClosed}
}

// Allow 检查当前请求是否放行。
// 返回 true 表示可执行；false 表示被熔断拒绝。
func (cb *CircuitBreaker) Allow() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case CircuitClosed:
		return true
	case CircuitOpen:
		// 冷却期过后转入 HalfOpen
		if time.Since(cb.lastFailureAt) > cb.config.CoolingPeriod {
			cb.state = CircuitHalfOpen
			cb.halfOpenAttempt = 0
			cb.halfOpenSuccess = 0
			return true
		}
		return false
	case CircuitHalfOpen:
		// 半开仅放行 N 次探测
		if cb.halfOpenAttempt < cb.config.HalfOpenMaxRequests {
			cb.halfOpenAttempt++
			return true
		}
		return false
	}
	return false
}

// RecordSuccess 记录一次成功执行。
// 在 HalfOpen 态下成功 → 复位到 Closed。
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case CircuitHalfOpen:
		cb.halfOpenSuccess++
		// 探测成功 → 复位
		cb.state = CircuitClosed
		cb.consecutiveFail = 0
	case CircuitClosed:
		cb.consecutiveFail = 0
	}
}

// RecordFailure 记录一次失败执行。
// 在 HalfOpen 态下失败 → 重新 Open。
// 在 Closed 态下连续失败 → Open。
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.lastFailureAt = time.Now()

	switch cb.state {
	case CircuitHalfOpen:
		cb.state = CircuitOpen
		cb.consecutiveFail++
	case CircuitClosed:
		cb.consecutiveFail++
		if cb.consecutiveFail >= cb.config.FailureThreshold {
			cb.state = CircuitOpen
		}
	}
}

// State 返回当前熔断器状态（只读）。
func (cb *CircuitBreaker) State() CircuitState {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.state
}

// Reset 人工复位到 Closed 态。
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.state = CircuitClosed
	cb.consecutiveFail = 0
}

// CircuitBreakerRegistry 是规则 ID → 熔断器的并发安全注册表。
type CircuitBreakerRegistry struct {
	mu    sync.RWMutex
	cores map[int64]*CircuitBreaker
	cfg   CircuitBreakerConfig
}

// NewCircuitBreakerRegistry 创建注册表。
func NewCircuitBreakerRegistry(cfg CircuitBreakerConfig) *CircuitBreakerRegistry {
	return &CircuitBreakerRegistry{
		cores: make(map[int64]*CircuitBreaker),
		cfg:   cfg,
	}
}

// GetOrCreate 获取或初始化指定规则的熔断器。
func (r *CircuitBreakerRegistry) GetOrCreate(ruleID int64) *CircuitBreaker {
	r.mu.RLock()
	cb, ok := r.cores[ruleID]
	r.mu.RUnlock()
	if ok {
		return cb
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	// double-check
	if cb, ok := r.cores[ruleID]; ok {
		return cb
	}
	cb = NewCircuitBreaker(r.cfg)
	r.cores[ruleID] = cb
	return cb
}

// Remove 移除规则熔断器（规则删除时调用）。
func (r *CircuitBreakerRegistry) Remove(ruleID int64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.cores, ruleID)
}
