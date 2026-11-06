// Package mcpserver — MCP (Model Context Protocol) 服务端。
//
// 对标 Plane 的原生 MCP Server：让 Claude / Cursor / ChatGPT / Codex 等
// AI 客户端通过标准 MCP 协议读写工作空间数据。
//
// 实现：
//   - JSON-RPC 2.0 over HTTP（Streamable HTTP transport）；
//   - 支持 initialize / tools/list / tools/call / resources/list / ping；
//   - 工具覆盖工作项 / 迭代 / 模块 / 项目 / 文档 / 标签 / 状态 / 成员 / 搜索等；
//   - 通过 API Token（ydz_）或 OAuth Token（oat_）鉴权。
//
// 路由：
//   POST /mcp            — JSON-RPC 主端点
//   GET  /mcp            — 能力发现（返回 server info + 协议版本）
//   GET  /mcp/tools      — 工具目录（便于调试）
package mcpserver

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// ProtocolVersion MCP 协议版本。
const ProtocolVersion = "2024-11-05"

// ServerInfo 服务端元信息。
var ServerInfo = map[string]any{
	"name":    "ydsz-plane",
	"version": "0.1.0",
	"title":   "Ydsz Plane MCP Server",
}

// JSONRPCRequest JSON-RPC 请求。
type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// JSONRPCError 错误对象。
type JSONRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// JSONRPCResponse JSON-RPC 响应。
type JSONRPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Result  any             `json:"result,omitempty"`
	Error   *JSONRPCError   `json:"error,omitempty"`
}

// ToolDefinition MCP 工具定义。
type ToolDefinition struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"inputSchema"`
}

// Handler MCP HTTP handler。
type Handler struct {
	// BackendURL 内部 API 基地址（用于转发 tools/call）
	BackendURL string
	// TokenValidator 校验 MCP 请求中的 token
	TokenValidator func(token string) (userID int64, ok bool)
}

// NewHandler 构造 handler。
func NewHandler(backendURL string, validator func(string) (int64, bool)) *Handler {
	return &Handler{BackendURL: backendURL, TokenValidator: validator}
}

// Register 注册 MCP 路由。
func (h *Handler) Register(r *gin.Engine) {
	r.GET("/mcp", h.Discover)
	r.GET("/mcp/tools", h.ToolCatalog)
	r.POST("/mcp", h.HandleRPC)
}

// Discover GET /mcp — 能力发现。
func (h *Handler) Discover(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"protocolVersion": ProtocolVersion,
		"serverInfo":      ServerInfo,
		"capabilities": gin.H{
			"tools":     gin.H{"listChanged": false},
			"resources": gin.H{"listChanged": false},
		},
	})
}

// ToolCatalog GET /mcp/tools — 工具目录。
func (h *Handler) ToolCatalog(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"tools": ToolDefinitions()})
}

// HandleRPC POST /mcp — JSON-RPC 主端点。
func (h *Handler) HandleRPC(c *gin.Context) {
	var req JSONRPCRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, jsonRPCErr(nil, -32700, "Parse error"))
		return
	}

	if req.Method != "initialize" && req.Method != "ping" {
		token := extractToken(c)
		if token == "" {
			c.JSON(http.StatusOK, jsonRPCErr(req.ID, -32001, "Missing authorization token"))
			return
		}
		userID, ok := h.TokenValidator(token)
		if !ok {
			c.JSON(http.StatusOK, jsonRPCErr(req.ID, -32002, "Invalid token"))
			return
		}
		c.Set("mcp_token", token)
		c.Set("mcp_user_id", userID)
	}

	switch req.Method {
	case "initialize":
		c.JSON(http.StatusOK, jsonRPCOK(req.ID, gin.H{
			"protocolVersion": ProtocolVersion,
			"serverInfo":      ServerInfo,
			"capabilities": gin.H{
				"tools":     gin.H{"listChanged": false},
				"resources": gin.H{"listChanged": false},
			},
		}))

	case "ping":
		c.JSON(http.StatusOK, jsonRPCOK(req.ID, gin.H{}))

	case "tools/list":
		c.JSON(http.StatusOK, jsonRPCOK(req.ID, gin.H{"tools": ToolDefinitions()}))

	case "tools/call":
		h.handleToolCall(c, &req)

	case "resources/list":
		c.JSON(http.StatusOK, jsonRPCOK(req.ID, gin.H{"resources": []any{}}))

	default:
		c.JSON(http.StatusOK, jsonRPCErr(req.ID, -32601, "Method not found: "+req.Method))
	}
}

// handleToolCall 执行工具调用。
func (h *Handler) handleToolCall(c *gin.Context, req *JSONRPCRequest) {
	var params struct {
		Name      string         `json:"name"`
		Arguments map[string]any `json:"arguments"`
	}
	if len(req.Params) > 0 {
		if err := json.Unmarshal(req.Params, &params); err != nil {
			c.JSON(http.StatusOK, jsonRPCErr(req.ID, -32602, "Invalid params"))
			return
		}
	}

	userID := c.GetInt64("mcp_user_id")
	token := c.GetString("mcp_token")

	result, err := Dispatch(c.Request.Context(), h.BackendURL, token, userID, params.Name, params.Arguments)
	if err != nil {
		c.JSON(http.StatusOK, jsonRPCOK(req.ID, gin.H{
			"content": []gin.H{{"type": "text", "text": "Error: " + err.Error()}},
			"isError": true,
		}))
		return
	}

	c.JSON(http.StatusOK, jsonRPCOK(req.ID, gin.H{
		"content": []gin.H{{"type": "text", "text": renderResult(result)}},
		"isError": false,
	}))
}

// --- helpers ---

func jsonRPCOK(id json.RawMessage, result any) JSONRPCResponse {
	return JSONRPCResponse{JSONRPC: "2.0", ID: id, Result: result}
}

func jsonRPCErr(id json.RawMessage, code int, msg string) JSONRPCResponse {
	return JSONRPCResponse{JSONRPC: "2.0", ID: id, Error: &JSONRPCError{Code: code, Message: msg}}
}

// extractToken 从 Authorization 头提取 token。
func extractToken(c *gin.Context) string {
	h := c.GetHeader("Authorization")
	if h == "" {
		return ""
	}
	if strings.HasPrefix(strings.ToLower(h), "bearer ") {
		return strings.TrimSpace(h[7:])
	}
	return strings.TrimSpace(h)
}

// renderResult 将工具结果渲染为文本。
func renderResult(v any) string {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Sprintf("%v", v)
	}
	return string(b)
}
