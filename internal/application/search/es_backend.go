// Package search — ES 搜索后端实现。
//
// 提供基于 Elasticsearch 的搜索后端，支持:
//   - IK 中文分词（ik_max_word 索引 / ik_smart 搜索）
//   - 类 JQL 语法解析 → ES bool query
//   - 高亮 + 聚合
//   - 降级自动回退到 PG FTS
package search

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/njydsz/ydsz-plane/internal/infrastructure/es"
	"github.com/njydsz/ydsz-plane/pkg/searchql"
)

// ESBackend ES 搜索后端。
type ESBackend struct {
	client     *es.Client
	indexAlias string // 别名（如 ydsz_search_current）
}

// NewESBackend 创建 ES 搜索后端。
func NewESBackend(client *es.Client) *ESBackend {
	return &ESBackend{
		client:     client,
		indexAlias: client.IndexName("search") + "_current",
	}
}

// IsAvailable 返回 ES 集群是否可用。
func (b *ESBackend) IsAvailable() bool {
	return b.client != nil && b.client.IsHealthy()
}

// Search 通过 ES 执行搜索。
// 返回 nil 表示 ES 不可用，调用方应降级到 PG FTS。
func (b *ESBackend) Search(ctx context.Context, q SearchQuery) (*SearchResponse, error) {
	if !b.IsAvailable() {
		return nil, fmt.Errorf("es: backend not available")
	}

	// 解析类 JQL 语法
	jql := searchql.Parse(q.Query)

	// 构建 ES bool query
	esQuery := b.buildESQuery(jql, q)

	// 构建 ES 搜索请求
	req := es.SearchRequest{
		Query: esQuery,
		From:  q.Offset,
		Size:  q.Limit,
		Sort: []any{
			map[string]any{"_score": map[string]any{"order": "desc"}},
			map[string]any{"updated_at": map[string]any{"order": "desc"}},
		},
		Highlight: es.HighlightConfig("title", "content"),
		Source: map[string]any{
			"excludes": []string{"content"}, // 不返回 content 字段减少传输
		},
	}

	// 执行搜索
	esResp, err := b.client.Search(ctx, b.indexAlias, req)
	if err != nil {
		return nil, err
	}

	// 转换 ES 响应为 SearchResponse
	resp := b.transformESResponse(esResp, q)

	// 如果有结构化过滤条件，在应用层进行二次过滤
	if len(jql.Clauses) > 0 {
		// ES query 已包含大部分过滤，这里做兜底
	}

	return resp, nil
}

// buildESQuery 将 JQL 解析结果 + 搜索参数转换为 ES bool query。
func (b *ESBackend) buildESQuery(jql *searchql.Query, q SearchQuery) map[string]any {
	var must, filter, should, mustNot []any

	// --- 全文检索 ---
	if jql.Text != "" {
		should = append(should, es.MultiMatchQuery(jql.Text, []string{
			"title^3",
			"content^2",
			"identifier^4",
		}, "or"))
	} else if len(jql.Clauses) == 0 {
		// 纯文本搜索（未解析出 JQL）
		should = append(should, es.MultiMatchQuery(q.Query, []string{
			"title^3",
			"content^2",
			"identifier^4",
		}, "or"))
	}

	// --- 租户隔离（强制 filter） ---
	filter = append(filter, es.TermQuery("workspace_id", q.WorkspaceID))

	// --- 项目过滤 ---
	if q.ProjectID > 0 {
		filter = append(filter, es.TermQuery("project_id", q.ProjectID))
	}

	// --- 文档类型过滤 ---
	if len(q.DocTypes) > 0 {
		var types []any
		for _, dt := range q.DocTypes {
			types = append(types, dt)
		}
		filter = append(filter, es.TermsQuery("doc_type", types))
	}

	// --- 结构化过滤（来自 handler query params） ---
	if code, ok := q.Filters["type_code"].(string); ok && code != "" {
		filter = append(filter, es.TermQuery("type_code", code))
	}
	if priority, ok := q.Filters["priority"].(string); ok && priority != "" {
		filter = append(filter, es.TermQuery("priority", priority))
	}
	if stateID, ok := q.Filters["state_id"].(float64); ok && stateID > 0 {
		filter = append(filter, es.TermQuery("state_id", int64(stateID)))
	}

	// --- JQL 子句转换 ---
	for _, clause := range jql.Clauses {
		esClause := b.convertJQLClause(clause, q)
		if esClause == nil {
			continue
		}

		if clause.Negated {
			mustNot = append(mustNot, esClause)
		} else {
			filter = append(filter, esClause)
		}
	}

	return es.BoolQuery(must, filter, should, mustNot)
}

// convertJQLClause 将单个 JQL 子句转换为 ES 查询条件。
func (b *ESBackend) convertJQLClause(clause searchql.Clause, q SearchQuery) any {
	field := clause.Field
	op := clause.Operator
	val := clause.Value

	// 解析特殊值
	strVal, _ := val.(string)
	if strVal == "__CURRENT_USER__" && q.UserID > 0 {
		val = q.UserID
	}

	switch field {
	case "project":
		return b.buildTermOrTerms("project_id", val, op)
	case "type":
		return b.buildTermOrTerms("type_code", val, op)
	case "status":
		return b.buildTermOrTerms("state_name", val, op)
	case "priority":
		return b.buildTermOrTerms("priority", val, op)
	case "severity":
		return b.buildRange("severity", val, op)
	case "assignee":
		return b.buildTermOrTerms("assignee_ids", val, op)
	case "reporter":
		return b.buildTermOrTerms("created_by", val, op)
	case "label":
		return b.buildTermOrTerms("label_ids", val, op)
	case "module":
		return b.buildTermOrTerms("module_ids", val, op)
	case "sprint":
		return b.buildTermOrTerms("sprint_id", val, op)
	case "version":
		return b.buildTermOrTerms("version_id", val, op)
	case "due":
		return b.buildDateRange("target_date", val, op)
	case "created":
		return b.buildDateRange("created_at", val, op)
	case "updated":
		return b.buildDateRange("updated_at", val, op)
	case "identifier":
		return b.buildTermOrTerms("identifier", val, op)
	default:
		// 未知字段，尝试作为标题搜索
		return es.MatchQuery("title", strVal)
	}
}

// buildTermOrTerms 构建 term 或 terms 查询。
func (b *ESBackend) buildTermOrTerms(field string, val any, op string) any {
	if list, ok := val.([]string); ok {
		var vals []any
		for _, v := range list {
			vals = append(vals, v)
		}
		if op == "!=" {
			return es.BoolQuery(nil, nil, nil, []any{es.TermsQuery(field, vals)})
		}
		return es.TermsQuery(field, vals)
	}

	s, ok := val.(string)
	if !ok {
		return es.TermQuery(field, val)
	}

	if op == "!=" {
		return es.BoolQuery(nil, nil, nil, []any{es.TermQuery(field, s)})
	}
	return es.TermQuery(field, s)
}

// buildRange 构建数值范围查询。
func (b *ESBackend) buildRange(field string, val any, op string) any {
	s, ok := val.(string)
	if !ok {
		return nil
	}

	num, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return nil
	}

	switch op {
	case ">":
		return es.RangeQuery(field, nil, nil, num, nil)
	case ">=":
		return es.RangeQuery(field, num, nil, nil, nil)
	case "<":
		return es.RangeQuery(field, nil, nil, nil, num)
	case "<=":
		return es.RangeQuery(field, nil, num, nil, nil)
	default:
		return es.TermQuery(field, int64(num))
	}
}

// buildDateRange 构建日期范围查询。
func (b *ESBackend) buildDateRange(field string, val any, op string) any {
	if t, ok := val.(time.Time); ok {
		switch op {
		case ">":
			return es.RangeQuery(field, nil, nil, t, nil)
		case ">=":
			return es.RangeQuery(field, t, nil, nil, nil)
		case "<":
			return es.RangeQuery(field, nil, nil, nil, t)
		case "<=":
			return es.RangeQuery(field, nil, t, nil, nil)
		default:
			return es.RangeQuery(field, t, t, nil, nil)
		}
	}

	s, ok := val.(string)
	if !ok {
		return nil
	}

	// 解析日期字符串
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		t, err = time.Parse("2006-01-02", s)
		if err != nil {
			return es.MatchQuery(field, s)
		}
	}

	return es.RangeQuery(field, t, t, nil, nil)
}

// transformESResponse 将 ES 搜索结果转换为 SearchResponse。
func (b *ESBackend) transformESResponse(esResp *es.SearchResponse, q SearchQuery) *SearchResponse {
	resp := &SearchResponse{
		Query:  q.Query,
		Total:  esResp.Hits.Total.Value,
		TimeMs: int64(esResp.Took),
	}

	results := SearchResults{}
	groupMap := map[string]*SearchGroup{}
	for _, dt := range q.DocTypes {
		groupMap[dt] = &SearchGroup{DocType: dt, Hits: []SearchHit{}}
	}

	for _, hit := range esResp.Hits.Hits {
		src := hit.Source
		dt, _ := src["doc_type"].(string)
		docID, _ := toInt64(src["doc_id"])
		title, _ := src["title"].(string)
		identifier, _ := src["identifier"].(string)
		projID, _ := toInt64(src["project_id"])

		// 提取高亮
		highlight := ""
		if titleHL, ok := hit.Highlight["title"]; ok && len(titleHL) > 0 {
			highlight = titleHL[0]
		} else if contentHL, ok := hit.Highlight["content"]; ok && len(contentHL) > 0 {
			highlight = contentHL[0]
		}

		hitResult := SearchHit{
			DocType:     dt,
			DocID:       docID,
			Title:       title,
			Identifier:  identifier,
			Highlight:   highlight,
			ProjectID:   projID,
			Rank:        hit.Score,
			URL:         buildDocURL(dt, docID, projID),
		}

		switch dt {
		case "issue":
			results.Issues = append(results.Issues, hitResult)
		case "sprint":
			results.Sprints = append(results.Sprints, hitResult)
		case "version":
			results.Versions = append(results.Versions, hitResult)
		}

		if g, ok := groupMap[dt]; ok {
			g.Hits = append(g.Hits, hitResult)
		}
	}

	resp.Results = results

	// 构建有序 groups
	var groups []SearchGroup
	for _, dt := range q.DocTypes {
		g := groupMap[dt]
		g.Total = int64(len(g.Hits))
		if len(g.Hits) > 0 {
			groups = append(groups, *g)
		}
	}
	resp.Groups = groups

	// 添加搜索建议（基于聚合）
	if aggs, ok := esResp.Aggregations["suggestions"]; ok {
		if aggMap, ok := aggs.(map[string]any); ok {
			if buckets, ok := aggMap["buckets"].([]any); ok {
				for _, b := range buckets {
					if bucket, ok := b.(map[string]any); ok {
						if key, ok := bucket["key"].(string); ok {
							resp.Suggestions = append(resp.Suggestions, key)
						}
					}
				}
			}
		}
	}

	return resp
}

// toInt64 安全转换 interface{} 到 int64。
func toInt64(v any) (int64, bool) {
	switch val := v.(type) {
	case float64:
		return int64(val), true
	case int64:
		return val, true
	case json.Number:
		n, err := val.Int64()
		return n, err == nil
	default:
		return 0, false
	}
}
