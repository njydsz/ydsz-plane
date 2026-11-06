// Package workflow — 工作流与审批 HTTP handlers。
//
// 路由挂载：
//   - /api/v1/workspaces/:workspace_id/projects/:project_id/workflows/...（项目级配置）
//   - /api/v1/workspaces/:workspace_id/approvals/...（工作空间级审批待办）
package workflow

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/njydsz/ydsz-plane/internal/interfaces/middleware"
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

// --- 项目级工作流配置 ---

// RegisterProject 注册项目级工作流配置路由。
func (h *Handler) RegisterProject(r *gin.RouterGroup) {
	g := r.Group("/workflows")
	g.GET("", h.ListWorkflows)
	g.POST("", h.CreateWorkflow)
	g.GET("/:workflow_id", h.GetWorkflowDetail)
	g.PATCH("/:workflow_id", h.UpdateWorkflow)
	g.DELETE("/:workflow_id", h.DeleteWorkflow)
	g.POST("/:workflow_id/transitions", h.AddTransition)
	g.DELETE("/:workflow_id/transitions/:transition_id", h.DeleteTransition)
	g.POST("/:workflow_id/transitions/:transition_id/validate", h.ValidateTransition)
}

// RegisterWorkspace 注册工作空间级审批路由。
func (h *Handler) RegisterWorkspace(r *gin.RouterGroup) {
	g := r.Group("/approvals")
	g.GET("", h.ListPendingApprovals)
	g.POST("", h.SubmitApproval)
	g.POST("/:approval_id/approve", h.ApproveApproval)
	g.POST("/:approval_id/reject", h.RejectApproval)
}

// --- helpers ---

func (h *Handler) fail(c *gin.Context, err error) {
	switch err {
	case ErrWorkflowNotFound, ErrTransitionNotFound, ErrApprovalNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case ErrTransitionNotAllowed, ErrApprovalRequired, ErrApprovalAlreadyResolved, ErrNotApprover:
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	default:
		// errs.Code 类型携带 HTTP 状态码；其余按 500 处理。
		if ec, ok := err.(interface {
			HTTPStatus() int
		}); ok {
			c.JSON(ec.HTTPStatus(), gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
	}
}

func intQuery(c *gin.Context, key string, def int) int {
	if v := c.Query(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
	}
	return def
}

func paramInt(c *gin.Context, key string) int64 {
	n, _ := strconv.ParseInt(c.Param(key), 10, 64)
	return n
}

// ListWorkflows godoc
//
//	@Summary		工作流列表
//	@Tags			workflow
//	@Produce		json
//	@Success		200	{object}	map[string]any
//	@Router			/workflows [get]
func (h *Handler) ListWorkflows(c *gin.Context) {
	projectID := c.GetInt64(middleware.CtxProjectID)
	items, err := h.d.Svc.List(c.Request.Context(), ListWorkflowsOptions{
		ProjectID: projectID,
		Limit:     intQuery(c, "limit", 100),
		Offset:    intQuery(c, "offset", 0),
	})
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

// CreateWorkflow godoc
//
//	@Summary		创建工作流
//	@Tags			workflow
//	@Accept			json
//	@Produce		json
//	@Success		201	{object}	Workflow
//	@Router			/workflows [post]
func (h *Handler) CreateWorkflow(c *gin.Context) {
	var body struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		IsDefault   bool   `json:"is_default"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	wf, err := h.d.Svc.Create(c.Request.Context(), CreateWorkflowInput{
		ProjectID:   c.GetInt64(middleware.CtxProjectID),
		Name:        body.Name,
		Description: body.Description,
		IsDefault:   body.IsDefault,
		CreatedBy:   c.GetInt64(middleware.CtxUserID),
	})
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusCreated, wf)
}

// GetWorkflowDetail godoc
//
//	@Summary		工作流详情（含流转规则）
//	@Tags			workflow
//	@Produce		json
//	@Success		200	{object}	WorkflowDetail
//	@Router			/workflows/{workflow_id} [get]
func (h *Handler) GetWorkflowDetail(c *gin.Context) {
	detail, err := h.d.Svc.GetDetail(c.Request.Context(), paramInt(c, "workflow_id"))
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, detail)
}

// UpdateWorkflow godoc
//
//	@Summary		更新工作流
//	@Tags			workflow
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	Workflow
//	@Router			/workflows/{workflow_id} [patch]
func (h *Handler) UpdateWorkflow(c *gin.Context) {
	var body struct {
		Name        *string `json:"name"`
		Description *string `json:"description"`
		IsDefault   *bool   `json:"is_default"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	wf, err := h.d.Svc.Update(c.Request.Context(), paramInt(c, "workflow_id"), UpdateWorkflowInput{
		Name:        body.Name,
		Description: body.Description,
		IsDefault:   body.IsDefault,
	})
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, wf)
}

// DeleteWorkflow godoc
//
//	@Summary		删除工作流
//	@Tags			workflow
//	@Produce		json
//	@Success		204
//	@Router			/workflows/{workflow_id} [delete]
func (h *Handler) DeleteWorkflow(c *gin.Context) {
	if err := h.d.Svc.Delete(c.Request.Context(), paramInt(c, "workflow_id")); err != nil {
		h.fail(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// AddTransition godoc
//
//	@Summary		新增流转规则
//	@Tags			workflow
//	@Accept			json
//	@Produce		json
//	@Success		201	{object}	WorkflowTransition
//	@Router			/workflows/{workflow_id}/transitions [post]
func (h *Handler) AddTransition(c *gin.Context) {
	var body struct {
		FromStateID      int64   `json:"from_state_id"`
		ToStateID        int64   `json:"to_state_id"`
		Name             string  `json:"name"`
		RequiresApproval bool    `json:"requires_approval"`
		ApproverRole     string  `json:"approver_role"`
		ApproverIDs      []int64 `json:"approver_ids"`
		Conditions       string  `json:"conditions"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	tr, err := h.d.Svc.AddTransition(c.Request.Context(), CreateTransitionInput{
		WorkflowID:       paramInt(c, "workflow_id"),
		FromStateID:      body.FromStateID,
		ToStateID:        body.ToStateID,
		Name:             body.Name,
		RequiresApproval: body.RequiresApproval,
		ApproverRole:     body.ApproverRole,
		ApproverIDs:      body.ApproverIDs,
		Conditions:       body.Conditions,
	})
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusCreated, tr)
}

// DeleteTransition godoc
//
//	@Summary		删除流转规则
//	@Tags			workflow
//	@Produce		json
//	@Success		204
//	@Router			/workflows/{workflow_id}/transitions/{transition_id} [delete]
func (h *Handler) DeleteTransition(c *gin.Context) {
	if err := h.d.Svc.DeleteTransition(c.Request.Context(), paramInt(c, "transition_id")); err != nil {
		h.fail(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ValidateTransition godoc
//
//	@Summary		校验流转是否允许
//	@Tags			workflow
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	ValidateResult
//	@Router			/workflows/{workflow_id}/transitions/{transition_id}/validate [post]
func (h *Handler) ValidateTransition(c *gin.Context) {
	var body struct {
		FromStateID int64 `json:"from_state_id"`
		ToStateID   int64 `json:"to_state_id"`
		WorkItemID  int64 `json:"work_item_id"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	res, err := h.d.Svc.ValidateTransition(c.Request.Context(), ValidateTransitionInput{
		WorkflowID:  paramInt(c, "workflow_id"),
		FromStateID: body.FromStateID,
		ToStateID:   body.ToStateID,
		WorkItemID:  body.WorkItemID,
	})
	if err != nil {
		h.fail(c, err)
		return
	}
	// 适配前端契约：{ valid, errors }
	out := gin.H{"valid": res.Allowed}
	if !res.Allowed {
		out["errors"] = []string{res.Reason}
	}
	c.JSON(http.StatusOK, out)
}

// --- 审批 ---

// SubmitApproval godoc
//
//	@Summary		提交审批申请
//	@Tags			workflow
//	@Accept			json
//	@Produce		json
//	@Success		201	{object}	WorkflowApproval
//	@Router			/approvals [post]
func (h *Handler) SubmitApproval(c *gin.Context) {
	var body struct {
		TransitionID int64  `json:"transition_id"`
		WorkItemID   int64  `json:"work_item_id"`
		WorkItemType string `json:"work_item_type"`
		ApproverID   int64  `json:"approver_id"`
		Comment      string `json:"comment"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ap, err := h.d.Svc.SubmitApproval(c.Request.Context(), SubmitApprovalInput{
		TransitionID: body.TransitionID,
		WorkItemID:   body.WorkItemID,
		WorkItemType: body.WorkItemType,
		RequesterID:  c.GetInt64(middleware.CtxUserID),
		ApproverID:   body.ApproverID,
		Comment:      body.Comment,
	})
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusCreated, ap)
}

// ListPendingApprovals godoc
//
//	@Summary		待我处理的审批列表
//	@Tags			workflow
//	@Produce		json
//	@Success		200	{object}	map[string]any
//	@Router			/approvals [get]
func (h *Handler) ListPendingApprovals(c *gin.Context) {
	limit := intQuery(c, "limit", 50)
	page := intQuery(c, "page", 1)
	if page < 1 {
		page = 1
	}
	items, err := h.d.Svc.ListPendingApprovals(c.Request.Context(), ListPendingApprovalsOptions{
		WorkspaceID: c.GetInt64(middleware.CtxWorkspaceID),
		UserID:      c.GetInt64(middleware.CtxUserID),
		Limit:       limit,
		Offset:      (page - 1) * limit,
	})
	if err != nil {
		h.fail(c, err)
		return
	}
	// 适配前端契约：{ items, total }
	total := len(items)
	c.JSON(http.StatusOK, gin.H{"items": items, "total": total})
}

// ApproveApproval godoc
//
//	@Summary		批准审批
//	@Tags			workflow
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	WorkflowApproval
//	@Router			/approvals/{approval_id}/approve [post]
func (h *Handler) ApproveApproval(c *gin.Context) {
	var body struct {
		Comment string `json:"comment"`
	}
	_ = c.ShouldBindJSON(&body)
	ap, err := h.d.Svc.ResolveApproval(c.Request.Context(), ResolveApprovalInput{
		ApprovalID: paramInt(c, "approval_id"),
		ApproverID: c.GetInt64(middleware.CtxUserID),
		Comment:    body.Comment,
	}, true)
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, ap)
}

// RejectApproval godoc
//
//	@Summary		驳回审批
//	@Tags			workflow
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	WorkflowApproval
//	@Router			/approvals/{approval_id}/reject [post]
func (h *Handler) RejectApproval(c *gin.Context) {
	var body struct {
		Comment string `json:"comment"`
	}
	_ = c.ShouldBindJSON(&body)
	ap, err := h.d.Svc.ResolveApproval(c.Request.Context(), ResolveApprovalInput{
		ApprovalID: paramInt(c, "approval_id"),
		ApproverID: c.GetInt64(middleware.CtxUserID),
		Comment:    body.Comment,
	}, false)
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, ap)
}
