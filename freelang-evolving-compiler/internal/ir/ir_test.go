package ir

import (
	"testing"

	"github.com/user/freelang-evolving-compiler/internal/parser"
)

// TestGenerateEmpty validates generation of empty program
func TestGenerateEmpty(t *testing.T) {
	gen := NewGenerator()
	prog, err := gen.Generate(nil)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if prog == nil {
		t.Errorf("expected program, got nil")
	}
}

// TestGenerateIntLit validates integer literal generation
func TestGenerateIntLit(t *testing.T) {
	code := "let x = 42"
	p := parser.New(code)
	prog, err := p.ParseProgram()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	gen := NewGenerator()
	irProg, err := gen.Generate(prog)
	if err != nil {
		t.Errorf("generation error: %v", err)
	}

	if len(irProg.Main) == 0 {
		t.Errorf("expected instructions in main")
	}
}

// TestGenerateBinaryAdd validates binary addition
func TestGenerateBinaryAdd(t *testing.T) {
	code := "let x = 3 + 4"
	p := parser.New(code)
	prog, err := p.ParseProgram()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	gen := NewGenerator()
	irProg, err := gen.Generate(prog)
	if err != nil {
		t.Errorf("generation error: %v", err)
	}

	// Find OpAdd instruction
	found := false
	for _, instr := range irProg.Main {
		if instr.Op == OpAdd {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected OpAdd instruction")
	}
}

// TestGenerateLetDecl validates let declaration
func TestGenerateLetDecl(t *testing.T) {
	code := "let x = 10"
	p := parser.New(code)
	prog, err := p.ParseProgram()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	gen := NewGenerator()
	irProg, err := gen.Generate(prog)
	if err != nil {
		t.Errorf("generation error: %v", err)
	}

	// Should have at least OpCopy for assignment
	found := false
	for _, instr := range irProg.Main {
		if instr.Op == OpCopy && instr.Dest.Name == "x" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected OpCopy with dest 'x'")
	}
}

// TestGenerateFnDecl validates function declaration
func TestGenerateFnDecl(t *testing.T) {
	code := "fn add(a, b) { a + b }"
	p := parser.New(code)
	prog, err := p.ParseProgram()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	gen := NewGenerator()
	irProg, err := gen.Generate(prog)
	if err != nil {
		t.Errorf("generation error: %v", err)
	}

	if len(irProg.Functions) != 1 {
		t.Errorf("expected 1 function, got %d", len(irProg.Functions))
	}

	if irProg.Functions[0].Name != "add" {
		t.Errorf("expected function name 'add', got %s", irProg.Functions[0].Name)
	}
}

// TestGenerateIfStmt validates if statement
func TestGenerateIfStmt(t *testing.T) {
	code := "if x { let y = 1 }"
	p := parser.New(code)
	prog, err := p.ParseProgram()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	gen := NewGenerator()
	irProg, err := gen.Generate(prog)
	if err != nil {
		t.Errorf("generation error: %v", err)
	}

	// Should have OpJumpIfFalse
	found := false
	for _, instr := range irProg.Main {
		if instr.Op == OpJumpIfFalse {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected OpJumpIfFalse instruction")
	}
}

// TestGenerateForStmt validates for loop
func TestGenerateForStmt(t *testing.T) {
	code := "for i in 0..10 { let x = i }"
	p := parser.New(code)
	prog, err := p.ParseProgram()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	gen := NewGenerator()
	irProg, err := gen.Generate(prog)
	if err != nil {
		t.Errorf("generation error: %v", err)
	}

	// Should have loop label and jump
	labels := 0
	jumps := 0
	for _, instr := range irProg.Main {
		if instr.Op == OpLabel {
			labels++
		}
		if instr.Op == OpJump || instr.Op == OpJumpIfFalse {
			jumps++
		}
	}
	if labels == 0 {
		t.Errorf("expected labels in for loop")
	}
	if jumps == 0 {
		t.Errorf("expected jumps in for loop")
	}
}

// TestGenerateCallExpr validates function call
func TestGenerateCallExpr(t *testing.T) {
	code := "let x = add(1, 2)"
	p := parser.New(code)
	prog, err := p.ParseProgram()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	gen := NewGenerator()
	irProg, err := gen.Generate(prog)
	if err != nil {
		t.Errorf("generation error: %v", err)
	}

	// Should have OpCall
	found := false
	for _, instr := range irProg.Main {
		if instr.Op == OpCall {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected OpCall instruction")
	}
}

// TestGenerateReturn validates return statement
func TestGenerateReturn(t *testing.T) {
	code := "fn get() { return 42 }"
	p := parser.New(code)
	prog, err := p.ParseProgram()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	gen := NewGenerator()
	irProg, err := gen.Generate(prog)
	if err != nil {
		t.Errorf("generation error: %v", err)
	}

	if len(irProg.Functions) == 0 {
		t.Errorf("expected function")
		return
	}

	// Function should have OpReturn
	found := false
	for _, instr := range irProg.Functions[0].Body {
		if instr.Op == OpReturn {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected OpReturn in function")
	}
}

// TestProgramByteSize validates byte size calculation
func TestProgramByteSize(t *testing.T) {
	code := "let x = 10 let y = 20"
	p := parser.New(code)
	prog, err := p.ParseProgram()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	gen := NewGenerator()
	irProg, err := gen.Generate(prog)
	if err != nil {
		t.Errorf("generation error: %v", err)
	}

	byteSize := irProg.ByteSize()
	if byteSize <= 0 {
		t.Errorf("expected positive byte size, got %d", byteSize)
	}

	// Should be (instruction count) * 4
	expectedCount := len(irProg.Main)
	expectedSize := expectedCount * 4
	if byteSize != expectedSize {
		t.Errorf("expected byte size %d, got %d", expectedSize, byteSize)
	}
}
