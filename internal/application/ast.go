package application

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

type ASTNode struct {
	IsLeaf   bool
	Value    float64
	Operator string
	Left     *ASTNode
	Right    *ASTNode
}

type parser struct {
	input string
	pos   int
}

func ParseAST(expr string) (*ASTNode, error) {
	expr = strings.ReplaceAll(expr, " ", "")
	p := &parser{input: expr}
	ast, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	if p.pos != len(p.input) {
		return nil, fmt.Errorf("unexpected input at %d", p.pos)
	}
	return ast, nil
}

func (p *parser) parseExpression() (*ASTNode, error) {
	return p.parseBinaryOp(p.parseTerm, []string{"+", "-"})
}

func (p *parser) parseTerm() (*ASTNode, error) {
	return p.parseBinaryOp(p.parseFactor, []string{"*", "/"})
}

func (p *parser) parseBinaryOp(next func() (*ASTNode, error), ops []string) (*ASTNode, error) {
	node, err := next()
	if err != nil {
		return nil, err
	}
	for {
		if p.pos >= len(p.input) {
			break
		}
		matched := ""
		for _, op := range ops {
			if string(p.input[p.pos]) == op {
				matched = op
				break
			}
		}
		if matched == "" {
			break
		}
		p.pos++
		right, err := next()
		if err != nil {
			return nil, err
		}
		node = &ASTNode{Operator: matched, Left: node, Right: right}
	}
	return node, nil
}

func (p *parser) parseFactor() (*ASTNode, error) {
	if p.pos < len(p.input) && p.input[p.pos] == '(' {
		p.pos++
		node, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		if p.pos >= len(p.input) || p.input[p.pos] != ')' {
			return nil, fmt.Errorf("missing )")
		}
		p.pos++
		return node, nil
	}
	start := p.pos
	if p.pos < len(p.input) && (p.input[p.pos] == '+' || p.input[p.pos] == '-') {
		p.pos++
	}
	for p.pos < len(p.input) && (unicode.IsDigit(rune(p.input[p.pos])) || p.input[p.pos] == '.') {
		p.pos++
	}
	numStr := p.input[start:p.pos]
	if numStr == "" {
		return nil, fmt.Errorf("expected number")
	}
	val, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid number %q", numStr)
	}
	return &ASTNode{IsLeaf: true, Value: val}, nil
}


