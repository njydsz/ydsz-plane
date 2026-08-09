// Package issue — Requirement 聚合根应用服务（requirement 表独立 CRUD）。
package issue

import (
	"context"
	"database/sql"
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

// RequirementService 提供 Requirement 领域应用服务。
type RequirementService struct {
	db       *pgxpool.Pool
	stateSvc *StateService
}

// NewRequirementService 创建 Requirement 服务。
func NewRequirementService(db *pgxpool.Pool) *RequirementService {
	return &RequirementService{db: db, stateSvc: NewStateService(db)}
}

// Create 创建需求。
func (s *RequirementService) Create(ctx context.Context, in CreateRequirementInput) (*Requirement, error) {
	if in.Priority == "" {
		in.Priority = PriorityNone
	}
	if strings.TrimSpace(in.Name) == "" {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "name", Reason: "需求名称不能为空"})
	}
	if len(in.Name) > 500 {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "name", Reason: "需求名称不能超过 500 字符"})
	}

	if in.ParentID != nil {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "parent_id", Reason: "需求暂不支持父子层级"})
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

	reqID, err := s.insertRequirement(ctx, in, stateID, seqID)
	if err != nil {
		return nil, err
	}

	return s.GetByID(ctx, in.WorkspaceID, reqID)
}

// GetByID 获取需求详情。
func (s *RequirementService) GetByID(ctx context.Context, wsID, reqID int64) (*Requirement, error) {
	var r Requirement
	var parentID sql.NullInt64
	var source sql.NullString
	var point sql.NullInt64
	var startDate, targetDate, completedAt sql.NullTime
	var stateName, stateColor, identifier string
	var stateGroup StateGroup
	var sprintID sql.NullInt64
	var versionID sql.NullInt64

	err := s.db.QueryRow(ctx, `
		SELECT r.id, r.public_id, r.workspace_id, r.project_id, r.sequence_id,
		       'requirement'::text, r.parent_id, r.depth, r.name,
		       r.description_json, r.description_html,
		       r.state_id, s.name, s.color, s."group",
		       r.priority, r.source, r.point,
		       r.start_date, r.target_date, r.completed_at, r.progress,
		       r.is_draft, r.version, r.created_by, r.created_at, r.updated_at,
		       p.identifier,
		       r.sprint_id, r.version_id
		FROM requirement r
		JOIN states s ON s.id = r.state_id
		JOIN projects p ON p.id = r.project_id
		WHERE r.id = $1 AND r.workspace_id = $2 AND r.deleted_at IS NULL`,
		reqID, wsID).Scan(
		&r.ID, &r.PublicID, &r.WorkspaceID, &r.ProjectID, &r.SequenceID,
		&r.TypeCode, &parentID, &r.Depth, &r.Name,
		&r.DescriptionJSON, &r.DescriptionHTML,
		&r.StateID, &stateName, &stateColor, &stateGroup,
		&r.Priority, &source, &point,
		&r.StartDate, &targetDate, &completedAt, &r.Progress,
		&r.IsDraft, &r.Version, &r.CreatedBy, &r.CreatedAt, &r.UpdatedAt,
		&identifier, &sprintID, &versionID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrNotFound
		}
		return nil, errs.ErrInternal.Wrap(err)
	}

	r.Identifier = identifier + "-" + strconv.FormatInt(r.SequenceID, 10)
	r.State = &State{ID: r.StateID, Name: stateName, Color: stateColor, Group: stateGroup}
	if parentID.Valid {
		v := parentID.Int64
		r.ParentID = &v
	}
	if source.Valid {
		v := source.String
		r.Source = &v
	}
	if point.Valid {
		v := int(point.Int64)
		r.Point = &v
	}
	if startDate.Valid {
		r.StartDate = &startDate.Time
	}
	if targetDate.Valid {
		r.TargetDate = &targetDate.Time
	}
	if completedAt.Valid {
		r.CompletedAt = &completedAt.Time
	}
	if sprintID.Valid {
		v := sprintID.Int64
		r.SprintID = &v
	}
	if versionID.Valid {
		v := versionID.Int64
		r.VersionID = &v
	}
	r.Assignees, _ = loadIntArray(ctx, s.db, `SELECT user_id FROM issue_assignees WHERE issue_id = $1`, reqID)
	r.Labels, _ = loadIntArray(ctx, s.db, `SELECT label_id FROM issue_labels WHERE issue_id = $1`, reqID)
	r.Modules, _ = loadIntArray(ctx, s.db, `SELECT module_id FROM issue_modules WHERE issue_id = $1`, reqID)
	r.Watchers, _ = loadIntArray(ctx, s.db, `SELECT user_id FROM issue_watchers WHERE issue_id = $1`, reqID)

	return &r, nil
}

// Update 更新需求。
func (s *RequirementService) Update(ctx context.Context, wsID, reqID int64, in UpdateRequirementInput) (*Requirement, error) {
	var result *Requirement
	err := s.withTx(ctx, wsID, func(tx pgx.Tx) error {
		current, err := s.getByIDTx(ctx, tx, reqID, wsID)
		if err != nil {
			return err
		}
		if in.Version != current.Version {
			return errs.ErrVersionConflict
		}
		if in.ParentID != nil && *in.ParentID != reqID {
			return errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "parent_id", Reason: "需求暂不支持父子层级"})
		}

		sets, args := buildRequirementUpdateSet(in)
		if len(sets) == 0 {
			return s.updateM2M(ctx, tx, reqID, in.Assignees, in.Labels, in.Modules)
		}

		sets = append(sets, "updated_at = now()")
		query := fmt.Sprintf(`UPDATE requirement SET %s WHERE id = $%d AND workspace_id = $%d AND version = $%d AND deleted_at IS NULL`,
			strings.Join(sets, ", "), len(args)+1, len(args)+2, len(args)+3)
		args = append(args, reqID, wsID, in.Version)

		tag, err := tx.Exec(ctx, query, args...)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return errs.ErrVersionConflict
		}
		return s.updateM2M(ctx, tx, reqID, in.Assignees, in.Labels, in.Modules)
	})
	if err != nil {
		return nil, err
	}
	return s.GetByID(ctx, wsID, reqID)
}

// SoftDelete 归档。
func (s *RequirementService) SoftDelete(ctx context.Context, wsID, reqID int64) error {
	return s.withTx(ctx, wsID, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `DELETE FROM sprint_issues WHERE issue_id = $1`, reqID); err != nil {
			return err
		}
		tag, err := tx.Exec(ctx, `
			UPDATE requirement SET deleted_at = now(), updated_at = now()
			WHERE id = $1 AND workspace_id = $2 AND deleted_at IS NULL`, reqID, wsID)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return errs.ErrNotFound
		}
		return nil
	})
}

// Restore 从回收站恢复。
func (s *RequirementService) Restore(ctx context.Context, wsID, reqID int64) error {
	tag, err := s.db.Exec(ctx, `
		UPDATE requirement SET deleted_at = NULL, updated_at = now()
		WHERE id = $1 AND workspace_id = $2 AND deleted_at IS NOT NULL`, reqID, wsID)
	if err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	if tag.RowsAffected() == 0 {
		return errs.ErrNotFound
	}
	return nil
}

// Transition 执行状态流转。
func (s *RequirementService) Transition(ctx context.Context, wsID, projectID, reqID, toStateID, userID int64) (*Requirement, error) {
	err := s.withTx(ctx, wsID, func(tx pgx.Tx) error {
		r, err := s.getByIDTx(ctx, tx, reqID, wsID)
		if err != nil {
			return err
		}
		if r.StateID == toStateID {
			return nil
		}
		if err := s.stateSvc.ValidateTransition(ctx, wsID, projectID, TransitionInput{
			IssueID: reqID, FromState: r.StateID, ToState: toStateID, TypeCode: r.TypeCode,
		}); err != nil {
			return err
		}

		toGroup, err := s.stateSvc.StateGroupByID(ctx, toStateID)
		if err != nil {
			return err
		}
		completedAtClause := "NULL"
		progress := r.Progress
		if toGroup == GroupCompleted {
			completedAtClause = "now()"
			progress = 100
		}
		query := fmt.Sprintf(`UPDATE requirement SET state_id = $1, completed_at = %s,
			progress = $2, version = version + 1, updated_at = now()
			WHERE id = $3 AND workspace_id = $4 AND deleted_at IS NULL`, completedAtClause)
		if _, err := tx.Exec(ctx, query, toStateID, progress, reqID, wsID); err != nil {
			return err
		}
		r.StateID = toStateID

		assignees := loadIntArrayTx(ctx, tx, `SELECT user_id FROM issue_assignees WHERE issue_id = $1`, reqID)
		var identifier, reqName string
		_ = tx.QueryRow(ctx, `SELECT p.identifier, r.name
			FROM requirement r JOIN projects p ON p.id = r.project_id
			WHERE r.id = $1`, reqID).Scan(&identifier, &reqName)
		return recordWorkitemEvent(ctx, tx, "workitem.status_changed", wsID, projectID, reqID, TypeRequirement,
			userID, identifier, reqName, assignees, loadStateName(ctx, tx, r.StateID), loadStateName(ctx, tx, toStateID))
	})
	if err != nil {
		return nil, err
	}
	return s.GetByID(ctx, wsID, reqID)
}

// --- 内部方法 ---

func (s *RequirementService) insertRequirement(ctx context.Context, in CreateRequirementInput, stateID, seqID int64) (int64, error) {
	var reqID int64
	err := s.withTx(ctx, in.WorkspaceID, func(tx pgx.Tx) error {
		depth := 1
		err := tx.QueryRow(ctx, `
			INSERT INTO requirement (workspace_id, project_id, sequence_id, parent_id, depth,
				name, description_json, description_html, state_id, priority,
				source, point, start_date, target_date, is_draft, created_by)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)
			RETURNING id`,
			in.WorkspaceID, in.ProjectID, seqID, in.ParentID, depth,
			in.Name, nil, in.DescriptionHTML, stateID, string(in.Priority),
			in.Source, in.Point, in.StartDate, in.TargetDate, in.IsDraft, in.CreatedBy).Scan(&reqID)
		if err != nil {
			return mapPgRequirementErr(err)
		}
		if err := insertM2M(ctx, tx, reqID, in.Assignees, in.Labels, in.Modules); err != nil {
			return err
		}
		var identifier string
		_ = tx.QueryRow(ctx, `SELECT identifier FROM projects WHERE id = $1`, in.ProjectID).Scan(&identifier)
		return recordWorkitemEvent(ctx, tx, "workitem.created", in.WorkspaceID, in.ProjectID, reqID, TypeRequirement,
			in.CreatedBy, identifier+"-"+strconv.FormatInt(seqID, 10), in.Name, in.Assignees, "", "")
	})
	return reqID, err
}

func (s *RequirementService) getByIDTx(ctx context.Context, tx pgx.Tx, reqID, wsID int64) (*Requirement, error) {
	var r Requirement
	var parentID sql.NullInt64
	err := tx.QueryRow(ctx, `
		SELECT id, workspace_id, project_id, sequence_id, 'requirement'::text, parent_id, depth,
		       name, state_id, priority, version, created_by
		FROM requirement WHERE id = $1 AND workspace_id = $2 AND deleted_at IS NULL`, reqID, wsID).Scan(
		&r.ID, &r.WorkspaceID, &r.ProjectID, &r.SequenceID,
		&r.TypeCode, &parentID, &r.Depth,
		&r.Name, &r.StateID, &r.Priority, &r.Version, &r.CreatedBy)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrNotFound
		}
		return nil, errs.ErrInternal.Wrap(err)
	}
	if parentID.Valid {
		v := parentID.Int64
		r.ParentID = &v
	}
	return &r, nil
}

func (s *RequirementService) nextSequenceID(ctx context.Context, projectID int64) (int64, error) {
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

func (s *RequirementService) withTx(ctx context.Context, wsID int64, fn func(tx pgx.Tx) error) error {
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

func (s *RequirementService) updateM2M(ctx context.Context, tx pgx.Tx, reqID int64, assignees, labels, modules []int64) error {
	return sharedUpdateM2M(ctx, tx, reqID, assignees, labels, modules)
}

func buildRequirementUpdateSet(in UpdateRequirementInput) ([]string, []interface{}) {
	var sets []string
	var args []interface{}
	arg := 1
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
	if in.Source != nil {
		sets = append(sets, "source = $"+strconv.Itoa(arg))
		args = append(args, *in.Source)
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
	return sets, args
}

func mapPgRequirementErr(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgErr.ConstraintName == "requirement_project_id_sequence_id_key" {
			return errs.New("REQUIREMENT.DUPLICATE_SEQ", "需求序号冲突，请重试", 409)
		}
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return errs.ErrNotFound
	}
	return errs.ErrInternal.Wrap(err)
}
