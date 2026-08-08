// Package searchv1 — SearchService gRPC API 类型桩代码（手工 stub）。
//
// 本文件为 Phase-0 开发期手工编写的轻量编译占位桩。
// 正式代码由 `buf generate` 从 search.proto 自动生成并覆盖本文件。
//
// 前置依赖（需在线安装）：
//   go get google.golang.org/grpc@v1.78.0
//   随后运行 buf generate 生成正式代码
package searchv1

import (
	context "context"
)

// ============== 领域消息 ==============

type SearchHit struct {
	Index  string
	Id     string
	Score  float32
	Fields map[string]string
}

type SearchQuery struct {
	WorkspaceId int64
	ProjectId   int64
	Q           string
	IndexFilter []string
	Jql         string
	OrderBy     string
	Descending  bool
	Limit       int32
	Offset      int32
}

type SearchResult struct {
	Hits   []*SearchHit
	Total  int64
	TookMs int64
}

type IndexRequest struct {
	Index string
	Id    string
	Doc   []byte
}

type BulkIndexRequest struct {
	Items []*IndexRequest
}

type BulkIndexResp struct {
	Indexed int64
	Failed  int64
}

type DeleteFromIndexRequest struct {
	Index string
	Id    string
}

type ReindexAllResp struct {
	TotalScanned int64
	TotalIndexed int64
	DurationMs   int64
}

type Empty struct{}

// ============== gRPC Server 接口 ==============

type SearchServiceServer interface {
	Search(context.Context, *SearchQuery) (*SearchResult, error)
	Index(context.Context, *IndexRequest) (*Empty, error)
	BulkIndex(context.Context, *BulkIndexRequest) (*BulkIndexResp, error)
	DeleteFromIndex(context.Context, *DeleteFromIndexRequest) (*Empty, error)
	ReindexAll(context.Context, *SearchQuery) (*ReindexAllResp, error)
	MustEmbedUnimplementedSearchServiceServer()
}

type UnimplementedSearchServiceServer struct{}

func (UnimplementedSearchServiceServer) Search(context.Context, *SearchQuery) (*SearchResult, error) {
	return nil, nil
}
func (UnimplementedSearchServiceServer) Index(context.Context, *IndexRequest) (*Empty, error)        { return nil, nil }
func (UnimplementedSearchServiceServer) BulkIndex(context.Context, *BulkIndexRequest) (*BulkIndexResp, error) {
	return nil, nil
}
func (UnimplementedSearchServiceServer) DeleteFromIndex(context.Context, *DeleteFromIndexRequest) (*Empty, error) {
	return nil, nil
}
func (UnimplementedSearchServiceServer) ReindexAll(context.Context, *SearchQuery) (*ReindexAllResp, error) {
	return nil, nil
}
func (UnimplementedSearchServiceServer) MustEmbedUnimplementedSearchServiceServer() {}

// RegisterSearchServiceServer 注册 gRPC 服务实现。
func RegisterSearchServiceServer(s interface{ RegisterService(desc interface{}, impl interface{}) }, srv SearchServiceServer) {
	// stub: 真实实现由 buf generate 生成后替换
	_ = srv
}

// ============== gRPC Client 接口 ==============

type SearchServiceClient interface {
	Search(context.Context, *SearchQuery) (*SearchResult, error)
	Index(context.Context, *IndexRequest) (*Empty, error)
	BulkIndex(context.Context, *BulkIndexRequest) (*BulkIndexResp, error)
	DeleteFromIndex(context.Context, *DeleteFromIndexRequest) (*Empty, error)
	ReindexAll(context.Context, *SearchQuery) (*ReindexAllResp, error)
}

type searchServiceClient struct {
	cc interface{}
}

func NewSearchServiceClient(cc interface{}) SearchServiceClient {
	return &searchServiceClient{cc: cc}
}

func (c *searchServiceClient) Search(ctx context.Context, in *SearchQuery) (*SearchResult, error) {
	return nil, nil
}
func (c *searchServiceClient) Index(ctx context.Context, in *IndexRequest) (*Empty, error) {
	return nil, nil
}
func (c *searchServiceClient) BulkIndex(ctx context.Context, in *BulkIndexRequest) (*BulkIndexResp, error) {
	return nil, nil
}
func (c *searchServiceClient) DeleteFromIndex(ctx context.Context, in *DeleteFromIndexRequest) (*Empty, error) {
	return nil, nil
}
func (c *searchServiceClient) ReindexAll(ctx context.Context, in *SearchQuery) (*ReindexAllResp, error) {
	return nil, nil
}
