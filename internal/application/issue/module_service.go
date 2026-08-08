// Package issue — Module 模块体系（项目内工作项分组机制）。
//
// 对标 Plane 的 Module 概念：模块是项目内工作项的逻辑分组，
// 通常对应功能模块、组件或子系统。一个工作项可以属于多个模块，
// 一个模块包含多个工作项（M:N 关系）。
//
// 模块属性：名称、描述、负责人、状态（活跃/已完成/归档）、起止日期、排序权重。
package issue

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// ModuleService 管理项目模块。
type ModuleService struct {
	db *pgxpool.Pool
}

// NewModuleService 创建模块服务。
func NewModuleService(db *pgxpool.Pool) *ModuleService {
	return &ModuleService{db: db}
}

// ---- 输入参数 ----

// CreateModuleInput 创建模块入参。
type CreateModuleInput struct {
	WorkspaceID int64
	ProjectID   int64
	Name        string
	Description string
	LeadID      *int64
	StartDate   *time.Time
	TargetDate  *time.Time
	SortOrder   float64
	CreatedBy   int64
}

// UpdateModuleInput 更新模块入参。
type UpdateModuleInput struct {
	ID          int64
	WorkspaceID int64
	ProjectID   int64
	Name        *string
	Description *string
	LeadID      *int64
	Status      *string
	StartDate   *time.Time
	TargetDate  *time.Time
	SortOrder   *float64
}

// ListModulesFilter 模块列表筛选。
type ListModulesFilter struct {
	WorkspaceID int64
	ProjectID   int64
	Status      string // 空=全部
}

// AssignIssuesInput 分配工作项到模块。
type AssignIssuesInput struct {
	ModuleID int64
	IssueIDs []int64
	CreatedBy int64
}

// ---- CRUD ----

// CreateModule 创建模块。
func (s *ModuleService) CreateModule(ctx context.Context, in CreateModuleInput) (*Module, error) {
	if strings.TrimSpace(in.Name) == "" {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "name", Reason: "模块名称不能为空"})
	}
	if len(in.Name) > 120 {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "name", Reason: "模块名称不能超过 120 字符"})
	}

	var m Module
	err := s.db.QueryRow(ctx, `
		INSERT INTO modules (workspace_id, project_id, name, description, lead_id, start_date, target_date, sort_order, created_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		RETURNING id, public_id, status, created_at, updated_at`,
		in.WorkspaceID, in.ProjectID, in.Name, nullIf(in.Description, ""),
		in.LeadID, in.StartDate, in.TargetDate,
		defaultSortOrder(in.SortOrder), in.CreatedBy,
	).Scan(&m.ID, &m.PublicID, &m.Status, &m.CreatedAt, &m.UpdatedAt)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}

	m.WorkspaceID = in.WorkspaceID
	m.ProjectID = in.ProjectID
	m.Name = in.Name
	m.Description = in.Description
	m.LeadID = in.LeadID
	m.StartDate = in.StartDate
	m.TargetDate = in.TargetDate
	m.SortOrder = in.SortOrder
	m.CreatedBy = in.CreatedBy
	return &m, nil
}

// GetModule 获取单个模块。
func (s *ModuleService) GetModule(ctx context.Context, moduleID int64, wsID int64) (*Module, error) {
	var m Module
	err := s.db.QueryRow(ctx, `
		SELECT id, public_id, workspace_id, project_id, name, description, lead_id, status,
		       start_date, target_date, sort_order, created_by, created_at, updated_at
		FROM modules WHERE id = $1 AND workspace_id = $2 AND deleted_at IS NULL`,
		moduleID, wsID,
	).Scan(&m.ID, &m.PublicID, &m.WorkspaceID, &m.ProjectID, &m.Name,
		&m.Description, &m.LeadID, &m.Status, &m.StartDate, &m.TargetDate,
		&m.SortOrder, &m.CreatedBy, &m.CreatedAt, &m.UpdatedAt)
	if err != nil {
		return nil, errs.ErrNotFound
	}
	return &m, nil
}

// ListModules 列出项目模块。
func (s *ModuleService) ListModules(ctx context.Context, f ListModulesFilter) ([]Module, error) {
	var args []interface{}
	var conds []string
	args = append(args, f.WorkspaceID, f.ProjectID)
	conds = append(conds, "workspace_id = $1", "project_id = $2", "deleted_at IS NULL")

	if f.Status != "" {
		args = append(args, f.Status)
		conds = append(conds, "status = $"+strconv.Itoa(len(args)))
	}

	rows, err := s.db.Query(ctx,
		`SELECT id, public_id, workspace_id, project_id, name, description, lead_id, status,
		        start_date, target_date, sort_order, created_by, created_at, updated_at
		 FROM modules WHERE `+strings.Join(conds, " AND ")+`
		 ORDER BY sort_order ASC, id ASC`, args...)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()

	var out []Module
	for rows.Next() {
		var m Module
		if err := rows.Scan(&m.ID, &m.PublicID, &m.WorkspaceID, &m.ProjectID, &m.Name,
			&m.Description, &m.LeadID, &m.Status, &m.StartDate, &m.TargetDate,
			&m.SortOrder, &m.CreatedBy, &m.CreatedAt, &m.UpdatedAt); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

// UpdateModule 更新模块。
func (s *ModuleService) UpdateModule(ctx context.Context, in UpdateModuleInput) (*Module, error) {
	var m Module
	err := s.db.QueryRow(ctx, `
		UPDATE modules SET
			name = COALESCE($4, name),
			description = COALESCE($5, description),
			lead_id = COALESCE($6, lead_id),
			status = COALESCE($7, status),
			start_date = COALESCE($8, start_date),
			target_date = COALESCE($9, target_date),
			sort_order = COALESCE($10, sort_order)
		WHERE id = $1 AND workspace_id = $2 AND project_id = $3 AND deleted_at IS NULL
		RETURNING id, public_id, workspace_id, project_id, name, description, lead_id, status,
		          start_date, target_date, sort_order, created_by, created_at, updated_at`,
		in.ID, in.WorkspaceID, in.ProjectID,
		in.Name, in.Description, in.LeadID, in.Status,
		in.StartDate, in.TargetDate, in.SortOrder,
	).Scan(&m.ID, &m.PublicID, &m.WorkspaceID, &m.ProjectID, &m.Name,
		&m.Description, &m.LeadID, &m.Status, &m.StartDate, &m.TargetDate,
		&m.SortOrder, &m.CreatedBy, &m.CreatedAt, &m.UpdatedAt)
	if err != nil {
		return nil, errs.ErrNotFound
	}
	return &m, nil
}

// DeleteModule 软删除模块。
func (s *ModuleService) DeleteModule(ctx context.Context, moduleID int64, wsID int64, projectID int64) error {
	cmd, err := s.db.Exec(ctx,
		`UPDATE modules SET deleted_at = now() WHERE id = $1 AND workspace_id = $2 AND project_id = $3 AND deleted_at IS NULL`,
		moduleID, wsID, projectID)
	if err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	if cmd.RowsAffected() == 0 {
		return errs.ErrNotFound
	}
	return nil
}

// ---- 工作项-模块关联 ----

// AssignIssues 分配工作项到模块（幂等）。
func (s *ModuleService) AssignIssues(ctx context.Context, in AssignIssuesInput) error {
	for _, issueID := range in.IssueIDs {
		_, err := s.db.Exec(ctx, `
			INSERT INTO module_issues (module_id, issue_id) VALUES ($1,$2)
			ON CONFLICT (module_id, issue_id) DO NOTHING`,
			in.ModuleID, issueID)
		if err != nil {
			return errs.ErrInternal.Wrap(err)
		}
	}
	return nil
}

// UnassignIssue 从模块移除工作项。
func (s *ModuleService) UnassignIssue(ctx context.Context, moduleID int64, issueID int64) error {
	_, err := s.db.Exec(ctx,
		`DELETE FROM module_issues WHERE module_id = $1 AND issue_id = $2`,
		moduleID, issueID)
	return err
}

// ListModuleIssues 列出模块下的工作项 ID。
func (s *ModuleService) ListModuleIssues(ctx context.Context, moduleID int64) ([]int64, error) {
	rows, err := s.db.Query(ctx,
		`SELECT issue_id FROM module_issues WHERE module_id = $1`, moduleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// ---- helpers ----

func nullIf(s string, empty string) *string {
	if s == empty {
		return nil
	}
	return &s
}

func defaultSortOrder(v float64) float64 {
	if v == 0 {
		return 65535
	}
	return v
}
