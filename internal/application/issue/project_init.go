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
func (s *ProjectInitService) InitializeForProject(ctx context.Context, wsID, projectID int64) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, "SELECT set_config('app.workspace_id', $1, true)", strconv.FormatInt(wsID, 10)); err != nil {
		return errs.ErrInternal.Wrap(err)
	}

	// 初始化研发流状态
	devStateIDs, err := s.insertTemplateSet(ctx, tx, wsID, projectID, DevFlowStates)
	if err != nil {
		return err
	}
	// 缺陷流
	defectStateIDs, err := s.insertTemplateSet(ctx, tx, wsID, projectID, DefectFlowStates)
	if err != nil {
		return err
	}

	// 写流转规则 - dev_flow（通用，type_code=all）
	if err := s.insertTransitions(ctx, tx, wsID, projectID, "all",
		BuiltInTransitions["dev_flow"], devStateIDs); err != nil {
		return err
	}
	// defect_flow
	if err := s.insertTransitions(ctx, tx, wsID, projectID, string(TypeDefect),
		BuiltInTransitions["defect_flow"], defectStateIDs); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// insertTemplateSet 插入状态模板并返回 name → ID 映射。
func (s *ProjectInitService) insertTemplateSet(ctx context.Context, tx pgx.Tx,
	wsID, projectID int64, states []State) (map[string]int64, error) {

	ids := make(map[string]int64, len(states))
	for _, st := range states {
		var id int64
		err := tx.QueryRow(ctx, `
			INSERT INTO states (workspace_id, project_id, name, "group", color, sequence, is_default, template_set)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
			RETURNING id`,
			wsID, projectID, st.Name, string(st.Group), st.Color, st.Sequence, st.IsDefault,
			"dev_flow").Scan(&id)
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
