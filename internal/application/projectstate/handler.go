// Package projectstate — 项目状态与更新 HTTP handlers。
package projectstate

import (
	"net/http"
	"strconv"
	"time"

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

// RegisterProject 注册项目级路由。
func (h *Handler) RegisterProject(r *gin.RouterGroup) {
	g := r.Group("/project-state")
	g.GET("", h.GetState)
	g.PUT("", h.UpsertState)

	u := r.Group("/project-updates")
	u.GET("", h.ListUpdates)
	u.POST("", h.CreateUpdate)
	u.DELETE("/:update_id", h.DeleteUpdate)
	u.POST("/:update_id/reactions", h.AddReaction)
	u.DELETE("/:update_id/reactions/:emoji", h.RemoveReaction)
	u.GET("/:update_id/reactions", h.ListReactions)
	u.GET("/:update_id/comments", h.ListComments)
	u.POST("/:update_id/comments", h.AddComment)
	u.DELETE("/comments/:comment_id", h.DeleteComment)
}

// RegisterWorkspace 注册工作空间级路由（按倡议查询更新）。
func (h *Handler) RegisterWorkspace(r *gin.RouterGroup) {
	r.Group("/initiative-updates").GET("", h.ListUpdates)
}

// --- helpers ---

func (h *Handler) fail(c *gin.Context, err error) {
	switch err {
	case ErrStateNotFound, ErrUpdateNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
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

func paramInt(c *gin.Context, key string) int64 {
	n, _ := strconv.ParseInt(c.Param(key), 10, 64)
	return n
}

// GetState godoc
//
//	@Summary		项目状态
//	@Tags			project-state
//	@Produce		json
//	@Success		200	{object}	ProjectState
//	@Router			/project-state [get]
func (h *Handler) GetState(c *gin.Context) {
	ps, err := h.d.Svc.GetState(c.Request.Context(), c.GetInt64(middleware.CtxProjectID))
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, ps)
}

// UpsertState godoc
//
//	@Summary		更新项目状态
//	@Tags			project-state
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	ProjectState
//	@Router			/project-state [put]
func (h *Handler) UpsertState(c *gin.Context) {
	var body struct {
		Status      string  `json:"status"`
		Priority    string  `json:"priority"`
		StartDate   *string `json:"start_date"`
		TargetDate  *string `json:"target_date"`
		OwnerID     int64   `json:"owner_id"`
		Description string  `json:"description"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ps, err := h.d.Svc.UpsertState(c.Request.Context(), UpsertStateInput{
		ProjectID:   c.GetInt64(middleware.CtxProjectID),
		Status:      body.Status,
		Priority:    body.Priority,
		StartDate:   parseDatePtr(body.StartDate),
		TargetDate:  parseDatePtr(body.TargetDate),
		OwnerID:     body.OwnerID,
		Description: body.Description,
	})
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, ps)
}

// ListUpdates godoc
//
//	@Summary		项目更新列表
//	@Tags			project-state
//	@Produce		json
//	@Success		200	{object}	map[string]any
//	@Router			/project-updates [get]
func (h *Handler) ListUpdates(c *gin.Context) {
	opts := ListUpdatesOptions{
		ProjectID:    c.GetInt64(middleware.CtxProjectID),
		InitiativeID: int64Query(c, "initiative_id"),
		Limit:        intQuery(c, "limit", 50),
		Offset:       intQuery(c, "offset", 0),
	}
	// 工作空间级路由没有项目上下文，允许通过 query 指定项目
	if opts.ProjectID == 0 {
		opts.ProjectID = int64Query(c, "project_id")
	}
	items, err := h.d.Svc.ListUpdates(c.Request.Context(), opts)
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

// CreateUpdate godoc
//
//	@Summary		发布项目更新
//	@Tags			project-state
//	@Accept			json
//	@Produce		json
//	@Success		201	{object}	ProjectUpdate
//	@Router			/project-updates [post]
func (h *Handler) CreateUpdate(c *gin.Context) {
	var body struct {
		Status          string  `json:"status"`
		Content         string  `json:"content"`
		ProgressPercent float64 `json:"progress_percent"`
		InitiativeID    int64   `json:"initiative_id"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	pu, err := h.d.Svc.CreateUpdate(c.Request.Context(), CreateUpdateInput{
		WorkspaceID:     c.GetInt64(middleware.CtxWorkspaceID),
		ProjectID:       c.GetInt64(middleware.CtxProjectID),
		AuthorID:        c.GetInt64(middleware.CtxUserID),
		Status:          body.Status,
		Content:         body.Content,
		ProgressPercent: body.ProgressPercent,
		InitiativeID:    body.InitiativeID,
	})
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusCreated, pu)
}

// DeleteUpdate godoc
//
//	@Summary		删除项目更新
//	@Tags			project-state
//	@Produce		json
//	@Success		204
//	@Router			/project-updates/{update_id} [delete]
func (h *Handler) DeleteUpdate(c *gin.Context) {
	if err := h.d.Svc.DeleteUpdate(c.Request.Context(), paramInt(c, "update_id")); err != nil {
		h.fail(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// AddReaction godoc
//
//	@Summary		添加更新表情回应
//	@Tags			project-state
//	@Accept			json
//	@Produce		json
//	@Success		204
//	@Router			/project-updates/{update_id}/reactions [post]
func (h *Handler) AddReaction(c *gin.Context) {
	var body struct {
		Emoji string `json:"emoji"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.d.Svc.AddReaction(c.Request.Context(), paramInt(c, "update_id"), c.GetInt64(middleware.CtxUserID), body.Emoji); err != nil {
		h.fail(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// RemoveReaction godoc
//
//	@Summary		移除更新表情回应
//	@Tags			project-state
//	@Produce		json
//	@Success		204
//	@Router			/project-updates/{update_id}/reactions/{emoji} [delete]
func (h *Handler) RemoveReaction(c *gin.Context) {
	if err := h.d.Svc.RemoveReaction(c.Request.Context(), paramInt(c, "update_id"), c.GetInt64(middleware.CtxUserID), c.Param("emoji")); err != nil {
		h.fail(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ListReactions godoc
//
//	@Summary		更新表情回应列表
//	@Tags			project-state
//	@Produce		json
//	@Success		200	{object}	map[string]any
//	@Router			/project-updates/{update_id}/reactions [get]
func (h *Handler) ListReactions(c *gin.Context) {
	items, err := h.d.Svc.ListReactions(c.Request.Context(), paramInt(c, "update_id"))
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

// ListComments godoc
//
//	@Summary		更新评论列表
//	@Tags			project-state
//	@Produce		json
//	@Success		200	{object}	map[string]any
//	@Router			/project-updates/{update_id}/comments [get]
func (h *Handler) ListComments(c *gin.Context) {
	items, err := h.d.Svc.ListComments(c.Request.Context(), paramInt(c, "update_id"))
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

// AddComment godoc
//
//	@Summary		添加更新评论
//	@Tags			project-state
//	@Accept			json
//	@Produce		json
//	@Success		201	{object}	ProjectUpdateComment
//	@Router			/project-updates/{update_id}/comments [post]
func (h *Handler) AddComment(c *gin.Context) {
	var body struct {
		Content  string `json:"content"`
		ParentID *int64 `json:"parent_id"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cm, err := h.d.Svc.AddComment(c.Request.Context(), paramInt(c, "update_id"), c.GetInt64(middleware.CtxUserID), body.Content, body.ParentID)
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusCreated, cm)
}

// DeleteComment godoc
//
//	@Summary		删除更新评论
//	@Tags			project-state
//	@Produce		json
//	@Success		204
//	@Router			/project-updates/comments/{comment_id} [delete]
func (h *Handler) DeleteComment(c *gin.Context) {
	if err := h.d.Svc.DeleteComment(c.Request.Context(), paramInt(c, "comment_id"), c.GetInt64(middleware.CtxUserID)); err != nil {
		h.fail(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// --- helpers ---

// parseDatePtr 将 RFC3339 字符串解析为 *time.Time。
func parseDatePtr(s *string) *time.Time {
	if s == nil || *s == "" {
		return nil
	}
	if t, err := time.Parse(time.RFC3339, *s); err == nil {
		return &t
	}
	return nil
}
