package parser

// Phase 2: Parser - Julia 파서 구현

import (
	"fmt"
	"juliacc/internal/ast"
	"juliacc/internal/lexer"
)

// Parser - Julia 파서
type Parser struct {
	tokens    []lexer.Token
	pos       int             // 현재 토큰 위치
	current   lexer.Token     // 현재 토큰
	errors    []ParserError
	precMap   map[lexer.TokenType]int // 연산자 우선순위
}

// ParserError - 파서 에러
type ParserError struct {
	Message string
	Token   lexer.Token
}

// NewParser - 새로운 파서 생성
func NewParser(tokens []lexer.Token) *Parser {
	p := &Parser{
		tokens:  tokens,
		pos:     0,
		errors:  []ParserError{},
		precMap: buildPrecedenceMap(),
	}
	p.advance()
	return p
}

// advance - 다음 토큰으로 이동
func (p *Parser) advance() {
	if p.pos < len(p.tokens) {
		p.current = p.tokens[p.pos]
		p.pos++
	}
}

// peek - n번째 앞의 토큰 보기
func (p *Parser) peek(n int) lexer.Token {
	pos := p.pos - 1 + n
	if pos >= len(p.tokens) {
		return p.tokens[len(p.tokens)-1] // EOF
	}
	return p.tokens[pos]
}

// match - 현재 토큰이 주어진 타입과 일치하는지 확인 및 소비
func (p *Parser) match(types ...lexer.TokenType) bool {
	for _, t := range types {
		if p.current.Type == t {
			p.advance()
			return true
		}
	}
	return false
}

// expect - 특정 토큰을 기대하고 소비, 실패 시 에러
func (p *Parser) expect(tokenType lexer.TokenType) lexer.Token {
	if p.current.Type != tokenType {
		p.error(fmt.Sprintf("예상 %v, 얻은 %v", tokenType, p.current.Type))
		return p.current
	}
	token := p.current
	p.advance()
	return token
}

// error - 에러 추가
func (p *Parser) error(msg string) {
	p.errors = append(p.errors, ParserError{
		Message: msg,
		Token:   p.current,
	})
}

// Parse - 프로그램 파싱
func (p *Parser) Parse() (*ast.Program, error) {
	stmts := []ast.Stmt{}

	for p.current.Type != lexer.TokenEOF {
		// 개행 건너뛰기
		if p.match(lexer.TokenNewline) {
			continue
		}

		stmt := p.parseStatement()
		if stmt != nil {
			stmts = append(stmts, stmt)
		}

		// 개행 또는 세미콜론 건너뛰기
		for p.match(lexer.TokenNewline, lexer.TokenSemicolon) {
		}
	}

	if len(p.errors) > 0 {
		return nil, fmt.Errorf("파싱 에러: %v", p.errors)
	}

	return &ast.Program{Statements: stmts}, nil
}

// parseStatement - 문 파싱
func (p *Parser) parseStatement() ast.Stmt {
	switch p.current.Type {
	case lexer.TokenKeywordFunction:
		return p.parseFunctionDecl()
	case lexer.TokenKeywordStruct, lexer.TokenKeywordMutable:
		return p.parseStructDecl()
	case lexer.TokenKeywordIf:
		return p.parseIfStmt()
	case lexer.TokenKeywordWhile:
		return p.parseWhileStmt()
	case lexer.TokenKeywordFor:
		return p.parseForStmt()
	case lexer.TokenKeywordReturn:
		return p.parseReturnStmt()
	case lexer.TokenKeywordBreak:
		p.advance()
		return &ast.BreakStmt{Break: p.tokens[p.pos-2]}
	case lexer.TokenKeywordContinue:
		p.advance()
		return &ast.ContinueStmt{Continue: p.tokens[p.pos-2]}
	case lexer.TokenKeywordTry:
		return p.parseTryStmt()
	case lexer.TokenKeywordConst:
		return p.parseConstDecl()
	case lexer.TokenKeywordLet:
		return p.parseLetStmt()
	default:
		return p.parseExprStmt()
	}
}

// parseExprStmt - 표현식 문 파싱
func (p *Parser) parseExprStmt() ast.Stmt {
	expr := p.parseExpression()
	if expr == nil {
		return nil
	}

	// 할당인지 확인
	if p.isAssignmentOp(p.current.Type) {
		op := p.current.Type
		opToken := p.current
		p.advance()
		right := p.parseExpression()
		if right == nil {
			p.error("할당의 우변이 필요합니다")
			return nil
		}
		return &ast.Assignment{
			Left:   expr,
			Right:  right,
			Op:     op,
			Equal:  opToken,
		}
	}

	return &ast.ExprStmt{Expr: expr}
}

// parseExpression - 표현식 파싱 (우선순위 클라이밍)
func (p *Parser) parseExpression() ast.Expr {
	return p.parseBinaryOp(0)
}

// parseBinaryOp - 이항 연산 파싱 (우선순위 클라이밍)
func (p *Parser) parseBinaryOp(minPrec int) ast.Expr {
	left := p.parseUnary()
	if left == nil {
		return nil
	}

	for p.isBinaryOp(p.current.Type) {
		prec := p.getPrecedence(p.current.Type)
		if prec < minPrec {
			break
		}

		op := p.current.Type
		opToken := p.current
		p.advance()

		// 우결합이면 prec, 좌결합이면 prec+1
		nextPrec := prec
		if p.isLeftAssociative(op) {
			nextPrec = prec + 1
		}

		right := p.parseBinaryOp(nextPrec)
		if right == nil {
			p.error("이항 연산자의 우변이 필요합니다")
			return nil
		}

		left = &ast.BinaryOp{
			Left:    left,
			Op:      op,
			Right:   right,
			OpToken: opToken,
		}
	}

	return left
}

// parseUnary - 단항 연산 파싱
func (p *Parser) parseUnary() ast.Expr {
	if p.isUnaryOp(p.current.Type) {
		op := p.current.Type
		opToken := p.current
		p.advance()
		operand := p.parseUnary()
		if operand == nil {
			p.error("단항 연산자의 피연산자가 필요합니다")
			return nil
		}
		return &ast.UnaryOp{
			Op:      op,
			Operand: operand,
			OpToken: opToken,
		}
	}

	return p.parsePostfix()
}

// parsePostfix - 후위 연산 파싱 (함수 호출, 인덱싱, 멤버 접근)
func (p *Parser) parsePostfix() ast.Expr {
	expr := p.parsePrimary()
	if expr == nil {
		return nil
	}

	for {
		switch p.current.Type {
		case lexer.TokenLparen:
			// 함수 호출
			lparen := p.current
			p.advance()
			args := p.parseArguments()
			rparen := p.expect(lexer.TokenRparen)
			expr = &ast.Call{
				Function:  expr,
				Arguments: args,
				LParen:    lparen,
				RParen:    rparen,
			}

		case lexer.TokenLbracket:
			// 인덱싱
			lbracket := p.current
			p.advance()
			indices := []ast.Expr{}
			for p.current.Type != lexer.TokenRbracket {
				idx := p.parseExpression()
				if idx == nil {
					p.error("인덱스 표현식이 필요합니다")
					return nil
				}
				indices = append(indices, idx)
				if !p.match(lexer.TokenComma) {
					break
				}
			}
			rbracket := p.expect(lexer.TokenRbracket)
			expr = &ast.Index{
				Object:   expr,
				Index:    indices,
				LBracket: lbracket,
				RBracket: rbracket,
			}

		case lexer.TokenDot:
			// 멤버 접근
			dot := p.current
			p.advance()
			if p.current.Type != lexer.TokenIdentifier {
				p.error("멤버 이름이 필요합니다")
				return nil
			}
			field := p.current.Lexeme
			p.advance()
			expr = &ast.MemberAccess{
				Object: expr,
				Field:  field,
				Dot:    dot,
			}

		case lexer.TokenDoubleColon:
			// 타입 주석
			colonColon := p.current
			p.advance()
			if p.current.Type != lexer.TokenIdentifier {
				p.error("타입 이름이 필요합니다")
				return nil
			}
			typeStr := p.current.Lexeme
			p.advance()
			expr = &ast.TypeAnnotation{
				Expr:       expr,
				Type:       typeStr,
				ColonColon: colonColon,
			}

		default:
			return expr
		}
	}
}

// parsePrimary - 기본 표현식 파싱
func (p *Parser) parsePrimary() ast.Expr {
	switch p.current.Type {
	case lexer.TokenInteger, lexer.TokenFloat, lexer.TokenString,
		lexer.TokenSymbol, lexer.TokenComplexNumber, lexer.TokenRationalNumber:
		// 리터럴
		token := p.current
		p.advance()
		return &ast.Literal{
			Token: token,
			Value: token.Lexeme,
		}

	case lexer.TokenKeywordTrue:
		token := p.current
		p.advance()
		return &ast.Literal{Token: token, Value: true}

	case lexer.TokenKeywordFalse:
		token := p.current
		p.advance()
		return &ast.Literal{Token: token, Value: false}

	case lexer.TokenKeywordNothing:
		token := p.current
		p.advance()
		return &ast.Literal{Token: token, Value: nil}

	case lexer.TokenIdentifier:
		// 식별자
		token := p.current
		p.advance()
		return &ast.Identifier{
			Token: token,
			Name:  token.Lexeme,
		}

	case lexer.TokenLparen:
		// 괄호 표현식 또는 튜플
		lparen := p.current
		p.advance()
		if p.match(lexer.TokenRparen) {
			// 빈 튜플
			return &ast.TupleLiteral{
				Elements: []ast.Expr{},
				LParen:   lparen,
				RParen:   p.tokens[p.pos-1],
			}
		}
		expr := p.parseExpression()
		if expr == nil {
			p.error("괄호 안에 표현식이 필요합니다")
			return nil
		}

		// 튜플인지 확인 (쉼표가 있으면 튜플)
		if p.match(lexer.TokenComma) {
			elements := []ast.Expr{expr}
			for p.current.Type != lexer.TokenRparen {
				e := p.parseExpression()
				if e == nil {
					p.error("튜플 요소가 필요합니다")
					return nil
				}
				elements = append(elements, e)
				if !p.match(lexer.TokenComma) {
					break
				}
			}
			rparen := p.expect(lexer.TokenRparen)
			return &ast.TupleLiteral{
				Elements: elements,
				LParen:   lparen,
				RParen:   rparen,
			}
		}

		rparen := p.expect(lexer.TokenRparen)
		// 단순 괄호 표현식은 그냥 내부 표현식 반환
		_ = rparen // 위치 정보 사용 가능
		return expr

	case lexer.TokenLbracket:
		// 배열 리터럴
		return p.parseArrayLiteral()

	default:
		p.error(fmt.Sprintf("예상치 못한 토큰: %v", p.current.Type))
		return nil
	}
}

// parseArrayLiteral - 배열 리터럴 파싱 ([1, 2, 3])
func (p *Parser) parseArrayLiteral() ast.Expr {
	lbracket := p.current
	p.advance()

	elements := [][]ast.Expr{}
	row := []ast.Expr{}

	for p.current.Type != lexer.TokenRbracket {
		expr := p.parseExpression()
		if expr == nil {
			p.error("배열 요소가 필요합니다")
			return nil
		}
		row = append(row, expr)

		if p.match(lexer.TokenComma) {
			continue
		} else if p.match(lexer.TokenSemicolon) {
			elements = append(elements, row)
			row = []ast.Expr{}
		} else {
			break
		}
	}

	if len(row) > 0 {
		elements = append(elements, row)
	}

	rbracket := p.expect(lexer.TokenRbracket)
	return &ast.ArrayLiteral{
		Elements: elements,
		LBracket: lbracket,
		RBracket: rbracket,
	}
}

// parseArguments - 함수 인자 파싱
func (p *Parser) parseArguments() []ast.Expr {
	args := []ast.Expr{}

	if p.current.Type == lexer.TokenRparen {
		return args
	}

	for {
		arg := p.parseExpression()
		if arg == nil {
			break
		}
		args = append(args, arg)

		if !p.match(lexer.TokenComma) {
			break
		}
	}

	return args
}

// parseFunctionDecl - 함수 선언 파싱
func (p *Parser) parseFunctionDecl() ast.Stmt {
	funcToken := p.current
	p.advance()

	if p.current.Type != lexer.TokenIdentifier {
		p.error("함수 이름이 필요합니다")
		return nil
	}
	name := p.current.Lexeme
	p.advance()

	// 매개변수 파싱
	p.expect(lexer.TokenLparen)
	params := []*ast.Parameter{}
	for p.current.Type != lexer.TokenRparen {
		if p.current.Type != lexer.TokenIdentifier {
			p.error("매개변수 이름이 필요합니다")
			return nil
		}
		paramName := p.current.Lexeme
		p.advance()

		param := &ast.Parameter{Name: paramName}
		// TODO: 타입, 기본값, 가변 인자 처리
		params = append(params, param)

		if !p.match(lexer.TokenComma) {
			break
		}
	}
	p.expect(lexer.TokenRparen)

	// 반환 타입 (옵션)
	var returnType string
	if p.match(lexer.TokenDoubleColon) {
		if p.current.Type != lexer.TokenIdentifier {
			p.error("반환 타입이 필요합니다")
			return nil
		}
		returnType = p.current.Lexeme
		p.advance()
	}

	// 함수 본체 파싱
	body := p.parseBlock(lexer.TokenKeywordEnd)
	p.expect(lexer.TokenKeywordEnd)

	return &ast.FunctionDecl{
		Name:       name,
		Parameters: params,
		ReturnType: returnType,
		Body:       body,
		Function:   funcToken,
	}
}

// parseStructDecl - 구조체 선언 파싱
func (p *Parser) parseStructDecl() ast.Stmt {
	isMutable := false
	var structToken lexer.Token

	if p.match(lexer.TokenKeywordMutable) {
		isMutable = true
	}

	structToken = p.current
	p.expect(lexer.TokenKeywordStruct)

	if p.current.Type != lexer.TokenIdentifier {
		p.error("구조체 이름이 필요합니다")
		return nil
	}
	name := p.current.Lexeme
	p.advance()

	// 필드 파싱
	fields := []*ast.StructField{}
	for p.current.Type != lexer.TokenKeywordEnd {
		if p.match(lexer.TokenNewline) {
			continue
		}

		if p.current.Type != lexer.TokenIdentifier {
			p.error("필드 이름이 필요합니다")
			return nil
		}
		fieldName := p.current.Lexeme
		p.advance()

		var fieldType string
		if p.match(lexer.TokenDoubleColon) {
			if p.current.Type != lexer.TokenIdentifier {
				p.error("필드 타입이 필요합니다")
				return nil
			}
			fieldType = p.current.Lexeme
			p.advance()
		}

		fields = append(fields, &ast.StructField{
			Name: fieldName,
			Type: fieldType,
		})
	}

	p.expect(lexer.TokenKeywordEnd)

	return &ast.StructDecl{
		Name:      name,
		IsMutable: isMutable,
		Fields:    fields,
		Struct:    structToken,
	}
}

// parseIfStmt - if 문 파싱
func (p *Parser) parseIfStmt() ast.Stmt {
	ifToken := p.current
	p.advance()

	condition := p.parseExpression()
	if condition == nil {
		p.error("if 조건이 필요합니다")
		return nil
	}

	thenBody := p.parseBlock(lexer.TokenKeywordElseif, lexer.TokenKeywordElse, lexer.TokenKeywordEnd)

	elseIfs := []*ast.ElseIfClause{}
	for p.match(lexer.TokenKeywordElseif) {
		elifToken := p.tokens[p.pos-1]
		elifCond := p.parseExpression()
		if elifCond == nil {
			p.error("elseif 조건이 필요합니다")
			return nil
		}
		elifBody := p.parseBlock(lexer.TokenKeywordElseif, lexer.TokenKeywordElse, lexer.TokenKeywordEnd)
		elseIfs = append(elseIfs, &ast.ElseIfClause{
			Condition: elifCond,
			Body:      elifBody,
			ElseIf:    elifToken,
		})
	}

	var elseBody []ast.Stmt
	if p.match(lexer.TokenKeywordElse) {
		elseBody = p.parseBlock(lexer.TokenKeywordEnd)
	}

	p.expect(lexer.TokenKeywordEnd)

	return &ast.IfStmt{
		Condition: condition,
		Then:      thenBody,
		ElseIfs:   elseIfs,
		Else:      elseBody,
		If:        ifToken,
	}
}

// parseWhileStmt - while 루프 파싱
func (p *Parser) parseWhileStmt() ast.Stmt {
	whileToken := p.current
	p.advance()

	condition := p.parseExpression()
	if condition == nil {
		p.error("while 조건이 필요합니다")
		return nil
	}

	body := p.parseBlock(lexer.TokenKeywordEnd)
	p.expect(lexer.TokenKeywordEnd)

	return &ast.WhileStmt{
		Condition: condition,
		Body:      body,
		While:     whileToken,
	}
}

// parseForStmt - for 루프 파싱
func (p *Parser) parseForStmt() ast.Stmt {
	forToken := p.current
	p.advance()

	if p.current.Type != lexer.TokenIdentifier {
		p.error("for 루프 변수가 필요합니다")
		return nil
	}
	loopVar := p.current.Lexeme
	p.advance()

	if !p.match(lexer.TokenKeywordIn) {
		p.error("for 루프에서 'in'이 필요합니다")
		return nil
	}

	iterator := p.parseExpression()
	if iterator == nil {
		p.error("반복 대상이 필요합니다")
		return nil
	}

	body := p.parseBlock(lexer.TokenKeywordEnd)
	p.expect(lexer.TokenKeywordEnd)

	return &ast.ForStmt{
		Variable: loopVar,
		Iterator: iterator,
		Body:     body,
		For:      forToken,
	}
}

// parseReturnStmt - return 문 파싱
func (p *Parser) parseReturnStmt() ast.Stmt {
	retToken := p.current
	p.advance()

	var value ast.Expr
	if p.current.Type != lexer.TokenNewline && p.current.Type != lexer.TokenEOF {
		value = p.parseExpression()
	}

	return &ast.ReturnStmt{
		Value:  value,
		Return: retToken,
	}
}

// parseBlock - 코드 블록 파싱
func (p *Parser) parseBlock(endTokens ...lexer.TokenType) []ast.Stmt {
	stmts := []ast.Stmt{}

	for {
		if p.isIn(p.current.Type, endTokens...) || p.current.Type == lexer.TokenEOF {
			break
		}

		if p.match(lexer.TokenNewline) {
			continue
		}

		stmt := p.parseStatement()
		if stmt != nil {
			stmts = append(stmts, stmt)
		}

		for p.match(lexer.TokenNewline, lexer.TokenSemicolon) {
		}
	}

	return stmts
}

// parseTryStmt - try-catch 문 파싱
func (p *Parser) parseTryStmt() ast.Stmt {
	tryToken := p.current
	p.advance()

	tryBody := p.parseBlock(lexer.TokenKeywordCatch, lexer.TokenKeywordFinally, lexer.TokenKeywordEnd)

	catches := []*ast.CatchClause{}
	for p.match(lexer.TokenKeywordCatch) {
		catchToken := p.tokens[p.pos-1]
		catchBody := p.parseBlock(lexer.TokenKeywordCatch, lexer.TokenKeywordFinally, lexer.TokenKeywordEnd)
		catches = append(catches, &ast.CatchClause{
			Body:  catchBody,
			Catch: catchToken,
		})
	}

	var finallyBody []ast.Stmt
	if p.match(lexer.TokenKeywordFinally) {
		finallyBody = p.parseBlock(lexer.TokenKeywordEnd)
	}

	p.expect(lexer.TokenKeywordEnd)

	return &ast.TryStmt{
		Try:     tryBody,
		Catches: catches,
		Finally: finallyBody,
		Try_:    tryToken,
	}
}

// parseConstDecl - const 선언 파싱
func (p *Parser) parseConstDecl() ast.Stmt {
	constToken := p.current
	p.advance()

	if p.current.Type != lexer.TokenIdentifier {
		p.error("상수 이름이 필요합니다")
		return nil
	}
	name := p.current.Lexeme
	p.advance()

	p.expect(lexer.TokenEqual)
	value := p.parseExpression()
	if value == nil {
		p.error("상수 값이 필요합니다")
		return nil
	}

	return &ast.ConstDecl{
		Name:  name,
		Value: value,
		Const: constToken,
	}
}

// parseLetStmt - let 선언 파싱
func (p *Parser) parseLetStmt() ast.Stmt {
	letToken := p.current
	p.advance()

	if p.current.Type != lexer.TokenIdentifier {
		p.error("변수 이름이 필요합니다")
		return nil
	}
	name := p.current.Lexeme
	p.advance()

	p.expect(lexer.TokenEqual)
	value := p.parseExpression()
	if value == nil {
		p.error("변수 값이 필요합니다")
		return nil
	}

	return &ast.VarDecl{
		Name:  name,
		Value: value,
		Let:   &letToken,
	}
}

// ============================================
// 헬퍼 함수
// ============================================

// buildPrecedenceMap - 연산자 우선순위 맵 구축
func buildPrecedenceMap() map[lexer.TokenType]int {
	return map[lexer.TokenType]int{
		lexer.TokenOr:            1,
		lexer.TokenAnd:           2,
		lexer.TokenEqualEqual:    3,
		lexer.TokenNotEqual:      3,
		lexer.TokenLess:          3,
		lexer.TokenLessEqual:     3,
		lexer.TokenGreater:       3,
		lexer.TokenGreaterEqual:  3,
		lexer.TokenPipe:          4,
		lexer.TokenAmpersand:     5,
		lexer.TokenLeftShift:     6,
		lexer.TokenRightShift:    6,
		lexer.TokenUnsignedRightShift: 6,
		lexer.TokenPlus:          7,
		lexer.TokenMinus:         7,
		lexer.TokenStar:          8,
		lexer.TokenSlash:         8,
		lexer.TokenPercent:       8,
		lexer.TokenCaret:         9,
		lexer.TokenDoubleDot:     10,
	}
}

// getPrecedence - 연산자의 우선순위 얻기
func (p *Parser) getPrecedence(tokenType lexer.TokenType) int {
	if prec, ok := p.precMap[tokenType]; ok {
		return prec
	}
	return 0
}

// isBinaryOp - 이항 연산자인지 확인
func (p *Parser) isBinaryOp(tokenType lexer.TokenType) bool {
	return p.getPrecedence(tokenType) > 0
}

// isUnaryOp - 단항 연산자인지 확인
func (p *Parser) isUnaryOp(tokenType lexer.TokenType) bool {
	switch tokenType {
	case lexer.TokenMinus, lexer.TokenNot, lexer.TokenTilde, lexer.TokenPlus:
		return true
	}
	return false
}

// isLeftAssociative - 좌결합 연산자인지 확인
func (p *Parser) isLeftAssociative(tokenType lexer.TokenType) bool {
	// 대부분의 연산자가 좌결합
	// 우결합: power (^) 등
	switch tokenType {
	case lexer.TokenCaret:
		return false // 우결합
	}
	return true
}

// isAssignmentOp - 할당 연산자인지 확인
func (p *Parser) isAssignmentOp(tokenType lexer.TokenType) bool {
	switch tokenType {
	case lexer.TokenEqual, lexer.TokenPlusAssign, lexer.TokenMinusAssign,
		lexer.TokenStarAssign, lexer.TokenSlashAssign, lexer.TokenPercentAssign,
		lexer.TokenCaretAssign, lexer.TokenAndAssign, lexer.TokenPipeAssign:
		return true
	}
	return false
}

// isIn - 주어진 토큰 타입이 목록에 있는지 확인
func (p *Parser) isIn(tokenType lexer.TokenType, types ...lexer.TokenType) bool {
	for _, t := range types {
		if tokenType == t {
			return true
		}
	}
	return false
}
