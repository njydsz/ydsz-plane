// Package repositories — Issue 与 Task 聚合根的 PostgreSQL 仓储实现。
//
// 遵循 persistence 包的 Repository 模式：
//   - 接口定义在调用方侧（repositories 子包内部），不依赖 domain 层。
//   - 实现使用 persistence.Querier 作为事务抽象，支持 *pgxpool.Pool 和 pgx.Tx。
//   - Issue 仓储封装跨三表（task / requirement / defect）的 UNION ALL 视图操作；
//     Task 仓储管理 task 主表与 task_relations / task_labels（tags）两个子表的联动。
//
// 参考:
//   - Martin Fowler《Patterns of Enterprise Application Architecture》— Repository
//   - persistence/repository.go 中的 BaseRepository 与 PostgresRepo 基类
package repositories

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

	"github.com/njydsz/ydsz-plane/internal/infrastructure/persistence"
	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// ---------------------------------------------------------------------------
// 共享查询选项与辅助
// ---------------------------------------------------------------------------

// ListOptions 跨类型列表查询选项。
type ListOptions struct {
	StateID    *int64
	Group      string
	TypeCode   string
	Priority   string
	ParentID   *int64
	Search     string
	SortBy     string
	SortDesc   bool
	Limit      int
	Offset     int
}

// normalize 填充 Limit / Offset 安全默认值。
func (o *ListOptions) normalize() {
	if o.Limit <= 0 || o.Limit > 100 {
		o.Limit = 50
	}
	if o.Offset < 0 {
		o.Offset = 0
	}
}

// buildWhere 拼装跨类型列表的 WHERE 子句；返回 (whereClause, args)。
// 参数索引从 startIdx 开始（调用方传入已有参数数量 + 1）。
func (o *ListOptions) buildWhere(wsID int64, projectID int64) (string, []interface{}) {
	clauses := []string{"1=1"}
	args := []interface{}{wsID}
	arg := 2
	if projectID != 0 {
		clauses = append(clauses, "i.project_id = $"+strconv.Itoa(arg))
		args = append(args, projectID)
		arg++
	}
	if o.StateID != nil {
		clauses = append(clauses, "i.state_id = $"+strconv.Itoa(arg))
		args = append(args, *o.StateID)
		arg++
	}
	if o.Group != "" {
		clauses = append(clauses, `s."group" = $`+strconv.Itoa(arg))
		args = append(args, o.Group)
		arg++
	}
	if o.TypeCode != "" {
		clauses = append(clauses, "i.type_code = $"+strconv.Itoa(arg))
		args = append(args, o.TypeCode)
		arg++
	}
	if o.Priority != "" {
		clauses = append(clauses, "i.priority = $"+strconv.Itoa(arg))
		args = append(args, o.Priority)
		arg++
	}
	if o.ParentID != nil {
		clauses = append(clauses, "i.parent_id = $"+strconv.Itoa(arg))
		args = append(args, *o.ParentID)
		arg++
	}
	if o.Search != "" {
		clauses = append(clauses, "i.name ILIKE $"+strconv.Itoa(arg))
		args = append(args, "%"+o.Search+"%")
		arg++
	}
	return "WHERE " + strings.Join(clauses, " AND "), args
}

// buildSort 返回跨类型列表的 ORDER BY 列表达式。
func (o *ListOptions) buildSort() string {
	dir := "ASC"
	if o.SortDesc {
		dir = "DESC"
	}
	switch o.SortBy {
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

// loadIntArray 在 Querier 上查询单列整数数组。
func loadIntArray(ctx context.Context, q persistence.Querier, query string, args ...interface{}) ([]int64, error) {
	rows, err := q.Query(ctx, query, args...)
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

// mapNotFound 将 pgx.ErrNoRows 统一转换为 errs.ErrNotFound。
func mapNotFound(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return errs.ErrNotFound
	}
	return errs.ErrInternal.Wrap(err)
}

// mapPgError 识别唯一约束冲突并返回业务错误；其余归类为 ErrInternal。
func mapPgError(err error, seqConstraint, dupCode, dupMsg string) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.ConstraintName == seqConstraint {
		return errs.New(dupCode, dupMsg, 409)
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return errs.ErrNotFound
	}
	return errs.ErrInternal.Wrap(err)
}

// ---------------------------------------------------------------------------
// IssueRepository — 跨类型视图层仓储
// ---------------------------------------------------------------------------

// Issue DTO — 跨类型只读投影，与 coordinator.go 的 Issue 结构对齐。
//
// 将最小字段集内联在此处，避免 repositories 包与 issue 包产生循环依赖。
// 后续如需扩展字段，可逐步迁移至共享的 issue.WorkitemView。
type Issue struct {
	ID          int64   `json:"id"`
	PublicID    string  `json:"public_id"`
	WorkspaceID int64   `json:"workspace_id"`
	ProjectID   int64   `json:"project_id"`
	SequenceID  int64   `json:"sequence_id"`
	Identifier  string  `json:"identifier"`
	TypeCode    string  `json:"type_code"`
	ParentID    *int64  `json:"parent_id,omitempty"`
	Depth       int     `json:"depth"`
	Name        string  `json:"name"`
	StateID     int64   `json:"state_id"`
	StateName   string  `json:"state_name,omitempty"`
	StateGroup  string  `json:"state_group,omitempty"`
	StateColor  string  `json:"state_color,omitempty"`
	Priority    string  `json:"priority"`
	SprintID    *int64  `json:"sprint_id,omitempty"`
	VersionID   *int64  `json:"version_id,omitempty"`
	Progress    int     `json:"progress"`
	TargetDate  *string `json:"target_date,omitempty"`
	IsDraft     bool    `json:"is_draft"`
	SortOrder   float64 `json:"sort_order"`
	Version     int     `json:"version"`
	CreatedBy   int64   `json:"created_by"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
	Severity    *int    `json:"severity,omitempty"`
	FoundPhase  *string `json:"found_phase,omitempty"`
	Category    *string `json:"category,omitempty"`
	Assignees   []int64 `json:"assignees,omitempty"`
	Labels      []int64 `json:"labels,omitempty"`
	Modules     []int64 `json:"modules,omitempty"`
}

// IssueRepository 定义跨类型 Issue 视图的仓储接口。
//
// Issue 是跨 task / requirement / defect 三表的统一视图，
// 不对应单张物理表；实现内部通过 UNION ALL 或按类型分派操作。
type IssueRepository interface {
	// GetByID 按全局 ID 查询工作项（跨类型检测）
	GetByID(ctx context.Context, id int64) (*Issue, error)

	// Update 更新工作项（跨类型分派，乐观锁）
	Update(ctx context.Context, wsID, id int64, updates map[string]any, version int) (*Issue, error)

	// Delete 软删除工作项（跨类型分派）
	Delete(ctx context.Context, wsID, id int64) (bool, error)

	// ListByWorkspace 列出工作空间下所有工作项
	ListByWorkspace(ctx context.Context, wsID int64, opts ListOptions) ([]Issue, error)

	// ListByProject 列出项目内所有工作项
	ListByProject(ctx context.Context, wsID, projectID int64, opts ListOptions) ([]Issue, error)
}

// postgresIssueRepository 是 IssueRepository 的 PostgreSQL 实现。
type postgresIssueRepository struct {
	db *pgxpool.Pool
}

// NewIssueRepository 构造 Issue 仓储实例。
func NewIssueRepository(db *pgxpool.Pool) IssueRepository {
	return &postgresIssueRepository{db: db}
}

// GetByID 按 ID 跨类型查询工作项。
func (r *postgresIssueRepository) GetByID(ctx context.Context, id int64) (*Issue, error) {
	row := r.db.QueryRow(ctx, `
		SELECT i.id, i.public_id, i.workspace_id, i.project_id, i.sequence_id,
		       i.type_code, i.parent_id, i.depth, i.name,
		       i.state_id, s.name, s."group", s.color,
		       i.priority, i.sprint_id, i.version_id, i.progress,
		       i.target_date::text, i.is_draft, i.sort_order, i.version,
		       i.created_by, i.created_at::text, i.updated_at::text
		FROM (
			SELECT id, public_id, workspace_id, project_id, sequence_id, 'task'::text AS type_code,
			       parent_id, depth, name, state_id, priority, sprint_id, version_id, progress,
			       target_date, is_draft, sort_order, version, created_by, created_at, updated_at,
			       project_id AS pid
			FROM task WHERE id = $1 AND deleted = false
			UNION ALL
			SELECT id, public_id, workspace_id, project_id, sequence_id, 'requirement'::text,
			       parent_id, depth, name, state_id, priority, sprint_id, version_id, progress,
			       target_date, is_draft, sort_order, version, created_by, created_at, updated_at,
			       project_id
			FROM requirement WHERE id = $1 AND deleted = false
			UNION ALL
			SELECT id, public_id, workspace_id, project_id, sequence_id, 'defect'::text,
			       parent_id, depth, name, state_id, priority, sprint_id, version_id, progress,
			       target_date, is_draft, sort_order, version, created_by, created_at, updated_at,
			       project_id
			FROM defect WHERE id = $1 AND deleted = false
		) i
		JOIN states s ON s.id = i.state_id
		LIMIT 1`, id)

	iss, err := scanIssue(row)
	if err != nil {
		return nil, mapNotFound(err)
	}
	// 加载 M2M：type_code 直接作为表前缀（task / requirement / defect）
	iss.Assignees, _ = loadAssignees(ctx, r.db, iss.TypeCode, id)
	iss.Labels, _ = loadIntArray(ctx, r.db,
		fmt.Sprintf("SELECT label_id FROM %s_labels WHERE %s_id = $1", iss.TypeCode, iss.TypeCode), id)
	iss.Modules, _ = loadIntArray(ctx, r.db,
		fmt.Sprintf("SELECT module_id FROM %s_modules WHERE %s_id = $1", iss.TypeCode, iss.TypeCode), id)
	iss.Identifier = r.loadIdentifier(ctx, iss.ProjectID, iss.SequenceID)
	return iss, nil
}

// loadIdentifier 拼接项目标识符 + 序号。
func (r *postgresIssueRepository) loadIdentifier(ctx context.Context, projectID, seqID int64) string {
	var identifier string
	_ = r.db.QueryRow(ctx, `SELECT identifier FROM projects WHERE id = $1`, projectID).Scan(&identifier)
	if identifier == "" {
		identifier = "PROJ"
	}
	return identifier + "-" + strconv.FormatInt(seqID, 10)
}

// loadAssignees 加载工作项的 assignee 列表。
func loadAssignees(ctx context.Context, db *pgxpool.Pool, typeCode string, id int64) ([]int64, error) {
	return loadIntArray(ctx, db,
		fmt.Sprintf("SELECT user_id FROM %s_assignees WHERE %s_id = $1", typeCode, typeCode), id)
}

// scanIssue 将一行扫描到 Issue 结构。
func scanIssue(row pgx.Row) (*Issue, error) {
	var iss Issue
	var parentID sql.NullInt64
	var sprintID sql.NullInt64
	var versionID sql.NullInt64
	var targetDate sql.NullString
	if err := row.Scan(
		&iss.ID, &iss.PublicID, &iss.WorkspaceID, &iss.ProjectID, &iss.SequenceID,
		&iss.TypeCode, &parentID, &iss.Depth, &iss.Name,
		&iss.StateID, &iss.StateName, &iss.StateGroup, &iss.StateColor,
		&iss.Priority, &sprintID, &versionID, &iss.Progress,
		&targetDate, &iss.IsDraft, &iss.SortOrder, &iss.Version,
		&iss.CreatedBy, &iss.CreatedAt, &iss.UpdatedAt); err != nil {
		return nil, err
	}
	if parentID.Valid {
		v := parentID.Int64
		iss.ParentID = &v
	}
	if sprintID.Valid {
		v := sprintID.Int64
		iss.SprintID = &v
	}
	if versionID.Valid {
		v := versionID.Int64
		iss.VersionID = &v
	}
	if targetDate.Valid {
		iss.TargetDate = &targetDate.String
	}
	return &iss, nil
}

// Update 跨类型分派更新工作项；支持部分字段 + 乐观锁。
func (r *postgresIssueRepository) Update(ctx context.Context, wsID, id int64, updates map[string]any, version int) (*Issue, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, "SELECT set_config('app.workspace_id', $1, true)", strconv.FormatInt(wsID, 10)); err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}

	// 检测类型
	table, err := detectTableTx(ctx, tx, id)
	if err != nil {
		return nil, err
	}

	sets, args := buildUpdateSets(updates)
	if len(sets) > 0 {
		sets = append(sets, "updated_at = now()")
		query := fmt.Sprintf("UPDATE %s SET %s WHERE id = $%d AND workspace_id = $%d AND version = $%d AND deleted = false",
			table, strings.Join(sets, ", "), len(args)+1, len(args)+2, len(args)+3)
		args = append(args, id, wsID, version)
		tag, err := tx.Exec(ctx, query, args...)
		if err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
		if tag.RowsAffected() == 0 {
			return nil, errs.ErrVersionConflict
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	return r.GetByID(ctx, id)
}

// Delete 跨类型分派软删除工作项。
func (r *postgresIssueRepository) Delete(ctx context.Context, wsID, id int64) (bool, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return false, errs.ErrInternal.Wrap(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, "SELECT set_config('app.workspace_id', $1, true)", strconv.FormatInt(wsID, 10)); err != nil {
		return false, errs.ErrInternal.Wrap(err)
	}

	table, err := detectTableTx(ctx, tx, id)
	if err != nil {
		return false, err
	}

	tag, err := tx.Exec(ctx, fmt.Sprintf(
		"UPDATE %s SET deleted = true, updated_at = now() WHERE id = $1 AND workspace_id = $2 AND deleted = false", table), id, wsID)
	if err != nil {
		return false, errs.ErrInternal.Wrap(err)
	}
	if tag.RowsAffected() == 0 {
		return false, errs.ErrNotFound
	}
	if err := tx.Commit(ctx); err != nil {
		return false, errs.ErrInternal.Wrap(err)
	}
	return true, nil
}

// ListByWorkspace 列出工作空间下所有工作项（跨类型）。
func (r *postgresIssueRepository) ListByWorkspace(ctx context.Context, wsID int64, opts ListOptions) ([]Issue, error) {
	return r.list(ctx, wsID, 0, opts)
}

// ListByProject 列出项目内所有工作项（跨类型）。
func (r *postgresIssueRepository) ListByProject(ctx context.Context, wsID, projectID int64, opts ListOptions) ([]Issue, error) {
	return r.list(ctx, wsID, projectID, opts)
}

// list 为 ListByWorkspace / ListByProject 的共用实现。
func (r *postgresIssueRepository) list(ctx context.Context, wsID, projectID int64, opts ListOptions) ([]Issue, error) {
	opts.normalize()
	where, args := opts.buildWhere(wsID, projectID)
	limitIdx := len(args) + 1
	offsetIdx := len(args) + 2
	args = append(args, opts.Limit, opts.Offset)

	rows, err := r.db.Query(ctx, `
		SELECT i.id, i.public_id, i.workspace_id, i.project_id, i.sequence_id,
		       i.type_code, i.parent_id, i.depth, i.name,
		       i.state_id, s.name, s."group", s.color,
		       i.priority, i.sprint_id, i.version_id, i.progress,
		       i.target_date::text, i.is_draft, i.sort_order, i.version,
		       i.created_by, i.created_at::text, i.updated_at::text
		FROM (
			SELECT id, public_id, workspace_id, project_id, sequence_id, 'task'::text AS type_code,
			       parent_id, depth, name, state_id, priority, sprint_id, version_id, progress,
			       target_date, is_draft, sort_order, version, created_by, created_at, updated_at,
			       project_id AS pid
			FROM task WHERE workspace_id = $1 AND deleted = false
			UNION ALL
			SELECT id, public_id, workspace_id, project_id, sequence_id, 'requirement'::text,
			       parent_id, depth, name, state_id, priority, sprint_id, version_id, progress,
			       target_date, is_draft, sort_order, version, created_by, created_at, updated_at,
			       project_id
			FROM requirement WHERE workspace_id = $1 AND deleted = false
			UNION ALL
			SELECT id, public_id, workspace_id, project_id, sequence_id, 'defect'::text,
			       parent_id, depth, name, state_id, priority, sprint_id, version_id, progress,
			       target_date, is_draft, sort_order, version, created_by, created_at, updated_at,
			       project_id
			FROM defect WHERE workspace_id = $1 AND deleted = false
		) i
		JOIN states s ON s.id = i.state_id
		`+where+`
		ORDER BY `+opts.buildSort()+`
		LIMIT $`+strconv.Itoa(limitIdx)+` OFFSET $`+strconv.Itoa(offsetIdx), args...)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()

	var issues []Issue
	for rows.Next() {
		iss, err := scanIssue(rows)
		if err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
		iss.Identifier = r.loadIdentifier(ctx, iss.ProjectID, iss.SequenceID)
		issues = append(issues, *iss)
	}
	return issues, rows.Err()
}

// ---------------------------------------------------------------------------
// IssueRepository 内部辅助
// ---------------------------------------------------------------------------

// detectTableTx 确定工作项对应的白名单表名。
func detectTableTx(ctx context.Context, tx pgx.Tx, id int64) (string, error) {
	var tc string
	err := tx.QueryRow(ctx, `
		SELECT 'task' FROM task WHERE id = $1 AND deleted = false
		UNION ALL
		SELECT 'requirement' FROM requirement WHERE id = $1 AND deleted = false
		UNION ALL
		SELECT 'defect' FROM defect WHERE id = $1 AND deleted = false
		LIMIT 1`, id).Scan(&tc)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", errs.ErrNotFound
		}
		return "", errs.ErrInternal.Wrap(err)
	}
	return tc, nil
}

// buildUpdateSets 从 map 生成 SET 子句和参数（白名单字段防止 SQL 注入）。
func buildUpdateSets(updates map[string]any) ([]string, []any) {
	allowed := map[string]bool{
		"name": true, "description_html": true, "priority": true,
		"state_id": true, "parent_id": true, "point": true,
		"target_date": true, "progress": true, "category": true,
		"severity": true, "found_phase": true, "is_draft": true,
	}
	var sets []string
	var args []any
	idx := 1
	// 按稳定顺序处理（map 遍历顺序不可预测，这里按字母排序可重现）
	var keys []string
	for k := range updates {
		if allowed[k] {
			keys = append(keys, k)
		}
	}
	// 简单冒泡即可（字段数 < 20）
	for i := 0; i < len(keys); i++ {
		for j := i + 1; j < len(keys); j++ {
			if keys[i] > keys[j] {
				keys[i], keys[j] = keys[j], keys[i]
			}
		}
	}
	for _, k := range keys {
		sets = append(sets, k+" = $"+strconv.Itoa(idx))
		args = append(args, updates[k])
		idx++
	}
	return sets, args
}
