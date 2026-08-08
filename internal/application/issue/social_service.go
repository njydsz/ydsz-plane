// Package issue — 工作项社交反馈（表情 Reaction + 投票 Vote）。
//
// 参考 Plane / Linear 的轻量协作反馈：
//   - Reaction：同人同工作项同表情唯一（UNIQUE），重复添加幂等返回已存在；删除不存在返回 ErrNotFound。
//   - Vote：同人同工作项一票（UNIQUE），支持改票（1/-1）与撤销（vote=0 删除）。
package issue

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// SocialService 管理工作项的 Reaction 与 Vote。
type SocialService struct {
	db *pgxpool.Pool
}

// NewSocialService 创建社交服务。
func NewSocialService(db *pgxpool.Pool) *SocialService {
	return &SocialService{db: db}
}

// --- Reaction ---

// AddReaction 添加表情反应（幂等：已存在则返回现有记录）。
func (s *SocialService) AddReaction(ctx context.Context, wsID, projectID, issueID, userID int64, reactionType string) (*IssueReaction, bool, error) {
	if reactionType == "" {
		return nil, false, errs.ErrValidation.WithDetails(errs.FieldDetail{
			Field: "reaction_type", Reason: "表情不能为空",
		})
	}
	if len([]rune(reactionType)) > 16 {
		return nil, false, errs.ErrValidation.WithDetails(errs.FieldDetail{
			Field: "reaction_type", Reason: "表情过长（最多 16 字符）",
		})
	}

	var r IssueReaction
	err := s.db.QueryRow(ctx, `
		INSERT INTO issue_reactions (workspace_id, project_id, issue_id, user_id, reaction_type)
		VALUES ($1,$2,$3,$4,$5)
		ON CONFLICT (issue_id, user_id, reaction_type) DO NOTHING
		RETURNING id, created_at`,
		wsID, projectID, issueID, userID, reactionType).Scan(&r.ID, &r.CreatedAt)

	created := true
	if err != nil {
		// ON CONFLICT DO NOTHING 冲突时无返回行
		if errors.Is(err, pgx.ErrNoRows) {
			created = false
			// 查出已存在记录的 id/created_at
			err = s.db.QueryRow(ctx, `
				SELECT id, created_at FROM issue_reactions
				WHERE issue_id = $1 AND user_id = $2 AND reaction_type = $3`,
				issueID, userID, reactionType).Scan(&r.ID, &r.CreatedAt)
		}
		if err != nil {
			return nil, false, errs.ErrInternal.Wrap(err)
		}
	}

	r.WorkspaceID = wsID
	r.ProjectID = projectID
	r.IssueID = issueID
	r.UserID = userID
	r.ReactionType = reactionType
	return &r, created, nil
}

// RemoveReaction 删除表情反应。
func (s *SocialService) RemoveReaction(ctx context.Context, wsID, issueID, userID int64, reactionType string) error {
	tag, err := s.db.Exec(ctx, `
		DELETE FROM issue_reactions
		WHERE issue_id = $1 AND user_id = $2 AND reaction_type = $3 AND workspace_id = $4`,
		issueID, userID, reactionType, wsID)
	if err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	if tag.RowsAffected() == 0 {
		return errs.ErrNotFound
	}
	return nil
}

// ListReactions 列出工作项的全部反应（按表情聚合计数 + 当前用户是否已反应）。
func (s *SocialService) ListReactions(ctx context.Context, wsID, issueID, userID int64) ([]ReactionSummary, error) {
	rows, err := s.db.Query(ctx, `
		SELECT reaction_type,
		       COUNT(*)::int AS cnt,
		       bool_or(user_id = $3) AS reacted
		FROM issue_reactions
		WHERE workspace_id = $1 AND issue_id = $2
		GROUP BY reaction_type
		ORDER BY cnt DESC, reaction_type ASC`, wsID, issueID, userID)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()

	summaries := make([]ReactionSummary, 0, 4)
	for rows.Next() {
		var s2 ReactionSummary
		if err := rows.Scan(&s2.ReactionType, &s2.Count, &s2.Reacted); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
		summaries = append(summaries, s2)
	}
	return summaries, rows.Err()
}

// --- Vote ---

// VoteIssue 投票/改票。vote=1 赞成，-1 反对；重复相同投票幂等返回当前状态。
func (s *SocialService) VoteIssue(ctx context.Context, wsID, projectID, issueID, userID int64, vote int) (*IssueVote, error) {
	if vote != 1 && vote != -1 {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{
			Field: "vote", Reason: "投票值只能为 1（赞成）或 -1（反对）",
		})
	}

	var v IssueVote
	err := s.db.QueryRow(ctx, `
		INSERT INTO issue_votes (workspace_id, project_id, issue_id, user_id, vote)
		VALUES ($1,$2,$3,$4,$5)
		ON CONFLICT (issue_id, user_id) DO UPDATE SET vote = EXCLUDED.vote, updated_at = now()
		RETURNING id, vote, created_at, updated_at`,
		wsID, projectID, issueID, userID, vote).Scan(&v.ID, &v.Vote, &v.CreatedAt, &v.UpdatedAt)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}

	v.WorkspaceID = wsID
	v.ProjectID = projectID
	v.IssueID = issueID
	v.UserID = userID
	return &v, nil
}

// RemoveVote 撤销投票。
func (s *SocialService) RemoveVote(ctx context.Context, wsID, issueID, userID int64) error {
	tag, err := s.db.Exec(ctx, `
		DELETE FROM issue_votes
		WHERE issue_id = $1 AND user_id = $2 AND workspace_id = $3`,
		issueID, userID, wsID)
	if err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	if tag.RowsAffected() == 0 {
		return errs.ErrNotFound
	}
	return nil
}

// VoteSummary 返回工作项投票聚合（赞成/反对/总分/当前用户投票）。
func (s *SocialService) VoteSummary(ctx context.Context, wsID, issueID, userID int64) (*VoteSummary, error) {
	var sum VoteSummary
	err := s.db.QueryRow(ctx, `
		SELECT
			COALESCE(SUM(CASE WHEN vote = 1 THEN 1 ELSE 0 END), 0)::int AS up,
			COALESCE(SUM(CASE WHEN vote = -1 THEN 1 ELSE 0 END), 0)::int AS down,
			COALESCE(SUM(vote), 0)::int AS score
		FROM issue_votes
		WHERE workspace_id = $1 AND issue_id = $2`, wsID, issueID).Scan(&sum.Upvotes, &sum.Downvotes, &sum.Score)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}

	// 当前用户投票
	var voted int
	err = s.db.QueryRow(ctx, `
		SELECT vote FROM issue_votes WHERE issue_id = $1 AND user_id = $2 AND workspace_id = $3`,
		issueID, userID, wsID).Scan(&voted)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			sum.Voted = nil
		} else {
			return nil, errs.ErrInternal.Wrap(err)
		}
	} else {
		sum.Voted = &voted
	}
	return &sum, nil
}
