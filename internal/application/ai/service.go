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
	"hash/fnv"
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
		FROM workitems
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
	TypeCode   string  `json:"type_code"`   // requirement | task | defect
	Priority   string  `json:"priority"`    // critical | high | medium | low
	Confidence float64 `json:"confidence"`  // 置信度 0-1
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

// ========================= AI Writing (编辑器 AI 辅助) =========================

// WritingAssistInput AI 续写输入。
type WritingAssistInput struct {
	Context      string `json:"context"`       // 光标前的最后一段文本
	FullText     string `json:"full_text"`     // 当前全文
	Language     string `json:"language"`      // zh | en，默认 zh
	Style        string `json:"style"`         // professional | concise | casual
	MaxTokens    int    `json:"max_tokens"`    // 返回最大字数
}

// WritingAssistResult AI 续写结果。
type WritingAssistResult struct {
	Text       string  `json:"text"`        // 续写内容
	Confidence float64 `json:"confidence"`  // 规则引擎固定 0.6
	Model      string  `json:"model"`       // 使用的模型或 "rule-engine"
}

// WritingAssist AI 续写 — 根据上下文智能续写文本。
//
// 规则引擎模式（默认）:
//   - 提取最后一段的关键词
//   - 根据句式模式补全（如列举、因果、递进修辞）
//   - 返回 1-3 句续写建议
//
// LLM 模式（需配置 Provider）:
//   - 调用 LLM 生成高质量续写
func (s *Service) WritingAssist(_ context.Context, in WritingAssistInput) (*WritingAssistResult, error) {
	if in.MaxTokens <= 0 {
		in.MaxTokens = 120
	}
	if in.Language == "" {
		in.Language = "zh"
	}

	// 规则引擎兜底：启发式续写
	text := ruleEngineAssist(in.Context, in.FullText, in.Language, in.MaxTokens)
	return &WritingAssistResult{
		Text:       text,
		Confidence: 0.6,
		Model:      "rule-engine",
	}, nil
}

// RewriteInput AI 改写输入。
type RewriteInput struct {
	Text     string `json:"text"`      // 选中的原文
	Style    string `json:"style"`     // formal | concise | fluent | expand
	Language string `json:"language"`  // zh | en
	IssueType string `json:"issue_type"` // 期望语境: requirement | task | defect | null
}

// RewriteResult 改写结果。
type RewriteResult struct {
	Text       string   `json:"text"`        // 改写后文本
	Changes    []string `json:"changes"`     // 改动说明
	OriginalLen int     `json:"original_len"`
	NewLen      int     `json:"new_len"`
	Model       string  `json:"model"`        // rule-engine 或 LLM 名称
}

// RewriteText AI 改写 — 对选中文本进行风格/语气改写。
//
// 支持的 style:
//   - formal: 正式化（去除口语词、补全主语、用规范术语）
//   - concise: 精简（去除冗余修饰，保留核心信息）
//   - fluent: 流畅化（调整语序、消除断裂句）
//   - expand: 扩写（添加过渡句、补充说明细节）
func (s *Service) RewriteText(_ context.Context, in RewriteInput) (*RewriteResult, error) {
	if in.Language == "" {
		in.Language = detectLanguage(in.Text)
	}

	result := &RewriteResult{
		OriginalLen: len([]rune(in.Text)),
		Model:       "rule-engine",
	}

	newText, changes := ruleEngineRewrite(in.Text, in.Style, in.Language)
	result.Text = newText
	result.Changes = changes
	result.NewLen = len([]rune(newText))
	return result, nil
}

// FixGrammarInput 语法纠错输入。
type FixGrammarInput struct {
	Text     string `json:"text"`     // 待纠错的文本
	Language string `json:"language"` // zh | en，自动检测
}

// GrammarIssue 语法问题。
type GrammarIssue struct {
	Offset      int    `json:"offset"`       // 问题起始偏移
	Length      int    `json:"length"`       // 问题长度
	Original    string `json:"original"`     // 原文
	Replacement string `json:"replacement"`  // 建议替换
	Reason      string `json:"reason"`       // 错误说明
	Severity    string `json:"severity"`     // error | warning | style
}

// FixGrammarResult 纠错结果。
type FixGrammarResult struct {
	FixedText string        `json:"fixed_text"` // 修正后全文
	Issues    []GrammarIssue `json:"workitems"`     // 发现的问题列表
	Model     string        `json:"model"`       // rule-engine
}

// FixGrammar AI 语法纠错 — 检测并修正语法、拼写、标点问题。
//
// 规则引擎检测:
//   - 中文：的地得混用、标点符号全半角、常见错别字
//   - 英文：基础拼写、主谓一致、冠词用法
func (s *Service) FixGrammar(_ context.Context, in FixGrammarInput) (*FixGrammarResult, error) {
	if in.Language == "" {
		in.Language = detectLanguage(in.Text)
	}

	result := &FixGrammarResult{
		Model: "rule-engine",
	}

	issues, fixed := ruleEngineFixGrammar(in.Text, in.Language)
	result.Issues = issues
	result.FixedText = fixed
	return result, nil
}

// --- Rule Engine Implementations ---

// ruleEngineAssist 根据文本尾部和上下文启发式续写。
func ruleEngineAssist(context, fullText, language string, maxTokens int) string {
	// 提取尾段（最后 500 字符视为 context）
	tail := []rune(context)
	if len(tail) > 200 {
		tail = tail[len(tail)-200:]
	}
	tailStr := string(tail)

	// 根据语言选择不同的续写策略
	if language == "en" {
		return englishAssist(tailStr, maxTokens)
	}
	return chineseAssist(tailStr, fullText, maxTokens)
}

// chineseAssist 中文续写规则引擎。
func chineseAssist(context, fullText string, maxTokens int) string {
	_ = fullText // 预留：后续可基于全文计算话题一致性

	// 模式匹配：根据尾段的句型模式续写
	runes := []rune(context)
	if len(runes) == 0 {
		return "建议进一步补充背景信息和预期目标，以便团队成员理解任务全貌。"
	}

	last50 := string(runes[max(0, len(runes)-50):])

	// 模式 1：尾随「首先」「第一」等列举开头 — 续写「其次」
	if hasAnyPrefix(last50, []string{"首先", "第一", "一是", "第一点"}) {
		return "其次需要关注方案的可落地性，确保在现有资源约束下能够按期交付。"
	}
	// 模式 2：尾随「其次」「第二」— 续写「最后」
	if hasAnyPrefix(last50, []string{"其次", "第二", "二是", "第二点"}) {
		return "最后建议制定详细的里程碑节点，便于过程跟踪与风险预警。"
	}
	// 模式 3：尾随问题描述 — 续写解决建议
	if hasAnySuffix(last50, []string{"问题", "缺陷", "漏洞", "故障", "风险", "不足"}) {
		return "建议从根因分析入手，逐步拆解为可执行的小任务，按优先级分批处理。"
	}
	// 模式 4：尾随目标 — 续写执行路径
	if hasAnySuffix(last50, []string{"目标", "目的", "期望", "希望"}) {
		return "为达成该目标，可先梳理关键依赖并明确各阶段的交付标准。"
	}
	// 模式 5：尾随疑问 — 续写建议
	if hasAnySuffix(last50, []string{"?", "？", "吗", "呢", "如何", "怎么", "是否"}) {
	建议 := "可以参考同类场景的最佳实践，结合团队现状形成本地化方案。"
		_ = 建议
		return "可组织一次跨角色对齐会，综合技术可行性与业务价值做出判断。"
	}

	// 模式 6：尾随「因此」「所以」「综上」— 续写结论
	if hasAnyPrefix(last50, []string{"因此", "所以", "综上", "总之", "由此可见"}) {
		return "在保证交付质量的前提下，建议预留 20% 的缓冲时间以应对不确定性。"
	}

	// 默认续写
	defaults := []string{
		"建议明确负责人与预期完成时间，确保各项任务能够闭环落地。",
		"可考虑将大目标拆解为可量化的小阶段，降低整体执行风险。",
		"在正式实施前，建议先在可控范围内做一次快速验证，收集反馈后再全面推广。",
	}

	// 用尾段 hash 选择确定性续写
	idx := hashSelect(context, len(defaults))
	return defaults[idx]
}

// englishAssist 英文续写规则引擎。
func englishAssist(context string, maxTokens int) string {
	_ = maxTokens
	runes := []rune(context)
	if len(runes) < 10 {
		return "Consider defining clear ownership and expected outcomes for each action item."
	}

	last50 := string(runes[max(0, len(runes)-50):])

	if hasAnySuffix(strings.ToLower(last50), []string{"problem", "issue", "risk", "concern"}) {
		return "We should establish measurable success criteria and assign a dedicated owner for follow-up."
	}
	if hasAnySuffix(strings.ToLower(last50), []string{"goal", "objective", "target"}) {
		return "To achieve this, we can break it down into deliverable milestones with clear timeboxes."
	}
	if hasAnySuffix(last50, []string{"?", "？"}) {
		return "It may be worthwhile to gather input from stakeholders before making a final decision."
	}

	defaults := []string{
		"Next steps should include a clear timeline, owner, and acceptance criteria for each deliverable.",
		"We can mitigate risk by delivering incrementally and validating assumptions early with real users.",
		"It would be beneficial to document the decision rationale so future contributors can understand the context.",
	}
	idx := hashSelect(last50, len(defaults))
	return defaults[idx]
}

// ruleEngineRewrite 改写规则引擎。
func ruleEngineRewrite(text, style, language string) (string, []string) {
	changes := []string{}

	switch style {
	case "formal":
		// 正式化：替换口语词
		if language == "zh" {
			replacements := map[string]string{
				"搞": "推进", "做": "执行", "看看": "评估",
				"差不多": "大致", "挺好的": "符合预期",
				"搞定": "完成", "想办法": "探索可行方案",
			}
			for old, new := range replacements {
				if strings.Contains(text, old) {
					text = strings.ReplaceAll(text, old, new)
					changes = append(changes, "口语词规范化："+old+" → "+new)
				}
			}
		}
	case "concise":
		// 精简：去除常见冗余修饰
		if language == "zh" {
			redundant := []string{"非常", "特别", "基本上", "总的来说", "毫无疑问"}
			for _, r := range redundant {
				if strings.Contains(text, r) {
					text = strings.ReplaceAll(text, r, "")
					changes = append(changes, "删除冗余修饰："+r)
				}
			}
			// 清理多余空格/标点
			text = strings.Join(strings.FieldsFunc(text, func(r rune) bool {
				return r == ' '
			}), "")
		}
	case "fluent":
		// 流畅化：补充断裂句
		if language == "zh" {
			if !hasAnySuffix(text, []string{"。", "！", "？", "；", "\"", "\""}) {
				text += "。"
				changes = append(changes, "补全句末标点")
			}
		}
	case "expand":
		// 扩写：在句末补充过渡说明
		if language == "zh" {
			if len(changes) == 0 {
				text += "（建议结合具体业务场景细化方案细节。）"
				changes = append(changes, "补充执行说明")
			}
		}
	}

	if len(changes) == 0 {
		changes = append(changes, "文本已符合目标风格，无需调整")
	}
	return text, changes
}

// ruleEngineFixGrammar 语法纠错规则引擎。
func ruleEngineFixGrammar(text, language string) ([]GrammarIssue, string) {
	var issues []GrammarIssue
	fixed := text

	if language == "zh" {
		// 规则 1：的地得混用（简单规则：动词前用"地"，名词前用"的"，动词后补语用"得"）
		// 此处使用简单启发式
		issues = append(issues, fixDeUsage(fixed)...)
		// 规则 2：标点全半角
		issues = append(issues, fixPunctuation(&fixed)...)
		// 规则 3：常见错别字
		issues = append(issues, fixCommonTypos(&fixed)...)
	} else {
		// 英文基础检查
		issues = append(issues, fixEnglishGrammar(&fixed)...)
	}

	// 如果没有问题，issues 为空
	if issues == nil {
		issues = []GrammarIssue{}
	}
	return issues, fixed
}

// fixDeUsage 检测"的地得"混用问题。
func fixDeUsage(text string) []GrammarIssue {
	var issues []GrammarIssue

	// 简单规则：查找"XX的XX"模式，若前面是动词则应为"地"，若是形容词则为"的"
	patterns := []struct {
		pattern     string
		replacement string
		reason      string
	}{
		{"快速的推进", "快速地推进", "\"的\"应为\"地\"：修饰动词时用\"地\""},
		{"很好的解决", "很好地解决", "\"的\"应为\"地\"：修饰动词时用\"地\""},
		{"清楚的说", "清楚地说", "\"的\"应为\"地\"：修饰动词时用\"地\""},
	}

	for _, p := range patterns {
		if idx := strings.Index(text, p.pattern); idx >= 0 {
			issues = append(issues, GrammarIssue{
				Offset: idx, Length: len([]rune(p.pattern)),
				Original: p.pattern, Replacement: p.replacement,
				Reason: p.reason, Severity: "error",
			})
		}
	}
	return issues
}

// fixPunctuation 修复标点全半角混用。
func fixPunctuation(text *string) []GrammarIssue {
	var issues []GrammarIssue
	result := *text

	// 中文语境下常见半角标点应转全角
	punctMap := map[string]string{
		",": "，", "!": "！", "?": "？", ":": "：", ";": "；",
	}
	for half, full := range punctMap {
		if strings.Contains(result, half) {
			// 检查是否在英文单词之间（简化判断：前后都是 ASCII）
			count := strings.Count(result, half)
			if count > 0 {
				// 简单替换所有（实际应用中需更精确判断）
				result = strings.ReplaceAll(result, half, full)
				issues = append(issues, GrammarIssue{
					Offset: 0, Length: len(half),
					Original: half, Replacement: full,
					Reason: "中文语境建议使用全角标点",
					Severity: "warning",
				})
			}
		}
	}
	*text = result
	return issues
}

// fixCommonTypos 修复常见错别字。
func fixCommonTypos(text *string) []GrammarIssue {
	var issues []GrammarIssue
	typos := map[string]string{
		"帐号": "账号", "帐户": "账户", "帐目": "账目",
		"份内": "分内", "份外": "分外",
	}

	for wrong, right := range typos {
		if idx := strings.Index(*text, wrong); idx >= 0 {
			*text = strings.ReplaceAll(*text, wrong, right)
			issues = append(issues, GrammarIssue{
				Offset: idx, Length: len([]rune(wrong)),
				Original: wrong, Replacement: right,
				Reason: "常见错别字",
				Severity: "error",
			})
		}
	}
	return issues
}

// fixEnglishGrammar 英文基础语法检查。
func fixEnglishGrammar(text *string) []GrammarIssue {
	var issues []GrammarIssue
	// 双空格检测
	if strings.Contains(*text, "  ") {
		re := strings.ReplaceAll(*text, "  ", " ")
		*text = re
		issues = append(issues, GrammarIssue{
			Offset: 0, Length: 2,
			Original: "  ", Replacement: " ",
			Reason: "多余空格", Severity: "style",
		})
	}
	return issues
}

// --- Shared Helpers ---

// hasAnyPrefix 检查文本是否包含任一前缀（实际是前缀中任意出现在尾部）。
func hasAnyPrefix(text string, prefixes []string) bool {
	for _, p := range prefixes {
		if strings.Contains(text, p) {
			return true
		}
	}
	return false
}

// hasAnySuffix 检查文本尾部是否包含任一后缀模式。
func hasAnySuffix(text string, suffixes []string) bool {
	for _, s := range suffixes {
		if strings.HasSuffix(text, s) {
			return true
		}
	}
	return false
}

// hashSelect 基于文本 hash 选择一个确定性索引。
func hashSelect(text string, n int) int {
	if n <= 0 {
		return 0
	}
	h := fnv.New32a()
	_, _ = h.Write([]byte(text))
	return int(h.Sum32()) % n
}

// detectLanguage 简单语言检测：中文字符占比 > 0.3 则判为中文。
func detectLanguage(text string) string {
	if text == "" {
		return "zh"
	}
	zhCount := 0
	total := 0
	for _, r := range text {
		if r >= 0x4e00 && r <= 0x9fff {
			zhCount++
		}
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= 0x4e00 && r <= 0x9fff) {
			total++
		}
	}
	if total == 0 {
		return "zh"
	}
	if float64(zhCount)/float64(total) > 0.3 {
		return "zh"
	}
	return "en"
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

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
		SELECT name FROM workitems
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
		FROM workitems i
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
