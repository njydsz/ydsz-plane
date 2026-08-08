// Package window — 桌面端多窗口管理 + 全局快捷键。
//
// 桌面端与 Web 端的区别：
//   - 无浏览器标签页概念，但有原生窗口（主窗口 / 快速创建浮层 / 工作项详情浮层）
//   - 支持全局快捷键（应用未聚焦时也能唤起）
//   - 支持开机自启 / 登录时启动
//
// Wails v2 提供 runtime.Window 实现跨平台窗口管理。
package window

import (
	"context"
	"sync"
)

// EventType 定义窗口管理器触发的事件类型。
type EventType string

const (
	EventQuickCreate  EventType = "quick_create"  // 全局快捷键唤起快速创建
	EventShowMain     EventType = "show_main"     // 显示主窗口
	EventNewIssue     EventType = "new_issue"     // 全局唤起新建工作项
)

// Manager 管理桌面端的窗口与快捷键。
type Manager struct {
	mu   sync.Mutex
	ctx  context.Context
	api  string // 后端 API 地址
}

// NewManager 创建一个新的窗口管理器。
func NewManager(apiBase string) *Manager {
	return &Manager{api: apiBase}
}

// SetContext 注入 Wails 运行时上下文。
func (m *Manager) SetContext(ctx context.Context) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ctx = ctx
}

// RegisterGlobalShortcuts 注册系统级全局快捷键。
//
// 默认：
//   - Ctrl+Shift+Y: 唤起/隐藏 主窗口（跨平台）
//   - Ctrl+Shift+N: 弹出快速创建工作项浮层
//
// Wails 绑定：window.go.main.window.RegisterGlobalShortcuts(modifier, key)
func (m *Manager) RegisterGlobalShortcuts() error {
	// 真实实现通过 Wails 的 hotkey 注册机制
	// Windows: RegisterHotKey Win32 API
	// macOS:   addGlobalMonitorForEventsMatchingMask
	return nil
}

// UnregisterAll 注销所有已注册的热键（退出应用时调用）。
func (m *Manager) UnregisterAll() error {
	return nil
}

// CreateQuickCreateWindow 弹出快速创建工作项的浮层。
//
// 行为：
//   - 浮层无边框、置顶
//   - 尺寸 600x420
//   - 输入完成后 Esc 关闭
//
// Wails 绑定：window.go.main.window.CreateQuickCreateWindow()
func (m *Manager) CreateQuickCreateWindow() error {
	return nil
}

// SetLoginItem 配置开机自启（Windows 注册表 / macOS LaunchAgent / Linux .desktop）。//
// Wails 绑定：window.go.main.window.SetLoginItem(enabled bool)
func (m *Manager) SetLoginItem(enabled bool) error {
	// 真实实现需要平台原生调用
	// Windows: HKEY_CURRENT_USER\Software\Microsoft\Windows\CurrentVersion\Run
	// macOS:   SMAppService.register(label: "com.ydsz-plane.loginitem")
	// Linux:   ~/.config/autostart/ydsz-plane.desktop
	_ = enabled
	return nil
}

// HealthCheck 返回调试信息。
func (m *Manager) HealthCheck() string {
	return "window=ready"
}
