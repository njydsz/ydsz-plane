// Package preference — 视图偏好持久化（看板/列表布局、列配置、过滤条件）。
package preference

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// ViewType 视图类型枚举。
type ViewType string

const (
	ViewKanban   ViewType = "kanban"
	ViewList     ViewType = "list"
	ViewCalendar ViewType = "calendar"
	ViewGantt    ViewType = "gantt"
)

// ViewPreference 视图偏好模型。
type ViewPreference struct {
	ID          int64           `json:"id"`
	WorkspaceID int64           `json:"workspace_id"`
	ProjectID   int64           `json:"project_id"`
	UserID      int64           `json:"user_id"`
	ViewType    ViewType        `json:"view_type"`
	Layout      string          `json:"layout"`
	Columns     json.RawMessage `json:"columns"`
	Filters     json.RawMessage `json:"filters"`
	Sort        json.RawMessage `json:"sort"`
	Extra       json.RawMessage `json:"extra"`
	CreatedAt   string          `json:"created_at"`
	UpdatedAt   string          `json:"updated_at"`
}

// Service 视图偏好服务。
type Service struct {
	db *pgxpool.Pool
}

// NewService 创建视图偏好服务。
func NewService(db *pgxpool.Pool) *Service {
	return &Service{db: db}
}

// Save 保存视图偏好（upsert）。
func (s *Service) Save(ctx context.Context, wsID, projectID, userID int64, vp *ViewPreference) (*ViewPreference, error) {
	if vp.ViewType == "" {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "view_type", Reason: "必须指定视图类型"})
	}
	if vp.Layout == "" {
		vp.Layout = "list"
	}
	cols := defaultJSON(vp.Columns, "[]")
	filters := defaultJSON(vp.Filters, "{}")
	sort := defaultJSON(vp.Sort, "{}")
	extra := defaultJSON(vp.Extra, "{}")

	err := s.db.QueryRow(ctx, `
		INSERT INTO view_preferences (workspace_id, project_id, user_id, view_type, layout, columns, filters, sort, extra)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		ON CONFLICT (workspace_id, project_id, user_id, view_type) DO UPDATE SET
			layout = EXCLUDED.layout,
			columns = EXCLUDED.columns,
			filters = EXCLUDED.filters,
			sort = EXCLUDED.sort,
			extra = EXCLUDED.extra,
			updated_at = now()
		RETURNING id, created_at::text, updated_at::text`,
		wsID, projectID, userID, string(vp.ViewType), vp.Layout,
		cols, filters, sort, extra).
		Scan(&vp.ID, &vp.CreatedAt, &vp.UpdatedAt)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	vp.WorkspaceID = wsID
	vp.ProjectID = projectID
	vp.UserID = userID
	vp.Columns = cols
	vp.Filters = filters
	vp.Sort = sort
	vp.Extra = extra
	return vp, nil
}

// Get 获取指定视图偏好（不存在返回 nil）。
func (s *Service) Get(ctx context.Context, wsID, projectID, userID int64, viewType ViewType) (*ViewPreference, error) {
	var vp ViewPreference
	var cols, filters, sort, extra []byte
	err := s.db.QueryRow(ctx, `
		SELECT id, workspace_id, project_id, user_id, view_type, layout,
		       columns, filters, sort, extra, created_at::text, updated_at::text
		FROM view_preferences
		WHERE workspace_id = $1 AND project_id = $2 AND user_id = $3 AND view_type = $4`,
		wsID, projectID, userID, string(viewType)).
		Scan(&vp.ID, &vp.WorkspaceID, &vp.ProjectID, &vp.UserID, &vp.ViewType, &vp.Layout,
			&cols, &filters, &sort, &extra, &vp.CreatedAt, &vp.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, errs.ErrInternal.Wrap(err)
	}
	vp.Columns = cols
	vp.Filters = filters
	vp.Sort = sort
	vp.Extra = extra
	return &vp, nil
}

// ListByUser 列出用户在某项目下的全部视图偏好。
func (s *Service) ListByUser(ctx context.Context, wsID, projectID, userID int64) ([]ViewPreference, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, workspace_id, project_id, user_id, view_type, layout,
		       columns, filters, sort, extra, created_at::text, updated_at::text
		FROM view_preferences
		WHERE workspace_id = $1 AND project_id = $2 AND user_id = $3
		ORDER BY view_type`, wsID, projectID, userID)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()

	var list []ViewPreference
	for rows.Next() {
		var vp ViewPreference
		var cols, filters, sort, extra []byte
		if err := rows.Scan(&vp.ID, &vp.WorkspaceID, &vp.ProjectID, &vp.UserID, &vp.ViewType, &vp.Layout,
			&cols, &filters, &sort, &extra, &vp.CreatedAt, &vp.UpdatedAt); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
		vp.Columns = cols
		vp.Filters = filters
		vp.Sort = sort
		vp.Extra = extra
		list = append(list, vp)
	}
	return list, rows.Err()
}

func defaultJSON(raw json.RawMessage, fallback string) json.RawMessage {
	if len(raw) == 0 {
		return json.RawMessage(fallback)
	}
	return raw
}
