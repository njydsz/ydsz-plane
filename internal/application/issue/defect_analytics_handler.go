// Package issue — 缺陷分析报表 HTTP handler。
//
// 提供 REST 端点供前端 ECharts 图表消费（对齐 Jira Issue Statistics / Plane Analytics）。
package issue

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/njydsz/ydsz-plane/internal/interfaces/http/export"
	"github.com/njydsz/ydsz-plane/internal/interfaces/middleware"
	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// DefectAnalyticsHandler 缺陷分析 HTTP handler 集合。
type DefectAnalyticsHandler struct {
	svc *DefectAnalyticsService
}

// NewDefectAnalyticsHandler 构造 handler。
func NewDefectAnalyticsHandler(svc *DefectAnalyticsService) *DefectAnalyticsHandler {
	return &DefectAnalyticsHandler{svc: svc}
}

// Register 注册缺陷分析路由（挂载到项目子路由组下，鉴权与 RBAC 已在父路由组应用）。
func (h *DefectAnalyticsHandler) Register(r *gin.RouterGroup) {
	analytics := r.Group("/analytics")
	{
		analytics.GET("/defects", h.getDefectAnalytics)
		analytics.GET("/defects/export", h.exportDefectAnalytics)
	}
}

// analyticsQuery 从查询参数解析分析过滤条件。
func analyticsQuery(c *gin.Context) AnalyticsQuery {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	projectID := c.GetInt64(middleware.CtxProjectID)
	q := AnalyticsQuery{WorkspaceID: wsID, ProjectID: projectID}

	if from := c.Query("date_from"); from != "" {
		if t, err := time.Parse("2006-01-02", from); err == nil {
			q.DateFrom = &t
		}
	}
	if to := c.Query("date_to"); to != "" {
		if t, err := time.Parse("2006-01-02", to); err == nil {
			q.DateTo = &t
		}
	}
	if sevFrom := c.Query("severity_from"); sevFrom != "" {
		if v, err := atoi(sevFrom); err == nil {
			q.SeverityFrom = &v
		}
	}
	if sevTo := c.Query("severity_to"); sevTo != "" {
		if v, err := atoi(sevTo); err == nil {
			q.SeverityTo = &v
		}
	}
	if mid := c.Query("module_id"); mid != "" {
		if v, err := atoi64(mid); err == nil {
			q.ModuleID = &v
		}
	}
	if vid := c.Query("version_id"); vid != "" {
		if v, err := atoi64(vid); err == nil {
			q.VersionID = &v
		}
	}
	return q
}

// getDefectAnalytics godoc
//
//	@Summary		缺陷分析聚合
//	@Description	输出按模块/严重程度/发现阶段/根因分类/缺陷龄分布/周趋势聚合的缺陷统计数据
//	@Tags			analytics
//	@Accept			json
//	@Produce		json
//	@Param			workspace_id	path		int		true	"工作空间 ID"
//	@Param			project_id		path		int		true	"项目 ID"
//	@Param			date_from		query		string	false	"起始日期（YYYY-MM-DD）"
//	@Param			date_to			query		string	false	"结束日期（YYYY-MM-DD）"
//	@Param			severity_from	query		int		false	"最低严重程度（1-5）"
//	@Param			severity_to		query		int		false	"最高严重程度（1-5）"
//	@Param			module_id		query		int		false	"按模块过滤"
//	@Success		200				{object}	DefectAnalytics
//	@Failure		400				{object}	errs.AppError
//	@Failure		401				{object}	errs.AppError
//	@Router			/analytics/defects [get]
func (h *DefectAnalyticsHandler) getDefectAnalytics(c *gin.Context) {
	q := analyticsQuery(c)
	data, err := h.svc.GetAnalytics(c.Request.Context(), q)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, data)
}

// exportDefectAnalytics 导出缺陷明细为 CSV 或 xlsx 文件。
//
//	@Summary		导出缺陷明细
//	@Description	按与聚合分析相同的过滤条件导出缺陷明细列表（CSV / xlsx）
//	@Tags			analytics
//	@Accept			json
//	@Produce		application/octet-stream
//	@Param			workspace_id	path		int		true	"工作空间 ID"
//	@Param			project_id		path		int		true	"项目 ID"
//	@Param			format			query		string	false	"导出格式：csv（默认）| xlsx"
//	@Param			date_from		query		string	false	"起始日期（YYYY-MM-DD）"
//	@Param			date_to			query		string	false	"结束日期（YYYY-MM-DD）"
//	@Param			severity_from	query		int		false	"最低严重程度（1-5）"
//	@Param			severity_to		query		int		false	"最高严重程度（1-5）"
//	@Param			module_id		query		int		false	"按模块过滤"
//	@Param			version_id		query		int		false	"按版本过滤（发现或修复版本）"
//	@Failure		400				{object}	errs.AppError
//	@Failure		401				{object}	errs.AppError
//	@Router			/analytics/defects/export [get]
func (h *DefectAnalyticsHandler) exportDefectAnalytics(c *gin.Context) {
	q := analyticsQuery(c)
	rows, err := h.svc.ExportDefects(c.Request.Context(), q, 5000)
	if err != nil {
		writeErr(c, err)
		return
	}

	headers := []string{"编号", "名称", "严重程度", "状态", "发现阶段", "根因分类", "模块", "创建时间", "完成时间", "缺陷龄(天)"}
	data := make([][]string, 0, len(rows))
	for _, r := range rows {
		severity := ""
		if r.Severity != nil {
			severity = fmt.Sprintf("%s(S%d)", severityLabels[*r.Severity], *r.Severity)
		}
		foundPhase := ""
		if r.FoundPhase != nil {
			foundPhase = *r.FoundPhase
		}
		rootCause := ""
		if r.RootCause != nil {
			rootCause = *r.RootCause
		}
		completedAt := ""
		if r.CompletedAt != nil {
			completedAt = r.CompletedAt.Format("2006-01-02 15:04")
		}
		data = append(data, []string{
			r.Identifier,
			r.Name,
			severity,
			r.StateName,
			foundPhase,
			rootCause,
			r.ModuleNames,
			r.CreatedAt.Format("2006-01-02 15:04"),
			completedAt,
			fmt.Sprintf("%.1f", r.AgeDays),
		})
	}

	date := time.Now().Format("20060102")
	switch c.Query("format") {
	case "xlsx":
		export.WriteXLSX(c, "缺陷明细", "defects-export-"+date+".xlsx", headers, data)
	default:
		export.WriteCSV(c, "defects-export-"+date+".csv", headers, data)
	}
}

// atoi 简易字符串 → int。
func atoi(s string) (int, error) {
	out := 0
	for _, ch := range s {
		if ch < '0' || ch > '9' {
			return 0, errs.ErrValidation
		}
		out = out*10 + int(ch-'0')
	}
	return out, nil
}

// atoi64 简易字符串 → int64。
func atoi64(s string) (int64, error) {
	var out int64
	for _, ch := range s {
		if ch < '0' || ch > '9' {
			return 0, errs.ErrValidation
		}
		out = out*10 + int64(ch-'0')
	}
	return out, nil
}
