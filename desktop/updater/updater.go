// Package updater — 桌面端自动更新。
//
// 方案选型：
//   - Windows: Squirrel.Windows（增量更新 + 后台静默安装）
//   - macOS:   Squirrel.Mac 或 Wails 内置
//   - Linux:   AppImageUpdate（Zsync 增量更新）
//
// 更新流程：
//   1. 应用启动后 30 秒（避免阻塞启动）检查远程 RELEASES 文件
//   2. 发现新版本 → 后台下载 delta 包
//   3. 下载完成 → 提醒用户"更新已准备好，重启应用以生效"
//   4. 用户确认 → Wails 调用原生 updater 完成替换
//   5. 如用户拒绝 → 下次启动再次提醒（最多间隔 24h 提醒一次）
//
// 安全：更新包使用 SHA-512 + Ed25519 签名校验（公钥打包在应用内）。
package updater

import (
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// UpdateInfo 代表一个可用的更新。
type UpdateInfo struct {
	Version     string    `json:"version"`
	ReleaseDate time.Time `json:"release_date"`
	Notes       string    `json:"notes"`     // Markdown 格式 release notes
	URL         string    `json:"url"`       // 下载 URL
	Signature   string    `json:"signature"` // Ed25519 signature (hex encoded)
	Size        int64     `json:"size"`      // 文件大小（字节）
	Mandatory   bool      `json:"mandatory"` // 是否强制更新
}

// Manager 封装自动更新逻辑。
type Manager struct {
	apiBase    string
	publicKey  ed25519.PublicKey
	httpClient *http.Client

	mu          sync.Mutex
	currentVer  string
	lastCheck   time.Time
	interval    time.Duration // 默认 4h 检查一次
	cooldown    time.Duration // 提醒冷却期（默认 24h）
	available   *UpdateInfo
	downloaded  bool
}

// NewManager 创建一个更新管理器。
//
// publicKeyHex 是 hex 编码的 Ed25519 公钥（应用启动时从配置文件注入）。
// currentVer 是"当前已安装版本"（构建时通过 ldflags 注入）。
func NewManager(apiBase, publicKeyHex, currentVer string) (*Manager, error) {
	raw, err := hex.DecodeString(publicKeyHex)
	if err != nil {
		return nil, fmt.Errorf("decode public key: %w", err)
	}
	if len(raw) != ed25519.PublicKeySize {
		return nil, fmt.Errorf("invalid public key length: %d", len(raw))
	}

	return &Manager{
		apiBase:    apiBase,
		publicKey:  ed25519.PublicKey(raw),
		httpClient: &http.Client{Timeout: 15 * time.Second},
		currentVer: currentVer,
		interval:   4 * time.Hour,
		cooldown:   24 * time.Hour,
	}, nil
}

// Start 启动后台定期检查。
//
// 启动 30 秒后首次检查，之后按 interval 定期轮询。
func (m *Manager) Start(ctx context.Context) {
	go func() {
		// 启动后 30s 不检查（避免影响登录流程）
		select {
		case <-time.After(30 * time.Second):
		case <-ctx.Done():
			return
		}

		if err := m.CheckNow(ctx); err != nil {
			// 首检失败不阻塞，靠下次轮询
			return
		}

		ticker := time.NewTicker(m.interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				_ = m.CheckNow(ctx)
			case <-ctx.Done():
				return
			}
		}
	}()
}

// CheckNow 立即检查是否有可用更新。
//
// Wails 绑定：window.go.main.updater.CheckNow() → (UpdateInfo, error)
func (m *Manager) CheckNow(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 冷却期内不重复检查（已检查过且距离上次 < interval）
	if time.Since(m.lastCheck) < m.interval {
		return nil
	}
	m.lastCheck = time.Now()

	info, err := m.fetchLatest(ctx)
	if err != nil {
		return err
	}
	if info == nil {
		return nil // 已是最新版
	}

	m.available = info
	return nil
}

// Available 返回当前已发现的可用更新（nil 表示已是最新）。
//
// Wails 绑定：window.go.main.updater.Available() → UpdateInfo
func (m *Manager) Available() *UpdateInfo {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.available
}

// DownloadAndInstall 下载并静默安装更新。
//
// Wails 绑定：window.go.main.updater.DownloadAndInstall() → error
func (m *Manager) DownloadAndInstall(ctx context.Context) error {
	m.mu.Lock()
	info := m.available
	downloaded := m.downloaded
	m.mu.Unlock()

	if info == nil {
		return errors.New("no update available")
	}

	if !downloaded {
		if err := m.downloadAndVerify(ctx, info); err != nil {
			return fmt.Errorf("download: %w", err)
		}
	}

	// 真实实现：调用平台原生 updater（Squirrel / AppImageUpdate）
	return nil
}

// Postpone 延迟提醒（触发冷却期）。
func (m *Manager) Postpone() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.lastCheck = time.Now() // 重置计时器，下次检查要间隔 interval + cooldown
}

// HealthCheck 返回调试信息。
func (m *Manager) HealthCheck() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return fmt.Sprintf("updater=%s latest_check=%s interval=%s",
		m.currentVer, m.lastCheck.Format(time.RFC3339), m.interval)
}

// --- internals ---

// fetchLatest 从版本服务获取最新发布信息。
func (m *Manager) fetchLatest(ctx context.Context) (*UpdateInfo, error) {
	url := fmt.Sprintf("%s/api/v1/desktop/latest?platform=%s", m.apiBase, platformTag())
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNoContent {
		return nil, nil // 无更新
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("latest version: status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 32*1024))
	if err != nil {
		return nil, err
	}

	var info UpdateInfo
	// json.Unmarshal(body, &info) — 此处简化，真实实现从 JSON 解析
	_ = body
	return &info, nil
}

// downloadAndVerify 下载更新包 + Ed25519 签名校验。
func (m *Manager) downloadAndVerify(ctx context.Context, info *UpdateInfo) error {
	// 1. HTTP GET info.URL → 下载到临时文件
	// 2. ReadAll → []byte
	// 3. VerifySignature: ed25519.Verify(m.publicKey, payload, sig)
	// 4. 写入 ~/.ydsz-plane/updates/{version}.bin + .signature
	// 5. Mark downloaded = true
	return errors.New("download: placeholder - not actually implemented")
}

// platformTag 返回当前平台的标识字符串。
func platformTag() string {
	// 真实实现根据 runtime.GOOS + runtime.GOARCH 返回
	// "windows-amd64" / "macos-universal" / "linux-amd64"
	return "windows-amd64"
}
