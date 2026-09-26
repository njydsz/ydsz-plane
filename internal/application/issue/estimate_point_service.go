// Package issue — Estimate Point 估算点数管理服务。
//
// 估算点数是项目级的工作量对标参照系（如斐波那契 1/2/3/5/8/13 或 T 恤码 S/M/L/XL）。
// 项目可自定义多套估算点集，工作项创建/编辑时可从中选择。
package issue

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// EstimatePoint 估算点数实体。
type EstimatePoint struct {
	ID          int64           `json:"id"`
	Code        string          `json:"code"`
	Name        string          `json:"name"`
	WorkspaceID int64           `json:"workspace_id"`
	ProjectID   int64           `json:"project_id"`
	Description string          `json:"description"`
	Points      json.RawMessage `json:"points"` // JSON 数组 [{label, value, color}, ...]
	IsDefault   bool            `json:"is_default"`
	Status      string          `json:"status"`
	CreatedBy   int64           `json:"created_by"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

// EstimatePointService 估算点数服务。
type EstimatePointService struct {
	db *pgxpool.Pool
}

// NewEstimatePointService 创建估算点数服务。
func NewEstimatePointService(db *pgxpool.Pool) *EstimatePointService {
	return &EstimatePointService{db: db}
}

// CreateInput 创建估算点数入参。
type CreateInput struct {
	WorkspaceID int64
	ProjectID   int64
	Name        string
	Description string
	Points      json.RawMessage
	IsDefault   bool
	CreatedBy   int64
}

// UpdateInput 更新估算点数入参。
type UpdateInput struct {
	ID          int64
	WorkspaceID int64
	ProjectID   int64
	Name        *string
	Description *string
	Points      json.RawMessage
	HasPoints   bool
	IsDefault   *bool
}

// ListFilter 列表筛选。
type ListFilter struct {
	WorkspaceID int64
	ProjectID   int64
	Status      string
}

// Create 创建估算点数。
func (s *EstimatePointService) Create(ctx context.Context, in CreateInput) (*EstimatePoint, error) {
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "name", Reason: "名称不能为空"})
	}
	if len([]rune(name)) > 120 {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "name", Reason: "名称不能超过 120 字符"})
	}

	var ep EstimatePoint
	err := s.db.QueryRow(ctx, `
		INSERT INTO estimate_points (id, workspace_id, project_id, name, description, points_config, is_default, status, created_by)
		VALUES ($1,$2,$3,$4,$5,coalesce($6,'[]'::jsonb),$7,'active',$8)
		RETURNING id, coalesce(code,''), name, workspace_id,
		          coalesce(description,''), coalesce(points_config,'[]'::jsonb),
		          is_default, status, created_by, created_at, updated_at`,
		genLabelID(), in.WorkspaceID, in.ProjectID, name,
		in.Description, in.Points, in.IsDefault, in.CreatedBy,
	).Scan(&ep.ID, &ep.Code, &ep.Name, &ep.WorkspaceID,
		&ep.Description, &ep.Points, &ep.IsDefault, &ep.Status,
		&ep.CreatedBy, &ep.CreatedAt, &ep.UpdatedAt)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	return &ep, nil
}

// Get 获取单个估算点数。
func (s *EstimatePointService) Get(ctx context.Context, wsID, id int64) (*EstimatePoint, error) {
	var ep EstimatePoint
	err := s.db.QueryRow(ctx, `
		SELECT id, coalesce(code,''), name, workspace_id,
		       coalesce(description,''), coalesce(points_config,'[]'::jsonb),
		       is_default, status, created_by, created_at, updated_at
		FROM estimate_points WHERE id=$1 AND workspace_id=$2 AND status<>'archived'`,
		id, wsID,
	).Scan(&ep.ID, &ep.Code, &ep.Name, &ep.WorkspaceID,
		&ep.Description, &ep.Points, &ep.IsDefault, &ep.Status,
		&ep.CreatedBy, &ep.CreatedAt, &ep.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errs.ErrNotFound.From()
		}
		return nil, errs.ErrInternal.Wrap(err)
	}
	return &ep, nil
}

// List 列出项目估算点数。
func (s *EstimatePointService) List(ctx context.Context, f ListFilter) ([]EstimatePoint, error) {
	query := `
		SELECT id, coalesce(code,''), name, workspace_id,
		       coalesce(description,''), coalesce(points_config,'[]'::jsonb),
		       is_default, status, created_by, created_at, updated_at
		FROM estimate_points
		WHERE workspace_id=$1 AND project_id=$2 AND status<>'archived'`
	args := []any{f.WorkspaceID, f.ProjectID}
	if f.Status != "" {
		query += ` AND status=$3`
		args = append(args, f.Status)
	}
	query += ` ORDER BY is_default DESC, created_at ASC`

	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()

	items := make([]EstimatePoint, 0)
	for rows.Next() {
		var ep EstimatePoint
		if err := rows.Scan(&ep.ID, &ep.Code, &ep.Name, &ep.WorkspaceID,
			&ep.Description, &ep.Points, &ep.IsDefault, &ep.Status,
			&ep.CreatedBy, &ep.CreatedAt, &ep.UpdatedAt); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
		items = append(items, ep)
	}
	return items, nil
}

// Update 更新估算点数。
func (s *EstimatePointService) Update(ctx context.Context, in UpdateInput) (*EstimatePoint, error) {
	ep, err := s.Get(ctx, in.WorkspaceID, in.ID)
	if err != nil {
		return nil, err
	}
	if in.Name != nil {
		name := strings.TrimSpace(*in.Name)
		if name == "" {
			return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "name", Reason: "名称不能为空"})
		}
		ep.Name = name
	}
	if in.Description != nil {
		ep.Description = *in.Description
	}
	if in.HasPoints {
		ep.Points = in.Points
	}
	if in.IsDefault != nil {
		ep.IsDefault = *in.IsDefault
	}

	_, err = s.db.Exec(ctx, `
		UPDATE estimate_points SET name=$1, description=$2, points_config=$3, is_default=$4, updated_at=NOW()
		WHERE id=$5 AND workspace_id=$6 AND status<>'archived'`,
		ep.Name, ep.Description, ep.Points, ep.IsDefault, ep.ID, in.WorkspaceID,
	)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	return ep, nil
}

// Delete 软删除估算点数。
func (s *EstimatePointService) Delete(ctx context.Context, wsID, id int64) error {
	ep, err := s.Get(ctx, wsID, id)
	if err != nil {
		return err
	}
	// 默认点集不可删除
	if ep.IsDefault {
		return errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "id", Reason: "默认估算点数不可删除"})
	}
	result, err := s.db.Exec(ctx, `
		UPDATE estimate_points SET status='archived', updated_at=NOW()
		WHERE id=$1 AND workspace_id=$2 AND status<>'archived'`,
		id, wsID,
	)
	if err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	if result.RowsAffected() == 0 {
		return errs.ErrNotFound.From()
	}
	return nil
}
