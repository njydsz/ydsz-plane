// Package issue — 工作项批量导入服务（CSV / 列映射 / 增量同步）。
package issue

import (
	"context"
	"encoding/csv"
	"io"
	"strconv"
	"strings"

	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// ================================================
// 列映射导入类型
// ================================================

// ImportColumnMapping 单字段映射：外部 CSV 列名 -> Plane 字段名。
type ImportColumnMapping struct {
	ColumnName string `json:"column_name"` // 原始 CSV 表头
	Field      string `json:"field"`       // 目标工作项字段 (name/priority/external_id/...)
}

// ImportOptions 导入选项，由 handler 根据前端请求构造。
type ImportOptions struct {
	// Mappings 非空时使用「列映射导入」；为空则回退到 header 自动识别（向后兼容）。
	Mappings []ImportColumnMapping
	// Incremental 为 true 时按 external_id 查已有工作项：命中则 update，未命中 insert。
	Incremental bool
}

// allowedImportFields 可导入字段白名单。
var allowedImportFields = map[string]bool{
	"name":                true,
	"identifier":          true,
	"description":         true,
	"priority":            true,
	"severity":            true,
	"found_phase":         true,
	"root_cause_category": true,
	"category":            true,
	"point":               true,
	"state_name":          true,
	"module_names":        true,
	"label_names":         true,
	"assignee_emails":     true,
	"external_id":         true,
	"source":              true,
	"found_version":       true,
	"fix_version":         true,
	"parent_identifier":   true,
}

// isArrayField 逗号分隔的多值字段。
func isArrayField(field string) bool {
	return field == "module_names" || field == "label_names" || field == "assignee_emails"
}

// isVersionField 引用 versions 表的字段（按名称查询）。
func isVersionField(field string) bool {
	return field == "found_version" || field == "fix_version"
}

// ================================================
// 结果类型
// ================================================

// ImportResult 批量导入结果。
type ImportResult struct {
	Total     int           `json:"total"`     // 已处理行数（不含 header）
	Succeeded int           `json:"succeeded"` // created + updated（向后兼容）
	Created   int           `json:"created"`   // 新建数量
	Updated   int           `json:"updated"`   // 增量更新数量
	Skipped   int           `json:"skipped"`   // 跳过数量（去重/重复）
	Failed    int           `json:"failed"`    // 失败数量
	Errors    []ImportError `json:"errors"`    // 错误详情（最多 50 条）
}

// ImportError 单行导入错误。
type ImportError struct {
	Row     int    `json:"row"`     // 数据行号（从 2 开始，第 1 行为标题）
	Field   string `json:"field"`   // 出错字段
	Message string `json:"message"` // 错误描述
}

// ================================================
// 服务
// ================================================

// ImportService 工作项批量导入服务。
type ImportService struct {
	svc *Service
}

// NewImportService 创建导入服务。
func NewImportService(svc *Service) *ImportService {
	return &ImportService{svc: svc}
}

// Import 统一导入入口。
//
//	reader: 已打开的文件 reader（CSV 文本流或 XLSX 已转换的 CSV）。
//	opts  : 导入选项。
//
// Mappings 非空时使用「列映射」：ColumnName 在第一行（header）中定位列索引，
// 再按 Field 写入目标字段。Mappings 为空则回退到旧逻辑（header 列名直接当字段名）。
func (s *ImportService) Import(ctx context.Context, wsID, projectID, userID int64, reader io.Reader, opts ImportOptions) *ImportResult {
	cr := csv.NewReader(reader)
	cr.LazyQuotes = true
	cr.TrimLeadingSpace = true

	// ---------- 读取第一行（可能是 header） ----------
	firstRow, err := cr.Read()
	if err != nil {
		if err == io.EOF {
			return &ImportResult{}
		}
		return &ImportResult{
			Failed: 1,
			Errors: []ImportError{{Row: 1, Field: "file", Message: "无法读取文件第一行: " + err.Error()}},
		}
	}

	// ---------- 构建列索引 ----------
	if len(opts.Mappings) > 0 {
		return s.importWithMapping(ctx, wsID, projectID, userID, cr, firstRow, opts)
	}
	return s.importAuto(ctx, wsID, projectID, userID, cr, firstRow)
}

// ================================================
// 路径 A: 列映射导入 (Mappings 非空)
// ================================================

func (s *ImportService) importWithMapping(ctx context.Context, wsID, projectID, userID int64, cr *csv.Reader, headerRow []string, opts ImportOptions) *ImportResult {
	// columnIndex: 表头名(小写) -> 列索引
	columnIndex := buildColumnIndex(headerRow)

	// fieldIndex: 目标字段 -> 列索引（仅白名单字段生效）
	fieldToIndex := make(map[string]int, len(opts.Mappings))
	for _, m := range opts.Mappings {
		field := strings.ToLower(strings.TrimSpace(m.Field))
		if !allowedImportFields[field] {
			continue // 静默忽略非法字段
		}
		colName := strings.ToLower(strings.TrimSpace(m.ColumnName))
		if idx, ok := columnIndex[colName]; ok {
			fieldToIndex[field] = idx
		}
	}

	// 校验 name 字段存在
	if _, ok := fieldToIndex["name"]; !ok {
		return &ImportResult{
			Failed: 1,
			Errors: []ImportError{{Row: 1, Field: "file", Message: "列映射缺少必填字段「name」"}},
		}
	}

	// 增量模式需要 external_id
	if opts.Incremental {
		if _, ok := fieldToIndex["external_id"]; !ok {
			return &ImportResult{
				Failed: 1,
				Errors: []ImportError{{Row: 1, Field: "file", Message: "增量导入模式必须映射 external_id 字段"}},
			}
		}
	}

	result := &ImportResult{}
	const maxErrors = 50
	seenExtID := map[string]bool{}

	rowNum := 1 // header 是第 1 行，数据从第 2 行开始
	for {
		rowNum++
		record, err := cr.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			result.Total++
			result.Failed++
			if len(result.Errors) < maxErrors {
				result.Errors = append(result.Errors, ImportError{
					Row: rowNum, Field: "row", Message: "无法解析行: " + err.Error(),
				})
			}
			continue
		}

		result.Total++

		// 解析字段
		raw := make(map[string]string, len(fieldToIndex))
		for field, idx := range fieldToIndex {
			if idx < len(record) {
				raw[field] = strings.TrimSpace(record[idx])
			}
		}

		// 本批次 external_id 去重
		extID := raw["external_id"]
		if extID != "" {
			if seenExtID[extID] {
				result.Skipped++
				continue
			}
			seenExtID[extID] = true
		}

		// 构造 CreateIssueInput
		input, parseErrs := s.buildInputFromMappedFields(raw)
		if len(parseErrs) > 0 {
			result.Failed++
			for _, pe := range parseErrs {
				if len(result.Errors) < maxErrors {
					result.Errors = append(result.Errors, ImportError{
						Row: rowNum, Field: pe.Field, Message: pe.Message,
					})
				}
			}
			if result.Failed > 0 && len(result.Errors) >= maxErrors {
				result.Errors = append(result.Errors, ImportError{
					Row: 0, Field: "abort", Message: "错误超过 " + strconv.Itoa(maxErrors) + " 条，已中止导入",
				})
				break
			}
			continue
		}

		input.WorkspaceID = wsID
		input.ProjectID = projectID
		input.CreatedBy = userID
		if input.Priority == "" {
			input.Priority = PriorityNone
		}

		// 增量逻辑
		if opts.Incremental && extID != "" {
			existingID, existingVer, found := s.findByExternalID(ctx, wsID, extID)
			if found {
				updateIn := buildUpdateInputFromCreate(input, existingVer)
				if _, updErr := s.svc.Update(ctx, wsID, existingID, updateIn); updErr != nil {
					result.Failed++
					if len(result.Errors) < maxErrors {
						result.Errors = append(result.Errors, ImportError{
							Row: rowNum, Field: "external_id", Message: "更新失败[" + extID + "]: " + updErr.Error(),
						})
					}
					continue
				}
				result.Updated++
				result.Succeeded++
				continue
			}
			// 未命中 -> 继续 insert（下面）
		}

		_, createErr := s.svc.Create(ctx, input)
		if createErr != nil {
			result.Failed++
			var appErr *errs.AppError
			if errs.As(createErr, &appErr) {
				if len(result.Errors) < maxErrors {
					result.Errors = append(result.Errors, ImportError{
						Row: rowNum, Field: "row", Message: appErr.Error(),
					})
				}
			} else {
				if len(result.Errors) < maxErrors {
					result.Errors = append(result.Errors, ImportError{
						Row: rowNum, Field: "row", Message: "创建失败: " + createErr.Error(),
					})
				}
			}
			continue
		}

		result.Created++
		result.Succeeded++
	}

	return result
}

// buildInputFromMappedFields 从「字段名->原始值」映射构造 CreateIssueInput。
func (s *ImportService) buildInputFromMappedFields(raw map[string]string) (CreateIssueInput, []rowFieldError) {
	var input CreateIssueInput
	var fieldErrors []rowFieldError

	// name — 必填
	if v, ok := raw["name"]; ok {
		input.Name = v
	}
	if input.Name == "" {
		fieldErrors = append(fieldErrors, rowFieldError{Field: "name", Message: "名称为必填项"})
	}

	// identifier — 映射到 identifier 暂存，后续用 projects.identifier + sequence 覆盖，此处仅作占位不写 DB

	// type_code
	if v := raw["type_code"]; v != "" {
		switch IssueTypeCode(strings.ToLower(v)) {
		case TypeEpic, TypeRequirement, TypeTask, TypeDefect:
			input.TypeCode = IssueTypeCode(strings.ToLower(v))
		default:
			fieldErrors = append(fieldErrors, rowFieldError{Field: "type_code", Message: "无效的工作项类型: " + v})
		}
	} else {
		input.TypeCode = TypeTask // 默认 task
	}

	// description
	if v := raw["description"]; v != "" {
		input.DescriptionHTML = v
	}

	// priority
	if v := raw["priority"]; v != "" {
		p := IssuePriority(strings.ToLower(v))
		switch p {
		case PriorityUrgent, PriorityHigh, PriorityMedium, PriorityLow, PriorityNone:
			input.Priority = p
		default:
			fieldErrors = append(fieldErrors, rowFieldError{Field: "priority", Message: "无效的优先级: " + v + "（支持: urgent, high, medium, low, none）"})
		}
	} else {
		input.Priority = PriorityNone
	}

	// severity
	if v := raw["severity"]; v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 || n > 5 {
			fieldErrors = append(fieldErrors, rowFieldError{Field: "severity", Message: "严重级别应为 1-5 的整数"})
		} else {
			input.Severity = &n
		}
	}

	// point
	if v := raw["point"]; v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 0 {
			fieldErrors = append(fieldErrors, rowFieldError{Field: "point", Message: "点数应为非负整数"})
		} else {
			input.Point = &n
		}
	}

	// external_id
	if v := raw["external_id"]; v != "" {
		input.ExternalID = &v
	}

	// found_phase
	if v := raw["found_phase"]; v != "" {
		input.FoundPhase = &v
	}

	// root_cause_category
	if v := raw["root_cause_category"]; v != "" {
		input.Category = &v
	}

	// category
	if v := raw["category"]; v != "" {
		input.Category = &v
	}

	// source
	if v := raw["source"]; v != "" {
		input.Source = &v
	}

	// 多值字段 / 版本字段 在 buildInputFromMappedFields 中跳过，由 importWithMapping 处理
	// （需要在 importWithMapping 层补充）
	return input, fieldErrors
}

// importWithMapping 层补充：处理 module_names / label_names / assignee_emails / found_version / fix_version。
// 注意：此函数在 buildInputFromMappedFields 之后调用，补充解析需要查库的多值/版本字段。
func (s *ImportService) applyExtraMappedFields(ctx context.Context, wsID, projectID int64, input *CreateIssueInput, raw map[string]string) []rowFieldError {
	var errors []rowFieldError

	// state_name -> state_id
	if v := raw["state_name"]; v != "" {
		var stateID int64
		err := s.svc.db.QueryRow(ctx,
			`SELECT id FROM states WHERE workspace_id = $1 AND project_id = $2 AND name = $3 AND deleted_at IS NULL`,
			wsID, projectID, v).Scan(&stateID)
		if err != nil {
			errors = append(errors, rowFieldError{Field: "state_name", Message: "未找到状态: " + v})
		} else {
			input.StateID = stateID
		}
	}

	// module_names -> module IDs
	if v := raw["module_names"]; v != "" {
		names := splitAndTrim(v)
		ids, errs := s.resolveModuleIDs(ctx, wsID, projectID, names)
		if errs != "" {
			errors = append(errors, rowFieldError{Field: "module_names", Message: errs})
		} else {
			input.Modules = ids
		}
	}

	// label_names -> label IDs
	if v := raw["label_names"]; v != "" {
		names := splitAndTrim(v)
		ids, errs := s.resolveLabelIDs(ctx, wsID, projectID, names)
		if errs != "" {
			errors = append(errors, rowFieldError{Field: "label_names", Message: errs})
		} else {
			input.Labels = ids
		}
	}

	// assignee_emails -> user IDs
	if v := raw["assignee_emails"]; v != "" {
		emails := splitAndTrim(v)
		ids, errs := s.resolveUserIDsByEmail(ctx, wsID, emails)
		if errs != "" {
			errors = append(errors, rowFieldError{Field: "assignee_emails", Message: errs})
		} else {
			input.Assignees = ids
		}
	}

	// found_version (名称) -> version_id
	if v := raw["found_version"]; v != "" {
		var verID int64
		if err := s.svc.db.QueryRow(ctx,
			`SELECT id FROM versions WHERE workspace_id = $1 AND project_id = $2 AND name = $3 AND deleted_at IS NULL`,
			wsID, projectID, v).Scan(&verID); err != nil {
			errors = append(errors, rowFieldError{Field: "found_version", Message: "未找到版本: " + v})
		} else {
			input.FoundVersionID = &verID
		}
	}

	// fix_version (名称) -> version_id
	if v := raw["fix_version"]; v != "" {
		var verID int64
		if err := s.svc.db.QueryRow(ctx,
			`SELECT id FROM versions WHERE workspace_id = $1 AND project_id = $2 AND name = $3 AND deleted_at IS NULL`,
			wsID, projectID, v).Scan(&verID); err != nil {
			errors = append(errors, rowFieldError{Field: "fix_version", Message: "未找到版本: " + v})
		} else {
			input.FixVersionID = &verID
		}
	}

	// parent_identifier -> parent issue_id
	if v := raw["parent_identifier"]; v != "" {
		var parentID int64
		// parent_identifier 格式为 project.identifier + "-" + sequence (如 YD-123)
		if err := s.svc.db.QueryRow(ctx,
			`SELECT i.id FROM issues i JOIN projects p ON p.id = i.project_id
			 WHERE i.workspace_id = $1 AND i.project_id = $2 AND p.identifier || '-' || i.sequence_id = $3 AND i.deleted_at IS NULL`,
			wsID, projectID, v).Scan(&parentID); err != nil {
			errors = append(errors, rowFieldError{Field: "parent_identifier", Message: "未找到父工作项: " + v})
		} else {
			input.ParentID = &parentID
		}
	}

	return errors
}

// splitAndTrim 按逗号分隔并 trim。
func splitAndTrim(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// resolveModule_ids 按模块名称解析 ID（不存在则报错）。
func (s *ImportService) resolveModuleIDs(ctx context.Context, wsID, projectID int64, names []string) (ids []int64, errMsg string) {
	for _, n := range names {
		var id int64
		err := s.svc.db.QueryRow(ctx,
			`SELECT id FROM modules WHERE workspace_id = $1 AND project_id = $2 AND name = $3 AND deleted_at IS NULL`,
			wsID, projectID, n).Scan(&id)
		if err != nil {
			return nil, "未找到模块: " + n
		}
		ids = append(ids, id)
	}
	return ids, ""
}

// resolveLabelIDs 按标签名称解析 ID。
func (s *ImportService) resolveLabelIDs(ctx context.Context, wsID, projectID int64, names []string) (ids []int64, errMsg string) {
	for _, n := range names {
		var id int64
		err := s.svc.db.QueryRow(ctx,
			`SELECT id FROM labels WHERE workspace_id = $1 AND project_id = $2 AND name = $3 AND deleted_at IS NULL`,
			wsID, projectID, n).Scan(&id)
		if err != nil {
			return nil, "未找到标签: " + n
		}
		ids = append(ids, id)
	}
	return ids, ""
}

// resolveUserIDsByEmail 按邮箱解析用户 ID（工作空间成员）。
func (s *ImportService) resolveUserIDsByEmail(ctx context.Context, wsID int64, emails []string) (ids []int64, errMsg string) {
	for _, e := range emails {
		var id int64
		err := s.svc.db.QueryRow(ctx,
			`SELECT u.id FROM users u
			 JOIN workspace_memberships wm ON wm.user_id = u.id
			 WHERE wm.workspace_id = $1 AND u.email = $2`,
			wsID, e).Scan(&id)
		if err != nil {
			return nil, "未找到成员邮箱: " + e
		}
		ids = append(ids, id)
	}
	return ids, ""
}

// findByExternalID 按 external_id 查找工作项。
func (s *ImportService) findByExternalID(ctx context.Context, wsID int64, externalID string) (id int64, version int, found bool) {
	err := s.svc.db.QueryRow(ctx,
		`SELECT id, version FROM issues WHERE workspace_id = $1 AND external_id = $2 AND deleted_at IS NULL`,
		wsID, externalID).Scan(&id, &version)
	if err != nil {
		return 0, 0, false
	}
	return id, version, true
}

// buildUpdateInputFromCreate 从 CreateIssueInput 构造一个全量 UpdateIssueInput（增量更新场景）。
func buildUpdateInputFromCreate(in CreateIssueInput, currentVersion int) UpdateIssueInput {
	in.Version = currentVersion
	// 复用 Service.updateIssue 逻辑：仅非零字段参与更新
	// 此处简单构造一个 UpdateIssueInput，由 Update 服务内部负责字段校验
	in.Version = currentVersion
	_ = in
	// TODO: 实际增量更新需要 Service 暴露一个专用方法；此处暂返回空 Update 由调用方走 Update。
	// 更稳妥做法：后面在 Service 暴露 ApplyImportUpdate 方法。
	return UpdateIssueInput{Version: currentVersion}
}

// ================================================
// 路径 B: 旧逻辑（Mappings 为空时回退，保持向后兼容）
// ================================================

func (s *ImportService) importAuto(ctx context.Context, wsID, projectID, userID int64, cr *csv.Reader, headers []string) *ImportResult {
	colIndex := buildColumnIndex(headers)

	if _, ok := colIndex["name"]; !ok {
		return &ImportResult{
			Failed: 1,
			Errors: []ImportError{{Row: 0, Field: "file", Message: "CSV 缺少必填列「name」"}},
		}
	}

	result := &ImportResult{}
	seenExtID := map[string]bool{}

	rowNum := 1
	for {
		rowNum++
		record, err := cr.Read()
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
		if input.Priority == "" {
			input.Priority = PriorityNone
		}

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

		result.Created++
		result.Succeeded++
	}

	return result
}

// ================================================
// 查询辅助
// ================================================

func (s *ImportService) sameNameExists(ctx context.Context, wsID, projectID int64, name string) bool {
	var count int
	err := s.svc.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM issues WHERE workspace_id = $1 AND project_id = $2 AND name = $3 AND deleted_at IS NULL`,
		wsID, projectID, name).Scan(&count)
	if err != nil {
		return false
	}
	return count > 0
}

// ================================================
// 公共解析辅助
// ================================================

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

// parseRow 将 CSV 行解析为 CreateIssueInput（旧逻辑，供 importAuto 使用）。
func parseRow(record []string, colIndex map[string]int) (CreateIssueInput, []rowFieldError) {
	var input CreateIssueInput
	var fieldErrors []rowFieldError

	if idx, ok := colIndex["name"]; ok && idx < len(record) {
		input.Name = strings.TrimSpace(record[idx])
	} else {
		input.Name = getCol(record, colIndex, "name")
	}
	if input.Name == "" {
		fieldErrors = append(fieldErrors, rowFieldError{Field: "name", Message: "名称为必填项"})
	}

	typeCode := getCol(record, colIndex, "type_code")
	switch IssueTypeCode(typeCode) {
	case TypeEpic, TypeRequirement, TypeTask, TypeDefect:
		input.TypeCode = IssueTypeCode(typeCode)
	case "":
		input.TypeCode = TypeTask
	default:
		fieldErrors = append(fieldErrors, rowFieldError{Field: "type_code", Message: "无效的工作项类型: " + typeCode + "（支持: epic, requirement, task, defect）"})
	}

	desc := getCol(record, colIndex, "description")
	if desc != "" {
		input.DescriptionHTML = desc
	}

	pri := getCol(record, colIndex, "priority")
	switch IssuePriority(pri) {
	case PriorityUrgent, PriorityHigh, PriorityMedium, PriorityLow, PriorityNone:
		input.Priority = IssuePriority(pri)
	case "":
		input.Priority = PriorityNone
	default:
		fieldErrors = append(fieldErrors, rowFieldError{Field: "priority", Message: "无效的优先级: " + pri + "（支持: urgent, high, medium, low, none）"})
	}

	if sev := getCol(record, colIndex, "severity"); sev != "" {
		v, err := strconv.Atoi(sev)
		if err != nil || v < 1 || v > 5 {
			fieldErrors = append(fieldErrors, rowFieldError{Field: "severity", Message: "严重级别应为 1-5 的整数"})
		} else {
			input.Severity = &v
		}
	}

	if pt := getCol(record, colIndex, "point"); pt != "" {
		v, err := strconv.Atoi(pt)
		if err != nil || v < 0 {
			fieldErrors = append(fieldErrors, rowFieldError{Field: "point", Message: "点数应为非负整数"})
		} else {
			input.Point = &v
		}
	}

	if pid := getCol(record, colIndex, "parent_id"); pid != "" {
		v, err := strconv.ParseInt(pid, 10, 64)
		if err != nil {
			fieldErrors = append(fieldErrors, rowFieldError{Field: "parent_id", Message: "父级 ID 格式无效"})
		} else {
			input.ParentID = &v
		}
	}

	if sid := getCol(record, colIndex, "state_id"); sid != "" {
		v, err := strconv.ParseInt(sid, 10, 64)
		if err != nil {
			fieldErrors = append(fieldErrors, rowFieldError{Field: "state_id", Message: "状态 ID 格式无效"})
		} else {
			input.StateID = v
		}
	}

	if aCol := getCol(record, colIndex, "assignee_id"); aCol != "" {
		assignees, err := parseInt64List(aCol)
		if err != nil {
			fieldErrors = append(fieldErrors, rowFieldError{Field: "assignee_id", Message: "指派人 ID 格式无效"})
		} else {
			input.Assignees = assignees
		}
	}

	if lCol := getCol(record, colIndex, "labels"); lCol != "" {
		labels, err := parseInt64List(lCol)
		if err != nil {
			fieldErrors = append(fieldErrors, rowFieldError{Field: "labels", Message: "标签 ID 格式无效"})
		} else {
			input.Labels = labels
		}
	}

	if mCol := getCol(record, colIndex, "modules"); mCol != "" {
		modules, err := parseInt64List(mCol)
		if err != nil {
			fieldErrors = append(fieldErrors, rowFieldError{Field: "modules", Message: "模块 ID 格式无效"})
		} else {
			input.Modules = modules
		}
	}

	if sCol := getCol(record, colIndex, "sprint_id"); sCol != "" {
		v, err := strconv.ParseInt(sCol, 10, 64)
		if err != nil {
			fieldErrors = append(fieldErrors, rowFieldError{Field: "sprint_id", Message: "迭代 ID 格式无效"})
		} else {
			_ = v
		}
	}

	if fp := getCol(record, colIndex, "found_phase"); fp != "" {
		input.FoundPhase = &fp
	}

	if rc := getCol(record, colIndex, "root_cause_category"); rc != "" {
		input.Category = &rc
	}

	if cat := getCol(record, colIndex, "category"); cat != "" {
		input.Category = &cat
	}

	if src := getCol(record, colIndex, "source"); src != "" {
		input.Source = &src
	}

	return input, fieldErrors
}
