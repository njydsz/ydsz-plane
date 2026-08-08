// Package issue — Issue 版本快照审计 HTTP handlers。
//
// 暴露：
//   - GET /issues/:issue_id/versions               版本历史列表
//   - GET /issues/:issue_id/versions/:version       单条快照
//   - POST /issues/:issue_id/versions/diff          两版本对比
package issue

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/njydsz/ydsz-plane/internal/interfaces/middleware"
	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// VersionHandler Gin handler（版本快照）。
type VersionHandler struct {
	svc *VersionService
}

// NewVersionHandler 构造版本 handler。
func NewVersionHandler(svc *VersionService) *VersionHandler {
	return &VersionHandler{svc: svc}
}

// Register 注册版本路由（在项目子路由组下）。
func (h *VersionHandler) Register(r *gin.RouterGroup) {
	vers := r.Group("/issues/:issue_id/versions")
	{
		vers.GET("", h.listVersions)
		vers.GET("/:version", h.getVersion)
		vers.POST("/diff", h.diffVersions)
	}
}

// ---- request ----

type diffVersionsRequest struct {
	FromVersion int `json:"from_version" binding:"required,gt=0"`
	ToVersion   int `json:"to_version" binding:"required,gt=0"`
}

// ---- handlers ----

// listVersions GET /api/v1/workspaces/:ws_id/projects/:project_id/issues/:issue_id/versions
// @Summary		列出版本历史
// @Tags			issue-version
// @Produce		json
// @Param			limit		query		int		false	"返回条数（1-100，默认 50）"
// @Success		200			{array}	IssueVersion
// @Router			/issues/{issue_id}/versions [get]
func (h *VersionHandler) listVersions(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	issueID := extractIssueID(c)
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	versions, err := h.svc.ListVersions(c.Request.Context(), wsID, issueID, limit)
	if err != nil {
		handleVersionErr(c, err)
		return
	}
	c.JSON(http.StatusOK, versions)
}

// getVersion GET /issues/:issue_id/versions/:version
func (h *VersionHandler) getVersion(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	issueID := extractIssueID(c)
	version, err := strconv.Atoi(c.Param("version"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的版本号"})
		return
	}

	v, err := h.svc.GetVersion(c.Request.Context(), wsID, issueID, version)
	if err != nil {
		handleVersionErr(c, err)
		return
	}
	c.JSON(http.StatusOK, v)
}

// diffVersions POST /issues/:issue_id/versions/diff
// @Summary		对比两个版本的字段差异
// @Tags			issue-version
// @Accept			json
// @Produce		json
// @Param			body	body		diffVersionsRequest	true	"from/to 版本号"
// @Success		200		{object}	VersionDiff
// @Router			/issues/{issue_id}/versions/diff [post]
func (h *VersionHandler) diffVersions(c *gin.Context) {
	var req diffVersionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": errs.ErrValidation.WithDetails(errs.FieldDetail{
			Field: "body", Reason: "请求体无效",
		})})
		return
	}

	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	issueID := extractIssueID(c)

	diff, err := h.svc.DiffVersions(c.Request.Context(), wsID, issueID, req.FromVersion, req.ToVersion)
	if err != nil {
		handleVersionErr(c, err)
		return
	}
	c.JSON(http.StatusOK, diff)
}

// ---- helpers ----

func extractIssueID(c *gin.Context) int64 {
	id, _ := strconv.ParseInt(c.Param("issue_id"), 10, 64)
	return id
}

func handleVersionErr(c *gin.Context, err error) {
	var appErr *errs.AppError
	if errs.As(err, &appErr) {
		middleware.AbortWithError(c, appErr)
		return
	}
	middleware.AbortWithError(c, errs.ErrInternal)
}
