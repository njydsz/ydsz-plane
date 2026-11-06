package main

import (
	applicationNotifications "github.com/wailsapp/wails/v3/pkg/services/notifications"
)

// NotificationService — 原生系统通知（封装 Wails 内置通知服务）。
type NotificationService struct {
	inner *applicationNotifications.NotificationService
}

// Notify 发送一条原生通知。
func (s *NotificationService) Notify(title, body, _ string) error {
	return s.inner.SendNotification(applicationNotifications.NotificationOptions{
		Title: title,
		Body:  body,
	})
}

// NotifyWithActions 发送带操作按钮的通知。
func (s *NotificationService) NotifyWithActions(title, body string, actions []string) error {
	return s.inner.SendNotificationWithActions(applicationNotifications.NotificationOptions{
		Title: title,
		Body:  body,
	})
}

// RequestPermission 请求通知权限（macOS）。
func (s *NotificationService) RequestPermission() bool {
	ok, _ := s.inner.RequestNotificationAuthorization()
	return ok
}

// HasPermission 返回是否已授权。
func (s *NotificationService) HasPermission() bool {
	ok, _ := s.inner.CheckNotificationAuthorization()
	return ok
}
