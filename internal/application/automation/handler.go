// Package automation 提供自动化规则引擎 HTTP Handler。
package automation

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/njydsz/ydsz-plane/internal/interfaces/middleware"
	"github.com/njydsz/ydsz-plane/pkg/errs"
)

type HandlerDeps struct {
	Svc *Service
}

// Handler 是自动化规则的 HTTP 处理器。
type Handler struct {
	deps *HandlerDeps
}

// NewHandler 创建自动化规则处理器。
func NewHandler(deps *HandlerDeps) *Handler {
	return &Handler{deps: deps}
}

// Register 注册自动化规则路由（项目级）。
func (h *Handler) Register(r gin.IRoutes) {
	// CRUD
	r.GET("", h.List)
	r.POST("", h.Create)
	r.GET("/templates", h.ListTemplates)
	r.POST("/dry-run", h.DryRun)
	r.POST("/from-template", h.CreateFromTemplate)
	r.GET("/executions", h.ListExecutions)
	r.GET("/:rule_id", h.Get)
	r.PATCH("/:rule_id", h.Update)
	r.DELETE("/:rule_id", h.Delete)
	r.POST("/:rule_id/toggle", h.Toggle)
}

// --- Query params ---

type listRulesQuery struct {
	Status      *string `form:"status" binding:"omitempty,oneof=draft active disabled error"`
	TriggerType *string `form:"trigger_type"`
	Limit       int     `form:"limit" binding:"omitempty,min=1,max=100"`
	Offset      int     `form:"offset" binding:"omitempty,min=0"`
}

// --- Handlers ---

// List godoc
//
//	@Summary		列出自动化规则
//	@Description	查询项目级 + 工作空间级自动化规则
//	@Tags			automation
//	@Produce		json
//	@Param			limit			query		int				false	"每页数 (1-100)"	default(50)
//	@Param			offset			query		int				false	"偏移"			default(0)
//	@Param			status			query		string	false	"状态"
//	@Success		200				{object}	ruleListResponse
//	@Security		Bearer
//	@Router			/automation [get]
func (h *Handler) List(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	projectID := c.GetInt64(middleware.CtxProjectID)

	var q listRulesQuery
	_ = c.ShouldBindQuery(&q)

	opts := ListRulesOptions{
		WorkspaceID: wsID,
		ProjectID:   &projectID,
		Limit:       q.Limit,
		Offset:      q.Offset,
	}
	if q.Status != nil {
		s := RuleStatus(*q.Status)
		opts.Status = &s
	}
	if q.TriggerType != nil {
		opts.TriggerType = q.TriggerType
	}

	rules, total, err := h.deps.Svc.List(c.Request.Context(), wsID, opts)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, ruleListResponse{Total: total, Items: rules})
}

// Create godoc
//
//	@Summary		创建自动化规则
//	@Description	创建一条新的自动化规则（含 DSL 校验）
//	@Tags			automation
//	@Accept			json
//	@Produce		json
//	@Param			body	body		createRuleRequest	true	"规则定义"
//	@Success		201		{object}	Rule
//	@Security		Bearer
//	@Router			/automation [post]
func (h *Handler) Create(c *gin.Context) {
	var req createRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithError(c, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "body", Reason: err.Error()}))
		return
	}

	userID := c.GetInt64(middleware.CtxUserID)
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	projectID := c.GetInt64(middleware.CtxProjectID)

	in := CreateRuleInput{
		WorkspaceID: wsID,
		ProjectID:   &projectID,
		Name:        req.Name,
		Description: req.Description,
		DSL:         req.DSL,
		Status:      req.Status,
		CreatedBy:   userID,
	}
	if in.Status == "" {
		in.Status = RuleStatusDraft
	}

	rule, err := h.deps.Svc.Create(c.Request.Context(), in)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, rule)
}

// CreateFromTemplate godoc
//
//	@Summary		从模板创建规则
//	@Description	选择内置模板克隆一条新规则
//	@Tags			automation
//	@Accept			json
//	@Produce		json
//	@Param			body	body		createFromTemplateRequest	true	"模板参数"
//	@Success		201		{object}	Rule
//	@Security		Bearer
//	@Router			/automation/from-template [post]
func (h *Handler) CreateFromTemplate(c *gin.Context) {
	var req createFromTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithError(c, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "body", Reason: err.Error()}))
		return
	}

	userID := c.GetInt64(middleware.CtxUserID)
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	projectID := c.GetInt64(middleware.CtxProjectID)

	in := CreateRuleInput{
		WorkspaceID: wsID,
		ProjectID:   &projectID,
		Name:        req.Name,
		Description: req.Description,
		Status:      RuleStatusDraft,
		CreatedBy:   userID,
	}

	rule, err := h.deps.Svc.CreateFromTemplate(c.Request.Context(), in, req.TemplateSlug)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, rule)
}

// Get godoc
//
//	@Summary		获取规则详情
//	@Description	通过 ID 查询单条规则
//	@Tags			automation
//	@Produce		json
//	@Param			rule_id		path		int	true	"规则ID"
//	@Success		200			{object}	Rule
//	@Security		Bearer
//	@Router			/automation/{rule_id} [get]
func (h *Handler) Get(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	ruleID, err := parseInt64Param(c, "rule_id")
	if err != nil {
		middleware.AbortWithError(c, errs.ErrValidation)
		return
	}

	rule, err := h.deps.Svc.GetByID(c.Request.Context(), wsID, ruleID)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, rule)
}

// Update godoc
//
//	@Summary		更新规则
//	@Description	部分更新规则（含 DSL 校验）
//	@Tags			automation
//	@Accept			json
//	@Produce		json
//	@Param			rule_id		path		int					true	"规则ID"
//	@Param			body		body		updateRuleRequest	true	"更新字段"
//	@Success		200			{object}	Rule
//	@Security		Bearer
//	@Router			/automation/{rule_id} [patch]
func (h *Handler) Update(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	ruleID, err := parseInt64Param(c, "rule_id")
	if err != nil {
		middleware.AbortWithError(c, errs.ErrValidation)
		return
	}

	var req updateRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithError(c, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "body", Reason: err.Error()}))
		return
	}

	in := UpdateRuleInput{Version: req.Version}
	if req.Name != nil {
		in.Name = req.Name
	}
	if req.Description != nil {
		in.Description = req.Description
	}
	if req.DSL != nil {
		in.DSL = req.DSL
	}
	if req.Status != nil {
		s := RuleStatus(*req.Status)
		in.Status = &s
	}

	rule, err := h.deps.Svc.Update(c.Request.Context(), wsID, ruleID, in)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, rule)
}

// Delete godoc
//
//	@Summary		删除规则
//	@Description	删除一条自动化规则
//	@Tags			automation
//	@Produce		json
//	@Param			rule_id	path	int	true	"规则ID"
//	@Success		204
//	@Security		Bearer
//	@Router			/automation/{rule_id} [delete]
func (h *Handler) Delete(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	ruleID, err := parseInt64Param(c, "rule_id")
	if err != nil {
		middleware.AbortWithError(c, errs.ErrValidation)
		return
	}

	if err := h.deps.Svc.Delete(c.Request.Context(), wsID, ruleID); err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// Toggle godoc
//
//	@Summary		启用/禁用规则
//	@Description	快速切换规则激活状态
//	@Tags			automation
//	@Accept			json
//	@Produce		json
//	@Param			rule_id	path		int					true	"规则ID"
//	@Param			body	body		toggleRuleRequest	true	"目标状态"
//	@Success		200		{object}	Rule
//	@Security		Bearer
//	@Router			/automation/{rule_id}/toggle [post]
func (h *Handler) Toggle(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	ruleID, err := parseInt64Param(c, "rule_id")
	if err != nil {
		middleware.AbortWithError(c, errs.ErrValidation)
		return
	}

	var req toggleRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithError(c, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "body", Reason: err.Error()}))
		return
	}

	status := RuleStatusDraft
	if req.Enable {
		status = RuleStatusActive
	}
	rule, err := h.deps.Svc.Update(c.Request.Context(), wsID, ruleID, UpdateRuleInput{
		Status:  &status,
		Version: req.Version,
	})
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, rule)
}

// DryRun godoc
//
//	@Summary		干跑测试规则
//	@Description	传入 DSL 模拟执行，返回"将执行"的动作列表（不实际操作）
//	@Tags			automation
//	@Accept			json
//	@Produce		json
//	@Param			body	body		dryRunRequest	true	"模拟运行参数"
//	@Success		200		{object}	dryRunResponse
//	@Security		Bearer
//	@Router			/automation/dry-run [post]
func (h *Handler) DryRun(c *gin.Context) {
	var req dryRunRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithError(c, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "body", Reason: err.Error()}))
		return
	}

	// 仅做 DSL 校验（完整版需要运行时上下文集成）
	result := ValidateDSL(req.DSL)
	c.JSON(http.StatusOK, dryRunResponse{
		Valid:       result.Valid,
		Errors:      result.Errors,
		Warnings:    result.Warnings,
		Actions:     len(req.DSL.Actions),
		TriggerType: req.DSL.Trigger.Type,
	})
}

// ListTemplates godoc
//
//	@Summary		列出内置模板
//	@Description	返回 7 条预置模板（规则编辑器"添加模板"面板数据源）
//	@Tags			automation
//	@Produce		json
//	@Success		200	{array}	Template
//	@Security		Bearer
//	@Router			/automation/templates [get]
func (h *Handler) ListTemplates(c *gin.Context) {
	templates, err := h.deps.Svc.ListTemplates(c.Request.Context())
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, templates)
}

// ListExecutions godoc
//
//	@Summary		列出执行历史
//	@Description	查询最近 N 条规则执行审计记录
//	@Tags			automation
//	@Produce		json
//	@Param			limit	query		int	false	"返回数量（1-100）"	default(20)
//	@Success		200		{array}	RuleExecution
//	@Security		Bearer
//	@Router			/automation/executions [get]
func (h *Handler) ListExecutions(c *gin.Context) {
	projectID := c.GetInt64(middleware.CtxProjectID)
	limit := 20
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = min(parsed, 100)
		}
	}

	execs, err := h.deps.Svc.ListRecentExecutions(c.Request.Context(), projectID, limit)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, execs)
}

// --- Request/Response Types ---

type ruleListResponse struct {
	Total int    `json:"total"`
	Items []Rule `json:"items"`
}

type createRuleRequest struct {
	Name        string     `json:"name" binding:"required,max=128"`
	Description string     `json:"description" binding:"max=500"`
	DSL         RuleDSL    `json:"dsl" binding:"required"`
	Status      RuleStatus `json:"status"`
}

type updateRuleRequest struct {
	Name        *string  `json:"name" binding:"omitempty,max=128"`
	Description *string  `json:"description" binding:"omitempty,max=500"`
	DSL         *RuleDSL `json:"dsl"`
	Status      *string  `json:"status" binding:"omitempty,oneof=draft active disabled error"`
	Version     int      `json:"version"`
}

type toggleRuleRequest struct {
	Enable  bool `json:"enable"`
	Version int  `json:"version"`
}

type createFromTemplateRequest struct {
	TemplateSlug string `json:"template_slug" binding:"required"`
	Name         string `json:"name"`
	Description  string `json:"description"`
}

type dryRunRequest struct {
	DSL RuleDSL `json:"dsl" binding:"required"`
}

type dryRunResponse struct {
	Valid       bool     `json:"valid"`
	Errors      []string `json:"errors,omitempty"`
	Warnings    []string `json:"warnings,omitempty"`
	Actions     int      `json:"actions_count"`
	TriggerType string   `json:"trigger_type"`
}

// --- Helpers ---

func writeError(c *gin.Context, err error) {
	var appErr *errs.AppError
	if errs.As(err, &appErr) {
		middleware.AbortWithError(c, appErr)
		return
	}
	middleware.AbortWithError(c, errs.ErrInternal.Wrap(err))
}

func parseInt64Param(c *gin.Context, name string) (int64, error) {
	v, err := strconv.ParseInt(c.Param(name), 10, 64)
	if err != nil {
		return 0, err
	}
	return v, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
