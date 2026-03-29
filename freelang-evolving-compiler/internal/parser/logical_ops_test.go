package parser

import (
	"testing"

	"github.com/user/freelang-evolving-compiler/internal/ast"
)

func TestLogicalNot(t *testing.T) {
	code := "let x = !true;"
	p := New(code)
	prog, err := p.ParseProgram()
	if err != nil {
		t.Fatalf("parsing failed: %v", err)
	}

	if len(prog.Children) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(prog.Children))
	}

	letNode := prog.Children[0]
	if letNode.Kind != ast.NodeLetDecl {
		t.Fatalf("expected NodeLetDecl, got %v", letNode.Kind)
	}

	// Check the expression (should be UnaryExpr with NOT)
	if len(letNode.Children) < 2 {
		t.Fatalf("expected at least 2 children in LetDecl")
	}

	exprNode := letNode.Children[1]
	if exprNode.Kind != ast.NodeUnaryExpr {
		t.Fatalf("expected NodeUnaryExpr, got %v", exprNode.Kind)
	}

	if exprNode.Value != "!" {
		t.Fatalf("expected !, got %s", exprNode.Value)
	}

	t.Log("✓ NOT operator parsed correctly")
}

func TestLogicalAnd(t *testing.T) {
	code := "let a = true && false;"
	p := New(code)
	prog, err := p.ParseProgram()
	if err != nil {
		t.Fatalf("parsing failed: %v", err)
	}

	letNode := prog.Children[0]
	exprNode := letNode.Children[1]

	if exprNode.Kind != ast.NodeBinaryExpr {
		t.Fatalf("expected NodeBinaryExpr, got %v", exprNode.Kind)
	}

	if exprNode.Value != "&&" {
		t.Fatalf("expected &&, got %s", exprNode.Value)
	}

	if len(exprNode.Children) != 2 {
		t.Fatalf("expected 2 operands, got %d", len(exprNode.Children))
	}

	t.Log("✓ AND operator parsed correctly")
}

func TestLogicalOr(t *testing.T) {
	code := "let b = true || false;"
	p := New(code)
	prog, err := p.ParseProgram()
	if err != nil {
		t.Fatalf("parsing failed: %v", err)
	}

	letNode := prog.Children[0]
	exprNode := letNode.Children[1]

	if exprNode.Kind != ast.NodeBinaryExpr {
		t.Fatalf("expected NodeBinaryExpr, got %v", exprNode.Kind)
	}

	if exprNode.Value != "||" {
		t.Fatalf("expected ||, got %s", exprNode.Value)
	}

	t.Log("✓ OR operator parsed correctly")
}

func TestComplexLogicalExpr(t *testing.T) {
	code := "let c = (true && false) || !true;"
	p := New(code)
	prog, err := p.ParseProgram()
	if err != nil {
		t.Fatalf("parsing failed: %v", err)
	}

	letNode := prog.Children[0]
	exprNode := letNode.Children[1]

	// Should be: (true && false) || !true
	// Top level: OR
	if exprNode.Kind != ast.NodeBinaryExpr || exprNode.Value != "||" {
		t.Fatalf("expected top-level OR")
	}

	t.Log("✓ Complex logical expression parsed correctly")
}
