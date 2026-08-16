// Package issue — 状态机服务（状态查询 + 流转校验）。
package issue

import (
	"context"
	"errors"
	"fmt"

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
