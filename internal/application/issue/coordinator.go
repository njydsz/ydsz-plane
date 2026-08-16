// Package issue — Service 协调器：聚合 Task / Requirement / Defect 三个独立服务，
// 对外暴露单一入口，内部根据 type_code 分派到对应聚合根服务。
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
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// ===== 过渡期 DTO（跨类型 / 外部消费者使用）=====

// Issue 仅用于跨类型过渡 API 返回；聚合根优先使用 Requirement / Task / Defect。
type Issue struct {
	ID          int64     `json:"id"`
	PublicID    string    `json:"public_id"`
	WorkspaceID int64     `json:"workspace_id"`
	ProjectID   int64     `json:"project_id"`
	SequenceID  int64     `json:"sequence_id"`
	Identifier  string    `json:"identifier"`
	TypeCode    IssueTypeCode `json:"type_code"`
	ParentID    *int64    `json:"parent_id,omitempty"`
	Depth       int       `json:"depth"`
	Name        string    `json:"name"`
	StateID     int64     `json:"state_id"`
	State       *State    `json:"state,omitempty"`
	Priority    IssuePriority `json:"priority"`
	SprintID    *int64    `json:"sprint_id,omitempty"`
	VersionID   *int64    `json:"version_id,omitempty"`
	Progress    int       `json:"progress"`
	StartDate   *time.Time `json:"start_date,omitempty"`
	TargetDate  *time.Time `json:"target_date,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	IsDraft     bool      `json:"is_draft"`
	SortOrder   float64   `json:"sort_order"`
	Version     int       `json:"version"`
	CreatedBy   int64     `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	// 按类型的可选字段
	Severity   *int     `json:"severity,omitempty"`
	FoundPhase *string  `json:"found_phase,omitempty"`
	Category   *string  `json:"category,omitempty"`
	Point      *int     `json:"point,omitempty"`
	Assignees  []int64  `json:"assignees,omitempty"`
	Labels     []int64  `json:"labels,omitempty"`
	Modules    []int64  `json:"modules,omitempty"`
}

// CreateIssueInput 创建工作项的入参（过渡 API 用）。
type CreateIssueInput struct {
	WorkspaceID      int64
	ProjectID        int64
	TypeCode         IssueTypeCode
	Name             string
	DescriptionHTML  string
	StateID          int64
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
	ExternalID       *string
	Point            *int
	Environment      map[string]any
	StartDate        *time.Time
	TargetDate       *time.Time
	IsDraft          bool
	CreatedBy        int64
	FoundVersionID   *int64
	FixVersionID     *int64
	ReleaseVersionID *int64
}

// UpdateIssueInput 更新工作项的入参（过渡 API 用）。
type UpdateIssueInput struct {
	Name              *string
	DescriptionHTML   *string
	Priority          *IssuePriority
	TypeCode          *IssueTypeCode
	ParentID          *int64
	Severity          *int
	FoundPhase        *string
	RootCauseCategory *string
	VerifierID        *int64
	ReproduceSteps    json.RawMessage
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
}

// ListIssuesOptions 保留作为 List 入口。
type ListIssuesOptions = ListWorkitemsOptions

// ReorderInput 看板拖拽排序输入。
type ReorderInput struct {
	PrevSortOrder *float64
	NextSortOrder *float64
	Version       *int
}

// ===== 协调器服务 =====

// Service 是跨类型协调器，对外暴露单一入口。
type Service struct {
	db          *pgxpool.Pool
	Task        *TaskService
	Requirement *RequirementService
	Defect      *DefectService
	stateSvc    *StateService
}

// NewService 构造协调器。
func NewService(db *pgxpool.Pool) *Service {
	return &Service{
		db:          db,
		Task:        NewTaskService(db),
		Requirement: NewRequirementService(db),
		Defect:      NewDefectService(db),
		stateSvc:    NewStateService(db),
	}
}

// Create 创建工作项 — 根据 TypeCode 分派。
func (s *Service) Create(ctx context.Context, in CreateIssueInput) (*Issue, error) {
	switch in.TypeCode {
	case TypeTask:
		if in.Severity != nil {
			return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "severity", Reason: "任务不支持 severity"})
		}
		created, err := s.Task.Create(ctx, CreateTaskInput{
			WorkspaceID: in.WorkspaceID, ProjectID: in.ProjectID, Name: in.Name,
			DescriptionHTML: in.DescriptionHTML, StateID: in.StateID, Priority: in.Priority,
			ParentID: in.ParentID, Point: in.Point, StartDate: in.StartDate, TargetDate: in.TargetDate,
			IsDraft: in.IsDraft, CreatedBy: in.CreatedBy, Category: in.Category,
			Assignees: in.Assignees, Labels: in.Labels, Modules: in.Modules,
		})
		if err != nil {
			return nil, err
		}
		return taskToIssue(created), nil
	case TypeRequirement:
		created, err := s.Requirement.Create(ctx, CreateRequirementInput{
			WorkspaceID: in.WorkspaceID, ProjectID: in.ProjectID, Name: in.Name,
			DescriptionHTML: in.DescriptionHTML, StateID: in.StateID, Priority: in.Priority,
			ParentID: in.ParentID, Point: in.Point, StartDate: in.StartDate, TargetDate: in.TargetDate,
			IsDraft: in.IsDraft, CreatedBy: in.CreatedBy, Source: in.Source,
			Assignees: in.Assignees, Labels: in.Labels, Modules: in.Modules,
		})
		if err != nil {
			return nil, err
		}
		return requirementToIssue(created), nil
	case TypeDefect:
		if in.Severity == nil || in.FoundPhase == nil {
			return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "defect_fields", Reason: "缺陷必须提供 severity 和 found_phase"})
		}
		created, err := s.Defect.Create(ctx, CreateDefectInput{
			WorkspaceID: in.WorkspaceID, ProjectID: in.ProjectID, Name: in.Name,
			DescriptionHTML: in.DescriptionHTML, StateID: in.StateID, Priority: in.Priority,
			ParentID: in.ParentID, Point: in.Point, StartDate: in.StartDate, TargetDate: in.TargetDate,
			IsDraft: in.IsDraft, CreatedBy: in.CreatedBy, Severity: *in.Severity,
			FoundPhase: *in.FoundPhase, ReproduceSteps: in.ReproduceSteps, Environment: in.Environment,
			SourceVersionID: in.FoundVersionID,
			Assignees: in.Assignees, Labels: in.Labels, Modules: in.Modules,
		})
		if err != nil {
			return nil, err
		}
		return defectToIssue(created), nil
	}
	return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "type_code", Reason: "无效的工作项类型"})
}

// GetByID 按类型分派查询。
func (s *Service) GetByID(ctx context.Context, wsID, issueID int64) (*Issue, error) {
	tc, err := s.detectType(ctx, wsID, issueID)
	if err != nil {
		return nil, err
	}
	switch tc {
	case TypeTask:
		t, err := s.Task.GetByID(ctx, wsID, issueID)
		if err != nil { return nil, err }
		return taskToIssue(t), nil
	case TypeRequirement:
		r, err := s.Requirement.GetByID(ctx, wsID, issueID)
		if err != nil { return nil, err }
		return requirementToIssue(r), nil
	case TypeDefect:
		d, err := s.Defect.GetByID(ctx, wsID, issueID)
		if err != nil { return nil, err }
		return defectToIssue(d), nil
	}
	return nil, errs.ErrNotFound
}

// Update 按类型分派更新。
func (s *Service) Update(ctx context.Context, wsID, issueID int64, in UpdateIssueInput) (*Issue, error) {
	tc, err := s.detectType(ctx, wsID, issueID)
	if err != nil {
		return nil, err
	}
	switch tc {
	case TypeTask:
		updated, err := s.Task.Update(ctx, wsID, issueID, UpdateTaskInput{
			Name: in.Name, DescriptionHTML: in.DescriptionHTML, Priority: in.Priority,
			ParentID: in.ParentID, Point: in.Point, TargetDate: in.TargetDate, Progress: in.Progress,
			Version: in.Version, Assignees: in.Assignees, Labels: in.Labels, Modules: in.Modules,
		})
		if err != nil { return nil, err }
		return taskToIssue(updated), nil
	case TypeRequirement:
		updated, err := s.Requirement.Update(ctx, wsID, issueID, UpdateRequirementInput{
			Name: in.Name, DescriptionHTML: in.DescriptionHTML, Priority: in.Priority,
			ParentID: in.ParentID, Point: in.Point, TargetDate: in.TargetDate, Progress: in.Progress,
			Version: in.Version, Assignees: in.Assignees, Labels: in.Labels, Modules: in.Modules,
		})
		if err != nil { return nil, err }
		return requirementToIssue(updated), nil
	case TypeDefect:
		updated, err := s.Defect.Update(ctx, wsID, issueID, UpdateDefectInput{
			Name: in.Name, DescriptionHTML: in.DescriptionHTML, Priority: in.Priority,
			ParentID: in.ParentID, Point: in.Point, TargetDate: in.TargetDate, Progress: in.Progress,
			Version: in.Version, Severity: in.Severity, FoundPhase: in.FoundPhase,
			RootCauseCategory: in.RootCauseCategory, Assignees: in.Assignees, Labels: in.Labels, Modules: in.Modules,
			FoundVersionID: in.FoundVersionID, FixVersionID: in.FixVersionID,
		})
		if err != nil { return nil, err }
		return defectToIssue(updated), nil
	}
	return nil, errs.ErrNotFound
}

// SoftDelete 跨类型删除。
func (s *Service) SoftDelete(ctx context.Context, wsID, issueID int64) error {
	tc, err := s.detectType(ctx, wsID, issueID)
	if err != nil { return err }
	switch tc {
	case TypeTask: return s.Task.SoftDelete(ctx, wsID, issueID)
	case TypeRequirement: return s.Requirement.SoftDelete(ctx, wsID, issueID)
	case TypeDefect: return s.Defect.SoftDelete(ctx, wsID, issueID)
	}
	return errs.ErrNotFound
}

// Restore 从回收站恢复。
func (s *Service) Restore(ctx context.Context, wsID, issueID int64) error {
	tc, err := s.detectType(ctx, wsID, issueID)
	if err != nil { return err }
	switch tc {
	case TypeTask: return s.Task.Restore(ctx, wsID, issueID)
	case TypeRequirement: return s.Requirement.Restore(ctx, wsID, issueID)
	case TypeDefect: return s.Defect.Restore(ctx, wsID, issueID)
	}
	return errs.ErrNotFound
}

// Transition 执行状态流转。
func (s *Service) Transition(ctx context.Context, wsID, projectID, issueID, toStateID, userID int64) (*Issue, error) {
	tc, err := s.detectType(ctx, wsID, issueID)
	if err != nil { return nil, err }
	switch tc {
	case TypeTask:
		t, err := s.Task.Transition(ctx, wsID, projectID, issueID, toStateID, userID)
		if err != nil { return nil, err }
		return taskToIssue(t), nil
	case TypeRequirement:
		r, err := s.Requirement.Transition(ctx, wsID, projectID, issueID, toStateID, userID)
		if err != nil { return nil, err }
		return requirementToIssue(r), nil
	case TypeDefect:
		d, err := s.Defect.Transition(ctx, wsID, projectID, issueID, toStateID, userID)
		if err != nil { return nil, err }
		return defectToIssue(d), nil
	}
	return nil, errs.ErrNotFound
}

// Reorder 看板拖拽排序 — 按类型分派。
func (s *Service) Reorder(ctx context.Context, wsID, issueID int64, in ReorderInput) (*Issue, error) {
	tc, err := s.detectType(ctx, wsID, issueID)
	if err != nil { return nil, err }

	var newOrder float64
	switch {
	case in.PrevSortOrder == nil && in.NextSortOrder == nil:
		return s.GetByID(ctx, wsID, issueID)
	case in.PrevSortOrder == nil:
		newOrder = *in.NextSortOrder - 1.0
	case in.NextSortOrder == nil:
		newOrder = *in.PrevSortOrder + 1.0
	default:
		newOrder = (*in.PrevSortOrder + *in.NextSortOrder) / 2.0
	}

	table := tc.Table()
	if in.Version != nil {
		tag, err := s.db.Exec(ctx, `
			UPDATE `+table+` SET sort_order = $1, updated_at = now(), version = version + 1
			WHERE id = $2 AND workspace_id = $3 AND deleted = false AND version = $4`,
			newOrder, issueID, wsID, *in.Version)
		if err != nil { return nil, errs.ErrInternal.Wrap(err) }
		if tag.RowsAffected() == 0 { return nil, errs.ErrVersionConflict }
	} else {
		_, err := s.db.Exec(ctx, `
			UPDATE `+table+` SET sort_order = $1, updated_at = now(), version = version + 1
			WHERE id = $2 AND workspace_id = $3 AND deleted = false`,
			newOrder, issueID, wsID)
		if err != nil { return nil, errs.ErrInternal.Wrap(err) }
	}
	return s.GetByID(ctx, wsID, issueID)
}

// BatchUpdate 批量操作 — 按 ID 分派。
func (s *Service) BatchUpdate(ctx context.Context, wsID, projectID, userID int64, in BatchUpdateInput) (BatchResult, error) {
	var result BatchResult
	for _, id := range in.IDs {
		var batchErr error
		switch {
		case in.Delete:
			batchErr = s.SoftDelete(ctx, wsID, id)
		case in.ToStateID != nil:
			_, batchErr = s.Transition(ctx, wsID, projectID, id, *in.ToStateID, userID)
		default:
			tc, dErr := s.detectType(ctx, wsID, id)
			if dErr != nil {
				batchErr = dErr
				break
			}
			var ver int
			batchErr = s.db.QueryRow(ctx, `SELECT version FROM `+tc.Table()+` WHERE id = $1 AND workspace_id = $2 AND deleted = false`, id, wsID).Scan(&ver)
			if batchErr != nil { break }
			batchErr = s.directUpdate(ctx, tc, wsID, id, in.AssigneeID, in.Priority, ver)
		}
		if batchErr != nil {
			result.Failed++
		} else {
			result.Succeeded++
		}
	}
	if result.Succeeded == 0 && result.Failed > 0 {
		return result, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "ids", Reason: "所有工作项操作均失败"})
	}
	return result, nil
}

func (s *Service) directUpdate(ctx context.Context, tc IssueTypeCode, wsID, issueID int64, assigneeID *int64, priority *string, expectedVersion int) error {
	prefix := workitemM2MPrefix(tc)
	idCol := prefix + "_id"
	return coordinatorWithTx(ctx, s.db, wsID, func(tx pgx.Tx) error {
		if assigneeID != nil {
			if _, err := tx.Exec(ctx, fmt.Sprintf(`DELETE FROM %s_assignees WHERE %s = $1`, prefix, idCol), issueID); err != nil { return err }
			if _, err := tx.Exec(ctx, fmt.Sprintf(`INSERT INTO %s_assignees (%s, user_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`, prefix, idCol), issueID, *assigneeID); err != nil { return err }
		}
		if priority != nil {
			tag, err := tx.Exec(ctx, `
				UPDATE `+tc.Table()+` SET priority = $1, updated_at = now(), version = version + 1
				WHERE id = $2 AND workspace_id = $3 AND deleted = false AND version = $4`,
				*priority, issueID, wsID, expectedVersion)
			if err != nil { return err }
			if tag.RowsAffected() == 0 { return errs.ErrVersionConflict }
		}
		return nil
	})
}

// List 跨类型列表 — UNION ALL 三表。
func (s *Service) List(ctx context.Context, opts ListIssuesOptions) ([]Issue, int64, error) {
	if opts.Limit <= 0 || opts.Limit > 100 { opts.Limit = 50 }
	if opts.Offset < 0 { opts.Offset = 0 }

	where, args := buildCoordinatorWhere(opts)
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
		FROM (
			SELECT id, public_id, workspace_id, project_id, sequence_id, 'task'::text AS type_code,
			       parent_id, depth, name, state_id, priority, NULL::int AS severity,
			       category, point, start_date, target_date, progress, version,
			       created_by, created_at, updated_at, project_id AS pid
			FROM task WHERE workspace_id = $2 AND deleted = false
			UNION ALL
			SELECT id, public_id, workspace_id, project_id, sequence_id, 'requirement'::text,
			       parent_id, depth, name, state_id, priority, NULL::int,
			       NULL::text, point, start_date, target_date, progress, version,
			       created_by, created_at, updated_at, project_id
			FROM requirement WHERE workspace_id = $2 AND deleted = false
			UNION ALL
			SELECT id, public_id, workspace_id, project_id, sequence_id, 'defect'::text,
			       parent_id, depth, name, state_id, priority, severity,
			       NULL::text, point, start_date, target_date, progress, version,
			       created_by, created_at, updated_at, project_id
			FROM defect WHERE workspace_id = $2 AND deleted = false
		) i
		JOIN states s ON s.id = i.state_id
		JOIN projects p ON p.id = i.pid
		`+where+`
		ORDER BY `+buildCoordinatorSort(opts)+`
		LIMIT $`+strconv.Itoa(limitIdx)+` OFFSET $`+strconv.Itoa(offsetIdx), args...)
	if err != nil { return nil, 0, errs.ErrInternal.Wrap(err) }
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
			return nil, 0, errs.ErrInternal.Wrap(err)
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
		if category.Valid { iss.Category = &category.String }
		if point.Valid {
			v := int(point.Int64)
			iss.Point = &v
		}
		if targetDate.Valid { iss.TargetDate = &targetDate.Time }
		iss.Identifier = identifier + "-" + strconv.FormatInt(iss.SequenceID, 10)
		issues = append(issues, iss)
	}

	// count 由调用方估算或另行处理
	return issues, 0, rows.Err()
}

// Watch / Unwatch 共用分表 {type}_watchers 表。
func (s *Service) Watch(ctx context.Context, wsID, issueID, userID int64) error {
	tc, err := detectWorkitemType(ctx, s.db, issueID)
	if err != nil {
		return err
	}
	prefix := workitemM2MPrefix(tc)
	_, err = s.db.Exec(ctx,
		fmt.Sprintf(`INSERT INTO %s_watchers (workspace_id, %s_id, user_id) VALUES ($1,$2,$3) ON CONFLICT DO NOTHING`, prefix, prefix),
		wsID, issueID, userID)
	if err != nil { return errs.ErrInternal.Wrap(err) }
	return nil
}

func (s *Service) Unwatch(ctx context.Context, wsID, issueID, userID int64) error {
	tc, err := detectWorkitemType(ctx, s.db, issueID)
	if err != nil {
		return err
	}
	prefix := workitemM2MPrefix(tc)
	_, err = s.db.Exec(ctx,
		fmt.Sprintf(`DELETE FROM %s_watchers WHERE %s_id = $1 AND user_id = $2`, prefix, prefix), issueID, userID)
	if err != nil { return errs.ErrInternal.Wrap(err) }
	return nil
}

func (s *Service) LoadWatchers(ctx context.Context, issueID int64) ([]int64, error) {
	tc, err := detectWorkitemType(ctx, s.db, issueID)
	if err != nil {
		return nil, err
	}
	prefix := workitemM2MPrefix(tc)
	return loadIntArray(ctx, s.db, fmt.Sprintf(`SELECT user_id FROM %s_watchers WHERE %s_id = $1`, prefix, prefix), issueID)
}

// detectType 确定工作项类型。
func (s *Service) detectType(ctx context.Context, wsID, issueID int64) (IssueTypeCode, error) {
	var tc string
	err := s.db.QueryRow(ctx, `
		SELECT 'task' FROM task WHERE id = $1 AND workspace_id = $2 AND deleted = false
		UNION ALL
		SELECT 'requirement' FROM requirement WHERE id = $1 AND workspace_id = $2 AND deleted = false
		UNION ALL
		SELECT 'defect' FROM defect WHERE id = $1 AND workspace_id = $2 AND deleted = false
		LIMIT 1`, issueID, wsID).Scan(&tc)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) { return "", errs.ErrNotFound }
		return "", errs.ErrInternal.Wrap(err)
	}
	return IssueTypeCode(tc), nil
}

func coordinatorWithTx(ctx context.Context, db *pgxpool.Pool, wsID int64, fn func(tx pgx.Tx) error) error {
	tx, err := db.Begin(ctx)
	if err != nil { return errs.ErrInternal.Wrap(err) }
	defer func() { _ = tx.Rollback(ctx) }()
	if _, err := tx.Exec(ctx, "SELECT set_config('app.workspace_id', $1, true)", strconv.FormatInt(wsID, 10)); err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	if err := fn(tx); err != nil { return err }
	return tx.Commit(ctx)
}

// Table 返回类型对应的表名。
func (tc IssueTypeCode) Table() string {
	switch tc {
	case TypeTask: return "task"
	case TypeRequirement: return "requirement"
	case TypeDefect: return "defect"
	}
	return "task"
}

// --- 聚合根 → Issue DTO 转换 ---

func taskToIssue(t *Task) *Issue {
	iss := &Issue{
		ID: t.ID, PublicID: t.PublicID, WorkspaceID: t.WorkspaceID, ProjectID: t.ProjectID,
		SequenceID: t.SequenceID, Identifier: t.Identifier, TypeCode: t.TypeCode,
		ParentID: t.ParentID, Depth: t.Depth, Name: t.Name, StateID: t.StateID, State: t.State,
		Priority: t.Priority, SprintID: t.SprintID, VersionID: t.VersionID, Progress: t.Progress,
		TargetDate: t.TargetDate, IsDraft: t.IsDraft, SortOrder: t.SortOrder, Version: t.Version,
		CreatedBy: t.CreatedBy, CreatedAt: t.CreatedAt, UpdatedAt: t.UpdatedAt,
		Category: t.Category, Assignees: t.Assignees, Labels: t.Labels, Modules: t.Modules,
	}
	return iss
}

func requirementToIssue(r *Requirement) *Issue {
	return &Issue{
		ID: r.ID, PublicID: r.PublicID, WorkspaceID: r.WorkspaceID, ProjectID: r.ProjectID,
		SequenceID: r.SequenceID, Identifier: r.Identifier, TypeCode: r.TypeCode,
		ParentID: r.ParentID, Depth: r.Depth, Name: r.Name, StateID: r.StateID, State: r.State,
		Priority: r.Priority, SprintID: r.SprintID, VersionID: r.VersionID, Progress: r.Progress,
		TargetDate: r.TargetDate, IsDraft: r.IsDraft, SortOrder: r.SortOrder, Version: r.Version,
		CreatedBy: r.CreatedBy, CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt,
		Assignees: r.Assignees, Labels: r.Labels, Modules: r.Modules,
	}
}

func defectToIssue(d *Defect) *Issue {
	severity := d.Severity
	return &Issue{
		ID: d.ID, PublicID: d.PublicID, WorkspaceID: d.WorkspaceID, ProjectID: d.ProjectID,
		SequenceID: d.SequenceID, Identifier: d.Identifier, TypeCode: d.TypeCode,
		ParentID: d.ParentID, Depth: d.Depth, Name: d.Name, StateID: d.StateID, State: d.State,
		Priority: d.Priority, SprintID: d.SprintID, VersionID: d.VersionID, Progress: d.Progress,
		TargetDate: d.TargetDate, IsDraft: d.IsDraft, SortOrder: d.SortOrder, Version: d.Version,
		CreatedBy: d.CreatedBy, CreatedAt: d.CreatedAt, UpdatedAt: d.UpdatedAt,
		Severity: &severity, FoundPhase: &d.FoundPhase,
		Assignees: d.Assignees, Labels: d.Labels, Modules: d.Modules,
	}
}

// --- 跨类型 where/sort builder ---

func buildCoordinatorWhere(opts ListIssuesOptions) (string, []interface{}) {
	clauses := []string{"1=1"}
	args := []interface{}{opts.WorkspaceID}
	arg := 2
	if opts.ProjectID != 0 {
		clauses = append(clauses, "i.project_id = $"+strconv.Itoa(arg)); args = append(args, opts.ProjectID); arg++
	}
	if opts.StateID != nil {
		clauses = append(clauses, "i.state_id = $"+strconv.Itoa(arg)); args = append(args, *opts.StateID); arg++
	}
	if opts.Group != nil {
		clauses = append(clauses, `s."group" = $`+strconv.Itoa(arg)); args = append(args, string(*opts.Group)); arg++
	}
	if opts.TypeCode != nil {
		clauses = append(clauses, "i.type_code = $"+strconv.Itoa(arg)); args = append(args, string(*opts.TypeCode)); arg++
	}
	if opts.Priority != nil {
		clauses = append(clauses, "i.priority = $"+strconv.Itoa(arg)); args = append(args, string(*opts.Priority)); arg++
	}
	if opts.ParentID != nil {
		clauses = append(clauses, "i.parent_id = $"+strconv.Itoa(arg)); args = append(args, *opts.ParentID); arg++
	}
	if opts.Search != "" {
		clauses = append(clauses, "i.name ILIKE $"+strconv.Itoa(arg)); args = append(args, "%"+opts.Search+"%"); arg++
	}
	if opts.AssigneeID != nil {
		clauses = append(clauses, "EXISTS(SELECT 1 FROM task_assignees WHERE task_id = i.id AND user_id = $"+strconv.Itoa(arg)+" UNION ALL SELECT 1 FROM requirement_assignees WHERE requirement_id = i.id AND user_id = $"+strconv.Itoa(arg)+" UNION ALL SELECT 1 FROM defect_assignees WHERE defect_id = i.id AND user_id = $"+strconv.Itoa(arg)+")")
		args = append(args, *opts.AssigneeID); arg++
	}
	if opts.LabelID != nil {
		clauses = append(clauses, "EXISTS(SELECT 1 FROM task_labels WHERE task_id = i.id AND label_id = $"+strconv.Itoa(arg)+" UNION ALL SELECT 1 FROM requirement_labels WHERE requirement_id = i.id AND label_id = $"+strconv.Itoa(arg)+" UNION ALL SELECT 1 FROM defect_labels WHERE defect_id = i.id AND label_id = $"+strconv.Itoa(arg)+")")
		args = append(args, *opts.LabelID); arg++
	}
	if opts.ModuleID != nil {
		clauses = append(clauses, "EXISTS(SELECT 1 FROM task_modules WHERE task_id = i.id AND module_id = $"+strconv.Itoa(arg)+" UNION ALL SELECT 1 FROM requirement_modules WHERE requirement_id = i.id AND module_id = $"+strconv.Itoa(arg)+" UNION ALL SELECT 1 FROM defect_modules WHERE defect_id = i.id AND module_id = $"+strconv.Itoa(arg)+")")
		args = append(args, *opts.ModuleID); arg++
	}
	if opts.SprintID != nil {
		clauses = append(clauses, "EXISTS(SELECT 1 FROM sprint_tasks WHERE task_id = i.id AND sprint_id = $"+strconv.Itoa(arg)+" UNION ALL SELECT 1 FROM sprint_requirements WHERE requirement_id = i.id AND sprint_id = $"+strconv.Itoa(arg)+" UNION ALL SELECT 1 FROM sprint_defects WHERE defect_id = i.id AND sprint_id = $"+strconv.Itoa(arg)+")")
		args = append(args, *opts.SprintID); arg++
	}
	if opts.StartDateFrom != nil {
		clauses = append(clauses, "i.start_date >= $"+strconv.Itoa(arg)+"::date"); args = append(args, *opts.StartDateFrom); arg++
	}
	if opts.TargetDateTo != nil {
		clauses = append(clauses, "i.target_date <= $"+strconv.Itoa(arg)+"::date"); args = append(args, *opts.TargetDateTo); arg++
	}
	if opts.SeverityFrom != nil {
		clauses = append(clauses, "i.severity >= $"+strconv.Itoa(arg)); args = append(args, *opts.SeverityFrom); arg++
	}
	return "WHERE " + strings.Join(clauses, " AND "), args
}

func buildCoordinatorSort(opts ListIssuesOptions) string {
	dir := "ASC"
	if opts.SortDesc { dir = "DESC" }
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

