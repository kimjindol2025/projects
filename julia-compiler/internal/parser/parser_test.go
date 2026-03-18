package parser

import (
	"juliacc/internal/ast"
	"juliacc/internal/lexer"
	"testing"
)

// Helper - 토큰 시퀀스 생성
func tokenize(input string) []lexer.Token {
	l := lexer.NewLexer(input)
	return l.ScanAll()
}

// TestPhase2BasicLiterals - 기본 리터럴 파싱
func TestPhase2BasicLiterals(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"integer", "42", "Literal(42)"},
		{"float", "3.14", "Literal(3.14)"},
		{"string", `"hello"`, `Literal("hello")`},
		{"symbol", ":test", "Literal(:test)"},
		{"true", "true", "Literal(true)"},
		{"false", "false", "Literal(false)"},
		{"nothing", "nothing", "Literal(<nil>)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens := tokenize(tt.input)
			p := NewParser(tokens)
			prog, err := p.Parse()

			if err != nil {
				t.Fatalf("파싱 에러: %v", err)
			}

			if len(prog.Statements) == 0 {
				t.Fatal("문이 없음")
			}

			exprStmt, ok := prog.Statements[0].(*ast.ExprStmt)
			if !ok {
				t.Fatalf("ExprStmt가 아님: %T", prog.Statements[0])
			}

			if exprStmt.Expr.String() != tt.expected {
				t.Errorf("예상: %s, 얻음: %s", tt.expected, exprStmt.Expr.String())
			}
		})
	}
}

// TestPhase2Identifiers - 식별자 파싱
func TestPhase2Identifiers(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple", "x", "x"},
		{"with_underscore", "var_name", "var_name"},
		{"bang", "push!", "push!"},
		{"question", "isempty?", "isempty?"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens := tokenize(tt.input)
			p := NewParser(tokens)
			prog, err := p.Parse()

			if err != nil {
				t.Fatalf("파싱 에러: %v", err)
			}

			exprStmt := prog.Statements[0].(*ast.ExprStmt)
			ident := exprStmt.Expr.(*ast.Identifier)

			if ident.Name != tt.expected {
				t.Errorf("예상: %s, 얻음: %s", tt.expected, ident.Name)
			}
		})
	}
}

// TestPhase2BinaryOperators - 이항 연산자 파싱
func TestPhase2BinaryOperators(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"addition", "1 + 2"},
		{"subtraction", "5 - 3"},
		{"multiplication", "4 * 6"},
		{"division", "10 / 2"},
		{"power", "2 ^ 3"},
		{"comparison_eq", "a == b"},
		{"comparison_lt", "x < 10"},
		{"logical_and", "true && false"},
		{"logical_or", "true || false"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens := tokenize(tt.input)
			p := NewParser(tokens)
			prog, err := p.Parse()

			if err != nil {
				t.Fatalf("파싱 에러: %v", err)
			}

			exprStmt := prog.Statements[0].(*ast.ExprStmt)
			binOp, ok := exprStmt.Expr.(*ast.BinaryOp)

			if !ok {
				t.Fatalf("BinaryOp가 아님: %T", exprStmt.Expr)
			}

			if binOp.Left == nil || binOp.Right == nil {
				t.Error("BinaryOp의 왼쪽 또는 오른쪽이 nil")
			}
		})
	}
}

// TestPhase2PrecedenceLeftAssoc - 왼쪽 결합 우선순위
func TestPhase2PrecedenceLeftAssoc(t *testing.T) {
	input := "2 + 3 - 1"
	tokens := tokenize(input)
	p := NewParser(tokens)
	prog, err := p.Parse()

	if err != nil {
		t.Fatalf("파싱 에러: %v", err)
	}

	exprStmt := prog.Statements[0].(*ast.ExprStmt)
	binOp := exprStmt.Expr.(*ast.BinaryOp)

	// (2 + 3) - 1 형태여야 함
	// 즉, 루트는 -, 왼쪽은 (2+3)
	if binOp.Op != lexer.TokenMinus {
		t.Errorf("루트 연산자가 -가 아님: %v", binOp.Op)
	}

	leftBinOp, ok := binOp.Left.(*ast.BinaryOp)
	if !ok {
		t.Fatalf("왼쪽 자식이 BinaryOp가 아님: %T", binOp.Left)
	}

	if leftBinOp.Op != lexer.TokenPlus {
		t.Errorf("왼쪽 자식 연산자가 +가 아님: %v", leftBinOp.Op)
	}
}

// TestPhase2PrecedenceRightAssoc - 오른쪽 결합 우선순위 (거듭제곱)
func TestPhase2PrecedenceRightAssoc(t *testing.T) {
	input := "2 ^ 3 ^ 2"
	tokens := tokenize(input)
	p := NewParser(tokens)
	prog, err := p.Parse()

	if err != nil {
		t.Fatalf("파싱 에러: %v", err)
	}

	exprStmt := prog.Statements[0].(*ast.ExprStmt)
	binOp := exprStmt.Expr.(*ast.BinaryOp)

	// 2 ^ (3 ^ 2) 형태여야 함
	if binOp.Op != lexer.TokenCaret {
		t.Errorf("루트 연산자가 ^가 아님: %v", binOp.Op)
	}

	rightBinOp, ok := binOp.Right.(*ast.BinaryOp)
	if !ok {
		t.Fatalf("오른쪽 자식이 BinaryOp가 아님: %T", binOp.Right)
	}

	if rightBinOp.Op != lexer.TokenCaret {
		t.Errorf("오른쪽 자식 연산자가 ^가 아님: %v", rightBinOp.Op)
	}
}

// TestPhase2UnaryOperators - 단항 연산자
func TestPhase2UnaryOperators(t *testing.T) {
	tests := []struct {
		name  string
		input string
		op    lexer.TokenType
	}{
		{"negation", "-x", lexer.TokenMinus},
		{"not", "!true", lexer.TokenNot},
		{"bitwise_not", "~a", lexer.TokenTilde},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens := tokenize(tt.input)
			p := NewParser(tokens)
			prog, err := p.Parse()

			if err != nil {
				t.Fatalf("파싱 에러: %v", err)
			}

			exprStmt := prog.Statements[0].(*ast.ExprStmt)
			unOp, ok := exprStmt.Expr.(*ast.UnaryOp)

			if !ok {
				t.Fatalf("UnaryOp가 아님: %T", exprStmt.Expr)
			}

			if unOp.Op != tt.op {
				t.Errorf("예상: %v, 얻음: %v", tt.op, unOp.Op)
			}
		})
	}
}

// TestPhase2FunctionCall - 함수 호출 파싱
func TestPhase2FunctionCall(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		funcName string
		argCount int
	}{
		{"no_args", "f()", "f", 0},
		{"one_arg", "sin(x)", "sin", 1},
		{"two_args", "add(a, b)", "add", 2},
		{"three_args", "max(x, y, z)", "max", 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens := tokenize(tt.input)
			p := NewParser(tokens)
			prog, err := p.Parse()

			if err != nil {
				t.Fatalf("파싱 에러: %v", err)
			}

			exprStmt := prog.Statements[0].(*ast.ExprStmt)
			call, ok := exprStmt.Expr.(*ast.Call)

			if !ok {
				t.Fatalf("Call이 아님: %T", exprStmt.Expr)
			}

			funcIdent := call.Function.(*ast.Identifier)
			if funcIdent.Name != tt.funcName {
				t.Errorf("함수명 예상: %s, 얻음: %s", tt.funcName, funcIdent.Name)
			}

			if len(call.Arguments) != tt.argCount {
				t.Errorf("인자 개수 예상: %d, 얻음: %d", tt.argCount, len(call.Arguments))
			}
		})
	}
}

// TestPhase2ArrayIndexing - 배열 인덱싱
func TestPhase2ArrayIndexing(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		indexCount int
	}{
		{"single_index", "a[1]", 1},
		{"two_indices", "matrix[1, 2]", 2},
		{"three_indices", "tensor[1, 2, 3]", 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens := tokenize(tt.input)
			p := NewParser(tokens)
			prog, err := p.Parse()

			if err != nil {
				t.Fatalf("파싱 에러: %v", err)
			}

			exprStmt := prog.Statements[0].(*ast.ExprStmt)
			idx, ok := exprStmt.Expr.(*ast.Index)

			if !ok {
				t.Fatalf("Index가 아님: %T", exprStmt.Expr)
			}

			if len(idx.Index) != tt.indexCount {
				t.Errorf("인덱스 개수 예상: %d, 얻음: %d", tt.indexCount, len(idx.Index))
			}
		})
	}
}

// TestPhase2MemberAccess - 멤버 접근
func TestPhase2MemberAccess(t *testing.T) {
	input := "point.x"
	tokens := tokenize(input)
	p := NewParser(tokens)
	prog, err := p.Parse()

	if err != nil {
		t.Fatalf("파싱 에러: %v", err)
	}

	exprStmt := prog.Statements[0].(*ast.ExprStmt)
	member, ok := exprStmt.Expr.(*ast.MemberAccess)

	if !ok {
		t.Fatalf("MemberAccess가 아님: %T", exprStmt.Expr)
	}

	if member.Field != "x" {
		t.Errorf("필드명 예상: x, 얻음: %s", member.Field)
	}
}

// TestPhase2TypeAnnotation - 타입 주석
func TestPhase2TypeAnnotation(t *testing.T) {
	input := "x::Int64"
	tokens := tokenize(input)
	p := NewParser(tokens)
	prog, err := p.Parse()

	if err != nil {
		t.Fatalf("파싱 에러: %v", err)
	}

	exprStmt := prog.Statements[0].(*ast.ExprStmt)
	typeAnnot, ok := exprStmt.Expr.(*ast.TypeAnnotation)

	if !ok {
		t.Fatalf("TypeAnnotation이 아님: %T", exprStmt.Expr)
	}

	if typeAnnot.Type != "Int64" {
		t.Errorf("타입명 예상: Int64, 얻음: %s", typeAnnot.Type)
	}
}

// TestPhase2Assignment - 할당
func TestPhase2Assignment(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"simple", "x = 5"},
		{"add_assign", "x += 1"},
		{"mult_assign", "y *= 2"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens := tokenize(tt.input)
			p := NewParser(tokens)
			prog, err := p.Parse()

			if err != nil {
				t.Fatalf("파싱 에러: %v", err)
			}

			_, ok := prog.Statements[0].(*ast.Assignment)
			if !ok {
				t.Fatalf("Assignment가 아님: %T", prog.Statements[0])
			}
		})
	}
}

// TestPhase2IfStatement - If 문
func TestPhase2IfStatement(t *testing.T) {
	input := `if x > 0
		y = 1
	elseif x < 0
		y = -1
	else
		y = 0
	end`

	tokens := tokenize(input)
	p := NewParser(tokens)
	prog, err := p.Parse()

	if err != nil {
		t.Fatalf("파싱 에러: %v", err)
	}

	ifStmt, ok := prog.Statements[0].(*ast.IfStmt)
	if !ok {
		t.Fatalf("IfStmt가 아님: %T", prog.Statements[0])
	}

	if ifStmt.Condition == nil {
		t.Error("조건이 nil")
	}

	if len(ifStmt.ElseIfs) != 1 {
		t.Errorf("elseif 개수 예상: 1, 얻음: %d", len(ifStmt.ElseIfs))
	}

	if ifStmt.Else == nil {
		t.Error("else 블록이 nil")
	}
}

// TestPhase2WhileStatement - While 문
func TestPhase2WhileStatement(t *testing.T) {
	input := `while x < 10
		x = x + 1
	end`

	tokens := tokenize(input)
	p := NewParser(tokens)
	prog, err := p.Parse()

	if err != nil {
		t.Fatalf("파싱 에러: %v", err)
	}

	whileStmt, ok := prog.Statements[0].(*ast.WhileStmt)
	if !ok {
		t.Fatalf("WhileStmt가 아님: %T", prog.Statements[0])
	}

	if whileStmt.Condition == nil {
		t.Error("조건이 nil")
	}

	if len(whileStmt.Body) == 0 {
		t.Error("while 블록이 비어있음")
	}
}

// TestPhase2ForStatement - For 문
func TestPhase2ForStatement(t *testing.T) {
	input := `for i in 1:10
		println(i)
	end`

	tokens := tokenize(input)
	p := NewParser(tokens)
	prog, err := p.Parse()

	if err != nil {
		t.Fatalf("파싱 에러: %v", err)
	}

	forStmt, ok := prog.Statements[0].(*ast.ForStmt)
	if !ok {
		t.Fatalf("ForStmt가 아님: %T", prog.Statements[0])
	}

	if forStmt.Variable != "i" {
		t.Errorf("변수명 예상: i, 얻음: %s", forStmt.Variable)
	}

	if forStmt.Iterator == nil {
		t.Error("iterator가 nil")
	}
}

// TestPhase2ReturnStatement - Return 문
func TestPhase2ReturnStatement(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"no_value", "return"},
		{"with_value", "return 42"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens := tokenize(tt.input)
			p := NewParser(tokens)
			prog, err := p.Parse()

			if err != nil {
				t.Fatalf("파싱 에러: %v", err)
			}

			_, ok := prog.Statements[0].(*ast.ReturnStmt)
			if !ok {
				t.Fatalf("ReturnStmt가 아님: %T", prog.Statements[0])
			}
		})
	}
}

// TestPhase2BreakStatement - Break 문
func TestPhase2BreakStatement(t *testing.T) {
	input := "break"
	tokens := tokenize(input)
	p := NewParser(tokens)
	prog, err := p.Parse()

	if err != nil {
		t.Fatalf("파싱 에러: %v", err)
	}

	_, ok := prog.Statements[0].(*ast.BreakStmt)
	if !ok {
		t.Fatalf("BreakStmt가 아님: %T", prog.Statements[0])
	}
}

// TestPhase2ContinueStatement - Continue 문
func TestPhase2ContinueStatement(t *testing.T) {
	input := "continue"
	tokens := tokenize(input)
	p := NewParser(tokens)
	prog, err := p.Parse()

	if err != nil {
		t.Fatalf("파싱 에러: %v", err)
	}

	_, ok := prog.Statements[0].(*ast.ContinueStmt)
	if !ok {
		t.Fatalf("ContinueStmt가 아님: %T", prog.Statements[0])
	}
}

// TestPhase2FunctionDeclaration - 함수 정의
func TestPhase2FunctionDeclaration(t *testing.T) {
	input := `function add(a::Int, b::Int)::Int
		return a + b
	end`

	tokens := tokenize(input)
	p := NewParser(tokens)
	prog, err := p.Parse()

	if err != nil {
		t.Fatalf("파싱 에러: %v", err)
	}

	funcDecl, ok := prog.Statements[0].(*ast.FunctionDecl)
	if !ok {
		t.Fatalf("FunctionDecl이 아님: %T", prog.Statements[0])
	}

	if funcDecl.Name != "add" {
		t.Errorf("함수명 예상: add, 얻음: %s", funcDecl.Name)
	}

	if len(funcDecl.Parameters) != 2 {
		t.Errorf("파라미터 개수 예상: 2, 얻음: %d", len(funcDecl.Parameters))
	}

	if funcDecl.ReturnType != "Int" {
		t.Errorf("반환 타입 예상: Int, 얻음: %s", funcDecl.ReturnType)
	}
}

// TestPhase2StructDeclaration - Struct 정의
func TestPhase2StructDeclaration(t *testing.T) {
	input := `struct Point
		x::Float64
		y::Float64
	end`

	tokens := tokenize(input)
	p := NewParser(tokens)
	prog, err := p.Parse()

	if err != nil {
		t.Fatalf("파싱 에러: %v", err)
	}

	structDecl, ok := prog.Statements[0].(*ast.StructDecl)
	if !ok {
		t.Fatalf("StructDecl이 아님: %T", prog.Statements[0])
	}

	if structDecl.Name != "Point" {
		t.Errorf("구조체명 예상: Point, 얻음: %s", structDecl.Name)
	}

	if len(structDecl.Fields) != 2 {
		t.Errorf("필드 개수 예상: 2, 얻음: %d", len(structDecl.Fields))
	}
}

// TestPhase2TryStatement - Try-Catch 문
func TestPhase2TryStatement(t *testing.T) {
	input := `try
		risky_operation()
	catch e
		println(e)
	end`

	tokens := tokenize(input)
	p := NewParser(tokens)
	prog, err := p.Parse()

	if err != nil {
		t.Fatalf("파싱 에러: %v", err)
	}

	tryStmt, ok := prog.Statements[0].(*ast.TryStmt)
	if !ok {
		t.Fatalf("TryStmt가 아님: %T", prog.Statements[0])
	}

	if len(tryStmt.Try) == 0 {
		t.Error("try 블록이 비어있음")
	}

	if len(tryStmt.Catches) != 1 {
		t.Errorf("catch 절 개수 예상: 1, 얻음: %d", len(tryStmt.Catches))
	}
}

// TestPhase2ComplexExpression - 복잡한 표현식
func TestPhase2ComplexExpression(t *testing.T) {
	input := "result = matrix[1, 2] + func(x::Int) * 3 ^ 2"
	tokens := tokenize(input)
	p := NewParser(tokens)
	prog, err := p.Parse()

	if err != nil {
		t.Fatalf("파싱 에러: %v", err)
	}

	_, ok := prog.Statements[0].(*ast.Assignment)
	if !ok {
		t.Fatalf("Assignment가 아님: %T", prog.Statements[0])
	}
}

// TestPhase2MultipleStatements - 여러 문
func TestPhase2MultipleStatements(t *testing.T) {
	input := `x = 5
y = 10
z = x + y`

	tokens := tokenize(input)
	p := NewParser(tokens)
	prog, err := p.Parse()

	if err != nil {
		t.Fatalf("파싱 에러: %v", err)
	}

	if len(prog.Statements) != 3 {
		t.Errorf("문의 개수 예상: 3, 얻음: %d", len(prog.Statements))
	}

	for i, stmt := range prog.Statements {
		if _, ok := stmt.(*ast.Assignment); !ok {
			t.Errorf("문 %d가 Assignment가 아님: %T", i, stmt)
		}
	}
}

// TestPhase2NestedCalls - 중첩된 함수 호출
func TestPhase2NestedCalls(t *testing.T) {
	input := "sin(cos(x))"
	tokens := tokenize(input)
	p := NewParser(tokens)
	prog, err := p.Parse()

	if err != nil {
		t.Fatalf("파싱 에러: %v", err)
	}

	exprStmt := prog.Statements[0].(*ast.ExprStmt)
	outerCall, ok := exprStmt.Expr.(*ast.Call)
	if !ok {
		t.Fatalf("외부 Call이 아님: %T", exprStmt.Expr)
	}

	if len(outerCall.Arguments) != 1 {
		t.Fatalf("외부 호출 인자 개수가 1이 아님: %d", len(outerCall.Arguments))
	}

	innerCall, ok := outerCall.Arguments[0].(*ast.Call)
	if !ok {
		t.Fatalf("내부 Call이 아님: %T", outerCall.Arguments[0])
	}

	if len(innerCall.Arguments) != 1 {
		t.Errorf("내부 호출 인자 개수 예상: 1, 얻음: %d", len(innerCall.Arguments))
	}
}

// BenchmarkParser - 파서 벤치마크
func BenchmarkParser(b *testing.B) {
	input := `
	function fibonacci(n::Int)::Int
		if n <= 1
			return n
		else
			return fibonacci(n - 1) + fibonacci(n - 2)
		end
	end

	for i in 1:100
		result = fibonacci(i)
		println(result)
	end

	struct Point{T <: Real}
		x::T
		y::T
	end

	function distance(p1::Point, p2::Point)::Float64
		return sqrt((p1.x - p2.x)^2 + (p1.y - p2.y)^2)
	end
	`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tokens := tokenize(input)
		p := NewParser(tokens)
		_, _ = p.Parse()
	}
}

// TestPhase2ArrayLiteral - 배열 리터럴
func TestPhase2ArrayLiteral(t *testing.T) {
	input := "[1 2; 3 4]"
	tokens := tokenize(input)
	p := NewParser(tokens)
	prog, err := p.Parse()

	if err != nil {
		t.Fatalf("파싱 에러: %v", err)
	}

	exprStmt := prog.Statements[0].(*ast.ExprStmt)
	_, ok := exprStmt.Expr.(*ast.ArrayLiteral)
	if !ok {
		t.Fatalf("ArrayLiteral이 아님: %T", exprStmt.Expr)
	}
}

// TestPhase2TupleLiteral - 튜플 리터럴
func TestPhase2TupleLiteral(t *testing.T) {
	input := "(1, 2, 3)"
	tokens := tokenize(input)
	p := NewParser(tokens)
	prog, err := p.Parse()

	if err != nil {
		t.Fatalf("파싱 에러: %v", err)
	}

	exprStmt := prog.Statements[0].(*ast.ExprStmt)
	tuple, ok := exprStmt.Expr.(*ast.TupleLiteral)
	if !ok {
		t.Fatalf("TupleLiteral이 아님: %T", exprStmt.Expr)
	}

	if len(tuple.Elements) != 3 {
		t.Errorf("요소 개수 예상: 3, 얻음: %d", len(tuple.Elements))
	}
}

// TestPhase2DictLiteral - 딕셔너리 리터럴
func TestPhase2DictLiteral(t *testing.T) {
	input := "Dict(\"a\" => 1, \"b\" => 2)"
	tokens := tokenize(input)
	p := NewParser(tokens)
	prog, err := p.Parse()

	if err != nil {
		t.Fatalf("파싱 에러: %v", err)
	}

	exprStmt := prog.Statements[0].(*ast.ExprStmt)
	call, ok := exprStmt.Expr.(*ast.Call)
	if !ok {
		t.Fatalf("Call이 아님: %T", exprStmt.Expr)
	}

	if len(call.Arguments) != 2 {
		t.Errorf("인자 개수 예상: 2, 얻음: %d", len(call.Arguments))
	}
}

// TestPhase2VarDeclaration - let 선언
func TestPhase2VarDeclaration(t *testing.T) {
	input := "let x = 5"
	tokens := tokenize(input)
	p := NewParser(tokens)
	prog, err := p.Parse()

	if err != nil {
		t.Fatalf("파싱 에러: %v", err)
	}

	varDecl, ok := prog.Statements[0].(*ast.VarDecl)
	if !ok {
		t.Fatalf("VarDecl이 아님: %T", prog.Statements[0])
	}

	if varDecl.Name != "x" {
		t.Errorf("변수명 예상: x, 얻음: %s", varDecl.Name)
	}
}

// TestPhase2ConstDeclaration - const 선언
func TestPhase2ConstDeclaration(t *testing.T) {
	input := "const PI = 3.14159"
	tokens := tokenize(input)
	p := NewParser(tokens)
	prog, err := p.Parse()

	if err != nil {
		t.Fatalf("파싱 에러: %v", err)
	}

	constDecl, ok := prog.Statements[0].(*ast.ConstDecl)
	if !ok {
		t.Fatalf("ConstDecl이 아님: %T", prog.Statements[0])
	}

	if constDecl.Name != "PI" {
		t.Errorf("상수명 예상: PI, 얻음: %s", constDecl.Name)
	}
}
