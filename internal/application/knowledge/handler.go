// Package knowledge — 知识库 HTTP handlers（REST API）。
package knowledge

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/njydsz/ydsz-plane/internal/interfaces/middleware"
	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// Handler Gin handler 集合。
type Handler struct {
	svc *Service
}

// NewHandler 构造知识库 handler。
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// RegisterRead 注册只读路由（GET，需 knowledge:read 权限）。
func (h *Handler) RegisterRead(r *gin.RouterGroup) {
	// 空间集合
	spaces := r.Group("/spaces")
	{
		spaces.GET("", h.listSpaces)
		spaces.GET("/:sid", h.getSpace)

		space := spaces.Group("/:sid")
		{
			// 文档树
			space.GET("/pages", h.getPageTree)

			// 文档 GET 详情
			pages := space.Group("/pages")
			{
				pages.GET("/:pid", h.getPage)

				// 版本历史
				pages.GET("/:pid/versions", h.listVersions)

				// 工作项关联（读取）
				pages.GET("/:pid/relations", h.listRelations)
			}
		}
	}

	// 全文检索（跨空间，workspace_id 均来自 RequireWorkspaceParam）
	r.GET("/search", h.search)
}

// RegisterWrite 注册写入路由（POST/PATCH/DELETE，需 knowledge:manage 权限）。
func (h *Handler) RegisterWrite(r *gin.RouterGroup) {
	// 空间集合
	spaces := r.Group("/spaces")
	{
		spaces.POST("", h.createSpace)

		space := spaces.Group("/:sid")
		{
			space.PATCH("", h.updateSpace)
			space.DELETE("", h.deleteSpace)

			// 文档 CRUD（写入类）
			pages := space.Group("/pages")
			{
				pages.POST("", h.createPage)
				pages.PATCH("/:pid", h.updatePage)
				pages.DELETE("/:pid", h.deletePage)

				// 版本回滚
				pages.POST("/:pid/revert", h.revertVersion)

				// 工作项关联（写入类）
				rels := pages.Group("/:pid/relations")
				{
					rels.POST("", h.addRelation)
					rels.DELETE("/:rid", h.removeRelation)
				}
			}
		}
	}
}

// --- 空间 handlers ---

// listSpaces 列出工作空间下的知识库空间。
func (h *Handler) listSpaces(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)

	opts := ListSpacesOptions{
		WorkspaceID: wsID,
		Limit:       intQuery(c, "limit", 50),
		Offset:      intQuery(c, "offset", 0),
		Keyword:     c.Query("keyword"),
	}
	if v := c.Query("project_id"); v != "" {
		if id, err := parseInt64(v); err == nil {
			opts.ProjectID = &id
		}
	}

	items, total, err := h.svc.ListSpaces(c.Request.Context(), opts)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"results": items, "total": total, "limit": opts.Limit, "offset": opts.Offset})
}

// getSpace 获取单个空间详情。
func (h *Handler) getSpace(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	sid := int64Param(c, "sid")

	sp, err := h.svc.GetSpace(c.Request.Context(), sid, wsID)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, sp)
}

// createSpace 创建知识库空间。
func (h *Handler) createSpace(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	userID := c.GetInt64(middleware.CtxUserID)

	var req struct {
		Name              string `json:"name" binding:"required"`
		Slug              string `json:"slug" binding:"required"`
		Description       string `json:"description"`
		DefaultPermission string `json:"default_permission"`
		IsPrivate         bool   `json:"is_private"`
		CoverImage        string `json:"cover_image"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithError(c, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "body", Reason: err.Error()}))
		return
	}

	var projectID *int64
	if v := c.Query("project_id"); v != "" {
		if id, err := parseInt64(v); err == nil {
			projectID = &id
		}
	}

	in := CreateSpaceInput{
		WorkspaceID:       wsID,
		ProjectID:         projectID,
		Name:              req.Name,
		Slug:              req.Slug,
		Description:       req.Description,
		OwnerID:           &userID,
		DefaultPermission: SpacePermission(req.DefaultPermission),
		IsPrivate:         req.IsPrivate,
		CoverImage:        req.CoverImage,
	}
	if in.DefaultPermission == "" {
		in.DefaultPermission = PermissionViewer
	}

	sp, err := h.svc.CreateSpace(c.Request.Context(), in, userID)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, sp)
}

// updateSpace 更新知识库空间。
func (h *Handler) updateSpace(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	sid := int64Param(c, "sid")

	var req struct {
		Name              *string `json:"name"`
		Description       *string `json:"description"`
		DefaultPermission *string `json:"default_permission"`
		IsPrivate         *bool   `json:"is_private"`
		CoverImage        *string `json:"cover_image"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithError(c, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "body", Reason: err.Error()}))
		return
	}

	in := UpdateSpaceInput{
		Name:        req.Name,
		Description: req.Description,
		IsPrivate:   req.IsPrivate,
		CoverImage:  req.CoverImage,
	}
	if req.DefaultPermission != nil {
		p := SpacePermission(*req.DefaultPermission)
		in.DefaultPermission = &p
	}

	sp, err := h.svc.UpdateSpace(c.Request.Context(), sid, wsID, in)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, sp)
}

// deleteSpace 软删除空间。
func (h *Handler) deleteSpace(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	sid := int64Param(c, "sid")

	if err := h.svc.DeleteSpace(c.Request.Context(), sid, wsID); err != nil {
		writeErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// --- 文档 handlers ---

// getPageTree 获取 space 下全部文档的树形结构。
func (h *Handler) getPageTree(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	spaceID := int64Param(c, "sid")

	tree, err := h.svc.ListPageTree(c.Request.Context(), wsID, spaceID)
	if err != nil {
		writeErr(c, err)
		return
	}
	for i := range tree {
		if tree[i].Children == nil {
			tree[i].Children = []KnowledgePageNode{}
		}
	}
	if tree == nil {
		tree = []KnowledgePageNode{}
	}
	c.JSON(http.StatusOK, gin.H{"results": tree})
}

// getPage 获取单个文档详情。
func (h *Handler) getPage(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	pageID := int64Param(c, "pid")

	page, err := h.svc.GetPage(c.Request.Context(), pageID, wsID)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, page)
}

// createPage 创建文档。
func (h *Handler) createPage(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	spaceID := int64Param(c, "sid")
	userID := c.GetInt64(middleware.CtxUserID)

	var req CreatePageInput
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithError(c, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "body", Reason: err.Error()}))
		return
	}

	in := CreatePageInput{
		WorkspaceID: wsID,
		SpaceID:     spaceID,
		ParentID:    req.ParentID,
		Title:       req.Title,
		ContentMD:   req.ContentMD,
		ContentHTML: req.ContentHTML,
		Status:      req.Status,
		SortOrder:   req.SortOrder,
	}

	page, err := h.svc.CreatePage(c.Request.Context(), in, userID)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, page)
}

// updatePage 更新文档（乐观锁）。
func (h *Handler) updatePage(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	pageID := int64Param(c, "pid")

	var req struct {
		Title         *string `json:"title"`
		ContentMD     *string `json:"content_md"`
		ContentHTML   *string `json:"content_html"`
		ParentID      *int64  `json:"parent_id"`
		Status        *string `json:"status"`
		SortOrder     *int64  `json:"sort_order"`
		IsPinned      *bool   `json:"is_pinned"`
		IsFeatured    *bool   `json:"is_featured"`
		Version       int64   `json:"version" binding:"required"`
		ChangeSummary *string `json:"change_summary"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithError(c, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "body", Reason: err.Error()}))
		return
	}

	in := UpdatePageInput{
		Title:         req.Title,
		ContentMD:     req.ContentMD,
		ContentHTML:   req.ContentHTML,
		ParentID:      req.ParentID,
		SortOrder:     req.SortOrder,
		IsPinned:      req.IsPinned,
		IsFeatured:    req.IsFeatured,
		Version:       req.Version,
		ChangeSummary: req.ChangeSummary,
	}
	if req.Status != nil {
		s := PageStatus(*req.Status)
		in.Status = &s
	}

	page, err := h.svc.UpdatePage(c.Request.Context(), pageID, wsID, req.Version, in)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, page)
}

// deletePage 软删除文档。
func (h *Handler) deletePage(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	pageID := int64Param(c, "pid")

	if err := h.svc.DeletePage(c.Request.Context(), pageID, wsID, true); err != nil {
		writeErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// --- 版本 handlers ---

// listVersions 获取文档的版本快照列表。
func (h *Handler) listVersions(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	pageID := int64Param(c, "pid")

	versions, err := h.svc.ListVersions(c.Request.Context(), wsID, pageID)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"results": versions})
}

// revertVersion 回滚到指定版本（复制版本快照内容为最新版本）。
func (h *Handler) revertVersion(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	pageID := int64Param(c, "pid")

	var req struct {
		Version *int64 `json:"version" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Version == nil {
		middleware.AbortWithError(c, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "version", Reason: "版本号不能为空"}))
		return
	}

	// 先获取文档当前状态
	page, err := h.svc.GetPage(c.Request.Context(), pageID, wsID)
	if err != nil {
		writeErr(c, err)
		return
	}

	// 查找对应版本的快照
	versions, err := h.svc.ListVersions(c.Request.Context(), wsID, pageID)
	if err != nil {
		writeErr(c, err)
		return
	}

	var targetVersion *KnowledgePageVersion
	for i := range versions {
		if versions[i].Version == *req.Version {
			targetVersion = &versions[i]
			break
		}
	}
	if targetVersion == nil {
		writeErr(c, errs.ErrNotFound)
		return
	}

	// 利用 updatePage 机制将快照内容写回当前文档，版本号自动 +1
	in := UpdatePageInput{
		Title:         &targetVersion.Title,
		ContentMD:     &targetVersion.ContentMD,
		ContentHTML:   &targetVersion.ContentHTML,
		Version:       page.Version,
		ChangeSummary: strPtr("回滚到版本 " + itoa(*req.Version)),
	}

	updated, err := h.svc.UpdatePage(c.Request.Context(), pageID, wsID, page.Version, in)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, updated)
}

// --- 关联 handlers ---

// addRelation 添加文档与工作项的关联。
func (h *Handler) addRelation(c *gin.Context) {
	pageID := int64Param(c, "pid")

	var req struct {
		IssueID      int64  `json:"issue_id" binding:"required"`
		RelationType string `json:"relation_type"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithError(c, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "body", Reason: err.Error()}))
		return
	}

	in := AddPageRelationInput{
		PageID:       pageID,
		IssueID:      req.IssueID,
		RelationType: PageRelationType(req.RelationType),
	}
	if in.RelationType == "" {
		in.RelationType = RelationReferenced
	}

	rel, err := h.svc.AddPageRelation(c.Request.Context(), in)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, rel)
}

// listRelations 列出文档的关联工作项。
func (h *Handler) listRelations(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	pageID := int64Param(c, "pid")

	rels, err := h.svc.ListPageRelations(c.Request.Context(), wsID, pageID)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"results": rels})
}

// removeRelation 移除文档与工作项的关联。
func (h *Handler) removeRelation(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	pageID := int64Param(c, "pid")
	relID := int64Param(c, "rid")

	if err := h.svc.RemovePageRelation(c.Request.Context(), relID, pageID, wsID); err != nil {
		writeErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// --- helpers ---

func int64Param(c *gin.Context, key string) int64 {
	v, _ := parseInt64(c.Param(key))
	return v
}

func parseInt64(s string) (int64, error) {
	var v int64
	_, err := fmt.Sscanf(s, "%d", &v)
	return v, err
}

func intQuery(c *gin.Context, key string, def int) int {
	v, err := parseInt(c.Query(key))
	if err != nil {
		return def
	}
	return v
}

func parseInt(s string) (int, error) {
	var v int
	_, err := fmt.Sscanf(s, "%d", &v)
	return v, err
}

func itoa(v int64) string {
	return fmt.Sprintf("%d", v)
}

func strPtr(s string) *string {
	return &s
}

// search 全文检索（PostgreSQL tsvector）。
// GET /api/v1/workspaces/:workspace_id/knowledge/search?q=keyword&space_id=optional
// workspace_id 由 RequireWorkspaceParam 中间件注入。
func (h *Handler) search(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	keyword := c.Query("q")
	if strings.TrimSpace(keyword) == "" {
		c.JSON(200, gin.H{"results": []KnowledgePage{}})
		return
	}
	var sid *int64
	if v, err := strconv.ParseInt(c.Query("space_id"), 10, 64); err == nil {
		sid = &v
	}
	items, err := h.svc.Search(c.Request.Context(), wsID, keyword, sid, 20)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"results": items, "total": len(items)})
}

func writeErr(c *gin.Context, err error) {
	var appErr *errs.AppError
	if errs.As(err, &appErr) {
		middleware.AbortWithError(c, appErr)
		return
	}
	middleware.AbortWithError(c, errs.ErrInternal)
}
