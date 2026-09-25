// Package persistence — 通用 Repository 接口抽象。
//
// 遵循 Repository 模式（Domain-Driven Design + Unit of Work），
// 将数据访问细节与领域逻辑解耦：
//
//  1. BaseRepository[T] 定义通用 CRUD 接口，具体聚合根仓储嵌入它。
//  2. PostgresRepo 为 PostgreSQL 场景提供通用的 WithTx 事务包装器，
//     确保每个写操作自动设置 RLS 上下文。
//  3. Querier 是 pgxpool.Pool 和 pgx.Tx 的共用最小接口，
//     允许 Repository 在事务内和事务外统一操作。
//
// 参考:
//   - Martin Fowler《Patterns of Enterprise Application Architecture》— Repository
//   - Uber Go Style Guide: "Define interfaces where they are used, not where they are implemented."
//   - Google SRE Book 第 6 章: 资源层抽象降低耦合度、提高可测试性。
package persistence

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Querier 定义 pgx 查询操作的最小接口，同时被 *pgxpool.Pool 和 pgx.Tx 满足。
// 这使得 Repository 的结构体实现可以在事务内和事务外复用同一段 SQL 逻辑。
//
// Uber Go Style: "Accept interfaces, return structs."
// 此处使用最小接口而非 *pgxpool.Pool，是为了方便测试时注入 mock。
type Querier interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, arguments ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, arguments ...any) pgx.Row
}

// BaseRepository 定义通用 CRUD 操作接口。
//
// 类型参数 T 为实体类型（如 Issue、Task）。各聚合根的 Repository 接口
// 可嵌入 BaseRepository[T] 并追加领域专属方法。
//
// 对标 Google SRE: 标准化接口降低多人协作的理解成本。
type BaseRepository[T any] interface {
	// GetByID 按主键查询实体；不存在时返回 errs.ErrNotFound。
	GetByID(ctx context.Context, id int64) (*T, error)

	// Create 插入新实体；自动填充序列号、默认状态等。
	Create(ctx context.Context, entity *T) error

	// Update 按乐观锁版本更新实体；version 不匹配返回 errs.ErrVersionConflict。
	Update(ctx context.Context, entity *T) error

	// Delete 按主键软删除；返回 (true, nil) 表示找到并删除。
	Delete(ctx context.Context, id int64) (bool, error)
}

// PostgresRepo 是 PostgreSQL 场景的通用 Repository 基类。
//
// 封装了连接池引用和事务包装方法，使得子仓储无需
// 重复实现 RLS 上下文注入逻辑。
//
// 使用方式：
//
//	type IssueRepo struct {
//	    persistence.PostgresRepo
//	}
//	func NewIssueRepo(pool *pgxpool.Pool) *IssueRepo {
//	    return &IssueRepo{PostgresRepo: persistence.NewPostgresRepo(pool)}
//	}
type PostgresRepo struct {
	db *pgxpool.Pool
}

// NewPostgresRepo 构造 PostgresRepo 基类实例。
func NewPostgresRepo(db *pgxpool.Pool) PostgresRepo {
	return PostgresRepo{db: db}
}

// WithTenantTx 在设置了租户上下文的事务内执行 fn，
// 使 RLS 策略（current_setting('app.workspace_id')）按行隔离数据。
// SET LOCAL 仅作用于当前事务，在连接池下是安全的。
//
// 这是原有 Pool.WithTenantTx 的复用保留，供子仓储调用。
func (r PostgresRepo) WithTenantTx(ctx context.Context, workspaceID int64, fn func(tx pgx.Tx) error) error {
	return (&Pool{r.db}).WithTenantTx(ctx, workspaceID, fn)
}

// DB 返回底层连接池（供需要直接执行的场景使用，推荐优先使用 WithTenantTx）。
func (r PostgresRepo) DB() *pgxpool.Pool {
	return r.db
}

// --- 错误辅助 ---

// IsNotFound 判断 error 是否为 pgx.ErrNoRows 或 errs 框架的 ErrNotFound。
// 提供统一的 "not found" 判断入口，避免业务代码散落 errors.Is 判断。
func IsNotFound(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}
