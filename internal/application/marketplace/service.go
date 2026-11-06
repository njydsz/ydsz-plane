// Package marketplace — 应用商店服务。
package marketplace

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// Service 应用商店服务。
type Service struct {
	db *pgxpool.Pool
}

// NewService 构造服务。
func NewService(db *pgxpool.Pool) *Service { return &Service{db: db} }

// --- 错误 ---

var (
	// ErrAppNotFound 应用不存在。
	ErrAppNotFound = errors.New("marketplace app not found")
	// ErrAppSlugTaken Slug 已被占用。
	ErrAppSlugTaken = errors.New("marketplace app slug already taken")
	// ErrAlreadyInstalled 已安装。
	ErrAlreadyInstalled = errors.New("app already installed in this workspace")
	// ErrInvalidRating 评分非法。
	ErrInvalidRating = errors.New("rating must be between 1 and 5")
)

// --- 分类 ---

// ListCategories 列出分类。
func (s *Service) ListCategories(ctx context.Context) ([]CategoryInfo, error) {
	rows, err := s.db.Query(ctx, `
		SELECT slug, name, COALESCE(description,''), COALESCE(icon,''), sort_order
		FROM marketplace_categories ORDER BY sort_order ASC`)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()

	var out []CategoryInfo
	for rows.Next() {
		var c CategoryInfo
		if err := rows.Scan(&c.Slug, &c.Name, &c.Description, &c.Icon, &c.SortOrder); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// --- 应用目录 ---

// Create 创建应用条目。
func (s *Service) Create(ctx context.Context, in CreateAppInput) (*App, error) {
	if strings.TrimSpace(in.Name) == "" {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "name", Reason: "required"})
	}
	if in.Slug == "" {
		in.Slug = slugify(in.Name)
	}
	if !IsValidCategory(string(in.Category)) {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "category", Reason: "invalid value"})
	}
	if in.Pricing == "" {
		in.Pricing = PricingFree
	}
	if in.InstallType == "" {
		in.InstallType = InstallBuiltin
	}
	if in.Publisher == "" {
		in.Publisher = "official"
	}

	cfgJSON := "{}"
	if len(in.InstallConfig) > 0 {
		if b, err := json.Marshal(in.InstallConfig); err == nil {
			cfgJSON = string(b)
		}
	}

	var a App
	err := s.db.QueryRow(ctx, `
		INSERT INTO marketplace_apps
			(id, slug, name, description, category, publisher, publisher_url, icon_url,
			 screenshots, tags, pricing, install_type, install_config, docs_url, created_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9::text[],$10::text[],$11,$12,$13::jsonb,$14,$15)
		RETURNING id, slug, name, COALESCE(description,''), category, publisher,
			COALESCE(publisher_url,''), COALESCE(icon_url,''), COALESCE(screenshots,'{}'),
			COALESCE(tags,'{}'), pricing, install_type, install_config, COALESCE(docs_url,''),
			rating_avg, rating_count, install_count, status, created_by, created_at, updated_at`,
		genID(), in.Slug, in.Name, nullIfEmpty(in.Description), string(in.Category),
		in.Publisher, nullIfEmpty(in.PublisherURL), nullIfEmpty(in.IconURL),
		strSlice(in.Screenshots), strSlice(in.Tags), string(in.Pricing),
		string(in.InstallType), cfgJSON, nullIfEmpty(in.DocsURL), in.CreatedBy,
	).Scan(&a.ID, &a.Slug, &a.Name, &a.Description, &a.Category, &a.Publisher,
		&a.PublisherURL, &a.IconURL, &a.Screenshots, &a.Tags, &a.Pricing,
		&a.InstallType, &a.InstallConfig, &a.DocsURL, &a.RatingAvg, &a.RatingCount,
		&a.InstallCount, &a.Status, &a.CreatedBy, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, ErrAppSlugTaken
		}
		return nil, errs.ErrInternal.Wrap(err)
	}
	return &a, nil
}

// GetBySlug 按 slug 查询。
func (s *Service) GetBySlug(ctx context.Context, slug string) (*App, error) {
	var a App
	err := s.db.QueryRow(ctx, `
		SELECT id, slug, name, COALESCE(description,''), category, publisher,
			COALESCE(publisher_url,''), COALESCE(icon_url,''), COALESCE(screenshots,'{}'),
			COALESCE(tags,'{}'), pricing, install_type, install_config, COALESCE(docs_url,''),
			rating_avg, rating_count, install_count, status, created_by, created_at, updated_at
		FROM marketplace_apps WHERE slug = $1`, slug,
	).Scan(&a.ID, &a.Slug, &a.Name, &a.Description, &a.Category, &a.Publisher,
		&a.PublisherURL, &a.IconURL, &a.Screenshots, &a.Tags, &a.Pricing,
		&a.InstallType, &a.InstallConfig, &a.DocsURL, &a.RatingAvg, &a.RatingCount,
		&a.InstallCount, &a.Status, &a.CreatedBy, &a.CreatedAt, &a.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrAppNotFound
	}
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	return &a, nil
}

// List 列出应用（可按分类/关键字过滤）。
func (s *Service) List(ctx context.Context, category Category, q string, limit, offset int) ([]*App, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	conds := []string{"status = 'published'"}
	args := []any{}
	if category != "" {
		args = append(args, string(category))
		conds = append(conds, "category = $"+itoa(len(args)))
	}
	if strings.TrimSpace(q) != "" {
		args = append(args, "%"+strings.TrimSpace(q)+"%")
		conds = append(conds, "(name ILIKE $"+itoa(len(args))+" OR description ILIKE $"+itoa(len(args))+")")
	}
	args = append(args, limit, offset)

	rows, err := s.db.Query(ctx, `
		SELECT id, slug, name, COALESCE(description,''), category, publisher,
			COALESCE(publisher_url,''), COALESCE(icon_url,''), COALESCE(screenshots,'{}'),
			COALESCE(tags,'{}'), pricing, install_type, install_config, COALESCE(docs_url,''),
			rating_avg, rating_count, install_count, status, created_by, created_at, updated_at
		FROM marketplace_apps
		WHERE `+strings.Join(conds, " AND ")+`
		ORDER BY install_count DESC, rating_avg DESC
		LIMIT $`+itoa(len(args)-1)+` OFFSET $`+itoa(len(args)), args...)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()

	var out []*App
	for rows.Next() {
		var a App
		if err := rows.Scan(&a.ID, &a.Slug, &a.Name, &a.Description, &a.Category, &a.Publisher,
			&a.PublisherURL, &a.IconURL, &a.Screenshots, &a.Tags, &a.Pricing,
			&a.InstallType, &a.InstallConfig, &a.DocsURL, &a.RatingAvg, &a.RatingCount,
			&a.InstallCount, &a.Status, &a.CreatedBy, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
		out = append(out, &a)
	}
	return out, rows.Err()
}

// Update 更新应用。
func (s *Service) Update(ctx context.Context, slug string, in UpdateAppInput) (*App, error) {
	sets := []string{}
	args := []any{}
	add := func(expr string, v any) {
		args = append(args, v)
		sets = append(sets, expr+"$"+itoa(len(args)))
	}
	if in.Name != nil {
		add("name = ", *in.Name)
	}
	if in.Description != nil {
		add("description = ", nullIfEmpty(*in.Description))
	}
	if in.PublisherURL != nil {
		add("publisher_url = ", nullIfEmpty(*in.PublisherURL))
	}
	if in.IconURL != nil {
		add("icon_url = ", nullIfEmpty(*in.IconURL))
	}
	if in.Screenshots != nil {
		add("screenshots = ", strSlice(in.Screenshots))
	}
	if in.Tags != nil {
		add("tags = ", strSlice(in.Tags))
	}
	if in.Pricing != nil {
		add("pricing = ", string(*in.Pricing))
	}
	if in.InstallConfig != nil {
		b, _ := json.Marshal(in.InstallConfig)
		add("install_config = ", string(b))
	}
	if in.DocsURL != nil {
		add("docs_url = ", nullIfEmpty(*in.DocsURL))
	}
	if in.Status != nil {
		add("status = ", string(*in.Status))
	}
	if len(sets) == 0 {
		return s.GetBySlug(ctx, slug)
	}
	sets = append(sets, "updated_at = now()")
	args = append(args, slug)

	var a App
	err := s.db.QueryRow(ctx, `
		UPDATE marketplace_apps SET `+strings.Join(sets, ", ")+`
		WHERE slug = $`+itoa(len(args))+`
		RETURNING id, slug, name, COALESCE(description,''), category, publisher,
			COALESCE(publisher_url,''), COALESCE(icon_url,''), COALESCE(screenshots,'{}'),
			COALESCE(tags,'{}'), pricing, install_type, install_config, COALESCE(docs_url,''),
			rating_avg, rating_count, install_count, status, created_by, created_at, updated_at`,
		args...,
	).Scan(&a.ID, &a.Slug, &a.Name, &a.Description, &a.Category, &a.Publisher,
		&a.PublisherURL, &a.IconURL, &a.Screenshots, &a.Tags, &a.Pricing,
		&a.InstallType, &a.InstallConfig, &a.DocsURL, &a.RatingAvg, &a.RatingCount,
		&a.InstallCount, &a.Status, &a.CreatedBy, &a.CreatedAt, &a.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrAppNotFound
	}
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	return &a, nil
}

// --- 安装 ---

// Install 安装应用到工作空间（幂等）。
func (s *Service) Install(ctx context.Context, in InstallInput) (*Install, error) {
	if in.WorkspaceID <= 0 || in.AppID <= 0 {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "workspace_id/app_id", Reason: "required"})
	}
	cfgJSON := "{}"
	if len(in.Config) > 0 {
		if b, err := json.Marshal(in.Config); err == nil {
			cfgJSON = string(b)
		}
	}

	var inst Install
	err := s.db.QueryRow(ctx, `
		INSERT INTO marketplace_installs (id, workspace_id, app_id, installed_by, config)
		VALUES ($1,$2,$3,$4,$5::jsonb)
		ON CONFLICT (workspace_id, app_id) DO UPDATE SET
			config = EXCLUDED.config,
			enabled = TRUE,
			updated_at = now()
		RETURNING id, workspace_id, app_id, installed_by, config, enabled, installed_at, updated_at`,
		genID(), in.WorkspaceID, in.AppID, in.InstalledBy, cfgJSON,
	).Scan(&inst.ID, &inst.WorkspaceID, &inst.AppID, &inst.InstalledBy,
		&inst.Config, &inst.Enabled, &inst.InstalledAt, &inst.UpdatedAt)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}

	// 更新安装计数
	_, _ = s.db.Exec(ctx, `
		UPDATE marketplace_apps
		SET install_count = (SELECT COUNT(*) FROM marketplace_installs WHERE app_id = $1 AND enabled),
		    updated_at = now()
		WHERE id = $1`, in.AppID)
	return &inst, nil
}

// Uninstall 卸载（软删除：enabled = false）。
func (s *Service) Uninstall(ctx context.Context, workspaceID, appID int64) error {
	tag, err := s.db.Exec(ctx, `
		UPDATE marketplace_installs SET enabled = false, updated_at = now()
		WHERE workspace_id = $1 AND app_id = $2`, workspaceID, appID)
	if err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	if tag.RowsAffected() == 0 {
		return ErrAppNotFound
	}
	_, _ = s.db.Exec(ctx, `
		UPDATE marketplace_apps
		SET install_count = (SELECT COUNT(*) FROM marketplace_installs WHERE app_id = $1 AND enabled),
		    updated_at = now()
		WHERE id = $1`, appID)
	return nil
}

// ListInstalled 列出工作空间已安装应用。
func (s *Service) ListInstalled(ctx context.Context, workspaceID int64) ([]*App, error) {
	rows, err := s.db.Query(ctx, `
		SELECT a.id, a.slug, a.name, COALESCE(a.description,''), a.category, a.publisher,
			COALESCE(a.publisher_url,''), COALESCE(a.icon_url,''), COALESCE(a.screenshots,'{}'),
			COALESCE(a.tags,'{}'), a.pricing, a.install_type, a.install_config, COALESCE(a.docs_url,''),
			a.rating_avg, a.rating_count, a.install_count, a.status, a.created_by, a.created_at, a.updated_at
		FROM marketplace_installs i
		JOIN marketplace_apps a ON a.id = i.app_id
		WHERE i.workspace_id = $1 AND i.enabled
		ORDER BY i.installed_at DESC`, workspaceID)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()

	var out []*App
	for rows.Next() {
		var a App
		if err := rows.Scan(&a.ID, &a.Slug, &a.Name, &a.Description, &a.Category, &a.Publisher,
			&a.PublisherURL, &a.IconURL, &a.Screenshots, &a.Tags, &a.Pricing,
			&a.InstallType, &a.InstallConfig, &a.DocsURL, &a.RatingAvg, &a.RatingCount,
			&a.InstallCount, &a.Status, &a.CreatedBy, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
		out = append(out, &a)
	}
	return out, rows.Err()
}

// --- 评分 ---

// CreateReview 提交/更新评分。
func (s *Service) CreateReview(ctx context.Context, in ReviewInput) (*Review, error) {
	if in.Rating < 1 || in.Rating > 5 {
		return nil, ErrInvalidRating
	}
	if in.AppID <= 0 || in.WorkspaceID <= 0 || in.UserID <= 0 {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "ids", Reason: "required"})
	}

	var r Review
	err := s.db.QueryRow(ctx, `
		INSERT INTO marketplace_reviews (id, app_id, workspace_id, user_id, rating, comment)
		VALUES ($1,$2,$3,$4,$5,$6)
		ON CONFLICT (app_id, user_id) DO UPDATE SET
			rating = EXCLUDED.rating,
			comment = EXCLUDED.comment,
			updated_at = now()
		RETURNING id, app_id, workspace_id, user_id, rating, COALESCE(comment,''), created_at, updated_at`,
		genID(), in.AppID, in.WorkspaceID, in.UserID, in.Rating, nullIfEmpty(in.Comment),
	).Scan(&r.ID, &r.AppID, &r.WorkspaceID, &r.UserID, &r.Rating, &r.Comment,
		&r.CreatedAt, &r.UpdatedAt)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}

	// 更新评分聚合
	_, _ = s.db.Exec(ctx, `
		UPDATE marketplace_apps
		SET rating_avg = COALESCE((SELECT AVG(rating)::numeric(3,2) FROM marketplace_reviews WHERE app_id = $1), 0),
		    rating_count = (SELECT COUNT(*) FROM marketplace_reviews WHERE app_id = $1),
		    updated_at = now()
		WHERE id = $1`, in.AppID)
	return &r, nil
}

// ListReviews 列出应用评分。
func (s *Service) ListReviews(ctx context.Context, appID int64, limit int) ([]*Review, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	rows, err := s.db.Query(ctx, `
		SELECT id, app_id, workspace_id, user_id, rating, COALESCE(comment,''), created_at, updated_at
		FROM marketplace_reviews WHERE app_id = $1
		ORDER BY created_at DESC LIMIT $2`, appID, limit)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()

	var out []*Review
	for rows.Next() {
		var r Review
		if err := rows.Scan(&r.ID, &r.AppID, &r.WorkspaceID, &r.UserID, &r.Rating,
			&r.Comment, &r.CreatedAt, &r.UpdatedAt); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
		out = append(out, &r)
	}
	return out, rows.Err()
}

// --- helpers ---

func nullIfEmpty(s string) any {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return s
}

func strSlice(v []string) any {
	if len(v) == 0 {
		return nil
	}
	return v
}

func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var sb strings.Builder
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			sb.WriteRune(r)
		case r == ' ' || r == '-' || r == '_':
			sb.WriteByte('-')
		}
	}
	return strings.Trim(sb.String(), "-")
}

func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "23505") || strings.Contains(err.Error(), "duplicate key")
}

func itoa(n int) string {
	if n < 10 {
		return string(rune('0' + n))
	}
	buf := []byte{}
	for n > 0 {
		buf = append([]byte{byte('0' + n%10)}, buf...)
		n /= 10
	}
	return string(buf)
}
