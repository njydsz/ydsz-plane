// Package auth — 桌面端 SSO/OAuth 回调捕获。
//
// 桌面端无法像 Web 端直接 302 跳转后接收回调，采用
// "localhost 临时 HTTP Server" 方案：
//
//   1. 桌面端启动临时 HTTP Server 监听 127.0.0.1:{随机端口}
//   2. 调用系统浏览器打开 SSO 授权 URL（redirect_uri = http://127.0.0.1:port/callback）
//   3. 用户在浏览器完成登录后，IdP 回调 → 桌面端临时 Server 接收 code + state
//   4. 桌面端用 code 换取 Token → 持久化到 OS Keychain
//   5. 显示"登录成功"原生通知 → 关闭临时 Server
//
// 安全约束：
//   - 临时 Server 仅绑定 127.0.0.1（不暴露到局域网）
//   - 单次监听，收到回调或 5 超时后立即关闭
//   - state 参数防 CSRF（auth 包内 crypto/rand 生成）
package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"
)

const (
	// ssoCallbackTimeout 临时 SSO 回调 Server 的最大等待时间。
	ssoCallbackTimeout = 5 * time.Minute

	// ssoCallbackHost 仅监听本地回环地址。
	ssoCallbackHost = "127.0.0.1"
)

// SSOManager 管理桌面端 SSO 认证流程。
type SSOManager struct {
	mu       sync.Mutex
	state    string
	issuer   string
	clientID string
	redirect string // 回调 URL（含实际端口）
	ch       chan *SSOCallback
	srv      *http.Server
}

// SSOCallback 是回调接收到的 code + state。
type SSOCallback struct {
	Code  string
	State string
	Error string
}

// NewSSOManager 创建 SSO 流程管理器。
func NewSSOManager(issuer, clientID string) *SSOManager {
	return &SSOManager{
		issuer:   issuer,
		clientID: clientID,
		ch:       make(chan *SSOCallback, 1),
	}
}

// Start 启动临时 HTTP Server 监听 SSO 回调。
//
// 返回实际监听的完整回调 URL（前端/系统浏览器使用）。
func (s *SSOManager) Start(ctx context.Context) (callbackURL string, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	port, err := pickFreePort()
	if err != nil {
		return "", fmt.Errorf("pick free port: %w", err)
	}

	s.state = generateState()
	s.redirect = fmt.Sprintf("http://%s:%d/callback", ssoCallbackHost, port)

	mux := http.NewServeMux()
	mux.HandleFunc("/callback", s.handleCallback)
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("ok"))
	})

	s.srv = &http.Server{
		Addr:              fmt.Sprintf("%s:%d", ssoCallbackHost, port),
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		BaseContext:       func(_ net.Listener) context.Context { return ctx },
	}

	// 超时自动关闭
	go func() {
		select {
		case <-time.After(ssoCallbackTimeout):
			_ = s.Stop()
		case <-ctx.Done():
			_ = s.Stop()
		}
	}()

	go func() {
		if err := s.srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.ch <- &SSOCallback{Error: err.Error()}
		}
	}()

	return s.redirect, nil
}

// Wait 阻塞等待回调到达或超时。
func (s *SSOManager) Wait(ctx context.Context) (*SSOCallback, error) {
	select {
	case cb := <-s.ch:
		if cb.Error != "" {
			return nil, fmt.Errorf("sso callback error: %s", cb.Error)
		}
		if cb.State != s.state {
			return nil, errors.New("sso: state mismatch (CSRF protection)")
		}
		return cb, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// Stop 关闭临时 HTTP Server。
func (s *SSOManager) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.srv == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err := s.srv.Shutdown(ctx)
	s.srv = nil
	return err
}

// --- internals ---

func (s *SSOManager) handleCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	errStr := r.URL.Query().Get("error")

	s.ch <- &SSOCallback{
		Code:  code,
		State: state,
		Error: errStr,
	}

	// 返回一个"可关闭"的 HTML 页面（用户可手动关闭标签页）
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = fmt.Fprint(w, ssoSuccessHTML)
}

func generateState() string {
	buf := make([]byte, 32)
	_, _ = rand.Read(buf)
	return hex.EncodeToString(buf)
}

func pickFreePort() (int, error) {
	l, err := net.Listen("tcp", ssoCallbackHost+":0")
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}

// ssoSuccessHTML 是浏览器回调成功后显示的页面。
var ssoSuccessHTML = `<!DOCTYPE html>
<html lang="zh-CN">
<head><meta charset="utf-8"><title>Ydsz Plane</title></head>
<body style="font-family:system-ui;margin:0 auto;text-align:center;padding:60px 20px;background:#fff">
  <h1>🎉 登录成功</h1>
  <p>请返回 <b>Ydsz Plane</b> 桌面应用继续使用。</p>
  <p style="color:#999;font-size:13px">此标签页可以安全关闭</p>
</body></html>`
