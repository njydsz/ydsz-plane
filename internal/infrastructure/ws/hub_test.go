// Package ws — WebSocket Hub 单元测试。
//
// 不使用真实 WebSocket 连接或 Redis，仅测试：
//   - NewHub / NewHubWithConfig 构造函数
//   - Stats() 返回初始零值
//   - Message JSON 序列化/反序列化
//   - workspaceChannel 频道名格式
package ws

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

// TestNewHub_CreatesHub 验证 NewHub 创建有效 Hub 实例。
// 不需要 Redis 连接（Hub 仅持有引用，不主动 Ping）。
func TestNewHub_CreatesHub(t *testing.T) {
	h := NewHub(nil)
	if h == nil {
		t.Fatal("NewHub returned nil")
	}

	// 各 channel 和 map 应已初始化
	if h.clients == nil {
		t.Fatal("clients map not initialized")
	}
	if h.broadcast == nil {
		t.Fatal("broadcast channel not initialized")
	}
	if h.register == nil {
		t.Fatal("register channel not initialized")
	}
	if h.unregister == nil {
		t.Fatal("unregister channel not initialized")
	}
	if h.userConns == nil {
		t.Fatal("userConns map not initialized")
	}
	if h.ipConns == nil {
		t.Fatal("ipConns map not initialized")
	}
}

// TestNewHubWithConfig_EmptyConfig 验证空配置（默认安全行为）下 Hub 字段正确初始化。
func TestNewHubWithConfig_EmptyConfig(t *testing.T) {
	h := NewHubWithConfig(nil, HubConfig{})
	if h == nil {
		t.Fatal("NewHubWithConfig returned nil")
	}
	if h.ctx == nil {
		t.Fatal("context not initialized")
	}
	if h.cancel == nil {
		t.Fatal("cancel function not initialized")
	}

	// Stats 应返回零值
	total, users, ips := h.Stats()
	if total != 0 || users != 0 || ips != 0 {
		t.Fatalf("Stats() = (%d, %d, %d), want (0, 0, 0)", total, users, ips)
	}
}

// TestNewHubWithConfig_CustomConfig 验证自定义配置正确传入。
func TestNewHubWithConfig_CustomConfig(t *testing.T) {
	reg := prometheus.NewRegistry()
	cfg := HubConfig{
		AllowedOrigins:     []string{"https://app.example.com", "https://admin.example.com"},
		MaxConnsPerUser:    5,
		MaxConnsPerIP:      10,
		MaxConnsGlobal:     1000,
		RequireOriginCheck: true,
		Registerer:         reg,
	}

	h := NewHubWithConfig(nil, cfg)
	if h == nil {
		t.Fatal("NewHubWithConfig returned nil")
	}

	// 验证配置被复制
	if h.cfg.MaxConnsPerUser != 5 {
		t.Fatalf("MaxConnsPerUser = %d, want 5", h.cfg.MaxConnsPerUser)
	}
	if h.cfg.MaxConnsGlobal != 1000 {
		t.Fatalf("MaxConnsGlobal = %d, want 1000", h.cfg.MaxConnsGlobal)
	}
	if !h.cfg.RequireOriginCheck {
		t.Fatal("RequireOriginCheck should be true")
	}
	if len(h.cfg.AllowedOrigins) != 2 {
		t.Fatalf("AllowedOrigins: got %d entries, want 2", len(h.cfg.AllowedOrigins))
	}
}

// TestMessage_JSON 验证 Message 结构体 JSON 往返。
func TestMessage_JSON(t *testing.T) {
	tests := []struct {
		name    string
		message Message
	}{
		{
			name:    "issue updated",
			message: Message{Type: "issue.updated", Data: json.RawMessage(`{"id":123,"state":"done"}`)},
		},
		{
			name:    "ping",
			message: Message{Type: "ping", Data: nil},
		},
		{
			name:    "notification",
			message: Message{Type: "notification.new", Data: json.RawMessage(`{"text":"hello"}`)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.message)
			if err != nil {
				t.Fatalf("Marshal: %v", err)
			}

			var decoded Message
			if err := json.Unmarshal(data, &decoded); err != nil {
				t.Fatalf("Unmarshal: %v", err)
			}

			if decoded.Type != tt.message.Type {
				t.Fatalf("Type: got %q, want %q", decoded.Type, tt.message.Type)
			}
			if tt.message.Data != nil {
				if string(decoded.Data) != string(tt.message.Data) {
					t.Fatalf("Data: got %q, want %q", decoded.Data, tt.message.Data)
				}
			}
		})
	}
}

// TestWorkspaceChannel 验证频道名格式遵循命名约定。
func TestWorkspaceChannel(t *testing.T) {
	tests := []struct {
		workspaceID int64
		wantPrefix  string
	}{
		{1, "plane:ws:1"},
		{100, "plane:ws:100"},
		{999999, "plane:ws:999999"},
		{0, "plane:ws:0"},
	}

	for _, tt := range tests {
		t.Run(tt.wantPrefix, func(t *testing.T) {
			got := workspaceChannel(tt.workspaceID)
			if got != tt.wantPrefix {
				t.Fatalf("workspaceChannel(%d) = %q, want %q",
					tt.workspaceID, got, tt.wantPrefix)
			}
			// 格式断言：以 plane:ws: 开头
			if !strings.HasPrefix(got, "plane:ws:") {
				t.Fatalf("workspaceChannel(%d) = %q, missing prefix",
					tt.workspaceID, got)
			}
		})
	}
}
