// Package workitemtype — 自定义工作项类型 HTTP handlers。
//
// 路由挂载于 /api/v1/workspaces/:workspace_id/projects/:project_id/work-item-types。
package workitemtype

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

// RegisterRead 注册只读路由。
func (h *Handler) RegisterRead(r *gin.RouterGroup) {
	g := r.Group("/work-item-types")
	g.GET("", h.List)
	g.GET("/:type_id", h.Get)
	g.GET("/:type_id/properties", h.ListProperties)
	g.GET("/:type_id/values/:work_item_id", h.GetCustomFieldValues)
}

// RegisterWrite 注册写路由。
func (h *Handler) RegisterWrite(r *gin.RouterGroup) {
	g := r.Group("/work-item-types")
	g.POST("", h.Create)
	g.PATCH("/:type_id", h.Update)
	g.DELETE("/:type_id", h.Delete)
	g.POST("/:type_id/archive", h.Archive)
	g.POST("/:type_id/restore", h.Restore)
	g.POST("/:type_id/properties", h.AddProperty)
	g.DELETE("/:type_id/properties/:prop_id", h.RemoveProperty)
	g.PUT("/:type_id/values", h.SetCustomFieldValue)
}

// Register 注册全部路由。
func (h *Handler) Register(r *gin.RouterGroup) {
	h.RegisterRead(r)
	h.RegisterWrite(r)
}

// --- helpers ---

func (h *Handler) fail(c *gin.Context, err error) {
	switch err {
	case ErrTypeNotFound, ErrPropertyNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case ErrTypeCodeConflict, ErrPropertyCodeConflict:
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case ErrTypeArchived, ErrInvalidFieldType:
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

// --- handlers ---

// List godoc
//
//	@Summary		工作项类型列表
//	@Tags			work-item-type
//	@Produce		json
//	@Success		200	{object}	map[string]any
//	@Router			/work-item-types [get]
func (h *Handler) List(c *gin.Context) {
	items, err := h.d.Svc.List(c.Request.Context(), ListTypesOptions{
		ProjectID: c.GetInt64(middleware.CtxProjectID),
		Limit:     intQuery(c, "limit", 100),
		Offset:    intQuery(c, "offset", 0),
	})
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

// Get godoc
//
//	@Summary		工作项类型详情
//	@Tags			work-item-type
//	@Produce		json
//	@Success		200	{object}	WorkItemType
//	@Router			/work-item-types/{type_id} [get]
func (h *Handler) Get(c *gin.Context) {
	t, err := h.d.Svc.GetByID(c.Request.Context(), paramInt(c, "type_id"))
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, t)
}

// Create godoc
//
//	@Summary		创建工作项类型
//	@Tags			work-item-type
//	@Accept			json
//	@Produce		json
//	@Success		201	{object}	WorkItemType
//	@Router			/work-item-types [post]
func (h *Handler) Create(c *gin.Context) {
	var body struct {
		Name        string `json:"name"`
		Code        string `json:"code"`
		Color       string `json:"color"`
		Icon        string `json:"icon"`
		Description string `json:"description"`
		Level       int    `json:"level"`
		IsDefault   bool   `json:"is_default"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	t, err := h.d.Svc.Create(c.Request.Context(), CreateTypeInput{
		ProjectID:   c.GetInt64(middleware.CtxProjectID),
		Name:        body.Name,
		Code:        body.Code,
		Color:       body.Color,
		Icon:        body.Icon,
		Description: body.Description,
		Level:       body.Level,
		IsDefault:   body.IsDefault,
	})
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusCreated, t)
}

// Update godoc
//
//	@Summary		更新工作项类型
//	@Tags			work-item-type
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	WorkItemType
//	@Router			/work-item-types/{type_id} [patch]
func (h *Handler) Update(c *gin.Context) {
	var body struct {
		Name        *string `json:"name"`
		Code        *string `json:"code"`
		Color       *string `json:"color"`
		Icon        *string `json:"icon"`
		Description *string `json:"description"`
		Level       *int    `json:"level"`
		IsActive    *bool   `json:"is_active"`
		IsDefault   *bool   `json:"is_default"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	t, err := h.d.Svc.Update(c.Request.Context(), paramInt(c, "type_id"), UpdateTypeInput{
		Name:        body.Name,
		Code:        body.Code,
		Color:       body.Color,
		Icon:        body.Icon,
		Description: body.Description,
		Level:       body.Level,
		IsActive:    body.IsActive,
		IsDefault:   body.IsDefault,
	})
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, t)
}

// Delete godoc
//
//	@Summary		删除工作项类型
//	@Tags			work-item-type
//	@Produce		json
//	@Success		204
//	@Router			/work-item-types/{type_id} [delete]
func (h *Handler) Delete(c *gin.Context) {
	if err := h.d.Svc.Delete(c.Request.Context(), paramInt(c, "type_id")); err != nil {
		h.fail(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// Archive godoc
//
//	@Summary		归档工作项类型
//	@Tags			work-item-type
//	@Produce		json
//	@Success		200	{object}	WorkItemType
//	@Router			/work-item-types/{type_id}/archive [post]
func (h *Handler) Archive(c *gin.Context) {
	id := paramInt(c, "type_id")
	if err := h.d.Svc.Archive(c.Request.Context(), id); err != nil {
		h.fail(c, err)
		return
	}
	t, _ := h.d.Svc.GetByID(c.Request.Context(), id)
	c.JSON(http.StatusOK, t)
}

// Restore godoc
//
//	@Summary		恢复工作项类型
//	@Tags			work-item-type
//	@Produce		json
//	@Success		200	{object}	WorkItemType
//	@Router			/work-item-types/{type_id}/restore [post]
func (h *Handler) Restore(c *gin.Context) {
	id := paramInt(c, "type_id")
	if err := h.d.Svc.Restore(c.Request.Context(), id); err != nil {
		h.fail(c, err)
		return
	}
	t, _ := h.d.Svc.GetByID(c.Request.Context(), id)
	c.JSON(http.StatusOK, t)
}

// AddProperty godoc
//
//	@Summary		新增自定义属性
//	@Tags			work-item-type
//	@Accept			json
//	@Produce		json
//	@Success		201	{object}	WorkItemTypeProperty
//	@Router			/work-item-types/{type_id}/properties [post]
func (h *Handler) AddProperty(c *gin.Context) {
	var body struct {
		Name      string `json:"name"`
		Code      string `json:"code"`
		FieldType string `json:"field_type"`
		Options   string `json:"options"`
		Required  bool   `json:"required"`
		SortOrder int    `json:"sort_order"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	p, err := h.d.Svc.AddProperty(c.Request.Context(), CreatePropertyInput{
		TypeID:    paramInt(c, "type_id"),
		Name:      body.Name,
		Code:      body.Code,
		FieldType: body.FieldType,
		Options:   body.Options,
		Required:  body.Required,
		SortOrder: body.SortOrder,
	})
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusCreated, p)
}

// ListProperties godoc
//
//	@Summary		自定义属性列表
//	@Tags			work-item-type
//	@Produce		json
//	@Success		200	{object}	map[string]any
//	@Router			/work-item-types/{type_id}/properties [get]
func (h *Handler) ListProperties(c *gin.Context) {
	items, err := h.d.Svc.ListProperties(c.Request.Context(), paramInt(c, "type_id"))
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

// RemoveProperty godoc
//
//	@Summary		删除自定义属性
//	@Tags			work-item-type
//	@Produce		json
//	@Success		204
//	@Router			/work-item-types/{type_id}/properties/{prop_id} [delete]
func (h *Handler) RemoveProperty(c *gin.Context) {
	if err := h.d.Svc.RemoveProperty(c.Request.Context(),
		paramInt(c, "type_id"), paramInt(c, "prop_id")); err != nil {
		h.fail(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// GetCustomFieldValues godoc
//
//	@Summary		工作项自定义属性值列表
//	@Tags			work-item-type
//	@Produce		json
//	@Success		200	{object}	map[string]any
//	@Router			/work-item-types/{type_id}/values/{work_item_id} [get]
func (h *Handler) GetCustomFieldValues(c *gin.Context) {
	items, err := h.d.Svc.GetCustomFieldValues(c.Request.Context(), paramInt(c, "work_item_id"))
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

// SetCustomFieldValue godoc
//
//	@Summary		写入工作项自定义属性值
//	@Tags			work-item-type
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	WorkItemCustomFieldValue
//	@Router			/work-item-types/{type_id}/values [put]
func (h *Handler) SetCustomFieldValue(c *gin.Context) {
	var body struct {
		PropertyID   int64  `json:"property_id"`
		WorkItemID   int64  `json:"work_item_id"`
		WorkItemType string `json:"work_item_type"`
		Value        string `json:"value"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	v, err := h.d.Svc.SetCustomFieldValue(c.Request.Context(), SetCustomFieldValueInput{
		PropertyID:   body.PropertyID,
		WorkItemID:   body.WorkItemID,
		WorkItemType: body.WorkItemType,
		Value:        body.Value,
	})
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, v)
}
