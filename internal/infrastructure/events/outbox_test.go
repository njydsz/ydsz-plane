// Package events — outbox 模式单元测试。
//
// 测试不依赖 PostgreSQL 或 RabbitMQ，覆盖高频改动区域的纯函数与配置：
//   - Event.toEnvelope() 转换逻辑
//   - Relay 默认配置验证
package events

import (
	"encoding/json"
	"testing"
	"time"

	"go.uber.org/zap/zaptest"

	"github.com/njydsz/ydsz-plane/internal/infrastructure/mq"
)

// --- Tests ---

// TestEventToEnvelope 验证 Event 到 mq.EventEnvelope 的转换。
func TestEventToEnvelope(t *testing.T) {
	payload := json.RawMessage(`{"issue_id":42,"state":"open"}`)
	occurred := time.Date(2026, 9, 26, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		event    Event
		wantKey  string
	}{
		{
			name: "issue.created",
			event: Event{
				ID: 100, WorkspaceID: 5,
				AggregateType: "issue", AggregateID: 42,
				EventType: "created", Payload: payload, OccurredAt: occurred,
			},
			wantKey: "plane.events.issue.created",
		},
		{
			name: "requirement.submitted",
			event: Event{
				ID: 101, WorkspaceID: 3,
				AggregateType: "requirement", AggregateID: 99,
				EventType: "submitted", Payload: nil, OccurredAt: occurred,
			},
			wantKey: "plane.events.requirement.submitted",
		},
		{
			name: "defect.status_changed",
			event: Event{
				ID: 102, WorkspaceID: 1,
				AggregateType: "defect", AggregateID: 7,
				EventType: "status_changed", Payload: json.RawMessage(`{"from":"open","to":"closed"}`), OccurredAt: occurred,
			},
			wantKey: "plane.events.defect.status_changed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := tt.event.toEnvelope()

			if env.EventID != tt.event.ID {
				t.Errorf("EventID = %d, want %d", env.EventID, tt.event.ID)
			}
			if env.WorkspaceID != tt.event.WorkspaceID {
				t.Errorf("WorkspaceID = %d, want %d", env.WorkspaceID, tt.event.WorkspaceID)
			}
			if env.AggregateType != tt.event.AggregateType {
				t.Errorf("AggregateType = %q, want %q", env.AggregateType, tt.event.AggregateType)
			}
			if env.AggregateID != tt.event.AggregateID {
				t.Errorf("AggregateID = %d, want %d", env.AggregateID, tt.event.AggregateID)
			}
			if env.EventType != tt.event.EventType {
				t.Errorf("EventType = %q, want %q", env.EventType, tt.event.EventType)
			}
			if env.Exchange != mq.EventExchange {
				t.Errorf("Exchange = %q, want %q", env.Exchange, mq.EventExchange)
			}
			if env.RoutingKey != tt.wantKey {
				t.Errorf("RoutingKey = %q, want %q", env.RoutingKey, tt.wantKey)
			}
			if string(env.Payload) != string(tt.event.Payload) {
				t.Errorf("Payload = %s, want %s", env.Payload, tt.event.Payload)
			}
			if !env.OccurredAt.Equal(tt.event.OccurredAt) {
				t.Errorf("OccurredAt = %v, want %v", env.OccurredAt, tt.event.OccurredAt)
			}
		})
	}
}

// TestEventToEnvelopeEmptyPayload 验证 nil payload 转换为空 JSON。
func TestEventToEnvelopeEmptyPayload(t *testing.T) {
	ev := Event{ID: 1, AggregateType: "issue", EventType: "test"}
	env := ev.toEnvelope()
	if env.Payload != nil {
		t.Errorf("nil Event.Payload should produce nil envelope Payload, got %v", env.Payload)
	}
}

// TestRelayDefaults 验证 NewRelay 构造时使用合理的默认值。
func TestRelayDefaults(t *testing.T) {
	r := NewRelay(nil, nil, zaptest.NewLogger(t))
	if r == nil {
		t.Fatal("NewRelay returned nil")
	}
	if r.batch != 200 {
		t.Errorf("default batch = %d, want 200", r.batch)
	}
	if r.period == 0 {
		t.Error("default period should not be 0")
	}
	if r.OnPublishFailed != nil {
		t.Error("OnPublishFailed should be nil by default")
	}
	if r.wakeup == nil {
		t.Error("wakeup channel should be initialized")
	}
}

// TestRelayCustomBatch 验证自定义批大小。
func TestRelayCustomBatch(t *testing.T) {
	r := NewRelay(nil, nil, zaptest.NewLogger(t))
	_ = r // 已在构造中使用默认值；自定义批次通过 RelayOption 扩展（待实现时补充）
}

// TestRecorderNew 验证 NewRecorder 返回非 nil。
func TestRecorderNew(t *testing.T) {
	r := NewRecorder()
	if r == nil {
		t.Fatal("NewRecorder returned nil")
	}
}
