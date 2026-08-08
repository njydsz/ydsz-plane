// Package auth — Token 自动刷新器。
//
// 桌面端的 JWT Token 通常有效期较短（15min~2h），需要 transparent refresh。
// 本模块在后台 Goroutine 定期检查 access_token 有效期，
// 在到期前自动用 refresh_token 换取新 Token。
//
// 与 Web 端的差异：
//   - Web: refresh_token 在 httpOnly Cookie 浏览器自动管理
//   - Desktop: token 存 OS Keychain，需主动发起 refresh 请求
package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// TokenRefresher 在后台静默刷新 JWT Token。
type TokenRefresher struct {
	mu      sync.Mutex
	store   *SessionStore
	apiBase string // e.g. "https://njydsz.com"
	client  *http.Client
	stopCh  chan struct{}
}

// NewTokenRefresher 创建一个 Token 刷新器。
//
// apiBase 为后端 REST API 基础 URL（如 https://njydsz.com）。
func NewTokenRefresher(store *SessionStore, apiBase string) *TokenRefresher {
	return &TokenRefresher{
		store:   store,
		apiBase: apiBase,
		client:  &http.Client{Timeout: 10 * time.Second},
		stopCh:  make(chan struct{}),
	}
}

// Start 启动后台刷新循环。
// 每 60s 检查一次 Token 是否在 refreshBuffer 内过期。
//
// refreshBuffer = Token 过期前多少秒触发 refresh（默认 5min）。
func (r *TokenRefresher) Start(ctx context.Context, refreshBuffer time.Duration) {
	if refreshBuffer == 0 {
		refreshBuffer = 5 * time.Minute
	}
	go r.loop(ctx, refreshBuffer)
}

// Stop 停止刷新循环。
func (r *TokenRefresher) Stop() {
	close(r.stopCh)
}

func (r *TokenRefresher) loop(ctx context.Context, buffer time.Duration) {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			r.maybeRefresh(ctx, buffer)
		case <-r.stopCh:
			return
		}
	}
}

// maybeRefresh 检查 Token 如即将过期则刷新。
func (r *TokenRefresher) maybeRefresh(ctx context.Context, buffer time.Duration) {
	r.mu.Lock()
	defer r.mu.Unlock()

	sess, err := r.store.Load(ctx)
	if err != nil {
		return // 无 session，跳过
	}

	// 尚未过期且在缓冲窗口外 → 不刷新
	if sess.ExpiresAt > time.Now().Add(buffer).Unix() {
		return
	}

	newSess, err := r.doRefresh(ctx, sess.RefreshToken)
	if err != nil {
		return // 刷新失败 → 下次重试
	}

	_ = r.store.Save(ctx, newSess)
}

// doRefresh 向后端发起 Token Refresh 请求。
//
// 后端 POST /api/v1/auth/refresh body: { refresh_token: "..." }
// 返回新 access_token + refresh_token 及过期时间。
func (r *TokenRefresher) doRefresh(ctx context.Context, refreshToken string) (*Session, error) {
	payload := map[string]string{"refresh_token": refreshToken}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		r.apiBase+"/api/v1/auth/refresh", bytesReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("refresh request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("refresh status %d", resp.StatusCode)
	}

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 16*1024))
	if err != nil {
		return nil, fmt.Errorf("read refresh resp: %w", err)
	}

	var result struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresAt    int64  `json:"expires_at"`
		UserID       int64  `json:"user_id"`
		Email        string `json:"email"`
		DisplayName  string `json:"display_name"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("parse refresh resp: %w", err)
	}

	return &Session{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		UserID:       result.UserID,
		Email:        result.Email,
		DisplayName:  result.DisplayName,
		ExpiresAt:    result.ExpiresAt,
	}, nil
}

// bytesReader 将 [] 包装为 io.Reader。
func bytesReader(b []byte) io.Reader {
	return &bytesReaderImpl{b: b}
}

type bytesReaderImpl struct {
	b   []byte
	pos int
}

func (r *bytesReaderImpl) Read(p []byte) (int, error) {
	if r.pos >= len(r.b) {
		return 0, io.EOF
	}
	n := copy(p, r.b[r.pos:])
	r.pos += n
	return n, nil
}
