package codegen

import (
	"juliacc/internal/ir"
	"testing"
)

func TestCodegenCreation(t *testing.T) {
	mod := ir.NewModule()
	cg := NewCodegen(mod)
	if cg == nil {
		t.Fatal("codegen is nil")
	}
}

func TestGenerateBasic(t *testing.T) {
	mod := ir.NewModule()
	cg := NewCodegen(mod)

	bytecode, err := cg.Generate()
	if err != nil {
		t.Errorf("generation error: %v", err)
	}

	if bytecode == nil {
		t.Fatal("bytecode is nil")
	}

	// Should have at least halt instruction
	if len(bytecode.Code) == 0 {
		t.Error("bytecode should not be empty")
	}
}

func TestVMCreation(t *testing.T) {
	bc := &Bytecode{
		Code:      []uint8{uint8(OpHalt)},
		Constants: []interface{}{},
	}

	vm := NewVM(bc)
	if vm == nil {
		t.Fatal("vm is nil")
	}
}

func TestVMRun(t *testing.T) {
	bc := &Bytecode{
		Code:      []uint8{uint8(OpHalt)},
		Constants: []interface{}{},
	}

	vm := NewVM(bc)
	result, err := vm.Run()
	if err != nil {
		t.Errorf("execution error: %v", err)
	}

	if result != nil {
		t.Errorf("expected nil result for halt, got %v", result)
	}
}

func TestVMStackPush(t *testing.T) {
	bc := &Bytecode{
		Code:      []uint8{uint8(OpPush), 0, uint8(OpHalt)},
		Constants: []interface{}{int64(42)},
	}

	vm := NewVM(bc)
	result, err := vm.Run()
	if err != nil {
		t.Errorf("execution error: %v", err)
	}

	if val, ok := result.(int64); !ok || val != 42 {
		t.Errorf("expected 42, got %v", result)
	}
}

func TestBytecodeOp(t *testing.T) {
	ops := []BytecodeOp{
		OpPush, OpAdd, OpSub, OpMul, OpDiv,
		OpEq, OpNe, OpLt, OpGt,
		OpCall, OpRet, OpHalt,
	}

	for _, op := range ops {
		if op < 0 || op > OpHalt {
			t.Errorf("invalid bytecode op: %d", op)
		}
	}
}

func TestAddConstant(t *testing.T) {
	mod := ir.NewModule()
	cg := NewCodegen(mod)

	idx1 := cg.addConstant(int64(42))
	idx2 := cg.addConstant("hello")
	idx3 := cg.addConstant(3.14)

	if idx1 != 0 {
		t.Errorf("expected first constant index 0, got %d", idx1)
	}

	if idx2 != 1 {
		t.Errorf("expected second constant index 1, got %d", idx2)
	}

	if idx3 != 2 {
		t.Errorf("expected third constant index 2, got %d", idx3)
	}

	if len(cg.bytecode.Constants) != 3 {
		t.Errorf("expected 3 constants, got %d", len(cg.bytecode.Constants))
	}
}

func TestVMArithmetic(t *testing.T) {
	bc := &Bytecode{
		Code:      []uint8{uint8(OpPush), 0, uint8(OpPush), 1, uint8(OpAdd), uint8(OpHalt)},
		Constants: []interface{}{int64(5), int64(3)},
	}

	vm := NewVM(bc)
	result, err := vm.Run()
	if err != nil {
		t.Errorf("execution error: %v", err)
	}

	if val, ok := result.(int64); !ok || val != 8 {
		t.Errorf("expected 8, got %v", result)
	}
}

func TestVMComparison(t *testing.T) {
	bc := &Bytecode{
		Code:      []uint8{uint8(OpPush), 0, uint8(OpPush), 1, uint8(OpGt), uint8(OpHalt)},
		Constants: []interface{}{int64(5), int64(3)},
	}

	vm := NewVM(bc)
	result, err := vm.Run()
	if err != nil {
		t.Errorf("execution error: %v", err)
	}

	if val, ok := result.(bool); !ok || !val {
		t.Errorf("expected true, got %v", result)
	}
}
