// Package mcpserver — 工具目录与分发。
package mcpserver

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// ToolDefinitions 返回 MCP 工具目录。
//
// 共 30 个工具，覆盖：
//   工作项 / 工作项类型 / 迭代 / 模块 / 里程碑 / 版本 / 标签
//   项目 / 工作空间 / 状态 / 成员 / 文档 / 集合 / 功能
//   需求受理 / 模板 / 客户 / 查询参考
func ToolDefinitions() []ToolDefinition {
	stringProp := func(desc string) map[string]any {
		return map[string]any{"type": "string", "description": desc}
	}
	numberProp := func(desc string) map[string]any {
		return map[string]any{"type": "number", "description": desc}
	}
	obj := func(props map[string]any, required ...string) map[string]any {
		return map[string]any{
			"type":       "object",
			"properties": props,
			"required":   required,
		}
	}

	return []ToolDefinition{
		// --- 工作项 ---
		{Name: "list_work_items", Description: "列出工作项（需求/任务/缺陷）", InputSchema: obj(map[string]any{
			"project_id": numberProp("项目 ID"),
			"type":       stringProp("类型过滤：requirement/task/defect"),
			"state":      stringProp("状态名称过滤"),
			"assignee":   stringProp("指派人（用户名或 me）"),
			"limit":      numberProp("返回数量上限"),
			"offset":     numberProp("偏移量"),
		})},
		{Name: "get_work_item", Description: "获取工作项详情", InputSchema: obj(map[string]any{
			"project_id": numberProp("项目 ID"),
			"type":       stringProp("类型：requirement/task/defect"),
			"id":         numberProp("工作项 ID"),
		}, "project_id", "type", "id")},
		{Name: "create_work_item", Description: "创建工作项", InputSchema: obj(map[string]any{
			"project_id": numberProp("项目 ID"),
			"type":       stringProp("类型：requirement/task/defect"),
			"title":      stringProp("标题"),
			"description": stringProp("描述"),
			"priority":   stringProp("优先级：urgent/high/medium/low/none"),
			"assignee_id": numberProp("指派人 ID"),
		}, "project_id", "type", "title")},
		{Name: "update_work_item", Description: "更新工作项字段", InputSchema: obj(map[string]any{
			"project_id": numberProp("项目 ID"),
			"type":       stringProp("类型：requirement/task/defect"),
			"id":         numberProp("工作项 ID"),
			"title":      stringProp("新标题"),
			"description": stringProp("新描述"),
			"priority":   stringProp("新优先级"),
			"state_id":   numberProp("新状态 ID"),
		}, "project_id", "type", "id")},
		{Name: "delete_work_item", Description: "删除工作项", InputSchema: obj(map[string]any{
			"project_id": numberProp("项目 ID"),
			"type":       stringProp("类型：requirement/task/defect"),
			"id":         numberProp("工作项 ID"),
		}, "project_id", "type", "id")},
		{Name: "transition_work_item", Description: "流转工作项状态", InputSchema: obj(map[string]any{
			"project_id": numberProp("项目 ID"),
			"type":       stringProp("类型"),
			"id":         numberProp("工作项 ID"),
			"state_id":   numberProp("目标状态 ID"),
		}, "project_id", "type", "id", "state_id")},
		{Name: "assign_work_item", Description: "指派工作项", InputSchema: obj(map[string]any{
			"project_id": numberProp("项目 ID"),
			"type":       stringProp("类型"),
			"id":         numberProp("工作项 ID"),
			"user_id":    numberProp("被指派用户 ID"),
		}, "project_id", "type", "id", "user_id")},
		{Name: "add_comment", Description: "给工作项添加评论", InputSchema: obj(map[string]any{
			"project_id": numberProp("项目 ID"),
			"type":       stringProp("类型"),
			"id":         numberProp("工作项 ID"),
			"content":    stringProp("评论内容"),
		}, "project_id", "type", "id", "content")},
		{Name: "list_work_item_activities", Description: "工作项活动日志", InputSchema: obj(map[string]any{
			"project_id": numberProp("项目 ID"),
			"type":       stringProp("类型"),
			"id":         numberProp("工作项 ID"),
		}, "project_id", "type", "id")},

		// --- 迭代 ---
		{Name: "list_cycles", Description: "列出迭代（Sprint）", InputSchema: obj(map[string]any{
			"project_id": numberProp("项目 ID"),
			"status":     stringProp("状态：planned/active/completed"),
		})},
		{Name: "get_cycle", Description: "迭代详情", InputSchema: obj(map[string]any{
			"project_id": numberProp("项目 ID"),
			"id":         numberProp("迭代 ID"),
		}, "project_id", "id")},
		{Name: "create_cycle", Description: "创建迭代", InputSchema: obj(map[string]any{
			"project_id": numberProp("项目 ID"),
			"name":       stringProp("名称"),
			"goal":       stringProp("目标"),
		}, "project_id", "name")},
		{Name: "start_cycle", Description: "启动迭代", InputSchema: obj(map[string]any{
			"project_id": numberProp("项目 ID"),
			"id":         numberProp("迭代 ID"),
		}, "project_id", "id")},
		{Name: "complete_cycle", Description: "完成迭代", InputSchema: obj(map[string]any{
			"project_id": numberProp("项目 ID"),
			"id":         numberProp("迭代 ID"),
		}, "project_id", "id")},
		{Name: "cycle_analytics", Description: "迭代分析（燃尽/速度）", InputSchema: obj(map[string]any{
			"project_id": numberProp("项目 ID"),
			"id":         numberProp("迭代 ID"),
		}, "project_id", "id")},

		// --- 模块 ---
		{Name: "list_modules", Description: "列出模块", InputSchema: obj(map[string]any{
			"project_id": numberProp("项目 ID"),
		})},
		{Name: "create_module", Description: "创建模块", InputSchema: obj(map[string]any{
			"project_id": numberProp("项目 ID"),
			"name":       stringProp("名称"),
			"description": stringProp("描述"),
		}, "project_id", "name")},
		{Name: "update_module", Description: "更新模块", InputSchema: obj(map[string]any{
			"project_id": numberProp("项目 ID"),
			"id":         numberProp("模块 ID"),
			"name":       stringProp("新名称"),
		}, "project_id", "id")},
		{Name: "delete_module", Description: "删除模块", InputSchema: obj(map[string]any{
			"project_id": numberProp("项目 ID"),
			"id":         numberProp("模块 ID"),
		}, "project_id", "id")},

		// --- 版本 ---
		{Name: "list_versions", Description: "列出版本", InputSchema: obj(map[string]any{
			"workspace_id": numberProp("工作空间 ID"),
		})},
		{Name: "create_version", Description: "创建版本", InputSchema: obj(map[string]any{
			"workspace_id": numberProp("工作空间 ID"),
			"name":         stringProp("版本号（semver）"),
			"description":  stringProp("描述"),
		}, "workspace_id", "name")},
		{Name: "release_version", Description: "发布版本", InputSchema: obj(map[string]any{
			"workspace_id": numberProp("工作空间 ID"),
			"id":           numberProp("版本 ID"),
		}, "workspace_id", "id")},

		// --- 项目 / 工作空间 ---
		{Name: "list_projects", Description: "列出项目", InputSchema: obj(map[string]any{})},
		{Name: "get_project", Description: "项目详情", InputSchema: obj(map[string]any{
			"project_id": numberProp("项目 ID"),
		}, "project_id")},
		{Name: "list_workspace_members", Description: "工作空间成员列表", InputSchema: obj(map[string]any{})},

		// --- 状态 / 标签 ---
		{Name: "list_states", Description: "项目状态列表", InputSchema: obj(map[string]any{
			"project_id": numberProp("项目 ID"),
		})},
		{Name: "list_labels", Description: "项目标签列表", InputSchema: obj(map[string]any{
			"project_id": numberProp("项目 ID"),
		})},

		// --- 文档 ---
		{Name: "list_pages", Description: "项目文档列表", InputSchema: obj(map[string]any{
			"project_id": numberProp("项目 ID"),
		})},
		{Name: "get_page", Description: "文档详情", InputSchema: obj(map[string]any{
			"project_id": numberProp("项目 ID"),
			"id":         numberProp("文档 ID"),
		}, "project_id", "id")},
		{Name: "create_page", Description: "创建文档", InputSchema: obj(map[string]any{
			"project_id": numberProp("项目 ID"),
			"title":      stringProp("标题"),
			"content":    stringProp("内容（HTML）"),
		}, "project_id", "title")},

		// --- 搜索 / 受理 / 客户 ---
		{Name: "search", Description: "全局搜索（跨项目）", InputSchema: obj(map[string]any{
			"query": stringProp("搜索关键词"),
			"limit": numberProp("返回数量"),
		}, "query")},
		{Name: "list_intake", Description: "需求受理列表", InputSchema: obj(map[string]any{
			"status": stringProp("状态过滤"),
		})},
		{Name: "list_customers", Description: "客户列表", InputSchema: obj(map[string]any{})},

		// --- 参考 ---
		{Name: "pql_reference", Description: "PQL 查询语言参考（字段/操作符/示例）", InputSchema: obj(map[string]any{})},
	}
}

// --- 分发 ---

// httpClient 共享客户端。
var httpClient = &http.Client{Timeout: 20 * time.Second}

// Dispatch 将工具调用转发到内部 API。
func Dispatch(ctx context.Context, backendURL, token string, userID int64, tool string, args map[string]any) (any, error) {
	switch tool {
	case "pql_reference":
		return pqlReference(), nil
	}

	path, method := routeOf(tool, args)
	if path == "" {
		return nil, fmt.Errorf("unknown tool: %s", tool)
	}

	url := fmt.Sprintf("%s/api/v1/workspaces/%d%s", backendURL, workspaceIDFromArgs(args), path)

	var body io.Reader
	if method == http.MethodPost || method == http.MethodPut || method == http.MethodPatch {
		b, _ := json.Marshal(args)
		body = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-MCP-User-ID", fmt.Sprintf("%d", userID))

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("backend returned %d: %s", resp.StatusCode, string(data))
	}

	var out any
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// routeOf 将工具名映射为内部 API 路径。
func routeOf(tool string, args map[string]any) (path, method string) {
	projectID := int64FromArgs(args, "project_id")
	wsID := int64FromArgs(args, "workspace_id")

	switch tool {
	case "list_work_items":
		return fmt.Sprintf("/projects/%d/issues?type=%v&limit=%v&offset=%v",
			projectID, strArg(args, "type"), numArg(args, "limit", 100), numArg(args, "offset", 0)), http.MethodGet
	case "get_work_item":
		return fmt.Sprintf("/projects/%d/issues/%d", projectID, int64FromArgs(args, "id")), http.MethodGet
	case "create_work_item":
		return fmt.Sprintf("/projects/%d/issues", projectID), http.MethodPost
	case "update_work_item":
		return fmt.Sprintf("/projects/%d/issues/%d", projectID, int64FromArgs(args, "id")), http.MethodPatch
	case "delete_work_item":
		return fmt.Sprintf("/projects/%d/issues/%d", projectID, int64FromArgs(args, "id")), http.MethodDelete
	case "transition_work_item":
		return fmt.Sprintf("/projects/%d/issues/%d/transition", projectID, int64FromArgs(args, "id")), http.MethodPost
	case "assign_work_item":
		return fmt.Sprintf("/projects/%d/issues/%d/assignees", projectID, int64FromArgs(args, "id")), http.MethodPut
	case "add_comment":
		return fmt.Sprintf("/projects/%d/issues/%d/comments", projectID, int64FromArgs(args, "id")), http.MethodPost
	case "list_work_item_activities":
		return fmt.Sprintf("/projects/%d/issues/%d/activities", projectID, int64FromArgs(args, "id")), http.MethodGet

	case "list_cycles":
		return fmt.Sprintf("/projects/%d/sprints", projectID), http.MethodGet
	case "get_cycle":
		return fmt.Sprintf("/projects/%d/sprints/%d", projectID, int64FromArgs(args, "id")), http.MethodGet
	case "create_cycle":
		return fmt.Sprintf("/projects/%d/sprints", projectID), http.MethodPost
	case "start_cycle":
		return fmt.Sprintf("/projects/%d/sprints/%d/start", projectID, int64FromArgs(args, "id")), http.MethodPost
	case "complete_cycle":
		return fmt.Sprintf("/projects/%d/sprints/%d/complete", projectID, int64FromArgs(args, "id")), http.MethodPost
	case "cycle_analytics":
		return fmt.Sprintf("/projects/%d/sprints/%d/analytics", projectID, int64FromArgs(args, "id")), http.MethodGet

	case "list_modules":
		return fmt.Sprintf("/projects/%d/modules", projectID), http.MethodGet
	case "create_module":
		return fmt.Sprintf("/projects/%d/modules", projectID), http.MethodPost
	case "update_module":
		return fmt.Sprintf("/projects/%d/modules/%d", projectID, int64FromArgs(args, "id")), http.MethodPatch
	case "delete_module":
		return fmt.Sprintf("/projects/%d/modules/%d", projectID, int64FromArgs(args, "id")), http.MethodDelete

	case "list_versions":
		return "/versions", http.MethodGet
	case "create_version":
		return "/versions", http.MethodPost
	case "release_version":
		return fmt.Sprintf("/versions/%d/release", int64FromArgs(args, "id")), http.MethodPost

	case "list_projects":
		return "/projects", http.MethodGet
	case "get_project":
		return fmt.Sprintf("/projects/%d", projectID), http.MethodGet
	case "list_workspace_members":
		return "/members", http.MethodGet

	case "list_states":
		return fmt.Sprintf("/projects/%d/states", projectID), http.MethodGet
	case "list_labels":
		return fmt.Sprintf("/projects/%d/labels", projectID), http.MethodGet

	case "list_pages":
		return fmt.Sprintf("/projects/%d/pages", projectID), http.MethodGet
	case "get_page":
		return fmt.Sprintf("/projects/%d/pages/%d", projectID, int64FromArgs(args, "id")), http.MethodGet
	case "create_page":
		return fmt.Sprintf("/projects/%d/pages", projectID), http.MethodPost

	case "search":
		return fmt.Sprintf("/search?q=%s&limit=%v", strArg(args, "query"), numArg(args, "limit", 20)), http.MethodGet
	case "list_intake":
		return "/intake/tickets", http.MethodGet
	case "list_customers":
		return "/customers", http.MethodGet
	}
	_ = wsID
	return "", ""
}

// workspaceIDFromArgs 从参数中取 workspace_id；
// 项目级工具需先通过 project → workspace 映射，此处由调用方在参数中传入。
func workspaceIDFromArgs(args map[string]any) int64 {
	if _, ok := args["workspace_id"]; ok {
		return int64FromArgs(args, "workspace_id")
	}
	return 0
}

func int64FromArgs(args map[string]any, key string) int64 {
	if v, ok := args[key]; ok {
		switch n := v.(type) {
		case float64:
			return int64(n)
		case int64:
			return n
		case json.Number:
			i, _ := n.Int64()
			return i
		}
	}
	return 0
}

func strArg(args map[string]any, key string) string {
	if v, ok := args[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func numArg(args map[string]any, key string, def int) int {
	if v, ok := args[key]; ok {
		if f, ok := v.(float64); ok {
			return int(f)
		}
	}
	return def
}

// pqlReference 返回 PQL 参考文档。
func pqlReference() map[string]any {
	return map[string]any{
		"description": "Plane Query Language (PQL) 参考",
		"operators":   []string{"=", "!=", ">", ">=", "<", "<=", "~ (包含)", "in (...)"},
		"logic":       []string{"AND", "OR", "NOT"},
		"special":     []string{"me (当前用户)", "@username (指定用户)"},
		"fields": []string{
			"state, state_group, priority, type, title, description",
			"assignee, creator, estimate, label, module, sprint",
			"created_at, updated_at, target_date, start_date, completed_at, severity",
		},
		"clauses": []string{"sort:<field> [asc|desc]", "limit:<n>", "offset:<n>"},
		"examples": []string{
			`state_group = "started" AND priority = "high"`,
			`assignee = me AND target_date <= "2026-12-31"`,
			`title ~ "登录" OR description ~ "鉴权"`,
			`label in ("backend", "urgent") AND NOT state_group = "cancelled"`,
			`type = "defect" AND severity >= 4 sort:created_at desc limit:20`,
		},
	}
}
