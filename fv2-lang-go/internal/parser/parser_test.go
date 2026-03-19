package parser

import (
	"testing"

	"fv2-lang/internal/ast"
	"fv2-lang/internal/lexer"
)

func TestParseFunctionDef(t *testing.T) {
	input := "fn main() { let x = 5; }"
	lex, _ := lexer.New(input)
	tokens, _ := lex.Tokenize()
	parser := New(tokens)
	program, err := parser.Parse()

	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(program.Definitions) != 1 {
		t.Errorf("Expected 1 definition, got %d", len(program.Definitions))
	}

	fn, ok := program.Definitions[0].(*ast.FunctionDef)
	if !ok {
		t.Fatalf("Expected FunctionDef, got %T", program.Definitions[0])
	}

	if fn.Name != "main" {
		t.Errorf("Expected function name 'main', got '%s'", fn.Name)
	}

	if len(fn.Body) != 1 {
		t.Errorf("Expected 1 statement in body, got %d", len(fn.Body))
	}
}

func TestParseFunctionWithParams(t *testing.T) {
	input := "fn add(x:i64, y:i64) i64 { let result = x + y; return result; }"
	lex, _ := lexer.New(input)
	tokens, _ := lex.Tokenize()
	parser := New(tokens)
	program, err := parser.Parse()

	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	fn, ok := program.Definitions[0].(*ast.FunctionDef)
	if !ok {
		t.Fatalf("Expected FunctionDef")
	}

	if len(fn.Parameters) != 2 {
		t.Errorf("Expected 2 parameters, got %d", len(fn.Parameters))
	}

	if fn.Parameters[0].Name != "x" {
		t.Errorf("Expected parameter name 'x', got '%s'", fn.Parameters[0].Name)
	}

	if fn.ReturnType == nil || fn.ReturnType.Name != "i64" {
		t.Errorf("Expected return type i64")
	}
}

func TestParseLetStatement(t *testing.T) {
	input := "fn main() { let x = 42; }"
	lex, _ := lexer.New(input)
	tokens, _ := lex.Tokenize()
	parser := New(tokens)
	program, err := parser.Parse()

	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	fn := program.Definitions[0].(*ast.FunctionDef)
	stmt := fn.Body[0].(*ast.LetStatement)

	if stmt.Name != "x" {
		t.Errorf("Expected 'x', got '%s'", stmt.Name)
	}

	if _, ok := stmt.Init.(*ast.IntegerLiteral); !ok {
		t.Errorf("Expected IntegerLiteral, got %T", stmt.Init)
	}
}

func TestParseConstStatement(t *testing.T) {
	input := "fn main() { const MAX = 100; }"
	lex, _ := lexer.New(input)
	tokens, _ := lex.Tokenize()
	parser := New(tokens)
	program, err := parser.Parse()

	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	fn := program.Definitions[0].(*ast.FunctionDef)
	stmt := fn.Body[0].(*ast.ConstStatement)

	if stmt.Name != "MAX" {
		t.Errorf("Expected 'MAX', got '%s'", stmt.Name)
	}
}

func TestParseIfStatement(t *testing.T) {
	input := "fn main() { if x > 0 { let y = 1; } else { let y = 2; } }"
	lex, _ := lexer.New(input)
	tokens, _ := lex.Tokenize()
	parser := New(tokens)
	program, err := parser.Parse()

	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	fn := program.Definitions[0].(*ast.FunctionDef)
	ifStmt, ok := fn.Body[0].(*ast.IfStatement)
	if !ok {
		t.Fatalf("Expected IfStatement, got %T", fn.Body[0])
	}

	if len(ifStmt.ThenBody) != 1 {
		t.Errorf("Expected 1 statement in then body, got %d", len(ifStmt.ThenBody))
	}

	if len(ifStmt.ElseBody) != 1 {
		t.Errorf("Expected 1 statement in else body, got %d", len(ifStmt.ElseBody))
	}
}

func TestParseForLoop(t *testing.T) {
	input := "fn main() { for i in 0..10 { let x = i; } }"
	lex, _ := lexer.New(input)
	tokens, _ := lex.Tokenize()
	parser := New(tokens)
	program, err := parser.Parse()

	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	fn := program.Definitions[0].(*ast.FunctionDef)
	forStmt, ok := fn.Body[0].(*ast.ForRangeStatement)
	if !ok {
		t.Fatalf("Expected ForRangeStatement, got %T", fn.Body[0])
	}

	if forStmt.Variable != "i" {
		t.Errorf("Expected 'i', got '%s'", forStmt.Variable)
	}

	if len(forStmt.Body) != 1 {
		t.Errorf("Expected 1 statement in body, got %d", len(forStmt.Body))
	}
}

func TestParseReturnStatement(t *testing.T) {
	input := "fn getValue() i64 { return 42; }"
	lex, _ := lexer.New(input)
	tokens, _ := lex.Tokenize()
	parser := New(tokens)
	program, err := parser.Parse()

	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	fn := program.Definitions[0].(*ast.FunctionDef)
	returnStmt, ok := fn.Body[0].(*ast.ReturnStatement)
	if !ok {
		t.Fatalf("Expected ReturnStatement, got %T", fn.Body[0])
	}

	if _, ok := returnStmt.Value.(*ast.IntegerLiteral); !ok {
		t.Errorf("Expected IntegerLiteral, got %T", returnStmt.Value)
	}
}

func TestParseBinaryExpression(t *testing.T) {
	input := "fn main() { let result = 10 + 20; }"
	lex, _ := lexer.New(input)
	tokens, _ := lex.Tokenize()
	parser := New(tokens)
	program, err := parser.Parse()

	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	fn := program.Definitions[0].(*ast.FunctionDef)
	letStmt := fn.Body[0].(*ast.LetStatement)
	binExpr, ok := letStmt.Init.(*ast.BinaryExpression)
	if !ok {
		t.Fatalf("Expected BinaryExpression, got %T", letStmt.Init)
	}

	if binExpr.Operator != "+" {
		t.Errorf("Expected '+', got '%s'", binExpr.Operator)
	}
}

func TestParseUnaryExpression(t *testing.T) {
	input := "fn main() { let x = -42; }"
	lex, _ := lexer.New(input)
	tokens, _ := lex.Tokenize()
	parser := New(tokens)
	program, err := parser.Parse()

	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	fn := program.Definitions[0].(*ast.FunctionDef)
	letStmt := fn.Body[0].(*ast.LetStatement)
	unaryExpr, ok := letStmt.Init.(*ast.UnaryExpression)
	if !ok {
		t.Fatalf("Expected UnaryExpression, got %T", letStmt.Init)
	}

	if unaryExpr.Operator != "-" {
		t.Errorf("Expected '-', got '%s'", unaryExpr.Operator)
	}
}

func TestParseFunctionCall(t *testing.T) {
	input := "fn main() { let result = add(10, 20); }"
	lex, _ := lexer.New(input)
	tokens, _ := lex.Tokenize()
	parser := New(tokens)
	program, err := parser.Parse()

	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	fn := program.Definitions[0].(*ast.FunctionDef)
	letStmt := fn.Body[0].(*ast.LetStatement)
	callExpr, ok := letStmt.Init.(*ast.CallExpression)
	if !ok {
		t.Fatalf("Expected CallExpression, got %T", letStmt.Init)
	}

	if len(callExpr.Arguments) != 2 {
		t.Errorf("Expected 2 arguments, got %d", len(callExpr.Arguments))
	}
}

func TestParseFieldAccess(t *testing.T) {
	input := "fn main() { let x = obj.field; }"
	lex, _ := lexer.New(input)
	tokens, _ := lex.Tokenize()
	parser := New(tokens)
	program, err := parser.Parse()

	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	fn := program.Definitions[0].(*ast.FunctionDef)
	letStmt := fn.Body[0].(*ast.LetStatement)
	fieldExpr, ok := letStmt.Init.(*ast.FieldExpression)
	if !ok {
		t.Fatalf("Expected FieldExpression, got %T", letStmt.Init)
	}

	if fieldExpr.Field != "field" {
		t.Errorf("Expected 'field', got '%s'", fieldExpr.Field)
	}
}

func TestParseIndexExpression(t *testing.T) {
	input := "fn main() { let x = arr[0]; }"
	lex, _ := lexer.New(input)
	tokens, _ := lex.Tokenize()
	parser := New(tokens)
	program, err := parser.Parse()

	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	fn := program.Definitions[0].(*ast.FunctionDef)
	letStmt := fn.Body[0].(*ast.LetStatement)
	indexExpr, ok := letStmt.Init.(*ast.IndexExpression)
	if !ok {
		t.Fatalf("Expected IndexExpression, got %T", letStmt.Init)
	}

	if _, ok := indexExpr.Index.(*ast.IntegerLiteral); !ok {
		t.Errorf("Expected IntegerLiteral index")
	}
}

func TestParseArrayLiteral(t *testing.T) {
	input := "fn main() { let arr = [1, 2, 3]; }"
	lex, _ := lexer.New(input)
	tokens, _ := lex.Tokenize()
	parser := New(tokens)
	program, err := parser.Parse()

	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	fn := program.Definitions[0].(*ast.FunctionDef)
	letStmt := fn.Body[0].(*ast.LetStatement)
	arrayExpr, ok := letStmt.Init.(*ast.ArrayExpression)
	if !ok {
		t.Fatalf("Expected ArrayExpression, got %T", letStmt.Init)
	}

	if len(arrayExpr.Elements) != 3 {
		t.Errorf("Expected 3 elements, got %d", len(arrayExpr.Elements))
	}
}

func TestParseStructDef(t *testing.T) {
	input := "struct Point { x i64, y i64 }"
	lex, _ := lexer.New(input)
	tokens, _ := lex.Tokenize()
	parser := New(tokens)
	program, err := parser.Parse()

	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	structDef, ok := program.Definitions[0].(*ast.StructDef)
	if !ok {
		t.Fatalf("Expected StructDef, got %T", program.Definitions[0])
	}

	if structDef.Name != "Point" {
		t.Errorf("Expected 'Point', got '%s'", structDef.Name)
	}

	if len(structDef.Fields) != 2 {
		t.Errorf("Expected 2 fields, got %d", len(structDef.Fields))
	}
}

func TestParseTypeDef(t *testing.T) {
	input := "type UserId = i64"
	lex, _ := lexer.New(input)
	tokens, _ := lex.Tokenize()
	parser := New(tokens)
	program, err := parser.Parse()

	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	typeDef, ok := program.Definitions[0].(*ast.TypeDef)
	if !ok {
		t.Fatalf("Expected TypeDef, got %T", program.Definitions[0])
	}

	if typeDef.Name != "UserId" {
		t.Errorf("Expected 'UserId', got '%s'", typeDef.Name)
	}
}

func TestParseMatchExpression(t *testing.T) {
	input := "fn main() { match x { 1 => { let y = 10; }, 2 => { let y = 20; }, _ => { let y = 0; } } }"
	lex, _ := lexer.New(input)
	tokens, _ := lex.Tokenize()
	parser := New(tokens)
	program, err := parser.Parse()

	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	fn := program.Definitions[0].(*ast.FunctionDef)
	matchStmt, ok := fn.Body[0].(*ast.MatchStatement)
	if !ok {
		t.Fatalf("Expected MatchStatement, got %T", fn.Body[0])
	}

	if len(matchStmt.Arms) != 3 {
		t.Errorf("Expected 3 arms, got %d", len(matchStmt.Arms))
	}
}

func TestParseOperatorPrecedence(t *testing.T) {
	input := "fn main() { let x = 2 + 3 * 4; }"
	lex, _ := lexer.New(input)
	tokens, _ := lex.Tokenize()
	parser := New(tokens)
	program, err := parser.Parse()

	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	fn := program.Definitions[0].(*ast.FunctionDef)
	letStmt := fn.Body[0].(*ast.LetStatement)
	binExpr := letStmt.Init.(*ast.BinaryExpression)

	// Should be 2 + (3 * 4), so root is +
	if binExpr.Operator != "+" {
		t.Errorf("Expected root operator '+', got '%s'", binExpr.Operator)
	}

	// Right side should be 3 * 4
	mulExpr, ok := binExpr.Right.(*ast.BinaryExpression)
	if !ok {
		t.Errorf("Expected multiplication on right side")
	}

	if mulExpr.Operator != "*" {
		t.Errorf("Expected '*' operator, got '%s'", mulExpr.Operator)
	}
}

func TestParseErrorPropagation(t *testing.T) {
	input := "fn main() { let result = getValue()?; }"
	lex, _ := lexer.New(input)
	tokens, _ := lex.Tokenize()
	parser := New(tokens)
	program, err := parser.Parse()

	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	fn := program.Definitions[0].(*ast.FunctionDef)
	letStmt := fn.Body[0].(*ast.LetStatement)
	errProp, ok := letStmt.Init.(*ast.ErrorPropagation)
	if !ok {
		t.Fatalf("Expected ErrorPropagation, got %T", letStmt.Init)
	}

	if _, ok := errProp.Expression.(*ast.CallExpression); !ok {
		t.Errorf("Expected CallExpression inside error propagation")
	}
}

func TestParseMultipleFunctions(t *testing.T) {
	input := `
		fn add(x:i64, y:i64) i64 { return x + y; }
		fn mul(x:i64, y:i64) i64 { return x * y; }
		fn main() { let result = add(2, 3); }
	`
	lex, _ := lexer.New(input)
	tokens, _ := lex.Tokenize()
	parser := New(tokens)
	program, err := parser.Parse()

	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(program.Definitions) != 3 {
		t.Errorf("Expected 3 functions, got %d", len(program.Definitions))
	}
}

func TestParseStringLiteral(t *testing.T) {
	input := `fn main() { let msg = "Hello, FV 2.0!"; }`
	lex, _ := lexer.New(input)
	tokens, _ := lex.Tokenize()
	parser := New(tokens)
	program, err := parser.Parse()

	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	fn := program.Definitions[0].(*ast.FunctionDef)
	letStmt := fn.Body[0].(*ast.LetStatement)
	strLit, ok := letStmt.Init.(*ast.StringLiteral)
	if !ok {
		t.Fatalf("Expected StringLiteral, got %T", letStmt.Init)
	}

	if strLit.Value != "Hello, FV 2.0!" {
		t.Errorf("Expected 'Hello, FV 2.0!', got '%s'", strLit.Value)
	}
}

func TestParseFloatLiteral(t *testing.T) {
	input := `fn main() { let pi = 3.14159; }`
	lex, _ := lexer.New(input)
	tokens, _ := lex.Tokenize()
	parser := New(tokens)
	program, err := parser.Parse()

	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	fn := program.Definitions[0].(*ast.FunctionDef)
	letStmt := fn.Body[0].(*ast.LetStatement)
	floatLit, ok := letStmt.Init.(*ast.FloatLiteral)
	if !ok {
		t.Fatalf("Expected FloatLiteral, got %T", letStmt.Init)
	}

	if floatLit.Value != 3.14159 {
		t.Errorf("Expected 3.14159, got %f", floatLit.Value)
	}
}

func TestParseBooleanLiteral(t *testing.T) {
	input := `fn main() { let flag = true; let other = false; }`
	lex, _ := lexer.New(input)
	tokens, _ := lex.Tokenize()
	parser := New(tokens)
	program, err := parser.Parse()

	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	fn := program.Definitions[0].(*ast.FunctionDef)
	letStmt1 := fn.Body[0].(*ast.LetStatement)
	boolLit1, ok := letStmt1.Init.(*ast.BoolLiteral)
	if !ok {
		t.Fatalf("Expected BoolLiteral, got %T", letStmt1.Init)
	}

	if !boolLit1.Value {
		t.Errorf("Expected true")
	}

	letStmt2 := fn.Body[1].(*ast.LetStatement)
	boolLit2 := letStmt2.Init.(*ast.BoolLiteral)

	if boolLit2.Value {
		t.Errorf("Expected false")
	}
}

func TestParseComplexExpression(t *testing.T) {
	input := `fn main() { let x = (a + b) * (c - d); }`
	lex, _ := lexer.New(input)
	tokens, _ := lex.Tokenize()
	parser := New(tokens)
	program, err := parser.Parse()

	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	fn := program.Definitions[0].(*ast.FunctionDef)
	letStmt := fn.Body[0].(*ast.LetStatement)
	mulExpr, ok := letStmt.Init.(*ast.BinaryExpression)
	if !ok {
		t.Fatalf("Expected BinaryExpression")
	}

	if mulExpr.Operator != "*" {
		t.Errorf("Expected '*', got '%s'", mulExpr.Operator)
	}
}

func TestParseMethodCall(t *testing.T) {
	input := `fn main() { let result = obj.method(arg); }`
	lex, _ := lexer.New(input)
	tokens, _ := lex.Tokenize()
	parser := New(tokens)
	program, err := parser.Parse()

	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	fn := program.Definitions[0].(*ast.FunctionDef)
	letStmt := fn.Body[0].(*ast.LetStatement)
	methodCall, ok := letStmt.Init.(*ast.MethodCallExpression)
	if !ok {
		t.Fatalf("Expected MethodCallExpression, got %T", letStmt.Init)
	}

	if methodCall.Method != "method" {
		t.Errorf("Expected 'method', got '%s'", methodCall.Method)
	}

	if len(methodCall.Arguments) != 1 {
		t.Errorf("Expected 1 argument, got %d", len(methodCall.Arguments))
	}
}

func TestParseLogicalOperators(t *testing.T) {
	input := `fn main() { let flag = true && false || true; }`
	lex, _ := lexer.New(input)
	tokens, _ := lex.Tokenize()
	parser := New(tokens)
	program, err := parser.Parse()

	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	fn := program.Definitions[0].(*ast.FunctionDef)
	letStmt := fn.Body[0].(*ast.LetStatement)

	// Should be (true && false) || true
	orExpr, ok := letStmt.Init.(*ast.BinaryExpression)
	if !ok {
		t.Fatalf("Expected BinaryExpression")
	}

	if orExpr.Operator != "||" {
		t.Errorf("Expected '||', got '%s'", orExpr.Operator)
	}
}

func TestParseNoneLiteral(t *testing.T) {
	input := `fn main() { let x = none; }`
	lex, _ := lexer.New(input)
	tokens, _ := lex.Tokenize()
	parser := New(tokens)
	program, err := parser.Parse()

	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	fn := program.Definitions[0].(*ast.FunctionDef)
	letStmt := fn.Body[0].(*ast.LetStatement)
	noneLit, ok := letStmt.Init.(*ast.NoneLiteral)
	if !ok {
		t.Fatalf("Expected NoneLiteral, got %T", letStmt.Init)
	}

	_ = noneLit
}

func TestParseIfExpressionAsValue(t *testing.T) {
	input := `fn main() { let x = if cond { 10 } else { 20 }; }`
	lex, _ := lexer.New(input)
	tokens, _ := lex.Tokenize()
	parser := New(tokens)
	program, err := parser.Parse()

	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	fn := program.Definitions[0].(*ast.FunctionDef)
	letStmt := fn.Body[0].(*ast.LetStatement)
	ifExpr, ok := letStmt.Init.(*ast.IfExpression)
	if !ok {
		t.Fatalf("Expected IfExpression, got %T", letStmt.Init)
	}

	if ifExpr.ThenExpr == nil || ifExpr.ElseExpr == nil {
		t.Errorf("Expected both then and else expressions")
	}
}
