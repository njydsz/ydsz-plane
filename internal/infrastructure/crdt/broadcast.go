// Package crdt — Redis 跨节点广播层。
//
// 协议：
//   - 频道：crdt:doc:{docID}
//   - 消息：JSON {type, data}
//     type = "update"      → data 为 base64 编码的 Y.js update
//     type = "awareness"   → data 为 base64 编码的 awareness update
//     type = "sync_req"    → 请求状态（新客户端加入）
//     type = "sync_res"    → 状态响应（data 为 base64 状态）
//
// 多节点行为：
//   - 本地客户端更新 → 发布到频道 → 其他节点转发给本地客户端
//   - 新客户端加入 → 发布 sync_req → 持有状态的节点响应
//   - 状态写入 Redis（crdt:state:{docID}），作为权威副本
package crdt

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

// Redis 频道与键前缀。
const (
	channelPrefix  = "crdt:doc:"
	stateKeyPrefix = "crdt:state:"
	// 状态 TTL（7 天；活跃文档会持续续期）
	stateTTL = 7 * 24 * time.Hour
)

// BroadcastMessageType 广播消息类型。
type BroadcastMessageType string

const (
	// MsgUpdate 增量更新。
	MsgUpdate BroadcastMessageType = "update"
	// MsgAwareness awareness 更新。
	MsgAwareness BroadcastMessageType = "awareness"
	// MsgSyncRequest 同步请求。
	MsgSyncRequest BroadcastMessageType = "sync_req"
	// MsgSyncResponse 同步响应。
	MsgSyncResponse BroadcastMessageType = "sync_res"
)

// BroadcastMessage 广播消息。
type BroadcastMessage struct {
	Type      BroadcastMessageType `json:"type"`
	Data      string               `json:"data,omitempty"`      // base64
	FromNode  string               `json:"from_node"`           // 发送节点标识
	FromUser  int64                `json:"from_user,omitempty"` // 发送者（awareness 用）
	DocID     string               `json:"doc_id"`
	Timestamp int64                `json:"ts"`
}

// RedisBroadcaster Redis 广播器。
type RedisBroadcaster struct {
	client *redis.Client
	nodeID string
	// 回调：收到远端消息时转发给本地房间
	onRemoteUpdate    func(docID string, data []byte)
	onRemoteAwareness func(docID string, data []byte)
	// 回调：收到同步请求时提供本地状态
	onSyncRequest func(docID string) ([]byte, bool)
}

// NewRedisBroadcaster 构造广播器。
func NewRedisBroadcaster(client *redis.Client, nodeID string) *RedisBroadcaster {
	return &RedisBroadcaster{
		client: client,
		nodeID: nodeID,
	}
}

// SetCallbacks 设置回调。
func (b *RedisBroadcaster) SetCallbacks(
	onUpdate func(docID string, data []byte),
	onAwareness func(docID string, data []byte),
	onSync func(docID string) ([]byte, bool),
) {
	b.onRemoteUpdate = onUpdate
	b.onRemoteAwareness = onAwareness
	b.onSyncRequest = onSync
}

// Start 启动订阅（后台 goroutine）。
func (b *RedisBroadcaster) Start(ctx context.Context, docIDs []string) {
	for _, docID := range docIDs {
		go b.subscribe(ctx, docID)
	}
}

// StartDynamic 启动动态订阅（文档按需加入）。
func (b *RedisBroadcaster) StartDynamic(ctx context.Context) {
	// 使用 Redis Pub/Sub 的 psubscribe 监听所有 crdt:doc:* 频道
	go b.psubscribe(ctx)
}

func (b *RedisBroadcaster) psubscribe(ctx context.Context) {
	pubsub := b.client.PSubscribe(ctx, channelPrefix+"*")
	defer pubsub.Close()

	ch := pubsub.Channel()
	for msg := range ch {
		if msg == nil {
			continue
		}
		docID := msg.Channel[len(channelPrefix):]
		b.handleMessage(docID, []byte(msg.Payload))
	}
}

// subscribe 订阅单个文档频道。
func (b *RedisBroadcaster) subscribe(ctx context.Context, docID string) {
	pubsub := b.client.Subscribe(ctx, channelPrefix+docID)
	defer pubsub.Close()

	ch := pubsub.Channel()
	for msg := range ch {
		if msg == nil {
			continue
		}
		b.handleMessage(docID, []byte(msg.Payload))
	}
}

// handleMessage 处理收到的广播消息。
func (b *RedisBroadcaster) handleMessage(docID string, payload []byte) {
	var msg BroadcastMessage
	if err := json.Unmarshal(payload, &msg); err != nil {
		return
	}
	// 忽略本节点发送的消息
	if msg.FromNode == b.nodeID {
		return
	}

	switch msg.Type {
	case MsgUpdate:
		data, err := base64.StdEncoding.DecodeString(msg.Data)
		if err != nil {
			return
		}
		if b.onRemoteUpdate != nil {
			b.onRemoteUpdate(docID, data)
		}

	case MsgAwareness:
		data, err := base64.StdEncoding.DecodeString(msg.Data)
		if err != nil {
			return
		}
		if b.onRemoteAwareness != nil {
			b.onRemoteAwareness(docID, data)
		}

	case MsgSyncRequest:
		if b.onSyncRequest != nil {
			if state, ok := b.onSyncRequest(docID); ok && len(state) > 0 {
				b.publish(docID, BroadcastMessage{
					Type:      MsgSyncResponse,
					Data:      base64.StdEncoding.EncodeToString(state),
					FromNode:  b.nodeID,
					DocID:     docID,
					Timestamp: time.Now().UnixMilli(),
				})
			}
		}

	case MsgSyncResponse:
		data, err := base64.StdEncoding.DecodeString(msg.Data)
		if err != nil {
			return
		}
		if b.onRemoteUpdate != nil {
			b.onRemoteUpdate(docID, data)
		}
	}
}

// PublishUpdate 发布增量更新到频道。
func (b *RedisBroadcaster) PublishUpdate(docID string, data []byte) error {
	return b.publish(docID, BroadcastMessage{
		Type:      MsgUpdate,
		Data:      base64.StdEncoding.EncodeToString(data),
		FromNode:  b.nodeID,
		DocID:     docID,
		Timestamp: time.Now().UnixMilli(),
	})
}

// PublishAwareness 发布 awareness 更新。
func (b *RedisBroadcaster) PublishAwareness(docID string, data []byte, userID int64) error {
	return b.publish(docID, BroadcastMessage{
		Type:      MsgAwareness,
		Data:      base64.StdEncoding.EncodeToString(data),
		FromNode:  b.nodeID,
		FromUser:  userID,
		DocID:     docID,
		Timestamp: time.Now().UnixMilli(),
	})
}

// RequestSync 请求其他节点提供状态。
func (b *RedisBroadcaster) RequestSync(docID string) error {
	return b.publish(docID, BroadcastMessage{
		Type:      MsgSyncRequest,
		FromNode:  b.nodeID,
		DocID:     docID,
		Timestamp: time.Now().UnixMilli(),
	})
}

// SaveState 保存状态到 Redis（权威副本）。
func (b *RedisBroadcaster) SaveState(docID string, state []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return b.client.Set(ctx, stateKeyPrefix+docID, state, stateTTL).Err()
}

// LoadState 从 Redis 加载状态。
func (b *RedisBroadcaster) LoadState(docID string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return b.client.Get(ctx, stateKeyPrefix+docID).Bytes()
}

// publish 发布消息到频道。
func (b *RedisBroadcaster) publish(docID string, msg BroadcastMessage) error {
	payload, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return b.client.Publish(ctx, channelPrefix+docID, payload).Err()
}

var _ = log.Printf
