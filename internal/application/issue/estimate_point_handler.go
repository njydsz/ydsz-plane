// Package issue — Estimate Point HTTP handlers（REST API）。
//
// 对齐 labels / modules 的路由模式：
//   - 列表 + 创建 挂在 /estimate-points
//   - 单资源 挂在 /estimate-points/:point_id
package issue

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/njydsz/ydsz-plane/internal/interfaces/middleware"
	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// EstimatePointHandler Gin handler 集合（估算点数 CRUD）。
type EstimatePointHandler struct {
	svc *EstimatePointService
}

// NewEstimatePointHandler 构造 handler。
func NewEstimatePointHandler(svc *EstimatePointService) *EstimatePointHandler {
	return &EstimatePointHandler{svc: svc}
}

// Register 注册估算点数路由。
func (h *EstimatePointHandler) Register(r *gin.RouterGroup) {
	eps := r.Group("/estimate-points")
	{
		eps.GET("", h.list)
		eps.POST("", h.create)
	}
	ep := eps.Group("/:point_id")
	{
		ep.GET("", h.get)
		ep.PATCH("", h.update)
		ep.DELETE("", h.delete)
	}
}

// ---- request DTOs ----

type createEstimatePointRequest struct {
	Name        string          `json:"name" binding:"required,max=120"`
	Description string          `json:"description"`
	Points      json.RawMessage `json:"points"`
	IsDefault   bool            `json:"is_default"`
}

type updateEstimatePointRequest struct {
	Name        *string         `json:"name" binding:"omitempty,max=120"`
	Description *string         `json:"description"`
	Points      json.RawMessage `json:"points"`
	IsDefault   *bool           `json:"is_default"`
}

// ---- handlers ----

// list 列出估算点数。
func (h *EstimatePointHandler) list(c *gin.Context) {
	items, err := h.svc.List(c.Request.Context(), ListFilter{
		WorkspaceID: extractWsID(c),
		ProjectID:   extractProjectID(c),
		Status:      c.Query("status"),
	})
	if err != nil {
		handleEstimatePointErr(c, err)
		return
	}
	if items == nil {
		items = []EstimatePoint{}
	}
	c.JSON(http.StatusOK, gin.H{"results": items})
}

// create 创建估算点数。
func (h *EstimatePointHandler) create(c *gin.Context) {
	var req createEstimatePointRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handleEstimatePointErr(c, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "body", Reason: "请求体无效: " + err.Error()}))
		return
	}
	ep, err := h.svc.Create(c.Request.Context(), CreateInput{
		WorkspaceID: extractWsID(c),
		ProjectID:   extractProjectID(c),
		Name:        req.Name,
		Description: req.Description,
		Points:      req.Points,
		IsDefault:   req.IsDefault,
		CreatedBy:   extractActorID(c),
	})
	if err != nil {
		handleEstimatePointErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, ep)
}

// get 获取单个估算点数。
func (h *EstimatePointHandler) get(c *gin.Context) {
	id, err := parseIDParam(c, "point_id")
	if err != nil {
		handleEstimatePointErr(c, err)
		return
	}
	ep, err := h.svc.Get(c.Request.Context(), extractWsID(c), id)
	if err != nil {
		handleEstimatePointErr(c, err)
		return
	}
	c.JSON(http.StatusOK, ep)
}

// update 更新估算点数。
func (h *EstimatePointHandler) update(c *gin.Context) {
	id, err := parseIDParam(c, "point_id")
	if err != nil {
		handleEstimatePointErr(c, err)
		return
	}
	var req updateEstimatePointRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handleEstimatePointErr(c, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "body", Reason: "请求体无效: " + err.Error()}))
		return
	}
	// 判断是否传入 points 字段
	hasPoints := len(req.Points) > 0 && string(req.Points) != "null"
	ep, err := h.svc.Update(c.Request.Context(), UpdateInput{
		ID:          id,
		WorkspaceID: extractWsID(c),
		ProjectID:   extractProjectID(c),
		Name:        req.Name,
		Description: req.Description,
		Points:      req.Points,
		HasPoints:   hasPoints,
		IsDefault:   req.IsDefault,
	})
	if err != nil {
		handleEstimatePointErr(c, err)
		return
	}
	c.JSON(http.StatusOK, ep)
}

// delete 删除估算点数。
func (h *EstimatePointHandler) delete(c *gin.Context) {
	id, err := parseIDParam(c, "point_id")
	if err != nil {
		handleEstimatePointErr(c, err)
		return
	}
	if err := h.svc.Delete(c.Request.Context(), extractWsID(c), id); err != nil {
		handleEstimatePointErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ---- helpers ----

// handleEstimatePointErr 统一错误响应。
func handleEstimatePointErr(c *gin.Context, err error) {
	if err == nil {
		return
	}
	var appErr *errs.AppError
	if errs.As(err, &appErr) {
		middleware.AbortWithError(c, appErr)
		return
	}
	middleware.AbortWithError(c, errs.ErrInternal)
}

// parseIDParam 解析路径中的 ID 参数。
func parseIDParam(c *gin.Context, name string) (int64, error) {
	v, err := strconv.ParseInt(c.Param(name), 10, 64)
	if err != nil || v <= 0 {
		return 0, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: name, Reason: "无效的 ID 参数"})
	}
	return v, nil
}
