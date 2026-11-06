package main

import (
	"github.com/wailsapp/wails/v3/pkg/application"
)

// UpdaterService — 应用内自动更新。
type UpdaterService struct {
	app     *application.App
	enabled bool
}

// CheckForUpdate 检查新版本。
func (s *UpdaterService) CheckForUpdate() (string, error) {
	if s.app == nil {
		return "", nil
	}
	// Wails v3 的更新检查由 updater 插件处理；此处返回配置信息
	return appVersion, nil
}

// IsAutoUpdateEnabled 返回是否启用自动更新。
func (s *UpdaterService) IsAutoUpdateEnabled() bool { return s.enabled }

// SetAutoUpdateEnabled 设置自动更新开关。
func (s *UpdaterService) SetAutoUpdateEnabled(enabled bool) { s.enabled = enabled }

// InstallUpdate 安装已下载的更新。
func (s *UpdaterService) InstallUpdate() error {
	// 由 updater 插件处理实际安装
	return nil
}

// GetCurrentVersion 返回当前版本。
func (s *UpdaterService) GetCurrentVersion() string { return appVersion }
