// Package cache — 桌面端本地 SQLite 离线缓存。
//
// 目的：
//   1. 保存最近访问的 工作项 / 项目元数据，网络断开时可继续浏览
//   2. 缓存搜索结果（减少 API 请求）
//   3. 暂存本地新建的工作项草稿（离线写后上线自动提交）
//
// WAL 模式 + 延迟写入 → 读写互不阻塞，适合高并发 UI 渲染。
//
// 表结构：
//   projects (id, name, key, workspace_id, ... )
//   issues (id, project_id, title, status, priority, assignee_id, updated_at, ... )
//   issue_drafts (id, project_id, title, body, created_at, synced bool )
//   search_keywords (keyword, last_used)
package cache

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	_ "github.com/glebarez/go-sqlite"
)

const (
	dbFilename = "ydsz-plane-cache.db"
	schemaV1   = `
CREATE TABLE IF NOT EXISTS projects (
	id            INTEGER PRIMARY KEY,
	name          TEXT    NOT NULL,
	key           TEXT    NOT NULL,
	workspace_id  INTEGER NOT NULL,
	description   TEXT    DEFAULT '',
	is_archived   INTEGER DEFAULT 0,
	updated_at    INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS issues (
	id            INTEGER PRIMARY KEY,
	project_id    INTEGER NOT NULL,
	title         TEXT    NOT NULL,
	status        TEXT    NOT NULL DEFAULT 'open',
	priority      INTEGER DEFAULT 0,
	assignee_id   INTEGER DEFAULT 0,
	updated_at    INTEGER NOT NULL,
	synced        INTEGER DEFAULT 1,
	body          TEXT    DEFAULT ''
);

CREATE INDEX IF NOT EXISTS idx_issues_project ON issues(project_id);
CREATE INDEX IF NOT EXISTS idx_issues_updated ON issues(updated_at);

CREATE TABLE IF NOT EXISTS issue_drafts (
	id            INTEGER PRIMARY KEY AUTOINCREMENT,
	project_id    INTEGER NOT NULL,
	title         TEXT    NOT NULL,
	body          TEXT    NOT NULL DEFAULT '',
	created_at    INTEGER NOT NULL,
	synced        INTEGER DEFAULT 0
);

CREATE TABLE IF NOT EXISTS search_keywords (
	keyword       TEXT PRIMARY KEY,
	last_used     INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS meta (
	key   TEXT PRIMARY KEY,
	value TEXT
);`
)

// Cache 封装 SQLite 缓存的 CRUD 接口。
type Cache struct {
	db       *sql.DB
	mu       sync.Mutex
	dbPath   string
}

// NewCache 打开（或创建）本地 SQLite 数据库。
func NewCache() (*Cache, error) {
	dir, err := getCacheDir()
	if err != nil {
		return nil, fmt.Errorf("resolve cache dir: %w", err)
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, fmt.Errorf("mkdir: %w", err)
	}

	dbPath := filepath.Join(dir, dbFilename)
	db, err := sql.Open("sqlite", dbPath+"?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)")
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	if _, err := db.Exec(schemaV1); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("init schema: %w", err)
	}

	return &Cache{db: db, dbPath: dbPath}, nil
}

// Close 关闭数据库。
func (c *Cache) Close() error {
	return c.db.Close()
}

// SaveProject 写入或更新项目元数据。
func (c *Cache) SaveProject(ctx context.Context, id int64, name, key string, workspaceID int64, updatedAt time.Time) error {
	_, err := c.db.ExecContext(ctx, `
		INSERT INTO projects(id, name, key, workspace_id, updated_at) VALUES(?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET name=excluded.name, key=excluded.key, updated_at=excluded.updated_at`,
		id, name, key, workspaceID, updatedAt.Unix())
	return err
}

// ListProjects 返回缓存的所有项目。
func (c *Cache) ListProjects(ctx context.Context, workspaceID int64) ([]map[string]interface{}, error) {
	rows, err := c.db.QueryContext(ctx,
		"SELECT id, name, key FROM projects WHERE workspace_id=? ORDER BY name", workspaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []map[string]interface{}
	for rows.Next() {
		var id int64
		var name, key string
		if err := rows.Scan(&id, &name, &key); err != nil {
			return nil, err
		}
		projects = append(projects, map[string]interface{}{
			"id": id, "name": name, "key": key,
		})
	}
	return projects, rows.Err()
}

// UpsertIssue 写入或更新工作项缓存。
func (c *Cache) UpsertIssue(ctx context.Context, id, projectID int64, title, status string, updatedAt time.Time) error {
	_, err := c.db.ExecContext(ctx, `
		INSERT INTO issues(id, project_id, title, status, updated_at) VALUES(?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET title=excluded.title, status=excluded.status, updated_at=excluded.updated_at`,
		id, projectID, title, status, updatedAt.Unix())
	return err
}

// ListIssuesByProject 返回项目下的缓存工作项。
func (c *Cache) ListIssuesByProject(ctx context.Context, projectID int64) ([]map[string]interface{}, error) {
	rows, err := c.db.QueryContext(ctx,
		"SELECT id, title, status, updated_at FROM issues WHERE project_id=? ORDER BY updated_at DESC LIMIT 200", projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var issues []map[string]interface{}
	for rows.Next() {
		var id int64
		var title, status string
		var updatedAt int64
		if err := rows.Scan(&id, &title, &status, &updatedAt); err != nil {
			return nil, err
		}
		issues = append(issues, map[string]interface{}{
			"id": id, "title": title, "status": status, "updated_at": updatedAt,
		})
	}
	return issues, rows.Err()
}

// InsertDraft 保存离线草稿。
func (c *Cache) InsertDraft(ctx context.Context, projectID int64, title, body string) (int64, error) {
	res, err := c.db.ExecContext(ctx,
		"INSERT INTO issue_drafts(project_id, title, body, created_at) VALUES(?, ?, ?, ?)",
		projectID, title, body, time.Now().Unix())
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// ListUnsyncedDrafts 返回尚未同步的离线草稿。
func (c *Cache) ListUnsyncedDrafts(ctx context.Context) ([]map[string]interface{}, error) {
	rows, err := c.db.QueryContext(ctx,
		"SELECT id, project_id, title, body, created_at FROM issue_drafts WHERE synced=0")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var drafts []map[string]interface{}
	for rows.Next() {
		var id, projectID, createdAt int64
		var title, body string
		if err := rows.Scan(&id, &projectID, &title, &body, &createdAt); err != nil {
			return nil, err
		}
		drafts = append(drafts, map[string]interface{}{
			"id": id, "project_id": projectID, "title": title, "body": body, "created_at": createdAt,
		})
	}
	return drafts, rows.Err()
}

// MarkDraftSynced 将草稿标记为已同步。
func (c *Cache) MarkDraftSynced(ctx context.Context, id int64) error {
	_, err := c.db.ExecContext(ctx, "UPDATE issue_drafts SET synced=1 WHERE id=?", id)
	return err
}

// Vacuum 压缩数据库（清除已删除数据占用的空间）。
func (c *Cache) Vacuum(ctx context.Context) error {
	_, err := c.db.ExecContext(ctx, "VACUUM")
	return err
}

// HealthCheck 返回缓存状态调试信息。
func (c *Cache) HealthCheck() string {
	return fmt.Sprintf("cache=sqlite path=%s", c.dbPath)
}

// getCacheDir 返回本地缓存目录（跨平台的用户数据目录）。
func getCacheDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".ydsz-plane", "cache"), nil
}
