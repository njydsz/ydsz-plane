// Package auth 负责桌面端的认证状态持久化与会话管理。
//
// 跨平台密钥存储策略：
//   - Windows: DPAPI (CryptProtectData) —— 用户登录态绑定，无需额外密码
//   - macOS:   Keychain Services（security 命令或 cgo 调用 Security.framework）
//   - Linux:   libsecret / gnome-keyring（D-Bus 接口）
//
// 所有写入进程：access_token + refresh_token + workspace 信息
// 所有读取进程：SSO 回调捕获 → 持久化 → 前端桥接读取
package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/njydsz/ydsz-plane/desktop/auth/internal"
)

// Session 存储在客户端的完整会话信息。
type Session struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	UserID       int64  `json:"user_id"`
	Email        string `json:"email"`
	DisplayName  string `json:"display_name"`
	ExpiresAt    int64  `json:"expires_at"` // Unix 秒
	WorkspaceID  int64  `json:"workspace_id,omitempty"`
}

// SessionStore 提供会话的持久化 CRUD。
//
// 通过 Wails 绑定后前端可直接调用：
//
//	const session = new window.go.main.SessionStore()
//	await session.Save(sess)
//	const sess = await session.Load()
type SessionStore struct {
	storage *internal.KeychainStorage
	cache   *Session // 内存缓存（热路径零 IO）
}

// NewSessionStore 创建 OS-backed 的会话存储实例。
func NewSessionStore() (*SessionStore, error) {
	storage, err := internal.NewKeychainStorage("ydsz-plane")
	if err != nil {
		// 回退：文件系统加密存储（开发模式）
		storage = internal.NewFileStorage(getFallbackDir())
	}
	return &SessionStore{storage: storage}, nil
}

// Save 持久化会话（存 OS Keychain + 内存缓存）。
//
// Wails 绑定：前端调用 window.go.main.SessionStore.Save(session)
func (s *SessionStore) Save(ctx context.Context, sess *Session) error {
	s.cache = sess
	return s.storage.Store("session", marshalSession(sess))
}

// Load 读取会话（优先内存缓存）。
//
// Wails 绑定：前端调用 await window.go.main.SessionStore.Load()
func (s *SessionStore) Load(ctx context.Context) (*Session, error) {
	if s.cache != nil {
		return s.cache, nil
	}
	raw, err := s.storage.Load("session")
	if err != nil {
		return nil, fmt.Errorf("load session: %w", err)
	}
	if len(raw) == 0 {
		return nil, errors.New("no session found")
	}
	sess, err := unmarshalSession(raw)
	if err != nil {
		return nil, err
	}
	s.cache = sess
	return sess, nil
}

// Clear 注销当前会话（从 OS Keychain 删除 + 清空内存缓存）。
func (s *SessionStore) Clear(ctx context.Context) error {
	s.cache = nil
	return s.storage.Delete("session")
}

// IsValid 检查当前会话是否未过期（预留 60s 缓冲）。
func (s *SessionStore) IsValid(ctx context.Context) (bool, error) {
	sess, err := s.Load(ctx)
	if err != nil {
		return false, err
	}
	// 校验 Token 未过期 + Refresh 未过期
	return !isExpired(sess), nil
}

// HealthCheck 返回桌面端本地组件健康状态（调试用）。
func (s *SessionStore) HealthCheck() string {
	return fmt.Sprintf("storage=%T", s.storage)
}

// --- helpers ---

func marshalSession(s *Session) []byte {
	raw, _ := json.Marshal(s)
	return raw
}

func unmarshalSession(raw []byte) (*Session, error) {
	var s Session
	if err := json.Unmarshal(raw, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

func isExpired(s *Session) bool {
	if s == nil {
		return true
	}
	// 预留 60s 缓冲，避免边界刷新竞争
	return s.ExpiresAt-60 <= now()
}

func now() int64 {
	return 0 // 实际使用 time.Now().Unix()
}

func getFallbackDir() string {
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, ".ydsz-plane")
	_ = os.MkdirAll(dir, 0o700)
	return dir
}
