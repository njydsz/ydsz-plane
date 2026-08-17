// Package intake — 收件箱 HTTP handlers（认证路由 + 公开路由）。
package intake

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/njydsz/ydsz-plane/internal/interfaces/middleware"
	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// HandlerDeps 收件箱 handler 依赖。
type HandlerDeps struct {
	Svc *Service
}

// Handler 认证路由 handler（工作空间级）。
type Handler struct {
	d *HandlerDeps
}

// NewHandler 构造认证 handler。
func NewHandler(d *HandlerDeps) *Handler {
	return &Handler{d: d}
}

// PublicHandler 公开（免登）路由 handler。
type PublicHandler struct {
	d *HandlerDeps
}

// NewPublicHandler 构造公开 handler。
func NewPublicHandler(d *HandlerDeps) *PublicHandler {
	return &PublicHandler{d: d}
}

// Register 注册全部认证路由（读+写，供内部/简化装配使用）。
func (h *Handler) Register(r *gin.RouterGroup) {
	h.RegisterRead(r)
	h.RegisterWrite(r)
}

// RegisterRead 注册只读路由（渠道列表/详情、工单列表/详情）。
func (h *Handler) RegisterRead(r *gin.RouterGroup) {
	g := r.Group("/intake")
	{
		g.GET("/channels", h.listChannels)
		g.GET("/channels/:channel_id", h.getChannel)
		g.GET("/issues", h.listIssues)
		g.GET("/issues/:issue_id", h.getIssue)
	}
}

// RegisterWrite 注册写路由（渠道管理、工单流转/转正）。
func (h *Handler) RegisterWrite(r *gin.RouterGroup) {
	g := r.Group("/intake")
	{
		g.POST("/channels", h.createChannel)
		g.PATCH("/channels/:channel_id", h.updateChannel)
		g.DELETE("/channels/:channel_id", h.deleteChannel)

		g.POST("/issues/:issue_id/accept", h.acceptIssue)
		g.POST("/issues/:issue_id/reject", h.rejectIssue)
		g.POST("/issues/:issue_id/archive", h.archiveIssue)
		g.POST("/issues/:issue_id/promote", h.promoteIssue)
	}
}

// RegisterPublic 注册公开路由（/api/v1/public/intake 前缀，免登）。
func (h *PublicHandler) RegisterPublic(r *gin.RouterGroup) {
	g := r.Group("/intake")
	{
		g.GET("/channels/:slug", h.publicGetChannel)
		g.POST("/issues", h.publicSubmitIssue)
		g.POST("/track", h.publicTrackIssue)
	}
}

// ---- 渠道（认证） ----

// listChannels GET /intake/channels?project_id=&active=
func (h *Handler) listChannels(c *gin.Context) {
	wsID := wsID(c)
	var projectID *int64
	if v, err := strconv.ParseInt(c.Query("project_id"), 10, 64); err == nil && v > 0 {
		projectID = &v
	}
	channels, err := h.d.Svc.ListChannels(c.Request.Context(), wsID, projectID, c.Query("active") == "true")
	if err != nil {
		writeIntakeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"results": channels, "total": len(channels)})
}

// createChannel POST /intake/channels
type createChannelRequest struct {
	Name        string         `json:"name" binding:"required"`
	Description string         `json:"description"`
	Slug        string         `json:"slug"`
	ProjectID   *int64         `json:"project_id"`
	Config      map[string]any `json:"config"`
}

func (h *Handler) createChannel(c *gin.Context) {
	var req createChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeIntakeErr(c, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "body", Reason: "请求体无效: " + err.Error()}))
		return
	}
	ch, err := h.d.Svc.CreateChannel(c.Request.Context(), CreateChannelInput{
		WorkspaceID: wsID(c),
		ProjectID:   req.ProjectID,
		Name:        req.Name,
		Description: req.Description,
		Slug:        req.Slug,
		Config:      req.Config,
		CreatedBy:   actorID(c),
	})
	if err != nil {
		writeIntakeErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, ch)
}

// getChannel GET /intake/channels/:channel_id
func (h *Handler) getChannel(c *gin.Context) {
	id, err := idParam(c, "channel_id")
	if err != nil {
		writeIntakeErr(c, err)
		return
	}
	ch, err := h.d.Svc.GetChannel(c.Request.Context(), wsID(c), id)
	if err != nil {
		writeIntakeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, ch)
}

// updateChannel PATCH /intake/channels/:channel_id
type updateChannelRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Slug        *string `json:"slug"`
	IsActive    *bool   `json:"is_active"`
	ProjectID   *int64  `json:"project_id"`
}

func (h *Handler) updateChannel(c *gin.Context) {
	id, err := idParam(c, "channel_id")
	if err != nil {
		writeIntakeErr(c, err)
		return
	}
	var req updateChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeIntakeErr(c, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "body", Reason: "请求体无效: " + err.Error()}))
		return
	}
	ch, err := h.d.Svc.UpdateChannel(c.Request.Context(), UpdateChannelInput{
		ID:          id,
		WorkspaceID: wsID(c),
		Name:        req.Name,
		Description: req.Description,
		Slug:        req.Slug,
		IsActive:    req.IsActive,
		ProjectID:   req.ProjectID,
	})
	if err != nil {
		writeIntakeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, ch)
}

// deleteChannel DELETE /intake/channels/:channel_id
func (h *Handler) deleteChannel(c *gin.Context) {
	id, err := idParam(c, "channel_id")
	if err != nil {
		writeIntakeErr(c, err)
		return
	}
	if err := h.d.Svc.DeleteChannel(c.Request.Context(), wsID(c), id); err != nil {
		writeIntakeErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ---- 工单（认证） ----

// listIssues GET /intake/issues?status=&channel_id=&project_id=&limit=&offset=
func (h *Handler) listIssues(c *gin.Context) {
	f := ListIssuesFilter{WorkspaceID: wsID(c), Status: c.Query("status")}
	if v, err := strconv.ParseInt(c.Query("channel_id"), 10, 64); err == nil && v > 0 {
		f.ChannelID = &v
	}
	if v, err := strconv.ParseInt(c.Query("project_id"), 10, 64); err == nil && v > 0 {
		f.ProjectID = &v
	}
	if v, err := strconv.Atoi(c.Query("limit")); err == nil && v > 0 {
		f.Limit = v
	}
	if v, err := strconv.Atoi(c.Query("offset")); err == nil && v >= 0 {
		f.Offset = v
	}
	issues, total, err := h.d.Svc.ListIssues(c.Request.Context(), f)
	if err != nil {
		writeIntakeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"results": issues, "total": total})
}

// getIssue GET /intake/issues/:issue_id
func (h *Handler) getIssue(c *gin.Context) {
	id, err := idParam(c, "issue_id")
	if err != nil {
		writeIntakeErr(c, err)
		return
	}
	it, err := h.d.Svc.GetIssue(c.Request.Context(), wsID(c), id)
	if err != nil {
		writeIntakeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, it)
}

// acceptIssue POST /intake/issues/:issue_id/accept
func (h *Handler) acceptIssue(c *gin.Context) {
	h.flowIssue(c, "accept")
}

// rejectIssue POST /intake/issues/:issue_id/reject
func (h *Handler) rejectIssue(c *gin.Context) {
	h.flowIssue(c, "reject")
}

// archiveIssue POST /intake/issues/:issue_id/archive
func (h *Handler) archiveIssue(c *gin.Context) {
	h.flowIssue(c, "archive")
}

func (h *Handler) flowIssue(c *gin.Context, action string) {
	id, err := idParam(c, "issue_id")
	if err != nil {
		writeIntakeErr(c, err)
		return
	}
	var it *IntakeIssue
	switch action {
	case "accept":
		it, err = h.d.Svc.AcceptIssue(c.Request.Context(), wsID(c), id, actorID(c))
	case "reject":
		it, err = h.d.Svc.RejectIssue(c.Request.Context(), wsID(c), id, actorID(c))
	case "archive":
		it, err = h.d.Svc.ArchiveIssue(c.Request.Context(), wsID(c), id, actorID(c))
	}
	if err != nil {
		writeIntakeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, it)
}

// promoteIssue POST /intake/issues/:issue_id/promote
type promoteIssueRequest struct {
	TypeCode   string  `json:"type_code"`
	Severity   *int    `json:"severity"`
	FoundPhase *string `json:"found_phase"`
	ProjectID  *int64  `json:"project_id"`
}

func (h *Handler) promoteIssue(c *gin.Context) {
	id, err := idParam(c, "issue_id")
	if err != nil {
		writeIntakeErr(c, err)
		return
	}
	var req promoteIssueRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeIntakeErr(c, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "body", Reason: "请求体无效: " + err.Error()}))
		return
	}
	it, err := h.d.Svc.PromoteIssue(c.Request.Context(), PromoteIssueInput{
		WorkspaceID: wsID(c),
		IssueID:     id,
		Operator:    actorID(c),
		TypeCode:    req.TypeCode,
		Severity:    req.Severity,
		FoundPhase:  req.FoundPhase,
		ProjectID:   req.ProjectID,
	})
	if err != nil {
		writeIntakeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, it)
}

// ---- 公开（免登） ----

// publicGetChannel GET /public/intake/channels/:slug
func (h *PublicHandler) publicGetChannel(c *gin.Context) {
	ch, err := h.d.Svc.GetChannelBySlug(c.Request.Context(), c.Param("slug"))
	if err != nil {
		writeIntakeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, ch)
}

// publicSubmitIssue POST /public/intake/issues
func (h *PublicHandler) publicSubmitIssue(c *gin.Context) {
	var req SubmitIssueInput
	if err := c.ShouldBindJSON(&req); err != nil {
		writeIntakeErr(c, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "body", Reason: "请求体无效: " + err.Error()}))
		return
	}
	// 公开提交不允许指定渠道 ID（防越权探测），仅允许 slug。
	req.ChannelID = 0
	it, err := h.d.Svc.SubmitIssue(c.Request.Context(), req)
	if err != nil {
		writeIntakeErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, it)
}

// publicTrackIssue POST /public/intake/track
type trackIssueRequest struct {
	TrackingID     string `json:"tracking_id" binding:"required"`
	SubmitterEmail string `json:"submitter_email" binding:"required"`
}

func (h *PublicHandler) publicTrackIssue(c *gin.Context) {
	var req trackIssueRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeIntakeErr(c, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "body", Reason: "请求体无效: " + err.Error()}))
		return
	}
	it, err := h.d.Svc.TrackIssue(c.Request.Context(), req.TrackingID, req.SubmitterEmail)
	if err != nil {
		writeIntakeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, it)
}

// ---- helpers ----

func wsID(c *gin.Context) int64 { return c.GetInt64(middleware.CtxWorkspaceID) }

func actorID(c *gin.Context) int64 { return c.GetInt64(middleware.CtxUserID) }

func idParam(c *gin.Context, name string) (int64, error) {
	v, err := strconv.ParseInt(c.Param(name), 10, 64)
	if err != nil || v <= 0 {
		return 0, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: name, Reason: "无效的 ID"})
	}
	return v, nil
}

func writeIntakeErr(c *gin.Context, err error) {
	var appErr *errs.AppError
	if errs.As(err, &appErr) {
		middleware.AbortWithError(c, appErr)
		return
	}
	middleware.AbortWithError(c, errs.ErrInternal)
}
