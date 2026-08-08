// Package metrics — 带缓存的效能度量 HTTP Handler 包装层。
//
// 在 Handler 外层叠加 Redis 读穿透缓存，对高频仪表盘查询进行性能优化。
// 缓存策略参见 cache.go。
//
// 失效策略：
//   - TTL 到期自动失效
//   - 工作项/迭代/版本变更事件触发主动失效（通过 Subscriber 监听）
package metrics

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/njydsz/ydsz-plane/internal/interfaces/middleware"
)

// CachedHandler 包装 Handler 并叠加缓存层。
type CachedHandler struct {
	*Handler
	cache *MetricCache
	log   *zap.Logger
}

// NewCachedHandler 创建带缓存的 Handler 包装。
//
// 如果 cli 为 nil，则回退为无缓存 Handler（开发模式降级）。
func NewCachedHandler(deps *HandlerDeps, cli *redis.Client, log *zap.Logger) *CachedHandler {
	inner := NewMetricsHandler(deps)
	var cache *MetricCache
	if cli != nil {
		cache = NewMetricCache(cli)
	}
	if log == nil {
		log = zap.NewNop()
	}
	return &CachedHandler{
		Handler: inner,
		cache:   cache,
		log:     log,
	}
}

// Register 注册带缓存的路由。
// 说明：velocity/lead-time/quality/dora/resource-load/snapshots 走读穿透
// 缓存；deployments（写入）、cfd/control-chart/throughput（低频分析查询）
// 委托内层 handler 直查，避免缓存复杂聚合结果造成口径漂移。
func (h *CachedHandler) Register(r *gin.RouterGroup) {
	r.GET("/velocity", h.GetVelocity)
	r.GET("/velocity/trend", h.GetVelocityTrend)
	r.GET("/lead-time", h.GetLeadTime)
	r.GET("/quality", h.GetQualityMetrics)
	r.GET("/dora", h.GetDORA)
	r.GET("/resource-load", h.GetResourceLoad)
	r.POST("/deployments", h.RecordDeployment)
	r.GET("/snapshots", h.ListSnapshots)

	// 未缓存的分析端点（委托内层 Handler，方法经嵌入提升）
	r.GET("/resource-load/detail", h.GetResourceLoadDetail)
	r.GET("/cfd", h.GetCFD)
	r.GET("/control-chart", h.GetControlChart)
	r.GET("/throughput", h.GetWeeklyThroughput)
}

// GetVelocity 查询项目速率统计（带缓存）。
func (h *CachedHandler) GetVelocity(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	projectID := c.GetInt64(middleware.CtxProjectID)
	lastN, _ := strconv.Atoi(c.DefaultQuery("last_n", "6"))

	// 1. 查缓存
	if h.cache != nil {
		if cached, ok := h.cache.GetVelocity(c.Request.Context(), wsID, projectID, lastN); ok {
			c.Header("X-Cache", "HIT")
			c.JSON(http.StatusOK, cached)
			return
		}
	}

	// 2. 缓存未命中，查 DB
	result, err := h.d.Svc.GetVelocity(c.Request.Context(), wsID, projectID, lastN)
	if err != nil {
		writeErr(c, err)
		return
	}

	// 3. 回写缓存
	if h.cache != nil {
		h.cache.SetVelocity(c.Request.Context(), wsID, projectID, lastN, result)
	}
	c.Header("X-Cache", "MISS")
	c.JSON(http.StatusOK, result)
}

// GetVelocityTrend 迭代速率趋势（复用 GetVelocity 逻辑，独立端点便于前端卡片绑定）。
func (h *CachedHandler) GetVelocityTrend(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	projectID := c.GetInt64(middleware.CtxProjectID)
	lastN, _ := strconv.Atoi(c.DefaultQuery("last_n", "6"))

	if h.cache != nil {
		if cached, ok := h.cache.GetVelocity(c.Request.Context(), wsID, projectID, lastN); ok {
			c.Header("X-Cache", "HIT")
			c.JSON(http.StatusOK, gin.H{
				"project_id":   cached.ProjectID,
				"trend":        cached.Trend,
				"average":      cached.Average,
				"sprint_count": cached.SprintCount,
			})
			return
		}
	}

	result, err := h.d.Svc.GetVelocity(c.Request.Context(), wsID, projectID, lastN)
	if err != nil {
		writeErr(c, err)
		return
	}
	if h.cache != nil {
		h.cache.SetVelocity(c.Request.Context(), wsID, projectID, lastN, result)
	}
	c.Header("X-Cache", "MISS")
	c.JSON(http.StatusOK, gin.H{
		"project_id":   result.ProjectID,
		"trend":        result.Trend,
		"average":      result.Average,
		"sprint_count": result.SprintCount,
	})
}

// GetLeadTime 需求前置时间（带缓存）。
func (h *CachedHandler) GetLeadTime(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	projectID := c.GetInt64(middleware.CtxProjectID)
	days, _ := strconv.Atoi(c.DefaultQuery("days", "90"))

	if h.cache != nil {
		if cached, ok := h.cache.GetLeadTime(c.Request.Context(), wsID, projectID, days); ok {
			c.Header("X-Cache", "HIT")
			c.JSON(http.StatusOK, cached)
			return
		}
	}

	result, err := h.d.Svc.GetLeadTime(c.Request.Context(), wsID, projectID, days)
	if err != nil {
		writeErr(c, err)
		return
	}
	if h.cache != nil {
		h.cache.SetLeadTime(c.Request.Context(), wsID, projectID, days, result)
	}
	c.Header("X-Cache", "MISS")
	c.JSON(http.StatusOK, result)
}

// GetQualityMetrics 质量指标（带缓存）。
func (h *CachedHandler) GetQualityMetrics(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	projectID := c.GetInt64(middleware.CtxProjectID)

	if h.cache != nil {
		if cached, ok := h.cache.GetQualityMetrics(c.Request.Context(), wsID, projectID); ok {
			c.Header("X-Cache", "HIT")
			c.JSON(http.StatusOK, cached)
			return
		}
	}

	result, err := h.d.Svc.GetQualityMetrics(c.Request.Context(), wsID, projectID)
	if err != nil {
		writeErr(c, err)
		return
	}
	if h.cache != nil {
		h.cache.SetQualityMetrics(c.Request.Context(), wsID, projectID, result)
	}
	c.Header("X-Cache", "MISS")
	c.JSON(http.StatusOK, result)
}

// GetDORA DORA 四指标（带缓存）。
func (h *CachedHandler) GetDORA(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	projectID := c.GetInt64(middleware.CtxProjectID)

	if h.cache != nil {
		if cached, ok := h.cache.GetDORA(c.Request.Context(), wsID, projectID); ok {
			c.Header("X-Cache", "HIT")
			c.JSON(http.StatusOK, cached)
			return
		}
	}

	result, err := h.d.Svc.GetDORA(c.Request.Context(), wsID, projectID)
	if err != nil {
		writeErr(c, err)
		return
	}
	if h.cache != nil {
		h.cache.SetDORA(c.Request.Context(), wsID, projectID, result)
	}
	c.Header("X-Cache", "MISS")
	c.JSON(http.StatusOK, result)
}

// GetResourceLoad 资源负载（带缓存，短 TTL）。
func (h *CachedHandler) GetResourceLoad(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	projectID := c.GetInt64(middleware.CtxProjectID)

	if h.cache != nil {
		if cached, ok := h.cache.GetResourceLoad(c.Request.Context(), wsID, projectID); ok {
			c.Header("X-Cache", "HIT")
			c.JSON(http.StatusOK, cached)
			return
		}
	}

	// 复用原始 handler 的查询逻辑
	c.Set("Cache-MISS", true)
	h.Handler.GetResourceLoad(c)
	if c.GetBool("Cache-MISS") && h.cache != nil && c.Writer.Status() == http.StatusOK {
		// 注意：这里无法直接读取 response body
		// 生产环境建议重构为 Service 层缓存或在中间件层面处理
	}
}

// ListSnapshots 指标快照列表（带缓存）。
func (h *CachedHandler) ListSnapshots(c *gin.Context) {
	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	projectID := c.GetInt64(middleware.CtxProjectID)
	metric := c.Query("metric")

	if h.cache != nil {
		if cached, ok := h.cache.GetSnapshots(c.Request.Context(), wsID, projectID, metric); ok {
			c.Header("X-Cache", "HIT")
			c.JSON(http.StatusOK, cached)
			return
		}
	}

	// 直查 DB
	query := `SELECT metric, value, dimensions, snapshot_date FROM metric_snapshots
			  WHERE workspace_id = $1 AND project_id = $2`
	args := []any{wsID, projectID}
	if metric != "" {
		query += ` AND metric = $3`
		args = append(args, metric)
	}
	query += ` ORDER BY snapshot_date DESC LIMIT 90`

	r, err := h.d.Svc.db.Query(c.Request.Context(), query, args...)
	if err != nil {
		writeErr(c, err)
		return
	}
	defer r.Close()

	var results []map[string]any
	for r.Next() {
		var m string
		var v float64
		var dims []byte
		var d string
		if err := r.Scan(&m, &v, &dims, &d); err != nil {
			continue
		}
		results = append(results, map[string]any{
			"metric":        m,
			"value":         v,
			"dims":          dims,
			"snapshot_date": d,
		})
	}

	if h.cache != nil {
		h.cache.SetSnapshots(c.Request.Context(), wsID, projectID, metric, results)
	}
	c.Header("X-Cache", "MISS")
	c.JSON(http.StatusOK, results)
}
