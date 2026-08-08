// +build ignore

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
)

func main() {
	conn, err := pgx.Connect(context.Background(), "postgres://postgres:Limw1020@127.0.0.1:5432/ydsz-plane?sslmode=disable")
	if err != nil {
		fmt.Fprintln(os.Stderr, "connect:", err)
		os.Exit(1)
	}
	defer conn.Close(context.Background())

	rows, err := conn.Query(context.Background(), `
		SELECT id, workspace_id, name, slug, identifier, description, network, icon, color, cover_image_url, template, status, sort_order, modules, created_by, created_at, updated_at
		FROM projects WHERE workspace_id = $1 AND deleted_at IS NULL
		ORDER BY sort_order, created_at ASC`, 3)
	if err != nil {
		fmt.Fprintln(os.Stderr, "query:", err)
		os.Exit(1)
	}
	defer rows.Close()

	for rows.Next() {
		var id, workspaceID, createdBy int64
		var name, slug, identifier, description, network, template, status string
		var modules []byte
		var createdAt, updatedAt json.RawMessage
		var sortOrder float64
		var icon, color, coverImageUrl *string

		if err := rows.Scan(&id, &workspaceID, &name, &slug, &identifier, &description,
			&network, &icon, &color, &coverImageUrl, &template, &status, &sortOrder, &modules, &createdBy, &createdAt, &updatedAt); err != nil {
			fmt.Fprintln(os.Stderr, "scan error:", err)
			os.Exit(1)
		}
		fmt.Printf("Row: id=%d, ws=%d, name=%s, slug=%s, sort_order=%f, modules=%s, created_by=%d\n",
			id, workspaceID, name, slug, sortOrder, string(modules), createdBy)
	}
	if err := rows.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "rows err:", err)
	}
}
