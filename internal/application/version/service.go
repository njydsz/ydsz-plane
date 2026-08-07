// Package version — 版本应用服务（CRUD / 生命周期 / 发布 / Release Notes / 质量门禁 / 交付报告）。
//
// 参考: Plane / Jira Fix Version / GitHub Release workflow。
// 设计要点:
//   - 同一项目内 semver 唯一(未删除)
//   - 状态机: planning → active → released → archived（不可逆）
//   - 发布动作 4 步: 清单全勾选 → 准出校验 → 生成 Notes+报告 → 状态推进
//   - 跨迭代进度聚合在版本详情中即时计算（事件失效缓存留给 M5 仪表盘）
//   - 迭代与版本是 N:1 关系: sprints.version_id FK（一个迭代只属于一个版本）
package version

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// Service 提供版本域应用服务。
type Service struct {
	db *pgxpool.Pool
}

// NewService 创建版本域服务。
func NewService(db *pgxpool.Pool) *Service {
	return &Service{db: db}
}

// ---- 状态机 ----

// canTransition 版本状态流转规则。
func canTransition(from, to VersionStatusCode) bool {
	switch from {
	case VersionPlanning:
		return to == VersionActive || to == VersionArchived
	case VersionActive:
		return to == VersionReleased || to == VersionArchived
	case VersionReleased:
		return to == VersionArchived
	case VersionArchived:
		return false
	}
	return false
}

// checklistAllRequiredChecked 检查清单全部 required 项是否已 checked。
func checklistAllRequiredChecked(items []ChecklistItem) bool {
	if len(items) == 0 {
		return true
	}
	for _, it := range items {
		if it.Required && !it.Checked {
			return false
		}
	}
	return true
}

// normalizeChecklist 对请求传入的清单规范化：补全 ID（客户端未传时）、trim label。
func normalizeChecklist(in []ChecklistItem) []ChecklistItem {
	if in == nil {
		return []ChecklistItem{}
	}
	out := make([]ChecklistItem, 0, len(in))
	for i, it := range in {
		it.Label = strings.TrimSpace(it.Label)
		if it.Label == "" {
			continue
		}
		if it.ID == "" {
			it.ID = fmt.Sprintf("chk-%d", i+1)
		}
		out = append(out, it)
	}
	return out
}

// ---- CRUD ----

// Create 创建版本。
func (s *Service) Create(ctx context.Context, in CreateVersionInput) (*Version, error) {
	if err := validateCreateInput(in); err != nil {
		return nil, err
	}
	// semver 校验
	if semErr, _ := ParseSemVer(in.Semver); semErr != nil {
		return nil, errs.ErrVersionSemverInvalid.WithDetails(errs.FieldDetail{Field: "semver", Reason: semErr.Error()})
	}
	checklist := normalizeChecklist(in.Checklist)

	var v *Version
	err := s.withTx(ctx, in.WorkspaceID, func(tx pgx.Tx) error {
		// 唯一性校验: 项目内 semver 未删除不重复
		var exists bool
		if err := tx.QueryRow(ctx,
			`SELECT EXISTS(SELECT 1 FROM versions WHERE project_id = $1 AND semver = $2 AND deleted_at IS NULL)`,
			in.ProjectID, in.Semver).Scan(&exists); err != nil {
			return errs.ErrInternal.Wrap(err)
		}
		if exists {
			return errs.ErrVersionDataConflict.WithDetails(errs.FieldDetail{Field: "semver", Reason: "项目下已有该语义版本"})
		}

		clRaw, _ := json.Marshal(checklist)
		var target, start, end interface{}
		if in.TargetDate != nil {
			target = *in.TargetDate
		}
		if in.StartDate != nil {
			start = *in.StartDate
		}
		if in.EndDate != nil {
			end = *in.EndDate
		}

		var id int64
		err := tx.QueryRow(ctx, `
			INSERT INTO versions (workspace_id, project_id, name, semver, description,
				status, checklist, start_date, end_date, target_date, created_by)
			VALUES ($1,$2,$3,$4,$5,'planning',$6,$7,$8,$9,$10)
			RETURNING id, workspace_id, project_id, name, semver, description, status,
				release_notes, delivered_at, start_date, end_date, target_date, archived_at, created_by, created_at, updated_at`,
			in.WorkspaceID, in.ProjectID, in.Name, in.Semver, in.Description,
			clRaw, start, end, target, in.CreatedBy).Scan(
			&id, &v.WorkspaceID, &v.ProjectID, &v.Name, &v.Semver,
			&v.Description, &v.Status, &v.ReleaseNotes, &v.DeliveredAt,
			&v.StartDate, &v.EndDate, &v.TargetDate,
			&v.ArchivedAt, &v.CreatedBy, &v.CreatedAt, &v.UpdatedAt)
		if err != nil {
			return errs.ErrInternal.Wrap(err)
		}
		v.ID = id
		v.Checklist = checklist
		return nil
	})
	if err != nil {
		return nil, s.mapPgError(err)
	}
	return v, nil
}

// GetByID 获取版本详情(包含聚合进度/质量)。
func (s *Service) GetByID(ctx context.Context, wsID, versionID int64) (*Version, error) {
	v, err := s.getVersion(ctx, wsID, versionID)
	if err != nil {
		return nil, err
	}
	v.Sprints = s.listSprints(ctx, wsID, v.ID)
	v.Progress = s.computeProgress(ctx, wsID, v)
	v.Quality = s.computeQuality(ctx, wsID, v)
	v.DeliveryReport = s.computeDeliveryReport(ctx, wsID, v, v.Progress, v.Quality)
	return v, nil
}

// List 查询版本列表。
func (s *Service) List(ctx context.Context, opts ListVersionsOptions) ([]Version, int64, error) {
	if opts.Limit <= 0 || opts.Limit > 100 {
		opts.Limit = 50
	}
	if opts.Offset < 0 {
		opts.Offset = 0
	}
	where := "WHERE v.deleted_at IS NULL AND v.project_id = $1 AND v.workspace_id = $2"
	args := []interface{}{opts.ProjectID, opts.WorkspaceID}
	arg := 3

	if opts.Status != nil {
		where += " AND v.status = $" + strconv.Itoa(arg)
		args = append(args, string(*opts.Status))
		arg++
	}

	var total int64
	_ = s.db.QueryRow(ctx, "SELECT count(*) FROM versions v "+where, args...).Scan(&total)

	limitIdx := len(args) + 1
	offsetIdx := len(args) + 2
	args = append(args, opts.Limit, opts.Offset)

	rows, err := s.db.Query(ctx, `
		SELECT v.id, v.workspace_id, v.project_id, v.name, v.semver, v.description,
		       v.status, v.release_notes, v.delivered_at,
		       v.start_date, v.end_date, v.target_date, v.archived_at,
		       v.created_by, v.created_at, v.updated_at
		FROM versions v `+where+`
		ORDER BY
			CASE v.status WHEN 'active' THEN 0 WHEN 'planning' THEN 1 WHEN 'released' THEN 2 ELSE 3 END,
			v.target_date NULLS LAST, v.created_at DESC
		LIMIT $`+strconv.Itoa(limitIdx)+` OFFSET $`+strconv.Itoa(offsetIdx), args...)
	if err != nil {
		return nil, 0, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()

	var out []Version
	for rows.Next() {
		v, err := scanVersion(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, *v)
	}
	return out, total, rows.Err()
}

// Update 更新版本字段(planning/active 状态允许修改；released/archived 不可)。
func (s *Service) Update(ctx context.Context, wsID, versionID int64, in UpdateVersionInput) (*Version, error) {
	var result *Version
	err := s.withTx(ctx, wsID, func(tx pgx.Tx) error {
		v, err := s.getVersionTx(ctx, tx, wsID, versionID)
		if err != nil {
			return err
		}
		if v.Status == VersionReleased || v.Status == VersionArchived {
			return errs.ErrVersionInvalidLifecycle
		}

		semverStr := v.Semver
		if in.Semver != nil && *in.Semver != v.Semver {
			if semErr, _ := ParseSemVer(*in.Semver); semErr != nil {
				return errs.ErrVersionSemverInvalid.WithDetails(errs.FieldDetail{Field: "semver", Reason: semErr.Error()})
			}
			var exists bool
			if err := tx.QueryRow(ctx,
				`SELECT EXISTS(SELECT 1 FROM versions WHERE project_id = $1 AND semver = $2 AND id <> $3 AND deleted_at IS NULL)`,
				v.ProjectID, *in.Semver, versionID).Scan(&exists); err != nil {
				return errs.ErrInternal.Wrap(err)
			}
			if exists {
				return errs.ErrVersionDataConflict.WithDetails(errs.FieldDetail{Field: "semver", Reason: "项目下已有该语义版本"})
			}
			semverStr = *in.Semver
		}

		cl := v.Checklist
		if in.Checklist != nil {
			cl = normalizeChecklist(in.Checklist)
		}

		clRaw, _ := json.Marshal(cl)
		sets, args := buildUpdateSet(in, semverStr, clRaw)
		if len(sets) == 0 {
			result = v
			return nil
		}
		args = append(args, versionID, wsID)
		query := fmt.Sprintf(`UPDATE versions SET %s WHERE id = $%d AND workspace_id = $%d AND status NOT IN ('released','archived')`,
			strings.Join(sets, ", "), len(args)-1, len(args))

		tag, err := tx.Exec(ctx, query, args...)
		if err != nil {
			return s.mapPgError(err)
		}
		if tag.RowsAffected() == 0 {
			return errs.ErrVersionDataConflict
		}
		result, err = s.getVersion(ctx, wsID, versionID)
		return err
	})
	if err != nil {
		return nil, s.mapPgError(err)
	}
	return result, nil
}

// SoftDelete 删除版本 (仅 planning/archived 可删除)。
func (s *Service) SoftDelete(ctx context.Context, wsID, versionID int64) error {
	return s.withTx(ctx, wsID, func(tx pgx.Tx) error {
		v, err := s.getVersionTx(ctx, tx, wsID, versionID)
		if err != nil {
			return err
		}
		if v.Status == VersionActive || v.Status == VersionReleased {
			return errs.ErrVersionInvalidLifecycle
		}
		tag, err := tx.Exec(ctx, `UPDATE versions SET deleted_at = now(), updated_at = now() WHERE id = $1 AND workspace_id = $2`,
			versionID, wsID)
		if err != nil {
			return errs.ErrInternal.Wrap(err)
		}
		if tag.RowsAffected() == 0 {
			return errs.ErrVersionNotFound
		}
		return nil
	})
}

// ---- 状态机操作 ----

// Activate 激活版本 (planning → active)。
func (s *Service) Activate(ctx context.Context, wsID, versionID int64) (*Version, error) {
	return s.transition(ctx, wsID, versionID, VersionActive)
}

// Release 发布版本 (active → released)。
func (s *Service) Release(ctx context.Context, wsID, versionID int64, in ReleaseVersionInput) (*Version, error) {
	var result *Version
	err := s.withTx(ctx, wsID, func(tx pgx.Tx) error {
		v, err := s.getVersionTx(ctx, tx, wsID, versionID)
		if err != nil {
			return err
		}
		if v.Status != VersionActive {
			return errs.ErrVersionInvalidLifecycle.WithDetails(
				errs.FieldDetail{Field: "status", Reason: "仅 active 状态可发布"})
		}

		// 1) 清单校验
		if !in.ForceChecklist && !checklistAllRequiredChecked(v.Checklist) {
			return errs.ErrVersionChecklistIncomplete
		}

		// 2) 准出校验
		quality := s.computeQuality(ctx, wsID, v)
		if !quality.PassQualityGate {
			return errs.ErrVersionNotQualityGate.WithDetails(
				errs.FieldDetail{Field: "quality", Reason: fmt.Sprintf("严重/致命未关闭缺陷 %d 个", quality.CriticalBugs)})
		}

		// 3) Release Notes
		progress := s.computeProgress(ctx, wsID, v)
		_ = s.computeDeliveryReport(ctx, wsID, v, progress, quality)
		noteSrc := s.buildReleaseNotesSource(ctx, wsID, v, in.AddKnownIssuesToNotes)
		notes := renderReleaseNotes(v, noteSrc)
		if strings.TrimSpace(in.DraftOverride) != "" {
			notes = in.DraftOverride
		}

		now := time.Now().UTC()
		tag, err := tx.Exec(ctx, `
			UPDATE versions SET status = 'released', release_notes = $1, delivered_at = $2,
				updated_at = now()
			WHERE id = $3 AND workspace_id = $4 AND status = 'active'`,
			notes, now, versionID, wsID)
		if err != nil {
			return errs.ErrInternal.Wrap(err)
		}
		if tag.RowsAffected() == 0 {
			return errs.ErrVersionDataConflict
		}

		result, err = s.getVersion(ctx, wsID, versionID)
		return err
	})
	if err != nil {
		return nil, s.mapPgError(err)
	}
	return result, nil
}

// transition 内部版本状态机推进。
func (s *Service) transition(ctx context.Context, wsID, versionID int64, to VersionStatusCode) (*Version, error) {
	var result *Version
	err := s.withTx(ctx, wsID, func(tx pgx.Tx) error {
		v, err := s.getVersionTx(ctx, tx, wsID, versionID)
		if err != nil {
			return err
		}
		if !canTransition(v.Status, to) {
			return errs.ErrVersionInvalidLifecycle
		}
		var archivedAt string
		if to == VersionArchived {
			archivedAt = "archived_at = now(),"
		}
		query := fmt.Sprintf(`UPDATE versions SET status = $1, %s updated_at = now()
			WHERE id = $2 AND workspace_id = $3 AND status = $4`, archivedAt)
		fromStatus := string(v.Status)
		tag, err := tx.Exec(ctx, query, string(to), versionID, wsID, fromStatus)
		if err != nil {
			return errs.ErrInternal.Wrap(err)
		}
		if tag.RowsAffected() == 0 {
			return errs.ErrVersionDataConflict
		}
		result, err = s.getVersion(ctx, wsID, versionID)
		return err
	})
	if err != nil {
		return nil, s.mapPgError(err)
	}
	return result, nil
}

// Archive 归档版本 (任何状态都可归档)。
func (s *Service) Archive(ctx context.Context, wsID, versionID int64) (*Version, error) {
	return s.transition(ctx, wsID, versionID, VersionArchived)
}

// ---- 迭代聚合 ----

// AddSprint 将迭代归属到版本 (N:1: 更新迭代的 version_id)。
// 一个迭代只能属于一个版本，设置新版本时会覆盖原有归属。
func (s *Service) AddSprint(ctx context.Context, wsID int64, in AddSprintInput) error {
	return s.withTx(ctx, wsID, func(tx pgx.Tx) error {
		v, err := s.getVersionTx(ctx, tx, wsID, in.VersionID)
		if err != nil {
			return err
		}
		if v.Status == VersionReleased || v.Status == VersionArchived {
			return errs.ErrVersionInvalidLifecycle
		}
		var sprintProject int64
		if err := tx.QueryRow(ctx,
			`SELECT project_id FROM sprints WHERE id = $1 AND workspace_id = $2 AND deleted_at IS NULL`,
			in.SprintID, wsID).Scan(&sprintProject); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return errs.ErrVersionNotFound.WithDetails(errs.FieldDetail{Field: "sprint_id", Reason: "迭代不存在"})
			}
			return errs.ErrInternal.Wrap(err)
		}
		if sprintProject != v.ProjectID {
			return errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "sprint_id", Reason: "迭代不属于当前项目"})
		}
		tag, err := tx.Exec(ctx, `
			UPDATE sprints SET version_id = $1, updated_at = now()
			WHERE id = $2 AND workspace_id = $3 AND deleted_at IS NULL`,
			in.VersionID, in.SprintID, wsID)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return errs.ErrVersionNotFound.WithDetails(errs.FieldDetail{Field: "sprint_id", Reason: "迭代更新失败"})
		}
		return nil
	})
}

// RemoveSprint 将迭代从版本解绑 (迭代的 version_id 置 NULL)。
func (s *Service) RemoveSprint(ctx context.Context, wsID, versionID, sprintID int64) error {
	return s.withTx(ctx, wsID, func(tx pgx.Tx) error {
		v, err := s.getVersionTx(ctx, tx, wsID, versionID)
		if err != nil {
			return err
		}
		if v.Status == VersionReleased || v.Status == VersionArchived {
			return errs.ErrVersionInvalidLifecycle
		}
		tag, err := tx.Exec(ctx, `
			UPDATE sprints SET version_id = NULL, updated_at = now()
			WHERE id = $1 AND version_id = $2 AND workspace_id = $3`,
			sprintID, versionID, wsID)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return errs.ErrVersionNotFound.WithDetails(errs.FieldDetail{Field: "sprint_id", Reason: "迭代未归属该版本"})
		}
		return nil
	})
}

// ---- 进度 / 质量 / 报告计算 ----

func (s *Service) computeProgress(ctx context.Context, wsID int64, v *Version) *VersionProgress {
	p := &VersionProgress{ByStateGroup: map[string]float64{}}

	// 直接从 sprints.version_id 聚合，不再使用 version_sprints 关联表
	rows, err := s.db.Query(ctx, `
		SELECT coalesce(sum(CASE WHEN i.point IS NOT NULL THEN i.point ELSE 0 END), 0),
		       coalesce(sum(CASE WHEN sg."group" = 'completed' AND i.point IS NOT NULL THEN i.point ELSE 0 END), 0),
		       count(DISTINCT i.id),
		       count(DISTINCT i.id) FILTER (WHERE sg."group" = 'completed')
		FROM sprints sp
		JOIN sprint_issues si ON si.sprint_id = sp.id
		JOIN issues i ON i.id = si.issue_id AND i.deleted_at IS NULL
		JOIN states sg ON sg.id = i.state_id
		WHERE sp.version_id = $1`, v.ID)
	if err == nil {
		defer rows.Close()
		if rows.Next() {
			_ = rows.Scan(&p.TotalPoints, &p.DonePoints, &p.TotalIssues, &p.DoneIssues)
		}
	}

	rows2, err := s.db.Query(ctx, `
		SELECT sg."group",
		       coalesce(sum(CASE WHEN i.point IS NOT NULL THEN i.point ELSE 0 END), 0)
		FROM sprints sp
		JOIN sprint_issues si ON si.sprint_id = sp.id
		JOIN issues i ON i.id = si.issue_id AND i.deleted_at IS NULL
		JOIN states sg ON sg.id = i.state_id
		WHERE sp.version_id = $1
		GROUP BY sg."group"`, v.ID)
	if err == nil {
		defer rows2.Close()
		for rows2.Next() {
			var grp string
			var pts float64
			if err := rows2.Scan(&grp, &pts); err == nil {
				p.ByStateGroup[grp] = pts
			}
		}
	}

	if p.TotalPoints > 0 {
		p.CompletionRate = math.Min(p.DonePoints/p.TotalPoints, 1)
	}

	_ = s.db.QueryRow(ctx, `SELECT count(*) FROM sprints WHERE version_id = $1 AND deleted_at IS NULL`, v.ID).Scan(&p.SprintCount)
	return p
}

func (s *Service) computeQuality(ctx context.Context, wsID int64, v *Version) *QualityMetrics {
	q := &QualityMetrics{BugBySeverity: map[int]int{}}

	rows, err := s.db.Query(ctx, `
		SELECT i.severity, sg."group", count(*)
		FROM issues i
		JOIN states sg ON sg.id = i.state_id
		WHERE i.project_id = $1 AND i.type_code = 'defect' AND i.found_version_id = $2 AND i.deleted_at IS NULL
		GROUP BY i.severity, sg."group"`, v.ProjectID, v.ID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var sev, cnt int
			var grp string
			if err := rows.Scan(&sev, &grp, &cnt); err == nil {
				q.FoundBugCount += cnt
				q.BugBySeverity[sev] += cnt
				isOpen := grp != "completed" && grp != "cancelled"
				if sev <= 1 {
					if isOpen {
						q.CriticalBugs += cnt
						q.OpenBugs += cnt
					}
				} else {
					if isOpen {
						q.OpenBugs += cnt
					}
					if sev == 2 {
						q.MajorBugs += cnt
					}
				}
			}
		}
	}

	_ = s.db.QueryRow(ctx, `
		SELECT count(*), count(*) FILTER (WHERE sg."group" = 'completed')
		FROM issues i
		JOIN states sg ON sg.id = i.state_id
		WHERE i.project_id = $1 AND i.type_code = 'defect' AND i.fix_version_id = $2 AND i.deleted_at IS NULL`,
		v.ProjectID, v.ID).Scan(&q.TotalBugs, &q.FixedBugCount)

	if q.FoundBugCount > 0 {
		q.FixRate = math.Min(float64(q.FixedBugCount)/float64(q.FoundBugCount), 1)
	} else {
		q.FixRate = 1
	}
	q.PassQualityGate = q.CriticalBugs == 0
	return q
}

func (s *Service) computeDeliveryReport(ctx context.Context, wsID int64, v *Version, p *VersionProgress, q *QualityMetrics) *DeliveryReport {
	return &DeliveryReport{
		GeneratedAt:       time.Now().UTC(),
		SprintCount:       p.SprintCount,
		TotalPoints:       p.TotalPoints,
		CompletedPoints:   p.DonePoints,
		TotalIssues:       p.TotalIssues,
		CompletedIssues:   p.DoneIssues,
		BugCount:          q.FoundBugCount,
		FixedBugCount:     q.FixedBugCount,
		PassRate:          p.CompletionRate,
		EligibleToRelease: q.PassQualityGate && p.CompletionRate >= 0.8,
	}
}

// Release Notes ----

func (s *Service) buildReleaseNotesSource(ctx context.Context, wsID int64, v *Version, includeKnownIssues bool) *ReleaseNotesData {
	src := &ReleaseNotesData{VersionName: v.Name, Semver: v.Semver}

	rows, err := s.db.Query(ctx, `
		SELECT DISTINCT i.identifier, i.name, st.name
		FROM sprints sp
		JOIN sprint_issues si ON si.sprint_id = sp.id
		JOIN issues i ON i.id = si.issue_id AND i.deleted_at IS NULL
		JOIN states st ON st.id = i.state_id AND st."group" = 'completed'
		WHERE sp.version_id = $1 AND i.type_code IN ('requirement','task')
		ORDER BY i.identifier`, v.ID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var r NoteIssueRef
			if err := rows.Scan(&r.Identifier, &r.Name, &r.StateName); err == nil {
				src.RequirementsDone = append(src.RequirementsDone, r)
			}
		}
	}

	rows2, err := s.db.Query(ctx, `
		SELECT i.identifier, i.name, st.name
		FROM issues i
		JOIN states st ON st.id = i.state_id AND st."group" = 'completed'
		WHERE i.project_id = $1 AND i.type_code = 'defect' AND i.fix_version_id = $2 AND i.deleted_at IS NULL
		ORDER BY i.identifier`, v.ProjectID, v.ID)
	if err == nil {
		defer rows2.Close()
		for rows2.Next() {
			var r NoteIssueRef
			if err := rows2.Scan(&r.Identifier, &r.Name, &r.StateName); err == nil {
				src.BugsFixed = append(src.BugsFixed, r)
			}
		}
	}

	if includeKnownIssues {
		rows3, err := s.db.Query(ctx, `
			SELECT i.identifier, i.name, st.name
			FROM issues i
			JOIN states st ON st.id = i.state_id
			WHERE i.project_id = $1 AND i.type_code = 'defect' AND i.found_version_id = $2
				AND i.deleted_at IS NULL AND st."group" NOT IN ('completed','cancelled')
			ORDER BY i.identifier`, v.ProjectID, v.ID)
		if err == nil {
			defer rows3.Close()
			for rows3.Next() {
				var r NoteIssueRef
				if err := rows3.Scan(&r.Identifier, &r.Name, &r.StateName); err == nil {
					src.KnownIssues = append(src.KnownIssues, r)
				}
			}
		}
	}
	return src
}

// renderReleaseNotes 三段式 Release Notes 渲染。
func renderReleaseNotes(v *Version, src *ReleaseNotesData) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("# %s v%s\n\n", v.Name, v.Semver))
	b.WriteString(fmt.Sprintf("> Released at: %s\n\n", time.Now().UTC().Format("2006-01-02")))
	if v.Description != "" {
		b.WriteString(v.Description + "\n\n")
	}

	b.WriteString("## ✅ 已完成需求与任务\n")
	if len(src.RequirementsDone) == 0 {
		b.WriteString("- （无完成的需求/任务）\n")
	} else {
		for _, r := range src.RequirementsDone {
			b.WriteString(fmt.Sprintf("- **%s** %s\n", r.Identifier, r.Name))
		}
	}

	b.WriteString("\n## 🐛 修复缺陷\n")
	if len(src.BugsFixed) == 0 {
		b.WriteString("- （无修复缺陷）\n")
	} else {
		for _, r := range src.BugsFixed {
			b.WriteString(fmt.Sprintf("- **%s** %s\n", r.Identifier, r.Name))
		}
	}

	if len(src.KnownIssues) > 0 {
		b.WriteString("\n## ⚠️ 已知问题\n")
		for _, r := range src.KnownIssues {
			b.WriteString(fmt.Sprintf("- **%s** %s (%s)\n", r.Identifier, r.Name, r.StateName))
		}
	}
	return b.String()
}

// ---- 进度查询 / 过滤 ----

func (s *Service) Progress(ctx context.Context, wsID, versionID int64) (*VersionProgress, error) {
	v, err := s.getVersion(ctx, wsID, versionID)
	if err != nil {
		return nil, err
	}
	return s.computeProgress(ctx, wsID, v), nil
}

// DefectPanel 版本缺陷面板(按 found_version 聚合)。
func (s *Service) DefectPanel(ctx context.Context, wsID, versionID int64) ([]BugVersionView, int64, error) {
	v, err := s.getVersion(ctx, wsID, versionID)
	if err != nil {
		return nil, 0, err
	}
	rows, err := s.db.Query(ctx, `
		SELECT i.id, i.identifier, i.name, i.severity, i.found_phase,
		       st.name, st."group", i.root_cause_category,
		       fv.semver, fxv.semver
		FROM issues i
		JOIN states st ON st.id = i.state_id
		LEFT JOIN versions fv ON fv.id = i.found_version_id
		LEFT JOIN versions fxv ON fxv.id = i.fix_version_id
		WHERE i.project_id = $1 AND i.type_code = 'defect' AND i.deleted_at IS NULL
			AND (i.found_version_id = $2 OR i.fix_version_id = $2)
		ORDER BY i.severity NULLS LAST, i.created_at DESC`,
		v.ProjectID, versionID)
	if err != nil {
		return nil, 0, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()

	var views []BugVersionView
	for rows.Next() {
		var bv BugVersionView
		var sev sql.NullInt64
		var fp, rc, fs, fx sql.NullString
		if err := rows.Scan(&bv.IssueID, &bv.Identifier, &bv.Name, &sev, &fp,
			&bv.StateName, &bv.StateGroup, &rc, &fs, &fx); err != nil {
			return nil, 0, errs.ErrInternal.Wrap(err)
		}
		if sev.Valid {
			n := int(sev.Int64)
			bv.Severity = &n
		}
		if fp.Valid {
			bv.FoundPhase = fp.String
		}
		if rc.Valid {
			bv.RootCause = rc.String
		}
		if fs.Valid {
			bv.FoundVersion = fs.String
		}
		if fx.Valid {
			bv.FixVersion = fx.String
		}
		views = append(views, bv)
	}
	return views, int64(len(views)), nil
}

// FilterDefects 跨版本缺陷过滤。
func (s *Service) FilterDefects(ctx context.Context, f BugVersionFilter) ([]BugVersionView, int64, error) {
	if f.Limit <= 0 || f.Limit > 200 {
		f.Limit = 50
	}
	if f.Offset < 0 {
		f.Offset = 0
	}

	where := `WHERE i.project_id = $1 AND i.workspace_id = $2 AND i.type_code = 'defect' AND i.deleted_at IS NULL`
	args := []interface{}{f.ProjectID, f.WorkspaceID}
	arg := 3

	if f.FoundVersionID != nil {
		where += " AND i.found_version_id = $" + strconv.Itoa(arg)
		args = append(args, *f.FoundVersionID)
		arg++
	}
	if f.FixVersionID != nil {
		where += " AND i.fix_version_id = $" + strconv.Itoa(arg)
		args = append(args, *f.FixVersionID)
		arg++
	}
	if f.StateGroup != nil {
		where += ` AND sg."group" = $` + strconv.Itoa(arg)
		args = append(args, *f.StateGroup)
		arg++
	}
	if f.Severity != nil {
		where += " AND i.severity = $" + strconv.Itoa(arg)
		args = append(args, *f.Severity)
		arg++
	}

	var total int64
	_ = s.db.QueryRow(ctx, `
		SELECT count(*) FROM issues i
		JOIN states sg ON sg.id = i.state_id `+where, args...).Scan(&total)

	limitIdx := len(args) + 1
	offsetIdx := len(args) + 2
	args = append(args, f.Limit, f.Offset)

	rows, err := s.db.Query(ctx, `
		SELECT i.id, i.identifier, i.name, i.severity, i.found_phase,
		       sg.name, sg."group", i.root_cause_category,
		       fv.semver, fxv.semver
		FROM issues i
		JOIN states sg ON sg.id = i.state_id
		LEFT JOIN versions fv ON fv.id = i.found_version_id
		LEFT JOIN versions fxv ON fxv.id = i.fix_version_id `+where+`
		ORDER BY i.severity NULLS LAST, i.created_at DESC
		LIMIT $`+strconv.Itoa(limitIdx)+` OFFSET $`+strconv.Itoa(offsetIdx), args...)
	if err != nil {
		return nil, 0, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()

	var views []BugVersionView
	for rows.Next() {
		var bv BugVersionView
		var sev sql.NullInt64
		var fp, rc, fs, fx sql.NullString
		if err := rows.Scan(&bv.IssueID, &bv.Identifier, &bv.Name, &sev, &fp,
			&bv.StateName, &bv.StateGroup, &rc, &fs, &fx); err != nil {
			return nil, 0, errs.ErrInternal.Wrap(err)
		}
		if sev.Valid {
			n := int(sev.Int64)
			bv.Severity = &n
		}
		if fp.Valid {
			bv.FoundPhase = fp.String
		}
		if rc.Valid {
			bv.RootCause = rc.String
		}
		if fs.Valid {
			bv.FoundVersion = fs.String
		}
		if fx.Valid {
			bv.FixVersion = fx.String
		}
		views = append(views, bv)
	}
	return views, total, rows.Err()
}

// ---- 低层查询 ----

func (s *Service) getVersion(ctx context.Context, wsID, versionID int64) (*Version, error) {
	row := s.db.QueryRow(ctx, `
		SELECT v.id, v.workspace_id, v.project_id, v.name, v.semver, v.description,
		       v.status, v.release_notes, v.delivered_at,
		       v.start_date, v.end_date, v.target_date, v.archived_at,
		       v.created_by, v.created_at, v.updated_at,
		       v.checklist
		FROM versions v
		WHERE v.id = $1 AND v.workspace_id = $2 AND v.deleted_at IS NULL`, versionID, wsID)
	v, err := scanVersion(row)
	if err != nil {
		return nil, s.mapPgError(err)
	}
	return v, nil
}

func (s *Service) getVersionTx(ctx context.Context, tx pgx.Tx, wsID, versionID int64) (*Version, error) {
	row := tx.QueryRow(ctx, `
		SELECT v.id, v.workspace_id, v.project_id, v.name, v.semver, v.description,
		       v.status, v.release_notes, v.delivered_at,
		       v.start_date, v.end_date, v.target_date, v.archived_at,
		       v.created_by, v.created_at, v.updated_at,
		       v.checklist
		FROM versions v
		WHERE v.id = $1 AND v.workspace_id = $2 AND v.deleted_at IS NULL`, versionID, wsID)
	return scanVersion(row)
}

func scanVersion(row pgx.Row) (*Version, error) {
	var v Version
	var desc, rn, sd, ed, td, arc sql.NullString
	var cbRaw []byte
	var delAt sql.NullTime

	err := row.Scan(&v.ID, &v.WorkspaceID, &v.ProjectID, &v.Name, &v.Semver,
		&desc, &v.Status, &rn, &delAt, &sd, &ed, &td, &arc,
		&v.CreatedBy, &v.CreatedAt, &v.UpdatedAt, &cbRaw)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrVersionNotFound
		}
		return nil, errs.ErrInternal.Wrap(err)
	}
	if desc.Valid {
		v.Description = desc.String
	}
	if rn.Valid {
		v.ReleaseNotes = rn.String
	}
	if delAt.Valid {
		v.DeliveredAt = &delAt.Time
	}
	if sd.Valid {
		v.StartDate = &sd.String
	}
	if ed.Valid {
		v.EndDate = &ed.String
	}
	if td.Valid {
		v.TargetDate = &td.String
	}
	if arc.Valid {
		if t, perr := time.Parse(time.RFC3339, arc.String); perr == nil {
			v.ArchivedAt = &t
		}
	}
	if len(cbRaw) > 0 {
		var items []ChecklistItem
		if err := json.Unmarshal(cbRaw, &items); err == nil {
			v.Checklist = items
		} else {
			v.Checklist = []ChecklistItem{}
		}
	} else {
		v.Checklist = []ChecklistItem{}
	}
	return &v, nil
}

func (s *Service) listSprints(ctx context.Context, wsID, versionID int64) []SprintRef {
	rows, err := s.db.Query(ctx, `
		SELECT sp.id, sp.name, sp.status, sp.start_date, sp.end_date, sp.completed_at,
		       sp.review_snapshot
		FROM sprints sp
		WHERE sp.version_id = $1 AND sp.workspace_id = $2 AND sp.deleted_at IS NULL
		ORDER BY sp.start_date NULLS LAST, sp.created_at`, versionID, wsID)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var refs []SprintRef
	for rows.Next() {
		var r SprintRef
		var start, end, completed sql.NullTime
		var reviewRaw []byte
		if err := rows.Scan(&r.SprintID, &r.Name, &r.Status, &start, &end, &completed, &reviewRaw); err != nil {
			continue
		}
		if start.Valid {
			s2 := start.Time.Format("2006-01-02")
			r.StartDate = &s2
		}
		if end.Valid {
			s2 := end.Time.Format("2006-01-02")
			r.EndDate = &s2
		}
		if completed.Valid {
			s2 := completed.Time.Format("2006-01-02")
			r.CompletedAt = &s2
		}
		if len(reviewRaw) > 0 {
			type reviewShort struct {
				CommittedPoints float64 `json:"committed_points"`
				CompletedPoints float64 `json:"completed_points"`
				CommittedIssues int     `json:"committed_issues"`
				CompletedIssues int     `json:"completed_issues"`
			}
			var snap reviewShort
			if err := json.Unmarshal(reviewRaw, &snap); err == nil {
				r.Progress = &SprintProgressRef{
					TotalPoints: snap.CommittedPoints,
					DonePoints:  snap.CompletedPoints,
					TotalIssues: snap.CommittedIssues,
					DoneIssues:  snap.CompletedIssues,
				}
			}
		}
		refs = append(refs, r)
	}
	return refs
}

// ---- 工具 ----

func (s *Service) withTx(ctx context.Context, wsID int64, fn func(tx pgx.Tx) error) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, "SELECT set_config('app.workspace_id', $1, true)", strconv.FormatInt(wsID, 10)); err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	if err := fn(tx); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (s *Service) mapPgError(err error) error {
	if err == nil {
		return nil
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			return errs.ErrVersionDataConflict
		case "23503":
			return errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "reference", Reason: "引用的资源不存在"})
		case "23514":
			if strings.Contains(pgErr.ConstraintName, "semver") {
				return errs.ErrVersionSemverInvalid.WithDetails(errs.FieldDetail{Field: "semver", Reason: pgErr.Detail})
			}
			return errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "version", Reason: pgErr.Detail})
		}
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return errs.ErrVersionNotFound
	}
	if _, ok := err.(*errs.AppError); ok {
		return err
	}
	return errs.ErrInternal.Wrap(err)
}

// ---- 校验 ----

func validateCreateInput(in CreateVersionInput) error {
	if in.WorkspaceID == 0 {
		return errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "workspace_id", Reason: "工作空间不能为空"})
	}
	if in.ProjectID == 0 {
		return errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "project_id", Reason: "项目不能为空"})
	}
	if strings.TrimSpace(in.Name) == "" || len(in.Name) > 120 {
		return errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "name", Reason: "版本名长度需在 1-120 之间"})
	}
	if strings.TrimSpace(in.Semver) == "" {
		return errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "semver", Reason: "语义版本号不能为空"})
	}
	return nil
}

func buildUpdateSet(in UpdateVersionInput, semver string, clRaw []byte) ([]string, []interface{}) {
	var sets []string
	var args []interface{}
	i := 1

	if in.Name != nil {
		sets = append(sets, fmt.Sprintf("name = $%d", i))
		args = append(args, *in.Name)
		i++
	}
	if in.Description != nil {
		sets = append(sets, fmt.Sprintf("description = $%d", i))
		args = append(args, *in.Description)
		i++
	}
	if in.StartDate != nil {
		sets = append(sets, fmt.Sprintf("start_date = $%d", i))
		args = append(args, *in.StartDate)
		i++
	}
	if in.EndDate != nil {
		sets = append(sets, fmt.Sprintf("end_date = $%d", i))
		args = append(args, *in.EndDate)
		i++
	}
	if in.TargetDate != nil {
		sets = append(sets, fmt.Sprintf("target_date = $%d", i))
		args = append(args, *in.TargetDate)
		i++
	}
	if in.Semver != nil && semver != "" {
		sets = append(sets, fmt.Sprintf("semver = $%d", i))
		args = append(args, semver)
		i++
	}
	if in.Checklist != nil {
		sets = append(sets, fmt.Sprintf("checklist = $%d", i))
		args = append(args, clRaw)
		i++
	}

	if len(sets) > 0 {
		sets = append(sets, "updated_at = now()")
	}
	return sets, args
}

// Import placeholder.
var _ = strconv.Itoa
