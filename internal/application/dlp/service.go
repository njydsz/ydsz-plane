// Package dlp — 扫描服务与存储。
package dlp

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// Service 扫描服务。
type Service struct {
	db       *pgxpool.Pool
	detector *Detector
	// 内存任务状态（生产可迁移到 DataJob）
	mu   sync.RWMutex
	jobs map[int64]*Job
}

// Job 扫描任务。
type Job struct {
	ID           int64      `json:"id"`
	WorkspaceID  int64      `json:"workspace_id"`
	Scope        string     `json:"scope"` // project / page / issue / global
	TargetID     int64      `json:"target_id,omitempty"`
	Status       string     `json:"status"` // pending / running / completed / failed
	TotalItems   int        `json:"total_items"`
	ScannedItems int        `json:"scanned_items"`
	Findings     int        `json:"findings"`
	StartedAt    *time.Time `json:"started_at,omitempty"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
	ErrorMessage string     `json:"error_message,omitempty"`
	CreatedBy    int64      `json:"created_by"`
	CreatedAt    time.Time  `json:"created_at"`
}

// Finding 扫描发现。
type Finding struct {
	ID          int64      `json:"id"`
	JobID       int64      `json:"job_id"`
	WorkspaceID int64      `json:"workspace_id"`
	Scope       string     `json:"scope"`
	TargetID    int64      `json:"target_id"`
	TargetTitle string     `json:"target_title,omitempty"`
	Category    Category   `json:"category"`
	Severity    Severity   `json:"severity"`
	RuleName    string     `json:"rule_name"`
	MaskedValue string     `json:"masked_value"`
	Offset      int        `json:"offset"`
	Confidence  float64    `json:"confidence"`
	Action      string     `json:"action"` // flagged / blocked / resolved
	ActionedBy  int64      `json:"actioned_by,omitempty"`
	ResolvedAt  *time.Time `json:"resolved_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

// ScanReport 扫描报告。
type ScanReport struct {
	JobID       int64          `json:"job_id"`
	WorkspaceID int64          `json:"workspace_id"`
	TotalItems  int            `json:"total_items"`
	Findings    int            `json:"findings"`
	ByCategory  map[string]int `json:"by_category"`
	BySeverity  map[string]int `json:"by_severity"`
	CompletedAt time.Time      `json:"completed_at"`
}

// NewService 构造服务。
func NewService(db *pgxpool.Pool) *Service {
	return &Service{
		db:       db,
		detector: NewDetector(),
		jobs:     map[int64]*Job{},
	}
}

// Detector 返回识别引擎（供配置规则）。
func (s *Service) Detector() *Detector { return s.detector }

// --- 错误 ---

var (
	// ErrJobNotFound 任务不存在。
	ErrJobNotFound = errors.New("dlp job not found")
)

// CreateJob 创建扫描任务。
func (s *Service) CreateJob(workspaceID int64, scope string, targetID, createdBy int64) (*Job, error) {
	if scope == "" {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "scope", Reason: "required"})
	}
	job := &Job{
		ID:          genID(),
		WorkspaceID: workspaceID,
		Scope:       scope,
		TargetID:    targetID,
		Status:      "pending",
		CreatedBy:   createdBy,
		CreatedAt:   time.Now(),
	}
	s.mu.Lock()
	s.jobs[job.ID] = job
	s.mu.Unlock()
	return job, nil
}

// GetJob 查询任务。
func (s *Service) GetJob(id int64) (*Job, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	j, ok := s.jobs[id]
	if !ok {
		return nil, ErrJobNotFound
	}
	return j, nil
}

// ListJobs 列出工作空间任务。
func (s *Service) ListJobs(workspaceID int64) []*Job {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []*Job
	for _, j := range s.jobs {
		if j.WorkspaceID == workspaceID {
			out = append(out, j)
		}
	}
	return out
}

// RunScan 执行扫描（异步）。
func (s *Service) RunScan(ctx context.Context, jobID int64) {
	go s.run(ctx, jobID)
}

func (s *Service) run(ctx context.Context, jobID int64) {
	s.mu.Lock()
	job, ok := s.jobs[jobID]
	if !ok {
		s.mu.Unlock()
		return
	}
	job.Status = "running"
	now := time.Now()
	job.StartedAt = &now
	s.mu.Unlock()

	// 采集待扫描文本
	items, err := s.collectItems(ctx, job)
	if err != nil {
		s.failJob(jobID, err)
		return
	}

	s.mu.Lock()
	job.TotalItems = len(items)
	s.mu.Unlock()

	var totalFindings int
	for i, item := range items {
		if ctx.Err() != nil {
			s.failJob(jobID, ctx.Err())
			return
		}
		matches := s.detector.Scan(item.Text)
		for _, m := range matches {
			f := &Finding{
				ID:          genID(),
				JobID:       jobID,
				WorkspaceID: job.WorkspaceID,
				Scope:       job.Scope,
				TargetID:    item.TargetID,
				TargetTitle: item.Title,
				Category:    m.Category,
				Severity:    m.Severity,
				RuleName:    m.RuleName,
				MaskedValue: m.Mask(),
				Offset:      m.Offset,
				Confidence:  m.Confidence,
				Action:      "flagged",
				CreatedAt:   time.Now(),
			}
			if err := s.saveFinding(ctx, f); err != nil {
				s.failJob(jobID, err)
				return
			}
			totalFindings++
		}

		s.mu.Lock()
		job.ScannedItems = i + 1
		job.Findings = totalFindings
		s.mu.Unlock()
	}

	s.mu.Lock()
	job.Status = "completed"
	done := time.Now()
	job.CompletedAt = &done
	s.mu.Unlock()
}

func (s *Service) failJob(jobID int64, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if j, ok := s.jobs[jobID]; ok {
		j.Status = "failed"
		j.ErrorMessage = err.Error()
		done := time.Now()
		j.CompletedAt = &done
	}
}

// --- 扫描项采集 ---

type scanItem struct {
	TargetID int64
	Title    string
	Text     string
}

// collectItems 采集待扫描文本。
func (s *Service) collectItems(ctx context.Context, job *Job) ([]scanItem, error) {
	switch job.Scope {
	case "project":
		return s.collectProject(ctx, job.WorkspaceID, job.TargetID)
	case "page":
		return s.collectPage(ctx, job.WorkspaceID, job.TargetID)
	case "issue":
		return s.collectIssue(ctx, job.WorkspaceID, job.TargetID)
	case "global":
		return s.collectGlobal(ctx, job.WorkspaceID)
	default:
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "scope", Reason: "unsupported"})
	}
}

// collectProject 采集项目下的工作项与文档。
func (s *Service) collectProject(ctx context.Context, wsID, projectID int64) ([]scanItem, error) {
	var out []scanItem

	// 工作项（需求/任务/缺陷）
	tables := []string{"requirement", "task", "defect"}
	for _, t := range tables {
		rows, err := s.db.Query(ctx, `
			SELECT id, title, COALESCE(description,'')
			FROM `+t+`
			WHERE workspace_id = $1 AND project_id = $2 AND deleted = false
			LIMIT 5000`, wsID, projectID)
		if err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
		for rows.Next() {
			var id int64
			var title, desc string
			if err := rows.Scan(&id, &title, &desc); err != nil {
				rows.Close()
				return nil, errs.ErrInternal.Wrap(err)
			}
			out = append(out, scanItem{
				TargetID: id,
				Title:    title,
				Text:     title + "\n" + desc,
			})
		}
		rows.Close()
	}

	// 文档
	rows, err := s.db.Query(ctx, `
		SELECT id, title, COALESCE(description_html,'')
		FROM pages
		WHERE workspace_id = $1 AND project_id = $2 AND deleted = false
		LIMIT 2000`, wsID, projectID)
	if err == nil {
		for rows.Next() {
			var id int64
			var title, html string
			if err := rows.Scan(&id, &title, &html); err != nil {
				rows.Close()
				return nil, errs.ErrInternal.Wrap(err)
			}
			out = append(out, scanItem{
				TargetID: id,
				Title:    title,
				Text:     title + "\n" + stripTags(html),
			})
		}
		rows.Close()
	}
	return out, nil
}

// collectPage 采集单个文档。
func (s *Service) collectPage(ctx context.Context, wsID, pageID int64) ([]scanItem, error) {
	var id int64
	var title, html string
	err := s.db.QueryRow(ctx, `
		SELECT id, title, COALESCE(description_html,'')
		FROM pages WHERE id = $1 AND workspace_id = $2 AND deleted = false`,
		pageID, wsID).Scan(&id, &title, &html)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	return []scanItem{{TargetID: id, Title: title, Text: title + "\n" + stripTags(html)}}, nil
}

// collectIssue 采集单个工作项。
func (s *Service) collectIssue(ctx context.Context, wsID, issueID int64) ([]scanItem, error) {
	tables := []string{"requirement", "task", "defect"}
	for _, t := range tables {
		var id int64
		var title, desc string
		err := s.db.QueryRow(ctx, `
			SELECT id, title, COALESCE(description,'')
			FROM `+t+` WHERE id = $1 AND workspace_id = $2 AND deleted = false`,
			issueID, wsID).Scan(&id, &title, &desc)
		if err == nil {
			return []scanItem{{TargetID: id, Title: title, Text: title + "\n" + desc}}, nil
		}
	}
	return nil, nil
}

// collectGlobal 采集工作空间全部内容。
func (s *Service) collectGlobal(ctx context.Context, wsID int64) ([]scanItem, error) {
	var out []scanItem
	tables := []string{"requirement", "task", "defect"}
	for _, t := range tables {
		rows, err := s.db.Query(ctx, `
			SELECT id, title, COALESCE(description,'')
			FROM `+t+`
			WHERE workspace_id = $1 AND deleted = false
			LIMIT 5000`, wsID)
		if err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
		for rows.Next() {
			var id int64
			var title, desc string
			if err := rows.Scan(&id, &title, &desc); err != nil {
				rows.Close()
				return nil, errs.ErrInternal.Wrap(err)
			}
			out = append(out, scanItem{
				TargetID: id,
				Title:    title,
				Text:     title + "\n" + desc,
			})
		}
		rows.Close()
	}
	return out, nil
}

// --- 发现管理 ---

// saveFinding 保存发现。
func (s *Service) saveFinding(ctx context.Context, f *Finding) error {
	_, err := s.db.Exec(ctx, `
		INSERT INTO dlp_findings
			(id, job_id, workspace_id, scope, target_id, target_title, category,
			 severity, rule_name, masked_value, offset_pos, confidence, action)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,'flagged')
		ON CONFLICT DO NOTHING`,
		f.ID, f.JobID, f.WorkspaceID, f.Scope, f.TargetID, f.TargetTitle,
		string(f.Category), string(f.Severity), f.RuleName, f.MaskedValue,
		f.Offset, f.Confidence)
	if err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	return nil
}

// ListFindings 列出发现（可按类别/严重级别过滤）。
func (s *Service) ListFindings(ctx context.Context, workspaceID int64, category Category, severity Severity, action string, limit int) ([]*Finding, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	conds := []string{"workspace_id = $1"}
	args := []any{workspaceID}
	if category != "" {
		args = append(args, string(category))
		conds = append(conds, "category = $"+itoa(len(args)))
	}
	if severity != "" {
		args = append(args, string(severity))
		conds = append(conds, "severity = $"+itoa(len(args)))
	}
	if action != "" {
		args = append(args, action)
		conds = append(conds, "action = $"+itoa(len(args)))
	}
	args = append(args, limit)

	rows, err := s.db.Query(ctx, `
		SELECT id, job_id, workspace_id, scope, target_id, COALESCE(target_title,''),
			category, severity, rule_name, masked_value, offset_pos, confidence,
			action, actioned_by, resolved_at, created_at
		FROM dlp_findings
		WHERE `+joinComma(conds)+`
		ORDER BY created_at DESC
		LIMIT $`+itoa(len(args)), args...)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()

	var out []*Finding
	for rows.Next() {
		var f Finding
		if err := rows.Scan(&f.ID, &f.JobID, &f.WorkspaceID, &f.Scope, &f.TargetID,
			&f.TargetTitle, &f.Category, &f.Severity, &f.RuleName, &f.MaskedValue,
			&f.Offset, &f.Confidence, &f.Action, &f.ActionedBy, &f.ResolvedAt, &f.CreatedAt); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
		out = append(out, &f)
	}
	return out, rows.Err()
}

// ResolveFinding 标记发现为已处理。
func (s *Service) ResolveFinding(ctx context.Context, id, userID int64) error {
	tag, err := s.db.Exec(ctx, `
		UPDATE dlp_findings
		SET action = 'resolved', actioned_by = $1, resolved_at = now()
		WHERE id = $2 AND action = 'flagged'`, userID, id)
	if err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	if tag.RowsAffected() == 0 {
		return ErrJobNotFound
	}
	return nil
}

// GetReport 生成扫描报告。
func (s *Service) GetReport(ctx context.Context, jobID int64) (*ScanReport, error) {
	s.mu.RLock()
	job, ok := s.jobs[jobID]
	s.mu.RUnlock()
	if !ok {
		return nil, ErrJobNotFound
	}

	report := &ScanReport{
		JobID:       jobID,
		WorkspaceID: job.WorkspaceID,
		TotalItems:  job.TotalItems,
		Findings:    job.Findings,
		ByCategory:  map[string]int{},
		BySeverity:  map[string]int{},
		CompletedAt: job.CreatedAt,
	}
	if job.CompletedAt != nil {
		report.CompletedAt = *job.CompletedAt
	}

	rows, err := s.db.Query(ctx, `
		SELECT category, severity, COUNT(*)
		FROM dlp_findings WHERE job_id = $1
		GROUP BY category, severity`, jobID)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()
	for rows.Next() {
		var cat, sev string
		var n int
		if err := rows.Scan(&cat, &sev, &n); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
		report.ByCategory[cat] += n
		report.BySeverity[sev] += n
	}
	return report, rows.Err()
}

// --- 写入前检查（拦截/告警） ---

// CheckAction 检查结果对应的处理动作。
type CheckAction string

const (
	// CheckAllow 允许写入。
	CheckAllow CheckAction = "allow"
	// CheckWarn 允许但告警。
	CheckWarn CheckAction = "warn"
	// CheckBlock 阻断写入。
	CheckBlock CheckAction = "block"
)

// Policy 扫描策略。
type Policy struct {
	// 阻断级别（含以上）
	BlockThreshold Severity
	// 告警级别
	WarnThreshold Severity
}

// DefaultPolicy 默认策略：critical/high 阻断，medium 告警。
var DefaultPolicy = Policy{
	BlockThreshold: SeverityHigh,
	WarnThreshold:  SeverityMedium,
}

// Check 检查文本并返回动作与匹配结果。
func (s *Service) Check(text string, policy Policy) (CheckAction, []Match) {
	matches := s.detector.Scan(text)
	if len(matches) == 0 {
		return CheckAllow, nil
	}
	sev := HighestSeverity(matches)
	switch sev {
	case SeverityCritical, SeverityHigh:
		if policy.BlockThreshold == "" || rank(sev) >= rank(policy.BlockThreshold) {
			return CheckBlock, matches
		}
	case SeverityMedium:
		if rank(sev) >= rank(policy.WarnThreshold) {
			return CheckWarn, matches
		}
	}
	return CheckAllow, matches
}

func rank(s Severity) int {
	switch s {
	case SeverityCritical:
		return 4
	case SeverityHigh:
		return 3
	case SeverityMedium:
		return 2
	case SeverityLow:
		return 1
	}
	return 0
}

// --- helpers ---

func genID() int64 {
	var b [8]byte
	_, _ = rand.Read(b[:])
	randBits := int64(binary.LittleEndian.Uint64(b[:]) & 0x3FFFFF)
	return (time.Now().UnixMilli() << 22) | randBits
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

// stripTags 简单去除 HTML 标签。
func stripTags(s string) string {
	var out []rune
	inTag := false
	for _, r := range s {
		switch {
		case r == '<':
			inTag = true
		case r == '>':
			inTag = false
		case !inTag:
			out = append(out, r)
		}
	}
	return string(out)
}
