// Package webhook — 投递编排器：匹配订阅、构造请求、执行投递、记录日志。
//
// 投递格式对标 GitHub Webhooks：
//
//	POST <target_url>
//	Content-Type: application/json
//	X-Ydsz-Event: issue.status_changed
//	X-Ydsz-Delivery: <delivery_id>        — 唯一投递 ID（接收方幂等）
//	X-Ydsz-Signature-256: sha256=<hmac>  — HMAC-SHA256(timestamp + "." + body)
//	X-Ydsz-Timestamp: 1786033228
//
// 重试策略：5xx / 429 / 超时 → TaskExchange 退避重试（1min / 5min / 30min），
// 共 3 次；最终失败标记 unhealthy。
//
// SSRF 防护：内网 / 保留 / link-local IP 一律拒绝。
package webhook

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"

	"github.com/njydsz/ydsz-plane/internal/infrastructure/mq"
)

// 投递配置常量
const (
	// DeliveryTimeout 是单次投递 HTTP 请求的最大耗时。
	DeliveryTimeout = 10 * time.Second
	// MaxRetries 是初始投递后的最大重试次数。
	MaxRetries = 3
	// UserAgent 是投递请求的 User-Agent 头。
	UserAgent = "Ydsz-Plane-Webhook/1.0"
)

// RetryBackoffs 是每次重试的等待时间（指数退避）。
var RetryBackoffs = []time.Duration{1 * time.Minute, 5 * time.Minute, 30 * time.Minute}

// Dispatcher 负责将领域事件分发给匹配的 Webhook 订阅并执行投递。
type Dispatcher struct {
	svc    *Service
	mq     *mq.TaskClient
	log    *zap.Logger
	client *http.Client
}

// NewDispatcher 构造投递器。
func NewDispatcher(svc *Service, mqClient *mq.TaskClient, log *zap.Logger) *Dispatcher {
	return &Dispatcher{
		svc: svc,
		mq:  mqClient,
		log: log,
		client: &http.Client{
			Timeout: DeliveryTimeout,
			// 禁用 redirect 以防止 SSRF bypass via 301/302
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					MinVersion: tls.VersionTLS12,
				},
			},
		},
	}
}

// DispatchEvent 处理单个领域事件：查找匹配订阅、投递/排队重试。
// 返回成功投递的订阅数与错误数（用于指标统计）。
func (d *Dispatcher) DispatchEvent(ctx context.Context, envelope mq.EventEnvelope) (ok int, failed int) {
	if d.mq == nil {
		return 0, 0
	}

	// 查找匹配的活跃订阅
	webhooks, err := d.svc.ListActiveForEvent(ctx, envelope.WorkspaceID, envelope.EventType)
	if err != nil {
		d.log.Error("webhook: list active failed",
			zap.String("event", envelope.EventType),
			zap.Error(err))
		return 0, 0
	}

	if len(webhooks) == 0 {
		return 0, 0
	}

	// SSRF 防护 + 投递到每一个匹配的订阅
	for _, w := range webhooks {
		if err := d.deliverOne(ctx, w, envelope, 1); err != nil {
			d.log.Warn("webhook: deliver failed, enqueue retry",
				zap.Int64("webhook_id", w.ID),
				zap.Error(err))
			if retryErr := d.enqueueRetry(ctx, w, envelope, 2); retryErr != nil {
				d.log.Error("webhook: enqueue retry failed",
					zap.Int64("webhook_id", w.ID),
					zap.Error(retryErr))
				failed++
			}
			continue
		}
		ok++
	}

	return ok, failed
}

// deliverOne 同步投递单条事件到单个 Webhook。
func (d *Dispatcher) deliverOne(ctx context.Context, w *Webhook, envelope mq.EventEnvelope, attempt int) error {
	// SSRF 防护：解析目标 URL 并校验 IP
	if err := ValidateTargetURL(w.TargetURL); err != nil {
		return err
	}

	// 构造投递负载
	payload := DeliveryPayload{
		Event:      envelope.EventType,
		Workspace:  d.svc.WorkspaceSlug(ctx, envelope.WorkspaceID),
		Data:       envelope.Payload,
		ActorName:  "",
		OccurredAt: envelope.OccurredAt,
	}
	if envelope.AggregateType == "project" && envelope.AggregateID > 0 {
		payload.Project = d.svc.ProjectSlug(ctx, envelope.AggregateID)
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	// 生成唯一投递 ID
	deliveryID := generateDeliveryID()

	// 计算签名
	timestamp := time.Now().Unix()
	sig := SignatureHeader(w.Secret, timestamp, body)

	// 构造请求
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, w.TargetURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", UserAgent)
	req.Header.Set("X-Ydsz-Event", envelope.EventType)
	req.Header.Set("X-Ydsz-Delivery", deliveryID)
	req.Header.Set("X-Ydsz-Signature-256", sig)
	req.Header.Set("X-Ydsz-Timestamp", fmt.Sprintf("%d", timestamp))

	// 执行投递
	start := time.Now()
	resp, err := d.client.Do(req)
	duration := time.Since(start)
	durMs := int(duration.Milliseconds())

	// 写日志 & 更新 webhook 状态（无论成功失败都记）
	logEntry := &WebhookLog{
		WebhookID:     w.ID,
		WorkspaceID:   w.WorkspaceID,
		DeliveryID:    deliveryID,
		EventType:     envelope.EventType,
		EventID:       &envelope.EventID,
		RequestURL:    w.TargetURL,
		RequestMethod: http.MethodPost,
		RequestBody:   string(body),
		Attempt:       int16(attempt),
		DurationMs:    &durMs,
		OccurredAt:    start,
	}

	if err != nil {
		logEntry.Status = LogStatusFailed
		logEntry.Error = err.Error()
		_ = d.svc.SaveLog(ctx, logEntry)
		_ = d.svc.RecordResult(ctx, w.ID, WebhookStatusFailed, err.Error())
		return fmt.Errorf("http do: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应体（限 1MB 防止内存耗尽）
	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		respBody = []byte(fmt.Sprintf("read body error: %v", err))
	}

	respStatus := resp.StatusCode
	logEntry.ResponseStatus = &respStatus
	logEntry.ResponseBody = string(respBody)

	if respStatus >= 200 && respStatus < 300 {
		logEntry.Status = LogStatusDelivered
		_ = d.svc.SaveLog(ctx, logEntry)
		_ = d.svc.RecordResult(ctx, w.ID, WebhookStatusSuccess, "")
		return nil
	}

	// 非 2xx 视为失败（4xx 客户端错误通常无法重试恢复）
	errText := fmt.Sprintf("unexpected status %d: %s", respStatus, truncate(string(respBody), 500))
	logEntry.Status = LogStatusFailed
	logEntry.Error = errText
	_ = d.svc.SaveLog(ctx, logEntry)
	_ = d.svc.RecordResult(ctx, w.ID, WebhookStatusFailed, errText)
	return fmt.Errorf("delivery rejected: %s", errText)
}

// ValidateTargetURL 执行 SSRF 防护：解析目标 URL 并校验最终 IP。
// 阻塞内网、保留、link-local 地址。
// 导出为包级函数，供自动化引擎的 webhook_call 动作等复用同一安全基线。
func ValidateTargetURL(targetURL string) error {
	u, err := url.Parse(targetURL)
	if err != nil {
		return fmt.Errorf("invalid target URL: %w", err)
	}

	// 仅允许 http / https 协议
	if !strings.EqualFold(u.Scheme, "http") && !strings.EqualFold(u.Scheme, "https") {
		return fmt.Errorf("scheme %q not allowed (only http/https)", u.Scheme)
	}

	// 域名解析 → IP
	host := u.Hostname()
	if host == "" {
		return fmt.Errorf("missing host in target URL")
	}

	ips, err := net.LookupIP(host)
	if err != nil {
		return fmt.Errorf("resolve host: %w", err)
	}

	for _, ip := range ips {
		if isBlockedIP(ip) {
			return fmt.Errorf("target IP %s is blocked (private/link-local)", ip.String())
		}
	}

	return nil
}

// isBlockedIP 报告 IP 是否为内网 / 保留地址。
func isBlockedIP(ip net.IP) bool {
	if ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsPrivate() {
		return true
	}
	// 额外检查 0.0.0.0/8, IPv6 未指定地址
	if ip.Equal(net.IPv4zero) || ip.Equal(net.IPv6unspecified) {
		return true
	}
	return false
}

// enqueueRetry 将失败投递排入 TaskExchange 的延迟重试队列。
func (d *Dispatcher) enqueueRetry(ctx context.Context, w *Webhook, envelope mq.EventEnvelope, attempt int) error {
	if attempt > MaxRetries {
		// 超过最大重试 → unhealthy
		_ = d.svc.RecordResult(ctx, w.ID, WebhookStatusUnhealthy, "max retries exhausted")
		return nil
	}

	payload, err := json.Marshal(RetryTask{
		WebhookID: w.ID,
		Envelope:  envelope,
		Attempt:   attempt,
	})
	if err != nil {
		return fmt.Errorf("marshal retry task: %w", err)
	}

	delay := RetryBackoffs[min(attempt-1, len(RetryBackoffs)-1)]
	return d.mq.Enqueue(ctx, mq.Task{
		ID:         fmt.Sprintf("webhook-retry-%d-%d", w.ID, time.Now().UnixNano()),
		Type:       "webhook.retry",
		Payload:    payload,
		MaxRetries: 0, // retry 任务不再重试
		Delay:      delay,
		Priority:   5,
		CreatedAt:  time.Now(),
	})
}

// RetryTask 是排入重试队列的任务结构。
type RetryTask struct {
	WebhookID int64            `json:"webhook_id"`
	Envelope  mq.EventEnvelope `json:"envelope"`
	Attempt   int              `json:"attempt"`
}

// HandleRetry 消费 retry 任务执行再次投递。
func (d *Dispatcher) HandleRetry(ctx context.Context, task mq.Task) error {
	var rt RetryTask
	if err := json.Unmarshal(task.Payload, &rt); err != nil {
		return fmt.Errorf("unmarshal retry task: %w", err)
	}

	// 重新查询 Webhook 配置（需要签名密钥）
	w, err := d.svc.GetByIDWithSecret(ctx, rt.Envelope.WorkspaceID, rt.WebhookID)
	if err != nil {
		return fmt.Errorf("webhook not found for retry: %w", err)
	}
	if !w.IsActive {
		d.log.Info("webhook: skip retry for inactive webhook",
			zap.Int64("webhook_id", w.ID))
		return nil
	}

	if err := d.deliverOne(ctx, w, rt.Envelope, rt.Attempt); err != nil {
		d.log.Warn("webhook: retry failed",
			zap.Int64("webhook_id", w.ID),
			zap.Int("attempt", rt.Attempt),
			zap.Error(err))
		// 继续重试下一级
		return d.enqueueRetry(ctx, w, rt.Envelope, rt.Attempt+1)
	}

	d.log.Info("webhook: retry succeeded",
		zap.Int64("webhook_id", w.ID),
		zap.Int("attempt", rt.Attempt))
	return nil
}

// ExecuteTestPing 执行测试投递（由 handler 同步调用简化流程）。
func (d *Dispatcher) ExecuteTestPing(ctx context.Context, w *Webhook) error {
	pingPayload := map[string]interface{}{
		"ping":        true,
		"hook_id":     w.ID,
		"test":        true,
		"event":       "ping",
		"occurred_at": time.Now().UTC(),
	}
	pingBody, _ := json.Marshal(pingPayload)

	envelope := mq.EventEnvelope{
		EventID:     -w.ID, // 合成 ID（负数以区分真实事件）
		EventType:   "ping",
		WorkspaceID: w.WorkspaceID,
		Payload:     pingBody,
		OccurredAt:  time.Now(),
	}
	return d.deliverOne(ctx, w, envelope, 1)
}

// RetryLog 手动重投一条历史投递日志（管理页"重投"按钮）。
//
// 实现要点：
//   - 从 webhook_logs 定位日志 → 校验归属（workspace + webhook）
//   - 通过日志记录的 event_id 回查 domain_events 重建原始事件（事件 30 天
//     保留期内可重投；超期/已清理返回明确错误）
//   - 同步重新投递（与初始投递共用 deliverOne：签名/SSRF/日志全套）
//
// 与自动重试（enqueueRetry 走 MQ 退避）不同，手动重投是用户主动操作，
// 采用同步语义，管理页可立即看到结果。API 进程无需 MQ 连接即可完成。
func (d *Dispatcher) RetryLog(ctx context.Context, workspaceID, webhookID, logID int64) error {
	logEntry, err := d.svc.GetLogByID(ctx, workspaceID, webhookID, logID)
	if err != nil {
		return err
	}
	if logEntry.EventID == nil || *logEntry.EventID <= 0 {
		return fmt.Errorf("webhook.retry: 该日志无原始事件可重投（测试 ping 或事件已清理）")
	}

	// 重建原始领域事件（domain_events 保留期内）
	var (
		eventType  string
		payload    json.RawMessage
		occurredAt time.Time
		aggType    string
		aggID      int64
	)
	err = d.svc.db.QueryRow(ctx, `
		SELECT event_type, payload, occurred_at, aggregate_type, aggregate_id
		FROM domain_events WHERE id = $1 AND workspace_id = $2`,
		*logEntry.EventID, workspaceID).Scan(&eventType, &payload, &occurredAt, &aggType, &aggID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("webhook.retry: 原始事件已超出保留期（30 天）或被清理")
		}
		return fmt.Errorf("webhook.retry: load event: %w", err)
	}

	envelope := mq.EventEnvelope{
		EventID:       *logEntry.EventID,
		EventType:     eventType,
		WorkspaceID:   workspaceID,
		AggregateType: aggType,
		AggregateID:   aggID,
		Payload:       payload,
		OccurredAt:    occurredAt,
	}

	w, err := d.svc.GetByIDWithSecret(ctx, workspaceID, webhookID)
	if err != nil {
		return fmt.Errorf("webhook.retry: load webhook: %w", err)
	}
	if !w.IsActive {
		return fmt.Errorf("webhook.retry: webhook 已停用，请先启用")
	}

	if err := d.deliverOne(ctx, w, envelope, 1); err != nil {
		return fmt.Errorf("webhook.retry: %w", err)
	}
	return nil
}

// --- 内部辅助 ---

// generateDeliveryID 生成唯一投递 ID。
func generateDeliveryID() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}

// truncate 截断字符串到指定长度。
func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

// min 返回两个 int 中较小者。
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
