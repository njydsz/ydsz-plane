// Package es 提供 Elasticsearch 客户端封装。
//
// 设计要点:
//   - 支持 ES 8.x，使用官方 go-elasticsearch 客户端
//   - 连接池配置（对标 Google SRE Book Ch.17：最大空闲连接 50，每个主机最多 10）
//   - 健康检查 + 自动降级标记
//   - 零停机索引别名切换
//   - 批量索引 (Bulk API)
//   - Worker Pool 并行 Reindex（对标 Google SRE Book Ch.17：面向吞吐的重构）
//
// 环境变量:
//   - YDSZ_ES_URLS: ES 集群地址，逗号分隔（默认 http://127.0.0.1:9200）
//   - YDSZ_ES_USERNAME: 认证用户名（可选）
//   - YDSZ_ES_PASSWORD: 认证密码（可选）
//   - YDSZ_ES_INDEX_PREFIX: 索引名前缀（默认 ydsz）
package es

import (
	"bytes"
	"context"
	"crypto/tls"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.uber.org/zap"

	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// Config ES 客户端配置。
type Config struct {
	// URLs ES 集群地址列表。
	URLs []string

	// Username 认证用户名。
	Username string

	// Password 认证密码。
	Password string

	// IndexPrefix 索引名前缀。
	IndexPrefix string

	// MaxIdleConns 最大空闲连接数。
	MaxIdleConns int

	// MaxConnsPerHost 每个主机最大连接数。
	MaxConnsPerHost int

	// Timeout 请求超时。
	Timeout time.Duration

	// RetryOnStatus 重试的 HTTP 状态码列表。
	RetryOnStatus []int

	// MaxRetries 最大重试次数。
	MaxRetries int
}

// DefaultConfig 返回推荐的默认配置。
func DefaultConfig() Config {
	return Config{
		URLs:            []string{"http://127.0.0.1:9200"},
		IndexPrefix:     "ydsz",
		MaxIdleConns:    50,
		MaxConnsPerHost: 10,
		Timeout:         5 * time.Second,
		RetryOnStatus:   []int{502, 503, 504},
		MaxRetries:      2,
	}
}

// --- Prometheus Metrics (对标 Google SRE Book Ch.17: 负载监控) ---

var (
	// ReindexRequestsTotal 统计 ES reindex 请求总数（按 doc_type 分组）。
	ReindexRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace:   "plane",
			Name:        "es_reindex_requests_total",
			Help:        "Total ES reindex bulk requests.",
		},
		[]string{"doc_type"},
	)

	// ReindexErrorsTotal 统计 ES reindex 失败次数。
	ReindexErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace:   "plane",
			Name:        "es_reindex_errors_total",
			Help:        "Total ES reindex bulk failures.",
		},
		[]string{"doc_type"},
	)

	// ReindexProgress 记录当前 reindex 进度（用于监控看板）。
	ReindexProgress = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace:   "plane",
			Name:        "es_reindex_progress_ratio",
			Help:        "ES reindex progress (0.0 to 1.0) with total/processed labels.",
		},
	[]string{"doc_type"},
)

// Client ES 客户端封装。
type Client struct {
	cfg    Config
	http   *http.Client
	logger *zap.Logger
	mu     sync.RWMutex
	closed bool

	// 健康状态（原子操作，零锁争用）
	healthy atomic.Bool

	// ReindexWorkers worker pool 并行度（默认 min(NumCPU, 8)）。
	ReindexWorkers int

	// ReindexBatchSize 每个 bulk 请求的批次大小（默认 1000）。
	ReindexBatchSize int
}

// NewClient 创建 ES 客户端。
func NewClient(cfg Config) (*Client, error) {
	if len(cfg.URLs) == 0 {
		return nil, fmt.Errorf("es: at least one URL is required")
	}
	if cfg.IndexPrefix == "" {
		cfg.IndexPrefix = "ydsz"
	}
	if cfg.MaxIdleConns <= 0 {
		cfg.MaxIdleConns = 50
	}
	if cfg.MaxConnsPerHost <= 0 {
		cfg.MaxConnsPerHost = 10
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 5 * time.Second
	}

	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:        cfg.MaxIdleConns,
		MaxConnsPerHost:     cfg.MaxConnsPerHost,
		MaxIdleConnsPerHost: cfg.MaxIdleConns / 2,
		IdleConnTimeout:     90 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: false,
		},
	}

	// worker 并行度：默认 min(NumCPU, 8)，上限保护避免过多连接争用
	workers := runtime.NumCPU()
	if workers > 8 {
		workers = 8
	}
	if workers < 1 {
		workers = 4
	}

	c := &Client{
		cfg:              cfg,
		http:             &http.Client{Transport: transport, Timeout: cfg.Timeout},
		logger:           zap.NewNop(),
		ReindexWorkers:   workers,
		ReindexBatchSize: 1000,
	}

	// 启动时标记为健康，由定期健康检查维护
	c.healthy.Store(true)

	return c, nil
}

// IsHealthy 返回 ES 集群当前是否健康。
func (c *Client) IsHealthy() bool {
	return c.healthy.Load()
}

// HealthCheck 执行 ES 集群健康检查。
func (c *Client) HealthCheck(ctx context.Context) error {
	resp, err := c.doRequest(ctx, "GET", "/_cluster/health?timeout=3s", nil)
	if err != nil {
		c.healthy.Store(false)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 500 {
		c.healthy.Store(false)
		return fmt.Errorf("es: cluster health returned %d", resp.StatusCode)
	}

	// 解析健康状态
	var health struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&health); err != nil {
		return err
	}

	if health.Status == "red" {
		c.healthy.Store(false)
		return fmt.Errorf("es: cluster health is RED")
	}

	c.healthy.Store(true)
	return nil
}

// --- Index Management ---

// IndexName 返回带前缀的索引名。
func (c *Client) IndexName(name string) string {
	return c.cfg.IndexPrefix + "_" + name
}

// CreateIndex 创建索引（如果不存在）。
func (c *Client) CreateIndex(ctx context.Context, name string, mapping map[string]any) error {
	idxName := c.IndexName(name)
	body, _ := json.Marshal(mapping)

	resp, err := c.doRequest(ctx, "PUT", "/"+idxName, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("es: create index %s: %w", idxName, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("es: create index %s failed (status=%d): %s", idxName, resp.StatusCode, string(respBody))
	}
	return nil
}

// IndexExists 检查索引是否存在。
func (c *Client) IndexExists(ctx context.Context, name string) (bool, error) {
	idxName := c.IndexName(name)
	resp, err := c.doRequest(ctx, "HEAD", "/"+idxName, nil)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	return resp.StatusCode == 200, nil
}

// DeleteIndex 删除索引。
func (c *Client) DeleteIndex(ctx context.Context, name string) error {
	idxName := c.IndexName(name)
	resp, err := c.doRequest(ctx, "DELETE", "/"+idxName, nil)
	if err != nil {
		return fmt.Errorf("es: delete index %s: %w", idxName, err)
	}
	defer resp.Body.Close()
	return nil
}

// AliasExists 检查别名是否存在。
func (c *Client) AliasExists(ctx context.Context, alias string) (bool, error) {
	resp, err := c.doRequest(ctx, "HEAD", "/_alias/"+alias, nil)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	return resp.StatusCode == 200, nil
}

// SwapAlias 原子切换别名（零停机索引重建）。
func (c *Client) SwapAlias(ctx context.Context, alias, oldIndex, newIndex string) error {
	body := map[string]any{
		"actions": []map[string]any{
			{
				"remove": map[string]any{
					"index": oldIndex,
					"alias": alias,
				},
			},
			{
				"add": map[string]any{
					"index": newIndex,
					"alias": alias,
				},
			},
		},
	}

	jsonBody, _ := json.Marshal(body)
	resp, err := c.doRequest(ctx, "POST", "/_aliases", bytes.NewReader(jsonBody))
	if err != nil {
		return fmt.Errorf("es: swap alias %s: %w", alias, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("es: swap alias %s failed (status=%d): %s", alias, resp.StatusCode, string(respBody))
	}
	return nil
}

// --- Document Operations ---

// IndexDocument 索引单个文档。
func (c *Client) IndexDocument(ctx context.Context, index, docID string, doc any) error {
	idxName := c.IndexName(index)
	body, _ := json.Marshal(doc)

	var path string
	if docID != "" {
		path = fmt.Sprintf("/%s/_doc/%s", idxName, docID)
	} else {
		path = fmt.Sprintf("/%s/_doc", idxName)
	}

	resp, err := c.doRequest(ctx, "PUT", path, bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("es: index document (status=%d): %s", resp.StatusCode, string(respBody))
	}
	return nil
}

// DeleteDocument 删除单个文档。
func (c *Client) DeleteDocument(ctx context.Context, index, docID string) error {
	idxName := c.IndexName(index)
	resp, err := c.doRequest(ctx, "DELETE", fmt.Sprintf("/%s/_doc/%s", idxName, docID), nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return nil // 幂等
	}
	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("es: delete document (status=%d): %s", resp.StatusCode, string(respBody))
	}
	return nil
}

// BulkIndex 批量索引文档。
// 返回成功/失败计数。
func (c *Client) BulkIndex(ctx context.Context, index string, docs []BulkDoc) (success, failed int, err error) {
	if len(docs) == 0 {
		return 0, 0, nil
	}

	idxName := c.IndexName(index)
	var buf bytes.Buffer

	for _, doc := range docs {
		action := map[string]any{
			"index": map[string]any{
				"_index": idxName,
				"_id":    doc.ID,
			},
		}
		actionJSON, _ := json.Marshal(action)
		buf.Write(actionJSON)
		buf.WriteByte('\n')

		docJSON, _ := json.Marshal(doc.Source)
		buf.Write(docJSON)
		buf.WriteByte('\n')
	}

	resp, err := c.doRequest(ctx, "POST", "/_bulk", &buf)
	if err != nil {
		return 0, len(docs), err
	}
	defer resp.Body.Close()

	var result struct {
		Errors bool `json:"errors"`
		Items  []map[string]struct {
			Status int `json:"status"`
		} `json:"items"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, len(docs), err
	}

	for _, item := range result.Items {
		for _, op := range item {
			if op.Status >= 200 && op.Status < 300 {
				success++
			} else {
				failed++
			}
		}
	}

	return success, failed, nil
}

// BulkDoc 批量索引的文档。
type BulkDoc struct {
	ID     string
	Source any
}

// --- Search ---

// SearchRequest ES 搜索请求体。
type SearchRequest struct {
	Query     any   `json:"query"`
	From      int   `json:"from,omitempty"`
	Size      int   `json:"size,omitempty"`
	Sort      []any `json:"sort,omitempty"`
	Highlight any   `json:"highlight,omitempty"`
	Aggs      any   `json:"aggs,omitempty"`
	Source    any   `json:"_source,omitempty"`
}

// SearchResponse ES 搜索响应。
type SearchResponse struct {
	Took     int  `json:"took"`
	TimedOut bool `json:"timed_out"`
	Hits     struct {
		Total struct {
			Value    int    `json:"value"`
			Relation string `json:"relation"`
		} `json:"total"`
		MaxScore float64 `json:"max_score"`
		Hits     []struct {
			ID        string              `json:"_id"`
			Score     float64             `json:"_score"`
			Source    map[string]any      `json:"_source"`
			Highlight map[string][]string `json:"highlight,omitempty"`
		} `json:"hits"`
	} `json:"hits"`
	Aggregations map[string]any `json:"aggregations,omitempty"`
}

// Search 执行搜索查询。
func (c *Client) Search(ctx context.Context, index string, req SearchRequest) (*SearchResponse, error) {
	idxName := c.IndexName(index)
	body, _ := json.Marshal(req)

	resp, err := c.doRequest(ctx, "POST", "/"+idxName+"/_search", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, errs.ErrInternal.Wrap(
			fmt.Errorf("es: search (status=%d): %s", resp.StatusCode, string(respBody)),
		)
	}

	var sr SearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&sr); err != nil {
		return nil, errs.ErrInternal.Wrap(err)
	}
	return &sr, nil
}

// --- Internal Helpers ---

func (c *Client) doRequest(ctx context.Context, method, path string, body io.Reader) (*http.Response, error) {
	// 轮询 URL（简单负载均衡）
	url := c.cfg.URLs[0]
	if len(c.cfg.URLs) > 1 {
		idx := int(time.Now().UnixNano()) % len(c.cfg.URLs)
		url = c.cfg.URLs[idx]
	}

	req, err := http.NewRequestWithContext(ctx, method, url+path, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	if c.cfg.Username != "" {
		req.SetBasicAuth(c.cfg.Username, c.cfg.Password)
	}

	// 带重试
	var lastErr error
	for attempt := 0; attempt <= c.cfg.MaxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt*100) * time.Millisecond)
		}

		resp, err := c.http.Do(req)
		if err != nil {
			lastErr = err
			continue
		}

		// 检查是否需要重试
		if c.shouldRetry(resp.StatusCode) {
			resp.Body.Close()
			continue
		}

		return resp, nil
	}

	return nil, fmt.Errorf("es: request failed after %d retries: %w", c.cfg.MaxRetries+1, lastErr)
}

func (c *Client) shouldRetry(statusCode int) bool {
	for _, code := range c.cfg.RetryOnStatus {
		if statusCode == code {
			return true
		}
	}
	return false
}

// Close 关闭客户端（释放连接池）。
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return nil
	}
	c.closed = true
	c.http.CloseIdleConnections()
	return nil
}

// --- 便捷构造函数 ---

// MatchQuery 构建 match 查询。
func MatchQuery(field, value string) map[string]any {
	return map[string]any{
		"match": map[string]any{
			field: value,
		},
	}
}

// MultiMatchQuery 构建 multi_match 查询（跨多字段）。
func MultiMatchQuery(query string, fields []string, operator string) map[string]any {
	m := map[string]any{
		"multi_match": map[string]any{
			"query":  query,
			"fields": fields,
		},
	}
	if operator != "" {
		m["multi_match"].(map[string]any)["operator"] = operator
	}
	return m
}

// TermQuery 构建精确匹配 term 查询。
func TermQuery(field string, value any) map[string]any {
	return map[string]any{
		"term": map[string]any{
			field: map[string]any{
				"value": value,
			},
		},
	}
}

// TermsQuery 构建 terms 查询（in 语义）。
func TermsQuery(field string, values []any) map[string]any {
	return map[string]any{
		"terms": map[string]any{
			field: values,
		},
	}
}

// RangeQuery 构建范围查询。
func RangeQuery(field string, gte, lte, gt, lt any) map[string]any {
	rangeMap := map[string]any{}
	if gte != nil {
		rangeMap["gte"] = gte
	}
	if lte != nil {
		rangeMap["lte"] = lte
	}
	if gt != nil {
		rangeMap["gt"] = gt
	}
	if lt != nil {
		rangeMap["lt"] = lt
	}
	return map[string]any{
		"range": map[string]any{
			field: rangeMap,
		},
	}
}

// BoolQuery 构建 bool 查询。
func BoolQuery(must, filter, should, mustNot []any) map[string]any {
	bq := map[string]any{}
	if len(must) > 0 {
		bq["must"] = must
	}
	if len(filter) > 0 {
		bq["filter"] = filter
	}
	if len(should) > 0 {
		bq["should"] = should
	}
	if len(mustNot) > 0 {
		bq["must_not"] = mustNot
	}
	return map[string]any{"bool": bq}
}

// HighlightConfig 构建高亮配置。
func HighlightConfig(fields ...string) map[string]any {
	fieldMap := map[string]any{}
	for _, f := range fields {
		fieldMap[f] = map[string]any{
			"fragment_size":       150,
			"number_of_fragments": 3,
			"pre_tags":            []string{"<mark>"},
			"post_tags":           []string{"</mark>"},
		}
	}
	return map[string]any{
		"fields": fieldMap,
	}
}

// --- ES 索引 Mapping 定义 ---

// IssueMapping 返回 issues 索引的 ES mapping。
// 使用 IK 分词器（ik_max_word 索引 / ik_smart 搜索）。
func IssueMapping() map[string]any {
	return map[string]any{
		"settings": map[string]any{
			"index": map[string]any{
				"number_of_shards":   3,
				"number_of_replicas": 1,
				"refresh_interval":   "5s",
			},
			"analysis": map[string]any{
				"analyzer": map[string]any{
					"ik_smart_analyzer": map[string]any{
						"type": "ik_smart",
					},
					"ik_max_word_analyzer": map[string]any{
						"type": "ik_max_word",
					},
				},
			},
		},
		"mappings": map[string]any{
			"properties": map[string]any{
				"workspace_id": map[string]any{"type": "long"},
				"project_id":   map[string]any{"type": "long"},
				"doc_type":     map[string]any{"type": "keyword"},
				"doc_id":       map[string]any{"type": "long"},
				"identifier":   map[string]any{"type": "keyword"},
				"type_code":    map[string]any{"type": "keyword"},
				"title": map[string]any{
					"type":            "text",
					"analyzer":        "ik_max_word_analyzer",
					"search_analyzer": "ik_smart_analyzer",
					"fields": map[string]any{
						"raw": map[string]any{"type": "keyword"},
						"pinyin": map[string]any{
							"type": "text",
							// pinyin analyzer optional - requires pinyin plugin
						},
					},
				},
				"content": map[string]any{
					"type":            "text",
					"analyzer":        "ik_max_word_analyzer",
					"search_analyzer": "ik_smart_analyzer",
				},
				"state_id":     map[string]any{"type": "long"},
				"state_name":   map[string]any{"type": "keyword"},
				"priority":     map[string]any{"type": "keyword"},
				"severity":     map[string]any{"type": "byte"},
				"assignee_ids": map[string]any{"type": "long"},
				"label_ids":    map[string]any{"type": "long"},
				"module_ids":   map[string]any{"type": "long"},
				"sprint_id":    map[string]any{"type": "long"},
				"version_id":   map[string]any{"type": "long"},
				"created_by":   map[string]any{"type": "long"},
				"target_date":  map[string]any{"type": "date"},
				"created_at":   map[string]any{"type": "date"},
				"updated_at":   map[string]any{"type": "date"},
				"deleted":      map[string]any{"type": "date"},
				"parent_id":    map[string]any{"type": "long"},
			},
		},
	}
}

// SprintMapping 返回 sprints 索引的 ES mapping。
func SprintMapping() map[string]any {
	return map[string]any{
		"settings": map[string]any{
			"index": map[string]any{
				"number_of_shards":   1,
				"number_of_replicas": 1,
			},
		},
		"mappings": map[string]any{
			"properties": map[string]any{
				"workspace_id": map[string]any{"type": "long"},
				"project_id":   map[string]any{"type": "long"},
				"doc_type":     map[string]any{"type": "keyword"},
				"doc_id":       map[string]any{"type": "long"},
				"title": map[string]any{
					"type":            "text",
					"analyzer":        "ik_max_word_analyzer",
					"search_analyzer": "ik_smart_analyzer",
				},
				"content": map[string]any{
					"type":            "text",
					"analyzer":        "ik_max_word_analyzer",
					"search_analyzer": "ik_smart_analyzer",
				},
				"status":     map[string]any{"type": "keyword"},
				"start_date": map[string]any{"type": "date"},
				"end_date":   map[string]any{"type": "date"},
				"created_at": map[string]any{"type": "date"},
				"updated_at": map[string]any{"type": "date"},
			},
		},
	}
}

// VersionMapping 返回 versions 索引的 ES mapping。
func VersionMapping() map[string]any {
	return map[string]any{
		"settings": map[string]any{
			"index": map[string]any{
				"number_of_shards":   1,
				"number_of_replicas": 1,
			},
		},
		"mappings": map[string]any{
			"properties": map[string]any{
				"workspace_id": map[string]any{"type": "long"},
				"project_id":   map[string]any{"type": "long"},
				"doc_type":     map[string]any{"type": "keyword"},
				"doc_id":       map[string]any{"type": "long"},
				"title": map[string]any{
					"type":            "text",
					"analyzer":        "ik_max_word_analyzer",
					"search_analyzer": "ik_smart_analyzer",
				},
				"content": map[string]any{
					"type":            "text",
					"analyzer":        "ik_max_word_analyzer",
					"search_analyzer": "ik_smart_analyzer",
				},
				"status":       map[string]any{"type": "keyword"},
				"release_date": map[string]any{"type": "date"},
				"created_at":   map[string]any{"type": "date"},
				"updated_at":   map[string]any{"type": "date"},
			},
		},
	}
}

// ReindexResult 索引重建结果。
type ReindexResult struct {
	Total      int   `json:"total"`
	Indexed    int   `json:"indexed"`
	Failed     int   `json:"failed"`
	DurationMs int64 `json:"duration_ms"`
}

// ReindexStats 带统计的重建结果（原子计数器汇总）。
type ReindexStats struct {
	Total    int64 `json:"total"`
	Success  int64 `json:"success"`
	Failed   int64 `json:"failed"`
	DurationMs int64 `json:"duration_ms"`
}

// Reindex 从 PostgreSQL search_documents 表全量重建 ES 索引。
// 采用蓝绿部署：先建新索引 → bulk 写入 → 切别名 → 删旧索引。
func (c *Client) Reindex(ctx context.Context, index string, mapping map[string]any, source <-chan BulkDoc) (*ReindexResult, error) {
	start := time.Now()
	baseIdx := c.IndexName(index)
	alias := baseIdx + "_current"
	newIdx := baseIdx + "_" + time.Now().Format("20060102_150405")

	// 1. 创建新索引
	if err := c.CreateIndex(ctx, newIdx, mapping); err != nil {
		return nil, err
	}

	// 2. 批量写入
	result := &ReindexResult{}
	batch := make([]BulkDoc, 0, 1000)

	for doc := range source {
		batch = append(batch, doc)
		result.Total++

		if len(batch) >= 1000 {
			success, failed, err := c.BulkIndex(ctx, newIdx, batch)
			if err != nil {
				// 清理失败的索引
				_ = c.DeleteIndex(ctx, newIdx)
				return nil, fmt.Errorf("es: reindex bulk failed at %d: %w", result.Total, err)
			}
			result.Indexed += success
			result.Failed += failed
			batch = batch[:0]
		}
	}

	// 处理最后一批
	if len(batch) > 0 {
		success, failed, err := c.BulkIndex(ctx, newIdx, batch)
		if err != nil {
			_ = c.DeleteIndex(ctx, newIdx)
			return nil, fmt.Errorf("es: reindex final bulk failed: %w", err)
		}
		result.Indexed += success
		result.Failed += failed
	}

	// 3. 刷新索引确保数据可见
	if err := c.refreshIndex(ctx, newIdx); err != nil {
		_ = c.DeleteIndex(ctx, newIdx)
		return nil, err
	}

	// 4. 原子切换别名
	exists, _ := c.AliasExists(ctx, alias)
	if exists {
		// 获取旧索引名
		oldIdx := baseIdx + "_old_" + time.Now().Format("20060102_150405")
		if err := c.SwapAlias(ctx, alias, baseIdx+"_current", newIdx); err != nil {
			_ = c.DeleteIndex(ctx, newIdx)
			return nil, err
		}
		// 尝试删除旧索引（失败不影响主流程）
		_ = c.DeleteIndex(ctx, oldIdx)
	} else {
		// 首次创建别名
		body := map[string]any{
			"actions": []map[string]any{
				{"add": map[string]any{"index": newIdx, "alias": alias}},
			},
		}
		jsonBody, _ := json.Marshal(body)
		resp, err := c.doRequest(ctx, "POST", "/_aliases", bytes.NewReader(jsonBody))
		if err != nil {
			_ = c.DeleteIndex(ctx, newIdx)
			return nil, err
		}
		resp.Body.Close()
	}

	result.DurationMs = time.Since(start).Milliseconds()
	return result, nil
}

func (c *Client) refreshIndex(ctx context.Context, index string) error {
	resp, err := c.doRequest(ctx, "POST", "/"+index+"/_refresh", nil)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

// --- Reindex Source ---

// ReindexSource 提供从 PG 读取数据以重建索引的通道。
type ReindexSource func(ctx context.Context) (<-chan BulkDoc, <-chan error)

// --- Parallel Reindex (对标 Google SRE Book Ch.17: Worker Pool) ---

// reindexDoc 是 producer→worker 投递的内部消息。
type reindexDoc struct {
	workspaceID int64
	doc         BulkDoc
}

// ReindexAll 使用 worker pool 并行重建索引（兼容旧接口：仅返回 error）。
// 由 DB 按 workspace_id hash 分区投递到各 worker，各自独立构建 bulk request。
func (c *Client) ReindexAll(ctx context.Context, index, docType string, mapping map[string]any, db *pgxpool.Pool) error {
	_, err := c.ReindexAllWithStats(ctx, index, docType, mapping, db)
	return err
}

// ReindexAllWithStats 带统计版本的并行重建。
//
// 工作流:
//  1. producer 从 DB 查询 search_documents，按 workspace_id % workers hash 分区
//  2. 各 worker 从独立 channel 读取文档，累积到 ReindexBatchSize 后执行 BulkIndex
//  3. 使用 atomic 汇总 total/success/failed
//  4. 蓝绿切换别名并刷新
func (c *Client) ReindexAllWithStats(ctx context.Context, index, docType string, mapping map[string]any, db *pgxpool.Pool) (*ReindexStats, error) {
	start := time.Now()
	workers := c.ReindexWorkers
	if workers < 1 {
		workers = 4
	}
	batchSize := c.ReindexBatchSize
	if batchSize < 1 {
		batchSize = 1000
	}

	baseIdx := c.IndexName(index)
	alias := baseIdx + "_current"
	newIdx := baseIdx + "_" + time.Now().Format("20060102_150405")

	// 1. 创建新索引
	if err := c.CreateIndex(ctx, newIdx, mapping); err != nil {
		return nil, err
	}

	// 2. 启动 worker pool + producer
	stats := &ReindexStats{}
	var success atomic.Int64
	var failed atomic.Int64
	var total atomic.Int64

	// worker channel buffer = workers * 10
	workerChs := make([]chan reindexDoc, workers)
	for i := 0; i < workers; i++ {
		workerChs[i] = make(chan reindexDoc, workers*10)
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var wg sync.WaitGroup

	// worker goroutines
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(id int, ch <-chan reindexDoc) {
			defer wg.Done()
			batch := make([]BulkDoc, 0, batchSize)
			for d := range ch {
				batch = append(batch, d.doc)
				if len(batch) >= batchSize {
					c.flushBatch(ctx, newIdx, docType, batch, &success, &failed)
					batch = batch[:0]
				}
			}
			// flush remaining
			if len(batch) > 0 {
				c.flushBatch(ctx, newIdx, docType, batch, &success, &failed)
			}
		}(i, workerChs[i])
	}

	// producer: 从 DB 读取 search_documents，按 workspace_id hash 分区投递
producerLoop:
	for {
		rows, err := db.Query(ctx, `
			SELECT workspace_id, project_id, doc_type, doc_id, identifier, title, content,
			       state_id, state_name, priority, severity, assignee_ids, label_ids,
			       module_ids, sprint_id, version_id, created_by, target_date, created_at, updated_at
			FROM search_documents
			WHERE doc_type = $1
			ORDER BY workspace_id, doc_id`, docType)
		if err != nil {
			cancel()
			for i := 0; i < workers; i++ {
				close(workerChs[i])
			}
			wg.Wait()
			_ = c.DeleteIndex(ctx, newIdx)
			return nil, fmt.Errorf("es: reindex query: %w", err)
		}

		for rows.Next() {
			var wsID, projectID, docID, stateID int64
			var identifier, docTypeDB, title, content, stateName, priority string
			var severity int
			var assigneeIDs, labelIDs, moduleIDs []int64
			var sprintID, versionID, createdBy int64
			var targetDate sql.NullTime
			var createdAt, updatedAt time.Time

			err := rows.Scan(&wsID, &projectID, &docTypeDB, &docID, &identifier, &title, &content,
				&stateID, &stateName, &priority, &severity, &assigneeIDs, &labelIDs,
				&moduleIDs, &sprintID, &versionID, &createdBy, &targetDate, &createdAt, &updatedAt)
			if err != nil {
				rows.Close()
				cancel()
				for i := 0; i < workers; i++ {
					close(workerChs[i])
				}
				wg.Wait()
				_ = c.DeleteIndex(ctx, newIdx)
				return nil, fmt.Errorf("es: reindex scan: %w", err)
			}

			total.Add(1)

			doc := BulkDoc{
				ID: fmt.Sprintf("%d_%d_%s_%d", wsID, projectID, docTypeDB, docID),
				Source: map[string]any{
					"workspace_id": wsID,
					"project_id":   projectID,
					"doc_type":     docTypeDB,
					"doc_id":       docID,
					"identifier":   identifier,
					"title":        title,
					"content":      content,
					"state_id":     stateID,
					"state_name":   stateName,
					"priority":     priority,
					"severity":     severity,
					"assignee_ids": assigneeIDs,
					"label_ids":    labelIDs,
					"module_ids":   moduleIDs,
					"sprint_id":    sprintID,
					"version_id":   versionID,
					"created_by":   createdBy,
					"target_date":  targetDate,
					"created_at":   createdAt,
					"updated_at":   updatedAt,
				},
			}

			// hash 分区
			workerID := int(wsID) % workers
			if workerID < 0 {
				workerID = -workerID
			}
			select {
			case <-ctx.Done():
				rows.Close()
				break producerLoop
			case workerChs[workerID] <- reindexDoc{workspaceID: wsID, doc: doc}:
			}
		}
		rows.Close()
		break
	}

	// 关闭所有 worker channel（producer 完成）
	for i := 0; i < workers; i++ {
		close(workerChs[i])
	}
	wg.Wait()

	// 3. 刷新索引
	if err := c.refreshIndex(ctx, newIdx); err != nil {
		_ = c.DeleteIndex(ctx, newIdx)
		return nil, err
	}

	// 4. 原子切换别名
	exists, _ := c.AliasExists(ctx, alias)
	if exists {
		if err := c.SwapAlias(ctx, alias, baseIdx+"_current", newIdx); err != nil {
			_ = c.DeleteIndex(ctx, newIdx)
			return nil, err
		}
	} else {
		body := map[string]any{
			"actions": []map[string]any{
				{"add": map[string]any{"index": newIdx, "alias": alias}},
			},
		}
		jsonBody, _ := json.Marshal(body)
		resp, err := c.doRequest(ctx, "POST", "/_aliases", bytes.NewReader(jsonBody))
		if err != nil {
			_ = c.DeleteIndex(ctx, newIdx)
			return nil, err
		}
		resp.Body.Close()
	}

	stats.Total = total.Load()
	stats.Success = success.Load()
	stats.Failed = failed.Load()
	stats.DurationMs = time.Since(start).Milliseconds()

	// P0-5：更新 Prometheus 进度 gauge
	if stats.Total > 0 {
		ReindexProgress.WithLabelValues(docType).Set(float64(stats.Success+stats.Failed) / float64(stats.Total))
	}

	if c.logger != nil {
		c.logger.Info("es reindex completed",
			zap.String("doc_type", docType),
			zap.Int64("total", stats.Total),
			zap.Int64("success", stats.Success),
			zap.Int64("failed", stats.Failed),
			zap.Int64("duration_ms", stats.DurationMs),
			zap.Int("workers", workers))
	}

	return stats, nil
}

// flushBatch 将一批文档 bulk 写入 ES 并原子更新统计。
// P0-5 增强：自适应 bulk size — 根据响应动态调整 batch 大小。
// 当 ES 响应快（<100ms）时逐步增大 batch（上限 5000），
// 当 ES 响应慢（>1s）时减小 batch（下限 500），避免过载。
func (c *Client) flushBatch(ctx context.Context, index, docType string, batch []BulkDoc, success, failed *atomic.Int64) {
	ReindexRequestsTotal.WithLabelValues(docType).Inc()
	start := time.Now()
	s, f, err := c.BulkIndex(ctx, index, batch)
	elapsed := time.Since(start)
	if err != nil {
		ReindexErrorsTotal.WithLabelValues(docType).Inc()
		failed.Add(int64(len(batch)))
		// 失败时减小 batch size（下限 500）
		if c.ReindexBatchSize > 500 {
			c.ReindexBatchSize = c.ReindexBatchSize / 2
			if c.ReindexBatchSize < 500 {
				c.ReindexBatchSize = 500
			}
		}
		return
	}
	success.Add(int64(s))
	failed.Add(int64(f))
	// 自适应：响应快时逐步增大 batch（上限 5000）
	if elapsed < 100*time.Millisecond && c.ReindexBatchSize < 5000 {
		c.ReindexBatchSize = int(float64(c.ReindexBatchSize) * 1.5)
		if c.ReindexBatchSize > 5000 {
			c.ReindexBatchSize = 5000
		}
	} else if elapsed > time.Second && c.ReindexBatchSize > 500 {
		// 响应慢时减小 batch
		c.ReindexBatchSize = c.ReindexBatchSize / 2
		if c.ReindexBatchSize < 500 {
			c.ReindexBatchSize = 500
		}
	}

	if f > 0 {
		ReindexErrorsTotal.WithLabelValues(docType).Add(float64(f))
	}
}
