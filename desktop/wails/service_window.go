package main

import (
	"github.com/wailsapp/wails/v3/pkg/application"
)

// WindowService — 窗口控制。
type WindowService struct {
	app *application.App
}

// SetTitle 设置窗口标题。
func (s *WindowService) SetTitle(title string) {
	if w := s.mainWindow(); w != nil {
		w.SetTitle(title)
	}
}

// Minimize 最小化窗口。
func (s *WindowService) Minimize() {
	if w := s.mainWindow(); w != nil {
		w.Minimise()
	}
}

// Maximize 最大化/还原窗口。
func (s *WindowService) Maximize() {
	if w := s.mainWindow(); w != nil {
		w.ToggleMaximise()
	}
}

// Close 关闭窗口（隐藏到托盘）。
func (s *WindowService) Close() {
	if w := s.mainWindow(); w != nil {
		w.Hide()
	}
}

// Quit 退出应用。
func (s *WindowService) Quit() {
	if s.app != nil {
		s.app.Quit()
	}
}

// SetAlwaysOnTop 置顶窗口。
func (s *WindowService) SetAlwaysOnTop(on bool) {
	if w := s.mainWindow(); w != nil {
		w.SetAlwaysOnTop(on)
	}
}

// IsMaximized 返回是否最大化。
func (s *WindowService) IsMaximized() bool {
	if w := s.mainWindow(); w != nil {
		return w.IsMaximised()
	}
	return false
}

// OpenURL 用系统默认浏览器打开链接。
func (s *WindowService) OpenURL(url string) {
	if s.app != nil {
		_ = s.app.Browser.OpenURL(url)
	}
}

// ShowAndFocus 显示并聚焦窗口。
func (s *WindowService) ShowAndFocus() {
	if w := s.mainWindow(); w != nil {
		w.Show()
		w.Focus()
	}
}

// mainWindow 获取主窗口。
func (s *WindowService) mainWindow() application.Window {
	if s.app == nil {
		return nil
	}
	w, ok := s.app.Window.GetByName("main")
	if !ok || w == nil {
		return nil
	}
	return w
}
