// Package notification — 多渠道通知分发 Worker。
//
// StartDispatchWorker 在 goroutine 中循环：
//   - 每 30s 捞 notification_deliveries WHERE status='pending' AND retry_count<3
//   - 按 channel 投递：email → mailSvc；wecom/dingtalk/feishu → webhook POST（加签名）
//   - 成功标 'sent'；失败 retry_count++、next_retry_at = now + 5^retry 分钟
//
// 投递记录由 consumer / HTTP handler 在生成通知时同步写入 notification_deliveries，
// dispatcher 仅负责异步外发。
package notification

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/njydsz/ydsz-plane/internal/infrastructure/mail"
)

// 常量：最大重试次数 / tick 间隔
const (
	dispatchMaxRetries = 3
	dispatchTick       = 30 * time.Second
)

// pendingDelivery 是待投递记录 + 关联的通知标题/动作 URL。
type pendingDelivery struct {
	ID             int64
	NotificationID int64
	Channel        Channel
	Recipient      string
	RetryCount     int
	Title          string
	ActionURL      string
}

// dispatchConfig 分发 worker 的外部依赖配置。
type dispatchConfig struct {
	// mailSvc 邮件发送器（由 main.go 装配）。
	mailSvc mail.EmailService
	// baseURL 前端基础 URL，用于拼接邮件/IM 动作链接。
	baseURL string
}

// StartDispatchWorker 阻塞循环，ctx 取消时退出。
// 应在 cmd/worker/main.go 中通过 go 协程调用。
//
// 参数：
//   - ctx    生命周期 context
//   - db     pgx 连接池
//   - mailSvc 邮件服务（可选；为 nil 时 email 投递直接返回失败）
//   - baseURL 前端基础 URL
//   - log    zap logger
func StartDispatchWorker(ctx context.Context, db *pgxpool.Pool, mailSvc mail.EmailService, baseURL string, log *zap.Logger) {
	cfg := &dispatchConfig{mailSvc: mailSvc, baseURL: baseURL}
	ticker := time.NewTicker(dispatchTick)
	defer ticker.Stop()

	log.Info("notification dispatcher: started",
		zap.Duration("tick", dispatchTick),
		zap.Int("max_retries", dispatchMaxRetries))

	for {
		select {
		case <-ctx.Done():
			log.Info("notification dispatcher: stopped")
			return
		case <-ticker.C:
			if err := cfg.processTick(ctx, db, log); err != nil {
				log.Warn("notification dispatcher: tick failed", zap.Error(err))
			}
		}
	}
}

// processTick 执行一轮投递循环。
func (c *dispatchConfig) processTick(ctx context.Context, db *pgxpool.Pool, log *zap.Logger) error {
	deliveries, err := c.fetchPending(ctx, db)
	if err != nil {
		return fmt.Errorf("fetch pending: %w", err)
	}
	if len(deliveries) == 0 {
		return nil
	}

	log.Info("notification dispatcher: tick",
		zap.Int("pending", len(deliveries)))

	for _, d := range deliveries {
		if err := c.processOne(ctx, db, &d, log); err != nil {
			log.Warn("notification dispatcher: delivery failed",
				zap.Int64("delivery_id", d.ID),
				zap.String("channel", string(d.Channel)),
				zap.Error(err))
		}
	}
	return nil
}

// fetchPending 获取待投递记录（status='pending' AND retry_count<3 AND next_retry_at<=now）。
func (c *dispatchConfig) fetchPending(ctx context.Context, db *pgxpool.Pool) ([]pendingDelivery, error) {
	rows, err := db.Query(ctx, `
		SELECT nd.id, nd.notification_id, nd.channel, nd.recipient, nd.retry_count,
		       COALESCE(n.title, '') AS title, COALESCE(n.action_url, '') AS action_url
		FROM notification_deliveries nd
		JOIN notifications n ON n.id = nd.notification_id
		WHERE nd.status = 'pending' AND nd.retry_count < $1
		  AND (nd.next_retry_at IS NULL OR nd.next_retry_at <= NOW())
		ORDER BY nd.created_at ASC
		LIMIT 50`,
		dispatchMaxRetries)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []pendingDelivery
	for rows.Next() {
		var d pendingDelivery
		if err := rows.Scan(&d.ID, &d.NotificationID, &d.Channel, &d.Recipient, &d.RetryCount, &d.Title, &d.ActionURL); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

// processOne 处理单条投递。
func (c *dispatchConfig) processOne(ctx context.Context, db *pgxpool.Pool, d *pendingDelivery, log *zap.Logger) error {
	var err error
	switch d.Channel {
	case ChannelEmail:
		err = c.deliverEmail(d)
	case ChannelWeCom, ChannelDingTalk, ChannelFeishu:
		err = c.deliverIM(d, log)
	default:
		err = fmt.Errorf("unknown channel: %s", d.Channel)
	}

	if err != nil {
		return c.markFailed(ctx, db, d.ID, d.RetryCount, err.Error())
	}
	return c.markSent(ctx, db, d.ID)
}

// deliverEmail 通过 mailSvc 投递邮件。
func (c *dispatchConfig) deliverEmail(d *pendingDelivery) error {
	if c.mailSvc == nil {
		return fmt.Errorf("mail service not configured")
	}
	body := d.Title
	if d.ActionURL != "" {
		body += "\n\n查看详情：" + d.ActionURL
	}
	return c.mailSvc.Send(mail.Message{
		To:      d.Recipient,
		Subject: "[Ydsz Plane] " + d.Title,
		Text:    body,
		HTML:    fmt.Sprintf("<p>%s</p><p><a href=\"%s\">查看详情</a></p>", d.Title, d.ActionURL),
	})
}

// deliverIM 投递企业微信/钉钉/飞书 webhook。
func (c *dispatchConfig) deliverIM(d *pendingDelivery, log *zap.Logger) error {
	webhook := imWebhookURL(d.Channel)
	if webhook == "" {
		return fmt.Errorf("%s webhook URL not set (env: YDSZ_NOTIF_%s_WEBHOOK)", d.Channel, imEnvKey(d.Channel))
	}

	switch d.Channel {
	case ChannelDingTalk:
		return c.deliverDingTalk(webhook, d, log)
	case ChannelFeishu:
		return c.deliverFeishu(webhook, d, log)
	default:
		// 企业微信：直接 POST markdown
		body, _ := json.Marshal(map[string]any{
			"msgtype": "markdown",
			"markdown": map[string]string{
				"content": "**" + d.Title + "**\n> [查看详情](" + d.ActionURL + ")",
			},
		})
		return postJSON(webhook, body)
	}
}

// deliverDingTalk 投递钉钉（签名模式：timestamp + hmac-sha256 → hex 追加到 URL）。
func (c *dispatchConfig) deliverDingTalk(webhook string, d *pendingDelivery, log *zap.Logger) error {
	secret := os.Getenv(fmt.Sprintf("YDSZ_NOTIF_%s_SECRET", imEnvKey(ChannelDingTalk)))
	body, _ := json.Marshal(map[string]any{
		"msgtype": "text",
		"text":    map[string]string{"content": d.Title + "\n" + d.ActionURL},
	})
	if secret == "" {
		// 无密钥则直接投递（自定义机器人关键词模式）
		return postJSON(webhook, body)
	}

	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)
	sign := dingTalkSign(secret, timestamp)
	webhook = webhook + "&timestamp=" + timestamp + "&sign=" + sign

	return postJSON(webhook, body)
}

// deliverFeishu 投递飞书（timestamp + hmac-sha256 → base64 嵌入 payload）。
func (c *dispatchConfig) deliverFeishu(webhook string, d *pendingDelivery, log *zap.Logger) error {
	secret := os.Getenv(fmt.Sprintf("YDSZ_NOTIF_%s_SECRET", imEnvKey(ChannelFeishu)))

	timestamp := time.Now().Unix()
	content := d.Title
	if d.ActionURL != "" {
		content += "\n" + d.ActionURL
	}

	if secret == "" {
		body, _ := json.Marshal(map[string]any{
			"msgtype": "text",
			"content": map[string]string{"text": content},
		})
		return postJSON(webhook, body)
	}

	sign := feishuSign(secret, timestamp)
	envelope := map[string]any{
		"timestamp": strconv.FormatInt(timestamp, 10),
		"sign":      sign,
		"msg_type":  "text",
		"content":   map[string]string{"text": content},
	}
	body, _ := json.Marshal(envelope)
	return postJSON(webhook, body)
}

// markSent 更新投递状态为 sent。
func (c *dispatchConfig) markSent(ctx context.Context, db *pgxpool.Pool, deliveryID int64) error {
	_, err := db.Exec(ctx,
		`UPDATE notification_deliveries SET status = 'sent', sent_at = now() WHERE id = $1`,
		deliveryID)
	return err
}

// markFailed 更新投递状态，retry_count++，next_retry_at = now + 5^retry 分钟。
func (c *dispatchConfig) markFailed(ctx context.Context, db *pgxpool.Pool, deliveryID int64, retryCount int, reason string) error {
	backoffMin := int(math.Pow(5, float64(retryCount+1))) // 5, 25, 125
	nextRetry := time.Now().Add(time.Duration(backoffMin) * time.Minute)
	_, err := db.Exec(ctx,
		`UPDATE notification_deliveries SET status = 'failed', error_msg = $2,
		    retry_count = retry_count + 1, next_retry_at = $3
		 WHERE id = $1`,
		deliveryID, reason, nextRetry)
	return err
}

// imWebhookURL 获取 IM webhook URL（来自环境变量 YDSZ_NOTIF_<CHANNEL>_WEBHOOK）。
func imWebhookURL(ch Channel) string {
	return os.Getenv(fmt.Sprintf("YDSZ_NOTIF_%s_WEBHOOK", imEnvKey(ch)))
}

// imEnvKey 返回环境变量中的渠道大写标识。
func imEnvKey(ch Channel) string {
	switch ch {
	case ChannelWeCom:
		return "WECOM"
	case ChannelDingTalk:
		return "DINGTALK"
	case ChannelFeishu:
		return "FEISHU"
	}
	return string(ch)
}

// dingTalkSign 钉钉签名：timestamp + "\n" + secret → hmac-sha256 → hex。
func dingTalkSign(secret, timestamp string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(timestamp + "\n" + secret))
	return hex.EncodeToString(mac.Sum(nil))
}

// feishuSign 飞书签名：timestamp + "\n" + secret → hmac-sha256 → base64。
func feishuSign(secret string, timestamp int64) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(fmt.Sprintf("%d\n%s", timestamp, secret)))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

// postJSON 发送 HTTP POST JSON，超时 10s。
func postJSON(url string, body []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("webhook status %d", resp.StatusCode)
	}
	return nil
}
