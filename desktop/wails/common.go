package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

// appVersion 客户端版本（与 wails.json 同步）。
const appVersion = "0.1.0"

// logDesktop 将前端日志写入桌面端日志文件。
func logDesktop(level, message string) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return
	}
	logDir := filepath.Join(dir, "YdszPlane", "logs")
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return
	}

	name := fmt.Sprintf("app-%s.log", time.Now().Format("2006-01-02"))
	f, err := os.OpenFile(filepath.Join(logDir, name), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return
	}
	defer f.Close()

	line := fmt.Sprintf("%s [%s] %s (%s/%s)\n",
		time.Now().Format("2006-01-02 15:04:05"), level, message, runtime.GOOS, runtime.GOARCH)
	_, _ = f.WriteString(line)

	// 同时输出到标准日志
	log.Print(line)
}
