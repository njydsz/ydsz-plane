// Package workspaceview — 工作空间视图 HTTP handlers。
//
// 路由挂载：
//   /api/v1/workspaces/:workspace_id/views
package workspaceview

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/njydsz/ydsz-plane/internal/interfaces/middleware"
	"github.com/njydsz/ydsz-plane/pkg/errs"
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

// RegisterRead 注册只读路由。
func (h *Handler) RegisterRead(r *gin.RouterGroup) {
	g := r.Group("/views")
	g.GET("", h.list)
	g.GET("/:view_id", h.get)
}

// RegisterWrite 注册写入路由。
func (h *Handler) RegisterWrite(r *gin.RouterGroup) {
	g := r.Group("/views")
	g.POST("", h.create)
	g.PATCH("/:view_id", h.update)
	g.DELETE("/:view_id", h.delete)
	g.POST("/:view_id/duplicate", h.duplicate)
	g.POST("/:view_id/favorite", h.toggleFavorite)
	g.POST("/:view_id/public", h.togglePublic)
}

// Register 注册全部路由。
func (h *Handler) Register(r *gin.RouterGroup) {
	h.RegisterRead(r)
	h.RegisterWrite(r)
}

// --- helpers ---

func writeErr(c *gin.Context, err error) {
	var appErr *errs.AppError
	if errs.As(err, &appErr) {
		middleware.AbortWithError(c, appErr)
		return
	}
	switch err {
	case ErrViewNotFound:
		middleware.AbortWithError(c, errs.ErrNotFound)
	default:
		middleware.AbortWithError(c, errs.ErrInternal)
	}
}

func intParam(c *gin.Context, key string) (int64, error) {
	id, err := strconv.ParseInt(c.Param(key), 10, 64)
	if err != nil || id <= 0 {
		return 0, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: key, Reason: "无效的 ID"})
	}
	return id, nil
}

func intQuery(c *gin.Context, key string, def int) int {
	if v := c.Query(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
	}
	return def
}

func boolQuery(c *gin.Context, key string) bool {
	v := c.Query(key)
	return v == "true" || v == "1"
}

func fieldDetail(err error) errs.FieldDetail {
	return errs.FieldDetail{Field: "body", Reason: err.Error()}
}

// --- handlers ---

// list 列出工作空间下的视图。
func (h *Handler) list(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	userID := c.GetInt64(middleware.CtxUserID)

	opts := ListOptions{
		WorkspaceID:  wsID,
		CreatedBy:    0,
		PublicOnly:   boolQuery(c, "public"),
		FavoriteOnly: boolQuery(c, "favorite"),
		Limit:        intQuery(c, "limit", 100),
		Offset:       intQuery(c, "offset", 0),
	}
	// mine=1 时仅返回当前用户创建的视图
	if boolQuery(c, "mine") {
		opts.CreatedBy = userID
	}

	list, err := h.d.Svc.List(c.Request.Context(), opts)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"results": list})
}

// get 查询视图详情。
func (h *Handler) get(c *gin.Context) {
	id, err := intParam(c, "view_id")
	if err != nil {
		writeErr(c, err)
		return
	}
	v, err := h.d.Svc.Get(c.Request.Context(), id)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, v)
}

// create 创建视图。
func (h *Handler) create(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	userID := c.GetInt64(middleware.CtxUserID)

	var req struct {
		Name           string `json:"name"`
		Description    string `json:"description"`
		Layout         string `json:"layout"`
		Filters        json.RawMessage `json:"filters"`
		DisplayOptions json.RawMessage `json:"display_options"`
		Sort           json.RawMessage `json:"sort"`
		IsPublic       bool   `json:"is_public"`
		IsFavorite     bool   `json:"is_favorite"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		writeErr(c, errs.ErrValidation.WithDetails(fieldDetail(err)))
		return
	}

	v, err := h.d.Svc.Create(c.Request.Context(), CreateInput{
		WorkspaceID:    wsID,
		Name:           req.Name,
		Description:    req.Description,
		Layout:         Layout(req.Layout),
		Filters:        req.Filters,
		DisplayOptions: req.DisplayOptions,
		Sort:           req.Sort,
		IsPublic:       req.IsPublic,
		IsFavorite:     req.IsFavorite,
		CreatedBy:      userID,
	})
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, v)
}

// update 更新视图。
func (h *Handler) update(c *gin.Context) {
	id, err := intParam(c, "view_id")
	if err != nil {
		writeErr(c, err)
		return
	}

	var req struct {
		Name           *string         `json:"name"`
		Description    *string         `json:"description"`
		Layout         *string         `json:"layout"`
		Filters        *json.RawMessage `json:"filters"`
		DisplayOptions *json.RawMessage `json:"display_options"`
		Sort           *json.RawMessage `json:"sort"`
		IsPublic       *bool           `json:"is_public"`
		IsFavorite     *bool           `json:"is_favorite"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		writeErr(c, errs.ErrValidation.WithDetails(fieldDetail(err)))
		return
	}

	in := UpdateInput{
		Name:           req.Name,
		Description:    req.Description,
		Filters:        req.Filters,
		DisplayOptions: req.DisplayOptions,
		Sort:           req.Sort,
		IsPublic:       req.IsPublic,
		IsFavorite:     req.IsFavorite,
	}
	if req.Layout != nil {
		l := Layout(*req.Layout)
		in.Layout = &l
	}

	v, err := h.d.Svc.Update(c.Request.Context(), id, in)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, v)
}

// delete 删除视图。
func (h *Handler) delete(c *gin.Context) {
	id, err := intParam(c, "view_id")
	if err != nil {
		writeErr(c, err)
		return
	}
	if err := h.d.Svc.Delete(c.Request.Context(), id); err != nil {
		writeErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// duplicate 复制视图。
func (h *Handler) duplicate(c *gin.Context) {
	id, err := intParam(c, "view_id")
	if err != nil {
		writeErr(c, err)
		return
	}
	userID := c.GetInt64(middleware.CtxUserID)
	v, err := h.d.Svc.Duplicate(c.Request.Context(), id, userID)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, v)
}

// toggleFavorite 切换收藏状态。
func (h *Handler) toggleFavorite(c *gin.Context) {
	id, err := intParam(c, "view_id")
	if err != nil {
		writeErr(c, err)
		return
	}
	isFav, err := h.d.Svc.ToggleFavorite(c.Request.Context(), id)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"is_favorite": isFav})
}

// togglePublic 切换公开状态。
func (h *Handler) togglePublic(c *gin.Context) {
	id, err := intParam(c, "view_id")
	if err != nil {
		writeErr(c, err)
		return
	}
	isPublic, err := h.d.Svc.TogglePublic(c.Request.Context(), id)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"is_public": isPublic})
}
