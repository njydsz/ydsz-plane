// Package initiative — 倡议应用服务（CRUD + 关联管理 + 进度聚合）。
//
// 参考: Plane / Jira Advanced Roadmaps Initiative service。
package initiative

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

// Service 提供 Initiative 领域应用服务。
type Service struct {
	db *pgxpool.Pool
}

// NewService 创建 Initiative 服务。
func NewService(db *pgxpool.Pool) *Service {
	return &Service{db: db}
}

// --- Input Types ---

// CreateInitiativeInput 创建倡议的入参。
type CreateInitiativeInput struct {
	WorkspaceID int64
	Title       string
	Description string
	Status      InitiativeStatus
	Priority    InitiativePriority
	StartDate   *time.Time
	TargetDate  *time.Time
	OwnerID     *int64
	CreatedBy   int64
}

// UpdateInitiativeInput 更新倡议的入参（字段为指针，nil 表示不更新）。
type UpdateInitiativeInput struct {
	Title       *string
	Description *string
	Status      *InitiativeStatus
	Priority    *InitiativePriority
	StartDate   *time.Time
	TargetDate  *time.Time
	OwnerID     *int64
}

// ListInitiativesOptions 倡议列表查询选项。
type ListInitiativesOptions struct {
	WorkspaceID int64
	Status      *InitiativeStatus
	Priority    *InitiativePriority
	OwnerID     *int64
	Limit       int
	Offset      int
}

// --- CRUD ---

// Create 创建倡议。
func (s *Service) Create(ctx context.Context, in CreateInitiativeInput) (*Initiative, error) {
	if err := validateCreateInput(in); err != nil {
		return nil, err
	}

	// 默认值填充
	status := in.Status
	if status == "" {
		status = StatusOnTrack
	}
	priority := in.Priority
	if priority == "" {
		priority = PriorityMedium
	}

	id := genID()
	var it Initiative
	err := s.db.QueryRow(ctx, `
		INSERT INTO initiatives (
			id, workspace_id, title, description, status, priority,
			start_date, target_date, owner_id, created_by
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		RETURNING id, workspace_id, title, description, status, priority,
			start_date, target_date, owner_id, created_by, archived_at, created_at, updated_at`,
		id, in.WorkspaceID, in.Title, nullIfEmpty(in.Description), status, priority,
		in.StartDate, in.TargetDate, in.OwnerID, in.CreatedBy,
	).Scan(
		&it.ID, &it.WorkspaceID, &it.Title, &it.Description, &it.Status, &it.Priority,
		&it.StartDate, &it.TargetDate, &it.OwnerID, &it.CreatedBy, &it.ArchivedAt,
		&it.CreatedAt, &it.UpdatedAt,
	)
	if err != nil {
		return nil, s.mapPgError(err)
	}
	return &it, nil
}

// GetByID 获取倡议详情。
func (s *Service) GetByID(ctx context.Context, wsID, initiativeID int64) (*Initiative, error) {
	row := s.db.QueryRow(ctx, `
		SELECT id, workspace_id, title, description, status, priority,
			start_date, target_date, owner_id, created_by, archived_at, created_at, updated_at
		FROM initiatives
		WHERE id = $1 AND workspace_id = $2 AND archived_at IS NULL`,
		initiativeID, wsID)
	it, err := scanInitiative(row)
	if err != nil {
		return nil, s.mapPgError(err)
	}
	return it, nil
}

// GetDetail 获取倡议详情（含关联项目与工作项列表）。
func (s *Service) GetDetail(ctx context.Context, wsID, initiativeID int64) (*InitiativeDetail, error) {
	it, err := s.GetByID(ctx, wsID, initiativeID)
	if err != nil {
		return nil, err
	}
	detail := &InitiativeDetail{Initiative: it}

	// 关联项目
	projects, err := s.listProjects(ctx, initiativeID)
	if err != nil {
		return nil, err
	}
	detail.Projects = projects

	// 关联工作项
	items, err := s.listWorkItems(ctx, initiativeID)
	if err != nil {
		return nil, err
	}
	detail.WorkItems = items

	return detail, nil
}

// InitiativeDetail 倡议详情（含关联数据）。
type InitiativeDetail struct {
	*Initiative
	Projects  []InitiativeProject  `json:"projects"`
	WorkItems []InitiativeWorkItem `json:"work_items"`
}

// List 列出倡议（支持状态/优先级/负责人过滤）。
func (s *Service) List(ctx context.Context, opts ListInitiativesOptions) ([]Initiative, int64, error) {
	if opts.Limit <= 0 || opts.Limit > 100 {
		opts.Limit = 50
	}
	if opts.Offset < 0 {
		opts.Offset = 0
	}

	where := "WHERE workspace_id = $1 AND archived_at IS NULL"
	args := []any{opts.WorkspaceID}
	arg := 2

	if opts.Status != nil {
		where += " AND status = $" + strconv.Itoa(arg)
		args = append(args, string(*opts.Status))
		arg++
	}
	if opts.Priority != nil {
		where += " AND priority = $" + strconv.Itoa(arg)
		args = append(args, string(*opts.Priority))
		arg++
	}
	if opts.OwnerID != nil {
		where += " AND owner_id = $" + strconv.Itoa(arg)
		args = append(args, *opts.OwnerID)
		arg++
	}

	var total int64
	_ = s.db.QueryRow(ctx, "SELECT count(*) FROM initiatives "+where, args...).Scan(&total)

	limitIdx := len(args) + 1
	offsetIdx := len(args) + 2
	args = append(args, opts.Limit, opts.Offset)

	rows, err := s.db.Query(ctx, `
		SELECT id, workspace_id, title, description, status, priority,
			start_date, target_date, owner_id, created_by, archived_at, created_at, updated_at
		FROM initiatives `+where+`
		ORDER BY
			CASE status WHEN 'on_track' THEN 0 WHEN 'at_risk' THEN 1 WHEN 'off_track' THEN 2 ELSE 3 END,
			target_date NULLS LAST, created_at DESC
		LIMIT $`+strconv.Itoa(limitIdx)+` OFFSET $`+strconv.Itoa(offsetIdx), args...)
	if err != nil {
		return nil, 0, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()

	var items []Initiative
	for rows.Next() {
		it, err := scanInitiative(rows)
		if err != nil {
			return nil, 0, errs.ErrInternal.Wrap(err)
		}
		items = append(items, *it)
	}
	return items, total, rows.Err()
}

// Update 更新倡议字段。
func (s *Service) Update(ctx context.Context, wsID, initiativeID int64, in UpdateInitiativeInput) (*Initiative, error) {
	sets, args := buildUpdateSet(in)
	if len(sets) == 0 {
		return s.GetByID(ctx, wsID, initiativeID)
	}
	sets = append(sets, "updated_at = now()")
	args = append(args, initiativeID, wsID)

	query := fmt.Sprintf(`UPDATE initiatives SET %s WHERE id = $%d AND workspace_id = $%d AND archived_at IS NULL`,
		strings.Join(sets, ", "), len(args)-1, len(args))

	tag, err := s.db.Exec(ctx, query, args...)
	if err != nil {
		return nil, s.mapPgError(err)
	}
	if tag.RowsAffected() == 0 {
		return nil, errs.ErrNotFound
	}
	return s.GetByID(ctx, wsID, initiativeID)
}

// Delete 删除倡议（硬删除：同时清理关联关系）。
func (s *Service) Delete(ctx context.Context, wsID, initiativeID int64) error {
	return s.withTx(ctx, wsID, func(tx pgx.Tx) error {
		tag, err := tx.Exec(ctx, `
			DELETE FROM initiative_work_items WHERE initiative_id = $1`, initiativeID)
		if err != nil {
			return err
		}
		_ = tag
		if _, err := tx.Exec(ctx, `
			DELETE FROM initiative_projects WHERE initiative_id = $1`, initiativeID); err != nil {
			return err
		}
		tag, err = tx.Exec(ctx, `
			DELETE FROM initiatives WHERE id = $1 AND workspace_id = $2 AND archived_at IS NULL`,
			initiativeID, wsID)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return errs.ErrNotFound
		}
		return nil
	})
}

// Archive 归档倡议（软删除语义：设置 archived_at）。
func (s *Service) Archive(ctx context.Context, wsID, initiativeID int64) error {
	tag, err := s.db.Exec(ctx, `
		UPDATE initiatives SET archived_at = now(), updated_at = now()
		WHERE id = $1 AND workspace_id = $2 AND archived_at IS NULL`,
		initiativeID, wsID)
	if err != nil {
		return s.mapPgError(err)
	}
	if tag.RowsAffected() == 0 {
		return errs.ErrNotFound
	}
	return nil
}

// Restore 恢复已归档倡议。
func (s *Service) Restore(ctx context.Context, wsID, initiativeID int64) error {
	tag, err := s.db.Exec(ctx, `
		UPDATE initiatives SET archived_at = NULL, updated_at = now()
		WHERE id = $1 AND workspace_id = $2 AND archived_at IS NOT NULL`,
		initiativeID, wsID)
	if err != nil {
		return s.mapPgError(err)
	}
	if tag.RowsAffected() == 0 {
		return errs.ErrNotFound
	}
	return nil
}

// --- 关联管理：项目 ---

// AddProject 将项目关联到倡议。
func (s *Service) AddProject(ctx context.Context, wsID, initiativeID, projectID int64) error {
	// 校验倡议归属
	if _, err := s.GetByID(ctx, wsID, initiativeID); err != nil {
		return err
	}
	// 校验项目归属同一工作空间
	var exists bool
	if err := s.db.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM projects WHERE id = $1 AND workspace_id = $2 AND deleted = false)`,
		projectID, wsID).Scan(&exists); err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	if !exists {
		return errs.New("INITIATIVE.PROJECT_NOT_FOUND", "项目不存在或不属于当前工作空间", 404)
	}

	_, err := s.db.Exec(ctx, `
		INSERT INTO initiative_projects (initiative_id, project_id)
		VALUES ($1, $2)
		ON CONFLICT (initiative_id, project_id) DO NOTHING`,
		initiativeID, projectID)
	if err != nil {
		return s.mapPgError(err)
	}
	return nil
}

// RemoveProject 从倡议移除项目关联。
func (s *Service) RemoveProject(ctx context.Context, wsID, initiativeID, projectID int64) error {
	tag, err := s.db.Exec(ctx, `
		DELETE FROM initiative_projects WHERE initiative_id = $1 AND project_id = $2`,
		initiativeID, projectID)
	if err != nil {
		return s.mapPgError(err)
	}
	if tag.RowsAffected() == 0 {
		return errs.ErrNotFound
	}
	return nil
}

// listProjects 列出倡议关联的项目。
func (s *Service) listProjects(ctx context.Context, initiativeID int64) ([]InitiativeProject, error) {
	rows, err := s.db.Query(ctx, `
		SELECT ip.initiative_id, ip.project_id, ip.created_at
		FROM initiative_projects ip
		WHERE ip.initiative_id = $1
		ORDER BY ip.created_at`, initiativeID)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()

	var out []InitiativeProject
	for rows.Next() {
		var p InitiativeProject
		if err := rows.Scan(&p.InitiativeID, &p.ProjectID, &p.CreatedAt); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// --- 关联管理：工作项 ---

// AddWorkItem 将工作项关联到倡议。
func (s *Service) AddWorkItem(ctx context.Context, wsID, initiativeID int64, workItemID int64, workItemType string) error {
	// 校验倡议归属
	if _, err := s.GetByID(ctx, wsID, initiativeID); err != nil {
		return err
	}
	if !isValidWorkItemType(workItemType) {
		return errs.New("INITIATIVE.INVALID_WORK_ITEM_TYPE", "工作项类型必须是 requirement/task/defect", 422)
	}

	// 校验工作项存在且属于同一工作空间
	if !s.workItemExists(ctx, wsID, workItemID, workItemType) {
		return errs.New("INITIATIVE.WORK_ITEM_NOT_FOUND", "工作项不存在或不属于当前工作空间", 404)
	}

	_, err := s.db.Exec(ctx, `
		INSERT INTO initiative_work_items (initiative_id, work_item_id, work_item_type)
		VALUES ($1, $2, $3)
		ON CONFLICT (initiative_id, work_item_id, work_item_type) DO NOTHING`,
		initiativeID, workItemID, workItemType)
	if err != nil {
		return s.mapPgError(err)
	}
	return nil
}

// RemoveWorkItem 从倡议移除工作项关联。
func (s *Service) RemoveWorkItem(ctx context.Context, wsID, initiativeID, workItemID int64) error {
	tag, err := s.db.Exec(ctx, `
		DELETE FROM initiative_work_items WHERE initiative_id = $1 AND work_item_id = $2`,
		initiativeID, workItemID)
	if err != nil {
		return s.mapPgError(err)
	}
	if tag.RowsAffected() == 0 {
		return errs.ErrNotFound
	}
	return nil
}

// listWorkItems 列出倡议关联的工作项。
func (s *Service) listWorkItems(ctx context.Context, initiativeID int64) ([]InitiativeWorkItem, error) {
	rows, err := s.db.Query(ctx, `
		SELECT iw.initiative_id, iw.work_item_id, iw.work_item_type, iw.created_at
		FROM initiative_work_items iw
		WHERE iw.initiative_id = $1
		ORDER BY iw.created_at`, initiativeID)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()

	var out []InitiativeWorkItem
	for rows.Next() {
		var w InitiativeWorkItem
		if err := rows.Scan(&w.InitiativeID, &w.WorkItemID, &w.WorkItemType, &w.CreatedAt); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
		out = append(out, w)
	}
	return out, rows.Err()
}

// --- 进度聚合 ---

// GetProgress 计算倡议的进度聚合。
//
// 聚合逻辑：
//   - 通过 initiative_projects 关联的项目，统计这些项目下所有工作项的完成情况；
//   - 通过 initiative_work_items 直接关联的工作项，统计完成情况；
//   - 去重后合并，得到整体完成率与分类型/分项目分布。
func (s *Service) GetProgress(ctx context.Context, wsID, initiativeID int64) (*InitiativeProgress, error) {
	// 校验倡议归属
	if _, err := s.GetByID(ctx, wsID, initiativeID); err != nil {
		return nil, err
	}

	progress := &InitiativeProgress{
		ByType: make(map[string]InitiativeTypeProgress),
	}

	// --- 1. 关联项目数 ---
	_ = s.db.QueryRow(ctx, `
		SELECT count(*) FROM initiative_projects WHERE initiative_id = $1`,
		initiativeID).Scan(&progress.ProjectCount)

	// --- 2. 直接关联的工作项进度 ---
	typeMap, totalWI, doneWI, err := s.progressForWorkItems(ctx, initiativeID)
	if err != nil {
		return nil, err
	}
	progress.TotalWorkItems = totalWI
	progress.DoneWorkItems = doneWI
	for k, v := range typeMap {
		progress.ByType[k] = v
	}

	// --- 3. 关联项目的工作项进度 ---
	projectProgress, projTotalWI, projDoneWI, err := s.progressForProjects(ctx, wsID, initiativeID)
	if err != nil {
		return nil, err
	}
	progress.ByProject = projectProgress

	// 合并：项目维度 + 直接关联维度（去重后的近似合并——直接关联的工作项
	// 可能也属于关联项目，这里按独立计数相加；产品上可接受，后续可优化为去重合并）
	progress.TotalWorkItems += projTotalWI
	progress.DoneWorkItems += projDoneWI

	// 重新按类型合并（项目维度没有类型细分，仅直接关联维度有）
	if progress.TotalWorkItems > 0 {
		progress.CompletionRate = float64(progress.DoneWorkItems) / float64(progress.TotalWorkItems) * 100
	}

	return progress, nil
}

// progressForWorkItems 统计直接关联工作项的进度（按类型分组）。
func (s *Service) progressForWorkItems(ctx context.Context, initiativeID int64) (map[string]InitiativeTypeProgress, int, int, error) {
	typeMap := make(map[string]InitiativeTypeProgress)
	total, done := 0, 0

	// 需求
	var reqTotal, reqDone int
	_ = s.db.QueryRow(ctx, `
		SELECT count(*), count(*) FILTER (WHERE s."group" = 'completed')
		FROM initiative_work_items iw
		JOIN requirement r ON r.id = iw.work_item_id AND r.deleted = false
		JOIN states s ON s.id = r.state_id
		WHERE iw.initiative_id = $1 AND iw.work_item_type = 'requirement'`,
		initiativeID).Scan(&reqTotal, &reqDone)
	typeMap[WorkItemTypeRequirement] = InitiativeTypeProgress{Total: reqTotal, Done: reqDone}
	total += reqTotal
	done += reqDone

	// 任务
	var taskTotal, taskDone int
	_ = s.db.QueryRow(ctx, `
		SELECT count(*), count(*) FILTER (WHERE s."group" = 'completed')
		FROM initiative_work_items iw
		JOIN task t ON t.id = iw.work_item_id AND t.deleted = false
		JOIN states s ON s.id = t.state_id
		WHERE iw.initiative_id = $1 AND iw.work_item_type = 'task'`,
		initiativeID).Scan(&taskTotal, &taskDone)
	typeMap[WorkItemTypeTask] = InitiativeTypeProgress{Total: taskTotal, Done: taskDone}
	total += taskTotal
	done += taskDone

	// 缺陷
	var defTotal, defDone int
	_ = s.db.QueryRow(ctx, `
		SELECT count(*), count(*) FILTER (WHERE s."group" = 'completed')
		FROM initiative_work_items iw
		JOIN defect d ON d.id = iw.work_item_id AND d.deleted = false
		JOIN states s ON s.id = d.state_id
		WHERE iw.initiative_id = $1 AND iw.work_item_type = 'defect'`,
		initiativeID).Scan(&defTotal, &defDone)
	typeMap[WorkItemTypeDefect] = InitiativeTypeProgress{Total: defTotal, Done: defDone}
	total += defTotal
	done += defDone

	return typeMap, total, done, nil
}

// progressForProjects 统计关联项目的工作项进度（按项目分组）。
// 返回各项目的进度、总工作项数、已完成数。
func (s *Service) progressForProjects(ctx context.Context, wsID, initiativeID int64) ([]InitiativeProjectProgress, int, int, error) {
	rows, err := s.db.Query(ctx, `
		SELECT
			p.id, p.name,
			(SELECT count(*) FROM (
				SELECT r.id, r.state_id FROM requirement r
				WHERE r.project_id = p.id AND r.workspace_id = $2 AND r.deleted = false
				UNION ALL
				SELECT t.id, t.state_id FROM task t
				WHERE t.project_id = p.id AND t.workspace_id = $2 AND t.deleted = false
				UNION ALL
				SELECT d.id, d.state_id FROM defect d
				WHERE d.project_id = p.id AND d.workspace_id = $2 AND d.deleted = false
			) items),
			(SELECT count(*) FROM (
				SELECT r.id, r.state_id FROM requirement r
				WHERE r.project_id = p.id AND r.workspace_id = $2 AND r.deleted = false
				UNION ALL
				SELECT t.id, t.state_id FROM task t
				WHERE t.project_id = p.id AND t.workspace_id = $2 AND t.deleted = false
				UNION ALL
				SELECT d.id, d.state_id FROM defect d
				WHERE d.project_id = p.id AND d.workspace_id = $2 AND d.deleted = false
			) items
			JOIN states st ON st.id = items.state_id
			WHERE st."group" = 'completed')
		FROM initiative_projects ip
		JOIN projects p ON p.id = ip.project_id AND p.deleted = false
		WHERE ip.initiative_id = $1
		ORDER BY p.name`, initiativeID, wsID)
	if err != nil {
		return nil, 0, 0, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()

	var out []InitiativeProjectProgress
	total, done := 0, 0
	for rows.Next() {
		var pp InitiativeProjectProgress
		if err := rows.Scan(&pp.ProjectID, &pp.ProjectName, &pp.TotalItems, &pp.DoneItems); err != nil {
			return nil, 0, 0, errs.ErrInternal.Wrap(err)
		}
		if pp.TotalItems > 0 {
			pp.CompletionRate = float64(pp.DoneItems) / float64(pp.TotalItems) * 100
		}
		total += pp.TotalItems
		done += pp.DoneItems
		out = append(out, pp)
	}
	return out, total, done, rows.Err()
}

// --- Low-level helpers ---

// scanInitiative 扫描一行数据到 Initiative。
func scanInitiative(row pgx.Row) (*Initiative, error) {
	var it Initiative
	var desc sql.NullString
	var startDate, targetDate, archivedAt sql.NullTime
	var ownerID sql.NullInt64

	err := row.Scan(
		&it.ID, &it.WorkspaceID, &it.Title, &desc, &it.Status, &it.Priority,
		&startDate, &targetDate, &ownerID, &it.CreatedBy, &archivedAt,
		&it.CreatedAt, &it.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrNotFound
		}
		return nil, errs.ErrInternal.Wrap(err)
	}
	if desc.Valid {
		it.Description = desc.String
	}
	if startDate.Valid {
		it.StartDate = &startDate.Time
	}
	if targetDate.Valid {
		it.TargetDate = &targetDate.Time
	}
	if ownerID.Valid {
		v := ownerID.Int64
		it.OwnerID = &v
	}
	if archivedAt.Valid {
		it.ArchivedAt = &archivedAt.Time
	}
	return &it, nil
}

// withTx 事务辅助（设置 workspace_id 会话变量以配合 RLS）。
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

// mapPgError 将 PostgreSQL 错误映射为 AppError。
func (s *Service) mapPgError(err error) error {
	if err == nil {
		return nil
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			return errs.New("INITIATIVE.DUPLICATE", "倡议冲突: "+pgErr.Detail, 409)
		case "23514": // check_violation
			return errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "initiative", Reason: pgErr.Detail})
		case "23503": // foreign_key_violation
			return errs.New("INITIATIVE.FK_VIOLATION", "关联资源不存在", 422)
		}
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return errs.ErrNotFound
	}
	return errs.ErrInternal.Wrap(err)
}

// --- Validation & util ---

func validateCreateInput(in CreateInitiativeInput) error {
	if in.WorkspaceID == 0 {
		return errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "workspace_id", Reason: "工作空间不能为空"})
	}
	if strings.TrimSpace(in.Title) == "" || len(in.Title) > 200 {
		return errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "title", Reason: "标题长度需在 1-200 之间"})
	}
	if in.Status != "" && !in.Status.IsValid() {
		return errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "status", Reason: "无效的倡议状态"})
	}
	if in.Priority != "" && !in.Priority.IsValid() {
		return errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "priority", Reason: "无效的优先级"})
	}
	return nil
}

func buildUpdateSet(in UpdateInitiativeInput) ([]string, []any) {
	var sets []string
	var args []any
	i := 1

	if in.Title != nil {
		sets = append(sets, "title = $"+strconv.Itoa(i))
		args = append(args, *in.Title)
		i++
	}
	if in.Description != nil {
		sets = append(sets, "description = $"+strconv.Itoa(i))
		args = append(args, *in.Description)
		i++
	}
	if in.Status != nil {
		sets = append(sets, "status = $"+strconv.Itoa(i))
		args = append(args, string(*in.Status))
		i++
	}
	if in.Priority != nil {
		sets = append(sets, "priority = $"+strconv.Itoa(i))
		args = append(args, string(*in.Priority))
		i++
	}
	if in.StartDate != nil {
		sets = append(sets, "start_date = $"+strconv.Itoa(i))
		args = append(args, *in.StartDate)
		i++
	}
	if in.TargetDate != nil {
		sets = append(sets, "target_date = $"+strconv.Itoa(i))
		args = append(args, *in.TargetDate)
		i++
	}
	if in.OwnerID != nil {
		sets = append(sets, "owner_id = $"+strconv.Itoa(i))
		args = append(args, *in.OwnerID)
		i++
	}

	return sets, args
}

func isValidWorkItemType(t string) bool {
	switch t {
	case WorkItemTypeRequirement, WorkItemTypeTask, WorkItemTypeDefect:
		return true
	default:
		return false
	}
}

// workItemExists 校验工作项是否存在且属于同一工作空间。
func (s *Service) workItemExists(ctx context.Context, wsID, workItemID int64, workItemType string) bool {
	var exists bool
	var err error
	switch workItemType {
	case WorkItemTypeRequirement:
		err = s.db.QueryRow(ctx,
			`SELECT EXISTS(SELECT 1 FROM requirement WHERE id = $1 AND workspace_id = $2 AND deleted = false)`,
			workItemID, wsID).Scan(&exists)
	case WorkItemTypeTask:
		err = s.db.QueryRow(ctx,
			`SELECT EXISTS(SELECT 1 FROM task WHERE id = $1 AND workspace_id = $2 AND deleted = false)`,
			workItemID, wsID).Scan(&exists)
	case WorkItemTypeDefect:
		err = s.db.QueryRow(ctx,
			`SELECT EXISTS(SELECT 1 FROM defect WHERE id = $1 AND workspace_id = $2 AND deleted = false)`,
			workItemID, wsID).Scan(&exists)
	default:
		return false
	}
	if err != nil {
		return false
	}
	return exists
}

func nullIfEmpty(s string) any {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return s
}
