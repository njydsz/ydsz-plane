// Package importer — CSV 导入器。
package importer

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

// CSVImporter CSV 导入器。
type CSVImporter struct{}

// NewCSVImporter 构造导入器。
func NewCSVImporter() *CSVImporter { return &CSVImporter{} }

// Source 返回来源。
func (c *CSVImporter) Source() Source { return SourceCSV }

// Fetch 解析 CSV（从 Reader）。
func (c *CSVImporter) Fetch(_ context.Context, _ Credentials, _ ImportOptions) ([]WorkItemRecord, error) {
	return nil, fmt.Errorf("use FetchFromReader for CSV import")
}

// FetchFromReader 从 io.Reader 解析 CSV。
//
// 支持的列（大小写不敏感，支持中英文别名）：
//
//	external_id / 外部ID
//	title / 标题 / 名称
//	description / 描述
//	type / 类型（requirement|task|defect|需求|任务|缺陷）
//	priority / 优先级（urgent|high|medium|low|none|紧急|高|中|低）
//	state / 状态
//	assignee / 指派人（邮箱）
//	creator / 创建人（邮箱）
//	labels / 标签（逗号分隔）
//	due_date / 截止日期（YYYY-MM-DD）
//	parent_external_id / 父项ID
//	estimate / 估算
func (c *CSVImporter) FetchFromReader(r io.Reader, opts ImportOptions) ([]WorkItemRecord, error) {
	reader := csv.NewReader(r)
	reader.TrimLeadingSpace = true
	reader.FieldsPerRecord = -1 // 允许变长

	rows, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(rows) < 2 {
		return nil, fmt.Errorf("CSV 至少需要表头与一行数据")
	}

	// 解析表头 → 字段索引
	header := rows[0]
	idx := map[string]int{}
	for i, h := range header {
		idx[normalizeHeader(h)] = i
	}

	var out []WorkItemRecord
	for lineNo, row := range rows[1:] {
		if isEmptyRow(row) {
			continue
		}
		rec := WorkItemRecord{
			ExternalID:       cell(row, idx, "external_id"),
			Title:            cell(row, idx, "title"),
			Description:      cell(row, idx, "description"),
			TypeCode:         normalizeType(cell(row, idx, "type"), opts.DefaultType),
			Priority:         normalizePriority(cell(row, idx, "priority")),
			StateName:        cell(row, idx, "state"),
			AssigneeEmail:    cell(row, idx, "assignee"),
			CreatorEmail:     cell(row, idx, "creator"),
			ParentExternalID: cell(row, idx, "parent_external_id"),
		}
		if rec.Title == "" {
			continue
		}
		if v := cell(row, idx, "labels"); v != "" {
			for _, l := range strings.Split(v, ",") {
				if l = strings.TrimSpace(l); l != "" {
					rec.Labels = append(rec.Labels, l)
				}
			}
		}
		if v := cell(row, idx, "due_date"); v != "" {
			if t, err := time.Parse("2006-01-02", v); err == nil {
				rec.DueDate = &t
			}
		}
		if v := cell(row, idx, "estimate"); v != "" {
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				rec.Estimate = &f
			}
		}
		if rec.ExternalID == "" {
			rec.ExternalID = fmt.Sprintf("csv-%d", lineNo+2)
		}
		out = append(out, rec)
	}
	return out, nil
}

// --- helpers ---

func normalizeHeader(h string) string {
	h = strings.ToLower(strings.TrimSpace(h))
	switch h {
	case "外部id", "外部编号", "编号", "id":
		return "external_id"
	case "标题", "名称", "主题", "summary":
		return "title"
	case "描述", "说明", "description":
		return "description"
	case "类型", "issue type", "issuetype":
		return "type"
	case "优先级", "priority":
		return "priority"
	case "状态", "state", "status":
		return "state"
	case "指派人", "负责人", "assignee":
		return "assignee"
	case "创建人", "报告人", "creator", "reporter":
		return "creator"
	case "标签", "labels":
		return "labels"
	case "截止日期", "到期日", "due_date", "duedate":
		return "due_date"
	case "父项id", "父工作项", "parent", "parent_external_id":
		return "parent_external_id"
	case "估算", "estimate", "story_points", "points":
		return "estimate"
	}
	return h
}

func cell(row []string, idx map[string]int, key string) string {
	i, ok := idx[key]
	if !ok || i >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[i])
}

func isEmptyRow(row []string) bool {
	for _, c := range row {
		if strings.TrimSpace(c) != "" {
			return false
		}
	}
	return true
}

func normalizeType(v, def string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "requirement", "req", "需求", "feature", "story":
		return "requirement"
	case "task", "任务":
		return "task"
	case "defect", "bug", "缺陷", "问题":
		return "defect"
	}
	if def != "" {
		return def
	}
	return "task"
}

func normalizePriority(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "urgent", "highest", "紧急", "最高", "p0":
		return "urgent"
	case "high", "高", "p1":
		return "high"
	case "low", "低", "p3":
		return "low"
	case "none", "无", "p4":
		return "none"
	}
	return "medium"
}
