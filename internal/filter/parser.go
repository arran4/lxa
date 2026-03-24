package filter

import (
	"errors"
	"fmt"
	"strings"

	"github.com/lxa-project/lxa/internal/xattr"
)

// Node represents a node in the filter AST.
type Node interface {
	Eval(md xattr.Metadata) bool
}

// Token types for lexing.
type tokenType int

const (
	tokenEOF tokenType = iota
	tokenIdent
	tokenAnd
	tokenOr
	tokenNot
	tokenLParen
	tokenRParen
)

type token struct {
	typ tokenType
	lit string
}

// lexer tokenizes a filter expression.
type lexer struct {
	input string
	pos   int
}

func newLexer(input string) *lexer {
	return &lexer{input: input, pos: 0}
}

func (l *lexer) nextToken() token {
	for l.pos < len(l.input) && isWhitespace(l.input[l.pos]) {
		l.pos++
	}

	if l.pos >= len(l.input) {
		return token{typ: tokenEOF, lit: ""}
	}

	ch := l.input[l.pos]
	if ch == '(' {
		l.pos++
		return token{typ: tokenLParen, lit: "("}
	}
	if ch == ')' {
		l.pos++
		return token{typ: tokenRParen, lit: ")"}
	}
	if ch == '!' {
		l.pos++
		return token{typ: tokenNot, lit: "!"}
	}
	if ch == '&' && l.pos+1 < len(l.input) && l.input[l.pos+1] == '&' {
		l.pos += 2
		return token{typ: tokenAnd, lit: "&&"}
	}
	if ch == '|' && l.pos+1 < len(l.input) && l.input[l.pos+1] == '|' {
		l.pos += 2
		return token{typ: tokenOr, lit: "||"}
	}

	// Identifier (e.g., has:tags, tag:foo, xdg)
	start := l.pos
	for l.pos < len(l.input) && !isWhitespace(l.input[l.pos]) &&
		l.input[l.pos] != '(' && l.input[l.pos] != ')' &&
		l.input[l.pos] != '!' && l.input[l.pos] != '&' && l.input[l.pos] != '|' {
		l.pos++
	}
	lit := l.input[start:l.pos]
	return token{typ: tokenIdent, lit: lit}
}

func isWhitespace(ch byte) bool {
	return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r'
}

// parser builds the AST from tokens.
type parser struct {
	lex  *lexer
	curr token
	errs []error
}

func newParser(input string) *parser {
	p := &parser{lex: newLexer(input)}
	p.next()
	return p
}

func (p *parser) next() {
	p.curr = p.lex.nextToken()
}

func (p *parser) parseExpression() Node {
	return p.parseOr()
}

func (p *parser) parseOr() Node {
	left := p.parseAnd()
	for p.curr.typ == tokenOr {
		p.next()
		right := p.parseAnd()
		left = &orNode{left: left, right: right}
	}
	return left
}

func (p *parser) parseAnd() Node {
	left := p.parseUnary()
	for p.curr.typ == tokenAnd {
		p.next()
		right := p.parseUnary()
		left = &andNode{left: left, right: right}
	}
	return left
}

func (p *parser) parseUnary() Node {
	if p.curr.typ == tokenNot {
		p.next()
		expr := p.parseUnary()
		return &notNode{expr: expr}
	}
	return p.parsePrimary()
}

func (p *parser) parsePrimary() Node {
	switch p.curr.typ {
	case tokenLParen:
		p.next()
		expr := p.parseExpression()
		if p.curr.typ != tokenRParen {
			p.errs = append(p.errs, errors.New("expected ')'"))
			return expr
		}
		p.next()
		return expr
	case tokenIdent:
		lit := p.curr.lit
		p.next()
		return &identNode{lit: lit}
	default:
		p.errs = append(p.errs, fmt.Errorf("unexpected token: %s", p.curr.lit))
		return nil
	}
}

// Parse converts a filter expression string into an AST.
func Parse(input string) (Node, error) {
	if strings.TrimSpace(input) == "" {
		return nil, nil // Empty filter allows everything
	}
	p := newParser(input)
	node := p.parseExpression()
	if len(p.errs) > 0 {
		return nil, p.errs[0]
	}
	if p.curr.typ != tokenEOF {
		return nil, fmt.Errorf("unexpected trailing characters: %s", p.curr.lit)
	}
	return node, nil
}

// AST Nodes
type andNode struct {
	left, right Node
}

func (n *andNode) Eval(md xattr.Metadata) bool {
	if n.left == nil || n.right == nil {
		return false
	}
	return n.left.Eval(md) && n.right.Eval(md)
}

type orNode struct {
	left, right Node
}

func (n *orNode) Eval(md xattr.Metadata) bool {
	if n.left == nil || n.right == nil {
		return false
	}
	return n.left.Eval(md) || n.right.Eval(md)
}

type notNode struct {
	expr Node
}

func (n *notNode) Eval(md xattr.Metadata) bool {
	if n.expr == nil {
		return false
	}
	return !n.expr.Eval(md)
}

type identNode struct {
	lit string
}

func (n *identNode) Eval(md xattr.Metadata) bool {
	switch n.lit {
	case "xdg":
		return md.HasXDG
	case "has:tags":
		return md.HasTags
	case "has:comment":
		return md.HasCmnt
	}

	if strings.HasPrefix(n.lit, "tag:") {
		t := strings.TrimPrefix(n.lit, "tag:")
		for _, tag := range md.Tags {
			if tag == t || strings.Contains(tag, t) {
				return true
			}
		}
		return false
	}
	if strings.HasPrefix(n.lit, "xattr:") {
		t := strings.TrimPrefix(n.lit, "xattr:")
		_, ok := md.All[t]
		return ok
	}
	if strings.HasPrefix(n.lit, "xdg:") {
		t := strings.TrimPrefix(n.lit, "xdg:")
		for k := range md.XDG {
			if strings.HasPrefix(k, "user.xdg."+t) {
				return true
			}
		}
		return false
	}

	return false
}
