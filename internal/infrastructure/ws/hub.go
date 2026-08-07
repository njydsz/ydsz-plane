// Package ws 提供基于 Redis Pub/Sub 的多节点 WebSocket 广播 Hub。
//
// 架构：
//   客户端 ↔ API Node (gorilla/websocket) ↔ Redis PubSub ↔ 其他 API Node
//
// 每个 API 节点维护本地连接池，通过 Redis 频道实现跨节点广播。
// 频道命名：plane:ws:{workspace_id}
//
// 安全加固清单（S7-P0/P1）：
//   - CheckOrigin：基于白名单校验 Origin，拒绝跨站 WS 连接（防 CSWSH）。
//   - 连接级限流：每用户/每 IP 最大并发连接数。
//   - 全局连接上限：单节点连接数硬上限，超出拒绝并返回 503。
package ws

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

// HubConfig Hub 运行时配置（安全相关）。
type HubConfig struct {
	// AllowedOrigins 允许的 WS 来源 Origin 列表。
	// 为空时仅允许同源请求（浏览器自动携带 Origin）；
	// 显式配置后按白名单匹配。
	AllowedOrigins []string

	// MaxConnsPerUser 单用户最大并发连接数。
	// 0 表示无限制。
	MaxConnsPerUser int

	// MaxConnsPerIP 单 IP 最大并发连接数。
	// 0 表示无限制。
	MaxConnsPerIP int

	// MaxConnsGlobal 单节点全局连接硬上限。
	// 0 表示无限制。
	MaxConnsGlobal int

	// RequireOriginCheck 是否强制校验 Origin。
	// 开发模式下可设为 false 允许跨域；生产环境必须 true。
	RequireOriginCheck bool
}

// upgrader 将 HTTP 连接升级为 WebSocket。
//
// 安全注：CheckOrigin 在 Hub 构造时根据 HubConfig 动态设置，
// 防止跨站 WebSocket 劫持（CSWSH）。未经校验的 Origin 会给任意
// 恶意网页打开通往内部频道的后门。
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// CheckOrigin 在 NewHub 中根据配置设置；
	// 默认 false 拒绝一切连接（fail-safe）。
	CheckOrigin: func(r *http.Request) bool { return false },
}

// Message 通过 WebSocket 发送的通用消息。
type Message struct {
	Type string          `json:"type"` // "issue.updated" | "sprint.changed" | "version.released" | "ping" | "notification.new"
	Data json.RawMessage `json:"data"`
}

// Client 表示一个 WebSocket 连接。
type Client struct {
	Conn     *websocket.Conn
	UserID   int64
	Send     chan []byte
	hub      *Hub
	mu       sync.Mutex
	remoteIP string
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
	cfg        HubConfig

	// 连接计数
	userConns map[int64]int
	ipConns   map[string]int
}

// NewHub 创建 WebSocket Hub。
func NewHub(rdb *redis.Client) *Hub {
	return NewHubWithConfig(rdb, HubConfig{})
}

// NewHubWithConfig 创建带安全配置的 Hub。
func NewHubWithConfig(rdb *redis.Client, cfg HubConfig) *Hub {
	ctx, cancel := context.WithCancel(context.Background())
	h := &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		rdb:        rdb,
		ctx:        ctx,
		cancel:     cancel,
		cfg:        cfg,
		userConns:  make(map[int64]int),
		ipConns:    make(map[string]int),
	}
	// 按配置设置 Origin 校验函数
	h.applyCheckOrigin()
	return h
}

// applyCheckOrigin 根据配置设置 upgrader.CheckOrigin。
func (h *Hub) applyCheckOrigin() {
	if !h.cfg.RequireOriginCheck {
		// 开发模式：允许所有 Origin（保持旧行为便于本地调试）
		upgrader.CheckOrigin = func(r *http.Request) bool {
			return true
		}
		return
	}

	if len(h.cfg.AllowedOrigins) == 0 {
		// 无白名单时只允许同源（浏览器自动带 Origin 作同源判断）
		upgrader.CheckOrigin = func(r *http.Request) bool {
			origin := r.Header.Get("Origin")
			// 无 Origin 头（非浏览器或同源）允许
			if origin == "" {
				return true
			}
			// 同源校验：Origin 与 Host 一致
			host := r.Host
			return origin == "http://"+host || origin == "https://"+host
		}
		return
	}

	// 白名单模式
	allowed := make(map[string]bool, len(h.cfg.AllowedOrigins))
	for _, o := range h.cfg.AllowedOrigins {
		allowed[o] = true
	}
	upgrader.CheckOrigin = func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		// 无 Origin 头的非浏览器客户端：白名单允许 "*"
		if origin == "" {
			return allowed["*"]
		}
		return allowed[origin]
	}
}

// Run 启动 Hub 的主事件循环。
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.userConns[client.UserID]++
			h.ipConns[client.remoteIP]++
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.Send)
				h.userConns[client.UserID]--
				if h.userConns[client.UserID] <= 0 {
					delete(h.userConns, client.UserID)
				}
				h.ipConns[client.remoteIP]--
				if h.ipConns[client.remoteIP] <= 0 {
					delete(h.ipConns, client.remoteIP)
				}
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
//
// 前置条件：调用方必须已通过 RequireAuth 中间件验证 JWT，
// 并将 user_id / workspace_id 注入请求上下文。
//
// 连接限流逻辑（按优先级）：
//  1. 全局硬上限（MaxConnsGlobal）
//  2. 单用户上限（MaxConnsPerUser）
//  3. 单 IP 上限（MaxConnsPerIP）
// 超出时返回 HTTP 503 拒绝升级，避免单用户/节点资源耗尽。
func (h *Hub) HandleWebSocket(w http.ResponseWriter, r *http.Request, userID int64, workspaceID int64) {
	// --- 连接限流 ---
	var remoteIP string
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		remoteIP = host
	} else {
		remoteIP = r.RemoteAddr
	}

	h.mu.RLock()
	currentGlobal := len(h.clients)
	currentUser := h.userConns[userID]
	currentIP := h.ipConns[remoteIP]
	h.mu.RUnlock()

	if h.cfg.MaxConnsGlobal > 0 && currentGlobal >= h.cfg.MaxConnsGlobal {
		http.Error(w, "Too many connections", http.StatusServiceUnavailable)
		return
	}
	if h.cfg.MaxConnsPerUser > 0 && currentUser >= h.cfg.MaxConnsPerUser {
		http.Error(w, "Too many connections for user", http.StatusTooManyRequests)
		return
	}
	if h.cfg.MaxConnsPerIP > 0 && currentIP >= h.cfg.MaxConnsPerIP {
		http.Error(w, "Too many connections from IP", http.StatusTooManyRequests)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("ws: upgrade error: %v", err)
		return
	}

	client := &Client{
		Conn:     conn,
		UserID:   userID,
		Send:     make(chan []byte, 64),
		hub:      h,
		remoteIP: remoteIP,
	}

	h.register <- client

	// 订阅 Redis 频道（跨节点扇出）
	pubsub := h.rdb.Subscribe(h.ctx, workspaceChannel(workspaceID))
	defer pubsub.Close()

	// 写入协程：从 Send 通道读，发给客户端
	go client.writePump()

	// 读取协程：心跳 + 关闭检测
	go client.readPump()

	// Redis 消息 → 客户端
	go func() {
		ch := pubsub.Channel()
		for msg := range ch {
			select {
			case client.Send <- []byte(msg.Payload):
			default:
				// 客户端消费不及时，丢弃（避免内存积压）
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

// Stats 返回当前连接统计（用于监控 / health 检查）。
func (h *Hub) Stats() (total int, users int, ips int) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients), len(h.userConns), len(h.ipConns)
}

func (h *Hub) unregisterClient(client *Client) {
	client.Conn.Close()
	select {
	case h.unregister <- client:
	default:
		// unregister 通道满（理论上不应发生），丢弃
	}
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
