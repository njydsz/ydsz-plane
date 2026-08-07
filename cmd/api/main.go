// Command api 启动 Ydsz Plane 的 HTTP API 服务。
package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/njydsz/ydsz-plane/internal/application/apitoken"
	"github.com/njydsz/ydsz-plane/internal/application/attachment"
	"github.com/njydsz/ydsz-plane/internal/application/auth"
	"github.com/njydsz/ydsz-plane/internal/application/dashboard"
	"github.com/njydsz/ydsz-plane/internal/application/issue"
	notif "github.com/njydsz/ydsz-plane/internal/application/notification"
	"github.com/njydsz/ydsz-plane/internal/application/preference"
	"github.com/njydsz/ydsz-plane/internal/application/search"
	"github.com/njydsz/ydsz-plane/internal/application/sprint"
	"github.com/njydsz/ydsz-plane/internal/application/version"
	"github.com/njydsz/ydsz-plane/internal/application/workbench"
	"github.com/njydsz/ydsz-plane/internal/application/workspace"
	"github.com/njydsz/ydsz-plane/internal/config"
	"github.com/njydsz/ydsz-plane/internal/infrastructure/cache"
	"github.com/njydsz/ydsz-plane/internal/infrastructure/mail"
	"github.com/njydsz/ydsz-plane/internal/infrastructure/persistence"
	"github.com/njydsz/ydsz-plane/internal/infrastructure/storage"
	"github.com/njydsz/ydsz-plane/internal/infrastructure/telemetry"
	"github.com/njydsz/ydsz-plane/internal/infrastructure/ws"
	httpapi "github.com/njydsz/ydsz-plane/internal/interfaces/http"
)

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

	rdb, err := cache.NewClient(ctx, cfg.Redis.Addr, cfg.Redis.Password, cfg.Redis.DB)
	if err != nil {
		return err
	}
	defer func() { _ = rdb.Close() }()

	// ---------- Services ----------
	authSvc := auth.NewService(pool.Pool, cfg.Auth.JWTSecret, cfg.Auth.JWTIssuer,
		cfg.Auth.AccessTokenTTL, cfg.Auth.RefreshTokenTTL, cfg.Auth.BcryptCost,
		cfg.Features.RegistrationOpen)
	resetSvc := auth.NewPasswordResetService(pool.Pool, nil, cfg.Email.AppBaseURL, cfg.Auth.BcryptCost)

	wsStore := auth.NewWorkspaceMembershipStore(pool.Pool)
	wsSvc := workspace.NewService(pool.Pool)
	memberSvc := workspace.NewMemberService(pool.Pool)
	projectSvc := workspace.NewProjectService(pool.Pool)
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
	wsHub := ws.NewHub(rdb)
	go wsHub.Run()
	defer wsHub.Shutdown()

	// ---------- Issue domain ----------
	issueSvc := issue.NewService(pool.Pool)
	stateSvc := issue.NewStateService(pool.Pool)
	activitySvc := issue.NewActivityService(pool.Pool)
	timeLogSvc := issue.NewTimeLogService(pool.Pool)
	relationSvc := issue.NewRelationService(pool.Pool)
	commentSvc := issue.NewCommentService(pool.Pool)
	issueHandler := issue.NewIssueHandler(&issue.HandlerDeps{
		IssueSvc:        issueSvc,
		StateSvc:        stateSvc,
		ActivitySvc:     activitySvc,
		TimeLogSvc:      timeLogSvc,
		RelationSvc:     relationSvc,
		CommentSvc:      commentSvc,
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

	// ---------- Search domain ----------
	searchSvc := search.NewService(pool.Pool)
	searchHandler := search.NewSearchHandler(&search.HandlerDeps{
		SearchSvc:      searchSvc,
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

	// ---------- Attachment / Storage ----------
	stClient, err := storage.New(cfg.Storage)
	if err != nil {
		return fmt.Errorf("storage: %w", err)
	}
	attSvc := attachment.NewService(pool.Pool, stClient)
	attHandler := attachment.NewHandler(&attachment.HandlerDeps{AttachmentSvc: attSvc})

	// ---------- View Preference ----------
	prefSvc := preference.NewService(pool.Pool)
	prefHandler := preference.NewHandler(prefSvc)

	// ---------- HTTP Engine ----------
	engine := httpapi.NewEngine(&httpapi.Deps{
		Cfg:             cfg,
		Log:             log,
		DB:              pool.Pool,
		Redis:           rdb,
		Auth:            authSvc,
		ResetSvc:        resetSvc,
		PrincipalParser: parsePrincipal,
		ApiTokenSvc:     apiTokenSvc,
		WorkspaceStore:  wsStore,
		WorkspaceSvc:    wsSvc,
		MemberSvc:       memberSvc,
		InvitationSvc:   invitationSvc,
		ProjectSvc:      projectSvc,
		AuditSvc:        auditSvc,
		Mail:            mailSvc,
		// Issue domain
		IssueSvc:     issueSvc,
		StateSvc:     stateSvc,
		ActivitySvc:  activitySvc,
		TimeLogSvc:   timeLogSvc,
		IssueHandler: issueHandler,
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
		// WebSocket Hub
		WSHub: wsHub,
		// Sprint domain
		SprintHandler: sprintHandler,
		// Version domain
		VersionHandler: versionHandler,
	})

	// 注册工作项路由（必须在 NewEngine 之后）
	httpapi.RegisterIssueRoutes(engine, &httpapi.Deps{
		Auth:            authSvc,
		PrincipalParser: parsePrincipal,
		WorkspaceStore:  wsStore,
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

	// 注册视图偏好路由（项目级）
	httpapi.RegisterPreferenceRoutes(engine, &httpapi.Deps{
		Auth:            authSvc,
		PrincipalParser: parsePrincipal,
		WorkspaceStore:  wsStore,
		PrefHandler:     prefHandler,
	})

	// 注册搜索路由（项目级 + 工作空间级）
	httpapi.RegisterSearchRoutes(engine, &httpapi.Deps{
		Auth:            authSvc,
		PrincipalParser: parsePrincipal,
		WorkspaceStore:  wsStore,
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
		DashboardHandler: dashboardHandler,
	})

	// 注册通知路由
	httpapi.RegisterNotificationRoutes(engine, &httpapi.Deps{
		Auth:                authSvc,
		PrincipalParser:     parsePrincipal,
		WorkspaceStore:      wsStore,
		NotificationHandler: notifHandler,
	})

	// 注册附件路由
	httpapi.RegisterAttachmentRoutes(engine, &httpapi.Deps{
		Auth:              authSvc,
		PrincipalParser:   parsePrincipal,
		WorkspaceStore:    wsStore,
		AttachmentHandler: attHandler,
	})

	// 注册 WebSocket 路由
	httpapi.RegisterWSRoutes(engine, &httpapi.Deps{
		Auth:            authSvc,
		PrincipalParser: parsePrincipal,
		WorkspaceStore:  wsStore,
		WSHub:           wsHub,
	})

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Server.Port),
		Handler: engine,
		// ReadHeaderTimeout 限制客户端发送请求头的时长。10s 对弱网客户端足够宽裕，
		// 同时能快速拒绝 Slowloris 类慢速攻击。
		ReadHeaderTimeout: 10 * time.Second,
		// ReadTimeout 覆盖读取完整请求体的时长。30s 可支撑单请求内的批量 CSV/JSON 上传。
		ReadTimeout: 30 * time.Second,
		// WriteTimeout 限制整个响应写入时长。60s 对分页工作项列表与迭代报告足够；
		// 需要更长时间的端点应当改用流式输出。
		WriteTimeout: 60 * time.Second,
		// IdleTimeout 回收静默超过 2 分钟的 keep-alive 连接，
		// 在高流量下限制空闲连接占用，同时保留典型突发请求的 keep-alive 收益。
		IdleTimeout: 120 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Info("api listening", zap.Int("port", cfg.Server.Port), zap.String("env", cfg.Server.Env))
		errCh <- srv.ListenAndServe()
	}()

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
