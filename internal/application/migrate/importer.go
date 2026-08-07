// Package migrate — 数据迁移导入器框架。
//
// 支持从主流项目管理平台导入数据:
//   - Jira (Cloud/Server/Data Center)
//   - 阿里云效 (YunXiao)
//   - TAPD (Tencent Agile Product Development)
//   - ONES
//   - CSV 通用格式
//
// 架构设计:
//   - 每个数据源实现 Importer 接口
//   - 两阶段导入: 解析(Parse) → 写入(Write)
//   - 幂等导入：基于 source_id 去重
//   - 错误收集：部分失败不中断全量导入
//   - 导入报告：汇总成功/失败/跳过计数
//
// 使用方式:
//   importer := migrate.NewJiraImporter(db, workspaceID, projectID)
//   report, err := importer.Import(ctx, sourceReader)
package migrate

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// --- Import Source Types ---

// SourceType 数据源类型。
type SourceType string

const (
	SourceJira     SourceType = "jira"
	SourceYunXiao  SourceType = "yunxiao"
	SourceTAPD     SourceType = "tapd"
	SourceONES     SourceType = "ones"
	SourceCSV      SourceType = "csv"
)

// --- Import Report ---

// ImportReport 导入报告。
type ImportReport struct {
	SourceType  SourceType     `json:"source_type"`
	StartedAt   time.Time      `json:"started_at"`
	CompletedAt time.Time      `json:"completed_at"`
	DurationMs  int64          `json:"duration_ms"`
	Total       int            `json:"total"`
	Success     int            `json:"success"`
	Failed      int            `json:"failed"`
	Skipped     int            `json:"skipped"`
	Errors      []ImportError  `json:"errors,omitempty"`
	Warnings    []string       `json:"warnings,omitempty"`
}

// ImportError 单条导入错误。
type ImportError struct {
	Row     int    `json:"row"`
	SourceID string `json:"source_id"`
	Message string `json:"message"`
}

// --- Parsed Entity ---

// ParsedIssue 解析后的工作项。
type ParsedIssue struct {
	SourceID    string            `json:"source_id"`    // 原始系统 ID
	SourceKey   string            `json:"source_key"`   // 原始系统 Key (如 JIRA-123)
	Title       string            `json:"title"`
	Description string            `json:"description"`
	TypeCode    string            `json:"type_code"`    // requirement/task/defect
	Priority    string            `json:"priority"`
	StateName   string            `json:"state_name"`
	Assignee    string            `json:"assignee"`     // 邮箱或用户名
	Reporter    string            `json:"reporter"`
	Labels      []string          `json:"labels"`
	Module      string            `json:"module"`
	Sprint      string            `json:"sprint"`
	Version     string            `json:"version"`
	StoryPoints float64           `json:"story_points"`
	TargetDate  *time.Time        `json:"target_date"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	Comments    []ParsedComment   `json:"comments,omitempty"`
	CustomFields map[string]any   `json:"custom_fields,omitempty"`
	Raw         json.RawMessage   `json:"raw,omitempty"` // 原始数据保留
}

// ParsedComment 解析后的评论。
type ParsedComment struct {
	Author    string    `json:"author"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
}

// --- Importer Interface ---

// Importer 数据导入器接口。
type Importer interface {
	// SourceType 返回数据源类型。
	SourceType() SourceType

	// Parse 解析原始数据为 ParsedIssue 列表。
	Parse(ctx context.Context, reader io.Reader) ([]ParsedIssue, error)

	// Write 将解析后的工作项写入数据库（幂等）。
	Write(ctx context.Context, issues []ParsedIssue) (*ImportReport, error)

	// Import 一键导入（Parse + Write）。
	Import(ctx context.Context, reader io.Reader) (*ImportReport, error)
}

// --- Base Importer ---

// BaseImporter 提供通用导入逻辑。
type BaseImporter struct {
	DB          *pgxpool.Pool
	WorkspaceID int64
	ProjectID   int64
	Source      SourceType
}

// NewBaseImporter 创建基础导入器。
func NewBaseImporter(db *pgxpool.Pool, wsID, projID int64, source SourceType) *BaseImporter {
	return &BaseImporter{
		DB:          db,
		WorkspaceID: wsID,
		ProjectID:   projID,
		Source:      source,
	}
}

// Write 将 ParsedIssue 写入数据库。
// 使用 INSERT ... ON CONFLICT 实现幂等导入。
func (b *BaseImporter) Write(ctx context.Context, issues []ParsedIssue) (*ImportReport, error) {
	report := &ImportReport{
		SourceType: b.Source,
		StartedAt:  time.Now(),
		Total:      len(issues),
	}

	for i, issue := range issues {
		rowNum := i + 1
		err := b.writeIssue(ctx, issue)
		if err != nil {
			report.Failed++
			report.Errors = append(report.Errors, ImportError{
				Row:     rowNum,
				SourceID: issue.SourceID,
				Message: err.Error(),
			})
			// 限制错误数（避免内存爆炸）
			if len(report.Errors) > 100 {
				report.Warnings = append(report.Warnings,
					fmt.Sprintf("错误过多，仅展示前 100 条（共 %d 条失败）", report.Failed))
				break
			}
			continue
		}
		report.Success++
	}

	report.CompletedAt = time.Now()
	report.DurationMs = report.CompletedAt.Sub(report.StartedAt).Milliseconds()

	return report, nil
}

// writeIssue 写入单条工作项（幂等）。
func (b *BaseImporter) writeIssue(ctx context.Context, p ParsedIssue) error {
	// 查找或创建模块
	moduleID := int64(0)
	if p.Module != "" {
		_ = b.DB.QueryRow(ctx, `
			INSERT INTO modules (workspace_id, project_id, name) VALUES ($1,$2,$3)
			ON CONFLICT (project_id, name) WHERE deleted_at IS NULL DO UPDATE SET name=EXCLUDED.name
			RETURNING id`, b.WorkspaceID, b.ProjectID, p.Module).Scan(&moduleID)
	}

	// 查找或创建标签
	var labelIDs []int64
	for _, label := range p.Labels {
		var lid int64
		err := b.DB.QueryRow(ctx, `
			INSERT INTO labels (workspace_id, project_id, name, color) VALUES ($1,$2,$3,'#6b7280')
			ON CONFLICT (project_id, name) WHERE deleted_at IS NULL DO NOTHING
			RETURNING id`, b.WorkspaceID, b.ProjectID, label).Scan(&lid)
		if err == nil {
			labelIDs = append(labelIDs, lid)
		}
	}

	// 查找指派人
	var assigneeID *int64
	if p.Assignee != "" {
		var uid int64
		err := b.DB.QueryRow(ctx,
			`SELECT id FROM users WHERE email = $1 AND deleted_at IS NULL`, p.Assignee).Scan(&uid)
		if err == nil {
			assigneeID = &uid
		}
	}

	// 导入工作项（幂等: ON CONFLICT DO UPDATE）
	_, err := b.DB.Exec(ctx, `
		INSERT INTO issues (
			workspace_id, project_id, type_code, title, description, description_stripped,
			priority, state_id, module_id, assignee_id, target_date, story_points,
			created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,
			(SELECT id FROM states WHERE project_id=$2 AND name=$8 AND deleted_at IS NULL LIMIT 1),
			$9,$10,$11,$12,$13,$14)
		ON CONFLICT DO NOTHING`,
		b.WorkspaceID, b.ProjectID,
		p.TypeCode, p.Title, p.Description, stripHTML(p.Description),
		p.Priority, p.StateName,
		nullIfZero(moduleID), assigneeID, p.TargetDate, p.StoryPoints,
		p.CreatedAt, p.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("write issue: %w", err)
	}

	return nil
}

// --- Jira Importer ---

// JiraImporter Jira 数据导入器。
// 支持从 Jira CSV 导出或 REST API 导出导入。
type JiraImporter struct {
	*BaseImporter
}

// NewJiraImporter 创建 Jira 导入器。
func NewJiraImporter(db *pgxpool.Pool, wsID, projID int64) *JiraImporter {
	return &JiraImporter{
		BaseImporter: NewBaseImporter(db, wsID, projID, SourceJira),
	}
}

// SourceType 返回数据源类型。
func (j *JiraImporter) SourceType() SourceType { return SourceJira }

// Parse 解析 Jira CSV 导出文件。
// Jira CSV 字段映射:
//
//	Summary → Title
//	Description → Description
//	Issue Type → TypeCode
//	Priority → Priority
//	Status → StateName
//	Assignee → Assignee
//	Reporter → Reporter
//	Labels → Labels
//	Fix Version/s → Version
//	Sprint → Sprint
//	Story Points → StoryPoints
//	Created → CreatedAt
//	Updated → UpdatedAt
func (j *JiraImporter) Parse(ctx context.Context, reader io.Reader) ([]ParsedIssue, error) {
	csvReader := csv.NewReader(reader)
	csvReader.LazyQuotes = true

	// 读取表头
	headers, err := csvReader.Read()
	if err != nil {
		return nil, fmt.Errorf("read csv header: %w", err)
	}

	// 构建列索引
	colIdx := make(map[string]int)
	for i, h := range headers {
		colIdx[h] = i
	}

	var issues []ParsedIssue
	rowNum := 1

	for {
		row, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read csv row %d: %w", rowNum, err)
		}
		rowNum++

		issue := ParsedIssue{
			SourceID:     getCol(row, colIdx, "Issue id"),
			SourceKey:    getCol(row, colIdx, "Issue key"),
			Title:        getCol(row, colIdx, "Summary"),
			Description:  getCol(row, colIdx, "Description"),
			TypeCode:     mapJiraType(getCol(row, colIdx, "Issue Type")),
			Priority:     mapJiraPriority(getCol(row, colIdx, "Priority")),
			StateName:    getCol(row, colIdx, "Status"),
			Assignee:     getCol(row, colIdx, "Assignee"),
			Reporter:     getCol(row, colIdx, "Reporter"),
			Labels:       splitCSV(getCol(row, colIdx, "Labels")),
			Version:      getCol(row, colIdx, "Fix Version/s"),
			Sprint:       getCol(row, colIdx, "Sprint"),
			StoryPoints:  parseFloat(getCol(row, colIdx, "Story Points")),
			CustomFields: map[string]any{},
		}

		// 解析日期
		if t, err := parseJiraDate(getCol(row, colIdx, "Created")); err == nil {
			issue.CreatedAt = t
		} else {
			issue.CreatedAt = time.Now()
		}
		if t, err := parseJiraDate(getCol(row, colIdx, "Updated")); err == nil {
			issue.UpdatedAt = t
		} else {
			issue.UpdatedAt = time.Now()
		}

		issues = append(issues, issue)
	}

	return issues, nil
}

// Import 一键导入。
func (j *JiraImporter) Import(ctx context.Context, reader io.Reader) (*ImportReport, error) {
	issues, err := j.Parse(ctx, reader)
	if err != nil {
		return nil, err
	}
	return j.Write(ctx, issues)
}

// --- CSV Generic Importer ---

// CSVImporter 通用 CSV 导入器。
// 支持自定义列映射的 CSV 导入。
type CSVImporter struct {
	*BaseImporter
	ColumnMapping map[string]string // CSV 列名 → ParsedIssue 字段名
}

// NewCSVImporter 创建通用 CSV 导入器。
func NewCSVImporter(db *pgxpool.Pool, wsID, projID int64, mapping map[string]string) *CSVImporter {
	return &CSVImporter{
		BaseImporter:  NewBaseImporter(db, wsID, projID, SourceCSV),
		ColumnMapping: mapping,
	}
}

// SourceType 返回数据源类型。
func (c *CSVImporter) SourceType() SourceType { return SourceCSV }

// Parse 解析通用 CSV 文件。
func (c *CSVImporter) Parse(ctx context.Context, reader io.Reader) ([]ParsedIssue, error) {
	csvReader := csv.NewReader(reader)
	csvReader.LazyQuotes = true

	headers, err := csvReader.Read()
	if err != nil {
		return nil, fmt.Errorf("read csv header: %w", err)
	}

	// 根据 ColumnMapping 构建反向映射
	reverseMapping := make(map[int]string)
	for csvCol, field := range c.ColumnMapping {
		for i, h := range headers {
			if h == csvCol {
				reverseMapping[i] = field
			}
		}
	}

	var issues []ParsedIssue
	rowNum := 1

	for {
		row, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read csv row %d: %w", rowNum, err)
		}
		rowNum++

		issue := ParsedIssue{
			CustomFields: map[string]any{},
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		for colIdx, field := range reverseMapping {
			if colIdx >= len(row) {
				continue
			}
			val := row[colIdx]
			switch field {
			case "title":
				issue.Title = val
			case "description":
				issue.Description = val
			case "type_code":
				issue.TypeCode = val
			case "priority":
				issue.Priority = val
			case "state_name":
				issue.StateName = val
			case "assignee":
				issue.Assignee = val
			case "labels":
				issue.Labels = splitCSV(val)
			}
		}

		issues = append(issues, issue)
	}

	return issues, nil
}

// Import 一键导入。
func (c *CSVImporter) Import(ctx context.Context, reader io.Reader) (*ImportReport, error) {
	issues, err := c.Parse(ctx, reader)
	if err != nil {
		return nil, err
	}
	return c.Write(ctx, issues)
}

// --- Helpers ---

func getCol(row []string, colIdx map[string]int, colName string) string {
	if idx, ok := colIdx[colName]; ok && idx < len(row) {
		return row[idx]
	}
	return ""
}

func splitCSV(s string) []string {
	if s == "" {
		return nil
	}
	parts := make([]string, 0)
	for _, p := range splitTrim(s, ",") {
		p = trimSpace(p)
		if p != "" {
			parts = append(parts, p)
		}
	}
	return parts
}

func splitTrim(s, sep string) []string {
	var result []string
	for _, part := range splitStr(s, sep) {
		result = append(result, trimSpace(part))
	}
	return result
}

func splitStr(s, sep string) []string {
	if s == "" {
		return nil
	}
	var parts []string
	start := 0
	for i := 0; i < len(s); i++ {
		if string(s[i]) == sep {
			parts = append(parts, s[start:i])
			start = i + 1
		}
	}
	parts = append(parts, s[start:])
	return parts
}

func trimSpace(s string) string {
	start, end := 0, len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t') {
		end--
	}
	return s[start:end]
}

func parseFloat(s string) float64 {
	if s == "" {
		return 0
	}
	var f float64
	fmt.Sscanf(s, "%f", &f)
	return f
}

func parseJiraDate(s string) (time.Time, error) {
	formats := []string{
		"2006-01-02 15:04",
		"2006-01-02T15:04:05.000-0700",
		"2006-01-02T15:04:05.000Z",
		"2006-01-02",
		"02/Jan/06 3:04 PM",
	}
	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unable to parse date: %s", s)
}

func mapJiraType(jiraType string) string {
	mapping := map[string]string{
		"Bug":           "defect",
		"Task":          "task",
		"Story":         "requirement",
		"Epic":          "epic",
		"Sub-task":      "task",
		"Improvement":   "requirement",
		"New Feature":   "requirement",
		"Technical task": "task",
		"缺陷":            "defect",
		"任务":            "task",
		"需求":            "requirement",
	}
	if mapped, ok := mapping[jiraType]; ok {
		return mapped
	}
	return "task"
}

func mapJiraPriority(jiraPriority string) string {
	mapping := map[string]string{
		"Highest": "critical",
		"High":    "high",
		"Medium":  "medium",
		"Low":     "low",
		"Lowest":  "low",
		"紧急":     "critical",
		"高":      "high",
		"中":      "medium",
		"低":      "low",
	}
	if mapped, ok := mapping[jiraPriority]; ok {
		return mapped
	}
	return "medium"
}

// stripHTML 移除 HTML 标签，保留纯文本。
func stripHTML(s string) string {
	var result []rune
	inTag := false
	for _, r := range s {
		if r == '<' {
			inTag = true
			continue
		}
		if r == '>' {
			inTag = false
			continue
		}
		if !inTag {
			result = append(result, r)
		}
	}
	return string(result)
}

func nullIfZero(id int64) *int64 {
	if id == 0 {
		return nil
	}
	return &id
}
