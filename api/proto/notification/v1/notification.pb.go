// Package notificationv1 — NotificationService gRPC API 类型桩代码（手工 stub）。
//
// 本文件为 Phase-0 开发期手工编写的轻量编译占位桩。
// 正式代码由 `buf generate` 从 notification.proto 自动生成并覆盖本文件。
//
// 前置依赖（需在线安装）：
//   go get google.golang.org/grpc@v1.78.0
//   随后运行 buf generate 生成正式代码
package notificationv1

import (
	context "context"
)

// ============== 领域消息（Plain Structs）==============

type Notification struct {
	Id          int64
	WorkspaceId int64
	RecipientId int64
	EventType   string
	EntityType  string
	EntityId    int64
	Title       string
	Body        string
	ActionUrl   string
	ActorId     int64
	ActorName   string
	IsRead      bool
	IsArchived  bool
	ReadAt      int64 // Unix ms
	CreatedAt   int64 // Unix ms
	Channel     string
	Payload     []byte
}

type ListNotificationsReq struct {
	WorkspaceId int64
	RecipientId int64
	IsRead      *bool
	EventType   *string
	Since       *int64
	Limit       int32
	Offset      int32
}

type ListNotificationsResp struct {
	Items []*Notification
	Total int64
}

type UnreadCountReq struct {
	WorkspaceId int64
	RecipientId int64
}

type UnreadCountResp struct {
	Count int64
}

type MarkReadReq struct {
	NotificationId int64
	RecipientId    int64
}

type MarkAllReadReq struct {
	WorkspaceId int64
	RecipientId int64
	EventType   *string
}

type ArchiveReq struct {
	NotificationId int64
	RecipientId    int64
}

type Preferences struct {
	UserId               int64
	WorkspaceId          int64
	InAppEnabled         bool
	EmailEnabled         bool
	EmailDigestEnabled   bool
	EmailDigestFrequency string
	MutedEventTypes      []string
}

type Empty struct{}

// ============== gRPC Server 接口 ==============

type NotificationServiceServer interface {
	List(context.Context, *ListNotificationsReq) (*ListNotificationsResp, error)
	UnreadCount(context.Context, *UnreadCountReq) (*UnreadCountResp, error)
	MarkRead(context.Context, *MarkReadReq) (*Empty, error)
	MarkAllRead(context.Context, *MarkAllReadReq) (*Empty, error)
	Archive(context.Context, *ArchiveReq) (*Empty, error)
	GetPreferences(context.Context, *UnreadCountReq) (*Preferences, error)
	UpdatePreferences(context.Context, *Preferences) (*Preferences, error)
	MustEmbedUnimplementedNotificationServiceServer()
}

type UnimplementedNotificationServiceServer struct{}

func (UnimplementedNotificationServiceServer) List(context.Context, *ListNotificationsReq) (*ListNotificationsResp, error) {
	return nil, nil
}
func (UnimplementedNotificationServiceServer) UnreadCount(context.Context, *UnreadCountReq) (*UnreadCountResp, error) {
	return nil, nil
}
func (UnimplementedNotificationServiceServer) MarkRead(context.Context, *MarkReadReq) (*Empty, error) {
	return nil, nil
}
func (UnimplementedNotificationServiceServer) MarkAllRead(context.Context, *MarkAllReadReq) (*Empty, error) {
	return nil, nil
}
func (UnimplementedNotificationServiceServer) Archive(context.Context, *ArchiveReq) (*Empty, error) {
	return nil, nil
}
func (UnimplementedNotificationServiceServer) GetPreferences(context.Context, *UnreadCountReq) (*Preferences, error) {
	return nil, nil
}
func (UnimplementedNotificationServiceServer) UpdatePreferences(context.Context, *Preferences) (*Preferences, error) {
	return nil, nil
}
func (UnimplementedNotificationServiceServer) MustEmbedUnimplementedNotificationServiceServer() {}

// RegisterNotificationServiceServer 注册 gRPC 服务实现。
//
// 注意：本 stub 使用 interface{} 避免强依赖 grpc（未安装时可编译）。
// 真实实现由 buf generate 生成后使用 grpc.ServiceRegistrar。
func RegisterNotificationServiceServer(s interface{ RegisterService(desc interface{}, impl interface{}) }, srv NotificationServiceServer) {
	// stub: 真实实现由 buf generate 生成后替换
	_ = srv
}

// ============== gRPC Client 接口 ==============

type NotificationServiceClient interface {
	List(context.Context, *ListNotificationsReq) (*ListNotificationsResp, error)
	UnreadCount(context.Context, *UnreadCountReq) (*UnreadCountResp, error)
	MarkRead(context.Context, *MarkReadReq) (*Empty, error)
	MarkAllRead(context.Context, *MarkAllReadReq) (*Empty, error)
	Archive(context.Context, *ArchiveReq) (*Empty, error)
	GetPreferences(context.Context, *UnreadCountReq) (*Preferences, error)
	UpdatePreferences(context.Context, *Preferences) (*Preferences, error)
}

type notificationServiceClient struct {
	cc interface{}
}

func NewNotificationServiceClient(cc interface{}) NotificationServiceClient {
	return &notificationServiceClient{cc: cc}
}

func (c *notificationServiceClient) List(ctx context.Context, in *ListNotificationsReq) (*ListNotificationsResp, error) {
	return nil, nil
}
func (c *notificationServiceClient) UnreadCount(ctx context.Context, in *UnreadCountReq) (*UnreadCountResp, error) {
	return nil, nil
}
func (c *notificationServiceClient) MarkRead(ctx context.Context, in *MarkReadReq) (*Empty, error) {
	return nil, nil
}
func (c *notificationServiceClient) MarkAllRead(ctx context.Context, in *MarkAllReadReq) (*Empty, error) {
	return nil, nil
}
func (c *notificationServiceClient) Archive(ctx context.Context, in *ArchiveReq) (*Empty, error) {
	return nil, nil
}
func (c *notificationServiceClient) GetPreferences(ctx context.Context, in *UnreadCountReq) (*Preferences, error) {
	return nil, nil
}
func (c *notificationServiceClient) UpdatePreferences(ctx context.Context, in *Preferences) (*Preferences, error) {
	return nil, nil
}
