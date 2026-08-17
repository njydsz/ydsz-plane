// Package issue — 内容模板（需求/任务/缺陷模板库）。
//
// 允许用户将常用工作项结构保存为模板，创建时从模板预填字段。
// 模板类型：requirement / task / defect。
// 存储：content_templates 表（JSONB 存储模板内容）。
package issue

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// ContentTemplateService 管理内容模板。
type ContentTemplateService struct {
	db *pgxpool.Pool
}

// NewContentTemplateService 创建内容模板服务。
func NewContentTemplateService(db *pgxpool.Pool) *ContentTemplateService {
	return &ContentTemplateService{db: db}
}

// ContentTemplate 内容模板实体。
type ContentTemplate struct {
	ID           int64           `json:"id"`
	TenantID     int64           `json:"tenant_id"`
	WorkspaceID  int64           `json:"workspace_id"`
	ProjectID    *int64          `json:"project_id,omitempty"`
	Name         string          `json:"name"`
	TemplateType string          `json:"template_type"` // requirement | task | defect
	ContentJSON  json.RawMessage `json:"content_json"`
	ContentHTML  string          `json:"content_html,omitempty"`
	IsDefault    bool            `json:"is_default"`
	Status       string          `json:"status"`
	CreatedBy    int64           `json:"created_by"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}

// CreateTemplateInput 创建模板入参。
type CreateTemplateInput struct {
	WorkspaceID  int64
	ProjectID    *int64
	Name         string
	TemplateType string
	ContentJSON  map[string]any
	ContentHTML  string
	IsDefault    bool
	CreatedBy    int64
}

// UpdateTemplateInput 更新模板入参。
type UpdateTemplateInput struct {
	ID           int64
	WorkspaceID  int64
	Name         *string
	ContentJSON  map[string]any
	ContentHTML  *string
	IsDefault    *bool
}

// ListTemplatesFilter 模板列表筛选。
type ListTemplatesFilter struct {
	WorkspaceID  int64
	ProjectID    *int64
	TemplateType string // 空 = 全部
}

// Create 创建内容模板。
func (s *ContentTemplateService) Create(ctx context.Context, in CreateTemplateInput) (*ContentTemplate, error) {
	if in.Name == "" {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "name", Reason: "模板名称不能为空"})
	}
	if in.TemplateType != "requirement" && in.TemplateType != "task" && in.TemplateType != "defect" {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "template_type", Reason: "模板类型必须是 requirement / task / defect"})
	}

	contentJSON, err := json.Marshal(in.ContentJSON)
	if err != nil {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "content_json", Reason: "无效的 JSON 内容"})
	}

	var t ContentTemplate
	err = s.db.QueryRow(ctx, `
		INSERT INTO content_templates (tenant_id, workspace_id, project_id, name, template_type, content_json, content_html, is_default, status, created_by)
		VALUES (1, $1, $2, $3, $4, $5, $6, $7, 'active', $8)
		RETURNING id, tenant_id, workspace_id, project_id, name, template_type, content_json, content_html, is_default, status, created_by, created_at, updated_at`,
		in.WorkspaceID, in.ProjectID, in.Name, in.TemplateType, contentJSON, in.ContentHTML, in.IsDefault, in.CreatedBy).Scan(
		&t.ID, &t.TenantID, &t.WorkspaceID, &t.ProjectID, &t.Name, &t.TemplateType,
		&t.ContentJSON, &t.ContentHTML, &t.IsDefault, &t.Status, &t.CreatedBy, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("content_template.Create: %w", err)
	}
	return &t, nil
}

// Get 获取单个模板。
func (s *ContentTemplateService) Get(ctx context.Context, wsID, templateID int64) (*ContentTemplate, error) {
	var t ContentTemplate
	err := s.db.QueryRow(ctx, `
		SELECT id, tenant_id, workspace_id, project_id, name, template_type, content_json, content_html, is_default, status, created_by, created_at, updated_at
		FROM content_templates
		WHERE id = $1 AND workspace_id = $2 AND deleted = false`,
		templateID, wsID).Scan(
		&t.ID, &t.TenantID, &t.WorkspaceID, &t.ProjectID, &t.Name, &t.TemplateType,
		&t.ContentJSON, &t.ContentHTML, &t.IsDefault, &t.Status, &t.CreatedBy, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errs.ErrNotFound
		}
		return nil, fmt.Errorf("content_template.Get: %w", err)
	}
	return &t, nil
}

// List 列出模板（按工作空间 + 可选类型筛选）。
func (s *ContentTemplateService) List(ctx context.Context, filter ListTemplatesFilter) ([]ContentTemplate, error) {
	var args []any
	argIdx := 1
	query := `SELECT id, tenant_id, workspace_id, project_id, name, template_type, content_json, content_html, is_default, status, created_by, created_at, updated_at
	          FROM content_templates WHERE workspace_id = $` + strconv.Itoa(argIdx) + ` AND deleted = false`
	args = append(args, filter.WorkspaceID)
	argIdx++

	if filter.TemplateType != "" {
		query += ` AND template_type = $` + strconv.Itoa(argIdx)
		args = append(args, filter.TemplateType)
		argIdx++
	}

	if filter.ProjectID != nil {
		query += ` AND (project_id = $` + strconv.Itoa(argIdx) + ` OR project_id IS NULL)`
		args = append(args, *filter.ProjectID)
		argIdx++
	}

	query += ` ORDER BY is_default DESC, created_at DESC`

	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("content_template.List: %w", err)
	}
	defer rows.Close()

	var templates []ContentTemplate
	for rows.Next() {
		var t ContentTemplate
		if err := rows.Scan(&t.ID, &t.TenantID, &t.WorkspaceID, &t.ProjectID, &t.Name, &t.TemplateType,
			&t.ContentJSON, &t.ContentHTML, &t.IsDefault, &t.Status, &t.CreatedBy, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, fmt.Errorf("content_template.List scan: %w", err)
		}
		templates = append(templates, t)
	}
	return templates, rows.Err()
}

// Update 更新模板。
func (s *ContentTemplateService) Update(ctx context.Context, in UpdateTemplateInput) (*ContentTemplate, error) {
	var sets []string
	var args []any
	argIdx := 1

	if in.Name != nil {
		sets = append(sets, fmt.Sprintf("name = $%d", argIdx))
		args = append(args, *in.Name)
		argIdx++
	}
	if in.ContentJSON != nil {
		contentJSON, err := json.Marshal(in.ContentJSON)
		if err != nil {
			return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "content_json", Reason: "无效的 JSON 内容"})
		}
		sets = append(sets, fmt.Sprintf("content_json = $%d", argIdx))
		args = append(args, contentJSON)
		argIdx++
	}
	if in.ContentHTML != nil {
		sets = append(sets, fmt.Sprintf("content_html = $%d", argIdx))
		args = append(args, *in.ContentHTML)
		argIdx++
	}
	if in.IsDefault != nil {
		sets = append(sets, fmt.Sprintf("is_default = $%d", argIdx))
		args = append(args, *in.IsDefault)
		argIdx++
	}

	if len(sets) == 0 {
		return s.Get(ctx, in.WorkspaceID, in.ID)
	}

	sets = append(sets, "updated_at = now()")
	args = append(args, in.ID, in.WorkspaceID)

	var t ContentTemplate
	err := s.db.QueryRow(ctx, `
		UPDATE content_templates SET `+fmt.Sprintf("%s", joinStrings(sets, ", "))+`
		WHERE id = $`+strconv.Itoa(argIdx)+` AND workspace_id = $`+strconv.Itoa(argIdx+1)+` AND deleted = false
		RETURNING id, tenant_id, workspace_id, project_id, name, template_type, content_json, content_html, is_default, status, created_by, created_at, updated_at`,
		args...).Scan(
		&t.ID, &t.TenantID, &t.WorkspaceID, &t.ProjectID, &t.Name, &t.TemplateType,
		&t.ContentJSON, &t.ContentHTML, &t.IsDefault, &t.Status, &t.CreatedBy, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errs.ErrNotFound
		}
		return nil, fmt.Errorf("content_template.Update: %w", err)
	}
	return &t, nil
}

// Delete 软删除模板。
func (s *ContentTemplateService) Delete(ctx context.Context, wsID, templateID int64) error {
	cmd, err := s.db.Exec(ctx, `
		UPDATE content_templates SET deleted = true, updated_at = now()
		WHERE id = $1 AND workspace_id = $2 AND deleted = false`,
		templateID, wsID)
	if err != nil {
		return fmt.Errorf("content_template.Delete: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return errs.ErrNotFound
	}
	return nil
}

// joinStrings 辅助函数。
func joinStrings(strs []string, sep string) string {
	result := ""
	for i, s := range strs {
		if i > 0 {
			result += sep
		}
		result += s
	}
	return result
}
