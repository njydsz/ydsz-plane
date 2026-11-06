// Package aiagent — 触发器对外接入点。
//
// 供其他领域服务（评论/工作项/迭代）在事件发生时调用，
// 由本模块决定是否触发 Agent Run。
package aiagent

import (
	"context"
	"encoding/json"
	"strings"
)

// TriggerHook 触发器钩子（供领域服务注入）。
//
// 用法：
//
//	// 在评论创建后
//	if hook != nil {
//	    hook.OnCommentCreated(ctx, workspaceID, projectID, issueID, issueType, actorID, content)
//	}
type TriggerHook struct {
	dispatcher *TriggerDispatcher
}

// NewTriggerHook 构造钩子。
func NewTriggerHook(dispatcher *TriggerDispatcher) *TriggerHook {
	if dispatcher == nil || !dispatcher.bridge.IsEnabled() {
		return nil
	}
	return &TriggerHook{dispatcher: dispatcher}
}

// IsEnabled 返回是否可用。
func (h *TriggerHook) IsEnabled() bool { return h != nil && h.dispatcher != nil }

// OnCommentCreated 评论创建事件（解析 @mention 触发）。
func (h *TriggerHook) OnCommentCreated(ctx context.Context, workspaceID, projectID, issueID int64, issueType string, actorID int64, content string) {
	if h == nil {
		return
	}
	mentioned := parseMentionNames(content)
	if len(mentioned) == 0 {
		return
	}
	_, _, _ = h.dispatcher.Dispatch(ctx, DomainEvent{
		Type:        TriggerMention,
		WorkspaceID: workspaceID,
		DedupKey:    dedupKey("mention", workspaceID, issueID, content),
		Payload: map[string]any{
			"payload": &EventPayload{
				ProjectID:      projectID,
				WorkItemID:     issueID,
				WorkItemType:   issueType,
				Action:         "commented",
				ActorID:        actorID,
				Content:        content,
				MentionedNames: mentioned,
			},
		},
	})
}

// OnWorkItemCreated 工作项创建事件。
func (h *TriggerHook) OnWorkItemCreated(ctx context.Context, workspaceID, projectID, issueID int64, issueType, title, description string, actorID int64) {
	if h == nil {
		return
	}
	_, _, _ = h.dispatcher.Dispatch(ctx, DomainEvent{
		Type:        TriggerEvent,
		WorkspaceID: workspaceID,
		DedupKey:    dedupKey("created", workspaceID, issueID, ""),
		Payload: map[string]any{
			"payload": &EventPayload{
				ProjectID:    projectID,
				WorkItemID:   issueID,
				WorkItemType: issueType,
				Action:       "created",
				ActorID:      actorID,
				Title:        title,
				Content:      description,
			},
		},
	})
}

// OnWorkItemAssigned 工作项指派事件。
func (h *TriggerHook) OnWorkItemAssigned(ctx context.Context, workspaceID, projectID, issueID int64, issueType, title string, assigneeID, actorID int64) {
	if h == nil {
		return
	}
	_, _, _ = h.dispatcher.Dispatch(ctx, DomainEvent{
		Type:        TriggerAssignment,
		WorkspaceID: workspaceID,
		DedupKey:    dedupKey("assigned", workspaceID, issueID, title),
		Payload: map[string]any{
			"payload": &EventPayload{
				ProjectID:    projectID,
				WorkItemID:   issueID,
				WorkItemType: issueType,
				Action:       "assigned",
				ActorID:      actorID,
				Title:        title,
			},
		},
	})
}

// OnStateChanged 状态变更事件。
func (h *TriggerHook) OnStateChanged(ctx context.Context, workspaceID, projectID, issueID int64, issueType, stateName string, actorID int64) {
	if h == nil {
		return
	}
	_, _, _ = h.dispatcher.Dispatch(ctx, DomainEvent{
		Type:        TriggerEvent,
		WorkspaceID: workspaceID,
		DedupKey:    dedupKey("state", workspaceID, issueID, stateName),
		Payload: map[string]any{
			"payload": &EventPayload{
				ProjectID:    projectID,
				WorkItemID:   issueID,
				WorkItemType: issueType,
				Action:       "state_changed",
				ActorID:      actorID,
				StateName:    stateName,
			},
		},
	})
}

// TriggerDueSchedules 触发到期调度（供定时任务调用）。
func (h *TriggerHook) TriggerDueSchedules(ctx context.Context, workspaceID int64) (int, error) {
	if h == nil {
		return 0, nil
	}
	runs, _, err := h.dispatcher.Dispatch(ctx, DomainEvent{
		Type:        TriggerSchedule,
		WorkspaceID: workspaceID,
		DedupKey:    dedupKey("schedule", workspaceID, 0, ""),
		Payload:     map[string]any{"payload": &EventPayload{Action: "scheduled"}},
	})
	return len(runs), err
}

// --- helpers ---

// parseMentionNames 从文本中提取 @username。
func parseMentionNames(content string) []string {
	var out []string
	for {
		i := strings.Index(content, "@")
		if i < 0 {
			break
		}
		rest := content[i+1:]
		end := strings.IndexAny(rest, " \t\r\n,，。.;；:：!！?？")
		name := rest
		if end >= 0 {
			name = rest[:end]
			content = rest[end:]
		} else {
			content = ""
		}
		name = strings.TrimSpace(name)
		if name != "" {
			out = append(out, name)
		}
	}
	return out
}

func dedupKey(parts ...any) string {
	b, _ := json.Marshal(parts)
	return string(b)
}
