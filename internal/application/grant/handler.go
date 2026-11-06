// Package grant — HTTP handlers。
package grant

import (
	"encoding/json"
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
func NewHandler(d *HandlerDeps) *Handler { return &Handler{d: d} }

// Register 注册路由（工作空间级）。
func (h *Handler) Register(r *gin.RouterGroup) {
	g := r.Group("/grant-rules")
	g.GET("", h.List)
	g.POST("", h.Create)
	g.GET("/:rule_id", h.Get)
	g.PATCH("/:rule_id", h.Update)
	g.DELETE("/:rule_id", h.Delete)
	g.POST("/:rule_id/enable", h.Enable)
	g.POST("/:rule_id/disable", h.Disable)
	g.POST("/evaluate", h.Evaluate)
}

// --- helpers ---

func (h *Handler) fail(c *gin.Context, err error) {
	switch err {
	case ErrRuleNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case ErrInvalidCondition:
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
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

func paramInt(c *gin.Context, key string) int64 {
	n, _ := strconv.ParseInt(c.Param(key), 10, 64)
	return n
}

// List godoc
//
//	@Summary		条件授权规则列表
//	@Tags			grant
//	@Produce		json
//	@Success		200	{object}	map[string]any
//	@Router			/grant-rules [get]
func (h *Handler) List(c *gin.Context) {
	items, err := h.d.Svc.List(c.Request.Context(), c.GetInt64(middleware.CtxWorkspaceID))
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

// Create godoc
//
//	@Summary		创建条件授权规则
//	@Tags			grant
//	@Accept			json
//	@Produce		json
//	@Success		201	{object}	Rule
//	@Router			/grant-rules [post]
func (h *Handler) Create(c *gin.Context) {
	var body struct {
		Name        string          `json:"name"`
		Description string          `json:"description"`
		RoleSlugs   []string        `json:"role_slugs"`
		Permissions []string        `json:"permissions"`
		Scope       string          `json:"scope"`
		Condition   json.RawMessage `json:"condition"`
		Priority    int             `json:"priority"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	r, err := h.d.Svc.Create(c.Request.Context(), CreateRuleInput{
		WorkspaceID: c.GetInt64(middleware.CtxWorkspaceID),
		Name:        body.Name,
		Description: body.Description,
		RoleSlugs:   body.RoleSlugs,
		Permissions: body.Permissions,
		Scope:       body.Scope,
		Condition:   body.Condition,
		Priority:    body.Priority,
		CreatedBy:   c.GetInt64(middleware.CtxUserID),
	})
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusCreated, r)
}

// Get godoc
//
//	@Summary		规则详情
//	@Tags			grant
//	@Produce		json
//	@Success		200	{object}	Rule
//	@Router			/grant-rules/{rule_id} [get]
func (h *Handler) Get(c *gin.Context) {
	r, err := h.d.Svc.GetByID(c.Request.Context(), paramInt(c, "rule_id"), c.GetInt64(middleware.CtxWorkspaceID))
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, r)
}

// Update godoc
//
//	@Summary		更新规则
//	@Tags			grant
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	Rule
//	@Router			/grant-rules/{rule_id} [patch]
func (h *Handler) Update(c *gin.Context) {
	var body struct {
		Name        *string         `json:"name"`
		Description *string         `json:"description"`
		RoleSlugs   []string        `json:"role_slugs"`
		Permissions []string        `json:"permissions"`
		Scope       *string         `json:"scope"`
		Condition   json.RawMessage `json:"condition"`
		Enabled     *bool           `json:"enabled"`
		Priority    *int            `json:"priority"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	r, err := h.d.Svc.Update(c.Request.Context(), paramInt(c, "rule_id"), c.GetInt64(middleware.CtxWorkspaceID), UpdateRuleInput{
		Name:        body.Name,
		Description: body.Description,
		RoleSlugs:   body.RoleSlugs,
		Permissions: body.Permissions,
		Scope:       body.Scope,
		Condition:   body.Condition,
		Enabled:     body.Enabled,
		Priority:    body.Priority,
	})
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, r)
}

// Delete godoc
//
//	@Summary		删除规则
//	@Tags			grant
//	@Produce		json
//	@Success		204
//	@Router			/grant-rules/{rule_id} [delete]
func (h *Handler) Delete(c *gin.Context) {
	if err := h.d.Svc.Delete(c.Request.Context(), paramInt(c, "rule_id"), c.GetInt64(middleware.CtxWorkspaceID)); err != nil {
		h.fail(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// Enable godoc
//
//	@Summary		启用规则
//	@Tags			grant
//	@Produce		json
//	@Success		200	{object}	Rule
//	@Router			/grant-rules/{rule_id}/enable [post]
func (h *Handler) Enable(c *gin.Context) {
	t := true
	r, err := h.d.Svc.Update(c.Request.Context(), paramInt(c, "rule_id"), c.GetInt64(middleware.CtxWorkspaceID), UpdateRuleInput{Enabled: &t})
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, r)
}

// Disable godoc
//
//	@Summary		停用规则
//	@Tags			grant
//	@Produce		json
//	@Success		200	{object}	Rule
//	@Router			/grant-rules/{rule_id}/disable [post]
func (h *Handler) Disable(c *gin.Context) {
	f := false
	r, err := h.d.Svc.Update(c.Request.Context(), paramInt(c, "rule_id"), c.GetInt64(middleware.CtxWorkspaceID), UpdateRuleInput{Enabled: &f})
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, r)
}

// Evaluate godoc
//
//	@Summary		试算条件表达式（调试用）
//	@Tags			grant
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	map[string]any
//	@Router			/grant-rules/evaluate [post]
func (h *Handler) Evaluate(c *gin.Context) {
	var body struct {
		Condition json.RawMessage `json:"condition"`
		Context   Context         `json:"context"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := Validate(body.Condition); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"valid": false, "error": err.Error()})
		return
	}
	var expr Expr
	if err := json.Unmarshal(body.Condition, &expr); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"valid": false, "error": err.Error()})
		return
	}
	ok, err := Evaluate(&expr, body.Context)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"valid": true, "result": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"valid": true, "result": ok})
}
