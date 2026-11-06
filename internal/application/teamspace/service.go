// Package teamspace — 团队空间应用服务。
package teamspace

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// Service 团队空间应用服务。
type Service struct {
	db *pgxpool.Pool
}

// NewService 创建服务。
func NewService(db *pgxpool.Pool) *Service {
	return &Service{db: db}
}

// --- 错误 ---

var (
	// ErrTeamspaceNotFound 团队空间不存在。
	ErrTeamspaceNotFound = errors.New("teamspace not found")
	// ErrTeamspaceArchived 团队空间已归档。
	ErrTeamspaceArchived = errors.New("teamspace is archived")
)

// --- CRUD ---

// Create 创建团队空间（含成员与项目关联）。
func (s *Service) Create(ctx context.Context, in CreateTeamspaceInput) (*Teamspace, error) {
	if in.WorkspaceID <= 0 {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "workspace_id", Reason: "required"})
	}
	if strings.TrimSpace(in.Name) == "" {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "name", Reason: "required"})
	}

	var ts Teamspace
	err := s.db.QueryRow(ctx, `
		INSERT INTO teamspaces (id, workspace_id, name, description, color, lead_id, created_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		RETURNING id, workspace_id, name, COALESCE(description,''), COALESCE(color,''),
			lead_id, created_by, archived_at, created_at, updated_at`,
		genID(), in.WorkspaceID, in.Name, nullIfEmpty(in.Description),
		nullIfEmpty(in.Color), in.LeadID, in.CreatedBy,
	).Scan(&ts.ID, &ts.WorkspaceID, &ts.Name, &ts.Description, &ts.Color,
		&ts.LeadID, &ts.CreatedBy, &ts.ArchivedAt, &ts.CreatedAt, &ts.UpdatedAt)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}

	// 成员关联：lead 优先，其余为 member。
	seen := map[int64]bool{}
	for _, uid := range append([]int64{in.LeadID}, in.MemberIDs...) {
		if uid <= 0 || seen[uid] {
			continue
		}
		seen[uid] = true
		role := string(MemberRoleMember)
		if uid == in.LeadID {
			role = string(MemberRoleLead)
		}
		if _, err := s.db.Exec(ctx, `
			INSERT INTO teamspace_members (teamspace_id, user_id, role)
			VALUES ($1,$2,$3) ON CONFLICT DO NOTHING`, ts.ID, uid, role); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
	}

	// 项目关联。
	for _, pid := range in.ProjectIDs {
		if pid <= 0 {
			continue
		}
		if _, err := s.db.Exec(ctx, `
			INSERT INTO teamspace_projects (teamspace_id, project_id)
			VALUES ($1,$2) ON CONFLICT DO NOTHING`, ts.ID, pid); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
	}
	return &ts, nil
}

// GetByID 查询团队空间。
func (s *Service) GetByID(ctx context.Context, id int64) (*Teamspace, error) {
	var ts Teamspace
	err := s.db.QueryRow(ctx, `
		SELECT id, workspace_id, name, COALESCE(description,''), COALESCE(color,''),
			lead_id, created_by, archived_at, created_at, updated_at
		FROM teamspaces WHERE id = $1`, id,
	).Scan(&ts.ID, &ts.WorkspaceID, &ts.Name, &ts.Description, &ts.Color,
		&ts.LeadID, &ts.CreatedBy, &ts.ArchivedAt, &ts.CreatedAt, &ts.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrTeamspaceNotFound
	}
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	return &ts, nil
}

// List 列出工作空间的团队空间。
func (s *Service) List(ctx context.Context, opts ListTeamspacesOptions) ([]*Teamspace, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, workspace_id, name, COALESCE(description,''), COALESCE(color,''),
			lead_id, created_by, archived_at, created_at, updated_at
		FROM teamspaces
		WHERE workspace_id = $1
		ORDER BY archived_at NULLS FIRST, created_at DESC
		LIMIT $2 OFFSET $3`,
		opts.WorkspaceID, opts.Limit, opts.Offset)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()

	var out []*Teamspace
	for rows.Next() {
		var ts Teamspace
		if err := rows.Scan(&ts.ID, &ts.WorkspaceID, &ts.Name, &ts.Description, &ts.Color,
			&ts.LeadID, &ts.CreatedBy, &ts.ArchivedAt, &ts.CreatedAt, &ts.UpdatedAt); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
		out = append(out, &ts)
	}
	return out, rows.Err()
}

// Update 更新团队空间。
func (s *Service) Update(ctx context.Context, id int64, in UpdateTeamspaceInput) (*Teamspace, error) {
	sets := []string{}
	args := []any{}
	add := func(expr string, val any) {
		args = append(args, val)
		sets = append(sets, expr+"$"+itoa(len(args)))
	}
	if in.Name != nil && strings.TrimSpace(*in.Name) != "" {
		add("name = ", *in.Name)
	}
	if in.Description != nil {
		add("description = ", nullIfEmpty(*in.Description))
	}
	if in.Color != nil {
		add("color = ", nullIfEmpty(*in.Color))
	}
	if in.LeadID != nil && *in.LeadID > 0 {
		add("lead_id = ", *in.LeadID)
	}
	if len(sets) == 0 {
		return s.GetByID(ctx, id)
	}
	args = append(args, id)
	query := "UPDATE teamspaces SET " + joinComma(sets) + ", updated_at = now() WHERE id = $" + itoa(len(args))
	if _, err := s.db.Exec(ctx, query, args...); err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	return s.GetByID(ctx, id)
}

// Archive 归档团队空间。
func (s *Service) Archive(ctx context.Context, id int64) error {
	tag, err := s.db.Exec(ctx, `
		UPDATE teamspaces SET archived_at = now(), updated_at = now()
		WHERE id = $1 AND archived_at IS NULL`, id)
	if err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	if tag.RowsAffected() == 0 {
		return ErrTeamspaceNotFound
	}
	return nil
}

// Restore 恢复团队空间。
func (s *Service) Restore(ctx context.Context, id int64) error {
	tag, err := s.db.Exec(ctx, `
		UPDATE teamspaces SET archived_at = NULL, updated_at = now()
		WHERE id = $1 AND archived_at IS NOT NULL`, id)
	if err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	if tag.RowsAffected() == 0 {
		return ErrTeamspaceNotFound
	}
	return nil
}

// Delete 删除团队空间（含关联）。
func (s *Service) Delete(ctx context.Context, id int64) error {
	if _, err := s.db.Exec(ctx, `DELETE FROM teamspace_members WHERE teamspace_id = $1`, id); err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	if _, err := s.db.Exec(ctx, `DELETE FROM teamspace_projects WHERE teamspace_id = $1`, id); err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	tag, err := s.db.Exec(ctx, `DELETE FROM teamspaces WHERE id = $1`, id)
	if err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	if tag.RowsAffected() == 0 {
		return ErrTeamspaceNotFound
	}
	return nil
}

// --- 成员 ---

// AddMember 添加成员。
func (s *Service) AddMember(ctx context.Context, teamspaceID, userID int64, role MemberRole) error {
	if userID <= 0 {
		return errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "user_id", Reason: "required"})
	}
	if !role.IsValid() {
		role = MemberRoleMember
	}
	_, err := s.db.Exec(ctx, `
		INSERT INTO teamspace_members (teamspace_id, user_id, role)
		VALUES ($1,$2,$3) ON CONFLICT (teamspace_id, user_id) DO UPDATE SET role = EXCLUDED.role`,
		teamspaceID, userID, string(role))
	if err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	return nil
}

// RemoveMember 移除成员。
func (s *Service) RemoveMember(ctx context.Context, teamspaceID, userID int64) error {
	_, err := s.db.Exec(ctx, `
		DELETE FROM teamspace_members WHERE teamspace_id = $1 AND user_id = $2`,
		teamspaceID, userID)
	if err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	return nil
}

// ListMembers 列出成员。
func (s *Service) ListMembers(ctx context.Context, teamspaceID int64) ([]TeamspaceMember, error) {
	rows, err := s.db.Query(ctx, `
		SELECT teamspace_id, user_id, role, created_at
		FROM teamspace_members WHERE teamspace_id = $1 ORDER BY created_at ASC`, teamspaceID)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()
	var out []TeamspaceMember
	for rows.Next() {
		var m TeamspaceMember
		if err := rows.Scan(&m.TeamspaceID, &m.UserID, &m.Role, &m.CreatedAt); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

// --- 项目 ---

// AddProject 关联项目。
func (s *Service) AddProject(ctx context.Context, teamspaceID, projectID int64) error {
	if projectID <= 0 {
		return errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "project_id", Reason: "required"})
	}
	_, err := s.db.Exec(ctx, `
		INSERT INTO teamspace_projects (teamspace_id, project_id)
		VALUES ($1,$2) ON CONFLICT DO NOTHING`, teamspaceID, projectID)
	if err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	return nil
}

// RemoveProject 解除项目关联。
func (s *Service) RemoveProject(ctx context.Context, teamspaceID, projectID int64) error {
	_, err := s.db.Exec(ctx, `
		DELETE FROM teamspace_projects WHERE teamspace_id = $1 AND project_id = $2`,
		teamspaceID, projectID)
	if err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	return nil
}

// ListProjects 列出关联项目 ID。
func (s *Service) ListProjects(ctx context.Context, teamspaceID int64) ([]int64, error) {
	rows, err := s.db.Query(ctx, `
		SELECT project_id FROM teamspace_projects WHERE teamspace_id = $1 ORDER BY created_at ASC`,
		teamspaceID)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()
	var out []int64
	for rows.Next() {
		var pid int64
		if err := rows.Scan(&pid); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
		out = append(out, pid)
	}
	return out, rows.Err()
}

// --- 进度 ---

// GetProgress 计算团队进度汇总。
//
// 数据来源：
//   - 成员维度：teamspace_members；
//   - 项目维度：teamspace_projects；
//   - 工作项维度：关联项目下、指派给团队成员的工作项（跨 requirement/task/defect 三表）。
func (s *Service) GetProgress(ctx context.Context, teamspaceID int64) (*TeamspaceProgress, error) {
	p := &TeamspaceProgress{TeamspaceID: teamspaceID}

	// 成员数与项目数
	if err := s.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM teamspace_members WHERE teamspace_id = $1`, teamspaceID).
		Scan(&p.MemberCount); err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	if err := s.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM teamspace_projects WHERE teamspace_id = $1`, teamspaceID).
		Scan(&p.ProjectCount); err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}

	// 活跃迭代数
	if err := s.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM sprints sp
		JOIN teamspace_projects tp ON tp.project_id = sp.project_id
		WHERE tp.teamspace_id = $1 AND sp.status = 'active'`, teamspaceID).
		Scan(&p.ActiveCycles); err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}

	// 工作项维度（跨 requirement / task / defect 三张表 + 各自 assignees 关联表）。
	// 逾期判定：target_date < today 且状态分组不属于 completed/cancelled。
	const aggQuery = `
		SELECT
			COUNT(*) AS total,
			COUNT(*) FILTER (WHERE st."group" IN ('completed','cancelled')) AS completed,
			COUNT(*) FILTER (WHERE st."group" NOT IN ('completed','cancelled')
			                      AND w.target_date IS NOT NULL
			                      AND w.target_date < CURRENT_DATE) AS overdue,
			COUNT(*) FILTER (WHERE st."group" = 'started') AS in_progress
		FROM (
			SELECT w.project_id, w.state_id, w.target_date, a.user_id
			FROM requirement w
			JOIN requirement_assignees a ON a.requirement_id = w.id AND a.deleted = false
			UNION ALL
			SELECT w.project_id, w.state_id, w.target_date, a.user_id
			FROM task w
			JOIN task_assignees a ON a.task_id = w.id AND a.deleted = false
			UNION ALL
			SELECT w.project_id, w.state_id, w.target_date, a.user_id
			FROM defect w
			JOIN defect_assignees a ON a.defect_id = w.id AND a.deleted = false
		) w
		JOIN teamspace_projects tp ON tp.project_id = w.project_id
		JOIN teamspace_members tm ON tm.teamspace_id = tp.teamspace_id AND tm.user_id = w.user_id
		JOIN states st ON st.id = w.state_id
		WHERE tp.teamspace_id = $1`
	if err := s.db.QueryRow(ctx, aggQuery, teamspaceID).
		Scan(&p.TotalWorkItems, &p.CompletedWorkItems, &p.OverdueWorkItems, &p.InProgressWorkItems); err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}

	if p.TotalWorkItems > 0 {
		p.CompletionRate = float64(p.CompletedWorkItems) / float64(p.TotalWorkItems) * 100
	}
	return p, nil
}

// ListMemberWorkloads 列出成员工作负载。
func (s *Service) ListMemberWorkloads(ctx context.Context, teamspaceID int64) ([]TeamspaceMemberWorkload, error) {
	rows, err := s.db.Query(ctx, `
		SELECT w.user_id,
		       COUNT(*) AS total,
		       COUNT(*) FILTER (WHERE st."group" = 'started') AS in_progress,
		       COUNT(*) FILTER (WHERE st."group" NOT IN ('completed','cancelled')
		                          AND w.target_date IS NOT NULL
		                          AND w.target_date < CURRENT_DATE) AS overdue,
		       CASE WHEN COUNT(*) > 0
		            THEN COUNT(*) FILTER (WHERE st."group" IN ('completed','cancelled'))::float8
		                 / COUNT(*) * 100
		            ELSE 0 END AS completion_rate
		FROM (
			SELECT w.project_id, w.state_id, w.target_date, a.user_id
			FROM requirement w
			JOIN requirement_assignees a ON a.requirement_id = w.id AND a.deleted = false
			UNION ALL
			SELECT w.project_id, w.state_id, w.target_date, a.user_id
			FROM task w
			JOIN task_assignees a ON a.task_id = w.id AND a.deleted = false
			UNION ALL
			SELECT w.project_id, w.state_id, w.target_date, a.user_id
			FROM defect w
			JOIN defect_assignees a ON a.defect_id = w.id AND a.deleted = false
		) w
		JOIN teamspace_projects tp ON tp.project_id = w.project_id
		JOIN teamspace_members tm ON tm.teamspace_id = tp.teamspace_id AND tm.user_id = w.user_id
		JOIN states st ON st.id = w.state_id
		WHERE tp.teamspace_id = $1
		GROUP BY w.user_id
		ORDER BY total DESC`, teamspaceID)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()
	var out []TeamspaceMemberWorkload
	for rows.Next() {
		var w TeamspaceMemberWorkload
		if err := rows.Scan(&w.UserID, &w.TotalWorkItems, &w.InProgressItems, &w.OverdueItems, &w.CompletionRate); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
		out = append(out, w)
	}
	return out, rows.Err()
}

// --- helpers ---

func nullIfEmpty(s string) any {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return s
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
