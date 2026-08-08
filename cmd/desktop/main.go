// Command desktop 启动 Ydsz Plane 原生桌面客户端（Wails v2 绑定）。
//
// 本入口将现有后端能力通过 Wails Runtime 暴露给前端：
//   - 会话持久化（OS Keychain/DPAPI 加密）
//   - 系统托盘 / 全局快捷键
//   - 本地 SQLite 离线缓存
//   - 自动更新检查
//
// 构建前置：
//   - Go 1.26+ · Node.js 20+
//   - Windows: WebView2 Runtime + TDM-GCC (cgo)
//   - macOS: Xcode Command Line Tools
//   - Linux: libwebkit2gtk-4.0-dev · gcc
//
// 构建命令：
//   wails build            # 开发构建
//   wails build -platform windows/amd64
//   wails build -platform darwin/universal
//   wails build -clean      # 清理后全量构建
package main

import (
	"context"
	"embed"
	"fmt"
	"log"
	"os"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/options/windows"

	"github.com/njydsz/ydsz-plane/desktop/auth"
	"github.com/njydsz/ydsz-plane/desktop/bridge"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// 日志输出到本地文件（桌面端不暴露 stderr）
	logFile, err := initDesktopLog()
	if err != nil {
		log.Fatalf("desktop: cannot init log: %v", err)
	}
	defer logFile.Close()
	log.SetOutput(logFile)

	// 装配桌面端服务
	sessionStore, err := auth.NewSessionStore()
	if err != nil {
		log.Fatalf("desktop: cannot init session store: %v", err)
	}

	evtBridge := bridge.NewEventBridge()

	// Wails 应用实例
	app := &DesktopApp{
		session: sessionStore,
		bridge:  evtBridge,
	}

	if err := wails.Run(app.opts(assets)); err != nil {
		log.Fatalf("desktop: wails run error: %v", err)
	}
}

// DesktopApp 是 Wails 应用的主结构体，
// 所有导出的方法（首字母大写）自动绑定为前端可调用函数。
type DesktopApp struct {
	session *auth.SessionStore
	bridge  *bridge.EventBridge
	ctx     context.Context
}

// opts 返回 Wails 启动选项。
func (a *DesktopApp) opts(assets embed.FS) *options.App {
	return &options.App{
		Title:             "Ydsz Plane",
		Width:             1280,
		Height:            800,
		MinWidth:          960,
		MinHeight:         600,
		DisableResize:     false,
		Fullscreen:        false,
		Frameless:         false,
		StartHidden:       false,
		HideWindowOnClose: true, // 点 × 最小化到托盘
		BackgroundColour:  &options.RGBA{R: 255, G: 255, B: 255, A: 255},
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup:  a.startup,
		OnDomReady: a.domReady,
		OnShutdown: a.shutdown,
		Bind: []interface{}{
			a.session,
			a.bridge,
		},
		Windows: &windows.Options{
			WebviewBrowserPath:   "",
			WebviewUserDataPath:  "",
			DisableWindowIcon:    false,
			DisableFramelessWindowDecorations: false,
			WebviewGpuIsDisabled: false,
			Theme:                windows.ThemeAuto,
		},
		Mac: &mac.Options{
			TitleBar:             mac.TitleBarHiddenInset,
			Appearance:           mac.NSAppearanceNameDarkAqua,
			WebviewIsTransparent: true,
			WindowIsTranslucent:  false,
			About: &mac.AboutInfo{
				Title:   "Ydsz Plane",
				Message: "面向中国软件团队的开源项目管理平台\n版本 0.1.0",
			},
		},
	}
}

// startup 在应用启动时执行（一次）。
func (a *DesktopApp) startup(ctx context.Context) {
	a.ctx = ctx
	a.bridge.SetContext(ctx)

	// 打印启动日志
	log.Println("desktop: startup complete, wails runtime ready")
}

// domReady 在前端 DOM 就绪后执行——此时可安全调用前端 JS。
func (a *DesktopApp) domReady(ctx context.Context) {
	log.Println("desktop: frontend DOM ready")
}

// shutdown 在应用退出前执行。
func (a *DesktopApp) shutdown(_ context.Context) {
	log.Println("desktop: shutting down")
}

// initDesktopLog 初始化桌面端本地日志文件。
func initDesktopLog() (*os.File, error) {
	logDir, err := getLogDir()
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return nil, fmt.Errorf("create log dir: %w", err)
	}
	return os.OpenFile(
		fmt.Sprintf("%s/desktop.log", logDir),
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0o644,
	)
}

// getLogDir 返回桌面端日志目录（跨平台）。
func getLogDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/.ydsz-plane/logs", home), nil
}
