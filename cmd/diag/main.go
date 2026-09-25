package main

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	db, err := pgxpool.New(context.Background(), "postgres://postgres:Limw1020@127.0.0.1:5432/ydsz-plane?sslmode=disable")
	if err != nil {
		fmt.Println("connect error:", err)
		os.Exit(1)
	}
	defer db.Close()

	// 1. Create sequences
	_, _ = db.Exec(context.Background(), `CREATE SEQUENCE IF NOT EXISTS workspaces_id_seq START 1 INCREMENT 1`)
	_, _ = db.Exec(context.Background(), `CREATE SEQUENCE IF NOT EXISTS workspace_members_id_seq START 1 INCREMENT 1`)

	// 2. Set sequences to max(id)+1 to avoid collisions
	_, err = db.Exec(context.Background(), `SELECT setval('workspaces_id_seq', (SELECT COALESCE(MAX(id), 0) + 1 FROM workspaces), false)`)
	fmt.Printf("=== set worksequences start value: err=%v ===\n", err)

	_, err = db.Exec(context.Background(), `SELECT setval('workspace_members_id_seq', (SELECT COALESCE(MAX(id), 0) + 1 FROM workspace_members), false)`)
	fmt.Printf("=== set workspace_members_id_seq start value: err=%v ===\n", err)

	// 3. Verify sequences
	rows, _ := db.Query(context.Background(), "SELECT relname, last_value FROM pg_class WHERE relkind='S' AND relname LIKE '%workspace%'")
	fmt.Println("=== sequences (with current value) ===")
	for rows.Next() {
		var name string
		var lastVal int64
		rows.Scan(&name, &lastVal)
		fmt.Printf("  %s (last_value=%d)\n", name, lastVal)
	}
	rows.Close()

	// 4. Full test insert
	var wsID int64
	err = db.QueryRow(context.Background(), `
		INSERT INTO workspaces (id, name, slug, owner_id)
		VALUES (nextval('workspaces_id_seq'), 'test-full-1111', 'test-full-1111', 1)
		RETURNING id
	`).Scan(&wsID)
	fmt.Printf("=== insert workspace: id=%d, err=%v ===\n", wsID, err)

	if err == nil {
		_, err = db.Exec(context.Background(), `
			INSERT INTO workspace_members (id, workspace_id, user_id, role)
			VALUES (nextval('workspace_members_id_seq'), $1, 1, 'owner')
		`, wsID)
		fmt.Printf("=== insert workspace_member: err=%v ===\n", err)
	}

	// 5. Cleanup test data
	if wsID > 0 {
		db.Exec(context.Background(), "DELETE FROM workspace_members WHERE workspace_id = $1", wsID)
		db.Exec(context.Background(), "DELETE FROM workspaces WHERE id = $1", wsID)
	}

	var seqCount, wsCount int
	db.QueryRow(context.Background(), "SELECT count(*) FROM pg_class WHERE relkind='S' AND relname LIKE 'workspaces_id_seq'").Scan(&seqCount)
	db.QueryRow(context.Background(), "SELECT count(*) FROM workspaces WHERE slug LIKE 'test-full%'").Scan(&wsCount)
	fmt.Printf("=== final: sequences=%d, leftover_test_rows=%d ===\n", seqCount, wsCount)
}
