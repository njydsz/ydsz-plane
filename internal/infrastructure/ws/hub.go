// Package ws 提供基于 Redis Pub/Sub 的多节点 WebSocket 广播 Hub。
//
// 架构：
//   客户端 ↔ API Node (gorilla/websocket) ↔ Redis PubSub ↔ 其他 API Node
//
// 每个 API 节点维护本地连接池，通过 Redis 频道实现跨节点广播。
// 频道命名：plane:ws:{workspace_id}
package ws

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

// upgrader 将 HTTP 连接升级为 WebSocket。
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // 开发环境允许所有来源，生产环境应限制
	},
}

// Message 通过 WebSocket 发送的通用消息。
type Message struct {
	Type string          `json:"type"` // "issue.updated" | "sprint.changed" | "version.released" | "ping"
	Data json.RawMessage `json:"data"`
}

// Client 表示一个 WebSocket 连接。
type Client struct {
	Conn     *websocket.Conn
	UserID   int64
	Send     chan []byte
	hub      *Hub
	mu       sync.Mutex
}

// Hub 管理所有 WebSocket 连接和 Redis Pub/Sub。
type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
	rdb        *redis.Client
	ctx        context.Context
	cancel     context.CancelFunc
}

// NewHub 创建 WebSocket Hub。
func NewHub(rdb *redis.Client) *Hub {
	ctx, cancel := context.WithCancel(context.Background())
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		rdb:        rdb,
		ctx:        ctx,
		cancel:     cancel,
	}
}

// Run 启动 Hub 的主事件循环。
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.Send)
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.Send <- message:
				default:
					// 客户端发送缓冲区满，关闭连接
					go h.unregisterClient(client)
				}
			}
			h.mu.RUnlock()
		}
	}
}

// Shutdown 优雅关闭 Hub。
func (h *Hub) Shutdown() {
	h.cancel()
}

// HandleWebSocket 处理 WebSocket 升级请求。
func (h *Hub) HandleWebSocket(w http.ResponseWriter, r *http.Request, userID int64, workspaceID int64) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("ws: upgrade error: %v", err)
		return
	}

	client := &Client{
		Conn:   conn,
		UserID: userID,
		Send:   make(chan []byte, 64),
		hub:    h,
	}

	h.register <- client

	// 订阅 Redis 频道
	pubsub := h.rdb.Subscribe(h.ctx, workspaceChannel(workspaceID))
	defer pubsub.Close()

	// 写入协程
	go client.writePump()

	// 读取协程（心跳 + 关闭检测）
	go client.readPump()

	// Redis 消息转发到客户端
	go func() {
		ch := pubsub.Channel()
		for msg := range ch {
			select {
			case client.Send <- []byte(msg.Payload):
			default:
			}
		}
	}()
}

// Publish 通过 Redis Pub/Sub 向指定工作空间广播消息。
// 支持跨节点广播：发布到 Redis，所有订阅该频道的节点都会收到。
func (h *Hub) Publish(ctx context.Context, workspaceID int64, msg Message) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return h.rdb.Publish(ctx, workspaceChannel(workspaceID), data).Err()
}

// BroadcastLocal 仅向本地连接广播（不经过 Redis）。
func (h *Hub) BroadcastLocal(msg Message) {
	data, _ := json.Marshal(msg)
	h.broadcast <- data
}

func (h *Hub) unregisterClient(client *Client) {
	client.Conn.Close()
	h.unregister <- client
}

func (c *Client) writePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			c.mu.Lock()
			err := c.Conn.WriteMessage(websocket.TextMessage, message)
			c.mu.Unlock()
			if err != nil {
				return
			}
		case <-ticker.C:
			// 心跳
			c.mu.Lock()
			err := c.Conn.WriteMessage(websocket.PingMessage, nil)
			c.mu.Unlock()
			if err != nil {
				return
			}
		}
	}
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.Conn.Close()
	}()
	c.Conn.SetReadLimit(4096)
	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})
	for {
		_, _, err := c.Conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

func workspaceChannel(workspaceID int64) string {
	return "plane:ws:" + itoa(workspaceID)
}

func itoa(v int64) string {
	if v == 0 {
		return "0"
	}
	neg := false
	if v < 0 {
		neg = true
		v = -v
	}
	var buf [20]byte
	i := len(buf)
	for v > 0 {
		i--
		buf[i] = byte(v%10) + '0'
		v /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
