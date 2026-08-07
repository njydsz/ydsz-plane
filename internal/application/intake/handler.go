// Package intake — Intake HTTP 处理器：通道管理、工单审核、公开提交、跟踪查询。
//
// 路由规划：
//   - /api/v1/workspaces/:workspace_id/intake/channels        （管理员：CRUD）
//   - /api/v1/workspaces/:workspace_id/intake/issues          （管理员：列表/审核）
//   - /api/v1/public/intake/:workspace_slug/:slug/submit     （公开提交，免登录）
//   - /api/v1/public/intake/track                             （提交者跟踪，邮箱+ID 查看）
package intake

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/njydsz/ydsz-plane/internal/application/auth"
	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// HandlerDeps 是 intake 处理器依赖。
type HandlerDeps struct {
	IntakeSvc      *Service
	WorkspaceStore *auth.WorkspaceMembershipStore
}

// Handler 是 intake HTTP handler。
type Handler struct {
	svc     *Service
	wsStore *auth.WorkspaceMembershipStore
}

// NewHandler 构造 intake handler。
func NewHandler(deps *HandlerDeps) *Handler {
	return &Handler{
		svc:     deps.IntakeSvc,
		wsStore: deps.WorkspaceStore,
	}
}

// Register 注册管理员路由（工作空间级）。
func (h *Handler) Register(r *gin.RouterGroup) {
	// 通道 CRUD
	channels := r.Group("/channels")
	{
		channels.POST("", h.CreateChannel)
		channels.GET("", h.ListChannels)
		channels.GET("/:channel_id", h.GetChannel)
		channels.PATCH("/:channel_id", h.UpdateChannel)
		channels.DELETE("/:channel_id", h.DeleteChannel)
	}

	// 收件工单审核
	issues := r.Group("/issues")
	{
		issues.GET("", h.ListIssues)
		issues.GET("/:issue_id", h.GetIssue)
		issues.POST("/:issue_id/review", h.ReviewIssue)
	}
}

// --- 通道管理 ---

func (h *Handler) CreateChannel(c *gin.Context) {
	wsID := c.GetInt64("workspace_id")
	userID := c.GetInt64("user_id")

	var req createChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, errs.Validation("INTAKE.INVALID_REQUEST", "无效请求"))
		return
	}

	input := CreateChannelInput{
		WorkspaceID:      wsID,
		ProjectID:        req.ProjectID,
		Slug:             req.Slug,
		Name:             req.Name,
		Description:      req.Description,
		IsPublic:         req.IsPublic,
		DefaultIssueType: req.DefaultIssueType,
		DefaultPriority:  req.DefaultPriority,
		AutoAssignRules:  req.AutoAssignRules,
		RateLimitPerMin:  req.RateLimitPerMin,
		RequireCaptcha:   req.RequireCaptcha,
		CustomFields:     req.CustomFields,
		Branding:         req.Branding,
		NotifyOnSubmit:   req.NotifyOnSubmit,
		NotifyUsers:      req.NotifyUsers,
		CreatedBy:        userID,
	}

	ch, err := h.svc.CreateChannel(c.Request.Context(), input)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, ch)
}

func (h *Handler) ListChannels(c *gin.Context) {
	wsID := c.GetInt64("workspace_id")
	input := ListChannelsInput{WorkspaceID: wsID}

	if v := c.Query("project_id"); v != "" {
		if id, err := strconv.ParseInt(v, 10, 64); err == nil {
			input.ProjectID = &id
		}
	}
	if v := c.Query("limit"); v != "" {
		n, _ := strconv.Atoi(v)
		input.Limit = n
	}
	if v := c.Query("offset"); v != "" {
		n, _ := strconv.Atoi(v)
		input.Offset = n
	}

	result, err := h.svc.ListChannels(c.Request.Context(), input)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"items":  result.Items,
		"total":  result.Total,
		"limit":  input.Limit,
		"offset": input.Offset,
	})
}

func (h *Handler) GetChannel(c *gin.Context) {
	wsID := c.GetInt64("workspace_id")
	channelID := c.GetInt64("channel_id")

	ch, err := h.svc.GetChannel(c.Request.Context(), wsID, channelID)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, ch)
}

func (h *Handler) UpdateChannel(c *gin.Context) {
	wsID := c.GetInt64("workspace_id")
	channelID := c.GetInt64("channel_id")

	var req updateChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, errs.Validation("INTAKE.INVALID_REQUEST", "无效请求"))
		return
	}

	input := UpdateChannelInput{
		Name:            req.Name,
		Description:     req.Description,
		IsPublic:        req.IsPublic,
		IsActive:        req.IsActive,
		AutoAssignRules: req.AutoAssignRules,
	}

	ch, err := h.svc.UpdateChannel(c.Request.Context(), wsID, channelID, input)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, ch)
}

func (h *Handler) DeleteChannel(c *gin.Context) {
	wsID := c.GetInt64("workspace_id")
	channelID := c.GetInt64("channel_id")

	if err := h.svc.DeleteChannel(c.Request.Context(), wsID, channelID); err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// --- 工单审核 ---

func (h *Handler) ListIssues(c *gin.Context) {
	wsID := c.GetInt64("workspace_id")
	input := ListIssuesInput{WorkspaceID: wsID}

	if v := c.Query("channel_id"); v != "" {
		if id, err := strconv.ParseInt(v, 10, 64); err == nil {
			input.ChannelID = &id
		}
	}
	if v := c.Query("status"); v != "" {
		input.Status = &v
	}
	if v := c.Query("limit"); v != "" {
		n, _ := strconv.Atoi(v)
		input.Limit = n
	}
	if v := c.Query("offset"); v != "" {
		n, _ := strconv.Atoi(v)
		input.Offset = n
	}

	result, err := h.svc.ListIssues(c.Request.Context(), input)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"items":  result.Items,
		"total":  result.Total,
		"limit":  input.Limit,
		"offset": input.Offset,
	})
}

func (h *Handler) GetIssue(c *gin.Context) {
	wsID := c.GetInt64("workspace_id")
	issueID := c.GetInt64("issue_id")

	is, err := h.svc.GetIssue(c.Request.Context(), wsID, issueID)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, is)
}

func (h *Handler) ReviewIssue(c *gin.Context) {
	wsID := c.GetInt64("workspace_id")
	issueID := c.GetInt64("issue_id")
	userID := c.GetInt64("user_id")

	var req reviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, errs.Validation("INTAKE.INVALID_REQUEST", "无效请求"))
		return
	}

	decision := ReviewDecision{
		Action:          req.Action,
		TargetIssueType: req.TargetIssueType,
		TargetProjectID:  req.TargetProjectID,
		Reason:          req.Reason,
		ReviewerID:      userID,
	}

	is, err := h.svc.ReviewIssue(c.Request.Context(), wsID, issueID, decision)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, is)
}

// --- 公开路由处理器（免登录） ---

// PublicHandler 公开提交与跟踪处理器。
type PublicHandler struct {
	svc *Service
}

// NewPublicHandler 构造公开 handler。
func NewPublicHandler(svc *Service) *PublicHandler {
	return &PublicHandler{svc: svc}
}

// PublicChannelView 公开通道视图（脱敏）。
type PublicChannelView struct {
	ID               int64           `json:"id"`
	Slug             string          `json:"slug"`
	Name             string          `json:"name"`
	Description      string          `json:"description,omitempty"`
	DefaultIssueType string          `json:"default_issue_type"`
	RequireCaptcha   bool            `json:"require_captcha"`
	CustomFields     json.RawMessage `json:"custom_fields"`
	Branding         json.RawMessage `json:"branding"`
}

// GetPublicChannel 公开获取通道配置（渲染表单）。
func (h *PublicHandler) GetPublicChannel(c *gin.Context) {
	wsID, err := resolveWorkspaceID(c, h.svc.db, c.Param("workspace"))
	if err != nil || wsID == 0 {
		respondError(c, errs.NotFound("WORKSPACE.NOT_FOUND", "工作空间不存在"))
		return
	}

	slug := c.Param("slug")
	ch, err := h.svc.GetChannelBySlug(c.Request.Context(), wsID, slug)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, PublicChannelView{
		ID:               ch.ID,
		Slug:             ch.Slug,
		Name:             ch.Name,
		Description:      ch.Description,
		DefaultIssueType: ch.DefaultIssueType,
		RequireCaptcha:   ch.RequireCaptcha,
		CustomFields:     ch.CustomFields,
		Branding:         ch.Branding,
	})
}

// SubmitPublicIssue 公开提交工单（免登录）。
func (h *PublicHandler) SubmitPublicIssue(c *gin.Context) {
	wsID, err := resolveWorkspaceID(c, h.svc.db, c.Param("workspace"))
	if err != nil || wsID == 0 {
		respondError(c, errs.NotFound("WORKSPACE.NOT_FOUND", "工作空间不存在"))
		return
	}

	slug := c.Param("slug")
	ch, err := h.svc.GetChannelBySlug(c.Request.Context(), wsID, slug)
	if err != nil {
		respondError(c, err)
		return
	}

	var req submitPublicRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, errs.Validation("INTAKE.INVALID_REQUEST", "无效请求"))
		return
	}

	// 限流检查
	allowed, err := h.svc.CheckRateLimit(c.Request.Context(), ch.WorkspaceID, ch.ID, c.ClientIP(), ch.RateLimitPerMin)
	if err != nil || !allowed {
		respondError(c, errs.Validation("INTAKE.RATE_LIMITED", "提交过快，请稍后再试"))
		return
	}

	it := req.IssueType
	if it == "" {
		it = ch.DefaultIssueType
	}

	input := SubmitInput{
		ChannelID:      ch.ID,
		WorkspaceID:    ch.WorkspaceID,
		Title:          req.Title,
		Description:    req.Description,
		SubmitterName:  req.SubmitterName,
		SubmitterEmail: req.SubmitterEmail,
		IssueType:      it,
		Priority:       ch.DefaultPriority,
		CustomFields:   req.CustomFields,
		AttachmentIDs:  req.AttachmentIDs,
	}

	iss, err := h.svc.SubmitIssue(c.Request.Context(), input)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"tracking_id":  iss.TrackingID,
		"status":       iss.Status,
		"submitted_at": iss.CreatedAt,
		"message":      "提交成功，我们会尽快处理",
	})
}

// TrackIssue 提交者跟踪工单状态。
func (h *PublicHandler) TrackIssue(c *gin.Context) {
	trackingID := c.Query("tracking_id")
	email := c.Query("email")
	if trackingID == "" || email == "" {
		respondError(c, errs.Validation("INTAKE.BAD_REQUEST", "请提供 tracking_id 和 email"))
		return
	}

	view, err := h.svc.GetPublicView(c.Request.Context(), trackingID, email)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, view)
}

// --- 请求 DTO ---

type createChannelRequest struct {
	Slug             string          `json:"slug" binding:"required"`
	Name             string          `json:"name" binding:"required"`
	Description      string          `json:"description,omitempty"`
	IsPublic         bool            `json:"is_public"`
	DefaultIssueType string          `json:"default_issue_type"`
	DefaultPriority  int16           `json:"default_priority"`
	AutoAssignRules  json.RawMessage `json:"auto_assign_rules,omitempty"`
	RateLimitPerMin  int16           `json:"rate_limit_per_min"`
	RequireCaptcha   bool            `json:"require_captcha"`
	CustomFields     json.RawMessage `json:"custom_fields,omitempty"`
	Branding         json.RawMessage `json:"branding,omitempty"`
	NotifyOnSubmit   bool            `json:"notify_on_submit"`
	NotifyUsers      []int64         `json:"notify_users,omitempty"`
	ProjectID        *int64          `json:"project_id,omitempty"`
}

type updateChannelRequest struct {
	Name            *string         `json:"name,omitempty"`
	Description     *string         `json:"description,omitempty"`
	IsPublic        *bool           `json:"is_public,omitempty"`
	IsActive        *bool           `json:"is_active,omitempty"`
	AutoAssignRules json.RawMessage `json:"auto_assign_rules,omitempty"`
}

type reviewRequest struct {
	Action          string `json:"action" binding:"required"`
	TargetIssueType string `json:"target_issue_type,omitempty"`
	TargetProjectID *int64 `json:"target_project_id,omitempty"`
	Reason          string `json:"reason,omitempty"`
}

type submitPublicRequest struct {
	Title          string          `json:"title" binding:"required"`
	Description    string          `json:"description"`
	SubmitterName  string          `json:"submitter_name" binding:"required"`
	SubmitterEmail string          `json:"submitter_email" binding:"required,email"`
	CustomFields   json.RawMessage `json:"custom_fields,omitempty"`
	AttachmentIDs  []int64         `json:"attachment_ids,omitempty"`
	IssueType      string          `json:"issue_type,omitempty"`
}

// --- 辅助函数 ---

// resolveWorkspaceID 从 URL 参数解析工作空间 ID（支持 slug）。
// 对外公开接口使用 slug 避免暴露内部 ID。
func resolveWorkspaceID(c *gin.Context, db *sql.DB, key string) (int64, error) {
	if key == "" {
		return 0, fmt.Errorf("missing workspace key")
	}
	// 尝试解析为 digital ID
	if id, err := strconv.ParseInt(key, 10, 64); err == nil {
		return id, nil
	}
	// 当作 slug 查询
	var id int64
	if err := db.QueryRowContext(c.Request.Context(),
		`SELECT id FROM workspaces WHERE slug = $1 AND is_active = true`, key).Scan(&id); err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, err
	}
	return id, nil
}

func respondError(c *gin.Context, err error) {
	var appErr *errs.AppError
	if errs.As(err, &appErr) {
		c.JSON(appErr.HTTP, gin.H{"error": appErr})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{
		"error": gin.H{"code": "INTERNAL.ERROR", "message": "服务内部错误"},
	})
}

var _ = context.Background // keep import
