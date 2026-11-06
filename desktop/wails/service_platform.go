package main

import (
	"os"
	"runtime"
)

// PlatformService — 平台信息与服务入口。
//
// 暴露给前端的只读方法与基础能力。
type PlatformService struct{}

// GetPlatform 返回操作系统标识（windows / darwin / linux）。
func (s *PlatformService) GetPlatform() string { return runtime.GOOS }

// GetArch 返回 CPU 架构。
func (s *PlatformService) GetArch() string { return runtime.GOARCH }

// GetVersion 返回客户端版本。
func (s *PlatformService) GetVersion() string { return appVersion }

// IsDesktop 标识当前是否运行在桌面壳内。
func (s *PlatformService) IsDesktop() bool { return true }

// GetHostname 返回主机名。
func (s *PlatformService) GetHostname() string {
	h, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return h
}

// GetHomeDir 返回用户主目录（用于文件选择默认路径）。
func (s *PlatformService) GetHomeDir() string {
	h, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return h
}

// GetAppDataDir 返回应用数据目录。
func (s *PlatformService) GetAppDataDir() string {
	base, err := os.UserConfigDir()
	if err != nil {
		return ""
	}
	dir := base + string(os.PathSeparator) + "YdszPlane"
	_ = os.MkdirAll(dir, 0o755)
	return dir
}

// Log 前端日志转发到桌面端（便于统一排查）。
func (s *PlatformService) Log(level, message string) {
	logDesktop(level, message)
}
