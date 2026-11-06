// Package llm — 工具定义与函数调用引擎。
package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
)

// ToolHandler 工具执行函数。
//
// args 为模型给出的 JSON 参数（已解析）；返回值会被序列化为 JSON 回传给模型。
type ToolHandler func(ctx context.Context, args json.RawMessage) (any, error)

// ToolEntry 工具注册表项。
type ToolEntry struct {
	Definition Tool
	Handler    ToolHandler
}

// ToolRegistry 工具注册表（并发安全）。
type ToolRegistry struct {
	mu    sync.RWMutex
	tools map[string]ToolEntry
}

// NewToolRegistry 创建注册表。
func NewToolRegistry() *ToolRegistry {
	return &ToolRegistry{tools: map[string]ToolEntry{}}
}

// Register 注册工具。
func (r *ToolRegistry) Register(def ToolDefinition, handler ToolHandler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tools[def.Name] = ToolEntry{
		Definition: Tool{Type: "function", Function: def},
		Handler:    handler,
	}
}

// Definitions 返回全部工具定义（用于 tools/list）。
func (r *ToolRegistry) Definitions() []Tool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]Tool, 0, len(r.tools))
	for _, e := range r.tools {
		out = append(out, e.Definition)
	}
	return out
}

// Has 判断工具是否已注册。
func (r *ToolRegistry) Has(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.tools[name]
	return ok
}

// Execute 执行单个工具调用。
func (r *ToolRegistry) Execute(ctx context.Context, name string, args json.RawMessage) (any, error) {
	r.mu.RLock()
	entry, ok := r.tools[name]
	r.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("llm: unknown tool %q", name)
	}
	if len(args) == 0 {
		args = json.RawMessage("{}")
	}
	return entry.Handler(ctx, args)
}

// --- 通用工具集 ---

// WorkspaceTools 返回工作空间操作工具定义（供 Agent / MCP 共用）。
//
// 共 12 个工具：查询类 6 + 写入类 6。
func WorkspaceTools() []ToolDefinition {
	str := func(desc string) map[string]any {
		return map[string]any{"type": "string", "description": desc}
	}
	num := func(desc string) map[string]any {
		return map[string]any{"type": "number", "description": desc}
	}
	object := func(props map[string]any, required ...string) map[string]any {
		return map[string]any{"type": "object", "properties": props, "required": required}
	}

	return []ToolDefinition{
		{
			Name:        "list_work_items",
			Description: "列出工作项（需求/任务/缺陷），支持按项目、类型、状态、指派人过滤",
			Parameters: object(map[string]any{
				"project_id": num("项目 ID"),
				"type":       str("类型：requirement/task/defect"),
				"state":      str("状态名称"),
				"assignee":   str("指派人用户名，或 me 表示当前用户"),
				"limit":      num("返回上限，默认 50"),
				"offset":     num("偏移量，默认 0"),
			}),
		},
		{
			Name:        "get_work_item",
			Description: "获取单个工作项详情（标题、描述、状态、优先级、指派人、迭代、模块、标签）",
			Parameters: object(map[string]any{
				"project_id": num("项目 ID"),
				"type":       str("类型：requirement/task/defect"),
				"id":         num("工作项 ID"),
			}, "project_id", "type", "id"),
		},
		{
			Name:        "search_workspace",
			Description: "跨项目搜索工作项、文档、迭代（全文检索）",
			Parameters: object(map[string]any{
				"query": str("搜索关键词"),
				"limit": num("返回上限，默认 20"),
			}, "query"),
		},
		{
			Name:        "list_cycles",
			Description: "列出迭代（Sprint），可按状态过滤（planned/active/completed）",
			Parameters: object(map[string]any{
				"project_id": num("项目 ID"),
				"status":     str("状态过滤"),
			}),
		},
		{
			Name:        "list_modules",
			Description: "列出项目模块",
			Parameters: object(map[string]any{
				"project_id": num("项目 ID"),
			}),
		},
		{
			Name:        "list_members",
			Description: "列出工作空间成员（用于指派人解析）",
			Parameters:  object(map[string]any{}),
		},

		// --- 写入类 ---
		{
			Name:        "create_work_item",
			Description: "创建需求/任务/缺陷",
			Parameters: object(map[string]any{
				"project_id":  num("项目 ID"),
				"type":        str("类型：requirement/task/defect"),
				"title":       str("标题"),
				"description": str("描述"),
				"priority":    str("优先级：urgent/high/medium/low/none"),
				"assignee_id": num("指派人 ID"),
			}, "project_id", "type", "title"),
		},
		{
			Name:        "update_work_item",
			Description: "更新工作项字段（标题/描述/优先级/状态）",
			Parameters: object(map[string]any{
				"project_id":  num("项目 ID"),
				"type":        str("类型：requirement/task/defect"),
				"id":          num("工作项 ID"),
				"title":       str("新标题"),
				"description": str("新描述"),
				"priority":    str("新优先级"),
				"state_id":    num("新状态 ID"),
			}, "project_id", "type", "id"),
		},
		{
			Name:        "transition_work_item",
			Description: "流转工作项状态",
			Parameters: object(map[string]any{
				"project_id": num("项目 ID"),
				"type":       str("类型"),
				"id":         num("工作项 ID"),
				"state_id":   num("目标状态 ID"),
			}, "project_id", "type", "id", "state_id"),
		},
		{
			Name:        "assign_work_item",
			Description: "指派工作项给指定成员",
			Parameters: object(map[string]any{
				"project_id": num("项目 ID"),
				"type":       str("类型"),
				"id":         num("工作项 ID"),
				"user_id":    num("被指派成员 ID"),
			}, "project_id", "type", "id", "user_id"),
		},
		{
			Name:        "add_comment",
			Description: "给工作项添加评论",
			Parameters: object(map[string]any{
				"project_id": num("项目 ID"),
				"type":       str("类型"),
				"id":         num("工作项 ID"),
				"content":    str("评论内容"),
			}, "project_id", "type", "id", "content"),
		},
		{
			Name:        "create_cycle",
			Description: "创建迭代（Sprint）",
			Parameters: object(map[string]any{
				"project_id": num("项目 ID"),
				"name":       str("迭代名称"),
				"goal":       str("迭代目标"),
			}, "project_id", "name"),
		},
	}
}

// MCPTools 返回 MCP Server 侧的工具定义（与 WorkspaceTools 对齐，供 MCP 复用）。
func MCPTools() []ToolDefinition { return WorkspaceTools() }
