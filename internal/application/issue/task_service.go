// Package issue — Task 聚合根应用服务（task 表独立 CRUD）。
package issue

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// TaskService 提供 Task 领域应用服务。
type TaskService struct {
	db       *pgxpool.Pool
	stateSvc *StateService
}

// NewTaskService 创建 Task 服务。
func NewTaskService(db *pgxpool.Pool) *TaskService {
	return &TaskService{db: db, stateSvc: NewStateService(db)}
}

// Create 创建任务。
func (s *TaskService) Create(ctx context.Context, in CreateTaskInput) (*Task, error) {
	if in.Priority == "" {
		in.Priority = PriorityNone
	}
	if strings.TrimSpace(in.Name) == "" {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "name", Reason: "任务名称不能为空"})
	}
	if len(in.Name) > 500 {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "name", Reason: "任务名称不能超过 500 字符"})
	}
	if in.ParentID != nil {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "parent_id", Reason: "任务暂不支持父子层级"})
	}

	stateID := in.StateID
	if stateID == 0 {
		defaultState, err := s.stateSvc.GetDefaultState(ctx, in.WorkspaceID, in.ProjectID)
		if err != nil {
			return nil, err
		}
		stateID = defaultState.ID
	}

	seqID, err := s.nextSequenceID(ctx, in.ProjectID)
	if err != nil {
		return nil, err
	}

	taskID, err := s.insertTask(ctx, in, stateID, seqID)
	if err != nil {
		return nil, err
	}

	return s.GetByID(ctx, in.WorkspaceID, taskID)
}

// GetByID 获取任务详情。
func (s *TaskService) GetByID(ctx context.Context, wsID, taskID int64) (*Task, error) {
	var t Task
	var parentID sql.NullInt64
	var category sql.NullString
	var point sql.NullInt64
	var startDate, targetDate, completedAt sql.NullTime
	var stateName, stateColor, identifier string
	var stateGroup StateGroup
	var sprintID sql.NullInt64
	var versionID sql.NullInt64

	err := s.db.QueryRow(ctx, `
		SELECT t.id, t.public_id, t.workspace_id, t.project_id, t.sequence_id,
		       'task'::text, t.parent_id, t.depth, t.name,
		       t.description_json, t.description_html,
		       t.state_id, s.name, s.color, s."group",
		       t.priority, t.category, t.point,
		       t.start_date, t.target_date, t.completed_at, t.progress,
		       t.is_draft, t.version, t.created_by, t.created_at, t.updated_at,
		       p.identifier, t.sprint_id, t.version_id
		FROM task t
		JOIN states s ON s.id = t.state_id
		JOIN projects p ON p.id = t.project_id
		WHERE t.id = $1 AND t.workspace_id = $2 AND t.deleted = false`,
		taskID, wsID).Scan(
		&t.ID, &t.PublicID, &t.WorkspaceID, &t.ProjectID, &t.SequenceID,
		&t.TypeCode, &parentID, &t.Depth, &t.Name,
		&t.DescriptionJSON, &t.DescriptionHTML,
		&t.StateID, &stateName, &stateColor, &stateGroup,
		&t.Priority, &category, &point,
		&t.StartDate, &targetDate, &completedAt, &t.Progress,
		&t.IsDraft, &t.Version, &t.CreatedBy, &t.CreatedAt, &t.UpdatedAt,
		&identifier, &sprintID, &versionID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrNotFound
		}
		return nil, errs.ErrInternal.Wrap(err)
	}

	t.Identifier = identifier + "-" + strconv.FormatInt(t.SequenceID, 10)
	t.State = &State{ID: t.StateID, Name: stateName, Color: stateColor, Group: stateGroup}
	if parentID.Valid { v := parentID.Int64; t.ParentID = &v }
	if category.Valid { v := category.String; t.Category = &v }
	if point.Valid { v := int(point.Int64); t.Point = &v }
	if startDate.Valid { t.StartDate = &startDate.Time }
	if targetDate.Valid { t.TargetDate = &targetDate.Time }
	if completedAt.Valid { t.CompletedAt = &completedAt.Time }
	if sprintID.Valid { v := sprintID.Int64; t.SprintID = &v }
	if versionID.Valid { v := versionID.Int64; t.VersionID = &v }
	t.Assignees, _ = loadIntArray(ctx, s.db, `SELECT user_id FROM task_assignees WHERE task_id = $1`, taskID)
	t.Labels, _ = loadIntArray(ctx, s.db, `SELECT label_id FROM task_labels WHERE task_id = $1`, taskID)
	t.Modules, _ = loadIntArray(ctx, s.db, `SELECT module_id FROM task_modules WHERE task_id = $1`, taskID)
	t.Watchers, _ = loadIntArray(ctx, s.db, `SELECT user_id FROM task_watchers WHERE task_id = $1`, taskID)

	return &t, nil
}

// Update 更新任务。
func (s *TaskService) Update(ctx context.Context, wsID, taskID int64, in UpdateTaskInput) (*Task, error) {
	err := s.withTx(ctx, wsID, func(tx pgx.Tx) error {
		current, err := s.getByIDTx(ctx, tx, taskID, wsID)
		if err != nil { return err }
		if in.Version != current.Version {
			return errs.ErrVersionConflict
		}
		if in.ParentID != nil && *in.ParentID != taskID {
			return errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "parent_id", Reason: "任务暂不支持父子层级"})
		}

		sets, args := buildTaskUpdateSet(in)
		if len(sets) == 0 {
			return sharedUpdateM2M(ctx, tx, TypeTask, wsID, current.ProjectID, taskID, in.Assignees, in.Labels, in.Modules)
		}

		sets = append(sets, "updated_at = now()")
		query := fmt.Sprintf(`UPDATE task SET %s WHERE id = $%d AND workspace_id = $%d AND version = $%d AND deleted = false`,
			strings.Join(sets, ", "), len(args)+1, len(args)+2, len(args)+3)
		args = append(args, taskID, wsID, in.Version)

		tag, err := tx.Exec(ctx, query, args...)
		if err != nil { return err }
		if tag.RowsAffected() == 0 { return errs.ErrVersionConflict }
		return sharedUpdateM2M(ctx, tx, TypeTask, wsID, current.ProjectID, taskID, in.Assignees, in.Labels, in.Modules)
	})
	if err != nil { return nil, err }
	return s.GetByID(ctx, wsID, taskID)
}

// SoftDelete 归档。
func (s *TaskService) SoftDelete(ctx context.Context, wsID, taskID int64) error {
	return s.withTx(ctx, wsID, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `DELETE FROM sprint_tasks WHERE task_id = $1`, taskID); err != nil {
			return err
		}
		tag, err := tx.Exec(ctx, `
			UPDATE task SET deleted = true, updated_at = now()
			WHERE id = $1 AND workspace_id = $2 AND deleted = false`, taskID, wsID)
		if err != nil { return err }
		if tag.RowsAffected() == 0 { return errs.ErrNotFound }
		return nil
	})
}

// Restore 从回收站恢复。
func (s *TaskService) Restore(ctx context.Context, wsID, taskID int64) error {
	tag, err := s.db.Exec(ctx, `
		UPDATE task SET deleted = false, updated_at = now()
		WHERE id = $1 AND workspace_id = $2 AND deleted = true`, taskID, wsID)
	if err != nil { return errs.ErrInternal.Wrap(err) }
	if tag.RowsAffected() == 0 { return errs.ErrNotFound }
	return nil
}

// Transition 执行状态流转。
func (s *TaskService) Transition(ctx context.Context, wsID, projectID, taskID, toStateID, userID int64) (*Task, error) {
	err := s.withTx(ctx, wsID, func(tx pgx.Tx) error {
		t, err := s.getByIDTx(ctx, tx, taskID, wsID)
		if err != nil { return err }
		if t.StateID == toStateID { return nil }
		if err := s.stateSvc.ValidateTransition(ctx, wsID, projectID, TransitionInput{
			IssueID: taskID, FromState: t.StateID, ToState: toStateID, TypeCode: t.TypeCode,
		}); err != nil {
			return err
		}

		toGroup, err := s.stateSvc.StateGroupByID(ctx, toStateID)
		if err != nil { return err }
		completedAtClause := "NULL"
		progress := t.Progress
		if toGroup == GroupCompleted {
			completedAtClause = "now()"
			progress = 100
		}
		query := fmt.Sprintf(`UPDATE task SET state_id = $1, completed_at = %s, progress = $2, version = version + 1, updated_at = now()
			WHERE id = $3 AND workspace_id = $4 AND deleted = false`, completedAtClause)
		if _, err := tx.Exec(ctx, query, toStateID, progress, taskID, wsID); err != nil {
			return err
		}
		t.StateID = toStateID

		assignees := loadIntArrayTx(ctx, tx, `SELECT user_id FROM task_assignees WHERE task_id = $1`, taskID)
		var identifier, name string
		_ = tx.QueryRow(ctx, `SELECT p.identifier, t.name FROM task t JOIN projects p ON p.id = t.project_id
			WHERE t.id = $1`, taskID).Scan(&identifier, &name)
		actorName := getUserNameTx(ctx, tx, userID)
		return recordWorkitemEvent(ctx, tx, "workitem.status_changed", wsID, projectID, taskID, TypeTask,
			userID, actorName, identifier, name, assignees, loadStateName(ctx, tx, t.StateID), loadStateName(ctx, tx, toStateID))
	})
	if err != nil { return nil, err }
	return s.GetByID(ctx, wsID, taskID)
}

// --- 内部方法 ---

func (s *TaskService) insertTask(ctx context.Context, in CreateTaskInput, stateID, seqID int64) (int64, error) {
	var taskID int64
	err := s.withTx(ctx, in.WorkspaceID, func(tx pgx.Tx) error {
		depth := 1
		err := tx.QueryRow(ctx, `
			INSERT INTO task (workspace_id, project_id, sequence_id, parent_id, depth,
				name, description_json, description_html, state_id, priority,
				category, point, start_date, target_date, is_draft, created_by)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)
			RETURNING id`,
			in.WorkspaceID, in.ProjectID, seqID, in.ParentID, depth,
			in.Name, nil, in.DescriptionHTML, stateID, string(in.Priority),
			in.Category, in.Point, in.StartDate, in.TargetDate, in.IsDraft, in.CreatedBy).Scan(&taskID)
		if err != nil {
			return mapTaskPgError(err)
		}
		if err := insertM2M(ctx, tx, TypeTask, in.WorkspaceID, in.ProjectID, taskID, in.Assignees, in.Labels, in.Modules); err != nil {
			return err
		}
		var identifier string
		_ = tx.QueryRow(ctx, `SELECT identifier FROM projects WHERE id = $1`, in.ProjectID).Scan(&identifier)
		actorName := getUserNameTx(ctx, tx, in.CreatedBy)
		return recordWorkitemEvent(ctx, tx, "workitem.created", in.WorkspaceID, in.ProjectID, taskID, TypeTask,
			in.CreatedBy, actorName, identifier+"-"+strconv.FormatInt(seqID, 10), in.Name, in.Assignees, "", "")
	})
	return taskID, err
}

func (s *TaskService) getByIDTx(ctx context.Context, tx pgx.Tx, taskID, wsID int64) (*Task, error) {
	var t Task
	var parentID sql.NullInt64
	err := tx.QueryRow(ctx, `
		SELECT id, workspace_id, project_id, sequence_id, 'task'::text, parent_id, depth,
		       name, state_id, priority, version, created_by
		FROM task WHERE id = $1 AND workspace_id = $2 AND deleted = false`, taskID, wsID).Scan(
		&t.ID, &t.WorkspaceID, &t.ProjectID, &t.SequenceID,
		&t.TypeCode, &parentID, &t.Depth,
		&t.Name, &t.StateID, &t.Priority, &t.Version, &t.CreatedBy)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) { return nil, errs.ErrNotFound }
		return nil, errs.ErrInternal.Wrap(err)
	}
	if parentID.Valid { v := parentID.Int64; t.ParentID = &v }
	return &t, nil
}

func (s *TaskService) nextSequenceID(ctx context.Context, projectID int64) (int64, error) {
	var seq int64
	err := s.db.QueryRow(ctx, `
		INSERT INTO project_sequences (project_id, next_value) VALUES ($1, 2)
		ON CONFLICT (project_id) DO UPDATE SET next_value = project_sequences.next_value + 1
		RETURNING next_value - 1`, projectID).Scan(&seq)
	if err != nil { return 0, errs.ErrInternal.Wrap(err) }
	return seq, nil
}

func (s *TaskService) withTx(ctx context.Context, wsID int64, fn func(tx pgx.Tx) error) error {
	tx, err := s.db.Begin(ctx)
	if err != nil { return errs.ErrInternal.Wrap(err) }
	defer func() { _ = tx.Rollback(ctx) }()
	if _, err := tx.Exec(ctx, "SELECT set_config('app.workspace_id', $1, true)", strconv.FormatInt(wsID, 10)); err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	if err := fn(tx); err != nil { return err }
	return tx.Commit(ctx)
}

func buildTaskUpdateSet(in UpdateTaskInput) ([]string, []interface{}) {
	var sets []string
	var args []interface{}
	arg := 1
	if in.Name != nil {
		sets = append(sets, "name = $"+strconv.Itoa(arg)); args = append(args, *in.Name); arg++
	}
	if in.DescriptionHTML != nil {
		sets = append(sets, "description_html = $"+strconv.Itoa(arg)); args = append(args, *in.DescriptionHTML); arg++
	}
	if in.Priority != nil {
		sets = append(sets, "priority = $"+strconv.Itoa(arg)); args = append(args, string(*in.Priority)); arg++
	}
	if in.ParentID != nil {
		sets = append(sets, "parent_id = $"+strconv.Itoa(arg)); args = append(args, *in.ParentID); arg++
	}
	if in.Category != nil {
		sets = append(sets, "category = $"+strconv.Itoa(arg)); args = append(args, *in.Category); arg++
	}
	if in.Point != nil {
		sets = append(sets, "point = $"+strconv.Itoa(arg)); args = append(args, *in.Point); arg++
	}
	if in.TargetDate != nil {
		sets = append(sets, "target_date = $"+strconv.Itoa(arg)); args = append(args, *in.TargetDate); arg++
	}
	if in.Progress != nil {
		sets = append(sets, "progress = $"+strconv.Itoa(arg)); args = append(args, *in.Progress); arg++
	}
	return sets, args
}

func mapTaskPgError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgErr.ConstraintName == "task_project_id_sequence_id_key" {
			return errs.New("TASK.DUPLICATE_SEQ", "任务序号冲突，请重试", 409)
		}
	}
	if errors.Is(err, pgx.ErrNoRows) { return errs.ErrNotFound }
	return errs.ErrInternal.Wrap(err)
}
