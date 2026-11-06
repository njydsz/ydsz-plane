// Package importer — 导入执行服务。
//
// 职责：
//   - 编排「拉取 → 校验 → 写入」流程；
//   - 维护导入任务状态（异步）；
//   - 处理父子关系二次绑定（先建父后建子）；
//   - 幂等（相同 external_id 跳过）。
package importer

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// Service 导入服务。
type Service struct {
	db *pgxpool.Pool
	// 任务状态（内存；生产可迁移到 DB/DataJob）
	mu   sync.RWMutex
	jobs map[int64]*Job
	next int64
}

// Job 导入任务。
type Job struct {
	ID          int64         `json:"id"`
	WorkspaceID int64         `json:"workspace_id"`
	ProjectID   int64         `json:"project_id"`
	Source      Source        `json:"source"`
	Status      string        `json:"status"` // pending / running / completed / failed
	Result      *ImportResult `json:"result,omitempty"`
	CreatedBy   int64         `json:"created_by"`
	CreatedAt   time.Time     `json:"created_at"`
	CompletedAt *time.Time    `json:"completed_at,omitempty"`
}

// NewService 构造服务。
func NewService(db *pgxpool.Pool) *Service {
	return &Service{db: db, jobs: map[int64]*Job{}}
}

// --- 错误 ---

var (
	// ErrJobNotFound 任务不存在。
	ErrJobNotFound = errors.New("import job not found")
	// ErrUnsupportedSource 不支持的来源。
	ErrUnsupportedSource = errors.New("unsupported import source")
)

// CreateJob 创建导入任务。
func (s *Service) CreateJob(workspaceID, projectID, createdBy int64, src Source) (*Job, error) {
	if !IsValidSource(string(src)) {
		return nil, ErrUnsupportedSource
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.next++
	job := &Job{
		ID:          genID(),
		WorkspaceID: workspaceID,
		ProjectID:   projectID,
		Source:      src,
		Status:      "pending",
		CreatedBy:   createdBy,
		CreatedAt:   time.Now(),
	}
	s.jobs[job.ID] = job
	return job, nil
}

// GetJob 查询任务。
func (s *Service) GetJob(id int64) (*Job, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	job, ok := s.jobs[id]
	if !ok {
		return nil, ErrJobNotFound
	}
	return job, nil
}

// ListJobs 列出项目任务。
func (s *Service) ListJobs(workspaceID, projectID int64) []*Job {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []*Job{}
	for _, j := range s.jobs {
		if j.WorkspaceID == workspaceID && j.ProjectID == projectID {
			out = append(out, j)
		}
	}
	return out
}

// RunCSV 执行 CSV 导入。
func (s *Service) RunCSV(ctx context.Context, jobID int64, data []byte, opts ImportOptions) (*ImportResult, error) {
	job, err := s.GetJob(jobID)
	if err != nil {
		return nil, err
	}
	s.setStatus(jobID, "running")

	imp := NewCSVImporter()
	records, err := imp.FetchFromReader(strings.NewReader(string(data)), opts)
	if err != nil {
		s.fail(jobID, err)
		return nil, err
	}

	result := s.importRecords(ctx, job, records)
	s.finish(jobID, result)
	return result, nil
}

// RunPlatform 执行平台导入（Jira/Linear/Asana）。
func (s *Service) RunPlatform(ctx context.Context, jobID int64, creds Credentials, opts ImportOptions) (*ImportResult, error) {
	job, err := s.GetJob(jobID)
	if err != nil {
		return nil, err
	}
	s.setStatus(jobID, "running")

	var imp Importer
	switch job.Source {
	case SourceJira:
		imp = NewJiraImporter()
	case SourceLinear:
		imp = NewLinearImporter()
	case SourceAsana:
		imp = NewAsanaImporter()
	default:
		s.fail(jobID, ErrUnsupportedSource)
		return nil, ErrUnsupportedSource
	}

	records, err := imp.Fetch(ctx, creds, opts)
	if err != nil {
		s.fail(jobID, err)
		return nil, err
	}
	result := s.importRecords(ctx, job, records)
	s.finish(jobID, result)
	return result, nil
}

// importRecords 将中间模型写入工作项。
//
// 策略：
//   - 先写父项（无 parent_external_id），建立 external_id → local_id 映射；
//   - 再写子项，绑定 parent_id；
//   - 相同 external_id 幂等跳过。
func (s *Service) importRecords(ctx context.Context, job *Job, records []WorkItemRecord) *ImportResult {
	result := &ImportResult{
		Source:       job.Source,
		StartedAt:    time.Now(),
		TotalRecords: len(records),
	}

	idMap := map[string]int64{}
	// 第一遍：父项
	for _, rec := range records {
		if rec.ParentExternalID != "" {
			continue
		}
		localID, err := s.upsertRecord(ctx, job, rec)
		if err != nil {
			result.Failed++
			result.Errors = append(result.Errors, ImportError{
				ExternalID: rec.ExternalID, Title: rec.Title, Reason: err.Error(),
			})
			continue
		}
		if localID > 0 {
			idMap[rec.ExternalID] = localID
			result.Imported++
		} else {
			result.Skipped++
		}
	}
	// 第二遍：子项
	for _, rec := range records {
		if rec.ParentExternalID == "" {
			continue
		}
		parentID, ok := idMap[rec.ParentExternalID]
		if !ok {
			result.Skipped++
			continue
		}
		localID, err := s.upsertRecord(ctx, job, rec)
		if err != nil {
			result.Failed++
			result.Errors = append(result.Errors, ImportError{
				ExternalID: rec.ExternalID, Title: rec.Title, Reason: err.Error(),
			})
			continue
		}
		if localID > 0 {
			idMap[rec.ExternalID] = localID
			// 绑定父子关系
			_, _ = s.db.Exec(ctx, `
				UPDATE `+workItemTable(rec.TypeCode)+`
				SET parent_id = $1, updated_at = now()
				WHERE id = $2`, parentID, localID)
			result.Imported++
		} else {
			result.Skipped++
		}
	}

	now := time.Now()
	result.CompletedAt = &now
	return result
}

// upsertRecord 幂等写入单条记录；返回 0 表示已存在（跳过）。
func (s *Service) upsertRecord(ctx context.Context, job *Job, rec WorkItemRecord) (int64, error) {
	table := workItemTable(rec.TypeCode)
	if table == "" {
		return 0, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "type_code", Reason: "unsupported type"})
	}

	// 幂等检查：按 project + 标题 + 创建人判重（无 external_id 列）
	var existing int64
	err := s.db.QueryRow(ctx, `
		SELECT id FROM `+table+`
		WHERE workspace_id = $1 AND project_id = $2 AND title = $3 AND deleted = false
		LIMIT 1`, job.WorkspaceID, job.ProjectID, rec.Title).Scan(&existing)
	if err == nil && existing > 0 {
		return 0, nil
	}
	if err != pgx.ErrNoRows {
		return 0, errs.ErrInternal.Wrap(err)
	}

	priority := rec.Priority
	if priority == "" {
		priority = "medium"
	}
	var due *time.Time
	if rec.DueDate != nil {
		due = rec.DueDate
	}

	var id int64
	err = s.db.QueryRow(ctx, `
		INSERT INTO `+table+` (id, workspace_id, project_id, title, description,
			priority, state_id, created_by, updated_by, target_date)
		VALUES ($1,$2,$3,$4,$5,$6,
			(SELECT id FROM states WHERE project_id = $3 AND "group" = 'backlog' ORDER BY sort_order LIMIT 1),
			$7,$7,$8)
		RETURNING id`, genID(), job.WorkspaceID, job.ProjectID, rec.Title,
		rec.Description, priority, job.CreatedBy, due).Scan(&id)
	if err != nil {
		return 0, errs.ErrInternal.Wrap(err)
	}
	return id, nil
}

// --- job 状态 ---

func (s *Service) setStatus(jobID int64, status string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if j, ok := s.jobs[jobID]; ok {
		j.Status = status
	}
}

func (s *Service) finish(jobID int64, result *ImportResult) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if j, ok := s.jobs[jobID]; ok {
		j.Status = "completed"
		j.Result = result
		now := time.Now()
		j.CompletedAt = &now
	}
}

func (s *Service) fail(jobID int64, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if j, ok := s.jobs[jobID]; ok {
		j.Status = "failed"
		now := time.Now()
		j.CompletedAt = &now
		j.Result = &ImportResult{
			StartedAt:   now,
			CompletedAt: &now,
			Errors:      []ImportError{{Reason: err.Error()}},
		}
	}
}

// --- helpers ---

func workItemTable(t string) string {
	switch strings.ToLower(t) {
	case "requirement", "req":
		return "requirement"
	case "task":
		return "task"
	case "defect", "bug":
		return "defect"
	}
	return ""
}

func genID() int64 {
	return time.Now().UnixMilli()<<22 | (time.Now().UnixNano() & 0x3FFFFF)
}

var _ = json.Marshal
