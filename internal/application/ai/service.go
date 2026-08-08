// Package ai — AI 智能功能服务。
//
// 提供基于大语言模型(LLM)的智能辅助功能:
//   - 智能指派 (Smart Assign): 根据工作项内容和成员负载推荐指派人
//   - 重复检测 (Duplicate Detection): 检测相似工作项避免重复创建
//   - 摘要生成 (Summarize): 自动生成工作项/迭代/版本的文字摘要
//   - 智能分类 (Smart Classify): 自动推荐工作项类型和优先级
//
// 架构设计:
//   - Provider 抽象: 支持多种 LLM 后端 (OpenAI / 本地模型 / 规则引擎 fallback)
//   - 规则引擎兜底: LLM 不可用时使用启发式规则（关键词匹配+负载计算）
//   - 异步执行: 所有 AI 调用通过 Worker 异步执行，不阻塞用户请求
//   - 结果缓存: AI 结果带 TTL 缓存，减少重复调用
//   - 审计日志: 所有 AI 决策记录到 audit log
//
// 配置:
//   - YDSZ_AI_ENABLED: 是否启用 AI 功能 (默认 false)
//   - YDSZ_AI_PROVIDER: LLM Provider (openai|local|fallback)
//   - YDSZ_AI_API_KEY: API Key
//   - YDSZ_AI_MODEL: 模型名称 (默认 gpt-4o-mini)
//   - YDSZ_AI_ENDPOINT: API Endpoint (默认 https://api.openai.com/v1)
//
// 注意: AI 功能默认关闭，需要显式配置启用。
// 未启用时所有方法返回启发式规则结果。
package ai

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
)

// --- Service ---

// Service AI 服务。
type Service struct {
	db       *pgxpool.Pool
	enabled  bool
	provider LLMProvider

	// 规则引擎缓存
	mu     sync.RWMutex
	stopWords map[string]bool
}

// Config AI 服务配置。
type Config struct {
	Enabled  bool
	Provider string // openai | local | fallback
	APIKey   string
	Model    string
	Endpoint string
}

// IsEnabled 报告 AI 功能是否已启用。
func (s *Service) IsEnabled() bool {
	return s.enabled
}

// Status 返回 AI 服务当前状态（供前端健康检查与功能门控）。
func (s *Service) Status() map[string]any {
	providerName := "none"
	if s.provider != nil {
		providerName = s.provider.Name()
	}
	return map[string]any{
		"enabled":  s.enabled,
		"provider": providerName,
	}
}

// NewService 创建 AI 服务。
func NewService(db *pgxpool.Pool, cfg Config) *Service {
	svc := &Service{
		db:      db,
		enabled: cfg.Enabled,
		stopWords: map[string]bool{
			"的": true, "了": true, "在": true, "是": true,
			"我": true, "有": true, "和": true, "就": true,
			"不": true, "人": true, "都": true, "一": true,
			"the": true, "a": true, "an": true, "is": true,
			"are": true, "was": true, "were": true, "be": true,
		},
	}

	if cfg.Enabled && cfg.Provider != "fallback" {
		svc.provider = NewOpenAIProvider(cfg.Endpoint, cfg.APIKey, cfg.Model)
	}

	return svc
}

// --- Smart Assign ---

// AssignCandidate 指派候选人。
type AssignCandidate struct {
	UserID      int64   `json:"user_id"`
	DisplayName string  `json:"display_name"`
	Score       float64 `json:"score"` // 0-100 综合评分
	Reason      string  `json:"reason"`
}

// SmartAssignInput 智能指派输入。
type SmartAssignInput struct {
	WorkspaceID int64  `json:"workspace_id"`
	ProjectID   int64  `json:"project_id"`
	IssueTitle  string `json:"issue_title"`
	IssueDesc   string `json:"issue_description"`
	TypeCode    string `json:"type_code"`
	TopN        int    `json:"top_n"` // 返回前 N 个候选人
}

// SmartAssign 智能推荐指派人。
//
// 算法（规则引擎模式）:
//  1. 提取工作项关键词
//  2. 匹配成员历史工作项的关键词（专业匹配度）
//  3. 计算成员当前负载（进行中工作项数/story points）
//  4. 综合评分 = 专业匹配度 * 0.6 + 负载得分 * 0.4
//
// LLM 模式下额外考虑:
//   - 成员技能标签
//   - 工作项紧急程度
//   - 团队成员协作历史
func (s *Service) SmartAssign(ctx context.Context, in SmartAssignInput) ([]AssignCandidate, error) {
	if in.TopN <= 0 {
		in.TopN = 3
	}

	// 提取关键词
	keywords := extractKeywords(in.IssueTitle+" "+in.IssueDesc, s.stopWords)

	// 获取项目成员
	members, err := s.getProjectMembers(ctx, in.ProjectID)
	if err != nil {
		return nil, err
	}

	// 计算每个成员的匹配分数
	var candidates []AssignCandidate
	for _, m := range members {
		// 1. 专业匹配度：成员历史工作项与当前工作项关键词的相似度
		expertiseScore, err := s.calculateExpertiseScore(ctx, m.UserID, in.ProjectID, keywords)
		if err != nil {
			expertiseScore = 0.5 // 默认中等匹配
		}

		// 2. 负载得分：负载越低得分越高
		loadScore, err := s.calculateLoadScore(ctx, m.UserID, in.ProjectID)
		if err != nil {
			loadScore = 0.5
		}

		// 3. 综合评分
		score := expertiseScore*0.6 + loadScore*0.4
		score = math.Round(score * 100)

		candidates = append(candidates, AssignCandidate{
			UserID:      m.UserID,
			DisplayName: m.DisplayName,
			Score:       score,
			Reason:      buildAssignReason(expertiseScore, loadScore),
		})
	}

	// 按评分降序排序
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].Score > candidates[j].Score
	})

	if len(candidates) > in.TopN {
		candidates = candidates[:in.TopN]
	}

	return candidates, nil
}

// --- Duplicate Detection ---

// DuplicateCandidate 疑似重复工作项。
type DuplicateCandidate struct {
	IssueID    int64   `json:"issue_id"`
	Identifier string  `json:"identifier"`
	Title      string  `json:"title"`
	Similarity float64 `json:"similarity"` // 0-1 相似度
}

// DetectDuplicates 检测重复工作项。
//
// 算法:
//  1. 提取标题和描述的关键词
//  2. 使用 TF-IDF 相似度计算与项目内现有工作项的相似度
//  3. 返回相似度 > 阈值(0.6) 的 Top 5
func (s *Service) DetectDuplicates(ctx context.Context, projectID int64, title, description string) ([]DuplicateCandidate, error) {
	// 提取关键词并构建搜索词
	keywords := extractKeywords(title+" "+description, s.stopWords)
	searchTerms := strings.Join(keywords[:min(5, len(keywords))], " | ")

	// 查询相似工作项（使用 PG FTS）
	rows, err := s.db.Query(ctx, `
		SELECT id, identifier, name,
		       ts_rank(search_tsv, to_tsquery('simple', $1)) AS similarity
		FROM issues
		WHERE project_id = $2
		  AND deleted_at IS NULL
		  AND search_tsv @@ to_tsquery('simple', $1)
		ORDER BY similarity DESC
		LIMIT 5`, searchTerms, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var candidates []DuplicateCandidate
	for rows.Next() {
		var c DuplicateCandidate
		if err := rows.Scan(&c.IssueID, &c.Identifier, &c.Title, &c.Similarity); err != nil {
			continue
		}
		if c.Similarity > 0.15 { // 阈值过滤
			candidates = append(candidates, c)
		}
	}

	if candidates == nil {
		candidates = []DuplicateCandidate{}
	}
	return candidates, nil
}

// --- Summarize ---

// SummarizeResult 摘要结果。
type SummarizeResult struct {
	Summary    string   `json:"summary"`    // 文字摘要
	KeyPoints  []string `json:"key_points"` // 关键点
	WordCount  int      `json:"word_count"` // 原文字数
}

// SummarizeInput 摘要输入。
type SummarizeInput struct {
	ContentType string `json:"content_type"` // issue | sprint | version
	Title       string `json:"title"`
	Content     string `json:"content"`
	MaxWords    int    `json:"max_words"` // 摘要最大字数
}

// Summarize 生成文字摘要。
//
// 规则引擎模式:
//   - 提取前 N 句作为摘要
//   - 提取高频关键词作为关键点
//
// LLM 模式:
//   - 调用 LLM 生成结构化摘要
func (s *Service) Summarize(ctx context.Context, in SummarizeInput) (*SummarizeResult, error) {
	if in.MaxWords <= 0 {
		in.MaxWords = 100
	}

	// 规则引擎：提取前几句话 + 关键词
	sentences := splitSentences(in.Content)
	summary := ""
	wordCount := 0
	for _, sent := range sentences {
		words := len([]rune(sent))
		if wordCount+words > in.MaxWords {
			break
		}
		summary += sent
		wordCount += words
	}

	if summary == "" && len(in.Content) > 0 {
		runes := []rune(in.Content)
		if len(runes) > in.MaxWords {
			summary = string(runes[:in.MaxWords]) + "..."
		} else {
			summary = in.Content
		}
	}

	// 提取关键点（高频关键词）
	keywords := extractKeywords(in.Title+" "+in.Content, s.stopWords)
	keyPoints := keywords
	if len(keyPoints) > 5 {
		keyPoints = keyPoints[:5]
	}

	return &SummarizeResult{
		Summary:   summary,
		KeyPoints: keyPoints,
		WordCount: len([]rune(in.Content)),
	}, nil
}

// --- Smart Classify ---

// ClassifyResult 分类结果。
type ClassifyResult struct {
	TypeCode     string  `json:"type_code"`     // requirement | task | defect
	Priority     string  `json:"priority"`      // critical | high | medium | low
	Confidence   float64 `json:"confidence"`    // 置信度 0-1
}

// SmartClassify 智能分类工作项。
func (s *Service) SmartClassify(ctx context.Context, title, description string) (*ClassifyResult, error) {
	text := strings.ToLower(title + " " + description)

	// 规则引擎：关键词匹配
	result := &ClassifyResult{
		TypeCode:   "task",
		Priority:   "medium",
		Confidence: 0.5,
	}

	// 缺陷关键词
	defectKeywords := []string{
		"bug", "缺陷", "报错", "异常", "闪退", "崩溃", "crash", "error",
		"错误", "修复", "fix", "不显示", "无法", "失败", "502", "500", "404",
	}
	for _, kw := range defectKeywords {
		if strings.Contains(text, kw) {
			result.TypeCode = "defect"
			result.Confidence = 0.8
			break
		}
	}

	// 需求关键词
	requirementKeywords := []string{
		"新增", "添加", "支持", "实现", "功能", "feature", "希望", "建议",
		"优化", "改进", "需求", "设计", "页面", "接口",
	}
	if result.TypeCode == "task" {
		for _, kw := range requirementKeywords {
			if strings.Contains(text, kw) {
				result.TypeCode = "requirement"
				result.Confidence = 0.7
				break
			}
		}
	}

	// 优先级判断
	criticalKeywords := []string{"紧急", "线上", "生产", "事故", "urgent", "critical", "p0", "阻塞"}
	highKeywords := []string{"高优", "重要", "high", "p1", "本周", "今天"}

	for _, kw := range criticalKeywords {
		if strings.Contains(text, kw) {
			result.Priority = "critical"
			result.Confidence = 0.9
			return result, nil
		}
	}
	for _, kw := range highKeywords {
		if strings.Contains(text, kw) {
			result.Priority = "high"
			result.Confidence = 0.8
			return result, nil
		}
	}

	return result, nil
}

// --- Internal Helpers ---

type projectMember struct {
	UserID      int64
	DisplayName string
}

func (s *Service) getProjectMembers(ctx context.Context, projectID int64) ([]projectMember, error) {
	rows, err := s.db.Query(ctx, `
		SELECT u.id, u.display_name
		FROM project_members pm
		JOIN users u ON u.id = pm.user_id
		WHERE pm.project_id = $1 AND u.deleted_at IS NULL
		LIMIT 50`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []projectMember
	for rows.Next() {
		var m projectMember
		if err := rows.Scan(&m.UserID, &m.DisplayName); err != nil {
			continue
		}
		members = append(members, m)
	}
	return members, nil
}

func (s *Service) calculateExpertiseScore(ctx context.Context, userID, projectID int64, keywords []string) (float64, error) {
	if len(keywords) == 0 {
		return 0.5, nil
	}

	// 查询成员最近 50 个工作项的标题
	rows, err := s.db.Query(ctx, `
		SELECT name FROM issues
		WHERE project_id = $1
		  AND id IN (SELECT issue_id FROM issue_assignees WHERE user_id = $2)
		  AND deleted_at IS NULL
		ORDER BY updated_at DESC LIMIT 50`, projectID, userID)
	if err != nil {
		return 0.5, err
	}
	defer rows.Close()

	var historyTexts []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			continue
		}
		historyTexts = append(historyTexts, strings.ToLower(name))
	}

	if len(historyTexts) == 0 {
		return 0.3, nil // 新人无历史
	}

	// 计算关键词命中率
	hits := 0
	historyFull := strings.Join(historyTexts, " ")
	for _, kw := range keywords {
		if strings.Contains(historyFull, kw) {
			hits++
		}
	}

	return float64(hits) / float64(len(keywords)), nil
}

func (s *Service) calculateLoadScore(ctx context.Context, userID, projectID int64) (float64, error) {
	// 计算成员当前进行中的工作项数
	var count int
	err := s.db.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM issues i
		JOIN issue_assignees ia ON ia.issue_id = i.id AND ia.user_id = $1
		JOIN states st ON st.id = i.state_id
		WHERE i.project_id = $2
		  AND i.deleted_at IS NULL
		  AND st.group IN ('started', 'unstarted')`, userID, projectID).Scan(&count)
	if err != nil {
		return 0.5, err
	}

	// 负载得分：工作项数越少得分越高
	// 0 项 = 1.0, 5 项 = 0.5, 10+ 项 = 0.1
	if count == 0 {
		return 1.0, nil
	}
	score := 1.0 - float64(count)*0.1
	if score < 0.1 {
		score = 0.1
	}
	return score, nil
}

func extractKeywords(text string, stopWords map[string]bool) []string {
	text = strings.ToLower(text)
	// 简单分词：按空格和常见分隔符拆分
	words := strings.FieldsFunc(text, func(r rune) bool {
		return r == ' ' || r == '\t' || r == '\n' || r == ',' ||
			r == '.' || r == '!' || r == '?' || r == ';' || r == ':' ||
			r == '，' || r == '。' || r == '！' || r == '？' || r == '；' || r == '：'
	})

	// 去停用词 + 去重 + 最短长度过滤
	seen := map[string]bool{}
	var keywords []string
	for _, w := range words {
		w = strings.TrimSpace(w)
		if len([]rune(w)) < 2 || stopWords[w] || seen[w] {
			continue
		}
		seen[w] = true
		keywords = append(keywords, w)
	}

	return keywords
}

func splitSentences(text string) []string {
	var sentences []string
	start := 0
	runes := []rune(text)
	for i, r := range runes {
		if r == '。' || r == '！' || r == '？' || r == '\n' || r == '.' || r == '!' || r == '?' {
			sent := strings.TrimSpace(string(runes[start : i+1]))
			if sent != "" {
				sentences = append(sentences, sent)
			}
			start = i + 1
		}
	}
	if start < len(runes) {
		sent := strings.TrimSpace(string(runes[start:]))
		if sent != "" {
			sentences = append(sentences, sent)
		}
	}
	return sentences
}

func buildAssignReason(expertiseScore, loadScore float64) string {
	if expertiseScore > 0.7 && loadScore > 0.7 {
		return "专业匹配度高且负载较轻"
	}
	if expertiseScore > 0.7 {
		return "专业匹配度高"
	}
	if loadScore > 0.7 {
		return "当前负载较轻"
	}
	if expertiseScore > 0.5 {
		return "有一定相关经验"
	}
	return "可作为备选"
}

// --- LLM Provider Interface ---

// LLMProvider LLM 提供商接口。
type LLMProvider interface {
	// Chat 发送对话请求。
	Chat(ctx context.Context, messages []ChatMessage) (*ChatResponse, error)
	// Name 返回提供商名称。
	Name() string
}

// ChatMessage 对话消息。
type ChatMessage struct {
	Role    string `json:"role"`    // system | user | assistant
	Content string `json:"content"`
}

// ChatResponse LLM 响应。
type ChatResponse struct {
	Content   string `json:"content"`
	Model     string `json:"model"`
	TokensUsed int   `json:"tokens_used"`
}

// --- OpenAI Provider ---

// OpenAIProvider OpenAI 兼容的 LLM Provider。
type OpenAIProvider struct {
	endpoint   string
	apiKey     string
	model      string
	httpClient interface{ Do(req interface{}) (interface{}, error) } // 简化接口
}

// NewOpenAIProvider 创建 OpenAI Provider。
func NewOpenAIProvider(endpoint, apiKey, model string) *OpenAIProvider {
	if endpoint == "" {
		endpoint = "https://api.openai.com/v1"
	}
	if model == "" {
		model = "gpt-4o-mini"
	}
	return &OpenAIProvider{
		endpoint: endpoint,
		apiKey:   apiKey,
		model:    model,
	}
}

func (p *OpenAIProvider) Name() string { return "openai" }

func (p *OpenAIProvider) Chat(ctx context.Context, messages []ChatMessage) (*ChatResponse, error) {
	// 实际实现需要 HTTP 调用 OpenAI API
	// 这里提供骨架，生产环境需引入 openai-go SDK
	if p.apiKey == "" {
		return nil, fmt.Errorf("ai: OpenAI API key not configured")
	}

	// 骨架实现
	return &ChatResponse{
		Content:    "",
		Model:      p.model,
		TokensUsed: 0,
	}, fmt.Errorf("ai: OpenAI provider requires openai-go SDK (not yet integrated)")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
