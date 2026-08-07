// Package search — 搜索 HTTP handlers（REST API）。
package search

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/njydsz/ydsz-plane/internal/application/auth"
	"github.com/njydsz/ydsz-plane/internal/interfaces/middleware"
	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// HandlerDeps search handler 依赖。
type HandlerDeps struct {
	SearchSvc      *Service
	WorkspaceStore *auth.WorkspaceMembershipStore
}

// SearchHandler Gin handler 集合。
type SearchHandler struct {
	d *HandlerDeps
}

// NewSearchHandler 构造 handler。
func NewSearchHandler(d *HandlerDeps) *SearchHandler {
	return &SearchHandler{d: d}
}

// Register 注册项目级搜索路由。
func (h *SearchHandler) Register(r *gin.RouterGroup) {
	r.GET("", h.Search)
	r.GET("/history", h.ListHistory)
	r.DELETE("/history/:history_id", h.DeleteHistory)
	r.DELETE("/history", h.ClearHistory)
	r.GET("/bookmarks", h.ListBookmarks)
	r.POST("/bookmarks", h.CreateBookmark)
	r.PATCH("/bookmarks/:bookmark_id", h.UpdateBookmark)
	r.DELETE("/bookmarks/:bookmark_id", h.DeleteBookmark)
}

// Search 全局搜索（REST handler）。
//
//	@Summary		全文搜索
//	@Description	跨对象（issues/sprints/versions）全文检索 + 高亮
//	@Tags			search
//	@Produce		json
//	@Param			q			query		string	true	"搜索词"
//	@Param			doc_type	query		string	false	"过滤类型 (issue|sprint|version)"
//	@Param			type		query		string	false	"工作项类型 (requirement|task|defect)"
//	@Param			priority	query		string	false	"优先级"
//	@Param			state_id	query		int		false	"状态 ID"
//	@Param			limit		query		int		false	"每页数量 (default 20, max 50)"
//	@Param			offset		query		int		false	"偏移量"
//	@Success		200			{object}	SearchResponse
//	@Router			/search [get]
func (h *SearchHandler) Search(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	projectID := c.GetInt64(middleware.CtxProjectID)
	userID := c.GetInt64(middleware.CtxUserID)

	// 支持 q / query 双参数（对齐 Jira/Plane 惯例 + 向后兼容）
	query := c.Query("q")
	if query == "" {
		query = c.Query("query")
	}
	if strings.TrimSpace(query) == "" {
		middleware.AbortWithError(c, errs.ErrValidation.WithDetails(errs.FieldDetail{
			Field:  "q",
			Reason: "搜索词不能为空",
		}))
		return
	}

	// 构建过滤条件
	filters := map[string]any{}
	if v := c.Query("type"); v != "" {
		filters["type_code"] = v
	}
	if v := c.Query("priority"); v != "" {
		filters["priority"] = v
	}
	if v := c.Query("state_id"); v != "" {
		if id, err := strconv.ParseInt(v, 10, 64); err == nil {
			filters["state_id"] = id
		}
	}

	// 文档类型（支持 types 多选 / doc_type 单选，对齐前后端）
	var docTypes []string
	if v := c.Query("types"); v != "" {
		docTypes = append(docTypes, strings.Split(v, ",")...)
	}
	if v := c.Query("doc_type"); v != "" {
		docTypes = append(docTypes, v)
	}

	resp, err := h.d.SearchSvc.Search(c.Request.Context(), SearchQuery{
		WorkspaceID: wsID,
		ProjectID:   projectID,
		UserID:      userID,
		Query:       query,
		DocTypes:    docTypes,
		Filters:     filters,
		Limit:       intQuery(c, "limit", 20),
		Offset:      intQuery(c, "offset", 0),
	})
	if err != nil {
		writeErr(c, err)
		return
	}

	// 异步记录搜索历史（不阻塞响应）
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = h.d.SearchSvc.RecordHistory(ctx, RecordHistoryInput{
			WorkspaceID: wsID,
			UserID:      userID,
			Query:       query,
			Filters:     filters,
			ResultCount: resp.Total,
		})
	}()

	c.JSON(http.StatusOK, resp)
}

// --- History ---

// ListHistory 列出搜索历史。
//
//	@Summary		搜索历史
//	@Tags			search
//	@Produce		json
//	@Success		200	{object}	searchHistoryListResponse
//	@Router			/search/history [get]
func (h *SearchHandler) ListHistory(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	userID := c.GetInt64(middleware.CtxUserID)

	entries, err := h.d.SearchSvc.ListHistory(c.Request.Context(), wsID, userID, intQuery(c, "limit", 20))
	if err != nil {
		writeErr(c, err)
		return
	}
	if entries == nil {
		entries = []SearchHistoryEntry{}
	}
	c.JSON(http.StatusOK, gin.H{"results": entries})
}

// DeleteHistory 删除单条搜索历史。
//
//	@Summary		删除搜索历史
//	@Tags			search
//	@Success		204
//	@Router			/search/history/{history_id} [delete]
func (h *SearchHandler) DeleteHistory(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	userID := c.GetInt64(middleware.CtxUserID)
	historyID := int64Param(c, "history_id")

	if err := h.d.SearchSvc.DeleteHistory(c.Request.Context(), wsID, userID, historyID); err != nil {
		writeErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ClearHistory 清空搜索历史。
//
//	@Summary		清空搜索历史
//	@Tags			search
//	@Success		204
//	@Router			/search/history [delete]
func (h *SearchHandler) ClearHistory(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	userID := c.GetInt64(middleware.CtxUserID)

	if err := h.d.SearchSvc.ClearHistory(c.Request.Context(), wsID, userID); err != nil {
		writeErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// --- Bookmarks ---

// ListBookmarks 列出搜索收藏。
//
//	@Summary		搜索收藏
//	@Tags			search
//	@Produce		json
//	@Success		200	{object}	searchBookmarkListResponse
//	@Router			/search/bookmarks [get]
func (h *SearchHandler) ListBookmarks(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	userID := c.GetInt64(middleware.CtxUserID)
	projectID := c.GetInt64(middleware.CtxProjectID)

	var projID *int64
	if projectID > 0 {
		projID = &projectID
	}

	bookmarks, err := h.d.SearchSvc.ListBookmarks(c.Request.Context(), wsID, userID, projID)
	if err != nil {
		writeErr(c, err)
		return
	}
	if bookmarks == nil {
		bookmarks = []SearchBookmark{}
	}
	c.JSON(http.StatusOK, gin.H{"results": bookmarks})
}

// CreateBookmark 创建搜索收藏。
//
//	@Summary		创建搜索收藏
//	@Tags			search
//	@Accept			json
//	@Produce		json
//	@Param			body	body		createBookmarkRequest	true	"收藏信息"
//	@Success		201		{object}	SearchBookmark
//	@Router			/search/bookmarks [post]
func (h *SearchHandler) CreateBookmark(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	projectID := c.GetInt64(middleware.CtxProjectID)
	userID := c.GetInt64(middleware.CtxUserID)

	var req createBookmarkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithError(c, errs.ErrValidation.WithDetails(fieldDetail(err)))
		return
	}

	var projID *int64
	if projectID > 0 {
		projID = &projectID
	}

	bm, err := h.d.SearchSvc.CreateBookmark(c.Request.Context(), CreateBookmarkInput{
		WorkspaceID: wsID,
		ProjectID:   projID,
		UserID:      userID,
		Name:        req.Name,
		Query:       req.Query,
		Filters:     req.Filters,
		IsShared:    req.IsShared,
	})
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, bm)
}

// UpdateBookmark 更新搜索收藏。
//
//	@Summary		更新搜索收藏
//	@Tags			search
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	SearchBookmark
//	@Router			/search/bookmarks/{bookmark_id} [patch]
func (h *SearchHandler) UpdateBookmark(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	userID := c.GetInt64(middleware.CtxUserID)
	bmID := int64Param(c, "bookmark_id")

	var req updateBookmarkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithError(c, errs.ErrValidation.WithDetails(fieldDetail(err)))
		return
	}

	bm, err := h.d.SearchSvc.UpdateBookmark(c.Request.Context(), wsID, userID, bmID, UpdateBookmarkInput{
		Name:     req.Name,
		Query:    req.Query,
		Filters:  req.Filters,
		IsShared: req.IsShared,
	})
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, bm)
}

// DeleteBookmark 删除搜索收藏。
//
//	@Summary		删除搜索收藏
//	@Tags			search
//	@Success		204
//	@Router			/search/bookmarks/{bookmark_id} [delete]
func (h *SearchHandler) DeleteBookmark(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	userID := c.GetInt64(middleware.CtxUserID)
	bmID := int64Param(c, "bookmark_id")

	if err := h.d.SearchSvc.DeleteBookmark(c.Request.Context(), wsID, userID, bmID); err != nil {
		writeErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// --- Request types ---

type createBookmarkRequest struct {
	Name     string         `json:"name" binding:"required,max=100"`
	Query    string         `json:"query"`
	Filters  map[string]any `json:"filters"`
	IsShared bool           `json:"is_shared"`
}

type updateBookmarkRequest struct {
	Name     *string        `json:"name"`
	Query    *string        `json:"query"`
	Filters  map[string]any `json:"filters"`
	IsShared *bool          `json:"is_shared"`
}

// --- Helpers ---

func int64Param(c *gin.Context, key string) int64 {
	v, _ := strconv.ParseInt(c.Param(key), 10, 64)
	return v
}

func intQuery(c *gin.Context, key string, def int) int {
	v, err := strconv.Atoi(c.Query(key))
	if err != nil {
		return def
	}
	return v
}

func fieldDetail(err error) errs.FieldDetail {
	return errs.FieldDetail{Field: "body", Reason: err.Error()}
}

func writeErr(c *gin.Context, err error) {
	var appErr *errs.AppError
	if errs.As(err, &appErr) {
		middleware.AbortWithError(c, appErr)
		return
	}
	middleware.AbortWithError(c, errs.ErrInternal)
}
