// Package parser implements an AST parser for mini FreeLang
package parser

import (
	"fmt"

	"github.com/user/freelang-evolving-compiler/internal/ast"
	"github.com/user/freelang-evolving-compiler/internal/lexer"
)

type Parser struct {
	l         *lexer.Lexer
	curToken  ast.Token
	peekToken ast.Token
}

// New creates a new parser for the input string
func New(input string) *Parser {
	l := lexer.New(input)
	p := &Parser{l: l}
	p.nextToken()
	p.nextToken()
	return p
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

// ParseProgram parses a complete program
func (p *Parser) ParseProgram() (*ast.Node, error) {
	prog := &ast.Node{
		Kind:     ast.NodeProgram,
		Children: []*ast.Node{},
	}

	for p.curToken.Type != ast.TokenEOF {
		stmt, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		if stmt != nil {
			prog.Children = append(prog.Children, stmt)
		}
		p.skipSemicolons()
	}

	return prog, nil
}

func (p *Parser) parseStatement() (*ast.Node, error) {
	switch p.curToken.Type {
	case ast.TokenLet:
		return p.parseLetDecl()
	case ast.TokenFn:
		return p.parseFnDecl()
	case ast.TokenIf:
		return p.parseIfStmt()
	case ast.TokenFor:
		return p.parseForStmt()
	case ast.TokenReturn:
		return p.parseReturnStmt()
	case ast.TokenStruct:
		return p.parseStructDecl()
	case ast.TokenLBrace:
		return p.parseBlockStmt()
	case ast.TokenSemicolon:
		p.nextToken()
		return nil, nil
	default:
		// Try to parse as expression statement
		expr, err := p.parseExpression(0)
		if err != nil {
			return nil, err
		}
		return expr, nil
	}
}

func (p *Parser) parseLetDecl() (*ast.Node, error) {
	letNode := &ast.Node{
		Kind: ast.NodeLetDecl,
		Line: p.curToken.Line,
		Col:  p.curToken.Col,
	}

	p.nextToken() // consume 'let'

	if p.curToken.Type != ast.TokenIdent {
		return nil, fmt.Errorf("expected identifier after 'let'")
	}

	nameNode := &ast.Node{
		Kind:  ast.NodeIdent,
		Value: p.curToken.Value,
		Line:  p.curToken.Line,
		Col:   p.curToken.Col,
	}
	letNode.Children = append(letNode.Children, nameNode)

	p.nextToken() // consume identifier

	// Check for type annotation: let x: int = ...
	if p.curToken.Type == ast.TokenColon {
		p.nextToken() // consume ':'
		if p.curToken.Type != ast.TokenIdent {
			return nil, fmt.Errorf("expected type name after ':'")
		}
		nameNode.TypeAnnotation = p.curToken.Value // "int", "bool", "Point", etc.
		p.nextToken()                              // consume type name
	}

	if p.curToken.Type != ast.TokenAssign {
		return nil, fmt.Errorf("expected '=' after identifier")
	}

	p.nextToken() // consume '='

	expr, err := p.parseExpression(0)
	if err != nil {
		return nil, err
	}
	letNode.Children = append(letNode.Children, expr)

	return letNode, nil
}

func (p *Parser) parseFnDecl() (*ast.Node, error) {
	fnNode := &ast.Node{
		Kind: ast.NodeFnDecl,
		Line: p.curToken.Line,
		Col:  p.curToken.Col,
	}

	p.nextToken() // consume 'fn'

	if p.curToken.Type != ast.TokenIdent {
		return nil, fmt.Errorf("expected function name")
	}

	fnNode.Value = p.curToken.Value
	p.nextToken()

	if p.curToken.Type != ast.TokenLParen {
		return nil, fmt.Errorf("expected '(' after function name")
	}

	p.nextToken()

	// Parse parameters with type annotations
	for p.curToken.Type != ast.TokenRParen {
		if p.curToken.Type != ast.TokenIdent {
			return nil, fmt.Errorf("expected parameter name")
		}
		param := &ast.Node{
			Kind:  ast.NodeIdent,
			Value: p.curToken.Value,
		}
		p.nextToken()

		// Parse parameter type annotation if present
		if p.curToken.Type == ast.TokenColon {
			p.nextToken()
			if p.curToken.Type == ast.TokenIdent {
				param.TypeAnnotation = p.curToken.Value // Store type annotation
				p.nextToken()
			}
		}

		fnNode.Children = append(fnNode.Children, param)

		if p.curToken.Type == ast.TokenComma {
			p.nextToken()
		}
	}

	p.nextToken() // consume ')'

	// Parse return type annotation if present: fn add(...): int
	if p.curToken.Type == ast.TokenColon {
		p.nextToken()
		if p.curToken.Type == ast.TokenIdent {
			fnNode.TypeAnnotation = p.curToken.Value // Store return type
			p.nextToken()
		}
	}

	if p.curToken.Type != ast.TokenLBrace {
		return nil, fmt.Errorf("expected '{' for function body")
	}

	body, err := p.parseBlockStmt()
	if err != nil {
		return nil, err
	}
	fnNode.Children = append(fnNode.Children, body)

	return fnNode, nil
}

func (p *Parser) parseIfStmt() (*ast.Node, error) {
	ifNode := &ast.Node{
		Kind: ast.NodeIfStmt,
		Line: p.curToken.Line,
		Col:  p.curToken.Col,
	}

	p.nextToken() // consume 'if'

	cond, err := p.parseExpression(0)
	if err != nil {
		return nil, err
	}
	ifNode.Children = append(ifNode.Children, cond)

	if p.curToken.Type != ast.TokenLBrace {
		return nil, fmt.Errorf("expected '{' after if condition")
	}

	body, err := p.parseBlockStmt()
	if err != nil {
		return nil, err
	}
	ifNode.Children = append(ifNode.Children, body)

	// Check for else branch
	if p.curToken.Type == ast.TokenElse {
		p.nextToken() // consume 'else'

		if p.curToken.Type != ast.TokenLBrace {
			return nil, fmt.Errorf("expected '{' after else")
		}

		elseBody, err := p.parseBlockStmt()
		if err != nil {
			return nil, err
		}
		ifNode.Children = append(ifNode.Children, elseBody)
	}

	return ifNode, nil
}

func (p *Parser) parseForStmt() (*ast.Node, error) {
	forNode := &ast.Node{
		Kind: ast.NodeForStmt,
		Line: p.curToken.Line,
		Col:  p.curToken.Col,
	}

	p.nextToken() // consume 'for'

	if p.curToken.Type != ast.TokenIdent {
		return nil, fmt.Errorf("expected iterator variable")
	}

	iter := &ast.Node{
		Kind:  ast.NodeIdent,
		Value: p.curToken.Value,
	}
	forNode.Children = append(forNode.Children, iter)
	p.nextToken()

	if p.curToken.Type != ast.TokenIn {
		return nil, fmt.Errorf("expected 'in' after iterator")
	}

	p.nextToken()

	rng, err := p.parseExpression(0)
	if err != nil {
		return nil, err
	}
	forNode.Children = append(forNode.Children, rng)

	if p.curToken.Type != ast.TokenLBrace {
		return nil, fmt.Errorf("expected '{' for loop body")
	}

	body, err := p.parseBlockStmt()
	if err != nil {
		return nil, err
	}
	forNode.Children = append(forNode.Children, body)

	return forNode, nil
}

func (p *Parser) parseReturnStmt() (*ast.Node, error) {
	retNode := &ast.Node{
		Kind: ast.NodeReturn,
		Line: p.curToken.Line,
		Col:  p.curToken.Col,
	}

	p.nextToken() // consume 'return'

	if p.curToken.Type == ast.TokenSemicolon || p.curToken.Type == ast.TokenRBrace {
		return retNode, nil
	}

	expr, err := p.parseExpression(0)
	if err != nil {
		return nil, err
	}
	retNode.Children = append(retNode.Children, expr)

	return retNode, nil
}

func (p *Parser) parseBlockStmt() (*ast.Node, error) {
	block := &ast.Node{
		Kind: ast.NodeBlockStmt,
		Line: p.curToken.Line,
		Col:  p.curToken.Col,
	}

	p.nextToken() // consume '{'

	for p.curToken.Type != ast.TokenRBrace && p.curToken.Type != ast.TokenEOF {
		stmt, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		if stmt != nil {
			block.Children = append(block.Children, stmt)
		}
		p.skipSemicolons()
	}

	if p.curToken.Type != ast.TokenRBrace {
		return nil, fmt.Errorf("expected '}'")
	}

	p.nextToken() // consume '}'

	return block, nil
}

func (p *Parser) parseExpression(prec int) (*ast.Node, error) {
	left, err := p.parsePrimary()
	if err != nil {
		return nil, err
	}

	for prec < p.peekPrecedence() {
		p.nextToken()
		left, err = p.parseInfix(left)
		if err != nil {
			return nil, err
		}
	}

	return left, nil
}

func (p *Parser) parsePrimary() (*ast.Node, error) {
	switch p.curToken.Type {
	case ast.TokenInt:
		node := &ast.Node{
			Kind:  ast.NodeIntLit,
			Value: p.curToken.Value,
			Line:  p.curToken.Line,
			Col:   p.curToken.Col,
		}
		return node, nil

	case ast.TokenTrue:
		node := &ast.Node{
			Kind:  ast.NodeBoolLit,
			Value: "true",
			Line:  p.curToken.Line,
			Col:   p.curToken.Col,
		}
		return node, nil

	case ast.TokenFalse:
		node := &ast.Node{
			Kind:  ast.NodeBoolLit,
			Value: "false",
			Line:  p.curToken.Line,
			Col:   p.curToken.Col,
		}
		return node, nil

	case ast.TokenString:
		node := &ast.Node{
			Kind:  ast.NodeStringLit,
			Value: p.curToken.Value,
			Line:  p.curToken.Line,
			Col:   p.curToken.Col,
		}
		return node, nil

	case ast.TokenIdent:
		name := p.curToken.Value
		node := &ast.Node{
			Kind:  ast.NodeIdent,
			Value: name,
			Line:  p.curToken.Line,
			Col:   p.curToken.Col,
		}
		// Check if next token is '{' for struct initialization
		if p.peekToken.Type == ast.TokenLBrace {
			p.nextToken() // move to '{'
			return p.parseStructLit(name)
		}
		return node, nil

	case ast.TokenLParen:
		p.nextToken()
		expr, err := p.parseExpression(0)
		if err != nil {
			return nil, err
		}
		if p.curToken.Type != ast.TokenRParen {
			return nil, fmt.Errorf("expected ')'")
		}
		return expr, nil

	case ast.TokenLBracket:
		return p.parseArrayLit()

	default:
		return nil, fmt.Errorf("unexpected token: %v", p.curToken.Type)
	}
}

func (p *Parser) parseInfix(left *ast.Node) (*ast.Node, error) {
	// Handle field access: obj.field
	if p.curToken.Type == ast.TokenDot {
		p.nextToken() // consume '.'
		if p.curToken.Type != ast.TokenIdent {
			return nil, fmt.Errorf("expected field name after '.'")
		}
		return &ast.Node{
			Kind:     ast.NodeFieldAccess,
			Value:    p.curToken.Value,
			Line:     p.curToken.Line,
			Col:      p.curToken.Col,
			Children: []*ast.Node{left},
		}, nil
	}

	// Handle array indexing: arr[i]
	if p.curToken.Type == ast.TokenLBracket {
		return p.parseIndexExpr(left)
	}

	node := &ast.Node{
		Kind:     ast.NodeBinaryExpr,
		Line:     p.curToken.Line,
		Col:      p.curToken.Col,
		Value:    p.curToken.Value,
		Children: []*ast.Node{left},
	}

	prec := p.curPrecedence()
	p.nextToken()

	right, err := p.parseExpression(prec)
	if err != nil {
		return nil, err
	}
	node.Children = append(node.Children, right)

	// Handle function calls
	if left.Kind == ast.NodeIdent && p.curToken.Type == ast.TokenLParen {
		call := &ast.Node{
			Kind:  ast.NodeCallExpr,
			Value: left.Value,
		}
		p.nextToken() // consume '('
		for p.curToken.Type != ast.TokenRParen {
			arg, err := p.parseExpression(0)
			if err != nil {
				return nil, err
			}
			call.Children = append(call.Children, arg)
			if p.curToken.Type == ast.TokenComma {
				p.nextToken()
			}
		}
		p.nextToken() // consume ')'
		return call, nil
	}

	return node, nil
}

func (p *Parser) peekPrecedence() int {
	return precedence(p.peekToken.Type)
}

func (p *Parser) curPrecedence() int {
	return precedence(p.curToken.Type)
}

func precedence(tok ast.TokenType) int {
	switch tok {
	case ast.TokenEq, ast.TokenNe, ast.TokenLt, ast.TokenGt, ast.TokenLe, ast.TokenGe:
		return 1
	case ast.TokenPlus, ast.TokenMinus:
		return 2
	case ast.TokenStar, ast.TokenSlash:
		return 3
	case ast.TokenDotDot:
		return 4
	case ast.TokenDot, ast.TokenLBracket:
		return 5
	default:
		return 0
	}
}

func (p *Parser) skipSemicolons() {
	for p.curToken.Type == ast.TokenSemicolon {
		p.nextToken()
	}
}

func (p *Parser) parseStructDecl() (*ast.Node, error) {
	structNode := &ast.Node{
		Kind: ast.NodeStructDecl,
		Line: p.curToken.Line,
		Col:  p.curToken.Col,
	}

	p.nextToken() // consume 'struct'

	if p.curToken.Type != ast.TokenIdent {
		return nil, fmt.Errorf("expected struct name")
	}

	structNode.Value = p.curToken.Value
	p.nextToken()

	if p.curToken.Type != ast.TokenLBrace {
		return nil, fmt.Errorf("expected '{' after struct name")
	}

	p.nextToken() // consume '{'

	// Parse fields
	for p.curToken.Type != ast.TokenRBrace {
		field, err := p.parseFieldDecl()
		if err != nil {
			return nil, err
		}
		structNode.Children = append(structNode.Children, field)

		if p.curToken.Type == ast.TokenSemicolon {
			p.nextToken()
		}
	}

	p.nextToken() // consume '}'

	return structNode, nil
}

func (p *Parser) parseFieldDecl() (*ast.Node, error) {
	if p.curToken.Type != ast.TokenIdent {
		return nil, fmt.Errorf("expected field name")
	}

	fieldNode := &ast.Node{
		Kind:  ast.NodeFieldDecl,
		Value: p.curToken.Value,
		Line:  p.curToken.Line,
		Col:   p.curToken.Col,
	}

	p.nextToken() // consume field name

	if p.curToken.Type != ast.TokenColon {
		return nil, fmt.Errorf("expected ':' after field name")
	}

	p.nextToken() // consume ':'

	if p.curToken.Type != ast.TokenIdent {
		return nil, fmt.Errorf("expected field type")
	}

	typeNode := &ast.Node{
		Kind:  ast.NodeIdent,
		Value: p.curToken.Value,
		Line:  p.curToken.Line,
		Col:   p.curToken.Col,
	}
	fieldNode.TypeAnnotation = p.curToken.Value // Store type annotation directly
	fieldNode.Children = append(fieldNode.Children, typeNode)

	p.nextToken() // consume type

	return fieldNode, nil
}

// parseStructLit parses struct initialization: Point{x: 1, y: 2}
func (p *Parser) parseStructLit(name string) (*ast.Node, error) {
	structLit := &ast.Node{
		Kind:  ast.NodeStructLit,
		Value: name,
		Line:  p.curToken.Line,
		Col:   p.curToken.Col,
	}

	p.nextToken() // consume '{'

	for p.curToken.Type != ast.TokenRBrace && p.curToken.Type != ast.TokenEOF {
		if p.curToken.Type != ast.TokenIdent {
			return nil, fmt.Errorf("expected field name in struct init")
		}

		fieldName := p.curToken.Value
		p.nextToken()

		if p.curToken.Type != ast.TokenColon {
			return nil, fmt.Errorf("expected ':' after field name in struct init")
		}

		p.nextToken()

		// Parse field value expression
		fieldValue, err := p.parseExpression(0)
		if err != nil {
			return nil, err
		}

		// Create field initialization node
		fieldInit := &ast.Node{
			Kind:  ast.NodeFieldDecl,
			Value: fieldName,
		}
		fieldInit.Children = append(fieldInit.Children, fieldValue)
		structLit.Children = append(structLit.Children, fieldInit)

		if p.curToken.Type == ast.TokenComma {
			p.nextToken()
		}
	}

	if p.curToken.Type != ast.TokenRBrace {
		return nil, fmt.Errorf("expected '}' to close struct init")
	}

	p.nextToken() // consume '}'

	return structLit, nil
}

// parseArrayLit parses array literals: [1, 2, 3]
func (p *Parser) parseArrayLit() (*ast.Node, error) {
	node := &ast.Node{
		Kind:  ast.NodeArrayLit,
		Line:  p.curToken.Line,
		Col:   p.curToken.Col,
		Children: []*ast.Node{},
	}

	p.nextToken() // skip '['

	for p.curToken.Type != ast.TokenRBracket && p.curToken.Type != ast.TokenEOF {
		elem, err := p.parseExpression(0)
		if err != nil {
			return nil, err
		}
		node.Children = append(node.Children, elem)

		if p.curToken.Type == ast.TokenComma {
			p.nextToken()
		} else if p.curToken.Type != ast.TokenRBracket {
			return nil, fmt.Errorf("expected ',' or ']' in array literal")
		}
	}

	if p.curToken.Type != ast.TokenRBracket {
		return nil, fmt.Errorf("expected ']' to close array literal")
	}

	p.nextToken() // skip ']'
	return node, nil
}

// parseIndexExpr parses array indexing: arr[i]
func (p *Parser) parseIndexExpr(obj *ast.Node) (*ast.Node, error) {
	node := &ast.Node{
		Kind:     ast.NodeIndexExpr,
		Line:     obj.Line,
		Col:      obj.Col,
		Children: []*ast.Node{obj},
	}

	p.nextToken() // skip '['

	idx, err := p.parseExpression(0)
	if err != nil {
		return nil, err
	}
	node.Children = append(node.Children, idx)

	if p.curToken.Type != ast.TokenRBracket {
		return nil, fmt.Errorf("expected ']' after array index")
	}

	p.nextToken() // skip ']'
	return node, nil
}
