// Package cache — Redis Pub/Sub 发布订阅封装。
//
// 参考: Redis 官方文档《Pub/Sub》
//
// 对标 Uber Go Style Guide:
//  "Keep interfaces small and focused; let composition scale."
// 这里仅提供 Publish 和 Subscribe 两个极简函数即可满足大多数场景。
package cache

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

// Publish 向指定 channel 发布一条二进制消息。
//
// 参数：
//   - ctx：调用上下文（超时/取消）。
//   - client：已连接的 Redis 客户端。
//   - channel：频道名（建议业务前缀，如 "events:workspace:123"）。
//   - payload：消息体（建议 protobuf 或 JSON 编码的字节流）。
//
// 返回 error 表示发布失败；发布成功返回 nil。
//
// 对标 Google SRE: Pub/Sub 不保证消息持久化，消费者离线时会丢失消息，
// 需要可靠传递的场景应使用 Redis Streams (XADD/XREADGROUP)。
func Publish(ctx context.Context, client *redis.Client, channel string, payload []byte) error {
	if err := client.Publish(ctx, channel, payload).Err(); err != nil {
		return fmt.Errorf("cache: publish to %s: %w", channel, err)
	}
	return nil
}

// Subscribe 订阅一个或多个频道，返回 *redis.PubSub 供消费者循环接收。
//
// 使用模式：
//
//	ps, err := cache.Subscribe(ctx, client, "events:workspace:123")
//	if err != nil { ... }
//	defer ps.Close()
//	ch := ps.Channel()
//	for msg := range ch {
//	    handle(msg.Payload)
//	}
//
// 对标 Uber Go Style: "Always close the subscription when done to avoid goroutine leaks."
func Subscribe(ctx context.Context, client *redis.Client, channels ...string) (*redis.PubSub, error) {
	if len(channels) == 0 {
		return nil, fmt.Errorf("cache: subscribe requires at least one channel")
	}
	ps := client.Subscribe(ctx, channels...)
	// 验证订阅是否成功（Ping 会阻塞直到服务端确认订阅）
	if err := ps.Ping(ctx); err != nil {
		_ = ps.Close()
		return nil, fmt.Errorf("cache: subscribe failed: %w", err)
	}
	return ps, nil
}
