// Package customer — 客户管理 HTTP handlers。
//
// 路由挂载于 /api/v1/workspaces/:workspace_id：
//   - /customers                       客户 CRUD + 列表筛选
//   - /customers/:id/stats             客户请求统计
//   - /customer-requests               客户请求 CRUD + 关联工作项
package customer

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
	g := r.Group("/customers")
	{
		g.GET("", h.listCustomers)
		g.GET("/search", h.searchCustomers)
		g.GET("/:id", h.getCustomer)
		g.GET("/:id/stats", h.getCustomerStats)
	}

	r.Group("/customer-requests").GET("", h.listCustomerRequests)
	r.Group("/customer-requests").GET("/:id", h.getCustomerRequest)
}

// RegisterWrite 注册写路由。
func (h *Handler) RegisterWrite(r *gin.RouterGroup) {
	g := r.Group("/customers")
	{
		g.POST("", h.createCustomer)
		g.PATCH("/:id", h.updateCustomer)
		g.DELETE("/:id", h.deleteCustomer)
	}

	cr := r.Group("/customer-requests")
	{
		cr.POST("", h.createCustomerRequest)
		cr.PATCH("/:id", h.updateCustomerRequest)
		cr.DELETE("/:id", h.deleteCustomerRequest)
		cr.POST("/:id/link", h.linkRequestToWorkItem)
	}
}

// Register 注册全部路由。
func (h *Handler) Register(r *gin.RouterGroup) {
	h.RegisterRead(r)
	h.RegisterWrite(r)
}

// ---- 客户 ----

// listCustomers GET /customers?stage=&tier=&industry=&search=&limit=&offset=
//
//	@Summary		客户列表
//	@Description	按 stage / tier / industry / 关键字分页列出客户
//	@Tags			customer
//	@Produce		json
//	@Success		200	{object}	map[string]any
//	@Router			/customers [get]
func (h *Handler) listCustomers(c *gin.Context) {
	f := ListCustomersFilter{
		WorkspaceID: wsID(c),
		Stage:       c.Query("stage"),
		Tier:        c.Query("tier"),
		Industry:    c.Query("industry"),
		Search:      c.Query("search"),
		Limit:       intQuery(c, "limit", 50),
		Offset:      intQuery(c, "offset", 0),
	}
	items, total, err := h.d.Svc.ListCustomers(c.Request.Context(), f)
	if err != nil {
		writeCustomerErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"results": items, "total": total})
}

// searchCustomers GET /customers/search?q=&limit=
//
//	@Summary		搜索客户
//	@Description	按名称 / 邮箱 / 行业模糊搜索客户
//	@Tags			customer
//	@Produce		json
//	@Param			q		query	string	false	"搜索关键字"
//	@Param			limit	query	int		false	"最大条数"	default(20)
//	@Success		200		{object}	map[string]any
//	@Router			/customers/search [get]
func (h *Handler) searchCustomers(c *gin.Context) {
	q := c.Query("q")
	limit := intQuery(c, "limit", 20)
	items, err := h.d.Svc.SearchCustomers(c.Request.Context(), wsID(c), q, limit)
	if err != nil {
		writeCustomerErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"results": items, "total": len(items)})
}

// getCustomer GET /customers/:id
//
//	@Summary		客户详情
//	@Tags		customer
//	@Produce	json
//	@Param		id	path	int	true	"客户 ID"
//	@Success	200	{object}	Customer
//	@Router		/customers/{id} [get]
func (h *Handler) getCustomer(c *gin.Context) {
	id, err := idParam(c, "id")
	if err != nil {
		writeCustomerErr(c, err)
		return
	}
	customer, err := h.d.Svc.GetCustomer(c.Request.Context(), wsID(c), id)
	if err != nil {
		writeCustomerErr(c, err)
		return
	}
	c.JSON(http.StatusOK, customer)
}

// createCustomer POST /customers
type createCustomerRequest struct {
	Name             string `json:"name" binding:"required"`
	Email            string `json:"email"`
	Website          string `json:"website"`
	Industry         string `json:"industry"`
	Stage            string `json:"stage"`
	Tier             string `json:"tier"`
	Region           string `json:"region"`
	Employees        int    `json:"employees"`
	Revenue          string `json:"revenue"`
	AccountManagerID int64  `json:"account_manager_id"`
	Notes            string `json:"notes"`
}

// createCustomer 创建客户。
//
//	@Summary		创建客户
//	@Tags		customer
//	@Accept		json
//	@Produce	json
//	@Success	201	{object}	Customer
//	@Router		/customers [post]
func (h *Handler) createCustomer(c *gin.Context) {
	var req createCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeCustomerErr(c, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "body", Reason: "请求体无效: " + err.Error()}))
		return
	}
	customer, err := h.d.Svc.CreateCustomer(c.Request.Context(), CreateCustomerInput{
		WorkspaceID:      wsID(c),
		Name:             req.Name,
		Email:            req.Email,
		Website:          req.Website,
		Industry:         req.Industry,
		Stage:            req.Stage,
		Tier:             req.Tier,
		Region:           req.Region,
		Employees:        req.Employees,
		Revenue:          req.Revenue,
		AccountManagerID: req.AccountManagerID,
		Notes:            req.Notes,
		CreatedBy:        actorID(c),
	})
	if err != nil {
		writeCustomerErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, customer)
}

// updateCustomer PATCH /customers/:id
type updateCustomerRequest struct {
	Name             *string `json:"name"`
	Email            *string `json:"email"`
	Website          *string `json:"website"`
	Industry         *string `json:"industry"`
	Stage            *string `json:"stage"`
	Tier             *string `json:"tier"`
	Region           *string `json:"region"`
	Employees        *int    `json:"employees"`
	Revenue          *string `json:"revenue"`
	AccountManagerID *int64  `json:"account_manager_id"`
	Notes            *string `json:"notes"`
}

// updateCustomer 更新客户。
//
//	@Summary		更新客户
//	@Tags		customer
//	@Accept		json
//	@Produce	json
//	@Param		id	path	int	true	"客户 ID"
//	@Success	200	{object}	Customer
//	@Router		/customers/{id} [patch]
func (h *Handler) updateCustomer(c *gin.Context) {
	id, err := idParam(c, "id")
	if err != nil {
		writeCustomerErr(c, err)
		return
	}
	var req updateCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeCustomerErr(c, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "body", Reason: "请求体无效: " + err.Error()}))
		return
	}
	customer, err := h.d.Svc.UpdateCustomer(c.Request.Context(), wsID(c), id, UpdateCustomerInput{
		Name:             req.Name,
		Email:            req.Email,
		Website:          req.Website,
		Industry:         req.Industry,
		Stage:            req.Stage,
		Tier:             req.Tier,
		Region:           req.Region,
		Employees:        req.Employees,
		Revenue:          req.Revenue,
		AccountManagerID: req.AccountManagerID,
		Notes:            req.Notes,
	})
	if err != nil {
		writeCustomerErr(c, err)
		return
	}
	c.JSON(http.StatusOK, customer)
}

// deleteCustomer DELETE /customers/:id
//
//	@Summary		删除客户
//	@Tags		customer
//	@Param		id	path	int	true	"客户 ID"
//	@Success	204
//	@Router		/customers/{id} [delete]
func (h *Handler) deleteCustomer(c *gin.Context) {
	id, err := idParam(c, "id")
	if err != nil {
		writeCustomerErr(c, err)
		return
	}
	if err := h.d.Svc.DeleteCustomer(c.Request.Context(), wsID(c), id); err != nil {
		writeCustomerErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// getCustomerStats GET /customers/:id/stats
//
//	@Summary		客户请求统计
//	@Tags		customer
//	@Produce	json
//	@Param		id	path	int	true	"客户 ID"
//	@Success	200	{object}	CustomerStats
//	@Router		/customers/{id}/stats [get]
func (h *Handler) getCustomerStats(c *gin.Context) {
	id, err := idParam(c, "id")
	if err != nil {
		writeCustomerErr(c, err)
		return
	}
	stats, err := h.d.Svc.GetCustomerStats(c.Request.Context(), wsID(c), id)
	if err != nil {
		writeCustomerErr(c, err)
		return
	}
	c.JSON(http.StatusOK, stats)
}

// ---- 客户请求 ----

// listCustomerRequests GET /customer-requests?status=&priority=&customer_id=&limit=&offset=
//
//	@Summary		客户请求列表
//	@Tags		customer
//	@Produce	json
//	@Success	200	{object}	map[string]any
//	@Router		/customer-requests [get]
func (h *Handler) listCustomerRequests(c *gin.Context) {
	f := ListCustomerRequestsFilter{
		WorkspaceID: wsID(c),
		Status:      c.Query("status"),
		Priority:    c.Query("priority"),
		Limit:       intQuery(c, "limit", 50),
		Offset:      intQuery(c, "offset", 0),
	}
	if v, err := strconv.ParseInt(c.Query("customer_id"), 10, 64); err == nil && v > 0 {
		f.CustomerID = &v
	}
	items, total, err := h.d.Svc.ListCustomerRequests(c.Request.Context(), f)
	if err != nil {
		writeCustomerErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"results": items, "total": total})
}

// getCustomerRequest GET /customer-requests/:id
//
//	@Summary		客户请求详情
//	@Tags		customer
//	@Produce	json
//	@Param		id	path	int	true	"请求 ID"
//	@Success	200	{object}	CustomerRequest
//	@Router		/customer-requests/{id} [get]
func (h *Handler) getCustomerRequest(c *gin.Context) {
	id, err := idParam(c, "id")
	if err != nil {
		writeCustomerErr(c, err)
		return
	}
	item, err := h.d.Svc.GetCustomerRequest(c.Request.Context(), wsID(c), id)
	if err != nil {
		writeCustomerErr(c, err)
		return
	}
	c.JSON(http.StatusOK, item)
}

// createCustomerRequest POST /customer-requests
type createCustomerRequestRequest struct {
	CustomerID   int64  `json:"customer_id" binding:"required"`
	Title        string `json:"title" binding:"required"`
	Description  string `json:"description"`
	Status       string `json:"status"`
	Priority     string `json:"priority"`
	Source       string `json:"source"`
	WorkItemID   *int64 `json:"work_item_id"`
	WorkItemType string `json:"work_item_type"`
	AssignedTo   int64  `json:"assigned_to"`
}

// createCustomerRequest 创建客户请求。
//
//	@Summary		创建客户请求
//	@Tags		customer
//	@Accept		json
//	@Produce	json
//	@Success	201	{object}	CustomerRequest
//	@Router		/customer-requests [post]
func (h *Handler) createCustomerRequest(c *gin.Context) {
	var req createCustomerRequestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeCustomerErr(c, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "body", Reason: "请求体无效: " + err.Error()}))
		return
	}
	item, err := h.d.Svc.CreateCustomerRequest(c.Request.Context(), CreateCustomerRequestInput{
		WorkspaceID:  wsID(c),
		CustomerID:   req.CustomerID,
		Title:        req.Title,
		Description:  req.Description,
		Status:       req.Status,
		Priority:     req.Priority,
		Source:       req.Source,
		WorkItemID:   req.WorkItemID,
		WorkItemType: req.WorkItemType,
		AssignedTo:   req.AssignedTo,
	})
	if err != nil {
		writeCustomerErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, item)
}

// updateCustomerRequest PATCH /customer-requests/:id
type updateCustomerRequestRequest struct {
	Title        *string `json:"title"`
	Description  *string `json:"description"`
	Status       *string `json:"status"`
	Priority     *string `json:"priority"`
	Source       *string `json:"source"`
	CustomerID   *int64  `json:"customer_id"`
	WorkItemID   *int64  `json:"work_item_id"`
	WorkItemType *string `json:"work_item_type"`
	AssignedTo   *int64  `json:"assigned_to"`
}

// updateCustomerRequest 更新客户请求。
//
//	@Summary		更新客户请求
//	@Tags		customer
//	@Accept		json
//	@Produce	json
//	@Param		id	path	int	true	"请求 ID"
//	@Success	200	{object}	CustomerRequest
//	@Router		/customer-requests/{id} [patch]
func (h *Handler) updateCustomerRequest(c *gin.Context) {
	id, err := idParam(c, "id")
	if err != nil {
		writeCustomerErr(c, err)
		return
	}
	var req updateCustomerRequestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeCustomerErr(c, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "body", Reason: "请求体无效: " + err.Error()}))
		return
	}
	item, err := h.d.Svc.UpdateCustomerRequest(c.Request.Context(), wsID(c), id, UpdateCustomerRequestInput{
		Title:        req.Title,
		Description:  req.Description,
		Status:       req.Status,
		Priority:     req.Priority,
		Source:       req.Source,
		CustomerID:   req.CustomerID,
		WorkItemID:   req.WorkItemID,
		WorkItemType: req.WorkItemType,
		AssignedTo:   req.AssignedTo,
	})
	if err != nil {
		writeCustomerErr(c, err)
		return
	}
	c.JSON(http.StatusOK, item)
}

// deleteCustomerRequest DELETE /customer-requests/:id
//
//	@Summary		删除客户请求
//	@Tags		customer
//	@Param		id	path	int	true	"请求 ID"
//	@Success	204
//	@Router		/customer-requests/{id} [delete]
func (h *Handler) deleteCustomerRequest(c *gin.Context) {
	id, err := idParam(c, "id")
	if err != nil {
		writeCustomerErr(c, err)
		return
	}
	if err := h.d.Svc.DeleteCustomerRequest(c.Request.Context(), wsID(c), id); err != nil {
		writeCustomerErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// linkRequestToWorkItem POST /customer-requests/:id/link
type linkRequestRequest struct {
	WorkItemID   int64  `json:"work_item_id" binding:"required"`
	WorkItemType string `json:"work_item_type"`
}

// linkRequestToWorkItem 关联客户请求到工作项。
//
//	@Summary		关联工作项
//	@Tags		customer
//	@Accept		json
//	@Produce	json
//	@Param		id	path	int	true	"请求 ID"
//	@Success	200	{object}	CustomerRequest
//	@Router		/customer-requests/{id}/link [post]
func (h *Handler) linkRequestToWorkItem(c *gin.Context) {
	id, err := idParam(c, "id")
	if err != nil {
		writeCustomerErr(c, err)
		return
	}
	var req linkRequestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeCustomerErr(c, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "body", Reason: "请求体无效: " + err.Error()}))
		return
	}
	item, err := h.d.Svc.LinkRequestToWorkItem(c.Request.Context(), LinkRequestInput{
		WorkspaceID:  wsID(c),
		RequestID:    id,
		WorkItemID:   req.WorkItemID,
		WorkItemType: req.WorkItemType,
	})
	if err != nil {
		writeCustomerErr(c, err)
		return
	}
	c.JSON(http.StatusOK, item)
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

func intQuery(c *gin.Context, key string, def int) int {
	if v := c.Query(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
	}
	return def
}

func writeCustomerErr(c *gin.Context, err error) {
	var appErr *errs.AppError
	if errs.As(err, &appErr) {
		middleware.AbortWithError(c, appErr)
		return
	}
	middleware.AbortWithError(c, errs.ErrInternal)
}
