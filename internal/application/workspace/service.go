// Package workspace — 工作空间应用服务（CRUD + 归档）。
//
// 设计参考: Plane / Linear / GitLab namespace 模型。
// 一个工作空间 = 一个租户，内含多个 Project。
// 创建工作空间时自动将创建者设为 owner 并写入 workspace_members。
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

// Workspace 是工作空间的域模型（API 响应 DTO）。
type Workspace struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	LogoURL     string    `json:"logo_url,omitempty"`
	Timezone    string    `json:"timezone"`
	Language    string    `json:"language"`
	Status      string    `json:"status"`
	OwnerID     int64     `json:"owner_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Role        string    `json:"role,omitempty"`
	MemberCount int64     `json:"member_count,omitempty"`
	BrandColor  string    `json:"brand_color,omitempty"`
}

// CreateInput 创建工作空间的入参。
type CreateInput struct {
	Name     string
	Slug     string
	Timezone string
	Language string
	OwnerID  int64
}

// UpdateInput 更新工作空间的入参（仅 owner/admin 可用）。
type UpdateInput struct {
	Name       *string
	Timezone   *string
	Language   *string
	LogoURL    *string
	BrandColor *string
}

// WorkspaceConfig 工作空间配置（存储在 config JSONB 中）。
type WorkspaceConfig struct {
	BrandColor string `json:"brand_color,omitempty"`
}

// Service 提供工作空间应用服务。
type Service struct {
	db *pgxpool.Pool
}

// NewService 创建工作空间服务。
func NewService(db *pgxpool.Pool) *Service {
	return &Service{db: db}
}

// Create 创建新工作空间，自动将创建者设为 owner。
func (s *Service) Create(ctx context.Context, in CreateInput) (*Workspace, error) {
	if strings.TrimSpace(in.Name) == "" {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{
			Field: "name", Reason: "工作空间名称不能为空",
		})
	}
	slug := normalizeSlug(in.Slug, in.Name)
	if slug == "" {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{
			Field: "slug", Reason: "无效的链接标识（slug）",
		})
	}
	tz := in.Timezone
	if tz == "" {
		tz = "Asia/Shanghai"
	}

	var ws *Workspace
	err := pgx.BeginTxFunc(ctx, s.db, pgx.TxOptions{}, func(tx pgx.Tx) error {
		var w Workspace
		err := tx.QueryRow(ctx, `
			INSERT INTO workspaces (name, slug, timezone, language, owner_id)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING id, name, slug, coalesce(logo_url,''), timezone, language, status, owner_id, created_at, updated_at`,
			in.Name, slug, tz, in.Language, in.OwnerID).
			Scan(&w.ID, &w.Name, &w.Slug, &w.LogoURL, &w.Timezone, &w.Language, &w.Status, &w.OwnerID, &w.CreatedAt, &w.UpdatedAt)
		if err != nil {
			if strings.Contains(err.Error(), "uq_workspaces_slug") {
				return errs.New("WORKSPACE.SLUG_TAKEN", "该链接标识已被使用", 409)
			}
			return errs.ErrInternal.Wrap(err)
		}
		_, err = tx.Exec(ctx, `
			INSERT INTO workspace_members (workspace_id, user_id, role, joined_at)
			VALUES ($1, $2, 'owner', now())`,
			w.ID, in.OwnerID)
		if err != nil {
			return errs.ErrInternal.Wrap(err)
		}
		w.Role = "owner"
		w.MemberCount = 1
		ws = &w
		return nil
	})
	if err != nil {
		return nil, err
	}
	return ws, nil
}

// Get 获取单个工作空间详情（带 member_count）。
func (s *Service) Get(ctx context.Context, wsID int64) (*Workspace, error) {
	var w Workspace
	err := s.db.QueryRow(ctx, `
		SELECT id, name, slug, coalesce(logo_url,''), timezone, language, status, owner_id,
		       (SELECT count(*) FROM workspace_members WHERE workspace_id = $1),
		       coalesce(config->>'brand_color', ''),
		       created_at, updated_at
		FROM workspaces WHERE id = $1 AND status = 'active'`, wsID).
		Scan(&w.ID, &w.Name, &w.Slug, &w.LogoURL, &w.Timezone, &w.Language, &w.Status,
			&w.OwnerID, &w.MemberCount, &w.BrandColor, &w.CreatedAt, &w.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrNotFound
		}
		return nil, errs.ErrInternal.Wrap(err)
	}
	return &w, nil
}

// GetBySlug 根据 slug 获取工作空间详情。
func (s *Service) GetBySlug(ctx context.Context, slug string) (*Workspace, error) {
	var w Workspace
	err := s.db.QueryRow(ctx, `
		SELECT id, name, slug, coalesce(logo_url,''), timezone, language, status, owner_id,
		       (SELECT count(*) FROM workspace_members wm WHERE wm.workspace_id = w.id),
		       coalesce(config->>'brand_color', ''),
		       created_at, updated_at
		FROM workspaces w WHERE slug = $1 AND status = 'active'`, slug).
		Scan(&w.ID, &w.Name, &w.Slug, &w.LogoURL, &w.Timezone, &w.Language, &w.Status,
			&w.OwnerID, &w.MemberCount, &w.BrandColor, &w.CreatedAt, &w.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrNotFound
		}
		return nil, errs.ErrInternal.Wrap(err)
	}
	return &w, nil
}

// ListByUser 列出用户参与的所有工作空间。
func (s *Service) ListByUser(ctx context.Context, userID int64) ([]Workspace, error) {
	rows, err := s.db.Query(ctx, `
		SELECT w.id, w.name, w.slug, coalesce(w.logo_url,''), w.timezone, w.language, w.status, w.owner_id,
		       wm.role,
		       (SELECT count(*) FROM workspace_members WHERE workspace_id = w.id),
		       coalesce(w.config->>'brand_color', ''),
		       w.created_at, w.updated_at
		FROM workspace_members wm
		JOIN workspaces w ON w.id = wm.workspace_id
		WHERE wm.user_id = $1 AND w.status = 'active'
		ORDER BY w.created_at DESC`, userID)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()

	var out = make([]Workspace, 0)
	for rows.Next() {
		var w Workspace
		if err := rows.Scan(&w.ID, &w.Name, &w.Slug, &w.LogoURL, &w.Timezone, &w.Language,
			&w.Status, &w.OwnerID, &w.Role, &w.MemberCount, &w.BrandColor, &w.CreatedAt, &w.UpdatedAt); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
		out = append(out, w)
	}
	return out, nil
}

// Update 更新工作空间信息。
func (s *Service) Update(ctx context.Context, wsID int64, in UpdateInput) (*Workspace, error) {
	var sets []string
	var args []any
	arg := 1

	if in.Name != nil {
		trimmed := strings.TrimSpace(*in.Name)
		if trimmed == "" {
			return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "name", Reason: "名称不能为空"})
		}
		sets = append(sets, "name = $"+strconv.Itoa(arg))
		args = append(args, trimmed)
		arg++
	}
	if in.Timezone != nil {
		sets = append(sets, "timezone = $"+strconv.Itoa(arg))
		args = append(args, *in.Timezone)
		arg++
	}
	if in.Language != nil {
		sets = append(sets, "language = $"+strconv.Itoa(arg))
		args = append(args, *in.Language)
		arg++
	}
	if in.LogoURL != nil {
		sets = append(sets, "logo_url = $"+strconv.Itoa(arg))
		args = append(args, *in.LogoURL)
		arg++
	}
	if in.BrandColor != nil {
		// 更新 config JSONB 中的 brand_color
		sets = append(sets, "config = jsonb_set(coalesce(config, '{}'::jsonb), '{brand_color}', to_jsonb($"+strconv.Itoa(arg)+"::text))")
		args = append(args, *in.BrandColor)
		arg++
	}
	if len(sets) == 0 {
		return s.Get(ctx, wsID)
	}
	sets = append(sets, "updated_at = now()")
	query := "UPDATE workspaces SET " + strings.Join(sets, ", ") +
		" WHERE id = $" + strconv.Itoa(arg) + " AND status = 'active'" +
		" RETURNING id, name, slug, coalesce(logo_url,''), timezone, language, status, owner_id, coalesce(config->>'brand_color', ''), created_at, updated_at"
	args = append(args, wsID)

	var w Workspace
	err := s.db.QueryRow(ctx, query, args...).
		Scan(&w.ID, &w.Name, &w.Slug, &w.LogoURL, &w.Timezone, &w.Language, &w.Status, &w.OwnerID, &w.BrandColor, &w.CreatedAt, &w.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrNotFound
		}
		return nil, errs.ErrInternal.Wrap(err)
	}
	return &w, nil
}

// GetBrandColor 获取工作空间品牌色（从 config JSONB 读取）。
func (s *Service) GetBrandColor(ctx context.Context, wsID int64) (string, error) {
	var brandColor string
	err := s.db.QueryRow(ctx, `
		SELECT coalesce(config->>'brand_color', '')
		FROM workspaces WHERE id = $1 AND status = 'active'`, wsID).Scan(&brandColor)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", errs.ErrNotFound
		}
		return "", errs.ErrInternal.Wrap(err)
	}
	return brandColor, nil
}

// Archive 归档工作空间（仅 owner 可用）。
func (s *Service) Archive(ctx context.Context, wsID int64) error {
	tag, err := s.db.Exec(ctx, `
		UPDATE workspaces SET status = 'archived', updated_at = now()
		WHERE id = $1 AND status = 'active'`, wsID)
	if err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	if tag.RowsAffected() == 0 {
		return errs.ErrNotFound
	}
	return nil
}

// normalizeSlug 规范化 slug：小写、替换非字母数字/汉字为 -。
func normalizeSlug(s, fallback string) string {
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
		} else if r >= 0x4E00 && r <= 0x9FFF {
			b.WriteRune(r)
			prevDash = false
		} else if !prevDash {
			b.WriteByte('-')
			prevDash = true
		}
	}
	result := strings.Trim(b.String(), "-")
	if len(result) > 60 {
		result = result[:60]
	}
	return result
}
