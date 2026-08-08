// Package preference — 视图偏好持久化及命名视图管理。
package preference

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// SavedViewScope 视图范围枚举。
type SavedViewScope string

const (
	ScopePersonal SavedViewScope = "personal"
	ScopeTeam     SavedViewScope = "team"
	ScopeDefault  SavedViewScope = "default"
)

// SavedView 命名视图模型。
type SavedView struct {
	ID          int64           `json:"id"`
	WorkspaceID int64           `json:"workspace_id"`
	ProjectID   int64           `json:"project_id"`
	Name        string          `json:"name"`
	Type        ViewType        `json:"type"`
	Scope       SavedViewScope  `json:"scope"`
	Config      json.RawMessage `json:"config"`
	OwnerID     int64           `json:"owner_id"`
	IsShared    bool            `json:"is_shared"`
	CreatedAt   string          `json:"created_at"`
	UpdatedAt   string          `json:"updated_at"`
}

// CreateViewInput 创建视图入参。
type CreateViewInput struct {
	Name     string          `json:"name" binding:"required,max=128"`
	Type     ViewType        `json:"type" binding:"required"`
	Scope    SavedViewScope  `json:"scope"`
	Config   json.RawMessage `json:"config"`
	IsShared bool            `json:"is_shared"`
}

// UpdateViewInput 更新视图入参（部分更新）。
type UpdateViewInput struct {
	Name     *string          `json:"name"`
	Type     *ViewType        `json:"type"`
	Scope    *SavedViewScope  `json:"scope"`
	Config   *json.RawMessage `json:"config"`
	IsShared *bool            `json:"is_shared"`
}

// ViewService 命名视图管理服务。
type ViewService struct {
	db *pgxpool.Pool
}

// NewViewService 创建视图管理服务。
func NewViewService(db *pgxpool.Pool) *ViewService {
	return &ViewService{db: db}
}

// EnsureSchema 确保 saved_views 表存在。
func (s *ViewService) EnsureSchema(ctx context.Context) error {
	_, err := s.db.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS saved_views (
			id BIGSERIAL PRIMARY KEY,
			workspace_id BIGINT NOT NULL,
			project_id BIGINT NOT NULL,
			name VARCHAR(128) NOT NULL,
			type VARCHAR(32) NOT NULL DEFAULT 'list',
			scope VARCHAR(32) NOT NULL DEFAULT 'personal',
			config JSONB NOT NULL DEFAULT '{}',
			owner_id BIGINT NOT NULL,
			is_shared BOOLEAN NOT NULL DEFAULT false,
			created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
		)`)
	return err
}

// Create 创建新视图。
func (s *ViewService) Create(ctx context.Context, wsID, projectID, userID int64, input *CreateViewInput) (*SavedView, error) {
	if input.Name == "" {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "name", Reason: "视图名称不能为空"})
	}
	if input.Type == "" {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "type", Reason: "必须指定视图类型"})
	}
	if input.Scope == "" {
		input.Scope = ScopePersonal
	}
	config := defaultJSON(input.Config, "{}")

	var v SavedView
	err := s.db.QueryRow(ctx, `
		INSERT INTO saved_views (workspace_id, project_id, name, type, scope, config, owner_id, is_shared)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		RETURNING id, workspace_id, project_id, name, type, scope, config, owner_id, is_shared, created_at::text, updated_at::text`,
		wsID, projectID, input.Name, string(input.Type), string(input.Scope),
		config, userID, input.IsShared).
		Scan(&v.ID, &v.WorkspaceID, &v.ProjectID, &v.Name, &v.Type, &v.Scope,
			&v.Config, &v.OwnerID, &v.IsShared, &v.CreatedAt, &v.UpdatedAt)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	return &v, nil
}

// Update 更新视图（仅 owner 可编辑）。
func (s *ViewService) Update(ctx context.Context, viewID, userID int64, input *UpdateViewInput) (*SavedView, error) {
	// 先查所有者
	ownerID, err := s.getOwnerID(ctx, viewID)
	if err != nil {
		return nil, err
	}
	if ownerID != userID {
		return nil, errs.ErrForbidden.WithDetails(errs.FieldDetail{Field: "view_id", Reason: "只有视图创建者可编辑"})
	}

	// 动态构建 UPDATE
	setClauses := []string{}
	args := []interface{}{viewID}
	argIdx := 2

	if input.Name != nil {
		setClauses = append(setClauses, "name = $"+itoa(argIdx))
		args = append(args, *input.Name)
		argIdx++
	}
	if input.Type != nil {
		setClauses = append(setClauses, "type = $"+itoa(argIdx))
		args = append(args, string(*input.Type))
		argIdx++
	}
	if input.Scope != nil {
		setClauses = append(setClauses, "scope = $"+itoa(argIdx))
		args = append(args, string(*input.Scope))
		argIdx++
	}
	if input.Config != nil {
		setClauses = append(setClauses, "config = $"+itoa(argIdx))
		args = append(args, *input.Config)
		argIdx++
	}
	if input.IsShared != nil {
		setClauses = append(setClauses, "is_shared = $"+itoa(argIdx))
		args = append(args, *input.IsShared)
		argIdx++
	}

	if len(setClauses) == 0 {
		// 无变更，直接返回现有视图
		return s.Get(ctx, viewID)
	}

	setClauses = append(setClauses, "updated_at = now()")
	query := `UPDATE saved_views SET ` + joinStrings(setClauses, ", ") + ` WHERE id = $1
		RETURNING id, workspace_id, project_id, name, type, scope, config, owner_id, is_shared, created_at::text, updated_at::text`

	var v SavedView
	err = s.db.QueryRow(ctx, query, args...).
		Scan(&v.ID, &v.WorkspaceID, &v.ProjectID, &v.Name, &v.Type, &v.Scope,
			&v.Config, &v.OwnerID, &v.IsShared, &v.CreatedAt, &v.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrNotFound.WithDetails(errs.FieldDetail{Field: "view_id", Reason: "视图不存在"})
		}
		return nil, errs.ErrInternal.Wrap(err)
	}
	return &v, nil
}

// Delete 删除视图（仅 owner 可删除）。
func (s *ViewService) Delete(ctx context.Context, viewID, userID int64) error {
	ownerID, err := s.getOwnerID(ctx, viewID)
	if err != nil {
		return err
	}
	if ownerID != userID {
		return errs.ErrForbidden.WithDetails(errs.FieldDetail{Field: "view_id", Reason: "只有视图创建者可删除"})
	}

	tag, err := s.db.Exec(ctx, `DELETE FROM saved_views WHERE id = $1`, viewID)
	if err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	if tag.RowsAffected() == 0 {
		return errs.ErrNotFound.WithDetails(errs.FieldDetail{Field: "view_id", Reason: "视图不存在"})
	}
	return nil
}

// Get 获取单个视图详情。
func (s *ViewService) Get(ctx context.Context, viewID int64) (*SavedView, error) {
	var v SavedView
	err := s.db.QueryRow(ctx, `
		SELECT id, workspace_id, project_id, name, type, scope, config, owner_id, is_shared, created_at::text, updated_at::text
		FROM saved_views WHERE id = $1`, viewID).
		Scan(&v.ID, &v.WorkspaceID, &v.ProjectID, &v.Name, &v.Type, &v.Scope,
			&v.Config, &v.OwnerID, &v.IsShared, &v.CreatedAt, &v.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrNotFound.WithDetails(errs.FieldDetail{Field: "view_id", Reason: "视图不存在"})
		}
		return nil, errs.ErrInternal.Wrap(err)
	}
	return &v, nil
}

// List 列出项目下指定范围的视图。
// personal → 仅当前用户创建的；team → 项目内 is_shared 的；default → 管理员设定的团队默认视图。
func (s *ViewService) List(ctx context.Context, wsID, projectID, userID int64, scope SavedViewScope) ([]SavedView, error) {
	var rows pgx.Rows
	var err error

	switch scope {
	case ScopePersonal:
		rows, err = s.db.Query(ctx, `
			SELECT id, workspace_id, project_id, name, type, scope, config, owner_id, is_shared, created_at::text, updated_at::text
			FROM saved_views
			WHERE workspace_id = $1 AND project_id = $2 AND owner_id = $3
			ORDER BY updated_at DESC`,
			wsID, projectID, userID)
	case ScopeDefault:
		rows, err = s.db.Query(ctx, `
			SELECT id, workspace_id, project_id, name, type, scope, config, owner_id, is_shared, created_at::text, updated_at::text
			FROM saved_views
			WHERE workspace_id = $1 AND project_id = $2 AND scope = 'default'
			ORDER BY updated_at DESC`,
			wsID, projectID)
	case ScopeTeam:
		fallthrough
	default:
		// 返回团队共享视图 + 个人视图（所有可用视图）
		rows, err = s.db.Query(ctx, `
			SELECT id, workspace_id, project_id, name, type, scope, config, owner_id, is_shared, created_at::text, updated_at::text
			FROM saved_views
			WHERE workspace_id = $1 AND project_id = $2 AND (
				(owner_id = $3 AND scope = 'personal')
				OR is_shared = true
				OR scope = 'team'
				OR scope = 'default'
			)
			ORDER BY updated_at DESC`,
			wsID, projectID, userID)
	}
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()

	var list []SavedView
	for rows.Next() {
		var v SavedView
		if err := rows.Scan(&v.ID, &v.WorkspaceID, &v.ProjectID, &v.Name, &v.Type, &v.Scope,
			&v.Config, &v.OwnerID, &v.IsShared, &v.CreatedAt, &v.UpdatedAt); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
		list = append(list, v)
	}
	return list, rows.Err()
}

// SetDefault 设置默认视图（管理员操作）。
func (s *ViewService) SetDefault(ctx context.Context, wsID, projectID, viewID int64) error {
	// 取消该项目的现有默认视图
	_, err := s.db.Exec(ctx, `
		UPDATE saved_views SET scope = 'team', updated_at = now()
		WHERE workspace_id = $1 AND project_id = $2 AND scope = 'default'`,
		wsID, projectID)
	if err != nil {
		return errs.ErrInternal.Wrap(err)
	}

	// 设置新默认视图
	tag, err := s.db.Exec(ctx, `
		UPDATE saved_views SET scope = 'default', updated_at = now()
		WHERE id = $1 AND workspace_id = $2 AND project_id = $3`,
		viewID, wsID, projectID)
	if err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	if tag.RowsAffected() == 0 {
		return errs.ErrNotFound.WithDetails(errs.FieldDetail{Field: "view_id", Reason: "视图不存在"})
	}
	return nil
}

// getOwnerID 获取视图的所有者 ID。
func (s *ViewService) getOwnerID(ctx context.Context, viewID int64) (int64, error) {
	var ownerID int64
	err := s.db.QueryRow(ctx, `SELECT owner_id FROM saved_views WHERE id = $1`, viewID).Scan(&ownerID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, errs.ErrNotFound.WithDetails(errs.FieldDetail{Field: "view_id", Reason: "视图不存在"})
		}
		return 0, errs.ErrInternal.Wrap(err)
	}
	return ownerID, nil
}

// --- helpers ---

func itoa(n int) string {
	res := make([]byte, 0, 4)
	if n == 0 {
		return "0"
	}
	for n > 0 {
		res = append([]byte{byte('0' + n%10)}, res...)
		n /= 10
	}
	return string(res)
}

func joinStrings(parts []string, sep string) string {
	if len(parts) == 0 {
		return ""
	}
	r := parts[0]
	for i := 1; i < len(parts); i++ {
		r += sep + parts[i]
	}
	return r
}
