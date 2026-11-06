// Package pql — 应用服务与 HTTP handlers。
package pql

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/njydsz/ydsz-plane/internal/interfaces/middleware"
)

// Service PQL 应用服务。
type Service struct{}

// NewService 创建服务。
func NewService() *Service { return &Service{} }

// Handler Gin handler 集合。
type Handler struct {
	svc *Service
}

// NewHandler 构造 handler。
func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// Register 注册路由。
func (h *Handler) Register(r *gin.RouterGroup) {
	g := r.Group("/pql")
	g.POST("/parse", h.Parse)
	g.POST("/validate", h.Validate)
	g.GET("/fields", h.ListFields)
	g.POST("/explain", h.Explain)
}

// ParseResult 解析结果。
type ParseResult struct {
	Valid       bool             `json:"valid"`
	Query       *Query           `json:"query,omitempty"`
	WhereSQL    string           `json:"where_sql,omitempty"`
	Args        []any            `json:"args,omitempty"`
	SortColumn  string           `json:"sort_column,omitempty"`
	SortDir     string           `json:"sort_dir,omitempty"`
	Limit       int              `json:"limit"`
	Offset      int              `json:"offset"`
	Error       string           `json:"error,omitempty"`
}

// Parse godoc
//
//	@Summary		解析 PQL 查询
//	@Tags			pql
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	ParseResult
//	@Router			/pql/parse [post]
func (h *Handler) Parse(c *gin.Context) {
	var body struct {
		Query string `json:"query"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	res := h.svc.Parse(body.Query, c.GetInt64(middleware.CtxUserID))
	c.JSON(http.StatusOK, res)
}

// Validate godoc
//
//	@Summary		校验 PQL 查询（仅返回合法性）
//	@Tags			pql
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	ParseResult
//	@Router			/pql/validate [post]
func (h *Handler) Validate(c *gin.Context) {
	var body struct {
		Query string `json:"query"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	res := h.svc.Parse(body.Query, c.GetInt64(middleware.CtxUserID))
	c.JSON(http.StatusOK, res)
}

// Explain godoc
//
//	@Summary		解释 PQL 查询（含生成的 SQL 与参数）
//	@Tags			pql
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	ParseResult
//	@Router			/pql/explain [post]
func (h *Handler) Explain(c *gin.Context) {
	var body struct {
		Query string `json:"query"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	res := h.svc.Parse(body.Query, c.GetInt64(middleware.CtxUserID))
	c.JSON(http.StatusOK, res)
}

// ListFields godoc
//
//	@Summary		PQL 可用字段列表
//	@Tags			pql
//	@Produce		json
//	@Success		200	{object}	map[string]any
//	@Router			/pql/fields [get]
func (h *Handler) ListFields(c *gin.Context) {
	fields := make([]map[string]any, 0, len(DefaultFieldMappings))
	for name, m := range DefaultFieldMappings {
		f := map[string]any{"name": name, "column": m.Column, "type": m.Kind}
		if len(m.EnumSet) > 0 {
			vals := make([]string, 0, len(m.EnumSet))
			for v := range m.EnumSet {
				vals = append(vals, v)
			}
			f["values"] = vals
		}
		fields = append(fields, f)
	}
	c.JSON(http.StatusOK, gin.H{"fields": fields, "sort_fields": AllowedSortFields})
}

// --- Service ---

// Parse 解析并编译 PQL 查询。
func (s *Service) Parse(input string, meUserID int64) *ParseResult {
	res := &ParseResult{Limit: 100}
	q, err := Parse(input)
	if err != nil {
		res.Valid = false
		res.Error = err.Error()
		return res
	}
	res.Valid = true
	res.Query = q
	res.Limit = q.Limit
	res.Offset = q.Offset
	res.SortDir = q.SortDir

	// 编译（无用户映射时，@user 会报错；此处仅做语法/字段校验）
	compiler := NewCompiler()
	compiled, cerr := compiler.Compile(q, meUserID, nil)
	if cerr != nil {
		// @user 未解析属于预期；标记为合法但记录提示
		if q.Where != nil && containsAtUser(q.Where) {
			res.WhereSQL = ""
			res.Args = nil
			res.SortColumn = ""
			return res
		}
		res.Valid = false
		res.Error = cerr.Error()
		return res
	}
	res.WhereSQL = compiled.WhereSQL
	res.Args = compiled.Args
	res.SortColumn = compiled.SortColumn
	return res
}

func containsAtUser(e Expr) bool {
	switch n := e.(type) {
	case *ComparisonExpr:
		return n.AtUser != ""
	case *LogicalExpr:
		return containsAtUser(n.Left) || containsAtUser(n.Right)
	case *NotExpr:
		return containsAtUser(n.Inner)
	}
	return false
}

var _ = strconv.Itoa
