// Package workspace — 项目应用服务（CRUD + 归档）。
//
// 项目是工作空间下的二级聚合根，包含模块/标签/迭代等子域。
package workspace

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// ProjectModuleToggles — 项目功能模块开关集合。
// 实现 driver.Valuer / sql.Scanner 以透明持久化为 JSONB。
type ProjectModuleToggles struct {
	Sprint   bool `json:"sprint"`
	Version  bool `json:"version"`
	Estimate bool `json:"estimate"`
}

// Value 实现 driver.Valuer，将结构体序列化为 JSON 字节。
func (m ProjectModuleToggles) Value() (driver.Value, error) {
	return json.Marshal(m)
}

// Scan 实现 sql.Scanner，从 JSON 字节反序列化。
func (m *ProjectModuleToggles) Scan(src any) error {
	var b []byte
	switch v := src.(type) {
	case []byte:
		b = v
	case string:
		b = []byte(v)
	case nil:
		return nil
	default:
		return errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "modules", Reason: "unsupported scan type"})
	}
	if len(b) == 0 {
		return nil
	}
	return json.Unmarshal(b, m)
}

// ProjectModuleAllEnabled 返回全部模块启用的默认开关集合。
func ProjectModuleAllEnabled() ProjectModuleToggles {
	return ProjectModuleToggles{Sprint: true, Version: true, Estimate: true}
}

// Project 项目 DTO。
type Project struct {
	ID             int64                  `json:"id"`
	WorkspaceID    int64                  `json:"workspace_id"`
	Name           string                 `json:"name"`
	Slug           string                 `json:"slug"`
	Identifier     string                 `json:"identifier"`
	Description    *string                `json:"description,omitempty"`
	Network        string                 `json:"network"`
	Icon           *string                `json:"icon,omitempty"`
	Color          *string                `json:"color,omitempty"`
	CoverImageUrl  *string                `json:"cover_image_url,omitempty"`
	Template       string                 `json:"template"`
	Status         string                 `json:"status"`
	SortOrder      float64                `json:"sort_order"`
	Modules        ProjectModuleToggles   `json:"modules"`
	CreatedBy      int64                  `json:"created_by"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
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
	// Template 项目模板代码（agile / waterfall / generic），默认 generic。
	Template string
	// Modules 功能模块开关集合；nil 表示全部启用。
	Modules *ProjectModuleToggles
	// CoverImageUrl 封面图片 URL；可选。
	CoverImageUrl *string
}

// ProjectUpdateInput 入参。
type ProjectUpdateInput struct {
	Name          *string
	Slug          *string
	Description   *string
	Network       *string
	Icon          *string
	Color         *string
	Modules       *ProjectModuleToggles
	CoverImageUrl *string
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
	if in.Network != "private" && in.Network != "internal" {
		in.Network = "public"
	}
	if in.Template != "agile" && in.Template != "waterfall" {
		in.Template = "generic"
	}

	sModules := ProjectModuleAllEnabled()
	if in.Modules != nil {
		sModules = *in.Modules
	}

	var p Project
	err := pgx.BeginTxFunc(ctx, s.db, pgx.TxOptions{}, func(tx pgx.Tx) error {
		if err := tx.QueryRow(ctx, `
			INSERT INTO projects (workspace_id, name, slug, identifier, description, network, icon, color, cover_image_url, template, modules, created_by)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
			RETURNING id, workspace_id, name, slug, identifier, description, network, icon, color, cover_image_url, template, status, sort_order, modules, created_by, created_at, updated_at`,
			in.WorkspaceID, in.Name, slug, identifier, in.Description, in.Network, in.Icon, in.Color, in.CoverImageUrl, in.Template, sModules, in.CreatedBy).
			Scan(&p.ID, &p.WorkspaceID, &p.Name, &p.Slug, &p.Identifier, &p.Description,
				&p.Network, &p.Icon, &p.Color, &p.CoverImageUrl, &p.Template, &p.Status, &p.SortOrder, &p.Modules, &p.CreatedBy, &p.CreatedAt, &p.UpdatedAt); err != nil {
			if strings.Contains(err.Error(), "idx_projects_workspace_slug") ||
				strings.Contains(err.Error(), "idx_projects_workspace_identifier") {
				return errs.New("PROJECT.DUPLICATE", "项目链接标识或前缀已存在", 409)
			}
			return errs.ErrInternal.Wrap(err)
		}
		// 创建者自动加入为项目 admin
		if _, err := tx.Exec(ctx, `
			INSERT INTO project_members (workspace_id, project_id, user_id, role, created_by)
			VALUES ($1, $2, $3, 'admin', $3)
			ON CONFLICT (workspace_id, project_id, user_id) DO UPDATE SET role = 'admin'`,
			in.WorkspaceID, p.ID, in.CreatedBy); err != nil {
			return errs.ErrInternal.Wrap(err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	// 创建默认风险规则（失败不阻塞项目创建）
	go EnsureProjectDefaultRiskRules(context.Background(), s.db, in.WorkspaceID, p.ID)
	return &p, nil
}

// Get 获取项目详情。
func (s *ProjectService) Get(ctx context.Context, wsID, projectID int64) (*Project, error) {
	var p Project
	err := s.db.QueryRow(ctx, `
		SELECT id, workspace_id, name, slug, identifier, description, network, icon, color, cover_image_url, template, status, sort_order, modules, created_by, created_at, updated_at
		FROM projects WHERE id = $1 AND workspace_id = $2 AND deleted = false`,
		projectID, wsID).
		Scan(&p.ID, &p.WorkspaceID, &p.Name, &p.Slug, &p.Identifier, &p.Description,
			&p.Network, &p.Icon, &p.Color, &p.CoverImageUrl, &p.Template, &p.Status, &p.SortOrder, &p.Modules, &p.CreatedBy, &p.CreatedAt, &p.UpdatedAt)
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
		SELECT id, workspace_id, name, slug, identifier, description, network, icon, color, cover_image_url, template, status, sort_order, modules, created_by, created_at, updated_at
		FROM projects WHERE workspace_id = $1 AND deleted = false
		ORDER BY sort_order, created_at ASC`, wsID)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()

	var out = make([]Project, 0)
	for rows.Next() {
		var p Project
		if err := rows.Scan(&p.ID, &p.WorkspaceID, &p.Name, &p.Slug, &p.Identifier, &p.Description,
			&p.Network, &p.Icon, &p.Color, &p.CoverImageUrl, &p.Template, &p.Status, &p.SortOrder, &p.Modules, &p.CreatedBy, &p.CreatedAt, &p.UpdatedAt); err != nil {
			fmt.Printf("DEBUG_PROJECT_SCAN_ERROR: %v\n", err)
			return nil, errs.ErrInternal.Wrap(fmt.Errorf("scan project row: %w", err))
		}
		out = append(out, p)
	}
	if err := rows.Err(); err != nil {
		fmt.Printf("DEBUG_PROJECT_ROWS_ERROR: %v\n", err)
		return nil, errs.ErrInternal.Wrap(fmt.Errorf("rows iteration: %w", err))
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
		if *in.Network != "private" && *in.Network != "internal" {
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
	if in.Modules != nil {
		sets = append(sets, "modules = $"+strconv.Itoa(arg))
		args = append(args, *in.Modules)
		arg++
	}
	if in.CoverImageUrl != nil {
		sets = append(sets, "cover_image_url = $"+strconv.Itoa(arg))
		args = append(args, *in.CoverImageUrl)
		arg++
	}

	if len(sets) == 0 {
		return s.Get(ctx, wsID, projectID)
	}
	sets = append(sets, "updated_at = now()")
	query := "UPDATE projects SET " + strings.Join(sets, ", ") +
		" WHERE id = $" + strconv.Itoa(arg) + " AND workspace_id = $" + strconv.Itoa(arg+1) + " AND deleted = false " +
		"RETURNING id, workspace_id, name, slug, identifier, description, network, icon, color, cover_image_url, template, status, sort_order, modules, created_by, created_at, updated_at"
	args = append(args, projectID, wsID)

	var p Project
	err := s.db.QueryRow(ctx, query, args...).
		Scan(&p.ID, &p.WorkspaceID, &p.Name, &p.Slug, &p.Identifier, &p.Description,
			&p.Network, &p.Icon, &p.Color, &p.CoverImageUrl, &p.Template, &p.Status, &p.SortOrder, &p.Modules, &p.CreatedBy, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrNotFound
		}
		if strings.Contains(err.Error(), "idx_projects_workspace") {
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

