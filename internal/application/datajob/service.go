// Package datajob — 异步数据任务应用服务。
//
// 提供任务的创建、进度更新、状态流转与列表查询，支撑「数据任务」监控页。
package datajob

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// Service 数据任务应用服务。
type Service struct {
	db *pgxpool.Pool
}

// NewService 创建数据任务服务。
func NewService(db *pgxpool.Pool) *Service {
	return &Service{db: db}
}

// --- CRUD ---

// Create 创建任务（默认 pending 状态）。
func (s *Service) Create(ctx context.Context, in CreateJobInput) (*DataJob, error) {
	if err := validateCreateInput(in); err != nil {
		return nil, err
	}

	id := genID()
	var job DataJob
	err := s.db.QueryRow(ctx, `
		INSERT INTO data_jobs (
			id, tenant_id, workspace_id, project_id, job_type, name, parameters,
			progress, status, scheduled_at, triggered_by, created_by, updated_by
		) VALUES ($1,$2,$3,$4,$5,$6,$7,0,'pending',$8,$9,$10,$10)
		RETURNING id, tenant_id, workspace_id, project_id, job_type, name,
		          parameters, progress, status, error_message, scheduled_at,
		          executed_at, completed_at, duration_ms, triggered_by,
		          created_by, created_at, updated_at`,
		id, in.TenantID, in.WorkspaceID, in.ProjectID, in.JobType, in.Name,
		jsonOrEmpty(in.Parameters), in.ScheduledAt, in.TriggeredBy, in.CreatedBy,
	).Scan(
		&job.ID, &job.TenantID, &job.WorkspaceID, &job.ProjectID,
		&job.JobType, &job.Name, &job.Parameters, &job.Progress, &job.Status,
		&job.ErrorMessage, &job.ScheduledAt, &job.ExecutedAt, &job.CompletedAt,
		&job.DurationMs, &job.TriggeredBy, &job.CreatedBy, &job.CreatedAt,
		&job.UpdatedAt,
	)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	return &job, nil
}

// GetByID 按 ID 查询任务。
func (s *Service) GetByID(ctx context.Context, id int64) (*DataJob, error) {
	var job DataJob
	err := s.db.QueryRow(ctx, `
		SELECT id, tenant_id, workspace_id, project_id, job_type, name,
		       parameters, progress, status, error_message, scheduled_at,
		       executed_at, completed_at, duration_ms, triggered_by,
		       created_by, created_at, updated_at
		FROM data_jobs
		WHERE id = $1 AND deleted = false`, id,
	).Scan(
		&job.ID, &job.TenantID, &job.WorkspaceID, &job.ProjectID,
		&job.JobType, &job.Name, &job.Parameters, &job.Progress, &job.Status,
		&job.ErrorMessage, &job.ScheduledAt, &job.ExecutedAt, &job.CompletedAt,
		&job.DurationMs, &job.TriggeredBy, &job.CreatedBy, &job.CreatedAt,
		&job.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errs.ErrNotFound
		}
		return nil, errs.ErrInternal.Wrap(err)
	}
	return &job, nil
}

// List 按过滤条件查询任务列表（默认按创建时间倒序）。
func (s *Service) List(ctx context.Context, f ListJobsFilter) ([]DataJob, error) {
	where := []string{"deleted = false"}
	args := []any{}
	idx := 1

	if f.WorkspaceID > 0 {
		args = append(args, f.WorkspaceID)
		where = append(where, fmt.Sprintf("workspace_id = $%d", idx))
		idx++
	}
	if f.ProjectID != nil {
		args = append(args, *f.ProjectID)
		where = append(where, fmt.Sprintf("project_id = $%d", idx))
		idx++
	}
	if f.JobType != "" {
		args = append(args, f.JobType)
		where = append(where, fmt.Sprintf("job_type = $%d", idx))
		idx++
	}
	if f.Status != "" {
		args = append(args, f.Status)
		where = append(where, fmt.Sprintf("status = $%d", idx))
		idx++
	}

	limit := f.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	offset := f.Offset
	if offset < 0 {
		offset = 0
	}
	args = append(args, limit, offset)

	query := `
		SELECT id, tenant_id, workspace_id, project_id, job_type, name,
		       parameters, progress, status, error_message, scheduled_at,
		       executed_at, completed_at, duration_ms, triggered_by,
		       created_by, created_at, updated_at
		FROM data_jobs
		WHERE ` + strings.Join(where, " AND ") + `
		ORDER BY created_at DESC
		LIMIT $` + fmt.Sprint(idx) + ` OFFSET $` + fmt.Sprint(idx+1)

	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()

	out := make([]DataJob, 0)
	for rows.Next() {
		var j DataJob
		if err := rows.Scan(
			&j.ID, &j.TenantID, &j.WorkspaceID, &j.ProjectID,
			&j.JobType, &j.Name, &j.Parameters, &j.Progress, &j.Status,
			&j.ErrorMessage, &j.ScheduledAt, &j.ExecutedAt, &j.CompletedAt,
			&j.DurationMs, &j.TriggeredBy, &j.CreatedBy, &j.CreatedAt,
			&j.UpdatedAt,
		); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
		out = append(out, j)
	}
	if err := rows.Err(); err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	return out, nil
}

// UpdateProgress 更新任务进度（0-100）。
//
// 仅在任务处于 pending/running 状态时允许更新，终态任务拒绝变更。
func (s *Service) UpdateProgress(ctx context.Context, id int64, progress int) error {
	if progress < 0 || progress > 100 {
		return errs.ErrValidation.WithDetails(
			errs.FieldDetail{Field: "progress", Reason: "must be between 0 and 100"},
		)
	}
	tag, err := s.db.Exec(ctx, `
		UPDATE data_jobs
		SET progress = $2, updated_at = now()
		WHERE id = $1 AND deleted = false AND status IN ('pending', 'running')`,
		id, progress,
	)
	if err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	if tag.RowsAffected() == 0 {
		return errs.ErrNotFound
	}
	return nil
}

// MarkRunning 将任务标记为执行中（记录 executed_at）。
func (s *Service) MarkRunning(ctx context.Context, id int64) error {
	now := time.Now()
	tag, err := s.db.Exec(ctx, `
		UPDATE data_jobs
		SET status = 'running', executed_at = $2, updated_at = now()
		WHERE id = $1 AND deleted = false AND status = 'pending'`, id, now,
	)
	if err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	if tag.RowsAffected() == 0 {
		return errs.ErrNotFound
	}
	return nil
}

// MarkSuccess 将任务标记为成功（记录 completed_at + duration_ms）。
func (s *Service) MarkSuccess(ctx context.Context, id int64) error {
	tag, err := s.db.Exec(ctx, `
		UPDATE data_jobs
		SET status = 'success', progress = 100, completed_at = now(),
		    duration_ms = EXTRACT(EPOCH FROM (now() - COALESCE(executed_at, created_at))) * 1000,
		    updated_at = now()
		WHERE id = $1 AND deleted = false AND status IN ('pending', 'running')`, id,
	)
	if err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	if tag.RowsAffected() == 0 {
		return errs.ErrNotFound
	}
	return nil
}

// MarkFailed 将任务标记为失败（记录错误信息与完成时间）。
func (s *Service) MarkFailed(ctx context.Context, id int64, errMsg string) error {
	if strings.TrimSpace(errMsg) == "" {
		return errs.ErrValidation.WithDetails(
			errs.FieldDetail{Field: "error_message", Reason: "required when marking failed"},
		)
	}
	tag, err := s.db.Exec(ctx, `
		UPDATE data_jobs
		SET status = 'failed', error_message = $2, completed_at = now(),
		    duration_ms = EXTRACT(EPOCH FROM (now() - COALESCE(executed_at, created_at))) * 1000,
		    updated_at = now()
		WHERE id = $1 AND deleted = false AND status IN ('pending', 'running')`, id, errMsg,
	)
	if err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	if tag.RowsAffected() == 0 {
		return errs.ErrNotFound
	}
	return nil
}

// Cancel 取消任务（仅 pending/running 可取消）。
func (s *Service) Cancel(ctx context.Context, id, operatorID int64) error {
	tag, err := s.db.Exec(ctx, `
		UPDATE data_jobs
		SET status = 'cancelled', completed_at = now(), updated_by = $2, updated_at = now()
		WHERE id = $1 AND deleted = false AND status IN ('pending', 'running')`, id, operatorID,
	)
	if err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	if tag.RowsAffected() == 0 {
		return errs.ErrNotFound
	}
	return nil
}

// Delete 软删除任务。
func (s *Service) Delete(ctx context.Context, id, operatorID int64) error {
	tag, err := s.db.Exec(ctx, `
		UPDATE data_jobs
		SET deleted = true, updated_by = $2, updated_at = now()
		WHERE id = $1 AND deleted = false`, id, operatorID,
	)
	if err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	if tag.RowsAffected() == 0 {
		return errs.ErrNotFound
	}
	return nil
}

// --- Internals ---

func validateCreateInput(in CreateJobInput) error {
	if in.WorkspaceID <= 0 {
		return errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "workspace_id", Reason: "required"})
	}
	if strings.TrimSpace(in.JobType) == "" {
		return errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "job_type", Reason: "required"})
	}
	if strings.TrimSpace(in.Name) == "" {
		return errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "name", Reason: "required"})
	}
	return nil
}

func jsonOrEmpty(m map[string]any) any {
	if len(m) == 0 {
		return "{}"
	}
	return m
}
