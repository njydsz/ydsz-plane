// Package issue — 工作项关联。
//
// 支持关联 (Relation)：双向、无序，6 种语义（duplicate/relates_to/blocked_by/start_before/finish_before/implemented_by）。
// 数据存储于跨类型通用关联表 biz_entity_relations（source_type/source_id/target_type/target_id 可跨 task/requirement/defect）。
package issue

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// RelationService 管理工作项之间的关联（无向）。
//
// 关联关系特性：同 source+target+type 重复创建视为幂等（唯一约束兜底）。
type RelationService struct {
	db *pgxpool.Pool
}

// NewRelationService 创建关联服务。
func NewRelationService(db *pgxpool.Pool) *RelationService {
	return &RelationService{db: db}
}

// --- 关联关系 ---

// CreateRelationInput 创建关联关系的入参。
type CreateRelationInput struct {
	WorkspaceID   int64  // 工作空间 ID（RLS 隔离）。
	ProjectID     int64  // 项目 ID。
	SourceIssueID int64  // 源工作项 ID。
	TargetIssueID int64  // 目标工作项 ID。
	RelationType  string // 关联类型：duplicate | relates_to | blocked_by | start_before | finish_before | implemented_by
	CreatedBy     int64  // 创建人 user_id。
}

// validRelationTypes 有效的关联类型集合。
var validRelationTypes = map[string]bool{
	"duplicate": true, "relates_to": true, "blocked_by": true,
	"start_before": true, "finish_before": true, "implemented_by": true,
}

// CreateRelation 建立两个工作项之间的关联。
// 工作项类型通过三主表推断（雪花 ID 全局唯一，跨类型关联由 biz_entity_relations 承载）。
func (s *RelationService) CreateRelation(ctx context.Context, in CreateRelationInput) (*IssueRelation, error) {
	if in.SourceIssueID == in.TargetIssueID {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{
			Field: "target_issue_id", Reason: "不能关联自己",
		})
	}
	if !validRelationTypes[in.RelationType] {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{
			Field: "relation_type", Reason: "无效的关联类型",
		})
	}

	srcType, err := detectWorkitemType(ctx, s.db, in.SourceIssueID)
	if err != nil {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{
			Field: "source_issue_id", Reason: "源工作项不存在",
		})
	}
	tgtType, err := detectWorkitemType(ctx, s.db, in.TargetIssueID)
	if err != nil {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{
			Field: "target_issue_id", Reason: "目标工作项不存在",
		})
	}

	var rel IssueRelation
	err = s.db.QueryRow(ctx, `
		INSERT INTO biz_entity_relations (workspace_id, project_id, source_type, source_id, target_type, target_id, relation_type, created_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		ON CONFLICT (source_type, source_id, target_type, target_id, relation_type) DO NOTHING
		RETURNING id, created_at`,
		in.WorkspaceID, in.ProjectID, srcType, in.SourceIssueID,
		tgtType, in.TargetIssueID, in.RelationType, in.CreatedBy).Scan(&rel.ID, &rel.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			// 幂等：已存在则返回既有记录
			rel.WorkspaceID = in.WorkspaceID
			rel.ProjectID = in.ProjectID
			rel.SourceIssueID = in.SourceIssueID
			rel.TargetIssueID = in.TargetIssueID
			rel.RelationType = in.RelationType
			rel.CreatedBy = in.CreatedBy
			return &rel, nil
		}
		return nil, errs.ErrInternal.Wrap(err)
	}

	rel.WorkspaceID = in.WorkspaceID
	rel.ProjectID = in.ProjectID
	rel.SourceIssueID = in.SourceIssueID
	rel.TargetIssueID = in.TargetIssueID
	rel.RelationType = in.RelationType
	rel.CreatedBy = in.CreatedBy

	return &rel, nil
}

// DeleteRelation 删除关联。
func (s *RelationService) DeleteRelation(ctx context.Context, wsID, relationID int64) error {
	tag, err := s.db.Exec(ctx,
		`DELETE FROM biz_entity_relations WHERE id = $1 AND workspace_id = $2`, relationID, wsID)
	if err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	if tag.RowsAffected() == 0 {
		return errs.ErrNotFound
	}
	return nil
}

// ListRelations 列出工作项的所有关联。
func (s *RelationService) ListRelations(ctx context.Context, wsID, issueID int64) ([]IssueRelation, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, workspace_id, project_id, source_id, target_id, relation_type, created_by, created_at
		FROM biz_entity_relations
		WHERE workspace_id = $1 AND (source_id = $2 OR target_id = $2) AND deleted = false
		ORDER BY created_at DESC`, wsID, issueID)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()

	var rels []IssueRelation
	for rows.Next() {
		var r IssueRelation
		var sourceID, targetID int64
		if err := rows.Scan(&r.ID, &r.WorkspaceID, &r.ProjectID, &sourceID, &targetID,
			&r.RelationType, &r.CreatedBy, &r.CreatedAt); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
		r.SourceIssueID = sourceID
		r.TargetIssueID = targetID
		rels = append(rels, r)
	}
	return rels, rows.Err()
}
