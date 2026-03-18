package ast

// Phase 2: Parser - AST (추상 구문 트리) 노드 정의

import (
	"fmt"
	"juliacc/internal/lexer"
)

// Node - 모든 AST 노드의 기본 인터페이스
type Node interface {
	node()
}

// Expr - 표현식 노드
type Expr interface {
	Node
	expr()
	String() string
}

// Stmt - 문(명령문) 노드
type Stmt interface {
	Node
	stmt()
	String() string
}

// ============================================
// 표현식 노드 (Expr)
// ============================================

// Literal - 리터럴 (정수, 부동소수, 문자열, 심볼, true/false/nothing)
type Literal struct {
	Token lexer.Token
	Value interface{}
}

func (l *Literal) node() {}
func (l *Literal) expr() {}
func (l *Literal) String() string {
	return fmt.Sprintf("Literal(%v)", l.Value)
}

// Identifier - 식별자
type Identifier struct {
	Token lexer.Token
	Name  string
}

func (i *Identifier) node() {}
func (i *Identifier) expr() {}
func (i *Identifier) String() string {
	return i.Name
}

// BinaryOp - 이항 연산 (a + b, a * b, etc)
type BinaryOp struct {
	Left     Expr
	Op       lexer.TokenType
	Right    Expr
	OpToken  lexer.Token
}

func (b *BinaryOp) node() {}
func (b *BinaryOp) expr() {}
func (b *BinaryOp) String() string {
	return fmt.Sprintf("(%s %v %s)", b.Left, b.OpToken.Lexeme, b.Right)
}

// UnaryOp - 단항 연산 (!a, -a, ~a, etc)
type UnaryOp struct {
	Op      lexer.TokenType
	Operand Expr
	OpToken lexer.Token
}

func (u *UnaryOp) node() {}
func (u *UnaryOp) expr() {}
func (u *UnaryOp) String() string {
	return fmt.Sprintf("%s%s", u.OpToken.Lexeme, u.Operand)
}

// Call - 함수 호출 (f(a, b, c))
type Call struct {
	Function  Expr
	Arguments []Expr
	LParen    lexer.Token
	RParen    lexer.Token
}

func (c *Call) node() {}
func (c *Call) expr() {}
func (c *Call) String() string {
	return fmt.Sprintf("%s(...)", c.Function)
}

// Index - 인덱싱 (a[i], matrix[i,j])
type Index struct {
	Object Expr
	Index  []Expr
	LBracket lexer.Token
	RBracket lexer.Token
}

func (i *Index) node() {}
func (i *Index) expr() {}
func (i *Index) String() string {
	return fmt.Sprintf("%s[...]", i.Object)
}

// MemberAccess - 멤버 접근 (obj.field)
type MemberAccess struct {
	Object Expr
	Field  string
	Dot    lexer.Token
}

func (m *MemberAccess) node() {}
func (m *MemberAccess) expr() {}
func (m *MemberAccess) String() string {
	return fmt.Sprintf("%s.%s", m.Object, m.Field)
}

// TypeAnnotation - 타입 주석 (x::Int, a::Vector{T})
type TypeAnnotation struct {
	Expr     Expr
	Type     string
	ColonColon lexer.Token
}

func (t *TypeAnnotation) node() {}
func (t *TypeAnnotation) expr() {}
func (t *TypeAnnotation) String() string {
	return fmt.Sprintf("%s::%s", t.Expr, t.Type)
}

// ArrayLiteral - 배열 리터럴 ([1, 2, 3], [a b; c d])
type ArrayLiteral struct {
	Elements [][]Expr // 2D for matrix support
	LBracket lexer.Token
	RBracket lexer.Token
}

func (a *ArrayLiteral) node() {}
func (a *ArrayLiteral) expr() {}
func (a *ArrayLiteral) String() string {
	return "[...]"
}

// DictLiteral - 딕셔너리 리터럴 (Dict(a=>1, b=>2))
type DictLiteral struct {
	Pairs [][2]Expr // key-value pairs
}

func (d *DictLiteral) node() {}
func (d *DictLiteral) expr() {}
func (d *DictLiteral) String() string {
	return "Dict(...)"
}

// TupleLiteral - 튜플 리터럴 ((1, 2, 3))
type TupleLiteral struct {
	Elements []Expr
	LParen   lexer.Token
	RParen   lexer.Token
}

func (t *TupleLiteral) node() {}
func (t *TupleLiteral) expr() {}
func (t *TupleLiteral) String() string {
	return "(...)"
}

// ============================================
// 문(Statement) 노드 (Stmt)
// ============================================

// ExprStmt - 표현식 문 (3 + 4; x;)
type ExprStmt struct {
	Expr Expr
}

func (e *ExprStmt) node() {}
func (e *ExprStmt) stmt() {}
func (e *ExprStmt) String() string {
	return e.Expr.String()
}

// Assignment - 할당 (x = 5, a[i] = v)
type Assignment struct {
	Left     Expr
	Right    Expr
	Op       lexer.TokenType
	Equal    lexer.Token
}

func (a *Assignment) node() {}
func (a *Assignment) stmt() {}
func (a *Assignment) String() string {
	return fmt.Sprintf("%s %s %s", a.Left, a.Equal.Lexeme, a.Right)
}

// VarDecl - 변수 선언 (x = 5; let x = 5)
type VarDecl struct {
	Name   string
	Type   string // optional
	Value  Expr
	Let    *lexer.Token // nil for implicit, or let token
}

func (v *VarDecl) node() {}
func (v *VarDecl) stmt() {}
func (v *VarDecl) String() string {
	return fmt.Sprintf("let %s = %s", v.Name, v.Value)
}

// ConstDecl - 상수 선언 (const x = 5)
type ConstDecl struct {
	Name  string
	Type  string
	Value Expr
	Const lexer.Token
}

func (c *ConstDecl) node() {}
func (c *ConstDecl) stmt() {}
func (c *ConstDecl) String() string {
	return fmt.Sprintf("const %s = %s", c.Name, c.Value)
}

// FunctionDecl - 함수 선언
type FunctionDecl struct {
	Name       string
	Parameters []*Parameter
	ReturnType string // optional
	Body       []Stmt
	Function   lexer.Token
}

func (f *FunctionDecl) node() {}
func (f *FunctionDecl) stmt() {}
func (f *FunctionDecl) String() string {
	return fmt.Sprintf("function %s(...)", f.Name)
}

// Parameter - 함수 매개변수
type Parameter struct {
	Name     string
	Type     string    // optional
	Default  Expr      // optional
	Variadic bool      // ... 마킹
}

// StructDecl - 구조체 선언
type StructDecl struct {
	Name      string
	IsMutable bool
	Fields    []*StructField
	Struct    lexer.Token
}

func (s *StructDecl) node() {}
func (s *StructDecl) stmt() {}
func (s *StructDecl) String() string {
	return fmt.Sprintf("struct %s", s.Name)
}

// StructField - 구조체 필드
type StructField struct {
	Name    string
	Type    string
	Default Expr // optional
}

// IfStmt - if 문
type IfStmt struct {
	Condition Expr
	Then      []Stmt
	ElseIfs   []*ElseIfClause
	Else      []Stmt // nil if no else
	If        lexer.Token
}

func (i *IfStmt) node() {}
func (i *IfStmt) stmt() {}
func (i *IfStmt) String() string {
	return "if ..."
}

// ElseIfClause - elseif 절
type ElseIfClause struct {
	Condition Expr
	Body      []Stmt
	ElseIf    lexer.Token
}

// WhileStmt - while 루프
type WhileStmt struct {
	Condition Expr
	Body      []Stmt
	While     lexer.Token
}

func (w *WhileStmt) node() {}
func (w *WhileStmt) stmt() {}
func (w *WhileStmt) String() string {
	return "while ..."
}

// ForStmt - for 루프
type ForStmt struct {
	Variable string  // loop variable
	Iterator Expr    // iterable expression
	Body     []Stmt
	For      lexer.Token
}

func (f *ForStmt) node() {}
func (f *ForStmt) stmt() {}
func (f *ForStmt) String() string {
	return fmt.Sprintf("for %s in ...", f.Variable)
}

// ReturnStmt - return 문
type ReturnStmt struct {
	Value  Expr // nil for bare return
	Return lexer.Token
}

func (r *ReturnStmt) node() {}
func (r *ReturnStmt) stmt() {}
func (r *ReturnStmt) String() string {
	return "return ..."
}

// BreakStmt - break 문
type BreakStmt struct {
	Break lexer.Token
}

func (b *BreakStmt) node() {}
func (b *BreakStmt) stmt() {}
func (b *BreakStmt) String() string {
	return "break"
}

// ContinueStmt - continue 문
type ContinueStmt struct {
	Continue lexer.Token
}

func (c *ContinueStmt) node() {}
func (c *ContinueStmt) stmt() {}
func (c *ContinueStmt) String() string {
	return "continue"
}

// TryStmt - try-catch-finally 문
type TryStmt struct {
	Try     []Stmt
	Catches []*CatchClause
	Finally []Stmt // optional
	Try_    lexer.Token
}

func (t *TryStmt) node() {}
func (t *TryStmt) stmt() {}
func (t *TryStmt) String() string {
	return "try ..."
}

// CatchClause - catch 절
type CatchClause struct {
	ExceptionType string // optional (e.g., "DomainError")
	Variable      string // variable name (optional, e.g., e)
	Body          []Stmt
	Catch         lexer.Token
}

// Program - 프로그램 (최상위 노드)
type Program struct {
	Statements []Stmt
}

func (p *Program) node() {}
func (p *Program) String() string {
	return fmt.Sprintf("Program(%d statements)", len(p.Statements))
}
