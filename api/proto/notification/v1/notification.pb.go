// Package notificationv1 — NotificationService gRPC API 类型桩代码。
//
// 本文件为手工编写的编译占位桩（stub），供 Phase-0 开发期编译通过。
// 正式代码通过 `buf generate` 从 notification.proto 生成（buf.yaml 配置）。
//
// 当运行 buf generate 后，本文件将被生成的代码覆盖——请勿手工维护。
package notificationv1

import (
	context "context"
	reflect "reflect"
	sync "sync"
	time "time"

	_ "google.golang.org/genproto/googleapis/api/annotations"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
	_ "google.golang.org/protobuf/types/known/timestamppb"
)

const _ = protoimpl.Version.Minimum.ProtoFileVersion

// --- Notification ---

type Notification struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Id          int64                  `protobuf:"varint,1,opt,name=id,proto3" json:"id"`
	WorkspaceId int64                  `protobuf:"varint,2,opt,name=workspace_id,proto3" json:"workspace_id"`
	RecipientId int64                  `protobuf:"varint,3,opt,name=recipient_id,proto3" json:"recipient_id"`
	EventType   string                 `protobuf:"bytes,4,opt,name=event_type,proto3" json:"event_type"`
	EntityType  string                 `protobuf:"bytes,5,opt,name=entity_type,proto3" json:"entity_type"`
	EntityId    int64                  `protobuf:"varint,6,opt,name=entity_id,proto3" json:"entity_id"`
	Title       string                 `protobuf:"bytes,7,opt,name=title,proto3" json:"title"`
	Body        string                 `protobuf:"bytes,8,opt,name=body,proto3" json:"body"`
	ActionUrl   string                 `protobuf:"bytes,9,opt,name=action_url,proto3" json:"action_url"`
	ActorId     int64                  `protobuf:"varint,10,opt,name=actor_id,proto3" json:"actor_id"`
	ActorName   string                 `protobuf:"bytes,11,opt,name=actor_name,proto3" json:"actor_name"`
	IsRead      bool                   `protobuf:"varint,12,opt,name=is_read,proto3" json:"is_read"`
	IsArchived  bool                   `protobuf:"varint,13,opt,name=is_archived,proto3" json:"is_archived"`
	ReadAt      *timestamppb.Timestamp `protobuf:"bytes,14,opt,name=read_at,proto3" json:"read_at"`
	CreatedAt   *timestamppb.Timestamp `protobuf:"bytes,15,opt,name=created_at,proto3" json:"created_at"`
	Channel     string                 `protobuf:"bytes,16,opt,name=channel,proto3" json:"channel"`
	Payload     []byte                 `protobuf:"bytes,17,opt,name=payload,proto3" json:"payload"`
	type protoimpl.UnknownFields
	sizeCache protoimpl.SizeCache
}

func (x *Notification) Reset() { *x = Notification{} }
func (x *Notification) String() string { return "notification" }
func (m *Notification) ProtoMessage() {}
func (x *Notification) GetId() int64 { return x.Id }
func (x *Notification) GetWorkspaceId() int64 { return x.WorkspaceId }
func (x *Notification) GetRecipientId() int64 { return x.RecipientId }
func (x *Notification) GetTitle() string { return x.Title }

// --- Request / Response types ---

type ListNotificationsReq struct {
	WorkspaceId int64   `json:"workspace_id"`
	RecipientId int64   `json:"recipient_id"`
	IsRead      *bool   `json:"is_read,omitempty"`
	EventType   *string `json:"event_type,omitempty"`
	Since       *int64  `json:"since,omitempty"`
	Limit       int32   `json:"limit"`
	Offset      int32   `json:"offset"`
}

type ListNotificationsResp struct {
	Items []*Notification `json:"items"`
	Total int64           `json:"total"`
}

type UnreadCountReq struct {
	WorkspaceId int64 `json:"workspace_id"`
	RecipientId int64 `json:"recipient_id"`
}

type UnreadCountResp struct {
	Count int64 `json:"count"`
}

type MarkReadReq struct {
	NotificationId int64 `json:"notification_id"`
	RecipientId    int64 `json:"recipient_id"`
}

type MarkAllReadReq struct {
	WorkspaceId int64   `json:"workspace_id"`
	RecipientId int64   `json:"recipient_id"`
	EventType   *string `json:"event_type,omitempty"`
}

type ArchiveReq struct {
	NotificationId int64 `json:"notification_id"`
	RecipientId    int64 `json:"recipient_id"`
}

type Preferences struct {
	UserId               int64    `json:"user_id"`
	WorkspaceId          int64    `json:"workspace_id"`
	InAppEnabled         bool     `json:"in_app_enabled"`
	EmailEnabled         bool     `json:"email_enabled"`
	EmailDigestEnabled   bool     `json:"email_digest_enabled"`
	EmailDigestFrequency string   `json:"email_digest_frequency"`
	MutedEventTypes      []string `json:"muted_event_types"`
}

type Empty struct{}

// --- gRPC Service Server Interface ---

type NotificationServiceServer interface {
	List(context.Context, *ListNotificationsReq) (*ListNotificationsResp, error)
	UnreadCount(context.Context, *UnreadCountReq) (*UnreadCountResp, error)
	MarkRead(context.Context, *MarkReadReq) (*Empty, error)
	MarkAllRead(context.Context, *MarkAllReadReq) (*Empty, error)
	Archive(context.Context, *ArchiveReq) (*Empty, error)
	GetPreferences(context.Context, *UnreadCountReq) (*Preferences, error)
	UpdatePreferences(context.Context, *Preferences) (*Preferences, error)
	mustEmbedUnimplementedNotificationServiceServer()
}

type UnimplementedNotificationServiceServer struct{}

func (UnimplementedNotificationServiceServer) List(context.Context, *ListNotificationsReq) (*ListNotificationsResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method List not implemented")
}
func (UnimplementedNotificationServiceServer) UnreadCount(context.Context, *UnreadCountReq) (*UnreadCountResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UnreadCount not implemented")
}
func (UnimplementedNotificationServiceServer) MarkRead(context.Context, *MarkReadReq) (*Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method MarkRead not implemented")
}
func (UnimplementedNotificationServiceServer) MarkAllRead(context.Context, *MarkAllReadReq) (*Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method MarkAllRead not implemented")
}
func (UnimplementedNotificationServiceServer) Archive(context.Context, *ArchiveReq) (*Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Archive not implemented")
}
func (UnimplementedNotificationServiceServer) GetPreferences(context.Context, *UnreadCountReq) (*Preferences, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetPreferences not implemented")
}
func (UnimplementedNotificationServiceServer) UpdatePreferences(context.Context, *Preferences) (*Preferences, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdatePreferences not implemented")
}
func (UnimplementedNotificationServiceServer) mustEmbedUnimplementedNotificationServiceServer() {}

type UnsafeNotificationServiceServer interface{ mustEmbedUnimplementedNotificationServiceServer() }

func RegisterNotificationServiceServer(s grpc.ServiceRegistrar, srv NotificationServiceServer) {
	s.RegisterService(&NotificationService_ServiceDesc, srv)
}

func _NotificationService_List_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	return nil, status.Errorf(codes.Unimplemented, "not implemented")
}

var NotificationService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "ydszplane.notification.v1.NotificationService",
	HandlerType: (*NotificationServiceServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams:     []grpc.StreamDesc{},
	Metadata:    "api/proto/notification/v1/notification.proto",
}

// --- gRPC Client Interface ---

type NotificationServiceClient interface {
	List(context.Context, *ListNotificationsReq, ...grpc.CallOption) (*ListNotificationsResp, error)
	UnreadCount(context.Context, *UnreadCountReq, ...grpc.CallOption) (*UnreadCountResp, error)
	MarkRead(context.Context, *MarkReadReq, ...grpc.CallOption) (*Empty, error)
	MarkAllRead(context.Context, *MarkAllReadReq, ...grpc.CallOption) (*Empty, error)
	Archive(context.Context, *ArchiveReq, ...grpc.CallOption) (*Empty, error)
	GetPreferences(context.Context, *UnreadCountReq, ...grpc.CallOption) (*Preferences, error)
	UpdatePreferences(context.Context, *Preferences, ...grpc.CallOption) (*Preferences, error)
}

type notificationServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewNotificationServiceClient(cc grpc.ClientConnInterface) NotificationServiceClient {
	return &notificationServiceClient{cc: cc}
}

func (c *notificationServiceClient) List(ctx context.Context, in *ListNotificationsReq, opts ...grpc.CallOption) (*ListNotificationsResp, error) {
	return nil, nil
}
func (c *notificationServiceClient) UnreadCount(ctx context.Context, in *UnreadCountReq, opts ...grpc.CallOption) (*UnreadCountResp, error) {
	return nil, nil
}
func (c *notificationServiceClient) MarkRead(ctx context.Context, in *MarkReadReq, opts ...grpc.CallOption) (*Empty, error) {
	return nil, nil
}
func (c *notificationServiceClient) MarkAllRead(ctx context.Context, in *MarkAllReadReq, opts ...grpc.CallOption) (*Empty, error) {
	return nil, nil
}
func (c *notificationServiceClient) Archive(ctx context.Context, in *ArchiveReq, opts ...grpc.CallOption) (*Empty, error) {
	return nil, nil
}
func (c *notificationServiceClient) GetPreferences(ctx context.Context, in *UnreadCountReq, opts ...grpc.CallOption) (*Preferences, error) {
	return nil, nil
}
func (c *notificationServiceClient) UpdatePreferences(ctx context.Context, in *Preferences, opts ...grpc.CallOption) (*Preferences, error) {
	return nil, nil
}
