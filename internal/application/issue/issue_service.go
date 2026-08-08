// Package issue — Issue 应用服务（CRUD + WBS + 发号器 + 状态流转）。
//
// 参考: Plane / Linear Issue domain service。
package issue

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// Service 提供 Issue 领域应用服务。
type Service struct {
	db       *pgxpool.Pool
	stateSvc *StateService
}

// NewService 创建 Issue 服务。
func NewService(db *pgxpool.Pool) *Service {
	return &Service{db: db, stateSvc: NewStateService(db)}
}

// --- Input Types ---

// CreateIssueInput 创建工作项的入参。
type CreateIssueInput struct {
	WorkspaceID      int64
	ProjectID        int64
	TypeCode         IssueTypeCode
	Name             string
	DescriptionHTML  string
	StateID          int64 // 可选；未指定则默认
	Priority         IssuePriority
	ParentID         *int64
	Severity         *int
	FoundPhase       *string
	ReproduceSteps   map[string]any
	Category         *string
	Source           *string
	Assignees        []int64
	Labels           []int64
	Modules          []int64
	Point            *int
	ExternalID       *string
	StartDate        *time.Time
	TargetDate       *time.Time
	IsDraft          bool
	CreatedBy        int64
	FoundVersionID   *int64
	FixVersionID     *int64
	ReleaseVersionID *int64
}

// UpdateIssueInput 更新工作项的入参。
// 注意：state_id不允许通过更新接口修改，所有状态变更必须走专用的transition端点
// 保证状态流转必须经过规则校验、required_fields校验
type UpdateIssueInput struct {
	Name              *string
	DescriptionHTML   *string
	Priority          *IssuePriority
	TypeCode          *IssueTypeCode
	ParentID          *int64 // nil=不更新, 设置值=更新
	Severity          *int
	FoundPhase        *string
	RootCauseCategory *string
	Category          *string
	Assignees         []int64
	Labels            []int64
	Modules           []int64
	Source            *string
	Point             *int
	TargetDate        *time.Time
	Progress          *int
	Version           int
	FoundVersionID    *int64
	FixVersionID      *int64
	ReleaseVersionID  *int64
	VerifierID       *int64         `json:"verifier_id,omitempty"`
	ReproduceSteps   json.RawMessage `json:"reproduce_steps,omitempty"`
}

// ListIssuesOptions 工作项列表查询选项。
type ListIssuesOptions struct {
	WorkspaceID      int64
	ProjectID        int64
	StateID          *int64
	Group            *StateGroup
	TypeCode         *IssueTypeCode
	Priority         *IssuePriority
	ParentID         *int64
	Search           string
	SortBy           string
	SortDesc         bool
	Limit            int
	Offset           int
	FoundVersionID   *int64
	FixVersionID     *int64
	ReleaseVersionID *int64
	AssigneeID       *int64
	LabelID          *int64
	ModuleID         *int64
	SprintID         *int64
	StartDateFrom    *string // ISO date string
	TargetDateTo     *string // ISO date string
	SeverityFrom     *int    // 缺陷最低严重级别
}

// ReorderInput 看板拖拽排序输入。
type ReorderInput struct {
	PrevSortOrder *float64 // 前一个工作项的 sort_order，nil 表示移到第一个
	NextSortOrder *float64 // 后一个工作项的 sort_order，nil 表示移到最后一个
	Version       *int     // 客户端获取的版本号，用于乐观锁防冲突（nil 表示跳过版本校验）
}

// BatchUpdateInput 批量操作输入。
type BatchUpdateInput struct {
	IssueIDs   []int64
	ToStateID  *int64
	AssigneeID *int64
	Priority   *string
	Delete     bool
}

// BatchResult 批量操作结果。
type BatchResult struct {
	Succeeded int
	Failed    int
}

// --- CRUD ---

// Create 创建工作项。
func (s *Service) Create(ctx context.Context, in CreateIssueInput) (*Issue, error) {
	if in.Priority == "" {
		in.Priority = PriorityNone
	}
	if err := validateCreateInput(in); err != nil {
		return nil, err
	}

	// WBS 规则校验
	if in.ParentID != nil {
		if err := s.validateWBSDepth(ctx, *in.ParentID); err != nil {
			return nil, err
		}
	}
	if err := validateTypeFields(in.TypeCode, in.Severity, in.FoundPhase); err != nil {
		return nil, err
	}

	// 状态未指定取默认
	stateID := in.StateID
	if stateID == 0 {
		defaultState, err := s.stateSvc.GetDefaultState(ctx, in.WorkspaceID, in.ProjectID)
		if err != nil {
			return nil, err
		}
		stateID = defaultState.ID
	}

	// 发号
	seqID, err := s.nextSequenceID(ctx, in.ProjectID)
	if err != nil {
		return nil, err
	}

	// 事务写入
	issueID, err := s.insertIssue(ctx, in, stateID, seqID)
	if err != nil {
		return nil, err
	}

	// 加载完整 issue
	return s.GetByID(ctx, in.WorkspaceID, issueID)
}

// GetByID 获取工作项详情。
func (s *Service) GetByID(ctx context.Context, wsID, issueID int64) (*Issue, error) {
	// 先查询工作项类型，判断在新旧表的哪个位置
	var typeCode string
	err := s.db.QueryRow(ctx, `
		SELECT 'task' FROM task WHERE id = $1 AND workspace_id = $2 AND deleted_at IS NULL
		UNION ALL
		SELECT 'requirement' FROM requirement WHERE id = $1 AND workspace_id = $2 AND deleted_at IS NULL
		UNION ALL
		SELECT 'defect' FROM defect WHERE id = $1 AND workspace_id = $2 AND deleted_at IS NULL
		LIMIT 1
	`, issueID, wsID).Scan(&typeCode)
	if err != nil {
		return nil, errs.ErrNotFound
	}

	var iss Issue
	var parentID sql.NullInt64
	var severity sql.NullInt64
	var foundPhase, category sql.NullString
	var point sql.NullInt64
	var startDate, targetDate, completedAt sql.NullTime
	var stateName, stateColor, identifier string
	var stateGroup StateGroup

	var foundVerID, fixVerID, releaseVerID sql.NullInt64
	var sprintID sql.NullInt64
		// 根据类型调用不同表的查询，兼容旧表迁移过渡期
	switch typeCode {
	case "task":
		err = s.db.QueryRow(ctx, `
			SELECT t.id, t.public_id, t.workspace_id, t.project_id, t.sequence_id,
			       'task' as type_code, t.parent_id, t.depth, t.name,
			       t.description_json, t.description_html,
			       t.state_id, s.name, s.color, s."group",
			       t.priority, NULL as severity, NULL as found_phase, t.category, t.point,
			       t.start_date, t.target_date, t.completed_at, t.progress,
			       t.is_draft, t.version, t.created_by, t.created_at, t.updated_at,
			       p.identifier,
			       NULL as found_version_id, NULL as fix_version_id, NULL as release_version_id,
			       t.sprint_id
			FROM task t
			JOIN states s ON s.id = t.state_id
			JOIN projects p ON p.id = t.project_id
			WHERE t.id = $1 AND t.workspace_id = $2 AND t.deleted_at IS NULL`,
			issueID, wsID).Scan(
			&iss.ID, &iss.PublicID, &iss.WorkspaceID, &iss.ProjectID, &iss.SequenceID,
			&iss.TypeCode, &parentID, &iss.Depth, &iss.Name, 
			&iss.DescriptionJSON, &iss.DescriptionHTML,
			&iss.StateID, &stateName, &stateColor, &stateGroup,
			&iss.Priority, &severity, &foundPhase, &category, &point,
			&iss.StartDate, &targetDate, &completedAt, &iss.Progress,
			&iss.IsDraft, &iss.Version, &iss.CreatedBy, &iss.CreatedAt, &iss.UpdatedAt,
			&identifier, &foundVerID, &fixVerID, &releaseVerID, &sprintID)
	case "requirement":
		err = s.db.QueryRow(ctx, `
			SELECT r.id, r.public_id, r.workspace_id, r.project_id, r.sequence_id,
			       'requirement' as type_code, r.parent_id, r.depth, r.name,
			       r.description_json, r.description_html,
			       r.state_id, s.name, s.color, s."group",
			       r.priority, NULL as severity, NULL as found_phase, NULL as category, r.point,
			       r.start_date, r.target_date, r.completed_at, r.progress,
			       r.is_draft, r.version, r.created_by, r.created_at, r.updated_at,
			       p.identifier,
			       NULL as found_version_id, NULL as fix_version_id, NULL as release_version_id,
			       r.sprint_id
			FROM requirement r
			JOIN states s ON s.id = r.state_id
			JOIN projects p ON p.id = r.project_id
			WHERE r.id = $1 AND r.workspace_id = $2 AND r.deleted_at IS NULL`,
			issueID, wsID).Scan(
			&iss.ID, &iss.PublicID, &iss.WorkspaceID, &iss.ProjectID, &iss.SequenceID,
			&iss.TypeCode, &parentID, &iss.Depth, &iss.Name, 
			&iss.DescriptionJSON, &iss.DescriptionHTML,
			&iss.StateID, &stateName, &stateColor, &stateGroup,
			&iss.Priority, &severity, &foundPhase, &category, &point,
			&iss.StartDate, &targetDate, &completedAt, &iss.Progress,
			&iss.IsDraft, &iss.Version, &iss.CreatedBy, &iss.CreatedAt, &iss.UpdatedAt,
			&identifier, &foundVerID, &fixVerID, &releaseVerID, &sprintID)
	case "defect":
		err = s.db.QueryRow(ctx, `
			SELECT d.id, d.public_id, d.workspace_id, d.project_id, d.sequence_id,
			       'defect' as type_code, d.parent_id, d.depth, d.name,
			       d.description_json, d.description_html,
			       d.state_id, s.name, s.color, s."group",
			       d.priority, d.severity, d.found_phase, NULL as category, d.point,
			       d.start_date, d.target_date, d.completed_at, d.progress,
			       d.is_draft, d.version, d.created_by, d.created_at, d.updated_at,
			       p.identifier,
			       d.found_version_id, d.fix_version_id, NULL as release_version_id,
			       d.sprint_id
			FROM defect d
			JOIN states s ON s.id = d.state_id
			JOIN projects p ON p.id = d.project_id
			WHERE d.id = $1 AND d.workspace_id = $2 AND d.deleted_at IS NULL`,
			issueID, wsID).Scan(
			&iss.ID, &iss.PublicID, &iss.WorkspaceID, &iss.ProjectID, &iss.SequenceID,
			&iss.TypeCode, &parentID, &iss.Depth, &iss.Name, 
			&iss.DescriptionJSON, &iss.DescriptionHTML,
			&iss.StateID, &stateName, &stateColor, &stateGroup,
			&iss.Priority, &severity, &foundPhase, &category, &point,
			&iss.StartDate, &targetDate, &completedAt, &iss.Progress,
			&iss.IsDraft, &iss.Version, &iss.CreatedBy, &iss.CreatedAt, &iss.UpdatedAt,
			&identifier, &foundVerID, &fixVerID, &releaseVerID, &sprintID)
	default:
		return nil, errs.ErrNotFound
	}
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrNotFound
		}
		return nil, errs.ErrInternal.Wrap(err)
	}

	iss.Identifier = identifier
	iss.State = &State{ID: iss.StateID, Name: stateName, Color: stateColor, Group: stateGroup}
	if parentID.Valid {
		v := parentID.Int64
		iss.ParentID = &v
	}
	if severity.Valid {
		v := int(severity.Int64)
		iss.Severity = &v
	}
	if foundPhase.Valid {
		iss.FoundPhase = &foundPhase.String
	}
	if category.Valid {
		iss.Category = &category.String
	}
	if point.Valid {
		v := int(point.Int64)
		iss.Point = &v
	}
	if startDate.Valid {
		iss.StartDate = &startDate.Time
	}
	if targetDate.Valid {
		iss.TargetDate = &targetDate.Time
	}
	if completedAt.Valid {
		iss.CompletedAt = &completedAt.Time
	}
	if foundVerID.Valid {
		v := foundVerID.Int64
		iss.FoundVersionID = &v
	}
	if fixVerID.Valid {
		v := fixVerID.Int64
		iss.FixVersionID = &v
	}
	if releaseVerID.Valid {
		v := releaseVerID.Int64
		iss.ReleaseVersionID = &v
	}
	if sprintID.Valid {
		v := sprintID.Int64
		iss.SprintID = &v
	}

	iss.Identifier = identifier + "-" + strconv.FormatInt(iss.SequenceID, 10)
	iss.Assignees, _ = s.loadAssignees(ctx, iss.ID)
	iss.Labels, _ = s.loadLabels(ctx, iss.ID)
	iss.Modules, _ = s.loadModules(ctx, iss.ID)
	iss.Watchers, _ = s.loadWatchers(ctx, iss.ID)

	return &iss, nil
}

// List 列出工作项。
func (s *Service) List(ctx context.Context, opts ListIssuesOptions) ([]Issue, int64, error) {
	if opts.Limit <= 0 || opts.Limit > 100 {
		opts.Limit = 50
	}
	if opts.Offset < 0 {
		opts.Offset = 0
	}

	where, args := buildIssueWhere(opts)
	limitIdx := len(args) + 1
	offsetIdx := len(args) + 2
	args = append(args, opts.Limit, opts.Offset)

	rows, err := s.db.Query(ctx, `
		SELECT i.id, i.public_id, i.workspace_id, i.project_id, i.sequence_id,
		       i.type_code, i.parent_id, i.depth, i.name,
		       i.state_id, s.name, s.color, s."group",
		       i.priority, i.severity, i.category, i.point,
		       i.start_date, i.target_date, i.progress, i.version,
		       i.created_by, i.created_at, i.updated_at,
		       p.identifier
		FROM issues i
		JOIN states s ON s.id = i.state_id
		JOIN projects p ON p.id = i.project_id
		`+where+`
		ORDER BY `+buildIssueSort(opts)+`
		LIMIT $`+strconv.Itoa(limitIdx)+` OFFSET $`+strconv.Itoa(offsetIdx), args...)
	if err != nil {
		return nil, 0, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()

	issues, err := scanIssueRows(rows)
	if err != nil {
		return nil, 0, err
	}

	// 统计总数（分页元数据）
	countWhere, countArgs := buildCountWhere(opts)
	var total int64
	_ = s.db.QueryRow(ctx, `SELECT count(*) FROM issues i `+countWhere, countArgs...).Scan(&total)

	return issues, total, nil
}

// Update 更新工作项。
func (s *Service) Update(ctx context.Context, wsID, issueID int64, in UpdateIssueInput) (*Issue, error) {
	var result *Issue
	err := s.withTx(ctx, wsID, func(tx pgx.Tx) error {
		current, err := s.getByIDTx(ctx, tx, issueID, wsID)
		if err != nil {
			return err
		}

		// 校验版本
		if in.Version != current.Version {
			return errs.ErrVersionConflict
		}

		// WBS parent 变更校验
		if in.ParentID != nil && *in.ParentID != issueID {
			if err := s.validateWBSDepth(ctx, *in.ParentID); err != nil {
				return err
			}
			if err := s.validateNoCircular(ctx, tx, issueID, *in.ParentID); err != nil {
				return err
			}
		}

		// 构建更新
		sets, args := buildUpdateSet(in, current)
		if len(sets) == 0 {
			return s.updateM2M(ctx, tx, issueID, in)
		}

		args = append(args, issueID, wsID, in.Version)
		idIdx := len(args) - 2
		wsIdx := len(args) - 1
		verIdx := len(args)
		_ = idIdx
		_ = wsIdx
		_ = verIdx

		query := fmt.Sprintf(`UPDATE issues SET %s WHERE id = $%d AND workspace_id = $%d AND version = $%d AND deleted_at IS NULL`,
			strings.Join(sets, ", "), len(args)-2, len(args)-1, len(args))

		tag, err := tx.Exec(ctx, query, args...)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return errs.ErrVersionConflict
		}

		return s.updateM2M(ctx, tx, issueID, in)
	})
	if err != nil {
		return nil, err
	}
	result, err = s.GetByID(ctx, wsID, issueID)
	return result, err
}

// SoftDelete 归档（含级联子项）。
func (s *Service) SoftDelete(ctx context.Context, wsID, issueID int64) error {
	return s.withTx(ctx, wsID, func(tx pgx.Tx) error {
		// 从 sprint_issues 中移除；软删除不触发 FK cascade
		_, err := tx.Exec(ctx, `DELETE FROM sprint_issues WHERE issue_id = $1`, issueID)
		if err != nil {
			return err
		}

		tag, err := tx.Exec(ctx, `
			UPDATE issues SET deleted_at = now(), updated_at = now()
			WHERE id = $1 AND workspace_id = $2 AND deleted_at IS NULL`, issueID, wsID)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return errs.ErrNotFound
		}
		// 级联子项
		_, err = tx.Exec(ctx, `
			UPDATE issues SET deleted_at = now(), updated_at = now()
			WHERE parent_id = $1 AND workspace_id = $2 AND deleted_at IS NULL`, issueID, wsID)
		return err
	})
}

// Restore 从回收站恢复。
func (s *Service) Restore(ctx context.Context, wsID, issueID int64) error {
	tag, err := s.db.Exec(ctx, `
		UPDATE issues SET deleted_at = NULL, updated_at = now()
		WHERE id = $1 AND workspace_id = $2 AND deleted_at IS NOT NULL`, issueID, wsID)
	if err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	if tag.RowsAffected() == 0 {
		return errs.ErrNotFound
	}
	return nil
}

// Transition 执行状态流转。
func (s *Service) Transition(ctx context.Context, wsID, projectID, issueID, toStateID, userID int64) (*Issue, error) {
	err := s.withTx(ctx, wsID, func(tx pgx.Tx) error {
		iss, err := s.getByIDTx(ctx, tx, issueID, wsID)
		if err != nil {
			return err
		}
		if iss.StateID == toStateID {
			return nil
		}

		// 校验流转
		if err := s.stateSvc.ValidateTransition(ctx, wsID, projectID, TransitionInput{
			IssueID: issueID, FromState: iss.StateID, ToState: toStateID, TypeCode: iss.TypeCode,
			Context: TransitionContext{
				RootCauseCategory: iss.RootCauseCategory,
				FixVersionID:      iss.FixVersionID,
			},
		}); err != nil {
			return err
		}

		// 目标 group
		toGroup, err := s.stateSvc.StateGroupByID(ctx, toStateID)
		if err != nil {
			return err
		}

		// 缺陷完成时必填校验: root_cause_category + fix_version_id
		if toGroup == GroupCompleted && iss.TypeCode == TypeDefect {
			var rc sql.NullString
			var fv sql.NullInt64
			if scanErr := tx.QueryRow(ctx,
				`SELECT root_cause_category, fix_version_id FROM issues WHERE id = $1 AND workspace_id = $2`,
				issueID, wsID).Scan(&rc, &fv); scanErr != nil {
				return errs.ErrInternal.Wrap(scanErr)
			}
			if !rc.Valid || rc.String == "" {
				return errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "root_cause_category", Reason: "缺陷关闭时根因分类为必填"})
			}
			if !fv.Valid {
				return errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "fix_version_id", Reason: "缺陷关闭时修复版本为必填"})
			}
		}

		// 更新
		completedAtClause := "NULL"
		progress := iss.Progress
		if toGroup == GroupCompleted {
			completedAtClause = "now()"
			progress = 100
		}

		query := fmt.Sprintf(`UPDATE issues SET state_id = $1, completed_at = %s,
			progress = $2, version = version + 1, updated_at = now()
			WHERE id = $3 AND workspace_id = $4 AND deleted_at IS NULL`, completedAtClause)
		if err := tx.QueryRow(ctx, query, toStateID, progress, issueID, wsID).Scan(&iss.Version); err != nil {
			if err == pgx.ErrNoRows {
				return errs.ErrNotFound
			}
			return err
		}

		iss.StateID = toStateID

		// 进度回写
		if iss.ParentID != nil && toGroup == GroupCompleted {
			s.triggerProgressRollup(ctx, tx, *iss.ParentID)
		}

		// 录制领域事件：issue.status_changed → 通知关注人
		assignees := s.loadAssigneesTx(ctx, tx, issueID)
		fromStateName := s.getStateName(ctx, tx, iss.StateID)
		toStateName := s.getStateName(ctx, tx, toStateID)
		var identifier, issueName string
		_ = tx.QueryRow(ctx, `
			SELECT p.identifier, i.name
			FROM issues i JOIN projects p ON p.id = i.project_id
			WHERE i.id = $1`, issueID).Scan(&identifier, &issueName)
		_ = recordIssueEvent(ctx, tx, "issue.status_changed", issueEventPayload{
			WorkspaceID: iss.WorkspaceID,
			ProjectID:   iss.ProjectID,
			ActorID:     userID,
			ActorName:   s.getUserName(ctx, tx, userID),
			IssueID:     issueID,
			Identifier:  identifier,
			Name:        issueName,
			AssigneeIDs: assignees,
			FromState:   fromStateName,
			ToState:     toStateName,
		})

		return nil
	})
	if err != nil {
		return nil, err
	}
	return s.GetByID(ctx, wsID, issueID)
}

// Reorder 看板拖拽排序 — 中值插入策略。
// prev 和 next 至少需要提供一个；两者都 nil 则无操作。
func (s *Service) Reorder(ctx context.Context, wsID, issueID int64, in ReorderInput) (*Issue, error) {
	var newOrder float64

	switch {
	case in.PrevSortOrder == nil && in.NextSortOrder == nil:
		// 无参照节点，不做任何事
		return s.GetByID(ctx, wsID, issueID)
	case in.PrevSortOrder == nil:
		// 移到第一个：比 next 稍小
		newOrder = *in.NextSortOrder - 1.0
	case in.NextSortOrder == nil:
		// 移到最后一个：比 prev 稍大
		newOrder = *in.PrevSortOrder + 1.0
	default:
		// 中值插入
		newOrder = (*in.PrevSortOrder + *in.NextSortOrder) / 2.0
	}

	// 乐观锁 CAS：若客户端传入 version，则校验一致性，冲突返回 ErrVersionConflict
	if in.Version != nil {
		tag, err := s.db.Exec(ctx, `
			UPDATE issues SET sort_order = $1, updated_at = now(), version = version + 1
			WHERE id = $2 AND workspace_id = $3 AND deleted_at IS NULL AND version = $4`,
			newOrder, issueID, wsID, *in.Version)
		if err != nil {
			return nil, errs.ErrInternal.Wrap(fmt.Errorf("reorder: %w", err))
		}
		if tag.RowsAffected() == 0 {
			return nil, errs.ErrVersionConflict
		}
	} else {
		_, err := s.db.Exec(ctx, `
			UPDATE issues SET sort_order = $1, updated_at = now(), version = version + 1
			WHERE id = $2 AND workspace_id = $3 AND deleted_at IS NULL`,
			newOrder, issueID, wsID)
		if err != nil {
			return nil, errs.ErrInternal.Wrap(fmt.Errorf("reorder: %w", err))
		}
	}

	return s.GetByID(ctx, wsID, issueID)
}

// BatchUpdate 批量操作工作项（流转/指派/优先级/删除）。
// 采用`尽力而为`策略：单个失败不影响其他项继续。
func (s *Service) BatchUpdate(ctx context.Context, wsID, projectID, userID int64, in BatchUpdateInput) (BatchResult, error) {
	var result BatchResult

	for _, id := range in.IssueIDs {
		var batchErr error

		switch {
		case in.Delete:
			batchErr = s.SoftDelete(ctx, wsID, id)
		case in.ToStateID != nil:
			_, batchErr = s.Transition(ctx, wsID, projectID, id, *in.ToStateID, userID)
		case in.AssigneeID != nil || in.Priority != nil:
			// 批量场景无客户端版本号：读取当前版本后按最新值更新（乐观锁以数据库为准）
			var ver int
			if err := s.db.QueryRow(ctx,
				`SELECT version FROM issues WHERE id = $1 AND workspace_id = $2 AND deleted_at IS NULL`,
				id, wsID).Scan(&ver); err != nil {
				batchErr = errs.ErrInternal.Wrap(err)
			} else {
				batchErr = s.directUpdateIssue(ctx, wsID, id, in.AssigneeID, in.Priority, ver)
			}
		default:
			continue
		}

		if batchErr != nil {
			result.Failed++
		} else {
			result.Succeeded++
		}
	}

	if result.Succeeded == 0 && result.Failed > 0 {
		return result, errs.ErrValidation.WithDetails(errs.FieldDetail{
			Field: "issue_ids", Reason: "所有工作项操作均失败",
		})
	}

	return result, nil
}

// directUpdateIssue 更新单个工作项的指派人/优先级，带乐观锁校验
// expectedVersion: 客户端获取到的最新版本号，更新时校验版本一致性，冲突返回errs.ErrVersionConflict
func (s *Service) directUpdateIssue(ctx context.Context, wsID, issueID int64, assigneeID *int64, priority *string, expectedVersion int) error {
	return s.withTx(ctx, wsID, func(tx pgx.Tx) error {
		// 更新指派人
		if assigneeID != nil {
			if _, err := tx.Exec(ctx, `DELETE FROM issue_assignees WHERE issue_id = $1`, issueID); err != nil {
				return err
			}
			if _, err := tx.Exec(ctx,
				`INSERT INTO issue_assignees (issue_id, user_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`,
				issueID, *assigneeID); err != nil {
				return err
			}
		}

		// 更新优先级（带乐观锁校验）
		if priority != nil {
			result, err := tx.Exec(ctx,
				`UPDATE issues SET priority = $1, updated_at = now(), version = version + 1
				 WHERE id = $2 AND workspace_id = $3 AND deleted_at IS NULL AND version = $4`,
				*priority, issueID, wsID, expectedVersion)
			if err != nil {
				return err
			}
			// 校验更新行数，0行说明版本冲突
			if result.RowsAffected() == 0 {
				return errs.ErrVersionConflict
			}
		}

		return nil
	})
}

// --- Helpers ---

// nextSequenceID 原子发号。
func (s *Service) nextSequenceID(ctx context.Context, projectID int64) (int64, error) {
	var seq int64
	err := s.db.QueryRow(ctx, `
		INSERT INTO project_sequences (project_id, next_value)
		VALUES ($1, 2)
		ON CONFLICT (project_id) DO UPDATE SET next_value = project_sequences.next_value + 1
		RETURNING next_value - 1`, projectID).Scan(&seq)
	if err != nil {
		return 0, errs.ErrInternal.Wrap(err)
	}
	return seq, nil
}

// insertIssue 在事务中插入 issue 并写 M2M + 领域事件。
func (s *Service) insertIssue(ctx context.Context, in CreateIssueInput, stateID, seqID int64) (int64, error) {
	var issueID int64
	err := s.withTx(ctx, in.WorkspaceID, func(tx pgx.Tx) error {
		depth := 1
		if in.ParentID != nil {
			d, err := s.getChildDepth(ctx, tx, *in.ParentID)
			if err != nil {
				return err
			}
			depth = d
		}

		var parent interface{}
		if in.ParentID != nil {
			parent = *in.ParentID
		}

		err := tx.QueryRow(ctx, `
		INSERT INTO issues (workspace_id, project_id, sequence_id, type_code, parent_id, depth,
			name, description_json, description_html, state_id, priority,
			severity, found_phase, reproduce_steps, category, source,
			point, external_id, start_date, target_date, is_draft, created_by,
			found_version_id, fix_version_id, release_version_id)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24,$25)
		RETURNING id`,
			in.WorkspaceID, in.ProjectID, seqID, string(in.TypeCode), parent, depth,
			in.Name, nil, in.DescriptionHTML, stateID, string(in.Priority),
			in.Severity, in.FoundPhase, in.ReproduceSteps, in.Category, in.Source,
			in.Point, in.ExternalID, in.StartDate, in.TargetDate, in.IsDraft, in.CreatedBy,
			in.FoundVersionID, in.FixVersionID, in.ReleaseVersionID).Scan(&issueID)
		if err != nil {
			return mapPgError(err)
		}

		if err := s.insertM2M(ctx, tx, issueID, in.Assignees, in.Labels, in.Modules); err != nil {
			return err
		}

		// 录制作领域事件：issue.created → 通知被分配人
		actorName := s.getUserName(ctx, tx, in.CreatedBy)
		var identifier string
		_ = tx.QueryRow(ctx, `SELECT identifier FROM projects WHERE id = $1`, in.ProjectID).Scan(&identifier)
		return recordIssueEvent(ctx, tx, "issue.created", issueEventPayload{
			WorkspaceID: in.WorkspaceID,
			ProjectID:   in.ProjectID,
			ActorID:     in.CreatedBy,
			ActorName:   actorName,
			IssueID:     issueID,
			Identifier:  identifier + "-" + strconv.FormatInt(seqID, 10),
			Name:        in.Name,
			AssigneeIDs: in.Assignees,
		})
	})
	return issueID, err
}

// getUserName 通过事务从 users 表查询用户显示名（事务安全）。
func (s *Service) getUserName(ctx context.Context, tx pgx.Tx, userID int64) string {
	var name string
	if tx != nil {
		_ = tx.QueryRow(ctx, `SELECT COALESCE(display_name,'') FROM users WHERE id = $1`, userID).Scan(&name)
	} else {
		_ = s.db.QueryRow(ctx, `SELECT COALESCE(display_name,'') FROM users WHERE id = $1`, userID).Scan(&name)
	}
	return name
}

// validateWBSDepth 校验父级深度。
func (s *Service) validateWBSDepth(ctx context.Context, parentID int64) error {
	var depth int
	err := s.db.QueryRow(ctx,
		`SELECT depth FROM issues WHERE id = $1 AND deleted_at IS NULL`, parentID).Scan(&depth)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errs.ErrNotFound
		}
		return errs.ErrInternal.Wrap(err)
	}
	if depth >= 3 {
		return errs.ErrWBSDepthExceeded
	}
	return nil
}

func (s *Service) getChildDepth(ctx context.Context, tx pgx.Tx, parentID int64) (int, error) {
	var depth int
	err := tx.QueryRow(ctx,
		`SELECT depth FROM issues WHERE id = $1 AND deleted_at IS NULL`, parentID).Scan(&depth)
	if err != nil {
		return 0, errs.ErrInternal.Wrap(err)
	}
	return depth + 1, nil
}

// validateNoCircular 校验新父级不在子树内。
func (s *Service) validateNoCircular(ctx context.Context, tx pgx.Tx, issueID, newParentID int64) error {
	var exists bool
	err := tx.QueryRow(ctx, `
		WITH RECURSIVE subtree AS (
			SELECT id FROM issues WHERE parent_id = $1 AND deleted_at IS NULL
			UNION ALL
			SELECT i.id FROM issues i JOIN subtree st ON i.parent_id = st.id WHERE i.deleted_at IS NULL
		)
		SELECT EXISTS(SELECT 1 FROM subtree WHERE id = $2)`, issueID, newParentID).Scan(&exists)
	if err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	if exists {
		return errs.ErrCircularParent
	}
	return nil
}

// --- Domain Event Recording ---
// 领域事件在业务事务内写入 domain_events 表，由 OutboxRelay 异步发布到 RabbitMQ。
// 事件消费者（NotificationConsumer/WebhookConsumer/SearchIndexConsumer）根据事件类型执行相应操作。

// issueEventPayload 是通知相关事件的标准 payload。
type issueEventPayload struct {
	WorkspaceID int64   `json:"workspace_id"`
	ProjectID   int64   `json:"project_id"`
	ActorID     int64   `json:"actor_id"`
	ActorName   string  `json:"actor_name"`
	IssueID     int64   `json:"issue_id"`
	Identifier  string  `json:"identifier"`
	Name        string  `json:"name"`
	AssigneeIDs []int64 `json:"assignee_ids"`
	FromState   string  `json:"from_state"`
	ToState     string  `json:"to_state"`
}

// recordIssueEvent 在既有事务内将 Issue 领域事件写入 Outbox（domain_events 表）。
// 事件类型约定：issue.created / issue.status_changed
func recordIssueEvent(ctx context.Context, tx pgx.Tx, eventType string, payload issueEventPayload) error {
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal event payload: %w", err)
	}
	_, err = tx.Exec(ctx, `
		INSERT INTO domain_events (workspace_id, aggregate_type, aggregate_id, event_type, payload)
		VALUES ($1, 'issue', $2, $3, $4)`,
		payload.WorkspaceID, payload.IssueID, eventType, payloadJSON)
	if err != nil {
		return fmt.Errorf("record issue event %s: %w", eventType, err)
	}
	return nil
}

// recordCommentEvent 在既有事务内将 Comment 领域事件写入 Outbox。
func recordCommentEvent(ctx context.Context, tx pgx.Tx, workspaceID, issueID, commentID, actorID int64, actorName, content string, assigneeIDs []int64) error {
	payload := struct {
		WorkspaceID int64   `json:"workspace_id"`
		ActorID     int64   `json:"actor_id"`
		ActorName   string  `json:"actor_name"`
		IssueID     int64   `json:"issue_id"`
		CommentID   int64   `json:"comment_id"`
		Content     string  `json:"content"`
		AssigneeIDs []int64 `json:"assignee_ids"`
	}{
		WorkspaceID: workspaceID,
		ActorID:     actorID,
		ActorName:   actorName,
		IssueID:     issueID,
		CommentID:   commentID,
		Content:     content,
		AssigneeIDs: assigneeIDs,
	}
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal comment payload: %w", err)
	}
	_, err = tx.Exec(ctx, `
		INSERT INTO domain_events (workspace_id, aggregate_type, aggregate_id, event_type, payload)
		VALUES ($1, 'comment', $2, 'comment.created', $3)`,
		workspaceID, issueID, payloadJSON)
	if err != nil {
		return fmt.Errorf("record comment event: %w", err)
	}
	return nil
}

// withTx 事务辅助（含 tenant context）。
func (s *Service) withTx(ctx context.Context, wsID int64, fn func(tx pgx.Tx) error) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, "SELECT set_config('app.workspace_id', $1, true)", strconv.FormatInt(wsID, 10)); err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	if err := fn(tx); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (s *Service) insertM2M(ctx context.Context, tx pgx.Tx, issueID int64, assignees, labels, modules []int64) error {
	for _, uid := range assignees {
		if _, err := tx.Exec(ctx,
			`INSERT INTO issue_assignees (issue_id, user_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`,
			issueID, uid); err != nil {
			return errs.ErrInternal.Wrap(err)
		}
	}
	for _, lid := range labels {
		if _, err := tx.Exec(ctx,
			`INSERT INTO issue_labels (issue_id, label_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`,
			issueID, lid); err != nil {
			return errs.ErrInternal.Wrap(err)
		}
	}
	for _, mid := range modules {
		if _, err := tx.Exec(ctx,
			`INSERT INTO issue_modules (issue_id, module_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`,
			issueID, mid); err != nil {
			return errs.ErrInternal.Wrap(err)
		}
	}
	return nil
}

func (s *Service) updateM2M(ctx context.Context, tx pgx.Tx, issueID int64, in UpdateIssueInput) error {
	if in.Assignees != nil {
		if _, err := tx.Exec(ctx, `DELETE FROM issue_assignees WHERE issue_id = $1`, issueID); err != nil {
			return err
		}
		for _, uid := range in.Assignees {
			if _, err := tx.Exec(ctx, `INSERT INTO issue_assignees (issue_id, user_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`, issueID, uid); err != nil {
				return err
			}
		}
	}
	if in.Labels != nil {
		if _, err := tx.Exec(ctx, `DELETE FROM issue_labels WHERE issue_id = $1`, issueID); err != nil {
			return err
		}
		for _, lid := range in.Labels {
			if _, err := tx.Exec(ctx, `INSERT INTO issue_labels (issue_id, label_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`, issueID, lid); err != nil {
				return err
			}
		}
	}
	if in.Modules != nil {
		if _, err := tx.Exec(ctx, `DELETE FROM issue_modules WHERE issue_id = $1`, issueID); err != nil {
			return err
		}
		for _, mid := range in.Modules {
			if _, err := tx.Exec(ctx, `INSERT INTO issue_modules (issue_id, module_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`, issueID, mid); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Service) loadAssigneesTx(ctx context.Context, tx pgx.Tx, issueID int64) []int64 {
	rows, err := tx.Query(ctx, `SELECT user_id FROM issue_assignees WHERE issue_id = $1`, issueID)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var ids []int64
	for rows.Next() {
		var uid int64
		if err := rows.Scan(&uid); err == nil {
			ids = append(ids, uid)
		}
	}
	return ids
}

// loadAssignees 通过连接池查询工作项指派人员（非事务版本）。
func (s *Service) loadAssignees(ctx context.Context, issueID int64) ([]int64, error) {
	return loadIntArray(ctx, s.db, `SELECT user_id FROM issue_assignees WHERE issue_id = $1`, issueID)
}

// getStateName 通过事务查状态名称。
func (s *Service) getStateName(ctx context.Context, tx pgx.Tx, stateID int64) string {
	var name string
	_ = tx.QueryRow(ctx, `SELECT name FROM states WHERE id = $1`, stateID).Scan(&name)
	return name
}

func (s *Service) loadLabels(ctx context.Context, issueID int64) ([]int64, error) {
	return loadIntArray(ctx, s.db, `SELECT label_id FROM issue_labels WHERE issue_id = $1`, issueID)
}

func (s *Service) loadModules(ctx context.Context, issueID int64) ([]int64, error) {
	return loadIntArray(ctx, s.db, `SELECT module_id FROM issue_modules WHERE issue_id = $1`, issueID)
}

// --- Watchers（关注订阅）---

// Watch 订阅关注工作项（幂等）。
func (s *Service) Watch(ctx context.Context, wsID, issueID, userID int64) error {
	// 校验工作项存在
	var exists bool
	if err := s.db.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM issues WHERE id = $1 AND workspace_id = $2 AND deleted_at IS NULL)`,
		issueID, wsID).Scan(&exists); err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	if !exists {
		return errs.ErrNotFound
	}

	if _, err := s.db.Exec(ctx,
		`INSERT INTO issue_watchers (issue_id, user_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`,
		issueID, userID); err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	return nil
}

// Unwatch 取消关注（幂等）。
func (s *Service) Unwatch(ctx context.Context, wsID, issueID, userID int64) error {
	if _, err := s.db.Exec(ctx,
		`DELETE FROM issue_watchers WHERE issue_id = $1 AND user_id = $2 AND issue_id IN (
			SELECT id FROM issues WHERE workspace_id = $3
		)`, issueID, userID, wsID); err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	return nil
}

func (s *Service) loadWatchers(ctx context.Context, issueID int64) ([]int64, error) {
	return loadIntArray(ctx, s.db, `SELECT user_id FROM issue_watchers WHERE issue_id = $1`, issueID)
}

// LoadWatchers 查询工作项关注者列表（供通知模块使用）。
func (s *Service) LoadWatchers(ctx context.Context, issueID int64) ([]int64, error) {
	return s.loadWatchers(ctx, issueID)
}

func (s *Service) getByIDTx(ctx context.Context, tx pgx.Tx, issueID, wsID int64) (*Issue, error) {
	var iss Issue
	var parentID sql.NullInt64
	err := tx.QueryRow(ctx, `
		SELECT id, workspace_id, project_id, sequence_id, type_code, parent_id, depth,
		       name, state_id, priority, version, created_by
		FROM issues WHERE id = $1 AND workspace_id = $2 AND deleted_at IS NULL`, issueID, wsID).Scan(
		&iss.ID, &iss.WorkspaceID, &iss.ProjectID, &iss.SequenceID,
		&iss.TypeCode, &parentID, &iss.Depth,
		&iss.Name, &iss.StateID, &iss.Priority, &iss.Version, &iss.CreatedBy)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrNotFound
		}
		return nil, errs.ErrInternal.Wrap(err)
	}
	if parentID.Valid {
		v := parentID.Int64
		iss.ParentID = &v
	}
	return &iss, nil
}

// triggerProgressRollup 递归回写父级进度。
func (s *Service) triggerProgressRollup(ctx context.Context, tx pgx.Tx, parentID int64) {
	var total, completed int
	err := tx.QueryRow(ctx, `
		SELECT count(*), count(*) FILTER (WHERE state_id IN (
		    SELECT id FROM states WHERE "group" = 'completed'
		))
		FROM issues WHERE parent_id = $1 AND workspace_id = (
		    SELECT workspace_id FROM issues WHERE id = $1
		) AND deleted_at IS NULL`, parentID).Scan(&total, &completed)
	if err != nil || total == 0 {
		return
	}
	progress := completed * 100 / total
	_, _ = tx.Exec(ctx, `UPDATE issues SET progress = $1, updated_at = now() WHERE id = $2`, progress, parentID)

	var grand sql.NullInt64
	_ = tx.QueryRow(ctx, `SELECT parent_id FROM issues WHERE id = $1`, parentID).Scan(&grand)
	if grand.Valid {
		s.triggerProgressRollup(ctx, tx, grand.Int64)
	}
}

// --- Query builders & validation ---

func buildIssueWhere(opts ListIssuesOptions) (string, []interface{}) {
	clauses := []string{"i.deleted_at IS NULL", "i.project_id = $1", "i.workspace_id = $2"}
	args := []interface{}{opts.ProjectID, opts.WorkspaceID}
	arg := 3

	if opts.StateID != nil {
		clauses = append(clauses, "i.state_id = $"+strconv.Itoa(arg))
		args = append(args, *opts.StateID)
		arg++
	}
	if opts.Group != nil {
		clauses = append(clauses, "s.\"group\" = $"+strconv.Itoa(arg))
		args = append(args, string(*opts.Group))
		arg++
	}
	if opts.TypeCode != nil {
		clauses = append(clauses, "i.type_code = $"+strconv.Itoa(arg))
		args = append(args, string(*opts.TypeCode))
		arg++
	}
	if opts.Priority != nil {
		clauses = append(clauses, "i.priority = $"+strconv.Itoa(arg))
		args = append(args, string(*opts.Priority))
		arg++
	}
	if opts.ParentID != nil {
		clauses = append(clauses, "i.parent_id = $"+strconv.Itoa(arg))
		args = append(args, *opts.ParentID)
		arg++
	}
	if opts.Search != "" {
		clauses = append(clauses, "i.name ILIKE $"+strconv.Itoa(arg))
		args = append(args, "%"+opts.Search+"%")
		arg++
	}
	if opts.FoundVersionID != nil {
		clauses = append(clauses, "i.found_version_id = $"+strconv.Itoa(arg))
		args = append(args, *opts.FoundVersionID)
		arg++
	}
	if opts.FixVersionID != nil {
		clauses = append(clauses, "i.fix_version_id = $"+strconv.Itoa(arg))
		args = append(args, *opts.FixVersionID)
		arg++
	}
	if opts.ReleaseVersionID != nil {
		clauses = append(clauses, "i.release_version_id = $"+strconv.Itoa(arg))
		args = append(args, *opts.ReleaseVersionID)
		arg++
	}
	if opts.AssigneeID != nil {
		clauses = append(clauses, "EXISTS(SELECT 1 FROM issue_assignees ia WHERE ia.issue_id = i.id AND ia.user_id = $"+strconv.Itoa(arg)+")")
		args = append(args, *opts.AssigneeID)
		arg++
	}
	if opts.LabelID != nil {
		clauses = append(clauses, "EXISTS(SELECT 1 FROM issue_labels il WHERE il.issue_id = i.id AND il.label_id = $"+strconv.Itoa(arg)+")")
		args = append(args, *opts.LabelID)
		arg++
	}
	if opts.ModuleID != nil {
		clauses = append(clauses, "EXISTS(SELECT 1 FROM issue_modules im WHERE im.issue_id = i.id AND im.module_id = $"+strconv.Itoa(arg)+")")
		args = append(args, *opts.ModuleID)
		arg++
	}
	if opts.SprintID != nil {
		clauses = append(clauses, "i.sprint_id = $"+strconv.Itoa(arg))
		args = append(args, *opts.SprintID)
		arg++
	}
	if opts.StartDateFrom != nil {
		clauses = append(clauses, "i.start_date >= $"+strconv.Itoa(arg)+"::date")
		args = append(args, *opts.StartDateFrom)
		arg++
	}
	if opts.TargetDateTo != nil {
		clauses = append(clauses, "i.target_date <= $"+strconv.Itoa(arg)+"::date")
		args = append(args, *opts.TargetDateTo)
		arg++
	}
	if opts.SeverityFrom != nil {
		clauses = append(clauses, "i.severity >= $"+strconv.Itoa(arg))
		args = append(args, *opts.SeverityFrom)
		arg++
	}

	return "WHERE " + strings.Join(clauses, " AND "), args
}

func buildCountWhere(opts ListIssuesOptions) (string, []interface{}) {
	clauses := []string{"deleted_at IS NULL", "project_id = $1", "workspace_id = $2"}
	args := []interface{}{opts.ProjectID, opts.WorkspaceID}
	arg := 3

	if opts.StateID != nil {
		clauses = append(clauses, "state_id = $"+strconv.Itoa(arg))
		args = append(args, *opts.StateID)
		arg++
	}
	if opts.TypeCode != nil {
		clauses = append(clauses, "type_code = $"+strconv.Itoa(arg))
		args = append(args, string(*opts.TypeCode))
		arg++
	}
	if opts.Priority != nil {
		clauses = append(clauses, "priority = $"+strconv.Itoa(arg))
		args = append(args, string(*opts.Priority))
		arg++
	}
	if opts.ParentID != nil {
		clauses = append(clauses, "parent_id = $"+strconv.Itoa(arg))
		args = append(args, *opts.ParentID)
		arg++
	}
	if opts.Search != "" {
		clauses = append(clauses, "name ILIKE $"+strconv.Itoa(arg))
		args = append(args, "%"+opts.Search+"%")
		arg++
	}

	return "WHERE " + strings.Join(clauses, " AND "), args
}

func buildIssueSort(opts ListIssuesOptions) string {
	dir := "ASC"
	if opts.SortDesc {
		dir = "DESC"
	}
	switch opts.SortBy {
	case "priority":
		return `CASE i.priority WHEN 'urgent' THEN 1 WHEN 'high' THEN 2 WHEN 'medium' THEN 3 WHEN 'low' THEN 4 ELSE 5 END ` + dir
	case "target_date":
		return "i.target_date " + dir
	case "created_at":
		return "i.created_at " + dir
	case "sequence":
		return "i.sequence_id " + dir
	default:
		return "i.updated_at " + dir
	}
}

func scanIssueRows(rows pgx.Rows) ([]Issue, error) {
	defer rows.Close()
	var issues []Issue
	for rows.Next() {
		var iss Issue
		var parentID sql.NullInt64
		var severity, point sql.NullInt64
		var category sql.NullString
		var targetDate sql.NullTime
		var stateName, stateColor, identifier string
		var stateGroup StateGroup

		if err := rows.Scan(
			&iss.ID, &iss.PublicID, &iss.WorkspaceID, &iss.ProjectID, &iss.SequenceID,
			&iss.TypeCode, &parentID, &iss.Depth, &iss.Name,
			&iss.StateID, &stateName, &stateColor, &stateGroup,
			&iss.Priority, &severity, &category, &point,
			&iss.StartDate, &targetDate, &iss.Progress, &iss.Version,
			&iss.CreatedBy, &iss.CreatedAt, &iss.UpdatedAt, &identifier); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}

		if parentID.Valid {
			v := parentID.Int64
			iss.ParentID = &v
		}
		iss.State = &State{ID: iss.StateID, Name: stateName, Color: stateColor, Group: stateGroup}
		if severity.Valid {
			v := int(severity.Int64)
			iss.Severity = &v
		}
		iss.Identifier = identifier + "-" + strconv.FormatInt(iss.SequenceID, 10)
		issues = append(issues, iss)
	}
	return issues, rows.Err()
}

func buildUpdateSet(in UpdateIssueInput, current *Issue) ([]string, []interface{}) {
	var sets []string
	var args []interface{}
	arg := 1
	_ = current

	if in.Name != nil {
		sets = append(sets, "name = $"+strconv.Itoa(arg))
		args = append(args, *in.Name)
		arg++
	}
	if in.DescriptionHTML != nil {
		sets = append(sets, "description_html = $"+strconv.Itoa(arg))
		args = append(args, *in.DescriptionHTML)
		arg++
	}
	if in.Priority != nil {
		sets = append(sets, "priority = $"+strconv.Itoa(arg))
		args = append(args, string(*in.Priority))
		arg++
	}
	if in.ParentID != nil {
		sets = append(sets, "parent_id = $"+strconv.Itoa(arg))
		args = append(args, *in.ParentID)
		arg++
	}
	if in.Severity != nil {
		sets = append(sets, "severity = $"+strconv.Itoa(arg))
		args = append(args, *in.Severity)
		arg++
	}
	if in.FoundPhase != nil {
		sets = append(sets, "found_phase = $"+strconv.Itoa(arg))
		args = append(args, *in.FoundPhase)
		arg++
	}
	if in.RootCauseCategory != nil {
		sets = append(sets, "root_cause_category = $"+strconv.Itoa(arg))
		args = append(args, *in.RootCauseCategory)
		arg++
	}
	if in.Category != nil {
		sets = append(sets, "category = $"+strconv.Itoa(arg))
		args = append(args, *in.Category)
		arg++
	}
	if in.FoundVersionID != nil {
		sets = append(sets, "found_version_id = $"+strconv.Itoa(arg))
		args = append(args, *in.FoundVersionID)
		arg++
	}
	if in.FixVersionID != nil {
		sets = append(sets, "fix_version_id = $"+strconv.Itoa(arg))
		args = append(args, *in.FixVersionID)
		arg++
	}
	if in.VerifierID != nil {
		sets = append(sets, "verifier_id = $"+strconv.Itoa(arg))
		args = append(args, *in.VerifierID)
		arg++
	}
	if len(in.ReproduceSteps) > 0 {
		sets = append(sets, "reproduce_steps = $"+strconv.Itoa(arg))
		args = append(args, in.ReproduceSteps)
		arg++
	}
	if in.ReleaseVersionID != nil {
		sets = append(sets, "release_version_id = $"+strconv.Itoa(arg))
		args = append(args, *in.ReleaseVersionID)
		arg++
	}
	if in.TypeCode != nil {
		sets = append(sets, "type_code = $"+strconv.Itoa(arg))
		args = append(args, string(*in.TypeCode))
		arg++
	}
	if in.Point != nil {
		sets = append(sets, "point = $"+strconv.Itoa(arg))
		args = append(args, *in.Point)
		arg++
	}
	if in.TargetDate != nil {
		sets = append(sets, "target_date = $"+strconv.Itoa(arg))
		args = append(args, *in.TargetDate)
		arg++
	}
	if in.Progress != nil {
		sets = append(sets, "progress = $"+strconv.Itoa(arg))
		args = append(args, *in.Progress)
		arg++
	}

	if len(sets) > 0 {
		sets = append(sets, "updated_at = now()")
	}
	return sets, args
}

// --- Validation ---

func validateCreateInput(in CreateIssueInput) error {
	if strings.TrimSpace(in.Name) == "" {
		return errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "name", Reason: "工作项名称不能为空"})
	}
	if len(in.Name) > 500 {
		return errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "name", Reason: "工作项名称不能超过 500 字符"})
	}
	if in.TypeCode != TypeEpic && in.TypeCode != TypeRequirement && in.TypeCode != TypeTask && in.TypeCode != TypeDefect {
		return errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "type_code", Reason: "无效的工作项类型"})
	}
	return nil
}

func validateTypeFields(tc IssueTypeCode, severity *int, foundPhase *string) error {
	if tc == TypeDefect {
		if severity == nil || *severity < 1 || *severity > 5 {
			return errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "severity", Reason: "缺陷严重程度为必填（1-5）"})
		}
		if foundPhase == nil || *foundPhase == "" {
			return errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "found_phase", Reason: "缺陷发现阶段为必填"})
		}
	}
	return nil
}

// --- Low-level ---

func loadIntArray(ctx context.Context, db *pgxpool.Pool, query string, arg interface{}) ([]int64, error) {
	rows, err := db.Query(ctx, query, arg)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []int64
	for rows.Next() {
		var v int64
		if err := rows.Scan(&v); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func mapPgError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.ConstraintName {
		case "issues_project_id_sequence_id_key":
			return errs.New("ISSUE.DUPLICATE_SEQ", "工作项序号冲突，请重试", 409)
		case "defect_required":
			return errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "severity", Reason: "缺陷严重程度和发现阶段为必填"})
		}
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return errs.ErrNotFound
	}
	return errs.ErrInternal.Wrap(err)
}

