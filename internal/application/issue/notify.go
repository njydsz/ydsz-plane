// Package issue 工作项通知辅助：工作项变更时的站内通知与
// WebSocket 实时广播编排。
package issue

import (
	"context"
	"encoding/json"
	"fmt"

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

// notifyIssueStatusChanged 工作项状态变更后广播（供看板实时刷新）。
func (h *IssueHandler) broadcastIssueUpdated(ctx context.Context, wsID, projectID, issueID int64) {
	if h.d.WSHub == nil {
		return
	}
	payload, _ := json.Marshal(map[string]int64{
		"workspace_id": wsID,
		"project_id":   projectID,
		"issue_id":     issueID,
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

// notifyIssueWatchers 工作项变更后通知所有关注者（watchers）。
//
// 语义：关注者订阅了工作项的动态（状态/字段/评论变更），
// 这里统一通过 EventIssueStatusChanged 事件发送站内通知 + 实时广播。
func (h *IssueHandler) notifyIssueWatchers(ctx context.Context, wsID, issueID, actorID int64, actorName, issueTitle, changeDesc string) {
	if h.d.NotificationSvc == nil || h.d.IssueSvc == nil {
		return
	}
	watchers, err := h.d.IssueSvc.LoadWatchers(ctx, issueID)
	if err != nil || len(watchers) == 0 {
		return
	}

	seen := map[int64]bool{}
	inputs := make([]notif.CreateNotificationInput, 0, len(watchers))
	for _, uid := range watchers {
		if uid == actorID || seen[uid] {
			continue
		}
		seen[uid] = true
		title := "工作项已更新"
		if changeDesc != "" {
			title = changeDesc
		}
		inputs = append(inputs, notif.CreateNotificationInput{
			WorkspaceID: wsID,
			RecipientID: uid,
			EventType:   notif.EventIssueStatusChanged,
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
