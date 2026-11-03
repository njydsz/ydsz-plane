// Package datajob — 异步数据任务域。
//
// 对标 Jira/Plane 的后台任务能力：大文件导入导出、数据同步、报表生成等
// 长耗时操作通过异步任务执行，用户可在「数据任务」页面查看进度与结果。
//
// 数据表：data_jobs（id 由应用层生成）。
//
// 设计要点：
//   - 状态机：pending → running → success / failed / cancelled；
//   - 进度可观测：progress 0-100，前端可渲染进度条；
//   - 错误可追溯：error_message 记录失败原因；
//   - 幂等：同工作空间同名任务不允许重复创建（部分唯一索引约束）。
package datajob

import (
	"crypto/rand"
	"encoding/binary"
	"time"
)

// JobStatus 任务状态（对应 data_job_status 枚举）。
type JobStatus string

const (
	// StatusPending 待执行。
	StatusPending JobStatus = "pending"
	// StatusRunning 执行中。
	StatusRunning JobStatus = "running"
	// StatusSuccess 成功。
	StatusSuccess JobStatus = "success"
	// StatusFailed 失败。
	StatusFailed JobStatus = "failed"
	// StatusCancelled 已取消。
	StatusCancelled JobStatus = "cancelled"
)

// IsValid 校验状态是否合法。
func (s JobStatus) IsValid() bool {
	switch s {
	case StatusPending, StatusRunning, StatusSuccess, StatusFailed, StatusCancelled:
		return true
	default:
		return false
	}
}

// IsTerminal 报告状态是否为终态（不再变化）。
func (s JobStatus) IsTerminal() bool {
	switch s {
	case StatusSuccess, StatusFailed, StatusCancelled:
		return true
	default:
		return false
	}
}

// JobType 任务类型常量。
const (
	// JobTypeImport 数据导入（CSV/Excel）。
	JobTypeImport = "import"
	// JobTypeExport 数据导出。
	JobTypeExport = "export"
	// JobTypeSync 数据同步。
	JobTypeSync = "sync"
	// JobTypeReport 报表生成。
	JobTypeReport = "report"
	// JobTypeReindex 搜索索引重建。
	JobTypeReindex = "reindex"
)

// DataJob 异步任务实体。
type DataJob struct {
	ID           int64      `json:"id"`
	TenantID     int64      `json:"tenant_id"`
	WorkspaceID  int64      `json:"workspace_id"`
	ProjectID    *int64     `json:"project_id,omitempty"`
	JobType      string     `json:"job_type"`
	Name         string     `json:"name"`
	Parameters   map[string]any `json:"parameters,omitempty"`
	Progress     int        `json:"progress"`
	Status       JobStatus  `json:"status"`
	ErrorMessage string     `json:"error_message,omitempty"`
	ScheduledAt  *time.Time `json:"scheduled_at,omitempty"`
	ExecutedAt   *time.Time `json:"executed_at,omitempty"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
	DurationMs   *int64     `json:"duration_ms,omitempty"`
	TriggeredBy  *int64     `json:"triggered_by,omitempty"`
	CreatedBy    int64      `json:"created_by"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// CreateJobInput 创建任务入参。
type CreateJobInput struct {
	TenantID    int64
	WorkspaceID int64
	ProjectID   *int64
	JobType     string
	Name        string
	Parameters  map[string]any
	ScheduledAt *time.Time
	TriggeredBy *int64
	CreatedBy   int64
}

// ListJobsFilter 任务列表过滤条件。
type ListJobsFilter struct {
	WorkspaceID int64
	ProjectID   *int64
	JobType     string
	Status      JobStatus
	Limit       int
	Offset      int
}

// --- ID 生成 ---

// genID 生成基于时间戳的伪随机 ID（与 calendar/intake 包策略一致）。
//
// 结构：高 42 位为毫秒时间戳，低 22 位为密码学随机数。
func genID() int64 {
	var b [8]byte
	_, _ = rand.Read(b[:])
	randBits := int64(binary.LittleEndian.Uint64(b[:]) & 0x3FFFFF)
	return (time.Now().UnixMilli() << 22) | randBits
}
