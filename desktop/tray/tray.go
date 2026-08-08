// Package tray — 系统托盘（Windows 托盘图标 / macOS 状态栏 / Linux AppIndicator）。
//
// 通过 Wails Runtime.Window 实现原生托盘能力：
//   - Windows: Wails 使用 Shell_NotifyIcon Win32 API
//   - macOS:   Wails 使用 NSStatusItem
//   - Linux:   Wails 需 libappindicator3（Ubuntu 自带）
//
// 行为：
//   - 应用启动：图标进入托盘
//   - 关闭窗口：最小化到托盘（不退出）
//   - 单击托盘：恢复主窗口
//   - 未读通知：托盘图标叠加红色角标
//   - 托盘菜单：工作项 / 通知 / 退出
package tray

import (
	"context"
	"sync"
)

// 托盘按钮 ID 常量。
const (
	MenuItemShowApp      = "show_app"
	MenuItemQuickCreate  = "quick_create"
	MenuItemNotification = "notification"
	MenuItemPreference   = "preference"
	MenuItemQuit         = "quit"
)

// Manager 封装托盘的生命周期。
type Manager struct {
	mu          sync.Mutex
	ctx         context.Context
	unreadCount int
	apiBase     string
}

// NewManager 创建一个未绑定的 Manager。
func NewManager(apiBase string) *Manager {
	return &Manager{
		apiBase: apiBase,
	}
}

// SetContext 注入 Wails 运行时上下文。
func (m *Manager) SetContext(ctx context.Context) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ctx = ctx
}

// Install 安装系统托盘图标 + 菜单。
//
// Wails 绑定：window.go.main.tray.Install()
func (m *Manager) Install() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	// 实际实现通过 Wails Runtime 调用平台原生 API
	// placeholder: 真实 Wails v2 使用 runtime.Window.SetBackgroundColour + systray 等机制
	return nil
}

// UpdateUnreadCount 更新未读计数角标。
//
// Wails 绑定：window.go.main.tray.UpdateUnreadCount(count)
func (m *Manager) UpdateUnreadCount(count int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.unreadCount = count
	// 实际实现：通过 Wails Runtime 更新托盘图标叠加层
	// Windows: 修改托盘图标为带数字的 ICO 文件
	// macOS:   Dock badge label
}

// Notify 发送一条原生系统通知。
//
// 跨平台：
//   - Windows: ToastNotification (Windows.Data.Xml.Dom)
//   - macOS:   UNUserNotificationCenter
//   - Linux:   notify-send (libnotify)
//
// Wails 绑定：window.go.main.tray.Notify(title, body)
func (m *Manager) Notify(title, body string, urgent bool) error {
	// placeholder: 真实实现依赖平台调用
	_ = urgent
	return nil
}

// Quit 彻底退出应用（非最小化到托盘）。
func (m *Manager) Quit() {
	m.mu.Lock()
	defer m.mu.Unlock()
	// 真实实现：Wails Runtime.Quit() 触发进程退出
}
