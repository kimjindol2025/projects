package codegen

import (
	"strings"
	"testing"

	"fv2-lang/internal/ast"
)

func TestBasicCGeneration(t *testing.T) {
	gen := New()

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

	code, err := gen.Generate(program)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(code, "long long x = 42;") {
		t.Errorf("expected 'long long x = 42;', got:\n%s", code)
	}

	if !strings.Contains(code, "#include <stdio.h>") {
		t.Errorf("missing stdio.h include")
	}

	if !strings.Contains(code, "int main()") {
		t.Errorf("missing main function")
	}
}

func TestFunctionGeneration(t *testing.T) {
	gen := New()

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

	code, err := gen.Generate(program)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(code, "long long add(long long x, long long y)") {
		t.Errorf("function signature not correct")
	}

	if !strings.Contains(code, "return (x + y);") {
		t.Errorf("function body not correct")
	}
}

func TestBinaryExpression(t *testing.T) {
	gen := New()

	program := &ast.Program{
		Definitions: []ast.Definition{},
		MainBody: []ast.Statement{
			&ast.LetStatement{
				Name: "result",
				Init: &ast.BinaryExpression{
					Left:     &ast.IntegerLiteral{Value: 5},
					Operator: "+",
					Right:    &ast.IntegerLiteral{Value: 3},
				},
			},
		},
	}

	code, err := gen.Generate(program)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(code, "(5 + 3)") {
		t.Errorf("binary expression not correct")
	}
}

func TestArrayGeneration(t *testing.T) {
	gen := New()

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

	code, err := gen.Generate(program)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(code, "{1, 2, 3}") {
		t.Errorf("array not correct")
	}
}

func TestIfStatementGeneration(t *testing.T) {
	gen := New()

	program := &ast.Program{
		Definitions: []ast.Definition{},
		MainBody: []ast.Statement{
			&ast.IfStatement{
				Condition: &ast.BoolLiteral{Value: true},
				ThenBody: []ast.Statement{
					&ast.ExpressionStatement{
						Expression: &ast.IntegerLiteral{Value: 42},
					},
				},
				ElseBody: []ast.Statement{},
			},
		},
	}

	code, err := gen.Generate(program)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(code, "if (true)") {
		t.Errorf("if statement not correct")
	}
}

func TestForRangeGeneration(t *testing.T) {
	gen := New()

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

	code, err := gen.Generate(program)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(code, "for (long long i = 0; i < 10; i++)") {
		t.Errorf("for-range not correct")
	}
}

func TestTypeGeneration(t *testing.T) {
	gen := New()

	tests := []struct {
		name     string
		astType  *ast.Type
		expected string
	}{
		{"i64", &ast.Type{Name: "i64"}, "long long"},
		{"f64", &ast.Type{Name: "f64"}, "double"},
		{"string", &ast.Type{Name: "string"}, "char*"},
		{"bool", &ast.Type{Name: "bool"}, "bool"},
		{"none", &ast.Type{Name: "none"}, "void"},
	}

	for _, test := range tests {
		result := gen.generateType(test.astType)
		if result != test.expected {
			t.Errorf("%s: expected %s, got %s", test.name, test.expected, result)
		}
	}
}

func TestStringLiteral(t *testing.T) {
	gen := New()

	program := &ast.Program{
		Definitions: []ast.Definition{},
		MainBody: []ast.Statement{
			&ast.LetStatement{
				Name: "greeting",
				Type: &ast.Type{Name: "string"},
				Init: &ast.StringLiteral{Value: "Hello, World!"},
			},
		},
	}

	code, err := gen.Generate(program)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(code, "\"Hello, World!\"") {
		t.Errorf("string literal not correct")
	}
}

func TestFunctionCall(t *testing.T) {
	gen := New()

	program := &ast.Program{
		Definitions: []ast.Definition{
			&ast.FunctionDef{
				Name:       "test",
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
					Function: &ast.Identifier{Name: "test"},
					Arguments: []ast.Expression{},
				},
			},
		},
	}

	code, err := gen.Generate(program)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(code, "test()") {
		t.Errorf("function call not correct")
	}
}

func TestStructGeneration(t *testing.T) {
	gen := New()

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

	code, err := gen.Generate(program)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(code, "struct Point") {
		t.Errorf("struct definition not correct")
	}

	if !strings.Contains(code, "long long x;") {
		t.Errorf("struct field x not correct")
	}

	if !strings.Contains(code, "long long y;") {
		t.Errorf("struct field y not correct")
	}
}

func TestConstStatement(t *testing.T) {
	gen := New()

	program := &ast.Program{
		Definitions: []ast.Definition{},
		MainBody: []ast.Statement{
			&ast.ConstStatement{
				Name:  "PI",
				Type:  &ast.Type{Name: "f64"},
				Value: &ast.FloatLiteral{Value: 3.14},
			},
		},
	}

	code, err := gen.Generate(program)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(code, "const double PI = 3.14;") {
		t.Errorf("const statement not correct")
	}
}

func TestComplexProgram(t *testing.T) {
	gen := New()

	program := &ast.Program{
		Definitions: []ast.Definition{
			&ast.FunctionDef{
				Name: "multiply",
				Parameters: []ast.Parameter{
					{Name: "a", Type: &ast.Type{Name: "i64"}},
					{Name: "b", Type: &ast.Type{Name: "i64"}},
				},
				ReturnType: &ast.Type{Name: "i64"},
				Body: []ast.Statement{
					&ast.ReturnStatement{
						Value: &ast.BinaryExpression{
							Left:     &ast.Identifier{Name: "a"},
							Operator: "*",
							Right:    &ast.Identifier{Name: "b"},
						},
					},
				},
			},
		},
		MainBody: []ast.Statement{
			&ast.LetStatement{
				Name: "x",
				Type: &ast.Type{Name: "i64"},
				Init: &ast.IntegerLiteral{Value: 5},
			},
			&ast.LetStatement{
				Name: "y",
				Type: &ast.Type{Name: "i64"},
				Init: &ast.IntegerLiteral{Value: 3},
			},
			&ast.LetStatement{
				Name: "result",
				Init: &ast.CallExpression{
					Function: &ast.Identifier{Name: "multiply"},
					Arguments: []ast.Expression{
						&ast.Identifier{Name: "x"},
						&ast.Identifier{Name: "y"},
					},
				},
			},
		},
	}

	code, err := gen.Generate(program)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(code, "long long multiply") {
		t.Errorf("function not generated")
	}

	if !strings.Contains(code, "int main()") {
		t.Errorf("main not generated")
	}

	if !strings.Contains(code, "long long x = 5;") {
		t.Errorf("variable x not generated")
	}

	if !strings.Contains(code, "multiply(x, y)") {
		t.Errorf("function call not generated")
	}
}

func TestExternFnCodegen(t *testing.T) {
	gen := New()

	program := &ast.Program{
		Definitions: []ast.Definition{
			&ast.ExternDef{
				Name: "printf",
				Parameters: []ast.Parameter{
					{Name: "fmt", Type: &ast.Type{Name: "string"}},
				},
				ReturnType: nil, // No return type (void)
			},
		},
		MainBody: []ast.Statement{},
	}

	code, err := gen.Generate(program)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(code, "extern void printf") {
		t.Errorf("expected 'extern void printf', got:\n%s", code)
	}
}

func TestArrayDeclarationCodegen(t *testing.T) {
	gen := New()

	program := &ast.Program{
		Definitions: []ast.Definition{},
		MainBody: []ast.Statement{
			&ast.LetStatement{
				Name: "nums",
				Type: nil,
				Init: &ast.ArrayExpression{
					Elements: []ast.Expression{
						&ast.IntegerLiteral{Value: 10},
						&ast.IntegerLiteral{Value: 20},
						&ast.IntegerLiteral{Value: 30},
					},
				},
			},
		},
	}

	code, err := gen.Generate(program)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(code, "long long nums[] = {10, 20, 30}") {
		t.Errorf("expected array declaration with [] syntax, got:\n%s", code)
	}
}

func TestIndexExpressionCodegen(t *testing.T) {
	gen := New()

	program := &ast.Program{
		Definitions: []ast.Definition{},
		MainBody: []ast.Statement{
			&ast.ExpressionStatement{
				Expression: &ast.IndexExpression{
					Object: &ast.Identifier{Name: "arr"},
					Index:  &ast.IntegerLiteral{Value: 0},
				},
			},
		},
	}

	code, err := gen.Generate(program)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(code, "arr[0]") {
		t.Errorf("expected 'arr[0]', got:\n%s", code)
	}
}

func TestElseIfCodegen(t *testing.T) {
	gen := New()

	program := &ast.Program{
		Definitions: []ast.Definition{},
		MainBody: []ast.Statement{
			&ast.IfStatement{
				Condition: &ast.BinaryExpression{
					Left:     &ast.Identifier{Name: "x"},
					Operator: ">",
					Right:    &ast.IntegerLiteral{Value: 20},
				},
				ThenBody: []ast.Statement{
					&ast.ExpressionStatement{
						Expression: &ast.CallExpression{
							Function: &ast.Identifier{Name: "println"},
							Arguments: []ast.Expression{
								&ast.StringLiteral{Value: "big"},
							},
						},
					},
				},
				ElseBody: []ast.Statement{
					&ast.IfStatement{
						Condition: &ast.BinaryExpression{
							Left:     &ast.Identifier{Name: "x"},
							Operator: ">",
							Right:    &ast.IntegerLiteral{Value: 5},
						},
						ThenBody: []ast.Statement{
							&ast.ExpressionStatement{
								Expression: &ast.CallExpression{
									Function: &ast.Identifier{Name: "println"},
									Arguments: []ast.Expression{
										&ast.StringLiteral{Value: "medium"},
									},
								},
							},
						},
						ElseBody: []ast.Statement{},
					},
				},
			},
		},
	}

	code, err := gen.Generate(program)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(code, "} else {") {
		t.Errorf("expected 'else' block, got:\n%s", code)
	}

	if !strings.Contains(code, "if (") {
		t.Errorf("expected 'if' statement, got:\n%s", code)
	}
}
