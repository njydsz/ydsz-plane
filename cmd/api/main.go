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
	})

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:           engine,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
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
