// Package workflow — 工作流与审批应用服务。
//
// 职责：
//   - 项目级工作流定义（Workflow）与状态流转规则（WorkflowTransition）的 CRUD；
//   - 流转合法性校验（仅允许已定义的 from → to）；
//   - 审批门禁（requires_approval 时执行流转前必须有已批准记录）；
//   - 审批记录的提交 / 批准 / 驳回与待办列表。
package workflow

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// Service 工作流应用服务。
type Service struct {
	db *pgxpool.Pool
}

// NewService 创建服务。
func NewService(db *pgxpool.Pool) *Service {
	return &Service{db: db}
}

// --- 错误 ---

var (
	// ErrWorkflowNotFound 工作流不存在。
	ErrWorkflowNotFound = errors.New("workflow not found")
	// ErrTransitionNotFound 流转规则不存在。
	ErrTransitionNotFound = errors.New("workflow transition not found")
	// ErrApprovalNotFound 审批记录不存在。
	ErrApprovalNotFound = errors.New("workflow approval not found")
	// ErrTransitionNotAllowed 该流转未被工作流定义允许。
	ErrTransitionNotAllowed = errors.New("transition not allowed by workflow")
	// ErrApprovalRequired 流转需要审批但尚无已批准记录。
	ErrApprovalRequired = errors.New("approval required before transition")
	// ErrApprovalAlreadyResolved 审批已终态，不可重复处理。
	ErrApprovalAlreadyResolved = errors.New("approval already resolved")
	// ErrNotApprover 当前用户不在审批人范围内。
	ErrNotApprover = errors.New("user is not an approver")
)

// --- Workflow CRUD ---

// Create 创建工作流。
//
// 若指定为默认工作流，则在同一项目内将其它工作流 is_default 置为 false。
func (s *Service) Create(ctx context.Context, in CreateWorkflowInput) (*Workflow, error) {
	if in.ProjectID <= 0 {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "project_id", Reason: "required"})
	}
	if in.Name == "" {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "name", Reason: "required"})
	}

	var wf Workflow
	err := s.db.QueryRow(ctx, `
		INSERT INTO workflows (id, project_id, name, description, is_default, created_by, updated_by)
		VALUES ($1,$2,$3,$4,$5,$6,$6)
		RETURNING id, project_id, name, description, is_default, created_at, updated_at`,
		genID(), in.ProjectID, in.Name, nullIfEmpty(in.Description), in.IsDefault, in.CreatedBy,
	).Scan(&wf.ID, &wf.ProjectID, &wf.Name, &wf.Description, &wf.IsDefault, &wf.CreatedAt, &wf.UpdatedAt)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	return &wf, nil
}

// GetByID 查询工作流。
func (s *Service) GetByID(ctx context.Context, id int64) (*Workflow, error) {
	var wf Workflow
	err := s.db.QueryRow(ctx, `
		SELECT id, project_id, name, description, is_default, created_at, updated_at
		FROM workflows WHERE id = $1 AND deleted_at IS NULL`,
		id,
	).Scan(&wf.ID, &wf.ProjectID, &wf.Name, &wf.Description, &wf.IsDefault, &wf.CreatedAt, &wf.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrWorkflowNotFound
	}
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	return &wf, nil
}

// List 列出项目的工作流。
func (s *Service) List(ctx context.Context, opts ListWorkflowsOptions) ([]*Workflow, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, project_id, name, description, is_default, created_at, updated_at
		FROM workflows
		WHERE project_id = $1 AND deleted_at IS NULL
		ORDER BY is_default DESC, created_at DESC
		LIMIT $2 OFFSET $3`,
		opts.ProjectID, opts.Limit, opts.Offset,
	)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()

	var out []*Workflow
	for rows.Next() {
		var wf Workflow
		if err := rows.Scan(&wf.ID, &wf.ProjectID, &wf.Name, &wf.Description, &wf.IsDefault, &wf.CreatedAt, &wf.UpdatedAt); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
		out = append(out, &wf)
	}
	return out, rows.Err()
}

// GetDetail 查询工作流详情（含全部流转规则）。
func (s *Service) GetDetail(ctx context.Context, id int64) (*WorkflowDetail, error) {
	wf, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	transitions, err := s.listTransitions(ctx, id)
	if err != nil {
		return nil, err
	}
	return &WorkflowDetail{Workflow: wf, Transitions: transitions}, nil
}

// Update 更新工作流。
func (s *Service) Update(ctx context.Context, id int64, in UpdateWorkflowInput) (*Workflow, error) {
	sets := []string{}
	args := []any{}
	add := func(expr string, val any) {
		args = append(args, val)
		sets = append(sets, expr+"$"+itoa(len(args)))
	}

	if in.Name != nil && *in.Name != "" {
		add("name = ", *in.Name)
	}
	if in.Description != nil {
		add("description = ", nullIfEmpty(*in.Description))
	}
	if in.IsDefault != nil {
		add("is_default = ", *in.IsDefault)
	}
	if len(sets) == 0 {
		return s.GetByID(ctx, id)
	}
	args = append(args, id)
	query := "UPDATE workflows SET " + joinComma(sets) + ", updated_at = now() WHERE id = $" + itoa(len(args)) + " AND deleted_at IS NULL"
	var wf Workflow
	err := s.db.QueryRow(ctx, query, args...).
		Scan(&wf.ID, &wf.ProjectID, &wf.Name, &wf.Description, &wf.IsDefault, &wf.CreatedAt, &wf.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrWorkflowNotFound
	}
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	return &wf, nil
}

// Delete 软删除工作流及其流转规则。
func (s *Service) Delete(ctx context.Context, id int64) error {
	tag, err := s.db.Exec(ctx, `
		UPDATE workflows SET deleted_at = now(), updated_at = now()
		WHERE id = $1 AND deleted_at IS NULL`, id)
	if err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	if tag.RowsAffected() == 0 {
		return ErrWorkflowNotFound
	}
	_, _ = s.db.Exec(ctx, `DELETE FROM workflow_transitions WHERE workflow_id = $1`, id)
	return nil
}

// --- Transition CRUD ---

// AddTransition 新增流转规则。
func (s *Service) AddTransition(ctx context.Context, in CreateTransitionInput) (*WorkflowTransition, error) {
	if in.WorkflowID <= 0 || in.FromStateID <= 0 || in.ToStateID <= 0 {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "state_id", Reason: "from/to required"})
	}
	if in.FromStateID == in.ToStateID {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "to_state_id", Reason: "must differ from from_state_id"})
	}
	if in.Conditions != "" {
		if !json.Valid([]byte(in.Conditions)) {
			return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "conditions", Reason: "must be valid JSON"})
		}
	}

	var tr WorkflowTransition
	err := s.db.QueryRow(ctx, `
		INSERT INTO workflow_transitions
			(id, workflow_id, from_state_id, to_state_id, name, requires_approval,
			 approver_role, approver_ids, conditions)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		RETURNING id, workflow_id, from_state_id, to_state_id, name, requires_approval,
			approver_role, approver_ids, conditions, created_at`,
		genID(), in.WorkflowID, in.FromStateID, in.ToStateID, nullIfEmpty(in.Name),
		in.RequiresApproval, nullIfEmpty(in.ApproverRole), int64Slice(in.ApproverIDs),
		nullIfEmpty(in.Conditions),
	).Scan(&tr.ID, &tr.WorkflowID, &tr.FromStateID, &tr.ToStateID, &tr.Name,
		&tr.RequiresApproval, &tr.ApproverRole, &tr.ApproverIDs, &tr.Conditions, &tr.CreatedAt)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	return &tr, nil
}

// listTransitions 列出工作流的流转规则。
func (s *Service) listTransitions(ctx context.Context, workflowID int64) ([]WorkflowTransition, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, workflow_id, from_state_id, to_state_id, name, requires_approval,
			COALESCE(approver_role, ''), COALESCE(approver_ids, '{}'), COALESCE(conditions, ''), created_at
		FROM workflow_transitions
		WHERE workflow_id = $1
		ORDER BY created_at ASC`, workflowID)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()

	var out []WorkflowTransition
	for rows.Next() {
		var tr WorkflowTransition
		if err := rows.Scan(&tr.ID, &tr.WorkflowID, &tr.FromStateID, &tr.ToStateID, &tr.Name,
			&tr.RequiresApproval, &tr.ApproverRole, &tr.ApproverIDs, &tr.Conditions, &tr.CreatedAt); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
		out = append(out, tr)
	}
	return out, rows.Err()
}

// DeleteTransition 删除流转规则。
func (s *Service) DeleteTransition(ctx context.Context, id int64) error {
	tag, err := s.db.Exec(ctx, `DELETE FROM workflow_transitions WHERE id = $1`, id)
	if err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	if tag.RowsAffected() == 0 {
		return ErrTransitionNotFound
	}
	return nil
}

// --- 校验与执行 ---

// ValidateTransition 校验 from → to 流转是否被允许，并返回是否需要审批。
func (s *Service) ValidateTransition(ctx context.Context, in ValidateTransitionInput) (*ValidateResult, error) {
	var tr WorkflowTransition
	err := s.db.QueryRow(ctx, `
		SELECT id, workflow_id, from_state_id, to_state_id, name, requires_approval,
			COALESCE(approver_role, ''), COALESCE(approver_ids, '{}'), COALESCE(conditions, ''), created_at
		FROM workflow_transitions
		WHERE workflow_id = $1 AND from_state_id = $2 AND to_state_id = $3`,
		in.WorkflowID, in.FromStateID, in.ToStateID,
	).Scan(&tr.ID, &tr.WorkflowID, &tr.FromStateID, &tr.ToStateID, &tr.Name,
		&tr.RequiresApproval, &tr.ApproverRole, &tr.ApproverIDs, &tr.Conditions, &tr.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return &ValidateResult{Allowed: false, Reason: "该流转未被工作流定义允许"}, nil
	}
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}

	res := &ValidateResult{Allowed: true, RequiresApproval: tr.RequiresApproval}
	if tr.RequiresApproval {
		// 若已存在该工作项针对该流转的已批准记录，则视为满足门禁。
		ok, err := s.hasApprovedApproval(ctx, tr.ID, in.WorkItemID)
		if err != nil {
			return nil, err
		}
		if !ok {
			res.Reason = "该流转需要审批后方可执行"
		}
	}
	return res, nil
}

// hasApprovedApproval 判断工作项是否已有针对指定流转的已批准记录。
func (s *Service) hasApprovedApproval(ctx context.Context, transitionID, workItemID int64) (bool, error) {
	var exists bool
	err := s.db.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM workflow_approvals
			WHERE transition_id = $1 AND work_item_id = $2 AND status = 'approved'
		)`, transitionID, workItemID).Scan(&exists)
	if err != nil {
		return false, errs.ErrInternal.Wrap(err)
	}
	return exists, nil
}

// --- 审批 ---

// SubmitApproval 提交审批申请。
func (s *Service) SubmitApproval(ctx context.Context, in SubmitApprovalInput) (*WorkflowApproval, error) {
	if in.WorkItemType == "" || !IsValidWorkItemType(in.WorkItemType) {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "work_item_type", Reason: "invalid value"})
	}
	if in.ApproverID <= 0 {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "approver_id", Reason: "required"})
	}

	var ap WorkflowApproval
	err := s.db.QueryRow(ctx, `
		INSERT INTO workflow_approvals
			(id, transition_id, work_item_id, work_item_type, requester_id, approver_id, status, comment)
		VALUES ($1,$2,$3,$4,$5,$6,'pending',$7)
		RETURNING id, transition_id, work_item_id, work_item_type, requester_id,
			approver_id, status, COALESCE(comment,''), created_at, resolved_at`,
		genID(), in.TransitionID, in.WorkItemID, in.WorkItemType, in.RequesterID,
		in.ApproverID, nullIfEmpty(in.Comment),
	).Scan(&ap.ID, &ap.TransitionID, &ap.WorkItemID, &ap.WorkItemType, &ap.RequesterID,
		&ap.ApproverID, &ap.Status, &ap.Comment, &ap.CreatedAt, &ap.ResolvedAt)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	return &ap, nil
}

// ResolveApproval 处理审批（批准 / 驳回）。
func (s *Service) ResolveApproval(ctx context.Context, in ResolveApprovalInput, approve bool) (*WorkflowApproval, error) {
	status := ApprovalRejected
	if approve {
		status = ApprovalApproved
	}
	var ap WorkflowApproval
	err := s.db.QueryRow(ctx, `
		UPDATE workflow_approvals
		SET status = $1, approver_id = $2, comment = COALESCE(NULLIF($3, ''), comment),
		    resolved_at = now()
		WHERE id = $4 AND status = 'pending'
		RETURNING id, transition_id, work_item_id, work_item_type, requester_id,
			approver_id, status, COALESCE(comment,''), created_at, resolved_at`,
		status, in.ApproverID, in.Comment, in.ApprovalID,
	).Scan(&ap.ID, &ap.TransitionID, &ap.WorkItemID, &ap.WorkItemType, &ap.RequesterID,
		&ap.ApproverID, &ap.Status, &ap.Comment, &ap.CreatedAt, &ap.ResolvedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrApprovalAlreadyResolved
	}
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	return &ap, nil
}

// ListPendingApprovals 列出工作空间内待用户处理的审批。
func (s *Service) ListPendingApprovals(ctx context.Context, opts ListPendingApprovalsOptions) ([]*WorkflowApproval, error) {
	rows, err := s.db.Query(ctx, `
		SELECT a.id, a.transition_id, a.work_item_id, a.work_item_type,
		       a.requester_id, a.approver_id, a.status, COALESCE(a.comment,''),
		       a.created_at, a.resolved_at
		FROM workflow_approvals a
		JOIN workflow_transitions t ON t.id = a.transition_id
		JOIN workflows w ON w.id = t.workflow_id
		WHERE w.project_id IN (SELECT id FROM projects WHERE workspace_id = $1)
		  AND a.status = 'pending' AND a.approver_id = $2
		ORDER BY a.created_at DESC
		LIMIT $3 OFFSET $4`,
		opts.WorkspaceID, opts.UserID, opts.Limit, opts.Offset)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()

	var out []*WorkflowApproval
	for rows.Next() {
		var ap WorkflowApproval
		if err := rows.Scan(&ap.ID, &ap.TransitionID, &ap.WorkItemID, &ap.WorkItemType,
			&ap.RequesterID, &ap.ApproverID, &ap.Status, &ap.Comment,
			&ap.CreatedAt, &ap.ResolvedAt); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
		out = append(out, &ap)
	}
	return out, rows.Err()
}

// GetTransitionHistory 查询工作项的流转历史。
func (s *Service) GetTransitionHistory(ctx context.Context, workItemID int64) ([]TransitionHistory, error) {
	rows, err := s.db.Query(ctx, `
		SELECT a.id, a.transition_id, t.from_state_id, t.to_state_id,
		       COALESCE(t.name,''), a.status, a.requester_id, a.approver_id,
		       COALESCE(a.comment,''), a.created_at, a.resolved_at
		FROM workflow_approvals a
		JOIN workflow_transitions t ON t.id = a.transition_id
		WHERE a.work_item_id = $1
		ORDER BY a.created_at DESC`, workItemID)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()

	var out []TransitionHistory
	for rows.Next() {
		var h TransitionHistory
		if err := rows.Scan(&h.ApprovalID, &h.TransitionID, &h.FromStateID, &h.ToStateID,
			&h.Name, &h.Status, &h.RequesterID, &h.ApproverID, &h.Comment,
			&h.CreatedAt, &h.ResolvedAt); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
		out = append(out, h)
	}
	return out, rows.Err()
}

// --- helpers ---

func nullIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func int64Slice(ids []int64) any {
	if len(ids) == 0 {
		return nil
	}
	return ids
}

func itoa(n int) string {
	if n < 10 {
		return string(rune('0' + n))
	}
	buf := []byte{}
	for n > 0 {
		buf = append([]byte{byte('0' + n%10)}, buf...)
		n /= 10
	}
	return string(buf)
}

func joinComma(parts []string) string {
	out := ""
	for i, p := range parts {
		if i > 0 {
			out += ", "
		}
		out += p
	}
	return out
}

var _ = time.Now
