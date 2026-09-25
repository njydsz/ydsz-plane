package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/njydsz/ydsz-plane/internal/infrastructure/persistence"
	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// ---------------------------------------------------------------------------
// Task 子表结构
// ---------------------------------------------------------------------------

// TaskRelation 任务关联关系（对应 task_relations 表）。
type TaskRelation struct {
	ID           int64  `json:"id"`
	SourceTaskID int64  `json:"source_task_id"`
	TargetTaskID int64  `json:"target_task_id"`
	RelationType string `json:"relation_type"`
}

// TaskTag 任务标签（对应 task_labels 表）。
type TaskTag struct {
	TaskID  int64 `json:"task_id"`
	LabelID int64 `json:"label_id"`
}

// Task entity — 独立聚合根，字段与 models.go 的 Task 结构对齐。
// 在 repositories 包中内联定义，避免反向依赖 issue 包。
type Task struct {
	ID              int64          `json:"id"`
	PublicID        string         `json:"public_id"`
	WorkspaceID     int64          `json:"workspace_id"`
	ProjectID       int64          `json:"project_id"`
	SequenceID      int64          `json:"sequence_id"`
	Identifier      string         `json:"identifier"`
	TypeCode        string         `json:"type_code"`
	ParentID        *int64         `json:"parent_id,omitempty"`
	Depth           int            `json:"depth"`
	Name            string         `json:"name"`
	DescriptionJSON map[string]any `json:"description_json,omitempty"`
	DescriptionHTML string         `json:"description_html,omitempty"`
	StateID         int64          `json:"state_id"`
	StateName       string         `json:"state_name,omitempty"`
	StateGroup      string         `json:"state_group,omitempty"`
	StateColor      string         `json:"state_color,omitempty"`
	Priority        string         `json:"priority"`
	Point           *int           `json:"point,omitempty"`
	SprintID        *int64         `json:"sprint_id,omitempty"`
	VersionID       *int64         `json:"version_id,omitempty"`
	Progress        int            `json:"progress"`
	StartDate       *string        `json:"start_date,omitempty"`
	TargetDate      *string        `json:"target_date,omitempty"`
	CompletedAt     *string        `json:"completed_at,omitempty"`
	IsDraft         bool           `json:"is_draft"`
	SortOrder       float64        `json:"sort_order"`
	Version         int            `json:"version"`
	Assignees       []int64        `json:"assignees,omitempty"`
	Labels          []int64        `json:"labels,omitempty"`
	Modules         []int64        `json:"modules,omitempty"`
	Watchers        []int64        `json:"watchers,omitempty"`
	CreatedBy       int64          `json:"created_by"`
	CreatedAt       string         `json:"created_at"`
	UpdatedAt       string         `json:"updated_at"`
	// 任务专属字段
	Category        *string  `json:"category,omitempty"`
	ActualEffort    *float64 `json:"actual_effort,omitempty"`
	RemainingEffort *float64 `json:"remaining_effort,omitempty"`
	DelayReason     *string  `json:"delay_reason,omitempty"`
}

// ---------------------------------------------------------------------------
// TaskRepository 接口
// ---------------------------------------------------------------------------

// TaskRepository 定义 Task 聚合根的仓储接口。
//
// 联动 task 主表 + task_relations（关联关系）+ task_labels（标签）两个子表。
// 写操作在主表变更的同事务内同步子表，保证一致性。
type TaskRepository interface {
	// GetByID 按 ID 查询任务（含子表加载：assignees/labels/modules/watchers + relations）
	GetByID(ctx context.Context, id int64) (*Task, error)

	// Create 创建任务（含子表写入）
	Create(ctx context.Context, in CreateTaskInput) (*Task, error)

	// Update 更新任务（含 relations / labels / assignees / modules 子表联动）
	Update(ctx context.Context, wsID int64, t *Task, relations []TaskRelation) error

	// Delete 软删除任务（同时清理 sprint_tasks / task_relations / task_labels）
	Delete(ctx context.Context, wsID, id int64) (bool, error)

	// ListByWorkspace 列出工作空间下任务
	ListByWorkspace(ctx context.Context, wsID int64, opts ListOptions) ([]Task, error)

	// ListByProject 列出项目内任务
	ListByProject(ctx context.Context, wsID, projectID int64, opts ListOptions) ([]Task, error)
}

// CreateTaskInput 创建任务入参。
type CreateTaskInput struct {
	WorkspaceID     int64
	ProjectID       int64
	Name            string
	DescriptionHTML string
	StateID         int64
	Priority        string
	ParentID        *int64
	Category        *string
	Assignees       []int64
	Labels          []int64
	Modules         []int64
	Watchers        []int64
	Point           *int
	StartDate       *string
	TargetDate      *string
	SprintID        *int64
	VersionID       *int64
	IsDraft         bool
	CreatedBy       int64
}

// ---------------------------------------------------------------------------
// PostgreSQL 实现
// ---------------------------------------------------------------------------

// postgresTaskRepository 是 TaskRepository 的 PostgreSQL 实现。
//
// 嵌入 persistence.PostgresRepo 以复用 WithTenantTx / DB() 等基础能力。
type postgresTaskRepository struct {
	persistence.PostgresRepo
}

// NewTaskRepository 构造 Task 仓储实例。
func NewTaskRepository(db *pgxpool.Pool) TaskRepository {
	return &postgresTaskRepository{PostgresRepo: persistence.NewPostgresRepo(db)}
}

// withTx 在租户事务内执行 fn（自动 set_config app.workspace_id）。
func (r *postgresTaskRepository) withTx(ctx context.Context, wsID int64, fn func(tx pgx.Tx) error) error {
	return r.WithTenantTx(ctx, wsID, fn)
}

// GetByID 按 ID 查询任务详情，含 M2M 子表加载。
func (r *postgresTaskRepository) GetByID(ctx context.Context, id int64) (*Task, error) {
	var t Task
	var parentID sql.NullInt64
	var category sql.NullString
	var point sql.NullInt64
	var startDate, targetDate, completedAt sql.NullString
	var stateName, stateColor, identifier, stateGroup string
	var sprintID sql.NullInt64
	var versionID sql.NullInt64

	err := r.DB().QueryRow(ctx, `
		SELECT t.id, t.public_id, t.workspace_id, t.project_id, t.sequence_id,
		       'task'::text, t.parent_id, t.depth, t.name,
		       t.description_json, t.description_html,
		       t.state_id, s.name, s.color, s."group",
		       t.priority, t.category, t.point,
		       t.start_date::text, t.target_date::text, t.completed_at::text, t.progress,
		       t.is_draft, t.version, t.created_by, t.created_at::text, t.updated_at::text,
		       p.identifier, t.sprint_id, t.version_id
		FROM task t
		JOIN states s ON s.id = t.state_id
		JOIN projects p ON p.id = t.project_id
		WHERE t.id = $1 AND t.deleted = false`, id).Scan(
		&t.ID, &t.PublicID, &t.WorkspaceID, &t.ProjectID, &t.SequenceID,
		&t.TypeCode, &parentID, &t.Depth, &t.Name,
		&t.DescriptionJSON, &t.DescriptionHTML,
		&t.StateID, &stateName, &stateColor, &stateGroup,
		&t.Priority, &category, &point,
		&t.StartDate, &targetDate, &completedAt, &t.Progress,
		&t.IsDraft, &t.Version, &t.CreatedBy, &t.CreatedAt, &t.UpdatedAt,
		&identifier, &sprintID, &versionID)
	if err != nil {
		return nil, mapNotFound(err)
	}

	t.Identifier = identifier + "-" + strconv.FormatInt(t.SequenceID, 10)
	t.StateName = stateName
	t.StateColor = stateColor
	t.StateGroup = stateGroup
	if parentID.Valid {
		v := parentID.Int64
		t.ParentID = &v
	}
	if category.Valid {
		v := category.String
		t.Category = &v
	}
	if point.Valid {
		v := int(point.Int64)
		t.Point = &v
	}
	if startDate.Valid {
		v := startDate.String
		t.StartDate = &v
	}
	if targetDate.Valid {
		v := targetDate.String
		t.TargetDate = &v
	}
	if completedAt.Valid {
		v := completedAt.String
		t.CompletedAt = &v
	}
	if sprintID.Valid {
		v := sprintID.Int64
		t.SprintID = &v
	}
	if versionID.Valid {
		v := versionID.Int64
		t.VersionID = &v
	}

	db := r.DB()
	t.Assignees, _ = loadIntArray(ctx, db, `SELECT user_id FROM task_assignees WHERE task_id = $1`, id)
	t.Labels, _ = loadIntArray(ctx, db, `SELECT label_id FROM task_labels WHERE task_id = $1`, id)
	t.Modules, _ = loadIntArray(ctx, db, `SELECT module_id FROM task_modules WHERE task_id = $1`, id)
	t.Watchers, _ = loadIntArray(ctx, db, `SELECT user_id FROM task_watchers WHERE task_id = $1`, id)
	return &t, nil
}

// Create 创建任务（含 M2M 子表写入）。
func (r *postgresTaskRepository) Create(ctx context.Context, in CreateTaskInput) (*Task, error) {
	var taskID int64
	err := r.withTx(ctx, in.WorkspaceID, func(tx pgx.Tx) error {
		seqID, err := nextSequenceIDTx(ctx, tx, in.ProjectID)
		if err != nil {
			return err
		}
		depth := 1
		err = tx.QueryRow(ctx, `
			INSERT INTO task (workspace_id, project_id, sequence_id, parent_id, depth,
				name, description_json, description_html, state_id, priority,
				category, point, start_date, target_date, sprint_id, version_id, is_draft, created_by)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18)
			RETURNING id`,
			in.WorkspaceID, in.ProjectID, seqID, in.ParentID, depth,
			in.Name, nil, in.DescriptionHTML, in.StateID, in.Priority,
			in.Category, in.Point, in.StartDate, in.TargetDate,
			in.SprintID, in.VersionID, in.IsDraft, in.CreatedBy).Scan(&taskID)
		if err != nil {
			return mapPgError(err, "task_project_id_sequence_id_key", "TASK.DUPLICATE_SEQ", "任务序号冲突，请重试")
		}
		return insertTaskM2M(ctx, tx, in.WorkspaceID, in.ProjectID, taskID,
			in.Assignees, in.Labels, in.Modules, in.Watchers)
	})
	if err != nil {
		return nil, err
	}
	return r.GetByID(ctx, taskID)
}

// Update 更新任务（含 relations / tags / assignees / modules 子表联动）。
//
// 语义：全字段写入 + 乐观锁（version 必须与当前 DB 值一致）。
// relations 为 nil 时不操作 task_relations；非 nil 时覆盖写入（先 DELETE 后 INSERT）。
// t.Labels / t.Assignees / t.Modules 为 nil 时不操作对应子表；非 nil 时覆盖写入。
// 这三个切片由 GetByID 完整填充，典型调用方：GetByID → 修改字段 → Update。
func (r *postgresTaskRepository) Update(ctx context.Context, wsID int64, t *Task, relations []TaskRelation) error {
	return r.withTx(ctx, wsID, func(tx pgx.Tx) error {
		// 乐观锁检查
		var curVer int
		if err := tx.QueryRow(ctx,
			`SELECT version FROM task WHERE id = $1 AND workspace_id = $2 AND deleted = false`,
			t.ID, wsID).Scan(&curVer); err != nil {
			return mapNotFound(err)
		}
		if curVer != t.Version {
			return errs.ErrVersionConflict
		}

		// 主表字段全量写入（固定占位符 $1..$15，WHERE 使用 $16/$17）
		sets, args := buildTaskUpdateSets(t)
		sets = append(sets, "updated_at = now()", "version = version + 1")
		args = append(args, t.ID, wsID)
		query := "UPDATE task SET " + strings.Join(sets, ", ") +
			" WHERE id = $16 AND workspace_id = $17 AND deleted = false"
		tag, err := tx.Exec(ctx, query, args...)
		if err != nil {
			return errs.ErrInternal.Wrap(err)
		}
		if tag.RowsAffected() == 0 {
			return errs.ErrNotFound
		}

		// task_labels (tags) 覆盖写入
		if t.Labels != nil {
			if _, err := tx.Exec(ctx, `DELETE FROM task_labels WHERE task_id = $1`, t.ID); err != nil {
				return errs.ErrInternal.Wrap(err)
			}
			for _, lid := range t.Labels {
				if _, err := tx.Exec(ctx,
					`INSERT INTO task_labels (workspace_id, project_id, task_id, label_id) VALUES ($1,$2,$3,$4) ON CONFLICT DO NOTHING`,
					wsID, t.ProjectID, t.ID, lid); err != nil {
					return errs.ErrInternal.Wrap(err)
				}
			}
		}

		// task_assignees 覆盖写入
		if t.Assignees != nil {
			if _, err := tx.Exec(ctx, `DELETE FROM task_assignees WHERE task_id = $1`, t.ID); err != nil {
				return errs.ErrInternal.Wrap(err)
			}
			for _, uid := range t.Assignees {
				if _, err := tx.Exec(ctx,
					`INSERT INTO task_assignees (workspace_id, project_id, task_id, user_id) VALUES ($1,$2,$3,$4) ON CONFLICT DO NOTHING`,
					wsID, t.ProjectID, t.ID, uid); err != nil {
					return errs.ErrInternal.Wrap(err)
				}
			}
		}

		// task_modules 覆盖写入
		if t.Modules != nil {
			if _, err := tx.Exec(ctx, `DELETE FROM task_modules WHERE task_id = $1`, t.ID); err != nil {
				return errs.ErrInternal.Wrap(err)
			}
			for _, mid := range t.Modules {
				if _, err := tx.Exec(ctx,
					`INSERT INTO task_modules (workspace_id, project_id, task_id, module_id) VALUES ($1,$2,$3,$4) ON CONFLICT DO NOTHING`,
					wsID, t.ProjectID, t.ID, mid); err != nil {
					return errs.ErrInternal.Wrap(err)
				}
			}
		}

		// task_relations 覆盖写入（入参非 nil 时删除旧关联再批量Insert）
		if relations != nil {
			if _, err := tx.Exec(ctx,
				`DELETE FROM task_relations WHERE source_task_id = $1 OR target_task_id = $1`, t.ID); err != nil {
				return errs.ErrInternal.Wrap(err)
			}
			for _, rel := range relations {
				if _, err := tx.Exec(ctx,
					`INSERT INTO task_relations (workspace_id, project_id, source_task_id, target_task_id, relation_type)
					 VALUES ($1,$2,$3,$4,$5)`,
					wsID, t.ProjectID, t.ID, rel.TargetTaskID, rel.RelationType); err != nil {
					return errs.ErrInternal.Wrap(err)
				}
			}
		}

		return nil
	})
}

// Delete 软删除任务（含 sprint_tasks / task_relations / task_labels 关联清理）。
func (r *postgresTaskRepository) Delete(ctx context.Context, wsID, id int64) (bool, error) {
	ok := false
	err := r.withTx(ctx, wsID, func(tx pgx.Tx) error {
		// 清理冲刺关联
		if _, err := tx.Exec(ctx, `DELETE FROM sprint_tasks WHERE task_id = $1`, id); err != nil {
			return errs.ErrInternal.Wrap(err)
		}
		// 清理 task_relations（双向）
		if _, err := tx.Exec(ctx, `DELETE FROM task_relations WHERE source_task_id = $1 OR target_task_id = $1`, id); err != nil {
			return errs.ErrInternal.Wrap(err)
		}
		// 主表软删除
		tag, err := tx.Exec(ctx, `
			UPDATE task SET deleted = true, updated_at = now()
			WHERE id = $1 AND workspace_id = $2 AND deleted = false`, id, wsID)
		if err != nil {
			return errs.ErrInternal.Wrap(err)
		}
		if tag.RowsAffected() == 0 {
			return errs.ErrNotFound
		}
		ok = true
		return nil
	})
	return ok, err
}

// ListByWorkspace 列出工作空间下任务。
func (r *postgresTaskRepository) ListByWorkspace(ctx context.Context, wsID int64, opts ListOptions) ([]Task, error) {
	return r.list(ctx, wsID, 0, opts)
}

// ListByProject 列出项目内任务。
func (r *postgresTaskRepository) ListByProject(ctx context.Context, wsID, projectID int64, opts ListOptions) ([]Task, error) {
	return r.list(ctx, wsID, projectID, opts)
}

// list 共用查询实现。
func (r *postgresTaskRepository) list(ctx context.Context, wsID, projectID int64, opts ListOptions) ([]Task, error) {
	opts.normalize()
	clauses := []string{"t.workspace_id = $1", "t.deleted = false"}
	args := []interface{}{wsID}
	arg := 2
	if projectID != 0 {
		clauses = append(clauses, "t.project_id = $"+strconv.Itoa(arg))
		args = append(args, projectID)
		arg++
	}
	if opts.StateID != nil {
		clauses = append(clauses, "t.state_id = $"+strconv.Itoa(arg))
		args = append(args, *opts.StateID)
		arg++
	}
	if opts.Priority != "" {
		clauses = append(clauses, "t.priority = $"+strconv.Itoa(arg))
		args = append(args, opts.Priority)
		arg++
	}
	if opts.ParentID != nil {
		clauses = append(clauses, "t.parent_id = $"+strconv.Itoa(arg))
		args = append(args, *opts.ParentID)
		arg++
	}
	if opts.Search != "" {
		clauses = append(clauses, "t.name ILIKE $"+strconv.Itoa(arg))
		args = append(args, "%"+opts.Search+"%")
		arg++
	}

	dir := "ASC"
	if opts.SortDesc {
		dir = "DESC"
	}
	var orderBy string
	switch opts.SortBy {
	case "priority":
		orderBy = `CASE t.priority WHEN 'urgent' THEN 1 WHEN 'high' THEN 2 WHEN 'medium' THEN 3 WHEN 'low' THEN 4 ELSE 5 END ` + dir
	case "target_date":
		orderBy = "t.target_date " + dir
	case "created_at":
		orderBy = "t.created_at " + dir
	case "sequence":
		orderBy = "t.sequence_id " + dir
	default:
		orderBy = "t.updated_at " + dir
	}

	limitIdx := len(args) + 1
	offsetIdx := len(args) + 2
	args = append(args, opts.Limit, opts.Offset)

	query := fmt.Sprintf(`
		SELECT t.id, t.public_id, t.workspace_id, t.project_id, t.sequence_id,
		       'task'::text, t.parent_id, t.depth, t.name,
		       t.description_json, t.description_html,
		       t.state_id, s.name, s.color, s."group",
		       t.priority, t.category, t.point,
		       t.start_date::text, t.target_date::text, t.completed_at::text, t.progress,
		       t.is_draft, t.version, t.created_by, t.created_at::text, t.updated_at::text,
		       p.identifier, t.sprint_id, t.version_id
		FROM task t
		JOIN states s ON s.id = t.state_id
		JOIN projects p ON p.id = t.project_id
		WHERE %s
		ORDER BY %s
		LIMIT $%d OFFSET $%d`,
		strings.Join(clauses, " AND "), orderBy, limitIdx, offsetIdx)

	rows, err := r.DB().Query(ctx, query, args...)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()

	db := r.DB()
	var tasks []Task
	for rows.Next() {
		var t Task
		var parentID sql.NullInt64
		var category sql.NullString
		var point sql.NullInt64
		var startDate, targetDate, completedAt sql.NullString
		var stateName, stateColor, identifier, stateGroup string
		var sprintID sql.NullInt64
		var versionID sql.NullInt64

		if err := rows.Scan(
			&t.ID, &t.PublicID, &t.WorkspaceID, &t.ProjectID, &t.SequenceID,
			&t.TypeCode, &parentID, &t.Depth, &t.Name,
			&t.DescriptionJSON, &t.DescriptionHTML,
			&t.StateID, &stateName, &stateColor, &stateGroup,
			&t.Priority, &category, &point,
			&t.StartDate, &targetDate, &completedAt, &t.Progress,
			&t.IsDraft, &t.Version, &t.CreatedBy, &t.CreatedAt, &t.UpdatedAt,
			&identifier, &sprintID, &versionID); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
		t.Identifier = identifier + "-" + strconv.FormatInt(t.SequenceID, 10)
		t.StateName = stateName
		t.StateColor = stateColor
		t.StateGroup = stateGroup
		if parentID.Valid {
			v := parentID.Int64
			t.ParentID = &v
		}
		if category.Valid {
			v := category.String
			t.Category = &v
		}
		if point.Valid {
			v := int(point.Int64)
			t.Point = &v
		}
		if startDate.Valid {
			v := startDate.String
			t.StartDate = &v
		}
		if targetDate.Valid {
			v := targetDate.String
			t.TargetDate = &v
		}
		if completedAt.Valid {
			v := completedAt.String
			t.CompletedAt = &v
		}
		if sprintID.Valid {
			v := sprintID.Int64
			t.SprintID = &v
		}
		if versionID.Valid {
			v := versionID.Int64
			t.VersionID = &v
		}
		t.Assignees, _ = loadIntArray(ctx, db, `SELECT user_id FROM task_assignees WHERE task_id = $1`, t.ID)
		t.Labels, _ = loadIntArray(ctx, db, `SELECT label_id FROM task_labels WHERE task_id = $1`, t.ID)
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}

// ---------------------------------------------------------------------------
// TaskRepository 内部辅助
// ---------------------------------------------------------------------------

// nextSequenceIDTx 在事务内获取下一个序列号。
func nextSequenceIDTx(ctx context.Context, tx pgx.Tx, projectID int64) (int64, error) {
	var seq int64
	err := tx.QueryRow(ctx, `
		INSERT INTO project_sequences (project_id, next_value) VALUES ($1, 2)
		ON CONFLICT (project_id) DO UPDATE SET next_value = project_sequences.next_value + 1
		RETURNING next_value - 1`, projectID).Scan(&seq)
	if err != nil {
		return 0, errs.ErrInternal.Wrap(err)
	}
	return seq, nil
}

// insertTaskM2M 写入任务 M2M 子表（assignees / labels / modules / watchers）。
func insertTaskM2M(ctx context.Context, tx pgx.Tx, wsID, projectID, taskID int64, assignees, labels, modules, watchers []int64) error {
	for _, uid := range assignees {
		if _, err := tx.Exec(ctx,
			`INSERT INTO task_assignees (workspace_id, project_id, task_id, user_id) VALUES ($1,$2,$3,$4) ON CONFLICT DO NOTHING`,
			wsID, projectID, taskID, uid); err != nil {
			return errs.ErrInternal.Wrap(err)
		}
	}
	for _, lid := range labels {
		if _, err := tx.Exec(ctx,
			`INSERT INTO task_labels (workspace_id, project_id, task_id, label_id) VALUES ($1,$2,$3,$4) ON CONFLICT DO NOTHING`,
			wsID, projectID, taskID, lid); err != nil {
			return errs.ErrInternal.Wrap(err)
		}
	}
	for _, mid := range modules {
		if _, err := tx.Exec(ctx,
			`INSERT INTO task_modules (workspace_id, project_id, task_id, module_id) VALUES ($1,$2,$3,$4) ON CONFLICT DO NOTHING`,
			wsID, projectID, taskID, mid); err != nil {
			return errs.ErrInternal.Wrap(err)
		}
	}
	for _, wid := range watchers {
		if _, err := tx.Exec(ctx,
			`INSERT INTO task_watchers (workspace_id, project_id, task_id, user_id) VALUES ($1,$2,$3,$4) ON CONFLICT DO NOTHING`,
			wsID, projectID, taskID, wid); err != nil {
			return errs.ErrInternal.Wrap(err)
		}
	}
	return nil
}

// buildTaskUpdateSets 生成全量写入 task 主表的非键字段 SET 子句与参数。
// version 等乐观锁字段由 Update 核心逻辑自动递增，不在此处写入。
func buildTaskUpdateSets(t *Task) ([]string, []any) {
	sets := []string{
		"name = $1",
		"description_html = $2",
		"description_json = $3",
		"priority = $4",
		"category = $5",
		"point = $6",
		"state_id = $7",
		"start_date = $8",
		"target_date = $9",
		"progress = $10",
		"delay_reason = $11",
		"sprint_id = $12",
		"version_id = $13",
		"parent_id = $14",
		"is_draft = $15",
	}
	args := []any{
		t.Name,
		t.DescriptionHTML,
		t.DescriptionJSON,
		t.Priority,
		t.Category,
		t.Point,
		t.StateID,
		t.StartDate,
		t.TargetDate,
		t.Progress,
		t.DelayReason,
		t.SprintID,
		t.VersionID,
		t.ParentID,
		t.IsDraft,
	}
	return sets, args
}
