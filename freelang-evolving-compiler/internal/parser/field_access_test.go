package parser

import (
	"testing"

	"github.com/user/freelang-evolving-compiler/internal/ast"
	"github.com/user/freelang-evolving-compiler/internal/codegen"
	"github.com/user/freelang-evolving-compiler/internal/ir"
)

// TestFieldAccessParse tests parsing of field access syntax
func TestFieldAccessParse(t *testing.T) {
	input := `let p = 0
let v = p.x`
	p := New(input)
	prog, err := p.ParseProgram()

	if err != nil {
		t.Fatalf("parsing failed: %v", err)
	}

	if prog == nil || len(prog.Children) < 2 {
		t.Fatal("expected at least 2 statements")
	}

	// Second statement should be: let v = p.x
	letDecl := prog.Children[1]
	if letDecl.Kind != ast.NodeLetDecl {
		t.Errorf("expected NodeLetDecl, got %v", letDecl.Kind)
	}

	if len(letDecl.Children) < 2 {
		t.Fatal("let declaration should have name and value")
	}

	// The value expression should be field access
	fieldAccess := letDecl.Children[1]
	if fieldAccess.Kind != ast.NodeFieldAccess {
		t.Errorf("expected NodeFieldAccess, got %v", fieldAccess.Kind)
	}

	if fieldAccess.Value != "x" {
		t.Errorf("expected field name 'x', got '%s'", fieldAccess.Value)
	}

	if len(fieldAccess.Children) == 0 {
		t.Fatal("field access should have object expression")
	}

	// Object should be identifier 'p'
	objExpr := fieldAccess.Children[0]
	if objExpr.Kind != ast.NodeIdent {
		t.Errorf("expected NodeIdent for object, got %v", objExpr.Kind)
	}

	if objExpr.Value != "p" {
		t.Errorf("expected object name 'p', got '%s'", objExpr.Value)
	}
}

// TestFieldAccessNodeKind validates NodeFieldAccess structure
func TestFieldAccessNodeKind(t *testing.T) {
	input := `let obj = 0
let val = obj.field`
	p := New(input)
	prog, err := p.ParseProgram()

	if err != nil {
		t.Fatalf("parsing failed: %v", err)
	}

	fieldAccess := prog.Children[1].Children[1]

	if fieldAccess.Kind != ast.NodeFieldAccess {
		t.Errorf("expected NodeFieldAccess (%d), got %d", ast.NodeFieldAccess, fieldAccess.Kind)
	}

	if fieldAccess.Value != "field" {
		t.Errorf("expected Value='field', got '%s'", fieldAccess.Value)
	}

	if len(fieldAccess.Children) != 1 {
		t.Errorf("expected 1 child (object), got %d", len(fieldAccess.Children))
	}
}

// TestFieldAccessChained tests multiple independent field accesses
func TestFieldAccessChained(t *testing.T) {
	input := `let p = 0
let x = p.x
let y = p.y`
	p := New(input)
	prog, err := p.ParseProgram()

	if err != nil {
		t.Fatalf("parsing failed: %v", err)
	}

	if len(prog.Children) < 3 {
		t.Fatal("expected at least 3 statements")
	}

	// Second statement: let x = p.x
	letX := prog.Children[1]
	fieldX := letX.Children[1]
	if fieldX.Kind != ast.NodeFieldAccess || fieldX.Value != "x" {
		t.Errorf("expected field 'x', got %v", fieldX.Value)
	}

	// Third statement: let y = p.y
	letY := prog.Children[2]
	fieldY := letY.Children[1]
	if fieldY.Kind != ast.NodeFieldAccess || fieldY.Value != "y" {
		t.Errorf("expected field 'y', got %v", fieldY.Value)
	}
}

// TestFieldAccessIRGen tests IR generation from field access
func TestFieldAccessIRGen(t *testing.T) {
	input := `struct Point { x: int; y: int }
let p = 0
let v = p.x`
	p := New(input)
	prog, err := p.ParseProgram()

	if err != nil {
		t.Fatalf("parsing failed: %v", err)
	}

	gen := ir.NewGenerator()
	irProg, err := gen.Generate(prog)

	if err != nil {
		t.Fatalf("IR generation failed: %v", err)
	}

	if irProg == nil {
		t.Fatal("expected non-nil IR program")
	}

	// Find OpFieldLoad in Main
	found := false
	for _, instr := range irProg.Main {
		if instr.Op == ir.OpFieldLoad {
			found = true
			// First field (x) should have offset 0
			if instr.Src2.ImmVal != 0 {
				t.Errorf("expected offset 0 for field 'x', got %d", instr.Src2.ImmVal)
			}
		}
	}

	if !found {
		t.Fatal("expected OpFieldLoad instruction in Main")
	}
}

// TestFieldAccessCodegen tests code generation from field access IR
func TestFieldAccessCodegen(t *testing.T) {
	input := `struct Point { x: int; y: int }
let p = 0
let v = p.x`
	p := New(input)
	prog, err := p.ParseProgram()

	if err != nil {
		t.Fatalf("parsing failed: %v", err)
	}

	gen := ir.NewGenerator()
	irProg, err := gen.Generate(prog)

	if err != nil {
		t.Fatalf("IR generation failed: %v", err)
	}

	cg := codegen.New()
	result := cg.Generate(irProg)

	if result.Code == "" {
		t.Fatal("expected non-empty generated code")
	}

	// Check for LOAD instruction with field offset
	if !contains(result.Code, "LOAD") || !contains(result.Code, "[") {
		t.Errorf("expected LOAD with offset in generated code\nGot:\n%s", result.Code)
	}
}
