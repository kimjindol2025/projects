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
