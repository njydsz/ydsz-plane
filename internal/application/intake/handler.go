// Package intake — 收件箱（匿名提报）HTTP handlers。
//
// 提供渠道管理（CRUD）、工单列表/详情、工单审批/转正/拒绝等管理端 REST 端点，
// 以及公开渠道查询、工单提交、工单跟踪等免登录端点。
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
//
//	@Tags	intake-public
func (h *PublicHandler) RegisterPublic(r *gin.RouterGroup) {
	g := r.Group("/intake")
	{
		g.GET("/channels/:slug", h.publicGetChannel)
		g.POST("/issues", h.publicSubmitIssue)
		g.POST("/track", h.publicTrackIssue)
	}
}

// ---- 渠道（认证） ----

// listChannels 列出收件箱渠道。
//
//	@Summary		列出收件箱渠道
//	@Description	按工作空间/项目/启用状态分页列出匿名提报渠道
//	@Tags			intake
//	@Produce		json
//	@Param			project_id	query	int		false	"项目 ID"
//	@Param			active		query	boolean	false	"仅启用"
//	@Success		200			{object}	map[string]any
//	@Router			/intake/channels [get]
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

// createChannel POST /intake/channels
//
//	@Summary		创建收件箱渠道
//	@Description	在工作空间下创建新的匿名提报渠道
//	@Tags			intake
//	@Accept			json
//	@Produce		json
//	@Success		201		{object}	IntakeChannel
//	@Failure		422		{object}	errs.AppError
//	@Router			/intake/channels [post]
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

// getChannel 获取收件箱渠道详情。
//
//	@Summary		获取渠道详情
//	@Tags			intake
//	@Produce		json
//	@Param			channel_id	path	int	true	"渠道 ID"
//	@Success		200			{object}	IntakeChannel
//	@Router			/intake/channels/{channel_id} [get]
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

// updateChannel 更新收件箱渠道。
//
//	@Summary		更新渠道
//	@Tags			intake
//	@Accept			json
//	@Produce		json
//	@Param			channel_id	path		int					true	"渠道 ID"
//	@Success		200			{object}	IntakeChannel
//	@Router			/intake/channels/{channel_id} [patch]
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

// deleteChannel 删除收件箱渠道。
//
//	@Summary		删除渠道
//	@Tags			intake
//	@Param			channel_id	path	int	true	"渠道 ID"
//	@Success		204
//	@Router			/intake/channels/{channel_id} [delete]
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
//
//	@Summary		列出收件箱工单
//	@Description	按渠道/状态分页列出匿名提报工单
//	@Tags			intake
//	@Produce		json
//	@Param			status		query		string	false	"工单状态"
//	@Param			channel_id	query		int		false	"渠道 ID"
//	@Param			project_id	query		int		false	"项目 ID"
//	@Param			limit		query		int		false	"每页数"	default(50)
//	@Param			offset		query		int		false	"偏移"	default(0)
//	@Success		200			{object}	map[string]any
//	@Router			/intake/issues [get]
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

// getIssue 获取收件箱工单详情。
//
//	@Summary		获取工单详情
//	@Tags			intake
//	@Produce		json
//	@Param			issue_id	path	int	true	"工单 ID"
//	@Success		200			{object}	IntakeIssue
//	@Router			/intake/issues/{issue_id} [get]
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

// acceptIssue 接受工单（转正）。
//
//	@Summary		接受工单
//	@Description	将匿名提报工单转正为项目内正式工作项
//	@Tags			intake
//	@Produce		json
//	@Param			issue_id	path	int	true	"工单 ID"
//	@Success		200			{object}	IntakeIssue
//	@Router			/intake/issues/{issue_id}/accept [post]
func (h *Handler) acceptIssue(c *gin.Context) {
	h.flowIssue(c, "accept")
}

// rejectIssue 拒绝工单。
//
//	@Summary		拒绝工单
//	@Tags			intake
//	@Produce		json
//	@Param			issue_id	path	int	true	"工单 ID"
//	@Success		200			{object}	IntakeIssue
//	@Router			/intake/issues/{issue_id}/reject [post]
func (h *Handler) rejectIssue(c *gin.Context) {
	h.flowIssue(c, "reject")
}

// archiveIssue 归档工单。
//
//	@Summary		归档工单
//	@Tags			intake
//	@Produce		json
//	@Param			issue_id	path	int	true	"工单 ID"
//	@Success		200			{object}	IntakeIssue
//	@Router			/intake/issues/{issue_id}/archive [post]
func (h *Handler) archiveIssue(c *gin.Context) {
	h.flowIssue(c, "archive")
}

// flowIssue 工单流转通用入口（accept/reject/archive）。
//
//	@Summary		工单流转
//	@Description	匿名提报工单流转操作（接受/拒绝/归档）
//	@Tags			intake
//	@Produce		json
//	@Param			issue_id	path	int	true	"工单 ID"
//	@Success		200			{object}	IntakeIssue
//	@Router			/intake/issues/{issue_id}/flow [post]
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

// promoteIssue 将工单提升为正式工作项。
//
//	@Summary		转正工单
//	@Description	将匿名提报工单转为项目内需求/任务/缺陷
//	@Tags			intake
//	@Accept			json
//	@Produce		json
//	@Param			issue_id	path		int					true	"工单 ID"
//	@Success		200			{object}	IntakeIssue
//	@Router			/intake/issues/{issue_id}/promote [post]
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

// publicGetChannel 按 slug 获取公开渠道信息。
//
//	@Summary		公开渠道查询
//	@Description	通过 slug 查询匿名提报渠道公开信息（免登录）
//	@Tags			intake-public
//	@Produce		json
//	@Param			slug	path	string	true	"渠道 Slug"
//	@Success		200		{object}	IntakeChannel
//	@Router			/public/intake/channels/{slug} [get]
func (h *PublicHandler) publicGetChannel(c *gin.Context) {
	ch, err := h.d.Svc.GetChannelBySlug(c.Request.Context(), c.Param("slug"))
	if err != nil {
		writeIntakeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, ch)
}

// publicSubmitIssue POST /public/intake/issues
//
//	@Summary		公开提交匿名工单
//	@Description	免登录提交匿名提报工单（无需 Authorization）
//	@Tags			intake-public
//	@Accept			json
//	@Produce		json
//	@Success		201		{object}	IntakeIssue
//	@Failure		422		{object}	errs.AppError
//	@Router			/public/intake/issues [post]
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

// publicTrackIssue 公开跟踪工单状态。
//
//	@Summary		公开跟踪工单
//	@Description	通过 tracking_id + 提交者邮箱查询工单状态（免登录）
//	@Tags			intake-public
//	@Accept			json
//	@Produce		json
//	@Success		200		{object}	IntakeIssue
//	@Router			/public/intake/track [post]
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
