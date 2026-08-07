// Package search — 搜索索引器（Indexer / SyncService）。
//
// 负责将 issues / sprints / versions 三类对象维护进 `search_documents` 统一索引表，
// 供 Service.Search 全文检索使用。
//
// 背景（P0 缺陷修复）:
//   - 0008 迁移只为 issues 建立了 DB 触发器自动填充，sprints / versions 从不会被索引，
//     且存量数据无回填，导致搜索 sprints/versions 永远为空、迁移前的历史 issues 无法命中。
//   - 本包提供: 全量回填（Backfill）+ 单对象同步（Sync*）+ 软删除清理（RemoveDocument），
//     并通过 RunConsumer 订阅领域事件、worker 消费 `search.index` 任务两条链路驱动。
//
// RLS 处理:
//   `search_documents`（以及 issues/sprints/versions）启用了 FORCE ROW LEVEL SECURITY，
//   写入/读取都必须先在同一事务内设置 `app.workspace_id`（SET LOCAL），
//   否则 WITH CHECK / USING 会因 `current_setting('app.workspace_id', true)` 为 NULL 而拒绝。
//   本实现复用 automation.appendActivity 的事务内 set_config 模式：
//   `SELECT set_config('app.workspace_id', $1, true)` 再执行 upsert。
package search

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"

	"github.com/njydsz/ydsz-plane/internal/infrastructure/mq"
)

// batchSize 是 Backfill 每批处理的记录条数。
const batchSize = 500

// Indexer 维护 search_documents 索引表。
// 持有一个 pgx 连接池，所有写操作均在设置了租户上下文的事务内执行。
type Indexer struct {
	db *pgxpool.Pool
}

// NewIndexer 创建索引器。
func NewIndexer(db *pgxpool.Pool) *Indexer {
	return &Indexer{db: db}
}

// --- Backfill 全量回填 ---

// Backfill 将现有所有未删除的 issues/sprints/versions 全量 upsert 进 search_documents。
// 按 (workspace_id, doc_type, doc_id) 幂等（INSERT ... ON CONFLICT DO UPDATE），
// 分批处理，返回本次回填的对象总数。
func (x *Indexer) Backfill(ctx context.Context) (int, error) {
	total := 0
	for _, typ := range []DocType{DocTypeIssue, DocTypeSprint, DocTypeVersion} {
		wsIDs, err := x.workspacesForType(ctx, typ)
		if err != nil {
			return total, fmt.Errorf("backfill: list workspaces for %s: %w", typ, err)
		}
		for _, wsID := range wsIDs {
			n, err := x.backfillTypeInWorkspace(ctx, typ, wsID)
			if err != nil {
				return total, fmt.Errorf("backfill: %s ws=%d: %w", typ, wsID, err)
			}
			total += n
		}
	}
	return total, nil
}

// workspacesForType 返回某类型对象存在的所有 workspace_id。
func (x *Indexer) workspacesForType(ctx context.Context, typ DocType) ([]int64, error) {
	var table string
	switch typ {
	case DocTypeIssue:
		table = "issues"
	case DocTypeSprint:
		table = "sprints"
	case DocTypeVersion:
		table = "versions"
	default:
		return nil, fmt.Errorf("unknown doc_type %q", typ)
	}
	q := "SELECT DISTINCT workspace_id FROM " + table + " WHERE deleted_at IS NULL ORDER BY workspace_id"
	rows, err := x.db.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []int64
	for rows.Next() {
		var ws int64
		if err := rows.Scan(&ws); err != nil {
			return nil, err
		}
		out = append(out, ws)
	}
	return out, rows.Err()
}

// backfillTypeInWorkspace 在单个 workspace 内按 id 分批回填指定类型。
func (x *Indexer) backfillTypeInWorkspace(ctx context.Context, typ DocType, wsID int64) (int, error) {
	count := 0
	lastID := int64(0)
	for {
		n, err := x.backfillBatch(ctx, typ, wsID, lastID)
		if err != nil {
			return count, err
		}
		count += n
		if n < batchSize {
			return count, nil
		}
		lastID += batchSize // 键集分页：id 连续自增，推进游标
	}
}

// backfillBatch 回填一批（≤ batchSize 条），从 lastID 之后按 id 升序取数。
func (x *Indexer) backfillBatch(ctx context.Context, typ DocType, wsID, lastID int64) (int, error) {
	table := sourceTable(typ)

	tx, err := x.db.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, "SELECT set_config('app.workspace_id', $1, true)", strconv.FormatInt(wsID, 10)); err != nil {
		return 0, err
	}

	// 以 id 键集分页，在设置了租户上下文的事务内 SELECT + INSERT 同批完成。
	sql := "INSERT INTO search_documents (workspace_id, project_id, doc_type, doc_id, title, identifier, content, search_tsv, metadata) " +
		"SELECT " + sourceColumnsFor(typ) + " FROM " + table + " src" +
		" WHERE src.workspace_id = $1 AND src.deleted_at IS NULL AND src.id > $2 ORDER BY src.id LIMIT " + strconv.Itoa(batchSize) +
		" ON CONFLICT (workspace_id, doc_type, doc_id) DO UPDATE SET " +
		"title = EXCLUDED.title, identifier = EXCLUDED.identifier, content = EXCLUDED.content, " +
		"search_tsv = EXCLUDED.search_tsv, metadata = EXCLUDED.metadata, updated_at = now()"

	tag, err := tx.Exec(ctx, sql, wsID, lastID)
	if err != nil {
		return 0, err
	}
	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}
	return int(tag.RowsAffected()), nil
}

// --- 单对象同步 ---

// SyncIssue 同步单条 issue 进 search_documents（upsert）。
func (x *Indexer) SyncIssue(ctx context.Context, issueID int64) error {
	wsID, err := x.resolveWorkspace(ctx, "issues", issueID)
	if err != nil {
		return err
	}
	return x.syncInWorkspace(ctx, wsID, DocTypeIssue, issueID)
}

// SyncSprint 同步单条 sprint 进 search_documents（upsert）。
func (x *Indexer) SyncSprint(ctx context.Context, sprintID int64) error {
	wsID, err := x.resolveWorkspace(ctx, "sprints", sprintID)
	if err != nil {
		return err
	}
	return x.syncInWorkspace(ctx, wsID, DocTypeSprint, sprintID)
}

// SyncVersion 同步单条 version 进 search_documents（upsert）。
func (x *Indexer) SyncVersion(ctx context.Context, versionID int64) error {
	wsID, err := x.resolveWorkspace(ctx, "versions", versionID)
	if err != nil {
		return err
	}
	return x.syncInWorkspace(ctx, wsID, DocTypeVersion, versionID)
}

// syncInWorkspace 在指定 workspace 事务内 upsert 单条对象。
// docID 对应 source 表的 id（即 doc_id 列）。
func (x *Indexer) syncInWorkspace(ctx context.Context, wsID int64, typ DocType, docID int64) error {
	table := sourceTable(typ)
	tx, err := x.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, "SELECT set_config('app.workspace_id', $1, true)", strconv.FormatInt(wsID, 10)); err != nil {
		return err
	}

	sql := "INSERT INTO search_documents (workspace_id, project_id, doc_type, doc_id, title, identifier, content, search_tsv, metadata) " +
		"SELECT " + sourceColumnsFor(typ) + " FROM " + table + " src" +
		" WHERE src.workspace_id = $1 AND src.deleted_at IS NULL AND src.id = $2" +
		" ON CONFLICT (workspace_id, doc_type, doc_id) DO UPDATE SET " +
		"title = EXCLUDED.title, identifier = EXCLUDED.identifier, content = EXCLUDED.content, " +
		"search_tsv = EXCLUDED.search_tsv, metadata = EXCLUDED.metadata, updated_at = now()"

	if _, err := tx.Exec(ctx, sql, wsID, docID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// RemoveDocument 软删除时清理索引文档。
func (x *Indexer) RemoveDocument(ctx context.Context, docType string, wsID, docID int64) error {
	if _, ok := validDocType(docType); !ok {
		return fmt.Errorf("remove: unknown doc_type %q", docType)
	}
	tx, err := x.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, "SELECT set_config('app.workspace_id', $1, true)", strconv.FormatInt(wsID, 10)); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx,
		"DELETE FROM search_documents WHERE workspace_id = $1 AND doc_type = $2 AND doc_id = $3",
		wsID, docType, docID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// resolveWorkspace 从 source 表解析对象所属 workspace_id。
// 注意：在未设置 app.workspace_id 的会话中，FORCE RLS 会使该查询返回空，
// 此时视为对象不存在（返回 pgx.ErrNoRows）。生产链路（RunConsumer / worker）
// 会优先通过事件/任务的 workspace_id 直接走 syncInWorkspace，避免此限制。
func (x *Indexer) resolveWorkspace(ctx context.Context, table string, id int64) (int64, error) {
	var ws int64
	err := x.db.QueryRow(ctx,
		"SELECT workspace_id FROM "+table+" WHERE id = $1 AND deleted_at IS NULL", id).Scan(&ws)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, fmt.Errorf("resolve workspace: %s %d: %w", table, id, err)
		}
		return 0, err
	}
	return ws, nil
}

// --- SQL 片段构建 ---

func sourceTable(typ DocType) string {
	switch typ {
	case DocTypeIssue:
		return "issues"
	case DocTypeSprint:
		return "sprints"
	case DocTypeVersion:
		return "versions"
	}
	return ""
}

// sourceColumnsFor 返回 source 表向 search_documents 投影的列表达式。
// 注意表别名统一为 "src"，配合 sourceFromClause 的 "FROM <table> src" 使用。
func sourceColumnsFor(typ DocType) string {
	switch typ {
	case DocTypeIssue:
		// workspace_id, project_id, doc_type, doc_id, title, identifier, content, search_tsv, metadata
		return "src.workspace_id, src.project_id, 'issue', src.id, src.name, src.sequence_id::text, src.description_stripped, " +
			"to_tsvector('simple', coalesce(src.name, '') || ' ' || coalesce(src.description_stripped, '')), " +
			"jsonb_build_object('type_code', src.type_code, 'state_id', src.state_id, 'priority', src.priority)"
	case DocTypeSprint:
		return "src.workspace_id, src.project_id, 'sprint', src.id, src.name, NULL, coalesce(src.goal, ''), " +
			"to_tsvector('simple', coalesce(src.name, '') || ' ' || coalesce(src.goal, '')), " +
			"jsonb_build_object('status', src.status)"
	case DocTypeVersion:
		return "src.workspace_id, src.project_id, 'version', src.id, src.name, src.semver, coalesce(src.description, ''), " +
			"to_tsvector('simple', coalesce(src.name, '') || ' ' || coalesce(src.description, '')), " +
			"jsonb_build_object('status', src.status)"
	}
	return ""
}

func validDocType(dt string) (DocType, bool) {
	switch DocType(dt) {
	case DocTypeIssue, DocTypeSprint, DocTypeVersion:
		return DocType(dt), true
	}
	return "", false
}

// --- 领域事件消费者 ---

// consumerQueue 是索引消费者绑定的队列名。
const consumerQueue = "search.index.events"

// routingPattern 订阅所有领域事件（内部按事件类型过滤出 issue/sprint/version）。
const routingPattern = "plane.events.#"

// RunConsumer 启动阻塞型索引事件消费者循环。
// 订阅 EventExchange 上 issue.*/sprint.*/version.* 事件，解析后调用 Sync*/RemoveDocument。
// 当 ctx 取消时优雅退出；连接断开后指数退避自动重连。
//
// 调用方应在独立 goroutine 中运行：
//
//	go search.RunConsumer(ctx, mqClient, indexer, log)
func RunConsumer(ctx context.Context, mqClient *mq.Client, indexer *Indexer, log *zap.Logger) {
	if log == nil {
		log = zap.NewNop()
	}
	log.Info("search consumer: starting",
		zap.String("queue", consumerQueue),
		zap.String("exchange", mq.EventExchange))

	for {
		select {
		case <-ctx.Done():
			log.Info("search consumer: stopped")
			return
		default:
		}

		if err := runConsumeLoop(ctx, mqClient, indexer, log); err != nil {
			if errors.Is(err, context.Canceled) || ctx.Err() != nil {
				log.Info("search consumer: stopped (context)")
				return
			}
			log.Warn("search consumer: connection lost, retrying",
				zap.Error(err), zap.Duration("backoff", 2*time.Second))
			select {
			case <-ctx.Done():
				return
			case <-time.After(2 * time.Second):
			}
		}
	}
}

// runConsumeLoop 单次消费循环：声明队列 → 消费。
func runConsumeLoop(ctx context.Context, mqClient *mq.Client, indexer *Indexer, log *zap.Logger) error {
	if _, err := mqClient.DeclareQueue(ctx, consumerQueue, mq.EventExchange, routingPattern, amqp.Table{
		"x-max-priority":         int64(5),
		"x-dead-letter-exchange": mq.DeadLetterExchange,
	}); err != nil {
		return errors.New("search: declare queue: " + err.Error())
	}

	return mqClient.Consume(ctx, consumerQueue, "search-index-consumer", false, func(delivery amqp.Delivery) error {
		var event mq.EventEnvelope
		if err := json.Unmarshal(delivery.Body, &event); err != nil {
			log.Warn("search consumer: bad payload, skipping", zap.Error(err))
			return nil // 无法解析直接 ACK，避免死信循环
		}
		if err := handleEvent(ctx, indexer, event); err != nil {
			log.Warn("search consumer: handle failed",
				zap.String("event_type", event.EventType),
				zap.Error(err))
			return err // NACK → 重试（受 MaxRetries 限制）
		}
		return nil // ACK
	})
}

// handleEvent 将领域事件映射为索引操作。
// 事件类型前缀决定 doc_type，`.deleted` 后缀走删除，其余走 upsert 同步。
func handleEvent(ctx context.Context, indexer *Indexer, event mq.EventEnvelope) error {
	typ, ok := validDocType(eventDocType(event.EventType))
	if !ok {
		return nil // 非 issue/sprint/version 事件，静默 ACK
	}

	// 删除事件（issue.deleted / sprint.deleted / version.deleted）
	if strings.HasSuffix(event.EventType, ".deleted") {
		return indexer.RemoveDocument(ctx, string(typ), event.WorkspaceID, event.AggregateID)
	}

	// 其余事件：通过事件自带 workspace_id 直接同步，避免 RLS 解析限制
	return indexer.syncInWorkspace(ctx, event.WorkspaceID, typ, event.AggregateID)
}

// eventDocType 从事件类型提取 doc_type（如 "issue.created" → "issue"）。
func eventDocType(eventType string) string {
	if i := strings.IndexByte(eventType, '.'); i > 0 {
		return eventType[:i]
	}
	return ""
}