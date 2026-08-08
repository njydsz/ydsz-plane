// Package issue — Issue HTTP handlers（REST API）。
package issue

import (
	"archive/zip"
	"context"
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	notif "github.com/njydsz/ydsz-plane/internal/application/notification"
	"github.com/njydsz/ydsz-plane/internal/application/auth"
	"github.com/njydsz/ydsz-plane/internal/interfaces/middleware"
	"github.com/njydsz/ydsz-plane/internal/infrastructure/ws"
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
	SocialSvc      *SocialService
	ProjectInit    *ProjectInitService
	WorkspaceStore *auth.WorkspaceMembershipStore
	// 通知与实时推送（可为 nil，未配置时静默跳过）
	NotificationSvc *notif.Service
	WSHub           *ws.Hub
	// Redis客户端，用于通知去重
	Redis *redis.Client
	// UserNameQuery 按用户 ID 查展示名（用于通知 actor 文案；nil 时回退 "用户"）
	UserNameQuery func(ctx context.Context, userID int64) string
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
	r.POST("/issues/batch", h.batchIssues)
	r.GET("/issues/export", h.exportIssues)

	// 模块（Module 体系，对标 Plane 的 Module 概念）
	modHandler := NewModuleHandler(NewModuleService(h.d.IssueSvc.db))
	modHandler.Register(r)

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
		issue.PATCH("/time-logs/:log_id", h.updateTimeLog)
		issue.DELETE("/time-logs/:log_id", h.deleteTimeLog)
		issue.GET("/relations", h.listRelations)
		issue.POST("/relations", h.createRelation)
		issue.DELETE("/relations/:relation_id", h.deleteRelation)
		issue.GET("/dependencies", h.listDependencies)
		issue.POST("/dependencies", h.createDependency)
		issue.DELETE("/dependencies/:dep_id", h.deleteDependency)
		// 社交反馈：表情反应 + 投票
		issue.GET("/reactions", h.listReactions)
		issue.POST("/reactions", h.addReaction)
		issue.DELETE("/reactions/:reaction_type", h.removeReaction)
		issue.GET("/vote", h.voteSummary)
		issue.POST("/vote", h.voteIssue)
		issue.DELETE("/vote", h.removeVote)
		// 关注（watchers 订阅）
		issue.POST("/watch", h.watchIssue)
		issue.DELETE("/watch", h.unwatchIssue)
		// 看板排序（预留，S4 完善）
		issue.PATCH("/reorder", h.reorderIssue)
		// 评论
		issue.GET("/comments", h.listComments)
		issue.POST("/comments", h.createComment)
		issue.PATCH("/comments/:comment_id", h.updateComment)
		issue.DELETE("/comments/:comment_id", h.deleteComment)
		// 版本快照审计（历史回溯 / 字段 diff）
		verHandler := NewVersionHandler(NewVersionService(h.d.IssueSvc.db))
		verHandler.Register(issue)
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
		Priority:         defaultPriority(req.Priority),
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

	// 通知被指派者 + 广播工作项创建
	h.notifyIssueCreated(c.Request.Context(), wsID, req.Assignees, userID,
		h.actorName(c, userID), iss.Name, iss.ID)
	h.broadcastIssueUpdated(c.Request.Context(), wsID, projectID, iss.ID, userID, int64(iss.Version))

	c.JSON(http.StatusCreated, iss)
}

// batchIssues 批量操作工作项（流转/指派/优先级/删除）。
//
//	@Summary		批量操作工作项
//	@Description	支持批量流转、指派、变更优先级、删除
//	@Tags			issue
//	@Accept			json
//	@Produce		json
//	@Param			body	body	batchIssuesRequest	true	"批量操作信息"
//	@Success		200		{object}	map[string]interface{}
//	@Failure		422		{object}	errs.AppError
//	@Router			/issues/batch [post]
func (h *IssueHandler) batchIssues(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	projectID := c.GetInt64(middleware.CtxProjectID)
	userID := c.GetInt64(middleware.CtxUserID)

	var req batchIssuesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithError(c, errs.ErrValidation.WithDetails(fieldDetail(err)))
		return
	}

	if len(req.IssueIDs) == 0 {
		middleware.AbortWithError(c, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "issue_ids", Reason: "至少选择一个工作项"}))
		return
	}
	if len(req.IssueIDs) > 100 {
		middleware.AbortWithError(c, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "issue_ids", Reason: "单次批量操作不超过100项"}))
		return
	}

	result, err := h.d.IssueSvc.BatchUpdate(c.Request.Context(), wsID, projectID, userID, BatchUpdateInput{
		IssueIDs:   req.IssueIDs,
		ToStateID:  req.ToStateID,
		AssigneeID: req.AssigneeID,
		Priority:   req.Priority,
		Delete:     req.Delete,
	})
	if err != nil {
		writeErr(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"succeeded": result.Succeeded, "failed": result.Failed})
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
	projectID := c.GetInt64(middleware.CtxProjectID)
	issueID := int64Param(c, "issue_id")
	userID := c.GetInt64(middleware.CtxUserID)

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
	if req.FoundVersionID != nil {
		in.FoundVersionID = req.FoundVersionID
	}
	if req.FixVersionID != nil {
		in.FixVersionID = req.FixVersionID
	}
	if req.ReleaseVersionID != nil {
		in.ReleaseVersionID = req.ReleaseVersionID
	}

	iss, err := h.d.IssueSvc.Update(c.Request.Context(), wsID, issueID, in)
	if err != nil {
		writeErr(c, err)
		return
	}
	// 广播 + 仅核心事件通知关注者
	h.broadcastIssueUpdated(c.Request.Context(), wsID, projectID, iss.ID, userID, int64(iss.Version))
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
	// 广播状态变更（看板实时刷新）
	h.broadcastIssueUpdated(c.Request.Context(), wsID, projectID, iss.ID, userID, int64(iss.Version))
	// 通知关注者，传递核心事件类型issue.status_changed，触发通知
	h.notifyIssueWatchers(c.Request.Context(), wsID, iss.ID, userID, "issue.status_changed", h.actorName(c, userID), iss.Name, "工作项状态已变更")
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
		Version:       req.Version,
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
//	@Param			type		query		string	false	"类型 (epic|requirement|task|defect)"
//	@Param			priority	query		string	false	"优先级"
//	@Param			parent_id	query		int		false	"父级 ID"
//	@Param			search		query		string	false	"名称搜索"
//	@Param			assignee_id	query		int		false	"指派人 ID"
//	@Param			label_id	query		int		false	"标签 ID"
//	@Param			module_id	query		int		false	"模块 ID"
//	@Param			sprint_id	query		int		false	"迭代 ID"
//	@Param			start_date_from	query	string	false	"开始日期起 (ISO)"
//	@Param			target_date_to	query	string	false	"截止日期止 (ISO)"
//	@Param			severity_from	query	int		false	"最低严重级别"
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
	if v := c.Query("assignee_id"); v != "" {
		if id, err := strconv.ParseInt(v, 10, 64); err == nil {
			opts.AssigneeID = &id
		}
	}
	if v := c.Query("label_id"); v != "" {
		if id, err := strconv.ParseInt(v, 10, 64); err == nil {
			opts.LabelID = &id
		}
	}
	if v := c.Query("module_id"); v != "" {
		if id, err := strconv.ParseInt(v, 10, 64); err == nil {
			opts.ModuleID = &id
		}
	}
	if v := c.Query("sprint_id"); v != "" {
		if id, err := strconv.ParseInt(v, 10, 64); err == nil {
			opts.SprintID = &id
		}
	}
	if v := c.Query("start_date_from"); v != "" {
		opts.StartDateFrom = &v
	}
	if v := c.Query("target_date_to"); v != "" {
		opts.TargetDateTo = &v
	}
	if v := c.Query("severity_from"); v != "" {
		if sv, err := strconv.Atoi(v); err == nil {
			opts.SeverityFrom = &sv
		}
	}

	issues, total, err := h.d.IssueSvc.List(c.Request.Context(), opts)
	if err != nil {
		writeErr(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"results": issues, "total": total, "limit": opts.Limit, "offset": opts.Offset})
}

// exportIssues 导出工作项为 CSV 或 xlsx 文件。
//
// 查询参数：
//   - format: csv（默认）| xlsx — 导出格式
//   - type / state_id / search: 过滤条件（同列表接口）
func (h *IssueHandler) exportIssues(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	projectID := c.GetInt64(middleware.CtxProjectID)

	opts := ListIssuesOptions{
		WorkspaceID: wsID,
		ProjectID:   projectID,
		Limit:       5000, // 导出上限
		Offset:      0,
		Search:      c.Query("search"),
	}
	if v := c.Query("type"); v != "" {
		t := IssueTypeCode(v)
		opts.TypeCode = &t
	}
	if v := c.Query("state_id"); v != "" {
		if id, err := strconv.ParseInt(v, 10, 64); err == nil {
			opts.StateID = &id
		}
	}

	issues, _, err := h.d.IssueSvc.List(c.Request.Context(), opts)
	if err != nil {
		writeErr(c, err)
		return
	}

	format := c.Query("format")
	if format == "xlsx" {
		writeXLSX(c, issues)
		return
	}

	// 默认 CSV 格式（含 UTF-8 BOM 兼容 Excel）
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition",
		fmt.Sprintf("attachment; filename=issues-export-%s.csv", time.Now().Format("20060102")))
	// BOM for Excel
	_, _ = c.Writer.Write([]byte{0xEF, 0xBB, 0xBF})

	w := csv.NewWriter(c.Writer)
	_ = w.Write([]string{"编号", "类型", "名称", "状态", "优先级", "严重级别", "点数", "指派人", "创建时间", "更新时间"})

	for _, iss := range issues {
		assignees := ""
		if len(iss.Assignees) > 0 {
			assignees = strings.Trim(strings.Join(strings.Fields(fmt.Sprint(iss.Assignees)), ","), "[]")
		}
		severity := ""
		if iss.Severity != nil {
			severity = fmt.Sprintf("S%d", *iss.Severity)
		}
		stateName := ""
		if iss.State != nil {
			stateName = iss.State.Name
		}
		_ = w.Write([]string{
			iss.Identifier,
			string(iss.TypeCode),
			iss.Name,
			stateName,
			string(iss.Priority),
			severity,
			fmt.Sprintf("%d", func() int { if iss.Point != nil { return *iss.Point }; return 0 }()),
			assignees,
			iss.CreatedAt.Format("2006-01-02 15:04"),
			iss.UpdatedAt.Format("2006-01-02 15:04"),
		})
	}
	w.Flush()
}

// xlsxTemplate 是 OOXML 最小化模板，用于纯标准库生成 .xlsx。
// 参考 ECMA-376 第 4 版 SpreadsheetML 规范。
const xlsxTemplate = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">
<cols>
  <col min="1" max="1" width="14" customWidth="1"/>
  <col min="2" max="2" width="10" customWidth="1"/>
  <col min="3" max="3" width="40" customWidth="1"/>
  <col min="4" max="4" width="12" customWidth="1"/>
  <col min="5" max="5" width="10" customWidth="1"/>
  <col min="6" max="6" width="10" customWidth="1"/>
  <col min="7" max="7" width="8" customWidth="1"/>
  <col min="8" max="8" width="16" customWidth="1"/>
  <col min="9" max="9" width="18" customWidth="1"/>
  <col min="10" max="10" width="18" customWidth="1"/>
</cols>
<sheetData>%s</sheetData></worksheet>`

// xlsxRow 是一行单元格 XML 片段。
const xlsxRow = `<row>%s</row>`

// xlsxCell 是一个单元格 XML 片段（内联字符串）。
const xlsxCell = `<c t="inlineStr"><is><t>%s</t></is></c>`

// writeXLSX 使用纯 Go 标准库生成 .xlsx 文件并写入响应体。
//
// .xlsx 本质是 ZIP 归档，内含固定的 XML 文件集合。
// 此处生成最小有效 xlsx（含 [Content_Types].xml / _rels/.rels / xl/workbook.xml / xl/worksheets/sheet1.xml）。
func writeXLSX(c *gin.Context, issues []Issue) {
	filename := fmt.Sprintf("issues-export-%s.xlsx", time.Now().Format("20060102"))
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))

	zw := zip.NewWriter(c.Writer)
	defer zw.Close()

	// 辅助函数：向 ZIP 中添加一个文件
	addZIPFile := func(name string, content string) {
		w, err := zw.CreateHeader(&zip.FileHeader{
			Name:   name,
			Method: zip.Deflate,
		})
		if err != nil {
			return
		}
		fmt.Fprint(w, content)
	}

	// [Content_Types].xml — 声明部件内容类型
	addZIPFile("[Content_Types].xml",
		`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`+
			`<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">`+
			`<Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>`+
			`<Default Extension="xml" ContentType="application/xml"/>`+
			`<Override PartName="/xl/workbook.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"/>`+
			`<Override PartName="/xl/worksheets/sheet1.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"/>`+
			`</Types>`)

	// _rels/.rels — 包级关系
	addZIPFile("_rels/.rels",
		`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`+
			`<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">`+
			`<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="xl/workbook.xml"/>`+
			`</Relationships>`)

	// xl/workbook.xml — 工作簿定义
	addZIPFile("xl/workbook.xml",
		`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`+
			`<workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">`+
			`<sheets><sheet name="工作项" sheetId="1" r:id="rId1"/></sheets>`+
			`</workbook>`)

	// xl/_rels/workbook.xml.rels
	addZIPFile("xl/_rels/workbook.xml.rels",
		`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`+
			`<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">`+
			`<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet1.xml"/>`+
			`</Relationships>`)

	// 构建表头行
	headerCells := buildXLSXRow([]string{
		"编号", "类型", "名称", "状态", "优先级", "严重级别", "点数", "指派人", "创建时间", "更新时间",
	})

	// 构建数据行
	var dataRows strings.Builder
	for _, iss := range issues {
		assignees := ""
		if len(iss.Assignees) > 0 {
			assignees = strings.Trim(strings.Join(strings.Fields(fmt.Sprint(iss.Assignees)), ","), "[]")
		}
		severity := ""
		if iss.Severity != nil {
			severity = fmt.Sprintf("S%d", *iss.Severity)
		}
		stateName := ""
		if iss.State != nil {
			stateName = iss.State.Name
		}
		row := buildXLSXRow([]string{
			iss.Identifier,
			string(iss.TypeCode),
			iss.Name,
			stateName,
			string(iss.Priority),
			severity,
			fmt.Sprintf("%d", func() int { if iss.Point != nil { return *iss.Point }; return 0 }()),
			assignees,
			iss.CreatedAt.Format("2006-01-02 15:04"),
			iss.UpdatedAt.Format("2006-01-02 15:04"),
		})
		dataRows.WriteString(row)
	}

	// xl/worksheets/sheet1.xml — 数据表
	sheetXML := fmt.Sprintf(xlsxTemplate, headerCells+dataRows.String())
	addZIPFile("xl/worksheets/sheet1.xml",
		`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`+sheetXML)
}

// buildXLSXRow 构建一行 xlsx 单元格 XML。
// 所有值作为内联字符串（t="inlineStr"）写入，避免数字/日期类型歧义。
func buildXLSXRow(vals []string) string {
	var cells strings.Builder
	for _, v := range vals {
		// XML 转义：& < > " '
		escaped := xmlEscape(v)
		cells.WriteString(fmt.Sprintf(xlsxCell, escaped))
	}
	return fmt.Sprintf(xlsxRow, cells.String())
}

// xmlEscape 转义 XML 特殊字符。
func xmlEscape(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	return s
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

func (h *IssueHandler) updateTimeLog(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	logID := int64Param(c, "log_id")

	var req struct {
		SpentDate       string `json:"spent_date" binding:"required"`
		DurationMinutes int    `json:"duration_minutes" binding:"required,min=1,max=1440"`
		Description     string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithError(c, errs.ErrValidation.WithDetails(fieldDetail(err)))
		return
	}

	tl, err := h.d.TimeLogSvc.Update(c.Request.Context(), wsID, logID,
		req.DurationMinutes, req.Description, parseDate(req.SpentDate))
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, tl)
}

func (h *IssueHandler) deleteTimeLog(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	logID := int64Param(c, "log_id")

	if err := h.d.TimeLogSvc.Delete(c.Request.Context(), wsID, logID); err != nil {
		writeErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// --- request/response types ---

type createIssueRequest struct {
	Type             string         `json:"type" binding:"required,oneof=epic requirement task defect"`
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
	FoundVersionID    *int64  `json:"found_version_id"`
	FixVersionID      *int64  `json:"fix_version_id"`
	ReleaseVersionID  *int64  `json:"release_version_id"`
}

type reorderIssueRequest struct {
	PrevSortOrder *float64 `json:"prev_sort_order"`
	NextSortOrder *float64 `json:"next_sort_order"`
	Version       *int     `json:"version"`
}

type batchIssuesRequest struct {
	IssueIDs   []int64 `json:"issue_ids" binding:"required,min=1,max=100"`
	ToStateID  *int64  `json:"to_state_id"`
	AssigneeID *int64  `json:"assignee_id"`
	Priority   *string `json:"priority"`
	Delete     bool    `json:"delete"`
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

// defaultPriority 请求未指定优先级时默认 'none'，满足数据库 CHECK 约束。
func defaultPriority(p string) IssuePriority {
	if p == "" {
		return PriorityNone
	}
	return IssuePriority(p)
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

	// XSS 防御：服务端二次净化 HTML（客户端 ProseMirror 渲染不可信）
	htmlSanitized := SanitizeHTML(req.ContentHTML)
	strippedSafe := StripHTML(req.ContentStripped)

	comment, err := h.d.CommentSvc.CreateWithEvent(c.Request.Context(), CreateCommentInput{
		IssueID:         issueID,
		WorkspaceID:     wsID,
		ProjectID:       projectID,
		ContentJSON:     []byte(req.ContentJSON),
		ContentHTML:     htmlSanitized,
		ContentStripped: strippedSafe,
		CreatedBy:       userID,
		Mentions:        req.Mentions,
		ParentID:        req.ParentID,
	})
	if err != nil {
		writeErr(c, err)
		return
	}

	// 通知被 @ 提及的用户 + 广播评论事件
	h.notifyCommentCreated(c.Request.Context(), wsID, issueID, req.Mentions, userID,
		comment.CreatorName, h.issueTitle(c, wsID, issueID))
	// 通知关注者（有人评论了我关注的工作项）
	h.notifyIssueWatchers(c.Request.Context(), wsID, issueID, userID, "issue.commented",
		comment.CreatorName, h.issueTitle(c, wsID, issueID), "你关注的工作项有新评论")
	h.broadcastIssueUpdated(c.Request.Context(), wsID, projectID, issueID, userID, -1)

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

	// XSS 防御：服务端二次净化 HTML
	htmlSanitized := SanitizeHTML(req.ContentHTML)
	strippedSafe := StripHTML(req.ContentStripped)

	comment, err := h.d.CommentSvc.Update(c.Request.Context(), commentID, userID, UpdateCommentInput{
		ContentJSON:     []byte(req.ContentJSON),
		ContentHTML:     htmlSanitized,
		ContentStripped: strippedSafe,
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

// --- Social handlers (Reaction / Vote / Watch) ---

func (h *IssueHandler) listReactions(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	issueID := int64Param(c, "issue_id")
	userID := c.GetInt64(middleware.CtxUserID)

	if h.d.SocialSvc == nil {
		middleware.AbortWithError(c, errs.ErrInternal.WithDetails(errs.FieldDetail{Field: "service", Reason: "社交服务未启用"}))
		return
	}
	reactions, err := h.d.SocialSvc.ListReactions(c.Request.Context(), wsID, issueID, userID)
	if err != nil {
		writeErr(c, err)
		return
	}
	if reactions == nil {
		reactions = []ReactionSummary{}
	}
	c.JSON(http.StatusOK, gin.H{"results": reactions})
}

func (h *IssueHandler) addReaction(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	projectID := c.GetInt64(middleware.CtxProjectID)
	issueID := int64Param(c, "issue_id")
	userID := c.GetInt64(middleware.CtxUserID)

	var req struct {
		ReactionType string `json:"reaction_type" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithError(c, errs.ErrValidation.WithDetails(fieldDetail(err)))
		return
	}

	reaction, created, err := h.d.SocialSvc.AddReaction(c.Request.Context(), wsID, projectID, issueID, userID, req.ReactionType)
	if err != nil {
		writeErr(c, err)
		return
	}
	status := http.StatusCreated
	if !created {
		status = http.StatusOK
	}
	c.JSON(status, reaction)
}

func (h *IssueHandler) removeReaction(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	issueID := int64Param(c, "issue_id")
	userID := c.GetInt64(middleware.CtxUserID)
	reactionType := c.Param("reaction_type")

	if err := h.d.SocialSvc.RemoveReaction(c.Request.Context(), wsID, issueID, userID, reactionType); err != nil {
		writeErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *IssueHandler) voteSummary(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	issueID := int64Param(c, "issue_id")
	userID := c.GetInt64(middleware.CtxUserID)

	summary, err := h.d.SocialSvc.VoteSummary(c.Request.Context(), wsID, issueID, userID)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, summary)
}

func (h *IssueHandler) voteIssue(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	projectID := c.GetInt64(middleware.CtxProjectID)
	issueID := int64Param(c, "issue_id")
	userID := c.GetInt64(middleware.CtxUserID)

	var req struct {
		Vote int `json:"vote" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithError(c, errs.ErrValidation.WithDetails(fieldDetail(err)))
		return
	}

	vote, err := h.d.SocialSvc.VoteIssue(c.Request.Context(), wsID, projectID, issueID, userID, req.Vote)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, vote)
}

func (h *IssueHandler) removeVote(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	issueID := int64Param(c, "issue_id")
	userID := c.GetInt64(middleware.CtxUserID)

	if err := h.d.SocialSvc.RemoveVote(c.Request.Context(), wsID, issueID, userID); err != nil {
		writeErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *IssueHandler) watchIssue(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	issueID := int64Param(c, "issue_id")
	userID := c.GetInt64(middleware.CtxUserID)

	if err := h.d.IssueSvc.Watch(c.Request.Context(), wsID, issueID, userID); err != nil {
		writeErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *IssueHandler) unwatchIssue(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	issueID := int64Param(c, "issue_id")
	userID := c.GetInt64(middleware.CtxUserID)

	if err := h.d.IssueSvc.Unwatch(c.Request.Context(), wsID, issueID, userID); err != nil {
		writeErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

