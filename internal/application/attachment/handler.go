// Package attachment 附件域 HTTP 处理器：列表查询、预签名直传与删除。
//
// 安全加固（S7-P0）：
//   - 单文件大小、MIME 类型、扩展名三重校验（客户端传入的 ContentType 不可信）；
//   - 单工作项附件总量限制；
//   - 文件名长度与 storage_key 格式二次校验。
package attachment

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/njydsz/ydsz-plane/internal/config"
	"github.com/njydsz/ydsz-plane/internal/interfaces/middleware"
	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// HandlerDeps handler 依赖。
type HandlerDeps struct {
	AttachmentSvc *serviceImpl
	Cfg           *config.AttachmentConfig
}

// Handler 附件 HTTP handler。
type Handler struct {
	d *HandlerDeps
}

// NewHandler 构造 handler。
func NewHandler(d *HandlerDeps) *Handler {
	return &Handler{d: d}
}

// serviceImpl 是 Service 接口的别名，用于解耦 handler 中的 typed nil 问题。
type serviceImpl = Service

// Register 注册附件路由（项目级）。
func (h *Handler) Register(r *gin.RouterGroup) {
	r.GET("/attachments", h.listAttachments)
	r.GET("/issues/:issue_id/attachments", h.listIssueAttachments)
	r.POST("/attachments/presigned-upload", h.getPresignedUploadURL)
	r.POST("/attachments/confirm", h.confirmUpload)
	r.DELETE("/attachments/:id", h.deleteAttachment)
}

// validateUploadInput 校验上传请求的合法性。
// 校验维度：文件名、MIME 类型（白名单）、扩展名（白名单）、文件大小。
func (h *Handler) validateUploadInput(fileName, contentType string, fileSize int64) error {
	cfg := h.d.Cfg

	// --- 文件大小 ---
	if cfg != nil && cfg.MaxFileSize > 0 && fileSize > 0 && fileSize > cfg.MaxFileSize {
		return errs.Validation("ATTACHMENT.FILE_TOO_LARGE",
			fmt.Sprintf("文件超过最大限制 %d MB", cfg.MaxFileSize/1024/1024))
	}

	// --- 扩展名白名单 ---
	ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(fileName)), ".")
	if cfg != nil && len(cfg.AllowedExtensions) > 0 && ext != "" {
		found := false
		for _, allowed := range cfg.AllowedExtensions {
			if strings.ToLower(allowed) == ext {
				found = true
				break
			}
		}
		if !found {
			return errs.Validation("ATTACHMENT.EXTENSION_NOT_ALLOWED",
				fmt.Sprintf("不支持的文件扩展名 .%s", ext))
		}
	}

	// --- MIME 类型白名单 ---
	if cfg != nil && len(cfg.AllowedContentTypes) > 0 && contentType != "" {
		// 去除可能的 charset 等参数（如 "image/jpeg; charset=utf-8"）
		ct := strings.Split(contentType, ";")[0]
		ct = strings.TrimSpace(ct)
		found := false
		for _, allowed := range cfg.AllowedContentTypes {
			if strings.ToLower(allowed) == ct {
				found = true
				break
			}
		}
		if !found {
			return errs.Validation("ATTACHMENT.CONTENT_TYPE_NOT_ALLOWED",
				fmt.Sprintf("不支持的文件类型 %s", ct))
		}
	}

	return nil
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

	// 上传限制校验（大小/MIME/扩展名白名单）
	if err := h.validateUploadInput(req.FileName, req.ContentType, 0); err != nil {
		if appErr, ok := err.(*errs.AppError); ok {
			middleware.AbortWithError(c, appErr)
		} else {
			writeErr(c, err)
		}
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

// listIssueAttachments GET /issues/:issue_id/attachments — 便捷路由，等价于 GET /attachments?entity_type=issue&entity_id=:issue_id
func (h *Handler) listIssueAttachments(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	projectID := c.GetInt64(middleware.CtxProjectID)
	issueID := int64Param(c, "issue_id")

	atts, err := h.d.AttachmentSvc.List(c.Request.Context(), wsID, projectID, "issue", issueID)
	if err != nil {
		writeErr(c, err)
		return
	}

	c.JSON(http.StatusOK, ListResponse{Results: atts})
}

// confirmUpload POST /attachments/confirm — 客户端 PUT 成功后提交，写入 DB 记录。
func (h *Handler) confirmUpload(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	projectID := c.GetInt64(middleware.CtxProjectID)
	userID := c.GetInt64(middleware.CtxUserID)

	var req struct {
		FileName    string `json:"file_name" binding:"required"`
		ContentType string `json:"content_type"`
		FileSize    int64  `json:"file_size"`
		EntityType  string `json:"entity_type" binding:"required"`
		EntityID    int64  `json:"entity_id" binding:"required"`
		StorageKey  string `json:"storage_key" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithError(c, errs.ErrValidation.WithDetails(fieldDetail(err)...))
		return
	}

	// confirm 阶段再次校验文件大小（客户端传回实际大小）
	if err := h.validateUploadInput(req.FileName, req.ContentType, req.FileSize); err != nil {
		if appErr, ok := err.(*errs.AppError); ok {
			middleware.AbortWithError(c, appErr)
		} else {
			writeErr(c, err)
		}
		return
	}

	// 单工作项附件总量限制
	cfg := h.d.Cfg
	if cfg != nil && cfg.MaxTotalSizePerIssue > 0 && req.EntityType == "issue" {
		atts, err := h.d.AttachmentSvc.List(c.Request.Context(), wsID, projectID, "issue", req.EntityID)
		if err != nil {
			writeErr(c, err)
			return
		}
		var existingTotal int64
		for _, a := range atts {
			existingTotal += a.FileSize
		}
		if existingTotal+req.FileSize > cfg.MaxTotalSizePerIssue {
			middleware.AbortWithError(c, errs.Validation("ATTACHMENT.TOTAL_SIZE_EXCEEDED",
				fmt.Sprintf("工作项附件总容量超过 %d MB 限制",
					cfg.MaxTotalSizePerIssue/1024/1024)))
			return
		}

	}

	att, err := h.d.AttachmentSvc.ConfirmUpload(c.Request.Context(), wsID, projectID, ConfirmUploadInput{
		FileName:    req.FileName,
		ContentType: req.ContentType,
		FileSize:    req.FileSize,
		EntityType:  req.EntityType,
		EntityID:    req.EntityID,
		StorageKey:  req.StorageKey,
		UploadedBy:  userID,
	})
	if err != nil {
		writeErr(c, err)
		return
	}

	c.JSON(http.StatusCreated, ConfirmUploadResult{Attachment: *att})
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
