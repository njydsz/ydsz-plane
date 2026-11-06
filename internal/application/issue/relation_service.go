// Package issue — 工作项关联。
//
// 支持关联 (Relation)：双向、无序，6 种语义（duplicate/relates_to/blocked_by/start_before/finish_before/implemented_by）。
// 数据存储于跨类型通用关联表 biz_entity_relations（source_type/source_id/target_type/target_id 可跨 task/requirement/defect）。
package issue

import (
	"context"
	"strings"

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

// ===========================================================================
// Plane 风格关系语义增强（blocked / blocking / duplicate）
//
// 在既有 6 种语义之上补充：
//   - blocked / blocking：有向阻塞关系（互为镜像，成对维护）；
//   - 循环阻塞检测（A 阻塞 B 且 B 阻塞 A 时拒绝）；
//   - duplicate 对称性（A duplicate B 时自动补建反向记录）。
// ===========================================================================

// Plane 风格关系类型。
const (
	// RelTypeBlocked 被阻塞。
	RelTypeBlocked = "blocked"
	// RelTypeBlocking 阻塞他人。
	RelTypeBlocking = "blocking"
	// RelTypeRelates 普通关联（Plane 命名）。
	RelTypeRelates = "relates"
	// RelTypeDuplicate 重复（对齐既有命名）。
	RelTypeDuplicate = "duplicate"
	// RelTypeRelatesTo 关联（对齐既有命名）。
	RelTypeRelatesTo = "relates_to"
)

// IsValidPlaneRelationType 校验 Plane 风格关系类型。
func IsValidPlaneRelationType(t string) bool {
	switch t {
	case RelTypeBlocked, RelTypeBlocking, RelTypeRelates,
		RelTypeDuplicate, RelTypeRelatesTo:
		return true
	}
	return false
}

// mirrorRelationType 返回关系的反向类型。
func mirrorRelationType(t string) string {
	switch t {
	case RelTypeBlocked:
		return RelTypeBlocking
	case RelTypeBlocking:
		return RelTypeBlocked
	case RelTypeDuplicate, RelTypeRelatesTo, RelTypeRelates:
		return t
	}
	return t
}

// CreateRelationWithSemantics 建立 Plane 风格关系。
//
// 相比 CreateRelation，额外提供：
//   - blocked/blocking 的镜像自动补建；
//   - 循环阻塞检测（避免 A blocked B 且 B blocked A）；
//   - duplicate 的对称记录。
//
// 返回主关系记录。
func (s *RelationService) CreateRelationWithSemantics(ctx context.Context, in CreateRelationInput) (*IssueRelation, error) {
	if in.SourceIssueID == in.TargetIssueID {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{
			Field: "target_issue_id", Reason: "不能关联自己",
		})
	}
	if !IsValidPlaneRelationType(in.RelationType) {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{
			Field: "relation_type", Reason: "无效的关联类型",
		})
	}

	if _, err := detectWorkitemType(ctx, s.db, in.SourceIssueID); err != nil {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{
			Field: "source_issue_id", Reason: "源工作项不存在",
		})
	}
	if _, err := detectWorkitemType(ctx, s.db, in.TargetIssueID); err != nil {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{
			Field: "target_issue_id", Reason: "目标工作项不存在",
		})
	}

	// 循环阻塞检测：若 target 已经 blocked source，则 source blocked target 会成环
	if in.RelationType == RelTypeBlocked {
		cycle, cerr := s.hasBlockingPath(ctx, in.TargetIssueID, in.SourceIssueID)
		if cerr != nil {
			return nil, cerr
		}
		if cycle {
			return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{
				Field: "relation_type", Reason: "存在循环阻塞，无法建立该关系",
			})
		}
	}

	rel, err := s.CreateRelation(ctx, in)
	if err != nil {
		return nil, err
	}

	// 补建镜像关系
	mirror := mirrorRelationType(in.RelationType)
	if mirror != in.RelationType {
		_, _ = s.CreateRelation(ctx, CreateRelationInput{
			WorkspaceID:   in.WorkspaceID,
			ProjectID:     in.ProjectID,
			SourceIssueID: in.TargetIssueID,
			TargetIssueID: in.SourceIssueID,
			RelationType:  mirror,
			CreatedBy:     in.CreatedBy,
		})
	}
	return rel, nil
}

// hasBlockingPath 判断从 from 出发沿 blocked 链能否到达 to（用于循环检测）。
func (s *RelationService) hasBlockingPath(ctx context.Context, from, to int64) (bool, error) {
	rows, err := s.db.Query(ctx, `
		WITH RECURSIVE chain AS (
			SELECT target_id AS node, 0 AS depth
			FROM biz_entity_relations
			WHERE source_id = $1 AND relation_type = 'blocked' AND NOT deleted
			UNION ALL
			SELECT r.target_id, c.depth + 1
			FROM biz_entity_relations r
			JOIN chain c ON r.source_id = c.node
			WHERE r.relation_type = 'blocked' AND NOT r.deleted AND c.depth < 20
		)
		SELECT 1 FROM chain WHERE node = $2 LIMIT 1`, from, to)
	if err != nil {
		return false, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()

	found := rows.Next()
	for rows.Next() {
	}
	return found, rows.Err()
}

// GetBlockingItems 返回阻塞指定工作项的所有工作项 ID。
func (s *RelationService) GetBlockingItems(ctx context.Context, wsID, issueID int64) ([]int64, error) {
	rows, err := s.db.Query(ctx, `
		SELECT source_id FROM biz_entity_relations
		WHERE target_id = $1 AND relation_type = 'blocked'
		  AND workspace_id = $2 AND NOT deleted
		ORDER BY created_at ASC`, issueID, wsID)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()

	out := []int64{}
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
		out = append(out, id)
	}
	return out, rows.Err()
}

// GetBlockedItems 返回被指定工作项阻塞的所有工作项 ID。
func (s *RelationService) GetBlockedItems(ctx context.Context, wsID, issueID int64) ([]int64, error) {
	rows, err := s.db.Query(ctx, `
		SELECT target_id FROM biz_entity_relations
		WHERE source_id = $1 AND relation_type = 'blocked'
		  AND workspace_id = $2 AND NOT deleted
		ORDER BY created_at ASC`, issueID, wsID)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()

	out := []int64{}
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
		out = append(out, id)
	}
	return out, rows.Err()
}

// IsBlocked 判断工作项是否被阻塞。
func (s *RelationService) IsBlocked(ctx context.Context, wsID, issueID int64) (bool, error) {
	var exists bool
	err := s.db.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM biz_entity_relations
			WHERE target_id = $1 AND relation_type = 'blocked'
			  AND workspace_id = $2 AND NOT deleted
		)`, issueID, wsID).Scan(&exists)
	if err != nil {
		return false, errs.ErrInternal.Wrap(err)
	}
	return exists, nil
}

// FindDuplicateCandidates 查找可能重复的工作项（同项目内标题相同或包含）。
func (s *RelationService) FindDuplicateCandidates(ctx context.Context, wsID, projectID int64, title string) ([]DuplicateCandidate, error) {
	norm := strings.ToLower(strings.TrimSpace(title))
	if norm == "" {
		return nil, nil
	}

	rows, err := s.db.Query(ctx, `
		SELECT id, 'requirement' AS type, title FROM requirement
		WHERE workspace_id = $1 AND project_id = $2 AND deleted = false
		  AND (LOWER(title) = $3 OR LOWER(title) LIKE '%' || $3 || '%')
		UNION ALL
		SELECT id, 'task', title FROM task
		WHERE workspace_id = $1 AND project_id = $2 AND deleted = false
		  AND (LOWER(title) = $3 OR LOWER(title) LIKE '%' || $3 || '%')
		UNION ALL
		SELECT id, 'defect', title FROM defect
		WHERE workspace_id = $1 AND project_id = $2 AND deleted = false
		  AND (LOWER(title) = $3 OR LOWER(title) LIKE '%' || $3 || '%')
		ORDER BY title ASC LIMIT 20`, wsID, projectID, norm)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()

	out := []DuplicateCandidate{}
	for rows.Next() {
		var c DuplicateCandidate
		if err := rows.Scan(&c.ID, &c.Type, &c.Title); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// DuplicateCandidate 重复候选。
type DuplicateCandidate struct {
	ID    int64  `json:"id"`
	Type  string `json:"type"`
	Title string `json:"title"`
}
