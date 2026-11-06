// Package workspaceview — 工作空间视图应用服务。
package workspaceview

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// Service 工作空间视图应用服务。
type Service struct {
	db *pgxpool.Pool
}

// NewService 创建服务。
func NewService(db *pgxpool.Pool) *Service {
	return &Service{db: db}
}

// --- 错误 ---

var (
	// ErrViewNotFound 视图不存在。
	ErrViewNotFound = errors.New("workspace view not found")
	// ErrInvalidLayout 布局类型非法。
	ErrInvalidLayout = errors.New("invalid layout")
)

const viewColumns = `id, workspace_id, name, COALESCE(description,''), layout,
	filters, display_options, sort, is_public, is_favorite, created_by, created_at, updated_at`

// Create 创建视图。
func (s *Service) Create(ctx context.Context, in CreateInput) (*WorkspaceView, error) {
	if strings.TrimSpace(in.Name) == "" {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "name", Reason: "名称不能为空"})
	}
	layout := in.Layout
	if layout == "" {
		layout = LayoutList
	}
	if !layout.IsValid() {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "layout", Reason: "无效的布局类型"})
	}

	var v WorkspaceView
	err := scanView(s.db.QueryRow(ctx, `
		INSERT INTO workspace_views
			(id, workspace_id, name, description, layout, filters, display_options, sort,
			 is_public, is_favorite, created_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		RETURNING `+viewColumns,
		genID(), in.WorkspaceID, in.Name, nullIfEmpty(in.Description), string(layout),
		rawJSONOrEmpty(in.Filters), rawJSONOrEmpty(in.DisplayOptions), rawJSONOrEmpty(in.Sort),
		in.IsPublic, in.IsFavorite, in.CreatedBy,
	), &v)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(fmt.Errorf("workspaceview.Create: %w", err))
	}
	return &v, nil
}

// Get 查询视图详情。
func (s *Service) Get(ctx context.Context, viewID int64) (*WorkspaceView, error) {
	var v WorkspaceView
	if err := scanView(s.db.QueryRow(ctx, `
		SELECT `+viewColumns+` FROM workspace_views WHERE id = $1`, viewID), &v); err != nil {
		return nil, err
	}
	return &v, nil
}

// List 列出工作空间下的视图。
func (s *Service) List(ctx context.Context, opts ListOptions) ([]WorkspaceView, error) {
	where := []string{"workspace_id = $1"}
	args := []any{opts.WorkspaceID}
	idx := 2
	if opts.CreatedBy > 0 {
		where = append(where, fmt.Sprintf("created_by = $%d", idx))
		args = append(args, opts.CreatedBy)
		idx++
	}
	if opts.PublicOnly {
		where = append(where, "is_public = TRUE")
	}
	if opts.FavoriteOnly {
		where = append(where, "is_favorite = TRUE")
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

	query := `SELECT ` + viewColumns + ` FROM workspace_views
		WHERE ` + strings.Join(where, " AND ") + `
		ORDER BY is_favorite DESC, created_at DESC
		LIMIT $` + fmt.Sprint(idx) + ` OFFSET $` + fmt.Sprint(idx+1)

	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(fmt.Errorf("workspaceview.List: %w", err))
	}
	defer rows.Close()

	out := make([]WorkspaceView, 0)
	for rows.Next() {
		var v WorkspaceView
		if err := scanView(rows, &v); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

// Update 更新视图（部分更新）。
func (s *Service) Update(ctx context.Context, viewID int64, in UpdateInput) (*WorkspaceView, error) {
	setCols := make([]string, 0, 8)
	args := make([]any, 0, 9)
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
	if in.Layout != nil {
		if !in.Layout.IsValid() {
			return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "layout", Reason: "无效的布局类型"})
		}
		setCols = append(setCols, fmt.Sprintf("layout = $%d", idx))
		args = append(args, string(*in.Layout))
		idx++
	}
	if in.Filters != nil {
		setCols = append(setCols, fmt.Sprintf("filters = $%d", idx))
		args = append(args, rawJSONOrEmpty(*in.Filters))
		idx++
	}
	if in.DisplayOptions != nil {
		setCols = append(setCols, fmt.Sprintf("display_options = $%d", idx))
		args = append(args, rawJSONOrEmpty(*in.DisplayOptions))
		idx++
	}
	if in.Sort != nil {
		setCols = append(setCols, fmt.Sprintf("sort = $%d", idx))
		args = append(args, rawJSONOrEmpty(*in.Sort))
		idx++
	}
	if in.IsPublic != nil {
		setCols = append(setCols, fmt.Sprintf("is_public = $%d", idx))
		args = append(args, *in.IsPublic)
		idx++
	}
	if in.IsFavorite != nil {
		setCols = append(setCols, fmt.Sprintf("is_favorite = $%d", idx))
		args = append(args, *in.IsFavorite)
		idx++
	}

	if len(setCols) == 0 {
		return s.Get(ctx, viewID)
	}

	setCols = append(setCols, "updated_at = now()")
	args = append(args, viewID)
	query := fmt.Sprintf(`UPDATE workspace_views SET %s WHERE id = $%d RETURNING %s`,
		strings.Join(setCols, ", "), idx, viewColumns)

	var v WorkspaceView
	if err := scanView(s.db.QueryRow(ctx, query, args...), &v); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrViewNotFound
		}
		return nil, errs.ErrInternal.Wrap(fmt.Errorf("workspaceview.Update: %w", err))
	}
	return &v, nil
}

// Delete 删除视图。
func (s *Service) Delete(ctx context.Context, viewID int64) error {
	tag, err := s.db.Exec(ctx, `DELETE FROM workspace_views WHERE id = $1`, viewID)
	if err != nil {
		return errs.ErrInternal.Wrap(fmt.Errorf("workspaceview.Delete: %w", err))
	}
	if tag.RowsAffected() == 0 {
		return ErrViewNotFound
	}
	return nil
}

// Duplicate 复制视图（生成一条新的独立副本，重置公开/收藏状态）。
func (s *Service) Duplicate(ctx context.Context, viewID, userID int64) (*WorkspaceView, error) {
	src, err := s.Get(ctx, viewID)
	if err != nil {
		return nil, err
	}
	return s.Create(ctx, CreateInput{
		WorkspaceID:    src.WorkspaceID,
		Name:           src.Name + " (副本)",
		Description:    src.Description,
		Layout:         src.Layout,
		Filters:        src.Filters,
		DisplayOptions: src.DisplayOptions,
		Sort:           src.Sort,
		IsPublic:       false,
		IsFavorite:     false,
		CreatedBy:      userID,
	})
}

// ToggleFavorite 切换收藏状态并返回最新值。
func (s *Service) ToggleFavorite(ctx context.Context, viewID int64) (bool, error) {
	var isFav bool
	err := s.db.QueryRow(ctx, `
		UPDATE workspace_views SET is_favorite = NOT is_favorite, updated_at = now()
		WHERE id = $1 RETURNING is_favorite`, viewID).Scan(&isFav)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, ErrViewNotFound
		}
		return false, errs.ErrInternal.Wrap(fmt.Errorf("workspaceview.ToggleFavorite: %w", err))
	}
	return isFav, nil
}

// TogglePublic 切换公开状态并返回最新值。
func (s *Service) TogglePublic(ctx context.Context, viewID int64) (bool, error) {
	var isPublic bool
	err := s.db.QueryRow(ctx, `
		UPDATE workspace_views SET is_public = NOT is_public, updated_at = now()
		WHERE id = $1 RETURNING is_public`, viewID).Scan(&isPublic)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, ErrViewNotFound
		}
		return false, errs.ErrInternal.Wrap(fmt.Errorf("workspaceview.TogglePublic: %w", err))
	}
	return isPublic, nil
}

// --- helpers ---

func nullIfEmpty(s string) any {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return s
}

// rawJSONOrEmpty 将 json.RawMessage 转为可落库的 JSONB 字节；
// 空 / null / 空对象 统一回退为 '{}'。
func rawJSONOrEmpty(raw json.RawMessage) []byte {
	trimmed := strings.TrimSpace(string(raw))
	if trimmed == "" || trimmed == "null" {
		return []byte("{}")
	}
	return []byte(trimmed)
}

func scanView(row pgx.Row, v *WorkspaceView) error {
	var desc string
	var filters, displayOpts, sort []byte
	err := row.Scan(
		&v.ID, &v.WorkspaceID, &v.Name, &desc, &v.Layout,
		&filters, &displayOpts, &sort, &v.IsPublic, &v.IsFavorite,
		&v.CreatedBy, &v.CreatedAt, &v.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrViewNotFound
		}
		return errs.ErrInternal.Wrap(err)
	}
	v.Description = desc
	v.Filters = normalizeRaw(filters)
	v.DisplayOptions = normalizeRaw(displayOpts)
	v.Sort = normalizeRaw(sort)
	return nil
}

// normalizeRaw 保证 JSONB 字段始终非 nil（避免 JSON 序列化输出 null）。
func normalizeRaw(b []byte) json.RawMessage {
	if len(b) == 0 {
		return json.RawMessage("{}")
	}
	return json.RawMessage(b)
}
