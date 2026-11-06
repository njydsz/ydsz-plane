// Package milestone — 项目里程碑应用服务。
package milestone

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// Service 里程碑应用服务。
type Service struct {
	db *pgxpool.Pool
}

// NewService 创建服务。
func NewService(db *pgxpool.Pool) *Service {
	return &Service{db: db}
}

// --- 错误 ---

var (
	// ErrMilestoneNotFound 里程碑不存在。
	ErrMilestoneNotFound = errors.New("milestone not found")
	// ErrWorkItemNotFound 关联的工作项不存在或已删除。
	ErrWorkItemNotFound = errors.New("work item not found")
	// ErrInvalidWorkItemType 工作项类型非法。
	ErrInvalidWorkItemType = errors.New("invalid work item type")
	// ErrAlreadyLinked 工作项已关联到该里程碑。
	ErrAlreadyLinked = errors.New("work item already linked")
)

const milestoneColumns = `id, workspace_id, project_id, name,
	COALESCE(description,''), target_date, status, created_by, created_at, updated_at`

// Create 创建里程碑。
func (s *Service) Create(ctx context.Context, in CreateInput) (*Milestone, error) {
	if strings.TrimSpace(in.Name) == "" {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "name", Reason: "名称不能为空"})
	}
	status := in.Status
	if status == "" {
		status = StatusNotStarted
	}
	if !status.IsValid() {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "status", Reason: "无效的状态值"})
	}

	var m Milestone
	err := scanMilestone(s.db.QueryRow(ctx, `
		INSERT INTO milestones (id, workspace_id, project_id, name, description, target_date, status, created_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		RETURNING `+milestoneColumns,
		genID(), in.WorkspaceID, in.ProjectID, in.Name, nullIfEmpty(in.Description),
		in.TargetDate, string(status), in.CreatedBy,
	), &m)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(fmt.Errorf("milestone.Create: %w", err))
	}
	return &m, nil
}

// Get 查询里程碑详情。
func (s *Service) Get(ctx context.Context, milestoneID int64) (*Milestone, error) {
	var m Milestone
	if err := scanMilestone(s.db.QueryRow(ctx, `
		SELECT `+milestoneColumns+` FROM milestones WHERE id = $1`, milestoneID), &m); err != nil {
		return nil, err
	}
	return &m, nil
}

// List 列出项目下的里程碑。
func (s *Service) List(ctx context.Context, opts ListOptions) ([]Milestone, error) {
	where := []string{"workspace_id = $1", "project_id = $2"}
	args := []any{opts.WorkspaceID, opts.ProjectID}
	idx := 3
	if opts.Status != "" {
		where = append(where, fmt.Sprintf("status = $%d", idx))
		args = append(args, string(opts.Status))
		idx++
	}

	limit := opts.Limit
	if limit <= 0 || limit > 200 {
		limit = 100
	}
	offset := opts.Offset
	if offset < 0 {
		offset = 0
	}
	args = append(args, limit, offset)

	query := `SELECT ` + milestoneColumns + ` FROM milestones
		WHERE ` + strings.Join(where, " AND ") + `
		ORDER BY COALESCE(target_date, '9999-12-31'::date) ASC, created_at DESC
		LIMIT $` + fmt.Sprint(idx) + ` OFFSET $` + fmt.Sprint(idx+1)

	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(fmt.Errorf("milestone.List: %w", err))
	}
	defer rows.Close()

	out := make([]Milestone, 0)
	for rows.Next() {
		var m Milestone
		if err := scanMilestone(rows, &m); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

// Update 更新里程碑。
func (s *Service) Update(ctx context.Context, milestoneID int64, in UpdateInput) (*Milestone, error) {
	setCols := make([]string, 0, 4)
	args := make([]any, 0, 5)
	idx := 1

	if in.Name != nil {
		if strings.TrimSpace(*in.Name) == "" {
			return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "name", Reason: "名称不能为空"})
		}
		setCols = append(setCols, fmt.Sprintf("name = $%d", idx))
		args = append(args, *in.Name)
		idx++
	}
	if in.Description != nil {
		setCols = append(setCols, fmt.Sprintf("description = $%d", idx))
		args = append(args, nullIfEmpty(*in.Description))
		idx++
	}
	if in.TargetDate != nil {
		setCols = append(setCols, fmt.Sprintf("target_date = $%d", idx))
		args = append(args, *in.TargetDate)
		idx++
	}
	if in.Status != nil {
		if !in.Status.IsValid() {
			return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "status", Reason: "无效的状态值"})
		}
		setCols = append(setCols, fmt.Sprintf("status = $%d", idx))
		args = append(args, string(*in.Status))
		idx++
	}

	if len(setCols) == 0 {
		return s.Get(ctx, milestoneID)
	}

	setCols = append(setCols, "updated_at = now()")
	args = append(args, milestoneID)
	query := fmt.Sprintf(`UPDATE milestones SET %s WHERE id = $%d RETURNING %s`,
		strings.Join(setCols, ", "), idx, milestoneColumns)

	var m Milestone
	if err := scanMilestone(s.db.QueryRow(ctx, query, args...), &m); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrMilestoneNotFound
		}
		return nil, errs.ErrInternal.Wrap(fmt.Errorf("milestone.Update: %w", err))
	}
	return &m, nil
}

// Delete 删除里程碑（并清理关联）。
func (s *Service) Delete(ctx context.Context, milestoneID int64) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return errs.ErrInternal.Wrap(fmt.Errorf("milestone.Delete begin: %w", err))
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `DELETE FROM milestone_work_items WHERE milestone_id = $1`, milestoneID); err != nil {
		return errs.ErrInternal.Wrap(fmt.Errorf("milestone.Delete links: %w", err))
	}
	tag, err := tx.Exec(ctx, `DELETE FROM milestones WHERE id = $1`, milestoneID)
	if err != nil {
		return errs.ErrInternal.Wrap(fmt.Errorf("milestone.Delete: %w", err))
	}
	if tag.RowsAffected() == 0 {
		return ErrMilestoneNotFound
	}
	if err := tx.Commit(ctx); err != nil {
		return errs.ErrInternal.Wrap(fmt.Errorf("milestone.Delete commit: %w", err))
	}
	return nil
}

// LinkWorkItem 关联工作项到里程碑（幂等）。
func (s *Service) LinkWorkItem(ctx context.Context, in LinkInput) error {
	if !isValidWorkItemType(in.WorkItemType) {
		return errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "work_item_type", Reason: "必须是 task、requirement 或 defect"})
	}
	// 校验里程碑存在
	if _, err := s.Get(ctx, in.MilestoneID); err != nil {
		return err
	}
	// 校验工作项存在且未删除
	var exists bool
	if err := s.db.QueryRow(ctx, fmt.Sprintf(
		`SELECT EXISTS(SELECT 1 FROM %s WHERE id = $1 AND deleted = false)`, in.WorkItemType),
		in.WorkItemID).Scan(&exists); err != nil {
		return errs.ErrInternal.Wrap(fmt.Errorf("milestone.LinkWorkItem check: %w", err))
	}
	if !exists {
		return errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "work_item_id", Reason: "工作项不存在或已删除"})
	}

	tag, err := s.db.Exec(ctx, `
		INSERT INTO milestone_work_items (milestone_id, work_item_id, work_item_type)
		VALUES ($1,$2,$3)
		ON CONFLICT (milestone_id, work_item_id) DO NOTHING`,
		in.MilestoneID, in.WorkItemID, in.WorkItemType)
	if err != nil {
		return errs.ErrInternal.Wrap(fmt.Errorf("milestone.LinkWorkItem: %w", err))
	}
	if tag.RowsAffected() == 0 {
		return ErrAlreadyLinked
	}
	return nil
}

// UnlinkWorkItem 解除工作项与里程碑的关联。
func (s *Service) UnlinkWorkItem(ctx context.Context, milestoneID, workItemID int64) error {
	tag, err := s.db.Exec(ctx, `
		DELETE FROM milestone_work_items WHERE milestone_id = $1 AND work_item_id = $2`,
		milestoneID, workItemID)
	if err != nil {
		return errs.ErrInternal.Wrap(fmt.Errorf("milestone.UnlinkWorkItem: %w", err))
	}
	if tag.RowsAffected() == 0 {
		return errs.ErrNotFound
	}
	return nil
}

// ListWorkItems 列出里程碑关联的工作项。
func (s *Service) ListWorkItems(ctx context.Context, milestoneID int64) ([]MilestoneWorkItem, error) {
	rows, err := s.db.Query(ctx, `
		SELECT milestone_id, work_item_id, work_item_type, created_at
		FROM milestone_work_items WHERE milestone_id = $1
		ORDER BY created_at ASC`, milestoneID)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(fmt.Errorf("milestone.ListWorkItems: %w", err))
	}
	defer rows.Close()

	out := make([]MilestoneWorkItem, 0)
	for rows.Next() {
		var w MilestoneWorkItem
		if err := rows.Scan(&w.MilestoneID, &w.WorkItemID, &w.WorkItemType, &w.CreatedAt); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
		out = append(out, w)
	}
	return out, rows.Err()
}

// GetProgress 聚合里程碑下关联工作项的进度。
//
// 通过 UNION 跨 task/requirement/defect 三张表，再 JOIN states 取状态组，
// 按 group 分类计数：done→completed、in_progress→in_progress、其余→not_started。
func (s *Service) GetProgress(ctx context.Context, milestoneID int64) (*MilestoneProgress, error) {
	var p MilestoneProgress
	p.ComputedAt = time.Now().UTC()

	rows, err := s.db.Query(ctx, `
		SELECT s."group", COUNT(*) AS cnt
		FROM milestone_work_items mwi
		JOIN (
		    SELECT id, state_id, deleted FROM task
		    UNION ALL
		    SELECT id, state_id, deleted FROM requirement
		    UNION ALL
		    SELECT id, state_id, deleted FROM defect
		) i ON i.id = mwi.work_item_id AND i.deleted = false
		JOIN states s ON s.id = i.state_id
		WHERE mwi.milestone_id = $1
		GROUP BY s."group"`, milestoneID)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(fmt.Errorf("milestone.GetProgress: %w", err))
	}
	defer rows.Close()

	for rows.Next() {
		var group string
		var cnt int
		if err := rows.Scan(&group, &cnt); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
		switch group {
		case "done":
			p.Completed += cnt
		case "in_progress":
			p.InProgress += cnt
		default:
			p.NotStarted += cnt
		}
	}
	if err := rows.Err(); err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}

	p.Total = p.Completed + p.InProgress + p.NotStarted
	if p.Total > 0 {
		p.Percent = float64(p.Completed) / float64(p.Total) * 100
		// 保留 2 位小数
		p.Percent = float64(int(p.Percent*100+0.5)) / 100
	}
	return &p, nil
}

// --- helpers ---

func isValidWorkItemType(t string) bool {
	switch t {
	case "task", "requirement", "defect":
		return true
	default:
		return false
	}
}

func nullIfEmpty(s string) any {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return s
}

func scanMilestone(row pgx.Row, m *Milestone) error {
	var targetDate sql.NullTime
	var status string
	err := row.Scan(
		&m.ID, &m.WorkspaceID, &m.ProjectID, &m.Name, &m.Description,
		&targetDate, &status, &m.CreatedBy, &m.CreatedAt, &m.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrMilestoneNotFound
		}
		return errs.ErrInternal.Wrap(err)
	}
	if targetDate.Valid {
		t := targetDate.Time
		m.TargetDate = &t
	}
	m.Status = MilestoneStatus(status)
	return nil
}
