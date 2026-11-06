// Package crdt — CRDT 文档协同（Y.js over WebSocket）。
//
// 协议：兼容 y-websocket 的二进制帧格式。
//
//	messageSync = 0      → [subtype, payload]
//	  subtype 0 (sync step1)：客户端发送 state vector
//	  subtype 1 (sync step2)：服务端响应差异更新
//	  subtype 2 (update)    ：增量更新（需广播给其他客户端）
//	messageAwareness = 1 → awareness 更新（广播）
//
// 设计：
//   - 每个文档（docID）一个房间（Room）；
//   - 房间维护文档状态的服务端镜像（用于新客户端同步与持久化）；
//   - 客户端连接 → 加入房间 → 双向同步；
//   - 消息广播：房间内其他连接（跨节点通过 Redis 预留）。
package crdt

import (
	"context"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Config CRDT Hub 配置。
type Config struct {
	MaxConnsPerUser int
	MaxConnsPerIP   int
	// 文档持久化（可选；nil 表示仅内存）
	Store DocumentStore
	// 文档最大字节数（防御）
	MaxDocBytes int
}

// DocumentStore 文档持久化接口。
type DocumentStore interface {
	Load(ctx context.Context, docID string) ([]byte, error)
	Save(ctx context.Context, docID string, state []byte) error
}

// Hub CRDT 协同 Hub。
type Hub struct {
	cfg      Config
	upgrader websocket.Upgrader
	mu       sync.RWMutex
	rooms    map[string]*Room
	// broadcaster 可选的 Redis 跨节点广播（nil = 单节点模式）
	broadcaster *RedisBroadcaster
}

// NewHub 构造 CRDT Hub。
func NewHub(cfg Config) *Hub {
	if cfg.MaxDocBytes <= 0 {
		cfg.MaxDocBytes = 16 << 20 // 16MB
	}
	return &Hub{
		cfg: cfg,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  4096,
			WriteBufferSize: 4096,
			CheckOrigin:     func(r *http.Request) bool { return true },
		},
		rooms: map[string]*Room{},
	}
}

// HandleWebSocket 升级并处理 CRDT 连接。
//
// 路由约定：/ws-collab/:workspace_id/:doc_id
func (h *Hub) HandleWebSocket(w http.ResponseWriter, r *http.Request, userID, workspaceID int64, docID string) {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("crdt: upgrade error: %v", err)
		return
	}

	room := h.getOrCreateRoom(docID)
	client := &Client{
		Conn:     conn,
		UserID:   userID,
		Send:     make(chan []byte, 64),
		room:     room,
		remoteIP: remoteIP(r),
	}
	room.join <- client

	go client.writePump()
	go client.readPump()
}

// WithRedisBroadcaster 注入 Redis 广播器（启用跨节点协同）。
//
// 启用后：
//   - 本地更新会发布到 Redis 频道，其他节点转发给各自客户端；
//   - 新客户端加入时向其他节点请求状态；
//   - 状态同时写入 Redis 作为权威副本。
func (h *Hub) WithRedisBroadcaster(b *RedisBroadcaster) {
	h.broadcaster = b
	b.SetCallbacks(
		func(docID string, data []byte) { h.onRemoteUpdate(docID, data) },
		func(docID string, data []byte) { h.onRemoteAwareness(docID, data) },
		func(docID string) ([]byte, bool) { return h.onSyncRequest(docID) },
	)
}

// onRemoteUpdate 处理远端增量更新：转发给本地房间 + 合并状态。
func (h *Hub) onRemoteUpdate(docID string, data []byte) {
	h.mu.RLock()
	room, ok := h.rooms[docID]
	h.mu.RUnlock()
	if !ok {
		return
	}
	// 转发给本地客户端
	room.mu.RLock()
	for c := range room.clients {
		select {
		case c.Send <- data:
		default:
		}
	}
	room.mu.RUnlock()
	// 合并到房间状态
	room.mu.Lock()
	if len(room.state) == 0 {
		room.state = append([]byte(nil), data...)
	} else {
		room.state = append(room.state, data...)
	}
	room.mu.Unlock()
}

// onRemoteAwareness 处理远端 awareness：转发给本地房间。
func (h *Hub) onRemoteAwareness(docID string, data []byte) {
	h.mu.RLock()
	room, ok := h.rooms[docID]
	h.mu.RUnlock()
	if !ok {
		return
	}
	room.mu.RLock()
	for c := range room.clients {
		select {
		case c.Send <- data:
		default:
		}
	}
	room.mu.RUnlock()
}

// onSyncRequest 响应远端同步请求：返回本地状态。
func (h *Hub) onSyncRequest(docID string) ([]byte, bool) {
	h.mu.RLock()
	room, ok := h.rooms[docID]
	h.mu.RUnlock()
	if !ok {
		return nil, false
	}
	room.mu.RLock()
	defer room.mu.RUnlock()
	if len(room.state) == 0 {
		return nil, false
	}
	return room.state, true
}

// Stats 返回统计。
func (h *Hub) Stats() (rooms, conns int) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.rooms), 0
}

func (h *Hub) getOrCreateRoom(docID string) *Room {
	h.mu.Lock()
	defer h.mu.Unlock()
	if r, ok := h.rooms[docID]; ok {
		return r
	}
	r := newRoom(docID, h.cfg, h)
	h.rooms[docID] = r
	go r.run()
	return r
}

func remoteIP(r *http.Request) string {
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return host
	}
	return r.RemoteAddr
}

// Room 文档房间。
type Room struct {
	docID string
	cfg   Config
	hub   *Hub

	join      chan *Client
	leave     chan *Client
	broadcast chan broadcastMsg

	mu      sync.RWMutex
	clients map[*Client]struct{}
	// 服务端文档状态（Y.js update 顺序拼接）
	state []byte
	// 是否有活跃客户端
	active bool
}

type broadcastMsg struct {
	data   []byte
	origin *Client
}

func newRoom(docID string, cfg Config, hub *Hub) *Room {
	return &Room{
		docID:     docID,
		cfg:       cfg,
		hub:       hub,
		join:      make(chan *Client, 16),
		leave:     make(chan *Client, 16),
		broadcast: make(chan broadcastMsg, 256),
		clients:   map[*Client]struct{}{},
	}
}

// run 房间主循环（单 goroutine）。
func (r *Room) run() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case c := <-r.join:
			r.mu.Lock()
			r.clients[c] = struct{}{}
			r.active = true
			state := r.state
			r.mu.Unlock()
			if len(state) > 0 {
				c.sendSyncStep2(state)
			} else if r.hub != nil && r.hub.broadcaster != nil {
				// 本地无状态：尝试从 Redis 加载 + 请求其他节点
				go r.hub.broadcaster.RequestSync(r.docID)
				if loaded, err := r.hub.broadcaster.LoadState(r.docID); err == nil && len(loaded) > 0 {
					c.sendSyncStep2(loaded)
				}
			}

		case c := <-r.leave:
			r.mu.Lock()
			delete(r.clients, c)
			empty := len(r.clients) == 0
			r.active = !empty
			r.mu.Unlock()
			close(c.Send)
			if empty && r.cfg.Store != nil && len(r.state) > 0 {
				_ = r.cfg.Store.Save(context.Background(), r.docID, r.state)
			}

		case msg := <-r.broadcast:
			r.broadcastExcept(msg.data, msg.origin)

		case <-ticker.C:
			if r.cfg.Store != nil {
				r.mu.RLock()
				state := r.state
				active := r.active
				r.mu.RUnlock()
				if active && len(state) > 0 {
					_ = r.cfg.Store.Save(context.Background(), r.docID, state)
				}
			}
		}
	}
}

// broadcastToRoom 广播消息到房间（由 Client 调用）。
func (r *Room) broadcastToRoom(data []byte, origin *Client) {
	select {
	case r.broadcast <- broadcastMsg{data: data, origin: origin}:
	default:
	}
}

func (r *Room) broadcastExcept(data []byte, origin *Client) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for c := range r.clients {
		if c == origin {
			continue
		}
		select {
		case c.Send <- data:
		default:
			// 客户端消费不及时，丢弃
		}
	}
}

// Client CRDT 客户端连接。
type Client struct {
	Conn     *websocket.Conn
	UserID   int64
	Send     chan []byte
	room     *Room
	remoteIP string
}

// writePump 写协程。
func (c *Client) writePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()
	for {
		select {
		case msg, ok := <-c.Send:
			if !ok {
				_ = c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.BinaryMessage, msg); err != nil {
				return
			}
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// readPump 读协程：解析 y-websocket 帧并分发。
func (c *Client) readPump() {
	defer func() {
		c.room.leave <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(int64(c.room.cfg.MaxDocBytes))
	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, data, err := c.Conn.ReadMessage()
		if err != nil {
			break
		}
		c.handleFrame(data)
	}
}

// handleFrame 处理 y-websocket 二进制帧。
func (c *Client) handleFrame(data []byte) {
	if len(data) == 0 {
		return
	}
	messageType := data[0]
	payload := data[1:]

	switch messageType {
	case 0: // messageSync
		if len(payload) == 0 {
			return
		}
		subtype := payload[0]
		body := payload[1:]
		switch subtype {
		case 0: // sync step1：客户端请求同步 → 服务端回复差异
			c.sendSyncFromStateVector(body)
		case 1: // sync step2：服务端下发（客户端一般不接收）
			c.applyUpdate(body)
		case 2: // update：增量 → 广播 + 合并到服务端状态
			c.applyUpdate(body)
			c.room.broadcastToRoom(data, c)
		}
	case 1: // messageAwareness：广播
		c.room.broadcastToRoom(data, c)
	}
}

// sendSyncFromStateVector 根据客户端 state vector 回复差异更新。
//
// 策略：服务端不解析 Y.js 结构，直接下发累积状态；客户端用
// Y.encodeStateAsUpdate(doc, remoteStateVector) 自行合并。
func (c *Client) sendSyncFromStateVector(_ []byte) {
	c.room.mu.RLock()
	state := c.room.state
	c.room.mu.RUnlock()
	c.sendSyncStep2(state)
}

// sendSyncStep2 发送 sync step2 帧。
func (c *Client) sendSyncStep2(state []byte) {
	frame := make([]byte, 2+len(state))
	frame[0] = 0 // messageSync
	frame[1] = 1 // syncStep2
	copy(frame[2:], state)
	c.Send <- frame
}

// applyUpdate 将增量更新合并到服务端状态（顺序拼接）。
func (c *Client) applyUpdate(update []byte) {
	if len(update) == 0 {
		return
	}
	c.room.mu.Lock()
	if len(c.room.state) == 0 {
		c.room.state = append([]byte(nil), update...)
	} else {
		c.room.state = append(c.room.state, update...)
	}
	c.room.mu.Unlock()

	// 跨节点广播
	if c.room.hub.broadcaster != nil {
		go c.room.hub.broadcaster.PublishUpdate(c.room.docID, update)
		go c.room.hub.broadcaster.SaveState(c.room.docID, c.room.state)
	}
}

// NewHubWithStore 创建带持久化的 Hub。
func NewHubWithStore(store DocumentStore, maxPerUser, maxPerIP int) *Hub {
	return NewHub(Config{
		MaxConnsPerUser: maxPerUser,
		MaxConnsPerIP:   maxPerIP,
		Store:           store,
	})
}

// --- 内存存储（默认） ---

// MemoryStore 内存文档存储。
type MemoryStore struct {
	mu   sync.RWMutex
	docs map[string][]byte
}

// NewMemoryStore 构造内存存储。
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{docs: map[string][]byte{}}
}

// Load 读取文档。
func (s *MemoryStore) Load(_ context.Context, docID string) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.docs[docID], nil
}

// Save 保存文档。
func (s *MemoryStore) Save(_ context.Context, docID string, state []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.docs[docID] = append([]byte(nil), state...)
	return nil
}

var _ = strings.TrimSpace
