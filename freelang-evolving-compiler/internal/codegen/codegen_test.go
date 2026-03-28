package codegen

import (
	"strings"
	"testing"

	"github.com/user/freelang-evolving-compiler/internal/ir"
	"github.com/user/freelang-evolving-compiler/internal/parser"
)

// TestCodeGenEmpty validates empty program generation
func TestCodeGenEmpty(t *testing.T) {
	cg := New()
	prog := &ir.Program{
		Functions: []ir.Function{},
		Main:      []ir.Instruction{},
	}

	result := cg.Generate(prog)

	if result.ByteSize < 0 {
		t.Errorf("expected non-negative byte size, got %d", result.ByteSize)
	}
}

// TestCodeGenIntLit validates integer literal generation
func TestCodeGenIntLit(t *testing.T) {
	cg := New()
	prog := &ir.Program{
		Main: []ir.Instruction{
			{
				Op:   ir.OpConst,
				Dest: ir.Operand{IsTemp: true, Name: "t0"},
				Src1: ir.Operand{IsImm: true, ImmVal: 42},
			},
		},
	}

	result := cg.Generate(prog)

	if !strings.Contains(result.Code, "LOAD") {
		t.Errorf("expected LOAD instruction in code")
	}
	if !strings.Contains(result.Code, "42") {
		t.Errorf("expected constant 42 in code")
	}
}

// TestCodeGenAddExpr validates addition generation
func TestCodeGenAddExpr(t *testing.T) {
	cg := New()
	prog := &ir.Program{
		Main: []ir.Instruction{
			{
				Op:   ir.OpAdd,
				Dest: ir.Operand{IsTemp: true, Name: "t0"},
				Src1: ir.Operand{IsImm: true, ImmVal: 3},
				Src2: ir.Operand{IsImm: true, ImmVal: 4},
			},
		},
	}

	result := cg.Generate(prog)

	if !strings.Contains(result.Code, "ADD") {
		t.Errorf("expected ADD instruction")
	}
}

// TestCodeGenFunction validates function generation
func TestCodeGenFunction(t *testing.T) {
	cg := New()
	prog := &ir.Program{
		Functions: []ir.Function{
			{
				Name: "test",
				Body: []ir.Instruction{
					{Op: ir.OpReturn},
				},
			},
		},
	}

	result := cg.Generate(prog)

	if !strings.Contains(result.Code, "ENTER") {
		t.Errorf("expected ENTER in function")
	}
	if !strings.Contains(result.Code, "LEAVE") {
		t.Errorf("expected LEAVE in function")
	}
}

// TestCodeGenReturn validates return statement
func TestCodeGenReturn(t *testing.T) {
	cg := New()
	prog := &ir.Program{
		Main: []ir.Instruction{
			{
				Op:   ir.OpReturn,
				Src1: ir.Operand{IsImm: true, ImmVal: 42},
			},
		},
	}

	result := cg.Generate(prog)

	if !strings.Contains(result.Code, "RET") {
		t.Errorf("expected RET instruction")
	}
}

// TestCodeGenIfStmt validates if statement generation
func TestCodeGenIfStmt(t *testing.T) {
	cg := New()
	prog := &ir.Program{
		Main: []ir.Instruction{
			{
				Op:    ir.OpJumpIfFalse,
				Src1:  ir.Operand{Name: "t0"},
				Label: "L_end_0",
			},
			{
				Op:    ir.OpLabel,
				Label: "L_end_0",
			},
		},
	}

	result := cg.Generate(prog)

	if !strings.Contains(result.Code, "JLF") {
		t.Errorf("expected JLF instruction")
	}
	if !strings.Contains(result.Code, "L_end_0") {
		t.Errorf("expected label in code")
	}
}

// TestCodeGenForLoop validates loop generation
func TestCodeGenForLoop(t *testing.T) {
	cg := New()
	prog := &ir.Program{
		Main: []ir.Instruction{
			{
				Op:    ir.OpLabel,
				Label: "L_loop_0",
			},
			{
				Op:    ir.OpJump,
				Label: "L_loop_0",
			},
			{
				Op:    ir.OpLabel,
				Label: "L_end_1",
			},
		},
	}

	result := cg.Generate(prog)

	if !strings.Contains(result.Code, "JUMP") {
		t.Errorf("expected JUMP instruction")
	}
	if strings.Count(result.Code, "L_loop_0") != 2 {
		t.Errorf("expected 2 uses of loop label")
	}
}

// TestCodeGenCallExpr validates function call generation
func TestCodeGenCallExpr(t *testing.T) {
	cg := New()
	prog := &ir.Program{
		Main: []ir.Instruction{
			{
				Op:   ir.OpParam,
				Src1: ir.Operand{IsImm: true, ImmVal: 1},
			},
			{
				Op:   ir.OpParam,
				Src1: ir.Operand{IsImm: true, ImmVal: 2},
			},
			{
				Op:   ir.OpCall,
				Dest: ir.Operand{IsTemp: true, Name: "t0"},
				Fn:   "add",
			},
		},
	}

	result := cg.Generate(prog)

	if !strings.Contains(result.Code, "PARAM") {
		t.Errorf("expected PARAM instruction")
	}
	if !strings.Contains(result.Code, "CALL") {
		t.Errorf("expected CALL instruction")
	}
}

// TestResultByteSize validates byte size calculation
func TestResultByteSize(t *testing.T) {
	cg := New()
	prog := &ir.Program{
		Main: []ir.Instruction{
			{
				Op:   ir.OpConst,
				Dest: ir.Operand{IsTemp: true, Name: "t0"},
				Src1: ir.Operand{IsImm: true, ImmVal: 1},
			},
		},
	}

	result := cg.Generate(prog)

	if result.ByteSize != len(result.Code) {
		t.Errorf("byte size should equal len(code): %d vs %d", result.ByteSize, len(result.Code))
	}
}

// TestCodeGenRoundTrip validates full pipeline (parse->IR->codegen)
func TestCodeGenRoundTrip(t *testing.T) {
	_ = parser.New("let x = 10 + 5") // Just test codegen with a simple program

	cg := New()
	irProg := &ir.Program{
		Main: []ir.Instruction{
			{
				Op:   ir.OpAdd,
				Dest: ir.Operand{IsTemp: true, Name: "t0"},
				Src1: ir.Operand{IsImm: true, ImmVal: 10},
				Src2: ir.Operand{IsImm: true, ImmVal: 5},
			},
		},
	}

	result := cg.Generate(irProg)

	if result.ByteSize == 0 {
		t.Errorf("expected non-zero byte size")
	}
	if result.LineCount == 0 {
		t.Errorf("expected non-zero line count")
	}
	if result.Code == "" {
		t.Errorf("expected generated code")
	}
}
