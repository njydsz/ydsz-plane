// Package issue — 需求评审工作流 HTTP handlers。
//
// 提供评审提交、评审人决定（采纳/驳回）、评审列表等 REST 端点。
package issue

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/njydsz/ydsz-plane/internal/interfaces/middleware"
	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// listReviews 查询评审记录。
//
//	@Summary		评审记录列表
//	@Tags			issue-review
//	@Produce		json
//	@Param			issue_id	path	int	true	"工作项 ID"
//	@Success		200			{object}	map[string]any
//	@Router			/issues/{issue_id}/reviews [get]
func (h *IssueHandler) listReviews(c *gin.Context) {
	issueID := int64Param(c, "issue_id")
	reviews, err := NewReviewService(h.d.IssueSvc.db).ListReviews(c.Request.Context(), issueID)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"results": reviews, "total": len(reviews)})
}

// submitReview 提交评审。
//
//	@Summary		提交评审
//	@Description	提交需求评审并指定评审人
//	@Tags			issue-review
//	@Accept			json
//	@Produce		json
//	@Param			issue_id	path	int	true	"工作项 ID"
//	@Success		201			{object}	Review
//	@Router			/issues/{issue_id}/review [post]
func (h *IssueHandler) submitReview(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	projectID := c.GetInt64(middleware.CtxProjectID)
	userID := c.GetInt64(middleware.CtxUserID)
	issueID := int64Param(c, "issue_id")

	var req struct {
		Name      string  `json:"name"`
		Reviewers []int64 `json:"reviewers"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		writeErr(c, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "body", Reason: "请求体解析失败"}))
		return
	}

	review, err := NewReviewService(h.d.IssueSvc.db).SubmitReview(c.Request.Context(), SubmitReviewInput{
		WorkspaceID: wsID, ProjectID: projectID, IssueID: issueID, UserID: userID,
		Name: req.Name, Reviewers: req.Reviewers,
	})
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, review)
}

// decideReview 评审决定（approved/rejected）。
//
//	@Summary		评审决定
//	@Description	评审人对需求评审做出通过或驳回决定
//	@Tags			issue-review
//	@Accept			json
//	@Produce		json
//	@Param			issue_id	path	int	true	"工作项 ID"
//	@Success		200			{interface}	interface{}
//	@Router			/issues/{issue_id}/review/decision [post]
func (h *IssueHandler) decideReview(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	userID := c.GetInt64(middleware.CtxUserID)
	issueID := int64Param(c, "issue_id")

	var req struct {
		Decision string `json:"decision" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		writeErr(c, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "decision", Reason: "decision 必填"}))
		return
	}

	if err := NewReviewService(h.d.IssueSvc.db).DecideReview(c.Request.Context(), DecideReviewInput{
		WorkspaceID: wsID, IssueID: issueID, UserID: userID, Decision: req.Decision,
	}); err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
