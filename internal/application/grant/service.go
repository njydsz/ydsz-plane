// Package grant — 条件授权存储与生效计算。
package grant

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// Service 条件授权服务。
type Service struct {
	db *pgxpool.Pool
}

// NewService 构造服务。
func NewService(db *pgxpool.Pool) *Service { return &Service{db: db} }

// --- 错误 ---

var (
	// ErrRuleNotFound 规则不存在。
	ErrRuleNotFound = errors.New("grant rule not found")
	// ErrInvalidCondition 条件表达式非法。
	ErrInvalidCondition = errors.New("invalid condition expression")
)

// Create 创建规则。
func (s *Service) Create(ctx context.Context, in CreateRuleInput) (*Rule, error) {
	if strings.TrimSpace(in.Name) == "" {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "name", Reason: "required"})
	}
	if in.WorkspaceID <= 0 {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "workspace_id", Reason: "required"})
	}
	if len(in.Permissions) == 0 {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "permissions", Reason: "required"})
	}
	if err := Validate(in.Condition); err != nil {
		return nil, ErrInvalidCondition
	}
	if in.Priority <= 0 {
		in.Priority = 100
	}

	var r Rule
	err := s.db.QueryRow(ctx, `
		INSERT INTO grant_rules
			(id, workspace_id, name, description, role_slugs, permissions, scope,
			 condition, priority, created_by)
		VALUES ($1,$2,$3,$4,$5::text[],$6::text[],$7,$8::jsonb,$9,$10)
		RETURNING id, workspace_id, name, COALESCE(description,''),
			role_slugs, permissions, COALESCE(scope,''), condition, enabled, priority,
			created_by, created_at, updated_at`,
		genID(), in.WorkspaceID, in.Name, nullIfEmpty(in.Description),
		strSlice(in.RoleSlugs), strSlice(in.Permissions), nullIfEmpty(in.Scope),
		string(in.Condition), in.Priority, in.CreatedBy,
	).Scan(&r.ID, &r.WorkspaceID, &r.Name, &r.Description, &r.RoleSlugs,
		&r.Permissions, &r.Scope, &r.Condition, &r.Enabled, &r.Priority,
		&r.CreatedBy, &r.CreatedAt, &r.UpdatedAt)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	return &r, nil
}

// GetByID 查询规则。
func (s *Service) GetByID(ctx context.Context, id, workspaceID int64) (*Rule, error) {
	var r Rule
	err := s.db.QueryRow(ctx, `
		SELECT id, workspace_id, name, COALESCE(description,''),
			role_slugs, permissions, COALESCE(scope,''), condition, enabled, priority,
			created_by, created_at, updated_at
		FROM grant_rules WHERE id = $1 AND workspace_id = $2`, id, workspaceID,
	).Scan(&r.ID, &r.WorkspaceID, &r.Name, &r.Description, &r.RoleSlugs,
		&r.Permissions, &r.Scope, &r.Condition, &r.Enabled, &r.Priority,
		&r.CreatedBy, &r.CreatedAt, &r.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrRuleNotFound
	}
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	return &r, nil
}

// List 列出工作空间规则。
func (s *Service) List(ctx context.Context, workspaceID int64) ([]*Rule, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, workspace_id, name, COALESCE(description,''),
			role_slugs, permissions, COALESCE(scope,''), condition, enabled, priority,
			created_by, created_at, updated_at
		FROM grant_rules WHERE workspace_id = $1
		ORDER BY priority ASC, created_at DESC`, workspaceID)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()

	var out []*Rule
	for rows.Next() {
		var r Rule
		if err := rows.Scan(&r.ID, &r.WorkspaceID, &r.Name, &r.Description, &r.RoleSlugs,
			&r.Permissions, &r.Scope, &r.Condition, &r.Enabled, &r.Priority,
			&r.CreatedBy, &r.CreatedAt, &r.UpdatedAt); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
		out = append(out, &r)
	}
	return out, rows.Err()
}

// Update 更新规则。
func (s *Service) Update(ctx context.Context, id, workspaceID int64, in UpdateRuleInput) (*Rule, error) {
	sets := []string{}
	args := []any{}
	add := func(expr string, v any) {
		args = append(args, v)
		sets = append(sets, expr+"$"+itoa(len(args)))
	}
	if in.Name != nil {
		add("name = ", *in.Name)
	}
	if in.Description != nil {
		add("description = ", nullIfEmpty(*in.Description))
	}
	if in.RoleSlugs != nil {
		add("role_slugs = ", strSlice(in.RoleSlugs))
	}
	if in.Permissions != nil {
		add("permissions = ", strSlice(in.Permissions))
	}
	if in.Scope != nil {
		add("scope = ", nullIfEmpty(*in.Scope))
	}
	if in.Condition != nil {
		if err := Validate(in.Condition); err != nil {
			return nil, ErrInvalidCondition
		}
		add("condition = ", string(in.Condition))
	}
	if in.Enabled != nil {
		add("enabled = ", *in.Enabled)
	}
	if in.Priority != nil {
		add("priority = ", *in.Priority)
	}
	if len(sets) == 0 {
		return s.GetByID(ctx, id, workspaceID)
	}
	sets = append(sets, "updated_at = now()")
	args = append(args, id, workspaceID)

	var r Rule
	err := s.db.QueryRow(ctx, `
		UPDATE grant_rules SET `+strings.Join(sets, ", ")+`
		WHERE id = $`+itoa(len(args)-1)+` AND workspace_id = $`+itoa(len(args))+`
		RETURNING id, workspace_id, name, COALESCE(description,''),
			role_slugs, permissions, COALESCE(scope,''), condition, enabled, priority,
			created_by, created_at, updated_at`, args...,
	).Scan(&r.ID, &r.WorkspaceID, &r.Name, &r.Description, &r.RoleSlugs,
		&r.Permissions, &r.Scope, &r.Condition, &r.Enabled, &r.Priority,
		&r.CreatedBy, &r.CreatedAt, &r.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrRuleNotFound
	}
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	return &r, nil
}

// Delete 删除规则。
func (s *Service) Delete(ctx context.Context, id, workspaceID int64) error {
	tag, err := s.db.Exec(ctx,
		`DELETE FROM grant_rules WHERE id = $1 AND workspace_id = $2`, id, workspaceID)
	if err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	if tag.RowsAffected() == 0 {
		return ErrRuleNotFound
	}
	return nil
}

// --- 生效计算 ---

// EffectivePermissions 计算给定角色在上下文下通过条件规则获得的额外权限。
func (s *Service) EffectivePermissions(ctx context.Context, workspaceID int64, roleSlug string, attrs Context) ([]string, error) {
	rules, err := s.List(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	out := []string{}
	for _, r := range rules {
		if !r.Enabled {
			continue
		}
		if len(r.RoleSlugs) > 0 && !containsFold(r.RoleSlugs, roleSlug) {
			continue
		}
		var expr Expr
		if err := json.Unmarshal(r.Condition, &expr); err != nil {
			continue
		}
		ok, err := Evaluate(&expr, attrs)
		if err != nil || !ok {
			continue
		}
		out = append(out, r.Permissions...)
	}
	return dedup(out), nil
}

// Check 判断给定角色在上下文下是否拥有某权限（仅条件授权维度）。
func (s *Service) Check(ctx context.Context, workspaceID int64, roleSlug, perm string, attrs Context) (bool, error) {
	perms, err := s.EffectivePermissions(ctx, workspaceID, roleSlug, attrs)
	if err != nil {
		return false, err
	}
	for _, p := range perms {
		if p == perm {
			return true, nil
		}
	}
	return false, nil
}

// --- helpers ---

func nullIfEmpty(s string) any {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return s
}

func strSlice(v []string) any {
	if len(v) == 0 {
		return nil
	}
	return v
}

func containsFold(list []string, target string) bool {
	for _, v := range list {
		if strings.EqualFold(strings.TrimSpace(v), target) {
			return true
		}
	}
	return false
}

func dedup(in []string) []string {
	seen := map[string]bool{}
	out := []string{}
	for _, v := range in {
		v = strings.TrimSpace(v)
		if v == "" || seen[v] {
			continue
		}
		seen[v] = true
		out = append(out, v)
	}
	return out
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
