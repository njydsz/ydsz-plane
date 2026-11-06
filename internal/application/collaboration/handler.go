// Package collaboration — 实时协作 HTTP handlers。
//
// 路由：/api/v1/workspaces/:workspace_id/collaboration/*
// 说明：WebSocket 长连接复用现有 /ws/:workspace_id 通道，
// 本组 HTTP 接口负责会话生命周期与操作提交。
package collaboration

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/njydsz/ydsz-plane/internal/interfaces/middleware"
)

// HandlerDeps handler 依赖。
type HandlerDeps struct {
	Svc *Service
}

// Handler Gin handler 集合。
type Handler struct {
	d *HandlerDeps
}

// NewHandler 构造 handler。
func NewHandler(d *HandlerDeps) *Handler {
	return &Handler{d: d}
}

// Register 注册路由。
func (h *Handler) Register(r *gin.RouterGroup) {
	g := r.Group("/collaboration")
	g.POST("/sessions", h.JoinSession)
	g.DELETE("/sessions", h.LeaveSession)
	g.POST("/heartbeat", h.Heartbeat)
	g.GET("/sessions", h.ListSessions)
	g.GET("/version", h.GetDocumentVersion)
	g.POST("/operations", h.SubmitOperation)
}

// --- helpers ---

func (h *Handler) fail(c *gin.Context, err error) {
	switch err {
	case ErrSessionNotFound, ErrDocumentNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case ErrVersionConflict:
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	default:
		if ec, ok := err.(interface{ HTTPStatus() int }); ok {
			c.JSON(ec.HTTPStatus(), gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
	}
}

func intQuery(c *gin.Context, key string, def int) int {
	if v := c.Query(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
	}
	return def
}

func int64Query(c *gin.Context, key string) int64 {
	n, _ := strconv.ParseInt(c.Query(key), 10, 64)
	return n
}

// docTypeAndID 从 query 读取 document_type / document_id。
func docTypeAndID(c *gin.Context) (DocumentType, int64, error) {
	dt := DocumentType(c.Query("document_type"))
	if !dt.IsValid() {
		return "", 0, localError{field: "document_type", reason: "invalid value"}
	}
	id := int64Query(c, "document_id")
	if id <= 0 {
		return "", 0, localError{field: "document_id", reason: "required"}
	}
	return dt, id, nil
}

// JoinSession godoc
//
//	@Summary		加入文档编辑会话
//	@Tags			collaboration
//	@Accept			json
//	@Produce		json
//	@Success		201	{object}	CollaborationSession
//	@Router			/collaboration/sessions [post]
func (h *Handler) JoinSession(c *gin.Context) {
	var body struct {
		DocumentType string `json:"document_type"`
		DocumentID   int64  `json:"document_id"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	sess, err := h.d.Svc.JoinSession(c.Request.Context(), JoinSessionInput{
		WorkspaceID:  c.GetInt64(middleware.CtxWorkspaceID),
		DocumentType: DocumentType(body.DocumentType),
		DocumentID:   body.DocumentID,
		UserID:       c.GetInt64(middleware.CtxUserID),
		UserName:     userName(c),
	})
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusCreated, sess)
}

// LeaveSession godoc
//
//	@Summary		离开文档编辑会话
//	@Tags			collaboration
//	@Produce		json
//	@Success		204
//	@Router			/collaboration/sessions [delete]
func (h *Handler) LeaveSession(c *gin.Context) {
	dt, id, err := docTypeAndID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.d.Svc.LeaveSession(c.Request.Context(), c.GetInt64(middleware.CtxWorkspaceID), string(dt), id, c.GetInt64(middleware.CtxUserID)); err != nil {
		h.fail(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// Heartbeat godoc
//
//	@Summary		编辑心跳
//	@Tags			collaboration
//	@Accept			json
//	@Produce		json
//	@Success		204
//	@Router			/collaboration/heartbeat [post]
func (h *Handler) Heartbeat(c *gin.Context) {
	var body struct {
		DocumentType string `json:"document_type"`
		DocumentID   int64  `json:"document_id"`
		BaseVersion  int64  `json:"base_version"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.d.Svc.Heartbeat(c.Request.Context(), HeartbeatInput{
		WorkspaceID:  c.GetInt64(middleware.CtxWorkspaceID),
		DocumentType: DocumentType(body.DocumentType),
		DocumentID:   body.DocumentID,
		UserID:       c.GetInt64(middleware.CtxUserID),
		BaseVersion:  body.BaseVersion,
	}); err != nil {
		h.fail(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ListSessions godoc
//
//	@Summary		文档在线编辑者列表
//	@Tags			collaboration
//	@Produce		json
//	@Success		200	{object}	map[string]any
//	@Router			/collaboration/sessions [get]
func (h *Handler) ListSessions(c *gin.Context) {
	dt, id, err := docTypeAndID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	items, err := h.d.Svc.ListSessions(c.Request.Context(), ListSessionsOptions{
		WorkspaceID:  c.GetInt64(middleware.CtxWorkspaceID),
		DocumentType: dt,
		DocumentID:   id,
		Limit:        intQuery(c, "limit", 50),
	})
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

// GetDocumentVersion godoc
//
//	@Summary		文档当前版本
//	@Tags			collaboration
//	@Produce		json
//	@Success		200	{object}	map[string]int64
//	@Router			/collaboration/version [get]
func (h *Handler) GetDocumentVersion(c *gin.Context) {
	dt, id, err := docTypeAndID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	v, err := h.d.Svc.GetDocumentVersion(c.Request.Context(), dt, id)
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"version": v})
}

// SubmitOperation godoc
//
//	@Summary		提交协作操作
//	@Tags			collaboration
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	SubmitResult
//	@Router			/collaboration/operations [post]
func (h *Handler) SubmitOperation(c *gin.Context) {
	var body struct {
		DocumentType string `json:"document_type"`
		DocumentID   int64  `json:"document_id"`
		OpType       string `json:"op_type"`
		BaseVersion  int64  `json:"base_version"`
		Payload      string `json:"payload"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	res, err := h.d.Svc.SubmitOperation(c.Request.Context(), SubmitOperationInput{
		WorkspaceID:  c.GetInt64(middleware.CtxWorkspaceID),
		DocumentType: DocumentType(body.DocumentType),
		DocumentID:   body.DocumentID,
		UserID:       c.GetInt64(middleware.CtxUserID),
		UserName:     userName(c),
		OpType:       body.OpType,
		BaseVersion:  body.BaseVersion,
		Payload:      body.Payload,
	})
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

// --- helpers ---

// localError 本地校验错误（映射为 400）。
type localError struct {
	field  string
	reason string
}

func (e localError) Error() string { return e.field + ": " + e.reason }

// HTTPStatus 实现 errs.Code 的状态映射。
func (e localError) HTTPStatus() int { return http.StatusBadRequest }

// userName 从 context 读取用户显示名（可选）。
func userName(c *gin.Context) string {
	if v, ok := c.Get("user_name"); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}
