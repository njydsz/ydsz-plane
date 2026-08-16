// Package issue — 需求评审工作流 handler。
package issue

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/njydsz/ydsz-plane/internal/interfaces/middleware"
	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// listReviews GET /issues/:issue_id/reviews — 查询评审记录。
func (h *IssueHandler) listReviews(c *gin.Context) {
	issueID := int64Param(c, "issue_id")
	reviews, err := NewReviewService(h.d.IssueSvc.db).ListReviews(c.Request.Context(), issueID)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"results": reviews, "total": len(reviews)})
}

// submitReview POST /issues/:issue_id/review — 提交评审。
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

// decideReview POST /issues/:issue_id/review/decision — 评审决定（approved/rejected）。
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
