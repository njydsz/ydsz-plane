// Command api 启动 Ydsz Plane 的 HTTP API 服务。
//
//	@title			Ydsz Plane API
//	@version		1.0.0
//	@description	项目管理平台 Ydsz Plane 的 REST API（需求/任务/缺陷/迭代/版本/自动化/Webhook/知识库/仪表盘/全局搜索/通知等）。
//	@host			localhost:8080
//	@BasePath		/api/v1
//
//	@securityDefinitions.apikey	Bearer
//	@in							header
//	@name						Authorization
//	@description				"格式: Bearer {access_token}。登录/注册接口返回 access_token 后填入。"
//
//	@tag.name			auth
//	@tag.description	认证与账号域
//	@tag.name			workspace
//	@tag.description	工作空间域
//	@tag.name			project
//	@tag.description	项目域
//	@tag.name			issue
//	@tag.description	需求/任务/缺陷域
//	@tag.name			sprint
//	@tag.description	迭代域
//	@tag.name			version
//	@tag.description	版本域
//	@tag.name			automation
//	@tag.description	自动化域
//	@tag.name			webhook
//	@tag.description	Webhook 集成域
//	@tag.name			knowledge
//	@tag.description	知识库域
//	@tag.name			ai
//	@tag.description	AI 智能域
//	@tag.name			search
//	@tag.description	全局搜索域
//	@tag.name			notification
//	@tag.description	通知域
//	@tag.name			intake
//	@tag.description	收件箱/匿名提报域
//	@tag.name			metrics
//	@tag.description	效能度量域
//	@tag.name			dashboard
//	@tag.description	仪表盘域
//	@tag.name			dlq
//	@tag.description	死信队列管理域
//	@tag.name			pages
//	@tag.description	页面管理域
//	@tag.name			preference
//	@tag.description	偏好设置域
//	@tag.name			workbench
//	@tag.description	工作台域
//	@tag.name			user
//	@tag.description	当前用户域
//	@tag.name			attachment
//	@tag.description	附件域
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"

	"github.com/njydsz/ydsz-plane/internal/application/ai"
	"github.com/njydsz/ydsz-plane/internal/application/aiagent"
	"github.com/njydsz/ydsz-plane/internal/application/apitoken"
	"github.com/njydsz/ydsz-plane/internal/application/attachment"
	"github.com/njydsz/ydsz-plane/internal/application/auth"
	"github.com/njydsz/ydsz-plane/internal/application/automation"
	"github.com/njydsz/ydsz-plane/internal/application/calendar"
	"github.com/njydsz/ydsz-plane/internal/application/collaboration"
	"github.com/njydsz/ydsz-plane/internal/application/customer"
	"github.com/njydsz/ydsz-plane/internal/application/dashboard"
	"github.com/njydsz/ydsz-plane/internal/application/datajob"
	"github.com/njydsz/ydsz-plane/internal/application/dlp"
	"github.com/njydsz/ydsz-plane/internal/application/dlq"
	"github.com/njydsz/ydsz-plane/internal/application/godmode"
	"github.com/njydsz/ydsz-plane/internal/application/grant"
	"github.com/njydsz/ydsz-plane/internal/application/importer"
	"github.com/njydsz/ydsz-plane/internal/application/initiative"
	"github.com/njydsz/ydsz-plane/internal/application/intake"
	"github.com/njydsz/ydsz-plane/internal/application/integration"
	"github.com/njydsz/ydsz-plane/internal/application/issue"
	"github.com/njydsz/ydsz-plane/internal/application/knowledge"
	"github.com/njydsz/ydsz-plane/internal/application/marketplace"
	"github.com/njydsz/ydsz-plane/internal/application/mcpserver"
	"github.com/njydsz/ydsz-plane/internal/application/metrics"
	"github.com/njydsz/ydsz-plane/internal/application/milestone"
	notif "github.com/njydsz/ydsz-plane/internal/application/notification"
	"github.com/njydsz/ydsz-plane/internal/application/oauthapp"
	"github.com/njydsz/ydsz-plane/internal/application/pagecomment"
	"github.com/njydsz/ydsz-plane/internal/application/pages"
	"github.com/njydsz/ydsz-plane/internal/application/pql"
	"github.com/njydsz/ydsz-plane/internal/application/preference"
	"github.com/njydsz/ydsz-plane/internal/application/projectcode"
	"github.com/njydsz/ydsz-plane/internal/application/projectstate"
	"github.com/njydsz/ydsz-plane/internal/application/recurring"
	"github.com/njydsz/ydsz-plane/internal/application/search"
	"github.com/njydsz/ydsz-plane/internal/application/sprint"
	"github.com/njydsz/ydsz-plane/internal/application/teamspace"
	"github.com/njydsz/ydsz-plane/internal/application/version"
	"github.com/njydsz/ydsz-plane/internal/application/webhook"
	"github.com/njydsz/ydsz-plane/internal/application/workbench"
	"github.com/njydsz/ydsz-plane/internal/application/workflow"
	"github.com/njydsz/ydsz-plane/internal/application/workitemtype"
	"github.com/njydsz/ydsz-plane/internal/application/workspace"
	"github.com/njydsz/ydsz-plane/internal/application/workspaceview"
	"github.com/njydsz/ydsz-plane/internal/config"
	"github.com/njydsz/ydsz-plane/internal/infrastructure/cache"
	"github.com/njydsz/ydsz-plane/internal/infrastructure/crdt"
	"github.com/njydsz/ydsz-plane/internal/infrastructure/llm"
	"github.com/njydsz/ydsz-plane/internal/infrastructure/mail"
	"github.com/njydsz/ydsz-plane/internal/infrastructure/persistence"
	"github.com/njydsz/ydsz-plane/internal/infrastructure/storage"
	"github.com/njydsz/ydsz-plane/internal/infrastructure/telemetry"
	"github.com/njydsz/ydsz-plane/internal/infrastructure/ws"
	httpapi "github.com/njydsz/ydsz-plane/internal/interfaces/http"
	"github.com/njydsz/ydsz-plane/internal/rbac"
	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// recurringIssueCreator 把 recurring.WorkItemTemplate 翻译为 issue.CreateIssueInput，
// 实现 recurring.WorkItemCreator 接口。用于重复规则自动生成工作项。
type recurringIssueCreator struct {
	issueSvc *issue.Service
}

// CreateWorkItem 按模板创建单一工作项（缺陷类型需额外字段时由调用方在规则中避免）。
func (a recurringIssueCreator) CreateWorkItem(ctx context.Context, in recurring.WorkItemTemplate) error {
	tc := issue.IssueTypeCode(in.TypeCode)
	if tc == "defect" {
		// 缺陷必须提供 severity / found_phase，重复规则暂不支持自动生成缺陷，
		// 通过校验错误显式拒绝而非生成非法数据。
		return errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "work_item_type", Reason: "重复规则暂不支持自动生成缺陷"})
	}
	priority := issue.IssuePriority(in.Priority)
	_, err := a.issueSvc.Create(ctx, issue.CreateIssueInput{
		WorkspaceID:     in.WorkspaceID,
		ProjectID:       in.ProjectID,
		TypeCode:        tc,
		Name:            in.Name,
		DescriptionHTML: in.DescriptionHTML,
		StateID:         in.StateID,
		Priority:        priority,
		Assignees:       in.Assignees,
		Labels:          in.Labels,
		CreatedBy:       in.CreatedBy,
	})
	return err
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "api: fatal:", err)
		os.Exit(1)
	}
}

// run 引导并启动 HTTP API 服务，阻塞直至收到退出信号。
//
// 初始化遵循严格的依赖顺序，确保每个下游组件都能拿到完整初始化的依赖：
//
//  1. 配置 + Logger —— 其余一切组件都依赖解析后的配置与可用的日志器。
//  2. PostgreSQL 连接池 —— 各服务启动时会对数据库 schema 做校验。
//  3. Redis 客户端 —— 用于会话、限流与 outbox 汇聚。
//  4. 领域服务 —— 在 DB 与 Redis 就绪后装配，便于各服务在底层存储不可达时快速失败。
//  5. HTTP 引擎与路由 —— 所有 handler 注册完成后最后挂载。
//  6. 监听与服务 —— 服务在独立 goroutine 中启动；主 goroutine 通过 select{}
//     等待信号或 ListenAndServe 错误。
//
// 收到 SIGINT/SIGTERM 后执行优雅关闭，排水超时 15 秒，之后关闭剩余连接并退出。
//
// run 返回非 nil 时进程以退出码 1 终止。
func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	log, err := telemetry.NewLogger(cfg.Log.Level, cfg.Log.Format)
	if err != nil {
		return err
	}
	defer func() { _ = log.Sync() }()

	// 安全提示：开发模式下 JWT 密钥每次重启都会轮换，重启前签发的 token
	// 将全部失效。这是有意设计，避免代码库中出现硬编码密钥。
	if cfg.IsDev() && strings.HasPrefix(cfg.Auth.JWTSecret, "dev-") {
		log.Warn("auth: using ephemeral dev JWT secret (set YDSZ_AUTH_JWT_SECRET to pin it)")
	}
	ctx := context.Background()

	pool, err := persistence.NewPool(ctx, cfg.Database.URL, cfg.Database.MaxConns)
	if err != nil {
		return err
	}
	defer pool.Close()

	// P1-6: 启动连接池 Prometheus 指标采集（每 15s 一轮）。
	// Uber Go Style: "defer ticker.Stop() guarantees cleanup."
	poolCollector := telemetry.NewPoolStatsCollector(pool.Pool)
	poolTicker := time.NewTicker(15 * time.Second)
	go poolCollector.Start(ctx, poolTicker)
	defer poolTicker.Stop()

	rdb, err := cache.NewClient(ctx, cfg.Redis.Addr, cfg.Redis.Password, cfg.Redis.DB)
	if err != nil {
		return err
	}
	defer func() { _ = rdb.Close() }()

	// ---------- RBAC (DB-backed) ----------
	rbacStore := rbac.NewStore(pool.Pool, log)
	if err := rbacStore.InitCache(ctx); err != nil {
		log.Warn("rbac: cache warm-up failed (will fallback to DB)", zap.Error(err))
	}

	// ---------- Services ----------
	authSvc := auth.NewService(pool.Pool, cfg.Auth.JWTSecret, cfg.Auth.JWTIssuer,
		cfg.Auth.AccessTokenTTL, cfg.Auth.RefreshTokenTTL, cfg.Auth.BcryptCost,
		cfg.Features.RegistrationOpen)
	resetSvc := auth.NewPasswordResetService(pool.Pool, nil, cfg.Email.AppBaseURL, cfg.Auth.BcryptCost)

	// S13 OIDC / SSO 服务（单点登录）；仅在有 SSO Provider 配置时可用
	// P0-2：client_secret 加密封存，需从环境变量 YDSZ_SSO_SECRET_KEY 加载 32 字节密钥
	var oidcService *auth.OIDCService
	if ssoCipher, cipherErr := auth.NewSecretCipher(); cipherErr != nil {
		log.Warn("SSO secret cipher not configured (client_secret will be stored in plaintext): "+cipherErr.Error(),
			zap.String("hint", "set YDSZ_SSO_SECRET_KEY env (32 bytes) to enable encryption"))
		oidcService = auth.NewOIDCService(pool.Pool, authSvc, cfg.Email.AppBaseURL)
	} else {
		// P0-2 安全加固：启动时密钥连通性自检（encrypt → decrypt round-trip）
		// 密钥错误时 fail-fast 阻断启动，避免运行时加解密失败导致 SSO 异常
		if selfTestErr := ssoCipher.SelfTest(); selfTestErr != nil {
			log.Fatal("SSO secret cipher self-test failed (YDSZ_SSO_SECRET_KEY may be corrupted)",
				zap.Error(selfTestErr))
			return selfTestErr
		}
		log.Info("SSO secret cipher initialized and self-test passed",
			zap.String("key_source", "YDSZ_SSO_SECRET_KEY"))
		oidcService = auth.NewOIDCServiceWithCipher(pool.Pool, authSvc, cfg.Email.AppBaseURL, ssoCipher)
	}

	wsStore := auth.NewWorkspaceMembershipStore(pool.Pool)
	wsSvc := workspace.NewService(pool.Pool)
	memberSvc := workspace.NewMemberService(pool.Pool)
	projectSvc := workspace.NewProjectService(pool.Pool)
	projectMemberSvc := workspace.NewProjectMemberService(pool.Pool)
	apiTokenSvc := apitoken.NewService(pool.Pool)

	// 复合认证解析器：先试会话 JWT，失败后尝试个人 API Token（ydz_ 前缀）。
	// 两种凭证最终都归一为 auth.Principal，中间件据此区分来源并施加 scope 收敛。
	parsePrincipal := func(token string) (auth.Principal, error) {
		if uid, err := authSvc.ParseAccess(token); err == nil {
			return auth.Principal{UserID: uid, Kind: auth.PrincipalJWT}, nil
		}
		return apiTokenSvc.ResolvePrincipal(ctx, token)
	}

	// 邮件服务：未配置 SMTP → Noop（dev/CI 不打真实邮件）
	var mailSvc mail.EmailService
	if cfg.Email.Enabled {
		mailSvc = mail.NewSmtpService(mail.SmtpConfig{
			Host: cfg.Email.Host, Port: cfg.Email.Port,
			Username: cfg.Email.Username, Password: cfg.Email.Password,
			From: cfg.Email.From, UseTLS: cfg.Email.UseTLS,
		})
	} else {
		mailSvc = mail.NewNoopService(0)
		log.Info("email: disabled (no-op mode) — set YDSZ_EMAIL_ENABLED=true to enable SMTP")
	}

	invitationSvc := workspace.NewInvitationService(pool.Pool, mailSvc, cfg.Email.AppBaseURL)
	auditSvc := workspace.NewAuditService(pool.Pool)

	// ---------- Notification domain (before Issue, for event fan-out) ----------
	notifSvc := notif.NewService(pool.Pool)
	notifHandler := notif.NewHandler(&notif.HandlerDeps{
		NotificationSvc: notifSvc,
		WorkspaceStore:  wsStore,
	})

	// ---------- WebSocket Hub (before Issue, for real-time broadcast) ----------
	// 生产环境安全配置：同源校验 + 连接级限流
	// P1-10：传入 prometheus.DefaultRegisterer 以注册 WS 指标（clientsActive/eventsPublished/eventsDropped/broadcastDuration）
	wsHub := ws.NewHubWithConfig(rdb, ws.HubConfig{
		MaxConnsPerUser:    3,
		MaxConnsPerIP:      10,
		MaxConnsGlobal:     1000,
		RequireOriginCheck: cfg.Server.Env != "development",
		Registerer:         prometheus.DefaultRegisterer,
	})
	go wsHub.Run()
	defer wsHub.Shutdown()

	// ---------- Issue domain ----------
	issueSvc := issue.NewService(pool.Pool)
	stateSvc := issue.NewStateService(pool.Pool)
	activitySvc := issue.NewActivityService(pool.Pool)
	timeLogSvc := issue.NewTimeLogService(pool.Pool)
	relationSvc := issue.NewRelationService(pool.Pool)
	// ---------- AI Agent（含 LLM 网关） ----------
	aiAgentSvc := aiagent.NewService(pool.Pool)
	llmPrimary, llmFallback, llmOK := llm.LoadFromEnv()
	llmClient := llm.NewClient(llmPrimary, llmFallback, llmOK)
	aiAgentRunSvc := aiagent.NewRunService(pool.Pool)
	aiAgentHandler := aiagent.NewHandler(&aiagent.HandlerDeps{
		Svc:    aiAgentSvc,
		LLM:    aiagent.NewLLMBridge(llmClient, pool.Pool, aiagent.NewWorkspaceToolProvider(pool.Pool)),
		RunSvc: aiAgentRunSvc,
	})

	// Agent 触发器钩子（mention / assignment / event / schedule）
	triggerDispatcher := aiagent.NewTriggerDispatcher(aiAgentSvc, aiAgentRunSvc,
		aiagent.NewLLMBridge(llmClient, pool.Pool, aiagent.NewWorkspaceToolProvider(pool.Pool)))
	triggerHook := aiagent.NewTriggerHook(triggerDispatcher)
	// 注入工作项事件触发（创建/状态变更）
	issueSvc.SetAgentTrigger(triggerHook)

	// ---------- 导入（Jira/Linear/Asana/CSV） ----------
	// ---------- 应用商店 ----------
	marketplaceSvc := marketplace.NewService(pool.Pool)
	marketplaceHandler := marketplace.NewHandler(&marketplace.HandlerDeps{Svc: marketplaceSvc})
	if err := marketplaceSvc.SeedIfEmpty(context.Background()); err != nil {
		slog.Warn("marketplace seed failed", "err", err)
	}

	// ---------- DLP 敏感数据扫描 ----------
	dlpSvc := dlp.NewService(pool.Pool)
	dlpHandler := dlp.NewHandler(&dlp.HandlerDeps{Svc: dlpSvc})

	// ---------- CRDT 协同（含 Redis 跨节点广播） ----------
	crdtStore := crdt.NewPGStore(pool.Pool)
	crdtHub := crdt.NewHubWithStore(crdtStore, 0, 0) // 0 = 不限流（由网关层控制）
	// 多节点部署时启用 Redis 广播：同一文档的更新会同步到其他节点
	if cfg.Redis.Addr != "" {
		if redisClient, rerr := cache.NewClient(context.Background(), cfg.Redis.Addr, cfg.Redis.Password, cfg.Redis.DB); rerr == nil {
			hostname, _ := os.Hostname()
			nodeID := fmt.Sprintf("%s-%d", hostname, cfg.Server.Port)
			broadcaster := crdt.NewRedisBroadcaster(redisClient, nodeID)
			crdtHub.WithRedisBroadcaster(broadcaster)
			broadcaster.StartDynamic(context.Background())
			slog.Info("crdt redis broadcaster enabled", "node", nodeID)
		} else {
			slog.Warn("crdt redis broadcaster disabled", "err", rerr)
		}
	}

	importerSvc := importer.NewService(pool.Pool)
	importerHandler := importer.NewHandler(&importer.HandlerDeps{Svc: importerSvc})

	// ---------- 条件授权（ABAC） ----------
	grantSvc := grant.NewService(pool.Pool)
	grantHandler := grant.NewHandler(&grant.HandlerDeps{Svc: grantSvc})

	// ---------- God Mode 实例管理 ----------
	godModeSvc := godmode.NewService()
	godModeHandler := godmode.NewHandler(&godmode.HandlerDeps{Svc: godModeSvc})
	if !llmOK {
		slog.Warn("llm not configured; ai-agent chat will return a configuration notice",
			"hint", "set YDSZ_LLM_API_BASE / YDSZ_LLM_API_KEY / YDSZ_LLM_MODEL")
	}

	commentSvc := issue.NewCommentService(pool.Pool)
	commentSvc.SetAgentTrigger(triggerHook)
	contentTplSvc := issue.NewContentTemplateService(pool.Pool)
	issueHandler := issue.NewIssueHandler(&issue.HandlerDeps{
		IssueSvc:        issueSvc,
		StateSvc:        stateSvc,
		ActivitySvc:     activitySvc,
		TimeLogSvc:      timeLogSvc,
		RelationSvc:     relationSvc,
		CommentSvc:      commentSvc,
		ContentTplSvc:   contentTplSvc,
		ImportSvc:       issue.NewImportService(issueSvc),
		WorkspaceStore:  wsStore,
		NotificationSvc: notifSvc,
		WSHub:           wsHub,
		UserNameQuery: func(ctx context.Context, userID int64) string {
			var name string
			_ = pool.Pool.QueryRow(ctx,
				`SELECT COALESCE(display_name,'') FROM users WHERE id=$1`, userID).Scan(&name)
			return name
		},
	})

	// ---------- Sprint domain ----------
	sprintSvc := sprint.NewService(pool.Pool)
	sprintHandler := sprint.NewHandler(sprintSvc,
		sprint.WithNotification(notifSvc),
		sprint.WithWSHub(wsHub),
	)

	// ---------- Version domain ----------
	versionSvc := version.NewService(version.Deps{
		DB:    pool.Pool,
		Redis: rdb,
		Audit: auditSvc,
		Log:   log,
	})
	versionHandler := version.NewHandler(versionSvc)

	// ---------- Knowledge domain ----------
	knowledgeSvc := knowledge.NewService(pool.Pool)
	knowledgeHandler := knowledge.NewHandler(knowledgeSvc)

	// ---------- Search domain ----------
	searchSvc := search.NewService(pool.Pool)
	searchIndexer := search.NewIndexer(pool.Pool)
	searchHandler := search.NewSearchHandler(&search.HandlerDeps{
		SearchSvc:      searchSvc,
		Indexer:        searchIndexer,
		WorkspaceStore: wsStore,
	})

	// ---------- Workbench domain ----------
	workbenchSvc := workbench.NewService(pool.Pool)
	workbenchHandler := workbench.NewWorkbenchHandler(&workbench.HandlerDeps{
		WorkbenchSvc: workbenchSvc,
	})

	// ---------- Dashboard domain ----------
	dashboardSvc := dashboard.NewService(pool.Pool)
	dashboardHandler := dashboard.NewDashboardHandler(&dashboard.HandlerDeps{
		DashboardSvc: dashboardSvc,
	})

	// ---------- Defect Analytics (S12) ----------
	defectAnalyticsSvc := issue.NewDefectAnalyticsService(pool.Pool)
	defectAnalyticsHandler := issue.NewDefectAnalyticsHandler(defectAnalyticsSvc)

	// ---------- Label / Module / EstimatePoint handlers ----------
	// P0-1 修复：标签与模块 Service + Handler 已就绪，实例化并挂载路由
	labelSvc := issue.NewLabelService(pool.Pool)
	labelHandler := issue.NewLabelHandler(labelSvc)
	moduleSvc := issue.NewModuleService(pool.Pool)
	moduleHandler := issue.NewModuleHandler(moduleSvc)
	estimatePointSvc := issue.NewEstimatePointService(pool.Pool)
	estimatePointHandler := issue.NewEstimatePointHandler(estimatePointSvc)

	// ---------- Issue domain init (templates) ----------
	projectInitSvc := issue.NewProjectInitService(pool.Pool)

	// ---------- Workspace Template service ----------
	templateSvc := workspace.NewTemplateService()

	// ---------- Attachment / Storage ----------
	stClient, err := storage.New(cfg.Storage)
	if err != nil {
		return fmt.Errorf("storage: %w", err)
	}
	attSvc := attachment.NewService(pool.Pool, stClient)
	attHandler := attachment.NewHandler(&attachment.HandlerDeps{
		AttachmentSvc: attSvc,
		Cfg:           &cfg.Attachment,
		DB:            pool.Pool,
	})

	// ---------- Automation domain (S11) ----------
	automationSvc := automation.NewService(pool.Pool)
	automationHandler := automation.NewHandler(&automation.HandlerDeps{
		Svc: automationSvc,
	})

	// ---------- Metrics domain (S11) ----------
	// 使用带 Redis 读穿透缓存的 Handler：velocity/lead-time/quality/dora 等
	// 高频仪表盘查询命中缓存（差异化 TTL + singleflight 防击穿），
	// cfd/control-chart/throughput 等分析端点直查 DB。
	// cli 为 nil（Redis 不可用）时自动降级为无缓存 Handler。
	metricsSvc := metrics.NewService(pool.Pool)
	metricsHandler := metrics.NewCachedHandler(&metrics.HandlerDeps{
		Svc: metricsSvc,
	}, rdb, log)

	// ---------- View Preference ----------
	prefSvc := preference.NewService(pool.Pool)
	prefHandler := preference.NewHandler(prefSvc)

	// ---------- Saved Views ----------
	viewSvc := preference.NewViewService(pool.Pool)
	if err := viewSvc.EnsureSchema(ctx); err != nil {
		log.Warn("view: schema ensure failed (will retry on next restart)", zap.Error(err))
	}
	viewHandler := preference.NewViewHandler(viewSvc)

	// ---------- Pages (文档页面) ----------
	pagesSvc := pages.NewService(pool.Pool)
	pagesHandler := pages.NewHandler(pagesSvc)
	pagesPublicHandler := pages.NewPublicShareHandler(pagesSvc)

	// ---------- Webhook (S10) ----------
	webhookSvc := webhook.NewService(pool.Pool)
	// Dispatcher 的 MQ 依赖（enqueueRetry 自动重试入队）仅在 worker 进程
	// 的事件投递链路中使用；API 进程内的 dispatcher 仅服务于管理页的
	// 同步操作（测试推送 / 手动重投 RetryLog），二者均为同步投递、不依赖
	// MQ。因此这里刻意传 nil，避免 API 进程引入 RabbitMQ 运行时依赖
	// （保持 API 无状态、RabbitMQ 故障不阻塞管理操作）。
	webhookDispatcher := webhook.NewDispatcher(webhookSvc, nil, log)
	webhookHandler := webhook.NewHandler(&webhook.HandlerDeps{
		WebhookSvc:     webhookSvc,
		WorkspaceStore: wsStore,
		Dispatcher:     webhookDispatcher,
	})

	// ---------- AI domain ----------
	aiSvc := ai.NewService(pool.Pool, ai.Config{
		Enabled:  cfg.AI.Enabled,
		Provider: cfg.AI.Provider,
		APIKey:   cfg.AI.APIKey,
		Model:    cfg.AI.Model,
		Endpoint: cfg.AI.Endpoint,
	})
	aiHandler := ai.NewHandler(&ai.HandlerDeps{AiSvc: aiSvc})

	// ---------- DLQ 管理（死信队列监控） ----------
	dlqSvc := dlq.NewService(pool.Pool)
	dlqHandler := dlq.NewHandler(dlqSvc)

	// ---------- Intake 收件箱（匿名提报） ----------
	// 转正能力依赖 issue.Service（在 Issue domain 之后装配）。
	intakeSvc := intake.NewService(pool.Pool).WithIssueCreator(issueSvc)
	intakeHandler := intake.NewHandler(&intake.HandlerDeps{Svc: intakeSvc})
	intakePublicHandler := intake.NewPublicHandler(&intake.HandlerDeps{Svc: intakeSvc})

	// ---------- Calendar 日历事件（独立事件：会议/里程碑/发布日） ----------
	calendarSvc := calendar.NewService(pool.Pool)
	calendarHandler := calendar.NewHandler(&calendar.HandlerDeps{Svc: calendarSvc})

	// ---------- DataJob 异步数据任务（导入/导出/同步/报表） ----------
	dataJobSvc := datajob.NewService(pool.Pool)
	dataJobHandler := datajob.NewHandler(&datajob.HandlerDeps{Svc: dataJobSvc})

	// ---------- Initiative 倡议（跨项目战略目标的进度聚合） ----------
	initiativeSvc := initiative.NewService(pool.Pool)
	initiativeHandler := initiative.NewHandler(&initiative.HandlerDeps{Svc: initiativeSvc})

	// ---------- Workflow 工作流与审批 ----------
	workflowSvc := workflow.NewService(pool.Pool)
	workflowHandler := workflow.NewHandler(&workflow.HandlerDeps{Svc: workflowSvc})

	// ---------- Teamspace 团队空间 ----------
	teamspaceSvc := teamspace.NewService(pool.Pool)
	teamspaceHandler := teamspace.NewHandler(&teamspace.HandlerDeps{Svc: teamspaceSvc})

	// ---------- ProjectState 项目状态与更新 ----------
	projectStateSvc := projectstate.NewService(pool.Pool)
	projectStateHandler := projectstate.NewHandler(&projectstate.HandlerDeps{Svc: projectStateSvc})

	// ---------- Projects-as-Code（YAML 定义项目） ----------
	projectCodeSvc := projectcode.NewService(pool.Pool)
	projectCodeHandler := projectcode.NewHandler(&projectcode.HandlerDeps{Svc: projectCodeSvc})

	// ---------- Collaboration 实时协作 ----------
	collabSvc := collaboration.NewService(pool.Pool)
	collabHandler := collaboration.NewHandler(&collaboration.HandlerDeps{Svc: collabSvc})

	// ---------- PQL 查询语言 ----------
	pqlSvc := pql.NewService()
	pqlHandler := pql.NewHandler(pqlSvc)

	// ---------- Customer 客户管理（CRM） ----------
	customerSvc := customer.NewService(pool.Pool)
	customerHandler := customer.NewHandler(&customer.HandlerDeps{Svc: customerSvc})

	// ---------- Work Item Type 自定义工作项类型 ----------
	workItemTypeSvc := workitemtype.NewService(pool.Pool)
	workItemTypeHandler := workitemtype.NewHandler(&workitemtype.HandlerDeps{Svc: workItemTypeSvc})

	// ---------- Page Comment 页面评论（页面级 / 行内评论） ----------
	pageCommentSvc := pagecomment.NewService(pool.Pool)
	pageCommentHandler := pagecomment.NewHandler(&pagecomment.HandlerDeps{Svc: pageCommentSvc})

	// ---------- Recurring 重复工作项（cron 模板自动生成） ----------
	// 通过适配器把 recurring 的 WorkItemTemplate 投影转发到 issue.Service.Create，
	// 避免 recurring 包直接依赖 issue 包的编译期耦合。
	recurringSvc := recurring.NewService(pool.Pool).WithWorkItemCreator(recurringIssueCreator{issueSvc: issueSvc})
	recurringHandler := recurring.NewHandler(&recurring.HandlerDeps{Svc: recurringSvc})

	// ---------- Milestone 里程碑 ----------
	milestoneSvc := milestone.NewService(pool.Pool)
	milestoneHandler := milestone.NewHandler(&milestone.HandlerDeps{Svc: milestoneSvc})

	// ---------- Workspace View 工作空间视图 ----------
	workspaceViewSvc := workspaceview.NewService(pool.Pool)
	workspaceViewHandler := workspaceview.NewHandler(&workspaceview.HandlerDeps{Svc: workspaceViewSvc})

	// ---------- Integration 集成（Marketplace + 第三方连接器） ----------
	// 凭证加密密钥来源：YDSZ_INTEGRATION_SECRET_KEY（32 字节）。
	// 未配置时服务降级为明文存储（仅开发/测试；生产必须配置，见 credentials.go 说明）。
	integrationCipher, cipherErr := integration.NewCredentialCipher()
	if cipherErr != nil {
		log.Warn("integration: credential cipher not configured (credentials will be stored in plaintext): "+cipherErr.Error(),
			zap.String("hint", "set YDSZ_INTEGRATION_SECRET_KEY env (32 bytes) to enable encryption"))
		integrationCipher = nil
	}
	integrationSvc := integration.NewService(pool.Pool, integrationCipher)
	integrationHandler := integration.NewHandler(&integration.HandlerDeps{Svc: integrationSvc})

	// ---------- OAuth 应用与授权（开发者应用 + OAuth2 端点） ----------
	oauthAppSvc := oauthapp.NewService(pool.Pool)
	oauthAppHandler := oauthapp.NewHandler(&oauthapp.HandlerDeps{Svc: oauthAppSvc})

	// ---------- HTTP Engine ----------
	engine := httpapi.NewEngine(&httpapi.Deps{
		Cfg:              cfg,
		Log:              log,
		DB:               pool.Pool,
		Redis:            rdb,
		Auth:             authSvc,
		ResetSvc:         resetSvc,
		PrincipalParser:  parsePrincipal,
		ApiTokenSvc:      apiTokenSvc,
		WorkspaceStore:   wsStore,
		RBACStore:        rbacStore,
		WorkspaceSvc:     wsSvc,
		MemberSvc:        memberSvc,
		InvitationSvc:    invitationSvc,
		ProjectSvc:       projectSvc,
		ProjectMemberSvc: projectMemberSvc,
		TemplateSvc:      templateSvc,
		ProjectInitSvc:   projectInitSvc,
		AuditSvc:         auditSvc,
		Mail:             mailSvc,
		// Issue domain
		IssueSvc:             issueSvc,
		StateSvc:             stateSvc,
		ActivitySvc:          activitySvc,
		TimeLogSvc:           timeLogSvc,
		IssueHandler:         issueHandler,
		LabelHandler:         labelHandler,
		ModuleHandler:        moduleHandler,
		EstimatePointHandler: estimatePointHandler,
		// Search domain
		SearchHandler: searchHandler,
		// Workbench domain
		WorkbenchHandler: workbenchHandler,
		// Dashboard domain
		DashboardHandler: dashboardHandler,
		// Notification domain
		NotificationHandler: notifHandler,
		// Attachment domain
		AttachmentHandler: attHandler,
		// Defect Analytics
		DefectAnalyticsHandler: defectAnalyticsHandler,
		// WebSocket Hub
		WSHub: wsHub,
		// Sprint domain
		SprintHandler: sprintHandler,
		// Version domain
		VersionHandler: versionHandler,
		// Knowledge domain (S9)
		KnowledgeHandler: knowledgeHandler,
		// Webhook domain (S10)
		WebhookHandler: webhookHandler,
		// Pages 公开分享域
		PagesPublicHandler: pagesPublicHandler,
		// Automation domain (S11)
		AutomationHandler: automationHandler,
		// Metrics domain (S11)
		MetricsHandler: metricsHandler,
		// AI domain
		AiHandler: aiHandler,
		// DLQ 域（死信队列管理）
		DLQHandler: dlqHandler,
		// Intake 收件箱域（匿名提报）
		IntakeHandler:       intakeHandler,
		IntakePublicHandler: intakePublicHandler,
		// 倡议域（跨项目战略目标的进度聚合）
		InitiativeHandler: initiativeHandler,
		// S13 OIDC / SSO 域（单点登录）
		OIDCService: oidcService,
		// Storage 客户端（Logo 上传等）
		Storage: stClient,
		// 自定义工作项类型域
		WorkItemTypeHandler: workItemTypeHandler,
		// 客户管理域（CRM：客户记录 + 客户请求）
		CustomerHandler: customerHandler,
		// 页面评论域（页面级 / 行内评论、线程、表情回应）
		PageCommentHandler: pageCommentHandler,
		// P2 增强域
		RecurringHandler:     recurringHandler,
		MilestoneHandler:     milestoneHandler,
		WorkspaceViewHandler: workspaceViewHandler,
		// 集成域（Marketplace + GitHub/GitLab/Slack/Sentry 连接器）
		IntegrationHandler: integrationHandler,
		// OAuth 应用与授权域（开发者应用 + OAuth2 端点）
		OAuthAppHandler: oauthAppHandler,
		OAuthAppSvc:     oauthAppSvc,
	})

	// 注册需求/任务/缺陷路由（必须在 NewEngine 之后）
	httpapi.RegisterIssueRoutes(engine, &httpapi.Deps{
		Auth:            authSvc,
		PrincipalParser: parsePrincipal,
		WorkspaceStore:  wsStore,
		RBACStore:       rbacStore,
		IssueHandler:    issueHandler,
	})

	// 注册迭代路由（独立于 Issue 路由）
	httpapi.RegisterSprintRoutes(engine, &httpapi.Deps{
		Auth:            authSvc,
		PrincipalParser: parsePrincipal,
		WorkspaceStore:  wsStore,
		SprintHandler:   sprintHandler,
	})

	// 注册版本路由（独立于 Issue 路由）
	httpapi.RegisterVersionRoutes(engine, &httpapi.Deps{
		Auth:            authSvc,
		PrincipalParser: parsePrincipal,
		WorkspaceStore:  wsStore,
		VersionHandler:  versionHandler,
	})

	// 注册知识库路由（工作空间级，可选项目级过滤）
	httpapi.RegisterKnowledgeRoutes(engine, &httpapi.Deps{
		Auth:             authSvc,
		PrincipalParser:  parsePrincipal,
		WorkspaceStore:   wsStore,
		RBACStore:        rbacStore,
		KnowledgeHandler: knowledgeHandler,
	})

	// 注册视图偏好路由（项目级）
	httpapi.RegisterPreferenceRoutes(engine, &httpapi.Deps{
		Auth:            authSvc,
		PrincipalParser: parsePrincipal,
		WorkspaceStore:  wsStore,
		PrefHandler:     prefHandler,
		RBACStore:       rbacStore,
	})

	// 注册命名视图管理路由（项目级）
	httpapi.RegisterViewRoutes(engine, &httpapi.Deps{
		Auth:            authSvc,
		PrincipalParser: parsePrincipal,
		WorkspaceStore:  wsStore,
		ViewHandler:     viewHandler,
		RBACStore:       rbacStore,
	})

	// 注册页面路由（项目级）
	httpapi.RegisterPagesRoutes(engine, &httpapi.Deps{
		Auth:            authSvc,
		PrincipalParser: parsePrincipal,
		WorkspaceStore:  wsStore,
		RBACStore:       rbacStore,
		PagesHandler:    pagesHandler,
	})

	// 注册搜索路由（项目级 + 工作空间级）
	httpapi.RegisterSearchRoutes(engine, &httpapi.Deps{
		Auth:            authSvc,
		PrincipalParser: parsePrincipal,
		WorkspaceStore:  wsStore,
		RBACStore:       rbacStore,
		SearchHandler:   searchHandler,
	})

	// 注册工作台路由（项目级 + 工作空间级）
	httpapi.RegisterWorkbenchRoutes(engine, &httpapi.Deps{
		Auth:             authSvc,
		PrincipalParser:  parsePrincipal,
		WorkspaceStore:   wsStore,
		WorkbenchHandler: workbenchHandler,
	})

	// 注册仪表盘路由（项目级 + 工作空间级）
	httpapi.RegisterDashboardRoutes(engine, &httpapi.Deps{
		Auth:             authSvc,
		PrincipalParser:  parsePrincipal,
		WorkspaceStore:   wsStore,
		RBACStore:        rbacStore,
		DashboardHandler: dashboardHandler,
	})

	// 注册通知路由
	httpapi.RegisterNotificationRoutes(engine, &httpapi.Deps{
		Auth:                authSvc,
		PrincipalParser:     parsePrincipal,
		WorkspaceStore:      wsStore,
		RBACStore:           rbacStore,
		NotificationHandler: notifHandler,
	})

	// 注册附件路由
	httpapi.RegisterAttachmentRoutes(engine, &httpapi.Deps{
		Auth:              authSvc,
		PrincipalParser:   parsePrincipal,
		WorkspaceStore:    wsStore,
		RBACStore:         rbacStore,
		AttachmentHandler: attHandler,
	})

	// 注册缺陷分析路由（项目级）
	httpapi.RegisterDefectAnalyticsRoutes(engine, &httpapi.Deps{
		Auth:                   authSvc,
		PrincipalParser:        parsePrincipal,
		WorkspaceStore:         wsStore,
		RBACStore:              rbacStore,
		DefectAnalyticsHandler: defectAnalyticsHandler,
	})

	// 注册 WebSocket 路由
	httpapi.RegisterWSRoutes(engine, &httpapi.Deps{
		Auth:            authSvc,
		PrincipalParser: parsePrincipal,
		WorkspaceStore:  wsStore,
		WSHub:           wsHub,
	})

	// 注册 Webhook 管理路由（工作空间级）
	httpapi.RegisterWebhookRoutes(engine, &httpapi.Deps{
		Auth:            authSvc,
		PrincipalParser: parsePrincipal,
		WorkspaceStore:  wsStore,
		RBACStore:       rbacStore,
		WebhookHandler:  webhookHandler,
	})

	// 注册文档公开分享路由（免登录）
	httpapi.RegisterPagesPublicRoutes(engine, &httpapi.Deps{
		PagesPublicHandler: pagesPublicHandler,
	})

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Server.Port),
		Handler: engine,
		// ReadHeaderTimeout 限制客户端发送请求头的时长。10s 对弱网客户端足够宽裕，
		// 同时能快速拒绝 Slowloris 类慢速攻击。
		ReadHeaderTimeout: 10 * time.Second,
		// ReadTimeout 覆盖读取完整请求体的时长。30s 可支撑单请求内的批量 CSV/JSON 上传。
		ReadTimeout: 30 * time.Second,
		// WriteTimeout 限制整个响应写入时长。60s 对分页需求/任务/缺陷列表与迭代报告足够；
		// 需要更长时间的端点应当改用流式输出。
		WriteTimeout: 60 * time.Second,
		// IdleTimeout 回收静默超过 2 分钟的 keep-alive 连接，
		// 在高流量下限制空闲连接占用，同时保留典型突发请求的 keep-alive 收益。
		IdleTimeout: 120 * time.Second,
	}

	// 注册自动化路由（S11，项目级）
	httpapi.RegisterAutomationRoutes(engine, &httpapi.Deps{
		Auth:              authSvc,
		PrincipalParser:   parsePrincipal,
		WorkspaceStore:    wsStore,
		RBACStore:         rbacStore,
		AutomationHandler: automationHandler,
	})

	// 注册效能度量路由（S11，项目级）
	httpapi.RegisterMetricsRoutes(engine, &httpapi.Deps{
		Auth:            authSvc,
		PrincipalParser: parsePrincipal,
		WorkspaceStore:  wsStore,
		RBACStore:       rbacStore,
		MetricsHandler:  metricsHandler,
	})

	// 注册 AI 智能辅助路由（项目级）
	httpapi.RegisterAIRoutes(engine, &httpapi.Deps{
		Auth:            authSvc,
		PrincipalParser: parsePrincipal,
		WorkspaceStore:  wsStore,
		RBACStore:       rbacStore,
		AiHandler:       aiHandler,
	})

	// 注册 DLQ 管理路由（工作空间级）
	httpapi.RegisterDLQRoutes(engine, &httpapi.Deps{
		Auth:            authSvc,
		PrincipalParser: parsePrincipal,
		WorkspaceStore:  wsStore,
		RBACStore:       rbacStore,
		DLQHandler:      dlqHandler,
	})

	// 注册 S13 SSO Provider 管理路由（工作空间级，workspace 管理员配置 SSO 集成）
	httpapi.RegisterSSOProviderRoutes(engine, &httpapi.Deps{
		Auth:            authSvc,
		PrincipalParser: parsePrincipal,
		WorkspaceStore:  wsStore,
		RBACStore:       rbacStore,
		OIDCService:     oidcService,
	})

	// 注册收件箱管理路由（工作空间级）
	httpapi.RegisterIntakeRoutes(engine, &httpapi.Deps{
		Auth:            authSvc,
		PrincipalParser: parsePrincipal,
		WorkspaceStore:  wsStore,
		RBACStore:       rbacStore,
		IntakeHandler:   intakeHandler,
	})

	// 注册收件箱公开路由（免登录：匿名提报 + 提交跟踪）
	httpapi.RegisterIntakePublicRoutes(engine, &httpapi.Deps{
		IntakePublicHandler: intakePublicHandler,
	})

	// 注册日历事件路由（工作空间级：读 workspace:read / 写 project:update）
	httpapi.RegisterCalendarRoutes(engine, &httpapi.Deps{
		Auth:            authSvc,
		PrincipalParser: parsePrincipal,
		WorkspaceStore:  wsStore,
		RBACStore:       rbacStore,
		CalendarHandler: calendarHandler,
	})

	// 注册数据任务路由（工作空间级：异步作业监控）
	httpapi.RegisterDataJobRoutes(engine, &httpapi.Deps{
		Auth:            authSvc,
		PrincipalParser: parsePrincipal,
		WorkspaceStore:  wsStore,
		RBACStore:       rbacStore,
		DataJobHandler:  dataJobHandler,
	})

	// 注册倡议路由（工作空间级：读 workspace:read / 写 project:update）
	httpapi.RegisterInitiativeRoutes(engine, &httpapi.Deps{
		Auth:              authSvc,
		PrincipalParser:   parsePrincipal,
		WorkspaceStore:    wsStore,
		RBACStore:         rbacStore,
		InitiativeHandler: initiativeHandler,
	})

	// 注册工作流与审批路由（项目级配置 + 工作空间级审批待办）
	httpapi.RegisterWorkflowRoutes(engine, &httpapi.Deps{
		Auth:            authSvc,
		PrincipalParser: parsePrincipal,
		WorkspaceStore:  wsStore,
		RBACStore:       rbacStore,
		WorkflowHandler: workflowHandler,
	})

	// 注册团队空间路由
	httpapi.RegisterTeamspaceRoutes(engine, &httpapi.Deps{
		Auth:             authSvc,
		PrincipalParser:  parsePrincipal,
		WorkspaceStore:   wsStore,
		RBACStore:        rbacStore,
		TeamspaceHandler: teamspaceHandler,
	})

	// 注册实时协作路由
	httpapi.RegisterCollaborationRoutes(engine, &httpapi.Deps{
		Auth:                 authSvc,
		PrincipalParser:      parsePrincipal,
		WorkspaceStore:       wsStore,
		RBACStore:            rbacStore,
		CollaborationHandler: collabHandler,
	})

	// 注册 PQL 查询语言路由
	httpapi.RegisterPQLRoutes(engine, &httpapi.Deps{
		Auth:            authSvc,
		PrincipalParser: parsePrincipal,
		WorkspaceStore:  wsStore,
		RBACStore:       rbacStore,
		PQLHandler:      pqlHandler,
	})

	// 注册 AI Agent 路由
	httpapi.RegisterAIAgentRoutes(engine, &httpapi.Deps{
		Auth:               authSvc,
		PrincipalParser:    parsePrincipal,
		WorkspaceStore:     wsStore,
		RBACStore:          rbacStore,
		AIAgentHandler:     aiAgentHandler,
		GodModeHandler:     godModeHandler,
		GrantHandler:       grantHandler,
		ImporterHandler:    importerHandler,
		MarketplaceHandler: marketplaceHandler,
		CRDTHub:            crdtHub,
		DLPHandler:         dlpHandler,
	})

	// 注册 MCP Server 路由（AI 客户端通过标准 MCP 协议接入）
	mcpHandler := mcpserver.NewHandler(fmt.Sprintf("http://127.0.0.1:%d", cfg.Server.Port), func(raw string) (int64, bool) {
		p, err := apiTokenSvc.ResolvePrincipal(context.Background(), raw)
		if err != nil {
			return 0, false
		}
		return p.UserID, true
	})
	mcpHandler.Register(engine)

	// 注册项目状态与更新路由
	httpapi.RegisterProjectStateRoutes(engine, &httpapi.Deps{
		Auth:                authSvc,
		PrincipalParser:     parsePrincipal,
		WorkspaceStore:      wsStore,
		RBACStore:           rbacStore,
		ProjectStateHandler: projectStateHandler,
	})

	// 注册 Projects-as-Code 路由（YAML 定义项目）
	httpapi.RegisterProjectCodeRoutes(engine, &httpapi.Deps{
		Auth:               authSvc,
		PrincipalParser:    parsePrincipal,
		WorkspaceStore:     wsStore,
		RBACStore:          rbacStore,
		ProjectCodeHandler: projectCodeHandler,
	})

	// 注册客户管理路由（工作空间级：读 customer:read / 写 customer:manage）
	httpapi.RegisterCustomerRoutes(engine, &httpapi.Deps{
		Auth:            authSvc,
		PrincipalParser: parsePrincipal,
		WorkspaceStore:  wsStore,
		RBACStore:       rbacStore,
		CustomerHandler: customerHandler,
	})

	// 注册自定义工作项类型路由（项目级）
	httpapi.RegisterWorkItemTypesRoutes(engine, &httpapi.Deps{
		Auth:                authSvc,
		PrincipalParser:     parsePrincipal,
		WorkspaceStore:      wsStore,
		RBACStore:           rbacStore,
		WorkItemTypeHandler: workItemTypeHandler,
	})

	// 注册页面评论路由（工作空间级：读 workspace:read / 写 comment:moderate）
	httpapi.RegisterPageCommentRoutes(engine, &httpapi.Deps{
		Auth:               authSvc,
		PrincipalParser:    parsePrincipal,
		WorkspaceStore:     wsStore,
		RBACStore:          rbacStore,
		PageCommentHandler: pageCommentHandler,
	})

	// 注册重复工作项规则路由（项目级）
	httpapi.RegisterRecurringRoutes(engine, &httpapi.Deps{
		Auth:             authSvc,
		PrincipalParser:  parsePrincipal,
		WorkspaceStore:   wsStore,
		RBACStore:        rbacStore,
		RecurringHandler: recurringHandler,
	})

	// 注册里程碑路由（项目级）
	httpapi.RegisterMilestoneRoutes(engine, &httpapi.Deps{
		Auth:             authSvc,
		PrincipalParser:  parsePrincipal,
		WorkspaceStore:   wsStore,
		RBACStore:        rbacStore,
		MilestoneHandler: milestoneHandler,
	})

	// 注册工作空间视图路由（工作空间级）
	httpapi.RegisterWorkspaceViewRoutes(engine, &httpapi.Deps{
		Auth:                 authSvc,
		PrincipalParser:      parsePrincipal,
		WorkspaceStore:       wsStore,
		RBACStore:            rbacStore,
		WorkspaceViewHandler: workspaceViewHandler,
	})

	// 注册集成管理路由（工作空间级：读 workspace:read / 写 workspace:update）
	httpapi.RegisterIntegrationRoutes(engine, &httpapi.Deps{
		Auth:               authSvc,
		PrincipalParser:    parsePrincipal,
		WorkspaceStore:     wsStore,
		RBACStore:          rbacStore,
		IntegrationHandler: integrationHandler,
	})

	// 注册 OAuth 应用与授权路由（应用管理 + OAuth2 端点）
	httpapi.RegisterOAuthAppRoutes(engine, &httpapi.Deps{
		Auth:            authSvc,
		PrincipalParser: parsePrincipal,
		WorkspaceStore:  wsStore,
		RBACStore:       rbacStore,
		OAuthAppHandler: oauthAppHandler,
		OAuthAppSvc:     oauthAppSvc,
	})

	errCh := make(chan error, 1)
	go func() {
		log.Info("api listening", zap.Int("port", cfg.Server.Port), zap.String("env", cfg.Server.Env))
		errCh <- srv.ListenAndServe()
	}()

	// pprof 独立端口（默认 6060），仅当 YDSZ_PPROF_ENABLE=true 时启动。
	// 独立端口避免性能分析端点与业务端口混用，生产环境默认关闭。
	if enablePprof(os.Getenv("YDSZ_PPROF_ENABLE")) {
		pprofPort := 6060
		if v := os.Getenv("YDSZ_PPROF_PORT"); v != "" {
			if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
				pprofPort = parsed
			}
		}
		go func() {
			pprofAddr := ":" + strconv.Itoa(pprofPort)
			log.Info("pprof listening", zap.String("addr", pprofAddr),
				zap.String("endpoints", "/debug/pprof/"))
			if err := http.ListenAndServe(pprofAddr, nil); err != nil {
				log.Error("pprof server error", zap.Error(err))
			}
		}()
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigCh:
		log.Info("shutdown signal", zap.String("signal", sig.String()))
	case err := <-errCh:
		if !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	return srv.Shutdown(shutdownCtx)
}

// enablePprof 判断字符串 v 是否为"启用"语义（true / 1 / yes / on）。
func enablePprof(v string) bool {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "true", "1", "yes", "on":
		return true
	default:
		return false
	}
}
