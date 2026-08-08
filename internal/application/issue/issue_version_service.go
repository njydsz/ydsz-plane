// Package issue — Issue 版本快照审计（IssueVersion 领域）。
//
// 对标 Plane 的 IssueVersion：每次工作项字段变更时保存旧值快照，
// 支持回溯历史与字段级 diff。由 Update 流程触发（应用层，非 DB 触发器）。
package issue

import (
	"context"
	"encoding/json"
	"reflect"
	"sort"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// IssueVersion 单条版本快照（与 issue_versions 表一一对应）。
type IssueVersion struct {
	ID            int64           `json:"id"`
	WorkspaceID   int64           `json:"workspace_id"`
	ProjectID     int64           `json:"project_id"`
	IssueID       int64           `json:"issue_id"`
	Version       int             `json:"version"`
	Snapshot      json.RawMessage `json:"snapshot"`
	ChangedFields []string        `json:"changed_fields"`
	ChangeType    string          `json:"change_type"`
	CreatedBy     *int64          `json:"created_by,omitempty"`
	CreatedAt     time.Time       `json:"created_at"`
}

// VersionSnapshotInput 保存快照的入参。
type VersionSnapshotInput struct {
	WorkspaceID   int64
	ProjectID     int64
	IssueID       int64
	Version       int
	Snapshot      interface{}
	ChangedFields []string
	ChangeType    string
	CreatedBy     *int64
}

// VersionService 版本快照服务。
type VersionService struct {
	db *pgxpool.Pool
}

// NewVersionService 构造版本服务。
func NewVersionService(db *pgxpool.Pool) *VersionService {
	return &VersionService{db: db}
}

// SaveSnapshot 写入单条版本快照。
func (s *VersionService) SaveSnapshot(ctx context.Context, in VersionSnapshotInput) (*IssueVersion, error) {
	data, err := json.Marshal(in.Snapshot)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}

	var v IssueVersion
	err = s.db.QueryRow(ctx, `
		INSERT INTO issue_versions (workspace_id, project_id, issue_id, version, snapshot, changed_fields, change_type, created_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		RETURNING id, created_at`,
		in.WorkspaceID, in.ProjectID, in.IssueID, in.Version,
		data, in.ChangedFields, in.ChangeType, in.CreatedBy,
	).Scan(&v.ID, &v.CreatedAt)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}

	v.WorkspaceID = in.WorkspaceID
	v.ProjectID = in.ProjectID
	v.IssueID = in.IssueID
	v.Version = in.Version
	v.Snapshot = data
	v.ChangedFields = in.ChangedFields
	v.ChangeType = in.ChangeType
	v.CreatedBy = in.CreatedBy
	return &v, nil
}

// ListVersions 列出版本历史（倒序）。
func (s *VersionService) ListVersions(ctx context.Context, wsID int64, issueID int64, limit int) ([]IssueVersion, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	rows, err := s.db.Query(ctx, `
		SELECT id, workspace_id, project_id, issue_id, version, snapshot,
		       changed_fields, change_type, created_by, created_at
		FROM issue_versions
		WHERE workspace_id = $1 AND issue_id = $2
		ORDER BY version DESC
		LIMIT $3`, wsID, issueID, limit)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()

	var out []IssueVersion
	for rows.Next() {
		var v IssueVersion
		if err := rows.Scan(&v.ID, &v.WorkspaceID, &v.ProjectID, &v.IssueID, &v.Version,
			&v.Snapshot, &v.ChangedFields, &v.ChangeType, &v.CreatedBy, &v.CreatedAt); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

// GetVersion 获取单条版本快照。
func (s *VersionService) GetVersion(ctx context.Context, wsID int64, issueID int64, version int) (*IssueVersion, error) {
	var v IssueVersion
	err := s.db.QueryRow(ctx, `
		SELECT id, workspace_id, project_id, issue_id, version, snapshot,
		       changed_fields, change_type, created_by, created_at
		FROM issue_versions
		WHERE workspace_id = $1 AND issue_id = $2 AND version = $3`,
		wsID, issueID, version,
	).Scan(&v.ID, &v.WorkspaceID, &v.ProjectID, &v.IssueID, &v.Version,
		&v.Snapshot, &v.ChangedFields, &v.ChangeType, &v.CreatedBy, &v.CreatedAt)
	if err != nil {
		return nil, errs.ErrNotFound
	}
	return &v, nil
}

// DiffVersions 计算两个版本之间的字段级差异。
func (s *VersionService) DiffVersions(ctx context.Context, wsID int64, issueID int64, fromVer int, toVer int) (*VersionDiff, error) {
	from, err := s.GetVersion(ctx, wsID, issueID, fromVer)
	if err != nil {
		return nil, errs.ErrNotFound.WithDetails(errs.FieldDetail{Field: "from", Reason: "版本不存在"})
	}
	to, err := s.GetVersion(ctx, wsID, issueID, toVer)
	if err != nil {
		return nil, errs.ErrNotFound.WithDetails(errs.FieldDetail{Field: "to", Reason: "版本不存在"})
	}

	return computeDiff(from, to), nil
}

// VersionDiff 字段级差异结果。
type VersionDiff struct {
	FromVersion int                `json:"from_version"`
	ToVersion   int                `json:"to_version"`
	Changes     []FieldChange      `json:"changes"`
}

// FieldChange 单条字段差异。
type FieldChange struct {
	Field    string `json:"field"`
	OldValue any    `json:"old_value"`
	NewValue any    `json:"new_value"`
}

// computeDiff 计算两个 JSONB 快照之间的字段级差异。
func computeDiff(from, to *IssueVersion) *VersionDiff {
	var fromMap, toMap map[string]any
	_ = json.Unmarshal(from.Snapshot, &fromMap)
	_ = json.Unmarshal(to.Snapshot, &toMap)

	diff := &VersionDiff{FromVersion: from.Version, ToVersion: to.Version}
	allKeys := mergeKeys(fromMap, toMap)
	for _, k := range allKeys {
		oldV, oldOk := fromMap[k]
		newV, newOk := toMap[k]
		if !oldOk || !newOk {
			continue
		}
		if !reflect.DeepEqual(oldV, newV) {
			diff.Changes = append(diff.Changes, FieldChange{
				Field: k, OldValue: oldV, NewValue: newV,
			})
		}
	}
	return diff
}

func mergeKeys(a, b map[string]any) []string {
	seen := make(map[string]bool)
	for k := range a {
		seen[k] = true
	}
	for k := range b {
		seen[k] = true
	}
	keys := make([]string, 0, len(seen))
	for k := range seen {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

