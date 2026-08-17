// Package issue — 项目 Issue 基础设施初始化（创建项目状态模板 + 流转规则）。
package issue

import (
	"context"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// ProjectInitService 在项目创建时初始化 issue 域基础设施。
type ProjectInitService struct {
	db *pgxpool.Pool
}

// NewProjectInitService 创建初始化服务。
func NewProjectInitService(db *pgxpool.Pool) *ProjectInitService {
	return &ProjectInitService{db: db}
}

// InitializeForProject 为项目初始化状态模板集 + 流转规则。
// 由 ProjectService.Create 在事务成功后调用（也可由独立 hook 调用）。
//
// 参数 templateCode 指定使用的预置模板：
//   - "agile":     DevFlow + DefectFlow（Scrum 风格）
//   - "waterfall": DevFlow + RequirementFlow（V 模型）
//   - "generic":   DevFlow 精简集（看板）
//
// 默认值 "generic" 保证向后兼容。
func (s *ProjectInitService) InitializeForProject(ctx context.Context, wsID, projectID int64, templateCode ...string) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, "SELECT set_config('app.workspace_id', $1, true)", strconv.FormatInt(wsID, 10)); err != nil {
		return errs.ErrInternal.Wrap(err)
	}

	tplCode := ProjectTemplateCode(TemplateGeneric)
	if len(templateCode) > 0 && ValidateTemplateCode(templateCode[0]) {
		tplCode = ProjectTemplateCode(templateCode[0])
	}
	tpl := ProjectTemplateByCode(string(tplCode))

	// 初始化任务研发流状态（所有模板通用）
	taskDevStateIDs, err := s.insertTemplateSet(ctx, tx, wsID, projectID, TaskDevFlowStates, "task_dev_flow")
	if err != nil {
		return err
	}
	if err := s.insertTransitions(ctx, tx, wsID, projectID, string(TypeTask),
		BuiltInTransitions["dev_flow"], taskDevStateIDs); err != nil {
		return err
	}

	// 初始化需求研发流状态（所有模板通用）
	reqDevStateIDs, err := s.insertTemplateSet(ctx, tx, wsID, projectID, RequirementDevFlowStates, "requirement_dev_flow")
	if err != nil {
		return err
	}
	if err := s.insertTransitions(ctx, tx, wsID, projectID, string(TypeRequirement),
		BuiltInTransitions["dev_flow"], reqDevStateIDs); err != nil {
		return err
	}

	// 按模板可选注入缺陷流
	if tpl.ApplyDefectFlow {
		defectStateIDs, err := s.insertTemplateSet(ctx, tx, wsID, projectID, DefectFlowStates, "defect_flow")
		if err != nil {
			return err
		}
		if err := s.insertTransitions(ctx, tx, wsID, projectID, string(TypeDefect),
			BuiltInTransitions["defect_flow"], defectStateIDs); err != nil {
			return err
		}
	}

	// 按模板可选注入需求评审流
	if tpl.ApplyRequirementFlow {
		reqReviewStateIDs, err := s.insertTemplateSet(ctx, tx, wsID, projectID, RequirementReviewFlowStates, "requirement_review_flow")
		if err != nil {
			return err
		}
		if err := s.insertTransitions(ctx, tx, wsID, projectID, string(TypeRequirement),
			BuiltInTransitions["requirement_flow"], reqReviewStateIDs); err != nil {
			return err
		}
	}

	// 按模板可选注入史诗流（默认所有模板均启用，作为顶层容器）
	if tpl.ApplyEpicFlow {
		epicStateIDs, err := s.insertTemplateSet(ctx, tx, wsID, projectID, EpicFlowStates, "epic_flow")
		if err != nil {
			return err
		}
		if err := s.insertTransitions(ctx, tx, wsID, projectID, string(TypeEpic),
			BuiltInTransitions["epic_flow"], epicStateIDs); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

// insertTemplateSet 插入状态模板并返回 name → ID 映射。
// templateSetLabel 写入 states.template_set 列，用于标识状态归属的模板分类。
func (s *ProjectInitService) insertTemplateSet(ctx context.Context, tx pgx.Tx,
	wsID, projectID int64, states []State, templateSetLabel string) (map[string]int64, error) {

	ids := make(map[string]int64, len(states))
	for _, st := range states {
		var id int64
		// 把适用的工作项类型数组转为PostgreSQL数组格式
		applicableTypes := "{"
		for i, t := range st.ApplicableTypes {
			if i > 0 {
				applicableTypes += ","
			}
			applicableTypes += t
		}
		applicableTypes += "}"
		err := tx.QueryRow(ctx, `
			INSERT INTO states (workspace_id, project_id, name, "group", color, sequence, is_default, template_set, applicable_types)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
			RETURNING id`,
			wsID, projectID, st.Name, string(st.Group), st.Color, st.Sequence, st.IsDefault,
			templateSetLabel, applicableTypes).Scan(&id)
		if err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
		ids[st.Name] = id
	}
	return ids, nil
}

// insertTransitions 根据模板名流转规则写入 DB。
func (s *ProjectInitService) insertTransitions(ctx context.Context, tx pgx.Tx,
	wsID, projectID int64, typeCode string, transitions []TransitionKey,
	nameToID map[string]int64) error {

	for _, t := range transitions {
		fromID := nameToID[t.From]
		if t.From == "*" {
			continue // 通配符规则：项目级任意状态都可以流转到任意 to（由应用层处理）
		}
		toID := nameToID[t.To]
		if fromID == 0 || toID == 0 {
			continue // 该模板中没有的状态
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO state_transitions (workspace_id, project_id, type_code, from_state_id, to_state_id)
			VALUES ($1,$2,$3,$4,$5)
			ON CONFLICT (project_id, type_code, from_state_id, to_state_id) DO NOTHING`,
			wsID, projectID, typeCode, fromID, toID); err != nil {
			return errs.ErrInternal.Wrap(err)
		}
	}
	return nil
}
