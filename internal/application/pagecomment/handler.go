// Package pagecomment — 页面评论 HTTP handlers。
//
// 路由挂载：
//   - /api/v1/workspaces/:workspace_id/pages/:page_id/comments（页面级列表 / 创建）
//   - /api/v1/workspaces/:workspace_id/comments/:id（编辑 / 删除）
//   - /api/v1/workspaces/:workspace_id/comments/:id/resolve|unresolve
//   - /api/v1/workspaces/:workspace_id/comments/:id/reactions
package pagecomment

import (
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

// RegisterRead 注册只读路由（GET）。
func (h *Handler) RegisterRead(r *gin.RouterGroup) {
	// 页面级评论列表：/api/v1/workspaces/:workspace_id/pages/:page_id/comments
	pages := r.Group("/pages/:page_id/comments")
	pages.GET("", h.listComments)

	// 单条评论的回应列表：/api/v1/workspaces/:workspace_id/comments/:id/reactions
	comments := r.Group("/comments/:id")
	comments.GET("/reactions", h.listReactions)
}

// RegisterWrite 注册写入路由（POST / PATCH / DELETE）。
func (h *Handler) RegisterWrite(r *gin.RouterGroup) {
	// 页面级评论创建：/api/v1/workspaces/:workspace_id/pages/:page_id/comments
	pages := r.Group("/pages/:page_id/comments")
	pages.POST("", h.createComment)

	// 评论编辑 / 删除：/api/v1/workspaces/:workspace_id/comments/:id
	comments := r.Group("/comments/:id")
	comments.PATCH("", h.updateComment)
	comments.DELETE("", h.deleteComment)

	// 线程解决 / 取消解决
	comments.POST("/resolve", h.resolveThread)
	comments.POST("/unresolve", h.unresolveThread)

	// 审核（隐藏 / 取消隐藏）：POST /pages/comments/:id/moderate
	comments.POST("/moderate", h.moderateComment)

	// 表情回应
	comments.POST("/reactions", h.addReaction)
	comments.DELETE("/reactions/:emoji", h.removeReaction)
}

// Register 注册全部路由。
func (h *Handler) Register(r *gin.RouterGroup) {
	h.RegisterRead(r)
	h.RegisterWrite(r)
}

// --- helpers ---

// fail 统一错误响应：将领域错误映射到 HTTP 状态码。
func (h *Handler) fail(c *gin.Context, err error) {
	switch err {
	case ErrCommentNotFound, ErrReactionNotFound:
		middleware.AbortWithError(c, errs.ErrNotFound)
	case ErrNotCommentAuthor:
		middleware.AbortWithError(c, errs.ErrForbidden)
	default:
		middleware.AbortWithError(c, errs.ErrInternal)
	}
}

// intParam 解析路径参数为 int64。
func intParam(c *gin.Context, key string) (int64, error) {
	id, err := strconv.ParseInt(c.Param(key), 10, 64)
	if err != nil || id <= 0 {
		return 0, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: key, Reason: "无效的 ID"})
	}
	return id, nil
}

// intQuery 解析查询参数为 int，失败时返回默认值。
func intQuery(c *gin.Context, key string, def int) int {
	if v := c.Query(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
	}
	return def
}

// boolQuery 解析布尔型查询参数（true/1）。
func boolQuery(c *gin.Context, key string) bool {
	v := c.Query(key)
	return v == "true" || v == "1"
}

// writeErr 统一错误响应（与 pages 模块一致）。
func writeErr(c *gin.Context, err error) {
	var appErr *errs.AppError
	if errs.As(err, &appErr) {
		middleware.AbortWithError(c, appErr)
		return
	}
	middleware.AbortWithError(c, errs.ErrInternal)
}

// --- handlers ---

// listComments 列出页面下的评论。
//
//	@Summary		列出页面评论
//	@Tags			pagecomment
//	@Produce		json
//	@Param			page_id		path	int		true	"页面 ID"
//	@Param			page_type	query	string	true	"页面类型 project|knowledge"
//	@Success		200			{object}	map[string]any
//	@Router			/pages/{page_id}/comments [get]
func (h *Handler) listComments(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	pageID, err := intParam(c, "page_id")
	if err != nil {
		writeErr(c, err)
		return
	}

	pageType := PageType(c.DefaultQuery("page_type", string(PageTypeProject)))
	if !pageType.IsValid() {
		writeErr(c, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "page_type", Reason: "页面类型必须是 project 或 knowledge"}))
		return
	}

	list, err := h.d.Svc.ListComments(c.Request.Context(), ListCommentsOptions{
		WorkspaceID:   wsID,
		PageID:        pageID,
		PageType:      pageType,
		Limit:         intQuery(c, "limit", 100),
		Offset:        intQuery(c, "offset", 0),
		IncludeHidden: boolQuery(c, "include_hidden"),
		HiddenOnly:    boolQuery(c, "hidden_only"),
	})
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"results": list})
}

// createComment 创建评论（支持回复与行内定位）。
//
//	@Summary		创建页面评论
//	@Tags			pagecomment
//	@Accept			json
//	@Produce		json
//	@Param			page_id		path	int					true	"页面 ID"
//	@Param			page_type	query	string				false	"页面类型 project|knowledge"
//	@Success		201				{object}	PageComment
//	@Router			/pages/{page_id}/comments [post]
func (h *Handler) createComment(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	userID := c.GetInt64(middleware.CtxUserID)
	pageID, err := intParam(c, "page_id")
	if err != nil {
		writeErr(c, err)
		return
	}

	pageType := PageType(c.DefaultQuery("page_type", string(PageTypeProject)))
	if !pageType.IsValid() {
		writeErr(c, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "page_type", Reason: "页面类型必须是 project 或 knowledge"}))
		return
	}

	var req struct {
		Content   string  `json:"content"`
		ParentID  *int64  `json:"parent_id"`
		BlockID   string  `json:"block_id"`
		QuoteText string  `json:"quote_text"`
		Mentions  []int64 `json:"mentions"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		writeErr(c, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "body", Reason: err.Error()}))
		return
	}

	comment, err := h.d.Svc.CreateComment(c.Request.Context(), CreateCommentInput{
		WorkspaceID: wsID,
		PageID:      pageID,
		PageType:    pageType,
		ParentID:    req.ParentID,
		AuthorID:    userID,
		Content:     req.Content,
		BlockID:     req.BlockID,
		QuoteText:   req.QuoteText,
		Mentions:    req.Mentions,
	})
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, comment)
}

// updateComment 编辑评论。
//
//	@Summary		编辑评论
//	@Tags			pagecomment
//	@Accept			json
//	@Produce		json
//	@Param			id	path	int	true	"评论 ID"
//	@Success		200	{object}	PageComment
//	@Router			/comments/{id} [patch]
func (h *Handler) updateComment(c *gin.Context) {
	commentID, err := intParam(c, "id")
	if err != nil {
		writeErr(c, err)
		return
	}
	userID := c.GetInt64(middleware.CtxUserID)

	var req struct {
		Content  *string `json:"content"`
		Mentions []int64 `json:"mentions"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		writeErr(c, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "body", Reason: err.Error()}))
		return
	}

	comment, err := h.d.Svc.UpdateComment(c.Request.Context(), commentID, userID, UpdateCommentInput{
		Content:  req.Content,
		Mentions: req.Mentions,
	})
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, comment)
}

// deleteComment 软删除评论。
//
//	@Summary		删除评论
//	@Tags			pagecomment
//	@Param			id	path	int	true	"评论 ID"
//	@Success		204
//	@Router			/comments/{id} [delete]
func (h *Handler) deleteComment(c *gin.Context) {
	commentID, err := intParam(c, "id")
	if err != nil {
		writeErr(c, err)
		return
	}
	userID := c.GetInt64(middleware.CtxUserID)

	if err := h.d.Svc.DeleteComment(c.Request.Context(), commentID, userID); err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// resolveThread 标记线程已解决。
//
//	@Summary		解决评论线程
//	@Tags			pagecomment
//	@Param			id	path	int	true	"评论 ID"
//	@Success		200	{object}	map[string]bool
//	@Router			/comments/{id}/resolve [post]
func (h *Handler) resolveThread(c *gin.Context) {
	commentID, err := intParam(c, "id")
	if err != nil {
		writeErr(c, err)
		return
	}
	userID := c.GetInt64(middleware.CtxUserID)

	if err := h.d.Svc.ResolveThread(c.Request.Context(), commentID, userID); err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// unresolveThread 取消线程解决状态。
//
//	@Summary		取消解决评论线程
//	@Tags			pagecomment
//	@Param			id	path	int	true	"评论 ID"
//	@Success		200	{object}	map[string]bool
//	@Router			/comments/{id}/unresolve [post]
func (h *Handler) unresolveThread(c *gin.Context) {
	commentID, err := intParam(c, "id")
	if err != nil {
		writeErr(c, err)
		return
	}
	userID := c.GetInt64(middleware.CtxUserID)

	if err := h.d.Svc.UnresolveThread(c.Request.Context(), commentID, userID); err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// moderateComment 审核评论（隐藏 / 取消隐藏）。
//
//	@Summary		审核评论
//	@Tags			pagecomment
//	@Accept			json
//	@Produce		json
//	@Param			id		path	int		true	"评论 ID"
//	@Param			body	body	moderationRequest	true	"审核动作"
//	@Success		200		{object}	map[string]bool
//	@Router			/comments/{id}/moderate [post]
func (h *Handler) moderateComment(c *gin.Context) {
	commentID, err := intParam(c, "id")
	if err != nil {
		writeErr(c, err)
		return
	}
	userID := c.GetInt64(middleware.CtxUserID)

	var req struct {
		Action string `json:"action"`
		Reason string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		writeErr(c, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "body", Reason: err.Error()}))
		return
	}

	if err := h.d.Svc.ModerateComment(c.Request.Context(), commentID, userID, req.Action, req.Reason); err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// addReaction 添加表情回应。
//
//	@Summary		添加评论表情回应
//	@Tags			pagecomment
//	@Accept			json
//	@Produce		json
//	@Param			id	path	int	true	"评论 ID"
//	@Success		201	{object}	PageCommentReaction
//	@Router			/comments/{id}/reactions [post]
func (h *Handler) addReaction(c *gin.Context) {
	commentID, err := intParam(c, "id")
	if err != nil {
		writeErr(c, err)
		return
	}
	userID := c.GetInt64(middleware.CtxUserID)

	var req struct {
		Emoji string `json:"emoji"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		writeErr(c, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "body", Reason: err.Error()}))
		return
	}

	reaction, err := h.d.Svc.AddReaction(c.Request.Context(), commentID, userID, req.Emoji)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, reaction)
}

// removeReaction 移除表情回应。
//
//	@Summary		移除评论表情回应
//	@Tags			pagecomment
//	@Param			id		path	int		true	"评论 ID"
//	@Param			emoji	path	string	true	"表情标识"
//	@Success		200		{object}	map[string]bool
//	@Router			/comments/{id}/reactions/{emoji} [delete]
func (h *Handler) removeReaction(c *gin.Context) {
	commentID, err := intParam(c, "id")
	if err != nil {
		writeErr(c, err)
		return
	}
	userID := c.GetInt64(middleware.CtxUserID)
	emoji := c.Param("emoji")
	if emoji == "" {
		writeErr(c, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "emoji", Reason: "表情标识不能为空"}))
		return
	}

	if err := h.d.Svc.RemoveReaction(c.Request.Context(), commentID, userID, emoji); err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// listReactions 列出评论的表情回应。
//
//	@Summary		列出评论表情回应
//	@Tags			pagecomment
//	@Produce		json
//	@Param			id	path	int	true	"评论 ID"
//	@Success		200	{object}	map[string]any
//	@Router			/comments/{id}/reactions [get]
func (h *Handler) listReactions(c *gin.Context) {
	commentID, err := intParam(c, "id")
	if err != nil {
		writeErr(c, err)
		return
	}

	reactions, err := h.d.Svc.ListReactions(c.Request.Context(), commentID)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"results": reactions})
}
