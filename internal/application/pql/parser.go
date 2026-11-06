// Package pql — Plane Query Language (PQL) 解析与执行。
//
// 目标：提供一套声明式查询语言，用于工作项列表过滤/搜索。
// 语法示例：
//
//	state = "in_progress" AND priority = "high"
//	assignee = "me" OR assignee = @zhangsan
//	created_at > "2026-01-01" AND created_at <= "2026-12-31"
//	title ~ "登录" OR description ~ "鉴权"
//	type = "task" AND estimate >= 5
//	label in ("backend", "urgent") AND NOT state = "cancelled"
//	state = "active" sort:created_at desc limit:20
//
// 设计：
//   - Lexer + 递归下降 parser（支持 AND/OR/NOT、比较符、in、like）；
//   - 解析为 AST；
//   - Compiler 将 AST 编译为参数化 SQL（防注入）；
//   - 字段白名单，禁止任意列名拼接。
package pql

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode"
)

// TokenKind 词法单元类型。
type TokenKind int

const (
	TokenEOF TokenKind = iota
	TokenIdent
	TokenString
	TokenNumber
	// 运算符
	TokenEq     // =
	TokenNe     // !=
	TokenGt     // >
	TokenGte    // >=
	TokenLt     // <
	TokenLte    // <=
	TokenTilde  // ~ (包含)
	TokenLParen // (
	TokenRParen // )
	TokenComma  // ,
	TokenColon  // :
	// 关键字
	TokenAnd
	TokenOr
	TokenNot
	TokenIn
	TokenSort
	TokenLimit
	TokenOffset
	TokenAsc
	TokenDesc
	TokenMe
	TokenAt
)

// String 返回 TokenKind 的可读名称（调试用）。
func (k TokenKind) String() string {
	switch k {
	case TokenEOF:
		return "EOF"
	case TokenIdent:
		return "IDENT"
	case TokenString:
		return "STRING"
	case TokenNumber:
		return "NUMBER"
	case TokenEq:
		return "="
	case TokenNe:
		return "!="
	case TokenGt:
		return ">"
	case TokenGte:
		return ">="
	case TokenLt:
		return "<"
	case TokenLte:
		return "<="
	case TokenTilde:
		return "~"
	case TokenLParen:
		return "("
	case TokenRParen:
		return ")"
	case TokenComma:
		return ","
	case TokenColon:
		return ":"
	case TokenAnd:
		return "AND"
	case TokenOr:
		return "OR"
	case TokenNot:
		return "NOT"
	case TokenIn:
		return "IN"
	case TokenSort:
		return "SORT"
	case TokenLimit:
		return "LIMIT"
	case TokenOffset:
		return "OFFSET"
	case TokenAsc:
		return "ASC"
	case TokenDesc:
		return "DESC"
	case TokenMe:
		return "ME"
	case TokenAt:
		return "AT"
	}
	return "UNKNOWN"
}

// Token 词法单元。
type Token struct {
	Kind TokenKind
	Text string
	Pos  int
}

// lexer 词法分析器。
type lexer struct {
	input string
	pos   int
}

func newLexer(input string) *lexer {
	return &lexer{input: input}
}

func (l *lexer) peek() byte {
	if l.pos >= len(l.input) {
		return 0
	}
	return l.input[l.pos]
}

func (l *lexer) next() byte {
	if l.pos >= len(l.input) {
		return 0
	}
	c := l.input[l.pos]
	l.pos++
	return c
}

// Tokenize 将输入切分为词法单元序列。
func (l *lexer) Tokenize() ([]Token, error) {
	var tokens []Token
	for {
		l.skipSpace()
		if l.pos >= len(l.input) {
			tokens = append(tokens, Token{Kind: TokenEOF, Pos: l.pos})
			return tokens, nil
		}
		start := l.pos
		c := l.peek()

		switch {
		case c == '"' || c == '\'':
			s, err := l.readQuoted(c)
			if err != nil {
				return nil, err
			}
			tokens = append(tokens, Token{Kind: TokenString, Text: s, Pos: start})
		case c == '@':
			l.next()
			name := l.readIdent()
			tokens = append(tokens, Token{Kind: TokenAt, Text: name, Pos: start})
		case c >= '0' && c <= '9':
			num := l.readNumber()
			tokens = append(tokens, Token{Kind: TokenNumber, Text: num, Pos: start})
		case c == ':':
			l.next()
			tokens = append(tokens, Token{Kind: TokenColon, Text: ":", Pos: start})
		case isIdentStart(rune(c)):
			word := l.readIdent()
			kind := TokenIdent
			switch strings.ToLower(word) {
			case "and":
				kind = TokenAnd
			case "or":
				kind = TokenOr
			case "not":
				kind = TokenNot
			case "in":
				kind = TokenIn
			case "sort":
				kind = TokenSort
			case "limit":
				kind = TokenLimit
			case "offset":
				kind = TokenOffset
			case "asc":
				kind = TokenAsc
			case "desc":
				kind = TokenDesc
			case "me":
				kind = TokenMe
			}
			tokens = append(tokens, Token{Kind: kind, Text: word, Pos: start})
		default:
			switch c {
			case '=':
				l.next()
				tokens = append(tokens, Token{Kind: TokenEq, Text: "=", Pos: start})
			case '!':
				l.next()
				if l.peek() == '=' {
					l.next()
					tokens = append(tokens, Token{Kind: TokenNe, Text: "!=", Pos: start})
				} else {
					return nil, fmt.Errorf("pql: unexpected character %q at %d", string(c), start)
				}
			case '>':
				l.next()
				if l.peek() == '=' {
					l.next()
					tokens = append(tokens, Token{Kind: TokenGte, Text: ">=", Pos: start})
				} else {
					tokens = append(tokens, Token{Kind: TokenGt, Text: ">", Pos: start})
				}
			case '<':
				l.next()
				if l.peek() == '=' {
					l.next()
					tokens = append(tokens, Token{Kind: TokenLte, Text: "<=", Pos: start})
				} else {
					tokens = append(tokens, Token{Kind: TokenLt, Text: "<", Pos: start})
				}
			case '~':
				l.next()
				tokens = append(tokens, Token{Kind: TokenTilde, Text: "~", Pos: start})
			case '(':
				l.next()
				tokens = append(tokens, Token{Kind: TokenLParen, Text: "(", Pos: start})
			case ')':
				l.next()
				tokens = append(tokens, Token{Kind: TokenRParen, Text: ")", Pos: start})
			case ',':
				l.next()
				tokens = append(tokens, Token{Kind: TokenComma, Text: ",", Pos: start})
			default:
				return nil, fmt.Errorf("pql: unexpected character %q at %d", string(c), start)
			}
		}
	}
}

func (l *lexer) skipSpace() {
	for l.pos < len(l.input) {
		if unicode.IsSpace(rune(l.input[l.pos])) {
			l.pos++
		} else {
			break
		}
	}
}

func (l *lexer) readQuoted(quote byte) (string, error) {
	l.next()
	var sb strings.Builder
	for l.pos < len(l.input) {
		c := l.next()
		if c == quote {
			return sb.String(), nil
		}
		if c == '\\' && l.pos < len(l.input) {
			sb.WriteByte(l.next())
			continue
		}
		sb.WriteByte(c)
	}
	return "", fmt.Errorf("pql: unterminated string")
}

func (l *lexer) readIdent() string {
	start := l.pos
	for l.pos < len(l.input) && isIdentPart(rune(l.input[l.pos])) {
		l.pos++
	}
	return l.input[start:l.pos]
}

func (l *lexer) readNumber() string {
	start := l.pos
	for l.pos < len(l.input) {
		c := l.input[l.pos]
		if (c >= '0' && c <= '9') || c == '.' {
			l.pos++
		} else {
			break
		}
	}
	return l.input[start:l.pos]
}

func isIdentStart(r rune) bool {
	return unicode.IsLetter(r) || r == '_'
}

func isIdentPart(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '.' || r == '-'
}

// --- AST ---

// Expr 查询表达式节点。
type Expr interface {
	exprNode()
}

// ComparisonExpr 比较表达式。
type ComparisonExpr struct {
	Field  string
	Op     TokenKind
	Value  string
	Values []string
	AtUser string
	IsMe   bool
}

func (*ComparisonExpr) exprNode() {}

// LogicalExpr 逻辑组合。
type LogicalExpr struct {
	Op    TokenKind
	Left  Expr
	Right Expr
}

func (*LogicalExpr) exprNode() {}

// NotExpr 否定。
type NotExpr struct {
	Inner Expr
}

func (*NotExpr) exprNode() {}

// Query 完整查询。
type Query struct {
	Where   Expr
	SortBy  string
	SortDir string
	Limit   int
	Offset  int
}

// parser 递归下降 parser。
type parser struct {
	tokens []Token
	pos    int
}

func newParser(tokens []Token) *parser {
	return &parser{tokens: tokens}
}

func (p *parser) cur() Token {
	if p.pos >= len(p.tokens) {
		return Token{Kind: TokenEOF}
	}
	return p.tokens[p.pos]
}

func (p *parser) advance() Token {
	t := p.cur()
	p.pos++
	return t
}

// Parse 解析完整查询（含 sort/limit/offset 子句）。
func Parse(input string) (*Query, error) {
	tokens, err := newLexer(input).Tokenize()
	if err != nil {
		return nil, err
	}
	p := newParser(tokens)
	q := &Query{SortDir: "asc", Limit: 100}

	for p.cur().Kind != TokenEOF {
		t := p.cur()
		switch t.Kind {
		case TokenSort, TokenLimit, TokenOffset:
			kw := strings.ToLower(p.advance().Text)
			if p.cur().Kind == TokenColon {
				p.advance()
			}
			val := p.advance()
			if val.Kind != TokenIdent && val.Kind != TokenNumber {
				return nil, fmt.Errorf("pql: expected value after %s", kw)
			}
			switch kw {
			case "sort":
				q.SortBy = val.Text
				if p.cur().Kind == TokenAsc || p.cur().Kind == TokenDesc {
					q.SortDir = strings.ToLower(p.advance().Text)
				}
			case "limit":
				n, err := strconv.Atoi(val.Text)
				if err != nil || n <= 0 || n > 1000 {
					return nil, fmt.Errorf("pql: invalid limit %q", val.Text)
				}
				q.Limit = n
			case "offset":
				n, err := strconv.Atoi(val.Text)
				if err != nil || n < 0 {
					return nil, fmt.Errorf("pql: invalid offset %q", val.Text)
				}
				q.Offset = n
			}
		default:
			if q.Where != nil {
				return nil, fmt.Errorf("pql: unexpected token %q at %d", t.Text, t.Pos)
			}
			expr, err := p.parseOr()
			if err != nil {
				return nil, err
			}
			q.Where = expr
		}
	}
	return q, nil
}

func (p *parser) parseOr() (Expr, error) {
	left, err := p.parseAnd()
	if err != nil {
		return nil, err
	}
	for p.cur().Kind == TokenOr {
		p.advance()
		right, err := p.parseAnd()
		if err != nil {
			return nil, err
		}
		left = &LogicalExpr{Op: TokenOr, Left: left, Right: right}
	}
	return left, nil
}

func (p *parser) parseAnd() (Expr, error) {
	left, err := p.parseNot()
	if err != nil {
		return nil, err
	}
	for p.cur().Kind == TokenAnd {
		p.advance()
		right, err := p.parseNot()
		if err != nil {
			return nil, err
		}
		left = &LogicalExpr{Op: TokenAnd, Left: left, Right: right}
	}
	return left, nil
}

func (p *parser) parseNot() (Expr, error) {
	if p.cur().Kind == TokenNot {
		p.advance()
		inner, err := p.parseNot()
		if err != nil {
			return nil, err
		}
		return &NotExpr{Inner: inner}, nil
	}
	return p.parseComparison()
}

func (p *parser) parseComparison() (Expr, error) {
	if p.cur().Kind == TokenLParen {
		p.advance()
		expr, err := p.parseOr()
		if err != nil {
			return nil, err
		}
		if p.cur().Kind != TokenRParen {
			return nil, fmt.Errorf("pql: expected ')' at %d", p.cur().Pos)
		}
		p.advance()
		return expr, nil
	}

	fieldTok := p.advance()
	if fieldTok.Kind != TokenIdent {
		return nil, fmt.Errorf("pql: expected field name at %d, got %q", fieldTok.Pos, fieldTok.Text)
	}
	field := fieldTok.Text

	if p.cur().Kind == TokenIn {
		p.advance()
		if p.cur().Kind != TokenLParen {
			return nil, fmt.Errorf("pql: expected '(' after in at %d", p.cur().Pos)
		}
		p.advance()
		var values []string
		for p.cur().Kind == TokenString || p.cur().Kind == TokenIdent || p.cur().Kind == TokenNumber {
			t := p.advance()
			values = append(values, t.Text)
			if p.cur().Kind == TokenComma {
				p.advance()
				continue
			}
			break
		}
		if p.cur().Kind != TokenRParen {
			return nil, fmt.Errorf("pql: expected ')' to close in-list at %d", p.cur().Pos)
		}
		p.advance()
		return &ComparisonExpr{Field: field, Op: TokenIn, Values: values}, nil
	}

	opTok := p.advance()
	switch opTok.Kind {
	case TokenEq, TokenNe, TokenGt, TokenGte, TokenLt, TokenLte, TokenTilde:
	default:
		return nil, fmt.Errorf("pql: expected operator after %q at %d", field, opTok.Pos)
	}

	valueTok := p.advance()
	expr := &ComparisonExpr{Field: field, Op: opTok.Kind}
	switch valueTok.Kind {
	case TokenString, TokenIdent, TokenNumber:
		expr.Value = valueTok.Text
	case TokenMe:
		expr.IsMe = true
	case TokenAt:
		expr.AtUser = valueTok.Text
	default:
		return nil, fmt.Errorf("pql: expected value at %d", valueTok.Pos)
	}
	return expr, nil
}

// --- 字段映射（白名单，防注入） ---

// FieldMapping 字段元数据。
type FieldMapping struct {
	Column  string
	Kind    string
	EnumSet map[string]bool
}

// DefaultFieldMappings 工作项默认字段映射。
var DefaultFieldMappings = map[string]FieldMapping{
	"state":        {Column: "state_name", Kind: "text"},
	"state_group":  {Column: "state_group", Kind: "enum", EnumSet: setOf("backlog", "started", "completed", "cancelled")},
	"priority":     {Column: "priority", Kind: "enum", EnumSet: setOf("urgent", "high", "medium", "low", "none")},
	"type":         {Column: "type_code", Kind: "enum", EnumSet: setOf("requirement", "task", "defect", "epic")},
	"title":        {Column: "title", Kind: "text"},
	"description":  {Column: "description", Kind: "text"},
	"assignee":     {Column: "assignee_name", Kind: "user"},
	"creator":      {Column: "creator_name", Kind: "user"},
	"estimate":     {Column: "estimate_point", Kind: "number"},
	"label":        {Column: "label_name", Kind: "text"},
	"module":       {Column: "module_name", Kind: "text"},
	"sprint":       {Column: "sprint_name", Kind: "text"},
	"created_at":   {Column: "created_at", Kind: "date"},
	"updated_at":   {Column: "updated_at", Kind: "date"},
	"target_date":  {Column: "target_date", Kind: "date"},
	"start_date":   {Column: "start_date", Kind: "date"},
	"completed_at": {Column: "completed_at", Kind: "date"},
	"severity":     {Column: "severity", Kind: "number"},
}

// AllowedSortFields 允许排序的字段。
var AllowedSortFields = map[string]bool{
	"created_at": true, "updated_at": true, "target_date": true,
	"start_date": true, "priority": true, "state": true, "title": true,
	"estimate": true, "completed_at": true,
}

func setOf(vals ...string) map[string]bool {
	m := make(map[string]bool, len(vals))
	for _, v := range vals {
		m[v] = true
	}
	return m
}

// NormalizeDate 规范化日期输入（支持 YYYY-MM-DD / YYYY/MM/DD）。
func NormalizeDate(s string) string {
	s = strings.ReplaceAll(s, "/", "-")
	if _, err := time.Parse("2006-01-02", s); err == nil {
		return s
	}
	return s
}
