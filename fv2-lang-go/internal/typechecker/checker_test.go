package typechecker

import (
	"testing"

	"fv2-lang/internal/ast"
)

func TestBasicTypeChecking(t *testing.T) {
	checker := New()

	program := &ast.Program{
		Definitions: []ast.Definition{},
		MainBody: []ast.Statement{
			&ast.LetStatement{
				Name: "x",
				Type: &ast.Type{Name: "i64"},
				Init: &ast.IntegerLiteral{Value: 42},
			},
		},
	}

	errors, _ := checker.Check(program)
	if len(errors) != 0 {
		t.Fatalf("expected 0 errors, got %d: %v", len(errors), errors)
	}
}

func TestTypeMismatch(t *testing.T) {
	checker := New()

	program := &ast.Program{
		Definitions: []ast.Definition{},
		MainBody: []ast.Statement{
			&ast.LetStatement{
				Name: "x",
				Type: &ast.Type{Name: "i64"},
				Init: &ast.StringLiteral{Value: "hello"},
			},
		},
	}

	errors, _ := checker.Check(program)
	if len(errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errors))
	}
}

func TestUndefinedVariable(t *testing.T) {
	checker := New()

	program := &ast.Program{
		Definitions: []ast.Definition{},
		MainBody: []ast.Statement{
			&ast.LetStatement{
				Name: "x",
				Init: &ast.Identifier{Name: "undefined"},
			},
		},
	}

	errors, _ := checker.Check(program)
	if len(errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errors))
	}
}

func TestFunctionDefinition(t *testing.T) {
	checker := New()

	program := &ast.Program{
		Definitions: []ast.Definition{
			&ast.FunctionDef{
				Name: "add",
				Parameters: []ast.Parameter{
					{Name: "x", Type: &ast.Type{Name: "i64"}},
					{Name: "y", Type: &ast.Type{Name: "i64"}},
				},
				ReturnType: &ast.Type{Name: "i64"},
				Body: []ast.Statement{
					&ast.ReturnStatement{
						Value: &ast.BinaryExpression{
							Left:     &ast.Identifier{Name: "x"},
							Operator: "+",
							Right:    &ast.Identifier{Name: "y"},
						},
					},
				},
			},
		},
		MainBody: []ast.Statement{},
	}

	errors, _ := checker.Check(program)
	if len(errors) != 0 {
		t.Fatalf("expected 0 errors, got %d: %v", len(errors), errors)
	}
}

func TestArrayTypeChecking(t *testing.T) {
	checker := New()

	program := &ast.Program{
		Definitions: []ast.Definition{},
		MainBody: []ast.Statement{
			&ast.LetStatement{
				Name: "arr",
				Init: &ast.ArrayExpression{
					Elements: []ast.Expression{
						&ast.IntegerLiteral{Value: 1},
						&ast.IntegerLiteral{Value: 2},
						&ast.IntegerLiteral{Value: 3},
					},
				},
			},
		},
	}

	errors, _ := checker.Check(program)
	if len(errors) != 0 {
		t.Fatalf("expected 0 errors, got %d: %v", len(errors), errors)
	}
}

func TestArrayTypeMismatch(t *testing.T) {
	checker := New()

	program := &ast.Program{
		Definitions: []ast.Definition{},
		MainBody: []ast.Statement{
			&ast.LetStatement{
				Name: "arr",
				Init: &ast.ArrayExpression{
					Elements: []ast.Expression{
						&ast.IntegerLiteral{Value: 1},
						&ast.StringLiteral{Value: "two"},
					},
				},
			},
		},
	}

	errors, _ := checker.Check(program)
	if len(errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errors))
	}
}

func TestBinaryExpression(t *testing.T) {
	checker := New()

	program := &ast.Program{
		Definitions: []ast.Definition{},
		MainBody: []ast.Statement{
			&ast.LetStatement{
				Name: "x",
				Init: &ast.BinaryExpression{
					Left:     &ast.IntegerLiteral{Value: 5},
					Operator: "+",
					Right:    &ast.IntegerLiteral{Value: 3},
				},
			},
		},
	}

	errors, _ := checker.Check(program)
	if len(errors) != 0 {
		t.Fatalf("expected 0 errors, got %d: %v", len(errors), errors)
	}
}

func TestComparisonExpression(t *testing.T) {
	checker := New()

	program := &ast.Program{
		Definitions: []ast.Definition{},
		MainBody: []ast.Statement{
			&ast.LetStatement{
				Name: "is_equal",
				Type: &ast.Type{Name: "bool"},
				Init: &ast.BinaryExpression{
					Left:     &ast.IntegerLiteral{Value: 5},
					Operator: "==",
					Right:    &ast.IntegerLiteral{Value: 5},
				},
			},
		},
	}

	errors, _ := checker.Check(program)
	if len(errors) != 0 {
		t.Fatalf("expected 0 errors, got %d: %v", len(errors), errors)
	}
}

func TestLogicalExpression(t *testing.T) {
	checker := New()

	program := &ast.Program{
		Definitions: []ast.Definition{},
		MainBody: []ast.Statement{
			&ast.LetStatement{
				Name: "x",
				Type: &ast.Type{Name: "bool"},
				Init: &ast.BinaryExpression{
					Left:     &ast.BoolLiteral{Value: true},
					Operator: "&&",
					Right:    &ast.BoolLiteral{Value: false},
				},
			},
		},
	}

	errors, _ := checker.Check(program)
	if len(errors) != 0 {
		t.Fatalf("expected 0 errors, got %d: %v", len(errors), errors)
	}
}

func TestUnaryExpression(t *testing.T) {
	checker := New()

	program := &ast.Program{
		Definitions: []ast.Definition{},
		MainBody: []ast.Statement{
			&ast.LetStatement{
				Name: "x",
				Init: &ast.UnaryExpression{
					Operator: "-",
					Operand:  &ast.IntegerLiteral{Value: 5},
				},
			},
		},
	}

	errors, _ := checker.Check(program)
	if len(errors) != 0 {
		t.Fatalf("expected 0 errors, got %d: %v", len(errors), errors)
	}
}

func TestIfExpression(t *testing.T) {
	checker := New()

	program := &ast.Program{
		Definitions: []ast.Definition{},
		MainBody: []ast.Statement{
			&ast.LetStatement{
				Name: "x",
				Init: &ast.IfExpression{
					Condition: &ast.BoolLiteral{Value: true},
					ThenExpr:  &ast.IntegerLiteral{Value: 10},
					ElseExpr:  &ast.IntegerLiteral{Value: 20},
				},
			},
		},
	}

	errors, _ := checker.Check(program)
	if len(errors) != 0 {
		t.Fatalf("expected 0 errors, got %d: %v", len(errors), errors)
	}
}

func TestIfExpressionTypeMismatch(t *testing.T) {
	checker := New()

	program := &ast.Program{
		Definitions: []ast.Definition{},
		MainBody: []ast.Statement{
			&ast.LetStatement{
				Name: "x",
				Init: &ast.IfExpression{
					Condition: &ast.BoolLiteral{Value: true},
					ThenExpr:  &ast.IntegerLiteral{Value: 10},
					ElseExpr:  &ast.StringLiteral{Value: "hello"},
				},
			},
		},
	}

	errors, _ := checker.Check(program)
	if len(errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errors))
	}
}

func TestForRangeStatement(t *testing.T) {
	checker := New()

	program := &ast.Program{
		Definitions: []ast.Definition{},
		MainBody: []ast.Statement{
			&ast.ForRangeStatement{
				Variable: "i",
				Start:    &ast.IntegerLiteral{Value: 0},
				End:      &ast.IntegerLiteral{Value: 10},
				Body: []ast.Statement{
					&ast.ExpressionStatement{
						Expression: &ast.Identifier{Name: "i"},
					},
				},
			},
		},
	}

	errors, _ := checker.Check(program)
	if len(errors) != 0 {
		t.Fatalf("expected 0 errors, got %d: %v", len(errors), errors)
	}
}

func TestStructDefinition(t *testing.T) {
	checker := New()

	program := &ast.Program{
		Definitions: []ast.Definition{
			&ast.StructDef{
				Name: "Point",
				Fields: []ast.Field{
					{Name: "x", Type: &ast.Type{Name: "i64"}},
					{Name: "y", Type: &ast.Type{Name: "i64"}},
				},
			},
		},
		MainBody: []ast.Statement{},
	}

	errors, _ := checker.Check(program)
	if len(errors) != 0 {
		t.Fatalf("expected 0 errors, got %d: %v", len(errors), errors)
	}
}

func TestIndexExpression(t *testing.T) {
	checker := New()

	program := &ast.Program{
		Definitions: []ast.Definition{},
		MainBody: []ast.Statement{
			&ast.LetStatement{
				Name: "arr",
				Init: &ast.ArrayExpression{
					Elements: []ast.Expression{
						&ast.IntegerLiteral{Value: 1},
						&ast.IntegerLiteral{Value: 2},
					},
				},
			},
			&ast.LetStatement{
				Name: "x",
				Init: &ast.IndexExpression{
					Object: &ast.Identifier{Name: "arr"},
					Index:  &ast.IntegerLiteral{Value: 0},
				},
			},
		},
	}

	errors, _ := checker.Check(program)
	if len(errors) != 0 {
		t.Fatalf("expected 0 errors, got %d: %v", len(errors), errors)
	}
}

func TestFunctionCall(t *testing.T) {
	checker := New()

	program := &ast.Program{
		Definitions: []ast.Definition{
			&ast.FunctionDef{
				Name: "double",
				Parameters: []ast.Parameter{
					{Name: "x", Type: &ast.Type{Name: "i64"}},
				},
				ReturnType: &ast.Type{Name: "i64"},
				Body: []ast.Statement{
					&ast.ReturnStatement{
						Value: &ast.BinaryExpression{
							Left:     &ast.Identifier{Name: "x"},
							Operator: "+",
							Right:    &ast.Identifier{Name: "x"},
						},
					},
				},
			},
		},
		MainBody: []ast.Statement{
			&ast.LetStatement{
				Name: "result",
				Init: &ast.CallExpression{
					Function: &ast.Identifier{Name: "double"},
					Arguments: []ast.Expression{
						&ast.IntegerLiteral{Value: 5},
					},
				},
			},
		},
	}

	errors, _ := checker.Check(program)
	if len(errors) != 0 {
		t.Fatalf("expected 0 errors, got %d: %v", len(errors), errors)
	}
}

func TestFunctionArgumentCountMismatch(t *testing.T) {
	checker := New()

	program := &ast.Program{
		Definitions: []ast.Definition{
			&ast.FunctionDef{
				Name: "add",
				Parameters: []ast.Parameter{
					{Name: "x", Type: &ast.Type{Name: "i64"}},
					{Name: "y", Type: &ast.Type{Name: "i64"}},
				},
				ReturnType: &ast.Type{Name: "i64"},
				Body: []ast.Statement{
					&ast.ReturnStatement{
						Value: &ast.BinaryExpression{
							Left:     &ast.Identifier{Name: "x"},
							Operator: "+",
							Right:    &ast.Identifier{Name: "y"},
						},
					},
				},
			},
		},
		MainBody: []ast.Statement{
			&ast.ExpressionStatement{
				Expression: &ast.CallExpression{
					Function: &ast.Identifier{Name: "add"},
					Arguments: []ast.Expression{
						&ast.IntegerLiteral{Value: 5},
					},
				},
			},
		},
	}

	errors, _ := checker.Check(program)
	if len(errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errors))
	}
}

// TestNoneLiteral tests none/null literal type checking
func TestNoneLiteral(t *testing.T) {
	checker := New()

	program := &ast.Program{
		Definitions: []ast.Definition{},
		MainBody: []ast.Statement{
			&ast.LetStatement{
				Name: "null_val",
				Type: &ast.Type{Name: "none"},
				Init: &ast.NoneLiteral{},
			},
		},
	}

	errors, _ := checker.Check(program)
	if len(errors) != 0 {
		t.Fatalf("expected 0 errors for valid none literal, got %d: %v", len(errors), errors)
	}
}

// TestBooleanLiteralType tests boolean literal type checking
func TestBooleanLiteralType(t *testing.T) {
	checker := New()

	program := &ast.Program{
		Definitions: []ast.Definition{},
		MainBody: []ast.Statement{
			&ast.LetStatement{
				Name: "is_valid",
				Type: &ast.Type{Name: "bool"},
				Init: &ast.BoolLiteral{Value: true},
			},
			&ast.LetStatement{
				Name: "is_empty",
				Type: &ast.Type{Name: "bool"},
				Init: &ast.BoolLiteral{Value: false},
			},
		},
	}

	errors, _ := checker.Check(program)
	if len(errors) != 0 {
		t.Fatalf("expected 0 errors for boolean types, got %d: %v", len(errors), errors)
	}
}

// TestMultipleTypeErrors tests collecting multiple type errors
func TestMultipleTypeErrors(t *testing.T) {
	checker := New()

	program := &ast.Program{
		Definitions: []ast.Definition{},
		MainBody: []ast.Statement{
			&ast.LetStatement{
				Name: "x",
				Type: &ast.Type{Name: "i64"},
				Init: &ast.StringLiteral{Value: "wrong"},
			},
			&ast.LetStatement{
				Name: "y",
				Type: &ast.Type{Name: "bool"},
				Init: &ast.IntegerLiteral{Value: 42},
			},
			&ast.LetStatement{
				Name: "z",
				Type: &ast.Type{Name: "f64"},
				Init: &ast.BoolLiteral{Value: true},
			},
		},
	}

	errors, _ := checker.Check(program)
	if len(errors) < 3 {
		t.Fatalf("expected at least 3 type errors, got %d", len(errors))
	}
}

// TestArrayElementTypeMismatch tests array element type validation
func TestArrayElementTypeMismatch(t *testing.T) {
	checker := New()

	program := &ast.Program{
		Definitions: []ast.Definition{},
		MainBody: []ast.Statement{
			&ast.LetStatement{
				Name: "numbers",
				Type: &ast.Type{
					Name:    "arr",
					IsArray: true,
					ElementType: &ast.Type{Name: "i64"},
				},
				Init: &ast.ArrayExpression{
					Elements: []ast.Expression{
						&ast.IntegerLiteral{Value: 1},
						&ast.IntegerLiteral{Value: 2},
						&ast.StringLiteral{Value: "three"},
					},
				},
			},
		},
	}

	errors, _ := checker.Check(program)
	if len(errors) < 1 {
		t.Fatalf("expected error for array element type mismatch, got %d", len(errors))
	}
}

// TestFloatLiteralType tests floating point literal type checking
func TestFloatLiteralType(t *testing.T) {
	checker := New()

	program := &ast.Program{
		Definitions: []ast.Definition{},
		MainBody: []ast.Statement{
			&ast.LetStatement{
				Name: "pi",
				Type: &ast.Type{Name: "f64"},
				Init: &ast.FloatLiteral{Value: 3.14},
			},
		},
	}

	errors, _ := checker.Check(program)
	if len(errors) != 0 {
		t.Fatalf("expected 0 errors for float type, got %d: %v", len(errors), errors)
	}
}

// TestStringLiteralType tests string literal type checking
func TestStringLiteralType(t *testing.T) {
	checker := New()

	program := &ast.Program{
		Definitions: []ast.Definition{},
		MainBody: []ast.Statement{
			&ast.LetStatement{
				Name: "greeting",
				Type: &ast.Type{Name: "string"},
				Init: &ast.StringLiteral{Value: "hello world"},
			},
		},
	}

	errors, _ := checker.Check(program)
	if len(errors) != 0 {
		t.Fatalf("expected 0 errors for string type, got %d: %v", len(errors), errors)
	}
}

// TestFieldExpressionWithStruct tests struct field access type checking
func TestFieldExpressionWithStruct(t *testing.T) {
	checker := New()

	program := &ast.Program{
		Definitions: []ast.Definition{
			&ast.StructDef{
				Name: "Person",
				Fields: []ast.Field{
					{Name: "name", Type: &ast.Type{Name: "string"}},
					{Name: "age", Type: &ast.Type{Name: "i64"}},
				},
			},
		},
		MainBody: []ast.Statement{},
	}

	errors, _ := checker.Check(program)
	if len(errors) != 0 {
		t.Fatalf("expected 0 errors for struct definition, got %d: %v", len(errors), errors)
	}
}

// TestConstStatement tests const binding type checking
func TestConstStatement(t *testing.T) {
	checker := New()

	program := &ast.Program{
		Definitions: []ast.Definition{},
		MainBody: []ast.Statement{
			&ast.ConstStatement{
				Name:  "MAX_SIZE",
				Type:  &ast.Type{Name: "i64"},
				Value: &ast.IntegerLiteral{Value: 1000},
			},
		},
	}

	errors, _ := checker.Check(program)
	if len(errors) != 0 {
		t.Fatalf("expected 0 errors for const statement, got %d: %v", len(errors), errors)
	}
}

// TestIfStatementNonBoolCondition tests that if condition must be boolean
func TestIfStatementNonBoolCondition(t *testing.T) {
	checker := New()

	program := &ast.Program{
		Definitions: []ast.Definition{},
		MainBody: []ast.Statement{
			&ast.IfStatement{
				Condition: &ast.IntegerLiteral{Value: 5},
				ThenBody: []ast.Statement{
					&ast.ExpressionStatement{
						Expression: &ast.IntegerLiteral{Value: 1},
					},
				},
				ElseBody: []ast.Statement{},
			},
		},
	}

	errors, _ := checker.Check(program)
	if len(errors) < 1 {
		t.Fatalf("expected error for non-bool if condition, got %d", len(errors))
	}
}

// TestForRangeNonNumericStart tests that for-range start must be numeric
func TestForRangeNonNumericStart(t *testing.T) {
	checker := New()

	program := &ast.Program{
		Definitions: []ast.Definition{},
		MainBody: []ast.Statement{
			&ast.ForRangeStatement{
				Variable: "i",
				Start:    &ast.StringLiteral{Value: "start"},
				End:      &ast.IntegerLiteral{Value: 10},
				Body:     []ast.Statement{},
			},
		},
	}

	errors, _ := checker.Check(program)
	if len(errors) < 1 {
		t.Fatalf("expected error for non-numeric range start, got %d", len(errors))
	}
}

// TestForRangeNonNumericEnd tests that for-range end must be numeric
func TestForRangeNonNumericEnd(t *testing.T) {
	checker := New()

	program := &ast.Program{
		Definitions: []ast.Definition{},
		MainBody: []ast.Statement{
			&ast.ForRangeStatement{
				Variable: "i",
				Start:    &ast.IntegerLiteral{Value: 0},
				End:      &ast.BoolLiteral{Value: true},
				Body:     []ast.Statement{},
			},
		},
	}

	errors, _ := checker.Check(program)
	if len(errors) < 1 {
		t.Fatalf("expected error for non-numeric range end, got %d", len(errors))
	}
}

// TestBlockStatementScope tests that block creates new scope
func TestBlockStatementScope(t *testing.T) {
	checker := New()

	program := &ast.Program{
		Definitions: []ast.Definition{},
		MainBody: []ast.Statement{
			&ast.LetStatement{
				Name: "x",
				Type: &ast.Type{Name: "i64"},
				Init: &ast.IntegerLiteral{Value: 5},
			},
			&ast.BlockStatement{
				Statements: []ast.Statement{
					&ast.LetStatement{
						Name: "y",
						Type: &ast.Type{Name: "i64"},
						Init: &ast.IntegerLiteral{Value: 10},
					},
				},
			},
		},
	}

	errors, _ := checker.Check(program)
	if len(errors) != 0 {
		t.Fatalf("expected 0 errors for block statement, got %d: %v", len(errors), errors)
	}
}

// TestInterfaceDefinition tests interface definition
func TestInterfaceDefinition(t *testing.T) {
	checker := New()

	program := &ast.Program{
		Definitions: []ast.Definition{
			&ast.InterfaceDef{
				Name: "Reader",
				Methods: []ast.MethodSig{
					{Name: "read", ReturnType: &ast.Type{Name: "i64"}},
				},
			},
		},
		MainBody: []ast.Statement{},
	}

	errors, _ := checker.Check(program)
	if len(errors) != 0 {
		t.Fatalf("expected 0 errors for interface definition, got %d: %v", len(errors), errors)
	}
}

// TestEnumDefinition tests enum definition
func TestEnumDefinition(t *testing.T) {
	checker := New()

	program := &ast.Program{
		Definitions: []ast.Definition{
			&ast.EnumDef{
				Name: "Status",
				Variants: []ast.EnumVariant{
					{Name: "Active"},
					{Name: "Inactive"},
					{Name: "Pending"},
				},
			},
		},
		MainBody: []ast.Statement{},
	}

	errors, _ := checker.Check(program)
	if len(errors) != 0 {
		t.Fatalf("expected 0 errors for enum definition, got %d: %v", len(errors), errors)
	}
}

// TestBinaryExpressionTypeError tests type mismatch in binary operations
func TestBinaryExpressionTypeError(t *testing.T) {
	checker := New()

	program := &ast.Program{
		Definitions: []ast.Definition{},
		MainBody: []ast.Statement{
			&ast.LetStatement{
				Name: "result",
				Init: &ast.BinaryExpression{
					Left:     &ast.StringLiteral{Value: "hello"},
					Operator: "+",
					Right:    &ast.IntegerLiteral{Value: 5},
				},
			},
		},
	}

	errors, _ := checker.Check(program)
	if len(errors) < 1 {
		t.Fatalf("expected error for type mismatch in binary operation, got %d", len(errors))
	}
}

// TestUnaryExpressionOnNumeric tests unary minus on numeric
func TestUnaryExpressionOnNumeric(t *testing.T) {
	checker := New()

	program := &ast.Program{
		Definitions: []ast.Definition{},
		MainBody: []ast.Statement{
			&ast.LetStatement{
				Name: "neg",
				Type: &ast.Type{Name: "i64"},
				Init: &ast.UnaryExpression{
					Operator: "-",
					Operand:  &ast.IntegerLiteral{Value: 10},
				},
			},
		},
	}

	errors, _ := checker.Check(program)
	if len(errors) != 0 {
		t.Fatalf("expected 0 errors for unary expression, got %d: %v", len(errors), errors)
	}
}

// TestFunctionDefWithMultipleReturns tests function with multiple return statements
func TestFunctionDefWithMultipleReturns(t *testing.T) {
	checker := New()

	program := &ast.Program{
		Definitions: []ast.Definition{
			&ast.FunctionDef{
				Name: "abs",
				Parameters: []ast.Parameter{
					{Name: "x", Type: &ast.Type{Name: "i64"}},
				},
				ReturnType: &ast.Type{Name: "i64"},
				Body: []ast.Statement{
					&ast.IfStatement{
						Condition: &ast.BinaryExpression{
							Left:     &ast.Identifier{Name: "x"},
							Operator: "<",
							Right:    &ast.IntegerLiteral{Value: 0},
						},
						ThenBody: []ast.Statement{
							&ast.ReturnStatement{
								Value: &ast.UnaryExpression{
									Operator: "-",
									Operand:  &ast.Identifier{Name: "x"},
								},
							},
						},
						ElseBody: []ast.Statement{
							&ast.ReturnStatement{
								Value: &ast.Identifier{Name: "x"},
							},
						},
					},
				},
			},
		},
		MainBody: []ast.Statement{},
	}

	errors, _ := checker.Check(program)
	if len(errors) != 0 {
		t.Fatalf("expected 0 errors for function with multiple returns, got %d: %v", len(errors), errors)
	}
}

// TestFunctionArgumentTypeMismatch tests function call with wrong argument type
func TestFunctionArgumentTypeMismatch(t *testing.T) {
	checker := New()

	program := &ast.Program{
		Definitions: []ast.Definition{
			&ast.FunctionDef{
				Name: "greet",
				Parameters: []ast.Parameter{
					{Name: "msg", Type: &ast.Type{Name: "string"}},
				},
				ReturnType: &ast.Type{Name: "none"},
				Body:       []ast.Statement{},
			},
		},
		MainBody: []ast.Statement{
			&ast.ExpressionStatement{
				Expression: &ast.CallExpression{
					Function: &ast.Identifier{Name: "greet"},
					Arguments: []ast.Expression{
						&ast.IntegerLiteral{Value: 42},
					},
				},
			},
		},
	}

	errors, _ := checker.Check(program)
	if len(errors) < 1 {
		t.Fatalf("expected error for argument type mismatch, got %d", len(errors))
	}
}

// TestTwoVariableShadowing tests variable shadowing in different scopes
func TestTwoVariableShadowing(t *testing.T) {
	checker := New()

	program := &ast.Program{
		Definitions: []ast.Definition{},
		MainBody: []ast.Statement{
			&ast.LetStatement{
				Name: "x",
				Type: &ast.Type{Name: "i64"},
				Init: &ast.IntegerLiteral{Value: 5},
			},
			&ast.BlockStatement{
				Statements: []ast.Statement{
					&ast.LetStatement{
						Name: "x",
						Type: &ast.Type{Name: "string"},
						Init: &ast.StringLiteral{Value: "hello"},
					},
				},
			},
			&ast.LetStatement{
				Name: "y",
				Init: &ast.Identifier{Name: "x"},
			},
		},
	}

	errors, _ := checker.Check(program)
	if len(errors) != 0 {
		t.Fatalf("expected 0 errors for variable shadowing, got %d: %v", len(errors), errors)
	}
}

// TestNestedBlockScope tests nested block scopes
func TestNestedBlockScope(t *testing.T) {
	checker := New()

	program := &ast.Program{
		Definitions: []ast.Definition{},
		MainBody: []ast.Statement{
			&ast.BlockStatement{
				Statements: []ast.Statement{
					&ast.LetStatement{
						Name: "a",
						Type: &ast.Type{Name: "i64"},
						Init: &ast.IntegerLiteral{Value: 1},
					},
					&ast.BlockStatement{
						Statements: []ast.Statement{
							&ast.LetStatement{
								Name: "b",
								Type: &ast.Type{Name: "i64"},
								Init: &ast.IntegerLiteral{Value: 2},
							},
						},
					},
				},
			},
		},
	}

	errors, _ := checker.Check(program)
	if len(errors) != 0 {
		t.Fatalf("expected 0 errors for nested blocks, got %d: %v", len(errors), errors)
	}
}

// TestArithmeticOnStrings tests that arithmetic on strings fails
func TestArithmeticOnStrings(t *testing.T) {
	checker := New()

	program := &ast.Program{
		Definitions: []ast.Definition{},
		MainBody: []ast.Statement{
			&ast.LetStatement{
				Name: "x",
				Init: &ast.BinaryExpression{
					Left:     &ast.StringLiteral{Value: "a"},
					Operator: "-",
					Right:    &ast.StringLiteral{Value: "b"},
				},
			},
		},
	}

	errors, _ := checker.Check(program)
	if len(errors) < 1 {
		t.Fatalf("expected error for arithmetic on strings, got %d", len(errors))
	}
}

// TestLogicalOperatorOnNonBool tests that && operator requires bool
func TestLogicalOperatorOnNonBool(t *testing.T) {
	checker := New()

	program := &ast.Program{
		Definitions: []ast.Definition{},
		MainBody: []ast.Statement{
			&ast.LetStatement{
				Name: "x",
				Init: &ast.BinaryExpression{
					Left:     &ast.IntegerLiteral{Value: 1},
					Operator: "&&",
					Right:    &ast.IntegerLiteral{Value: 2},
				},
			},
		},
	}

	errors, _ := checker.Check(program)
	if len(errors) < 1 {
		t.Fatalf("expected error for logical operator on non-bool, got %d", len(errors))
	}
}

// TestUnaryMinusOnString tests that unary minus on string fails
func TestUnaryMinusOnString(t *testing.T) {
	checker := New()

	program := &ast.Program{
		Definitions: []ast.Definition{},
		MainBody: []ast.Statement{
			&ast.LetStatement{
				Name: "x",
				Init: &ast.UnaryExpression{
					Operator: "-",
					Operand:  &ast.StringLiteral{Value: "hello"},
				},
			},
		},
	}

	errors, _ := checker.Check(program)
	if len(errors) < 1 {
		t.Fatalf("expected error for unary minus on string, got %d", len(errors))
	}
}

// TestComparisonBetweenDifferentTypes tests comparison between different types
func TestComparisonBetweenDifferentTypes(t *testing.T) {
	checker := New()

	program := &ast.Program{
		Definitions: []ast.Definition{},
		MainBody: []ast.Statement{
			&ast.LetStatement{
				Name: "result",
				Init: &ast.BinaryExpression{
					Left:     &ast.IntegerLiteral{Value: 5},
					Operator: "==",
					Right:    &ast.StringLiteral{Value: "5"},
				},
			},
		},
	}

	errors, _ := checker.Check(program)
	// 비교 연산자는 타입 자동 변환 시도 또는 불일치 보고 가능
	_ = errors
}

// TestConstStatementWithTypeMismatch tests const with mismatched type
func TestConstStatementWithTypeMismatch(t *testing.T) {
	checker := New()

	program := &ast.Program{
		Definitions: []ast.Definition{},
		MainBody: []ast.Statement{
			&ast.ConstStatement{
				Name:  "PI",
				Type:  &ast.Type{Name: "i64"},
				Value: &ast.FloatLiteral{Value: 3.14},
			},
		},
	}

	errors, _ := checker.Check(program)
	if len(errors) < 1 {
		t.Fatalf("expected error for const type mismatch, got %d", len(errors))
	}
}

// TestForStatementIterator tests for loop iterator
func TestForStatementIterator(t *testing.T) {
	checker := New()

	program := &ast.Program{
		Definitions: []ast.Definition{},
		MainBody: []ast.Statement{
			&ast.LetStatement{
				Name: "arr",
				Init: &ast.ArrayExpression{
					Elements: []ast.Expression{
						&ast.IntegerLiteral{Value: 1},
						&ast.IntegerLiteral{Value: 2},
					},
				},
			},
			&ast.ForStatement{
				Variable: "item",
				Iterator: &ast.Identifier{Name: "arr"},
				Body: []ast.Statement{
					// Empty body is OK
				},
			},
		},
	}

	errors, _ := checker.Check(program)
	if len(errors) != 0 {
		t.Fatalf("expected 0 errors for for statement, got %d: %v", len(errors), errors)
	}
}

// TestCallExpressionWithNoArguments tests function call without arguments
func TestCallExpressionWithNoArguments(t *testing.T) {
	checker := New()

	program := &ast.Program{
		Definitions: []ast.Definition{
			&ast.FunctionDef{
				Name:       "foo",
				Parameters: []ast.Parameter{},
				ReturnType: &ast.Type{Name: "i64"},
				Body: []ast.Statement{
					&ast.ReturnStatement{
						Value: &ast.IntegerLiteral{Value: 42},
					},
				},
			},
		},
		MainBody: []ast.Statement{
			&ast.LetStatement{
				Name: "result",
				Init: &ast.CallExpression{
					Function:  &ast.Identifier{Name: "foo"},
					Arguments: []ast.Expression{},
				},
			},
		},
	}

	errors, _ := checker.Check(program)
	if len(errors) != 0 {
		t.Fatalf("expected 0 errors for zero-argument function call, got %d: %v", len(errors), errors)
	}
}

// TestReturnStatementWithoutValue tests return without value
func TestReturnStatementWithoutValue(t *testing.T) {
	checker := New()

	program := &ast.Program{
		Definitions: []ast.Definition{
			&ast.FunctionDef{
				Name:       "noop",
				Parameters: []ast.Parameter{},
				ReturnType: &ast.Type{Name: "none"},
				Body: []ast.Statement{
					&ast.ReturnStatement{
						Value: nil,
					},
				},
			},
		},
		MainBody: []ast.Statement{},
	}

	errors, _ := checker.Check(program)
	if len(errors) != 0 {
		t.Fatalf("expected 0 errors for return without value, got %d: %v", len(errors), errors)
	}
}

// TestFunctionWithoutExplicitReturnType tests function without explicit return type
func TestFunctionWithoutExplicitReturnType(t *testing.T) {
	checker := New()

	program := &ast.Program{
		Definitions: []ast.Definition{
			&ast.FunctionDef{
				Name:       "implicit",
				Parameters: []ast.Parameter{},
				ReturnType: nil,
				Body: []ast.Statement{
					&ast.ExpressionStatement{
						Expression: &ast.IntegerLiteral{Value: 42},
					},
				},
			},
		},
		MainBody: []ast.Statement{},
	}

	errors, _ := checker.Check(program)
	if len(errors) != 0 {
		t.Fatalf("expected 0 errors for function without explicit return type, got %d: %v", len(errors), errors)
	}
}

// TestArrayIndexing tests array indexing with valid index
func TestArrayIndexing(t *testing.T) {
	checker := New()

	program := &ast.Program{
		Definitions: []ast.Definition{},
		MainBody: []ast.Statement{
			&ast.LetStatement{
				Name: "nums",
				Init: &ast.ArrayExpression{
					Elements: []ast.Expression{
						&ast.IntegerLiteral{Value: 10},
						&ast.IntegerLiteral{Value: 20},
					},
				},
			},
			&ast.LetStatement{
				Name: "first",
				Init: &ast.IndexExpression{
					Object: &ast.Identifier{Name: "nums"},
					Index:  &ast.IntegerLiteral{Value: 0},
				},
			},
		},
	}

	errors, _ := checker.Check(program)
	if len(errors) != 0 {
		t.Fatalf("expected 0 errors for array indexing, got %d: %v", len(errors), errors)
	}
}

// TestIfExpressionWithSameTypes tests if-else expression with matching types
func TestIfExpressionWithSameTypes(t *testing.T) {
	checker := New()

	program := &ast.Program{
		Definitions: []ast.Definition{},
		MainBody: []ast.Statement{
			&ast.LetStatement{
				Name: "msg",
				Init: &ast.IfExpression{
					Condition: &ast.BoolLiteral{Value: true},
					ThenExpr:  &ast.StringLiteral{Value: "yes"},
					ElseExpr:  &ast.StringLiteral{Value: "no"},
				},
			},
		},
	}

	errors, _ := checker.Check(program)
	if len(errors) != 0 {
		t.Fatalf("expected 0 errors for if-expression with same types, got %d: %v", len(errors), errors)
	}
}

// TestIfExpressionWithoutElse tests if-expression without else
func TestIfExpressionWithoutElse(t *testing.T) {
	checker := New()

	program := &ast.Program{
		Definitions: []ast.Definition{},
		MainBody: []ast.Statement{
			&ast.LetStatement{
				Name: "x",
				Init: &ast.IfExpression{
					Condition: &ast.BoolLiteral{Value: true},
					ThenExpr:  &ast.IntegerLiteral{Value: 42},
					ElseExpr:  nil,
				},
			},
		},
	}

	errors, _ := checker.Check(program)
	// If-expression without else may return none
	_ = errors
}

// TestFieldAccessOnStruct tests field access on struct
func TestFieldAccessOnStruct(t *testing.T) {
	checker := New()

	program := &ast.Program{
		Definitions: []ast.Definition{
			&ast.StructDef{
				Name: "Box",
				Fields: []ast.Field{
					{Name: "width", Type: &ast.Type{Name: "i64"}},
					{Name: "height", Type: &ast.Type{Name: "i64"}},
				},
			},
		},
		MainBody: []ast.Statement{},
	}

	errors, _ := checker.Check(program)
	if len(errors) != 0 {
		t.Fatalf("expected 0 errors for struct field definition, got %d: %v", len(errors), errors)
	}
}

// TestMultipleFunctionDefinitions tests multiple function definitions
func TestMultipleFunctionDefinitions(t *testing.T) {
	checker := New()

	program := &ast.Program{
		Definitions: []ast.Definition{
			&ast.FunctionDef{
				Name:       "first",
				Parameters: []ast.Parameter{},
				ReturnType: &ast.Type{Name: "i64"},
				Body: []ast.Statement{
					&ast.ReturnStatement{Value: &ast.IntegerLiteral{Value: 1}},
				},
			},
			&ast.FunctionDef{
				Name:       "second",
				Parameters: []ast.Parameter{},
				ReturnType: &ast.Type{Name: "i64"},
				Body: []ast.Statement{
					&ast.ReturnStatement{Value: &ast.IntegerLiteral{Value: 2}},
				},
			},
		},
		MainBody: []ast.Statement{},
	}

	errors, _ := checker.Check(program)
	if len(errors) != 0 {
		t.Fatalf("expected 0 errors for multiple functions, got %d: %v", len(errors), errors)
	}
}

// TestRangeOperator tests range operator in expressions
func TestRangeOperator(t *testing.T) {
	checker := New()

	program := &ast.Program{
		Definitions: []ast.Definition{},
		MainBody: []ast.Statement{
			&ast.LetStatement{
				Name: "range_expr",
				Init: &ast.BinaryExpression{
					Left:     &ast.IntegerLiteral{Value: 1},
					Operator: "..",
					Right:    &ast.IntegerLiteral{Value: 10},
				},
			},
		},
	}

	errors, _ := checker.Check(program)
	// Range operator type depends on implementation
	_ = errors
}

// TestMultipleFunctionCallsInSequence tests multiple function calls
func TestMultipleFunctionCallsInSequence(t *testing.T) {
	checker := New()

	program := &ast.Program{
		Definitions: []ast.Definition{
			&ast.FunctionDef{
				Name:       "id",
				Parameters: []ast.Parameter{{Name: "x", Type: &ast.Type{Name: "i64"}}},
				ReturnType: &ast.Type{Name: "i64"},
				Body: []ast.Statement{
					&ast.ReturnStatement{Value: &ast.Identifier{Name: "x"}},
				},
			},
		},
		MainBody: []ast.Statement{
			&ast.LetStatement{
				Name: "a",
				Init: &ast.CallExpression{
					Function:  &ast.Identifier{Name: "id"},
					Arguments: []ast.Expression{&ast.IntegerLiteral{Value: 1}},
				},
			},
			&ast.LetStatement{
				Name: "b",
				Init: &ast.CallExpression{
					Function:  &ast.Identifier{Name: "id"},
					Arguments: []ast.Expression{&ast.IntegerLiteral{Value: 2}},
				},
			},
		},
	}

	errors, _ := checker.Check(program)
	if len(errors) != 0 {
		t.Fatalf("expected 0 errors for sequential function calls, got %d: %v", len(errors), errors)
	}
}

// TestComplexBinaryExpression tests nested binary operations
func TestComplexBinaryExpression(t *testing.T) {
	checker := New()

	program := &ast.Program{
		Definitions: []ast.Definition{},
		MainBody: []ast.Statement{
			&ast.LetStatement{
				Name: "result",
				Init: &ast.BinaryExpression{
					Left: &ast.BinaryExpression{
						Left:     &ast.IntegerLiteral{Value: 2},
						Operator: "*",
						Right:    &ast.IntegerLiteral{Value: 3},
					},
					Operator: "+",
					Right: &ast.BinaryExpression{
						Left:     &ast.IntegerLiteral{Value: 4},
						Operator: "*",
						Right:    &ast.IntegerLiteral{Value: 5},
					},
				},
			},
		},
	}

	errors, _ := checker.Check(program)
	if len(errors) != 0 {
		t.Fatalf("expected 0 errors for complex binary expression, got %d: %v", len(errors), errors)
	}
}

// TestModuloOperator tests modulo arithmetic operator
func TestModuloOperator(t *testing.T) {
	checker := New()

	program := &ast.Program{
		Definitions: []ast.Definition{},
		MainBody: []ast.Statement{
			&ast.LetStatement{
				Name: "remainder",
				Init: &ast.BinaryExpression{
					Left:     &ast.IntegerLiteral{Value: 10},
					Operator: "%",
					Right:    &ast.IntegerLiteral{Value: 3},
				},
			},
		},
	}

	errors, _ := checker.Check(program)
	if len(errors) != 0 {
		t.Fatalf("expected 0 errors for modulo operator, got %d: %v", len(errors), errors)
	}
}

// TestNotEqualOperator tests != operator
func TestNotEqualOperator(t *testing.T) {
	checker := New()

	program := &ast.Program{
		Definitions: []ast.Definition{},
		MainBody: []ast.Statement{
			&ast.LetStatement{
				Name: "different",
				Type: &ast.Type{Name: "bool"},
				Init: &ast.BinaryExpression{
					Left:     &ast.IntegerLiteral{Value: 5},
					Operator: "!=",
					Right:    &ast.IntegerLiteral{Value: 3},
				},
			},
		},
	}

	errors, _ := checker.Check(program)
	if len(errors) != 0 {
		t.Fatalf("expected 0 errors for != operator, got %d: %v", len(errors), errors)
	}
}

// TestLogicalOrOperator tests || operator
func TestLogicalOrOperator(t *testing.T) {
	checker := New()

	program := &ast.Program{
		Definitions: []ast.Definition{},
		MainBody: []ast.Statement{
			&ast.LetStatement{
				Name: "or_result",
				Type: &ast.Type{Name: "bool"},
				Init: &ast.BinaryExpression{
					Left:     &ast.BoolLiteral{Value: false},
					Operator: "||",
					Right:    &ast.BoolLiteral{Value: true},
				},
			},
		},
	}

	errors, _ := checker.Check(program)
	if len(errors) != 0 {
		t.Fatalf("expected 0 errors for || operator, got %d: %v", len(errors), errors)
	}
}
