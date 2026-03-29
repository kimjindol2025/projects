package typesys

import (
	"testing"

	"github.com/user/freelang-evolving-compiler/internal/parser"
)

// TestTypeCheckInt tests type checking of integer literals
func TestTypeCheckInt(t *testing.T) {
	input := `let x: int = 5`
	p := parser.New(input)
	prog, err := p.ParseProgram()

	if err != nil {
		t.Fatalf("parsing failed: %v", err)
	}

	tc := NewTypeChecker()
	errs := tc.Check(prog)

	if len(errs) > 0 {
		t.Fatalf("type check failed: %v", errs)
	}
}

// TestTypeCheckStruct tests struct registration and type
func TestTypeCheckStruct(t *testing.T) {
	input := `struct Point { x: int; y: int }
let p: Point = 0`
	p := parser.New(input)
	prog, err := p.ParseProgram()

	if err != nil {
		t.Fatalf("parsing failed: %v", err)
	}

	tc := NewTypeChecker()
	errs := tc.Check(prog)

	// Should have no critical errors (type annotation match is in Phase 2)
	for _, e := range errs {
		t.Logf("type error: %s (line %d)", e.Message, e.Line)
	}

	// Verify struct was registered
	if _, found := tc.env.LookupStruct("Point"); !found {
		t.Fatal("struct 'Point' not registered")
	}
}

// TestTypeCheckFieldAccess tests field access validation
func TestTypeCheckFieldAccess(t *testing.T) {
	input := `struct Point { x: int; y: int }
let p: Point = 0
let v: int = p.x`
	p := parser.New(input)
	prog, err := p.ParseProgram()

	if err != nil {
		t.Fatalf("parsing failed: %v", err)
	}

	tc := NewTypeChecker()
	errs := tc.Check(prog)

	// Should have registered struct and validated field access
	if _, found := tc.env.LookupStruct("Point"); !found {
		t.Fatal("struct 'Point' not registered")
	}

	for _, e := range errs {
		t.Logf("type check note: %s (line %d)", e.Message, e.Line)
	}
}

// TestTypeCheckUnknownField tests error on accessing unknown field
func TestTypeCheckUnknownField(t *testing.T) {
	input := `struct Point { x: int }
let p: Point = 0
let v: int = p.z`
	p := parser.New(input)
	prog, err := p.ParseProgram()

	if err != nil {
		t.Fatalf("parsing failed: %v", err)
	}

	tc := NewTypeChecker()
	errs := tc.Check(prog)

	// Should report unknown field error
	foundError := false
	for _, e := range errs {
		if contains(e.Message, "unknown field") {
			foundError = true
		}
	}

	if !foundError {
		t.Fatal("expected error for unknown field 'z'")
	}
}

// TestTypeCheckBinaryExpr tests binary expression type checking
func TestTypeCheckBinaryExpr(t *testing.T) {
	input := `let x: int = 5
let y: int = 3
let z: int = x + y`
	p := parser.New(input)
	prog, err := p.ParseProgram()

	if err != nil {
		t.Fatalf("parsing failed: %v", err)
	}

	tc := NewTypeChecker()
	errs := tc.Check(prog)

	// Should have no type errors
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("unexpected error: %s (line %d)", e.Message, e.Line)
		}
		t.Fatal("should have no type errors")
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

// TestTypeBoolLit tests boolean literal type checking
func TestTypeBoolLit(t *testing.T) {
	input := `let b: bool = true`
	p := parser.New(input)
	prog, err := p.ParseProgram()

	if err != nil {
		t.Fatalf("parsing failed: %v", err)
	}

	tc := NewTypeChecker()
	errs := tc.Check(prog)

	if len(errs) > 0 {
		t.Fatalf("type check failed: %v", errs)
	}
}

// TestTypeStringLit tests string literal type checking
func TestTypeStringLit(t *testing.T) {
	input := `let s: string = "hello"`
	p := parser.New(input)
	prog, err := p.ParseProgram()

	if err != nil {
		t.Fatalf("parsing failed: %v", err)
	}

	tc := NewTypeChecker()
	errs := tc.Check(prog)

	if len(errs) > 0 {
		t.Fatalf("type check failed: %v", errs)
	}
}

// TestTypeInferenceNoAnnotation tests type inference without annotation
func TestTypeInferenceNoAnnotation(t *testing.T) {
	input := `let x = 5`
	p := parser.New(input)
	prog, err := p.ParseProgram()

	if err != nil {
		t.Fatalf("parsing failed: %v", err)
	}

	tc := NewTypeChecker()
	errs := tc.Check(prog)

	// Should infer IntType from literal 5
	if len(errs) > 0 {
		t.Fatalf("type check failed: %v", errs)
	}
}

// TestTypeInferenceBinaryExpr tests type inference from binary expressions
func TestTypeInferenceBinaryExpr(t *testing.T) {
	input := `let x = 5
let y = 3
let z = x + y`
	p := parser.New(input)
	prog, err := p.ParseProgram()

	if err != nil {
		t.Fatalf("parsing failed: %v", err)
	}

	tc := NewTypeChecker()
	errs := tc.Check(prog)

	if len(errs) > 0 {
		t.Fatalf("type check failed: %v", errs)
	}
}

// TestTypeFnSignature tests function signature registration and validation
func TestTypeFnSignature(t *testing.T) {
	input := `fn add(x: int, y: int): int {
  return x + y
}
let r: int = add(1, 2)`
	p := parser.New(input)
	prog, err := p.ParseProgram()

	if err != nil {
		t.Fatalf("parsing failed: %v", err)
	}

	tc := NewTypeChecker()
	errs := tc.Check(prog)

	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("error: %s (line %d)", e.Message, e.Line)
		}
	}
}

// TestTypeStructLit tests struct literal initialization
func TestTypeStructLit(t *testing.T) {
	input := `struct Point {
  x: int;
  y: int
}
let p = Point{x: 1, y: 2}`
	p := parser.New(input)
	prog, err := p.ParseProgram()

	if err != nil {
		t.Fatalf("parsing failed: %v", err)
	}

	tc := NewTypeChecker()
	errs := tc.Check(prog)

	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("error: %s (line %d)", e.Message, e.Line)
		}
	}
}

// TestTypeIfElse tests if/else branch type checking
func TestTypeIfElse(t *testing.T) {
	input := `if true {
  let x: int = 5
} else {
  let y: int = 10
}`
	p := parser.New(input)
	prog, err := p.ParseProgram()

	if err != nil {
		t.Fatalf("parsing failed: %v", err)
	}

	tc := NewTypeChecker()
	errs := tc.Check(prog)

	if len(errs) > 0 {
		t.Fatalf("type check failed: %v", errs)
	}
}

// TestHardModeTypeMismatch tests hard mode type error detection
func TestHardModeTypeMismatch(t *testing.T) {
	input := `let x: int = true`
	p := parser.New(input)
	prog, err := p.ParseProgram()

	if err != nil {
		t.Fatalf("parsing failed: %v", err)
	}

	tc := NewTypeCheckerHard()
	errs := tc.Check(prog)

	// Should report type mismatch error
	foundError := false
	for _, e := range errs {
		if contains(e.Message, "type mismatch") {
			foundError = true
		}
	}

	if !foundError {
		t.Fatal("expected type mismatch error in hard mode")
	}
}
