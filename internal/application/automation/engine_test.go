// Package automation 测试：校验自动化引擎构造与辅助方法。
//
// EvaluateEvent 的完整链路测试（防循环、规则匹配、动作分发）需要真实 DB，
// 通过 go test -tags dockertest 在集成测试文件中激活。
// 本文件覆盖无 DB 依赖的纯路径：构造器、链式配置、extractProjectID。
package automation

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/njydsz/ydsz-plane/internal/infrastructure/mq"
)

// mockActionExecutor 零实现。
type mockActionExecutor struct{}

func (m *mockActionExecutor) TransitionIssueStatus(_, _, _, _ int64, _ string) error { return nil }
func (m *mockActionExecutor) AssignIssue(_, _, _, _, _ int64) error                   { return nil }
func (m *mockActionExecutor) UpdateIssueField(_, _, _, _ int64, _ string, _ any) error { return nil }
func (m *mockActionExecutor) SendNotification(_ interface{ Done() <-chan struct{} }, _ NotificationRequest) error {
	return nil
}
func (m *mockActionExecutor) CreateIssue(_ interface{ Done() <-chan struct{} }, _ CreateIssueRequest) (int64, error) {
	return 0, nil
}

type mockCtxProvider struct{}

func (m *mockCtxProvider) BuildContext(_ interface{ Done() <-chan struct{} }, _ mq.EventEnvelope) (*ExecutionContext, error) {
	return &ExecutionContext{}, nil
}

// TestEngine_NewEngineNilLogger verifies nil logger is handled gracefully.
func TestEngine_NewEngineNilLogger(t *testing.T) {
	eng := NewEngine(&Service{}, &mockActionExecutor{}, &mockCtxProvider{}, nil)
	assert.NotNil(t, eng)
	assert.NotNil(t, eng.metrics)
	assert.NotNil(t, eng.breakers)
}

// TestEngine_WithMetrics verifies chainable WithMetrics returns the same instance.
func TestEngine_WithMetrics(t *testing.T) {
	eng := NewEngine(&Service{}, &mockActionExecutor{}, &mockCtxProvider{}, zap.NewNop())
	result := eng.WithMetrics(&Metrics{})
	assert.Same(t, eng, result)
}

// TestEngine_WithCircuitBreaker verifies chainable WithCircuitBreaker returns the same instance.
func TestEngine_WithCircuitBreaker(t *testing.T) {
	eng := NewEngine(&Service{}, &mockActionExecutor{}, &mockCtxProvider{}, zap.NewNop())
	cb := NewCircuitBreakerRegistry(DefaultCircuitBreakerConfig)
	result := eng.WithCircuitBreaker(cb)
	assert.Same(t, eng, result)
}

// TestEngine_ExtractProjectID verifies project ID is extracted from event payload.
func TestEngine_ExtractProjectID(t *testing.T) {
	eng := NewEngine(&Service{}, &mockActionExecutor{}, &mockCtxProvider{}, zap.NewNop())

	// Case 1: payload has project_id
	event := mq.EventEnvelope{
		Payload: map[string]any{"project_id": float64(42)},
	}
	pid := eng.extractProjectID(event)
	assert.NotNil(t, pid)
	assert.Equal(t, int64(42), *pid)

	// Case 2: payload is nil — should return nil without panic
	event2 := mq.EventEnvelope{Payload: nil}
	pid2 := eng.extractProjectID(event2)
	assert.Nil(t, pid2)

	// Case 3: payload without project_id
	event3 := mq.EventEnvelope{Payload: map[string]any{"other": "value"}}
	pid3 := eng.extractProjectID(event3)
	assert.Nil(t, pid3)
}

// TestEngine_ChainableConfig verifies full chain configuration.
func TestEngine_ChainableConfig(t *testing.T) {
	eng := NewEngine(&Service{}, &mockActionExecutor{}, &mockCtxProvider{}, nil)
	metrics := &Metrics{}
	cb := NewCircuitBreakerRegistry(DefaultCircuitBreakerConfig)

	result := eng.WithMetrics(metrics).WithCircuitBreaker(cb)
	assert.Same(t, eng, result)
	assert.Equal(t, metrics, eng.metrics)
	assert.Equal(t, cb, eng.breakers)
}
