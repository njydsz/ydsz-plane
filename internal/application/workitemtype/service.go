// Package workitemtype — 自定义工作项类型应用服务。
//
// 职责：
//   - 工作项类型（WorkItemType）的 CRUD / 归档；
//   - 类型下自定义属性（WorkItemTypeProperty）的增删查；
//   - 工作项自定义属性值（WorkItemCustomFieldValue）的读写。
package workitemtype

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// Service 自定义工作项类型应用服务。
type Service struct {
	db *pgxpool.Pool
}

// NewService 创建服务。
func NewService(db *pgxpool.Pool) *Service {
	return &Service{db: db}
}

// --- 错误 ---

var (
	// ErrTypeNotFound 工作项类型不存在。
	ErrTypeNotFound = errors.New("work item type not found")
	// ErrPropertyNotFound 自定义属性不存在。
	ErrPropertyNotFound = errors.New("work item type property not found")
	// ErrTypeCodeConflict 同一项目内类型 code 已存在。
	ErrTypeCodeConflict = errors.New("work item type code already exists in project")
	// ErrPropertyCodeConflict 同一类型内属性 code 已存在。
	ErrPropertyCodeConflict = errors.New("work item type property code already exists")
	// ErrTypeArchived 类型已归档，不可修改。
	ErrTypeArchived = errors.New("work item type is archived")
	// ErrInvalidFieldType 字段类型不合法。
	ErrInvalidFieldType = errors.New("invalid field type")
)

// ---------------------------------------------------------------------------
// WorkItemType CRUD
// ---------------------------------------------------------------------------

// Create 创建工作项类型。
//
// code 在同一项目内唯一；若指定为默认类型，则将同项目其它类型 is_default 置 false。
func (s *Service) Create(ctx context.Context, in CreateTypeInput) (*WorkItemType, error) {
	if in.ProjectID <= 0 {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "project_id", Reason: "required"})
	}
	code := strings.TrimSpace(in.Code)
	if code == "" {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "code", Reason: "required"})
	}
	if strings.TrimSpace(in.Name) == "" {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "name", Reason: "required"})
	}
	if in.Level < 0 || in.Level > 1 {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "level", Reason: "must be 0 or 1"})
	}
	if err := s.checkTypeCodeUnique(ctx, in.ProjectID, code, 0); err != nil {
		return nil, err
	}

	// 默认类型：同项目内互斥。
	if in.IsDefault {
		if _, err := s.db.Exec(ctx,
			`UPDATE work_item_types SET is_default = FALSE, updated_at = now()
			 WHERE project_id = $1 AND deleted_at IS NULL`, in.ProjectID); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
	}

	var t WorkItemType
	err := s.db.QueryRow(ctx, `
		INSERT INTO work_item_types
			(id, project_id, name, code, color, icon, description, level, is_default)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		RETURNING id, project_id, name, code, color, icon, COALESCE(description,''),
			level, is_active, is_default, created_at, updated_at, archived_at`,
		genID(), in.ProjectID, strings.TrimSpace(in.Name), code,
		nullIfEmpty(in.Color), nullIfEmpty(in.Icon), nullIfEmpty(in.Description),
		in.Level, in.IsDefault,
	).Scan(&t.ID, &t.ProjectID, &t.Name, &t.Code, &t.Color, &t.Icon, &t.Description,
		&t.Level, &t.IsActive, &t.IsDefault, &t.CreatedAt, &t.UpdatedAt, &t.ArchivedAt)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	return &t, nil
}

// GetByID 查询工作项类型。
func (s *Service) GetByID(ctx context.Context, id int64) (*WorkItemType, error) {
	var t WorkItemType
	err := s.db.QueryRow(ctx, `
		SELECT id, project_id, name, code, color, icon, COALESCE(description,''),
			level, is_active, is_default, created_at, updated_at, archived_at
		FROM work_item_types WHERE id = $1 AND deleted_at IS NULL`, id,
	).Scan(&t.ID, &t.ProjectID, &t.Name, &t.Code, &t.Color, &t.Icon, &t.Description,
		&t.Level, &t.IsActive, &t.IsDefault, &t.CreatedAt, &t.UpdatedAt, &t.ArchivedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrTypeNotFound
	}
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	return &t, nil
}

// List 列出项目的工作项类型（不含已归档）。
func (s *Service) List(ctx context.Context, opts ListTypesOptions) ([]*WorkItemType, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, project_id, name, code, color, icon, COALESCE(description,''),
			level, is_active, is_default, created_at, updated_at, archived_at
		FROM work_item_types
		WHERE project_id = $1 AND deleted_at IS NULL AND archived_at IS NULL
		ORDER BY is_default DESC, created_at DESC
		LIMIT $2 OFFSET $3`,
		opts.ProjectID, opts.Limit, opts.Offset)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()

	var out []*WorkItemType
	for rows.Next() {
		var t WorkItemType
		if err := rows.Scan(&t.ID, &t.ProjectID, &t.Name, &t.Code, &t.Color, &t.Icon, &t.Description,
			&t.Level, &t.IsActive, &t.IsDefault, &t.CreatedAt, &t.UpdatedAt, &t.ArchivedAt); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
		out = append(out, &t)
	}
	return out, rows.Err()
}

// Update 更新工作项类型。
func (s *Service) Update(ctx context.Context, id int64, in UpdateTypeInput) (*WorkItemType, error) {
	// 归档类型不允许修改。
	cur, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if cur.ArchivedAt != nil {
		return nil, ErrTypeArchived
	}

	sets := []string{}
	args := []any{}
	add := func(expr string, val any) {
		args = append(args, val)
		sets = append(sets, expr+"$"+itoa(len(args)))
	}
	if in.Name != nil && strings.TrimSpace(*in.Name) != "" {
		add("name = ", strings.TrimSpace(*in.Name))
	}
	if in.Code != nil && strings.TrimSpace(*in.Code) != "" {
		code := strings.TrimSpace(*in.Code)
		if err := s.checkTypeCodeUnique(ctx, cur.ProjectID, code, id); err != nil {
			return nil, err
		}
		add("code = ", code)
	}
	if in.Color != nil {
		add("color = ", nullIfEmpty(*in.Color))
	}
	if in.Icon != nil {
		add("icon = ", nullIfEmpty(*in.Icon))
	}
	if in.Description != nil {
		add("description = ", nullIfEmpty(*in.Description))
	}
	if in.Level != nil {
		if *in.Level < 0 || *in.Level > 1 {
			return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "level", Reason: "must be 0 or 1"})
		}
		add("level = ", *in.Level)
	}
	if in.IsActive != nil {
		add("is_active = ", *in.IsActive)
	}
	if in.IsDefault != nil {
		add("is_default = ", *in.IsDefault)
	}
	if len(sets) == 0 {
		return s.GetByID(ctx, id)
	}

	// 默认类型互斥：置为默认时清掉同项目其它默认。
	if in.IsDefault != nil && *in.IsDefault {
		if _, err := s.db.Exec(ctx,
			`UPDATE work_item_types SET is_default = FALSE, updated_at = now()
			 WHERE project_id = $1 AND id <> $2 AND deleted_at IS NULL`, cur.ProjectID, id); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
	}

	args = append(args, id)
	query := "UPDATE work_item_types SET " + joinComma(sets) + ", updated_at = now() WHERE id = $" + itoa(len(args)) + " AND deleted_at IS NULL"
	var t WorkItemType
	err = s.db.QueryRow(ctx, query, args...).
		Scan(&t.ID, &t.ProjectID, &t.Name, &t.Code, &t.Color, &t.Icon, &t.Description,
			&t.Level, &t.IsActive, &t.IsDefault, &t.CreatedAt, &t.UpdatedAt, &t.ArchivedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrTypeNotFound
	}
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	return &t, nil
}

// Delete 软删除工作项类型（级联删除属性与值）。
func (s *Service) Delete(ctx context.Context, id int64) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	tag, err := tx.Exec(ctx, `DELETE FROM work_item_types WHERE id = $1 AND deleted_at IS NULL`, id)
	if err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	if tag.RowsAffected() == 0 {
		return ErrTypeNotFound
	}
	// 级联清理属性定义与属性值。
	if _, err := tx.Exec(ctx, `DELETE FROM work_item_custom_values WHERE property_id IN
		(SELECT id FROM work_item_type_properties WHERE type_id = $1)`, id); err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	if _, err := tx.Exec(ctx, `DELETE FROM work_item_type_properties WHERE type_id = $1`, id); err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	if err := tx.Commit(ctx); err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	return nil
}

// Archive 归档工作项类型（归档后不可修改，不再出现在列表）。
func (s *Service) Archive(ctx context.Context, id int64) error {
	tag, err := s.db.Exec(ctx, `
		UPDATE work_item_types SET archived_at = now(), updated_at = now(), is_active = FALSE
		WHERE id = $1 AND deleted_at IS NULL AND archived_at IS NULL`, id)
	if err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	if tag.RowsAffected() == 0 {
		return ErrTypeNotFound
	}
	return nil
}

// Restore 恢复已归档的工作项类型。
func (s *Service) Restore(ctx context.Context, id int64) error {
	tag, err := s.db.Exec(ctx, `
		UPDATE work_item_types SET archived_at = NULL, updated_at = now(), is_active = TRUE
		WHERE id = $1 AND deleted_at IS NULL AND archived_at IS NOT NULL`, id)
	if err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	if tag.RowsAffected() == 0 {
		return ErrTypeNotFound
	}
	return nil
}

// GetTypeByCode 按项目 + code 查询工作项类型。
func (s *Service) GetTypeByCode(ctx context.Context, projectID int64, code string) (*WorkItemType, error) {
	var t WorkItemType
	err := s.db.QueryRow(ctx, `
		SELECT id, project_id, name, code, color, icon, COALESCE(description,''),
			level, is_active, is_default, created_at, updated_at, archived_at
		FROM work_item_types
		WHERE project_id = $1 AND code = $2 AND deleted_at IS NULL`,
		projectID, strings.TrimSpace(code),
	).Scan(&t.ID, &t.ProjectID, &t.Name, &t.Code, &t.Color, &t.Icon, &t.Description,
		&t.Level, &t.IsActive, &t.IsDefault, &t.CreatedAt, &t.UpdatedAt, &t.ArchivedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrTypeNotFound
	}
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	return &t, nil
}

// ---------------------------------------------------------------------------
// Property CRUD
// ---------------------------------------------------------------------------

// AddProperty 为类型新增自定义属性。
func (s *Service) AddProperty(ctx context.Context, in CreatePropertyInput) (*WorkItemTypeProperty, error) {
	if in.TypeID <= 0 {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "type_id", Reason: "required"})
	}
	if strings.TrimSpace(in.Name) == "" {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "name", Reason: "required"})
	}
	code := strings.TrimSpace(in.Code)
	if code == "" {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "code", Reason: "required"})
	}
	ft := FieldType(strings.TrimSpace(in.FieldType))
	if !ft.IsValid() {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "field_type", Reason: "invalid value"})
	}
	if (ft == FieldTypeSelect || ft == FieldTypeMultiselect) && strings.TrimSpace(in.Options) != "" {
		if !json.Valid([]byte(in.Options)) {
			return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "options", Reason: "must be valid JSON"})
		}
	}
	// 校验类型存在。
	if _, err := s.GetByID(ctx, in.TypeID); err != nil {
		return nil, err
	}

	var p WorkItemTypeProperty
	err := s.db.QueryRow(ctx, `
		INSERT INTO work_item_type_properties
			(id, type_id, name, code, field_type, options, required, sort_order)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		RETURNING id, type_id, name, code, field_type, COALESCE(options,''),
			required, sort_order, created_at`,
		genID(), in.TypeID, strings.TrimSpace(in.Name), code, string(ft),
		nullIfEmpty(in.Options), in.Required, in.SortOrder,
	).Scan(&p.ID, &p.TypeID, &p.Name, &p.Code, &p.FieldType, &p.Options,
		&p.Required, &p.SortOrder, &p.CreatedAt)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, ErrPropertyCodeConflict
		}
		return nil, errs.ErrInternal.Wrap(err)
	}
	return &p, nil
}

// ListProperties 列出类型下的全部属性（按 sort_order 升序）。
func (s *Service) ListProperties(ctx context.Context, typeID int64) ([]*WorkItemTypeProperty, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, type_id, name, code, field_type, COALESCE(options,''),
			required, sort_order, created_at
		FROM work_item_type_properties
		WHERE type_id = $1
		ORDER BY sort_order ASC, created_at ASC`, typeID)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()

	var out []*WorkItemTypeProperty
	for rows.Next() {
		var p WorkItemTypeProperty
		if err := rows.Scan(&p.ID, &p.TypeID, &p.Name, &p.Code, &p.FieldType, &p.Options,
			&p.Required, &p.SortOrder, &p.CreatedAt); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
		out = append(out, &p)
	}
	return out, rows.Err()
}

// RemoveProperty 删除属性定义（级联删除该属性的全部值）。
func (s *Service) RemoveProperty(ctx context.Context, typeID, propertyID int64) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	tag, err := tx.Exec(ctx,
		`DELETE FROM work_item_type_properties WHERE id = $1 AND type_id = $2`, propertyID, typeID)
	if err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	if tag.RowsAffected() == 0 {
		return ErrPropertyNotFound
	}
	if _, err := tx.Exec(ctx, `DELETE FROM work_item_custom_values WHERE property_id = $1`, propertyID); err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	if err := tx.Commit(ctx); err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	return nil
}

// ---------------------------------------------------------------------------
// Custom Field Values
// ---------------------------------------------------------------------------

// SetCustomFieldValue 写入/更新工作项的自定义属性值（upsert）。
func (s *Service) SetCustomFieldValue(ctx context.Context, in SetCustomFieldValueInput) (*WorkItemCustomFieldValue, error) {
	if in.PropertyID <= 0 {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "property_id", Reason: "required"})
	}
	if in.WorkItemID <= 0 {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "work_item_id", Reason: "required"})
	}
	if strings.TrimSpace(in.WorkItemType) == "" {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "work_item_type", Reason: "required"})
	}
	// 校验属性存在。
	var typeID int64
	err := s.db.QueryRow(ctx,
		`SELECT type_id FROM work_item_type_properties WHERE id = $1`, in.PropertyID).Scan(&typeID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrPropertyNotFound
	}
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	_ = typeID

	var v WorkItemCustomFieldValue
	err = s.db.QueryRow(ctx, `
		INSERT INTO work_item_custom_values
			(id, property_id, work_item_id, work_item_type, value)
		VALUES ($1,$2,$3,$4,$5)
		ON CONFLICT (property_id, work_item_id) DO UPDATE
			SET value = EXCLUDED.value, work_item_type = EXCLUDED.work_item_type, updated_at = now()
		RETURNING id, property_id, work_item_id, work_item_type, value, created_at, updated_at`,
		genID(), in.PropertyID, in.WorkItemID, strings.TrimSpace(in.WorkItemType), in.Value,
	).Scan(&v.ID, &v.PropertyID, &v.WorkItemID, &v.WorkItemType, &v.Value, &v.CreatedAt, &v.UpdatedAt)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	return &v, nil
}

// GetCustomFieldValues 查询工作项的全部自定义属性值。
func (s *Service) GetCustomFieldValues(ctx context.Context, workItemID int64) ([]*WorkItemCustomFieldValue, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, property_id, work_item_id, work_item_type, value, created_at, updated_at
		FROM work_item_custom_values
		WHERE work_item_id = $1
		ORDER BY created_at ASC`, workItemID)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()

	var out []*WorkItemCustomFieldValue
	for rows.Next() {
		var v WorkItemCustomFieldValue
		if err := rows.Scan(&v.ID, &v.PropertyID, &v.WorkItemID, &v.WorkItemType, &v.Value,
			&v.CreatedAt, &v.UpdatedAt); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
		out = append(out, &v)
	}
	return out, rows.Err()
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

// checkTypeCodeUnique 校验项目内 code 唯一（excludeID > 0 时排除自身）。
func (s *Service) checkTypeCodeUnique(ctx context.Context, projectID int64, code string, excludeID int64) error {
	var exists bool
	err := s.db.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM work_item_types
			WHERE project_id = $1 AND code = $2 AND deleted_at IS NULL
			  AND ($3 <= 0 OR id <> $3)
		)`, projectID, code, excludeID).Scan(&exists)
	if err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	if exists {
		return ErrTypeCodeConflict
	}
	return nil
}

// isUniqueViolation 判断是否为 PostgreSQL 唯一约束冲突（SQLSTATE 23505）。
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}

func nullIfEmpty(s string) any {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return s
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

func joinComma(parts []string) string {
	out := ""
	for i, p := range parts {
		if i > 0 {
			out += ", "
		}
		out += p
	}
	return out
}

var _ = time.Now
