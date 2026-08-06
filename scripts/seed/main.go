// Command seed inserts minimal development seed data:
// one admin user and one demo workspace (idempotent).
// Demo data (projects/issues/...) lands with the corresponding Sprints.
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/ydszopen/ydsz-plane/internal/application/auth"
	"github.com/ydszopen/ydsz-plane/internal/config"
	"github.com/ydszopen/ydsz-plane/internal/infrastructure/persistence"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "seed:", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	ctx := context.Background()

	pool, err := persistence.NewPool(ctx, cfg.Database.URL, cfg.Database.MaxConns)
	if err != nil {
		return err
	}
	defer pool.Close()

	authSvc := auth.NewService(pool.Pool, cfg.Auth.JWTSecret, cfg.Auth.JWTIssuer,
		cfg.Auth.AccessTokenTTL, cfg.Auth.RefreshTokenTTL, cfg.Auth.BcryptCost, true)

	hash, err := authSvc.HashPassword("Admin@123")
	if err != nil {
		return err
	}

	var userID int64
	err = pool.QueryRow(ctx, `
		INSERT INTO users (email, password_hash, display_name)
		VALUES ('admin@ydsz.dev', $1, '系统管理员')
		ON CONFLICT (email) DO UPDATE SET email = EXCLUDED.email
		RETURNING id`, hash).Scan(&userID)
	if err != nil {
		return fmt.Errorf("seed user: %w", err)
	}

	var wsID int64
	err = pool.QueryRow(ctx, `
		INSERT INTO workspaces (name, slug, owner_id)
		VALUES ('演示工作空间', 'demo', $1)
		ON CONFLICT (slug) WHERE status <> 'archived' DO NOTHING
		RETURNING id`, userID).Scan(&wsID)
	if err != nil {
		if err.Error() == "no rows in result set" {
			if err := pool.QueryRow(ctx, `SELECT id FROM workspaces WHERE slug = 'demo'`).Scan(&wsID); err != nil {
				return fmt.Errorf("seed workspace lookup: %w", err)
			}
		} else {
			return fmt.Errorf("seed workspace: %w", err)
		}
	}

	if _, err := pool.Exec(ctx, `
		INSERT INTO workspace_members (workspace_id, user_id, role)
		VALUES ($1, $2, 'owner')
		ON CONFLICT DO NOTHING`, wsID, userID); err != nil {
		return fmt.Errorf("seed member: %w", err)
	}

	fmt.Println("seed ok: admin@ydsz.dev / Admin@123, workspace: demo")
	return nil
}
