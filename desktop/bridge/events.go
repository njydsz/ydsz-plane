// Package bridge 负责桌面端 Go ↔ JS 的事件桥接。
//
// 复用现有 WebSocket Hub 的模式：在 Go 端维护事件发射器，
// 前端通过 Wails Runtime 的 Events 机制监听。
//
// 前端调用示例：
//
//	import { Events } from '../wailsjs/runtime/runtime'
//	Events.On("notification", (n) => { ... })
//
// Go 端调用：
//
//	app.bridge.Emit("notification", payload)
package bridge

import (
	"context"
	"sync"
)

// EventBridge 维护一个事件发射器的上下文引用。
type EventBridge struct {
	ctx   context.Context
	mu    sync.RWMutex
}

// NewEventBridge 创建一个未绑定的 EventBridge 实例。
func NewEventBridge() *EventBridge {
	return &EventBridge{}
}

// SetContext 注入 Wails 运行时上下文（OnStartup 时由 Wails 调用）。
func (e *EventBridge) SetContext(ctx context.Context) {
	e.ctx = ctx
}

// EmitNotification 向前端推送一条新通知事件。
//
// Wails 绑定后前端监听：
//
//	Events.On("ydsz:notification", (payload) => { ... })
func (e *EventBridge) EmitNotification(title, body string, count int) {
	if e.ctx == nil {
		return
	}
	_ = title
	_ = body
	_ = count
	// 实际实现依赖 wails.Runtime.Events.Emit(e.ctx, "ydsz:notification", payload)
	// 此处为占位，接入 Wails Runtime 后替换
}

// EmitUnreadCount 更新未读计数（驱动系统托盘角标 + 应用徽章）。
func (e *EventBridge) EmitUnreadCount(count int) {
	if e.ctx == nil {
		return
	}
	_ = count
}

// HealthCheck 返回桥接器调试信息。
func (e *EventBridge) HealthCheck() string {
	if e.ctx == nil {
		return "bridge=uninitialized"
	}
	return "bridge=ready"
}
