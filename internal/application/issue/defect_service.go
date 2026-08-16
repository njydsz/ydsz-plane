// Package issue — Defect 聚合根应用服务（defect 表独立 CRUD）。
package issue

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// DefectService 提供 Defect 领域应用服务。
type DefectService struct {
	db       *pgxpool.Pool
	stateSvc *StateService
}

// NewDefectService 创建 Defect 服务。
func NewDefectService(db *pgxpool.Pool) *DefectService {
	return &DefectService{db: db, stateSvc: NewStateService(db)}
}

// Create 创建缺陷。
func (s *DefectService) Create(ctx context.Context, in CreateDefectInput) (*Defect, error) {
	if in.Priority == "" {
		in.Priority = PriorityNone
	}
	if err := validateWorkitemName(in.Name); err != nil {
		return nil, err
	}
	if strings.TrimSpace(in.Name) == "" {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "name", Reason: "缺陷名称不能为空"})
	}
	if in.Severity < 1 || in.Severity > 5 {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "severity", Reason: "缺陷严重程度为必填（1-5）"})
	}
	if in.FoundPhase == "" {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "found_phase", Reason: "缺陷发现阶段为必填"})
	}

	if in.ParentID != nil {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "parent_id", Reason: "缺陷暂不支持父子层级"})
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

	defectID, err := s.insertDefect(ctx, in, stateID, seqID)
	if err != nil {
		return nil, err
	}

	return s.GetByID(ctx, in.WorkspaceID, defectID)
}

// GetByID 获取缺陷详情。
func (s *DefectService) GetByID(ctx context.Context, wsID, defectID int64) (*Defect, error) {
	var d Defect
	var parentID sql.NullInt64
	var foundVerID, fixVerID sql.NullInt64
	var point sql.NullInt64
	var startDate, targetDate, completedAt sql.NullTime
	var stateName, stateColor, identifier string
	var stateGroup StateGroup
	var sprintID sql.NullInt64
	var versionID sql.NullInt64

	err := s.db.QueryRow(ctx, `
		SELECT d.id, d.public_id, d.workspace_id, d.project_id, d.sequence_id,
		       'defect'::text, d.parent_id, d.depth, d.name,
		       d.description_json, d.description_html,
		       d.state_id, s.name, s.color, s."group",
		       d.priority, d.severity, d.found_phase, d.point,
		       d.start_date, d.target_date, d.completed_at, d.progress,
		       d.is_draft, d.version, d.created_by, d.created_at, d.updated_at,
		       p.identifier,
		       d.sprint_id, d.version_id,
		       d.found_version_id, d.fix_version_id
		FROM defect d
		JOIN states s ON s.id = d.state_id
		JOIN projects p ON p.id = d.project_id
		WHERE d.id = $1 AND d.workspace_id = $2 AND d.deleted = false`,
		defectID, wsID).Scan(
		&d.ID, &d.PublicID, &d.WorkspaceID, &d.ProjectID, &d.SequenceID,
		&d.TypeCode, &parentID, &d.Depth, &d.Name,
		&d.DescriptionJSON, &d.DescriptionHTML,
		&d.StateID, &stateName, &stateColor, &stateGroup,
		&d.Priority, &d.Severity, &d.FoundPhase, &point,
		&d.StartDate, &targetDate, &completedAt, &d.Progress,
		&d.IsDraft, &d.Version, &d.CreatedBy, &d.CreatedAt, &d.UpdatedAt,
		&identifier, &sprintID, &versionID,
		&foundVerID, &fixVerID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrNotFound
		}
		return nil, errs.ErrInternal.Wrap(err)
	}

	d.Identifier = identifier + "-" + strconv.FormatInt(d.SequenceID, 10)
	d.State = &State{ID: d.StateID, Name: stateName, Color: stateColor, Group: stateGroup}
	if parentID.Valid {
		v := parentID.Int64
		d.ParentID = &v
	}
	if point.Valid {
		v := int(point.Int64)
		d.Point = &v
	}
	if startDate.Valid {
		d.StartDate = &startDate.Time
	}
	if targetDate.Valid {
		d.TargetDate = &targetDate.Time
	}
	if completedAt.Valid {
		d.CompletedAt = &completedAt.Time
	}
	if sprintID.Valid {
		v := sprintID.Int64
		d.SprintID = &v
	}
	if versionID.Valid {
		v := versionID.Int64
		d.VersionID = &v
	}
	if foundVerID.Valid {
		v := foundVerID.Int64
		d.FoundVersionID = &v
	}
	if fixVerID.Valid {
		v := fixVerID.Int64
		d.FixVersionID = &v
	}
	d.Assignees, _ = loadIntArray(ctx, s.db, `SELECT user_id FROM defect_assignees WHERE defect_id = $1`, defectID)
	d.Labels, _ = loadIntArray(ctx, s.db, `SELECT label_id FROM defect_labels WHERE defect_id = $1`, defectID)
	d.Modules, _ = loadIntArray(ctx, s.db, `SELECT module_id FROM defect_modules WHERE defect_id = $1`, defectID)
	d.Watchers, _ = loadIntArray(ctx, s.db, `SELECT user_id FROM defect_watchers WHERE defect_id = $1`, defectID)

	return &d, nil
}

// Update 更新缺陷。
func (s *DefectService) Update(ctx context.Context, wsID, defectID int64, in UpdateDefectInput) (*Defect, error) {
	err := s.withTx(ctx, wsID, func(tx pgx.Tx) error {
		current, err := s.getByIDTx(ctx, tx, defectID, wsID)
		if err != nil {
			return err
		}
		if in.Version != current.Version {
			return errs.ErrVersionConflict
		}
		if in.ParentID != nil && *in.ParentID != defectID {
			return errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "parent_id", Reason: "缺陷暂不支持父子层级"})
		}

		sets, args := buildDefectUpdateSet(in)
		if len(sets) == 0 {
			return s.updateM2M(ctx, tx, wsID, current.ProjectID, defectID, in.Assignees, in.Labels, in.Modules)
		}

		sets = append(sets, "updated_at = now()")
		query := fmt.Sprintf(`UPDATE defect SET %s WHERE id = $%d AND workspace_id = $%d AND version = $%d AND deleted = false`,
			strings.Join(sets, ", "), len(args)+1, len(args)+2, len(args)+3)
		args = append(args, defectID, wsID, in.Version)

		tag, err := tx.Exec(ctx, query, args...)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return errs.ErrVersionConflict
		}
		return s.updateM2M(ctx, tx, wsID, current.ProjectID, defectID, in.Assignees, in.Labels, in.Modules)
	})
	if err != nil {
		return nil, err
	}
	return s.GetByID(ctx, wsID, defectID)
}

// SoftDelete 归档。
func (s *DefectService) SoftDelete(ctx context.Context, wsID, defectID int64) error {
	return s.withTx(ctx, wsID, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `DELETE FROM sprint_defects WHERE defect_id = $1`, defectID); err != nil {
			return err
		}
		tag, err := tx.Exec(ctx, `
			UPDATE defect SET deleted = true, updated_at = now()
			WHERE id = $1 AND workspace_id = $2 AND deleted = false`, defectID, wsID)
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
func (s *DefectService) Restore(ctx context.Context, wsID, defectID int64) error {
	tag, err := s.db.Exec(ctx, `
		UPDATE defect SET deleted = false, updated_at = now()
		WHERE id = $1 AND workspace_id = $2 AND deleted = true`, defectID, wsID)
	if err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	if tag.RowsAffected() == 0 {
		return errs.ErrNotFound
	}
	return nil
}

// Transition 执行状态流转。缺陷完成时需要额外校验 root_cause_category + fix_version_id。
func (s *DefectService) Transition(ctx context.Context, wsID, projectID, defectID, toStateID, userID int64) (*Defect, error) {
	err := s.withTx(ctx, wsID, func(tx pgx.Tx) error {
		d, err := s.getByIDTx(ctx, tx, defectID, wsID)
		if err != nil {
			return err
		}
		if d.StateID == toStateID {
			return nil
		}

		toGroup, err := s.stateSvc.StateGroupByID(ctx, toStateID)
		if err != nil {
			return err
		}
		if err := s.stateSvc.ValidateTransition(ctx, wsID, projectID, TransitionInput{
			IssueID: defectID, FromState: d.StateID, ToState: toStateID, TypeCode: d.TypeCode,
			Context: TransitionContext{RootCauseCategory: d.RootCauseCategory, FixVersionID: d.FixVersionID},
		}); err != nil {
			return err
		}

		// 缺陷完成时必填校验
		if toGroup == GroupCompleted {
			var rc sql.NullString
			var fv sql.NullInt64
			if scanErr := tx.QueryRow(ctx,
				`SELECT root_cause_category, fix_version_id FROM defect WHERE id = $1 AND workspace_id = $2`,
				defectID, wsID).Scan(&rc, &fv); scanErr != nil {
				return errs.ErrInternal.Wrap(scanErr)
			}
			if !rc.Valid || rc.String == "" {
				return errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "root_cause_category", Reason: "缺陷关闭时根因分类为必填"})
			}
			if !fv.Valid {
				return errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "fix_version_id", Reason: "缺陷关闭时修复版本为必填"})
			}
		}

		completedAtClause := "NULL"
		progress := d.Progress
		if toGroup == GroupCompleted {
			completedAtClause = "now()"
			progress = 100
		}
		query := fmt.Sprintf(`UPDATE defect SET state_id = $1, completed_at = %s,
			progress = $2, version = version + 1, updated_at = now()
			WHERE id = $3 AND workspace_id = $4 AND deleted = false`, completedAtClause)
		if _, err := tx.Exec(ctx, query, toStateID, progress, defectID, wsID); err != nil {
			return err
		}
		d.StateID = toStateID

		assignees := loadIntArrayTx(ctx, tx, `SELECT user_id FROM defect_assignees WHERE defect_id = $1`, defectID)
		var identifier, defectName string
		_ = tx.QueryRow(ctx, `SELECT p.identifier, d.name
			FROM defect d JOIN projects p ON p.id = d.project_id
			WHERE d.id = $1`, defectID).Scan(&identifier, &defectName)
		actorName := getUserNameTx(ctx, tx, userID)
		return recordWorkitemEvent(ctx, tx, "workitem.status_changed", wsID, projectID, defectID, TypeDefect,
			userID, actorName, identifier, defectName, assignees, loadStateName(ctx, tx, d.StateID), loadStateName(ctx, tx, toStateID))
	})
	if err != nil {
		return nil, err
	}
	return s.GetByID(ctx, wsID, defectID)
}

// --- 内部方法 ---

func (s *DefectService) insertDefect(ctx context.Context, in CreateDefectInput, stateID, seqID int64) (int64, error) {
	var defectID int64
	err := s.withTx(ctx, in.WorkspaceID, func(tx pgx.Tx) error {
		depth := 1
		var sourceVer interface{}
		if in.SourceVersionID != nil {
			sourceVer = *in.SourceVersionID
		}
		var repro interface{}
		if in.ReproduceSteps != nil {
			repro, _ = json.Marshal(in.ReproduceSteps)
		}
		var env interface{}
		if in.Environment != nil {
			env, _ = json.Marshal(in.Environment)
		}

		err := tx.QueryRow(ctx, `
			INSERT INTO defect (workspace_id, project_id, sequence_id, parent_id, depth,
				name, description_json, description_html, state_id, priority,
				severity, found_phase, found_version_id, reproduce_steps, environment,
				point, start_date, target_date, is_draft, created_by)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20)
			RETURNING id`,
			in.WorkspaceID, in.ProjectID, seqID, in.ParentID, depth,
			in.Name, nil, in.DescriptionHTML, stateID, string(in.Priority),
			in.Severity, in.FoundPhase, sourceVer, repro, env,
			in.Point, in.StartDate, in.TargetDate, in.IsDraft, in.CreatedBy).Scan(&defectID)
		if err != nil {
			return mapPgDefectErr(err)
		}
		if err := insertM2M(ctx, tx, TypeDefect, in.WorkspaceID, in.ProjectID, defectID, in.Assignees, in.Labels, in.Modules); err != nil {
			return err
		}
		var identifier string
		_ = tx.QueryRow(ctx, `SELECT identifier FROM projects WHERE id = $1`, in.ProjectID).Scan(&identifier)
		actorName := getUserNameTx(ctx, tx, in.CreatedBy)
		return recordWorkitemEvent(ctx, tx, "workitem.created", in.WorkspaceID, in.ProjectID, defectID, TypeDefect,
			in.CreatedBy, actorName, identifier+"-"+strconv.FormatInt(seqID, 10), in.Name, in.Assignees, "", "")
	})
	return defectID, err
}

func (s *DefectService) getByIDTx(ctx context.Context, tx pgx.Tx, defectID, wsID int64) (*Defect, error) {
	var d Defect
	var parentID sql.NullInt64
	err := tx.QueryRow(ctx, `
		SELECT id, workspace_id, project_id, sequence_id, 'defect'::text, parent_id, depth,
		       name, state_id, priority, severity, found_phase, version, created_by
		FROM defect WHERE id = $1 AND workspace_id = $2 AND deleted = false`, defectID, wsID).Scan(
		&d.ID, &d.WorkspaceID, &d.ProjectID, &d.SequenceID,
		&d.TypeCode, &parentID, &d.Depth,
		&d.Name, &d.StateID, &d.Priority, &d.Severity, &d.FoundPhase, &d.Version, &d.CreatedBy)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrNotFound
		}
		return nil, errs.ErrInternal.Wrap(err)
	}
	if parentID.Valid {
		v := parentID.Int64
		d.ParentID = &v
	}
	return &d, nil
}

func (s *DefectService) nextSequenceID(ctx context.Context, projectID int64) (int64, error) {
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

func (s *DefectService) withTx(ctx context.Context, wsID int64, fn func(tx pgx.Tx) error) error {
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

func (s *DefectService) updateM2M(ctx context.Context, tx pgx.Tx, wsID, projectID, defectID int64, assignees, labels, modules []int64) error {
	return sharedUpdateM2M(ctx, tx, TypeDefect, wsID, projectID, defectID, assignees, labels, modules)
}

func buildDefectUpdateSet(in UpdateDefectInput) ([]string, []interface{}) {
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
	return sets, args
}

func mapPgDefectErr(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgErr.ConstraintName == "defect_project_id_sequence_id_key" {
			return errs.New("DEFECT.DUPLICATE_SEQ", "缺陷序号冲突，请重试", 409)
		}
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return errs.ErrNotFound
	}
	return errs.ErrInternal.Wrap(err)
}

// ImportServiceCreateIssueInput 兼容旧 import 流程，aggregated 入口转换为 CreateDefectInput 等。
type importAdapter struct{}
