package issue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/olivere/elastic/v7"

	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// WorkitemSearchService 工作项搜索服务
// 支持分类型ES索引，ES不可用时自动降级到PostgreSQL全文搜索，对齐项目现有降级策略
type WorkitemSearchService struct {
	db          *pgxpool.Pool
	esClient    *elastic.Client
	redisClient *redis.Client
	useES       bool // 是否使用ES，ES故障时自动置为false
}

// NewWorkitemSearchService 创建搜索服务
func NewWorkitemSearchService(db *pgxpool.Pool, esClient *elastic.Client, redisClient *redis.Client) *WorkitemSearchService {
	return &WorkitemSearchService{
		db:          db,
		esClient:    esClient,
		redisClient: redisClient,
		useES:       esClient != nil,
	}
}

// IndexWorkitem 索引工作项到ES（分类型索引，索引名规则：workitem_requirement / workitem_task / workitem_defect）
func (s *WorkitemSearchService) IndexWorkitem(ctx context.Context, wsID int64, entityType IssueTypeCode, entity any) error {
	if !s.useES || s.esClient == nil {
		return nil
	}
	
	indexName := s.getIndexName(entityType)
	_, err := s.esClient.Index().
		Index(indexName).
		Id(fmt.Sprintf("%d_%d", wsID, getEntityID(entity))).
		BodyJson(entity).
		Do(ctx)
	if err != nil {
		// 索引失败时记录，不影响主业务
		return nil
	}
	return nil
}

// DeleteWorkitemIndex 删除工作项的ES索引
func (s *WorkitemSearchService) DeleteWorkitemIndex(ctx context.Context, wsID int64, entityType IssueTypeCode, entityID int64) error {
	if !s.useES || s.esClient == nil {
		return nil
	}
	
	indexName := s.getIndexName(entityType)
	_, err := s.esClient.Delete().
		Index(indexName).
		Id(fmt.Sprintf("%d_%d", wsID, entityID)).
		Do(ctx)
	if err != nil {
		return nil
	}
	return nil
}

// Search 搜索工作项，支持按类型筛选，ES不可用时自动降级到PG搜索
func (s *WorkitemSearchService) Search(ctx context.Context, wsID, projectID int64, entityType IssueTypeCode, keyword string, page, perPage int) ([]map[string]any, int, error) {
	if s.useES && s.esClient != nil {
		// 先尝试ES搜索
		_, total, err := s.searchFromES(ctx, wsID, projectID, entityType, keyword, page, perPage)
		if err == nil {
			return nil, total, nil
		}
		// ES失败，降级到PG搜索
		s.useES = false
		// 5分钟后重新尝试ES
		go func() {
			time.Sleep(5 * time.Minute)
			s.useES = s.esClient != nil
		}()
	}
	
	// PG全文搜索降级
	return s.searchFromPG(ctx, wsID, projectID, entityType, keyword, page, perPage)
}

// searchFromES 从ES搜索
func (s *WorkitemSearchService) searchFromES(ctx context.Context, wsID, projectID int64, entityType IssueTypeCode, keyword string, page, perPage int) ([]map[string]any, int, error) {
	indexName := s.getIndexName(entityType)
	query := elastic.NewBoolQuery().
		Must(elastic.NewMultiMatchQuery(keyword, "name", "description_stripped")).
		Filter(elastic.NewTermQuery("workspace_id", wsID)).
		Filter(elastic.NewTermQuery("project_id", projectID)).
		Filter(elastic.NewTermQuery("archived_at", nil))
	
	from := (page - 1) * perPage
	result, err := s.esClient.Search().
		Index(indexName).
		Query(query).
		From(from).
		Size(perPage).
		Do(ctx)
	if err != nil {
		return nil, 0, err
	}
	
	var items []map[string]any
	for _, hit := range result.Hits.Hits {
		var item map[string]any
		_ = json.Unmarshal(hit.Source, &item)
		items = append(items, item)
	}
	return items, int(result.Hits.TotalHits.Value), nil
}

// searchFromPG 从PostgreSQL全文搜索（降级逻辑，使用已有的GIN索引）
func (s *WorkitemSearchService) searchFromPG(ctx context.Context, wsID, projectID int64, entityType IssueTypeCode, keyword string, page, perPage int) ([]map[string]any, int, error) {
	var tableName string
	switch entityType {
	case TypeTask:
		tableName = "task"
	case TypeRequirement:
		tableName = "requirement"
	case TypeDefect:
		tableName = "defect"
	default:
		return nil, 0, errs.ErrValidation.WithDetails(errs.FieldDetail{Field: "entity_type", Reason: "不支持的工作项类型"})
	}
	
	from := (page - 1) * perPage
	rows, err := s.db.Query(ctx, fmt.Sprintf(`
		SELECT id, name, description_stripped, state_id, priority, created_at, updated_at
		FROM %s 
		WHERE workspace_id = $1 
		  AND project_id = $2 
		  AND archived_at IS NULL
		  AND to_tsvector('simple', coalesce(name,'') || ' ' || coalesce(description_stripped,'')) @@ plainto_tsquery('simple', $3)
		ORDER BY ts_rank(to_tsvector('simple', coalesce(name,'') || ' ' || coalesce(description_stripped,'')), plainto_tsquery('simple', $3)) DESC, updated_at DESC
		LIMIT $4 OFFSET $5
	`, tableName), wsID, projectID, keyword, perPage, from)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	
	var items []map[string]any
	for rows.Next() {
		var id, stateID int64
		var name, desc string
		var priority string
		var createdAt, updatedAt time.Time
		if err := rows.Scan(&id, &name, &desc, &stateID, &priority, &createdAt, &updatedAt); err != nil {
			return nil, 0, err
		}
		items = append(items, map[string]any{
			"id": id,
			"name": name,
			"description_stripped": desc,
			"state_id": stateID,
			"priority": priority,
			"created_at": createdAt,
			"updated_at": updatedAt,
		})
	}
	
	// 查询总数
	var total int
	err = s.db.QueryRow(ctx, fmt.Sprintf(`
		SELECT COUNT(*) FROM %s 
		WHERE workspace_id = $1 
		  AND project_id = $2 
		  AND archived_at IS NULL
		  AND to_tsvector('simple', coalesce(name,'') || ' ' || coalesce(description_stripped,'')) @@ plainto_tsquery('simple', $3)
	`, tableName), wsID, projectID, keyword).Scan(&total)
	if err != nil {
		return nil, 0, err
	}
	
	return items, total, nil
}

// getIndexName 生成ES索引名
func (s *WorkitemSearchService) getIndexName(entityType IssueTypeCode) string {
	return fmt.Sprintf("workitem_%s", entityType)
}

// getEntityID 辅助函数，获取工作项ID，多种类型适配
func getEntityID(entity any) int64 {
	switch e := entity.(type) {
	case Task:
		return e.ID
	case Requirement:
		return e.ID
	case Defect:
		return e.ID
	}
	return 0
}
