// Package calendar — 项目日历应用服务。
//
// 提供日历事件的 CRUD 与列表查询，支撑「项目日历」页面的独立事件管理
// （会议 / 里程碑 / 发布日 / 截止日），并与工作项时间排布叠加展示。
package calendar

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// Service 日历应用服务。
type Service struct {
	db *pgxpool.Pool
}

// NewService 创建日历服务。
func NewService(db *pgxpool.Pool) *Service {
	return &Service{db: db}
}

// --- CRUD ---

// Create 创建日历事件。
//
// 幂等策略：当 IdempotencyKey 非空且已存在同工作空间同键的活跃事件时，
// 直接返回既有事件（不重复创建），支撑外部系统重复上报场景。
func (s *Service) Create(ctx context.Context, in CreateEventInput) (*CalendarEvent, error) {
	if err := validateCreateInput(in); err != nil {
		return nil, err
	}

	// 幂等去重：同工作空间 + 同 idempotency_key 的活跃事件只保留一条。
	if in.IdempotencyKey != "" {
		existing, err := s.findByIdempotencyKey(ctx, in.WorkspaceID, in.IdempotencyKey)
		if err != nil {
			return nil, err
		}
		if existing != nil {
			return existing, nil
		}
	}

	id := genID()
	var event CalendarEvent
	err := s.db.QueryRow(ctx, `
		INSERT INTO calendar_events (
			id, tenant_id, workspace_id, project_id, title, description,
			start_time, end_time, is_all_day, location, event_type,
			source_type, source_id, idempotency_key, organizer_id,
			status, version, created_by, updated_by
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,'active',1,$16,$16)
		RETURNING id, tenant_id, workspace_id, project_id, title, description,
			start_time, end_time, is_all_day, location, event_type,
			source_type, source_id, idempotency_key, organizer_id,
			status, version, created_by, created_at, updated_at`,
		id, in.TenantID, in.WorkspaceID, in.ProjectID, in.Title, in.Description,
		in.StartTime, in.EndTime, in.IsAllDay, in.Location, in.EventType,
		nullIfEmpty(in.SourceType), in.SourceID, nullIfEmpty(in.IdempotencyKey),
		in.OrganizerID, in.CreatedBy,
	).Scan(
		&event.ID, &event.TenantID, &event.WorkspaceID, &event.ProjectID,
		&event.Title, &event.Description, &event.StartTime, &event.EndTime,
		&event.IsAllDay, &event.Location, &event.EventType, &event.SourceType,
		&event.SourceID, &event.IdempotencyKey, &event.OrganizerID,
		&event.Status, &event.Version, &event.CreatedBy, &event.CreatedAt,
		&event.UpdatedAt,
	)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	return &event, nil
}

// Update 更新日历事件（乐观锁：version 不匹配返回 ErrVersionConflict）。
func (s *Service) Update(ctx context.Context, in UpdateEventInput) (*CalendarEvent, error) {
	if in.ID <= 0 {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "id", Reason: "required"})
	}
	if strings.TrimSpace(in.Title) == "" {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "title", Reason: "required"})
	}
	if in.EndTime.Before(in.StartTime) {
		return nil, errs.ErrValidation.WithDetails(
			errs.FieldDetail{Field: "end_time", Reason: "must be after start_time"},
		)
	}
	if !in.EventType.IsValid() {
		return nil, errs.ErrValidation.WithDetails(
			errs.FieldDetail{Field: "event_type", Reason: "invalid value"},
		)
	}

	var event CalendarEvent
	err := s.db.QueryRow(ctx, `
		UPDATE calendar_events
		SET title = $1, description = $2, start_time = $3, end_time = $4,
		    is_all_day = $5, location = $6, event_type = $7, organizer_id = $8,
		    version = version + 1, updated_by = $9, updated_at = now()
		WHERE id = $10 AND version = $11 AND deleted = false
		RETURNING id, tenant_id, workspace_id, project_id, title, description,
			start_time, end_time, is_all_day, location, event_type,
			source_type, source_id, idempotency_key, organizer_id,
			status, version, created_by, created_at, updated_at`,
		in.Title, in.Description, in.StartTime, in.EndTime, in.IsAllDay,
		in.Location, in.EventType, in.OrganizerID, in.UpdatedBy,
		in.ID, in.Version,
	).Scan(
		&event.ID, &event.TenantID, &event.WorkspaceID, &event.ProjectID,
		&event.Title, &event.Description, &event.StartTime, &event.EndTime,
		&event.IsAllDay, &event.Location, &event.EventType, &event.SourceType,
		&event.SourceID, &event.IdempotencyKey, &event.OrganizerID,
		&event.Status, &event.Version, &event.CreatedBy, &event.CreatedAt,
		&event.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errs.ErrVersionConflict
		}
		return nil, errs.ErrInternal.Wrap(err)
	}
	return &event, nil
}

// GetByID 按 ID 查询事件（含已软删除的过滤）。
func (s *Service) GetByID(ctx context.Context, id int64) (*CalendarEvent, error) {
	var event CalendarEvent
	err := s.db.QueryRow(ctx, `
		SELECT id, tenant_id, workspace_id, project_id, title, description,
		       start_time, end_time, is_all_day, location, event_type,
		       source_type, source_id, idempotency_key, organizer_id,
		       status, version, created_by, created_at, updated_at
		FROM calendar_events
		WHERE id = $1 AND deleted = false`, id,
	).Scan(
		&event.ID, &event.TenantID, &event.WorkspaceID, &event.ProjectID,
		&event.Title, &event.Description, &event.StartTime, &event.EndTime,
		&event.IsAllDay, &event.Location, &event.EventType, &event.SourceType,
		&event.SourceID, &event.IdempotencyKey, &event.OrganizerID,
		&event.Status, &event.Version, &event.CreatedBy, &event.CreatedAt,
		&event.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errs.ErrNotFound
		}
		return nil, errs.ErrInternal.Wrap(err)
	}
	return &event, nil
}

// List 按过滤条件查询事件列表（默认只返回活跃状态）。
func (s *Service) List(ctx context.Context, f ListEventsFilter) ([]CalendarEvent, error) {
	where := []string{"deleted = false", "status = 'active'"}
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
	if !f.Start.IsZero() {
		args = append(args, f.Start)
		where = append(where, fmt.Sprintf("end_time >= $%d", idx))
		idx++
	}
	if !f.End.IsZero() {
		args = append(args, f.End)
		where = append(where, fmt.Sprintf("start_time <= $%d", idx))
		idx++
	}
	if f.EventType != "" {
		args = append(args, f.EventType)
		where = append(where, fmt.Sprintf("event_type = $%d", idx))
		idx++
	}
	if f.SourceType != "" {
		args = append(args, f.SourceType)
		where = append(where, fmt.Sprintf("source_type = $%d", idx))
		idx++
	}
	if f.SourceID != nil {
		args = append(args, *f.SourceID)
		where = append(where, fmt.Sprintf("source_id = $%d", idx))
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
		SELECT id, tenant_id, workspace_id, project_id, title, description,
		       start_time, end_time, is_all_day, location, event_type,
		       source_type, source_id, idempotency_key, organizer_id,
		       status, version, created_by, created_at, updated_at
		FROM calendar_events
		WHERE ` + strings.Join(where, " AND ") + `
		ORDER BY start_time ASC
		LIMIT $` + fmt.Sprint(idx) + ` OFFSET $` + fmt.Sprint(idx+1)

	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()

	out := make([]CalendarEvent, 0)
	for rows.Next() {
		var e CalendarEvent
		if err := rows.Scan(
			&e.ID, &e.TenantID, &e.WorkspaceID, &e.ProjectID,
			&e.Title, &e.Description, &e.StartTime, &e.EndTime,
			&e.IsAllDay, &e.Location, &e.EventType, &e.SourceType,
			&e.SourceID, &e.IdempotencyKey, &e.OrganizerID,
			&e.Status, &e.Version, &e.CreatedBy, &e.CreatedAt,
			&e.UpdatedAt,
		); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
		out = append(out, e)
	}
	if err := rows.Err(); err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	return out, nil
}

// Archive 归档事件（软删除语义：置为 archived，保留历史）。
func (s *Service) Archive(ctx context.Context, id, operatorID int64) error {
	if id <= 0 {
		return errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "id", Reason: "required"})
	}
	tag, err := s.db.Exec(ctx, `
		UPDATE calendar_events
		SET status = 'archived', deleted = true, updated_by = $2, updated_at = now()
		WHERE id = $1 AND deleted = false`, id, operatorID)
	if err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	if tag.RowsAffected() == 0 {
		return errs.ErrNotFound
	}
	return nil
}

// --- Internals ---

func validateCreateInput(in CreateEventInput) error {
	if strings.TrimSpace(in.Title) == "" {
		return errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "title", Reason: "required"})
	}
	if in.WorkspaceID <= 0 {
		return errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "workspace_id", Reason: "required"})
	}
	if in.StartTime.IsZero() || in.EndTime.IsZero() {
		return errs.ErrValidation.WithDetails(
			errs.FieldDetail{Field: "start_time/end_time", Reason: "required"},
		)
	}
	if in.EndTime.Before(in.StartTime) {
		return errs.ErrValidation.WithDetails(
			errs.FieldDetail{Field: "end_time", Reason: "must be after start_time"},
		)
	}
	if !in.EventType.IsValid() {
		return errs.ErrValidation.WithDetails(
			errs.FieldDetail{Field: "event_type", Reason: "invalid value"},
		)
	}
	return nil
}

func (s *Service) findByIdempotencyKey(ctx context.Context, wsID int64, key string) (*CalendarEvent, error) {
	var e CalendarEvent
	err := s.db.QueryRow(ctx, `
		SELECT id, tenant_id, workspace_id, project_id, title, description,
		       start_time, end_time, is_all_day, location, event_type,
		       source_type, source_id, idempotency_key, organizer_id,
		       status, version, created_by, created_at, updated_at
		FROM calendar_events
		WHERE workspace_id = $1 AND idempotency_key = $2
		  AND deleted = false AND status = 'active'
		ORDER BY id DESC LIMIT 1`, wsID, key,
	).Scan(
		&e.ID, &e.TenantID, &e.WorkspaceID, &e.ProjectID,
		&e.Title, &e.Description, &e.StartTime, &e.EndTime,
		&e.IsAllDay, &e.Location, &e.EventType, &e.SourceType,
		&e.SourceID, &e.IdempotencyKey, &e.OrganizerID,
		&e.Status, &e.Version, &e.CreatedBy, &e.CreatedAt,
		&e.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, errs.ErrInternal.Wrap(err)
	}
	return &e, nil
}

func nullIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}

// Ensure time import is used (for potential future duration helpers).
var _ = time.Time{}
