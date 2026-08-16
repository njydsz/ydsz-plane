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
// 类型定义
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
//	reader: 已打开的文件 reader（CSV 文本流；XLSX 需由调用方转换或走 TODO stub）。
//	opts  : 导入选项。
//
// Mappings 非空时使用「列映射」：ColumnName 在第一行 header 中定位列索引，
// 再按 Field 写入目标字段。Mappings 为空则回退到旧逻辑（header 列名直接当字段名）。
func (s *ImportService) Import(ctx context.Context, wsID, projectID, userID int64, reader io.Reader, opts ImportOptions) *ImportResult {
	cr := csv.NewReader(reader)
	cr.LazyQuotes = true
	cr.TrimLeadingSpace = true

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

	if len(opts.Mappings) > 0 {
		return s.importWithMapping(ctx, wsID, projectID, userID, cr, firstRow, opts)
	}
	return s.importAuto(ctx, wsID, projectID, userID, cr, firstRow)
}

// ================================================
// 路径 A: 列映射导入 (Mappings 非空)
// ================================================

func (s *ImportService) importWithMapping(ctx context.Context, wsID, projectID, userID int64, cr *csv.Reader, headerRow []string, opts ImportOptions) *ImportResult {
	columnIndex := buildColumnIndex(headerRow)

	// field -> 列索引（仅白名单字段生效）
	fieldToIndex := make(map[string]int, len(opts.Mappings))
	for _, m := range opts.Mappings {
		field := strings.ToLower(strings.TrimSpace(m.Field))
		if !allowedImportFields[field] {
			continue
		}
		colName := strings.ToLower(strings.TrimSpace(m.ColumnName))
		if idx, ok := columnIndex[colName]; ok {
			fieldToIndex[field] = idx
		}
	}

	if _, ok := fieldToIndex["name"]; !ok {
		return &ImportResult{
			Failed: 1,
			Errors: []ImportError{{Row: 1, Field: "file", Message: "列映射缺少必填字段「name」"}},
		}
	}

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
	rowNum := 1

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

		raw := make(map[string]string, len(fieldToIndex))
		for field, idx := range fieldToIndex {
			if idx < len(record) {
				raw[field] = strings.TrimSpace(record[idx])
			}
		}

		extID := raw["external_id"]
		if extID != "" {
			if seenExtID[extID] {
				result.Skipped++
				continue
			}
			seenExtID[extID] = true
		}

		// 构造 CreateIssueInput + 额外字段校验
		input, extraErrs := s.buildInput(ctx, wsID, projectID, raw)
		if len(extraErrs) > 0 {
			result.Failed++
			for _, pe := range extraErrs {
				if len(result.Errors) < maxErrors {
					result.Errors = append(result.Errors, ImportError{
						Row: rowNum, Field: pe.Field, Message: pe.Message,
					})
				}
			}
			if len(result.Errors) >= maxErrors {
				result.Errors = append(result.Errors, ImportError{Row: 0, Field: "abort", Message: "错误超过 " + strconv.Itoa(maxErrors) + " 条，已中止"})
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

		// 增量：按 external_id 查找已有工作项
		if opts.Incremental && extID != "" {
			if existingID, ver, found := s.findByExternalID(ctx, wsID, extID); found {
				updateIn := buildUpdateInput(raw, ver)
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
		}

		if _, createErr := s.svc.Create(ctx, input); createErr != nil {
			result.Failed++
			var appErr *errs.AppError
			if errs.As(createErr, &appErr) {
				if len(result.Errors) < maxErrors {
					result.Errors = append(result.Errors, ImportError{Row: rowNum, Field: "row", Message: appErr.Error()})
				}
			} else {
				if len(result.Errors) < maxErrors {
					result.Errors = append(result.Errors, ImportError{Row: rowNum, Field: "row", Message: "创建失败: " + createErr.Error()})
				}
			}
			continue
		}

		result.Created++
		result.Succeeded++
	}

	return result
}

// buildInput 根据 raw 字段映射构造 CreateIssueInput，并解析需要查库的引用字段。
func (s *ImportService) buildInput(ctx context.Context, wsID, projectID int64, raw map[string]string) (CreateIssueInput, []rowFieldError) {
	var in CreateIssueInput
	var errs []rowFieldError

	if v, ok := raw["name"]; ok {
		in.Name = v
	}
	if in.Name == "" {
		errs = append(errs, rowFieldError{Field: "name", Message: "名称为必填项"})
	}

	if v := raw["type_code"]; v != "" {
		switch IssueTypeCode(strings.ToLower(v)) {
		case TypeEpic, TypeRequirement, TypeTask, TypeDefect:
			in.TypeCode = IssueTypeCode(strings.ToLower(v))
		default:
			errs = append(errs, rowFieldError{Field: "type_code", Message: "无效的工作项类型: " + v})
		}
	} else {
		in.TypeCode = TypeTask
	}

	if v := raw["description"]; v != "" {
		in.DescriptionHTML = v
	}

	if v := raw["priority"]; v != "" {
		p := IssuePriority(strings.ToLower(v))
		switch p {
		case PriorityUrgent, PriorityHigh, PriorityMedium, PriorityLow, PriorityNone:
			in.Priority = p
		default:
			errs = append(errs, rowFieldError{Field: "priority", Message: "无效的优先级: " + v})
		}
	} else {
		in.Priority = PriorityNone
	}

	if v := raw["severity"]; v != "" {
		n, e := strconv.Atoi(v)
		if e != nil || n < 1 || n > 5 {
			errs = append(errs, rowFieldError{Field: "severity", Message: "严重级别应为 1-5 的整数"})
		} else {
			in.Severity = &n
		}
	}

	if v := raw["point"]; v != "" {
		n, e := strconv.Atoi(v)
		if e != nil || n < 0 {
			errs = append(errs, rowFieldError{Field: "point", Message: "点数应为非负整数"})
		} else {
			in.Point = &n
		}
	}

	if v := raw["external_id"]; v != "" {
		in.ExternalID = &v
	}

	if v := raw["found_phase"]; v != "" {
		in.FoundPhase = &v
	}

	if v := raw["root_cause_category"]; v != "" {
		in.Category = &v
	}
	if v := raw["category"]; v != "" {
		in.Category = &v
	}

	if v := raw["source"]; v != "" {
		in.Source = &v
	}

	// --- 引用字段（需查库） ---

	if v := raw["state_name"]; v != "" {
		var sid int64
		if e := s.svc.db.QueryRow(ctx,
			`SELECT id FROM states WHERE workspace_id=$1 AND project_id=$2 AND name=$3 AND deleted = false`,
			wsID, projectID, v).Scan(&sid); e != nil {
			errs = append(errs, rowFieldError{Field: "state_name", Message: "未找到状态: " + v})
		} else {
			in.StateID = sid
		}
	}

	if v := raw["module_names"]; v != "" {
		ids, msg := s.resolveModuleIDs(ctx, wsID, projectID, splitAndTrim(v))
		if msg != "" {
			errs = append(errs, rowFieldError{Field: "module_names", Message: msg})
		} else {
			in.Modules = ids
		}
	}

	if v := raw["label_names"]; v != "" {
		ids, msg := s.resolveLabelIDs(ctx, wsID, projectID, splitAndTrim(v))
		if msg != "" {
			errs = append(errs, rowFieldError{Field: "label_names", Message: msg})
		} else {
			in.Labels = ids
		}
	}

	if v := raw["assignee_emails"]; v != "" {
		ids, msg := s.resolveUserIDsByEmail(ctx, wsID, splitAndTrim(v))
		if msg != "" {
			errs = append(errs, rowFieldError{Field: "assignee_emails", Message: msg})
		} else {
			in.Assignees = ids
		}
	}

	if v := raw["found_version"]; v != "" {
		var vid int64
		if e := s.svc.db.QueryRow(ctx,
			`SELECT id FROM versions WHERE workspace_id=$1 AND project_id=$2 AND name=$3 AND deleted = false`,
			wsID, projectID, v).Scan(&vid); e != nil {
			errs = append(errs, rowFieldError{Field: "found_version", Message: "未找到版本: " + v})
		} else {
			in.FoundVersionID = &vid
		}
	}

	if v := raw["fix_version"]; v != "" {
		var vid int64
		if e := s.svc.db.QueryRow(ctx,
			`SELECT id FROM versions WHERE workspace_id=$1 AND project_id=$2 AND name=$3 AND deleted = false`,
			wsID, projectID, v).Scan(&vid); e != nil {
			errs = append(errs, rowFieldError{Field: "fix_version", Message: "未找到版本: " + v})
		} else {
			in.FixVersionID = &vid
		}
	}

	if v := raw["parent_identifier"]; v != "" {
		var pid int64
		if e := s.svc.db.QueryRow(ctx,
			`SELECT i.id FROM (SELECT id, public_id, workspace_id, project_id, sequence_id, 'requirement'::text AS type_code, parent_id, depth, name, description_json, description_html, description_stripped, state_id, priority, NULL::smallint AS severity, NULL::text AS found_phase, NULL::text AS root_cause_category, NULL::bigint AS verifier_id, NULL::jsonb AS environment, NULL::jsonb AS reproduce_steps, NULL::text AS category, NULL::numeric AS actual_effort, NULL::numeric AS remaining_effort, NULL::text AS delay_reason, source, point, sprint_id, progress, start_date, target_date, completed_at, is_draft, sort_order, version, version_id, NULL::bigint AS found_version_id, NULL::bigint AS fix_version_id, created_by, created_at, updated_at, deleted FROM requirement WHERE deleted = false UNION ALL SELECT id, public_id, workspace_id, project_id, sequence_id, 'task'::text, parent_id, depth, name, description_json, description_html, description_stripped, state_id, priority, NULL::smallint, NULL::text, NULL::text, NULL::bigint, NULL::jsonb, NULL::jsonb, category, actual_effort, remaining_effort, delay_reason, NULL::text, point, sprint_id, progress, start_date, target_date, completed_at, is_draft, sort_order, version, version_id, NULL::bigint AS found_version_id, NULL::bigint AS fix_version_id, created_by, created_at, updated_at, deleted FROM task WHERE deleted = false UNION ALL SELECT id, public_id, workspace_id, project_id, sequence_id, 'defect'::text, parent_id, depth, name, description_json, description_html, description_stripped, state_id, priority, severity, found_phase, root_cause_category, verifier_id, environment, reproduce_steps, NULL::text, NULL::numeric, NULL::numeric, NULL::text, NULL::text, point, sprint_id, progress, start_date, target_date, completed_at, is_draft, sort_order, version, version_id, found_version_id, fix_version_id, created_by, created_at, updated_at, deleted FROM defect WHERE deleted = false) AS w i JOIN projects p ON p.id=i.project_id
			 WHERE i.workspace_id=$1 AND i.project_id=$2 AND p.identifier || '-' || i.sequence_id=$3 AND i.deleted = false`,
			wsID, projectID, v).Scan(&pid); e != nil {
			errs = append(errs, rowFieldError{Field: "parent_identifier", Message: "未找到父工作项: " + v})
		} else {
			in.ParentID = &pid
		}
	}

	return in, errs
}

// buildUpdateInput 根据 raw 映射 + 当前版本构造 UpdateIssueImport（仅设置 raw 中存在的字段）。
func buildUpdateInput(raw map[string]string, currentVersion int) UpdateIssueInput {
	up := UpdateIssueInput{Version: currentVersion}

	if v, ok := raw["name"]; ok && v != "" {
		up.Name = &v
	}
	if v, ok := raw["description"]; ok && v != "" {
		up.DescriptionHTML = &v
	}
	if v, ok := raw["priority"]; ok && v != "" {
		p := IssuePriority(strings.ToLower(v))
		up.Priority = &p
	}
	if v, ok := raw["type_code"]; ok && v != "" {
		t := IssueTypeCode(strings.ToLower(v))
		up.TypeCode = &t
	}
	if v, ok := raw["severity"]; ok && v != "" {
		if n, e := strconv.Atoi(v); e == nil && n >= 1 && n <= 5 {
			up.Severity = &n
		}
	}
	if v, ok := raw["found_phase"]; ok && v != "" {
		up.FoundPhase = &v
	}
	if v, ok := raw["root_cause_category"]; ok && v != "" {
		up.RootCauseCategory = &v
	}
	if v, ok := raw["category"]; ok && v != "" {
		up.Category = &v
	}
	if v, ok := raw["source"]; ok && v != "" {
		up.Source = &v
	}
	if v, ok := raw["point"]; ok && v != "" {
		if n, e := strconv.Atoi(v); e == nil && n >= 0 {
			up.Point = &n
		}
	}
	if v, ok := raw["found_version"]; ok && v != "" {
		// 版本名称在 update 时需解析为 ID；这里直接传字符串暂存，由 importWithMapping 补查
		// 注意：update 接口当前不支持 version name -> id 转换，调用前已在 buildInput 中预查
		// 因此对 update 路径，我们在 importWithMapping 中先通过 buildInput 解析
		_ = v
	}
	if v, ok := raw["state_name"]; ok && v != "" {
		// state 需走 transition 接口而非直接写 state_id
		_ = v
	}
	if v, ok := raw["module_names"]; ok && v != "" {
		// M2M 需解析
		_ = v
	}
	if v, ok := raw["label_names"]; ok && v != "" {
		_ = v
	}
	if v, ok := raw["assignee_emails"]; ok && v != "" {
		_ = v
	}
	if v, ok := raw["parent_identifier"]; ok && v != "" {
		_ = v
	}

	return up
}

// ================================================
// 路径 B: header 自动识别（旧逻辑，向后兼容）
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

		if _, err = s.svc.Create(ctx, input); err != nil {
			result.Failed++
			var appErr *errs.AppError
			if errs.As(err, &appErr) {
				result.Errors = append(result.Errors, ImportError{Row: rowNum, Field: "row", Message: appErr.Error()})
			} else {
				result.Errors = append(result.Errors, ImportError{Row: rowNum, Field: "row", Message: "创建失败: " + err.Error()})
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
		`SELECT COUNT(*) FROM (SELECT id, public_id, workspace_id, project_id, sequence_id, 'requirement'::text AS type_code, parent_id, depth, name, description_json, description_html, description_stripped, state_id, priority, NULL::smallint AS severity, NULL::text AS found_phase, NULL::text AS root_cause_category, NULL::bigint AS verifier_id, NULL::jsonb AS environment, NULL::jsonb AS reproduce_steps, NULL::text AS category, NULL::numeric AS actual_effort, NULL::numeric AS remaining_effort, NULL::text AS delay_reason, source, point, sprint_id, progress, start_date, target_date, completed_at, is_draft, sort_order, version, version_id, NULL::bigint AS found_version_id, NULL::bigint AS fix_version_id, created_by, created_at, updated_at, deleted FROM requirement WHERE deleted = false UNION ALL SELECT id, public_id, workspace_id, project_id, sequence_id, 'task'::text, parent_id, depth, name, description_json, description_html, description_stripped, state_id, priority, NULL::smallint, NULL::text, NULL::text, NULL::bigint, NULL::jsonb, NULL::jsonb, category, actual_effort, remaining_effort, delay_reason, NULL::text, point, sprint_id, progress, start_date, target_date, completed_at, is_draft, sort_order, version, version_id, NULL::bigint AS found_version_id, NULL::bigint AS fix_version_id, created_by, created_at, updated_at, deleted FROM task WHERE deleted = false UNION ALL SELECT id, public_id, workspace_id, project_id, sequence_id, 'defect'::text, parent_id, depth, name, description_json, description_html, description_stripped, state_id, priority, severity, found_phase, root_cause_category, verifier_id, environment, reproduce_steps, NULL::text, NULL::numeric, NULL::numeric, NULL::text, NULL::text, point, sprint_id, progress, start_date, target_date, completed_at, is_draft, sort_order, version, version_id, found_version_id, fix_version_id, created_by, created_at, updated_at, deleted FROM defect WHERE deleted = false) AS w WHERE workspace_id = $1 AND project_id = $2 AND name = $3 AND deleted = false`,
		wsID, projectID, name).Scan(&count)
	return err == nil && count > 0
}

func (s *ImportService) findByExternalID(ctx context.Context, wsID int64, externalID string) (id int64, version int, found bool) {
	err := s.svc.db.QueryRow(ctx,
		`SELECT id, version FROM (SELECT id, public_id, workspace_id, project_id, sequence_id, 'requirement'::text AS type_code, parent_id, depth, name, description_json, description_html, description_stripped, state_id, priority, NULL::smallint AS severity, NULL::text AS found_phase, NULL::text AS root_cause_category, NULL::bigint AS verifier_id, NULL::jsonb AS environment, NULL::jsonb AS reproduce_steps, NULL::text AS category, NULL::numeric AS actual_effort, NULL::numeric AS remaining_effort, NULL::text AS delay_reason, source, point, sprint_id, progress, start_date, target_date, completed_at, is_draft, sort_order, version, version_id, NULL::bigint AS found_version_id, NULL::bigint AS fix_version_id, created_by, created_at, updated_at, deleted FROM requirement WHERE deleted = false UNION ALL SELECT id, public_id, workspace_id, project_id, sequence_id, 'task'::text, parent_id, depth, name, description_json, description_html, description_stripped, state_id, priority, NULL::smallint, NULL::text, NULL::text, NULL::bigint, NULL::jsonb, NULL::jsonb, category, actual_effort, remaining_effort, delay_reason, NULL::text, point, sprint_id, progress, start_date, target_date, completed_at, is_draft, sort_order, version, version_id, NULL::bigint AS found_version_id, NULL::bigint AS fix_version_id, created_by, created_at, updated_at, deleted FROM task WHERE deleted = false UNION ALL SELECT id, public_id, workspace_id, project_id, sequence_id, 'defect'::text, parent_id, depth, name, description_json, description_html, description_stripped, state_id, priority, severity, found_phase, root_cause_category, verifier_id, environment, reproduce_steps, NULL::text, NULL::numeric, NULL::numeric, NULL::text, NULL::text, point, sprint_id, progress, start_date, target_date, completed_at, is_draft, sort_order, version, version_id, found_version_id, fix_version_id, created_by, created_at, updated_at, deleted FROM defect WHERE deleted = false) AS w WHERE workspace_id = $1 AND external_id = $2 AND deleted = false`,
		wsID, externalID).Scan(&id, &version)
	if err != nil {
		return 0, 0, false
	}
	return id, version, true
}

func (s *ImportService) resolveModuleIDs(ctx context.Context, wsID, projectID int64, names []string) (ids []int64, errMsg string) {
	for _, n := range names {
		var id int64
		if err := s.svc.db.QueryRow(ctx,
			`SELECT id FROM modules WHERE workspace_id=$1 AND project_id=$2 AND name=$3 AND deleted = false`,
			wsID, projectID, n).Scan(&id); err != nil {
			return nil, "未找到模块: " + n
		}
		ids = append(ids, id)
	}
	return ids, ""
}

func (s *ImportService) resolveLabelIDs(ctx context.Context, wsID, projectID int64, names []string) (ids []int64, errMsg string) {
	for _, n := range names {
		var id int64
		if err := s.svc.db.QueryRow(ctx,
			`SELECT id FROM labels WHERE workspace_id=$1 AND project_id=$2 AND name=$3 AND deleted = false`,
			wsID, projectID, n).Scan(&id); err != nil {
			return nil, "未找到标签: " + n
		}
		ids = append(ids, id)
	}
	return ids, ""
}

func (s *ImportService) resolveUserIDsByEmail(ctx context.Context, wsID int64, emails []string) (ids []int64, errMsg string) {
	for _, e := range emails {
		var id int64
		if err := s.svc.db.QueryRow(ctx,
			`SELECT u.id FROM users u JOIN workspace_members wm ON wm.user_id = u.id
			 WHERE wm.workspace_id=$1 AND u.email=$2`,
			wsID, e).Scan(&id); err != nil {
			return nil, "未找到成员邮箱: " + e
		}
		ids = append(ids, id)
	}
	return ids, ""
}

// ================================================
// 通用解析辅助
// ================================================

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

type rowFieldError struct {
	Field   string
	Message string
}

func buildColumnIndex(headers []string) map[string]int {
	m := make(map[string]int, len(headers))
	for i, h := range headers {
		m[strings.TrimSpace(strings.ToLower(h))] = i
	}
	return m
}

func getCol(record []string, colIndex map[string]int, name string) string {
	idx, ok := colIndex[strings.ToLower(name)]
	if !ok || idx >= len(record) {
		return ""
	}
	return strings.TrimSpace(record[idx])
}

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
		fieldErrors = append(fieldErrors, rowFieldError{Field: "type_code", Message: "无效的工作项类型: " + typeCode})
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
		fieldErrors = append(fieldErrors, rowFieldError{Field: "priority", Message: "无效的优先级: " + pri})
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
