// Package pages — 项目文档页面 HTTP handlers。
package pages

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/njydsz/ydsz-plane/internal/interfaces/middleware"
	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// Handler 页面 HTTP handler。
type Handler struct {
	svc *Service
}

// NewHandler 创建 handler。
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// Register 注册页面路由（挂在项目子路由组下）。
// 路由前缀：/api/v1/workspaces/:workspace_id/projects/:project_id/pages
func (h *Handler) Register(r *gin.RouterGroup) {
	pages := r.Group("/pages")
	{
		pages.GET("", h.listPages)
		pages.POST("", h.createPage)
		pages.GET("/:page_id", h.getPage)
		pages.PATCH("/:page_id", h.updatePage)
		pages.DELETE("/:page_id", h.deletePage)

		// 版本历史
		pages.GET("/:page_id/versions", h.listVersions)
		pages.GET("/:page_id/versions/:version_id", h.getVersion)
		pages.POST("/:page_id/versions/:version_id/rollback", h.rollbackToVersion)

		// 文档关联
		pages.POST("/:page_id/links", h.createLink)
		pages.GET("/:page_id/links", h.listLinks)
		pages.DELETE("/:page_id/links/:link_id", h.deleteLink)

		// 文档公开分享
		pages.POST("/:page_id/shares", h.createShare)
		pages.GET("/:page_id/shares", h.listShares)
		pages.PATCH("/:page_id/shares/:share_id", h.updateShare)
		pages.DELETE("/:page_id/shares/:share_id", h.revokeShare)
	}

	// 文档模板（同级路由组：/templates）
	templates := r.Group("/templates")
	{
		templates.GET("", h.listTemplates)
		templates.POST("", h.createTemplate)
		templates.PATCH("/:template_id", h.updateTemplate)
		templates.DELETE("/:template_id", h.deleteTemplate)
	}
}

// listPages 列出项目下全部页面。
//
//	@Summary		列出页面
//	@Tags			pages
//	@Produce		json
//	@Success		200	{object}	[]Page
//	@Router			/pages [get]
func (h *Handler) listPages(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	projectID := c.GetInt64(middleware.CtxProjectID)

	list, err := h.svc.List(c.Request.Context(), wsID, projectID)
	if err != nil {
		writeErr(c, err)
		return
	}
	if list == nil {
		list = []Page{}
	}
	c.JSON(http.StatusOK, gin.H{"results": list})
}

// createPage 创建页面。
//
//	@Summary		创建页面
//	@Tags			pages
//	@Accept			json
//	@Produce		json
//	@Param			body	body	CreatePageInput	true	"页面内容"
//	@Success		200		{object}	Page
//	@Router			/pages [post]
func (h *Handler) createPage(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	projectID := c.GetInt64(middleware.CtxProjectID)
	userID := c.GetInt64(middleware.CtxUserID)

	var req CreatePageInput
	if err := c.ShouldBindJSON(&req); err != nil {
		writeErr(c, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "body", Reason: err.Error()}))
		return
	}

	page, err := h.svc.Create(c.Request.Context(), wsID, projectID, userID, req)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, page)
}

// getPage 获取单个页面。
//
//	@Summary		获取页面
//	@Tags			pages
//	@Produce		json
//	@Success		200	{object}	Page
//	@Router			/pages/{page_id} [get]
func (h *Handler) getPage(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	projectID := c.GetInt64(middleware.CtxProjectID)

	pageID, err := pageIDParam(c)
	if err != nil {
		writeErr(c, err)
		return
	}

	page, err := h.svc.Get(c.Request.Context(), wsID, projectID, pageID)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, page)
}

// updatePage 更新页面（乐观锁）。
//
//	@Summary		更新页面
//	@Tags			pages
//	@Accept			json
//	@Produce		json
//	@Param			body	body	UpdatePageInput	true	"页面内容"
//	@Success		200		{object}	Page
//	@Router			/pages/{page_id} [patch]
func (h *Handler) updatePage(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	projectID := c.GetInt64(middleware.CtxProjectID)
	userID := c.GetInt64(middleware.CtxUserID)

	pageID, err := pageIDParam(c)
	if err != nil {
		writeErr(c, err)
		return
	}

	var req UpdatePageInput
	if err := c.ShouldBindJSON(&req); err != nil {
		writeErr(c, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "body", Reason: err.Error()}))
		return
	}

	page, err := h.svc.Update(c.Request.Context(), wsID, projectID, pageID, userID, req)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, page)
}

// deletePage 软删除页面。
//
//	@Summary		删除页面
//	@Tags			pages
//	@Success		200	{object}	map[string]bool
//	@Router			/pages/{page_id} [delete]
func (h *Handler) deletePage(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	projectID := c.GetInt64(middleware.CtxProjectID)
	userID := c.GetInt64(middleware.CtxUserID)

	pageID, err := pageIDParam(c)
	if err != nil {
		writeErr(c, err)
		return
	}

	if err := h.svc.Delete(c.Request.Context(), wsID, projectID, pageID, userID); err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// pageIDParam 解析路径参数 :page_id。
func pageIDParam(c *gin.Context) (int64, error) {
	id, err := strconv.ParseInt(c.Param("page_id"), 10, 64)
	if err != nil {
		return 0, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "page_id", Reason: "无效的页面 ID"})
	}
	return id, nil
}

// versionIDParam 解析路径参数 :version_id（版本号 int32）。
func versionIDParam(c *gin.Context) (int32, error) {
	v, err := strconv.ParseInt(c.Param("version_id"), 10, 32)
	if err != nil {
		return 0, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "version_id", Reason: "无效的版本号"})
	}
	return int32(v), nil
}

// linkIDParam 解析路径参数 :link_id。
func linkIDParam(c *gin.Context) (int64, error) {
	id, err := strconv.ParseInt(c.Param("link_id"), 10, 64)
	if err != nil {
		return 0, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "link_id", Reason: "无效的关联 ID"})
	}
	return id, nil
}

// --- 版本历史 handlers ---

// listVersions 获取页面所有版本历史。
func (h *Handler) listVersions(c *gin.Context) {
	pageID, err := pageIDParam(c)
	if err != nil {
		writeErr(c, err)
		return
	}

	list, err := h.svc.ListVersions(c.Request.Context(), pageID)
	if err != nil {
		writeErr(c, err)
		return
	}
	if list == nil {
		list = []DocumentVersion{}
	}
	c.JSON(http.StatusOK, gin.H{"results": list})
}

// getVersion 获取指定版本快照。
func (h *Handler) getVersion(c *gin.Context) {
	pageID, err := pageIDParam(c)
	if err != nil {
		writeErr(c, err)
		return
	}
	versionNumber, err := versionIDParam(c)
	if err != nil {
		writeErr(c, err)
		return
	}

	dv, err := h.svc.GetVersion(c.Request.Context(), pageID, versionNumber)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, dv)
}

// rollbackToVersion 回滚页面到指定版本。
func (h *Handler) rollbackToVersion(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	projectID := c.GetInt64(middleware.CtxProjectID)
	userID := c.GetInt64(middleware.CtxUserID)

	pageID, err := pageIDParam(c)
	if err != nil {
		writeErr(c, err)
		return
	}
	versionNumber, err := versionIDParam(c)
	if err != nil {
		writeErr(c, err)
		return
	}

	page, err := h.svc.RollbackToVersion(c.Request.Context(), wsID, projectID, pageID, userID, versionNumber)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, page)
}

// --- 文档关联 handlers ---

// createLink 创建文档与外部实体的关联。
func (h *Handler) createLink(c *gin.Context) {
	userID := c.GetInt64(middleware.CtxUserID)

	pageID, err := pageIDParam(c)
	if err != nil {
		writeErr(c, err)
		return
	}

	var req struct {
		LinkableType string `json:"linkable_type"`
		LinkableID   int64  `json:"linkable_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		writeErr(c, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "body", Reason: err.Error()}))
		return
	}
	if req.LinkableType == "" {
		writeErr(c, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "linkable_type", Reason: "关联类型不能为空"}))
		return
	}
	if req.LinkableID <= 0 {
		writeErr(c, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "linkable_id", Reason: "关联 ID 无效"}))
		return
	}

	link, err := h.svc.CreateLink(c.Request.Context(), pageID, userID, req.LinkableType, req.LinkableID)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, link)
}

// listLinks 获取页面所有关联。
func (h *Handler) listLinks(c *gin.Context) {
	pageID, err := pageIDParam(c)
	if err != nil {
		writeErr(c, err)
		return
	}

	list, err := h.svc.ListLinks(c.Request.Context(), pageID)
	if err != nil {
		writeErr(c, err)
		return
	}
	if list == nil {
		list = []DocumentLink{}
	}
	c.JSON(http.StatusOK, gin.H{"results": list})
}

// deleteLink 删除文档关联。
func (h *Handler) deleteLink(c *gin.Context) {
	pageID, err := pageIDParam(c)
	if err != nil {
		writeErr(c, err)
		return
	}
	linkID, err := linkIDParam(c)
	if err != nil {
		writeErr(c, err)
		return
	}

	if err := h.svc.DeleteLink(c.Request.Context(), pageID, linkID); err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// --- 文档模板 handlers ---

// listTemplates 列出项目可用模板（工作空间级 + 项目级 + 全局内置）。
func (h *Handler) listTemplates(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	projectID := c.GetInt64(middleware.CtxProjectID)

	list, err := h.svc.ListTemplates(c.Request.Context(), wsID, projectID)
	if err != nil {
		writeErr(c, err)
		return
	}
	if list == nil {
		list = []PageTemplate{}
	}
	c.JSON(http.StatusOK, gin.H{"results": list})
}

// createTemplate 创建文档模板。
func (h *Handler) createTemplate(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	projectID := c.GetInt64(middleware.CtxProjectID)
	userID := c.GetInt64(middleware.CtxUserID)

	var req CreateTemplateInput
	if err := c.ShouldBindJSON(&req); err != nil {
		writeErr(c, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "body", Reason: err.Error()}))
		return
	}

	tmpl, err := h.svc.CreateTemplate(c.Request.Context(), wsID, projectID, userID, req)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, tmpl)
}

// updateTemplate 更新文档模板。
func (h *Handler) updateTemplate(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)

	templateID, err := templateIDParam(c)
	if err != nil {
		writeErr(c, err)
		return
	}

	var req UpdateTemplateInput
	if err := c.ShouldBindJSON(&req); err != nil {
		writeErr(c, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "body", Reason: err.Error()}))
		return
	}

	tmpl, err := h.svc.UpdateTemplate(c.Request.Context(), wsID, templateID, req)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, tmpl)
}

// deleteTemplate 删除文档模板。
func (h *Handler) deleteTemplate(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)

	templateID, err := templateIDParam(c)
	if err != nil {
		writeErr(c, err)
		return
	}

	if err := h.svc.DeleteTemplate(c.Request.Context(), wsID, templateID); err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// --- 文档公开分享 handlers ---

// createShare 为文档创建公开分享链接。
func (h *Handler) createShare(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	projectID := c.GetInt64(middleware.CtxProjectID)
	userID := c.GetInt64(middleware.CtxUserID)

	pageID, err := pageIDParam(c)
	if err != nil {
		writeErr(c, err)
		return
	}

	var req CreateShareInput
	if err := c.ShouldBindJSON(&req); err != nil {
		writeErr(c, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "body", Reason: err.Error()}))
		return
	}

	share, err := h.svc.CreateShare(c.Request.Context(), wsID, projectID, pageID, userID, req)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, share)
}

// listShares 列出文档所有分享链接。
func (h *Handler) listShares(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)

	pageID, err := pageIDParam(c)
	if err != nil {
		writeErr(c, err)
		return
	}

	list, err := h.svc.ListShares(c.Request.Context(), wsID, pageID)
	if err != nil {
		writeErr(c, err)
		return
	}
	if list == nil {
		list = []PageShare{}
	}
	c.JSON(http.StatusOK, gin.H{"results": list})
}

// updateShare 更新分享链接（状态 / 密码 / 过期时间）。
func (h *Handler) updateShare(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)

	shareID, err := shareIDParam(c)
	if err != nil {
		writeErr(c, err)
		return
	}

	var req UpdateShareInput
	if err := c.ShouldBindJSON(&req); err != nil {
		writeErr(c, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "body", Reason: err.Error()}))
		return
	}

	share, err := h.svc.UpdateShare(c.Request.Context(), wsID, shareID, req)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, share)
}

// revokeShare 吊销分享链接。
func (h *Handler) revokeShare(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)

	shareID, err := shareIDParam(c)
	if err != nil {
		writeErr(c, err)
		return
	}

	if err := h.svc.RevokeShare(c.Request.Context(), wsID, shareID); err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// --- 公开分享访问 handler（免登录） ---

// PublicShareHandler 公开分享处理器（与认证路由分开注册）。
type PublicShareHandler struct {
	svc *Service
}

// NewPublicShareHandler 构造公开分享 handler。
func NewPublicShareHandler(svc *Service) *PublicShareHandler {
	return &PublicShareHandler{svc: svc}
}

// GetSharedPage 获取公开分享页面内容。
// 查询参数：?password=xxx （可选，当分享设置了密码时需提供）。
func (h *PublicShareHandler) GetSharedPage(c *gin.Context) {
	token := c.Param("token")
	password := c.Query("password")

	view, _, err := h.svc.GetSharedPageView(c.Request.Context(), token, password)
	if err != nil {
		var appErr *errs.AppError
		if errs.As(err, &appErr) {
			// 特殊标记：需要密码时返回 401 + 提示
			if appErr.Code == "PAGES.SHARE_PASSWORD_REQUIRED" {
				c.JSON(http.StatusUnauthorized, gin.H{"error": appErr, "require_password": true})
				return
			}
			c.JSON(appErr.HTTP, gin.H{"error": appErr})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": errs.ErrInternal})
		return
	}
	c.JSON(http.StatusOK, view)
}

// writeErr 统一错误响应（同 preference 模块）。
func writeErr(c *gin.Context, err error) {
	var appErr *errs.AppError
	if errs.As(err, &appErr) {
		middleware.AbortWithError(c, appErr)
		return
	}
	middleware.AbortWithError(c, errs.ErrInternal)
}

// templateIDParam 解析路径参数 :template_id。
func templateIDParam(c *gin.Context) (int64, error) {
	id, err := strconv.ParseInt(c.Param("template_id"), 10, 64)
	if err != nil {
		return 0, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "template_id", Reason: "无效的模板 ID"})
	}
	return id, nil
}

// shareIDParam 解析路径参数 :share_id。
func shareIDParam(c *gin.Context) (int64, error) {
	id, err := strconv.ParseInt(c.Param("share_id"), 10, 64)
	if err != nil {
		return 0, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "share_id", Reason: "无效的分享 ID"})
	}
	return id, nil
}
