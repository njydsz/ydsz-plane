// Package issue — 工作项关联 & 依赖。
//
// 支持两种关系：
//   - 关联 (Relation)：双向、无序，6 种语义（duplicate/relates_to/blocked_by/start_before/finish_before/implemented_by）。
//   - 依赖 (Dependency)：有向、4 种类型（FS/SS/FF/SF），通过 BFS 防止循环依赖。
package issue

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// RelationService 管理工作项之间的关联（无向）和依赖（有向）。
//
// 关联关系特性：ON CONFLICT 幂等（同 source+target+type 重复创建视为更新）。
// 依赖关系特性：创建前 BFS 防环；删除不影响已建成的 DAG 其它节点。
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
	validTypes    []string
}

// validRelationTypes 有效的关联类型集合。
var validRelationTypes = map[string]bool{
	"duplicate": true, "relates_to": true, "blocked_by": true,
	"start_before": true, "finish_before": true, "implemented_by": true,
}

// CreateRelation 建立两个工作项之间的关联。
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

	var rel IssueRelation
	err := s.db.QueryRow(ctx, `
		INSERT INTO issue_relations (workspace_id, project_id, source_issue_id, target_issue_id, relation_type, created_by)
		VALUES ($1,$2,$3,$4,$5,$6)
		ON CONFLICT (source_issue_id, target_issue_id, relation_type) DO UPDATE SET source_issue_id = EXCLUDED.source_issue_id
		RETURNING id, created_at`,
		in.WorkspaceID, in.ProjectID, in.SourceIssueID, in.TargetIssueID,
		in.RelationType, in.CreatedBy).Scan(&rel.ID, &rel.CreatedAt)

	if err != nil {
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
		`DELETE FROM issue_relations WHERE id = $1 AND workspace_id = $2`, relationID, wsID)
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
		SELECT id, workspace_id, project_id, source_issue_id, target_issue_id, relation_type, created_by, created_at
		FROM issue_relations
		WHERE workspace_id = $1 AND (source_issue_id = $2 OR target_issue_id = $2)
		ORDER BY created_at DESC`, wsID, issueID)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()

	var rels []IssueRelation
	for rows.Next() {
		var r IssueRelation
		if err := rows.Scan(&r.ID, &r.WorkspaceID, &r.ProjectID, &r.SourceIssueID, &r.TargetIssueID,
			&r.RelationType, &r.CreatedBy, &r.CreatedAt); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
		rels = append(rels, r)
	}
	return rels, rows.Err()
}

// --- 依赖 ---

// CreateDependencyInput 建立依赖关系的入参。
type CreateDependencyInput struct {
	WorkspaceID    int64  // 工作空间 ID（RLS 隔离）。
	ProjectID      int64  // 项目 ID。
	PredecessorID  int64  // 前置工作项 ID。
	SuccessorID    int64  // 后续工作项 ID。
	DependencyType string // 依赖类型：FS (完成-开始) | SS (开始-开始) | FF (完成-完成) | SF (开始-完成)
	LagDays        int    // 延后天数；语义由 DependencyType 决定（见 DurationModeByDependencyType）。
	CreatedBy      int64  // 创建人 user_id。
}

// validDependencyTypes 有效的依赖类型集合。
var validDependencyTypes = map[string]bool{
	"FS": true, "SS": true, "FF": true, "SF": true,
}

// CreateDependency 建立依赖（含 DFS 防环）。
func (s *RelationService) CreateDependency(ctx context.Context, in CreateDependencyInput) (*IssueDependency, error) {
	if in.PredecessorID == in.SuccessorID {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{
			Field: "successor_id", Reason: "前置和后续不能是同一工作项",
		})
	}
	if !validDependencyTypes[in.DependencyType] {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{
			Field: "dependency_type", Reason: "无效的依赖类型（FS/SS/FF/SF）",
		})
	}

	// DFS 环检测：successor 不能再是 predecessor 的前置（形成环）。
	hasCycle, err := s.hasCycle(ctx, in.PredecessorID, in.SuccessorID)
	if err != nil {
		return nil, err
	}
	if hasCycle {
		return nil, errs.ErrCircularParent.WithDetails(errs.FieldDetail{
			Field: "successor_id", Reason: "添加该依赖将形成循环依赖",
		})
	}

	var dep IssueDependency
	err = s.db.QueryRow(ctx, `
		INSERT INTO issue_dependencies (workspace_id, project_id, predecessor_id, successor_id, dependency_type, lag_days, created_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		ON CONFLICT (predecessor_id, successor_id, dependency_type) DO UPDATE SET predecessor_id = EXCLUDED.predecessor_id
		RETURNING id, created_at`,
		in.WorkspaceID, in.ProjectID, in.PredecessorID, in.SuccessorID,
		in.DependencyType, in.LagDays, in.CreatedBy).Scan(&dep.ID, &dep.CreatedAt)

	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}

	dep.WorkspaceID = in.WorkspaceID
	dep.ProjectID = in.ProjectID
	dep.PredecessorID = in.PredecessorID
	dep.SuccessorID = in.SuccessorID
	dep.DependencyType = in.DependencyType
	dep.LagDays = in.LagDays
	dep.CreatedBy = in.CreatedBy

	return &dep, nil
}

// hasCycle 检测在 predecessorID → successorID 之间添加依赖是否会形成环。
//
// 算法：从 successor 出发，沿 predecessor_id 边做 BFS；若可到达 predecessor，则添加新边后成环。
// 时间复杂度 O(V+E)，V/E 为可达子图规模；工程上通常很小（几十个节点）。
func (s *RelationService) hasCycle(ctx context.Context, predecessorID, successorID int64) (bool, error) {
	// BFS/DFS：从 successor 开始，沿 predecessor_id 边搜索（即找 successor 的所有前置）。
	// 如果找到 predecessorID，则添加 predecessorID -> successorID 会形成环。
	visited := make(map[int64]bool)
	queue := []int64{successorID}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if visited[current] {
			continue
		}
		visited[current] = true

		if current == predecessorID {
			return true, nil
		}

		// 找 current 的所有直接前置
		rows, err := s.db.Query(ctx,
			`SELECT predecessor_id FROM issue_dependencies WHERE successor_id = $1`, current)
		if err != nil {
			return false, errs.ErrInternal.Wrap(err)
		}
		for rows.Next() {
			var pred int64
			if err := rows.Scan(&pred); err != nil {
				rows.Close()
				return false, errs.ErrInternal.Wrap(err)
			}
			if !visited[pred] {
				queue = append(queue, pred)
			}
		}
		rows.Close()
	}

	return false, nil
}

// DeleteDependency 删除依赖。
func (s *RelationService) DeleteDependency(ctx context.Context, wsID, depID int64) error {
	tag, err := s.db.Exec(ctx,
		`DELETE FROM issue_dependencies WHERE id = $1 AND workspace_id = $2`, depID, wsID)
	if err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	if tag.RowsAffected() == 0 {
		return errs.ErrNotFound
	}
	return nil
}

// ListDependencies 列出工作项的所有依赖（作为前置 + 作为后续）。
func (s *RelationService) ListDependencies(ctx context.Context, wsID, issueID int64) (predecessors []IssueDependency, successors []IssueDependency, err error) {
	// 作为前置（我的前置是谁）
	predRows, err := s.db.Query(ctx, `
		SELECT id, workspace_id, project_id, predecessor_id, successor_id, dependency_type, lag_days, created_by, created_at
		FROM issue_dependencies
		WHERE workspace_id = $1 AND successor_id = $2
		ORDER BY created_at DESC`, wsID, issueID)
	if err != nil {
		return nil, nil, errs.ErrInternal.Wrap(err)
	}
	defer predRows.Close()

	for predRows.Next() {
		var d IssueDependency
		if err := predRows.Scan(&d.ID, &d.WorkspaceID, &d.ProjectID, &d.PredecessorID, &d.SuccessorID,
			&d.DependencyType, &d.LagDays, &d.CreatedBy, &d.CreatedAt); err != nil {
			return nil, nil, errs.ErrInternal.Wrap(err)
		}
		predecessors = append(predecessors, d)
	}

	// 作为后续（我的后续是谁）
	succRows, err := s.db.Query(ctx, `
		SELECT id, workspace_id, project_id, predecessor_id, successor_id, dependency_type, lag_days, created_by, created_at
		FROM issue_dependencies
		WHERE workspace_id = $1 AND predecessor_id = $2
		ORDER BY created_at DESC`, wsID, issueID)
	if err != nil {
		return nil, nil, errs.ErrInternal.Wrap(err)
	}
	defer succRows.Close()

	for succRows.Next() {
		var d IssueDependency
		if err := succRows.Scan(&d.ID, &d.WorkspaceID, &d.ProjectID, &d.PredecessorID, &d.SuccessorID,
			&d.DependencyType, &d.LagDays, &d.CreatedBy, &d.CreatedAt); err != nil {
			return nil, nil, errs.ErrInternal.Wrap(err)
		}
		successors = append(successors, d)
	}

	return predecessors, successors, nil
}

// --- Duration helpers ---

// DurationModeByDependencyType 返回每种依赖类型对应的 lag 取值的含义说明。
// FS(完成-开始): "前置完成后 X 天，后续才能开始"
// SS(开始-开始): "前置开始后 X 天，后续才能开始"
// FF(完成-完成): "前置完成后 X 天，后续才能完成"
// SF(开始-完成): "前置开始后 X 天，后续才能完成"
var DurationModeByDependencyType = map[string]string{
	"FS": "前置完成后 X 天，后续才能开始",
	"SS": "前置开始后 X 天，后续才能开始",
	"FF": "前置完成后 X 天，后续才能完成",
	"SF": "前置开始后 X 天，后续才能完成",
}

