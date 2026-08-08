// Package eventbus — 事件发布抽象层。
//
// 目的：在单体 ↔ 微服务切换时，无需修改业务代码。
//
// 实现（按优先级）：
//   1. LocalPublisher  - 当前单体模式，直接调用本地 Notification Service（零开销）
//   2. RabbitMQPublisher - 微服务模式，通过消息队列投递给独立 Notification Service 消费
//   3. gRPCPublisher   - 微服务模式，通过 gRPC 同步调用（需要注意 RPC 失败降级为事件）
//
// 切换方式：通过配置项 eventbus.publisher = "local" | "rabbitmq" | "grpc"
package eventbus

import (
	"context"
	"fmt"
	"sync"
)

// DomainEvent 是应用内领域事件的统一信封格式。
type DomainEvent struct {
	EventType  string `json:"event_type"` // e.g. "issue.created"
	EntityType string `json:"entity_type"`
	EntityID   int64  `json:"entity_id"`
	WorkspaceID int64 `json:"workspace_id"`
	ActorID    int64  `json:"actor_id"`
	ActorName  string `json:"actor_name"`
	Payload    []byte `json:"payload"`    // JSON 格式的详情
}

// Publisher 事件发布接口。
type Publisher interface {
	// Publish 投递一个领域事件（异步语义 - 非阻塞，写入 outbox 或 MQ）。
	Publish(ctx context.Context, event DomainEvent) error

	// PublishSync 同步投递（等待下游确认，用于关键业务通知）。
	PublishSync(ctx context.Context, event DomainEvent) error

	// Close 关闭发布连接。
	Close() error
}

// --- LocalPublisher（单体模式默认实现）---

// LocalPublisher 通过直接函数调用投递事件给本地订阅者。
// 在单体模式下零网络开销，单元测试友好。
type LocalPublisher struct {
	mu        sync.RWMutex
	handlers  []func(ctx context.Context, event DomainEvent) error
}

// NewLocalPublisher 创建一个本地事件发布器。
func NewLocalPublisher() *LocalPublisher {
	return &LocalPublisher{}
}

// RegisterHandler 注册事件处理函数。
func (p *LocalPublisher) RegisterHandler(handler func(ctx context.Context, event DomainEvent) error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.handlers = append(p.handlers, handler)
}

// Publish 异步投递（把事件放入内存 channel 或直接调用 handler）。
func (p *LocalPublisher) Publish(ctx context.Context, event DomainEvent) error {
	p.mu.RLock()
	handlers := make([]func(ctx context.Context, event DomainEvent) error, len(p.handlers))
	copy(handlers, p.handlers)
	p.mu.RUnlock()

	var lastErr error
	for _, h := range handlers {
		if err := h(ctx, event); err != nil {
			lastErr = err
		}
	}
	return lastErr
}

// PublishSync 同步投递（与 Publish 同义 - 因为是本地调用）。
func (p *LocalPublisher) PublishSync(ctx context.Context, event DomainEvent) error {
	return p.Publish(ctx, event)
}

// Close 关闭本地发布器（无-op）。
func (p *LocalPublisher) Close() error {
	return nil
}

// --- Factory ---

// Config 创建 Publisher 的配置参数。
type Config struct {
	Mode string // "local" | "rabbitmq" | "grpc"

	// RabbitMQ 参数
	RabbitURL string

	// gRPC 参数
	GRPCTarget string // "notification-service:9090"
}

// NewPublisher 根据配置创建对应实现。
func NewPublisher(cfg Config) (Publisher, error) {
	switch cfg.Mode {
	case "local", "":
		return NewLocalPublisher(), nil
	case "rabbitmq":
		return newRabbitMQPublisher(cfg.RabbitURL)
	case "grpc":
		return newGRPCPublisher(cfg.GRPCTarget)
	default:
		return nil, fmt.Errorf("unknown eventbus publisher mode: %s", cfg.Mode)
	}
}

// --- Stubs (to be implemented when rabbitmq/grpc split happens) ---

func newRabbitMQPublisher(rabbitURL string) (Publisher, error) {
	return nil, fmt.Errorf("rabbitmq publisher not yet implemented - pending S14 Phase-1")
}

func newGRPCPublisher(target string) (Publisher, error) {
	return nil, fmt.Errorf("grpc publisher not yet implemented - pending S14 Phase-2")
}
