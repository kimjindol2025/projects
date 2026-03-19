// Package parser implements V-compatible syntax parsing
package parser

import (
	"fmt"
	"strconv"

	"fv2-lang/internal/ast"
	"fv2-lang/internal/lexer"
)

// Parser converts tokens to AST
type Parser struct {
	tokens   []lexer.Token
	pos      int
	errors   []string
}

// New creates a new parser
func New(tokens []lexer.Token) *Parser {
	return &Parser{
		tokens: tokens,
		pos:    0,
		errors: []string{},
	}
}

// Parse parses tokens into an AST
func (p *Parser) Parse() (*ast.Program, error) {
	var definitions []ast.Definition
	var mainBody []ast.Statement

	for !p.isAtEnd() {
		if p.match(lexer.TknEof) {
			break
		}

		if p.check(lexer.TknFn) {
			def, err := p.parseFunctionDef()
			if err != nil {
				return nil, err
			}
			if def != nil {
				definitions = append(definitions, def)
			}
		} else if p.check(lexer.TknType) {
			def, err := p.parseTypeDef()
			if err != nil {
				return nil, err
			}
			if def != nil {
				definitions = append(definitions, def)
			}
		} else if p.check(lexer.TknStruct) {
			def, err := p.parseStructDef()
			if err != nil {
				return nil, err
			}
			if def != nil {
				definitions = append(definitions, def)
			}
		} else {
			stmt, err := p.parseStatement()
			if err != nil {
				return nil, err
			}
			if stmt != nil {
				mainBody = append(mainBody, stmt)
			}
		}
	}

	return &ast.Program{
		Definitions: definitions,
		MainBody:    mainBody,
	}, nil
}

// parseFunctionDef parses a function definition
func (p *Parser) parseFunctionDef() (*ast.FunctionDef, error) {
	if !p.match(lexer.TknFn) {
		return nil, nil
	}

	name := p.current().Text
	if !p.match(lexer.TknIdentifier) {
		return nil, fmt.Errorf("expected function name at %d:%d", p.current().Line, p.current().Column)
	}

	if !p.match(lexer.TknLParen) {
		return nil, fmt.Errorf("expected '(' at %d:%d", p.current().Line, p.current().Column)
	}

	// Parse parameters
	var params []ast.Parameter
	if !p.check(lexer.TknRParen) {
		for {
			paramName := p.current().Text
			if !p.match(lexer.TknIdentifier) {
				return nil, fmt.Errorf("expected parameter name at %d:%d", p.current().Line, p.current().Column)
			}

			var paramType *ast.Type
			if p.match(lexer.TknColon) {
				// Type annotation
				paramType = p.parseType()
			}

			params = append(params, ast.Parameter{
				Name: paramName,
				Type: paramType,
			})

			if !p.match(lexer.TknComma) {
				break
			}
		}
	}

	if !p.match(lexer.TknRParen) {
		return nil, fmt.Errorf("expected ')' at %d:%d", p.current().Line, p.current().Column)
	}

	// Parse return type (V syntax: type directly after params, no ->)
	var returnType *ast.Type
	if !p.check(lexer.TknLBrace) && !p.isAtEnd() {
		// Look ahead to see if next is a type (identifier or keyword type)
		if p.check(lexer.TknIdentifier) || p.isPrimitiveType() {
			returnType = p.parseType()
		}
	}

	// Parse body
	if !p.match(lexer.TknLBrace) {
		return nil, fmt.Errorf("expected '{' at %d:%d", p.current().Line, p.current().Column)
	}

	var body []ast.Statement
	for !p.check(lexer.TknRBrace) && !p.isAtEnd() {
		stmt, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		if stmt != nil {
			body = append(body, stmt)
		}
	}

	if !p.match(lexer.TknRBrace) {
		return nil, fmt.Errorf("expected '}' at %d:%d", p.current().Line, p.current().Column)
	}

	return &ast.FunctionDef{
		Name:       name,
		Parameters: params,
		ReturnType: returnType,
		Body:       body,
	}, nil
}

// parseTypeDef parses a type definition
// V syntax: type UserId = i64
func (p *Parser) parseTypeDef() (*ast.TypeDef, error) {
	if !p.match(lexer.TknType) {
		return nil, nil
	}

	name := p.current().Text
	if !p.match(lexer.TknIdentifier) {
		return nil, fmt.Errorf("expected type name at %d:%d", p.current().Line, p.current().Column)
	}

	if !p.match(lexer.TknAssign) {
		return nil, fmt.Errorf("expected '=' at %d:%d", p.current().Line, p.current().Column)
	}

	// Parse the underlying type (simple type alias)
	underlyingType := p.parseType()

	return &ast.TypeDef{
		Name: name,
		Fields: []ast.Field{
			{
				Name: "value",
				Type: underlyingType,
			},
		},
	}, nil
}

// parseStructDef parses a struct definition
// V syntax: struct Point { x i64, y i64 }
// Field syntax: name type (no colon)
func (p *Parser) parseStructDef() (*ast.StructDef, error) {
	if !p.match(lexer.TknStruct) {
		return nil, nil
	}

	name := p.current().Text
	if !p.match(lexer.TknIdentifier) {
		return nil, fmt.Errorf("expected struct name at %d:%d", p.current().Line, p.current().Column)
	}

	if !p.match(lexer.TknLBrace) {
		return nil, fmt.Errorf("expected '{' at %d:%d", p.current().Line, p.current().Column)
	}

	var fields []ast.Field
	for !p.check(lexer.TknRBrace) && !p.isAtEnd() {
		isMutable := p.match(lexer.TknMut)

		fieldName := p.current().Text
		if !p.match(lexer.TknIdentifier) {
			return nil, fmt.Errorf("expected field name at %d:%d", p.current().Line, p.current().Column)
		}

		// V syntax: field type (space-separated, no colon)
		fieldType := p.parseType()

		fields = append(fields, ast.Field{
			Name:    fieldName,
			Type:    fieldType,
			Mutable: isMutable,
		})

		if !p.match(lexer.TknComma) {
			break
		}
	}

	if !p.match(lexer.TknRBrace) {
		return nil, fmt.Errorf("expected '}' at %d:%d", p.current().Line, p.current().Column)
	}

	return &ast.StructDef{
		Name:   name,
		Fields: fields,
	}, nil
}

// parseStatement parses a statement
func (p *Parser) parseStatement() (ast.Statement, error) {
	if p.match(lexer.TknLet) {
		return p.parseLetStatement()
	}
	if p.match(lexer.TknConst) {
		return p.parseConstStatement()
	}
	if p.match(lexer.TknIf) {
		return p.parseIfStatement()
	}
	if p.match(lexer.TknFor) {
		return p.parseForStatement()
	}
	if p.match(lexer.TknMatch) {
		return p.parseMatchStatement()
	}
	if p.match(lexer.TknReturn) {
		return p.parseReturnStatement()
	}
	if p.match(lexer.TknLBrace) {
		return p.parseBlockStatement()
	}

	// Expression statement
	expr, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	// Consume optional semicolon
	p.match(lexer.TknSemicolon)

	return &ast.ExpressionStatement{Expression: expr}, nil
}

// parseLetStatement parses a let binding
func (p *Parser) parseLetStatement() (ast.Statement, error) {
	isMutable := p.match(lexer.TknMut)

	name := p.current().Text
	if !p.match(lexer.TknIdentifier) {
		return nil, fmt.Errorf("expected variable name at %d:%d", p.current().Line, p.current().Column)
	}

	var varType *ast.Type
	if p.match(lexer.TknColon) {
		varType = p.parseType()
	}

	var init ast.Expression
	if p.match(lexer.TknAssign) || p.match(lexer.TknColonAssign) {
		var err error
		init, err = p.parseExpression()
		if err != nil {
			return nil, err
		}
	}

	p.match(lexer.TknSemicolon)

	return &ast.LetStatement{
		Name:    name,
		Type:    varType,
		Init:    init,
		Mutable: isMutable,
	}, nil
}

// parseConstStatement parses a const binding
func (p *Parser) parseConstStatement() (ast.Statement, error) {
	name := p.current().Text
	if !p.match(lexer.TknIdentifier) {
		return nil, fmt.Errorf("expected constant name at %d:%d", p.current().Line, p.current().Column)
	}

	var varType *ast.Type
	if p.match(lexer.TknColon) {
		varType = p.parseType()
	}

	if !p.match(lexer.TknAssign) {
		return nil, fmt.Errorf("expected '=' at %d:%d", p.current().Line, p.current().Column)
	}

	value, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	p.match(lexer.TknSemicolon)

	return &ast.ConstStatement{
		Name:  name,
		Type:  varType,
		Value: value,
	}, nil
}

// parseIfStatement parses an if statement
func (p *Parser) parseIfStatement() (ast.Statement, error) {
	cond, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	if !p.match(lexer.TknLBrace) {
		return nil, fmt.Errorf("expected '{' at %d:%d", p.current().Line, p.current().Column)
	}

	var thenBody []ast.Statement
	for !p.check(lexer.TknRBrace) && !p.isAtEnd() {
		stmt, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		if stmt != nil {
			thenBody = append(thenBody, stmt)
		}
	}

	if !p.match(lexer.TknRBrace) {
		return nil, fmt.Errorf("expected '}' at %d:%d", p.current().Line, p.current().Column)
	}

	var elseBody []ast.Statement
	if p.match(lexer.TknElse) {
		if p.match(lexer.TknLBrace) {
			for !p.check(lexer.TknRBrace) && !p.isAtEnd() {
				stmt, err := p.parseStatement()
				if err != nil {
					return nil, err
				}
				if stmt != nil {
					elseBody = append(elseBody, stmt)
				}
			}

			if !p.match(lexer.TknRBrace) {
				return nil, fmt.Errorf("expected '}' at %d:%d", p.current().Line, p.current().Column)
			}
		}
	}

	return &ast.IfStatement{
		Condition: cond,
		ThenBody:  thenBody,
		ElseBody:  elseBody,
	}, nil
}

// parseForStatement parses a for loop
func (p *Parser) parseForStatement() (ast.Statement, error) {
	varName := p.current().Text
	if !p.match(lexer.TknIdentifier) {
		return nil, fmt.Errorf("expected variable name at %d:%d", p.current().Line, p.current().Column)
	}

	if !p.match(lexer.TknIn) {
		return nil, fmt.Errorf("expected 'in' at %d:%d", p.current().Line, p.current().Column)
	}

	iter, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	if !p.match(lexer.TknLBrace) {
		return nil, fmt.Errorf("expected '{' at %d:%d", p.current().Line, p.current().Column)
	}

	var body []ast.Statement
	for !p.check(lexer.TknRBrace) && !p.isAtEnd() {
		stmt, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		if stmt != nil {
			body = append(body, stmt)
		}
	}

	if !p.match(lexer.TknRBrace) {
		return nil, fmt.Errorf("expected '}' at %d:%d", p.current().Line, p.current().Column)
	}

	// Check if iter is a range (BinaryExpression with .. operator)
	if binExpr, ok := iter.(*ast.BinaryExpression); ok && binExpr.Operator == ".." {
		return &ast.ForRangeStatement{
			Variable: varName,
			Start:    binExpr.Left,
			End:      binExpr.Right,
			Body:     body,
		}, nil
	}

	return &ast.ForStatement{
		Variable: varName,
		Iterator: iter,
		Body:     body,
	}, nil
}

// parseMatchStatement parses a match statement
func (p *Parser) parseMatchStatement() (ast.Statement, error) {
	expr, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	if !p.match(lexer.TknLBrace) {
		return nil, fmt.Errorf("expected '{' at %d:%d", p.current().Line, p.current().Column)
	}

	var arms []ast.MatchArm
	for !p.check(lexer.TknRBrace) && !p.isAtEnd() {
		// Parse pattern
		pattern, err := p.parsePattern()
		if err != nil {
			return nil, err
		}

		if !p.match(lexer.TknFatArrow) {
			return nil, fmt.Errorf("expected '=>' at %d:%d", p.current().Line, p.current().Column)
		}

		// Parse body
		var body []ast.Statement
		if p.match(lexer.TknLBrace) {
			for !p.check(lexer.TknRBrace) && !p.isAtEnd() {
				stmt, err := p.parseStatement()
				if err != nil {
					return nil, err
				}
				if stmt != nil {
					body = append(body, stmt)
				}
			}
			if !p.match(lexer.TknRBrace) {
				return nil, fmt.Errorf("expected '}' at %d:%d", p.current().Line, p.current().Column)
			}
		} else {
			stmt, err := p.parseStatement()
			if err != nil {
				return nil, err
			}
			if stmt != nil {
				body = append(body, stmt)
			}
		}

		arms = append(arms, ast.MatchArm{
			Pattern: pattern,
			Body:    body,
		})

		if !p.match(lexer.TknComma) {
			break
		}
	}

	if !p.match(lexer.TknRBrace) {
		return nil, fmt.Errorf("expected '}' at %d:%d", p.current().Line, p.current().Column)
	}

	return &ast.MatchStatement{
		Expression: expr,
		Arms:       arms,
	}, nil
}

// parseReturnStatement parses a return statement
func (p *Parser) parseReturnStatement() (ast.Statement, error) {
	var value ast.Expression
	if !p.check(lexer.TknSemicolon) && !p.check(lexer.TknRBrace) && !p.isAtEnd() {
		var err error
		value, err = p.parseExpression()
		if err != nil {
			return nil, err
		}
	}

	p.match(lexer.TknSemicolon)

	return &ast.ReturnStatement{Value: value}, nil
}

// parseBlockStatement parses a block statement
func (p *Parser) parseBlockStatement() (ast.Statement, error) {
	var statements []ast.Statement

	for !p.check(lexer.TknRBrace) && !p.isAtEnd() {
		stmt, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		if stmt != nil {
			statements = append(statements, stmt)
		}
	}

	if !p.match(lexer.TknRBrace) {
		return nil, fmt.Errorf("expected '}' at %d:%d", p.current().Line, p.current().Column)
	}

	return &ast.BlockStatement{Statements: statements}, nil
}

// parseExpression parses an expression
func (p *Parser) parseExpression() (ast.Expression, error) {
	return p.parseBinaryExpression(0)
}

// parseBinaryExpression parses binary expressions with precedence
func (p *Parser) parseBinaryExpression(minPrec int) (ast.Expression, error) {
	left, err := p.parseUnary()
	if err != nil {
		return nil, err
	}

	for {
		if p.isAtEnd() || !p.isOperator(p.current().Type) {
			break
		}

		op := p.current()
		prec := getPrecedence(op.Type)
		if prec < minPrec {
			break
		}

		p.advance()

		right, err := p.parseBinaryExpression(prec + 1)
		if err != nil {
			return nil, err
		}

		left = &ast.BinaryExpression{
			Left:     left,
			Operator: op.Text,
			Right:    right,
		}
	}

	return left, nil
}

// parseUnary parses unary expressions
func (p *Parser) parseUnary() (ast.Expression, error) {
	if p.match(lexer.TknNot) {
		operand, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		return &ast.UnaryExpression{
			Operator: "!",
			Operand:  operand,
		}, nil
	}

	if p.match(lexer.TknMinus) {
		operand, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		return &ast.UnaryExpression{
			Operator: "-",
			Operand:  operand,
		}, nil
	}

	if p.match(lexer.TknAmpersand) {
		operand, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		return &ast.UnaryExpression{
			Operator: "&",
			Operand:  operand,
		}, nil
	}

	return p.parsePostfix()
}

// parsePostfix parses postfix expressions
func (p *Parser) parsePostfix() (ast.Expression, error) {
	expr, err := p.parsePrimary()
	if err != nil {
		return nil, err
	}

	for {
		if p.match(lexer.TknLParen) {
			// Function call
			var args []ast.Expression
			if !p.check(lexer.TknRParen) {
				for {
					arg, err := p.parseExpression()
					if err != nil {
						return nil, err
					}
					args = append(args, arg)

					if !p.match(lexer.TknComma) {
						break
					}
				}
			}

			if !p.match(lexer.TknRParen) {
				return nil, fmt.Errorf("expected ')' at %d:%d", p.current().Line, p.current().Column)
			}

			expr = &ast.CallExpression{
				Function:  expr,
				Arguments: args,
			}
		} else if p.match(lexer.TknDot) {
			// Field access or method call
			fieldName := p.current().Text
			if !p.match(lexer.TknIdentifier) {
				return nil, fmt.Errorf("expected field name at %d:%d", p.current().Line, p.current().Column)
			}

			if p.match(lexer.TknLParen) {
				// Method call
				var args []ast.Expression
				if !p.check(lexer.TknRParen) {
					for {
						arg, err := p.parseExpression()
						if err != nil {
							return nil, err
						}
						args = append(args, arg)

						if !p.match(lexer.TknComma) {
							break
						}
					}
				}

				if !p.match(lexer.TknRParen) {
					return nil, fmt.Errorf("expected ')' at %d:%d", p.current().Line, p.current().Column)
				}

				expr = &ast.MethodCallExpression{
					Object:    expr,
					Method:    fieldName,
					Arguments: args,
				}
			} else {
				// Field access
				expr = &ast.FieldExpression{
					Object: expr,
					Field:  fieldName,
				}
			}
		} else if p.match(lexer.TknLBracket) {
			// Index access
			index, err := p.parseExpression()
			if err != nil {
				return nil, err
			}

			if !p.match(lexer.TknRBracket) {
				return nil, fmt.Errorf("expected ']' at %d:%d", p.current().Line, p.current().Column)
			}

			expr = &ast.IndexExpression{
				Object: expr,
				Index:  index,
			}
		} else if p.match(lexer.TknQuestion) {
			// Error propagation
			expr = &ast.ErrorPropagation{Expression: expr}
		} else {
			break
		}
	}

	return expr, nil
}

// parsePrimary parses primary expressions
func (p *Parser) parsePrimary() (ast.Expression, error) {
	// Literals
	if p.match(lexer.TknInteger) {
		val, err := strconv.ParseInt(p.previous().Text, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid integer: %s", p.previous().Text)
		}
		return &ast.IntegerLiteral{Value: val}, nil
	}

	if p.match(lexer.TknFloat) {
		val, err := strconv.ParseFloat(p.previous().Text, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid float: %s", p.previous().Text)
		}
		return &ast.FloatLiteral{Value: val}, nil
	}

	if p.match(lexer.TknString) {
		return &ast.StringLiteral{Value: p.previous().Text}, nil
	}

	if p.match(lexer.TknTrue) {
		return &ast.BoolLiteral{Value: true}, nil
	}

	if p.match(lexer.TknFalse) {
		return &ast.BoolLiteral{Value: false}, nil
	}

	if p.match(lexer.TknNone) {
		return &ast.NoneLiteral{}, nil
	}

	if p.match(lexer.TknIdentifier) {
		return &ast.Identifier{Name: p.previous().Text}, nil
	}

	if p.match(lexer.TknLBracket) {
		// Array literal
		var elements []ast.Expression
		if !p.check(lexer.TknRBracket) {
			for {
				elem, err := p.parseExpression()
				if err != nil {
					return nil, err
				}
				elements = append(elements, elem)

				if !p.match(lexer.TknComma) {
					break
				}
			}
		}

		if !p.match(lexer.TknRBracket) {
			return nil, fmt.Errorf("expected ']' at %d:%d", p.current().Line, p.current().Column)
		}

		return &ast.ArrayExpression{Elements: elements}, nil
	}

	if p.match(lexer.TknIf) {
		// Parse if expression
		cond, err := p.parseExpression()
		if err != nil {
			return nil, err
		}

		if !p.match(lexer.TknLBrace) {
			return nil, fmt.Errorf("expected '{' at %d:%d", p.current().Line, p.current().Column)
		}

		thenExpr, err := p.parseExpression()
		if err != nil {
			return nil, err
		}

		if !p.match(lexer.TknRBrace) {
			return nil, fmt.Errorf("expected '}' at %d:%d", p.current().Line, p.current().Column)
		}

		var elseExpr ast.Expression
		if p.match(lexer.TknElse) {
			if p.match(lexer.TknLBrace) {
				var err error
				elseExpr, err = p.parseExpression()
				if err != nil {
					return nil, err
				}

				if !p.match(lexer.TknRBrace) {
					return nil, fmt.Errorf("expected '}' at %d:%d", p.current().Line, p.current().Column)
				}
			}
		}

		return &ast.IfExpression{
			Condition: cond,
			ThenExpr:  thenExpr,
			ElseExpr:  elseExpr,
		}, nil
	}

	if p.match(lexer.TknLParen) {
		expr, err := p.parseExpression()
		if err != nil {
			return nil, err
		}

		if !p.match(lexer.TknRParen) {
			return nil, fmt.Errorf("expected ')' at %d:%d", p.current().Line, p.current().Column)
		}

		return expr, nil
	}

	return nil, fmt.Errorf("unexpected token: %v at %d:%d", p.current().Type, p.current().Line, p.current().Column)
}

// parsePattern parses a pattern
func (p *Parser) parsePattern() (ast.Pattern, error) {
	if p.match(lexer.TknIdentifier) {
		return &ast.IdentifierPattern{Name: p.previous().Text}, nil
	}

	if p.match(lexer.TknInteger) {
		val, _ := strconv.ParseInt(p.previous().Text, 10, 64)
		return &ast.LiteralPattern{Value: &ast.IntegerLiteral{Value: val}}, nil
	}

	if p.match(lexer.TknString) {
		return &ast.LiteralPattern{Value: &ast.StringLiteral{Value: p.previous().Text}}, nil
	}

	if p.match(lexer.TknTilde) {
		return &ast.WildcardPattern{}, nil
	}

	return nil, fmt.Errorf("unexpected pattern at %d:%d", p.current().Line, p.current().Column)
}

// parseType parses a type
func (p *Parser) parseType() *ast.Type {
	// Handle option types (?)
	isOption := p.match(lexer.TknQuestion)

	// Get base type
	var typeName string
	if p.match(lexer.TknIdentifier) {
		typeName = p.previous().Text
	} else {
		typeName = "unknown"
	}

	// Handle array types (prefix [])
	isArray := false
	if p.check(lexer.TknLBracket) {
		p.advance()
		if p.match(lexer.TknRBracket) {
			isArray = true
		} else {
			p.pos-- // backtrack
		}
	}

	return &ast.Type{
		Name:       typeName,
		IsOption:   isOption,
		IsArray:    isArray,
		IsPrimitive: isPrimitiveType(typeName),
	}
}

// Helper methods

func (p *Parser) current() lexer.Token {
	if p.pos < len(p.tokens) {
		return p.tokens[p.pos]
	}
	return lexer.Token{Type: lexer.TknEof}
}

func (p *Parser) previous() lexer.Token {
	if p.pos > 0 {
		return p.tokens[p.pos-1]
	}
	return lexer.Token{Type: lexer.TknEof}
}

func (p *Parser) advance() lexer.Token {
	curr := p.current()
	if !p.isAtEnd() {
		p.pos++
	}
	return curr
}

func (p *Parser) match(types ...lexer.TokenType) bool {
	for _, t := range types {
		if p.check(t) {
			p.advance()
			return true
		}
	}
	return false
}

func (p *Parser) check(t lexer.TokenType) bool {
	if p.isAtEnd() {
		return false
	}
	return p.current().Type == t
}

func (p *Parser) isAtEnd() bool {
	return p.current().Type == lexer.TknEof
}

func (p *Parser) isOperator(t lexer.TokenType) bool {
	// Check if token is an operator
	switch t {
	case lexer.TknPlus, lexer.TknMinus, lexer.TknStar, lexer.TknSlash, lexer.TknPercent:
		return true
	case lexer.TknAmpersand, lexer.TknPipe, lexer.TknCaret, lexer.TknTilde:
		return true
	case lexer.TknLeftShift, lexer.TknRightShift:
		return true
	case lexer.TknEq, lexer.TknNe, lexer.TknLt, lexer.TknLe, lexer.TknGt, lexer.TknGe:
		return true
	case lexer.TknLogicalAnd, lexer.TknLogicalOr, lexer.TknNot:
		return true
	case lexer.TknDoubleDot:
		return true
	default:
		return false
	}
}

func getPrecedence(t lexer.TokenType) int {
	switch t {
	case lexer.TknLogicalOr:
		return 1
	case lexer.TknLogicalAnd:
		return 2
	case lexer.TknEq, lexer.TknNe, lexer.TknLt, lexer.TknLe, lexer.TknGt, lexer.TknGe:
		return 3
	case lexer.TknPlus, lexer.TknMinus:
		return 4
	case lexer.TknStar, lexer.TknSlash, lexer.TknPercent:
		return 5
	case lexer.TknCaret:
		return 6
	case lexer.TknDoubleDot:
		return 3 // Same as comparison for range expressions
	default:
		return 0
	}
}

func (p *Parser) isPrimitiveType() bool {
	if p.current().Type == lexer.TknIdentifier {
		return isPrimitiveType(p.current().Text)
	}
	// Also check for V primitive keywords
	switch p.current().Type {
	case lexer.TknTrue, lexer.TknFalse:
		return true
	}
	return false
}

func isPrimitiveType(name string) bool {
	switch name {
	case "i8", "i16", "i32", "i64", "u8", "u16", "u32", "u64", "f32", "f64", "bool", "string", "char", "int", "float":
		return true
	}
	return false
}
