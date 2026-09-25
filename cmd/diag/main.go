package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Simulate the exact service.Create flow
func main() {
	db, err := pgxpool.New(context.Background(), "postgres://postgres:Limw1020@127.0.0.1:5432/ydsz-plane?sslmode=disable")
	if err != nil {
		fmt.Println("connect error:", err)
		os.Exit(1)
	}
	defer db.Close()

	// Ensure sequences exist
	_, _ = db.Exec(context.Background(), `CREATE SEQUENCE IF NOT EXISTS workspaces_id_seq START 1 INCREMENT 1`)
	_, _ = db.Exec(context.Background(), `CREATE SEQUENCE IF NOT EXISTS workspace_members_id_seq START 1 INCREMENT 1`)
	_, _ = db.Exec(context.Background(), `SELECT setval('workspaces_id_seq', (SELECT COALESCE(MAX(id), 0) + 1 FROM workspaces), false)`)
	_, _ = db.Exec(context.Background(), `SELECT setval('workspace_members_id_seq', (SELECT COALESCE(MAX(id), 0) + 1 FROM workspace_members), false)`)

	// Simulate workspace normalizelike service.normalizeSlug
	name := "test-sim-1111"
	slug := strings.ToLower(strings.ReplaceAll(name, " ", "-"))

	// Transaction
	var wsID, memberID, memberCount int64
	var wsName, wsSlug, wsRole string
	err = pgx.BeginTxFunc(context.Background(), db, pgx.TxOptions{}, func(tx pgx.Tx) error {
		if err := tx.QueryRow(context.Background(), "SELECT nextval('workspaces_id_seq')").Scan(&wsID); err != nil {
			return fmt.Errorf("nextval workspace: %w", err)
		}
		var w struct {
			id       int64
			name     string
			slug     string
			logoURL  string
			timezone string
			language string
			status   string
			ownerID  int64
		}
		err := tx.QueryRow(context.Background(), `
			INSERT INTO workspaces (id, name, slug, timezone, language, owner_id)
			VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING id, name, slug, coalesce(logo_url,''), timezone, language, status, owner_id`,
			wsID, name, slug, "Asia/Shanghai", "", 1).
			Scan(&w.id, &w.name, &w.slug, &w.logoURL, &w.timezone, &w.language, &w.status, &w.ownerID)
		if err != nil {
			return fmt.Errorf("insert workspace: %w", err)
		}
		if err := tx.QueryRow(context.Background(), "SELECT nextval('workspace_members_id_seq')").Scan(&memberID); err != nil {
			return fmt.Errorf("nextval member: %w", err)
		}
		if _, err := tx.Exec(context.Background(), `
			INSERT INTO workspace_members (id, workspace_id, user_id, role, joined_at)
			VALUES ($1, $2, $3, 'owner', now())`,
			memberID, w.id, 1); err != nil {
			return fmt.Errorf("insert member: %w", err)
		}
		wsID = w.id
		wsName = w.name
		wsSlug = w.slug
		wsRole = "owner"
		memberCount = 1
		return nil
	})
	fmt.Printf("=== Create result: id=%d name=%s slug=%s role=%s members=%d err=%v ===\n",
		wsID, wsName, wsSlug, wsRole, memberCount, err)

	// Cleanup
	if wsID > 0 {
		db.Exec(context.Background(), "DELETE FROM workspace_members WHERE workspace_id = $1", wsID)
		db.Exec(context.Background(), "DELETE FROM workspaces WHERE id = $1", wsID)
		fmt.Println("=== cleanup done ===")
	}
}
