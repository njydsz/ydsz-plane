// Package searchql 提供类 Jira JQL 的搜索语法解析器。
//
// 设计参考:
//   - Jira JQL (Jira Query Language)
//   - Elasticsearch Query DSL
//   - Linear search syntax
//
// 语法 (EBNF):
//
//	query      := orExpr
//	orExpr     := andExpr (OR andExpr)*
//	andExpr    := factor (AND factor | factor)*     // 空格隐式 AND
//	factor     := NOT factor | '(' orExpr ')' | clause | phrase
//	clause     := field op value
//	field      := project|type|status|priority|severity|assignee|
//	              reporter|label|module|sprint|version|due|created|updated
//	op         := : | = | != | > | >= | < | <= | in
//	value      := word | "quoted" | (v1, v2) | me() | currentUser() | now() | now(-7d)
//	phrase     := 裸关键词
//
// 使用示例:
//
//	"登录页 闪退"                             → 全文检索
//	"project:YD status:todo assignee:me()"   → 结构化过滤
//	"type:defect severity>=3 created>now(-7d)" → 范围查询
//	""支付回调" AND module:支付"              → 短语+字段
//
// 解析失败时降级为纯全文查询（不报错，保证用户搜索体验）。
package searchql

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Query 解析后的搜索查询 AST。
type Query struct {
	// Raw 原始查询字符串。
	Raw string `json:"raw"`

	// Text 纯文本搜索词（用于全文检索）。
	Text string `json:"text,omitempty"`

	// Clauses 结构化过滤条件。
	Clauses []Clause `json:"clauses,omitempty"`

	// IsDegraded 为 true 表示解析失败，降级为纯文本搜索。
	IsDegraded bool `json:"is_degraded,omitempty"`
}

// Clause 表示一个 field op value 过滤条件。
type Clause struct {
	Field    string `json:"field"`
	Operator string `json:"op"`   // :, =, !=, >, >=, <, <=, in
	Value    any    `json:"value"` // string, []string, int, time.Time
	Negated  bool   `json:"negated,omitempty"`
}

// --- Token types ---

type tokenType int

const (
	tokEOF tokenType = iota
	tokWord
	tokQuoted
	tokLParen
	tokRParen
	tokColon
	tokEq
	tokNeq
	tokGt
	tokGte
	tokLt
	tokLte
	tokAnd
	tokOr
	tokNot
	tokIn
)

type token struct {
	typ tokenType
	val string
}

// --- Lexer ---

type lexer struct {
	input []rune
	pos   int
}

func newLexer(input string) *lexer {
	return &lexer{input: []rune(input), pos: 0}
}

func (l *lexer) peek() rune {
	if l.pos >= len(l.input) {
		return 0
	}
	return l.input[l.pos]
}

func (l *lexer) next() rune {
	r := l.peek()
	if r != 0 {
		l.pos++
	}
	return r
}

func (l *lexer) skipWS() {
	for l.pos < len(l.input) && (l.peek() == ' ' || l.peek() == '\t' || l.peek() == '\n' || l.peek() == '\r') {
		l.next()
	}
}

func (l *lexer) readWord() string {
	start := l.pos
	for l.pos < len(l.input) {
		r := l.peek()
		if r == ' ' || r == '\t' || r == '\n' || r == '\r' ||
			r == ':' || r == '=' || r == '!' || r == '>' || r == '<' ||
			r == 0 {
			break
		}
		// 遇到 '(' 时检查是否是函数调用语法 (word + '(' + ... + ')')
		if r == '(' {
			// 保存位置，尝试读取函数调用
			saved := l.pos
			l.next() // skip '('
			depth := 1
			for l.pos < len(l.input) && depth > 0 {
				c := l.peek()
				if c == '(' {
					depth++
				} else if c == ')' {
					depth--
				}
				l.next()
			}
			if depth == 0 {
				// 成功匹配函数调用语法
				return string(l.input[start:l.pos])
			}
			// 未匹配，回退
			l.pos = saved
			break
		}
		if r == ')' {
			break
		}
		l.next()
	}
	return string(l.input[start:l.pos])
}

func (l *lexer) readQuoted() (string, error) {
	// skip opening quote
	l.next()
	start := l.pos
	for l.pos < len(l.input) {
		r := l.peek()
		if r == '"' {
			val := string(l.input[start:l.pos])
			l.next() // skip closing quote
			return val, nil
		}
		if r == '\\' {
			l.next() // skip backslash, keep next char
		}
		l.next()
	}
	return "", fmt.Errorf("unterminated quoted string")
}

func (l *lexer) readInList() (string, error) {
	// skip opening (
	l.next()
	start := l.pos
	depth := 1
	for l.pos < len(l.input) && depth > 0 {
		r := l.peek()
		if r == '(' {
			depth++
		} else if r == ')' {
			depth--
			if depth == 0 {
				val := string(l.input[start:l.pos])
				l.next() // skip closing )
				return val, nil
			}
		}
		l.next()
	}
	return "", fmt.Errorf("unterminated in-list")
}

func (l *lexer) nextToken() (token, error) {
	l.skipWS()
	r := l.peek()
	if r == 0 {
		return token{typ: tokEOF}, nil
	}

	switch r {
	case '(':
		l.next()
		return token{typ: tokLParen, val: "("}, nil
	case ')':
		l.next()
		return token{typ: tokRParen, val: ")"}, nil
	case ':':
		l.next()
		return token{typ: tokColon, val: ":"}, nil
	case '=':
		l.next()
		return token{typ: tokEq, val: "="}, nil
	case '!':
		l.next()
		if l.peek() == '=' {
			l.next()
			return token{typ: tokNeq, val: "!="}, nil
		}
		return token{typ: tokNot, val: "!"}, nil
	case '>':
		l.next()
		if l.peek() == '=' {
			l.next()
			return token{typ: tokGte, val: ">="}, nil
		}
		return token{typ: tokGt, val: ">"}, nil
	case '<':
		l.next()
		if l.peek() == '=' {
			l.next()
			return token{typ: tokLte, val: "<="}, nil
		}
		return token{typ: tokLt, val: "<"}, nil
	case '"':
		val, err := l.readQuoted()
		if err != nil {
			return token{}, err
		}
		return token{typ: tokQuoted, val: val}, nil
	default:
		word := l.readWord()
		upper := strings.ToUpper(word)
		switch upper {
		case "AND":
			return token{typ: tokAnd, val: word}, nil
		case "OR":
			return token{typ: tokOr, val: word}, nil
		case "NOT":
			return token{typ: tokNot, val: word}, nil
		case "IN":
			return token{typ: tokIn, val: word}, nil
		}
		return token{typ: tokWord, val: word}, nil
	}
}

// --- Parser (递归下降) ---

type parser struct {
	lex       *lexer
	tok       token
	err       error
	textParts []string
	clauses   []Clause
}

// Parse 解析类 JQL 查询字符串。
// 解析失败时返回 degraded=true，Query.Text 保留原始输入用于全文搜索。
func Parse(input string) *Query {
	if strings.TrimSpace(input) == "" {
		return &Query{Raw: input}
	}

	p := &parser{
		lex:       newLexer(input),
		textParts: []string{},
		clauses:   []Clause{},
	}

	// 尝试解析
	if err := p.nextToken(); err != nil {
		return &Query{Raw: input, Text: input, IsDegraded: true}
	}

	p.parseOrExpr()

	// 如果解析出错或没有结构化子句，降级为全文搜索
	if p.err != nil || len(p.clauses) == 0 {
		return &Query{Raw: input, Text: input, IsDegraded: p.err != nil}
	}

	text := strings.Join(p.textParts, " ")

	return &Query{
		Raw:        input,
		Text:       text,
		Clauses:    p.clauses,
		IsDegraded: false,
	}
}

func (p *parser) nextToken() error {
	tok, err := p.lex.nextToken()
	if err != nil {
		p.err = err
		return err
	}
	p.tok = tok
	return nil
}

// parseOrExpr: andExpr (OR andExpr)*
func (p *parser) parseOrExpr() {
	p.parseAndExpr()
	for p.tok.typ == tokOr {
		p.nextToken()
		p.parseAndExpr()
	}
}

// parseAndExpr: factor (AND factor | factor)*
func (p *parser) parseAndExpr() {
	p.parseFactor()
	for p.tok.typ != tokEOF && p.tok.typ != tokRParen && p.tok.typ != tokOr {
		if p.tok.typ == tokAnd {
			p.nextToken()
		}
		p.parseFactor()
	}
}

// parseFactor: NOT factor | '(' orExpr ')' | clause | phrase
func (p *parser) parseFactor() {
	switch p.tok.typ {
	case tokNot:
		p.nextToken()
		p.parseFactor()
		// Mark last clause as negated
		if len(p.clauses) > 0 {
			p.clauses[len(p.clauses)-1].Negated = true
		}
	case tokLParen:
		p.nextToken()
		p.parseOrExpr()
		if p.tok.typ == tokRParen {
			p.nextToken()
		}
	default:
		if p.isFieldStart() {
			p.parseClause()
		} else {
			p.parsePhrase()
		}
	}
}

// isFieldStart 检查当前位置是否可能是 field:value 子句的开始。
func (p *parser) isFieldStart() bool {
	if p.tok.typ != tokWord {
		return false
	}
	// 向前探测一个 token
	saved := p.lex.pos
	fieldTok := p.tok

	// 先跳过当前 token 并看下一个
	nextTok, err := p.lex.nextToken()
	if err != nil {
		p.lex.pos = saved
		return false
	}

	// 如果下一个是 : 或 IN 或比较运算符，则是 field 子句
	result := nextTok.typ == tokColon || nextTok.typ == tokIn ||
		nextTok.typ == tokEq || nextTok.typ == tokNeq ||
		nextTok.typ == tokGt || nextTok.typ == tokGte ||
		nextTok.typ == tokLt || nextTok.typ == tokLte

	// 还原
	p.lex.pos = saved
	p.tok = fieldTok

	return result
}

// parseClause: field op value
func (p *parser) parseClause() {
	field := strings.ToLower(p.tok.val)
	p.nextToken()

	op := ":"
	switch p.tok.typ {
	case tokColon:
		op = ":"
	case tokEq:
		op = "="
	case tokNeq:
		op = "!="
	case tokGt:
		op = ">"
	case tokGte:
		op = ">="
	case tokLt:
		op = "<"
	case tokLte:
		op = "<="
	case tokIn:
		op = "in"
	default:
		// 不是有效的操作符，回退
		p.textParts = append(p.textParts, field)
		return
	}
	p.nextToken()

	val := p.parseValue()
	if val == nil {
		return
	}

	// 特殊值解析
	parsedVal := parseSpecialValue(field, val)
	if parsedVal != nil {
		val = parsedVal
	}

	p.clauses = append(p.clauses, Clause{
		Field:    field,
		Operator: op,
		Value:    val,
	})
}

// parseValue: word | "quoted" | (v1, v2)
func (p *parser) parseValue() any {
	switch p.tok.typ {
	case tokWord:
		val := p.tok.val
		p.nextToken()
		return val
	case tokQuoted:
		val := p.tok.val
		p.nextToken()
		return val
	case tokLParen:
		// in-list: (v1, v2, v3)
		list, err := p.lex.readInList()
		if err != nil {
			p.err = err
			return nil
		}
		parts := strings.Split(list, ",")
		var values []string
		for _, part := range parts {
			part = strings.TrimSpace(part)
			part = strings.Trim(part, "\"")
			if part != "" {
				values = append(values, part)
			}
		}
		p.nextToken() // consume next token after ')'
		return values
	default:
		return nil
	}
}

// parsePhrase 解析纯文本搜索词（用于全文检索）。
func (p *parser) parsePhrase() {
	switch p.tok.typ {
	case tokWord:
		p.textParts = append(p.textParts, p.tok.val)
		p.nextToken()
	case tokQuoted:
		// 引号内的短语保留原样用于精确匹配
		p.textParts = append(p.textParts, `"`+p.tok.val+`"`)
		p.nextToken()
	}
}

// --- 特殊值解析 ---

func parseSpecialValue(field string, val any) any {
	s, ok := val.(string)
	if !ok {
		return nil
	}

	lower := strings.ToLower(s)

	// me() / currentUser() → 保持原字符串，由调用方替换
	if lower == "me()" || lower == "currentuser()" {
		return "__CURRENT_USER__"
	}

	// now() / now(-7d) / now(+1w) → 保持原字符串，由调用方替换
	if strings.HasPrefix(lower, "now(") && strings.HasSuffix(lower, ")") {
		offset := lower[4 : len(lower)-1]
		return parseNowOffset(offset)
	}

	return nil
}

func parseNowOffset(offset string) any {
	offset = strings.TrimSpace(offset)
	now := time.Now()

	if offset == "" {
		return now.Truncate(24 * time.Hour)
	}

	// parse: -7d, +1w, -3h
	neg := false
	if strings.HasPrefix(offset, "-") {
		neg = true
		offset = offset[1:]
	} else if strings.HasPrefix(offset, "+") {
		offset = offset[1:]
	}

	numStr := ""
	unit := ""
	for _, ch := range offset {
		if ch >= '0' && ch <= '9' {
			numStr += string(ch)
		} else {
			unit += string(ch)
		}
	}

	num, err := strconv.Atoi(numStr)
	if err != nil || num == 0 {
		return now.Truncate(24 * time.Hour)
	}

	var d time.Duration
	switch strings.ToLower(unit) {
	case "d", "day", "days":
		d = time.Duration(num) * 24 * time.Hour
	case "w", "week", "weeks":
		d = time.Duration(num) * 7 * 24 * time.Hour
	case "h", "hour", "hours":
		d = time.Duration(num) * time.Hour
	case "m", "min", "minute", "minutes":
		d = time.Duration(num) * time.Minute
	default:
		return now.Truncate(24 * time.Hour)
	}

	if neg {
		return now.Add(-d)
	}
	return now.Add(d)
}

// --- 已知字段列表 ---

// KnownFields 返回所有支持的搜索字段。
func KnownFields() []string {
	return []string{
		"project", "type", "status", "priority", "severity",
		"assignee", "reporter", "label", "module", "sprint",
		"version", "due", "created", "updated", "identifier",
	}
}

// IsValidField 检查字段名是否有效。
func IsValidField(field string) bool {
	for _, f := range KnownFields() {
		if f == field {
			return true
		}
	}
	return false
}
