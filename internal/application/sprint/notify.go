package sprint

import (
	"context"
	"encoding/json"
	"fmt"

	notif "github.com/njydsz/ydsz-plane/internal/application/notification"
	"github.com/njydsz/ydsz-plane/internal/infrastructure/ws"
)

// notifyLifecycleChange 迭代启动/结束后：
//  1) 通知迭代内工作项的指派者（排除操作者本人）
//  2) 实时广播到工作空间
//
// 遵循 issue 域 notify.go 的既有模式（NotificationSvc / WSHub 可空注入，
// 缺失时静默跳过，不影响主流程）。
func (h *Handler) notifyLifecycleChange(ctx context.Context, wsID, sprintID int64, actorID int64, actorName string, eventType notif.EventType, sprintName string) {
	if h.NotificationSvc == nil {
		return
	}

	// 查询迭代内工作项的去重指派者列表
	assigneeIDs, err := h.sprintAssigneeIDs(ctx, wsID, sprintID)
	if err != nil || len(assigneeIDs) == 0 {
		return
	}

	actionURL := fmt.Sprintf("/sprints/%d", sprintID)
	title := fmt.Sprintf("%s %s「%s」", actorName, notif.EventTitles[eventType], sprintName)
	body := fmt.Sprintf("迭代「%s」中的工作项需要你关注", sprintName)

	seen := map[int64]bool{}
	inputs := make([]notif.CreateNotificationInput, 0, len(assigneeIDs))
	for _, uid := range assigneeIDs {
		if uid == actorID || seen[uid] {
			continue
		}
		seen[uid] = true
		inputs = append(inputs, notif.CreateNotificationInput{
			WorkspaceID: wsID,
			RecipientID: uid,
			EventType:   eventType,
			EntityType:  notif.EntitySprint,
			EntityID:    sprintID,
			Title:       title,
			Body:        body,
			ActionURL:   actionURL,
			ActorID:     &actorID,
			ActorName:   actorName,
			Channel:     notif.ChannelInApp,
		})
	}
	if len(inputs) == 0 {
		return
	}

	if _, err := h.NotificationSvc.CreateBatch(ctx, inputs); err != nil {
		return
	}

	// 实时广播：工作空间内所有在线客户端收到通知信号
	if h.WSHub != nil {
		_ = h.WSHub.Publish(ctx, wsID, ws.Message{
			Type: "notification.created",
			Data: json.RawMessage(`{}`),
		})
	}
}

// sprintAssigneeIDs 返回迭代内工作项的去重指派者（供通知使用）。
func (h *Handler) sprintAssigneeIDs(ctx context.Context, wsID, sprintID int64) ([]int64, error) {
	// 通过 service 暴露的数据库访问不可行时，走 handler 侧查询。
	// 说明：Service 未暴露原始 pool，这里通过 svc 的内部查询能力兜底。
	if h.svc == nil {
		return nil, nil
	}
	return h.svc.AssigneeIDs(ctx, wsID, sprintID)
}
