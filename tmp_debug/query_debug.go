package main

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, "postgres://postgres:Limw1020@127.0.0.1:5432/ydsz-plane?sslmode=disable")
	if err != nil {
		fmt.Fprintf(os.Stderr, "pool error: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()

	// Test 1: Simple count
	var count int
	err = pool.QueryRow(ctx, "SELECT count(*) FROM issues WHERE project_id = $1 AND workspace_id = $2 AND deleted_at IS NULL", int64(1), int64(3)).Scan(&count)
	if err != nil {
		fmt.Printf("ERROR count: %v\n", err)
	} else {
		fmt.Printf("Count issues (project 1, ws 3): %d\n", count)
	}

	// Test 2: The full query from listIssues
	rows, err := pool.Query(ctx, `
		SELECT i.id, i.public_id, i.workspace_id, i.project_id, i.sequence_id,
		       i.type_code, i.parent_id, i.depth, i.name,
		       i.state_id, s.name, s.color, s."group",
		       i.priority, i.severity, i.category, i.point,
		       i.start_date, i.target_date, i.progress, i.version,
		       i.created_by, i.created_at, i.updated_at,
		       p.identifier
		FROM issues i
		JOIN states s ON s.id = i.state_id
		JOIN projects p ON p.id = i.project_id
		WHERE i.deleted_at IS NULL AND i.project_id = $1 AND i.workspace_id = $2
		ORDER BY i.updated_at DESC
		LIMIT $3 OFFSET $4`, int64(1), int64(3), int64(50), int64(0))
	if err != nil {
		fmt.Printf("ERROR query: %v\n", err)
	} else {
		rows.Close()
		fmt.Println("Query OK")
	}

	// Test 3: Simple join test
	var identifier string
	err = pool.QueryRow(ctx, `
		SELECT p.identifier FROM projects p WHERE p.id = $1`, int64(1)).Scan(&identifier)
	if err != nil {
		fmt.Printf("ERROR project query: %v\n", err)
	} else {
		fmt.Printf("Project identifier (id=1): %s\n", identifier)
	}

	// Test 4: Does table exists?
	var exists bool
	err = pool.QueryRow(ctx, "SELECT EXISTS(SELECT FROM information_schema.tables WHERE table_name='issues')").Scan(&exists)
	if err != nil {
		fmt.Printf("ERROR exists query: %v\n", err)
	} else {
		fmt.Printf("Issues table exists: %v\n", exists)
	}
}
