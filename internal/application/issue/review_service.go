// Package issue — 需求评审工作流（PRD 5.3.3）。
//
// 评审以工作项（重点为需求）为粒度：提交评审 → 评审人决定 → 采纳/驳回。
// 状态通过 requirement.review_status 承载；评审活动与评审人落 reviews / review_assignments。
package issue

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// ReviewStatus 需求评审状态。
type ReviewStatus = string

const (
	ReviewDraft    ReviewStatus = "draft"     // 草稿
	ReviewInReview ReviewStatus = "in_review" // 评审中
	ReviewApproved ReviewStatus = "approved"  // 已采纳
	ReviewRejected ReviewStatus = "rejected"  // 已拒绝
)

// Review 评审活动记录（reviews 表）。
type Review struct {
	ID            int64      `json:"id"`
	WorkspaceID   int64      `json:"workspace_id"`
	ProjectID     *int64     `json:"project_id,omitempty"`
	Name          string     `json:"name"`
	ReviewType    string     `json:"review_type"`
	EntityType    string     `json:"entity_type"`
	EntityID      *int64     `json:"entity_id,omitempty"`
	Status        string     `json:"status"`
	Description   string     `json:"description,omitempty"`
	DueDate       *time.Time `json:"due_date,omitempty"`
	CreatedDate   *time.Time `json:"created_date,omitempty"`
	CompletedDate *time.Time `json:"completed_date,omitempty"`
	CreatedBy     int64      `json:"created_by"`
	CreatedAt     time.Time  `json:"created_at"`
	Reviewers     []int64    `json:"reviewers,omitempty"`
}

// ReviewService 评审工作流应用服务。
type ReviewService struct {
	db *pgxpool.Pool
}

// NewReviewService 创建评审服务。
func NewReviewService(db *pgxpool.Pool) *ReviewService {
	return &ReviewService{db: db}
}

// SubmitReviewInput 提交评审入参。
type SubmitReviewInput struct {
	WorkspaceID int64
	ProjectID   int64
	IssueID     int64
	UserID      int64
	Name        string
	Reviewers   []int64
}

// SubmitReview 提交评审：创建评审记录 + 指定评审人 + 将需求状态置为 in_review。
func (s *ReviewService) SubmitReview(ctx context.Context, in SubmitReviewInput) (*Review, error) {
	tc, err := detectWorkitemType(ctx, s.db, in.IssueID)
	if err != nil {
		return nil, err
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var reviewID int64
	name := in.Name
	if name == "" {
		name = "需求评审"
	}
	err = tx.QueryRow(ctx, `
		INSERT INTO reviews (workspace_id, project_id, name, review_type, entity_type, entity_id, status, created_by)
		VALUES ($1,$2,$3,'requirement_review',$4,$5,'active',$6)
		RETURNING id`,
		in.WorkspaceID, in.ProjectID, name, string(tc), in.IssueID, in.UserID).Scan(&reviewID)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}

	for _, uid := range in.Reviewers {
		if _, err := tx.Exec(ctx, `
			INSERT INTO review_assignments (workspace_id, project_id, review_id, assignee_id)
			VALUES ($1,$2,$3,$4) ON CONFLICT (review_id, assignee_id) DO NOTHING`,
			in.WorkspaceID, in.ProjectID, reviewID, uid); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
	}

	// 仅需求支持 review_status
	if tc == TypeRequirement {
		if _, err := tx.Exec(ctx, `
			UPDATE requirement SET review_status = $1, updated_at = now()
			WHERE id = $2 AND workspace_id = $3`,
			ReviewInReview, in.IssueID, in.WorkspaceID); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	return s.GetByID(ctx, reviewID)
}

// DecideReviewInput 评审决定入参。
type DecideReviewInput struct {
	WorkspaceID int64
	IssueID     int64
	UserID      int64
	Decision    string // approved | rejected
}

// DecideReview 评审人作出决定：更新 review_status + 评审记录 completed_date。
func (s *ReviewService) DecideReview(ctx context.Context, in DecideReviewInput) error {
	if in.Decision != ReviewApproved && in.Decision != ReviewRejected {
		return errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "decision", Reason: "decision 须为 approved 或 rejected"})
	}
	tc, err := detectWorkitemType(ctx, s.db, in.IssueID)
	if err != nil {
		return err
	}
	if tc != TypeRequirement {
		return errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "issue_id", Reason: "仅需求支持评审"})
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	tag, err := tx.Exec(ctx, `
		UPDATE requirement SET review_status = $1, updated_at = now()
		WHERE id = $2 AND workspace_id = $3 AND review_status = $4`,
		in.Decision, in.IssueID, in.WorkspaceID, ReviewInReview)
	if err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	if tag.RowsAffected() == 0 {
		return errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "review_status", Reason: "当前状态不可评审（须处于评审中）"})
	}

	if _, err := tx.Exec(ctx, `
		UPDATE reviews SET completed_date = CURRENT_DATE, updated_at = now()
		WHERE entity_id = $1 AND entity_type = 'requirement' AND completed_date IS NULL`,
		in.IssueID); err != nil {
		return errs.ErrInternal.Wrap(err)
	}

	return tx.Commit(ctx)
}

// ListReviews 查询工作项的评审记录。
func (s *ReviewService) ListReviews(ctx context.Context, issueID int64) ([]Review, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, workspace_id, project_id, name, review_type, entity_type, entity_id,
		       status, description, due_date, created_date, completed_date, created_by, created_at
		FROM reviews
		WHERE entity_id = $1 AND deleted = false
		ORDER BY created_at DESC`, issueID)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()

	var reviews []Review
	for rows.Next() {
		var r Review
		if err := rows.Scan(&r.ID, &r.WorkspaceID, &r.ProjectID, &r.Name, &r.ReviewType,
			&r.EntityType, &r.EntityID, &r.Status, &r.Description, &r.DueDate,
			&r.CreatedDate, &r.CompletedDate, &r.CreatedBy, &r.CreatedAt); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
		r.Reviewers, _ = loadIntArray(ctx, s.db, `SELECT assignee_id FROM review_assignments WHERE review_id = $1 AND deleted = false`, r.ID)
		reviews = append(reviews, r)
	}
	return reviews, rows.Err()
}

// GetByID 查询单条评审记录。
func (s *ReviewService) GetByID(ctx context.Context, reviewID int64) (*Review, error) {
	var r Review
	err := s.db.QueryRow(ctx, `
		SELECT id, workspace_id, project_id, name, review_type, entity_type, entity_id,
		       status, description, due_date, created_date, completed_date, created_by, created_at
		FROM reviews WHERE id = $1 AND deleted = false`, reviewID).Scan(
		&r.ID, &r.WorkspaceID, &r.ProjectID, &r.Name, &r.ReviewType,
		&r.EntityType, &r.EntityID, &r.Status, &r.Description, &r.DueDate,
		&r.CreatedDate, &r.CompletedDate, &r.CreatedBy, &r.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrNotFound
		}
		return nil, errs.ErrInternal.Wrap(err)
	}
	return &r, nil
}
