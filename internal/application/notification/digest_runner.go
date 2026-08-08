// Package notification — 通知摘要聚合 Worker。
//
// DigestRunner 每分钟检查是否有用户的摘要到达触发时刻。
// 到达后调用 DigestService.BuildDigest 聚合时间窗内的通知，
// 然后通过 email 投递（或写入 notification_digests 待 dispatcher 处理）。
//
// 本质：将"分散的 in_app 通知"按 daily/weekly 频率合并为一封邮件/IM 摘要。
// 业务价值：降低用户通知噪音（参考 Linear/GitHub Digest）。
package notification

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/njydsz/ydsz-plane/internal/infrastructure/mail"
)

// 常量
const (
	digestTick      = 1 * time.Minute  // 每分钟检查一次
	digestBatchSize = 50               // 单次最多处理的待摘要数
	baseURLEnv      = "YDSZ_BASE_URL"  // 前端基础 URL 环境变量
)

// DigestDeps 摘要 Worker 依赖。
type DigestDeps struct {
	DB      *pgxpool.Pool
	MailSvc mail.EmailService
	Log     *zap.Logger
	// BaseURL 拼接邮件内动作链接。
	BaseURL string
}

// StartDigestRunner 阻塞循环，ctx 取消时优雅退出。
// 应在 cmd/worker/main.go 中通过 go 协程调用。
func StartDigestRunner(ctx context.Context, deps *DigestDeps) {
	if deps.Log == nil {
		deps.Log = zap.NewNop()
	}
	if deps.BaseURL == "" {
		deps.BaseURL = os.Getenv(baseURLEnv)
	}

	digestSvc := NewDigestService(deps.DB)

	deps.Log.Info("notification digest runner: started",
		zap.Duration("tick", digestTick))

	ticker := time.NewTicker(digestTick)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			deps.Log.Info("notification digest runner: stopped")
			return
		case <-ticker.C:
			if err := runDigestTick(ctx, digestSvc, deps); err != nil {
				deps.Log.Warn("notification digest runner: tick failed",
					zap.Error(err))
			}
		}
	}
}

// runDigestTick 执行一轮摘要检查。
func runDigestTick(ctx context.Context, digestSvc *DigestService, deps *DigestDeps) error {
	now := time.Now()

	// 1. 查找所有到达触发时刻的 (用户,工作空间,频率)
	pending, err := digestSvc.PendingDigests(ctx, now)
	if err != nil {
		return fmt.Errorf("pending digests: %w", err)
	}
	if len(pending) == 0 {
		return nil
	}

	deps.Log.Info("notification digest runner: tick",
		zap.Int("pending", len(pending)))

	// 2. 逐个处理
	processed := 0
	for i := range pending {
		if processed >= digestBatchSize {
			break
		}
		if err := processOneDigest(ctx, digestSvc, &pending[i], deps, now); err != nil {
			deps.Log.Warn("notification digest runner: process failed",
				zap.Int64("user_id", pending[i].UserID),
				zap.Int64("workspace_id", pending[i].WorkspaceID),
				zap.Error(err))
			continue
		}
		processed++
	}

	if processed > 0 {
		deps.Log.Info("notification digest runner: tick done",
			zap.Int("processed", processed))
	}
	return nil
}

// processOneDigest 处理单个摘要任务。
func processOneDigest(ctx context.Context, digestSvc *DigestService, pending *PendingDigest, deps *DigestDeps, now time.Time) error {
	// 查询上次摘要时间（确定聚合时间窗起点）
	lastAt, err := digestSvc.LastDigestAt(ctx, pending.WorkspaceID, pending.UserID, pending.DigestType)
	if err != nil {
		return fmt.Errorf("last digest at: %w", err)
	}

	// 聚合通知
	payload, err := digestSvc.BuildDigest(ctx, pending.WorkspaceID, pending.UserID, pending.DigestType, lastAt, now)
	if err != nil {
		return fmt.Errorf("build digest: %w", err)
	}

	// 无通知也记录一次"空摘要"避免重复扫描：WriteEmpty=true 标志告知 RecordDigest。
	// 写入摘要记录 + 归档原始通知
	if err := digestSvc.RecordDigest(ctx, *pending, payload); err != nil {
		return fmt.Errorf("record digest: %w", err)
	}

	// 通过 email 投递（若 mailSvc 已配置且用户有 email 收件人）
	if deps.MailSvc != nil && len(pending.Recipients) > 0 {
		subject := BuildDigestSubject(pending.DigestType, fmt.Sprintf("workspace-%d", pending.WorkspaceID))
		html := BuildDigestHTML(payload, fmt.Sprintf("workspace-%d", pending.WorkspaceID))
		for _, recipient := range pending.Recipients {
			if err := deps.MailSvc.Send(mail.Message{
				To:      recipient,
				Subject: subject,
				HTML:    html,
				Text:    digestTextSummary(payload),
			}); err != nil {
				deps.Log.Warn("digest email failed",
					zap.String("recipient", recipient),
					zap.Error(err))
			}
		}
	}

	return nil
}

// digestTextSummary 生成纯文本摘要（邮件纯文本备选/文本格式）。
func digestTextSummary(payload *DigestPayload) string {
	s := fmt.Sprintf("通知摘要 (%s — %s)\n\n",
		payload.PeriodStart.Format("2006-01-02 15:04"),
		payload.PeriodEnd.Format("2006-01-02 15:04"))
	s += fmt.Sprintf("共 %d 条新通知:\n\n", payload.TotalCount)
	for i, item := range payload.Items {
		s += fmt.Sprintf("%d. %s\n", i+1, item.Title)
		if item.Body != "" {
			s += fmt.Sprintf("   %s\n", item.Body)
		}
		if item.ActionURL != "" {
			s += fmt.Sprintf("   链接: %s\n", item.ActionURL)
		}
		s += "\n"
	}
	s += "— Ydsz Plane 自动发送"
	return s
}
