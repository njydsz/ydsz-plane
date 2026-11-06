// Package iam — HTTP handlers。
package iam

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/njydsz/ydsz-plane/internal/interfaces/middleware"
)

// HandlerDeps handler 依赖。
type HandlerDeps struct {
	Svc *AuditService
}

// Handler Gin handler 集合。
type Handler struct {
	d *HandlerDeps
}

// NewHandler 构造 handler。
func NewHandler(d *HandlerDeps) *Handler { return &Handler{d: d} }

// Register 注册路由（工作空间级）。
func (h *Handler) Register(r *gin.RouterGroup) {
	g := r.Group("/permission-audit")
	g.GET("", h.List)
	g.GET("/:id", h.Get)
}

// --- helpers ---

func intQuery(c *gin.Context, key string, def int) int {
	if v := c.Query(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
	}
	return def
}

func intParam(c *gin.Context, key string) int64 {
	n, _ := strconv.ParseInt(c.Param(key), 10, 64)
	return n
}

// List godoc
//
//	@Summary		权限审计日志
//	@Tags			iam
//	@Produce		json
//	@Success		200	{object}	map[string]any
//	@Router			/permission-audit [get]
func (h *Handler) List(c *gin.Context) {
	items, err := h.d.Svc.List(c.Request.Context(), AuditQuery{
		WorkspaceID: c.GetInt64(middleware.CtxWorkspaceID),
		RoleSlug:    c.Query("role_slug"),
		ActorID:     int64(intQuery(c, "actor_id", 0)),
		Action:      c.Query("action"),
		Limit:       intQuery(c, "limit", 100),
		Offset:      intQuery(c, "offset", 0),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

// Get godoc
//
//	@Summary		审计日志详情
//	@Tags			iam
//	@Produce		json
//	@Success		200	{object}	AuditEntry
//	@Router			/permission-audit/{id} [get]
func (h *Handler) Get(c *gin.Context) {
	entry, err := h.d.Svc.GetByID(c.Request.Context(), intParam(c, "id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, entry)
}
