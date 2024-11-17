package ast

import (
	"velox.eparker.dev/src/tokenizer"
)

type PrefixParseFn func() *ASTNode
type InfixParseFn func(*ASTNode) *ASTNode

type PrattParser struct {
	parser         *Parser
	tokens         []tokenizer.Token
	current        int
	prefixParseFns map[tokenizer.TokenType]PrefixParseFn
	infixParseFns  map[tokenizer.TokenType]InfixParseFn
}

func NewPrattParser(tokens []tokenizer.Token) *PrattParser {
	p := &PrattParser{
		tokens:         tokens,
		current:        0,
		prefixParseFns: make(map[tokenizer.TokenType]PrefixParseFn),
		infixParseFns:  make(map[tokenizer.TokenType]InfixParseFn),
	}

	p.registerParseFns()
	return p
}

func (p *PrattParser) registerParseFns() {
	p.prefixParseFns[tokenizer.Number] = p.parseNumberLiteral
	p.prefixParseFns[tokenizer.Identifier] = p.parseIdentifier
	p.prefixParseFns[tokenizer.Punctuation] = p.parseGroupedExpression

	p.infixParseFns[tokenizer.Operator] = p.parseInfixExpression
}

func (p *PrattParser) parseGroupedExpression() *ASTNode {
	p.consumeToken() // consume '('
	exp := p.parseExpression(0)
	p.expectToken(tokenizer.Punctuation, ")") // consume ')'
	return exp
}

func (p *PrattParser) peekToken() tokenizer.Token {
	if p.current >= len(p.tokens) {
		return tokenizer.Token{Type: tokenizer.Invalid, Value: ""}
	}
	return p.tokens[p.current]
}

func (p *PrattParser) consumeToken() tokenizer.Token {
	token := p.peekToken()
	p.current++
	return token
}

func (p *PrattParser) expectToken(tokenType tokenizer.TokenType, value string) {
	if p.peekToken().Type != tokenType || p.peekToken().Value != value {
		p.parser.UnexpectedError(p.peekToken())
	}
	p.consumeToken()
}

func (p *PrattParser) peekPrecedence() int {
	if p.current >= len(p.tokens) {
		return 0
	}
	return getPrecedence(p.tokens[p.current])
}

func getPrecedence(token tokenizer.Token) int {
	switch token.Type {
	case tokenizer.Operator:
		switch token.Value {
		case "**":
			return 8
		case "*", "/", "%":
			return 7
		case "+", "-":
			return 6
		case "<", ">", "<=", ">=":
			return 5
		case "==", "!=":
			return 4
		case "&&":
			return 3
		case "||":
			return 2
		case "=", "+=", "-=", "*=", "/=":
			return 1
		}
	}
	return 0
}

func (p *PrattParser) Parse() *ASTNode {
	p.current = p.parser.current
	expr := p.parseExpression(0)
	p.parser.current = p.current
	return expr
}

func (p *PrattParser) parseExpression(precedence int) *ASTNode {
	prefix := p.prefixParseFns[p.peekToken().Type]
	if prefix == nil {
		p.parser.UnexpectedError(p.peekToken())
		return nil
	}

	leftExp := prefix()

	for precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken().Type]
		if infix == nil {
			return leftExp
		}
		leftExp = infix(leftExp)
	}

	return leftExp
}

func (p *PrattParser) parseNumberLiteral() *ASTNode {
	token := p.consumeToken()
	return &ASTNode{Type: Literal, Name: token.Value}
}

func (p *PrattParser) parseIdentifier() *ASTNode {
	token := p.consumeToken()

	// Check if the identifier is a function call
	if p.peekToken().Type == tokenizer.Punctuation && p.peekToken().Value == "(" {
		p.consumeToken() // consume '('
		node := &ASTNode{Type: FunctionCall, Name: token.Value}
		for p.peekToken().Type != tokenizer.Punctuation || p.peekToken().Value != ")" {
			node.Children = append(node.Children, p.parseExpression(0))
			if p.peekToken().Type == tokenizer.Punctuation && p.peekToken().Value == "," {
				p.consumeToken() // consume ','
			}
		}
		p.expectToken(tokenizer.Punctuation, ")") // consume ')'
		return node
	}

	return &ASTNode{Type: Identifier, Name: token.Value}
}

func (p *PrattParser) parseInfixExpression(left *ASTNode) *ASTNode {
	token := p.consumeToken()
	precedence := getPrecedence(token)
	right := p.parseExpression(precedence)
	return &ASTNode{
		Type:     BinaryExpression,
		Name:     token.Value,
		Children: []*ASTNode{left, right},
	}
}
