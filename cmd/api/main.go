// Command api runs the Ydsz Plane HTTP API server.
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

	"github.com/njydsz/ydsz-plane/internal/application/auth"
	"github.com/njydsz/ydsz-plane/internal/application/issue"
	"github.com/njydsz/ydsz-plane/internal/application/sprint"
	"github.com/njydsz/ydsz-plane/internal/application/version"
	"github.com/njydsz/ydsz-plane/internal/application/workspace"
	"github.com/njydsz/ydsz-plane/internal/config"
	"github.com/njydsz/ydsz-plane/internal/infrastructure/cache"
	"github.com/njydsz/ydsz-plane/internal/infrastructure/mail"
	"github.com/njydsz/ydsz-plane/internal/infrastructure/persistence"
	"github.com/njydsz/ydsz-plane/internal/infrastructure/telemetry"
	httpapi "github.com/njydsz/ydsz-plane/internal/interfaces/http"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "api: fatal:", err)
		os.Exit(1)
	}
}

// run bootstraps the HTTP API server and blocks until shutdown.
//
// Initialization follows a strict ordering to guarantee that every
// downstream component receives a fully-initialized dependency:
//
//  1. Configuration +Logger — everything else depends on the parsed
//     config and a working logger.
//  2. PostgreSQL pool — services will validate their DB schema on boot.
//  3. Redis client — used for sessions, rate-limiting and the outbox sink.
//  4. Domain Services — wired after both DB and Redis are available so that
//     each service can fail fast if its underlying store is unreachable.
//  5. HTTP Engine + Routes — mounted last once all handlers are registered.
//  6. Listen & Serve — the server starts in its own goroutine; the main
//     goroutine then blocks on a select{} that waits for either a signal or
//     a ListenAndServe error.
//
// On SIGINT/SIGTERM the server performs a graceful shutdown with a 15 s
// drain timeout, after which remaining connections are closed and the
// process exits.
//
// A non-nil return from run terminates the process with exit code 1.
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

	// Security notice: dev-mode JWT secret rotates each restart — tokens issued
	// before restart become invalid. This is intentional to avoid hardcoded
	// secrets in the codebase.
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

	// ---------- Issue domain ----------
	issueSvc := issue.NewService(pool.Pool)
	stateSvc := issue.NewStateService(pool.Pool)
	activitySvc := issue.NewActivityService(pool.Pool)
	timeLogSvc := issue.NewTimeLogService(pool.Pool)
	issueHandler := issue.NewIssueHandler(&issue.HandlerDeps{
		IssueSvc:       issueSvc,
		StateSvc:       stateSvc,
		ActivitySvc:    activitySvc,
		TimeLogSvc:     timeLogSvc,
		WorkspaceStore: wsStore,
	})

	// ---------- Sprint domain ----------
	sprintSvc := sprint.NewService(pool.Pool)
	sprintHandler := sprint.NewHandler(sprintSvc)

	// ---------- Version domain ----------
	versionSvc := version.NewService(pool.Pool)
	versionHandler := version.NewHandler(versionSvc)

	// ---------- HTTP Engine ----------
	engine := httpapi.NewEngine(&httpapi.Deps{
		Cfg:            cfg,
		Log:            log,
		DB:             pool.Pool,
		Redis:          rdb,
		Auth:           authSvc,
		ResetSvc:       resetSvc,
		WorkspaceStore: wsStore,
		WorkspaceSvc:   wsSvc,
		MemberSvc:      memberSvc,
		InvitationSvc:  invitationSvc,
		ProjectSvc:     projectSvc,
		AuditSvc:       auditSvc,
		Mail:           mailSvc,
		// Issue domain
		IssueSvc:     issueSvc,
		StateSvc:     stateSvc,
		ActivitySvc:  activitySvc,
		TimeLogSvc:   timeLogSvc,
		IssueHandler: issueHandler,
		// Sprint domain
		SprintHandler: sprintHandler,
		// Version domain
		VersionHandler: versionHandler,
	})

	// 注册工作项路由（必须在 NewEngine 之后）
	httpapi.RegisterIssueRoutes(engine, &httpapi.Deps{
		Auth:           authSvc,
		WorkspaceStore: wsStore,
		IssueHandler:   issueHandler,
		SprintHandler:  sprintHandler,
		VersionHandler: versionHandler,
	})

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Server.Port),
		Handler: engine,
		// ReadHeaderTimeout bounds the time a client has to finish sending
		// request headers. 10 s is generous enough for slow clients on
		// cellular but rejects Slowloris-style attacks quickly.
		ReadHeaderTimeout: 10 * time.Second,
		// ReadTimeout covers reading the full request body. 30 s handles
		// bulk CSV/JSON uploads within a single request window.
		ReadTimeout: 30 * time.Second,
		// WriteTimeout limits the entire response write. 60 s is sufficient
		// for paginated issue lists and sprint reports — endpoints that need
		// more time should stream instead.
		WriteTimeout: 60 * time.Second,
		// IdleTimeout recycles keep-alive connections that have been silent
		// for 2 minutes. This limits idle slot consumption under high traffic
		// while still benefiting from HTTP keep-alive for typical bursts.
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
