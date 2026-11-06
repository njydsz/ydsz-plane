// Package recurring — 重复工作项 HTTP handlers。
//
// 路由挂载：
//   /api/v1/workspaces/:workspace_id/projects/:project_id/recurring-rules
package recurring

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

// RegisterRead 注册只读路由。
func (h *Handler) RegisterRead(r *gin.RouterGroup) {
	g := r.Group("/recurring-rules")
	g.GET("", h.listRules)
	g.GET("/:rule_id", h.getRule)
}

// RegisterWrite 注册写入路由。
func (h *Handler) RegisterWrite(r *gin.RouterGroup) {
	g := r.Group("/recurring-rules")
	g.POST("", h.createRule)
	g.PATCH("/:rule_id", h.updateRule)
	g.DELETE("/:rule_id", h.deleteRule)
	g.POST("/:rule_id/enable", h.enableRule)
	g.POST("/:rule_id/disable", h.disableRule)
	g.POST("/:rule_id/run-now", h.runNow)
}

// Register 注册全部路由。
func (h *Handler) Register(r *gin.RouterGroup) {
	h.RegisterRead(r)
	h.RegisterWrite(r)
}

// --- helpers ---

func writeErr(c *gin.Context, err error) {
	var appErr *errs.AppError
	if errs.As(err, &appErr) {
		middleware.AbortWithError(c, appErr)
		return
	}
	switch err {
	case ErrRuleNotFound:
		middleware.AbortWithError(c, errs.ErrNotFound)
	default:
		middleware.AbortWithError(c, errs.ErrInternal)
	}
}

func intParam(c *gin.Context, key string) (int64, error) {
	id, err := strconv.ParseInt(c.Param(key), 10, 64)
	if err != nil || id <= 0 {
		return 0, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: key, Reason: "无效的 ID"})
	}
	return id, nil
}

func intQuery(c *gin.Context, key string, def int) int {
	if v := c.Query(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
	}
	return def
}

func fieldDetail(err error) errs.FieldDetail {
	return errs.FieldDetail{Field: "body", Reason: err.Error()}
}

// --- handlers ---

// listRules 列出项目下的重复规则。
func (h *Handler) listRules(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	projectID := c.GetInt64(middleware.CtxProjectID)

	opts := ListRulesOptions{
		WorkspaceID: wsID,
		ProjectID:   projectID,
		Limit:       intQuery(c, "limit", 100),
		Offset:      intQuery(c, "offset", 0),
	}
	if v := c.Query("enabled"); v != "" {
		b := v == "true" || v == "1"
		opts.Enabled = &b
	}

	rules, err := h.d.Svc.ListRules(c.Request.Context(), opts)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"results": rules})
}

// getRule 查询单条规则。
func (h *Handler) getRule(c *gin.Context) {
	ruleID, err := intParam(c, "rule_id")
	if err != nil {
		writeErr(c, err)
		return
	}
	rule, err := h.d.Svc.GetRule(c.Request.Context(), ruleID)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, rule)
}

// createRule 创建重复规则。
func (h *Handler) createRule(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	projectID := c.GetInt64(middleware.CtxProjectID)
	userID := c.GetInt64(middleware.CtxUserID)

	var req struct {
		WorkItemType        string  `json:"work_item_type"`
		TitleTemplate       string  `json:"title_template"`
		DescriptionTemplate string  `json:"description_template"`
		CronExpr            string  `json:"cron_expr"`
		AssigneeIDs         []int64 `json:"assignee_ids"`
		LabelIDs            []int64 `json:"label_ids"`
		StateID             int64   `json:"state_id"`
		Enabled             bool    `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		writeErr(c, errs.ErrValidation.WithDetails(fieldDetail(err)))
		return
	}

	rule, err := h.d.Svc.CreateRule(c.Request.Context(), CreateRuleInput{
		WorkspaceID:         wsID,
		ProjectID:           projectID,
		WorkItemType:        req.WorkItemType,
		TitleTemplate:       req.TitleTemplate,
		DescriptionTemplate: req.DescriptionTemplate,
		CronExpr:            req.CronExpr,
		AssigneeIDs:         req.AssigneeIDs,
		LabelIDs:            req.LabelIDs,
		StateID:             req.StateID,
		Enabled:             req.Enabled,
		CreatedBy:           userID,
	})
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, rule)
}

// updateRule 更新重复规则。
func (h *Handler) updateRule(c *gin.Context) {
	ruleID, err := intParam(c, "rule_id")
	if err != nil {
		writeErr(c, err)
		return
	}

	var req struct {
		TitleTemplate       *string  `json:"title_template"`
		DescriptionTemplate *string  `json:"description_template"`
		CronExpr            *string  `json:"cron_expr"`
		AssigneeIDs         []int64  `json:"assignee_ids"`
		LabelIDs            []int64  `json:"label_ids"`
		StateID             *int64   `json:"state_id"`
		Enabled             *bool    `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		writeErr(c, errs.ErrValidation.WithDetails(fieldDetail(err)))
		return
	}

	rule, err := h.d.Svc.UpdateRule(c.Request.Context(), ruleID, UpdateRuleInput{
		TitleTemplate:       req.TitleTemplate,
		DescriptionTemplate: req.DescriptionTemplate,
		CronExpr:            req.CronExpr,
		AssigneeIDs:         req.AssigneeIDs,
		LabelIDs:            req.LabelIDs,
		StateID:             req.StateID,
		Enabled:             req.Enabled,
	})
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, rule)
}

// deleteRule 删除重复规则。
func (h *Handler) deleteRule(c *gin.Context) {
	ruleID, err := intParam(c, "rule_id")
	if err != nil {
		writeErr(c, err)
		return
	}
	if err := h.d.Svc.DeleteRule(c.Request.Context(), ruleID); err != nil {
		writeErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// enableRule 启用重复规则。
func (h *Handler) enableRule(c *gin.Context) {
	ruleID, err := intParam(c, "rule_id")
	if err != nil {
		writeErr(c, err)
		return
	}
	rule, err := h.d.Svc.EnableRule(c.Request.Context(), ruleID)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, rule)
}

// disableRule 禁用重复规则。
func (h *Handler) disableRule(c *gin.Context) {
	ruleID, err := intParam(c, "rule_id")
	if err != nil {
		writeErr(c, err)
		return
	}
	rule, err := h.d.Svc.DisableRule(c.Request.Context(), ruleID)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, rule)
}

// runNow 立即执行一次规则。
func (h *Handler) runNow(c *gin.Context) {
	ruleID, err := intParam(c, "rule_id")
	if err != nil {
		writeErr(c, err)
		return
	}
	if err := h.d.Svc.RunNow(c.Request.Context(), ruleID); err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
