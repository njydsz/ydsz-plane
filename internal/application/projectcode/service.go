// Package projectcode — Projects-as-Code 应用服务。
package projectcode

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"gopkg.in/yaml.v3"

	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// Service Projects-as-Code 应用服务。
type Service struct {
	db *pgxpool.Pool
}

// NewService 创建服务。
func NewService(db *pgxpool.Pool) *Service {
	return &Service{db: db}
}

// --- 错误 ---

var (
	// ErrConfigNotFound 配置不存在。
	ErrConfigNotFound = errors.New("project code config not found")
	// ErrProjectNotFound 项目不存在。
	ErrProjectNotFound = errors.New("project not found")
	// ErrInvalidYAML YAML 解析失败。
	ErrInvalidYAML = errors.New("invalid yaml")
)

// --- 常量 ---

// 合法状态分组（与 issue.StateGroup 对齐）。
var validStateGroups = map[string]bool{
	"backlog":   true,
	"started":   true,
	"completed": true,
	"cancelled": true,
}

// 合法成员角色（与 project_role_enum 对齐）。
var validMemberRoles = map[string]bool{
	"admin":  true,
	"member": true,
}

// 合法 spec 版本格式（semver 主版本号）。
var specVersionRe = regexp.MustCompile(`^[0-9]+(\.[0-9]+)*$`)

// --- 解析与校验 ---

// ParseSpec 解析并校验 YAML 项目定义。
//
// 校验规则：
//   - version / name / key 必填；
//   - key 必须为 2-6 位大写字母数字；
//   - 每个 state 必须有 name 与合法 group；
//   - states / labels / modules / members 名称唯一（大小写不敏感）；
//   - member email 必须合法；member role 必须合法。
func ParseSpec(yamlText string) (*ProjectSpec, error) {
	trimmed := strings.TrimSpace(yamlText)
	if trimmed == "" {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "yaml", Reason: "YAML 内容不能为空"})
	}

	var spec ProjectSpec
	if err := yaml.Unmarshal([]byte(yamlText), &spec); err != nil {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{
			Field:  "yaml",
			Reason: "YAML 解析失败: " + err.Error(),
		})
	}

	if err := validateSpec(&spec); err != nil {
		return nil, err
	}
	return &spec, nil
}

// validateSpec 对解析后的规范做语义校验。
func validateSpec(spec *ProjectSpec) error {
	spec.Version = strings.TrimSpace(spec.Version)
	spec.Name = strings.TrimSpace(spec.Name)
	spec.Key = strings.TrimSpace(strings.ToUpper(spec.Key))

	if spec.Version == "" {
		return errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "version", Reason: "spec version 不能为空"})
	}
	if !specVersionRe.MatchString(spec.Version) {
		return errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "version", Reason: "spec version 必须为数字版本号，如 1.0"})
	}
	if spec.Name == "" {
		return errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "name", Reason: "项目名称不能为空"})
	}
	if spec.Key == "" {
		return errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "key", Reason: "项目标识 key 不能为空"})
	}
	if !identifierRe.MatchString(spec.Key) {
		return errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "key", Reason: "项目标识 key 必须为 2-6 位大写字母或数字"})
	}

	// 状态：name + group 必填，名称唯一。
	stateNames := map[string]bool{}
	for i, st := range spec.States {
		st.Name = strings.TrimSpace(st.Name)
		st.Group = strings.TrimSpace(strings.ToLower(st.Group))
		if st.Name == "" {
			return errs.ErrValidation.WithDetails(errs.FieldDetail{Field: fmt.Sprintf("states[%d].name", i), Reason: "状态名不能为空"})
		}
		if !validStateGroups[st.Group] {
			return errs.ErrValidation.WithDetails(errs.FieldDetail{
				Field:  fmt.Sprintf("states[%d].group", i),
				Reason: fmt.Sprintf("无效的状态分组 %q（合法值：backlog/started/completed/cancelled）", st.Group),
			})
		}
		key := strings.ToLower(st.Name)
		if stateNames[key] {
			return errs.ErrValidation.WithDetails(errs.FieldDetail{Field: fmt.Sprintf("states[%d].name", i), Reason: fmt.Sprintf("状态名 %q 重复", st.Name)})
		}
		stateNames[key] = true
		spec.States[i] = st
	}

	// 标签：name 必填，名称唯一。
	labelNames := map[string]bool{}
	for i, lb := range spec.Labels {
		lb.Name = strings.TrimSpace(lb.Name)
		if lb.Name == "" {
			return errs.ErrValidation.WithDetails(errs.FieldDetail{Field: fmt.Sprintf("labels[%d].name", i), Reason: "标签名不能为空"})
		}
		key := strings.ToLower(lb.Name)
		if labelNames[key] {
			return errs.ErrValidation.WithDetails(errs.FieldDetail{Field: fmt.Sprintf("labels[%d].name", i), Reason: fmt.Sprintf("标签名 %q 重复", lb.Name)})
		}
		labelNames[key] = true
		spec.Labels[i] = lb
	}

	// 模块：name 必填，名称唯一。
	moduleNames := map[string]bool{}
	for i, m := range spec.Modules {
		m.Name = strings.TrimSpace(m.Name)
		if m.Name == "" {
			return errs.ErrValidation.WithDetails(errs.FieldDetail{Field: fmt.Sprintf("modules[%d].name", i), Reason: "模块名不能为空"})
		}
		key := strings.ToLower(m.Name)
		if moduleNames[key] {
			return errs.ErrValidation.WithDetails(errs.FieldDetail{Field: fmt.Sprintf("modules[%d].name", i), Reason: fmt.Sprintf("模块名 %q 重复", m.Name)})
		}
		moduleNames[key] = true
		spec.Modules[i] = m
	}

	// 成员：email 必填且合法，role 合法（默认 member），email 唯一。
	memberEmails := map[string]bool{}
	for i, mb := range spec.Members {
		mb.Email = strings.TrimSpace(strings.ToLower(mb.Email))
		if mb.Email == "" {
			return errs.ErrValidation.WithDetails(errs.FieldDetail{Field: fmt.Sprintf("members[%d].email", i), Reason: "成员邮箱不能为空"})
		}
		if !emailRe.MatchString(mb.Email) {
			return errs.ErrValidation.WithDetails(errs.FieldDetail{Field: fmt.Sprintf("members[%d].email", i), Reason: fmt.Sprintf("邮箱 %q 格式不正确", mb.Email)})
		}
		mb.Role = strings.TrimSpace(strings.ToLower(mb.Role))
		if mb.Role == "" {
			mb.Role = "member"
		}
		if !validMemberRoles[mb.Role] {
			return errs.ErrValidation.WithDetails(errs.FieldDetail{
				Field:  fmt.Sprintf("members[%d].role", i),
				Reason: fmt.Sprintf("无效的成员角色 %q（合法值：admin/member）", mb.Role),
			})
		}
		if memberEmails[mb.Email] {
			return errs.ErrValidation.WithDetails(errs.FieldDetail{Field: fmt.Sprintf("members[%d].email", i), Reason: fmt.Sprintf("成员邮箱 %q 重复", mb.Email)})
		}
		memberEmails[mb.Email] = true
		spec.Members[i] = mb
	}

	return nil
}

// identifierRe 项目标识：2-6 位大写字母或数字。
var identifierRe = regexp.MustCompile(`^[A-Z0-9]{2,6}$`)

// emailRe 简化邮箱校验（足够覆盖常见格式）。
var emailRe = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// --- 配置存储 ---

// GetConfig 查询单个配置。
func (s *Service) GetConfig(ctx context.Context, workspaceID, id int64) (*Config, error) {
	var cfg Config
	var specJSON []byte
	err := s.db.QueryRow(ctx, `
		SELECT id, workspace_id, COALESCE(project_id,0), name, COALESCE(spec_key,''), version,
		       yaml, spec_json, sha256, config_version, created_by, updated_by, created_at, updated_at
		FROM project_code_configs
		WHERE id = $1 AND workspace_id = $2 AND deleted = false`, id, workspaceID,
	).Scan(&cfg.ID, &cfg.WorkspaceID, &cfg.ProjectID, &cfg.Name, &cfg.SpecKey, &cfg.Version,
		&cfg.YAML, &specJSON, &cfg.SHA256, &cfg.ConfigVersion, &cfg.CreatedBy, &cfg.UpdatedBy,
		&cfg.CreatedAt, &cfg.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrConfigNotFound
	}
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	cfg.Spec = decodeSpec(specJSON)
	return &cfg, nil
}

// ListConfigs 列出工作空间下的配置。
func (s *Service) ListConfigs(ctx context.Context, opts ListConfigsOptions) ([]*Config, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, workspace_id, COALESCE(project_id,0), name, COALESCE(spec_key,''), version,
		       yaml, spec_json, sha256, config_version, created_by, updated_by, created_at, updated_at
		FROM project_code_configs
		WHERE workspace_id = $1 AND deleted = false
		  AND (CASE WHEN $2 > 0 THEN project_id = $2 ELSE TRUE END)
		ORDER BY updated_at DESC
		LIMIT $3 OFFSET $4`,
		opts.WorkspaceID, opts.ProjectID, opts.Limit, opts.Offset)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()

	var out []*Config
	for rows.Next() {
		var cfg Config
		var specJSON []byte
		if err := rows.Scan(&cfg.ID, &cfg.WorkspaceID, &cfg.ProjectID, &cfg.Name, &cfg.SpecKey, &cfg.Version,
			&cfg.YAML, &specJSON, &cfg.SHA256, &cfg.ConfigVersion, &cfg.CreatedBy, &cfg.UpdatedBy,
			&cfg.CreatedAt, &cfg.UpdatedAt); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
		cfg.Spec = decodeSpec(specJSON)
		out = append(out, &cfg)
	}
	return out, rows.Err()
}

// SaveConfig 保存（新增或更新）YAML 配置。
//
// 幂等策略：同一 workspace 下若已存在相同 sha256 且同 project_id 的配置，
// 则复用该记录并自增 config_version（不新建）。
func (s *Service) SaveConfig(ctx context.Context, in SaveConfigInput) (*Config, error) {
	if strings.TrimSpace(in.YAML) == "" {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "yaml", Reason: "YAML 内容不能为空"})
	}
	spec, err := ParseSpec(in.YAML)
	if err != nil {
		return nil, err
	}

	sum := sha256.Sum256([]byte(in.YAML))
	sha := hex.EncodeToString(sum[:])

	specJSON, err := json.Marshal(spec)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}

	// 幂等：同 workspace + 同 project + 同 sha256 → 更新既有记录的版本号。
	var existingID int64
	err = s.db.QueryRow(ctx, `
		SELECT id FROM project_code_configs
		WHERE workspace_id = $1 AND sha256 = $2
		  AND (project_id IS NOT DISTINCT FROM $3)
		  AND deleted = false
		ORDER BY updated_at DESC LIMIT 1`,
		in.WorkspaceID, sha, nullProjectID(in.ProjectID)).Scan(&existingID)
	if err == nil {
		return s.bumpConfigVersion(ctx, existingID, in)
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, errs.ErrInternal.Wrap(err)
	}

	name := in.Name
	if strings.TrimSpace(name) == "" {
		name = spec.Name
	}

	var cfg Config
	err = s.db.QueryRow(ctx, `
		INSERT INTO project_code_configs
			(id, workspace_id, project_id, name, spec_key, version, yaml, spec_json, sha256, config_version, created_by, updated_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,1,$10,$10)
		RETURNING id, workspace_id, COALESCE(project_id,0), name, COALESCE(spec_key,''), version,
		          yaml, spec_json, sha256, config_version, created_by, updated_by, created_at, updated_at`,
		genID(), in.WorkspaceID, nullProjectID(in.ProjectID), name, nullIfEmpty(spec.Key),
		spec.Version, in.YAML, specJSON, sha, in.CreatedBy,
	).Scan(&cfg.ID, &cfg.WorkspaceID, &cfg.ProjectID, &cfg.Name, &cfg.SpecKey, &cfg.Version,
		&cfg.YAML, &specJSON, &cfg.SHA256, &cfg.ConfigVersion, &cfg.CreatedBy, &cfg.UpdatedBy,
		&cfg.CreatedAt, &cfg.UpdatedAt)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	cfg.Spec = decodeSpec(specJSON)
	return &cfg, nil
}

// bumpConfigVersion 幂等命中时自增 config_version 并刷新内容。
func (s *Service) bumpConfigVersion(ctx context.Context, id int64, in SaveConfigInput) (*Config, error) {
	spec, err := ParseSpec(in.YAML)
	if err != nil {
		return nil, err
	}
	specJSON, err := json.Marshal(spec)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	name := in.Name
	if strings.TrimSpace(name) == "" {
		name = spec.Name
	}
	var cfg Config
	err = s.db.QueryRow(ctx, `
		UPDATE project_code_configs
		SET yaml = $1, spec_json = $2, name = $3, spec_key = COALESCE(NULLIF($4,''), spec_key),
		    version = $5, config_version = config_version + 1, updated_by = $6, updated_at = now()
		WHERE id = $7
		RETURNING id, workspace_id, COALESCE(project_id,0), name, COALESCE(spec_key,''), version,
		          yaml, spec_json, sha256, config_version, created_by, updated_by, created_at, updated_at`,
		in.YAML, specJSON, name, nullIfEmpty(spec.Key), spec.Version, in.UpdatedBy, id,
	).Scan(&cfg.ID, &cfg.WorkspaceID, &cfg.ProjectID, &cfg.Name, &cfg.SpecKey, &cfg.Version,
		&cfg.YAML, &specJSON, &cfg.SHA256, &cfg.ConfigVersion, &cfg.CreatedBy, &cfg.UpdatedBy,
		&cfg.CreatedAt, &cfg.UpdatedAt)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	cfg.Spec = decodeSpec(specJSON)
	return &cfg, nil
}

// DeleteConfig 删除配置（软删除）。
func (s *Service) DeleteConfig(ctx context.Context, workspaceID, id int64) error {
	tag, err := s.db.Exec(ctx, `
		UPDATE project_code_configs SET deleted = true, updated_at = now()
		WHERE id = $1 AND workspace_id = $2 AND deleted = false`, id, workspaceID)
	if err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	if tag.RowsAffected() == 0 {
		return ErrConfigNotFound
	}
	return nil
}

// --- 导出 ---

// ExportProject 从既有项目生成规范（反向 YAML）。
func (s *Service) ExportProject(ctx context.Context, workspaceID, projectID int64) (*ProjectSpec, error) {
	// 项目本体。
	var name, slug, identifier, description, icon, color *string
	var existingID int64
	err := s.db.QueryRow(ctx, `
		SELECT id, name, COALESCE(slug,''), COALESCE(identifier,''), description, icon, color
		FROM projects
		WHERE id = $1 AND workspace_id = $2 AND deleted = false`,
		projectID, workspaceID).Scan(&existingID, &name, &slug, &identifier, &description, &icon, &color)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrProjectNotFound
	}
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}

	spec := &ProjectSpec{
		Version:     "1.0",
		Name:        deref(name),
		Key:         deref(identifier),
		Color:       deref(color),
		Icon:        deref(icon),
		Description: deref(description),
	}

	// 状态。
	stateRows, err := s.db.Query(ctx, `
		SELECT name, "group", COALESCE(color,'')
		FROM states WHERE project_id = $1 AND workspace_id = $2 AND deleted = false
		ORDER BY sequence, id`, projectID, workspaceID)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	for stateRows.Next() {
		var st StateSpec
		if err := stateRows.Scan(&st.Name, &st.Group, &st.Color); err != nil {
			stateRows.Close()
			return nil, errs.ErrInternal.Wrap(err)
		}
		spec.States = append(spec.States, st)
	}
	stateRows.Close()
	if err := stateRows.Err(); err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}

	// 标签。
	labelRows, err := s.db.Query(ctx, `
		SELECT name, COALESCE(color,'')
		FROM labels WHERE project_id = $1 AND workspace_id = $2 AND deleted = false
		ORDER BY created_at`, projectID, workspaceID)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	for labelRows.Next() {
		var lb LabelSpec
		if err := labelRows.Scan(&lb.Name, &lb.Color); err != nil {
			labelRows.Close()
			return nil, errs.ErrInternal.Wrap(err)
		}
		spec.Labels = append(spec.Labels, lb)
	}
	labelRows.Close()
	if err := labelRows.Err(); err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}

	// 模块（含 lead email）。
	moduleRows, err := s.db.Query(ctx, `
		SELECT m.name, COALESCE(u.email,'')
		FROM modules m
		LEFT JOIN users u ON u.id = m.lead_id
		WHERE m.project_id = $1 AND m.workspace_id = $2 AND m.deleted = false
		ORDER BY m.sort_order, m.id`, projectID, workspaceID)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	for moduleRows.Next() {
		var m ModuleSpec
		if err := moduleRows.Scan(&m.Name, &m.Lead); err != nil {
			moduleRows.Close()
			return nil, errs.ErrInternal.Wrap(err)
		}
		spec.Modules = append(spec.Modules, m)
	}
	moduleRows.Close()
	if err := moduleRows.Err(); err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}

	// 成员（含 role）。
	memberRows, err := s.db.Query(ctx, `
		SELECT u.email, pm.role::text
		FROM project_members pm
		JOIN users u ON u.id = pm.user_id
		WHERE pm.project_id = $1 AND pm.workspace_id = $2
		ORDER BY pm.joined_at`, projectID, workspaceID)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	for memberRows.Next() {
		var mb MemberSpec
		if err := memberRows.Scan(&mb.Email, &mb.Role); err != nil {
			memberRows.Close()
			return nil, errs.ErrInternal.Wrap(err)
		}
		spec.Members = append(spec.Members, mb)
	}
	memberRows.Close()
	if err := memberRows.Err(); err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}

	return spec, nil
}

// MarshalYAML 将规范序列化为 YAML 文本（用于 export / CLI pull）。
func MarshalYAML(spec *ProjectSpec) ([]byte, error) {
	if spec == nil {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "spec", Reason: "规范不能为空"})
	}
	return yaml.Marshal(spec)
}

// --- 差异 ---

// DiffConfig 比较既有项目与规范，返回结构化差异。
//
// 仅做只读比对，不写入任何数据。用于 apply 前的预览（dry-run）。
func (s *Service) DiffConfig(ctx context.Context, workspaceID, projectID int64, spec *ProjectSpec) (*DiffResult, error) {
	if spec == nil {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "spec", Reason: "规范不能为空"})
	}

	// 读取既有项目快照为 spec 形式。
	existing, err := s.ExportProject(ctx, workspaceID, projectID)
	if err != nil {
		return nil, err
	}

	res := &DiffResult{
		ProjectID: projectID,
		SpecName:  spec.Name,
		Items:     []DiffItem{},
	}

	// 项目本体字段。
	diffProjectFields(res, existing, spec)

	// 状态 / 标签 / 模块 / 成员：按 name/email 对齐。
	diffNamedSlice(res, "state", specStatesToNamed(existing.States), specStatesToNamed(spec.States),
		func(a, b namedItem) []DiffItem {
			var items []DiffItem
			if a.field1 != b.field1 {
				items = append(items, DiffItem{Kind: ChangeUpdate, Resource: "state", Name: b.name, Field: "group", Old: a.field1, New: b.field1})
			}
			if a.field2 != b.field2 {
				items = append(items, DiffItem{Kind: ChangeUpdate, Resource: "state", Name: b.name, Field: "color", Old: a.field2, New: b.field2})
			}
			return items
		})

	diffNamedSlice(res, "label", specLabelsToNamed(existing.Labels), specLabelsToNamed(spec.Labels),
		func(a, b namedItem) []DiffItem {
			var items []DiffItem
			if a.field1 != b.field1 {
				items = append(items, DiffItem{Kind: ChangeUpdate, Resource: "label", Name: b.name, Field: "color", Old: a.field1, New: b.field1})
			}
			return items
		})

	diffNamedSlice(res, "module", specModulesToNamed(existing.Modules), specModulesToNamed(spec.Modules),
		func(a, b namedItem) []DiffItem {
			var items []DiffItem
			if a.field1 != b.field1 {
				items = append(items, DiffItem{Kind: ChangeUpdate, Resource: "module", Name: b.name, Field: "lead", Old: a.field1, New: b.field1})
			}
			return items
		})

	diffNamedSlice(res, "member", specMembersToNamed(existing.Members), specMembersToNamed(spec.Members),
		func(a, b namedItem) []DiffItem {
			var items []DiffItem
			if a.field1 != b.field1 {
				items = append(items, DiffItem{Kind: ChangeUpdate, Resource: "member", Name: b.name, Field: "role", Old: a.field1, New: b.field1})
			}
			return items
		})

	// 汇总。
	for _, it := range res.Items {
		switch it.Kind {
		case ChangeAdd:
			res.AddCount++
		case ChangeUpdate:
			res.UpdateCount++
		case ChangeRemove:
			res.RemoveCount++
		}
	}
	res.HasChanges = res.AddCount+res.UpdateCount+res.RemoveCount > 0
	return res, nil
}

// diffProjectFields 比对项目本体字段。
func diffProjectFields(res *DiffResult, existing, spec *ProjectSpec) {
	add := func(field, oldVal, newVal string) {
		if oldVal == newVal {
			return
		}
		res.Items = append(res.Items, DiffItem{Kind: ChangeUpdate, Resource: "project", Field: field, Old: oldVal, New: newVal})
	}
	add("name", existing.Name, spec.Name)
	add("key", existing.Key, spec.Key)
	add("color", existing.Color, spec.Color)
	add("icon", existing.Icon, spec.Icon)
	add("description", existing.Description, spec.Description)
	add("version", existing.Version, spec.Version)
}

// namedItem 通用命名项（用于 diff）。
type namedItem struct {
	name   string
	field1 string
	field2 string
}

// diffNamedSlice 对齐新旧命名集合，产出 add/update/remove 差异。
func diffNamedSlice(res *DiffResult, resource string, oldItems, newItems []namedItem,
	compare func(a, b namedItem) []DiffItem) {

	oldMap := make(map[string]namedItem, len(oldItems))
	for _, it := range oldItems {
		oldMap[strings.ToLower(it.name)] = it
	}
	newMap := make(map[string]namedItem, len(newItems))
	for _, it := range newItems {
		newMap[strings.ToLower(it.name)] = it
	}

	// 新增与更新（保持新集合顺序）。
	for _, nw := range newItems {
		key := strings.ToLower(nw.name)
		old, ok := oldMap[key]
		if !ok {
			res.Items = append(res.Items, DiffItem{Kind: ChangeAdd, Resource: resource, Name: nw.name})
			continue
		}
		res.Items = append(res.Items, compare(old, nw)...)
	}
	// 移除（旧集合有、新集合无）。
	for _, od := range oldItems {
		key := strings.ToLower(od.name)
		if _, ok := newMap[key]; !ok {
			res.Items = append(res.Items, DiffItem{Kind: ChangeRemove, Resource: resource, Name: od.name})
		}
	}
}

func specStatesToNamed(in []StateSpec) []namedItem {
	out := make([]namedItem, 0, len(in))
	for _, s := range in {
		out = append(out, namedItem{name: s.Name, field1: s.Group, field2: s.Color})
	}
	return out
}

func specLabelsToNamed(in []LabelSpec) []namedItem {
	out := make([]namedItem, 0, len(in))
	for _, s := range in {
		out = append(out, namedItem{name: s.Name, field1: s.Color})
	}
	return out
}

func specModulesToNamed(in []ModuleSpec) []namedItem {
	out := make([]namedItem, 0, len(in))
	for _, s := range in {
		out = append(out, namedItem{name: s.Name, field1: s.Lead})
	}
	return out
}

func specMembersToNamed(in []MemberSpec) []namedItem {
	out := make([]namedItem, 0, len(in))
	for _, s := range in {
		out = append(out, namedItem{name: s.Email, field1: s.Role})
	}
	return out
}

// --- 应用 ---

// ApplySpec 以规范为准创建或更新项目及其子资源。
//
// 策略：
//   - project_id > 0：更新既有项目（名称/标识/颜色/图标/描述）；
//   - project_id == 0：新建项目（校验 identifier 唯一）；
//   - 状态 / 标签 / 模块 / 成员：以规范为准增量对齐（新增缺失项，更新已存在项的可变字段，
//     不删除规范外的既有项，避免误删用户手工创建的数据）。
//
// 整个过程在单个事务内完成，失败回滚。
func (s *Service) ApplySpec(ctx context.Context, in ApplySpecInput) (*ApplyResult, error) {
	if in.Spec == nil {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "spec", Reason: "规范不能为空"})
	}
	if err := validateSpec(in.Spec); err != nil {
		return nil, err
	}
	if in.WorkspaceID <= 0 {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "workspace_id", Reason: "required"})
	}

	spec := in.Spec
	created := in.ProjectID <= 0

	var projectID int64
	var slug, identifier string

	err := pgx.BeginTxFunc(ctx, s.db, pgx.TxOptions{}, func(tx pgx.Tx) error {
		if created {
			// 新建项目。
			slug = normalizeSlug(spec.Key, spec.Name)
			identifier = spec.Key

			// 校验 identifier 在 workspace 内唯一。
			var exists int
			if err := tx.QueryRow(ctx, `
				SELECT COUNT(*) FROM projects
				WHERE workspace_id = $1 AND identifier = $2 AND deleted = false`,
				in.WorkspaceID, identifier).Scan(&exists); err != nil {
				return errs.ErrInternal.Wrap(err)
			}
			if exists > 0 {
				return errs.New("PROJECT.DUPLICATE", "项目标识已存在", 409)
			}

			err := tx.QueryRow(ctx, `
				INSERT INTO projects (id, workspace_id, name, slug, identifier, description, icon, color, network, template, created_by)
				VALUES ($1,$2,$3,$4,$5,$6,$7,$8,'public','generic',$9)
				RETURNING id`,
				genID(), in.WorkspaceID, spec.Name, slug, identifier,
				nullIfEmpty(spec.Description), nullIfEmpty(spec.Icon), nullIfEmpty(spec.Color),
				in.CreatedBy).Scan(&projectID)
			if err != nil {
				if strings.Contains(err.Error(), "idx_projects_workspace") {
					return errs.New("PROJECT.DUPLICATE", "项目链接标识或前缀已存在", 409)
				}
				return errs.ErrInternal.Wrap(err)
			}

			// 创建者自动成为项目 admin。
			if _, err := tx.Exec(ctx, `
				INSERT INTO project_members (workspace_id, project_id, user_id, role, created_by)
				VALUES ($1,$2,$3,'admin',$3)
				ON CONFLICT (workspace_id, project_id, user_id) DO UPDATE SET role = 'admin'`,
				in.WorkspaceID, projectID, in.CreatedBy); err != nil {
				return errs.ErrInternal.Wrap(err)
			}
		} else {
			// 更新既有项目。
			projectID = in.ProjectID
			var wsID int64
			var existingName string
			err := tx.QueryRow(ctx, `
				SELECT workspace_id, name FROM projects
				WHERE id = $1 AND deleted = false`, projectID).Scan(&wsID, &existingName)
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrProjectNotFound
			}
			if err != nil {
				return errs.ErrInternal.Wrap(err)
			}
			if wsID != in.WorkspaceID {
				return ErrProjectNotFound
			}
			slug = normalizeSlug(spec.Key, spec.Name)
			identifier = spec.Key

			// 若 identifier 变更，校验唯一性。
			var dupCount int
			if err := tx.QueryRow(ctx, `
				SELECT COUNT(*) FROM projects
				WHERE workspace_id = $1 AND identifier = $2 AND id <> $3 AND deleted = false`,
				in.WorkspaceID, identifier, projectID).Scan(&dupCount); err != nil {
				return errs.ErrInternal.Wrap(err)
			}
			if dupCount > 0 {
				return errs.New("PROJECT.DUPLICATE", "项目标识已存在", 409)
			}

			if _, err := tx.Exec(ctx, `
				UPDATE projects
				SET name = $1, slug = $2, identifier = $3,
				    description = $4, icon = $5, color = $6, updated_at = now()
				WHERE id = $7`,
				spec.Name, slug, identifier,
				nullIfEmpty(spec.Description), nullIfEmpty(spec.Icon), nullIfEmpty(spec.Color),
				projectID); err != nil {
				if strings.Contains(err.Error(), "idx_projects_workspace") {
					return errs.New("PROJECT.DUPLICATE", "项目链接标识或前缀已存在", 409)
				}
				return errs.ErrInternal.Wrap(err)
			}
			_ = existingName
		}

		// 子资源对齐。
		if err := s.applyStates(ctx, tx, in.WorkspaceID, projectID, spec, in.CreatedBy); err != nil {
			return err
		}
		if err := s.applyLabels(ctx, tx, in.WorkspaceID, projectID, spec, in.CreatedBy); err != nil {
			return err
		}
		if err := s.applyModules(ctx, tx, in.WorkspaceID, projectID, spec, in.CreatedBy); err != nil {
			return err
		}
		if err := s.applyMembers(ctx, tx, in.WorkspaceID, projectID, spec, in.CreatedBy); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &ApplyResult{
		ProjectID:   projectID,
		Created:     created,
		ProjectSlug: slug,
		Identifier:  identifier,
	}, nil
}

// applyStates 增量对齐状态（新增缺失，更新 group/color）。
func (s *Service) applyStates(ctx context.Context, tx pgx.Tx, wsID, projectID int64, spec *ProjectSpec, userID int64) error {
	// 既有状态名 → id/group/color。
	rows, err := tx.Query(ctx, `
		SELECT id, name, "group", COALESCE(color,'')
		FROM states WHERE project_id = $1 AND workspace_id = $2 AND deleted = false`,
		projectID, wsID)
	if err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	type existingState struct {
		id    int64
		group string
		color string
	}
	existing := map[string]existingState{}
	for rows.Next() {
		var id int64
		var name, group, color string
		if err := rows.Scan(&id, &name, &group, &color); err != nil {
			rows.Close()
			return errs.ErrInternal.Wrap(err)
		}
		existing[strings.ToLower(name)] = existingState{id: id, group: group, color: color}
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return errs.ErrInternal.Wrap(err)
	}

	for i, st := range spec.States {
		seq := float64((i + 1) * 1000)
		if ex, ok := existing[strings.ToLower(st.Name)]; ok {
			if ex.group != st.Group || ex.color != st.Color {
				if _, err := tx.Exec(ctx, `
					UPDATE states SET "group" = $1, color = $2, updated_at = now()
					WHERE id = $3`, st.Group, defaultColor(st.Color, "#8DA2C2"), ex.id); err != nil {
					return errs.ErrInternal.Wrap(err)
				}
			}
			continue
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO states (id, workspace_id, project_id, name, "group", color, sequence, is_default, applicable_types, created_by, updated_by)
			VALUES ($1,$2,$3,$4,$5,$6,$7,false,'{all}',$8,$8)`,
			genID(), wsID, projectID, st.Name, st.Group, defaultColor(st.Color, "#8DA2C2"), seq, userID); err != nil {
			return errs.ErrInternal.Wrap(err)
		}
	}
	return nil
}

// applyLabels 增量对齐标签（新增缺失，更新 color）。
func (s *Service) applyLabels(ctx context.Context, tx pgx.Tx, wsID, projectID int64, spec *ProjectSpec, userID int64) error {
	rows, err := tx.Query(ctx, `
		SELECT id, name, COALESCE(color,'')
		FROM labels WHERE project_id = $1 AND workspace_id = $2 AND deleted = false`,
		projectID, wsID)
	if err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	type existingLabel struct {
		id    int64
		color string
	}
	existing := map[string]existingLabel{}
	for rows.Next() {
		var id int64
		var name, color string
		if err := rows.Scan(&id, &name, &color); err != nil {
			rows.Close()
			return errs.ErrInternal.Wrap(err)
		}
		existing[strings.ToLower(name)] = existingLabel{id: id, color: color}
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return errs.ErrInternal.Wrap(err)
	}

	for _, lb := range spec.Labels {
		if ex, ok := existing[strings.ToLower(lb.Name)]; ok {
			if ex.color != lb.Color {
				if _, err := tx.Exec(ctx, `
					UPDATE labels SET color = $1, updated_at = now() WHERE id = $2`,
					defaultColor(lb.Color, "#8DA2C2"), ex.id); err != nil {
					return errs.ErrInternal.Wrap(err)
				}
			}
			continue
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO labels (id, workspace_id, project_id, name, color, status, created_by)
			VALUES ($1,$2,$3,$4,$5,'active',$6)`,
			genID(), wsID, projectID, lb.Name, defaultColor(lb.Color, "#8DA2C2"), userID); err != nil {
			return errs.ErrInternal.Wrap(err)
		}
	}
	return nil
}

// applyModules 增量对齐模块（新增缺失，更新 lead）。
func (s *Service) applyModules(ctx context.Context, tx pgx.Tx, wsID, projectID int64, spec *ProjectSpec, userID int64) error {
	rows, err := tx.Query(ctx, `
		SELECT id, name, COALESCE(lead_id,0)
		FROM modules WHERE project_id = $1 AND workspace_id = $2 AND deleted = false`,
		projectID, wsID)
	if err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	type existingModule struct {
		id     int64
		leadID int64
	}
	existing := map[string]existingModule{}
	for rows.Next() {
		var id, leadID int64
		var name string
		if err := rows.Scan(&id, &name, &leadID); err != nil {
			rows.Close()
			return errs.ErrInternal.Wrap(err)
		}
		existing[strings.ToLower(name)] = existingModule{id: id, leadID: leadID}
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return errs.ErrInternal.Wrap(err)
	}

	for i, m := range spec.Modules {
		leadID, err := s.resolveUserIDByEmail(ctx, tx, wsID, m.Lead)
		if err != nil {
			return err
		}
		if ex, ok := existing[strings.ToLower(m.Name)]; ok {
			if !ptrEqual(ex.leadID, leadID) {
				if _, err := tx.Exec(ctx, `
					UPDATE modules SET lead_id = $1, updated_at = now() WHERE id = $2`,
					nullInt64(leadID), ex.id); err != nil {
					return errs.ErrInternal.Wrap(err)
				}
			}
			continue
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO modules (id, workspace_id, project_id, name, description, lead_id, sort_order, created_by)
			VALUES ($1,$2,$3,$4,'',$5,$6,$7)`,
			genID(), wsID, projectID, m.Name, nullInt64(leadID), float64((i+1)*1000), userID); err != nil {
			return errs.ErrInternal.Wrap(err)
		}
	}
	return nil
}

// applyMembers 增量对齐成员（新增缺失，更新 role）。
//
// 成员 email 必须是同工作空间成员；否则跳过并记为「未解析」，不阻塞整体 apply。
func (s *Service) applyMembers(ctx context.Context, tx pgx.Tx, wsID, projectID int64, spec *ProjectSpec, userID int64) error {
	rows, err := tx.Query(ctx, `
		SELECT pm.user_id, u.email, pm.role::text
		FROM project_members pm
		JOIN users u ON u.id = pm.user_id
		WHERE pm.project_id = $1 AND pm.workspace_id = $2`,
		projectID, wsID)
	if err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	type existingMember struct {
		userID int64
		role   string
	}
	existing := map[string]existingMember{}
	for rows.Next() {
		var uid int64
		var email, role string
		if err := rows.Scan(&uid, &email, &role); err != nil {
			rows.Close()
			return errs.ErrInternal.Wrap(err)
		}
		existing[strings.ToLower(email)] = existingMember{userID: uid, role: role}
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return errs.ErrInternal.Wrap(err)
	}

	for _, mb := range spec.Members {
		uid, err := s.resolveUserIDByEmail(ctx, tx, wsID, mb.Email)
		if err != nil {
			return err
		}
		if uid == nil {
			// 非工作空间成员：跳过（不阻塞 apply）。
			continue
		}
		if ex, ok := existing[strings.ToLower(mb.Email)]; ok {
			if ex.role != mb.Role {
				if _, err := tx.Exec(ctx, `
					UPDATE project_members SET role = $1, updated_at = now()
					WHERE workspace_id = $2 AND project_id = $3 AND user_id = $4`,
					mb.Role, wsID, projectID, *uid); err != nil {
					return errs.ErrInternal.Wrap(err)
				}
			}
			continue
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO project_members (workspace_id, project_id, user_id, role, created_by)
			VALUES ($1,$2,$3,$4,$5)
			ON CONFLICT (workspace_id, project_id, user_id) DO UPDATE SET role = EXCLUDED.role`,
			wsID, projectID, *uid, mb.Role, userID); err != nil {
			return errs.ErrInternal.Wrap(err)
		}
	}
	return nil
}

// resolveUserIDByEmail 通过邮箱解析用户 ID（限定在同工作空间内）。
//
// 返回 nil 表示该邮箱不是当前工作空间成员（调用方据此跳过，不阻塞 apply）。
func (s *Service) resolveUserIDByEmail(ctx context.Context, tx pgx.Tx, wsID int64, email string) (*int64, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" {
		return nil, nil
	}
	var uid int64
	err := tx.QueryRow(ctx, `
		SELECT u.id FROM users u
		JOIN workspace_members wm ON wm.user_id = u.id
		WHERE u.email = $1 AND wm.workspace_id = $2
		LIMIT 1`, email, wsID).Scan(&uid)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	return &uid, nil
}

// --- helpers ---

func decodeSpec(b []byte) *ProjectSpec {
	if len(b) == 0 {
		return nil
	}
	var spec ProjectSpec
	if err := json.Unmarshal(b, &spec); err != nil {
		return nil
	}
	return &spec
}

func nullIfEmpty(s string) any {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return s
}

func nullProjectID(id int64) any {
	if id <= 0 {
		return nil
	}
	return id
}

func nullInt64(p *int64) any {
	if p == nil || *p <= 0 {
		return nil
	}
	return *p
}

func ptrEqual(a int64, b *int64) bool {
	if b == nil {
		return a == 0
	}
	return a == *b
}

func defaultColor(c, def string) string {
	if strings.TrimSpace(c) == "" {
		return def
	}
	return c
}

func deref(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

// normalizeSlug 生成项目 slug。
func normalizeSlug(key, name string) string {
	base := strings.TrimSpace(key)
	if base == "" {
		base = strings.TrimSpace(name)
	}
	base = strings.ToLower(base)
	var b strings.Builder
	prevDash := false
	for _, r := range base {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			prevDash = false
		} else if !prevDash {
			b.WriteByte('-')
			prevDash = true
		}
	}
	out := strings.Trim(b.String(), "-")
	if out == "" {
		return "project"
	}
	return out
}
