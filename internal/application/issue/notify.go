// Package issue 工作项通知辅助：工作项变更时的站内通知与
// WebSocket 实时广播编排。
package issue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"

	notif "github.com/njydsz/ydsz-plane/internal/application/notification"
	"github.com/njydsz/ydsz-plane/internal/infrastructure/ws"
)

// actorName 返回当前用户展示名（优先 UserNameQuery，回退 "用户"）。
func (h *IssueHandler) actorName(c *gin.Context, userID int64) string {
	if h.d.UserNameQuery != nil {
		if name := h.d.UserNameQuery(c.Request.Context(), userID); name != "" {
			return name
		}
	}
	return "用户"
}

// issueTitle 查询工作项标题（用于通知文案）。
func (h *IssueHandler) issueTitle(ctx context.Context, wsID, issueID int64) string {
	iss, err := h.d.IssueSvc.GetByID(ctx, wsID, issueID)
	if err != nil || iss == nil {
		return "工作项"
	}
	return iss.Name
}

// notifyCommentCreated 创建评论后通知被 @ 的用户，并实时广播。
func (h *IssueHandler) notifyCommentCreated(ctx context.Context, wsID, issueID int64, mentions []int64, actorID int64, actorName, issueTitle string) {
	if h.d.NotificationSvc == nil {
		return
	}
	if len(mentions) == 0 {
		return
	}

	// 去重（避免重复提及同一用户）
	seen := map[int64]bool{}
	inputs := make([]notif.CreateNotificationInput, 0, len(mentions))
	for _, uid := range mentions {
		if uid == actorID || seen[uid] {
			continue
		}
		seen[uid] = true
		inputs = append(inputs, notif.CreateNotificationInput{
			WorkspaceID: wsID,
			RecipientID: uid,
			EventType:   notif.EventCommentCreated,
			EntityType:  notif.EntityIssue,
			EntityID:    issueID,
			Title:       fmt.Sprintf("%s 在评论中提到了你", actorName),
			Body:        fmt.Sprintf("关于工作项「%s」", issueTitle),
			ActionURL:   fmt.Sprintf("/issues/%d", issueID),
			ActorID:     &actorID,
			ActorName:   actorName,
			Channel:     notif.ChannelInApp,
		})
	}
	if len(inputs) == 0 {
		return
	}

	created, err := h.d.NotificationSvc.CreateBatch(ctx, inputs)
	if err != nil || created == 0 {
		return
	}

	// 实时广播：工作空间内所有在线客户端收到新通知信号
	if h.d.WSHub != nil {
		_ = h.d.WSHub.Publish(ctx, wsID, ws.Message{
			Type: "notification.created",
			Data: json.RawMessage(`{}`),
		})
	}
}

// notifyIssueAssigned 工作项指派后通知被指派者。
func (h *IssueHandler) notifyIssueAssigned(ctx context.Context, wsID int64, assignees []int64, actorID int64, actorName, issueTitle string, issueID int64) {
	if h.d.NotificationSvc == nil || len(assignees) == 0 {
		return
	}

	seen := map[int64]bool{}
	inputs := make([]notif.CreateNotificationInput, 0, len(assignees))
	for _, uid := range assignees {
		if uid == actorID || seen[uid] {
			continue
		}
		seen[uid] = true
		inputs = append(inputs, notif.CreateNotificationInput{
			WorkspaceID: wsID,
			RecipientID: uid,
			EventType:   notif.EventIssueAssigned,
			EntityType:  notif.EntityIssue,
			EntityID:    issueID,
			Title:       fmt.Sprintf("%s 将工作项分配给了你", actorName),
			Body:        issueTitle,
			ActionURL:   fmt.Sprintf("/issues/%d", issueID),
			ActorID:     &actorID,
			ActorName:   actorName,
			Channel:     notif.ChannelInApp,
		})
	}
	if len(inputs) == 0 {
		return
	}

	created, err := h.d.NotificationSvc.CreateBatch(ctx, inputs)
	if err != nil || created == 0 {
		return
	}
	if h.d.WSHub != nil {
		_ = h.d.WSHub.Publish(ctx, wsID, ws.Message{
			Type: "notification.created",
			Data: json.RawMessage(`{}`),
		})
	}
}

// broadcastIssueUpdated 工作项状态变更后广播（供看板实时刷新）。
// payload 携带 actor_id 与 new_version，前端用于：
//   - 区分自己触发的广播（actor_id == 当前 user_id → 跳过处理，本地已乐观更新）
//   - 检测版本冲突（new_version > 本地 version → 拉取详情覆盖）
func (h *IssueHandler) broadcastIssueUpdated(ctx context.Context, wsID, projectID, issueID, actorID, newVersion int64) {
	if h.d.WSHub == nil {
		return
	}
	payload, _ := json.Marshal(map[string]int64{
		"workspace_id": wsID,
		"project_id":   projectID,
		"issue_id":     issueID,
		"actor_id":     actorID,
		"new_version":  newVersion,
	})
	_ = h.d.WSHub.Publish(ctx, wsID, ws.Message{
		Type: "issue.updated",
		Data: payload,
	})
}

// notifyIssueCreated 工作项创建后通知被指派者 + 广播。
func (h *IssueHandler) notifyIssueCreated(ctx context.Context, wsID int64, assignees []int64, actorID int64, actorName, issueTitle string, issueID int64) {
	h.notifyIssueAssigned(ctx, wsID, assignees, actorID, actorName, issueTitle, issueID)
}

// 核心通知事件类型，只有这些类型的工作项变更才会触发通知，避免噪音
var coreEventTypes = map[string]bool{
	"issue.created":       true,
	"issue.assigned":      true,
	"issue.status_changed": true,
	"issue.priority_changed": true,
	"issue.commented":     true,
	"issue.attachment_added": true,
}

// notifyIssueWatchers 工作项核心变更后通知所有关注者（watchers）。
//
// 优化点：
// 1. 仅核心变更事件（创建、指派、状态流转等）触发通知，避免普通字段修改的噪音
// 2. 同一用户对同一工作项5分钟内的多次变更合并为一条通知，避免通知风暴
func (h *IssueHandler) notifyIssueWatchers(ctx context.Context, wsID, issueID, actorID int64, eventType string, actorName, issueTitle, changeDesc string) {
	if h.d.NotificationSvc == nil || h.d.IssueSvc == nil {
		return
	}
	// 校验是否是核心事件，非核心事件不触发通知
	if !coreEventTypes[eventType] {
		return
	}
	watchers, err := h.d.IssueSvc.LoadWatchers(ctx, issueID)
	if err != nil || len(watchers) == 0 {
		return
	}

	seen := map[int64]bool{}
	inputs := make([]notif.CreateNotificationInput, 0, len(watchers))
	mergeTTL := 5 * time.Minute // 5分钟合并窗口
	for _, uid := range watchers {
		if uid == actorID || seen[uid] {
			continue
		}
		seen[uid] = true
		
		// 通知去重：同一用户对同一工作项5分钟内只发一次通知
		mergeKey := fmt.Sprintf("notif:merge:%d:%d", issueID, uid)
		// 如果key已存在，说明5分钟内已经发过通知，跳过
		count, _ := h.d.Redis.Exists(ctx, mergeKey).Result()
		if count > 0 {
			continue
		}
		// 设置key，有效期5分钟
		_ = h.d.Redis.Set(ctx, mergeKey, "1", mergeTTL).Err()
		
		title := "工作项已更新"
		if changeDesc != "" {
			title = changeDesc
		}
		inputs = append(inputs, notif.CreateNotificationInput{
			WorkspaceID: wsID,
			RecipientID: uid,
			EventType:   notif.EventType(eventType),
			EntityType:  notif.EntityIssue,
			EntityID:    issueID,
			Title:       title,
			Body:        fmt.Sprintf("你关注的工作项「%s」有新动态", issueTitle),
			ActionURL:   fmt.Sprintf("/issues/%d", issueID),
			ActorID:     &actorID,
			ActorName:   actorName,
			Channel:     notif.ChannelInApp,
		})
	}
	if len(inputs) == 0 {
		return
	}

	created, err := h.d.NotificationSvc.CreateBatch(ctx, inputs)
	if err != nil || created == 0 {
		return
	}
	if h.d.WSHub != nil {
		_ = h.d.WSHub.Publish(ctx, wsID, ws.Message{
			Type: "notification.created",
			Data: json.RawMessage(`{}`),
		})
	}
}
