package parser

import (
	"testing"

	"github.com/user/freelang-evolving-compiler/internal/ast"
	"github.com/user/freelang-evolving-compiler/internal/codegen"
	"github.com/user/freelang-evolving-compiler/internal/ir"
)

// TestStructSimple tests parsing of a simple struct with integer fields
func TestStructSimple(t *testing.T) {
	input := `struct Point { x: int; y: int }`
	p := New(input)
	prog, err := p.ParseProgram()

	if err != nil {
		t.Fatalf("parsing failed: %v", err)
	}

	if prog == nil || len(prog.Children) == 0 {
		t.Fatal("expected struct declaration in program")
	}

	structDecl := prog.Children[0]
	if structDecl.Kind != ast.NodeStructDecl {
		t.Errorf("expected NodeStructDecl, got %v", structDecl.Kind)
	}

	if structDecl.Value != "Point" {
		t.Errorf("expected struct name 'Point', got '%s'", structDecl.Value)
	}

	if len(structDecl.Children) != 2 {
		t.Errorf("expected 2 fields, got %d", len(structDecl.Children))
	}

	// Check field names
	for i, expectedName := range []string{"x", "y"} {
		if structDecl.Children[i].Value != expectedName {
			t.Errorf("field %d: expected '%s', got '%s'", i, expectedName, structDecl.Children[i].Value)
		}
	}
}

// TestStructFieldStr tests parsing of struct with string fields
func TestStructFieldStr(t *testing.T) {
	input := `struct Person { name: string; age: int }`
	p := New(input)
	prog, err := p.ParseProgram()

	if err != nil {
		t.Fatalf("parsing failed: %v", err)
	}

	if prog == nil || len(prog.Children) == 0 {
		t.Fatal("expected struct declaration in program")
	}

	structDecl := prog.Children[0]
	if structDecl.Kind != ast.NodeStructDecl {
		t.Errorf("expected NodeStructDecl, got %v", structDecl.Kind)
	}

	if structDecl.Value != "Person" {
		t.Errorf("expected struct name 'Person', got '%s'", structDecl.Value)
	}

	if len(structDecl.Children) != 2 {
		t.Errorf("expected 2 fields, got %d", len(structDecl.Children))
	}

	// Check field names
	expectedFields := map[int]string{0: "name", 1: "age"}
	for i, expectedName := range expectedFields {
		if structDecl.Children[i].Value != expectedName {
			t.Errorf("field %d: expected '%s', got '%s'", i, expectedName, structDecl.Children[i].Value)
		}
	}
}

// TestStructNodeKind validates that NodeStructDecl is correct
func TestStructNodeKind(t *testing.T) {
	input := `struct Rectangle { width: int; height: int }`
	p := New(input)
	prog, err := p.ParseProgram()

	if err != nil {
		t.Fatalf("parsing failed: %v", err)
	}

	structDecl := prog.Children[0]
	if structDecl.Kind != ast.NodeStructDecl {
		t.Errorf("expected NodeStructDecl (%d), got %d", ast.NodeStructDecl, structDecl.Kind)
	}

	// Verify children are FieldDecl nodes
	for i, field := range structDecl.Children {
		if field.Kind != ast.NodeFieldDecl {
			t.Errorf("field %d: expected NodeFieldDecl, got %v", i, field.Kind)
		}
	}
}

// TestStructIRGen tests IR generation from parsed struct
func TestStructIRGen(t *testing.T) {
	input := `struct Data { value: int; flag: int }`
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

	// Check that OpStructDef was emitted in Main
	if len(irProg.Main) == 0 {
		t.Fatal("expected OpStructDef instruction in Main")
	}

	found := false
	for _, instr := range irProg.Main {
		if instr.Op == ir.OpStructDef && instr.Fn == "Data" {
			found = true
			// Size should be 2 fields * 8 bytes = 16
			if instr.Src1.ImmVal != 16 {
				t.Errorf("expected struct size 16, got %d", instr.Src1.ImmVal)
			}
		}
	}

	if !found {
		t.Fatal("expected OpStructDef instruction for struct 'Data'")
	}
}

// TestStructCodegen tests code generation from struct IR
func TestStructCodegen(t *testing.T) {
	input := `struct Point { x: int; y: int }`
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

	// Check for struct definition comment in output
	if !contains(result.Code, "; STRUCT Point size=16") {
		t.Errorf("expected struct definition in generated code\nGot:\n%s", result.Code)
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	for i := 0; i < len(s)-len(substr)+1; i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
