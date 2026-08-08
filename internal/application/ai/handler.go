// Package ai — AI 智能功能 HTTP handlers。
//
// 暴露项目级 AI 辅助端点：智能指派、重复检测、智能分类、摘要生成。
// 所有端点遵循统一响应格式，未启用时返回 501。
package ai

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/njydsz/ydsz-plane/internal/interfaces/middleware"
	"github.com/njydsz/ydsz-plane/pkg/errs"
)

// HandlerDeps AI handler 依赖。
type HandlerDeps struct {
	AiSvc *Service
}

// Handler Gin handler 集合。
type Handler struct {
	d *HandlerDeps
}

// NewHandler 构造 handler。
func NewHandler(d *HandlerDeps) *Handler {
	return &Handler{d: d}
}

// Register 注册 AI 路由（项目级）。
func (h *Handler) Register(r *gin.RouterGroup) {
	r.GET("/status", h.Status)
	r.POST("/smart-assign", h.SmartAssign)
	r.POST("/detect-duplicates", h.DetectDuplicates)
	r.POST("/classify", h.SmartClassify)
	r.POST("/summarize", h.Summarize)
	r.POST("/assist", h.WritingAssist)
	r.POST("/rewrite", h.RewriteText)
	r.POST("/fix-grammar", h.FixGrammar)
}

// Status godoc
//
//	@Summary		AI 功能状态
//	@Description	返回 AI 功能的启用状态和当前 Provider
//	@Tags			ai
//	@Produce		json
//	@Success		200	{object}	map[string]any
//	@Router			/projects/{project_id}/ai/status [get]
func (h *Handler) Status(c *gin.Context) {
	c.JSON(http.StatusOK, h.d.AiSvc.Status())
}

// SmartAssign godoc
//
//	@Summary		智能指派推荐
//	@Description	根据工作项内容与成员负载推荐指派人
//	@Tags			ai
//	@Accept			json
//	@Produce		json
//	@Param			body	body		SmartAssignInput	true	"指派输入"
//	@Success		200		{array}		AssignCandidate
//	@Router			/projects/{project_id}/ai/smart-assign [post]
func (h *Handler) SmartAssign(c *gin.Context) {
	if !h.d.AiSvc.IsEnabled() {
		middleware.AbortWithError(c, errs.ErrNotImplemented.WithDetails(errs.FieldDetail{
			Field:  "ai",
			Reason: "AI 功能未启用，请配置 YDSZ_AI_ENABLED=true",
		}))
		return
	}

	var in SmartAssignInput
	if err := c.ShouldBindJSON(&in); err != nil {
		middleware.AbortWithError(c, errs.ErrValidation.WithDetails(fieldDetail(err)))
		return
	}

	wsID := c.GetInt64(middleware.CtxWorkspaceID)
	projectID := c.GetInt64(middleware.CtxProjectID)
	in.WorkspaceID = wsID
	in.ProjectID = projectID

	candidates, err := h.d.AiSvc.SmartAssign(c.Request.Context(), in)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, candidates)
}

// DetectDuplicates godoc
//
//	@Summary		重复工作项检测
//	@Description	检测项目内与输入标题/描述相似的工作项
//	@Tags			ai
//	@Accept			json
//	@Produce		json
//	@Param			body	body		detectDuplicatesRequest	true	"检测输入"
//	@Success		200		{array}		DuplicateCandidate
//	@Router			/projects/{project_id}/ai/detect-duplicates [post]
func (h *Handler) DetectDuplicates(c *gin.Context) {
	if !h.d.AiSvc.IsEnabled() {
		middleware.AbortWithError(c, errs.ErrNotImplemented.WithDetails(errs.FieldDetail{
			Field:  "ai",
			Reason: "AI 功能未启用，请配置 YDSZ_AI_ENABLED=true",
		}))
		return
	}

	var req detectDuplicatesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithError(c, errs.ErrValidation.WithDetails(fieldDetail(err)))
		return
	}

	projectID := c.GetInt64(middleware.CtxProjectID)
	candidates, err := h.d.AiSvc.DetectDuplicates(c.Request.Context(), projectID, req.Title, req.Description)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, candidates)
}

// SmartClassify godoc
//
//	@Summary		智能分类推荐
//	@Description	自动推荐工作项类型和优先级
//	@Tags			ai
//	@Accept			json
//	@Produce		json
//	@Param			body	body		classifyRequest	true	"分类输入"
//	@Success		200		{object}	ClassifyResult
//	@Router			/projects/{project_id}/ai/classify [post]
func (h *Handler) SmartClassify(c *gin.Context) {
	if !h.d.AiSvc.IsEnabled() {
		middleware.AbortWithError(c, errs.ErrNotImplemented.WithDetails(errs.FieldDetail{
			Field:  "ai",
			Reason: "AI 功能未启用，请配置 YDSZ_AI_ENABLED=true",
		}))
		return
	}

	var req classifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithError(c, errs.ErrValidation.WithDetails(fieldDetail(err)))
		return
	}

	result, err := h.d.AiSvc.SmartClassify(c.Request.Context(), req.Title, req.Description)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

// Summarize godoc
//
//	@Summary		生成文字摘要
//	@Description	为工作项/迭代/版本内容生成文字摘要
//	@Tags			ai
//	@Accept			json
//	@Produce		json
//	@Param			body	body		SummarizeInput	true	"摘要输入"
//	@Success		200		{object}	SummarizeResult
//	@Router			/projects/{project_id}/ai/summarize [post]
func (h *Handler) Summarize(c *gin.Context) {
	if !h.d.AiSvc.IsEnabled() {
		middleware.AbortWithError(c, errs.ErrNotImplemented.WithDetails(errs.FieldDetail{
			Field:  "ai",
			Reason: "AI 功能未启用，请配置 YDSZ_AI_ENABLED=true",
		}))
		return
	}

	var in SummarizeInput
	if err := c.ShouldBindJSON(&in); err != nil {
		middleware.AbortWithError(c, errs.ErrValidation.WithDetails(fieldDetail(err)))
		return
	}

	result, err := h.d.AiSvc.Summarize(c.Request.Context(), in)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

// --- Request DTOs ---

type detectDuplicatesRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type classifyRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type writingAssistRequest struct {
	Context   string `json:"context"`
	FullText  string `json:"full_text"`
	Language  string `json:"language"`
	Style     string `json:"style"`
	MaxTokens int    `json:"max_tokens"`
}

type rewriteRequest struct {
	Text      string `json:"text"`
	Style     string `json:"style"`
	Language  string `json:"language"`
	IssueType string `json:"issue_type"`
}

type fixGrammarRequest struct {
	Text     string `json:"text"`
	Language string `json:"language"`
}

// WritingAssist godoc
//
//	@Summary		AI 续写
//	@Description	根据上下文智能续写文本（规则引擎兜底，LLM 需配置）
//	@Tags			ai
//	@Accept			json
//	@Produce		json
//	@Param			body	body		writingAssistRequest	true	"续写输入"
//	@Success		200		{object}	WritingAssistResult
//	@Router			/projects/{project_id}/ai/assist [post]
func (h *Handler) WritingAssist(c *gin.Context) {
	var req writingAssistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithError(c, errs.ErrValidation.WithDetails(fieldDetail(err)))
		return
	}
	result, err := h.d.AiSvc.WritingAssist(c.Request.Context(), WritingAssistInput{
		Context: req.Context, FullText: req.FullText, Language: req.Language,
		Style: req.Style, MaxTokens: req.MaxTokens,
	})
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

// RewriteText godoc
//
//	@Summary		文本改写
//	@Description	对选中文本进行风格改写（formal / concise / fluent / expand）
//	@Tags			ai
//	@Accept			json
//	@Produce		json
//	@Param			body	body		rewriteRequest	true	"改写输入"
//	@Success		200		{object}	RewriteResult
//	@Router			/projects/{project_id}/ai/rewrite [post]
func (h *Handler) RewriteText(c *gin.Context) {
	var req rewriteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithError(c, errs.ErrValidation.WithDetails(fieldDetail(err)))
		return
	}
	result, err := h.d.AiSvc.RewriteText(c.Request.Context(), RewriteInput{
		Text: req.Text, Style: req.Style, Language: req.Language, IssueType: req.IssueType,
	})
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

// FixGrammar godoc
//
//	@Summary		语法纠错
//	@Description	检测并修正语法、拼写、标点问题
//	@Tags			ai
//	@Accept			json
//	@Produce		json
//	@Param			body	body		fixGrammarRequest	true	"纠错输入"
//	@Success		200		{object}	FixGrammarResult
//	@Router			/projects/{project_id}/ai/fix-grammar [post]
func (h *Handler) FixGrammar(c *gin.Context) {
	var req fixGrammarRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithError(c, errs.ErrValidation.WithDetails(fieldDetail(err)))
		return
	}
	result, err := h.d.AiSvc.FixGrammar(c.Request.Context(), FixGrammarInput{
		Text: req.Text, Language: req.Language,
	})
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

// --- Helpers ---

func fieldDetail(err error) errs.FieldDetail {
	return errs.FieldDetail{Field: "body", Reason: err.Error()}
}

func writeErr(c *gin.Context, err error) {
	var appErr *errs.AppError
	if errs.As(err, &appErr) {
		middleware.AbortWithError(c, appErr)
		return
	}
	middleware.AbortWithError(c, errs.ErrInternal)
}
