package parser

import (
	"testing"

	"github.com/user/freelang-evolving-compiler/internal/ast"
)

func TestParseLetDecl(t *testing.T) {
	tests := []struct {
		input   string
		varName string
		wantErr bool
	}{
		{"let x = 10", "x", false},
		{"let y = 20", "y", false},
		{"let abc = 100 + 200", "abc", false},
	}

	for _, tt := range tests {
		p := New(tt.input)
		prog, err := p.ParseProgram()
		if (err != nil) != tt.wantErr {
			t.Errorf("Input %q: error = %v, wantErr %v", tt.input, err, tt.wantErr)
			continue
		}
		if err != nil {
			continue
		}
		if len(prog.Children) != 1 {
			t.Errorf("Input %q: expected 1 statement, got %d", tt.input, len(prog.Children))
			continue
		}
		stmt := prog.Children[0]
		if stmt.Kind != ast.NodeLetDecl {
			t.Errorf("Input %q: expected NodeLetDecl, got %v", tt.input, stmt.Kind)
			continue
		}
		if len(stmt.Children) < 1 || stmt.Children[0].Value != tt.varName {
			t.Errorf("Input %q: expected var name %q, got %q", tt.input, tt.varName,
				func() string {
					if len(stmt.Children) > 0 {
						return stmt.Children[0].Value
					}
					return ""
				}())
		}
	}
}

func TestParseFnDecl(t *testing.T) {
	input := "fn add(a, b) { return a + b }"
	p := New(input)
	prog, err := p.ParseProgram()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(prog.Children) != 1 {
		t.Fatalf("Expected 1 statement, got %d", len(prog.Children))
	}
	stmt := prog.Children[0]
	if stmt.Kind != ast.NodeFnDecl {
		t.Errorf("Expected NodeFnDecl, got %v", stmt.Kind)
	}
	if stmt.Value != "add" {
		t.Errorf("Expected function name 'add', got %q", stmt.Value)
	}
}

func TestParseIfStmt(t *testing.T) {
	input := "if x > 5 { let y = 10 }"
	p := New(input)
	prog, err := p.ParseProgram()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(prog.Children) != 1 {
		t.Fatalf("Expected 1 statement, got %d", len(prog.Children))
	}
	stmt := prog.Children[0]
	if stmt.Kind != ast.NodeIfStmt {
		t.Errorf("Expected NodeIfStmt, got %v", stmt.Kind)
	}
}

func TestParseForStmt(t *testing.T) {
	input := "for i in 0..10 { let sum = sum + i }"
	p := New(input)
	prog, err := p.ParseProgram()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(prog.Children) != 1 {
		t.Fatalf("Expected 1 statement, got %d", len(prog.Children))
	}
	stmt := prog.Children[0]
	if stmt.Kind != ast.NodeForStmt {
		t.Errorf("Expected NodeForStmt, got %v", stmt.Kind)
	}
}

func TestParseMultipleStatements(t *testing.T) {
	input := `
		let x = 10
		let y = 20
		let z = x + y
	`
	p := New(input)
	prog, err := p.ParseProgram()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(prog.Children) != 3 {
		t.Errorf("Expected 3 statements, got %d", len(prog.Children))
	}
	for i, stmt := range prog.Children {
		if stmt.Kind != ast.NodeLetDecl {
			t.Errorf("Statement %d: expected NodeLetDecl, got %v", i, stmt.Kind)
		}
	}
}

func TestParseBinaryExpressions(t *testing.T) {
	tests := []struct {
		input string
		want  []ast.TokenType
	}{
		{"2 + 3", []ast.TokenType{ast.TokenPlus}},
		{"a * b", []ast.TokenType{ast.TokenStar}},
		{"x > 5", []ast.TokenType{ast.TokenGt}},
		{"a == b", []ast.TokenType{ast.TokenEq}},
	}

	for _, tt := range tests {
		p := New(tt.input)
		expr, err := p.parseExpression(0)
		if err != nil {
			t.Errorf("Input %q: unexpected error %v", tt.input, err)
			continue
		}
		if expr.Kind != ast.NodeBinaryExpr {
			t.Errorf("Input %q: expected binary expr, got %v", tt.input, expr.Kind)
		}
	}
}

func TestParseNesting(t *testing.T) {
	input := `
		if x > 5 {
			for i in 0..10 {
				let sum = sum + i
			}
		}
	`
	p := New(input)
	prog, err := p.ParseProgram()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(prog.Children) != 1 {
		t.Fatalf("Expected 1 statement, got %d", len(prog.Children))
	}
}
