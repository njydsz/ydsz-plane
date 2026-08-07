// Package issue — Issue HTTP handlers（REST API）。
package issue

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/njydsz/ydsz-plane/internal/application/auth"
	"github.com/njydsz/ydsz-plane/internal/interfaces/middleware"
	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// HandlerDeps handler 依赖。
type HandlerDeps struct {
	IssueSvc       *Service
	StateSvc       *StateService
	ActivitySvc    *ActivityService
	TimeLogSvc     *TimeLogService
	RelationSvc    *RelationService
	CommentSvc     *CommentService
	ProjectInit    *ProjectInitService
	WorkspaceStore *auth.WorkspaceMembershipStore
}

// IssueHandler Gin handler 集合。
type IssueHandler struct {
	d *HandlerDeps
}

// NewIssueHandler 构造 handler。
func NewIssueHandler(d *HandlerDeps) *IssueHandler {
	return &IssueHandler{d: d}
}

// Register 注册 Issue 路由。
func (h *IssueHandler) Register(r *gin.RouterGroup, wsMiddleware []gin.HandlerFunc, projectMiddleware []gin.HandlerFunc) {
	// 在当前项目子路由组下注册状态/工作项路由
	// wsMiddleware / projectMiddleware 已在父路由组应用（auth / parse params / rbac）
	_ = wsMiddleware
	_ = projectMiddleware

	// 集合
	r.GET("/states", h.listStates)
	r.GET("/issues", h.listIssues)
	r.POST("/issues", h.createIssue)

	// 单资源
	issue := r.Group("/issues/:issue_id")
	{
		issue.GET("", h.getIssue)
		issue.PATCH("", h.updateIssue)
		issue.DELETE("", h.deleteIssue)
		issue.POST("/transition", h.transition)
		issue.GET("/activities", h.listActivities)
		issue.GET("/time-logs", h.listTimeLogs)
		issue.POST("/time-logs", h.createTimeLog)
		issue.GET("/relations", h.listRelations)
		issue.POST("/relations", h.createRelation)
		issue.DELETE("/relations/:relation_id", h.deleteRelation)
		issue.GET("/dependencies", h.listDependencies)
		issue.POST("/dependencies", h.createDependency)
		issue.DELETE("/dependencies/:dep_id", h.deleteDependency)
		// 看板排序（预留，S4 完善）
		issue.PATCH("/reorder", h.reorderIssue)
		// 评论
		issue.GET("/comments", h.listComments)
		issue.POST("/comments", h.createComment)
		issue.PATCH("/comments/:comment_id", h.updateComment)
		issue.DELETE("/comments/:comment_id", h.deleteComment)
	}
}

// --- State handlers ---

func (h *IssueHandler) listStates(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	projectID := c.GetInt64(middleware.CtxProjectID)

	states, err := h.d.StateSvc.GetProjectStates(c.Request.Context(), wsID, projectID)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"results": states})
}

// --- Issue handlers ---

// createIssue 创建工作项（REST handler，Swagger 注解见下）。
//
//	@Summary		创建工作项
//	@Description	在工作空间中创建需求/任务/缺陷
//	@Tags			issue
//	@Accept			json
//	@Produce		json
//	@Param			body	body		createIssueRequest	true	"工作项信息"
//	@Success		201		{object}	Issue
//	@Failure		422		{object}	errs.AppError
//	@Router			/issues [post]
func (h *IssueHandler) createIssue(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	projectID := c.GetInt64(middleware.CtxProjectID)
	userID := c.GetInt64(middleware.CtxUserID)

	var req createIssueRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithError(c, errs.ErrValidation.WithDetails(fieldDetail(err)))
		return
	}

	var severityPtr *int
	if req.Severity != nil {
		severityPtr = req.Severity
	}
	var foundPhasePtr *string
	if req.FoundPhase != nil {
		foundPhasePtr = req.FoundPhase
	}
	var categoryPtr *string
	if req.Category != nil {
		categoryPtr = req.Category
	}
	var sourcePtr *string
	if req.Source != nil {
		sourcePtr = req.Source
	}
	var pointPtr *int
	if req.Point != nil {
		pointPtr = req.Point
	}
	var parentIDPtr *int64
	if req.ParentID != nil {
		parentIDPtr = req.ParentID
	}

	iss, err := h.d.IssueSvc.Create(c.Request.Context(), CreateIssueInput{
		WorkspaceID:      wsID,
		ProjectID:        projectID,
		TypeCode:         IssueTypeCode(req.Type),
		Name:             req.Name,
		DescriptionHTML:  req.DescriptionHTML,
		StateID:          req.StateID,
		Priority:         IssuePriority(req.Priority),
		ParentID:         parentIDPtr,
		Severity:         severityPtr,
		FoundPhase:       foundPhasePtr,
		ReproduceSteps:   req.ReproduceSteps,
		Category:         categoryPtr,
		Source:           sourcePtr,
		Assignees:        req.Assignees,
		Labels:           req.Labels,
		Modules:          req.Modules,
		Point:            pointPtr,
		IsDraft:          req.IsDraft,
		CreatedBy:        userID,
	})
	if err != nil {
		writeErr(c, err)
		return
	}

	c.JSON(http.StatusCreated, iss)
}

// getIssue 获取工作项详情（REST handler，Swagger 注解见下）。
//
//	@Summary		获取工作项详情
//	@Tags			issue
//	@Produce		json
//	@Success		200	{object}	Issue
//	@Failure		404	{object}	errs.AppError
//	@Router			/issues/{issue_id} [get]
func (h *IssueHandler) getIssue(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	issueID := int64Param(c, "issue_id")

	iss, err := h.d.IssueSvc.GetByID(c.Request.Context(), wsID, issueID)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, iss)
}

// updateIssue 更新工作项（REST handler，Swagger 注解见下）。
//
//	@Summary		更新工作项
//	@Tags			issue
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	Issue
//	@Failure		409	{object}	errs.AppError	"版本冲突"
//	@Router			/issues/{issue_id} [patch]
func (h *IssueHandler) updateIssue(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	issueID := int64Param(c, "issue_id")

	var req updateIssueRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithError(c, errs.ErrValidation.WithDetails(fieldDetail(err)))
		return
	}

	in := UpdateIssueInput{Version: req.Version}
	if req.Name != nil {
		in.Name = req.Name
	}
	if req.DescriptionHTML != nil {
		in.DescriptionHTML = req.DescriptionHTML
	}
	if req.StateID != nil {
		in.StateID = req.StateID
	}
	if req.Priority != nil {
		p := IssuePriority(*req.Priority)
		in.Priority = &p
	}
	if req.ParentID != nil {
		in.ParentID = req.ParentID
	}
	if req.Severity != nil {
		in.Severity = req.Severity
	}
	if req.FoundPhase != nil {
		in.FoundPhase = req.FoundPhase
	}
	if req.RootCauseCategory != nil {
		in.RootCauseCategory = req.RootCauseCategory
	}
	if req.Category != nil {
		in.Category = req.Category
	}
	in.Assignees = req.Assignees
	in.Labels = req.Labels
	in.Modules = req.Modules
	if req.Source != nil {
		in.Source = req.Source
	}

	iss, err := h.d.IssueSvc.Update(c.Request.Context(), wsID, issueID, in)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, iss)
}

// deleteIssue 删除（归档）工作项（REST handler，Swagger 注解见下）。
//
//	@Summary		删除工作项（归档）
//	@Tags			issue
//	@Success		204
//	@Router			/issues/{issue_id} [delete]
func (h *IssueHandler) deleteIssue(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	issueID := int64Param(c, "issue_id")

	if err := h.d.IssueSvc.SoftDelete(c.Request.Context(), wsID, issueID); err != nil {
		writeErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// transition 流转工作项状态（REST handler，Swagger 注解见下）。
//
//	@Summary		流转状态
//	@Description	将工作项流转到目标状态
//	@Tags			issue
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	Issue
//	@Failure		422	{object}	errs.AppError	"非法流转"
//	@Router			/issues/{issue_id}/transition [post]
func (h *IssueHandler) transition(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	projectID := c.GetInt64(middleware.CtxProjectID)
	issueID := int64Param(c, "issue_id")
	userID := c.GetInt64(middleware.CtxUserID)

	var req struct {
		ToStateID int64 `json:"to_state_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithError(c, errs.ErrValidation.WithDetails(fieldDetail(err)))
		return
	}

	iss, err := h.d.IssueSvc.Transition(c.Request.Context(), wsID, projectID, issueID, req.ToStateID, userID)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, iss)
}

// reorderIssue 看板拖拽排序（中值插入策略）。
//
//	@Summary		看板排序
//	@Description	在看板视图拖拽工作项到新位置
//	@Tags			issue
//	@Accept			json
//	@Produce		json
//	@Param			body	body	reorderIssueRequest	true	"排序信息"
//	@Success		200		{object}	Issue
//	@Router			/issues/{issue_id}/reorder [patch]
func (h *IssueHandler) reorderIssue(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	issueID := int64Param(c, "issue_id")

	var req reorderIssueRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithError(c, errs.ErrValidation.WithDetails(fieldDetail(err)))
		return
	}

	iss, err := h.d.IssueSvc.Reorder(c.Request.Context(), wsID, issueID, ReorderInput{
		PrevSortOrder: req.PrevSortOrder,
		NextSortOrder: req.NextSortOrder,
	})
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, iss)
}

// listIssues 列出工作项（REST handler，Swagger 注解见下）。
//
//	@Summary		列出工作项
//	@Tags			issue
//	@Produce		json
//	@Param			state_id	query		int		false	"状态 ID"
//	@Param			group		query		string	false	"状态分组 (backlog|started|completed|cancelled)"
//	@Param			type		query		string	false	"类型 (requirement|task|defect)"
//	@Param			priority	query		string	false	"优先级"
//	@Param			parent_id	query		int		false	"父级 ID"
//	@Param			search		query		string	false	"名称搜索"
//	@Param			sort		query		string	false	"排序字段 (-updated_at, priority, target_date, created_at)"
//	@Param			limit		query		int		false	"每页数量 (default 50, max 100)"
//	@Param			offset		query		int		false	"偏移量"
//	@Success		200			{object}	issueListResponse
//	@Router			/issues [get]
func (h *IssueHandler) listIssues(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	projectID := c.GetInt64(middleware.CtxProjectID)

	opts := ListIssuesOptions{
		WorkspaceID: wsID,
		ProjectID:   projectID,
		Limit:       intQuery(c, "limit", 50),
		Offset:      intQuery(c, "offset", 0),
		Search:      c.Query("search"),
		SortBy:      sortField(c.Query("sort")),
		SortDesc:    strings.HasPrefix(c.Query("sort"), "-"),
	}

	if v := c.Query("state_id"); v != "" {
		if id, err := strconv.ParseInt(v, 10, 64); err == nil {
			opts.StateID = &id
		}
	}
	if v := c.Query("group"); v != "" {
		g := StateGroup(v)
		opts.Group = &g
	}
	if v := c.Query("type"); v != "" {
		t := IssueTypeCode(v)
		opts.TypeCode = &t
	}
	if v := c.Query("priority"); v != "" {
		p := IssuePriority(v)
		opts.Priority = &p
	}
	if v := c.Query("parent_id"); v != "" {
		if id, err := strconv.ParseInt(v, 10, 64); err == nil {
			opts.ParentID = &id
		}
	}

	issues, total, err := h.d.IssueSvc.List(c.Request.Context(), opts)
	if err != nil {
		writeErr(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"results": issues, "total": total, "limit": opts.Limit, "offset": opts.Offset})
}

// --- Activity handlers ---

func (h *IssueHandler) listActivities(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	issueID := int64Param(c, "issue_id")
	limit := intQuery(c, "limit", 50)
	offset := intQuery(c, "offset", 0)

	activities, total, err := h.d.ActivitySvc.ListByIssue(c.Request.Context(), wsID, issueID, limit, offset)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"results": activities, "total": total})
}

// --- Time log handlers ---

func (h *IssueHandler) listTimeLogs(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	issueID := int64Param(c, "issue_id")
	limit := intQuery(c, "limit", 50)
	offset := intQuery(c, "offset", 0)

	logs, total, err := h.d.TimeLogSvc.ListByIssue(c.Request.Context(), wsID, issueID, limit, offset)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"results": logs, "total": total})
}

func (h *IssueHandler) createTimeLog(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	projectID := c.GetInt64(middleware.CtxProjectID)
	issueID := int64Param(c, "issue_id")
	userID := c.GetInt64(middleware.CtxUserID)

	var req struct {
		SpentDate       string `json:"spent_date" binding:"required"`
		DurationMinutes int    `json:"duration_minutes" binding:"required,min=1,max=1440"`
		Description     string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithError(c, errs.ErrValidation.WithDetails(fieldDetail(err)))
		return
	}

	tl, err := h.d.TimeLogSvc.Create(c.Request.Context(), CreateTimeLogInput{
		WorkspaceID:     wsID,
		ProjectID:       projectID,
		IssueID:         issueID,
		UserID:          userID,
		SpentDate:        parseDate(req.SpentDate),
		DurationMinutes: req.DurationMinutes,
		Description:     req.Description,
	})
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, tl)
}

// --- request/response types ---

type createIssueRequest struct {
	Type             string         `json:"type" binding:"required,oneof=requirement task defect"`
	Name             string         `json:"name" binding:"required,max=500"`
	DescriptionHTML  string         `json:"description_html"`
	StateID          int64          `json:"state_id"`
	Priority         string         `json:"priority" binding:"omitempty,oneof=urgent high medium low none"`
	ParentID         *int64         `json:"parent_id"`
	Severity         *int           `json:"severity"`
	FoundPhase       *string        `json:"found_phase"`
	ReproduceSteps   map[string]any `json:"reproduce_steps"`
	Category         *string        `json:"category"`
	Source           *string        `json:"source"`
	Assignees        []int64        `json:"assignees"`
	Labels           []int64        `json:"labels"`
	Modules          []int64        `json:"modules"`
	Point            *int           `json:"point"`
	IsDraft          bool           `json:"is_draft"`
}

type updateIssueRequest struct {
	Name              *string `json:"name"`
	DescriptionHTML   *string `json:"description_html"`
	StateID           *int64  `json:"state_id"`
	Priority          *string `json:"priority"`
	ParentID          *int64  `json:"parent_id"`
	Severity          *int    `json:"severity"`
	FoundPhase        *string `json:"found_phase"`
	RootCauseCategory *string `json:"root_cause_category"`
	Category          *string `json:"category"`
	Assignees         []int64 `json:"assignees"`
	Labels            []int64 `json:"labels"`
	Modules           []int64 `json:"modules"`
	Source            *string `json:"source"`
	Version           int     `json:"version" binding:"required"`
}

type reorderIssueRequest struct {
	PrevSortOrder *float64 `json:"prev_sort_order"`
	NextSortOrder *float64 `json:"next_sort_order"`
}

// --- helpers ---

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

func sortField(s string) string {
	if strings.HasPrefix(s, "-") {
		return s[1:]
	}
	return s
}

func parseDate(s string) time.Time {
	if s == "" {
		return time.Time{}
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return time.Time{}
	}
	return t
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

// --- Relation handlers ---

func (h *IssueHandler) listRelations(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	issueID := int64Param(c, "issue_id")

	rels, err := h.d.RelationSvc.ListRelations(c.Request.Context(), wsID, issueID)
	if err != nil {
		writeErr(c, err)
		return
	}
	if rels == nil {
		rels = []IssueRelation{}
	}
	c.JSON(http.StatusOK, gin.H{"results": rels})
}

func (h *IssueHandler) createRelation(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	projectID := c.GetInt64(middleware.CtxProjectID)
	issueID := int64Param(c, "issue_id")
	userID := c.GetInt64(middleware.CtxUserID)

	var req struct {
		TargetIssueID int64  `json:"target_issue_id" binding:"required"`
		RelationType  string `json:"relation_type" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithError(c, errs.ErrValidation.WithDetails(fieldDetail(err)))
		return
	}

	rel, err := h.d.RelationSvc.CreateRelation(c.Request.Context(), CreateRelationInput{
		WorkspaceID:   wsID,
		ProjectID:     projectID,
		SourceIssueID: issueID,
		TargetIssueID: req.TargetIssueID,
		RelationType:  req.RelationType,
		CreatedBy:     userID,
	})
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, rel)
}

func (h *IssueHandler) deleteRelation(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	relationID := int64Param(c, "relation_id")

	if err := h.d.RelationSvc.DeleteRelation(c.Request.Context(), wsID, relationID); err != nil {
		writeErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// --- Dependency handlers ---

func (h *IssueHandler) listDependencies(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	issueID := int64Param(c, "issue_id")

	predecessors, successors, err := h.d.RelationSvc.ListDependencies(c.Request.Context(), wsID, issueID)
	if err != nil {
		writeErr(c, err)
		return
	}
	if predecessors == nil {
		predecessors = []IssueDependency{}
	}
	if successors == nil {
		successors = []IssueDependency{}
	}
	c.JSON(http.StatusOK, gin.H{"predecessors": predecessors, "successors": successors})
}

func (h *IssueHandler) createDependency(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	projectID := c.GetInt64(middleware.CtxProjectID)
	userID := c.GetInt64(middleware.CtxUserID)

	var req struct {
		PredecessorID  int64  `json:"predecessor_id" binding:"required"`
		SuccessorID    int64  `json:"successor_id" binding:"required"`
		DependencyType string `json:"dependency_type" binding:"required,oneof=FS SS FF SF"`
		LagDays        int    `json:"lag_days"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithError(c, errs.ErrValidation.WithDetails(fieldDetail(err)))
		return
	}

	dep, err := h.d.RelationSvc.CreateDependency(c.Request.Context(), CreateDependencyInput{
		WorkspaceID:    wsID,
		ProjectID:      projectID,
		PredecessorID:  req.PredecessorID,
		SuccessorID:    req.SuccessorID,
		DependencyType: req.DependencyType,
		LagDays:        req.LagDays,
		CreatedBy:      userID,
	})
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, dep)
}

func (h *IssueHandler) deleteDependency(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	depID := int64Param(c, "dep_id")

	if err := h.d.RelationSvc.DeleteDependency(c.Request.Context(), wsID, depID); err != nil {
		writeErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// --- Comment handlers ---

func (h *IssueHandler) listComments(c *gin.Context) {
	issueID := int64Param(c, "issue_id")

	comments, err := h.d.CommentSvc.ListByIssue(c.Request.Context(), issueID)
	if err != nil {
		writeErr(c, err)
		return
	}
	if comments == nil {
		comments = []Comment{}
	}
	c.JSON(http.StatusOK, gin.H{"results": comments})
}

func (h *IssueHandler) createComment(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	projectID := c.GetInt64(middleware.CtxProjectID)
	issueID := int64Param(c, "issue_id")
	userID := c.GetInt64(middleware.CtxUserID)

	var req struct {
		ContentJSON     string  `json:"content_json"`
		ContentHTML     string  `json:"content_html"`
		ContentStripped string  `json:"content_stripped"`
		Mentions        []int64 `json:"mentions"`
		ParentID        *int64  `json:"parent_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithError(c, errs.ErrValidation.WithDetails(fieldDetail(err)))
		return
	}

	comment, err := h.d.CommentSvc.Create(c.Request.Context(), CreateCommentInput{
		IssueID:         issueID,
		WorkspaceID:     wsID,
		ProjectID:       projectID,
		ContentJSON:     []byte(req.ContentJSON),
		ContentHTML:     req.ContentHTML,
		ContentStripped: req.ContentStripped,
		CreatedBy:       userID,
		Mentions:        req.Mentions,
		ParentID:        req.ParentID,
	})
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, comment)
}

func (h *IssueHandler) updateComment(c *gin.Context) {
	commentID := int64Param(c, "comment_id")
	userID := c.GetInt64(middleware.CtxUserID)

	var req struct {
		ContentJSON     string  `json:"content_json"`
		ContentHTML     string  `json:"content_html"`
		ContentStripped string  `json:"content_stripped"`
		Mentions        []int64 `json:"mentions"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithError(c, errs.ErrValidation.WithDetails(fieldDetail(err)))
		return
	}

	comment, err := h.d.CommentSvc.Update(c.Request.Context(), commentID, userID, UpdateCommentInput{
		ContentJSON:     []byte(req.ContentJSON),
		ContentHTML:     req.ContentHTML,
		ContentStripped: req.ContentStripped,
		Mentions:        req.Mentions,
	})
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, comment)
}

func (h *IssueHandler) deleteComment(c *gin.Context) {
	commentID := int64Param(c, "comment_id")
	userID := c.GetInt64(middleware.CtxUserID)

	if err := h.d.CommentSvc.Delete(c.Request.Context(), commentID, userID); err != nil {
		writeErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

