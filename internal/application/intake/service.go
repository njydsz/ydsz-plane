// Package intake — Intake 收件箱应用服务：通道管理、工单提交、转正流程。
package intake

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// Service 提供 Intake 通道管理与收件工单生命周期。
type Service struct {
	db *pgxpool.Pool
}

// NewService 创建 Intake 服务。
func NewService(db *pgxpool.Pool) *Service {
	return &Service{db: db}
}

// --- Channel CRUD ---

// CreateChannelInput 创建通道入参。
type CreateChannelInput struct {
	WorkspaceID      int64
	ProjectID        *int64
	Slug             string
	Name             string
	Description      string
	IsPublic         bool
	DefaultIssueType string
	DefaultPriority  int16
	AutoAssignRules  json.RawMessage
	RateLimitPerMin  int16
	RequireCaptcha   bool
	CustomFields     json.RawMessage
	Branding         json.RawMessage
	NotifyOnSubmit   bool
	NotifyUsers      []int64
	CreatedBy        int64
}

// CreateChannel 创建表单通道。
func (s *Service) CreateChannel(ctx context.Context, input CreateChannelInput) (*Channel, error) {
	if input.Slug == "" || input.Name == "" {
		return nil, errs.Validation("INTAKE.CHANNEL_REQUIRED", "slug 和 name 不能为空")
	}
	if !isValidSlug(input.Slug) {
		return nil, errs.Validation("INTAKE.BAD_SLUG", "slug 只允许字母、数字、连字符、下划线")
	}
	if input.DefaultIssueType == "" {
		input.DefaultIssueType = IssueTypeRequirement
	}
	// 确保 JSON 非空
	if len(input.AutoAssignRules) == 0 {
		input.AutoAssignRules = json.RawMessage("[]")
	}
	if len(input.CustomFields) == 0 {
		input.CustomFields = json.RawMessage("[]")
	}
	if len(input.Branding) == 0 {
		input.Branding = json.RawMessage("{}")
	}

	var pid *int64
	if input.ProjectID != nil {
		pid = input.ProjectID
	}

	var c Channel
	err := s.db.QueryRow(ctx, `
		INSERT INTO intake_channels
			(workspace_id, project_id, slug, name, description, is_public,
			 default_issue_type, default_priority, auto_assign_rules,
			 rate_limit_per_min, require_captcha, custom_fields, branding,
			 notify_on_submit, notify_users, is_active, created_by, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,true,$16,NOW(),NOW())
		RETURNING id, workspace_id, project_id, slug, name, description, is_public,
			default_issue_type, default_priority, auto_assign_rules,
			rate_limit_per_min, require_captcha, custom_fields, branding,
			notify_on_submit, notify_users, is_active, created_by, created_at, updated_at`,
		input.WorkspaceID, pid, input.Slug, input.Name, input.Description, input.IsPublic,
		input.DefaultIssueType, input.DefaultPriority, input.AutoAssignRules,
		input.RateLimitPerMin, input.RequireCaptcha, input.CustomFields, input.Branding,
		input.NotifyOnSubmit, input.NotifyUsers, input.CreatedBy,
	).Scan(
		&c.ID, &c.WorkspaceID, &c.ProjectID, &c.Slug, &c.Name, &c.Description, &c.IsPublic,
		&c.DefaultIssueType, &c.DefaultPriority, &c.AutoAssignRules,
		&c.RateLimitPerMin, &c.RequireCaptcha, &c.CustomFields, &c.Branding,
		&c.NotifyOnSubmit, &c.NotifyUsers, &c.IsActive, &c.CreatedBy, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return nil, errs.Validation("INTAKE.DUPLICATE_SLUG", "该 slug 已被占用")
		}
		return nil, fmt.Errorf("intake.CreateChannel: %w", err)
	}
	return &c, nil
}

// ListChannelsInput 查询入参。
type ListChannelsInput struct {
	WorkspaceID int64
	ProjectID   *int64
	Limit       int
	Offset      int
}

// ListChannelsResult 查询结果。
type ListChannelsResult struct {
	Items []Channel `json:"items"`
	Total int64     `json:"total"`
}

// ListChannels 列出通道。
func (s *Service) ListChannels(ctx context.Context, input ListChannelsInput) (*ListChannelsResult, error) {
	if input.Limit <= 0 {
		input.Limit = 20
	}
	if input.Limit > 100 {
		input.Limit = 100
	}

	var where string
	var args []interface{}
	if input.ProjectID != nil {
		where = "WHERE workspace_id = $1 AND project_id = $2"
		args = append(args, input.WorkspaceID, *input.ProjectID)
	} else {
		where = "WHERE workspace_id = $1"
		args = append(args, input.WorkspaceID)
	}

	var total int64
	if err := s.db.QueryRow(ctx, "SELECT COUNT(*) FROM intake_channels "+where, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("intake.ListChannels count: %w", err)
	}

	query := strings.Replace(where, "WHERE", "WHERE", 1)
	query = fmt.Sprintf(`
		SELECT id, workspace_id, project_id, slug, name, description, is_public,
			default_issue_type, default_priority, auto_assign_rules,
			rate_limit_per_min, require_captcha, custom_fields, branding,
			notify_on_submit, notify_users, is_active, created_by, created_at, updated_at
		FROM intake_channels %s ORDER BY id DESC LIMIT $%d OFFSET $%d`, query, len(args)+1, len(args)+2)

	rows, err := s.db.Query(ctx, query, append(args, input.Limit, input.Offset)...)
	if err != nil {
		return nil, fmt.Errorf("intake.ListChannels: %w", err)
	}
	defer rows.Close()

	var items []Channel
	for rows.Next() {
		var c Channel
		if err := rows.Scan(
			&c.ID, &c.WorkspaceID, &c.ProjectID, &c.Slug, &c.Name, &c.Description, &c.IsPublic,
			&c.DefaultIssueType, &c.DefaultPriority, &c.AutoAssignRules,
			&c.RateLimitPerMin, &c.RequireCaptcha, &c.CustomFields, &c.Branding,
			&c.NotifyOnSubmit, &c.NotifyUsers, &c.IsActive, &c.CreatedBy, &c.CreatedAt, &c.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("intake.ListChannels scan: %w", err)
		}
		items = append(items, c)
	}
	return &ListChannelsResult{Items: items, Total: total}, nil
}

// GetChannel 读取通道详情。
func (s *Service) GetChannel(ctx context.Context, workspaceID, channelID int64) (*Channel, error) {
	var c Channel
	err := s.db.QueryRow(ctx, `
		SELECT id, workspace_id, project_id, slug, name, description, is_public,
			default_issue_type, default_priority, auto_assign_rules,
			rate_limit_per_min, require_captcha, custom_fields, branding,
			notify_on_submit, notify_users, is_active, created_by, created_at, updated_at
		FROM intake_channels WHERE id = $1 AND workspace_id = $2`,
		channelID, workspaceID,
	).Scan(
		&c.ID, &c.WorkspaceID, &c.ProjectID, &c.Slug, &c.Name, &c.Description, &c.IsPublic,
		&c.DefaultIssueType, &c.DefaultPriority, &c.AutoAssignRules,
		&c.RateLimitPerMin, &c.RequireCaptcha, &c.CustomFields, &c.Branding,
		&c.NotifyOnSubmit, &c.NotifyUsers, &c.IsActive, &c.CreatedBy, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.NotFound("INTAKE.CHANNEL_NOT_FOUND", "通道不存在")
		}
		return nil, fmt.Errorf("intake.GetChannel: %w", err)
	}
	return &c, nil
}

// 公开查询（免登录）：通过 slug 获取公开通道。
func (s *Service) GetChannelBySlug(ctx context.Context, workspaceID int64, slug string) (*Channel, error) {
	var c Channel
	err := s.db.QueryRow(ctx, `
		SELECT id, workspace_id, project_id, slug, name, description, is_public,
			default_issue_type, default_priority, auto_assign_rules,
			rate_limit_per_min, require_captcha, custom_fields, branding,
			notify_on_submit, notify_users, is_active, created_by, created_at, updated_at
		FROM intake_channels WHERE workspace_id = $1 AND slug = $2 AND is_public = true AND is_active = true`,
		workspaceID, slug,
	).Scan(
		&c.ID, &c.WorkspaceID, &c.ProjectID, &c.Slug, &c.Name, &c.Description, &c.IsPublic,
		&c.DefaultIssueType, &c.DefaultPriority, &c.AutoAssignRules,
		&c.RateLimitPerMin, &c.RequireCaptcha, &c.CustomFields, &c.Branding,
		&c.NotifyOnSubmit, &c.NotifyUsers, &c.IsActive, &c.CreatedBy, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		return nil, errs.NotFound("INTAKE.CHANNEL_NOT_FOUND", "该通道不存在或未公开")
	}
	return &c, nil
}

// UpdateChannel 更新通道。
func (s *Service) UpdateChannel(ctx context.Context, workspaceID, channelID int64, input UpdateChannelInput) (*Channel, error) {
	// 动态 SET。
	var sets []string
	var args []interface{}
	argIdx := 1

	appendSet := func(cond string, val interface{}) {
		sets = append(sets, fmt.Sprintf(cond, argIdx))
		args = append(args, val)
		argIdx++
	}

	if input.Name != nil {
		appendSet("name = $%d", *input.Name)
	}
	if input.Description != nil {
		appendSet("description = $%d", *input.Description)
	}
	if input.IsPublic != nil {
		appendSet("is_public = $%d", *input.IsPublic)
	}
	if input.IsActive != nil {
		appendSet("is_active = $%d", *input.IsActive)
	}
	if input.AutoAssignRules != nil {
		appendSet("auto_assign_rules = $%d", input.AutoAssignRules)
	}

	if len(sets) == 0 {
		return s.GetChannel(ctx, workspaceID, channelID)
	}

	sets = append(sets, "updated_at = NOW()")
	query := fmt.Sprintf("UPDATE intake_channels SET %s WHERE id = $%d AND workspace_id = $%d",
		strings.Join(sets, ", "), argIdx, argIdx+1)
	args = append(args, channelID, workspaceID)

	if _, err := s.db.Exec(ctx, query, args...); err != nil {
		return nil, fmt.Errorf("intake.UpdateChannel: %w", err)
	}
	return s.GetChannel(ctx, workspaceID, channelID)
}

// UpdateChannelInput 更新入参。
type UpdateChannelInput struct {
	Name            *string         `json:"name,omitempty"`
	Description     *string         `json:"description,omitempty"`
	IsPublic        *bool           `json:"is_public,omitempty"`
	IsActive        *bool           `json:"is_active,omitempty"`
	AutoAssignRules json.RawMessage `json:"auto_assign_rules,omitempty"`
}

// DeleteChannel 删除通道。
func (s *Service) DeleteChannel(ctx context.Context, workspaceID, channelID int64) error {
	tag, err := s.db.Exec(ctx, `DELETE FROM intake_channels WHERE id = $1 AND workspace_id = $2`, channelID, workspaceID)
	if err != nil {
		return fmt.Errorf("intake.DeleteChannel: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return errs.NotFound("INTAKE.CHANNEL_NOT_FOUND", "通道不存在")
	}
	return nil
}

// --- Issue（收件工单）操作 ---

// SubmitInput 公开提交入参（外部用户）。
type SubmitInput struct {
	ChannelID     int64
	WorkspaceID   int64
	Title         string
	Description   string
	SubmitterName string
	SubmitterEmail string
	IssueType     string
	Priority      int16
	CustomFields  json.RawMessage
	AttachmentIDs []int64
	SubmitterUserID *int64
}

// SubmitIssue 创建收件工单。
// 由公开表单调用（免登录）。tracking_id 自动生成。
func (s *Service) SubmitIssue(ctx context.Context, input SubmitInput) (*Issue, error) {
	if input.Title == "" || input.SubmitterName == "" || input.SubmitterEmail == "" {
		return nil, errs.Validation("INTAKE.FIELD_REQUIRED", "标题、姓名、邮箱必填")
	}
	trackingID := generateTrackingID()

	if len(input.CustomFields) == 0 {
		input.CustomFields = json.RawMessage("{}")
	}
	if input.IssueType == "" {
		input.IssueType = IssueTypeRequirement
	}

	var is Issue
	err := s.db.QueryRow(ctx, `
		INSERT INTO intake_issues
			(channel_id, workspace_id, tracking_id, submitter_name, submitter_email, submitter_user_id,
			 title, description, issue_type, priority, custom_fields, attachment_ids,
			 status, notify_on_status, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,'open',true,NOW(),NOW())
		RETURNING id, channel_id, workspace_id, project_id, tracking_id,
			submitter_name, submitter_email, submitter_user_id,
			title, description, issue_type, priority, custom_fields, attachment_ids,
			status, status_reason, converted_issue_id, assigned_to,
			reviewed_by, reviewed_at, notify_on_status, created_at, updated_at`,
		input.ChannelID, input.WorkspaceID, trackingID,
		input.SubmitterName, input.SubmitterEmail, input.SubmitterUserID,
		input.Title, input.Description, input.IssueType, input.Priority,
		input.CustomFields, input.AttachmentIDs,
	).Scan(
		&is.ID, &is.ChannelID, &is.WorkspaceID, &is.ProjectID, &is.TrackingID,
		&is.SubmitterName, &is.SubmitterEmail, &is.SubmitterUserID,
		&is.Title, &is.Description, &is.IssueType, &is.Priority,
		&is.CustomFields, &is.AttachmentIDs,
		&is.Status, &is.StatusReason, &is.ConvertedIssueID, &is.AssignedTo,
		&is.ReviewedBy, &is.ReviewedAt, &is.NotifyOnStatus, &is.CreatedAt, &is.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("intake.SubmitIssue: %w", err)
	}
	return &is, nil
}

// 查询收件工单列表（管理员视图）。
type ListIssuesInput struct {
	WorkspaceID int64
	ChannelID   *int64
	Status      *string
	AssignedTo  *int64
	Limit       int
	Offset      int
}

// ListIssuesResult 列表结果。
type ListIssuesResult struct {
	Items []Issue `json:"items"`
	Total int64   `json:"total"`
}

// ListIssues 查询收件工单。
func (s *Service) ListIssues(ctx context.Context, input ListIssuesInput) (*ListIssuesResult, error) {
	if input.Limit <= 0 {
		input.Limit = 20
	}
	if input.Limit > 100 {
		input.Limit = 100
	}

	var args []interface{}
	var where string
	argIdx := 1
	conjunction := func(cond string, val interface{}) {
		prefix := "WHERE"
		if where != "" {
			prefix = "AND"
		}
		where = fmt.Sprintf("%s %s %s $%d", where, prefix, cond, argIdx)
		args = append(args, val)
		argIdx++
	}

	conjunction("workspace_id =", input.WorkspaceID)
	if input.ChannelID != nil {
		conjunction("channel_id =", *input.ChannelID)
	}
	if input.Status != nil {
		conjunction("status =", *input.Status)
	}
	if input.AssignedTo != nil {
		conjunction("assigned_to =", *input.AssignedTo)
	}

	var total int64
	if err := s.db.QueryRow(ctx, "SELECT COUNT(*) FROM intake_issues "+where, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("intake.ListIssues count: %w", err)
	}

	query := strings.Replace(where, "WHERE", "", 1)
	if query == "" {
		query = "ORDER BY id DESC"
	} else {
		query += " ORDER BY id DESC"
	}
	query = fmt.Sprintf(`
		SELECT id, channel_id, workspace_id, project_id, tracking_id,
			submitter_name, submitter_email, submitter_user_id,
			title, description, issue_type, priority, custom_fields, attachment_ids,
			status, status_reason, converted_issue_id, assigned_to,
			reviewed_by, reviewed_at, notify_on_status, created_at, updated_at
		FROM intake_issues %s LIMIT $%d OFFSET $%d`, query, len(args)+1, len(args)+2)

	rows, err := s.db.Query(ctx, query, append(args, input.Limit, input.Offset)...)
	if err != nil {
		return nil, fmt.Errorf("intake.ListIssues: %w", err)
	}
	defer rows.Close()

	var items []Issue
	for rows.Next() {
		var is Issue
		if err := rows.Scan(
			&is.ID, &is.ChannelID, &is.WorkspaceID, &is.ProjectID, &is.TrackingID,
			&is.SubmitterName, &is.SubmitterEmail, &is.SubmitterUserID,
			&is.Title, &is.Description, &is.IssueType, &is.Priority,
			&is.CustomFields, &is.AttachmentIDs,
			&is.Status, &is.StatusReason, &is.ConvertedIssueID, &is.AssignedTo,
			&is.ReviewedBy, &is.ReviewedAt, &is.NotifyOnStatus, &is.CreatedAt, &is.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("intake.ListIssues scan: %w", err)
		}
		items = append(items, is)
	}
	return &ListIssuesResult{Items: items, Total: total}, nil
}

// GetIssue 读取单条工单（管理员视图）。
func (s *Service) GetIssue(ctx context.Context, workspaceID, issueID int64) (*Issue, error) {
	var is Issue
	err := s.db.QueryRow(ctx, `
		SELECT id, channel_id, workspace_id, project_id, tracking_id,
			submitter_name, submitter_email, submitter_user_id,
			title, description, issue_type, priority, custom_fields, attachment_ids,
			status, status_reason, converted_issue_id, assigned_to,
			reviewed_by, reviewed_at, notify_on_status, created_at, updated_at
		FROM intake_issues WHERE id = $1 AND workspace_id = $2`,
		issueID, workspaceID,
	).Scan(
		&is.ID, &is.ChannelID, &is.WorkspaceID, &is.ProjectID, &is.TrackingID,
		&is.SubmitterName, &is.SubmitterEmail, &is.SubmitterUserID,
		&is.Title, &is.Description, &is.IssueType, &is.Priority,
		&is.CustomFields, &is.AttachmentIDs,
		&is.Status, &is.StatusReason, &is.ConvertedIssueID, &is.AssignedTo,
		&is.ReviewedBy, &is.ReviewedAt, &is.NotifyOnStatus, &is.CreatedAt, &is.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.NotFound("INTAKE.ISSUE_NOT_FOUND", "工单不存在")
		}
		return nil, fmt.Errorf("intake.GetIssue: %w", err)
	}
	return &is, nil
}

// ReviewDecision 管理员审核动作。
type ReviewDecision struct {
	Action        string // "accept" | "reject" | "archive"
	TargetIssueType string // accept 时选择 "requirement" | "defect" | "task"
	TargetProjectID *int64 // accept 时选择项目
	Reason        string // reject/archive 时填写原因
	ReviewerID    int64
}

// ReviewIssue 管理员审核收件工单（accept/reject/archive）。
// accept 时会在 ProjectService 创建正式 Issue（由调用方编排）。
func (s *Service) ReviewIssue(ctx context.Context, workspaceID, issueID int64, decision ReviewDecision) (*Issue, error) {
	iss, err := s.GetIssue(ctx, workspaceID, issueID)
	if err != nil {
		return nil, err
	}
	if iss.Status != string(IntakeStatusOpen) {
		return nil, errs.Validation("INTAKE.BAD_STATUS", "当前状态不允许此操作")
	}

	var newStatus string
	switch decision.Action {
	case "accept":
		newStatus = string(IntakeStatusAccepted)
	case "reject":
		newStatus = string(IntakeStatusRejected)
	case "archive":
		newStatus = string(IntakeStatusArchived)
	default:
		return nil, errs.Validation("INTAKE.BAD_ACTION", "无效操作")
	}

	_, err = s.db.Exec(ctx, `
		UPDATE intake_issues
		SET status = $1, status_reason = $2, reviewed_by = $3, reviewed_at = NOW(), updated_at = NOW()
		WHERE id = $4 AND workspace_id = $5 AND status = 'open'`,
		newStatus, decision.Reason, decision.ReviewerID, issueID, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("intake.ReviewIssue: %w", err)
	}
	return s.GetIssue(ctx, workspaceID, issueID)
}

// LinkConvertedIssue 在 intake 工单上记录转正的 Issue ID。
func (s *Service) LinkConvertedIssue(ctx context.Context, workspaceID, issueID, convertedIssueID int64, targetProjectID *int64) error {
	_, err := s.db.Exec(ctx, `
		UPDATE intake_issues
		SET converted_issue_id = $1, project_id = $2, updated_at = NOW()
		WHERE id = $3 AND workspace_id = $4`,
		convertedIssueID, targetProjectID, issueID, workspaceID)
	return err
}

// AssignIssue 自动/手动指派处理人。
func (s *Service) AssignIssue(ctx context.Context, workspaceID, issueID, assignToUserID int64) error {
	_, err := s.db.Exec(ctx, `
		UPDATE intake_issues SET assigned_to = $1, updated_at = NOW()
		WHERE id = $2 AND workspace_id = $3`,
		assignToUserID, issueID, workspaceID)
	return err
}

// --- 公开跟踪视图 ---

// GetPublicView 提交者看到的脱敏视图。
func (s *Service) GetPublicView(ctx context.Context, trackingID, email string) (*PublicView, error) {
	var is Issue
	err := s.db.QueryRow(ctx, `
		SELECT id, channel_id, workspace_id, project_id, tracking_id,
			submitter_name, submitter_email, submitter_user_id,
			title, description, issue_type, priority, custom_fields, attachment_ids,
			status, status_reason, converted_issue_id, assigned_to,
			reviewed_by, reviewed_at, notify_on_status, created_at, updated_at
		FROM intake_issues WHERE tracking_id = $1 AND submitter_email = $2`,
		trackingID, email,
	).Scan(
		&is.ID, &is.ChannelID, &is.WorkspaceID, &is.ProjectID, &is.TrackingID,
		&is.SubmitterName, &is.SubmitterEmail, &is.SubmitterUserID,
		&is.Title, &is.Description, &is.IssueType, &is.Priority,
		&is.CustomFields, &is.AttachmentIDs,
		&is.Status, &is.StatusReason, &is.ConvertedIssueID, &is.AssignedTo,
		&is.ReviewedBy, &is.ReviewedAt, &is.NotifyOnStatus, &is.CreatedAt, &is.UpdatedAt,
	)
	if err != nil {
		return nil, errs.NotFound("INTAKE.NOT_FOUND", "未找到匹配的工单（请确认 tracking_id 与邮箱）")
	}

	return &PublicView{
		TrackingID:       is.TrackingID,
		Title:            is.Title,
		Status:           is.Status,
		StatusText:       StatusText(is.Status),
		Priority:         is.Priority,
		IssueType:        IssueTypeText(is.IssueType),
		SubmittedAt:      is.CreatedAt,
		ReviewedAt:       is.ReviewedAt,
		StatusReason:     is.StatusReason,
		ConvertedIssueID: is.ConvertedIssueID,
	}, nil
}

// --- 限流检查 ---

// CheckRateLimit 使用 Redis 风格的 INCR+TTL 实现滑动窗口限流。
// 简化为 DB 计数（MVP），后续可替换为 Redis INCR EXPIRE。
func (s *Service) CheckRateLimit(ctx context.Context, workspaceID, channelID int64, clientIP string, limitPerMin int16) (bool, error) {
	if limitPerMin <= 0 {
		return true, nil
	}
	var count int64
	since := time.Now().Add(-1 * time.Minute)
	err := s.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM intake_issues
		WHERE channel_id = $1 AND created_at > $2`,
		channelID, since).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("intake.CheckRateLimit: %w", err)
	}
	// 注意：这里用 channel 维度全局限流代替 IP 维度，生产环境建议使用 Redis
	return count < int64(limitPerMin), nil
}

// --- 辅助 ---

func isValidSlug(s string) bool {
	if len(s) == 0 || len(s) > 64 {
		return false
	}
	for _, c := range s {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' || c == '_') {
			return false
		}
	}
	return true
}

func generateTrackingID() string {
	const prefix = "YD-IN-"
	now := time.Now()
	datePart := now.Format("060102")
	randomPart := fmt.Sprintf("%04d", now.Nanosecond()%10000)
	return prefix + datePart + "-" + randomPart
}

// IsActive 报告通道是否生效；便于 handler 校验。
func (s *Service) IsChannelActive(ctx context.Context, channelID int64) (bool, error) {
	var isActive bool
	err := s.db.QueryRow(ctx, `SELECT is_active FROM intake_channels WHERE id = $1`, channelID).Scan(&isActive)
	if err != nil {
		return false, err
	}
	return isActive, nil
}

// DB 返回底层连接池（供 handler 内部查询使用）。
// 注意：仅在必要时使用，通常应通过 Service 方法访问。
func (s *Service) DB() *pgxpool.Pool {
	return s.db
}

// ResolveWorkspaceID 从 slug 或数字 ID 解析工作空间。
func (s *Service) ResolveWorkspaceID(ctx context.Context, key string) (int64, error) {
	if key == "" {
		return 0, fmt.Errorf("missing workspace")
	}
	if id, err := strconv.ParseInt(key, 10, 64); err == nil {
		return id, nil
	}
	var id int64
	if err := s.db.QueryRow(ctx, `SELECT id FROM workspaces WHERE slug = $1 AND is_active = true`, key).Scan(&id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, nil
		}
		return 0, err
	}
	return id, nil
}
