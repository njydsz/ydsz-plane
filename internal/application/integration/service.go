// Package integration — 集成应用服务。
//
// 职责：
//   - 集成实例 CRUD（list / get / create / update / delete）；
//   - Connect：校验 + 测试 + 加密存储凭证；
//   - Disconnect / Pause / Resume；
//   - ListEvents：按集成 / 方向 / 状态过滤事件日志；
//   - SyncNow：触发一次同步并记录事件；
//   - Provider registry：map[string]IntegrationConnector；
//   - 凭证加解密：AES-256-GCM，工作空间派生密钥。
package integration

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

// Service 集成应用服务。
type Service struct {
	db     *pgxpool.Pool
	cipher credentialCipher
	// connectors 是 provider → connector 的注册表。
	connectors map[string]IntegrationConnector
}

// NewService 创建集成服务，并注册全部内置连接器。
//
// cipher 可为 nil（测试场景），此时凭证以明文存储（不推荐生产使用）。
func NewService(db *pgxpool.Pool, cipher credentialCipher) *Service {
	s := &Service{db: db, cipher: cipher, connectors: make(map[string]IntegrationConnector)}
	s.RegisterConnector(NewGitHubConnector())
	s.RegisterConnector(NewGitLabConnector())
	s.RegisterConnector(NewSlackConnector())
	s.RegisterConnector(NewSentryConnector())
	s.RegisterConnector(NewJiraConnector())
	return s
}

// RegisterConnector 注册一个连接器（后注册的同名 provider 覆盖先注册的）。
func (s *Service) RegisterConnector(c IntegrationConnector) {
	if c == nil {
		return
	}
	s.connectors[c.Provider()] = c
}

// Connector 返回指定 provider 的连接器，未注册时返回 nil。
func (s *Service) Connector(provider string) IntegrationConnector {
	return s.connectors[strings.TrimSpace(provider)]
}

// --- Provider 元数据 ---

// ListProviders 返回全部已注册 provider 的元数据（供 GET /providers）。
func (s *Service) ListProviders() []ProviderMeta {
	return []ProviderMeta{
		{
			Provider:       string(ProviderGitHub),
			Name:           "GitHub",
			Icon:           "github",
			Description:    "同步 GitHub 仓库的 Issues 到工作项，支持状态 / 标签 / 指派人映射。",
			DocsURL:        "https://docs.github.com/en/rest",
			SupportsSync:   true,
			SupportsNotify: false,
			Fields: []ProviderField{
				{Key: "token", Label: "Personal Access Token", Required: true, Secret: true, Placeholder: "ghp_..."},
				{Key: "repo", Label: "Repository (owner/repo)", Required: true, Secret: false, Placeholder: "owner/repo"},
			},
		},
		{
			Provider:       string(ProviderGitLab),
			Name:           "GitLab",
			Icon:           "gitlab",
			Description:    "连接 GitLab 项目，同步 Issues（同步能力开发中）。",
			DocsURL:        "https://docs.gitlab.com/ee/api/",
			SupportsSync:   false,
			SupportsNotify: false,
			Fields: []ProviderField{
				{Key: "token", Label: "Access Token", Required: true, Secret: true, Placeholder: "glpat-..."},
				{Key: "base_url", Label: "Base URL (可选)", Required: false, Secret: false, Placeholder: "https://gitlab.example.com"},
			},
		},
		{
			Provider:       string(ProviderSlack),
			Name:           "Slack",
			Icon:           "slack",
			Description:    "向 Slack 频道发送工作项通知（创建 / 状态变更 / 评论）。",
			DocsURL:        "https://api.slack.com/",
			SupportsSync:   false,
			SupportsNotify: true,
			Fields: []ProviderField{
				{Key: "bot_token", Label: "Bot Token", Required: true, Secret: true, Placeholder: "xoxb-..."},
				{Key: "channel", Label: "Channel", Required: true, Secret: false, Placeholder: "#general"},
			},
		},
		{
			Provider:       string(ProviderSentry),
			Name:           "Sentry",
			Icon:           "sentry",
			Description:    "连接 Sentry 组织，监控错误与异常（同步能力开发中）。",
			DocsURL:        "https://docs.sentry.io/api/",
			SupportsSync:   false,
			SupportsNotify: false,
			Fields: []ProviderField{
				{Key: "token", Label: "Auth Token", Required: true, Secret: true, Placeholder: "sntrys_..."},
				{Key: "base_url", Label: "Base URL (可选)", Required: false, Secret: false, Placeholder: "https://sentry.io"},
			},
		},
		{
			Provider:       string(ProviderJira),
			Name:           "Jira",
			Icon:           "jira",
			Description:    "连接 Jira 实例，同步 Issue（同步能力开发中）。",
			DocsURL:        "https://developer.atlassian.com/cloud/jira/platform/rest/v3/",
			SupportsSync:   false,
			SupportsNotify: false,
			Fields: []ProviderField{
				{Key: "token", Label: "API Token", Required: true, Secret: true, Placeholder: ""},
				{Key: "base_url", Label: "Jira Base URL", Required: true, Secret: false, Placeholder: "https://your-domain.atlassian.net"},
			},
		},
	}
}

// --- 凭证加解密 ---

// encryptCredentials 将明文凭证 JSON 加密为密文。
func (s *Service) encryptCredentials(workspaceID int64, creds map[string]string) (string, error) {
	plain, err := json.Marshal(creds)
	if err != nil {
		return "", errs.ErrInternal.Wrap(err)
	}
	if s.cipher == nil {
		// 测试降级：明文存储（生产必须配置 YDSZ_INTEGRATION_SECRET_KEY）。
		return string(plain), nil
	}
	enc, err := s.cipher.Encrypt(workspaceID, string(plain))
	if err != nil {
		return "", errs.ErrInternal.Wrap(err)
	}
	return enc, nil
}

// decryptCredentials 解密并解析凭证 JSON。
func (s *Service) decryptCredentials(workspaceID int64, ciphertext string) (map[string]string, error) {
	if ciphertext == "" {
		return map[string]string{}, nil
	}
	var plain string
	var err error
	if s.cipher == nil {
		plain = ciphertext
	} else {
		plain, err = s.cipher.Decrypt(workspaceID, ciphertext)
		if err != nil {
			return nil, errs.ErrInternal.Wrap(err)
		}
	}
	creds := map[string]string{}
	if plain == "" {
		return creds, nil
	}
	if err := json.Unmarshal([]byte(plain), &creds); err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	return creds, nil
}

// marshalConfig 将配置 map 序列化为 JSON 字符串（空 map 返回空串）。
func marshalConfig(config map[string]string) string {
	if len(config) == 0 {
		return ""
	}
	buf, err := json.Marshal(config)
	if err != nil {
		return ""
	}
	return string(buf)
}

// parseConfig 解析配置 JSON（空串返回空 map）。
func parseConfig(raw string) map[string]string {
	if raw == "" {
		return map[string]string{}
	}
	config := map[string]string{}
	_ = json.Unmarshal([]byte(raw), &config)
	return config
}

// --- CRUD ---

// Create 创建（连接）集成：校验 provider → 校验凭证 → 测试连接 → 加密存储。
//
// 测试失败时不落库，返回 422 并携带错误详情。
func (s *Service) Create(ctx context.Context, in CreateIntegrationInput) (*Integration, error) {
	provider := strings.TrimSpace(in.Provider)
	if !Provider(provider).IsValid() {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "provider", Reason: "不支持的 provider"})
	}
	connector := s.Connector(provider)
	if connector == nil {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "provider", Reason: "provider 未注册"})
	}
	if len(in.Credentials) == 0 {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "credentials", Reason: "凭证不能为空"})
	}
	if err := connector.ValidateCredentials(in.Credentials); err != nil {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "credentials", Reason: err.Error()})
	}
	// 测试连接：失败则拒绝创建。
	if err := connector.TestConnection(in.Credentials); err != nil {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "credentials", Reason: "连接测试失败: " + err.Error()})
	}

	encCreds, err := s.encryptCredentials(in.WorkspaceID, in.Credentials)
	if err != nil {
		return nil, err
	}
	name := strings.TrimSpace(in.Name)
	if name == "" {
		name = connector.Provider()
	}

	var itg Integration
	err = s.db.QueryRow(ctx, `
		INSERT INTO integrations (id, workspace_id, provider, name, status, credentials, config, created_by, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8, NOW(), NOW())
		RETURNING id, workspace_id, provider, name, status, credentials, COALESCE(config,''),
			last_sync_at, COALESCE(last_error,''), created_by, created_at, updated_at`,
		genID(), in.WorkspaceID, provider, name, string(StatusActive),
		encCreds, marshalConfig(in.Config), in.CreatedBy,
	).Scan(&itg.ID, &itg.WorkspaceID, &itg.Provider, &itg.Name, &itg.Status,
		&itg.Credentials, &itg.Config, &itg.LastSyncAt, &itg.LastError,
		&itg.CreatedBy, &itg.CreatedAt, &itg.UpdatedAt)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	return &itg, nil
}

// GetByID 查询单个集成（含密文凭证，供内部同步使用）。
func (s *Service) GetByID(ctx context.Context, workspaceID, integrationID int64) (*Integration, error) {
	var itg Integration
	err := s.db.QueryRow(ctx, `
		SELECT id, workspace_id, provider, name, status, credentials, COALESCE(config,''),
			last_sync_at, COALESCE(last_error,''), created_by, created_at, updated_at
		FROM integrations WHERE id = $1 AND workspace_id = $2`,
		integrationID, workspaceID,
	).Scan(&itg.ID, &itg.WorkspaceID, &itg.Provider, &itg.Name, &itg.Status,
		&itg.Credentials, &itg.Config, &itg.LastSyncAt, &itg.LastError,
		&itg.CreatedBy, &itg.CreatedAt, &itg.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errs.ErrNotFound.WithCodeMessage("INTEGRATION.NOT_FOUND", "集成不存在")
	}
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	return &itg, nil
}

// GetByIDPublic 查询单个集成（不含凭证），供 API 响应使用。
func (s *Service) GetByIDPublic(ctx context.Context, workspaceID, integrationID int64) (*Integration, error) {
	itg, err := s.GetByID(ctx, workspaceID, integrationID)
	if err != nil {
		return nil, err
	}
	itg.Credentials = ""
	return itg, nil
}

// List 列出工作空间内的集成。
func (s *Service) List(ctx context.Context, opts ListIntegrationsOptions) ([]Integration, int, error) {
	where := []string{"workspace_id = $1"}
	args := []any{opts.WorkspaceID}
	idx := 2
	if opts.Provider != "" {
		where = append(where, "provider = $"+strconv.Itoa(idx))
		args = append(args, opts.Provider)
		idx++
	}
	if opts.Status != "" {
		where = append(where, "status = $"+strconv.Itoa(idx))
		args = append(args, opts.Status)
		idx++
	}
	cond := strings.Join(where, " AND ")

	var total int
	if err := s.db.QueryRow(ctx, `SELECT COUNT(*) FROM integrations WHERE `+cond, args...).Scan(&total); err != nil {
		return nil, 0, errs.ErrInternal.Wrap(err)
	}

	limit, offset := opts.Limit, opts.Offset
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	args = append(args, limit, offset)

	rows, err := s.db.Query(ctx, `
		SELECT id, workspace_id, provider, name, status, '' AS credentials, COALESCE(config,''),
			last_sync_at, COALESCE(last_error,''), created_by, created_at, updated_at
		FROM integrations WHERE `+cond+`
		ORDER BY created_at DESC
		LIMIT $`+strconv.Itoa(idx)+` OFFSET $`+strconv.Itoa(idx+1), args...)
	if err != nil {
		return nil, 0, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()

	out := make([]Integration, 0)
	for rows.Next() {
		var itg Integration
		if err := rows.Scan(&itg.ID, &itg.WorkspaceID, &itg.Provider, &itg.Name, &itg.Status,
			&itg.Credentials, &itg.Config, &itg.LastSyncAt, &itg.LastError,
			&itg.CreatedBy, &itg.CreatedAt, &itg.UpdatedAt); err != nil {
			return nil, 0, errs.ErrInternal.Wrap(err)
		}
		out = append(out, itg)
	}
	return out, total, rows.Err()
}

// Update 更新集成（name / credentials / config）。凭证变更时重新加密。
func (s *Service) Update(ctx context.Context, workspaceID, integrationID int64, in UpdateIntegrationInput) (*Integration, error) {
	itg, err := s.GetByID(ctx, workspaceID, integrationID)
	if err != nil {
		return nil, err
	}

	sets := []string{}
	args := []any{}
	idx := 1
	add := func(expr string, val any) {
		args = append(args, val)
		sets = append(sets, expr+"$"+strconv.Itoa(idx))
		idx++
	}

	if in.Name != nil {
		name := strings.TrimSpace(*in.Name)
		if name == "" {
			return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "name", Reason: "名称不能为空"})
		}
		add("name = ", name)
	}
	if in.Config != nil {
		add("config = ", marshalConfig(in.Config))
	}
	if in.Credentials != nil {
		connector := s.Connector(itg.Provider)
		if connector == nil {
			return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "credentials", Reason: "provider 未注册"})
		}
		if err := connector.ValidateCredentials(in.Credentials); err != nil {
			return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "credentials", Reason: err.Error()})
		}
		enc, err := s.encryptCredentials(workspaceID, in.Credentials)
		if err != nil {
			return nil, err
		}
		add("credentials = ", enc)
	}
	if len(sets) == 0 {
		return s.GetByIDPublic(ctx, workspaceID, integrationID)
	}
	sets = append(sets, "updated_at = NOW()")
	args = append(args, integrationID, workspaceID)
	query := "UPDATE integrations SET " + strings.Join(sets, ", ") +
		" WHERE id = $" + strconv.Itoa(idx) + " AND workspace_id = $" + strconv.Itoa(idx+1)
	if _, err := s.db.Exec(ctx, query, args...); err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	return s.GetByIDPublic(ctx, workspaceID, integrationID)
}

// Delete 删除集成（含事件日志）。
func (s *Service) Delete(ctx context.Context, workspaceID, integrationID int64) error {
	tag, err := s.db.Exec(ctx, `DELETE FROM integrations WHERE id = $1 AND workspace_id = $2`, integrationID, workspaceID)
	if err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	if tag.RowsAffected() == 0 {
		return errs.ErrNotFound.WithCodeMessage("INTEGRATION.NOT_FOUND", "集成不存在")
	}
	_, _ = s.db.Exec(ctx, `DELETE FROM integration_events WHERE integration_id = $1`, integrationID)
	return nil
}

// Disconnect 断开集成（删除凭证，状态置为 pending）。
func (s *Service) Disconnect(ctx context.Context, workspaceID, integrationID int64) (*Integration, error) {
	tag, err := s.db.Exec(ctx, `
		UPDATE integrations SET credentials = '', status = $1, last_error = '', updated_at = NOW()
		WHERE id = $2 AND workspace_id = $3`,
		string(StatusPending), integrationID, workspaceID)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	if tag.RowsAffected() == 0 {
		return nil, errs.ErrNotFound.WithCodeMessage("INTEGRATION.NOT_FOUND", "集成不存在")
	}
	return s.GetByIDPublic(ctx, workspaceID, integrationID)
}

// Pause 暂停集成。
func (s *Service) Pause(ctx context.Context, workspaceID, integrationID int64) (*Integration, error) {
	return s.setStatus(ctx, workspaceID, integrationID, StatusPaused)
}

// Resume 恢复集成。
func (s *Service) Resume(ctx context.Context, workspaceID, integrationID int64) (*Integration, error) {
	return s.setStatus(ctx, workspaceID, integrationID, StatusActive)
}

// setStatus 更新集成状态（内部辅助）。
func (s *Service) setStatus(ctx context.Context, workspaceID, integrationID int64, status Status) (*Integration, error) {
	tag, err := s.db.Exec(ctx, `
		UPDATE integrations SET status = $1, updated_at = NOW()
		WHERE id = $2 AND workspace_id = $3`,
		string(status), integrationID, workspaceID)
	if err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	if tag.RowsAffected() == 0 {
		return nil, errs.ErrNotFound.WithCodeMessage("INTEGRATION.NOT_FOUND", "集成不存在")
	}
	return s.GetByIDPublic(ctx, workspaceID, integrationID)
}

// --- 测试与同步 ---

// TestConnection 使用已存储的凭证测试连接。
func (s *Service) TestConnection(ctx context.Context, workspaceID, integrationID int64) error {
	itg, err := s.GetByID(ctx, workspaceID, integrationID)
	if err != nil {
		return err
	}
	connector := s.Connector(itg.Provider)
	if connector == nil {
		return errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "provider", Reason: "provider 未注册"})
	}
	creds, err := s.decryptCredentials(workspaceID, itg.Credentials)
	if err != nil {
		return err
	}
	if err := connector.TestConnection(creds); err != nil {
		// 记录失败事件并更新状态。
		_, _ = s.db.Exec(ctx, `
			UPDATE integrations SET status = $1, last_error = $2, updated_at = NOW()
			WHERE id = $3`, string(StatusError), err.Error(), integrationID)
		_ = s.RecordEvent(ctx, integrationID, string(DirectionOutbound), "test_connection", string(EventFailed), "", err.Error())
		return errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "credentials", Reason: "连接测试失败: " + err.Error()})
	}
	// 测试成功：清除错误并置为 active。
	_, _ = s.db.Exec(ctx, `
		UPDATE integrations SET status = $1, last_error = '', updated_at = NOW()
		WHERE id = $2`, string(StatusActive), integrationID)
	_ = s.RecordEvent(ctx, integrationID, string(DirectionOutbound), "test_connection", string(EventSuccess), "", "")
	return nil
}

// SyncNow 触发一次同步：拉取第三方工作项并记录事件。
//
// 当前实现拉取并归一化工作项，写入事件日志；与内部工作项的持久化映射在后续迭代接入。
func (s *Service) SyncNow(ctx context.Context, workspaceID, integrationID int64) (*SyncResult, error) {
	itg, err := s.GetByID(ctx, workspaceID, integrationID)
	if err != nil {
		return nil, err
	}
	if itg.Status == string(StatusPaused) {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "status", Reason: "集成已暂停，无法同步"})
	}
	connector := s.Connector(itg.Provider)
	if connector == nil {
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "provider", Reason: "provider 未注册"})
	}
	creds, err := s.decryptCredentials(workspaceID, itg.Credentials)
	if err != nil {
		return nil, err
	}
	config := parseConfig(itg.Config)

	result, syncErr := connector.SyncWorkItems(ctx, creds, config)
	if syncErr != nil {
		_, _ = s.db.Exec(ctx, `
			UPDATE integrations SET status = $1, last_error = $2, updated_at = NOW()
			WHERE id = $3`, string(StatusError), syncErr.Error(), integrationID)
		_ = s.RecordEvent(ctx, integrationID, string(DirectionInbound), "sync", string(EventFailed), "", syncErr.Error())
		return nil, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "sync", Reason: "同步失败: " + syncErr.Error()})
	}

	// 同步成功：更新 last_sync_at，记录事件。
	_, _ = s.db.Exec(ctx, `
		UPDATE integrations SET last_sync_at = NOW(), status = $1, last_error = '', updated_at = NOW()
		WHERE id = $2`, string(StatusActive), integrationID)
	_ = s.RecordEvent(ctx, integrationID, string(DirectionInbound), "sync", string(EventSuccess),
		fmt.Sprintf(`{"items_fetched":%d,"items_imported":%d}`, result.ItemsFetched, result.ItemsImported), "")
	return &result, nil
}

// --- 事件日志 ---

// RecordEvent 写入一条集成事件日志。
func (s *Service) RecordEvent(ctx context.Context, integrationID int64, direction, eventType, status, payload, eventErr string) error {
	_, err := s.db.Exec(ctx, `
		INSERT INTO integration_events (id, integration_id, direction, event_type, status, payload, error, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7, NOW())`,
		genID(), integrationID, direction, eventType, status, nullIfEmpty(payload), nullIfEmpty(eventErr))
	if err != nil {
		return errs.ErrInternal.Wrap(err)
	}
	return nil
}

// ListEvents 查询集成事件日志。
func (s *Service) ListEvents(ctx context.Context, opts ListEventsOptions) ([]IntegrationEvent, int, error) {
	where := []string{"e.workspace_id = $1"}
	args := []any{opts.WorkspaceID}
	idx := 2
	if opts.IntegrationID != nil {
		where = append(where, "e.integration_id = $"+strconv.Itoa(idx))
		args = append(args, *opts.IntegrationID)
		idx++
	}
	if opts.Direction != "" {
		where = append(where, "e.direction = $"+strconv.Itoa(idx))
		args = append(args, opts.Direction)
		idx++
	}
	if opts.Status != "" {
		where = append(where, "e.status = $"+strconv.Itoa(idx))
		args = append(args, opts.Status)
		idx++
	}
	if opts.EventType != "" {
		where = append(where, "e.event_type = $"+strconv.Itoa(idx))
		args = append(args, opts.EventType)
		idx++
	}
	cond := strings.Join(where, " AND ")

	var total int
	if err := s.db.QueryRow(ctx, `SELECT COUNT(*) FROM integration_events e WHERE `+cond, args...).Scan(&total); err != nil {
		return nil, 0, errs.ErrInternal.Wrap(err)
	}

	limit, offset := opts.Limit, opts.Offset
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	args = append(args, limit, offset)

	rows, err := s.db.Query(ctx, `
		SELECT e.id, e.integration_id, e.direction, e.event_type, e.status,
			COALESCE(e.payload,''), COALESCE(e.error,''), e.created_at
		FROM integration_events e WHERE `+cond+`
		ORDER BY e.created_at DESC
		LIMIT $`+strconv.Itoa(idx)+` OFFSET $`+strconv.Itoa(idx+1), args...)
	if err != nil {
		return nil, 0, errs.ErrInternal.Wrap(err)
	}
	defer rows.Close()

	out := make([]IntegrationEvent, 0)
	for rows.Next() {
		var ev IntegrationEvent
		if err := rows.Scan(&ev.ID, &ev.IntegrationID, &ev.Direction, &ev.EventType, &ev.Status,
			&ev.Payload, &ev.Error, &ev.CreatedAt); err != nil {
			return nil, 0, errs.ErrInternal.Wrap(err)
		}
		out = append(out, ev)
	}
	return out, total, rows.Err()
}

// --- helpers ---

// nullIfEmpty 将空串转为 NULL（用于 PostgreSQL 可选列）。
func nullIfEmpty(s string) any {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return s
}

// 确保 time 包被引用（SyncNow 当前未直接使用 time，但保留导入以便后续扩展）。
var _ = time.Now
