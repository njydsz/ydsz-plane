// Package aiagent — Agent 触发器运行时。
//
// 对标 Plane Agents 的触发方式：
//   - mention：工作项描述/评论中被 @ 提及；
//   - assignment：工作项被分配给 Agent 可服务的成员（或 Agent 自身）；
//   - event：工作项事件（创建/状态变更/评论）；
//   - schedule：定时调度（cron）。
//
// 设计：
//   - TriggerDispatcher 接收领域事件，匹配已注册 Agent 的 triggers，
//     为每个匹配项创建 AgentRun 并异步执行；
//   - 执行复用 LLMBridge 的 Chat 能力；
//   - 幂等：同一事件 + 同一 Agent 在短时间窗内只触发一次。
package aiagent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"
)

// TriggerType 触发类型。
type TriggerType string

const (
	// TriggerMention 被提及。
	TriggerMention TriggerType = "mention"
	// TriggerAssignment 被指派。
	TriggerAssignment TriggerType = "assignment"
	// TriggerEvent 工作项事件。
	TriggerEvent TriggerType = "event"
	// TriggerSchedule 定时调度。
	TriggerSchedule TriggerType = "schedule"
)

// IsValidTrigger 校验触发类型。
func IsValidTrigger(t string) bool {
	switch TriggerType(t) {
	case TriggerMention, TriggerAssignment, TriggerEvent, TriggerSchedule:
		return true
	}
	return false
}

// DomainEvent 领域事件（触发源）。
type DomainEvent struct {
	Type        TriggerType
	WorkspaceID int64
	AgentIDs    []int64 // 仅 event/mention 场景可指定
	// 事件上下文
	Payload map[string]any
	// 幂等键（相同键在窗口内去重）
	DedupKey string
}

// EventPayload 事件载荷（统一结构）。
type EventPayload struct {
	// 工作项上下文
	ProjectID    int64  `json:"project_id,omitempty"`
	WorkItemID   int64  `json:"work_item_id,omitempty"`
	WorkItemType string `json:"work_item_type,omitempty"`
	// 事件内容
	Action    string `json:"action,omitempty"`   // created / state_changed / commented / assigned
	ActorID   int64  `json:"actor_id,omitempty"` // 触发者
	Content   string `json:"content,omitempty"`  // 评论/描述内容（用于 @ 解析）
	Title     string `json:"title,omitempty"`
	StateName string `json:"state_name,omitempty"`
	// 提及的用户名
	MentionedNames []string `json:"mentioned_names,omitempty"`
}

// TriggerDispatcher 触发器分发器。
type TriggerDispatcher struct {
	svc        *Service
	runSvc     *RunService
	bridge     *LLMBridge
	agentStore AgentReader

	// 去重窗口
	dedupTTL  time.Duration
	dedupMu   sync.Mutex
	dedupSeen map[string]time.Time
}

// AgentReader Agent 读取接口（避免与具体存储耦合）。
type AgentReader interface {
	ListAgents(ctx context.Context, workspaceID int64) ([]*Agent, error)
}

// NewTriggerDispatcher 构造分发器。
func NewTriggerDispatcher(svc *Service, runSvc *RunService, bridge *LLMBridge) *TriggerDispatcher {
	return &TriggerDispatcher{
		svc:        svc,
		runSvc:     runSvc,
		bridge:     bridge,
		agentStore: svc,
		dedupTTL:   5 * time.Minute,
		dedupSeen:  map[string]time.Time{},
	}
}

// Dispatch 处理领域事件，为匹配的 Agent 创建并执行 Run。
//
// 返回本次触发创建的 Run 列表（含被去重跳过的数量）。
func (d *TriggerDispatcher) Dispatch(ctx context.Context, ev DomainEvent) ([]*AgentRun, int, error) {
	if d.bridge == nil || !d.bridge.IsEnabled() {
		return nil, 0, nil
	}

	// 去重
	if ev.DedupKey != "" && !d.markSeen(ev.DedupKey) {
		return nil, 1, nil
	}

	agents, err := d.agentStore.ListAgents(ctx, ev.WorkspaceID)
	if err != nil {
		return nil, 0, err
	}

	var created []*AgentRun
	skipped := 0

	for _, agent := range agents {
		if agent.Status != "active" {
			continue
		}
		if !agentHasTrigger(agent, ev.Type) {
			continue
		}
		if len(ev.AgentIDs) > 0 && !containsID(ev.AgentIDs, agent.ID) {
			continue
		}

		run, err := d.execute(ctx, agent, ev)
		if err != nil {
			// 单个 Agent 失败不影响其他 Agent
			skipped++
			continue
		}
		if run != nil {
			created = append(created, run)
		}
	}
	return created, skipped, nil
}

// execute 为单个 Agent 创建并执行 Run。
func (d *TriggerDispatcher) execute(ctx context.Context, agent *Agent, ev DomainEvent) (*AgentRun, error) {
	if d.runSvc == nil {
		return nil, fmt.Errorf("run service not configured")
	}

	input := buildInput(ev)
	mode := ModeAutopilot
	if agentContainsSkill(agent, "ask-only") {
		mode = ModeAsk
	}

	run, err := d.runSvc.CreateRun(ctx, CreateRunInput{
		WorkspaceID:    ev.WorkspaceID,
		AgentID:        agent.ID,
		UserID:         payloadActor(ev.Payload),
		ConversationID: 0,
		Mode:           mode,
		Input:          input,
		Trigger:        string(ev.Type),
		TriggerPayload: marshalPayload(ev.Payload),
	})
	if err != nil {
		return nil, err
	}

	if _, err := d.runSvc.Transition(ctx, run.ID, RunInProgress, "", ""); err != nil {
		return run, nil
	}

	// 异步执行（不阻塞事件源）
	go d.runAsync(context.Background(), agent, run, ev)

	return run, nil
}

// runAsync 在后台执行 Agent 并回写结果。
func (d *TriggerDispatcher) runAsync(ctx context.Context, agent *Agent, run *AgentRun, ev DomainEvent) {
	if d.bridge == nil {
		_, _ = d.runSvc.Transition(ctx, run.ID, RunFailed, "", "LLM 未配置")
		return
	}

	ctx = WithActor(ctx, ev.WorkspaceID, run.UserID)
	result, err := d.bridge.Chat(ctx, agent, run.Mode, "", nil, run.Input, agent.Temperature)
	if err != nil {
		_, _ = d.runSvc.Transition(ctx, run.ID, RunFailed, "", err.Error())
		return
	}

	if result.FinalText != "" {
		_, _ = d.runSvc.AddStep(ctx, run.ID, "response", "生成回复", result.FinalText)
	}
	for _, tc := range result.ToolCalls {
		_, _ = d.runSvc.AddStep(ctx, run.ID, "action", tc.Name, fmt.Sprintf("%v", tc.Args))
	}
	_ = d.runSvc.RecordUsage(ctx, run.ID, result.Usage.PromptTokens, result.Usage.CompletionTokens, len(result.ToolCalls))
	_, _ = d.runSvc.Transition(ctx, run.ID, RunCompleted, "", result.FinalText)
}

// --- helpers ---

// markSeen 记录去重键，返回是否为首次（true=首次）。
func (d *TriggerDispatcher) markSeen(key string) bool {
	d.dedupMu.Lock()
	defer d.dedupMu.Unlock()

	now := time.Now()
	// 清理过期
	for k, t := range d.dedupSeen {
		if now.Sub(t) > d.dedupTTL {
			delete(d.dedupSeen, k)
		}
	}
	if _, ok := d.dedupSeen[key]; ok {
		return false
	}
	d.dedupSeen[key] = now
	return true
}

func agentHasTrigger(a *Agent, t TriggerType) bool {
	for _, tr := range a.Triggers {
		if strings.EqualFold(strings.TrimSpace(tr), string(t)) {
			return true
		}
	}
	return false
}

func agentContainsSkill(a *Agent, skill string) bool {
	for _, s := range a.Skills {
		if strings.EqualFold(strings.TrimSpace(s), skill) {
			return true
		}
	}
	return false
}

func containsID(ids []int64, id int64) bool {
	for _, v := range ids {
		if v == id {
			return true
		}
	}
	return false
}

// buildInput 根据事件构造 Agent 输入。
func buildInput(ev DomainEvent) string {
	p, _ := ev.Payload["payload"].(*EventPayload)
	if p == nil {
		if raw, ok := ev.Payload["payload"].(map[string]any); ok {
			b, _ := json.Marshal(raw)
			var parsed EventPayload
			_ = json.Unmarshal(b, &parsed)
			p = &parsed
		}
	}
	if p == nil {
		return fmt.Sprintf("收到 %s 事件：%v", ev.Type, ev.Payload)
	}

	var sb strings.Builder
	switch ev.Type {
	case TriggerMention:
		sb.WriteString("你在工作项中被提及了。")
	case TriggerAssignment:
		sb.WriteString("你被指派处理一个工作项。")
	case TriggerEvent:
		sb.WriteString(fmt.Sprintf("工作项发生了 %s 事件。", p.Action))
	case TriggerSchedule:
		sb.WriteString("定时任务触发。")
	}
	if p.Title != "" {
		sb.WriteString("\n标题：")
		sb.WriteString(p.Title)
	}
	if p.Content != "" {
		sb.WriteString("\n内容：")
		sb.WriteString(p.Content)
	}
	sb.WriteString("\n\n请根据以上信息判断是否需要处理；如无需处理请简要说明原因。")
	return sb.String()
}

func payloadActor(payload map[string]any) int64 {
	if p, ok := payload["payload"].(*EventPayload); ok && p != nil {
		return p.ActorID
	}
	if v, ok := payload["actor_id"].(float64); ok {
		return int64(v)
	}
	return 0
}

func marshalPayload(payload map[string]any) string {
	b, err := json.Marshal(payload)
	if err != nil {
		return ""
	}
	return string(b)
}
