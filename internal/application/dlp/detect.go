// Package dlp — 数据防泄漏（Data Loss Prevention）。
//
// 对标 Plane 的 Argus 敏感数据扫描：
//   - 识别敏感数据（正则规则 + 关键词 + 指纹）
//   - 扫描任务（异步）与结果存储
//   - 写入前拦截/告警
//   - 扫描报告与规则配置
package dlp

import (
	"regexp"
	"strings"
	"sync"
)

// Category 敏感数据类型。
type Category string

const (
	// CategoryIDCard 身份证号。
	CategoryIDCard Category = "id_card"
	// CategoryBankCard 银行卡号。
	CategoryBankCard Category = "bank_card"
	// CategoryPhone 手机号。
	CategoryPhone Category = "phone"
	// CategoryEmail 邮箱。
	CategoryEmail Category = "email"
	// CategoryPassword 密码/密钥。
	CategoryPassword Category = "password"
	// CategoryAPIKey API 密钥。
	CategoryAPIKey Category = "api_key"
	// CategoryAddress 地址。
	CategoryAddress Category = "address"
	// CategoryPassport 护照号。
	CategoryPassport Category = "passport"
	// CategoryLicense 营业执照。
	CategoryLicense Category = "license"
	// CategoryCustom 自定义。
	CategoryCustom Category = "custom"
)

// Severity 严重级别。
type Severity string

const (
	// SeverityCritical 严重（阻断）。
	SeverityCritical Severity = "critical"
	// SeverityHigh 高（需审批）。
	SeverityHigh Severity = "high"
	// SeverityMedium 中（告警）。
	SeverityMedium Severity = "medium"
	// SeverityLow 低（记录）。
	SeverityLow Severity = "low"
)

// Match 单条匹配结果。
type Match struct {
	Category Category `json:"category"`
	Severity Severity `json:"severity"`
	// 匹配到的文本（已脱敏）
	Value string `json:"value"`
	// 原文长度
	RawLength int `json:"raw_length"`
	// 起始位置
	Offset int `json:"offset"`
	// 规则名
	RuleName string `json:"rule_name"`
	// 置信度 0-1
	Confidence float64 `json:"confidence"`
}

// Rule 检测规则。
type Rule struct {
	Name     string
	Category Category
	Severity Severity
	// 正则（优先）
	Pattern *regexp.Regexp
	// 关键词（正则不可用时）
	Keywords []string
	// 校验函数（如身份证校验位、银行卡 Luhn）
	Validator func(string) bool
	// 置信度
	Confidence float64
	// 是否启用
	Enabled bool
}

// Detector 敏感数据识别引擎。
type Detector struct {
	mu    sync.RWMutex
	rules map[Category][]Rule
}

// NewDetector 构造引擎并加载默认规则。
func NewDetector() *Detector {
	d := &Detector{rules: map[Category][]Rule{}}
	d.registerDefaults()
	return d
}

// Scan 扫描文本，返回匹配结果。
func (d *Detector) Scan(text string) []Match {
	d.mu.RLock()
	defer d.mu.RUnlock()

	var out []Match
	for cat, rules := range d.rules {
		for _, r := range rules {
			if !r.Enabled {
				continue
			}
			matches := d.applyRule(cat, r, text)
			out = append(out, matches...)
		}
	}
	return out
}

// ScanWithFilter 按类别过滤扫描。
func (d *Detector) ScanWithFilter(text string, cats []Category) []Match {
	all := d.Scan(text)
	if len(cats) == 0 {
		return all
	}
	want := map[Category]bool{}
	for _, c := range cats {
		want[c] = true
	}
	var out []Match
	for _, m := range all {
		if want[m.Category] {
			out = append(out, m)
		}
	}
	return out
}

// HighestSeverity 返回最高严重级别。
func HighestSeverity(matches []Match) Severity {
	if len(matches) == 0 {
		return ""
	}
	rank := map[Severity]int{
		SeverityLow: 1, SeverityMedium: 2, SeverityHigh: 3, SeverityCritical: 4,
	}
	max := SeverityLow
	for _, m := range matches {
		if rank[m.Severity] > rank[max] {
			max = m.Severity
		}
	}
	return max
}

// Mask 脱敏匹配值（保留前 2 后 2）。
func (m Match) Mask() string {
	if m.RawLength <= 4 {
		return strings.Repeat("*", m.RawLength)
	}
	if len(m.Value) <= 4 {
		return strings.Repeat("*", m.RawLength)
	}
	return m.Value[:2] + strings.Repeat("*", m.RawLength-4) + m.Value[len(m.Value)-2:]
}

// --- 规则注册 ---

func (d *Detector) registerDefaults() {
	d.Register(Rule{
		Name:       "id_card_cn",
		Category:   CategoryIDCard,
		Severity:   SeverityCritical,
		Pattern:    regexp.MustCompile(`\b[1-9]\d{5}(?:19|20)\d{2}(?:0[1-9]|1[0-2])(?:0[1-9]|[12]\d|3[01])\d{3}[\dXx]\b`),
		Validator:  validateCNIDCard,
		Confidence: 0.95,
		Enabled:    true,
	})

	d.Register(Rule{
		Name:       "bank_card_luhn",
		Category:   CategoryBankCard,
		Severity:   SeverityCritical,
		Pattern:    regexp.MustCompile(`\b(?:4[0-9]{12}(?:[0-9]{3})?|5[1-5][0-9]{14}|6(?:011|5[0-9]{2})[0-9]{12}|3[47][0-9]{13}|(?:2131|1800|35\d{3})\d{11})\b`),
		Validator:  validateLuhn,
		Confidence: 0.9,
		Enabled:    true,
	})

	d.Register(Rule{
		Name:       "phone_cn",
		Category:   CategoryPhone,
		Severity:   SeverityHigh,
		Pattern:    regexp.MustCompile(`\b1[3-9]\d{9}\b`),
		Confidence: 0.85,
		Enabled:    true,
	})

	d.Register(Rule{
		Name:       "email",
		Category:   CategoryEmail,
		Severity:   SeverityMedium,
		Pattern:    regexp.MustCompile(`\b[A-Za-z0-9._%+\-]+@[A-Za-z0-9.\-]+\.[A-Za-z]{2,}\b`),
		Confidence: 0.9,
		Enabled:    true,
	})

	d.Register(Rule{
		Name:       "api_key",
		Category:   CategoryAPIKey,
		Severity:   SeverityCritical,
		Pattern:    regexp.MustCompile(`\b(?:sk|pk|api[_-]?key|secret|token)[_-]?[A-Za-z0-9]{16,}\b`),
		Confidence: 0.8,
		Enabled:    true,
	})

	d.Register(Rule{
		Name:       "password_field",
		Category:   CategoryPassword,
		Severity:   SeverityHigh,
		Pattern:    regexp.MustCompile(`(?i)(password|passwd|pwd)\s*[:=]\s*["']?[^\s"']{6,}["']?`),
		Confidence: 0.75,
		Enabled:    true,
	})

	d.Register(Rule{
		Name:       "passport_cn",
		Category:   CategoryPassport,
		Severity:   SeverityCritical,
		Pattern:    regexp.MustCompile(`\b[EG]\d{8}\b`),
		Confidence: 0.8,
		Enabled:    true,
	})

	d.Register(Rule{
		Name:       "license_cn",
		Category:   CategoryLicense,
		Severity:   SeverityHigh,
		Pattern:    regexp.MustCompile(`\b[0-9A-HJ-NPQRTUWXY]{2}\d{6}[0-9A-HJ-NPQRTUWXY]{10}\b`),
		Confidence: 0.75,
		Enabled:    true,
	})
}

// Register 注册规则。
func (d *Detector) Register(r Rule) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.rules[r.Category] = append(d.rules[r.Category], r)
}

// Unregister 移除规则。
func (d *Detector) Unregister(name string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	for cat, rules := range d.rules {
		out := rules[:0]
		for _, r := range rules {
			if r.Name != name {
				out = append(out, r)
			}
		}
		d.rules[cat] = out
	}
}

// Rules 返回当前规则快照。
func (d *Detector) Rules() []Rule {
	d.mu.RLock()
	defer d.mu.RUnlock()
	var out []Rule
	for _, rs := range d.rules {
		out = append(out, rs...)
	}
	return out
}

// applyRule 应用单条规则。
func (d *Detector) applyRule(cat Category, r Rule, text string) []Match {
	var out []Match

	if r.Pattern != nil {
		for _, loc := range r.Pattern.FindAllStringIndex(text, -1) {
			raw := text[loc[0]:loc[1]]
			if r.Validator != nil && !r.Validator(raw) {
				continue
			}
			out = append(out, Match{
				Category:   cat,
				Severity:   r.Severity,
				Value:      raw,
				RawLength:  len(raw),
				Offset:     loc[0],
				RuleName:   r.Name,
				Confidence: r.Confidence,
			})
		}
		return out
	}

	for _, kw := range r.Keywords {
		idx := strings.Index(strings.ToLower(text), strings.ToLower(kw))
		if idx >= 0 {
			out = append(out, Match{
				Category:   cat,
				Severity:   r.Severity,
				Value:      text[idx : idx+len(kw)],
				RawLength:  len(kw),
				Offset:     idx,
				RuleName:   r.Name,
				Confidence: r.Confidence,
			})
		}
	}
	return out
}

// --- 校验函数 ---

// validateCNIDCard 校验中国大陆身份证号（ISO 7064:1983 MOD 11-2）。
func validateCNIDCard(s string) bool {
	if len(s) != 18 {
		return false
	}
	weights := []int{7, 9, 10, 5, 8, 4, 2, 1, 6, 3, 7, 9, 10, 5, 8, 4, 2}
	checks := []byte{'1', '0', 'X', '9', '8', '7', '6', '5', '4', '3', '2'}

	sum := 0
	for i := 0; i < 17; i++ {
		c := s[i]
		if c < '0' || c > '9' {
			return false
		}
		sum += int(c-'0') * weights[i]
	}
	expected := checks[sum%11]
	last := s[17]
	if last == 'x' {
		last = 'X'
	}
	return last == expected
}

// validateLuhn 银行卡号 Luhn 校验。
func validateLuhn(s string) bool {
	var sum int
	alternate := false
	for i := len(s) - 1; i >= 0; i-- {
		c := s[i]
		if c < '0' || c > '9' {
			return false
		}
		n := int(c - '0')
		if alternate {
			n *= 2
			if n > 9 {
				n -= 9
			}
		}
		sum += n
		alternate = !alternate
	}
	return sum%10 == 0
}
