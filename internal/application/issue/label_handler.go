// Package issue — Label HTTP handlers（REST API）。
//
// 对齐 modules 的路由模式：
//   - 列表 + 创建 挂在 /labels
//   - 单资源 挂在 /labels/:label_id
package issue

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/njydsz/ydsz-plane/internal/interfaces/middleware"
	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// LabelHandler Gin handler 集合（标签 CRUD）。
type LabelHandler struct {
	svc *LabelService
}

// NewLabelHandler 构造标签 handler。
func NewLabelHandler(svc *LabelService) *LabelHandler {
	return &LabelHandler{svc: svc}
}

// Register 注册标签路由（在项目子路由组下）。
func (h *LabelHandler) Register(r *gin.RouterGroup) {
	labels := r.Group("/labels")
	{
		labels.GET("", h.listLabels)
		labels.POST("", h.createLabel)
	}
	label := labels.Group("/:label_id")
	{
		label.GET("", h.getLabel)
		label.PATCH("", h.updateLabel)
		label.DELETE("", h.deleteLabel)
	}
}

// ---- request DTOs ----

type createLabelRequest struct {
	Name        string `json:"name" binding:"required"`
	Color       string `json:"color"`
	Description string `json:"description"`
}

type updateLabelRequest struct {
	Name        *string `json:"name"`
	Color       *string `json:"color"`
	Description *string `json:"description"`
}

// ---- handlers ----

// listLabels GET /labels?status=
func (h *LabelHandler) listLabels(c *gin.Context) {
	labels, err := h.svc.ListLabels(c.Request.Context(), ListLabelsFilter{
		WorkspaceID: extractWsID(c),
		ProjectID:   extractProjectID(c),
		Status:      c.Query("status"),
	})
	if err != nil {
		handleLabelErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"results": labels})
}

// createLabel POST /labels
func (h *LabelHandler) createLabel(c *gin.Context) {
	var req createLabelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handleLabelErr(c, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "body", Reason: "请求体无效: " + err.Error()}))
		return
	}
	l, err := h.svc.CreateLabel(c.Request.Context(), CreateLabelInput{
		WorkspaceID: extractWsID(c),
		ProjectID:   extractProjectID(c),
		Name:        req.Name,
		Color:       req.Color,
		Description: req.Description,
		CreatedBy:   extractActorID(c),
	})
	if err != nil {
		handleLabelErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, l)
}

// getLabel GET /labels/:label_id
func (h *LabelHandler) getLabel(c *gin.Context) {
	id, err := labelIDParam(c)
	if err != nil {
		handleLabelErr(c, err)
		return
	}
	l, err := h.svc.GetLabel(c.Request.Context(), extractWsID(c), id)
	if err != nil {
		handleLabelErr(c, err)
		return
	}
	c.JSON(http.StatusOK, l)
}

// updateLabel PATCH /labels/:label_id
func (h *LabelHandler) updateLabel(c *gin.Context) {
	id, err := labelIDParam(c)
	if err != nil {
		handleLabelErr(c, err)
		return
	}
	var req updateLabelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handleLabelErr(c, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "body", Reason: "请求体无效: " + err.Error()}))
		return
	}
	l, err := h.svc.UpdateLabel(c.Request.Context(), UpdateLabelInput{
		ID:          id,
		WorkspaceID: extractWsID(c),
		ProjectID:   extractProjectID(c),
		Name:        req.Name,
		Color:       req.Color,
		Description: req.Description,
	})
	if err != nil {
		handleLabelErr(c, err)
		return
	}
	c.JSON(http.StatusOK, l)
}

// deleteLabel DELETE /labels/:label_id
func (h *LabelHandler) deleteLabel(c *gin.Context) {
	id, err := labelIDParam(c)
	if err != nil {
		handleLabelErr(c, err)
		return
	}
	if err := h.svc.DeleteLabel(c.Request.Context(), extractWsID(c), extractProjectID(c), id); err != nil {
		handleLabelErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ---- helpers ----

func labelIDParam(c *gin.Context) (int64, error) {
	v, err := strconv.ParseInt(c.Param("label_id"), 10, 64)
	if err != nil || v <= 0 {
		return 0, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "label_id", Reason: "无效的标签 ID"})
	}
	return v, nil
}

func handleLabelErr(c *gin.Context, err error) {
	var appErr *errs.AppError
	if errs.As(err, &appErr) {
		middleware.AbortWithError(c, appErr)
		return
	}
	middleware.AbortWithError(c, errs.ErrInternal)
}
