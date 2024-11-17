package ast

import (
	"fmt"
	"os"

	"velox.eparker.dev/src/tokenizer"
)

type ASTNodeType int

const (
	Invalid ASTNodeType = iota
	Program
	PreprocessorDirective
	Block
	Statement
	Expression
	Assignment
	ReturnStatement
	FunctionDeclaration
	VariableDeclaration
	ClassDeclaration
	ConstructorDeclaration
	ArrayInitializer
	Parameters
	WhileStatement
	BinaryExpression
	UnaryExpression
	PostfixExpression
	ArrayAccess
	MemberAccess
	MacroExpansion
	Literal
	Identifier
	FunctionCall
)

var ASTNodeTypeNames map[ASTNodeType]string = map[ASTNodeType]string{
	Invalid:                "Invalid",
	Program:                "Program",
	PreprocessorDirective:  "PreprocessorDirective",
	Block:                  "Block",
	Statement:              "Statement",
	Expression:             "Expression",
	Assignment:             "Assignment",
	ReturnStatement:        "ReturnStatement",
	FunctionDeclaration:    "FunctionDeclaration",
	VariableDeclaration:    "VariableDeclaration",
	ClassDeclaration:       "ClassDeclaration",
	ConstructorDeclaration: "ConstructorDeclaration",
	ArrayInitializer:       "ArrayInitializer",
	Parameters:             "Parameters",
	WhileStatement:         "WhileStatement",
	BinaryExpression:       "BinaryExpression",
	UnaryExpression:        "UnaryExpression",
	PostfixExpression:      "PostfixExpression",
	ArrayAccess:            "ArrayAccess",
	MemberAccess:           "MemberAccess",
	MacroExpansion:         "MacroExpansion",
	Literal:                "Literal",
	Identifier:             "Identifier",
	FunctionCall:           "FunctionCall",
}

type Parser struct {
	tokens      []tokenizer.Token
	current     int
	debugMode   bool
	prattParser *PrattParser
}

func NewParser(tokens []tokenizer.Token, debugMode bool) *Parser {
	out := &Parser{
		tokens:      tokens,
		debugMode:   debugMode,
		current:     0,
		prattParser: NewPrattParser(tokens),
	}

	out.prattParser.parser = out

	return out
}

func (p *Parser) Parse() *ASTNode {
	program := &ASTNode{Type: Program, Children: []*ASTNode{}, Name: "Program"}

	for p.current < len(p.tokens) {
		token := p.Peek()

		switch token.Type {
		case tokenizer.Preprocessor:
			program.Children = append(program.Children, p.ParsePreprocessorDirective())
		case tokenizer.Keyword:
			switch token.Value {
			case "int", "float", "char", "void":
				program.Children = append(program.Children, p.ParseFunctionDeclaration())
			default:
				p.UnexpectedError(token)
			}
		default:
			p.UnexpectedError(token)
		}
	}

	return program
}

func (p *Parser) Error(format string, tokens ...tokenizer.Token) {
	if len(tokens) > 0 {
		token := tokens[0]
		fmt.Printf("Line %d, Column %d: %s\n", token.Line, token.Column, format)
	} else {
		fmt.Println(format)
	}

	if p.debugMode {
		panic("Parser error")
	}

	os.Exit(1)
}

func (p *Parser) UnexpectedError(token tokenizer.Token) {
	p.Error(fmt.Sprintf("Unexpected fucc given %s", token.String()), token)
}

func (p *Parser) ExpectedError(format string, token tokenizer.Token) {
	p.Error(fmt.Sprintf("Expected fuccing %s, got %s, an unexpected fucc", format, token.String()), token)
}

func (p *Parser) Peek() tokenizer.Token {
	if p.current >= len(p.tokens) {
		return tokenizer.Token{}
	}

	return p.tokens[p.current]
}

func (p *Parser) PeekNext() tokenizer.Token {
	if p.current+1 >= len(p.tokens) {
		return tokenizer.Token{}
	}

	return p.tokens[p.current+1]
}

func (p *Parser) Consume() tokenizer.Token {
	token := p.Peek()
	p.current++
	return token
}

func (p *Parser) Match(tokenType tokenizer.TokenType) bool {
	token := p.Peek()
	return token.Type == tokenType
}

func (p *Parser) MatchValue(tokenType tokenizer.TokenType, value string) bool {
	token := p.Peek()
	return token.Type == tokenType && token.Value == value
}

func (p *Parser) Expect(tokenType tokenizer.TokenType) tokenizer.Token {
	token := p.Consume()

	if token.Type != tokenType {
		p.ExpectedError(tokenizer.TokenTypeNames[tokenType], token)
	}

	return token
}

func (p *Parser) ExpectValue(tokenType tokenizer.TokenType, value string) tokenizer.Token {
	token := p.Consume()

	if token.Type != tokenType || token.Value != value {
		p.ExpectedError(fmt.Sprintf("%s(%s)", tokenizer.TokenTypeNames[tokenType], value), token)
	}

	return token
}

func (p *Parser) ParsePreprocessorDirective() *ASTNode {
	node := &ASTNode{Type: PreprocessorDirective}

	node.Name = p.Expect(tokenizer.Preprocessor).Value
	child := p.Expect(tokenizer.Identifier)
	node.Children = append(node.Children, &ASTNode{Type: Identifier, Name: child.Value})
	if p.Match(tokenizer.Number) {
		child := p.Expect(tokenizer.Number)
		node.Children = append(node.Children, &ASTNode{Type: Literal, Name: child.Value})
	} else if p.Match(tokenizer.String) {
		child := p.Expect(tokenizer.String)
		node.Children = append(node.Children, &ASTNode{Type: Literal, Name: child.Value})
	}

	return node
}

func (p *Parser) ParseExpression() *ASTNode {
	return p.prattParser.Parse()
}

func (p *Parser) ParseBinaryExpression(precedence int) *ASTNode {
	left := p.ParseUnary()

	for {
		op := p.Peek()
		opPrecedence := getPrecedence(op)

		if opPrecedence < precedence {
			break
		}

		p.Consume() // consume the operator

		right := p.ParseBinaryExpression(opPrecedence + 1)

		left = &ASTNode{
			Type:     BinaryExpression,
			Name:     op.Value,
			Children: []*ASTNode{left, right},
		}
	}

	return left
}

func (p *Parser) ParseUnary() *ASTNode {
	if p.Match(tokenizer.Operator) {
		op := p.Consume()
		fmt.Println(op.String())
		return &ASTNode{
			Type:     UnaryExpression,
			Name:     op.Value,
			Children: []*ASTNode{p.ParseUnary()},
		}
	}

	return p.ParsePrimary()
}

func (p *Parser) ParsePrimary() *ASTNode {
	if p.Match(tokenizer.Number) {
		return &ASTNode{Type: Literal, Name: p.Consume().Value}
	} else if p.Match(tokenizer.Identifier) {
		return &ASTNode{Type: Identifier, Name: p.Consume().Value}
	} else if p.MatchValue(tokenizer.Punctuation, "(") {
		p.Consume() // consume '('
		node := p.ParseExpression()
		p.ExpectValue(tokenizer.Punctuation, ")")
		return node
	} else {
		p.UnexpectedError(p.Peek())
		return nil
	}
}

func (p *Parser) ParseFunctionDeclaration() *ASTNode {
	returnType := p.Expect(tokenizer.Keyword) // Expect return type (e.g., "int")
	name := p.Expect(tokenizer.Identifier)    // Expect function name
	node := &ASTNode{
		Type: FunctionDeclaration,
		Name: name.Value,
		Children: []*ASTNode{
			{Type: Identifier, Name: returnType.Value}, // Store return type
		},
	}

	p.ExpectValue(tokenizer.Punctuation, "(")
	params := p.ParseParameters()
	node.Children = append(node.Children, params)
	p.ExpectValue(tokenizer.Punctuation, ")")

	// Parse function body
	body := p.ParseBlock()
	node.Children = append(node.Children, body)

	return node
}

func (p *Parser) ParseParameters() *ASTNode {
	params := &ASTNode{Type: Parameters, Name: "Parameters"}

	for !p.MatchValue(tokenizer.Punctuation, ")") {
		if p.Match(tokenizer.Keyword) {
			paramType := p.Consume()
			paramName := p.Expect(tokenizer.Identifier)
			param := &ASTNode{
				Type: VariableDeclaration,
				Name: paramName.Value,
				Children: []*ASTNode{
					{Type: Identifier, Name: paramType.Value},
				},
			}
			params.Children = append(params.Children, param)

			if p.MatchValue(tokenizer.Punctuation, ",") {
				p.Consume()
			} else if !p.MatchValue(tokenizer.Punctuation, ")") {
				p.ExpectedError("',' or ')'", p.Peek())
			}
		} else if p.MatchValue(tokenizer.Punctuation, ")") {
			break // Allow empty parameter list
		} else {
			p.ExpectedError("parameter type or ')'", p.Peek())
		}
	}

	return params
}

func (p *Parser) ParseBlock() *ASTNode {
	node := &ASTNode{Type: Block, Name: "Block"}

	p.ExpectValue(tokenizer.Punctuation, "{")

	for !p.MatchValue(tokenizer.Punctuation, "}") {
		if p.current >= len(p.tokens) {
			p.Error("Unexpected end of input while parsing block")
			return node
		}
		stmt := p.ParseStatement()
		if stmt != nil {
			node.Children = append(node.Children, stmt)
		}
	}

	p.ExpectValue(tokenizer.Punctuation, "}")

	return node
}

var assignmentOperators = []string{"=", "+=", "-=", "*=", "/=", "%=", "&=", "|=", "^=", "<<=", ">>="}

func (p *Parser) ParseStatement() *ASTNode {
	if p.Match(tokenizer.Keyword) {
		switch p.Peek().Value {
		case "return":
			return p.ParseReturnStatement()
		case "int", "float", "char":
			return p.ParseVariableDeclaration()
		case "if":
			return p.ParseConditional()
		case "while":
			return p.ParseWhileStatement()
		case "continue", "break":
			return p.ParseControlFlow()
		default:
			p.UnexpectedError(p.Peek())
		}
	}

	if p.Match(tokenizer.Identifier) && p.PeekNext().Value == "(" && p.PeekNext().Type == tokenizer.Punctuation {
		return p.ParseFunctionCall()
	}

	if p.Match(tokenizer.Identifier) {
		for _, op := range assignmentOperators {
			if p.PeekNext().Value == op {
				return p.ParseAssignment()
			}
		}
	}

	p.UnexpectedError(p.Peek())
	return nil
}

func (p *Parser) ParseVariableDeclaration() *ASTNode {
	vType := p.Expect(tokenizer.Keyword) // Expect type (e.g., "int", "float", "char")
	arrayType := false
	if p.MatchValue(tokenizer.Punctuation, "[") {
		p.Consume()                               // Consume "["
		p.ExpectValue(tokenizer.Punctuation, "]") // Expect "]"
		arrayType = true
	}

	name := p.Expect(tokenizer.Identifier) // Expect variable name

	node := &ASTNode{Type: VariableDeclaration, Name: name.Value}
	typeNode := &ASTNode{Type: Identifier, Name: vType.Value}
	if arrayType {
		typeNode.Name += "[]" // Mark as an array type
	}
	node.Children = append(node.Children, typeNode)

	if p.MatchValue(tokenizer.Operator, "=") {
		p.Consume() // Consume "="
		if arrayType {
			node.Children = append(node.Children, p.ParseArrayInitializer())
		} else {
			node.Children = append(node.Children, p.ParseExpression())
		}
	}

	p.ExpectValue(tokenizer.Punctuation, ";")
	return node
}

func (p *Parser) ParseArrayInitializer() *ASTNode {
	node := &ASTNode{Type: ArrayInitializer, Name: "ArrayInitializer"}

	p.ExpectValue(tokenizer.Punctuation, "{")
	for !p.MatchValue(tokenizer.Punctuation, "}") {
		node.Children = append(node.Children, p.ParseExpression())

		if p.MatchValue(tokenizer.Punctuation, ",") {
			p.Consume()
		} else if !p.MatchValue(tokenizer.Punctuation, "}") {
			p.ExpectedError("',' or '}'", p.Peek())
		}
	}
	p.ExpectValue(tokenizer.Punctuation, "}")

	return node
}

func (p *Parser) ParseReturnStatement() *ASTNode {
	p.Expect(tokenizer.Keyword) // Consume "return"
	node := &ASTNode{Type: ReturnStatement, Name: "return"}
	node.Children = append(node.Children, p.ParseExpression())
	p.ExpectValue(tokenizer.Punctuation, ";")
	return node
}

func (p *Parser) ParseFunctionCall() *ASTNode {
	name := p.Expect(tokenizer.Identifier)
	node := &ASTNode{Type: FunctionCall, Name: name.Value}

	p.ExpectValue(tokenizer.Punctuation, "(")
	for !p.MatchValue(tokenizer.Punctuation, ")") {
		node.Children = append(node.Children, p.ParseExpression())

		if p.MatchValue(tokenizer.Punctuation, ",") {
			p.Consume()
		} else if !p.MatchValue(tokenizer.Punctuation, ")") {
			p.ExpectedError("',' or ')'", p.Peek())
		}
	}
	p.ExpectValue(tokenizer.Punctuation, ")")
	p.ExpectValue(tokenizer.Punctuation, ";")

	return node
}

// { type: conditional, children: [condition, block, recursive conditional for else if or else ]}
func (p *Parser) ParseConditional() *ASTNode {
	node := &ASTNode{Type: Statement, Name: "if"}

	p.ExpectValue(tokenizer.Keyword, "if")
	p.ExpectValue(tokenizer.Punctuation, "(")
	node.Children = append(node.Children, p.ParseExpression())
	p.ExpectValue(tokenizer.Punctuation, ")")

	node.Children = append(node.Children, p.ParseBlock())

	if p.MatchValue(tokenizer.Keyword, "else") {
		p.Consume()

		if p.MatchValue(tokenizer.Keyword, "if") {
			node.Children = append(node.Children, p.ParseConditional())
		} else {
			node.Children = append(node.Children, p.ParseBlock())
		}
	}

	return node
}

func (p *Parser) ParseAssignment() *ASTNode {
	name := p.Expect(tokenizer.Identifier)
	node := &ASTNode{Type: Assignment, Name: name.Value}

	op := p.Expect(tokenizer.Operator)
	node.Children = append(node.Children, &ASTNode{Type: Identifier, Name: op.Value})
	node.Children = append(node.Children, p.ParseExpression())

	p.ExpectValue(tokenizer.Punctuation, ";")

	return node
}

func (p *Parser) ParseWhileStatement() *ASTNode {
	node := &ASTNode{Type: WhileStatement, Name: "while"}

	p.ExpectValue(tokenizer.Keyword, "while")
	p.ExpectValue(tokenizer.Punctuation, "(")
	node.Children = append(node.Children, p.ParseExpression())
	p.ExpectValue(tokenizer.Punctuation, ")")

	node.Children = append(node.Children, p.ParseBlock())

	return node
}

func (p *Parser) ParseControlFlow() *ASTNode {
	node := &ASTNode{Type: Statement, Name: p.Peek().Value}
	p.Consume()
	p.ExpectValue(tokenizer.Punctuation, ";")
	return node
}
