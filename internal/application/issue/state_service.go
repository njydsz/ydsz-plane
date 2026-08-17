// Package issue — 状态机服务（状态查询 + 流转校验）。
package issue

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// StateService 提供状态查询与流转校验。
type StateService struct {
	db *pgxpool.Pool
}

// NewStateService 创建状态服务。
func NewStateService(db *pgxpool.Pool) *StateService {
	return &StateService{db: db}
}

// GetProjectStates 列出项目全部状态（按 sequence 排序）。
func (s *StateService) GetProjectStates(ctx context.Context, wsID, projectID int64) ([]State, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, workspace_id, project_id, name, "group", color, sequence, is_default, created_at, updated_at
		FROM states
		WHERE project_id = $1 AND workspace_id = $2 AND deleted = false
		ORDER BY sequence, id`, projectID, wsID)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()

	var states []State
	for rows.Next() {
		var st State
		if err := rows.Scan(&st.ID, &st.WorkspaceID, &st.ProjectID, &st.Name,
			&st.Group, &st.Color, &st.Sequence, &st.IsDefault, &st.CreatedAt, &st.UpdatedAt); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
		states = append(states, st)
	}
	return states, rows.Err()
}

// GetStateByID 根据 ID 获取单个状态。
func (s *StateService) GetStateByID(ctx context.Context, wsID, stateID int64) (*State, error) {
	var st State
	err := s.db.QueryRow(ctx, `
		SELECT id, workspace_id, project_id, name, "group", color, sequence, is_default, created_at, updated_at
		FROM states WHERE id = $1 AND workspace_id = $2 AND deleted = false`,
		stateID, wsID).Scan(&st.ID, &st.WorkspaceID, &st.ProjectID, &st.Name,
		&st.Group, &st.Color, &st.Sequence, &st.IsDefault, &st.CreatedAt, &st.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrNotFound
		}
		return nil, errs.ErrInternal.Wrap(err)
	}
	return &st, nil
}

// GetDefaultState 获取项目的默认状态（用于行创建 issue）。
// 优先取 dev_flow 的默认；若无则取 sequence 最小的。
func (s *StateService) GetDefaultState(ctx context.Context, wsID, projectID int64) (*State, error) {
	var st State
	err := s.db.QueryRow(ctx, `
		SELECT id, workspace_id, project_id, name, "group", color, sequence, is_default, created_at, updated_at
		FROM states
		WHERE project_id = $1 AND workspace_id = $2 AND deleted = false
		ORDER BY is_default DESC, sequence ASC
		LIMIT 1`, projectID, wsID).Scan(&st.ID, &st.WorkspaceID, &st.ProjectID, &st.Name,
		&st.Group, &st.Color, &st.Sequence, &st.IsDefault, &st.CreatedAt, &st.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.New("ISSUE.NO_STATES", "项目尚未初始化状态集", 422)
		}
		return nil, errs.ErrInternal.Wrap(err)
	}
	return &st, nil
}

// StateNameIndex 项目状态名 → ID 的缓存索引（用于模板流转规则映射）。
type StateNameIndex map[string]int64

// GetStateNameIndex 构建项目状态名 → ID 映射。
func (s *StateService) GetStateNameIndex(ctx context.Context, wsID, projectID int64) (StateNameIndex, error) {
	states, err := s.GetProjectStates(ctx, wsID, projectID)
	if err != nil {
		return nil, err
	}
	index := make(StateNameIndex, len(states))
	for _, st := range states {
		index[st.Name] = st.ID
	}
	return index, nil
}

// TransitionInput 状态流转输入。
type TransitionInput struct {
	IssueID   int64
	FromState int64
	ToState   int64
	TypeCode  IssueTypeCode
	Context   TransitionContext // 当前工作项字段值（用于 required_fields 校验）
}

// TransitionContext 流转校验所需的当前字段值。
type TransitionContext struct {
	RootCauseCategory *string
	FixVersionID      *int64
}

// ValidateTransition 校验状态流转是否合法。
// 1. 查 state_transitions 表
// 2. 校验 required_fields
// 返回错误码 ISSUE.INVALID_STATE_TRANSITION 或 nil。
func (s *StateService) ValidateTransition(ctx context.Context, wsID, projectID int64, in TransitionInput) error {
	if in.FromState == in.ToState {
		return nil // 同状态无需流转
	}

	// 查显式流转规则
	allowed, err := s.checkTransitionRule(ctx, wsID, projectID, in)
	if err != nil {
		return err
	}
	if !allowed {
		return errs.ErrInvalidTransition
	}

	// 校验 required_fields
	return s.checkRequiredFields(ctx, wsID, in)
}

func (s *StateService) checkTransitionRule(ctx context.Context, wsID, projectID int64, in TransitionInput) (bool, error) {
	// 先查精确匹配（type_code 指定）
	var exists bool
	err := s.db.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM state_transitions
			WHERE project_id = $1 AND workspace_id = $2
			  AND from_state_id = $3 AND to_state_id = $4
			  AND (type_code = $5 OR type_code = 'all')
		)`, projectID, wsID, in.FromState, in.ToState, string(in.TypeCode)).Scan(&exists)
	if err != nil {
		return false, errs.ErrInternal.Wrap(err)
	}
	return exists, nil
}

func (s *StateService) checkRequiredFields(ctx context.Context, wsID int64, in TransitionInput) error {
	// 查询 from/to 状态名称
	fromState, err := s.GetStateByID(ctx, wsID, in.FromState)
	if err != nil {
		return err
	}
	toState, err := s.GetStateByID(ctx, wsID, in.ToState)
	if err != nil {
		return err
	}

	// 构造流转 key： "from_name -> to_name"
	transitionKey := fromState.Name + " -> " + toState.Name

	// 从内置规则查找 required_fields
	if fields, ok := RequiredFieldsForTransition[transitionKey]; ok {
		return validateFields(in.Context, fields)
	}

	return nil
}

// validateFields 校验必填字段。
func validateFields(ctx TransitionContext, fields []string) error {
	for _, f := range fields {
		switch f {
		case "root_cause_category":
			if ctx.RootCauseCategory == nil || *ctx.RootCauseCategory == "" {
				return errs.ErrValidation.WithDetails(errs.FieldDetail{
					Field: "root_cause_category", Reason: "流转到该状态时根因分类为必填",
				})
			}
		case "fix_version_id":
			if ctx.FixVersionID == nil {
				return errs.ErrValidation.WithDetails(errs.FieldDetail{
					Field: "fix_version_id", Reason: "流转到该状态时修复版本为必填",
				})
			}
		}
	}
	return nil
}

// 辅助占位（完整实现在 issue_service 中查 DB）
func fieldsFromDB(in interface{}) []string { return nil }

func validateFieldsFromTransitions(_ interface{}, ctx TransitionContext, stateName string) error {
	// 内置规则已在 checkRequiredFields 前置查过；DB 自定义规则在 IssueService 中补充
	_ = ctx
	_ = stateName
	return nil
}

// InitializeProjectStates 为项目初始化状态模板（项目创建时调用）。
// 将默认模板套装写入 states + state_transitions 表。
// 注：实际初始化逻辑已迁移至 ProjectInitService.InitializeForProject()。
func (s *StateService) InitializeProjectStates(ctx context.Context, wsID, projectID int64, nameIndex StateNameIndex) error {
	return nil
}

// StateGroupByID 根据 state ID 查询其 group。
func (s *StateService) StateGroupByID(ctx context.Context, stateID int64) (StateGroup, error) {
	var group StateGroup
	err := s.db.QueryRow(ctx, `SELECT "group" FROM states WHERE id = $1 AND deleted = false`, stateID).Scan(&group)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", errs.ErrNotFound
		}
		return "", errs.ErrInternal.Wrap(err)
	}
	return group, nil
}

// --- 状态 CRUD ---

// CreateStateInput 创建状态入参。
type CreateStateInput struct {
	WorkspaceID     int64
	ProjectID       int64
	Name            string
	Group           StateGroup
	Color           string
	Sequence        float64
	IsDefault       bool
	ApplicableTypes []string
	CreatedBy       int64
}

// CreateState 创建自定义状态。
func (s *StateService) CreateState(ctx context.Context, in CreateStateInput) (*State, error) {
	if in.Name == "" {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "name", Reason: "状态名不能为空"})
	}
	if in.Group == "" {
		in.Group = GroupBacklog
	}
	if in.Color == "" {
		in.Color = "#8DA2C2"
	}
	if len(in.ApplicableTypes) == 0 {
		in.ApplicableTypes = []string{"all"}
	}

	// 生成 Snowflake ID（简化：使用 pkg/snowflake 或数据库序列）
	id, err := s.nextID(ctx)
	if err != nil {
		return nil, err
	}

	_, err = s.db.Exec(ctx, `
		INSERT INTO states (id, workspace_id, project_id, name, "group", color, sequence, is_default, applicable_types, created_by, updated_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$10)`,
		id, in.WorkspaceID, in.ProjectID, in.Name, string(in.Group), in.Color, in.Sequence, in.IsDefault, in.ApplicableTypes, in.CreatedBy)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}

	return s.GetStateByID(ctx, in.WorkspaceID, id)
}

// UpdateStateInput 更新状态入参。
type UpdateStateInput struct {
	WorkspaceID     int64
	StateID         int64
	Name            *string
	Group           *StateGroup
	Color           *string
	Sequence        *float64
	IsDefault       *bool
	ApplicableTypes []string
	UpdatedBy       int64
}

// UpdateState 更新状态属性。
func (s *StateService) UpdateState(ctx context.Context, in UpdateStateInput) (*State, error) {
	// 构建动态 SQL：只更新非空字段
	var sets []string
	var args []any
	argIdx := 1

	if in.Name != nil {
		sets = append(sets, fmt.Sprintf("name = $%d", argIdx))
		args = append(args, *in.Name)
		argIdx++
	}
	if in.Group != nil {
		sets = append(sets, fmt.Sprintf(`"group" = $%d`, argIdx))
		args = append(args, string(*in.Group))
		argIdx++
	}
	if in.Color != nil {
		sets = append(sets, fmt.Sprintf("color = $%d", argIdx))
		args = append(args, *in.Color)
		argIdx++
	}
	if in.Sequence != nil {
		sets = append(sets, fmt.Sprintf("sequence = $%d", argIdx))
		args = append(args, *in.Sequence)
		argIdx++
	}
	if in.IsDefault != nil {
		sets = append(sets, fmt.Sprintf("is_default = $%d", argIdx))
		args = append(args, *in.IsDefault)
		argIdx++
	}
	if in.ApplicableTypes != nil {
		sets = append(sets, fmt.Sprintf("applicable_types = $%d", argIdx))
		args = append(args, in.ApplicableTypes)
		argIdx++
	}

	if len(sets) == 0 {
		return s.GetStateByID(ctx, in.WorkspaceID, in.StateID)
	}

	sets = append(sets, fmt.Sprintf("updated_by = $%d", argIdx))
	args = append(args, in.UpdatedBy)
	argIdx++

	sets = append(sets, fmt.Sprintf("updated_at = now()"))

	sql := fmt.Sprintf("UPDATE states SET %s WHERE id = $%d AND workspace_id = $%d AND deleted = false",
		strings.Join(sets, ", "), argIdx, argIdx+1)
	args = append(args, in.StateID, in.WorkspaceID)

	_, err := s.db.Exec(ctx, sql, args...)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	return s.GetStateByID(ctx, in.WorkspaceID, in.StateID)
}

// DeleteState 软删除状态（若有工作项在使用则拒绝）。
func (s *StateService) DeleteState(ctx context.Context, wsID, stateID int64) error {
	// 检查是否有工作项仍在此状态
	var inUse bool
	err := s.db.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM requirement WHERE state_id = $1 AND workspace_id = $2 AND deleted = false
			UNION ALL
			SELECT 1 FROM task WHERE state_id = $1 AND workspace_id = $2 AND deleted = false
			UNION ALL
			SELECT 1 FROM defect WHERE state_id = $1 AND workspace_id = $2 AND deleted = false
		)`, stateID, wsID).Scan(&inUse)
	if err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	if inUse {
		return errs.New("ISSUE.STATE_IN_USE", "仍有工作项处于此状态，无法删除", 409)
	}

	// 软删除状态 + 关联的流转规则
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, `UPDATE states SET deleted = true, updated_at = now() WHERE id = $1 AND workspace_id = $2`, stateID, wsID); err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	if _, err := tx.Exec(ctx, `UPDATE state_transitions SET deleted = true, updated_at = now() WHERE (from_state_id = $1 OR to_state_id = $1) AND workspace_id = $2`, stateID, wsID); err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	return tx.Commit(ctx)
}

// --- 流转规则 CRUD ---

// TransitionRule 流转规则（读出时附带状态名方便前端展示）。
type TransitionRule struct {
	ID              int64    `json:"id"`
	FromStateID     int64    `json:"from_state_id"`
	FromStateName   string   `json:"from_state_name,omitempty"`
	ToStateID       int64    `json:"to_state_id"`
	ToStateName     string   `json:"to_state_name,omitempty"`
	TypeCode        string   `json:"type_code"`
	RequiredFields  []string `json:"required_fields"`
}

// ListTransitions 列出项目的全部流转规则。
func (s *StateService) ListTransitions(ctx context.Context, wsID, projectID int64) ([]TransitionRule, error) {
	rows, err := s.db.Query(ctx, `
		SELECT t.id, t.from_state_id, f.name AS from_name, t.to_state_id, t.name AS to_name, t.type_code, t.required_fields
		FROM state_transitions t
		JOIN states f ON f.id = t.from_state_id AND f.deleted = false
		JOIN states u ON u.id = t.to_state_id AND u.deleted = false
		WHERE t.project_id = $1 AND t.workspace_id = $2 AND t.deleted = false
		ORDER BY f.sequence, u.sequence`, projectID, wsID)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()

	var rules []TransitionRule
	for rows.Next() {
		var r TransitionRule
		if err := rows.Scan(&r.ID, &r.FromStateID, &r.FromStateName, &r.ToStateID, &r.ToStateName, &r.TypeCode, &r.RequiredFields); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
		rules = append(rules, r)
	}
	return rules, rows.Err()
}

// AddTransitionInput 添加流转规则入参。
type AddTransitionInput struct {
	WorkspaceID    int64
	ProjectID      int64
	FromStateID    int64
	ToStateID      int64
	TypeCode       string
	RequiredFields []string
	CreatedBy      int64
}

// AddTransition 添加流转规则。
func (s *StateService) AddTransition(ctx context.Context, in AddTransitionInput) (*TransitionRule, error) {
	if in.FromStateID == in.ToStateID {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "to_state_id", Reason: "起止状态不能相同"})
	}
	id, err := s.nextID(ctx)
	if err != nil {
		return nil, err
	}
	requiredJSON := in.RequiredFields
	if requiredJSON == nil {
		requiredJSON = []string{}
	}

	_, err = s.db.Exec(ctx, `
		INSERT INTO state_transitions (id, workspace_id, project_id, type_code, from_state_id, to_state_id, required_fields, created_by, updated_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$8)
		ON CONFLICT (from_state_id, to_state_id, type_code) DO UPDATE SET deleted = false, required_fields = $7, updated_by = $8, updated_at = now()`,
		id, in.WorkspaceID, in.ProjectID, in.TypeCode, in.FromStateID, in.ToStateID, requiredJSON, in.CreatedBy)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}

	return &TransitionRule{
		ID:             id,
		FromStateID:    in.FromStateID,
		ToStateID:      in.ToStateID,
		TypeCode:       in.TypeCode,
		RequiredFields: requiredJSON,
	}, nil
}

// RemoveTransition 删除流转规则。
func (s *StateService) RemoveTransition(ctx context.Context, wsID, transitionID int64) error {
	tag, err := s.db.Exec(ctx, `
		UPDATE state_transitions SET deleted = true, updated_at = now()
		WHERE id = $1 AND workspace_id = $2`, transitionID, wsID)
	if err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	if tag.RowsAffected() == 0 {
		return errs.ErrNotFound
	}
	return nil
}

// nextID 生成下一个 Snowflake ID（简化实现，生产应使用 pkg/snowflake）。
func (s *StateService) nextID(ctx context.Context) (int64, error) {
	var id int64
	err := s.db.QueryRow(ctx, "SELECT nextval('states_id_seq')").Scan(&id)
	if err != nil {
		return 0, errs.ErrInternal.Wrap(err)
	}
	return id, nil
}

// CanTransitionQuick 快速判断是否允许流转（不校验字段，仅看规则）。
// 用于前端按钮的 disabled 状态展示。
func (s *StateService) CanTransitionQuick(ctx context.Context, wsID, projectID int64, fromStateID, toStateID int64, tc IssueTypeCode) (bool, error) {
	if fromStateID == toStateID {
		return false, nil
	}
	var exists bool
	err := s.db.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM state_transitions
			WHERE project_id = $1 AND workspace_id = $2
			  AND from_state_id = $3 AND to_state_id = $4
			  AND (type_code = $5 OR type_code = 'all')
		)`, projectID, wsID, fromStateID, toStateID, string(tc)).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("state check: %w", err)
	}
	return exists, nil
}
