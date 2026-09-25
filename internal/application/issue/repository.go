// Package issue — Repository 接口定义（S16 P0-2）。
//
// 引入 Repository 接口层的目的：
//  1. 单元测试可 Mock → 领域层覆盖率可从当前基线提升至 ≥70%
//  2. BatchUpdateV2 批量 SQL 的封装复用
//  3. 为将来拆微服务准备接口边界
//
// 所有 Service 结构体已隐式实现这些接口（Go 的 structural typing）。
// 未来计划：添加 impl 结构体将 DB 访问从 Service 中分离，
// 测试时注入 fake/mock 实现。
package issue

import (
	"context"

	"github.com/jackc/pgx/v5"
)

// === Issue / Workitem 仓储接口 ===

// WorkitemReader 读取工作项的接口。
type WorkitemReader interface {
	// GetByID 按工作项 ID 查询（跨 task/requirement/defect）。
	GetByID(ctx context.Context, wsID, issueID int64) (*Issue, error)
	// List 跨类型列表查询。
	List(ctx context.Context, opts ListIssuesOptions) ([]Issue, int64, error)
}

// WorkitemBatchOperator 批量操作的接口。
// 现有 *Service 结构体已隐式实现此接口。
type WorkitemBatchOperator interface {
	// BatchUpdateV2 批量更新（优先级/分配人）— 按 type_code 分桶后批量 SQL。
	BatchUpdateV2(ctx context.Context, wsID, projectID, userID int64, in BatchUpdateInput) (BatchResult, error)
}

// WorkitemWriter 写入单个工作项的接口。
type WorkitemWriter interface {
	// Create 创建单个工作项（按 type 分派到 Task/Requirement/Defect 服务）。
	Create(ctx context.Context, in CreateIssueInput) (*Issue, error)
	// Update 更新单个工作项。
	Update(ctx context.Context, wsID, issueID int64, in UpdateIssueInput) (*Issue, error)
	// SoftDelete 软删除。
	SoftDelete(ctx context.Context, wsID, issueID int64) error
	// Restore 从回收站恢复。
	Restore(ctx context.Context, wsID, issueID int64) error
	// Transition 执行状态流转。
	Transition(ctx context.Context, wsID, projectID, issueID, toStateID, userID int64) (*Issue, error)
}

// TxWorkitemQuery 在已提供的事务内查询工作项。
// 供 BatchUpdateV2 分桶查询使用。
type TxWorkitemQuery interface {
	// GroupByType 按 type_code 分桶查询指定 IDs 的类型和版本信息。
	GroupByType(ctx context.Context, tx pgx.Tx, wsID int64, ids []int64) (map[string][]workitemTypeRow, error)
}

// Compile-time 接口实现检查（确保 *Service 结构体满足接口约束）。
var (
	_ WorkitemReader        = (*Service)(nil)
	_ WorkitemWriter        = (*Service)(nil)
	_ WorkitemBatchOperator = (*Service)(nil)
	_ TxWorkitemQuery       = (*Service)(nil)
)

// GroupByType 是 Service 对 TxWorkitemQuery 接口的实现。
// 代理到包级 batchGroupByType 函数。
func (s *Service) GroupByType(ctx context.Context, tx pgx.Tx, wsID int64, ids []int64) (map[string][]workitemTypeRow, error) {
	return batchGroupByType(ctx, tx, wsID, ids)
}
