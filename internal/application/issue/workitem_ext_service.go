package issue

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/xeipuuv/gojsonschema"

	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// WorkitemExtService 工作项扩展属性服务
// 支持三类工作项的自定义字段配置、值的CRUD和schema校验
type WorkitemExtService struct {
	db *pgxpool.Pool
}

// NewWorkitemExtService 创建扩展属性服务
func NewWorkitemExtService(db *pgxpool.Pool) *WorkitemExtService {
	return &WorkitemExtService{db: db}
}

// ValidateExtValue 校验扩展属性值是否符合schema定义
func (s *WorkitemExtService) ValidateExtValue(fieldSchema map[string]any, fieldValue map[string]any) error {
	if fieldSchema == nil || len(fieldSchema) == 0 {
		return nil
	}
	schemaLoader := gojsonschema.NewGoLoader(fieldSchema)
	valueLoader := gojsonschema.NewGoLoader(fieldValue)
	result, err := gojsonschema.Validate(schemaLoader, valueLoader)
	if err != nil {
		return errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "ext_field", Reason: fmt.Sprintf("schema解析失败: %v", err)})
	}
	if !result.Valid() {
		var errDetails []string
		for _, desc := range result.Errors() {
			errDetails = append(errDetails, desc.String())
		}
		return errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "ext_field", Reason: fmt.Sprintf("字段校验失败: %v", errDetails)})
	}
	return nil
}

// CreateOrUpdate 创建或更新扩展属性值
func (s *WorkitemExtService) CreateOrUpdate(ctx context.Context, wsID, projectID int64, entityType IssueTypeCode, entityID int64, fieldName string, fieldValue map[string]any, fieldSchema map[string]any, userID int64) error {
	// 先校验值是否符合schema
	if err := s.ValidateExtValue(fieldSchema, fieldValue); err != nil {
		return err
	}
	
	return s.withTx(ctx, wsID, func(tx pgx.Tx) error {
		var tableName string
		var entityColName string
		switch entityType {
		case TypeTask:
			tableName = "task_ext"
			entityColName = "task_id"
		case TypeRequirement:
			tableName = "requirement_ext"
			entityColName = "requirement_id"
		case TypeDefect:
			tableName = "defect_ext"
			entityColName = "defect_id"
		default:
			return errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "entity_type", Reason: "不支持的工作项类型"})
		}
		
		valueByte, _ := json.Marshal(fieldValue)
		schemaByte, _ := json.Marshal(fieldSchema)
		
		//  upsert操作：存在则更新，不存在则创建
		_, err := tx.Exec(ctx, fmt.Sprintf(`
			INSERT INTO %s (workspace_id, project_id, %s, field_name, field_value, field_schema, created_by, updated_at)
			VALUES ($1, $2, $3, $4, $5::jsonb, $6::jsonb, $7, now())
			ON CONFLICT (%s, field_name) DO UPDATE SET
				field_value = EXCLUDED.field_value,
				field_schema = EXCLUDED.field_schema,
				updated_at = now()
		`, tableName, entityColName, entityColName),
			wsID, projectID, entityID, fieldName, string(valueByte), string(schemaByte), userID)
		return err
	})
}

// GetByEntity 查询工作项的所有扩展属性
func (s *WorkitemExtService) GetByEntity(ctx context.Context, wsID int64, entityType IssueTypeCode, entityID int64) ([]WorkitemExtension, error) {
	var tableName string
	var entityColName string
	switch entityType {
	case TypeTask:
		tableName = "task_ext"
		entityColName = "task_id"
	case TypeRequirement:
		tableName = "requirement_ext"
		entityColName = "requirement_id"
	case TypeDefect:
		tableName = "defect_ext"
		entityColName = "defect_id"
	default:
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "entity_type", Reason: "不支持的工作项类型"})
	}
	
	rows, err := s.db.Query(ctx, fmt.Sprintf(`
		SELECT id, workspace_id, project_id, %s, field_name, field_value, field_schema, created_by, created_at, updated_at
		FROM %s 
		WHERE workspace_id = $1 AND %s = $2 AND deleted = false
	`, entityColName, tableName, entityColName), wsID, entityID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var exts []WorkitemExtension
	for rows.Next() {
		var ext WorkitemExtension
		var valueByte, schemaByte []byte
		if err := rows.Scan(&ext.ID, &ext.WorkspaceID, &ext.ProjectID, &ext.EntityID, &ext.FieldName, &valueByte, &schemaByte, &ext.CreatedBy, &ext.CreatedAt, &ext.UpdatedAt); err != nil {
			return nil, err
		}
		ext.EntityType = entityType
		_ = json.Unmarshal(valueByte, &ext.FieldValue)
		_ = json.Unmarshal(schemaByte, &ext.FieldSchema)
		exts = append(exts, ext)
	}
	return exts, nil
}

// Delete 删除工作项的某个扩展属性
func (s *WorkitemExtService) Delete(ctx context.Context, wsID, entityID int64, entityType IssueTypeCode, fieldName string) error {
	var tableName string
	var entityColName string
	switch entityType {
	case TypeTask:
		tableName = "task_ext"
		entityColName = "task_id"
	case TypeRequirement:
		tableName = "requirement_ext"
		entityColName = "requirement_id"
	case TypeDefect:
		tableName = "defect_ext"
		entityColName = "defect_id"
	default:
		return errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "entity_type", Reason: "不支持的工作项类型"})
	}
	
	_, err := s.db.Exec(ctx, fmt.Sprintf(`
		DELETE FROM %s WHERE workspace_id = $1 AND %s = $2 AND field_name = $3
	`, tableName, entityColName), wsID, entityID, fieldName)
	return err
}

// DeleteAllByEntity 删除工作项的所有扩展属性（工作项删除时调用）
func (s *WorkitemExtService) DeleteAllByEntity(ctx context.Context, wsID, entityID int64, entityType IssueTypeCode) error {
	var tableName string
	var entityColName string
	switch entityType {
	case TypeTask:
		tableName = "task_ext"
		entityColName = "task_id"
	case TypeRequirement:
		tableName = "requirement_ext"
		entityColName = "requirement_id"
	case TypeDefect:
		tableName = "defect_ext"
		entityColName = "defect_id"
	default:
		return errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "entity_type", Reason: "不支持的工作项类型"})
	}
	
	_, err := s.db.Exec(ctx, fmt.Sprintf(`
		DELETE FROM %s WHERE workspace_id = $1 AND %s = $2
	`, tableName, entityColName), wsID, entityID)
	return err
}

// withTx 事务辅助函数，和issue_service里的对齐
func (s *WorkitemExtService) withTx(ctx context.Context, wsID int64, fn func(pgx.Tx) error) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	
	if _, err := tx.Exec(ctx, "SELECT set_config('app.workspace_id', $1, true)", wsID); err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	
	if err := fn(tx); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
