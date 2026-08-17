// Package issue — Label 标签管理服务。
//
// 标签是项目级轻量分类维度（对标 Plane / Linear 的 Label）：
// 需求/任务/缺陷可打多个标签（M2M 经 issue_labels 关联），
// 标签本身独立维护（名称/颜色/描述），支持软删除。
package issue

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// LabelService 标签管理服务。
type LabelService struct {
	db *pgxpool.Pool
}

// NewLabelService 创建标签服务。
func NewLabelService(db *pgxpool.Pool) *LabelService {
	return &LabelService{db: db}
}

// CreateLabelInput 创建标签入参。
type CreateLabelInput struct {
	WorkspaceID int64
	ProjectID   int64
	Name        string
	Color       string
	Description string
	CreatedBy   int64
}

// UpdateLabelInput 更新标签入参。
type UpdateLabelInput struct {
	ID          int64
	WorkspaceID int64
	ProjectID   int64
	Name        *string
	Color       *string
	Description *string
}

// ListLabelsFilter 标签列表筛选。
type ListLabelsFilter struct {
	WorkspaceID int64
	ProjectID   int64
	Status      string // active|archived；空=全部
}

// CreateLabel 创建标签。
func (s *LabelService) CreateLabel(ctx context.Context, in CreateLabelInput) (*Label, error) {
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "name", Reason: "标签名称不能为空"})
	}
	if len([]rune(name)) > 120 {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "name", Reason: "标签名称不能超过 120 字符"})
	}
	// 同项目同名唯一
	var exists int
	if err := s.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM labels WHERE workspace_id=$1 AND project_id=$2 AND name=$3 AND deleted=false`,
		in.WorkspaceID, in.ProjectID, name).Scan(&exists); err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	if exists > 0 {
		return nil, errs.Conflict("LABEL.NAME_TAKEN", "该标签名称已存在")
	}

	var l Label
	err := s.db.QueryRow(ctx, `
		INSERT INTO labels (id, workspace_id, project_id, name, color, description, status, created_by)
		VALUES ($1,$2,$3,$4,$5,$6,'active',$7)
		RETURNING id, created_at, updated_at`,
		genLabelID(), in.WorkspaceID, in.ProjectID, name,
		nullLabelValue(in.Color), nullLabelValue(in.Description), in.CreatedBy,
	).Scan(&l.ID, &l.CreatedAt, &l.UpdatedAt)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	l.Name = name
	l.WorkspaceID = in.WorkspaceID
	l.ProjectID = in.ProjectID
	l.Color = in.Color
	l.Description = in.Description
	l.Status = "active"
	l.CreatedBy = in.CreatedBy
	return &l, nil
}

// GetLabel 获取单个标签。
func (s *LabelService) GetLabel(ctx context.Context, wsID, labelID int64) (*Label, error) {
	var l Label
	err := s.db.QueryRow(ctx, `
		SELECT id, coalesce(code,''), name, workspace_id, project_id,
		       coalesce(color,''), coalesce(description,''), status, created_by, created_at, updated_at
		FROM labels WHERE id=$1 AND workspace_id=$2 AND deleted=false`,
		labelID, wsID,
	).Scan(&l.ID, &l.Code, &l.Name, &l.WorkspaceID, &l.ProjectID,
		&l.Color, &l.Description, &l.Status, &l.CreatedBy, &l.CreatedAt, &l.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errs.ErrNotFound.From()
		}
		return nil, errs.ErrInternal.Wrap(err)
	}
	return &l, nil
}

// ListLabels 列出项目标签（含使用计数）。
func (s *LabelService) ListLabels(ctx context.Context, f ListLabelsFilter) ([]Label, error) {
	query := `
		SELECT l.id, coalesce(l.code,''), l.name, l.workspace_id, l.project_id,
		       coalesce(l.color,''), coalesce(l.description,''), l.status, l.created_by,
		       l.created_at, l.updated_at,
		       ((SELECT COUNT(*) FROM requirement_labels rl WHERE rl.label_id=l.id AND rl.deleted=false)
		         + (SELECT COUNT(*) FROM task_labels tl WHERE tl.label_id=l.id AND tl.deleted=false)
		         + (SELECT COUNT(*) FROM defect_labels dl WHERE dl.label_id=l.id AND dl.deleted=false)) AS issue_count
		FROM labels l
		WHERE l.workspace_id=$1 AND l.project_id=$2 AND l.deleted=false`
	args := []any{f.WorkspaceID, f.ProjectID}
	if f.Status != "" {
		query += ` AND l.status=$3`
		args = append(args, f.Status)
	}
	query += ` ORDER BY l.created_at DESC`

	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()

	labels := make([]Label, 0)
	for rows.Next() {
		var l Label
		if err := rows.Scan(&l.ID, &l.Code, &l.Name, &l.WorkspaceID, &l.ProjectID,
			&l.Color, &l.Description, &l.Status, &l.CreatedBy,
			&l.CreatedAt, &l.UpdatedAt, &l.IssueCount); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
		labels = append(labels, l)
	}
	return labels, nil
}

// UpdateLabel 更新标签。
func (s *LabelService) UpdateLabel(ctx context.Context, in UpdateLabelInput) (*Label, error) {
	cur, err := s.GetLabel(ctx, in.WorkspaceID, in.ID)
	if err != nil {
		return nil, err
	}
	if in.Name != nil {
		name := strings.TrimSpace(*in.Name)
		if name == "" {
			return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "name", Reason: "标签名称不能为空"})
		}
		var exists int
		if err := s.db.QueryRow(ctx,
			`SELECT COUNT(*) FROM labels WHERE workspace_id=$1 AND project_id=$2 AND name=$3 AND id<>$4 AND deleted=false`,
			in.WorkspaceID, in.ProjectID, name, in.ID).Scan(&exists); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
		if exists > 0 {
			return nil, errs.Conflict("LABEL.NAME_TAKEN", "该标签名称已存在")
		}
		cur.Name = name
	}
	if in.Color != nil {
		cur.Color = *in.Color
	}
	if in.Description != nil {
		cur.Description = *in.Description
	}

	_, err = s.db.Exec(ctx, `
		UPDATE labels SET name=$1, color=$2, description=$3, updated_at=now()
		WHERE id=$4 AND workspace_id=$5 AND deleted=false`,
		cur.Name, nullLabelValue(cur.Color), nullLabelValue(cur.Description), in.ID, in.WorkspaceID)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	return s.GetLabel(ctx, in.WorkspaceID, in.ID)
}

// DeleteLabel 软删除标签。
func (s *LabelService) DeleteLabel(ctx context.Context, wsID, projectID, labelID int64) error {
	tag, err := s.db.Exec(ctx,
		`UPDATE labels SET deleted=true, updated_at=now() WHERE id=$1 AND workspace_id=$2 AND project_id=$3 AND deleted=false`,
		labelID, wsID, projectID)
	if err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	if tag.RowsAffected() == 0 {
		return errs.ErrNotFound.From()
	}
	return nil
}

// ---- helpers ----

func nullLabelValue(s string) any {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return s
}

// genLabelID 生成应用层 BIGINT 主键（labels 表 id 无数据库默认值）。
func genLabelID() int64 {
	var b [8]byte
	_, _ = rand.Read(b[:])
	randBits := int64(binary.LittleEndian.Uint64(b[:]) & 0x3FFFFF)
	return (time.Now().UnixMilli() << 22) | randBits
}
