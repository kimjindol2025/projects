package ir

import (
	"testing"
)

func TestModuleCreation(t *testing.T) {
	mod := NewModule()
	if mod == nil {
		t.Fatal("module is nil")
	}

	if len(mod.Functions) != 0 {
		t.Errorf("expected 0 functions, got %d", len(mod.Functions))
	}
}

func TestBasicBlockIR(t *testing.T) {
	bb := NewBasicBlock("test")
	if bb.Name != "test" {
		t.Errorf("expected block name 'test', got %s", bb.Name)
	}

	inst := &Instruction{Type: InstNop}
	bb.AddInstruction(inst)

	if len(bb.Insts) != 1 {
		t.Errorf("expected 1 instruction, got %d", len(bb.Insts))
	}

	if inst.ID != 0 {
		t.Errorf("expected instruction ID 0, got %d", inst.ID)
	}
}

func TestFunctionCreation(t *testing.T) {
	fn := NewFunction("test", "i64", []Value{})
	if fn.Name != "test" {
		t.Errorf("expected function name 'test', got %s", fn.Name)
	}

	if fn.ReturnType != "i64" {
		t.Errorf("expected return type 'i64', got %s", fn.ReturnType)
	}
}

func TestInstructionLiteral(t *testing.T) {
	inst := &Instruction{
		Type: InstLiteral,
		Meta: int64(42),
		Result: Value{
			ID:   0,
			Type: "i64",
		},
	}

	if inst.String() == "" {
		t.Error("instruction string should not be empty")
	}
}

func TestInstructionBinOp(t *testing.T) {
	inst := &Instruction{
		Type:   InstBinOp,
		OpType: "add",
		Ops:    []Value{{ID: 0}, {ID: 1}},
		Result: Value{ID: 2},
	}

	str := inst.String()
	if str == "" {
		t.Error("instruction string should not be empty")
	}
}

func TestModulePrint(t *testing.T) {
	mod := NewModule()
	fn := NewFunction("test", "i64", []Value{})
	block := NewBasicBlock("entry")
	fn.AddBlock(block)
	mod.Functions = append(mod.Functions, fn)

	output := mod.PrintModule()
	if len(output) == 0 {
		t.Error("module print output should not be empty")
	}

	if output == "" {
		t.Error("module print should produce output")
	}
}

func TestMultipleFunctions(t *testing.T) {
	mod := NewModule()

	fn1 := NewFunction("fn1", "i64", []Value{})
	fn2 := NewFunction("fn2", "i64", []Value{})

	mod.Functions = append(mod.Functions, fn1)
	mod.Functions = append(mod.Functions, fn2)

	if len(mod.Functions) != 2 {
		t.Errorf("expected 2 functions, got %d", len(mod.Functions))
	}
}

func TestBuilderBasic(t *testing.T) {
	builder := NewBuilder()
	if builder == nil {
		t.Fatal("builder should not be nil")
	}

	if builder.module == nil {
		t.Fatal("builder module should not be nil")
	}
}

func TestValue(t *testing.T) {
	v := Value{
		ID:   42,
		Type: "i64",
		Name: "test",
	}

	if v.ID != 42 {
		t.Errorf("expected ID 42, got %d", v.ID)
	}

	if v.Type != "i64" {
		t.Errorf("expected type i64, got %s", v.Type)
	}

	if v.Name != "test" {
		t.Errorf("expected name test, got %s", v.Name)
	}
}
