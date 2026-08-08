// Package issue — 工作项批量导入服务。
package issue

import (
	"context"
	"encoding/csv"
	"io"
	"strconv"
	"strings"

	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// ImportResult 批量导入结果。
type ImportResult struct {
	Total     int           `json:"total"`
	Succeeded int           `json:"succeeded"`
	Skipped   int           `json:"skipped"`
	Failed    int           `json:"failed"`
	Errors    []ImportError `json:"errors"`
}

// ImportError 单行导入错误。
type ImportError struct {
	Row     int    `json:"row"`     // 数据行号（从 2 开始，第 1 行为标题）
	Field   string `json:"field"`   // 出错的字段名
	Message string `json:"message"` // 错误描述
}

// ImportService 工作项批量导入服务。
type ImportService struct {
	svc *Service
}

// NewImportService 创建导入服务。
func NewImportService(svc *Service) *ImportService {
	return &ImportService{svc: svc}
}

// Import 从 CSV reader 批量导入工作项。
func (s *ImportService) Import(ctx context.Context, wsID, projectID, userID int64, reader io.Reader) *ImportResult {
	r := csv.NewReader(reader)
	r.LazyQuotes = true
	r.TrimLeadingSpace = true

	// 读标题行
	headers, err := r.Read()
	if err != nil {
		if err == io.EOF {
			return &ImportResult{}
		}
		return &ImportResult{
			Failed: 1,
			Errors: []ImportError{{Row: 0, Field: "file", Message: "无法读取 CSV 标题行: " + err.Error()}},
		}
	}

	colIndex := buildColumnIndex(headers)

	// 校验 name 列是否存在
	if _, ok := colIndex["name"]; !ok {
		return &ImportResult{
			Failed: 1,
			Errors: []ImportError{{Row: 0, Field: "file", Message: "CSV 缺少必填列「name」"}},
		}
	}

	result := &ImportResult{}

	// 本批次 external_id 去重
	seenExtID := map[string]bool{}

	rowNum := 1 // 标题行是第 1 行，数据行从第 2 行开始
	for {
		rowNum++
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			result.Failed++
			result.Total++
			result.Errors = append(result.Errors, ImportError{
				Row: rowNum, Field: "row", Message: "无法解析行: " + err.Error(),
			})
			continue
		}

		result.Total++

		// external_id 去重
		if extIDIdx, ok := colIndex["external_id"]; ok && extIDIdx < len(record) {
			extID := strings.TrimSpace(record[extIDIdx])
			if extID != "" {
				if seenExtID[extID] {
					result.Skipped++
					continue
				}
				seenExtID[extID] = true
			}
		}

		// 解析字段
		input, parseErrors := parseRow(record, colIndex)
		if len(parseErrors) > 0 {
			result.Failed++
			for _, pe := range parseErrors {
				result.Errors = append(result.Errors, ImportError{
					Row: rowNum, Field: pe.Field, Message: pe.Message,
				})
			}
			continue
		}

		input.WorkspaceID = wsID
		input.ProjectID = projectID
		input.CreatedBy = userID
		// 默认优先级
		if input.Priority == "" {
			input.Priority = PriorityNone
		}

		// 检查同批次中同名工作项是否已导入
		if s.sameNameExists(ctx, wsID, projectID, input.Name) {
			result.Skipped++
			continue
		}

		_, err = s.svc.Create(ctx, input)
		if err != nil {
			result.Failed++
			var appErr *errs.AppError
			if errs.As(err, &appErr) {
				result.Errors = append(result.Errors, ImportError{
					Row: rowNum, Field: "row", Message: appErr.Error(),
				})
			} else {
				result.Errors = append(result.Errors, ImportError{
					Row: rowNum, Field: "row", Message: "创建失败: " + err.Error(),
				})
			}
			continue
		}

		result.Succeeded++
	}

	return result
}

// sameNameExists 检查同名工作项是否已存在（用于增量去重）。
func (s *ImportService) sameNameExists(ctx context.Context, wsID, projectID int64, name string) bool {
	var count int
	err := s.svc.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM issues WHERE workspace_id = $1 AND project_id = $2 AND name = $3 AND deleted_at IS NULL`,
		wsID, projectID, name).Scan(&count)
	if err != nil {
		return false // 查询失败保守起见不跳过
	}
	return count > 0
}

// rowFieldError 辅助结构用于解析行时收集字段错误。
type rowFieldError struct {
	Field   string
	Message string
}

// buildColumnIndex 构建列名到索引的映射（大小写不敏感）。
func buildColumnIndex(headers []string) map[string]int {
	m := make(map[string]int, len(headers))
	for i, h := range headers {
		m[strings.TrimSpace(strings.ToLower(h))] = i
	}
	return m
}

// parseRow 将 CSV 行解析为 CreateIssueInput，返回字段级错误列表。
func parseRow(record []string, colIndex map[string]int) (CreateIssueInput, []rowFieldError) {
	var input CreateIssueInput
	var fieldErrors []rowFieldError

	// name — 必填
	if idx, ok := colIndex["name"]; ok && idx < len(record) {
		input.Name = strings.TrimSpace(record[idx])
	} else {
		input.Name = getCol(record, colIndex, "name")
	}
	if input.Name == "" {
		fieldErrors = append(fieldErrors, rowFieldError{Field: "name", Message: "名称为必填项"})
	}

	// type_code — 映射到 requirement / task / defect，默认 task
	typeCode := getCol(record, colIndex, "type_code")
	switch IssueTypeCode(typeCode) {
	case TypeEpic, TypeRequirement, TypeTask, TypeDefect:
		input.TypeCode = IssueTypeCode(typeCode)
	case "":
		input.TypeCode = TypeTask
	default:
		fieldErrors = append(fieldErrors, rowFieldError{Field: "type_code", Message: "无效的工作项类型: " + typeCode + "（支持: epic, requirement, task, defect）"})
	}

	// description
	desc := getCol(record, colIndex, "description")
	if desc != "" {
		input.DescriptionHTML = desc
	}

	// priority
	pri := getCol(record, colIndex, "priority")
	switch IssuePriority(pri) {
	case PriorityUrgent, PriorityHigh, PriorityMedium, PriorityLow, PriorityNone:
		input.Priority = IssuePriority(pri)
	case "":
		input.Priority = PriorityNone
	default:
		fieldErrors = append(fieldErrors, rowFieldError{Field: "priority", Message: "无效的优先级: " + pri + "（支持: urgent, high, medium, low, none）"})
	}

	// severity — 仅 defect 类型使用
	if sev := getCol(record, colIndex, "severity"); sev != "" {
		v, err := strconv.Atoi(sev)
		if err != nil || v < 1 || v > 5 {
			fieldErrors = append(fieldErrors, rowFieldError{Field: "severity", Message: "严重级别应为 1-5 的整数"})
		} else {
			input.Severity = &v
		}
	}

	// point
	if pt := getCol(record, colIndex, "point"); pt != "" {
		v, err := strconv.Atoi(pt)
		if err != nil || v < 0 {
			fieldErrors = append(fieldErrors, rowFieldError{Field: "point", Message: "点数应为非负整数"})
		} else {
			input.Point = &v
		}
	}

	// parent_id
	if pid := getCol(record, colIndex, "parent_id"); pid != "" {
		v, err := strconv.ParseInt(pid, 10, 64)
		if err != nil {
			fieldErrors = append(fieldErrors, rowFieldError{Field: "parent_id", Message: "父级 ID 格式无效"})
		} else {
			input.ParentID = &v
		}
	}

	// state_id — 空值时由后端取默认
	if sid := getCol(record, colIndex, "state_id"); sid != "" {
		v, err := strconv.ParseInt(sid, 10, 64)
		if err != nil {
			fieldErrors = append(fieldErrors, rowFieldError{Field: "state_id", Message: "状态 ID 格式无效"})
		} else {
			input.StateID = v
		}
	}

	// assignee_id — 可多个（逗号分隔）
	if aCol := getCol(record, colIndex, "assignee_id"); aCol != "" {
		assignees, err := parseInt64List(aCol)
		if err != nil {
			fieldErrors = append(fieldErrors, rowFieldError{Field: "assignee_id", Message: "指派人 ID 格式无效"})
		} else {
			input.Assignees = assignees
		}
	}

	// labels — 可多个（逗号分隔）
	if lCol := getCol(record, colIndex, "labels"); lCol != "" {
		labels, err := parseInt64List(lCol)
		if err != nil {
			fieldErrors = append(fieldErrors, rowFieldError{Field: "labels", Message: "标签 ID 格式无效"})
		} else {
			input.Labels = labels
		}
	}

	// modules — 可多个（逗号分隔）
	if mCol := getCol(record, colIndex, "modules"); mCol != "" {
		modules, err := parseInt64List(mCol)
		if err != nil {
			fieldErrors = append(fieldErrors, rowFieldError{Field: "modules", Message: "模块 ID 格式无效"})
		} else {
			input.Modules = modules
		}
	}

	// sprint_id
	if sCol := getCol(record, colIndex, "sprint_id"); sCol != "" {
		v, err := strconv.ParseInt(sCol, 10, 64)
		if err != nil {
			fieldErrors = append(fieldErrors, rowFieldError{Field: "sprint_id", Message: "迭代 ID 格式无效"})
		} else {
			// sprint_id 不在 CreateIssueInput 中，通过更新处理
			// 暂不处理：导入时不直接关联迭代，需要后续单独更新
			_ = v
		}
	}

	// found_phase
	if fp := getCol(record, colIndex, "found_phase"); fp != "" {
		input.FoundPhase = &fp
	}

	// root_cause_category — 缺陷根因分类
	if rc := getCol(record, colIndex, "root_cause_category"); rc != "" {
		input.Category = &rc
	}

	// category
	if cat := getCol(record, colIndex, "category"); cat != "" {
		input.Category = &cat
	}

	// source
	if src := getCol(record, colIndex, "source"); src != "" {
		input.Source = &src
	}

	return input, fieldErrors
}

// getCol 从 CSV 行中按列名获取值。
func getCol(record []string, colIndex map[string]int, name string) string {
	idx, ok := colIndex[strings.ToLower(name)]
	if !ok || idx >= len(record) {
		return ""
	}
	return strings.TrimSpace(record[idx])
}

// parseInt64List 解析逗号分隔的 int64 列表。
func parseInt64List(s string) ([]int64, error) {
	parts := strings.Split(s, ",")
	var result []int64
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		v, err := strconv.ParseInt(p, 10, 64)
		if err != nil {
			return nil, err
		}
		result = append(result, v)
	}
	return result, nil
}

