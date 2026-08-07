package attachment

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

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

// Register 注册附件路由。
func (h *Handler) Register(r *gin.RouterGroup) {
	r.GET("/attachments", h.listAttachments)
	r.POST("/attachments/presigned-upload", h.getPresignedUploadURL)
	r.GET("/attachments/:id/download", h.downloadAttachment)
	r.DELETE("/attachments/:id", h.deleteAttachment)
}

// listAttachments GET /attachments?entity_type=issue&entity_id=123
func (h *Handler) listAttachments(c *gin.Context) {
	entityType := c.Query("entity_type")
	entityID := int64Param(c, "entity_id")

	atts, err := h.d.AttachmentSvc.ListByEntity(c.Request.Context(), entityType, entityID)
	if err != nil {
		writeErr(c, err)
		return
	}

	// 为返回的附件生成下载 URL
	for i := range atts {
		url, err := h.d.AttachmentSvc.Storage().PresignedDownloadURL(
			c.Request.Context(), atts[i].StorageKey, 15*time.Minute,
		)
		if err == nil {
			atts[i].StorageURL = url
		}
	}

	c.JSON(http.StatusOK, gin.H{"results": atts})
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

	if req.ContentType == "" {
		req.ContentType = "application/octet-stream"
	}

	// 生成唯一的存储 key：{ws}/{project}/{entity_type}/{entity_id}/{timestamp}_{filename}
	ext := filepath.Ext(req.FileName)
	storageKey := fmt.Sprintf("%d/%d/%s/%d/%d_%s%s",
		wsID, projectID, req.EntityType, req.EntityID,
		time.Now().UnixMilli(), sanitizeFilename(req.FileName), ext)

	uploadURL, err := h.d.AttachmentSvc.Storage().PresignedUploadURL(
		c.Request.Context(), storageKey, 15*time.Minute, req.ContentType,
	)
	if err != nil {
		writeErr(c, err)
		return
	}

	// 预创建数据库记录
	att, err := h.d.AttachmentSvc.Create(c.Request.Context(), CreateInput{
		WorkspaceID: wsID,
		ProjectID:   projectID,
		EntityType:  req.EntityType,
		EntityID:    req.EntityID,
		FileName:    req.FileName,
		FileSize:    0,
		ContentType: req.ContentType,
		StorageKey:  storageKey,
		UploadedBy:  userID,
	})
	if err != nil {
		writeErr(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"upload_url":  uploadURL,
		"storage_key": storageKey,
		"attachment":  att,
	})
}

// downloadAttachment 生成预签名下载 URL 并 302 重定向。
// GET /attachments/:id/download
func (h *Handler) downloadAttachment(c *gin.Context) {
	id := int64Param(c, "id")

	att, err := h.d.AttachmentSvc.Get(c.Request.Context(), id)
	if err != nil {
		writeErr(c, err)
		return
	}

	downloadURL, err := h.d.AttachmentSvc.Storage().PresignedDownloadURL(
		c.Request.Context(), att.StorageKey, 15*time.Minute,
	)
	if err != nil {
		writeErr(c, err)
		return
	}

	c.Redirect(http.StatusFound, downloadURL)
}

// deleteAttachment DELETE /attachments/:id
func (h *Handler) deleteAttachment(c *gin.Context) {
	id := int64Param(c, "id")
	userID := c.GetInt64(middleware.CtxUserID)

	if err := h.d.AttachmentSvc.Delete(c.Request.Context(), id, userID); err != nil {
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

func sanitizeFilename(name string) string {
	name = filepath.Base(name)
	name = strings.Map(func(r rune) rune {
		if r == ' ' || r == '\\' || r == '/' || r == ':' || r == '*' || r == '?' || r == '"' || r == '<' || r == '>' || r == '|' {
			return '_'
		}
		return r
	}, name)
	if len(name) > 200 {
		name = name[:200]
	}
	return name
}

func fieldDetail(err error) []errs.FieldDetail {
	// 提供基础的字段级错误详情
	if err != nil {
		return []errs.FieldDetail{
			{Field: "body", Reason: err.Error()},
		}
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
