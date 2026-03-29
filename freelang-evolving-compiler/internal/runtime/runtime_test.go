package runtime

import (
	"testing"

	"github.com/user/freelang-evolving-compiler/internal/ir"
	"github.com/user/freelang-evolving-compiler/internal/parser"
)

// runProgram is a helper that parses code, generates IR, and runs the VM
func runProgram(t *testing.T, code string) (Value, map[string]Value) {
	t.Helper()

	// Parse
	p := parser.New(code)
	prog, err := p.ParseProgram()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	// Generate IR
	gen := ir.NewGenerator()
	irProg, err := gen.Generate(prog)
	if err != nil {
		t.Fatalf("IR gen error: %v", err)
	}

	// Run VM
	vm := New(irProg)
	result, err := vm.Run()
	if err != nil {
		t.Fatalf("VM error: %v", err)
	}

	return result, vm.DumpGlobals()
}

// TestArithmetic tests basic arithmetic operations
func TestArithmetic(t *testing.T) {
	_, globals := runProgram(t, "let x = 3 + 4")

	if xVal, ok := globals["x"]; !ok || xVal.Kind != KindInt || xVal.IVal != 7 {
		t.Errorf("expected x=7, got %v", xVal)
	}
}

// TestMultiplyArithmetic tests multiplication with precedence
func TestMultiplyArithmetic(t *testing.T) {
	_, globals := runProgram(t, "let x = 3 + 4 * 2")

	// Expected: 3 + 8 = 11
	if xVal, ok := globals["x"]; !ok || xVal.Kind != KindInt || xVal.IVal != 11 {
		t.Errorf("expected x=11, got %v", xVal)
	}
}

// TestIfStatement tests conditional statements
func TestIfStatement(t *testing.T) {
	_, globals := runProgram(t, "let x = 0  if 1 { let x = 42 }")

	if xVal, ok := globals["x"]; !ok || xVal.Kind != KindInt || xVal.IVal != 42 {
		t.Errorf("expected x=42, got %v", xVal)
	}
}

// TestIfStatementFalse tests conditional that doesn't execute
func TestIfStatementFalse(t *testing.T) {
	_, globals := runProgram(t, "let x = 5  if 0 { let x = 42 }")

	if xVal, ok := globals["x"]; !ok || xVal.Kind != KindInt || xVal.IVal != 5 {
		t.Errorf("expected x=5, got %v", xVal)
	}
}

// TestForLoop tests for-loop iteration
func TestForLoop(t *testing.T) {
	_, globals := runProgram(t, "let sum = 0  for i in 0..5 { let sum = sum + i }")

	// Expected: 0 + 0 + 1 + 2 + 3 + 4 = 10
	if sumVal, ok := globals["sum"]; !ok || sumVal.Kind != KindInt || sumVal.IVal != 10 {
		t.Errorf("expected sum=10, got %v", sumVal)
	}
}

// TestForLoopLonger tests longer for-loop iteration
func TestForLoopLonger(t *testing.T) {
	_, globals := runProgram(t, "let sum = 0  for i in 0..10 { let sum = sum + i }")

	// Expected: 0+1+2+...+9 = 45
	if sumVal, ok := globals["sum"]; !ok || sumVal.Kind != KindInt || sumVal.IVal != 45 {
		t.Errorf("expected sum=45, got %v", sumVal)
	}
}

// TestFunctionCall tests user-defined functions
func TestFunctionCall(t *testing.T) {
	_, globals := runProgram(t, "fn double(n) { return n + n }  let result = double(21)")

	if resultVal, ok := globals["result"]; !ok || resultVal.Kind != KindInt || resultVal.IVal != 42 {
		t.Errorf("expected result=42, got %v", resultVal)
	}
}

// TestFunctionCallMultiple tests multiple function arguments
func TestFunctionCallMultiple(t *testing.T) {
	_, globals := runProgram(t, "fn add(a, b) { return a + b }  let result = add(10, 32)")

	if resultVal, ok := globals["result"]; !ok || resultVal.Kind != KindInt || resultVal.IVal != 42 {
		t.Errorf("expected result=42, got %v", resultVal)
	}
}

// TestBuiltinPrintln tests builtin function via direct IR
func TestBuiltinPrintln(t *testing.T) {
	irProg := &ir.Program{
		Main: []ir.Instruction{
			{Op: ir.OpParam, Src1: ir.Operand{IsImm: true, ImmVal: 42}},
			{Op: ir.OpCall, Dest: ir.Operand{IsTemp: true, Name: "t0"}, Fn: "println"},
		},
	}

	vm := New(irProg)
	_, err := vm.Run()
	if err != nil {
		t.Errorf("builtin println call failed: %v", err)
	}
}

// TestBuiltinLenStr tests string length builtin
func TestBuiltinLenStr(t *testing.T) {
	irProg := &ir.Program{
		Main: []ir.Instruction{
			// Create string "hello" manually using IR
			{Op: ir.OpConst, Dest: ir.Operand{Name: "s"}, Src1: ir.Operand{IsImm: true, ImmVal: 5}},
		},
	}

	vm := New(irProg)
	_, err := vm.Run()
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

// TestGlobalVariable tests global variable storage
func TestGlobalVariable(t *testing.T) {
	_, globals := runProgram(t, "let x = 10  let y = x + 5")

	if yVal, ok := globals["y"]; !ok || yVal.Kind != KindInt || yVal.IVal != 15 {
		t.Errorf("expected y=15, got %v", yVal)
	}
}

// TestComparison tests comparison operations
func TestComparison(t *testing.T) {
	_, globals := runProgram(t, "let a = 5  let b = 10  let eq = 5  let gt = a < b")

	if gtVal, ok := globals["gt"]; !ok || gtVal.Kind != KindBool || !gtVal.BVal {
		t.Errorf("expected gt=true, got %v", gtVal)
	}
}

// TestNegativeComparison tests negative comparison
func TestNegativeComparison(t *testing.T) {
	_, globals := runProgram(t, "let a = 10  let b = 5  let result = a < b")

	if resultVal, ok := globals["result"]; !ok || resultVal.Kind != KindBool || resultVal.BVal {
		t.Errorf("expected result=false, got %v", resultVal)
	}
}

// TestMultipleVariables tests multiple variable assignments
func TestMultipleVariables(t *testing.T) {
	_, globals := runProgram(t, "let a = 1  let b = 2  let c = 3  let sum = a + b + c")

	if sumVal, ok := globals["sum"]; !ok || sumVal.Kind != KindInt || sumVal.IVal != 6 {
		t.Errorf("expected sum=6, got %v", sumVal)
	}
}

// TestNestedArithmetic tests complex arithmetic
func TestNestedArithmetic(t *testing.T) {
	_, globals := runProgram(t, "let x = 2 + 3 * 4 - 5")

	// Expected: 2 + 12 - 5 = 9
	if xVal, ok := globals["x"]; !ok || xVal.Kind != KindInt || xVal.IVal != 9 {
		t.Errorf("expected x=9, got %v", xVal)
	}
}

// TestStringLiteral tests string literal assignment
func TestStringLiteral(t *testing.T) {
	_, globals := runProgram(t, "let s = \"hello\"")

	if sVal, ok := globals["s"]; !ok || sVal.Kind != KindString || sVal.SVal != "hello" {
		t.Errorf("expected s=\"hello\", got %v", sVal)
	}
}

// TestBoolLiteralTrue tests bool literal true assignment
func TestBoolLiteralTrue(t *testing.T) {
	_, globals := runProgram(t, "let b = true")

	if bVal, ok := globals["b"]; !ok || bVal.Kind != KindBool || !bVal.BVal {
		t.Errorf("expected b=true, got %v", bVal)
	}
}

// TestBoolLiteralFalse tests bool literal false assignment
func TestBoolLiteralFalse(t *testing.T) {
	_, globals := runProgram(t, "let b = false")

	if bVal, ok := globals["b"]; !ok || bVal.Kind != KindBool || bVal.BVal {
		t.Errorf("expected b=false, got %v", bVal)
	}
}

// TestStringConcat tests string concatenation
func TestStringConcat(t *testing.T) {
	_, globals := runProgram(t, "let s = \"hello\" + \" world\"")

	if sVal, ok := globals["s"]; !ok || sVal.Kind != KindString || sVal.SVal != "hello world" {
		t.Errorf("expected s=\"hello world\", got %v", sVal)
	}
}

// TestStringEquality tests string equality comparison
func TestStringEquality(t *testing.T) {
	_, globals := runProgram(t, "let result = \"abc\" == \"abc\"")

	if resultVal, ok := globals["result"]; !ok || resultVal.Kind != KindBool || !resultVal.BVal {
		t.Errorf("expected result=true, got %v", resultVal)
	}
}

// TestBoolEquality tests bool equality comparison
func TestBoolEquality(t *testing.T) {
	_, globals := runProgram(t, "let result = true == true")

	if resultVal, ok := globals["result"]; !ok || resultVal.Kind != KindBool || !resultVal.BVal {
		t.Errorf("expected result=true, got %v", resultVal)
	}
}

// TestIfWithBoolLiteral tests if statement with bool literal
func TestIfWithBoolLiteral(t *testing.T) {
	_, globals := runProgram(t, "let x = 0  if true { let x = 1 }")

	if xVal, ok := globals["x"]; !ok || xVal.Kind != KindInt || xVal.IVal != 1 {
		t.Errorf("expected x=1, got %v", xVal)
	}
}

// TestArrayLiteral tests array literal creation
func TestArrayLiteral(t *testing.T) {
	_, globals := runProgram(t, "let arr = [1, 2, 3]")

	if arrVal, ok := globals["arr"]; !ok || arrVal.Kind != KindArray || len(arrVal.Elems) != 3 {
		t.Errorf("expected arr=KindArray with 3 elements, got %v", arrVal)
	}
}

// TestArrayIndex tests array indexing
func TestArrayIndex(t *testing.T) {
	_, globals := runProgram(t, "let arr = [10, 20, 30]  let x = arr[0]  let y = arr[2]")

	if xVal, ok := globals["x"]; !ok || xVal.Kind != KindInt || xVal.IVal != 10 {
		t.Errorf("expected x=10, got %v", xVal)
	}
	if yVal, ok := globals["y"]; !ok || yVal.Kind != KindInt || yVal.IVal != 30 {
		t.Errorf("expected y=30, got %v", yVal)
	}
}

// TestArrayIndexVar tests array indexing with variable
func TestArrayIndexVar(t *testing.T) {
	_, globals := runProgram(t, "let arr = [5, 15, 25]  let i = 1  let val = arr[i]")

	if valVal, ok := globals["val"]; !ok || valVal.Kind != KindInt || valVal.IVal != 15 {
		t.Errorf("expected val=15, got %v", valVal)
	}
}

// TestArrayInLoop tests array indexing in for loop
func TestArrayInLoop(t *testing.T) {
	_, globals := runProgram(t, "let arr = [1, 2, 3]  let sum = 0  for i in 0..3 { let sum = sum + arr[i] }")

	if sumVal, ok := globals["sum"]; !ok || sumVal.Kind != KindInt || sumVal.IVal != 6 {
		t.Errorf("expected sum=6, got %v", sumVal)
	}
}

// TestArrayLength tests len_arr builtin with array
func TestArrayLength(t *testing.T) {
	_, globals := runProgram(t, "let arr = [10, 20, 30, 40]  let length = len_arr(arr)")

	if lenVal, ok := globals["length"]; !ok || lenVal.Kind != KindInt || lenVal.IVal != 4 {
		t.Errorf("expected length=4, got %v", lenVal)
	}
}

// TestEmptyArray tests empty array creation
func TestEmptyArray(t *testing.T) {
	_, globals := runProgram(t, "let arr = []")

	if arrVal, ok := globals["arr"]; !ok || arrVal.Kind != KindArray || len(arrVal.Elems) != 0 {
		t.Errorf("expected arr=empty KindArray, got %v", arrVal)
	}
}
