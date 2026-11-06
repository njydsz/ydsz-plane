// Ydsz Plane 桌面客户端（Wails v3）。
//
// 架构：
//   - 复用现有 Go 后端能力（通过 Service 暴露给前端）；
//   - 前端为现有 Vue 3 SPA（desktop/wails/frontend → 构建产物）；
//   - 系统能力：原生通知、窗口控制、剪贴板、自动更新。
//
// 构建：
//   wails3 build           # 生产构建
//   wails3 dev             # 开发（热重载）
//   wails3 build -platform windows/amd64
package main

import (
	"embed"
	"log"

	"github.com/wailsapp/wails/v3/pkg/application"
	applicationNotifications "github.com/wailsapp/wails/v3/pkg/services/notifications"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	notifier := &NotificationService{ inner: applicationNotifications.New() }
	windows := &WindowService{}

	app := application.New(application.Options{
		Name:        "Ydsz Plane",
		Description: "开源项目管理平台桌面客户端",
		Services: []application.Service{
			application.NewService(&PlatformService{}),
			application.NewService(notifier),
			application.NewService(windows),
			application.NewService(&UpdaterService{}),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: false,
		},
	})

	// 注入 app 引用
	windows.app = app

	// 主窗口
	app.Window.NewWithOptions(application.WebviewWindowOptions{
		Name:           "main",
		Title:          "Ydsz Plane",
		URL:            "/",
		Width:          1440,
		Height:         900,
		MinWidth:       1024,
		MinHeight:      640,
		DisableResize:  false,
		StartState:     application.WindowStateNormal,
		BackgroundType: application.BackgroundTypeSolid,
		BackgroundColour: application.NewRGBA(255, 255, 255, 255),
		EnableFileDrop: true,
	})

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
