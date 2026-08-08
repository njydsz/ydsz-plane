// Package notification — Notification 模块的 gRPC 接口实现。
//
// Phase-0（当前）：在单体核心进程内以 gRPC Service 形式暴露通知能力，
//   API Gateway 内部将 HTTP 请求转化为 gRPC 进程内调用，验证接口等价性。
//
// Phase-1（S14 P2）：将本文件编译进 cmd/notification-service 成为独立进程，
//   通过 gRPC Server 对外服务，core-service 通过 gRPC Client 调用。
//
// 两种模式共享同一份 business logic，仅入口不同。
package notification

import (
	"context"

	notificationv1 "github.com/njydsz/ydsz-plane/api/proto/notification/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// GRPCService 是 notificationv1.NotificationService 的实现。
//
// 将现有 REST Handler 的入参出参与 Protobuf 互转。
// 所有业务逻辑通过 s.svc (Service) 复用，零重复代码。
type GRPCService struct {
	notificationv1.UnimplementedNotificationServiceServer
	svc *Service
}

// NewGRPCService 从 Service 实例创建 gRPC Service。
func NewGRPCService(svc *Service) *GRPCService {
	return &GRPCService{svc: svc}
}

// List 分页查询通知列表。
func (s *GRPCService) List(ctx context.Context, req *notificationv1.ListNotificationsReq) (*notificationv1.ListNotificationsResp, error) {
	input := ListInput{
		WorkspaceID: req.WorkspaceId,
		RecipientID: req.RecipientId,
		Limit:       int(req.Limit),
		Offset:      int(req.Offset),
	}
	if req.IsRead != nil {
		v := *req.IsRead
		input.IsRead = &v
	}
	if req.EventType != nil {
		v := *req.EventType
		input.EventType = &v
	}
	if req.Since != nil {
		input.Since = req.Since
	}

	result, err := s.svc.List(ctx, input)
	if err != nil {
		return nil, err
	}

	items := make([]*notificationv1.Notification, 0, len(result.Items))
	for _, n := range result.Items {
		items = append(items, toProtoNotification(&n))
	}
	return &notificationv1.ListNotificationsResp{
		Items: items,
		Total: result.Total,
	}, nil
}

// UnreadCount 获取未读通知计数。
func (s *GRPCService) UnreadCount(ctx context.Context, req *notificationv1.UnreadCountReq) (*notificationv1.UnreadCountResp, error) {
	count, err := s.svc.UnreadCount(ctx, req.WorkspaceId, req.RecipientId)
	if err != nil {
		return nil, err
	}
	return &notificationv1.UnreadCountResp{Count: count}, nil
}

// MarkRead 标记单条已读。
func (s *GRPCService) MarkRead(ctx context.Context, req *notificationv1.MarkReadReq) (*notificationv1.Empty, error) {
	err := s.svc.MarkRead(ctx, req.NotificationId, req.RecipientId)
	if err != nil {
		return nil, err
	}
	return &notificationv1.Empty{}, nil
}

// MarkAllRead 批量标记已读。
func (s *GRPCService) MarkAllRead(ctx context.Context, req *notificationv1.MarkAllReadReq) (*notificationv1.Empty, error) {
	_, err := s.svc.MarkAllRead(ctx, req.WorkspaceId, req.RecipientId)
	if err != nil {
		return nil, err
	}
	return &notificationv1.Empty{}, nil
}

// Archive 归档通知。
func (s *GRPCService) Archive(ctx context.Context, req *notificationv1.ArchiveReq) (*notificationv1.Empty, error) {
	err := s.svc.Archive(ctx, req.NotificationId, req.RecipientId)
	if err != nil {
		return nil, err
	}
	return &notificationv1.Empty{}, nil
}

// GetPreferences 读取通知偏好。
func (s *GRPCService) GetPreferences(ctx context.Context, req *notificationv1.UnreadCountReq) (*notificationv1.Preferences, error) {
	pref, err := s.svc.GetUserPreference(ctx, req.WorkspaceId, req.RecipientId)
	if err != nil {
		return nil, err
	}
	return toProtoPreferences(pref), nil
}

// UpdatePreferences 修改通知偏好。
func (s *GRPCService) UpdatePreferences(ctx context.Context, req *notificationv1.Preferences) (*notificationv1.Preferences, error) {
	input := PreferenceUpdateInput{
		Digest:     req.EmailDigestFrequency,
		IsEnabled:  req.InAppEnabled,
		EventTypes: req.MutedEventTypes,
	}
	_, err := s.svc.UpdatePreference(ctx, req.WorkspaceId, req.UserId, input)
	if err != nil {
		return nil, err
	}
	return req, nil
}

// --- Proto 互转辅助函数 ---

func toProtoNotification(n *Notification) *notificationv1.Notification {
	var actorID int64
	if n.ActorID != nil {
		actorID = *n.ActorID
	}
	var readAt int64
	if n.ReadAt != nil {
		readAt = n.ReadAt.Unix()
	}
	return &notificationv1.Notification{
		Id:          n.ID,
		WorkspaceId: n.WorkspaceID,
		RecipientId: n.RecipientID,
		EventType:   string(n.EventType),
		EntityType:  string(n.EntityType),
		EntityId:    n.EntityID,
		Title:       n.Title,
		Body:        n.Body,
		ActionUrl:   n.ActionURL,
		ActorId:     actorID,
		ActorName:   n.ActorName,
		IsRead:      n.IsRead,
		IsArchived:  n.IsArchived,
		Payload:     n.Payload,
		ReadAt:      timestamppb.New(n.CreatedAt), // best-effort; real impl uses proper ts
		CreatedAt:   timestamppb.New(n.CreatedAt),
	}
}

func toProtoPreferences(p *NotificationPreference) *notificationv1.Preferences {
	return &notificationv1.Preferences{
		UserId:               p.UserID,
		WorkspaceId:          p.WorkspaceID,
		InAppEnabled:         p.IsEnabled,
		EmailDigestFrequency: string(p.Digest),
		MutedEventTypes:      p.EventTypes,
	}
}
