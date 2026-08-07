// Package attachment 附件域 HTTP 处理器：列表查询、预签名直传与删除。
package attachment

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/njydsz/ydsz-plane/internal/interfaces/middleware"
	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// HandlerDeps handler 依赖。
type HandlerDeps struct {
	AttachmentSvc *Service
}

// Handler 附件 HTTP handler。
type Handler struct {
	d *HandlerDeps
}

// NewHandler 构造 handler。
func NewHandler(d *HandlerDeps) *Handler {
	return &Handler{d: d}
}

// Register 注册附件路由（项目级）。
func (h *Handler) Register(r *gin.RouterGroup) {
	r.GET("/attachments", h.listAttachments)
	r.POST("/attachments/presigned-upload", h.getPresignedUploadURL)
	r.DELETE("/attachments/:id", h.deleteAttachment)
}

// listAttachments GET /attachments?entity_type=issue&entity_id=123
func (h *Handler) listAttachments(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	projectID := c.GetInt64(middleware.CtxProjectID)
	entityType := c.Query("entity_type")
	entityID := int64Param(c, "entity_id")

	if entityType == "" {
		middleware.AbortWithError(c, errs.ErrValidation.WithDetails(
			errs.FieldDetail{Field: "entity_type", Reason: "required"},
		))
		return
	}

	atts, err := h.d.AttachmentSvc.List(c.Request.Context(), wsID, projectID, entityType, entityID)
	if err != nil {
		writeErr(c, err)
		return
	}

	c.JSON(http.StatusOK, ListResponse{Results: atts})
}

// getPresignedUploadURL 获取预签名上传 URL。
// POST /attachments/presigned-upload
// Body: { "file_name": "xxx.png", "content_type": "image/png", "entity_type": "issue", "entity_id": 1 }
func (h *Handler) getPresignedUploadURL(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	projectID := c.GetInt64(middleware.CtxProjectID)
	userID := c.GetInt64(middleware.CtxUserID)

	var req struct {
		FileName    string `json:"file_name" binding:"required"`
		ContentType string `json:"content_type"`
		EntityType  string `json:"entity_type" binding:"required"`
		EntityID    int64  `json:"entity_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithError(c, errs.ErrValidation.WithDetails(fieldDetail(err)...))
		return
	}

	result, err := h.d.AttachmentSvc.CreatePresignedUpload(c.Request.Context(), wsID, projectID, PresignedUploadInput{
		FileName:    req.FileName,
		ContentType: req.ContentType,
		EntityType:  req.EntityType,
		EntityID:    req.EntityID,
		UploadedBy:  userID,
	})
	if err != nil {
		writeErr(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// deleteAttachment DELETE /attachments/:id
func (h *Handler) deleteAttachment(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	projectID := c.GetInt64(middleware.CtxProjectID)
	userID := c.GetInt64(middleware.CtxUserID)
	id := int64Param(c, "id")

	if err := h.d.AttachmentSvc.Delete(c.Request.Context(), wsID, projectID, id, userID); err != nil {
		writeErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ---------- 工具函数 ----------

func int64Param(c *gin.Context, key string) int64 {
	v, _ := c.Params.Get(key)
	if v == "" {
		v = c.Query(key)
	}
	var n int64
	fmt.Sscanf(v, "%d", &n)
	return n
}

func fieldDetail(err error) []errs.FieldDetail {
	if err != nil {
		return []errs.FieldDetail{{Field: "body", Reason: err.Error()}}
	}
	return nil
}

func writeErr(c *gin.Context, err error) {
	if appErr, ok := err.(*errs.AppError); ok {
		c.JSON(appErr.HTTP, gin.H{"error": appErr})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{
		"code":    "INTERNAL",
		"message": err.Error(),
	}})
}
