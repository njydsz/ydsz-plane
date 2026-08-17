// Package intake — 收件箱（匿名提报）应用服务。
package intake

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/njydsz/ydsz-plane/internal/application/issue"
	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// IssueCreator 抽象「转正」时创建正式需求/任务/缺陷的能力（由 issue.Service 实现）。
type IssueCreator interface {
	Create(ctx context.Context, in issue.CreateIssueInput) (*issue.Issue, error)
}

// Service 收件箱应用服务。
type Service struct {
	db    *pgxpool.Pool
	issue IssueCreator // 可为 nil（仅转正需要）
}

// NewService 创建收件箱服务。
func NewService(db *pgxpool.Pool) *Service {
	return &Service{db: db}
}

// WithIssueCreator 注入正式工作项创建器（转正能力）。
func (s *Service) WithIssueCreator(ic IssueCreator) *Service {
	s.issue = ic
	return s
}

// ============================================================
// 渠道 CRUD
// ============================================================

// CreateChannelInput 创建渠道入参。
type CreateChannelInput struct {
	WorkspaceID int64
	ProjectID   *int64
	Name        string
	Description string
	Slug        string
	Config      map[string]any
	CreatedBy   int64
}

// CreateChannel 创建入口渠道。
func (s *Service) CreateChannel(ctx context.Context, in CreateChannelInput) (*IntakeChannel, error) {
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "name", Reason: "渠道名称不能为空"})
	}
	slug := strings.TrimSpace(in.Slug)
	if slug == "" {
		slug = genSlug(name)
	}
	// slug 唯一性校验（同工作空间）
	var exists int
	if err := s.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM intake_channels WHERE workspace_id=$1 AND slug=$2 AND deleted=false`,
		in.WorkspaceID, slug).Scan(&exists); err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	if exists > 0 {
		return nil, errs.Conflict("INTAKE.SLUG_TAKEN", "该渠道链接标识已被使用").WithDetails(errs.FieldDetail{Field: "slug", Reason: "已被其他渠道占用"})
	}

	var ch IntakeChannel
	err := s.db.QueryRow(ctx, `
		INSERT INTO intake_channels
			(id, code, name, slug, workspace_id, project_id, description, is_active, config, status, created_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9::jsonb,'active',$10)
		RETURNING id, created_at, updated_at`,
		genID(), genCode("CH", 6), name, slug, in.WorkspaceID, in.ProjectID,
		nullIf(in.Description, ""), true, jsonOrEmpty(in.Config), in.CreatedBy,
	).Scan(&ch.ID, &ch.CreatedAt, &ch.UpdatedAt)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	ch.Name = name
	ch.Slug = slug
	ch.WorkspaceID = in.WorkspaceID
	ch.ProjectID = in.ProjectID
	ch.Description = in.Description
	ch.IsActive = true
	ch.Status = string(ChannelActive)
	ch.Config = in.Config
	ch.CreatedBy = in.CreatedBy
	return &ch, nil
}

// ListChannels 列出工作空间（可限定项目）下的渠道。
func (s *Service) ListChannels(ctx context.Context, wsID int64, projectID *int64, onlyActive bool) ([]IntakeChannel, error) {
	query := `
		SELECT c.id, coalesce(c.code,''), c.name, c.slug, c.workspace_id, c.project_id,
		       coalesce(c.description,''), c.is_active, coalesce(c.config,'{}'::jsonb), c.status,
		       (SELECT COUNT(*) FROM intake_issues i WHERE i.channel_id=c.id AND i.deleted=false) AS issue_count,
		       c.created_by, c.created_at, c.updated_at
		FROM intake_channels c
		WHERE c.workspace_id=$1 AND c.deleted=false`
	args := []any{wsID}
	if projectID != nil {
		query += ` AND c.project_id=$2`
		args = append(args, *projectID)
	}
	if onlyActive {
		query += ` AND c.is_active=true`
	}
	query += ` ORDER BY c.created_at DESC`

	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()

	channels := make([]IntakeChannel, 0)
	for rows.Next() {
		var ch IntakeChannel
		var config []byte
		if err := rows.Scan(&ch.ID, &ch.Code, &ch.Name, &ch.Slug, &ch.WorkspaceID, &ch.ProjectID,
			&ch.Description, &ch.IsActive, &config, &ch.Status, &ch.IssueCount,
			&ch.CreatedBy, &ch.CreatedAt, &ch.UpdatedAt); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
		ch.Config = jsonMap(config)
		channels = append(channels, ch)
	}
	return channels, nil
}

// GetChannel 获取单个渠道。
func (s *Service) GetChannel(ctx context.Context, wsID, channelID int64) (*IntakeChannel, error) {
	var ch IntakeChannel
	var config []byte
	err := s.db.QueryRow(ctx, `
		SELECT id, coalesce(code,''), name, slug, workspace_id, project_id,
		       coalesce(description,''), is_active, coalesce(config,'{}'::jsonb), status,
		       (SELECT COUNT(*) FROM intake_issues i WHERE i.channel_id=id AND i.deleted=false),
		       created_by, created_at, updated_at
		FROM intake_channels
		WHERE id=$1 AND workspace_id=$2 AND deleted=false`,
		channelID, wsID,
	).Scan(&ch.ID, &ch.Code, &ch.Name, &ch.Slug, &ch.WorkspaceID, &ch.ProjectID,
		&ch.Description, &ch.IsActive, &config, &ch.Status, &ch.IssueCount,
		&ch.CreatedBy, &ch.CreatedAt, &ch.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errs.ErrNotFound.From()
		}
		return nil, errs.ErrInternal.Wrap(err)
	}
	ch.Config = jsonMap(config)
	return &ch, nil
}

// UpdateChannelInput 更新渠道入参。
type UpdateChannelInput struct {
	ID          int64
	WorkspaceID int64
	Name        *string
	Description *string
	Slug        *string
	IsActive    *bool
	ProjectID   *int64
}

// UpdateChannel 更新渠道。
func (s *Service) UpdateChannel(ctx context.Context, in UpdateChannelInput) (*IntakeChannel, error) {
	cur, err := s.GetChannel(ctx, in.WorkspaceID, in.ID)
	if err != nil {
		return nil, err
	}
	if in.Name != nil {
		if strings.TrimSpace(*in.Name) == "" {
			return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "name", Reason: "渠道名称不能为空"})
		}
		cur.Name = *in.Name
	}
	if in.Description != nil {
		cur.Description = *in.Description
	}
	if in.Slug != nil {
		slug := strings.TrimSpace(*in.Slug)
		if slug == "" {
			return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "slug", Reason: "链接标识不能为空"})
		}
		var exists int
		if err := s.db.QueryRow(ctx,
			`SELECT COUNT(*) FROM intake_channels WHERE workspace_id=$1 AND slug=$2 AND id<>$3 AND deleted=false`,
			in.WorkspaceID, slug, in.ID).Scan(&exists); err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
		if exists > 0 {
			return nil, errs.Conflict("INTAKE.SLUG_TAKEN", "该渠道链接标识已被使用")
		}
		cur.Slug = slug
	}
	if in.IsActive != nil {
		cur.IsActive = *in.IsActive
	}
	if in.ProjectID != nil {
		cur.ProjectID = in.ProjectID
	}

	_, err = s.db.Exec(ctx, `
		UPDATE intake_channels SET name=$1, description=$2, slug=$3, is_active=$4, project_id=$5,
		       updated_at=now()
		WHERE id=$6 AND workspace_id=$7 AND deleted=false`,
		cur.Name, nullIf(cur.Description, ""), cur.Slug, cur.IsActive, cur.ProjectID,
		in.ID, in.WorkspaceID)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	return s.GetChannel(ctx, in.WorkspaceID, in.ID)
}

// DeleteChannel 归档渠道（软删除）。
func (s *Service) DeleteChannel(ctx context.Context, wsID, channelID int64) error {
	tag, err := s.db.Exec(ctx,
		`UPDATE intake_channels SET deleted=true, updated_at=now() WHERE id=$1 AND workspace_id=$2 AND deleted=false`,
		channelID, wsID)
	if err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	if tag.RowsAffected() == 0 {
		return errs.ErrNotFound.From()
	}
	return nil
}

// GetChannelBySlug 公开：按 slug 获取启用中的渠道。
func (s *Service) GetChannelBySlug(ctx context.Context, slug string) (*IntakeChannel, error) {
	var ch IntakeChannel
	var config []byte
	err := s.db.QueryRow(ctx, `
		SELECT id, coalesce(code,''), name, slug, workspace_id, project_id,
		       coalesce(description,''), is_active, coalesce(config,'{}'::jsonb), status,
		       0, created_by, created_at, updated_at
		FROM intake_channels
		WHERE slug=$1 AND deleted=false AND is_active=true`,
		slug,
	).Scan(&ch.ID, &ch.Code, &ch.Name, &ch.Slug, &ch.WorkspaceID, &ch.ProjectID,
		&ch.Description, &ch.IsActive, &config, &ch.Status, &ch.IssueCount,
		&ch.CreatedBy, &ch.CreatedAt, &ch.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errs.ErrNotFound.WithCodeMessage("INTAKE.CHANNEL_NOT_FOUND", "提报渠道不存在或已停用")
		}
		return nil, errs.ErrInternal.Wrap(err)
	}
	ch.Config = jsonMap(config)
	return &ch, nil
}

// ============================================================
// 工单
// ============================================================

// SubmitIssueInput 提交工单入参（认证与公开共用）。
type SubmitIssueInput struct {
	WorkspaceID    int64  `json:"-"`
	ChannelID      int64  `json:"channel_id"`
	ChannelSlug    string `json:"channel_slug"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	SubmitterName  string `json:"submitter_name"`
	SubmitterEmail string `json:"submitter_email"`
	Priority       string `json:"priority"`
	CreatedBy      int64  `json:"-"`
}

// SubmitIssue 提交工单（渠道定位：优先 channel_id，其次 channel_slug）。
func (s *Service) SubmitIssue(ctx context.Context, in SubmitIssueInput) (*IntakeIssue, error) {
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "name", Reason: "标题不能为空"})
	}
	email := strings.TrimSpace(in.SubmitterEmail)
	if email == "" {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "submitter_email", Reason: "联系邮箱不能为空"})
	}
	priority := strings.TrimSpace(in.Priority)
	if priority == "" {
		priority = "medium"
	}
	if !validPriority(priority) {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "priority", Reason: "无效的优先级"})
	}

	// 定位渠道
	var ch IntakeChannel
	if in.ChannelID > 0 {
		if in.WorkspaceID <= 0 {
			return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "channel_id", Reason: "认证提交必须携带工作空间上下文"})
		}
		c, err := s.GetChannel(ctx, in.WorkspaceID, in.ChannelID)
		if err != nil {
			return nil, err
		}
		ch = *c
	} else if in.ChannelSlug != "" {
		c, err := s.GetChannelBySlug(ctx, in.ChannelSlug)
		if err != nil {
			return nil, err
		}
		ch = *c
	} else {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "channel", Reason: "缺少渠道标识（channel_id 或 channel_slug）"})
	}

	iss := &IntakeIssue{}
	err := s.db.QueryRow(ctx, `
		INSERT INTO intake_issues
			(id, code, name, workspace_id, project_id, channel_id, tracking_id,
			 submitter_name, submitter_email, description, priority, status, created_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,'open',$12)
		RETURNING id, created_at, updated_at`,
		genID(), genCode("IN", 8), name, ch.WorkspaceID, ch.ProjectID, ch.ID,
		newTrackingID(), nullIf(in.SubmitterName, ""), email,
		in.Description, priority, in.CreatedBy,
	).Scan(&iss.ID, &iss.CreatedAt, &iss.UpdatedAt)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	iss.Name = name
	iss.WorkspaceID = ch.WorkspaceID
	iss.ProjectID = ch.ProjectID
	iss.ChannelID = ch.ID
	iss.ChannelName = ch.Name
	iss.SubmitterName = in.SubmitterName
	iss.SubmitterEmail = email
	iss.Description = in.Description
	iss.Priority = priority
	iss.Status = IssueOpen
	return iss, nil
}

// ListIssuesFilter 工单列表筛选。
type ListIssuesFilter struct {
	WorkspaceID int64
	ProjectID   *int64
	ChannelID   *int64
	Status      string // open|accepted|rejected|archived；空=全部
	Limit       int
	Offset      int
}

// ListIssues 列出工单。
func (s *Service) ListIssues(ctx context.Context, f ListIssuesFilter) ([]IntakeIssue, int, error) {
	where := []string{"i.workspace_id=$1", "i.deleted=false"}
	args := []any{f.WorkspaceID}
	idx := 2
	if f.ProjectID != nil {
		where = append(where, "i.project_id=$"+strconv.Itoa(idx))
		args = append(args, *f.ProjectID)
		idx++
	}
	if f.ChannelID != nil {
		where = append(where, "i.channel_id=$"+strconv.Itoa(idx))
		args = append(args, *f.ChannelID)
		idx++
	}
	if f.Status != "" {
		where = append(where, "i.status=$"+strconv.Itoa(idx))
		args = append(args, f.Status)
		idx++
	}
	cond := strings.Join(where, " AND ")

	var total int
	if err := s.db.QueryRow(ctx, `SELECT COUNT(*) FROM intake_issues i WHERE `+cond, args...).Scan(&total); err != nil {
		return nil, 0, errs.ErrInternal.Wrap(err)
	}

	limit, offset := f.Limit, f.Offset
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	args = append(args, limit, offset)

	rows, err := s.db.Query(ctx, `
		SELECT i.id, coalesce(i.code,''), i.name, i.workspace_id, i.project_id, i.channel_id,
		       coalesce(c.name,''), coalesce(i.tracking_id,''), coalesce(i.submitter_name,''),
		       i.submitter_email, coalesce(i.description,''), i.priority, i.status,
		       coalesce(i.linked_entity_type,''), i.linked_entity_id,
		       i.resolved_at, i.resolved_by, i.created_at, i.updated_at
		FROM intake_issues i
		LEFT JOIN intake_channels c ON c.id=i.channel_id
		WHERE `+cond+` ORDER BY i.created_at DESC LIMIT $`+strconv.Itoa(idx)+` OFFSET $`+strconv.Itoa(idx+1),
		args...)
	if err != nil {
		return nil, 0, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()

	issues := make([]IntakeIssue, 0)
	for rows.Next() {
		var it IntakeIssue
		if err := rows.Scan(&it.ID, &it.Code, &it.Name, &it.WorkspaceID, &it.ProjectID, &it.ChannelID,
			&it.ChannelName, &it.TrackingID, &it.SubmitterName, &it.SubmitterEmail, &it.Description,
			&it.Priority, &it.Status, &it.LinkedEntityType, &it.LinkedEntityID,
			&it.ResolvedAt, &it.ResolvedBy, &it.CreatedAt, &it.UpdatedAt); err != nil {
			return nil, 0, errs.ErrInternal.Wrap(err)
		}
		issues = append(issues, it)
	}
	return issues, total, nil
}

// GetIssue 获取单个工单。
func (s *Service) GetIssue(ctx context.Context, wsID, issueID int64) (*IntakeIssue, error) {
	var it IntakeIssue
	err := s.db.QueryRow(ctx, `
		SELECT i.id, coalesce(i.code,''), i.name, i.workspace_id, i.project_id, i.channel_id,
		       coalesce(c.name,''), coalesce(i.tracking_id,''), coalesce(i.submitter_name,''),
		       i.submitter_email, coalesce(i.description,''), i.priority, i.status,
		       coalesce(i.linked_entity_type,''), i.linked_entity_id,
		       coalesce(li.identifier,''), i.resolved_at, i.resolved_by, i.created_at, i.updated_at
		FROM intake_issues i
		LEFT JOIN intake_channels c ON c.id=i.channel_id
		LEFT JOIN LATERAL (
			SELECT (p.identifier || '-' || i2.sequence_id) AS identifier
			FROM (SELECT id, public_id, workspace_id, project_id, sequence_id, 'requirement'::text AS type_code, parent_id, depth, name, description_json, description_html, description_stripped, state_id, priority, NULL::smallint AS severity, NULL::text AS found_phase, NULL::text AS root_cause_category, NULL::bigint AS verifier_id, NULL::jsonb AS environment, NULL::jsonb AS reproduce_steps, NULL::text AS category, NULL::numeric AS actual_effort, NULL::numeric AS remaining_effort, NULL::text AS delay_reason, source, point, sprint_id, progress, start_date, target_date, completed_at, is_draft, sort_order, version, version_id, NULL::bigint AS found_version_id, NULL::bigint AS fix_version_id, created_by, created_at, updated_at, deleted FROM requirement WHERE deleted=false
			 UNION ALL SELECT id, public_id, workspace_id, project_id, sequence_id, 'task'::text, parent_id, depth, name, description_json, description_html, description_stripped, state_id, priority, NULL::smallint, NULL::text, NULL::text, NULL::bigint, NULL::jsonb, NULL::jsonb, category, actual_effort, remaining_effort, delay_reason, NULL::text, point, sprint_id, progress, start_date, target_date, completed_at, is_draft, sort_order, version, version_id, NULL::bigint, NULL::bigint, created_by, created_at, updated_at, deleted FROM task WHERE deleted=false
			 UNION ALL SELECT id, public_id, workspace_id, project_id, sequence_id, 'defect'::text, parent_id, depth, name, description_json, description_html, description_stripped, state_id, priority, severity, found_phase, root_cause_category, verifier_id, environment, reproduce_steps, NULL::text, NULL::numeric, NULL::numeric, NULL::text, NULL::text, point, sprint_id, progress, start_date, target_date, completed_at, is_draft, sort_order, version, version_id, found_version_id, fix_version_id, created_by, created_at, updated_at, deleted FROM defect WHERE deleted=false
			) i2 JOIN projects p ON p.id=i2.project_id
			WHERE i2.id=i.linked_entity_id AND i2.workspace_id=i.workspace_id
			LIMIT 1
		) li ON true
		WHERE i.id=$1 AND i.workspace_id=$2 AND i.deleted=false`,
		issueID, wsID,
	).Scan(&it.ID, &it.Code, &it.Name, &it.WorkspaceID, &it.ProjectID, &it.ChannelID,
		&it.ChannelName, &it.TrackingID, &it.SubmitterName, &it.SubmitterEmail, &it.Description,
		&it.Priority, &it.Status, &it.LinkedEntityType, &it.LinkedEntityID, &it.LinkedEntityIdent,
		&it.ResolvedAt, &it.ResolvedBy, &it.CreatedAt, &it.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errs.ErrNotFound.From()
		}
		return nil, errs.ErrInternal.Wrap(err)
	}
	return &it, nil
}

// setStatus 统一状态流转。
func (s *Service) setStatus(ctx context.Context, wsID, issueID int64, to IssueStatus, operator int64) (*IntakeIssue, error) {
	cur, err := s.GetIssue(ctx, wsID, issueID)
	if err != nil {
		return nil, err
	}
	allowed := map[IssueStatus]map[IssueStatus]bool{
		IssueOpen:     {IssueAccepted: true, IssueRejected: true, IssueArchived: true},
		IssueAccepted: {IssueArchived: true},
		IssueRejected: {IssueArchived: true},
	}
	if !allowed[cur.Status][to] {
		return nil, errs.ErrInvalidTransition.WithDetails(errs.FieldDetail{
			Field: "status", Reason: "当前状态不允许该流转",
		})
	}
	if _, err := s.db.Exec(ctx, `
		UPDATE intake_issues SET status=$1::intake_issue_status, resolved_at=now(), resolved_by=$2, updated_at=now()
		WHERE id=$3 AND workspace_id=$4 AND deleted=false`,
		string(to), operator, issueID, wsID); err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	return s.GetIssue(ctx, wsID, issueID)
}

// AcceptIssue 接受工单。
func (s *Service) AcceptIssue(ctx context.Context, wsID, issueID, operator int64) (*IntakeIssue, error) {
	return s.setStatus(ctx, wsID, issueID, IssueAccepted, operator)
}

// RejectIssue 拒绝工单。
func (s *Service) RejectIssue(ctx context.Context, wsID, issueID, operator int64) (*IntakeIssue, error) {
	return s.setStatus(ctx, wsID, issueID, IssueRejected, operator)
}

// ArchiveIssue 归档工单。
func (s *Service) ArchiveIssue(ctx context.Context, wsID, issueID, operator int64) (*IntakeIssue, error) {
	return s.setStatus(ctx, wsID, issueID, IssueArchived, operator)
}

// PromoteIssueInput 转正入参。
type PromoteIssueInput struct {
	WorkspaceID int64
	IssueID     int64
	Operator    int64
	TypeCode    string  // requirement|task|defect，默认 requirement
	Severity    *int    // defect 必填 1-5
	FoundPhase  *string // defect 必填
	ProjectID   *int64  // 覆盖目标项目（默认渠道所属项目）
}

// PromoteIssue 转正：将已接受工单转为正式需求/任务/缺陷，并回写关联。
func (s *Service) PromoteIssue(ctx context.Context, in PromoteIssueInput) (*IntakeIssue, error) {
	if s.issue == nil {
		return nil, errs.ErrNotImplemented.WithDetails(errs.FieldDetail{Field: "promote", Reason: "转正服务未装配"})
	}
	cur, err := s.GetIssue(ctx, in.WorkspaceID, in.IssueID)
	if err != nil {
		return nil, err
	}
	if cur.Status != IssueAccepted {
		return nil, errs.ErrInvalidTransition.WithDetails(errs.FieldDetail{
			Field: "status", Reason: "仅已接受（accepted）的工单可以转正",
		})
	}

	typeCode := strings.TrimSpace(in.TypeCode)
	if typeCode == "" {
		typeCode = "requirement"
	}
	if !validType(typeCode) {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "type_code", Reason: "无效的工作项类型"})
	}

	projectID := in.ProjectID
	if projectID == nil {
		projectID = cur.ProjectID
	}
	if projectID == nil {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "project_id", Reason: "工单未归属项目，请指定目标项目"})
	}

	createIn := issue.CreateIssueInput{
		WorkspaceID:     in.WorkspaceID,
		ProjectID:       *projectID,
		TypeCode:        issue.IssueTypeCode(typeCode),
		Name:            cur.Name,
		DescriptionHTML: cur.Description,
		Priority:        issue.IssuePriority(cur.Priority),
		CreatedBy:       in.Operator,
		Source:          ptrStr("intake"),
	}
	if typeCode == "defect" {
		if in.Severity == nil || in.FoundPhase == nil {
			return nil, errs.ErrValidation.WithDetails(
				errs.FieldDetail{Field: "severity", Reason: "缺陷转正必须提供严重级别"},
				errs.FieldDetail{Field: "found_phase", Reason: "缺陷转正必须提供发现阶段"},
			)
		}
		createIn.Severity = in.Severity
		createIn.FoundPhase = in.FoundPhase
	}

	created, err := s.issue.Create(ctx, createIn)
	if err != nil {
		return nil, err
	}

	if _, err := s.db.Exec(ctx, `
		UPDATE intake_issues SET linked_entity_type=$1, linked_entity_id=$2, updated_at=now()
		WHERE id=$3 AND workspace_id=$4 AND deleted=false`,
		typeCode, created.ID, in.IssueID, in.WorkspaceID); err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	return s.GetIssue(ctx, in.WorkspaceID, in.IssueID)
}

// TrackIssue 公开：按 tracking_id + 提交邮箱查询工单状态。
func (s *Service) TrackIssue(ctx context.Context, trackingID, email string) (*IntakeIssue, error) {
	var it IntakeIssue
	err := s.db.QueryRow(ctx, `
		SELECT i.id, coalesce(i.code,''), i.name, i.workspace_id, i.project_id, i.channel_id,
		       coalesce(c.name,''), coalesce(i.tracking_id,''), coalesce(i.submitter_name,''),
		       i.submitter_email, coalesce(i.description,''), i.priority, i.status,
		       coalesce(i.linked_entity_type,''), i.linked_entity_id,
		       i.resolved_at, i.resolved_by, i.created_at, i.updated_at
		FROM intake_issues i
		LEFT JOIN intake_channels c ON c.id=i.channel_id
		WHERE i.tracking_id=$1 AND lower(i.submitter_email)=lower($2) AND i.deleted=false`,
		trackingID, email,
	).Scan(&it.ID, &it.Code, &it.Name, &it.WorkspaceID, &it.ProjectID, &it.ChannelID,
		&it.ChannelName, &it.TrackingID, &it.SubmitterName, &it.SubmitterEmail, &it.Description,
		&it.Priority, &it.Status, &it.LinkedEntityType, &it.LinkedEntityID,
		&it.ResolvedAt, &it.ResolvedBy, &it.CreatedAt, &it.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errs.ErrNotFound.WithCodeMessage("INTAKE.NOT_FOUND", "未找到对应的提报记录")
		}
		return nil, errs.ErrInternal.Wrap(err)
	}
	return &it, nil
}

// ============================================================
// Helpers
// ============================================================

func validPriority(p string) bool {
	switch p {
	case "urgent", "high", "medium", "low", "none":
		return true
	}
	return false
}

func validType(t string) bool {
	switch t {
	case "requirement", "task", "defect":
		return true
	}
	return false
}

func nullIf(s, empty string) any {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return s
}

func jsonOrEmpty(m map[string]any) any {
	if m == nil {
		return "{}"
	}
	b, _ := json.Marshal(m)
	return string(b)
}

func jsonMap(b []byte) map[string]any {
	m := map[string]any{}
	if len(b) > 0 {
		_ = json.Unmarshal(b, &m)
	}
	return m
}

func ptrStr(s string) *string { return &s }

func genCode(prefix string, n int) string {
	b := make([]byte, n/2+1)
	_, _ = rand.Read(b)
	return prefix + strings.ToUpper(hex.EncodeToString(b)[:n])
}

func newTrackingID() string {
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	return "INC-" + strings.ToUpper(hex.EncodeToString(b))
}

func genSlug(name string) string {
	b := make([]byte, 3)
	_, _ = rand.Read(b)
	return "channel-" + hex.EncodeToString(b)
}
