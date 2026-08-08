package issue

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// WorkitemArchiveService 工作项归档服务
// 实现差异化归档策略：
// - 缺陷(defect)：关闭超过2年自动归档
// - 任务(task)：关闭超过3年自动归档
// - 需求(requirement)：关闭超过5年自动归档
// 已归档的工作项不参与日常查询、搜索索引，只读归档库可查询
type WorkitemArchiveService struct {
	db                 *pgxpool.Pool
	defectArchiveYears  int
	taskArchiveYears    int
	requirementArchiveYears int
}

// NewWorkitemArchiveService 创建归档服务，传入各类工作项的归档年限
func NewWorkitemArchiveService(db *pgxpool.Pool) *WorkitemArchiveService {
	return &WorkitemArchiveService{
		db:                 db,
		defectArchiveYears:  2,
		taskArchiveYears:    3,
		requirementArchiveYears: 5,
	}
}

// ArchiveExpiredWorkitems 执行过期工作项归档，返回归档的工作项数量
func (s *WorkitemArchiveService) ArchiveExpiredWorkitems(ctx context.Context) (int, error) {
	totalArchived := 0
	
	// 1. 归档过期缺陷
	defectCutoff := time.Now().AddDate(-s.defectArchiveYears, 0, 0)
	var defectCount int
	err := s.db.QueryRow(ctx, `
		UPDATE defect 
		SET archived_at = now()
		WHERE completed_at IS NOT NULL 
		  AND completed_at < $1 
		  AND archived_at IS NULL
		RETURNING COUNT(*)
	`, defectCutoff).Scan(&defectCount)
	if err == nil {
		totalArchived += defectCount
	}
	
	// 2. 归档过期任务
	taskCutoff := time.Now().AddDate(-s.taskArchiveYears, 0, 0)
	var taskCount int
	err = s.db.QueryRow(ctx, `
		UPDATE task 
		SET archived_at = now()
		WHERE completed_at IS NOT NULL 
		  AND completed_at < $1 
		  AND archived_at IS NULL
		RETURNING COUNT(*)
	`, taskCutoff).Scan(&taskCount)
	if err == nil {
		totalArchived += taskCount
	}
	
	// 3. 归档过期需求
	reqCutoff := time.Now().AddDate(-s.requirementArchiveYears, 0, 0)
	var reqCount int
	err = s.db.QueryRow(ctx, `
		UPDATE requirement 
		SET archived_at = now()
		WHERE completed_at IS NOT NULL 
		  AND completed_at < $1 
		  AND archived_at IS NULL
		RETURNING COUNT(*)
	`, reqCutoff).Scan(&reqCount)
	if err == nil {
		totalArchived += reqCount
	}
	
	return totalArchived, nil
}

// RestoreArchivedWorkitem 恢复已归档的工作项
func (s *WorkitemArchiveService) RestoreArchivedWorkitem(ctx context.Context, wsID int64, entityType IssueTypeCode, entityID int64) error {
	var tableName string
	switch entityType {
	case TypeTask:
		tableName = "task"
	case TypeRequirement:
		tableName = "requirement"
	case TypeDefect:
		tableName = "defect"
	default:
		return errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "entity_type", Reason: "不支持的工作项类型"})
	}
	
	_, err := s.db.Exec(ctx, `
		UPDATE `+tableName+` 
		SET archived_at = NULL
		WHERE id = $1 AND workspace_id = $2 AND archived_at IS NOT NULL
	`, entityID, wsID)
	return err
}

// IsArchived 检查工作项是否已归档
func (s *WorkitemArchiveService) IsArchived(ctx context.Context, entityType IssueTypeCode, entityID int64) (bool, error) {
	var tableName string
	switch entityType {
	case TypeTask:
		tableName = "task"
	case TypeRequirement:
		tableName = "requirement"
	case TypeDefect:
		tableName = "defect"
	default:
		return false, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "entity_type", Reason: "不支持的工作项类型"})
	}
	
	var archived bool
	err := s.db.QueryRow(ctx, `
		SELECT archived_at IS NOT NULL FROM `+tableName+` WHERE id = $1
	`, entityID).Scan(&archived)
	return archived, err
}
