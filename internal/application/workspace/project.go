// Package workspace — 项目应用服务（CRUD + 归档）。
//
// 项目是工作空间下的二级聚合根，包含模块/标签/迭代等子域。
package workspace

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// Project 项目 DTO。
type Project struct {
	ID          int64     `json:"id"`
	WorkspaceID int64     `json:"workspace_id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Identifier  string    `json:"identifier"`
	Description string    `json:"description,omitempty"`
	Network     string    `json:"network"`
	Icon        string    `json:"icon,omitempty"`
	Color       string    `json:"color,omitempty"`
	Status      string    `json:"status"`
	SortOrder   float64   `json:"sort_order"`
	CreatedBy   int64     `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ProjectCreateInput 入参。
type ProjectCreateInput struct {
	WorkspaceID int64
	Name        string
	Slug        string
	Identifier  string
	Description string
	Network     string
	Icon        string
	Color       string
	CreatedBy   int64
}

// ProjectUpdateInput 入参。
type ProjectUpdateInput struct {
	Name        *string
	Slug        *string
	Description *string
	Network     *string
	Icon        *string
	Color       *string
}

// ProjectService 项目应用服务。
type ProjectService struct {
	db *pgxpool.Pool
}

// NewProjectService 创建项目服务。
func NewProjectService(db *pgxpool.Pool) *ProjectService {
	return &ProjectService{db: db}
}

// Create 创建项目。
func (s *ProjectService) Create(ctx context.Context, in ProjectCreateInput) (*Project, error) {
	if strings.TrimSpace(in.Name) == "" {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "name", Reason: "项目名称不能为空"})
	}
	slug := normalizeProjectSlug(in.Slug, in.Name)
	identifier := normalizeIdentifier(in.Identifier, slug)
	if in.Network != "private" {
		in.Network = "public"
	}

	var p Project
	err := s.db.QueryRow(ctx, `
		INSERT INTO projects (workspace_id, name, slug, identifier, description, network, icon, color, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, workspace_id, name, slug, identifier, description, network, icon, color, status, sort_order, created_by, created_at, updated_at`,
		in.WorkspaceID, in.Name, slug, identifier, in.Description, in.Network, in.Icon, in.Color, in.CreatedBy).
		Scan(&p.ID, &p.WorkspaceID, &p.Name, &p.Slug, &p.Identifier, &p.Description,
			&p.Network, &p.Icon, &p.Color, &p.Status, &p.SortOrder, &p.CreatedBy, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if strings.Contains(err.Error(), "projects_workspace_id_slug") ||
			strings.Contains(err.Error(), "projects_workspace_id_identifier") {
			return nil, errs.New("PROJECT.DUPLICATE", "项目链接标识或前缀已存在", 409)
		}
		return nil, errs.ErrInternal.Wrap(err)
	}
	return &p, nil
}

// Get 获取项目详情。
func (s *ProjectService) Get(ctx context.Context, wsID, projectID int64) (*Project, error) {
	var p Project
	err := s.db.QueryRow(ctx, `
		SELECT id, workspace_id, name, slug, identifier, description, network, icon, color, status, sort_order, created_by, created_at, updated_at
		FROM projects WHERE id = $1 AND workspace_id = $2 AND deleted_at IS NULL`,
		projectID, wsID).
		Scan(&p.ID, &p.WorkspaceID, &p.Name, &p.Slug, &p.Identifier, &p.Description,
			&p.Network, &p.Icon, &p.Color, &p.Status, &p.SortOrder, &p.CreatedBy, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrNotFound
		}
		return nil, errs.ErrInternal.Wrap(err)
	}
	return &p, nil
}

// ListByWorkspace 列出工作空间下的全部项目。
func (s *ProjectService) ListByWorkspace(ctx context.Context, wsID int64) ([]Project, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, workspace_id, name, slug, identifier, description, network, icon, color, status, sort_order, created_by, created_at, updated_at
		FROM projects WHERE workspace_id = $1 AND deleted_at IS NULL
		ORDER BY sort_order, created_at ASC`, wsID)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()

	var out []Project
	for rows.Next() {
		var p Project
		if err := rows.Scan(&p.ID, &p.WorkspaceID, &p.Name, &p.Slug, &p.Identifier, &p.Description,
			&p.Network, &p.Icon, &p.Color, &p.Status, &p.SortOrder, &p.CreatedBy, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
		out = append(out, p)
	}
	return out, nil
}

// Update 更新项目。
func (s *ProjectService) Update(ctx context.Context, wsID, projectID int64, in ProjectUpdateInput) (*Project, error) {
	var sets []string
	var args []any
	arg := 1

	if in.Name != nil {
		if strings.TrimSpace(*in.Name) == "" {
			return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "name", Reason: "项目名称不能为空"})
		}
		sets = append(sets, "name = $"+strconv.Itoa(arg))
		args = append(args, *in.Name)
		arg++
	}
	if in.Slug != nil {
		sets = append(sets, "slug = $"+strconv.Itoa(arg))
		args = append(args, normalizeProjectSlug(*in.Slug, *in.Name))
		arg++
	}
	if in.Description != nil {
		sets = append(sets, "description = $"+strconv.Itoa(arg))
		args = append(args, *in.Description)
		arg++
	}
	if in.Network != nil {
		if *in.Network != "private" {
			*in.Network = "public"
		}
		sets = append(sets, "network = $"+strconv.Itoa(arg))
		args = append(args, *in.Network)
		arg++
	}
	if in.Icon != nil {
		sets = append(sets, "icon = $"+strconv.Itoa(arg))
		args = append(args, *in.Icon)
		arg++
	}
	if in.Color != nil {
		sets = append(sets, "color = $"+strconv.Itoa(arg))
		args = append(args, *in.Color)
		arg++
	}

	if len(sets) == 0 {
		return s.Get(ctx, wsID, projectID)
	}
	sets = append(sets, "updated_at = now()")
	query := "UPDATE projects SET " + strings.Join(sets, ", ") +
		" WHERE id = $" + strconv.Itoa(arg) + " AND workspace_id = $" + strconv.Itoa(arg+1) + " AND deleted_at IS NULL"
	args = append(args, projectID, wsID)

	var p Project
	err := s.db.QueryRow(ctx, query, args...).
		Scan(&p.ID, &p.WorkspaceID, &p.Name, &p.Slug, &p.Identifier, &p.Description,
			&p.Network, &p.Icon, &p.Color, &p.Status, &p.SortOrder, &p.CreatedBy, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrNotFound
		}
		if strings.Contains(err.Error(), "projects_workspace_id") {
			return nil, errs.New("PROJECT.DUPLICATE", "项目链接标识或前缀已存在", 409)
		}
		return nil, errs.ErrInternal.Wrap(err)
	}
	return &p, nil
}

// Archive 归档项目（软删除）。
func (s *ProjectService) Archive(ctx context.Context, wsID, projectID int64) error {
	tag, err := s.db.Exec(ctx, `
		UPDATE projects SET status = 'archived', updated_at = now()
		WHERE id = $1 AND workspace_id = $2 AND status = 'active'`,
		projectID, wsID)
	if err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	if tag.RowsAffected() == 0 {
		return errs.ErrNotFound
	}
	return nil
}

// --- helpers ---

func normalizeProjectSlug(s, fallback string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		s = fallback
	}
	s = strings.ToLower(s)
	var b strings.Builder
	prevDash := false
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			prevDash = false
		} else if !prevDash {
			b.WriteByte('-')
			prevDash = true
		}
	}
	return strings.Trim(b.String(), "-")
}

// normalizeIdentifier 生成项目前缀（2-6 位大写）。
func normalizeIdentifier(s, fallback string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		// 从 fallback 生成
		words := strings.FieldsFunc(fallback, func(r rune) bool {
			return r == ' ' || r == '-' || r == '_'
		})
		if len(words) == 0 {
			return "P"
		}
		if len(words) == 1 {
			w := strings.ToUpper(words[0])
			if len(w) > 6 {
				w = w[:6]
			}
			return w
		}
		var b strings.Builder
		for _, w := range words {
			if len(b.String()) >= 6 {
				break
			}
			if len(w) > 0 {
				b.WriteByte(strings.ToUpper(w)[0])
			}
		}
		return b.String()
	}
	s = strings.ToUpper(s)
	if len(s) > 6 {
		s = s[:6]
	}
	return s
}

