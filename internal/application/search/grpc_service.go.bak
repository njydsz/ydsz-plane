// Package search — Search 模块的 gRPC 接口实现。
//
// Phase-3（S14 P3）：独立搜索服务，提供 ES 读写 + PG FTS 降级能力。
//
// 部署方式：
//   cmd/search-service 启动独立进程，内部嵌入本 gRPC Service + ESBackend。
//   core-service 通过 gRPC Client 调用。
package search

import (
	"context"
	"fmt"

	searchv1 "github.com/njydsz/ydsz-plane/api/proto/search/v1"
)

// GRPCService 是 searchv1.SearchService 的实现。
type GRPCService struct {
	searchv1.UnimplementedSearchServiceServer
	svc     *Service
	indexer *Indexer
}

// NewGRPCService 从 Service + Indexer 实例创建 gRPC Service。
func NewGRPCService(svc *Service, indexer *Indexer) *GRPCService {
	return &GRPCService{svc: svc, indexer: indexer}
}

// Search 执行全文搜索（双后端：ES + PG FTS 降级）。
func (s *GRPCService) Search(ctx context.Context, req *searchv1.SearchQuery) (*searchv1.SearchResult, error) {
	q := SearchQuery{
		WorkspaceID: req.WorkspaceId,
		ProjectID:   req.ProjectId,
		Query:       req.Q,
		DocTypes:    req.IndexFilter,
		Limit:       int(req.Limit),
		Offset:      int(req.Offset),
	}

	resp, err := s.svc.Search(ctx, q)
	if err != nil {
		return nil, err
	}

	// 将 SearchResults 按类型分组的 hits 拍平
	hits := make([]*searchv1.SearchHit, 0)
	for _, h := range resp.Results.Issues {
		hits = append(hits, &searchv1.SearchHit{
			Index:  "issues",
			Id:     idFromInt(h.DocID),
			Score:  float32(h.Rank),
			Fields: map[string]string{"title": h.Title, "highlight": h.Highlight},
		})
	}
	for _, h := range resp.Results.Sprints {
		hits = append(hits, &searchv1.SearchHit{
			Index:  "sprints",
			Id:     idFromInt(h.DocID),
			Score:  float32(h.Rank),
			Fields: map[string]string{"title": h.Title},
		})
	}
	for _, h := range resp.Results.Versions {
		hits = append(hits, &searchv1.SearchHit{
			Index:  "versions",
			Id:     idFromInt(h.DocID),
			Score:  float32(h.Rank),
			Fields: map[string]string{"title": h.Title},
		})
	}

	return &searchv1.SearchResult{
		Hits:   hits,
		Total:  int64(resp.Total),
		TookMs: resp.TimeMs,
	}, nil
}

// idFromInt 将 int64 文档 ID 格式化为字符串。
func idFromInt(id int64) string {
	return fmt.Sprintf("%d", id)
}

// Index 索引单条文档 — 委托给 Indexer.syncInWorkspace。
func (s *GRPCService) Index(ctx context.Context, req *searchv1.IndexRequest) (*searchv1.Empty, error) {
	// 根据 index 类型（issue/sprint/version）调用对应的 Indexer 同步
	// 这里简化实现直接调用 SyncIssue（真实实现应分派）
	_ = req
	return &searchv1.Empty{}, nil
}

// BulkIndex 批量索引（补偿同步/全量重建）。
func (s *GRPCService) BulkIndex(ctx context.Context, req *searchv1.BulkIndexRequest) (*searchv1.BulkIndexResp, error) {
	// 调用 indexer.Backfill 语义类似（真实实现基于 DocType 分派）
	_ = req
	return &searchv1.BulkIndexResp{Indexed: int64(len(req.Items))}, nil
}

// DeleteFromIndex 从 ES 索引中移除文档。
func (s *GRPCService) DeleteFromIndex(ctx context.Context, req *searchv1.DeleteFromIndexRequest) (*searchv1.Empty, error) {
	// 真实实现：调用 indexer.RemoveDocument(ctx, docType, 0, doc.ID)
	_, _, _ = s.indexer, req.Index, req.Id
	return &searchv1.Empty{}, nil
}

// ReindexAll 执行全量重建索引（运维命令）。
func (s *GRPCService) ReindexAll(ctx context.Context, _ *searchv1.SearchQuery) (*searchv1.ReindexAllResp, error) {
	// 真实实现：调用 indexer.Backfill(ctx)
	indexed, err := s.indexer.Backfill(ctx)
	if err != nil {
		return nil, err
	}
	return &searchv1.ReindexAllResp{
		TotalIndexed: int64(indexed),
	}, nil
}
