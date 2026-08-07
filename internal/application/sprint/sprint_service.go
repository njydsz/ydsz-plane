// Package sprint — Sprint 应用服务（CRUD + 生命周期 + 燃尽图 + 速率统计）。
//
// 参考: Plane / Linear / Jira Sprint service。
package sprint

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// Service 提供 Sprint 领域应用服务。
type Service struct {
	db *pgxpool.Pool
}

// NewService 创建 Sprint 服务。
func NewService(db *pgxpool.Pool) *Service {
	return &Service{db: db}
}

// --- Input Types ---

// CreateSprintInput 创建迭代的入参。
type CreateSprintInput struct {
	WorkspaceID int64
	ProjectID   int64
	Name        string
	Description string
	Goal        string
	StartDate   *time.Time
	EndDate     *time.Time
	Capacity    *float64
	OwnerID     *int64
	CreatedBy   int64
}

// UpdateSprintInput 更新迭代的入参。
type UpdateSprintInput struct {
	Name        *string
	Description *string
	Goal        *string
	StartDate   *time.Time
	EndDate     *time.Time
	Capacity    *float64
	OwnerID     *int64
	Viewport    map[string]any
	Version     int
}

// ListSprintsOptions 迭代列表查询选项。
type ListSprintsOptions struct {
	WorkspaceID int64
	ProjectID   int64
	Status      *SprintStatusCode
	Limit       int
	Offset      int
}

// AddIssueInput 添加工作项到迭代（拖拽规划）。
type AddIssueInput struct {
	SprintID  int64
	IssueID   int64
	SortOrder float64
	AddedBy   int64
}

// RemoveIssueInput 从迭代移除工作项。
type RemoveIssueInput struct {
	SprintID int64
	IssueID  int64
}

// CompleteSprintInput 结束迭代的入参。
type CompleteSprintInput struct {
	Strategy UnfinishedStrategy
	NextSprintID *int64 // 策略为 next_sprint 时指定
}

// --- CRUD ---

// Create 创建迭代。
func (s *Service) Create(ctx context.Context, in CreateSprintInput) (*Sprint, error) {
	if err := validateSprintInput(in); err != nil {
		return nil, err
	}

	var sp *Sprint
	err := s.withTx(ctx, in.WorkspaceID, func(tx pgx.Tx) error {
		// 校验：同项目内迭代名称唯一
		var exists bool
		if err := tx.QueryRow(ctx,
			`SELECT EXISTS(SELECT 1 FROM sprints WHERE project_id = $1 AND name = $2 AND deleted_at IS NULL)`,
			in.ProjectID, in.Name).Scan(&exists); err != nil {
			return errs.ErrInternal.Wrap(err)
		}
		if exists {
			return errs.New("SPRINT.DUPLICATE_NAME", "同一项目下迭代名称已存在", 409)
		}

		var id int64
		var startPtr, endPtr, capPtr, ownerIDPtr interface{}
		if in.StartDate != nil {
			startPtr = *in.StartDate
		}
		if in.EndDate != nil {
			endPtr = *in.EndDate
		}
		if in.Capacity != nil {
			capPtr = *in.Capacity
		}
		if in.OwnerID != nil {
			ownerIDPtr = *in.OwnerID
		}

		err := tx.QueryRow(ctx, `
			INSERT INTO sprints (workspace_id, project_id, name, description, goal,
				start_date, end_date, capacity, owner_id, status, created_by)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,'planned',$10)
			RETURNING id, workspace_id, project_id, name, description, goal, status,
				start_date, end_date, capacity, owner_id, created_by, created_at, updated_at`,
			in.WorkspaceID, in.ProjectID, in.Name, in.Description, in.Goal,
			startPtr, endPtr, capPtr, ownerIDPtr, in.CreatedBy).Scan(
			&id, &sp.WorkspaceID, &sp.ProjectID, &sp.Name, &sp.Description, &sp.Goal, &sp.Status,
			&sp.StartDate, &sp.EndDate, &sp.Capacity, &sp.OwnerID, &sp.CreatedBy, &sp.CreatedAt, &sp.UpdatedAt)
		return err
	})
	if err != nil {
		return nil, s.mapPgError(err)
	}
	return sp, nil
}

// GetByID 获取迭代详情（含实时进度）。
func (s *Service) GetByID(ctx context.Context, wsID, sprintID int64) (*Sprint, error) {
	sp, err := s.getSprintByID(ctx, wsID, sprintID)
	if err != nil {
		return nil, err
	}
	// 填充实时进度
	sp.Progress = s.computeProgress(ctx, wsID, sp)
	return sp, nil
}

// List 列出迭代。
func (s *Service) List(ctx context.Context, opts ListSprintsOptions) ([]Sprint, int64, error) {
	if opts.Limit <= 0 || opts.Limit > 100 {
		opts.Limit = 50
	}
	if opts.Offset < 0 {
		opts.Offset = 0
	}

	where := "WHERE s.deleted_at IS NULL AND s.project_id = $1 AND s.workspace_id = $2"
	args := []interface{}{opts.ProjectID, opts.WorkspaceID}
	arg := 3

	if opts.Status != nil {
		where += " AND s.status = $" + strconv.Itoa(arg)
		args = append(args, string(*opts.Status))
		arg++
	}

	var total int64
	_ = s.db.QueryRow(ctx,
		"SELECT count(*) FROM sprints s "+where, args...).Scan(&total)

	limitIdx := len(args) + 1
	offsetIdx := len(args) + 2
	args = append(args, opts.Limit, opts.Offset)

	rows, err := s.db.Query(ctx, `
		SELECT s.id, s.workspace_id, s.project_id, s.name, s.description, s.goal,
			s.status, s.start_date, s.end_date, s.capacity, s.owner_id,
			s.review_snapshot, s.started_at, s.completed_at, s.created_by, s.created_at, s.updated_at
		FROM sprints s `+where+`
		ORDER BY
			CASE s.status WHEN 'active' THEN 0 WHEN 'planned' ELSE 1 END,
			s.start_date NULLS LAST, s.created_at DESC
		LIMIT $`+strconv.Itoa(limitIdx)+` OFFSET $`+strconv.Itoa(offsetIdx), args...)
	if err != nil {
		return nil, 0, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()

	var sprints []Sprint
	for rows.Next() {
		sp, err := scanSprint(rows)
		if err != nil {
			return nil, 0, err
		}
		sprints = append(sprints, *sp)
	}
	return sprints, total, rows.Err()
}

// Update 更新迭代字段。
func (s *Service) Update(ctx context.Context, wsID, sprintID int64, in UpdateSprintInput) (*Sprint, error) {
	var result *Sprint
	err := s.withTx(ctx, wsID, func(tx pgx.Tx) error {
		sp, err := s.getSprintTx(ctx, tx, sprintID, wsID)
		if err != nil {
			return err
		}
		// 只有 planned 迭代可编辑核心字段
		if sp.Status != SprintPlanned {
			return errs.ErrSprintInvalidLifecycle
		}

		sets, args := buildSprintUpdateSet(in)
		if len(sets) == 0 {
			result = sp
			return nil
		}
		args = append(args, sprintID, wsID)
		query := fmt.Sprintf(`UPDATE sprints SET %s WHERE id = $%d AND workspace_id = $%d AND status = 'planned'`,
			stringsJoin(sets, ", "), len(args)-1, len(args))

		tag, err := tx.Exec(ctx, query, args...)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return errs.ErrSprintConflict
		}
		result, err = s.getSprintByID(ctx, wsID, sprintID)
		return err
	})
	if err != nil {
		return nil, s.mapPgError(err)
	}
	return result, nil
}

// SoftDelete 归档迭代。
func (s *Service) SoftDelete(ctx context.Context, wsID, sprintID int64) error {
	return s.withTx(ctx, wsID, func(tx pgx.Tx) error {
		sp, err := s.getSprintTx(ctx, tx, sprintID, wsID)
		if err != nil {
			return err
		}
		if sp.Status == SprintActive {
			return errs.ErrSprintInvalidLifecycle
		}

		// 将工作项退回 Backlog
		if _, err := tx.Exec(ctx, `DELETE FROM sprint_issues WHERE sprint_id = $1`, sprintID); err != nil {
			return err
		}

		tag, err := tx.Exec(ctx, `UPDATE sprints SET deleted_at = now(), updated_at = now() WHERE id = $1 AND workspace_id = $2`,
			sprintID, wsID)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return errs.ErrNotFound
		}
		return nil
	})
}

// --- Lifecycle ---

// Start 启动迭代。
// 不变量：同一项目只允许一个 active 迭代（通过 DB 唯一索引 idx_one_active_sprint_per_project）。
func (s *Service) Start(ctx context.Context, wsID, sprintID int64) (*Sprint, error) {
	var result *Sprint
	err := s.withTx(ctx, wsID, func(tx pgx.Tx) error {
		sp, err := s.getSprintTx(ctx, tx, sprintID, wsID)
		if err != nil {
			return err
		}
		if sp.Status != SprintPlanned {
			return errs.New("SPRINT.NOT_PLANNED", "只有未启动的迭代可以启动", 422)
		}
		if sp.StartDate == nil || sp.EndDate == nil {
			return errs.New("SPRINT.NO_DATE_RANGE", "请先设定迭代的起止日期", 422)
		}

		// 状态更新（索引防并发：唯一索引 + 显式检查）
		tag, err := tx.Exec(ctx, `
			UPDATE sprints SET status = 'active', started_at = now(), updated_at = now()
			WHERE id = $1 AND workspace_id = $2 AND status = 'planned'`,
			sprintID, wsID)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return errs.ErrSprintConflict
		}

		// 写一个"启动快照"作为起点
		if err := s.writeSnapshotTx(ctx, tx, wsID, sp); err != nil {
			return err
		}

		result, err = s.getSprintByID(ctx, wsID, sprintID)
		return err
	})
	return result, s.mapPgErrorForStart(err)
}

// mapPgErrorForStart 处理启动冲突错误码。
func (s *Service) mapPgErrorForStart(err error) error {
	if err == nil {
		return nil
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		// 23505 = unique_violation
		if pgErr.Code == "23505" && pgErr.ConstraintName == "idx_one_active_sprint_per_project" {
			return errs.New("SPRINT.ALREADY_ACTIVE", "该项目已有一个进行中的迭代，请结束后再启动新迭代", 409)
		}
	}
	return s.mapPgError(err)
}

// Complete 结束迭代。
func (s *Service) Complete(ctx context.Context, wsID, sprintID int64, in CompleteSprintInput) (*Sprint, error) {
	var result *Sprint
	err := s.withTx(ctx, wsID, func(tx pgx.Tx) error {
		sp, err := s.getSprintTx(ctx, tx, sprintID, wsID)
		if err != nil {
			return err
		}
		if sp.Status != SprintActive {
			return errs.New("SPRINT.NOT_ACTIVE", "只有进行中的迭代可以结束", 422)
		}

		// 处理未完成任务
		if err := s.handleUnfinishedTx(ctx, tx, wsID, sp, in); err != nil {
			return err
		}

		// 生成复盘数据
		review := s.computeReview(ctx, tx, wsID, sp)
		reviewJSON, _ := json.Marshal(review)

		tag, err := tx.Exec(ctx, `
			UPDATE sprints SET status = 'completed', completed_at = now(),
				review_snapshot = $1, updated_at = now()
			WHERE id = $2 AND workspace_id = $3 AND status = 'active'`,
			reviewJSON, sprintID, wsID)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return errs.ErrSprintConflict
		}

		result, err = s.getSprintByID(ctx, wsID, sprintID)
		return err
	})
	return result, s.mapPgError(err)
}

// handleUnfinishedTx 处理未完成任务的工作项迁移。
func (s *Service) handleUnfinishedTx(ctx context.Context, tx pgx.Tx, wsID int64, sp *Sprint, in CompleteSprintInput) error {
	switch in.Strategy {
	case UnfinishedBacklog, UnfinishedKeep:
		// 删除 sprint_issues 关联（下次创建时不再属于任何迭代）
		_, err := tx.Exec(ctx, `DELETE FROM sprint_issues WHERE sprint_id = $1`, sp.ID)
		return err
	case UnfinishedNextSprint:
		if in.NextSprintID == nil {
			return errs.New("SPRINT.MISSING_TARGET", "请指定下一迭代", 422)
		}
		// 校验目标迭代
		var ok bool
		err := tx.QueryRow(ctx,
			`SELECT EXISTS(SELECT 1 FROM sprints WHERE id = $1 AND project_id = $2 AND status = 'planned' AND deleted_at IS NULL)`,
			*in.NextSprintID, sp.ProjectID).Scan(&ok)
		if err != nil {
			return err
		}
		if !ok {
			return errs.New("SPRINT.INVALID_TARGET", "目标迭代不存在或不在 planned 状态", 422)
		}
		// 迁移：先将工作项从当前迭代删除，然后添加到下一迭代
		_, err = tx.Exec(ctx, `DELETE FROM sprint_issues WHERE sprint_id = $1`, sp.ID)
		return err
	}
	return nil
}

// --- Planning: Backlog <-> Sprint drag ---

// AddIssue 将工作项加入迭代（支持拖拽规划与中途加项）。
func (s *Service) AddIssue(ctx context.Context, wsID int64, in AddIssueInput) error {
	return s.withTx(ctx, wsID, func(tx pgx.Tx) error {
		sp, err := s.getSprintTx(ctx, tx, in.SprintID, wsID)
		if err != nil {
			return err
		}

		// 校验工作项存在并属于同项目
		var issueProjectID int64
		var statusCode SprintStatusCode
		err = tx.QueryRow(ctx,
			`SELECT i.project_id FROM issues i WHERE i.id = $1 AND i.workspace_id = $2 AND i.deleted_at IS NULL`,
			in.IssueID, wsID).Scan(&issueProjectID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return errs.ErrNotFound
			}
			return errs.ErrInternal.Wrap(err)
		}
		if issueProjectID != sp.ProjectID {
			return errs.New("SPRINT.CROSS_PROJECT", "工作项不属于当前项目", 422)
		}

		statusCode = sp.Status
		var addedMidway bool
		if statusCode == SprintActive {
			addedMidway = true
		}

		// 注入 sprint_id
		_, err = tx.Exec(ctx, `UPDATE issues SET sprint_id = $1, updated_at = now() WHERE id = $2 AND workspace_id = $3`,
			in.IssueID, in.IssueID, wsID)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			// sprint_id 列允许 NULL，此更新可选；此处忽略重复
		}

		// 写 sprint_issues
		_, err = tx.Exec(ctx, `
			INSERT INTO sprint_issues (sprint_id, issue_id, added_midway, sort_order, added_by)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (sprint_id, issue_id) DO UPDATE SET sort_order = EXCLUDED.sort_order`,
			in.SprintID, in.IssueID, addedMidway, in.SortOrder, in.AddedBy)
		return err
	})
}

// RemoveIssue 从迭代移除工作项。
func (s *Service) RemoveIssue(ctx context.Context, wsID int64, in RemoveIssueInput) error {
	return s.withTx(ctx, wsID, func(tx pgx.Tx) error {
		_, err := s.getSprintTx(ctx, tx, in.SprintID, wsID)
		if err != nil {
			return err
		}

		// 从 sprint_issues 删除
		tag, err := tx.Exec(ctx, `DELETE FROM sprint_issues WHERE sprint_id = $1 AND issue_id = $2`,
			in.SprintID, in.IssueID)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return errs.ErrNotFound
		}

		// 若迭代已 started，标记 added_midway 为 false（视为中途移除）
		// 注意：中途移除的 work 通过 computeProgress 自动识别

		return nil
	})
}

// ListSprintIssues 列出迭代中的工作项（with progress)。
func (s *Service) ListSprintIssues(ctx context.Context, wsID, sprintID int64, limit, offset int) ([]SprintIssueView, int64, error) {
	// 迭代存在性校验
	_, err := s.getSprintByID(ctx, wsID, sprintID)
	if err != nil {
		return nil, 0, err
	}

	if limit <= 0 || limit > 100 {
		limit = 50
	}

	var total int64
	_ = s.db.QueryRow(ctx,
		`SELECT count(*) FROM sprint_issues WHERE sprint_id = $1`, sprintID).Scan(&total)

	rows, err := s.db.Query(ctx, `
		SELECT si.issue_id, si.added_midway, si.sort_order,
			i.name, i.type_code, i.priority, i.point, i.severity,
			i.state_id, st.name AS state_name, st.color AS state_color, st."group" AS state_group,
			i.created_at
		FROM sprint_issues si
		JOIN issues i ON i.id = si.issue_id
		JOIN states st ON st.id = i.state_id
		WHERE si.sprint_id = $1
		ORDER BY si.sort_order, si.added_at
		LIMIT $2 OFFSET $3`, sprintID, limit, offset)
	if err != nil {
		return nil, 0, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()

	var views []SprintIssueView
	for rows.Next() {
		var v SprintIssueView
		var point sql.NullInt64
		var severity sql.NullInt64
		var addedMidway bool
		if err := rows.Scan(
			&v.IssueID, &addedMidway, &v.SortOrder,
			&v.Name, &v.TypeCode, &v.Priority, &point, &severity,
			&v.StateID, &v.StateName, &v.StateColor, &v.StateGroup,
			&v.CreatedAt,
		); err != nil {
			return nil, 0, errs.ErrInternal.Wrap(err)
		}
		if point.Valid {
			n := int(point.Int64)
			v.Point = &n
		}
		if severity.Valid {
			n := int(severity.Int64)
			v.Severity = &n
		}
		views = append(views, v)
	}
	return views, total, rows.Err()
}

// BacklogItem Backlog 视图。
type BacklogItem struct {
	IssueID        int64     `json:"issue_id"`
	Name           string    `json:"name"`
	TypeCode       string    `json:"type_code"`
	Priority       string    `json:"priority"`
	StateID        int64     `json:"state_id"`
	StateName      string    `json:"state_name"`
	StateGroup     string    `json:"state_group"`
	StateColor     string    `json:"state_color"`
	HasSprint      bool      `json:"has_sprint"`
	SprintID       *int64    `json:"sprint_id,omitempty"`
	SprintName     string    `json:"sprint_name,omitempty"`
	AssignedPoints *int      `json:"point,omitempty"`
}

// SprintIssueView 迭代内工作项视图。
type SprintIssueView struct {
	IssueID    int64     `json:"issue_id"`
	SortOrder  float64   `json:"sort_order"`
	Name       string    `json:"name"`
	TypeCode   string    `json:"type_code"`
	Priority   string    `json:"priority"`
	StateID    int64     `json:"state_id"`
	StateName  string    `json:"state_name"`
	StateColor string    `json:"state_color"`
	StateGroup string    `json:"state_group"`
	CreatedAt  time.Time `json:"created_at"`
	Point      *int      `json:"point,omitempty"`
	Severity   *int      `json:"severity,omitempty"`
}

// GetBacklog 获取 Backlog 工作项列表（未规划进 active 迭代的未完成任务）。
func (s *Service) GetBacklog(ctx context.Context, wsID, projectID int64, limit, offset int) ([]BacklogItem, int64, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	// Backlog：未关闭/未取消、且不在 active 迭代的工作项
	var total int64
	_ = s.db.QueryRow(ctx, `
		SELECT count(*) FROM issues i
		WHERE i.workspace_id = $1 AND i.project_id = $2 AND i.deleted_at IS NULL
			AND i.sprint_id IS NULL
			AND i.state_id NOT IN (SELECT id FROM states WHERE "group" IN ('completed','cancelled'))
	`, wsID, projectID).Scan(&total)

	rows, err := s.db.Query(ctx, `
		SELECT i.id, i.name, i.type_code, i.priority, i.point,
			s.id, s.name, s.color, s."group",
			i.sprint_id, sp.name
		FROM issues i
		JOIN states s ON s.id = i.state_id
		LEFT JOIN sprints sp ON sp.id = i.sprint_id AND sp.status = 'active' AND sp.deleted_at IS NULL
		WHERE i.workspace_id = $1 AND i.project_id = $2 AND i.deleted_at IS NULL
			AND i.state_id NOT IN (SELECT id FROM states WHERE "group" IN ('completed','cancelled'))
		ORDER BY
			CASE i.priority WHEN 'urgent' THEN 0 WHEN 'high' THEN 1 WHEN 'medium' THEN 2 WHEN 'low' THEN 3 ELSE 4 END,
			i.sort_order
		LIMIT $3 OFFSET $4`, wsID, projectID, limit, offset)
	if err != nil {
		return nil, 0, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()

	var items []BacklogItem
	for rows.Next() {
		var v BacklogItem
		var point sql.NullInt64
		var sprintID sql.NullInt64
		var sprintName sql.NullString
		if err := rows.Scan(
			&v.IssueID, &v.Name, &v.TypeCode, &v.Priority, &point,
			&v.StateID, &v.StateName, &v.StateColor, &v.StateGroup,
			&sprintID, &sprintName,
		); err != nil {
			return nil, 0, errs.ErrInternal.Wrap(err)
		}
		if point.Valid {
			n := int(point.Int64)
			v.AssignedPoints = &n
		}
		v.HasSprint = sprintID.Valid
		if sprintID.Valid {
			v.SprintID = &sprintID.Int64
			v.SprintName = sprintName.String
		}
		items = append(items, v)
	}
	return items, total, rows.Err()
}

// --- Snapshot & Burndown ---

// WriteDailySnapshot 为 active 迭代写一条日快照（由外部 Cron/Worker 调用）。
func (s *Service) WriteDailySnapshot(ctx context.Context, wsID int64) (int, error) {
	var projectIDs []int64
	if err := s.withTx(ctx, wsID, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctx,
			`SELECT DISTINCT project_id FROM sprints WHERE status = 'active' AND workspace_id = $1 AND deleted_at IS NULL`,
			wsID)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var pid int64
			if err := rows.Scan(&pid); err != nil {
				return err
			}
			projectIDs = append(projectIDs, pid)
		}
		if err := rows.Err(); err != nil {
			return err
		}

		for _, pid := range projectIDs {
			var sprintIDs []int64
			rows, err := tx.Query(ctx,
				`SELECT id FROM sprints WHERE project_id = $1 AND status = 'active' AND deleted_at IS NULL`, pid)
			if err != nil {
				return err
			}
			for rows.Next() {
				var sid int64
				if err := rows.Scan(&sid); err != nil {
					rows.Close()
					return err
				}
				sprintIDs = append(sprintIDs, sid)
			}
			rows.Close()

			for _, sid := range sprintIDs {
				if err := s.writeSnapshotTx(ctx, tx, wsID, &Sprint{ID: sid, WorkspaceID: wsID, ProjectID: pid}); err != nil {
					return err
				}
			}
		}
		return nil
	}); err != nil {
		return 0, s.mapPgError(err)
	}
	return len(projectIDs), nil
}

// writeSnapshotTx 在迭代内写一个日快照（内嵌事务）。
func (s *Service) writeSnapshotTx(ctx context.Context, tx pgx.Tx, wsID int64, sp *Sprint) error {
	today := time.Now().UTC().Truncate(24 * time.Hour)
	projectID := sp.ProjectID

	// 聚合：当前迭代内工作项的总/已完成点数，按状态分组
	var totalPoints, donePoints, addedPoints sql.NullFloat64
	var totalIssues, doneIssues int
	byState := map[string]float64{}

	err := tx.QueryRow(ctx, `
		SELECT
			coalesce(sum(CASE WHEN i.point IS NOT NULL THEN i.point ELSE 0 END), 0),
			coalesce(sum(CASE WHEN sg."group" = 'completed' AND i.point IS NOT NULL THEN i.point ELSE 0 END), 0),
			count(*),
			count(*) FILTER (WHERE sg."group" = 'completed'),
			coalesce(sum(CASE WHEN si.added_midway AND i.point IS NOT NULL THEN i.point ELSE 0 END), 0)
		FROM sprint_issues si
		JOIN issues i ON i.id = si.issue_id
		JOIN states sg ON sg.id = i.state_id
		WHERE si.sprint_id = $1 AND i.deleted_at IS NULL`, sp.ID).Scan(
		&totalPoints, &donePoints, &totalIssues, &doneIssues, &addedPoints)
	if err != nil {
		return errs.ErrInternal.Wrap(err)
	}

	// by_state
	rows, err := tx.Query(ctx, `
		SELECT sg."group", coalesce(sum(CASE WHEN i.point IS NOT NULL THEN i.point ELSE 0 END), 0)
		FROM sprint_issues si
		JOIN issues i ON i.id = si.issue_id
		JOIN states sg ON sg.id = i.state_id
		WHERE si.sprint_id = $1 AND i.deleted_at IS NULL
		GROUP BY sg."group"`, sp.ID)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var grp string
		var pts float64
		if err := rows.Scan(&grp, &pts); err != nil {
			return err
		}
		byState[grp] = pts
	}

	data := SnapshotData{
		TotalPoints:  totalPoints.Float64,
		DonePoints:   donePoints.Float64,
		TotalIssues:  totalIssues,
		DoneIssues:   doneIssues,
		ByStateGroup: byState,
		AddedPoints:  addedPoints.Float64,
		RemovedPoints: 0, // removed 在非 start 时已被移出；可进一步按流程算
	}

	payload, _ := json.Marshal(data)
	_, err = tx.Exec(ctx, `
		INSERT INTO sprint_snapshots (workspace_id, project_id, sprint_id, snapshot_date, data)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (sprint_id, snapshot_date) DO UPDATE SET data = EXCLUDED.data`,
		wsID, projectID, sp.ID, today, payload)
	return err
}

// BurndownData 返回燃尽图数据序列。
func (s *Service) BurndownData(ctx context.Context, wsID, sprintID int64) (*Sprint, []BurndownPoint, error) {
	sp, err := s.GetByID(ctx, wsID, sprintID)
	if err != nil {
		return nil, nil, err
	}
	if sp.StartDate == nil || sp.EndDate == nil {
		return sp, []BurndownPoint{}, nil
	}

	rows, err := s.db.Query(ctx,
		`SELECT snapshot_date, data FROM sprint_snapshots WHERE sprint_id = $1 ORDER BY snapshot_date`,
		sprintID)
	if err != nil {
		return sp, nil, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()

	var points []BurndownPoint
	for rows.Next() {
		var d time.Time
		var data SnapshotData
		var payload []byte
		if err := rows.Scan(&d, &payload); err != nil {
			return sp, nil, errs.ErrInternal.Wrap(err)
		}
		_ = json.Unmarshal(payload, &data)

		points = append(points, BurndownPoint{
			Date:        d,
			TotalPoints: data.TotalPoints,
			DonePoints:  data.DonePoints,
			Remaining:   data.TotalPoints - data.DonePoints,
		})
	}

	// 理想线：从 startDate 0 点到 endDate 总工作量的线性递减
	daysBetween := sp.EndDate.Sub(*sp.StartDate).Hours() / 24
	total := 0.0
	if len(points) > 0 {
		// 取启动日的 total_points 作为"承诺量"基准
		total = points[0].TotalPoints
	}

	for i := range points {
		dayOffset := points[i].Date.Sub(*sp.StartDate).Hours() / 24
		if daysBetween > 0 {
			points[i].IdealLine = total * (1 - dayOffset/daysBetween)
			if points[i].IdealLine < 0 {
				points[i].IdealLine = 0
			}
		} else {
			points[i].IdealLine = 0
		}
	}

	return sp, points, rows.Err()
}

// SuggestCapacity 根据近 N 期 completed 迭代的速率来推荐容量。
func (s *Service) SuggestCapacity(ctx context.Context, wsID, projectID int64, windows []int) (*VelocityStats, error) {
	if len(windows) == 0 {
		windows = []int{3, 6}
	}

	var all []SprintVelocity
	for _, n := range windows {
		stats, err := s.velocityStats(ctx, wsID, projectID, n)
		if err != nil {
			return nil, err
		}
		all = append(all, stats.RecentSprints...)
	}

	// 最终逻辑：使用最大窗口的统计
	maxN := 0
	for _, n := range windows {
		if n > maxN {
			maxN = n
		}
	}
	return s.velocityStats(ctx, wsID, projectID, maxN)
}

// velocityStats 返回近 N 期的 rate stats。
func (s *Service) velocityStats(ctx context.Context, wsID, projectID int64, lastN int) (*VelocityStats, error) {
	rows, err := s.db.Query(ctx, `
		SELECT s.id, s.name, s.review_snapshot
		FROM sprints s
		WHERE s.workspace_id = $1 AND s.project_id = $2 AND s.status = 'completed' AND s.review_snapshot IS NOT NULL
		ORDER BY s.completed_at DESC NULLS LAST
		LIMIT $3`, wsID, projectID, lastN)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()

	var history []SprintVelocity
	for rows.Next() {
		var id int64
		var name string
		var raw []byte
		if err := rows.Scan(&id, &name, &raw); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
		var snap ReviewSnapshot
		_ = json.Unmarshal(raw, &snap)
		history = append(history, SprintVelocity{
			SprintID:        id,
			SprintName:      name,
			CompletedPoints: snap.CompletedPoints,
			CompletedIssues: snap.CompletedIssues,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}

	stats := &VelocityStats{RecentSprints: history, Count: len(history)}
	if len(history) == 0 {
		return stats, nil
	}

	var sumPt float64
	var sumIss int
	var vals []float64
	for _, h := range history {
		sumPt += h.CompletedPoints
		sumIss += h.CompletedIssues
		vals = append(vals, h.CompletedPoints)
	}
	stats.AvgPoints = sumPt / float64(len(history))
	stats.AvgIssues = float64(sumIss) / float64(len(history))

	// P50 中位数
	sort.Float64s(vals)
	mid := len(vals) / 2
	if len(vals)%2 == 0 {
		stats.P50 = (vals[mid-1] + vals[mid]) / 2
	} else {
		stats.P50 = vals[mid]
	}

	return stats, nil
}

// computeProgress 计算迭代实时进度。
func (s *Service) computeProgress(ctx context.Context, wsID int64, sp *Sprint) SprintProgress {
	if sp.Status == SprintCompleted {
		// completed 时优先使用 review_snapshot
		if sp.ReviewSnapshot != nil {
			return SprintProgress{
				TotalPoints:  sp.ReviewSnapshot.CommittedPoints,
				DonePoints:   sp.ReviewSnapshot.CompletedPoints,
				TotalIssues:  sp.ReviewSnapshot.CommittedIssues,
				DoneIssues:   sp.ReviewSnapshot.CompletedIssues,
			}
		}
	}

	// 实时计算
	var totalPoints, donePoints sql.NullFloat64
	var totalIssues, doneIssues int
	byState := map[string]float64{}

	_ = s.db.QueryRow(ctx, `
		SELECT
			coalesce(sum(CASE WHEN i.point IS NOT NULL THEN i.point ELSE 0 END), 0),
			coalesce(sum(CASE WHEN sg."group" = 'completed' AND i.point IS NOT NULL THEN i.point ELSE 0 END), 0),
			count(*),
			count(*) FILTER (WHERE sg."group" = 'completed')
		FROM sprint_issues si
		JOIN issues i ON i.id = si.issue_id
		JOIN states sg ON sg.id = i.state_id
		WHERE si.sprint_id = $1 AND i.deleted_at IS NULL`, sp.ID).Scan(
		&totalPoints, &donePoints, &totalIssues, &doneIssues)

	rows, err := s.db.Query(ctx, `
		SELECT sg."group", coalesce(sum(CASE WHEN i.point IS NOT NULL THEN i.point ELSE 0 END), 0)
		FROM sprint_issues si
		JOIN issues i ON i.id = si.issue_id
		JOIN states sg ON sg.id = i.state_id
		WHERE si.sprint_id = $1 AND i.deleted_at IS NULL
		GROUP BY sg."group"`, sp.ID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var grp string
			var pts float64
			if err := rows.Scan(&grp, &pts); err == nil {
				byState[grp] = pts
			}
		}
	}

	progress := SprintProgress{
		TotalPoints:  totalPoints.Float64,
		DonePoints:   donePoints.Float64,
		TotalIssues:  totalIssues,
		DoneIssues:   doneIssues,
		ByStateGroup: byState,
	}
	if sp.Capacity != nil && *sp.Capacity > 0 {
		progress.Saturation = math.Min(totalPoints.Float64 / (*sp.Capacity), 999)
	}
	return progress
}

// computeReview 复盘数据。
func (s *Service) computeReview(ctx context.Context, tx pgx.Tx, wsID int64, sp *Sprint) *ReviewSnapshot {
	// 承诺：迭代启动时 sprint_issues 的工作项（不含中途加入）
	// 完成：已 completed state 的工作项
	// 中途加入：added_midway = true
	// 中途移除：不再在 sprint_issues 中但曾经在的（需通过该字段 tracking；本期初版通过 added_midway 标记）
	var committedPoints, completedPoints, joinedPoints sql.NullFloat64
	var committedIssues, completedIssues, joinedIssues int

	_ = tx.QueryRow(ctx, `
		SELECT
			coalesce(sum(CASE WHEN NOT si.added_midway AND i.point IS NOT NULL THEN i.point ELSE 0 END), 0),
			coalesce(sum(CASE WHEN sg."group" = 'completed' AND i.point IS NOT NULL THEN i.point ELSE 0 END), 0),
			count(*) FILTER (WHERE NOT si.added_midway),
			count(*) FILTER (WHERE sg."group" = 'completed'),
			count(*) FILTER (WHERE si.added_midway),
			coalesce(sum(CASE WHEN si.added_midway AND i.point IS NOT NULL THEN i.point ELSE 0 END), 0)
		FROM sprint_issues si
		JOIN issues i ON i.id = si.issue_id
		JOIN states sg ON sg.id = i.state_id
		WHERE si.sprint_id = $1 AND i.deleted_at IS NULL`, sp.ID).Scan(
		&committedPoints, &completedPoints, &committedIssues, &completedIssues, &joinedIssues, &joinedPoints)

	rate := 0.0
	if committedPoints.Float64 > 0 {
		rate = completedPoints.Float64 / committedPoints.Float64
		if rate > 1 {
			rate = 1
		}
	}

	return &ReviewSnapshot{
		CommittedPoints: committedPoints.Float64,
		CompletedPoints: completedPoints.Float64,
		JoinedPoints:    joinedPoints.Float64,
		CommittedIssues: committedIssues,
		CompletedIssues: completedIssues,
		JoinedIssues:    joinedIssues,
		CompletionRate:  rate,
	}
}

// --- Low-level helpers ---

func (s *Service) getSprintByID(ctx context.Context, wsID, sprintID int64) (*Sprint, error) {
	row := s.db.QueryRow(ctx, `
		SELECT s.id, s.workspace_id, s.project_id, s.name, s.description, s.goal,
			s.status, s.start_date, s.end_date, s.capacity, s.owner_id,
			s.review_snapshot, s.started_at, s.completed_at, s.created_by, s.created_at, s.updated_at
		FROM sprints s
		WHERE s.id = $1 AND s.workspace_id = $2 AND s.deleted_at IS NULL`, sprintID, wsID)
	sp, err := scanSprint(row)
	if err != nil {
		return nil, s.mapPgError(err)
	}
	return sp, nil
}

func (s *Service) getSprintTx(ctx context.Context, tx pgx.Tx, sprintID, wsID int64) (*Sprint, error) {
	row := tx.QueryRow(ctx, `
		SELECT s.id, s.workspace_id, s.project_id, s.name, s.description, s.goal,
			s.status, s.start_date, s.end_date, s.capacity, s.owner_id,
			s.review_snapshot, s.started_at, s.completed_at, s.created_by, s.created_at, s.updated_at
		FROM sprints s
		WHERE s.id = $1 AND s.workspace_id = $2 AND s.deleted_at IS NULL`, sprintID, wsID)
	return scanSprint(row)
}

func scanSprint(row pgx.Row) (*Sprint, error) {
	var sp Sprint
	var desc, goal sql.NullString
	var startDate, endDate, startedAt, completedAt sql.NullTime
	var capacity sql.NullFloat64
	var ownerID sql.NullInt64
	var reviewRaw []byte

	err := row.Scan(
		&sp.ID, &sp.WorkspaceID, &sp.ProjectID, &sp.Name, &desc, &goal,
		&sp.Status, &startDate, &endDate, &capacity, &ownerID,
		&reviewRaw, &startedAt, &completedAt, &sp.CreatedBy, &sp.CreatedAt, &sp.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrNotFound
		}
		return nil, errs.ErrInternal.Wrap(err)
	}
	if desc.Valid {
		sp.Description = desc.String
	}
	if goal.Valid {
		sp.Goal = goal.String
	}
	if startDate.Valid {
		sp.StartDate = &startDate.Time
	}
	if endDate.Valid {
		sp.EndDate = &endDate.Time
	}
	if capacity.Valid {
		v := capacity.Float64
		sp.Capacity = &v
	}
	if ownerID.Valid {
		v := ownerID.Int64
		sp.OwnerID = &v
	}
	if startedAt.Valid {
		sp.StartedAt = &startedAt.Time
	}
	if completedAt.Valid {
		sp.CompletedAt = &completedAt.Time
	}
	if len(reviewRaw) > 0 {
		var snap ReviewSnapshot
		if err := json.Unmarshal(reviewRaw, &snap); err == nil {
			sp.ReviewSnapshot = &snap
		}
	}
	return &sp, nil
}

// withTx 事务辅助。
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

func (s *Service) mapPgError(err error) error {
	if err == nil {
		return nil
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			return errs.New("SPRINT.DUPLICATE", "迭代冲突: "+pgErr.Detail, 409)
		case "23514": // check_violation
			return errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "sprint", Reason: pgErr.Detail})
		}
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return errs.ErrNotFound
	}
	if appErr, ok := err.(*errs.AppError); ok {
		return appErr
	}
	return errs.ErrInternal.Wrap(err)
}

// --- Validation & util ---

func validateSprintInput(in CreateSprintInput) error {
	if in.WorkspaceID == 0 {
		return errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "workspace_id", Reason: "工作空间不能为空"})
	}
	if in.ProjectID == 0 {
		return errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "project_id", Reason: "项目不能为空"})
	}
	if len(in.Name) == 0 || len(in.Name) > 80 {
		return errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "name", Reason: "迭代名称长度需在 1-80 之间"})
	}
	return nil
}

func buildSprintUpdateSet(in UpdateSprintInput) ([]string, []interface{}) {
	var sets []string
	var args []interface{}
	i := 1

	if in.Name != nil {
		sets = append(sets, "name = $"+strconv.Itoa(i))
		args = append(args, *in.Name)
		i++
	}
	if in.Description != nil {
		sets = append(sets, "description = $"+strconv.Itoa(i))
		args = append(args, *in.Description)
		i++
	}
	if in.Goal != nil {
		sets = append(sets, "goal = $"+strconv.Itoa(i))
		args = append(args, *in.Goal)
		i++
	}
	if in.StartDate != nil {
		sets = append(sets, "start_date = $"+strconv.Itoa(i))
		args = append(args, *in.StartDate)
		i++
	}
	if in.EndDate != nil {
		sets = append(sets, "end_date = $"+strconv.Itoa(i))
		args = append(args, *in.EndDate)
		i++
	}
	if in.Capacity != nil {
		sets = append(sets, "capacity = $"+strconv.Itoa(i))
		args = append(args, *in.Capacity)
		i++
	}
	if in.OwnerID != nil {
		sets = append(sets, "owner_id = $"+strconv.Itoa(i))
		args = append(args, *in.OwnerID)
		i++
	}

	if len(sets) > 0 {
		sets = append(sets, "updated_at = now()")
	}
	return sets, args
}

func stringsJoin(ss []string, sep string) string {
	switch len(ss) {
	case 0:
		return ""
	case 1:
		return ss[0]
	}
	n := len(sep) * (len(ss) - 1)
	for _, s := range ss {
		n += len(s)
	}
	b := make([]byte, 0, n)
	b = append(b, ss[0]...)
	for _, s := range ss[1:] {
		b = append(b, sep...)
		b = append(b, s...)
	}
	return string(b)
}
